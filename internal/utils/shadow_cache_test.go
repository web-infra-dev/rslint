package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

// shadowCacheCorpus holds files whose scopes the cache has to key apart. Each
// one is asked about every name any of them binds, so a source that never
// mentions a name still exercises the "file declares nothing" fast path for it.
var shadowCacheCorpus = []string{
	// Nothing declares the name: the index alone has to answer, for every
	// node in the file.
	`Promise.resolve(1); function f() { return Promise.resolve(2); } const g = () => Promise.reject(3);`,

	// One scope shadows and a sibling does not, which is what a per-name
	// file-wide cache would get wrong.
	`function outer() { const Promise = 1; return Promise; }
Promise.resolve(1);
function other() { return Promise.resolve(2); }`,

	// Two references in the same block share one answer.
	`{ const Promise = 1; Promise; Promise; }
{ Promise; Promise; }`,

	// Hoisted declarations reach backwards within their scope.
	`function f() { Promise; var Promise; }
function g() { Promise; }
function h() { Promise; { var Promise; } }`,

	// A function's parameter environment is not its body's variable
	// environment, so entry has to be part of the key.
	`function f(a = Promise) { var Promise; return Promise; }`,
	`function f(a = Promise, Promise = 1) { return Promise; }`,

	// A parameter decorator escapes the decorated function's scope.
	`class C { m(@dec(Promise) x: number) { var Promise; } }`,
	`class C { m(@dec(Promise) x: number, Promise: any) { } }`,
	`class C { m(@dec(() => Promise) x: number, Promise: any) { } }`,

	// A class expression's own name is in scope for its body but not for a
	// scope created inside its decorators.
	`const k = class Promise { m() { return Promise; } };`,
	`class Promise { m(@dec(() => Promise) x: number) { } }`,

	// Namespace bodies, case blocks, catch clauses, enums, static blocks and
	// the three `for` heads are all scopes IsShadowed inspects.
	`namespace N { const Promise = 1; Promise; } Promise;`,
	`namespace N { var Promise = 1; { Promise; } } Promise;`,
	`switch (a) { case 1: const Promise = 1; case 2: Promise; } Promise;`,
	`try { Promise; } catch (Promise) { Promise; }`,
	`enum Promise { A = Promise.B } Promise;`,
	`class C { static { var Promise; Promise; } m() { return Promise; } }`,
	`for (var Promise = 0; Promise; Promise++) Promise;`,
	`for (const Promise in o) Promise;`,
	`for (let Promise of o) Promise;`,

	// Member names must stay out of the declaration index, or every file with
	// an unrelated `Promise` property would pay for the walk — and, more
	// importantly, must not be mistaken for bindings.
	`const helper = { Promise: 1, Promise() {} }; Promise.resolve(1);`,
	`class C { Promise = 1; Promise2() {} } Promise.resolve(1);`,
	`enum E { Promise } Promise.resolve(1);`,
	`interface I { Promise: number } Promise.resolve(1);`,
	`function f<Promise>() { return Promise; } Promise.resolve(1);`,

	// `undefined` travels the same path from the mock-factory suggestions.
	`function f() { const undefined = 1; return undefined; } undefined;`,
	`const o = { undefined: 1 }; undefined;`,

	// The source-file arm is now answered once per name for the whole file, so
	// each of its three ingredients needs a case: a top-level declaration, a
	// `var` hoisted out of a nested block, and a function declaration hoisted
	// out of one. Each file also asks from several scopes of its own.
	`const Promise = 1;
function f() { return Promise; }
function g() { { return Promise; } }
Promise;`,
	`{ var Promise = 1; }
function f() { return Promise; }
Promise;`,
	`if (b) function Promise() {}
function f() { return Promise; }
Promise;`,
	`switch (a) { case 1: { var Promise = 1; } }
function f() { return Promise; }
Promise;`,
	// The same shapes with nothing at file scope, so the arm has to answer
	// false while inner scopes still decide for themselves.
	`function f() { { var Promise = 1; } return Promise; }
function g() { return Promise; }
Promise;`,

	// Nested functions, so the walk crosses several scopes before the file.
	`function a() { function b() { function c() { return Promise; } } var Promise; }`,
	`function a() { const Promise = 1; return function b() { return function c() { return Promise; }; }; }`,
	`function unrelated(Promise) {}
function setup() { const a = () => Promise.resolve(1); const b = () => Promise.resolve(2); }`,
	`function f(a = (() => Promise)()) { var Promise; const b = () => Promise; const c = () => Promise; }`,
	`function outer() { { var Promise; } const a = () => Promise; const b = () => Promise; }
function sibling() { const a = () => Promise; const b = () => Promise; }`,
	`function outer() { const a = () => Promise; const b = () => Promise; function Promise() {} }`,
	`function unrelated(Promise) {}
namespace N { const a = () => Promise; const b = () => Promise; }
switch (x) { case 0: (() => Promise)(); break; default: (() => Promise)(); }
class C { static { (() => Promise)(); (() => Promise)(); } }`,
}

// shadowCacheNames are asked of every file in the corpus.
var shadowCacheNames = []string{"Promise", "undefined", "N", "C", "f", "a", "Promise2", "Missing"}

// TestShadowCacheMatchesIsShadowed is the differential the cache has to pass:
// for every identifier of every corpus file and every name, the cached answer
// and utils.IsShadowed must agree. Asking at every node — not only at the ones
// a rule would ask about — is what covers the scope and entry keying, since two
// nodes that share a cache entry must genuinely share an answer.
func TestShadowCacheMatchesIsShadowed(t *testing.T) {
	for _, code := range shadowCacheCorpus {
		t.Run(code, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, code, core.ScriptKindTS)

			identifiers := collectIdentifierNodes(sourceFile.AsNode())
			if len(identifiers) == 0 {
				t.Fatal("corpus entry has no identifiers")
			}

			for _, name := range shadowCacheNames {
				// One cache per name is not how a rule uses it, so share a
				// single cache across every name as well.
				cache := NewShadowCache(sourceFile)
				for _, node := range identifiers {
					want := IsShadowed(node, name)
					// Ask twice: the first call fills the entry, the second has
					// to read it back unchanged.
					for attempt := range 2 {
						if got := cache.IsShadowed(node, name); got != want {
							t.Fatalf("cache.IsShadowed(%q at %d, attempt %d) = %v, IsShadowed = %v",
								node.Text(), node.Pos(), attempt, got, want)
						}
					}
				}
			}
		})
	}
}

// TestShadowCacheDeclaresName locks the index down to bindings a scope walk can
// find, since a member name that leaked into it would take every file with an
// unrelated property of that name off the fast path.
func TestShadowCacheDeclaresName(t *testing.T) {
	tests := []struct {
		code     string
		declares bool
	}{
		{`const Promise = 1;`, true},
		{`let Promise;`, true},
		{`var Promise;`, true},
		{`function Promise() {}`, true},
		{`class Promise {}`, true},
		{`enum Promise {}`, true},
		{`namespace Promise {}`, true},
		{`import Promise from "m";`, true},
		{`import { x as Promise } from "m";`, true},
		{`import * as Promise from "m";`, true},
		{`function f(Promise) {}`, true},
		{`try {} catch (Promise) {}`, true},
		{`const { Promise } = o;`, true},
		{`const { x: Promise } = o;`, true},
		{`type Promise = number;`, true},
		{`interface Promise {}`, true},

		{`Promise.resolve(1);`, false},
		{`const o = { Promise: 1 };`, false},
		{`const o = { Promise() {} };`, false},
		{`const { Promise: renamed } = o;`, false},
		{`class C { Promise = 1; }`, false},
		{`class C { Promise() {} }`, false},
		{`class C { get Promise() { return 1; } }`, false},
		{`interface I { Promise: number }`, false},
		{`enum E { Promise }`, false},
		{`function f<Promise>() {}`, false},
		{`const jsx = <div Promise="x" />;`, false},
	}

	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.tsx",
				Path:     "/test.tsx",
			}, test.code, core.ScriptKindTSX)
			cache := NewShadowCache(sourceFile)
			if got := cache.DeclaresName("Promise"); got != test.declares {
				t.Errorf("DeclaresName() = %v, want %v", got, test.declares)
			}
		})
	}
}

func TestShadowCacheReusesOuterDeclarationScans(t *testing.T) {
	for _, test := range []struct {
		code string
		kind shadowScanKind
	}{
		{`function outer() { const a = () => Promise; const b = () => Promise; }`, shadowScanBlock},
		{`function outer() { const a = () => Promise; const b = () => Promise; }`, shadowScanHoistedVar},
		{`function outer(value) { const a = () => Promise; const b = () => Promise; }`, shadowScanParameters},
		{`namespace N { const a = () => Promise; const b = () => Promise; }`, shadowScanModuleBlock},
		{`switch (value) { case 1: const a = () => Promise; break; default: const b = () => Promise; }`, shadowScanCaseBlock},
		{`class C { static { const a = () => Promise; const b = () => Promise; } }`, shadowScanHoistedVar},
	} {
		t.Run(fmt.Sprintf("%d/%s", test.kind, test.code), func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts", Path: "/test.ts",
			}, "function unrelated(Promise) {}\n"+test.code, core.ScriptKindTS)
			var references []*ast.Node
			for _, node := range collectIdentifierNodes(source.AsNode()) {
				if node.Text() == "Promise" && node.Parent.Kind == ast.KindArrowFunction {
					references = append(references, node)
				}
			}
			if len(references) != 2 {
				t.Fatalf("got %d references, want 2", len(references))
			}
			cache := NewShadowCache(source)
			if cache.IsShadowed(references[0], "Promise") {
				t.Fatal("sibling parameter must not shadow Promise")
			}
			for outer := references[0].Parent.Parent; outer != nil; outer = outer.Parent {
				key := shadowScanKey{node: outer, name: "Promise", kind: test.kind}
				if result, ok := cache.scans[key]; ok {
					if result {
						t.Fatal("outer declaration scan must be negative")
					}
					// A sentinel proves the second sibling reads the cache instead of rescanning the AST.
					cache.scans[key] = true
					if !cache.IsShadowed(references[1], "Promise") {
						t.Fatal("second sibling did not reuse the outer declaration scan")
					}
					return
				}
			}
			t.Fatal("missing cached outer declaration scan")
		})
	}
}

func BenchmarkShadowCacheSiblingFactories(b *testing.B) {
	for _, wrapper := range []struct{ name, prefix, suffix string }{
		{"top-level", "", ""},
		{"describe", "describe('suite', () => {\n", "});"},
		{"setup", "function setup() {\n", "}"},
	} {
		for _, size := range []int{1000, 2000, 4000} {
			b.Run(fmt.Sprintf("%s/%d", wrapper.name, size), func(b *testing.B) {
				code := "function unrelated(Promise) {}\n" + wrapper.prefix +
					strings.Repeat("rs.doMock('m', () => Promise.resolve({}));\n", size) + wrapper.suffix
				source := parser.ParseSourceFile(ast.SourceFileParseOptions{
					FileName: "/bench.ts", Path: "/bench.ts",
				}, code, core.ScriptKindTS)
				var references []*ast.Node
				for _, node := range collectIdentifierNodes(source.AsNode()) {
					if node.Text() == "Promise" && node.Parent.Kind == ast.KindPropertyAccessExpression {
						references = append(references, node)
					}
				}
				if len(references) != size {
					b.Fatalf("got %d references, want %d", len(references), size)
				}
				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					cache := NewShadowCache(source)
					for _, reference := range references {
						if cache.IsShadowed(reference, "Promise") {
							b.Fatal("unrelated parameter must not shadow Promise")
						}
					}
				}
			})
		}
	}
}

func collectIdentifierNodes(root *ast.Node) []*ast.Node {
	var identifiers []*ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			identifiers = append(identifiers, node)
		}
		node.ForEachChild(visit)
		return false
	}
	root.ForEachChild(visit)
	return identifiers
}
