package rule

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
)

func TestRefStoreReferenceFactsAccessAndSignature(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/reference-facts.ts", core.ScriptKindTS, `
export {};
let target = 0;
const key = "key", fallback = 1, source = {}, delta = 1, values = [];
({ [key]: target = fallback } = source);
target += delta;
target++;
for (target of values) {}
type TargetType = typeof target;
`)

	var got []ReferenceFacts
	for _, node := range identifiers(sourceFile.AsNode(), "target") {
		if facts, ok := refs.ReferenceFacts(node); ok {
			got = append(got, facts)
		}
	}
	if len(got) != 5 {
		t.Fatalf("reference facts = %d, want 5", len(got))
	}
	wantAccess := []ReferenceAccess{
		ReferenceAccessWrite,
		ReferenceAccessRead | ReferenceAccessWrite,
		ReferenceAccessRead | ReferenceAccessWrite,
		ReferenceAccessWrite,
		ReferenceAccessRead,
	}
	for index, facts := range got {
		if facts.Access != wantAccess[index] {
			t.Errorf("reference %d access = %v, want %v", index, facts.Access, wantAccess[index])
		}
		if facts.Node == nil || facts.SyntacticBlockContainer == nil || facts.SyntacticScopeContainer == nil {
			t.Errorf("reference %d omitted its node/container facts: %#v", index, facts)
		}
	}
	last := got[len(got)-1]
	if last.Syntax&ReferenceSyntaxTypeQuery == 0 || last.Syntax&ReferenceSyntaxTypePosition == 0 {
		t.Fatalf("typeof reference syntax = %v, want type-query and type-position", last.Syntax)
	}
	if last.Space&ReferenceSpaceValue == 0 {
		t.Fatalf("typeof reference space = %v, want value-capable", last.Space)
	}
}

func TestRefStoreReferenceFactsSeeThroughSatisfiesAssignmentTargets(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/satisfies-write.ts", core.ScriptKindTS, `
let target = 0;
(target satisfies number) = 1;
(target satisfies number) += 1;
`)
	target := identifiers(sourceFile.AsNode(), "target")
	if len(target) != 3 {
		t.Fatalf("target occurrences = %d, want declaration and two writes", len(target))
	}
	want := []ReferenceAccess{
		ReferenceAccessWrite,
		ReferenceAccessRead | ReferenceAccessWrite,
	}
	for index, node := range target[1:] {
		facts, ok := refs.ReferenceFacts(node)
		if !ok || facts.Access != want[index] {
			t.Errorf("satisfies target %d facts = %#v, %v; want access %v", index, facts, ok, want[index])
		}
	}
}

func TestRefStoreReferenceFactsDistinguishParameterEnvironment(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/parameter-environment.ts", core.ScriptKindTS, `
export {};
const outer = 1;
function f(value = outer) {
  var outer = 2;
  return outer;
}
`)
	outer := identifiers(sourceFile.AsNode(), "outer")
	if len(outer) != 4 {
		t.Fatalf("outer occurrences = %d, want 4", len(outer))
	}
	initializerFacts, ok := refs.ReferenceFacts(outer[1])
	if !ok || !initializerFacts.OutsideFunctionBody {
		t.Fatalf("parameter initializer facts = %#v, want function-signature reference", initializerFacts)
	}
	bodyFacts, ok := refs.ReferenceFacts(outer[3])
	if !ok || bodyFacts.OutsideFunctionBody {
		t.Fatalf("body facts = %#v, want body reference", bodyFacts)
	}
	if initializerFacts.NearestFunction == nil || initializerFacts.NearestFunction != bodyFacts.NearestFunction {
		t.Fatal("initializer and body should share the enclosing function while retaining environment distinction")
	}

	topLevel := outer[0].Parent.Symbol()
	bodySymbol := outer[2].Parent.Symbol()
	if topLevel == nil || bodySymbol == nil {
		t.Fatal("fixture declarations were not bound")
	}
	if refs.Resolve(outer[1]) != topLevel {
		t.Fatal("parameter initializer resolved to the function-body var instead of the outer binding")
	}
	if refs.Resolve(outer[3]) != bodySymbol {
		t.Fatal("body reference did not resolve to the function-body var")
	}
	if got := refs.References(topLevel); len(got) != 1 || got[0] != outer[1] {
		t.Fatalf("outer binding references = %v, want parameter initializer", got)
	}
	if got := refs.References(bodySymbol); len(got) != 1 || got[0] != outer[3] {
		t.Fatalf("body binding references = %v, want return reference", got)
	}
}

func TestRefStoreReferenceFactsMarkExpressionShapedTypePositions(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/type-position-facts.ts", core.ScriptKindTS, `
namespace N { export interface T {}; export class Base {} }
class Implements implements N.T {}
interface Extends extends N.T {}
class RuntimeExtends extends N.Base {}
type Foo = {};
export type { Foo };
export { type Foo as Alias };
`)

	namespaceUses := identifiers(sourceFile.AsNode(), "N")
	if len(namespaceUses) != 4 {
		t.Fatalf("N occurrences = %d, want declaration plus three uses", len(namespaceUses))
	}
	for index, node := range namespaceUses[1:] {
		facts, ok := refs.ReferenceFacts(node)
		if !ok {
			t.Fatalf("N use %d has no reference facts", index)
		}
		wantType := index < 2
		if got := facts.Syntax&ReferenceSyntaxTypePosition != 0; got != wantType {
			t.Errorf("N use %d type-position = %v, want %v (facts=%#v)", index, got, wantType, facts)
		}
	}

	fooUses := identifiers(sourceFile.AsNode(), "Foo")
	if len(fooUses) != 3 {
		t.Fatalf("Foo occurrences = %d, want declaration plus two local type exports", len(fooUses))
	}
	for index, node := range fooUses[1:] {
		facts, ok := refs.ReferenceFacts(node)
		if !ok || facts.Syntax&(ReferenceSyntaxLocalExport|ReferenceSyntaxTypePosition) !=
			ReferenceSyntaxLocalExport|ReferenceSyntaxTypePosition || facts.Space&ReferenceSpaceValue != 0 {
			t.Errorf("Foo type export %d facts = %#v, %v", index, facts, ok)
		}
	}
}

func TestRefStoreReferenceFactsKeepComputedNamesSyntactic(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/computed-name-facts.ts", core.ScriptKindTS, `
const key = "method";
class C { [key]() {} }
`)
	key := identifiers(sourceFile.AsNode(), "key")
	if len(key) != 2 {
		t.Fatalf("key occurrences = %d, want declaration and computed-name use", len(key))
	}
	facts, ok := refs.ReferenceFacts(key[1])
	if !ok || facts.NearestFunction == nil || facts.NearestFunction.Kind != ast.KindMethodDeclaration ||
		!facts.OutsideFunctionBody {
		t.Fatalf("computed-name facts = %#v, %v; want exact outside-method-body syntax", facts, ok)
	}
}

func TestRefStoreJSXReferenceFactsAreExplicitOnly(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/reference-facts.tsx", core.ScriptKindTSX, `
import React from "react";
const Component = () => null;
declare const NS: any;
const view = <><Component /><NS.Component /><div /><Foo-bar /><foo:bar /></>;
`)

	tests := []struct {
		name      string
		wantFacts int
		wantJSX   bool
	}{
		{name: "Component", wantFacts: 1, wantJSX: true},
		{name: "NS", wantFacts: 1, wantJSX: true},
		{name: "div"},
		{name: "Foo-bar"},
		{name: "foo"},
		{name: "bar"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			count := 0
			for _, node := range identifiers(sourceFile.AsNode(), testCase.name) {
				facts, ok := refs.ReferenceFacts(node)
				if !ok {
					continue
				}
				count++
				if testCase.wantJSX && facts.Syntax&ReferenceSyntaxJSXTag == 0 {
					t.Errorf("%s facts omit JSX-tag syntax", testCase.name)
				}
			}
			if count != testCase.wantFacts {
				t.Fatalf("%s reference facts = %d, want %d", testCase.name, count, testCase.wantFacts)
			}
		})
	}
	react := identifiers(sourceFile.AsNode(), "React")
	if len(react) != 1 {
		t.Fatalf("React occurrences = %d, want only the import binding", len(react))
	}
	if binding, ok := refs.ImportBinding(react[0]); !ok || binding.ImportedName != "default" {
		t.Fatalf("React import binding = %#v, %v", binding, ok)
	}
	if got := refs.References(react[0].Parent.Symbol()); len(got) != 0 {
		t.Fatalf("React references = %v, want no synthesized JSX factory reference", got)
	}
}

func TestRefStoreImportBindingsPreserveLocalAliases(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/import-bindings.ts", core.ScriptKindTS, `
import Default, * as NS from "m";
import { Foo, Bar as Baz, type Qux as Q } from "m";
import type TypeDefault from "m";
import Alias = require("m");
Default(); NS.value; Foo(); Baz();
`)

	bindings := refs.ImportBindings()
	if len(bindings) != 7 {
		t.Fatalf("import bindings = %d, want 7: %#v", len(bindings), bindings)
	}
	want := []struct {
		local    string
		imported string
		kind     ImportBindingKind
		typeOnly bool
	}{
		{local: "Default", imported: "default", kind: ImportBindingDefault},
		{local: "NS", imported: "*", kind: ImportBindingNamespace},
		{local: "Foo", imported: "Foo", kind: ImportBindingNamed},
		{local: "Baz", imported: "Bar", kind: ImportBindingNamed},
		{local: "Q", imported: "Qux", kind: ImportBindingNamed, typeOnly: true},
		{local: "TypeDefault", imported: "default", kind: ImportBindingDefault, typeOnly: true},
		{local: "Alias", kind: ImportBindingEquals},
	}
	for index, expected := range want {
		got := bindings[index]
		if got.LocalName.Text() != expected.local || got.ImportedName != expected.imported ||
			got.Kind != expected.kind || got.TypeOnly != expected.typeOnly {
			t.Errorf("binding %d = (%s, %s, %v, typeOnly=%v), want (%s, %s, %v, typeOnly=%v)",
				index, got.LocalName.Text(), got.ImportedName, got.Kind, got.TypeOnly,
				expected.local, expected.imported, expected.kind, expected.typeOnly)
		}
		if got.Symbol == nil || got.Declaration == nil || got.Specifier == nil ||
			got.ModuleSpecifier == nil || got.ModuleSpecifier.Text() != "m" {
			t.Errorf("binding %d omitted syntax provenance: %#v", index, got)
		}
	}

	defaultOccurrences := identifiers(sourceFile.AsNode(), "Default")
	if len(defaultOccurrences) != 2 {
		t.Fatalf("Default occurrences = %d, want 2", len(defaultOccurrences))
	}
	fromReference, ok := refs.ImportBinding(defaultOccurrences[1])
	if !ok || fromReference.LocalName != defaultOccurrences[0] {
		t.Fatalf("ImportBinding(reference) = %#v, %v; want Default declaration", fromReference, ok)
	}
	bar := identifiers(sourceFile.AsNode(), "Bar")
	if len(bar) != 1 {
		t.Fatalf("Bar occurrences = %d, want imported-side name only", len(bar))
	}
	if _, ok := refs.ImportBinding(bar[0]); ok {
		t.Fatal("imported-side property name must not be treated as a local import binding")
	}
	if bindings[2].Symbol == bindings[3].Symbol {
		t.Fatal("different local aliases were merged by import provenance")
	}
}

func TestRefStoreImportBindingsDoNotMergeAliasesOfSameTarget(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/import-alias-identity.ts", core.ScriptKindTS, `
import { Foo as A, Foo as B } from "m";
type TA = A;
A();
B();
`)
	a := identifiers(sourceFile.AsNode(), "A")
	b := identifiers(sourceFile.AsNode(), "B")
	if len(a) != 3 || len(b) != 2 {
		t.Fatalf("alias occurrences = A:%d B:%d, want 3 and 2", len(a), len(b))
	}
	aBinding, aOK := refs.ImportBinding(a[0])
	bBinding, bOK := refs.ImportBinding(b[0])
	if !aOK || !bOK || aBinding.Symbol == nil || bBinding.Symbol == nil || aBinding.Symbol == bBinding.Symbol {
		t.Fatalf("local import identities were not kept separate: A=%#v B=%#v", aBinding, bBinding)
	}
	if got := refs.References(aBinding.Symbol); len(got) != 2 || got[0] != a[1] || got[1] != a[2] {
		t.Fatalf("A references = %v, want type and value uses", got)
	}
	if got := refs.References(bBinding.Symbol); len(got) != 1 || got[0] != b[1] {
		t.Fatalf("B references = %v, want only B call", got)
	}
	if typeFacts, ok := refs.ReferenceFacts(a[1]); !ok || typeFacts.Space&ReferenceSpaceType == 0 {
		t.Fatalf("type alias reference facts = %#v, %v", typeFacts, ok)
	}
	if valueFacts, ok := refs.ReferenceFacts(a[2]); !ok || valueFacts.Space&ReferenceSpaceValue == 0 {
		t.Fatalf("call reference facts = %#v, %v", valueFacts, ok)
	}
}

func TestRefStoreImportBindingsIncludeReparsedJSDocImports(t *testing.T) {
	_, refs := newBoundRefStore(t, "/jsdoc-import.js", core.ScriptKindJS, `
/** @import { Source as Local } from "module-name" */
/** @type {Local} */
let value;
`)
	bindings := refs.ImportBindings()
	if len(bindings) != 1 {
		t.Fatalf("JSDoc import bindings = %#v, want one Local alias", bindings)
	}
	binding := bindings[0]
	if binding.LocalName.Text() != "Local" || binding.ImportedName != "Source" ||
		!binding.TypeOnly || binding.ModuleSpecifier == nil || binding.ModuleSpecifier.Text() != "module-name" {
		t.Fatalf("JSDoc import binding = %#v, want type-only Source as Local provenance", binding)
	}
}

func TestRefStoreBindingDeclarationsKeepOccurrences(t *testing.T) {
	_, refs := newBoundRefStore(t, "/binding-declarations.ts", core.ScriptKindTS, `
export {};
interface Merged {}
const Merged = 1;
const { nested: local = Merged } = source;
function named(parameter: typeof Merged) { return local + parameter; }
`)

	declarations := refs.BindingDeclarations()
	byName := make(map[string][]BindingDeclarationFacts)
	for _, declaration := range declarations {
		byName[declaration.Name.Text()] = append(byName[declaration.Name.Text()], declaration)
		if declaration.Declaration == nil || declaration.RootDeclaration == nil ||
			declaration.SyntacticBlockContainer == nil || declaration.SyntacticScopeContainer == nil {
			t.Errorf("declaration omitted exact occurrence/container facts: %#v", declaration)
		}
	}
	if len(byName["Merged"]) != 2 {
		t.Fatalf("Merged declaration occurrences = %d, want interface and const", len(byName["Merged"]))
	}
	if len(byName["local"]) != 1 || byName["local"][0].Declaration.Kind != ast.KindBindingElement {
		t.Fatalf("local declaration = %#v, want BindingElement occurrence", byName["local"])
	}
	if len(byName["parameter"]) != 1 || !byName["parameter"][0].OutsideFunctionBody {
		t.Fatalf("parameter declaration = %#v, want function-signature occurrence", byName["parameter"])
	}
}

func TestRefStoreBindingDeclarationsPreserveNamedExpressionBindings(t *testing.T) {
	_, refs := newBoundRefStore(t, "/named-expression-bindings.ts", core.ScriptKindTS, `
export {};
const holder = function inner() { return inner; };
const ClassHolder = class Inner { method() { return Inner; } };
`)
	declarations := refs.BindingDeclarations()
	byName := make(map[string]BindingDeclarationFacts)
	for _, declaration := range declarations {
		byName[declaration.Name.Text()] = declaration
	}
	for _, name := range []string{"inner", "Inner"} {
		declaration, ok := byName[name]
		if !ok || declaration.Symbol == nil {
			t.Fatalf("named expression binding %q = %#v, want raw binder symbol", name, declaration)
		}
		references := refs.References(declaration.Symbol)
		if len(references) != 1 || references[0].Text() != name {
			t.Fatalf("%s references = %v, want its inner self-reference", name, references)
		}
	}
}

func TestRefStoreBindingDeclarationSeparatesParameterPropertyIdentities(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/parameter-property-binding.ts", core.ScriptKindTS, `
class C {
  constructor(public parameter: number) { consume(parameter); }
}
`)
	parameter := identifiers(sourceFile.AsNode(), "parameter")
	if len(parameter) != 2 {
		t.Fatalf("parameter occurrences = %d, want declaration and body use", len(parameter))
	}
	facts, ok := refs.BindingDeclaration(parameter[0])
	if !ok || facts.Symbol == nil || facts.MemberSymbol == nil || facts.Symbol == facts.MemberSymbol {
		t.Fatalf("parameter-property facts = %#v, %v; want distinct lexical/member identities", facts, ok)
	}
	if resolved := refs.Resolve(parameter[1]); resolved != facts.Symbol {
		t.Fatalf("body use resolved to %p, want lexical parameter symbol %p", resolved, facts.Symbol)
	}
	if got := refs.References(facts.Symbol); len(got) != 1 || got[0] != parameter[1] {
		t.Fatalf("lexical parameter references = %v, want body use", got)
	}
	if got := refs.References(facts.MemberSymbol); len(got) != 0 {
		t.Fatalf("member identity references = %v, want no lexical references", got)
	}

	var listed *BindingDeclarationFacts
	declarations := refs.BindingDeclarations()
	for index := range declarations {
		declaration := &declarations[index]
		if declaration.Name == parameter[0] {
			listed = declaration
			break
		}
	}
	if listed == nil || listed.Symbol != facts.Symbol || listed.MemberSymbol != facts.MemberSymbol {
		t.Fatalf("listed parameter-property facts = %#v, want direct-query identities %#v", listed, facts)
	}
}

func TestRefStoreBindingDeclarationsDocumentDeclarationBoundary(t *testing.T) {
	t.Run("JSDoc type alias", func(t *testing.T) {
		_, refs := newBoundRefStore(t, "/jsdoc-bindings.js", core.ScriptKindJS, `
/** @typedef {number} JSDocType */
`)
		declarations := refs.BindingDeclarations()
		if len(declarations) != 1 || declarations[0].Name.Text() != "JSDocType" || declarations[0].Symbol == nil {
			t.Fatalf("binding declarations = %#v, want re-parsed JSDoc type alias", declarations)
		}
	})

	t.Run("augmentation syntax", func(t *testing.T) {
		_, refs := newBoundRefStore(t, "/augmentation-bindings.d.ts", core.ScriptKindTS, `
export {};
export as namespace GlobalAlias;
declare global { var augmented: number; }
`)
		declarations := refs.BindingDeclarations()
		if len(declarations) != 1 || declarations[0].Name.Text() != "augmented" {
			t.Fatalf("binding declarations = %#v, want augmented and no augmentation sentinels", declarations)
		}
	})
}

func TestRefStoreBindingDeclarationsExcludeSyntheticJSDocOverloadNames(t *testing.T) {
	_, refs := newBoundRefStore(t, "/jsdoc-overload.js", core.ScriptKindJS, `
/**
 * @overload
 * @param {string} value
 * @returns {string}
 */
/** @param {unknown} value */
function overloaded(value) { return value; }
`)
	count := 0
	for _, declaration := range refs.BindingDeclarations() {
		if declaration.Name.Text() == "overloaded" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("overloaded binding occurrences = %d, want only the authored implementation name", count)
	}
}

func TestRefStoreReferencesReconcileGlobalAugmentationIdentity(t *testing.T) {
	sourceFile, refs, cleanup := newCheckedRefStore(t, `
export {};
declare global {
  interface Map<K, V> { custom: true }
  type Inside = Map<string, string>;
}
type Outside = Map<string, string>;
`)
	defer cleanup()
	mapNames := identifiers(sourceFile.AsNode(), "Map")
	if len(mapNames) != 3 {
		t.Fatalf("Map occurrences = %d, want declaration plus inside/outside uses", len(mapNames))
	}
	declaration, ok := refs.BindingDeclaration(mapNames[0])
	resolved := refs.Resolve(mapNames[2])
	if !ok || declaration.Symbol == nil || resolved == nil || declaration.Symbol == resolved {
		t.Fatalf("augmentation identities = declaration %#v resolved %p, want distinct raw/merged symbols", declaration, resolved)
	}
	if got := refs.References(declaration.Symbol); len(got) != 2 || got[0] != mapNames[1] || got[1] != mapNames[2] {
		t.Fatalf("raw augmentation references = %v, want source-ordered inside/outside uses", got)
	}
	if got := refs.References(resolved); len(got) != 2 || got[0] != mapNames[1] || got[1] != mapNames[2] {
		t.Fatalf("merged augmentation references = %v, want same complete bucket", got)
	}
}

func TestRefStoreReferenceFactsDoNotRequireCheckerResolution(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/unresolved-reference-facts.ts", core.ScriptKindTS, `
export {};
window;
const object = { window: 1 };
`)
	window := identifiers(sourceFile.AsNode(), "window")
	if len(window) != 2 {
		t.Fatalf("window occurrences = %d, want unresolved use and property label", len(window))
	}
	if refs.Resolve(window[0]) != nil {
		t.Fatal("binder-only fixture unexpectedly resolved window")
	}
	if facts, ok := refs.ReferenceFacts(window[0]); !ok || facts.Node != window[0] {
		t.Fatalf("unresolved explicit reference facts = %#v, %v", facts, ok)
	}
	if _, ok := refs.ReferenceFacts(window[1]); ok {
		t.Fatal("property label was accepted as an explicit reference")
	}
}

func TestRefStoreReferenceFactsExcludeConstAssertionSentinel(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/const-assertion.ts", core.ScriptKindTS, `
const value = {} as const;
`)
	constNames := identifiers(sourceFile.AsNode(), "const")
	if len(constNames) != 1 {
		t.Fatalf("const assertion identifiers = %d, want one parser sentinel", len(constNames))
	}
	if _, ok := refs.ReferenceFacts(constNames[0]); ok {
		t.Fatal("const assertion sentinel was treated as a lexical reference")
	}
	if refs.Resolve(constNames[0]) != nil || refs.ResolveInFile(constNames[0]) != nil {
		t.Fatal("const assertion sentinel unexpectedly resolved")
	}
}

func TestRefStoreReferenceFactsDistinguishLocalAndReExports(t *testing.T) {
	sourceFile, refs := newBoundRefStore(t, "/export-reference-facts.ts", core.ScriptKindTS, `
const value = 1;
export { value as alias };
export { foreign as externalAlias } from "m";
`)
	value := identifiers(sourceFile.AsNode(), "value")
	if len(value) != 2 {
		t.Fatalf("value occurrences = %d, want declaration and local export", len(value))
	}
	facts, ok := refs.ReferenceFacts(value[1])
	if !ok || facts.Syntax&ReferenceSyntaxLocalExport == 0 ||
		facts.Space != ReferenceSpaceValue|ReferenceSpaceType|ReferenceSpaceNamespace {
		t.Fatalf("local export facts = %#v, %v", facts, ok)
	}
	for _, name := range []string{"alias", "foreign", "externalAlias"} {
		for _, node := range identifiers(sourceFile.AsNode(), name) {
			if _, ok := refs.ReferenceFacts(node); ok {
				t.Errorf("%s export label/re-export name was treated as a local reference", name)
			}
		}
	}
}
