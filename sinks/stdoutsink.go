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
	"os"

	"k8s.io/api/core/v1"
)

// StdoutSink is the other basic sink
// By default, Fluentd/ElasticSearch won't index glog formatted lines
// By logging raw JSON to stdout, we will get automated indexing which
// can be queried in Kibana.
type StdoutSink struct {
	// TODO: create a channel and buffer for scaling
	namespace string
}


// NewStdoutSink will create a new StdoutSink with default options, returned as
// an EventSinkInterface
func NewStdoutSink(namespace string) EventSinkInterface {
	return &StdoutSink{
		namespace: namespace}
}

// UpdateEvents implements the EventSinkInterface
func (gs *StdoutSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	eData := NewEventData(eNew, eOld)
	
	if len(gs.namespace) > 0 {
		namespacedData := map[string]interface{}{}
		namespacedData[gs.namespace] = eData
		if eJSONBytes, err := json.Marshal(namespacedData); err == nil {
			fmt.Println(string(eJSONBytes))
		} else {
			fmt.Fprintf(os.Stderr, "Failed to json serialize event: %v", err)
		}
	} else {
		if eJSONBytes, err := json.Marshal(eData); err == nil {
			fmt.Println(string(eJSONBytes))
		} else {
			fmt.Fprintf(os.Stderr, "Failed to json serialize event: %v", err)
		}
	}
}
