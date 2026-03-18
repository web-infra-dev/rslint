package no_new_wrappers

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-new-wrappers
var wrapperObjects = map[string]bool{
	"String":  true,
	"Number":  true,
	"Boolean": true,
}

var NoNewWrappersRule = rule.Rule{
	Name: "no-new-wrappers",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				if newExpr == nil {
					return
				}

				callee := newExpr.Expression
				if callee == nil || callee.Kind != ast.KindIdentifier {
					return
				}

				name := callee.Text()
				if !wrapperObjects[name] {
					return
				}

				// Use TypeChecker to verify this refers to the global built-in,
				// not a locally shadowed variable/parameter/import.
				if ctx.TypeChecker != nil {
					sym := ctx.TypeChecker.GetSymbolAtLocation(callee)
					if sym != nil && isLocallyShadowed(sym, ctx.SourceFile) {
						return
					}
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noConstructor",
					Description: fmt.Sprintf("Do not use %s as a constructor.", name),
				})
			},
		}
	},
}

// isLocallyShadowed checks if a symbol is declared in a user source file
// (not a .d.ts lib file), meaning it's a local variable that shadows the global.
func isLocallyShadowed(sym *ast.Symbol, currentFile *ast.SourceFile) bool {
	decl := sym.ValueDeclaration
	if decl == nil {
		return false
	}

	// Walk up to find the SourceFile containing this declaration
	current := decl
	for current != nil && current.Kind != ast.KindSourceFile {
		current = current.Parent
	}
	if current == nil {
		return false
	}

	declFile := current.AsSourceFile()
	if declFile == nil {
		return false
	}

	// If declared in a .d.ts file, it's a lib declaration (global built-in)
	fileName := declFile.FileName()
	if strings.HasSuffix(fileName, ".d.ts") {
		return false
	}

	// If declared in the current file (or another non-.d.ts file), it's local
	return true
}
