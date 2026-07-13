// TestNoBufferConstructorExtras locks in branches and edge shapes that the upstream test suite does not cover.
// Position assertions cover line/column for every invalid case.
// Upstream-migrated cases live in the no_buffer_constructor_upstream_test.go file.
package no_buffer_constructor

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoBufferConstructorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoBufferConstructorRule,
		// ---- Extras Valid cases ----
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: Access / key forms ----
			{Code: "new Buffer.from(5)"},

			// ---- Dimension 4: Declaration / container forms (shadowing) ----
			{Code: "function run(Buffer) { Buffer(5); }"},
			{Code: "var Buffer = require('buffer'); new Buffer(5);"},
			{Code: "let Buffer = class {}; new Buffer(5);"},
			{Code: "const Buffer = 1; Buffer(5);"},
			{Code: "try {} catch (Buffer) { Buffer(5); }"},

			// ---- Real-user: global override ----
			{Code: "new Buffer(5);", Globals: map[string]bool{"Buffer": false}},
		},
		// ---- Extras Invalid cases ----
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: Receiver / expression wrappers (parentheses) ----
			{
				Code: "(Buffer)(5)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "deprecated",
						Message:   "Buffer() is deprecated. Use Buffer.from(), Buffer.alloc(), or Buffer.allocUnsafe() instead.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "new (Buffer)(5)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "deprecated",
						Message:   "new Buffer() is deprecated. Use Buffer.from(), Buffer.alloc(), or Buffer.allocUnsafe() instead.",
						Line:      1,
						Column:    1,
					},
				},
			},
			// ---- Dimension 4: Receiver / expression wrappers (TS type expressions) ----
			{
				Code: "(Buffer as any)(5)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "deprecated",
						Message:   "Buffer() is deprecated. Use Buffer.from(), Buffer.alloc(), or Buffer.allocUnsafe() instead.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "new (Buffer as any)(5)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "deprecated",
						Message:   "new Buffer() is deprecated. Use Buffer.from(), Buffer.alloc(), or Buffer.allocUnsafe() instead.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "(Buffer satisfies any)(5)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "deprecated",
						Message:   "Buffer() is deprecated. Use Buffer.from(), Buffer.alloc(), or Buffer.allocUnsafe() instead.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "Buffer!(5)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "deprecated",
						Message:   "Buffer() is deprecated. Use Buffer.from(), Buffer.alloc(), or Buffer.allocUnsafe() instead.",
						Line:      1,
						Column:    1,
					},
				},
			},
			// ---- Dimension 4: Optional chaining ----
			{
				Code: "Buffer?.(5)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "deprecated",
						Message:   "Buffer() is deprecated. Use Buffer.from(), Buffer.alloc(), or Buffer.allocUnsafe() instead.",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}
