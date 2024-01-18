package main

import (
	"log"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func filter(args extenderv1.ExtenderArgs) *extenderv1.ExtenderFilterResult {
	log.Println("my-scheduler-extender filter called.")

	return &extenderv1.ExtenderFilterResult{
		Nodes:     args.Nodes,
		NodeNames: args.NodeNames,
	}
}
