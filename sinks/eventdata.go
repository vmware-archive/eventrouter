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
	"encoding/json"
	"fmt"
	"io"

	"github.com/crewjam/rfc5424"
	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"k8s.io/api/core/v1"

	"github.com/nytlabs/gojsonexplode"
)

// EventData encodes an eventrouter event and previous event, with a verb for
// whether the event is created or updated.
type EventData struct {
	Verb     string    `json:"verb"`
	Event    *v1.Event `json:"event"`
	OldEvent *v1.Event `json:"old_event,omitempty"`
}

// NewEventData constructs an EventData struct from an old and new event,
// setting the verb accordingly
func NewEventData(eNew *v1.Event, eOld *v1.Event) EventData {
	var eData EventData
	if eOld == nil {
		eData = EventData{
			Verb:  "ADDED",
			Event: eNew,
		}
	} else {
		eData = EventData{
			Verb:     "UPDATED",
			Event:    eNew,
			OldEvent: eOld,
		}
	}

	return eData
}

// WriteRFC5424 writes the current event data to the given io.Writer using
// RFC5424 (syslog over TCP) syntax.
func (e *EventData) WriteRFC5424(w io.Writer) (int64, error) {
	var eJSONBytes []byte
	var err error
	if eJSONBytes, err = json.Marshal(e); err != nil {
		return 0, fmt.Errorf("failed to json serialize event: %v", err)
	}

	// Each message should look like an RFC5424 syslog message:
	// <NumberOfBytes/ASCII encoded integer><Space character><RFC5424 message:NumberOfBytes long>
	//
	// Note: There are some restrictions on length and character space for
	// Hostname and AppName, see
	// https://github.com/crewjam/rfc5424/blob/master/marshal.go#L90. There's no
	// attempt at trying to clean them up here because hostnames and component
	// names already adhere to this convention in practice.
	msg := rfc5424.Message{
		Priority:  rfc5424.Daemon,
		Timestamp: e.Event.LastTimestamp.Time,
		Hostname:  e.Event.Source.Host,
		AppName:   e.Event.Source.Component,
		Message:   eJSONBytes,
	}

	return msg.WriteTo(w)
}

// WriteFlattenedJSON writes the json to the file in the below format
// 1) Flattens the json into a not nested key:value
// 2) Convert the json into snake format
// Eg: {"event_involved_object_kind":"pod", "event_metadata_namespace":"kube-system"}
func (e *EventData) WriteFlattenedJSON(w io.Writer) (int64, error) {
	var eJSONBytes []byte
	var err error
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
	if eJSONBytes, err = jsoniter.Marshal(e); err != nil {
		return 0, fmt.Errorf("failed to json serialize event: %v", err)
	}

	result, err := gojsonexplode.Explodejsonstr(string(eJSONBytes), "_")
	if err != nil {
		return 0, fmt.Errorf("failed to flatten json: %v", err)
	}

	written, err := w.Write([]byte(result))
	return int64(written), err
}
