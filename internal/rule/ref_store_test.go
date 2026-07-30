package rule

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/utils"
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
	return sourceFile, NewRefStore(sourceFile, &core.CompilerOptions{}, nil)
}

// newCheckedRefStore builds a real Program (via the shared fixtures
// tsconfig, which includes the DOM lib) for source and returns its parsed
// source file, a type-aware RefStore, and a cleanup func the caller must
// defer. Unlike newBoundRefStore, this exercises the TypeChecker fallback:
// the source file is one among the fixtures project's files, so identifiers
// naming a lib/ambient global genuinely can't be resolved by the binder scope
// walk alone.
func newCheckedRefStore(t *testing.T, source string) (*ast.SourceFile, *RefStore, func()) {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "ref-store-tc-fallback.ts")
	fs := utils.NewOverlayVFSForFile(filePath, source)

	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", utils.CreateCompilerHost(rootDir, fs))
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		t.Fatalf("GetSourceFile(%q) = nil", filePath)
	}
	tc, done := program.GetTypeChecker(t.Context())
	return sourceFile, NewRefStore(sourceFile, program.Options(), tc), done
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

func TestRefStoreResolve(t *testing.T) {
	// Resolve is the forward counterpart to References: given a reference
	// identifier, find its declaring symbol.
	sourceFile, refs := newBoundRefStore(t, "/resolve.ts", core.ScriptKindTS,
		"export {}; function f() { var x = 1; return x; }")

	occurrences := identifiers(sourceFile.AsNode(), "x")
	if len(occurrences) != 2 {
		t.Fatalf("expected 2 occurrences of x, got %d", len(occurrences))
	}
	declIdent, useIdent := occurrences[0], occurrences[1]
	wantSym := declIdent.Parent.Symbol()
	if wantSym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	if got := refs.Resolve(useIdent); got != wantSym {
		t.Fatalf("Resolve(use) = %v, want %v", got, wantSym)
	}
	// Resolving the declaration name itself is meaningless the same way
	// References excludes it — declaration names aren't reference positions.
	if got := refs.Resolve(declIdent); got != nil {
		t.Fatalf("Resolve(decl) = %v, want nil", got)
	}
}

func TestRefStoreResolveShorthandPropertyWrite(t *testing.T) {
	// Resolve must handle shorthand destructuring writes the same way
	// References does: the shorthand name is a reference position even
	// though IsDeclarationName also treats it as one.
	sourceFile, refs := newBoundRefStore(t, "/resolve-shorthand.ts", core.ScriptKindTS,
		"export {}; function f() { var x; ({ x } = { y: 1 }); return x; }")

	occurrences := identifiers(sourceFile.AsNode(), "x")
	if len(occurrences) != 3 {
		t.Fatalf("expected 3 occurrences of x, got %d", len(occurrences))
	}
	declIdent, writeIdent := occurrences[0], occurrences[1]
	wantSym := declIdent.Parent.Symbol()
	if wantSym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	if got := refs.Resolve(writeIdent); got != wantSym {
		t.Fatalf("Resolve(shorthand write) = %v, want %v", got, wantSym)
	}
}

func TestRefStoreResolveExcludedPositions(t *testing.T) {
	// Property names, import bindings, and other non-reference positions
	// must resolve to nil, matching isReferencePosition's exclusions.
	sourceFile, refs := newBoundRefStore(t, "/resolve-excluded.ts", core.ScriptKindTS,
		"export {}; function f() { var x = 1; return { x: x }.x; }")

	occurrences := identifiers(sourceFile.AsNode(), "x")
	if len(occurrences) != 4 {
		t.Fatalf("expected 4 occurrences of x, got %d", len(occurrences))
	}
	// occurrences: [0] var x decl, [1] property key `x:`, [2] value `x`, [3] `.x` property access
	propertyKey, propertyAccess := occurrences[1], occurrences[3]

	if got := refs.Resolve(propertyKey); got != nil {
		t.Fatalf("Resolve(property key) = %v, want nil", got)
	}
	if got := refs.Resolve(propertyAccess); got != nil {
		t.Fatalf("Resolve(property access name) = %v, want nil", got)
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

func TestRefStoreExportAssignmentResolvesTypeOnlySymbol(t *testing.T) {
	// Export assignments resolve entity names in value, type, and namespace
	// space. An interface consumed by `export =` is therefore a real reference.
	sourceFile, refs := newBoundRefStore(t, "/export-assignment.ts", core.ScriptKindTS,
		"interface Foo { bar: string }\nexport = Foo;\n")

	occurrences := identifiers(sourceFile.AsNode(), "Foo")
	if len(occurrences) != 2 {
		t.Fatalf("expected 2 occurrences of Foo, got %d", len(occurrences))
	}
	declIdent, exportIdent := occurrences[0], occurrences[1]
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	got := refs.References(sym)
	if len(got) != 1 || got[0] != exportIdent {
		t.Fatalf("References = %v, want [%v] (the `export = Foo` reference)", got, exportIdent)
	}
}

func TestRefStoreExportDefaultNamedFunction(t *testing.T) {
	// A named default export is bound to an export symbol named "default", so
	// looking its candidates up by symbol name finds nothing: the references
	// are bucketed under the name they are written with.
	sourceFile, refs := newBoundRefStore(t, "/export-default.ts", core.ScriptKindTS,
		"export default function foo() { foo = 1; }\nfoo = 2;\n")

	occurrences := identifiers(sourceFile.AsNode(), "foo")
	if len(occurrences) != 3 {
		t.Fatalf("expected 3 occurrences of foo, got %d", len(occurrences))
	}
	declIdent, innerIdent, outerIdent := occurrences[0], occurrences[1], occurrences[2]
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
		return
	}
	if sym.Name != ast.InternalSymbolNameDefault {
		t.Fatalf("declaration symbol name = %q, want %q", sym.Name, ast.InternalSymbolNameDefault)
	}

	got := refs.References(sym)
	if len(got) != 2 || got[0] != innerIdent || got[1] != outerIdent {
		t.Fatalf("References = %v, want [%v %v] (both assignments to foo)", got, innerIdent, outerIdent)
	}
}

func TestRefStoreLocalExportSpecifierRead(t *testing.T) {
	// `export { foo }` (no rename) reads the local `foo` — ESLint's
	// scope-manager marks this as a reference, unlike import/export bindings
	// in general.
	sourceFile, refs := newBoundRefStore(t, "/export-specifier-read.ts", core.ScriptKindTS,
		"var foo = 1;\nexport { foo };\n")

	occurrences := identifiers(sourceFile.AsNode(), "foo")
	if len(occurrences) != 2 {
		t.Fatalf("expected 2 occurrences of foo, got %d", len(occurrences))
	}
	declIdent, specifierIdent := occurrences[0], occurrences[1]
	if specifierIdent.Parent == nil || specifierIdent.Parent.Kind != ast.KindExportSpecifier {
		t.Fatalf("expected the second foo to be the ExportSpecifier name, got parent kind %v", specifierIdent.Parent.Kind)
	}
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	if got := refs.Resolve(specifierIdent); got != sym {
		t.Fatalf("Resolve(export specifier name) = %v, want %v", got, sym)
	}
	got := refs.References(sym)
	if len(got) != 1 || got[0] != specifierIdent {
		t.Fatalf("References = %v, want [%v] (the `export { foo }` read)", got, specifierIdent)
	}
}

func TestRefStoreLocalExportSpecifierAliasReadsPropertyName(t *testing.T) {
	// `export { foo as bar }` reads the local `foo` through PropertyName;
	// `bar` is only the external export label and never a reference to
	// anything in this file.
	sourceFile, refs := newBoundRefStore(t, "/export-specifier-alias.ts", core.ScriptKindTS,
		"var foo = 1;\nexport { foo as bar };\n")

	fooOccurrences := identifiers(sourceFile.AsNode(), "foo")
	if len(fooOccurrences) != 2 {
		t.Fatalf("expected 2 occurrences of foo, got %d", len(fooOccurrences))
	}
	declIdent, propertyNameIdent := fooOccurrences[0], fooOccurrences[1]
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	if got := refs.Resolve(propertyNameIdent); got != sym {
		t.Fatalf("Resolve(PropertyName foo) = %v, want %v", got, sym)
	}

	barOccurrences := identifiers(sourceFile.AsNode(), "bar")
	if len(barOccurrences) != 1 {
		t.Fatalf("expected 1 occurrence of bar, got %d", len(barOccurrences))
	}
	if got := refs.Resolve(barOccurrences[0]); got != nil {
		t.Fatalf("Resolve(export label bar) = %v, want nil (it's not a reference to anything)", got)
	}
}

func TestRefStoreReExportSpecifierExcluded(t *testing.T) {
	// `export { foo } from 'mod'` references the other module's binding, not
	// anything in this file, even though a same-named local exists here.
	sourceFile, refs := newBoundRefStore(t, "/re-export.ts", core.ScriptKindTS,
		`var foo = 1;
export { foo } from "./mod";
foo;
`)

	occurrences := identifiers(sourceFile.AsNode(), "foo")
	if len(occurrences) != 3 {
		t.Fatalf("expected 3 occurrences of foo, got %d", len(occurrences))
	}
	declIdent, specifierIdent, useIdent := occurrences[0], occurrences[1], occurrences[2]
	if specifierIdent.Parent == nil || specifierIdent.Parent.Kind != ast.KindExportSpecifier {
		t.Fatalf("expected the second foo to be the ExportSpecifier name, got parent kind %v", specifierIdent.Parent.Kind)
	}
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	if got := refs.Resolve(specifierIdent); got != nil {
		t.Fatalf("Resolve(re-export specifier name) = %v, want nil", got)
	}
	got := refs.References(sym)
	if len(got) != 1 || got[0] != useIdent {
		t.Fatalf("References = %v, want [%v] (the re-export specifier must be excluded)", got, useIdent)
	}
}

func TestRefStoreReExportSpecifierAliasExcluded(t *testing.T) {
	// `export { foo as bar } from 'mod'` — neither the source name nor the
	// alias label references anything in this file.
	sourceFile, refs := newBoundRefStore(t, "/re-export-alias.ts", core.ScriptKindTS,
		`export { foo as bar } from "./mod";`)

	fooOccurrences := identifiers(sourceFile.AsNode(), "foo")
	if len(fooOccurrences) != 1 {
		t.Fatalf("expected 1 occurrence of foo, got %d", len(fooOccurrences))
	}
	if got := refs.Resolve(fooOccurrences[0]); got != nil {
		t.Fatalf("Resolve(re-export PropertyName foo) = %v, want nil", got)
	}

	barOccurrences := identifiers(sourceFile.AsNode(), "bar")
	if len(barOccurrences) != 1 {
		t.Fatalf("expected 1 occurrence of bar, got %d", len(barOccurrences))
	}
	if got := refs.Resolve(barOccurrences[0]); got != nil {
		t.Fatalf("Resolve(re-export label bar) = %v, want nil", got)
	}
}

func TestRefStoreExportSpecifierResolvesTypeOnlySymbol(t *testing.T) {
	// `export { Foo }` can re-export a type-only local; referenceMeaning must
	// include SymbolFlagsType or this never resolves.
	sourceFile, refs := newBoundRefStore(t, "/export-specifier-type.ts", core.ScriptKindTS,
		"type Foo = {};\nexport { Foo };\n")

	occurrences := identifiers(sourceFile.AsNode(), "Foo")
	if len(occurrences) != 2 {
		t.Fatalf("expected 2 occurrences of Foo, got %d", len(occurrences))
	}
	declIdent, specifierIdent := occurrences[0], occurrences[1]
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	got := refs.References(sym)
	if len(got) != 1 || got[0] != specifierIdent {
		t.Fatalf("References = %v, want [%v] (the `export { Foo }` reference to the type alias)", got, specifierIdent)
	}
}

func TestRefStoreTypeOnlyExportSpecifierIndexed(t *testing.T) {
	// `export type { x }` / `export { type x }` resolves to the binding like
	// any other local export — ESLint's scope-manager records a type-flagged
	// reference for it. Consumers that only follow runtime reads classify the
	// reference site (utils.IsNonReferenceIdentifier reports these as
	// non-reads); the index itself keeps them.
	for name, source := range map[string]string{
		"export type clause":  "let x;\nexport type { x };\nx = 1;\n",
		"type specifier":      "let x;\nexport { type x };\nx = 1;\n",
		"export type aliased": "let x;\nexport type { x as y };\nx = 1;\n",
	} {
		t.Run(name, func(t *testing.T) {
			sourceFile, refs := newBoundRefStore(t, "/type-only-export.ts", core.ScriptKindTS, source)

			occurrences := identifiers(sourceFile.AsNode(), "x")
			if len(occurrences) != 3 {
				t.Fatalf("expected 3 occurrences of x, got %d", len(occurrences))
			}
			declIdent, specifierIdent, writeIdent := occurrences[0], occurrences[1], occurrences[2]
			sym := declIdent.Parent.Symbol()
			if sym == nil {
				t.Fatal("declaration identifier has no bound symbol")
			}

			if got := refs.Resolve(specifierIdent); got != sym {
				t.Fatalf("Resolve(type-only export specifier x) = %v, want %v", got, sym)
			}
			got := refs.References(sym)
			if len(got) != 2 || got[0] != specifierIdent || got[1] != writeIdent {
				t.Fatalf("References = %v, want [%v %v] (the specifier and the `x = 1` write)", got, specifierIdent, writeIdent)
			}
		})
	}
}

func TestRefStoreTypeOnlyExportSpecifierResolvesTypeSymbol(t *testing.T) {
	// The type-space half of the meaning must stay: `export type { Foo }` of
	// an actual type still references it.
	sourceFile, refs := newBoundRefStore(t, "/type-only-export-type.ts", core.ScriptKindTS,
		"type Foo = {};\nexport type { Foo };\n")

	occurrences := identifiers(sourceFile.AsNode(), "Foo")
	if len(occurrences) != 2 {
		t.Fatalf("expected 2 occurrences of Foo, got %d", len(occurrences))
	}
	declIdent, specifierIdent := occurrences[0], occurrences[1]
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	if got := refs.Resolve(specifierIdent); got != sym {
		t.Fatalf("Resolve(type-only export specifier Foo) = %v, want %v", got, sym)
	}
	got := refs.References(sym)
	if len(got) != 1 || got[0] != specifierIdent {
		t.Fatalf("References = %v, want [%v] (the `export type { Foo }` reference)", got, specifierIdent)
	}
}

func TestRefStoreTypeOnlyExportSpecifierWithChecker(t *testing.T) {
	// With a TypeChecker present, the aliased type-only specifier's
	// PropertyName must still resolve through the binder onto the same raw
	// symbol the write does — not detour through the checker fallback, which
	// would file it under a different (checker-merged alias) identity and
	// split the index.
	sourceFile, refs, done := newCheckedRefStore(t,
		"let x;\nexport type { x as y };\nx = 1;\n")
	defer done()

	occurrences := identifiers(sourceFile.AsNode(), "x")
	if len(occurrences) != 3 {
		t.Fatalf("expected 3 occurrences of x, got %d", len(occurrences))
	}
	declIdent, specifierIdent, writeIdent := occurrences[0], occurrences[1], occurrences[2]
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	if got := refs.Resolve(specifierIdent); got != sym {
		t.Fatalf("Resolve(type-only export PropertyName x) = %v, want %v", got, sym)
	}
	got := refs.References(sym)
	if len(got) != 2 || got[0] != specifierIdent || got[1] != writeIdent {
		t.Fatalf("References = %v, want [%v %v] (the specifier and the `x = 1` write)", got, specifierIdent, writeIdent)
	}
}

func TestRefStoreExportedFunctionDeclaration(t *testing.T) {
	// A non-default export is bound to an export symbol too; the local symbol
	// left in the file's locals carries only the ExportValue marker, so
	// references must still resolve to the queried export symbol.
	sourceFile, refs := newBoundRefStore(t, "/export-named.ts", core.ScriptKindTS,
		"export function foo() {}\nfoo = 1;\n")

	occurrences := identifiers(sourceFile.AsNode(), "foo")
	if len(occurrences) != 2 {
		t.Fatalf("expected 2 occurrences of foo, got %d", len(occurrences))
	}
	declIdent, useIdent := occurrences[0], occurrences[1]
	sym := declIdent.Parent.Symbol()
	if sym == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	got := refs.References(sym)
	if len(got) != 1 || got[0] != useIdent {
		t.Fatalf("References = %v, want [%v] (the `foo = 1` assignment)", got, useIdent)
	}
}

func TestRefStoreResolveCheckerFallback(t *testing.T) {
	// window is declared in lib.dom.d.ts, not this file — the binder scope
	// walk structurally can't resolve it, so this only succeeds through
	// Resolve's TypeChecker fallback.
	sourceFile, refs, done := newCheckedRefStore(t, "export {}; function f() { window; window = 1 as any; }")
	defer done()

	occurrences := identifiers(sourceFile.AsNode(), "window")
	if len(occurrences) != 2 {
		t.Fatalf("expected 2 occurrences of window, got %d", len(occurrences))
	}
	readIdent, writeIdent := occurrences[0], occurrences[1]

	sym := refs.Resolve(readIdent)
	if sym == nil {
		t.Fatal("Resolve(window) = nil, want the lib.dom.d.ts global symbol")
	}
	// Two different reference sites for the same external declaration must
	// resolve to the identical symbol object — References keys on this.
	if got := refs.Resolve(writeIdent); got != sym {
		t.Fatalf("Resolve(second window reference) = %v, want %v (same symbol object)", got, sym)
	}

	// The fallback must also make References work for this symbol, not just
	// Resolve: resolvePending has to drain the "window" candidate bucket
	// through the TypeChecker too, or these references would stay pending
	// forever and every rule asking "was this written" would silently see
	// nothing.
	got := refs.References(sym)
	if len(got) != 2 || got[0] != readIdent || got[1] != writeIdent {
		t.Fatalf("References = %v, want [%v %v] (both window references)", got, readIdent, writeIdent)
	}
}

func TestRefStoreResolveCheckerFallbackExcludedPositions(t *testing.T) {
	// A TypeChecker resolves a property key's own declaration name to the
	// property's symbol (its only declaration is that very key), not to
	// anything the identifier "references". Resolve must still treat it as a
	// non-reference position and return nil, the same as without a checker —
	// the TypeChecker fallback is only for reference positions the binder
	// scope walk can't place, not a way to resolve declaration names into
	// their own symbol.
	sourceFile, refs, done := newCheckedRefStore(t,
		"export {}; function f() { var obj = { foo: 2 }; return obj.foo; }")
	defer done()

	occurrences := identifiers(sourceFile.AsNode(), "foo")
	if len(occurrences) != 2 {
		t.Fatalf("expected 2 occurrences of foo, got %d", len(occurrences))
	}
	propertyKey := occurrences[0]

	if got := refs.Resolve(propertyKey); got != nil {
		t.Fatalf("Resolve(property key) = %v, want nil", got)
	}
}

func TestRefStoreResolveNoCheckerFallback(t *testing.T) {
	// Without a TypeChecker, Resolve stays binder-only: an identifier the
	// binder can't resolve stays unresolved.
	sourceFile, refs := newBoundRefStore(t, "/no-fallback.ts", core.ScriptKindTS,
		"export {}; function f() { window; }")

	occurrences := identifiers(sourceFile.AsNode(), "window")
	if len(occurrences) != 1 {
		t.Fatalf("expected 1 occurrence of window, got %d", len(occurrences))
	}
	if got := refs.Resolve(occurrences[0]); got != nil {
		t.Fatalf("Resolve(window) = %v, want nil (no TypeChecker supplied)", got)
	}
}

func TestRefStoreMergedGlobalSymbolIdentity(t *testing.T) {
	// A global script file's `interface Map` merges with lib.d.ts's `Map`,
	// and the two reference sites below reach that one declaration by
	// different routes: the type reference through the binder scope walk
	// (the file's own locals are the resolver's globals table), the value
	// reference only through the checker fallback, since the local interface
	// carries no value meaning. Both identities must index into one bucket
	// or References returns whichever half matches the queried symbol.
	sourceFile, refs, done := newCheckedRefStore(t,
		"Map;\ntype Alias = Map<string, string>;\ninterface Map<K, V> {}\n")
	defer done()

	occurrences := identifiers(sourceFile.AsNode(), "Map")
	if len(occurrences) != 3 {
		t.Fatalf("expected 3 occurrences of Map, got %d", len(occurrences))
	}
	valueRef, typeRef := occurrences[0], occurrences[1]

	fromChecker := refs.Resolve(valueRef)
	if fromChecker == nil {
		t.Fatal("Resolve(value Map) = nil, want the merged lib.d.ts symbol")
	}
	fromBinder := refs.Resolve(typeRef)
	if fromBinder == nil {
		t.Fatal("Resolve(type Map) = nil, want the file's interface symbol")
	}

	want := []*ast.Node{valueRef, typeRef}
	for _, sym := range []*ast.Symbol{fromChecker, fromBinder} {
		got := refs.References(sym)
		if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("References = %v, want %v (both the value and the type reference)", got, want)
		}
	}
}
