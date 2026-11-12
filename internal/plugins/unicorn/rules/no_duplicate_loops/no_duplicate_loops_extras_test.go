package no_duplicate_loops

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDuplicateLoopsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDuplicateLoopsRule, []rule_tester.ValidTestCase{
		// Access forms.
		validCase("for (const x of (items?.map)(fn)) {}", "case.js"),
		validCase("for (const x of items[`map`](fn)) {}", "case.js"),
		validCase("class Box { #map(fn) { return []; } loop() { for (const x of this.#map(fn)) {} } }", "case.js"),
		// Iterator sources and helper argument bounds.
		validCase("for (const x of items.values?.().map(fn)) {}", "case.js"),
		validCase("for (const x of Iterator?.from(items).filter(fn)) {}", "case.js"),
		validCase("for (const x of items.matchAll().map(fn)) {}", "case.js"),
		// Local bindings and generator calls.
		validCase("const it = items.values(); const chain = it.filter(fn); for (const x of chain.map(fn)) {}", "case.js"),
		validCase("function* values() { yield 1; } const alias = values; for (const x of alias().filter(fn)) {}", "case.js"),
		validCase("for (const x of (function* () { yield 1; })().map(fn)) {}", "case.js"),
		// Authored JavaScript types.
		validCase(`/** @param {Iterator<number> | undefined} it */
function consume(it) { for (const x of it.filter(fn)) {} }`, "case.js"),
		validCase(`/** @type {IteratorObject<number>} */
const it = make();
for (const x of it.map(fn)) {}`, "case.js"),
		validCase(`/** @returns {Iterator<number>} */
function make() {}
for (const x of make().map(fn)) {}`, "case.js"),
		validCase(`/** @type {function(): Iterator<number>} */
const make = factory;
for (const x of make().map(fn)) {}`, "case.js"),
		validCase(`/** @type {Iterator<number>} */
const it = make();
for (const x of (before(), it).map(fn)) {}
for (const x of (temp = it).map(fn)) {}`, "case.js"),
		// TypeScript expression and type forms.
		validCase("for (const x of (items.map as typeof items.map)(fn)) {}", "case.ts"),
		validCase("type Rows = IteratorObject<number> | null; function consume(it: Rows) { for (const x of it!.map(fn)) {} }", "case.ts"),
		validCase("interface Rows extends IteratorObject<number> {} function consume(it: Rows) { for (const x of it.map(fn)) {} }", "case.ts"),
		validCase("function consume(it: IteratorObject<number>) { for (const x of (it as any).map(fn)) {} }", "case.ts"),
		validCase("const make = ((): IteratorObject<number> => factory()) satisfies Function; for (const x of make().map(fn)) {}", "case.ts"),
		// A TypeScript project can resolve the return type of a member call.
		validCase(`declare const factories: { rows(): IteratorObject<number> };
for (const x of factories.rows().map(fn)) {}`, "case.ts"),
	}, []rule_tester.InvalidTestCase{
		// Access forms.
		invalidCase("for (const x of items?.list.map(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 29, 32)),
		invalidCase("for (const x of (items.map)(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 24, 27)),
		invalidCase("for (const x of items.map(...args)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 23, 26)),
		invalidCase("const label = \"🌱\"; for (const x of items.m\\u0061p(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 43, 51)),
		// Iterator sources and helper argument bounds.
		invalidCase("for (const x of items.matchAll(...args).map(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 41, 44)),
		invalidCase("for (const x of Object.entries(items).map(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 39, 42)),
		invalidCase("for (const x of items.values(arg).map(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 35, 38)),
		invalidCase("for (const x of Iterator.from(items).drop().map(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 45, 48)),
		invalidCase("for (const x of Iterator[\"from\"](items).filter(fn)) {}", "case.js", loopDiagnostic("filter", "for…of", 1, 41, 47)),
		invalidCase("for (const x of globalThis?.Iterator.from(items).map(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 50, 53)),
		invalidCase("for (const x of globalThis[\"Iterator\"].from(items).map(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 52, 55)),
		// Local bindings and generator calls.
		invalidCase("const a = b, b = a; for (const x of a.map(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 39, 42)),
		invalidCase("let it = items.values(); it = items; for (const x of it.map(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 57, 60)),
		invalidCase(`function* values() { yield 1; }
for (const x of values().map(fn)) {}
{
  function* values() { yield 2; }
  [values] = other;
  for (const x of values().filter(fn)) {}
}
for (const x of values().map(fn)) {}`, "case.js", loopDiagnostic("filter", "for…of", 6, 28, 34)),
		invalidCase("async function* values() { yield 1; } for await (const x of values().filter(fn)) {}", "case.js", loopDiagnostic("filter", "for await…of", 1, 70, 76)),
		invalidCase("function* values() { yield 1; } for (const x of values?.().map(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 60, 63)),
		invalidCase("const { it } = { it: items.values() }; for (const x of it.map(fn)) {}", "case.js", loopDiagnostic("map", "for…of", 1, 59, 62)),
		// Authored JavaScript types.
		invalidCase(`/** @type {Iterator<number> | number[]} */
const it = make();
for (const x of it.map(fn)) {}`, "case.js", loopDiagnostic("map", "for…of", 3, 20, 23)),
		// TypeScript expression and type forms.
		invalidCase("for (const x of <unknown[]>items.map(fn)) {}", "case.ts", loopDiagnostic("map", "for…of", 1, 34, 37)),
		invalidCase("for (const x of (items.filter(fn) satisfies unknown[])) {}", "case.ts", loopDiagnostic("filter", "for…of", 1, 24, 30)),
		invalidCase("for (const x of (items.map(fn))<number>) {}", "case.ts", loopDiagnostic("map", "for…of", 1, 24, 27)),
		invalidCase("function consume(it: Iterator<number> | number[]) { for (const x of it.map(fn)) {} }", "case.ts", loopDiagnostic("map", "for…of", 1, 72, 75)),
		invalidCase("type Iterator = { map(fn: Function): number[] }; function consume(it: Iterator) { for (const x of it.map(fn)) {} }", "case.ts", loopDiagnostic("map", "for…of", 1, 102, 105)),
		invalidCase("const Iterator = {}; function consume(it: Iterator<number>) { for (const x of it.map(fn)) {} }", "case.ts", loopDiagnostic("map", "for…of", 1, 82, 85)),
		// TSX loop body.
		invalidCase("for (const x of items.filter(fn)) { use(<Row key={x.id} />); }", "case.tsx", loopDiagnostic("filter", "for…of", 1, 23, 29)),
		// Iterator is an ES2025 global and can be disabled explicitly.
		{
			Code: "for (const x of Iterator.from(items).map(fn)) {}", FileName: "case.js",
			LanguageOptions: rule.LanguageOptions{SourceType: "module", ECMAVersion: 2024},
			Errors:          []rule_tester.InvalidTestCaseError{loopDiagnostic("map", "for…of", 1, 38, 41)},
		},
		{
			Code: "for (const x of Iterator.from(items).map(fn)) {}", FileName: "case.js",
			LanguageOptions: module,
			Globals:         map[string]any{"Iterator": "off"},
			Errors:          []rule_tester.InvalidTestCaseError{loopDiagnostic("map", "for…of", 1, 38, 41)},
		},
	})
}
