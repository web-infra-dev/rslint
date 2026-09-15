// TestRequireUnicodeRegexpExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// Dimension walk notes for require-unicode-regexp:
//   - Dimension 4 (declaration/container forms): N/A — the rule targets
//     regex literals and RegExp() calls, never a function or class
//     declaration.
//   - Dimension 4 (graceful degradation: SpreadAssignment/RestElement inside
//     an object literal or binding pattern, empty destructuring pattern,
//     overload signatures/abstract/declare members): N/A — none of these
//     shapes can appear as a regex literal, a RegExp call, or one of its two
//     arguments.
package require_unicode_regexp

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireUnicodeRegexpExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireUnicodeRegexpRule,
		[]rule_tester.ValidTestCase{
			// ReferenceTracker treats authored program-scope type declarations
			// as modifications of a script global, but not of a module global.
			{Code: "interface RegExp {}\nRegExp('x')", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			{Code: "type RegExp = unknown;\nRegExp('x')", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			{Code: "interface globalThis {}\nglobalThis.RegExp('x')", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			// Namespace and type-only import declarations own their scope variable
			// in both declaration spaces, so neither reference is the builtin.
			{Code: "namespace RegExp {}\nRegExp('x')", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			{Code: "import type { RegExp } from 'x';\nRegExp('x')", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// Static RegExp-literal properties that already carry u suppress the
			// report; mutation makes an alias unknown and conservatively skipped.
			{Code: `new RegExp("x", /u/u.source)`},
			{Code: `new RegExp("x", /x/u.flags)`},
			{Code: `const re = /g/u; re.lastIndex = 1; new RegExp("x", re.source)`},
			// eslint-utils 4.10.1 does not whitelist the unicodeSets getter.
			{Code: `new RegExp("x", /x/v.unicodeSets)`, Options: []any{map[string]any{"requireFlag": "v"}}},
			// A RegExp used as __proto__ supplies branded getters that throw for
			// the outer object, so its String value is not statically known.
			{Code: `new RegExp("x", String({__proto__: /u/u}))`},
			// ---- Locks in upstream trackMap arm: patternNode SpreadElement skips
			// the whole call, even with further arguments present ----
			{Code: "RegExp(...args, 'gi')"},
			// ---- Dimension 2: shadowing a global-object alias suppresses the
			// member-access form, same as shadowing RegExp itself ----
			{Code: "function f(globalThis) { return new globalThis.RegExp('foo') }"},
			// ---- Locks in the isDeclaredGlobal callback's negative branch: an
			// explicitly un-declared RegExp global is not a builtin callee ----
			{Code: "RegExp('foo')", Globals: map[string]any{"RegExp": "off"}},
			// ---- Dimension 3: a non-static (side-effecting) flags argument is
			// never reported, matching a non-literal flags identifier ----
			{Code: "new RegExp('foo', getFlags())"},
			// ReferenceTracker drops a global root after any write to that
			// binding, so a reassigned RegExp must not be treated as the builtin.
			{Code: "RegExp = custom; RegExp('foo')"},
		},
		[]rule_tester.InvalidTestCase{
			// A top-level type declaration in a module does not replace the
			// runtime global, while the same declaration in script mode does.
			{
				Code:            "interface RegExp {}\nRegExp('x')",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "interface RegExp {}\nRegExp('x', \"u\")"}},
				}},
			},
			{
				Code:            "type RegExp = unknown;\nRegExp('x')",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "type RegExp = unknown;\nRegExp('x', \"u\")"}},
				}},
			},
			{
				Code:            "{ interface RegExp {} RegExp('x') }",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "{ interface RegExp {} RegExp('x', \"u\") }"}},
				}},
			},
			{
				Code:            "interface RegExp {}\nglobalThis.RegExp('x')",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "interface RegExp {}\nglobalThis.RegExp('x', \"u\")"}},
				}},
			},
			{
				Code:            "interface globalThis {}\nglobalThis.RegExp('x')",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "interface globalThis {}\nglobalThis.RegExp('x', \"u\")"}},
				}},
			},
			// TSInstantiationExpression is transparent in eslint-utils' tracker.
			{
				Code: `(RegExp<string>)("x")`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `(RegExp<string>)("x", "u")`}},
				}},
			},
			{
				Code: `const R = RegExp<string>; R("x")`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `const R = RegExp<string>; R("x", "u")`}},
				}},
			},
			{
				Code: `const R = globalThis.RegExp<string>; new R("x")`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `const R = globalThis.RegExp<string>; new R("x", "u")`}},
				}},
			},
			// SourceFile.HasIdentifier normalizes escaped identifier spellings, so
			// the constructor fast path cannot hide this global reference.
			{
				Code: `R\u0065gExp("x")`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `R\u0065gExp("x", "u")`}},
				}},
			},
			// RegExp-literal properties participate in shared static evaluation.
			// Member-expression flags are reportable but intentionally not fixed.
			{
				Code:   `new RegExp("x", /g/u.source)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}},
			},
			{
				Code:   `const re = /g/u; new RegExp("x", re["source"])`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}},
			},
			{
				Code:   `new RegExp("x", ({__proto__: /g/u}).lastIndex)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}},
			},
			// ReferenceTracker follows a stable local alias of the builtin.
			{
				Code: "const R = RegExp; R('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId:   "requireUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `const R = RegExp; R('foo', "u")`}},
					},
				},
			},
			{
				Code: "let R; R = RegExp; new R('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId:   "requireUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `let R; R = RegExp; new R('foo', "u")`}},
					},
				},
			},
			{
				Code: "const { RegExp: R } = globalThis; R('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId:   "requireUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `const { RegExp: R } = globalThis; R('foo', "u")`}},
					},
				},
			},
			// Assignment-pattern shorthand carries globalThis.RegExp into the
			// existing local binding just like an explicit property assignment.
			{
				Code: "let RegExp; ({RegExp} = globalThis); RegExp('x')",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `let RegExp; ({RegExp} = globalThis); RegExp('x', "u")`}},
				}},
			},
			{
				Code: "function fake() {} (fake, RegExp)('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId:   "requireUFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `function fake() {} (fake, RegExp)('foo', "u")`}},
					},
				},
			},
			// ---- Dimension 4: optional call — tsgo keeps `?.` as a flag on the
			// CallExpression rather than an ESTree ChainExpression wrapper ----
			{
				Code: "RegExp?.('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `RegExp?.('foo', "u")`}},
					},
				},
			},
			// ---- Dimension 4: TS non-null assertion on the callee ----
			{
				Code: "RegExp!('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `RegExp!('foo', "u")`}},
					},
				},
			},
			// ---- Dimension 4: TS `as` type-expression wrapper on the callee ----
			{
				Code: "(RegExp as any)('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `(RegExp as any)('foo', "u")`}},
					},
				},
			},
			// ---- Dimension 4: single and multi-level parenthesized callee ----
			{
				Code: "(RegExp)('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `(RegExp)('foo', "u")`}},
					},
				},
			},
			{
				Code: "((RegExp))('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `((RegExp))('foo', "u")`}},
					},
				},
			},
			// ---- Dimension 4: element access on a global-object alias with a
			// static string-literal key ----
			{
				Code: "globalThis['RegExp']('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `globalThis['RegExp']('foo', "u")`}},
					},
				},
			},
			// ---- Dimension 4: element access on a global-object alias with a
			// static no-substitution template key ----
			{
				Code: "globalThis[`RegExp`]('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "globalThis[`RegExp`]('foo', \"u\")"}},
					},
				},
			},
			// ---- Dimension 2: detection reaches a RegExp call 3+ scopes deep
			// (class method -> nested function) without bleeding into or out of
			// the intervening scopes ----
			{
				Code: "class C { m() { function f() { return new RegExp('foo') } return f } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 39, EndLine: 1, EndColumn: 56,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `class C { m() { function f() { return new RegExp('foo', "u") } return f } }`}},
					},
				},
			},
			// ---- Dimension 3: a comment between the pattern and flags arguments
			// is preserved by the fix, which only replaces the flags node's own
			// range ----
			{
				Code: "new RegExp('foo', /* flags */ 'gimy')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 38,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "new RegExp('foo', /* flags */ 'gimyu')"}},
					},
				},
			},
			// ---- Dimension 3: multi-line call — the diagnostic spans the whole
			// call across lines and the fix targets only the flags node's own line ----
			{
				Code: "new RegExp(\n  'foo',\n  'gimy'\n)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 4, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "new RegExp(\n  'foo',\n  'gimyu'\n)"}},
					},
				},
			},
			// Locks in upstream Program() listener arm: the comma operator's
			// value is its right operand — the left operand (an unresolvable
			// identifier here) does not gate evaluation, matching
			// getStringIfConstant's SequenceExpression handling.
			{
				Code: "new RegExp('foo', (unknownVar, 'gi'))",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},
			// Locks in isValidWithUnicodeFlag's ecmaVersion <= 5 arm for the
			// default (unset) requireFlag: pre-ES2015 environments never get a
			// `u`-flag suggestion.
			{
				Code:            "/foo/",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// Pattern features are gated by the configured edition, not just by
			// the availability of the u/v flag itself.
			{
				Code:            `new RegExp("(?<a>x)")`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag"},
				},
			},
			{
				Code:            `new RegExp("(?i:a)")`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag"},
				},
			},
			// Locks in isValidWithUnicodeFlag's ecmaVersion <= 2023 arm at its
			// exact boundary (2023, one edition below the `v` flag's ES2024
			// introduction) — upstream only tests the boundary at ecmaVersion 6.
			{
				Code:            "/foo/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2023},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// Locks in the fixer's "we intentionally don't suggest concatenating
			// to non-literals" arm for a PropertyAccessExpression flags node that
			// evaluates successfully (so the report itself fires) but is
			// syntactically ineligible for a fix — upstream's own tests only
			// exercise a bare identifier there.
			{
				Code: "const obj = { flags: 'gi' }; new RegExp('foo', obj.flags)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 30, EndLine: 1, EndColumn: 58},
				},
			},
			// ---- Dimension 2: nested RegExp calls are each evaluated
			// independently — the outer call's non-static flags node (a nested
			// `new RegExp(...)`) suppresses only the outer report ----
			{
				Code: "new RegExp('foo', new RegExp('bar'))",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 19, EndLine: 1, EndColumn: 36,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `new RegExp('foo', new RegExp('bar', "u"))`}},
					},
				},
			},
			// Locks in the fixer's append branch (flags does not contain the
			// opposite flag) staying unguarded by the raw-backslash check that
			// only applies to the includes(flag) branch: a `g`-escaped
			// template still gets a plain `u` appended.
			{
				Code: "new RegExp('foo', `\\u0067`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "new RegExp('foo', `\\u0067u`)"}},
					},
				},
			},
			// ---- A lone `]` or `}` outside a class is a literal without the flag
			// but a SyntaxError with it, so no suggestion may be offered ----
			{
				Code: "/]/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 4},
				},
			},
			{
				Code: "/}/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 4},
				},
			},
			// ---- Under `u` the inner `[` is literal, so the trailing `]` is
			// stray and the pattern can't take the flag ----
			{
				Code: "/[[a][b]]/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			// ---- The same pattern nests cleanly under `v`, where the suggestion
			// stays available ----
			{
				Code:            "/[[a][b]]/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: "/[[a][b]]/v"}},
					},
				},
			},
			// ---- Braces closing a quantifier are structural, not stray ----
			{
				Code: "/a{1,2}/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "/a{1,2}/u"}},
					},
				},
			},
			// ---- Only the `v` grammar makes these class characters syntax
			// characters, so the suggestion has to be withheld under it ----
			{
				Code:            "/[(]/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			{
				Code:            "/[-]/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			{
				Code:            "/[a-]/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			// ---- A doubled punctuator is reserved under `v`; `&&` is the one
			// that stays legal, and only between two operands ----
			{
				Code:            "/[..]/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code:            "/[&&]/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code:            "/[a&&b]/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: "/[a&&b]/v"}},
					},
				},
			},
			// Each side of a v-class intersection must be one class-set operand;
			// a multi-element union cannot be used as the right operand.
			{
				Code:            "/[a&&bc]/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag"},
				},
			},
			// Escaped reserved punctuators are v-only class-set characters and
			// must not be rejected by an intermediate u-mode compilation.
			{
				Code:            `/[\&]/`,
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId:   "requireVFlag",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: `/[\&]/v`}},
					},
				},
			},
			// ---- Ordinary ranges and unions keep their suggestion under `v` ----
			{
				Code:            "/[a-z0-9_]/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 12,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: "/[a-z0-9_]/v"}},
					},
				},
			},
			// ---- `\k<name>` is literal text without the flag but a
			// SyntaxError with it unless the group exists ----
			{
				Code: "/\\k<a>/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code: "/(?<a>x)\\k<a>/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "/(?<a>x)\\k<a>/u"}},
					},
				},
			},
			// Escaped RegExpIdentifierName spellings normalize before the full
			// ECMAScript grammar gate, including escaped references and `$`.
			{
				Code: `/(?<\u0061>x)\k<a>/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `/(?<\u0061>x)\k<a>/u`}},
				}},
			},
			{
				Code: `/(?<a>x)\k<\u0061>/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `/(?<a>x)\k<\u0061>/u`}},
				}},
			},
			{
				Code: `/(?<\u0024>x)\k<$>/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `/(?<\u0024>x)\k<$>/u`}},
				}},
			},
			{
				Code: `new RegExp("(?<\\u{00000061}>x)\\k<a>")`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `new RegExp("(?<\\u{00000061}>x)\\k<a>", "u")`}},
				}},
			},
			// The tsgo grammar gate rejects unsafe u-mode conversions that the
			// previous regexp2-plus-hand-scanner path accepted.
			{Code: `new RegExp("\\q")`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}}},
			{Code: `/[\k<a>]/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}}},
			{Code: `/[\B]/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}}},
			{Code: `/\1/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}}},
			{Code: `/(a)\2/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}}},
			{Code: `/\01/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}}},
			{Code: `new RegExp("(?<\\u{D835}\\u{DC9C}>x)")`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}}},
			// Relaxed duplicate names only became valid in ES2025.
			{
				Code:            `/(?<a>x)|(?<a>y)/`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}},
			},
			{
				Code:            `/(?<a>x)|(?<a>y)/`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `/(?<a>x)|(?<a>y)/u`}},
				}},
			},
			{
				Code:            `/(?:(?<a>x)|(?:y|(?<a>z)))/`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `/(?:(?<a>x)|(?:y|(?<a>z)))/u`}},
				}},
			},
			{
				Code:            `/(?:x|(?<a>y))(?<a>z)/`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}},
			},
			{
				Code:            `/(?=(?<a>x))(?<a>y)/`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}},
			},
			// Full positive-set v grammar and canonical property names are
			// delegated to tsgo after the adapter's narrow safety checks.
			{
				Code:            `/[\q{abc}]/`,
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireVFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: `/[\q{abc}]/v`}},
				}},
			},
			{
				Code: `/\p{Script=Greek}/`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireUFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `/\p{Script=Greek}/u`}},
				}},
			},
			{
				Code:            `/[a^^b]/`,
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "requireVFlag"}},
			},
			{
				Code:            `/[^a\q{ab}]/`,
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "requireVFlag"}},
			},
			{
				Code:            `/[^a\q{b|c}]/`,
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireVFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: `/[^a\q{b|c}]/v`}},
				}},
			},
			{
				Code:            `/[^a\p{Letter}]/`,
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireVFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: `/[^a\p{Letter}]/v`}},
				}},
			},
			{
				Code:            `/[^\q{ab}&&a]/`,
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "requireVFlag",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: `/[^\q{ab}&&a]/v`}},
				}},
			},
			{
				Code:            `/[^a\p{Basic_Emoji}]/`,
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "requireVFlag"}},
			},
			{
				Code:            `/[^a-z\q{bc}]/`,
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "requireVFlag"}},
			},
			// Group-like text inside a character class does not declare a named
			// capture for an external backreference.
			{
				Code: `/[(?<a>x)]\k<a>/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag"},
				},
			},
			// Invalid named captures still require the flag but cannot receive an
			// unsafe suggestion that would make the pattern a syntax error.
			{
				Code:   `new RegExp("(?<a>x)(?<a>y)")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}},
			},
			{
				Code:   `new RegExp("(?<1>x)")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireUFlag"}},
			},
			// ---- Dimension 4: parenthesized flags argument — ESTree has no
			// parenthesis node, so the fix rewrites the literal inside ----
			{
				Code: "new RegExp('foo', ('gi'))",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "new RegExp('foo', ('giu'))"}},
					},
				},
			},
			// ---- Real-user: a Unicode property escape only becomes meaningful
			// under the `u` flag (Annex B reads `\p{L}` as literal `p{L}` without
			// it) — a common pattern for validating letters/digits ----
			{
				Code: "const isLetter = /^\\p{L}+$/;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 18, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "const isLetter = /^\\p{L}+$/u;"}},
					},
				},
			},
		},
	)
}

// TestRequireUnicodeRegexpSourceOnlyConstantFlags exercises the supported
// source-only Program path where the linter supplies binder references but no
// TypeChecker. Local constant flags must still be folded.
func TestRequireUnicodeRegexpSourceOnlyConstantFlags(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	rawProgram, sourceFile, err := helper.CreateTestProgram(
		`const flags = "g"; RegExp("foo", flags);`,
		"source-only.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}
	sourceProgram, err := lintprogram.NewFromBoundSources(rawProgram, rawProgram.SourceFiles())
	if err != nil {
		t.Fatal(err)
	}

	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     sourceProgram,
		File:        sourceFile.FileName(),
		HasTypeInfo: false,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     RequireUnicodeRegexpRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Fatal("source-only Program unexpectedly supplied a TypeChecker")
					}
					return RequireUnicodeRegexpRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})

	if len(diagnostics) != 1 || diagnostics[0].Message.Id != "requireUFlag" {
		t.Fatalf("diagnostics = %#v, want one requireUFlag", diagnostics)
	}
}

// TestRequireUnicodeRegexpEditDemand exercises Dimension 3 (autofix
// boundaries) for a suggestion-only rule: diagnostic identity (count,
// message, range) must stay identical across every edit demand, and the
// suggestion must materialize only when the suggestion or all-edits demand is
// requested — this rule never produces an autofix.
func TestRequireUnicodeRegexpEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"var a = /foo/;\n",
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     RequireUnicodeRegexpRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return RequireUnicodeRegexpRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for _, diagnostics := range [][]rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly} {
		if got, want := withoutEdits(diagnostics[0]), withoutEdits(allEdits[0]); !reflect.DeepEqual(got, want) {
			t.Errorf("changed diagnostic identity:\ngot  %#v\nwant %#v", got, want)
		}
	}

	if diagnosticsOnly[0].FixesPtr != nil || autofixOnly[0].FixesPtr != nil || suggestionOnly[0].FixesPtr != nil || allEdits[0].FixesPtr != nil {
		t.Fatalf("require-unicode-regexp never emits autofixes, but a fix was materialized")
	}

	if diagnosticsOnly[0].Suggestions != nil || autofixOnly[0].Suggestions != nil {
		t.Fatalf("non-suggestion demand unexpectedly materialized suggestions")
	}
	if suggestionOnly[0].Suggestions == nil || allEdits[0].Suggestions == nil {
		t.Fatalf("suggestion/all demand did not materialize suggestions")
	}
	if !reflect.DeepEqual(*suggestionOnly[0].Suggestions, *allEdits[0].Suggestions) {
		t.Fatalf("suggestion artifacts differ between suggestion-only and all demand")
	}
	if len(*suggestionOnly[0].Suggestions) != 1 {
		t.Fatalf("suggestions = %#v, want 1", *suggestionOnly[0].Suggestions)
	}
}
