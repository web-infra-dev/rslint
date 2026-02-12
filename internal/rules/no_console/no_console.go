package no_console

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-console
var NoConsoleRule = rule.Rule{
	Name: "no-console",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		reportIfConsole := func(node *ast.Node, consoleIdent *ast.Node, propertyName string) {
			// Check if this property is allowed
			if opts.isAllowed(propertyName) {
				return
			}

			// Check if console is shadowed by a local declaration
			if ctx.TypeChecker != nil {
				symbol := ctx.TypeChecker.GetSymbolAtLocation(consoleIdent)
				if symbol != nil {
					for _, declaration := range symbol.Declarations {
						declarationSourceFile := ast.GetSourceFileOfNode(declaration)
						if declarationSourceFile != nil && declarationSourceFile == ctx.SourceFile {
							return
						}
					}
				}
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpected",
				Description: "Unexpected console statement.",
			})
		}

		return rule.RuleListeners{
			// Handle console.log, console.warn, etc.
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				propAccess := node.AsPropertyAccessExpression()
				if propAccess == nil {
					return
				}

				if propAccess.Expression.Kind != ast.KindIdentifier {
					return
				}

				objectName := propAccess.Expression.AsIdentifier().Text
				if objectName != "console" {
					return
				}

				propertyName := propAccess.Name().Text()
				reportIfConsole(node, propAccess.Expression, propertyName)
			},

			// Handle console["log"], console["warn"], etc.
			ast.KindElementAccessExpression: func(node *ast.Node) {
				elemAccess := node.AsElementAccessExpression()
				if elemAccess == nil {
					return
				}

				if elemAccess.Expression.Kind != ast.KindIdentifier {
					return
				}

				objectName := elemAccess.Expression.AsIdentifier().Text
				if objectName != "console" {
					return
				}

				// Only report if the argument is a static string
				if elemAccess.ArgumentExpression == nil {
					return
				}
				if elemAccess.ArgumentExpression.Kind != ast.KindStringLiteral {
					return
				}

				propertyName := elemAccess.ArgumentExpression.AsStringLiteral().Text
				reportIfConsole(node, elemAccess.Expression, propertyName)
			},
		}
	},
}

type consoleOptions struct {
	allow map[string]bool
}

func (o *consoleOptions) isAllowed(method string) bool {
	return o.allow[method]
}

func parseOptions(opts any) consoleOptions {
	result := consoleOptions{
		allow: make(map[string]bool),
	}

	if opts == nil {
		return result
	}

	var optsMap map[string]interface{}
	if arr, ok := opts.([]interface{}); ok && len(arr) > 0 {
		optsMap, _ = arr[0].(map[string]interface{})
	} else {
		optsMap, _ = opts.(map[string]interface{})
	}

	if optsMap != nil {
		if allowArr, ok := optsMap["allow"].([]interface{}); ok {
			for _, item := range allowArr {
				if str, ok := item.(string); ok {
					result.allow[str] = true
				}
			}
		}
	}

	return result
}
