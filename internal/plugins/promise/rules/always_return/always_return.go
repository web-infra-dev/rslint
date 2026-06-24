package always_return

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const skipTransparent = ast.OEKParentheses

type Options struct {
	IgnoreLastCallback       bool
	IgnoreAssignmentVariable []string
}

func buildThenShouldReturnOrThrowMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "thenShouldReturnOrThrow",
		Description: "Each then() should return a value or throw",
	}
}

func isInlineThenFunctionExpression(node *ast.Node) bool {
	if node == nil || !isFunctionWithBlockStatement(node) {
		return false
	}
	parent := node.Parent
	for parent != nil && ast.IsOuterExpression(parent, skipTransparent) {
		parent = parent.Parent
	}
	if parent == nil || !ast.IsCallExpression(parent) || !isMemberCall(parent, "then") {
		return false
	}
	args := parent.Arguments()
	if len(args) == 0 {
		return false
	}
	firstArg := ast.SkipOuterExpressions(args[0], skipTransparent)
	return firstArg == node
}

func isFunctionWithBlockStatement(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionExpression:
		return node.Body() != nil
	case ast.KindArrowFunction:
		body := node.Body()
		return body != nil && body.Kind == ast.KindBlock
	default:
		return false
	}
}

func isMemberCall(node *ast.Node, memberName string) bool {
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}
	callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
	if callee == nil || !ast.IsPropertyAccessExpression(callee) {
		return false
	}
	name := callee.AsPropertyAccessExpression().Name()
	return name != nil && ast.IsIdentifier(name) && name.AsIdentifier().Text == memberName
}

func isLastCallback(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	target := node.Parent
	for target != nil && ast.IsOuterExpression(target, skipTransparent) {
		target = target.Parent
	}
	parent := target.Parent
	for parent != nil {
		if ast.IsOuterExpression(parent, skipTransparent) {
			target = parent
			parent = target.Parent
			continue
		}
		switch parent.Kind {
		case ast.KindExpressionStatement:
			return true
		case ast.KindVoidExpression:
			return true
		case ast.KindAwaitExpression:
			target = parent
			parent = target.Parent
			continue
		case ast.KindBinaryExpression:
			bin := parent.AsBinaryExpression()
			if bin == nil || bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindCommaToken {
				return false
			}
			if ast.SkipOuterExpressions(bin.Right, skipTransparent) != target {
				return true
			}
			target = parent
			parent = target.Parent
			continue
		case ast.KindPropertyAccessExpression:
			prop := parent.AsPropertyAccessExpression()
			if prop == nil || ast.SkipOuterExpressions(prop.Expression, skipTransparent) != target {
				return false
			}
			call := parent.Parent
			if call != nil && ast.IsCallExpression(call) && (isMemberCall(call, "catch") || isMemberCall(call, "finally")) {
				target = call
				parent = target.Parent
				continue
			}
			return false
		default:
			return false
		}
	}
	return false
}

func bodyAlwaysReturnsOrThrows(fn *ast.Node) bool {
	body := fn.Body()
	if body == nil {
		return false
	}
	if body.Kind != ast.KindBlock {
		return true
	}
	return utils.StatementListCompletion(body.Statements()) == utils.CompletionTerminates
}

func hasIgnoredAssignment(body *ast.Node, ignoredVars []string) bool {
	if body == nil || body.Kind != ast.KindBlock || len(ignoredVars) == 0 {
		return false
	}
	for _, stmt := range body.Statements() {
		if isIgnoredAssignment(stmt, ignoredVars) {
			return true
		}
	}
	return false
}

func isIgnoredAssignment(stmt *ast.Node, ignoredVars []string) bool {
	if stmt == nil || stmt.Kind != ast.KindExpressionStatement {
		return false
	}
	expr := ast.SkipOuterExpressions(stmt.AsExpressionStatement().Expression, skipTransparent)
	if expr == nil || !ast.IsBinaryExpression(expr) {
		return false
	}
	bin := expr.AsBinaryExpression()
	if bin == nil || bin.OperatorToken == nil || !ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
		return false
	}
	rootName := rootObjectName(bin.Left)
	return rootName != "" && slices.Contains(ignoredVars, rootName)
}

func rootObjectName(node *ast.Node) string {
	node = ast.SkipOuterExpressions(node, skipTransparent)
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		return rootObjectName(node.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		return rootObjectName(node.AsElementAccessExpression().Expression)
	default:
		return ""
	}
}

var AlwaysReturnRule = rule.Rule{
	Name: "promise/always-return",
	Schema0: rule.Object(map[string]rule.Schema{
		"ignoreLastCallback":       rule.Bool().Default(false),
		"ignoreAssignmentVariable": rule.Union(rule.Array(rule.String())).Default([]any{"globalThis"}),
	}),
	RunWithOptions: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		optsMap, _ := options.(map[string]any)

		var ignoreAssignmentVariable []string
		if arr, ok := optsMap["ignoreAssignmentVariable"].([]any); ok {
			ignoreAssignmentVariable = make([]string, len(arr))
			for i, v := range arr {
				s, _ := v.(string)
				ignoreAssignmentVariable[i] = s
			}
		} else {
			ignoreAssignmentVariable = []string{"globalThis"}
		}

		ignoreLastCallback, _ := optsMap["ignoreLastCallback"].(bool)
		opts := Options{
			IgnoreLastCallback:       ignoreLastCallback,
			IgnoreAssignmentVariable: ignoreAssignmentVariable,
		}

		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
				checkFunction(ctx, opts, node)
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				checkFunction(ctx, opts, node)
			},
		}
	},
}

func checkFunction(ctx rule.RuleContext, opts Options, node *ast.Node) {
	if !isInlineThenFunctionExpression(node) {
		return
	}
	if opts.IgnoreLastCallback && isLastCallback(node) {
		return
	}
	if hasIgnoredAssignment(node.Body(), opts.IgnoreAssignmentVariable) && isLastCallback(node) {
		return
	}
	if !bodyAlwaysReturnsOrThrows(node) {
		ctx.ReportNode(node, buildThenShouldReturnOrThrowMessage())
	}
}
