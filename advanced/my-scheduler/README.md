# 定制调度器

## 调度器的作用

Kubernetes 的调度器是一个独立的进程，它负责将 Pod 调度到集群中的节点上。

`kube-scheduler` 是 Kubernetes 集群的默认调度器，它会根据 Pod 的资源需求和节点的资源容量，将 Pod 调度到合适的节点上。

## 实现一个简单的调度器

1. 实现一个简单的调度器，并指定调度器的名称为 `my-scheduler`；
2. 指定 `Pod.Spec.SchedulerName` 字段值为 `my-scheduler`，这样 Pod 就会使用我们自定义的调度器。

### 调度器的实现

编写一一个调度器，原理：通过协调 Pod ，选择一个适合的节点，创建 Binding 对象，将 Pod 绑定到指定的节点上。

- 创建项目

```bash
mkdir my-scheduler && cd my-scheduler
go mod init my-scheduler
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

    err = (&MyScheduler{
        Client: mgr.GetClient(),
        Scheme: mgr.GetScheme(),
    }).SetupWithManager(mgr)
    if err != nil {
        log.Fatalf("setup scheduler err: %s", err.Error())
    }

    err = mgr.Start(context.Background())
    if err != nil {
        log.Fatalf("start manager err: %s", err.Error())
    }
}

const mySchedulerName = "my-scheduler"

type MyScheduler struct {
    Client client.Client
    Scheme *runtime.Scheme
}

func (s *MyScheduler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
func (s *MyScheduler) SetupWithManager(mgr ctrl.Manager) error {
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

### 调度器的使用

运行自定义调度器：

```bash
go run main.go
```

运行一个 Pod，指定调度器为 `my-scheduler`：

```bash
kubectl run nginx-by-my-scheduler --image=nginx --overrides='{"spec":{"schedulerName":"my-scheduler"}}'
```
