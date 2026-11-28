package test_framework

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func TestResolveModuleImportSpecifierTypeOnlyPolicy(t *testing.T) {
	// The `type` modifier is recorded on the specifier when it is written
	// inline and on the enclosing clause when it heads the declaration. Both
	// name a binding that does not exist at runtime.
	for _, code := range []string{
		`import { type Mock } from "@rstest/core";`,
		`import type { Mock } from "@rstest/core";`,
	} {
		t.Run(code, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, code, core.ScriptKindTS)
			specifier := findFirstNodeOfKind(sourceFile, ast.KindImportSpecifier)
			if specifier == nil {
				t.Fatal("expected an import specifier")
			}

			modules := []string{"@rstest/core"}
			if _, _, ok := resolveModuleImportSpecifier(specifier, modules, false); ok {
				t.Fatal("function reference resolution accepted a type-only specifier")
			}
			name, _, ok := resolveModuleImportSpecifier(specifier, modules, true)
			if !ok || name != "Mock" {
				t.Fatalf("type reference resolution = %q, %v; want Mock, true", name, ok)
			}
		})
	}
}

func TestResolveFunctionReferenceForModules(t *testing.T) {
	modules := []string{"@jest/globals"}

	if name, node, _ := ResolveFunctionReferenceForModules(nil, "test", nil, nil, nil, modules); name != "" || node != nil {
		t.Fatalf("nil node = (%q, %v), want (\"\", nil)", name, node)
	}

	ident := parseFirstNodeOfKind(t, `test;`, ast.KindIdentifier)
	if name, node, _ := ResolveFunctionReferenceForModules(ident, "test", ident, nil, nil, modules); name != "" || node != nil {
		t.Fatalf("non-call node = (%q, %v), want (\"\", nil)", name, node)
	}

	_, regexCall := parseFirstCall(t, `/a/.test(text);`)
	if name, node, _ := ResolveFunctionReferenceForModules(regexCall, "test", nil, nil, nil, modules); name != "" || node != nil {
		t.Fatalf("non-identifier receiver call = (%q, %v), want (\"\", nil)", name, node)
	}

	_, testCall := parseFirstCall(t, `test("case", () => {});`)
	name, node, mode := ResolveFunctionReferenceForModules(testCall, "test", nil, nil, nil, modules)
	if name != "test" || node == nil || node.Kind != ast.KindIdentifier || node.AsIdentifier().Text != "test" || mode != ReferenceModeGlobal {
		t.Fatalf("identifier call without checker = (%q, %v, %v), want (\"test\", identifier, global)", name, node, mode)
	}
}

func TestIsImportEqualsOfModules(t *testing.T) {
	decl := parseFirstNodeOfKind(t, `import r = require("@rstest/core");`, ast.KindImportEqualsDeclaration)
	if !IsImportEqualsOfModules(decl, []string{"@rstest/core"}) {
		t.Fatal("expected matching import-equals declaration to return true")
	}
	if IsImportEqualsOfModules(decl, []string{"vitest"}) {
		t.Fatal("expected non-matching import-equals declaration to return false")
	}
	if IsImportEqualsOfModules(nil, []string{"@rstest/core"}) {
		t.Fatal("expected nil declaration to return false")
	}
}

func findRequireCallForModule(sourceFile *ast.SourceFile, moduleName string) *ast.Node {
	var reqCall *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindCallExpression && ast.IsRequireCall(node, false) {
			if args := node.AsCallExpression().Arguments; args != nil && len(args.Nodes) == 1 {
				if arg := ast.SkipParentheses(args.Nodes[0]); arg != nil &&
					arg.Kind == ast.KindStringLiteral && arg.AsStringLiteral().Text == moduleName {
					reqCall = node
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile.AsNode())
	return reqCall
}

func TestIsModuleRequireCallModulesIgnoresShadowedRequire(t *testing.T) {
	modules := []string{"@rstest/core"}

	for _, code := range []string{
		`const require = () => ({}); require("@rstest/core");`,
		`function run(require: any) { return require("@rstest/core"); }`,
		`if (true) { var require = () => ({}); } require("@rstest/core");`,
		`function f() { if (true) { var require = () => ({}); } return require("@rstest/core"); }`,
		`class C { static { var require = () => ({}); require("@rstest/core"); } }`,
		`namespace N { var require = () => ({}); require("@rstest/core"); }`,
		`if (true) function require() {} require("@rstest/core");`,
		`import { createRequire } from "node:module"; const { resolve: require } = createRequire(import.meta.url); require("@rstest/core");`,
		`import { createRequire } from "node:module"; let require = createRequire(import.meta.url); require = () => ({}); require("@rstest/core");`,
		// Same-scope TDZ self-reference must terminate cleanly without stack overflow.
		`const { createRequire } = require("node:module"); const require = createRequire(import.meta.url); require("@rstest/core");`,
		`const mod = require("node:module"); const require = mod.createRequire(import.meta.url); require("@rstest/core");`,
		// Hoisted var in block reassigned outside block.
		`import { createRequire } from "node:module"; let fn: any; { var require = createRequire(import.meta.url); fn = () => require("@rstest/core"); } require = () => ({});`,
		// Parameter shadowed by unexecuted hoisted var.
		`import { createRequire } from "node:module"; function run(require: any) { if (false) { var require = createRequire(import.meta.url); } return require("@rstest/core"); }`,
		// Unexecuted hoisted var + unbraced hoisted function declaration.
		`import { createRequire } from "node:module"; if (false) { var require = createRequire(import.meta.url); } if (true) function require() { return {}; } require("@rstest/core");`,
		// Catch binding TDZ default referencing catch's own `require` binding.
		`import { createRequire } from "node:module"; const require = createRequire(import.meta.url); try {} catch ({ a = require("@rstest/core"), require }) {}`,
		// Default parameter / computed method key reassigns outer `let require` even when body declares `let require`.
		`import { createRequire } from "node:module"; let require = createRequire(import.meta.url); function hijack(x = (require = () => ({}))) { let require = 1; } hijack(); require("@rstest/core");`,
		`import { createRequire } from "node:module"; let require = createRequire(import.meta.url); const o = { [(require = () => ({}))]() { let require = 1; } }; require("@rstest/core");`,
		// Ambient `declare const require` inside an inner scope must not hide an outer runtime shadow.
		`const require = () => ({}); namespace N { declare const require: any; export const r = require("@rstest/core"); }`,
		// Mutable `let` bindings are rejected conservatively rather than scanning for reassignments.
		`import { createRequire } from "node:module"; let require = createRequire(import.meta.url); require("@rstest/core");`,
	} {
		t.Run("shadowed/"+code, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, code, core.ScriptKindTS)
			reqCall := findRequireCallForModule(sourceFile, "@rstest/core")
			if reqCall == nil {
				t.Fatal("expected a require call in test code")
			}
			if IsModuleRequireCallModules(reqCall, modules) {
				t.Fatalf("expected shadowed require call to be ignored: %s", code)
			}
		})
	}

	for _, code := range []string{
		`require("@rstest/core");`,
		`declare const require: any; require("@rstest/core");`,
		`declare function require(id: string): any; require("@rstest/core");`,
		`import { createRequire } from "node:module"; const require = createRequire(import.meta.url); require("@rstest/core");`,
		`import { createRequire } from "node:module"; const cr = createRequire; const require = cr(import.meta.url); require("@rstest/core");`,
		`import { createRequire as makeReq } from "module"; const require = makeReq(import.meta.url); require("@rstest/core");`,
		`import * as mod from "node:module"; const require = mod.createRequire(import.meta.url); require("@rstest/core");`,
		`import mod from "node:module"; const require = mod.createRequire(import.meta.url); require("@rstest/core");`,
		`import { default as mod } from "node:module"; const require = mod.createRequire(import.meta.url); require("@rstest/core");`,
		`import mod from "node:module"; const { createRequire } = mod; const require = createRequire(import.meta.url); require("@rstest/core");`,
		`const { createRequire } = require("node:module"); function run() { const require = createRequire(import.meta.url); require("@rstest/core"); }`,
		`const mod = require("node:module"); { const require = mod.createRequire(import.meta.url); require("@rstest/core"); }`,
		`const make = require("node:module").createRequire; { const require = make(import.meta.url); require("@rstest/core"); }`,
		`import type { require } from "./types"; require("@rstest/core");`,
		`import { type require } from "./types"; require("@rstest/core");`,
		`import { createRequire } from "node:module"; const require = createRequire(import.meta.url); function run() { { using require = makeDisposable(); } require("@rstest/core"); }`,
		`class C { @dec(require("@rstest/core")) m(require: any) {} }`,
		`import { createRequire } from "node:module"; const require = createRequire(import.meta.url); class C { m(@dec(() => require("@rstest/core")) require: any) {} }`,
		`function f() { { function require() {} } return require("@rstest/core"); }`,
		`function f(x = require("@rstest/core")) { var require = () => ({}); return x; }`,
	} {
		t.Run("genuine/"+code, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, code, core.ScriptKindTS)
			reqCall := findRequireCallForModule(sourceFile, "@rstest/core")
			if reqCall == nil {
				t.Fatal("expected a require call in test code")
			}
			if !IsModuleRequireCallModules(reqCall, modules) {
				t.Fatalf("expected genuine require call to match: %s", code)
			}
		})
	}
}

func TestResolveModuleRequireBinding(t *testing.T) {
	modules := []string{"@rstest/core"}
	for _, tc := range []struct {
		code     string
		wantName string
		wantOK   bool
	}{
		{`const { expect } = require("@rstest/core");`, "expect", true},
		{`const { expect: check } = require("@rstest/core");`, "expect", true},
		{`const { ["expect"]: check } = require("@rstest/core");`, "expect", true},
		{`const { "other": expect } = require("@rstest/core");`, "other", true},
		{`const { ...expect } = require("@rstest/core");`, "", false},
		{`const { nested: { expect } } = require("@rstest/core");`, "", false},
		{`const [expect] = require("@rstest/core");`, "", false},
	} {
		t.Run(tc.code, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, tc.code, core.ScriptKindTS)
			var lastBinding *ast.Node
			var visit func(*ast.Node)
			visit = func(node *ast.Node) {
				if node == nil {
					return
				}
				if node.Kind == ast.KindBindingElement {
					lastBinding = node
				}
				node.ForEachChild(func(child *ast.Node) bool {
					visit(child)
					return false
				})
			}
			visit(sourceFile.AsNode())
			if lastBinding == nil {
				t.Fatal("expected a binding element")
			}
			gotName, _, gotOK := resolveModuleRequireBinding(lastBinding, modules)
			if gotOK != tc.wantOK || gotName != tc.wantName {
				t.Fatalf("resolveModuleRequireBinding = (%q, %v), want (%q, %v)", gotName, gotOK, tc.wantName, tc.wantOK)
			}
		})
	}
}
