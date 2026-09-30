// cspell:ignore unscopables

package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

type staticSymbolKind uint8

const (
	staticSymbolWellKnown staticSymbolKind = iota
	staticSymbolRegistry
)

type staticSymbolValue struct {
	kind staticSymbolKind
	key  string
}

var staticWellKnownSymbolNames = map[string]struct{}{
	"asyncIterator": {}, "hasInstance": {}, "isConcatSpreadable": {},
	"iterator": {}, "match": {}, "matchAll": {}, "replace": {},
	"search": {}, "species": {}, "split": {}, "toPrimitive": {},
	"toStringTag": {}, "unscopables": {}, "dispose": {}, "asyncDispose": {},
}

func (staticEvaluator *StaticStringEvaluator) evalWellKnownSymbolMember(node *ast.Node) staticEvalResult {
	object := SkipAssertionsAndParens(AccessExpressionObject(node))
	if !staticEvaluator.isBuiltinSymbolValue(object, map[*ast.Symbol]bool{}) {
		return staticEvalResult{}
	}
	name, ok := staticEvaluator.evalAccessExpressionKey(node)
	if !ok {
		return staticEvalResult{}
	}
	if _, ok := staticWellKnownSymbolNames[name]; !ok {
		return staticEvalResult{}
	}
	return staticEvalResult{
		value: staticSymbolValue{kind: staticSymbolWellKnown, key: name},
		ok:    true,
	}
}

func (staticEvaluator *StaticStringEvaluator) evalSymbolCall(node *ast.Node) staticEvalResult {
	callee := SkipAssertionsAndParens(node.AsCallExpression().Expression)
	resolving := map[*ast.Symbol]bool{}
	if staticEvaluator.isBuiltinMethodValue(callee, "Symbol", "for", resolving) {
		arguments, ok := staticEvaluator.evalCallArguments(node)
		if !ok {
			return staticEvalResult{}
		}
		key := "undefined"
		if len(arguments) > 0 {
			key, ok = staticValueToString(arguments[0])
			if !ok {
				return staticEvalResult{}
			}
		}
		return staticEvalResult{
			value: staticSymbolValue{kind: staticSymbolRegistry, key: key},
			ok:    true,
		}
	}

	resolving = map[*ast.Symbol]bool{}
	if staticEvaluator.isBuiltinMethodValue(callee, "Symbol", "keyFor", resolving) {
		arguments, ok := staticEvaluator.evalCallArguments(node)
		if !ok || len(arguments) == 0 {
			return staticEvalResult{}
		}
		symbol, ok := arguments[0].(staticSymbolValue)
		if !ok {
			return staticEvalResult{}
		}
		if symbol.kind == staticSymbolRegistry {
			return staticEvalResult{value: symbol.key, ok: true}
		}
		return staticEvalResult{value: staticUndefinedValue{}, ok: true}
	}
	return staticEvalResult{}
}
func (staticEvaluator *StaticStringEvaluator) isBuiltinSymbolValue(
	node *ast.Node,
	resolvingAliases map[*ast.Symbol]bool,
) bool {
	return staticEvaluator.isBuiltinValue(node, "Symbol", resolvingAliases)
}

func (staticEvaluator *StaticStringEvaluator) builtinMethodName(
	node *ast.Node,
	objectName string,
	resolvingAliases map[*ast.Symbol]bool,
) (string, bool) {
	node = SkipAssertionsAndParens(node)
	if ast.IsAccessExpression(node) {
		name, ok := staticEvaluator.evalAccessExpressionKey(node)
		if !ok || !staticEvaluator.isBuiltinValue(
			AccessExpressionObject(node),
			objectName,
			resolvingAliases,
		) {
			return "", false
		}
		return name, true
	}
	initializer, symbol, ok := staticEvaluator.resolveIdentifierInitializer(node)
	if !ok || resolvingAliases[symbol] {
		return "", false
	}
	resolvingAliases[symbol] = true
	defer delete(resolvingAliases, symbol)
	return staticEvaluator.builtinMethodName(
		initializer,
		objectName,
		resolvingAliases,
	)
}

func (staticEvaluator *StaticStringEvaluator) isBuiltinMethodValue(
	node *ast.Node,
	objectName string,
	methodName string,
	resolvingAliases map[*ast.Symbol]bool,
) bool {
	name, ok := staticEvaluator.builtinMethodName(
		node,
		objectName,
		resolvingAliases,
	)
	return ok && name == methodName
}
