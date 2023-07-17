# admission-controller

## 示例

写一个 mutating webhook，当一个带有 `pk.poneding.com/echo-hello-sidecar=true` Label 的 Pod 被创建或者更新时，自动为其注册一个 `busybox` 的 sidecar 容器，sidecar 容器的启动命令为 `echo hello`。

## 参考
- [Kubernetes admission webhook server 开发教程](https://www.zeng.dev/post/2021-denyenv-validating-admission-webhook/#cert-manager-%E7%AD%BE%E5%8F%91-tls-%E8%AF%81%E4%B9%A6)
- [example-webhook-admission-controller](https://github.com/caesarxuchao/example-webhook-admission-controller)