package no_useless_recursion_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_recursion"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessRecursionExtras(t *testing.T) {
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
			// The implicit arguments binding shadows the function name in script code.
			{
				Code:            `function arguments() { return arguments(); }`,
				FileName:        "case.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
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
