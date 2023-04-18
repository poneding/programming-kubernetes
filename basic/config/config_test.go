package config_test

import (
	"context"
	"fmt"
	"testing"

	"pk/basic/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestConfig(t *testing.T) {
	clientset, _ := kubernetes.NewForConfig(config.KubeConfigFromFlags())
	// clientset2, _ := kubernetes.NewForConfig(config.KubeConfigFromInClusterConfig())
	// clientset3, _ := kubernetes.NewForConfig(config.KubeConfigFromCtrlRuntime())
	// clientset4, _ := kubernetes.NewForConfig(config.KubeConfigFromConfigContent())
	pods, _ := clientset.CoreV1().Pods(metav1.NamespaceDefault).List(context.TODO(), metav1.ListOptions{})
	if pods != nil {
		for _, pod := range pods.Items {
			fmt.Println(pod.Name)
		}
	}
}
