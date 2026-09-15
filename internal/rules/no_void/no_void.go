package no_void

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_void.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/no-void
var NoVoidRule = rule.Rule{
	Name:   "no-void",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindVoidExpression: func(node *ast.Node) {
				if opts.allowAsStatement {
					// ESTree has no ParenthesizedExpression node, so a
					// void expression wrapped only in parentheses still has
					// the surrounding ExpressionStatement as its parent.
					parent := ast.WalkUpParenthesizedExpressions(node.Parent)
					if parent != nil && parent.Kind == ast.KindExpressionStatement {
						return
					}
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noVoid",
					Description: "Expected 'undefined' and instead saw 'void'.",
				})
			},
		}
	},
}

type noVoidOptions struct {
	allowAsStatement bool
}

func parseOptions(options []any) noVoidOptions {
	opts := noVoidOptions{allowAsStatement: false}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if value, ok := optsMap["allowAsStatement"].(bool); ok {
		opts.allowAsStatement = value
	}
	return opts
}
