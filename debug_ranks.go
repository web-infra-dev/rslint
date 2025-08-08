package main

import "fmt"

func main() {
	order := []string{
		"signature",
		"call-signature", 
		"public-static-field",
		"protected-static-field",
		"private-static-field",
		"#private-static-field",
		"public-decorated-field",
		"protected-decorated-field",
		"private-decorated-field",
		"public-instance-field",
		"protected-instance-field",
		"private-instance-field",
		"#private-instance-field",
		"public-abstract-field",
		"protected-abstract-field",
		"public-field",
		"protected-field",
		"private-field",
		"#private-field",
		"static-field",
		"instance-field",
		"abstract-field",
		"decorated-field",
		"field",
		"static-initialization",
		"public-constructor",
		"protected-constructor", 
		"private-constructor",
		"constructor",
	}
	for i, item := range order {
		if item == "signature" { 
			fmt.Printf("signature: %d\n", i) 
		}
		if item == "field" { 
			fmt.Printf("field: %d\n", i) 
		}
		if item == "constructor" { 
			fmt.Printf("constructor: %d\n", i) 
		}
	}
}