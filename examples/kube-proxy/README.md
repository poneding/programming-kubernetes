# kube-proxy

[<- 返回上级](../index.md)

我们可以使用 `kubectl proxy` 代理 kube-apiserver 服务。

那么，我们可不可以创建一个 http 服务来代理 kube-apiserver 服务呢？那是完全没问题的！

1、首先我们需要导入必要的包：

```go
import (
    "net/http"
    "net/url"
    "strings"

    "k8s.io/apimachinery/pkg/util/proxy"
    "k8s.io/client-go/rest"
    ctrl "sigs.k8s.io/controller-runtime"
)
```

2、编写 http 服务的代码如下：

```go
func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        kconfig := ctrl.GetConfigOrDie()
        transport, err := rest.TransportFor(kconfig)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            w.Write([]byte(err.Error()))
            return
        }

        proxyUrl := *r.URL
        u, err := url.Parse(kconfig.Host)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            w.Write([]byte(err.Error()))
            return
        }

        proxyUrl.Host = u.Host
        proxyUrl.Scheme = u.Scheme

        kubeProxy := proxy.NewUpgradeAwareHandler(&proxyUrl, transport, true, false, nil)
        kubeProxy.UpgradeTransport = proxy.NewUpgradeRequestRoundTripper(transport, transport)
        kubeProxy.ServeHTTP(w, r)
    })
    http.ListenAndServe(":8888", nil)
}
```

如果你想为你的服务添加 pathbase，那么你可以为 `proxyUrl` 对象添加 `Path` 属性：

```go
func main() {
    http.HandleFunc("/kube-proxy", func(w http.ResponseWriter, r *http.Request) {
        ...
        proxyUrl.Path = strings.TrimPrefix(proxyUrl.EscapedPath(), "/kube-proxy")
        ...
    })
    http.ListenAndServe(":8888", nil)
}
```

3、运行：

```bash
go run main.go
```

4、测试：

```bash
# 获取集群中所有的命名空间
curl http://localhost:8888/api/v1/namespaces
```

如果顺利的话，你可以得到 json 格式的命名空间列表数据的输出。
