package main

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
)

// 实现一个简单的命令行工具，可以通过该命令行工具来查看集群中的资源对象
// go run main.go --help
// go run main.go po
// go run main.go pod
// go run main.go pods
// go run main.go services,deployments
// go run main.go --namespace=default service/kubernetes
// go run main.go --namespace default service kubernetes

func main() {
	configFlags := genericclioptions.NewConfigFlags(true)

	cmd := &cobra.Command{
		Use:  "kubectl (well, almost)",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			builder := resource.NewBuilder(configFlags)

			namespace := ""
			if configFlags.Namespace != nil {
				namespace = *configFlags.Namespace
			}

			obj, err := builder.
				WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
				NamespaceParam(namespace).
				DefaultNamespace().
				ResourceTypeOrNameArgs(true, args...).
				Do().
				Object()
			if err != nil {
				panic(err.Error())
			}

			printer := printers.NewTypeSetter(scheme.Scheme).ToPrinter(printers.NewTablePrinter(printers.PrintOptions{
				WithNamespace: true,
			}))
			if err := printer.PrintObj(obj, os.Stdout); err != nil {
				panic(err.Error())
			}
		},
	}
	configFlags.AddFlags(cmd.PersistentFlags())

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
