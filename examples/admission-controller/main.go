package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	certFile = "/etc/webhook/certs/tls.crt"
	keyFile  = "/etc/webhook/certs/tls.key"
)

var (
	dockerAccount           = os.Getenv("DOCKER_ACCPOUNT")
	targetRegistryServer    = os.Getenv("TARGET_REGISTRY_SERVER")
	targetRegistryNamespace = os.Getenv("TARGET_REGISTRY_NAMESPACE")
)

func init() {
	_, err := os.Stat(certFile)
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = os.Stat(keyFile)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if targetRegistryServer == "" || targetRegistryNamespace == "" {
		log.Fatalln("env TARGET_REGISTRY_SERVER or TARGET_REGISTRY_NAMESPACE is empty.")
	}
}

func main() {
	http.HandleFunc("/mutating", mutatePod)
	klog.Info("start serving admission webhook...")
	if err := http.ListenAndServeTLS(":443", certFile, keyFile, nil); err != nil {
		log.Fatalln("failed to listen and serve admission webhook.", err.Error())
	}
}

func replcacePodImages(pod *corev1.Pod) {
	for i := range pod.Spec.Containers {
		pod.Spec.Containers[i].Image = parseImage(pod.Spec.Containers[i].Image)
	}

	for i := range pod.Spec.InitContainers {
		pod.Spec.InitContainers[i].Image = parseImage(pod.Spec.InitContainers[i].Image)
	}

	for i := range pod.Spec.EphemeralContainers {
		pod.Spec.EphemeralContainers[i].Image = parseImage(pod.Spec.EphemeralContainers[i].Image)
	}
}

func mutatePod(w http.ResponseWriter, r *http.Request) {
	log.Println("request admission webhook mutating...")
	var (
		reviewRequest, reviewResponse v1.AdmissionReview
		pod                           = &corev1.Pod{}
	)

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reviewRequest); err != nil {
		log.Println("decode body failed.")
		http.Error(w, fmt.Sprintf("could not decode body: %v", err), http.StatusBadRequest)
		return
	}

	var raw []byte
	if reviewRequest.Request.Operation == v1.Create || reviewRequest.Request.Operation == v1.Update {
		raw = reviewRequest.Request.Object.Raw
	} else {
		raw = reviewRequest.Request.OldObject.Raw
	}

	log.Println(string(raw))

	if err := json.Unmarshal(raw, pod); err != nil {
		log.Println("unmarshal pod object failed.", err.Error())
		http.Error(w, fmt.Sprintf("could not unmarshal pod object: %v", err), http.StatusBadRequest)
		return
	}

	reviewResponse.TypeMeta = reviewRequest.TypeMeta
	reviewResponse.Response = &v1.AdmissionResponse{
		UID:     reviewRequest.Request.UID,
		Allowed: true,
		Result:  nil,
		PatchType: func() *v1.PatchType {
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}

	if v, ok := pod.Labels["pk.io/webhook-test"]; ok && v == "true" {
		reviewResponse.Response.Result = &metav1.Status{
			Message: "matched pod labels pk.io/webhook-test=true, migrate image to target registry.",
		}

		replcacePodImages(pod)

		var patchs = []map[string]any{
			{
				"op":    "replace",
				"path":  "/spec/containers",
				"value": pod.Spec.Containers,
			},
			{
				"op":    "replace",
				"path":  "/spec/initContainers",
				"value": pod.Spec.InitContainers,
			},
			{
				"op":    "replace",
				"path":  "/spec/ephemeralContainers",
				"value": pod.Spec.EphemeralContainers,
			},
		}

		patchBytes, err := json.Marshal(patchs)
		if err != nil {
			log.Println("marshal patch failed.")
			http.Error(w, fmt.Sprintf("could not marshal patch: %v", err), http.StatusInternalServerError)
			return
		}

		reviewResponse.Response.Patch = patchBytes
	}

	if err := json.NewEncoder(w).Encode(reviewResponse); err != nil {
		log.Println("encode response failed.")
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func parseImage(image string) string {
	parts := strings.Split(image, "/")
	if len(parts) != 2 || parts[0] != dockerAccount {
		return image
	}

	return fmt.Sprintf("%s/%s/%s", targetRegistryServer, targetRegistryNamespace, parts[1])
}
