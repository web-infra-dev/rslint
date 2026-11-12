// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
package no_mismatched_map_key_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_mismatched_map_key"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-mismatched-map-key.js
func TestNoMismatchedMapKeyUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Guard and access ordering.
		{FileName: "case.js", Code: "const value = map.get(otherKey) && map.has(key);"},
		{FileName: "case.js", Code: "const value = map.get(otherKey) || map.has(key);"},
		{FileName: "case.js", Code: "if (map.has(key) && condition) { map.get(anotherKey); }"},
		{FileName: "case.js", Code: "do { map.get(anotherKey); } while (map.has(key));"},
		{FileName: "case.js", Code: "if (condition || map.has(key)) { map.get(anotherKey); }"},

		// Receiver writes, matching keys and an explicitly typed non-Map receiver.
		{FileName: "case.js", Code: "class A { f() { if (this.map.has(key)) { (this.map = other, this.map.get(anotherKey)); } } }"},
		{FileName: "case.js", Code: "class A { f() { if (this.map.has(key)) { for (this.map of maps) { this.map.get(anotherKey); } } } }"},
		{FileName: "case.js", Code: "const value = map.has(key) && map.get(key);"},
		{FileName: "case.js", Code: "if (map.has(key)) { (map++, map.get(anotherKey)); }"},
		{FileName: "case.ts", Code: `function f(cache: {has(key: string): boolean; get(key: string): number}) { if (cache.has("a")) { return cache.get("b"); } }`},
		{FileName: "case.js", Code: "map.has(key) ? map.get(key) : value;"},
		{FileName: "case.js", Code: "!map.has(key) ? map.set(key, value) : value;"},
		{FileName: "case.js", Code: "if (map.has(key)) { const value = map.get(key); }"},
		{FileName: "case.js", Code: "if (!map.has(key)) { map.set(key, value); }"},

		// Other receivers, calls in keys and unsupported method call forms.
		{FileName: "case.js", Code: "if (map.has(key)) { otherMap.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { map.get(getKey()); }"},
		{FileName: "case.js", Code: "if (map.has(getKey())) { map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has(...key)) { map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { map.set(anotherKey, ...values); }"},
		{FileName: "case.js", Code: "if (map?.has(key)) { map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has?.(key)) { map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map[has](key)) { map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { map[get](anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has(key) && other) { map.get(anotherKey); }"},

		// Equal static keys and a call in a key.
		{FileName: "case.js", Code: "if (map.has(1 + 1)) { map.get(2); }"},
		{FileName: "case.js", Code: "if (map.has(NaN)) { map.get(NaN); }"},
		{FileName: "case.js", Code: "if (map.has(-0)) { map.get(0); }"},
		{FileName: "case.js", Code: `if (map.has(String("key"))) { map.get("anotherKey"); }`},

		// Receiver shadowing and writes.
		{FileName: "case.js", Code: "if (map.has(key)) { const map = otherMap, value = map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (cache.map.has(key)) { const cache = otherCache, value = cache.map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { const {map} = cache, value = map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { try {} catch (map) { map.get(anotherKey); } }"},
		{FileName: "case.js", Code: "if (map.has(key)) { for (map of maps) { map.get(anotherKey); } }"},
		{FileName: "case.js", Code: "if (map.has(key)) { map = otherMap; map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { var map = otherMap; map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { (map.get(anotherKey), map = otherMap); }"},
		{FileName: "case.js", Code: "if (cache.map.has(key)) { (cache.map = otherMap, cache.map.get(anotherKey)); }"},
		{FileName: "case.js", Code: "if (cache.map.has(key)) { for (cache.map of maps) { cache.map.get(anotherKey); } }"},
		{FileName: "case.js", Code: "if (map.has(key)) { ({map} = other); map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { var {map} = other; map.get(anotherKey); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { for ({map} of maps) { map.get(anotherKey); } }"},

		// Nested guards on the same receiver.
		{FileName: "case.js", Code: "if (map.has(key)) { if (map.has(anotherKey)) { map.get(anotherKey); } }"},
		{FileName: "case.js", Code: "if (map.has(key)) { map.has(anotherKey) ? map.get(anotherKey) : value; }"},
		{FileName: "case.js", Code: "if (map.has(key)) { if (!map.has(anotherKey)) { map.set(anotherKey, value); } }"},
		{FileName: "case.js", Code: "if (map.has(key)) { if (condition && map.has(anotherKey)) { map.get(anotherKey); } }"},
		{FileName: "case.js", Code: "if (map.has(key)) { condition && map.has(anotherKey) ? map.get(anotherKey) : value; }"},

		// Conditional keys, multi-statement branches and nested scope boundaries.
		{FileName: "case.js", Code: `if (map.has(true ? String("key") : "x")) { map.get("anotherKey"); }`},
		{FileName: "case.js", Code: "if (map.has(key)) { map.property = value; map.get(anotherKey); }"},
		{FileName: "case.js", Code: `if (map.has(key)) {
	map.set(key, value);
	map.set(anotherKey, value);
}`},
		{FileName: "case.js", Code: `if (map.has(key)) {
	for (const map of maps) {
		map.get(anotherKey);
	}
}`},
		{FileName: "case.js", Code: `if (map.has(key)) {
	for (const item of items) {
		const map = otherMap;
		map.get(anotherKey);
	}
}`},
		{FileName: "case.js", Code: `if (map.has(key)) {
	for (const item of items) {
		class map {}
		map.get(anotherKey);
	}
}`},
		{FileName: "case.js", Code: `if (map.has(key)) {
	for (const item of items) {
		function map() {}
		map.get(anotherKey);
	}
}`},
		{FileName: "case.js", Code: `if (map.has(key)) {
	function getValue() {
		return map.get(anotherKey);
	}
}`},
		{FileName: "case.js", Code: `if (map.has(key)) {
	class Value {
		static value = map.get(anotherKey);
	}
}`},

		// TypeScript assertions and mutable member reads.
		{FileName: "case.ts", Code: "if (map.has(key as string)) { map.get(key); }"},
		{FileName: "case.js", Code: `const modes = new Set();
modes.add('foo');
if (map.has(modes.size)) {
	map.get(1);
}`},
		{FileName: "case.js", Code: `const modes = new Set(["foo"]); modes.clear(); if (map.has((modes.size && 1) || key)) { map.get(1); }`},
		{FileName: "case.js", Code: `const object = {value: true}; Object.defineProperty(object, "value", {get() { return false; }}); if (map.has(object.value ? 2 : key)) { map.get(1); }`},
		{FileName: "case.ts", Code: `const modes = new Set(["foo"]); modes.clear(); if (map.has((modes.size ? 1 : key) as number)) { map.get(1); }`},

		// Destructuring, non-primitive keys and branch boundaries.
		{FileName: "case.js", Code: "if (map.has(key)) { use([, map] = values, map.get(anotherKey)); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { use({...map} = values, map.get(anotherKey)); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { use([map = fallback] = values, map.get(anotherKey)); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { use([...map] = values, map.get(anotherKey)); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { map.get([anotherKey]); }"},
		{FileName: "case.js", Code: "if (map.has(key)) { map.get(key); } else if (condition) { map.get(anotherKey); }"},
		{FileName: "case.js", Code: "while (map.has(key)) { map.get(anotherKey); next(); }"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// An explicitly typed Map.
		keyCase("case.ts", `function f(map: Map<string, number>) { if (map.has("a")) { return map.get("b"); } }`, 1, 75, 1, 78),

		// Conditional branches and WeakMap-shaped receivers.
		keyCase("case.js", "map.has(key) ? map.get(anotherKey) : value;", 1, 24, 1, 34),
		keyCase("case.js", "map.has(key) ? map.set(anotherKey, value) : value;", 1, 24, 1, 34),
		keyCase("case.js", "if (weakMap.has(key)) { weakMap.get(anotherKey); }", 1, 37, 1, 47),
		keyCase("case.js", "weakMap.has(key) ? weakMap.get(anotherKey) : value;", 1, 32, 1, 42),
		keyCase("case.js", "if (map.has(key)) { const value = map.get(anotherKey); }", 1, 43, 1, 53),
		keyCase("case.js", "if (map.has(key)) { map.set(anotherKey, value); }", 1, 29, 1, 39),
		keyCase("case.js", "if (!map.has(key)) { const value = map.get(anotherKey); }", 1, 44, 1, 54),
		keyCase("case.js", "if (!map.has(key)) { map.set(anotherKey, value); }", 1, 30, 1, 40),
		keyCase("case.js", "map.has(key) ? value : map.get(anotherKey);", 1, 32, 1, 42),
		keyCase("case.js", "!map.has(key) ? value : map.get(anotherKey);", 1, 33, 1, 43),
		keyCase("case.js", "if (map.has(key)) {} else { map.get(anotherKey); }", 1, 37, 1, 47),
		keyCase("case.js", "if (!map.has(key)) {} else { map.get(anotherKey); }", 1, 38, 1, 48),

		// Nested branches and shadowed key bindings.
		keyCase("case.js", "if (map.has(key)) { for (const item of items) { map.get(anotherKey); } }", 1, 57, 1, 67),
		keyCase("case.js", "map.has(key) ? (map.has(key) ? map.get(anotherKey) : value) : value;", 1, 40, 1, 50),
		keyCase("case.js", "if (map.has(key)) { if (condition) { map.get(anotherKey); } }", 1, 46, 1, 56),
		keyCase("case.js", "if (map.has(key)) { condition ? map.get(anotherKey) : value; }", 1, 41, 1, 51),
		keyCase("case.js", "if (map.has(key)) { const key = anotherKey, value = map.get(key); }", 1, 61, 1, 64),
		keyCase("case.js", "if (map.has(object.key)) { const object = other, value = map.get(object.key); }", 1, 66, 1, 76),

		// Literal and member keys, including a TypeScript assertion.
		keyCase("case.js", `if (map.has("key")) { map.get("anotherKey"); }`, 1, 31, 1, 43),
		keyCase("case.js", "if (map.has(1)) { map.get(2); }", 1, 27, 1, 28),
		keyCase("case.js", "if (map.has(object.key)) { map.get(object.anotherKey); }", 1, 36, 1, 53),
		keyCase("case.js", "if (this.map.has(this.key)) { this.map.get(this.anotherKey); }", 1, 44, 1, 59),
		keyCase("case.ts", "if (map.has(key as string)) { map.get(anotherKey as string); }", 1, 39, 1, 59),

		// The final && operand, while loops and short-circuit expressions.
		keyCase("case.js", "if (condition && map.has(key)) { map.get(anotherKey); }", 1, 42, 1, 52),
		keyCase("case.js", "while (map.has(key)) { map.get(anotherKey); }", 1, 32, 1, 42),
		keyCase("case.js", "while (!map.has(key)) { map.set(anotherKey, value); }", 1, 33, 1, 43),
		keyCase("case.js", "while (map.has(key)) { const x = map.get(anotherKey); }", 1, 42, 1, 52),
		keyCase("case.js", "const value = map.has(key) && map.get(otherKey);", 1, 39, 1, 47),
		keyCase("case.js", "if (!map.has(key) && map.get(otherKey)) {}", 1, 30, 1, 38),
		keyCase("case.js", "const value = map.has(key) || map.get(otherKey);", 1, 39, 1, 47),

		// An unbraced branch, unrelated writes and a nested function.
		keyCase("case.js", "if (map.has(key)) map.get(anotherKey);", 1, 27, 1, 37),
		keyCase("case.js", "if (map.has(key)) { use([object.value] = values, map.get(anotherKey)); }", 1, 58, 1, 68),
		keyCase("case.js", "if (map.has(key)) { (count = 1, map.get(anotherKey)); }", 1, 41, 1, 51),
		keyCase("case.js", "if (map.has(key)) { (count++, map.get(anotherKey)); }", 1, 39, 1, 49),
		keyCase("case.js", "if (map.has(key)) { if (check(() => map.has(anotherKey))) { map.get(anotherKey); } }", 1, 69, 1, 79),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_mismatched_map_key.NoMismatchedMapKeyRule, valid, invalid)
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-mismatched-map-key.md
func TestNoMismatchedMapKeyDocumentation(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{FileName: "case.js", Code: "const value = map.has(key) ? map.get(key) : fallback;"},
		{FileName: "case.js", Code: `if (!map.has(key)) {
	map.set(key, value);
}`},
	}
	invalid := []rule_tester.InvalidTestCase{
		keyCase("case.js", "const value = map.has(key) ? map.get(anotherKey) : fallback;", 1, 38, 1, 48),
		keyCase("case.js", `if (!map.has(key)) {
	map.set(anotherKey, value);
}`, 2, 10, 2, 20),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_mismatched_map_key.NoMismatchedMapKeyRule, valid, invalid)
}

func keyCase(filename, code string, line, column, endLine, endColumn int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		FileName: filename,
		Code:     code,
		Errors:   []rule_tester.InvalidTestCaseError{keyError(line, column, endLine, endColumn)},
	}
}

func keyError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "no-mismatched-map-key",
		Message:   "Do not check a Map key before accessing a different key.",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}
