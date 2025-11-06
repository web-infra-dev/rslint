package array_callback_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options for array-callback-return rule
type Options struct {
	AllowImplicit bool `json:"allowImplicit"`
	CheckForEach  bool `json:"checkForEach"`
	AllowVoid     bool `json:"allowVoid"`
}

func parseOptions(options any) Options {
	opts := Options{
		AllowImplicit: false,
		CheckForEach:  false,
		AllowVoid:     false,
	}

	if options == nil {
		return opts
	}

	// Parse options with dual-format support (handles both array and object formats)
	var optsMap map[string]interface{}
	var ok bool

	// Handle array format: [{ option: value }]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, ok = optArray[0].(map[string]interface{})
	} else {
		// Handle direct object format: { option: value }
		optsMap, ok = options.(map[string]interface{})
	}

	if ok {
		if v, ok := optsMap["allowImplicit"].(bool); ok {
			opts.AllowImplicit = v
		}
		if v, ok := optsMap["checkForEach"].(bool); ok {
			opts.CheckForEach = v
		}
		if v, ok := optsMap["allowVoid"].(bool); ok {
			opts.AllowVoid = v
		}
	}
	return opts
}

// Message builders
func buildExpectedReturnValue(methodName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedReturnValue",
		Description: "Array.prototype." + methodName + "() expects a return value from arrow function.",
	}
}

func buildExpectedInside(methodName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedInside",
		Description: "Array.prototype." + methodName + "() expects a value to be returned at the end of arrow function.",
	}
}

func buildExpectedNoReturnValue(methodName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedNoReturnValue",
		Description: "Array.prototype." + methodName + "() expects no useless return value from arrow function.",
	}
}

func buildExpectedAtEnd(methodName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedAtEnd",
		Description: "Array.prototype." + methodName + "() expects a return value at the end of function.",
	}
}

// Target array methods that require return values
var targetMethods = map[string]bool{
	"every":         true,
	"filter":        true,
	"find":          true,
	"findIndex":     true,
	"findLast":      true,
	"findLastIndex": true,
	"flatMap":       true,
	"map":           true,
	"reduce":        true,
	"reduceRight":   true,
	"some":          true,
	"sort":          true,
	"toSorted":      true,
	// from is handled separately
}

// Methods that should not have return values
var forEachMethod = "forEach"

// isTargetMethod checks if the method is a target array method
func isTargetMethod(name string) bool {
	return targetMethods[name]
}

// isForEachMethod checks if the method is forEach
func isForEachMethod(name string) bool {
	return name == forEachMethod
}

// getMethodName extracts the method name from a CallExpression
func getMethodName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindCallExpression {
		return ""
	}

	expr := node.Expression()
	if expr == nil {
		return ""
	}

	// Handle PropertyAccessExpression (e.g., arr.map)
	if expr.Kind == ast.KindPropertyAccessExpression {
		name := expr.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.Text()
		}
	}

	return ""
}

// isArrayFromCall checks if this is Array.from()
func isArrayFromCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}

	expr := node.Expression()
	if expr == nil || expr.Kind != ast.KindPropertyAccessExpression {
		return false
	}

	obj := expr.Expression()
	if obj == nil || obj.Kind != ast.KindIdentifier || obj.Text() != "Array" {
		return false
	}

	name := expr.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}

	return name.Text() == "from"
}

// isFunctionNode checks if a node is a function
func isFunctionNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindFunctionDeclaration ||
		node.Kind == ast.KindFunctionExpression ||
		node.Kind == ast.KindArrowFunction ||
		node.Kind == ast.KindMethodDeclaration
}

// isGeneratorOrAsyncFunction checks if a node is a generator or async function
func isGeneratorOrAsyncFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check if it has async modifier
	modifiers := node.Modifiers()
	if modifiers != nil {
		for _, mod := range modifiers.Nodes {
			if mod != nil && mod.Kind == ast.KindAsyncKeyword {
				return true
			}
		}
	}

	// Check for generator function (function*)
	// Generators contain yield expressions
	// We check recursively through the body
	if containsYield(node.Body()) {
		return true
	}

	return false
}

// containsYield recursively checks if a node contains a yield expression
func containsYield(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindYieldExpression {
		return true
	}
	// Recursively check children
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if containsYield(child) {
			found = true
			return true // Stop iteration
		}
		return false // Continue iteration
	})
	return found
}

// isVoidExpression checks if an expression is a void expression
func isVoidExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindVoidExpression
}

// checkCallbackReturn validates that a callback has proper return statements
func checkCallbackReturn(ctx rule.RuleContext, funcNode *ast.Node, methodName string, opts Options, checkForEach bool) {
	if funcNode == nil {
		return
	}

	// Skip generator and async functions - they have different control flow semantics
	if isGeneratorOrAsyncFunction(funcNode) {
		return
	}

	body := funcNode.Body()
	if body == nil {
		return
	}

	// For arrow functions with expression bodies, check if it's an implicit return
	if funcNode.Kind == ast.KindArrowFunction {
		// Expression body (no braces) - always returns the expression value
		if body.Kind != ast.KindBlock {
			if checkForEach {
				// forEach shouldn't return values, but expression bodies always do
				// unless it's a void expression
				if opts.AllowVoid && isVoidExpression(body) {
					return
				}
				ctx.ReportNode(funcNode, buildExpectedNoReturnValue(methodName))
			}
			// For other methods, expression bodies always return, which is good
			return
		}
	}

	// Analyze return statements
	result := analyzeCallbackReturns(body, opts.AllowImplicit)

	if checkForEach {
		// forEach callbacks should not return values
		if result.hasReturnWithValue {
			ctx.ReportNode(funcNode, buildExpectedNoReturnValue(methodName))
		}
	} else {
		// Other methods require return values
		if result.hasNoReturns {
			// No return statements at all, or only empty returns
			if funcNode.Kind == ast.KindArrowFunction {
				ctx.ReportNode(funcNode, buildExpectedReturnValue(methodName))
			} else {
				ctx.ReportNode(funcNode, buildExpectedAtEnd(methodName))
			}
		} else if !result.allPathsReturn {
			// Some paths return a value, but not all paths do
			if funcNode.Kind == ast.KindArrowFunction {
				ctx.ReportNode(funcNode, buildExpectedInside(methodName))
			} else {
				ctx.ReportNode(funcNode, buildExpectedAtEnd(methodName))
			}
		}
	}
}

// callbackReturnResult holds the result of callback return analysis
type callbackReturnResult struct {
	hasNoReturns       bool // true if there are no return statements with values
	allPathsReturn     bool // true if all code paths return a value
	hasReturnWithValue bool // true if there's at least one return with value
}

// analyzeCallbackReturns performs control flow analysis on a callback body
func analyzeCallbackReturns(body *ast.Node, allowImplicit bool) callbackReturnResult {
	if body == nil {
		return callbackReturnResult{hasNoReturns: true, allPathsReturn: false, hasReturnWithValue: false}
	}

	hasReturnWithValue := false
	hasReturnWithoutValue := false

	// Use ForEachReturnStatement to find all return statements
	ast.ForEachReturnStatement(body, func(stmt *ast.Node) bool {
		expr := stmt.Expression()
		if expr != nil {
			hasReturnWithValue = true
		} else {
			hasReturnWithoutValue = true
		}
		return false // Continue iterating
	})

	// Track if we found at least one return with value
	result := callbackReturnResult{
		hasReturnWithValue: hasReturnWithValue,
	}

	// If allowImplicit is true, empty returns are acceptable
	if allowImplicit && hasReturnWithoutValue && !hasReturnWithValue {
		return callbackReturnResult{
			hasNoReturns:       false,
			allPathsReturn:     true,
			hasReturnWithValue: false,
		}
	}

	// Determine if this is a simple case or complex case
	isSingleReturn := false
	if body.Kind == ast.KindBlock {
		statements := body.Statements()
		if len(statements) == 1 && statements[0].Kind == ast.KindReturnStatement {
			isSingleReturn = true
		}
	}

	// If we have no return with value at all (and empty returns don't count unless allowImplicit)
	if !hasReturnWithValue && (!allowImplicit || !hasReturnWithoutValue) {
		return callbackReturnResult{
			hasNoReturns:       true,
			allPathsReturn:     false,
			hasReturnWithValue: false,
		}
	}

	// Heuristic for determining if all paths return:
	// This is a simplified heuristic that works for common cases
	countReturnsWithValue := 0
	ast.ForEachReturnStatement(body, func(stmt *ast.Node) bool {
		if stmt.Expression() != nil {
			countReturnsWithValue++
		}
		return false
	})

	// For try-catch blocks, if the try block has a return, we consider it valid
	// even if the catch doesn't (since the catch only runs on exception)
	hasTryStatement := false
	if body.Kind == ast.KindBlock {
		statements := body.Statements()
		for _, stmt := range statements {
			if stmt != nil {
				if stmt.Kind == ast.KindTryStatement {
					hasTryStatement = true
				}
			}
		}
	}

	// Check if all paths return based on the structure
	// We use several heuristics:
	// 1. Single return statement (obviously all paths return)
	// 2. Simple body with no empty returns and at least one return
	// 3. Try-catch with a return in the try block
	// 4. If-else statements where both branches return
	hasIfElseWithReturns := checkIfElseReturns(body)

	// Note: We intentionally don't try to detect all control flow patterns perfectly
	// as this requires proper control flow analysis which is complex
	allPathsReturn := isSingleReturn ||
		(!hasReturnWithoutValue && isSimpleBody(body) && hasReturnWithValue) ||
		(hasTryStatement && hasReturnWithValue) ||
		hasIfElseWithReturns

	result.hasNoReturns = false
	result.allPathsReturn = allPathsReturn

	return result
}

// isSimpleBody checks if a function body is simple enough that we can assume all paths return
func isSimpleBody(body *ast.Node) bool {
	if body == nil || body.Kind != ast.KindBlock {
		return true
	}

	statements := body.Statements()
	if len(statements) == 0 {
		return true
	}

	// Check for control flow statements (if, switch, loops, etc.)
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindIfStatement, ast.KindSwitchStatement,
			ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindWhileStatement, ast.KindDoStatement,
			ast.KindTryStatement:
			// Has control flow - not simple
			return false
		}
	}

	return true
}

// checkIfElseReturns checks if an if-else statement has returns in all branches
// This properly analyzes if-else chains to ensure all branches return
func checkIfElseReturns(body *ast.Node) bool {
	if body == nil || body.Kind != ast.KindBlock {
		return false
	}

	statements := body.Statements()
	if len(statements) == 0 {
		return false
	}

	// Check if the function body consists of a single if-else chain that covers all paths
	// For example:
	// if (a) { return x; } else { return y; }  -> all paths return
	// if (a) { return x; } else if (b) { return y; }  -> NOT all paths (missing final else)
	// if (a) { return x; } else if (b) { return y; } else { return z; }  -> all paths return

	for _, stmt := range statements {
		if stmt == nil {
			continue
		}
		if stmt.Kind == ast.KindIfStatement {
			if ifStatementCoversAllPaths(stmt) {
				return true
			}
		}
	}

	return false
}

// ifStatementCoversAllPaths checks if an if-statement covers all code paths with returns
func ifStatementCoversAllPaths(ifStmt *ast.Node) bool {
	if ifStmt == nil || ifStmt.Kind != ast.KindIfStatement {
		return false
	}

	// Access the IfStatement structure
	ifStmtData := ifStmt.AsIfStatement()
	if ifStmtData == nil {
		return false
	}

	// Get the then statement
	thenStmt := ifStmtData.ThenStatement
	if thenStmt == nil || !blockReturnsValue(thenStmt) {
		return false
	}

	// Get the else statement
	elseStmt := ifStmtData.ElseStatement
	if elseStmt == nil {
		// No else clause - doesn't cover all paths
		return false
	}

	// If the else statement is another if-statement (else if), check it recursively
	if elseStmt.Kind == ast.KindIfStatement {
		return ifStatementCoversAllPaths(elseStmt)
	}

	// Otherwise, check if the else block returns a value
	return blockReturnsValue(elseStmt)
}

// blockReturnsValue checks if a block/statement returns a value
func blockReturnsValue(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Direct return statement
	if node.Kind == ast.KindReturnStatement {
		return node.Expression() != nil
	}

	// Block statement - check if it ends with a return
	if node.Kind == ast.KindBlock {
		statements := node.Statements()
		if len(statements) == 0 {
			return false
		}
		// Check the last statement
		lastStmt := statements[len(statements)-1]
		if lastStmt != nil && lastStmt.Kind == ast.KindReturnStatement {
			return lastStmt.Expression() != nil
		}
	}

	return false
}

// ArrayCallbackReturnRule enforces return statements in callbacks of array methods
var ArrayCallbackReturnRule = rule.CreateRule(rule.Rule{
	Name: "array-callback-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				methodName := getMethodName(node)
				isArrayFrom := false

				// Check if it's Array.from (needs special handling)
				if isArrayFromCall(node) {
					methodName = "from"
					isArrayFrom = true
				}

				// Check if we have a method name
				if methodName == "" {
					return
				}

				// Check if it's a target method
				isTarget := isTargetMethod(methodName)
				isForEach := opts.CheckForEach && isForEachMethod(methodName)

				if !isTarget && !isForEach && !isArrayFrom {
					return
				}

				// Get the callback argument
				args := node.Arguments()
				if len(args) == 0 {
					return
				}

				var callbackArg *ast.Node
				if isArrayFrom {
					// Array.from(arr, callback) - callback is second argument
					if len(args) >= 2 {
						callbackArg = args[1]
					}
				} else {
					// For other methods, callback is first argument
					callbackArg = args[0]
				}

				if callbackArg == nil {
					return
				}

				// Check if the argument is a function
				if !isFunctionNode(callbackArg) {
					return
				}

				// Check the callback
				checkCallbackReturn(ctx, callbackArg, methodName, opts, isForEach)
			},
		}
	},
})
