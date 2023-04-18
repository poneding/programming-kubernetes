package main

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	informers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

// 1. start the typed informer factory
// go run main.go
// 2. create & update & delete a configmap
// kubectl apply -f cm.yaml
// kubectl delete -f cm.yaml

func main() {
	config := controllerruntime.GetConfigOrDie()
	client := kubernetes.NewForConfigOrDie(config)

	factory := informers.NewSharedInformerFactoryWithOptions(client, 5*time.Second, informers.WithNamespace("default"), informers.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.LabelSelector = "foo=bar"
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmInformer := factory.Core().V1().ConfigMaps()

	cmInformer.Informer().AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cm := obj.(*corev1.ConfigMap)
			fmt.Printf("Informer event: ConfigMap ADDED %s/%s\n", cm.GetNamespace(), cm.GetName())
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*corev1.ConfigMap)
			new := newObj.(*corev1.ConfigMap)
			if old.GetResourceVersion() != new.GetResourceVersion() {
				fmt.Printf("Informer event: ConfigMap UPDATED %s/%s\n", new.GetNamespace(), new.GetName())
			}
		},
		DeleteFunc: func(obj interface{}) {
			cm := obj.(*corev1.ConfigMap)
			fmt.Printf("Informer event: ConfigMap DELETED %s/%s\n", cm.GetNamespace(), cm.GetName())
		},
	})

	factory.Start(ctx.Done())

	for gvr, ok := range factory.WaitForCacheSync(ctx.Done()) { // gvr?
		if !ok {
			panic(fmt.Sprintf("failed to sync cache for %v", gvr))
		}
		fmt.Printf("synced cache for %v\n", gvr)
	}

	// 通过 lister 获取所有的 configmap
	cmobjs, err := cmInformer.Lister().List(labels.Everything())
	// cmobjs, err := cmInformer.Lister().Get("dynamic-informer-cm")
	if err != nil {
		panic(err)
	}

	for _, cmobj := range cmobjs {
		fmt.Printf("cmobj: %s/%s\n", cmobj.GetNamespace(), cmobj.GetName())
	}

	for {
		time.Sleep(1 * time.Second)
	}
}
