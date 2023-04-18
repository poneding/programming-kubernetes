package main

import (
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

func main() {
	obj := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm",
			Namespace: "default",
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}

	// 对象实例 => JSON
	// 第一种方式
	encoder := jsonserializer.NewSerializerWithOptions(
		nil, // jsonserializer.MetaFactory
		nil, // runtime.ObjectCreator
		nil, // runtime.ObjectTyper
		jsonserializer.SerializerOptions{
			Yaml:   false, // 是否以 yaml 格式输出
			Pretty: false, // 是否以 pretty 格式输出, 仅在 Yaml 为 false 时有效
			Strict: false, // 是否以 strict 格式输出
		},
	)

	encoded1, err := runtime.Encode(encoder, obj)
	if err != nil {
		panic(err)
	}
	fmt.Printf("encoded1: \n%s\n", string(encoded1))
	fmt.Println()

	// 第二种方式
	encoded2, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	fmt.Printf("encoded2: \n%s\n", string(encoded2))
	fmt.Println()

	// JSON => 对象实例
	encoder = jsonserializer.NewSerializerWithOptions(
		jsonserializer.DefaultMetaFactory,
		scheme.Scheme,
		scheme.Scheme,
		jsonserializer.SerializerOptions{
			Yaml:   false,
			Pretty: false,
			Strict: false,
		})
	decoded, err := runtime.Decode(encoder, encoded1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded: %#v\n", decoded)
}
