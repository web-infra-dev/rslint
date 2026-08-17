// TestBuild* lock in the shape of the scope tree the shared model produces —
// which scopes exist, what nests inside what, and which scope owns each
// binding. Rules read those three things and nothing else, so a regression
// here is a regression in every consumer.
package scope

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
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
	assertDeclares(t, m.Scopes[2], "blockOnly", DefVariable)
	if len(m.Scopes[2].Declarations("hoisted")) != 0 {
		t.Error("var declaration should not also live in the block scope")
	}
	if m.Scopes[2].VariableScope() != m.Scopes[1] {
		t.Error("block scope's variable scope should be the enclosing function")
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
	// JSX: a lower-case tag is an intrinsic element, not a binding.
	assertReferences(t, `<div />;`)
	assertReferences(t, `<App />;`, "App->?")
	assertReferences(t, `<ns.Widget />;`, "ns->?")
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
