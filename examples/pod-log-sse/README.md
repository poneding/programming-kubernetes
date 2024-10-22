# pod-log-sse

利用 SSE 实现前端实时获取后端 Pod 的日志并不断推送。

运行：

```bash
# 1. 创建 Pod
kubectl apply -f pod.yaml

# 2. 运行
go run main.go

# 3. 浏览器访问 http://localhost:8080
```
