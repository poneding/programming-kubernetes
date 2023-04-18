package client

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

func DiscoveryClient(config *rest.Config) *discovery.DiscoveryClient {
	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err)
	}
	return client
}
