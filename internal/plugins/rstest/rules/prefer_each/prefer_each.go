package prefer_each

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func buildPreferEachMessage(fn string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferEach",
		Description: "prefer using `" + fn + ".each` rather than a manual loop",
		Data: map[string]string{
			"fn": fn,
		},
	}
}

type pendingRegistration struct {
	kind rstestUtils.RstestFnType
	name string
}

func recommendFn(pending []pendingRegistration) string {
	if len(pending) == 1 && pending[0].kind == rstestUtils.RstestFnTypeTest {
		if pending[0].name == "it" {
			return "it"
		}
		return "test"
	}
	return "describe"
}

var PreferEachRule = rule.Rule{
	Name:   "rstest/prefer-each",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		callbacks := analysis.Callbacks().Functions
		// NOTE: Unlike eslint-plugin-jest's single boolean, rstest tracks the
		// callback depth from analysis.Callbacks().Functions so nested test
		// registrations inside a test callback do not accidentally re-enable
		// reporting for later business loops in the outer callback.
		pending := make([]pendingRegistration, 0, 4)
		testCallbackDepth := 0

		enterLoop := func(node *ast.Node) {
			if testCallbackDepth > 0 || len(pending) == 0 {
				return
			}
			pending = pending[:0]
		}

		exitLoop := func(node *ast.Node) {
			if testCallbackDepth > 0 || len(pending) == 0 {
				return
			}
			ctx.ReportNode(node, buildPreferEachMessage(recommendFn(pending)))
			pending = pending[:0]
		}

		enterFunction := func(node *ast.Node) {
			if callbacks[node] {
				testCallbackDepth++
			}
		}

		exitFunction := func(node *ast.Node) {
			if callbacks[node] {
				testCallbackDepth--
			}
		}

		return rule.RuleListeners{
			ast.KindForStatement:                             enterLoop,
			ast.KindForInStatement:                           enterLoop,
			ast.KindForOfStatement:                           enterLoop,
			rule.ListenerOnExit(ast.KindForStatement):        exitLoop,
			rule.ListenerOnExit(ast.KindForInStatement):      exitLoop,
			rule.ListenerOnExit(ast.KindForOfStatement):      exitLoop,
			ast.KindFunctionDeclaration:                      enterFunction,
			ast.KindFunctionExpression:                       enterFunction,
			ast.KindArrowFunction:                            enterFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			rule.ListenerOnExit(ast.KindArrowFunction):       exitFunction,
			ast.KindCallExpression: func(node *ast.Node) {
				if testCallbackDepth > 0 {
					return
				}
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return
				}
				switch parsed.Kind {
				case testFramework.FnKindTest,
					testFramework.FnKindDescribe,
					testFramework.FnKindHook:
					pending = append(pending, pendingRegistration{
						kind: parsed.Kind,
						name: parsed.Name,
					})
				}
			},
		}
	},
}
