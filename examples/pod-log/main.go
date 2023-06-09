package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func main() {
	// kubectl apply -f log-foo.yaml
	// go run main.go
	// kubectl delete -f log-foo.yaml

	// 获取日志内容 - 单次获取
	getLogString()

	// 获取日志流
	getLogStream()
}

// 一次性获取日志内容
func getLogString() {
	config := controllerruntime.GetConfigOrDie()
	client := kubernetes.NewForConfigOrDie(config)

	logReq := client.CoreV1().Pods("default").GetLogs("log-foo", &corev1.PodLogOptions{
		Container: "foo",
	})

	rc, err := logReq.Stream(context.Background())
	if err != nil {
		log.Fatalf("stream error: %v", err)
	}

	defer rc.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, rc)
	if err != nil {
		log.Fatalf("copy error: %v", err)
	}
	log.Printf("log: %s", buf.String())
}

// 获取日志流
func getLogStream() {
	config := controllerruntime.GetConfigOrDie()
	client := kubernetes.NewForConfigOrDie(config)

	logReq := client.CoreV1().Pods("default").GetLogs("log-foo", &corev1.PodLogOptions{
		Container: "foo",
		Follow:    true,
		TailLines: &[]int64{10}[0],
	})

	rc, err := logReq.Stream(context.Background())
	if err != nil {
		log.Fatalf("stream error: %v", err)
	}

	defer rc.Close()
	for {
		buf := make([]byte, 2048)
		cnt, err := rc.Read(buf)
		if cnt == 0 {
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("read error: %v", err)
		}
		fmt.Print(string(buf[:cnt]))
	}
}
