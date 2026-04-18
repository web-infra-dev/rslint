// cspell:ignore aeiou DFFB DDEF dedup
package no_misleading_character_class

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMisleadingCharacterClassRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoMisleadingCharacterClassRule,
		[]rule_tester.ValidTestCase{
			// ---- Baseline: u/v flag or non-class contexts ----
			{Code: `var r = /[👍]/u`},
			{Code: `var r = /[\uD83D\uDC4D]/u`},
			{Code: `var r = /[\u{1F44D}]/u`},
			{Code: `var r = /❇️/`},
			{Code: `var r = /Á/`},
			{Code: `var r = /[❇]/`},
			{Code: `var r = /👶🏻/`},
			{Code: `var r = /[👶]/u`},
			{Code: `var r = /🇯🇵/`},
			{Code: `var r = /[JP]/`},
			{Code: `var r = /👨‍👩‍👦/`},
			{Code: `new RegExp()`},
			{Code: `var r = RegExp(/[👍]/u)`},
			{Code: `const regex = /[👍]/u; new RegExp(regex);`},

			// ---- Solo code points are OK ----
			{Code: `var r = /[\uD83D]/`},
			{Code: `var r = /[\uDC4D]/`},
			{Code: `var r = /[\uD83D]/u`},
			{Code: `var r = /[\uDC4D]/u`},
			{Code: `var r = /[\u0301]/`},
			{Code: `var r = /[\uFE0F]/`},
			{Code: `var r = /[\u0301]/u`},
			{Code: `var r = /[\uFE0F]/u`},
			{Code: `var r = /[\u{1F3FB}]/u`},
			{Code: `var r = /[🇯]/u`},
			{Code: `var r = /[🇵]/u`},
			{Code: `var r = /[\u200D]/`},
			{Code: `var r = /[\u200D]/u`},

			// ---- Non-RegExp / non-literal call paths ----
			{Code: `new RegExp('[Á] [ ');`},           // syntax error in pattern → ignored
			{Code: `var r = new RegExp('[Á] [ ');`},  // ditto
			{Code: `var r = RegExp('{ [Á]', 'u');`}, // ditto
			{Code: `var r = RegExp(` + "`" + `${x}[👍]` + "`" + `)`},
			{Code: `var r = new RegExp('[🇯🇵]', ` + "`${foo}`" + `)`},
			{Code: `var r = new RegExp("[👍]", flags)`},
			{Code: `const args = ['[👍]', 'i']; new RegExp(...args);`},

			// ---- ES2024 v flag ----
			{Code: `var r = /[👍]/v`},
			{Code: `var r = /^[\q{👶🏻}]$/v`},
			{Code: `var r = /[🇯\q{abc}🇵]/v`},
			{Code: `var r = /[🇯[A]🇵]/v`},
			{Code: `var r = /[🇯[A--B]🇵]/v`},

			// ---- allowEscape ----
			{Code: `/[\ud83d\udc4d]/`, Options: map[string]interface{}{"allowEscape": true}},
			{Code: `/[A\u0301]/`, Options: map[string]interface{}{"allowEscape": true}},
			{Code: `/[👶\u{1f3fb}]/u`, Options: map[string]interface{}{"allowEscape": true}},
			{Code: `/[\u{1F1EF}\u{1F1F5}]/u`, Options: map[string]interface{}{"allowEscape": true}},
			{Code: `/[\u00B7\u0300-\u036F]/u`, Options: map[string]interface{}{"allowEscape": true}},
			{Code: `/[\n\u0305]/`, Options: map[string]interface{}{"allowEscape": true}},
			{Code: `RegExp("[\uD83D\uDC4D]")`, Options: map[string]interface{}{"allowEscape": true}},
			{Code: `RegExp("[A\u0301]")`, Options: map[string]interface{}{"allowEscape": true}},

			// ---- Identifier resolved to a u-flagged regex literal — OK ----
			{Code: `const regex = /[👍]/u; new RegExp(regex);`},
			// ---- Identifier resolved to a safe string literal ----
			{Code: `const pattern = "[abc]"; new RegExp(pattern);`},
			// ---- let binding with a write reference is NOT resolved ----
			// (Any reassignment makes the initial value unreliable; mirrors
			// ESLint's `getStaticValue` which requires no write references.)
			{Code: `let pattern = "[abc]"; pattern = "[👍]"; new RegExp(pattern);`},
			// ---- let binding with no writes + safe initial value ----
			{Code: `let pattern = "[abc]"; new RegExp(pattern);`},
			// ---- let with emoji pair + u flag override — pattern is safe under u ----
			{Code: `let pattern = "[👍]"; new RegExp(pattern, "u");`},

			// ---- Breaker (\d) splits a would-be emoji modifier pair ----
			{Code: `var r = /[👶\d🏻]/u`},
			// ---- Breaker (\s) splits a would-be surrogate pair without u flag ----
			{Code: `var r = /[\uD83D\s\uDC4D]/`},
			// ---- Breaker (\p{Letter}) splits a would-be combining sequence ----
			{Code: `var r = /[A\p{Letter}\u0301]/u`},
			// ---- Solo surrogate is fine outside a pair ----
			{Code: `var r = /[\uD83D abc]/`},
			// ---- Range endpoints don't trigger detectors when splitting ----
			{Code: `var r = /[a-Á]/`},
			// ---- v-flag \q{...} containing what looks like a pair ----
			{Code: `var r = /[\q{👶🏻}]/v`},
			// ---- Solo emoji modifier (no base) inside a class ----
			{Code: `var r = /[\u{1F3FB}]/u`},

			// ---- RegExp(regexLiteralViaIdentifier, overrideFlags) must NOT
			// re-analyze the regex under override flags. Matches ESLint's
			// `getStaticValueOrRegex` which returns null for RegExp objects,
			// so flag-stripping patterns (`/[👍]/u` safe → `new RegExp(r, "")`
			// technically misleading at runtime) are intentionally not flagged
			// through the constructor. The standalone literal listener still
			// fires on misleading literals themselves.
			{Code: `const r = /[👍]/u; new RegExp(r, "");`},

			// ---- ESLint-aligned "don't resolve" cases (all valid) ----
			// ESLint's getStaticValue does NOT resolve:
			//   - method calls like `.repeat(n)`
			//   - conditional expressions
			//   - object / array destructuring bindings
			// We match this behavior.
			{Code: `new RegExp("[👍]".repeat(1));`},
			{Code: `new RegExp(cond ? "[👍]" : "a");`},
			{Code: `const {pattern} = {pattern: "[👍]"}; new RegExp(pattern);`},
			{Code: `const [pattern] = ["[👍]"]; new RegExp(pattern);`},

			// ---- Breaker `\p{...}` splits a would-be emoji modifier pair ----
			{Code: `var r = /[👶\p{Letter}🏻]/u`},
			// ---- Deep v-flag nesting with no misleading sequence ----
			{Code: `var r = /[[[[[a]]]]]/v`},
			// ---- v-flag set operations without misleading content ----
			{Code: `var r = /[[a-z]--[aeiou]]/v`},
			{Code: `var r = /[[a-z]&&[0-9]]/v`},
			// ---- Range endpoints with combining mark — split by range ----
			// `[A-Z\u0301]`: [A-Z] is range, then \u0301 on its own — no pair.
			// Actually `\u0301` is combining, and the range `max` is `Z` which
			// IS non-combining. So sequence [Z, \u0301] fires combiningClass.
			// Instead test a case where the sequence is safely split.
			{Code: `var r = /[\u0300-\u036F]/u`},
			// ---- ZWJ as solo character in class (not flanked by non-ZWJ on both sides) ----
			{Code: `var r = /[\u200D]/u`},
			{Code: `var r = /[\u200D\u200D]/u`},
			// ---- Misleading sequence outside any class ----
			// The rule only inspects character classes, so literal 👍 outside
			// any `[...]` is not flagged regardless of flags.
			{Code: `var r = /👶🏻/u`},
			{Code: `var r = /🇯🇵/u`},
			{Code: `var r = /👨‍👩‍👦/u`},
		},

		[]rule_tester.InvalidTestCase{
			// ---- Regex literals: surrogatePairWithoutUFlag ----
			{
				Code: `var r = /[👍]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[👍]/u`},
						},
					},
				},
			},
			{
				Code: `var r = /[\uD83D\uDC4D]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[\uD83D\uDC4D]/u`},
						},
					},
				},
			},
			{
				Code: `var r = /before[\uD83D\uDC4D]after/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /before[\uD83D\uDC4D]after/u`},
						},
					},
				},
			},
			{
				Code: `var r = /[before\uD83D\uDC4Dafter]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[before\uD83D\uDC4Dafter]/u`},
						},
					},
				},
			},
			{
				Code: `var r = /\uDC4D[\uD83D\uDC4D]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /\uDC4D[\uD83D\uDC4D]/u`},
						},
					},
				},
			},

			// ---- Regex literals: combiningClass ----
			// Using A + combining acute accent (U+0301) explicitly.
			{
				Code: "var r = /[A\u0301]/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 11},
				},
			},
			{
				Code: "var r = /[A\u0301]/u",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[\u0041\u0301]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[\u0041\u0301]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[\u{41}\u{301}]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 11},
				},
			},
			{
				Code: "var r = /[\u2747\uFE0F]/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 11},
				},
			},
			{
				Code: "var r = /[\u2747\uFE0F]/u",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[\u2747\uFE0F]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[\u{2747}\u{FE0F}]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 11},
				},
			},

			// ---- emojiModifier ----
			{
				Code: `var r = /[👶🏻]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[a\uD83C\uDFFB]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[\uD83D\uDC76\uD83C\uDFFB]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[\u{1F476}\u{1F3FB}]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier", Line: 1, Column: 11},
				},
			},

			// ---- regionalIndicatorSymbol ----
			{
				Code: `var r = /[🇯🇵]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regionalIndicatorSymbol", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[\uD83C\uDDEF\uD83C\uDDF5]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regionalIndicatorSymbol", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[\u{1F1EF}\u{1F1F5}]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regionalIndicatorSymbol", Line: 1, Column: 11},
				},
			},

			// ---- zwj ----
			{
				Code: `var r = /[👨‍👩‍👦]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "zwj", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[👩‍👦]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "zwj", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[\uD83D\uDC68\u200D\uD83D\uDC69\u200D\uD83D\uDC66]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "zwj", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /[\u{1F468}\u{200D}\u{1F469}\u{200D}\u{1F466}]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "zwj", Line: 1, Column: 11},
				},
			},

			// ---- Under u/v flag: surrogatePair (via \u{...}) ----
			{
				Code: `/[\ud83d\u{dc4d}]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePair"},
				},
			},
			{
				Code: `/[\u{d83d}\udc4d]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePair"},
				},
			},
			{
				Code: `/[\u{d83d}\u{dc4d}]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePair"},
				},
			},
			{
				Code: `/[\uD83D\u{DC4d}]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePair"},
				},
			},

			// ---- Multiple matches in one regex ----
			{
				Code: `var r = /[👶🏻]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[👶🏻]/u`},
						},
					},
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[👶🏻]/u`},
						},
					},
				},
			},
			{
				Code: `var r = /[🇯🇵]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[🇯🇵]/u`},
						},
					},
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[🇯🇵]/u`},
						},
					},
				},
			},
			{
				Code: `var r = /[🇯🇵]/i`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[🇯🇵]/iu`},
						},
					},
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[🇯🇵]/iu`},
						},
					},
				},
			},

			// ---- zwj across multiple classes ----
			{
				Code: `var r = /[👩‍👦][👩‍👦]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "zwj", Line: 1, Column: 11},
					{MessageId: "zwj", Line: 1, Column: 18},
				},
			},
			{
				Code: `var r = /[👨‍👩‍👦]foo[👨‍👩‍👦]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "zwj", Line: 1, Column: 11},
					{MessageId: "zwj", Line: 1, Column: 24},
				},
			},

			// ---- Adjacency within one class ----
			{
				Code: `var r = /[👨‍👩‍👦👩‍👦]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "zwj", Line: 1, Column: 11},
					{MessageId: "zwj", Line: 1, Column: 19},
				},
			},

			// ---- RegExp constructor: string literal pattern ----
			{
				Code: `var r = RegExp("[👍]", "")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = RegExp("[👍]", "u")`},
						},
					},
				},
			},
			{
				Code: `var r = new RegExp("[👍]", "")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = new RegExp("[👍]", "u")`},
						},
					},
				},
			},
			{
				Code: "var r = new RegExp(\"[A\u0301]\", \"\")",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 22},
				},
			},
			{
				Code: "var r = new RegExp(\"[A\u0301]\", \"u\")",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 22},
				},
			},
			{
				Code: `var r = new RegExp("[\u0041\u0301]", "")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 22},
				},
			},
			{
				Code: `var r = new RegExp("[👶🏻]", "u")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier", Line: 1, Column: 22},
				},
			},
			{
				Code: `var r = new RegExp("[🇯🇵]", "u")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regionalIndicatorSymbol", Line: 1, Column: 22},
				},
			},

			// ---- new globalThis.RegExp(...) ----
			{
				Code: "var r = new globalThis.RegExp(\"[\u2747\uFE0F]\", \"\")",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 33},
				},
			},

			// ---- allowEscape: still flags non-escaped combinations ----
			{
				Code:    "/[A\u0301]/",
				Options: map[string]interface{}{"allowEscape": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass"},
				},
			},

			// ==== Extended ESLint parity: regex literal boundary cases ====

			// ---- Suggestions are null when pattern is invalid with u flag ----
			// Pattern `[👍]\a` — `\a` identity escape would be invalid under u.
			{
				Code: `var r = /[👍]\a/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 11},
				},
			},
			{
				Code: `var r = /\a[👍]\a/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 13},
				},
			},
			// ---- Lookbehind with a misleading class ----
			{
				Code: `var r = /(?<=[👍])/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /(?<=[👍])/u`},
						},
					},
				},
			},
			// ---- Pattern that would be invalid with u flag → no suggestion ----
			{
				Code: `var r = /[👍]\a/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 11},
				},
			},
			// ---- Multiple surrogate pairs in one class ----
			{
				Code: `var r = /[👶🏻👶🏻]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[👶🏻👶🏻]/u`},
						},
					},
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[👶🏻👶🏻]/u`},
						},
					},
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[👶🏻👶🏻]/u`},
						},
					},
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /[👶🏻👶🏻]/u`},
						},
					},
				},
			},
			// ---- Mixed hex + u-brace for surrogate pair under u flag ----
			{
				Code: `/[\u{d83d}\udc4d]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePair"},
				},
			},
			{
				Code: `/[\u{d83d}\u{dc4d}]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePair"},
				},
			},

			// ==== Extended parity: RegExp constructor positions ====

			// ---- Surrogate pair via \\uHHHH in a string literal ----
			{
				Code: `var r = RegExp("[\\uD83D\\uDC4D]", "")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = RegExp("[\\uD83D\\uDC4D]", "u")`},
						},
					},
				},
			},
			{
				Code: `var r = RegExp("before[\\uD83D\\uDC4D]after", "")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = RegExp("before[\\uD83D\\uDC4D]after", "u")`},
						},
					},
				},
			},
			{
				Code: `var r = RegExp("[before\\uD83D\\uDC4Dafter]", "")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = RegExp("[before\\uD83D\\uDC4Dafter]", "u")`},
						},
					},
				},
			},
			// ---- Combining class via \\uHHHH\\uHHHH in a string literal ----
			{
				Code: `var r = new RegExp("[\\u0041\\u0301]", "")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 22},
				},
			},
			{
				Code: `var r = new RegExp("[\\u0041\\u0301]", "u")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 22},
				},
			},
			{
				Code: `var r = new RegExp("[\\u{41}\\u{301}]", "u")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 1, Column: 22},
				},
			},
			// ---- Template literal (no substitution) argument ----
			{
				Code: "var r = RegExp(`[\u2747\uFE0F]`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass"},
				},
			},
			{
				Code: "var r = new RegExp('[👍]', ``)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: "var r = new RegExp('[👍]', `u`)"},
						},
					},
				},
			},
			// ---- Multi-line template literal: source line tracking ----
			{
				Code: "var r = new RegExp(`\n                [👍]`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 2, Column: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: "var r = new RegExp(`\n                [👍]`, \"u\")"},
						},
					},
				},
			},
			{
				Code: "var r = new RegExp(`\n                [\u2747\uFE0F]`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass", Line: 2, Column: 18},
				},
			},
			// ---- Parenthesized args ----
			{
				Code: `var r = new RegExp(("[🇯🇵]"))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = new RegExp(("[🇯🇵]"), "u")`},
						},
					},
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = new RegExp(("[🇯🇵]"), "u")`},
						},
					},
				},
			},
			{
				Code: `var r = new RegExp((("[🇯🇵]")))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = new RegExp((("[🇯🇵]")), "u")`},
						},
					},
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = new RegExp((("[🇯🇵]")), "u")`},
						},
					},
				},
			},
			// ---- globalThis.RegExp variants ----
			{
				Code: `var r = new globalThis.RegExp("[👶🏻]", "u")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier", Line: 1, Column: 33},
				},
			},
			{
				Code: `var r = new globalThis.RegExp("[🇯🇵]", "")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = new globalThis.RegExp("[🇯🇵]", "u")`},
						},
					},
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 35,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = new globalThis.RegExp("[🇯🇵]", "u")`},
						},
					},
				},
			},
			// ---- Report only on regex literal when no flags argument ----
			{
				Code: "RegExp(/[👍]/)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: "RegExp(/[👍]/u)"},
						},
					},
				},
			},
			// ---- Regex literal + flags arg: flags override ----
			{
				Code: "RegExp(/[👍]/, 'i');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: "RegExp(/[👍]/, 'iu');"},
						},
					},
				},
			},
			// ---- Regex literal with u flag in first arg, no flags in second: skip ----
			// (covered by valid test above: `var r = RegExp(/[👍]/u)`)

			// ==== ES2024 v-flag ====

			{
				Code: `var r = /[[👶🏻]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier", Line: 1, Column: 12},
				},
			},
			{
				Code: "new RegExp(/^[👍]$/v, '')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: "new RegExp(/^[👍]$/v, 'u')"},
						},
					},
				},
			},

			// ==== allowEscape extended ====

			{
				Code:    `/[A\u0301]/`,
				Options: map[string]interface{}{"allowEscape": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass"},
				},
			},
			// ---- allowEscape: identity-escape of astral remains flagged ----
			// Even though `\👍` "looks escaped", the backslash is an identity
			// escape of the high surrogate (in JS UTF-16 semantics), so only
			// the high half is covered by the backslash; the low is raw.
			// Pair is still detected.
			{
				Code:    `RegExp('[\👍]')`,
				Options: map[string]interface{}{"allowEscape": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `RegExp('[\👍]', "u")`},
						},
					},
				},
			},
			{
				// `/[\👍]/` still has the pair (identity escape doesn't hide
				// it) but adding `u` makes the pattern a syntax error
				// (`\<non-syntax-char>` is invalid under u), so no
				// suggestUnicodeFlag fix.
				Code:    `/[\👍]/`,
				Options: map[string]interface{}{"allowEscape": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag"},
				},
			},

			// ==== Multi-identifier constant folding ====
			{
				// Concatenate two const-bound strings. tsgo's evaluator
				// recursively resolves each Identifier via our evaluateEntity
				// callback, then folds the BinaryExpression `+`.
				Code: `const a = "["; const b = "👍]"; new RegExp(a + b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `const a = "["; const b = "👍]"; new RegExp(a + b, "u");`},
						},
					},
				},
			},
			{
				// A const whose initializer is itself a TemplateExpression
				// with a constant span. Resolves through two levels.
				Code: "const pat = `[${\"👍\"}]`; new RegExp(pat);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: "const pat = `[${\"👍\"}]`; new RegExp(pat, \"u\");"},
						},
					},
				},
			},

			// ==== Destructured RegExp alias via TypeChecker ====
			{
				Code: `const {RegExp: A} = globalThis; new A("[👍]", "");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `const {RegExp: A} = globalThis; new A("[👍]", "u");`},
						},
					},
				},
			},
			{
				Code: `const {RegExp: R} = globalThis; R("[🇯🇵]", "u");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regionalIndicatorSymbol"},
				},
			},

			// ==== Heuristic: pattern validity under u flag ====
			// `\a` identity-letter → suggestion suppressed
			{
				Code: `var r = /[👍]\z/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag"}, // no suggestion
				},
			},
			// Bad `\xH` hex (only one digit) → also suppressed
			{
				Code: `var r = /[👍]\x1/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag"}, // no suggestion
				},
			},
			// Truncated `\u{...` → suppressed
			{
				Code: `var r = /[👍]\u{1F/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag"}, // no suggestion
				},
			},

			// ==== Regression: trailing-comma fix for `new RegExp(..., )` ====
			{
				// The fix must produce `new RegExp("...", "u",)` — NOT
				// `new RegExp("..." "u",,)` (which was the pre-fix bug where
				// the trailing-comma branch inserted text before the existing
				// comma, producing a double-comma + missing separator).
				Code: `var r = new RegExp("[🇯🇵]",)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = new RegExp("[🇯🇵]", "u",)`},
						},
					},
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = new RegExp("[🇯🇵]", "u",)`},
						},
					},
				},
			},

			// ==== Regression: dedup by parsed u/v state (not raw flag string) ====
			// `g` vs `i` have equivalent parser semantics (both non-uv), so the
			// rule must not double-report when a literal with one of those
			// flags is passed to RegExp() with the other.
			{
				Code: `const r = /[🇯🇵]/g; new RegExp(r, "i");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `const r = /[🇯🇵]/gu; new RegExp(r, "i");`},
						},
					},
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `const r = /[🇯🇵]/gu; new RegExp(r, "i");`},
						},
					},
				},
			},
			// Flag reordering — `"gu"` vs `"ug"` — should also dedup.
			{
				Code: `const r = /[🇯🇵]/gu; new RegExp(r, "ug");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regionalIndicatorSymbol"},
				},
			},

			// ==== Template literal with static expression substitution ====
			{
				// Template's only span is a StringLiteral — foldable.
				// All 4 pair-detects collapse into 1 diagnostic at node loc.
				Code: `new RegExp(` + "`${\"[👍🇯🇵]\"}[😊]`" + `);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `new RegExp(` + "`${\"[👍🇯🇵]\"}[😊]`" + `, "u");`},
						},
					},
				},
			},
			// ==== Binary-concat flag string ====
			{
				// `"u" + ""` → flags="u"; the overriding `u` suppresses the
				// pair-without-u detector but the class still has an emoji
				// modifier pair, so emojiModifier fires.
				Code: `new RegExp("[👶🏻]", "" + "u");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier"},
				},
			},
			// ==== String.raw tagged template as pattern ====
			{
				Code: "new RegExp(String.raw`[👍]`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: "new RegExp(String.raw`[👍]`, \"u\")"},
						},
					},
				},
			},
			// ==== let with literal initializer, no reassignments ====
			{
				Code: `let pattern = "[👍]"; new RegExp(pattern);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `let pattern = "[👍]"; new RegExp(pattern, "u");`},
						},
					},
				},
			},

			// ==== Identifier resolution via TypeChecker ====
			{
				Code: `const pattern = "[👍]"; new RegExp(pattern);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `const pattern = "[👍]"; new RegExp(pattern, "u");`},
						},
					},
				},
			},
			{
				Code: `const pattern = "[A\u0301]"; new RegExp(pattern);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass"},
				},
			},
			{
				Code: `const pattern = "[👶🏻]"; new RegExp(pattern, "u");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier"},
				},
			},
			{
				Code: `const pattern = /[🇯🇵]/u; new RegExp(pattern, "u");`,
				// When first arg is a regex literal and a flags arg is supplied,
				// the flags arg overrides — the rule should detect via the
				// call path. Resolving the identifier gives us a regex literal
				// which we then treat with the override flags (u here).
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regionalIndicatorSymbol"},
				},
			},

			// ==== Additional boundary cases (rslint-specific edge testing) ====

			// ---- Adjacent classes, only one misleading (decomposed Á) ----
			{
				Code: "var r = /[abc][A\u0301]/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass"},
				},
			},
			// ---- Escaped bracket right before a misleading class ----
			{
				Code: `var r = /\[[👍]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag", Line: 1, Column: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `var r = /\[[👍]/u`},
						},
					},
				},
			},
			// ---- Negated class containing a misleading sequence ----
			{
				Code: `var r = /[^🇯🇵]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regionalIndicatorSymbol"},
				},
			},

			// ==== Nested v-flag classes ====

			// ---- Misleading sequence in an inner class ----
			{
				Code: `var r = /[[👶🏻]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier", Line: 1, Column: 12},
				},
			},
			// ---- Two-level nesting with misleading content ----
			{
				Code: `var r = /[a[b[🇯🇵]]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regionalIndicatorSymbol"},
				},
			},
			// ---- Misleading sequence in BOTH outer and inner classes ----
			// Iteration order is innermost-first: 👶🏻 before 🇯🇵.
			{
				Code: `var r = /[🇯🇵[👶🏻]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier"},
					{MessageId: "regionalIndicatorSymbol"},
				},
			},
			// ---- Set subtraction LHS has misleading sequence ----
			{
				Code: `var r = /[🇯🇵--X]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regionalIndicatorSymbol"},
				},
			},
			// ---- Set subtraction RHS has misleading sequence ----
			{
				Code: `var r = /[X--[👶🏻]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "emojiModifier"},
				},
			},
			// ---- Set intersection with misleading sequence ----
			{
				Code: `var r = /[[👨‍👩‍👦]&&.]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "zwj"},
				},
			},
			// ---- Misleading sequence interrupted by property class ----
			{
				Code: `var r = /[👶\p{Letter}🏻]/u`,
				// \p{Letter} is a breaker — the pair is split, no emojiModifier.
				// Expected: no errors — but this is an invalid test case slot,
				// so must produce at least one error. We test via valid above instead.
				Skip: true,
			},
			// ---- Combining sequence inside nested class ----
			{
				Code: "var r = /[a[A\u0301]b]/v",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass"},
				},
			},

			// ==== Complex range edge cases ====

			// ---- Range from BMP to astral (u flag) ----
			// `[a-🏻]` — `a` to 🏻; `a` is not in modifier range, `🏻` is. No
			// detector fires because range endpoints split into separate sequences.
			// But `X🏻` (adjacent-to-range-end) detection depends on the next
			// elements. Here there's nothing after → valid.
			//
			// We test the inverse: range split correctly prevents false pair.
			{
				Code: "var r = /[a-b\u0301]/u",
				// `a-b` is range, then `\u0301` follows `b` (range max).
				// After range, sequence restarts at `b`. So we have
				// sequence [b, \u0301] → combiningClass fires.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass"},
				},
			},

			// ==== Template literal edge cases ====

			// ---- Raw CRLF in template between parts of misleading sequence ----
			{
				Code: "var r = RegExp(`[A\r\n\u0301]`)",
				// Under template cooked value: `[A\n\u0301]` (4 chars in class).
				// Between `A` and combining, there's an LF — it's a breaker-ish
				// intermediate code unit (not a regex breaker, just a char).
				// Sequence: [A, \n, \u0301] — combining follows \n, which is
				// also a non-combining char, so combiningClass fires.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass"},
				},
			},
			// ---- Escaped-quote inside string literal containing misleading class ----
			{
				Code: `var r = new RegExp("\"[Á]\"")`,
				// String value: `"[Á]"` (includes the double-quote chars).
				// Inside the class: Á (single code point U+00C1) — NOT a
				// combining sequence. No error. Move to valid... but
				// InvalidTestCase needs Errors, so we test a combining form:
				Skip: true,
			},
			{
				Code: "var r = new RegExp(\"\\\"[A\u0301]\\\"\")",
				// String resolved: `"[A + combining]"`. Class contains
				// [A + combining] → combiningClass.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "combiningClass"},
				},
			},

			// ==== ZWJ chain edges ====

			// ---- ZWJ at class boundary (ZWJ is last char before `]`) ----
			{
				Code: `var r = /[👨‍👩]/u`,
				// Two-person ZWJ sequence. Should fire zwj.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "zwj"},
				},
			},
			// ---- Multiple independent ZWJ sequences separated by a literal ----
			{
				Code: `var r = /[👨‍👩X👩‍👦]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "zwj"},
					{MessageId: "zwj"},
				},
			},
		},
	)
}
