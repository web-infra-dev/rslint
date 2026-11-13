package unicornutil

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var mapOrFilterCall = DotMethodCallOptions{
	Methods:          []string{"map", "filter"},
	MinimumArguments: new(1),
}

var iteratorMethodCall = DotMethodCallOptions{
	AllowOptionalCall: true, AllowOptionalMember: true,
}

var lazyIteratorHelperCall = DotMethodCallOptions{
	Methods:             []string{"drop", "filter", "flatMap", "map", "take"},
	MinimumArguments:    new(1),
	AllowOptionalCall:   true,
	AllowOptionalMember: true,
}

// MatchMapOrFilterCall matches the dot calls with arguments considered by both loop rules.
func MatchMapOrFilterCall(node *ast.Node) (DotMethodCall, bool) {
	return MatchDotMethodCall(node, mapOrFilterCall)
}

// UnwrapIteratorExpression skips parentheses and TypeScript wrappers around
// an iterable expression. Optional-chain tokens remain on their original nodes.
func UnwrapIteratorExpression(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments)
}

// IsIteratorExpression reports whether the expression produces a known iterator.
func IsIteratorExpression(ctx *rule.RuleContext, expression *ast.Node) bool {
	return isIteratorExpression(ctx, expression, make(map[*ast.Node]struct{}))
}

// IsLazyIteratorHelperCall matches a lazy helper called on a known iterator.
func IsLazyIteratorHelperCall(ctx *rule.RuleContext, node *ast.Node) bool {
	call, ok := MatchDotMethodCall(node, lazyIteratorHelperCall)
	return ok && IsIteratorExpression(ctx, call.Object)
}

func isIteratorExpression(ctx *rule.RuleContext, expression *ast.Node, visited map[*ast.Node]struct{}) bool {
	node := UnwrapIteratorExpression(expression)
	if node == nil {
		return false
	}
	if _, seen := visited[node]; seen {
		return false
	}
	visited[node] = struct{}{}

	switch node.Kind {
	case ast.KindCallExpression:
		if isIteratorCall(ctx, node, visited) {
			return true
		}
	case ast.KindIdentifier:
		if value := immutableValue(ctx, node); value != nil && isIteratorExpression(ctx, value, visited) {
			return true
		}
	}
	return isKnownIteratorType(ctx, expression)
}

func isIteratorCall(ctx *rule.RuleContext, node *ast.Node, visited map[*ast.Node]struct{}) bool {
	call, ok := MatchDotMethodCall(node, iteratorMethodCall)
	if ok {
		args := node.Arguments()
		method := call.Property.Text()
		switch method {
		case "from", "concat", "zip", "zipKeyed":
			return isGlobalIterator(ctx, call.Object)
		case "entries", "keys", "values":
			return len(args) == 0
		case "matchAll":
			return len(args) == 0 || (len(args) == 1 && args[0].Kind != ast.KindSpreadElement)
		}
		return len(args) > 0 && matchesMethodName(method, lazyIteratorHelperCall) &&
			isIteratorExpression(ctx, call.Object, visited)
	}
	return !ast.IsOptionalChainRoot(node) && isSynchronousGenerator(ctx, node.Expression(), visited)
}

func isGlobalIterator(ctx *rule.RuleContext, node *ast.Node) bool {
	node = UnwrapIteratorExpression(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindPropertyAccessExpression && !ast.IsOptionalChainRoot(node) {
		property := node.Name()
		if property.Kind != ast.KindIdentifier || property.Text() != "Iterator" {
			return false
		}
		return isGlobalName(ctx, UnwrapIteratorExpression(node.Expression()), "globalThis")
	}
	return isGlobalName(ctx, node, "Iterator")
}

func isGlobalName(ctx *rule.RuleContext, node *ast.Node, name string) bool {
	return node != nil && node.Kind == ast.KindIdentifier && node.Text() == name &&
		ctx.Globals.Access(name).IsDeclared() && ctx.Refs.IsGlobalReference(node)
}

func immutableValue(ctx *rule.RuleContext, identifier *ast.Node) *ast.Node {
	symbol := ctx.Refs.ResolveInFile(identifier)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		if bindingIsUnchanged(ctx, symbol) {
			return declaration
		}
	case ast.KindVariableDeclaration:
		if value := declaration.Initializer(); value != nil && ast.IsVarConst(declaration) &&
			(declaration.Type() == nil || utils.IsJSDocSyntaxNode(declaration.Type())) &&
			bindingIsUnchanged(ctx, symbol) {
			return value
		}
	}
	return nil
}

type unchangedBindingCacheKey struct{}

func bindingIsUnchanged(ctx *rule.RuleContext, symbol *ast.Symbol) bool {
	// Several loop headers can share one binding with many references.
	cache := rule.CachedByFile(*ctx, unchangedBindingCacheKey{}, func() map[*ast.Symbol]bool {
		return make(map[*ast.Symbol]bool)
	})
	if unchanged, ok := cache[symbol]; ok {
		return unchanged
	}
	unchanged := !slices.ContainsFunc(ctx.Refs.References(symbol), utils.IsWriteReference)
	cache[symbol] = unchanged
	return unchanged
}

func isSynchronousGenerator(ctx *rule.RuleContext, node *ast.Node, visited map[*ast.Node]struct{}) bool {
	for node != nil {
		node = UnwrapIteratorExpression(node)
		if _, seen := visited[node]; seen {
			return false
		}
		visited[node] = struct{}{}
		switch node.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
			return node.BodyData().AsteriskToken != nil && !ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
		case ast.KindIdentifier:
			node = immutableValue(ctx, node)
		default:
			return false
		}
	}
	return false
}
