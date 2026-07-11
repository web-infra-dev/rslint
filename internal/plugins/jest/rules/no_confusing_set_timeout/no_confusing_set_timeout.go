package no_confusing_set_timeout

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildGlobalSetTimeoutMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "globalSetTimeout",
		Description: "`jest.setTimeout` should be call in `global` scope",
	}
}

func buildMultipleSetTimeoutsMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "multipleSetTimeouts",
		Description: "Do not call `jest.setTimeout` multiple times, as only the last call will have an effect",
	}
}

func buildOrderSetTimeoutMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "orderSetTimeout",
		Description: "`jest.setTimeout` should be placed before any other jest methods",
	}
}

func isJestSetTimeout(jestFnCall *utils.ParsedJestFnCall) bool {
	return jestFnCall != nil &&
		jestFnCall.Kind == utils.JestFnTypeJest &&
		len(jestFnCall.Members) == 1 &&
		len(jestFnCall.MemberEntries) == 1 &&
		jestFnCall.MemberEntries[0].Node != nil &&
		jestFnCall.MemberEntries[0].Node.Kind == ast.KindIdentifier &&
		jestFnCall.Members[0] == "setTimeout"
}

func isInGlobalOrModuleScope(node *ast.Node) bool {
	if node == nil {
		return false
	}

	return ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(n) || ast.IsClassLike(n) {
			return true
		}
		switch n.Kind {
		case ast.KindBlock,
			ast.KindCatchClause,
			ast.KindEnumDeclaration,
			ast.KindModuleBlock,
			ast.KindSwitchStatement,
			ast.KindWithStatement:
			return true
		case ast.KindForStatement:
			initializer := n.AsForStatement().Initializer
			return initializer != nil &&
				initializer.Kind == ast.KindVariableDeclarationList &&
				initializer.Flags&ast.NodeFlagsBlockScoped != 0
		case ast.KindForInStatement, ast.KindForOfStatement:
			initializer := n.AsForInOrOfStatement().Initializer
			return initializer != nil &&
				initializer.Kind == ast.KindVariableDeclarationList &&
				initializer.Flags&ast.NodeFlagsBlockScoped != 0
		}
		return false
	}) == nil
}

var NoConfusingSetTimeoutRule = rule.Rule{
	Name: "jest/no-confusing-set-timeout",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		seenJestTimeout := false
		shouldEmitOrderSetTimeout := false

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil {
					return
				}

				if !isJestSetTimeout(jestFnCall) {
					shouldEmitOrderSetTimeout = true
					return
				}

				if !isInGlobalOrModuleScope(node) {
					ctx.ReportNode(node, buildGlobalSetTimeoutMessage())
				}
				if shouldEmitOrderSetTimeout {
					ctx.ReportNode(node, buildOrderSetTimeoutMessage())
				}
				if seenJestTimeout {
					ctx.ReportNode(node, buildMultipleSetTimeoutsMessage())
				}

				seenJestTimeout = true
			},
		}
	},
}
