package no_console

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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

			// Check if console is shadowed by a local declaration.
			// TypeChecker is expected to always be available at runtime.
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

			// Handle console["log"], console[foo], etc.
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

				// Try to get static property name for allow-list check.
				// Dynamic access (e.g., console[foo]) is always reported since
				// we can't prove it's in the allow list.
				var propertyName string
				if elemAccess.ArgumentExpression != nil {
					switch elemAccess.ArgumentExpression.Kind {
					case ast.KindStringLiteral:
						propertyName = elemAccess.ArgumentExpression.AsStringLiteral().Text
					case ast.KindNoSubstitutionTemplateLiteral:
						propertyName = elemAccess.ArgumentExpression.AsNoSubstitutionTemplateLiteral().Text
					}
				}

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

	optsMap := utils.GetOptionsMap(opts)
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
