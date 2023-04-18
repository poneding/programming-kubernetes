package main

import (
	"fmt"
	"net/http"
	"time"

	"pk/basic/client"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	sharedInformers informers.SharedInformerFactory
)

func init() {
	setupInformers()
}

func setupInformers() {
	konfig := ctrl.GetConfigOrDie()
	fmt.Printf("konfig.Host: %v\n", konfig.Host)
	fmt.Printf("konfig.ServerName: %v\n", konfig.ServerName)
	fmt.Printf("konfig.APIPath: %v\n", konfig.APIPath)
	clientset := client.Clientset(konfig)

	sharedInformers = informers.NewSharedInformerFactory(clientset, time.Minute*5)

	gvrs := []schema.GroupVersionResource{
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "daemonsets"},
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "namespaces"},
		{Group: "", Version: "v1", Resource: "configmaps"},
		{Group: "", Version: "v1", Resource: "secrets"},
	}

	for _, gvr := range gvrs {
		_, err := sharedInformers.ForResource(gvr)
		if err != nil {
			panic(err)
		}
	}
	sharedInformers.Start(wait.NeverStop)
	sharedInformers.WaitForCacheSync(wait.NeverStop)
	fmt.Println("synced done.")
}

func main() {
	e := gin.Default()
	e.GET("/pods", listPods)
	e.GET("/:g/:v/:r", listResources)

	e.Run(":8888")
}

func listPods(c *gin.Context) {
	pods, err := sharedInformers.Core().V1().Pods().Lister().Pods("default").List(labels.Everything())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
		return
	}

	var res []string
	for _, pod := range pods {
		res = append(res, pod.Name)
	}
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  fmt.Sprintln(res),
	})
}

func listResources(c *gin.Context) {
	g := c.Param("g")
	v := c.Param("v")
	r := c.Param("r")

	gvr := schema.GroupVersionResource{
		Group:    g,
		Version:  v,
		Resource: r,
	}
	gi, err := sharedInformers.ForResource(gvr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
	}

	objs, err := gi.Lister().List(labels.Everything())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
	}

	var res []interface{}
	for _, obj := range objs {
		unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  err.Error(),
			})
			return
		}

		name, ok, err := unstructured.NestedString(unstructuredObj, "metadata", "name")
		if err != nil || !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "get name failed",
			})
			return
		}
		namespace, _, _ := unstructured.NestedString(unstructuredObj, "metadata", "namespace")
		res = append(res, struct {
			Group     string `json:"group"`
			Version   string `json:"version"`
			Kind      string `json:"kind"`
			Name      string `json:"name"`
			Namespace string `json:"namespace,omitempty"`
		}{
			Group:     obj.GetObjectKind().GroupVersionKind().Group,
			Version:   obj.GetObjectKind().GroupVersionKind().Version,
			Kind:      obj.GetObjectKind().GroupVersionKind().Kind,
			Name:      name,
			Namespace: namespace,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  res,
	})
}
