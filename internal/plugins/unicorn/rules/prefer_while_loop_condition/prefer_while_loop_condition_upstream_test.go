// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-while-loop-condition.js
package prefer_while_loop_condition_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_while_loop_condition"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferWhileLoopConditionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_while_loop_condition.PreferWhileLoopConditionRule,
		[]rule_tester.ValidTestCase{
			// test.snapshot valid.
			{Code: "while (hasMore()) {\n\tprocessNext();\n}", FileName: "case.js"},
			{Code: "while (false) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js"},
			{Code: "while (1) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js"},
			{Code: "for (; false;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js"},
			{Code: "for (let index = 0;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js"},
			{Code: "for (;; update()) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js"},
			{Code: "do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} while (false);", FileName: "case.js"},
			{Code: "let done = false;\n\nwhile (true) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tlet done = true;\n\twork();\n}", FileName: "case.js"},
			{Code: "let done = false;\n\nfor (;;) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tconst done = true;\n\twork();\n}", FileName: "case.js"},
			{Code: "let done = false;\n\ndo {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tclass done {}\n\twork();\n} while (true);", FileName: "case.js"},
			{Code: "let done = false;\n\nwhile (true) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tfunction done() {}\n\twork();\n}", FileName: "case.js"},
			{Code: "loop: for (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js"},
			{Code: "loop: do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} while (true);", FileName: "case.js"},
			{Code: "while (true) {\n\tsetup();\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js"},
			{Code: "while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t} else {\n\t\tprocessNext();\n\t}\n}", FileName: "case.js"},
			{Code: "function process() {\n\twhile (true) {\n\t\tif (!hasMore()) {\n\t\t\treturn;\n\t\t}\n\t}\n}", FileName: "case.js"},
			{Code: "while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t\tprocessNext();\n\t}\n}", FileName: "case.js"},
			{Code: "loop: while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js"},
			{Code: "loop: {\n\twhile (true) {\n\t\tif (!hasMore()) {\n\t\t\tbreak loop;\n\t\t}\n\t}\n}", FileName: "case.js"},
			{Code: "while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tif (done) {\n\t\tbreak;\n\t}\n}", FileName: "case.js"},
			{Code: "while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\t{\n\t\tif (done) {\n\t\t\tbreak;\n\t\t}\n\t}\n}", FileName: "case.js"},
			// documentation example 2.
			{Code: "// ✅\nwhile (hasMore()) {\n\tprocessNext();\n}", FileName: "case.js"},
			// documentation example 4.
			{Code: "// ✅\nwhile (hasMore()) {\n\tprocessNext();\n}", FileName: "case.js"},
			// documentation example 6.
			{Code: "// ✅\nwhile (!done) {\n\tprocessNext();\n}", FileName: "case.js"},
			// documentation example 8.
			{Code: "// ✅\nwhile (hasMore()) {\n\tprocessNext();\n}", FileName: "case.js"},
		},
		[]rule_tester.InvalidTestCase{
			// test.snapshot invalid.
			{
				Code: "while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tfor (const item of items) {\n\t\tif (done(item)) {\n\t\t\tbreak;\n\t\t}\n\t}\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\n\tfor (const item of items) {\n\t\tif (done(item)) {\n\t\t\tbreak;\n\t\t}\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tswitch (value) {\n\t\tcase 1:\n\t\t\tbreak;\n\t}\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\n\tswitch (value) {\n\t\tcase 1:\n\t\t\tbreak;\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tfunction stop() {\n\t\twhile (done) {\n\t\t\tbreak;\n\t\t}\n\t}\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\n\tfunction stop() {\n\t\twhile (done) {\n\t\t\tbreak;\n\t\t}\n\t}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (!hasMore()) break;\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 2, EndColumn: 24},
				},
			},
			{
				Code: "while (true) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{"while (!done) {\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (done || cancelled) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{"while (!(done || cancelled)) {\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "for (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "for (; true;) {\n\tif (!hasMore()) break;\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 2, EndColumn: 24},
				},
			},
			{
				Code: "for (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tif (skip) {\n\t\tcontinue;\n\t}\n\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\n\tif (skip) {\n\t\tcontinue;\n\t}\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n} while (true);", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tif (skip) {\n\t\tcontinue;\n\t}\n\n\tprocessNext();\n} while (true);", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\n\tif (skip) {\n\t\tcontinue;\n\t}\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "do {\n\tif (!hasMore()) break;\n\tprocessNext();\n} while (true);", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 2, EndColumn: 24},
				},
			},
			{
				Code: "do {\n\tif (done || cancelled) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n} while (true);", FileName: "case.js",
				Output: []string{"while (!(done || cancelled)) {\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tvar done = true;\n\twork();\n}", FileName: "case.js",
				Output: []string{"while (!done) {\n\n\tvar done = true;\n\twork();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "for (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "for (; true;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} while (true);", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (!value!) {\n\t\tbreak;\n\t}\n\n\tprocess(value);\n}", FileName: "case.ts",
				Output: []string{"while (value!) {\n\n\tprocess(value);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (value!) {\n\t\tbreak;\n\t}\n\n\tprocess(value);\n}", FileName: "case.ts",
				Output: []string{"while (!(value!)) {\n\n\tprocess(value);\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (/* keep */ true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "for (;;) {\n\tif (\n\t\t// Check before processing.\n\t\t!hasMore()\n\t) {\n\t\tbreak;\n\t}\n}", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 7, EndColumn: 3},
				},
			},
			{
				Code: "for /* keep */ (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "for (; /* keep */ true;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n}", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} /* keep */ while (true);", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "do /* keep */ {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} while (true);", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "do {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n} while (true); // Preserve this comment.", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (\n\t\t// Check before processing.\n\t\t!hasMore()\n\t) {\n\t\tbreak;\n\t}\n}", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 7, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (!hasMore()) {\n\t\t// Preserve this comment.\n\t\tbreak;\n\t}\n}", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 5, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\t// Preserve this comment.\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 3, Column: 2, EndLine: 5, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (!hasMore()) break; // Preserve this comment.\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 2, EndColumn: 24},
				},
			},
			{
				Code: "while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t} // Preserve this comment.\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "while (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\t// Process the remaining work.\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{"while (hasMore()) {\n\n\t// Process the remaining work.\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: "do {\r\n  if (done) {\r\n    break;\r\n  }\r\n  processNext(/* Keep this argument comment. */ value);\r\n} while (true);", FileName: "case.js",
				Output: []string{"while (!done) {\r\n  processNext(/* Keep this argument comment. */ value);\r\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 2, Column: 3, EndLine: 4, EndColumn: 4},
				},
			},
			// documentation example 1.
			{
				Code: "// ❌\nwhile (true) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{"// ❌\nwhile (hasMore()) {\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 3, Column: 2, EndLine: 5, EndColumn: 3},
				},
			},
			// documentation example 3.
			{
				Code: "// ❌\ndo {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n} while (true);", FileName: "case.js",
				Output: []string{"// ❌\nwhile (hasMore()) {\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 3, Column: 2, EndLine: 5, EndColumn: 3},
				},
			},
			// documentation example 5.
			{
				Code: "// ❌\nwhile (true) {\n\tif (done) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{"// ❌\nwhile (!done) {\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 3, Column: 2, EndLine: 5, EndColumn: 3},
				},
			},
			// documentation example 7.
			{
				Code: "// ❌\nfor (;;) {\n\tif (!hasMore()) {\n\t\tbreak;\n\t}\n\n\tprocessNext();\n}", FileName: "case.js",
				Output: []string{"// ❌\nwhile (hasMore()) {\n\n\tprocessNext();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-while-loop-condition", Message: "Prefer putting the condition in the `while` statement.", Line: 3, Column: 2, EndLine: 5, EndColumn: 3},
				},
			},
		},
	)
}
