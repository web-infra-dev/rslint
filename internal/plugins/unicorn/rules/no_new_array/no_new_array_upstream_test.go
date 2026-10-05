// Complete tests and documentation examples from eslint-plugin-unicorn v77.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/no-new-array.js
package no_new_array_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_new_array"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const errorMessage = "`new Array()` is unclear in intent; use an array literal or `Array.from()`."

func arrayCase(code, filename string, line, column, endLine, endColumn int, output, suggestion string) rule_tester.InvalidTestCase {
	diagnostic := rule_tester.InvalidTestCaseError{
		MessageId: "error", Message: errorMessage,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
	if suggestion != "" {
		diagnostic.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{MessageId: "spread", Output: suggestion}}
	}
	test := rule_tester.InvalidTestCase{Code: code, FileName: filename, Errors: []rule_tester.InvalidTestCaseError{diagnostic}}
	if output != "" {
		test.Output = []string{output}
	}
	return test
}

func TestNoNewArrayUpstreamSnapshots(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_new_array.NoNewArrayRule, []rule_tester.ValidTestCase{
		{Code: "const array = Array.from({length: 1})", FileName: "file.js"},
		{Code: "const array = new Array()", FileName: "file.js"},
		{Code: "const array = new Array", FileName: "file.js"},
		{Code: "const array = new Array(1, 2)", FileName: "file.js"},
		{Code: "const array = Array(1, 2)", FileName: "file.js"},
		{Code: "const array = Array(1)", FileName: "file.js"},
	},
		[]rule_tester.InvalidTestCase{
			arrayCase("const modes = new Set(['foo']);\nmodes.clear();\nnew Array(modes.size ? 1 : 'x');", "file.js", 3, 1, 3, 32, "", ""),
			arrayCase("const modes = new Set([\"foo\"]); modes.clear(); new Array((modes.size && 1) || value);", "file.js", 1, 48, 1, 85, "", ""),
			arrayCase("const object = {value: true}; Object.defineProperty(object, \"value\", {get() { return false; }}); new Array(object.value ? 1 : value);", "file.js", 1, 98, 1, 133, "", ""),
			arrayCase("const modes = new Set([\"foo\"]); modes.clear(); new Array((modes.size ? \"x\" : 1) as number);", "file.ts", 1, 48, 1, 91, "", ""),
			arrayCase("const array = new Array(1)", "file.js", 1, 15, 1, 27, "", ""),
			arrayCase("const zero = 0;\nconst array = new Array(zero);", "file.js", 2, 15, 2, 30, "", ""),
			arrayCase("const length = 1;\nconst array = new Array(length);", "file.js", 2, 15, 2, 32, "", ""),
			arrayCase("const array = new Array(1.5)", "file.js", 1, 15, 1, 29, "", ""),
			arrayCase("const array = new Array(Number(\"1\"))", "file.js", 1, 15, 1, 37, "", ""),
			arrayCase("const array = new Array(\"1\")", "file.js", 1, 15, 1, 29, "const array = [\"1\"]", ""),
			arrayCase("const array = new Array(null)", "file.js", 1, 15, 1, 30, "const array = [null]", ""),
			arrayCase("const array = new Array((\"1\"))", "file.js", 1, 15, 1, 31, "const array = [(\"1\")]", ""),
			arrayCase("const array = new Array((0, 1))", "file.js", 1, 15, 1, 32, "", ""),
			arrayCase("const foo = []\nnew Array(\"bar\").forEach(baz)", "file.js", 2, 1, 2, 17, "const foo = []\n;[\"bar\"].forEach(baz)", ""),
			arrayCase("new Array(0xff)", "file.js", 1, 1, 1, 16, "", ""),
			arrayCase("new Array(Math.PI | foo)", "file.js", 1, 1, 1, 25, "", ""),
			arrayCase("new Array(Math.min(foo, bar))", "file.js", 1, 1, 1, 30, "", ""),
			arrayCase("new Array(Number(foo))", "file.js", 1, 1, 1, 23, "", ""),
			arrayCase("new Array(Number.MAX_SAFE_INTEGER)", "file.js", 1, 1, 1, 35, "", ""),
			arrayCase("new Array(parseInt(foo))", "file.js", 1, 1, 1, 25, "", ""),
			arrayCase("new Array(Number.parseInt(foo))", "file.js", 1, 1, 1, 32, "", ""),
			arrayCase("new Array(+foo)", "file.js", 1, 1, 1, 16, "", ""),
			arrayCase("new Array(-Math.PI)", "file.js", 1, 1, 1, 20, "", ""),
			arrayCase("new Array(-\"-2\")", "file.js", 1, 1, 1, 17, "", ""),
			arrayCase("new Array(foo.length)", "file.js", 1, 1, 1, 22, "", ""),
			arrayCase("const array = [1]; new Array(array.length)", "file.js", 1, 20, 1, 43, "", ""),
			arrayCase("const text = \"foo\"; new Array(text.length)", "file.js", 1, 21, 1, 43, "", ""),
			arrayCase("const foo = 1; new Array(foo + 2)", "file.js", 1, 16, 1, 34, "", ""),
			arrayCase("new Array(foo - 2)", "file.js", 1, 1, 1, 19, "", ""),
			arrayCase("new Array(foo -= 2)", "file.js", 1, 1, 1, 20, "", ""),
			arrayCase("new Array(foo ? 1 : 2)", "file.js", 1, 1, 1, 23, "", ""),
			arrayCase("const truthy = \"truthy\"; new Array(truthy ? 1 : foo)", "file.js", 1, 26, 1, 53, "", ""),
			arrayCase("const falsy = !\"truthy\"; new Array(falsy ? foo : 1)", "file.js", 1, 26, 1, 52, "", ""),
			arrayCase("new Array((1n, 2))", "file.js", 1, 1, 1, 19, "", ""),
			arrayCase("new Array(Number.NaN)", "file.js", 1, 1, 1, 22, "", ""),
			arrayCase("new Array(NaN)", "file.js", 1, 1, 1, 15, "", ""),
			arrayCase("new Array(foo >>> bar)", "file.js", 1, 1, 1, 23, "", ""),
			arrayCase("new Array(foo >>>= bar)", "file.js", 1, 1, 1, 24, "", ""),
			arrayCase("new Array(++bar.length)", "file.js", 1, 1, 1, 24, "", ""),
			arrayCase("new Array(bar.length++)", "file.js", 1, 1, 1, 24, "", ""),
			arrayCase("new Array(foo = bar.length)", "file.js", 1, 1, 1, 28, "", ""),
			arrayCase("new Array(\"0xff\")", "file.js", 1, 1, 1, 18, "[\"0xff\"]", ""),
			arrayCase("new Array(Math.NON_EXISTS_PROPERTY)", "file.js", 1, 1, 1, 36, "", ""),
			arrayCase("new Array(Math.NON_EXISTS_METHOD(foo))", "file.js", 1, 1, 1, 39, "", ""),
			arrayCase("new Array(Math[min](foo, bar))", "file.js", 1, 1, 1, 31, "", ""),
			arrayCase("new Array(Number[MAX_SAFE_INTEGER])", "file.js", 1, 1, 1, 36, "", ""),
			arrayCase("new Array(new Number(foo))", "file.js", 1, 1, 1, 27, "", ""),
			arrayCase("const foo = 1; new Array(foo + \"2\")", "file.js", 1, 16, 1, 36, "const foo = 1; [foo + \"2\"]", ""),
			arrayCase("new Array(foo - 2n)", "file.js", 1, 1, 1, 20, "", ""),
			arrayCase("new Array(foo -= 2n)", "file.js", 1, 1, 1, 21, "", ""),
			arrayCase("new Array(foo instanceof 1)", "file.js", 1, 1, 1, 28, "", ""),
			arrayCase("new Array(foo || 1)", "file.js", 1, 1, 1, 20, "", ""),
			arrayCase("new Array(foo ||= 1)", "file.js", 1, 1, 1, 21, "", ""),
			arrayCase("new Array(foo ? 1n : 2)", "file.js", 1, 1, 1, 24, "", ""),
			arrayCase("new Array((1, 2n))", "file.js", 1, 1, 1, 19, "[(1, 2n)]", ""),
			arrayCase("new Array(-foo)", "file.js", 1, 1, 1, 16, "", ""),
			arrayCase("new Array(~foo)", "file.js", 1, 1, 1, 16, "", ""),
			arrayCase("new Array(typeof 1)", "file.js", 1, 1, 1, 20, "[typeof 1]", ""),
			arrayCase("const truthy = \"truthy\"; new Array(truthy ? foo : 1)", "file.js", 1, 26, 1, 53, "", ""),
			arrayCase("const falsy = !\"truthy\"; new Array(falsy ? 1 : foo)", "file.js", 1, 26, 1, 52, "", ""),
			arrayCase("new Array(unknown ? foo : 1)", "file.js", 1, 1, 1, 29, "", ""),
			arrayCase("new Array(unknown ? 1 : foo)", "file.js", 1, 1, 1, 29, "", ""),
			arrayCase("new Array(++foo)", "file.js", 1, 1, 1, 17, "", ""),
			arrayCase("const array = new Array(foo)", "file.js", 1, 15, 1, 29, "", ""),
			arrayCase("const array = new Array(length)", "file.js", 1, 15, 1, 32, "", ""),
			arrayCase("const foo = []\nnew Array(bar).forEach(baz)", "file.js", 2, 1, 2, 15, "", ""),
			arrayCase("const array = new Array(...[foo])", "file.js", 1, 15, 1, 34, "", "const array = [...[foo]]"),
			arrayCase("const array = new Array(...foo)", "file.js", 1, 15, 1, 32, "", "const array = [...foo]"),
			arrayCase("const array = new Array(...[...foo])", "file.js", 1, 15, 1, 37, "", "const array = [...[...foo]]"),
			arrayCase("const array = new Array(...[1])", "file.js", 1, 15, 1, 32, "", "const array = [...[1]]"),
			arrayCase("const array = new Array(...[\"1\"])", "file.js", 1, 15, 1, 34, "", "const array = [...[\"1\"]]"),
			arrayCase("const array = new Array(...[1, \"1\"])", "file.js", 1, 15, 1, 37, "", "const array = [...[1, \"1\"]]"),
			arrayCase("const foo = []\nnew Array(...bar).forEach(baz)", "file.js", 2, 1, 2, 18, "", "const foo = []\n;[...bar].forEach(baz)"),
		})
}

func TestNoNewArrayUpstreamExplicit(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_new_array.NoNewArrayRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			arrayCase("const object = {}; Object.defineProperty(object, \"length\", {get() { return \"foo\"; }}); new Array(object.length);", "file.js", 1, 88, 1, 112, "", ""),
			arrayCase("const array = []; new Array(1 + array.length);", "file.js", 1, 19, 1, 46, "", ""),
			arrayCase("const alias = condition; var condition = true; new Array(alias ? \"x\" : 1);", "file.js", 1, 48, 1, 74, "", ""),
			arrayCase("new Array(1.5)", "file.js", 1, 1, 1, 15, "", ""),
			arrayCase("new Array(-1)", "file.js", 1, 1, 1, 14, "", ""),
			arrayCase("new Array(1 / 2)", "file.js", 1, 1, 1, 17, "", ""),
			arrayCase("const size = 3; new Array(size / 2);", "file.js", 1, 17, 1, 36, "", ""),
			arrayCase("const size = 3; new Array(size * 2);", "file.js", 1, 17, 1, 36, "", ""),
			arrayCase("new Array(-0)", "file.js", 1, 1, 1, 14, "", ""),
			arrayCase("new Array(2.5 * 2)", "file.js", 1, 1, 1, 19, "", ""),
			arrayCase("new Array(Infinity)", "file.js", 1, 1, 1, 20, "", ""),
			arrayCase("const array = new Array(0xff);", "file.js", 1, 15, 1, 30, "", ""),
			arrayCase("const array = new Array(1e2);", "file.js", 1, 15, 1, 29, "", ""),
			arrayCase("const array = new Array(0b1);", "file.js", 1, 15, 1, 29, "", ""),
			arrayCase("const array = new Array(1_0);", "file.js", 1, 15, 1, 29, "", ""),
			arrayCase("new Array(/* the size */ 3);", "file.js", 1, 1, 1, 28, "", ""),
			arrayCase("new /* c */ Array(1);", "file.js", 1, 1, 1, 21, "", ""),
			arrayCase("new Array(1 /* c */);", "file.js", 1, 1, 1, 21, "", ""),
			arrayCase("new Array(/* c */ \"x\");", "file.js", 1, 1, 1, 23, "", ""),
			arrayCase("new /* c */ Array(\"x\");", "file.js", 1, 1, 1, 23, "", ""),
			arrayCase("new Array(/* c */ ...foo);", "file.js", 1, 1, 1, 26, "", ""),
		})
}

func TestNoNewArrayUpstreamDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_new_array.NoNewArrayRule, []rule_tester.ValidTestCase{
		{Code: "const array = [onlyElement];", FileName: "file.js"},
		{Code: "const items = ['foo', 'bar'];\nconst array = [...items];", FileName: "file.js"},
	},
		[]rule_tester.InvalidTestCase{
			arrayCase("const length = 10;\nconst array = new Array(length);", "file.js", 2, 15, 2, 32, "", ""),
			arrayCase("const array = new Array(3);\narray.filter(() => true).length;\nArray.from({length: 3}).filter(() => true).length;", "file.js", 1, 15, 1, 27, "", ""),
			arrayCase("const array = new Array(onlyElement);", "file.js", 1, 15, 1, 37, "", ""),
			arrayCase("const items = ['foo', 'bar'];\nconst array = new Array(...items);", "file.js", 2, 15, 2, 34, "", "const items = ['foo', 'bar'];\nconst array = [...items];"),
		})
}
