// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-array-index-of.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/shared/simple-array-search-rule-tests.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-array-index-of.md
package prefer_array_index_of_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_index_of"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

type indexSearch struct {
	method, replacement string
}

func (search indexSearch) expand(code string) string {
	return strings.NewReplacer("FIND", search.method, "INDEX", search.replacement).Replace(code)
}

func (search indexSearch) invalid(filename, template, fixed, suggestion string) rule_tester.InvalidTestCase {
	result := rule_tester.InvalidTestCase{FileName: filename, Code: search.expand(template)}
	if fixed != "" {
		result.Output = []string{search.expand(fixed)}
	}
	prefix := "prefer-" + search.replacement + "-over-" + search.method + "/"
	// Each FIND marker is an expected report on a method name.
	for before := ""; ; {
		segment, rest, found := strings.Cut(template, "FIND")
		if !found {
			break
		}
		before += search.expand(segment)
		line := strings.Count(before, "\n") + 1
		column := ecmascript.StringCodeUnitCount(before[strings.LastIndexByte(before, '\n')+1:]) + 1
		result.Errors = append(result.Errors, rule_tester.InvalidTestCaseError{
			MessageId: prefix + "error",
			Message:   "Use `." + search.replacement + "()` instead of `." + search.method + "()` when looking for the index of an item.",
			Line:      line, Column: column, EndLine: line, EndColumn: column + len(search.method),
		})
		before += search.method
		template = rest
	}
	if suggestion != "" {
		result.Errors[0].Suggestions = []rule_tester.InvalidTestCaseSuggestion{{MessageId: prefix + "suggestion", Output: search.expand(suggestion)}}
	}
	return result
}

func TestPreferArrayIndexOfUpstream(t *testing.T) {
	for _, search := range []indexSearch{{"findIndex", "indexOf"}, {"findLastIndex", "lastIndexOf"}} {
		t.Run(search.method, func(t *testing.T) {
			t.Run("syntax", func(t *testing.T) { testUpstreamSyntax(t, search) })
			t.Run("types", func(t *testing.T) { testUpstreamTypes(t, search) })
		})
	}
}

func testUpstreamSyntax(t *testing.T, search indexSearch) {
	valid := []rule_tester.ValidTestCase{
		// Member and call shapes.
		{Code: search.expand("const FIND = foo.FIND"), FileName: "case.js"},
		{Code: search.expand("foo.FIND()"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(function (x) {return x === 1;}, bar)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(...[function (x) {return x === 1;}])"), FileName: "case.js"},
		{Code: search.expand("new foo.FIND(x => x === 1)"), FileName: "case.js"},
		{Code: search.expand("FIND(x => x === 1)"), FileName: "case.js"},
		{Code: search.expand("foo[\"FIND\"](x => x === 1)"), FileName: "case.js"},
		{Code: search.expand("foo[FIND](x => x === 1)"), FileName: "case.js"},
		{Code: search.expand("foo.not_FIND(x => x === 1)"), FileName: "case.js"},
		// Callback kind and parameters.
		{Code: search.expand("foo.FIND(myFunction)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND((x, i) => x === i)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND((x, i) => {return x === i})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(function(x, i) {return x === i})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(({x}) => x === 1)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(({x}) => {return x === 1})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(function({x}) {return x === 1})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(function * (x) {return x === 1})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(async (x) => x === 1)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(async (x) => {return x === 1})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(async function(x) {return x === 1})"), FileName: "case.js"},
		// Callback body.
		{Code: search.expand("foo.FIND(x => {return;})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(function (x) {return;})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND((x = 1) => {return;})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(() => {return;})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(({x}) => {noop();return x === 1})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(({x}) => {bar(x === 1)})"), FileName: "case.js"},
		// Comparison and references.
		{Code: search.expand("foo.FIND(x => x - 1)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(x => {return x - 1})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(function (x){return x - 1})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(x => x !== 1)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(x => {return x !== 1})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(function (x){return x !== 1})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(x => 1 === 1.0)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(x => {return 1 === 1.0})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(function (x){return 1 === 1.0})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(x => x === x)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(x => y === 1)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(x => x + \"foo\" === \"foo\" + x)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(x => x === \"foo\" + x)"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(x => x === (function (){return x === \"1\"})())"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(function fn(x) {return x === fn(y)})"), FileName: "case.js"},
		// Function scope and receiver type.
		{Code: search.expand("foo.FIND(function(x) {return x === arguments.length})"), FileName: "case.js"},
		{Code: search.expand("foo.FIND(function(x) {return x === this.length})"), FileName: "case.js"},
		{Code: search.expand("foo.INDEX(0)"), FileName: "case.js"},
		// Vue template parser is not available to the native rule tester.
		{Code: search.expand("<template><div v-if=\"values.FIND(x => x === 'foo')\"></div></template>"), FileName: "case.vue", Skip: true},
		{Code: search.expand("function f(foo: Set<number>) { return foo.FIND(x => x === 1); }"), FileName: "case.ts"},
		// Documentation examples.
		{Code: search.expand("const index = foo.INDEX('foo');"), FileName: "case.js"},
		{Code: search.expand("const index = foo.FIND(x => x == undefined);"), FileName: "case.js"},
		{Code: search.expand("const index = foo.FIND(x => x !== 'foo');"), FileName: "case.js"},
		{Code: search.expand("const index = foo.FIND((x, index) => x === index);"), FileName: "case.js"},
		{Code: search.expand("const index = foo.FIND(x => (x === 'foo') && isValid());"), FileName: "case.js"},
		{Code: search.expand("const index = foo.FIND(x => y === 'foo');"), FileName: "case.js"},
		{Code: search.expand("const index = foo.FIND(x => y.x === 'foo');"), FileName: "case.js"},
		{Code: search.expand("const index = foo.FIND(x => {\n\tconst bar = getBar();\n\treturn x === bar;\n});"), FileName: "case.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		search.invalid("case.js", "values.FIND(x => x === \"foo\")", "values.INDEX(\"foo\")", ""),
		search.invalid("case.js", "values?.FIND(x => x === \"foo\")", "values?.INDEX(\"foo\")", ""),
		search.invalid("case.js", "values.FIND(x => \"foo\" === x)", "values.INDEX(\"foo\")", ""),
		search.invalid("case.js", "values.FIND(x => {return x === \"foo\";})", "values.INDEX(\"foo\")", ""),
		search.invalid("case.js", "values.FIND(function (x) {return x === \"foo\";})", "values.INDEX(\"foo\")", ""),
		search.invalid("case.js", "// 1\n(0, values)\n\t// 2\n\t./* 3 */FIND /* 3 */ (\n\t\t/* 4 */\n\t\tx /* 5 */ => /* 6 */ x /* 7 */ === /* 8 */ \"foo\" /* 9 */\n\t) /* 10 */", "", ""),
		search.invalid("case.js", "foo.FIND(function (element) {\n\treturn element === bar.FIND(x => x === 1);\n});", "foo.FIND(function (element) {\n\treturn element === bar.INDEX(1);\n});", "foo.INDEX(bar.FIND(x => x === 1));"),
		search.invalid("case.js", "values.FIND(x => x === (0, \"foo\"))", "values.INDEX((0, \"foo\"))", ""),
		search.invalid("case.js", "values.FIND((x => x === (0, \"foo\")))", "values.INDEX((0, \"foo\"))", ""),
		search.invalid("case.js", "function fn() {\n\tfoo.FIND(x => x === arguments.length)\n}", "function fn() {\n\tfoo.INDEX(arguments.length)\n}", ""),
		search.invalid("case.js", "function fn() {\n\tfoo.FIND(x => x === this[1])\n}", "function fn() {\n\tfoo.INDEX(this[1])\n}", ""),
		search.invalid("case.js", "values.FIND(x => x === foo())", "", "values.INDEX(foo())"),
		search.invalid("case.js", "foo.FIND(function a(x) {\n\treturn x === (function (a) {\n\t\treturn a(this) === arguments[1]\n\t}).call(thisObject, anotherFunctionNamedA, secondArgument)\n})", "", "foo.INDEX((function (a) {\n\t\treturn a(this) === arguments[1]\n\t}).call(thisObject, anotherFunctionNamedA, secondArgument))"),
		search.invalid("case.js", "values.FIND(item => item === target);", "values.INDEX(target);", ""),
		search.invalid("case.js", "values.FIND(item => item === target.id);", "values.INDEX(target.id);", ""),
		search.invalid("case.js", "values.FIND(item => item === target?.id);", "values.INDEX(target?.id);", ""),
		search.invalid("case.js", "values.FIND(item => item === 1 + 2);", "values.INDEX(1 + 2);", ""),
		search.invalid("case.js", "values.FIND(item => item === -2n);", "values.INDEX(-2n);", ""),
		search.invalid("case.js", "values.FIND(item => item === +1n);", "values.INDEX(+1n);", ""),
		search.invalid("case.js", "values.FIND(item => item === undefined);", "values.INDEX(undefined);", ""),
		search.invalid("case.js", "values.FIND(item => item === /item/u);", "values.INDEX(/item/u);", ""),
		search.invalid("case.js", "values.FIND(item => item === `item`);", "values.INDEX(`item`);", ""),
		search.invalid("case.js", "const target = 2; values.FIND(item => item === target);", "const target = 2; values.INDEX(target);", ""),
		search.invalid("case.js", "values.FIND(item => item === /* keep */ target.id);", "", ""),
		// TypeScript syntax without a project.
		search.invalid("case.ts", "const values = [\"small\", \"medium\", \"large\"] as const; declare const input: string; values.FIND(value => value === input);", "const values = [\"small\", \"medium\", \"large\"] as const; declare const input: string; values.INDEX(input);", ""),
		search.invalid("case.ts", "values.FIND(item => item === (target as number));", "values.INDEX((target as number));", ""),
		search.invalid("case.ts", "values.FIND(item => item === (target satisfies number));", "values.INDEX((target satisfies number));", ""),
		search.invalid("case.ts", "values.FIND(item => item === target!);", "values.INDEX(target!);", ""),
		search.invalid("case.ts", "values.FIND(item => item === <number>target);", "values.INDEX(<number>target);", ""),
		search.invalid("case.ts", "values.FIND(item => item === (-2 as number));", "values.INDEX((-2 as number));", ""),
		search.invalid("case.ts", "values.FIND(item => item === (2 satisfies number));", "values.INDEX((2 satisfies number));", ""),
		search.invalid("case.ts", "values.FIND(item => item === 2!);", "values.INDEX(2!);", ""),
		search.invalid("case.ts", "values.FIND(item => item === <number>2);", "values.INDEX(<number>2);", ""),
		search.invalid("case.ts", "values.FIND(item => item === (getTarget() as number));", "", "values.INDEX((getTarget() as number));"),
		search.invalid("case.ts", "function foo() {\n\treturn (bar as string[]).FIND(x => x === \"foo\");\n}", "function foo() {\n\treturn (bar as string[]).INDEX(\"foo\");\n}", ""),
		// Documentation examples.
		search.invalid("case.js", "const index = foo.FIND(x => x === 'foo');", "const index = foo.INDEX('foo');", ""),
		search.invalid("case.js", "const index = foo.FIND(x => 'foo' === x);", "const index = foo.INDEX('foo');", ""),
		search.invalid("case.js", "const index = foo.FIND(x => {\n\treturn x === 'foo';\n});", "const index = foo.INDEX('foo');", ""),
		// NaN uses strict equality in both methods.
		search.invalid("case.js", "foo.FIND(element => element === NaN);", "foo.INDEX(NaN);", ""),
	}

	// Side effect tests: calls and assignments are suggestions, even on empty arrays.
	invalid = append(invalid,
		search.invalid("case.js",
			`let calls = 0; function getValue() { calls++; return calls + 1; } const index = [1, 2].FIND(item => item === getValue()); JSON.stringify({index, calls});`,
			"",
			`let calls = 0; function getValue() { calls++; return calls + 1; } const index = [1, 2].INDEX(getValue()); JSON.stringify({index, calls});`),
		search.invalid("case.js",
			`let value = 1; const index = [].FIND(item => item === (value += 1)); JSON.stringify({index, value});`,
			"",
			`let value = 1; const index = [].INDEX((value += 1)); JSON.stringify({index, value});`),
	)

	// The upstream runtime matrix also checks these fixes on empty and nonempty arrays.
	for _, value := range []string{
		"0", "-0", "+2", "-(-2)", "1e999", "-1e999", `"item"`, "true", "false", "null", "2n",
		"undefined", "`item`", "/item/u", "target", "object.id", "object?.id", "object.missing",
	} {
		for _, values := range []string{`[]`, `[0, 2, "item", true, false, null, 2n, 1e999, undefined]`} {
			base := `const target = 2; const object = {id: "item"}; ` + values
			invalid = append(invalid, search.invalid("case.js",
				base+".FIND(item => item === "+value+");",
				base+".INDEX("+value+");", ""))
		}
	}

	// The Go tester always builds a compiler Program. Hide the checker for
	// fixtures run upstream with a syntax-only parser, including the wider
	// string search in a tuple. The IPC mirror uses a real source-only Program.
	syntaxOnly := prefer_array_index_of.PreferArrayIndexOfRule
	run := syntaxOnly.Run
	syntaxOnly.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return run(ctx, options)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &syntaxOnly, valid, invalid)
}

func testUpstreamTypes(t *testing.T, search indexSearch) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, tc := range []struct {
		name, extension, template, callback, searchValue string
		suggestion                                       bool
	}{
		{"wider string", "ts", "const values = [\"small\", \"medium\", \"large\"] as const; declare const input: string; values.SEARCH;", "value => value === input", "", false},
		{"wider object", "ts", "declare const values: readonly {id: string; label: string}[]; declare const input: {id: string}; values.SEARCH;", "value => value === input", "", false},
		{"reversed comparison", "ts", "const values = [\"small\", \"medium\", \"large\"] as const; declare const input: string; values.SEARCH;", "value => input === value", "", false},
		{"generic unknown", "ts", "function check<Type>(value: unknown, values: readonly Type[]): number { return values.SEARCH; }", "option => option === value", "", false},
		{"constrained literal receiver", "ts", "function check<Type extends readonly [\"small\", \"large\"]>(values: Type, input: string) { return values.SEARCH; }", "value => value === input", "", false},
		{"optional value with block body", "ts", "declare const values: string[]; declare const input: string | undefined; values.SEARCH;", "value => {return value === input;}", "", false},
		{"incompatible side effect", "ts", "const values = [\"small\", \"medium\", \"large\"] as const; declare function getInput(): string; values.SEARCH;", "value => value === getInput()", "", false},
		{"widened callback parameter", "ts", "const values = [\"small\", \"medium\", \"large\"] as const; declare const input: string; values.SEARCH;", "(value: string) => value === input", "", false},
		{"incompatible union receiver", "ts", "declare const values: string[] | number[]; declare const input: string; values.SEARCH;", "value => value === input", "", false},
		{"incompatible optional receiver", "ts", "declare const values: readonly [\"small\"] | undefined; declare const input: string; values?.SEARCH;", "value => value === input", "", false},
		{"compatible string", "ts", "declare const values: string[]; declare const input: string; values.SEARCH;", "value => value === input", "input", false},
		{"compatible object with extra property", "ts", "declare const values: readonly {id: string; label: string}[]; declare const input: {id: string; label: string; extra: boolean}; values.SEARCH;", "value => value === input", "input", false},
		{"compatible side effect", "ts", "declare const values: string[]; declare function getInput(): string; values.SEARCH;", "value => value === getInput()", "getInput()", true},
		{"compatible literal", "ts", "const values = [\"small\", \"medium\", \"large\"] as const; values.SEARCH;", "value => value === \"small\"", "\"small\"", false},
		{"compatible generic", "ts", "function check<Type>(input: Type, values: readonly Type[]) { return values.SEARCH; }", "value => value === input", "input", false},
		{"compatible constrained receiver", "ts", "function check<Type extends readonly string[]>(values: Type, input: string) { return values.SEARCH; }", "value => value === input", "input", false},
		{"compatible optional value", "ts", "declare const values: (string | undefined)[]; declare const input: string | undefined; values.SEARCH;", "value => value === input", "input", false},
		{"compatible union receiver", "ts", "declare const values: string[] | (string | number)[]; declare const input: string; values.SEARCH;", "value => value === input", "input", false},
		{"compatible optional receiver", "ts", "declare const values: string[] | undefined; declare const input: string; values?.SEARCH;", "value => value === input", "input", false},
		{"compatible typed array", "ts", "declare const values: Uint8Array; declare const input: number; values.SEARCH;", "value => value === input", "input", false},
		{"BigInt typed array with unknown search value", "ts", "declare const values: BigInt64Array; declare const input: unknown; values.SEARCH;", "value => value === input", "", false},
		{"compatible BigInt typed array", "ts", "declare const values: BigInt64Array; declare const input: bigint; values.SEARCH;", "value => value === input", "input", false},
		{"JSDoc wider string", "js", "/** @type {readonly (\"small\" | \"large\")[]} */\nconst values = [\"small\", \"large\"];\n/** @param {string} input */\nfunction check(input) { return values.SEARCH; }", "value => value === input", "", false},
		{"JSDoc compatible string", "js", "/** @type {string[]} */\nconst values = [\"small\", \"large\"];\n/** @param {string} input */\nfunction check(input) { return values.SEARCH; }", "value => value === input", "input", false},
		{"unknown receiver", "ts", "declare const values: any; declare const input: string; values.SEARCH;", "(value: unknown) => value === input", "input", false},
		{"narrowed value", "ts", "function check(values: string[], input: string | undefined) { if (typeof input !== \"string\") { return; } return values.SEARCH; }", "value => value === input", "input", false},
	} {
		filename := "case." + tc.extension
		code := strings.Replace(tc.template, "SEARCH", "FIND("+tc.callback+")", 1)
		if tc.searchValue == "" {
			valid = append(valid, rule_tester.ValidTestCase{FileName: filename, Code: search.expand(code)})
			continue
		}
		output := strings.Replace(tc.template, "SEARCH", "INDEX("+tc.searchValue+")", 1)
		if tc.suggestion {
			invalid = append(invalid, search.invalid(filename, code, "", output))
		} else {
			invalid = append(invalid, search.invalid(filename, code, output, ""))
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_array_index_of.PreferArrayIndexOfRule, valid, invalid)
}
