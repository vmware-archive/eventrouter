package sinks

import (
	"context"
	"encoding/json"

	eventhub "github.com/Azure/azure-event-hubs-go/v2"
	"github.com/eapache/channels"
	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
)

const maxMessageSize = 1046528

// EventHubSink sends events to an Azure Event Hub.
type EventHubSink struct {
	hub     *eventhub.Hub
	eventCh channels.Channel
}

// NewEventHubSink constructs a new EventHubSink given a event hub connection string
// and buffering options.
//
// ```
// export EVENTHUB_RESOURCE_GROUP=eventrouter
// export EVENTHUB_NAMESPACE=eventrouter-ns                                                                                                               <<<
// export EVENTHUB_NAME=eventrouter
// export EVENTHUB_REGION=westus2
// export EVENTHUB_RULE_NAME=eventrouter-send
//
// az group create -g ${EVENTHUB_RESOURCE_GROUP} -l ${EVENTHUB_REGION}
// az eventhubs namespace create -g ${EVENTHUB_RESOURCE_GROUP} -n ${EVENTHUB_NAMESPACE} -l ${EVENTHUB_REGION}
// az eventhubs eventhub create -g ${EVENTHUB_RESOURCE_GROUP} --namespace-name ${EVENTHUB_NAMESPACE} -n ${EVENTHUB_NAME}
// az eventhubs eventhub authorization-rule create -g ${EVENTHUB_RESOURCE_GROUP} --namespace-name ${EVENTHUB_NAMESPACE} --eventhub-name ${EVENTHUB_NAME} -n ${EVENTHUB_RULE_NAME} --rights Send
// export EVENTHUB_CONNECTION_STRING=$(az eventhubs eventhub authorization-rule keys list -g ${EVENTHUB_RESOURCE_GROUP} --namespace-name ${EVENTHUB_NAMESPACE} --eventhub-name ${EVENTHUB_NAME} -n ${EVENTHUB_RULE_NAME} | jq -r '.primaryConnectionString')
//
// cat yaml/eventrouter-azure.yaml | envsubst | kubectl apply -f
// ```
//
// connString expects the Azure Event Hub connection string format:
//		`Endpoint=sb://YOUR_ENDPOINT.servicebus.windows.net/;SharedAccessKeyName=YOUR_ACCESS_KEY_NAME;SharedAccessKey=YOUR_ACCESS_KEY;EntityPath=YOUR_EVENT_HUB_NAME`
func NewEventHubSink(connString string, overflow bool, bufferSize int) (*EventHubSink, error) {
	hub, err := eventhub.NewHubFromConnectionString(connString)
	if err != nil {
		return nil, err
	}
	var eventCh channels.Channel
	if overflow {
		eventCh = channels.NewOverflowingChannel(channels.BufferCap(bufferSize))
	} else {
		eventCh = channels.NewNativeChannel(channels.BufferCap(bufferSize))
	}

	return &EventHubSink{hub: hub, eventCh: eventCh}, nil
}

// UpdateEvents implements the EventSinkInterface. It really just writes the
// event data to the event OverflowingChannel, which should never block.
// Messages that are buffered beyond the bufferSize specified for this EventHubSink
// are discarded.
func (h *EventHubSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	h.eventCh.In() <- NewEventData(eNew, eOld)
}

// Run sits in a loop, waiting for data to come in through h.eventCh,
// and forwarding them to the event hub sink. If multiple events have happened
// between loop iterations, it puts all of them in one request instead of
// making a single request per event.
func (h *EventHubSink) Run(stopCh <-chan bool) {
loop:
	for {
		select {
		case e := <-h.eventCh.Out():
			var evt EventData
			var ok bool
			evt, ok = e.(EventData)
			if !ok {
				glog.Warningf("Invalid type sent through event channel: %T", e)
				continue loop
			}

			// Start with just this event...
			arr := []EventData{evt}

			// Consume all buffered events into an array, in case more have been written
			// since we last forwarded them
			numEvents := h.eventCh.Len()
			for i := 0; i < numEvents; i++ {
				e := <-h.eventCh.Out()
				if evt, ok = e.(EventData); ok {
					arr = append(arr, evt)
				} else {
					glog.Warningf("Invalid type sent through event channel: %T", e)
				}
			}

			h.drainEvents(arr)
		case <-stopCh:
			break loop
		}
	}
}

// drainEvents takes an array of event data and sends it to the receiving event hub.
func (h *EventHubSink) drainEvents(events []EventData) {
	var messageSize int
	var evts []*eventhub.Event
	for _, evt := range events {
		eJSONBytes, err := json.Marshal(evt)
		if err != nil {
			glog.Warningf("Failed to flatten json: %v", err)
			return
		}
		glog.V(4).Infof("%s", string(eJSONBytes))
		messageSize += len(eJSONBytes)
		if messageSize > maxMessageSize {
			h.sendBatch(evts)
			evts = nil
			messageSize = 0
		}
		evts = append(evts, eventhub.NewEvent(eJSONBytes))
	}
	h.sendBatch(evts)
}

func (h *EventHubSink) sendBatch(evts []*eventhub.Event) {
	if err := h.hub.SendBatch(context.Background(), eventhub.NewEventBatchIterator(evts...)); err != nil {
		glog.Errorf("Failed to send batch of %d: %v", len(evts), err)
	}
}
