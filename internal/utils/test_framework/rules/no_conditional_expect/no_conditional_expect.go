package no_conditional_expect

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type Runtime struct {
	TestCallbackFunctions map[*ast.Node]bool
	ClassifyCall          func(*ast.Node) (isTest bool, isExpect bool)
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
}

func buildConditionalExpectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionalExpect",
		Description: "Avoid calling `expect` conditionally",
	}
}

func isPromiseCatchCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	name := calleeChainName(node.AsCallExpression().Expression)
	return name == "catch" || strings.HasSuffix(name, ".catch")
}

func calleeChainName(node *ast.Node) string {
	entries := testFramework.GetMemberEntries(node)
	return testFramework.JoinMemberEntries(entries)
}

func isConditionalNode(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindCatchClause,
		ast.KindIfStatement,
		ast.KindSwitchStatement,
		ast.KindConditionalExpression:
		return true
	case ast.KindBinaryExpression:
		return ast.IsLogicalExpression(node)
	default:
		return false
	}
}

type callExpressionFrame struct {
	isTest  bool
	isCatch bool
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			testCaseDepth := 0
			conditionalDepth := 0
			inPromiseCatch := false
			callExpressionFrames := map[*ast.Node]callExpressionFrame{}

			inTestCase := func() bool {
				return testCaseDepth > 0
			}

			enterTestCase := func() {
				testCaseDepth++
			}

			exitTestCase := func() {
				if testCaseDepth > 0 {
					testCaseDepth--
				}
			}

			enterConditional := func(node *ast.Node) {
				if inTestCase() && isConditionalNode(node) {
					conditionalDepth++
				}
			}

			exitConditional := func(node *ast.Node) {
				if inTestCase() && conditionalDepth > 0 && isConditionalNode(node) {
					conditionalDepth--
				}
			}

			enterTestCallbackFunction := func(node *ast.Node) {
				if runtime.TestCallbackFunctions[node] {
					enterTestCase()
				}
			}

			exitTestCallbackFunction := func(node *ast.Node) {
				if runtime.TestCallbackFunctions[node] {
					exitTestCase()
				}
			}

			return rule.RuleListeners{
				ast.KindFunctionDeclaration:                      enterTestCallbackFunction,
				rule.ListenerOnExit(ast.KindFunctionDeclaration): exitTestCallbackFunction,
				ast.KindFunctionExpression:                       enterTestCallbackFunction,
				rule.ListenerOnExit(ast.KindFunctionExpression):  exitTestCallbackFunction,
				ast.KindArrowFunction:                            enterTestCallbackFunction,
				rule.ListenerOnExit(ast.KindArrowFunction):       exitTestCallbackFunction,

				ast.KindCatchClause:                                enterConditional,
				rule.ListenerOnExit(ast.KindCatchClause):           exitConditional,
				ast.KindIfStatement:                                enterConditional,
				rule.ListenerOnExit(ast.KindIfStatement):           exitConditional,
				ast.KindSwitchStatement:                            enterConditional,
				rule.ListenerOnExit(ast.KindSwitchStatement):       exitConditional,
				ast.KindConditionalExpression:                      enterConditional,
				rule.ListenerOnExit(ast.KindConditionalExpression): exitConditional,
				ast.KindBinaryExpression:                           enterConditional,
				rule.ListenerOnExit(ast.KindBinaryExpression):      exitConditional,

				ast.KindCallExpression: func(node *ast.Node) {
					isTest := false
					isExpect := false
					if runtime.ClassifyCall != nil {
						isTest, isExpect = runtime.ClassifyCall(node)
					}
					isCatch := isPromiseCatchCall(node)
					callExpressionFrames[node] = callExpressionFrame{
						isTest:  isTest,
						isCatch: isCatch,
					}

					if isTest {
						enterTestCase()
					}
					if isCatch {
						inPromiseCatch = true
					}
					if !isExpect {
						return
					}
					if inTestCase() && conditionalDepth > 0 {
						ctx.ReportNode(node, buildConditionalExpectMessage())
					}
					if inPromiseCatch {
						ctx.ReportNode(node, buildConditionalExpectMessage())
					}
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					frame, ok := callExpressionFrames[node]
					if ok {
						delete(callExpressionFrames, node)
					}
					if frame.isTest {
						exitTestCase()
					}
					if frame.isCatch {
						inPromiseCatch = false
					}
				},
			}
		},
	}
}
