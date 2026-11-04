// Executed in JavaScript during the audit: preserve values, effects and exceptions.
package no_redundant_comparison_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_redundant_comparison"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoRedundantComparisonEditSafety(t *testing.T) {
	valid := []rule_tester.ValidTestCase{}
	invalid := []rule_tester.InvalidTestCase{
		// Unsafe edits withheld; proven local removals retained
		{
			Code: "(() => { const a = 1, b = 1, label = \"label\"; return a === b && a === 1 && label && b === 1; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 85, EndLine: 1, EndColumn: 92,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { const a = 1, b = 1, label = \"\"; return b !== 1 && label && a === 1 && a !== b; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 49, EndLine: 1, EndColumn: 56,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let a = 1, b = 1; const guard = {get ready() { b = 2; return true; }}; return a === b && a === 1 && guard.ready && b === 1; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 125, EndLine: 1, EndColumn: 132,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let a = 1, b = 1; const guard = {valueOf() { b = 2; return 1; }}; return a === b && a === 1 && guard > 0 && b === 1; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 118, EndLine: 1, EndColumn: 125,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let a = 1, b = 1; const guard = {toString() { b = 2; return \"ok\"; }}; return a === b && a === 1 && `${guard}` && b === 1; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 123, EndLine: 1, EndColumn: 130,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let a = 1, b = 1; const key = {toString() { b = 2; return \"key\"; }}; return a === b && a === 1 && ({[key]: 1}) && b === 1; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 124, EndLine: 1, EndColumn: 131,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let a = 1, b = 1; const source = {get ready() { b = 2; return true; }}; return a === b && a === 1 && ({...source}) && b === 1; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 128, EndLine: 1, EndColumn: 135,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let a = 1, b = 1; const source = {[Symbol.iterator]: function* () { b = 2; yield 1; }}; return a === b && a === 1 && [...source] && b === 1; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 142, EndLine: 1, EndColumn: 149,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let a = 1, b = 1; const Type = {[Symbol.hasInstance]() { b = 2; return true; }}; return a === b && a === 1 && ({} instanceof Type) && b === 1; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 144, EndLine: 1, EndColumn: 151,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let a = 1, b = 1; const object = new Proxy({}, {has() { b = 2; return true; }}); return a === b && a === 1 && (\"key\" in object) && b === 1; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 141, EndLine: 1, EndColumn: 148,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { const a = 2; return b !== 1 && a === 1 && a !== b; let b; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 30, EndLine: 1, EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let reads = 0; const object = {get value() { return ++reads === 1 ? 1 : 2; }}; const b = 1; return object.value === b && b === 1 && object.value === 1; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 142, EndLine: 1, EndColumn: 160,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let reads = 0; Object.defineProperty(globalThis, \"accessed\", {get() { return ++reads === 1 ? 1 : 2; }}); const a = 1; return a === accessed && a === 1 && accessed === 1; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 164, EndLine: 1, EndColumn: 178,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let reads = 0; const env = {a: 1, get b() { return ++reads === 1 ? 1 : 2; }}; with (env) { return a === b && a === 1 && b === 1; } })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 130, EndLine: 1, EndColumn: 137,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { function check(a, b) { return a === b && a === 1 && b === 1; } return check(1, 1); })()", FileName: "case.js",
			Output: []string{"(() => { function check(a, b) { return a === b && a === 1; } return check(1, 1); })()"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 62, EndLine: 1, EndColumn: 69,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { const a = 1, b = 1, label = \"\"; if (a === b && a === 1 && label && b === 1) return true; return false; })()", FileName: "case.js",
			Output: []string{"(() => { const a = 1, b = 1, label = \"\"; if (a === b && a === 1 && label) return true; return false; })()"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 77, EndLine: 1, EndColumn: 84,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { const a = 1, b = 1, label = \"label\"; return !(a === b && a === 1 && label && b === 1); })()", FileName: "case.js",
			Output: []string{"(() => { const a = 1, b = 1, label = \"label\"; return !(a === b && a === 1 && label); })()"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 87, EndLine: 1, EndColumn: 94,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { const a = 1, b = 1, label = true; return a === b && a === 1 && label && b === 1; })()", FileName: "case.js",
			Output: []string{"(() => { const a = 1, b = 1, label = true; return a === b && a === 1 && label; })()"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 82, EndLine: 1, EndColumn: 89,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { const a = 1, b = 1, label = \"label\"; return a === b && a === 1 && b === 1 && label; })()", FileName: "case.js",
			Output: []string{"(() => { const a = 1, b = 1, label = \"label\"; return a === b && a === 1 && label; })()"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 76, EndLine: 1, EndColumn: 83,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { let a = 1, b = 1; const guard = {get ready() { b = 2; return \"ok\"; }}; return a === b && a === 1 && b === 1 && guard.ready; })()", FileName: "case.js",
			Output: []string{"(() => { let a = 1, b = 1; const guard = {get ready() { b = 2; return \"ok\"; }}; return a === b && a === 1 && guard.ready; })()"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 110, EndLine: 1, EndColumn: 117,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { const a = 1, b = 2; return b !== 1 && a === 1 && a !== b; })()", FileName: "case.js",
			Output: []string{"(() => { const a = 1, b = 2; return a === 1 && a !== b; })()"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 37, EndLine: 1, EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { const a = 1, b = 1; return a === b && a > 0 && b > 0; })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 57, EndLine: 1, EndColumn: 62,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "no-redundant-comparison/suggestion", Output: "(() => { const a = 1, b = 1; return a === b && a > 0; })()"},
					},
				},
			},
		},
		{
			Code: "(() => { function check(a, b) { return /** @type {boolean} */ (a === b && a === 1) && b === 1; } return check(1, 1); })()", FileName: "case.js",
			Output: []string{"(() => { function check(a, b) { return /** @type {boolean} */ (a === b && a === 1); } return check(1, 1); })()"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 87, EndLine: 1, EndColumn: 94,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(() => { Boolean = value => value; const a = 1, b = 1, label = \"label\"; return Boolean(a === b && a === 1 && label && b === 1); })()", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 119, EndLine: 1, EndColumn: 126,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
	}
	for i := range valid {
		valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "script"}
	}
	for i := range invalid {
		invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "script"}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_redundant_comparison.NoRedundantComparisonRule, valid, invalid)
}

// Script var bindings can be object properties; let bindings remain lexical.
func TestNoRedundantComparisonScriptBindings(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_redundant_comparison.NoRedundantComparisonRule, nil, []rule_tester.InvalidTestCase{
		{Code: "var a = 1, b = 1; a === b && a === 1 && b === 1;", FileName: "bindings.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 41, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}}},
		},
		{Code: "let a = 1, b = 1; a === b && a === 1 && b === 1;", FileName: "bindings.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Output: []string{"let a = 1, b = 1; a === b && a === 1;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 41, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}}},
		},
	})
}
