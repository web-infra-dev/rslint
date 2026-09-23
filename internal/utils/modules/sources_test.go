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
			names = append(names, source.Specifier.Text())
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
		expressions = append(expressions, strings.TrimSpace(file.Text()[source.Specifier.Pos():source.Specifier.End()]))
		types = append(types, source.TypeOnly)
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
		expressions = append(expressions, strings.TrimSpace(file.Text()[source.Specifier.Pos():source.Specifier.End()]))
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
		names = append(names, source.Specifier.Text())
	}
	if want := []string{"parenthesized", "optional"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("require sources = %q, want %q", names, want)
	}
}
