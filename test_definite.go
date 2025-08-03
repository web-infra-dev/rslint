package main

import (
	"fmt"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func main() {
	sourceText := "class A { a\!: number = 1; }"
	sourceFile, _ := parser.ParseSourceFile("test.ts", sourceText, false, false, nil)
	classDecl := sourceFile.AsSourceFile().Statements.Nodes[0].AsClassDeclaration()
	prop := classDecl.Members()[0].AsPropertyDeclaration()
	
	fmt.Printf("Property name: %v\n", prop.Name().AsIdentifier().Text)
	fmt.Printf("Has type: %v\n", prop.Type \!= nil)
	fmt.Printf("Has initializer: %v\n", prop.Initializer \!= nil)
	fmt.Printf("PostfixToken: %v\n", prop.PostfixToken \!= nil)
	if prop.PostfixToken \!= nil {
		fmt.Printf("PostfixToken kind: %v\n", prop.PostfixToken.Kind)
	}
	fmt.Printf("ExclamationToken: %v\n", prop.ExclamationToken \!= nil)
	if prop.ExclamationToken \!= nil {
		fmt.Printf("ExclamationToken kind: %v\n", prop.ExclamationToken.Kind)
	}
}
EOF < /dev/null