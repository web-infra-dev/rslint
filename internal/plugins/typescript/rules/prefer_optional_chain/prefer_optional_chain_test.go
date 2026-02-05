package prefer_optional_chain

import (
	"regexp"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferOptionalChainRule(t *testing.T) {
	// =====================================================================
	// GENERATED BASE CASES (mirrors JS BaseCases invocations)
	// =====================================================================
	generatedValid := []rule_tester.ValidTestCase{}
	generatedInvalid := []rule_tester.InvalidTestCase{}

	// --- && operator group ---

	// 1. BaseCases({ operator: '&&' }) — boolean truthy, 26 invalid cases
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "&&",
	})...)

	// 2. BaseCases({ mutateCode: c => c.replace(/;$/, ' && bing;'), operator: '&&' })
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "&&",
		mutateCode: func(c string) string {
			return strings.TrimSuffix(c, ";") + " && bing;"
		},
	})...)

	// 3. BaseCases({ mutateCode: c => c.replace(/;$/, ' && bing.bong;'), operator: '&&' })
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "&&",
		mutateCode: func(c string) string {
			return strings.TrimSuffix(c, ";") + " && bing.bong;"
		},
	})...)

	// 4. !== null valid (guard strips undefined from type → chain already safe)
	generatedValid = append(generatedValid, generateValidBaseCases(baseCaseOptions{
		operator: "&&",
		mutateCode: func(c string) string {
			return strings.ReplaceAll(c, "&&", "!== null &&")
		},
	})...)

	// 5. !== null invalid (mutateDeclaration removes `| undefined`)
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "&&",
		mutateCode: func(c string) string {
			return strings.ReplaceAll(c, "&&", "!== null &&")
		},
		mutateDeclaration: func(c string) string {
			return strings.ReplaceAll(c, "| undefined", "")
		},
		mutateOutput:       identity,
		useSuggestionFixer: true,
	})...)

	// 6. != null invalid
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "&&",
		mutateCode: func(c string) string {
			return strings.ReplaceAll(c, "&&", "!= null &&")
		},
		mutateOutput:       identity,
		useSuggestionFixer: true,
	})...)

	// 7. !== undefined valid (skipIds [20, 26])
	generatedValid = append(generatedValid, generateValidBaseCases(baseCaseOptions{
		operator: "&&",
		mutateCode: func(c string) string {
			return strings.ReplaceAll(c, "&&", "!== undefined &&")
		},
		skipIds: []int{20, 26},
	})...)

	// 8. !== undefined invalid (mutateDeclaration removes `| null`)
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "&&",
		mutateCode: func(c string) string {
			return strings.ReplaceAll(c, "&&", "!== undefined &&")
		},
		mutateDeclaration: func(c string) string {
			return strings.ReplaceAll(c, "| null", "")
		},
		mutateOutput:       identity,
		useSuggestionFixer: true,
	})...)

	// 9. != undefined invalid
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "&&",
		mutateCode: func(c string) string {
			return strings.ReplaceAll(c, "&&", "!= undefined &&")
		},
		mutateOutput:       identity,
		useSuggestionFixer: true,
	})...)

	// --- || operator group ---

	// 10. || boolean negation invalid
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "||",
		mutateCode: func(c string) string {
			return "!" + strings.ReplaceAll(c, "||", "|| !")
		},
		mutateOutput: func(c string) string {
			return "!" + c
		},
	})...)

	// 11. === null valid
	generatedValid = append(generatedValid, generateValidBaseCases(baseCaseOptions{
		operator: "||",
		mutateCode: func(c string) string {
			return strings.ReplaceAll(c, "||", "=== null ||")
		},
	})...)

	// 12. === null invalid (mutateDeclaration removes `| undefined`)
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "||",
		mutateCode: func(c string) string {
			s := strings.ReplaceAll(c, "||", "=== null ||")
			return regexp.MustCompile(`;$`).ReplaceAllString(s, " === null;")
		},
		mutateDeclaration: func(c string) string {
			return strings.ReplaceAll(c, "| undefined", "")
		},
		mutateOutput: func(c string) string {
			return regexp.MustCompile(`;$`).ReplaceAllString(c, " === null;")
		},
		useSuggestionFixer: true,
	})...)

	// 13. == null invalid
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "||",
		mutateCode: func(c string) string {
			s := strings.ReplaceAll(c, "||", "== null ||")
			return regexp.MustCompile(`;$`).ReplaceAllString(s, " == null;")
		},
		mutateOutput: func(c string) string {
			return regexp.MustCompile(`;$`).ReplaceAllString(c, " == null;")
		},
	})...)

	// 14. === undefined valid (skipIds [20, 26])
	generatedValid = append(generatedValid, generateValidBaseCases(baseCaseOptions{
		operator: "||",
		mutateCode: func(c string) string {
			return strings.ReplaceAll(c, "||", "=== undefined ||")
		},
		skipIds: []int{20, 26},
	})...)

	// 15. === undefined invalid (mutateDeclaration removes `| null`)
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "||",
		mutateCode: func(c string) string {
			s := strings.ReplaceAll(c, "||", "=== undefined ||")
			return regexp.MustCompile(`;$`).ReplaceAllString(s, " === undefined;")
		},
		mutateDeclaration: func(c string) string {
			return strings.ReplaceAll(c, "| null", "")
		},
		mutateOutput: func(c string) string {
			return regexp.MustCompile(`;$`).ReplaceAllString(c, " === undefined;")
		},
	})...)

	// 16. == undefined invalid
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "||",
		mutateCode: func(c string) string {
			s := strings.ReplaceAll(c, "||", "== undefined ||")
			return regexp.MustCompile(`;$`).ReplaceAllString(s, " == undefined;")
		},
		mutateOutput: func(c string) string {
			return regexp.MustCompile(`;$`).ReplaceAllString(c, " == undefined;")
		},
	})...)

	// 17. Whitespace sanity check
	dotWithSpaces := regexp.MustCompile(`\.`)
	bracketContent := regexp.MustCompile(`(\[.+\])`)
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "&&",
		mutateCode: func(c string) string {
			return dotWithSpaces.ReplaceAllString(c, ".      ")
		},
		mutateOutput: func(c string) string {
			return bracketContent.ReplaceAllStringFunc(c, func(m string) string {
				return dotWithSpaces.ReplaceAllString(m, ".      ")
			})
		},
	})...)

	// 18. Newline sanity check
	generatedInvalid = append(generatedInvalid, generateInvalidBaseCases(baseCaseOptions{
		operator: "&&",
		mutateCode: func(c string) string {
			return dotWithSpaces.ReplaceAllString(c, ".\n")
		},
		mutateOutput: func(c string) string {
			return bracketContent.ReplaceAllStringFunc(c, func(m string) string {
				return dotWithSpaces.ReplaceAllString(m, ".\n")
			})
		},
	})...)

	// =====================================================================
	// HAND-WRITTEN VALID CASES (edge cases not covered by generator)
	// =====================================================================
	edgeCaseValid := []rule_tester.ValidTestCase{
		// --- Already using optional chain ---
		{Code: `foo?.bar;`},
		{Code: `foo?.bar?.baz;`},
		{Code: `foo?.bar?.baz?.qux;`},

		// --- Different objects - not a chain ---
		{Code: `foo && bar.baz;`},
		{Code: `foo.bar && baz.qux;`},
		{Code: `foo && fooBar.baz;`},

		// --- Non-chain patterns ---
		{Code: `foo && bar;`},
		{Code: `foo || bar;`},
		{Code: `foo ?? bar;`},
		{Code: `foo && foo;`},
		{Code: `foo || foo.bar;`},
		{Code: `foo ?? foo.bar;`},

		// --- || {} valid patterns ---
		{Code: `foo || {};`},
		{Code: `(foo || {})?.bar;`},
		{Code: `(foo || { bar: 1 }).bar;`},
		{Code: `foo ?? {};`},
		{Code: `(foo ?? {})?.bar;`},

		// --- Comparison not forming a chain ---
		{Code: `foo == bar && foo.bar == null;`},
		{Code: `foo === 1 && foo.toFixed();`},
		{Code: `foo !== null && foo !== undefined;`},

		// --- Non-matching arguments ---
		{Code: `foo.bar(a) && foo.bar(a, b).baz;`},

		// --- Non-matching type parameters ---
		{Code: `foo.bar<a>() && foo.bar<a, b>().baz;`},

		// --- Strict null check with type that also has undefined ---
		{
			Code: "declare const foo: {bar: (() => number) | null | undefined};\nfoo.bar !== null && foo.bar();",
		},

		// --- Private identifiers (cannot use optional chaining) ---
		{Code: `foo && foo.#bar;`},
		{Code: `!foo || !foo.#bar;`},

		// --- this at start (not handled as chain) ---
		{Code: `this && this.foo;`},
		{Code: `!this || !this.foo;`},

		// --- Various non-chain patterns from official tests ---
		{Code: `!a || !b;`},
		{Code: `!a || a.b;`},
		{Code: `!a && a.b;`},
		{Code: `!a && !a.b;`},
		{Code: `!a.b || a.b?.();`},
		{Code: `!a.b || a.b();`},
		{Code: `foo ||= bar;`},
		{Code: `foo ||= bar?.baz;`},
		{Code: `result && this.options.shouldPreserveNodeMaps;`},
		{Code: `match && match$1 !== undefined;`},

		// --- Short-circuiting chains (already optional in guard) ---
		{Code: `(foo?.a).b && foo.a.b.c;`},
		{Code: `(foo?.a)() && foo.a().b;`},
		{Code: `(foo?.a)() && foo.a()();`},

		// --- Computed property mismatch ---
		{Code: `!foo[1 + 1] || !foo[1 + 2];`},
		{Code: `!foo[1 + 1] || !foo[1 + 2].foo;`},

		// --- Weird non-constant cases ---
		{Code: "({}) && {}.toString();"},
		{Code: `[] && [].length;`},

		// --- Various operator patterns ---
		{Code: "typeof foo === 'number' && foo.toFixed();"},
		{Code: "foo === 'undefined' && foo.length;"},

		// --- Assignment patterns ---
		{Code: `(x = {}) && (x.y = true) != null && x.y.toString();`},

		// --- Falsy unions: discriminated falsy union without null/undefined ---
		{Code: "declare const x: false | { a: string };\nx && x.a;"},
		{Code: "declare const x: false | { a: string };\n!x || x.a;"},
		{Code: "declare const x: '' | { a: string };\nx && x.a;"},
		{Code: "declare const x: '' | { a: string };\n!x || x.a;"},
		{Code: "declare const x: 0 | { a: string };\nx && x.a;"},
		{Code: "declare const x: 0 | { a: string };\n!x || x.a;"},
		{Code: "declare const x: 0n | { a: string };\nx && x.a;"},
		{Code: "declare const x: 0n | { a: string };\n!x || x.a;"},

		// --- import.meta / new.target valid patterns ---
		{Code: `import.meta || true;`},
		{Code: `import.meta || import.meta.foo;`},
		{Code: `!import.meta && false;`},
		{Code: `!import.meta && !import.meta.foo;`},
		{Code: `new.target || new.target.length;`},
		{Code: `!new.target || true;`},

		// --- Template literals (different instances, not a valid chain) ---
		{Code: "`x` && `x`.length;"},

		// --- Strict null check: data && data.value !== null should be valid ---
		{Code: `data && data.value !== null;`},

		// --- Strict null pair without chain ---
		{Code: `foo !== null && foo !== undefined;`},
		{Code: "x['y'] !== undefined && x['y'] !== null;"},

		// --- Private identifiers in inner chain ---
		{Code: `a.#foo?.bar;`},
		{Code: `!a.#foo?.bar;`},

		// --- Options: checkAny=false ---
		{
			Code:    "declare const x: any;\nx && x.length;",
			Options: PreferOptionalChainOptions{CheckAny: boolPtr(false)},
		},
		// --- Options: checkString=false ---
		{
			Code:    "declare const x: string;\nx && x.length;",
			Options: PreferOptionalChainOptions{CheckString: boolPtr(false)},
		},
		// --- Options: checkNumber=false ---
		{
			Code:    "declare const x: number;\nx && x.length;",
			Options: PreferOptionalChainOptions{CheckNumber: boolPtr(false)},
		},
		// --- Options: checkBoolean=false ---
		{
			Code:    "declare const x: boolean;\nx && x.length;",
			Options: PreferOptionalChainOptions{CheckBoolean: boolPtr(false)},
		},
		// --- Options: checkBigInt=false ---
		{
			Code:    "declare const x: bigint;\nx && x.length;",
			Options: PreferOptionalChainOptions{CheckBigInt: boolPtr(false)},
		},
		// --- Options: checkUnknown=false ---
		{
			Code:    "declare const x: unknown;\nx && x.length;",
			Options: PreferOptionalChainOptions{CheckUnknown: boolPtr(false)},
		},

		// --- requireNullish valid cases ---
		{
			Code:    "declare const x: string;\nx && x.length;",
			Options: PreferOptionalChainOptions{RequireNullish: boolPtr(true)},
		},
		{
			Code:    "declare const foo: string;\nfoo && foo.toString();",
			Options: PreferOptionalChainOptions{RequireNullish: boolPtr(true)},
		},
		{
			Code:    "declare const foo: { bar: string };\nfoo && foo.bar && foo.bar.toString();",
			Options: PreferOptionalChainOptions{RequireNullish: boolPtr(true)},
		},
		{
			Code:    "declare const foo: string;\n(foo || {}).toString();",
			Options: PreferOptionalChainOptions{RequireNullish: boolPtr(true)},
		},
	}

	// =====================================================================
	// HAND-WRITTEN INVALID CASES (edge cases not covered by generator)
	// =====================================================================
	edgeCaseInvalid := []rule_tester.InvalidTestCase{
		// =================================================================
		// Category 5: (foo || {}).bar patterns (→ suggestions)
		// =================================================================
		{
			Code: `(foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo?.bar;`},
					},
				},
			},
		},
		{
			Code: `(foo ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo?.bar;`},
					},
				},
			},
		},
		{
			Code: `(foo || {})[bar];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo?.[bar];`},
					},
				},
			},
		},
		{
			Code: `(foo.bar || {})[baz];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo.bar?.[baz];`},
					},
				},
			},
		},
		{
			Code: `(foo1?.foo2 || {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo1?.foo2?.foo3;`},
					},
				},
			},
		},
		{
			Code: `const foo = (bar || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `const foo = bar?.baz;`},
					},
				},
			},
		},
		{
			Code: `(foo ?? {})[baz];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo?.[baz];`},
					},
				},
			},
		},
		{
			Code: `(foo1?.foo2 ?? {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `foo1?.foo2?.foo3;`},
					},
				},
			},
		},
		{
			Code: `(this || {}).foo;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `this?.foo;`},
					},
				},
			},
		},

		// =================================================================
		// Category 2: Nullish comparisons (hand-crafted, non-base-case)
		// =================================================================
		// foo != null && foo.bar != null
		{
			Code:   `foo != null && foo.bar != null;`,
			Output: []string{`foo?.bar != null;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		// foo && foo.bar != null (truthy + nullish last operand)
		{
			Code:   "declare const foo: { bar: number } | null | undefined;\nfoo && foo.bar != null;",
			Output: []string{"declare const foo: { bar: number } | null | undefined;\nfoo?.bar != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		// foo && foo.bar != undefined
		{
			Code:   `foo && foo.bar != undefined;`,
			Output: []string{`foo?.bar != undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 3: typeof checks
		// =================================================================
		{
			Code:   "declare const foo: { bar: number } | undefined;\nfoo && typeof foo.bar !== 'undefined';",
			Output: []string{"declare const foo: { bar: number } | undefined;\ntypeof foo?.bar !== 'undefined';"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "declare const foo: { bar: number } | undefined;\nfoo && 'undefined' !== typeof foo.bar;",
			Output: []string{"declare const foo: { bar: number } | undefined;\n'undefined' !== typeof foo?.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 13: Parenthesized expressions
		// =================================================================
		{
			Code:   "a && (a.b && a.b.c)",
			Output: []string{"a?.b?.c"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Column: 1},
			},
		},
		{
			Code:   "(a && a.b) && a.b.c",
			Output: []string{"a?.b?.c"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Column: 1},
			},
		},
		{
			Code:   "((a && a.b)) && a.b.c",
			Output: []string{"a?.b?.c"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Column: 1},
			},
		},
		{
			Code:   "foo(a && (a.b && a.b.c))",
			Output: []string{"foo(a?.b?.c)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Column: 5},
			},
		},
		{
			Code:   "foo(a && a.b && a.b.c)",
			Output: []string{"foo(a?.b?.c)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Column: 5},
			},
		},
		{
			Code:   "!foo || !foo.bar || ((((!foo.bar.baz || !foo.bar.baz()))));",
			Output: []string{"!foo?.bar?.baz?.();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Column: 1},
			},
		},
		{
			Code:   "a !== undefined && ((a !== null && a.prop));",
			Output: []string{"a?.prop;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Column: 1},
			},
		},

		// =================================================================
		// Category 14: Two-error (multi-chain) cases
		// =================================================================
		{
			Code:   "foo && foo.bar && foo.bar.baz || baz && baz.bar && baz.bar.foo",
			Output: []string{"foo?.bar?.baz || baz?.bar?.foo"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo && foo.a && bar && bar.a;",
			Output: []string{"foo?.a && bar?.a;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo1 != null && foo1.bar != null && foo2 != null && foo2.bar != null;",
			Output: []string{"foo1?.bar != null && foo2?.bar != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "(!foo || !foo.bar || !foo.bar.baz) && (!baz || !baz.bar || !baz.bar.foo);",
			Output: []string{"(!foo?.bar?.baz) && (!baz?.bar?.foo);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 15: this.bar chains (not bare this)
		// =================================================================
		{
			Code:   "this.bar && this.bar.baz;",
			Output: []string{"this.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "!this.bar || !this.bar.baz;",
			Output: []string{"!this.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 16: import.meta and new.target chains
		// =================================================================
		{
			Code:   "import.meta && import.meta?.baz;",
			Output: []string{"import.meta?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "!import.meta || !import.meta?.baz;",
			Output: []string{"!import.meta?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "import.meta && import.meta?.() && import.meta?.().baz;",
			Output: []string{"import.meta?.()?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 16b: new.target chains
		// =================================================================
		{
			Code: "class Foo {\n  constructor() {\n    new.target && new.target.length;\n  }\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: "class Foo {\n  constructor() {\n    new.target?.length;\n  }\n}"},
					},
				},
			},
		},

		// =================================================================
		// Category 17: || negated chains (more patterns)
		// =================================================================
		{
			Code:   "!foo[bar] || !foo[bar]?.[baz];",
			Output: []string{"!foo[bar]?.[baz];"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "!foo || !foo?.bar.baz;",
			Output: []string{"!foo?.bar.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "!foo() || !foo().bar;",
			Output: []string{"!foo()?.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 18: Non-null assertion chains
		// =================================================================
		{
			Code:   "!foo!.bar!.baz || !foo!.bar!.baz!.paz;",
			Output: []string{"!foo!.bar.baz?.paz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "!foo.bar!.baz || !foo.bar!.baz!.paz;",
			Output: []string{"!foo.bar.baz?.paz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 19: String/template literal element access
		// =================================================================
		{
			Code:   "foo && foo['some long string'] && foo['some long string'].baz;",
			Output: []string{"foo?.['some long string']?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo && foo[`some long string`] && foo[`some long string`].baz;",
			Output: []string{"foo?.[`some long string`]?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 20: Complex computed properties
		// =================================================================
		{
			Code:   "foo && foo[1 + 2] && foo[1 + 2].baz;",
			Output: []string{"foo?.[1 + 2] && foo[1 + 2].baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo && foo[typeof bar] && foo[typeof bar].baz;",
			Output: []string{"foo?.[typeof bar] && foo[typeof bar].baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 21: Mixed binary checks (long chains)
		// =================================================================
		{
			Code:   "a &&\n  a.b != null &&\n  a.b.c !== undefined &&\n  a.b.c !== null &&\n  a.b.c.d != null &&\n  a.b.c.d.e !== null &&\n  a.b.c.d.e !== undefined &&\n  a.b.c.d.e.f != undefined &&\n  typeof a.b.c.d.e.f.g !== 'undefined' &&\n  a.b.c.d.e.f.g !== null &&\n  a.b.c.d.e.f.g.h;",
			Output: []string{"a?.b?.c?.d?.e?.f?.g?.h;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "!a ||\n  a.b == null ||\n  a.b.c === undefined ||\n  a.b.c === null ||\n  a.b.c.d == null ||\n  a.b.c.d.e === null ||\n  a.b.c.d.e === undefined ||\n  a.b.c.d.e.f == undefined ||\n  typeof a.b.c.d.e.f.g === 'undefined' ||\n  a.b.c.d.e.f.g === null ||\n  !a.b.c.d.e.f.g.h;",
			Output: []string{"!a?.b?.c?.d?.e?.f?.g?.h;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 22: Yoda checks
		// =================================================================
		{
			Code:   "undefined !== foo && null !== foo && null != foo.bar && foo.bar.baz;",
			Output: []string{"foo?.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "null != foo &&\n  'undefined' !== typeof foo.bar &&\n  null !== foo.bar &&\n  foo.bar.baz;",
			Output: []string{"foo?.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 23: foo && foo?.() patterns
		// =================================================================
		{
			Code:   "foo && foo?.();",
			Output: []string{"foo?.();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo() && foo()(bar);",
			Output: []string{"foo()?.(bar);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 24: Type parameter chains
		// =================================================================
		{
			Code:   "foo && foo<string>() && foo<string>().bar;",
			Output: []string{"foo?.<string>()?.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo && foo<string>() && foo<string, number>().bar;",
			Output: []string{"foo?.<string>() && foo<string, number>().bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		// Type reference (non-keyword) type arguments
		{
			Code:   "foo && foo<T>() && foo<T>().bar;",
			Output: []string{"foo?.<T>()?.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo && foo<T>() && foo<U>().bar;",
			Output: []string{"foo?.<T>() && foo<U>().bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 25: Chain breaking with inconsistent checks
		// =================================================================
		{
			Code:   "foo && foo.bar != null && foo.bar.baz !== undefined && foo.bar.baz.buzz;",
			Output: []string{"foo?.bar?.baz?.buzz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 26: Await expression chains
		// =================================================================
		{
			Code:   "(await foo).bar && (await foo).bar.baz;",
			Output: []string{"(await foo).bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 27: Optional chain tokens in earlier operands
		// =================================================================
		{
			Code:   "foo.bar.baz != null && foo?.bar?.baz.bam != null;",
			Output: []string{"foo?.bar?.baz?.bam != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo?.bar?.baz != null && foo.bar.baz.bam != null;",
			Output: []string{"foo.bar.baz?.bam != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 28: Non-null assertions from earlier operands
		// =================================================================
		{
			Code:   "foo.bar.baz != null && foo!.bar!.baz.bam != null;",
			Output: []string{"foo!.bar.baz?.bam != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo!.bar!.baz != null && foo.bar.baz.bam != null;",
			Output: []string{"foo.bar.baz?.bam != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 29: Unrelated prefix with chain
		// =================================================================
		{
			Code:   "unrelated != null && foo != null && foo.bar != null;",
			Output: []string{"unrelated != null && foo?.bar != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "unrelated1 != null && unrelated2 != null && foo != null && foo.bar != null;",
			Output: []string{"unrelated1 != null && unrelated2 != null && foo?.bar != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 30: globalThis typeof pattern
		// =================================================================
		{
			Code:   "function foo(globalThis?: { Array: Function }) {\n  typeof globalThis !== 'undefined' && globalThis.Array();\n}",
			Output: []string{"function foo(globalThis?: { Array: Function }) {\n  globalThis?.Array();\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 6: || negated chains (hand-crafted with declarations)
		// =================================================================
		{
			Code:   "declare const foo: {bar: {baz: number} | null | undefined};\n!foo.bar || !foo.bar.baz;",
			Output: []string{"declare const foo: {bar: {baz: number} | null | undefined};\n!foo.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "declare const a: {b: (() => number) | null | undefined};\n!a.b || !a.b();",
			Output: []string{"declare const a: {b: (() => number) | null | undefined};\n!a.b?.();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 7: Element access chains (hand-crafted with declarations)
		// =================================================================
		{
			Code:   "declare const bar: string;\ndeclare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;\nfoo && foo[bar] && foo[bar].baz && foo[bar].baz.buzz;",
			Output: []string{"declare const bar: string;\ndeclare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;\nfoo?.[bar]?.baz?.buzz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "declare const bar: {baz: string};\ndeclare const foo: {[k: string]: {buzz: number} | null | undefined} | null | undefined;\nfoo && foo[bar.baz] && foo[bar.baz].buzz;",
			Output: []string{"declare const bar: {baz: string};\ndeclare const foo: {[k: string]: {buzz: number} | null | undefined} | null | undefined;\nfoo?.[bar.baz]?.buzz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 8: this.bar chains (with declaration)
		// =================================================================
		{
			Code:   "declare const _this: {bar: {baz: number} | null | undefined};\n_this.bar && _this.bar.baz;",
			Output: []string{"declare const _this: {bar: {baz: number} | null | undefined};\n_this.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 9: Non-null expression chains
		// =================================================================
		{
			Code:   "declare const foo: {bar: {baz: number} | null | undefined};\n!foo!.bar || !foo!.bar.baz;",
			Output: []string{"declare const foo: {bar: {baz: number} | null | undefined};\n!foo!.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 10: requireNullish option
		// =================================================================
		{
			Code:    "declare const thing1: string | null;\nthing1 && thing1.toString();",
			Options: PreferOptionalChainOptions{RequireNullish: boolPtr(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: "declare const thing1: string | null;\nthing1?.toString();"},
					},
				},
			},
		},
		{
			Code:    "declare const foo: { bar: string | null | undefined } | null | undefined;\nfoo && foo.bar && foo.bar.toString();",
			Options: PreferOptionalChainOptions{RequireNullish: boolPtr(true)},
			Output:  []string{"declare const foo: { bar: string | null | undefined } | null | undefined;\nfoo?.bar?.toString();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:    "declare const foo: string | null;\n(foo || {}).toString();",
			Options: PreferOptionalChainOptions{RequireNullish: boolPtr(true)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: "declare const foo: string | null;\nfoo?.toString();"},
					},
				},
			},
		},

		// =================================================================
		// Category 11: allowPotentiallyUnsafeFixesThatModifyTheReturnType
		// =================================================================
		{
			Code:    "declare const foo: { bar: number } | null | undefined;\nfoo != undefined && foo.bar;",
			Options: PreferOptionalChainOptions{AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: boolPtr(true)},
			Output:  []string{"declare const foo: { bar: number } | null | undefined;\nfoo?.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:    "declare const foo: { bar: number } | null | undefined;\nfoo != undefined && foo.bar;",
			Options: PreferOptionalChainOptions{AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: boolPtr(false)},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: "declare const foo: { bar: number } | null | undefined;\nfoo?.bar;"},
					},
				},
			},
		},

		// =================================================================
		// Category 12: Already-optional partial chains
		// =================================================================
		{
			Code:   "declare const foo: {bar: (() => number) | null | undefined};\nfoo.bar && foo.bar?.();",
			Output: []string{"declare const foo: {bar: (() => number) | null | undefined};\nfoo.bar?.();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 1: Basic && truthy chains (hand-crafted with declarations)
		// =================================================================
		{
			Code:   `declare const foo: {bar: number} | null | undefined; foo && foo.bar;`,
			Output: []string{`declare const foo: {bar: number} | null | undefined; foo?.bar;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   `declare const foo: {bar: {baz: number} | null | undefined}; foo.bar && foo.bar.baz;`,
			Output: []string{`declare const foo: {bar: {baz: number} | null | undefined}; foo.bar?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   `declare const foo: {bar: {baz: number} | null | undefined} | null | undefined; foo && foo.bar && foo.bar.baz;`,
			Output: []string{`declare const foo: {bar: {baz: number} | null | undefined} | null | undefined; foo?.bar?.baz;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;\nfoo && foo.bar && foo.bar.baz && foo.bar.baz.buzz;",
			Output: []string{"declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;\nfoo?.bar?.baz?.buzz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "declare const foo: (() => number) | null | undefined;\nfoo && foo();",
			Output: []string{"declare const foo: (() => number) | null | undefined;\nfoo?.();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   `declare const foo: {bar: (() => number) | null | undefined}; foo.bar && foo.bar();`,
			Output: []string{`declare const foo: {bar: (() => number) | null | undefined}; foo.bar?.();`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 2b: Nullish comparisons with declarations (suggestion-only)
		// =================================================================
		{
			Code: `declare const foo: {bar: number} | null; foo != null && foo.bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `declare const foo: {bar: number} | null; foo?.bar;`},
					},
				},
			},
		},
		{
			Code: `declare const foo: {bar: number} | undefined; foo !== undefined && foo.bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `declare const foo: {bar: number} | undefined; foo?.bar;`},
					},
				},
			},
		},

		// =================================================================
		// Category 3b: typeof checks with declarations (suggestion-only)
		// =================================================================
		{
			Code: `declare const foo: {bar: number} | undefined; typeof foo !== 'undefined' && foo.bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `declare const foo: {bar: number} | undefined; foo?.bar;`},
					},
				},
			},
		},

		// =================================================================
		// Category 4: Yoda conditions (suggestion-only)
		// =================================================================
		{
			Code: `declare const foo: {bar: number} | null; null != foo && foo.bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "optionalChainSuggest", Output: `declare const foo: {bar: number} | null; foo?.bar;`},
					},
				},
			},
		},
	}

	// Combine generated and hand-written cases
	allValid := append(generatedValid, edgeCaseValid...)
	allInvalid := append(generatedInvalid, edgeCaseInvalid...)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		allValid, allInvalid)
}
