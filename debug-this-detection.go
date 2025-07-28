package main

import (
	"fmt"
	"strings"
)

func main() {
	// Test case from valid-39
	testCode := `
/** @this Obj */ function foo() {
  console.log(this);
  z(x => console.log(x, this));
}
`
	
	// Simulate finding function position (rough estimate)
	funcPos := strings.Index(testCode, "function foo()")
	if funcPos == -1 {
		fmt.Println("Function not found")
		return
	}
	
	searchStart := 0 
	if funcPos > 500 {
		searchStart = funcPos - 500
	}
	searchText := testCode[searchStart:funcPos]
	
	fmt.Printf("Function position: %d\n", funcPos)
	fmt.Printf("Search text: %q\n", searchText)
	fmt.Printf("Contains @this: %v\n", strings.Contains(searchText, "@this"))
	
	if strings.Contains(searchText, "@this") {
		thisIdx := strings.Index(searchText, "@this")
		fmt.Printf("@this position in search text: %d\n", thisIdx)
		
		beforeThis := searchText[:thisIdx]
		afterThis := searchText[thisIdx:]
		
		fmt.Printf("Before @this: %q\n", beforeThis)
		fmt.Printf("After @this: %q\n", afterThis)
		
		// Check for JSDoc comment /** ... @this ... */
		jsdocStart := strings.LastIndex(beforeThis, "/**")
		fmt.Printf("JSDoc start position: %d\n", jsdocStart)
		
		if jsdocStart != -1 {
			jsdocEnd := strings.Index(afterThis, "*/")
			fmt.Printf("JSDoc end position: %d\n", jsdocEnd)
			
			if jsdocEnd != -1 {
				// Make sure there's no comment end between JSDoc start and @this
				noEndBetween := strings.Index(beforeThis[jsdocStart:], "*/") == -1
				fmt.Printf("No end between: %v\n", noEndBetween)
				
				if noEndBetween {
					fmt.Println("FOUND VALID @this JSDoc!")
				}
			}
		}
	}
}