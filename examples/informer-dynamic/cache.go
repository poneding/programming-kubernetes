package main

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	dynamicSharedInformer dynamicinformer.DynamicSharedInformerFactory
	// myCRLister            cache.GenericLister
	podLister cache.GenericLister
)

func init() {
	setupInformers()
}

func setupInformers() {
	konfig := ctrl.GetConfigOrDie()
	fmt.Printf("konfig.Host: %v\n", konfig.Host)
	fmt.Printf("konfig.ServerName: %v\n", konfig.ServerName)
	fmt.Printf("konfig.APIPath: %v\n", konfig.APIPath)
	dynamicClient := dynamic.NewForConfigOrDie(konfig)

	dynamicSharedInformer = dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, time.Minute*5)

	// 自定义资源 Informer
	// myCRInformer := dynamicSharedInformer.ForResource(schema.GroupVersionResource{
	// 	Group:    "<your-group>",
	// 	Version:  "<your-version>",
	// 	Resource: "<your-resource>",
	// })

	podInformer := dynamicSharedInformer.ForResource(corev1.SchemeGroupVersion.WithResource("pods"))

	// myCRLister = myCRInformer.Lister()
	podLister = podInformer.Lister()

	dynamicSharedInformer.Start(wait.NeverStop)
	dynamicSharedInformer.WaitForCacheSync(wait.NeverStop)
	fmt.Println("synced done.")
}

func main() {
	listResources()
}

func listResources() {
	// mycrs, err := myCRLister.ByNamespace(metav1.NamespaceDefault).List(labels.Everything())
	// if err != nil {
	// 	panic(err)
	// }
	//
	// for _, mycr := range mycrs {
	// 	var cr corev1.Pod
	// 	err = runtime.DefaultUnstructuredConverter.FromUnstructured(mycr.(*unstructured.Unstructured).Object, &cr)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("%s/%s\n", cr.GetNamespace(), cr.GetName())
	// }

	resources, err := podLister.ByNamespace(metav1.NamespaceDefault).List(labels.Everything())
	if err != nil {
		panic(err)
	}

	for _, r := range resources {
		var pod corev1.Pod
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(r.(*unstructured.Unstructured).Object, &pod)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s/%s\n", pod.GetNamespace(), pod.GetName())
	}
}
