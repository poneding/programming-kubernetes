# 扩展调度器：Scheduler Extender

Kubernetes 调度器扩展：Scheduler Extender

## 编译

```bash
./build.sh
```

## 部署

```bash
kubectl apply -f deploy/manifests.yaml
```

## 测试

```bash
kubectl run nginx-by-my-scheduler-extender --image=nginx --overrides='{"spec":{"schedulerName":"my-scheduler-with-extender"}'
```

## 验证

```bash
kubectl logs deploy/my-scheduler-extender -n kube-system -f
```

可以看到类似如下的日志：

```bash
2024/01/12 08:07:27 my-scheduler-extender filter called.
2024/01/12 08:07:27 my-scheduler-extender bind called.
2024/01/12 08:07:27 pod default/nginx-by-my-scheduler-extender is bind to dev
```

并且将观察到 Pod 被正常调度到了节点上：

```bash
kubectl get pod -o wide
```
