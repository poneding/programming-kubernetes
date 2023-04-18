package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func main() {
	yamlContent := []byte(`
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: default
data:
  foo: bar
`)

	jsonContent, err := yaml.ToJSON(yamlContent)
	if err != nil {
		panic(err)
	}
	fmt.Printf("jsonContent: %s\n", string(jsonContent))

	unstructuredObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsonContent)
	fmt.Printf("unstructuredObj: %v\n", unstructuredObj)
}
