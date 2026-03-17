package no_this_before_super

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message builder
func buildNoBeforeSuper(kind string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noBeforeSuper",
		Description: "'" + kind + "' is not allowed before 'super()'.",
	}
}

// getClassNode gets the class declaration/expression for a constructor
func getClassNode(constructorNode *ast.Node) *ast.Node {
	if constructorNode == nil {
		return nil
	}

	current := constructorNode.Parent
	for current != nil {
		if current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
			return current
		}
		current = current.Parent
	}

	return nil
}

// hasValidExtends checks if a class extends a valid constructor (not null)
func hasValidExtends(classNode *ast.Node) bool {
	if classNode == nil {
		return false
	}

	heritageClauses := utils.GetHeritageClauses(classNode)
	if heritageClauses == nil || len(heritageClauses.Nodes) == 0 {
		return false
	}

	for _, clause := range heritageClauses.Nodes {
		if clause == nil {
			continue
		}
		heritageClause := clause.AsHeritageClause()
		if heritageClause == nil {
			continue
		}
		if heritageClause.Token == ast.KindExtendsKeyword {
			types := heritageClause.Types
			if types != nil && len(types.Nodes) > 0 {
				exprWithType := types.Nodes[0].AsExpressionWithTypeArguments()
				if exprWithType != nil && exprWithType.Expression != nil {
					// Skip extends null
					if exprWithType.Expression.Kind == ast.KindNullKeyword {
						return false
					}
					return true
				}
			}
		}
	}

	return false
}

// isFunctionOrClassBoundary checks if a node creates a new scope boundary
// for this/super references (nested functions and classes).
func isFunctionOrClassBoundary(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindClassDeclaration, ast.KindClassExpression:
		return true
	}
	return false
}

// superStatus represents the state of super() calls across code paths.
type superStatus int

const (
	// superNone means super() has not been called on any path.
	superNone superStatus = iota
	// superSome means super() has been called on some but not all paths.
	superSome
	// superAll means super() has been called on all paths.
	superAll
)

// NoThisBeforeSuperRule disallows this/super before calling super() in constructors
// https://eslint.org/docs/latest/rules/no-this-before-super
var NoThisBeforeSuperRule = rule.Rule{
	Name: "no-this-before-super",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindConstructor: func(node *ast.Node) {
				// Get the class this constructor belongs to
				classNode := getClassNode(node)
				if classNode == nil {
					return
				}

				// Only applies to derived classes (classes with extends clause)
				if !hasValidExtends(classNode) {
					return
				}

				// Get constructor body
				body := node.Body()
				if body == nil || body.Kind != ast.KindBlock {
					return
				}

				// Walk the constructor body and find this/super references before super()
				checkStatements(body.Statements(), &ctx)
			},
		}
	},
}

// checkStatements walks a list of statements sequentially and reports any
// this/super usage that occurs before super() has been called.
// Returns the super status after processing all statements.
func checkStatements(statements []*ast.Node, ctx *rule.RuleContext) superStatus {
	status := superNone

	for _, stmt := range statements {
		if stmt == nil {
			continue
		}

		// If super has already been called on all paths, no need to check further
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
		return checkStatements(stmt.Statements(), ctx)

	case ast.KindReturnStatement:
		// Check the return expression for this/super usage
		returnStmt := stmt.AsReturnStatement()
		if returnStmt != nil && returnStmt.Expression != nil && status != superAll {
			checkExpressionForViolations(returnStmt.Expression, status, ctx)
		}
		return status

	case ast.KindThrowStatement:
		// Check the throw expression for this/super usage
		expr := stmt.Expression()
		if expr != nil && status != superAll {
			checkExpressionForViolations(expr, status, ctx)
		}
		return status

	case ast.KindVariableStatement:
		// Check variable declarations for this/super usage
		if status != superAll {
			checkNodeForViolations(stmt, status, ctx)
		}
		return status

	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
		ast.KindWhileStatement, ast.KindDoStatement:
		// Check loop body and expressions for this/super usage
		if status != superAll {
			checkNodeForViolations(stmt, status, ctx)
		}
		// Loops might call super, but we can't know statically if the loop executes
		if nodeContainsSuperCall(stmt) {
			return superSome
		}
		return status

	case ast.KindSwitchStatement:
		return checkSwitchStatement(stmt, status, ctx)

	case ast.KindTryStatement:
		return checkTryStatement(stmt, status, ctx)

	default:
		// For other statement types, check for violations
		if status != superAll {
			checkNodeForViolations(stmt, status, ctx)
		}
		return status
	}
}

// checkExpressionStatement handles an expression statement.
func checkExpressionStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	expr := stmt.Expression()
	if expr == nil {
		return status
	}

	// Check if this is a direct super() call
	if isSuperCall(expr) {
		// Check arguments of super() for this/super violations first
		checkSuperCallArgs(expr, status, ctx)
		return superAll
	}

	// Check if this is a ternary with super() in both branches
	if expr.Kind == ast.KindConditionalExpression {
		condExpr := expr.AsConditionalExpression()
		if condExpr != nil {
			// Check condition for violations
			if condExpr.Condition != nil && status != superAll {
				checkExpressionForViolations(condExpr.Condition, status, ctx)
			}

			thenHasSuper := expressionContainsSuperCall(condExpr.WhenTrue)
			elseHasSuper := expressionContainsSuperCall(condExpr.WhenFalse)

			if thenHasSuper && elseHasSuper {
				return superAll
			}
			if thenHasSuper || elseHasSuper {
				// Check the non-super branch for violations
				if !thenHasSuper && status != superAll {
					checkExpressionForViolations(condExpr.WhenTrue, status, ctx)
				}
				if !elseHasSuper && status != superAll {
					checkExpressionForViolations(condExpr.WhenFalse, status, ctx)
				}
				return superSome
			}
		}
	}

	// Check if this is a logical expression with super() like `x && super()`
	if expr.Kind == ast.KindBinaryExpression {
		binExpr := expr.AsBinaryExpression()
		if binExpr != nil {
			op := binExpr.OperatorToken
			if op != nil && (op.Kind == ast.KindAmpersandAmpersandToken || op.Kind == ast.KindBarBarToken || op.Kind == ast.KindQuestionQuestionToken) {
				// Check left side for violations
				if status != superAll {
					checkExpressionForViolations(binExpr.Left, status, ctx)
				}
				// If right side has super call, it's conditional
				if expressionContainsSuperCall(binExpr.Right) {
					return superSome
				}
				if status != superAll {
					checkExpressionForViolations(binExpr.Right, status, ctx)
				}
				return status
			}
		}
	}

	// Not a super() call - check for this/super violations
	if status != superAll {
		checkExpressionForViolations(expr, status, ctx)
	}

	return status
}

// checkIfStatement handles an if statement and tracks super status across branches.
func checkIfStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	ifStmt := stmt.AsIfStatement()
	if ifStmt == nil {
		return status
	}

	// Check the condition for this/super violations
	if ifStmt.Expression != nil && status != superAll {
		checkExpressionForViolations(ifStmt.Expression, status, ctx)
	}

	// Check then branch
	thenStatus := checkBranchStatement(ifStmt.ThenStatement, status, ctx)

	// Check else branch
	if ifStmt.ElseStatement != nil {
		elseStatus := checkBranchStatement(ifStmt.ElseStatement, status, ctx)

		// Merge branch statuses
		if thenStatus == superAll && elseStatus == superAll {
			return superAll
		}
		if thenStatus == superAll || elseStatus == superAll ||
			thenStatus == superSome || elseStatus == superSome {
			return superSome
		}
		return superNone
	}

	// No else branch - if then has super, it's only some paths
	if thenStatus == superAll || thenStatus == superSome {
		return superSome
	}
	return status
}

// checkBranchStatement checks a statement that forms a branch (then/else body).
func checkBranchStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	if stmt == nil {
		return status
	}

	if stmt.Kind == ast.KindBlock {
		return checkStatements(stmt.Statements(), ctx)
	}

	return checkStatement(stmt, status, ctx)
}

// checkSwitchStatement handles switch statements.
func checkSwitchStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	// For switch, check all expressions for violations and check if super is called
	if status != superAll {
		// Check the switch expression
		switchExpr := stmt.Expression()
		if switchExpr != nil {
			checkExpressionForViolations(switchExpr, status, ctx)
		}
	}

	// Get case clauses via direct accessors
	switchStmt := stmt.AsSwitchStatement()
	if switchStmt == nil || switchStmt.CaseBlock == nil {
		return status
	}

	caseBlock := switchStmt.CaseBlock.AsCaseBlock()
	if caseBlock == nil || caseBlock.Clauses == nil || len(caseBlock.Clauses.Nodes) == 0 {
		return status
	}

	clauses := caseBlock.Clauses.Nodes

	hasDefault := false
	allHaveSuper := true

	for _, clause := range clauses {
		if clause == nil {
			continue
		}
		if clause.Kind == ast.KindDefaultClause {
			hasDefault = true
		}

		clauseStatus := checkStatements(clause.Statements(), ctx)
		if clauseStatus != superAll {
			allHaveSuper = false
		}
	}

	if hasDefault && allHaveSuper {
		return superAll
	}

	if nodeContainsSuperCall(stmt) {
		return superSome
	}

	return status
}

// checkTryStatement handles try/catch/finally statements.
func checkTryStatement(stmt *ast.Node, status superStatus, ctx *rule.RuleContext) superStatus {
	// Use direct accessors for try/catch/finally blocks
	tryStmt := stmt.AsTryStatement()
	if tryStmt == nil {
		return status
	}

	// Check the try block
	tryStatus := status
	if tryStmt.TryBlock != nil {
		tryStatus = checkStatements(tryStmt.TryBlock.Statements(), ctx)
	}

	// Check the catch block
	catchStatus := status
	if tryStmt.CatchClause != nil {
		catchBlock := tryStmt.CatchClause.AsCatchClause()
		if catchBlock != nil && catchBlock.Block != nil {
			catchStatus = checkStatements(catchBlock.Block.Statements(), ctx)
		}
	}

	// Check the finally block
	if tryStmt.FinallyBlock != nil {
		finallyStatus := checkStatements(tryStmt.FinallyBlock.Statements(), ctx)
		if finallyStatus == superAll {
			return superAll
		}
	}

	// If both try and catch have super, super is called on all paths
	if tryStmt.CatchClause != nil {
		if tryStatus == superAll && catchStatus == superAll {
			return superAll
		}
		if tryStatus == superAll || catchStatus == superAll ||
			tryStatus == superSome || catchStatus == superSome {
			return superSome
		}
	} else {
		return tryStatus
	}

	return status
}

// isSuperCall checks if an expression is a super() call.
func isSuperCall(expr *ast.Node) bool {
	if expr == nil || expr.Kind != ast.KindCallExpression {
		return false
	}
	callee := expr.Expression()
	return callee != nil && callee.Kind == ast.KindSuperKeyword
}

// expressionContainsSuperCall checks if an expression contains a super() call.
func expressionContainsSuperCall(expr *ast.Node) bool {
	if expr == nil {
		return false
	}
	if isSuperCall(expr) {
		return true
	}

	found := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		if isFunctionOrClassBoundary(node) {
			return
		}
		if isSuperCall(node) {
			found = true
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	walk(expr)
	return found
}

// nodeContainsSuperCall checks if a node (statement) contains a super() call,
// not crossing function/class boundaries.
func nodeContainsSuperCall(node *ast.Node) bool {
	if node == nil {
		return false
	}

	found := false
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}
		if isFunctionOrClassBoundary(n) {
			return
		}
		if isSuperCall(n) {
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

// checkSuperCallArgs checks the arguments of a super() call for this/super violations.
func checkSuperCallArgs(callExpr *ast.Node, status superStatus, ctx *rule.RuleContext) {
	if callExpr == nil || status == superAll {
		return
	}

	// The arguments are children of the call expression (after the callee)
	callExpr.ForEachChild(func(child *ast.Node) bool {
		if child == nil {
			return false
		}
		// Skip the super keyword itself (the callee)
		if child.Kind == ast.KindSuperKeyword {
			return false
		}
		checkExpressionForViolations(child, status, ctx)
		return false
	})
}

// checkExpressionForViolations recursively checks an expression for
// this/super usage violations.
func checkExpressionForViolations(expr *ast.Node, status superStatus, ctx *rule.RuleContext) {
	if expr == nil || status == superAll {
		return
	}

	// Don't cross function/class boundaries
	if isFunctionOrClassBoundary(expr) {
		return
	}

	switch expr.Kind {
	case ast.KindThisKeyword:
		ctx.ReportNode(expr, buildNoBeforeSuper("this"))
		return

	case ast.KindSuperKeyword:
		// super is only a violation when used for property access (super.foo),
		// not when used as super() call. The parent check handles this.
		parent := expr.Parent
		if parent != nil && parent.Kind == ast.KindCallExpression {
			// This is super() call - the callee of the call. Not a violation.
			// (The call itself would be handled at the statement level.)
			return
		}
		// super.foo or super[foo] - this is a violation
		ctx.ReportNode(expr, buildNoBeforeSuper("super"))
		return

	case ast.KindCallExpression:
		// If this is a super() call, don't report - it will be handled by checkStatement
		if isSuperCall(expr) {
			// But check arguments for violations
			checkSuperCallArgs(expr, status, ctx)
			return
		}
	}

	// Recurse into children
	expr.ForEachChild(func(child *ast.Node) bool {
		checkExpressionForViolations(child, status, ctx)
		return false
	})
}

// checkNodeForViolations walks a node (statement) looking for this/super violations,
// without crossing function/class boundaries.
func checkNodeForViolations(node *ast.Node, status superStatus, ctx *rule.RuleContext) {
	if node == nil || status == superAll {
		return
	}

	if isFunctionOrClassBoundary(node) {
		return
	}

	switch node.Kind {
	case ast.KindThisKeyword:
		ctx.ReportNode(node, buildNoBeforeSuper("this"))
		return

	case ast.KindSuperKeyword:
		parent := node.Parent
		if parent != nil && parent.Kind == ast.KindCallExpression {
			// This is the callee of super() - not a violation
			return
		}
		ctx.ReportNode(node, buildNoBeforeSuper("super"))
		return

	case ast.KindCallExpression:
		if isSuperCall(node) {
			// super() call found inside a nested expression - check args
			checkSuperCallArgs(node, status, ctx)
			return
		}
	}

	node.ForEachChild(func(child *ast.Node) bool {
		checkNodeForViolations(child, status, ctx)
		return false
	})
}
