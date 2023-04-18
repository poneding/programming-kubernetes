package main

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

// 运行：
// go run main.go
// kubectl run nginx --image=nginx --labels=foo=bar
// kubectl delete pod nginx

func main() {
	// 创建 store 对象
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)

	// 创建 clientset 对象
	clientset := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())

	// 创建 ListWatch 对象
	lw := cache.NewFilteredListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", "default", func(options *metav1.ListOptions) {
		options.LabelSelector = "foo=bar"
	})

	// 创建 DeltaFIFO 对象
	df := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KeyFunction:  cache.MetaNamespaceKeyFunc,
		KnownObjects: store,
	})

	// 创建 Reflector 对象
	rf := cache.NewReflector(lw, &corev1.Pod{}, df, 0)

	stop := make(chan struct{})
	defer close(stop)

	go func() {
		// 不断的使用 listwatch 从 DeltaFIFO 中取出对象和后续增量
		rf.Run(stop)
	}()

	for {
		popFunc := func(pop interface{}) error {
			for _, delta := range pop.(cache.Deltas) {
				pod := delta.Object.(*corev1.Pod)
				key, _ := df.KeyOf(pod)
				fmt.Printf("%s: %s - %s\n", delta.Type, key, pod.Status.Phase)
				// 根据 delta.Type 的类型，更新 store 中的对象
				switch delta.Type {
				case cache.Added, cache.Sync:
					store.Add(pod)
				case cache.Updated:
					store.Update(pod)
				case cache.Deleted:
					store.Delete(pod)
				}
			}
			return nil
		}
		_, err := df.Pop(popFunc)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
	}
}
