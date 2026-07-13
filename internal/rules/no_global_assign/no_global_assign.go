package no_global_assign

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type options struct {
	exceptions map[string]bool
}

func parseOptions(opts any) options {
	result := options{exceptions: make(map[string]bool)}
	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if exceptions, ok := optsMap["exceptions"].([]interface{}); ok {
			for _, e := range exceptions {
				if s, ok := e.(string); ok {
					result.exceptions[s] = true
				}
			}
		}
	}
	return result
}

// isWriteThroughTypeAssertion checks if the identifier reaches its assignment target
// through an AsExpression or TypeAssertionExpression. ESLint's scope analysis does not
// track writes through these TS-specific wrappers, so we skip them to match ESLint.
func isWriteThroughTypeAssertion(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression:
			return true
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
			current = current.Parent
			continue
		default:
			return false
		}
	}
	return false
}

// NoGlobalAssignRule disallows assignments to native objects or read-only global variables
var NoGlobalAssignRule = rule.Rule{
	Name: "no-global-assign",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				name := node.Text()
				if !utils.IsECMAScriptGlobal(name) || opts.exceptions[name] {
					return
				}

				if declared, ok := ctx.Globals[name]; ok && !declared {
					return
				}

				if !utils.IsWriteReference(node) {
					return
				}

				if isWriteThroughTypeAssertion(node) {
					return
				}

				if utils.IsShadowed(node, name) {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "globalShouldNotBeModified",
					Description: fmt.Sprintf("Read-only global '%s' should not be modified.", name),
				})
			},
		}
	},
}
