package no_dynamic_require_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_dynamic_require"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	requireMessage = "Calls to require() should use string literals"
	importMessage  = "Calls to import() should use string literals"
)

// Every semantic case from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-dynamic-require.js
// and https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-dynamic-require.md.
// Parser duplicates use the same tsgo representation. Upstream reports literal
// messages without message IDs, fixes, or suggestions; the Go tester checks all three.
func TestNoDynamicRequireUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_dynamic_require.NoDynamicRequireRule,
		[]rule_tester.ValidTestCase{
			// Upstream valid: imports and require calls.
			{
				Code: "import _ from \"lodash\"",
			},
			{
				Code: "require(\"foo\")",
			},
			{
				Code: "require(`foo`)",
			},
			{
				Code: "require(\"./foo\")",
			},
			{
				Code: "require(\"@scope/foo\")",
			},
			{
				Code: "require()",
			},
			{
				Code: "require(\"./foo\", \"bar\" + \"okay\")",
			},
			{
				Code: "var foo = require(\"foo\")",
			},
			{
				Code: "var foo = require(`foo`)",
			},
			{
				Code: "var foo = require(\"./foo\")",
			},
			{
				Code: "var foo = require(\"@scope/foo\")",
			},
			// Upstream dynamic imports: Espree and Babel repeat these same cases.
			{
				Code:    "import(\"foo\")",
				Options: []any{map[string]any{"esmodule": true}},
			},
			{
				Code:    "import(`foo`)",
				Options: []any{map[string]any{"esmodule": true}},
			},
			{
				Code:    "import(\"./foo\")",
				Options: []any{map[string]any{"esmodule": true}},
			},
			{
				Code:    "import(\"@scope/foo\")",
				Options: []any{map[string]any{"esmodule": true}},
			},
			{
				Code:    "var foo = import(\"foo\")",
				Options: []any{map[string]any{"esmodule": true}},
			},
			{
				Code:    "var foo = import(`foo`)",
				Options: []any{map[string]any{"esmodule": true}},
			},
			{
				Code:    "var foo = import(\"./foo\")",
				Options: []any{map[string]any{"esmodule": true}},
			},
			{
				Code:    "var foo = import(\"@scope/foo\")",
				Options: []any{map[string]any{"esmodule": true}},
			},
			// Upstream dynamic imports are allowed when esmodule is omitted.
			{
				Code: "import(\"../\" + name)",
			},
			{
				Code: "import(`../${name}`)",
			},
			// Examples from the pinned upstream documentation.
			{
				Code: "require('../name');\nrequire(`../name`);",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream invalid: require calls.
			{
				Code: "require(\"../\" + name)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: "require(`../${name}`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: "require(name)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "require(name())",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "require(name + \"foo\", \"bar\")",
				Options: []any{map[string]any{"esmodule": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},
			// Upstream dynamic imports: both parser variants have the same semantics.
			{
				Code:    "import(\"../\" + name)",
				Options: []any{map[string]any{"esmodule": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "import(`../${name}`)",
				Options: []any{map[string]any{"esmodule": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "import(name)",
				Options: []any{map[string]any{"esmodule": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    "import(name())",
				Options: []any{map[string]any{"esmodule": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			// Upstream interpolated require arguments.
			{
				Code: "require(`foo${x}`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: "var foo = require(`foo${x}`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 11, EndLine: 1, EndColumn: 29},
				},
			},
			// Examples from the pinned upstream documentation.
			{
				Code: "require(name);\nrequire('../' + name);\nrequire(`../${name}`);\nrequire(name());",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: requireMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
					{MessageId: "", Message: requireMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 22},
					{MessageId: "", Message: requireMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 22},
					{MessageId: "", Message: requireMessage, Line: 4, Column: 1, EndLine: 4, EndColumn: 16},
				},
			},
		},
	)
}
