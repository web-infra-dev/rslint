package prefer_object_spread

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferObjectSpreadExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// Dimension 3 (autofix boundaries around side-effecting arguments) is N/A for
// this rule: upstream's fixer has no side-effect bail-out — it always
// autofixes every reported Object.assign call, so there is nothing to lock in
// beyond the comment-preservation cases already migrated in the upstream
// file. Declaration/container-form rows (class/function shape) are also N/A:
// this rule only ever matches a CallExpression callee, never a declaration.
func TestPreferObjectSpreadExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferObjectSpreadRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: numeric-literal key does not match "assign" ----
			{Code: `Object[0]({}, foo)`},

			// ---- Branch lock-in: "Object" shadowed by a function parameter ----
			{Code: `function f(Object) { return Object.assign({}, a); }`},

			// ---- Branch lock-in: "Object" declared off via languageOptions.globals ----
			{Code: `Object.assign({}, foo)`, Globals: map[string]bool{"Object": false}},

			// ---- Branch lock-in: "globalThis" declared off via languageOptions.globals ----
			{Code: `globalThis.Object.assign({}, foo)`, Globals: map[string]bool{"globalThis": false}},

			// ---- Branch lock-in: "window" declared off via languageOptions.globals ----
			{Code: `window.Object.assign({}, foo)`, Globals: map[string]bool{"window": false}},

			// ---- Branch lock-in: "window" shadowed by a function parameter ----
			{Code: `function f(window) { return window.Object.assign({}, a); }`},

			// ---- Member-name evaluation: a computed member name that cannot
			// be statically folded must not match ----
			{Code: `Object["as" + sign]({}, foo)`},

			// ---- Alias tracking: a computed destructured property name that
			// cannot be statically folded must not match ----
			{Code: `const { [key]: a } = Object; a({}, foo)`},

			// ---- Modified-global tracking: a bare write to the global
			// `Object` untracks every reference positioned after it (calls
			// before the write still match — see the invalid section) ----
			{Code: `Object = {}; Object.assign({}, foo);`},
			{Code: `Object ||= {}; Object.assign({}, foo);`},

			// ---- Modified-global tracking: same for a written global-object
			// entry name ----
			{Code: `window = {}; window.Object.assign({}, foo);`},

			// ---- Modified-global tracking: an alias whose initializer reads
			// the global after it was modified must not match (capture-position
			// semantics; contrast with the capture-before-write invalid case) ----
			{Code: `Object = {}; const o = Object; o.assign({}, foo);`},

			// ---- Modified-global tracking: for-of assignment targets and
			// destructuring-assignment targets count as writes too ----
			{Code: `for (Object of xs); Object.assign({}, foo)`},
			{Code: `({ Object } = x); Object.assign({}, foo)`},

			// ---- Global-object tracking: an alias of something that is not a
			// global object must not match ----
			{Code: `const g = foo; g.Object.assign({}, x)`},

			// ---- Global-object tracking: destructuring `Object` off a
			// non-global source must not match ----
			{Code: `const { Object: O } = foo; O.assign({}, x)`},

			// ---- Branch lock-in: nested Object.assign whose own "Object" is
			// shadowed by an inner function parameter — the outer call still
			// matches, the inner one must not (scope boundary correctness) ----
			{Code: `(function (Object) { return Object.assign({}, x); })()`},

			// ---- Alias tracking: the last write before the call assigns a
			// non-Object value, so the alias no longer matches at that point ----
			{Code: `let o = Object; o = foo; o.assign({}, bar)`},

			// ---- Alias tracking: same for a reassigned Object.assign alias ----
			{Code: `let assign = Object.assign; assign = foo; assign({}, bar)`},

			// ---- Flow tracking: a local write positioned after the call does
			// not make the alias match at the call site ----
			{Code: `let o; o.assign({}, x); o = Object;`},

			// ---- Flow tracking: compound assignment is an opaque write ----
			{Code: `let o = foo; o = Object; o ||= bar; o.assign({}, x)`},

			// ---- Flow tracking: destructuring-assignment writes are opaque ----
			{Code: `let assign; ({ assign } = Object); assign({}, x)`},

			// ---- Cycle safety: mutually- and self-referential alias chains
			// must resolve to unknown instead of recursing forever ----
			{Code: `var a = b; var b = a; b.assign({}, x)`},
			{Code: `var a = a; a.assign({}, x)`},

			// ---- Nested destructuring: a default value makes the bound value
			// untrackable ----
			{Code: `const { Object: { assign } = {} } = globalThis; assign({}, x)`},

			// ---- Nested destructuring: a non-global root must not match ----
			{Code: `const { Object: { assign } } = foo; assign({}, x)`},

			// ---- Alias tracking: destructuring a property other than
			// "assign" off Object must not match ----
			{Code: `const {keys} = Object; keys({}, bar)`},

			// ---- Alias tracking: rest element in a destructure off Object
			// must not match (DotDotDotToken guard) ----
			{Code: `const {...rest} = Object; rest.assign({}, bar)`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver, single and multi-level ----
			{
				Code:   `(Object).assign({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `((Object)).assign({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: TS non-null assertion on the receiver ----
			{
				Code:   `Object!.assign({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: TS type-expression wrapper on the receiver ----
			{
				Code:   `(Object as any).assign({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: optional chain forms ----
			{
				Code:   `Object?.assign({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `Object.assign?.({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Dimension 4 / element-access key forms: bracket string and
			// static-template-literal property name for "assign" ----
			{
				Code:   `Object['assign']({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   "Object[`assign`]({}, foo)",
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Branch lock-in: globalThis.Object.assign reached through a
			// bracket access on the "Object" hop ----
			{
				Code:   `globalThis['Object'].assign({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Branch lock-in: needsWrappingParens default arm for a
			// non-assignment BinaryExpression operator other than "+"/comma
			// (already covered upstream) — nullish-coalescing here ----
			{
				Code:   `let a = foo ?? Object.assign({}, bar)`,
				Output: []string{`let a = foo ?? ({ ...bar})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 16}},
			},

			// ---- Branch lock-in: needsSpreadParens must NOT double-wrap an
			// argument that is already parenthesized ----
			{
				Code:   `Object.assign({}, (a ? b : c))`,
				Output: []string{`({ ...(a ? b : c)})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Branch lock-in: object argument's own trailing comma
			// (properties.length > 0 disjunct of the comma-dedup check, as
			// opposed to the properties.length === 0 disjunct already covered
			// upstream by the empty-object cases) ----
			{
				Code:   `Object.assign({ a: 1, }, b)`,
				Output: []string{`({a: 1, ...b})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: string-literal object-argument property key
			// (not just computed / identifier / numeric, already covered
			// upstream) ----
			{
				Code:   `Object.assign({}, { "foo-bar": 1 })`,
				Output: []string{`({ "foo-bar": 1})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Real-user: common props-merge pattern (default props
			// merged with an instance's own props) ----
			{
				Code:   `const props = Object.assign({}, defaultProps, this.props);`,
				Output: []string{`const props = { ...defaultProps, ...this.props};`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 15}},
			},

			// ---- Real-user: Object.assign nested as a call argument (legacy
			// React setState pattern), exercising the "parent is a
			// CallExpression" no-extra-parens branch with real member
			// expressions rather than bare identifiers ----
			{
				Code:   `this.setState(Object.assign({}, this.state, { loading: true }));`,
				Output: []string{`this.setState({ ...this.state, loading: true});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 15}},
			},

			// ---- Branch lock-in: a multi-byte UTF-8 property name butts
			// directly against the object literal's closing brace — the
			// boundary walk must decode whole runes, not individual UTF-8
			// bytes, or it corrupts the fix into invalid UTF-8 ----
			{
				Code:   "Object.assign({}, {à})",
				Output: []string{"({ à})"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Branch lock-in: a comment sits between the last property
			// and the object argument's own trailing comma — the
			// trailing-comma probe must skip comments (like the
			// separator-comma probe already does), not just whitespace, or
			// it leaves a stray double comma in the fix ----
			{
				Code:   "Object.assign({a: 1 /* x */,}, {b: 2})",
				Output: []string{"({a: 1 /* x */, b: 2})"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Branch lock-in: TS assertion on the whole `Object.assign`
			// callee (not just the `Object` receiver, already covered above)
			// ----
			{
				Code:   `(Object.assign as any)({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Branch lock-in: needsWrappingParens for a
			// ComputedPropertyName parent — tsgo wraps a computed key's
			// expression in its own node, unlike a normal property value, so
			// it needs its own no-wrap case alongside PropertyAssignment ----
			{
				Code:   `const o = {[Object.assign({}, a)]: 1};`,
				Output: []string{`const o = {[{ ...a}]: 1};`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 13}},
			},

			// ---- Alias tracking: `Object` receiver reached through a stable
			// `let` alias (ESLint's ReferenceTracker follows this; mirrors
			// isObjectReference's evaluator.ResolveIdentifierInitializer
			// path) ----
			{
				Code:   `let o = Object; o.assign({}, foo)`,
				Output: []string{`let o = Object; ({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 17}},
			},

			// ---- Alias tracking: `Object.assign` itself reached through a
			// stable `const` alias, including across a nested function scope
			// ----
			{
				Code:   `const assign = Object.assign; assign({}, foo)`,
				Output: []string{`const assign = Object.assign; ({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 31}},
			},
			{
				Code:   `function g() { const assign = Object.assign; return assign({}, foo); }`,
				Output: []string{`function g() { const assign = Object.assign; return { ...foo}; }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 53}},
			},

			// ---- Alias tracking: `assign` destructured off `Object`, both
			// shorthand and renamed, and regardless of const/let/var ----
			{
				Code:   `const {assign} = Object; assign({}, foo)`,
				Output: []string{`const {assign} = Object; ({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 26}},
			},
			{
				Code:   `var {assign: a} = Object; a({}, foo)`,
				Output: []string{`var {assign: a} = Object; ({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 27}},
			},

			// ---- Alias tracking: trailing comma-operator sequence around
			// the Object.assign reference (`unwrapValue`'s comma-descent),
			// including a multi-comma chain ----
			{
				Code:   `(0, Object.assign)({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `(1, 2, Object.assign)({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `(0, Object).assign({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Member-name evaluation: computed member name folded by the
			// static string evaluator (`"as" + "sign"`), like ESLint's
			// getStaticValue-based ReferenceTracker ----
			{
				Code:   `Object["as" + "sign"]({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Alias tracking: `Object` receiver reached through a
			// multi-hop alias chain (P -> O -> Object) ----
			{
				Code:   `const O = Object; const P = O; P.assign({}, foo)`,
				Output: []string{`const O = Object; const P = O; ({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 32}},
			},

			// ---- Alias tracking: `assign` destructured off `Object` via a
			// string-literal and a statically-folded computed property name ----
			{
				Code:   `const { "assign": a } = Object; a({}, foo)`,
				Output: []string{`const { "assign": a } = Object; ({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 33}},
			},
			{
				Code:   `const { ["as" + "sign"]: b } = Object; b({}, foo)`,
				Output: []string{`const { ["as" + "sign"]: b } = Object; ({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 40}},
			},

			// ---- Global-object entries: window / self / global alongside the
			// globalThis case already covered above ----
			{
				Code:   `window.Object.assign({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `self.Object.assign({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `global.Object.assign({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Modified-global tracking: a write to a local binding that
			// shadows `Object` must NOT disable tracking of the real global ----
			{
				Code:   "function f(Object) { Object = {}; }\nObject.assign({}, foo)",
				Output: []string{"function f(Object) { Object = {}; }\n({ ...foo})"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 2, Column: 1}},
			},

			// ---- Global-object tracking: the global object itself reached
			// through a stable local alias ----
			{
				Code:   `const g = globalThis; g.Object.assign({}, foo)`,
				Output: []string{`const g = globalThis; ({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 23}},
			},

			// ---- Global-object tracking: `Object` destructured off a global
			// object, renamed and shorthand (the shorthand also shadows the
			// global `Object` name, so it must be tracked via the destructure
			// path, not isGlobalIdentifier) ----
			{
				Code:   `const { Object: O } = globalThis; O.assign({}, foo)`,
				Output: []string{`const { Object: O } = globalThis; ({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 35}},
			},
			{
				Code:   `const { Object } = globalThis; Object.assign({}, foo)`,
				Output: []string{`const { Object } = globalThis; ({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 32}},
			},

			// ---- __proto__ preservation: a source object literal with a
			// prototype-setting `__proto__:` property must be kept whole
			// behind a spread — unwrapping it into the merged literal would
			// set the result's prototype, which Object.assign never does for
			// a source (its `__proto__:` creates no own properties, so the
			// call copies nothing) ----
			{
				Code:   `Object.assign({}, { __proto__: proto })`,
				Output: []string{`({ ...{ __proto__: proto }})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `Object.assign({}, { "__proto__": proto })`,
				Output: []string{`({ ...{ "__proto__": proto }})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `Object.assign({}, { __proto__: a }, { b: 1 })`,
				Output: []string{`({ ...{ __proto__: a }, b: 1})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- __proto__ preservation: the FIRST argument keeps
			// unwrapping — its `__proto__:` sets the target's prototype in
			// the original call too, so the merged literal is equivalent ----
			{
				Code:   `Object.assign({ __proto__: base }, foo)`,
				Output: []string{`({__proto__: base, ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- __proto__ preservation: a computed `["__proto__"]` creates
			// an ordinary own property, not a prototype, so the literal still
			// unwraps ----
			{
				Code:   `Object.assign({}, { ["__proto__"]: proto })`,
				Output: []string{`({ ["__proto__"]: proto})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- needsWrappingParens: as the callee of an outer call the
			// fixed literal must be parenthesized (`{ ...foo }()` does not
			// parse), unlike the argument position covered above ----
			{
				Code:   `Object.assign({}, foo)()`,
				Output: []string{`({ ...foo})()`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Modified-global tracking is flow-sensitive: a call before
			// the bare write still matches; the call after it does not ----
			{
				Code:   `Object.assign({}, before); Object = {}; Object.assign({}, after);`,
				Output: []string{`({ ...before}); Object = {}; Object.assign({}, after);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Modified-global tracking: an alias whose initializer read
			// the global BEFORE the write captured the pristine Object, so
			// the call still matches at runtime ----
			{
				Code:   `const o = Object; Object = {}; o.assign({}, foo);`,
				Output: []string{`const o = Object; Object = {}; ({ ...foo});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 32}},
			},

			// ---- Flow tracking: a reassigned alias reports the calls made
			// while it held Object and only those ----
			{
				Code:   `let o = Object; o.assign({}, a); o = foo; o.assign({}, b);`,
				Output: []string{`let o = Object; ({ ...a}); o = foo; o.assign({}, b);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 17}},
			},

			// ---- Flow tracking: aliases established by a plain assignment
			// (not a declaration initializer) match too, for both the Object
			// receiver and Object.assign itself ----
			{
				Code:   `let o; o = Object; o.assign({}, x)`,
				Output: []string{`let o; o = Object; ({ ...x})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 20}},
			},
			{
				Code:   `let assign; assign = Object.assign; assign({}, x)`,
				Output: []string{`let assign; assign = Object.assign; ({ ...x})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 37}},
			},

			// ---- Nested destructuring: `assign` bound through a nested
			// pattern off a global object ----
			{
				Code:   `const { Object: { assign } } = globalThis; assign({}, x)`,
				Output: []string{`const { Object: { assign } } = globalThis; ({ ...x})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 44}},
			},

			// ---- Alias chains longer than the old 8-hop cap still resolve
			// (cycle detection replaced the fixed depth limit) ----
			{
				Code:   `const a = Object; const b = a; const c = b; const d = c; const e = d; const f = e; const g = f; const h = g; const i = h; const j = i; j.assign({}, foo)`,
				Output: []string{`const a = Object; const b = a; const c = b; const d = c; const e = d; const f = e; const g = f; const h = g; const i = h; const j = i; ({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 136}},
			},
		},
	)
}
