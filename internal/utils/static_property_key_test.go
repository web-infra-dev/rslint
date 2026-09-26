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
	for _, evaluator := range []*StaticStringEvaluator{nil, NewStaticStringEvaluatorWithoutScope()} {
		if evaluator.HasStaticAccessExpressionKey(member) || evaluator.HasStaticAccessExpressionKey(nil) {
			t.Fatal("unresolved key must remain unknown")
		}
	}
	if NewStaticStringEvaluator(nil).isStaticSymbolKey(nil) {
		t.Fatal("missing expression must remain unknown")
	}
}
