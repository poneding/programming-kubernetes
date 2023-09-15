# my-apiserver

自定义的 apiserver：基于内存管理 Foo 资源。

## 部署

```bash
kubectl apply -f https://raw.githubusercontent.com/poneding/programming-kubernetes/master/examples/my-apiserver/deploy.yaml
```

## 示例

```bash
kubectl apply -f https://raw.githubusercontent.com/poneding/programming-kubernetes/master/examples/my-apiserver/example-pod.yaml
```

## 操作 foo 资源：

```bash
kubectl get foos.play.poneding.com

kubectl get foos.play.poneding.com example-foo -o yaml

kubectl edit foos.play.poneding.com example-foo

kubectl delete foos.play.poneding.com example-foo
```