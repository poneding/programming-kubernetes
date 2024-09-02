package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func streamLogs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	defer conn.Close()

	config := controllerruntime.GetConfigOrDie()

	client := kubernetes.NewForConfigOrDie(config)

	logReq := client.CoreV1().Pods("wutong-dev").GetLogs("test-app-helloweb-5685765d6f-wdzdq", &corev1.PodLogOptions{
		Container: "helloweb",
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
			log.Printf("read error: %v", err)
			break
		}
		err = conn.WriteMessage(websocket.TextMessage, buf[:cnt])
		if err != nil {
			log.Printf("write error: %v", err)
			break
		}
	}
}

func main() {
	http.HandleFunc("/ws", streamLogs)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
