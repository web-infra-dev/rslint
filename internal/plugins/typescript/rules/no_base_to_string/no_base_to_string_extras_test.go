package no_base_to_string

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// These regressions pin messages and UTF-16 ranges against typescript-eslint
// v8.28.0, including native AST parentheses and script bindings merged with libs.
func TestNoBaseToStringRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoBaseToStringRule,
		[]rule_tester.ValidTestCase{
			{Code: "function String(value: unknown) { return value; } String({});"},
			{Code: "var String = (value: unknown) => value; String({});"},
			{Code: "export {}; function String(value: unknown) { return value; } String({});"},
			{Code: "function convert(String: (value: unknown) => unknown) { String({}); }"},
			{Code: "{ const String = (value: unknown) => value; String({}); }"},
			{Code: "import { String } from 'foo'; String({});"},
			{Code: "String(42); String(...[{}, {}]);"},
			{Code: "({})['toString'](); [{}]['join']();"},
		},
		[]rule_tester.InvalidTestCase{
			// The astral character occupies two UTF-16 columns before the report.
			{Code: "String(/* 😀 */ {});", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseToString", Message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.", Line: 1, Column: 17, EndLine: 1, EndColumn: 19}}},
			{Code: "String({});", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseToString", Message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.", Line: 1, Column: 8, EndLine: 1, EndColumn: 10}}},
			{Code: "String((( /* leading */ {} )));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseToString", Message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.", Line: 1, Column: 25, EndLine: 1, EndColumn: 27}}},
			{Code: "(( /* leading */ {} )).toLocaleString();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseToString", Message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.", Line: 1, Column: 18, EndLine: 1, EndColumn: 20}}},
			{Code: "'' + ((( /* leading */ {} )));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseToString", Message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.", Line: 1, Column: 24, EndLine: 1, EndColumn: 26}}},
			{Code: "(( /* leading */ [{}] )).join();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseArrayJoin", Message: "Using `join()` for [{}] will use Object's default stringification format ('[object Object]') when stringified.", Line: 1, Column: 18, EndLine: 1, EndColumn: 22}}},
			{Code: "(String)({});", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseToString", Message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.", Line: 1, Column: 10, EndLine: 1, EndColumn: 12}}},
			{Code: "(({}).toString)();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseToString", Message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.", Line: 1, Column: 3, EndLine: 1, EndColumn: 5}}},
			{Code: "({})[toString]();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseToString", Message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.", Line: 1, Column: 2, EndLine: 1, EndColumn: 4}}},
			{Code: "[{}][join]();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseArrayJoin", Message: "Using `join()` for [{}] will use Object's default stringification format ('[object Object]') when stringified.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5}}},
			{Code: "const 類型 = {};\nString(類型);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseToString", Message: "'類型' will use Object's default stringification format ('[object Object]') when stringified.", Line: 2, Column: 8, EndLine: 2, EndColumn: 10}}},
			{Code: "String(\n /* before object */ {\n  value: 1,\n }\n);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseToString", Message: "'{\n  value: 1,\n }' will use Object's default stringification format ('[object Object]') when stringified.", Line: 2, Column: 22, EndLine: 4, EndColumn: 3}}},
			{Code: "declare const value: {} | string; String(value);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "baseToString", Message: "'value' may use Object's default stringification format ('[object Object]') when stringified.", Line: 1, Column: 42, EndLine: 1, EndColumn: 47}}},
		},
	)
}

func TestNoBaseToStringOptionsSchema(t *testing.T) {
	// checkUnknown belongs to a later upstream version and is not implemented
	// by this v8.28.0 rule. Reject it instead of silently ignoring its value.
	if err := NoBaseToStringRule.Schema.Validate([]any{map[string]any{"checkUnknown": true}}); err == nil {
		t.Fatal("checkUnknown must be rejected by the v8.28.0 schema")
	}
	if err := NoBaseToStringRule.Schema.Validate([]any{map[string]any{"ignoredTypeNames": []any{"Custom"}}}); err != nil {
		t.Fatalf("ignoredTypeNames must be accepted: %v", err)
	}
}
