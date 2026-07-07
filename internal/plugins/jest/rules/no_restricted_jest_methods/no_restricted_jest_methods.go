package no_restricted_jest_methods

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type restrictedMethod struct {
	Message string
	HasMsg  bool
}

func buildRestrictedJestMethodMessage(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "restrictedJestMethod",
		Description: "Use of `" + method + "` is disallowed",
		Data: map[string]string{
			"restriction": method,
		},
	}
}

func buildRestrictedJestMethodWithMessage(method string, message string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "restrictedJestMethodWithMessage",
		Description: message,
		Data: map[string]string{
			"message":     message,
			"restriction": method,
		},
	}
}

func parseOptions(options any) map[string]restrictedMethod {
	normalized := rule.NormalizeOptions(options)
	if len(normalized) == 0 {
		return nil
	}

	raw, ok := normalized[0].(map[string]interface{})
	if !ok {
		return nil
	}

	restricted := make(map[string]restrictedMethod, len(raw))
	for method, rawMessage := range raw {
		if rawMessage == nil {
			restricted[method] = restrictedMethod{}
			continue
		}

		message, ok := rawMessage.(string)
		if !ok {
			continue
		}
		if message == "" {
			restricted[method] = restrictedMethod{}
			continue
		}

		restricted[method] = restrictedMethod{
			Message: message,
			HasMsg:  true,
		}
	}

	return restricted
}

func restrictedRange(entries []jestUtils.ParsedJestFnMemberEntry) (core.TextRange, bool) {
	if len(entries) == 0 || entries[0].Node == nil || entries[len(entries)-1].Node == nil {
		return core.TextRange{}, false
	}

	return core.NewTextRange(entries[0].Node.Pos(), entries[len(entries)-1].Node.End()), true
}

func isNestedJestFnCall(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	return node.Parent.Kind == ast.KindCallExpression ||
		node.Parent.Kind == ast.KindPropertyAccessExpression ||
		node.Parent.Kind == ast.KindElementAccessExpression
}

var NoRestrictedJestMethodsRule = rule.Rule{
	Name: "jest/no-restricted-jest-methods",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		restrictedMethods := parseOptions(options)
		if len(restrictedMethods) == 0 {
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if isNestedJestFnCall(node) {
					return
				}

				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil || jestFnCall.Kind != jestUtils.JestFnTypeJest || len(jestFnCall.Members) == 0 {
					return
				}

				method := jestFnCall.Members[0]
				restriction, ok := restrictedMethods[method]
				if !ok {
					return
				}

				reportRange, ok := restrictedRange(jestFnCall.MemberEntries)
				if !ok {
					return
				}

				if restriction.HasMsg {
					ctx.ReportRange(reportRange, buildRestrictedJestMethodWithMessage(method, restriction.Message))
				} else {
					ctx.ReportRange(reportRange, buildRestrictedJestMethodMessage(method))
				}
			},
		}
	},
}
