package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// FunctionReturnAnalysis holds the result of analyzing return behavior of a function.
type FunctionReturnAnalysis struct {
	EndReachable       bool // Whether the function's end is reachable (can fall through without return/throw)
	HasReturnWithValue bool // Whether any return statement returns a value
	HasEmptyReturn     bool // Whether any return statement is empty (return;)
}

// AnalyzeFunctionReturns analyzes a function node's return behavior using
// the binder's control flow graph and ForEachReturnStatement.
// The node must be a function-like node (FunctionDeclaration, FunctionExpression,
// ArrowFunction, Constructor, MethodDeclaration, GetAccessor, SetAccessor).
//
// The binder sets EndFlowNode on function bodies only when the function end is
// reachable (i.e., some code path falls through without returning or throwing).
// If EndFlowNode is nil and the body is present, all paths return or throw.
func AnalyzeFunctionReturns(node *ast.Node) FunctionReturnAnalysis {
	result := FunctionReturnAnalysis{
		EndReachable: true,
	}

	if node == nil {
		return result
	}

	body := node.Body()
	if body == nil {
		return result
	}

	// The binder sets NodeFlagsHasImplicitReturn when the function end is reachable.
	// This flag is set at the same time as EndFlowNode, but is simpler to check.
	result.EndReachable = node.Flags&ast.NodeFlagsHasImplicitReturn != 0

	// Scan return statements (ForEachReturnStatement skips nested functions)
	ast.ForEachReturnStatement(body, func(stmt *ast.Node) bool {
		if stmt.Expression() != nil {
			result.HasReturnWithValue = true
		} else {
			result.HasEmptyReturn = true
		}
		return false
	})

	return result
}

// IsFunctionEndReachable checks if a function's end is reachable (can fall through
// without a return or throw statement). Uses the binder's NodeFlagsHasImplicitReturn flag.
func IsFunctionEndReachable(node *ast.Node) bool {
	if node == nil {
		return true
	}
	return node.Flags&ast.NodeFlagsHasImplicitReturn != 0
}

// CanBlockThrow checks if a block can throw before reaching a non-throwing
// terminal. Used to determine if a catch clause is reachable.
//
// A block "can throw" if it contains any statement that may raise an exception
// before control reaches a guaranteed non-throwing terminal (break, continue,
// or return without expression). Specifically:
//   - break / continue: non-throwing terminals → returns false
//   - return (no expression): non-throwing → returns false
//   - return (with expression): expression evaluation may throw → returns true
//   - throw: always throws → returns true
//   - empty statement: no effect → continues checking
//   - nested block: recurses
//   - try with finally that terminates: finally overrides → returns false
//   - any other statement (expression, if, for, etc.): may throw → returns true
func CanBlockThrow(block *ast.Node) bool {
	statements := block.Statements()
	if len(statements) == 0 {
		return false
	}
	for _, stmt := range statements {
		switch stmt.Kind {
		case ast.KindBreakStatement, ast.KindContinueStatement:
			return false
		case ast.KindReturnStatement:
			rs := stmt.AsReturnStatement()
			return rs != nil && rs.Expression != nil
		case ast.KindThrowStatement:
			return true
		case ast.KindEmptyStatement:
			continue
		case ast.KindBlock:
			return CanBlockThrow(stmt)
		case ast.KindTryStatement:
			ts := stmt.AsTryStatement()
			if ts != nil && ts.FinallyBlock != nil && BlockEndsWithTerminal(ts.FinallyBlock) {
				return false
			}
			return true
		default:
			return true
		}
	}
	return true
}

// BlockEndsWithTerminal checks if a block's last statement is a control flow
// terminal (break/return/throw/continue), possibly nested in inner blocks.
func BlockEndsWithTerminal(block *ast.Node) bool {
	nodes := block.Statements()
	if len(nodes) == 0 {
		return false
	}
	last := nodes[len(nodes)-1]
	switch last.Kind {
	case ast.KindBreakStatement, ast.KindContinueStatement,
		ast.KindReturnStatement, ast.KindThrowStatement:
		return true
	case ast.KindBlock:
		return BlockEndsWithTerminal(last)
	}
	return false
}

// skipOEKParentheses strips only parentheses when walking outer expressions.
const skipOEKParentheses = ast.OEKParentheses

// IsFunctionLikeContainer reports whether node introduces a new function scope.
// Covers function declarations/expressions, arrow functions, methods, getters,
// setters, and constructors. Use this to stop traversals at function boundaries.
func IsFunctionLikeContainer(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return true
	}
	return false
}

// Completion represents the control-flow completion kind of a statement.
type Completion int

const (
	CompletionFallsThrough Completion = iota // execution continues normally to the next statement
	CompletionTerminates                     // return / throw / process.exit terminates the function
	CompletionStops                          // break / continue exits the current loop or switch arm
)

// StatementListCompletion returns the completion of executing a sequence of
// statements in order. It stops at the first non-FallsThrough statement.
func StatementListCompletion(statements []*ast.Node) Completion {
	for _, stmt := range statements {
		switch StatementCompletion(stmt) {
		case CompletionTerminates:
			return CompletionTerminates
		case CompletionStops:
			return CompletionStops
		}
	}
	return CompletionFallsThrough
}

// StatementCompletion returns the control-flow completion of a single statement,
// recursively analyzing compound statements (if/try/switch/loops/labels).
// process.exit() and process.abort() are treated as CompletionTerminates.
func StatementCompletion(stmt *ast.Node) Completion {
	if stmt == nil {
		return CompletionFallsThrough
	}
	switch stmt.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement:
		return CompletionTerminates
	case ast.KindBreakStatement, ast.KindContinueStatement:
		return CompletionStops
	case ast.KindExpressionStatement:
		if isProcessExitOrAbortCall(stmt.AsExpressionStatement().Expression) {
			return CompletionTerminates
		}
		return CompletionFallsThrough
	case ast.KindBlock:
		return StatementListCompletion(stmt.Statements())
	case ast.KindIfStatement:
		ifStmt := stmt.AsIfStatement()
		if ifStmt == nil || ifStmt.ElseStatement == nil {
			return CompletionFallsThrough
		}
		thenC := StatementCompletion(ifStmt.ThenStatement)
		elseC := StatementCompletion(ifStmt.ElseStatement)
		if thenC == CompletionTerminates && elseC == CompletionTerminates {
			return CompletionTerminates
		}
		if thenC == CompletionStops || elseC == CompletionStops {
			return CompletionStops
		}
		return CompletionFallsThrough
	case ast.KindTryStatement:
		tryStmt := stmt.AsTryStatement()
		if tryStmt == nil {
			return CompletionFallsThrough
		}
		if tryStmt.FinallyBlock != nil && StatementListCompletion(tryStmt.FinallyBlock.Statements()) == CompletionTerminates {
			return CompletionTerminates
		}
		if tryStmt.TryBlock == nil || StatementListCompletion(tryStmt.TryBlock.Statements()) != CompletionTerminates {
			return CompletionFallsThrough
		}
		if tryStmt.CatchClause == nil {
			return CompletionTerminates
		}
		catchBlock := tryStmt.CatchClause.AsCatchClause().Block
		if catchBlock != nil && StatementListCompletion(catchBlock.Statements()) == CompletionTerminates {
			return CompletionTerminates
		}
		if tryBlockCannotReachCatch(tryStmt.TryBlock) {
			return CompletionTerminates
		}
		return CompletionFallsThrough
	case ast.KindSwitchStatement:
		if switchTerminatesAll(stmt) {
			return CompletionTerminates
		}
		return CompletionFallsThrough
	case ast.KindWhileStatement:
		whileStmt := stmt.AsWhileStatement()
		if whileStmt != nil && isTruthyLiteral(whileStmt.Expression) && !containsBreakExitingLoop(stmt, whileStmt.Statement) {
			return CompletionTerminates
		}
		return CompletionFallsThrough
	case ast.KindForStatement:
		forStmt := stmt.AsForStatement()
		if forStmt != nil && (forStmt.Condition == nil || isTruthyLiteral(forStmt.Condition)) && !containsBreakExitingLoop(stmt, forStmt.Statement) {
			return CompletionTerminates
		}
		return CompletionFallsThrough
	case ast.KindDoStatement:
		doStmt := stmt.AsDoStatement()
		if doStmt == nil {
			return CompletionFallsThrough
		}
		if isTruthyLiteral(doStmt.Expression) {
			if !containsBreakExitingLoop(stmt, doStmt.Statement) {
				return CompletionTerminates
			}
			return CompletionFallsThrough
		}
		if loopBodyTerminates(stmt, doStmt.Statement) {
			return CompletionTerminates
		}
		return CompletionFallsThrough
	case ast.KindLabeledStatement:
		labeledStmt := stmt.AsLabeledStatement()
		if labeledStmt == nil {
			return CompletionFallsThrough
		}
		return StatementCompletion(labeledStmt.Statement)
	default:
		return CompletionFallsThrough
	}
}

func switchTerminatesAll(stmt *ast.Node) bool {
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
		if !clausesTerminateFrom(clauses, i, stmt) {
			return false
		}
	}
	return true
}

func clausesTerminateFrom(clauses []*ast.Node, start int, switchNode *ast.Node) bool {
	for i := start; i < len(clauses); i++ {
		clause := clauses[i]
		if clause == nil {
			continue
		}
		caseOrDefault := clause.AsCaseOrDefaultClause()
		if caseOrDefault == nil || caseOrDefault.Statements == nil {
			continue
		}
		switch caseClauseCompletion(caseOrDefault.Statements.Nodes, switchNode) {
		case CompletionTerminates:
			return true
		case CompletionStops:
			return false
		}
	}
	return false
}

// caseClauseCompletion is StatementListCompletion specialized for switch clause
// bodies: a break that exits switchNode along a reachable path counts as
// CompletionStops (the clause leaves the switch rather than terminating the
// function). This catches a conditional break such as `if (c) break;` that a
// later `return` would otherwise mask.
func caseClauseCompletion(statements []*ast.Node, switchNode *ast.Node) Completion {
	for _, stmt := range statements {
		if stmtCanBreakOutOfSwitch(stmt, switchNode) {
			return CompletionStops
		}
		switch StatementCompletion(stmt) {
		case CompletionTerminates:
			return CompletionTerminates
		case CompletionStops:
			return CompletionStops
		}
	}
	return CompletionFallsThrough
}

// stmtCanBreakOutOfSwitch reports whether executing stmt can reach a break that
// exits switchNode without first terminating. It descends into blocks and if
// branches (tracking reachability) but not into nested loops/switches, whose
// unlabeled breaks target themselves.
func stmtCanBreakOutOfSwitch(stmt *ast.Node, switchNode *ast.Node) bool {
	if stmt == nil {
		return false
	}
	switch stmt.Kind {
	case ast.KindBreakStatement:
		return breakExitsTarget(stmt, switchNode)
	case ast.KindBlock:
		for _, s := range stmt.Statements() {
			if stmtCanBreakOutOfSwitch(s, switchNode) {
				return true
			}
			switch StatementCompletion(s) {
			case CompletionTerminates, CompletionStops:
				return false
			}
		}
		return false
	case ast.KindIfStatement:
		ifStmt := stmt.AsIfStatement()
		if ifStmt == nil {
			return false
		}
		if stmtCanBreakOutOfSwitch(ifStmt.ThenStatement, switchNode) {
			return true
		}
		return ifStmt.ElseStatement != nil && stmtCanBreakOutOfSwitch(ifStmt.ElseStatement, switchNode)
	case ast.KindLabeledStatement:
		labeledStmt := stmt.AsLabeledStatement()
		if labeledStmt == nil {
			return false
		}
		return stmtCanBreakOutOfSwitch(labeledStmt.Statement, switchNode)
	default:
		return false
	}
}

func isTruthyLiteral(expr *ast.Node) bool {
	expr = ast.SkipOuterExpressions(expr, skipOEKParentheses)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindTrueKeyword:
		return true
	case ast.KindNumericLiteral:
		return NormalizeNumericLiteral(expr.AsNumericLiteral().Text) != "0"
	case ast.KindStringLiteral:
		return expr.AsStringLiteral().Text != ""
	case ast.KindNoSubstitutionTemplateLiteral:
		return expr.AsNoSubstitutionTemplateLiteral().Text != ""
	default:
		return false
	}
}

func loopBodyTerminates(loop *ast.Node, body *ast.Node) bool {
	return StatementCompletion(body) == CompletionTerminates && !containsBreakExitingLoop(loop, body)
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
		if node != body && IsFunctionLikeContainer(node) {
			return false
		}
		if node.Kind == ast.KindBreakStatement && breakExitsTarget(node, loop) {
			found = true
			return true
		}
		node.ForEachChild(visit)
		return found
	}
	visit(body)
	return found
}

// breakExitsTarget reports whether breakStmt exits target, which may be a loop
// or a switch statement.
func breakExitsTarget(breakStmt *ast.Node, target *ast.Node) bool {
	stmt := breakStmt.AsBreakStatement()
	if stmt == nil {
		return false
	}
	if stmt.Label == nil {
		return nearestUnlabeledBreakTarget(breakStmt) == target
	}
	labelName := stmt.Label.Text()
	for parent := breakStmt.Parent; parent != nil; parent = parent.Parent {
		if IsFunctionLikeContainer(parent) {
			return false
		}
		if parent.Kind != ast.KindLabeledStatement {
			continue
		}
		labeledStmt := parent.AsLabeledStatement()
		if labeledStmt != nil && labeledStmt.Label != nil && labeledStmt.Label.Text() == labelName {
			return isAncestorOrSelf(parent, target)
		}
	}
	return false
}

func nearestUnlabeledBreakTarget(node *ast.Node) *ast.Node {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindSwitchStatement, ast.KindWhileStatement, ast.KindDoStatement,
			ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
			return parent
		}
		if IsFunctionLikeContainer(parent) {
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

func isProcessExitOrAbortCall(expr *ast.Node) bool {
	expr = ast.SkipOuterExpressions(expr, skipOEKParentheses)
	if expr == nil || !ast.IsCallExpression(expr) {
		return false
	}
	callee := ast.SkipOuterExpressions(expr.AsCallExpression().Expression, skipOEKParentheses)
	if callee == nil || !ast.IsPropertyAccessExpression(callee) {
		return false
	}
	prop := callee.AsPropertyAccessExpression()
	name := prop.Name()
	if name == nil || !ast.IsIdentifier(name) || (name.AsIdentifier().Text != "exit" && name.AsIdentifier().Text != "abort") {
		return false
	}
	object := ast.SkipOuterExpressions(prop.Expression, skipOEKParentheses)
	return object != nil && ast.IsIdentifier(object) && object.AsIdentifier().Text == "process"
}

// tryBlockCannotReachCatch reports whether the try block reaches a return whose
// argument cannot throw, with every preceding statement also unable to throw, so
// the catch clause is unreachable. In that case a non-terminating catch does not
// stop the function from always returning. Mirrors ESLint's code-path analysis,
// which only routes to catch when a statement may throw.
func tryBlockCannotReachCatch(tryBlock *ast.Node) bool {
	if tryBlock == nil {
		return false
	}
	for _, s := range tryBlock.Statements() {
		if s == nil {
			return false
		}
		switch s.Kind {
		case ast.KindReturnStatement:
			arg := s.AsReturnStatement().Expression
			return arg == nil || expressionCannotThrow(arg)
		case ast.KindThrowStatement:
			return false
		default:
			if !statementCannotThrow(s) {
				return false
			}
		}
	}
	return false
}

// statementCannotThrow reports whether a non-terminating statement preceding a
// return in a try block cannot throw. Only simple side-effect-free forms qualify;
// anything else is treated conservatively as possibly throwing.
func statementCannotThrow(stmt *ast.Node) bool {
	switch stmt.Kind {
	case ast.KindEmptyStatement:
		return true
	case ast.KindExpressionStatement:
		return expressionCannotThrow(stmt.AsExpressionStatement().Expression)
	case ast.KindVariableStatement:
		list := stmt.AsVariableStatement().DeclarationList
		if list == nil {
			return false
		}
		for _, d := range list.AsVariableDeclarationList().Declarations.Nodes {
			if d == nil {
				return false
			}
			decl := d.AsVariableDeclaration()
			if name := decl.Name(); name == nil || !ast.IsIdentifier(name) {
				return false
			}
			if decl.Initializer != nil && !expressionCannotThrow(decl.Initializer) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// expressionCannotThrow reports whether evaluating expr cannot throw. It is a
// conservative whitelist mirroring ESLint: literals and operations whose
// operands all cannot throw. Anything that dereferences a binding (identifier,
// member access, call, new, spread, computed key, or a template/expression with
// a variable) may throw and returns false.
func expressionCannotThrow(expr *ast.Node) bool {
	expr = ast.SkipOuterExpressions(expr, skipOEKParentheses)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	case ast.KindPrefixUnaryExpression:
		return expressionCannotThrow(expr.AsPrefixUnaryExpression().Operand)
	case ast.KindVoidExpression:
		return expressionCannotThrow(expr.AsVoidExpression().Expression)
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		return expressionCannotThrow(bin.Left) && expressionCannotThrow(bin.Right)
	case ast.KindConditionalExpression:
		cond := expr.AsConditionalExpression()
		return expressionCannotThrow(cond.Condition) &&
			expressionCannotThrow(cond.WhenTrue) &&
			expressionCannotThrow(cond.WhenFalse)
	case ast.KindObjectLiteralExpression:
		for _, prop := range expr.AsObjectLiteralExpression().Properties.Nodes {
			if prop == nil || prop.Kind != ast.KindPropertyAssignment {
				return false
			}
			pa := prop.AsPropertyAssignment()
			if name := pa.Name(); name != nil && ast.IsComputedPropertyName(name) {
				return false
			}
			if !expressionCannotThrow(pa.Initializer) {
				return false
			}
		}
		return true
	case ast.KindArrayLiteralExpression:
		for _, el := range expr.AsArrayLiteralExpression().Elements.Nodes {
			if el == nil || el.Kind == ast.KindSpreadElement || !expressionCannotThrow(el) {
				return false
			}
		}
		return true
	case ast.KindTemplateExpression:
		for _, span := range expr.AsTemplateExpression().TemplateSpans.Nodes {
			if span == nil || !expressionCannotThrow(span.AsTemplateSpan().Expression) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
