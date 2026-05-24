// TestCommaSpacingExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
package comma_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/comma_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCommaSpacingExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&comma_spacing.CommaSpacingRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: empty containers (graceful degradation) ----
			{Code: `var a = [];`},
			{Code: `var a = {};`},
			{Code: `fn();`},
			{Code: `new C();`},
			{Code: `function fn() {}`},
			{Code: `class C<T> {}`},

			// ---- Dimension 4: TS tuple types ----
			{Code: `type T = [number, string];`},
			{Code: `type T = [number, string, boolean];`},
			{Code: `type T = readonly [number, string];`},
			{Code: `let t: [number, string] = [1, 'a'];`},

			// ---- Dimension 4: TS type-argument lists ----
			{Code: `let m: Map<string, number>;`},
			{Code: `function f<T>(x: Array<T>): Array<T, number> { return x; }`},
			{Code: `type P = Pick<T, 'a' | 'b'>;`},

			// ---- Dimension 4: TS enum members ----
			{Code: `enum E { A, B, C }`},
			{Code: `enum E { A = 1, B = 2 }`},
			{Code: `const enum E { A, B }`},

			// ---- Dimension 4: import / export specifiers ----
			{Code: `import { a, b } from 'mod';`},
			{Code: `import { a, b, } from 'mod';`},
			{Code: `import { a, b as c } from 'mod';`},
			{Code: `export { a, b };`},
			{Code: `export { a, b, };`},

			// ---- Dimension 4: spread / rest with commas ----
			{Code: `function fn(a, b, ...rest) {}`},
			{Code: `var [a, b, ...rest] = arr;`},
			{Code: `var [, , c] = arr;`},
			{Code: `var obj = { ...a, b: 1 };`},
			{Code: `var { a, ...rest } = obj;`},
			{Code: `fn(a, ...args);`},

			// ---- Dimension 4: nested arrays / objects / mixed ----
			{Code: `var a = [[1, 2], [3, 4]];`},
			{Code: `var a = [{x: 1, y: 2}, {x: 3, y: 4}];`},
			{Code: `fn({a: 1, b: 2}, [3, 4]);`},
			{Code: `var deep = {a: [{b: {c: [1, 2, 3]}}]};`},

			// ---- Dimension 4: declaration / container forms ----
			{Code: `class C { m(a, b) {} }`},
			{Code: `class C { static m(a, b) {} }`},
			{Code: `class C { get p() { return 1; } set p(v) {} }`},
			{Code: `var o = { m(a, b) {} };`},
			{Code: `var o = { get p() { return 1; }, set p(v) {} };`},
			{Code: `async function f(a, b) {}`},
			{Code: `function* g(a, b) {}`},
			{Code: `async function* ag(a, b) {}`},
			{Code: `var f = async (a, b) => {};`},
			{Code: `var f = async (a, b) => a + b;`},

			// ---- Dimension 4: TS heritage / generics interplay ----
			{Code: `class C<T extends U, V extends W> {}`},
			{Code: `interface I<T, U> extends A<T>, B<U> {}`},

			// ---- Dimension 4: BinaryExpression with CommaToken (sequence) ----
			// Locks in upstream: validateCommaSpacing sequence-expression path.
			{Code: `var x = (a, b, c);`},
			{Code: `for (var i = 0, j = 0; i < 10; i++, j++) {}`},

			// ---- Dimension 4: for-statement initializer commas ----
			{Code: `for (let i = 0, j = 10; i < j; i++) {}`},

			// ---- Dimension 4: receiver-wrapper interactions ----
			// tsgo preserves ParenthesizedExpression where ESTree flattens.
			// Locks in: prev-token end position correctly anchored to `)`.
			{Code: `fn((a), b);`},
			{Code: `fn((a + b), c, (d));`},
			{Code: `var x = [(1), (2), (3)];`},

			// ---- Dimension 4: TS non-null + type assertion + satisfies wrappers ----
			{Code: `fn(a!, b);`},
			{Code: `fn(a as number, b);`},
			{Code: `fn(a satisfies number, b);`},
			{Code: `var arr = [a!, b!];`},

			// ---- Dimension 4: optional chain receivers ----
			{Code: `fn(a?.b, c?.d);`},
			{Code: `fn(a?.(), b);`},

			// ---- Dimension 4: computed property keys ----
			{Code: `var o = {[k]: 1, [k2]: 2};`},
			{Code: `var o = {[k++]: 1, [k++]: 2};`},

			// ---- Real-user: literal string with comma must not be flagged ----
			// Strings are single tokens at scanner level — defensive cover.
			{Code: `var s = "a,b,c";`},
			{Code: `var s = 'a,b';`},
			{Code: "var s = `a,b`;"},
			{Code: "var s = `a${1},b`;"},

			// ---- Real-user: regex literal containing commas ----
			// Locks in collectExcludeRanges' KindRegularExpressionLiteral arm.
			{Code: `var r = /a,b,c/;`},
			{Code: `var r = /,/g;`},
			{Code: `var arr = [/,/, /,,/];`},

			// ---- Real-user: template literal between substitutions with commas ----
			// Locks in collectExcludeRanges' KindTemplateMiddle/Tail arms.
			{Code: "var s = `${a}, ${b}`;"},
			{Code: "var s = `a${b}, ${c}d`;"},
			{Code: "var s = `${a}, ${b}, ${c}`;"},

			// ---- Real-user: JSX text with commas ----
			{Code: `<a>hello, world</a>`, Tsx: true},
			{Code: `<a>{x}, {y}</a>`, Tsx: true},
			{Code: `<a attr="a,b,c" />`, Tsx: true},

			// ---- Real-user: line comment immediately after comma (before-side) ----
			// Locks in the `nextToken === Line && !spaceAfter` exemption.
			{Code: "var a = 1,// comment\nb = 2;", Options: optsNeither()},
			{Code: "fn(a,// c\n b);", Options: optsNeither()},

			// ---- Real-user: comma at very end of file (no next token) ----
			{Code: `var x,`},

			// ---- Locks in upstream validateCommaSpacing arm: prev=null when prev is comma ----
			// In `[a,,b]` the middle comma's prev token (in our token stream)
			// is `,1`, which collapses to nil. No before-report should fire on
			// `,2` even though there's no space before it.
			{Code: `var arr = [a,,b];`},
			{Code: `var arr = [a, ,b];`},
			{Code: `var arr = [a,, b];`},

			// ---- Locks in upstream validateCommaSpacing arm: next=null when next is comma ----
			{Code: `var arr = [a,,b];`},

			// ---- Locks in same-line short-circuit on before-side ----
			// `,` on its own line. prev (`1`) is on prior line ⇒ before-check skipped.
			{Code: "var arr = [\n  1\n  , 2\n];"},
			// Same shape, but with before:false, after:false — only the
			// after-check should run (it does match: comma is followed by
			// newline so after-check also skips since `2` is on next line).
			{Code: "var arr = [\n  1\n  ,\n  2\n];", Options: optsNeither()},

			// ---- Locks in same-line short-circuit on after-side ----
			// `2` is on the next line, so the after-check is skipped even
			// when there's no space after the comma.
			{Code: "var arr = [1,\n  2];"},
			// optsBoth requires space before too; we provide it. The after side
			// is short-circuited by the line break.
			{Code: "var arr = [1 ,\n  2];", Options: optsBoth()},

			// ---- Locks in close-paren exemption ----
			// The trailing comma's `next` token is `)`, which suppresses the
			// after-side check even when `, )` has a space — and conversely
			// `,)` with no space is fine under `optsNeither`.
			{Code: `fn(a,);`},
			{Code: `fn(a,b,);`, Options: optsNeither()},

			// ---- Locks in close-brace exemption ----
			{Code: `var o = {a: 1,};`, Options: optsNeither()},
			{Code: `var { a, } = x;`, Options: optsNeither()},

			// ---- Locks in close-bracket exemption ----
			{Code: `var a = [1,];`, Options: optsNeither()},

			// ---- Multiple comma-sites in one file ----
			{Code: `var x = [1, 2]; var y = {a: 1, b: 2}; fn(x, y);`},
			// optsBefore: space before comma, no space after.
			{Code: `var x = [1 ,2]; var y = {a: 1 ,b: 2}; fn(x ,y);`, Options: optsBefore()},

			// ---- Options shape: array-wrapped (exercises GetOptionsMap array branch) ----
			{Code: `var a = 1 ,b = 2;`, Options: []interface{}{map[string]interface{}{"before": true, "after": false}}},
			{Code: `var a = 1, b = 2;`, Options: []interface{}{map[string]interface{}{"before": false, "after": true}}},

			// ---- Options shape: partial — only `before` specified, `after` defaults to true ----
			{Code: `var a = 1, b = 2;`, Options: map[string]interface{}{"before": false}},

			// ---- TS-specific Dimension 4: trailing comma in TSX arrow generics ----
			// Upstream covers `<T,>(foo) =>` once; broaden to multi-param +
			// constrained variants.
			{Code: `const f = <T,>(x: T) => x;`, Tsx: true},
			{Code: `const f = <T, U,>(x: T, y: U) => x;`, Tsx: true},
			{Code: `const f = <T extends number,>(x: T) => x;`, Tsx: true},

			// ---- Real-user: JSX fragment with text comma ----
			// JsxFragment <>...</> children include JsxText same as JsxElement;
			// the comma inside the text content must NOT fire.
			{Code: `var x = <>,</>;`, Tsx: true},
			{Code: `var x = <>a, b, c</>;`, Tsx: true},

			// ---- Real-user: TS template literal type with literal-text commas ----
			// Template literal types reuse KindTemplateHead/Middle/Tail.
			// Comma INSIDE the literal portion must NOT fire; commas in
			// substituted type unions don't exist (unions use `|`).
			{Code: "type T = `Hello, ${string}`;"},
			{Code: "type T = `${A}, ${B}`;"},
			{Code: "type T = `a, b, c`;"},

			// ---- Real-user: shebang at start of file ----
			// scanShebangTrivia runs before any tokens; our scanner pass
			// starting at byte 0 must not see a leading `#!` as a stray token
			// stream that throws off downstream comma detection.
			{Code: "#!/usr/bin/env node\nvar a = 1, b = 2;"},

			// ---- Real-user: decorator with comma-separated arguments ----
			{Code: `@deco(a, b) class C {}`},
			{Code: `@deco(a, b, c) function fn() {}`},

			// ---- Real-user: TS conditional type with infer + tuple ----
			// The `, ` between `infer A` and `infer B` inside the tuple must
			// be checked normally; the conditional's `?:` shouldn't interfere.
			{Code: `type Head<T> = T extends [infer A, infer B] ? A : never;`},
			{Code: `type Tail<T> = T extends [any, ...infer R] ? R : never;`},

			// ---- Real-user: import type / export type with named specifiers ----
			{Code: `import type { A, B } from 'mod';`},
			{Code: `export type { A, B } from 'mod';`},
			{Code: `import { type A, type B } from 'mod';`},

			// ---- Real-user: class fields don't use comma separators, but
			// initializer expressions can contain comma-separated lists ----
			{Code: `class C { a = [1, 2]; b = { x: 1, y: 2 }; }`},

			// ---- Real-user: tagged template + sequence expression ----
			{Code: "var r = tag`hello`, x = 1;"},

			// ---- Real-user: destructuring with defaults ----
			{Code: `var { a = 1, b = 2 } = obj;`},
			{Code: `var [a = 1, b = 2] = arr;`},
			{Code: `function f({ a = 1, b = 2 } = {}, [c, d] = []) {}`},

			// ---- Real-user: parameter properties (TS constructor) ----
			{Code: `class C { constructor(public a: number, private b: string) {} }`},
			{Code: `class C { constructor(readonly a: number, protected b: string) {} }`},

			// ---- Dimension 4: ParenthesizedExpression deep nesting ----
			// tsgo preserves every paren level (ESTree flattens). The prev/next
			// token-end positions stay anchored to the LAST `)` or FIRST `(`,
			// independent of nesting depth.
			{Code: `fn(((a)), (((b))));`},
			{Code: `fn(((((a))))), (b);`},
			{Code: `var arr = [((a)), ((b))];`},
			{Code: `var x = (((a, b, c)));`},

			// ---- Dimension 4: Optional chain across receiver + access + call ----
			// In tsgo optional chains are flag-bearing (no ChainExpression
			// wrapper). The comma's prev token is the last child's end.
			{Code: `fn(a?.b, c?.d);`},
			{Code: `fn(a?.b?.c, d?.e?.f);`},
			{Code: `fn(a?.(), b?.(c, d));`},
			{Code: `fn(a?.[0], b?.[1]);`},
			{Code: `fn(a?.b?.c?.[0]?.(d, e), f);`},

			// ---- Dimension 4: TS expression wrapper combinations ----
			{Code: `fn(a!, b?.c, d as T);`},
			{Code: `fn((a as T)!, b satisfies U);`},
			{Code: `fn(a as T | U, b as V & W);`},
			{Code: `fn(<T>a, <U>b);`}, // legacy type assertion (non-JSX)

			// ---- Dimension 4: TS type-argument list combinations ----
			{Code: `var m: Map<string, number>;`},
			{Code: `var m: Map<Array<string>, Set<number>>;`},
			{Code: `var m: Map<keyof T, T[keyof T]>;`},
			{Code: `type R = Record<'a' | 'b', number>;`},
			{Code: `var x = fn<A, B, C>();`},
			{Code: `var x = new Cls<A, B>(p, q);`},

			// ---- Dimension 4: tuple type variants ----
			{Code: `type T = [];`},
			{Code: `type T = [number];`},
			{Code: `type T = [number, string];`},
			{Code: `type T = [number, string?];`},               // optional element
			{Code: `type T = [number, ...string[]];`},           // rest element
			{Code: `type T = [a: number, b: string];`},          // labeled tuple
			{Code: `type T = [a: number, b?: string];`},
			{Code: `type T = [a: number, ...rest: string[]];`},
			{Code: `type T = readonly [A, B];`},
			{Code: `type T = [[A, B], [C, D]];`},                // nested
			{Code: `type T = [{a: A, b: B}, [C, D]];`},          // mixed

			// ---- Dimension 4: union / intersection inside list element ----
			{Code: `function f(x: A | B, y: C & D) {}`},
			{Code: `var arr: (A | B)[] = [a, b];`},

			// ---- Dimension 4: complex destructuring ----
			{Code: `var {a: [b, c], d: {e, f}} = obj;`},
			{Code: `var [{a, b}, {c, d}] = arr;`},
			{Code: `var [[a, b], [c, d]] = matrix;`},
			{Code: `var { a: [b, , c], ...rest } = obj;`},
			{Code: `function fn({ a, b: { c, d } } = {}, [e, ...f] = []) {}`},

			// ---- Dimension 4: arrow function variants ----
			{Code: `var f = (a, b) => a + b;`},
			{Code: `var f = async (a, b) => a + b;`},
			{Code: `var f = (a, b): number => a + b;`},
			{Code: `var f = async (a, b): Promise<number> => a + b;`},
			{Code: `var f = <T>(a: T, b: T): T => a;`},
			{Code: `var f = async <T>(a: T, b: T): Promise<T> => a;`},

			// ---- Dimension 4: generator / async generator ----
			{Code: `function* g(a, b) { yield a; yield b; }`},
			{Code: `async function* ag(a, b) { yield a; yield b; }`},
			{Code: `class C { *m(a, b) {} async *m2(a, b) {} }`},

			// ---- Dimension 4: class member variants ----
			{Code: `class C { static a = 1; static b = 2; }`},
			{Code: `class C { readonly a: number = 1; readonly b: number = 2; }`},
			{Code: `class C { #priv = 1; #other = 2; m() { return [this.#priv, this.#other]; } }`},
			{Code: `abstract class C { abstract m(a: A, b: B): void; }`},

			// ---- Dimension 4: TS overload signatures ----
			{Code: `function f(a: A): void; function f(a: A, b: B): void; function f(a: A, b?: B): void {}`},
			{Code: `interface I { f(a: A): void; f(a: A, b: B): void; }`},

			// ---- Dimension 4: enum variants ----
			{Code: `enum E { A = 1, B = 2, C = 3 }`},
			{Code: `enum E { A = 'a', B = 'b' }`},
			{Code: `enum E { A = 1, B = 'b' }`}, // heterogeneous
			{Code: `enum E { 'a-b' = 1, 'c-d' = 2 }`},

			// ---- Dimension 4: namespace / module ----
			{Code: `namespace N { export const a = 1, b = 2; }`},
			{Code: `declare module 'm' { export const a: number, b: string; }`},

			// ---- Dimension 4: heritage clauses ----
			{Code: `class C extends Base implements A, B, C {}`},
			{Code: `interface I extends A, B, C {}`},
			{Code: `class C<T, U> extends Base<T, U> implements I<T>, J<U> {}`},

			// ---- Dimension 4: JSX combinations ----
			{Code: `<C a={[1, 2]} b={{x: 1, y: 2}} c={fn(a, b)} />`, Tsx: true},
			{Code: `<C {...props} a={1} />`, Tsx: true},
			{Code: `<C><D a={1} b={2} /><E /></C>`, Tsx: true},
			{Code: `<C>{[a, b, c].map(x => x)}</C>`, Tsx: true},

			// ---- Dimension 4: nested templates with substitutions and commas ----
			{Code: "var s = `${[1, 2]}, ${[3, 4]}`;"},
			{Code: "var s = `outer ${ `inner ${a, b}` }, ${c}`;"}, // nested template
			{Code: "var s = tag`${a}, ${b}`;"},                    // tagged template

			// ---- Dimension 4: regex variants ----
			{Code: `var r = /[,]/g;`},                  // regex char class with comma
			{Code: `var r = /,{2,4}/;`},                // regex quantifier braces (also contains comma!)
			{Code: `fn(/foo/g, /bar/i, /baz/m);`},     // multiple regex args
			{Code: `var arr = [/,/, /;/, /[,]/g];`},   // regex in array

			// ---- Dimension 4: for-loop variants ----
			{Code: `for (let i = 0, j = 10; i < j; i++, j--) {}`},
			{Code: `for (var [a, b] of arr) {}`},
			{Code: `for (var [a, , b] of arr) {}`},
			{Code: `for (var {a, b} in obj) {}`},

			// ---- Dimension 4: try-catch destructuring ----
			{Code: `try {} catch ({code, message}) {}`},
			{Code: `try {} catch (e: any) {}`}, // TS catch type

			// ---- Real-user: multi-line + multi-byte identifiers ----
			{Code: "var café = 1, naïve = 2, fiancé = 3;"},
			{Code: "fn(δ, λ, μ);"},
			{Code: "var emoji = '🌟', star = '⭐';"},

			// ---- Real-user: typeof / keyof / infer in type expressions ----
			{Code: `type T = typeof a; var x: T, y: T;`},
			{Code: `type T<U> = U extends [infer A, infer B, infer C] ? [A, B, C] : never;`},
			{Code: `type K = keyof T; var x: Record<K, number>;`},

			// ---- Real-user: const assertion ----
			{Code: `var arr = [1, 2, 3] as const;`},
			{Code: `var obj = { a: 1, b: 2 } as const;`},
			{Code: `fn([1, 2] as const, {a: 1} as const);`},

			// ---- Real-user: switch case lists (no commas, but defensive) ----
			{Code: `switch (x) { case 1: case 2: case 3: break; }`},

			// ---- Real-user: complex import patterns ----
			{Code: `import a, { b, c } from 'mod';`},                 // default + named
			{Code: `import a, * as ns from 'mod';`},                  // default + namespace
			{Code: `import { a as a1, b as b1, c } from 'mod';`},     // aliasing
			{Code: `export { a as a1, b as b1, c };`},
			{Code: `export { a, b } from 'mod';`},
			{Code: `export * from 'mod';`},                           // no commas
			{Code: `export * as ns from 'mod';`},                     // no commas

			// ---- Real-user: TS `import type` mixed forms ----
			{Code: `import { type A, type B, c } from 'mod';`},

			// ---- Real-user: function with `this` parameter ----
			{Code: `function fn(this: T, a: A, b: B) {}`},
			{Code: `function fn(this: void, a: A, b: B): void {}`},

			// ---- Real-user: complex default initializers ----
			{Code: `function fn(a = (1, 2), b = [3, 4], c = { x: 5 }) {}`},
			{Code: `function fn(a = fn2(x, y), b = obj.m(z, w)) {}`},

			// ---- Real-user: type predicates & asserts ----
			{Code: `function isFoo(x: any): x is Foo { return true; }`},
			{Code: `function assertFoo(x: any): asserts x is Foo {}`},
			{Code: `function fn(x: T, y: U): x is T & U { return true; }`},

			// ---- Real-user: yield + sequence ----
			{Code: `function* g() { yield (a, b, c); }`},
			{Code: `function* g() { var r = yield a; r = yield b; }`},

			// ---- Real-user: bigint / numeric separator literal in list ----
			{Code: `var arr = [1n, 2n, 3n];`},
			{Code: `var arr = [1_000_000, 2_000_000];`},
			{Code: `var arr = [0xff, 0o77, 0b11];`},

			// ---- Real-user: deep mixed nesting (worst-case integration) ----
			{Code: `var deep = { a: [{ b: fn(c, [d, e]), f: g?.h?.(i, j) }, [k, ...l]], m: <C a={[1, 2]}><D /></C> };`, Tsx: true},

			// ---- Real-user: comment between operand and comma (multi-shape) ----
			{Code: `fn(a /* x */ /* y */, b);`},
			{Code: `fn(a /*1*/, /*2*/ b, /*3*/ c);`},
			// `,//` is only valid under {after:false}; with default
			// (after:true) the comma needs a space before `//`.
			{Code: "fn(\n  a,// trailing\n  b,/* trailing */\n  c\n);", Options: optsNeither()},

			// ---- Real-user: trailing comma in TSX arrow generic ----
			// TSX-arrow trailing comma `<T,>` is the only place a trailing
			// comma in a type-arg-shaped list is currently legal in TS; it's
			// a type-parameter DECLARATION (not arguments). collectIgnored-
			// Commas adds the trailing comma to ignored so spacing doesn't
			// fire on it.
			{Code: `const a = <T,>(x: T) => x;`, Tsx: true},
			{Code: `const a = <T, U,>(x: T, y: U) => x;`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Locks in: comma after `(` exemption — the prev=`(` case ----
			// `fn( ,a)` is a syntax error so we use a normal call with the
			// space *between* args.
			{
				Code:    `fn(a , b);`,
				Output:  []string{`fn(a, b);`},
				Options: optsAfter(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},

			// ---- Locks in: sequence-expression branch ----
			{
				Code:    `(a , b);`,
				Output:  []string{`(a, b);`},
				Options: optsAfter(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 4},
				},
			},

			// ---- Locks in: for-statement update list with comma ----
			{
				Code:    `for (var i = 0,j = 0; i < 10; i++,j++) {}`,
				Output:  []string{`for (var i = 0, j = 0; i < 10; i++, j++) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 15},
					{MessageId: "missing", Line: 1, Column: 34},
				},
			},

			// ---- Locks in: import-specifier comma fires ----
			{
				Code:    `import {a ,b} from 'm';`,
				Output:  []string{`import {a, b} from 'm';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
					{MessageId: "missing", Line: 1, Column: 11},
				},
			},

			// ---- Locks in: export-specifier comma fires ----
			{
				Code:    `export {a ,b};`,
				Output:  []string{`export {a, b};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
					{MessageId: "missing", Line: 1, Column: 11},
				},
			},

			// ---- Locks in: enum members ----
			{
				Code:   `enum E { A ,B }`,
				Output: []string{`enum E { A, B }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
					{MessageId: "missing", Line: 1, Column: 12},
				},
			},

			// ---- Locks in: tuple type ----
			{
				Code:   `type T = [number ,string];`,
				Output: []string{`type T = [number, string];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
					{MessageId: "missing", Line: 1, Column: 18},
				},
			},

			// ---- Locks in: type-argument list ----
			{
				Code:   `let m: Map<string ,number>;`,
				Output: []string{`let m: Map<string, number>;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
					{MessageId: "missing", Line: 1, Column: 19},
				},
			},

			// ---- Locks in: rest element comma fires correctly ----
			{
				Code:   `function fn(a ,...rest) {}`,
				Output: []string{`function fn(a, ...rest) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
					{MessageId: "missing", Line: 1, Column: 15},
				},
			},

			// ---- Locks in: spread element comma fires correctly ----
			{
				Code:   `fn(...a ,b);`,
				Output: []string{`fn(...a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
					{MessageId: "missing", Line: 1, Column: 9},
				},
			},

			// ---- Locks in: paren-wrapped receiver ----
			// tsgo preserves ParenthesizedExpression — make sure the
			// prev-token range still falls on `)` and not inside the parens.
			{
				Code:   `fn((a) ,b);`,
				Output: []string{`fn((a), b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
					{MessageId: "missing", Line: 1, Column: 8},
				},
			},

			// ---- Locks in: non-null assertion wrapper as comma neighbor ----
			{
				Code:   `fn(a! ,b);`,
				Output: []string{`fn(a!, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 7},
					{MessageId: "missing", Line: 1, Column: 7},
				},
			},

			// ---- Locks in: `as` type assertion wrapper as comma neighbor ----
			{
				Code:   `fn(a as number ,b);`,
				Output: []string{`fn(a as number, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
					{MessageId: "missing", Line: 1, Column: 16},
				},
			},

			// ---- Locks in: optional chain receiver as comma neighbor ----
			{
				Code:   `fn(a?.b ,c);`,
				Output: []string{`fn(a?.b, c);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
					{MessageId: "missing", Line: 1, Column: 9},
				},
			},

			// ---- Locks in: nested array — only inner comma triggers ----
			{
				Code:   `var a = [[1 ,2], [3, 4]];`,
				Output: []string{`var a = [[1, 2], [3, 4]];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
					{MessageId: "missing", Line: 1, Column: 13},
				},
			},

			// ---- Locks in: class method params ----
			{
				Code:   `class C { m(a ,b) {} }`,
				Output: []string{`class C { m(a, b) {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
					{MessageId: "missing", Line: 1, Column: 15},
				},
			},

			// ---- Locks in: shorthand object property ----
			{
				Code:   `var o = {a ,b};`,
				Output: []string{`var o = {a, b};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
					{MessageId: "missing", Line: 1, Column: 12},
				},
			},

			// ---- Locks in: array-wrapped option still routes through GetOptionsMap ----
			{
				Code:    `var a = 1, b = 2;`,
				Output:  []string{`var a = 1 ,b = 2;`},
				Options: []interface{}{map[string]interface{}{"before": true, "after": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 10},
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},

			// ---- Locks in: multi-byte (Unicode) identifier — column is
			// per-character, not per-byte. `é` occupies 1 UTF-16 unit
			// (2 UTF-8 bytes), so the comma after `café = 1 ` sits at
			// column 14 the same as for `var cafe = 1 ,b`. Pins the
			// behavior so a future column-counting change is caught.
			{
				Code:   `var café = 1 ,b = 2;`,
				Output: []string{`var café = 1, b = 2;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
					{MessageId: "missing", Line: 1, Column: 14},
				},
			},

			// ---- Locks in: comma in JSX expression child fires for its operand ----
			{
				Code:   `<a>{fn(b ,c)}</a>`,
				Output: []string{`<a>{fn(b, c)}</a>`},
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
					{MessageId: "missing", Line: 1, Column: 10},
				},
			},

			// ---- Locks in: JSX text content commas don't trigger; only commas in
			// embedded expressions do (this case has both — only one should fire) ----
			{
				Code:   `<a>foo, bar {fn(1 ,2)}</a>`,
				Output: []string{`<a>foo, bar {fn(1, 2)}</a>`},
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
					{MessageId: "missing", Line: 1, Column: 19},
				},
			},

			// ---- Locks in: regex literal containing commas does NOT fire,
			// but a real comma immediately after the regex DOES ----
			{
				Code:   `fn(/,/ ,b);`,
				Output: []string{`fn(/,/, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
					{MessageId: "missing", Line: 1, Column: 8},
				},
			},

			// ---- Locks in: template literal text commas don't fire, but
			// commas in `${...}` substitutions do ----
			{
				Code:   "var s = `${a ,b}`;",
				Output: []string{"var s = `${a, b}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
					{MessageId: "missing", Line: 1, Column: 14},
				},
			},

			// ---- Locks in: for-statement update list fires on real comma ----
			{
				Code:   `for (var i = 0, j = 0; i < 10; i++ ,j++) {}`,
				Output: []string{`for (var i = 0, j = 0; i < 10; i++, j++) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36},
					{MessageId: "missing", Line: 1, Column: 36},
				},
			},

			// ---- Locks in: decorator argument list fires ----
			{
				Code:   `@deco(a ,b) class C {}`,
				Output: []string{`@deco(a, b) class C {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
					{MessageId: "missing", Line: 1, Column: 9},
				},
			},

			// ---- Locks in: TS infer-tuple comma fires ----
			{
				Code:   `type Head<T> = T extends [infer A ,infer B] ? A : never;`,
				Output: []string{`type Head<T> = T extends [infer A, infer B] ? A : never;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
					{MessageId: "missing", Line: 1, Column: 35},
				},
			},

			// ---- Locks in: JSX fragment text content doesn't fire, but
			// commas inside embedded expressions do ----
			{
				Code:   `var x = <>{fn(1 ,2)}</>;`,
				Output: []string{`var x = <>{fn(1, 2)}</>;`},
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
					{MessageId: "missing", Line: 1, Column: 17},
				},
			},

			// ---- Locks in: parameter property modifiers ----
			{
				Code:   `class C { constructor(public a: number ,private b: string) {} }`,
				Output: []string{`class C { constructor(public a: number, private b: string) {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 40},
					{MessageId: "missing", Line: 1, Column: 40},
				},
			},

			// ---- Locks in: shebang doesn't shift positions ----
			// The comma is on line 2 (after the shebang newline). If
			// shebang handling broke positions or kind detection, this
			// would mis-fire or mis-report.
			{
				Code:   "#!/usr/bin/env node\nvar a = 1 ,b = 2;",
				Output: []string{"#!/usr/bin/env node\nvar a = 1, b = 2;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 11},
					{MessageId: "missing", Line: 2, Column: 11},
				},
			},
		},
	)
}
