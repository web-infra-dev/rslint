// cspell:ignore datepicker fakeDOM foobar

package aria_role

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// invalidRoleError is the canonical invalid-role diagnostic used across this
// file's invalid suite — shorthand alias.
func invalidRoleError() rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: "invalidAriaRole", Message: errorMessage}
}

// TestAriaRoleExtras covers cases NOT in upstream's
// `__tests__/src/rules/aria-role-test.js` — universal edge shapes
// (Dimension 4 of the port skill), TS-only syntax forms, JSX shape variants,
// option-shape coverage, and lock-in tests for each branch in the upstream
// `JSXAttribute` listener.
func TestAriaRoleExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AriaRoleRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Parenthesized values — tsgo preserves; upstream Literal
			// extractor uses ESTree paren-flattening. LiteralPropAriaValue
			// strips parens internally so single / nested parens both
			// resolve to the string "button".
			// ============================================================
			{Code: `<div role={("button")} />`, Tsx: true},
			{Code: `<div role={(("button"))} />`, Tsx: true},

			// ============================================================
			// TS wrapper kinds — LITERAL_TYPES upstream:
			//   TSAsExpression:        noop → null → skip
			//   TSNonNullExpression:   noop → null → skip
			//   TSSatisfiesExpression: not in LITERAL_TYPES → noop default → skip
			// LiteralPropAriaValue mirrors: returns NoLit. Step-3 skip.
			//
			// These look like valid roles to a human reader, but the rule
			// can't statically verify them through the TS wrapper, so it
			// declines to flag — same as upstream.
			// ============================================================
			{Code: `<div role={"button" as any} />`, Tsx: true},
			{Code: `<div role={"button" as const} />`, Tsx: true},
			{Code: `<div role={"button"!} />`, Tsx: true},
			{Code: `<div role={"button" satisfies string} />`, Tsx: true},

			// ============================================================
			// Locks upstream listener gates 1–3:
			//  - Identifier `undefined` → AriaLiteralUndef → step-3 skip.
			//  - MemberExpression → LITERAL_TYPES noop → AriaLiteralNoLit → skip.
			//  - CallExpression → noop → skip.
			//  - ConditionalExpression → noop → skip.
			//  - LogicalExpression (||, &&, ??) → noop → skip.
			//  - BinaryExpression (`+`) → noop → skip.
			//  - JSXElement → noop → skip.
			// Each branch upstream has its own `LITERAL_TYPES[Kind] = noop`
			// entry. Lock all branches so a future refactor that adds
			// recursive extraction (or accidentally promotes one of these
			// to extractable) is caught here.
			// ============================================================
			{Code: `<div role={undefined} />`, Tsx: true},
			{Code: `<div role={obj.role} />`, Tsx: true},
			{Code: `<div role={getRole()} />`, Tsx: true},
			{Code: `<div role={cond ? "button" : "link"} />`, Tsx: true},
			{Code: `<div role={a && "button"} />`, Tsx: true},
			{Code: `<div role={a ?? "button"} />`, Tsx: true},
			{Code: `<div role={"button" + ""} />`, Tsx: true},
			{Code: `<div role={<span />} />`, Tsx: true},

			// ============================================================
			// Template literals — no-substitution preserves the verbatim
			// string and skips the case-insensitive "true"/"false"
			// boolean coerce; substitution synthesizes a placeholder
			// string from quasis + extracted parts.
			// ============================================================
			{Code: "<div role={`button`} />", Tsx: true},
			{Code: "<div role={`doc-abstract`} />", Tsx: true},

			// ============================================================
			// JSX shape variants — both open + closing tag forms must
			// fire the listener.
			// ============================================================
			{Code: `<div role="button">x</div>`, Tsx: true},

			// ============================================================
			// SpreadAttribute — listener is JsxAttribute only, so spread
			// attributes never fire the rule even when their evaluated
			// shape would carry a role.
			// ============================================================
			{Code: `<div {...spread} />`, Tsx: true},
			{Code: `<div role="button" {...spread} />`, Tsx: true},

			// ============================================================
			// Name case — `propName(attr).toUpperCase() === 'ROLE'` matches
			// mixed-case attribute names. But these are NOT real ARIA
			// attributes at the DOM level — JSX prop names are case-
			// sensitive in React, and only lowercase `role` ends up as the
			// HTML `role` attribute. Upstream nonetheless validates them
			// because the rule never re-asserts the lowercase form.
			//
			// Valid because "button" is in the role set:
			// ============================================================
			{Code: `<div Role="button" />`, Tsx: true},
			{Code: `<div ROLE="button" />`, Tsx: true},

			// ============================================================
			// ignoreNonDOM + native HTML element with a valid role —
			// rule runs, validates "button", passes.
			// ============================================================
			{Code: `<button role="button" />`, Tsx: true, Options: ignoreNonDOMOption},

			// ============================================================
			// Combined options (ignoreNonDOM + allowedInvalidRoles): DOM
			// element + invalid token that IS allow-listed → valid.
			// ============================================================
			{
				Code: `<div role="invalid-role tabpanel" />`, Tsx: true,
				Options: map[string]interface{}{
					"ignoreNonDOM":        true,
					"allowedInvalidRoles": []interface{}{"invalid-role"},
				},
			},

			// ============================================================
			// Option-shape coverage (Phase 4 contract): the rule must work
			// with both the bare-map (CLI / single-option) form AND the
			// array-wrapped (multi-element rule_tester) form.
			// ============================================================
			{
				Code:    `<img role="invalid-role" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"allowedInvalidRoles": []interface{}{"invalid-role"}}},
			},

			// ============================================================
			// Empty options object — defaults: ignoreNonDOM=false,
			// allowedInvalidRoles=[]. Valid role still passes.
			// ============================================================
			{Code: `<div role="button" />`, Tsx: true, Options: map[string]interface{}{}},

			// ============================================================
			// Non-string entries in allowedInvalidRoles are silently
			// skipped — defensive but mirrors upstream's `new Set(opts...)`
			// which would include the raw value but `validRoles.has(val)`
			// type-mismatches on a string-vs-number compare.
			// ============================================================
			{
				Code: `<div role="button" />`, Tsx: true,
				Options: map[string]interface{}{"allowedInvalidRoles": []interface{}{42, "foo"}},
			},

			// ============================================================
			// polymorphicAllowList: only `<Box>` consults polymorphic
			// prop; `<Foo>` does not. Both with ignoreNonDOM and a custom
			// role: Foo is not DOM and not on the allow-list, so polymorphic
			// is skipped; Foo remains "Foo" → not DOM → skip.
			// ============================================================
			{
				Code: `<Foo as="div" role="invalid-role" />`, Tsx: true,
				Options: ignoreNonDOMOption,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"polymorphicPropName":  "as",
						"polymorphicAllowList": []interface{}{"Box"},
					},
				},
			},

			// ============================================================
			// Nesting / traversal boundaries — listener is per-JsxAttribute,
			// no scope concept, so the rule fires independently on each
			// element. Inner / outer / Fragment / mapped lists / render
			// callbacks all keep their attribute-level verdicts independent.
			// ============================================================
			{Code: `<div><span role="button" /></div>`, Tsx: true},
			{Code: `<><div role="button" /></>`, Tsx: true},
			{Code: `<Fragment><div role="link" /></Fragment>`, Tsx: true},
			{Code: `items.map(x => <li key={x.id} role="listitem" />)`, Tsx: true},
			{Code: `<Comp render={() => <div role="button" />} />`, Tsx: true},
			// Member-access tag (`<Module.Comp>`) and JSX namespaced name
			// (`<svg:circle>`) — GetElementType resolves them without panic.
			{Code: `<Module.Comp role="button" />`, Tsx: true},
			{Code: `<svg><circle role="img" /></svg>`, Tsx: true},

			// ============================================================
			// ArrayLiteralExpression valid cases — upstream's LITERAL_TYPES
			// evaluates each element and joins with `,`. `String(['button'])`
			// is `"button"` → split by space → `["button"]` → valid.
			// Null / undefined elements are filtered before stringification.
			// ============================================================
			{Code: `<div role={["button"]} />`, Tsx: true},
			{Code: `<div role={["doc-abstract"]} />`, Tsx: true},
			{Code: `<div role={[null, "button"]} />`, Tsx: true},
			{Code: `<div role={["button", undefined]} />`, Tsx: true},

			// ============================================================
			// allowedInvalidRoles option — edge values.
			// ============================================================
			// Empty-string allow: `<div role="" />` becomes valid because
			// `""` matches the allow-set. Tests the contract that the
			// allow-list short-circuits BEFORE the canonical role check.
			{
				Code: `<div role="" />`, Tsx: true,
				Options: map[string]interface{}{"allowedInvalidRoles": []interface{}{""}},
			},
			// Overlap with canonical list — listing a real role in the
			// allow-set is a no-op (it was already valid). Locks
			// short-circuit-first behavior.
			{
				Code: `<div role="button" />`, Tsx: true,
				Options: map[string]interface{}{"allowedInvalidRoles": []interface{}{"button"}},
			},
			// Allow-set is case-sensitive, matching `Set.has` semantics —
			// "Button" in the allow-set ONLY allows the literal string
			// "Button", not "button".
			{
				Code: `<div role="Button" />`, Tsx: true,
				Options: map[string]interface{}{"allowedInvalidRoles": []interface{}{"Button"}},
			},
			// Unknown option keys are silently ignored (forward-compat).
			{
				Code: `<div role="button" />`, Tsx: true,
				Options: map[string]interface{}{
					"allowedInvalidRoles": []interface{}{},
					"unknownFutureOption": true,
				},
			},
			// Whole options is nil / wrong type → defaults apply (no
			// allowed list, ignoreNonDOM = false).
			{Code: `<div role="button" />`, Tsx: true, Options: nil},
			{Code: `<div role="button" />`, Tsx: true, Options: "not-a-map"},
			{Code: `<div role="button" />`, Tsx: true, Options: 42},
			{Code: `<div role="button" />`, Tsx: true, Options: []interface{}{}},

			// ============================================================
			// ignoreNonDOM option — non-boolean values.
			// rslint's `parseOptions` does a type-asserted bool read;
			// non-bool values (string "true", numeric 1, nil) silently fall
			// back to the default `false`, so the rule continues to validate
			// custom components. ESLint would reject these at schema-load
			// time; we match the runtime behavior on rslint-side.
			// ============================================================
			{Code: `<div role="button" />`, Tsx: true, Options: map[string]interface{}{"ignoreNonDOM": "true"}},
			{Code: `<div role="button" />`, Tsx: true, Options: map[string]interface{}{"ignoreNonDOM": 1}},

			// ============================================================
			// Settings — missing / malformed.
			// GetElementType must tolerate (a) no settings at all,
			// (b) settings without `jsx-a11y`, (c) `jsx-a11y` as a wrong
			// type, (d) `components` / `polymorphicPropName` as wrong types
			// — falling back to the raw JSX tag name in each case.
			// ============================================================
			{Code: `<Foo role="button" />`, Tsx: true, Settings: nil},
			{Code: `<Foo role="button" />`, Tsx: true, Settings: map[string]interface{}{}},
			{Code: `<Foo role="button" />`, Tsx: true, Settings: map[string]interface{}{"jsx-a11y": nil}},
			{Code: `<Foo role="button" />`, Tsx: true, Settings: map[string]interface{}{"jsx-a11y": "not-a-map"}},
			{
				Code: `<Foo role="button" />`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"components": "not-a-map",
					},
				},
			},
			// components map with a non-string value — entry is ignored;
			// `Foo` keeps its raw type. Rule still runs and accepts the
			// valid "button".
			{
				Code: `<Foo role="button" />`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"components": map[string]interface{}{"Foo": 42},
					},
				},
			},
			// polymorphicPropName as empty string — no resolution attempt.
			{
				Code: `<Foo as="div" role="button" />`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"polymorphicPropName": "",
					},
				},
			},

			// ============================================================
			// Multiple attributes on the same element — only `role`
			// triggers the validation branch. id / className / style /
			// other a11y attrs co-existing must not change verdict.
			// ============================================================
			{Code: `<div id="x" className="y" role="button" data-test="z" />`, Tsx: true},
			{Code: `<div role="button" aria-label="ok" aria-hidden="false" />`, Tsx: true},

			// ============================================================
			// Spread + role coexistence — JsxSpreadAttribute doesn't fire
			// the listener (different node kind). Listener sees only the
			// JsxAttribute(s).
			// ============================================================
			{Code: `<div {...rest} role="button" />`, Tsx: true},
			{Code: `<div role="button" {...rest} />`, Tsx: true},
			{Code: `<div role="button" {...a} {...b} />`, Tsx: true},

			// ============================================================
			// Deep nesting — listener fires on each JsxAttribute regardless
			// of nesting depth. All-valid case: no diagnostic anywhere.
			// ============================================================
			{
				Code: `<section role="region"><nav role="navigation"><ul><li role="listitem"><a role="link">x</a></li></ul></nav></section>`,
				Tsx:  true,
			},

			// ============================================================
			// Repeated tokens — `every` accepts duplicates because each
			// individual token is still in the valid set.
			// ============================================================
			{Code: `<div role="button button" />`, Tsx: true},
			{Code: `<div role="tabpanel tabpanel tabpanel" />`, Tsx: true},

			// ============================================================
			// JsxNamespacedName attribute names — `xlink:role`, `role:foo`
			// uppercase to "XLINK:ROLE" / "ROLE:FOO", neither equals
			// "ROLE", so they skip. Locks the `propName(attr).toUpperCase
			// () === 'ROLE'` exact-string check against namespace-name
			// false-positives.
			// ============================================================
			{Code: `<svg xlink:role="invalid" />`, Tsx: true},

			// ============================================================
			// Real-world component-library patterns — verify the rule
			// integrates cleanly with the polymorphic / asChild idioms
			// shipped by Radix, Headless UI, styled-components, etc.
			// ============================================================
			// Radix Slot pattern — `asChild` polymorphic + ARIA role.
			{
				Code: `<Slot asChild role="button"><button>Click</button></Slot>`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"polymorphicPropName": "asChild",
					},
				},
			},
			// styled-components: custom component name; ignoreNonDOM
			// silences it.
			{Code: `<StyledButton role="button" />`, Tsx: true, Options: ignoreNonDOMOption},
			// Conditional inside Fragment-wrapped list — listener iterates
			// every child independently.
			{
				Code: `<>{items.map(item => <div role="listitem" key={item.id}>{item.label}</div>)}</>`,
				Tsx:  true,
			},
			// JSX in a ternary; both branches valid.
			{Code: `cond ? <div role="alert" /> : <div role="status" />`, Tsx: true},
			// JSX as a function argument — listener still fires.
			{Code: `render(<div role="button">Click</div>)`, Tsx: true},
			// Component returning JSX with computed-but-static role.
			{Code: `function C() { return <div role="region" aria-label="x" />; }`, Tsx: true},

			// ============================================================
			// JSX children siblings — listener handles back-to-back JSX
			// elements in the same parent without bleed.
			// ============================================================
			{Code: `<div><span role="button" />text<span role="link" /></div>`, Tsx: true},

			// ============================================================
			// `<div role={null}>` with allowedInvalidRoles=["null"] —
			// LITERAL_TYPES.Literal returns the STRING "null"; if the user
			// allow-lists "null", it becomes valid. Locks the
			// allow-list-vs-canonical short-circuit ordering.
			// ============================================================
			{
				Code: `<div role={null} />`, Tsx: true,
				Options: map[string]interface{}{"allowedInvalidRoles": []interface{}{"null"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// jsx-ast-utils' Literal extractor coerces case-insensitive
			// "true"/"false" string literals to actual JS booleans. The
			// rule then String()s them back to "true"/"false" — neither is
			// a valid role.
			// ============================================================
			{Code: `<div role="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role="True" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role="TRUE" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role="false" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role="False" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},

			// ============================================================
			// Boolean / numeric / bigint literals via JSXExpressionContainer
			// — stringified and fail the role lookup.
			//   String(true) === "true"
			//   String(false) === "false"
			//   String(1)     === "1"
			//   String(-1)    === "-1"
			//   String(1n)    === "1"
			// ============================================================
			{Code: `<div role={true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={false} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={-1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={1n} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},

			// ============================================================
			// Whitespace handling — JS `String.prototype.split(' ')` splits
			// on a single ASCII space only; tabs / newlines remain part of
			// the token; leading / trailing / double spaces produce empty
			// tokens which never match a real role.
			// ============================================================
			{Code: `<div role="button  tabpanel" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role=" button" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role="button " />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},

			// ============================================================
			// Name case — uppercase / mixed-case prop names are still
			// validated. Locks upstream's `propName(attr).toUpperCase()`.
			// ============================================================
			{Code: `<div ROLE="foobar" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div Role="foobar" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},

			// ============================================================
			// Position assertion — `<div role="foobar" />` invariant:
			// the diagnostic lands on the JsxAttribute, which begins at
			// the prop name `role`. Multi-line case exercises a non-1
			// line position too.
			// ============================================================
			{
				Code:   `<div role="foobar" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage, Line: 1, Column: 6, EndLine: 1, EndColumn: 19}},
			},
			{
				Code: "<div\n  role=\"foobar\"\n/>", Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaRole", Message: errorMessage, Line: 2, Column: 3, EndLine: 2, EndColumn: 16}},
			},

			// ============================================================
			// Option shapes — array-wrapped and bare-map both flow through
			// GetOptionsMap. Pair with a Valid case above to fully cover
			// the option-parsing path.
			// ============================================================
			{
				Code: `<div role="datepicker" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{}},
				Errors:  []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{
				Code: `<div role="datepicker" />`, Tsx: true,
				Options: map[string]interface{}{},
				Errors:  []rule_tester.InvalidTestCaseError{invalidRoleError()}},

			// ============================================================
			// ignoreNonDOM=true + custom component → skip; but the same
			// component WITHOUT ignoreNonDOM still validates.
			// ============================================================
			{Code: `<Foo role="datepicker" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},

			// ============================================================
			// Polymorphic prop resolves Box → div (DOM); rule runs and
			// "datepicker" fails.
			// ============================================================
			{
				Code: `<Box asChild="div" role="datepicker" />`, Tsx: true,
				Settings: customDivSettings,
				Errors:   []rule_tester.InvalidTestCaseError{invalidRoleError()}},

			// ============================================================
			// Polymorphic with allow-list — `<Foo>` is on the allow-list,
			// `as="div"` resolves Foo → div → DOM → rule runs → fail.
			// Locks the GetElementType polymorphicAllowList branch.
			// ============================================================
			{
				Code: `<Foo as="div" role="datepicker" />`, Tsx: true,
				Options: ignoreNonDOMOption,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"polymorphicPropName":  "as",
						"polymorphicAllowList": []interface{}{"Foo"},
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},

			// ============================================================
			// TemplateExpression with substitution — staticEvalTemplate
			// synthesizes a placeholder string from each quasi + extracted
			// substitution. The synthesized string almost never matches a
			// real role and falls through to the diagnostic. Locks the
			// template path so a future refactor that adds full substitution
			// resolution still passes the equivalent test.
			// ============================================================
			{
				Code: "<div role={`prefix-${name}-suffix`} />", Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},

			// ============================================================
			// Nesting — inner JSX with invalid role MUST still report,
			// even when the outer JSX has a valid role.
			// ============================================================
			{
				Code: `<div role="button"><span role="datepicker" /></div>`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},
			{
				Code: `<><div role="invalid" /></>`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},

			// ============================================================
			// Multiple invalid attributes on the SAME element — listener
			// fires per JsxAttribute, so multiple `role` attrs (TS allows
			// duplicate prop names with a warning) all report. We assert
			// only the well-formed single-`role` case here; duplicate-
			// attribute behavior depends on the parser dedup pass and is
			// outside the rule's responsibility.
			// ============================================================
			{
				Code: `<div role="datepicker" aria-hidden />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},

			// ============================================================
			// ArrayLiteralExpression invalid cases — upstream's LITERAL_TYPES
			// evaluates the array, filters null, joins with `,`, then
			// String().split(' ') produces a single token containing the
			// comma. None of these match a canonical role.
			//
			//   String([])               === ""
			//   String(['foobar'])       === "foobar"
			//   String(['button','x'])   === "button,x"  // single token
			//   String([1])              === "1"
			//   String([null])           === ""          // after filter
			// ============================================================
			{Code: `<div role={[]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={["foobar"]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={["range"]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={["button", "tabpanel"]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={[1]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={[null]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={[undefined]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},

			// ============================================================
			// Case-sensitive `validRoles.has(val)` — every common variation
			// of a real role name fails because the set uses lowercase
			// canonical keys only.
			// ============================================================
			{Code: `<div role="BUTTON" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role="Tab" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role="LinK" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			// DPUB role with capitalized form.
			{Code: `<div role="Doc-Abstract" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			// Graphics role with leading capital.
			{Code: `<div role="Graphics-Document" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},

			// ============================================================
			// Numeric-literal role values — String(n) coerces to JS's
			// canonical decimal form; never matches a role.
			// ============================================================
			{Code: `<div role={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={1.5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={0.5} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			// Numeric prefixes (hex / binary / octal) are decimal-
			// normalized by tsgo's parser; result is always a base-10
			// integer. Locks tsgo's normalization quirk.
			{Code: `<div role={0x1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={0b1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={0o7} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},

			// ============================================================
			// Tab / newline inside the role string — `split(' ')` only
			// splits on the literal space, so the token contains the raw
			// whitespace and fails the role lookup.
			// ============================================================
			{Code: `<div role={"button\ttabpanel"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},
			{Code: `<div role={"button\nlink"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()}},

			// ============================================================
			// Non-bool `ignoreNonDOM` falls back to default `false`. A
			// custom component therefore still validates and the invalid
			// role reports.
			// ============================================================
			{
				Code: `<Foo role="datepicker" />`, Tsx: true,
				Options: map[string]interface{}{"ignoreNonDOM": "true"},
				Errors:  []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},
			{
				Code: `<Foo role="datepicker" />`, Tsx: true,
				Options: map[string]interface{}{"ignoreNonDOM": 1},
				Errors:  []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},

			// ============================================================
			// components map redirects a custom name to a DOM element →
			// rule runs against the resolved type (no ignoreNonDOM bypass).
			// ============================================================
			{
				Code: `<MyDiv role="datepicker" />`, Tsx: true,
				Options: ignoreNonDOMOption,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"components": map[string]interface{}{"MyDiv": "div"},
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},
			// components map redirects div → NotADomElement → ignoreNonDOM
			// suppresses the rule. We assert by flipping ignoreNonDOM off:
			// the diagnostic must surface.
			{
				Code: `<div role="datepicker" />`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"components": map[string]interface{}{"div": "NotADomElement"},
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},

			// ============================================================
			// Spread + invalid role — spread doesn't suppress the
			// JsxAttribute listener. Diagnostic still surfaces on `role`.
			// ============================================================
			{
				Code: `<div {...props} role="datepicker" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},
			{
				Code: `<div role="datepicker" {...props} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},

			// ============================================================
			// Same element, role + sibling props — listener fires only on
			// `role`. Single diagnostic.
			// ============================================================
			{
				Code: `<div id="x" className="y" role="datepicker" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},

			// ============================================================
			// Deep nesting — diagnostics fire INDEPENDENTLY at every
			// JsxAttribute. Three invalid roles at different depths →
			// three diagnostics.
			// ============================================================
			{
				Code: `<section role="invalid-outer"><nav><ul><li role="invalid-middle"><a role="invalid-inner">x</a></li></ul></nav></section>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidRoleError(),
					invalidRoleError(),
					invalidRoleError(),
				},
			},

			// ============================================================
			// Siblings — two adjacent JSX elements, each with a different
			// invalid role.
			// ============================================================
			{
				Code: `<div><span role="invalid-a" /><span role="invalid-b" /></div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidRoleError(),
					invalidRoleError(),
				},
			},

			// ============================================================
			// JSX inside a Fragment, inside a ternary, with mapped list —
			// real-world list-rendering patterns. Two invalid roles fire
			// independently even when the JSX context wraps them.
			// ============================================================
			{
				Code: `<>{cond ? <div role="invalid-a" /> : <div role="invalid-b" />}</>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					invalidRoleError(),
					invalidRoleError(),
				},
			},

			// ============================================================
			// Real-world: Radix Slot with bad role. Polymorphic asChild +
			// invalid role on the wrapper.
			// ============================================================
			{
				Code: `<Slot asChild role="datepicker"><button>Click</button></Slot>`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"polymorphicPropName": "asChild",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},

			// ============================================================
			// Real-world: list inside a Component returning JSX. Stresses
			// listener traversal into nested function/arrow bodies.
			// ============================================================
			{
				Code: `function List() { return items.map(x => <li role="invalid" key={x.id} />); }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},

			// ============================================================
			// Self-closing vs open-close emit identical diagnostics — lock
			// in both forms to guard against a future shape-sensitive
			// regression.
			// ============================================================
			{
				Code: `<div role="datepicker"></div>`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},
			{
				Code: `<div role="datepicker" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{invalidRoleError()},
			},
		},
	)
}
