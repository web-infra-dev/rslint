// TestArrowParensExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk / GitHub issue
// it covers, so future refactors can't silently regress them without breaking
// a named lock-in.
//
// Dimension 4 walk for @stylistic/arrow-parens:
//
//   - Receiver / expression wrappers — N/A. The rule fires on the
//     ArrowFunction node itself; surrounding wrappers (`(arrow).then(...)`,
//     `arrow as any`) are out of scope.
//   - Access / key forms — N/A. The rule doesn't inspect property or
//     computed-key access.
//   - Declaration / container forms — covered: arrow with single Identifier
//     vs destructuring (Array/Object binding) vs rest vs default vs typed
//     param; async-modified arrows; arrows used as object/class field
//     initializers; arrows nested inside other arrows.
//   - Nesting / traversal boundaries — covered: arrow-in-arrow where each
//     instance is independently checked; ensures the listener doesn't bleed
//     across boundaries.
//   - Graceful degradation — covered: zero-param `()` (never reported),
//     multi-param `(a, b)` (never reported), single TS optional `(a?)`
//     (keep parens), trailing comma `(a,)`, multi-line inside parens.
//
// Branch walk for upstream's `arrow-parens.ts`:
//
//   - `params.length === 1` filter
//   - `shouldHaveParens = !asNeeded || requireForBlockBody && hasBlockBody`
//   - `findOpeningParenOfParams` (returns paren or null)
//   - Each of the as-needed remove-paren guard clauses:
//     `param.type === 'Identifier'`, `!param.optional`,
//     `!param.typeAnnotation`, `!node.returnType`,
//     `!hasCommentsInParensOfParams`, `!hasUnexpectedTokensBeforeOpeningParen`
//   - Fix's `canTokensBeAdjacent` space-insertion branch
//
// Issue-tracker anchors (real-user shapes):
//
//   - eslint-stylistic#498 — `(opts?) =>` (TS optional, no type annotation)
//   - eslint#12570 — complex nested generics with multiple constraints
//   - eslint#8834 — trailing comma in multi-line arrow params
//   - eslint#6311 / #8682 — async + as-needed interplay
package arrow_parens_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/arrow_parens"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestArrowParensExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&arrow_parens.ArrowParensRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Dimension 4: declaration / container forms
			// ============================================================

			// ---- Dimension 4: arrow inside object literal value ----
			{Code: `const o = { f: (a) => a };`, Options: optsAlways()},
			{Code: `const o = { f: a => a };`, Options: optsAsNeeded()},

			// ---- Dimension 4: arrow as class field initializer ----
			{Code: `class C { handler = (a) => a; }`, Options: optsAlways()},
			{Code: `class C { handler = a => a; }`, Options: optsAsNeeded()},

			// ---- Dimension 4: arrow as default param value of regular fn ----
			// Locks in that the listener fires on the arrow nested inside a
			// FunctionDeclaration's parameter initializer, not on the outer fn.
			{Code: `function f(cb = (a) => a) { return cb; }`, Options: optsAlways()},
			{Code: `function f(cb = a => a) { return cb; }`, Options: optsAsNeeded()},

			// ---- Dimension 4: arrow returned from another arrow's body ----
			{Code: `const curry = (x) => (y) => x + y;`, Options: optsAlways()},
			{Code: `const curry = x => y => x + y;`, Options: optsAsNeeded()},

			// ---- Dimension 4: three-level nested arrow ----
			// Each level visited independently.
			{Code: `const a = x => y => z => x + y + z;`, Options: optsAsNeeded()},
			{Code: `const a = (x) => (y) => (z) => x + y + z;`, Options: optsAlways()},

			// ============================================================
			// Dimension 4: nesting boundaries
			// ============================================================

			// ---- Nesting boundaries: nested arrow that should NOT bleed ----
			// The outer arrow's parens are evaluated independently of the
			// inner arrow. With requireForBlockBody, outer block body needs
			// parens; inner expression body without parens is fine.
			{Code: `(a) => { return b => b + a; }`, Options: optsAsNeededBlock()},

			// ---- Nesting boundaries: arrow inside body of regular function ----
			{Code: `function f() { return a => a; }`, Options: optsAsNeeded()},

			// ============================================================
			// Dimension 4: graceful degradation
			// ============================================================

			// ---- Multi-param arrow under any mode — `params.length === 1` filter ----
			{Code: `(a, b) => a + b`, Options: optsAlways()},
			{Code: `(a, b) => a + b`, Options: optsAsNeeded()},
			{Code: `(a, b) => { return a + b; }`, Options: optsAsNeededBlock()},
			{Code: `(a, b, c) => a + b + c`, Options: optsAsNeeded()},

			// ---- Zero-param arrow ----
			{Code: `() => 42`, Options: optsAlways()},
			{Code: `() => 42`, Options: optsAsNeeded()},
			{Code: `() => { return 42; }`, Options: optsAsNeededBlock()},

			// ---- Empty body / expression body / paren-wrapped body ----
			{Code: `(a) => {}`, Options: optsAlways()},
			{Code: `a => ({})`, Options: optsAsNeeded()},
			{Code: `a => (a + 1)`, Options: optsAsNeeded()},
			{Code: `a => (a)`, Options: optsAsNeededBlock()},

			// ============================================================
			// Branch lock-ins: `param.type === 'Identifier'` guard
			// ============================================================

			// ---- Destructuring patterns: parens stay under as-needed ----
			// In tsgo: ArrayBindingPattern / ObjectBindingPattern under
			// ParameterDeclaration.Name(). Equivalent to ESTree's
			// `params[0].type` being `ArrayPattern` / `ObjectPattern`.
			{Code: `([a]) => a;`, Options: optsAsNeeded()},
			{Code: `([a, b]) => a + b;`, Options: optsAsNeeded()},
			{Code: `([a, , c]) => a + c;`, Options: optsAsNeeded()}, // sparse / hole
			{Code: `([a, ...rest]) => rest.length;`, Options: optsAsNeeded()},
			{Code: `({a}) => a;`, Options: optsAsNeeded()},
			{Code: `({a, b}) => a + b;`, Options: optsAsNeeded()},
			{Code: `({a: x, b: y}) => x + y;`, Options: optsAsNeeded()},   // aliasing
			{Code: `({a = 1, b = 2}) => a + b;`, Options: optsAsNeeded()}, // pattern default
			{Code: `({a, ...rest}) => rest;`, Options: optsAsNeeded()},    // pattern rest

			// ---- Rest param `...a` ----
			// tsgo: paramDecl.DotDotDotToken != nil. ESTree: RestElement
			// (not Identifier).
			{Code: `(...rest) => rest;`, Options: optsAsNeeded()},

			// ---- Default param `a = 10` ----
			// tsgo: paramDecl.Initializer != nil. ESTree: AssignmentPattern
			// (not Identifier).
			{Code: `(a = 1) => a;`, Options: optsAsNeeded()},
			{Code: `(a = () => 1) => a;`, Options: optsAsNeeded()}, // arrow as default value of arrow param

			// ============================================================
			// Branch lock-ins: TS-specific guards (!optional, !type, !returnType)
			// ============================================================

			// ---- TS type annotation on param: keep parens ----
			{Code: `(a: number) => a;`, Options: optsAsNeeded()},
			{Code: `(a: number) => a;`, Options: optsAsNeededBlock()},
			{Code: `(a: string) => a.length;`, Options: optsAsNeeded()},
			{Code: `(a: { x: number }) => a.x;`, Options: optsAsNeeded()},  // object type
			{Code: `(a: number | string) => a;`, Options: optsAsNeeded()},  // union type
			{Code: `(a: Array<number>) => a[0];`, Options: optsAsNeeded()}, // generic type ref

			// ---- TS optional `a?`: keep parens ----
			{Code: `(a?: number) => a;`, Options: optsAsNeeded()},
			{Code: `(a?: number) => a;`, Options: optsAsNeededBlock()},
			// Real-user: eslint-stylistic#498 — `(opts?) => doSomething(opts)`
			// Locks: TS optional WITHOUT a type annotation. Upstream's
			// guard `!param.optional` must fire even when no type follows.
			{Code: `myObj.something = (opts?) => doSomething(opts);`, Options: optsAsNeeded()},
			{Code: `myObj.something = (opts?) => doSomething(opts);`, Options: optsAsNeededBlock()},

			// ---- TS arrow return-type `(a): T =>`: keep parens ----
			{Code: `(a): number => a;`, Options: optsAsNeeded()},
			{Code: `(a): number => a;`, Options: optsAsNeededBlock()},
			{Code: `async (a): Promise<number> => a;`, Options: optsAsNeeded()},
			// Real-user: TS type-guard return type `(x): x is string => ...`
			// — type predicates also use the return-type slot, so parens stay.
			{Code: `(x): x is string => typeof x === 'string';`, Options: optsAsNeeded()},
			{Code: `(x): asserts x is number => { if (typeof x !== 'number') throw new Error(); };`, Options: optsAsNeeded()},

			// ============================================================
			// Branch lock-ins: `!hasUnexpectedTokensBeforeOpeningParen`
			// ============================================================

			// ---- Generics: parens forced by `<T>` ----
			{Code: `<T,>(a) => a;`, Options: optsAsNeeded()},
			{Code: `<T, U>(a: T, b: U) => a;`, Options: optsAsNeeded()},
			{Code: `<const T>(a) => a;`, Options: optsAsNeeded()},
			{Code: `<T extends keyof U, U>(a: T) => a;`, Options: optsAsNeeded()},
			// Real-user: eslint#12570 — nested generic constraint with
			// complex keyof / Extract usage. Body is block, but params is
			// empty so the rule should NOT fire.
			{Code: `const useABTests = <T extends ABTests, TName extends Extract<keyof T, string | number>>() => { return null; };`, Options: optsAsNeeded()},
			// Same shape with single param + complex generic — locks the
			// "generics present means keep parens" path under heavy nesting.
			{Code: `const useABTests = <T extends ABTests, TName extends Extract<keyof T, string | number>>(name: TName) => name;`, Options: optsAsNeeded()},

			// ============================================================
			// Branch lock-ins: `!hasCommentsInParensOfParams`
			// ============================================================

			// ---- Comment position variants inside parens ----
			{Code: `const f = ( a /* trailing block */ ) => a;`, Options: optsAsNeeded()},
			{Code: `const f = ( /* leading block */ a ) => a;`, Options: optsAsNeeded()},
			{Code: "const f = ( a // line\n) => a;", Options: optsAsNeeded()},
			{Code: "const f = (\n  // line comment alone\n  a\n) => a;", Options: optsAsNeeded()},
			{Code: "const f = (a /* line1\nline2 */) => a;", Options: optsAsNeeded()},     // multi-line block comment
			{Code: `const f = (/** JSDoc */ a) => a;`, Options: optsAsNeeded()},           // JSDoc block
			{Code: `const f = (a /* end */ , /* sep */ ) => a;`, Options: optsAsNeeded()}, // comment between param and trailing comma + after

			// ---- Comments OUTSIDE parens — should NOT block paren removal ----
			// Locks the as-needed comment scan window: only the [openParen+1,
			// closeParen) range matters, not surrounding source.
			{Code: `a => a;`, Options: optsAsNeeded()},

			// ============================================================
			// requireForBlockBody flag variations
			// ============================================================

			// ---- requireForBlockBody=false → plain as-needed ----
			{Code: `a => a`, Options: optsAsNeededFlag(false)},
			{Code: `a => ({})`, Options: optsAsNeededFlag(false)},
			{Code: `a => { return a; }`, Options: optsAsNeededFlag(false)}, // block body, parens NOT required

			// ---- requireForBlockBody with arrow-body-is-arrow-with-block ----
			// The outer arrow has an expression body (an inner arrow), the
			// inner arrow has a block body — requireForBlockBody affects each
			// individually.
			{Code: `a => (b) => { return a + b; }`, Options: optsAsNeededBlock()},

			// ============================================================
			// Async modifier combinations
			// ============================================================

			// ---- async + various param shapes under as-needed ----
			{Code: `async a => a;`, Options: optsAsNeeded()},
			{Code: `async (a: number) => a;`, Options: optsAsNeeded()},          // TS type → keep
			{Code: `async (a?) => a;`, Options: optsAsNeeded()},                 // TS optional → keep
			{Code: `async (a): Promise<number> => a;`, Options: optsAsNeeded()}, // return type → keep
			{Code: `async (...args) => args.length;`, Options: optsAsNeeded()},  // rest → keep
			{Code: `async ({a, b}) => a + b;`, Options: optsAsNeeded()},         // destructuring → keep
			{Code: `async <T>(a: T) => a;`, Options: optsAsNeeded()},            // generic → keep
			{Code: `async <T>(/* doc */a: T) => a;`, Options: optsAsNeeded()},   // generic + comment

			// ---- async + requireForBlockBody ----
			{Code: `async a => a`, Options: optsAsNeededBlock()},
			{Code: `async (a) => { return a; }`, Options: optsAsNeededBlock()},
			{Code: `async a => ({a})`, Options: optsAsNeededBlock()}, // body is paren-wrapped expression, not block

			// ============================================================
			// Real-user code patterns (without issue refs)
			// ============================================================

			// ---- Real-user: Promise.then chain ----
			{Code: `fetch('/api').then((res) => res.json()).then((data) => data);`, Options: optsAlways()},
			{Code: `fetch('/api').then(res => res.json()).then(data => data);`, Options: optsAsNeeded()},

			// ---- Real-user: Array.prototype methods ----
			{Code: `items.map(item => item.id);`, Options: optsAsNeeded()},
			{Code: `items.filter(item => item.active);`, Options: optsAsNeeded()},
			{Code: `items.reduce((acc, x) => acc + x, 0);`, Options: optsAsNeeded()}, // multi-param skipped
			{Code: `items.find(item => item.id === target);`, Options: optsAsNeeded()},
			{Code: `items.some(item => item.flag);`, Options: optsAsNeeded()},

			// ---- Real-user: React useState updater pattern ----
			{Code: `setCount(prev => prev + 1);`, Options: optsAsNeeded()},
			{Code: `setItems(prev => [...prev, newItem]);`, Options: optsAsNeeded()},

			// ---- Real-user: useEffect / dependency-array hook ----
			// Zero-param arrow — never reported.
			{Code: `useEffect(() => { document.title = title; }, [title]);`, Options: optsAsNeededBlock()},
			{Code: `useMemo(() => computeValue(), [a]);`, Options: optsAsNeeded()},

			// ---- Real-user: IIFE ----
			{Code: `((x) => x)(1);`, Options: optsAlways()},
			{Code: `(x => x)(1);`, Options: optsAsNeeded()},

			// ---- Real-user: arrow + TS `as` in body ----
			{Code: `(x) => (x as number) + 1;`, Options: optsAlways()},
			{Code: `x => (x as number) + 1;`, Options: optsAsNeeded()},

			// ---- Real-user: arrow + TS `satisfies` in body ----
			{Code: `(x) => x satisfies object;`, Options: optsAlways()},
			{Code: `x => x satisfies object;`, Options: optsAsNeeded()},

			// ---- Real-user: arrow body uses non-null + optional chain ----
			{Code: `x => x!.y;`, Options: optsAsNeeded()},
			{Code: `x => x?.y;`, Options: optsAsNeeded()},

			// ============================================================
			// Operator precedence / placement (where arrow can appear)
			// ============================================================

			// ---- Conditional expression branches ----
			{Code: `flag ? (a) => a : (b) => b;`, Options: optsAlways()},
			{Code: `flag ? a => a : b => b;`, Options: optsAsNeeded()},

			// ---- Sequence expression — multiple arrows ----
			{Code: `(a => a, b => b);`, Options: optsAsNeeded()},

			// ---- Spread argument in call ----
			{Code: `Promise.all([(a) => a, (b) => b]);`, Options: optsAlways()},
			{Code: `Promise.all([a => a, b => b]);`, Options: optsAsNeeded()},

			// ---- Arrow as JSX prop / inside JSX expression container ----
			{Code: `const C = () => <button onClick={(e) => e.preventDefault()}>x</button>;`, Options: optsAlways(), Tsx: true},
			{Code: `const C = () => <button onClick={e => e.preventDefault()}>x</button>;`, Options: optsAsNeeded(), Tsx: true},

			// ---- Arrow inside template substitution ----
			{Code: "const s = `${((a) => a)(1)}`;", Options: optsAlways()},
			{Code: "const s = `${(a => a)(1)}`;", Options: optsAsNeeded()},

			// ---- Arrow in tagged template substitution ----
			{Code: "tag`prefix ${(a) => a} suffix`;", Options: optsAlways()},

			// ============================================================
			// Unicode / encoding boundaries
			// ============================================================

			// ---- Unicode identifier as param name ----
			// `λ` is a valid ES identifier-start character per the Unicode ID
			// spec. Our needsSpaceBeforeOpenParen uses scanner.IsIdentifierPart
			// which handles this correctly.
			{Code: `λ => λ * 2;`, Options: optsAsNeeded()},
			{Code: `(λ) => λ * 2;`, Options: optsAlways()},

			// ---- Param name with leading Unicode (Greek letter) ----
			{Code: `const f = (αβγ) => αβγ;`, Options: optsAlways()},
			{Code: `const f = αβγ => αβγ;`, Options: optsAsNeeded()},

			// ---- Comment containing multi-byte chars ----
			// Tests that the byte-level substring search in hasCommentsBetween
			// (looking for `//` and `/*`) isn't fooled by Unicode in the
			// payload. The presence of multi-byte chars between the comment
			// markers shouldn't matter.
			{Code: `const f = (/* 中文 注释 */ a) => a;`, Options: optsAsNeeded()},
			{Code: "const f = (a /* 中文 */ ) => a;", Options: optsAsNeeded()},

			// ============================================================
			// Robustness boundaries
			// ============================================================

			// ---- Same line, multiple arrows ----
			// Independent reports per arrow; order matches source.
			{Code: `[(a) => a, (b) => b].length;`, Options: optsAlways()},

			// ---- Empty options array ----
			{Code: `(a) => a`, Options: []any{}},

			// ---- Bare-string options form ----
			{Code: `a => a`, Options: "as-needed"},
			{Code: `(a) => a`, Options: "always"},

			// ---- Options as nested array (rule_tester / config-loader path) ----
			// `['as-needed']` is the canonical shape; verify nil inner option.
			{Code: `a => a`, Options: []any{"as-needed"}},

			// ---- Options shape with explicit requireForBlockBody=true and a body ----
			{Code: `a => a`, Options: []any{"as-needed", map[string]any{"requireForBlockBody": true}}},

			// ---- requireForBlockBody set when mode is 'always' — flag IGNORED ----
			// Upstream: `requireForBlockBody = asNeeded && options?.requireForBlockBody`.
			// With 'always' mode, the second-arg flag must be ignored;
			// `a => {}` (no parens, block body) should still report under
			// 'always' (covered in invalid below). Here we lock that
			// `(a) => {}` is VALID — it would be UNEXPECTED if the flag
			// silently activated.
			{Code: `(a) => {}`, Options: []any{"always", map[string]any{"requireForBlockBody": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Token-fusion fix branch
			// ============================================================

			// ---- Non-identifier prefix → no space inserted ----
			// `=` is not part of an identifier; safe to be adjacent.
			{
				Code:    `var x =(a) => a;`,
				Output:  []string{`var x =a => a;`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			// `,` is not part of an identifier; safe.
			{
				Code:    `[1,(a) => a];`,
				Output:  []string{`[1,a => a];`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
				},
			},
			// `?` (ternary) is not part of an identifier; safe.
			{
				Code:    `cond?(a) => a:b;`,
				Output:  []string{`cond?a => a:b;`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
				},
			},
			// ---- Identifier-char prefix on a Unicode keyword-like identifier ----
			// `myFunc` ends in identifier char. Because this is a CallExpression
			// (`myFunc(arg)`) at parse time and `(a) => a` would be the
			// single argument, the inner arrow's parens removal needs no
			// space — the outer `(` belongs to `myFunc`. Locks the
			// `parenPos >= node.Pos()` exclusion path.
			{
				Code:    `myFunc((a) => a);`,
				Output:  []string{`myFunc(a => a);`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},

			// ============================================================
			// Multi-line / trailing comma fix
			// ============================================================

			// ---- Multi-line param with trailing comma ----
			{
				Code:    "(\n  a,\n) => a",
				Output:  []string{`a => a`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 2, Column: 3, EndLine: 2, EndColumn: 4},
				},
			},
			// ---- Real-user: eslint#8834 — multi-line trailing comma + paren body ----
			{
				Code:    "const foo = (\n  bar,\n) => ({});",
				Output:  []string{"const foo = bar => ({});"},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 2, Column: 3, EndLine: 2, EndColumn: 6},
				},
			},

			// ============================================================
			// Nesting — independent reports per arrow
			// ============================================================

			// ---- Two-level nested arrow, both fire (always mode) ----
			{
				Code:   `a => b => a + b`,
				Output: []string{`(a) => (b) => a + b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
					{MessageId: "expectedParens", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				},
			},
			// ---- Two-level nested arrow, both fire (as-needed mode) ----
			{
				Code:    `(a) => (b) => a + b`,
				Output:  []string{`a => b => a + b`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
					{MessageId: "unexpectedParens", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			// ---- Three-level nesting ----
			{
				Code:   `a => b => c => a + b + c`,
				Output: []string{`(a) => (b) => (c) => a + b + c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
					{MessageId: "expectedParens", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
					{MessageId: "expectedParens", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				},
			},

			// ============================================================
			// requireForBlockBody messageId selection
			// ============================================================

			// ---- requireForBlockBody, expression body — unexpectedParensInline ----
			{
				Code:    `(a) => a + 1`,
				Output:  []string{`a => a + 1`},
				Options: optsAsNeededBlock(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParensInline", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},
			// ---- requireForBlockBody, block body — expectedParensBlock ----
			{
				Code:    `a => { return a; }`,
				Output:  []string{`(a) => { return a; }`},
				Options: optsAsNeededBlock(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParensBlock", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- requireForBlockBody falsy on 'always' mode — flag ignored ----
			// 'always' with requireForBlockBody:true should still emit plain
			// `expectedParens` (NOT `expectedParensBlock`) because upstream
			// only sets requireForBlockBody internal flag when asNeeded.
			{
				Code:    `a => {}`,
				Output:  []string{`(a) => {}`},
				Options: []any{"always", map[string]any{"requireForBlockBody": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},

			// ============================================================
			// Trailing-comma / paren-position fixes
			// ============================================================

			// ---- Trailing comma without comment — parens removed ----
			{
				Code:    `var foo = (a,) => b;`,
				Output:  []string{`var foo = a => b;`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- Arrow inside argument list — outer `(` ignored ----
			{
				Code:    `foo((a) => a)`,
				Output:  []string{`foo(a => a)`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				},
			},
			// ---- Arrow wrapped in ParenthesizedExpression ----
			{
				Code:    `((a) => a)`,
				Output:  []string{`(a => a)`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 3, EndLine: 1, EndColumn: 4},
				},
			},
			// ---- Double-wrapped ParenthesizedExpression ----
			// Locks the case where multiple layers of ParenthesizedExpression
			// wrap the arrow but don't interfere with inner-paren detection.
			{
				Code:    `(((a) => a))`,
				Output:  []string{`((a => a))`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 4, EndLine: 1, EndColumn: 5},
				},
			},

			// ============================================================
			// Default mode (no options)
			// ============================================================

			// ---- No options → defaults to 'always' ----
			{
				Code:   `a => a`,
				Output: []string{`(a) => a`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedParens", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- String-only options form ----
			{
				Code:    `(a) => a`,
				Output:  []string{`a => a`},
				Options: "as-needed",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},

			// ============================================================
			// Real-user / operator-placement fixes
			// ============================================================

			// ---- Real-user: arrow in conditional branches ----
			{
				Code:    `const fn = flag ? (a) => a : (b) => b;`,
				Output:  []string{`const fn = flag ? a => a : b => b;`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
					{MessageId: "unexpectedParens", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- Real-user: chained .then with single-arg callback ----
			{
				Code:    `fetch('/').then((res) => res.json()).then((data) => data);`,
				Output:  []string{`fetch('/').then(res => res.json()).then(data => data);`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 18, EndLine: 1, EndColumn: 21},
					{MessageId: "unexpectedParens", Line: 1, Column: 44, EndLine: 1, EndColumn: 48},
				},
			},
			// ---- Real-user: Array.prototype.map under as-needed ----
			{
				Code:    `items.map((item) => item.id);`,
				Output:  []string{`items.map(item => item.id);`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 12, EndLine: 1, EndColumn: 16},
				},
			},
			// ---- Real-user: React setState updater ----
			{
				Code:    `setCount((prev) => prev + 1);`,
				Output:  []string{`setCount(prev => prev + 1);`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
				},
			},
			// ---- Real-user: IIFE ----
			{
				Code:    `((x) => x)(1);`,
				Output:  []string{`(x => x)(1);`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 3, EndLine: 1, EndColumn: 4},
				},
			},
			// ---- Real-user: arrow as JSX prop value ----
			{
				Code:    `const C = () => <button onClick={(e) => e.preventDefault()}>x</button>;`,
				Output:  []string{`const C = () => <button onClick={e => e.preventDefault()}>x</button>;`},
				Options: optsAsNeeded(),
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 35, EndLine: 1, EndColumn: 36},
				},
			},

			// ============================================================
			// Token-fusion guard: all keyword-prefix shapes that can legally
			// precede an arrow's `(` without intervening whitespace
			// ============================================================
			//
			// `async(a) => a` and `yield(a) => a` are in upstream's invalid
			// suite. The keyword set that can legally prefix an arrow
			// without space is small — these locks complete the coverage.

			// ---- `throw(a) => a;` — throw statement + arrow ----
			// `throw` followed by an arrow expression. Removing parens
			// requires a space between `throw` and `a` to avoid fusing the
			// keyword and the parameter into a single identifier token.
			{
				Code:    `function f() { throw(a) => a; }`,
				Output:  []string{`function f() { throw a => a; }`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- `return(a) => a;` — return statement + arrow ----
			{
				Code:    `function g() { return(a) => a; }`,
				Output:  []string{`function g() { return a => a; }`},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},

			// ============================================================
			// FullSignature (JSDoc @type) — must NOT block paren removal
			// ============================================================
			//
			// tsgo's reparser sets `arrowFn.FullSignature` from a JSDoc
			// `@type` tag (see typescript-go/internal/parser/reparser.go).
			// Upstream's `node.returnType` (ESTree) is unaffected by JSDoc,
			// so ESLint also wouldn't block removal here. The rule must
			// follow upstream — only the SYNTACTIC return type
			// (`arrowFn.Type` / `(x): T =>`) blocks removal; FullSignature
			// (JSDoc-inferred) does not.

			// ---- JSDoc @type before arrow, as-needed — parens removed ----
			{
				Code:    "/** @type {(x: number) => string} */\nconst h = (x) => x;",
				Output:  []string{"/** @type {(x: number) => string} */\nconst h = x => x;"},
				Options: optsAsNeeded(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedParens", Line: 2, Column: 12, EndLine: 2, EndColumn: 13},
				},
			},
		},
	)
}
