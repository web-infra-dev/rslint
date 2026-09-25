package unambiguous_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/unambiguous"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/unambiguous.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/unambiguous.md
func TestUnambiguousUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &unambiguous.UnambiguousRule,
		[]rule_tester.ValidTestCase{
			// Upstream's string cases use RuleTester's script default.
			{Code: "function x() {}", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			{Code: "\"use strict\"; function y() {}", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			{Code: "import y from \"z\"; function x() {}"},
			{Code: "import * as y from \"z\"; function x() {}"},
			{Code: "import { y } from \"z\"; function x() {}"},
			{Code: "import z, { y } from \"z\"; function x() {}"},
			{Code: "function x() {}; export {}"},
			{Code: "function x() {}; export { x }"},
			{Code: "function x() {}; export { y } from \"z\""},
			// Upstream uses Babel for this syntax; tsgo supports it natively.
			{Code: "function x() {}; export * as y from \"z\""},
			{Code: "export function x() {}"},
			// Upstream documentation.
			{Code: "import 'foo'\nfunction x() { return 42 }"},
			{Code: "export function x() { return 42 }"},
			{Code: "(function x() { return 42 })()\nexport {} // simple way to mark side-effects-only file as 'module' without any imports/exports"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:            "function x() {}",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors:          []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 16)},
			},
			// Upstream documentation.
			{
				Code:   "(function x() { return 42 })()",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 31)},
			},
		},
	)
}

func unambiguousError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "",
		Message:   "This module could be parsed as a valid script.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}
