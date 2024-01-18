package main

import (
	"context"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var kconfig *rest.Config
var kruntimeclient client.Client

func init() {
	kconfig = ctrl.GetConfigOrDie()
	var err error
	kruntimeclient, err = client.New(kconfig, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		log.Fatalf("failed to create k8s runtime client: %v", err)
	}
}

func bind(args extenderv1.ExtenderBindingArgs) *extenderv1.ExtenderBindingResult {
	log.Println("my-scheduler-extender bind called.")
	log.Printf("pod %s/%s is bind to %s", args.PodNamespace, args.PodName, args.Node)

	// 创建绑定关系
	binding := new(corev1.Binding)
	binding.Name = args.PodName
	binding.Namespace = args.PodNamespace
	binding.Target = corev1.ObjectReference{
		Kind:       "Node",
		APIVersion: "v1",
		Name:       args.Node,
	}

	result := new(extenderv1.ExtenderBindingResult)

	err := kruntimeclient.Create(context.Background(), binding)
	if err != nil {
		result.Error = err.Error()
	}

	return result
}
