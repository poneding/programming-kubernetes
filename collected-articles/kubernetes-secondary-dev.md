# kubernetes-secondary-dev

[<- 收藏文章](./index.md)

> 本文转自 ShadowYD 的博客，[**原文**](https://juejin.cn/post/7203690731276517432)，版权归原作者所有。

## **1. 简介**

当使用 Kubernetes 进行应用程序的开发和部署时，**client-go** 是一个非常重要的工具。它是 Kubernetes 的官方客户端库，提供了与 Kubernetes ApiServer 进行通信的接口和实现。

client-go 主要提供以下几个功能：

1. **与 Kubernetes ApiServer 进行通信**：client-go 提供了与 Kubernetes ApiServer 进行通信的接口和实现，包括基本的 http 请求和更深层次的封装。开发人员可以使用 client-go 创建、更新和删除 Kubernetes 中的资源。
2. **访问 Kubernetes ApiServer 中的资源**：client-go 提供了访问 Kubernetes ApiServer 中资源的方法，包括使用 `ClientSet` 进行基于对象的访问和使用 `DynamicClient` 进行基于无类型的访问。
3. **处理 Kubernetes 资源的事件**：client-go 提供了一种称为 `Informer` 的机制，它可以监听 Kubernetes ApiServer 中的资源变更事件。开发人员可以使用 `Informer` 实现资源的快速检索和本地缓存，从而减轻对 ApiServer 的访问压力。
4. **发现 Kubernetes ApiServer 中的资源**：client-go 还提供了 `DiscoveryClient` 接口，该接口可以用于在 Kubernetes ApiServer 中查找特定资源的详细信息。

总的来说，client-go 是 Kubernetes 开发人员不可或缺的工具之一。它提供了丰富的功能和灵活的接口，使开发人员能够更轻松地构建和管理 Kubernetes 应用程序。

上述的要点下文都会一一的酌情展开，因为我需要开发多集群管理平台和一些 K8s 组件所以在 client-go 上有深度的使用 ，在 client-go 上的一些小坑和解决技巧会在下一篇文章中列出，本文更多关注 client-go 关于 **Informer** 的详细用法。

## **2. Client**

> 这里只简单介绍其封装好的几个 client，调用起来都比较方便所以就不展开了。

### **2.1 加载 kubeconfig 配置**

加载 kubeconfig 及各客户端初始化的方法：

```go
package config

import (
        "k8s.io/client-go/discovery"
        "k8s.io/client-go/dynamic"
        "k8s.io/client-go/kubernetes"
        "k8s.io/client-go/rest"
        "k8s.io/client-go/tools/clientcmd"
        "log"
)

const kubeConfigFilePath = "/Users/ShadowYD/.kube/config"

type K8sConfig struct {
}

func NewK8sConfig() *K8sConfig {
        return &K8sConfig{}
}
// 读取kubeconfig 配置文件
func (this *K8sConfig) K8sRestConfig() *rest.Config {
        config, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)

        if err != nil {
                log.Fatal(err)
        }

        return config
}
// 初始化 clientSet
func (this *K8sConfig) InitClient() *kubernetes.Clientset {
        c, err := kubernetes.NewForConfig(this.K8sRestConfig())

        if err != nil {
                log.Fatal(err)
        }

        return c
}

// 初始化 dynamicClient
func (this *K8sConfig) InitDynamicClient() dynamic.Interface {
        c, err := dynamic.NewForConfig(this.K8sRestConfig())

        if err != nil {
                log.Fatal(err)
        }

        return c
}

// 初始化 DiscoveryClient
func (this *K8sConfig) InitDiscoveryClient() *discovery.DiscoveryClient {
        return discovery.NewDiscoveryClient(this.InitClient().RESTClient())
}
```

### **2.2 ClientSet**

ClientSet 是比较常用的一个 client，常用于对 K8s 内部资源做 CRUD 或查询当前集群拥有什么资源：

```go
func main () {
// 使用的是上文提到的配置加载对象
    cliset := NewK8sConfig().InitClient()
    configMaps, err := cliset.CoreV1().ConfigMaps(ns).List(metav1.ListOptions{})
    if err != nil {
       panic(err)
    }
    for _, cm := range configMaps.Items {
       fmt.Printf("configName: %v, configData: %v \n", cm.Name, cm.Data)
    }
    return nil
}
```

### **2.3 DynamicClient**

DynamicClient 也是比较常用的 client 之一，但频繁度不及 ClientSet，它主要作用是用于 CRD (自定义资源)。当然它也可以用于 K8s 的内部资源，我们在项目内就用它来开发出可以对任意资源做 CRUD 的接口。

下面将演示使用 dynamicClient 创建资源，先在 tpls/deployment.yaml 测试配置：

```go
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myngx
  namespace: default
spec:
  selector:
    matchLabels:
      app: myngx
  replicas: 1
  template:
    metadata:
      labels:
        app: myngx
    spec:
      containers:
        - name: myngx-container
          image: nginx:1.18-alpine
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 80
```

使用 DynamicClient 创建测试配置：

```go
package main

import (
   "context"
   _ "embed"
   "k8s-clientset/config"
   metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
   "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
   "k8s.io/apimachinery/pkg/runtime/schema"
   "k8s.io/apimachinery/pkg/util/yaml"
   "log"
)

// 这个是新特性使用注释加载配置
//go:embed tpls/deployment.yaml
var deployTpl string

// dynamic client 创建 Deploy
func main()  {

// 动态客户端
   dynamicCli := config.NewK8sConfig().InitDynamicClient()

// 可以随意指定集群拥有的资源, 进行创建
   deployGVR := schema.GroupVersionResource{
      Group: "apps",
      Version: "v1",
      Resource: "deployments",
   }

   deployObj := &unstructured.Unstructured{}
   if err := yaml.Unmarshal([]byte(deployTpl), deployObj); err != nil {
       log.Fatalln(err)
   }

   if _, err = dynamicCli.
      Resource(deployGVR).
      Namespace("default").
      Create(context.Background(), deployObj, metav1.CreateOptions{});
      err != nil {
      log.Fatalln(err)
   }

   log.Println("Create deploy succeed")
}
```

### **2.4 DiscoveryClient**

DiscoveryClient 顾名思义就是用于发现 K8s 资源的，当我们不知道当前集群有什么资源时就会用该客户端封装好的方法进行查询。`kubectl api-resources` 命令就是用它实现的：

```go
package main

import (
        "fmt"
        "k8s-clientset/config"
)

func main() {
        client := config.NewK8sConfig().InitDiscoveryClient()
// 可以看到当前集群的 gvr
        preferredResources, _ := client.ServerPreferredResources()
        for _, pr := range preferredResources {
                fmt.Println(pr.String())
        }

// _, _, _ = client.ServerGroupsAndResources()

}
```

## **3. Informer**

### **3.1 前言**

本文重点就是放在 Informer 的源码的调试，以及如何去使用 Informer 达到对多集群查询目的之余也不会对集群的 API  Server 造成压力。下面将沿着 Informer 架构图一步一步的剖析每个环节，你将知道 informer 每一步的运作方式。全网可能独一份，是不是该 **点赞👍** 以示支持一下？！

### **3.2 Informer 架构图**

> 该图其实还有下半部分是关于 **Custom Controller**, 想了解请跳转 👉Controller 源码解析。

[https://mmbiz.qpic.cn/mmbiz/qFG6mghhA4aJpwMWCeGeSpkBWrQ0qdbHeticibOibu6iaoBDLBF10m8VRkzcOhpRBKNhawoF68rw35KdeLTlhs5iaOg/640?wx_fmt=other&wxfrom=5&wx_lazy=1&wx_co=1](https://mmbiz.qpic.cn/mmbiz/qFG6mghhA4aJpwMWCeGeSpkBWrQ0qdbHeticibOibu6iaoBDLBF10m8VRkzcOhpRBKNhawoF68rw35KdeLTlhs5iaOg/640?wx_fmt=other&wxfrom=5&wx_lazy=1&wx_co=1)

上图的流程解析 :

1. Reflector(反射器) 通过 http trunk 协议监听 K8s apiserver 服务的资源变更事件 , 事件主要分为三个动作 `ADD`、`UPDATE`、`DELETE`；
2. Reflector(反射器) 将事件添加到 Delta 队列中等待；
3. Informer 从队列获取新的事件；
4. Informer 调用 Indexer (索引器 , 该索引器内包含 Store 对象), 默认索引器是以 namespace 和 name 作为每种资源的索引名；
5. Indexer 通过调用 Store 存储对象按资源分类存储；

### **3.3 源码调试与分析**

> 下面部分示例需要把部分源码 copy 到一个可导入的目录下，因为有些源码是私有化不允许通过包 import。

### **3.3.1 从头说起 List & Watch**

在 Reflector 包中，存在着 ListWatch 客户端，其中包含了 list 和 watch 两个对象。list 对象主要用于列出指定资源（如 pods）的当前列表版本，而 watch 对象则用于追踪指定资源的当前版本并监听其后续的所有变更事件。

在 watch 的过程中，API Server 不可能长时间保留我们 watch  的某个资源版本。因此，每个资源版本都会有一个过期时间。一旦版本过期，watch 就会中断并返回 expired  相关的错误。此时，如果我们想持续监听并避免遗漏变更事件，就需要持续记录资源版本号（或记录 API Server  传递的标记版本号）。一旦之前我们监听的版本号过期，我们就可以从记录的版本号开始重新监听。

watch 对象使用的是 http 的 chunk 协议（数据分块协议），在制作浏览器进度条时，我们也会使用该协议进行长连接。

用代码调试一下如何 watch Pod 资源，下面仅仅是代码片段需要自行补全：

```go
package main

import (
        "fmt"
        "k8s-clientset/deep_client_go/reflector/helper"
        v1 "k8s.io/api/core/v1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/apimachinery/pkg/fields"
        "k8s.io/client-go/tools/cache"
        "log"
)

// create pods list & watch
func main() {
// helper 只是一个类似上文演示的 config, 只要用于初始化各种客户端
        cliset := helper.InitK8SClient()
        lwc := cache.NewListWatchFromClient(cliset.CoreV1().RESTClient(), "pods", "kube-system", fields.Everything())
        watcher, err := lwc.Watch(metav1.ListOptions{})
        if err != nil {
                log.Fatalln(err)
        }
        for {
                select {
                case v, ok := <-watcher.ResultChan():
                        if ok {
                                fmt.Println(v.Type, ":", v.Object.(*v1.Pod).Name, "-", v.Object.(*v1.Pod).Status.Phase)
                        }

                }
        }
}

// 输出结果
// ADDED : kube-apiserver-k8s-01 - Running
// ADDED : kube-scheduler-k8s-01 - Running
// ADDED : coredns-65c54cc984-26zx9 - Running
// ADDED : metrics-server-7fd564dc66-sm29c - Running
// ADDED : kube-proxy-6jl96 - Running
// ADDED : coredns-65c54cc984-bgmpm - Running
// ADDED : etcd-k8s-01 - Running
// ADDED : kube-controller-manager-k8s-01 - Running
```

当你做 Pod 资源变更时便可以接收到变更事件：

```txt
// 执行 kubectl apply -f  deploy.yaml
//ADDED : mygott-7565765f4d-2t4z8 - Pending
//MODIFIED : mygott-7565765f4d-2t4z8 - Pending
//MODIFIED : mygott-7565765f4d-2t4z8 - Pending
//MODIFIED : mygott-7565765f4d-2t4z8 - Running

// 执行 kubectl delete deploy mygott
//MODIFIED : mygott-7565765f4d-2t4z8 - Running
//MODIFIED : mygott-7565765f4d-2t4z8 - Running
//MODIFIED : mygott-7565765f4d-2t4z8 - Running
//DELETED : mygott-7565765f4d-2t4z8 - Running
```

### **3.3.2 入列 DeltaFifo**

从 reflector 中获取到资源事件然后放入先进先出队列，事件对象包含了 2 个属性如下所示：

```txt
type Event struct {
// 事件类型
        Type EventType
// 资源对象
        Object runtime.Object
}
// 事件类型如下:
// 资源添加事件
Added    EventType = "ADDED"
// 资源修改事件
Modified EventType = "MODIFIED"
// 资源删除事件
Deleted  EventType = "DELETED"
// 标记资源版本号事件, 这个就是用于可重新watch的版本号
Bookmark EventType = "BOOKMARK"
// 错误事件
Error    EventType = "ERROR"
```

DeltaFifo 队列源码调试，添加 Pod 资源入队列：

```go
package main

import (
        "fmt"
        "k8s.io/client-go/tools/cache"
)

type Pod struct {
        Name  string
        Value int
}

func NewPod(name string, v int) Pod {
        return Pod{Name: name, Value: v}
}

// 需要提供一个资源的唯一标识的字符串给到 DeltaFifo， 这样它就能追踪某个资源的变化
func PodKeyFunc(obj interface{}) (string, error) {
        return obj.(Pod).Name, nil
}

func main() {
        df := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{KeyFunction: PodKeyFunc})

// ADD3个object 进入 fifo
        pod1 := NewPod("pod-1", 1)
        pod2 := NewPod("pod-2", 2)
        pod3 := NewPod("pod-3", 3)
        df.Add(pod1)
        df.Add(pod2)
        df.Add(pod3)
// Update pod-1
        pod1.Value = 11
        df.Update(pod1)
        df.Delete(pod1)

// 当前df 的列表
        fmt.Println(df.List())

// 循环抛出事件
        for {
            df.Pop(func(i interface{}) error {
                for _, delta := range i.(cache.Deltas) {
                    switch delta.Type {
                    case cache.Added:
                            fmt.Printf("Add Event: %v \n", delta.Object)
                            break
                    case cache.Updated:
                            fmt.Printf("Update Event: %v \n", delta.Object)
                            break
                    case cache.Deleted:
                            fmt.Printf("Delete Event: %v \n", delta.Object)
                            break
                    case cache.Sync:
                            fmt.Printf("Sync Event: %v \n", delta.Object)
                            break
                    case cache.Replaced:
                            fmt.Printf("Replaced Event: %v \n", delta.Object)
                            break
                    }
                }
                return nil
            })
        }
}

// 输出结果, 可以看到先入列的资源事件会被先抛出
// 这是由于底层是是用 map 来记录资源的唯一标识起到快速索引和去重复的作用;
//[{pod-1 11} {pod-2 2} {pod-3 3}]
//Add Event: {pod-1 1}
//Update Event: {pod-1 11}
//Delete Event: {pod-1 11}
//Add Event: {pod-2 2}
//Add Event: {pod-3 3}
```

### **3.3.3 Reflector 的构造**

上述 2 个小节已经把 listWatch 客户端和 DeltaFifo 如何工作的方法说明了一下，本小节演示 Reflector 对象整合 listWatch 和 DeltaFifo。

```go
package main

import (
        "fmt"
        "k8s-clientset/deep_client_go/reflector/helper"
        v1 "k8s.io/api/core/v1"
        "k8s.io/apimachinery/pkg/fields"
        "k8s.io/client-go/tools/cache"
        "time"
)

// simulate  K8s simple reflector creation process
func main() {

        cliset := helper.InitK8SClient()
// 使用 store 进行存储，这样本地才有一份数据；
// 如果本地没有存储到被删除的资源， 则不需要 Pop 该资源的 Delete 事件；
// 所以我们为了准确接收到delete时接收到 Delete 事件, 所以预先创建一下 store
// cache.MetaNamespaceKeyFunc 是用于返回资源的唯一标识, {namespace}/{name} 或 {name}
        store := cache.NewStore(cache.MetaNamespaceKeyFunc)

// create list & watch Client
        lwc := cache.NewListWatchFromClient(cliset.CoreV1().RESTClient(),
                helper.Resource,
                helper.Namespace,
                fields.Everything(),
        )

// create deltafifo
        df := cache.NewDeltaFIFOWithOptions(
                cache.DeltaFIFOOptions{
                        KeyFunction:  cache.MetaNamespaceKeyFunc,
                        KnownObjects: store,
                })

// crete reflector
        rf := cache.NewReflector(lwc, &v1.Pod{}, df, time.Second*0)
        rsCH := make(chan struct{})
        go func() {
                rf.Run(rsCH)
        }()

// fetch delta event
        for {
            df.Pop(func(i interface{}) error {
// deltas
                for _, d := range i.(cache.Deltas) {
                    fmt.Println(d.Type, ":", d.Object.(*v1.Pod).Name,
                            "-", d.Object.(*v1.Pod).Status.Phase)
                    switch d.Type {
                    case cache.Sync, cache.Added:
// 向store中添加对象
                            store.Add(d.Object)
                    case cache.Updated:
                            store.Update(d.Object)
                    case cache.Deleted:
                            store.Delete(d.Object)
                    }
                }
                return nil
            })
        }
}

// 输出结果
//Sync : pod-1 - Running
//Sync : web-sts-1 - Running
//Sync : web-sts-0 - Running
//Sync : ngx-8669b5c9d-xwljg - Running

// 执行 kubectl apply -f  deploy.yaml
//Added : mygott-7565765f4d-x6znf - Pending
//Updated : mygott-7565765f4d-x6znf - Pending
//Updated : mygott-7565765f4d-x6znf - Pending
//Updated : mygott-7565765f4d-x6znf - Running

// 执行 kubectl delete deploy mygott
//Updated : mygott-7565765f4d-x6znf - Running
//Updated : mygott-7565765f4d-x6znf - Running
//Updated : mygott-7565765f4d-x6znf - Running
//Deleted : mygott-7565765f4d-wcml6 - Running
```

### **3.3.4 Indexer 与 Store**

### **>> Store**

**Store** 是如何存储资源对象的？其实通过 `NewStore` 方法就能立刻找到的答案，底层则是一个 `ThreadSafeStore` 的对象来存储资源的。而它的核心数据结构是一个 map 并且配合互斥锁保证并发安全，下面的源码的 item 字段就是其存储的核心：

```go
func NewStore(keyFunc KeyFunc) Store {
    return &cache{
            cacheStorage: NewThreadSafeStore(Indexers{}, Indices{}),
            keyFunc:      keyFunc,
        }
}

// NewThreadSafeStore creates a new instance of ThreadSafeStore.
func NewThreadSafeStore(indexers Indexers, indices Indices) ThreadSafeStore {
    return &threadSafeMap{
        items:    map[string]interface{}{},
        indexers: indexers,
        indices:  indices,
    }
}

// threadSafeMap implements ThreadSafeStore
type threadSafeMap struct {
    lock  sync.RWMutex
    items map[string]interface{}

// indexers maps a name to an IndexFunc
    indexers Indexers
// indices maps a name to an Index
    indices Indices
}
```

我们可以一起看看 `ThreadSafeStore` 所含有的的一些动作，便很容易理解其工作的方式：

```go
type ThreadSafeStore interface {
        Add(key string, obj interface{})
        Update(key string, obj interface{})
        Delete(key string)
        Get(key string) (item interface{}, exists bool)
        List() []interface{}
        ListKeys() []string
        Replace(map[string]interface{}, string)
        Index(indexName string, obj interface{}) ([]interface{}, error)
        IndexKeys(indexName, indexKey string) ([]string, error)
        ListIndexFuncValues(name string) []string
        ByIndex(indexName, indexKey string) ([]interface{}, error)
        GetIndexers() Indexers
        AddIndexers(newIndexers Indexers) error
        Resync() error
}
```

在 `threadSafeMap` 上还有一层用于 Store 的标准接口，用于存储 K8s 资源即 runtime.Object 的专用实现（runtime.Object 在 K8s 二开中是一个很重要的概念）：

```go
type Store interface {
        Add(obj interface{}) error
        Update(obj interface{}) error
        Delete(obj interface{}) error
        List() []interface{}
        ListKeys() []string
        Get(obj interface{}) (item interface{}, exists bool, err error)
        GetByKey(key string) (item interface{}, exists bool, err error)
        Replace([]interface{}, string) error
        Resync() error
}
```

到此我们大概知道 Store 是怎么工作的了，Store 的调用演示可以查看 **3.3.3 章节**。

### **>> Indexer**

**Indexer** 用于对资源进行快速检索，它也是通过几个 map 做相互映射实现。而我们外部是通过 `IndexFunc` 的定义进行控制反转， `IndexFunc` 是定义了该资源需要用什么字段作为索引值，如默认提供的索引方法返回的就是 `{namespace}` 这个字符串。

Indexer 使用的几种数据结构：

```go
// Index maps the indexed value to a set of keys in the store that match on that value
type Index map[string]sets.String

// Indexers maps a name to an IndexFunc
type Indexers map[string]IndexFunc

// Indices maps a name to an Index
type Indices map[string]Index
```

默认提供的 IndexFunc，构建通过 namespace 进行索引资源的索引器，当我们检索 namespace 下的资源时便可以使用该索引器建立索引与资源的存储关系：

```go
func MetaNamespaceIndexFunc(obj interface{}) ([]string, error) {
        meta, err := meta.Accessor(obj)
        if err != nil {
                return []string{""}, fmt.Errorf("object has no meta: %v", err)
        }
        return []string{meta.GetNamespace()}, nil
}
```

我们可以手动调用下带 Indexer 的 Store 是如何使用的，因为我是在源码内调试的所以我的包名是 `cache`：

```go
package cache

import (
        "fmt"
        v1 "k8s.io/api/core/v1"
        "k8s.io/apimachinery/pkg/api/meta"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "testing"
)

// LabelsIndexFunc 用作给出可检索所有的索引值
func LabelsIndexFunc(obj interface{}) ([]string, error) {
        metaD, err := meta.Accessor(obj)
        if err != nil {
                return []string{""}, fmt.Errorf("object has no meta: %v", err)
        }
        return []string{metaD.GetLabels()["app"]}, nil
}

func TestIndexer(t *testing.T) {
// 建立一个名为 app 的 Indexer, 并使用我们自己编写的 索引方法
        idxs := Indexers{"app": LabelsIndexFunc}

// 伪造2个pod资源
        pod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{
 Name:      "pod1",
 Namespace: "ns1",
 Labels: map[string]string{
  "app": "l1",
 }}}

        pod2 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{
 Name:      "pod2",
 Namespace: "ns2",
 Labels: map[string]string{
  "app": "l2",
 }}}
// 初始化 Indexer
        myIdx := NewIndexer(MetaNamespaceKeyFunc, idxs)
// 添加pod
        myIdx.Add(pod1)
        myIdx.Add(pod2)
// 打印通过索引检索的资源
        fmt.Println(myIdx.IndexKeys("app", "l1"))

}
// Output
// 结果只返回 app=l1 的 pod
// [ns1/pod1] <nil>

```

我们已经了解了 Informer 如何存储和检索资源。在调用 Informer 时，通常我们会看到许多不同的选项，例如 `NewInformer`、`NewIndexInfomer`、`NewShareInformer` 和 `NewShareIndexInformer` 等等。此外，还有其他几种选项没有列举出来。如果我们了解了上述内容，就会发现当我们看到 “Index” 这个词时，就知道我们可以传入自己构造的 Indexer。至于如何选择初始化方式，则取决于具体情况。 见 **3.4 章节**。

### **3.3.5 EventHandler 事件处理**

从前面几小节的内容可以看出，我们一直在接收变更事件并将其存储起来，以实现本地存储和远程存储的一致，从而减少对 API Server 的请求压力。不过，我们还需要考虑如何处理这些事件。接下来，我们将通过一个简单的例子来解释这一过程，并对源代码进行一些分析。

```go
package main

import (
        "fmt"
        "k8s-clientset/config"
        "k8s.io/api/core/v1"
        "k8s.io/apimachinery/pkg/fields"
        "k8s.io/apimachinery/pkg/util/wait"
        "k8s.io/client-go/tools/cache"
)

type CmdHandler struct {
}

// 当接收到添加事件便会执行该回调, 后面的方法以此类推
func (this *CmdHandler) OnAdd(obj interface{}) {
        fmt.Println("Add: ", obj.(*v1.ConfigMap).Name)
}

func (this *CmdHandler) OnUpdate(obj interface{}, newObj interface{}) {
        fmt.Println("Update: ", newObj.(*v1.ConfigMap).Name)
}

func (this *CmdHandler) OnDelete(obj interface{}) {
        fmt.Println("Delete: ", obj.(*v1.ConfigMap).Name)
}

func main() {
        cliset := config.NewK8sConfig().InitClient()
// 通过 clientset 返回一个 listwatcher, 仅支持 default/configmaps 资源
        listWatcher := cache.NewListWatchFromClient(
                cliset.CoreV1().RESTClient(),
                "configmaps",
                "default",
                fields.Everything(),
        )
// 初始化一个informer, 传入了监听器, 资源名, 间隔同步时间
// 最后一个是我们定义的 Handler 用于接收我们监听的资源变更事件;
        _, c := cache.NewInformer(listWatcher, &v1.ConfigMap{}, 0, &CmdHandler{})

// 启动循环监听
        c.Run(wait.NeverStop)
}
```

通过上面的例子，我们可以监听集群中 default/configmaps 资源的变更。它实际上接收变化的方式与前面的一些调试例子类似，但为了更加直观，我们可以直接看一下源代码是如何实现的。我删除了一些不必要的代码，只保留了重要的部分。完整的代码路径为 `client-go/tools/cache/controller.go`。在 `processDeltas` 的外层，有一个 `processLoop` 循环，它会不断地从队列中抛出事件，使得 `handler` 可以持续地流式处理事件。

```go
func processDeltas(
        handler ResourceEventHandler,
        clientState Store,
        transformer TransformFunc,
        deltas Deltas,
) error {
// from oldest to newest
        for _, d := range deltas {
                ...
                switch d.Type {
                case Sync, Replaced, Added, Updated:
                        if old, exists, err := clientState.Get(obj); err == nil && exists {
                                if err := clientState.Update(obj); err != nil {
                                        return err
                                }
                                handler.OnUpdate(old, obj)
                        } else {
                                if err := clientState.Add(obj); err != nil {
                                        return err
                                }
                                handler.OnAdd(obj)
                        }
                case Deleted:
                        if err := clientState.Delete(obj); err != nil {
                                return err
                        }
                        handler.OnDelete(obj)
                }
        }
        return nil
}
```

### **3.4 熟能生巧**

### **3.4.1 入门技巧**

上文提到 Informer 有非常多的初始化方式，本小节主要介绍 `NewInformer`、 `NewShareInformer` 和 `NewIndexInformer`。

### **>> NewInformer**

在 **[3.3.5 章节]** 中，我们介绍了 EventHandler 并演示了如何使用 `NewInformer` 方法创建 Informer。实际上，Informer 会向我们返回两个对象：`Store` 和 `Controller`。其中，Controller 主要用于控制监听事件的循环过程，而 Store 对象实际上与之前所讲的内容相同，我们可以直接从本地缓存中获取我们所监听的资源。在这个过程中，我们不需要担心数据的缺失或错误，因为 Informer 的监听机制可以保证数据的一致性。

参考示例：

```go
...
...
func main () {
        cliset := config.NewK8sConfig().InitClient()
// 获取configmap
        listWatcher := cache.NewListWatchFromClient(
                cliset.CoreV1().RESTClient(),
                "configmaps",
                "default",
                fields.Everything(),
        )
// CmdHandler 和上述的 EventHandler (参考 3.3.5)
        store, controller := cache.NewInformer(listWatcher, &v1.ConfigMap{}, 0, &CmdHandler{})
// 开启一个goroutine 避免主线程堵塞
        go controller.Run(wait.NeverStop)
// 等待3秒 同步缓存
        time.Sleep(3 * time.Second)
// 从缓存中获取监听到的 configmap 资源
        fmt.Println(store.List())

}

// Output:
// Add:  kube-root-ca.crt
// Add:  istio-ca-root-cert
// [... configmap 对象]
```

### **>> NewIndexInformer**

在 NewInformer 基础上接收 Indexer，注意这次我们例子中把资源变更 Pod，在 EventHandler 中的类型转换也要进行变成 Pod。

```go
import (
    "fmt"
    "k8s-clientset/config"
    "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/meta"
    "k8s.io/apimachinery/pkg/fields"
    "k8s.io/apimachinery/pkg/util/wait"
    "k8s.io/client-go/tools/cache"
    "time"
)

...

// LabelsIndexFunc 用作给出可检索的索引值
func LabelsIndexFunc(obj interface{}) ([]string, error) {
        metaD, err := meta.Accessor(obj)
        if err != nil {
                return []string{""}, fmt.Errorf("object has no meta: %v", err)
        }
        return []string{metaD.GetLabels()["app"]}, nil
}

func main () {
        cliset := config.NewK8sConfig().InitClient()
// 获取configmap
        listWatcher := cache.NewListWatchFromClient(
                cliset.CoreV1().RESTClient(),
                "configmaps",
                "default",
                fields.Everything(),
        )
// 创建索引其并指定名字
        myIndexer := cache.Indexers{"app": LabelsIndexFunc}
// CmdHandler 和上述的 EventHandler (参考 3.3.5)
        i, c := cache.NewIndexerInformer(listWatcher, &v1.Pod{}, 0, &CmdHandler{}, myIndexer)
// 开启一个goroutine 避免主线程堵塞
        go controller.Run(wait.NeverStop)
// 等待3秒 同步缓存
        time.Sleep(3 * time.Second)
// 通过 IndexStore 指定索引器获取我们需要的索引值
// busy-box 索引值是由于 我在某个 pod 上打了一个 label 为 app: busy-box
        objList, err := i.ByIndex("app", "busy-box")
        if err != nil {
                panic(err)
        }

        fmt.Println(objList[0].(*v1.Pod).Name)

}

// Output:
// Add:  cloud-enterprise-7f84df95bc-7vwxb
// Add:  busy-box-6698d6dff6-jmwfs
// busy-box-6698d6dff6-jmwfs
//
```

### **>> NewSharedInformer**

Share Informer 和 Informer 的主要区别就是可以添加多个 EventHandler，代码比较类似我就只展示重要的部分：

```go
...
...
func main() {
        cliset := config.NewK8sConfig().InitClient()
        listWarcher := cache.NewListWatchFromClient(
                cliset.CoreV1().RESTClient(),
                "configmaps",
                "default",
                fields.Everything(),
        )
// 全量同步时间
        shareInformer := cache.NewSharedInformer(listWarcher, &v1.ConfigMap{}, 0)
// 可以增加多个Event handler
        shareInformer.AddEventHandler(&handlers.CmdHandler{})
        shareInformer.AddEventHandler(&handlers.CmdHandler2{})
        shareInformer.Run(wait.NeverStop)
}
```

最后 `NewSharedIndexInformer` 和 `NewSharedInformer` 的区别就是可以添加 Indexer。

### **3.4.2 大集合才是硬道理**

在开发云原生应用或者进行多集群管理时，我们通常需要监听更多的资源，甚至是所有可操作的资源。因此，我们需要介绍一种更加灵活的 Informer 创建方式——`NewSharedInformerFactoryWithOptions`。使用该方法可以创建一个 Informer 工厂对象，在该工厂对象启动前，我们可以向其中添加任意 Kubernetes 内置的资源以及任意 Indexer。 看代码演示：

```go
package main

import (
        "fmt"
        "k8s-clientset/config"
        "k8s-clientset/dc/handlers"
        "k8s.io/apimachinery/pkg/labels"
        "k8s.io/apimachinery/pkg/runtime/schema"
        "k8s.io/apimachinery/pkg/util/wait"
        "k8s.io/client-go/informers"
)

func main() {

        cliset := config.NewK8sConfig().InitClient()
        informerFactory := informers.NewSharedInformerFactoryWithOptions(
                cliset,
                0,
// 指定的namespace 空间，如果需要所有空间，则不指定该参数
                informers.WithNamespace("default"),
        )
// 添加 ConfigMap 资源
        cmGVR := schema.GroupVersionResource{
                Group:    "",
                Version:  "v1",
                Resource: "configmaps",
        }
        cmInformer, _ := informerFactory.ForResource(cmGVR)
// 增加对 ConfigMap 事件的处理
        cmInformer.Informer().AddEventHandler(&handlers.CmdHandler{})

// 添加 Pod 资源
        podGVR := schema.GroupVersionResource{
                Group:    "",
                Version:  "v1",
                Resource: "pods",
        }
        _, _ = informerFactory.ForResource(podGVR)

// 启动 informerFactory
        informerFactory.Start(wait.NeverStop)
// 等待所有资源完成本地同步
        informerFactory.WaitForCacheSync(wait.NeverStop)

// 打印资源信息
        listConfigMap, _ := informerFactory.Core().V1().ConfigMaps().Lister().List(labels.Everything())
        fmt.Println("Configmap:")
        for _, obj := range listConfigMap {
                fmt.Printf("%s/%s \n", obj.Namespace, obj.Name)
        }
        fmt.Println("Pod:")
        listPod, _ := informerFactory.Core().V1().Pods().Lister().List(labels.Everything())
        for _, obj := range listPod {
                fmt.Printf("%s/%s \n", obj.Namespace, obj.Name)
        }
        select {}
}

// Ouput:

// Configmap:
// default/istio-ca-root-cert
// default/kube-root-ca.crt
// default/my-config
// Pod:
// default/cloud-enterprise-7f84df95bc-csdqp
// default/busy-box-6698d6dff6-42trb
```

如果想监听所有可操作的内部资源，可以使用 `DiscoveryClient` 去获取当前集群的资源版本再调用 `InformerFactory` 进行资源缓存。

### **3.5 埋点坑**

- Informer 获取的资源对象会丢失的 Kind 和 Version，该如何解决？
- Informer 在通过信号停止后，它却没有清理已占用的缓存，该如何在不重启的情况下清理膨胀的缓存？
