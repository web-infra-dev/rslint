package main

import (
	"fmt"
	"strings"
)

func debugProximity(name string, code string) {
	fmt.Printf("\n=== %s ===\n", name)
	
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
	
	// Simulate search area
	searchStart := 0
	if funcPos > 500 {
		searchStart = funcPos - 500
	}
	searchText := code[searchStart:funcPos]
	
	fmt.Printf("Function at position: %d\n", funcPos)
	fmt.Printf("Search text: %q\n", searchText)
	
	if strings.Contains(searchText, "@this") {
		thisIdx := strings.Index(searchText, "@this")
		beforeThis := searchText[:thisIdx]
		afterThis := searchText[thisIdx:]
		
		fmt.Printf("@this at: %d\n", thisIdx)
		
		// Check JSDoc
		jsdocStart := strings.LastIndex(beforeThis, "/**")
		if jsdocStart != -1 {
			jsdocEnd := strings.Index(afterThis, "*/")
			if jsdocEnd != -1 {
				noEndBetween := strings.Index(beforeThis[jsdocStart:], "*/") == -1
				if noEndBetween {
					// Check proximity
					commentEndPos := thisIdx + jsdocEnd + 2
					remainingText := strings.TrimSpace(searchText[commentEndPos:])
					
					fmt.Printf("Comment ends at: %d\n", commentEndPos)
					fmt.Printf("Remaining text: %q\n", remainingText)
					fmt.Printf("Starts with function: %v\n", strings.HasPrefix(remainingText, "function"))
					
					if len(remainingText) == 0 || strings.HasPrefix(remainingText, "function") {
						fmt.Println("*** SHOULD BE VALID (JSDoc) ***")
					} else {
						fmt.Println("*** REJECTED: Non-function text between comment and function ***")
					}
				}
			}
		}
		
		// Check block comment
		blockStart := strings.LastIndex(beforeThis, "/*")
		if blockStart != -1 {
			blockEnd := strings.Index(afterThis, "*/")
			if blockEnd != -1 {
				noEndBetween := strings.Index(beforeThis[blockStart:], "*/") == -1
				if noEndBetween {
					// Check proximity
					commentEndPos := thisIdx + blockEnd + 2
					remainingText := strings.TrimSpace(searchText[commentEndPos:])
					
					fmt.Printf("Block comment ends at: %d\n", commentEndPos)
					fmt.Printf("Remaining text: %q\n", remainingText)
					fmt.Printf("Starts with function: %v\n", strings.HasPrefix(remainingText, "function"))
					
					if len(remainingText) == 0 || strings.HasPrefix(remainingText, "function") {
						fmt.Println("*** SHOULD BE VALID (Block) ***")
					} else {
						fmt.Println("*** REJECTED: Non-function text between comment and function ***")
					}
				}
			}
		}
	}
}

func main() {
	debugProximity("valid-39", `
/** @this Obj */ function foo() {
  console.log(this);
  z(x => console.log(x, this));
}`)

	debugProximity("valid-40", `
foo(
  /* @this Obj */ function () {
    console.log(this);
    z(x => console.log(x, this));
  },
);`)

	debugProximity("invalid-33", `
/** @this Obj */ foo(function () {
  console.log(this);
  z(x => console.log(x, this));
});`)
}