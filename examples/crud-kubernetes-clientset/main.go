package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

func main() {
	// 从本地 kubeconfig 文件中加载配置
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err.Error())
	}

	// 创建 kubernetes.Clientset
	client := kubernetes.NewForConfigOrDie(config)

	namespace := "default"
	desired := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "kubernetes-clientset-crud-config",
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}

	// 创建 ConfigMap
	created, err := client.CoreV1().ConfigMaps(namespace).Create(context.Background(), desired, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Created ConfigMap %s/%s\n", namespace, created.GetName())

	// 获取 ConfigMap
	got, err := client.CoreV1().ConfigMaps(namespace).Get(context.Background(), created.GetName(), metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Got ConfigMap %s/%s\n", namespace, got.GetName())

	// 获取 ConfigMap 列表
	list, err := client.CoreV1().ConfigMaps(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, item := range list.Items {
		fmt.Printf("List ConfigMap %s/%s\n", namespace, item.GetName())
	}

	// 更新 ConfigMap
	// got.Data["foo"] = "baz"
	// updated, err := client.CoreV1().ConfigMaps(namespace).Update(context.Background(), got, metav1.UpdateOptions{})
	// if err != nil {
	// 	panic(err.Error())
	// }

	// 在官方的示例中，建议了一种更新资源的方式：当更新资源时，有可能会发生冲突而导致更新失败，这时候需要重新获取资源，然后再次更新。
	// 冲突的原因：可能你获取到的资源，已经被其他人更新或者删除了
	// 底层通过 ResourceVersion 来判断资源是否发生了变化
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// 获取 ConfigMap
		got, getErr := client.CoreV1().ConfigMaps(namespace).Get(context.Background(), created.GetName(), metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("failed to get latest version of ConfigMap: %v", getErr))
		}

		got.Data["foo"] = "baz"
		_, updateErr := client.CoreV1().ConfigMaps(namespace).Update(context.Background(), got, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}

	fmt.Printf("Updated ConfigMap %s/%s\n", namespace, created.GetName())

	// 删除 ConfigMap
	err = client.CoreV1().ConfigMaps(namespace).Delete(context.Background(), created.GetName(), metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Deleted ConfigMap %s/%s\n", namespace, created.GetName())
}
