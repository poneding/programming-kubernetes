package main

import (
	"context"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func main() {
	// kubectl apply -f scale-foo-deploy.yaml
	// go run main.go
	// kubectl delete -f scale-foo-deploy.yaml

	config := controllerruntime.GetConfigOrDie()
	cli := kubernetes.NewForConfigOrDie(config)

	s, err := cli.AppsV1().Deployments("default").GetScale(context.Background(), "scale-foo", metav1.GetOptions{})
	if err != nil {
		log.Fatalf("get scale error: %v", err)
	}

	// Spec 只有一个 Replicas 字段
	s.Spec.Replicas = 3
	_, err = cli.AppsV1().Deployments("default").UpdateScale(context.Background(), "scale-foo", s, metav1.UpdateOptions{})
	if err != nil {
		log.Fatalf("update scale error: %v", err)
	}
	log.Println("update scale success.")
}
