package prefer_optional_chain

import (
	"regexp"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
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
	// Skip ID 20: call chain with trailing call + strict undefined equality
	// triggers partial fold (matches TS-ESLint's hand-crafted test cases).
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
		skipIds:            []int{20},
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
			Options: PreferOptionalChainOptions{CheckAny: utils.Ref(false)},
		},
		// --- Options: checkString=false ---
		{
			Code:    "declare const x: string;\nx && x.length;",
			Options: PreferOptionalChainOptions{CheckString: utils.Ref(false)},
		},
		// --- Options: checkNumber=false ---
		{
			Code:    "declare const x: number;\nx && x.length;",
			Options: PreferOptionalChainOptions{CheckNumber: utils.Ref(false)},
		},
		// --- Options: checkBoolean=false ---
		{
			Code:    "declare const x: boolean;\nx && x.length;",
			Options: PreferOptionalChainOptions{CheckBoolean: utils.Ref(false)},
		},
		// --- Options: checkBigInt=false ---
		{
			Code:    "declare const x: bigint;\nx && x.length;",
			Options: PreferOptionalChainOptions{CheckBigInt: utils.Ref(false)},
		},
		// --- Options: checkUnknown=false ---
		{
			Code:    "declare const x: unknown;\nx && x.length;",
			Options: PreferOptionalChainOptions{CheckUnknown: utils.Ref(false)},
		},

		// --- requireNullish valid cases ---
		{
			Code:    "declare const x: string;\nx && x.length;",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
		},
		{
			Code:    "declare const foo: string;\nfoo && foo.toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
		},
		{
			Code:    "declare const foo: { bar: string };\nfoo && foo.bar && foo.bar.toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
		},
		{
			Code:    "declare const foo: string;\n(foo || {}).toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
		},
		// --- Additional valid patterns (false positive prevention) ---
		{Code: `foo || ({} as any);`},
		{Code: `foo ||= bar || {};`},
		{Code: `foo ||= bar?.baz?.buzz;`},
		{Code: "file !== 'index.ts' && file.endsWith('.ts');"},
		{Code: `!foo().#a || a;`},
		{Code: `!a.b.#a || a;`},
		{Code: "!(foo as any).bar || 'anything';"},
		{Code: "(() => {}) && (() => {}).name;"},
		{Code: "(function () {}) && function () {}.name;"},
		{Code: "new Map().get('a') && new Map().get('a').what;"},
		{Code: `foo[x++] && foo[x++].bar;`},
		{Code: `a = b && (a = b).wtf;`},
		{Code: `(x || y) != null && (x || y).foo;`},
		{Code: `(await foo) && (await foo).bar;`},
		// Falsy literal unions (discriminated unions, not null guards)
		{Code: "declare const x: false | { a: string };\nx && x.a;"},
		{Code: "declare const x: '' | { a: string };\nx && x.a;"},
		{Code: "declare const x: 0 | { a: string };\nx && x.a;"},
		{Code: "declare const x: 0n | { a: string };\nx && x.a;"},
		// || {} / ?? {} valid patterns
		{Code: `(undefined && (foo || {})).bar;`},
		{Code: `(foo1 ? foo2 : foo3 || {}).foo4;`},
		{Code: `(foo = 2 || {}).bar;`},
		{Code: `func(foo || {}).bar;`},
		{Code: `foo ||= bar ?? {};`},
		// Private identifiers
		{Code: `this.#a && this.#b;`},
		{Code: `!this.#a || !this.#b;`},
		{Code: `!new A().#b || a;`},
		{Code: `!(await a).#b || a;`},
		// Various non-chain patterns
		{Code: `nextToken && sourceCode.isSpaceBetweenTokens(prevToken, nextToken);`},
		{Code: `!entity.__helper!.__initialized || options.refresh;`},
		{Code: `[1, 2].length && [1, 2, 3].length.toFixed();`},
		{Code: `(class Foo {}) && class Foo {}.constructor;`},
		{Code: `foo[yield x] && foo[yield x].bar;`},
		// Template literal / type assertion patterns
		{Code: "`x${a}` && `x${a}`.length;"},
		{Code: "('x' as `${'x'}`) && ('x' as `${'x'}`).length;"},
		// JSX valid patterns
		{Code: `<div /> && (<div />).wtf;`, Tsx: true},
		{Code: `<></> && (<></>).wtf;`, Tsx: true},
		// globalThis typeof (valid: no chain formed)
		{Code: "typeof globalThis !== 'undefined' && globalThis.Array();"},
		// Issue #8380: cross-variable null/undefined checks (not optional chain candidates)
		{Code: "const a = null;\nconst b = 0;\na === undefined || b === null || b === undefined;"},
		{Code: "const a = 0;\nconst b = 0;\na === undefined || b === undefined || b === null;"},
		{Code: "const b = 0;\nb === null || b === undefined;"},
		{Code: "const a = 0;\nconst b = 0;\nb === null || a === undefined || b === undefined;"},
		{Code: "const a = 0;\nconst b = 0;\nb != null && a !== null && a !== undefined;"},
		{Code: `foo ||= bar?.baz || {};`},
		{Code: "[1,].length && [1, 2].length.toFixed();"},
		// requireNullish additional valid
		{
			Code:    "declare const x: string | number | boolean | object;\nx && x.toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
		},
		{
			Code:    "declare const foo: string;\nfoo && foo.toString() && foo.toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
		},
		{
			Code:    "declare const foo: { bar: string };\nfoo && foo.bar && foo.bar.toString() && foo.bar.toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
		},
		{
			Code:    "declare const foo1: { bar: string | null };\nfoo1 && foo1.bar;",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
		},
		{
			Code:    "declare const foo: string | null;\n(foo || 'a' || {}).toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
		},

		// --- Diverging property access paths (same terminal name, different chains) ---
		// foo.x and foo.y.x share the terminal name "x" but are different access paths
		{Code: `foo.x && foo.y.x;`},
		{Code: `!foo.x || !foo.y.x;`},
		{Code: "declare const entry: {success: boolean; result?: {success: boolean}};\nentry.success && entry.result?.success;"},
		// a.bar and a.baz.bar share terminal name "bar"
		{Code: `a.bar && a.baz.bar;`},
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
			Output: []string{"!foo!.bar!.baz?.paz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "!foo.bar!.baz || !foo.bar!.baz!.paz;",
			Output: []string{"!foo.bar!.baz?.paz;"},
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
			Output: []string{"foo?.[1 + 2]?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo && foo[typeof bar] && foo[typeof bar].baz;",
			Output: []string{"foo?.[typeof bar]?.baz;"},
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
		// Inconsistent checks (loose != null then strict !== undefined) break chain
		{
			Code:   "foo && foo.bar != null && foo.bar.baz !== undefined && foo.bar.baz.buzz;",
			Output: []string{"foo?.bar != null && foo.bar.baz !== undefined && foo.bar.baz.buzz;"},
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
		// Existing ?. from guard operands are preserved in output by scanning
		// the guard's source text via GetSourceTextOfNodeFromSourceFile.
		{
			Code:   "foo.bar.baz != null && foo?.bar?.baz.bam != null;",
			Output: []string{"foo.bar.baz?.bam != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo?.bar.baz != null && foo.bar?.baz.bam != null;",
			Output: []string{"foo?.bar.baz?.bam != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo?.bar?.baz != null && foo.bar.baz.bam != null;",
			Output: []string{"foo?.bar?.baz?.bam != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 28: Non-null assertions from earlier operands
		// =================================================================
		// Non-null assertions: the output uses the GUARD's text for the base portion,
		// preserving ! assertions from the guard, not the target.
		{
			Code:   "foo!.bar.baz != null && foo.bar!.baz.bam != null;",
			Output: []string{"foo!.bar.baz?.bam != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:   "foo!.bar!.baz != null && foo.bar.baz.bam != null;",
			Output: []string{"foo!.bar!.baz?.bam != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},

		// =================================================================
		// Category 29: Unrelated prefix with chain (with Column assertions)
		// =================================================================
		{
			Code:   "unrelated != null && foo != null && foo.bar != null;",
			Output: []string{"unrelated != null && foo?.bar != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Column: 22},
			},
		},
		{
			Code:   "unrelated1 != null && unrelated2 != null && foo != null && foo.bar != null;",
			Output: []string{"unrelated1 != null && unrelated2 != null && foo?.bar != null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Column: 45},
			},
		},

		// =================================================================
		// Category 31: Complementary strict-equality pair recovery
		// =================================================================
		// Truthy guard + complementary pair on deeper property → merge to != null
		{
			Code: "declare const existing: {id: number | null | undefined} | null | undefined;\nexisting && existing.id !== null && existing.id !== undefined;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    "declare const existing: {id: number | null | undefined} | null | undefined;\nexisting?.id != null;",
						},
					},
				},
			},
		},
		// Same pattern with || chain (DeMorgan)
		{
			Code: "declare const existing: {id: number | null | undefined} | null | undefined;\n!existing || existing.id === null || existing.id === undefined;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalChain",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "optionalChainSuggest",
							Output:    "declare const existing: {id: number | null | undefined} | null | undefined;\nexisting?.id == null;",
						},
					},
				},
			},
		},
		// Reversed order: !== undefined first, then !== null.
		// !== undefined doesn't trigger wouldChangeTruthiness, so the chain
		// proceeds as a normal chain-with-tail (no complementary merge needed).
		{
			Code:   "declare const data: {value: string | null | undefined} | null | undefined;\ndata && data.value !== undefined && data.value !== null;",
			Output: []string{"declare const data: {value: string | null | undefined} | null | undefined;\ndata?.value !== undefined && data.value !== null;"},
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
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
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
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
			Output:  []string{"declare const foo: { bar: string | null | undefined } | null | undefined;\nfoo?.bar?.toString();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:    "declare const foo: string | null;\n(foo || {}).toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
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
			Options: PreferOptionalChainOptions{AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: utils.Ref(true)},
			Output:  []string{"declare const foo: { bar: number } | null | undefined;\nfoo?.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain"},
			},
		},
		{
			Code:    "declare const foo: { bar: number } | null | undefined;\nfoo != undefined && foo.bar;",
			Options: PreferOptionalChainOptions{AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: utils.Ref(false)},
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
		// =================================================================
		// Additional || {} / ?? {} patterns (from TS test suite)
		// =================================================================
		{
			Code: `(foo || ({})).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `foo?.bar;`},
				}},
			},
		},
		{
			Code: `(await foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `(await foo)?.bar;`},
				}},
			},
		},
		{
			Code: `(foo || undefined || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `(foo || undefined)?.bar;`},
				}},
			},
		},
		{
			Code: `(foo() || bar || {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `(foo() || bar)?.baz;`},
				}},
			},
		},
		{
			Code: `((foo1 ? foo2 : foo3) || {}).foo4;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `(foo1 ? foo2 : foo3)?.foo4;`},
				}},
			},
		},
		// ?? {} variants
		{
			Code: `(foo ?? ({})).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `foo?.bar;`},
				}},
			},
		},
		{
			Code: `(await foo ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `(await foo)?.bar;`},
				}},
			},
		},
		{
			Code: `(foo.bar ?? {})[baz];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `foo.bar?.[baz];`},
				}},
			},
		},

		// =================================================================
		// Additional && chain patterns
		// =================================================================
		// Arrow with typeof in argument
		{
			Code:   "foo && foo.bar(baz => typeof baz);",
			Output: []string{"foo?.bar(baz => typeof baz);"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// JSX whitespace preservation
		{
			Code:   "foo && foo.bar(baz => <This Requires Spaces />);",
			Tsx:    true,
			Output: []string{"foo?.bar(baz => <This Requires Spaces />);"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Template literal with interpolation in computed property
		{
			Code:   "foo && foo[`some ${long} string`] && foo[`some ${long} string`].baz;",
			Output: []string{"foo?.[`some ${long} string`]?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Type assertion in computed property
		{
			Code:   "foo && foo[bar as string] && foo[bar as string].baz;",
			Output: []string{"foo?.[bar as string]?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Mismatched arguments break chain
		{
			Code:   "foo && foo.bar(a) && foo.bar(a, b).baz;",
			Output: []string{"foo?.bar(a) && foo.bar(a, b).baz;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Comments in call expression preserved
		{
			Code:   "foo && foo.bar(/* comment */a,\n  // comment2\n  b);",
			Output: []string{"foo?.bar(/* comment */a,\n  // comment2\n  b);"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// foo && foo.bar != null && baz (trailing unrelated)
		{
			Code:   "foo && foo.bar != null && baz;",
			Output: []string{"foo?.bar != null && baz;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// foo.bar && foo.bar?.() without declaration
		{
			Code:   "foo.bar && foo.bar?.();",
			Output: []string{"foo.bar?.();"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// !foo!.bar || !foo!.bar.baz without declaration
		{
			Code:   "!foo!.bar || !foo!.bar.baz;",
			Output: []string{"!foo!.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Type parameters: matching
		{
			Code:   "foo && foo<string>() && foo<string>().bar;",
			Output: []string{"foo?.<string>()?.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Type parameters: mismatching stops chain
		{
			Code:   "foo && foo<string>() && foo<string, number>().bar;",
			Output: []string{"foo?.<string>() && foo<string, number>().bar;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Type parameters: reference type
		{
			Code:   "foo && foo<T>() && foo<T>().bar;",
			Output: []string{"foo?.<T>()?.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// =================================================================
		// Deep mixed checks (long chains with mixed operators)
		// =================================================================
		{
			Code: "a &&\n  a.b != null &&\n  a.b.c !== undefined &&\n  a.b.c !== null &&\n  a.b.c.d != null &&\n  a.b.c.d.e !== null &&\n  a.b.c.d.e !== undefined &&\n  a.b.c.d.e.f != undefined &&\n  typeof a.b.c.d.e.f.g !== 'undefined' &&\n  a.b.c.d.e.f.g !== null &&\n  a.b.c.d.e.f.g.h;",
			Output: []string{"a?.b?.c?.d?.e?.f?.g?.h;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// || chain variant
		{
			Code:   "!a ||\n  a.b == null ||\n  a.b.c === undefined ||\n  a.b.c === null ||\n  a.b.c.d == null ||\n  a.b.c.d.e === null ||\n  a.b.c.d.e === undefined ||\n  a.b.c.d.e.f == undefined ||\n  typeof a.b.c.d.e.f.g === 'undefined' ||\n  a.b.c.d.e.f.g === null ||\n  !a.b.c.d.e.f.g.h;",
			Output: []string{"!a?.b?.c?.d?.e?.f?.g?.h;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// =================================================================
		// Yoda with typeof
		// =================================================================
		{
			Code:   "undefined !== foo && null !== foo && null != foo.bar && foo.bar.baz;",
			Output: []string{"foo?.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   "null != foo &&\n  'undefined' !== typeof foo.bar &&\n  null !== foo.bar &&\n  foo.bar.baz;",
			Output: []string{"foo?.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// =================================================================
		// Additional || {} patterns (IIFE, nested, complex LHS)
		// =================================================================
		{
			Code: `((() => foo())() || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(() => foo())()?.bar;`},
			}}},
		},
		{
			Code: `((foo1 || {}).foo2 || {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `(foo1 || {}).foo2?.foo3;`},
				}},
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `(foo1?.foo2 || {}).foo3;`},
				}},
			},
		},
		{
			Code: `(undefined && foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(undefined && foo)?.bar;`},
			}}},
		},
		{
			Code: `(a > b || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(a > b)?.bar;`},
			}}},
		},
		{
			Code: `(void foo() || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(void foo())?.bar;`},
			}}},
		},
		{
			Code: `((a instanceof Error) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(a instanceof Error)?.bar;`},
			}}},
		},
		{
			Code: `((a << b) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(a << b)?.bar;`},
			}}},
		},
		{
			Code: `((foo ** 2) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(foo ** 2)?.bar;`},
			}}},
		},
		{
			Code: `(foo ** 2 || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(foo ** 2)?.bar;`},
			}}},
		},
		{
			Code: `(foo++ || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(foo++)?.bar;`},
			}}},
		},
		{
			Code: `(+foo || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(+foo)?.bar;`},
			}}},
		},

		// =================================================================
		// Additional ?? {} patterns
		// =================================================================
		{
			Code: `((() => foo())() ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(() => foo())()?.bar;`},
			}}},
		},
		{
			Code: `const foo = (bar ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `const foo = bar?.baz;`},
			}}},
		},
		{
			Code: `((foo1 ?? {}).foo2 ?? {}).foo3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `(foo1 ?? {}).foo2?.foo3;`},
				}},
				{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "optionalChainSuggest", Output: `(foo1?.foo2 ?? {}).foo3;`},
				}},
			},
		},
		{
			Code: `(foo ?? undefined ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(foo ?? undefined)?.bar;`},
			}}},
		},
		{
			Code: `(foo() ?? bar ?? {}).baz;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(foo() ?? bar)?.baz;`},
			}}},
		},
		{
			Code: `((foo1 ? foo2 : foo3) ?? {}).foo4;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(foo1 ? foo2 : foo3)?.foo4;`},
			}}},
		},

		// =================================================================
		// Additional bare && chains (no declarations)
		// =================================================================
		{
			Code:   "foo && foo.bar != null;",
			Output: []string{"foo?.bar != null;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   "!foo.bar || !foo.bar.baz;",
			Output: []string{"!foo.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   "!a.b || !a.b();",
			Output: []string{"!a.b?.();"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   "foo.bar.baz != null && foo!.bar!.baz.bam != null;",
			Output: []string{"foo.bar.baz?.bam != null;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// =================================================================
		// Chain-breaking: multi-line inconsistent check
		// =================================================================
		{
			Code:   "foo.bar &&\n  foo.bar.baz != null &&\n  foo.bar.baz.qux !== undefined &&\n  foo.bar.baz.qux.buzz;",
			Output: []string{"foo.bar?.baz != null &&\n  foo.bar.baz.qux !== undefined &&\n  foo.bar.baz.qux.buzz;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// =================================================================
		// || chain: reordered null/undefined variant
		// =================================================================
		{
			Code:   "!a ||\n  a.b == null ||\n  a.b.c === null ||\n  a.b.c === undefined ||\n  a.b.c.d == null ||\n  a.b.c.d.e === null ||\n  a.b.c.d.e === undefined ||\n  a.b.c.d.e.f == undefined ||\n  typeof a.b.c.d.e.f.g === 'undefined' ||\n  a.b.c.d.e.f.g === null ||\n  !a.b.c.d.e.f.g.h;",
			Output: []string{"!a?.b?.c?.d?.e?.f?.g?.h;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// =================================================================
		// Strict null checks with declarations
		// =================================================================
		{
			Code: "declare const foo: { bar: string } | null;\nfoo !== null && foo.bar !== null;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "declare const foo: { bar: string } | null;\nfoo?.bar !== null;"},
			}}},
		},
		{
			Code:   "declare const foo: { bar: string | null } | null;\nfoo !== null && foo.bar != null;",
			Output: []string{"declare const foo: { bar: string | null } | null;\nfoo?.bar != null;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// typeof on bare identifier preserves the typeof prefix, chain starts after
		{
			Code: "typeof globalThis !== 'undefined' && globalThis.Array && globalThis.Array();",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "typeof globalThis !== 'undefined' && globalThis.Array?.();"},
			}}},
		},

		// =================================================================
		// allowPotentiallyUnsafe with acceptsBoolean
		// =================================================================
		{
			Code:    "declare const foo: { bar: boolean } | null | undefined;\ndeclare function acceptsBoolean(arg: boolean): void;\nacceptsBoolean(foo != null && foo.bar);",
			Options: PreferOptionalChainOptions{AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: utils.Ref(true)},
			Output:  []string{"declare const foo: { bar: boolean } | null | undefined;\ndeclare function acceptsBoolean(arg: boolean): void;\nacceptsBoolean(foo?.bar);"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// =================================================================
		// requireNullish additional invalid
		// =================================================================
		{
			Code:    "declare const thing1: string | null;\nthing1 && thing1.toString() && true;",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "declare const thing1: string | null;\nthing1?.toString() && true;"},
			}}},
		},
		// Chains with duplicate last operands stop before the duplicate
		{
			Code:    "declare const foo: string | null;\nfoo && foo.toString() && foo.toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "declare const foo: string | null;\nfoo?.toString() && foo.toString();"},
			}}},
		},
		{
			Code:    "declare const foo: { bar: string | null | undefined } | null | undefined;\nfoo && foo.bar && foo.bar.toString() && foo.bar.toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
			Output:  []string{"declare const foo: { bar: string | null | undefined } | null | undefined;\nfoo?.bar?.toString() && foo.bar.toString();"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// =================================================================
		// || {} / ?? {} inside if statements
		// =================================================================
		{
			Code: "(foo1?.foo2 || ({})).foo3;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "foo1?.foo2?.foo3;"},
			}}},
		},
		{
			Code: "if (foo) {\n  (foo || {}).bar;\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "if (foo) {\n  foo?.bar;\n}"},
			}}},
		},
		{
			Code: "if ((foo || {}).bar) {\n  foo.bar;\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "if (foo?.bar) {\n  foo.bar;\n}"},
			}}},
		},
		{
			Code: `if (foo) { (foo ?? {}).bar; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `if (foo) { foo?.bar; }`},
			}}},
		},
		{
			Code: `if ((foo ?? {}).bar) { foo.bar; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `if (foo?.bar) { foo.bar; }`},
			}}},
		},
		{
			Code: `(undefined && foo ?? {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(undefined && foo)?.bar;`},
			}}},
		},
		{
			Code: `(((typeof x) as string) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `((typeof x) as string)?.bar;`},
			}}},
		},
		{
			Code: `((a ? b : c) || {}).bar;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: `(a ? b : c)?.bar;`},
			}}},
		},

		// =================================================================
		// Nullish comparison with declarations
		// =================================================================
		{
			Code: "declare const foo: { bar: string | null } | null;\nfoo != null && foo.bar !== null;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "declare const foo: { bar: string | null } | null;\nfoo?.bar !== null;"},
			}}},
		},

		// =================================================================
		// Yoda: null != foo?.bar?.baz
		// =================================================================
		{
			Code:   "null != foo &&\n  'undefined' !== typeof foo.bar &&\n  null !== foo.bar &&\n  null != foo.bar.baz;",
			Output: []string{"null != foo?.bar?.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// =================================================================
		// Retain split strict equals at end of chain
		// =================================================================
		{
			Code: "null != foo &&\n  'undefined' !== typeof foo.bar &&\n  null !== foo.bar &&\n  null !== foo.bar.baz &&\n  'undefined' !== typeof foo.bar.baz;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "null !== foo?.bar?.baz &&\n  'undefined' !== typeof foo.bar.baz;"},
			}}},
		},
		{
			Code: "foo != null &&\n  typeof foo.bar !== 'undefined' &&\n  foo.bar !== null &&\n  foo.bar.baz !== null &&\n  typeof foo.bar.baz !== 'undefined';",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "foo?.bar?.baz !== null &&\n  typeof foo.bar.baz !== 'undefined';"},
			}}},
		},
		{
			Code: "null != foo &&\n  'undefined' !== typeof foo.bar &&\n  null !== foo.bar &&\n  null !== foo.bar.baz &&\n  undefined !== foo.bar.baz;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "null !== foo?.bar?.baz &&\n  undefined !== foo.bar.baz;"},
			}}},
		},
		{
			Code: "foo != null &&\n  typeof foo.bar !== 'undefined' &&\n  foo.bar !== null &&\n  foo.bar.baz !== null &&\n  foo.bar.baz !== undefined;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "foo?.bar?.baz !== null &&\n  foo.bar.baz !== undefined;"},
			}}},
		},
		{
			Code:   "null != foo &&\n  'undefined' !== typeof foo.bar &&\n  null !== foo.bar &&\n  undefined !== foo.bar.baz &&\n  null !== foo.bar.baz;",
			Output: []string{"undefined !== foo?.bar?.baz &&\n  null !== foo.bar.baz;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   "foo != null &&\n  typeof foo.bar !== 'undefined' &&\n  foo.bar !== null &&\n  foo.bar.baz !== undefined &&\n  foo.bar.baz !== null;",
			Output: []string{"foo?.bar?.baz !== undefined &&\n  foo.bar.baz !== null;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// =================================================================
		// || chain: == null before final operand
		// =================================================================
		{
			Code:   "!a ||\n  a.b == null ||\n  a.b.c === undefined ||\n  a.b.c === null ||\n  a.b.c.d == null ||\n  a.b.c.d.e === null ||\n  a.b.c.d.e === undefined ||\n  a.b.c.d.e.f == undefined ||\n  a.b.c.d.e.f.g == null ||\n  a.b.c.d.e.f.g.h;",
			Output: []string{"a?.b?.c?.d?.e?.f?.g == null ||\n  a.b.c.d.e.f.g.h;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// =================================================================
		// requireNullish: (foo || undefined || {}).toString()
		// =================================================================
		{
			Code:    "declare const foo: string;\n(foo || undefined || {}).toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "declare const foo: string;\n(foo || undefined)?.toString();"},
			}}},
		},
		{
			Code:    "declare const foo: string | null;\n(foo || undefined || {}).toString();",
			Options: PreferOptionalChainOptions{RequireNullish: utils.Ref(true)},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "declare const foo: string | null;\n(foo || undefined)?.toString();"},
			}}},
		},

		// =================================================================
		// Extra: !== undefined / === undefined call chains
		// When the last operand is a bare call/negation and the chain includes
		// strict equality guards on call results, trim to avoid changing call count.
		// =================================================================
		// Cases with !== undefined / === undefined call chains.
		// Aligned with TS-ESLint: strict equality on call expression result stops
		// optional chain extension because each call is a separate invocation.
		// Case 1: trailing bare call → partial fold (suggestion)
		{
			Code: "declare const foo: {bar: () => ({baz: {buzz: (() => number) | undefined} | undefined}) | undefined};\nfoo.bar !== undefined &&\n  foo.bar() !== undefined &&\n  foo.bar().baz !== undefined &&\n  foo.bar().baz.buzz !== undefined &&\n  foo.bar().baz.buzz();",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain", Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "optionalChainSuggest", Output: "declare const foo: {bar: () => ({baz: {buzz: (() => number) | undefined} | undefined}) | undefined};\nfoo.bar?.() !== undefined &&\n  foo.bar().baz !== undefined &&\n  foo.bar().baz.buzz !== undefined &&\n  foo.bar().baz.buzz();"},
			}}},
		},
		// Case 2: trailing comparison (auto-fix: !== undefined preserves boolean result)
		{
			Code:   "declare const foo: {bar: () => ({baz: {buzz: (() => number) | undefined} | undefined}) | undefined};\nfoo.bar !== undefined &&\n  foo.bar() !== undefined &&\n  foo.bar().baz !== undefined &&\n  foo.bar().baz.buzz !== undefined;",
			Output: []string{"declare const foo: {bar: () => ({baz: {buzz: (() => number) | undefined} | undefined}) | undefined};\nfoo.bar?.()?.baz?.buzz !== undefined;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Case 3: || trailing negated call → partial fold (auto-fix)
		{
			Code:   "declare const foo: {bar: () => ({baz: {buzz: (() => number) | undefined} | undefined}) | undefined};\nfoo.bar === undefined ||\n  foo.bar() === undefined ||\n  foo.bar().baz === undefined ||\n  foo.bar().baz.buzz === undefined ||\n  !foo.bar().baz.buzz();",
			Output: []string{"declare const foo: {bar: () => ({baz: {buzz: (() => number) | undefined} | undefined}) | undefined};\nfoo.bar?.() === undefined ||\n  foo.bar().baz === undefined ||\n  foo.bar().baz.buzz === undefined ||\n  !foo.bar().baz.buzz();"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		// Case 4: || trailing comparison (auto-fix, full fold)
		{
			Code:   "declare const foo: {bar: () => ({baz: {buzz: (() => number) | undefined} | undefined}) | undefined};\nfoo.bar === undefined ||\n  foo.bar() === undefined ||\n  foo.bar().baz === undefined ||\n  foo.bar().baz.buzz === undefined;",
			Output: []string{"declare const foo: {bar: () => ({baz: {buzz: (() => number) | undefined} | undefined}) | undefined};\nfoo.bar?.()?.baz?.buzz === undefined;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	}

	// Combine generated and hand-written cases
	allValid := append(generatedValid, edgeCaseValid...)
	allInvalid := append(generatedInvalid, edgeCaseInvalid...)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule,
		allValid, allInvalid)
}
