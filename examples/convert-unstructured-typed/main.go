package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func main() {
	// 定义一个非结构化的 ConfigMap 实例
	unstructuredConfigMap := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"creationTimestamp": nil,
				"namespace":         "default",
				"name":              "my-configmap",
			},
			"data": map[string]interface{}{
				"foo": "bar",
			},
		},
	}

	// 非结构化对象 => 结构化对象
	var structuredConfigMap corev1.ConfigMap
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredConfigMap.Object, &structuredConfigMap)
	if err != nil {
		panic(err.Error())
	}
	if structuredConfigMap.GetName() != "my-configmap" {
		panic("name not equal")
	}

	// 结构化对象 => 非结构化对象
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&structuredConfigMap)
	if err != nil {
		panic(err.Error())
	}

	j1, _ := json.Marshal(unstructuredConfigMap)
	j2, _ := json.Marshal(unstructured.Unstructured{Object: unstructuredObj})
	fmt.Printf("j1: %v\n", string(j1))
	fmt.Printf("j2: %v\n", string(j2))

	if !reflect.DeepEqual(unstructuredConfigMap, unstructured.Unstructured{Object: unstructuredObj}) {
		panic("not equal")
	}
}
