// TestBuild* lock in the shape of the scope tree the shared model produces —
// which scopes exist, what nests inside what, and which scope owns each
// binding. Rules read those three things and nothing else, so a regression
// here is a regression in every consumer.
package scope

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

func build(t *testing.T, source string) *Manager {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)
	return Build(sourceFile, Options{})
}

// lookup finds the innermost scope declaring `name` starting from the scope
// created at index `scopeIndex`, mirroring how a rule resolves outward.
func lookup(m *Manager, scopeIndex int, name string) *Variable {
	for current := m.Scopes[scopeIndex]; current != nil; current = current.Parent {
		if declarations := current.Declarations(name); len(declarations) > 0 {
			return declarations[0]
		}
	}
	return nil
}

func assertKinds(t *testing.T, m *Manager, want []Kind) {
	t.Helper()
	if len(m.Scopes) != len(want) {
		t.Fatalf("got %d scopes, want %d", len(m.Scopes), len(want))
	}
	for i, kind := range want {
		if m.Scopes[i].Kind != kind {
			t.Errorf("scope %d has kind %d, want %d", i, m.Scopes[i].Kind, kind)
		}
	}
}

func assertDeclares(t *testing.T, s *Scope, name string, kind DefKind) {
	t.Helper()
	declarations := s.Declarations(name)
	if len(declarations) == 0 {
		t.Fatalf("scope does not declare %q", name)
	}
	if declarations[0].Kind != kind {
		t.Errorf("%q has def kind %d, want %d", name, declarations[0].Kind, kind)
	}
}

func TestBuildHoistsVarToTheVariableScope(t *testing.T) {
	m := build(t, `
function outer() {
  if (cond) { var hoisted = 1; let blockOnly = 2; }
}
`)
	assertKinds(t, m, []Kind{KindGlobal, KindFunction, KindBlock})

	assertDeclares(t, m.Scopes[0], "outer", DefFunctionName)
	// `var` lands in the function scope even though it is written in a block.
	assertDeclares(t, m.Scopes[1], "hoisted", DefVariable)
	if got := len(m.Scopes[1].Declarations("hoisted")); got != 1 {
		t.Fatalf("function scope has %d definitions of hoisted, want 1", got)
	}
	assertDeclares(t, m.Scopes[2], "blockOnly", DefVariable)
	if len(m.Scopes[2].Declarations("hoisted")) != 0 {
		t.Error("var declaration should not also live in the block scope")
	}
	if m.Scopes[2].VariableScope() != m.Scopes[1] {
		t.Error("block scope's variable scope should be the enclosing function")
	}
}

func TestBuildHoistsFunctionsThroughScopeLessStatements(t *testing.T) {
	m := build(t, `
function outer() {
  if (condition) function fromIf() {}
  label: function fromLabel() {}
  while (condition) function fromWhile() {}
  do function fromDo() {} while (condition);
  for (; condition;) function fromFor() {}
  for (let index = 0; condition;) function inFor() {}
  if (condition) { function blockOnly() {} }
}
`)
	outer := m.Scopes[1]
	for _, name := range []string{"fromIf", "fromLabel", "fromWhile", "fromDo", "fromFor"} {
		assertDeclares(t, outer, name, DefFunctionName)
	}
	if len(outer.Declarations("inFor")) != 0 {
		t.Error("a let-initialized for loop should keep its function binding in the loop scope")
	}
	if len(outer.Declarations("blockOnly")) != 0 {
		t.Error("a braced body should keep its function binding in the block scope")
	}

	var forScope, blockScope *Scope
	for _, candidate := range m.Scopes {
		if candidate.Block == nil {
			continue
		}
		switch candidate.Block.Kind {
		case ast.KindForStatement:
			if len(candidate.Declarations("inFor")) != 0 {
				forScope = candidate
			}
		case ast.KindBlock:
			if len(candidate.Declarations("blockOnly")) != 0 {
				blockScope = candidate
			}
		}
	}
	if forScope == nil {
		t.Error("let-initialized loop scope does not declare inFor")
	}
	if blockScope == nil {
		t.Error("braced block scope does not declare blockOnly")
	}
}

func TestBuildDottedNamespaceSegmentsDoNotDeclareVariables(t *testing.T) {
	m := build(t, `
let A = [];
namespace A.B { export const value = 1; }
`)
	declarations := m.Global.Declarations("A")
	if len(declarations) != 1 || declarations[0].Kind != DefVariable {
		t.Fatalf("global A definitions = %#v, want only the variable", declarations)
	}
	for _, candidate := range m.Scopes {
		if len(candidate.Declarations("B")) != 0 {
			t.Fatalf("scope %v unexpectedly declares dotted namespace segment B", candidate.Kind)
		}
	}
	assertKinds(t, m, []Kind{KindGlobal, KindModule})
	assertDeclares(t, m.Scopes[1], "value", DefVariable)
}

func TestBuildKeepsExplicitNestedNamespaces(t *testing.T) {
	m := build(t, `namespace A { namespace B { namespace C { const value = 1; } } }`)
	assertKinds(t, m, []Kind{KindGlobal, KindModule, KindModule, KindModule})
	assertDeclares(t, m.Scopes[1], "B", DefNamespaceName)
	assertDeclares(t, m.Scopes[2], "C", DefNamespaceName)
	assertDeclares(t, m.Scopes[3], "value", DefVariable)
}

func TestBuildWithBodyHasItsOwnEnvironment(t *testing.T) {
	m := build(t, `with (object) { function inner() {} }`)
	assertKinds(t, m, []Kind{KindGlobal, KindWith, KindBlock, KindFunction})
	assertDeclares(t, m.Scopes[2], "inner", DefFunctionName)
	if m.Scopes[1].VariableScope() != m.Global {
		t.Fatal("with must not introduce a var hoisting target")
	}
}

func TestAcquireUsesTheESTreePosition(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		code    string
		kind    Kind
		binding string
	}{
		{"parameter default", `function f(x = probe()) { const local = 1; }`, KindFunction, "local"},
		{"named function", `const f = function named(local) { probe(); };`, KindFunction, "local"},
		{"method key", `class C<Local> { [probe()](Local) {} }`, KindClass, "Local"},
		{"method decorator", `class C<Local> { @probe() method(Local) {} }`, KindClass, "Local"},
		{"parameter decorator", `class C { method(@probe() x, local) {} }`, KindFunction, "local"},
		{"class decorator", `@probe() class C<Local> {}`, KindClass, "Local"},
		{"field initializer", `class C { field = probe(); }`, KindClassFieldInitializer, ""},
		{"switch discriminant", `switch (probe()) { case 0: const local = 1; }`, KindBlock, "local"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			m := build(t, testCase.code)
			var call *ast.Node
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindCallExpression && node.AsCallExpression().Expression.Kind == ast.KindIdentifier && node.AsCallExpression().Expression.Text() == "probe" {
					call = node
				}
				return node.ForEachChild(visit)
			}
			m.SourceFile.AsNode().ForEachChild(visit)
			if call == nil {
				t.Fatal("missing probe call")
			}
			acquired := m.Acquire(call)
			if acquired.Kind != testCase.kind {
				t.Fatalf("kind = %v, want %v", acquired.Kind, testCase.kind)
			}
			if testCase.binding != "" && len(acquired.Declarations(testCase.binding)) == 0 {
				t.Fatalf("missing declaration %q", testCase.binding)
			}
			if m.Acquire(call) != acquired {
				t.Fatal("cached acquisition changed")
			}
		})
	}
}

func TestBuildFunctionExpressionNameGetsItsOwnScope(t *testing.T) {
	m := build(t, `var alias = function named(param) { return named; };`)
	assertKinds(t, m, []Kind{KindGlobal, KindFunctionExprName, KindFunction})

	assertDeclares(t, m.Scopes[0], "alias", DefVariable)
	assertDeclares(t, m.Scopes[1], "named", DefFnExprName)
	assertDeclares(t, m.Scopes[2], "param", DefParameter)
	// The name is reachable from the body but not from the enclosing scope.
	if lookup(m, 2, "named") == nil {
		t.Error("function body should see the function expression's own name")
	}
	if len(m.Scopes[0].Declarations("named")) != 0 {
		t.Error("function expression name must not leak to the enclosing scope")
	}
}

func TestBuildClassScopeHoldsInnerNameAndTypeParameters(t *testing.T) {
	m := build(t, `class Box<T> extends Base { method(value: T) {} }`)
	assertKinds(t, m, []Kind{KindGlobal, KindClass, KindFunction})

	assertDeclares(t, m.Scopes[0], "Box", DefClassName)
	assertDeclares(t, m.Scopes[1], "Box", DefClassInnerName)
	assertDeclares(t, m.Scopes[1], "T", DefTypeParameter)
	assertDeclares(t, m.Scopes[2], "value", DefParameter)
	if lookup(m, 2, "T") == nil {
		t.Error("method scope should see the class type parameter")
	}
}

func TestBuildCatchClauseNestsABlockScope(t *testing.T) {
	m := build(t, `try {} catch (err) { let inner = err; }`)
	// The try block, the catch clause, and the catch body are separate scopes.
	assertKinds(t, m, []Kind{KindGlobal, KindBlock, KindCatch, KindBlock})

	assertDeclares(t, m.Scopes[2], "err", DefCatch)
	assertDeclares(t, m.Scopes[3], "inner", DefVariable)
	if lookup(m, 3, "err") == nil {
		t.Error("catch body should resolve the catch parameter")
	}
}

func TestBuildTypeScriptDeclarationKinds(t *testing.T) {
	m := build(t, `
import type { Imported } from './m';
interface Shape {}
type Alias = Shape;
enum Level { Low }
namespace Space {}
`)
	global := m.Scopes[0]
	assertDeclares(t, global, "Imported", DefImport)
	assertDeclares(t, global, "Shape", DefType)
	assertDeclares(t, global, "Alias", DefType)
	assertDeclares(t, global, "Level", DefEnumName)
	assertDeclares(t, global, "Space", DefNamespaceName)

	if !global.Declarations("Imported")[0].IsTypeOnlyImport {
		t.Error("`import type` binding should be flagged type-only")
	}
	if global.Declarations("Shape")[0].IsValueBinding {
		t.Error("an interface should not be a value binding")
	}
}

func TestBuildGlobalAugmentationIsMarked(t *testing.T) {
	m := build(t, `declare global { var injected: string; }`)
	assertKinds(t, m, []Kind{KindGlobal, KindModule})
	if m.Scopes[0].GlobalAugmentation {
		t.Error("the file scope is not an augmentation")
	}
	if !m.Scopes[1].GlobalAugmentation {
		t.Error("`declare global` scope should be marked as an augmentation")
	}
}

func TestBuildMergesDeclarationsByNameInOrder(t *testing.T) {
	m := build(t, `
enum Merged { A }
namespace Merged {}
`)
	declarations := m.Scopes[0].Declarations("Merged")
	if len(declarations) != 2 {
		t.Fatalf("got %d declarations of Merged, want 2", len(declarations))
	}
	if declarations[0].Kind != DefEnumName || declarations[1].Kind != DefNamespaceName {
		t.Errorf("declarations should be returned in source order, got %d then %d",
			declarations[0].Kind, declarations[1].Kind)
	}
}

func TestBuildEmptySourceFileHasOnlyTheGlobalScope(t *testing.T) {
	m := build(t, ``)
	assertKinds(t, m, []Kind{KindGlobal})
	if m.Global != m.Scopes[0] {
		t.Error("Global should be the first created scope")
	}
	if m.Global.VariableScope() != m.Global {
		t.Error("the global scope is its own variable scope")
	}
}

func buildWithReferences(t *testing.T, source string) *Manager {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.tsx",
		Path:     "/test.tsx",
	}, source, core.ScriptKindTSX)
	return Build(sourceFile, Options{CollectReferences: true})
}

// referencedNames lists every reference in source order as
// "name->resolvedDeclarationKind", using "?" when the name resolves to nothing
// in this file.
func referencedNames(m *Manager) []string {
	out := make([]string, 0, len(m.References))
	for _, ref := range m.References {
		name := ref.Identifier.Text()
		if resolved := ref.Resolved(); resolved != nil {
			out = append(out, name+"->declared")
		} else {
			out = append(out, name+"->?")
		}
	}
	return out
}

func assertReferences(t *testing.T, source string, want ...string) {
	t.Helper()
	got := referencedNames(buildWithReferences(t, source))
	if len(got) != len(want) {
		t.Fatalf("%s\n got %v\nwant %v", source, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s\n got %v\nwant %v", source, got, want)
			return
		}
	}
}

func TestBuildCollectsReferencesOnlyWhenAsked(t *testing.T) {
	if refs := build(t, `var a = 1; a;`).References; len(refs) != 0 {
		t.Errorf("references collected without Options.CollectReferences: %d", len(refs))
	}
	if refs := buildWithReferences(t, `var a = 1; a;`).References; len(refs) != 1 {
		t.Errorf("got %d references, want 1", len(refs))
	}
}

func TestBuildFiltersCollectedReferencesWithoutChangingResolution(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, `const kept = 1, skipped = 2; kept; skipped; missing;`, core.ScriptKindTS)
	m := Build(sourceFile, Options{
		CollectReferences: true,
		ReferenceNames:    map[string]struct{}{"kept": {}, "missing": {}},
	})
	got := referencedNames(m)
	want := []string{"kept->declared", "missing->?"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestBuildReferenceIdentifierPositions(t *testing.T) {
	// Declaration names are not references; eslint-scope models them as
	// `init: true` references that every consumer filters out.
	assertReferences(t, `var a = b;`, "b->?")
	// Only the object of a member access, never the member name.
	assertReferences(t, `a.b.c;`, "a->?")
	assertReferences(t, `a[b];`, "a->?", "b->?")
	// Object literal keys are not bindings, but shorthand values are.
	assertReferences(t, `({ a: 1 });`)
	assertReferences(t, `({ a });`, "a->?")
	assertReferences(t, `({ [a]: 1 });`, "a->?")
	// Labels are not bindings.
	assertReferences(t, `outer: while (x) { break outer; }`, "x->?")
	// Import specifiers declare; export specifiers reference their local name.
	assertReferences(t, `import { a as b } from './m';`)
	assertReferences(t, `const a = 1; export { a as b };`, "a->declared")
	// Type positions reference too.
	assertReferences(t, `let x: Foo;`, "Foo->?")
	assertReferences(t, `let x: typeof Foo.Bar;`, "Foo->?")
	// ...but an `import(...)` type names another module's exports.
	assertReferences(t, `type T = typeof import('./m').a.b;`)
	assertReferences(t, `export as namespace Missing;`, "Missing->?")
	assertReferences(t, `const Present = {}; export as namespace Present;`, "Present->declared")
	assertReferences(t, `interface TypeOnly {}; export as namespace TypeOnly;`, "TypeOnly->?")
	// JSX: a lower-case tag is an intrinsic element, not a binding.
	assertReferences(t, `<div />;`)
	assertReferences(t, `<App />;`, "App->?")
	assertReferences(t, `<ns.Widget />;`, "ns->?")
	assertReferences(t, `<App></App>;`, "App->?", "App->?")
	assertReferences(t, `<ns.Widget></ns.Widget>;`, "ns->?", "ns->?")
	// TSESTree visits both pieces of a namespaced TSX tag, but never the
	// pieces of a namespaced attribute.
	assertReferences(t, `<foo:bar />;`, "foo->?", "bar->?")
	assertReferences(t, `const foo = 1, bar = 2; <foo:bar />;`, "foo->declared", "bar->declared")
	assertReferences(t, `<Component foo:bar="value" />;`, "Component->?")
	assertReferences(t, `<foo:bar></foo:bar>;`, "foo->?", "bar->?", "foo->?", "bar->?")
}

func TestBuildJsxReferencesExcludeEspreeClosingTags(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.jsx",
		Path:     "/test.jsx",
	}, `<App></App>; <ns.Widget></ns.Widget>; <foo:bar></foo:bar>;`, core.ScriptKindJSX)
	m := Build(sourceFile, Options{CollectReferences: true})
	got := referencedNames(m)
	want := []string{"App->?", "ns->?"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestIsJsxComponentNameMatchesJavaScriptCodeUnitCasing(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "snake_case", want: false},
		{name: "snake_case-tag", want: false},
		{name: "Snake_case-tag", want: true},
		{name: "é_snake", want: false},
		{name: "ß_snake", want: false},
		{name: "ﬁ_snake", want: false},
		{name: "ı_snake", want: false},
		{name: "ǅ_snake", want: false},
		{name: "É_snake", want: true},
		{name: "中_snake", want: true},
		// JavaScript indexes the first high surrogate, whose uppercase mapping
		// is unchanged, rather than the astral letter as a whole.
		{name: "𐐨_snake", want: true},
		{name: "this", want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsJsxComponentName(test.name); got != test.want {
				t.Errorf("IsJsxComponentName(%q) = %v, want %v", test.name, got, test.want)
			}
		})
	}
}

func TestBuildResolvesReferencesThroughTheScopeChain(t *testing.T) {
	m := buildWithReferences(t, `
const outer = 1;
function f(param) {
  return outer + param + missing;
}
`)
	if len(m.References) != 3 {
		t.Fatalf("got %d references, want 3", len(m.References))
	}
	outer, param, missing := m.References[0], m.References[1], m.References[2]

	if got := outer.Resolved(); got == nil || got.Kind != DefVariable || got.Scope != m.Global {
		t.Errorf("`outer` should resolve to the global const, got %+v", got)
	}
	if got := param.Resolved(); got == nil || got.Kind != DefParameter {
		t.Errorf("`param` should resolve to the parameter, got %+v", got)
	}
	if missing.Resolved() != nil {
		t.Error("`missing` is declared nowhere in this file and must stay unresolved")
	}
	// Every reference in the body is evaluated in the function scope.
	if outer.From != param.From || outer.From.Kind != KindFunction {
		t.Errorf("references should come from the function scope, got %d", outer.From.Kind)
	}
}

func TestBuildClassInitializerScopes(t *testing.T) {
	m := buildWithReferences(t, `class C { field = 1; static staticField = 2; static { let inBlock; } }`)
	assertKinds(t, m, []Kind{
		KindGlobal, KindClass,
		KindClassFieldInitializer, KindClassFieldInitializer, KindClassStaticBlock,
	})

	// Both field initializers and static blocks are execution contexts, so they
	// are variable scopes in their own right.
	for _, i := range []int{2, 3, 4} {
		if m.Scopes[i].VariableScope() != m.Scopes[i] {
			t.Errorf("scope %d should be its own variable scope", i)
		}
	}
	// The class scope is not — `var` inside it hoists to the file scope.
	if m.Scopes[1].VariableScope() != m.Global {
		t.Error("a class scope is not a variable scope")
	}
	assertDeclares(t, m.Scopes[4], "inBlock", DefVariable)
}

func TestBuildGlobalAugmentationBindsNothing(t *testing.T) {
	// `declare global { ... }` reopens the global scope; it does not declare a
	// namespace named `global`.
	m := buildWithReferences(t, `global.foo = true; declare global { var injected: string; }`)
	if len(m.Global.Declarations("global")) != 0 {
		t.Error("`declare global` must not bind the name `global`")
	}
	if got := referencedNames(m); len(got) != 1 || got[0] != "global->?" {
		t.Errorf("got %v, want [global->?]", got)
	}
}

func findReference(t *testing.T, m *Manager, name string) *Reference {
	t.Helper()
	for _, ref := range m.References {
		if ref.Identifier.Text() == name {
			return ref
		}
	}
	t.Fatalf("reference %q not found", name)
	return nil
}

func TestBuildTypeDeclarationReferences(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		declaration string
		kinds       []Kind
		owner       int
	}{
		{"alias", `type T = typeof value;`, []Kind{KindGlobal}, 0},
		{"interface", `interface T { value: typeof value }`, []Kind{KindGlobal}, 0},
		{"generic alias", `type T<U> = typeof value;`, []Kind{KindGlobal, KindType}, 1},
		{"generic interface", `interface T<U> { value: typeof value }`, []Kind{KindGlobal, KindType}, 1},
		{"function type", `type T = () => typeof value;`, []Kind{KindGlobal, KindFunctionType}, 1},
		{"mapped type", `type T = { [K in keyof typeof value]: K };`, []Kind{KindGlobal, KindType}, 1},
		{"conditional check", `type T = typeof value extends unknown ? 1 : 2;`, []Kind{KindGlobal, KindType}, 1},
		{"conditional extends", `type T = unknown extends typeof value ? 1 : 2;`, []Kind{KindGlobal, KindType}, 1},
		{"conditional true", `type T = unknown extends unknown ? typeof value : 2;`, []Kind{KindGlobal, KindType}, 1},
		{"conditional false", `type T = unknown extends unknown ? 1 : typeof value;`, []Kind{KindGlobal, KindType}, 0},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			m := buildWithReferences(t, `declare const value: unknown; `+testCase.declaration)
			assertKinds(t, m, testCase.kinds)
			assertDeclares(t, m.Global, "T", DefType)
			reference := findReference(t, m, "value")
			if reference.From != m.Scopes[testCase.owner] {
				t.Errorf("value reference belongs to kind %d, want kind %d", reference.From.Kind, m.Scopes[testCase.owner].Kind)
			}
			if reference.Resolved() != m.Global.Declarations("value")[0] {
				t.Error("value reference should resolve to the outer declaration")
			}
		})
	}
}

func TestBuildKeepsTheFirstFunctionOverloadAsTheBindingAnchor(t *testing.T) {
	m := buildWithReferences(t, `function f(x: string): void; f(); function f(x: any) {}`)
	declarations := m.Global.Declarations("f")
	if len(declarations) != 2 {
		t.Fatalf("got %d f definitions, want the overload and implementation", len(declarations))
	}
	if declarations[0].DefNode.Body() != nil {
		t.Fatal("the first overload signature should be the binding anchor")
	}

	ref := findReference(t, m, "f")
	if ref.Resolved() != declarations[0] {
		t.Fatal("the call should resolve to the overload binding")
	}
	if got := ref.ResolvedIdentifier(); got != declarations[0].ID {
		t.Fatalf("resolved identifier = %p, want first overload identifier %p", got, declarations[0].ID)
	}
}

func TestBuildMarksBodylessFunctionsInAmbientContexts(t *testing.T) {
	m := build(t, `declare namespace N { function f(): void }`)
	for _, s := range m.Scopes {
		if declarations := s.Declarations("f"); len(declarations) > 0 {
			if !declarations[0].DeclareModifier {
				t.Fatal("function nested in a declare namespace should be ambient")
			}
			return
		}
	}
	t.Fatal("ambient function declaration not found")
}

func TestBuildKeepsReferenceSpacesIndependent(t *testing.T) {
	t.Run("type query is a value reference", func(t *testing.T) {
		ref := findReference(t, buildWithReferences(t, `const x = 1 as typeof x;`), "x")
		if !ref.IsValueReference() {
			t.Fatal("typeof query must retain its value-space reference")
		}
	})

	t.Run("default export identifier is dual", func(t *testing.T) {
		for _, source := range []string{
			`export default T; type T = number;`,
			`export default ((T)); type T = number;`,
		} {
			ref := findReference(t, buildWithReferences(t, source), "T")
			if !ref.IsValueReference() || !ref.IsTypeReference() {
				t.Fatalf("%s: export default identifier spaces = value:%v type:%v, want both", source, ref.IsValueReference(), ref.IsTypeReference())
			}
			if ref.Resolved() == nil || ref.Resolved().Kind != DefType {
				t.Fatalf("%s: dual default-export reference should resolve to a type-only binding", source)
			}
		}
	})

	t.Run("type predicate parameter is a value reference", func(t *testing.T) {
		ref := findReference(t, buildWithReferences(t, `function f(): x is string { return true; } const x = 1;`), "x")
		if !ref.IsValueReference() || ref.IsTypeReference() {
			t.Fatalf("type predicate spaces = value:%v type:%v, want value-only", ref.IsValueReference(), ref.IsTypeReference())
		}
		if ref.Resolved() == nil || ref.Resolved().Kind != DefVariable {
			t.Fatal("type predicate parameter should resolve to the value binding")
		}
	})
}

func TestClassifyReferenceSpacesKeepsESLintAndTypeScriptMeaningsIndependent(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		eslint     ReferenceSpace
		typescript ast.SemanticMeaning
	}{
		{
			name:       "direct export assignment",
			source:     `export default Subject; interface Subject {}`,
			eslint:     ReferenceDual,
			typescript: ast.SemanticMeaningAll,
		},
		{
			name:       "parenthesized export assignment",
			source:     `export default ((Subject)); interface Subject {}`,
			eslint:     ReferenceDual,
			typescript: ast.SemanticMeaningValue,
		},
		{
			name:       "direct type reference",
			source:     `type T = Subject; interface Subject {}`,
			eslint:     ReferenceType,
			typescript: ast.SemanticMeaningType,
		},
		{
			name:       "qualified type root",
			source:     `type T = Subject.Member; declare namespace Subject { interface Member {} }`,
			eslint:     ReferenceType,
			typescript: ast.SemanticMeaningNamespace,
		},
		{
			name:       "interface heritage root",
			source:     `interface I extends Subject.Member {} declare namespace Subject { interface Member {} }`,
			eslint:     ReferenceType,
			typescript: ast.SemanticMeaningNamespace,
		},
		{
			name:       "class implements root",
			source:     `class C implements Subject.Member {} declare namespace Subject { interface Member {} }`,
			eslint:     ReferenceType,
			typescript: ast.SemanticMeaningNamespace,
		},
		{
			name:       "class extends root",
			source:     `class C extends Subject.Base {} declare namespace Subject { class Base {} }`,
			eslint:     ReferenceValue,
			typescript: ast.SemanticMeaningValue,
		},
		{
			name:       "import equals root",
			source:     `import Alias = Subject.Member; declare namespace Subject { interface Member {} }`,
			eslint:     ReferenceValue,
			typescript: ast.SemanticMeaningNamespace,
		},
		{
			name:       "type query operand",
			source:     `type T = typeof Subject; declare const Subject: unknown;`,
			eslint:     ReferenceValue,
			typescript: ast.SemanticMeaningValue,
		},
		{
			name:       "computed type key",
			source:     `type T = { [Subject]: unknown }; declare const Subject: unique symbol;`,
			eslint:     ReferenceValue,
			typescript: ast.SemanticMeaningValue,
		},
		{
			name:       "type-only export specifier",
			source:     `interface Subject {} export type { Subject };`,
			eslint:     ReferenceType,
			typescript: ast.SemanticMeaningType | ast.SemanticMeaningNamespace,
		},
		{
			name:       "regular export specifier",
			source:     `interface Subject {} export { Subject };`,
			eslint:     ReferenceDual,
			typescript: ast.SemanticMeaningAll,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reference := findReference(t, buildWithReferences(t, test.source), "Subject")
			spaces := ClassifyReferenceSpaces(reference.Identifier)
			if spaces.ESLint != test.eslint {
				t.Errorf("ESLint space = %v, want %v", spaces.ESLint, test.eslint)
			}
			if spaces.TypeScript != test.typescript {
				t.Errorf("TypeScript meaning = %v, want %v", spaces.TypeScript, test.typescript)
			}
		})
	}
}

func TestIsReferenceIdentifierRejectsReExportAndImportTypeNames(t *testing.T) {
	for _, source := range []string{
		`export { Subject } from "mod";`,
		`type T = import("mod").Subject;`,
	} {
		sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: "/test.ts",
			Path:     "/test.ts",
		}, source, core.ScriptKindTS)
		for _, identifier := range identifiersWithText(sourceFile.AsNode(), "Subject") {
			if IsReferenceIdentifier(identifier) {
				t.Errorf("%s: Subject unexpectedly classified as a local reference", source)
			}
		}
	}
}

func TestIsReferenceIdentifierMatchesTypeScriptParserEdges(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		identifier string
		want       bool
	}{
		{name: "type query this", source: `type T = typeof this;`, identifier: "this"},
		{name: "qualified type query this", source: `type T = typeof this.member;`, identifier: "this"},
		{name: "namespace export", source: `export as namespace Missing;`, identifier: "Missing", want: true},
		{name: "import type source", source: `type T = import(Missing).Box;`, identifier: "Missing"},
		{name: "import type qualifier", source: `type T = typeof import("pkg").Missing;`, identifier: "Missing"},
		{name: "import type attributes", source: `type T = import("pkg", { with: { type: Missing } }).Box;`, identifier: "Missing"},
		{name: "import type argument", source: `type T = import("pkg").Box<Missing>;`, identifier: "Missing", want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, test.source, core.ScriptKindTS)
			identifiers := identifiersWithText(sourceFile.AsNode(), test.identifier)
			if len(identifiers) != 1 {
				t.Fatalf("found %d %q identifiers, want 1", len(identifiers), test.identifier)
			}
			if got := IsReferenceIdentifier(identifiers[0]); got != test.want {
				t.Errorf("IsReferenceIdentifier(%q) = %t, want %t", test.identifier, got, test.want)
			}
		})
	}
}

func TestIsTypeScriptJsxThisReference(t *testing.T) {
	tests := []struct {
		name       string
		fileName   string
		scriptKind core.ScriptKind
		source     string
		want       []bool
	}{
		{name: "self closing TSX", fileName: "/test.tsx", scriptKind: core.ScriptKindTSX, source: `<this />`, want: []bool{true}},
		{name: "paired TSX", fileName: "/test.tsx", scriptKind: core.ScriptKindTSX, source: `<this></this>`, want: []bool{true, true}},
		{name: "member TSX", fileName: "/test.tsx", scriptKind: core.ScriptKindTSX, source: `<this.Member />`, want: []bool{false}},
		{name: "expression TSX", fileName: "/test.tsx", scriptKind: core.ScriptKindTSX, source: `const value = this;`, want: []bool{false}},
		{name: "Espree JSX", fileName: "/test.jsx", scriptKind: core.ScriptKindJSX, source: `<this />`, want: []bool{false}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: test.fileName,
				Path:     tspath.Path(test.fileName),
			}, test.source, test.scriptKind)
			var got []bool
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindThisKeyword {
					got = append(got, IsTypeScriptJsxThisReference(node))
				}
				return node.ForEachChild(visit)
			}
			sourceFile.AsNode().ForEachChild(visit)
			if len(got) != len(test.want) {
				t.Fatalf("found %d this nodes, want %d", len(got), len(test.want))
			}
			for i := range test.want {
				if got[i] != test.want[i] {
					t.Errorf("this node %d = %t, want %t", i, got[i], test.want[i])
				}
			}
		})
	}
}

func identifiersWithText(root *ast.Node, text string) []*ast.Node {
	var identifiers []*ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier && node.Text() == text {
			identifiers = append(identifiers, node)
		}
		node.ForEachChild(visit)
		return false
	}
	root.ForEachChild(visit)
	return identifiers
}

func TestBuildFunctionTypeScopeKeepsComputedMethodKeyOutside(t *testing.T) {
	m := buildWithReferences(t, `interface I { [key](arg: Later): Result } declare const key: unique symbol; interface Later {} interface Result {}`)
	key := findReference(t, m, "key")
	if key.From == nil || key.From.Kind == KindFunctionType {
		t.Fatal("computed method key must be evaluated outside the function-type scope")
	}
	for _, name := range []string{"Later", "Result"} {
		ref := findReference(t, m, name)
		if ref.From == nil || ref.From.Kind != KindFunctionType {
			t.Fatalf("%s reference scope = %v, want KindFunctionType", name, ref.From)
		}
	}
}

func TestResolvedIdentifierSkipsAnonymousMergedDefinitions(t *testing.T) {
	m := buildWithReferences(t, `enum E { b = a, "a" = 1, a = 2 }`)
	ref := findReference(t, m, "a")
	if ref.Resolved() == nil || !ref.Resolved().Anonymous {
		t.Fatal("the first resolved definition should remain the string-literal member")
	}
	identifier := ref.ResolvedIdentifier()
	if identifier == nil || identifier.Kind != ast.KindIdentifier || identifier.Text() != "a" {
		t.Fatalf("resolved identifier = %#v, want the later named enum member", identifier)
	}
}
