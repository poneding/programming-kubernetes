package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

var (
	namespace         = "default"
	ConfigMapResource = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "configmaps",
	}
)

func main() {
	client := createClientOrDie()

	// workqueue 的特点：
	// - 公平：按照添加的顺序处理
	// - 吝啬：单个 item 不会被并发处理，如果一个 item 在处理之前被添加多次，那么它只会被处理一次
	// - 多租户：多个消费者和生产者。特别是，允许在 item 被处理时重新入队
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer queue.ShutDown()

	// 队列通常由一个或多个 informer 来填充，informer 会监听 Kubernetes 资源的事件。
	// 一个“惯用”的方式是通过 SharedInformerFactory 来获取 informer。
	// SharedInformerFactory 会为每个资源类型创建一个 informer，然后将它们缓存起来。
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		client, 5*time.Second, namespace, func(opts *metav1.ListOptions) {
			opts.LabelSelector = "foo=bar"
		},
	)
	dynamicInformer := factory.ForResource(ConfigMapResource)

	dynamicInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// key 是一个  <namespace>/<name> 格式的字符串 (对于集群范围的对象则是 <name> 格式)
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				fmt.Printf("New event: ADD %s\n", key)
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				fmt.Printf("New event: UPDATE %s\n", key)
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamespaceKeyFunc 在调用 MetaNamespaceKeyFunc
			// 之前检查 DeletedFinalStateUnknown 对象。
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				fmt.Printf("New event: DELETE %s\n", key)
				queue.Add(key)
			}
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动 informer
	factory.Start(ctx.Done())

	for gvr, ok := range factory.WaitForCacheSync(ctx.Done()) {
		if !ok {
			panic(fmt.Sprintf("Failed to sync cache for resource %v", gvr))
		}
	}

	for i := 0; i < 3; i++ {
		// 一个更好的方式是使用 wait.Until() 来为每个 worker 创建一个 goroutine
		// "k8s.io/apimachinery/pkg/util/wait"
		fmt.Printf("Starting worker %d\n", i)

		// worker()
		go func(n int) {
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("Controller's done! Worker %d exiting...\n", n)
					return
				default:
				}

				key, quit := queue.Get()
				if quit {
					fmt.Printf("Work queue has been shut down! Worker %d exiting...\n", n)
					return
				}
				fmt.Printf("Worker %d is about to start process new item %s.\n", n, key)

				// processSingleItem() - 一个处理 item 的函数
				func() {
					// 通知队列已经处理完毕
					// 它会解锁队列，允许其他 worker 并行处理
					// 因为具有相同的 key 的两个对象不会并行被处理
					defer queue.Done(key)

					// 控制器逻辑
					obj, err := dynamicInformer.Lister().Get(key.(string))
					if err == nil {
						fmt.Printf("Worker %d found ConfigMap object in informer's cahce %#v.\n", n, obj)
						if n == 1 {
							err = fmt.Errorf("worker %d is a chronic failure", n)
						}
					} else {
						fmt.Printf("Worker %d got error %v while looking up ConfigMap object in informer's cache.\n", n, err)
					}

					if err == nil {
						// 该 key 已经被成功处理 -> Forget
						fmt.Printf("Worker %d reconciled ConfigMap %s successfully. Removing it from te queue.\n", n, key)
						// Forget 表示一个项目已经完成重试。不管它是失败还是成功，我们都会停止速率限制器跟踪它。
						// 这只清除 rateLimiter，你仍然需要在队列上调用 Done()。
						queue.Forget(key)
						return
					}

					// 重试不超过 5 次
					if queue.NumRequeues(key) >= 5 {
						fmt.Printf("Worker %d gave up on processing %s. Removing it from the queue.\n", n, key)
						queue.Forget(key)
						return
					}

					// 重新放回队列，等待再次被处理
					fmt.Printf("Worker %d failed to process %s. Putting it back to the queue to retry later.\n", n, key)
					queue.AddRateLimited(key)
				}()
			}
		}(i)
	}

	cm1 := createConfigMap(client)
	cm2 := createConfigMap(client)
	cm3 := createConfigMap(client)
	cm4 := createConfigMap(client)
	cm5 := createConfigMap(client)

	deleteConfigMap(client, cm1)
	deleteConfigMap(client, cm2)
	deleteConfigMap(client, cm3)
	deleteConfigMap(client, cm4)
	deleteConfigMap(client, cm5)

	time.Sleep(10 * time.Second)
	queue.ShutDown()
	cancel()
	time.Sleep(1 * time.Second)
}

func createClientOrDie() dynamic.Interface {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", path.Join(home, ".kube/config"))
	if err != nil {
		panic(err.Error())
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return client
}

func createConfigMap(client dynamic.Interface) *unstructured.Unstructured {
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"namespace":    namespace,
				"generateName": "workqueue-",
				"labels": map[string]interface{}{
					"foo": "bar",
				},
			},
			"data": map[string]interface{}{
				"foo": "bar",
			},
		},
	}

	cm, err := client.
		Resource(ConfigMapResource).
		Namespace(namespace).
		Create(context.Background(), cm, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Created ConfigMap %s/%s\n", cm.GetNamespace(), cm.GetName())
	return cm
}

func deleteConfigMap(client dynamic.Interface, cm *unstructured.Unstructured) {
	err := client.
		Resource(ConfigMapResource).
		Namespace(cm.GetNamespace()).
		Delete(context.Background(), cm.GetName(), metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Deleted ConfigMap %s/%s\n", cm.GetNamespace(), cm.GetName())
}
