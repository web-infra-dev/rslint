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
	return statementCompletion(stmt) == completionTerminates
}

type completion int

const (
	completionFallsThrough completion = iota
	completionTerminates
	completionStops
)

func statementListCompletion(statements []*ast.Node) completion {
	for _, stmt := range statements {
		switch statementCompletion(stmt) {
		case completionTerminates:
			return completionTerminates
		case completionStops:
			return completionStops
		}
	}
	return completionFallsThrough
}

func statementCompletion(stmt *ast.Node) completion {
	if stmt == nil {
		return completionFallsThrough
	}
	switch stmt.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement:
		return completionTerminates
	case ast.KindBreakStatement, ast.KindContinueStatement:
		return completionStops
	case ast.KindExpressionStatement:
		if isProcessExitOrAbortCall(stmt.AsExpressionStatement().Expression) {
			return completionTerminates
		}
		return completionFallsThrough
	case ast.KindBlock:
		return statementListCompletion(stmt.Statements())
	case ast.KindIfStatement:
		ifStmt := stmt.AsIfStatement()
		if ifStmt == nil || ifStmt.ElseStatement == nil {
			return completionFallsThrough
		}
		thenCompletion := statementCompletion(ifStmt.ThenStatement)
		elseCompletion := statementCompletion(ifStmt.ElseStatement)
		if thenCompletion == completionTerminates && elseCompletion == completionTerminates {
			return completionTerminates
		}
		if thenCompletion == completionStops || elseCompletion == completionStops {
			return completionStops
		}
		return completionFallsThrough
	case ast.KindTryStatement:
		tryStmt := stmt.AsTryStatement()
		if tryStmt == nil {
			return completionFallsThrough
		}
		if tryStmt.FinallyBlock != nil && blockTerminates(tryStmt.FinallyBlock) {
			return completionTerminates
		}
		if tryStmt.TryBlock == nil || !blockTerminates(tryStmt.TryBlock) {
			return completionFallsThrough
		}
		if tryStmt.CatchClause == nil {
			return completionTerminates
		}
		catchBlock := tryStmt.CatchClause.AsCatchClause().Block
		if catchBlock != nil && blockTerminates(catchBlock) {
			return completionTerminates
		}
		return completionFallsThrough
	case ast.KindSwitchStatement:
		if switchTerminates(stmt) {
			return completionTerminates
		}
		return completionFallsThrough
	case ast.KindWhileStatement:
		whileStmt := stmt.AsWhileStatement()
		if whileStmt != nil && isLiteralTrue(whileStmt.Expression) && loopBodyTerminates(stmt, whileStmt.Statement) {
			return completionTerminates
		}
		return completionFallsThrough
	case ast.KindForStatement:
		forStmt := stmt.AsForStatement()
		if forStmt != nil && (forStmt.Condition == nil || isLiteralTrue(forStmt.Condition)) && loopBodyTerminates(stmt, forStmt.Statement) {
			return completionTerminates
		}
		return completionFallsThrough
	case ast.KindDoStatement:
		doStmt := stmt.AsDoStatement()
		if doStmt != nil && loopBodyTerminates(stmt, doStmt.Statement) {
			return completionTerminates
		}
		return completionFallsThrough
	case ast.KindLabeledStatement:
		labeledStmt := stmt.AsLabeledStatement()
		if labeledStmt == nil {
			return completionFallsThrough
		}
		return statementCompletion(labeledStmt.Statement)
	default:
		return completionFallsThrough
	}
}

func switchTerminates(stmt *ast.Node) bool {
	switchStmt := stmt.AsSwitchStatement()
	if switchStmt == nil || switchStmt.CaseBlock == nil {
		return false
	}
	caseBlock := switchStmt.CaseBlock.AsCaseBlock()
	if caseBlock == nil || caseBlock.Clauses == nil || len(caseBlock.Clauses.Nodes) == 0 {
		return false
	}
	clauses := caseBlock.Clauses.Nodes
	hasDefault := false
	for _, clause := range clauses {
		if clause != nil && clause.Kind == ast.KindDefaultClause {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		return false
	}
	for i := range clauses {
		if !clausesTerminateFrom(clauses, i) {
			return false
		}
	}
	return true
}

func clausesTerminateFrom(clauses []*ast.Node, start int) bool {
	for i := start; i < len(clauses); i++ {
		clause := clauses[i]
		if clause == nil {
			continue
		}
		caseOrDefault := clause.AsCaseOrDefaultClause()
		if caseOrDefault == nil || caseOrDefault.Statements == nil {
			continue
		}
		switch statementListCompletion(caseOrDefault.Statements.Nodes) {
		case completionTerminates:
			return true
		case completionStops:
			return false
		}
	}
	return false
}

func isLiteralTrue(expr *ast.Node) bool {
	expr = ast.SkipOuterExpressions(expr, skipTransparent)
	return expr != nil && expr.Kind == ast.KindTrueKeyword
}

func loopBodyTerminates(loop *ast.Node, body *ast.Node) bool {
	return statementTerminates(body) && !containsBreakExitingLoop(loop, body)
}

func containsBreakExitingLoop(loop *ast.Node, body *ast.Node) bool {
	if loop == nil || body == nil {
		return false
	}
	found := false
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil || found {
			return found
		}
		if node != body && isFunctionLikeContainer(node) {
			return false
		}
		if node.Kind == ast.KindBreakStatement && breakExitsLoop(node, loop) {
			found = true
			return true
		}
		node.ForEachChild(visit)
		return found
	}
	visit(body)
	return found
}

func breakExitsLoop(breakStmt *ast.Node, loop *ast.Node) bool {
	stmt := breakStmt.AsBreakStatement()
	if stmt == nil {
		return false
	}
	if stmt.Label == nil {
		return nearestUnlabeledBreakTarget(breakStmt) == loop
	}
	labelName := stmt.Label.Text()
	for parent := breakStmt.Parent; parent != nil; parent = parent.Parent {
		if isFunctionLikeContainer(parent) {
			return false
		}
		if parent.Kind != ast.KindLabeledStatement {
			continue
		}
		labeledStmt := parent.AsLabeledStatement()
		if labeledStmt != nil && labeledStmt.Label != nil && labeledStmt.Label.Text() == labelName {
			return isAncestorOrSelf(parent, loop)
		}
	}
	return false
}

func nearestUnlabeledBreakTarget(node *ast.Node) *ast.Node {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindSwitchStatement, ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
			return parent
		}
		if isFunctionLikeContainer(parent) {
			return nil
		}
	}
	return nil
}

func isAncestorOrSelf(ancestor *ast.Node, node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}

func isFunctionLikeContainer(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return true
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
