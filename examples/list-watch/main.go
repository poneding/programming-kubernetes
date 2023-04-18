package main

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

/*
在 Reflector 包中，存在着 ListWatch 客户端，其中包含了 list 和 watch 两个对象。
list 对象主要用于列出指定资源（如 pods）的当前列表版本，
而 watch 对象则用于追踪指定资源的当前版本并监听其后续的所有变更事件。
在 watch 的过程中，API Server 不可能长时间保留我们 watch  的某个资源版本。
因此，每个资源版本都会有一个过期时间。一旦版本过期，
watch 就会中断并返回 expired 相关的错误。此时，如果我们想持续监听并避免遗漏变更事件，
就需要持续记录资源版本号（或记录 API Server  传递的标记版本号）。
一旦之前我们监听的版本号过期，我们就可以从记录的版本号开始重新监听。
watch 对象使用的是 http 的 chunk 协议（数据分块协议），
在制作浏览器进度条时，我们也会使用该协议进行长连接。
*/

// 运行：
// go run main.go
// kubectl apply -f cm.yaml
// kubectl delete -f cm.yaml
func main() {
	clientset := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	// lw := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "configmaps", "default", fields.Everything())
	lw := cache.NewFilteredListWatchFromClient(clientset.CoreV1().RESTClient(), "configmaps", "default", func(options *metav1.ListOptions) {
		options.LabelSelector = labels.SelectorFromSet(labels.Set{
			"foo": "bar",
		}).String()
	})

	lister, err := lw.List(metav1.ListOptions{}) // 获取到监听事件对象列表
	if err != nil {
		panic(err)
	}
	cmlist := lister.(*corev1.ConfigMapList)
	fmt.Printf("lister ResourceVersion: %v\n", cmlist.ResourceVersion)
	for _, cm := range cmlist.Items {
		fmt.Printf("list cm.ResourceVersion: %v\n", cm.ResourceVersion)
		fmt.Printf("%s/%s\n", cm.Namespace, cm.Name)
	}

	fmt.Println("=======================================")

	watcher, err := lw.Watch(metav1.ListOptions{
		ResourceVersion: cmlist.ResourceVersion, // 从上一次监听的版本号开始监听
	})
	if err != nil {
		panic(err)
	}

	for {
		we, ok := <-watcher.ResultChan() // 获取到监听事件对象，对象中包括事件类型和事件对象
		if ok {
			cm := we.Object.(*corev1.ConfigMap)
			fmt.Printf("watch cm.ResourceVersion: %v\n", cm.ResourceVersion)
			fmt.Printf("%s: %s/%s\n", we.Type, cm.Namespace, cm.Name)
		}
	}
}
