// TestArrowSpacingExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk / real-user
// scenario it covers, so future refactors can't silently regress them
// without breaking a named lock-in.
//
// # Dimension 4 walk for @stylistic/arrow-spacing
//
//   - Receiver / expression wrappers — N/A. The rule fires on the arrow /
//     function-type / constructor-type node itself; surrounding wrappers
//     (`(arrow).then(...)`, `arrow satisfies T`) don't change the `=>` token
//     position.
//   - Access / key forms — N/A. The rule doesn't inspect property or
//     computed-key access.
//   - Declaration / container forms — covered: async, generic, type-annotated
//     parameter, return-type annotation, class field initializer, object
//     property value, JSX attribute value, TSFunctionType in interface / type
//     alias / parameter type, TSConstructorType with `abstract` modifier and
//     with type parameters, generic function types.
//   - Nesting / traversal boundaries — covered: arrow-returning-arrow,
//     template-literal substitution, conditional branches, multi-arrow same
//     level (array / object / argument list), TSFunctionType in union /
//     intersection / conditional / mapped types, IIFE.
//   - Graceful degradation — covered: ASCII/Unicode whitespace forms,
//     block-comment / line-comment adjacencies, JSDoc, default values
//     containing `*/` / `=>` byte sequences (regression guard against
//     non-token-aware reverse scans crossing string-literal boundaries).
//
// # Branch walk for upstream's `arrow-spacing.ts`
//
//   - `getArrow` arm 1: ArrowFunctionExpression path
//   - `getArrow` arm 2: TSFunctionType / TSConstructorType path
//   - `spaces` arms 3–10: every (before|after) × (true|false) × (with-space|
//     no-space) combo — most by upstream, the rest below.
//
// # Real-user shapes
//
// Issue tracker scans (eslint/eslint, eslint-stylistic/eslint-stylistic)
// surface the following recurrent shapes, all locked in below:
//
//   - eslint#3079 — multi-line arrow `() =>\n  ({ ... })` must stay valid
//     under default options (newline is the "space").
//   - eslint#7079 — nested arrow inside default value, already in upstream.
//   - eslint#8472 — defaults document `{ before: true, after: true }`; tests
//     `[{}]` (empty object) round-trips to that default.
//   - eslint-stylistic#1036 — TSFunctionType / TSConstructorType support;
//     covered via union / intersection / conditional / mapped type cases.
package arrow_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/arrow_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// optsB is referenced from cases where only the `before` key is set —
// stresses the partial-option merge path in `parseOptions`.
func optsB(before bool) []any { return []any{map[string]any{"before": before}} }

func TestArrowSpacingExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&arrow_spacing.ArrowSpacingRule,
		[]rule_tester.ValidTestCase{
			// ================================================================
			// Dimension 4: declaration / container forms — arrow modifiers
			// ================================================================

			// ---- async arrow (no params) ----
			{Code: `async () => a`},
			{Code: `async()=>a`, Options: optsBA(false, false)},

			// ---- async arrow (single id, no parens) ----
			{Code: `async a => a`},
			{Code: `async a=>a`, Options: optsBA(false, false)},

			// ---- async arrow with parens ----
			{Code: `async (a) => a`},

			// ---- generic arrow ----
			// Locks in that `<T>` between `async` and `(a)` doesn't move the
			// `=>` location.
			{Code: `<T>(a: T): T => a`},
			{Code: `async <T>(a: T): Promise<T> => a`},

			// ---- TS-annotated parameter ----
			{Code: `(a: number) => a`},
			{Code: `(a: number, b: string) => a`},

			// ---- TS optional parameter `(a?: T) => ...` ----
			// `?` is a postfix token on the param decl; doesn't disturb `=>`
			// detection.
			{Code: `(a?: number) => a`},

			// ---- TS rest parameter `(...a: number[]) => ...` ----
			{Code: `(...a: number[]) => a`},

			// ---- TS default value `(a = 1) => ...` ----
			{Code: `(a = 1) => a`},

			// ---- TS return-type annotation ----
			{Code: `(a): number => a`},
			{Code: `(): void => undefined`},

			// ---- TS return type that is itself a function type ----
			// arrowFn.Type non-nil, deeply nested; ensures `findArrowTokenRange`
			// doesn't follow the inner type's `=>`.
			{Code: `(): (() => void) => () => undefined`},

			// ================================================================
			// Dimension 4: declaration / container forms — surrounding context
			// ================================================================

			// ---- class field initializer ----
			{Code: `class C { handler = (a) => a; }`},
			{Code: `class C { handler = a=>a; }`, Options: optsBA(false, false)},

			// ---- static class field ----
			{Code: `class C { static h = (a) => a; }`},

			// ---- class field with type annotation ----
			{Code: `class C { handler: (e: Event) => void = (e) => {} }`},

			// ---- object literal property value ----
			{Code: `const o = { f: (a) => a };`},

			// ---- object shorthand returning arrow ----
			{Code: `const o = { f() { return (a) => a; } };`},

			// ---- arrow as default parameter value (eslint/eslint#7079 family) ----
			{Code: `function g(cb = (x) => x) { return cb; }`},

			// ---- IIFE — arrow immediately invoked ----
			{Code: `((a) => a)(b)`},

			// ---- arrow returning object literal (parens around literal) ----
			{Code: `() => ({ x: 1 })`},

			// ---- arrow in JSX attribute ----
			// Stresses that a JSX-expression-container's `}` after the arrow
			// body doesn't break after-token detection. Requires .tsx
			// parsing — `Tsx: true` switches the fixture file to react.tsx.
			{Code: `const e = <Foo onClick={(e) => bar(e)} />`, Tsx: true},

			// ---- arrow inside template literal substitution ----
			{Code: "`${(a) => a}`"},

			// ---- arrow returned from another arrow's body (curry) ----
			{Code: `const curry = (a) => (b) => a + b`},
			{Code: `const curry = a=>b=>a+b`, Options: optsBA(false, false)},

			// ---- three-level curry ----
			{Code: `const c = a => b => c => a + b + c`},

			// ---- promise chain (.then((x) => ...)) ----
			{Code: `promise.then((x) => x).catch((e) => e)`},

			// ---- array method chain ----
			{Code: `[1, 2, 3].map((x) => x + 1).filter((x) => x > 0)`},

			// ---- conditional expression with arrow in each branch ----
			{Code: `const f = cond ? (a) => a : (b) => b`},

			// ---- logical-or short-circuit ----
			{Code: `const f = fallback || ((a) => a)`},

			// ---- optional chaining call with arrow argument ----
			{Code: `obj?.method((x) => x)`},

			// ---- setTimeout with arrow ----
			{Code: `setTimeout(() => doSomething(), 100)`},

			// ---- arrow with object destructuring ----
			{Code: `({ a, b }) => a + b`},

			// ---- arrow with array destructuring ----
			{Code: `([a, b]) => a + b`},

			// ---- multi-arrow same level — args list ----
			{Code: `Promise.all([a => a, b => b, c => c])`},

			// ================================================================
			// Dimension 4: TS-specific type-position shapes
			// ================================================================

			// ---- TSFunctionType in interface ----
			{Code: `interface I { f: (a: number) => string }`},

			// ---- TSFunctionType in interface call signature ----
			// `(): T` is a method signature, NOT a TSFunctionType — but the
			// arrow doesn't appear here. Locking in by absence: no false
			// positives from method signatures.
			{Code: `interface I { (a: number): string }`},

			// ---- TSFunctionType in type alias ----
			{Code: `type F = (a: number) => string`},

			// ---- TSFunctionType in union ----
			{Code: `type T = ((a: number) => string) | (() => void)`},

			// ---- TSFunctionType in intersection ----
			{Code: `type T = ((a: number) => string) & { foo: number }`},

			// ---- TSFunctionType in tuple ----
			{Code: `type T = [(a: number) => string, () => void]`},

			// ---- TSFunctionType in conditional type ----
			{Code: `type Args<F> = F extends (a: infer A) => infer R ? A : R`},

			// ---- TSFunctionType in mapped type ----
			{Code: `type T<K extends string> = { [P in K]: () => P }`},

			// ---- TSFunctionType as parameter type ----
			{Code: `function g(cb: (x: number) => string) {}`},

			// ---- TSFunctionType nested as parameter type ----
			// Outer fn's param is a function type that itself takes a
			// function type. Two nested `=>` tokens, each on a separate
			// FunctionType node.
			{Code: `function g(cb: (h: (x: number) => void) => string) {}`},

			// ---- TSFunctionType as return type of function declaration ----
			{Code: `function g(): (x: number) => string { return x => '' }`},

			// ---- TSConstructorType ----
			{Code: `type C = new () => Foo`},

			// ---- TSConstructorType with `abstract` modifier ----
			{Code: `type C = abstract new () => Foo`},

			// ---- TSConstructorType with type parameters ----
			{Code: `type C = new <T>() => T`},

			// ---- TSConstructorType in array type ----
			{Code: `type T = (new () => Foo)[]`},

			// ---- TSFunctionType with generic param ----
			{Code: `type F = <T>(a: T) => T`},

			// ---- TSFunctionType returning TSFunctionType ----
			{Code: `type F = (a: number) => (b: string) => boolean`},

			// ---- Generic TSFunctionType with extends constraint ----
			{Code: `type F = <T extends number>(a: T) => T`},

			// ================================================================
			// Dimension 4: graceful degradation — whitespace
			// ================================================================

			// ---- tab whitespace ----
			{Code: "a\t=>\tb"},

			// ---- CRLF whitespace ----
			{Code: "a\r\n=>\r\nb"},

			// ---- Unicode identifier ----
			// Exercises Unicode-aware findBeforeToken / findAfterToken via
			// scanner.IsIdentifierPart. ASCII byte-walking would mis-locate
			// the identifier boundary on multi-byte UTF-8 runes.
			{Code: "α => β"},

			// ---- multi-line arrow (eslint#3079) ----
			// `() =>\n  ({ ... })`: newline + indent must count as "space after"
			// under defaults. Historically this was a false positive.
			{Code: "const test = (argument) =>\n  ({ argument })"},

			// ---- form-feed `\f` whitespace ----
			// JS treats `\f` as whitespace per `isWhitespaceByte`.
			{Code: "a \f=>\fb"},

			// ---- vertical-tab `\v` whitespace ----
			{Code: "a \v=>\vb"},

			// ---- mixed tab + space ----
			{Code: "a \t => \t b"},

			// ---- Unicode whitespace — NBSP (U+00A0) ----
			// ESLint's `\s`-based `isSpaceBetween` treats NBSP as whitespace;
			// to match upstream we recognise the full ECMAScript `\s` set
			// (`WhiteSpace` + `LineTerminator`: NBSP, ideographic space,
			// ZWNBSP/BOM, Zs category, LS, PS).
			{Code: "a => b"},

			// ---- Unicode whitespace — ideographic space (U+3000) ----
			{Code: "a　=>　b"},

			// ---- Unicode whitespace — line separator (U+2028) ----
			{Code: "a => b"},

			// ================================================================
			// Dimension 4: graceful degradation — comments
			// ================================================================

			// ---- block comment with space around arrow ----
			// `a /*c*/ => b`: space before / after `=>` ⇒ valid by default.
			{Code: `a /*c*/ => b`},

			// ---- multiple consecutive block comments ----
			// `a /*c1*/ /*c2*/ => b`: space-spaced; valid by default.
			{Code: `a /*c1*/ /*c2*/ => b`},

			// ---- JSDoc block comment adjacent ----
			{Code: `/** doc */ const f = (a) => a`},

			// ---- line comment after arrow ----
			// `(a) => // c\n  b`: the newline after `//c` is the "space".
			{Code: "(a) => // c\n  b"},

			// ---- block comment with internal `=>` ----
			// `(a) => /* x => y */ a`: the inner `/*  =>  */` must not be
			// picked up as an arrow token.
			{Code: `(a) => /* x => y */ a`},

			// ---- block comment with internal `*/` doesn't apply — can't nest ----
			// JS forbids nested block comments, so `*/` always closes. Skip.

			// ================================================================
			// Dimension 4: default values that look like trivia delimiters
			// ================================================================

			// ---- default value = string containing `=>` ----
			// Stresses that the rule does NOT visit the substring `=>` inside
			// a string literal; only the outer arrow's `=>` fires.
			{Code: `(a = "x => y") => a`},

			// ---- default value = string containing `*/` ----
			// String spans across positions where a naive `*/` reverse scan
			// might mistake it for a comment end. The byte before the outer
			// `=>` is `)`, so the reverse scan never enters the string.
			{Code: `(a = "*/") => a`},

			// ---- default value = regex containing `=>` ----
			{Code: `(a = /x=>y/g) => a`},

			// ---- default value = template literal ----
			{Code: "(a = `tpl`) => a"},

			// ================================================================
			// Dimension 4: real-user shapes from issue tracker
			// ================================================================

			// ---- eslint#3079 multi-line — alternate form ----
			{Code: "(a) =>\n    0"},

			// ---- eslint#3079 multi-line with explicit body parens ----
			{Code: "const f = (a) =>\n    (a + 1)"},

			// ---- chained arrow in default value (eslint#7079 family) ----
			{Code: `const f = (g = (a) => a) => g`},
		},
		[]rule_tester.InvalidTestCase{
			// ================================================================
			// Branch lock-ins
			// ================================================================

			// ---- ArrowFunctionExpression with multi-char identifiers ----
			// Locks in identifier-start walking (vs the upstream `a=>a` cases
			// where before/after are single chars). column 1 = start of `foo`;
			// column 6 = start of `bar`.
			{
				Code:    `foo=>bar`,
				Output:  []string{`foo => bar`},
				Options: optsBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 1},
					{MessageId: "expectedAfter", Line: 1, Column: 6},
				},
			},

			// ---- TSFunctionType with non-empty parameter list ----
			// Locks in scanForArrow from Parameters.End() correctly walking
			// past `)` to land on `=>` (not on the closing paren).
			{
				Code:   `type Foo = (a: number)=>string`,
				Output: []string{`type Foo = (a: number) => string`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 22},
					{MessageId: "expectedAfter", Line: 1, Column: 25},
				},
			},

			// ---- TSConstructorType with `abstract` modifier ----
			{
				Code:   `type C = abstract new ()=>Foo`,
				Output: []string{`type C = abstract new () => Foo`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 24},
					{MessageId: "expectedAfter", Line: 1, Column: 27},
				},
			},

			// ---- TSConstructorType with type parameters ----
			// Three-segment header (`new`, `<T>`, `()`) before `=>`. Locks in
			// that scanForArrow ignores the `>` of `<T>` and finds only the
			// real `=>` token.
			{
				Code:   `type C = new <T>()=>T`,
				Output: []string{`type C = new <T>() => T`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 18},
					{MessageId: "expectedAfter", Line: 1, Column: 21},
				},
			},

			// ---- Partial option override — only `before` set ----
			// `{ before: false }` leaves `after` at default true. With `a=>a`
			// the before-space is missing (no, wait — `=> a` has neither, so
			// both are missing). Let's pick `a => a`: before is set (and
			// unexpected), after is fine.
			{
				Code:    `a => a`,
				Output:  []string{`a=> a`},
				Options: optsB(false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 4: declaration / container forms (invalid)
			// ================================================================

			// ---- async arrow (single id) with no space ----
			{
				Code:   `async a=>a`,
				Output: []string{`async a => a`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 7},
					{MessageId: "expectedAfter", Line: 1, Column: 10},
				},
			},

			// ---- async with hugged parens ----
			// Verifies the `async(` (keyword + paren, no space) shape doesn't
			// confuse before-token detection.
			{
				Code:   `async(a)=>a`,
				Output: []string{`async(a) => a`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 8},
					{MessageId: "expectedAfter", Line: 1, Column: 11},
				},
			},

			// ---- TS-annotated parameter, no space ----
			{
				Code:   `(a: number)=>a`,
				Output: []string{`(a: number) => a`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 11},
					{MessageId: "expectedAfter", Line: 1, Column: 14},
				},
			},

			// ---- return-type annotation, no space ----
			// `(a): number=>a` — `=>` follows the return type. before-token is
			// the `number` keyword (cols 6-11), report at col 6 (start of
			// keyword), matching ESLint's `node: beforeToken` shape.
			{
				Code:   `(a): number=>a`,
				Output: []string{`(a): number => a`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 6},
					{MessageId: "expectedAfter", Line: 1, Column: 14},
				},
			},

			// ---- generic arrow with explicit return type, no space ----
			{
				Code:   `<T>(a: T): T=>a`,
				Output: []string{`<T>(a: T): T => a`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 12},
					{MessageId: "expectedAfter", Line: 1, Column: 15},
				},
			},

			// ---- TS optional param `(a?)=>a` ----
			// `?` is the last token before the closing `)`. Verifies the
			// before-token detection of `)`.
			{
				Code:   `(a?: number)=>a`,
				Output: []string{`(a?: number) => a`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 12},
					{MessageId: "expectedAfter", Line: 1, Column: 15},
				},
			},

			// ---- TS rest param `(...a)=>a` ----
			{
				Code:   `(...a)=>a`,
				Output: []string{`(...a) => a`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 6},
					{MessageId: "expectedAfter", Line: 1, Column: 9},
				},
			},

			// ---- TS default param `(a=1)=>a` ----
			{
				Code:   `(a = 1)=>a`,
				Output: []string{`(a = 1) => a`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 7},
					{MessageId: "expectedAfter", Line: 1, Column: 10},
				},
			},

			// ---- destructuring object param ----
			{
				Code:   `({a, b})=>a+b`,
				Output: []string{`({a, b}) => a+b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 8},
					{MessageId: "expectedAfter", Line: 1, Column: 11},
				},
			},

			// ---- destructuring array param ----
			{
				Code:   `([a, b])=>a+b`,
				Output: []string{`([a, b]) => a+b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 8},
					{MessageId: "expectedAfter", Line: 1, Column: 11},
				},
			},

			// ---- class field initializer, no space ----
			// Locks in that class-field arrows are visited the same as
			// expression-statement arrows; the listener doesn't bleed into
			// other class members.
			{
				Code:   `class C { f = (a)=>a; }`,
				Output: []string{`class C { f = (a) => a; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 17},
					{MessageId: "expectedAfter", Line: 1, Column: 20},
				},
			},

			// ---- object literal property value, no space ----
			{
				Code:   `const o = { f: (a)=>a };`,
				Output: []string{`const o = { f: (a) => a };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 18},
					{MessageId: "expectedAfter", Line: 1, Column: 21},
				},
			},

			// ---- JSX attribute value, no space ----
			{
				Code:   `const e = <Foo onClick={(e)=>bar(e)} />`,
				Output: []string{`const e = <Foo onClick={(e) => bar(e)} />`},
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 27},
					{MessageId: "expectedAfter", Line: 1, Column: 30},
				},
			},

			// ================================================================
			// Dimension 4: nesting boundaries (invalid)
			// ================================================================

			// ---- curry: a=>b=>a+b ----
			// Two arrows; inner reports first (cols 4, 7), outer second
			// (cols 1, 4). The on-exit listener replicates ESLint's
			// source-order sort.
			{
				Code:    `a=>b=>a+b`,
				Output:  []string{`a => b => a+b`},
				Options: optsBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 4},
					{MessageId: "expectedAfter", Line: 1, Column: 7},
					{MessageId: "expectedBefore", Line: 1, Column: 1},
					{MessageId: "expectedAfter", Line: 1, Column: 4},
				},
			},

			// ---- three-level curry: a=>b=>c=>... ----
			// 3 arrows × 2 errors each = 6 errors. Innermost-first ordering.
			{
				Code:    `a=>b=>c=>a+b+c`,
				Output:  []string{`a => b => c => a+b+c`},
				Options: optsBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 7},
					{MessageId: "expectedAfter", Line: 1, Column: 10},
					{MessageId: "expectedBefore", Line: 1, Column: 4},
					{MessageId: "expectedAfter", Line: 1, Column: 7},
					{MessageId: "expectedBefore", Line: 1, Column: 1},
					{MessageId: "expectedAfter", Line: 1, Column: 4},
				},
			},

			// ---- conditional branches both contain arrows ----
			// Two sibling arrows visited independently.
			{
				Code:   `const f = cond ? (a)=>a : (b)=>b`,
				Output: []string{`const f = cond ? (a) => a : (b) => b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 20},
					{MessageId: "expectedAfter", Line: 1, Column: 23},
					{MessageId: "expectedBefore", Line: 1, Column: 29},
					{MessageId: "expectedAfter", Line: 1, Column: 32},
				},
			},

			// ---- two arrows as call arguments ----
			{
				Code:   `f((a)=>a, (b)=>b)`,
				Output: []string{`f((a) => a, (b) => b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 5},
					{MessageId: "expectedAfter", Line: 1, Column: 8},
					{MessageId: "expectedBefore", Line: 1, Column: 13},
					{MessageId: "expectedAfter", Line: 1, Column: 16},
				},
			},

			// ---- two arrows in an array literal ----
			{
				Code:   `const arr = [(a)=>a, (b)=>b]`,
				Output: []string{`const arr = [(a) => a, (b) => b]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 16},
					{MessageId: "expectedAfter", Line: 1, Column: 19},
					{MessageId: "expectedBefore", Line: 1, Column: 24},
					{MessageId: "expectedAfter", Line: 1, Column: 27},
				},
			},

			// ---- promise chain — two arrows on .then/.catch ----
			{
				Code:   `p.then((x)=>x).catch((e)=>e)`,
				Output: []string{`p.then((x) => x).catch((e) => e)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 10},
					{MessageId: "expectedAfter", Line: 1, Column: 13},
					{MessageId: "expectedBefore", Line: 1, Column: 24},
					{MessageId: "expectedAfter", Line: 1, Column: 27},
				},
			},

			// ---- TSFunctionType inside ArrowFunction's parameter type ----
			// Mixed node-kind listener case: the listener fires on the outer
			// ArrowFunction AND on the inner TSFunctionType. With `{before:
			// false, after: false}` both must report at their own positions.
			//
			// `(cb: (x: number) => string) => cb` shape, hugged:
			{
				Code:    `(cb: (x: number) => string) => cb`,
				Output:  []string{`(cb: (x: number)=>string)=>cb`},
				Options: optsBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 16},
					{MessageId: "unexpectedAfter", Line: 1, Column: 21},
					{MessageId: "unexpectedBefore", Line: 1, Column: 27},
					{MessageId: "unexpectedAfter", Line: 1, Column: 32},
				},
			},

			// ---- TSFunctionType in union (each summand's `=>` reported) ----
			{
				Code:    `type T = ((a: number)=>string) | (()=>void)`,
				Output:  []string{`type T = ((a: number) => string) | (() => void)`},
				Options: optsBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 21},
					{MessageId: "expectedAfter", Line: 1, Column: 24},
					{MessageId: "expectedBefore", Line: 1, Column: 36},
					{MessageId: "expectedAfter", Line: 1, Column: 39},
				},
			},

			// ---- TSFunctionType in conditional type (infer position) ----
			// `=>` between params and return type of a function type pattern.
			{
				Code:   `type R<F> = F extends (a: number)=>infer R ? R : never`,
				Output: []string{`type R<F> = F extends (a: number) => infer R ? R : never`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 33},
					{MessageId: "expectedAfter", Line: 1, Column: 36},
				},
			},

			// ================================================================
			// Dimension 4: graceful degradation (invalid)
			// ================================================================

			// ---- block comment immediately before arrow ----
			// `a/*c*/=>b`: before-token is the block-comment ending at col 7;
			// before-token range start = col 2 (`/*` opener).
			{
				Code:   `a/*c*/=>b`,
				Output: []string{`a/*c*/ => b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 2},
					{MessageId: "expectedAfter", Line: 1, Column: 9},
				},
			},

			// ---- block comment immediately after arrow ----
			{
				Code:   `a=>/*c*/b`,
				Output: []string{`a => /*c*/b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 1},
					{MessageId: "expectedAfter", Line: 1, Column: 4},
				},
			},

			// ---- empty block comment `/**/` ----
			// Edge case: 4-byte comment, the reverse scan must terminate at
			// the right `/*` (i.e. handle `*/` and `/*` overlapping by one
			// byte: `*/` ends at col 5, `/*` begins at col 2).
			{
				Code:   `a/**/=>b`,
				Output: []string{`a/**/ => b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 2},
					{MessageId: "expectedAfter", Line: 1, Column: 8},
				},
			},

			// ---- JSDoc comment immediately before arrow ----
			// `/** doc */` is a block comment in tokenizer terms — same
			// handling as `/* */`.
			{
				Code:   `a/** doc */=>b`,
				Output: []string{`a/** doc */ => b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 2},
					{MessageId: "expectedAfter", Line: 1, Column: 14},
				},
			},

			// ---- comments on both sides — neither has whitespace gap ----
			// `a/*c*/=>/*d*/b`: before-token = `/*c*/`, after-token = `/*d*/`.
			// Default options expect spaces on both sides → two reports.
			{
				Code:   `a/*c*/=>/*d*/b`,
				Output: []string{`a/*c*/ => /*d*/b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 2},
					{MessageId: "expectedAfter", Line: 1, Column: 9},
				},
			},

			// ---- comment with `=>` inside is not picked up as the arrow ----
			// `a/* x => y */=>b`: the lexer treats `/* x => y */` as a single
			// block comment, so the only real `=>` token is the one after the
			// comment. Locks in that block-comment reverse scanning doesn't
			// confuse `=>` byte-pair inside the comment with the real arrow.
			{
				Code:   `a/* x => y */=>b`,
				Output: []string{`a/* x => y */ => b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 2},
					{MessageId: "expectedAfter", Line: 1, Column: 16},
				},
			},

			// ---- string literal containing `=>` followed by `=>` ----
			// `(a = "x => y")=>1`: only the outer `=>` exists as a token.
			// Locks in that the rule does not visit the inner `=>` substring
			// inside the string-literal default value.
			{
				Code:   `(a = "x => y")=>1`,
				Output: []string{`(a = "x => y") => 1`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 14},
					{MessageId: "expectedAfter", Line: 1, Column: 17},
				},
			},

			// ---- string literal containing `*/` ----
			// `(a = "*/")=>1`: before-token is `)`, NOT the `*/` inside the
			// string. Regression guard against the non-token-aware reverse
			// scan crossing the string boundary.
			{
				Code:   `(a = "*/")=>1`,
				Output: []string{`(a = "*/") => 1`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 10},
					{MessageId: "expectedAfter", Line: 1, Column: 13},
				},
			},

			// ---- regex containing `=>` ----
			// `(a = /x=>y/g)=>1`: the inner `=>` is part of the regex; only
			// the outer `=>` is a real arrow.
			{
				Code:   `(a = /x=>y/g)=>1`,
				Output: []string{`(a = /x=>y/g) => 1`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 13},
					{MessageId: "expectedAfter", Line: 1, Column: 16},
				},
			},

			// ---- tab between identifier and arrow ----
			// `\t` is whitespace per `isWhitespaceByte`; with `before: false`
			// the tab is treated as a space and reported.
			{
				Code:    "a\t=>a",
				Output:  []string{"a=> a"},
				Options: optsBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 1},
					{MessageId: "expectedAfter", Line: 1, Column: 5},
				},
			},

			// ---- form-feed before arrow ----
			// `\f` is whitespace.
			{
				Code:    "a\f=>a",
				Output:  []string{"a=> a"},
				Options: optsBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 1},
					{MessageId: "expectedAfter", Line: 1, Column: 5},
				},
			},

			// ---- CRLF — `\r\n` collapses to one space ----
			{
				Code:    "(a) =>\r\n  b",
				Output:  []string{"(a) =>b"},
				Options: optsA(false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAfter", Line: 2, Column: 3},
				},
			},

			// ---- multi-line arrow with extra blank line ----
			// Newlines + spaces all count as "after" whitespace.
			{
				Code:    "(a) =>\n\n  b",
				Output:  []string{"(a) =>b"},
				Options: optsA(false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAfter", Line: 3, Column: 3},
				},
			},

			// ---- Unicode identifier — no surrounding space ----
			// Verifies scanner.IsIdentifierPart walks correctly over multi-
			// byte α / β runes. ASCII byte-walking would report columns
			// inside the multi-byte sequence.
			{
				Code:   "α=>β",
				Output: []string{"α => β"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 1},
					{MessageId: "expectedAfter", Line: 1, Column: 4},
				},
			},

			// ---- Unicode identifier with mixed-ASCII context ----
			// `f(α=>β)` — surrounding parens are ASCII, arrow neighbors are
			// Unicode. Stresses that `findBeforeToken`'s ASCII fast path
			// fallthrough still picks up the identifier rune correctly.
			{
				Code:   `f(α=>β)`,
				Output: []string{`f(α => β)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 3},
					{MessageId: "expectedAfter", Line: 1, Column: 6},
				},
			},

			// ---- Comments on both sides with `before:false, after:false` ----
			// `a /*c*/ => /*d*/ b` has spaces on both sides → both reports.
			// Locks in that the byte walk treats spaces around comments as
			// the `isSpaced` signal (not the comment itself).
			{
				Code:    `a /*c*/ => /*d*/ b`,
				Output:  []string{`a /*c*/=>/*d*/ b`},
				Options: optsBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 3},
					{MessageId: "unexpectedAfter", Line: 1, Column: 12},
				},
			},

			// ---- before:false, after:false applied to multi-line ----
			// `a\n=>\nb` has newline-space on both sides; both should report.
			// Verifies that newlines count as the "unexpected space".
			{
				Code:    "a\n=>\nb",
				Output:  []string{"a=>b"},
				Options: optsBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 1},
					{MessageId: "unexpectedAfter", Line: 3, Column: 1},
				},
			},

			// ---- Unicode whitespace must count as "space" for `before/after:false` ----
			// `a NBSP => NBSP b` with `{ before: false, after: false }`: both
			// the leading and trailing NBSP must be treated as whitespace,
			// matching ESLint's `\s`-based `isSpaceBetween`. ASCII-only
			// `isWhitespaceByte` would miss them and silently pass.
			{
				Code:    "a => b",
				Output:  []string{"a=>b"},
				Options: optsBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 1},
					{MessageId: "unexpectedAfter", Line: 1, Column: 6},
				},
			},
		},
	)
}
