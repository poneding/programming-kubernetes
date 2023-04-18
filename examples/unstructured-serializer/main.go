package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func main() {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-cm",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"foo": "bar",
			},
		},
	}

	// 非结构化对象 => JSON
	// 第一种方式
	encoded1, err := runtime.Encode(unstructured.UnstructuredJSONScheme, obj)
	if err != nil {
		panic(err)
	}
	fmt.Printf("encoded1: %s\n", string(encoded1))
	fmt.Println()

	// 第二种方式：unstructured.UnstructuredJSONScheme 对象实现了 MarshalJSON 方式
	encoded2, err := obj.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Printf("encoded2: %s\n", string(encoded2))
	fmt.Println()

	// JSON => 非结构化对象
	decoded, err := runtime.Decode(unstructured.UnstructuredJSONScheme, encoded1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded: %v\n", decoded)
	fmt.Println()
}
