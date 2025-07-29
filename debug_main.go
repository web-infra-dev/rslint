package main

import (
    "fmt"
    "github.com/typescript-eslint/rslint/internal/rules/adjacent_overload_signatures"
    "github.com/typescript-eslint/rslint/internal/ts"
)

func main() {
    code := `
function foo(s: string);
function foo(n: number);
function bar(): void {}
function foo(sn: string | number) {}
`
    diagnostics, err := ts.LintText(code, adjacent_overload_signatures.AdjacentOverloadSignaturesRule, nil)
    if err \!= nil {
        panic(err)
    }
    
    fmt.Printf("Number of diagnostics: %d\n", len(diagnostics))
    for _, d := range diagnostics {
        fmt.Printf("Diagnostic: %+v\n", d)
    }
}
EOF < /dev/null