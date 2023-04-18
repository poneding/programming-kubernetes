package main

import (
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
)

func main() {
	// fields.Set 是一个 map[string]string 类型，用于存储字段选择器的键值对
	fs := fields.Set{"foo": "bar", "baz": "qux"}

	// 从 fields.Set 中创建一个 fields.Selector： foo=bar,baz=qux
	sel := fields.SelectorFromSet(fs)

	if sel.Matches(fs) {
		fmt.Printf("Selector %v matches field set %v\n", sel, fs)
	} else {
		fmt.Printf("Selector %v does not match field set %v\n", sel, fs)
	}

	// 从 kv 对中创建一个 fields.Selector： foo=bar
	sel = fields.OneTermEqualSelector("foo", "bar")
	if sel.Matches(fs) {
		fmt.Printf("Selector %v matches field set %v\n", sel, fs)
	} else {
		fmt.Printf("Selector %v does not match field set %v\n", sel, fs)
	}

	// 从 kv 对中创建一个 fields.Selector： foo != bar2
	sel = fields.OneTermNotEqualSelector("foo", "bar2")
	if sel.Matches(fs) {
		fmt.Printf("Selector %v matches field set %v\n", sel, fs)
	} else {
		fmt.Printf("Selector %v does not match field set %v\n", sel, fs)
	}

	// 从 kv 对中创建一个 fields.Selector： foo=bar & baz=qux
	sel = fields.AndSelectors(fields.OneTermEqualSelector("foo", "bar"),
		fields.OneTermEqualSelector("baz", "qux"))
	if sel.Matches(fs) {
		fmt.Printf("Selector %v matches field set %v\n", sel, fs)
	} else {
		fmt.Printf("Selector %v does not match field set %v\n", sel, fs)
	}

	// 解析表达式，得到一个 fields.Selector
	sel = fields.ParseSelectorOrDie("foo=bar,baz=qux")
	if sel.Matches(fs) {
		fmt.Printf("Selector %v matches field set %v\n", sel, fs)
	} else {
		fmt.Printf("Selector %v does not match field set %v\n", sel, fs)
	}
}
