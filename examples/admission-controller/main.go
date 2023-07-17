package main

import (
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"log"
	"net/http"
	"os"
)

const (
	certFile = "/etc/webhook/certs/tls.crt"
	keyFile  = "/etc/webhook/certs/tls.key"
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
}

func main() {
	http.HandleFunc("/mutating", mutatePod)
	klog.Info("start serving admission webhook...")
	if err := http.ListenAndServeTLS(":443", certFile, keyFile, nil); err != nil {
		log.Fatalln("failed to listen and serve admission webhook.", err.Error())
	}
}

func mutatePod(w http.ResponseWriter, r *http.Request) {
	log.Println("request admission webhook mutating...")
	var (
		reviewRequest, reviewResponse v1.AdmissionReview
		pod                           = &corev1.Pod{}
	)

	if err := json.NewDecoder(r.Body).Decode(&reviewRequest); err != nil {
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

	if v, ok := pod.Labels["pk.poneding.com/echo-hello-sidecar"]; ok && v == "true" {
		reviewResponse.Response.Result = &metav1.Status{
			Message: "matched pod labels pk.poneding.com/echo-hello-sidecar=true, migrate image to target registry.",
		}

		tryRegisterSidecar(pod)

		var patches = []map[string]any{
			{
				"op":    "replace",
				"path":  "/spec/initContainers",
				"value": pod.Spec.InitContainers,
			},
		}

		patchBytes, err := json.Marshal(patches)
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

func tryRegisterSidecar(pod *corev1.Pod) {
	for _, ic := range pod.Spec.InitContainers {
		if ic.Name == "echo-hello-sidecar" {
			return
		}
	}

	pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{
		Name:  "echo-hello-sidecar",
		Image: "busybox",
		Command: []string{
			"/bin/sh",
			"-c",
			"echo hello",
		},
	})
}
