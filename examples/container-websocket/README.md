# container-websocket

[<- 返回上级](../index.md)

1. 通过 websocket 连接到容器，以终端的方式实现容器的交互式操作；
2. 通过 websocket 获取容器的日志信息；

## 运行

```bash
kubectl apply -f container-websocket-test-pod.yaml

go run main.go
```

## 预览

1. 打开浏览器;
2. 访问容器终端 `http://localhost:8080/cterm`；
3. 访问容器日志 `http://localhost:8080/clogs`；
