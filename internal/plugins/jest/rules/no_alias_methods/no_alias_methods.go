package no_alias_methods

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
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
	Name: "jest/no-alias-methods",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil {
					return
				}

				if jestFnCall.Kind != utils.JestFnTypeExpect {
					return
				}

				for _, memberEntry := range jestFnCall.MemberEntries {
					if canonicalName, ok := methodNames[memberEntry.Name]; ok {
						start := memberEntry.Node.Pos()
						end := memberEntry.Node.End()

						if memberEntry.Node.Kind != ast.KindIdentifier {
							start = start + 1
							end = end - 1
						}

						ctx.ReportNodeWithFixes(
							memberEntry.Node, buildErrorAliasMethodMessage(memberEntry.Name, canonicalName),
							rule.RuleFix{
								Text:  canonicalName,
								Range: core.NewTextRange(start, end),
							},
						)
						break
					}
				}
			},
		}
	},
}
