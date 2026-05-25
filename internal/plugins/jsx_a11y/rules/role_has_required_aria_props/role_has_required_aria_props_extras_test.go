// cspell:ignore eading inspectable

package role_has_required_aria_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestRoleHasRequiredAriaPropsExtras covers cases NOT in upstream's
// `__tests__/src/rules/role-has-required-aria-props-test.js` — universal
// edge shapes (Dimension 4 of the port skill), tsgo AST quirks, and
// lock-in tests for each branch in the upstream `JSXAttribute` listener
// that the upstream test file leaves uncovered.
func TestRoleHasRequiredAriaPropsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RoleHasRequiredAriaPropsRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Boolean-form / null / boolean-coerced role values — every
			// shape that String()-coerces to a non-role token, so
			// validRoles ends up empty and the rule skips.
			// ============================================================

			// `<div role />` (boolean form) — getLiteralPropValue returns
			// true (boolean), `String(true)` = "true", which isn't a
			// valid role token → validRoles empty → skip. Upstream's
			// aria-role reports this; role-has-required-aria-props does
			// not.
			{Code: `<div role />`, Tsx: true},
			// `<div role={null} />` — LITERAL_TYPES.Literal override →
			// magic string "null". Not a valid role → skip.
			{Code: `<div role={null} />`, Tsx: true},
			// `<div role={true} />` / `<div role={false} />` — String()
			// coerces to "true"/"false", neither is a valid role.
			{Code: `<div role={true} />`, Tsx: true},
			{Code: `<div role={false} />`, Tsx: true},
			// `<div role="" />` — empty string, splits to [""], not a
			// valid role.
			{Code: `<div role="" />`, Tsx: true},

			// ============================================================
			// JsxExpression-wrapped string literal — same handling as
			// direct string. Slider needs aria-valuenow which IS present.
			// ============================================================
			{Code: `<div role={"slider"} aria-valuenow={50} />`, Tsx: true},

			// ============================================================
			// Lowercase normalization — role="HEADING" still validates as
			// "heading" via the String.toLowerCase().split path, and
			// aria-level is present.
			// ============================================================
			{Code: `<div role="HEADING" aria-level={2} />`, Tsx: true},

			// ============================================================
			// Space-separated multiple roles — both heading and slider
			// have their required props satisfied. Locks in the
			// `normalized.split(' ')` + per-role forEach branches.
			// ============================================================
			{
				Code: `<div role="heading slider" aria-level={2} aria-valuenow={5} />`,
				Tsx:  true,
			},

			// ============================================================
			// Mixed valid + invalid token — only "heading" is in
			// roleKeys; the "invalid-role" token is filtered out by the
			// validRoles filter. aria-level present → pass. Locks in the
			// upstream `.filter(...)` arm.
			// ============================================================
			{Code: `<div role="heading invalid-role" aria-level={2} />`, Tsx: true},

			// ============================================================
			// Semantic-role skip — every concept entry in
			// `jsxa11yutil.SemanticRoleConcepts` must be exercised so a
			// future axobject-query refresh that removes a row trips a
			// test. None of these need their normally-required props
			// because the semantic skip bypasses the check.
			// ============================================================

			// `<select>` → ["combobox", "listbox"]; both should skip.
			{Code: `<select role="combobox" />`, Tsx: true},
			{Code: `<select role="listbox" />`, Tsx: true},
			// `<option>` → ["option"]; option has aria-selected as required.
			{Code: `<option role="option" />`, Tsx: true},
			// `<h1>` … `<h6>` → ["heading"]; heading needs aria-level but
			// the semantic skip bypasses the check.
			{Code: `<h1 role="heading" />`, Tsx: true},
			{Code: `<h2 role="heading" />`, Tsx: true},
			{Code: `<h3 role="heading" />`, Tsx: true},
			{Code: `<h4 role="heading" />`, Tsx: true},
			{Code: `<h5 role="heading" />`, Tsx: true},
			{Code: `<h6 role="heading" />`, Tsx: true},
			// `<input type="radio" role="radio" />` — radio needs
			// aria-checked but the semantic skip kicks in.
			{Code: `<input type="radio" role="radio" />`, Tsx: true},
			// `<input type="range" role="slider" />` — slider needs
			// aria-valuenow but the semantic skip kicks in.
			{Code: `<input type="range" role="slider" />`, Tsx: true},

			// ============================================================
			// JsxExpression-wrapped `type` attribute on input — the inner
			// expression is a string literal, so LiteralPropStringValue
			// resolves "checkbox" and the concept-attribute match
			// succeeds. Locks in the upstream `getLiteralPropValue(attr)`
			// path inside isSemanticRoleElement.
			// ============================================================
			{Code: `<input type={"checkbox"} role="switch" />`, Tsx: true},

			// ============================================================
			// Array literal role — upstream's LITERAL_TYPES.ArrayExpression
			// evaluates elements via TYPES and `,`-joins per
			// `String([...])`. Empty array → "" → empty split token →
			// validRoles empty → skip.
			// ============================================================
			{Code: `<div role={[]} />`, Tsx: true},
			// Array with single token that comma-joins to a non-role
			// string — `["heading,slider"]` joins to `"heading,slider"`,
			// which split on space yields one token containing comma →
			// not a valid role → skip.
			{Code: `<div role={["heading,slider"]} />`, Tsx: true},
			// Multi-element array — `["heading", "slider"]` joins to
			// `"heading,slider"`, same result.
			{Code: `<div role={["heading", "slider"]} />`, Tsx: true},

			// ============================================================
			// Backtick template literal without substitutions — upstream
			// treats `\`heading\`` the same as a string literal (no
			// boolean coerce), so role="heading" requires aria-level.
			// ============================================================
			{Code: "<div role={`heading`} aria-level={2} />", Tsx: true},

			// ============================================================
			// Template with substitution — staticEvalTemplate synthesizes
			// a placeholder string like "heading{x}" which doesn't match
			// any valid role token. validRoles = []. Skip.
			// ============================================================
			{Code: "<div role={`heading${x}`} />", Tsx: true},

			// ============================================================
			// Parenthesized expressions — tsgo preserves; upstream's
			// ESTree parser flattens. LiteralPropAriaValue strips parens
			// so the inner string literal still resolves.
			// ============================================================
			{Code: `<div role={("row")} />`, Tsx: true},
			{Code: `<div role={(("heading"))} aria-level={2} />`, Tsx: true},

			// ============================================================
			// TypeScript wrappers — `as` / `satisfies` / non-null `!` all
			// noop → NoLit in LITERAL_TYPES → skip. Locks in upstream's
			// TS-wrapper handling: even though the inner literal would
			// otherwise be inspectable, the wrapper makes the value opaque
			// to the literal extractor.
			// ============================================================
			{Code: `<div role={"heading" as const} />`, Tsx: true},
			{Code: `<div role={"heading" as string} />`, Tsx: true},
			{Code: `<div role={"heading"!} />`, Tsx: true},
			{Code: `<div role={"heading" satisfies string} />`, Tsx: true},

			// ============================================================
			// Member access / optional chain / call expression role value
			// — all LITERAL_TYPES noop → NoLit → skip. These represent
			// real-world patterns where the role is computed.
			// ============================================================
			{Code: `<div role={obj.role} />`, Tsx: true},
			{Code: `<div role={obj?.role} />`, Tsx: true},
			{Code: `<div role={getRole()} />`, Tsx: true},
			{Code: `<div role={obj.deep?.nested?.role} />`, Tsx: true},
			{Code: `<div role={cond ? "checkbox" : "radio"} />`, Tsx: true},
			// String concat — LITERAL_TYPES.BinaryExpression noop for `+`.
			// Even though staticEval could resolve "head" + "ing", the
			// literal-extractor path doesn't.
			{Code: `<div role={"head" + "ing"} />`, Tsx: true},
			// Nullish coalescing / logical falls through to noop too.
			{Code: `<div role={role ?? "button"} />`, Tsx: true},

			// ============================================================
			// Numeric / BigInt role values — String coerce to numeric
			// digits which aren't valid roles. validRoles empty → skip.
			// ============================================================
			{Code: `<div role={0} />`, Tsx: true},
			{Code: `<div role={42} />`, Tsx: true},
			{Code: `<div role={1.5} />`, Tsx: true},
			{Code: `<div role={42n} />`, Tsx: true},
			// `NaN` / `Infinity` are Identifiers (not numeric literals)
			// → LITERAL_TYPES.Identifier returns null for non-undefined
			// identifiers → skip.
			{Code: `<div role={NaN} />`, Tsx: true},
			{Code: `<div role={Infinity} />`, Tsx: true},

			// ============================================================
			// Whitespace-only role values — single space splits to ["", ""],
			// validRoles empty.
			// ============================================================
			{Code: `<div role=" " />`, Tsx: true},
			{Code: `<div role="   " />`, Tsx: true},
			// Tabs / newlines in role — upstream splits on single ASCII
			// space only, so non-space whitespace stays in the token and
			// the token isn't a valid role.
			{Code: "<div role=\"heading\tslider\" />", Tsx: true},

			// ============================================================
			// HTML entity decoding on DIRECT string attribute — JSX
			// parsers decode `&#NN;` / `&NAME;` at parse time, so
			// `<div role="&#104;eading">` resolves to "heading" and the
			// aria-level requirement applies. tsgo doesn't decode by
			// default — the LiteralPropStringValue helper invokes
			// `jsxtransforms.DecodeEntities` to realign. The {…}-wrapped
			// form is a JS string literal (NOT JSX attribute text), so
			// it does NOT get entity-decoded — that's tested by the
			// invalid case below.
			// ============================================================
			{Code: `<div role="&#104;eading" aria-level={2} />`, Tsx: true},

			// ============================================================
			// Required-prop presence-check semantics — upstream `getProp`
			// returns the JsxAttribute when the name matches, regardless
			// of value type. So presence (not value validity) is what
			// matters for the required-props check. Every shape below
			// counts as "present".
			// ============================================================
			// Boolean form: `<div role="checkbox" aria-checked />` →
			// aria-checked has no value (null in upstream's getProp), but
			// the attribute exists → present.
			{Code: `<div role="checkbox" aria-checked />`, Tsx: true},
			// `={null}` / `={undefined}` — attribute exists with a value
			// expression, even if the value is nullish.
			{Code: `<div role="checkbox" aria-checked={null} />`, Tsx: true},
			{Code: `<div role="checkbox" aria-checked={undefined} />`, Tsx: true},
			// `={true}` / `={false}` — booleans count as present.
			{Code: `<div role="checkbox" aria-checked={true} />`, Tsx: true},
			{Code: `<div role="checkbox" aria-checked={false} />`, Tsx: true},
			// Case-insensitive name lookup — `getProp` uses ignoreCase:true
			// by default.
			{Code: `<div role="checkbox" Aria-Checked="false" />`, Tsx: true},
			{Code: `<div role="checkbox" ARIA-CHECKED="false" />`, Tsx: true},
			// Literal-object spread with computed-IDENTIFIER key — upstream's
			// `getProp` matches `{[ariaChecked]: "false"}` by inspecting
			// the inner Identifier without consulting the `computed`
			// flag. This is upstream's quirk: bracket-wrapped Identifier
			// `[someIdent]` reads as a matchable key, even though at
			// runtime the bracket form evaluates the expression.
			// (Symbolic — won't compile as TypeScript without ariaChecked
			// in scope, so we test via a runnable shape with an in-scope
			// const.)
			// Note: kebab-case names like `aria-checked` are NOT legal
			// JS identifiers, so the only way to have aria-checked in a
			// literal-object spread is via a string-literal key (which
			// upstream does NOT match — see the invalid case below).

			// ============================================================
			// Multi-required-prop role with both satisfied — combobox
			// needs aria-controls AND aria-expanded.
			// ============================================================
			{
				Code: `<div role="combobox" aria-controls="list" aria-expanded={true} />`,
				Tsx:  true,
			},
			{
				Code: `<div role="scrollbar" aria-controls="content" aria-valuenow={50} />`,
				Tsx:  true,
			},

			// ============================================================
			// JSX-namespaced tag (svg:rect) — IsDOMElement returns false
			// because aria-query's dom map keys on bare HTML tag names.
			// Skip without checking required props.
			// ============================================================
			{Code: `<svg:rect role="checkbox" />`, Tsx: true},

			// ============================================================
			// Member-expression tag (Foo.Bar) — non-DOM → skip.
			// ============================================================
			{Code: `<Foo.Bar role="checkbox" />`, Tsx: true},
			{Code: `<Foo.Bar.Baz role="checkbox" />`, Tsx: true},

			// ============================================================
			// SVG tag (not in aria-query's dom map) — `<svg>` is not a
			// recognized DOM element per aria-query, so the rule skips.
			// ============================================================
			{Code: `<svg role="checkbox" />`, Tsx: true},

			// ============================================================
			// Multi-line JSX with required prop present — the rule fires
			// per attribute regardless of source-level newlines.
			// ============================================================
			{
				Code: "<div\n  role=\"heading\"\n  aria-level={2}\n/>",
				Tsx:  true,
			},

			// ============================================================
			// JSX inside a conditional / ternary / map — listener fires
			// on the inner element's attributes regardless of nesting.
			// These are real-world rendering patterns.
			// ============================================================
			{Code: `<div>{cond && <span role="row" />}</div>`, Tsx: true},
			{Code: `<div>{cond ? <span role="row" /> : null}</div>`, Tsx: true},
			{Code: `<>{[1,2].map(x => <span key={x} role="row" />)}</>`, Tsx: true},

			// ============================================================
			// JSX fragment around role-bearing elements — fragment has
			// no attributes; inner role attribute fires normally.
			// ============================================================
			{Code: `<><div role="row" /></>`, Tsx: true},

			// ============================================================
			// Roles WITHOUT required props on DOM elements — should NOT
			// report. Sanity-check that the rule doesn't accidentally
			// flag roles whose requiredProps set is empty.
			// ============================================================
			{Code: `<div role="button" />`, Tsx: true},
			{Code: `<div role="link" />`, Tsx: true},
			{Code: `<div role="alert" />`, Tsx: true},
			{Code: `<div role="presentation" />`, Tsx: true},
			{Code: `<div role="none" />`, Tsx: true},
			{Code: `<div role="dialog" />`, Tsx: true},
			{Code: `<div role="tabpanel" />`, Tsx: true},
			{Code: `<div role="alertdialog" />`, Tsx: true},

			// ============================================================
			// `<button role="checkbox" aria-checked="false" />` — button
			// is a DOM element. Button as concept maps only to ["button"]
			// (not in SemanticRoleConcepts since button has no required
			// props), so no semantic skip for "checkbox". aria-checked
			// IS present → no report.
			// ============================================================
			{Code: `<button role="checkbox" aria-checked="false" />`, Tsx: true},

			// ============================================================
			// Anchor with non-required-props role — anchor is DOM; role
			// "tab" has no required props → no report.
			// ============================================================
			{Code: `<a href="#" role="tab" />`, Tsx: true},

			// ============================================================
			// `<input>` without role — no role attribute means the
			// listener doesn't fire on a role match. No report.
			// ============================================================
			{Code: `<input type="checkbox" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Space-separated multiple roles, BOTH missing required
			// props → two diagnostics, one per role, in upstream-forEach
			// order. Locks in the per-role report loop branch upstream
			// doesn't test.
			// ============================================================
			{
				Code: `<div role="heading slider" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "role-has-required-aria-props",
						Message:   errorMessage("heading", []string{"aria-level"}),
						Line:      1, Column: 6,
					},
					{
						MessageId: "role-has-required-aria-props",
						Message:   errorMessage("slider", []string{"aria-valuenow"}),
						Line:      1, Column: 6,
					},
				},
			},

			// ============================================================
			// Mixed: one role satisfied, one missing — only the failing
			// role reports. heading has aria-level; slider missing
			// aria-valuenow.
			// ============================================================
			{
				Code: `<div role="heading slider" aria-level={2} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("slider", []string{"aria-valuenow"}),
					Line:      1, Column: 6,
				}},
			},

			// ============================================================
			// Uppercase role value → toLowerCase()-normalized for
			// validRoles filter, so "HEADING" is treated as "heading"
			// and the missing aria-level still reports. Locks in the
			// `.toLowerCase().split(' ')` step.
			// ============================================================
			{
				Code: `<div role="HEADING" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("heading", []string{"aria-level"}),
					Line:      1, Column: 6,
				}},
			},

			// ============================================================
			// Uppercase role value on a semantic-role element: the
			// semantic-role check uses the RAW (case-preserved) value,
			// so it does NOT skip; the validRoles filter still reports.
			// `<select role="COMBOBOX" />`: roleAttrValue = "COMBOBOX",
			// isSemanticRoleElement comparing "COMBOBOX" to aria-query's
			// lowercase "combobox" returns false → no skip → missing
			// required props reported. Locks in the case-sensitivity
			// quirk between the two paths.
			// ============================================================
			{
				Code: `<select role="COMBOBOX" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("combobox", []string{"aria-controls", "aria-expanded"}),
					Line:      1, Column: 9,
				}},
			},

			// ============================================================
			// JsxExpression-wrapped string literal as role value — same
			// handling as direct string. slider missing aria-valuenow.
			// ============================================================
			{
				Code: `<div role={"slider"} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("slider", []string{"aria-valuenow"}),
					Line:      1, Column: 6,
				}},
			},

			// ============================================================
			// `<input type="CHECKBOX" role="switch" />` — concept-attribute
			// value match is case-sensitive ("CHECKBOX" != "checkbox") →
			// semantic skip fails → "switch" requires aria-checked,
			// missing → report. Locks in case-sensitive value comparison
			// inside isSemanticRoleElement.
			// ============================================================
			{
				Code: `<input type="CHECKBOX" role="switch" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("switch", []string{"aria-checked"}),
					Line:      1, Column: 24,
				}},
			},

			// ============================================================
			// Dynamic `type={typeVar}` on input — LiteralPropStringValue
			// returns ("", false) for the type prop, so the
			// concept-attribute value match fails → semantic skip fails
			// → switch requires aria-checked, missing → report. Locks in
			// the "dynamic value, no semantic skip" branch.
			// ============================================================
			{
				Code: `<input type={typeVar} role="switch" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("switch", []string{"aria-checked"}),
					Line:      1, Column: 23,
				}},
			},

			// ============================================================
			// Bare `<input role="checkbox" />` without type=checkbox —
			// the concept requires the type attribute to match, so no
			// semantic skip; missing aria-checked reports. Locks in the
			// "attribute absent → concept doesn't match" branch.
			// ============================================================
			{
				Code: `<input role="checkbox" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("checkbox", []string{"aria-checked"}),
					Line:      1, Column: 8,
				}},
			},

			// ============================================================
			// Array literal role on a semantic-role element — upstream's
			// strict `role.name === roleAttrValue` comparison treats
			// the array value as NOT a string, so the semantic skip
			// MUST NOT fire even though the joined string would have
			// matched. `<input type="checkbox" role={["switch"]} />`:
			// validRoles = ["switch"], semantic skip refuses (array
			// value), missing aria-checked → report.
			// ============================================================
			{
				Code: `<input type="checkbox" role={["switch"]} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("switch", []string{"aria-checked"}),
					Line:      1, Column: 24,
				}},
			},
			// Same shape with `["checkbox"]` instead of `["switch"]` —
			// no semantic skip; checkbox needs aria-checked → report.
			{
				Code: `<select role={["combobox"]} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("combobox", []string{"aria-controls", "aria-expanded"}),
					Line:      1, Column: 9,
				}},
			},

			// ============================================================
			// Array role with valid token but missing required props —
			// `<div role={["heading"]} />` joins to "heading", validRoles
			// = ["heading"], missing aria-level → report.
			// ============================================================
			{
				Code: `<div role={["heading"]} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("heading", []string{"aria-level"}),
					Line:      1, Column: 6,
				}},
			},

			// ============================================================
			// Literal-spread with STRING-LITERAL key — upstream's
			// `getProp` only walks Identifier-typed keys. `aria-checked`
			// can't be a JS Identifier (hyphenated names aren't legal),
			// so the only way to put aria-* in a literal spread is via
			// a string-literal key — which getProp does NOT match. The
			// rule therefore reports as if aria-checked is absent, even
			// though the spread "logically" provides it. Locks in
			// upstream's quirky-but-deliberate Identifier-only key
			// match.
			// ============================================================
			{
				Code: `<div role="checkbox" {...{"aria-checked": "false"}} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("checkbox", []string{"aria-checked"}),
					Line:      1, Column: 6,
				}},
			},

			// ============================================================
			// Non-literal spread — opaque. FindAttributeByName cannot see
			// inside `{...someObj}`, so the required-prop is treated as
			// absent → report. Locks in the upstream `getProp`
			// strict-spread behavior.
			// ============================================================
			{
				Code: `<div role="checkbox" {...someObj} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("checkbox", []string{"aria-checked"}),
					Line:      1, Column: 6,
				}},
			},

			// ============================================================
			// Mixed case attribute name in concept — upstream's
			// `cAttr.name === propName(attr)` is case-SENSITIVE
			// ("type" != "Type"), so `<input Type="checkbox">` does NOT
			// match the input/type=checkbox concept → no semantic
			// skip → switch's aria-checked missing → report.
			// ============================================================
			{
				Code: `<input Type="checkbox" role="switch" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("switch", []string{"aria-checked"}),
					Line:      1, Column: 24,
				}},
			},

			// ============================================================
			// Nested JSX — each role attribute fires its own listener
			// invocation independently, so a missing required-prop on
			// the inner element reports without interference from the
			// outer element's role.
			// ============================================================
			{
				Code: `<div role="heading" aria-level={2}><span role="checkbox" /></div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("checkbox", []string{"aria-checked"}),
					Line:      1, Column: 42,
				}},
			},

			// ============================================================
			// Backtick template with substitution — upstream's
			// TemplateLiteral extractor synthesizes a placeholder string
			// like "heading{x}" which split on space won't yield
			// "heading" as a token. validRoles = []. Skip. But a
			// pure-quasi template (no substitution) is a string and
			// IS extractable — covered in the valid section.
			// ============================================================
			// (No invalid case for the with-substitution form — it's a
			// valid case because validRoles is empty.)

			// ============================================================
			// Per-role data-table lock-in for the 6 roles upstream's
			// invalid suite doesn't exercise. The generated
			// `basicValidityTests` test that EACH role passes when its
			// required props are present, but that suite builds the JSX
			// FROM the same data table — so a typo in the table would
			// produce a matching JSX and silently pass. Invalid tests
			// where the prop is GENUINELY ABSENT lock in the prop name.
			// ============================================================

			// menuitemcheckbox — requires aria-checked.
			{
				Code: `<div role="menuitemcheckbox" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("menuitemcheckbox", []string{"aria-checked"}),
					Line:      1, Column: 6,
				}},
			},
			// menuitemradio — requires aria-checked.
			{
				Code: `<div role="menuitemradio" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("menuitemradio", []string{"aria-checked"}),
					Line:      1, Column: 6,
				}},
			},
			// meter — requires aria-valuenow.
			{
				Code: `<div role="meter" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("meter", []string{"aria-valuenow"}),
					Line:      1, Column: 6,
				}},
			},
			// radio — requires aria-checked. Use <div> rather than
			// <input type="radio"> to avoid the semantic skip.
			{
				Code: `<div role="radio" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("radio", []string{"aria-checked"}),
					Line:      1, Column: 6,
				}},
			},
			// switch — requires aria-checked. Use <div> rather than
			// <input type="checkbox"> (which maps to ["checkbox", "switch"]
			// via semantic skip).
			{
				Code: `<div role="switch" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("switch", []string{"aria-checked"}),
					Line:      1, Column: 6,
				}},
			},
			// treeitem — requires aria-selected.
			{
				Code: `<div role="treeitem" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("treeitem", []string{"aria-selected"}),
					Line:      1, Column: 6,
				}},
			},

			// ============================================================
			// Multi-required-prop partial satisfaction — combobox with
			// only ONE of its two required props present still reports.
			// Locks in the `requiredProps.every(...)` short-circuit
			// (which means "fail on first missing prop").
			// ============================================================
			{
				Code: `<div role="combobox" aria-controls="list" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("combobox", []string{"aria-controls", "aria-expanded"}),
					Line:      1, Column: 6,
				}},
			},
			{
				Code: `<div role="combobox" aria-expanded={true} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("combobox", []string{"aria-controls", "aria-expanded"}),
					Line:      1, Column: 6,
				}},
			},

			// ============================================================
			// HTML entity decoding via the JsxExpression form — a JS
			// string literal inside `{…}` is NOT entity-decoded
			// (entities are a JSX-attribute-text feature, not a JS
			// string feature). So `<div role={"&#104;eading"} />` has
			// the literal value `&#104;eading` (12 chars), which is NOT
			// a valid role → validRoles empty → skip → no report.
			// (Valid case — confirms the asymmetry between direct and
			// JsxExpression-wrapped string forms.)
			// ============================================================
			// (No invalid case — this is asserted by the valid case in
			// the corresponding section above; the JsxExpression-wrapped
			// form does NOT decode, so it doesn't trigger a "heading"
			// validRoles entry.)

			// ============================================================
			// Switch role on input WITHOUT type=checkbox — the semantic
			// concept requires type=checkbox specifically, so other
			// input types do not trigger the skip. aria-checked missing →
			// report. Locks in the "concept attribute value must match
			// exactly" branch.
			// ============================================================
			{
				Code: `<input type="text" role="switch" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("switch", []string{"aria-checked"}),
					Line:      1, Column: 20,
				}},
			},
			// Slider role on input WITHOUT type=range — similar pattern.
			{
				Code: `<input type="number" role="slider" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("slider", []string{"aria-valuenow"}),
					Line:      1, Column: 22,
				}},
			},
			// Radio role on input WITHOUT type=radio — similar.
			{
				Code: `<input type="text" role="radio" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("radio", []string{"aria-checked"}),
					Line:      1, Column: 20,
				}},
			},

			// ============================================================
			// Multi-line JSX with missing required prop — verifies that
			// the column / line assertion still works when the role
			// attribute is on a non-first line. Real-world prettier-formatted
			// JSX wraps long elements across multiple lines.
			// ============================================================
			{
				Code: "<div\n  role=\"checkbox\"\n/>",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("checkbox", []string{"aria-checked"}),
					Line:      2, Column: 3,
				}},
			},

			// ============================================================
			// Required prop present elsewhere on the parent's siblings
			// does NOT count — the rule only inspects the element's own
			// attributes. Sanity check.
			// ============================================================
			{
				Code: `<div aria-checked="true"><span role="checkbox" /></div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("checkbox", []string{"aria-checked"}),
					Line:      1, Column: 32,
				}},
			},

			// ============================================================
			// Nested role-bearing elements inside conditional / map —
			// real-world rendering pattern. Each inner role fires
			// independently.
			// ============================================================
			{
				Code: `<div>{cond && <span role="checkbox" />}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("checkbox", []string{"aria-checked"}),
					Line:      1, Column: 21,
				}},
			},
			{
				Code: `<>{[1,2].map(x => <span key={x} role="checkbox" />)}</>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("checkbox", []string{"aria-checked"}),
					Line:      1, Column: 33,
				}},
			},

			// ============================================================
			// HTML entity decoding on a value pointing at a checkbox-like
			// pattern (`&#115;` is 's' → "&#115;witch" → "switch") —
			// confirms entity decoding happens at literal extraction.
			// ============================================================
			{
				Code: `<div role="&#115;witch" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("switch", []string{"aria-checked"}),
					Line:      1, Column: 6,
				}},
			},

			// ============================================================
			// Spread + explicit role, where required prop is genuinely
			// absent (spread is opaque). Locks in non-literal spread →
			// opaque behavior.
			// ============================================================
			{
				Code: `<div {...spreadProps} role="combobox" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("combobox", []string{"aria-controls", "aria-expanded"}),
					Line:      1, Column: 23,
				}},
			},

			// ============================================================
			// Outer element with required prop + inner element with
			// missing required prop on the SAME role. Outer should not
			// "donate" its aria-checked to the inner element's check.
			// ============================================================
			{
				Code: `<div role="checkbox" aria-checked="true"><span role="checkbox" /></div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("checkbox", []string{"aria-checked"}),
					Line:      1, Column: 48,
				}},
			},

			// ============================================================
			// Heading role on multiple deeply nested elements — each
			// reports independently.
			// ============================================================
			{
				Code: `<div><div><span role="heading" /></div></div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "role-has-required-aria-props",
					Message:   errorMessage("heading", []string{"aria-level"}),
					Line:      1, Column: 17,
				}},
			},
		},
	)
}
