package no_process_env

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_process_env.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-process-env.js
var NoProcessEnvRule = rule.Rule{
	Name:   "node/no-process-env",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var allowedVariables []any
		if len(options) > 0 {
			if opts, ok := options[0].(map[string]any); ok {
				allowedVariables, _ = opts["allowedVariables"].([]any)
			}
		}

		check := func(node *ast.Node) {
			if utils.IsInJsxTagName(node) {
				return
			}
			object, property := utils.MemberExpressionParts(node)
			object = utils.ESTreeRuntimeExpression(object)
			property = utils.ESTreeRuntimeExpression(property)
			if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "process" || property == nil {
				return
			}
			if node.Kind == ast.KindElementAccessExpression {
				if property.Kind != ast.KindStringLiteral || property.Text() != "env" {
					return
				}
			} else if (property.Kind != ast.KindIdentifier || property.Text() != "env") &&
				(property.Kind != ast.KindPrivateIdentifier || property.Text() != "#env") {
				return
			}

			// At the end of an optional chain, ESTree inserts a ChainExpression
			// parent, so a following member cannot grant an allowedVariables exemption.
			if len(allowedVariables) > 0 {
				parent := utils.ESTreeParent(node)
				if ast.IsOptionalChain(node) && (!ast.IsOptionalChain(node.Parent) || node.Parent.Expression() != node) {
					parent = nil
				}
				_, child := utils.MemberExpressionParts(parent)
				child = utils.ESTreeRuntimeExpression(child)
				if child != nil && ((parent.Kind != ast.KindElementAccessExpression && child.Kind == ast.KindIdentifier) ||
					(parent.Kind == ast.KindElementAccessExpression && child.Kind == ast.KindStringLiteral)) &&
					slices.Contains(allowedVariables, any(child.Text())) {
					return
				}
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpectedProcessEnv",
				Description: "Unexpected use of process.env.",
			})
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: check,
			ast.KindElementAccessExpression:  check,
			ast.KindQualifiedName:            check,
		}
	},
}
