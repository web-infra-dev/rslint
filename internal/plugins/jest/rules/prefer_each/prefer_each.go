package prefer_each

import (
	"github.com/microsoft/typescript-go/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builder

func buildPreferEachMessage(fn string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferEach",
		Description: "prefer using `" + fn + ".each` rather than a manual loop",
		Data: map[string]string{
			"fn": fn,
		},
	}
}

var PreferEachRule = rule.Rule{
	Name: "jest/prefer-each",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var jestFnCalls []jestUtils.JestFnType
		inTestCaseCall := false

		recommendFn := func() string {
			if len(jestFnCalls) == 1 && jestFnCalls[0] == jestUtils.JestFnTypeTest {
				return "it"
			}
			return "describe"
		}

		enterForLoop := func(node *ast.Node) {
			if len(jestFnCalls) == 0 || inTestCaseCall {
				return
			}
			jestFnCalls = jestFnCalls[:0]
		}

		exitForLoop := func(node *ast.Node) {
			if len(jestFnCalls) == 0 || inTestCaseCall {
				return
			}
			ctx.ReportNode(node, buildPreferEachMessage(recommendFn()))
			jestFnCalls = jestFnCalls[:0]
		}

		return rule.RuleListeners{
			ast.KindForStatement:                        enterForLoop,
			ast.KindForInStatement:                      enterForLoop,
			ast.KindForOfStatement:                      enterForLoop,
			rule.ListenerOnExit(ast.KindForStatement):   exitForLoop,
			rule.ListenerOnExit(ast.KindForInStatement): exitForLoop,
			rule.ListenerOnExit(ast.KindForOfStatement): exitForLoop,
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil {
					return
				}
				switch jestFnCall.Kind {
				case jestUtils.JestFnTypeHook,
					jestUtils.JestFnTypeDescribe,
					jestUtils.JestFnTypeTest:
					jestFnCalls = append(jestFnCalls, jestFnCall.Kind)
				}
				if jestFnCall.Kind == jestUtils.JestFnTypeTest {
					inTestCaseCall = true
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall != nil && jestFnCall.Kind == jestUtils.JestFnTypeTest {
					inTestCaseCall = false
				}
			},
		}
	},
}
