package main

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
)

func main() {
	namespace := "default"
	obj := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "printer-configmap",
			Labels: map[string]string{
				"test": "printer",
			},
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}

	// YAML
	fmt.Println("# YAML ConfigMap representation")
	printer := printers.NewTypeSetter(scheme.Scheme).ToPrinter(&printers.YAMLPrinter{})
	if err := printer.PrintObj(obj, os.Stdout); err != nil {
		panic(err.Error())
	}

	// 如果直接使用 printers.YAMLPrinter{}，会报错：missing apiVersion or kind; 必须为 obj 设置 GroupVersionKind
	// obj.GetObjectKind().SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))
	// 或者为 obj 设置 GroupVersionKind
	// obj.APIVersion = corev1.SchemeGroupVersion.String()
	// obj.Kind = "ConfigMap"

	// yamlprinter := &printers.YAMLPrinter{}
	// err := yamlprinter.PrintObj(obj, os.Stdout)
	// if err != nil {
	// 	panic(err.Error())
	// }

	fmt.Println()

	// JSON
	fmt.Println("# JSON ConfigMap representation")
	printer = printers.NewTypeSetter(scheme.Scheme).ToPrinter(&printers.JSONPrinter{})
	if err := printer.PrintObj(obj, os.Stdout); err != nil {
		panic(err.Error())
	}

	fmt.Println()

	// Table(可读性更好，默认)
	fmt.Println("# Table ConfigMap representation")
	printer = printers.NewTypeSetter(scheme.Scheme).ToPrinter(printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: true,
		ShowLabels:    true,
	}))
	if err := printer.PrintObj(obj, os.Stdout); err != nil {
		panic(err.Error())
	}

	fmt.Println()

	// JSONPath
	fmt.Println("# JSONPath ConfigMap representation -> .data.foo")
	jsonprinter, err := printers.NewJSONPathPrinter("{.data.foo}")
	if err != nil {
		panic(err.Error())
	}
	printer = printers.NewTypeSetter(scheme.Scheme).ToPrinter(jsonprinter)
	if err := printer.PrintObj(obj, os.Stdout); err != nil {
		panic(err.Error())
	}

	fmt.Println()
}
