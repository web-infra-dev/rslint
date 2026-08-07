// TestConsistentReturnExtrasFunctionNames locks the public diagnostic text to
// ESLint for function owners whose ESTree and tsgo parent shapes differ. The
// three message IDs are all covered because they share the same display-name
// helper but are emitted from different rule branches.
package consistent_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentReturnExtrasFunctionNames(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentReturnRule,
		nil,
		[]rule_tester.InvalidTestCase{
			// ---- PrivateIdentifier: ESLint never quotes a private member name ----
			{
				Code: `class A { #foo() { if (a) return 1; return; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Private method #foo expected a return value.",
					Line:      1,
					Column:    37,
					EndLine:   1,
					EndColumn: 44,
				}},
			},
			{
				Code: `class A { #foo() { if (a) return; return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedReturnValue",
					Message:   "Private method #foo expected no return value.",
					Line:      1,
					Column:    35,
					EndLine:   1,
					EndColumn: 44,
				}},
			},

			// ---- Parenthesized property values: ESTree removes every paren layer ----
			{
				Code: `({ foo: (() => { if (a) return 1; }) })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of method 'foo'.",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 15,
				}},
			},
			{
				Code: `({ foo: ((() => { if (a) return 1; return; })) })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Method 'foo' expected a return value.",
					Line:      1,
					Column:    36,
					EndLine:   1,
					EndColumn: 43,
				}},
			},
			{
				Code: `({ foo: ((() => { if (a) return; return 1; })) })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedReturnValue",
					Message:   "Method 'foo' expected no return value.",
					Line:      1,
					Column:    34,
					EndLine:   1,
					EndColumn: 43,
				}},
			},
			{
				Code: `class C { static #f = ((() => { if (a) return 1; })) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of static private method #f.",
					Line:      1,
					Column:    28,
					EndLine:   1,
					EndColumn: 30,
				}},
			},

			// ---- Auto-accessors: AccessorProperty is not a PropertyDefinition ----
			{
				Code: `class C { accessor x = () => { if (a) return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of arrow function.",
					Line:      1,
					Column:    27,
					EndLine:   1,
					EndColumn: 29,
				}},
			},
			{
				Code: `class C { accessor x = () => { if (a) return 1; return; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturnValue",
					Message:   "Arrow function expected a return value.",
					Line:      1,
					Column:    49,
					EndLine:   1,
					EndColumn: 56,
				}},
			},
			{
				Code: `class C { accessor x = () => { if (a) return; return 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedReturnValue",
					Message:   "Arrow function expected no return value.",
					Line:      1,
					Column:    47,
					EndLine:   1,
					EndColumn: 56,
				}},
			},
			{
				Code: `class C { static accessor #x = (async () => { if (a) return 1; }) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of async arrow function.",
					Line:      1,
					Column:    42,
					EndLine:   1,
					EndColumn: 44,
				}},
			},

			// A TypeScript expression wrapper is visible in ESTree and must stop
			// the owner walk; only parentheses are transparent.
			{
				Code: `class C { f = ((() => { if (a) return 1; }) as unknown) }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingReturn",
					Message:   "Expected to return a value at the end of arrow function.",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 22,
				}},
			},
		},
	)
}
