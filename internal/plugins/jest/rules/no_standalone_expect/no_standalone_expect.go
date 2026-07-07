package no_standalone_expect

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type blockType string

const (
	blockTypeTest     blockType = "test"
	blockTypeFunction blockType = "function"
	blockTypeDescribe blockType = "describe"
	blockTypeArrow    blockType = "arrow"
	blockTypeTemplate blockType = "template"
)

type options struct {
	additionalTestBlockFunctions map[string]struct{}
}

type scopeEntry struct {
	kind  blockType
	start int
	end   int
}

// Message Builder

func buildUnexpectedExpectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedExpect",
		Description: "Expect must be inside of a test block",
	}
}

func parseOptions(raw any) options {
	opts := options{additionalTestBlockFunctions: map[string]struct{}{}}
	if raw == nil {
		return opts
	}

	optArray := rule.NormalizeOptions(raw)
	if len(optArray) == 0 {
		return opts
	}

	optsMap, ok := optArray[0].(map[string]interface{})
	if !ok {
		return opts
	}

	rawFns, ok := optsMap["additionalTestBlockFunctions"]
	if !ok || rawFns == nil {
		return opts
	}

	list, ok := rawFns.([]interface{})
	if !ok {
		return opts
	}

	for _, item := range list {
		if name, ok := item.(string); ok {
			opts.additionalTestBlockFunctions[name] = struct{}{}
		}
	}

	return opts
}

func getBlockType(block *ast.Node, ctx rule.RuleContext) blockType {
	fn := block.Parent
	if !utils.IsFunction(fn) {
		return ""
	}

	if ast.IsFunctionDeclaration(fn) {
		return blockTypeFunction
	}

	if fn.Parent != nil && fn.Parent.Kind == ast.KindVariableDeclaration {
		return blockTypeFunction
	}

	switch fn.Kind {
	case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		return blockTypeFunction
	}

	parent := fn.Parent
	if parent != nil && parent.Kind == ast.KindCallExpression &&
		utils.IsTypeOfJestFnCall(parent, ctx, utils.JestFnTypeDescribe) {
		return blockTypeDescribe
	}

	return ""
}

func isStaticExpectPropertyCall(jestFnCall *utils.ParsedJestFnCall) bool {
	return jestFnCall != nil &&
		jestFnCall.Kind == utils.JestFnTypeExpect &&
		utils.IsStaticExpectMatcher(jestFnCall.Matcher, jestFnCall.Head.Local.Node)
}

func shouldReportExpectCall(jestFnCall *utils.ParsedJestFnCall) bool {
	return !isStaticExpectPropertyCall(jestFnCall)
}

func calleeIsTaggedTemplateExpression(node *ast.Node) bool {
	callExpr := node.AsCallExpression()
	return callExpr != nil &&
		callExpr.Expression != nil &&
		callExpr.Expression.Kind == ast.KindTaggedTemplateExpression
}

func getFunctionBodyRange(fn *ast.Node) (int, int, bool) {
	if fn == nil {
		return 0, 0, false
	}

	body := fn.Body()
	if body == nil {
		return 0, 0, false
	}

	return body.Pos(), body.End(), true
}

func getTestScopeRange(node *ast.Node) (int, int, bool) {
	callExpr := node.AsCallExpression()
	if callExpr == nil || callExpr.Arguments == nil {
		return 0, 0, false
	}

	for i := len(callExpr.Arguments.Nodes) - 1; i >= 0; i-- {
		arg := callExpr.Arguments.Nodes[i]
		if start, end, ok := getFunctionBodyRange(arg); ok {
			return start, end, true
		}
	}

	return 0, 0, false
}

func getTemplateScopeRange(node *ast.Node) (int, int, bool) {
	callExpr := node.AsCallExpression()
	if callExpr == nil || callExpr.Expression == nil || callExpr.Expression.Kind != ast.KindTaggedTemplateExpression {
		return 0, 0, false
	}

	return callExpr.Expression.Pos(), callExpr.Expression.End(), true
}

var NoStandaloneExpectRule = rule.Rule{
	Name: "jest/no-standalone-expect",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.UnwrapOptions(_options)
		opts := parseOptions(options)
		scopeStack := make([]scopeEntry, 0, 8)

		isCustomTestBlockFunction := func(node *ast.Node) bool {
			callExpr := node.AsCallExpression()
			if callExpr == nil {
				return false
			}
			name := utils.CalleeChainName(callExpr.Expression)
			if name == "" {
				return false
			}
			_, ok := opts.additionalTestBlockFunctions[name]
			return ok
		}

		currentScopeKind := func() blockType {
			if len(scopeStack) == 0 {
				return ""
			}
			return scopeStack[len(scopeStack)-1].kind
		}

		pushScope := func(kind blockType, start, end int) {
			scopeStack = append(scopeStack, scopeEntry{kind: kind, start: start, end: end})
		}

		popScope := func(kind blockType) {
			if len(scopeStack) == 0 || scopeStack[len(scopeStack)-1].kind != kind {
				return
			}

			scopeStack = scopeStack[:len(scopeStack)-1]
		}

		getContainingBlock := func(node *ast.Node) blockType {
			pos := node.Pos()
			for i := len(scopeStack) - 1; i >= 0; i-- {
				scope := scopeStack[i]
				if scope.start <= pos && pos < scope.end {
					return scope.kind
				}
			}

			return ""
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall != nil && jestFnCall.Kind == utils.JestFnTypeExpect {
					if !shouldReportExpectCall(jestFnCall) {
						return
					}

					parent := getContainingBlock(node)
					if parent == "" || parent == blockTypeDescribe {
						ctx.ReportNode(node, buildUnexpectedExpectMessage())
					}
					return
				}

				if (jestFnCall != nil && jestFnCall.Kind == utils.JestFnTypeTest) || isCustomTestBlockFunction(node) {
					if start, end, ok := getTestScopeRange(node); ok {
						pushScope(blockTypeTest, start, end)
					}
				}

				if calleeIsTaggedTemplateExpression(node) {
					if start, end, ok := getTemplateScopeRange(node); ok {
						pushScope(blockTypeTemplate, start, end)
					}
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if calleeIsTaggedTemplateExpression(node) {
					popScope(blockTypeTemplate)
				}

				if (utils.IsTypeOfJestFnCall(node, ctx, utils.JestFnTypeTest) || isCustomTestBlockFunction(node)) &&
					currentScopeKind() == blockTypeTest {
					popScope(blockTypeTest)
				}
			},
			ast.KindBlock: func(node *ast.Node) {
				if blockType := getBlockType(node, ctx); blockType != "" {
					pushScope(blockType, node.Pos(), node.End())
				}
			},
			rule.ListenerOnExit(ast.KindBlock): func(node *ast.Node) {
				blockType := getBlockType(node, ctx)
				if blockType != "" && currentScopeKind() == blockType {
					popScope(blockType)
				}
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				if node.Parent != nil && node.Parent.Kind != ast.KindCallExpression {
					if start, end, ok := getFunctionBodyRange(node); ok {
						pushScope(blockTypeArrow, start, end)
					}
				}
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				if currentScopeKind() == blockTypeArrow {
					popScope(blockTypeArrow)
				}
			},
		}
	},
}
