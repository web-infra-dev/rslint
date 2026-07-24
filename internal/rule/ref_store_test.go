package rule

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
)

// newBoundRefStore parses source as fileName, binds it (populating symbols
// and per-file locals the way the linter guarantees before constructing a
// RefStore in production), and returns both the bound source file and its
// RefStore for direct inspection.
func newBoundRefStore(t *testing.T, fileName string, scriptKind core.ScriptKind, source string) (*ast.SourceFile, *RefStore) {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: fileName,
		Path:     tspath.Path(fileName),
	}, source, scriptKind)
	binder.BindSourceFile(sourceFile)
	return sourceFile, NewRefStore(sourceFile, &core.CompilerOptions{})
}

// identifiers returns every Identifier node under root with the given text,
// in source order.
func identifiers(root *ast.Node, text string) []*ast.Node {
	var found []*ast.Node
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier && n.Text() == text {
			found = append(found, n)
		}
		n.ForEachChild(visit)
		return false
	}
	root.ForEachChild(visit)
	return found
}

func TestRefStoreScriptFileGlobals(t *testing.T) {
	// A script file (no import/export) never puts its top-level locals in
	// scope during the resolver's walk — they're conceptually merged into
	// the global symbol table — so NewRefStore must hand the resolver this
	// file's locals as its Globals or the reference below never resolves.
	sourceFile, refs := newBoundRefStore(t, "/script.ts", core.ScriptKindTS, "var x = 1;\nx++;\n")

	occurrences := identifiers(sourceFile.AsNode(), "x")
	if len(occurrences) != 2 {
		t.Fatalf("expected 2 occurrences of x, got %d", len(occurrences))
	}
	declIdent, useIdent := occurrences[0], occurrences[1]
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	got := refs.References(sym)
	if len(got) != 1 || got[0] != useIdent {
		t.Fatalf("References = %v, want [%v] (the `x++` use)", got, useIdent)
	}
}

func TestRefStoreShorthandPropertyRead(t *testing.T) {
	// `{ x }` reads x; IsDeclarationName also treats this name as a
	// declaration (it declares the object's own property), so a naive
	// exclusion would discard this real reference.
	sourceFile, refs := newBoundRefStore(t, "/shorthand-read.ts", core.ScriptKindTS,
		"export {}; function f() { var x = 1; return { x }; }")

	occurrences := identifiers(sourceFile.AsNode(), "x")
	if len(occurrences) != 2 {
		t.Fatalf("expected 2 occurrences of x, got %d", len(occurrences))
	}
	declIdent, shorthandIdent := occurrences[0], occurrences[1]
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	got := refs.References(sym)
	if len(got) != 1 || got[0] != shorthandIdent {
		t.Fatalf("References = %v, want [%v] (the `{ x }` shorthand read)", got, shorthandIdent)
	}
}

func TestRefStoreShorthandPropertyWrite(t *testing.T) {
	// `({ x } = obj)` writes to x once the object literal is reinterpreted
	// as an assignment pattern; this is a real reference distinct from the
	// declaration name.
	sourceFile, refs := newBoundRefStore(t, "/shorthand-write.ts", core.ScriptKindTS,
		"export {}; function f() { var x; ({ x } = { y: 1 }); return x; }")

	occurrences := identifiers(sourceFile.AsNode(), "x")
	if len(occurrences) != 3 {
		t.Fatalf("expected 3 occurrences of x, got %d", len(occurrences))
	}
	declIdent, writeIdent, useIdent := occurrences[0], occurrences[1], occurrences[2]
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	got := refs.References(sym)
	if len(got) != 2 || got[0] != writeIdent || got[1] != useIdent {
		t.Fatalf("References = %v, want [%v %v] (the `{ x }` write and the later use)", got, writeIdent, useIdent)
	}
}

func TestRefStoreMetaPropertyExcluded(t *testing.T) {
	// `target` in `new.target` is a syntactic name, not an identifier that
	// can reference a same-named variable.
	sourceFile, refs := newBoundRefStore(t, "/meta-property.ts", core.ScriptKindTS,
		"export {}; function f() { var target = 1; if (new.target) {} return target; }")

	occurrences := identifiers(sourceFile.AsNode(), "target")
	if len(occurrences) != 3 {
		t.Fatalf("expected 3 occurrences of target, got %d", len(occurrences))
	}
	declIdent, metaIdent, useIdent := occurrences[0], occurrences[1], occurrences[2]
	if metaIdent.Parent == nil || metaIdent.Parent.Kind != ast.KindMetaProperty {
		t.Fatalf("expected the second `target` to be new.target's MetaProperty name, got parent kind %v", metaIdent.Parent.Kind)
	}
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	got := refs.References(sym)
	if len(got) != 1 || got[0] != useIdent {
		t.Fatalf("References = %v, want [%v] (new.target's `target` must be excluded)", got, useIdent)
	}
}

func TestRefStoreImportAttributeNameExcluded(t *testing.T) {
	// `type` in an import attribute is a syntactic key, not a reference to a
	// same-named variable.
	sourceFile, refs := newBoundRefStore(t, "/import-attribute.ts", core.ScriptKindTS,
		`import data from "pkg" with { type: "json" }; var type = 1; type++;`)

	occurrences := identifiers(sourceFile.AsNode(), "type")
	if len(occurrences) != 3 {
		t.Fatalf("expected 3 occurrences of type, got %d", len(occurrences))
	}
	attributeIdent, declIdent, useIdent := occurrences[0], occurrences[1], occurrences[2]
	if attributeIdent.Parent == nil || attributeIdent.Parent.Kind != ast.KindImportAttribute {
		t.Fatalf("expected the first `type` to be an ImportAttribute name, got parent kind %v", attributeIdent.Parent.Kind)
	}
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	got := refs.References(sym)
	if len(got) != 1 || got[0] != useIdent {
		t.Fatalf("References = %v, want [%v] (the import attribute key must be excluded)", got, useIdent)
	}
}

func TestRefStoreJsxNamespacedNameExcluded(t *testing.T) {
	// `bar` in the namespaced JSX tag name `<bar:qux />` is syntactic, not a
	// reference to a same-named variable.
	sourceFile, refs := newBoundRefStore(t, "/jsx-namespaced.tsx", core.ScriptKindTSX,
		"export {}; function f() { var bar = 1; const el = <bar:qux />; return bar; }")

	occurrences := identifiers(sourceFile.AsNode(), "bar")
	if len(occurrences) != 3 {
		t.Fatalf("expected 3 occurrences of bar, got %d", len(occurrences))
	}
	declIdent, tagIdent, useIdent := occurrences[0], occurrences[1], occurrences[2]
	if tagIdent.Parent == nil || tagIdent.Parent.Kind != ast.KindJsxNamespacedName {
		t.Fatalf("expected the second `bar` to be the JsxNamespacedName's namespace, got parent kind %v", tagIdent.Parent.Kind)
	}
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	got := refs.References(sym)
	if len(got) != 1 || got[0] != useIdent {
		t.Fatalf("References = %v, want [%v] (the namespaced JSX tag's `bar` must be excluded)", got, useIdent)
	}
}
