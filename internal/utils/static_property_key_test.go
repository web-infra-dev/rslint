package utils

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func TestStaticAccessExpressionKeys(t *testing.T) {
	for _, test := range []struct {
		code  string
		known bool
	}{
		{`object["foo"]`, true},
		{`object[Symbol.iterator]`, true},
		{`object[Symbol["iterator"]]`, true},
		{`object[Symbol.asyncDispose]`, true},
		{`object[Symbol.for("foo")]`, true},
		{`object[Symbol["for"]("foo")]`, true},
		{`object[Symbol.for()]`, true},
		{`object[Symbol.for(...["foo"])]`, true},
		{`object[Symbol.for(42)]`, true},
		{`object[Math.max(1, 2)]`, true},
		{`object[true ? Symbol.iterator : Symbol.match]`, true},
		{`object[Symbol.keyFor(Symbol.for("foo"))]`, true},
		{`object[Symbol.keyFor(Symbol.iterator)]`, true},
		{`object[Math.random()]`, false},
		{`object[Math.max(key)]`, false},
		{`object[Math.max(Symbol.iterator)]`, false},
		{`object[Symbol.for(Symbol.iterator)]`, false},
		{`object[Symbol.keyFor()]`, false},
		{`object[Symbol.unknown]`, false},
		{`object[Symbol[key]]`, false},
		{`object[Symbol.keyFor("foo")]`, false},
		{`object[Symbol.for(key)]`, false},
		{`object[Symbol.for(...keys)]`, false},
		{`object[Symbol("foo")]`, false},
		{`object[unknown()]`, false},
		{`object[key]`, false},
		{`object[other.iterator]`, false},
		{`function f(Symbol) { object[Symbol.iterator]; }`, false},
	} {
		t.Run(test.code, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/key.ts", Path: "/key.ts"}, test.code, core.ScriptKindTS)
			member := findFirstNodeOfKind(t, source, ast.KindElementAccessExpression)
			evaluator := NewStaticStringEvaluatorWithSourceFile(nil, source)
			if got := evaluator.HasStaticAccessExpressionKey(member); got != test.known {
				t.Fatalf("known key = %v, want %v", got, test.known)
			}
		})
	}
}

func TestStaticSymbolKeysRequireScope(t *testing.T) {
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/key.ts", Path: "/key.ts"}, `object[Symbol.iterator]`, core.ScriptKindTS)
	member := findFirstNodeOfKind(t, source, ast.KindElementAccessExpression)
	identifier := findFirstNodeOfKind(t, source, ast.KindIdentifier)
	for _, evaluator := range []*StaticStringEvaluator{nil, NewStaticStringEvaluatorWithoutScope()} {
		if evaluator.HasStaticAccessExpressionKey(member) || evaluator.HasStaticAccessExpressionKey(nil) {
			t.Fatal("unresolved key must remain unknown")
		}
	}
	evaluator := NewStaticStringEvaluatorWithSourceFile(nil, source)
	if evaluator.HasStaticAccessExpressionKey(identifier) {
		t.Fatal("non-member node must not be treated as a property key")
	}
	if evaluator.isBuiltinValue(nil, "String", map[*ast.Symbol]bool{}) {
		t.Fatal("missing built-in value must remain unresolved")
	}
}

func TestStaticSymbolValueSemantics(t *testing.T) {
	iterator := staticSymbolValue{kind: staticSymbolWellKnown, key: "iterator"}
	match := staticSymbolValue{kind: staticSymbolWellKnown, key: "match"}

	if staticValueIsTsgoSafe(iterator) {
		t.Fatal("static Symbol value must not escape into the tsgo evaluator")
	}
	if !staticValueIsTsgoSafe(struct{}{}) {
		t.Fatal("ordinary host values should remain tsgo-safe")
	}
	if staticValueKindOf(iterator) != staticKindSymbol {
		t.Fatal("static Symbol value has the wrong kind")
	}
	if truthy, ok := staticValueTruthy(iterator); !ok || !truthy {
		t.Fatal("Symbol values must be statically truthy")
	}
	if _, ok := staticValueToString(iterator); ok {
		t.Fatal("implicit Symbol string coercion must remain invalid")
	}
	if equal, ok := staticValuesStrictEqual(iterator, iterator); !ok || !equal {
		t.Fatal("identical well-known Symbols must compare equal")
	}
	if equal, ok := staticValuesStrictEqual(iterator, match); !ok || equal {
		t.Fatal("different well-known Symbols must compare unequal")
	}
}
