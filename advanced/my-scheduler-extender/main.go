package main

import (
	"encoding/json"
	"log"
	"net/http"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func Filter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var args extenderv1.ExtenderArgs
	var result *extenderv1.ExtenderFilterResult

	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		result = &extenderv1.ExtenderFilterResult{
			Error: err.Error(),
		}
	} else {
		result = filter(args)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("failed to encode result: %v", err)
	}
}

func Prioritize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var args extenderv1.ExtenderArgs
	var result *extenderv1.HostPriorityList

	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		result = &extenderv1.HostPriorityList{}
	} else {
		result = prioritize(args)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("failed to encode result: %v", err)
	}
}

func Bind(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var args extenderv1.ExtenderBindingArgs
	var result *extenderv1.ExtenderBindingResult

	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		result = &extenderv1.ExtenderBindingResult{
			Error: err.Error(),
		}
	} else {
		result = bind(args)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("failed to encode result: %v", err)
	}
}

func main() {
	http.HandleFunc("/filter", Filter)
	http.HandleFunc("/priority", Prioritize)
	http.HandleFunc("/bind", Bind)
	http.ListenAndServe(":8080", nil)
}
