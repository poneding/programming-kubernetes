package client

import (
	"net/http"

	"k8s.io/client-go/rest"
)

func HTTPClient(config *rest.Config) *http.Client {
	client, err := rest.HTTPClientFor(config)
	if err != nil {
		panic(err)
	}
	return client
}
