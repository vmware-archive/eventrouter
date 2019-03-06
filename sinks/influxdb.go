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
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	influxdb "github.com/influxdata/influxdb/client"

	"k8s.io/api/core/v1"
)

var (
	LabelPodId = LabelDescriptor{
		Key:         "pod_id",
		Description: "The unique ID of the pod",
	}

	LabelPodName = LabelDescriptor{
		Key:         "pod_name",
		Description: "The name of the pod",
	}

	LabelNamespaceName = LabelDescriptor{
		Key:         "namespace_name",
		Description: "The name of the namespace",
	}

	LabelHostname = LabelDescriptor{
		Key:         "hostname",
		Description: "Hostname where the container ran",
	}
)

const (
	eventMeasurementName = "k8s_events"
	// Event special tags
	eventUID = "uid"
	// Value Field name
	valueField = "value"
	// Event special tags
	dbNotFoundError = "database not found"
)

type LabelDescriptor struct {
	// Key to use for the label.
	Key string `json:"key,omitempty"`

	// Description of the label.
	Description string `json:"description,omitempty"`
}

type InfluxDBSink struct {
	config InfluxdbConfig
	client *influxdb.Client
	sync.RWMutex
	dbExists bool
}

type InfluxdbConfig struct {
	User                  string
	Password              string
	Secure                bool
	Host                  string
	DbName                string
	WithFields            bool
	InsecureSsl           bool
	RetentionPolicy       string
	ClusterName           string
	DisableCounterMetrics bool
	Concurrency           int
}

// Returns a thread-safe implementation of EventSinkInterface for InfluxDB.
func NewInfuxdbSink(cfg InfluxdbConfig) (EventSinkInterface, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}

	return &InfluxDBSink{
		config:   cfg,
		client:   client,
		dbExists: false,
	}, nil
}

func newClient(c InfluxdbConfig) (*influxdb.Client, error) {
	url := &url.URL{
		Scheme: "http",
		Host:   c.Host,
	}

	if c.Secure {
		url.Scheme = "https"
	}

	iConfig := &influxdb.Config{
		URL:       *url,
		Username:  c.User,
		Password:  c.Password,
		UnsafeSsl: c.InsecureSsl,
	}

	client, err := influxdb.NewClient(*iConfig)
	if err != nil {
		return nil, err
	}

	if _, _, err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping influxDB server at %q - %v", c.Host, err)
	}

	return client, nil
}

func (sink *InfluxDBSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	sink.Lock()
	defer sink.Unlock()

	var point *influxdb.Point
	var err error
	if sink.config.WithFields {
		point, err = eventToPointWithFields(eNew)
	} else {
		point, err = eventToPoint(eNew)
	}
	if err != nil {
		glog.Warningf("Failed to convert event to point: %v", err)
	}

	point.Tags["cluster_name"] = sink.config.ClusterName

	dataPoints := make([]influxdb.Point, 0, 10)
	dataPoints = append(dataPoints, *point)
	sink.sendData(dataPoints)
}

// Generate point value for event
func getEventValue(event *v1.Event) (string, error) {
	bytes, err := json.MarshalIndent(event, "", " ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func eventToPointWithFields(event *v1.Event) (*influxdb.Point, error) {
	point := influxdb.Point{
		Measurement: "events",
		Time:        event.LastTimestamp.Time.UTC(),
		Fields: map[string]interface{}{
			"message": event.Message,
		},
		Tags: map[string]string{
			eventUID: string(event.UID),
		},
	}
	if event.InvolvedObject.Kind == "Pod" {
		point.Tags[LabelPodId.Key] = string(event.InvolvedObject.UID)
	}
	point.Tags["object_name"] = event.InvolvedObject.Name
	point.Tags["type"] = event.Type
	point.Tags["kind"] = event.InvolvedObject.Kind
	point.Tags["component"] = event.Source.Component
	point.Tags["reason"] = event.Reason
	point.Tags[LabelNamespaceName.Key] = event.Namespace
	point.Tags[LabelHostname.Key] = event.Source.Host
	return &point, nil
}

func eventToPoint(event *v1.Event) (*influxdb.Point, error) {
	value, err := getEventValue(event)
	if err != nil {
		return nil, err
	}

	point := influxdb.Point{
		Measurement: eventMeasurementName,
		Time:        event.LastTimestamp.Time.UTC(),
		Fields: map[string]interface{}{
			valueField: value,
		},
		Tags: map[string]string{
			eventUID: string(event.UID),
		},
	}
	if event.InvolvedObject.Kind == "Pod" {
		point.Tags[LabelPodId.Key] = string(event.InvolvedObject.UID)
		point.Tags[LabelPodName.Key] = event.InvolvedObject.Name
	}
	point.Tags[LabelHostname.Key] = event.Source.Host
	return &point, nil
}

func (sink *InfluxDBSink) sendData(dataPoints []influxdb.Point) {
	if err := sink.createDatabase(); err != nil {
		glog.Errorf("Failed to create influxdb: %v", err)
		return
	}
	bp := influxdb.BatchPoints{
		Points:          dataPoints,
		Database:        sink.config.DbName,
		RetentionPolicy: "default",
	}

	start := time.Now()
	if _, err := sink.client.Write(bp); err != nil {
		glog.Errorf("InfluxDB write failed: %v", err)
		if strings.Contains(err.Error(), dbNotFoundError) {
			sink.resetConnection()
		} else if _, _, err := sink.client.Ping(); err != nil {
			glog.Errorf("InfluxDB ping failed: %v", err)
			sink.resetConnection()
		}
	}
	end := time.Now()
	glog.V(4).Infof("Exported %d data to influxDB in %s", len(dataPoints), end.Sub(start))
}

func (sink *InfluxDBSink) resetConnection() {
	glog.Infof("Influxdb connection reset")
	sink.dbExists = false
	sink.client = nil
	sink.config = InfluxdbConfig{}
}

func (sink *InfluxDBSink) createDatabase() error {
	if sink.client == nil {
		client, err := newClient(sink.config)
		if err != nil {
			return err
		}
		sink.client = client
	}

	if sink.dbExists {
		return nil
	}

	q := influxdb.Query{
		Command: fmt.Sprintf(`CREATE DATABASE %s WITH NAME "default"`, sink.config.DbName),
	}

	if resp, err := sink.client.Query(q); err != nil {
		// We want to return error only if it is not "already exists" error.
		if !(resp != nil && resp.Err != nil && strings.Contains(resp.Err.Error(), "existing policy")) {
			err := sink.createRetentionPolicy()
			if err != nil {
				return err
			}
		}
	}

	sink.dbExists = true
	glog.Infof("Created database %q on influxDB server at %q", sink.config.DbName, sink.config.Host)
	return nil
}

func (sink *InfluxDBSink) createRetentionPolicy() error {
	q := influxdb.Query{
		Command: fmt.Sprintf(`CREATE RETENTION POLICY "default" ON %s DURATION 0d REPLICATION 1 DEFAULT`, sink.config.DbName),
	}

	if resp, err := sink.client.Query(q); err != nil {
		if !(resp != nil && resp.Err != nil && strings.Contains(resp.Err.Error(), "already exists")) {
			return fmt.Errorf("Retention policy creation failed: %v", err)
		}
	}

	glog.Infof("Created database %q on influxDB server at %q", sink.config.DbName, sink.config.Host)
	return nil
}
