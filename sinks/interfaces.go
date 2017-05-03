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
	"errors"

	"github.com/golang/glog"
	"github.com/spf13/viper"

	"k8s.io/client-go/pkg/api/v1"
)

// EventSinkInterface is the interface used to shunt events
type EventSinkInterface interface {
	UpdateEvents(eNew *v1.Event, eOld *v1.Event)
}

// ManufactureSink will manufacture a sink according to viper configs
// TODO: Determine if it should return an array of sinks
func ManufactureSink() (e EventSinkInterface) {
	s := viper.GetString("sink")
	glog.Infof("Sink is [%v]", s)
	switch s {
	case "glog":
		e = NewGlogSink()
	case "stdout":
		e = NewStdoutSink()
	// case "kafka"
	// case "logfile"
	default:
		err := errors.New("Invalid Sink Specified")
		panic(err.Error())
	}
	return e
}
