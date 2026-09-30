// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/consistent-assert.js
package consistent_assert_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_assert"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "consistent-assert/error"

func valid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:            code,
		FileName:        "file.js",
		LanguageOptions: rule.LanguageOptions{SourceType: "module"},
	}
}

func validTS(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:            code,
		FileName:        "file.ts",
		LanguageOptions: rule.LanguageOptions{SourceType: "module"},
	}
}

func location(code, target string, occurrence int) (line, column, endLine, endColumn int) {
	searchFrom := 0
	start := -1
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], target)
		if relative < 0 {
			panic("target not found in consistent-assert fixture: " + target)
		}
		start = searchFrom + relative
		searchFrom = start + len(target)
	}

	line = strings.Count(code[:start], "\n") + 1
	lineStart := strings.LastIndex(code[:start], "\n") + 1
	column = start - lineStart + 1
	end := start + len(target)
	endLine = strings.Count(code[:end], "\n") + 1
	endLineStart := strings.LastIndex(code[:end], "\n") + 1
	endColumn = end - endLineStart + 1
	return
}

func expectedError(code, name string, occurrence int) rule_tester.InvalidTestCaseError {
	line, column, endLine, endColumn := location(code, name, occurrence)
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   "Prefer `" + name + ".ok(…)` over `" + name + "(…)`.",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func expectedFooCallError(code, name string) rule_tester.InvalidTestCaseError {
	line, column, _, _ := location(code, name+"(foo)", 0)
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   "Prefer `" + name + ".ok(…)` over `" + name + "(…)`.",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len(name),
	}
}

func invalid(code, output, name string, occurrence int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:            code,
		FileName:        "file.js",
		LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		Output:          []string{output},
		Errors:          []rule_tester.InvalidTestCaseError{expectedError(code, name, occurrence)},
	}
}

func TestConsistentAssertUpstream(t *testing.T) {
	typeOnlyForms := []string{
		`import type assert from "node:assert/strict";`,
		`import {type strict as assert} from "node:assert/strict";`,
		`import type {strict as assert} from "node:assert/strict";`,
	}
	validCases := []rule_tester.ValidTestCase{
		valid(`assert(foo)`),
		valid(`import assert from "assert";`),
		valid("import assert from 'node:assert';\nassert;"),
		valid("import customAssert from 'node:assert';\nassert(foo);"),
		valid("function foo (assert) {\n\tassert(bar);\n}"),
		valid("import assert from 'node:assert';\n\nfunction foo (assert) {\n\tassert(bar);\n}"),
		valid("import {strict} from 'node:assert/strict';\n\nstrict(foo);"),
		valid("import * as assert from 'node:assert';\nassert(foo);"),
		valid("export * as assert from 'node:assert';\nassert(foo);"),
		valid("export {default as assert} from 'node:assert';\nexport {assert as strict} from 'node:assert';\nassert(foo);"),
		valid("import assert from 'node:assert/strict';\nconsole.log(assert)"),
		valid("import {'strict' as assert} from 'assert';\nassert(foo)"),
	}
	for _, code := range typeOnlyForms {
		validCases = append(validCases, validTS(code), validTS(code+"\nassert();"))
	}

	allCases := "import a, {strict as b, default as c} from 'node:assert';\n" +
		"import d, {strict as e, default as f} from 'assert';\n" +
		"import g, {default as h} from 'node:assert/strict';\n" +
		"import i, {default as j} from 'assert/strict';\n" +
		"a(foo);\nb(foo);\nc(foo);\nd(foo);\ne(foo);\nf(foo);\ng(foo);\nh(foo);\ni(foo);\nj(foo);"
	allOutput := "import a, {strict as b, default as c} from 'node:assert';\n" +
		"import d, {strict as e, default as f} from 'assert';\n" +
		"import g, {default as h} from 'node:assert/strict';\n" +
		"import i, {default as j} from 'assert/strict';\n" +
		"a.ok(foo);\nb.ok(foo);\nc.ok(foo);\nd.ok(foo);\ne.ok(foo);\nf.ok(foo);\ng.ok(foo);\nh.ok(foo);\ni.ok(foo);\nj.ok(foo);"

	multiple := "import assert from 'assert';\nassert(foo)\nassert(bar)\nassert(baz)"
	multipleOutput := "import assert from 'assert';\nassert.ok(foo)\nassert.ok(bar)\nassert.ok(baz)"
	multipleErrors := []rule_tester.InvalidTestCaseError{
		expectedError(multiple, "assert", 1),
		expectedError(multiple, "assert", 2),
		expectedError(multiple, "assert", 3),
	}

	invalidCases := []rule_tester.InvalidTestCase{
		invalid("import assert from 'assert';\nassert(foo)", "import assert from 'assert';\nassert.ok(foo)", "assert", 1),
		invalid("import assert from 'node:assert';\nassert(foo)", "import assert from 'node:assert';\nassert.ok(foo)", "assert", 1),
		invalid("import assert from 'assert/strict';\nassert(foo)", "import assert from 'assert/strict';\nassert.ok(foo)", "assert", 1),
		invalid("import assert from 'node:assert/strict';\nassert(foo)", "import assert from 'node:assert/strict';\nassert.ok(foo)", "assert", 1),
		invalid("import customAssert from 'assert';\ncustomAssert(foo)", "import customAssert from 'assert';\ncustomAssert.ok(foo)", "customAssert", 1),
		invalid("import customAssert from 'node:assert';\ncustomAssert(foo)", "import customAssert from 'node:assert';\ncustomAssert.ok(foo)", "customAssert", 1),
		{
			Code:            multiple,
			FileName:        "file.js",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Output:          []string{multipleOutput},
			Errors:          multipleErrors,
		},
		invalid("import {strict} from 'assert';\nstrict(foo)", "import {strict} from 'assert';\nstrict.ok(foo)", "strict", 1),
		invalid("import {strict as assert} from 'assert';\nassert(foo)", "import {strict as assert} from 'assert';\nassert.ok(foo)", "assert", 1),
		{
			Code:            allCases,
			FileName:        "file.js",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Output:          []string{allOutput},
			Errors: []rule_tester.InvalidTestCaseError{
				expectedFooCallError(allCases, "a"),
				expectedFooCallError(allCases, "b"),
				expectedFooCallError(allCases, "c"),
				expectedFooCallError(allCases, "d"),
				expectedFooCallError(allCases, "e"),
				expectedFooCallError(allCases, "f"),
				expectedFooCallError(allCases, "g"),
				expectedFooCallError(allCases, "h"),
				expectedFooCallError(allCases, "i"),
				expectedFooCallError(allCases, "j"),
			},
		},
		invalid("import assert from 'node:assert';\nassert?.(foo)", "import assert from 'node:assert';\nassert.ok?.(foo)", "assert", 1),
		invalid(
			"import assert from 'assert';\n\n((\n\t/* comment */ ((\n\t\t/* comment */\n\t\tassert\n\t\t/* comment */\n\t\t)) /* comment */\n\t\t(/* comment */ typeof foo === 'string', 'foo must be a string' /** after comment */)\n));",
			"import assert from 'assert';\n\n((\n\t/* comment */ ((\n\t\t/* comment */\n\t\tassert.ok\n\t\t/* comment */\n\t\t)) /* comment */\n\t\t(/* comment */ typeof foo === 'string', 'foo must be a string' /** after comment */)\n));",
			"assert",
			1,
		),
	}

	if len(validCases) != 18 || len(invalidCases) != 12 {
		t.Fatalf("upstream coverage accounting changed: valid=%d invalid=%d", len(validCases), len(invalidCases))
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_assert.ConsistentAssertRule,
		validCases,
		invalidCases,
	)
}

func TestConsistentAssertDocumentation(t *testing.T) {
	cases := []struct {
		code   string
		output string
	}{
		{
			code:   "import assert from 'node:assert/strict';\n\nassert.strictEqual(actual, expected);\nassert.deepStrictEqual(actual, expected);\nassert(divide(10, 2) === 5);",
			output: "import assert from 'node:assert/strict';\n\nassert.strictEqual(actual, expected);\nassert.deepStrictEqual(actual, expected);\nassert.ok(divide(10, 2) === 5);",
		},
		{
			code:   "import assert from 'node:assert';\n\nassert.strictEqual(actual, expected);\nassert.deepStrictEqual(actual, expected);\nassert(divide(10, 2) === 5);",
			output: "import assert from 'node:assert';\n\nassert.strictEqual(actual, expected);\nassert.deepStrictEqual(actual, expected);\nassert.ok(divide(10, 2) === 5);",
		},
		{
			code:   "import {strict as assert} from 'node:assert';\n\nassert.strictEqual(actual, expected);\nassert.deepStrictEqual(actual, expected);\nassert(divide(10, 2) === 5);",
			output: "import {strict as assert} from 'node:assert';\n\nassert.strictEqual(actual, expected);\nassert.deepStrictEqual(actual, expected);\nassert.ok(divide(10, 2) === 5);",
		},
	}

	invalidCases := make([]rule_tester.InvalidTestCase, 0, len(cases))
	validCases := make([]rule_tester.ValidTestCase, 0, len(cases))
	for _, testCase := range cases {
		invalidCases = append(invalidCases, invalid(testCase.code, testCase.output, "assert", 3))
		validCases = append(validCases, valid(testCase.output))
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_assert.ConsistentAssertRule,
		validCases,
		invalidCases,
	)
}
