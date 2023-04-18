package client

import "k8s.io/client-go/rest"

func RESTClient(config *rest.Config) *rest.RESTClient {
	client, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}
	return client
}
