package prefer_set_has

// This file ports the array-shape recognition, static-size evaluation, unique-
// array detection, and TypeScript type-annotation rewriting that
// eslint-plugin-unicorn's prefer-set-has rule relies on. These helpers encode
// the rule's own semantics (what counts as an "array source" and a "known
// unique array" for THIS rule), so they stay local rather than being extracted
// to unicornutil.

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// isArrayMethodCall mirrors upstream's `isArrayMethodCall`: whether `node`
// evaluates to an array. `visited` guards the recursive const-initializer walk
// against cycles.
func isArrayMethodCall(ctx rule.RuleContext, node *ast.Node, visited map[*ast.Symbol]bool) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	// `[]`
	if ast.IsArrayLiteralExpression(node) {
		return true
	}

	// `Array()` and `new Array()`
	if isArrayConstructorCall(node) {
		return true
	}

	// `Array.from()` and `Array.of()`
	if matchArrayStaticCall(node, "from") != nil || matchArrayStaticCall(node, "of") != nil {
		return true
	}

	// Methods that always return an array.
	if match, ok := matchListedMethodCall(node, methodsReturnsArray); ok {
		_ = match
		return true
	}

	// Methods that can return an array OR a string: only treat as array when
	// the receiver is provably not a string.
	if match, ok := matchListedMethodCall(node, methodsReturnsArrayAndString); ok {
		object := ast.SkipParentheses(match.Object)
		// Upstream gates on `getStaticStringValue(object) === undefined`, i.e.
		// the receiver is not a string literal / no-substitution template
		// (both of which upstream can read a static string from, including the
		// empty string).
		if _, isString := staticStringLike(object); isString {
			return false
		}
		if object.Kind != ast.KindIdentifier {
			return true
		}
		return isIdentifierInitializedWithArray(ctx, object, visited)
	}

	return false
}

// staticStringLike reports whether node is syntactically a string literal or
// template with no substitutions — the cases upstream's getStaticStringValue
// recognizes as a string (including the empty string).
func staticStringLike(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text, true
	}
	return "", false
}

// isIdentifierInitializedWithArray mirrors upstream's
// `isIdentifierInitializedWithArray`: the identifier resolves to a single
// `const` declaration whose initializer is itself an array source.
func isIdentifierInitializedWithArray(ctx rule.RuleContext, node *ast.Node, visited map[*ast.Symbol]bool) bool {
	if node.Kind != ast.KindIdentifier || ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.Resolve(node)
	if symbol == nil || visited[symbol] || len(symbol.Declarations) != 1 {
		return false
	}
	visited[symbol] = true

	declarationNode := symbol.Declarations[0]
	if declarationNode == nil || declarationNode.Kind != ast.KindVariableDeclaration {
		return false
	}
	declaration := declarationNode.AsVariableDeclaration()
	if declaration.Initializer == nil {
		return false
	}
	declarationList := declarationNode.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList ||
		!ast.IsVarConst(declarationList) {
		return false
	}
	return isArrayMethodCall(ctx, declaration.Initializer, visited)
}

// isArrayConstructorCall reports whether node is `Array(...)` or
// `new Array(...)` (non-optional, identifier callee named Array).
func isArrayConstructorCall(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call.QuestionDotToken != nil {
			return false
		}
		return isIdentifierNamed(ast.SkipParentheses(call.Expression), "Array")
	case ast.KindNewExpression:
		return isIdentifierNamed(ast.SkipParentheses(node.AsNewExpression().Expression), "Array")
	}
	return false
}

// matchArrayStaticCall matches `Array.<method>(...)` (non-optional call and
// member, identifier object named Array, identifier property named method) and
// returns the call node, else nil.
func matchArrayStaticCall(node *ast.Node, method string) *ast.Node {
	match, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{Method: method})
	if !ok {
		return nil
	}
	if !isIdentifierNamed(ast.SkipParentheses(match.Object), "Array") {
		return nil
	}
	return match.Call
}

// matchListedMethodCall matches `<receiver>.<method>(...)` for any method in
// `methods` (non-optional call and member, identifier property).
func matchListedMethodCall(node *ast.Node, methods map[string]bool) (unicornutil.DotMethodCall, bool) {
	if node == nil || !ast.IsCallExpression(node) {
		return unicornutil.DotMethodCall{}, false
	}
	call := node.AsCallExpression()
	if call.QuestionDotToken != nil {
		return unicornutil.DotMethodCall{}, false
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil || !ast.IsPropertyAccessExpression(callee) || ast.IsOptionalChainRoot(callee) {
		return unicornutil.DotMethodCall{}, false
	}
	property := callee.AsPropertyAccessExpression().Name()
	if property == nil || !ast.IsIdentifier(property) || !methods[property.AsIdentifier().Text] {
		return unicornutil.DotMethodCall{}, false
	}
	return unicornutil.DotMethodCall{
		Call:     node,
		Callee:   callee,
		Object:   callee.AsPropertyAccessExpression().Expression,
		Property: property,
	}, true
}

func isIdentifierNamed(node *ast.Node, name string) bool {
	return node != nil && ast.IsIdentifier(node) && node.AsIdentifier().Text == name
}
