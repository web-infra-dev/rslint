package unicornutil

import "github.com/microsoft/typescript-go/shim/ast"

// DotMethodCall is a non-computed method call matched by MatchDotMethodCall.
type DotMethodCall struct {
	Call      *ast.Node
	RawCallee *ast.Node
	Callee    *ast.Node
	Object    *ast.Node
	Property  *ast.Node
}

// DotMethodCallOptions mirrors the subset of unicorn's isMethodCall options
// used by native rslint rules. Nil argument limits disable that check.
type DotMethodCallOptions struct {
	Method              string
	ArgumentsLength     *int
	MinimumArguments    *int
	AllowOptionalCall   bool
	AllowOptionalMember bool
}

// MatchDotMethodCall matches a dot-property CallExpression. Static bracket
// access is intentionally excluded because unicorn's isMethodCall defaults to
// computed:false. Parentheses are transparent except around optional-chain
// callees, matching the upstream helper's semantics.
func MatchDotMethodCall(node *ast.Node, options DotMethodCallOptions) (DotMethodCall, bool) {
	if node == nil || !ast.IsCallExpression(node) {
		return DotMethodCall{}, false
	}

	call := node.AsCallExpression()
	if !options.AllowOptionalCall && ast.IsOptionalChainRoot(node) {
		return DotMethodCall{}, false
	}

	args := node.Arguments()
	if options.ArgumentsLength != nil && len(args) != *options.ArgumentsLength {
		return DotMethodCall{}, false
	}
	if options.MinimumArguments != nil && len(args) < *options.MinimumArguments {
		return DotMethodCall{}, false
	}

	rawCallee := call.Expression
	callee := ast.SkipParentheses(rawCallee)
	if callee == nil || (rawCallee != callee && ast.IsOptionalChain(callee)) ||
		!ast.IsPropertyAccessExpression(callee) {
		return DotMethodCall{}, false
	}

	propertyAccess := callee.AsPropertyAccessExpression()
	if propertyAccess == nil ||
		(!options.AllowOptionalMember && ast.IsOptionalChainRoot(callee)) {
		return DotMethodCall{}, false
	}

	property := propertyAccess.Name()
	if property == nil || !ast.IsIdentifier(property) ||
		property.AsIdentifier().Text != options.Method {
		return DotMethodCall{}, false
	}

	return DotMethodCall{
		Call:      node,
		RawCallee: rawCallee,
		Callee:    callee,
		Object:    propertyAccess.Expression,
		Property:  property,
	}, true
}
