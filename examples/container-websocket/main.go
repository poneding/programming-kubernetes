package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

var (
	Namespace     = "default"
	PodName       = "container-websocket-test-pod"
	ContainerName = "test"
)

var clientset kubernetes.Interface
var restconfig *rest.Config

func restconfigOrDie() *rest.Config {
	if restconfig == nil {
		restconfig = controllerruntime.GetConfigOrDie()
	}
	return restconfig
}

func clientsetOrDie() kubernetes.Interface {
	if clientset == nil {
		clientset = kubernetes.NewForConfigOrDie(restconfigOrDie())
	}
	return clientset
}

type websocketReader struct {
	conn *websocket.Conn
}

type websocketWriter struct {
	conn *websocket.Conn
}

// WebSocketReader 读取 WebSocket 消息作为 stdin
func newWebSocketReader(conn *websocket.Conn) *websocketReader {
	return &websocketReader{conn: conn}
}

// WebSocketWriter 将 stdout 和 stderr 写回 WebSocket
func newWebSocketWriter(conn *websocket.Conn) *websocketWriter {
	return &websocketWriter{conn: conn}
}

func (r *websocketReader) Read(p []byte) (int, error) {
	_, message, err := r.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	copy(p, message)
	return len(message), nil
}

func (w *websocketWriter) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 在生产环境中，请实现更严格的检查
		},
	}
)

func handleContainerTerminal(w http.ResponseWriter, r *http.Request) {
	// 升级 HTTP 连接到 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// 创建一个执行请求
	req := clientsetOrDie().CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(PodName).
		Namespace(Namespace).
		SubResource("exec").
		Param("container", ContainerName).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "true").
		Param("command", "/bin/sh")

	// 创建SPDY executor
	executor, err := remotecommand.NewSPDYExecutor(restconfigOrDie(), "POST", req.URL())
	if err != nil {
		log.Printf("Error creating SPDY Executor: %v", err)
		return
	}

	// 创建远程命令的流选项
	streamOptions := remotecommand.StreamOptions{
		Stdin:  newWebSocketReader(conn),
		Stdout: newWebSocketWriter(conn),
		Stderr: newWebSocketWriter(conn),
		Tty:    true,
	}

	// 执行远程命令
	err = executor.StreamWithContext(context.Background(), streamOptions)
	if err != nil {
		log.Printf("Error streaming to Pod: %v", err)
		return
	}
}

func handleContainerLogs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// 创建一个日志请求
	logReq := clientsetOrDie().CoreV1().Pods(Namespace).GetLogs(PodName, &corev1.PodLogOptions{
		Container: ContainerName,
		Follow:    true,
		TailLines: &[]int64{10}[0],
	})

	// 读取日志
	rc, err := logReq.Stream(context.Background())
	if err != nil {
		log.Printf("stream error: %v", err)
		return
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
	http.HandleFunc("/ws/cterm", handleContainerTerminal)
	http.HandleFunc("/ws/clogs", handleContainerLogs)

	http.HandleFunc("/cterm", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "cterm.html")
	})
	http.HandleFunc("/clogs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "clogs.html")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
