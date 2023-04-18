package config

import (
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

func KubeConfigFromCtrlRuntime() *rest.Config {
	config, err := ctrl.GetConfig()
	if err != nil {
		panic(err)
	}
	return config
}
