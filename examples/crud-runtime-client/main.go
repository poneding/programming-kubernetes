package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	// 从本地 kubeconfig 文件中加载配置
	// config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	config := ctrl.GetConfigOrDie()

	cli, err := client.New(config, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		panic(err.Error())
	}

	namespace := "default"
	desired := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "runtime-client-crud-config",
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}

	key := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}

	// 创建 ConfigMap
	err = cli.Create(context.Background(), desired)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Created ConfigMap %s/%s\n", namespace, desired.GetName())
	fmt.Printf("desired.ResourceVersion: %v\n", desired.ResourceVersion)

	// 获取 ConfigMap
	got := &corev1.ConfigMap{}
	err = cli.Get(context.Background(), key, got)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Got ConfigMap %s/%s\n", namespace, got.GetName())

	// 获取 ConfigMap 列表
	list := &corev1.ConfigMapList{}
	err = cli.List(context.Background(), list, &client.ListOptions{Namespace: namespace})
	if err != nil {
		panic(err.Error())
	}
	for _, item := range list.Items {
		fmt.Printf("List ConfigMap %s/%s\n", namespace, item.GetName())
	}

	// 更新 ConfigMap
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// 获取 ConfigMap
		got := &corev1.ConfigMap{}
		err = cli.Get(context.Background(), key, got)
		if err != nil {
			return err
		}

		// 更新 ConfigMap
		got.Data["foo"] = "bar2"
		err = cli.Update(context.Background(), got)
		if err != nil {
			return err
		}
		fmt.Printf("Updated ConfigMap %s/%s\n", namespace, got.GetName())
		return nil
	})
	if err != nil {
		panic(err.Error())
	}

	// 删除 ConfigMap
	err = cli.Delete(context.Background(), desired)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Deleted ConfigMap %s/%s\n", namespace, desired.GetName())
}
