package main

import (
	"fmt"

	"github.com/tailscale/hujson"
)

func main() {
	input := `{"key": "value"} // comment`
	ast, err := hujson.Parse([]byte(input))
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}
	ast.Standardize()
	result := string(ast.Pack())
	fmt.Printf("Input: %q\n", input)
	fmt.Printf("Result: %q\n", result)
}
