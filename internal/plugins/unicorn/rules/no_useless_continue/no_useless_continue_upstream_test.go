// Ported from eslint-plugin-unicorn v76.0.0 test/no-useless-continue.js and documentation; see LICENSE.
package no_useless_continue_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_continue"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessContinueUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "for (const x of xs) {\n\tif (skip(x)) {\n\t\tcontinue;\n\t}\n\n\tprocess(x);\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\tif (a) {\n\t\tcontinue;\n\t}\n\n\tdoMore();\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\tif (a) {\n\t\tcontinue;\n\t} else {\n\t\tdoMore();\n\t}\n\n\tuse(x);\n}\n", FileName: "case.js"},
		{Code: "while (cond) continue;", FileName: "case.js"},
		{Code: "for (;;) continue;", FileName: "case.js"},
		{Code: "for (const x of xs) continue;", FileName: "case.js"},
		{Code: "outer: for (const x of xs) {\n\tfor (const y of ys) {\n\t\tcontinue outer;\n\t}\n}\n", FileName: "case.js"},
		{Code: "loop: for (const x of xs) {\n\tdoX();\n\tcontinue loop;\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\tswitch (x) {\n\t\tcase 1:\n\t\t\tcontinue;\n\t}\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\ttry {\n\t\tcontinue;\n\t} finally {\n\t\tcleanup();\n\t}\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\ttry {\n\t\tdoX();\n\t\tcontinue;\n\t} catch {}\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\ttry {\n\t\tdoX();\n\t} catch {\n\t\tcontinue;\n\t}\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\ttry {\n\t\tdoX();\n\t} finally {\n\t\tcontinue;\n\t}\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\tfor (const y of ys) {\n\t\tif (skip(y)) {\n\t\t\tcontinue;\n\t\t}\n\n\t\tprocess(y);\n\t}\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\tcontinue;\n\tdoX();\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\tcontinue;\n\tfunction f() {}\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\tif (a) {\n\t\tcontinue;\n\t}\n\t;\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\t{\n\t\tcontinue;\n\t}\n\n\tdoMore();\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\tblock: {\n\t\tcontinue;\n\t}\n}\n", FileName: "case.js"},
		{Code: "for (const x of xs) {\n\tif (a) {\n\t\tif (b) {\n\t\t\tcontinue;\n\t\t}\n\t}\n\n\tdoMore();\n}\n", FileName: "case.js"},
		// Examples from docs/rules/no-useless-continue.md.
		{Code: "for (const item of items) {\n\tprocess(item);\n}\n", FileName: "case.js"},
		{Code: "for (const item of items) {\n\tif (shouldSkip(item)) {\n\t\tcontinue;\n\t}\n\tprocess(item);\n}\n", FileName: "case.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: "for (const x of xs) {\n\tprocess(x);\n\tcontinue;\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tprocess(x);\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 2, EndLine: 3, EndColumn: 11}}},
		{Code: "for (const x of xs) {\n\tcontinue;\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 2, Column: 2, EndLine: 2, EndColumn: 11}}},
		{Code: "while (cond) {\n\tdoX();\n\tcontinue;\n}\n", FileName: "case.js", Output: []string{"while (cond) {\n\tdoX();\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 2, EndLine: 3, EndColumn: 11}}},
		{Code: "do {\n\tdoX();\n\tcontinue;\n} while (cond);\n", FileName: "case.js", Output: []string{"do {\n\tdoX();\n} while (cond);\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 2, EndLine: 3, EndColumn: 11}}},
		{Code: "for (const x in object) {\n\tdoX();\n\tcontinue;\n}\n", FileName: "case.js", Output: []string{"for (const x in object) {\n\tdoX();\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 2, EndLine: 3, EndColumn: 11}}},
		{Code: "for (let i = 0; i < n; i++) {\n\tdoX();\n\tcontinue;\n}\n", FileName: "case.js", Output: []string{"for (let i = 0; i < n; i++) {\n\tdoX();\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 2, EndLine: 3, EndColumn: 11}}},
		{Code: "for (const x of xs) if (a) { continue; }", FileName: "case.js", Output: []string{"for (const x of xs) if (a) {  }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 1, Column: 30, EndLine: 1, EndColumn: 39}}},
		{Code: "for (const x of xs) {\n\tif (a) {\n\t\tcontinue;\n\t}\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tif (a) {\n\t}\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 3, EndLine: 3, EndColumn: 12}}},
		{Code: "for (const x of xs) {\n\tif (a) {\n\t\tif (b) {\n\t\t\tcontinue;\n\t\t}\n\t}\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tif (a) {\n\t\tif (b) {\n\t\t}\n\t}\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 4, Column: 4, EndLine: 4, EndColumn: 13}}},
		{Code: "for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t\tcontinue;\n\t} else {\n\t\tdoY();\n\t}\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else {\n\t\tdoY();\n\t}\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 4, Column: 3, EndLine: 4, EndColumn: 12}}},
		{Code: "for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else {\n\t\tdoY();\n\t\tcontinue;\n\t}\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else {\n\t\tdoY();\n\t}\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 6, Column: 3, EndLine: 6, EndColumn: 12}}},
		{Code: "for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else if (b) {\n\t\tdoY();\n\t\tcontinue;\n\t}\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else if (b) {\n\t\tdoY();\n\t}\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 6, Column: 3, EndLine: 6, EndColumn: 12}}},
		{Code: "async function run() {\n\tfor await (const x of xs) {\n\t\tprocess(x);\n\t\tcontinue;\n\t}\n}\n", FileName: "case.js", Output: []string{"async function run() {\n\tfor await (const x of xs) {\n\t\tprocess(x);\n\t}\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 4, Column: 3, EndLine: 4, EndColumn: 12}}},
		{Code: "for (const x of xs) {\n\t{\n\t\tcontinue;\n\t}\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\t{\n\t}\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 3, EndLine: 3, EndColumn: 12}}},
		{Code: "for (const x of xs) {\n\tfor (const y of ys) {\n\t\tprocess(y);\n\t\tcontinue;\n\t}\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tfor (const y of ys) {\n\t\tprocess(y);\n\t}\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 4, Column: 3, EndLine: 4, EndColumn: 12}}},
		{Code: "for (const x of xs) {\n\tfor (const y of ys) {\n\t\tcontinue;\n\t}\n\n\tcontinue;\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tfor (const y of ys) {\n\t}\n\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 3, EndLine: 3, EndColumn: 12}, {MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 6, Column: 2, EndLine: 6, EndColumn: 11}}},
		{Code: "outer: for (const x of xs) {\n\tdoX();\n\tcontinue;\n}\n", FileName: "case.js", Output: []string{"outer: for (const x of xs) {\n\tdoX();\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 2, EndLine: 3, EndColumn: 11}}},
		{Code: "for (const x of xs) {\n\tcontinue;\n\tcontinue;\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tcontinue;\n}\n", "for (const x of xs) {\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 2, EndLine: 3, EndColumn: 11}}},
		{Code: "for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else {\n\t\tif (b) {\n\t\t\tcontinue;\n\t\t}\n\t}\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tif (a) {\n\t\tdoX();\n\t} else {\n\t\tif (b) {\n\t\t}\n\t}\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 6, Column: 4, EndLine: 6, EndColumn: 13}}},
		{Code: "do {\n\tif (a) {\n\t\tcontinue;\n\t}\n} while (cond);\n", FileName: "case.js", Output: []string{"do {\n\tif (a) {\n\t}\n} while (cond);\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 3, EndLine: 3, EndColumn: 12}}},
		{Code: "for (const x of xs) {\n\tdoX();\n\tcontinue; // trailing comment\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tdoX();\n\t // trailing comment\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 2, EndLine: 3, EndColumn: 11}}},
		{Code: "for (const x of xs) {\n\tdoX();\n\t// leading comment\n\tcontinue;\n}\n", FileName: "case.js", Output: []string{"for (const x of xs) {\n\tdoX();\n\t// leading comment\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 4, Column: 2, EndLine: 4, EndColumn: 11}}},
		// Example from docs/rules/no-useless-continue.md.
		{Code: "for (const item of items) {\n\tprocess(item);\n\tcontinue;\n}\n", FileName: "case.js", Output: []string{"for (const item of items) {\n\tprocess(item);\n}\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 2, EndLine: 3, EndColumn: 11}}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_continue.NoUselessContinueRule, valid, invalid)
}
