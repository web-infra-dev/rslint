// TestArrayBracketSpacingExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking
// a named lock-in.
//
// Dimension 4 walk for @stylistic/array-bracket-spacing:
//
//   - Receiver / expression wrappers (outside the array) — N/A. The rule
//     fires on the array literal / array binding pattern itself; outer
//     receiver wrappers (`(arr).x`, `arr!.x`, `arr?.x`) sit outside the
//     listener. Wrappers in ELEMENT position are highly relevant — see
//     "paren-wrapped element" + "TS type wrappers must NOT match" lock-ins.
//   - Access / key forms — N/A. The rule doesn't inspect property accesses
//     or computed keys; it only looks at the array brackets and a slice of
//     trivia on either side.
//   - Declaration / container forms — covered: array literal in variable
//     init, function param default, JSX prop, decorator argument, hook
//     deps, template substitution, `as const` wrapper, ObjectLit property
//     value, callback param, for-of binding, arrow param.
//   - Nesting / traversal boundaries — covered: same-kind nesting up to
//     three levels (`[[[1]]]`), nested empty (`[[]]`), nested + rest
//     (`[[a, ...b], c]`), and the inner-array-bracket exception
//     interleaving (see the enter/exit emit order discussion in the
//     rule's package comment).
//   - Graceful degradation — covered: empty arrays in both modes, single-
//     element arrays, sparse arrays with omitted first / last elements,
//     trailing comma in destructuring, comments between brackets and
//     tokens, comments WITH `]` / `[` inside their payload, string /
//     template literals containing `]` or `[`, multi-byte (UTF-8)
//     element content.
//
// Options-resilience subgroup covers: empty options array, unknown keys
// in the second arg, non-bool exception values — all silently ignored
// (rslint does not run JSON-schema validation that upstream's loader does).
package array_bracket_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/array_bracket_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestArrayBracketSpacingExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&array_bracket_spacing.ArrayBracketSpacingRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: paren-wrapped element triggers objectsInArrays ----
			// tsgo preserves the ParenthesizedExpression that ESLint's TS
			// parser flattens; without SkipParentheses on the first/last
			// element this would silently miss the exception. With the
			// exception firing, opening + closing must be spaced even
			// though the rule's primary mode is "never".
			{Code: `var foo = [ ({ 'bar': 'baz' }), 1, 5];`, Options: []interface{}{"never", map[string]interface{}{"objectsInArrays": true}}},
			{Code: `var foo = [1, 5, ({ 'bar': 'baz' }) ];`, Options: []interface{}{"never", map[string]interface{}{"objectsInArrays": true}}},
			// ---- Dimension 4: paren-wrapped element triggers arraysInArrays ----
			{Code: `var arr = [ ([1, 2]), 2, 3, 4];`, Options: []interface{}{"never", map[string]interface{}{"arraysInArrays": true}}},

			// ---- Dimension 4: same-kind nesting, three levels deep ----
			// Verifies the rule visits every layer (enter+exit) and emits in
			// document order, not in visit (parent-first) order.
			{Code: `var arr = [ [ [ 1 ] ] ];`, Options: []interface{}{"always"}},
			{Code: `var arr = [[[1]]];`, Options: []interface{}{"never"}},

			// ---- Dimension 4: TS type wrappers must NOT match the exception ----
			// Upstream's rule reads `node.type === 'ObjectExpression'` and
			// returns false for TSAsExpression / TSSatisfiesExpression /
			// TSNonNullExpression — even when the wrapped value is an
			// object. Our skipping is paren-only, so these should fall
			// through to the default spaced behavior. With "never" and
			// objectsInArrays:true the FIRST element check sees AsExpression
			// (not ObjectLiteralExpression), so the default 'never' wins
			// and no leading space is required.
			{Code: `var foo = [{a: 1} as any, 2];`, Options: []interface{}{"never", map[string]interface{}{"objectsInArrays": true}}},
			{Code: `var foo = [{a: 1} satisfies object, 2];`, Options: []interface{}{"never", map[string]interface{}{"objectsInArrays": true}}},
			{Code: `var foo = [{a: 1}!, 2];`, Options: []interface{}{"never", map[string]interface{}{"objectsInArrays": true}}},

			// ---- Dimension 4: TS array binding with type annotation on parent ----
			// In tsgo the type annotation lives on VariableDeclaration, not
			// on the binding pattern itself, so node.End() of the binding
			// pattern stops at `]` without us needing to consult any
			// type-annotation field. Upstream covers this via babel/flow;
			// we lock the simpler tsgo path. The type annotation
			// `[number, number]` is a TupleType node and not visited by
			// this rule, so it can stay tight regardless of mode.
			{Code: `var [ a, b ]: [number, number] = [ 1, 2 ];`, Options: []interface{}{"always"}},
			{Code: `var [a, b]: [number, number] = [1, 2];`, Options: []interface{}{"never"}},
			{Code: `function f([ a, b ]: [number, number]) {}`, Options: []interface{}{"always"}},
			{Code: `function f([a, b]: [number, number]) {}`, Options: []interface{}{"never"}},

			// ---- Dimension 4: arrow param destructuring with type annotation ----
			{Code: `const f = ([a, b]: [number, number]) => a + b;`, Options: []interface{}{"never"}},
			{Code: `const f = ([ a, b ]: [number, number]) => a + b;`, Options: []interface{}{"always"}},

			// ---- Dimension 4: for-of destructuring ----
			{Code: `for (const [a, b] of arr) {}`, Options: []interface{}{"never"}},
			{Code: `for (const [ a, b ] of arr) {}`, Options: []interface{}{"always"}},

			// ---- Dimension 4: nested binding with object pattern ----
			// BindingElement.Name() is ObjectBindingPattern → matches the
			// objectsInArrays exception under our bindingElementBound
			// unwrap. Without the unwrap this would fall through to
			// `KindBindingElement` and silently miss the exception.
			{Code: `var [ { x, y }, z] = arr;`, Options: []interface{}{"never", map[string]interface{}{"objectsInArrays": true}}},
			{Code: `var [z, { x, y } ] = arr;`, Options: []interface{}{"never", map[string]interface{}{"objectsInArrays": true}}},

			// ---- Dimension 4: rest element (BindingRest) does NOT match exceptions ----
			// `bindingElementBound` keeps DotDotDotToken-bearing
			// BindingElements as-is so the rest target doesn't accidentally
			// be classified as the inner array/object kind.
			{Code: `var [...rest] = arr;`, Options: []interface{}{"never", map[string]interface{}{"arraysInArrays": true}}},
			{Code: `var [ ...rest ] = arr;`, Options: []interface{}{"always", map[string]interface{}{"arraysInArrays": false}}},

			// ---- Dimension 4: line comment near closing bracket ----
			// A `//` line comment forces a newline before `]`, which the
			// `isTokenOnSameLine` (newline) check uses to skip the closing
			// report. Mirrors upstream's behavior under
			// getTokenBefore({includeComments:false}).
			{Code: "var arr = [\n  1,\n  2 // comment\n];", Options: []interface{}{"never"}},

			// ---- Dimension 4: trailing block comment with newline ----
			{Code: "var arr = [\n  1,\n  2 /* comment */\n];", Options: []interface{}{"never"}},

			// ---- Real-user: array in function argument with always-spacing ----
			// Common shape that previously triggered false reports when the
			// nested array exception was applied to the wrong layer.
			{Code: `fn([ a, b ], { c });`, Options: []interface{}{"always", map[string]interface{}{"objectsInArrays": false}}},

			// ---- Real-user: array literal in object property value ----
			{Code: `const obj = { items: [1, 2, 3] };`, Options: []interface{}{"never"}},
			{Code: `const obj = { items: [ 1, 2, 3 ] };`, Options: []interface{}{"always"}},

			// ---- Locks in upstream validateArraySpacing() arm: empty array under 'always' early-return ----
			// `if (options.spaced && node.elements.length === 0) return` —
			// upstream doesn't iterate further, so the input `[ ]` is
			// valid under 'always' even though there IS a space between
			// the brackets.
			{Code: `var foo = [ ];`, Options: []interface{}{"always"}},

			// ---- Locks in upstream validateArraySpacing() arm: 'never' empty array with no closing report ----
			// `first !== penultimate` short-circuits closing — for `[]`
			// the closing report is silently dropped even though the
			// closing check would otherwise emit nothing. Verify the
			// empty array under 'never' is clean.
			{Code: `var foo = [];`, Options: []interface{}{"never"}},

			// ---- Dimension 4: nested empty arrays ----
			// Outer's first element is itself an ArrayLit (the inner `[]`),
			// but without an arraysInArrays:true exception the outer keeps
			// its default 'never' behavior. The inner's `spaced && len===0`
			// (here: spaced=false, so no early-return) still produces no
			// reports because there's nothing between the inner brackets.
			{Code: `var arr = [[]];`, Options: []interface{}{"never"}},
			{Code: `var arr = [[], []];`, Options: []interface{}{"never"}},

			// ---- Dimension 4: SpreadElement / RestElement never match
			//      Array/Object exceptions ----
			// SpreadElement is not ArrayLit/ObjectLit even when its operand
			// is, so `objectsInArrays` / `arraysInArrays` exceptions never
			// fire. Mirrors upstream's `isArrayType(spreadElement) === false`.
			{Code: `var arr = [...other];`, Options: []interface{}{"never"}},
			{Code: `var arr = [ ...other ];`, Options: []interface{}{"always"}},
			{Code: `var arr = [...arr1, ...arr2];`, Options: []interface{}{"never", map[string]interface{}{"arraysInArrays": true}}},

			// ---- Dimension 4: BindingElement with default initializer ----
			// `bindingElementBound` unwraps to the BindingElement's name
			// (an Identifier), not to the Initializer expression — so a
			// default of `1` or `[]` doesn't accidentally match the
			// arraysInArrays exception via the wrong slot.
			{Code: `var [a = 1, b = 2] = arr;`, Options: []interface{}{"never"}},
			{Code: `var [ a = 1, b = 2 ] = arr;`, Options: []interface{}{"always"}},
			// Parameter default initializer is itself an ArrayLit, but
			// it's visited as a sibling listener call (not as a child of
			// the binding pattern), so the two arrays are scored
			// independently. Under 'always' the empty `[]` initializer
			// hits the `spaced && len===0` early-return path.
			{Code: `function f([a, b] = []) {}`, Options: []interface{}{"never"}},
			{Code: `function f([ a, b ] = []) {}`, Options: []interface{}{"always"}},

			// ---- Dimension 4: non-Array/non-Object element kinds (no exception) ----
			// new / typeof / unary / optional chain / await — none of
			// these are ArrayLit or ObjectLit, so even with both
			// exceptions enabled the rule falls through to its default
			// spaced setting.
			{Code: `var arr = [new X(), -1, !flag, typeof x, void y];`, Options: []interface{}{"never", map[string]interface{}{"objectsInArrays": true, "arraysInArrays": true}}},
			{Code: `var arr = [a?.b, c?.d()];`, Options: []interface{}{"never"}},
			{Code: `async function f() { return [await one, await two]; }`, Options: []interface{}{"never"}},
			{Code: `function * g() { yield [1, 2]; }`, Options: []interface{}{"never"}},

			// ---- Dimension 4: parser robustness — bracket-like chars
			//      inside string / template / comment payloads ----
			// The parser correctly closes the outer ArrayLit at its own
			// `]`. Our `text[end-1] == ']'` invariant holds because we
			// query `node.End()` from the AST, not via byte heuristics.
			{Code: `var arr = ["a]b", "c[d", "e]f[g"];`, Options: []interface{}{"never"}},
			{Code: "var arr = [`a]b`, `c[d`];", Options: []interface{}{"never"}},
			{Code: `var arr = [1 /* ] [ */, 2];`, Options: []interface{}{"never"}},

			// ---- Adversarial: token suffixes that collide with `*/` ----
			// Reverse byte scan is not token-aware. Without bounding it at
			// `rawLast.End()` it would mistake the trailing bytes of these
			// last elements for a block-comment terminator and over-run
			// into the token body, generating spurious reports + a
			// destructive autofix that deletes the element. Locked in by
			// the `scanLow = rawLast.End()` bound — see PR #983 review.
			//
			// Regex literal forms with various `*/`-adjacent endings:
			{Code: `var arr = [/abc*/];`, Options: []interface{}{"never"}},   // pattern ends in `*/`
			{Code: `var arr = [/\*/];`, Options: []interface{}{"never"}},     // body is escaped star + closing `/`
			{Code: `var arr = [/foo*\//];`, Options: []interface{}{"never"}}, // escaped slash in body
			{Code: `var arr = [a, /xy*/];`, Options: []interface{}{"never"}}, // regex as LAST of multi-element
			{Code: `var arr = [ /abc*/ ];`, Options: []interface{}{"always"}}, // 'always' mode + regex
			// String literal ending with `*/`:
			{Code: `var arr = ["x*/"];`, Options: []interface{}{"never"}},
			{Code: `var arr = [1, "trailing*/"];`, Options: []interface{}{"never"}},
			// Template literal ending with `*/`:
			{Code: "var arr = [`hello*/`];", Options: []interface{}{"never"}},
			{Code: "var arr = [1, `tail*/`];", Options: []interface{}{"never"}},

			// ---- Dimension 4: ArrayLit inside template substitution ----
			// Nested visit through a TemplateExpression. Confirms the
			// listener fires on the inner array even though the byte
			// scan never crosses `}` boundaries.
			{Code: "var t = `pre ${[1, 2]} post`;", Options: []interface{}{"never"}},
			{Code: "var t = `pre ${[ 1, 2 ]} post`;", Options: []interface{}{"always"}},

			// ---- Dimension 4: multi-byte / non-ASCII content in elements ----
			// Unicode strings / identifiers in element position don't
			// shift the bracket positions returned by node.End(). Our
			// byte-level scan only inspects ASCII characters
			// immediately adjacent to `[` and `]`, so multi-byte
			// payloads pass through transparently.
			{Code: `var arr = ["✓", "✗", "中文"];`, Options: []interface{}{"never"}},
			{Code: `var arr = [ "✓", "✗", "中文" ];`, Options: []interface{}{"always"}},
			{Code: `var arr = [π, ε, μ];`, Options: []interface{}{"never"}},

			// ---- Real-user: React hook dependency array ----
			// `useEffect(() => {}, [a, b])` is the single most common
			// shape this rule fires on in real React codebases. Lock
			// both modes.
			{Code: `useEffect(() => {}, [a, b]);`, Options: []interface{}{"never"}},
			{Code: `useEffect(() => {}, [ a, b ]);`, Options: []interface{}{"always"}},
			{Code: `useEffect(() => {}, []);`, Options: []interface{}{"never"}},

			// ---- Real-user: JSX prop value containing an array ----
			// Verifies the rule runs in .tsx files and inside JsxExpression
			// containers. The array lives inside `{...}` so the listener
			// must traverse through JsxAttribute → JsxExpression.
			{Code: `const el = <List items={[1, 2, 3]} />;`, Options: []interface{}{"never"}, Tsx: true},
			{Code: `const el = <Grid rows={[[1, 2], [3, 4]]} />;`, Options: []interface{}{"never"}, Tsx: true},

			// ---- Real-user: array literal as call argument with `as const` ----
			// `[1, 2, 3] as const` is an AsExpression wrapping the
			// ArrayLit; our rule visits the inner ArrayLit only.
			{Code: `var arr = [1, 2, 3] as const;`, Options: []interface{}{"never"}},
			{Code: `var arr = [ 1, 2, 3 ] as const;`, Options: []interface{}{"always"}},

			// ---- Real-user: nested destructuring with rest tail ----
			// Lock the recursive descent into BindingPattern children.
			{Code: `var [[a, ...b], c] = arr;`, Options: []interface{}{"never"}},
			{Code: `var [ [ a, ...b ], c ] = arr;`, Options: []interface{}{"always"}},

			// ---- Real-user: array destructuring in callback parameter ----
			// `.reduce((acc, [k, v]) => …)` is one of the most common
			// destructuring shapes outside variable declarations.
			{Code: `arr.reduce((acc, [k, v]) => acc + v, 0);`, Options: []interface{}{"never"}},
			{Code: `arr.reduce((acc, [ k, v ]) => acc + v, 0);`, Options: []interface{}{"always"}},
			{Code: `arr.forEach(([k, v]) => { console.log(k, v); });`, Options: []interface{}{"never"}},

			// ---- Real-user: ArrayLit inside an ObjectLit property value ----
			// The classic "config object" shape. Verifies the listener
			// reaches arrays nested in PropertyAssignment.Initializer.
			{Code: `const cfg = { items: [1, 2], modes: ["a", "b"] };`, Options: []interface{}{"never"}},

			// ---- Real-user: trailing comma in destructuring + 'never' ----
			// Sparse-trailing form. `[a, b, ,]` is the upstream `[ ,x, ]`
			// variant flipped to the trailing edge.
			{Code: `var [a, b, ,] = arr;`, Options: []interface{}{"never"}},

			// ---- Real-user: TS decorator with array argument ----
			// `@Decorator([1, 2])` — array literal in a decorator's
			// argument list. Confirms the rule runs inside Decorator
			// nodes.
			{Code: `@Decorator([1, 2]) class X {}`, Options: []interface{}{"never"}},
			{Code: `@Decorator([ 1, 2 ]) class X {}`, Options: []interface{}{"always"}},

			// ---- Options resilience: empty options array → 'never' defaults ----
			// rslint's CLI can dispatch options as `[]` (level only); we
			// fall through to spaced=false with no exceptions.
			{Code: `var arr = [1, 2];`, Options: []interface{}{}},

			// ---- Options resilience: unknown keys in the second arg are ignored ----
			// Defensive: rslint does not perform JSON-schema validation, so
			// we must silently ignore keys upstream's schema would reject.
			{Code: `var arr = [1, 2];`, Options: []interface{}{"never", map[string]interface{}{"unknownKey": true}}},

			// ---- Options resilience: non-bool exception values are ignored ----
			// `singleValue: "true"` (string) must NOT flip the exception
			// — only literal `true` / `false` matter (mirrors upstream's
			// `=== !spaced` reference-equality check).
			{Code: `var arr = [1];`, Options: []interface{}{"never", map[string]interface{}{"singleValue": "true"}}},
			{Code: `var arr = [1];`, Options: []interface{}{"never", map[string]interface{}{"singleValue": 1}}},

			// ---- Unicode WhiteSpace + LineTerminator (ECMAScript §12.2/§12.3) ----
			// Parity with @stylistic/eslint-plugin's array-bracket-spacing
			// (verified via local ESLint probe). NBSP / IDEO / ZWNBSP count
			// as space → satisfy `always` mode; LS / PS count as line
			// terminator → cross-line short-circuit (both modes valid).
			// `always` with NBSP / IDEO / BOM inside brackets — valid.
			{Code: "[\u00A0foo\u00A0]", Options: []interface{}{"always"}},
			{Code: "[\u3000foo\u3000]", Options: []interface{}{"always"}},
			{Code: "[\uFEFFfoo\uFEFF]", Options: []interface{}{"always"}},
			// `never` + LS / PS one-side → that side cross-line; opposite
			// side hugs the bracket so both sides valid.
			{Code: "[foo\u2028]", Options: []interface{}{"never"}},
			{Code: "[\u2028foo]", Options: []interface{}{"never"}},
			{Code: "[foo\u2029]", Options: []interface{}{"never"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Locks in upstream validateArraySpacing() arm:
			// objectsInArrays + paren-wrapped element ----
			// Without SkipParentheses on elements[0] the leading-bracket
			// check would NOT flip the spaced default, and this would be
			// reported as a regular 'never' violation. With SkipParentheses
			// the exception fires and the opening space becomes required
			// (closing space becomes required too because elements[len-1]
			// is also paren-wrapped). Fix output and positions verify the
			// exception is applied symmetrically.
			{
				Code:    `var foo = [({a: 1}), 2, ({b: 2})];`,
				Output:  []string{`var foo = [ ({a: 1}), 2, ({b: 2}) ];`},
				Options: []interface{}{"never", map[string]interface{}{"objectsInArrays": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 33, EndLine: 1, EndColumn: 34},
				},
			},

			// ---- Locks in upstream validateArraySpacing() arm:
			// arraysInArrays + paren-wrapped first element ----
			{
				Code:    `var arr = [([1, 2]), 2];`,
				Output:  []string{`var arr = [ ([1, 2]), 2];`},
				Options: []interface{}{"never", map[string]interface{}{"arraysInArrays": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				},
			},

			// ---- Dimension 4: nested ArrayLiteral in 'always' — verifies
			// enter/exit emit order matches source order across the layer
			// boundary. Without the enter-on-open / exit-on-close split,
			// the outer closing report would emit before the inner reports
			// (parent-first traversal). The expected column sequence
			// 11 → 18 → 23 → 24 exercises this.
			{
				Code:    `var arr = [1, 2, [3, 4]];`,
				Output:  []string{`var arr = [ 1, 2, [ 3, 4 ] ];`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Locks in upstream validateArraySpacing() arm:
			// singleElementException with 'never' ----
			// length === 1 triggers the exception arm that flips the
			// opening-must-be-spaced AND closing-must-be-spaced from false
			// to true, regardless of which other exceptions match. Verify
			// both reports fire on a paren-wrapped single-element value
			// (so neither object- nor array-in-arrays exceptions apply).
			{
				Code:    `var arr = [(x)];`,
				Output:  []string{`var arr = [ (x) ];`},
				Options: []interface{}{"never", map[string]interface{}{"singleValue": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},

			// ---- Dimension 4: TS array binding pattern with type annotation, 'always' ----
			// Confirms node.End() of the binding pattern stops at `]` and
			// the type annotation `: [number, number]` on the parent
			// declaration doesn't trip up our open/close position
			// resolution. The annotation itself is a TupleType node and
			// is not visited by this rule (no second-pass fix).
			{
				Code:    `var [a, b]: [number, number] = [1, 2];`,
				Output:  []string{`var [ a, b ]: [number, number] = [ 1, 2 ];`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 32, EndLine: 1, EndColumn: 33},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 37, EndLine: 1, EndColumn: 38},
				},
			},

			// ---- Real-user: deeply nested array literal flagged at every layer ----
			// Triple-level nesting in 'never' mode with arraysInArrays:true
			// — every layer whose first/last element is itself an array
			// has the exception fired independently. Confirms the
			// per-node exception derivation doesn't bleed across nesting
			// levels (e.g. the innermost `[2]` keeps default 'never'
			// because its element `2` isn't an array, while its parent
			// `[[2]]` flips to spaced). Also exercises the enter/exit
			// emit ordering across three layers: 11 → 17 → 21 → 22.
			{
				Code:    `var arr = [[1], [[2]]];`,
				Output:  []string{`var arr = [ [1], [ [2] ] ];`},
				Options: []interface{}{"never", map[string]interface{}{"arraysInArrays": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},

			// ---- Locks in upstream babel block: arrow param destructuring
			// with TS type annotation, 'never' ----
			// Mirrors upstream's babel/Flow test `([ a, b ]: Array<any>) => {}`
			// rewritten in tsgo-friendly TS syntax. Verifies node.End() of
			// the ArrayBindingPattern stops at `]` even when the parent
			// parameter carries a type annotation, so the closing report's
			// reported range stays on `]` and the trailing `:` doesn't
			// leak into the fix.
			{
				Code:    `([ a, b ]: Array<any>) => {}`,
				Output:  []string{`([a, b]: Array<any>) => {}`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 3, EndLine: 1, EndColumn: 4},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    `([a, b]: Array<any>) => {}`,
				Output:  []string{`([ a, b ]: Array<any>) => {}`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
				},
			},

			// ---- Real-user: array as call argument, multiple spaces ----
			// Mirrors the 'multiple spaces' upstream cases but inside a
			// function call rather than a variable initializer. Confirms
			// that the rule's text-based scan locates the array's own
			// brackets even when the surrounding context has its own
			// whitespace.
			{
				Code:    `fn([   1, 2   ]);`,
				Output:  []string{`fn([1, 2]);`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 8},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 12, EndLine: 1, EndColumn: 15},
				},
			},

			// ---- Dimension 4: nested empty array under 'always' ----
			// `[[]]` — outer needs spaces (default 'always'), inner empty
			// hits early-return. Confirms the two layers don't interfere.
			{
				Code:    `var arr = [[]];`,
				Output:  []string{`var arr = [ [] ];`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},

			// ---- Dimension 4: SpreadElement does NOT match
			//      arraysInArrays — only the outer 'never' rule fires ----
			{
				Code:    `var arr = [ ...other ];`,
				Output:  []string{`var arr = [...other];`},
				Options: []interface{}{"never", map[string]interface{}{"arraysInArrays": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
				},
			},

			// ---- Dimension 4: paren-wrapped array element triggers
			//      arraysInArrays on both ends ----
			// Symmetric companion to the [({a:1}), ...] objectsInArrays
			// invalid case. Confirms SkipParentheses → ArrayLit
			// classification works for first AND last positions.
			{
				Code:    `var arr = [([1]), ([2])];`,
				Output:  []string{`var arr = [ ([1]), ([2]) ];`},
				Options: []interface{}{"never", map[string]interface{}{"arraysInArrays": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Dimension 4: BindingElement with default initializer
			//      under 'always' ----
			// The defaults (`= 1`, `= 2`) are part of BindingElement, not
			// elements at array level — they don't affect bracket
			// classification.
			{
				Code:    `var [a = 1, b = 2] = arr;`,
				Output:  []string{`var [ a = 1, b = 2 ] = arr;`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},

			// ---- Dimension 4: parameter default array initializer is
			//      visited as a sibling, not a child ----
			// Verifies the rule fires on the outer destructuring AND the
			// initializer's empty array independently — the empty `[]`
			// initializer hits early-return so only the destructuring
			// brackets get reported.
			{
				Code:    `function f([a, b] = []) {}`,
				Output:  []string{`function f([ a, b ] = []) {}`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},

			// ---- Branch lock-in: three-level always — single-pass fix
			//      across six independent insert points ----
			// Verifies all six fixes (3 opening + 3 closing, each at a
			// distinct byte position) compose in one ApplyRuleFixes call
			// without overlap. Also exercises the enter/exit emit order
			// across three layers in a row.
			{
				Code:    `var arr = [[[1]]];`,
				Output:  []string{`var arr = [ [ [ 1 ] ] ];`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},

			// ---- Locks in upstream destructive-fix behavior:
			//      trailing block comment removed with the offending
			//      space ----
			// Upstream's `removeRange([penultimate.range[1], last.range[0]])`
			// removes the comment between the last token and `]`. Our
			// reverse-scan delivers the same byte range, so the fix
			// matches upstream byte-for-byte.
			{
				Code:    `var arr = [1, 2 /* trailing */ ];`,
				Output:  []string{`var arr = [1, 2];`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 16, EndLine: 1, EndColumn: 32},
				},
			},

			// ---- Locks in upstream destructive-fix behavior:
			//      leading block comment removed too ----
			// Symmetric lock for the leading edge — `scanner.SkipTrivia`
			// crosses the block comment, the fix removes everything
			// between `[` and the first real token.
			{
				Code:    `var arr = [ /* leading */ 1, 2];`,
				Output:  []string{`var arr = [1, 2];`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 27},
				},
			},

			// ---- Real-user: React hook deps with 'always' ----
			{
				Code:    `useEffect(() => {}, [a, b]);`,
				Output:  []string{`useEffect(() => {}, [ a, b ]);`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
				},
			},

			// ---- Real-user: JSX prop array under 'always' ----
			// Confirms the rule visits ArrayLit inside JsxExpression and
			// the .tsx file path is exercised.
			{
				Code:    `const el = <List items={[1, 2, 3]} />;`,
				Output:  []string{`const el = <List items={[ 1, 2, 3 ]} />;`},
				Options: []interface{}{"always"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 33, EndLine: 1, EndColumn: 34},
				},
			},

			// ---- Real-user: TS decorator argument needs spaces under
			//      'always' ----
			{
				Code:    `@Decorator([1, 2]) class X {}`,
				Output:  []string{`@Decorator([ 1, 2 ]) class X {}`},
				Options: []interface{}{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},

			// ---- Real-user: callback destructuring with default
			//      'never' violation ----
			// `.reduce((acc, [ k, v ]) => …)` is a common shape that
			// gets accidentally formatted with spaces — verify the
			// rule normalizes it.
			{
				Code:    `arr.reduce((acc, [ k, v ]) => acc + v, 0);`,
				Output:  []string{`arr.reduce((acc, [k, v]) => acc + v, 0);`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Real-user: array literal in template substitution ----
			{
				Code:    "var t = `${[ 1, 2 ]}`;",
				Output:  []string{"var t = `${[1, 2]}`;"},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},

			// ---- Multi-byte safety: UTF-8 string content + 'never' violation ----
			// The byte-level scan reads ASCII at the bracket boundary;
			// the multi-byte payload between is opaque. Position
			// reporting must use UTF-16 char offsets (via
			// scanner.GetECMALineAndUTF16CharacterOfPosition).
			{
				Code:    `var arr = [ "中" ];`,
				Output:  []string{`var arr = ["中"];`},
				Options: []interface{}{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},

			// ---- Real-user: deeply nested config-style array ----
			// `[{ items: [1, 2] }, { items: [3, 4] }]` — common
			// fixture-style data shape. Lock with 'always' +
			// objectsInArrays:false so only the inner ArrayLits get
			// spaced, not the wrapper objects.
			{
				Code:    `var data = [{ items: [1, 2] }, { items: [3, 4] }];`,
				Output:  []string{`var data = [{ items: [ 1, 2 ] }, { items: [ 3, 4 ] }];`},
				Options: []interface{}{"always", map[string]interface{}{"objectsInArrays": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 27, EndLine: 1, EndColumn: 28},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 41, EndLine: 1, EndColumn: 42},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 46, EndLine: 1, EndColumn: 47},
				},
			},
		},
	)
}
