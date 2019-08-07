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

	// construct query
	q := models.QueryRequest{
		Sql: &models.QueryRequestSql{
			Query: "select * from \"_events\" limit 1",
		},
	}

	// query
	res, _, err := client.Query(q)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	// print result
	res.PrintResponse()
}
