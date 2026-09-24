package no_dynamic_require_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_dynamic_require"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expected diagnostics were checked against eslint-plugin-import v2.32.0.
func TestNoDynamicRequireExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &no_dynamic_require.NoDynamicRequireRule,
		[]rule_tester.ValidTestCase{
			// ESTree literals are accepted even when they are not strings.
			{
				Code: "require(0); require(1n); require(true); require(false); require(null); require(/name/); require(''); require(``);",
			},
			{
				Code:    `import(0); import(1n); import(true); import(false); import(null); import(/name/);`,
				Options: []any{map[string]any{"esmodule": true}},
			},
			// Parentheses, optional calls, and comments do not make literals dynamic.
			{
				Code: "((require))((('name'))); require?.((`name`)); require(/* path */ ('name'));",
			},
			{
				Code:    "import((('name'))); import((`name`));",
				Options: []any{map[string]any{"esmodule": true}},
			},
			// Only direct calls named require are checked.
			{
				Code: "module.require(name); require.resolve(name); require[name](name); require['resolve'](name); obj?.require(name); obj.require?.(name); (require?.resolve)(name); new require(name); require`${name}`; (0, require)(name);",
			},
			{
				Code: `class Loader { #require; load(name) { return this.#require(name); } }`,
			},
			// Authored TypeScript callee wrappers remain visible to ESTree.
			{
				Code: `(require as any)(name); (require satisfies any)(name); require!(name); (<any>require)(name); const typed = require<string>;`,
			},
			{
				Code: `import module = require('name'); type Module = import('name');`,
			},
			// Both forms of the default option leave dynamic imports alone.
			{
				Code:    `import(name);`,
				Options: []any{map[string]any{}},
			},
			{
				Code:    `import(name);`,
				Options: []any{map[string]any{"esmodule": false}},
			},
			// Only the first argument matters, including import attributes.
			{
				Code:    `require('name', dynamic); import('name', { with: attributes });`,
				Options: []any{map[string]any{"esmodule": true}},
			},
			// JavaScript JSDoc casts are transparent, like source parentheses.
			{
				Code:     "/** @type {any} */ (require)(/** @type {string} */ ('name')); require(/** @satisfies {string} */ (`name`));",
				FileName: "case.js",
			},
			// JSX tags are not calls; runtime literal arguments remain allowed.
			{
				Code: `const node = <require.resolve>{require('name')}</require.resolve>;`,
				Tsx:  true,
			},
			// Alias and global-object calls are outside the direct-call contract.
			{
				Code: `const load = require; load(name); globalThis.require(name);`,
			},
			// A TypeScript instantiation expression is not an identifier callee.
			{
				Code: `(require<string>)(name);`,
			},
			// The pinned rule checks plain import(), not import meta-property calls.
			{
				Code:    `import.meta.resolve(name); import.defer(name);`,
				Options: []any{map[string]any{"esmodule": true}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Constant expressions are not evaluated.
			{
				Code: "const name = 'module'; require(name); require('a' + 'b'); require(`${'name'}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 37},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 39, EndLine: 1, EndColumn: 57},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 59, EndLine: 1, EndColumn: 79},
				},
			},
			// Unary expressions and undefined are not ESTree literals.
			{
				Code: `require(-1); require(+1); require(undefined); require(void 0);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 14, EndLine: 1, EndColumn: 25},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 27, EndLine: 1, EndColumn: 45},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 47, EndLine: 1, EndColumn: 62},
				},
			},
			// Spread, object, array, and function arguments are dynamic.
			{
				Code: `require(...names); require({name}); require([name]); require(() => name);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 35},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 37, EndLine: 1, EndColumn: 52},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 54, EndLine: 1, EndColumn: 73},
				},
			},
			// Parenthesized callees and arguments match ESTree.
			{
				Code: `((require))((name));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// Optional direct calls are checked.
			{
				Code: `require?.(name);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// An outer call of an optional call is not another require call.
			{
				Code: `(require?.(name))(other);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 2, EndLine: 1, EndColumn: 17},
				},
			},
			// Shadowed require bindings are still checked upstream.
			{
				Code: `function load(require) { return require(name); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 33, EndLine: 1, EndColumn: 46},
				},
			},
			// Nested calls are visited even when the outer argument is static.
			{
				Code: `require('name', require(name)); require(require(name));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 30},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 33, EndLine: 1, EndColumn: 55},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 41, EndLine: 1, EndColumn: 54},
				},
			},
			// Empty and explicit default options continue to check require.
			{
				Code:    `require(name);`,
				Options: []any{map[string]any{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    `require(name); import(name);`,
				Options: []any{map[string]any{"esmodule": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			// Type arguments keep a direct call; assertion arguments are not literals.
			{
				Code: `require<string>(name); require('name' as string); require('name' satisfies string); require('name'!); require(<string>'name');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 49},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 51, EndLine: 1, EndColumn: 83},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 85, EndLine: 1, EndColumn: 101},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 103, EndLine: 1, EndColumn: 126},
				},
			},
			// Dynamic imports inspect the source and include attributes in their range.
			{
				Code:    `import((name), { with: { type: 'json' } });`,
				Options: []any{map[string]any{"esmodule": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
				},
			},
			{
				Code:    "import('name' as string); import(`${'name'}`);",
				Options: []any{map[string]any{"esmodule": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
					{MessageId: "", Message: importMessage, Line: 1, Column: 27, EndLine: 1, EndColumn: 46},
				},
			},
			// JSDoc casts around dynamic calls and arguments stay transparent.
			{
				Code:     `/** @type {any} */ (require)(/** @type {string} */ (name));`,
				FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 59},
				},
			},
			// Calls inside JSX expressions are checked.
			{
				Code:    `const node = <Loader path={require(name)}>{import(name)}</Loader>;`,
				Tsx:     true,
				Options: []any{map[string]any{"esmodule": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 28, EndLine: 1, EndColumn: 41},
					{MessageId: "", Message: importMessage, Line: 1, Column: 44, EndLine: 1, EndColumn: 56},
				},
			},
			// Report complete multiline call spans and UTF-16 columns.
			{
				Code: "'😀'; require(\n  `./${name}`\n);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 7, EndLine: 3, EndColumn: 2},
				},
			},
			{
				Code:    "'😀'; import(\n  name,\n  { with: { type: 'json' } }\n);",
				Options: []any{map[string]any{"esmodule": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 7, EndLine: 4, EndColumn: 2},
				},
			},
			// Real loader patterns still need a literal module path.
			{
				Code: `const load = name => require(path.join(__dirname, name));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 57},
				},
			},
			// Escaped identifiers are decoded before matching require.
			{
				Code: `r\u0065quire(name); r\u{65}quire(name);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
					{MessageId: "", Message: requireMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 39},
				},
			},
			// Nested JavaScript JSDoc casts preserve the complete call range.
			{
				Code:     `((/** @type {*} */ (require)))(/* path */ (name));`,
				FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 50},
				},
			},
			// Tagged templates are expressions even without substitutions.
			{
				Code:    "require(String.raw`name`); import(String.raw`name`);",
				Options: []any{map[string]any{"esmodule": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
					{MessageId: "", Message: importMessage, Line: 1, Column: 28, EndLine: 1, EndColumn: 52},
				},
			},
			// A BOM, hashbang, and CRLF line breaks do not shift the diagnostic.
			{
				Code:     "\ufeff#!/usr/bin/env node\r\nrequire(name);\r\n",
				FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 14},
				},
			},
		},
	)
}
