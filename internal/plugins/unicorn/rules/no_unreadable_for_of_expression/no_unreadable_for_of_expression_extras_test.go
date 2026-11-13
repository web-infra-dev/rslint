package no_unreadable_for_of_expression

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnreadableForOfExpressionExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnreadableForOfExpressionRule, []rule_tester.ValidTestCase{
		validCase("for (const x of (getItems)(getArgument())) {}", "case.js"),
		validCase("for (const x of combine((getArgument)(), (object).value, true, false, null, 1n, /x/, this)) {}", "case.js"),
		validCase("for (const x of [compute(seed), ...more]) {}", "case.js"),
		validCase("for (const x of items.map()) {}", "case.js"),
		validCase("for (const x of items.filter(...args)) {}", "case.js"),
		validCase("class C { #map() {} m() { for (const x of this.#map(fn)) {} } }", "case.js"),
		// Only the first member in this chain uses optional access, so the final
		// .map() is handled by no-duplicate-loops.
		validCase("for (const x of items?.list.map(fn)) {}", "case.js"),
		validCase("for (const x in (head(), items)) {}", "case.js"),
		validCase("for (const x of Iterator.from(items).take()) {}", "case.js"),
		validCase("for (const x of items.matchAll(...args).map(fn)) {}", "case.js"),
		validCase("for (const x of items.values(arg).filter(fn)) {}", "case.js"),
		validCase("for (const x of Object.entries(items).map(fn)) {}", "case.js"),
		validCase("for (const x of globalThis?.Iterator.from(items).map(fn)) {}", "case.js"),
		validCase("let it = items.values(); it = items; for (const x of it.map(fn)) {}", "case.js"),
		validCase("const a = b, b = a; for (const x of a.map(fn)) {}", "case.js"),
		validCase("async function* values() { yield 1; } for await (const x of values().filter(fn)) {}", "case.js"),
		validCase("for (const x of /** @type {Iterable<number>} */ (items)) {}", "case.js"),
		validCase("for (const x of <Iterable<number>>items) {}", "case.ts"),
		validCase("for (const x of (getItems)<number>()) {}", "case.ts"),
		validCase("for (const x of getItems<number>) {}", "case.ts"),
		validCase("for (const x of items) { use(<Row key={x.id} />); }", "case.tsx"),
		{
			Code: "for (const x of Iterator.from(items).map(fn)) {}", FileName: "case.js",
			LanguageOptions: rule.LanguageOptions{SourceType: "module", ECMAVersion: 2024},
		},
		{
			Code: "for (const x of Iterator.from(items).map(fn)) {}", FileName: "case.js",
			LanguageOptions: module,
			Globals:         map[string]any{"Iterator": "off"},
		},
	}, []rule_tester.InvalidTestCase{
		invalidCase("for (const x of (items?.map)(fn)) {}", "case.js", loopDiagnostic("for…of", 1, 17, 1, 33)),
		invalidCase("for (const x of combine((items?.get)())) {}", "case.js", loopDiagnostic("for…of", 1, 17, 1, 40)),
		invalidCase("for (const x of items[\"map\"](fn)) {}", "case.js", loopDiagnostic("for…of", 1, 17, 1, 33)),
		invalidCase("for (const x of items.map(fn).entries()) {}", "case.js", loopDiagnostic("for…of", 1, 17, 1, 40)),
		invalidCase("for (const x of combine(...args)) {}", "case.js", loopDiagnostic("for…of", 1, 17, 1, 33)),
		invalidCase("for (const x of combine(import(\"data.js\"))) {}", "case.js", loopDiagnostic("for…of", 1, 17, 1, 43)),
		invalidCase("for (const x of combine(object[index])) {}", "case.js", loopDiagnostic("for…of", 1, 17, 1, 39)),
		invalidCase("for (const x of `prefix-${items}`) {}", "case.js", loopDiagnostic("for…of", 1, 17, 1, 34)),
		invalidCase("for (const x of (head(), items)) {}", "case.js", loopDiagnostic("for…of", 1, 18, 1, 31)),
		invalidCase("const label = \"🌱\"; for (const x of (/* note */ first || fallback)) {}", "case.js", loopDiagnostic("for…of", 1, 49, 1, 66)),
		invalidCase(`for (const x of (
  getItems(
    createArgument(seed),
  )
)) {}`, "case.js", loopDiagnostic("for…of", 2, 3, 4, 4)),
		invalidCase("for (const x of Iterator.from(items).take(...args)) {}", "case.js", loopDiagnostic("for…of", 1, 17, 1, 51)),
		invalidCase("for (const x of Iterator.from(items).map?.(fn)) {}", "case.js", loopDiagnostic("for…of", 1, 17, 1, 47)),
		invalidCase("function* values() { yield 1; } const alias = values; for (const x of alias().filter(fn)) {}", "case.js", loopDiagnostic("for…of", 1, 71, 1, 89)),
		invalidCase("for (const x of (function* () { yield 1; })().map(fn)) {}", "case.js", loopDiagnostic("for…of", 1, 17, 1, 54)),
		invalidCase("for (const x of /** @type {Iterable<number>} */ (first || fallback)) {}", "case.js", loopDiagnostic("for…of", 1, 50, 1, 67)),
		// A type wrapper can be unwrapped at the loop or argument boundary,
		// but a wrapped callee is itself part of the call's complexity.
		invalidCase("for (const x of (getItems<number>)()) {}", "case.ts", loopDiagnostic("for…of", 1, 17, 1, 37)),
		invalidCase("for (const x of (getItems as typeof getItems)()) {}", "case.ts", loopDiagnostic("for…of", 1, 17, 1, 48)),
		invalidCase("for (const x of getItems!(seed)) {}", "case.ts", loopDiagnostic("for…of", 1, 17, 1, 32)),
		invalidCase("for (const x of (items?.get)!()) {}", "case.ts", loopDiagnostic("for…of", 1, 17, 1, 32)),
		invalidCase("for (const x of ((first || second) as Iterable<number>)) {}", "case.ts", loopDiagnostic("for…of", 1, 18, 1, 55)),
		invalidCase("function consume(it: IteratorObject<number> | undefined) { for (const x of it!.map(fn)) {} }", "case.ts", loopDiagnostic("for…of", 1, 76, 1, 87)),
		invalidCase("for (const x of rows(<Row />)) {}", "case.tsx", loopDiagnostic("for…of", 1, 17, 1, 30)),
		// The TypeScript project exposes the iterator return type of this method.
		invalidCase(`declare const factories: { rows(): IteratorObject<number> };
for (const x of factories.rows().map(fn)) {}`, "case.ts", loopDiagnostic("for…of", 2, 17, 2, 41)),
	})
}
