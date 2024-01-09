# admission-controller

## 示例

写一个 mutating webhook，当一个带有 `pk.poneding.com/echo-hello-sidecar=true` Label 的 Pod 被创建或者更新时，自动为其注册一个 `busybox` 的 sidecar 容器并运行 `echo hello` 命令。

## 部署

依赖 `cert-manager` 生成自签证书，安装命令：

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml
```

安装 webhook：

```bash
kubectl apply -f ./deploy
```

## 参考
- [准入控制器参考](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/)
- [动态准入控制](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/extensible-admission-controllers/)
- [Kubernetes admission webhook server 开发教程](https://www.zeng.dev/post/2021-denyenv-validating-admission-webhook/#cert-manager-%E7%AD%BE%E5%8F%91-tls-%E8%AF%81%E4%B9%A6)
- [example-webhook-admission-controller](https://github.com/caesarxuchao/example-webhook-admission-controller)