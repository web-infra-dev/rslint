// cspell:ignore aeiou DFFB DDEF dedup
package no_misleading_character_class

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
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
			{Code: `new RegExp('[Á] [ ');`},         // syntax error in pattern → ignored
			{Code: `var r = new RegExp('[Á] [ ');`}, // ditto
			{Code: `var r = RegExp('{ [Á]', 'u');`}, // ditto
			{Code: `var r = RegExp(` + "`" + `${x}[👍]` + "`" + `)`},
			{Code: `var r = new RegExp('[🇯🇵]', ` + "`${foo}`" + `)`},
			{Code: `var r = new RegExp("[👍]", flags)`},
			{Code: `const args = ['[👍]', 'i']; new RegExp(...args);`},
			// Keep the constructor gate open while repeatedly exercising a
			// non-RegExp callee type; cached negative results must stay local
			// to that type and must not produce diagnostics.
			{Code: `function fake(pattern: string, flags: string) {} const marker = globalThis; fake("[👍]", ""); fake("[🇯🇵]", "u");`},

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
			// ---- Regex literal handled through RegExp(..., static flags) ----
			{Code: `const flags = "u"; RegExp(/[👍]/, flags);`},
			// ---- Dynamic override flags make the constructor own the literal ----
			{Code: `const flags = getFlags(); RegExp(/[👍]/, flags);`},

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
			//   - conditional expressions with an unknown test
			//   - object / array destructuring bindings
			// We match this behavior.
			{Code: `new RegExp("[👍]".repeat(1));`},
			{Code: `new RegExp(cond ? "[👍]" : "a");`},
			{Code: `const {pattern} = {pattern: "[👍]"}; new RegExp(pattern);`},
			{Code: `const [pattern] = ["[👍]"]; new RegExp(pattern);`},
			{Code: "{ const String = { raw: () => \"[👍]\" }; new RegExp(String.raw`[👍]`); }"},

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
				Code: `var r = R\u0065gExp("[👍]", "")`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "surrogatePairWithoutUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestUnicodeFlag",
						Output:    `var r = R\u0065gExp("[👍]", "u")`,
					}},
				}},
			},
			{
				Code: `var r = globalThis["RegExp"]("[👍]", "")`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "surrogatePairWithoutUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestUnicodeFlag",
						Output:    `var r = globalThis["RegExp"]("[👍]", "u")`,
					}},
				}},
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
			{
				// Interleave cached false and true callee types, then reuse
				// both. This attacks accidental cache poisoning between an
				// ordinary function and a destructured built-in alias.
				Code: `function fake(pattern: string, flags: string) {} const {RegExp: A} = globalThis; fake("[👍]", ""); A("[👍]", ""); fake("[🇯🇵]", "u"); A("[🇯🇵]", "u");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `function fake(pattern: string, flags: string) {} const {RegExp: A} = globalThis; fake("[👍]", ""); A("[👍]", "u"); fake("[🇯🇵]", "u"); A("[🇯🇵]", "u");`},
						},
					},
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
			{
				Code: "const RawString = String; new RegExp(RawString.raw`[👍]`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: "const RawString = String; new RegExp(RawString.raw`[👍]`, \"u\")"},
						},
					},
				},
			},
			// ==== Static conditional pattern ====
			{
				Code: `new RegExp(true ? "[👍]" : "[a]");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `new RegExp(true ? "[👍]" : "[a]", "u");`},
						},
					},
				},
			},
			// ==== Static String() pattern ====
			{
				Code: `new RegExp(String("[👍]"));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `new RegExp(String("[👍]"), "u");`},
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
			{
				Code: `const pattern = "[👍]" as string; new RegExp(pattern);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestUnicodeFlag", Output: `const pattern = "[👍]" as string; new RegExp(pattern, "u");`},
						},
					},
				},
			},
			// ==== Regex literal with static non-u flag override ====
			{
				Code: `const flags = ""; RegExp(/[👍]/, flags);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag"},
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

func TestNoMisleadingCharacterClassDefersSuggestions(t *testing.T) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.ts",
		Path:     "/edit-demand.ts",
	}, `/[\uD83D\uDC4D]/`, core.ScriptKindTS)

	for _, testCase := range []struct {
		name            string
		demand          rule.EditDemand
		wantBuilds      int
		wantSuggestions bool
	}{
		{name: "diagnostics only", demand: rule.EditDemandNone},
		{name: "autofix only", demand: rule.EditDemandAutofix},
		{name: "suggestions", demand: rule.EditDemandSuggestion, wantBuilds: 1, wantSuggestions: true},
		{name: "all edits", demand: rule.EditDemandAll, wantBuilds: 1, wantSuggestions: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			builds := 0
			var diagnostics []rule.RuleDiagnostic
			ctx := rule.RuleContext{SourceFile: sourceFile}.WithDiagnosticConsumer(
				NoMisleadingCharacterClassRule.Name,
				rule.SeverityError,
				rule.DiagnosticConsumer{
					Demand: testCase.demand,
					Report: func(diagnostic rule.RuleDiagnostic) {
						diagnostics = append(diagnostics, diagnostic)
					},
				},
			)

			emitMatch(ctx, foundMatch{
				kind:     "surrogatePairWithoutUFlag",
				srcStart: 2,
				srcEnd:   14,
			}, `[\uD83D\uDC4D]`, func() []rule.RuleFix {
				builds++
				return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(0, 0), "u")}
			})

			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
			}
			if builds != testCase.wantBuilds {
				t.Fatalf("suggestion builds = %d, want %d", builds, testCase.wantBuilds)
			}
			hasSuggestions := diagnostics[0].Suggestions != nil
			if hasSuggestions != testCase.wantSuggestions {
				t.Fatalf("suggestions present = %v, want %v", hasSuggestions, testCase.wantSuggestions)
			}
		})
	}
}

func TestSourceMayUseRegExp(t *testing.T) {
	for _, testCase := range []struct {
		name string
		code string
		want bool
	}{
		{name: "ordinary calls", code: `fn(); other();`},
		// Global-object identifiers deliberately keep the gate open. Their
		// computed property can be assembled from arbitrary static pieces,
		// so rejecting the file here could hide a real RegExp reference.
		{name: "conservative global reference", code: `const value = global; fn();`, want: true},
		{name: "direct constructor", code: `RegExp("x")`, want: true},
		{name: "escaped constructor identifier", code: `R\u0065gExp("x")`, want: true},
		{name: "computed global member", code: `globalThis["RegExp"]("x")`, want: true},
		{name: "folded computed global member", code: `globalThis["Reg" + "Exp"]("x")`, want: true},
		{name: "escaped computed global member", code: `globalThis["\u0052egExp"]("x")`, want: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/source-gate.ts",
				Path:     "/source-gate.ts",
			}, testCase.code, core.ScriptKindTS)
			if got := sourceMayUseRegExp(sourceFile); got != testCase.want {
				t.Fatalf("sourceMayUseRegExp(%q) = %v, want %v", testCase.code, got, testCase.want)
			}
		})
	}
}

// TestRunDetectorsOnCharsParity exhaustively compares the allocation-free
// detector path with the frozen pre-optimization implementation below. The
// alphabet covers every detector category, both sequence sentinels, escaped
// and unescaped forms, and the u-brace discriminator.
func TestRunDetectorsOnCharsParity(t *testing.T) {
	const maxLength = 4
	for length := 0; length <= maxLength; length++ {
		combinations := 1
		for range length {
			combinations *= detectorParityAlphabetSize
		}
		for encoded := range combinations {
			chars := detectorParityChars(encoded, length)
			for _, flags := range []utils.RegexFlags{
				{},
				{Unicode: true},
				{UnicodeSets: true},
			} {
				for _, allowEscape := range []bool{false, true} {
					assertDetectorParity(t, chars, flags, allowEscape)
				}
			}
		}
	}

	// Length four covers an individual ZWJ triplet; these explicit inputs
	// cover chained and interrupted sequences beyond the exhaustive bound.
	for _, chars := range [][]regexChar{
		detectorParityCharsFromIndexes(5, 9, 5, 9, 5),
		detectorParityCharsFromIndexes(5, 9, 5, 10, 5, 9, 5),
		detectorParityCharsFromIndexes(5, 9, 11, 9, 5),
		detectorParityCharsFromIndexes(5, 9, 5, 12, 5, 9, 5),
	} {
		for _, flags := range []utils.RegexFlags{{}, {Unicode: true}} {
			for _, allowEscape := range []bool{false, true} {
				assertDetectorParity(t, chars, flags, allowEscape)
			}
		}
	}
}

func TestCollapseSurrogatePairsInPlace(t *testing.T) {
	input := []regexChar{
		{value: 0xD83D, srcStart: 0, srcEnd: 6, raw: `\uD83D`},
		{value: 0xDC4D, srcStart: 6, srcEnd: 12, raw: `\uDC4D`},
		{value: 'A', srcStart: 12, srcEnd: 13, raw: "A"},
		{value: 0xD83C, srcStart: 13, srcEnd: 19, raw: `\uD83C`},
		{value: 0xDDEF, srcStart: 19, srcEnd: 25, raw: `\uDDEF`},
		{value: 0xD83D, srcStart: 25, srcEnd: 34, raw: `\u{D83D}`, isUBrace: true},
		{value: 0xDC4D, srcStart: 34, srcEnd: 40, raw: `\uDC4D`},
	}
	first := &input[0]
	got := collapseSurrogatePairs(input)
	want := []regexChar{
		{value: 0x1F44D, srcStart: 0, srcEnd: 12, raw: `\uD83D\uDC4D`},
		{value: 'A', srcStart: 12, srcEnd: 13, raw: "A"},
		{value: 0x1F1EF, srcStart: 13, srcEnd: 25, raw: `\uD83C\uDDEF`},
		{value: 0xD83D, srcStart: 25, srcEnd: 34, raw: `\u{D83D}`, isUBrace: true},
		{value: 0xDC4D, srcStart: 34, srcEnd: 40, raw: `\uDC4D`},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collapsed chars = %+v, want %+v", got, want)
	}
	if len(got) > 0 && &got[0] != first {
		t.Fatal("collapseSurrogatePairs allocated a new backing slice")
	}
}

func FuzzRunDetectorsOnCharsParity(f *testing.F) {
	for _, seed := range [][]byte{
		nil,
		{0, 1, 2, 3},
		{5, 9, 5, 9, 5},
		{6, 7},
		{11, 6, 7, 12},
		{5, 3, 4},
	} {
		for mode := range uint8(3) {
			f.Add(seed, mode, false)
			f.Add(seed, mode, true)
		}
	}

	f.Fuzz(func(t *testing.T, input []byte, mode uint8, allowEscape bool) {
		if len(input) > 256 {
			t.Skip()
		}
		chars := make([]regexChar, len(input))
		for i, value := range input {
			chars[i] = detectorParityChar(int(value)%detectorParityAlphabetSize, i)
		}
		var flags utils.RegexFlags
		switch mode % 3 {
		case 1:
			flags.Unicode = true
		case 2:
			flags.UnicodeSets = true
		}
		assertDetectorParity(t, chars, flags, allowEscape)
	})
}

func assertDetectorParity(t *testing.T, chars []regexChar, flags utils.RegexFlags, allowEscape bool) {
	t.Helper()
	got := runDetectorsOnChars(chars, flags, ruleOptions{allowEscape: allowEscape})
	want := runDetectorsBeforeOptimization(chars, flags, ruleOptions{allowEscape: allowEscape})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"detector mismatch for chars %+v, flags %+v, allowEscape %v:\ngot  %+v\nwant %+v",
			chars, flags, allowEscape, got, want,
		)
	}
}

const detectorParityAlphabetSize = 13

func detectorParityChars(encoded, length int) []regexChar {
	chars := make([]regexChar, length)
	for i := range chars {
		chars[i] = detectorParityChar(encoded%detectorParityAlphabetSize, i)
		encoded /= detectorParityAlphabetSize
	}
	return chars
}

func detectorParityCharsFromIndexes(indexes ...int) []regexChar {
	chars := make([]regexChar, len(indexes))
	for i, index := range indexes {
		chars[i] = detectorParityChar(index, i)
	}
	return chars
}

func detectorParityChar(index, position int) regexChar {
	char := [...]regexChar{
		{value: 'A', raw: "A"},
		{value: 0x0301, raw: "\u0301"},
		{value: 0x0301, raw: `\u0301`},
		{value: 0x1F466, raw: "👦"},
		{value: 0x1F3FB, raw: `\u{1F3FB}`, isUBrace: true},
		{value: 0x1F1EF, raw: "🇯"},
		{value: 0xD83D, raw: `\uD83D`},
		{value: 0xDC4D, raw: `\uDC4D`},
		{value: 0xD83D, raw: `\u{D83D}`, isUBrace: true},
		{value: 0x200D, raw: "\u200D"},
		{value: sentinelBreaker},
		{value: sentinelRangeBoundary},
		{value: 0x1F1F5, raw: "🇵"},
	}[index]
	char.srcStart = position * 3
	char.srcEnd = char.srcStart + 2
	return char
}

// runDetectorsBeforeOptimization is a frozen copy of the pre-optimization
// sequence materialization and detector dispatch. It is intentionally kept in
// test code as an independent behavioral oracle for the flat-slice rewrite.
func runDetectorsBeforeOptimization(chars []regexChar, flags utils.RegexFlags, opts ruleOptions) []foundMatch {
	var matches []foundMatch
	for _, sequence := range splitOnBreakerBeforeOptimization(chars) {
		active := sequence
		if opts.allowEscape {
			active = make([]*regexChar, len(sequence))
			for i, char := range sequence {
				if char == nil || isAcceptableEscape(char) {
					continue
				}
				active[i] = char
			}
		}
		matches = appendDetectorMatchesBeforeOptimization(matches, active, sequence, flags)
	}
	return matches
}

func splitOnBreakerBeforeOptimization(chars []regexChar) [][]*regexChar {
	var sequences [][]*regexChar
	var sequence []*regexChar
	flush := func() {
		if len(sequence) == 0 {
			return
		}
		sequences = append(sequences, sequence)
		sequence = nil
	}
	for _, char := range chars {
		switch char.value {
		case sentinelBreaker, sentinelRangeBoundary:
			flush()
		default:
			charCopy := char
			sequence = append(sequence, &charCopy)
		}
	}
	flush()
	return sequences
}

func appendDetectorMatchesBeforeOptimization(matches []foundMatch, chars, unfiltered []*regexChar, flags utils.RegexFlags) []foundMatch {
	if !flags.UV() {
		for i := 1; i < len(chars); i++ {
			previous, current := chars[i-1], chars[i]
			if previous != nil && current != nil &&
				isSurrogatePair(previous.value, current.value) &&
				!previous.isUBrace && !current.isUBrace {
				matches = append(matches, foundMatch{
					kind:     "surrogatePairWithoutUFlag",
					srcStart: previous.srcStart,
					srcEnd:   current.srcEnd,
				})
			}
		}
	} else {
		for i := 1; i < len(chars); i++ {
			previous, current := chars[i-1], chars[i]
			if previous != nil && current != nil &&
				isSurrogatePair(previous.value, current.value) &&
				(previous.isUBrace || current.isUBrace) {
				matches = append(matches, foundMatch{
					kind:     "surrogatePair",
					srcStart: previous.srcStart,
					srcEnd:   current.srcEnd,
				})
			}
		}
	}

	for i := 1; i < len(chars); i++ {
		previous, current := unfiltered[i-1], chars[i]
		if previous != nil && current != nil &&
			isCombiningCharacter(current.value) &&
			!isCombiningCharacter(previous.value) {
			matches = append(matches, foundMatch{
				kind:     "combiningClass",
				srcStart: previous.srcStart,
				srcEnd:   current.srcEnd,
			})
		}
	}
	for i := 1; i < len(chars); i++ {
		previous, current := chars[i-1], chars[i]
		if previous != nil && current != nil &&
			isEmojiModifier(current.value) &&
			!isEmojiModifier(previous.value) {
			matches = append(matches, foundMatch{
				kind:     "emojiModifier",
				srcStart: previous.srcStart,
				srcEnd:   current.srcEnd,
			})
		}
	}
	for i := 1; i < len(chars); i++ {
		previous, current := chars[i-1], chars[i]
		if previous != nil && current != nil &&
			isRegionalIndicatorSymbol(current.value) &&
			isRegionalIndicatorSymbol(previous.value) {
			matches = append(matches, foundMatch{
				kind:     "regionalIndicatorSymbol",
				srcStart: previous.srcStart,
				srcEnd:   current.srcEnd,
			})
		}
	}

	sequenceStart, sequenceEnd := -1, -1
	for i := 1; i < len(chars)-1; i++ {
		previous, current, next := chars[i-1], chars[i], chars[i+1]
		if previous == nil || current == nil || next == nil {
			continue
		}
		if current.value == 0x200D && previous.value != 0x200D && next.value != 0x200D {
			if sequenceStart >= 0 && sequenceEnd == previous.srcEnd {
				sequenceEnd = next.srcEnd
			} else {
				if sequenceStart >= 0 {
					matches = append(matches, foundMatch{
						kind:     "zwj",
						srcStart: sequenceStart,
						srcEnd:   sequenceEnd,
					})
				}
				sequenceStart = previous.srcStart
				sequenceEnd = next.srcEnd
			}
		}
	}
	if sequenceStart >= 0 {
		matches = append(matches, foundMatch{
			kind:     "zwj",
			srcStart: sequenceStart,
			srcEnd:   sequenceEnd,
		})
	}
	return matches
}
