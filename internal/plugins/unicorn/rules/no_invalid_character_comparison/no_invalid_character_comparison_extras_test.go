// Additional AST and numeric/string coverage checked against Unicorn v79.0.0.
package no_invalid_character_comparison_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_invalid_character_comparison"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidCharacterComparisonExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&no_invalid_character_comparison.NoInvalidCharacterComparisonRule,
		[]rule_tester.ValidTestCase{
			// Call and optional-chain boundaries.
			{Code: "\"abc\".charAt(0, 1) === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\".at(0, 1) === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\".charAt(...args) === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\".at(...args) === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"charAt\"](0) === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\".codePointAt(0) === \"ab\"", FileName: "case.js"},
			{Code: "new (\"abc\".charAt)(0) === \"ab\"", FileName: "case.js"},
			{Code: "class Text { #charAt() { return \"ab\"; } test() { return this.#charAt() === \"ab\"; } }", FileName: "case.js"},
			{Code: "\"abc\".charAt?.(0) === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"?.[0] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"?.trim().charAt(0) === \"ab\"", FileName: "case.js"},
			{Code: "(\"abc\"?.charAt)(0) === \"ab\"", FileName: "case.js"},
			// Static numeric keys.
			{Code: "\"abc\"[1n] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[-1] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[+1] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"-0\"] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"-1\"] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"1.5\"] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"1.0\"] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"1e0\"] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"0x1\"] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\" 1\"] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"NaN\"] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"Infinity\"] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"9007199254740992\"] === \"ab\"", FileName: "case.js"},
			{Code: "const index = 0; \"abc\"[index] === \"ab\"", FileName: "case.js"},
			// Literal and language wrappers.
			{Code: "\"abc\"[0] === \"a\" + \"b\"", FileName: "case.js"},
			{Code: "\"abc\"[0] === `a${\"b\"}`", FileName: "case.js"},
			{Code: "const target = \"ab\"; \"abc\"[0] === target", FileName: "case.js"},
			{Code: "\"abc\"[0] === \"\\uD800\"", FileName: "case.js"},
			{Code: "(\"abc\"[0] as string) === \"ab\"", FileName: "case.ts"},
			{Code: "\"abc\"[0] === (\"ab\" as const)", FileName: "case.ts"},
			{Code: "\"abc\"[0]! === \"ab\"", FileName: "case.ts"},
			{Code: "\"abc\"[0 satisfies number] === \"ab\"", FileName: "case.ts"},
			// Receiver proof.
			{Code: "[\"ab\"].at(0) === \"ab\"", FileName: "case.js"},
			{Code: "new String(\"abc\").charAt(0) === \"ab\"", FileName: "case.js"},
			{Code: "declare const text: string | undefined; text[0] === \"ab\"", FileName: "case.ts"},
			{Code: "let text = \"abc\"; text = [\"ab\"]; text.at(0) === \"ab\"", FileName: "case.js"},
			{Code: "declare const read: () => string; read().charAt(0) === \"ab\";", FileName: "case.ts"},
			// Additional optional-chain, TypeScript and numeric edges.
			{Code: "(\"abc\"?.trim()).charAt(0) === \"ab\"", FileName: "case.js"},
			{Code: "(\"abc\"?.trim())[0] === \"ab\"", FileName: "case.js"},
			{Code: "(\"abc\"?.at(0))[0] === \"ab\"", FileName: "case.js"},
			{Code: "(\"abc\".charAt<string>)(0) === \"ab\"", FileName: "case.ts"},
			{Code: "\"abc\".charAt!(0) === \"ab\"", FileName: "case.ts"},
			{Code: "\"abc\"[0 as number] === \"ab\"", FileName: "case.ts"},
			{Code: "\"abc\"[0!] === \"ab\"", FileName: "case.ts"},
			{Code: "type Text<T extends string> = T; declare const text: Text<string>; text[0] === \"ab\"", FileName: "case.ts"},
			{Code: "\"abc\"[0] === String.raw`ab`", FileName: "case.js"},
			{Code: "\"abc\"[-0] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"+0\"] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"0.0\"] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"0\\uFEFF\"] === \"ab\"", FileName: "case.js"},
			// A local String factory returns an array; Unicorn v79.0.0 reports this comparison.
			{Code: "function String() { return [\"ab\"]; } String().at(0) === \"ab\"", FileName: "case.js"},
		}, []rule_tester.InvalidTestCase{
			// Call and optional-chain boundaries.
			{Code: "(\"abc\".charAt)(0) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Static numeric keys.
			{Code: "\"abc\"[\"9007199254740991\"] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\"[((0x1))] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\"[0.5] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\"[1e100] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Literal and language wrappers.
			{Code: "\"abc\"[0] === \"e\\u0301\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `é`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @type {string} */ (\"abc\"[0]) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 23, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\"[0] === /** @type {string} */ (\"ab\")", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const view = <div>{\"abc\"[0] === \"ab\"}</div>;", FileName: "case.tsx", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 20, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const emoji = \"😀\";\nif (\n  (\"abc\"[0])\n  !== `a\nb`\n) {}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `a\nb`, so this comparison is always true.", Line: 3, Column: 3, EndLine: 5, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const emoji = \"😀\"; \"abc\"[0] === \"ab\";", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 21, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"ab\" === (\"abc\"[0] === \"cd\")", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `cd`, so this comparison is always false.", Line: 1, Column: 11, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Receiver proof.
			{Code: "const text = \"abc\"; const alias = text; alias[0] !== \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always true.", Line: 1, Column: 41, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function check(text: string) { return text.at(0) == \"12\"; }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `12`, so this comparison is always false.", Line: 1, Column: 39, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "declare const read: () => string; (read() as string).charAt(0) === \"ab\";", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 35, EndLine: 1, EndColumn: 72, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "(test ? \"abc\" : \"def\")[0] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Additional optional-chain, TypeScript and numeric edges.
			{Code: "(\"abc\"?.[0]).charAt(0) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\".charAt<string>(0) === \"ab\"", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "(<string>source)[0] === \"ab\"", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "type Text = string; declare const text: Text; text.at(0) === \"ab\"", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 47, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function check<T extends string>(text: T) { return text[0] === \"ab\"; }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 52, EndLine: 1, EndColumn: 68, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "String.raw`abc`[0] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/abc/.source[0] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\"[1e999] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
}
