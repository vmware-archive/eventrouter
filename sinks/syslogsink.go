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
	"bytes"
	"log/syslog"
	"strconv"

	"github.com/eapache/channels"
	"github.com/golang/glog"

	"k8s.io/api/core/v1"
)

/*
The Syslog sink is a sink that sends events over TCP using RFC5424 (syslog)
compatible messages. It establishes a connection with the remote
endpoint, sending messages as individual lines with the RFC5424 syntax:

<NumberOfBytes/ASCII encoded integer><Space character><RFC5424 message:NumberOfBytes long>

This is compatible with the protocol used by Heroku's Logplex:

https://github.com/heroku/logplex/blob/master/doc/README.http_drains.md

Many events may be coalesced into one request if they happen faster than we
can send them, if not, a single Syslog request is made for each event.

But with the payload of the messages being a serialized JSON object
containing the kubernetes v1.Event.
*/

// SyslogSink implements the EvenSinkInterface
type SyslogSink struct {
	eventCh      channels.Channel
	syslogClient *syslog.Writer
	bodyBuf      *bytes.Buffer
}

// NewSyslogSink constructs a new SyslogSink given a sink URL and buffer size
func NewSyslogSink(endpoint string, port int, protocol string, overflow bool, bufferSize int, tag string) *SyslogSink {
	s := &SyslogSink{}

	if overflow {
		s.eventCh = channels.NewOverflowingChannel(channels.BufferCap(bufferSize))
	} else {
		s.eventCh = channels.NewNativeChannel(channels.BufferCap(bufferSize))
	}

	syslogAddress := endpoint + ":" + strconv.Itoa(port)

	syslogConnection, err := syslog.Dial(protocol, syslogAddress,
		syslog.LOG_WARNING|syslog.LOG_DAEMON, tag)
	if err != nil {
		glog.Warningf(err.Error())
	}

	s.syslogClient = syslogConnection

	// Let the body buffer be 4096 bytes at the start. It will be grown if
	// necessary.
	s.bodyBuf = bytes.NewBuffer(make([]byte, 0, 4096))

	return s
}

// UpdateEvents implements the EventSinkInterface. It really just writes the
// event data to the event OverflowingChannel, which should never block.
// Messages that are buffered beyond the bufferSize specified for this SyslogSink
// are discarded.
func (s *SyslogSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	s.eventCh.In() <- NewEventData(eNew, eOld)
}

// Run sits in a loop, waiting for data to come in through s.eventCh,
// and forwarding them to the Syslog sink. If multiple events have happened
// between loop iterations, it puts all of them in one request instead of
// making a single request per event.
func (s *SyslogSink) Run(stopCh <-chan bool) {
loop:
	for {
		select {
		case e := <-s.eventCh.Out():
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
			numEvents := s.eventCh.Len()
			for i := 0; i < numEvents; i++ {
				e := <-s.eventCh.Out()
				if evt, ok = e.(EventData); ok {
					arr = append(arr, evt)
				} else {
					glog.Warningf("Invalid type sent through event channel: %T", e)
				}
			}

			s.drainEvents(arr)
		case <-stopCh:
			break loop
		}
	}
}

// drainEvents takes an array of event data and sends it to the receiving Syslog
// server. This function is *NOT* re-entrant: it re-uses the same body buffer
// for each call, truncating it each time to avoid extra memory allocations.
func (s *SyslogSink) drainEvents(events []EventData) {
	// Reuse the body buffer for each request
	s.bodyBuf.Truncate(0)

	var written int64
	for _, evt := range events {
		w, err := evt.WriteRFC5424(s.bodyBuf)
		written += w
		if err != nil {
			glog.Warningf("Could not write to event request body (wrote %v) bytes: %v", written, err)
			return
		}

		s.bodyBuf.Write([]byte{'\n'})
		written++
	}

	glog.V(5).Infof("Sending Message:\n%v", s.bodyBuf)
	_, err := s.syslogClient.Write(s.bodyBuf.Bytes())
	if err != nil {
		glog.Warningf(err.Error())
		return
	}
}
