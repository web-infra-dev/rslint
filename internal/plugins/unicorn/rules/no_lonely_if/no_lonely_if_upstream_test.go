package no_lonely_if_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_lonely_if"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const lonelyIfMessage = "Unexpected `if` as the only statement in a `if` block without `else`."

// Upstream: eslint-plugin-unicorn v76.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/no-lonely-if.js
// Includes the snapshot suite, all 14 regressions, and every documentation example.

func TestNoLonelyIfUpstreamSnapshots(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_lonely_if.NoLonelyIfRule,
		[]rule_tester.ValidTestCase{
			{Code: `if (a) {
	if (b) {
	}
} else {}`},
			{Code: `if (a) {
	if (b) {
	}
	foo();
} else {}`},
			{Code: `if (a) {
} else {
	if (y) {}
}`},
			{Code: `if (a) {
	b ? c() : d()
}`},
		},
		[]rule_tester.InvalidTestCase{
			// snapshot 1
			{
				Code: `if (a) {
	if (b) {
	}
}`,
				Output: []string{`if (a && b) {
	}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 2, EndLine: 3, EndColumn: 3},
				},
			},
			// snapshot 2
			{
				Code: `if (a) if (b) {
	foo();
}`,
				Output: []string{`if (a && b) {
	foo();
}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 8, EndLine: 3, EndColumn: 2},
				},
			},
			// snapshot 3
			{
				Code: `if (a) {
	if (b) foo();
}`,
				Output: []string{"if (a && b) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 2, EndLine: 2, EndColumn: 15},
				},
			},
			// snapshot 4
			{
				Code: `if (a) /* comment */ {
	if (b) foo();
}`,
				Output: []string{"/* comment */ if (a && b) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 2, EndLine: 2, EndColumn: 15},
				},
			},
			// snapshot 5
			{
				Code:   "if (a) if (b) foo();",
				Output: []string{"if (a && b) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 21},
				},
			},
			// snapshot 6
			{
				Code: `if (a) {
	if (b) foo()
}`,
				Output: []string{"if (a && b) foo()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 2, EndLine: 2, EndColumn: 14},
				},
			},
			// snapshot 7
			{
				Code:   "if (a) if (b);",
				Output: []string{"if (a && b);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			// snapshot 8
			{
				Code: `if (a) {
	if (b) {
		// Should not report
	}
} else if (c) {
	if (d) {
	}
}`,
				Output: []string{`if (a) {
	if (b) {
		// Should not report
	}
} else if (c && d) {
	}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 6, Column: 2, EndLine: 7, EndColumn: 3},
				},
			},
			// snapshot 9
			{
				Code: `function * foo() {
	if (a || b)
	if (a ?? b)
	if (a ? b : c)
	if (a = b)
	if (a += b)
	if (a -= b)
	if (a &&= b)
	if (yield a)
	if (a, b);
}`,
				Output: []string{`function * foo() {
	if ((a || b) && (a ?? b))
	if (a ? b : c)
	if (a = b)
	if (a += b)
	if (a -= b)
	if (a &&= b)
	if (yield a)
	if (a, b);
}`,
					`function * foo() {
	if ((a || b) && (a ?? b) && (a ? b : c))
	if (a = b)
	if (a += b)
	if (a -= b)
	if (a &&= b)
	if (yield a)
	if (a, b);
}`,
					`function * foo() {
	if ((a || b) && (a ?? b) && (a ? b : c) && (a = b))
	if (a += b)
	if (a -= b)
	if (a &&= b)
	if (yield a)
	if (a, b);
}`,
					`function * foo() {
	if ((a || b) && (a ?? b) && (a ? b : c) && (a = b) && (a += b))
	if (a -= b)
	if (a &&= b)
	if (yield a)
	if (a, b);
}`,
					`function * foo() {
	if ((a || b) && (a ?? b) && (a ? b : c) && (a = b) && (a += b) && (a -= b))
	if (a &&= b)
	if (yield a)
	if (a, b);
}`,
					`function * foo() {
	if ((a || b) && (a ?? b) && (a ? b : c) && (a = b) && (a += b) && (a -= b) && (a &&= b))
	if (yield a)
	if (a, b);
}`,
					`function * foo() {
	if ((a || b) && (a ?? b) && (a ? b : c) && (a = b) && (a += b) && (a -= b) && (a &&= b) && (yield a))
	if (a, b);
}`,
					`function * foo() {
	if ((a || b) && (a ?? b) && (a ? b : c) && (a = b) && (a += b) && (a -= b) && (a &&= b) && (yield a) && (a, b));
}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 3, Column: 2, EndLine: 10, EndColumn: 12},
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 4, Column: 2, EndLine: 10, EndColumn: 12},
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 5, Column: 2, EndLine: 10, EndColumn: 12},
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 6, Column: 2, EndLine: 10, EndColumn: 12},
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 7, Column: 2, EndLine: 10, EndColumn: 12},
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 8, Column: 2, EndLine: 10, EndColumn: 12},
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 9, Column: 2, EndLine: 10, EndColumn: 12},
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 10, Column: 2, EndLine: 10, EndColumn: 12},
				},
			},
			// snapshot 10
			{
				Code: `async function foo() {
	if (a)
	if (await a)
	if (a.b)
	if (a && b);
}`,
				Output: []string{`async function foo() {
	if (a && await a)
	if (a.b)
	if (a && b);
}`,
					`async function foo() {
	if (a && await a && a.b)
	if (a && b);
}`,
					`async function foo() {
	if (a && await a && a.b && a && b);
}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 3, Column: 2, EndLine: 5, EndColumn: 14},
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 4, Column: 2, EndLine: 5, EndColumn: 14},
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 5, Column: 2, EndLine: 5, EndColumn: 14},
				},
			},
			// snapshot 11
			{
				Code:   "if (((a || b))) if (((c || d)));",
				Output: []string{"if (((a || b)) && ((c || d)));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 33},
				},
			},
			// snapshot 12
			{
				Code: `if // 1
(
	// 2
	a // 3
		.b // 4
) // 5
{
	// 6
	if (
		// 7
		c // 8
			.d // 9
	) {
		// 10
		foo();
		// 11
	}
	// 12
}`,
				Output: []string{"// 6\n\tif // 1\n(\n\t// 2\n\ta // 3\n\t\t.b // 4\n && \n\t\t// 7\n\t\tc // 8\n\t\t\t.d // 9\n\t) // 5\n {\n\t\t// 10\n\t\tfoo();\n\t\t// 11\n\t\n\t// 12\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 9, Column: 2, EndLine: 17, EndColumn: 3},
				},
			},
			// snapshot 13
			{
				Code: `if (a) {
	if (b) foo()
}
[].forEach(bar)`,
				Output: []string{`if (a && b) foo();
[].forEach(bar)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 2, EndLine: 2, EndColumn: 14},
				},
			},
			// snapshot 14
			{
				Code: `if (a)
	if (b) foo()
;[].forEach(bar)`,
				Output: []string{`if (a && b) foo()
;[].forEach(bar)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 2, EndLine: 3, EndColumn: 2},
				},
			},
			// snapshot 15
			{
				Code: `if (a) {
	if (b) foo()
}
;[].forEach(bar)`,
				Output: []string{`if (a && b) foo()
;[].forEach(bar)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 2, EndLine: 2, EndColumn: 14},
				},
			},
			// snapshot 16
			{
				Code: `if (a) /* comment */ {
	if (b) foo()
}`,
				Output: []string{"/* comment */ if (a && b) foo()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 2, EndLine: 2, EndColumn: 14},
				},
			},
		},
	)
}

func TestNoLonelyIfUpstreamRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_lonely_if.NoLonelyIfRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// fix should not produce invalid code when another rule replaces the original range
			{
				Code: `function some(value) {
    if (value < 10) {
        if (value < 5) {
            console.log('this is a long string this is a long string this is a long string this is a long string', value);
        }
    }
    return 0
}`,
				Output: []string{`function some(value) {
    if (value < 10 && value < 5) {
            console.log('this is a long string this is a long string this is a long string this is a long string', value);
        }
    return 0
}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 3, Column: 9, EndLine: 5, EndColumn: 10},
				},
			},
			// fix should preserve text between the outer condition and block
			{
				Code:   "if (a) /* comment */ { if (b) { foo(); } }",
				Output: []string{"/* comment */ if (a && b) { foo(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 41},
				},
			},
			// fix should preserve text between the outer condition and non-block consequent
			{
				Code:   "if (a) /* comment */ { if (b) foo(); }",
				Output: []string{"/* comment */ if (a && b) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 37},
				},
			},
			// fix should preserve comments before the inner if inside the outer block
			{
				Code:   "if (a) { /* before */ if (b) foo(); }",
				Output: []string{"/* before */ if (a && b) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 23, EndLine: 1, EndColumn: 36},
				},
			},
			// fix should preserve comments after the inner if inside the outer block
			{
				Code:   "if (a) { if (b) foo(); /* after */ }",
				Output: []string{"if (a && b) foo(); /* after */ "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 23},
				},
			},
			// fix should keep pragma comments from before the inner if attached to the merged if
			{
				Code:   "if (a) { /* @keep-next */ if (b) foo(); }",
				Output: []string{"/* @keep-next */ if (a && b) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 27, EndLine: 1, EndColumn: 40},
				},
			},
			// fix should keep pragma comments from the outer condition gap attached to the merged if
			{
				Code:   "if (a) /* @keep-next */ { if (b) foo(); }",
				Output: []string{"/* @keep-next */ if (a && b) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 27, EndLine: 1, EndColumn: 40},
				},
			},
			// fix should preserve eslint-disable-next-line before the inner if
			{
				Code: `if (a) {
	// eslint-disable-next-line no-console
	if (b) console.log('foo');
}`,
				Output: []string{`// eslint-disable-next-line no-console
	if (a && b) console.log('foo');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 3, Column: 2, EndLine: 3, EndColumn: 28},
				},
			},
			// fix should preserve block eslint-disable-next-line before the inner if
			{
				Code: `if (a) {
	/* eslint-disable-next-line no-console */
	if (b) console.log('foo');
}`,
				Output: []string{`/* eslint-disable-next-line no-console */
	if (a && b) console.log('foo');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 3, Column: 2, EndLine: 3, EndColumn: 28},
				},
			},
			// fix should preserve eslint-disable-line between the outer condition and block
			{
				Code: `if (true) // eslint-disable-line no-constant-condition
{
	if (true) foo();
}`,
				Output: []string{`if (true && true) // eslint-disable-line no-constant-condition
 foo();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 3, Column: 2, EndLine: 3, EndColumn: 18},
				},
			},
			// fix should preserve comments inside merged conditions
			{
				Code:   "if (/* outer */ a) { if (b /* inner */) foo(); }",
				Output: []string{"if (/* outer */ a && b /* inner */) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 47},
				},
			},
			// fix should preserve comments between if and opening parenthesis
			{
				Code:   "if/* outer */(a) if/* inner */(b) foo();",
				Output: []string{"if/* outer */(a && /* inner */b) foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 18, EndLine: 1, EndColumn: 41},
				},
			},
			// fix should preserve ASI-safe semicolon insertion when keeping outer-gap text
			{
				Code:   "if (a) /* comment */ { if (b) foo() } [].forEach(bar)",
				Output: []string{"/* comment */ if (a && b) foo();[].forEach(bar)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 36},
				},
			},
			// fix should preserve ASI-safe semicolon insertion when keeping trailing text from the outer block
			{
				Code:   "if (a) { if (b) foo() /* after */ } [].forEach(bar)",
				Output: []string{"if (a && b) foo() /* after */ ;[].forEach(bar)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 22},
				},
			},
		},
	)
}

func TestNoLonelyIfUpstreamDocs(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_lonely_if.NoLonelyIfRule,
		[]rule_tester.ValidTestCase{
			{Code: `if (foo && bar) {
	// …
}`},
			{Code: `if (foo) {
	// …
} else if (bar && baz) {
	// …
}`},
			{Code: `if (foo) {
	// …
} else if (bar) {
	if (baz) {
		// …
	}
} else {
	// …
}`},
			{Code: "// Built-in rule `no-lonely-if` case https://eslint.org/docs/rules/no-lonely-if\nif (foo) {\n\t// …\n} else {\n\tif (bar) {\n\t\t// …\n\t}\n}"},
		},
		[]rule_tester.InvalidTestCase{
			// documentation 1
			{
				Code: `if (foo) {
	if (bar) {
		// …
	}
}`,
				Output: []string{`if (foo && bar) {
		// …
	}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			// documentation 2
			{
				Code: `if (foo) {
	// …
} else if (bar) {
	if (baz) {
		// …
	}
}`,
				Output: []string{`if (foo) {
	// …
} else if (bar && baz) {
		// …
	}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "no-lonely-if", Message: lonelyIfMessage, Line: 4, Column: 2, EndLine: 6, EndColumn: 3},
				},
			},
		},
	)
}
