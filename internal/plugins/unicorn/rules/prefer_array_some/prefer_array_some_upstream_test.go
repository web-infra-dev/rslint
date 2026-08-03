// TestPreferArraySomeUpstream migrates the full valid/invalid suite from
// upstream test/prefer-array-some.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the prefer_array_some_extras_test.go file.
//
// Upstream's Vue test cases (test.vue) are omitted: rslint does not support
// Vue single-file components.
package prefer_array_some_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_some"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageSome    = "some"
	messageFilter  = "filter"
	messageSuggest = "some-suggestion"
)

// findError builds the expected `.some` error with its suggestion for a
// `.find` / `.findLast` case.
func findError(output string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: messageSome,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: messageSuggest, Output: output},
		},
	}
}

func TestPreferArraySomeUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_array_some.PreferArraySomeRule, []rule_tester.ValidTestCase{
		// ---- `.find()` result is not used as a boolean ----
		{Code: `const bar = foo.find(fn)`},
		{Code: `const bar = foo.find(fn) || baz`},
		{Code: `if (foo.find(fn) ?? bar) {}`},
		{Code: `let hasFoo = foo.find(fn); if (hasFoo) {}`},
		{Code: `const hasFoo = foo.find(fn); hasFoo.value;`},
		{Code: `const hasFoo = foo.find(fn); if (hasFoo) { const value = hasFoo; }`},
		{Code: `const hasFoo = foo.find(fn); const value = other || hasFoo;`},
		{Code: `const {hasFoo} = foo.find(fn); if (hasFoo) {}`},
		{Code: `export const hasFoo = foo.find(fn);`},
		{Code: `const hasFoo = foo.find(fn); export {hasFoo};`},
		{Code: `const hasFoo = foo.find(fn);`},
		{Code: `const hasFoo = foo.find(fn); if (hasFoo === undefined) {}`},
		{Code: `const hasFoo = foo.find(fn); if (hasFoo || !hasFoo || Boolean(hasFoo)) {}`},
		// `Boolean` is shadowed → not the global boolean cast.
		{Code: `const Boolean = value => value; if (Boolean(foo.find(fn))) {}`},
		{Code: `function unicorn(Boolean) { if (Boolean(foo.find(fn))) {} }`},
		{Code: `const font = this.form[$globalData].fontFinder.find(typeface); if (!font) {}`},

		// ---- receiver is definitely not an array (new X(), string) ----
		{Code: `if (new Foo().find(fn)) {}`},
		{Code: `if (new Foo().findLast(fn)) {}`},
		{Code: `new Foo().findIndex(fn) !== -1`},
		{Code: `new Foo().findLastIndex(fn) !== -1`},
		{Code: `new Foo().filter(fn).length > 0`},
		{Code: `if ("foo".find(fn)) {}`},
		{Code: `const value = "foo"; if (value.find(fn)) {}`},
		{Code: "if ((`foo${bar}`).find(fn)) {}"},
		{Code: `if (({find() {}}).find(fn)) {}`},
		{Code: `if ((function () {}).find(fn)) {}`},
		{Code: `if ((() => {}).find(fn)) {}`},
		{Code: `if ((class {}).find(fn)) {}`},
		{Code: "(`foo${bar}`).findIndex(fn) !== -1"},
		{Code: "(`foo${bar}`).filter(fn).length > 0"},
		{Code: `({findIndex() {}}).findIndex(fn) !== -1`},
		{Code: `({filter() {}}).filter(fn).length > 0`},

		// ---- receiver is a non-indexed collection or unknown-with-shape ----
		{Code: `const collection = {find() {}}; if (collection.find(fn)) {}`, Tsx: false},
		{Code: `const store = new Store(); if (store.find(fn)) {}`},
		{Code: `interface QueryCache {find(queryKey: string[]): unknown} declare const queryClient: {getQueryCache(): QueryCache}; queryClient.getQueryCache().find(["foo"]) ? "foo" : "bar";`, Tsx: false},
		{Code: `interface QueryCache {findIndex(predicate: Function): number} declare const queryCache: QueryCache; queryCache.findIndex(fn) !== -1;`, Tsx: false},
		{Code: `interface QueryCache {filter(predicate: Function): {length: number}} declare const queryCache: QueryCache; queryCache.filter(fn).length > 0;`, Tsx: false},
		{Code: `interface QueryCache {findIndex(predicate: Function): number} declare function getQueryCache(): QueryCache; getQueryCache().findIndex(fn) !== -1;`, Tsx: false},
		{Code: `interface QueryCache {filter(predicate: Function): {length: number}} declare function getQueryCache(): QueryCache; getQueryCache().filter(fn).length > 0;`, Tsx: false},
		{Code: `class Store {find(key: string): boolean {return key.length > 0}} const store = new Store(); if (store.find("key")) {}`, Tsx: false},
		{Code: `class Store extends HTMLElement {find(key: string): boolean {return key.length > 0}} const store = new Store(); if (store.find("key")) {}`, Tsx: false},
		{Code: `function foo(collection: {find(predicate: Function): unknown}) { if (collection.find(fn)) {} }`, Tsx: false},
		{Code: `function foo<T extends {find(predicate: Function): unknown}>(collection: T) { if (collection.find(fn)) {} }`, Tsx: false},
		{Code: `function foo(collection: {findIndex(predicate: Function): number}) { collection.findIndex(fn) !== -1; }`, Tsx: false},
		{Code: `function foo(collection: {filter(predicate: Function): {length: number}}) { collection.filter(fn).length > 0; }`, Tsx: false},
		{Code: `type Collection = {find(predicate: Function): unknown}; const collection: Collection = [] as never; if (collection.find(fn)) {}`, Tsx: false},
		{Code: `type Collection = {findIndex(predicate: Function): number}; ([] as Collection).findIndex(fn) !== -1`, Tsx: false},
		{Code: `type Collection = {filter(predicate: Function): {length: number}}; ([] as Collection).filter(fn).length > 0`, Tsx: false},

		// ---- explicit type arguments mean the value is used ----
		{Code: `const hasFoo: Item | undefined = foo.find(fn); if (hasFoo) {}`, Tsx: false},
		{Code: `const hasFoo = foo.find<Item>(fn); if (hasFoo) {}`, Tsx: false},
		{Code: `const hasFoo = foo.findLast<Item>(fn); if (hasFoo) {}`, Tsx: false},
		{Code: `if (foo.find<Item>(fn)) {}`, Tsx: false},
		{Code: `if (foo.findLast<Item>(fn)) {}`, Tsx: false},
		{Code: `foo.find<Item>(fn) !== undefined`, Tsx: false},
		{Code: `foo.findLast<Item>(fn) !== undefined`, Tsx: false},

		// ---- not a matched CallExpression ----
		{Code: `if (new foo.find(fn)) {}`},
		{Code: `if (new foo.findLast(fn)) {}`},
		{Code: `if (find(fn)) {}`},
		{Code: `if (findLast(fn)) {}`},
		{Code: `if (foo["find"](fn)) {}`},
		{Code: `if (foo["findLast"](fn)) {}`},
		{Code: `if (foo["fi" + "nd"](fn)) {}`},
		{Code: `if (foo["fi" + "ndLast"](fn)) {}`},
		{Code: "if (foo[`find`](fn)) {}"},
		{Code: "if (foo[`findLast`](fn)) {}"},
		{Code: `if (foo[find](fn)) {}`},
		{Code: `if (foo[findLast](fn)) {}`},
		{Code: `if (foo.notFind(fn)) {}`},
		{Code: `if (foo.notFindLast(fn)) {}`},
		{Code: `if (foo.find()) {}`},
		{Code: `if (foo.findLast()) {}`},
		{Code: `if (foo.find(fn, thisArgument, extraArgument)) {}`},
		{Code: `if (foo.findLast(fn, thisArgument, extraArgument)) {}`},
		{Code: `if (foo.find(...argumentsArray)) {}`},
		{Code: `if (foo.findLast(...argumentsArray)) {}`},

		// ---- `.filter().length` valid comparisons (explicit-length-check) ----
		{Code: `array.filter(fn).length > 0.`},
		{Code: `array.filter(fn).length > .0`},
		{Code: `array.filter(fn).length > 0.0`},
		{Code: `array.filter(fn).length > 0x00`},
		{Code: `array.filter(fn).length < 0`},
		{Code: `array.filter(fn).length >= 0`},
		{Code: `0 > array.filter(fn).length`},
		{Code: `array.filter(fn).length !== 0.`},
		{Code: `array.filter(fn).length !== .0`},
		{Code: `array.filter(fn).length !== 0.0`},
		{Code: `array.filter(fn).length !== 0x00`},
		{Code: `array.filter(fn).length != 0`},
		{Code: `array.filter(fn).length === 0`},
		{Code: `array.filter(fn).length == 0`},
		{Code: `array.filter(fn).length = 0`},
		{Code: `0 !== array.filter(fn).length`},
		{Code: `array.filter(fn).length >= 1`},
		{Code: `array.filter(fn).length >= 1.`},
		{Code: `array.filter(fn).length >= 1.0`},
		{Code: `array.filter(fn).length >= 0x1`},
		{Code: `array.filter(fn).length > 1`},
		{Code: `array.filter(fn).length < 1`},
		{Code: `array.filter(fn).length = 1`},
		{Code: `array.filter(fn).length += 1`},
		{Code: `1 >= array.filter(fn).length`},
		{Code: `array.filter(fn)?.length > 0`},
		{Code: `array.filter(fn)[length] > 0`},
		{Code: `array.filter(fn).notLength > 0`},
		{Code: `array.filter(fn).length() > 0`},
		{Code: `+array.filter(fn).length >= 1`},
		{Code: `array.filter?.(fn).length > 0`},
		{Code: `array?.filter(fn).length > 0`},
		{Code: `array.notFilter(fn).length > 0`},
		{Code: `array.filter.length > 0`},
		{Code: `$element.filter(":visible").length > 0`},
		{Code: `var res = $module.filter(selector.disabled).length > 0;`},
		{Code: `$module.filter(fn).length !== 0`},

		// ---- `.findIndex()` valid comparisons ----
		{Code: `foo.notMatchedMethod(bar) !== -1`},
		{Code: `new foo.findIndex(bar) !== -1`},
		{Code: `foo.findIndex(bar, extraArgument) !== -1`},
		{Code: `foo.findIndex(bar) instanceof -1`},
		{Code: `foo.findIndex(...bar) !== -1`},
		{Code: `_.findIndex(bar)`},
		{Code: `_.findIndex(foo, bar)`},

		// ---- compare-with-undefined valid ----
		{Code: `foo.find(fn) == 0`},
		{Code: `foo.find(fn) != ""`},
		{Code: `foo.find(fn) === null`},
		{Code: `foo.find(fn) !== "null"`},
		{Code: `foo.find(fn) >= undefined`},
		{Code: `foo.find(fn) instanceof undefined`},
		// We are not checking `typeof` right now.
		{Code: `typeof foo.find(fn) === "undefined"`},
	}, []rule_tester.InvalidTestCase{
		// ---- boolean-position `.find()` / `.findLast()` ----
		{Code: `const bar = !foo.find(fn)`, Errors: []rule_tester.InvalidTestCaseError{findError(`const bar = !foo.some(fn)`)}},
		{Code: `const bar = !foo.findLast(fn)`, Errors: []rule_tester.InvalidTestCaseError{findError(`const bar = !foo.some(fn)`)}},
		{Code: `const bar = Boolean(foo.find(fn))`, Errors: []rule_tester.InvalidTestCaseError{findError(`const bar = Boolean(foo.some(fn))`)}},
		{Code: `const bar = Boolean(foo.findLast(fn))`, Errors: []rule_tester.InvalidTestCaseError{findError(`const bar = Boolean(foo.some(fn))`)}},
		{Code: `const bar = !!Boolean(foo.find(fn))`, Errors: []rule_tester.InvalidTestCaseError{findError(`const bar = !!Boolean(foo.some(fn))`)}},
		{Code: `const bar = !!Boolean(foo.findLast(fn))`, Errors: []rule_tester.InvalidTestCaseError{findError(`const bar = !!Boolean(foo.some(fn))`)}},
		{Code: `if (foo.find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`if (foo.some(fn)) {}`)}},
		{Code: `if (foo.findLast(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`if (foo.some(fn)) {}`)}},
		{Code: `const bar = foo.find(fn) ? 1 : 2`, Errors: []rule_tester.InvalidTestCaseError{findError(`const bar = foo.some(fn) ? 1 : 2`)}},
		{Code: `const bar = foo.findLast(fn) ? 1 : 2`, Errors: []rule_tester.InvalidTestCaseError{findError(`const bar = foo.some(fn) ? 1 : 2`)}},
		{Code: `while (foo.find(fn)) foo.shift();`, Errors: []rule_tester.InvalidTestCaseError{findError(`while (foo.some(fn)) foo.shift();`)}},
		{Code: `while (foo.findLast(fn)) foo.shift();`, Errors: []rule_tester.InvalidTestCaseError{findError(`while (foo.some(fn)) foo.shift();`)}},
		{Code: `do {foo.shift();} while (foo.find(fn));`, Errors: []rule_tester.InvalidTestCaseError{findError(`do {foo.shift();} while (foo.some(fn));`)}},
		{Code: `do {foo.shift();} while (foo.findLast(fn));`, Errors: []rule_tester.InvalidTestCaseError{findError(`do {foo.shift();} while (foo.some(fn));`)}},
		{Code: `for (; foo.find(fn); ) foo.shift();`, Errors: []rule_tester.InvalidTestCaseError{findError(`for (; foo.some(fn); ) foo.shift();`)}},
		{Code: `for (; foo.findLast(fn); ) foo.shift();`, Errors: []rule_tester.InvalidTestCaseError{findError(`for (; foo.some(fn); ) foo.shift();`)}},

		// ---- variable used only as boolean (array receiver) ----
		{Code: `const hasFoo = [].find(fn); if (hasFoo) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`const hasFoo = [].some(fn); if (hasFoo) {}`)}},
		{Code: `const hasFoo = [].findLast(fn); if (hasFoo) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`const hasFoo = [].some(fn); if (hasFoo) {}`)}},
		{Code: `const array = []; const hasFoo = array.find(fn); if (hasFoo) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`const array = []; const hasFoo = array.some(fn); if (hasFoo) {}`)}},
		{Code: `const array = []; const hasFoo = array.findLast(fn); if (hasFoo) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`const array = []; const hasFoo = array.some(fn); if (hasFoo) {}`)}},

		// ---- comments: suggestion keeps comments when only replacing method ----
		{
			Code:   `console.log(foo /* comment 1 */ . /* comment 2 */ find /* comment 3 */ (fn) ? a : b)`,
			Errors: []rule_tester.InvalidTestCaseError{findError(`console.log(foo /* comment 1 */ . /* comment 2 */ some /* comment 3 */ (fn) ? a : b)`)},
		},
		// jQuery.find is always truthy but upstream still reports.
		{Code: `if (jQuery.find(".outer > div")) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`if (jQuery.some(".outer > div")) {}`)}},

		// ---- actual message text ----
		{
			Code: `if (bar.find(fn)) {}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: messageSome,
				Message:   "Prefer `.some(…)` over `.find(…)`.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: messageSuggest, Output: `if (bar.some(fn)) {}`},
				},
			}},
		},
		{
			Code: `if (bar.findLast(fn)) {}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: messageSome,
				Message:   "Prefer `.some(…)` over `.findLast(…)`.",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: messageSuggest, Output: `if (bar.some(fn)) {}`},
				},
			}},
		},

		// ---- array receiver forms (type-aware, .ts) ----
		{Code: `if ([].find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`if ([].some(fn)) {}`)}},
		{Code: `if (Array().find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`if (Array().some(fn)) {}`)}},
		{Code: `if (new Array().find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`if (new Array().some(fn)) {}`)}},
		{Code: `if (Array.from(iterable).find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`if (Array.from(iterable).some(fn)) {}`)}},
		{Code: `if (Array.of(value).find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`if (Array.of(value).some(fn)) {}`)}},
		{Code: `const array = []; if (array.find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`const array = []; if (array.some(fn)) {}`)}},
		{Code: `const array = Array(); if (array.find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`const array = Array(); if (array.some(fn)) {}`)}},
		{Code: `const array = new Array(); if (array.find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`const array = new Array(); if (array.some(fn)) {}`)}},
		{Code: `const array = Array.from(iterable); if (array.find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`const array = Array.from(iterable); if (array.some(fn)) {}`)}},
		{Code: `const array = Array.of(value); if (array.find(fn)) {}`, Errors: []rule_tester.InvalidTestCaseError{findError(`const array = Array.of(value); if (array.some(fn)) {}`)}},
		{Code: `function foo(array: string[]) { if (array.find(fn)) {} }`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`function foo(array: string[]) { if (array.some(fn)) {} }`)}},
		{Code: `function foo(array: readonly string[]) { if (array.find(fn)) {} }`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`function foo(array: readonly string[]) { if (array.some(fn)) {} }`)}},
		{Code: `function foo(array: ReadonlyArray<string>) { if (array.find(fn)) {} }`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`function foo(array: ReadonlyArray<string>) { if (array.some(fn)) {} }`)}},
		{Code: `function foo(array: [string, string]) { if (array.find(fn)) {} }`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`function foo(array: [string, string]) { if (array.some(fn)) {} }`)}},
		{Code: `function foo<T extends string[]>(array: T) { if (array.find(fn)) {} }`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`function foo<T extends string[]>(array: T) { if (array.some(fn)) {} }`)}},
		{Code: `type Items = string[]; function foo(items: Items) { if (items.find(fn)) {} }`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`type Items = string[]; function foo(items: Items) { if (items.some(fn)) {} }`)}},
		{Code: `interface Items extends Array<string> {} function foo(items: Items) { if (items.find(fn)) {} }`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`interface Items extends Array<string> {} function foo(items: Items) { if (items.some(fn)) {} }`)}},
		{Code: `function foo<T>(items: string[] | {find(predicate: Function): unknown} | T) { if (items.find(fn)) {} }`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`function foo<T>(items: string[] | {find(predicate: Function): unknown} | T) { if (items.some(fn)) {} }`)}},
		{Code: `type Items = string[] | {find(predicate: Function): unknown}; function foo(items: Items) { if (items.find(fn)) {} }`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`type Items = string[] | {find(predicate: Function): unknown}; function foo(items: Items) { if (items.some(fn)) {} }`)}},
		{Code: `type Collection = {find(predicate: Function): unknown}; if (([] satisfies Collection).find(fn)) {}`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`type Collection = {find(predicate: Function): unknown}; if (([] satisfies Collection).some(fn)) {}`)}},
		{Code: `declare function getItems(): string[]; if (getItems().find(fn)) {}`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`declare function getItems(): string[]; if (getItems().some(fn)) {}`)}},
		{Code: `function foo(array: string[]) { const hasFoo = array.find(fn); if (hasFoo) {} }`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`function foo(array: string[]) { const hasFoo = array.some(fn); if (hasFoo) {} }`)}},
		{Code: `declare function getItems(): string[]; const hasFoo = getItems().find(fn); if (hasFoo) {}`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`declare function getItems(): string[]; const hasFoo = getItems().some(fn); if (hasFoo) {}`)}},
		{Code: `function foo<T>(collection: T) { if (collection.find(fn)) {} }`, Tsx: false, Errors: []rule_tester.InvalidTestCaseError{findError(`function foo<T>(collection: T) { if (collection.some(fn)) {} }`)}},

		// ---- snapshot group: find used as boolean ----
		{
			Code:   `if (array.find(element => element === "🦄")) {}`,
			Errors: []rule_tester.InvalidTestCaseError{findError(`if (array.some(element => element === "🦄")) {}`)},
		},
		{
			Code:   `const foo = array.find(element => element === "🦄") ? bar : baz;`,
			Errors: []rule_tester.InvalidTestCaseError{findError(`const foo = array.some(element => element === "🦄") ? bar : baz;`)},
		},

		// ---- `.filter().length` reported forms ----
		{Code: `array.filter(fn).length > 0`, Output: []string{`array.some(fn)`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},
		{Code: `array.filter(fn).length !== 0`, Output: []string{`array.some(fn)`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},
		// Don't drop comments in the removed part → no fix.
		{Code: `array.filter(fn).length /* keep */ > 0`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},
		// Parentheses around `.length` survive the removal.
		{Code: `((array.filter(fn).length)) > 0`, Output: []string{`((array.some(fn)))`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},
		{Code: `(( array.filter(fn).length )) > 0`, Output: []string{`(( array.some(fn) ))`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},
		{Code: `module$.filter(fn).length > 0`, Output: []string{`module$.some(fn)`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},
		{Code: `type Collection = {filter(predicate: Function): {length: number}}; ([] satisfies Collection).filter(fn).length > 0`, Tsx: false, Output: []string{`type Collection = {filter(predicate: Function): {length: number}}; ([] satisfies Collection).some(fn)`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},
		{Code: `type Items = string[] | {filter(predicate: Function): {length: number}}; function foo(items: Items) { items.filter(fn).length > 0; }`, Tsx: false, Output: []string{`type Items = string[] | {filter(predicate: Function): {length: number}}; function foo(items: Items) { items.some(fn); }`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},
		{Code: `declare function getItems(): string[]; getItems().filter(fn).length > 0;`, Tsx: false, Output: []string{`declare function getItems(): string[]; getItems().some(fn);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},
		// A typed array shares `Array#filter()` and `Array#some()`.
		{Code: `function f(array: Int8Array) { return array.filter(fn).length > 0; }`, Tsx: false, Output: []string{`function f(array: Int8Array) { return array.some(fn); }`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageFilter}}},

		// ---- `.findIndex()` / `.findLastIndex()` compared ----
		{Code: `foo.findIndex(bar) !== -1`, Output: []string{`foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findLastIndex(bar) !== -1`, Output: []string{`foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findIndex(bar) != -1`, Output: []string{`foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findLastIndex(bar) != -1`, Output: []string{`foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findIndex(bar) > - 1`, Output: []string{`foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findLastIndex(bar) > - 1`, Output: []string{`foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findIndex(bar) === -1`, Output: []string{`!foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findLastIndex(bar) === -1`, Output: []string{`!foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findIndex(bar) == - 1`, Output: []string{`!foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findLastIndex(bar) == - 1`, Output: []string{`!foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findIndex(bar) >= 0`, Output: []string{`foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findLastIndex(bar) >= 0`, Output: []string{`foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findIndex(bar) < 0`, Output: []string{`!foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findLastIndex(bar) < 0`, Output: []string{`!foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findIndex(bar) !== (( - 1 ))`, Output: []string{`foo.some(bar) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `foo.findIndex(element => element.bar === 1) !== (( - 1 ))`, Output: []string{`foo.some(element => element.bar === 1) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		// Don't drop comments in the removed part → no fix.
		{Code: `foo.findIndex(bar) !== /* keep */ -1`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `type Collection = {findIndex(predicate: Function): number}; ([] satisfies Collection).findIndex(fn) !== -1`, Tsx: false, Output: []string{`type Collection = {findIndex(predicate: Function): number}; ([] satisfies Collection).some(fn) `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `type Items = string[] | {findIndex(predicate: Function): number}; function foo(items: Items) { items.findIndex(fn) !== -1; }`, Tsx: false, Output: []string{`type Items = string[] | {findIndex(predicate: Function): number}; function foo(items: Items) { items.some(fn) ; }`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
		{Code: `declare function getItems(): string[]; getItems().findIndex(fn) !== -1;`, Tsx: false, Output: []string{`declare function getItems(): string[]; getItems().some(fn) ;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},

		// ---- compare-with-undefined reported forms ----
		{Code: `foo.find(fn) == null`, Errors: []rule_tester.InvalidTestCaseError{findError(`!foo.some(fn)`)}},
		{Code: `foo.find(fn) == undefined`, Errors: []rule_tester.InvalidTestCaseError{findError(`!foo.some(fn)`)}},
		{Code: `foo.find(fn) === undefined`, Errors: []rule_tester.InvalidTestCaseError{findError(`!foo.some(fn)`)}},
		{Code: `foo.find(fn) != null`, Errors: []rule_tester.InvalidTestCaseError{findError(`foo.some(fn)`)}},
		{Code: `foo.find(fn) != undefined`, Errors: []rule_tester.InvalidTestCaseError{findError(`foo.some(fn)`)}},
		{Code: `foo.find(fn) !== undefined`, Errors: []rule_tester.InvalidTestCaseError{findError(`foo.some(fn)`)}},
		{Code: `a = (( ((foo.find(fn))) == ((null)) )) ? "no" : "yes";`, Errors: []rule_tester.InvalidTestCaseError{findError(`a = (( !((foo.some(fn))) )) ? "no" : "yes";`)}},
		// Don't drop comments in the removed comparison part → no suggestion.
		{Code: `foo.find(fn) === /* keep */ undefined`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageSome}}},
	})
}
