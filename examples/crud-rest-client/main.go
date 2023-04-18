package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 从本地 kubeconfig 文件中加载配置
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err.Error())
	}

	// 当初始化 rest.RESTClient 时，需要指定 GroupVersion 和 NegotiatedSerializer
	config.APIPath = "/api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs

	config.GroupVersion = &corev1.SchemeGroupVersion
	// 创建 rest.RESTClient
	client, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}

	namespace := "default"
	desired := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "rest-client-crud-config",
			Labels: map[string]string{
				"app": "rest-client-crud",
			},
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}

	// 创建 ConfigMap
	// 使用 Into() 方法接收返回的结果，需要传入一个非 nil 的指针
	// var created *corev1.ConfigMap // 如果传入 nil 的指针，会 panic: expected pointer, but got nil
	var created = &corev1.ConfigMap{} // 非 nil 的指针
	err = client.Post().Namespace(namespace).Resource("configmaps").Body(desired).Do(context.Background()).Into(created)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Created ConfigMap %s/%s\n", namespace, created.GetName())

	// 获取 ConfigMap
	var got = &corev1.ConfigMap{}
	err = client.Get().Namespace(namespace).Resource("configmaps").Name(created.GetName()).Do(context.Background()).Into(got)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Got ConfigMap %s/%s\n", namespace, got.GetName())

	// 获取 ConfigMap 列表
	var list = &corev1.ConfigMapList{}
	err = client.Get().Namespace(namespace).Resource("configmaps").Do(context.Background()).Into(list)
	// 可以通过 Param() 方法传入查询参数：labelSelector，fieldSelector，resourceVersion，timeoutSeconds，watch 等，例如：
	// err = client.Get().Namespace(namespace).Resource("configmaps").Param("labelSelector", "app=rest-client-crud").Do(context.Background()).Into(list)
	// 通过 label 过滤

	if err != nil {
		panic(err.Error())
	}
	for _, item := range list.Items {
		fmt.Printf("List ConfigMap %s/%s\n", namespace, item.GetName())
	}

	// 更新 ConfigMap
	got.Data["foo"] = "baz"
	var updated = &corev1.ConfigMap{}
	err = client.Put().Namespace(namespace).Resource("configmaps").Name(created.GetName()).Body(got).Do(context.Background()).Into(updated)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Updated ConfigMap %s/%s\n", namespace, updated.GetName())

	// 删除 ConfigMap
	err = client.Delete().Namespace(namespace).Resource("configmaps").Name(created.GetName()).Do(context.Background()).Error()
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Deleted ConfigMap %s/%s\n", namespace, created.GetName())
}
