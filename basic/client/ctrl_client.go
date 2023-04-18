package client

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	runtimecli "sigs.k8s.io/controller-runtime/pkg/client"
	// 假设定义了一个 foo CRD，里面包含 register.go 中定义了 AddToScheme 实例
	// foosv1 "pkg/apis/foos/v1"
)

func RuntimeClient(config *rest.Config) runtimecli.Client {
	client, err := runtimecli.New(config, runtimecli.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		panic(err)
	}
	return client
}

func RuntimeClientForCRD(config *rest.Config) runtimecli.Client {
	crScheme := runtime.NewScheme()
	// foosv1.AddToScheme(scheme.Scheme)
	client, err := runtimecli.New(config, runtimecli.Options{
		Scheme: crScheme,
	})
	if err != nil {
		panic(err)
	}
	return client
}
