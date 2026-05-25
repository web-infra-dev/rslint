package no_this_before_super

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildNoBeforeSuper(kind string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noBeforeSuper",
		Description: "'" + kind + "' is not allowed before 'super()'.",
	}
}

// superStatus represents the state of super() calls across code paths.
type superStatus int

const (
	superNone superStatus = iota // super() not called on any path
	superSome                    // super() called on some but not all paths
	superAll                     // super() called on all paths
)

// mergeBranches computes the combined super status when two branches merge.
// superAll only if both branches are superAll.
func mergeBranches(a, b superStatus) superStatus {
	if a == superAll && b == superAll {
		return superAll
	}
	if a == superAll || b == superAll || a == superSome || b == superSome {
		return superSome
	}
	return superNone
}

// isScopeBoundary returns true if the node creates a scope that isolates
// this/super from the enclosing constructor. Uses ast.IsFunctionLikeDeclaration
// (covers FunctionDeclaration, FunctionExpression, ArrowFunction, MethodDeclaration,
// Constructor, GetAccessor, SetAccessor) and ast.IsClassLike (ClassDeclaration,
// ClassExpression).
func isScopeBoundary(node *ast.Node) bool {
	return node != nil && (ast.IsFunctionLikeDeclaration(node) || ast.IsClassLike(node))
}


// containsSuperCall checks if a node contains a super() call anywhere,
// not crossing scope boundaries.
func containsSuperCall(node *ast.Node) bool {
	if node == nil {
		return false
	}
	found := false
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}
		if isScopeBoundary(n) {
			return
		}
		if ast.IsSuperCall(n) {
			found = true
			return
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	walk(node)
	return found
}

// bodyAlwaysExits returns true if a statement (or block) always exits via break/return/throw,
// meaning code after it in a loop is unreachable (e.g., the incrementor of a for-loop).
func bodyAlwaysExits(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindBlock:
		statements := node.Statements()
		for _, s := range statements {
			if s != nil && bodyAlwaysExits(s) {
				return true
			}
		}
		return false
	case ast.KindBreakStatement, ast.KindReturnStatement, ast.KindThrowStatement, ast.KindContinueStatement:
		return true
	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt == nil || ifStmt.ElseStatement == nil {
			return false
		}
		return bodyAlwaysExits(ifStmt.ThenStatement) && bodyAlwaysExits(ifStmt.ElseStatement)
	default:
		return false
	}
}

// NoThisBeforeSuperRule disallows this/super before calling super() in constructors.
// https://eslint.org/docs/latest/rules/no-this-before-super
var NoThisBeforeSuperRule = rule.Rule{
	Name: "no-this-before-super",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindConstructor: func(node *ast.Node) {
				classNode := ast.GetContainingClass(node)
				if classNode == nil {
					return
				}

				// Only applies to derived classes (with extends clause, not extends null)
				extendsElem := ast.GetClassExtendsHeritageElement(classNode)
				if extendsElem == nil {
					return
				}
				exprWithType := extendsElem.AsExpressionWithTypeArguments()
				if exprWithType == nil || exprWithType.Expression == nil ||
					exprWithType.Expression.Kind == ast.KindNullKeyword {
					return
				}

				body := node.Body()
				if body == nil || body.Kind != ast.KindBlock {
					return
				}

				// Check default parameter values for this/super violations
				for _, param := range node.Parameters() {
					if param == nil {
						continue
					}
					p := param.AsParameterDeclaration()
					if p != nil && p.Initializer != nil {
						checkViolations(p.Initializer, superNone, &ctx)
					}
				}

				checkStatements(body.Statements(), &ctx)
			},
		}
	},
}

// checkStatements walks a list of statements sequentially, starting from superNone.
func checkStatements(statements []*ast.Node, ctx *rule.RuleContext) superStatus {
	return checkStatementsFrom(statements, superNone, ctx)
}

// checkStatementsFrom walks a list of statements sequentially, starting from the given status.
func checkStatementsFrom(statements []*ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}
		if status == superAll {
			break
		}
		status = checkStatement(stmt, status, ctx)
	}
	return status
}

// checkStatement processes a single statement and returns updated super status.
func checkStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	if stmt == nil {
		return status
	}

	switch stmt.Kind {
	case ast.KindExpressionStatement:
		return checkExpressionStatement(stmt, status, ctx)

	case ast.KindIfStatement:
		return checkIfStatement(stmt, status, ctx)

	case ast.KindBlock:
		return checkStatementsFrom(stmt.Statements(), status, ctx)

	case ast.KindLabeledStatement:
		labeled := stmt.AsLabeledStatement()
		if labeled != nil && labeled.Statement != nil {
			return checkStatement(labeled.Statement, status, ctx)
		}
		return status

	case ast.KindReturnStatement:
		returnStmt := stmt.AsReturnStatement()
		if returnStmt != nil && returnStmt.Expression != nil {
			checkExpressionStatus(returnStmt.Expression, status, ctx)
		}
		// Return terminates the code path; subsequent statements are unreachable.
		// Using superAll causes checkStatements to skip them.
		return superAll

	case ast.KindThrowStatement:
		expr := stmt.Expression()
		if expr != nil {
			checkExpressionStatus(expr, status, ctx)
		}
		return superAll

	case ast.KindVariableStatement:
		return checkVariableStatement(stmt, status, ctx)

	case ast.KindDoStatement:
		return checkDoStatement(stmt, status, ctx)

	case ast.KindForStatement:
		return checkForStatement(stmt, status, ctx)

	case ast.KindWhileStatement:
		return checkWhileStatement(stmt, status, ctx)

	case ast.KindForInStatement, ast.KindForOfStatement:
		// The iterable/object expression is always evaluated
		exprStatus := status
		if expr := stmt.Expression(); expr != nil {
			exprStatus = checkExpressionStatus(expr, status, ctx)
		}
		// The variable declaration and body may or may not execute (0 iterations),
		// but track super status within the body to handle { super(); this.a = 0; }
		if exprStatus != superAll {
			// Check the variable binding for violations
			if initNode := stmt.Initializer(); initNode != nil {
				checkViolations(initNode, exprStatus, ctx)
			}
			if body := stmt.Statement(); body != nil {
				checkBranchBody(body, exprStatus, ctx)
			}
		}
		// If body contains super, it's conditional
		if exprStatus != superAll && stmt.Statement() != nil && containsSuperCall(stmt.Statement()) {
			return superSome
		}
		return exprStatus

	case ast.KindSwitchStatement:
		return checkSwitchStatement(stmt, status, ctx)

	case ast.KindTryStatement:
		return checkTryStatement(stmt, status, ctx)

	default:
		if status != superAll {
			checkViolations(stmt, status, ctx)
		}
		return status
	}
}

// checkExpressionStatement processes an expression statement, tracking super status
// through the expression's evaluation order.
func checkExpressionStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	expr := stmt.Expression()
	if expr == nil {
		return status
	}
	return checkExpressionStatus(expr, status, ctx)
}

// checkExpressionStatus evaluates an expression, reports violations, and returns
// the resulting super status. It follows JavaScript evaluation order.
func checkExpressionStatus(expr *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	if expr == nil {
		return status
	}

	expr = ast.SkipParentheses(expr)

	// Direct super() call
	if ast.IsSuperCall(expr) {
		checkSuperCallArgs(expr, status, ctx)
		return superAll
	}

	switch expr.Kind {
	case ast.KindConditionalExpression:
		return checkConditionalExprStatus(expr, status, ctx)

	case ast.KindBinaryExpression:
		return checkBinaryExprStatus(expr, status, ctx)

	case ast.KindVariableDeclarationList:
		// For-loop initializer: for (let x = super();;)
		return checkVarDeclList(expr, status, ctx)

	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
		ast.KindVoidExpression, ast.KindTypeOfExpression, ast.KindDeleteExpression:
		// Unary operators always evaluate their operand
		return checkExpressionChildren(expr, status, ctx)

	case ast.KindCallExpression:
		// Non-super call (super call handled above). Evaluate callee first,
		// then arguments left-to-right. e.g., super().toString() or foo(super()).
		calleeStatus := checkExpressionStatus(expr.Expression(), status, ctx)
		for _, arg := range expr.Arguments() {
			if arg != nil {
				calleeStatus = checkExpressionStatus(arg, calleeStatus, ctx)
			}
		}
		return calleeStatus

	case ast.KindNewExpression:
		// new Foo(super()) — evaluate expression then arguments
		curStatus := checkExpressionStatus(expr.Expression(), status, ctx)
		for _, arg := range expr.Arguments() {
			if arg != nil {
				curStatus = checkExpressionStatus(arg, curStatus, ctx)
			}
		}
		return curStatus

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		// e.g., super().toString — evaluate the expression/object first
		return checkExpressionStatus(expr.Expression(), status, ctx)

	case ast.KindArrayLiteralExpression:
		// Elements evaluated left to right; any may contain super()
		curStatus := status
		for _, elem := range expr.Elements() {
			if elem != nil {
				curStatus = checkExpressionStatus(elem, curStatus, ctx)
			}
		}
		return curStatus

	case ast.KindObjectLiteralExpression:
		// Property values evaluated left to right
		curStatus := status
		for _, prop := range expr.Properties() {
			if prop != nil {
				if isScopeBoundary(prop) {
					continue // getter/setter/method are scope boundaries
				}
				curStatus = checkExpressionChildren(prop, curStatus, ctx)
			}
		}
		return curStatus

	case ast.KindTemplateExpression:
		// Template: head + spans. Each span has an expression evaluated left to right.
		curStatus := status
		expr.ForEachChild(func(child *ast.Node) bool {
			if child != nil && child.Kind == ast.KindTemplateSpan {
				// TemplateSpan has Expression + Literal
				spanExpr := child.Expression()
				if spanExpr != nil {
					curStatus = checkExpressionStatus(spanExpr, curStatus, ctx)
				}
			}
			return false
		})
		return curStatus

	case ast.KindSpreadElement:
		// Spread evaluates its expression
		return checkExpressionStatus(expr.Expression(), status, ctx)

	case ast.KindTaggedTemplateExpression:
		// tag`...` — evaluate tag expression then template
		tagStatus := checkExpressionStatus(expr.AsTaggedTemplateExpression().Tag, status, ctx)
		return checkExpressionChildren(expr.AsTaggedTemplateExpression().Template, tagStatus, ctx)
	}

	// Scope boundaries (functions, classes): don't traverse into them
	if isScopeBoundary(expr) {
		return status
	}

	// Any other expression: check for violations, no status change
	if status != superAll {
		checkViolations(expr, status, ctx)
	}
	return status
}

// checkExpressionChildren recursively walks an expression's children left to right,
// tracking super status through each child expression.
func checkExpressionChildren(node *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	if node == nil {
		return status
	}
	curStatus := status
	node.ForEachChild(func(child *ast.Node) bool {
		if child != nil {
			curStatus = checkExpressionStatus(child, curStatus, ctx)
		}
		return false
	})
	return curStatus
}

// checkConditionalExprStatus handles ternary expressions (a ? b : c).
func checkConditionalExprStatus(expr *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	condExpr := expr.AsConditionalExpression()
	if condExpr == nil {
		return status
	}

	// Condition is always evaluated — track super status through it
	condStatus := status
	if condExpr.Condition != nil {
		condStatus = checkExpressionStatus(condExpr.Condition, status, ctx)
	}

	// Recursively evaluate each branch to get proper super status
	// (handles nested ternaries, super() calls, etc.)
	thenStatus := checkExpressionStatus(condExpr.WhenTrue, condStatus, ctx)
	elseStatus := checkExpressionStatus(condExpr.WhenFalse, condStatus, ctx)
	return mergeBranches(thenStatus, elseStatus)
}

// checkBinaryExprStatus handles binary expressions, including comma, logical, and assignment.
func checkBinaryExprStatus(expr *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	binExpr := expr.AsBinaryExpression()
	if binExpr == nil {
		return status
	}
	op := binExpr.OperatorToken
	if op == nil {
		return status
	}

	switch op.Kind {
	case ast.KindCommaToken:
		// Comma: left is always evaluated first, then right
		leftStatus := checkExpressionStatus(binExpr.Left, status, ctx)
		return checkExpressionStatus(binExpr.Right, leftStatus, ctx)

	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
		ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
		// Short-circuit: left is always evaluated; right is conditional
		leftStatus := checkExpressionStatus(binExpr.Left, status, ctx)
		if containsSuperCall(binExpr.Right) {
			if leftStatus == superAll {
				// Left already has super; right is conditional but left covers it
				return superAll
			}
			return superSome
		}
		if leftStatus != superAll {
			checkViolations(binExpr.Right, leftStatus, ctx)
		}
		return leftStatus

	case ast.KindEqualsToken:
		// Assignment: both sides are always evaluated (left for target, right for value)
		// Check left side for violations (but skip the assignment target identifier)
		if status != superAll {
			checkViolations(binExpr.Left, status, ctx)
		}
		// Right side is always evaluated; it can contain super()
		return checkExpressionStatus(binExpr.Right, status, ctx)

	default:
		// Other binary operators (+, -, *, in, instanceof, etc.):
		// both sides always evaluated left to right
		leftStatus := checkExpressionStatus(binExpr.Left, status, ctx)
		return checkExpressionStatus(binExpr.Right, leftStatus, ctx)
	}
}

// checkDoStatement handles do-while statements. The body always executes at least once.
func checkDoStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	doStmt := stmt.AsDoStatement()
	if doStmt == nil {
		return status
	}

	bodyStatus := status
	if doStmt.Statement != nil {
		bodyStatus = checkBranchBody(doStmt.Statement, status, ctx)
	}
	// Condition is always evaluated after body
	condStatus := bodyStatus
	if doStmt.Expression != nil {
		condStatus = checkExpressionStatus(doStmt.Expression, bodyStatus, ctx)
	}
	return condStatus
}

// checkForStatement handles for statements. The initializer always executes.
func checkForStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	forStmt := stmt.AsForStatement()
	if forStmt == nil {
		return status
	}

	// The initializer always executes
	initStatus := status
	if forStmt.Initializer != nil {
		initStatus = checkExpressionStatus(forStmt.Initializer, status, ctx)
	}

	// Condition is always evaluated at least once (to decide whether to enter the loop)
	condStatus := initStatus
	if forStmt.Condition != nil {
		condStatus = checkExpressionStatus(forStmt.Condition, initStatus, ctx)
	}

	// If no condition (for(;;)), the body always executes (like while(true))
	if forStmt.Condition == nil {
		bodyStatus := condStatus
		if forStmt.Statement != nil {
			bodyStatus = checkBranchBody(forStmt.Statement, condStatus, ctx)
		}
		return bodyStatus
	}

	// Body and incrementor may not execute (condition may be false immediately),
	// but track super status within the body to handle { super(); this.a = 0; }
	if condStatus != superAll {
		bodyEndStatus := condStatus
		if forStmt.Statement != nil {
			bodyEndStatus = checkBranchBody(forStmt.Statement, condStatus, ctx)
		}
		// Incrementor is unreachable if the body always exits (break/return/throw)
		if forStmt.Incrementor != nil && !bodyAlwaysExits(forStmt.Statement) {
			checkExpressionStatus(forStmt.Incrementor, bodyEndStatus, ctx)
		}
	}

	// If body/incrementor contain super, it's conditional (loop might not execute)
	if condStatus != superAll && (containsSuperCall(forStmt.Statement) ||
		(forStmt.Incrementor != nil && containsSuperCall(forStmt.Incrementor))) {
		return superSome
	}
	return condStatus
}

// checkWhileStatement handles while statements. The condition is always evaluated at least once.
func checkWhileStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	whileStmt := stmt.AsWhileStatement()
	if whileStmt == nil {
		return status
	}

	// The condition is always evaluated at least once
	condStatus := status
	if whileStmt.Expression != nil {
		condStatus = checkExpressionStatus(whileStmt.Expression, status, ctx)
	}

	// If condition is literal `true`, the body always executes (like do-while)
	if whileStmt.Expression != nil && whileStmt.Expression.Kind == ast.KindTrueKeyword {
		bodyStatus := condStatus
		if whileStmt.Statement != nil {
			bodyStatus = checkBranchBody(whileStmt.Statement, condStatus, ctx)
		}
		return bodyStatus
	}

	// The body may not execute (condition may be false immediately),
	// but track super status within the body to handle { super(); this.a = 0; }
	if condStatus != superAll && whileStmt.Statement != nil {
		checkBranchBody(whileStmt.Statement, condStatus, ctx)
	}

	// If body contains super, it's conditional (might not execute)
	if condStatus != superAll && whileStmt.Statement != nil && containsSuperCall(whileStmt.Statement) {
		return superSome
	}
	return condStatus
}

// checkVariableStatement handles variable declarations, tracking super in initializers.
func checkVariableStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	varStmt := stmt.AsVariableStatement()
	if varStmt == nil || varStmt.DeclarationList == nil {
		return status
	}
	return checkVarDeclList(varStmt.DeclarationList, status, ctx)
}

// checkVarDeclList walks a variable declaration list, tracking super in initializers.
func checkVarDeclList(node *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	declList := node.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return status
	}

	for _, decl := range declList.Declarations.Nodes {
		if decl == nil {
			continue
		}
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil {
			continue
		}
		// Check binding pattern defaults (e.g., const { a = this.b } = {})
		if name := decl.Name(); name != nil && status != superAll {
			checkViolations(name, status, ctx)
		}
		// Check initializer for both violations and super status tracking
		if varDecl.Initializer != nil {
			status = checkExpressionStatus(varDecl.Initializer, status, ctx)
		}
	}
	return status
}

// checkIfStatement handles if/else branching.
func checkIfStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	ifStmt := stmt.AsIfStatement()
	if ifStmt == nil {
		return status
	}

	// Condition is always evaluated — track super status through it
	condStatus := status
	if ifStmt.Expression != nil {
		condStatus = checkExpressionStatus(ifStmt.Expression, status, ctx)
	}

	thenStatus := checkBranchBody(ifStmt.ThenStatement, condStatus, ctx)

	if ifStmt.ElseStatement != nil {
		elseStatus := checkBranchBody(ifStmt.ElseStatement, condStatus, ctx)
		return mergeBranches(thenStatus, elseStatus)
	}

	// No else: if then has super, it's only some paths
	if thenStatus == superAll || thenStatus == superSome {
		return superSome
	}
	return status
}

// checkBranchBody checks a statement that forms a branch body (then/else/case).
func checkBranchBody(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	if stmt == nil {
		return status
	}
	if stmt.Kind == ast.KindBlock {
		return checkStatementsFrom(stmt.Statements(), status, ctx)
	}
	return checkStatement(stmt, status, ctx)
}

// checkSwitchStatement handles switch statements.
func checkSwitchStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	// Switch discriminant is always evaluated — track super status through it
	exprStatus := status
	if switchExpr := stmt.Expression(); switchExpr != nil {
		exprStatus = checkExpressionStatus(switchExpr, status, ctx)
	}

	switchStmt := stmt.AsSwitchStatement()
	if switchStmt == nil || switchStmt.CaseBlock == nil {
		return exprStatus
	}
	caseBlock := switchStmt.CaseBlock.AsCaseBlock()
	if caseBlock == nil || caseBlock.Clauses == nil || len(caseBlock.Clauses.Nodes) == 0 {
		return exprStatus
	}

	hasDefault := false
	allHaveSuper := true
	clauses := caseBlock.Clauses.Nodes

	for i, clause := range clauses {
		if clause == nil {
			continue
		}
		if clause.Kind == ast.KindDefaultClause {
			hasDefault = true
		}

		stmtList := clause.Statements()
		// Compute this clause's status by running its statements
		clauseStatus := checkStatementsFrom(stmtList, exprStatus, ctx)
		if clauseStatus == superAll {
			continue // This clause guarantees super
		}

		// Check if this clause falls through (empty or non-breaking statements)
		// and reaches super in a subsequent clause
		if !clauseBodyBreaks(stmtList) {
			// Falls through — accumulate statements from subsequent clauses
			chainStatus := clauseStatus
			chainFound := false
			for j := i + 1; j < len(clauses); j++ {
				nextClause := clauses[j]
				if nextClause == nil {
					continue
				}
				nextStatements := nextClause.Statements()
				chainStatus = checkStatementsFrom(nextStatements, chainStatus, ctx)
				if chainStatus == superAll {
					chainFound = true
					break
				}
			}
			if chainFound {
				continue // Fallthrough chain reaches super
			}
		}

		allHaveSuper = false
	}

	if hasDefault && allHaveSuper {
		return superAll
	}
	if containsSuperCall(stmt) {
		return superSome
	}
	return exprStatus
}

// clauseBodyBreaks returns true if a switch case's statement list definitely
// breaks/returns/throws (i.e., does NOT fall through to the next case).
func clauseBodyBreaks(statements []*ast.Node) bool {
	for _, s := range statements {
		if s != nil && bodyAlwaysExits(s) {
			return true
		}
	}
	return false
}

// checkTryStatement handles try/catch/finally statements.
func checkTryStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	tryStmt := stmt.AsTryStatement()
	if tryStmt == nil {
		return status
	}

	tryStatus := status
	if tryStmt.TryBlock != nil {
		tryStatus = checkStatementsFrom(tryStmt.TryBlock.Statements(), status, ctx)
	}

	if tryStmt.FinallyBlock != nil {
		// With a finally block: try might throw at any point, so super() in try
		// is uncertain. The finally block sees the worst case — the status before
		// try started (an exception could occur before super() in try).
		// If there's a catch, it starts from the pre-try status as well.
		mergedStatus := status // worst case: try threw before super()
		if tryStatus == superAll && status == superAll {
			mergedStatus = superAll
		} else if tryStatus != superNone || status != superNone {
			// If try had super or started with some, it's conditional
			if tryStatus == superAll || tryStatus == superSome {
				mergedStatus = superSome
			}
		}

		if tryStmt.CatchClause != nil {
			catchStatus := status
			catchBlock := tryStmt.CatchClause.AsCatchClause()
			if catchBlock != nil && catchBlock.Block != nil {
				catchStatus = checkStatementsFrom(catchBlock.Block.Statements(), status, ctx)
			}
			mergedStatus = mergeBranches(mergedStatus, catchStatus)
		}

		finallyStatus := checkStatementsFrom(tryStmt.FinallyBlock.Statements(), mergedStatus, ctx)
		if finallyStatus == superAll {
			return superAll
		}
		return finallyStatus
	}

	// No finally: merge try and catch branches
	mergedStatus := tryStatus
	if tryStmt.CatchClause != nil {
		catchStatus := status
		catchBlock := tryStmt.CatchClause.AsCatchClause()
		if catchBlock != nil && catchBlock.Block != nil {
			catchStatus = checkStatementsFrom(catchBlock.Block.Statements(), status, ctx)
		}
		mergedStatus = mergeBranches(tryStatus, catchStatus)
	}

	return mergedStatus
}

// checkSuperCallArgs checks super() arguments for this/super violations.
func checkSuperCallArgs(callExpr *ast.Node, status superStatus, ctx *rule.RuleContext) {
	if callExpr == nil || status == superAll {
		return
	}
	callExpr.ForEachChild(func(child *ast.Node) bool {
		if child == nil || child.Kind == ast.KindSuperKeyword {
			return false
		}
		checkViolations(child, status, ctx)
		return false
	})
}

// checkViolations recursively checks a node for this/super usage violations,
// not crossing scope boundaries.
func checkViolations(node *ast.Node, status superStatus, ctx *rule.RuleContext) {
	if node == nil || status == superAll {
		return
	}
	if isScopeBoundary(node) {
		return
	}

	switch node.Kind {
	case ast.KindThisKeyword:
		ctx.ReportNode(node, buildNoBeforeSuper("this"))
		return

	case ast.KindSuperKeyword:
		parent := node.Parent
		if parent != nil && parent.Kind == ast.KindCallExpression {
			return // super() callee, not a violation
		}
		ctx.ReportNode(node, buildNoBeforeSuper("super"))
		return

	case ast.KindCallExpression:
		if ast.IsSuperCall(node) {
			checkSuperCallArgs(node, status, ctx)
			return
		}
	}

	node.ForEachChild(func(child *ast.Node) bool {
		checkViolations(child, status, ctx)
		return false
	})
}
