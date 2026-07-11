// TestRequireArrayJoinSeparatorExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch, Dimension 4 row, or real-user
// scenario it covers.
package require_array_join_separator_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	require_array_join_separator "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_array_join_separator"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireArrayJoinSeparatorExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_array_join_separator.RequireArrayJoinSeparatorRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: computed and literal member keys are excluded ----
			{Code: "foo[`join`]()", FileName: "file.ts"},
			{Code: `foo[0]()`, FileName: "file.ts"},
			{Code: `foo[Symbol.join]()`, FileName: "file.ts"},
			{Code: `Array.prototype["join"].call(foo)`, FileName: "file.ts"},
			// ---- Dimension 4: optional call is excluded ----
			{Code: `foo.join.call?.(foo)`, FileName: "file.ts"},
			{Code: `[].join.call?.(foo)`, FileName: "file.ts"},
			// ---- Dimension 4: non-empty array receiver is excluded ----
			{Code: `[,].join.call(foo)`, FileName: "file.ts"},
			// N/A: private member keys are not valid on these unrelated receivers.
			// N/A: declaration/container forms; the rule only targets call expressions.
			// N/A: nesting and ancestor walks; the rule performs no ancestor traversal.
			// N/A: body-absent declarations and empty containers; only call arguments are inspected.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver wrappers ----
			{
				Code:     `(foo).join()`,
				FileName: "file.ts",
				Output:   []string{`(foo).join(',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 11, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:     `((foo)).join()`,
				FileName: "file.ts",
				Output:   []string{`((foo)).join(',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 13, EndLine: 1, EndColumn: 15}},
			},
			// ---- Dimension 4: TypeScript receiver wrappers ----
			{
				Code:     `foo!.join()`,
				FileName: "file.ts",
				Output:   []string{`foo!.join(',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 10, EndLine: 1, EndColumn: 12}},
			},
			{
				Code:     `(foo as any).join()`,
				FileName: "file.ts",
				Output:   []string{`(foo as any).join(',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 18, EndLine: 1, EndColumn: 20}},
			},
			{
				Code:     `(foo satisfies ArrayLike<unknown>).join()`,
				FileName: "file.ts",
				Output:   []string{`(foo satisfies ArrayLike<unknown>).join(',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 40, EndLine: 1, EndColumn: 42}},
			},
			// ---- Dimension 4: parenthesized prototype receivers ----
			{
				Code:     `( [] ).join.call(foo)`,
				FileName: "file.ts",
				Output:   []string{`( [] ).join.call(foo, ',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 21, EndLine: 1, EndColumn: 22}},
			},
			{
				Code:     `(Array.prototype).join.call(foo)`,
				FileName: "file.ts",
				Output:   []string{`(Array.prototype).join.call(foo, ',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 32, EndLine: 1, EndColumn: 33}},
			},
			// ---- Dimension 3: type arguments and comments are preserved ----
			{
				Code:     `foo.join<string>()`,
				FileName: "file.ts",
				Output:   []string{`foo.join<string>(',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 17, EndLine: 1, EndColumn: 19}},
			},
			{
				Code:     `foo.join( /* keep */ )`,
				FileName: "file.ts",
				Output:   []string{`foo.join( /* keep */ ',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 9, EndLine: 1, EndColumn: 23}},
			},
			// ---- Dimension 3: nested closing parentheses determine the report range ----
			{
				Code:     `[].join.call(f(x,))`,
				FileName: "file.ts",
				Output:   []string{`[].join.call(f(x,), ',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 19, EndLine: 1, EndColumn: 20}},
			},
			{
				Code:     `[].join.call((foo))`,
				FileName: "file.ts",
				Output:   []string{`[].join.call((foo), ',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 19, EndLine: 1, EndColumn: 20}},
			},
			// ---- Dimension 3: regex parentheses inside the argument are not re-scanned ----
			{
				Code:     `[].join.call(s.replace(/\)/g, ''))`,
				FileName: "file.ts",
				Output:   []string{`[].join.call(s.replace(/\)/g, ''), ',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 34, EndLine: 1, EndColumn: 35}},
			},
			{
				Code:     `[].join.call(s.split(/\(/))`,
				FileName: "file.ts",
				Output:   []string{`[].join.call(s.split(/\(/), ',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 27, EndLine: 1, EndColumn: 28}},
			},
			// ---- Real-user: array-like values borrowed through an empty array ----
			// Locks in upstream isPrototypeProperty() empty-array arm.
			{
				Code:     `[].join.call(arrayLike)`,
				FileName: "file.ts",
				Output:   []string{`[].join.call(arrayLike, ',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 23, EndLine: 1, EndColumn: 24}},
			},
			// ---- Real-user: array-like values borrowed from Array.prototype ----
			// Locks in upstream isPrototypeProperty() Array.prototype arm.
			{
				Code:     `Array.prototype.join.call(arrayLike)`,
				FileName: "file.ts",
				Output:   []string{`Array.prototype.join.call(arrayLike, ',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 36, EndLine: 1, EndColumn: 37}},
			},
			// Locks in upstream create() direct isMethodCall() arm and appendArgument() no-argument arm.
			{
				Code:     `foo.join()`,
				FileName: "file.ts",
				Output:   []string{`foo.join(',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 9, EndLine: 1, EndColumn: 11}},
			},
			// Locks in upstream appendArgument() existing-argument arm.
			{
				Code:     `[].join.call(foo)`,
				FileName: "file.ts",
				Output:   []string{`[].join.call(foo, ',')`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 17, EndLine: 1, EndColumn: 18}},
			},
			// Locks in upstream appendArgument() trailing-comma arm.
			{
				Code:     `[].join.call(foo,)`,
				FileName: "file.ts",
				Output:   []string{`[].join.call(foo, ',',)`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: 1, Column: 18, EndLine: 1, EndColumn: 19}},
			},
		},
	)
}
