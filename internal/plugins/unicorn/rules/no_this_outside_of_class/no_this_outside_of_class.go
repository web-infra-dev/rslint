package no_this_outside_of_class

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-this-outside-of-class"

var defaultMessage = rule.RuleMessage{
	Id:          messageID,
	Description: "Do not use `this` outside of classes.",
}

// NoThisOutsideOfClassRule disallows using `this` outside of classes.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v72.0.0/rules/no-this-outside-of-class.js
var NoThisOutsideOfClassRule = rule.Rule{
	Name:   "unicorn/no-this-outside-of-class",
	Schema: rule.EmptyArraySchema,
	Run:    run,
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	checkThis := func(node *ast.Node) {
		if utils.IsInJsxTagName(node) {
			return
		}
		if node.Parent != nil && ast.IsThisParameter(node.Parent) && node.Parent.Name() == node {
			return
		}
		if !isAllowedThisBinding(node) {
			ctx.ReportNode(node, defaultMessage)
		}
	}

	return rule.RuleListeners{
		ast.KindThisKeyword: checkThis,
		ast.KindThisType: func(node *ast.Node) {
			if isTypeQueryRoot(node) {
				checkThis(node)
			}
		},
		ast.KindIdentifier: func(node *ast.Node) {
			if node.AsIdentifier().Text == "this" && isTypeQueryRoot(node) {
				checkThis(node)
			}
		},
	}
}

func isAllowedThisBinding(node *ast.Node) bool {
	child := node
	parent := node.Parent

	for parent != nil {
		if parent.Kind == ast.KindArrowFunction {
			child = parent
			parent = parent.Parent
			continue
		}

		if parent.Kind == ast.KindFunctionDeclaration || parent.Kind == ast.KindFunctionExpression {
			// In TypeScript, an overload signature without a body has no ESTree
			// FunctionExpression representation (TSDeclareFunction) and does not
			// bind `this`; it passes through to the enclosing scope.
			if parent.Body() == nil {
				child = parent
				parent = parent.Parent
				continue
			}
			return hasThisParameter(parent)
		}

		if isMethodLike(parent) {
			// In ESTree, a class method's FunctionExpression wraps parameters,
			// type parameters, return type, and body — but only when implemented.
			// Body-less signatures (overloads, abstract methods) do not bind `this`
			// and continue upward to the enclosing scope.
			if isInsideImplementedMethodFunction(parent, child) {
				return (parent.Parent != nil && ast.IsClassLike(parent.Parent)) || hasThisParameter(parent)
			}
			child = parent
			parent = parent.Parent
			continue
		}

		if parent.Kind == ast.KindClassStaticBlockDeclaration {
			return true
		}

		if parent.Kind == ast.KindPropertyDeclaration {
			prop := parent.AsPropertyDeclaration()
			if prop != nil && prop.Initializer == child && parent.Parent != nil && ast.IsClassLike(parent.Parent) {
				return true
			}
			child = parent
			parent = parent.Parent
			continue
		}

		child = parent
		parent = parent.Parent
	}

	return false
}

func isMethodLike(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindMethodDeclaration,
		ast.KindConstructor,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		return true
	default:
		return false
	}
}

func isInsideImplementedMethodFunction(method *ast.Node, child *ast.Node) bool {
	if child == nil || method.Body() == nil {
		return false
	}
	// Name (computed property name) is the method's key in ESTree terms,
	// and modifiers (decorators) attach to the MethodDefinition, outside the FunctionExpression.
	if child == method.Name() || child.Kind == ast.KindDecorator {
		return false
	}
	return true
}

func hasThisParameter(node *ast.Node) bool {
	for _, p := range node.Parameters() {
		if ast.IsThisParameter(p) {
			return true
		}
	}
	return false
}

func isTypeQueryRoot(node *ast.Node) bool {
	current := node
	for current != nil && current.Parent != nil && current.Parent.Kind == ast.KindQualifiedName {
		qualified := current.Parent.AsQualifiedName()
		if qualified == nil || qualified.Left != current {
			return false
		}
		current = current.Parent
	}
	return current != nil && current.Parent != nil && current.Parent.Kind == ast.KindTypeQuery
}
