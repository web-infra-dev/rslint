package no_unreadable_for_of_expression

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-unreadable-for-of-expression.js
func TestNoUnreadableForOfExpressionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnreadableForOfExpressionRule, []rule_tester.ValidTestCase{
		// Template literals
		validCase("for (const character of `unicorn`) {}", "case.js"),
		validCase("for await (const character of ``) {}", "case.js"),
		validCase("for (const character of `uni\\u0063orn`) {}", "case.js"),
		// Upstream snapshot cases
		validCase("for (const [name, score] of Iterator.zip([names, scores])) {}", "case.js"),
		validCase("for (const item of combine([first, object.second, getThird(), \"fourth\"])) {}", "case.js"),
		validCase("for (const item of combine([])) {}", "case.js"),
		validCase("for (const item of items) {}", "case.js"),
		validCase("for await (const item of items) {}", "case.js"),
		validCase("for (const item of object.items) {}", "case.js"),
		validCase("for (const item of this.items) {}", "case.js"),
		validCase("class Foo extends Bar { method() { for (const item of super.items) {} } }", "case.js"),
		validCase("for (const item of [first, second]) {}", "case.js"),
		validCase("for (const [index, item] of [first, second].entries()) {}", "case.js"),
		validCase("for (const item of getItems()) {}", "case.js"),
		validCase("for (const item of getItems(argument)) {}", "case.js"),
		validCase("for (const item of getItems(first, object.second, \"third\")) {}", "case.js"),
		validCase("for (const item of getItems(`argument`)) {}", "case.js"),
		validCase("for (const item of getItems(createArgument())) {}", "case.js"),
		validCase("for (const character of \"abc\") {}", "case.js"),
		validCase("for (const part of \"a,b\".split(\",\")) {}", "case.js"),
		validCase("for (const item of getItems(`prefix-${argument}`)) {}", "case.js"),
		validCase("for (const element of root.querySelectorAll(`[data-x=${value}]`)) {}", "case.js"),
		validCase("for (const item of object.getItems(argument)) {}", "case.js"),
		validCase("for (const item of object.getItems()) {}", "case.js"),
		validCase("for (const item of Object.keys(object)) {}", "case.js"),
		validCase("for (const item of Object.values(object.items)) {}", "case.js"),
		validCase("for (const [key, value] of Object.entries(getObject())) {}", "case.js"),
		validCase("const Object = {keys: value => value}; for (const item of Object.keys(object)) {}", "case.js"),
		validCase("for (const item of (items)) {}", "case.js"),
		validCase("for (const item of ((object.items))) {}", "case.js"),
		validCase("for (const item of items.map(item => item.value)) {}", "case.js"),
		validCase("for (const item of items.map(callback)) {}", "case.js"),
		validCase("for (const item of (items.map(callback))) {}", "case.js"),
		validCase("for (const item of items.filter(item => item.isEnabled)) {}", "case.js"),
		validCase("for await (const item of items.map(callback)) {}", "case.js"),
		validCase("for await (const item of items.filter(callback)) {}", "case.js"),
		validCase("const Iterator = {from: value => value}; for (const item of Iterator.from(items).map(callback)) {}", "case.js"),
		validCase("const Iterator = {concat: value => value}; for (const item of Iterator.concat(items).map(callback)) {}", "case.js"),
		validCase("const Iterator = {zip: value => value}; for (const item of Iterator.zip(items).filter(callback)) {}", "case.js"),
		validCase("const Iterator = {zipKeyed: value => value}; for (const item of Iterator.zipKeyed(items).take(1)) {}", "case.js"),
		validCase("const globalThis = {Iterator}; for (const item of globalThis.Iterator.from(items).map(callback)) {}", "case.js"),
		validCase("for (const item of items as string[]) {}", "case.ts"),
		validCase("for (const item of (items satisfies Iterable<string>)) {}", "case.ts"),
		validCase("for (const item of items!) {}", "case.ts"),
		validCase("for (const item of getItems() as string[]) {}", "case.ts"),
		validCase("for (const item of items.map(callback) as string[]) {}", "case.ts"),
		validCase("for (const item of items.filter(callback) as string[]) {}", "case.ts"),
		validCase("for (const pair of Iterator.zip([names, scores] as const)) {}", "case.ts"),
		validCase("for (const pair of Iterator.zip([names, scores] satisfies Iterable<unknown>[])) {}", "case.ts"),
		validCase("for (const pair of Iterator.zip(([names, scores])!)) {}", "case.ts"),
		validCase("for (const pair of Iterator.zip([names as string[], scores!])) {}", "case.ts"),
		validCase("class Foo { #items; method() { for (const item of this.#items) {} } }", "case.js"),
		// Documentation examples (https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-unreadable-for-of-expression.md)
		validCase(`const argument = createArgument(seed);
for (const item of getItems(argument)) {
	console.log(item);
}`, "case.js"),
		validCase(`const object = Object.fromEntries(entries);
for (const key of Object.keys(object)) {
	console.log(key);
}`, "case.js"),
		validCase(`for (const item of getItems(argument)) {
	console.log(item);
}`, "case.js"),
		validCase(`for (const item of items) {
	console.log(item);
}`, "case.js"),
		validCase(`for (const value of Object.values(object)) {
	console.log(value);
}`, "case.js"),
	}, []rule_tester.InvalidTestCase{
		// Upstream snapshot cases
		invalidCase("for (const pair of Iterator.zip([names, getScores(seed)])) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 58)),
		invalidCase("for (const pair of Iterator.zip([...inputs])) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 45)),
		invalidCase("for (const pair of Iterator.zip([[first], [second]])) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 53)),
		invalidCase("for (const pair of Iterator.zip([first, , second])) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 51)),
		invalidCase("for (const item of getItems(createArgument(seed))) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 50)),
		invalidCase("for (const item of getItems(argument || fallback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 50)),
		invalidCase("for (const item of Object.keys({a: 1})) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 39)),
		invalidCase("for (const item of (header || \"\").split(\",\")) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 45)),
		invalidCase("for (const item of (condition ? first : second).values()) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 57)),
		invalidCase("for (const item of items[method]()) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 35)),
		invalidCase("for (const item of items[index]) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 32)),
		invalidCase("for (const item of items?.values()) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 35)),
		invalidCase("for (const item of condition ? first : second) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 46)),
		invalidCase("for (const item of first || second) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 35)),
		invalidCase("for (const item of new Set(items)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 34)),
		invalidCase("for (const item of tag`items`) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 30)),
		invalidCase("for (const item of Object.keys(Object.fromEntries(entries))) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 60)),
		invalidCase("for (const item of Object.values(items.filter(item => item.isEnabled))) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 71)),
		invalidCase("for (const item of items?.map(item => item.value)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 50)),
		invalidCase("for (const item of items?.map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 40)),
		invalidCase("for (const item of items.filter?.(item => item.isEnabled)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 58)),
		invalidCase("for (const item of items.filter?.(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 44)),
		invalidCase("for (const item of globalThis.Iterator.from(items).map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 65)),
		invalidCase("for (const item of (globalThis).Iterator.from(items).map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 67)),
		invalidCase("for (const item of (Iterator).from(items).map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 56)),
		invalidCase("for (const item of Iterator.from(items).map(item => item.value)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 64)),
		invalidCase("for (const item of Iterator.from(items).map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 54)),
		invalidCase("for (const item of Iterator.from(items).filter(item => item.isEnabled)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 71)),
		invalidCase("for (const item of Iterator.from(items).filter(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 57)),
		invalidCase("for (const item of Iterator.from(items).take(1)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 48)),
		invalidCase("for (const item of Iterator.from(items).drop(1)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 48)),
		invalidCase("for (const item of Iterator.from(items).flatMap(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 58)),
		invalidCase("for (const item of Iterator.concat(first, second).map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 64)),
		invalidCase("for (const item of Iterator.zip(iterables).filter(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 60)),
		invalidCase("for (const item of Iterator.zipKeyed(iterables).take(1)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 56)),
		invalidCase("for (const item of Iterator.from(items).take(1).map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 62)),
		invalidCase("for (const item of items.values().map(item => item.value)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 58)),
		invalidCase("for (const item of items.values().map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 48)),
		invalidCase("for (const item of items.values().filter(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 51)),
		invalidCase("for (const item of items.values().take(1)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 42)),
		invalidCase("for (const item of items.values().take(1).map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 56)),
		invalidCase("for (const item of items.keys().map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 46)),
		invalidCase("for (const item of items.entries().filter(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 52)),
		invalidCase("for (const item of items.matchAll(pattern).map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 20, 1, 57)),
		invalidCase("for await (const item of getItems(createArgument(seed))) {}", "case.js", loopDiagnostic("for await…of", 1, 26, 1, 56)),
		invalidCase("for (const item of getItems(createArgument(seed)) as string[]) {}", "case.ts", loopDiagnostic("for…of", 1, 20, 1, 62)),
		invalidCase("for (const item of (Iterator as typeof Iterator).from(items).map(callback)) {}", "case.ts", loopDiagnostic("for…of", 1, 20, 1, 75)),
		invalidCase("for (const item of (Iterator.from(items) as Iterator<string>).map(callback)) {}", "case.ts", loopDiagnostic("for…of", 1, 20, 1, 76)),
		invalidCase("for (const item of Iterator.from(items).take(1) as Iterator<string>) {}", "case.ts", loopDiagnostic("for…of", 1, 20, 1, 68)),
		invalidCase("for (const item of (items[index] satisfies Iterable<string>)) {}", "case.ts", loopDiagnostic("for…of", 1, 21, 1, 60)),
		invalidCase("for (const pair of Iterator.zip([names, getScores(seed)] as const)) {}", "case.ts", loopDiagnostic("for…of", 1, 20, 1, 67)),
		invalidCase("for (const pair of Iterator.zip([[names], [scores]] as const)) {}", "case.ts", loopDiagnostic("for…of", 1, 20, 1, 62)),
		invalidCase("for (const pair of Iterator.zip([getNames(seed) as string[], scores])) {}", "case.ts", loopDiagnostic("for…of", 1, 20, 1, 70)),
		invalidCase(`for (
	const item of getItems(createArgument(seed))
) {}`, "case.js", loopDiagnostic("for…of", 2, 16, 2, 46)),
		invalidCase("const iterator = items.values(); for (const item of iterator.map(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 53, 1, 75)),
		invalidCase("function * generate() { yield 1; } for (const item of generate().filter(callback)) {}", "case.js", loopDiagnostic("for…of", 1, 55, 1, 82)),
		// Documentation examples (https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-unreadable-for-of-expression.md)
		invalidCase(`for (const item of getItems(createArgument(seed))) {
	console.log(item);
}`, "case.js", loopDiagnostic("for…of", 1, 20, 1, 50)),
		invalidCase(`for (const key of Object.keys(Object.fromEntries(entries))) {
	console.log(key);
}`, "case.js", loopDiagnostic("for…of", 1, 19, 1, 59)),
	})
}

var module = rule.LanguageOptions{SourceType: "module"}

func validCase(code, filename string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: filename, LanguageOptions: module}
}

func invalidCase(code, filename string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, FileName: filename, LanguageOptions: module, Errors: errors}
}

func loopDiagnostic(kind string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "no-unreadable-for-of-expression",
		Message:   "Move the complex iterable expression out of the `" + kind + "` loop header.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}
