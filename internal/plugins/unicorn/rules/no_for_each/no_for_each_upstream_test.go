// Ported from eslint-plugin-unicorn v77.0.0 test/no-for-each.js.
// cspell:ignore callbag
package no_for_each_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_for_each"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	errorMessageID      = "no-for-each/error"
	suggestionMessageID = "no-for-each/suggestion"
)

func valid(code, fileName string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: fileName}
}

func forEachError(code string, occurrence int) rule_tester.InvalidTestCaseError {
	start := 0
	offset := -1
	for range occurrence + 1 {
		relative := strings.Index(code[start:], "forEach")
		if relative < 0 {
			panic("forEach not found in test case")
		}
		offset = start + relative
		start = offset + len("forEach")
	}
	prefix := code[:offset]
	line := strings.Count(prefix, "\n") + 1
	lastNewline := strings.LastIndex(prefix, "\n")
	column := offset + 1
	if lastNewline >= 0 {
		column = offset - lastNewline
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: errorMessageID,
		Message:   "Use `for…of` instead of `.forEach(…)`.",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len("forEach"),
	}
}

func invalid(code, fileName string, output string) rule_tester.InvalidTestCase {
	testCase := rule_tester.InvalidTestCase{
		Code:     code,
		FileName: fileName,
		Errors:   []rule_tester.InvalidTestCaseError{forEachError(code, 0)},
	}
	if output != "" {
		testCase.Output = []string{output}
	}
	return testCase
}

func TestNoForEachUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_for_each.NoForEachRule,
		[]rule_tester.ValidTestCase{
			valid(`new foo.forEach(element => bar())`, "file.js"),
			valid(`forEach(element => bar())`, "file.js"),
			valid(`foo.notForEach(element => bar())`, "file.js"),
			valid(`React.Children.forEach(children, child => {});`, "file.js"),
			valid(`Children.forEach(children, child => {});`, "file.js"),
			valid(`await pIteration.forEach(plugins, async pluginName => {});`, "file.js"),
			valid(`Effect.forEach([1, 2, 3], n => Effect.succeed(n));`, "file.js"),
			valid(`const map = new Map(); map.forEach(value => console.log(value));`, "file.js"),
			valid(`const set = new Set(); set.forEach(value => console.log(value));`, "file.js"),
			valid("import * as CB from 'strict-callbag-basics';\nCB.forEach(x => console.log(x));", "file.js"),
			valid(`function foo(map: Map<string, string>) { map.forEach(value => console.log(value)); }`, "file.ts"),
			valid(`function foo(set: ReadonlySet<string>) { set.forEach(value => console.log(value)); }`, "file.ts"),
			valid(`type Array<T> = Map<T, T>; function foo(value: Array<string>) { value.forEach(x => use(x)); }`, "file.ts"),
			valid(`type Uint8Array = Set<number>; function foo(value: Uint8Array) { value.forEach(x => use(x)); }`, "file.ts"),
		},
		[]rule_tester.InvalidTestCase{
			invalid(`foo.forEach?.(element => bar(element))`, "file.js", ""),
			invalid(`foo.forEach(element => bar(element), thisArgument)`, "file.js", ""),
			invalid(`foo.forEach()`, "file.js", ""),
			invalid(`const baz = foo.forEach(element => bar(element))`, "file.js", ""),
			invalid(`foo.forEach(bar)`, "file.js", ""),
			invalid(`foo.forEach(async function(element) {})`, "file.js", ""),
			invalid(`foo.forEach(function * (element) {})`, "file.js", ""),
			invalid(`foo.forEach(() => bar())`, "file.js", ""),
			invalid(`foo.forEach((element, index, array) => bar())`, "file.js", ""),
			invalid(`property.forEach(({property}) => bar(property))`, "file.js", ""),
			invalid(`foo.forEach((element = {}) => call(element))`, "file.js", ""),
			invalid(`foo.forEach((...args) => bar(...args))`, "file.js", ""),
			invalid(`[1, 2, 3].forEach(element => bar(element))`, "file.js", `for (const element of [1, 2, 3]) bar(element)`),
			invalid(`const array = []; array.forEach(element => bar(element));`, "file.js", `const array = []; for (const element of array) bar(element);`),
			invalid(`const array = []; array.forEach((element, index) => bar(element, index));`, "file.js", `const array = []; for (const [index, element] of array.entries()) bar(element, index);`),
			invalid(`const array = []; (array).forEach(element => bar(element));`, "file.js", `const array = []; for (const element of (array)) bar(element);`),
			invalid(`const array = []; array.forEach((element => bar(element)));`, "file.js", `const array = []; for (const element of array) bar(element);`),
			invalid(`const array = []; array.forEach(element => { bar(element); });`, "file.js", `const array = []; for (const element of array) { bar(element); }`),
			invalid(`const array = []; array.forEach(element => {/* comment */ bar(element);});`, "file.js", `const array = []; for (const element of array) {/* comment */ bar(element);}`),
			invalid(`const array = []; array.forEach(/* comment */ element => bar(element));`, "file.js", ""),
			invalid(`const array = []; array.forEach(element => bar(element),);`, "file.js", `const array = []; for (const element of array) bar(element);`),
			invalid(`const array = []; array.forEach(async element => bar(element));`, "file.js", ""),
			invalid(`const array = []; const result = array.forEach(element => bar(element));`, "file.js", ""),
			invalid(`const array = []; array.forEach(element => { for (const item of element) { return; } });`, "file.js", ""),
			invalid("const array = [];\narray.forEach(element => { if (element) bar(); else return; });", "file.js", "const array = [];\nfor (const element of array) { if (element) bar(); else continue; }"),
			invalid("const array = [];\narray.forEach(element => {\n\tfoo()\n\treturn [element]\n});", "file.js", "const array = [];\nfor (const element of array) {\n\tfoo()\n\t ;[element]; continue;\n}"),
			invalid(`const typedArray = new Uint8Array(); typedArray.forEach(value => console.log(value));`, "file.js", ""),
			invalid(`function foo(typedArray: Uint8Array) { typedArray.forEach(value => console.log(value)); }`, "file.ts", ""),
			invalid(`function foo(array: Array<string>) { array.forEach(value => console.log(value)); }`, "file.ts", `function foo(array: Array<string>) { for (const value of array) console.log(value); }`),
			invalid(`function foo(array: ReadonlyArray<string>) { array.forEach(value => console.log(value)); }`, "file.ts", `function foo(array: ReadonlyArray<string>) { for (const value of array) console.log(value); }`),
			invalid(`function foo(array: [string, string]) { array.forEach(value => console.log(value)); }`, "file.ts", `function foo(array: [string, string]) { for (const value of array) console.log(value); }`),
			invalid(`type Strings = string[]; function foo(array: Strings) { array.forEach(value => console.log(value)); }`, "file.ts", `type Strings = string[]; function foo(array: Strings) { for (const value of array) console.log(value); }`),
			invalid(`type Map = string[]; function foo(array: Map) { array.forEach(value => console.log(value)); }`, "file.ts", `type Map = string[]; function foo(array: Map) { for (const value of array) console.log(value); }`),
			invalid("interface PageInfo {}\ndeclare const staticPages: string[];\ndeclare const allStaticPages: Set<string>;\ndeclare const pageInfos: Map<string, PageInfo>;\ndeclare const allPageInfos: Map<string, PageInfo>;\n\nstaticPages.forEach(pg => allStaticPages.add(pg));\npageInfos.forEach((info, key) => allPageInfos.set(key, info));", "file.ts", "interface PageInfo {}\ndeclare const staticPages: string[];\ndeclare const allStaticPages: Set<string>;\ndeclare const pageInfos: Map<string, PageInfo>;\ndeclare const allPageInfos: Map<string, PageInfo>;\n\nfor (const pg of staticPages) allStaticPages.add(pg);\npageInfos.forEach((info, key) => allPageInfos.set(key, info));"),
			invalid("declare const elements: string[];\ndeclare function cloakElement(element: string): string;\nconst cloakVals: string[] = [];\nelements.forEach(element => cloakVals.push(cloakElement(element)));", "file.ts", "declare const elements: string[];\ndeclare function cloakElement(element: string): string;\nconst cloakVals: string[] = [];\nfor (const element of elements) cloakVals.push(cloakElement(element));"),
			invalid("declare function getStrings(): string[];\ngetStrings().forEach(value => console.log(value));", "file.ts", "declare function getStrings(): string[];\nfor (const value of getStrings()) console.log(value);"),
			{
				Code:     `function foo(collection: string[] | {forEach(callback: (value: string) => void): void}) { collection.forEach(value => console.log(value)); }`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					forEachError(`function foo(collection: string[] | {forEach(callback: (value: string) => void): void}) { collection.forEach(value => console.log(value)); }`, 1),
				},
			},
			invalid(`const element = 5; console.log(element); [1, 2, 3].forEach(element => bar(element));`, "file.js", `const element = 5; console.log(element); for (const element of [1, 2, 3]) bar(element);`),
			invalid(`const array = []; array.forEach(element => { element = foo(element); bar(element); });`, "file.js", `const array = []; for (let element of array) { element = foo(element); bar(element); }`),
			invalid("const foo = [1, 2, 3];\nfoo.forEach(x => function () {});", "file.js", "const foo = [1, 2, 3];\nfor (const x of foo) (function () {});"),
			invalid("const foo = [1, 2, 3];\nfoo.forEach(x => class {});", "file.js", "const foo = [1, 2, 3];\nfor (const x of foo) (class {});"),
			invalid("const foo = [1, 2, 3];\nfoo.forEach(x => ({bar: x}));", "file.js", "const foo = [1, 2, 3];\nfor (const x of foo) ({bar: x});"),
			invalid(`const array = []; array.forEach(element => {if (foo) return class {}; bar(element);});`, "file.js", `const array = []; for (const element of array) {if (foo)  { (class {}); continue; } bar(element);}`),
		},
	)
}
