/*
Copyright 2017 Heptio Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sinks

import (
	"bytes"
	"net/http"

	"github.com/eapache/channels"
	"github.com/golang/glog"
	"github.com/sethgrid/pester"

	"k8s.io/api/core/v1"
)

/*
The HTTP sink is a sink that sends events over HTTP using RFC5424 (syslog)
compatible messages. It establishes an HTTP connection with the remote
endpoint, sending messages as individual lines with the RFC5424 syntax:

<NumberOfBytes/ASCII encoded integer><Space character><RFC5424 message:NumberOfBytes long>

This is compatible with the protocol used by Heroku's Logplex:

https://github.com/heroku/logplex/blob/master/doc/README.http_drains.md

Many events may be coalesced into one request if they happen faster than we
can send them, if not, a single HTTP request is made for each event.
(Hopefully in a single keep-alive http connection, which is go's default.)

But with the payload of the messages being a serialized JSON object
containing the kubernetes v1.Event.
*/

// HTTPSink wraps an HTTP endpoint that messages should be sent to.
type HTTPSink struct {
	SinkURL string

	eventCh    channels.Channel
	httpClient *pester.Client
	bodyBuf    *bytes.Buffer
}

// NewHTTPSink constructs a new HTTPSink given a sink URL and buffer size
func NewHTTPSink(sinkURL string, overflow bool, bufferSize int) *HTTPSink {
	h := &HTTPSink{
		SinkURL: sinkURL,
	}

	if overflow {
		h.eventCh = channels.NewOverflowingChannel(channels.BufferCap(bufferSize))
	} else {
		h.eventCh = channels.NewNativeChannel(channels.BufferCap(bufferSize))
	}

	h.httpClient = pester.New()
	h.httpClient.Backoff = pester.ExponentialJitterBackoff
	h.httpClient.MaxRetries = 10
	// Let the body buffer be 4096 bytes at the start. It will be grown if
	// necessary.
	h.bodyBuf = bytes.NewBuffer(make([]byte, 0, 4096))

	return h
}

// UpdateEvents implements the EventSinkInterface. It really just writes the
// event data to the event OverflowingChannel, which should never block.
// Messages that are buffered beyond the bufferSize specified for this HTTPSink
// are discarded.
func (h *HTTPSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	h.eventCh.In() <- NewEventData(eNew, eOld)
}

// Run sits in a loop, waiting for data to come in through h.eventCh,
// and forwarding them to the HTTP sink. If multiple events have happened
// between loop iterations, it puts all of them in one request instead of
// making a single request per event.
func (h *HTTPSink) Run(stopCh <-chan bool) {
loop:
	for {
		select {
		case e := <-h.eventCh.Out():
			var evt EventData
			var ok bool
			if evt, ok = e.(EventData); !ok {
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

// drainEvents takes an array of event data and sends it to the receiving HTTP
// server. This function is *NOT* re-entrant: it re-uses the same body buffer
// for each call, truncating it each time to avoid extra memory allocations.
func (h *HTTPSink) drainEvents(events []EventData) {
	// Reuse the body buffer for each request
	h.bodyBuf.Truncate(0)

	var written int64
	for _, evt := range events {
		w, err := evt.WriteRFC5424(h.bodyBuf)
		written += w
		if err != nil {
			glog.Warningf("Could not write to event request body (wrote %v) bytes: %v", written, err)
			return
		}

		h.bodyBuf.Write([]byte{'\n'})
		written++
	}

	req, err := http.NewRequest("POST", h.SinkURL, h.bodyBuf)
	if err != nil {
		glog.Warningf(err.Error())
		return
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		glog.Warningf(err.Error())
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		glog.Warningf("Got HTTP code %v from %v", resp.StatusCode, h.SinkURL)
	}
}
