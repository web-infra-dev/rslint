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
	return Build(sourceFile)
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
