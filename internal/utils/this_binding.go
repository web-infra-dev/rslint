package utils

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
)


// IsDefaultThisBinding checks whether a function's 'this' binding defaults to the
// global object. This mirrors ESLint's astUtils.isDefaultThisBinding.
// Returns true when 'this' defaults to global; false when explicitly bound.
func IsDefaultThisBinding(node *ast.Node) bool {
	current := node
	for {
		// Walk current → parent, skipping transparent wrappers (parens, TS assertions).
		// After this, current is the outermost wrapper and parent is the real consumer.
		parent := current.Parent
		for parent != nil && ast.IsOuterExpression(parent, skipTransparentKinds) {
			current = parent
			parent = current.Parent
		}
		if parent == nil {
			return true
		}

		switch parent.Kind {
		// { foo: function() {} } → this is the object
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
			return false

		// obj.foo = function() {} → this is the object
		case ast.KindBinaryExpression:
			bin := parent.AsBinaryExpression()
			opKind := bin.OperatorToken.Kind
			if opKind == ast.KindEqualsToken && bin.Right == current {
				left := ast.SkipParentheses(bin.Left)
				if left != nil && (ast.IsPropertyAccessExpression(left) || ast.IsElementAccessExpression(left)) {
					return false
				}
				return true
			}
			// Logical (||, &&, ??) → transparent
			if opKind == ast.KindBarBarToken || opKind == ast.KindAmpersandAmpersandToken ||
				opKind == ast.KindQuestionQuestionToken {
				current = parent
				continue
			}
			// Comma → last element transparent
			if opKind == ast.KindCommaToken && bin.Right == current {
				current = parent
				continue
			}
			return true

		case ast.KindConditionalExpression:
			current = parent
			continue

			// return function() {} — only transparent if inside an IIFE
		case ast.KindReturnStatement:
			fn := ast.GetContainingFunction(parent)
			if fn != nil && IsCallee(fn) {
				current = fn
				continue
			}
			return true

		// (function(){})() — callee of call → not default
		case ast.KindCallExpression:
			call := parent.AsCallExpression()
			if call.Expression == current {
				return false
			}
			// Check for known methods with thisArg: arr.forEach(cb, thisArg)
			return !isCallbackWithThisArg(call, current)

		case ast.KindNewExpression:
			if parent.AsNewExpression().Expression == current {
				return false
			}
			return true

		// (function(){}).call(obj) / .apply(obj) / .bind(obj)
		case ast.KindPropertyAccessExpression:
			prop := parent.AsPropertyAccessExpression()
			if prop.Expression == current {
				name := prop.Name()
				if name != nil && isCallApplyBind(name.Text()) {
					// Check if the member expression is actually called: func.call(obj)
					callGrandparent := parent.Parent
					for callGrandparent != nil && ast.IsOuterExpression(callGrandparent, skipTransparentKinds) {
						callGrandparent = callGrandparent.Parent
					}
					if callGrandparent != nil && ast.IsCallExpression(callGrandparent) {
						call := callGrandparent.AsCallExpression()
						if call.Arguments != nil && len(call.Arguments.Nodes) >= 1 &&
							!IsNullOrUndefined(call.Arguments.Nodes[0]) {
							return false
						}
					}
				}
			}
			return true

		// var Foo = function() {} — uppercase name convention → constructor
		case ast.KindVariableDeclaration:
			varDecl := parent.AsVariableDeclaration()
			if varDecl.Initializer == current {
				varName := varDecl.Name()
				if varName != nil && ast.IsIdentifier(varName) && startsWithUpperCase(varName.AsIdentifier().Text) {
					// Anonymous function assigned to uppercase variable → likely constructor
					if isFuncAnonymous(node) {
						return false
					}
				}
			}
			return true

		default:
			return true
		}
	}
}

// startsWithUpperCase checks if a string starts with an uppercase ASCII letter.
func startsWithUpperCase(s string) bool {
	return len(s) > 0 && s[0] >= 'A' && s[0] <= 'Z'
}

// isFuncAnonymous checks if a function expression has no name.
func isFuncAnonymous(node *ast.Node) bool {
	if node.Kind == ast.KindFunctionExpression {
		return node.AsFunctionExpression().Name() == nil
	}
	return false
}

func isCallApplyBind(name string) bool {
	return name == "call" || name == "apply" || name == "bind"
}

// IsNullOrUndefined checks if a node is a null literal, undefined identifier,
// or void expression, unwrapping parentheses.
func IsNullOrUndefined(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNullKeyword {
		return true
	}
	if ast.IsIdentifier(node) && node.AsIdentifier().Text == "undefined" {
		return true
	}
	if node.Kind == ast.KindVoidExpression {
		return true
	}
	return false
}

// Methods known to accept a thisArg parameter: method(callback, thisArg).
var thisArgMethods = []string{
	"every", "filter", "find", "findIndex", "findLast", "findLastIndex",
	"flatMap", "forEach", "map", "some",
}

// isCallbackWithThisArg checks if the callback is passed to a known method
// with a non-null thisArg: arr.forEach(callback, thisArg).
func isCallbackWithThisArg(call *ast.CallExpression, callback *ast.Node) bool {
	if call.Arguments == nil {
		return false
	}
	args := call.Arguments.Nodes

	// Reflect.apply(callback, thisArg, args)
	if isSpecificMemberAccess(call.Expression, "Reflect", "apply") {
		return len(args) >= 3 && args[0] == callback && !IsNullOrUndefined(args[1])
	}

	// Array.from(iterable, callback, thisArg) / Array.fromAsync(...)
	if isSpecificMemberAccess(call.Expression, "Array", "from") ||
		isSpecificMemberAccess(call.Expression, "Array", "fromAsync") {
		return len(args) >= 3 && args[1] == callback && !IsNullOrUndefined(args[2])
	}

	// arr.forEach(callback, thisArg), arr.map(callback, thisArg), etc.
	callee := call.Expression
	if ast.IsPropertyAccessExpression(callee) {
		prop := callee.AsPropertyAccessExpression()
		name := prop.Name()
		if name != nil && slices.Contains(thisArgMethods, name.Text()) {
			return len(args) == 2 && args[0] == callback && !IsNullOrUndefined(args[1])
		}
	}

	return false
}

// isSpecificMemberAccess checks if node is Object.method pattern.
func isSpecificMemberAccess(node *ast.Node, objectName, methodName string) bool {
	if node == nil || !ast.IsPropertyAccessExpression(node) {
		return false
	}
	prop := node.AsPropertyAccessExpression()
	name := prop.Name()
	if name == nil || name.Text() != methodName {
		return false
	}
	obj := ast.SkipParentheses(prop.Expression)
	return obj != nil && ast.IsIdentifier(obj) && obj.AsIdentifier().Text == objectName
}
