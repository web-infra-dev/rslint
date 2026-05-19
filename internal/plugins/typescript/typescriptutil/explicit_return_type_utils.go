// Package typescriptutil holds helpers shared across typescript-eslint rules.
//
// This file mirrors upstream's util/explicitReturnTypeUtils.ts. It is consumed
// by both `explicit-function-return-type` and `explicit-module-boundary-types`,
// which share semantic predicates for deciding whether a function expression
// is in a typed context (parent variable annotation, type assertion, JSX
// container, ...) and whether an ancestor already supplies the return type.
package typescriptutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// IsFunction reports whether a node is FunctionDeclaration, FunctionExpression
// or ArrowFunction. Matches ESLint's ASTUtils.isFunction — excludes methods,
// getters, setters, and constructors.
func IsFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return ast.IsFunctionDeclaration(node) || ast.IsFunctionExpressionOrArrowFunction(node)
}

// IsTypeAssertion reports whether a node is `x as T` or `<T>x`.
func IsTypeAssertion(node *ast.Node) bool {
	return ast.IsAsExpression(node) || ast.IsTypeAssertion(node)
}

// GetEffectiveParent returns the first meaningful parent, skipping
// ParenthesizedExpressions. tsgo preserves parens that ESLint strips, so this
// bridges the gap when walking from a function to its containing context.
func GetEffectiveParent(node *ast.Node) *ast.Node {
	if node == nil || node.Parent == nil {
		return nil
	}
	return ast.WalkUpParenthesizedExpressions(node.Parent)
}

// IsVariableDeclaratorWithTypeAnnotation reports whether a node is a
// VariableDeclaration (tsgo's equivalent of ESTree's VariableDeclarator) with
// an explicit type annotation, e.g. `const x: Foo = ...`.
func IsVariableDeclaratorWithTypeAnnotation(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindVariableDeclaration {
		return false
	}
	return node.AsVariableDeclaration().Type != nil
}

// IsPropertyDefinitionWithTypeAnnotation reports whether a node is a
// PropertyDeclaration with a type annotation, e.g. `public x: Foo = ...`.
func IsPropertyDefinitionWithTypeAnnotation(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPropertyDeclaration {
		return false
	}
	return node.AsPropertyDeclaration().Type != nil
}

// IsDefaultFunctionParameterWithTypeAnnotation reports whether a node is a
// Parameter with both a type annotation and an initializer (default value),
// e.g. `(param: Type = () => {})`.
func IsDefaultFunctionParameterWithTypeAnnotation(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindParameter {
		return false
	}
	param := node.AsParameterDeclaration()
	return param.Type != nil && param.Initializer != nil
}

// IsFunctionArgument reports whether `parent` is a CallExpression and
// `funcNode` is one of its arguments (not its callee — that would be an IIFE).
func IsFunctionArgument(parent *ast.Node, funcNode *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
	return callee != funcNode
}

// IsTypedJSX reports whether a node is JsxExpression or JsxSpreadAttribute —
// the two JSX containers ESTree models as `JSXExpressionContainer` /
// `JSXSpreadAttribute`.
func IsTypedJSX(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindJsxExpression || node.Kind == ast.KindJsxSpreadAttribute
}

// IsConstructorArgument reports whether a node is a NewExpression. Used to
// detect functions passed directly as constructor arguments, e.g.
// `new Foo(() => {})`.
func IsConstructorArgument(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindNewExpression
}

// IsTypedParent reports whether the parent of a function expression provides
// type context: type assertion, typed variable declarator, default parameter
// with type, typed property declaration, function-call argument, or JSX
// container.
func IsTypedParent(parent *ast.Node, funcNode *ast.Node) bool {
	if parent == nil {
		return false
	}
	return IsTypeAssertion(parent) ||
		IsVariableDeclaratorWithTypeAnnotation(parent) ||
		IsDefaultFunctionParameterWithTypeAnnotation(parent) ||
		IsPropertyDefinitionWithTypeAnnotation(parent) ||
		IsFunctionArgument(parent, funcNode) ||
		IsTypedJSX(parent)
}

// IsPropertyOfObjectWithType reports whether a node is a property (or nested
// property) of a typed object expression. Mirrors upstream's
// `isPropertyOfObjectWithType` walk.
func IsPropertyOfObjectWithType(property *ast.Node, funcNode *ast.Node) bool {
	if property == nil {
		return false
	}
	if property.Kind != ast.KindPropertyAssignment &&
		property.Kind != ast.KindShorthandPropertyAssignment {
		return false
	}
	objectExpr := property.Parent
	if objectExpr == nil || objectExpr.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	parent := GetEffectiveParent(objectExpr)
	if parent == nil {
		return false
	}
	return IsTypedParent(parent, funcNode) || IsPropertyOfObjectWithType(parent, funcNode)
}

// IsTypedFunctionExpression reports whether a function expression is in a
// typed context, gated by allowTypedFunctionExpressions.
func IsTypedFunctionExpression(node *ast.Node, allowTypedFunctionExpressions bool) bool {
	if !allowTypedFunctionExpressions {
		return false
	}
	parent := GetEffectiveParent(node)
	if parent == nil {
		return false
	}
	return IsTypedParent(parent, node) ||
		IsPropertyOfObjectWithType(parent, node) ||
		IsConstructorArgument(parent)
}

// IsValidFunctionExpressionReturnType reports whether a function expression's
// return type is typed (per IsTypedFunctionExpression), is wrapped in a
// permitted expression context (allowExpressions), or is an arrow function
// returning `as const` (allowDirectConstAssertionInArrowFunctions).
func IsValidFunctionExpressionReturnType(
	node *ast.Node,
	allowTypedFunctionExpressions bool,
	allowExpressions bool,
	allowDirectConstAssertionInArrowFunctions bool,
) bool {
	if IsTypedFunctionExpression(node, allowTypedFunctionExpressions) {
		return true
	}

	if allowExpressions {
		parent := GetEffectiveParent(node)
		if parent != nil &&
			parent.Kind != ast.KindVariableDeclaration &&
			parent.Kind != ast.KindMethodDeclaration &&
			parent.Kind != ast.KindExportAssignment &&
			parent.Kind != ast.KindPropertyDeclaration {
			return true
		}
	}

	if !allowDirectConstAssertionInArrowFunctions || node.Kind != ast.KindArrowFunction {
		return false
	}

	af := node.AsArrowFunction()
	body := af.Body
	if body == nil {
		return false
	}
	// tsgo preserves ParenthesizedExpression and exposes `satisfies` as
	// SatisfiesExpression — both must be peeled to reach a possible `as const`.
	body = ast.SkipParentheses(body)
	for body.Kind == ast.KindSatisfiesExpression {
		body = ast.SkipParentheses(body.AsSatisfiesExpression().Expression)
	}
	return ast.IsConstAssertion(body)
}

// DoesImmediatelyReturnFunctionExpression reports whether `node` is a function
// whose body (or every `return` statement) yields another function expression.
// Mirrors upstream's `doesImmediatelyReturnFunctionExpression`.
func DoesImmediatelyReturnFunctionExpression(node *ast.Node, returns []*ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindArrowFunction {
		af := node.AsArrowFunction()
		if af.Body != nil && af.Body.Kind != ast.KindBlock {
			return IsFunction(ast.SkipParentheses(af.Body))
		}
	}
	if len(returns) == 0 {
		return false
	}
	for _, ret := range returns {
		arg := ret.Expression()
		if arg == nil || !IsFunction(ast.SkipParentheses(arg)) {
			return false
		}
	}
	return true
}

// AncestorHasReturnType reports whether any ancestor of `node` already supplies
// a return type, making `node`'s own return type unnecessary. Mirrors
// upstream's `ancestorHasReturnType`.
func AncestorHasReturnType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	ancestor := node.Parent
	// tsgo preserves ParenthesizedExpression that ESLint strips — peel them
	// before checking the "is this a return value?" gate.
	for ancestor != nil && ancestor.Kind == ast.KindParenthesizedExpression {
		ancestor = ancestor.Parent
	}

	// ESLint's model: `if (ancestor.type === Property) ancestor = ancestor.value;`
	// `Property.value` for `arrowFn: () => 'test'` is the ArrowFunction itself.
	// In tsgo, PropertyAssignment.Initializer holds the expression.
	if ancestor != nil && ancestor.Kind == ast.KindPropertyAssignment {
		pa := ancestor.AsPropertyAssignment()
		if pa.Initializer != nil {
			ancestor = ast.SkipParentheses(pa.Initializer)
		}
	}

	isReturnStatement := ancestor != nil && ancestor.Kind == ast.KindReturnStatement
	isBodylessArrow := ancestor != nil &&
		ancestor.Kind == ast.KindArrowFunction &&
		ancestor.AsArrowFunction().Body != nil &&
		ancestor.AsArrowFunction().Body.Kind != ast.KindBlock
	if !isReturnStatement && !isBodylessArrow {
		return false
	}

	for ancestor != nil {
		switch ancestor.Kind {
		case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindFunctionDeclaration,
			ast.KindMethodDeclaration, ast.KindGetAccessor:
			// In tsgo, methods and getters are their own node types (not
			// FunctionExpression inside MethodDefinition like ESLint). Check
			// their return type the same way.
			if ancestor.Type() != nil {
				return true
			}
		case ast.KindVariableDeclaration:
			return ancestor.AsVariableDeclaration().Type != nil
		case ast.KindPropertyDeclaration:
			return ancestor.AsPropertyDeclaration().Type != nil
		case ast.KindExpressionStatement:
			return false
		}
		ancestor = ancestor.Parent
	}
	return false
}
