package main

import (
	"fmt"
	"strings"
)

func testCase(name string, code string) {
	fmt.Printf("\n=== Testing %s ===\n", name)
	fmt.Printf("Code: %s\n", code)
	
	// Find function position
	funcPos := -1
	if strings.Contains(code, "function foo") {
		funcPos = strings.Index(code, "function foo")
	} else if strings.Contains(code, "function ()") {
		funcPos = strings.Index(code, "function ()")
	}
	
	if funcPos == -1 {
		fmt.Println("No function found")
		return
	}
	
	fmt.Printf("Function position: %d\n", funcPos)
	
	// Simulate search area
	searchStart := 0
	if funcPos > 500 {
		searchStart = funcPos - 500
	}
	searchText := code[searchStart:funcPos]
	
	fmt.Printf("Search text: %q\n", searchText)
	fmt.Printf("Contains @this: %v\n", strings.Contains(searchText, "@this"))
	
	if strings.Contains(searchText, "@this") {
		thisIdx := strings.Index(searchText, "@this")
		beforeThis := searchText[:thisIdx]
		afterThis := searchText[thisIdx:]
		
		fmt.Printf("@this at: %d\n", thisIdx)
		fmt.Printf("Before: %q\n", beforeThis)
		fmt.Printf("After: %q\n", afterThis)
		
		// Check JSDoc
		jsdocStart := strings.LastIndex(beforeThis, "/**")
		if jsdocStart != -1 {
			jsdocEnd := strings.Index(afterThis, "*/")
			if jsdocEnd != -1 {
				noEndBetween := strings.Index(beforeThis[jsdocStart:], "*/") == -1
				fmt.Printf("JSDoc found: start=%d, end=%d, no end between=%v\n", jsdocStart, jsdocEnd, noEndBetween)
				if noEndBetween {
					fmt.Println("*** SHOULD BE VALID ***")
				}
			}
		}
		
		// Check block comment
		blockStart := strings.LastIndex(beforeThis, "/*")
		if blockStart != -1 {
			blockEnd := strings.Index(afterThis, "*/")
			if blockEnd != -1 {
				noEndBetween := strings.Index(beforeThis[blockStart:], "*/") == -1
				fmt.Printf("Block comment found: start=%d, end=%d, no end between=%v\n", blockStart, blockEnd, noEndBetween)
				if noEndBetween {
					fmt.Println("*** SHOULD BE VALID ***")
				}
			}
		}
	}
}

func main() {
	// Test cases
	testCase("valid-39", `
/** @this Obj */ function foo() {
  console.log(this);
  z(x => console.log(x, this));
}`)

	testCase("valid-40", `
foo(
  /* @this Obj */ function () {
    console.log(this);
    z(x => console.log(x, this));
  },
);`)

	testCase("invalid-33", `
/** @this Obj */ foo(function () {
  console.log(this);
  z(x => console.log(x, this));
});`)
}