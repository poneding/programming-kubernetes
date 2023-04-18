package main

import (
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 从本地 kubeconfig 文件中加载配置
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err.Error())
	}

	// 创建 kubernetes.Clientset
	client := kubernetes.NewForConfigOrDie(config)

	// 获取 API 资源列表
	// 获取服务端版本
	version, err := client.ServerVersion()
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Server Version: %#v\n", version)
	// 打印的结果如下（是不是有种似曾相识的感觉，试一下 kubectl version）
	// Server Version: &version.Info{Major:"1", Minor:"22", GitVersion:"v1.22.6+k3s1", GitCommit:"3228d9cb9a4727d48f60de4f1ab472f7c50df904", GitTreeState:"clean", BuildDate:"2022-01-25T01:27:44Z", GoVersion:"go1.16.10", Compiler:"gc", Platform:"linux/amd64"}

	// 查看 kubernetes.Clientset 的源码，可以看到内嵌了 discovery.DiscoveryClient 结构
	// 实质上 ServerVersion() 方法就是调用了 discovery.DiscoveryClient 的 ServerVersion() 方法
	disscoveryClient := discovery.NewDiscoveryClientForConfigOrDie(config)
	version, err = disscoveryClient.ServerVersion()
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Server Version: %#v\n", version)
}
