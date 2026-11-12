// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-accidental-bitwise-operator.js
package no_accidental_bitwise_operator_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_accidental_bitwise_operator"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAccidentalBitwiseOperatorUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_accidental_bitwise_operator.NoAccidentalBitwiseOperatorRule, []rule_tester.ValidTestCase{
		// Snapshot cases
		// Intentional bitwise operations
		jsCase("a | b;"),
		jsCase("a & b;"),
		jsCase("flags & MASK;"),
		jsCase("x | 0;"),
		jsCase("x | 1;"),
		jsCase("options | someVariable;"),
		jsCase("a ^ b;"),
		jsCase("a << b;"),
		jsCase("a >> b;"),
		jsCase("a >>> b;"),
		jsCase("~x;"),
		jsCase("x | (a + b);"),
		jsCase("foo() | bar();"),
		// Numeric coercion, BigInt, and regular expressions
		jsCase("x | null;"),
		jsCase("x | 1n;"),
		jsCase("x | /regex/;"),
		// Logical operators
		jsCase("a && b;"),
		jsCase("a || b;"),
		jsCase("obj && obj.prop;"),
		// Operands outside the same-identifier guard pattern
		jsCase("obj1 & obj2.a;"),
		jsCase("obj.a & obj.b;"),
		jsCase("obj & obj.a.b;"),
		jsCase("a & b.c;"),
		jsCase("this & this.a;"),
		jsCase("obj & obj;"),
		jsCase("obj & obj.prop();"),
		jsCase("obj & obj?.a;"),
		// Other assignments and numeric or unknown fallback values
		jsCase("x &= {};"),
		jsCase("x &= 1;"),
		jsCase("x |= 1;"),
		jsCase("x |= y;"),
		// TypeScript assertion
		{Code: "options | ({} as Foo);", FileName: "case.ts", LanguageOptions: module},
		// Documentation examples (https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-accidental-bitwise-operator.md)
		jsCase("if (object && object.property) {\n\t// …\n}"),
		jsCase("options = options || {};"),
		jsCase("input ||= '';"),
		jsCase("const masked = flags & MASK;\nconst truncated = value | 0;"),
	}, []rule_tester.InvalidTestCase{
		// Grouping and statement boundaries
		suggestionCase("ready && value | '';", "ready && (value || '');", "|", 1, 16),
		suggestionCase("obj & obj.prop | fallback;", "(obj && obj.prop) | fallback;", "&", 1, 5),
		suggestionCase("value | '' ?? fallback;", "(value || '') ?? fallback;", "|", 1, 7),
		suggestionCase("fallback ?? value | '';", "fallback ?? (value || '');", "|", 1, 19),
		suggestionCase("ready && (value | '');", "ready && (value || '');", "|", 1, 17),
		suggestionCase("ready && value /* keep */ | '';", "ready && (value /* keep */ || '');", "|", 1, 27),
		suggestionCase("(obj & obj.prop) === expected;", "(obj && obj.prop) === expected;", "&", 1, 6),
		suggestionCase("foo()\nobj & obj.prop | fallback", "foo()\n;(obj && obj.prop) | fallback", "&", 2, 5),
		suggestionCase("foo()\nvalue | '' ?? fallback", "foo()\n;(value || '') ?? fallback", "|", 2, 7),
		suggestionCase("if (ready)\nobj & obj.prop | fallback", "if (ready)\n(obj && obj.prop) | fallback", "&", 2, 5),
		// Snapshot cases
		// Short-circuit guards
		suggestionCase("obj & obj.a;", "obj && obj.a;", "&", 1, 5),
		suggestionCase("if (obj & obj.prop) {}", "if (obj && obj.prop) {}", "&", 1, 9),
		suggestionCase("obj & obj[key];", "obj && obj[key];", "&", 1, 5),
		suggestionCase("(obj) & obj.a;", "(obj) && obj.a;", "&", 1, 7),
		suggestionCase("obj /* comment */ & obj.a;", "obj /* comment */ && obj.a;", "&", 1, 19),
		// Fallback expressions
		suggestionCase("options | {};", "options || {};", "|", 1, 9),
		suggestionCase("options | '';", "options || '';", "|", 1, 9),
		suggestionCase("options | true;", "options || true;", "|", 1, 9),
		suggestionCase("options | [];", "options || [];", "|", 1, 9),
		suggestionCase("options | `template`;", "options || `template`;", "|", 1, 9),
		suggestionCase("x | function () {};", "x || function () {};", "|", 1, 3),
		suggestionCase("x | (() => {});", "x || (() => {});", "|", 1, 3),
		suggestionCase("x | class {};", "x || class {};", "|", 1, 3),
		suggestionCase("foo() | {};", "foo() || {};", "|", 1, 7),
		suggestionCase("a.b | {};", "a.b || {};", "|", 1, 5),
		suggestionCase("a | b | {};", "a | b || {};", "|", 1, 7),
		// Fallback assignments
		suggestionCase("input |= '';", "input ||= '';", "|=", 1, 7),
		suggestionCase("input |= {};", "input ||= {};", "|=", 1, 7),
		suggestionCase("input |= false;", "input ||= false;", "|=", 1, 7),
		// TypeScript
		withFileName(suggestionCase("obj & obj.a;", "obj && obj.a;", "&", 1, 5), "case.ts"),
		withFileName(suggestionCase("options | {};", "options || {};", "|", 1, 9), "case.ts"),
		// Compatible logical precedence
		suggestionCase("ready || object & object.value;", "ready || object && object.value;", "&", 1, 17),
		suggestionCase("ready && object & object.value;", "ready && object && object.value;", "&", 1, 17),
		suggestionCase("value | '' || fallback;", "value || '' || fallback;", "|", 1, 7),
		// Documentation examples
		suggestionCase("if (object & object.property) {\n\t// …\n}", "if (object && object.property) {\n\t// …\n}", "&", 1, 12),
		suggestionCase("options = options | {};", "options = options || {};", "|", 1, 19),
		suggestionCase("input |= '';", "input ||= '';", "|=", 1, 7),
	})
}

var module = rule.LanguageOptions{SourceType: "module"}

func jsCase(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "case.js", LanguageOptions: module}
}

func suggestionCase(code, output, bitwise string, line, column int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code, FileName: "case.js", LanguageOptions: module,
		Errors: []rule_tester.InvalidTestCaseError{operatorError(bitwise, line, column, output)},
	}
}

func withFileName(testCase rule_tester.InvalidTestCase, fileName string) rule_tester.InvalidTestCase {
	testCase.FileName = fileName
	return testCase
}

func operatorError(bitwise string, line, column int, output string) rule_tester.InvalidTestCaseError {
	logical := map[string]string{"&": "&&", "|": "||", "|=": "||="}[bitwise]
	return rule_tester.InvalidTestCaseError{
		MessageId: "no-accidental-bitwise-operator/error",
		Message:   "Unexpected bitwise operator `" + bitwise + "`. Did you mean the logical operator `" + logical + "`?",
		Line:      line, Column: column, EndLine: line, EndColumn: column + len(bitwise),
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: "no-accidental-bitwise-operator/suggestion", Output: output},
		},
	}
}
