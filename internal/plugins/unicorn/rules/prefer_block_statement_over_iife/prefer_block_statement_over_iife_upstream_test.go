// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-block-statement-over-iife.js
package prefer_block_statement_over_iife_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_block_statement_over_iife"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferBlockStatementOverIIFEUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// The shared upstream runner also checks the empty file.
		{Code: "", FileName: "case.js"},
		{Code: "const value = (() => { return getValue(); })();", FileName: "case.js"},
		{Code: "const value = (() => { getValue(); })();", FileName: "case.js"},
		{Code: "(() => value)();", FileName: "case.js"},
		{Code: "(async () => { await run(); })();", FileName: "case.js"},
		{Code: "(function * () { yield value; })();", FileName: "case.js"},
		{Code: "(function named() { run(); })();", FileName: "case.js"},
		{Code: "(function (value) { run(value); })(value);", FileName: "case.js"},
		{Code: "(() => { return value; })();", FileName: "case.js"},
		{Code: "(() => { if (condition) { return; } })();", FileName: "case.js"},
		{Code: "(() => { var value = 1; run(value); })();", FileName: "case.js"},
		{Code: "(() => { if (condition) { var value = 1; } })();", FileName: "case.js"},
		{Code: "(() => { eval(code); })();", FileName: "case.js"},
		{Code: "(() => { if (condition) { eval(code); } })();", FileName: "case.js"},
		{Code: "(() => { (eval)(code); })();", FileName: "case.js"},
		{Code: "(() => { (eval as any)('var x = 1'); })();", FileName: "case.ts"},
		{Code: "(() => { eval!('var x = 1'); })();", FileName: "case.ts"},
		{Code: "(function () { this.run(); })();", FileName: "case.js"},
		{Code: "(function () { arguments[0](); })();", FileName: "case.js"},
		{Code: "(function () { console.log(new.target); })();", FileName: "case.js"},
		{Code: "(function () { const run = () => this.run(); run(); })();", FileName: "case.js"},
		{Code: "(function () { const getArguments = () => arguments; getArguments(); })();", FileName: "case.js"},
		{Code: "(function () { const getNewTarget = () => new.target; getNewTarget(); })();", FileName: "case.js"},
		{Code: "(function () { 'use strict'; run(); })();", FileName: "case.js"},
		{Code: "(() => { 'use strict'; run(); })();", FileName: "case.js"},
		{Code: "(function () { run(); }).call(this);", FileName: "case.js"},
		{Code: "(() => { run(); })?.();", FileName: "case.js"},
		{Code: "(() => {\n\tfunction foo() {}\n\tfoo();\n})();", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		{Code: "(() => {\n\tif (condition) {\n\t\tfunction foo() {}\n\t\tfoo();\n\t}\n})();", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},

		// Documentation examples. Both incorrect examples are in the suite below.
		{Code: "{\n\tconst value = getValue();\n\trun(value);\n}", FileName: "case.js"},
		{Code: "const value = (() => {\n\treturn getValue();\n})();", FileName: "case.js"},
		{Code: "(async () => {\n\tawait run();\n})();", FileName: "case.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		{
			Code:     "(() => { eval?.(code); })();",
			FileName: "case.js",
			Output:   []string{"{ eval?.(code); }"},
			Errors:   iifeDiagnostic(1, 1, 1, 28),
		},
		{
			Code:     "(() => {\n\tconst value = getValue();\n\trun(value);\n})();",
			FileName: "case.js",
			Output:   []string{"{\n\tconst value = getValue();\n\trun(value);\n}"},
			Errors:   iifeDiagnostic(1, 1, 4, 5),
		},
		{
			Code:     "(function () {\n\tconst value = getValue();\n\trun(value);\n})();",
			FileName: "case.js",
			Output:   []string{"{\n\tconst value = getValue();\n\trun(value);\n}"},
			Errors:   iifeDiagnostic(1, 1, 4, 5),
		},
		{
			Code:     "if (condition) {\n\t(() => {\n\t\tconst value = getValue();\n\t\trun(value);\n\t})();\n}",
			FileName: "case.js",
			Output:   []string{"if (condition) {\n\t{\n\t\tconst value = getValue();\n\t\trun(value);\n\t}\n}"},
			Errors:   iifeDiagnostic(2, 2, 5, 6),
		},
		{
			Code:     "(() => {\n\t// Keep this comment.\n\trun();\n})();",
			FileName: "case.js",
			Output:   []string{"{\n\t// Keep this comment.\n\trun();\n}"},
			Errors:   iifeDiagnostic(1, 1, 4, 5),
		},
		{
			Code:     "(() => {\n\tthis.run();\n})();",
			FileName: "case.js",
			Output:   []string{"{\n\tthis.run();\n}"},
			Errors:   iifeDiagnostic(1, 1, 3, 5),
		},
		{
			Code:     "(() => {\n\tfunction nested() {\n\t\treturn value;\n\t}\n\n\trun(nested);\n})();",
			FileName: "case.js",
			Output:   []string{"{\n\tfunction nested() {\n\t\treturn value;\n\t}\n\n\trun(nested);\n}"},
			Errors:   iifeDiagnostic(1, 1, 7, 5),
		},
		{
			Code:     "(() => {\n\tfunction nested() {\n\t\tthis.run();\n\t}\n\n\trun(nested);\n})();",
			FileName: "case.js",
			Output:   []string{"{\n\tfunction nested() {\n\t\tthis.run();\n\t}\n\n\trun(nested);\n}"},
			Errors:   iifeDiagnostic(1, 1, 7, 5),
		},
		{
			Code:     "(() => {\n\tconst value = getValue() satisfies number;\n\trun(value);\n})();",
			FileName: "case.ts",
			Output:   []string{"{\n\tconst value = getValue() satisfies number;\n\trun(value);\n}"},
			Errors:   iifeDiagnostic(1, 1, 4, 5),
		},
		{
			Code:            "(() => {\n\tfunction foo() {}\n\tfoo();\n})();",
			FileName:        "case.js",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Output:          []string{"{\n\tfunction foo() {}\n\tfoo();\n}"},
			Errors:          iifeDiagnostic(1, 1, 4, 5),
		},
		{
			Code:     "(/* comment */ () => {\n\trun();\n})();",
			FileName: "case.js",
			Output:   []string{},
			Errors:   iifeDiagnostic(1, 1, 3, 5),
		},
		{
			Code:     "(() => {\n\trun();\n} /* comment */)();",
			FileName: "case.js",
			Output:   []string{},
			Errors:   iifeDiagnostic(1, 1, 3, 19),
		},
		{
			Code:     "(() => {\n\trun();\n})() /* comment */;",
			FileName: "case.js",
			Output:   []string{},
			Errors:   iifeDiagnostic(1, 1, 3, 5),
		},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_block_statement_over_iife.PreferBlockStatementOverIIFERule, valid, invalid)
}

func iifeDiagnostic(line, column, endLine, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "prefer-block-statement-over-iife",
		Message:   "Prefer a block statement over an IIFE used only for scoping.",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}}
}
