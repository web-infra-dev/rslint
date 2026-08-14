// TestNoVoidUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/no-void.js 1:1. Position assertions cover
// line/column/endLine/endColumn for every invalid case. rslint-specific
// lock-in cases live in the no_void_extras_test.go file.
package no_void

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoVoidUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoVoidRule,
		[]rule_tester.ValidTestCase{
			// ---- valid ----
			{Code: `var foo = bar()`},
			{Code: `foo.void()`},
			{Code: `foo.void = bar`},
			{Code: `delete foo;`},
			{
				Code:    `void 0`,
				Options: []any{map[string]any{"allowAsStatement": true}},
			},
			{
				Code:    `void(0)`,
				Options: []any{map[string]any{"allowAsStatement": true}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid ----
			{
				Code: `void 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Message: "Expected 'undefined' and instead saw 'void'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code:    `void 0`,
				Options: []any{map[string]any{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code:    `void 0`,
				Options: []any{map[string]any{"allowAsStatement": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code: `void(0)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code: `var foo = void 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    `var foo = void 0`,
				Options: []any{map[string]any{"allowAsStatement": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
		},
	)
}
