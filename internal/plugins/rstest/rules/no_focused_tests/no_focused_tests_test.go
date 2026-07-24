package no_focused_tests_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_focused_tests"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func focused(code, output string, line, column, endColumn int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "focusedTest",
				Line:      line,
				Column:    column,
				EndLine:   line,
				EndColumn: endColumn,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{
						MessageId: "suggestRemoveFocus",
						Output:    output,
					},
				},
			},
		},
	}
}

func focusedWithoutSuggestion(code string, line, column, endColumn int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			{
				MessageId: "focusedTest",
				Line:      line,
				Column:    column,
				EndLine:   line,
				EndColumn: endColumn,
			},
		},
	}
}

func TestNoFocusedTests(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_focused_tests.NoFocusedTestsRule,
		[]rule_tester.ValidTestCase{
			{Code: "describe()"},
			{Code: "it()"},
			{Code: "test()"},
			{Code: "describe.skip()"},
			{Code: "it.skip()"},
			{Code: "test.skip()"},
			{Code: "test.todo()"},
			{Code: "test.concurrent()"},
			{Code: "test.sequential()"},
			{Code: "test.fails()"},
			{Code: "test.each()()"},
			{Code: "test.each`table`()"},
			{Code: "test.for()()"},
			{Code: "describe.for()()"},
			{Code: "test.runIf(true)()"},
			{Code: "describe.skipIf(false)()"},

			// A factory access is not itself a test registration.
			{Code: "test.only.each(rows)"},
			{Code: "test.only.for(rows)"},
			{Code: "test.only.runIf(condition)"},
			{Code: "describe.only.skipIf(condition)"},
			{Code: "test.extend(fixtures)"},

			// Reject member orders and APIs that Rstest's types/runtime do not support.
			{Code: "test.each(rows).only('case', fn)"},
			{Code: "test.for(rows).concurrent('case', fn)"},
			{Code: "describe.each(rows).only('suite', fn)"},
			{Code: "test.only.extend(fixtures)('case', fn)"},
			{Code: "describe.fails.only('suite', fn)"},
			{Code: "describe.extend(fixtures).only('suite', fn)"},
			{Code: "test.unknown.only('case', fn)"},
			{Code: "test[modifier].only('case', fn)"},

			// Rstest has no Jest focus aliases.
			{Code: "fit()"},
			{Code: "fdescribe()"},
			{Code: "xit()"},
			{Code: "xdescribe()"},

			// Calls not resolved to @rstest/core APIs are ignored.
			{Code: "import { test } from 'node:test';\n\ntest.only()"},
			{Code: "const test = createRunner();\ntest.only('case', fn)"},

			// Modifier provenance hidden behind an alias is outside this rule's boundary.
			{Code: "const focused = test.only;\nfocused('case', fn)"},
			{Code: "const focusedEach = test.only.each(rows);\nfocusedEach('case', fn)"},
			{Code: "var appliedOnly = describe.only; appliedOnly.apply(describe)"},
			{Code: "var calledOnly = it.only; calledOnly.call(it)"},
		},
		[]rule_tester.InvalidTestCase{
			focused("describe.only()", "describe()", 1, 10, 14),
			focused("it.only()", "it()", 1, 4, 8),
			focused("test.only()", "test()", 1, 6, 10),

			// Getter order does not matter: committing any explicit only is forbidden.
			focused("test.concurrent.only()", "test.concurrent()", 1, 17, 21),
			focused("test.only.concurrent()", "test.concurrent()", 1, 6, 10),
			focused("test.only.fails()", "test.fails()", 1, 6, 10),
			focused("test.fails.only()", "test.fails()", 1, 12, 16),
			focused("describe.sequential.only()", "describe.sequential()", 1, 21, 25),
			focused("describe.only.skip()", "describe.skip()", 1, 10, 14),
			focused("test.skip.only()", "test.skip()", 1, 11, 15),

			// Conditional and parameterized factories are reported only at registration.
			focused("test.only.runIf(true)()", "test.runIf(true)()", 1, 6, 10),
			focused("test.runIf(true).only()", "test.runIf(true)()", 1, 18, 22),
			focused("test.only.skipIf(false)()", "test.skipIf(false)()", 1, 6, 10),
			focused("test.skipIf(false).only()", "test.skipIf(false)()", 1, 20, 24),
			focused("describe.only.runIf(true)()", "describe.runIf(true)()", 1, 10, 14),
			focused("describe.skipIf(false).only()", "describe.skipIf(false)()", 1, 24, 28),
			focused("test.only.for()()", "test.for()()", 1, 6, 10),
			focused("test.only.each`table`()", "test.each`table`()", 1, 6, 10),
			focused("describe.only.each()()", "describe.each()()", 1, 10, 14),
			focused("describe.only.for<Row>(rows)()", "describe.for<Row>(rows)()", 1, 10, 14),

			// Extended APIs retain the complete test API surface.
			focused("test.extend(fixtures).only()", "test.extend(fixtures)()", 1, 23, 27),
			focused(
				"test.extend(a).extend(b).concurrent.only()",
				"test.extend(a).extend(b).concurrent()",
				1, 37, 41,
			),
			focused(
				"const appTest = test.extend(fixtures);\nappTest.only()",
				"const appTest = test.extend(fixtures);\nappTest()",
				2, 9, 13,
			),
			focused(
				"const alias = test;\nconst adminTest = alias.extend(fixtures);\nadminTest.skipIf(condition).only()",
				"const alias = test;\nconst adminTest = alias.extend(fixtures);\nadminTest.skipIf(condition)()",
				3, 29, 33,
			),

			// Imports, require bindings, and namespace APIs are resolved by the parser.
			focused(
				"import { describe as describeThis } from '@rstest/core';\n\ndescribeThis.only()",
				"import { describe as describeThis } from '@rstest/core';\n\ndescribeThis()",
				3, 14, 18,
			),
			focused(
				"const { test } = require('@rstest/core');\n\ntest.only()",
				"const { test } = require('@rstest/core');\n\ntest()",
				3, 6, 10,
			),
			focused(
				"import * as rstest from '@rstest/core';\nrstest.describe.only()",
				"import * as rstest from '@rstest/core';\nrstest.describe()",
				2, 17, 21,
			),
			focused(
				"const rstest = require('@rstest/core');\nrstest.test.only()",
				"const rstest = require('@rstest/core');\nrstest.test()",
				2, 13, 17,
			),

			// Suggestions remove the complete accessor, including brackets/optionality.
			focused(`describe["only"]()`, "describe()", 1, 10, 16),
			focused(`test[  "only"  ]()`, "test()", 1, 8, 14),
			focused("test[\n  `only`\n]()", "test()", 2, 3, 9),
			focused(`test[("only")]()`, "test()", 1, 7, 13),
			focused("test?.only()", "test()", 1, 7, 11),
			focused(`test?.["only"]()`, "test()", 1, 8, 14),

			// One suggestion removes every focus marker from a registration.
			focused("test.only.only()", "test()", 1, 6, 10),
			focused("test.only.skip.only()", "test.skip()", 1, 6, 10),
			focused("describe.only.concurrent.only()", "describe.concurrent()", 1, 10, 14),

			// Reporting remains safe when deleting an accessor would delete comments.
			focusedWithoutSuggestion("test /* keep */.only()", 1, 17, 21),
			focusedWithoutSuggestion(`test[/* keep */ "only"]()`, 1, 17, 23),
		},
	)
}
