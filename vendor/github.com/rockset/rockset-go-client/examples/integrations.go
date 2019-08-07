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

	// create integration
	iinfo := models.CreateIntegrationRequest{
		Name:        "my-first-integration",
		Description: "my first integration",
		Aws: &models.AwsKeyIntegration{
			AwsAccessKeyId:     ".....",
			AwsSecretAccessKey: ".....",
		},
	}

	res, _, err := client.Integration.Create(iinfo)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	res.PrintResponse()
}
