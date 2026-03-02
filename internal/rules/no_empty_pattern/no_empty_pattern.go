package no_empty_pattern

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-empty-pattern
var NoEmptyPatternRule = rule.Rule{
	Name: "no-empty-pattern",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindObjectBindingPattern: func(node *ast.Node) {
				bp := node.AsBindingPattern()
				if bp == nil {
					return
				}

				// Check if it has any elements
				if bp.Elements != nil && len(bp.Elements.Nodes) > 0 {
					return
				}

				// Allow empty object patterns as parameters if option is set
				if opts.allowObjectPatternsAsParameters && isParameterWithEmptyDefault(node) {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpected",
					Description: "Unexpected empty object pattern.",
				})
			},

			ast.KindArrayBindingPattern: func(node *ast.Node) {
				bp := node.AsBindingPattern()
				if bp == nil {
					return
				}

				// Check if it has any elements
				if bp.Elements != nil && len(bp.Elements.Nodes) > 0 {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpected",
					Description: "Unexpected empty array pattern.",
				})
			},
		}
	},
}

// isParameterWithEmptyDefault checks if the binding pattern is a function parameter,
// either directly (function foo({}) {}) or with an empty object default (function foo({} = {}) {}).
// Matches ESLint's behavior: only allows when the default value is specifically {}.
func isParameterWithEmptyDefault(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Direct parameter without default: function foo({}) {}
	if parent.Kind == ast.KindParameter {
		param := parent.AsParameterDeclaration()
		// No initializer — direct empty pattern as parameter
		if param.Initializer == nil {
			return true
		}
		// Has initializer — only allow if it's an empty object literal {}
		return isEmptyObjectLiteral(param.Initializer)
	}

	return false
}

// isEmptyObjectLiteral checks if a node is an empty object literal expression {}.
func isEmptyObjectLiteral(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	objLit := node.AsObjectLiteralExpression()
	return objLit.Properties == nil || len(objLit.Properties.Nodes) == 0
}

type emptyPatternOptions struct {
	allowObjectPatternsAsParameters bool
}

func parseOptions(opts any) emptyPatternOptions {
	result := emptyPatternOptions{
		allowObjectPatternsAsParameters: false,
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
		if allow, ok := optsMap["allowObjectPatternsAsParameters"].(bool); ok {
			result.allowObjectPatternsAsParameters = allow
		}
	}

	return result
}
