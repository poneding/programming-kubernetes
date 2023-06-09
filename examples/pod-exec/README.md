# pod-exec

[<- 返回上级](../index.md)

进入 Pod 容器终端。

## 代码示例

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh/terminal"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func main() {
	// kubectl apply -f exec-foo.yaml
	// go run main.go
	// kubectl delete -f exec-foo.yaml

	config := controllerruntime.GetConfigOrDie()

	client := kubernetes.NewForConfigOrDie(config)

	execReq := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name("exec-foo").
		Namespace("default").
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: "foo",
			Command:   []string{"/bin/bash"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", execReq.URL())
	if err != nil {
		log.Fatalf("create executor error: %v", err)
	}

	// 检查是不是终端
	if !terminal.IsTerminal(0) || !terminal.IsTerminal(1) {
		log.Fatalf("stdin/stdout should be terminal")
	}

	// 读取当前状态
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		fmt.Println(err.Error())
	}

	defer terminal.Restore(fd, oldState)

	if err := executor.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    true,
	}); err != nil {
		log.Fatalf("stream error: %v", err)
	}
}
```
