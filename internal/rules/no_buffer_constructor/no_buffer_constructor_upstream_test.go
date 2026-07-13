// TestNoBufferConstructorUpstream migrates the full valid/invalid suite from upstream tests/lib/rules/no-buffer-constructor.js 1:1.
// Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in the no_buffer_constructor_extras_test.go file.
package no_buffer_constructor

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoBufferConstructorUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoBufferConstructorRule,
		// ---- Upstream Valid cases ----
		[]rule_tester.ValidTestCase{
			{Code: "Buffer.alloc(5)"},
			{Code: "Buffer.allocUnsafe(5)"},
			{Code: "new Buffer.Foo()"},
			{Code: "Buffer.from([1, 2, 3])"},
			{Code: "foo(Buffer)"},
			{Code: "Buffer.alloc(res.body.amount)"},
			{Code: "Buffer.from(res.body.values)"},
		},
		// ---- Upstream Invalid cases ----
		[]rule_tester.InvalidTestCase{
			{
				Code: "Buffer(5)",
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
				Code: "new Buffer(5)",
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
				Code: "Buffer([1, 2, 3])",
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
				Code: "new Buffer([1, 2, 3])",
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
				Code: "new Buffer(res.body.amount)",
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
				Code: "new Buffer(res.body.values)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "deprecated",
						Message:   "new Buffer() is deprecated. Use Buffer.from(), Buffer.alloc(), or Buffer.allocUnsafe() instead.",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}
