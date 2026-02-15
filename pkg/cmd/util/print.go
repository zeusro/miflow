// Package util provides shared utilities for cmd packages.
package util

import (
	"encoding/json"
	"fmt"
)

// PrintResult prints the result to stdout.
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
