package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// HasStaticAccessExpressionKey reports whether a member has a statically known
// string or symbol key. Unlike EvalAccessExpressionName, it does not coerce a
// symbol into a string name.
func (staticEvaluator *StaticStringEvaluator) HasStaticAccessExpressionKey(node *ast.Node) bool {
	if _, known := staticEvaluator.EvalAccessExpressionName(node); known {
		return true
	}
	return staticEvaluator != nil && staticEvaluator.resolveIdentifiers &&
		node != nil && node.Kind == ast.KindElementAccessExpression &&
		staticEvaluator.isStaticSymbolKey(node.AsElementAccessExpression().ArgumentExpression)
}

func (staticEvaluator *StaticStringEvaluator) isStaticSymbolKey(node *ast.Node) bool {
	node = SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if ast.IsIdentifier(node) {
		initializer, symbol, known := staticEvaluator.resolveIdentifierInitializer(node)
		if !known || staticEvaluator.resolving[symbol] {
			return false
		}
		staticEvaluator.resolving[symbol] = true
		defer delete(staticEvaluator.resolving, symbol)
		return staticEvaluator.isStaticSymbolKey(initializer)
	}

	member := node
	if ast.IsCallExpression(node) {
		member = SkipAssertionsAndParens(node.AsCallExpression().Expression)
	}
	if !ast.IsAccessExpression(member) ||
		!staticEvaluator.isBuiltinIdentifier(SkipAssertionsAndParens(AccessExpressionObject(member)), "Symbol") {
		return false
	}
	name, known := staticEvaluator.EvalAccessExpressionName(member)
	if !known {
		return false
	}
	if !ast.IsCallExpression(node) {
		// These standard Symbol properties are immutable well-known symbols.
		switch name {
		case "asyncIterator", "hasInstance", "isConcatSpreadable", "iterator",
			"match", "matchAll", "replace", "search", "species", "split",
			"toPrimitive", "toStringTag", "unscopables", "dispose", "asyncDispose":
			return true
		}
		return false
	}
	if name != "for" {
		return false
	}
	arguments, known := staticEvaluator.evalCallArguments(node)
	if !known {
		return false
	}
	if len(arguments) == 0 {
		return true
	}
	_, known = staticValueToString(arguments[0])
	return known
}
