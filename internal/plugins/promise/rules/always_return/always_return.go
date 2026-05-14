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

func parseOptions(options any) Options {
	opts := Options{IgnoreAssignmentVariable: []string{"globalThis"}}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["ignoreLastCallback"].(bool); ok {
		opts.IgnoreLastCallback = v
	}
	if arr, ok := optsMap["ignoreAssignmentVariable"].([]interface{}); ok {
		opts.IgnoreAssignmentVariable = make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				opts.IgnoreAssignmentVariable = append(opts.IgnoreAssignmentVariable, s)
			}
		}
	}
	return opts
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
	return blockTerminates(body)
}

func blockTerminates(block *ast.Node) bool {
	for _, stmt := range block.Statements() {
		if statementTerminates(stmt) {
			return true
		}
	}
	return false
}

func statementTerminates(stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}
	switch stmt.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement:
		return true
	case ast.KindExpressionStatement:
		return isProcessExitOrAbortCall(stmt.AsExpressionStatement().Expression)
	case ast.KindBlock:
		return blockTerminates(stmt)
	case ast.KindIfStatement:
		ifStmt := stmt.AsIfStatement()
		return ifStmt != nil && ifStmt.ElseStatement != nil && statementTerminates(ifStmt.ThenStatement) && statementTerminates(ifStmt.ElseStatement)
	case ast.KindTryStatement:
		tryStmt := stmt.AsTryStatement()
		if tryStmt == nil {
			return false
		}
		if tryStmt.FinallyBlock != nil && blockTerminates(tryStmt.FinallyBlock) {
			return true
		}
		if tryStmt.TryBlock == nil || !blockTerminates(tryStmt.TryBlock) {
			return false
		}
		if tryStmt.CatchClause == nil {
			return true
		}
		catchBlock := tryStmt.CatchClause.AsCatchClause().Block
		return catchBlock != nil && blockTerminates(catchBlock)
	default:
		return false
	}
}

func isProcessExitOrAbortCall(expr *ast.Node) bool {
	expr = ast.SkipOuterExpressions(expr, skipTransparent)
	if expr == nil || !ast.IsCallExpression(expr) {
		return false
	}
	callee := ast.SkipOuterExpressions(expr.AsCallExpression().Expression, skipTransparent)
	if callee == nil || !ast.IsPropertyAccessExpression(callee) {
		return false
	}
	prop := callee.AsPropertyAccessExpression()
	name := prop.Name()
	if name == nil || !ast.IsIdentifier(name) || (name.AsIdentifier().Text != "exit" && name.AsIdentifier().Text != "abort") {
		return false
	}
	object := ast.SkipOuterExpressions(prop.Expression, skipTransparent)
	return object != nil && ast.IsIdentifier(object) && object.AsIdentifier().Text == "process"
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
	if bin == nil || bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindEqualsToken {
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
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
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
