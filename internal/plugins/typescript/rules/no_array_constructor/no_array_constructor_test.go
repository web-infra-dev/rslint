package no_array_constructor

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoArrayConstructorRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoArrayConstructorRule, []rule_tester.ValidTestCase{
		// Single argument (creates array with size)
		{Code: `new Array(x);`},
		{Code: `Array(x);`},
		{Code: `new Array(9);`},
		{Code: `Array(9);`},

		// Namespaced (not global Array)
		{Code: `new foo.Array();`},
		{Code: `foo.Array();`},
		{Code: `new Array.foo();`},
		{Code: `Array.foo();`},

		// TypeScript with type arguments
		{Code: `new Array<Foo>(1, 2, 3);`},
		{Code: `new Array<Foo>();`},
		{Code: `Array<Foo>(1, 2, 3);`},
		{Code: `Array<Foo>();`},

		// Optional chaining with single argument
		{Code: `Array?.(x);`},
		{Code: `Array?.(9);`},
		{Code: `foo?.Array();`},
		{Code: `Array?.foo();`},
		{Code: `foo.Array?.();`},
		{Code: `Array.foo?.();`},
		{Code: `Array?.<Foo>(1, 2, 3);`},
		{Code: `Array?.<Foo>();`},
	}, []rule_tester.InvalidTestCase{
		// new Array (without parentheses)
		{
			Code: `new Array;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 10,
				},
			},
			Output: []string{`[];`},
		},
		// new Array()
		{
			Code: `new Array();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 12,
				},
			},
			Output: []string{`[];`},
		},
		// Array()
		{
			Code: `Array();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 8,
				},
			},
			Output: []string{`[];`},
		},
		// Optional chaining with no args
		{
			Code: `Array?.();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 10,
				},
			},
			Output: []string{`[];`},
		},
		// new Array with multiple args
		{
			Code: `new Array(x, y);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 16,
				},
			},
			Output: []string{`[x, y];`},
		},
		// Array with multiple args
		{
			Code: `Array(x, y);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 12,
				},
			},
			Output: []string{`[x, y];`},
		},
		// Optional chaining with multiple args
		{
			Code: `Array?.(x, y);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 14,
				},
			},
			Output: []string{`[x, y];`},
		},
		// new Array with numeric args
		{
			Code: `new Array(0, 1, 2);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 19,
				},
			},
			Output: []string{`[0, 1, 2];`},
		},
		// Array with numeric args
		{
			Code: `Array(0, 1, 2);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 15,
				},
			},
			Output: []string{`[0, 1, 2];`},
		},
		// Optional chaining with numeric args
		{
			Code: `Array?.(0, 1, 2);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 17,
				},
			},
			Output: []string{`[0, 1, 2];`},
		},
		// With comments (no args)
		{
			Code: `/* a */ /* b */ Array /* c */ /* d */ /* e */ /* f */?.(); /* g */ /* h */`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 58,
				},
			},
			Output: []string{`/* a */ /* b */ []; /* g */ /* h */`},
		},
		// With comments (with args)
		{
			Code: `/* a */ /* b */ Array /* c */ /* d */ /* e */ /* f */?.(x, y); /* g */ /* h */`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 62,
				},
			},
			Output: []string{`/* a */ /* b */ [x, y]; /* g */ /* h */`},
		},
		// Multi-line
		{
			Code: `
new Array(0, 1, 2);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 19,
				},
			},
			Output: []string{`
[0, 1, 2];
`},
		},
		// Multi-line with comments
		{
			Code: `
/* a */ /* b */ Array /* c */ /* d */ /* e */ /* f */?.(
  0,
  1,
  2,
); /* g */ /* h */
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useLiteral",
					Line:      2,
					Column:    17,
				},
			},
			Output: []string{`
/* a */ /* b */ [
  0,
  1,
  2,
]; /* g */ /* h */
`},
		},
	})
}
