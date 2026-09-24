package no_amd_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_amd"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	defineMessage  = "Expected imports instead of AMD define()."
	requireMessage = "Expected imports instead of AMD require()."
)

// Every case from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-amd.js
func TestNoAmdUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_amd.NoAmdRule,
		[]rule_tester.ValidTestCase{
			{Code: `import "x";`},
			{Code: `import x from "x"`},
			{Code: `var x = require("x")`},
			{Code: `require("x")`},
			// Two arguments, but not an array.
			{Code: `require("x", "y")`},
			// Other functions and non-identifier callees.
			{Code: `setTimeout(foo, 100)`},
			{Code: `(a || b)(1, 2, 3)`},
			// Nested scopes.
			{Code: `function x() { define(["a"], function (a) {}) }`},
			{Code: `function x() { require(["a"], function (a) {}) }`},
			// Unmatched argument types/counts.
			{Code: `define(0, 1, 2)`},
			{Code: `define("a")`},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream's ESLint >= 4 branch; messages have no message IDs.
			{
				Code: `define([], function() {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: `define(["a"], function(a) { console.log(a); })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code: `require([], function() {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: `require(["a"], function(a) { console.log(a); })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
				},
			},
			// Examples from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-amd.md
			{
				Code: `define(["a", "b"], function (a, b) { /* ... */ })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: defineMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 50},
				},
			},
			{
				Code: `require(["b", "c"], function (b, c) { /* ... */ })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 51},
				},
			},
		},
	)
}
