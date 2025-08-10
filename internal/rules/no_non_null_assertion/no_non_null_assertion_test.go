package no_non_null_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoNonNullAssertionRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// Basic valid cases - no non-null assertions
		{Code: `const foo = "hello"; console.log(foo);`},
		{Code: `function foo(bar: string) { console.log(bar); }`},
		{Code: `const foo: string | null = "hello"; if (foo) { console.log(foo); }`},
		{Code: `const foo: string | undefined = "hello"; if (foo !== undefined) { console.log(foo); }`},
		{Code: `const foo: string | null = "hello"; const bar = foo || "default";`},
		{Code: `const foo: string | null = "hello"; const bar = foo ?? "default";`},
		{Code: `const foo: string | null = "hello"; const bar = foo?.length || 0;`},
		{Code: `const foo: string | null = "hello"; if (foo !== null) { console.log(foo); }`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// Basic non-null assertion - should report error
		{
			Code: `const foo: string | null = "hello"; const bar = foo!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in property access
		{
			Code: `const foo: string | null = "hello"; const bar = foo!.length;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in function call
		{
			Code: `const foo: string | null = "hello"; const bar = foo!.toUpperCase();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in array access
		{
			Code: `const foo: string[] | null = ["hello"]; const bar = foo![0];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Chained non-null assertions - should report 2 errors
		{
			Code: `const foo: string | null = "hello"; const bar = foo!!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in conditional expression
		{
			Code: `const foo: string | null = "hello"; const bar = foo! ? "yes" : "no";`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in logical expression
		{
			Code: `const foo: string | null = "hello"; const bar = foo! && "yes";`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in return statement
		{
			Code: `function test(): string { const foo: string | null = "hello"; return foo!; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in variable declaration
		{
			Code: `let foo: string | null = "hello"; foo = foo!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in parameters
		{
			Code: `function test(foo: string | null) { const bar = foo!; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in object properties
		{
			Code: `const obj = { foo: "hello" as string | null }; const bar = obj.foo!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in template strings
		{
			Code: "const foo: string | null = \"hello\"; const bar = `Value: ${foo!}`;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in type assertion
		{
			Code: `const foo: string | null = "hello"; const bar = (foo! as string).length;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in generics
		{
			Code: `function test<T extends string | null>(foo: T): T { return foo!; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in union types
		{
			Code: `const foo: (string | null)[] = ["hello"]; const bar = foo[0]!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in nested expressions
		{
			Code: `const foo: string | null = "hello"; const bar = (foo! + "world").length;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// Non-null assertion in ternary expressions
		{
			Code: `const foo: string | null = "hello"; const bar = foo! ? foo!.length : 0;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
				{
					MessageId: "noNonNull",
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNonNullAssertionRule, validTestCases, invalidTestCases)
}

func TestNoNonNullAssertionOptionsParsing(t *testing.T) {
	// Test basic rule information
	rule := NoNonNullAssertionRule
	if rule.Name != "@typescript-eslint/no-non-null-assertion" {
		t.Errorf("Expected rule name to be '@typescript-eslint/no-non-null-assertion', got %s", rule.Name)
	}
}

func TestNoNonNullAssertionMessage(t *testing.T) {
	msg := buildNoNonNullAssertionMessage()
	if msg.Id != "noNonNull" {
		t.Errorf("Expected message ID to be 'noNonNull', got %s", msg.Id)
	}
	if msg.Description != "Non-null assertion operator (!) is not allowed." {
		t.Errorf("Expected description to be 'Non-null assertion operator (!) is not allowed.', got %s", msg.Description)
	}
}

// Test edge cases
func TestNoNonNullAssertionEdgeCases(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// Nested assignment expressions
		{Code: `let obj: { prop?: string } = {}; obj.prop! = "value";`},

		// Non-null assertion in destructuring assignment
		{Code: `let arr: (string | null)[] = ["hello"]; [arr[0]!] = ["world"];`},

		// Complex assignment expressions
		{Code: `let foo: string | null = "hello"; (foo! as any) = "world";`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// These test cases are already included in the main test function
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNonNullAssertionRule, validTestCases, invalidTestCases)
}
