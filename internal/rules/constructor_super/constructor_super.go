package constructor_super

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message builders
func buildMissingAll() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAll",
		Description: "Expected to call 'super()' in all paths.",
	}
}

func buildMissingSome() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingSome",
		Description: "Expected to call 'super()' in some paths.",
	}
}

func buildDuplicate() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "duplicate",
		Description: "Unexpected duplicate 'super()' call.",
	}
}

func buildBadSuper() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "badSuper",
		Description: "Unexpected 'super()' call because this class does not extend a valid constructor.",
	}
}

// isConstructor checks if a node is a constructor method
func isConstructor(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check if it's a constructor method
	if node.Kind == ast.KindConstructor {
		return true
	}

	// For method declarations, check if it's named "constructor"
	if node.Kind == ast.KindMethodDeclaration {
		name := node.Name()
		if name != nil && name.Kind == ast.KindIdentifier && name.Text() == "constructor" {
			return true
		}
	}

	return false
}

// getClassNode gets the class declaration/expression for a constructor
func getClassNode(constructorNode *ast.Node) *ast.Node {
	if constructorNode == nil {
		return nil
	}

	// Walk up the parent chain to find the class
	current := constructorNode.Parent
	for current != nil {
		if current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
			return current
		}
		current = current.Parent
	}

	return nil
}

// hasValidExtends checks if a class extends a valid constructor
func hasValidExtends(classNode *ast.Node) bool {
	if classNode == nil {
		return false
	}

	// Get heritage clause (extends clause)
	heritageClauses := utils.GetHeritageClauses(classNode)
	if heritageClauses == nil || len(heritageClauses.Nodes) == 0 {
		return false
	}

	// Look for extends clause (token = ExtendsKeyword)
	for _, clause := range heritageClauses.Nodes {
		if clause == nil {
			continue
		}
		heritageClause := clause.AsHeritageClause()
		if heritageClause == nil {
			continue
		}
		// Check if this is an extends clause
		if heritageClause.Token == ast.KindExtendsKeyword {
			types := heritageClause.Types
			if types != nil && len(types.Nodes) > 0 {
				// Check if the extended type is valid (not null, not a primitive)
				exprWithType := types.Nodes[0].AsExpressionWithTypeArguments()
				if exprWithType != nil {
					extendsExpr := exprWithType.Expression
					if extendsExpr != nil && !isInvalidExtends(extendsExpr) {
						return true
					}
				}
			}
		}
	}

	return false
}

// isInvalidExtends checks if an extends expression is invalid (null, literal, etc.)
func isInvalidExtends(node *ast.Node) bool {
	if node == nil {
		return true
	}

	switch node.Kind {
	case ast.KindNullKeyword:
		return true
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindParenthesizedExpression:
		// Check inner expression
		expr := node.Expression()
		return isInvalidExtends(expr)
	case ast.KindBinaryExpression:
		// For binary expressions, check if the result could be a constructor
		// Assignment like (B = C) is valid if C could be a constructor
		return !isPossibleConstructor(node)
	}

	return false
}

// isPossibleConstructor checks if an expression could be a constructor
// This handles cases like: class A extends (B = C) where we need to check if the result is a constructor
func isPossibleConstructor(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindClassExpression:
		return true
	case ast.KindIdentifier:
		// Could be a constructor reference
		return true
	case ast.KindParenthesizedExpression:
		// Check inner expression
		expr := node.Expression()
		return isPossibleConstructor(expr)
	case ast.KindBinaryExpression:
		// For assignments like (B = C), check the right side
		binExpr := node.AsBinaryExpression()
		if binExpr != nil && binExpr.OperatorToken != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken {
			// Assignment - check if right side could be a constructor
			return isPossibleConstructor(binExpr.Right)
		}
		// For logical expressions (&&, ||), check the right side
		if binExpr != nil && binExpr.OperatorToken != nil {
			if binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken ||
			   binExpr.OperatorToken.Kind == ast.KindBarBarToken {
				return isPossibleConstructor(binExpr.Right)
			}
		}
		// For other operators (+=, -=, *=, /=), it's not a valid constructor
		return false
	case ast.KindConditionalExpression:
		// For ternary, both branches must be constructors
		condExpr := node.AsConditionalExpression()
		if condExpr != nil {
			return isPossibleConstructor(condExpr.WhenTrue) && isPossibleConstructor(condExpr.WhenFalse)
		}
		return false
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		// Could be a constructor reference like Class.Static
		return true
	case ast.KindCallExpression:
		// Could return a constructor
		return true
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		// Literals are definitely not constructors
		return false
	}

	// For other node types, assume it could be a constructor
	return true
}

// findDuplicateSuperCalls finds super() calls that could execute sequentially (real duplicates)
// Returns super() calls that are duplicates (excluding the first one)
func findDuplicateSuperCalls(body *ast.Node, superCallLocations []*ast.Node) []*ast.Node {
	if body == nil || body.Kind != ast.KindBlock {
		return nil
	}

	statements := body.Statements()
	var duplicates []*ast.Node
	foundFirstSuper := false

	for _, stmt := range statements {
		if stmt == nil {
			continue
		}

		// Check if this statement directly contains a super() call
		if hasSuperCall(stmt) {
			if foundFirstSuper {
				// This is a duplicate - super after another super at the same level
				// Find the corresponding super call node
				for _, loc := range superCallLocations {
					if isSuperCallInStatement(stmt, loc) {
						duplicates = append(duplicates, loc)
					}
				}
			} else {
				foundFirstSuper = true
			}
		}

		// After finding first super, any additional super calls anywhere are duplicates
		if foundFirstSuper {
			// Find all super calls in this statement (recursively)
			findSuperCallsInNode(stmt, superCallLocations, &duplicates)
		}

		// Check branching statements that have super in all branches
		switch stmt.Kind {
		case ast.KindIfStatement:
			ifStmt := stmt.AsIfStatement()
			if ifStmt != nil && ifStmt.ElseStatement != nil {
				// If-else with super in both branches
				if statementHasSuper(ifStmt.ThenStatement) && statementHasSuper(ifStmt.ElseStatement) {
					foundFirstSuper = true
				}
			}
		case ast.KindSwitchStatement:
			if switchHasSuper(stmt) {
				foundFirstSuper = true
			}
		case ast.KindExpressionStatement:
			// Check ternary expression with super in both branches
			expr := stmt.Expression()
			if expr != nil && expr.Kind == ast.KindConditionalExpression {
				condExpr := expr.AsConditionalExpression()
				if condExpr != nil {
					if expressionHasSuper(condExpr.WhenTrue) && expressionHasSuper(condExpr.WhenFalse) {
						foundFirstSuper = true
					}
				}
			}
		}
	}

	return duplicates
}

// findSuperCallsInNode finds super() calls in a node (excluding direct super calls at the statement level)
func findSuperCallsInNode(node *ast.Node, superCallLocations []*ast.Node, duplicates *[]*ast.Node) {
	if node == nil {
		return
	}

	// Don't check direct expression statements with super (already handled)
	if node.Kind == ast.KindExpressionStatement {
		if hasSuperCall(node) {
			return
		}
	}

	// Check if this node contains any super calls
	for _, loc := range superCallLocations {
		if isSuperCallInNode(node, loc) && !contains(*duplicates, loc) {
			*duplicates = append(*duplicates, loc)
		}
	}
}

// isSuperCallInNode checks if a super call is inside a node (not at the direct statement level)
func isSuperCallInNode(node *ast.Node, superCall *ast.Node) bool {
	// Don't match if this IS the super call itself at expression statement level
	if node.Kind == ast.KindExpressionStatement {
		expr := node.Expression()
		if expr != nil && expr == superCall {
			return false
		}
	}

	return isSuperCallInStatement(node, superCall)
}

// contains checks if a slice contains a node
func contains(slice []*ast.Node, item *ast.Node) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// isSuperCallInStatement checks if a super call node is contained in a statement
func isSuperCallInStatement(stmt *ast.Node, superCall *ast.Node) bool {
	found := false
	var check func(*ast.Node)
	check = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		if node == superCall {
			found = true
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			check(child)
			return found // Stop if found
		})
	}
	check(stmt)
	return found
}

// hasBranchingWithEarlyReturn checks if there's any branching where some paths have early return/throw
func hasBranchingWithEarlyReturn(body *ast.Node) bool {
	if body == nil || body.Kind != ast.KindBlock {
		return false
	}

	statements := body.Statements()
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}

		// Check if this is an if statement with early return in one branch but not all
		if stmt.Kind == ast.KindIfStatement {
			ifStmt := stmt.AsIfStatement()
			if ifStmt != nil {
				thenTerminates := branchTerminates(ifStmt.ThenStatement)
				elseStmt := ifStmt.ElseStatement

				if elseStmt != nil {
					elseTerminates := branchTerminates(elseStmt)
					// If one branch terminates but not both, there's branching with partial early return
					if thenTerminates != elseTerminates {
						return true
					}
				} else {
					// If only has then branch and it terminates, that's branching with partial early return
					if thenTerminates {
						return true
					}
				}
			}
		}

		// Check for switch statements with partial early returns
		if stmt.Kind == ast.KindSwitchStatement {
			// If switch has any branches, consider it branching
			return true
		}
	}

	return false
}

// allPathsTerminateEarly checks if all code paths terminate early (return/throw) without calling super
func allPathsTerminateEarly(body *ast.Node) bool {
	if body == nil || body.Kind != ast.KindBlock {
		return false
	}

	statements := body.Statements()
	if len(statements) == 0 {
		return false
	}

	// Check if all paths lead to early termination
	return statementTerminatesEarly(statements)
}

// statementTerminatesEarly checks if a statement or sequence of statements terminates early
func statementTerminatesEarly(statements []*ast.Node) bool {
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}

		// Direct termination
		if stmt.Kind == ast.KindReturnStatement || stmt.Kind == ast.KindThrowStatement {
			return true
		}

		// If-else where both branches terminate
		if stmt.Kind == ast.KindIfStatement {
			ifStmt := stmt.AsIfStatement()
			if ifStmt != nil && ifStmt.ElseStatement != nil {
				thenTerminates := branchTerminates(ifStmt.ThenStatement)
				elseTerminates := branchTerminates(ifStmt.ElseStatement)
				if thenTerminates && elseTerminates {
					return true
				}
			}
		}
	}

	return false
}

// branchTerminates checks if a branch (statement or block) terminates early
func branchTerminates(stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}

	if stmt.Kind == ast.KindReturnStatement || stmt.Kind == ast.KindThrowStatement {
		return true
	}

	if stmt.Kind == ast.KindBlock {
		return statementTerminatesEarly(stmt.Statements())
	}

	return false
}

// analyzeSuperCalls analyzes super() calls in a constructor body
type superCallAnalysis struct {
	hasSuperCall       bool     // true if any super() call exists
	allPathsHaveSuper  bool     // true if all paths call super()
	superCallLocations []*ast.Node // locations of all super() calls
}

// analyzeSuperCallsInBody analyzes super() calls in a constructor body
func analyzeSuperCallsInBody(body *ast.Node) superCallAnalysis {
	result := superCallAnalysis{
		superCallLocations: make([]*ast.Node, 0),
	}

	if body == nil {
		result.allPathsHaveSuper = false
		return result
	}

	// Find all super() calls
	findSuperCalls(body, &result.superCallLocations)
	result.hasSuperCall = len(result.superCallLocations) > 0

	// Analyze if all paths have super
	result.allPathsHaveSuper = checkAllPathsHaveSuper(body)

	return result
}

// findSuperCalls recursively finds all super() call expressions
func findSuperCalls(node *ast.Node, locations *[]*ast.Node) {
	if node == nil {
		return
	}

	// Check if this is a super() call
	if node.Kind == ast.KindCallExpression {
		expr := node.Expression()
		if expr != nil && expr.Kind == ast.KindSuperKeyword {
			// Store the call expression to get the correct position
			*locations = append(*locations, node)
		}
	}

	// Don't recurse into nested functions/classes
	if isFunctionOrClassNode(node) {
		return
	}

	// Recurse into children
	node.ForEachChild(func(child *ast.Node) bool {
		findSuperCalls(child, locations)
		return false // Continue iteration
	})
}

// isFunctionOrClassNode checks if a node is a function or class (boundary for super call search)
func isFunctionOrClassNode(node *ast.Node) bool {
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

// checkAllPathsHaveSuper checks if all code paths in the body call super()
func checkAllPathsHaveSuper(body *ast.Node) bool {
	if body == nil {
		return false
	}

	if body.Kind != ast.KindBlock {
		return false
	}

	statements := body.Statements()
	if len(statements) == 0 {
		return false
	}

	// Use a simplified control flow analysis
	return analyzeStatements(statements)
}

// analyzeStatements checks if super() is called in all code paths
func analyzeStatements(statements []*ast.Node) bool {
	for i, stmt := range statements {
		if stmt == nil {
			continue
		}

		// If we find a super() call at this level, all paths have it
		if hasSuperCall(stmt) {
			return true
		}

		// Check control flow statements
		switch stmt.Kind {
		case ast.KindIfStatement:
			// If-else with super/return/throw in all branches
			ifStmt := stmt.AsIfStatement()
			if ifStmt != nil {
				thenHasSuper := statementHasSuper(ifStmt.ThenStatement)
				elseStmt := ifStmt.ElseStatement

				if elseStmt != nil {
					// Has else clause - both branches must have super or terminate
					elseHasSuper := statementHasSuper(elseStmt)
					if thenHasSuper && elseHasSuper {
						// Both branches have super or terminate, so we can continue
						// Check remaining statements
						if i+1 < len(statements) {
							return analyzeStatements(statements[i+1:])
						}
						return true
					}
				}
			}

		case ast.KindSwitchStatement:
			// Switch with super in all cases (including default)
			if switchHasSuper(stmt) {
				// All switch paths have super, check remaining statements
				if i+1 < len(statements) {
					return analyzeStatements(statements[i+1:])
				}
				return true
			}

		case ast.KindReturnStatement, ast.KindThrowStatement:
			// Early return/throw means this path terminates without needing super
			// This is only valid if it's the only statement or all preceding paths also terminate
			return true

		case ast.KindExpressionStatement:
			// Check if this is a ternary expression with super in both branches
			expr := stmt.Expression()
			if expr != nil && expr.Kind == ast.KindConditionalExpression {
				condExpr := expr.AsConditionalExpression()
				if condExpr != nil {
					// Check if both branches have super
					whenTrueHasSuper := expressionHasSuper(condExpr.WhenTrue)
					whenFalseHasSuper := expressionHasSuper(condExpr.WhenFalse)
					if whenTrueHasSuper && whenFalseHasSuper {
						return true
					}
				}
			}
		}
	}

	return false
}

// expressionHasSuper checks if an expression contains a super() call
func expressionHasSuper(expr *ast.Node) bool {
	if expr == nil {
		return false
	}

	// Direct super() call
	if expr.Kind == ast.KindCallExpression {
		callExpr := expr.Expression()
		if callExpr != nil && callExpr.Kind == ast.KindSuperKeyword {
			return true
		}
	}

	// For conditional expressions, need super in both branches
	if expr.Kind == ast.KindConditionalExpression {
		condExpr := expr.AsConditionalExpression()
		if condExpr != nil {
			return expressionHasSuper(condExpr.WhenTrue) && expressionHasSuper(condExpr.WhenFalse)
		}
	}

	return false
}

// hasSuperCall checks if a statement contains a direct super() call
func hasSuperCall(stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}

	// Direct call
	if stmt.Kind == ast.KindExpressionStatement {
		expr := stmt.Expression()
		if expr != nil && expr.Kind == ast.KindCallExpression {
			callExpr := expr.Expression()
			if callExpr != nil && callExpr.Kind == ast.KindSuperKeyword {
				return true
			}
		}
	}

	return false
}

// statementHasSuper checks if a statement (or block) has super call or terminates
func statementHasSuper(stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}

	if stmt.Kind == ast.KindBlock {
		return analyzeStatements(stmt.Statements())
	}

	// Return and throw statements are valid path terminators
	if stmt.Kind == ast.KindReturnStatement || stmt.Kind == ast.KindThrowStatement {
		return true
	}

	return hasSuperCall(stmt)
}

// switchHasSuper checks if a switch statement has super in all branches
func switchHasSuper(switchStmt *ast.Node) bool {
	if switchStmt == nil || switchStmt.Kind != ast.KindSwitchStatement {
		return false
	}

	// Find the case block by traversing children
	var caseBlock *ast.Node
	switchStmt.ForEachChild(func(child *ast.Node) bool {
		if child != nil && child.Kind == ast.KindCaseBlock {
			caseBlock = child
			return true // Stop iteration
		}
		return false // Continue iteration
	})

	if caseBlock == nil {
		return false
	}

	// Get clauses from the case block
	var clauses []*ast.Node
	caseBlock.ForEachChild(func(child *ast.Node) bool {
		if child != nil && (child.Kind == ast.KindCaseClause || child.Kind == ast.KindDefaultClause) {
			clauses = append(clauses, child)
		}
		return false // Continue iteration
	})

	if len(clauses) == 0 {
		return false
	}

	hasDefault := false
	allClausesHaveSuper := true

	for _, clause := range clauses {
		if clause == nil {
			continue
		}

		if clause.Kind == ast.KindDefaultClause {
			hasDefault = true
		}

		// Check if this clause has super
		// Get statements from the clause using ForEachChild
		var statements []*ast.Node
		clause.ForEachChild(func(child *ast.Node) bool {
			statements = append(statements, child)
			return false // Continue iteration
		})
		if !analyzeStatements(statements) {
			allClausesHaveSuper = false
		}
	}

	// All cases must have super AND there must be a default case
	return hasDefault && allClausesHaveSuper
}

// ConstructorSuperRule enforces proper super() calls in constructors
var ConstructorSuperRule = rule.CreateRule(rule.Rule{
	Name: "constructor-super",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindConstructor: func(node *ast.Node) {
				// Check if this is a constructor
				if !isConstructor(node) {
					return
				}

				// Get the class this constructor belongs to
				classNode := getClassNode(node)
				if classNode == nil {
					return
				}

				// Check if the class extends something
				hasExtends := hasValidExtends(classNode)

				// Get the constructor body
				body := node.Body()

				// Analyze super() calls
				analysis := analyzeSuperCallsInBody(body)

				if hasExtends {
					// Derived class: must call super()
					if !analysis.hasSuperCall {
						// No super() call at all
						if allPathsTerminateEarly(body) {
							// All paths terminate early, no error
						} else if hasBranchingWithEarlyReturn(body) {
							// Some paths have early return/throw, report missingSome
							ctx.ReportNode(node, buildMissingSome())
						} else {
							// No branching or early returns, report missingAll
							ctx.ReportNode(node, buildMissingAll())
						}
					} else if !analysis.allPathsHaveSuper {
						// super() called in some paths but not all
						ctx.ReportNode(node, buildMissingSome())
					} else if analysis.allPathsHaveSuper && len(analysis.superCallLocations) > 1 {
						// Check for actual duplicates (super calls that can execute sequentially)
						duplicates := findDuplicateSuperCalls(body, analysis.superCallLocations)
						for _, superCall := range duplicates {
							// Report on the super keyword, not the call expression
							superKeyword := superCall.Expression()
							if superKeyword != nil && superKeyword.Kind == ast.KindSuperKeyword {
								ctx.ReportNode(superKeyword, buildDuplicate())
							} else {
								ctx.ReportNode(superCall, buildDuplicate())
							}
						}
					}
				} else {
					// Non-derived class or extends null/invalid: must NOT call super()
					if analysis.hasSuperCall {
						// Report each super() call as invalid
						for _, superCall := range analysis.superCallLocations {
							// Report on the super keyword, not the call expression
							superKeyword := superCall.Expression()
							if superKeyword != nil && superKeyword.Kind == ast.KindSuperKeyword {
								ctx.ReportNode(superKeyword, buildBadSuper())
							} else {
								ctx.ReportNode(superCall, buildBadSuper())
							}
						}
					}
				}
			},
		}
	},
})
