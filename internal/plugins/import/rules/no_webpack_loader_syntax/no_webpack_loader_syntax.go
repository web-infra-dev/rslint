package no_webpack_loader_syntax

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	importutil "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
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
				call := importutil.GetRequireCallWithStringLiteralArgument(node)
				if call == nil {
					return
				}
				arg := ast.SkipParentheses(call.Arguments.Nodes[0])
				modulePath := arg.AsStringLiteral().Text
				if hasWebpackLoaderSyntax(modulePath) {
					// report at the string literal argument location for accuracy
					ctx.ReportNode(arg, buildRuleMessage(modulePath))
				}
			},
		}
	},
}
