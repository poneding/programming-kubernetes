package client

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func DynamicClient(config *rest.Config) dynamic.Interface {
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return client
}
