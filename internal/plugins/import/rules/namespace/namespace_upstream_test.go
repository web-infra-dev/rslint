package namespace_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/namespace"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	notFoundNamesC = `'c' not found in imported namespace 'names'.`
	computedNames  = `Unable to validate computed reference to imported namespace 'names'.`
)

// TestNamespaceUpstream migrates the valid/invalid suite from upstream
// tests/src/rules/namespace.js where tsgo supports the syntax and framework
// hooks. Some fixture paths are renamed by behavior to keep this repository
// readable. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in namespace_extras_test.go.
func TestNamespaceUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&namespace.NamespaceRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid: no import specifiers / no export map ----
			{Code: `import "./malformed.js"`},
			{Code: `import * as foo from './empty-folder';`},
			{Code: `import * as foo from './common';`},

			// ---- upstream valid: direct namespace members ----
			{Code: `import * as names from "./named-exports"; console.log((names.b).c); `},
			{Code: `import * as names from "./named-exports"; console.log(names.a);`},
			{Code: `import * as names from "./re-export-names"; console.log(names.foo);`},
			{Code: `import * as elements from "./jsx";`},
			{
				Code: `
					import * as foo from "./jsx/re-export";
					console.log(foo.jsxFoo);
				`,
			},
			{
				Code: `
					import * as components from "./jsx/component-exports";
					console.log(components.Baz1);
					console.log(components.Baz2);
					console.log(components.Qux1);
					console.log(components.Qux2);
				`,
			},

			// ---- upstream valid: destructuring namespaces ----
			{Code: `import * as names from "./named-exports"; const { a } = names`},
			{Code: `import * as names from "./named-exports"; const { d: c } = names`},
			{Code: `import * as names from "./named-exports"; const { c } = foo, { length } = "names", alt = names;`},
			{Code: `import * as names from "./named-exports"; const { ExportedClass: { length } } = names`},

			// ---- upstream valid: detect scope redefinition ----
			{Code: `import * as names from "./named-exports"; function b(names) { const { c } = names }`},
			{Code: `import * as names from "./named-exports"; function b() { let names = null; const { c } = names }`},
			{Code: `import * as names from "./named-exports"; const x = function names() { const { c } = names }`},

			// ---- upstream valid: export namespace specifier ----
			{Code: `export * as names from "./named-exports"`},
			// SKIP: rslint does not parse Babel's default plus namespace re-export proposal.
			{Code: `export defaultExport, * as names from "./named-exports"`, Skip: true},
			{Code: `export * as names from "./does-not-exist"`},

			// ---- upstream valid: endpoint namespace fixture and hoisting ----
			{Code: `import * as Endpoints from "./endpoint-namespace-module/endpoints"; console.log(Endpoints.Users)`},
			{Code: `function x() { console.log((names.b).c); } import * as names from "./named-exports"; `},
			{Code: `import * as names from './default-export';`},
			{Code: `import * as names from './default-export'; console.log(names.default)`},
			{Code: `export * as names from "./default-export"`},
			// SKIP: rslint does not parse Babel's default plus namespace re-export proposal.
			{Code: `export defaultExport, * as names from "./default-export"`, Skip: true},

			// ---- upstream valid: options and object-rest properties ----
			{Code: `import * as names from './named-exports'; console.log(names['a']);`, Options: []interface{}{map[string]interface{}{"allowComputed": true}}},
			{Code: `import * as names from './named-exports'; console.log(names['a']);`, Options: map[string]interface{}{"allowComputed": true}},
			{Code: `import * as names from './named-exports'; const {a, b, ...rest} = names;`},
			{Code: `import * as ns from './re-export-common'; const {foo} = ns;`},

			// ---- upstream valid: JSX ----
			{Code: `import * as Names from "./named-exports"; const Foo = <Names.a/>`, Tsx: true},

			// ---- upstream valid: TypeScript namespace export ----
			{Code: `import * as foo from "./typescript-declare-nested"; foo.bar.MyFunction()`},
			{Code: `import { foobar } from "./typescript-declare-interface"`},
			{Code: `export * from "typescript/lib/typescript.d"`},
			{Code: `export = function name() {}`},

			// ---- upstream valid: syntax smoke cases with no namespace diagnostics ----
			{Code: `for (let { foo, bar } of baz) {}`},
			{Code: `for (let [ foo, bar ] of baz) {}`},
			{Code: `const { x, y } = bar`},
			{Code: `const { x, y, ...z } = bar`},
			{Code: `let x; export { x }`},
			{Code: `let x; export { x as y }`},
			{Code: `export const x = null`},
			{Code: `export var x = null`},
			{Code: `export let x = null`},
			{Code: `export default x`},
			{Code: `export default class x {}`},
			{Code: `import json from "./data.json"`},
			{Code: `import foo from "./foobar.json";`},
			{Code: `import foo from "./foobar";`},
			{Code: `import { foo } from "./issue-370-commonjs-namespace/bar"`},
			{Code: `export * from "./issue-370-commonjs-namespace/bar"`},
			{Code: `import * as a from "./commonjs-namespace/a"; a.b`},
			{Code: `import { foo } from "./ignore.invalid.extension"`},

			// ---- upstream valid: shadowing and deep namespace re-exports ----
			{Code: `import * as color from './color'; export const getBackgroundFromColor = (color) => color.bg; export const getExampleColor = () => color.example`},
			{Code: `import * as aliasChain from './namespace-alias'; console.log(aliasChain.myName);`, FileName: "export-namespace-alias-chain/consumer.ts"},
			{Code: `import * as reexports from './namespace-reexports'; console.log(reexports.abc.A); console.log(reexports.def.D);`, FileName: "multiple-namespace-reexports/consumer.ts"},
			{Code: `import * as names from './default-export-string';`},
			{Code: `import * as names from './default-export-string'; console.log(names.default)`},
			{Code: `import * as names from './default-export-namespace-string';`},
			{Code: `import * as names from './default-export-namespace-string'; console.log(names.default)`},
			{Code: `import { "b" as b } from "./deep-namespace-chain/entry"; console.log(b.c.d.e)`},
			{Code: `import { "b" as b } from "./deep-namespace-chain/entry"; var {c:{d:{e}}} = b`},
			{Code: `import * as a from "./deep-namespace-chain/entry"; console.log(a.b.c.d.e)`},
			{Code: `import { b } from "./deep-namespace-chain/entry"; console.log(b.c.d.e)`},
			{Code: `import * as a from "./deep-namespace-chain/entry"; console.log(a.b.c.d.e.f)`},
			{Code: `import * as a from "./deep-namespace-chain/entry"; var {b:{c:{d:{e}}}} = a`},
			{Code: `import { b } from "./deep-namespace-chain/entry"; var {c:{d:{e}}} = b`},
			{Code: `import * as a from "./deep-namespace-chain/entry"; console.log(a.b.default)`},
			// eslint-plugin-import exposes `export * as` aliases as top-level
			// names, but does not attach namespace metadata for deeper checks.
			{Code: `import b from './deep-namespace-chain/default'; console.log(b.e)`},
			{Code: `import { "b" as b } from "./deep-namespace-chain/entry"; console.log(b.e)`},
			{Code: `import { "b" as b } from "./deep-namespace-chain/entry"; console.log(b.c.e)`},
			{Code: `import * as aliasChain from './namespace-alias'; console.log(aliasChain.myName.b);`, FileName: "export-namespace-alias-chain/consumer.ts"},
			{Code: `import { myName } from './namespace-alias'; console.log(myName.b);`, FileName: "export-namespace-alias-chain/consumer.ts"},
			{Code: `import * as a from "./deep-namespace-chain/entry"; console.log(a.b.e)`},
			{Code: `import { b } from "./deep-namespace-chain/entry"; console.log(b.e)`},
			{Code: `import * as a from "./deep-namespace-chain/entry"; console.log(a.b.c.e)`},
			{Code: `import { b } from "./deep-namespace-chain/entry"; console.log(b.c.e)`},
			{Code: `import * as a from "./deep-namespace-chain/entry"; var {b:{ e }} = a`},
			{Code: `import * as a from "./deep-namespace-chain/entry"; var {b:{c:{ e }}} = a`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `import * as names from './named-exports'; console.log(names.c)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 61, EndLine: 1, EndColumn: 62},
				},
			},
			{
				Code: `import * as names from './named-exports'; console.log(names['a']);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "computed", Message: computedNames, Line: 1, Column: 61, EndLine: 1, EndColumn: 64},
				},
			},
			{
				Code: `import * as foo from './bar'; foo.foo = 'y';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignment", Message: `Assignment to member of namespace 'foo'.`, Line: 1, Column: 31, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code: `import * as foo from './bar'; foo.x = 'y';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignment", Message: `Assignment to member of namespace 'foo'.`, Line: 1, Column: 31, EndLine: 1, EndColumn: 42},
					{MessageId: "notFound", Message: `'x' not found in imported namespace 'foo'.`, Line: 1, Column: 35, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code: `import * as names from "./named-exports"; const { c } = names`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 51, EndLine: 1, EndColumn: 52},
				},
			},
			{
				Code: `import * as names from "./named-exports"; function b() { const { c } = names }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 66, EndLine: 1, EndColumn: 67},
				},
			},
			{
				Code: `import * as names from "./named-exports"; const { c: d } = names`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 51, EndLine: 1, EndColumn: 52},
				},
			},
			{
				Code: `import * as names from "./named-exports"; const { c: { d } } = names`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 51, EndLine: 1, EndColumn: 52},
				},
			},
			{
				Code: `import * as Endpoints from "./endpoint-namespace-module/endpoints"; console.log(Endpoints.Foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: `'Foo' not found in imported namespace 'Endpoints'.`, Line: 1, Column: 91, EndLine: 1, EndColumn: 94},
				},
			},
			// SKIP: rslint's test harness creates programs before rules run, so imported parse errors are framework-level failures.
			{
				Code: `import * as namespace from './malformed.js';`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "parseError"},
				},
			},
			{
				Code: `console.log(names.c); import * as names from './named-exports';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: `function x() { console.log(names.c) } import * as names from './named-exports';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 34, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code: `import * as ree from "./re-export"; console.log(ree.default)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: `'default' not found in imported namespace 'ree'.`, Line: 1, Column: 53, EndLine: 1, EndColumn: 60},
				},
			},
			{
				Code: `import * as Names from "./named-exports"; const Foo = <Names.e/>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: `'e' not found in imported namespace 'Names'.`, Line: 1, Column: 62, EndLine: 1, EndColumn: 63},
				},
			},
		},
	)
}
