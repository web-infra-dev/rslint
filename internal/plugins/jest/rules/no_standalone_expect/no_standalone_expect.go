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

	optArray, ok := raw.([]interface{})
	if !ok || len(optArray) == 0 {
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

	if ast.FindAncestorKind(fn, ast.KindVariableDeclaration) != nil {
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

var NoStandaloneExpectRule = rule.Rule{
	Name: "jest/no-standalone-expect",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		callStack := make([]blockType, 0, 8)

		isCustomTestBlockFunction := func(node *ast.Node) bool {
			name := utils.GetNodeName(node)
			if name == "" {
				return false
			}
			_, ok := opts.additionalTestBlockFunctions[name]
			return ok
		}

		currentBlock := func() blockType {
			if len(callStack) == 0 {
				return ""
			}
			return callStack[len(callStack)-1]
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall != nil && jestFnCall.Kind == utils.JestFnTypeExpect {
					if !shouldReportExpectCall(jestFnCall) {
						return
					}

					parent := currentBlock()
					if parent == "" || parent == blockTypeDescribe {
						ctx.ReportNode(node, buildUnexpectedExpectMessage())
					}
					return
				}

				if (jestFnCall != nil && jestFnCall.Kind == utils.JestFnTypeTest) || isCustomTestBlockFunction(node) {
					callStack = append(callStack, blockTypeTest)
				}

				if calleeIsTaggedTemplateExpression(node) {
					callStack = append(callStack, blockTypeTemplate)
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				top := currentBlock()
				if top == blockTypeTest &&
					(utils.IsTypeOfJestFnCall(node, ctx, utils.JestFnTypeTest) || isCustomTestBlockFunction(node)) {
					callStack = callStack[:len(callStack)-1]
					return
				}

				if top == blockTypeTemplate && calleeIsTaggedTemplateExpression(node) {
					callStack = callStack[:len(callStack)-1]
				}
			},
			ast.KindBlock: func(node *ast.Node) {
				if blockType := getBlockType(node, ctx); blockType != "" {
					callStack = append(callStack, blockType)
				}
			},
			rule.ListenerOnExit(ast.KindBlock): func(node *ast.Node) {
				blockType := getBlockType(node, ctx)
				if blockType != "" && currentBlock() == blockType {
					callStack = callStack[:len(callStack)-1]
				}
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				if node.Parent != nil && node.Parent.Kind != ast.KindCallExpression {
					callStack = append(callStack, blockTypeArrow)
				}
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				if currentBlock() == blockTypeArrow {
					callStack = callStack[:len(callStack)-1]
				}
			},
		}
	},
}
