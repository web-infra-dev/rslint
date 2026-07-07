package no_conditional_expect

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

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

	name := jestUtils.CalleeChainName(node.AsCallExpression().Expression)
	return name == "catch" || strings.HasSuffix(name, ".catch")
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

func collectTestFunctionCallbacks(ctx rule.RuleContext) map[*ast.Node]bool {
	callbacks := map[*ast.Node]bool{}
	pendingNames := map[string]bool{}

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}

		if node.Kind == ast.KindCallExpression {
			jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
			if jestFnCall != nil && jestFnCall.Kind == jestUtils.JestFnTypeTest {
				info := jestUtils.ResolveTestCallbackFunction(ctx, node.AsCallExpression())
				if info.FunctionNode != nil {
					callbacks[info.FunctionNode] = true
				} else if info.Name != "" {
					pendingNames[info.Name] = true
				}
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}

	if ctx.SourceFile != nil {
		visit(ctx.SourceFile.Node.AsNode())
	}

	if len(pendingNames) > 0 {
		resolvePendingTestCallbackNames(ctx, pendingNames, callbacks)
	}

	return callbacks
}

func resolvePendingTestCallbackNames(
	ctx rule.RuleContext,
	names map[string]bool,
	callbacks map[*ast.Node]bool,
) {
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}

		switch node.Kind {
		case ast.KindFunctionDeclaration:
			fn := node.AsFunctionDeclaration()
			if fn != nil && fn.Name() != nil && names[fn.Name().Text()] {
				callbacks[node] = true
				delete(names, fn.Name().Text())
			}
		case ast.KindVariableDeclaration:
			vd := node.AsVariableDeclaration()
			if vd == nil {
				break
			}
			id := vd.Name()
			if id == nil || id.Kind != ast.KindIdentifier {
				break
			}
			name := id.AsIdentifier().Text
			if !names[name] {
				break
			}
			init := ast.SkipParentheses(vd.Initializer)
			if ast.IsFunctionExpressionOrArrowFunction(init) {
				callbacks[init] = true
				delete(names, name)
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}

	if ctx.SourceFile != nil {
		visit(ctx.SourceFile.Node.AsNode())
	}
}

type callExpressionFrame struct {
	jestFnCall *jestUtils.ParsedJestFnCall
	isCatch    bool
}

var NoConditionalExpectRule = rule.Rule{
	Name: "jest/no-conditional-expect",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.UnwrapOptions(_options)
		_ = options

		testCallbackFunctions := collectTestFunctionCallbacks(ctx)
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
			if !inTestCase() {
				return
			}
			if node.Kind == ast.KindBinaryExpression && !ast.IsLogicalExpression(node) {
				return
			}
			if isConditionalNode(node) {
				conditionalDepth++
			}
		}

		exitConditional := func(node *ast.Node) {
			if !inTestCase() || conditionalDepth == 0 {
				return
			}
			if node.Kind == ast.KindBinaryExpression && !ast.IsLogicalExpression(node) {
				return
			}
			if isConditionalNode(node) {
				conditionalDepth--
			}
		}

		enterTestCallbackFunction := func(node *ast.Node) {
			if testCallbackFunctions[node] {
				enterTestCase()
			}
		}

		exitTestCallbackFunction := func(node *ast.Node) {
			if testCallbackFunctions[node] {
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
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				isCatch := isPromiseCatchCall(node)
				callExpressionFrames[node] = callExpressionFrame{
					jestFnCall: jestFnCall,
					isCatch:    isCatch,
				}

				if jestFnCall != nil && jestFnCall.Kind == jestUtils.JestFnTypeTest {
					enterTestCase()
				}

				if isCatch {
					inPromiseCatch = true
				}

				if jestFnCall == nil || jestFnCall.Kind != jestUtils.JestFnTypeExpect {
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

				if frame.jestFnCall != nil && frame.jestFnCall.Kind == jestUtils.JestFnTypeTest {
					exitTestCase()
				}

				if frame.isCatch {
					inPromiseCatch = false
				}
			},
		}
	},
}
