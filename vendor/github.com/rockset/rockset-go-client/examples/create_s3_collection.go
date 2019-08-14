package main

import (
	"fmt"
	"os"

	apiclient "github.com/rockset/rockset-go-client"
	models "github.com/rockset/rockset-go-client/lib/go"
)

func main() {
	apiKey := os.Getenv("ROCKSET_APIKEY")
	apiServer := os.Getenv("ROCKSET_APISERVER")

	// create the API client
	client := apiclient.Client(apiKey, apiServer)

	// create collection
	cinfo := models.CreateCollectionRequest{
		Name:        "s3_collection",
		Description: "s3 collection",
		Sources: []models.Source{
			{
				IntegrationName: "my-first-integration",
				S3: &models.SourceS3{
					Bucket: "<bucket>",
				},
			},
		},
	}

	rockresponse, _, err := client.Collection.Create(cinfo)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	rockresponse.PrintResponse()
}
