package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func main() {
	// 运行
	// go run main.go
	// kubectl apply -f cm.yaml

	config := controllerruntime.GetConfigOrDie()
	client := kubernetes.NewForConfigOrDie(config)

	watch, err := client.CoreV1().ConfigMaps("default").Watch(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: "foo=bar",
			// Watch:         true, // 即使这里不设置，调用 Watch 也会将其设置为 true
			TimeoutSeconds: &[]int64{30}[0], // 持续 Watch 的时间
			Limit:          10,              // 一次 Watch 的最多返回的资源数量
		},
	)
	if err != nil {
		panic(err)
	}

	stop := make(chan struct{})
	go func() {
		for event := range watch.ResultChan() {
			obj, ok := event.Object.(*corev1.ConfigMap)
			if !ok {
				fmt.Printf("event: %s %s\n", event.Object.GetObjectKind().GroupVersionKind().Kind, event.Type)
				continue
			}
			fmt.Printf("event: %s/%s %s\n", obj.GetNamespace(), obj.GetName(), event.Type)
		}
		stop <- struct{}{}
	}()

	<-stop
}
