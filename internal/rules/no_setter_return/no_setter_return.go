package no_setter_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildSetterMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "returnsValue",
		Description: "Setter cannot return a value.",
	}
}

// findEnclosingFunction walks up from a node to find the nearest enclosing
// function-like boundary.
func findEnclosingFunction(node *ast.Node) *ast.Node {
	return ast.FindAncestor(node.Parent, ast.IsFunctionLikeDeclaration)
}

// isGlobalReference checks if an identifier refers to a global variable (not shadowed locally).
func isGlobalReference(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	if ctx.TypeChecker == nil {
		return true
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}
	return utils.IsSymbolFromDefaultLibrary(ctx.Program, symbol)
}

// matchGlobalMethodCall checks if a CallExpression calls objectName.methodName
// where objectName is a global reference. Handles dot access, bracket access,
// optional chaining, and parenthesized callee expressions.
func matchGlobalMethodCall(callNode *ast.Node, ctx rule.RuleContext, objectName, methodName string) bool {
	callee := ast.SkipParentheses(callNode.Expression())
	if callee == nil {
		return false
	}

	var objNode *ast.Node
	var accessedName string

	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		objNode = callee.Expression()
		nameNode := callee.Name()
		if nameNode != nil && nameNode.Kind == ast.KindIdentifier {
			accessedName = nameNode.Text()
		}
	case ast.KindElementAccessExpression:
		objNode = callee.Expression()
		argExpr := callee.AsElementAccessExpression().ArgumentExpression
		if argExpr != nil {
			val, ok := utils.GetStaticExpressionValue(argExpr)
			if ok {
				accessedName = val
			}
		}
	default:
		return false
	}

	if accessedName != methodName || objNode == nil {
		return false
	}
	if objNode.Kind != ast.KindIdentifier || objNode.Text() != objectName {
		return false
	}
	return isGlobalReference(objNode, ctx)
}

// isPropertyDescriptor checks if an ObjectLiteralExpression is used as a property descriptor
// in calls to Object.defineProperty, Reflect.defineProperty, Object.defineProperties, or Object.create.
func isPropertyDescriptor(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return false
	}

	parent := node.Parent
	if parent == nil {
		return false
	}

	// Direct descriptor: 3rd argument of Object.defineProperty or Reflect.defineProperty
	if parent.Kind == ast.KindCallExpression {
		args := parent.Arguments()
		if len(args) >= 3 && args[2] == node {
			if matchGlobalMethodCall(parent, ctx, "Object", "defineProperty") ||
				matchGlobalMethodCall(parent, ctx, "Reflect", "defineProperty") {
				return true
			}
		}
	}

	// Nested descriptor: property value of 2nd argument of Object.create or Object.defineProperties
	if parent.Kind == ast.KindPropertyAssignment && parent.Initializer() == node {
		grandparent := parent.Parent
		if grandparent != nil && grandparent.Kind == ast.KindObjectLiteralExpression {
			greatGrandparent := grandparent.Parent
			if greatGrandparent != nil && greatGrandparent.Kind == ast.KindCallExpression {
				args := greatGrandparent.Arguments()
				if len(args) >= 2 && args[1] == grandparent {
					if matchGlobalMethodCall(greatGrandparent, ctx, "Object", "create") ||
						matchGlobalMethodCall(greatGrandparent, ctx, "Object", "defineProperties") {
						return true
					}
				}
			}
		}
	}

	return false
}

// isSetterFunction checks if a function-like node is used as a setter.
func isSetterFunction(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil {
		return false
	}

	// Set accessor in object literal or class
	if node.Kind == ast.KindSetAccessor {
		return true
	}

	parent := node.Parent
	if parent == nil {
		return false
	}

	// PropertyAssignment: { set: function(val) { ... } } or { set: (val) => ... }
	if parent.Kind == ast.KindPropertyAssignment && parent.Initializer() == node {
		propName, ok := utils.GetStaticPropertyName(parent.Name())
		if ok && propName == "set" {
			descriptor := parent.Parent
			if descriptor != nil && descriptor.Kind == ast.KindObjectLiteralExpression {
				return isPropertyDescriptor(descriptor, ctx)
			}
		}
		return false
	}

	// MethodDeclaration: { set(val) { ... } } shorthand in a descriptor
	if node.Kind == ast.KindMethodDeclaration {
		propName, ok := utils.GetStaticPropertyName(node.Name())
		if ok && propName == "set" {
			if parent.Kind == ast.KindObjectLiteralExpression {
				return isPropertyDescriptor(parent, ctx)
			}
		}
	}

	return false
}

// NoSetterReturnRule disallows returning a value from a setter.
// Setters cannot meaningfully return values; any return value is silently ignored.
// A bare `return;` (without a value) is allowed for control flow.
var NoSetterReturnRule = rule.Rule{
	Name: "no-setter-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				ret := node.AsReturnStatement()
				if ret.Expression == nil {
					return
				}
				enclosing := findEnclosingFunction(node)
				if enclosing != nil && isSetterFunction(enclosing, ctx) {
					ctx.ReportNode(node, buildSetterMessage())
				}
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				// Arrow with expression body used as setter in property descriptor:
				// e.g., { set: val => val }
				body := node.Body()
				if body == nil || body.Kind == ast.KindBlock {
					return
				}
				if isSetterFunction(node, ctx) {
					ctx.ReportNode(body, buildSetterMessage())
				}
			},
		}
	},
}
