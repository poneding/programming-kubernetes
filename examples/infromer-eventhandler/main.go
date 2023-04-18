package main

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

// 运行：
// go run main.go
// kubectl apply -f cm.yaml
// kubectl delete -f cm.yaml

func main() {
	clientset := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())

	lw := cache.NewFilteredListWatchFromClient(clientset.CoreV1().RESTClient(), "configmaps", "default", func(options *metav1.ListOptions) {
		options.LabelSelector = "foo=bar"
	})

	store, controller := cache.NewInformer(lw, nil, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cm := obj.(*corev1.ConfigMap)
			fmt.Printf("ConfigMap Added: %s/%s\n", cm.Namespace, cm.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			cm := newObj.(*corev1.ConfigMap)
			fmt.Printf("ConfigMap Updated: %s/%s\n", cm.Namespace, cm.Name)
		},
		DeleteFunc: func(obj interface{}) {
			cm := obj.(*corev1.ConfigMap)
			fmt.Printf("ConfigMap Deleted: %s/%s\n", cm.Namespace, cm.Name)
		},
	})

	go controller.Run(wait.NeverStop)

	time.Sleep(5 * time.Second)

	fmt.Printf("stored keys: %v\n", store.ListKeys())
}
