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
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// setup a signal hander to gracefully exit
func sigHandler() <-chan struct{} {
	stop := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // Termination Request
			syscall.SIGSEGV, // FullDerp
			syscall.SIGABRT, // Abnormal termination
			syscall.SIGILL,  // illegal instruction
			syscall.SIGFPE)  // floating point - this is why we can't have nice things
		sig := <-c
		glog.Warningf("Signal (%v) Detected, Shutting Down", sig)
		close(stop)
	}()
	return stop
}

func main() {
	// TODO: update kubeconfig to use internal version
	// will need RBAC'd config
	kubeconfig := flag.String("kubeconfig", "/var/run/kubernetes/admin.kubeconfig", "absolute path to the kubeconfig file")
	rsInterval := flag.Duration("resync-interval", time.Minute*30, "default resync interval")
	// TODO: see EventAppender
	flag.Parse()

	// uses the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset from kubeconfig
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// initialize the shared informers
	sharedInformers := informers.NewSharedInformerFactory(clientset, *rsInterval)
	eventsInformer := sharedInformers.Core().V1().Events()

	// create a new event router
	// TODO: Do locking for HA, b/c this is super important
	// probably a closure to select on stop or os.Signal
	eventRouter := NewEventRouter(clientset, eventsInformer)
	stop := sigHandler()
	glog.Infof("Starting Event Router")
	go eventRouter.Run(stop)
	glog.Infof("Starting Informer(s)")
	sharedInformers.Start(stop)
	<-stop
	glog.Warningf("Exiting main()")
}
