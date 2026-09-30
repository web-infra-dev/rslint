package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsLogicalExpression matches Unicorn's boolean.js helper: nullish coalescing
// does not coerce its operands to booleans.
func IsLogicalExpression(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	op := node.AsBinaryExpression().OperatorToken.Kind
	return op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken
}

func IsLogicalNot(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindPrefixUnaryExpression &&
		node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
}

// IsGlobalBooleanCall recognizes a single, non-spread argument to the global
// Boolean function. Authored TypeScript wrappers remain significant.
func IsGlobalBooleanCall(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression || node.AsCallExpression().QuestionDotToken != nil {
		return false
	}
	callee := utils.ESTreeCallCallee(node.Expression())
	args := node.Arguments()
	return callee != nil && ast.IsIdentifier(callee) && callee.Text() == "Boolean" &&
		len(args) == 1 && args[0].Kind != ast.KindSpreadElement && IsGlobalReference(ctx, callee)
}

func isBooleanCallArgument(ctx rule.RuleContext, node, parent *ast.Node) bool {
	return IsGlobalBooleanCall(ctx, parent) && utils.ESTreeRuntimeExpression(parent.Arguments()[0]) == node
}

// BooleanAncestor folds only explicit ! and Boolean(...) wrappers, returning
// both the reporting node and the parity of its negations.
func BooleanAncestor(ctx rule.RuleContext, node *ast.Node) (*ast.Node, bool) {
	negative := false
	for {
		parent := utils.ESTreeParent(node)
		if IsLogicalNot(parent) {
			negative = !negative
		} else if !isBooleanCallArgument(ctx, node, parent) {
			return node, negative
		}
		node = parent
	}
}

// IsBooleanExpression reports explicit boolean coercion, including through
// logical operands. It does not infer types or treat comparisons as coercions.
func IsBooleanExpression(ctx rule.RuleContext, node *ast.Node) bool {
	for node != nil {
		parent := utils.ESTreeParent(node)
		if IsLogicalNot(node) || IsLogicalNot(parent) || IsGlobalBooleanCall(ctx, node) || isBooleanCallArgument(ctx, node, parent) {
			return true
		}
		if !IsLogicalExpression(parent) {
			return false
		}
		node = parent
	}
	return false
}

// IsControlFlowTest follows && and || operands to the containing test.
func IsControlFlowTest(node *ast.Node) bool {
	for node != nil {
		parent := utils.ESTreeParent(node)
		if parent == nil {
			return false
		}
		var test *ast.Node
		switch parent.Kind {
		case ast.KindIfStatement, ast.KindWhileStatement, ast.KindDoStatement:
			test = parent.Expression()
		case ast.KindConditionalExpression:
			test = parent.AsConditionalExpression().Condition
		case ast.KindForStatement:
			test = parent.AsForStatement().Condition
		}
		if utils.ESTreeRuntimeExpression(test) == node {
			return true
		}
		if !IsLogicalExpression(parent) {
			return false
		}
		node = parent
	}
	return false
}
