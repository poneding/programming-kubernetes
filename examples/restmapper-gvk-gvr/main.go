package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 从本地 kubeconfig 文件中加载配置
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err.Error())
	}

	// 创建 discovery.DiscoveryClient
	client := discovery.NewDiscoveryClientForConfigOrDie(config)

	// 创建 restmapper.RESTMapper
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(client))

	// 1. GVK -> GVR
	gvk := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "ConfigMap",
	}
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		panic(err.Error())
	}
	gvr := mapping.Resource
	fmt.Printf("获取到 GVR: %#v\n", gvr)

	// 2. GVR -> GVK
	gvk, err = mapper.KindFor(gvr)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("获取到 GVK: %#v\n", gvk)

	//
	//
	// 按照 dynameic-client-crud 的实践，我们就可以通过 unstructured.Unstructured 来创建 ConfigMap
	namespace := "default"
	desired := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      "dynamic-client-crud-config",
			},
			"data": map[string]interface{}{
				"foo": "bar",
			},
		},
	}

	dynamicClient := dynamic.NewForConfigOrDie(config)

	// 创建 ConfigMap
	created, err := dynamicClient.Resource(mapping.Resource).Namespace(namespace).Create(context.Background(), desired, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Created ConfigMap %s/%s\n", namespace, created.GetName())

	// 删除 ConfigMap
	err = dynamicClient.Resource(mapping.Resource).Namespace(namespace).Delete(context.Background(), created.GetName(), metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Deleted ConfigMap %s/%s\n", namespace, created.GetName())
}
