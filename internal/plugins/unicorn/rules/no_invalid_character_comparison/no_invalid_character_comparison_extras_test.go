// Additional AST and numeric/string coverage checked against Unicorn v79.0.0.
package no_invalid_character_comparison_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_invalid_character_comparison"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// The pinned source-only cases do not enable upstream parser type services.
func sourceOnlyRule() rule.Rule {
	r := no_invalid_character_comparison.NoInvalidCharacterComparisonRule
	run := r.Run
	r.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return run(ctx, options)
	}
	return r
}

func TestNoInvalidCharacterComparisonExtras(t *testing.T) {
	r := sourceOnlyRule()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&r,
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

func TestNoInvalidCharacterComparisonProjectTypes(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `declare const read: () => string | number; read().charAt(0) === "ab";`},
		{Code: `declare const text: {value: unknown}; text.value.at(0) === "ab";`},
		{Code: `/** @type {unknown} */ const text = "abc"; text[0] === "ab";`},
		{Code: `let text = "abc"; text = replacement; text[0] === "ab";`},
		{Code: `function read(): string { return "abc"; } read = replacement; read().charAt(0) === "ab";`},
		{Code: `let text = "abc"; text = replacement; const alias = text; alias[0] === "ab";`},
		{Code: `function read(): string { return "abc"; } read = replacement; (flag ? read() : "abc")[0] === "ab";`},
		// Ambient constructor types must not override local script bindings.
		{Code: `function String() { return ["ab"]; } String().at(0) === "ab";`, FileName: "case.js"},
		{Code: `function Date() { return ["ab"]; } Date().at(0) === "ab";`, FileName: "case.js"},
		{Code: `const String = {fromCharCode() { return ["ab"]; }}; String.fromCharCode(97).at(0) === "ab";`, FileName: "case.js"},
		{Code: `function String() { return ["ab"]; } const read = String; read().at(0) === "ab";`, FileName: "case.js"},
		{Code: `export {}; function String() { return ["ab"]; } String().at(0) === "ab";`, FileName: "case.js"},
		{Code: `/** @param {unknown} s */ function f(s) { s[0] === "ab"; }`, FileName: "case.js"},
		{Code: `/** @param {string | null} text */ function f(text) { text[0] === "ab"; }`, FileName: "case.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: `declare const read: () => string; read().charAt(0) === "ab";`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 35, EndLine: 1, EndColumn: 60}}},
		{Code: `declare const text: {value: string}; text.value.at(0) !== "ab";`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always true.", Line: 1, Column: 38, EndLine: 1, EndColumn: 63}}},
		{Code: `interface Read {(): string}; declare const read: Read; read()[0] === "ab";`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 56, EndLine: 1, EndColumn: 74}}},
		{Code: `declare const read: () => string; const text = read(); text[0] === "ab";`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 56, EndLine: 1, EndColumn: 72}}},
		{Code: `export {}; type CharacterText<T extends string> = T; declare const text: CharacterText<string>; text.at(0) === "ab";`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 97, EndLine: 1, EndColumn: 116}}},
		{Code: `export {}; function check(text?: string) { if (text) { text[0] === "ab"; } }`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 56, EndLine: 1, EndColumn: 72}}},
		{Code: `export {}; function check(text: unknown) { if (typeof text === "string") { text[0] === "ab"; } }`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 76, EndLine: 1, EndColumn: 92}}},
		{Code: `/** @param {unknown} s */ function f(s) { if (typeof s === "string") { s[0] === "ab"; } }`, FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 72, EndLine: 1, EndColumn: 85}}},
		{Code: `/** @param {string | null} text */ function f(text) { if (text) { text[0] === "ab"; } }`, FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 67, EndLine: 1, EndColumn: 83}}},
		{Code: `/** @param {unknown} s */ function f(s) { if (typeof s === "string") { const alias = s; alias[0] === "ab"; } }`, FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 89, EndLine: 1, EndColumn: 106}}},
		{Code: `export {}; let s: unknown = "x"; if (typeof s === "string") { s[0] === "ab"; } s = 0;`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 63, EndLine: 1, EndColumn: 76}}},
		{Code: `export {}; let s: unknown = "x"; s = read(); if (typeof s === "string") { s[0] === "ab"; }`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 75, EndLine: 1, EndColumn: 88}}},
		{Code: `export {}; let s: unknown = "x"; if (typeof s === "string") { const alias = s; alias[0] === "ab"; } s = 0;`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 80, EndLine: 1, EndColumn: 97}}},
		{Code: `export interface Read {(): string}; export declare const read: Read; const value = read(); value[0] === "ab";`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 92, EndLine: 1, EndColumn: 109}}},
		{Code: `/** @param {{value: string}} text */ function f(text) { text.value.at(0) === "ab"; }`, FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 57, EndLine: 1, EndColumn: 82}}},
		{Code: `/** @param {any} s */ function f(s) { if (typeof s === "string") { s[0] === "ab"; } }`, FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 68, EndLine: 1, EndColumn: 81}}},
		{Code: `export {}; let s: unknown = read(); if (typeof s === "string") { const alias = s; s = 0; alias[0] === "ab"; }`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 90, EndLine: 1, EndColumn: 107}}},
	}
	for _, depth := range []int{10, 11, 20} {
		code := []byte("declare const flag: boolean; declare const read: () => string;\nconst s0 = read();\n")
		for level := 1; level <= depth; level++ {
			code = fmt.Appendf(code, "const s%d = flag ? s%d : \"p\" + s%d;\n", level, level-1, level-1)
		}
		code = fmt.Appendf(code, "s%d[0] === \"ab\";", depth)
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: string(code), Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: depth + 3, Column: 1, EndLine: depth + 3, EndColumn: 16}}})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&no_invalid_character_comparison.NoInvalidCharacterComparisonRule, valid, invalid)
}

func TestNoInvalidCharacterComparisonLoneSurrogates(t *testing.T) {
	r := sourceOnlyRule()
	value := ecmascript.StringFromCodeUnits([]uint16{0xd800, 'a'})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &r, nil,
		[]rule_tester.InvalidTestCase{{Code: `"x"[0] === "\uD800a";`, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `" + value + "`, so this comparison is always false.",
			Line: 1, Column: 1, EndLine: 1, EndColumn: 21,
		}}}})
}
