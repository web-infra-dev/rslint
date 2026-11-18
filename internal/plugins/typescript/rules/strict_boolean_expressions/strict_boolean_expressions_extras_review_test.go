package strict_boolean_expressions

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// reviewError specifies the authored range independently of the rule's AST.
func reviewError(t *testing.T, code, operand, id, message string, suggestions []rule_tester.InvalidTestCaseSuggestion) rule_tester.InvalidTestCaseError {
	t.Helper()
	start := strings.LastIndex(code, operand)
	if start < 0 {
		t.Fatalf("missing diagnostic operand %q", operand)
	}
	prefix := code[:start]
	line := strings.Count(prefix, "\n") + 1
	column := start - strings.LastIndex(prefix, "\n")
	return rule_tester.InvalidTestCaseError{MessageId: id, Message: message, Line: line, Column: column, EndLine: line, EndColumn: column + len(operand), Suggestions: suggestions}
}

func TestStrictBooleanExpressionsReviewSafeLength(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, test := range []struct {
		declaration, expression string
		negated                 bool
	}{
		{`declare const user: {name: string} | undefined;`, `user?.name!`, false},
		{`declare const user: {name: string} | undefined;`, `user?.name!`, true},
		{`declare const user: {name: string} | undefined;`, `(user?.name!) as string`, false},
		{`declare const user: {name: string} | undefined;`, `(user?.name!) as string`, true},
		{`declare const user: {getName(): string} | undefined;`, `user?.getName()!`, false},
		{`declare const user: {getName(): string} | undefined;`, `user?.getName()!`, true},
		{`declare const getUser: () => {name: string} | undefined;`, `getUser()?.name!`, false},
		{`declare const getUser: () => {name: string} | undefined;`, `getUser()?.name!`, true},
		{`declare const user: {name: string} | undefined;`, `user!?.name`, false},
		{`declare const user: {name: string} | undefined;`, `user!?.name`, true},
	} {
		target := test.expression
		if test.negated {
			target = "!(" + target + ")"
		}
		code := test.declaration + " if (" + target + ") {}"
		inner := "(" + test.expression + ")"
		length := "(" + inner + ` ?? "").length > 0`
		compare := inner + ` !== ""`
		cast := "Boolean(" + inner + ")"
		if test.negated {
			length = "(" + inner + ` ?? "").length === 0`
			compare = inner + ` === ""`
			cast = "!Boolean(" + inner + ")"
		}
		replace := func(text string) string { return strings.Replace(code, target, text, 1) }
		suggestions := []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: "conditionFixCompareStringLength", Output: replace(length)},
			{MessageId: "conditionFixCompareEmptyString", Output: replace(compare)},
			{MessageId: "conditionFixCastBoolean", Output: replace(cast)},
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: map[string]any{"allowString": false}, Errors: []rule_tester.InvalidTestCaseError{reviewError(t, code, test.expression, "conditionErrorString", "Unexpected string value in conditional. An explicit empty string check is required.", suggestions)}})
	}
	for _, negated := range []bool{false, true} {
		expression := `(await user?.getName()!)!`
		target := expression
		if negated {
			target = "!(" + expression + ")"
		}
		code := `declare const user: {getName(): Promise<string>} | undefined; async function f() { if (` + target + `) {} }`
		length := "(" + expression + ` ?? "").length > 0`
		compare := expression + ` !== ""`
		cast := "Boolean(" + expression + ")"
		if negated {
			length = "(" + expression + ` ?? "").length === 0`
			compare = expression + ` === ""`
			cast = "!Boolean(" + expression + ")"
		}
		suggestions := []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: "conditionFixCompareStringLength", Output: strings.Replace(code, target, length, 1)},
			{MessageId: "conditionFixCompareEmptyString", Output: strings.Replace(code, target, compare, 1)},
			{MessageId: "conditionFixCastBoolean", Output: strings.Replace(code, target, cast, 1)},
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: map[string]any{"allowString": false}, Errors: []rule_tester.InvalidTestCaseError{reviewError(t, code, expression, "conditionErrorString", "Unexpected string value in conditional. An explicit empty string check is required.", suggestions)}})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, nil, invalid)
}

func TestStrictBooleanExpressionsReviewSwitchASI(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, test := range []struct {
		before, after  string
		needsSemicolon bool
	}{
		{"switch (1) { case 1: record()\n", "\nbreak; }", true},
		{"switch (1) { default: record()\n", "\nbreak; }", true},
		{"switch (1) { case 0: case 1: record()\n", "\nbreak; }", true},
		{"switch (1) { case 1: record() // comment\n", "\nbreak; }", true},
		{"switch (1) { case 1: record()\r\n", "\r\nbreak; }", true},
		{"switch (1) { case 1: record(); /* comment */\n", "\nbreak; }", false},
		{"switch (1) { case 1: ", "\nbreak; }", false},
		{"switch (1) { case 0: record()\ncase 1: ", "\nbreak; }", false},
		{"record()\n", ";", true},
		{"{ record()\n", "; }", true},
	} {
		prefix := "declare const state: {active?: boolean} | undefined; declare function record(): void; "
		code := prefix + test.before + "!state?.active" + test.after
		compare := "(state?.active) === false"
		if test.needsSemicolon {
			compare = ";" + compare
		}
		suggestions := []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: "conditionFixDefaultFalse", Output: strings.Replace(code, "!state?.active", "!((state?.active) ?? false)", 1)},
			{MessageId: "conditionFixCompareFalse", Output: strings.Replace(code, "!state?.active", compare, 1)},
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{reviewError(t, code, "state?.active", "conditionErrorNullableBoolean", "Unexpected nullable boolean value in conditional. Please handle the nullish case explicitly.", suggestions)}})
	}
	for _, tail := range []string{" ? 1 : 0", " || true", " && true"} {
		code := "declare const state: {active?: boolean} | undefined; declare function record(): void; switch (1) { case 1: record()\n!state?.active" + tail + "\nbreak; }"
		suggestions := []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: "conditionFixDefaultFalse", Output: strings.Replace(code, "!state?.active", "!((state?.active) ?? false)", 1)},
			{MessageId: "conditionFixCompareFalse", Output: strings.Replace(code, "!state?.active", ";((state?.active) === false)", 1)},
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{reviewError(t, code, "state?.active", "conditionErrorNullableBoolean", "Unexpected nullable boolean value in conditional. Please handle the nullish case explicitly.", suggestions)}})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, nil, invalid)
}

func TestStrictBooleanExpressionsReviewCallEscapes(t *testing.T) {
	const declarations = `type Options = {method: 'some' | 'forEach'}; declare function configure(opts: Options): void; const opts: Options = {method: 'some'}; `
	valid := []rule_tester.ValidTestCase{
		{Code: declarations + `configure(opts); [1, 2][opts.method](n => console.log(n));`},
		{Code: declarations + `const alias = opts; configure(alias); [1, 2][opts.method](n => console.log(n));`},
		{Code: declarations + `const holder = {opts}; declare function change(value: typeof holder): void; change(holder); [1, 2][opts.method](n => console.log(n));`},
		{Code: declarations + `declare const change: ((opts: Options) => void) | undefined; change?.(opts); [1, 2][opts.method](n => console.log(n));`},
		{Code: declarations + `declare const change: {run(): void}; change.run(); [1, 2].forEach(n => console.log(n));`},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, suffix := range []string{
		`[1, 2][opts.method](n => console.log(n));`,
		`const method = opts.method; configure(opts); [1, 2][method](n => console.log(n));`,
		`configure({method: 'some'}); [1, 2][opts.method](n => console.log(n));`,
		`const copied = {method: opts.method}; configure(opts); [1, 2][copied.method](n => console.log(n));`,
		`[1, 2][opts.method]((configure(opts), (n: number) => console.log(n)));`,
	} {
		code := declarations + suffix
		callback := "n => console.log(n)"
		if strings.Contains(suffix, "(n: number)") {
			callback = "(n: number) => console.log(n)"
		}
		suggestion := strings.Replace(code, callback, strings.Replace(callback, " =>", ": boolean =>", 1), 1)
		// A parenless arrow needs parentheses around its parameter.
		if callback == "n => console.log(n)" {
			suggestion = strings.Replace(code, callback, "(n): boolean => console.log(n)", 1)
		}
		// Sequence expressions are not editable inline function nodes.
		var suggestions []rule_tester.InvalidTestCaseSuggestion
		operand := callback
		if strings.Contains(suffix, "(configure(opts)") {
			operand = "configure(opts), (n: number) => console.log(n)"
		} else {
			suggestions = []rule_tester.InvalidTestCaseSuggestion{{MessageId: "explicitBooleanReturnType", Output: suggestion}}
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{reviewError(t, code, operand, "conditionErrorNullish", "Unexpected nullish value in conditional. The condition is always false.", suggestions)}})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, valid, invalid)
}

func TestStrictBooleanExpressionsReviewSpreadArguments(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `const args = [(n: number) => n > 0, {}] as const; [1, 2].filter(...args);`},
		{Code: `const args: [predicate: (n: number) => boolean, thisArg: object] = [n => n > 0, {}]; [1, 2].filter(...args);`},
		{Code: `declare const args: readonly [(n: number) => boolean, object] | readonly [(n: number) => boolean]; [1, 2].filter(...args);`},
		{Code: `function f<T extends readonly [(n: number) => boolean, object]>(args: T) { [1, 2].filter(...args); }`},
		{Code: `const empty = [] as const; [1, 2].filter(...empty, (n: number) => n > 0, {});`},
		{Code: `declare const empty: readonly []; [1, 2].filter(...empty, ...empty, (n: number) => n > 0);`},
		{Code: `declare const predicates: ((n: number) => boolean)[]; [1, 2].filter(...predicates);`},
		{Code: `declare const args: readonly [(n: number) => boolean, ...object[]]; [1, 2].filter(...args);`},
		{Code: `declare const predicates: readonly [(n: number) => boolean]; [1, 2].filter(...predicates, {});`},
		{Code: `declare const args: readonly [...(() => boolean)[], () => boolean]; [1, 2].filter(...args, {});`},
		{Code: `declare const args: readonly [...(() => boolean)[], () => boolean, object]; [1, 2].filter(...args);`},
		{Code: `function f<T extends readonly (() => boolean)[]>(args: readonly [...T, () => boolean, object]) { [1, 2].filter(...args); }`},

		{Code: `function f<T extends readonly (() => boolean)[]>(args: readonly [...T, () => boolean]) { [1, 2].filter(...args, {}); }`},

		{Code: `declare const args: readonly [] | readonly [(n: number) => boolean]; [1, 2].filter(...args, (n: number) => true);`},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, code := range []string{
		`const args = [(n: number) => {}, {}] as const; [1, 2].filter(...args);`,
		`declare const args: readonly [() => void, ...object[]]; [1, 2].filter(...args);`,
		`function f<T extends readonly [() => void, object]>(args: T) { [1, 2].filter(...args); }`,
		`declare const args: readonly [() => void] | readonly [() => void, object]; [1, 2].filter(...args);`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{reviewError(t, code, "...args", "conditionErrorNullish", "Unexpected nullish value in conditional. The condition is always false.", nil)}})
	}
	for _, code := range []string{
		`declare const args: readonly [...(() => boolean)[], object]; [1, 2].filter(...args);`,
		`declare const args: readonly [predicate?: () => boolean]; [1, 2].filter(...args, {});`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{reviewError(t, code, "...args", "conditionErrorOther", "Unexpected value in conditional. A boolean expression is required.", nil)}})
	}
	code := `declare const args: readonly [() => boolean, ...object[]] | readonly [() => void, object]; [1, 2].filter(...args);`
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{reviewError(t, code, "...args", "conditionErrorNullableBoolean", "Unexpected nullable boolean value in array predicate return type. Please handle the nullish case explicitly.", nil)}})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictBooleanExpressionsRule, valid, invalid)
}
