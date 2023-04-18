package config

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func KubeConfigFromFlags() *rest.Config {
	// config, err := clientcmd.BuildConfigFromFlags("", path.Join(clientcmd.RecommendedHomeDir, ".kube", "config"))
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}
	return config
}
