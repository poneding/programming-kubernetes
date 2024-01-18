package main

import (
	"log"

	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func prioritize(args extenderv1.ExtenderArgs) *extenderv1.HostPriorityList {
	log.Println("my-scheduler-extender prioritize called.")

	var result extenderv1.HostPriorityList
	for i, node := range args.Nodes.Items {
		result = append(result, extenderv1.HostPriority{
			Host:  node.Name,
			Score: int64(i),
		})
	}

	return &result
}
