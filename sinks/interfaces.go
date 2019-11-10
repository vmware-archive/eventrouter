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
	v1 "k8s.io/api/core/v1"
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
		viper.SetDefault("stdoutJSONNamespace", "")
		stdoutNamespace := viper.GetString("stdoutJSONNamespace")
		e = NewStdoutSink(stdoutNamespace)
	case "http":
		url := viper.GetString("httpSinkUrl")
		if url == "" {
			panic("http sink specified but no httpSinkUrl")
		}

		// By default we buffer up to 1500 events, and drop messages if more than
		// 1500 have come in without getting consumed
		viper.SetDefault("httpSinkBufferSize", 1500)
		viper.SetDefault("httpSinkDiscardMessages", true)

		bufferSize := viper.GetInt("httpSinkBufferSize")
		overflow := viper.GetBool("httpSinkDiscardMessages")

		h := NewHTTPSink(url, overflow, bufferSize)
		go h.Run(make(chan bool))
		return h
	case "kafka":
		viper.SetDefault("kafkaBrokers", []string{"kafka:9092"})
		viper.SetDefault("kafkaTopic", "eventrouter")
		viper.SetDefault("kafkaAsync", true)
		viper.SetDefault("kafkaRetryMax", 5)
		viper.SetDefault("kafkaSaslUser", "")
		viper.SetDefault("kafkaSaslPwd", "")

		brokers := viper.GetStringSlice("kafkaBrokers")
		topic := viper.GetString("kafkaTopic")
		async := viper.GetBool("kakfkaAsync")
		retryMax := viper.GetInt("kafkaRetryMax")
		saslUser := viper.GetString("kafkaSaslUser")
		saslPwd := viper.GetString("kafkaSaslPwd")

		e, err := NewKafkaSink(brokers, topic, async, retryMax, saslUser, saslPwd)
		if err != nil {
			panic(err.Error())
		}
		return e
	case "s3sink":
		accessKeyID := viper.GetString("s3SinkAccessKeyID")
		if accessKeyID == "" {
			panic("s3 sink specified but s3SinkAccessKeyID not specified")
		}

		secretAccessKey := viper.GetString("s3SinkSecretAccessKey")
		if secretAccessKey == "" {
			panic("s3 sink specified but s3SinkSecretAccessKey not specified")
		}

		region := viper.GetString("s3SinkRegion")
		if region == "" {
			panic("s3 sink specified but s3SinkRegion not specified")
		}

		bucket := viper.GetString("s3SinkBucket")
		if bucket == "" {
			panic("s3 sink specified but s3SinkBucket not specified")
		}

		bucketDir := viper.GetString("s3SinkBucketDir")
		if bucketDir == "" {
			panic("s3 sink specified but s3SinkBucketDir not specified")
		}

		// By default the json is pushed to s3 in not flatenned rfc5424 write format
		// The option to write to s3 is in the flattened json format which will help in
		// using the data in redshift with least effort
		viper.SetDefault("s3SinkOutputFormat", "rfc5424")
		outputFormat := viper.GetString("s3SinkOutputFormat")
		if outputFormat != "rfc5424" && outputFormat != "flatjson" {
			panic("s3 sink specified, but incorrect s3SinkOutputFormat specifed. Supported formats are: rfc5424 (default) and flatjson")
		}

		// By default we buffer up to 1500 events, and drop messages if more than
		// 1500 have come in without getting consumed
		viper.SetDefault("s3SinkBufferSize", 1500)
		viper.SetDefault("s3SinkDiscardMessages", true)

		viper.SetDefault("s3SinkUploadInterval", 120)
		uploadInterval := viper.GetInt("s3SinkUploadInterval")

		bufferSize := viper.GetInt("s3SinkBufferSize")
		overflow := viper.GetBool("s3SinkDiscardMessages")

		s, err := NewS3Sink(accessKeyID, secretAccessKey, region, bucket, bucketDir, uploadInterval, overflow, bufferSize, outputFormat)
		if err != nil {
			panic(err.Error())
		}

		go s.Run(make(chan bool))
		return s
	case "influxdb":
		host := viper.GetString("influxdbHost")
		if host == "" {
			panic("influxdb sink specified but influxdbHost not specified")
		}

		username := viper.GetString("influxdbUsername")
		if username == "" {
			panic("influxdb sink specified but influxdbUsername not specified")
		}

		password := viper.GetString("influxdbPassword")
		if password == "" {
			panic("influxdb sink specified but influxdbPassword not specified")
		}

		viper.SetDefault("influxdbName", "k8s")
		viper.SetDefault("influxdbSecure", false)
		viper.SetDefault("influxdbWithFields", false)
		viper.SetDefault("influxdbInsecureSsl", false)
		viper.SetDefault("influxdbRetentionPolicy", "0")
		viper.SetDefault("influxdbClusterName", "default")
		viper.SetDefault("influxdbDisableCounterMetrics", false)
		viper.SetDefault("influxdbConcurrency", 1)

		dbName := viper.GetString("influxdbName")
		secure := viper.GetBool("influxdbSecure")
		withFields := viper.GetBool("influxdbWithFields")
		insecureSsl := viper.GetBool("influxdbInsecureSsl")
		retentionPolicy := viper.GetString("influxdbRetentionPolicy")
		cluterName := viper.GetString("influxdbClusterName")
		disableCounterMetrics := viper.GetBool("influxdbDisableCounterMetrics")
		concurrency := viper.GetInt("influxdbConcurrency")

		cfg := InfluxdbConfig{
			User:                  username,
			Password:              password,
			Secure:                secure,
			Host:                  host,
			DbName:                dbName,
			WithFields:            withFields,
			InsecureSsl:           insecureSsl,
			RetentionPolicy:       retentionPolicy,
			ClusterName:           cluterName,
			DisableCounterMetrics: disableCounterMetrics,
			Concurrency:           concurrency,
		}

		influx, err := NewInfuxdbSink(cfg)
		if err != nil {
			panic(err.Error())
		}
		return influx
	case "rockset":
		rocksetAPIKey := viper.GetString("rocksetAPIKey")
		if rocksetAPIKey == "" {
			panic("Rockset sink specified but rocksetAPIKey not specified")
		}

		rocksetCollectionName := viper.GetString("rocksetCollectionName")
		if rocksetCollectionName == "" {
			panic("Rockset sink specified but rocksetCollectionName not specified")
		}
		rocksetWorkspaceName := viper.GetString("rocksetWorkspaceName")
		if rocksetCollectionName == "" {
			panic("Rockset sink specified but rocksetWorkspaceName not specified")
		}
		e = NewRocksetSink(rocksetAPIKey, rocksetCollectionName, rocksetWorkspaceName)
	case "eventhub":
		connString := viper.GetString("eventHubConnectionString")
		if connString == "" {
			panic("eventhub sink specified but eventHubConnectionString not specified")
		}
		// By default we buffer up to 1500 events, and drop messages if more than
		// 1500 have come in without getting consumed
		viper.SetDefault("eventHubSinkBufferSize", 1500)
		viper.SetDefault("eventHubSinkDiscardMessages", true)

		bufferSize := viper.GetInt("eventHubSinkBufferSize")
		overflow := viper.GetBool("eventHubSinkDiscardMessages")
		eh, err := NewEventHubSink(connString, overflow, bufferSize)
		if err != nil {
			panic(err.Error())
		}
		go eh.Run(make(chan bool))
		return eh
	// case "logfile"
	default:
		err := errors.New("Invalid Sink Specified")
		panic(err.Error())
	}
	return e
}
