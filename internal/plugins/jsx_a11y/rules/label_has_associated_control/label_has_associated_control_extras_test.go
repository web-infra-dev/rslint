// label_has_associated_control_extras_test.go locks in branches that
// upstream's `__tests__/src/rules/label-has-associated-control-test.js`
// does not directly exercise but are reachable through the rule's listener
// gate, helper short-circuits, or schema-default fallbacks. Each case is a
// minimal snippet aimed at a single upstream branch.
//
// Coverage axes (mirrors .agents/skills/port-rule Dimension 1–4 + upstream
// semantic-walk lock-ins):
//
//   - Dimension 1 (tsgo AST shape): self-closing vs paired label root;
//     JsxExpression child as "best-effort" nested-control / accessible-label
//     marker; TS non-null assertion on attribute values; literal-object
//     spread vs opaque-identifier spread.
//   - Dimension 2 (scoping / nesting): control deeper than the configured
//     `depth` cap; settings.components rewires the listener gate; minimatch
//     glob in `labelComponents`.
//   - Dimension 4 (universal edge shapes): whitespace-only text; explicit
//     `null` / `undefined` / `false` / `""` htmlFor values; unknown
//     `assert` enum value silently no-ops; position assertions (single-
//     line and multi-line) on opening element.
//
// Lock-ins also cover upstream's quirky `hasProp` `spreadStrict: true`
// default: literal-resolvable spreads carrying `htmlFor` do NOT satisfy the
// presence check.
//
// Shared variables (`expectedAccessibleLabel`, `expectedHtmlFor`,
// `componentsSettings`, etc.) and helpers (`mergeOptions`, …) are declared
// in `label_has_associated_control_upstream_test.go`.
package label_has_associated_control

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestLabelHasAssociatedControl_LockInExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Listener gate: minimatch on the tag name.
			// ============================================================
			// labelComponents with glob — `Form*Label` matches FormFieldLabel
			// and renders as a label-like component; the htmlFor + aria-label
			// satisfy `either` mode.
			{Code: `<FormFieldLabel htmlFor="x" aria-label="x" />`, Tsx: true, Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"Form*Label"}}}},

			// ============================================================
			// Listener gate (negative): tag name does NOT match a label
			// pattern → skip without checks. Even though the element has no
			// text, the rule does not fire.
			// ============================================================
			{Code: `<Badge />`, Tsx: true},
			{Code: `<Badge htmlFor="x" />`, Tsx: true},

			// ============================================================
			// settings.components: CustomLabel → label rewires the listener
			// gate. Without label-attributes / labelComponents options.
			// ============================================================
			{Code: `<CustomLabel htmlFor="x" aria-label="x" />`, Tsx: true, Settings: componentsSettings},

			// ============================================================
			// settings.attributes.for: alternate htmlFor attribute name.
			// ============================================================
			{Code: `<label for="x" aria-label="x" />`, Tsx: true, Settings: attributesSettings},

			// ============================================================
			// validateHtmlFor truthiness — JsxExpression with Identifier as
			// htmlFor value is truthy under `getPropValue` (synthesized
			// non-empty Identifier name).
			// ============================================================
			{Code: `<label htmlFor={id}>Save<input /></label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "either"}}},

			// ============================================================
			// MayContainChildComponent expression-container short-circuit:
			// `<label>{children}</label>` → JsxExpression child counts as
			// both a nested control AND accessible text.
			// ============================================================
			{Code: `<label htmlFor="x">{children}</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "both"}}},

			// ============================================================
			// depth=0 still inspects the root for labelling props (depth
			// check is `depth > maxDepth`, so depth 0 at maxDepth 0 OK).
			// `<label />` has aria-label, root inspection finds it; no
			// descendants are visited.
			// ============================================================
			{Code: `<label htmlFor="x" aria-label="Save" />`, Tsx: true, Options: []interface{}{map[string]interface{}{"depth": float64(0), "assert": "htmlFor"}}},

			// ============================================================
			// depth cap at 25: a depth=999 setting clamps to 25 — still
			// resolves a shallow label.
			// ============================================================
			{Code: `<label htmlFor="x"><span>Save</span></label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"depth": float64(999), "assert": "htmlFor"}}},

			// ============================================================
			// `assert` with an unknown enum value: ESLint schema rejects
			// AND re-invokes with default options (assert=either). rslint
			// has no schema, so parseOptions normalizes non-enum values to
			// 'either' for the same observable behavior.
			//
			// VALID side: `<label htmlFor="x">Save</label>` would be valid
			// under either mode (htmlFor present), so the unknown-assert
			// behavior matches.
			// ============================================================
			{Code: `<label htmlFor="x">Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "garbage"}}},
			// Empty assert string also normalizes to 'either'.
			{Code: `<label htmlFor="x">Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": ""}}},

			// ============================================================
			// Spread-attribute opacity: `<label {...props}>Save<input/></label>`
			// — spread keeps `accessibleLabel` satisfied via HasLabellingProp's
			// short-circuit; htmlFor stays false (strict-spread `hasProp`);
			// nesting passes via the literal <input />. With assert: 'nesting'
			// the case is valid.
			// ============================================================
			{Code: `<label {...props}>Save<input /></label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "nesting"}}},

			// ============================================================
			// Self-closing label with aria-label satisfies the accessible-
			// label check even without descendants (root labelling-prop
			// branch in mayHaveAccessibleLabel).
			// ============================================================
			{Code: `<label htmlFor="x" aria-label="Save" />`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}}},

			// ============================================================
			// TS-wrapper on htmlFor value: `htmlFor={id!}` — upstream's
			// `getPropValue` extracts the wrapped Identifier name (with
			// `!` appended) as a non-empty string → truthy.
			// ============================================================
			{Code: `<label htmlFor={id!} aria-label="Save" />`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}}},

			// ============================================================
			// Spread + direct, spread FIRST with truthy value, direct AFTER
			// with empty value. Counter-intuitively VALID under upstream's
			// `getProp` first-match-wins semantic: spread carries
			// `htmlFor: "js_id"`, the direct empty `htmlFor=""` is shadowed.
			// hasProp(strict)=true (direct exists), getProp scans in source
			// order → spread first → returns spread's "js_id" → truthy.
			// ============================================================
			{Code: `<label {...{htmlFor: "js_id"}} htmlFor="" aria-label="Save" />`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}}},

			// ============================================================
			// Spread + direct, direct FIRST. getProp source-order returns
			// direct → "x" → truthy.
			// ============================================================
			{Code: `<label htmlFor="x" {...{htmlFor: ""}} aria-label="Save" />`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}}},

			// ============================================================
			// Direct + literal-spread, both truthy, direct first. getProp
			// returns direct, value "x" → truthy.
			// ============================================================
			{Code: `<label htmlFor="x" {...{htmlFor: "y"}} aria-label="Save" />`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}}},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// validateHtmlFor: explicit empty string is falsy.
			// ============================================================
			{Code: `<label htmlFor="">Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedHtmlFor}},

			// ============================================================
			// jsx-ast-utils `hasProp` default `spreadStrict: true`: even a
			// literal-object spread carrying `htmlFor` does NOT satisfy the
			// presence check. `<label {...{htmlFor: "js_id"}} />` → upstream
			// hasProp returns false → validateHtmlFor returns false → htmlFor
			// mode reports.
			// ============================================================
			{Code: `<label {...{htmlFor: "js_id"}}>Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedHtmlFor}},

			// ============================================================
			// Same strict-spread semantic for non-literal spread:
			// `{...props}` is opaque and does NOT count.
			// ============================================================
			{Code: `<label {...props}>Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedHtmlFor}},

			// ============================================================
			// Spread + direct, spread FIRST with falsy value, direct AFTER
			// with truthy value. hasProp(strict)=true via direct; getProp's
			// `find` scans source-order, finds the spread first, returns the
			// inner empty string → falsy → reports htmlFor.
			// (Counter-intuitive but mirrors jsx-ast-utils' two-helper
			// asymmetry. The direct `htmlFor="x"` is ignored.)
			// ============================================================
			{Code: `<label {...{htmlFor: ""}} htmlFor="x">Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedHtmlFor}},

			// ============================================================
			// Spread + direct, both falsy: regardless of order, getProp's
			// first-match returns one of them, and both are "" → falsy.
			// ============================================================
			{Code: `<label htmlFor="" {...{htmlFor: ""}}>Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedHtmlFor}},

			// ============================================================
			// Unknown `assert` enum → normalized to 'either' (verified via
			// `npx eslint` on upstream main: schema rejects the invalid
			// option then re-invokes the rule with defaults; observable
			// behavior is `either` mode). `<label>Save</label>` has neither
			// htmlFor nor nested control → reports `either`.
			// ============================================================
			{Code: `<label>Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "garbage"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedEither}},
			{Code: `<label>Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": ""}},
				Errors: []rule_tester.InvalidTestCaseError{expectedEither}},

			// ============================================================
			// validateHtmlFor: `htmlFor={undefined}` resolves to undefined,
			// not truthy under getPropValue.
			// ============================================================
			{Code: `<label htmlFor={undefined}>Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedHtmlFor}},

			// ============================================================
			// validateHtmlFor: `htmlFor={null}` — `getPropValue` returns
			// null → falsy → not a valid htmlFor.
			// ============================================================
			{Code: `<label htmlFor={null}>Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedHtmlFor}},

			// ============================================================
			// validateHtmlFor: `htmlFor={false}` — explicit boolean false
			// → falsy.
			// ============================================================
			{Code: `<label htmlFor={false}>Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedHtmlFor}},

			// ============================================================
			// Listener gate: `<label>` with whitespace-only text fails the
			// accessible-label check (trim → "" → falsy).
			// ============================================================
			{Code: `<label htmlFor="x">   </label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedAccessibleLabel}},

			// ============================================================
			// htmlFor missing in either mode, but text present → `either`
			// error.
			// ============================================================
			{Code: `<label>Save</label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "either"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedEither}},

			// ============================================================
			// Nested control deeper than depth → not detected.
			// `<label htmlFor=...>label<span><span><input /></span></span></label>`
			// at depth=2 (default) — span at depth 1, span at depth 2,
			// input at depth 3 (> maxDepth=2) → not seen → both mode fails
			// on nesting.
			// ============================================================
			{Code: `<label htmlFor="x">label<span><span><input /></span></span></label>`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "both"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedBoth}},

			// ============================================================
			// Empty `<label />` with no attributes — accessibleLabel error
			// (root labelling-prop check returns false; no descendants).
			// ============================================================
			{Code: `<label />`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "either"}},
				Errors: []rule_tester.InvalidTestCaseError{expectedAccessibleLabel}},

			// ============================================================
			// Settings: attributes.for replaces default — `htmlFor=…` no
			// longer counts when settings.for is ['for']. Provide
			// htmlFor="js_id" + accessibleLabel; htmlFor mode still fails
			// because the lookup walks `['for']`, finds nothing, returns
			// false.
			// ============================================================
			{Code: `<label htmlFor="js_id" aria-label="Save" />`, Tsx: true, Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"attributes": map[string]interface{}{"for": []interface{}{"for"}}}},
				Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedHtmlFor}},

			// ============================================================
			// Position assertion: report points at the OPENING element of
			// the paired label, not the JsxElement root. The opening
			// element starts at column 1.
			// ============================================================
			{Code: `<label htmlFor="x" />`, Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "either"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "accessibleLabel",
					Message:   msgAccessibleLabel,
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 22,
				}}},

			// Multi-line position assertion.
			{Code: "\n<label\n  htmlFor=\"x\"\n/>",
				Tsx: true, Options: []interface{}{map[string]interface{}{"assert": "either"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "accessibleLabel",
					Message:   msgAccessibleLabel,
					Line:      2,
					Column:    1,
					EndLine:   4,
					EndColumn: 3,
				}}},
		})
}

// TestLabelHasAssociatedControl_OptionParsing locks in the JSON-decoded option
// shape — both single-element CLI (`['warn', {...}]` → bare map) and array-
// wrapped rule_tester (`[{...}]` → []interface{}{...}). Catches regressions
// where GetOptionsMap stops handling either shape.
func TestLabelHasAssociatedControl_OptionParsing(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule,
		[]rule_tester.ValidTestCase{
			// Bare map options (single-element CLI shape).
			{Code: `<label htmlFor="x" aria-label="x" />`, Tsx: true,
				Options: map[string]interface{}{"assert": "htmlFor"}},
			// Array-wrapped options (rule_tester / multi-element shape).
			{Code: `<label htmlFor="x" aria-label="x" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}}},
			// Nil options — defaults apply (assert='either').
			{Code: `<label>Save<input /></label>`, Tsx: true},
			// Empty options array — defaults apply.
			{Code: `<label>Save<input /></label>`, Tsx: true,
				Options: []interface{}{}},
		},
		[]rule_tester.InvalidTestCase{
			// Bare map options with assert: 'both' fires `both`.
			{Code: `<label htmlFor="x">Save</label>`, Tsx: true,
				Options: map[string]interface{}{"assert": "both"},
				Errors:  []rule_tester.InvalidTestCaseError{expectedBoth}},
			// Array-wrapped options with assert: 'both' fires `both`.
			{Code: `<label htmlFor="x">Save</label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"assert": "both"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedBoth}},
		})
}

// TestLabelHasAssociatedControl_HtmlForExpressionForms locks in
// `validateHtmlFor` against the full surface of `getPropValue`'s
// truthiness extractor. Every row covers one expression kind the
// upstream `extractValueFromExpression` table dispatches on.
func TestLabelHasAssociatedControl_HtmlForExpressionForms(t *testing.T) {
	withAssertHtmlFor := func(code string) rule_tester.ValidTestCase {
		return rule_tester.ValidTestCase{Code: code, Tsx: true,
			Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}}}
	}
	withAssertHtmlForInvalid := func(code string, err rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
		return rule_tester.InvalidTestCase{Code: code, Tsx: true,
			Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}},
			Errors:  []rule_tester.InvalidTestCaseError{err}}
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule,
		[]rule_tester.ValidTestCase{
			// --------- Identifier (non-undefined) → synthesized non-empty string → truthy
			withAssertHtmlFor(`<label htmlFor={id} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={someLongName} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={$id} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={_id} aria-label="x" />`),

			// --------- MemberExpression / OptionalMemberExpression → synthesized → truthy
			withAssertHtmlFor(`<label htmlFor={form.id} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={form.field.id} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={form?.id} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={obj["key"]} aria-label="x" />`),

			// --------- CallExpression / OptionalCallExpression → synthesized → truthy
			withAssertHtmlFor(`<label htmlFor={getId()} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={getId(arg)} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={getId?.()} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={form.getId()} aria-label="x" />`),

			// --------- ConditionalExpression — both branches truthy → truthy
			withAssertHtmlFor(`<label htmlFor={cond ? "a" : "b"} aria-label="x" />`),
			// ConditionalExpression — non-empty Identifier branches synth-truthy
			withAssertHtmlFor(`<label htmlFor={cond ? a : b} aria-label="x" />`),

			// --------- LogicalExpression (||, &&, ??) — synthesized truthy
			withAssertHtmlFor(`<label htmlFor={id || "fallback"} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={id ?? "fallback"} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={id && id.value} aria-label="x" />`),

			// --------- TemplateExpression (with substitution) → synth non-empty string
			withAssertHtmlFor(`<label htmlFor={` + "`id-${x}`" + `} aria-label="x" />`),
			// NoSubstitutionTemplateLiteral, non-empty
			withAssertHtmlFor(`<label htmlFor={` + "`id`" + `} aria-label="x" />`),

			// --------- BinaryExpression (string concat) → "ab" → truthy
			withAssertHtmlFor(`<label htmlFor={"a" + "b"} aria-label="x" />`),
			// BinaryExpression numeric — 1+2=3 → truthy
			withAssertHtmlFor(`<label htmlFor={1 + 2} aria-label="x" />`),

			// --------- NumericLiteral truthy (non-zero)
			withAssertHtmlFor(`<label htmlFor={1} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={42} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={-1} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={1.5} aria-label="x" />`),

			// --------- BooleanLiteral true → truthy
			withAssertHtmlFor(`<label htmlFor={true} aria-label="x" />`),

			// --------- TS wrappers transparent — all truthy via unwrap loop
			withAssertHtmlFor(`<label htmlFor={(id)} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={((id))} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={id as string} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={(id as any)} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={id!} aria-label="x" />`),
			withAssertHtmlFor(`<label htmlFor={(id as string)!} aria-label="x" />`),

			// --------- Direct StringLiteral with HTML entities — non-empty after decode
			withAssertHtmlFor(`<label htmlFor="js&#95;id" aria-label="x" />`),
			// Direct StringLiteral with multiple ids (HTML technically only allows one,
			// but the rule only checks truthiness)
			withAssertHtmlFor(`<label htmlFor="id1 id2" aria-label="x" />`),
		},
		[]rule_tester.InvalidTestCase{
			// --------- NumericLiteral 0 → falsy → invalid
			withAssertHtmlForInvalid(`<label htmlFor={0}>Save</label>`, expectedHtmlFor),
			// NumericLiteral -0 → falsy
			withAssertHtmlForInvalid(`<label htmlFor={-0}>Save</label>`, expectedHtmlFor),

			// --------- BooleanLiteral false → falsy
			// (already covered by existing extras, kept here for grouping)
			withAssertHtmlForInvalid(`<label htmlFor={false}>Save</label>`, expectedHtmlFor),

			// --------- Direct StringLiteral whitespace — non-empty string → truthy
			// (Counter-intuitive but matches upstream's `!!getPropValue` — no trim
			// applied at validateHtmlFor's level. Documented to lock in.)
			// We test this via the *opposite* expectation: even " " is truthy.
			// — moved to valid block above? Actually upstream `getPropValue("  ")` is "  "
			// which is truthy. So this is VALID, not invalid. Skip the case here.

			// --------- Empty NoSubstitutionTemplateLiteral → "" → falsy
			withAssertHtmlForInvalid(`<label htmlFor={`+"``"+`}>Save</label>`, expectedHtmlFor),

			// --------- Empty JsxExpression `{}` — upstream's "type not in TYPES" → null → falsy
			withAssertHtmlForInvalid(`<label htmlFor={}>Save</label>`, expectedHtmlFor),

			// --------- ConditionalExpression — both branches falsy string literals
			withAssertHtmlForInvalid(`<label htmlFor={cond ? "" : ""}>Save</label>`, expectedHtmlFor),
		})
}

// TestLabelHasAssociatedControl_NestedTraversal locks in mayHaveAccessibleLabel
// and mayContainChildComponent against tsgo-specific child-traversal shapes:
// JsxFragment transparency, conditional rendering, map rendering, nested
// labels, and the exact depth boundary.
func TestLabelHasAssociatedControl_NestedTraversal(t *testing.T) {
	assertEither := []interface{}{map[string]interface{}{"assert": "either"}}
	assertBoth := []interface{}{map[string]interface{}{"assert": "both"}}
	assertNesting := []interface{}{map[string]interface{}{"assert": "nesting"}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// JsxFragment as direct child — Fragment is walked transparently
			// (GetJsxChildren returns its children); `<input/>` is reached
			// at depth 2 (Fragment is depth 1).
			// ============================================================
			{Code: `<label>Save<><input /></></label>`, Tsx: true, Options: assertNesting},
			// Fragment wrapping the text only
			{Code: `<label><>Save</>1<input /></label>`, Tsx: true, Options: assertNesting},
			// Both children inside Fragment
			{Code: `<label><>Save<input /></></label>`, Tsx: true, Options: assertNesting},

			// ============================================================
			// Conditional rendering: && / ternary inside JsxExpression —
			// the JsxExpression child short-circuits mayContainChildComponent
			// to true and mayHaveAccessibleLabel to true.
			// ============================================================
			{Code: `<label>Save{cond && <input />}</label>`, Tsx: true, Options: assertNesting},
			{Code: `<label>Save{cond ? <input /> : <textarea />}</label>`, Tsx: true, Options: assertNesting},
			{Code: `<label>{labelText}<input /></label>`, Tsx: true, Options: assertEither},

			// ============================================================
			// map rendering: returned <input/> is inside JsxExpression →
			// short-circuit.
			// ============================================================
			{Code: `<label>Save{items.map(i => <input key={i.id} />)}</label>`, Tsx: true, Options: assertNesting},

			// ============================================================
			// Nested labels: BOTH labels are independently checked. Outer
			// has text + nested input + nested label; Inner has text +
			// input. Both valid in nesting mode.
			// ============================================================
			{Code: `<label>Outer<label>Inner<input /></label></label>`, Tsx: true, Options: assertNesting},
			// Outer label has just a nested label that contains input — Outer
			// finds nested input at depth 2; Inner finds input at depth 1.
			{Code: `<label>Outer<label>x<input/></label></label>`, Tsx: true, Options: assertNesting},
			// Outer has no own text but mayHaveAccessibleLabel recurses into
			// inner's text at depth 2 → accessibleLabel satisfied; outer's
			// nested-control scan also finds the inner input at depth 2 →
			// either mode passes for outer; inner trivially passes.
			{Code: `<label><label>x<input /></label></label>`, Tsx: true, Options: assertEither},

			// ============================================================
			// Depth boundary — input EXACTLY at maxDepth still detected.
			// Default depth=2: input at depth 2 (label→span→input) → found.
			// ============================================================
			{Code: `<label>Save<span><input /></span></label>`, Tsx: true, Options: assertNesting},

			// ============================================================
			// Multi-line, mixed whitespace — tsgo's JsxText / JsxTextAllWhiteSpaces
			// split. Non-whitespace text still satisfies accessible-label.
			// ============================================================
			{Code: "<label>\n  Save\n  <input />\n</label>", Tsx: true, Options: assertNesting},

			// ============================================================
			// `<label>{}<input/></label>` — empty JsxExpression child counts
			// as accessible-label under upstream's unconditional return-true.
			// ============================================================
			{Code: `<label>{}<input /></label>`, Tsx: true, Options: assertNesting},

			// ============================================================
			// JSX comment-only child: `<label>{/* x */}<input/></label>` —
			// tsgo emits a JsxExpression with no Expression; mayHaveAccessibleLabel
			// treats it as accessible (matches upstream).
			// ============================================================
			{Code: `<label>{/* placeholder */}<input /></label>`, Tsx: true, Options: assertNesting},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Depth boundary — input one level BEYOND maxDepth not detected.
			// Default depth=2: input at depth 3 (label→span→span→input).
			// ============================================================
			{Code: `<label htmlFor="x">Save<span><span><input /></span></span></label>`, Tsx: true, Options: assertBoth,
				Errors: []rule_tester.InvalidTestCaseError{expectedBoth}},

			// ============================================================
			// Whitespace-only text + Fragment-wrapped input — text trim →
			// falsy → mayHaveAccessibleLabel returns false → accessibleLabel.
			// ============================================================
			{Code: `<label>   <></></label>`, Tsx: true, Options: assertNesting,
				Errors: []rule_tester.InvalidTestCaseError{expectedAccessibleLabel}},

			// ============================================================
			// Two empty labels at the top level — both fire `accessibleLabel`
			// independently. Locks in that the listener visits every JsxElement
			// (not just the outermost).
			// ============================================================
			{Code: `<><label /><label /></>`, Tsx: true, Options: assertEither,
				Errors: []rule_tester.InvalidTestCaseError{expectedAccessibleLabel, expectedAccessibleLabel}},

			// ============================================================
			// Outer label has only text (no nested input, no htmlFor) → fires
			// `either`. Sanity check that the listener doesn't get confused by
			// an inner non-label child.
			// ============================================================
			{Code: `<label>outerText<span>innerText</span></label>`, Tsx: true, Options: assertEither,
				Errors: []rule_tester.InvalidTestCaseError{expectedEither}},
		})
}

// TestLabelHasAssociatedControl_TagShapes locks in the listener gate against
// the full surface of JSX tag-name shapes — uppercase variant, member
// expression, namespaced names, and minimatch case-sensitivity.
func TestLabelHasAssociatedControl_TagShapes(t *testing.T) {
	assertEither := []interface{}{map[string]interface{}{"assert": "either"}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// minimatch is CASE-SENSITIVE: `<LABEL>` (uppercase) is NOT a
			// label match — it's a React component named LABEL. Rule skips
			// regardless of content.
			// ============================================================
			{Code: `<LABEL />`, Tsx: true, Options: assertEither},
			{Code: `<LABEL>Save</LABEL>`, Tsx: true, Options: assertEither},

			// ============================================================
			// Member-expression tag: getElementType returns "Foo.Label" —
			// minimatch("Foo.Label", "label") is false (literal "." in
			// pattern would need escaping). Rule skips.
			// ============================================================
			{Code: `<Foo.Label htmlFor="x" />`, Tsx: true, Options: assertEither},
			{Code: `<Foo.Bar.Label />`, Tsx: true, Options: assertEither},

			// ============================================================
			// PropertyAccess CAN be targeted explicitly via labelComponents:
			// `'Foo.Label'` matches exactly.
			// ============================================================
			{Code: `<Foo.Label htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"Foo.Label"}, "assert": "either"}}},

			// ============================================================
			// Namespaced JSX tag: `svg:label` — full type "svg:label" does
			// NOT match minimatch "label" exact (colon is literal).
			// ============================================================
			{Code: `<svg:label />`, Tsx: true, Options: assertEither},
			{Code: `<svg:label>Save</svg:label>`, Tsx: true, Options: assertEither},

			// ============================================================
			// `<Label />` (capitalized) does NOT match 'label' minimatch
			// either — case-sensitive. Custom React component; skipped.
			// ============================================================
			{Code: `<Label />`, Tsx: true, Options: assertEither},

			// ============================================================
			// Unicode-aware first-character classification in
			// mayHaveAccessibleLabel's React-component fallback. Mirrors
			// JS's `c === c.toUpperCase()`:
			//
			//   - Uppercase Latin / CJK / `$` / `_` → React component →
			//     fallback returns true → label is "accessible".
			//
			// Empty `<label>` containing one of these triggers the fallback
			// → accessibleLabel passes → only the assert error fires (here
			// either's `<label><Foo/></label>` gets `either`, not
			// `accessibleLabel`). These are the *invalid* cases (no htmlFor,
			// no nested control); see invalid block below.
			//
			// For the VALID side, putting htmlFor on satisfies either mode.
			{Code: `<label htmlFor="x"><ÉFoo /></label>`, Tsx: true, Options: assertEither},
			{Code: `<label htmlFor="x"><中Foo /></label>`, Tsx: true, Options: assertEither},
			{Code: `<label htmlFor="x"><$Foo /></label>`, Tsx: true, Options: assertEither},
			{Code: `<label htmlFor="x"><_Foo /></label>`, Tsx: true, Options: assertEither},
		},
		[]rule_tester.InvalidTestCase{
			// PropertyAccess WITH a matching labelComponents entry — fires.
			{Code: `<Foo.Label>Save</Foo.Label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"Foo.Label"}, "assert": "either"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedEither}},

			// ============================================================
			// Unicode-aware lowercase classification — upstream verified
			// with `npx eslint` against eslint-plugin-jsx-a11y@main:
			//
			//   <label><éFoo/></label>  → accessibleLabel
			//   <label><àFoo/></label>  → accessibleLabel
			//
			// `'é'.toUpperCase() === 'É' !== 'é'` → not a React component →
			// fallback does NOT trigger → mayHaveAccessibleLabel returns
			// false for the empty label → accessibleLabel fires.
			// ============================================================
			{Code: `<label><éFoo /></label>`, Tsx: true, Options: assertEither,
				Errors: []rule_tester.InvalidTestCaseError{expectedAccessibleLabel}},
			{Code: `<label><àFoo /></label>`, Tsx: true, Options: assertEither,
				Errors: []rule_tester.InvalidTestCaseError{expectedAccessibleLabel}},

			// ============================================================
			// Uppercase non-ASCII / non-cased characters DO trigger the
			// React-component fallback — mayHaveAccessibleLabel returns
			// true, then the missing htmlFor / nested-control reports the
			// `either` error.
			// ============================================================
			{Code: `<label><ÉFoo /></label>`, Tsx: true, Options: assertEither,
				Errors: []rule_tester.InvalidTestCaseError{expectedEither}},
			{Code: `<label><中Foo /></label>`, Tsx: true, Options: assertEither,
				Errors: []rule_tester.InvalidTestCaseError{expectedEither}},
			{Code: `<label><$Foo /></label>`, Tsx: true, Options: assertEither,
				Errors: []rule_tester.InvalidTestCaseError{expectedEither}},
			{Code: `<label><_Foo /></label>`, Tsx: true, Options: assertEither,
				Errors: []rule_tester.InvalidTestCaseError{expectedEither}},
		})
}

// TestLabelHasAssociatedControl_GlobPatterns locks in the full minimatch
// glob surface against the listener gate (labelComponents) and the inner
// helpers (controlComponents).
func TestLabelHasAssociatedControl_GlobPatterns(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule,
		[]rule_tester.ValidTestCase{
			// --------- Star glob: prefix
			{Code: `<MyLabel htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"My*"}, "assert": "either"}}},
			// --------- Star glob: suffix
			{Code: `<FormLabel htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"*Label"}, "assert": "either"}}},
			// --------- Star glob: middle
			{Code: `<MyCustomLabel htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"My*Label"}, "assert": "either"}}},

			// --------- ? glob: single character
			{Code: `<XLabel htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"?Label"}, "assert": "either"}}},
			// --------- ??? glob: exactly 3 chars
			{Code: `<FooLabel htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"???Label"}, "assert": "either"}}},

			// --------- Brace expansion: {a,b}
			{Code: `<AlphaLabel htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"{Alpha,Beta}Label"}, "assert": "either"}}},
			{Code: `<BetaLabel htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"{Alpha,Beta}Label"}, "assert": "either"}}},

			// --------- Character class: [Ll]abel — but only if "label"
			// minimatch matches the BUILTIN, since 'label' is always added.
			// Use a more demonstrative custom name.
			{Code: `<ALabel htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"[AB]Label"}, "assert": "either"}}},

			// --------- Negation (!Foo): every name EXCEPT Foo matches.
			// `<label>` matches `!Foo` (since "label" != "Foo"); the rule
			// still triggers and finds htmlFor + aria-label.
			{Code: `<label htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"!Foo"}, "assert": "either"}}},

			// --------- Multiple patterns: any-match wins. `<MyInput/>` matches
			// `*Input` but not `*Label`; rule triggers and finds htmlFor +
			// aria-label.
			{Code: `<MyInput htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"*Label", "*Input"}, "assert": "either"}}},
			// Same multi-pattern, this time the OTHER alternative matches.
			{Code: `<MyLabel htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"*Label", "*Input"}, "assert": "either"}}},

			// --------- controlComponents glob inside mayContainChildComponent
			{Code: `<label>Save<MyInput /></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"controlComponents": []interface{}{"My*"}, "assert": "nesting"}}},
		},
		[]rule_tester.InvalidTestCase{
			// Glob doesn't match → rule doesn't trigger → no report.
			// Negative shape: this CASE WOULD FIRE if glob matched. Construct
			// a sample where the glob fails: labelComponents=['ZZZ*'],
			// `<MyLabel/>` doesn't match → rule skipped → valid. To produce
			// an invalid case we use the builtin 'label' (always added).

			// --------- Glob in labelComponents adds patterns, but builtin
			// 'label' still always matches `<label>`.
			{Code: `<label>Save</label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"WontMatchAnything"}, "assert": "either"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedEither}},

			// --------- controlComponents glob match — `<MyInput/>` matches
			// `My*`, but `<OtherInput/>` does not → nested control NOT
			// detected → nesting mode fails.
			{Code: `<label>Save<OtherInput /></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"controlComponents": []interface{}{"My*"}, "assert": "nesting"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedNesting}},
		})
}

// TestLabelHasAssociatedControl_SettingsInteractions locks in the
// `settings['jsx-a11y']` interactions: components map, polymorphicPropName,
// polymorphicAllowList, and attributes.for (replace-not-merge default).
func TestLabelHasAssociatedControl_SettingsInteractions(t *testing.T) {
	polymorphicSettings := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"polymorphicPropName": "as",
		},
	}
	polymorphicAllowListSettings := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"polymorphicPropName":  "as",
			"polymorphicAllowList": []interface{}{"Box"},
		},
	}
	componentsPlusAttributes := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"components": map[string]interface{}{"MyLabel": "label"},
			"attributes": map[string]interface{}{"for": []interface{}{"htmlFor", "for"}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule,
		[]rule_tester.ValidTestCase{
			// --------- polymorphicPropName: `<Box as="label">…</Box>` resolves
			// element type to "label" and triggers the rule.
			{Code: `<Box as="label" htmlFor="x" aria-label="y" />`, Tsx: true,
				Settings: polymorphicSettings,
				Options:  []interface{}{map[string]interface{}{"assert": "either"}}},

			// --------- polymorphicAllowList: only `Box` may be remapped via
			// `as`. `<Other as="label" />` keeps type "Other" → not a label.
			{Code: `<Other as="label" />`, Tsx: true,
				Settings: polymorphicAllowListSettings,
				Options:  []interface{}{map[string]interface{}{"assert": "either"}}},

			// --------- polymorphic + Box in allow list → remapped → triggers
			{Code: `<Box as="label" htmlFor="x" aria-label="y" />`, Tsx: true,
				Settings: polymorphicAllowListSettings,
				Options:  []interface{}{map[string]interface{}{"assert": "either"}}},

			// --------- components + attributes.for joint: MyLabel→label AND
			// `for` attribute is accepted.
			{Code: `<MyLabel for="js_id" aria-label="y" />`, Tsx: true,
				Settings: componentsPlusAttributes,
				Options:  []interface{}{map[string]interface{}{"assert": "htmlFor"}}},

			// --------- attributes.for INCLUDING htmlFor — both attributes are
			// honored, first present wins. `htmlFor` first in list, found
			// first.
			{Code: `<label htmlFor="js_id" for="" aria-label="y" />`, Tsx: true,
				Settings: componentsPlusAttributes,
				Options:  []interface{}{map[string]interface{}{"assert": "htmlFor"}}},
		},
		[]rule_tester.InvalidTestCase{
			// --------- polymorphic remap to label, but no htmlFor → either fails.
			{Code: `<Box as="label">Save</Box>`, Tsx: true,
				Settings: polymorphicSettings,
				Options:  []interface{}{map[string]interface{}{"assert": "either"}},
				Errors:   []rule_tester.InvalidTestCaseError{expectedEither}},

			// --------- attributes.for first-attribute-wins: htmlFor="" earlier
			// in user list → wins lookup → falsy → reports despite `for="x"`.
			{Code: `<label htmlFor="" for="x" aria-label="y" />`, Tsx: true,
				Settings: componentsPlusAttributes,
				Options:  []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors:   []rule_tester.InvalidTestCaseError{expectedHtmlFor}},

			// --------- components remap MyLabel→label; no htmlFor → either.
			{Code: `<MyLabel>Save</MyLabel>`, Tsx: true,
				Settings: componentsPlusAttributes,
				Options:  []interface{}{map[string]interface{}{"assert": "either"}},
				Errors:   []rule_tester.InvalidTestCaseError{expectedEither}},
		})
}

// TestLabelHasAssociatedControl_DepthBoundaries locks in the depth-cap
// semantics: clamp to 25, lower-bound passthrough, and exact boundary
// detection for both mayHaveAccessibleLabel (depth 0 entry) and
// mayContainChildComponent (depth 1 entry).
func TestLabelHasAssociatedControl_DepthBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule,
		[]rule_tester.ValidTestCase{
			// --------- depth=1: input at depth 1 detected
			{Code: `<label>Save<input /></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"depth": float64(1), "assert": "nesting"}}},

			// --------- depth=0: mayHaveAccessibleLabel inspects root (aria-label)
			// and mayContainChildComponent bails before any child — but
			// `<label aria-label="x" />` has no descendants, so nesting is
			// trivially false; in `htmlFor` mode we have htmlFor → valid.
			{Code: `<label htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"depth": float64(0), "assert": "htmlFor"}}},

			// --------- depth=25 (cap): deep nested input found
			{Code: `<label htmlFor="x" aria-label="y"><span><span><span><span><span><span><span><span><input /></span></span></span></span></span></span></span></span></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"depth": float64(25), "assert": "both"}}},
		},
		[]rule_tester.InvalidTestCase{
			// --------- depth=1, input at depth 2 → NOT found
			{Code: `<label htmlFor="x" aria-label="y"><span><input /></span></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"depth": float64(1), "assert": "both"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedBoth}},

			// --------- depth=0: nested input never visited
			{Code: `<label htmlFor="x" aria-label="y"><input /></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"depth": float64(0), "assert": "both"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedBoth}},

			// --------- Negative depth: mayHaveAccessibleLabel's `depth > maxDepth`
			// bail fires at the ROOT (depth=0 > maxDepth=-1 is true), so even the
			// root labelling-prop check is skipped → hasAccessibleLabel=false →
			// accessibleLabel reports BEFORE the assert switch.
			{Code: `<label htmlFor="x" aria-label="y"><input /></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"depth": float64(-1), "assert": "both"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedAccessibleLabel}},

			// --------- Same root-bail for depth=-5: aria-label on root is NOT
			// inspected — `0 > -5` is true, so mayHaveAccessibleLabel returns
			// false immediately. Mirrors upstream's `Math.min(depth, 25)` (no
			// lower-bound clamp) + `depth > maxDepth` bail.
			{Code: `<label htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"depth": float64(-5), "assert": "htmlFor"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedAccessibleLabel}},

			// --------- depth=2 (default): input at depth 3 not found
			{Code: `<label htmlFor="x">Save<span><span><input /></span></span></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"assert": "both"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedBoth}},
		})
}

// TestLabelHasAssociatedControl_RealWorldPatterns covers the JSX shapes that
// actually show up in production codebases — render props, HOCs, conditional
// rendering at the call site, fragment-wrapped sibling groups, etc.
func TestLabelHasAssociatedControl_RealWorldPatterns(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// react-hook-form `register` pattern — label wraps text + input
			// receiving a spread of register's return value.
			// ============================================================
			{Code: `<label>Email<input {...register("email")} /></label>`, Tsx: true},

			// ============================================================
			// Common UI library custom label component — only fires
			// when `labelComponents` is configured.
			// ============================================================
			{Code: `<FormLabel htmlFor="email">Email</FormLabel>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"FormLabel"}, "assert": "htmlFor"}}},
			{Code: `<InputLabel htmlFor="my-input">Name</InputLabel>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"InputLabel"}, "assert": "htmlFor"}}},

			// ============================================================
			// `<label>` inside a controlled form section — text + input
			// is the canonical pattern.
			// ============================================================
			{Code: `<form><label>Name<input type="text" /></label></form>`, Tsx: true},

			// ============================================================
			// `<label>` inside React.Fragment with siblings — fragment is
			// not a label, but the inner label is checked independently.
			// ============================================================
			{Code: `<React.Fragment><label>Name<input /></label></React.Fragment>`, Tsx: true},

			// ============================================================
			// `<label>` rendered conditionally — wrapped in JsxExpression,
			// the label child fires the rule.
			// ============================================================
			{Code: `<div>{showLabel ? <label>Name<input /></label> : null}</div>`, Tsx: true},

			// ============================================================
			// `<label>` rendered in a list — each label fires the rule.
			// ============================================================
			{Code: `<>{items.map(item => <label key={item.id} htmlFor={item.id}>{item.label}<input id={item.id} /></label>)}</>`, Tsx: true},

			// ============================================================
			// Forwarded children pattern (HOC) — spread carries htmlFor
			// opaquely; spread is enough to satisfy nesting mode's
			// accessibleLabel via HasLabellingProp; nesting child is a
			// concrete `<input/>`.
			// ============================================================
			{Code: `const MyLabel = (props) => <label {...props}>{props.children}<input /></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"assert": "nesting"}}},

			// ============================================================
			// Suspense / Boundary patterns
			// ============================================================
			{Code: `<Suspense fallback={<Spinner />}><label htmlFor="x" aria-label="y" /></Suspense>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}}},
			{Code: `<ErrorBoundary><label>Save<input /></label></ErrorBoundary>`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// `<label>` rendered conditionally with no input — fires the
			// rule from inside the JsxExpression.
			// ============================================================
			{Code: `<div>{showLabel && <label>Name</label>}</div>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"assert": "either"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedEither}},

			// ============================================================
			// `<label>` in a list with no association — fires for each.
			// Two labels in source → two reports.
			// ============================================================
			{Code: `<><label>A</label><label>B</label></>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"assert": "either"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedEither, expectedEither}},

			// ============================================================
			// Custom label component without configuring `labelComponents`
			// — rule doesn't fire on `<FormLabel/>` (not a `label` tag).
			// But the inner `<label/>` (if rendered) would. Sanity: an
			// empty inner label fires accessibleLabel.
			// ============================================================
			{Code: `<FormLabel><label /></FormLabel>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"assert": "either"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedAccessibleLabel}},
		})
}

// TestLabelHasAssociatedControl_DefensiveOptions exercises options-parsing
// against malformed inputs (non-string values where strings expected,
// mixed-type arrays, deeply nested structures). The rule must not panic
// and must fall back to defaults gracefully.
func TestLabelHasAssociatedControl_DefensiveOptions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LabelHasAssociatedControlRule,
		[]rule_tester.ValidTestCase{
			// --------- assert: non-string falls back to default 'either'
			{Code: `<label>Save<input /></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"assert": float64(42)}}},

			// --------- depth: string instead of number → ignored → default 2
			{Code: `<label>Save<input /></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"depth": "not-a-number", "assert": "nesting"}}},

			// --------- depth: nil → use default
			{Code: `<label>Save<input /></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"depth": nil, "assert": "nesting"}}},

			// --------- labelComponents: array with mixed types → string entries kept
			{Code: `<MyLabel htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{"MyLabel", float64(123), nil, true}, "assert": "htmlFor"}}},

			// --------- labelComponents: not an array → ignored → only builtin 'label'
			{Code: `<label htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": "not-array", "assert": "htmlFor"}}},

			// --------- empty labelComponents array — builtin 'label' still applies
			{Code: `<label htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"labelComponents": []interface{}{}, "assert": "htmlFor"}}},

			// --------- empty controlComponents array — only builtins (input, …) apply
			{Code: `<label>Save<input /></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"controlComponents": []interface{}{}, "assert": "nesting"}}},

			// --------- settings entirely absent — defaults flow through
			{Code: `<label htmlFor="x" aria-label="y" />`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"assert": "htmlFor"}}},

			// --------- jsx-a11y settings is wrong shape (string) → fall back
			{Code: `<label htmlFor="x" aria-label="y" />`, Tsx: true,
				Settings: map[string]interface{}{"jsx-a11y": "bogus"},
				Options:  []interface{}{map[string]interface{}{"assert": "htmlFor"}}},

			// --------- attributes.for is wrong shape (object instead of array)
			// → fall back to default ['htmlFor']
			{Code: `<label htmlFor="x" aria-label="y" />`, Tsx: true,
				Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"attributes": map[string]interface{}{"for": map[string]interface{}{"nope": true}}}},
				Options:  []interface{}{map[string]interface{}{"assert": "htmlFor"}}},

			// --------- attributes.for with mixed-type entries — only string entries kept
			{Code: `<label htmlFor="x" aria-label="y" />`, Tsx: true,
				Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"attributes": map[string]interface{}{"for": []interface{}{"htmlFor", float64(1), nil}}}},
				Options:  []interface{}{map[string]interface{}{"assert": "htmlFor"}}},
		},
		[]rule_tester.InvalidTestCase{
			// --------- attributes.for explicit empty array — htmlFor lookup
			// walks no attributes → returns false → htmlFor mode fails.
			{Code: `<label htmlFor="x" aria-label="y" />`, Tsx: true,
				Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"attributes": map[string]interface{}{"for": []interface{}{}}}},
				Options:  []interface{}{map[string]interface{}{"assert": "htmlFor"}},
				Errors:   []rule_tester.InvalidTestCaseError{expectedHtmlFor}},

			// --------- depth: nil with deeply nested input → falls back to
			// default 2 → input at depth 3 not found → both mode fails.
			{Code: `<label htmlFor="x">label<span><span><input /></span></span></label>`, Tsx: true,
				Options: []interface{}{map[string]interface{}{"depth": nil, "assert": "both"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedBoth}},
		})
}
