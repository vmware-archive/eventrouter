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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ref "k8s.io/client-go/tools/reference"

	"k8s.io/api/core/v1"
)

func TestUpdateEvents(t *testing.T) {
	stopCh := make(chan bool, 1)
	doneCh := make(chan bool, 1)

	got := bytes.NewBuffer(nil)
	seenRequests := make([]*http.Request, 0)
	mockStatus := http.StatusOK

	// Make a test server to send stuff too... it just copies its input to the
	// `got` buffer and records the request.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenRequests = append(seenRequests, r)
		io.Copy(got, r.Body)
		w.WriteHeader(mockStatus)
	}))
	defer srv.Close()

	testPod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			SelfLink:  "/api/version/pods/foo",
			Name:      "foo",
			Namespace: "baz",
			UID:       "bar",
		},
		Spec: v1.PodSpec{},
	}
	podRef, err := ref.GetReference(scheme.Scheme, testPod)
	if err != nil {
		t.Fatalf(err.Error())
	}

	evt := makeFakeEvent(podRef, v1.EventTypeWarning, "CreateInCluster", "Fake pod creation event")

	// 1. Try with a synchronous channel
	sink := NewHTTPSink(srv.URL, false, 0)
	go func() {
		sink.Run(stopCh)
		doneCh <- true
	}()

	// Send the event
	sink.UpdateEvents(evt, nil)
	stopCh <- true
	<-doneCh

	if got.Len() == 0 {
		t.Errorf("Sent logs but didn't read any back")
	}

	// 2. Try with the server returning 500's, test retries
	got.Truncate(0)
	seenRequests = make([]*http.Request, 0)
	sink = NewHTTPSink(srv.URL, false, 10)

	go func() {
		sink.Run(stopCh)
		doneCh <- true
	}()

	// Send the event, sleep to ensure the request is attempted
	mockStatus = http.StatusInternalServerError
	sink.UpdateEvents(evt, nil)
	// TODO(SLEEP): this can result in flakes if the events aren't sent yet.
	time.Sleep(100 * time.Millisecond)
	mockStatus = http.StatusOK

	// Start the server, then send the stop chan. Since it's synchronous, the HTTP
	// client should still be trying to retry, so it won't read from the stop chan
	// again until it's finished retrying
	stopCh <- true
	<-doneCh

	if got.Len() == 0 {
		t.Errorf("Sent logs but didn't read any back. HTTP error log: %v", sink.httpClient.ErrLog)
	}
	if len(seenRequests) < 2 {
		t.Errorf("Tried to simulate server errors for retry, more than one request should have been sent")
	}

	// 3. Try with an overflowing channel, write a bunch of events out, only 10
	// should be consumed (the rest discarded, since we're not running the
	// processing loop yet.)
	numExpected := 10
	got.Truncate(0)
	seenRequests = make([]*http.Request, 0)
	sink = NewHTTPSink(srv.URL, true, numExpected)

	for i := 0; i < 1000; i++ {
		evt.Message = "msg " + strconv.Itoa(i)
		sink.UpdateEvents(evt, nil)
	}

	go func() {
		sink.Run(stopCh)
		doneCh <- true
	}()

	// TODO(SLEEP): Let the events go through (yes, sleeping is lame but there's
	// no easy way to synchronize this since the code is supposed to be
	// non-blocking.)
	time.Sleep(100 * time.Millisecond)

	stopCh <- true
	<-doneCh

	newlines := strings.Count(got.String(), "\n")
	if newlines != numExpected {
		t.Errorf("Got wrong number of lines back (got %v, expected %v)", newlines, numExpected)
	}
	if len(seenRequests) > 1 {
		t.Errorf("Pending logs should have been coalesced into one request, but got %v requests", len(seenRequests))
	}
}

func makeFakeEvent(ref *v1.ObjectReference, eventtype, reason, message string) *v1.Event {
	tm := metav1.Time{
		Time: time.Now(),
	}
	return &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v.%x", ref.Name, tm.UnixNano()),
			Namespace: ref.Namespace,
		},
		InvolvedObject: *ref,
		Reason:         reason,
		Message:        message,
		FirstTimestamp: tm,
		LastTimestamp:  tm,
		Count:          1,
		Type:           eventtype,
	}
}
