/*
Copyright 2017 The Contributors

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
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/xdg/scram"
	v1 "k8s.io/api/core/v1"
)

var (
	SHA256 scram.HashGeneratorFcn = sha256.New
	SHA512 scram.HashGeneratorFcn = sha512.New
)

type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}

func (x *XDGSCRAMClient) Step(challenge string) (response string, err error) {
	response, err = x.ClientConversation.Step(challenge)
	return
}

func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}

// KafkaSink implements the EventSinkInterface
type KafkaSink struct {
	Topic    string
	producer interface{}
}

// NewKafkaSinkSink will create a new KafkaSink with default options, returned as an EventSinkInterface
func NewKafkaSink(brokers []string, topic string, async bool, retryMax int, saslUser string, saslPwd string, saslMechanism string) (EventSinkInterface, error) {

	p, err := sinkFactory(brokers, async, retryMax, saslUser, saslPwd, saslMechanism)

	if err != nil {
		return nil, err
	}

	return &KafkaSink{
		Topic:    topic,
		producer: p,
	}, err
}

func sinkFactory(brokers []string, async bool, retryMax int, saslUser string, saslPwd string, saslMechanism string) (interface{}, error) {
	config := sarama.NewConfig()
	config.Producer.Retry.Max = retryMax
	config.Producer.RequiredAcks = sarama.WaitForAll

	if saslUser != "" && saslPwd != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = saslUser
		config.Net.SASL.Password = saslPwd

		if saslMechanism == "sha256" {
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA256} }
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		} else if saslMechanism == "sha512" {
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA512} }
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		}
	}

	if async {
		return sarama.NewAsyncProducer(brokers, config)
	}

	config.Producer.Return.Successes = true
	return sarama.NewSyncProducer(brokers, config)

}

// UpdateEvents implements EventSinkInterface.UpdateEvents
func (ks *KafkaSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {

	eData := NewEventData(eNew, eOld)

	eJSONBytes, err := json.Marshal(eData)
	if err != nil {
		glog.Errorf("Failed to json serialize event: %v", err)
		return
	}
	msg := &sarama.ProducerMessage{
		Topic: ks.Topic,
		Key:   sarama.StringEncoder(eNew.InvolvedObject.Name),
		Value: sarama.ByteEncoder(eJSONBytes),
	}

	switch p := ks.producer.(type) {
	case sarama.SyncProducer:
		partition, offset, err := p.SendMessage(msg)
		if err != nil {
			glog.Errorf("Failed to send to: topic(%s)/partition(%d)/offset(%d)\n",
				ks.Topic, partition, offset)
		}

	case sarama.AsyncProducer:
		select {
		case p.Input() <- msg:
		case err := <-p.Errors():
			glog.Errorf("Failed to produce message: %v", err)
		}

	default:
		glog.Errorf("Unhandled producer type: %s", p)
	}

}
