// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-invalid-character-comparison.js
// Parser-only repeats share cases; JSDoc behavior is checked in .ts files as upstream does.
package no_invalid_character_comparison_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidCharacterComparisonUpstream(t *testing.T) {
	r := sourceOnlyRule()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&r,
		[]rule_tester.ValidTestCase{
			// JSDoc / default parser.
			{Code: "/** @param {?string} text */ function run(text) { return text.charAt(0) === \"ab\"; }", FileName: "file.ts"},
			{Code: "/** @param {Array<string>} values */ function run(values) { return values.at(0) === \"ab\"; }", FileName: "file.ts"},
			{Code: "/** @param {string | Array<string>} value */ function run(value) { return value[0] === \"ab\"; }", FileName: "file.ts"},
			{Code: "/** @param {unknown} text */ function run(text) { return text.charAt(0) === \"ab\"; }", FileName: "file.ts"},
			{Code: "/** @returns {string} */ async function read() { return source; } read().at(0) === \"ab\";", FileName: "file.ts"},
			{Code: "/** @param {string} text */ function run(text) { return text?.[0] === \"ab\"; }", FileName: "file.ts"},
			// TypeScript annotation precedence.
			{Code: "/** @param {string} text */ function run(text: {charAt(): string}) { return text.charAt() === \"ab\"; }", FileName: "file.ts"},
			// Snapshot: Single-character comparison value.
			{Code: "str.charAt(4) === \"a\"", FileName: "case.js"},
			{Code: "\"a\" === str.charAt(4)", FileName: "case.js"},
			{Code: "str.charAt(4) === \"\\n\"", FileName: "case.js"},
			{Code: "str.charAt(4) !== \"\\t\"", FileName: "case.js"},
			{Code: "\"abc\".at(0) === \"a\"", FileName: "case.js"},
			{Code: "\"abc\"[0] === \"a\"", FileName: "case.js"},
			// Snapshot: Empty string (legitimate out-of-bounds check).
			{Code: "str.charAt(100) === \"\"", FileName: "case.js"},
			{Code: "str.at(0) === \"\"", FileName: "case.js"},
			{Code: "\"abc\"[100] === \"\"", FileName: "case.js"},
			// Snapshot: Not a method call.
			{Code: "charAt(4) === \"ab\"", FileName: "case.js"},
			{Code: "new charAt(4) === \"ab\"", FileName: "case.js"},
			// Snapshot: Wrong argument count.
			{Code: "str.charAt(1, 2) === \"ab\"", FileName: "case.js"},
			// Snapshot: Character access on a receiver that is not provably a string.
			{Code: "str.charAt(0) === \"ab\"", FileName: "case.js"},
			{Code: "str.charAt() === \"ab\"", FileName: "case.js"},
			{Code: "tokenizer.charAt(0) === \"ab\"", FileName: "case.js"},
			{Code: "array.at(0) === \"ab\"", FileName: "case.js"},
			{Code: "array[0] === \"ab\"", FileName: "case.js"},
			{Code: "foo.at(0) === \"ab\"", FileName: "case.js"},
			{Code: "foo[0] === \"ab\"", FileName: "case.js"},
			// Snapshot: Non-computed access or non-index key.
			{Code: "\"abc\".length === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[\"length\"] === \"ab\"", FileName: "case.js"},
			{Code: "String(foo)[index] === \"ab\"", FileName: "case.js"},
			{Code: "String(foo)[property] != \"12\"", FileName: "case.js"},
			{Code: "String(foo)[i + 1] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[``] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[`length`] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[`00`] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[`0${index}`] === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"[tag`0`] === \"ab\"", FileName: "case.js"},
			// Snapshot: Dynamic comparison value.
			{Code: "str.charAt(0) === someVariable", FileName: "case.js"},
			// Snapshot: Non-equality operator.
			{Code: "str.charAt(0) < \"ab\"", FileName: "case.js"},
			{Code: "str.charAt(0) + \"ab\"", FileName: "case.js"},
			// Snapshot: Optional chaining is intentionally not handled.
			{Code: "str?.charAt(0) === \"ab\"", FileName: "case.js"},
			{Code: "\"abc\"?.at(0) === \"ab\"", FileName: "case.js"},
			// Documentation examples.
			{Code: "'unicorn'.charAt(0) === 'u';", FileName: "case.js"},
			{Code: "'unicorn'.charAt(3) === '\\n';", FileName: "case.js"},
			{Code: "'unicorn'.at(0) === 'a';", FileName: "case.js"},
			{Code: "'unicorn'[0] === 'a';", FileName: "case.js"},
			{Code: "'unicorn'.charAt(100) === '';", FileName: "case.js"},
		}, []rule_tester.InvalidTestCase{
			// JSDoc / default parser.
			{Code: "/** @param {string} text */ function run(text) { return text.charAt(0) === \"😀\"; }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `😀`, so this comparison is always false.", Line: 1, Column: 57, EndLine: 1, EndColumn: 80, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @returns {string} */ function read() { return source; } read().at(-1) !== \"ab\";", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always true.", Line: 1, Column: 61, EndLine: 1, EndColumn: 83, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @param {() => string} read */ function run(read) { return read()[\"0\"] == \"ab\"; }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 63, EndLine: 1, EndColumn: 82, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @type {string} */ const text = source; const alias = text; alias.charAt() === \"ab\";", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 64, EndLine: 1, EndColumn: 87, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @param {string} text */ function run(text, cache) { return (cache = text)[0] === \"ab\"; }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 64, EndLine: 1, EndColumn: 90, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "/** @param {\"left\" | \"right\"} text */ function run(text) { return \"ab\" !== text.at(1); }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always true.", Line: 1, Column: 67, EndLine: 1, EndColumn: 86, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// TypeScript annotation precedence.
			{Code: "/** @param {Array<string>} text */ function run(text: string) { return text.charAt(0) === \"ab\"; }", FileName: "file.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 72, EndLine: 1, EndColumn: 95, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: All equality operators.
			{Code: "\"abc\".charAt(4) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\".charAt(4) !== \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always true.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\".charAt(0) == \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\".charAt(0) != \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always true.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: Reversed operand order.
			{Code: "\"ab\" === \"abc\".charAt(4)", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: Parenthesized operands.
			{Code: "(\"abc\".charAt(0)) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\".charAt(0) === (\"ab\")", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: Typo motivation (`/n` instead of `\n`).
			{Code: "\"abc\".charAt(3) === \"/n\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `/n`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\".charAt(822) !== \"foo\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `foo`, so this comparison is always true.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: Omitted positions default to `0`.
			{Code: "\"abc\".charAt() === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\".at() !== \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always true.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: Emoji is two UTF-16 code units (OXC misses this).
			{Code: "\"abc\".charAt(0) === \"😀\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `😀`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\".charAt(0) === \"\\u{1F600}\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `😀`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\".charAt(0) === \"\\r\\n\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `\r\n`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: Template literal comparison value.
			{Code: "\"abc\".charAt(0) === `ab`", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: `charAt` on a provable string receiver.
			{Code: "String(foo).charAt(0) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "String.fromCharCode(65).charAt(0) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "`${foo}`.charAt(0) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "(foo + \"bar\").charAt(0) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: `at` on a provable string receiver.
			{Code: "\"abc\".at(0) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\".at(0) !== \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always true.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\".at(-1) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "String(foo).at(0) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "String.fromCharCode(65).at(0) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "`${foo}`.at(0) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "(foo + \"bar\").at(0) === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: `[index]` on a provable string receiver.
			{Code: "\"abc\"[0] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\"[0] != \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always true.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\"[\"0\"] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: TypeScript: receiver proven via type annotation.
			{Code: "declare const s: string;\nconst x = s.at(0) === \"ab\";", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 2, Column: 11, EndLine: 2, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "declare const s: string;\nconst x = s.charAt(0) === \"ab\";", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 2, Column: 11, EndLine: 2, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function isAb(s: string) {\n\treturn s[\"0\"] === \"ab\";\n}", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 2, Column: 9, EndLine: 2, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: TypeScript: receiver proven via `as` assertion.
			{Code: "(foo as string)[0] === \"ab\";", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: A single UTF-16 code unit can never equal an astral character, which is two code units.
			{Code: "'😀'[0] === '😀'", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `😀`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "'😀'.charAt(0) === '😀'", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `😀`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "'a😀'.at(0) === 'a😀'", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `a😀`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Snapshot: `String#at()` returns a UTF-16 code unit too, not a whole code point.
			{Code: "'😀'.at(0) === '😀'", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `😀`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\"[`0`] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\"[(`1`)] === `ab`", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "\"abc\"[`\\u{30}`] === \"ab\"", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function isAb(value: string) { return value[`0`] === `ab`; }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 39, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Documentation examples.
			{Code: "'unicorn'.charAt(0) === 'ab';", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "'unicorn'.charAt(3) === '/n'; // Typo: meant '\\n'", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `/n`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "'unicorn'.at(0) === 'ab';", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "'unicorn'[0] === 'ab';", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-invalid-character-comparison", Message: "A single character can never equal the multi-character string `ab`, so this comparison is always false.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
}
