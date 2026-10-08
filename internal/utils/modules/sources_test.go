package modules

import (
	"reflect"
	"strings"
	"testing"
)

func TestCollectEmptyStaticSources(t *testing.T) {
	file := parseModuleSpecifierCacheFile("import 'first'; import ''; export * from ''; export {x} from 'last';")
	for _, kinds := range []ReferenceKinds{ESModuleReferences, AllModuleReferences} {
		var names []string
		for _, source := range Collect(file, kinds) {
			names = append(names, source.Specifier().Text())
		}
		if want := []string{"first", "", "", "last"}; !reflect.DeepEqual(names, want) {
			t.Fatalf("kinds %d: sources = %q, want %q", kinds, names, want)
		}
	}
}

func TestCollectKeepsSourceExpressions(t *testing.T) {
	file := parseModuleSpecifierCacheFile("import type {X} from 'types'; import {type Y} from 'named'; export type * from 'exported'; import(('dep')); import('typed' as string); import(1); import(`template`); import(dynamic);")
	sources := Collect(file, ESModuleReferences)
	var expressions []string
	var types []bool
	for _, source := range sources {
		expressions = append(expressions, strings.TrimSpace(file.Text()[source.Specifier().Pos():source.Specifier().End()]))
		types = append(types, source.TypeOnly())
	}
	want := []string{"'types'", "'named'", "'exported'", "('dep')", "'typed' as string", "1", "`template`", "dynamic"}
	if !reflect.DeepEqual(expressions, want) {
		t.Fatalf("source expressions = %q, want %q", expressions, want)
	}
	if !reflect.DeepEqual(types, []bool{true, true, true, false, false, false, false, false}) {
		t.Fatalf("type-only flags = %v", types)
	}
}

func TestCollectNestedAndNonLiteralSources(t *testing.T) {
	file := parseModuleSpecifierCacheFile("declare module 'outer' { import x from 'inner'; } require(name); define(['amd', dynamic]);")
	sources := Collect(file, AllModuleReferences)
	var expressions []string
	for _, source := range sources {
		expressions = append(expressions, strings.TrimSpace(file.Text()[source.Specifier().Pos():source.Specifier().End()]))
	}
	if want := []string{"'inner'", "name", "'amd'", "dynamic"}; !reflect.DeepEqual(expressions, want) {
		t.Fatalf("sources = %q, want %q", expressions, want)
	}
	if Collect(nil, AllModuleReferences) != nil || Collect(file, 0) != nil {
		t.Fatal("empty selection returned sources")
	}
}

func TestCollectParenthesizedRequire(t *testing.T) {
	file := parseModuleSpecifierCacheFile(`(require)('parenthesized'); require?.('optional'); require('two', 'args'); (require as any)('asserted');`)
	var names []string
	for _, source := range Collect(file, CommonJSReferences) {
		names = append(names, source.Specifier().Text())
	}
	if want := []string{"parenthesized", "optional"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("require sources = %q, want %q", names, want)
	}
}

func TestCollectImportAttributes(t *testing.T) {
	file := parseModuleSpecifierCacheFile(`
		import first from "./data.json" with { type: "json", mode: "strict" };
		import second from "./data.json" with { "mode": "strict", "type": "json" };
		export { value } from "./data.json" with { type: "text" };
		import("./data.json", { with: { type: "json" } });
		import("./data.json", options);
		import duplicate from "./data.json" with { type: "json", type: "text" };
	`)
	sources := Collect(file, ESModuleReferences)
	if len(sources) != 6 {
		t.Fatalf("collected %d sources, want 6", len(sources))
	}
	if sources[0].Attributes().Key() == "" || sources[0].Attributes().Key() != sources[1].Attributes().Key() {
		t.Fatalf("reordered attributes have keys %q and %q", sources[0].Attributes().Key(), sources[1].Attributes().Key())
	}
	if value, ok := sources[2].Attributes().Value("type"); !ok || value != "text" {
		t.Fatalf("export type attribute = (%q, %v), want (text, true)", value, ok)
	}
	if value, ok := sources[3].Attributes().Value("type"); !ok || value != "json" {
		t.Fatalf("dynamic import type attribute = (%q, %v), want (json, true)", value, ok)
	}
	if sources[4].Attributes().State != AttributesDynamic {
		t.Fatalf("dynamic options state = %v, want AttributesDynamic", sources[4].Attributes().State)
	}
	if sources[0].Attributes().Key() == sources[2].Attributes().Key() {
		t.Fatal("different attribute values produced the same key")
	}
	if sources[5].Attributes().State != AttributesInvalid {
		t.Fatalf("duplicate attributes state = %v, want AttributesInvalid", sources[5].Attributes().State)
	}
}

func TestDynamicImportAttributePresence(t *testing.T) {
	file := parseModuleSpecifierCacheFile(`
		import("./b.js");
		import("./b.js", {});
		import("./b.js", { signal });
		import("./b.js", { signal: value });
		import("./b.js", { ["signal"]: value });
		import("./b.js", options);
		import("./b.js", { with: attrs });
		import("./b.js", { with: attrs, with: other });
		import("./b.js", { ...options });
		import("./b.js", { [key]: value });
	`)
	sources := Collect(file, ESModuleReferences)
	want := []AttributeState{
		AttributesNone,
		AttributesNone,
		AttributesNone,
		AttributesNone,
		AttributesNone,
		AttributesDynamic,
		AttributesDynamic,
		AttributesDynamic,
		AttributesDynamic,
		AttributesDynamic,
	}
	if len(sources) != len(want) {
		t.Fatalf("collected %d sources, want %d", len(sources), len(want))
	}
	for i, state := range want {
		if got := sources[i].Attributes().State; got != state {
			t.Errorf("source %d state = %v, want %v", i, got, state)
		}
	}
}

func TestImportAttributesIgnoreTypeScriptWrappers(t *testing.T) {
	file := parseModuleSpecifierCacheFile(`
		import("a", { with: { type: "json" } } as const);
		import("a", { with: { type: "json" } satisfies object });
		import("a", { with: ({ type: "json" } as const) });
		import("a", { with: { type: "json" as const } });
		import("a", { with: { type: ("json" satisfies string) } });
		import("a", {} as const);
		import("a", { with: { type: "json" } }!);
		import json from "a" with { type: "json" };
	`)
	sources := Collect(file, ESModuleReferences)
	if len(sources) != 8 {
		t.Fatalf("collected %d sources, want 8", len(sources))
	}
	want := sources[7].Attributes().Key()
	for i := 0; i < 5; i++ {
		if got := sources[i].Attributes(); got.State != AttributesStatic || got.Key() != want {
			t.Errorf("source %d = (%v, %q), want static %q", i, got.State, got.Key(), want)
		}
	}
	if got := sources[5].Attributes(); got.State != AttributesNone && !(got.State == AttributesStatic && got.Key() == "") {
		t.Errorf("empty options state = %v, want no attributes", got.State)
	}
	if got := sources[6].Attributes(); got.State != AttributesStatic || got.Key() != want {
		t.Errorf("non-null options = (%v, %q), want static %q", got.State, got.Key(), want)
	}
}

func TestImportAttributesPrototypeSetter(t *testing.T) {
	file := parseModuleSpecifierCacheFile(`
		import("./b", { with: { __proto__: null } });
		import("./b", { with: { "__proto__": null, type: "json" } });
		import("./b", { with: { ["__proto__"]: "x" } });
		import("./b", { __proto__: null, with: { type: "json" } });
		import("./b", { __proto__: proto, with: { type: "json" } });
		import json from "./b" with { type: "json" };
	`)
	sources := Collect(file, ESModuleReferences)
	if len(sources) != 6 {
		t.Fatalf("collected %d sources, want 6", len(sources))
	}
	if got := sources[0].Attributes(); got.State != AttributesStatic || got.Key() != "" || len(got.Entries()) != 0 {
		t.Errorf("proto-only set = (%v, %q), want empty static", got.State, got.Key())
	}
	json := sources[5].Attributes().Key()
	if got := sources[1].Attributes(); got.State != AttributesStatic || got.Key() != json {
		t.Errorf("proto plus type = (%v, %q), want static %q", got.State, got.Key(), json)
	}
	if value, ok := sources[2].Attributes().Value("__proto__"); !ok || value != "x" {
		t.Errorf("computed __proto__ = (%q, %v), want ordinary attribute (x, true)", value, ok)
	}
	if got := sources[3].Attributes(); got.State != AttributesStatic || got.Key() != json {
		t.Errorf("null-prototype options = (%v, %q), want static %q", got.State, got.Key(), json)
	}
	if got := sources[4].Attributes().State; got != AttributesDynamic {
		t.Errorf("non-null prototype options state = %v, want AttributesDynamic", got)
	}
}
