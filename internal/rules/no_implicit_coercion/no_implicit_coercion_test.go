package no_implicit_coercion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoImplicitCoercion(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoImplicitCoercionRule,
		[]rule_tester.ValidTestCase{
			// Idiomatic replacements (what the rule recommends).
			{Code: `Boolean(foo)`},
			{Code: `foo.indexOf(1) !== -1`},
			{Code: `Number(foo)`},
			{Code: `parseInt(foo)`},
			{Code: `parseFloat(foo)`},
			{Code: `String(foo)`},

			// Single unary/binary operator on non-coerce target.
			{Code: `!foo`},
			{Code: `~foo`},
			{Code: `-foo`},
			{Code: `+1234`},
			{Code: `-1234`},
			{Code: `- -1234`},
			{Code: `+Number(lol)`},
			{Code: `-parseFloat(lol)`},

			// Multiplicative chains that aren't `1 * non-numeric`.
			{Code: `2 * foo`},
			{Code: `1 * 1234`},
			{Code: `123 - 0`},
			{Code: `1 * Number(foo)`},
			{Code: `1 * parseInt(foo)`},
			{Code: `1 * parseFloat(foo)`},
			{Code: `Number(foo) * 1`},
			{Code: `Number(foo) - 0`},
			{Code: `parseInt(foo) * 1`},
			{Code: `parseFloat(foo) * 1`},
			{Code: `- -Number(foo)`},
			{Code: `1 * 1234 * 678 * Number(foo)`},
			{Code: `1 * 1234 * 678 * parseInt(foo)`},
			{Code: `(1 - 0) * parseInt(foo)`},
			{Code: `1234 * 1 * 678 * Number(foo)`},
			{Code: `1234 * 1 * Number(foo) * Number(bar)`},
			{Code: `1234 * 1 * Number(foo) * parseInt(bar)`},
			{Code: `1234 * 1 * Number(foo) * parseFloat(bar)`},
			{Code: `1234 * 1 * parseInt(foo) * parseFloat(bar)`},
			{Code: `1234 * 1 * parseInt(foo) * Number(bar)`},
			{Code: `1234 * 1 * parseFloat(foo) * Number(bar)`},
			{Code: `1234 * Number(foo) * 1 * Number(bar)`},
			{Code: `1234 * parseInt(foo) * 1 * Number(bar)`},
			{Code: `1234 * parseFloat(foo) * 1 * parseInt(bar)`},
			{Code: `1234 * parseFloat(foo) * 1 * Number(bar)`},
			{Code: `(- -1234) * (parseFloat(foo) - 0) * (Number(bar) - 0)`},
			{Code: `1234*foo*1`},
			{Code: `1234*1*foo`},
			{Code: `1234*bar*1*foo`},
			{Code: `1234*1*foo*bar`},
			{Code: `1234*1*foo*Number(bar)`},
			{Code: `1234*1*Number(foo)*bar`},
			{Code: `1234*1*parseInt(foo)*bar`},
			{Code: `0 + foo`},
			{Code: `~foo.bar()`},

			// String concatenation with a non-empty literal is fine.
			{Code: `foo + 'bar'`},
			{Code: "foo + `${bar}`"},

			// Option toggles — rule types individually disabled.
			{Code: `!!foo`, Options: map[string]interface{}{"boolean": false}},
			{Code: `~foo.indexOf(1)`, Options: map[string]interface{}{"boolean": false}},
			{Code: `+foo`, Options: map[string]interface{}{"number": false}},
			{Code: `-(-foo)`, Options: map[string]interface{}{"number": false}},
			{Code: `foo - 0`, Options: map[string]interface{}{"number": false}},
			{Code: `1*foo`, Options: map[string]interface{}{"number": false}},
			{Code: `""+foo`, Options: map[string]interface{}{"string": false}},
			{Code: `foo += ""`, Options: map[string]interface{}{"string": false}},

			// Allowlist entries.
			{Code: `var a = !!foo`, Options: map[string]interface{}{"boolean": true, "allow": []interface{}{"!!"}}},
			{Code: `var a = ~foo.indexOf(1)`, Options: map[string]interface{}{"boolean": true, "allow": []interface{}{"~"}}},
			{Code: `var a = ~foo`, Options: map[string]interface{}{"boolean": true}},
			{Code: `var a = 1 * foo`, Options: map[string]interface{}{"boolean": true, "allow": []interface{}{"*"}}},
			{Code: `- -foo`, Options: map[string]interface{}{"number": true, "allow": []interface{}{"- -"}}},
			{Code: `foo - 0`, Options: map[string]interface{}{"number": true, "allow": []interface{}{"-"}}},
			{Code: `var a = +foo`, Options: map[string]interface{}{"boolean": true, "allow": []interface{}{"+"}}},
			{Code: `var a = "" + foo`, Options: map[string]interface{}{"boolean": true, "string": true, "allow": []interface{}{"+"}}},

			// https://github.com/eslint/eslint/issues/7057 — both operands already string.
			{Code: `'' + 'foo'`},
			{Code: "`` + 'foo'"},
			{Code: "'' + `${foo}`"},
			{Code: `'foo' + ''`},
			{Code: "'foo' + ``"},
			{Code: "`${foo}` + ''"},
			{Code: `foo += 'bar'`},
			{Code: "foo += `${bar}`"},

			// disallowTemplateShorthand: non-shorthand templates don't trigger.
			{Code: "`a${foo}`", Options: map[string]interface{}{"disallowTemplateShorthand": true}},
			{Code: "`${foo}b`", Options: map[string]interface{}{"disallowTemplateShorthand": true}},
			{Code: "`${foo}${bar}`", Options: map[string]interface{}{"disallowTemplateShorthand": true}},
			{Code: "tag`${foo}`", Options: map[string]interface{}{"disallowTemplateShorthand": true}},
			// Default is off.
			{Code: "`${foo}`"},
			{Code: "`${foo}`", Options: map[string]interface{}{"disallowTemplateShorthand": false}},
			{Code: `+42`},

			// https://github.com/eslint/eslint/issues/14623 — String(...) operand is already string.
			{Code: `'' + String(foo)`},
			{Code: `String(foo) + ''`},
			{Code: "`` + String(foo)"},
			{Code: "String(foo) + ``"},
			{Code: "`${'foo'}`", Options: map[string]interface{}{"disallowTemplateShorthand": true}},
			{Code: "`${`foo`}`", Options: map[string]interface{}{"disallowTemplateShorthand": true}},
			{Code: "`${String(foo)}`", Options: map[string]interface{}{"disallowTemplateShorthand": true}},

			// https://github.com/eslint/eslint/issues/16373 — fraction-of-one pattern.
			{Code: `console.log(Math.PI * 1/4)`},
			{Code: `a * 1 / 2`},
			{Code: `a * 1 / b`},

			// Parenthesised callee — ESLint treats `(Number)(x)` as numeric and
			// `(String)(x)` as string, since parens are transparent.
			{Code: `+(Number)(foo)`},
			{Code: `- -(Number)(foo)`},
			{Code: `(Number)(foo) * 1`},
			{Code: `(Number)(foo) - 0`},
			{Code: `'' + (String)(foo)`},
			{Code: `(String)(foo) + ''`},
			{Code: "`` + (String)(foo)"},
			{Code: "(String)(foo) + ``"},
			{Code: "`${(String)(foo)}`", Options: map[string]interface{}{"disallowTemplateShorthand": true}},
			// Doubly-parenthesised callee.
			{Code: `+((Number))(foo)`},
		},
		[]rule_tester.InvalidTestCase{
			// !!foo — fix applied when Boolean not shadowed. Autofix cases have
			// no suggestions (ESLint drops the suggestion when `shouldFix`).
			{
				Code:   `!!foo`,
				Output: []string{`Boolean(foo)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			{
				Code:   `!!(foo + bar)`,
				Output: []string{`Boolean(foo + bar)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// Boolean shadowed — suggestion only, no autofix.
			{
				Code: `!!(foo + bar); var Boolean = null;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Boolean(foo + bar); var Boolean = null;`},
						},
					},
				},
			},
			// ~foo.indexOf(x) — no fix, no suggestion (reported only).
			{
				Code: `~foo.indexOf(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			{
				Code: `~foo.bar.indexOf(2)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// +foo — suggestion only.
			{
				Code: `+foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			{
				Code: `-(-foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			{
				Code: `+foo.bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo.bar)`},
						},
					},
				},
			},
			// Multiply by 1.
			{
				Code: `1*foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			{
				Code: `foo*1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			{
				Code: `1*foo.bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo.bar)`},
						},
					},
				},
			},
			// Subtract 0.
			{
				Code: `foo.bar-0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo.bar)`},
						},
					},
				},
			},
			// "" + x / x + "".
			{
				Code: `""+foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo)`},
						},
					},
				},
			},
			{
				Code: "``+foo",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo)`},
						},
					},
				},
			},
			{
				Code: `foo+""`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo)`},
						},
					},
				},
			},
			{
				Code: "foo+``",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo)`},
						},
					},
				},
			},
			{
				Code: `""+foo.bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo.bar)`},
						},
					},
				},
			},
			{
				Code: "``+foo.bar",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo.bar)`},
						},
					},
				},
			},
			{
				Code: `foo.bar+""`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo.bar)`},
						},
					},
				},
			},
			{
				Code: "foo.bar+``",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo.bar)`},
						},
					},
				},
			},
			// Template shorthand.
			{
				Code:    "`${foo}`",
				Options: map[string]interface{}{"disallowTemplateShorthand": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo)`},
						},
					},
				},
			},
			// x += "".
			{
				Code: `foo += ""`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `foo = String(foo)`},
						},
					},
				},
			},
			{
				Code: "foo += ``",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `foo = String(foo)`},
						},
					},
				},
			},
			// allow list doesn't suppress a different operator.
			{
				Code:    `var a = !!foo`,
				Output:  []string{`var a = Boolean(foo)`},
				Options: map[string]interface{}{"boolean": true, "allow": []interface{}{"~"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 9},
				},
			},
			{
				Code:    `var a = ~foo.indexOf(1)`,
				Options: map[string]interface{}{"boolean": true, "allow": []interface{}{"!!"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 9},
				},
			},
			{
				Code:    `var a = 1 * foo`,
				Options: map[string]interface{}{"boolean": true, "allow": []interface{}{"+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `var a = Number(foo)`},
						},
					},
				},
			},
			{
				Code:    `var a = +foo`,
				Options: map[string]interface{}{"boolean": true, "allow": []interface{}{"*"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `var a = Number(foo)`},
						},
					},
				},
			},
			{
				Code:    `var a = "" + foo`,
				Options: map[string]interface{}{"boolean": true, "allow": []interface{}{"*"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `var a = String(foo)`},
						},
					},
				},
			},
			{
				Code:    "var a = `` + foo",
				Options: map[string]interface{}{"boolean": true, "allow": []interface{}{"*"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `var a = String(foo)`},
						},
					},
				},
			},
			// typeof+foo — leading-space handling keeps output lexable.
			{
				Code: `typeof+foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `typeof Number(foo)`},
						},
					},
				},
			},
			{
				Code: `typeof +foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `typeof Number(foo)`},
						},
					},
				},
			},
			// BigInt operand: `'' + 1n` still flagged (1n is not isNumeric).
			{
				Code: `let x ='' + 1n;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `let x =String(1n);`},
						},
					},
				},
			},
			// Optional chaining — `>= 0` recommendation.
			{
				Code: `~foo?.indexOf(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// Parens around the optional chain callee break the chain — `!== -1`.
			{
				Code: `~(foo?.indexOf)(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// Fraction-of-one regression cases (issue 16373).
			{
				Code: `1 * a / 2`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(a) / 2`},
						},
					},
				},
			},
			{
				Code: `(a * 1) / 2`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `(Number(a)) / 2`},
						},
					},
				},
			},
			{
				Code: `a * 1 / (b * 1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `a * 1 / (Number(b))`},
						},
					},
				},
			},
			{
				Code: `a * 1 + 2`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(a) + 2`},
						},
					},
				},
			},

			// --- Additional edge cases: parenthesised variants ---

			// `!!` with parens between the two `!`s (transparent in ESLint).
			{
				Code:   `!(!foo)`,
				Output: []string{`Boolean(foo)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// Deeply nested parens between `!`s.
			{
				Code:   `!((!foo))`,
				Output: []string{`Boolean(foo)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// `!!` on a parenthesised inner — recommendation drops the parens.
			{
				Code:   `!!((foo))`,
				Output: []string{`Boolean(foo)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// Triple negation: the outer `!!` (`!!` of `!foo`) and the middle
			// `!!` (`!!foo`) each match, same as ESLint.
			{
				Code:   `!!!foo`,
				Output: []string{`Boolean(!foo)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
					{MessageId: "implicitCoercion", Line: 1, Column: 2},
				},
			},
			// TS `as` expression wrapped in `!!` — target text keeps the cast.
			{
				Code:   `!!(foo as any)`,
				Output: []string{`Boolean(foo as any)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},

			// `~indexOf` variants.
			{
				Code: `~foo.lastIndexOf(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			{
				Code: `~foo['indexOf'](1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			{
				Code: "~foo[`indexOf`](1)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// Outer parens around the call argument.
			{
				Code: `~(foo.indexOf(1))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// Optional call: `?.(` on the call itself — still an optional chain.
			{
				Code: `~foo.indexOf?.(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// Chained member before `.indexOf` with optional access — chain flag propagates.
			{
				Code: `~foo?.bar.indexOf(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},

			// `+foo` with parens.
			{
				Code: `+(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			// `+` on BigInt (not isNumeric) — still reported; `Number(1n)` throws
			// at runtime but that mirrors ESLint's behavior exactly.
			{
				Code: `+1n`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(1n)`},
						},
					},
				},
			},
			// Stacked `+`: both are reported independently.
			{
				Code: `+ +foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(+foo)`},
						},
					},
					{
						MessageId: "implicitCoercion", Line: 1, Column: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `+ Number(foo)`},
						},
					},
				},
			},
			// `- -` with inner parens.
			{
				Code: `-(- foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			// `- -` with double parens.
			{
				Code: `-((-foo))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},

			// `*1` / `1*` with parens around the numeric or operand.
			{
				Code: `(1) * foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			{
				Code: `foo * (1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			// Normalized numeric literals (1.0, 0x1, 1e0).
			{
				Code: `1.0 * foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			{
				Code: `0x1 * foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			{
				Code: `1e0 * foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			// Stacked `*1`: inner is flagged but outer (non-literal chain) isn't.
			{
				Code: `1 * foo * 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo) * 1`},
						},
					},
				},
			},

			// `- 0` with parens on operand (node replaced whole, parens absorbed).
			{
				Code: `(foo) - 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			{
				Code: `foo - (0)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			// Normalized zero (0.0, 0x0).
			{
				Code: `foo - 0.0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},
			{
				Code: `foo - 0x0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo)`},
						},
					},
				},
			},

			// `"" + x` with parens on operand — parens dropped in recommendation.
			{
				Code: `"" + (foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo)`},
						},
					},
				},
			},
			{
				Code: `(foo) + ""`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo)`},
						},
					},
				},
			},
			// Member access / element access.
			{
				Code: `obj.prop + ""`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(obj.prop)`},
						},
					},
				},
			},
			{
				Code: `arr[0] + ""`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(arr[0])`},
						},
					},
				},
			},
			// "" + (x + y) — parens preserved in the inner expression text.
			{
				Code: `"" + (foo + bar)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo + bar)`},
						},
					},
				},
			},

			// `` `${expr}` `` variants.
			{
				Code:    "`${foo.bar}`",
				Options: map[string]interface{}{"disallowTemplateShorthand": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo.bar)`},
						},
					},
				},
			},
			// Parens around the interpolated expression — stripped in output.
			{
				Code:    "`${(foo)}`",
				Options: map[string]interface{}{"disallowTemplateShorthand": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo)`},
						},
					},
				},
			},
			// Nested coercion: outer template is a string (inner template counts as
			// string type), so only the inner `` `${foo}` `` is flagged.
			{
				Code:    "`${`${foo}`}`",
				Options: map[string]interface{}{"disallowTemplateShorthand": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 4,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: "`${String(foo)}`"},
						},
					},
				},
			},

			// `+=` with parens or member targets.
			{
				Code: `obj.prop += ""`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `obj.prop = String(obj.prop)`},
						},
					},
				},
			},
			{
				Code: `arr[0] += ""`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `arr[0] = String(arr[0])`},
						},
					},
				},
			},

			// Scope/shadowing edge cases for `!!`.
			{
				// Shadowed in a function — suggestion only.
				Code: `function f() { var Boolean = 1; return !!foo; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 40,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `function f() { var Boolean = 1; return Boolean(foo); }`},
						},
					},
				},
			},
			{
				// Shadowed by a block-scoped let — suggestion only.
				Code: `{ let Boolean = 1; !!foo; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `{ let Boolean = 1; Boolean(foo); }`},
						},
					},
				},
			},
			{
				// Shadowing in an enclosing scope outside the block — autofix
				// applies because the reference is not shadowed where it appears.
				Code:   `{ let Boolean = 1; } !!foo;`,
				Output: []string{`{ let Boolean = 1; } Boolean(foo);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 22},
				},
			},

			// Mixed/nested coercion chains (each inner pattern reported).
			{
				// `"" + foo + bar` parses as `("" + foo) + bar`; the inner is
				// the coercion, the outer is plain concatenation.
				Code: `"" + foo + bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo) + bar`},
						},
					},
				},
			},
			{
				// `!!foo && !!bar` — both outer `!!` patterns are reported.
				Code:   `!!foo && !!bar`,
				Output: []string{`Boolean(foo) && Boolean(bar)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
					{MessageId: "implicitCoercion", Line: 1, Column: 10},
				},
			},
			// `!!` inside a conditional — each branch reported independently.
			{
				Code:   `cond ? !!foo : !!bar`,
				Output: []string{`cond ? Boolean(foo) : Boolean(bar)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 8},
					{MessageId: "implicitCoercion", Line: 1, Column: 16},
				},
			},
			// `!!` in an arrow body.
			{
				Code:   `const f = x => !!x;`,
				Output: []string{`const f = x => Boolean(x);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 16},
				},
			},
			// `- -` applied to a call result (non-numeric wrapper).
			{
				Code: `- -foo()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo())`},
						},
					},
				},
			},
			// `- -` with inner numeric: skipped (no coercion needed).
			// (Already covered as valid: `- -1234`, `- -Number(foo)`.)

			// String concat inside a larger chain — ensure the outer (which
			// treats the left side as a non-string BinaryExpression) isn't
			// flagged; only the inner `"" + foo` is.
			{
				Code: `"" + foo + bar + baz`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo) + bar + baz`},
						},
					},
				},
			},

			// Multiply inside a division chain that doesn't form a `(x*1)/y`
			// fraction — the inner `x*1` is still coercion.
			{
				Code: `x * 1 * y / z`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						// inner `x * 1`: parent is `* y`, not `/`, so fraction
						// check fails and we report.
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(x) * y / z`},
						},
					},
				},
			},

			// Optional chain variants for the `~` pattern — recommendation text
			// keeps the callee intact.
			{
				Code: `~foo?.bar.indexOf(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},

			// --- More nesting: coercion patterns inside other expression contexts ---

			// `!!` inside object property value.
			{
				Code:   `const o = { a: !!foo };`,
				Output: []string{`const o = { a: Boolean(foo) };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 16},
				},
			},
			// `!!` inside destructuring default.
			{
				Code:   `const { x = !!foo } = obj;`,
				Output: []string{`const { x = Boolean(foo) } = obj;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 13},
				},
			},
			// `!!` inside spread in array literal.
			{
				Code:   `const a = [...!!foo];`,
				Output: []string{`const a = [...Boolean(foo)];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 15},
				},
			},
			// `!!` over non-null assertion (TS-specific node, preserved verbatim).
			{
				Code:   `!!(foo!)`,
				Output: []string{`Boolean(foo!)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},

			// `"" + foo` inside a template — inner BinExpr is flagged.
			{
				Code: "`${'' + foo}`",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 4,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: "`${String(foo)}`"},
						},
					},
				},
			},

			// `*1` multiplied with BigInt on the other side — non-numeric so reported.
			{
				Code: `1n * 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(1n)`},
						},
					},
				},
			},

			// `+` on a call result (non-numeric) — still flagged.
			{
				Code: `+foo()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `Number(foo())`},
						},
					},
				},
			},

			// Both `(x * 1)` and `(y * 1)` parenthesised — neither qualifies as
			// a fraction-of-one, so both are reported.
			{
				Code: `(a * 1) * (b * 1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `(Number(a)) * (b * 1)`},
						},
					},
					{
						MessageId: "implicitCoercion", Line: 1, Column: 12,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `(a * 1) * (Number(b))`},
						},
					},
				},
			},

			// --- Remaining corner cases ---

			// `Boolean` shadowed as a function parameter — suggestion only.
			{
				Code: `function f(Boolean) { return !!foo; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 30,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `function f(Boolean) { return Boolean(foo); }`},
						},
					},
				},
			},
			// `"" + foo + ""` — parses as `("" + foo) + ""`; both layers match
			// ESLint's literal check (outer's left is a BinaryExpression, which
			// isStringType does not recognize as string), so both are reported.
			{
				Code: `"" + foo + ""`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo) + ""`},
						},
					},
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String("" + foo)`},
						},
					},
				},
			},
			// Template shorthand with a trailing line-continuation (cooked tail
			// still empty, so the shorthand pattern applies).
			{
				Code:    "`${foo}\\\n`",
				Options: map[string]interface{}{"disallowTemplateShorthand": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `String(foo)`},
						},
					},
				},
			},
			// `!!` on an awaited expression — flagged, inner text preserved.
			{
				Code:   `async function f() { return !!(await g()); }`,
				Output: []string{`async function f() { return Boolean(await g()); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 29},
				},
			},
			// `~` on `indexOf` applied to a call-result receiver.
			{
				Code: `~foo().indexOf(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// `~` with parens around the non-optional callee.
			{
				Code: `~(foo.indexOf)(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// Computed property key wrapped in parens — ESLint treats parens
			// transparently in the `isSpecificMemberAccess` check.
			{
				Code: `~foo[('indexOf')](1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			{
				Code: "~foo[(`lastIndexOf`)](1)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "implicitCoercion", Line: 1, Column: 1},
				},
			},
			// `Boolean` shadowed by a class name at the top level — suggestion only.
			{
				Code: `class Boolean {} !!foo;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `class Boolean {} Boolean(foo);`},
						},
					},
				},
			},
			// `Boolean` shadowed by a namespace at the top level — suggestion only.
			{
				Code: `namespace Boolean {} !!foo;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `namespace Boolean {} Boolean(foo);`},
						},
					},
				},
			},
			// Keyword adjacency: `void+foo` — fix must keep a separating space.
			{
				Code: `void+foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 5,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `void Number(foo)`},
						},
					},
				},
			},
			// `~` applied to `+call.indexOf(...)` — the outer `~` doesn't match
			// (operand isn't a CallExpression), but the inner `+call` does.
			{
				Code: `~+foo.indexOf(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "implicitCoercion", Line: 1, Column: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useRecommendation", Output: `~Number(foo.indexOf(1))`},
						},
					},
				},
			},
		},
	)
}
