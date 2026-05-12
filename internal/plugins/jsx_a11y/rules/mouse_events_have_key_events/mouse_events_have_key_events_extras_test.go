package mouse_events_have_key_events

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// This file holds test cases NOT in upstream's own test file. Upstream-parity
// cases live in mouse_events_have_key_events_upstream_test.go so it stays
// trivially comparable against future upstream updates via diff.
//
// Coverage axes:
//
//   - Dimension 1 (AST node shape): tsgo's parenthesized / `as` /
//     `satisfies` / `!` wrappers on hover-handler values; JsxSelfClosingElement
//     vs paired JsxElement boundary; literal-spread vs non-literal-spread
//     for both the hover handler and the pair attribute.
//   - Dimension 2 (scoping / nesting): nested JSX where listener must
//     classify each element independently — outer hover-paired, inner
//     missing pair, and vice versa.
//   - Dimension 4 (universal edge shapes): attribute existence forms
//     (boolean, `={null}`, `={undefined}`), case-insensitive matching
//     on both hover-handler names and the `onFocus` / `onBlur` pair,
//     multi-line opening-element position anchoring, multi-attribute
//     same-element reports.
//   - Upstream branch lock-in: hoverInHandlers `.find(handler => …)`
//     short-circuits on the FIRST handler with a non-null value — the
//     subsequent handler is irrelevant to the report; settings-style
//     misuse (`settings.components`) is intentionally not honored per
//     upstream's `node.name.name` raw read.

// extras-only error templates — reused across the file.
var (
	pointerEnterErrorAtCol6 = rule_tester.InvalidTestCaseError{
		MessageId: "mouseOver",
		Message:   "onPointerEnter must be accompanied by onFocus for accessibility.",
		Line:      1,
		Column:    6,
	}
	pointerLeaveErrorAtCol6 = rule_tester.InvalidTestCaseError{
		MessageId: "mouseOut",
		Message:   "onPointerLeave must be accompanied by onBlur for accessibility.",
		Line:      1,
		Column:    6,
	}
	// {mouseOver,mouseOut}ErrorAnyPos — the standard diagnostic without
	// fixed Line/Column. Use when the JSX opening element is buried inside
	// a wrapper (forwardRef / memo / HOC / map / class / IIFE / async /
	// generator / conditional / fragment) where the column offset is
	// brittle and the test's value is "fires at all", not "fires exactly
	// at column N". The Line/Column == 0 fields are skipped by the
	// rule_tester (see internal/rule_tester/rule_tester.go assertions).
	mouseOverErrorAnyPos = rule_tester.InvalidTestCaseError{
		MessageId: "mouseOver",
		Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
	}
	mouseOutErrorAnyPos = rule_tester.InvalidTestCaseError{
		MessageId: "mouseOut",
		Message:   "onMouseOut must be accompanied by onBlur for accessibility.",
	}
)

// TestMouseEventsHaveKeyEventsExtras locks in branches that upstream's
// test file doesn't exercise but are reachable through the rule's
// listener gate.
func TestMouseEventsHaveKeyEventsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MouseEventsHaveKeyEventsRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Dimension 1: hover-handler value forms that count as nullish
			// (propValue == null) — rule short-circuits before checking
			// the pair attribute.
			// ============================================================

			// ---- onMouseOver={null} — extracted value is null → not
			//      counted as a hover-in handler → no pair check. ----
			{Code: `<div onMouseOver={null} />`, Tsx: true},
			// ---- onMouseOver={undefined} — extracted undefined → skipped. ----
			{Code: `<div onMouseOver={undefined} />`, Tsx: true},
			// ---- onMouseOut={null} / {undefined} — same on the hover-out side. ----
			{Code: `<div onMouseOut={null} />`, Tsx: true},
			{Code: `<div onMouseOut={undefined} />`, Tsx: true},

			// ---- Parens + `as` are unwrapped by staticEval (skipTransparent
			//      strips parens + TS assertions); `!` (NonNullExpression) is
			//      NOT — upstream jsx-ast-utils' TSNonNullExpression extractor
			//      returns the stringified `"<inner>!"` form which is non-null
			//      non-undefined truthy, so `<X attr={null!} />` does NOT
			//      classify as nullish. The `null!` / `undefined!` cases move
			//      to invalid below. ----
			{Code: `<div onMouseOver={(null)} />`, Tsx: true},
			{Code: `<div onMouseOver={null as any} />`, Tsx: true},
			// ---- Pair value `null!` — upstream jsx-ast-utils' TSNonNullExpression
			//      extractor returns the stringified `"null!"` (non-null
			//      non-undefined truthy), so the pair is counted as present
			//      and no diagnostic fires. Differential-validated against
			//      eslint-plugin-jsx-a11y@6.10.2 — this is the corner case
			//      that #915's jsxa11yutil fix aligned to upstream. ----
			{Code: `<div onMouseOver={fn} onFocus={null!} />`, Tsx: true},
			// ---- Pair value `undefined!` — same shape: string `"undefined!"`
			//      is truthy non-null → pair counted as present → valid. ----
			{Code: `<div onMouseOver={fn} onFocus={undefined!} />`, Tsx: true},

			// ============================================================
			// Dimension 1: paired-form `<div ...></div>` — tsgo emits
			// KindJsxOpeningElement; rule fires on the OpeningElement
			// regardless of whether `</div>` follows.
			// ============================================================

			// ---- Paired form, properly paired handlers. ----
			{Code: `<div onMouseOver={() => void 0} onFocus={() => void 0}>label</div>`, Tsx: true},
			{Code: `<div onMouseOut={() => void 0} onBlur={() => void 0}>label</div>`, Tsx: true},

			// ============================================================
			// Dimension 1: case-insensitive attribute matching (upstream's
			// getProp default `ignoreCase: true`). Lock both sides.
			// ============================================================

			// ---- All-lowercase hover + pair — both match case-insensitively. ----
			{Code: `<div onmouseover={() => {}} onfocus={() => {}} />`, Tsx: true},
			{Code: `<div onmouseout={() => {}} onblur={() => {}} />`, Tsx: true},
			// ---- ALL-CAPS hover + pair. ----
			{Code: `<div ONMOUSEOVER={() => {}} ONFOCUS={() => {}} />`, Tsx: true},
			// ---- Mixed case: hover in canonical case, pair in lowercase. ----
			{Code: `<div onMouseOver={() => {}} onfocus={() => {}} />`, Tsx: true},
			{Code: `<div onMouseOut={() => {}} onblur={() => {}} />`, Tsx: true},

			// ============================================================
			// Dimension 1: TS-wrapped non-null pair value — staticEval
			// unwraps so the pair still resolves to a non-null value.
			// ============================================================

			// ---- onFocus={(fn)} — paren wrapper. ----
			{Code: `<div onMouseOver={fn} onFocus={(fn)} />`, Tsx: true},
			// ---- onFocus={fn as any} — TS as cast. ----
			{Code: `<div onMouseOver={fn} onFocus={fn as any} />`, Tsx: true},
			// ---- onFocus={fn!} — TS non-null assertion. ----
			{Code: `<div onMouseOver={fn} onFocus={fn!} />`, Tsx: true},

			// ============================================================
			// Element-type classification: non-DOM tag shapes skip the rule
			// before any pairing check. Member-expression and namespaced
			// tag names stringify to a form NOT in aria-query's `dom` map.
			// ============================================================

			// ---- `<Foo.Bar onMouseOver={fn} />` — "Foo.Bar" not in dom map. ----
			{Code: `<Foo.Bar onMouseOver={() => {}} />`, Tsx: true},
			// ---- `<Foo.Bar.Baz onMouseOut={fn} />` — same, deeper chain. ----
			{Code: `<Foo.Bar.Baz onMouseOut={() => {}} />`, Tsx: true},
			// ---- `<this.Foo onMouseOver={fn} />` — "this.Foo" not in dom map. ----
			{Code: `<this.Foo onMouseOver={() => {}} />`, Tsx: true},
			// ---- Lowercase member-expression also escapes — even though
			//      it would be a DOM-component classification, the dotted
			//      string is not in aria-query's `dom`. ----
			{Code: `<foo.bar onMouseOver={() => {}} />`, Tsx: true},
			// ---- Namespaced tag — "svg:circle" not in aria-query's `dom`. ----
			{Code: `<svg:circle onMouseOver={() => {}} />`, Tsx: true},

			// ============================================================
			// Upstream NEVER respects settings.components / polymorphicPropName
			// for this rule (it reads `node.name.name` directly). Lock the
			// raw-name behavior — `<Footer>` with components map should
			// NOT trigger even when the map would resolve to `footer`.
			// ============================================================

			// ---- settings.components map ignored — rule reads raw name. ----
			{
				Code: `<Footer onMouseOver={() => {}} />`,
				Tsx:  true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"components": map[string]interface{}{
							"Footer": "footer",
						},
					},
				},
			},
			// ---- settings.polymorphicPropName ignored too. ----
			{
				Code: `<Foo as="div" onMouseOver={() => {}} />`,
				Tsx:  true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"polymorphicPropName": "as",
					},
				},
			},

			// ============================================================
			// Nesting: pair-check applies independently to each element.
			// ============================================================

			// ---- Outer + inner BOTH have pairs. ----
			{
				Code: `<div onMouseOver={fn} onFocus={fn}>` +
					`<span onMouseOver={fn} onFocus={fn} />` +
					`</div>`,
				Tsx: true,
			},

			// ============================================================
			// Multiple hover-in handlers in custom options — find returns
			// the FIRST one with a non-null value. Add an onFocus and the
			// rule passes regardless of which one was found first.
			// ============================================================
			{
				Code: `<div onMouseOver={fn} onMouseEnter={fn} onFocus={fn} />`,
				Tsx:  true,
				Options: []interface{}{map[string]interface{}{
					"hoverInHandlers": []interface{}{"onMouseOver", "onMouseEnter"},
				}},
			},
			// ---- First listed handler is null, second has a value — only
			//      the second triggers the pair check, which onFocus satisfies. ----
			{
				Code: `<div onMouseOver={null} onMouseEnter={fn} onFocus={fn} />`,
				Tsx:  true,
				Options: []interface{}{map[string]interface{}{
					"hoverInHandlers": []interface{}{"onMouseOver", "onMouseEnter"},
				}},
			},

			// ============================================================
			// Real-world component patterns — properly paired across
			// HOC / forwardRef / map / hooks shapes.
			// ============================================================

			// ---- forwardRef with paired handlers. ----
			{
				Code: `const Btn = React.forwardRef((props, ref) => <div ref={ref} onMouseOver={props.onHover} onFocus={props.onFocus} />);`,
				Tsx:  true,
			},
			// ---- map() over items, each item paired. ----
			{
				Code: `const list = items.map(item => <li key={item.id} onMouseOver={() => hover(item)} onFocus={() => focus(item)}>{item.label}</li>);`,
				Tsx:  true,
			},
			// ---- async function returning paired element. ----
			{
				Code: `async function render() { return <div onMouseOver={fn} onFocus={fn} />; }`,
				Tsx:  true,
			},
			// ---- Conditional rendering: both branches paired. ----
			{
				Code: `const x = cond ? <div onMouseOver={fn} onFocus={fn} /> : <span onMouseOut={fn} onBlur={fn} />;`,
				Tsx:  true,
			},

			// ============================================================
			// staticEval branches: `||`, `&&`, `??`, ternary on hover-handler
			// value. Statically-nullish results short-circuit before pair
			// check; statically-truthy results trigger pair check normally.
			// ============================================================

			// ---- `null || null` → null → nullish → no pair check needed. ----
			{Code: `<div onMouseOver={null || null} />`, Tsx: true},
			// ---- `null ?? undefined` → undefined → nullish → skip. ----
			{Code: `<div onMouseOut={null ?? undefined} />`, Tsx: true},
			// ---- `false ? f : null` → static-false → use whenFalse `null`
			//      → nullish → skip. Locks ConditionalExpression branching. ----
			{Code: `<div onMouseOver={false ? f : null} />`, Tsx: true},
			// ---- `false && fn` → false (LHS falsy short-circuits) →
			//      coerced-falsy not equal to null exactly, BUT staticEval
			//      returns the boolean false, which is `false != null` →
			//      counted as PRESENT. Pair must be present too. With
			//      paired onFocus, valid. ----
			{
				Code: `<div onMouseOver={false && fn} onFocus={fn} />`,
				Tsx:  true,
			},

			// ============================================================
			// Boolean attribute form on the PAIR. `<div ... onFocus />`
			// extracts to JS `true` (extractValue's null-attribute path) →
			// pair is counted as present → no report.
			// ============================================================

			// ---- Boolean-form hover + boolean-form pair. ----
			{Code: `<div onMouseOver onFocus />`, Tsx: true},
			{Code: `<div onMouseOut onBlur />`, Tsx: true},
			// ---- Mixed: explicit hover, boolean pair. ----
			{Code: `<div onMouseOver={fn} onFocus />`, Tsx: true},
			{Code: `<div onMouseOut={fn} onBlur />`, Tsx: true},

			// ============================================================
			// Non-aria-query DOM elements skip the rule entirely. SVG tags
			// (`<svg>`, `<path>`, `<circle>`, …) are NOT in aria-query's
			// `dom` map — even though they're real DOM elements at runtime,
			// upstream's `dom.get('svg')` returns undefined → skip.
			// ============================================================

			{Code: `<svg onMouseOver={() => {}} />`, Tsx: true},
			{Code: `<path onMouseOver={() => {}} />`, Tsx: true},
			{Code: `<circle onMouseOver={() => {}} />`, Tsx: true},

			// ============================================================
			// Options shape: bare-map form (CLI / single-option config).
			// PORT_RULE.md requires exercising the JSON path on both shapes.
			// utils.GetOptionsMap accepts either array-wrapped or bare map;
			// the linter unwraps single-element option arrays to bare map
			// before calling Run, so the CLI shape is what real users hit.
			// ============================================================

			// ---- Bare-map options (CLI shape) with custom hoverIn. ----
			{
				Code:    `<div onMouseEnter={() => {}} onFocus={() => {}} />`,
				Tsx:     true,
				Options: map[string]interface{}{"hoverInHandlers": []interface{}{"onMouseEnter"}},
			},
			// ---- Bare-map options with custom hoverOut. ----
			{
				Code:    `<div onMouseLeave={() => {}} onBlur={() => {}} />`,
				Tsx:     true,
				Options: map[string]interface{}{"hoverOutHandlers": []interface{}{"onMouseLeave"}},
			},
			// ---- Bare-map empty object → no keys → fall back to defaults
			//      → `<div onMouseOver={fn} onFocus={fn} />` is paired. ----
			{
				Code:    `<div onMouseOver={fn} onFocus={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{},
			},
			// ---- Array-wrapped empty object → same as bare empty map →
			//      defaults apply. ----
			{
				Code:    `<div onMouseOver={fn} onFocus={fn} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{}},
			},
			// ---- Empty option array → falls back to all defaults. ----
			{
				Code:    `<div onMouseOver={fn} onFocus={fn} />`,
				Tsx:     true,
				Options: []interface{}{},
			},
			// ---- Malformed options: non-string entries are filtered by
			//      StringSliceOption. Remaining ["onMouseOver"] becomes
			//      the effective hoverIn list. ----
			{
				Code:    `<div onMouseOver={fn} onFocus={fn} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverInHandlers": []interface{}{1, true, "onMouseOver"}}},
			},
			// ---- Malformed hoverInHandlers — not an array → StringSliceOption
			//      returns nil → defaults apply → `<div onMouseOver={fn}
			//      onFocus={fn} />` paired. ----
			{
				Code:    `<div onMouseOver={fn} onFocus={fn} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverInHandlers": "onMouseOver"}},
			},

			// ============================================================
			// ShorthandPropertyAssignment in literal spread — `getProp`
			// walks literal spreads (default spreadStrict: false). The
			// shorthand `{...{onFocus}}` matches just like
			// `{...{onFocus: onFocus}}`.
			// ============================================================

			// ---- Pair via shorthand spread. ----
			{Code: `<div onMouseOver={fn} {...{onFocus}} />`, Tsx: true},
			// ---- Hover via shorthand spread, pair via direct attribute. ----
			{Code: `<div {...{onMouseOver}} onFocus={fn} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Dimension 1: boolean attribute form — extractValue's null-attr
			// path yields JS true → `true != null` → counted as present
			// hover handler → pair check fires → missing → reports.
			// ============================================================

			// ---- `<div onMouseOver />` — boolean form, no onFocus. ----
			{
				Code:   `<div onMouseOver />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			// ---- `<div onMouseOut />` boolean form, no onBlur. ----
			{
				Code:   `<div onMouseOut />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOutErrorAtCol6},
			},
			// ---- Boolean hover + boolean pair: pair extractValue = true
			//      → != null → pair COUNTED → no report. (Validates that
			//      `<div onMouseOver onFocus />` is the boolean-pairing
			//      escape hatch.) ----

			// ============================================================
			// Dimension 1: pair attribute set to a nullish value — counted
			// as missing pair → reports. Mirrors upstream's `null ||
			// undefined` arm.
			// ============================================================

			// ---- onFocus={null} — same effect as `onFocus={undefined}`. ----
			{
				Code:   `<div onMouseOver={fn} onFocus={null} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			// ---- onBlur={null}. ----
			{
				Code:   `<div onMouseOut={fn} onBlur={null} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOutErrorAtCol6},
			},

			// ============================================================
			// Dimension 1: case-insensitive — uppercase / lowercase handler
			// matches against the default `onMouseOver` / `onMouseOut`.
			// ============================================================

			// ---- All-lowercase handler, no pair. ----
			{
				Code:   `<div onmouseover={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			{
				Code:   `<div onmouseout={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOutErrorAtCol6},
			},
			// ---- ALL-CAPS handler with no pair. ----
			{
				Code: `<div ONMOUSEOVER={() => {}} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
					Line:      1,
					Column:    6,
				}},
			},

			// ============================================================
			// Both hover-in and hover-out missing on the same element with
			// default options — both reports fire. Defaults only check
			// onMouseOver / onMouseOut, so we exercise the canonical pair.
			// ============================================================

			{
				Code: `<div onMouseOver={() => {}} onMouseOut={() => {}} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "mouseOver",
						Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
						Line:      1,
						Column:    6,
					},
					{
						MessageId: "mouseOut",
						Message:   "onMouseOut must be accompanied by onBlur for accessibility.",
						Line:      1,
						Column:    29,
					},
				},
			},

			// ============================================================
			// Multi-line opening element — Line/Column anchor to the
			// hover-handler attribute, NOT the element start.
			// ============================================================

			{
				Code: `<div` + "\n" +
					`  onMouseOver={() => void 0}` + "\n" +
					`/>;`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
					Line:      2,
					Column:    3,
				}},
			},

			// ============================================================
			// Nested JSX — listener visits inner element independently.
			// Outer paired (valid), inner missing pair (invalid).
			// ============================================================

			{
				Code: `<div onMouseOver={fn} onFocus={fn}>` + "\n" +
					`  <span onMouseOver={fn} />` + "\n" +
					`</div>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
					Line:      2,
					Column:    9,
				}},
			},

			// ============================================================
			// Custom hoverIn list with multiple entries — `find` returns
			// the FIRST handler whose value is non-null. The report carries
			// the FIRST matched name. Locks the iteration-order guarantee.
			// ============================================================

			{
				Code: `<div onMouseOver={fn} onMouseEnter={fn} />`,
				Tsx:  true,
				Options: []interface{}{map[string]interface{}{
					"hoverInHandlers": []interface{}{"onMouseOver", "onMouseEnter"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
					Line:      1,
					Column:    6,
				}},
			},
			// ---- First handler is null → skipped; second handler has
			//      value → reported under the SECOND name. ----
			{
				Code: `<div onMouseOver={null} onMouseEnter={fn} />`,
				Tsx:  true,
				Options: []interface{}{map[string]interface{}{
					"hoverInHandlers": []interface{}{"onMouseOver", "onMouseEnter"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseEnter must be accompanied by onFocus for accessibility.",
					Line:      1,
					Column:    25,
				}},
			},

			// ============================================================
			// Pointer-event variant lock-in via custom options.
			// ============================================================

			{
				Code: `<div onPointerEnter={() => void 0} />`,
				Tsx:  true,
				Options: []interface{}{map[string]interface{}{
					"hoverInHandlers": []interface{}{"onPointerEnter"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{pointerEnterErrorAtCol6},
			},
			{
				Code: `<div onPointerLeave={() => void 0} />`,
				Tsx:  true,
				Options: []interface{}{map[string]interface{}{
					"hoverOutHandlers": []interface{}{"onPointerLeave"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{pointerLeaveErrorAtCol6},
			},

			// ============================================================
			// Two-element same-line fragment — each fires independently.
			// ============================================================

			{
				Code: `<><div onMouseOver={fn} /><span onMouseOut={fn} /></>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "mouseOver",
						Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
						Line:      1,
						Column:    8,
					},
					{
						MessageId: "mouseOut",
						Message:   "onMouseOut must be accompanied by onBlur for accessibility.",
						Line:      1,
						Column:    33,
					},
				},
			},

			// ============================================================
			// Deep nesting (3+ levels): listener visits every JsxOpening /
			// JsxSelfClosing element regardless of depth. Outer paired
			// (no report), middle unpaired hover-in (reports), inner
			// unpaired hover-out (reports). Locks the per-element
			// classification independence.
			// ============================================================

			{
				Code: `<section onMouseOver={fn} onFocus={fn}>` + "\n" +
					`  <div onMouseOver={fn}>` + "\n" +
					`    <span onMouseOut={fn}>label</span>` + "\n" +
					`  </div>` + "\n" +
					`</section>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "mouseOver",
						Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
						Line:      2,
						Column:    8,
					},
					{
						MessageId: "mouseOut",
						Message:   "onMouseOut must be accompanied by onBlur for accessibility.",
						Line:      3,
						Column:    11,
					},
				},
			},
			// ---- 4-level nesting, all DOM, mixed valid / invalid. ----
			{
				Code: `<main onMouseOver={fn} onFocus={fn}>` + "\n" +
					`  <section onMouseOver={fn} onFocus={fn}>` + "\n" +
					`    <article onMouseOver={fn}>` + "\n" +
					`      <p onMouseOut={fn}>x</p>` + "\n" +
					`    </article>` + "\n" +
					`  </section>` + "\n" +
					`</main>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "mouseOver",
						Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
						Line:      3,
						Column:    14,
					},
					{
						MessageId: "mouseOut",
						Message:   "onMouseOut must be accompanied by onBlur for accessibility.",
						Line:      4,
						Column:    10,
					},
				},
			},

			// ============================================================
			// Real-world component patterns (reports). Position is brittle
			// inside wrapper code, so mouseOverErrorAnyPos / mouseOutErrorAnyPos
			// skip Line/Column — the assertion locks "fires at all",
			// "carries the right MessageId", and "carries the right Message
			// text", which is the contract for an a11y rule embedded inside
			// arbitrary surrounding code.
			// ============================================================

			// ---- React.forwardRef wrapping a non-paired div. The
			//      inner <div onMouseOver={...} /> must still be classified
			//      as a DOM element with a missing onFocus. ----
			{
				Code:   `const Btn = React.forwardRef((props, ref) => <div ref={ref} onMouseOver={props.onHover} />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAnyPos},
			},
			// ---- React.memo wrapping a non-paired section. ----
			{
				Code:   `const Pane = React.memo(({ id, onHover }) => <section id={id} onMouseOver={onHover} />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAnyPos},
			},
			// ---- Custom HOC wrapping a non-paired div. ----
			{
				Code:   `const Enhanced = withTracking(({ value, onHover }) => <div data-value={value} onMouseOver={onHover} />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAnyPos},
			},
			// ---- Array.map producing children with missing onFocus. ----
			{
				Code:   `const list = items.map(item => <li key={item.id} onMouseOver={() => hover(item)}>{item.label}</li>);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAnyPos},
			},
			// ---- Class component render() returning non-paired root. ----
			{
				Code:   `class Card extends React.Component { render() { return <article onMouseOver={this.onHover}>x</article>; } }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAnyPos},
			},
			// ---- Generator yielding non-paired elements — both fire. ----
			{
				Code:   `function* render() { yield <div onMouseOver={fn} />; yield <span onMouseOut={fn} />; }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAnyPos, mouseOutErrorAnyPos},
			},
			// ---- Async function returning non-paired root. ----
			{
				Code:   `async function render() { return <div onMouseOver={fn} />; }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAnyPos},
			},
			// ---- IIFE returning non-paired root. ----
			{
				Code:   `const x = (() => <article onMouseOver={fn} />)();`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAnyPos},
			},
			// ---- Conditional with one non-paired branch. The valid branch
			//      stays silent; only the missing-onFocus branch reports. ----
			{
				Code:   `const x = cond ? <div onMouseOver={fn} /> : null;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAnyPos},
			},
			// ---- `&&` short-circuit producing non-paired element. ----
			{
				Code:   `const x = cond && <div onMouseOver={fn} />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAnyPos},
			},
			// ---- Fragment with mixed valid / invalid children — only
			//      non-paired children report. <button> is DOM and still
			//      requires pairing (this rule has no "interactive element"
			//      exemption — different from click-events-have-key-events). ----
			{
				Code:   `const x = (<><button onMouseOver={fn} onFocus={fn} /><div onMouseOver={fn} /><a href="x" onMouseOut={fn} onBlur={fn} /><section onMouseOver={fn} /></>);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAnyPos, mouseOverErrorAnyPos},
			},
			// ---- Two-function file: A reports at line 1, B at line 2 —
			//      Line is stable (each top-level function on its own line),
			//      Column is not (depends on inner whitespace). ----
			{
				Code: `function A() { return <div onMouseOver={fn} />; }` + "\n" +
					`function B() { return <section onMouseOut={fn} />; }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "mouseOver",
						Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
						Line:      1,
					},
					{
						MessageId: "mouseOut",
						Message:   "onMouseOut must be accompanied by onBlur for accessibility.",
						Line:      2,
					},
				},
			},

			// ============================================================
			// staticEval branches: statically-truthy hover value → pair
			// check fires; statically-nullish pair value → counted as
			// missing.
			// ============================================================

			// ---- `fn || null` — LHS truthy → uses LHS fn → not nullish →
			//      hover counted; pair (onFocus) missing → report. ----
			{
				Code:   `<div onMouseOver={fn || null} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			// ---- `someVar ?? fallback` — both Identifiers,
			//      `someVar` resolves to a string ("someVar") → not nullish
			//      → use LHS → counted; pair missing → report. ----
			{
				Code:   `<div onMouseOver={someVar ?? fallback} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			// ---- `true ? fn : null` — static-true → use whenTrue fn →
			//      counted; pair missing → report. ----
			{
				Code:   `<div onMouseOver={true ? fn : null} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			// ---- Hover handler set, pair statically resolves to null
			//      via ternary → pair counted as missing → report. ----
			{
				Code:   `<div onMouseOver={fn} onFocus={false ? f : null} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			// ---- Pair value `undefined as any` — `as` strips → undefined
			//      → nullish → missing → report. ----
			{
				Code:   `<div onMouseOver={fn} onFocus={undefined as any} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			// ---- Hover value `undefined!` — upstream jsx-ast-utils returns
			//      string `"undefined!"` (truthy, non-null) → counted as
			//      present hover handler → pair check fires → onFocus
			//      missing → report. Differential-validated against
			//      eslint-plugin-jsx-a11y@6.10.2. ----
			{
				Code:   `<div onMouseOver={undefined!} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			// ---- Hover value `null!` — same shape: string `"null!"` is
			//      truthy non-null → counted as present hover handler →
			//      onFocus missing → report. ----
			{
				Code:   `<div onMouseOver={null!} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},

			// ============================================================
			// Bare-map options (CLI shape) — must produce the same
			// diagnostics as the array-wrapped form.
			// ============================================================

			// ---- Bare-map custom hoverIn: missing onFocus → report. ----
			{
				Code:    `<div onMouseEnter={() => {}} />`,
				Tsx:     true,
				Options: map[string]interface{}{"hoverInHandlers": []interface{}{"onMouseEnter"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseEnter must be accompanied by onFocus for accessibility.",
					Line:      1,
					Column:    6,
				}},
			},
			// ---- Bare-map custom hoverOut: missing onBlur → report. ----
			{
				Code:    `<div onMouseLeave={() => {}} />`,
				Tsx:     true,
				Options: map[string]interface{}{"hoverOutHandlers": []interface{}{"onMouseLeave"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOut",
					Message:   "onMouseLeave must be accompanied by onBlur for accessibility.",
					Line:      1,
					Column:    6,
				}},
			},

			// ============================================================
			// Literal-spread containing a nullish pair value → pair is
			// counted as missing (PropValueIsNullish sees null/undefined
			// via the PropertyAssignment.Initializer staticEval).
			// ============================================================

			// ---- Hover direct, pair via literal spread with null. ----
			{
				Code:   `<div onMouseOver={fn} {...{onFocus: null}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			// ---- Hover direct, pair via literal spread with undefined. ----
			{
				Code:   `<div onMouseOut={fn} {...{onBlur: undefined}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOutErrorAtCol6},
			},
			// ---- Hover via literal spread, pair absent. ----
			{
				Code: `<div {...{onMouseOver: () => {}}} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
					// Position not pinned — the report node is the inner
					// PropertyAssignment inside the spread; tsgo column
					// differs from upstream's synthesized JSXAttribute.
				}},
			},

			// ============================================================
			// Spread + direct attribute order — spread before / after /
			// multiple — does not affect the rule's classification.
			// ============================================================

			// ---- Spread before hover, no pair → report. ----
			{
				Code:   `<div {...props} onMouseOver={fn} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
					Line:      1,
					Column:    17,
				}},
			},
			// ---- Multiple spreads + direct hover, no pair → report. ----
			{
				Code:   `<div {...a} {...b} onMouseOver={fn} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
					Line:      1,
					Column:    20,
				}},
			},

			// ============================================================
			// Self-closing void elements with hover handlers — the rule
			// still applies (any DOM tag, any pairing miss → report).
			// ============================================================

			// ---- `<br>` — self-closing-only void DOM element. ----
			{
				Code: `<br onMouseOver={fn} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
					Line:      1,
					Column:    5,
				}},
			},
			// ---- `<input>` — `<input>` IS in aria-query's dom map and
			//      this rule does NOT exempt interactive elements (unlike
			//      click-events-have-key-events). So even `<input type="text">`
			//      with onMouseOver and no onFocus reports. ----
			{
				Code: `<input type="text" onMouseOver={fn} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
					Line:      1,
					Column:    20,
				}},
			},
			// ---- `<button>` — same: button gets no exemption from this
			//      rule (in contrast to click-events-have-key-events,
			//      where button is interactive → exempt). ----
			{
				Code: `<button onMouseOver={fn}>label</button>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
					Line:      1,
					Column:    9,
				}},
			},
		},
	)
}
