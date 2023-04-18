package config

import (
	"k8s.io/client-go/rest"
)

func KubeConfigFromInClusterConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	return config
}
