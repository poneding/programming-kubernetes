# 定制调度器

## 调度器的作用

Kubernetes 的调度器是一个独立的进程，它负责将 Pod 调度到集群中的节点上。

`kube-scheduler` 是 Kubernetes 集群的默认调度器，它会根据 Pod 的资源需求和节点的资源容量，将 Pod 调度到合适的节点上。

## 实现一个简单的调度器

我们可以自定义一个调度器，并通过指定 PodSpec 的 `schedulerName` 字段，指定使用我们自定义的调度器。

一般来说，调度器的实现需要以下几个步骤：

- Filter：过滤掉不符合要求的节点
- Prioritize：对符合要求的节点排优先级，选出最优节点
- Bind：将 Pod 绑定到最优节点上

### 调度器的实现

我们可以通过编写一个 Pod 控制器，来实现一个调度器。原理：通过监听 Pod 的创建，然后创建 Binding 对象，将 Pod 绑定到指定的节点上。

- 创建项目

```bash
mkdir my-scheduler && cd my-scheduler
kubebuilder init
kubebuilder create api --group core --version v1 --kind Pod --controller=true
```

- 编写 `pod_controller.go` 控制逻辑，实现 Pod 控制器

```go
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    nodes := new(corev1.NodeList)
    err := r.List(ctx, nodes)
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

    err = r.Create(ctx, binding)
    if err != nil {
        return ctrl.Result{Requeue: true}, err
    }

    return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
    // 过滤目标 Pod
    filter := predicate.Funcs{
        CreateFunc: func(e event.CreateEvent) bool {
            pod, ok := e.Object.(*corev1.Pod)
            if ok {
                return pod.Spec.SchedulerName == "my-scheduler" && pod.Spec.NodeName == ""
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
        Complete(r)
}
```

### 调度器的使用

运行控制器：

```bash
make run
```

运行一个 Pod，指定调度器为 `my-scheduler`：

```bash
kubectl run nginx-by-my-scheduler --image=nginx --overrides='{"spec":{"schedulerName":"my-scheduler"}}'
```
