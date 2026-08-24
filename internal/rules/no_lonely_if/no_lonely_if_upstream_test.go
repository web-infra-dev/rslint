// TestNoLonelyIfUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/no-lonely-if.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the no_lonely_if_extras_test.go file.
package no_lonely_if

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoLonelyIfUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLonelyIfRule,
		[]rule_tester.ValidTestCase{
			{Code: "if (a) {;} else if (b) {;}"},
			{Code: "if (a) {;} else { if (b) {;} ; }"},
			{Code: "if (a) if (a) {} else { if (b) {} } else {}"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "if (a) {;} else { if (b) {;} }",
				Output: []string{"if (a) {;} else if (b) {;}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 1, Column: 19},
				},
			},
			{
				Code: "if (a) {\n" +
					"  foo();\n" +
					"} else {\n" +
					"  if (b) {\n" +
					"    bar();\n" +
					"  }\n" +
					"}",
				Output: []string{"if (a) {\n" +
					"  foo();\n" +
					"} else if (b) {\n" +
					"    bar();\n" +
					"  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 4, Column: 3},
				},
			},
			{
				Code: "if (a) {\n" +
					"  foo();\n" +
					"} else /* comment */ {\n" +
					"  if (b) {\n" +
					"    bar();\n" +
					"  }\n" +
					"}",
				Output: []string{"if (a) {\n" +
					"  foo();\n" +
					"} else /* comment */ if (b) {\n" +
					"    bar();\n" +
					"  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 4, Column: 3},
				},
			},
			{
				Code: "if (a) {\n" +
					"  foo();\n" +
					"} else {\n" +
					"  /* otherwise, do the other thing */ if (b) {\n" +
					"    bar();\n" +
					"  }\n" +
					"}",
				// Not fixed: a comment sits between the else block's brace and the inner if.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 4, Column: 39},
				},
			},
			{
				Code: "if (a) {\n" +
					"  foo();\n" +
					"} else {\n" +
					"  if /* this comment is ok */ (b) {\n" +
					"    bar();\n" +
					"  }\n" +
					"}",
				Output: []string{"if (a) {\n" +
					"  foo();\n" +
					"} else if /* this comment is ok */ (b) {\n" +
					"    bar();\n" +
					"  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 4, Column: 3},
				},
			},
			{
				Code: "if (a) {\n" +
					"  foo();\n" +
					"} else {\n" +
					"  if (b) {\n" +
					"    bar();\n" +
					"  } /* this comment will prevent this test case from being autofixed. */\n" +
					"}",
				// Not fixed: a comment sits between the inner if and the else block's closing brace.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 4, Column: 3},
				},
			},
			{
				Code:   "if (foo) {} else { if (bar) baz(); }",
				Output: []string{"if (foo) {} else if (bar) baz();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 1, Column: 20},
				},
			},
			{
				// Not fixed; removing the braces would cause a SyntaxError.
				Code: "if (foo) {} else { if (bar) baz() } qux();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 1, Column: 20},
				},
			},
			{
				// This is fixed because there is a semicolon after baz().
				Code:   "if (foo) {} else { if (bar) baz(); } qux();",
				Output: []string{"if (foo) {} else if (bar) baz(); qux();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 1, Column: 20},
				},
			},
			{
				// Not fixed; removing the braces would change the semantics due to ASI.
				Code: "if (foo) {\n" +
					"} else {\n" +
					"  if (bar) baz()\n" +
					"}\n" +
					"[1, 2, 3].forEach(foo);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 3, Column: 3},
				},
			},
			{
				// Not fixed; removing the braces would change the semantics due to ASI.
				Code: "if (foo) {\n" +
					"} else {\n" +
					"  if (bar) baz++\n" +
					"}\n" +
					"foo;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 3, Column: 3},
				},
			},
			{
				// This is fixed because there is a semicolon after baz++
				Code: "if (foo) {\n" +
					"} else {\n" +
					"  if (bar) baz++;\n" +
					"}\n" +
					"foo;",
				Output: []string{"if (foo) {\n" +
					"} else if (bar) baz++;\n" +
					"foo;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 3, Column: 3},
				},
			},
			{
				// Not fixed; bar() would be interpreted as a template literal tag.
				Code: "if (a) {\n" +
					"  foo();\n" +
					"} else {\n" +
					"  if (b) bar()\n" +
					"}\n" +
					"`template literal`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 4, Column: 3},
				},
			},
			{
				Code: "if (a) {\n" +
					"  foo();\n" +
					"} else {\n" +
					"  if (b) {\n" +
					"    bar();\n" +
					"  } else if (c) {\n" +
					"    baz();\n" +
					"  } else {\n" +
					"    qux();\n" +
					"  }\n" +
					"}",
				Output: []string{"if (a) {\n" +
					"  foo();\n" +
					"} else if (b) {\n" +
					"    bar();\n" +
					"  } else if (c) {\n" +
					"    baz();\n" +
					"  } else {\n" +
					"    qux();\n" +
					"  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 4, Column: 3},
				},
			},
		},
	)
}
