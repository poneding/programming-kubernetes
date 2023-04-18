package main

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func main() {
	// 本质上来说，Indexer 就是一个 Store，只不过它可以根据指定的索引函数，为对象创建索引。
	foobarIndexFunc := func(obj interface{}) ([]string, error) {
		o, err := meta.Accessor(obj)
		if err != nil {
			return nil, err
		}
		return []string{o.GetLabels()["foo"], o.GetLabels()["bar"]}, nil
	}

	fooIndexName := "foo"
	barIndexName := "bar"
	myIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{
		fooIndexName: foobarIndexFunc,
		barIndexName: foobarIndexFunc,
	})

	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
			Labels: map[string]string{
				"foo": "foox",
				"bar": "barx",
			},
		},
	}

	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "default",
			Labels: map[string]string{
				"foo": "foox",
				"bar": "bary",
			},
		},
	}

	pod3 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod3",
			Namespace: "default",
			Labels: map[string]string{
				"foo": "fooy",
				"bar": "bary",
			},
		},
	}

	myIndexer.Add(pod1)
	myIndexer.Add(pod2)
	myIndexer.Add(pod3)

	fmt.Println("----- foox -----")

	fooxKeys, err := myIndexer.IndexKeys(fooIndexName, "foox")
	if err != nil {
		panic(err)
	}
	for _, k := range fooxKeys {
		fmt.Println(k) // default/pod1, default/pod2
	}

	fmt.Println("----- bary -----")

	baryKeys, err := myIndexer.IndexKeys(barIndexName, "bary")
	if err != nil {
		panic(err)
	}
	for _, k := range baryKeys {
		fmt.Println(k) // default/pod2, default/pod3
	}
}
