package main

import (
	"fmt"

	"k8s.io/client-go/tools/cache"
)

type User struct {
	Name string
	Age  int
}

func UserKeyFunc(obj interface{}) (string, error) {
	return obj.(*User).Name, nil
}

func main() {
	// DeltaFiFO: map[string]Deltas
	// Deltas: []Delta
	// Delta.Type: Added, Updated, Deleted, Sync, Replaced (5 种类型)
	// Delta.Object: API 对象
	// Kubernetes client-go 中，DeltaFIFO 中存储着每个对象的所有变更，key 一般是 API 对象的 <namespace>/<name>
	df := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{KeyFunction: UserKeyFunc})
	u1 := &User{Name: "user1", Age: 10}
	u2 := &User{Name: "user2", Age: 20}
	u3 := &User{Name: "user3", Age: 30}

	df.Add(u1)
	df.Add(u2)
	df.Add(u3)

	u1.Age = 11
	df.Update(u1)
	df.Delete(u1)

	fmt.Printf("df.List(): %v\n", df.List())
	popFunc := func(pop interface{}, isInInitialList bool) error {
		for _, delta := range pop.(cache.Deltas) {
			switch delta.Type {
			case cache.Added:
				fmt.Printf("Added: %v\n", delta.Object)
			case cache.Updated:
				fmt.Printf("Updated: %v\n", delta.Object)
			case cache.Deleted:
				fmt.Printf("Deleted: %v\n", delta.Object)
			case cache.Sync:
				fmt.Printf("Sync: %v\n", delta.Object)
			case cache.Replaced:
				fmt.Printf("Replaced: %v\n", delta.Object)
			}
		}
		return nil
	}
	_, err := df.Pop(popFunc)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
}
