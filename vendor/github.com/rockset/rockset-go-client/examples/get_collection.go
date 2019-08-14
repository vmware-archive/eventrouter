package main

import (
	"fmt"
	"os"

	apiclient "github.com/rockset/rockset-go-client"
)

func main() {
	apiKey := os.Getenv("ROCKSET_APIKEY")
	apiServer := os.Getenv("ROCKSET_APISERVER")

	// create the API client
	client := apiclient.Client(apiKey, apiServer)

	// get collections
	res, _, err := client.Collection.Get("_events")
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	res.PrintResponse()
}
