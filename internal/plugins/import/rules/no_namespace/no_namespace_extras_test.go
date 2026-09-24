package no_namespace_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_namespace"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNamespaceIgnore(t *testing.T) {
	options := func(patterns ...any) map[string]any { return map[string]any{"ignore": append([]any{}, patterns...)} }
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_namespace.NoNamespaceRule,
		[]rule_tester.ValidTestCase{
			{Code: `export * as ns from 'm'; import name = require('m'); import('m');`},
			{Code: `import * as ns from '@pkg/nested/file.svg';`, Options: options("*.svg")},
			{Code: `import * as ns from './a/file.svg';`, Options: options("./a/**")},
			{Code: `import * as ns from './file.json';`, Options: options("*.{svg,json}")},
			{Code: `import * as ns from './file.svg';`, Options: options("*.@(svg|json)")},
			{Code: `import * as ns from './file.json';`, Options: options("!*.js")},
			{Code: `import * as ns from './.hidden';`, Options: options(".*")},
			{Code: `import * as ns from './f\u0069le.svg';`, Options: options("file.svg")},
			{Code: `import * as ns from './file.json';`, Options: options("*.js", "*.json")},
			{Code: `import * as ns from '';`, Options: options("**", "")},
			{Code: `import * as ns from '';`, Options: options(" ", "**")},
		}, []rule_tester.InvalidTestCase{
			{Code: `import * as ns from 'm';`, Options: map[string]any{}, Errors: namespaceError(1, 8, 1, 15)},
			{Code: `import * as ns from 'm';`, Options: options(), Errors: namespaceError(1, 8, 1, 15)},
			{Code: `import * as ns from './file.SVG';`, Options: options("*.svg"), Errors: namespaceError(1, 8, 1, 15)},
			{Code: `import * as ns from './.hidden';`, Options: options("*"), Errors: namespaceError(1, 8, 1, 15)},
			{Code: `import * as ns from './b/file.svg';`, Options: options("./a/*"), Errors: namespaceError(1, 8, 1, 15)},
			{Code: `import * as ns from './file.js';`, Options: options("!*.js"), Errors: namespaceError(1, 8, 1, 15)},
			// Array.find returns an empty string, which is falsy upstream.
			{Code: `import * as ns from '';`, Options: options(""), Errors: namespaceError(1, 8, 1, 15)},
			{Code: `import * as ns from '';`, Options: options("", "**"), Errors: namespaceError(1, 8, 1, 15)},
			{Code: `import * as ns from 'm';`, Options: options("#m"), Errors: namespaceError(1, 8, 1, 15)},
			// The upstream schema intentionally allows unrecognized properties.
			{Code: `import * as ns from 'm';`, Options: map[string]any{"other": true}, Errors: namespaceError(1, 8, 1, 15)},
		})
}

func TestNoNamespaceFixes(t *testing.T) {
	const prefix = `import * as ns from 'm'; `
	tests := []struct{ code, output string }{
		{`ns.foo; ns.foo; ns.bar.baz;`, `import { foo, bar } from 'm'; foo; foo; bar.baz;`},
		{`(ns).foo; ns[('bar')]; ns?.baz; ns?.['qux'];`, `import { foo, bar, baz, qux } from 'm'; foo; bar; baz; qux;`},
		{`class C extends ns.Base {}`, `import { Base } from 'm'; class C extends Base {}`},
		{`interface T extends ns.Foo {}`, `import { Foo } from 'm'; interface T extends Foo {}`},
		{`class C implements ns.Foo {}`, `import { Foo } from 'm'; class C implements Foo {}`},
		{`class C<Foo> implements ns.Foo {}`, `import { Foo as ns_Foo } from 'm'; class C<Foo> implements ns_Foo {}`},
		{`function f<foo>() { return ns.foo; }`, `import { foo as ns_foo } from 'm'; function f<foo>() { return ns_foo; }`},
		{`ns.foo satisfies T; ns.bar!;`, `import { foo, bar } from 'm'; foo satisfies T; bar!;`},
		{`const { value } = ns.foo; ns.bar<string>();`, `import { foo, bar } from 'm'; const { value } = foo; bar<string>();`},
		{`function f(ns) { ns.foo; } ns.bar;`, `import { bar } from 'm'; function f(ns) { ns.foo; } bar;`},
		{`const f = function foo(arg) { return ns.foo + ns.arg; };`, `import { foo as ns_foo, arg as ns_arg } from 'm'; const f = function foo(arg) { return ns_foo + ns_arg; };`},
		{`function f(foo) { return ns.foo; } function g(ns_foo, ns_foo_1) { return ns.foo; }`, `import { foo as ns_foo_2 } from 'm'; function f(foo) { return ns_foo_2; } function g(ns_foo, ns_foo_1) { return ns_foo_2; }`},
		{`try {} catch (foo) { ns.foo; }`, `import { foo as ns_foo } from 'm'; try {} catch (foo) { ns_foo; }`},
		{`ns.type; ns.from;`, `import { type, from } from 'm'; type; from;`},
		{`ns.变量; ns['\u0066oo'];`, `import { 变量, foo } from 'm'; 变量; foo;`},
		{`ns['\u{10400}'];`, `import { 𐐀 } from 'm'; 𐐀;`},
		{`delete ns.foo.bar; ns.foo.bar++;`, `import { foo } from 'm'; delete foo.bar; foo.bar++;`},
		{`class C { static { const foo = 1; ns.foo; } value = ns.bar; }`, `import { foo as ns_foo, bar } from 'm'; class C { static { const foo = 1; ns_foo; } value = bar; }`},
		// Avoid upstream's outer-scope collision and duplicate generated locals.
		{`const foo = 1; function f() { { ns.foo; } }`, `import { foo as ns_foo } from 'm'; const foo = 1; function f() { { ns_foo; } }`},
		{`ns.foo; ns.ns_foo; const foo = 1;`, `import { foo as ns_foo, ns_foo as ns_ns_foo } from 'm'; ns_foo; ns_ns_foo; const foo = 1;`},
		// Upstream can join return with the following identifier.
		{`function f() { return(ns).foo; }`, `import { foo } from 'm'; function f() { return foo; }`},
		{`function f() { throw(ns).foo; }`, `import { foo } from 'm'; function f() { throw foo; }`},
		// Plain object property names crash the upstream conflict accumulator.
		{`ns.constructor; ns.__proto__;`, `import { constructor as ns_constructor, __proto__ } from 'm'; ns_constructor; __proto__;`},
	}
	invalid := make([]rule_tester.InvalidTestCase, 0, len(tests)+4)
	for _, tc := range tests {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: prefix + tc.code, Output: []string{tc.output}, Errors: namespaceError(1, 8, 1, 15),
		})
	}
	invalid = append(invalid,
		rule_tester.InvalidTestCase{
			Code:   `import value, * as ns from 'm' with { type: 'json' }; ns.foo;`,
			Output: []string{`import value, { foo } from 'm' with { type: 'json' }; foo;`},
			Errors: namespaceError(1, 15, 1, 22),
		},
		rule_tester.InvalidTestCase{
			Code:   `import type * as ns from 'm'; interface C extends ns.Base {}`,
			Output: []string{`import type { Base } from 'm'; interface C extends Base {}`},
			Errors: namespaceError(1, 13, 1, 20),
		},
		rule_tester.InvalidTestCase{
			Code: prefix + `const view = <div>{ns.foo}</div>;`, Tsx: true,
			Output: []string{`import { foo } from 'm'; const view = <div>{foo}</div>;`},
			Errors: namespaceError(1, 8, 1, 15),
		},
		rule_tester.InvalidTestCase{
			Code: prefix + `ns.console; ns.Map;`, Globals: map[string]any{"console": "readonly"},
			Output: []string{`import { console as ns_console, Map as ns_Map } from 'm'; ns_console; ns_Map;`},
			Errors: namespaceError(1, 8, 1, 15),
		},
		rule_tester.InvalidTestCase{
			Code:   "import\n  * as\n  命名 from 'm'; 命名.foo;",
			Output: []string{"import\n  { foo } from 'm'; foo;"},
			Errors: namespaceError(2, 3, 3, 5),
		},
		rule_tester.InvalidTestCase{
			Code:   `/*😀*/ import * as ns from 'm'; ns.foo;`,
			Output: []string{`/*😀*/ import { foo } from 'm'; foo;`},
			Errors: namespaceError(1, 15, 1, 22),
		},
		rule_tester.InvalidTestCase{
			Code:   "import * as ns from 'm'; ns.foo;\nimport * as other from 'n'; other.foo;",
			Output: []string{"import { foo } from 'm'; foo;\nimport { foo as other_foo } from 'n'; other_foo;"},
			Errors: append(namespaceError(1, 8, 1, 15), namespaceError(2, 8, 2, 18)...),
		},
	)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_namespace.NoNamespaceRule, nil, invalid)
}

func TestNoNamespaceWithoutFixes(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, use := range []string{
		`consume(ns); ns.foo;`, `ns();`, `const { foo } = ns;`, `const copy = { ...ns };`,
		`export { ns };`, `function f(ns) { return ns.foo; }`,
		`type T = ns.Foo;`, `type T = typeof ns.foo;`,
		`(ns as any).foo;`, `ns!.foo;`, `(ns satisfies T).foo;`,
		`const element = <ns.Component />;`, `class C { #foo; f() { return ns.#foo; } }`,
		// Upstream misinterprets dynamic keys and namespace references used as keys.
		`ns[key];`, `obj[ns];`, `ns[ns.foo];`, `ns[getKey()];`, "ns[`foo`];",
		`ns[0];`, `ns[1n];`, `ns[true];`, `ns[null];`, `ns[/foo/];`, `ns['not-an-identifier'];`,
		`ns.default;`, `ns.await;`, `ns.yield;`, `ns.arguments;`, `ns.eval;`,
		// Do not introduce writes to imported bindings or invalid delete targets.
		`ns.foo = 1;`, `ns.foo++;`, `ns.foo ||= 1;`, `({ value: ns.foo } = obj);`,
		`[ns.foo] = values;`, `for (ns.foo of values) {}`, `(ns.foo as any) = 1;`,
		`delete ns.foo;`, `delete ((ns.foo));`, `delete (ns.foo as any);`,
		// Preserve comments which upstream's member replacement removes.
		`ns./*keep*/foo;`, `ns[/*keep*/'foo'];`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: `import * as ns from 'm'; ` + use, Tsx: true, Errors: namespaceError(1, 8, 1, 15),
		})
	}
	invalid = append(invalid,
		rule_tester.InvalidTestCase{Code: `import * /*keep*/ as ns from 'm'; ns.foo;`, Errors: namespaceError(1, 8, 1, 24)},
		rule_tester.InvalidTestCase{Code: `import type * as ns from 'm'; type T = ns.Foo;`, Errors: namespaceError(1, 13, 1, 20)},
		rule_tester.InvalidTestCase{Code: `import * as ns from 'm'; (/** @type {any} */ (ns)).foo;`, FileName: "cast.js", TSConfig: "tsconfig.allow-js.json", Errors: namespaceError(1, 8, 1, 15)},
	)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_namespace.NoNamespaceRule, nil, invalid)
}

func TestNoNamespaceSchema(t *testing.T) {
	for _, options := range [][]any{
		{map[string]any{"ignore": "*.js"}},
		{map[string]any{"ignore": []any{1}}},
		{map[string]any{"ignore": []any{"m", "m"}}},
		{true},
		{map[string]any{}, map[string]any{}},
	} {
		if err := no_namespace.NoNamespaceRule.Schema.Validate(options); err == nil {
			t.Errorf("accepted invalid options: %#v", options)
		}
	}
}

func TestNoNamespaceEditDemand(t *testing.T) {
	for _, use := range []string{`ns.foo;`, `consume(ns);`} {
		t.Run(use, func(t *testing.T) {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", Path: "/test.ts"}, `import * as ns from 'm'; `+use, core.ScriptKindTS)
			binder.BindSourceFile(file)
			node := file.Statements.Nodes[0].AsImportDeclaration().ImportClause.AsImportClause().NamedBindings
			var all rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
				var diagnostics []rule.RuleDiagnostic
				ctx := (rule.RuleContext{
					SourceFile: file,
					Refs:       rule.NewRefStore(file, &core.CompilerOptions{}, nil, rule.RefStoreInit{}),
				}).WithDiagnosticConsumer(no_namespace.NoNamespaceRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
					Demand: demand,
					Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
				})
				no_namespace.NoNamespaceRule.Run(ctx, nil)[ast.KindNamespaceImport](node)
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(diagnostics))
				}
				got := diagnostics[0]
				if demand == rule.EditDemandAll {
					all = got
				}
				wantFix := use == `ns.foo;` && (demand == rule.EditDemandAll || demand == rule.EditDemandAutofix)
				if (len(got.Fixes()) > 0) != wantFix || got.Suggestions != nil {
					t.Fatalf("demand %d: incorrect edit availability", demand)
				}
				if wantFix && !reflect.DeepEqual(got.FixesPtr, all.FixesPtr) {
					t.Fatalf("demand %d: edits differ from all edits", demand)
				}
				want := all
				got.FixesPtr, want.FixesPtr = nil, nil
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("demand %d: diagnostic identity changed", demand)
				}
			}
		})
	}
}
