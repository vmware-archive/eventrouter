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

	"github.com/golang/glog"
	"k8s.io/client-go/pkg/api/v1"
)

type jsonSink struct {
	// TODO: create a channel and buffer for scaling
}

// Provides a sync that embeds the event(s) and the verb "add" or "update" accordingly, for easy parsing.
func NewJSONSink() EventSinkInterface {
	return &jsonSink{}
}

// UpdateEvents implements the EventSinkInterface
func (gs *jsonSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	var eventBlob map[string]interface{}
	if eOld == nil {
		eventBlob = map[string]interface{}{
			"verb":  "add",
			"event": eNew,
		}
	} else {
		eventBlob = map[string]interface{}{
			"verb":      "update",
			"event":     eNew,
			"old_event": eOld,
		}
	}
	if eventBlobBytes, err := json.Marshal(eventBlob); err == nil {
		glog.Info(string(eventBlobBytes))
	} else {
		glog.Warningf("Failed to json serialize event: %v", err)
	}
}
