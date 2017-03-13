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

// GlogSink is the most basic sink
// Useful when you already have EFK Stack
type GlogSink struct {
	// TODO: create a channel and buffer for scaling
}

// NewGlogSink will create a new
func NewGlogSink() EventSinkInterface {
	return &GlogSink{}
}

// UpdateEvents implements the EventSinkInterface
func (gs *GlogSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	nEvent, err := json.Marshal(eNew)
	if err != nil {
		glog.Warningf("Failed to json serialize new element:\n%v", eNew)
	} else {
		if eOld != nil {
			oEvent, err := json.Marshal(eOld)
			if err != nil {
				glog.Warningf("Failed to json serialize old element:\n%v", eOld)
			} else {
				glog.Infof("Event UPDATED in the system FROM:\n%v\nTO:%v", string(oEvent), string(nEvent))
			}
		} else {
			glog.Infof("Event ADDED to the system:\n%v", string(nEvent))
		}
	}
}
