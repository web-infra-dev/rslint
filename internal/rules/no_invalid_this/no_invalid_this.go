package no_invalid_this

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoInvalidThisOptions struct {
	CapIsConstructor bool `json:"capIsConstructor"`
}

type contextTracker struct {
	stack []bool
}

func (ct *contextTracker) pushValid() {
	ct.stack = append(ct.stack, true)
}

func (ct *contextTracker) pushInvalid() {
	ct.stack = append(ct.stack, false)
}

func (ct *contextTracker) pop() {
	if len(ct.stack) > 0 {
		ct.stack = ct.stack[:len(ct.stack)-1]
	}
}

func (ct *contextTracker) isCurrentValid() bool {
	if len(ct.stack) == 0 {
		return false
	}
	return ct.stack[len(ct.stack)-1]
}

// Check if a function has a 'this' parameter (TypeScript feature)
func hasThisParameter(node *ast.Node) bool {
	var params []*ast.Node

	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl.Parameters != nil {
			params = funcDecl.Parameters.Nodes
		}
	case ast.KindFunctionExpression:
		funcExpr := node.AsFunctionExpression()
		if funcExpr.Parameters != nil {
			params = funcExpr.Parameters.Nodes
		}
	case ast.KindArrowFunction:
		arrow := node.AsArrowFunction()
		if arrow.Parameters != nil {
			params = arrow.Parameters.Nodes
		}
	default:
		return false
	}

	// Check all parameters for TypeScript 'this' parameter
	for _, param := range params {
		if param.Kind == ast.KindParameter {
			paramDecl := param.AsParameterDeclaration()
			if paramDecl.Name() != nil && ast.IsIdentifier(paramDecl.Name()) {
				if paramDecl.Name().AsIdentifier().Text == "this" {
					return true
				}
			}
		}
	}

	return false
}

// Check if a function is a constructor based on naming convention
func isConstructor(node *ast.Node, capIsConstructor bool) bool {
	if !capIsConstructor {
		return false
	}

	var name string

	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		nameNode := funcDecl.Name()
		if nameNode != nil && ast.IsIdentifier(nameNode) {
			name = nameNode.AsIdentifier().Text
		}
	case ast.KindFunctionExpression:
		funcExpr := node.AsFunctionExpression()
		nameNode := funcExpr.Name()
		if nameNode != nil && ast.IsIdentifier(nameNode) {
			name = nameNode.AsIdentifier().Text
		}
	}

	if name != "" && len(name) > 0 {
		// Check if first character is uppercase
		return strings.ToUpper(name[:1]) == name[:1]
	}

	// Check if this is being assigned to a capitalized variable
	if node.Parent != nil {
		switch node.Parent.Kind {
		case ast.KindVariableDeclaration:
			varDecl := node.Parent.AsVariableDeclaration()
			// Variable declarations in no_invalid_this are VariableDeclaration, not VariableStatement
			nameNode := varDecl.Name()
			if nameNode != nil && ast.IsIdentifier(nameNode) {
				name = nameNode.AsIdentifier().Text
				if len(name) > 0 {
					return strings.ToUpper(name[:1]) == name[:1]
				}
			}
		case ast.KindBinaryExpression:
			binExpr := node.Parent.AsBinaryExpression()
			if binExpr.OperatorToken.Kind == ast.KindEqualsToken {
				if ast.IsIdentifier(binExpr.Left) {
					name = binExpr.Left.AsIdentifier().Text
					if len(name) > 0 {
						return strings.ToUpper(name[:1]) == name[:1]
					}
				}
			}
		case ast.KindParameter:
			// Check if this function is a default value for a parameter
			param := node.Parent.AsParameterDeclaration()
			if param.Name() != nil && ast.IsIdentifier(param.Name()) {
				name = param.Name().AsIdentifier().Text
				if len(name) > 0 {
					return strings.ToUpper(name[:1]) == name[:1]
				}
			}
		}
	}

	return false
}

// Check if node has @this JSDoc tag
func hasThisJSDocTag(node *ast.Node, sourceFile *ast.SourceFile) bool {
	text := string(sourceFile.Text())
	nodeStart := int(node.Pos())

	// For function expressions, check different patterns
	if node.Kind == ast.KindFunctionExpression {
		// Pattern 1: foo(/* @this Obj */ function () {})
		// Check if there's a @this comment immediately before this function expression
		// First check within the node itself (comment may be included in node range)
		nodeEnd := int(node.End())
		nodeText := text[nodeStart:nodeEnd]

		// Find the function keyword position within the node
		funcKeywordPos := strings.Index(nodeText, "function")
		if funcKeywordPos != -1 {
			// Check the text before the function keyword for @this comment
			beforeFuncText := nodeText[:funcKeywordPos]
			if strings.Contains(beforeFuncText, "@this") {
				// Find the last @this before the function keyword
				lastThisIdx := strings.LastIndex(beforeFuncText, "@this")
				if lastThisIdx != -1 {
					beforeThis := beforeFuncText[:lastThisIdx]
					afterThis := beforeFuncText[lastThisIdx:]

					// Check if it's in a block comment /* @this ... */
					blockStart := strings.LastIndex(beforeThis, "/*")
					if blockStart != -1 {
						blockEnd := strings.Index(afterThis, "*/")
						if blockEnd != -1 {
							// Make sure there's no intervening */ between /* and @this
							noEndBetween := strings.Index(beforeThis[blockStart:], "*/") == -1
							if noEndBetween {
								// Check that the comment is immediately before the function
								// There should be only whitespace between comment end and function keyword
								commentEndPos := lastThisIdx + blockEnd + 2
								remainingText := strings.TrimSpace(beforeFuncText[commentEndPos:])
								// Only allow whitespace
								if len(remainingText) == 0 {
									return true
								}
							}
						}
					}
				}
			}
		}

		// Also check before the node (original logic)
		searchStart := max(0, nodeStart-200)
		searchText := text[searchStart:nodeStart]

		// Look for @this that's immediately before the function keyword
		if strings.Contains(searchText, "@this") {
			// Find the last @this before the function
			lastThisIdx := strings.LastIndex(searchText, "@this")
			if lastThisIdx != -1 {
				beforeThis := searchText[:lastThisIdx]
				afterThis := searchText[lastThisIdx:]

				// Check if it's in a block comment /* @this ... */
				blockStart := strings.LastIndex(beforeThis, "/*")
				if blockStart != -1 {
					blockEnd := strings.Index(afterThis, "*/")
					if blockEnd != -1 {
						// Make sure there's no intervening */ between /* and @this
						noEndBetween := strings.Index(beforeThis[blockStart:], "*/") == -1
						if noEndBetween {
							// Check that the comment is immediately before the function
							// There should be only whitespace between comment end and function keyword
							commentEndPos := lastThisIdx + blockEnd + 2
							remainingText := strings.TrimSpace(searchText[commentEndPos:])
							// Only allow whitespace - the function keyword will be after the node start
							if len(remainingText) == 0 {
								return true
							}
						}
					}
				}
			}
		}

		// Pattern 2: return /** @this Obj */ function bar() {}
		// Check if we're in a return statement context with @this comment
		if node.Parent != nil && node.Parent.Kind == ast.KindReturnStatement {
			returnStmtStart := int(node.Parent.Pos())
			returnStmtEnd := int(node.Parent.End())
			returnText := text[returnStmtStart:returnStmtEnd]

			// Check if there's @this comment in the return statement before the function
			if strings.Contains(returnText, "@this") {
				thisIdx := strings.Index(returnText, "@this")
				funcIdx := strings.Index(returnText, "function")
				if thisIdx != -1 && funcIdx != -1 && thisIdx < funcIdx {
					// Check if @this is in a comment
					beforeThis := returnText[:thisIdx]
					afterThis := returnText[thisIdx:]

					// Check for JSDoc comment /** @this ... */
					jsdocStart := strings.LastIndex(beforeThis, "/**")
					if jsdocStart != -1 {
						jsdocEnd := strings.Index(afterThis, "*/")
						if jsdocEnd != -1 {
							noEndBetween := strings.Index(beforeThis[jsdocStart:], "*/") == -1
							if noEndBetween {
								return true
							}
						}
					}

					// Check for block comment /* @this ... */
					blockStart := strings.LastIndex(beforeThis, "/*")
					if blockStart != -1 {
						blockEnd := strings.Index(afterThis, "*/")
						if blockEnd != -1 {
							noEndBetween := strings.Index(beforeThis[blockStart:], "*/") == -1
							if noEndBetween {
								return true
							}
						}
					}
				}
			}
		}

		return false
	}

	// For function declarations, check for JSDoc comments before the function
	if node.Kind == ast.KindFunctionDeclaration {
		// The function node may include leading JSDoc comments, so check the entire node range
		nodeEnd := int(node.End())
		nodeText := text[nodeStart:nodeEnd]

		// Find the position of the actual "function" keyword within the node
		funcKeywordPos := strings.Index(nodeText, "function")
		if funcKeywordPos == -1 {
			return false
		}

		// Search the text before the function keyword for @this
		searchText := nodeText[:funcKeywordPos]
		if strings.Contains(searchText, "@this") {
			// Find all @this occurrences
			thisIndices := []int{}
			searchIndex := 0
			for {
				idx := strings.Index(searchText[searchIndex:], "@this")
				if idx == -1 {
					break
				}
				thisIndices = append(thisIndices, searchIndex+idx)
				searchIndex += idx + 5
			}

			// Check each @this occurrence (starting from the last one)
			for i := len(thisIndices) - 1; i >= 0; i-- {
				thisIdx := thisIndices[i]
				beforeThis := searchText[:thisIdx]
				afterThis := searchText[thisIdx:]

				// Check for JSDoc comment /** ... @this ... */
				jsdocStart := strings.LastIndex(beforeThis, "/**")
				if jsdocStart != -1 {
					jsdocEnd := strings.Index(afterThis, "*/")
					if jsdocEnd != -1 {
						// Make sure there's no comment end between JSDoc start and @this
						noEndBetween := strings.Index(beforeThis[jsdocStart:], "*/") == -1
						if noEndBetween {
							// Ensure the JSDoc comment is immediately before the function
							commentEndPos := thisIdx + jsdocEnd + 2
							// Allow some whitespace but no other code
							remainingText := strings.TrimSpace(searchText[commentEndPos:])
							if len(remainingText) == 0 || strings.HasPrefix(remainingText, "function") {
								return true
							}
						}
					}
				}

				// Check for multiline JSDoc with * @this pattern
				lineStartIdx := strings.LastIndex(beforeThis, "\n")
				if lineStartIdx != -1 {
					lineContent := strings.TrimSpace(beforeThis[lineStartIdx+1:] + "@this")
					if strings.HasPrefix(lineContent, "*") && strings.Contains(lineContent, "@this") {
						// Make sure we're still inside a JSDoc comment
						lastJSDocStart := strings.LastIndex(beforeThis[:lineStartIdx], "/**")
						lastJSDocEnd := strings.LastIndex(beforeThis[:lineStartIdx], "*/")
						if lastJSDocStart != -1 && (lastJSDocEnd == -1 || lastJSDocStart > lastJSDocEnd) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Check if function is an argument to a function call
func isFunctionArgument(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// If parent is a call expression, this function is likely an argument
	if parent.Kind == ast.KindCallExpression {
		return true
	}

	// Check if parent is an argument list, which means we're in a call
	current := parent
	for current != nil {
		if current.Kind == ast.KindCallExpression {
			return true
		}
		current = current.Parent
	}

	return false
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// isNullOrUndefined checks if node represents null, undefined, or void expression
func isNullOrUndefined(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindNullKeyword:
		return true
	case ast.KindUndefinedKeyword:
		return true
	case ast.KindVoidExpression:
		return true
	}

	if ast.IsIdentifier(node) {
		text := node.AsIdentifier().Text
		return text == "undefined" || text == "null"
	}

	return false
}

// Check if a function is assigned as a method or has valid this binding
func isValidMethodContext(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindPropertyAssignment:
		// Direct object method assignment
		return true
	case ast.KindBinaryExpression:
		// Check logical operators like obj.foo = bar || function() {}
		binExpr := parent.AsBinaryExpression()
		if binExpr.OperatorToken.Kind == ast.KindBarBarToken || binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
			// Check if this binary expression is part of a property assignment
			grandParent := parent.Parent
			if grandParent != nil && grandParent.Kind == ast.KindPropertyAssignment {
				return true
			}
			// Check if this binary expression is part of a binary assignment
			if grandParent != nil && grandParent.Kind == ast.KindBinaryExpression {
				return isValidAssignmentContext(grandParent, parent)
			}
		}
		// Check various assignment patterns
		return isValidAssignmentContext(parent, node)
	case ast.KindMethodDeclaration:
		// Class method
		return true
	case ast.KindPropertyDeclaration:
		// Class property with function value
		return true
	case ast.KindConditionalExpression:
		// Check ternary expressions (e.g., foo ? bar : function() {})
		return isValidConditionalContext(parent, node)
	case ast.KindCallExpression:
		// Check function passed to certain methods
		return isValidCallContext(parent, node)
	}

	// Check if nested in object definition patterns or other valid contexts
	return false
}

// Check if the assignment is to a property (obj.foo = function() {})
func isValidAssignmentContext(parent *ast.Node, funcNode *ast.Node) bool {
	binExpr := parent.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != ast.KindEqualsToken {
		return false
	}

	// Direct property assignment (obj.foo = function)
	if ast.IsPropertyAccessExpression(binExpr.Left) {
		return true
	}

	// Check if the function is the direct assignment value
	if binExpr.Right == funcNode {
		return ast.IsPropertyAccessExpression(binExpr.Left)
	}

	// Check if in conditional/logical assignment
	return isNestedInValidAssignment(binExpr.Right, funcNode)
}

// Check if function is nested in a valid assignment pattern
func isNestedInValidAssignment(node *ast.Node, target *ast.Node) bool {
	if node == target {
		return true
	}

	switch node.Kind {
	case ast.KindConditionalExpression:
		cond := node.AsConditionalExpression()
		return cond.WhenTrue == target || cond.WhenFalse == target ||
			isNestedInValidAssignment(cond.WhenTrue, target) ||
			isNestedInValidAssignment(cond.WhenFalse, target)
	case ast.KindBinaryExpression:
		bin := node.AsBinaryExpression()
		if bin.OperatorToken.Kind == ast.KindBarBarToken || bin.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
			return bin.Left == target || bin.Right == target ||
				isNestedInValidAssignment(bin.Left, target) ||
				isNestedInValidAssignment(bin.Right, target)
		}
	case ast.KindCallExpression:
		// Check IIFE patterns: (function() { return function() {}; })()
		call := node.AsCallExpression()
		if ast.IsFunctionExpression(call.Expression) {
			return containsTargetFunction(call.Expression, target)
		}
	case ast.KindArrowFunction:
		// Arrow function returning a function: (() => function() {})()
		arrow := node.AsArrowFunction()
		return arrow.Body == target || isNestedInValidAssignment(arrow.Body, target)
	}

	return false
}

// Check if target function is contained within another function
func containsTargetFunction(container *ast.Node, target *ast.Node) bool {
	if container == target {
		return true
	}

	if ast.IsFunctionExpression(container) {
		funcExpr := container.AsFunctionExpression()
		if funcExpr.Body != nil {
			return findFunctionInBody(funcExpr.Body, target)
		}
	}

	return false
}

// Recursively find target function in function body
func findFunctionInBody(body *ast.Node, target *ast.Node) bool {
	if body == target {
		return true
	}

	if body.Kind == ast.KindBlock {
		block := body.AsBlock()
		for _, stmt := range block.Statements.Nodes {
			if stmt.Kind == ast.KindReturnStatement {
				ret := stmt.AsReturnStatement()
				if ret.Expression == target {
					return true
				}
			}
		}
	}

	return false
}

// Check conditional expressions (ternary operator)
func isValidConditionalContext(parent *ast.Node, funcNode *ast.Node) bool {
	// If in a conditional, check if the parent context is valid
	grandParent := parent.Parent
	if grandParent == nil {
		return false
	}

	switch grandParent.Kind {
	case ast.KindBinaryExpression:
		return isValidAssignmentContext(grandParent, parent)
	case ast.KindPropertyAssignment:
		// conditional in object property: {foo: bar ? func1 : func2}
		return true
	case ast.KindConditionalExpression:
		// nested conditional: a ? (b ? func1 : func2) : func3
		return isValidConditionalContext(grandParent, parent)
	}

	return false
}

// Check if function is in Object.defineProperty or similar contexts
func isInDefinePropertyContext(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if ast.IsCallExpression(current) {
			callExpr := current.AsCallExpression()
			if ast.IsPropertyAccessExpression(callExpr.Expression) {
				propAccess := callExpr.Expression.AsPropertyAccessExpression()
				nameNode := propAccess.Name()
				if ast.IsIdentifier(propAccess.Expression) && ast.IsIdentifier(nameNode) {
					objName := propAccess.Expression.AsIdentifier().Text
					methodName := nameNode.AsIdentifier().Text
					if objName == "Object" && (methodName == "defineProperty" || methodName == "defineProperties") {
						return true
					}
				}
			}
			if ast.IsIdentifier(callExpr.Expression) {
				methodName := callExpr.Expression.AsIdentifier().Text
				if methodName == "defineProperty" || methodName == "defineProperties" {
					return true
				}
			}
		}
		current = current.Parent
	}
	return false
}

// Check if function is in object literal context
func isInObjectLiteralContext(node *ast.Node) bool {
	current := node.Parent
	for current != nil && current.Kind != ast.KindFunctionExpression && current.Kind != ast.KindFunctionDeclaration {
		if current.Kind == ast.KindObjectLiteralExpression {
			return true
		}
		current = current.Parent
	}
	return false
}

// Check call contexts (bind, call, apply, array methods)
func isValidCallContext(parent *ast.Node, funcNode *ast.Node) bool {
	callExpr := parent.AsCallExpression()
	args := callExpr.Arguments.Nodes

	// Find the position of the function in arguments
	funcArgIndex := -1
	for i, arg := range args {
		if arg == funcNode {
			funcArgIndex = i
			break
		}
	}

	if funcArgIndex == -1 {
		return false
	}

	// Only consider property access expressions (like obj.method())
	// Plain identifier calls like foo() should not be valid contexts
	if ast.IsPropertyAccessExpression(callExpr.Expression) {
		propAccess := callExpr.Expression.AsPropertyAccessExpression()
		nameNode := propAccess.Name()
		if ast.IsIdentifier(nameNode) {
			methodName := nameNode.AsIdentifier().Text
			switch methodName {
			case "bind", "call", "apply":
				// These methods make 'this' valid only if thisArg is not null/undefined
				if len(args) > 0 {
					return !isNullOrUndefined(args[0])
				}
				return false
			case "every", "filter", "find", "findIndex", "forEach", "map", "some":
				// Array methods: function at index 0, optional thisArg at index 1
				if funcArgIndex == 0 {
					if len(args) > 1 {
						return !isNullOrUndefined(args[1])
					}
					return false // No thisArg provided
				}
			}
		}
	}

	// Check Array.from(iterable, mapFn, thisArg)
	if ast.IsPropertyAccessExpression(callExpr.Expression) {
		propAccess := callExpr.Expression.AsPropertyAccessExpression()
		nameNode := propAccess.Name()
		if ast.IsIdentifier(propAccess.Expression) && ast.IsIdentifier(nameNode) {
			if propAccess.Expression.AsIdentifier().Text == "Array" &&
				nameNode.AsIdentifier().Text == "from" {
				if funcArgIndex == 1 && len(args) > 2 {
					return !isNullOrUndefined(args[2])
				}
				if funcArgIndex == 1 && len(args) <= 2 {
					return false // No thisArg provided
				}
			}
		}
	}

	// Check Reflect.apply(target, thisArgument, argumentsList)
	if ast.IsPropertyAccessExpression(callExpr.Expression) {
		propAccess := callExpr.Expression.AsPropertyAccessExpression()
		nameNode := propAccess.Name()
		if ast.IsIdentifier(propAccess.Expression) && ast.IsIdentifier(nameNode) {
			if propAccess.Expression.AsIdentifier().Text == "Reflect" &&
				nameNode.AsIdentifier().Text == "apply" {
				if funcArgIndex == 0 && len(args) > 1 {
					return !isNullOrUndefined(args[1])
				}
			}
		}
	}

	return false
}

// Check if we're in a class context
func isInClassContext(node *ast.Node) bool {
	current := node
	for current != nil {
		switch current.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return true
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
			// Stop at function boundaries
			return false
		}
		current = current.Parent
	}
	return false
}

// Check if function is bound with call/apply/bind
func isInFunctionBinding(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		if parent.Kind == ast.KindCallExpression {
			callExpr := parent.AsCallExpression()
			if ast.IsPropertyAccessExpression(callExpr.Expression) {
				propAccess := callExpr.Expression.AsPropertyAccessExpression()
				// Check if our function is the target of the method call
				// Handle parenthesized expressions
				targetNode := propAccess.Expression
				for targetNode != nil && targetNode.Kind == ast.KindParenthesizedExpression {
					targetNode = targetNode.AsParenthesizedExpression().Expression
				}

				if targetNode == node {
					nameNode := propAccess.Name()
					if ast.IsIdentifier(nameNode) {
						methodName := nameNode.AsIdentifier().Text
						if methodName == "call" || methodName == "apply" || methodName == "bind" {
							// Check if thisArg is provided and not null/undefined
							args := callExpr.Arguments.Nodes
							if len(args) > 0 {
								return !isNullOrUndefined(args[0])
							}
						}
					}
				}
			}
		}
		parent = parent.Parent
	}
	return false
}

// Check if function is returned from an IIFE that's in a valid context
func isReturnedFromIIFE(node *ast.Node) bool {
	// Two cases to handle:
	// 1. Explicit return statement: return function() {}
	// 2. Arrow function implicit return: () => function() {}

	parent := node.Parent
	var containingFunc *ast.Node

	if parent != nil && parent.Kind == ast.KindReturnStatement {
		// Case 1: Explicit return statement
		// Walk up to find the containing function
		current := parent.Parent
		for current != nil {
			if current.Kind == ast.KindFunctionExpression || current.Kind == ast.KindArrowFunction {
				containingFunc = current
				break
			}
			// Stop at other function boundaries
			if current.Kind == ast.KindFunctionDeclaration {
				return false
			}
			current = current.Parent
		}
	} else if parent != nil && parent.Kind == ast.KindArrowFunction {
		// Case 2: Arrow function implicit return
		arrow := parent.AsArrowFunction()
		// Check if this node is the body of the arrow function (implicit return)
		if arrow.Body == node {
			containingFunc = parent
		}
	}

	if containingFunc == nil {
		return false
	}

	// Check if the containing function is an IIFE
	funcParent := containingFunc.Parent
	// Handle direct call without parentheses (for arrow functions)
	if funcParent != nil && funcParent.Kind == ast.KindCallExpression {
		callExpr := funcParent.AsCallExpression()
		if callExpr.Expression == containingFunc {
			// Check if the IIFE result is in a valid context
			iifeParent := funcParent.Parent
			if iifeParent != nil {
				switch iifeParent.Kind {
				case ast.KindPropertyAssignment:
					return true
				case ast.KindBinaryExpression:
					// Check if assigned to an object property
					binExpr := iifeParent.AsBinaryExpression()
					if binExpr.OperatorToken.Kind == ast.KindEqualsToken && ast.IsPropertyAccessExpression(binExpr.Left) {
						return true
					}
				}
			}
		}
	}
	// Handle parenthesized function expressions
	if funcParent != nil && funcParent.Kind == ast.KindParenthesizedExpression {
		parenParent := funcParent.Parent
		if parenParent != nil && parenParent.Kind == ast.KindCallExpression {
			callExpr := parenParent.AsCallExpression()
			// Check if it's being called immediately with no/few arguments (typical IIFE)
			if callExpr.Expression == funcParent {
				// Check if the IIFE result is in a valid context
				iifeParent := parenParent.Parent
				if iifeParent != nil {
					switch iifeParent.Kind {
					case ast.KindPropertyAssignment:
						return true
					case ast.KindBinaryExpression:
						// Check if assigned to an object property
						binExpr := iifeParent.AsBinaryExpression()
						if binExpr.OperatorToken.Kind == ast.KindEqualsToken && ast.IsPropertyAccessExpression(binExpr.Left) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

var NoInvalidThisRule = rule.Rule{
	Name: "no-invalid-this",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoInvalidThisOptions{
			CapIsConstructor: true,
		}
		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
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
				if capIsConstructor, ok := optsMap["capIsConstructor"].(bool); ok {
					opts.CapIsConstructor = capIsConstructor
				}
			}
		}

		tracker := &contextTracker{
			stack: []bool{false}, // Start with global scope (invalid)
		}

		return rule.RuleListeners{
			// Class contexts
			ast.KindClassDeclaration: func(node *ast.Node) {
				tracker.pushValid()
			},
			rule.ListenerOnExit(ast.KindClassDeclaration): func(node *ast.Node) {
				tracker.pop()
			},
			ast.KindClassExpression: func(node *ast.Node) {
				tracker.pushValid()
			},
			rule.ListenerOnExit(ast.KindClassExpression): func(node *ast.Node) {
				tracker.pop()
			},

			// Property definitions (class properties)
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				tracker.pushValid()
			},
			rule.ListenerOnExit(ast.KindPropertyDeclaration): func(node *ast.Node) {
				tracker.pop()
			},

			// Note: TypeScript accessor properties are handled as PropertyDeclaration with modifiers

			// Constructor
			ast.KindConstructor: func(node *ast.Node) {
				tracker.pushValid()
			},
			rule.ListenerOnExit(ast.KindConstructor): func(node *ast.Node) {
				tracker.pop()
			},

			// Methods
			ast.KindMethodDeclaration: func(node *ast.Node) {
				tracker.pushValid()
			},
			rule.ListenerOnExit(ast.KindMethodDeclaration): func(node *ast.Node) {
				tracker.pop()
			},

			// Getter/Setter
			ast.KindGetAccessor: func(node *ast.Node) {
				tracker.pushValid()
			},
			rule.ListenerOnExit(ast.KindGetAccessor): func(node *ast.Node) {
				tracker.pop()
			},
			ast.KindSetAccessor: func(node *ast.Node) {
				tracker.pushValid()
			},
			rule.ListenerOnExit(ast.KindSetAccessor): func(node *ast.Node) {
				tracker.pop()
			},

			// Function declarations
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				valid := false
				// TypeScript 'this' parameter always makes function valid
				if hasThisParameter(node) {
					valid = true
				} else if hasThisJSDocTag(node, ctx.SourceFile) {
					valid = true
				} else if isConstructor(node, opts.CapIsConstructor) {
					valid = true
				} else if isInClassContext(node) {
					valid = true
				}

				if valid {
					tracker.pushValid()
				} else {
					tracker.pushInvalid()
				}
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				tracker.pop()
			},

			// Function expressions
			ast.KindFunctionExpression: func(node *ast.Node) {
				valid := false
				// TypeScript 'this' parameter always makes function valid
				if hasThisParameter(node) {
					valid = true
				} else if hasThisJSDocTag(node, ctx.SourceFile) {
					valid = true
				} else if isConstructor(node, opts.CapIsConstructor) {
					valid = true
				} else if isValidMethodContext(node) {
					valid = true
				} else if isInDefinePropertyContext(node) {
					valid = true
				} else if isInObjectLiteralContext(node) {
					valid = true
				} else if isInFunctionBinding(node) {
					valid = true
				} else if isReturnedFromIIFE(node) {
					valid = true
				} else if isInClassContext(node) {
					valid = true
				} else {
					// Check if in a valid call context (array methods with thisArg, etc.)
					parent := node.Parent
					if parent != nil && parent.Kind == ast.KindCallExpression {
						if isValidCallContext(parent, node) {
							valid = true
						}
					}
				}

				if valid {
					tracker.pushValid()
				} else {
					tracker.pushInvalid()
				}
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
				tracker.pop()
			},

			// Arrow functions
			ast.KindArrowFunction: func(node *ast.Node) {
				// Arrow functions inherit 'this' from parent scope
				// Don't change the stack
			},

			// ThisExpression - the actual check
			ast.KindThisKeyword: func(node *ast.Node) {
				if !tracker.isCurrentValid() {
					// Report on the exact 'this' keyword position
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedThis",
						Description: "Unexpected 'this'.",
					})
				}
			},
		}
	},
}
