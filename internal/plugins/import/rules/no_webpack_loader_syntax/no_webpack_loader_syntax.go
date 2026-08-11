package no_webpack_loader_syntax

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func hasWebpackLoaderSyntax(modulePath string) bool {
	return strings.Contains(modulePath, "!")
}

func buildRuleMessage(modulePath string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "import/no-webpack-loader-syntax",
		// https://github.com/import-js/eslint-plugin-import/blob/01c9eb04331d2efa8d63f2d7f4bfec3bc44c94f3/src/rules/no-webpack-loader-syntax.js#L6C27-L6C110
		Description: fmt.Sprintf("Unexpected '!' in '%s'. Do not use import syntax to configure webpack loaders.", modulePath),
	}
}

// See: https://github.com/import-js/eslint-plugin-import/blob/01c9eb04331d2efa8d63f2d7f4bfec3bc44c94f3/src/rules/no-webpack-loader-syntax.js
var NoWebpackLoaderSyntax = rule.Rule{
	Name:   "import/no-webpack-loader-syntax",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				specifier := node.ModuleSpecifier()
				if specifier == nil || specifier.Kind != ast.KindStringLiteral {
					return
				}
				modulePath := specifier.AsStringLiteral().Text

				if hasWebpackLoaderSyntax(modulePath) {
					ctx.ReportNode(specifier, buildRuleMessage(modulePath))
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				callExpression := node.AsCallExpression()
				// This rule cannot use importutil.GetRequireCall: upstream accepts
				// require("path", extra) as long as the first argument is static,
				// whereas the shared helper deliberately requires exactly one.
				expr := ast.SkipParentheses(callExpression.Expression)
				if expr == nil || expr.Kind != ast.KindIdentifier || expr.AsIdentifier().Text != "require" {
					return
				}
				// ensure there is at least one argument
				if len(callExpression.Arguments.Nodes) == 0 {
					return
				}
				arg := ast.SkipParentheses(callExpression.Arguments.Nodes[0])
				if arg == nil || arg.Kind != ast.KindStringLiteral {
					return
				}
				modulePath := arg.AsStringLiteral().Text
				if hasWebpackLoaderSyntax(modulePath) {
					// report at the string literal argument location for accuracy
					ctx.ReportNode(arg, buildRuleMessage(modulePath))
				}
			},
		}
	},
}
