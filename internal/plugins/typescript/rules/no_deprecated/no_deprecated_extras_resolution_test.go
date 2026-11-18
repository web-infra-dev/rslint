package no_deprecated

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/utils"
	"strings"
	"testing"
)

// Expected diagnostics come from typescript-eslint v8.72.2 with TypeScript 6.0.2.
func TestNoDeprecatedReviewResolution(t *testing.T) {
	type diagnostic struct {
		id, message string
		start, end  int
	}
	for _, tc := range []struct {
		name, code string
		want       []diagnostic
	}{
		{name: "parameter-default", code: `/** @deprecated Use current instead. */
const legacy = 1;
function f(x = legacy) {}`, want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current instead.", 73, 79}}},
		{name: "arrow-default", code: `/** @deprecated old */
const legacy=1;
const f=(x=legacy)=>x;`, want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. old", 50, 56}}},
		{name: "constructor-default", code: `/** @deprecated old */
const legacy=1;
class C {constructor(public x=legacy){}}`, want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. old", 69, 75}}},
		{name: "destructured-default", code: `/** @deprecated old */
const legacy=1;
function f({x=legacy}={}){}`, want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. old", 53, 59}}},
		{name: "parameter-shadow", code: `/** @deprecated old */
const legacy=1;
function f(legacy=2){return legacy;}`, want: []diagnostic{}},
		{name: "parameter-call-default", code: `/** @deprecated old */
function legacy(){return 1;}
function f(x=legacy()){}`, want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. old", 65, 71}}},
		{name: "parameter-property-undocumented", code: `class Base {
/** @deprecated base */ x=0;
}
class Child extends Base {

constructor(public override x:number){super();}
}
new Child(1).x;`, want: []diagnostic{{"deprecatedWithReason", "`x` is deprecated. base", 135, 136}}},
		{name: "parameter-property-named-tag", code: `class Base {
/** @deprecated base */ x=0;
}
class Child extends Base {
/** @param x local */
constructor(public override x:number){super();}
}
new Child(1).x;`, want: []diagnostic{}},
		{name: "parameter-property-unrelated-tag", code: `class Base {
/** @deprecated base */ x=0;
}
class Child extends Base {
/** @param other unrelated */
constructor(public override x:number){super();}
}
new Child(1).x;`, want: []diagnostic{{"deprecatedWithReason", "`x` is deprecated. base", 164, 165}}},
		{name: "parameter-property-empty-tag", code: `class Base {
/** @deprecated base */ x=0;
}
class Child extends Base {
/** @param x */
constructor(public override x:number){super();}
}
new Child(1).x;`, want: []diagnostic{}},
		{name: "parameter-property-typed-tag", code: `class Base {
/** @deprecated base */ x=0;
}
class Child extends Base {
/** @param {number} x local */
constructor(public override x:number){super();}
}
new Child(1).x;`, want: []diagnostic{}},
		{name: "field-initializer-undocumented", code: `class Base {
/** @deprecated base */ fn=()=>1;
}
class Child extends Base {fn= ()=>2;}
new Child().fn;
new Child().fn();`, want: []diagnostic{{"deprecatedWithReason", "`fn` is deprecated. base", 99, 101}, {"deprecatedWithReason", "`fn` is deprecated. base", 115, 117}}},
		{name: "field-initializer-return-tag", code: `class Base {
/** @deprecated base */ fn=()=>1;
}
class Child extends Base {fn=/** @returns local */ ()=>2;}
new Child().fn;
new Child().fn();`, want: []diagnostic{}},
		{name: "field-initializer-deprecated-tag", code: `class Base {
/** @deprecated base */ fn=()=>1;
}
class Child extends Base {fn=/** @deprecated initializer */ ()=>2;}
new Child().fn;
new Child().fn();`, want: []diagnostic{{"deprecatedWithReason", "`fn` is deprecated. initializer", 129, 131}, {"deprecatedWithReason", "`fn` is deprecated. initializer", 145, 147}}},
		{name: "field-initializer-inherit-tag", code: `class Base {
/** @deprecated base */ fn=()=>1;
}
class Child extends Base {fn=/** @inheritDoc */ ()=>2;}
new Child().fn;
new Child().fn();`, want: []diagnostic{{"deprecatedWithReason", "`fn` is deprecated. base", 117, 119}, {"deprecatedWithReason", "`fn` is deprecated. base", 133, 135}}},
		{name: "initializer-variable-reference", code: `const legacy=/** @deprecated initializer */ ()=>1;
legacy;
legacy();`, want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. initializer", 51, 57}, {"deprecatedWithReason", "`legacy` is deprecated. initializer", 59, 65}}},
		{name: "getter-first-undocumented", code: `class Base {
/** @returns value */ get foo(){return 1;}
/** @deprecated setter */ set foo(value:number){}
}
class Child extends Base {

get foo(){return 2;}
set foo(value:number){}
}
const child=new Child();
child.foo;
child.foo=3;`, want: []diagnostic{}},
		{name: "getter-first-inherit-tag", code: `class Base {
/** @returns value */ get foo(){return 1;}
/** @deprecated setter */ set foo(value:number){}
}
class Child extends Base {
/** @inheritDoc */
get foo(){return 2;}
set foo(value:number){}
}
const child=new Child();
child.foo;
child.foo=3;`, want: []diagnostic{}},
		{name: "getter-first-local-tag", code: `class Base {
/** @returns value */ get foo(){return 1;}
/** @deprecated setter */ set foo(value:number){}
}
class Child extends Base {
/** @returns local */
get foo(){return 2;}
set foo(value:number){}
}
const child=new Child();
child.foo;
child.foo=3;`, want: []diagnostic{}},
		{name: "setter-first-undocumented", code: `class Base {
/** @deprecated setter */ set foo(value:number){}
/** @returns value */ get foo(){return 1;}
}
class Child extends Base {

set foo(value:number){}
get foo(){return 2;}
}
const child=new Child();
child.foo;
child.foo=3;`, want: []diagnostic{{"deprecatedWithReason", "`foo` is deprecated. setter", 214, 217}, {"deprecatedWithReason", "`foo` is deprecated. setter", 225, 228}}},
		{name: "setter-first-inherit-tag", code: `class Base {
/** @deprecated setter */ set foo(value:number){}
/** @returns value */ get foo(){return 1;}
}
class Child extends Base {
/** @inheritDoc */
set foo(value:number){}
get foo(){return 2;}
}
const child=new Child();
child.foo;
child.foo=3;`, want: []diagnostic{{"deprecatedWithReason", "`foo` is deprecated. setter", 232, 235}, {"deprecatedWithReason", "`foo` is deprecated. setter", 243, 246}}},
		{name: "setter-first-local-tag", code: `class Base {
/** @deprecated setter */ set foo(value:number){}
/** @returns value */ get foo(){return 1;}
}
class Child extends Base {
/** @returns local */
set foo(value:number){}
get foo(){return 2;}
}
const child=new Child();
child.foo;
child.foo=3;`, want: []diagnostic{}},
		{name: "getter-fallback-setter", code: `class Base {
get foo(){return 1;}
/** @deprecated setter */ set foo(value:number){}
}
class Child extends Base {get foo(){return 2;}}
new Child().foo;`, want: []diagnostic{{"deprecatedWithReason", "`foo` is deprecated. setter", 146, 149}}},
		{name: "getter-only-deep-empty", code: `class C0 {get value(){return 1;}}
class C1 extends C0 {get value(){return 1;}}
class C2 extends C1 {get value(){return 1;}}
class C3 extends C2 {get value(){return 1;}}
new C3().value;`, want: []diagnostic{}},
		{name: "cycle", code: `class C implements C {get value(){return 1;}}
new C().value;`, want: []diagnostic{}},
		{name: "interface-diamond", code: `interface A {/** @deprecated base */ value:number;}
interface B extends A {value:number;}
interface C extends A {value:number;}
interface D extends B,C {value:number;}
declare const d:D; d.value;`, want: []diagnostic{}},
		{name: "static-getter", code: `class A {
/** @deprecated base */ static get value(){return 1;}
}
class B extends A {static get value(){return 2;}}
B.value;`, want: []diagnostic{{"deprecatedWithReason", "`value` is deprecated. base", 118, 123}}},
		{name: "object-parenthesized-key", code: `interface Options {
/** @deprecated old */ old:number;
}
const opts:Options={[('old')]:1};`, want: []diagnostic{{"deprecatedWithReason", "`old` is deprecated. old", 79, 84}}},
		{name: "element-parenthesized-key", code: `interface Options {
/** @deprecated old */ old:number;
}
declare const opts:Options;
opts[('old')];`, want: []diagnostic{{"deprecatedWithReason", "`old` is deprecated. old", 91, 96}}},
		{name: "object-nested-parentheses", code: `interface Options {
/** @deprecated old */ old:number;
}
const opts:Options={[((('old')))]:1};`, want: []diagnostic{{"deprecatedWithReason", "`old` is deprecated. old", 81, 86}}},
		{name: "element-nested-parentheses", code: `interface Options {
/** @deprecated old */ old:number;
}
declare const opts:Options;
opts[((('old')))];`, want: []diagnostic{{"deprecatedWithReason", "`old` is deprecated. old", 93, 98}}},
		{name: "object-parenthesized-assertion", code: `interface Options {
/** @deprecated old */ old:number;
}
const opts:Options={[('old' as const)]:1};`, want: []diagnostic{{"deprecatedWithReason", "`old` is deprecated. old", 79, 93}}},
		{name: "element-parenthesized-assertion", code: `interface Options {
/** @deprecated old */ old:number;
}
declare const opts:Options;
opts[('old' as const)];`, want: []diagnostic{{"deprecatedWithReason", "`old` is deprecated. old", 91, 105}}},
		{name: "object-parenthesized-satisfies", code: `interface Options {
/** @deprecated old */ old:number;
}
const opts:Options={[(('old' satisfies string))]:1};`, want: []diagnostic{{"deprecatedWithReason", "`old` is deprecated. old", 80, 102}}},
		{name: "element-parenthesized-satisfies", code: `interface Options {
/** @deprecated old */ old:number;
}
declare const opts:Options;
opts[(('old' satisfies string))];`, want: []diagnostic{{"deprecatedWithReason", "`old` is deprecated. old", 92, 114}}},
		{name: "typedef-without-return", code: `/** @typedef {number} NewType
 * @deprecated old
 */
const legacy=1; legacy;`, want: []diagnostic{}},
		{name: "typedef-with-return", code: `/** @typedef {number} NewType
 * @deprecated old
 * @returns value
 */
const legacy=1; legacy;`, want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. old", 87, 93}}},
		{name: "callback-without-return", code: `/** @callback {number} NewType
 * @deprecated old
 */
const legacy=1; legacy;`, want: []diagnostic{}},
		{name: "callback-with-return", code: `/** @callback {number} NewType
 * @deprecated old
 * @returns value
 */
const legacy=1; legacy;`, want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. old", 88, 94}}},
		{name: "mutual-implements-empty", code: `class A implements B {get value(){return 1;}}
class B implements A {get value(){return 1;}}
declare const a:A;declare const b:B; a.value; b.value;`, want: []diagnostic{}},
		{name: "mutual-implements-tagged", code: `class A implements B {
/** @deprecated base */ get value(){return 1;}
}
class B implements A {get value(){return 1;}}
declare const a:A;declare const b:B; b.value; a.value;`, want: []diagnostic{{"deprecatedWithReason", "`value` is deprecated. base", 157, 162}, {"deprecatedWithReason", "`value` is deprecated. base", 166, 171}}},
		{name: "expression-getter-inheritance", code: `class Base {
/** @deprecated base */ get value(){return 1;}
}
const C=class extends Base {get value(){return 1;}};new C().value;`, want: []diagnostic{{"deprecatedWithReason", "`value` is deprecated. base", 122, 127}}},
		{name: "parameter-reference-with-constructor-tags", code: `class Base {
/** @deprecated base */ x=0;
}
class Child extends Base {
/** @param x local */
constructor(public override x:number){super(); x;}
}
new Child(1).x;`, want: []diagnostic{}},
		{name: "initializer-before-field-tags", code: `class Base {
/** @deprecated base */ fn=()=>1;
}
class Child extends Base {
/** @deprecated field */ fn=/** @deprecated initializer */ ()=>2;
}
const c=new Child();c.fn;c.fn();`, want: []diagnostic{{"deprecatedWithReason", "`fn` is deprecated. initializer", 166, 168}, {"deprecatedWithReason", "`fn` is deprecated. initializer", 171, 173}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.tsx": tc.code}, "main.tsx", nil)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d diagnostics, want %d: %#v", len(got), len(tc.want), got)
			}
			for i, w := range tc.want {
				d := got[i]
				if d.Message.Id != w.id || d.Message.Description != w.message || d.Range.Pos() != w.start || d.Range.End() != w.end {
					t.Errorf("got %s %q [%d,%d), want %s %q [%d,%d)", d.Message.Id, d.Message.Description, d.Range.Pos(), d.Range.End(), w.id, w.message, w.start, w.end)
				}
				if d.FixesPtr != nil || d.Suggestions != nil {
					t.Error("unexpected fix or suggestion")
				}
			}
		})
	}
}

func newReviewResolver(t *testing.T, code string) (*deprecationResolver, *ast.SourceFile) {
	t.Helper()
	root := fixtures.GetRootDir()
	filename := tspath.ResolvePath(root.Dir, "main.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{filename: code})
	program, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", utils.CreateCompilerHost(root.Dir, fs))
	if err != nil {
		t.Fatal(err)
	}
	source := program.GetSourceFile(filename)
	if source == nil {
		t.Fatal("missing test source")
	}
	c, done := program.GetTypeCheckerForFile(t.Context(), source)
	t.Cleanup(done)
	return newDeprecationResolver(c), source
}

func TestNoDeprecatedResolverWork(t *testing.T) {
	for _, depth := range []int{1, 4, 8, 16, 32} {
		for _, setters := range []bool{false, true} {
			t.Run(fmt.Sprintf("depth=%d/setters=%t", depth, setters), func(t *testing.T) {
				var code strings.Builder
				for i := range depth {
					fmt.Fprintf(&code, "class C%d", i)
					if i > 0 {
						fmt.Fprintf(&code, " extends C%d", i-1)
					}
					code.WriteString(" { get value(){return 1;} ")
					if setters {
						code.WriteString("set value(v:number){}")
					}
					code.WriteString("}\n")
				}
				fmt.Fprintf(&code, "declare const target:C%d; target.value;", depth-1)
				r, source := newReviewResolver(t, code.String())
				access := source.Statements.Nodes[len(source.Statements.Nodes)-1].AsExpressionStatement().Expression
				symbol := r.checker.GetSymbolAtLocation(access.Name())
				if symbol == nil {
					t.Fatal("missing property symbol")
				}
				reads := 0
				bound := depth
				if setters {
					bound = 4 * depth
				}
				readOwn := r.readOwn
				r.readOwn = func(node *ast.Node) deprecationInfo {
					reads++
					if reads > bound {
						t.Fatalf("%d reads exceed linear bound %d", reads, bound)
					}
					return readOwn(node)
				}
				modes := []ast.Kind{ast.KindUnknown, ast.KindGetAccessor}
				if setters {
					modes = append(modes, ast.KindSetAccessor)
				}
				for _, mode := range modes {
					if info := r.symbolTags(symbol, mode); info.hasTags || info.deprecated {
						t.Fatalf("unexpected tags: %#v", info)
					}
				}
				if reads < depth {
					t.Fatalf("only %d reads for %d reachable declarations", reads, depth)
				}
				initial := reads
				for range 100 {
					for _, mode := range modes {
						r.symbolTags(symbol, mode)
					}
				}
				if reads != initial {
					t.Fatalf("repeated negative lookups did %d extra reads", reads-initial)
				}
				for _, entry := range r.results {
					if !entry.complete {
						t.Fatal("unfinished cache entry after lookup")
					}
				}
				t.Logf("depth %d: %d declaration reads; repeated lookups: 0", depth, reads)
			})
		}
	}
}

func TestNoDeprecatedResolverContexts(t *testing.T) {
	const code = `class Base {
/** @returns value */ get value(){return 1;}
/** @deprecated setter */ set value(v:number){}
}
class Child extends Base {get value(){return 2;} set value(v:number){}}
declare const child:Child; child.value;`
	for _, order := range [][]ast.Kind{
		{ast.KindUnknown, ast.KindGetAccessor, ast.KindSetAccessor},
		{ast.KindSetAccessor, ast.KindGetAccessor, ast.KindUnknown},
	} {
		r, source := newReviewResolver(t, code)
		access := source.Statements.Nodes[len(source.Statements.Nodes)-1].AsExpressionStatement().Expression
		symbol := r.checker.GetSymbolAtLocation(access.Name())
		for _, mode := range order {
			info := r.symbolTags(symbol, mode)
			want := mode == ast.KindSetAccessor
			if info.deprecated != want || (want && info.reason != "setter") {
				t.Fatalf("mode %v: %#v", mode, info)
			}
		}
	}
	const overloads = `/** @deprecated zero */ function f(x:0):void;
function f(x:1):void;
function f(x:number){}
f;`
	r, source := newReviewResolver(t, overloads)
	for _, i := range []int{1, 0, 1} {
		deprecated, reason := r.declarationDeprecation(source.Statements.Nodes[i])
		if deprecated != (i == 0) || (i == 0 && reason != "zero") {
			t.Fatalf("signature %d: %t %q", i, deprecated, reason)
		}
	}
	// A new resolver must compute its own results even with the same checker.
	other := newDeprecationResolver(r.checker)
	reads := 0
	readOwn := other.readOwn
	other.readOwn = func(node *ast.Node) deprecationInfo { reads++; return readOwn(node) }
	other.declarationDeprecation(source.Statements.Nodes[0])
	if reads == 0 {
		t.Fatal("results escaped their resolver")
	}
}

// Tag ownership is independently pinned to TypeScript 6.0.2 getJSDocTags.
func TestNoDeprecatedLocalTagOwnership(t *testing.T) {
	type tag struct {
		kind ast.Kind
		name string
	}
	for _, tc := range []struct {
		name, code string
		kind       ast.Kind
		index      int
		tags       []tag
	}{
		{name: "named-parameter", code: `/** @param left L
 * @param right R */
function f(left:number,right:number){}`, kind: ast.KindParameter, index: 1, tags: []tag{{ast.KindJSDocParameterTag, "right"}}},
		{name: "destructured-parameter", code: `/** @param first A
 * @param second B */
function f({x},[y]){}`, kind: ast.KindParameter, index: 1, tags: []tag{{ast.KindJSDocParameterTag, "second"}}},
		{name: "rest-parameter", code: `/** @param args values */
function f(...args:number[]){}`, kind: ast.KindParameter, index: 0, tags: []tag{{ast.KindJSDocParameterTag, "args"}}},
		{name: "unmatched-parameter", code: `/** @param absent values */
function f(value:number){}`, kind: ast.KindParameter, index: 0, tags: []tag{}},
		{name: "template-parameter", code: `/** @template T,U */
function f<T,U>(){}`, kind: ast.KindTypeParameter, index: 1, tags: []tag{{ast.KindJSDocTemplateTag, ""}}},
		{name: "initializer-type-cast", code: `const x=/** @type {number} */ (1);`, kind: ast.KindVariableDeclaration, index: 0, tags: []tag{}},
		{name: "initializer-return-tag", code: `const x=/** @returns number */ ()=>1;`, kind: ast.KindVariableDeclaration, index: 0, tags: []tag{{ast.KindJSDocReturnTag, ""}}},
		{name: "last-documentation-block", code: `/** @deprecated old */
/** @returns value */
function f(){}`, kind: ast.KindFunctionDeclaration, index: 0, tags: []tag{{ast.KindJSDocReturnTag, ""}}},
		{name: "earlier-overload-block", code: `/** @overload */
/** @returns value */
function f(){}`, kind: ast.KindFunctionDeclaration, index: 0, tags: []tag{{ast.KindJSDocOverloadTag, ""}, {ast.KindJSDocReturnTag, ""}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, source := newReviewResolver(t, tc.code)
			var found *ast.Node
			index := tc.index
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if node.Kind == tc.kind {
					if index == 0 {
						found = node
						return true
					}
					index--
				}
				return node.ForEachChild(visit)
			}
			visit(source.AsNode())
			if found == nil {
				t.Fatal("missing declaration")
			}
			got := localJSDocTags(found)
			if len(got) != len(tc.tags) {
				t.Fatalf("got %d tags, want %d", len(got), len(tc.tags))
			}
			for i, want := range tc.tags {
				name := ""
				if got[i].Kind == ast.KindJSDocParameterTag {
					name = got[i].Name().Text()
				}
				if got[i].Kind != want.kind || name != want.name {
					t.Errorf("tag %d: %v %q, want %v %q", i, got[i].Kind, name, want.kind, want.name)
				}
			}
		})
	}
}
