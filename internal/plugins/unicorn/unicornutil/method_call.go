package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// DotMethodCall is a non-computed method call matched by MatchDotMethodCall.
type DotMethodCall struct {
	Call      *ast.Node
	RawCallee *ast.Node
	Callee    *ast.Node
	Object    *ast.Node
	Property  *ast.Node
}

// DotMethodCallOptions mirrors the subset of unicorn's isMethodCall options
// used by native rslint rules. Nil argument limits disable that check. Method
// matches a single name; Methods matches any name in the set. Supply at most
// one of them.
type DotMethodCallOptions struct {
	Method           string
	Methods          []string
	ArgumentsLength  *int
	MinimumArguments *int
	MaximumArguments *int
	// RejectSpreadElement opts into upstream isMethodCall's default
	// `allowSpreadElement: false` behavior: a SpreadElement argument appearing
	// below the effective argument bound disqualifies the match. It defaults to
	// off so existing callers keep matching spread-argument calls.
	RejectSpreadElement bool
	AllowOptionalCall   bool
	AllowOptionalMember bool
}

// MatchDotMethodCall matches a dot-property CallExpression. Static bracket
// access is intentionally excluded because unicorn's isMethodCall defaults to
// computed:false. Parentheses and JavaScript JSDoc casts are transparent
// except around optional-chain callees, matching the ESTree view upstream.
// Authored TypeScript assertions remain visible and therefore unmatched.
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
	if options.MaximumArguments != nil && len(args) > *options.MaximumArguments {
		return DotMethodCall{}, false
	}
	// Mirror upstream's spread rejection: when spread elements are not allowed,
	// a SpreadElement appearing at an index below the effective maximum-args
	// bound (maximumArguments, else argumentsLength) disqualifies the match.
	if options.RejectSpreadElement {
		bound := -1
		if options.MaximumArguments != nil {
			bound = *options.MaximumArguments
		} else if options.ArgumentsLength != nil {
			bound = *options.ArgumentsLength
		}
		if bound >= 0 {
			for i, arg := range args {
				if i < bound && arg.Kind == ast.KindSpreadElement {
					return DotMethodCall{}, false
				}
			}
		}
	}

	rawCallee := call.Expression
	callee := utils.ESTreeRuntimeExpression(rawCallee)
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
		!matchesMethodName(property.AsIdentifier().Text, options) {
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

// matchesMethodName reports whether name satisfies the option's Method /
// Methods constraint. Methods takes precedence when non-empty; otherwise Method
// is matched. When neither is set, any name matches.
func matchesMethodName(name string, options DotMethodCallOptions) bool {
	if len(options.Methods) > 0 {
		for _, m := range options.Methods {
			if name == m {
				return true
			}
		}
		return false
	}
	if options.Method != "" {
		return name == options.Method
	}
	return true
}
