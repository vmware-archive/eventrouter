package sinks

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	v1 "k8s.io/api/core/v1"
)

type PubSubSink struct {
	ProjectId  string
	client     *pubsub.Client
	topic      *pubsub.Topic
	deadLetter *pubsub.Topic
}

func NewPubSubSink(ctx context.Context, projectId string, topic string) (*PubSubSink, error) {
	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		return nil, err
	}
	clientTopic := client.Topic(topic)
	ok, err := clientTopic.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("topic %s does not exist", topic)
	}

	return &PubSubSink{
		ProjectId:  projectId,
		topic:      clientTopic,
		deadLetter: nil,
	}, nil
}

func NewPubSubSinkWithDeadLetter(ctx context.Context, projectId string, topic string, deadLetterTopic string) (*PubSubSink, error) {
	ps, err := NewPubSubSink(ctx, projectId, topic)
	if err != nil {
		return nil, err
	}
	clientDeadLetterTopic := ps.client.Topic(deadLetterTopic)
	ok, err := clientDeadLetterTopic.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("dead-letter topic %s does not exist", deadLetterTopic)
	}
	ps.deadLetter = clientDeadLetterTopic
	return ps, nil
}

func (x *PubSubSink) publishAsync(ctx context.Context, message []byte, topic *pubsub.Topic) *pubsub.PublishResult {
	return topic.Publish(ctx, &pubsub.Message{
		Data: message,
	})
}

func (x *PubSubSink) publishSync(ctx context.Context, message []byte, topic *pubsub.Topic) (string, error) {
	res := x.publishAsync(ctx, message, topic)
	msgID, err := res.Get(ctx) // blocks until published
	if err != nil {
		return "", err
	}
	return msgID, nil
}

func (x *PubSubSink) Cleanup() {
	x.topic.Stop()
	if x.deadLetter != nil {
		x.deadLetter.Stop()
	}
}

// UpdateEvents implements EventSinkInterface.UpdateEvents
func (x *PubSubSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	eData := NewEventData(eNew, eOld)

	if eJSONBytes, err := json.Marshal(eData); err == nil {
		x.publishAsync(context.Background(), eJSONBytes, x.topic)
	} else {
		if x.deadLetter != nil {
			x.publishAsync(context.Background(), eJSONBytes, x.deadLetter)
		}
	}
}
