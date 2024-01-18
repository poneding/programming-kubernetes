package main

import (
	"context"
	"log"
	"math/rand"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func main() {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Fatalf("new manager err: %s", err.Error())
	}

	err = (&PodReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)
	if err != nil {
		log.Fatalf("setup scheduler controller err: %s", err.Error())
	}

	err = mgr.Start(context.Background())
	if err != nil {
		log.Fatalf("start manager controller err: %s", err.Error())
	}
}

const mySchedulerName = "my-scheduler-controller"

type PodReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (s *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	nodes := new(corev1.NodeList)
	err := s.Client.List(ctx, nodes)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// 随机选择一个节点
	targetNode := nodes.Items[rand.Intn(len(nodes.Items))].Name

	// 创建绑定关系
	binding := new(corev1.Binding)
	binding.Name = req.Name
	binding.Namespace = req.Namespace
	binding.Target = corev1.ObjectReference{
		Kind:       "Node",
		APIVersion: "v1",
		Name:       targetNode,
	}

	err = s.Client.Create(ctx, binding)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (s *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// 过滤目标 Pod
	filter := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			pod, ok := e.Object.(*corev1.Pod)
			if ok {
				return pod.Spec.SchedulerName == mySchedulerName && pod.Spec.NodeName == ""
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(filter).
		Complete(s)
}
