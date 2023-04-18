package main

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func main() {
	// 假设你从某个地方获取到了一个 runtime.Object 对象
	// 你想从这个对象中获取到它的 namespace 和 name
	// 你可以使用 meta.Accessor 方法来获取到对象的元数据
	var ro runtime.Object = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}

	// 从 runtime.Object 对象中获取到元数据
	o, err := meta.Accessor(ro)
	if err != nil {
		panic(err)
	}
	fmt.Printf("o.GetNamespace(): %v\n", o.GetNamespace())
	fmt.Printf("o.GetName(): %v\n", o.GetName())
	fmt.Printf("o.GetLabels(): %v\n", o.GetLabels())
}

/*
$ go run main.go
o.GetNamespace(): default
o.GetName(): test-cm
o.GetLabels(): map[app:test]
*/
