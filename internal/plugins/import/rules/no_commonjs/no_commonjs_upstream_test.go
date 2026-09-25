package no_commonjs_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_commonjs"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	exportMessage = `Expected "export" or "export default"`
	importMessage = `Expected "import" instead of "require()"`
)

// Every case from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-commonjs.js
// Upstream uses literal messages, so every diagnostic has an empty message ID.
func TestNoCommonjsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_commonjs.NoCommonjsRule,
		[]rule_tester.ValidTestCase{
			// Upstream valid: ES imports.
			{
				Code: `import "x";`,
			},
			{
				Code: `import x from "x"`,
			},
			{
				Code: `import { x } from "x"`,
			},
			// Upstream valid: ES exports and a local exports binding.
			{
				Code: `export default "x"`,
			},
			{
				Code: `export function house() {}`,
			},
			{
				Code: `
        function someFunc() {
          const exports = someComputation();
          expect(exports.someProp).toEqual({ a: 'value' });
        }
      `,
			},
			// Upstream valid: allowed requires and option exceptions.
			{
				Code: `function a() { var x = require("y"); }`,
			},
			{
				Code: `var a = c && require("b")`,
			},
			{
				Code: `require.resolve("help")`,
			},
			{
				Code: `require.ensure([])`,
			},
			{
				Code: `require([], function(a, b, c) {})`,
			},
			{
				Code: `var bar = require('./bar', true);`,
			},
			{
				Code: `var bar = proxyquire('./bar');`,
			},
			{
				Code: `var bar = require('./ba' + 'r');`,
			},
			{
				Code: "var bar = require(`x${1}`);",
			},
			{
				Code: `var zero = require(0);`,
			},
			{
				Code:    `require("x")`,
				Options: []any{map[string]any{"allowRequire": true}},
			},
			{
				Code:    `require(rootRequire("x"))`,
				Options: []any{map[string]any{"allowRequire": true}},
			},
			{
				Code:    `require(String("x"))`,
				Options: []any{map[string]any{"allowRequire": true}},
			},
			{
				Code:    `require(["x", "y", "z"].join("/"))`,
				Options: []any{map[string]any{"allowRequire": true}},
			},
			{
				Code:    `rootRequire("x")`,
				Options: []any{map[string]any{"allowRequire": true}},
			},
			{
				Code:    `rootRequire("x")`,
				Options: []any{map[string]any{"allowRequire": false}},
			},
			{
				Code:    `module.exports = function () {}`,
				Options: []any{"allow-primitive-modules"},
			},
			{
				Code:    `module.exports = function () {}`,
				Options: []any{map[string]any{"allowPrimitiveModules": true}},
			},
			{
				Code:    `module.exports = "foo"`,
				Options: []any{"allow-primitive-modules"},
			},
			{
				Code:    `module.exports = "foo"`,
				Options: []any{map[string]any{"allowPrimitiveModules": true}},
			},
			{
				Code:    `if (typeof window !== "undefined") require("x")`,
				Options: []any{map[string]any{"allowRequire": true}},
			},
			{
				Code:    `if (typeof window !== "undefined") require("x")`,
				Options: []any{map[string]any{"allowRequire": false}},
			},
			{
				Code:    `if (typeof window !== "undefined") { require("x") }`,
				Options: []any{map[string]any{"allowRequire": true}},
			},
			{
				Code:    `if (typeof window !== "undefined") { require("x") }`,
				Options: []any{map[string]any{"allowRequire": false}},
			},
			{
				Code: `try { require("x") } catch (error) {}`,
			},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream invalid: require calls (the ESLint >= 4 branch).
			{
				Code: `var x = require("x")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 9, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `x = require("x")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 5, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code: `require("x")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code: "require(`x`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code:    `if (typeof window !== "undefined") require("x")`,
				Options: []any{map[string]any{"allowConditionalRequire": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 36, EndLine: 1, EndColumn: 43},
				},
			},
			{
				Code:    `if (typeof window !== "undefined") { require("x") }`,
				Options: []any{map[string]any{"allowConditionalRequire": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 38, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code:    `try { require("x") } catch (error) {}`,
				Options: []any{map[string]any{"allowConditionalRequire": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 7, EndLine: 1, EndColumn: 14},
				},
			},
			// Upstream invalid: CommonJS exports.
			{
				Code: `exports.face = "palm"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: `module.exports.face = "palm"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `module.exports = face`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `exports = module.exports = {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: `var x = module.exports = {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 9, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `module.exports = {}`,
				Options: []any{"allow-primitive-modules"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `var x = module.exports`,
				Options: []any{"allow-primitive-modules"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 9, EndLine: 1, EndColumn: 23},
				},
			},
		},
	)
}

// All code examples from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-commonjs.md
// The conditional example also runs with allowConditionalRequire: false.
func TestNoCommonjsDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_commonjs.NoCommonjsRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `/*eslint no-commonjs: [2, { allowRequire: true }]*/
var mod = require('./mod');`,
				Options: []any{map[string]any{"allowRequire": true}},
			},
			{
				Code: `var a = b && require("c")

if (typeof window !== "undefined") {
  require('that-ugly-thing');
}

var fs = null;
try {
  fs = require("fs")
} catch (error) {}`,
			},
			{
				Code: `/*eslint no-commonjs: [2, { allowPrimitiveModules: true }]*/

module.exports = "foo"
module.exports = function rule(context) { return { /* ... */ } }`,
				Options: []any{map[string]any{"allowPrimitiveModules": true}},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `var mod = require('./mod')
  , common = require('./common')
  , fs = require('fs')
  , whateverModule = require('./not-found')

module.exports = { a: "b" }
exports.c = "d"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
					{MessageId: "", Message: importMessage, Line: 2, Column: 14, EndLine: 2, EndColumn: 21},
					{MessageId: "", Message: importMessage, Line: 3, Column: 10, EndLine: 3, EndColumn: 17},
					{MessageId: "", Message: importMessage, Line: 4, Column: 22, EndLine: 4, EndColumn: 29},
					{MessageId: "", Message: exportMessage, Line: 6, Column: 1, EndLine: 6, EndColumn: 15},
					{MessageId: "", Message: exportMessage, Line: 7, Column: 1, EndLine: 7, EndColumn: 10},
				},
			},
			{
				Code: `var a = b && require("c")

if (typeof window !== "undefined") {
  require('that-ugly-thing');
}

var fs = null;
try {
  fs = require("fs")
} catch (error) {}`,
				Options: []any{map[string]any{"allowConditionalRequire": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 14, EndLine: 1, EndColumn: 21},
					{MessageId: "", Message: importMessage, Line: 4, Column: 3, EndLine: 4, EndColumn: 10},
					{MessageId: "", Message: importMessage, Line: 9, Column: 8, EndLine: 9, EndColumn: 15},
				},
			},
			{
				Code: `/*eslint no-commonjs: [2, { allowPrimitiveModules: true }]*/

module.exports = { x: "y" }
exports.z = function boop() { /* ... */ }`,
				Options: []any{map[string]any{"allowPrimitiveModules": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 15},
					{MessageId: "", Message: exportMessage, Line: 4, Column: 1, EndLine: 4, EndColumn: 10},
				},
			},
		},
	)
}
