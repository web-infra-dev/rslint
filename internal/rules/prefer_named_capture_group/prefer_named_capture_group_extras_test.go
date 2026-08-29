package prefer_named_capture_group

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferNamedCaptureGroupExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without
// breaking a named lock-in.
func TestPreferNamedCaptureGroupExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferNamedCaptureGroupRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: access/key forms that are not the RegExp global ----
			{
				Code:            "globalThis['NotRegExp']('(a)');",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
			},
			{
				Code:            "const key = 'RegExp'; globalThis[key]('(a)');",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
			},

			// ---- Dimension 4: shadowed globals are not built-ins ----
			{Code: "const RegExp = function() {}; RegExp('(a)');"},
			{
				Code:            "function f(globalThis) { globalThis.RegExp('(a)'); }",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
			},
			{Code: "let RegExp; new RegExp('(a)');"},
			// ReferenceTracker ignores every global RegExp use when the global has
			// been reassigned anywhere in the file.
			{Code: "RegExp = custom; RegExp('(a)' + '');"},

			// ---- Config `/* global RegExp: off */` / `languageOptions.globals` un-declares the builtin ----
			{Code: "new RegExp('(a)');", Globals: map[string]any{"RegExp": "off"}},
			{
				Code:            "new globalThis.RegExp('(a)');",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Globals:         map[string]any{"globalThis": "off"},
			},

			// ---- Real-user: StaticStringEvaluator has no scope, matching getStringIfConstant(node) called without a scope argument — a local variable's literal initializer is not resolved ----
			{Code: "const pattern = '(a)'; new RegExp(pattern);"},
			{Code: "const pattern = '(a)'; new RegExp('a' + pattern);"},

			// Documented difference from ESLint: rslint traces supported call-site
			// expressions back to the global constructor, but doesn't follow a local
			// alias assigned from that constructor.
			{Code: "const R = RegExp; new R('(a)');"},

			// ---- Dimension 4: empty pattern stays valid regardless of form ----
			{Code: "new RegExp(``);"},
			{Code: "RegExp('' + '');"},

			// ---- Dimension 4: lookaround constructs are not capturing groups ----
			{Code: "/(?=foo)/;"},
			{Code: "/(?!foo)/;"},
			{Code: "/(?<=foo)/;"},
			{Code: "/(?<!foo)/;"},

			// ---- Dimension 4: graceful degradation on malformed patterns ----
			{Code: "RegExp('[');"},                  // unterminated character class
			{Code: "RegExp('a{');"},                 // standalone quantifier-shaped literal, no group
			{Code: "RegExp('(?<dup>a)(?<dup>b)');"}, // duplicate group name — no unnamed group to report
			{Code: "new RegExp('(a){2,1}');"},       // descending quantifier bounds
			{Code: "new RegExp('(a)[z-a]');"},       // descending character-class range
			{Code: "new RegExp('(a)', 'u' + 'v');"}, // mutually exclusive Unicode modes

			// ---- ES2025 modifier-group headers the host engine rejects ----
			{Code: "RegExp('(?ii:(a))');"},  // flag repeated within one set
			{Code: "RegExp('(?i-i:(a))');"}, // same flag both added and removed
			{Code: "RegExp('(?-:(a))');"},   // both modifier sets empty

			// ---- Escapes that are only valid in the context they're missing ----
			{Code: `RegExp('\\q{a}(b)', 'v');`}, // \q{...} is a class-only string disjunction
			{Code: `RegExp('\\c(a)', 'u');`},    // \c without a control letter is a u-mode syntax error

			// N/A: declaration/container forms (class/function shape) do not affect
			// this expression-only rule — every observable branch is reached through
			// a regex literal or a RegExp() call argument, not a declaration site.
			// N/A: nesting/traversal boundaries specific to function or class scope
			// bodies do not apply; the only "nesting" this rule's own logic performs
			// is inside the regex pattern itself, covered by the deeply-nested-group
			// invalid case below.
		},
		[]rule_tester.InvalidTestCase{
			// ReferenceTracker follows conditional/logical pass-through values and
			// global-object aliases wrapped in a comma expression. Concatenated
			// patterns intentionally have no source-mapped suggestions.
			{
				Code:   "(true ? RegExp : other)('(a)' + '');",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "required"}},
			},
			{
				Code:   "(RegExp || other)('(a)' + '');",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "required"}},
			},
			{
				Code:            "(0, globalThis).RegExp('(a)' + '');",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "required"}},
			},

			// ---- ReferenceTracker-compatible callee expression forms ----
			{
				Code:            "globalThis['Reg' + 'Exp']('(a)');",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "globalThis['Reg' + 'Exp']('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "globalThis['Reg' + 'Exp']('(?:a)');"},
					},
				}},
			},
			{
				Code: "(0, RegExp)('(a)');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "(0, RegExp)('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "(0, RegExp)('(?:a)');"},
					},
				}},
			},

			// ---- Dimension 4: parenthesized/assertion/optional callee wrappers ----
			{
				Code: "((RegExp))('(a)');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "((RegExp))('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "((RegExp))('(?:a)');"},
					},
				}},
			},
			{
				Code: "(RegExp as any)('(a)' as string);",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "(RegExp as any)('(?<temp1>a)' as string);"},
						{MessageId: "addNonCapture", Output: "(RegExp as any)('(?:a)' as string);"},
					},
				}},
			},
			{
				Code: "(RegExp!)('(a)');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "(RegExp!)('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "(RegExp!)('(?:a)');"},
					},
				}},
			},
			{
				Code: "RegExp?.('(a)');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "RegExp?.('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "RegExp?.('(?:a)');"},
					},
				}},
			},
			{
				Code:            "globalThis?.RegExp?.('(a)');",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "globalThis?.RegExp?.('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "globalThis?.RegExp?.('(?:a)');"},
					},
				}},
			},

			// ---- Dimension 4: global-object member/element access forms ----
			{
				Code:            "new (globalThis.RegExp)('(a)');",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 31,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "new (globalThis.RegExp)('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "new (globalThis.RegExp)('(?:a)');"},
					},
				}},
			},
			{
				Code:            "globalThis['RegExp']('(a)');",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "globalThis['RegExp']('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "globalThis['RegExp']('(?:a)');"},
					},
				}},
			},
			{
				Code:            "globalThis[`RegExp`]('(a)');",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "globalThis[`RegExp`]('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "globalThis[`RegExp`]('(?:a)');"},
					},
				}},
			},

			// ---- Dimension 4: window/self/global are equally recognized global-object aliases ----
			{
				Code:    "window.RegExp('(a)');",
				Globals: map[string]any{"window": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "window.RegExp('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "window.RegExp('(?:a)');"},
					},
				}},
			},
			{
				Code:    "self.RegExp('(a)');",
				Globals: map[string]any{"self": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "self.RegExp('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "self.RegExp('(?:a)');"},
					},
				}},
			},
			{
				Code:    "global.RegExp('(a)');",
				Globals: map[string]any{"global": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "global.RegExp('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "global.RegExp('(?:a)');"},
					},
				}},
			},

			// Config declares RegExp as a writable/readonly global — still the builtin.
			{
				Code:    "new RegExp('(a)');",
				Globals: map[string]any{"RegExp": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "new RegExp('(?<temp1>a)');"},
						{MessageId: "addNonCapture", Output: "new RegExp('(?:a)');"},
					},
				}},
			},

			// ---- Dimension 4: pattern argument wrapped in parens/assertions still gets a suggestion ----
			{
				Code: "RegExp(('(a)'));",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "RegExp(('(?<temp1>a)'));"},
						{MessageId: "addNonCapture", Output: "RegExp(('(?:a)'));"},
					},
				}},
			},
			{
				Code: "RegExp('(a)' as string);",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    1,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "RegExp('(?<temp1>a)' as string);"},
						{MessageId: "addNonCapture", Output: "RegExp('(?:a)' as string);"},
					},
				}},
			},

			// ---- Dimension 4: lookaround groups don't suppress a sibling capturing group ----
			{
				Code: "/(?=foo)(bar)/;",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Message:   "Capture group '(bar)' should be converted to a named or non-capturing group.",
					Line:      1,
					Column:    1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "/(?=foo)(?<temp1>bar)/;"},
						{MessageId: "addNonCapture", Output: "/(?=foo)(?:bar)/;"},
					},
				}},
			},
			{
				Code: "/(?<!foo)(bar)/;",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Message:   "Capture group '(bar)' should be converted to a named or non-capturing group.",
					Line:      1,
					Column:    1,
					EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "/(?<!foo)(?<temp1>bar)/;"},
						{MessageId: "addNonCapture", Output: "/(?<!foo)(?:bar)/;"},
					},
				}},
			},

			// ---- Dimension 4: deeply nested unnamed groups (3+ levels), reported outer-to-inner ----
			{
				Code: "/(((a)))/;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(((a)))' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/(?<temp1>((a)))/;"},
							{MessageId: "addNonCapture", Output: "/(?:((a)))/;"},
						},
					},
					{
						MessageId: "required",
						Message:   "Capture group '((a))' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/((?<temp1>(a)))/;"},
							{MessageId: "addNonCapture", Output: "/((?:(a)))/;"},
						},
					},
					{
						MessageId: "required",
						Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/(((?<temp1>a)))/;"},
							{MessageId: "addNonCapture", Output: "/(((?:a)))/;"},
						},
					},
				},
			},

			// Locks in the ES2025 modifier-group header parser: combined flags with
			// a trailing `-` group don't create a capturing group, and the group
			// inside is still discovered.
			{
				Code:            "/(?ms-i:(a))bar/;",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
					Line:      1,
					Column:    1,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "/(?ms-i:(?<temp1>a))bar/;"},
						{MessageId: "addNonCapture", Output: "/(?ms-i:(?:a))bar/;"},
					},
				}},
			},

			// An empty removal set is still a well-formed header, so the group
			// inside is reported.
			{
				Code:            "/(?i-:(a))bar/;",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
					Line:      1,
					Column:    1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "/(?i-:(?<temp1>a))bar/;"},
						{MessageId: "addNonCapture", Output: "/(?i-:(?:a))bar/;"},
					},
				}},
			},

			// `\c` not followed by a control letter is Annex B's literal
			// backslash, so the `(` right after it still opens a capture group.
			{
				Code: `/\c(a)/;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
					Line:      1,
					Column:    1,
					EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: `/\c(?<temp1>a)/;`},
						{MessageId: "addNonCapture", Output: `/\c(?:a)/;`},
					},
				}},
			},
			{
				Code: `RegExp('\\c(a)');`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
					Line:      1,
					Column:    1,
					EndColumn: 17,
				}},
			},

			// Outside u/v mode `\q` is an identity escape, so the group after it
			// is reached.
			{
				Code: `RegExp('\\q{a}(b)');`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Message:   "Capture group '(b)' should be converted to a named or non-capturing group.",
					Line:      1,
					Column:    1,
					EndColumn: 20,
				}},
			},

			// Escaped IdentifierName characters are decoded when named
			// backreferences resolve, so the later unnamed group is still seen.
			{
				Code: `/(?<\u0061>a)\k<a>(b)/u;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Message:   "Capture group '(b)' should be converted to a named or non-capturing group.",
					Line:      1,
					Column:    1,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: `/(?<\u0061>a)\k<a>(?<temp1>b)/u;`},
						{MessageId: "addNonCapture", Output: `/(?<\u0061>a)\k<a>(?:b)/u;`},
					},
				}},
			},

			// Locks in the standalone `{` (outside u/v mode) being treated as a
			// literal character rather than a syntax error, so a later group in the
			// same pattern is still found.
			{
				Code: "RegExp('a{(b)');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Message:   "Capture group '(b)' should be converted to a named or non-capturing group.",
					Line:      1,
					Column:    1,
					EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "RegExp('a{(?<temp1>b)');"},
						{MessageId: "addNonCapture", Output: "RegExp('a{(?:b)');"},
					},
				}},
			},

			// Locks in the non-constant-flags branch: an unresolvable flags argument
			// doesn't suppress reporting — it's just treated as no flags.
			{
				Code: "const flags = 'g'; RegExp('(a)', flags);",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "required",
					Line:      1,
					Column:    20,
					EndColumn: 40,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addGroupName", Output: "const flags = 'g'; RegExp('(?<temp1>a)', flags);"},
						{MessageId: "addNonCapture", Output: "const flags = 'g'; RegExp('(?:a)', flags);"},
					},
				}},
			},

			// A regex literal nested inside a RegExp() call is checked twice: the
			// constructor reads it as the string `/(a)/`, slashes included, and the
			// regex-literal listener reads it as the pattern `(a)`.
			{
				Code: "new RegExp(/(a)/);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Line:      1,
						Column:    1,
						EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "new RegExp(/(a?<temp1>)/);"},
							{MessageId: "addNonCapture", Output: "new RegExp(/(a?:)/);"},
						},
					},
					{
						MessageId: "required",
						Line:      1,
						Column:    12,
						EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "new RegExp(/(?<temp1>a)/);"},
							{MessageId: "addNonCapture", Output: "new RegExp(/(?:a)/);"},
						},
					},
				},
			},

			// ---- Real-user: common multi-capture date/URL parsing shapes ----
			{
				Code: "const isoDate = /^(\\d{4})-(\\d{2})-(\\d{2})$/;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required", Message: "Capture group '(\\d{4})' should be converted to a named or non-capturing group.",
						Line: 1, Column: 17, EndColumn: 44,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "const isoDate = /^(?<temp1>\\d{4})-(\\d{2})-(\\d{2})$/;"},
							{MessageId: "addNonCapture", Output: "const isoDate = /^(?:\\d{4})-(\\d{2})-(\\d{2})$/;"},
						},
					},
					{
						MessageId: "required", Message: "Capture group '(\\d{2})' should be converted to a named or non-capturing group.",
						Line: 1, Column: 17, EndColumn: 44,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "const isoDate = /^(\\d{4})-(?<temp1>\\d{2})-(\\d{2})$/;"},
							{MessageId: "addNonCapture", Output: "const isoDate = /^(\\d{4})-(?:\\d{2})-(\\d{2})$/;"},
						},
					},
					{
						MessageId: "required", Message: "Capture group '(\\d{2})' should be converted to a named or non-capturing group.",
						Line: 1, Column: 17, EndColumn: 44,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "const isoDate = /^(\\d{4})-(\\d{2})-(?<temp1>\\d{2})$/;"},
							{MessageId: "addNonCapture", Output: "const isoDate = /^(\\d{4})-(\\d{2})-(?:\\d{2})$/;"},
						},
					},
				},
			},
			{
				Code: "const urlPattern = new RegExp('^(https?)://([^/]+)(/.*)?$');",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required", Message: "Capture group '(https?)' should be converted to a named or non-capturing group.",
						Line: 1, Column: 20, EndColumn: 60,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "const urlPattern = new RegExp('^(?<temp1>https?)://([^/]+)(/.*)?$');"},
							{MessageId: "addNonCapture", Output: "const urlPattern = new RegExp('^(?:https?)://([^/]+)(/.*)?$');"},
						},
					},
					{
						MessageId: "required", Message: "Capture group '([^/]+)' should be converted to a named or non-capturing group.",
						Line: 1, Column: 20, EndColumn: 60,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "const urlPattern = new RegExp('^(https?)://(?<temp1>[^/]+)(/.*)?$');"},
							{MessageId: "addNonCapture", Output: "const urlPattern = new RegExp('^(https?)://(?:[^/]+)(/.*)?$');"},
						},
					},
					{
						MessageId: "required", Message: "Capture group '(/.*)' should be converted to a named or non-capturing group.",
						Line: 1, Column: 20, EndColumn: 60,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "const urlPattern = new RegExp('^(https?)://([^/]+)(?<temp1>/.*)?$');"},
							{MessageId: "addNonCapture", Output: "const urlPattern = new RegExp('^(https?)://([^/]+)(?:/.*)?$');"},
						},
					},
				},
			},
		},
	)
}

// TestPreferNamedCaptureGroupEditDemand locks in that this rule's diagnostics
// are suggestion-only: diagnostic identity (message, range) is invariant
// across edit demand, no autofix is ever materialized, and suggestions
// appear only under EditDemandSuggestion/EditDemandAll.
func TestPreferNamedCaptureGroupEditDemand(t *testing.T) {
	t.Parallel()

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.ts",
		Path:     "/edit-demand.ts",
	}, `/(a)/;`, core.ScriptKindTS)
	statement := sourceFile.Statements.Nodes[0].AsExpressionStatement()
	literalNode := statement.Expression

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			DisableManager: rule.NewDisableManager(sourceFile, rule.NewCommentStore(sourceFile)),
		}.WithDiagnosticConsumer(PreferNamedCaptureGroupRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		})
		checkRegex(ctx, literalNode, literalNode, "(a)", "")
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
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
	for demand, diagnostic := range map[rule.EditDemand]rule.RuleDiagnostic{
		rule.EditDemandNone:       diagnosticsOnly,
		rule.EditDemandAutofix:    autofixOnly,
		rule.EditDemandSuggestion: suggestionOnly,
	} {
		if got, want := withoutEdits(diagnostic), withoutEdits(allEdits); !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic identity:\ngot  %#v\nwant %#v", demand, got, want)
		}
	}
	if diagnosticsOnly.Suggestions != nil || autofixOnly.Suggestions != nil {
		t.Fatal("non-suggestion demand materialized suggestions")
	}
	if diagnosticsOnly.FixesPtr != nil || autofixOnly.FixesPtr != nil ||
		suggestionOnly.FixesPtr != nil || allEdits.FixesPtr != nil {
		t.Fatal("suggestion-only rule materialized autofixes")
	}
	if suggestionOnly.Suggestions == nil ||
		!reflect.DeepEqual(suggestionOnly.Suggestions, allEdits.Suggestions) {
		t.Fatalf("suggestions differ between suggestion-only and all demand")
	}
	if suggestions := *allEdits.Suggestions; len(suggestions) != 2 ||
		suggestions[0].Message.Id != "addGroupName" ||
		suggestions[1].Message.Id != "addNonCapture" {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
}
