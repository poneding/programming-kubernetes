package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

// SSE 推送日志给前端
func sseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// 设置响应头，告知客户端这是一个 SSE 流
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 获取 Pod 的 namespace 和 name
	namespace := "default"
	podName := "log-foo"

	// 初始化 Kubernetes 客户端
	config := controllerruntime.GetConfigOrDie()
	client := kubernetes.NewForConfigOrDie(config)

	// 获取 Pod 日志
	tailLines := int64(100)
	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Follow:     true,       // 持续获取日志
		TailLines:  &tailLines, // 获取最后 100 行日志
		Timestamps: true,       // 显示时间戳
	})

	stream, err := req.Stream(r.Context())
	if err != nil {
		http.Error(w, "Failed to stream pod logs", http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	scaner := bufio.NewScanner(stream)

	for {
		select {
		case <-r.Context().Done():
			log.Printf("context done: %v", r.Context().Err())
			return
		default:
			if scaner.Scan() {
				txt := scaner.Text()
				fmt.Fprintf(w, "data: %s\n\n", txt) // [data: ] 是 SSE 的固定格式
				flusher.Flush()
			}
		}
	}
}

func main() {
	// index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	// 创建 HTTP 路由
	http.HandleFunc("/logs", sseHandler)

	// 启动服务器
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
