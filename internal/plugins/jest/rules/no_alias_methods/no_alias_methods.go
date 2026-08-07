package no_alias_methods

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// Message builder

func buildErrorAliasMethodMessage(alias string, canonical string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceAlias",
		Description: fmt.Sprintf("Replace %s() with its canonical name of %s()", alias, canonical),
	}
}

var methodNames = map[string]string{
	"toBeCalled":       "toHaveBeenCalled",
	"toBeCalledTimes":  "toHaveBeenCalledTimes",
	"toBeCalledWith":   "toHaveBeenCalledWith",
	"lastCalledWith":   "toHaveBeenLastCalledWith",
	"nthCalledWith":    "toHaveBeenNthCalledWith",
	"toReturn":         "toHaveReturned",
	"toReturnTimes":    "toHaveReturnedTimes",
	"toReturnWith":     "toHaveReturnedWith",
	"lastReturnedWith": "toHaveLastReturnedWith",
	"nthReturnedWith":  "toHaveNthReturnedWith",
	"toThrowError":     "toThrow",
}

var NoAliasMethodsRule = rule.Rule{
	Name:   "jest/no-alias-methods",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil {
					return
				}

				if jestFnCall.Kind != utils.JestFnTypeExpect {
					return
				}

				if isNestedCallExpressionInMemberChain(node) {
					return
				}

				if len(jestFnCall.MemberEntries) == 0 {
					return
				}

				for _, memberEntry := range jestFnCall.MemberEntries {
					canonicalName, ok := methodNames[memberEntry.Name]
					if !ok {
						continue
					}

					message := buildErrorAliasMethodMessage(memberEntry.Name, canonicalName)

					// A computed identifier key such as `expect(a)[toBeCalled]()`
					// names a variable, not the matcher, so it still gets
					// reported but must not be rewritten.
					fixRange, fixText, ok := testFramework.AccessorReplacement(ctx.SourceFile, memberEntry.Node, canonicalName)
					if !ok {
						ctx.ReportNode(memberEntry.Node, message)
						continue
					}

					ctx.ReportNodeWithFixes(
						memberEntry.Node, message,
						rule.RuleFix{
							Text:  fixText,
							Range: fixRange,
						},
					)
				}
			},
		}
	},
}

func isNestedCallExpressionInMemberChain(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	switch node.Parent.Kind {
	case ast.KindPropertyAccessExpression:
		return node.Parent.AsPropertyAccessExpression().Expression == node
	case ast.KindElementAccessExpression:
		return node.Parent.AsElementAccessExpression().Expression == node
	case ast.KindCallExpression:
		return node.Parent.AsCallExpression().Expression == node
	case ast.KindTaggedTemplateExpression:
		return node.Parent.AsTaggedTemplateExpression().Tag == node
	default:
		return false
	}
}
