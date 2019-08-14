package rockset

import (
	"log"
	"net/http"
	"strings"

	api "github.com/rockset/rockset-go-client/lib/go"
)

var Version="0.6.0"

type RockClient struct {
	apiServer string
	common    api.Service

	// API Services
	ApiKeys     *api.ApiKeysApiService
	Collection  *api.CollectionsApiService
	Integration *api.IntegrationsApiService
	Documents   *api.DocumentsApiService
	QueryApi    *api.QueriesApiService
	Users       *api.UsersApiService
}

/*
Create a Client object to securely connect to Rockset using an API key
Optionally, an alternate API server host can also be provided.
*/
func Client(apiKey string, apiServer string) *RockClient {
	// TODO read from credentials file if it exists
	if apiKey == "" {
		log.Fatal("apiKey needs to be specified")
	}

	if apiServer == "" {
		apiServer = "https://api.rs2.usw2.rockset.com"
	}

	if !strings.HasPrefix(apiServer, "http://") && !strings.HasPrefix(apiServer, "https://") {
		apiServer = "https://" + apiServer
	}

	c := &RockClient{}
	cfg := api.NewConfiguration()
	cfg.BasePath = apiServer
	cfg.Version = Version
	c.common.Client = api.ApiClient(cfg, apiKey)

	// API Services
	c.ApiKeys = (*api.ApiKeysApiService)(&c.common)
	c.Collection = (*api.CollectionsApiService)(&c.common)
	c.Integration = (*api.IntegrationsApiService)(&c.common)
	c.Documents = (*api.DocumentsApiService)(&c.common)
	c.QueryApi = (*api.QueriesApiService)(&c.common)
	c.Users = (*api.UsersApiService)(&c.common)

	return c
}

// Execute a query against Rockset
func (c *RockClient) Query(body api.QueryRequest) (api.QueryResponse, *http.Response, error) {
	return c.QueryApi.Query(body)
}
