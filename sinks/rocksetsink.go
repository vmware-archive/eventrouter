/*
Copyright 2019 The Contributors.

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
	"os"

	apiclient "github.com/rockset/rockset-go-client"
	models "github.com/rockset/rockset-go-client/lib/go"
	v1 "k8s.io/api/core/v1"
)

/*
RocksetSink is a sink that uploads the kubernetes events as json object
and converts them to documents inside of a Rockset collection.

Rockset can later be used with
many different connectors such as Tableau or Redash to use this data.
*/
type RocksetSink struct {
	client                *apiclient.RockClient
	rocksetCollectionName string
	rocksetWorkspaceName  string
}

// NewRocksetSink will create a new RocksetSink with default options, returned as
// an EventSinkInterface
func NewRocksetSink(rocksetAPIKey string, rocksetCollectionName string, rocksetWorkspaceName string) EventSinkInterface {
	client := apiclient.Client(rocksetAPIKey, "https://api.rs2.usw2.rockset.com")
	return &RocksetSink{
		client:                client,
		rocksetCollectionName: rocksetCollectionName,
		rocksetWorkspaceName:  rocksetWorkspaceName,
	}
}

// UpdateEvents implements the EventSinkInterface
func (rs *RocksetSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	eData := NewEventData(eNew, eOld)

	if eJSONBytes, err := json.Marshal(eData); err == nil {
		var m map[string]interface{}
		json.Unmarshal(eJSONBytes, &m)
		docs := []interface{}{
			m,
		}
		dinfo := models.AddDocumentsRequest{
			Data: docs,
		}
		rs.client.Documents.Add(rs.rocksetWorkspaceName, rs.rocksetCollectionName, dinfo)
	} else {
		fmt.Fprintf(os.Stderr, "Failed to json serialize event: %v", err)
	}
}
