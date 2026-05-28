package prefer_each

import (
	"github.com/microsoft/typescript-go/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type scopeRange struct {
	start int
	end   int
}

type loopScope struct {
	scopeRange
	jestFnCalls []jestUtils.JestFnType
}

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

func recommendFn(jestFnCalls []jestUtils.JestFnType) string {
	if len(jestFnCalls) == 1 && jestFnCalls[0] == jestUtils.JestFnTypeTest {
		return "it"
	}
	return "describe"
}

func getLoopScopeRange(node *ast.Node) (scopeRange, bool) {
	if node == nil {
		return scopeRange{}, false
	}

	switch node.Kind {
	case ast.KindForStatement:
		stmt := node.AsForStatement().Statement
		if stmt == nil {
			return scopeRange{}, false
		}
		return scopeRange{start: stmt.Pos(), end: stmt.End()}, true
	case ast.KindForInStatement, ast.KindForOfStatement:
		stmt := node.AsForInOrOfStatement().Statement
		if stmt == nil {
			return scopeRange{}, false
		}
		return scopeRange{start: stmt.Pos(), end: stmt.End()}, true
	default:
		return scopeRange{}, false
	}
}

func getFunctionBodyRange(fn *ast.Node) (scopeRange, bool) {
	if fn == nil {
		return scopeRange{}, false
	}

	body := fn.Body()
	if body == nil {
		return scopeRange{}, false
	}

	return scopeRange{start: body.Pos(), end: body.End()}, true
}

func getTestCaseScopeRange(node *ast.Node) (scopeRange, bool) {
	callExpr := node.AsCallExpression()
	if callExpr == nil || callExpr.Arguments == nil {
		return scopeRange{}, false
	}

	for i := len(callExpr.Arguments.Nodes) - 1; i >= 0; i-- {
		if bodyRange, ok := getFunctionBodyRange(callExpr.Arguments.Nodes[i]); ok {
			return bodyRange, true
		}
	}

	return scopeRange{}, false
}

func rangeContainsNode(scope scopeRange, node *ast.Node) bool {
	if node == nil {
		return false
	}

	pos := node.Pos()
	return scope.start <= pos && pos < scope.end
}

var PreferEachRule = rule.Rule{
	Name: "jest/prefer-each",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		loops := make([]loopScope, 0, 4)
		testCases := make([]scopeRange, 0, 4)

		enterForLoop := func(node *ast.Node) {
			if bodyRange, ok := getLoopScopeRange(node); ok {
				loops = append(loops, loopScope{scopeRange: bodyRange})
			}
		}

		exitForLoop := func(node *ast.Node) {
			if len(loops) == 0 {
				return
			}

			currentLoop := loops[len(loops)-1]
			loops = loops[:len(loops)-1]
			if len(currentLoop.jestFnCalls) == 0 {
				return
			}

			ctx.ReportNode(node, buildPreferEachMessage(recommendFn(currentLoop.jestFnCalls)))
		}

		isInsideTestCase := func(node *ast.Node) bool {
			for i := len(testCases) - 1; i >= 0; i-- {
				if rangeContainsNode(testCases[i], node) {
					return true
				}
			}
			return false
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
					if !isInsideTestCase(node) {
						for i := range loops {
							if rangeContainsNode(loops[i].scopeRange, node) {
								loops[i].jestFnCalls = append(loops[i].jestFnCalls, jestFnCall.Kind)
							}
						}
					}
					if jestFnCall.Kind == jestUtils.JestFnTypeTest {
						if bodyRange, ok := getTestCaseScopeRange(node); ok {
							testCases = append(testCases, bodyRange)
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall != nil && jestFnCall.Kind == jestUtils.JestFnTypeTest && len(testCases) > 0 {
					testCases = testCases[:len(testCases)-1]
				}
			},
		}
	},
}
