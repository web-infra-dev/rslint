package no_default_export_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_default_export"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	preferNamedMessage           = "Prefer named exports."
	noAliasDefaultMessage        = "Do not alias `foo` as `default`. Just export `foo` itself instead."
	noAliasUndefinedMessage      = "Do not alias `undefined` as `default`. Just export `undefined` itself instead."
	noAliasDefaultDefaultMessage = "Do not alias `default` as `default`. Just export `default` itself instead."
)

// TestNoDefaultExportUpstream migrates the full valid/invalid suite from
// upstream tests/src/rules/no-default-export.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// no_default_export_extras_test.go.
func TestNoDefaultExportUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_default_export.NoDefaultExportRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid: CommonJS is not an ES default export ----
			{Code: `module.exports = function foo() {}`},
			{Code: `module.exports = function foo() {}`},

			// ---- upstream valid: named exports ----
			{Code: "export const foo = 'foo';\nexport const bar = 'bar';"},
			{Code: "export const foo = 'foo';\nexport function bar() {};"},
			{Code: `export const foo = 'foo';`},
			{Code: "const foo = 'foo';\nexport { foo };"},
			{Code: `let foo, bar; export { foo, bar }`},
			{Code: `export const { foo, bar } = item;`},
			{Code: `export const { foo, bar: baz } = item;`},
			{Code: `export const { foo: { bar, baz } } = item;`},
			{Code: "let item;\nexport const foo = item;\nexport { item };"},
			{Code: `export * from './foo';`},
			{Code: `export const { foo } = { foo: "bar" };`},
			{Code: `export const { foo: { bar } } = { foo: { bar: "baz" } };`},
			{Code: `export { a, b } from "foo.js"`},

			// ---- upstream valid: no exports at all ----
			{Code: `import * as foo from './foo';`},
			{Code: `import foo from './foo';`},
			{Code: `import {default as foo} from './foo';`},

			// ---- upstream valid: type exports and unsupported Babel proposal syntax ----
			{Code: `export type UserId = number;`},
			// SKIP: rslint does not parse Babel's default re-export proposal `export foo from`.
			{Code: `export foo from "foo.js"`, Skip: true},
			// SKIP: rslint does not parse Babel's default re-export proposal `export foo, { bar } from`.
			{Code: `export Memory, { MemoryValue } from './Memory'`, Skip: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid: direct default exports ----
			{
				Code: `export default function bar() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: "export const foo = 'foo';\nexport default bar;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 2, Column: 8, EndLine: 2, EndColumn: 15},
				},
			},
			{
				Code: `export default class Bar {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `export default function() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `export default class {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},

			// ---- upstream invalid: aliasing a local export as default ----
			{
				Code: `let foo; export { foo as default }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAliasDefault", Message: noAliasDefaultMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			// SKIP: rslint does not parse Babel's default re-export proposal `export default from`.
			{
				Code: `export default from "foo.js"`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed"},
				},
			},
			{
				Code: `let foo; export { foo as "default" }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAliasDefault", Message: noAliasDefaultMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
		},
	)
}
