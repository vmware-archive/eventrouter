package sinks

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"k8s.io/api/core/v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/eapache/channels"
	"github.com/golang/glog"
)

/*
S3Sink is the sink that uploads the kubernetes events as json object stored in a file.
The sinker uploads it to s3 if any of the below criteria gets fullfilled
1) Time(uploadInterval): If the specfied time has passed since the last upload it uploads
2) [TODO] Data size: If the total data getting uploaded becomes greater than N bytes

S3 is cheap and the sink can be used to store events data. S3 can later then be used with
Redshift and other visualization tools to use this data.

*/
type S3Sink struct {
	// uploader is the uploader client from aws which makes the API call to aws for upload
	uploader *s3manager.Uploader

	// bucket is the s3 bucket name where the events data would be stored
	bucket string

	// bucketDir is the first level directory in the bucket where the events would be stored
	bucketDir string

	// outPutFormat is the format in which the data is stored in the s3 file
	outputFormat string

	// lastUploadTimestamp stores the timestamp when the last upload to s3 happened
	lastUploadTimestamp int64

	// uploadInterval tells after how many seconds the next upload can happen
	// sink waits till this time is passed before next upload can happen
	uploadInterval time.Duration

	// eventCh is used to interact eventRouter and the sharedInformer
	eventCh channels.Channel

	// bodyBuf stores all the event captured data in a buffer before upload
	bodyBuf *bytes.Buffer
}

// NewS3Sink is the factory method constructing a new S3Sink
func NewS3Sink(awsAccessKeyID string, s3SinkSecretAccessKey string, s3SinkRegion string, s3SinkBucket string, s3SinkBucketDir string, s3SinkUploadInterval int, overflow bool, bufferSize int, outputFormat string) (*S3Sink, error) {
	awsConfig := &aws.Config{
		Region:      aws.String(s3SinkRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, s3SinkSecretAccessKey, ""),
	}

	awsConfig = awsConfig.WithCredentialsChainVerboseErrors(true)
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	uploader := s3manager.NewUploader(sess)

	s := &S3Sink{
		uploader:       uploader,
		bucket:         s3SinkBucket,
		bucketDir:      s3SinkBucketDir,
		uploadInterval: time.Second * time.Duration(s3SinkUploadInterval),
		outputFormat:   outputFormat,
		bodyBuf:        bytes.NewBuffer(make([]byte, 0, 4096)),
	}

	if overflow {
		s.eventCh = channels.NewOverflowingChannel(channels.BufferCap(bufferSize))
	} else {
		s.eventCh = channels.NewNativeChannel(channels.BufferCap(bufferSize))
	}

	return s, nil
}

// UpdateEvents implements the EventSinkInterface. It really just writes the
// event data to the event OverflowingChannel, which should never block.
// Messages that are buffered beyond the bufferSize specified for this HTTPSink
// are discarded.
func (s *S3Sink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	s.eventCh.In() <- NewEventData(eNew, eOld)
}

// Run sits in a loop, waiting for data to come in through h.eventCh,
// and forwarding them to the HTTP sink. If multiple events have happened
// between loop iterations, it puts all of them in one request instead of
// making a single request per event.
func (s *S3Sink) Run(stopCh <-chan bool) {
loop:
	for {
		select {
		case e := <-s.eventCh.Out():
			var evt EventData
			var ok bool
			if evt, ok = e.(EventData); !ok {
				glog.Warningf("Invalid type sent through event channel: %T", e)
				continue loop
			}

			// Start with just this event...
			arr := []EventData{evt}

			// Consume all buffered events into an array, in case more have been written
			// since we last forwarded them
			numEvents := s.eventCh.Len()
			for i := 0; i < numEvents; i++ {
				e := <-s.eventCh.Out()
				if evt, ok = e.(EventData); ok {
					arr = append(arr, evt)
				} else {
					glog.Warningf("Invalid type sent through event channel: %T", e)
				}
			}

			s.drainEvents(arr)
		case <-stopCh:
			break loop
		}
	}
}

// drainEvents takes an array of event data and sends it to s3
func (s *S3Sink) drainEvents(events []EventData) {
	var written int64
	for _, evt := range events {
		switch s.outputFormat {
		case "rfc5424":
			w, err := evt.WriteRFC5424(s.bodyBuf)
			written += w
			if err != nil {
				glog.Warningf("Could not write to event request body (wrote %v) bytes: %v", written, err)
				return
			}
		case "flatjson":
			w, err := evt.WriteFlattenedJSON(s.bodyBuf)
			written += w
			if err != nil {
				glog.Warningf("Could not write to event request body (wrote %v) bytes: %v", written, err)
				return
			}
		default:
			err := errors.New("Invalid Sink Output Format specified")
			panic(err.Error())
		}
		s.bodyBuf.Write([]byte{'\n'})
		written++
	}

	if s.canUpload() == false {
		return
	}

	s.upload()
}

// canUpload verifies the conditions suitable for a new file upload and upload the data
func (s *S3Sink) canUpload() bool {
	now := time.Now().UnixNano()
	if (s.lastUploadTimestamp + s.uploadInterval.Nanoseconds()) < now {
		return true
	}
	return false
}

// getNewKey gets the key name based on time
func (s *S3Sink) getNewKey(t time.Time) string {
	return fmt.Sprintf("%s/%d/%d/%d/%d.txt", s.bucketDir, t.Year(), t.Month(), t.Day(), t.UnixNano())
}

// upload uploads the events stored in buffer to s3 in the specified key
// and clears the buffer
func (s *S3Sink) upload() {
	now := time.Now()
	key := s.getNewKey(now)

	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   s.bodyBuf,
	})
	if err != nil {
		glog.Errorf("Error uploading %s to s3, %v", key, err)
	}
	glog.Infof("Uploaded at %s", key)
	s.lastUploadTimestamp = now.UnixNano()

	s.bodyBuf.Truncate(0)
}
