package main

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
)

// 实现一个简单的命令行工具，可以通过该命令行工具来查看 yaml 文件中定义的资源对象
// go run main.go resources.yaml
// go run main.go https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.6.4/deploy/static/provider/cloud/deploy.yaml

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

			// 你可以使用这种方式从远程文件中读取资源对象
			obj, err := builder.
				WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
				NamespaceParam(namespace).
				DefaultNamespace().
				FilenameParam(false, &resource.FilenameOptions{Filenames: args}).
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
