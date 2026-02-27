// Package util 为 cmd 包提供共享工具。
package util

import (
	"encoding/json"
	"fmt"
)

// PrintResult 将结果打印到 stdout。
func PrintResult(v interface{}) {
	switch t := v.(type) {
	case string:
		fmt.Println(t)
	case nil:
		fmt.Println("null")
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Println(v)
			return
		}
		fmt.Println(string(b))
	}
}
