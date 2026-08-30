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
	return newBoundRefStoreWithLanguageOptions(t, fileName, scriptKind, source, LanguageOptions{})
}

func newBoundRefStoreWithLanguageOptions(t *testing.T, fileName string, scriptKind core.ScriptKind, source string, languageOptions LanguageOptions) (*ast.SourceFile, *RefStore) {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: fileName,
		Path:     tspath.Path(fileName),
	}, source, scriptKind)
	binder.BindSourceFile(sourceFile)
	_, refsInit, _ := ResolveLanguageDefaults(fileName, languageOptions)
	return sourceFile, NewRefStore(sourceFile, &core.CompilerOptions{}, nil, refsInit)
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
	filePath := tspath.ResolvePath(rootDir.Dir, "ref-store-tc-fallback.ts")
	fs := utils.NewOverlayVFS(rootDir.FS, map[string]string{filePath: source})

	program, err := utils.CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", utils.CreateCompilerHost(rootDir.Dir, fs))
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		t.Fatalf("GetSourceFile(%q) = nil", filePath)
	}
	tc, done := program.GetTypeChecker(t.Context())
	_, refsInit, _ := ResolveLanguageDefaults(sourceFile.FileName(), LanguageOptions{})
	return sourceFile, NewRefStore(sourceFile, program.Options(), tc, refsInit), done
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

func TestRefStoreGlobalReferenceUsesEffectiveProgramScope(t *testing.T) {
	t.Parallel()

	valueMeaning := ast.SymbolFlagsValue | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	tests := []struct {
		name       string
		source     string
		sourceType string
		want       bool
	}{
		{
			name:       "script type declaration joins global scope",
			source:     "type Boolean = string; Boolean(value);",
			sourceType: "script",
			want:       false,
		},
		{
			name:       "module type declaration has no value meaning",
			source:     "type Boolean = string; Boolean(value);",
			sourceType: "module",
			want:       true,
		},
		{
			name:       "module namespace declaration has value meaning",
			source:     "namespace Boolean {} Boolean(value);",
			sourceType: "module",
			want:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			sourceFile, refs := newBoundRefStoreWithLanguageOptions(
				t,
				"/effective-program-scope.ts",
				core.ScriptKindTS,
				test.source,
				LanguageOptions{SourceType: test.sourceType},
			)
			occurrences := identifiers(sourceFile.AsNode(), "Boolean")
			if len(occurrences) != 2 {
				t.Fatalf("found %d Boolean identifiers, want 2", len(occurrences))
			}
			if got := refs.IsGlobalNameReference(occurrences[1], "Boolean", valueMeaning); got != test.want {
				t.Errorf("IsGlobalNameReference(Boolean) = %v, want %v", got, test.want)
			}
		})
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

func TestRefStoreResolveTypeOrNamespaceUsesExactReferenceMeaning(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		identifier   string
		wantResolved bool
	}{
		{
			name:         "direct export resolves type",
			source:       `export = Record;`,
			identifier:   "Record",
			wantResolved: true,
		},
		{
			name:       "parenthesized export is value-only",
			source:     `export = ((Record));`,
			identifier: "Record",
		},
		{
			name:       "import equals rejects type-only target",
			source:     `import Alias = Record;`,
			identifier: "Record",
		},
		{
			name:         "import equals resolves namespace target",
			source:       `import Alias = Intl.Collator;`,
			identifier:   "Intl",
			wantResolved: true,
		},
		{
			name:         "qualified type root resolves namespace",
			source:       `type T = Intl.CollatorOptions;`,
			identifier:   "Intl",
			wantResolved: true,
		},
		{
			name:       "type query remains value-only",
			source:     `type T = typeof Intl;`,
			identifier: "Intl",
		},
		{
			name:       "class extends remains value-only",
			source:     `class C extends Intl.Collator {}`,
			identifier: "Intl",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile, refs, done := newCheckedRefStore(t, test.source)
			defer done()
			occurrences := identifiers(sourceFile.AsNode(), test.identifier)
			if len(occurrences) != 1 {
				t.Fatalf("identifier occurrences = %d, want 1", len(occurrences))
			}
			resolved := refs.ResolveTypeOrNamespace(occurrences[0])
			if (resolved != nil) != test.wantResolved {
				t.Fatalf("ResolveTypeOrNamespace(%s) = %v, wantResolved %v", test.identifier, resolved, test.wantResolved)
			}
		})
	}
}

func TestRefStoreResolveTypeOrNamespaceKeepsTheExactTarget(t *testing.T) {
	t.Run("import equals uses namespace meaning", func(t *testing.T) {
		sourceFile, refs, done := newCheckedRefStore(t, `import Alias = Record;`)
		defer done()
		occurrences := identifiers(sourceFile.AsNode(), "Record")
		if len(occurrences) != 1 {
			t.Fatalf("identifier occurrences = %d, want 1", len(occurrences))
		}
		if got := refs.Resolve(occurrences[0]); got != nil {
			t.Fatalf("Resolve(type-only import-equals root) = %v, want nil", got)
		}
	})

	t.Run("local value shadows external type", func(t *testing.T) {
		sourceFile, refs, done := newCheckedRefStore(t, `const Record = 1; export default Record;`)
		defer done()
		occurrences := identifiers(sourceFile.AsNode(), "Record")
		if len(occurrences) != 2 {
			t.Fatalf("identifier occurrences = %d, want 2", len(occurrences))
		}
		if got := refs.ResolveTypeOrNamespace(occurrences[1]); got != nil {
			t.Fatalf("ResolveTypeOrNamespace(local value) = %v, want nil", got)
		}
		if got, want := refs.Resolve(occurrences[1]), occurrences[0].Parent.Symbol(); got != want {
			t.Fatalf("Resolve(local value) = %v, want %v", got, want)
		}
	})

	t.Run("local alias retains its authored identity", func(t *testing.T) {
		sourceFile, refs, done := newCheckedRefStore(t, `import type { T as Alias } from "./foo"; export default Alias;`)
		defer done()
		occurrences := identifiers(sourceFile.AsNode(), "Alias")
		if len(occurrences) != 2 {
			t.Fatalf("identifier occurrences = %d, want 2", len(occurrences))
		}
		got := refs.ResolveTypeOrNamespace(occurrences[1])
		want := occurrences[0].Parent.Symbol()
		if got == nil || got != want {
			t.Fatalf("ResolveTypeOrNamespace(local alias) = %v, want %v", got, want)
		}
	})
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

func TestRefStoreTypeOnlyExportSpecifierExcludesValueOnlySymbol(t *testing.T) {
	// A type-only export cannot resolve to a value-only declaration. This is
	// declaration-space filtering, not merely read classification: the
	// specifier must be absent from the value symbol's reference list.
	for name, source := range map[string]string{
		"export type clause":   "let x;\nexport type { x };\nx = 1;\n",
		"inline type":          "let x;\nexport { type x };\nx = 1;\n",
		"aliased":              "let x;\nexport type { x as y };\nx = 1;\n",
		"value-only function":  "function x() {}\nexport type { x };\nx();\n",
		"inline function type": "function x() {}\nexport { type x };\nx();\n",
	} {
		t.Run(name, func(t *testing.T) {
			sourceFile, refs := newBoundRefStore(t, "/type-only-export.ts", core.ScriptKindTS, source)

			occurrences := identifiers(sourceFile.AsNode(), "x")
			if len(occurrences) != 3 {
				t.Fatalf("expected 3 occurrences of x, got %d", len(occurrences))
			}
			declIdent, specifierIdent, valueUseIdent := occurrences[0], occurrences[1], occurrences[2]
			sym := declIdent.Parent.Symbol()
			if sym == nil {
				t.Fatal("declaration identifier has no bound symbol")
			}

			if got := refs.Resolve(specifierIdent); got != nil {
				t.Fatalf("Resolve(type-only export specifier x) = %v, want nil for a value-only declaration", got)
			}
			got := refs.References(sym)
			if len(got) != 1 || got[0] != valueUseIdent {
				t.Fatalf("References = %v, want [%v] (only the value-space use)", got, valueUseIdent)
			}
		})
	}
}

func TestRefStoreTypeOnlyExportSpecifierResolvesTypeCapableSymbol(t *testing.T) {
	// Type aliases/interfaces, dual-space declarations, namespaces, merged
	// symbols, and import aliases are all type-capable in scope-manager's
	// model and must keep their type-only export reference.
	for name, source := range map[string]string{
		"type alias":          "type X = {};\nexport type { X };\n",
		"interface":           "interface X {}\nexport type { X };\n",
		"class":               "class X {}\nexport type { X };\n",
		"enum":                "enum X { A }\nexport type { X };\n",
		"namespace":           "namespace X {}\nexport type { X };\n",
		"function namespace":  "function X() {}\nnamespace X {}\nexport type { X };\n",
		"interface and value": "interface X {}\nlet X;\nexport type { X };\n",
		"import alias":        "import { Foo as X } from './missing';\nexport type { X };\n",
		"type import alias":   "import type { Foo as X } from './missing';\nexport type { X };\n",
	} {
		t.Run(name, func(t *testing.T) {
			sourceFile, refs := newBoundRefStore(t, "/type-only-export-type.ts", core.ScriptKindTS, source)
			occurrences := identifiers(sourceFile.AsNode(), "X")
			if len(occurrences) < 2 {
				t.Fatalf("expected at least 2 occurrences of X, got %d", len(occurrences))
			}
			declIdent, specifierIdent := occurrences[0], occurrences[len(occurrences)-1]
			sym := declIdent.Parent.Symbol()
			if sym == nil {
				t.Fatal("declaration identifier has no bound symbol")
			}

			if got := refs.Resolve(specifierIdent); got != sym {
				t.Fatalf("Resolve(type-only export specifier X) = %v, want %v", got, sym)
			}
			got := refs.References(sym)
			if len(got) != 1 || got[0] != specifierIdent {
				t.Fatalf("References = %v, want [%v] (the type-only export reference)", got, specifierIdent)
			}
		})
	}
}

func TestRefStoreTypeOnlyExportSpecifierWithCheckerExcludesValueOnlySymbol(t *testing.T) {
	// The checker fallback must not reintroduce the value binding after the
	// binder correctly rejects it.
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

	if got := refs.Resolve(specifierIdent); got != nil {
		t.Fatalf("Resolve(type-only export PropertyName x) = %v, want nil", got)
	}
	got := refs.References(sym)
	if len(got) != 1 || got[0] != writeIdent {
		t.Fatalf("References = %v, want [%v] (only the `x = 1` write)", got, writeIdent)
	}
}

func TestRefStoreTypeOnlyExportCheckerFallbackRespectsMeaningAndShadowing(t *testing.T) {
	// A local value-only HTMLElement does not shadow the outer/global
	// HTMLElement type. A broad checker lookup would pick the local value and
	// either misfile the reference or lose the real type target; the requested
	// declaration-space meaning must participate in the checker scope walk.
	sourceFile, refs, done := newCheckedRefStore(t,
		"let HTMLElement;\nexport type { HTMLElement as LocalHTMLElement };\n")
	defer done()

	occurrences := identifiers(sourceFile.AsNode(), "HTMLElement")
	if len(occurrences) != 2 {
		t.Fatalf("expected 2 occurrences of HTMLElement, got %d", len(occurrences))
	}
	declIdent, specifierIdent := occurrences[0], occurrences[1]
	localValue := declIdent.Parent.Symbol()
	if localValue == nil {
		t.Fatal("declaration identifier has no bound symbol")
	}

	target := refs.Resolve(specifierIdent)
	if target == nil {
		t.Fatal("Resolve(type-only export HTMLElement) = nil, want the global type symbol")
	}
	if target == localValue {
		t.Fatal("type-only export resolved to the shadowing local value symbol")
	}
	if target.Flags&(ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) == 0 {
		t.Fatalf("resolved checker symbol flags = %v, want a type-capable symbol", target.Flags)
	}
	if got := refs.References(localValue); len(got) != 0 {
		t.Fatalf("References(local value) = %v, want none", got)
	}
	got := refs.References(target)
	if len(got) != 1 || got[0] != specifierIdent {
		t.Fatalf("References(global type) = %v, want [%v]", got, specifierIdent)
	}
}

func TestRefStoreNestedLocalExports(t *testing.T) {
	tests := []struct {
		name            string
		source          string
		identifier      string
		occurrenceCount int
		targetIndex     int
		localValueIndex int
	}{
		{
			name: "type export skips a value shadow",
			source: `
type X = {};
namespace N {
	let X;
	export type { X };
}`,
			identifier:      "X",
			occurrenceCount: 3,
			targetIndex:     0,
			localValueIndex: 1,
		},
		{
			name: "type export keeps its local type",
			source: `
namespace N {
	type X = {};
	export type { X };
}`,
			identifier:      "X",
			occurrenceCount: 2,
			targetIndex:     0,
			localValueIndex: -1,
		},
		{
			name: "deeply nested type export finds its lexical type",
			source: `
namespace Outer {
	type X = {};
	export namespace Inner {
		let X;
		export type { X };
	}
}`,
			identifier:      "X",
			occurrenceCount: 3,
			targetIndex:     0,
			localValueIndex: 1,
		},
		{
			name: "type export rejects its own value-only alias",
			source: `
namespace N {
	let x;
	export type { x };
}`,
			identifier:      "x",
			occurrenceCount: 2,
			targetIndex:     -1,
			localValueIndex: 0,
		},
		{
			name: "ordinary export keeps its local value",
			source: `
namespace N {
	const x = 1;
	export { x };
}`,
			identifier:      "x",
			occurrenceCount: 2,
			targetIndex:     0,
			localValueIndex: -1,
		},
		{
			name: "ordinary export finds its outer value",
			source: `
const x = 1;
namespace N {
	export { x };
}`,
			identifier:      "x",
			occurrenceCount: 2,
			targetIndex:     0,
			localValueIndex: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile, refs := newBoundRefStore(t, "/nested-local-export.ts", core.ScriptKindTS, test.source)
			occurrences := identifiers(sourceFile.AsNode(), test.identifier)
			if len(occurrences) != test.occurrenceCount {
				t.Fatalf("identifier occurrences = %d, want %d", len(occurrences), test.occurrenceCount)
			}
			specifier := occurrences[len(occurrences)-1]

			if test.targetIndex < 0 {
				if got := refs.Resolve(specifier); got != nil {
					t.Fatalf("Resolve(export specifier) = %v, want nil", got)
				}
			} else {
				target := occurrences[test.targetIndex].Parent.Symbol()
				if target == nil {
					t.Fatal("target declaration identifier has no bound symbol")
				}
				if got := refs.Resolve(specifier); got != target {
					t.Fatalf("Resolve(export specifier) = %v, want %v", got, target)
				}
				if got := refs.References(target); len(got) != 1 || got[0] != specifier {
					t.Fatalf("References(target) = %v, want [%v]", got, specifier)
				}
			}

			if test.localValueIndex >= 0 {
				localValue := occurrences[test.localValueIndex].Parent.Symbol()
				if localValue == nil {
					t.Fatal("local value declaration identifier has no bound symbol")
				}
				if got := refs.References(localValue); len(got) != 0 {
					t.Fatalf("References(local value) = %v, want none", got)
				}
			}
		})
	}
}

func TestRefStoreNestedTypeOnlyExportsWithChecker(t *testing.T) {
	for _, test := range []struct {
		name       string
		identifier string
		source     string
		wantOuter  bool
	}{
		{
			name:       "skips own alias and resolves global type",
			identifier: "HTMLElement",
			source: `
namespace N {
	let HTMLElement;
	export type { HTMLElement };
}`,
			wantOuter: true,
		},
		{
			name:       "does not return own alias without outer type",
			identifier: "localValue",
			source: `
namespace N {
	let localValue;
	export type { localValue };
}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceFile, refs, done := newCheckedRefStore(t, test.source)
			defer done()

			occurrences := identifiers(sourceFile.AsNode(), test.identifier)
			if len(occurrences) != 2 {
				t.Fatalf("identifier occurrences = %d, want 2", len(occurrences))
			}
			localValue := occurrences[0].Parent.Symbol()
			specifier := occurrences[1]
			if localValue == nil {
				t.Fatal("local value declaration identifier has no bound symbol")
			}

			target := refs.Resolve(specifier)
			if !test.wantOuter {
				if target != nil {
					t.Fatalf("Resolve(type-only export) = %v, want nil", target)
				}
			} else {
				if target == nil || target == localValue || isOwnExportAlias(specifier, target) {
					t.Fatalf("Resolve(type-only export) = %v, want the outer type", target)
				}
				if target.Flags&(ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) == 0 {
					t.Fatalf("resolved checker symbol flags = %v, want a type-capable symbol", target.Flags)
				}
				if got := refs.References(target); len(got) != 1 || got[0] != specifier {
					t.Fatalf("References(outer type) = %v, want [%v]", got, specifier)
				}
			}
			if got := refs.References(localValue); len(got) != 0 {
				t.Fatalf("References(local value) = %v, want none", got)
			}
		})
	}
}

func TestRefStoreLocalExportCheckerFallbackResolvesGlobalValue(t *testing.T) {
	// The meaning-aware ExportSpecifier fallback also preserves ordinary
	// dual-space exports of globals; it must not be limited to type-only
	// syntax.
	sourceFile, refs, done := newCheckedRefStore(t,
		"export { window as browserWindow };\n")
	defer done()

	occurrences := identifiers(sourceFile.AsNode(), "window")
	if len(occurrences) != 1 {
		t.Fatalf("expected 1 occurrence of window, got %d", len(occurrences))
	}
	specifierIdent := occurrences[0]
	target := refs.Resolve(specifierIdent)
	if target == nil {
		t.Fatal("Resolve(local export window) = nil, want the global value symbol")
	}
	if target.Flags&ast.SymbolFlagsValue == 0 {
		t.Fatalf("resolved checker symbol flags = %v, want a value-capable symbol", target.Flags)
	}
	got := refs.References(target)
	if len(got) != 1 || got[0] != specifierIdent {
		t.Fatalf("References(global value) = %v, want [%v]", got, specifierIdent)
	}
}

func TestRefStoreGlobalNameReferenceSkipsOwnExportAlias(t *testing.T) {
	const typeMeaning = ast.SymbolFlagsType | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	for _, test := range []struct {
		name   string
		source string
		want   bool
	}{
		{
			name:   "implicit global type target",
			source: `export type { Record as Safe };`,
			want:   true,
		},
		{
			name:   "authored local type target",
			source: `type Record = {}; export type { Record as Safe };`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceFile, refs := newBoundRefStore(t, "/export-global-name.ts", core.ScriptKindTS, test.source)
			occurrences := identifiers(sourceFile.AsNode(), "Record")
			specifier := occurrences[len(occurrences)-1]
			if got := refs.IsGlobalNameReference(specifier, "Record", typeMeaning); got != test.want {
				t.Fatalf("IsGlobalNameReference(Record) = %v, want %v", got, test.want)
			}
			if safe := identifiers(sourceFile.AsNode(), "Safe"); len(safe) == 1 &&
				refs.IsGlobalNameReference(safe[0], "Safe", typeMeaning) {
				t.Fatal("exported alias label Safe was treated as a global reference")
			}
		})
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

func TestRefStoreResolveInFileNeverUsesCheckerFallback(t *testing.T) {
	// ResolveInFile must preserve local/import scope resolution while refusing
	// the same lib.dom symbol that Resolve finds through its checker fallback.
	sourceFile, refs, done := newCheckedRefStore(t, "export {}; const local = 1; local; window;")
	defer done()

	localOccurrences := identifiers(sourceFile.AsNode(), "local")
	if len(localOccurrences) != 2 {
		t.Fatalf("expected declaration + reference for local, got %d", len(localOccurrences))
	}
	if got := refs.ResolveInFile(localOccurrences[1]); got == nil {
		t.Fatal("ResolveInFile(local reference) = nil, want the current-file binding")
	}

	windowOccurrences := identifiers(sourceFile.AsNode(), "window")
	if len(windowOccurrences) != 1 {
		t.Fatalf("expected one window reference, got %d", len(windowOccurrences))
	}
	window := windowOccurrences[0]
	if got := refs.ResolveInFile(window); got != nil {
		t.Fatalf("ResolveInFile(window) = %v, want nil for a lib.d.ts binding", got)
	}
	if got := refs.Resolve(window); got == nil {
		t.Fatal("Resolve(window) = nil, want proof that the checker fallback is available")
	}
}

func TestRefStoreImplicitFileArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fileName string
		source   string
		want     []bool
	}{
		{
			name:     "commonjs top level arrow and function",
			fileName: "/file.cjs",
			source:   "arguments; (() => arguments)(); function f() { return arguments; }",
			want:     []bool{true, true, true},
		},
		{
			name:     "module top level arrow and function",
			fileName: "/file.js",
			source:   "arguments; (() => arguments)(); function f() { return arguments; }",
			want:     []bool{false, false, true},
		},
		{
			name:     "local declaration wins",
			fileName: "/file.cjs",
			source:   "let arguments = 1; arguments;",
			want:     []bool{false, true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			sourceFile, refs := newBoundRefStore(t, test.fileName, core.ScriptKindJS, test.source)
			occurrences := identifiers(sourceFile.AsNode(), "arguments")
			if len(occurrences) != len(test.want) {
				t.Fatalf("found %d arguments identifiers, want %d", len(occurrences), len(test.want))
			}
			for i, node := range occurrences {
				if got := refs.IsDefinedInFile(node); got != test.want[i] {
					t.Errorf("IsDefinedInFile(arguments occurrence %d) = %v, want %v", i, got, test.want[i])
				}
			}
		})
	}
}

func TestRefStoreCommonJSGlobalReference(t *testing.T) {
	for _, test := range []struct {
		name   string
		source string
		want   bool
	}{
		{name: "wrapper global", source: "exports.foo = 1;", want: true},
		{name: "authored declaration", source: "const exports = {}; exports.foo = 1;"},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceFile, refs := newBoundRefStore(t, "/file.cjs", core.ScriptKindJS, test.source)
			exports := identifiers(sourceFile.AsNode(), "exports")
			reference := exports[len(exports)-1]
			if got := refs.IsGlobalReference(reference); got != test.want {
				t.Fatalf("IsGlobalReference(exports) = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRefStoreGlobalReferenceProgramScopeSemantics(t *testing.T) {
	tests := []struct {
		name       string
		fileName   string
		scriptKind core.ScriptKind
		sourceType string
		source     string
		want       bool
	}{
		{
			name:       "authored import retained beside JSDoc import",
			fileName:   "/file.js",
			scriptKind: core.ScriptKindJS,
			sourceType: "module",
			source:     "/** @import { Promise } from \"x\" */\nimport { Promise } from \"y\";\nPromise;",
		},
		{
			name:       "type-only module declaration leaves value global",
			fileName:   "/file.ts",
			scriptKind: core.ScriptKindTS,
			sourceType: "module",
			source:     "interface Promise {} Promise;",
			want:       true,
		},
		{
			name:       "namespace module declaration binds value reference",
			fileName:   "/file.ts",
			scriptKind: core.ScriptKindTS,
			sourceType: "module",
			source:     "namespace Promise {} Promise;",
			want:       false,
		},
		{
			name:       "dotted namespace leaves built-in global uncontrollable",
			fileName:   "/file.ts",
			scriptKind: core.ScriptKindTS,
			sourceType: "script",
			source:     "namespace Promise.Inner {} Promise;",
			want:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: test.fileName,
				Path:     tspath.Path(test.fileName),
			}, test.source, test.scriptKind)
			binder.BindSourceFile(sourceFile)
			_, refsInit, _ := ResolveLanguageDefaults(test.fileName, LanguageOptions{SourceType: test.sourceType})
			refs := NewRefStore(sourceFile, &core.CompilerOptions{}, nil, refsInit)
			occurrences := identifiers(sourceFile.AsNode(), "Promise")
			if len(occurrences) == 0 {
				t.Fatal("found no Promise identifier")
			}
			reference := occurrences[len(occurrences)-1]
			if got := refs.IsGlobalReference(reference); got != test.want {
				t.Fatalf("IsGlobalReference(Promise) = %v, want %v", got, test.want)
			}
		})
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

func TestRefStoreTypeOnlyHeritageQualifierSkipsValueShadow(t *testing.T) {
	for _, test := range []struct {
		name   string
		source string
	}{
		{
			name: "class implements",
			source: `
export {};
namespace N { export interface T {} }
function f() {
	let N = 0;
	class C implements N.T {}
}`,
		},
		{
			name: "interface extends",
			source: `
export {};
namespace N { export interface T {} }
function f() {
	let N = 0;
	interface I extends N.T {}
}`,
		},
		{
			name: "nested qualifier",
			source: `
export {};
namespace N { export namespace Inner { export interface T {} } }
function f() {
	let N = 0;
	class C implements N.Inner.T {}
}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceFile, refs := newBoundRefStore(t, "/type-only-heritage.ts", core.ScriptKindTS, test.source)
			occurrences := identifiers(sourceFile.AsNode(), "N")
			if len(occurrences) != 3 {
				t.Fatalf("identifier occurrences = %d, want 3", len(occurrences))
			}

			namespace := occurrences[0].Parent.Symbol()
			localValue := occurrences[1].Parent.Symbol()
			use := occurrences[2]
			if namespace == nil || localValue == nil {
				t.Fatal("declaration identifier has no bound symbol")
			}
			if got := refs.Resolve(use); got != namespace {
				t.Fatalf("Resolve(heritage qualifier) = %v, want namespace %v", got, namespace)
			}
			if got := refs.References(namespace); len(got) != 1 || got[0] != use {
				t.Fatalf("References(namespace) = %v, want [%v]", got, use)
			}
			if got := refs.References(localValue); len(got) != 0 {
				t.Fatalf("References(local value) = %v, want none", got)
			}
		})
	}
}

func TestRefStoreExpressionShapedTypeQualifierPreservesValueContexts(t *testing.T) {
	for _, test := range []struct {
		name   string
		source string
	}{
		{
			// Unlike `implements` and an interface's `extends`, a class
			// `extends` expression executes at runtime.
			name: "class extends",
			source: `
export {};
namespace N { export class Base {} }
function f() {
	const N = { Base: class {} };
	class C extends N.Base {}
	}`,
		},
		{
			// A type query names the runtime value even though the query itself
			// is type syntax.
			name: "type query",
			source: `
export {};
namespace N { export const value = 1; }
function f() {
	const N = { value: 1 };
	type T = typeof N.value;
}`,
		},
		{
			// An arbitrary expression inside `implements` is still a value use;
			// only a pure dotted entity name acts as a namespace qualifier.
			name: "call expression in implements",
			source: `
export {};
namespace N { export interface T {} }
function f() {
	function N() { return { T: class {} }; }
	class C implements N().T {}
}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceFile, refs := newBoundRefStore(t, "/value-qualifier.ts", core.ScriptKindTS, test.source)
			occurrences := identifiers(sourceFile.AsNode(), "N")
			if len(occurrences) != 3 {
				t.Fatalf("identifier occurrences = %d, want 3", len(occurrences))
			}
			namespace := occurrences[0].Parent.Symbol()
			localValue := occurrences[1].Parent.Symbol()
			use := occurrences[2]
			if namespace == nil || localValue == nil {
				t.Fatal("declaration identifier has no bound symbol")
			}
			if got := refs.Resolve(use); got != localValue {
				t.Fatalf("Resolve(value qualifier) = %v, want local value %v", got, localValue)
			}
			if got := refs.References(localValue); len(got) != 1 || got[0] != use {
				t.Fatalf("References(local value) = %v, want [%v]", got, use)
			}
			if got := refs.References(namespace); len(got) != 0 {
				t.Fatalf("References(namespace) = %v, want none", got)
			}
		})
	}
}

func TestRefStoreTypeOnlyHeritageQualifierCheckerFallbackRespectsMeaning(t *testing.T) {
	// The binder cannot see namespaces declared in lib files. Its namespace
	// lookup must fall back to a meaning-aware checker lookup rather than let
	// GetSymbolAtLocation select the same-named local value.
	sourceFile, refs, done := newCheckedRefStore(t, `
export {};
function f() {
	let Intl = 0;
	class C implements Intl.CollatorOptions {}
}`)
	defer done()

	occurrences := identifiers(sourceFile.AsNode(), "Intl")
	if len(occurrences) != 2 {
		t.Fatalf("identifier occurrences = %d, want 2", len(occurrences))
	}
	localValue := occurrences[0].Parent.Symbol()
	use := occurrences[1]
	if localValue == nil {
		t.Fatal("local declaration identifier has no bound symbol")
	}
	target := refs.Resolve(use)
	if target == nil || target == localValue {
		t.Fatalf("Resolve(heritage qualifier) = %v, want the global Intl namespace", target)
	}
	if target.Flags&(ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) == 0 {
		t.Fatalf("resolved checker symbol flags = %v, want a namespace-capable symbol", target.Flags)
	}
	if got := refs.References(target); len(got) != 1 || got[0] != use {
		t.Fatalf("References(global namespace) = %v, want [%v]", got, use)
	}
	if got := refs.References(localValue); len(got) != 0 {
		t.Fatalf("References(local value) = %v, want none", got)
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
