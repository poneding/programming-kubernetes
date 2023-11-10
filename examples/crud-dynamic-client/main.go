package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"maps"
)

func main() {
	// 从本地 kubeconfig 文件中加载配置
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err.Error())
	}

	// 创建 dynamic.Client
	client := dynamic.NewForConfigOrDie(config)

	namespace := "default"
	// 用 schema.GroupVersionResource 描述资源（GVR）
	// 目的是为了让 dynamic.Client 可以通过 GVR 来识别资源，找到对应的 endpoint
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}

	// 以非结构化的方式创建 ConfigMap
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

	// 创建 ConfigMap
	created, err := client.Resource(res).Namespace(namespace).Create(context.Background(), desired, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Created ConfigMap %s/%s\n", namespace, created.GetName())

	// 获取 ConfigMap
	got, err := client.Resource(res).Namespace(namespace).Get(context.Background(), created.GetName(), metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Got ConfigMap %s/%s\n", namespace, got.GetName())

	// 获取 ConfigMap 的列表
	list, err := client.Resource(res).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, item := range list.Items {
		fmt.Printf("List ConfigMap %s/%s\n", namespace, item.GetName())
	}

	// 检查 ConfigMap 的数据
	// 如果没有找到 map 的键，则第二个返回值为 false
	// 如果传入的 obj 结构不是 map[string]interface{} 或者 map 值类型不是 string map，则第三个返回值为 error
	data, _, _ := unstructured.NestedStringMap(got.Object, "data")
	if !maps.Equal(data, map[string]string{"foo": "bar"}) {
		panic("Got ConfigMap has unexpected data")
	}

	// 更新 ConfigMap
	// SetNestedField 函数第二个参数是要修改的值，第二个参数之后是要修改的字段路径
	// 例如：SetNestedField(got.Object, "baz", "data", "foo")，表示将 got.Object["data"]["foo"] 的值修改为 "baz"
	unstructured.SetNestedField(got.Object, "baz", "data", "foo")
	updated, err := client.Resource(res).Namespace(namespace).Update(context.Background(), got, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Updated ConfigMap %s/%s\n", namespace, updated.GetName())
	data, _, _ = unstructured.NestedStringMap(updated.Object, "data")
	if !maps.Equal(data, map[string]string{"foo": "baz"}) {
		panic("Got ConfigMap has unexpected data")
	}

	// 删除 ConfigMap
	err = client.Resource(res).Namespace(namespace).Delete(context.Background(), updated.GetName(), metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Deleted ConfigMap %s/%s\n", namespace, updated.GetName())

	// 自定义资源操作，需要集群中已经存在该自定义资源
	// get, err := client.Resource(schema.GroupVersionResource{
	// 	Group:    "kubevirt.io",
	// 	Version:  "v1",
	// 	Resource: "virtualmachines",
	// }).Namespace(metav1.NamespaceDefault).Get(context.Background(), "testvm", metav1.GetOptions{})
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("get: %v\n", get.GetName())
}
