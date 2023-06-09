package main

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"

	// "k8s.io/client-go/tools/clientcmd"

	controllerruntime "sigs.k8s.io/controller-runtime"
)

func main() {
	// 从本地 kubeconfig 文件中加载配置
	// config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	// if err != nil {
	// 	panic(err.Error())
	// }
	config := controllerruntime.GetConfigOrDie()

	// 创建 discovery.DiscoveryClient
	client := discovery.NewDiscoveryClientForConfigOrDie(config)

	// 获取 APIGroup 列表
	apiGroups, err := client.ServerGroups()
	if err != nil {
		panic(err.Error())
	}

	for _, apiGroup := range apiGroups.Groups {
		// 无法获取到 APIGroup 的 Kind 和 APIVersion
		// fmt.Printf("apiGroup.Kind: %v\n", apiGroup.Kind)
		// fmt.Printf("apiGroup.APIVersion: %v\n", apiGroup.APIVersion)
		fmt.Printf("apiGroup.Name: %v\n", apiGroup.Name)
		// APIGroup 中可能包含多个版本，例如 apps/v1、apps/v1beta1、apps/v1beta2
		// PreferredVersion 表示首选的版本
		fmt.Printf("apiGroup.PreferredVersion.GroupVersion: %v\n", apiGroup.PreferredVersion.GroupVersion)
	}

	fmt.Println("====================================")

	// 获取服务端 APIGroup 列表和 APIResourceList 列表
	// 一个 Group 下面可以包含多个 ResourceList
	discovery.ServerGroupsAndResources(client)
	_, resourceLists, err := client.ServerGroupsAndResources()
	if err != nil {
		panic(err.Error())
	}
	resourceCounts := 0
	for _, rl := range resourceLists {
		// fmt.Printf("rl.GroupVersion: %v\n", rl.GroupVersion)
		gv, err := schema.ParseGroupVersion(rl.GroupVersion) // 解析 GroupVersion
		if err != nil {
			panic(err.Error())
		}
		for _, r := range rl.APIResources {
			// 可以获取到单个资源的 Group、Version、Kind 信息
			if rl.GroupVersion == "apps/v1" && r.Kind == "Deployment" {
				fmt.Printf("\tr.Group: %v\n", gv.Group)     // 此时是无法直接通过 r.Group 获取到 Group 信息的
				fmt.Printf("\tr.Version: %v\n", gv.Version) // 此时是无法直接通过 r.Version 获取到 Version 信息的
				fmt.Printf("\tr.Kind: %v\n", r.Kind)
				fmt.Printf("\tr.Name: %v\n", r.Name)
				fmt.Printf("\tr.ShortNames: %v\n", r.ShortNames)         // 短名称，比如 Deployment 的短名称是 deploy
				fmt.Printf("\tr.SingularName: %v\n", r.SingularName)     // 单数形式的资源名称，比如 Pod 的单数形式是 pod
				fmt.Printf("\tr.Verbs.String(): %v\n", r.Verbs.String()) // 支持的操作，比如 get、list、watch、create、update、patch、delete、deletecollection

				fmt.Println()
			}
		}

		resourceCounts += len(rl.APIResources)
	}
	fmt.Println("====================================")

	// 获取首选的 APIResourceList 列表， 每个 APIGroup 只有一个首选的 APIResourceList
	preferredAPIResourceLists, err := client.ServerPreferredResources()
	if err != nil {
		panic(err.Error())
	}

	preferredResourceCounts := 0
	for _, parl := range preferredAPIResourceLists {
		// fmt.Printf("parl.GroupVersion: %v\n", parl.GroupVersion)
		// 非首选的 APIGroupVersion 会返回空的 APIResource 列表，例如 autoscaling/v1beta2
		if len(parl.APIResources) == 0 {
			fmt.Println("parl.APIResources is empty, gv: " + parl.GroupVersion)
		}
		preferredResourceCounts += len(parl.APIResources)
	}

	fmt.Println()
	// 比较 resourceCounts 和 preferredResourceCounts 的数量，可以发现 preferredResourceCounts 的数量要少
	fmt.Printf("resourceCounts: %v\n", resourceCounts)
	fmt.Printf("preferredResourceCounts: %v\n", preferredResourceCounts)
}
