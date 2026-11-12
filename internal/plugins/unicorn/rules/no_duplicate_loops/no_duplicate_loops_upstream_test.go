package no_duplicate_loops

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-duplicate-loops.js
func TestNoDuplicateLoopsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDuplicateLoopsRule, []rule_tester.ValidTestCase{
		validCase("const iterator = items.values(); for (const item of iterator.map(callback)) {}", "case.js"),
		validCase("function * generate() { yield 1; } for (const item of generate().filter(callback)) {}", "case.js"),
		validCase("for (const item of items) {}", "case.js"),
		validCase("for await (const item of items) {}", "case.js"),
		validCase("for (const key in items.map(callback)) {}", "case.js"),
		validCase("items.map(callback);", "case.js"),
		validCase("items.filter(callback);", "case.js"),
		validCase("const mapped = items.map(callback); for (const item of mapped) {}", "case.js"),
		validCase("for (const item of items.map()) {}", "case.js"),
		validCase("for (const item of items.filter()) {}", "case.js"),
		validCase("for (const item of items?.map(callback)) {}", "case.js"),
		validCase("for (const item of items.map?.(callback)) {}", "case.js"),
		validCase("for (const item of items[\"map\"](callback)) {}", "case.js"),
		validCase("for (const item of items[method](callback)) {}", "case.js"),
		validCase("for (const [index, item] of items.entries()) {}", "case.js"),
		validCase("for (const [index, item] of items.map(callback).entries()) {}", "case.js"),
		validCase("for (const [index, item] of items.filter(callback).entries()) {}", "case.js"),
		validCase("for (const item of Array.from(items)) {}", "case.js"),
		validCase("for (const item of [...items]) {}", "case.js"),
		validCase("for (const item of Array.prototype.map.call(items, callback)) {}", "case.js"),
		validCase("for (const item of Iterator.from(items).map(callback)) {}", "case.js"),
		validCase("for (const item of Iterator.from(items).filter(callback)) {}", "case.js"),
		validCase("for (const item of Iterator.from(items).map(callback).filter(callback)) {}", "case.js"),
		validCase("for (const item of (Iterator).from(items).map(callback)) {}", "case.js"),
		validCase("for (const item of Iterator.concat(first, second).filter(callback)) {}", "case.js"),
		validCase("for (const item of Iterator.zip(iterables).map(callback)) {}", "case.js"),
		validCase("for (const item of Iterator.zipKeyed(iterables).filter(callback)) {}", "case.js"),
		validCase("for (const item of globalThis.Iterator.from(items).map(callback)) {}", "case.js"),
		validCase("for (const item of (globalThis).Iterator.from(items).map(callback)) {}", "case.js"),
		validCase("for (const item of globalThis.Iterator.zip(iterables).filter(callback)) {}", "case.js"),
		validCase("for (const item of items.values().map(callback)) {}", "case.js"),
		validCase("for (const item of items.keys().map(callback)) {}", "case.js"),
		validCase("for (const item of items.entries().filter(callback)) {}", "case.js"),
		validCase("for (const item of items.matchAll(pattern).map(callback)) {}", "case.js"),
		validCase("for (const item of Iterator.from(items).take(1).map(callback)) {}", "case.js"),
		validCase("for (const item of Iterator.from(items).flatMap(callback).drop(1).filter(callback)) {}", "case.js"),
		validCase("function foo(items: string[]) { for (const item of items) {} }", "case.ts"),
		validCase("function foo(iterator: IteratorObject<string> | undefined) { for (const item of iterator!.map(callback)) {} }", "case.ts"),
		validCase("for (const item of (Iterator as typeof Iterator).from(items).map(callback)) {}", "case.ts"),
		validCase("for (const item of (Iterator.from(items) as Iterator<string>).map(callback)) {}", "case.ts"),
		validCase("for (const item of (items as Iterator<string>).map(callback)) {}", "case.ts"),
		// Documentation examples (https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-duplicate-loops.md)
		validCase(`for (const value of values) {
	console.log(value.property);
}`, "case.js"),
		validCase(`for (const value of values) {
	if (value.isEnabled) {
		console.log(value);
	}
}`, "case.js"),
	}, []rule_tester.InvalidTestCase{
		invalidCase("for (const item of items.map(callback)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 26, 29)),
		invalidCase("for (const item of items.filter(callback)) {}", "case.js", loopDiagnostic("filter", "for…of", 1, 26, 32)),
		invalidCase("for (const item of items.map(callback, thisArgument)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 26, 29)),
		invalidCase("for (const item of items.filter(callback, thisArgument)) {}", "case.js", loopDiagnostic("filter", "for…of", 1, 26, 32)),
		invalidCase("for (const item of (items).map(callback)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 28, 31)),
		invalidCase("for (const item of ((items)).filter(callback)) {}", "case.js", loopDiagnostic("filter", "for…of", 1, 30, 36)),
		invalidCase("for (const item of (items.map(callback))) {}", "case.js", loopDiagnostic("map", "for…of", 1, 27, 30)),
		invalidCase("for await (const item of items.map(callback)) {}", "case.js", loopDiagnostic("map", "for await…of", 1, 32, 35)),
		invalidCase("for await (const item of items.filter(async item => item.isEnabled)) {}", "case.js", loopDiagnostic("filter", "for await…of", 1, 32, 38)),
		invalidCase("for (const item of items.map(/* comment */ callback)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 26, 29)),
		invalidCase("const Iterator = {from: items => items}; for (const item of Iterator.from(items).map(callback)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 82, 85)),
		invalidCase("const Iterator = {zip: iterables => iterables}; for (const item of Iterator.zip(iterables).filter(callback)) {}", "case.js", loopDiagnostic("filter", "for…of", 1, 92, 98)),
		invalidCase("const globalThis = {Iterator: {from: items => items}}; for (const item of globalThis.Iterator.from(items).map(callback)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 107, 110)),
		invalidCase(`for (const item of items
	.map(item => item.value)
) {}`, "case.js", loopDiagnostic("map", "for…of", 2, 3, 6)),
		invalidCase(`for (
	const item of items.filter(item => item.isEnabled)
) {}`, "case.js", loopDiagnostic("filter", "for…of", 2, 22, 28)),
		invalidCase("function foo(items: string[]) { for (const item of (items as string[]).map(item => item.trim())) {} }", "case.ts", loopDiagnostic("map", "for…of", 1, 72, 75)),
		invalidCase("for (const item of items.map(callback) as string[]) {}", "case.ts", loopDiagnostic("map", "for…of", 1, 26, 29)),
		// Documentation examples (https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-duplicate-loops.md)
		invalidCase(`for (const value of values.map(value => value.property)) {
	console.log(value);
}`, "case.js", loopDiagnostic("map", "for…of", 1, 28, 31)),
		invalidCase(`for (const value of values.filter(value => value.isEnabled)) {
	console.log(value);
}`, "case.js", loopDiagnostic("filter", "for…of", 1, 28, 34)),
	})
}

var module = rule.LanguageOptions{SourceType: "module"}

func validCase(code, filename string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: filename, LanguageOptions: module}
}

func invalidCase(code, filename string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, FileName: filename, LanguageOptions: module, Errors: errors}
}

func loopDiagnostic(method, kind string, line, column, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "no-duplicate-loops",
		Message:   "Do not use `." + method + "()` directly in a `" + kind + "` loop header. It creates an intermediate array before the loop.",
		Line:      line, Column: column, EndLine: line, EndColumn: endColumn,
	}
}
