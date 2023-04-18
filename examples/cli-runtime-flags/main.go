package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	// 返回一个 *genericclioptions.ConfigFlags 实例
	cf := genericclioptions.NewConfigFlags(true)

	cmd := cobra.Command{
		Use: "kubectl (well, almost)",
		Run: func(cmd *cobra.Command, args []string) {
			// 有意思的几个函数：

			// - RawConfig() 可以看成是 kubeconfig 文件内容反序列化后的实例结果
			kubeconfig, err := cf.ToRawKubeConfigLoader().RawConfig()
			if err != nil {
				panic(err)
			}
			for c := range kubeconfig.Clusters {
				fmt.Printf("cluster: %v\n", c)
			}

			// - ToRESTConfig()
			restconfig, err := cf.ToRESTConfig()
			if err != nil {
				panic(err)
			}
			fmt.Printf("restconfig.Host: %v\n", restconfig.Host)

			// - ToDiscoveryClient()
			discoveryclient, err := cf.ToDiscoveryClient()
			if err != nil {
				panic(err)
			}
			serverversion, err := discoveryclient.ServerVersion()
			if err != nil {
				panic(err)
			}
			fmt.Printf("serverversion: %v\n", serverversion.GitVersion)

			// - ToRESTMapper()
			restmapper, err := cf.ToRESTMapper()
			if err != nil {
				panic(err)
			}
			gvk, err := restmapper.KindFor(schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			})
			if err != nil {
				panic(err)
			}
			fmt.Printf("gvk.Kind: %v\n", gvk.Kind)
		},
	}

	cf.AddFlags(cmd.PersistentFlags())
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
