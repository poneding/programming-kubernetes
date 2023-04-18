package main

import (
	"k8s.io/apimachinery/pkg/api/errors"
)

func main() {
	// 对 Kubernetes 资源进行 CRUD 操作时，错误类型的判断
	var err error // 假设这是一个 CRUD 返回的错误

	// 1. 判断资源是否已经存在，一般用于创建资源返回的错误
	errors.IsAlreadyExists(err)

	// 2. 判断资源是否不存在，一般用于更新资源返回的错误
	errors.IsNotFound(err)

	// 3. 判断资源是否更新冲突，用于更新资源返回的错误
	errors.IsConflict(err)

	// 4. 判断资源是否被禁止访问
	errors.IsForbidden(err)

	// 5. 还有很多其他的错误类型判断，可以查阅该包下的定义 k8s.io/apimachinery/pkg/api/errors
}
