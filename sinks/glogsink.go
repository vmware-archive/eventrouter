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
	"github.com/golang/glog"
	"k8s.io/client-go/pkg/api/v1"
)

// GlogSink is the most basic sink
// Useful when you already have EFK Stack
type GlogSink struct {
	// TODO: maybe create a channel here?
}

// NewGlogSink will create a new
func NewGlogSink() EventSinkInterface {
	return &GlogSink{}
}

// UpdateEvents implements the EventSinkInterface
func (gs *GlogSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	if eOld != nil {
		glog.Infof("Event Updated from the system FROM:\n%v\nTO:%v", eOld, eNew)
	} else {
		glog.Infof("Event Added to the system:\n%v", eNew)
	}
}
