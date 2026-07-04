package no_new_native_nonconstructor

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoNewNativeNonconstructorUpstream migrates the full valid/invalid suite from upstream
// tests/lib/rules/no-new-native-nonconstructor.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in the
// no_new_native_nonconstructor_extras_test.go file.
func TestNoNewNativeNonconstructorUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewNativeNonconstructorRule,
		[]rule_tester.ValidTestCase{
			// ---- Symbol ----
			{Code: `var foo = Symbol('foo');`},
			{Code: `function bar(Symbol) { var baz = new Symbol('baz');}`},
			{Code: `function Symbol() {} new Symbol();`},
			{Code: `new foo(Symbol);`},
			{Code: `new foo(bar, Symbol);`},

			// ---- BigInt ----
			{Code: `var foo = BigInt(9007199254740991);`},
			{Code: `function bar(BigInt) { var baz = new BigInt(9007199254740991);}`},
			{Code: `function BigInt() {} new BigInt();`},
			{Code: `new foo(BigInt);`},
			{Code: `new foo(bar, BigInt);`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Symbol ----
			{
				Code: `var foo = new Symbol('foo');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noNewNonconstructor",
						Message:   "`Symbol` cannot be called as a constructor.",
						Line:      1,
						Column:    15,
					},
				},
			},
			{
				Code: `function bar() { return function Symbol() {}; } var baz = new Symbol('baz');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 63},
				},
			},

			// ---- BigInt ----
			{
				Code: `var foo = new BigInt(9007199254740991);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noNewNonconstructor",
						Message:   "`BigInt` cannot be called as a constructor.",
						Line:      1,
						Column:    15,
					},
				},
			},
			{
				Code: `function bar() { return function BigInt() {}; } var baz = new BigInt(9007199254740991);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNewNonconstructor", Line: 1, Column: 63},
				},
			},
		},
	)
}
