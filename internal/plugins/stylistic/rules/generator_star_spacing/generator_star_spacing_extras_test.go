// TestGeneratorStarSpacingExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// Coverage summary:
//   - Dimension 4 walk for the function-header location (most rows are N/A
//     because the `*` is not a child expression the rule inspects — those
//     rows carry an explicit `// N/A: ...` marker)
//   - tsgo trivia detection: tabs, multiple spaces, CR/LF newlines, line +
//     block comments between `function`/`*` and `*`/name
//   - container variants: class declaration vs class expression, named vs
//     anonymous function expression, TS class with `declare`, async generator
//     in class / object literal
//   - same-kind nesting: generator inside generator reports both
//   - upstream-branch lock-ins: `shorthand ?? method` fallback, partial
//     object overrides inheriting from defaults, unknown-string fallback
//   - real-user shapes from the ES module ecosystem
package generator_star_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/generator_star_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestGeneratorStarSpacingExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&generator_star_spacing.GeneratorStarSpacingRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: trivia kinds — any non-zero distance between tokens counts as "spaced" ----

			// multiple spaces between `function` and `*` — distance > 0, treated
			// as spaced by upstream's `!!(rightToken.range[0] - leftToken.range[1])`
			{Code: "function   *foo(){}"},
			// tab between `function` and `*`
			{Code: "function\t*foo(){}"},
			// newline between `function` and `*` — Layer-2: tsgo line maps line/col
			// independently; we only care that the trivia counts as spacing
			{Code: "function\n*foo(){}"},
			// CRLF newline between
			{Code: "function\r\n*foo(){}"},
			// block comment between `function` and `*` — non-zero distance, valid
			// under default 'before'
			{Code: "function/*c*/ *foo(){}"},
			// line comment terminated by newline between `function` and `*`
			{Code: "function //c\n*foo(){}"},

			// ---- Dimension 4: method name forms ----

			// string-literal key in object shorthand
			{Code: "var foo = { *'bar'(){} };"},
			// numeric-literal key in object shorthand
			{Code: "var foo = { *0(){} };"},
			// computed key in object shorthand (already in upstream, but with brackets-direct)
			{Code: "var foo = { *[x](){} };"},
			// private identifier in class method (TS / ES2022)
			{Code: "class Foo { *#bar(){} }"},
			// string-literal key in class method
			{Code: "class Foo { *'bar'(){} }"},
			// numeric-literal key in class method
			{Code: "class Foo { *0(){} }"},

			// ---- Dimension 4: container forms ----

			// class expression
			{Code: "var X = class { *foo(){} };"},
			// named class expression
			{Code: "var X = class Y { *foo(){} };"},
			// anonymous function expression named via assignment
			{Code: "var foo = function *(){};"},
			// FE inside object literal as PropertyAssignment value (NOT shorthand)
			// — kind is `anonymous`, before-check applies (function keyword)
			{Code: "var foo = { bar: function *() {} };"},

			// ---- Dimension 4: async generators (modifier presence enables before-check) ----

			// async generator: function declaration form
			{Code: "async function *foo(){}"},
			// async generator: anonymous function expression
			{Code: "var foo = async function *(){};"},
			// async generator: class method without static
			{Code: "class Foo { async *foo(){} }"},
			// N/A: `*` cannot appear on an arrow function — arrow functions are
			// not generators in any ECMAScript version

			// ---- Dimension 4: nesting (same-kind boundaries) ----

			// outer generator with inner non-generator — only outer is a generator
			{Code: "function *outer() { function inner() {} }"},

			// ---- Dimension 4: graceful degradation ----

			// abstract generator method (TS-only) — body absent; before-check
			// still applies because `abstract` is a modifier
			{Code: "abstract class Foo { abstract *foo(): IterableIterator<number>; }"},
			// declared class with generator method overload (body-absent)
			{Code: "declare class Foo { *foo(): IterableIterator<number>; }"},
			// overload signature followed by implementation, both generators
			{Code: "class Foo { *foo(x: number): IterableIterator<number>; *foo(x: string): IterableIterator<string>; *foo(x: any) { yield x; } }"},

			// N/A: paren wrapper around the function — `(*foo()...)` is not valid
			//      syntax; `*` is part of the header not an expression child
			// N/A: optional chain on `*` — same reason
			// N/A: TS `as`/`satisfies`/non-null on `*` — same reason

			// ---- Branch lock-ins ----

			// Locks in upstream `shorthand ?? method` arm: when shorthand override
			// is omitted, falls back to the method override (not to top-level
			// before/after). Without the `??` fallback, this would error on the
			// missing space after `*`.
			{
				Code:    "var foo = { *foo() {} };",
				Options: optObj(map[string]any{"before": false, "after": false, "method": "neither"}),
			},

			// Locks in upstream `optionToDefinition` object arm: partial object
			// merges with defaults. Here `named.after` is omitted, so it must
			// inherit from top-level `after: false` (not from any built-in
			// default that would be `false` only by coincidence).
			{
				Code:    "function *foo(){}",
				Options: optObj(map[string]any{"before": true, "after": true, "named": map[string]any{"after": false}}),
			},

			// Locks in upstream `optionToDefinition` unknown-string arm: an
			// unknown string falls through to the defaults (top-level before:
			// true, after: false), so `function *foo(){}` stays valid.
			{
				Code:    "function *foo(){}",
				Options: optStr("nonsense"),
			},

			// Locks in upstream `defaultOptions` resolution: an empty object at
			// top level falls back to the rule's defaults (before: true,
			// after: false), so the default-before behavior holds.
			{
				Code:    "function *foo(){}",
				Options: optObj(map[string]any{}),
			},

			// ---- Real-user shapes (ESM exports — common in libraries) ----

			// named export of a generator declaration
			{Code: "export function *foo() { yield 1; }"},
			// default export of an anonymous generator
			{Code: "export default function *() { yield 1; }"},
			// default export of a named generator
			{Code: "export default function *foo() { yield 1; }"},
			// const-bound generator expression
			{Code: "export const foo = function *() { yield 1; };"},

			// ---- Real-user shapes (redux-saga style — generator as the dominant pattern) ----

			// saga: named generator with yield-call (the dominant pattern in the
			// redux-saga ecosystem; the rule must not flag the inner `yield call`).
			{Code: "function *saga() { yield call(api, payload); }"},
			// saga with `yield* delegate(...)` — yield delegation token is `*`
			// too, but it lives on YieldExpression (NOT on a function/method),
			// so the rule must not see it. Locks in that the listener set is
			// scoped correctly.
			{Code: "function *saga() { yield* delegate(); }"},
			// saga loop with multiple yield-delegations
			{Code: "function *root() { while (true) { yield* listen(); yield* poll(); } }"},

			// ---- Robustness: listener must NOT fire on non-monitored AST kinds ----

			// N/A: TS interface / type-literal method signatures cannot use `*`
			// at all (TS parse error: "Property or signature expected") because
			// generators require a body. So there is no MethodSignature-shaped
			// regression to lock in here — the rule's listener exclusion of
			// MethodSignature is correct *and* untestable through valid input.

			// Getter / setter on class — KindGetAccessor / KindSetAccessor,
			// not monitored by the rule, must not error
			{Code: "class Foo { get bar() { return 1; } set bar(v) {} }"},
			// Constructor — KindConstructor, not monitored
			{Code: "class Foo { constructor() {} }"},
			// async non-generator method — has no AsteriskToken, checkGenerator
			// returns early on nil
			{Code: "class Foo { async bar() {} }"},
			// regular non-generator method
			{Code: "class Foo { bar() {} }"},
			// Object property assignment with non-generator function expression
			// — checkGenerator returns early because AsteriskToken is nil
			{Code: "var foo = { bar: function() {} };"},

			// ---- Robustness: container nesting (the rule walks into every body) ----

			// generator IIFE — anonymous FunctionExpression as a CallExpression's
			// callee. Locks in that traversal reaches CallExpression children.
			{Code: "(function *() { yield 1; })()"},
			// generator as a CallExpression argument (common: setTimeout etc.)
			{Code: "setTimeout(function *() { yield 1; }, 0);"},
			// generator inside a default parameter value
			{Code: "function foo(cb = function *() { yield 1; }) {}"},
			// generator returned from a generator
			{Code: "function *outer() { return function *() { yield 1; }; }"},
			// 4-level nesting: class > method (gen) > inner gen > deepest gen
			{
				Code: "class A {\n" +
					"  *m() {\n" +
					"    var x = class B {\n" +
					"      *n() {\n" +
					"        return function *() { return function *() { yield 1; }; };\n" +
					"      }\n" +
					"    };\n" +
					"    yield 1;\n" +
					"  }\n" +
					"}",
			},

			// ---- Robustness: TS-specific shapes ----

			// TS generic generator — type parameters between `name` and `(`
			// don't affect the `*` spacing check
			{Code: "function *foo<T>(x: T): IterableIterator<T> { yield x; }"},
			// TS generic anonymous generator expression
			{Code: "var foo = function *<T>(x: T): IterableIterator<T> { yield x; };"},
			// TS generic generator class method
			{Code: "class Foo { *bar<T>(x: T): IterableIterator<T> { yield x; } }"},
			// TS parameter destructuring + rest — params don't touch `*`
			{Code: "function *foo({ a, b }, ...rest) { yield a; yield b; }"},
			// TS multi-modifier class method: public + static + async + generator
			{Code: "class Foo { public static async *bar() { yield 1; } }"},
			// TS protected modifier + generator
			{Code: "class Foo { protected *bar() { yield 1; } }"},
			// TS private modifier + generator
			{Code: "class Foo { private *bar() { yield 1; } }"},
			// TS override modifier + generator
			{Code: "class Bar extends Foo { override *bar() { yield 1; } }"},
			// private identifier method name (#priv) + static modifier
			{Code: "class Foo { static *#priv() { yield 1; } }"},
			// Decorator + generator method — decorators sit in tsgo's
			// `Modifiers()` list as `KindDecorator` entries (see
			// class_methods_use_this.go:576), so `hasAnyModifier` returns
			// true and the before-check engages. Without this behavior we
			// would mis-skip the before-check on decorated generators and
			// silently miss spacing violations on `@dec*foo()`.
			{Code: "class Foo { @dec *bar() { yield 1; } }"},

			// ---- Robustness: control-flow containment ----

			// generator declaration inside switch case — function declarations
			// are hoisted to the enclosing block; the listener must still fire
			{Code: "switch (x) { case 1: function *foo() { yield 1; } break; }"},
			// generator declaration inside if-block
			{Code: "if (x) { function *foo() { yield 1; } }"},
			// generator declaration inside for-loop body
			{Code: "for (var i = 0; i < 1; i++) { (function *() { yield i; })(); }"},

			// ---- Mixed class members: only the generator member is inspected ----

			// generator + getter + setter + normal in same class — only *gen
			// would have been inspected; here it satisfies default 'before'
			{Code: "class Mixed { *gen() { yield 1; } normal() {} get prop() { return 1; } set prop(v) {} }"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: comment trivia interacts with autofix ----

			// Block comment after `*` — under default 'before', the after-side
			// is `unexpected` since distance from `*` to `/` is 0... wait, the
			// block comment IS distance > 0 — actually `function*/*c*/foo`:
			// `*` at col 9, `/` at col 10, name `f` at col 15. After-distance =
			// 15 - 10 = 5 > 0 → spaced. With 'before' default (after: false),
			// this is `unexpected after`. Autofix removes the comment.
			{
				Code:    "function */*c*/foo(){}",
				Output:  []string{"function *foo(){}"},
				Errors:  []rule_tester.InvalidTestCaseError{errUA(10)},
			},

			// ---- Dimension 4: nesting — both generators reported independently ----

			// Outer + inner both violate default 'before'
			{
				Code:   "function* outer() { function* inner() { yield 1; } yield 2; }",
				Output: []string{"function *outer() { function *inner() { yield 1; } yield 2; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					errMB(9), errUA(9),
					{MessageId: "missingBefore", Line: 1, Column: 29, EndLine: 1, EndColumn: 30},
					{MessageId: "unexpectedAfter", Line: 1, Column: 29, EndLine: 1, EndColumn: 30},
				},
			},

			// Generator method nested inside a class expression assigned to a property
			// — class expression body is traversed; the inner generator reports.
			// No modifier on the method, so before-check is skipped (upstream's
			// `(method||shorthand) && star === firstToken(parent)` short-circuit).
			{
				Code:    "var x = { y: class { *foo(){} } };",
				Output:  []string{"var x = { y: class { * foo(){} } };"},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},

			// ---- Dimension 4: TS-modifier prefixes engage before-check on methods ----

			// `public` modifier in class — `*` is no longer the first token
			{
				Code:    "class Foo { public*foo(){} }",
				Output:  []string{"class Foo { public * foo(){} }"},
				Options: optBA(true, true),
				Errors:  []rule_tester.InvalidTestCaseError{errMB(19), errMA(19)},
			},

			// async generator object shorthand with 'neither' — both unexpected
			{
				Code:    "var foo = { async * bar(){} };",
				Output:  []string{"var foo = { async*bar(){} };"},
				Options: optBA(false, false),
				Errors:  []rule_tester.InvalidTestCaseError{errUB(19), errUA(19)},
			},

			// ---- Branch lock-in: shorthand fallback to method ----

			// Locks in upstream `shorthand ?? method`: shorthand undefined,
			// method=both → object shorthand requires both spaces. Without the
			// fallback, would report under top-level (false,false) instead.
			{
				Code:    "var foo = { *bar() {} };",
				Output:  []string{"var foo = { * bar() {} };"},
				Options: optObj(map[string]any{"before": false, "after": false, "method": "both"}),
				Errors:  []rule_tester.InvalidTestCaseError{errMA(13)},
			},

			// Locks in upstream override-isolation: anonymous override applies
			// to anonymous FE only; named FunctionDeclaration uses top-level.
			{
				Code:    "function * foo(){}",
				Output:  []string{"function*foo(){}"},
				Options: optObj(map[string]any{"before": false, "after": false, "anonymous": "both"}),
				Errors:  []rule_tester.InvalidTestCaseError{errUB(10), errUA(10)},
			},

			// Locks in upstream `Object.assign({}, defaults, option)`: partial
			// object override inherits unspecified key from top-level (here
			// `named.after` inherits `after: true`, so missing after on
			// `function *foo` reports).
			{
				Code:    "function *foo(){}",
				Output:  []string{"function * foo(){}"},
				Options: optObj(map[string]any{"before": true, "after": true, "named": map[string]any{"before": true}}),
				Errors:  []rule_tester.InvalidTestCaseError{errMA(10)},
			},

			// ---- Real-user shapes ----

			// Named export with generator — default options ('before')
			{
				Code:   "export function*foo(){ yield 1; }",
				Output: []string{"export function *foo(){ yield 1; }"},
				Errors: []rule_tester.InvalidTestCaseError{errMB(16)},
			},

			// Default export anonymous generator — default options ('before')
			{
				Code:   "export default function* (){ yield 1; }",
				Output: []string{"export default function *(){ yield 1; }"},
				Errors: []rule_tester.InvalidTestCaseError{errMB(24), errUA(24)},
			},

			// Locks in upstream `!node.id → anonymous` for the
			// FunctionDeclaration path: `export default function() {}` is a
			// FunctionDeclaration whose Name is nil. The anonymous override
			// must apply to it. Without this routing, top-level defaults
			// (false, false) would be used and no error would be reported.
			{
				Code:    "export default function*() { yield 1; }",
				Output:  []string{"export default function * () { yield 1; }"},
				Options: optObj(map[string]any{"before": false, "after": false, "anonymous": "both"}),
				Errors:  []rule_tester.InvalidTestCaseError{errMB(24), errMA(24)},
			},

			// Class with static async generator — common modern shape
			{
				Code:    "class Foo { static async*bar() { yield 1; } }",
				Output:  []string{"class Foo { static async * bar() { yield 1; } }"},
				Options: optBA(true, true),
				Errors:  []rule_tester.InvalidTestCaseError{errMB(25), errMA(25)},
			},

			// ---- Multi-line cases (Contract Alignment Checklist: ≥1 multi-line per container) ----

			// Multi-line function declaration with violation on line 1
			{
				Code: "function*foo() {\n  yield 1;\n}",
				Output: []string{"function *foo() {\n  yield 1;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingBefore", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},

			// Multi-line class method
			{
				Code: "class Foo {\n  static *bar() {\n    yield 1;\n  }\n}",
				Output: []string{"class Foo {\n  static* bar() {\n    yield 1;\n  }\n}"},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 2, Column: 10, EndLine: 2, EndColumn: 11},
					{MessageId: "missingAfter", Line: 2, Column: 10, EndLine: 2, EndColumn: 11},
				},
			},

			// Multi-line object shorthand
			{
				Code: "var x = {\n  *foo() {\n    yield 1;\n  },\n};",
				Output: []string{"var x = {\n  * foo() {\n    yield 1;\n  },\n};"},
				Options: optStr("after"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingAfter", Line: 2, Column: 3, EndLine: 2, EndColumn: 4},
				},
			},

			// ---- Real-user shape: redux-saga ----

			// Common saga style: named generator, must not let the yield-delegation
			// `*` (yield* call(...)) interfere with the function-`*` check
			{
				Code:   "function*saga() { yield call(api); yield* delegate(); }",
				Output: []string{"function *saga() { yield call(api); yield* delegate(); }"},
				Errors: []rule_tester.InvalidTestCaseError{errMB(9)},
			},

			// ---- Robustness: container nesting (rule must traverse into bodies) ----

			// IIFE generator violating 'neither' — anonymous, modifier-less
			// FunctionExpression as CallExpression callee
			{
				Code:    "(function * () { yield 1; })()",
				Output:  []string{"(function*() { yield 1; })()"},
				Options: optStr("neither"),
				Errors:  []rule_tester.InvalidTestCaseError{errUB(11), errUA(11)},
			},

			// Generator inside default parameter — anonymous FE, default 'before'
			// catches `function*` with no leading space
			{
				Code:   "function foo(cb = function*() { yield 1; }) {}",
				Output: []string{"function foo(cb = function *() { yield 1; }) {}"},
				Errors: []rule_tester.InvalidTestCaseError{errMB(27)},
			},

			// Generator as setTimeout argument — anonymous FE in CallExpression
			// arguments list. Default 'before', missing before.
			{
				Code:   "setTimeout(function*() { yield 1; }, 0);",
				Output: []string{"setTimeout(function *() { yield 1; }, 0);"},
				Errors: []rule_tester.InvalidTestCaseError{errMB(20)},
			},

			// 4-level nesting reports every violating generator independently.
			// Default 'before' / no option:
			//   - col 11 `*m()` and col 36 `*n()` are MethodDeclarations with
			//     no modifier → before-check skipped, after-check passes (no
			//     space before `(`), so they don't report.
			//   - col 58 `function* ()` anonymous FE: missingBefore (function
			//     immediately precedes `*`) AND unexpectedAfter (space after `*`)
			//   - col 80 `function*()` anonymous FE: missingBefore only
			{
				Code: "class A { *m() { var x = class B { *n() { return function* () { return function*() { yield 1; }; }; } }; } }",
				Output: []string{
					"class A { *m() { var x = class B { *n() { return function *() { return function *() { yield 1; }; }; } }; } }",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingBefore", Line: 1, Column: 58, EndLine: 1, EndColumn: 59},
					{MessageId: "unexpectedAfter", Line: 1, Column: 58, EndLine: 1, EndColumn: 59},
					{MessageId: "missingBefore", Line: 1, Column: 80, EndLine: 1, EndColumn: 81},
				},
			},

			// ---- Robustness: TS-specific modifier combinations engage before-check ----

			// Multi-modifier method: `public static async *foo` violating 'neither'
			// — three modifiers, then `*`. Before is spaced (after async), after
			// is spaced (before foo). Both unexpected.
			{
				Code:    "class Foo { public static async * foo() { yield 1; } }",
				Output:  []string{"class Foo { public static async*foo() { yield 1; } }"},
				Options: optStr("neither"),
				Errors:  []rule_tester.InvalidTestCaseError{errUB(33), errUA(33)},
			},

			// `override` modifier (TS) engages before-check
			{
				Code:    "class Bar extends Foo { override*bar() { yield 1; } }",
				Output:  []string{"class Bar extends Foo { override * bar() { yield 1; } }"},
				Options: optBA(true, true),
				Errors:  []rule_tester.InvalidTestCaseError{errMB(33), errMA(33)},
			},

			// Private identifier + static modifier: `static*#priv` violates
			// default 'before' — missing space between `static` and `*`
			{
				Code:   "class Foo { static*#priv() { yield 1; } }",
				Output: []string{"class Foo { static *#priv() { yield 1; } }"},
				Errors: []rule_tester.InvalidTestCaseError{errMB(19)},
			},

			// Decorator + missing space before `*` — decorator is part of
			// the Modifiers list (tsgo design), so before-check engages and
			// catches `@dec*bar()` violating default 'before'.
			{
				Code:   "class Foo { @dec*bar() { yield 1; } }",
				Output: []string{"class Foo { @dec *bar() { yield 1; } }"},
				Errors: []rule_tester.InvalidTestCaseError{errMB(17)},
			},

			// TS generic generator: type parameters don't affect `*` check
			{
				Code:   "function*foo<T>(x: T): IterableIterator<T> { yield x; }",
				Output: []string{"function *foo<T>(x: T): IterableIterator<T> { yield x; }"},
				Errors: []rule_tester.InvalidTestCaseError{errMB(9)},
			},

			// ---- Robustness: control-flow containment ----

			// generator declaration inside switch case — listener must fire
			{
				Code:   "switch (x) { case 1: function*foo() { yield 1; } break; }",
				Output: []string{"switch (x) { case 1: function *foo() { yield 1; } break; }"},
				Errors: []rule_tester.InvalidTestCaseError{errMB(30)},
			},

			// generator declaration inside if-block
			{
				Code:   "if (x) { function*foo() { yield 1; } }",
				Output: []string{"if (x) { function *foo() { yield 1; } }"},
				Errors: []rule_tester.InvalidTestCaseError{errMB(18)},
			},

			// ---- Robustness: mixed class members — only the generator member reports ----

			// Class with generator + regular + getter + setter. With default
			// 'before', only `*gen` and `static*g2` are checked; the others
			// are different node kinds (regular MethodDeclaration with no
			// AsteriskToken, GetAccessor, SetAccessor) and must not produce
			// false positives. `*gen` passes (no modifier → before skipped);
			// `static*g2` violates (modifier present, no space before `*`).
			{
				Code:   "class Mixed { *gen() { yield 1; } normal() {} get prop() { return 1; } set prop(v) {} static*g2() {} }",
				Output: []string{"class Mixed { *gen() { yield 1; } normal() {} get prop() { return 1; } set prop(v) {} static *g2() {} }"},
				Errors: []rule_tester.InvalidTestCaseError{errMB(93)},
			},
		},
	)
}
