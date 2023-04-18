package main

import (
	"net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

// 运行
// go run main.go
// 1. 查看所有可用接口
// curl http://localhost:8888/kube-proxy/
// 2. 查看所有命名空间
// curl http://localhost:8888/kube-proxy/api/v1/namespaces
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
	// 设置 pathbase
	// http.HandleFunc("/kube-proxy", func(w http.ResponseWriter, r *http.Request) {
	// 	...
	// 	proxyUrl.Path = strings.TrimPrefix(proxyUrl.EscapedPath(), "/kube-proxy")
	// 	...
	// })
	http.ListenAndServe(":8888", nil)
}
