package no_restricted_jest_methods

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_restricted_jest_methods.schema.json
var schemaJSON []byte

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

func parseOptions(options []any) map[string]restrictedMethod {
	if len(options) == 0 {
		return nil
	}

	raw, ok := options[0].(map[string]interface{})
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

func isNestedJestFnCall(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}

	if parent == nil {
		return false
	}

	return parent.Kind == ast.KindCallExpression ||
		parent.Kind == ast.KindPropertyAccessExpression ||
		parent.Kind == ast.KindElementAccessExpression
}

var NoRestrictedJestMethodsRule = rule.Rule{
	Name:   "jest/no-restricted-jest-methods",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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

				reportRange, ok := jestUtils.JestFnMemberEntriesRange(jestFnCall.MemberEntries)
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
