package control_has_associated_label

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedErrorAnyPos is the standard expected diagnostic without a fixed
// line/column. Use when the JSX opening element is not at column 1 — e.g.
// inside a `function X() { return <button /> }` wrapper or `arr.map(...)`.
var expectedErrorAnyPos = rule_tester.InvalidTestCaseError{
	MessageId: "controlHasAssociatedLabel",
	Message:   errorMessage,
}

// polymorphicSettings exercises `settings['jsx-a11y'].polymorphicPropName`.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// componentsMapSettings demotes named components to specific tags via the
// settings.components map. `MyButton` becomes `button` (interactive),
// `MyDiv` becomes `div` (non-interactive — won't trigger).
var componentsMapSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"MyButton": "button",
			"MyDiv":    "div",
		},
	},
}

// TestControlHasAssociatedLabelExtras locks in branches that upstream's
// test file does not exercise but are reachable through the rule's listener
// gate. Coverage axes:
//
//   - Dimension 1 (tsgo AST shape): parenthesized / `as` / `satisfies` /
//     non-null wrappers on attribute values; self-closing vs paired forms
//     on every branch (root, recursion, React-component fallback).
//   - Dimension 1 (literal shapes): JsxExpression-with-StringLiteral
//     vs direct StringLiteral; NoSubstitutionTemplateLiteral; TemplateExpression
//     with substitutions (non-literal under literalPropValue).
//   - Dimension 2 (nesting / containers): JSX inside React.forwardRef /
//     React.memo / HOC / hooks / class render / map / fragment / generator /
//     async / IIFE. Each opening element classified independently.
//   - Dimension 4 (universal edge shapes):
//   - Spread attribute opacity (`{...x}` / `{...{...}}` count as
//     labelling under upstream's `attribute.type !== 'JSXAttribute'`
//     short-circuit).
//   - controlComponents glob matching INSIDE mayHaveAccessibleLabel
//     (different from the top-level exact match).
//   - settings.components demoting custom → DOM removes the React-component
//     fallback and re-enables interactive checks.
//   - depth cap at 25 — supplying depth=100 behaves identically to
//     depth=25.
//   - Listener gate branches:
//   - hard-coded `link` ignore (cannot be disabled via ignoreElements
//     absent or set to []).
//   - ignoreRoles via literal vs non-literal role expression.
//   - aria-hidden via TS-wrapper / JsxExpression / case-insensitive
//     string "true".
//   - empty `<X />` root + uppercase tag + not in controlComponents →
//     fallback applies at root depth 0.
func TestControlHasAssociatedLabelExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ControlHasAssociatedLabelRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Listener gate: hard-coded `link` ignore. Always exempt
			// regardless of ignoreElements / options.
			// ============================================================
			{Code: `<link />`, Tsx: true},
			{Code: `<link />`, Tsx: true, Options: map[string]interface{}{
				"ignoreElements": []interface{}{}, // explicit empty
			}},

			// ============================================================
			// Listener gate: ignoreRoles literal-only match. A non-literal
			// `role={someVar}` does NOT match ignoreRoles (literalPropValue
			// returns null for Identifier) — so this falls through to the
			// trigger / label check.
			// ============================================================
			// Literal role match → skip.
			{Code: `<div role="grid" />`, Tsx: true, Options: map[string]interface{}{
				"ignoreRoles": []interface{}{"grid"},
			}},
			// JsxExpression-wrapped StringLiteral role → still extractable
			// via literalPropValue → match → skip.
			{Code: `<div role={"grid"} />`, Tsx: true, Options: map[string]interface{}{
				"ignoreRoles": []interface{}{"grid"},
			}},
			// NoSubstitutionTemplateLiteral role → extractable → match.
			{Code: "<div role={`grid`} />", Tsx: true, Options: map[string]interface{}{
				"ignoreRoles": []interface{}{"grid"},
			}},

			// ============================================================
			// Listener gate: isHiddenFromScreenReader paths.
			// ============================================================
			// aria-hidden boolean form — boolean attribute → true.
			{Code: `<button aria-hidden />`, Tsx: true},
			// aria-hidden={true} — direct boolean.
			{Code: `<button aria-hidden={true} />`, Tsx: true},
			// aria-hidden="true" — string coerced to bool true.
			{Code: `<button aria-hidden="true" />`, Tsx: true},
			// aria-hidden with paren wrapping (single + multi level).
			{Code: `<button aria-hidden={(true)} />`, Tsx: true},
			{Code: `<button aria-hidden={((true))} />`, Tsx: true},
			// aria-hidden with TS `as` cast — staticEval strips TSAsExpression.
			{Code: `<button aria-hidden={true as boolean} />`, Tsx: true},
			// <input type="hidden"> case-insensitive.
			{Code: `<input type="hidden" />`, Tsx: true,
				Options: map[string]interface{}{
					// drop input from ignoreElements so we exercise the
					// IsHiddenFromScreenReader branch, not the early ignore.
					"ignoreElements": []interface{}{},
				}},
			{Code: `<input type="HIDDEN" />`, Tsx: true,
				Options: map[string]interface{}{"ignoreElements": []interface{}{}}},
			{Code: `<input type="Hidden" />`, Tsx: true,
				Options: map[string]interface{}{"ignoreElements": []interface{}{}}},

			// ============================================================
			// hasLabellingProp: spread attributes are opaque and count
			// as labelling. Even literal-resolvable spreads.
			// ============================================================
			{Code: `<button {...props} />`, Tsx: true},
			{Code: `<button {...{title: 'x'}} />`, Tsx: true},

			// ============================================================
			// hasLabellingProp: trim semantics — whitespace-only values are
			// treated as missing labels (upstream `tryTrim`). Each case
			// pairs with a non-empty sibling that DOES count, to lock both
			// arms of the conditional.
			// ============================================================
			// Non-empty label string.
			{Code: `<button aria-label="Save" />`, Tsx: true},
			// Boolean form on aria-label — counts as label (upstream null-attr
			// → true → tryTrim non-string passes through truthy).
			{Code: `<button aria-label />`, Tsx: true},
			// JsxExpression with non-whitespace string.
			{Code: `<button aria-label={"Save"} />`, Tsx: true},
			// ---- aria-label={x!} (TS non-null assertion). Upstream's
			//      `TSNonNullExpression` extractor stringifies and appends
			//      "!" → "x!" (non-empty, !!tryTrim non-empty) → counts
			//      as labelled. rslint's staticEval emits the same
			//      stringified form via the dedicated KindNonNullExpression
			//      arm → also counts as labelled → no report. ALIGNED. ----
			{Code: `<button aria-label={x!} />`, Tsx: true},
			// Same on aria-labelledby and alt (shared labellingValueIsPresent
			// path; same staticEval emission).
			{Code: `<button aria-labelledby={ref!} />`, Tsx: true},
			{Code: `<button><img alt={src!} /></button>`, Tsx: true},

			// ============================================================
			// labelAttributes: user-configured extra label attribute names.
			// ============================================================
			{Code: `<button label="Save" />`, Tsx: true,
				Options: map[string]interface{}{"labelAttributes": []interface{}{"label"}}},

			// ============================================================
			// JsxExpression child — assumed to render label regardless of
			// expression content (even {undefined} / {null}).
			// ============================================================
			{Code: `<button>{maybeLabel}</button>`, Tsx: true},
			{Code: `<button>{undefined}</button>`, Tsx: true},
			{Code: `<button>{null}</button>`, Tsx: true},
			{Code: `<button>{0}</button>`, Tsx: true},

			// ============================================================
			// Recursion: JsxFragment as a transparent child container.
			// Upstream's checkElement walks `node.children` regardless of
			// node type — fragments don't have a switch arm but the
			// children loop still picks up their kids.
			// ============================================================
			{Code: `<button><>Save</></button>`, Tsx: true},

			// ============================================================
			// React-component fallback inside recursion: an uppercase-named
			// self-closing component (not in controlComponents) is assumed
			// to render a label.
			// ============================================================
			{Code: `<button><MyLabel /></button>`, Tsx: true},
			{Code: `<button><MyLabel></MyLabel></button>`, Tsx: true},

			// ============================================================
			// React-component fallback: minimatch in controlComponents.
			// `MyControl*` glob matches `MyControlA` → fallback NOT
			// triggered → falls through; but since this case has a label,
			// it stays valid. Pairs with the invalid version below.
			// ============================================================
			{Code: `<button><MyControlA aria-label="Save" /></button>`, Tsx: true,
				Options: map[string]interface{}{"controlComponents": []interface{}{"MyControl*"}}},

			// ============================================================
			// settings.polymorphicPropName: <Foo as="button">Save</Foo>
			// becomes button → interactive → check label → "Save" present.
			// ============================================================
			{Code: `<Foo as="button">Save</Foo>`, Tsx: true, Settings: polymorphicSettings},

			// ============================================================
			// settings.components: <MyButton aria-label="Save" /> demotes
			// to button (interactive) → label found.
			// ============================================================
			{Code: `<MyButton aria-label="Save" />`, Tsx: true, Settings: componentsMapSettings},

			// ============================================================
			// depth cap: depth=100 is clamped to 25, but a 3-level
			// label is still found. (Pure regression: don't accidentally
			// invert the cap.)
			// ============================================================
			{Code: `<button><span><span>Save</span></span></button>`, Tsx: true,
				Options: map[string]interface{}{"depth": float64(100)}},

			// ============================================================
			// Real-world component patterns.
			// ============================================================
			// React.forwardRef wrapping interactive button with label.
			{Code: `const Btn = React.forwardRef((props, ref) => <button ref={ref} aria-label="Save" />);`, Tsx: true},
			// Array.map producing labelled interactive children.
			{Code: `const list = items.map(item => <button key={item.id} aria-label={item.label} />);`, Tsx: true},
			// Class component render() with labelled root.
			{Code: `class Form extends React.Component { render() { return <button>Save</button>; } }`, Tsx: true},

			// ============================================================
			// Complex nesting: outer non-control wrapping interactive
			// child with label → only inner is classified; outer doesn't
			// require a label.
			// ============================================================
			{Code: `<div><button>Save</button></div>`, Tsx: true},
			{Code: `<section><a href="#" aria-label="Home" /></section>`, Tsx: true},

			// ============================================================
			// Nested interactive elements — both interactive, each
			// independently labelled.
			// ============================================================
			{Code: `<button aria-label="outer"><a href="#" aria-label="inner" /></button>`, Tsx: true},

			// ============================================================
			// Multi-level fragments + label deep within (counts fragments
			// in the depth budget).
			// ============================================================
			// Fragment + div + span + text: depth needs to reach JsxText.
			// button(0) -> fragment(1) -> JsxText(2) at depth 2 still works
			// at default depth 2.
			{Code: `<button><>Save</></button>`, Tsx: true},
			// button(0) -> fragment(1) -> fragment(2) -> JsxText(3) at depth=3.
			{Code: `<button><><>Save</></></button>`, Tsx: true,
				Options: map[string]interface{}{"depth": float64(3)}},

			// ============================================================
			// React-component fallback at root (no children) — uppercase-
			// named element NOT in controlComponents, but it IS triggered
			// via controlComponents EXACT-match at top level. Pair sanity
			// case: the rule must NOT trigger for an uppercase-named tag
			// that is NOT in controlComponents at all (no DOM/role match
			// either) — no diagnostic.
			// ============================================================
			{Code: `<UnrelatedComponent />`, Tsx: true}, // no trigger anywhere
			{Code: `<UnrelatedComponent>Anything</UnrelatedComponent>`, Tsx: true},

			// ============================================================
			// Nested labelling via spread on the inner control element —
			// hasLabellingProp's spread short-circuit kicks in.
			// ============================================================
			{Code: `<button><span {...props} /></button>`, Tsx: true},

			// ============================================================
			// Deep-tree React component fallback at the leaf: the deepest
			// child is an uppercase-named component (not in controlComponents)
			// → fallback returns true → outer label requirement satisfied.
			// ============================================================
			{Code: `<button><div><Label /></div></button>`, Tsx: true},

			// ============================================================
			// controlComponents exact-match at top + minimatch in recursion
			// — `<CustomButton><CustomIcon /></CustomButton>` with
			// controlComponents=['CustomButton', 'CustomIcon']:
			//   - Top: CustomButton ∈ list → trigger.
			//   - depth 1: CustomIcon name in list (exact or glob) → fallback NOT
			//     triggered → empty recurse → false → REPORT.
			// Sanity: add aria-label to inner makes it valid.
			// ============================================================
			{Code: `<CustomButton><CustomIcon aria-label="Open" /></CustomButton>`, Tsx: true,
				Options: map[string]interface{}{
					"controlComponents": []interface{}{"CustomButton", "CustomIcon"},
				}},

			// ============================================================
			// Trigger combinations × ignoreRoles / settings interactions.
			// ============================================================
			// DOM-interactive + ignoreRoles match → ignoreRoles checked BEFORE
			// trigger → skip regardless of interactivity.
			{Code: `<button role="grid" />`, Tsx: true,
				Options: map[string]interface{}{"ignoreRoles": []interface{}{"grid"}}},
			// DOM-interactive + role outside ignoreRoles → trigger fires →
			// label required → here Save text satisfies.
			{Code: `<button role="grid">Save</button>`, Tsx: true},
			// Custom component in controlComponents + ignoreRoles match → skip.
			{Code: `<Widget role="grid" />`, Tsx: true,
				Options: map[string]interface{}{
					"controlComponents": []interface{}{"Widget"},
					"ignoreRoles":       []interface{}{"grid"},
				}},
			// Custom component + ignoreElements match → skip (ignoreElements
			// applies to the resolved tag name, even custom ones).
			{Code: `<Widget aria-label="x" />`, Tsx: true,
				Options: map[string]interface{}{
					"controlComponents": []interface{}{"Widget"},
					"ignoreElements":    []interface{}{"Widget"},
				}},
			// ignoreElements containing `link` (redundant — `link` is
			// always ignored anyway) — no-op, both `link` and `audio` skip.
			{Code: `<link />`, Tsx: true,
				Options: map[string]interface{}{"ignoreElements": []interface{}{"link", "audio"}}},
			{Code: `<audio />`, Tsx: true,
				Options: map[string]interface{}{"ignoreElements": []interface{}{"link", "audio"}}},

			// ============================================================
			// Role values: case-sensitivity in ignoreRoles AND
			// case-insensitivity in isInteractiveRole.
			// ============================================================
			// ignoreRoles is case-SENSITIVE (`includes` uses ===). role="GRID"
			// does NOT match ignoreRoles=['grid']. But isInteractiveRole
			// lowercases internally so "GRID" still resolves to interactive.
			// Combined: rule fires (not ignored, is interactive role on div) →
			// must label → "Save" satisfies → valid.
			{Code: `<div role="GRID">Save</div>`, Tsx: true,
				Options: map[string]interface{}{"ignoreRoles": []interface{}{"grid"}}},
			// Multi-role `role="button switch"` — first valid role wins.
			// "button" is interactive → trigger → label required.
			{Code: `<div role="button switch">Save</div>`, Tsx: true},
			// Multi-role with non-interactive first → no trigger via role
			// (still depends on element).
			{Code: `<div role="alert button" />`, Tsx: true},

			// ============================================================
			// `depth` extremes — 0, max=25, above max (clamped).
			// ============================================================
			// depth=0: only root inspected. Root has labelling prop → valid.
			{Code: `<button aria-label="Save" />`, Tsx: true,
				Options: map[string]interface{}{"depth": float64(0)}},
			// depth=0: root has no labelling prop, no children fall within
			// budget → invalid (tested in the invalid array below).
			// depth=25 (max): deep enough for 25-level nested text.
			{Code: `<button><span><span><span><span><span><span><span><span><span><span>Save</span></span></span></span></span></span></span></span></span></span></button>`, Tsx: true,
				Options: map[string]interface{}{"depth": float64(25)}},
			// depth=999: clamped to 25, still enough for the 11-level case
			// above (way under 25).
			{Code: `<button><span>Save</span></button>`, Tsx: true,
				Options: map[string]interface{}{"depth": float64(999)}},

			// ============================================================
			// labelAttributes edge cases.
			// ============================================================
			// labelAttributes with hyphenated prop name.
			{Code: `<button data-label="Save" />`, Tsx: true,
				Options: map[string]interface{}{"labelAttributes": []interface{}{"data-label"}}},
			// labelAttributes with duplicate entries — defensive.
			{Code: `<button title="Save" />`, Tsx: true,
				Options: map[string]interface{}{"labelAttributes": []interface{}{"title", "title"}}},
			// labelAttributes containing a builtin name (alt) — redundant
			// but harmless.
			{Code: `<button><img alt="Save" /></button>`, Tsx: true,
				Options: map[string]interface{}{"labelAttributes": []interface{}{"alt"}}},

			// ============================================================
			// Mixed children patterns (real-world JSX shapes).
			// ============================================================
			// Text + element child — text alone satisfies at depth 1.
			{Code: `<button>Save<span /></button>`, Tsx: true},
			// Element + text — same.
			{Code: `<button><span />Save</button>`, Tsx: true},
			// Expression + text — JsxExpression unconditional true.
			{Code: `<button>{prefix} Save</button>`, Tsx: true},
			// Non-ASCII text content.
			{Code: `<button>→</button>`, Tsx: true},
			{Code: `<button>保存</button>`, Tsx: true},
			// Newlines + indentation in text — TrimSpace handles.
			{
				Code: `<button>` + "\n" +
					`  Save` + "\n" +
					`</button>`,
				Tsx: true,
			},
			// Comment-only JsxExpression — tsgo emits KindJsxExpression
			// with no inner Expression. Our switch returns true
			// unconditionally for KindJsxExpression, matching upstream's
			// `case 'JSXExpressionContainer': return true`.
			{Code: `<button>{/* coming soon */}</button>`, Tsx: true},

			// ============================================================
			// Custom event handlers don't substitute for a label.
			// Already covered in invalid via `<button />` etc. but lock
			// in `<button onClick={fn} />` and friends here as INVALID.
			// (Covered below in invalid section.)
			// ============================================================

			// ============================================================
			// `controlComponents` with universal glob `["*"]` in recursion.
			// Top-level still uses exact match — `["*"]` doesn't exact-match
			// any tag → no top-level trigger from controlComponents. But
			// the rule still triggers via interactive-element / role on a
			// labelled inner: outer `<button>` triggers, inner `<Label />`
			// is uppercase + matches `*` glob → fallback OFF → recurse →
			// empty → no label → REPORT (covered in invalid below). The
			// VALID counterpart adds explicit label at the outer level.
			// ============================================================
			{Code: `<button aria-label="Save"><Anything /></button>`, Tsx: true,
				Options: map[string]interface{}{"controlComponents": []interface{}{"*"}}},

			// ============================================================
			// Settings: polymorphicAllowList narrows the polymorphic prop.
			// `<Bar as="button">` with allowList=['Foo'] → Bar NOT remapped
			// → stays custom → no trigger → valid.
			// ============================================================
			{Code: `<Bar as="button" />`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"polymorphicPropName":  "as",
						"polymorphicAllowList": []interface{}{"Foo"},
					},
				}},
			// Same settings + `<Foo as="button">` → Foo IS in allowList →
			// remapped to `button` → interactive → label required → "Save"
			// satisfies.
			{Code: `<Foo as="button">Save</Foo>`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"polymorphicPropName":  "as",
						"polymorphicAllowList": []interface{}{"Foo"},
					},
				}},

			// ============================================================
			// JSX tag name shapes — member expression, namespaced.
			// ============================================================
			// Member expression `<Foo.Bar />` — resolves to "Foo.Bar"
			// (uppercase first) → not DOM → no trigger.
			{Code: `<Foo.Bar />`, Tsx: true},
			// Member expression as controlComponent (must match the full
			// dotted form exactly at top level).
			{Code: `<Foo.Bar aria-label="x" />`, Tsx: true,
				Options: map[string]interface{}{"controlComponents": []interface{}{"Foo.Bar"}}},
			// Namespaced `<svg:path />` — resolves to "svg:path" (lowercase
			// prefix) → not in dom map → no trigger.
			{Code: `<svg:path />`, Tsx: true},

			// ============================================================
			// Real-world component patterns (extended).
			// ============================================================
			// React.lazy + Suspense (Suspense is a custom component → no
			// trigger; inner button has label).
			{Code: `<Suspense fallback={<Spinner />}><button>Save</button></Suspense>`, Tsx: true},
			// Render-prop pattern — children-as-function. The function
			// body's JSX is in a JsxExpression; the OUTER component is
			// custom (no trigger). Inner button is interactive but labelled.
			{Code: `<DataLoader>{(data) => <button>Save</button>}</DataLoader>`, Tsx: true},
			// Spread props onto a custom component — the spread is opaque
			// at the custom-component level (no trigger anyway).
			{Code: `<MyButton {...props} />`, Tsx: true},
			// Conditional rendering with both branches labelled.
			{Code: `<>{cond ? <button>A</button> : <button>B</button>}</>`, Tsx: true},
			// Logical `&&` short-circuit with labelled button.
			{Code: `<>{loading && <button aria-label="Loading" />}</>`, Tsx: true},
			// Portal-style wrapper (custom component) with labelled child.
			{Code: `<Portal><button aria-label="Close">×</button></Portal>`, Tsx: true},
			// TypeScript generic on a JSX component (not a control).
			{Code: `<Select<string> options={opts} aria-label="x" />`, Tsx: true,
				Options: map[string]interface{}{"controlComponents": []interface{}{"Select"}}},

			// ============================================================
			// Settings.components: cascading remaps + override.
			// ============================================================
			// MyButton → button (interactive, needs label) — labelled here.
			{Code: `<MyButton>Save</MyButton>`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"components": map[string]interface{}{"MyButton": "button"},
					},
				}},
			// MyDiv → div (no trigger).
			{Code: `<MyDiv />`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"components": map[string]interface{}{"MyDiv": "div"},
					},
				}},

			// ============================================================
			// Settings.components + ignoreElements interaction. Component
			// resolves to "input", which is in default-recommended
			// ignoreElements → skipped. Locks in the resolution order:
			// componentsMap FIRST, then ignore checks against the
			// resolved name.
			// ============================================================
			{Code: `<MyInput />`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"components": map[string]interface{}{"MyInput": "input"},
					},
				},
				Options: map[string]interface{}{
					"ignoreElements": []interface{}{"input"},
				}},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Listener gate: `link` is hard-coded ignored but `a` is not.
			// `<a href="#" />` triggers via interactive elementRoles.
			// ============================================================
			{Code: `<a href="#" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// `aria-hidden={true!}` (TS non-null assertion on boolean
			// literal). Upstream `TSNonNullExpression` extractor stringifies
			// the inner value and appends "!" → "true!" (a string, NOT
			// bool true). `aria-hidden === true` fails → element NOT
			// hidden → trigger check fires → no label → REPORT. rslint
			// emits the same stringified form via the dedicated
			// `case ast.KindNonNullExpression` arm in static_eval.go →
			// aligned with upstream.
			// ============================================================
			{Code: `<button aria-hidden={true!} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Listener gate: non-literal role does NOT match ignoreRoles
			// — `<div role={someRole} />` with ignoreRoles=['grid'] does
			// not skip; falls through to the trigger check. But here,
			// without a literal role, isInteractiveRole returns false →
			// trigger is false → no diagnostic. So the case is VALID, not
			// invalid. We test the invalid sibling: literal role NOT in
			// ignoreRoles → trigger fires → no label.
			// ============================================================
			{Code: `<div role="button" />`, Tsx: true,
				Options: map[string]interface{}{"ignoreRoles": []interface{}{"grid"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// hasLabellingProp trim: whitespace-only aria-label fails.
			// ============================================================
			{Code: `<button aria-label="   " />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<button aria-label="" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// JsxExpression with empty string.
			{Code: `<button aria-label={""} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// JsxExpression with whitespace-only string.
			{Code: `<button aria-label={"   "} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// aria-label={false} → falsy.
			{Code: `<button aria-label={false} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// aria-label={null} → falsy.
			{Code: `<button aria-label={null} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// aria-label={undefined} → falsy.
			{Code: `<button aria-label={undefined} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// React-component fallback: name matched in controlComponents
			// via minimatch DOES NOT save the parent. With
			// controlComponents=['MyControl*'] and `<button><MyControlA /></button>`,
			// the inner MyControlA matches the glob → fallback NOT
			// triggered → empty recurse → button reports.
			// ============================================================
			{Code: `<button><MyControlA /></button>`, Tsx: true,
				Options: map[string]interface{}{"controlComponents": []interface{}{"MyControl*"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// React-component fallback at root: `<CustomControl />` with
			// controlComponents=['CustomControl'] — the rule fires
			// because controlComponents triggers at top level, then
			// mayHaveAccessibleLabel checks the root which has empty
			// children — name='CustomControl' IS in controlComponents
			// minimatch list → fallback NOT triggered → empty recurse →
			// false → REPORT.
			// ============================================================
			{Code: `<CustomControl />`, Tsx: true,
				Options: map[string]interface{}{"controlComponents": []interface{}{"CustomControl"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// settings.components: <MyButton /> demotes to `button`. Now
			// `button` is interactive → trigger. Empty children, name is
			// now 'button' (lowercase, after componentsMap) → not React
			// component → fallback off → REPORT.
			// ============================================================
			{Code: `<MyButton />`, Tsx: true, Settings: componentsMapSettings,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos}},

			// ============================================================
			// settings.polymorphicPropName: <Foo as="button" /> becomes
			// button → interactive → no children, no label → REPORT.
			// ============================================================
			{Code: `<Foo as="button" />`, Tsx: true, Settings: polymorphicSettings,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos}},

			// ============================================================
			// Depth cap: depth=100 clamps to 25, but depth=2 (default)
			// can't reach a 4-level-nested label.
			// ============================================================
			{Code: `<button><span><span><span>Save</span></span></span></button>`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Multi-line opening element — Line/Column anchor to `<`.
			// ============================================================
			{
				Code: `<button` + "\n" +
					`/>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "controlHasAssociatedLabel",
					Message:   errorMessage,
					Line:      1,
					Column:    1,
				}},
			},

			// ============================================================
			// Nested non-interactive parent with interactive child: only
			// the inner control reports.
			// ============================================================
			{
				Code: `<div>` + "\n" +
					`  <button />` + "\n" +
					`</div>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "controlHasAssociatedLabel",
					Message:   errorMessage,
					Line:      2,
					Column:    3,
				}},
			},

			// ============================================================
			// Two same-line non-interactive controls both report.
			// ============================================================
			{
				Code: `<><button /><a href="x" /></>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "controlHasAssociatedLabel", Message: errorMessage, Line: 1, Column: 3},
					{MessageId: "controlHasAssociatedLabel", Message: errorMessage, Line: 1, Column: 13},
				},
			},

			// ============================================================
			// Real-world component patterns (reports).
			// ============================================================
			// React.forwardRef wrapping interactive button with no label.
			{Code: `const Btn = React.forwardRef((props, ref) => <button ref={ref} />);`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos}},
			// Array.map producing unlabelled interactive children.
			{Code: `const list = items.map(item => <button key={item.id} />);`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos}},
			// Generator yielding unlabelled controls — each reports.
			{Code: `function* render() { yield <button />; yield <button />; }`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos, expectedErrorAnyPos}},

			// ============================================================
			// Complex nesting: a non-control parent wrapping an unlabelled
			// interactive child. Only the inner control reports — outer
			// `div` is not interactive and doesn't trigger.
			// ============================================================
			{Code: `<div><button /></div>`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos}},

			// ============================================================
			// Nested controls — each independently classified. Outer
			// button has no label; inner has none either; BOTH report.
			// Position assertions lock the two-fire pattern.
			// ============================================================
			{
				Code: `<button><a href="#" /></button>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "controlHasAssociatedLabel", Message: errorMessage, Line: 1, Column: 1},
					{MessageId: "controlHasAssociatedLabel", Message: errorMessage, Line: 1, Column: 9},
				},
			},

			// ============================================================
			// Fragment-only label deep beyond depth budget.
			// button(0) → fragment(1) → fragment(2) → fragment(3) → JsxText(4)
			// At default depth=2, depth(3)>2 → recursion bails before
			// reaching the text. REPORT.
			// ============================================================
			{Code: `<button><><><>Save</></></></button>`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// controlComponents EXACT match at root vs glob inside:
			// `<CustomA />` with controlComponents=['CustomA*'] does NOT
			// trigger at root (top-level uses indexOf — exact only). So
			// no diagnostic fires here even though the glob would match
			// in a recursive context. Sanity: this case must be VALID
			// (asymmetry is upstream-faithful). Pair with the actual
			// invalid `<CustomA />` with exact-match controlComponents=['CustomA'].
			// ============================================================
			// Exact match at top → trigger → no label → REPORT.
			{Code: `<CustomA />`, Tsx: true,
				Options: map[string]interface{}{
					"controlComponents": []interface{}{"CustomA"},
				},
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// React-component fallback at root: `<CustomA />` is in
			// controlComponents → triggers → fallback at depth 0 checks
			// anyMinimatch and finds CustomA in the list → fallback off
			// → empty recurse → false → REPORT.
			// (Already partially covered above; this locks in glob-only
			// match for the controlComponents list — 'C*' covers 'CustomA'.)
			// ============================================================
			{Code: `<CustomA />`, Tsx: true,
				Options: map[string]interface{}{
					"controlComponents": []interface{}{"CustomA", "C*"},
				},
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Common real-world failure: icon button without label.
			// ============================================================
			// Bare button wrapping an SVG — svg lowercase, not in DomElements
			// strictly used by IsInteractiveElement, but it IS in dom map so
			// IsDOMElement(svg)=true; isReactComponent('svg')=false → fallback
			// off → no label.
			{Code: `<button><svg /></button>`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// Icon class on void child — i element, no children.
			{Code: `<button><i className="icon-save" /></button>`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// Span with className but no text.
			{Code: `<button><span className="material-icons" /></button>`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// onClick / event handlers don't substitute for a text label.
			// ============================================================
			{Code: `<button onClick={fn} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<a href="#" onClick={fn} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="button" onClick={fn} onKeyDown={fn} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// depth=0: only root inspected, descendants invisible. A
			// labelled grandchild does NOT save the button.
			// ============================================================
			{Code: `<button><span>Save</span></button>`, Tsx: true,
				Options: map[string]interface{}{"depth": float64(0)},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// depth=25 (max) — text at depth 26 not reached.
			// ============================================================
			{Code: `<button><span><span><span><span><span><span><span><span><span><span><span><span><span><span><span><span><span><span><span><span><span><span><span><span><span>Save</span></span></span></span></span></span></span></span></span></span></span></span></span></span></span></span></span></span></span></span></span></span></span></span></span></button>`, Tsx: true,
				Options: map[string]interface{}{"depth": float64(25)},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Trigger fires regardless of `role` value when element itself
			// is interactive. `<button role="img" />`, `<button role="presentation" />`,
			// `<button role={unknown} />` — all still need a label.
			// ============================================================
			{Code: `<button role="img" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<button role="presentation" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<button role={unknown} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Multi-role: first valid role wins for isInteractiveRole.
			// "presentation" is non-interactive; on a `div` with this as the
			// FIRST role, no role-trigger; div not interactive → no trigger →
			// valid actually. So this case must use BUTTON (interactive
			// element) to get a trigger. Or use role="button presentation"
			// on div: first valid = "button" (interactive) → trigger.
			// ============================================================
			{Code: `<div role="button presentation" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Cross-listener report: nested controls in a real component.
			// Each opening element classified independently.
			// ============================================================
			{
				Code: `function Toolbar() {` + "\n" +
					`  return (` + "\n" +
					`    <div>` + "\n" +
					`      <button />` + "\n" +
					`      <a href="#" />` + "\n" +
					`      <button aria-label="Save" />` + "\n" +
					`    </div>` + "\n" +
					`  );` + "\n" +
					`}`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "controlHasAssociatedLabel", Message: errorMessage, Line: 4, Column: 7},
					{MessageId: "controlHasAssociatedLabel", Message: errorMessage, Line: 5, Column: 7},
				},
			},

			// ============================================================
			// Render-prop pattern producing an unlabelled control.
			// ============================================================
			{Code: `<DataLoader>{(data) => <button />}</DataLoader>`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos}},

			// ============================================================
			// `controlComponents=["*"]` glob: inner Anything matches → fallback off.
			// Outer `<button>` still triggers via element-level interactivity;
			// recursion hits Anything at depth 1 (children empty + matches
			// `*`) → fallback off → empty recurse at deeper level → false →
			// no label found → outer button REPORTS.
			// ============================================================
			{Code: `<button><Anything /></button>`, Tsx: true,
				Options: map[string]interface{}{"controlComponents": []interface{}{"*"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// `controlComponents` glob with brace expansion (minimatch
			// supports `{a,b}` style).
			// ============================================================
			{Code: `<button><CustomA /></button>`, Tsx: true,
				Options: map[string]interface{}{"controlComponents": []interface{}{"Custom{A,B,C}"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Settings.components remaps the element to a non-ignored DOM
			// tag that IS interactive — rule fires.
			// ============================================================
			{Code: `<Submit />`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"components": map[string]interface{}{"Submit": "button"},
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos}},

			// ============================================================
			// Settings.polymorphicAllowList NOT covering the component →
			// stays custom → no trigger → valid. But for the variant where
			// allowList covers it AND `as` resolves to interactive →
			// trigger fires.
			// ============================================================
			{Code: `<Foo as="th" />`, Tsx: true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"polymorphicPropName":  "as",
						"polymorphicAllowList": []interface{}{"Foo"},
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos}},

			// ============================================================
			// Multi-line position assertions.
			// ============================================================
			// Self-closing on its own line — column on `<`.
			{
				Code: `(` + "\n" +
					`  <button` + "\n" +
					`    onClick={fn}` + "\n" +
					`  />` + "\n" +
					`)`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "controlHasAssociatedLabel",
					Message:   errorMessage,
					Line:      2,
					Column:    3,
				}},
			},
			// Opening tag with attribute spanning multiple lines, then
			// children, then closing — report column on `<` of opening.
			{
				Code: `<button` + "\n" +
					`  className="primary"` + "\n" +
					`  onClick={fn}` + "\n" +
					`>` + "\n" +
					`</button>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "controlHasAssociatedLabel",
					Message:   errorMessage,
					Line:      1,
					Column:    1,
				}},
			},
		},
	)
}
