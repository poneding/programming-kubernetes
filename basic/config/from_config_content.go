package config

import (
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var fakeConfig = `---
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: cert-xxx
    server: https://127.0.0.1:6443
  name: cluster-xxx
contexts:
- context:
    cluster: cluster-xxx
    namespace: default
    user: user-xxx
  name: context-xxx
current-context: context-xxx
kind: Config
preferences: {}
users:
- name: user-xxx
  user:
    token: token-xxx
`

func KubeConfigFromConfigContent() *rest.Config {
	realConfig, _ := os.ReadFile(clientcmd.RecommendedHomeFile)
	fakeConfig = string(realConfig)

	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(fakeConfig))
	if err != nil {
		panic(err)
	}
	return config
}
