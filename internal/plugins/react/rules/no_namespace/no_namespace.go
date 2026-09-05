package no_namespace

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const msgNoNamespace = "React component {{name}} must not be in a namespace, as React does not support them"

var NoNamespaceRule = rule.Rule{
	Name:   "react/no-namespace",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		pragma := reactutil.GetReactPragmaFromContext(ctx)

		reportIfNamespaced := func(node *ast.Node, name string) {
			if name == "" || strings.IndexByte(name, ':') == -1 {
				return
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noNamespace",
				Description: strings.Replace(msgNoNamespace, "{{name}}", name, 1),
			})
		}

		checkJSX := func(node *ast.Node) {
			// ESTree exposes both JSX element forms through JSXOpeningElement;
			// tsgo separates paired and self-closing JSX, so listen to both.
			reportIfNamespaced(node, reactutil.GetJsxElementTypeString(node))
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     checkJSX,
			ast.KindJsxSelfClosingElement: checkJSX,
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !reactutil.IsCreateElementCallWithRefs(call.Expression, pragma, ctx.TypeChecker, ctx.Refs) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}

				// Upstream checks Literal, not TemplateLiteral or a TypeScript
				// wrapper. Parentheses are skipped because ESTree flattens them.
				argument := utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0])
				if argument == nil || argument.Kind != ast.KindStringLiteral {
					return
				}
				reportIfNamespaced(node, argument.AsStringLiteral().Text)
			},
		}
	},
}
