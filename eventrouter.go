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

package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/heptiolabs/eventrouter/sinks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var (
	kubernetesWarningEventCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "heptio_eventrouter_warnings_total",
		Help: "Total number of warning events in the kubernetes cluster",
	}, []string{
		"involved_object_kind",
		"involved_object_name",
		"involved_object_namespace",
		"reason",
		"source",
	})
	kubernetesNormalEventCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "heptio_eventrouter_normal_total",
		Help: "Total number of normal events in the kubernetes cluster",
	}, []string{
		"involved_object_kind",
		"involved_object_name",
		"involved_object_namespace",
		"reason",
		"source",
	})
	kubernetesInfoEventCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "heptio_eventrouter_info_total",
		Help: "Total number of info events in the kubernetes cluster",
	}, []string{
		"involved_object_kind",
		"involved_object_name",
		"involved_object_namespace",
		"reason",
		"source",
	})
	kubernetesUnknownEventCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "heptio_eventrouter_unknown_total",
		Help: "Total number of events of unknown type in the kubernetes cluster",
	}, []string{
		"involved_object_kind",
		"involved_object_name",
		"involved_object_namespace",
		"reason",
		"source",
	})
)

// EventRouter is responsible for maintaining a stream of kubernetes
// system Events and pushing them to another channel for storage
type EventRouter struct {
	// kubeclient is the main kubernetes interface
	kubeClient kubernetes.Interface

	// store of events populated by the shared informer
	eLister corelisters.EventLister

	// returns true if the event store has been synced
	eListerSynched cache.InformerSynced

	// event sink
	// TODO: Determine if we want to support multiple sinks.
	eSink sinks.EventSinkInterface

	lastSeenResourceVersion     string
	lastResourceVersionPosition func(string)
}

// NewEventRouter will create a new event router using the input params
func NewEventRouter(kubeClient kubernetes.Interface, eventsInformer coreinformers.EventInformer,
	lastSeenResourceVersion string, lastResourceVersionPosition func(rv string)) *EventRouter {
	if viper.GetBool("enable-prometheus") {
		prometheus.MustRegister(kubernetesWarningEventCounterVec)
		prometheus.MustRegister(kubernetesNormalEventCounterVec)
		prometheus.MustRegister(kubernetesInfoEventCounterVec)
		prometheus.MustRegister(kubernetesUnknownEventCounterVec)
	}

	er := &EventRouter{
		kubeClient:                  kubeClient,
		eSink:                       sinks.ManufactureSink(),
		lastSeenResourceVersion:     lastSeenResourceVersion,
		lastResourceVersionPosition: lastResourceVersionPosition,
	}
	eventsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    er.addEvent,
		UpdateFunc: er.updateEvent,
		DeleteFunc: er.deleteEvent,
	})
	er.eLister = eventsInformer.Lister()
	er.eListerSynched = eventsInformer.Informer().HasSynced
	return er
}

// Run starts the EventRouter/Controller.
func (er *EventRouter) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer glog.Infof("Shutting down EventRouter")

	glog.Infof("Starting EventRouter")

	// here is where we kick the caches into gear
	if !cache.WaitForCacheSync(stopCh, er.eListerSynched) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	<-stopCh
}

// addEvent is called when an event is created, or during the initial list
func (er *EventRouter) addEvent(obj interface{}) {
	e := obj.(*v1.Event)
	if cast.ToInt(er.lastSeenResourceVersion) < cast.ToInt(e.ResourceVersion) {
		prometheusEvent(e)
		er.eSink.UpdateEvents(e, nil)
		er.lastResourceVersionPosition(e.ResourceVersion)
	} else {
		glog.V(5).Infof("Event had already been processed:\n%v", e)
	}
}

// updateEvent is called any time there is an update to an existing event
func (er *EventRouter) updateEvent(objOld interface{}, objNew interface{}) {
	eOld := objOld.(*v1.Event)
	eNew := objNew.(*v1.Event)
	if cast.ToInt(er.lastSeenResourceVersion) < cast.ToInt(eNew.ResourceVersion) {
		prometheusEvent(eNew)
		er.eSink.UpdateEvents(eNew, eOld)
		er.lastResourceVersionPosition(eNew.ResourceVersion)
	} else {
		glog.V(5).Infof("Event had already been processed:\n%v", eNew)
	}
}

// prometheusEvent is called when an event is added or updated
func prometheusEvent(event *v1.Event) {
	if !viper.GetBool("enable-prometheus") {
		return
	}
	var counter prometheus.Counter
	var err error

	switch event.Type {
	case "Normal":
		counter, err = kubernetesNormalEventCounterVec.GetMetricWithLabelValues(
			event.InvolvedObject.Kind,
			event.InvolvedObject.Name,
			event.InvolvedObject.Namespace,
			event.Reason,
			event.Source.Host,
		)
	case "Warning":
		counter, err = kubernetesWarningEventCounterVec.GetMetricWithLabelValues(
			event.InvolvedObject.Kind,
			event.InvolvedObject.Name,
			event.InvolvedObject.Namespace,
			event.Reason,
			event.Source.Host,
		)
	case "Info":
		counter, err = kubernetesInfoEventCounterVec.GetMetricWithLabelValues(
			event.InvolvedObject.Kind,
			event.InvolvedObject.Name,
			event.InvolvedObject.Namespace,
			event.Reason,
			event.Source.Host,
		)
	default:
		counter, err = kubernetesUnknownEventCounterVec.GetMetricWithLabelValues(
			event.InvolvedObject.Kind,
			event.InvolvedObject.Name,
			event.InvolvedObject.Namespace,
			event.Reason,
			event.Source.Host,
		)
	}

	if err != nil {
		// Not sure this is the right place to log this error?
		glog.Warning(err)
	} else {
		counter.Add(1)
	}
}

// deleteEvent should only occur when the system garbage collects events via TTL expiration
func (er *EventRouter) deleteEvent(obj interface{}) {
	e := obj.(*v1.Event)
	// NOTE: This should *only* happen on TTL expiration there
	// is no reason to push this to a sink
	glog.V(5).Infof("Event Deleted from the system:\n%v", e)
}
