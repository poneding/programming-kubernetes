# 扩展调度器：定制调度控制器

1. 实现一个简单的调度控制器，并指定调度器的名称为 `my-scheduler-controller`；
2. 指定 `Pod.Spec.SchedulerName` 字段值为 `my-scheduler-controller`，这样 Pod 就会使用我们自定义的调度器。

## 实现

编写一一个调度器，原理：通过协调 Pod ，选择一个适合的节点，创建 Binding 对象，将 Pod 绑定到指定的节点上。

- 创建项目

```bash
mkdir my-scheduler-controller && cd my-scheduler-controller
go mod init my-scheduler-controller
touch main.go
```

- 编写 `main.go` 调度器逻辑（本质是一个 Pod 的协调控制器）

```go
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
```

## 使用

运行自定义调度器：

```bash
# 本地运行
go run main.go

# 部署到集群
kubectl apply -f deploy/manifests.yaml
```

运行一个 Pod，指定调度器为 `my-scheduler-controller`：

```bash
kubectl run nginx-by-my-scheduler-controller --image=nginx --overrides='{"spec":{"schedulerName":"my-scheduler-controller"}}'
```

## 验证

查看 Pod 信息：

```bash
kubectl get pod nginx-by-my-scheduler-controller -o wide
```

如果一切正常，Pod 将被正常调度到节点上。
