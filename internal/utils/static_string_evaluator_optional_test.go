package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

type countingStaticReferenceResolver struct{ calls int }

func (resolver *countingStaticReferenceResolver) Resolve(*ast.Node) *ast.Symbol {
	resolver.calls++
	return nil
}

func TestStaticOptionalChainUnknownReceiverCost(t *testing.T) {
	for _, test := range []struct {
		suffix string
		links  int
	}{{"?.x", 1}, {"?.[0]", 1}, {"?.()", 1}, {"?.x?.[0]?.()", 3}} {
		for _, length := range []int{8, 16, 32, 64} {
			t.Run(fmt.Sprintf("%s/%d", test.suffix, length), func(t *testing.T) {
				code := "const value = unknown" + strings.Repeat(test.suffix, length)
				source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/chain.ts", Path: "/chain.ts"}, code, core.ScriptKindTS)
				resolver := &countingStaticReferenceResolver{}
				evaluator := NewStaticStringEvaluatorWithReferenceResolver(nil, source, resolver)
				if _, known := evaluator.EvalValue(findVariableInitializer(t, source, "value")); known {
					t.Fatal("unknown receiver must remain unknown")
				}
				if resolver.calls > 8*(test.links*length+1) {
					t.Fatalf("%d resolutions exceed linear bound for %d links", resolver.calls, length)
				}
				if evaluator.evaluation != nil {
					t.Fatal("evaluation memo must be discarded after the outer call")
				}
			})
		}
	}
}

func TestStaticOptionalJoinClassification(t *testing.T) {
	for _, test := range []struct {
		expression string
		want       string
		kind       StaticEvalValueKind
	}{
		{`null?.join()`, "", StaticEvalNonString},
		{`undefined?.join()`, "", StaticEvalNonString},
		{`({join: undefined}).join?.()`, "", StaticEvalNonString},
		{`["x"]?.join?.()`, "x", StaticEvalString},
		{`unknown?.join()`, "", StaticEvalUnknown},
		{`(null?.join).call()`, "", StaticEvalUnknown},
	} {
		t.Run(test.expression, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/chain.ts", Path: "/chain.ts"}, "const value = "+test.expression, core.ScriptKindTS)
			evaluator := NewStaticStringEvaluatorWithSourceFile(nil, source)
			got, kind := evaluator.EvalStringValue(findVariableInitializer(t, source, "value"))
			if got != test.want || kind != test.kind {
				t.Fatalf("EvalStringValue = (%q, %v), want (%q, %v)", got, kind, test.want, test.kind)
			}
		})
	}
}

func TestStaticOptionalChainEvaluationBoundaries(t *testing.T) {
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/chain.ts", Path: "/chain.ts"}, `
		const value = undefined?.x;
		const symbol = Symbol?.for("x");
	`, core.ScriptKindTS)
	evaluator := NewStaticStringEvaluatorWithSourceFile(nil, source)
	value := findVariableInitializer(t, source, "value")
	for _, access := range []GlobalAccess{GlobalAccessReadonly, GlobalAccessOff, GlobalAccessReadonly} {
		evaluator.GlobalAccess = func(string) GlobalAccess { return access }
		_, known := evaluator.EvalValue(value)
		if known != (access != GlobalAccessOff) {
			t.Fatalf("known = %v with global access %v", known, access)
		}
	}
	if result, known := evaluator.EvalValue(findVariableInitializer(t, source, "symbol")); !known {
		t.Fatal("optional builtin call must remain supported")
	} else if _, isSymbol := result.(staticSymbolValue); !isSymbol {
		t.Fatal("optional Symbol.for must preserve its typed result")
	}
}
