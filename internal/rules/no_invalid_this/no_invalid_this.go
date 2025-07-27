package no_invalid_this

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
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
	
	// For function expressions in call arguments, we need to look at the parent context
	searchNode := node
	if node.Parent != nil && node.Parent.Kind == ast.KindCallExpression {
		// Check if this function is an argument to a call
		callExpr := node.Parent.AsCallExpression()
		for _, arg := range callExpr.Arguments.Nodes {
			if arg == node {
				// This function is a call argument, search from the call expression
				searchNode = node.Parent
				break
			}
		}
	}
	
	// Look backwards from node position for JSDoc comment
	// We need to find the actual function keyword position, not the start of JSDoc
	searchPos := searchNode.Pos()
	
	// For function declarations, the position might include the JSDoc
	// So let's look from the node position up to the end position
	searchEnd := searchNode.End()
	if searchEnd > len(text) {
		searchEnd = len(text)
	}
	
	// Look for the actual function text to determine where to search from
	nodeText := text[searchPos:searchEnd]
	funcIndex := strings.Index(nodeText, "function")
	if funcIndex >= 0 {
		// Adjust search position to the actual function keyword
		searchPos = searchPos + funcIndex
	}
	
	searchStart := searchPos
	if searchStart > 500 {
		searchStart = searchStart - 500
	} else {
		searchStart = 0
	}
	
	commentArea := text[searchStart:searchPos]
	
	// Check for @this in different comment styles
	// 1. /** @this ... */ (JSDoc block comment)
	// 2. /* @this ... */ (regular block comment)  
	// 3. // @this ... (line comment - less common but possible)
	
	// First check if there's any @this tag
	if !strings.Contains(commentArea, "@this") {
		return false
	}
	
	// Now verify it's in a proper comment context
	// Look for comment patterns and ensure @this is within them
	inBlockComment := false
	inLineComment := false
	
	for i := 0; i < len(commentArea); i++ {
		// Check for block comment start
		if i < len(commentArea)-1 && commentArea[i] == '/' && commentArea[i+1] == '*' {
			inBlockComment = true
			i++ // Skip the *
			continue
		}
		
		// Check for block comment end
		if inBlockComment && i < len(commentArea)-1 && commentArea[i] == '*' && commentArea[i+1] == '/' {
			inBlockComment = false
			i++ // Skip the /
			continue
		}
		
		// Check for line comment start
		if i < len(commentArea)-1 && commentArea[i] == '/' && commentArea[i+1] == '/' {
			inLineComment = true
			i++ // Skip the second /
			continue
		}
		
		// Line comments end at newline
		if inLineComment && commentArea[i] == '\n' {
			inLineComment = false
			continue
		}
		
		// Check if we found @this within a comment
		if (inBlockComment || inLineComment) && i+5 <= len(commentArea) && commentArea[i:i+5] == "@this" {
			// Verify it's followed by whitespace or end of comment
			if i+5 == len(commentArea) || isWhitespace(commentArea[i+5]) || commentArea[i+5] == '*' {
				return true
			}
		}
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
	
	// Check if nested in object definition patterns
	return isInObjectLiteralContext(node) || isInDefinePropertyContext(node) || isInFunctionBinding(node) || isReturnedFromIIFE(node)
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
		if options != nil {
			if optsMap, ok := options.(map[string]interface{}); ok {
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
				// TypeScript 'this' parameter always makes function valid
				if hasThisParameter(node) {
					tracker.pushValid()
				} else if hasThisJSDocTag(node, ctx.SourceFile) ||
				   isConstructor(node, opts.CapIsConstructor) ||
				   isInClassContext(node) {
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
				// TypeScript 'this' parameter always makes function valid
				if hasThisParameter(node) {
					tracker.pushValid()
				} else if hasThisJSDocTag(node, ctx.SourceFile) ||
				   isConstructor(node, opts.CapIsConstructor) ||
				   isValidMethodContext(node) ||
				   isInClassContext(node) {
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
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedThis",
						Description: "Unexpected 'this'.",
					})
				}
			},
		}
	},
}