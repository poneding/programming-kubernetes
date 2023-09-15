package main

import (
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"golang.org/x/exp/maps"
	v1 "k8s.io/api/admissionregistration/v1"
	apidiscovery "k8s.io/api/apidiscovery/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"net/http"
	"os"
	"path"
	"pk/examples/my-apiserver/generated"
	"strings"
)

func main() {
	r := gin.New()
	// kube-apiserver 探活
	r.GET("/openapi/v2", OpenAPI)
	r.GET("/apis", APIs)
	r.GET("/apis/play.poneding.com", APIGroup)
	r.GET("/apis/play.poneding.com/v1", APIResources)
	r.GET("/apis/play.poneding.com/v1/foos", AllFoos)
	r.GET("/apis/play.poneding.com/v1/namespaces/:namespace/foos", AllNamespacedFoos)
	r.GET("/apis/play.poneding.com/v1/namespaces/:namespace/foos/:name", GetFoo)
	r.POST("/apis/play.poneding.com/v1/namespaces/:namespace/foos", CreateFoo)
	r.PUT("/apis/play.poneding.com/v1/namespaces/:namespace/foos/:name", UpdateFoo)
	r.PATCH("/apis/play.poneding.com/v1/namespaces/:namespace/foos/:name", PatchFoo)
	r.DELETE("/apis/play.poneding.com/v1/namespaces/:namespace/foos/:name", DeleteFoo)

	if certPath := os.Getenv("CERT_DIR"); certPath != "" {
		crt := path.Join(certPath, "tls.crt")
		key := path.Join(certPath, "tls.key")
		r.RunTLS(":8443", crt, key)
	} else {
		r.Run(":8080")
	}
}

var apiGroup = metav1.APIGroup{
	TypeMeta: metav1.TypeMeta{
		Kind:       "APIGroup",
		APIVersion: "v1",
	},
	Name: "hello.zeng.dev",
	Versions: []metav1.GroupVersionForDiscovery{
		{
			GroupVersion: "play.poneding.com/v1",
			Version:      "v1",
		},
	},
	PreferredVersion: metav1.GroupVersionForDiscovery{GroupVersion: "play.poneding.com/v1", Version: "v1"},
}

var apiGroups = metav1.APIGroupList{
	TypeMeta: metav1.TypeMeta{
		Kind:       "APIGroupList",
		APIVersion: "v1",
	},
	Groups: []metav1.APIGroup{apiGroup},
}

var apiGroupDiscoveries = apidiscovery.APIGroupDiscoveryList{
	Items: []apidiscovery.APIGroupDiscovery{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "play.poneding.com",
			},
			Versions: []apidiscovery.APIVersionDiscovery{
				{
					Version: "v1",
					Resources: []apidiscovery.APIResourceDiscovery{
						{
							Resource: "foos",
							ResponseKind: &metav1.GroupVersionKind{
								Group:   "play.poneding.com",
								Kind:    "Foo",
								Version: "v1",
							},
							Scope:            apidiscovery.ResourceScope(v1.NamespacedScope),
							ShortNames:       []string{"fo"},
							SingularResource: "foo",
							Verbs: []string{
								"delete",
								"get",
								"list",
								"patch",
								"create",
								"update",
							},
						},
					},
				},
			},
		},
	},
}

var apiResources = metav1.APIResourceList{
	GroupVersion: "play.poneding.com/v1",
	TypeMeta: metav1.TypeMeta{
		Kind:       "APIResourceList",
		APIVersion: "v1",
	},
	APIResources: []metav1.APIResource{
		{
			Group:        "play.poneding.com",
			Version:      "v1",
			Name:         "foos",
			Kind:         "Foo",
			Namespaced:   true,
			SingularName: "foo",
			ShortNames:   []string{"fo"},
			Categories:   []string{"all"},
			Verbs: []string{
				"delete",
				"get",
				"list",
				"patch",
				"create",
				"update",
			},
		},
	},
}

// +k8s:openapi-gen=true
type Foo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FooSpec `json:"spec,omitempty"`
}

type FooSpec struct {
	Msg string `json:"msg,omitempty"`
}

type FooList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Foo `json:"items"`
}

var resourceInstances = map[string]Foo{
	"default/init-foo": {
		TypeMeta: metav1.TypeMeta{
			APIVersion: "play.poneding.com/v1",
			Kind:       "Foo",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "init-foo",
			Namespace:         "default",
			ResourceVersion:   "1",
			CreationTimestamp: metav1.Now(),
		},
		Spec: FooSpec{
			Msg: "Hello init-foo!",
		},
	},
}

func OpenAPI(ctx *gin.Context) {
	definitions := generated.GetOpenAPIDefinitions(func(ref string) spec.Ref {
		r, _ := spec.NewRef(ref)
		return r
	})
	ctx.JSON(http.StatusOK, definitions)
}

// @Router /apis [get]
func APIs(ctx *gin.Context) {
	var gvk [3]string
	for _, acceptPart := range strings.Split(ctx.GetHeader("Accept"), ";") {
		if g_v_k := strings.Split(acceptPart, "="); len(g_v_k) == 2 {
			switch g_v_k[0] {
			case "g":
				gvk[0] = g_v_k[1]
			case "v":
				gvk[1] = g_v_k[1]
			case "as":
				gvk[2] = g_v_k[1]
			}
		}
	}

	if gvk[0] == "apidiscovery.k8s.io" && gvk[2] == "APIGroupDiscoveryList" {
		ctx.JSON(http.StatusOK, apiGroupDiscoveries)
	} else {
		ctx.JSON(http.StatusOK, apiGroups)
	}
}

func APIGroup(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, apiGroup)
}

func APIResources(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, apiResources)
}

func AllFoos(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, &FooList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "play.poneding.com/v1",
			Kind:       "FooList",
		},
		Items: maps.Values(resourceInstances),
	})
}

func AllNamespacedFoos(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	foos := make([]Foo, 0)
	for _, foo := range resourceInstances {
		if foo.Namespace == namespace {
			foos = append(foos, foo)
		}
	}

	ctx.JSON(http.StatusOK, &FooList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "play.poneding.com/v1",
			Kind:       "FooList",
		},
		Items: foos,
	})
}

func GetFoo(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	if resource, ok := resourceInstances[namespace+"/"+name]; ok {
		ctx.JSON(http.StatusOK, resource)
	} else {
		ctx.JSON(http.StatusNotFound, nil)
	}
}

func CreateFoo(ctx *gin.Context) {
	var foo Foo
	if err := ctx.ShouldBindJSON(&foo); err != nil {
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	if _, ok := resourceInstances[fooKey(foo)]; ok {
		ctx.JSON(http.StatusConflict, nil)
		return
	}
	foo.CreationTimestamp = metav1.Now()
	foo.ResourceVersion = "1"
	resourceInstances[fooKey(foo)] = foo
	ctx.JSON(http.StatusOK, foo) // 遵循 k8s 的规范，创建资源时，返回资源对象
}

func UpdateFoo(ctx *gin.Context) {
	var foo Foo
	if err := ctx.ShouldBindJSON(&foo); err != nil {
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	if _, ok := resourceInstances[fooKey(foo)]; !ok {
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	foo.ResourceVersion = cast.ToString(cast.ToInt(foo.ResourceVersion) + 1)
	resourceInstances[fooKey(foo)] = foo
	ctx.JSON(http.StatusOK, foo) // 遵循 k8s 的规范，更新资源时，返回资源对象
}

func PatchFoo(ctx *gin.Context) {
	var namespace = ctx.Param("namespace")
	var name = ctx.Param("name")
	patch, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	foo, ok := resourceInstances[namespace+"/"+name]
	if !ok {
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	marshal, err := json.Marshal(foo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	mergePatch, err := jsonpatch.MergePatch(marshal, patch)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	var patchedFoo Foo
	err = json.Unmarshal(mergePatch, &patchedFoo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	patchedFoo.ResourceVersion = cast.ToString(cast.ToInt(patchedFoo.ResourceVersion) + 1)
	resourceInstances[fooKey(patchedFoo)] = patchedFoo
	ctx.JSON(http.StatusOK, patchedFoo)
}

func DeleteFoo(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	foo, ok := resourceInstances[namespace+"/"+name]
	if !ok {
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	delete(resourceInstances, namespace+"/"+name)

	ctx.JSON(http.StatusOK, foo) // 遵循 k8s 的规范，删除资源时，返回资源对象
}

func fooKey(foo Foo) string {
	return foo.Namespace + "/" + foo.Name
}
