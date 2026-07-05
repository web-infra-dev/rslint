// TestNoEmptyStaticBlockUpstream migrates the full valid/invalid suite from
// upstream eslint/tests/lib/rules/no-empty-static-block.js 1:1. Position
// assertions cover line/column/endLine/endColumn for every invalid case.
// rslint-specific lock-in cases live in the no_empty_static_block_extras_test.go
// file.
package no_empty_static_block

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyStaticBlockUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyStaticBlockRule,
		[]rule_tester.ValidTestCase{
			// ---- valid ----
			{Code: `class Foo { static { bar(); } }`},
			{Code: `class Foo { static { /* comments */ } }`},
			{Code: "class Foo { static {\n// comment\n} }"},
			{Code: `class Foo { static { bar(); } static { bar(); } }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid ----
			invalidStaticBlockCase(`class Foo { static {} }`, 1, 20, 1, 22, `class Foo { static { /* empty */ } }`),
			invalidStaticBlockCase(`class Foo { static { } }`, 1, 20, 1, 23, `class Foo { static { /* empty */ } }`),
			invalidStaticBlockCase("class Foo { static { \n\n } }", 1, 20, 3, 3, `class Foo { static { /* empty */ } }`),
			invalidStaticBlockCase(`class Foo { static { bar(); } static {} }`, 1, 38, 1, 40, `class Foo { static { bar(); } static { /* empty */ } }`),
			invalidStaticBlockCase("class Foo { static // comment\n {} }", 2, 2, 2, 4, "class Foo { static // comment\n { /* empty */ } }"),
			invalidStaticBlockCase(`class Foo { static /* empty */ {} /* empty */ }`, 1, 32, 1, 34, `class Foo { static /* empty */ { /* empty */ } /* empty */ }`),
		},
	)
}

func invalidStaticBlockCase(code string, line int, column int, endLine int, endColumn int, output string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "unexpected",
				Message:   "Unexpected empty static block.",
				Line:      line,
				Column:    column,
				EndLine:   endLine,
				EndColumn: endColumn,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{
						MessageId: "suggestComment",
						Output:    output,
					},
				},
			},
		},
	}
}
