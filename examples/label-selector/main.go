package main

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func main() {
	// labels.Set 是一个 map[string]string 类型，用于存储标签选择器的键值对
	ls := labels.Set{"foo": "bar", "baz": "qux"}

	// 从 labels.Set 中创建一个 labels.Selector： foo=bar,baz=qux
	sel := labels.SelectorFromSet(ls)

	if sel.Matches(ls) {
		fmt.Printf("Selector %v matches labels set %v\n", sel, ls)
	} else {
		fmt.Printf("Selector %v does not match labels set %v\n", sel, ls)
	}

	sel = labels.NewSelector()
	req, _ := labels.NewRequirement("foo", selection.Equals, []string{"bar"})

	sel.Add(*req)
	if sel.Matches(ls) {
		fmt.Printf("Selector %v matches labels set %v\n", sel, ls)
	} else {
		fmt.Printf("Selector %v does not match labels set %v\n", sel, ls)
	}

	// 解析表达式，得到一个 labels.Selector
	sel, _ = labels.Parse("foo=bar,baz=qux")
	if sel.Matches(ls) {
		fmt.Printf("Selector %v matches labels set %v\n", sel, ls)
	} else {
		fmt.Printf("Selector %v does not match labels set %v\n", sel, ls)
	}

	fmt.Println()
	// 比较两个 lables 是否相等
	l1 := labels.Set{"foo": "bar", "baz": "qux"}
	l2 := labels.Set{"baz": "qux", "foo": "bar"}
	if labels.Equals(l1, l2) {
		fmt.Printf("l1 %v equals l2 %v\n", l1, l2)
	} else {
		fmt.Printf("l1 %v not equals l2 %v\n", l1, l2)
	}
}
