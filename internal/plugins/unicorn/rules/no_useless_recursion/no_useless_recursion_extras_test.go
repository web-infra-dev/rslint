package no_useless_recursion_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_recursion"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessRecursionExtras(t *testing.T) {
	script := rule.LanguageOptions{SourceType: "script"}
	namedExpression := recursiveReturn(`with ({ walk: () => 1 }) {
  const saved = function walk() {
    walk = () => 1;
    return walk();
  };
}`, "case.js", 4, 12, 16)
	namedExpression.LanguageOptions = script

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_recursion.NoUselessRecursionRule,
		[]rule_tester.ValidTestCase{
			// These returns do not have a plain call with an identifier callee.
			{Code: `async function walk(n) {
  if (n === 0) return;
  if (n === 1) return ((walk?.()));
  if (n === 2) return (walk?.())();
  if (n === 3) return new walk();
  if (n === 4) return walk["call"](null);
  return await walk();
}`, FileName: "case.js"},
			{Code: `export default function () { return walk(); }
const generator = async function* walk() { return walk(); };
const shadowed = function walk(walk) { return walk(); };`, FileName: "case.js"},
			{Code: `function walk() { { function walk() {} return walk(); } }`, FileName: "case.js"},
			{Code: `function walk() {
  walk = () => 1;
  return walk();
}`, FileName: "case.js"},
			// Calling a saved reference can run the original body after the binding changes.
			{Code: `export default function walk(): number {
  return walk();
}
const saved = walk;
({ replacement: walk } = handlers);
saved();`, FileName: "case.ts"},
			{
				Code: `function walk() { return walk(); }
const saved = walk;
var { walk } = handlers;
saved();`,
				FileName:        "case.js",
				LanguageOptions: script,
			},
			{
				Code:            `function walk() { with ({ walk: () => 1 }) { return walk(); } }`,
				FileName:        "case.js",
				LanguageOptions: script,
			},
			// The implicit arguments binding shadows the function name in script code.
			{
				Code:            `function arguments() { return arguments(); }`,
				FileName:        "case.js",
				LanguageOptions: script,
			},
			// Returns in methods, accessors, constructors, and arrows belong to those functions.
			{Code: `function walk() {
  const arrow = () => { return walk(); };
  class Example {
    constructor() { return walk(); }
    get next() { return walk(); }
    set next(value) { return walk(); }
    #walk() { return walk(); }
  }
}`, FileName: "case.js"},
			{Code: `function walk() {
  try { run(); }
  catch { return walk(); }
  finally { return walk(); }
}`, FileName: "case.js"},
			{Code: `function walk() {
  using resource = acquire();
  if (ready) {
    return walk();
  }
}
async function iterate(items) {
  for await (using resource of items) {
    return iterate(items);
  }
}`, FileName: "case.js"},
			// Authored TypeScript wrappers stay visible to the rule.
			{Code: `function walk<T>(x: number): unknown {
  if (x === 0) return walk(x) as unknown;
  if (x === 1) return (walk as (n: number) => unknown)(x);
  if (x === 2) return walk(x)!;
  if (x === 3) return (walk!)(x);
  if (x === 4) return (<typeof walk>walk)(x);
  if (x === 5) return (walk<T>)(x);
  return walk(x) satisfies unknown;
}`, FileName: "case.ts"},
			// ESLint's lexical lookup includes type-only bindings.
			{Code: `function walk() {
  type walk = string;
  return walk();
}
function visit<visit>(value: visit): unknown {
  return visit(value);
}`, FileName: "case.ts"},
		},
		[]rule_tester.InvalidTestCase{
			// The inner name is immutable and in scope before the surrounding with.
			namedExpression,
			recursiveReturn(`function walk() {
  { let walk; walk = () => 1; }
  walk.count++;
  return walk();
}`, "case.js", 4, 10, 14),
			recursiveReturn(`function walk() { return (((walk))()); }`, "case.js", 1, 29, 33),
			recursiveReturn(`function walk() { return /** @type {number} */ ((/** @type {() => number} */ (walk))()); }`, "case.js", 1, 79, 83),
			recursiveReturn(`function walk() { return /** @satisfies {number} */ ((/** @satisfies {() => number} */ (walk))()); }`, "case.js", 1, 89, 93),
			// The names resolve to the same binding; the reported columns count UTF-16 units.
			recursiveReturn(`function \u{10400}() { /* 😀 */ return 𐐀(); }`, "case.js", 1, 40, 42),
			recursiveReturn(`function 𐐀() { /* 😀 */ return \u{10400}(); }`, "case.js", 1, 33, 42),
			// The surrounding function owns the resource and the try statement.
			recursiveReturn(`function outer() {
  using resource = acquire();
  try {
    function walk() { return walk(); }
  } finally { release(); }
}`, "case.js", 4, 30, 34),
			recursiveReturn(`function walk() {
  { using resource = acquire(); }
  if (ready) return walk();
  using resource = acquire();
}`, "case.js", 3, 21, 25),
			recursiveReturn(`function walk(value) {
  switch (value) {
    case 0: {
      using resource = acquire();
      return walk(1);
    }
    default:
      return walk(value);
  }
}`, "case.js", 8, 14, 18),
			recursiveReturn(`function walk() {
  for (let i = 0; i < count; i++) {
    for (const using of items) {
      return walk();
    }
  }
}`, "case.js", 4, 14, 18),
			recursiveReturn(`export default function walk() { return walk(); }`, "case.js", 1, 41, 45),
			recursiveReturn(`export function run() { return run(); }`, "case.js", 1, 32, 35),
			recursiveReturn(`namespace Tree { export function walk() { return walk(); } }`, "case.ts", 1, 50, 54),
			recursiveReturn(`function walk<T>(value: T): T { return (walk)<T>(value); }`, "case.ts", 1, 41, 45),
			recursiveReturn(`function walk(value: number): number;
function walk(value: number): number { return walk(value); }`, "case.ts", 2, 47, 51),
			recursiveReturn(`function walk(value: unknown): unknown { return walk(<div />); }`, "case.tsx", 1, 49, 53),
		},
	)
}

func TestNoUselessRecursionBindingExtras(t *testing.T) {
	script := rule.LanguageOptions{SourceType: "script"}
	scriptRecursion := func(code string, line, column int) rule_tester.InvalidTestCase {
		test := recursiveReturn(code, "case.js", line, column, column+4)
		test.LanguageOptions = script
		return test
	}
	const globalWrite = `function walk() {
  globalThis.walk = () => 1;
  return walk();
}`
	const blockFunction = `function walk(n) {
  if (true) { function walk() { return 1; } }
  if (n) return walk(n - 1);
  return 0;
}`

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_recursion.NoUselessRecursionRule,
		[]rule_tester.ValidTestCase{
			{Code: `function outer(walk) {
  function walk(walk) { return walk(); }
}`, FileName: "case.js"},
			// Unlike a parameter-default write, this updates the body binding.
			{Code: `function outer(walk = () => 0) {
  function walk() { return walk(); }
  walk = () => 1;
}`, FileName: "case.js"},
			{Code: globalWrite, FileName: "case.js", LanguageOptions: script},
			{Code: `function walk() {
  (globalThis)["wa" + "lk"] = () => 1;
  return walk();
}`, FileName: "case.js", LanguageOptions: script},
			{Code: blockFunction, FileName: "case.js", LanguageOptions: script},
			{Code: `function walk(n) {
  if (n) return walk(n - 1);
  return 0;
}
const saved = walk;
{ function walk() { return 1; } }
saved(1);`, FileName: "case.js", LanguageOptions: script},
			{Code: `const saved = function walk(n) {
  { function walk() { return 1; } }
  if (n) return walk(n - 1);
  return 0;
};`, FileName: "case.js", LanguageOptions: script},
			{Code: `function walk(n) {
  try { throw 0; } catch (walk) { { function walk() { return 1; } } }
  if (n) return walk(n - 1);
  return 0;
}`, FileName: "case.js", LanguageOptions: script},
		},
		[]rule_tester.InvalidTestCase{
			recursiveReturn(`function outer(walk) {
  function walk() {
    return walk();
  }
}`, "case.js", 3, 12, 16),
			recursiveReturn(`function outer({ walk }: { walk: () => number } = { walk: () => 0 }, change = (walk = () => 1)) {
  function walk(): number {
    return walk();
  }
}`, "case.ts", 3, 12, 16),
			recursiveReturn(globalWrite, "case.js", 3, 10, 14),
			scriptRecursion(`function walk(globalThis) {
  globalThis.walk = () => 1;
  return walk();
}`, 3, 10),
			scriptRecursion(`{
  function walk() {
    globalThis.walk = () => 1;
    return walk();
  }
}`, 4, 12),
			recursiveReturn(blockFunction, "case.js", 3, 17, 21),
			scriptRecursion(`"use strict";`+"\n"+blockFunction, 4, 17),
			// Lexical declarations can prevent the block function from escaping.
			scriptRecursion(`function walk(n) {
  { let walk; { function walk() { return 1; } } }
  if (n) return walk(n - 1);
  return 0;
}`, 3, 17),
			scriptRecursion(`function walk(n) {
  try { throw { walk: 0 }; } catch ({ walk }) { { function walk() { return 1; } } }
  if (n) return walk(n - 1);
  return 0;
}`, 3, 17),
			// An assignment to the local copy does not change the global function.
			scriptRecursion(`function walk() {
  return walk();
}
function other() {
  { function walk() { return 1; } }
  walk = () => 2;
}`, 2, 10),
		},
	)
}
