package no_autofocus

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// polymorphicSettings mirrors upstream's `polymorphicPropName: 'as'` config.
// `<Box as="input" autoFocus />` should resolve to `input` and be reported
// even with ignoreNonDOM:true.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// polymorphicWithComponentsSettings combines polymorphicPropName + components
// — locks in the resolution order in jsxa11yutil.GetElementType (polymorphic
// runs first, then components-map remap on the unresolved rawType).
var polymorphicWithComponentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
		"components": map[string]interface{}{
			"Box": "div",
		},
	},
}

// componentsToCustomSettings maps a custom component name to ANOTHER custom
// component name (not in the dom set). Used to verify ignoreNonDOM still
// skips when components remap stays outside the dom set.
var componentsToCustomSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Foo": "Bar",
		},
	},
}

// emptyJsxA11ySettings has the `jsx-a11y` key but no inner config — exercises
// the GetElementType / IsDOMElement defensive paths when the settings tree
// exists but is empty.
var emptyJsxA11ySettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{},
}

// polymorphicAllowListSettings combines polymorphicPropName +
// polymorphicAllowList — the allow-list restricts which raw tag names the
// `as` swap applies to. Locks in that GetElementType honors the allow-list
// for ignoreNonDOM resolution.
var polymorphicAllowListSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
		"polymorphicAllowList": []interface{}{"Box"},
	},
}

// TestNoAutofocusExtras is the catch-all for everything beyond upstream's
// 18-case suite (which lives in no_autofocus_upstream_test.go). Cases here
// fall in three groups, kept in one suite so a regression bisects easily:
//
//  1. **Dimension 4 universal edge shapes** — TS wrappers, namespaced names,
//     paired vs self-closing element kinds, satisfies opaqueness, template
//     substitution placeholders, listener boundary across nested elements.
//  2. **Upstream getPropValue branches lacking dedicated upstream coverage**
//     — boolean form, numeric / null / bigint / call / member / template
//     literals, the "true"/"false" string coercion direction, exact position
//     assertions per JSX surface.
//  3. **ignoreNonDOM × polymorphicPropName × components × polymorphicAllowList
//     resolution matrix**, plus the bare/array-wrapped/malformed options
//     shapes that exercise GetOptionsMap.
//  4. **Real-world React / TS patterns** — TS generics, hyphenated tags,
//     hooks / forwardRef / memo / HOC wrappers, fragments + portals,
//     conditional rendering, multi-component files, generator/async/IIFE
//     bodies. These don't lock new semantics; they certify the listener
//     fires reliably across the AST shapes a real codebase produces.
func TestNoAutofocusExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoAutofocusRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Group 1: Case-sensitivity locks
		// ============================================================
		// upstream `propName(attribute) === 'autoFocus'` is strict; lowercase
		// HTML attribute / random casings must NOT match.
		{Code: `<div autofocus />`, Tsx: true},
		{Code: `<div AUTOFOCUS />`, Tsx: true},
		{Code: `<div AutoFocus />`, Tsx: true},
		{Code: `<div autoFocuS />`, Tsx: true},
		// Namespaced attribute name `xml:autoFocus` becomes the composite
		// "xml:autoFocus" via reactutil.GetJsxPropName, which is not equal to
		// the bare "autoFocus" — listener does NOT match.
		{Code: `<div xml:autoFocus />`, Tsx: true},

		// ============================================================
		// Group 2: "false" literal forms — all coerce to boolean false
		// ============================================================
		// Locks in jsxAstUtilsLiteralCoerce's case-insensitive coverage and
		// the TS-wrapper unwrapping inside staticEval.
		{Code: `<div autoFocus="False" />`, Tsx: true},
		{Code: `<div autoFocus="FALSE" />`, Tsx: true},
		{Code: `<div autoFocus={"false"} />`, Tsx: true},
		{Code: `<div autoFocus={"False"} />`, Tsx: true},
		{Code: `<div autoFocus={("false")} />`, Tsx: true},
		{Code: `<div autoFocus={"false" as string} />`, Tsx: true},
		{Code: `<div autoFocus={"false"!} />`, Tsx: true},
		{Code: `<div autoFocus={false as boolean} />`, Tsx: true},
		// NoSubstitutionTemplateLiteral does NOT route through
		// jsxAstUtilsLiteralCoerce, so `` `false` `` extracts to string
		// "false", which matches the upstream `=== 'false'` defensive check.
		{Code: "<div autoFocus={`false`} />", Tsx: true},

		// ============================================================
		// Group 3: Boolean / Conditional resolving to false
		// ============================================================
		{Code: `<div autoFocus={true && false} />`, Tsx: true},
		{Code: `<div autoFocus={true ? false : true} />`, Tsx: true},
		{Code: `<div autoFocus={false || false} />`, Tsx: true},

		// ============================================================
		// Group 4: ignoreNonDOM matrix — skips
		// ============================================================
		// ignoreNonDOM: true skips custom components and namespaced /
		// member-expression tag names that aren't in the dom set.
		{Code: `<Foo autoFocus />`, Tsx: true, Options: ignoreNonDOMOption},
		{Code: `<UX.Layout autoFocus />`, Tsx: true, Options: ignoreNonDOMOption},
		// `svg:circle` is a JsxNamespacedName — the composite "svg:circle" is
		// not an exact key in aria-query's dom map, so ignoreNonDOM skips.
		{Code: `<svg:circle autoFocus />`, Tsx: true, Options: ignoreNonDOMOption},
		// ignoreNonDOM:true — components map promotes `Button` → `button`
		// (in dom set), so ignoreNonDOM does not skip; explicit `false`
		// short-circuits the report.
		{Code: `<Button autoFocus={false} />`, Tsx: true, Options: ignoreNonDOMOption, Settings: componentsSettings},
		// ignoreNonDOM:true with polymorphicPropName: `<Box as="div" />`
		// resolves to `div` (in dom set), ignoreNonDOM doesn't skip, but
		// `false` short-circuits.
		{Code: `<Box as="div" autoFocus={false} />`, Tsx: true, Options: ignoreNonDOMOption, Settings: polymorphicSettings},
		// `<Box as="ComponentName" autoFocus />` resolves to `ComponentName`
		// — not in the dom set → ignoreNonDOM skips.
		{Code: `<Box as="ComponentName" autoFocus />`, Tsx: true, Options: ignoreNonDOMOption, Settings: polymorphicSettings},

		// ============================================================
		// Group 5: Default-options shapes — JSON path coverage
		// ============================================================
		// Bare `{}` (single-option CLI shape) and array-wrapped `[{}]`
		// (multi-element / rule_tester shape) — exercise both arms of
		// utils.GetOptionsMap. Default ignoreNonDOM=false → custom Foo
		// would normally fire, but explicit `false` value short-circuits.
		{Code: `<Foo autoFocus={false} />`, Tsx: true, Options: map[string]interface{}{}},
		{Code: `<Foo autoFocus={false} />`, Tsx: true, Options: []interface{}{map[string]interface{}{}}},

		// ============================================================
		// Group 6: Spread attributes are NOT JsxAttribute → listener never fires
		// ============================================================
		// upstream's listener is `JSXAttribute` only; SpreadAttribute is its
		// own AST kind. Even `{...{autoFocus: true}}` cannot trigger this
		// rule because no JsxAttribute named autoFocus exists in the tree.
		{Code: `<div {...{autoFocus: true}} />`, Tsx: true},
		{Code: `<div {...props} />`, Tsx: true},
		{Code: `<div {...{autoFocus: true}} {...props} />`, Tsx: true},

		// ============================================================
		// Group 7: Real-world patterns with explicit false (no-report)
		// ============================================================
		{
			Code: `function Modal({ open }) { return <dialog open={open}><input autoFocus={false} placeholder="search" /></dialog>; }`,
			Tsx:  true,
		},
		// useRef + manual focus pattern (recommended a11y replacement) — no
		// autoFocus prop, so nothing to report.
		{
			Code: `function Search() { const ref = useRef(null); useEffect(() => ref.current?.focus(), []); return <input ref={ref} />; }`,
			Tsx:  true,
		},
		// HOC wrapping — no autoFocus on the inner component.
		{
			Code: `const Enhanced = withTracking(({ value }) => <input value={value} />);`,
			Tsx:  true,
		},
		// React.forwardRef with autoFocus={false} — explicit false suppresses.
		{
			Code: `const FocusInput = React.forwardRef((props, ref) => <input ref={ref} autoFocus={false} {...props} />);`,
			Tsx:  true,
		},
		// React.memo + arrow component — autoFocus={false} suppresses.
		{
			Code: `const Item = React.memo(({ id }) => <li id={id} autoFocus={false}>{id}</li>);`,
			Tsx:  true,
		},
		// ignoreNonDOM × Suspense / Portal — these are framework components,
		// not in dom set; the listener must not crash on the surrounding tree.
		{
			Code:    `<Suspense fallback={<div autoFocus={false} />}><div>x</div></Suspense>`,
			Tsx:     true,
			Options: ignoreNonDOMOption,
		},

		// ============================================================
		// Group 8: TypeScript generic JSX components
		// ============================================================
		// `<List<string> autoFocus={false}>` — the type-arg is a TS-only
		// extension; tsgo parses it as a JsxOpeningElement with type args.
		// The autoFocus attribute is still reachable as a plain JsxAttribute.
		{Code: `<List<string> autoFocus={false} />`, Tsx: true},
		{
			Code:    `<List<{a: number}> autoFocus />`,
			Tsx:     true,
			Options: ignoreNonDOMOption,
		},

		// ============================================================
		// Group 9: Long-chain member-expression tags
		// ============================================================
		// `<Foo.Bar.Baz autoFocus={false} />` — element name resolves to the
		// dotted string. With ignoreNonDOM:true, the resolved string isn't in
		// the dom set → skip; without, the explicit false suppresses.
		{Code: `<Foo.Bar.Baz autoFocus={false} />`, Tsx: true},
		{
			Code:    `<Foo.Bar.Baz autoFocus />`,
			Tsx:     true,
			Options: ignoreNonDOMOption,
		},

		// ============================================================
		// Group 10: Hyphenated DOM tags (web components)
		// ============================================================
		// `<my-element>` is a custom element; jsx-ast-utils sees it as a
		// lowercase string. `dom.get('my-element')` is undefined, so with
		// ignoreNonDOM:true → not in dom set → skipped.
		{
			Code:    `<my-element autoFocus />`,
			Tsx:     true,
			Options: ignoreNonDOMOption,
		},

		// ============================================================
		// Group 11: components / settings edge shapes
		// ============================================================
		// components map promoting custom → custom (still custom): `Foo` →
		// `Bar`, neither in dom set → ignoreNonDOM still skips.
		{
			Code:     `<Foo autoFocus />`,
			Tsx:      true,
			Options:  ignoreNonDOMOption,
			Settings: componentsToCustomSettings,
		},
		// Empty `jsx-a11y` settings: settings present but inner is empty →
		// IsDOMElement falls back to raw element name; `Foo` not in dom set
		// → ignoreNonDOM skips.
		{
			Code:     `<Foo autoFocus />`,
			Tsx:      true,
			Options:  ignoreNonDOMOption,
			Settings: emptyJsxA11ySettings,
		},

		// ============================================================
		// Group 12: polymorphicAllowList restricts the `as` swap
		// ============================================================
		// `<Box as="input" />` IS in the allow-list, so rawType becomes
		// "input"; explicit false suppresses.
		{
			Code:     `<Box as="input" autoFocus={false} />`,
			Tsx:      true,
			Options:  ignoreNonDOMOption,
			Settings: polymorphicAllowListSettings,
		},
		// `<Other as="input" />` is NOT in the allow-list, so the `as` swap
		// is skipped → rawType stays "Other" → not in dom set → ignoreNonDOM
		// skips even though autoFocus is truthy.
		{
			Code:     `<Other as="input" autoFocus />`,
			Tsx:      true,
			Options:  ignoreNonDOMOption,
			Settings: polymorphicAllowListSettings,
		},

		// ============================================================
		// Group 13: Comments around the prop don't break extraction
		// ============================================================
		{Code: `<div /* before */ autoFocus={false} /* after */ />`, Tsx: true},
		{Code: `<div autoFocus={/* false */ false} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Group 1: Position assertions — JsxAttribute node range
		// ============================================================
		// `<div autoFocus />` — the JsxAttribute spans columns 6..14
		// (1-based, end-exclusive), covering exactly `autoFocus`.
		{
			Code: `<div autoFocus />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAutoFocus",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 15,
			}},
		},
		// Position with explicit value — JsxAttribute spans `autoFocus={true}`
		// (16 chars wide), columns 6..22 (1-based, end-exclusive).
		{
			Code: `<div autoFocus={true} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAutoFocus",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 22,
			}},
		},
		// Multi-line attribute — position spans the entire attribute.
		{
			Code: "<div\n  autoFocus={\n    true\n  } />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAutoFocus",
				Message:   errorMessage,
				Line:      2, Column: 3, EndLine: 4, EndColumn: 4,
			}},
		},
		// Paired (non-self-closing) element — listener still fires on the
		// JsxAttribute inside the JsxOpeningElement.
		{
			Code: `<div autoFocus>child</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAutoFocus",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 15,
			}},
		},
		// String "true" → coerce to true → reports; the JsxAttribute spans
		// `autoFocus="true"` (16 chars wide), columns 6..22.
		{
			Code: `<div autoFocus="true" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAutoFocus",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 22,
			}},
		},

		// ============================================================
		// Group 2: Listener boundary — nested elements report independently
		// ============================================================
		// Outer `<a autoFocus>` AND inner `<span autoFocus />` each emit a
		// diagnostic. Locks in that the listener doesn't dedupe and doesn't
		// bleed across the nesting boundary.
		{
			Code: `<a autoFocus><span autoFocus /></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noAutoFocus", Message: errorMessage, Line: 1, Column: 4, EndLine: 1, EndColumn: 13},
				{MessageId: "noAutoFocus", Message: errorMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 29},
			},
		},
		// Multiple autoFocus attributes on one element each report.
		// JsxAttribute listener fires once per matching attribute; tsgo
		// preserves duplicates (legal source though typically a typo).
		{
			Code: `<div autoFocus autoFocus />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noAutoFocus", Message: errorMessage, Line: 1, Column: 6, EndLine: 1, EndColumn: 15},
				{MessageId: "noAutoFocus", Message: errorMessage, Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
			},
		},

		// ============================================================
		// Group 3: Element kind survey (ignoreNonDOM unset)
		// ============================================================
		// Rule fires regardless of element type when ignoreNonDOM is unset.
		{Code: `<a autoFocus />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<input autoFocus />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Component autoFocus />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<UX.Layout autoFocus />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<svg:circle autoFocus />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 4: Truthy literal coercions
		// ============================================================
		// jsxAstUtilsLiteralCoerce handles "true" / "TRUE" / "True" → boolean
		// true → !== false → reports.
		{Code: `<div autoFocus="True" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus="TRUE" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={"true"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 5: Non-coerced falsy values still report
		// ============================================================
		// upstream's check is `!== false` (strict) — null / 0 / "" / NaN are
		// all falsy in JS but distinct from boolean false → still report.
		{Code: `<div autoFocus={null} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={""} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Empty string attribute: extracts to "" via the StringLiteral path
		// (no coerce — not "true"/"false"). "" !== false → reports.
		{Code: `<div autoFocus="" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `void 0` evaluates to undefined under upstream's UnaryExpression
		// extractor → undefined !== false → reports.
		{Code: `<div autoFocus={void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 6: Numeric / BigInt / Identifier expressions are truthy
		// ============================================================
		{Code: `<div autoFocus={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={42} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={1n} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Identifier extracts to the name string under jsx-ast-utils — non-empty
		// truthy, but more importantly !== false / 'false'.
		{Code: `<div autoFocus={someVar} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 7: Conditional / logical resolving to NOT-false
		// ============================================================
		// `true && true` → true → reports. `false || true` → true → reports.
		{Code: `<div autoFocus={true && true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={false || true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={"x" && true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 8: Call / Member / TaggedTemplate — upstream truthy synthesis
		// ============================================================
		// Upstream synthesizes a non-empty truthy string for these via the
		// call / member / tagged-template extractors, all !== false.
		{Code: `<div autoFocus={fn()} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={obj.x} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={obj?.x} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<div autoFocus={tag`x`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// TemplateExpression with substitution → upstream renders a string of
		// the form "${...}" → non-empty, !== false / 'false'.
		{Code: "<div autoFocus={`${x}`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Even `` `true` `` (NoSubstitutionTemplateLiteral) extracts to the
		// raw string "true" — does not coerce to boolean → !== false → reports.
		{Code: "<div autoFocus={`true`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 9: TS wrappers around a truthy expression
		// ============================================================
		// skipTransparent unwraps parens / TS assertion wrappers inside
		// staticEval. Locks in that wrapped truthy values still trip the rule.
		{Code: `<div autoFocus={true as boolean} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={(true)} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={(true)!} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 10: `satisfies` is OPAQUE — wrapped value falls through
		// ============================================================
		// jsx-ast-utils' TYPES table has no entry for SatisfiesExpression →
		// console.error → null. staticEval mirrors by deliberately excluding
		// OEKSatisfies from skipTransparent, so the SatisfiesExpression node
		// hits the default arm → jsNull. PropStaticBoolValue → (false, false)
		// and PropStaticStringValue → ("", false), so neither short-circuit
		// matches → reports. Even `<div autoFocus={false satisfies boolean} />`
		// reports because satisfies hides the boolean false from extraction.
		{Code: `<div autoFocus={false satisfies boolean} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div autoFocus={"false" satisfies string} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 11: TemplateExpression with `${false}` substitution
		// ============================================================
		// staticEvalTemplate renders the literal `${false}` placeholder, so the
		// full string is "${Expression}" (or similar) — non-empty truthy and
		// not equal to "false". Even an inner `false` substitution does not
		// short-circuit the report.
		{Code: "<div autoFocus={`${false}`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<div autoFocus={`prefix${false}suffix`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 12: Deep nesting × ignoreNonDOM
		// ============================================================
		// The listener visits every JsxAttribute regardless of ancestor
		// boundary; ignoreNonDOM is decided per attribute, not per subtree.
		// Outer `<Custom autoFocus />` is skipped (not in dom set), but inner
		// `<input autoFocus />` is in dom set → reports. Locks in that the
		// rule isn't accidentally inheriting an ancestor's "isCustom" verdict.
		{
			Code:    `<Outer><Mid><input autoFocus /></Mid></Outer>`,
			Tsx:     true,
			Options: ignoreNonDOMOption,
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Both outer and inner are DOM under ignoreNonDOM:true → both report.
		{
			Code:    `<div autoFocus><span autoFocus /></div>`,
			Tsx:     true,
			Options: ignoreNonDOMOption,
			Errors:  []rule_tester.InvalidTestCaseError{expectedError, expectedError},
		},
		// Mixed: DOM + custom interleaved under ignoreNonDOM:true → only the
		// DOM ones report. `<Custom autoFocus>` is skipped, `<input autoFocus />`
		// reports.
		{
			Code:    `<Custom autoFocus><input autoFocus /></Custom>`,
			Tsx:     true,
			Options: ignoreNonDOMOption,
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 13: ignoreNonDOM × DOM element → still reports
		// ============================================================
		// `input` is in the dom set → ignoreNonDOM skip does NOT apply.
		{Code: `<input autoFocus />`, Tsx: true, Options: ignoreNonDOMOption, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<button autoFocus />`, Tsx: true, Options: ignoreNonDOMOption, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Group 14: ignoreNonDOM × polymorphicPropName resolving to DOM
		// ============================================================
		// `<Box as="input" autoFocus />` resolves to `input` (in dom set)
		// → ignoreNonDOM does NOT skip → reports.
		{Code: `<Box as="input" autoFocus />`, Tsx: true, Options: ignoreNonDOMOption, Settings: polymorphicSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `<Box as="div" autoFocus />` — same shape, different tag.
		{Code: `<Box as="div" autoFocus />`, Tsx: true, Options: ignoreNonDOMOption, Settings: polymorphicSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// polymorphic + components map: `<Box as="input" autoFocus />` →
		// polymorphic runs first → rawType = "input" (in dom set) → reports.
		// The components map's `Box → div` is not consulted because the
		// polymorphic prop short-circuits the rawType.
		{Code: `<Box as="input" autoFocus />`, Tsx: true, Options: ignoreNonDOMOption, Settings: polymorphicWithComponentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// ignoreNonDOM × polymorphicAllowList: `<Box as="input" autoFocus />`
		// — `Box` IS in allow-list → swap applies → rawType = "input" → reports.
		{
			Code:     `<Box as="input" autoFocus />`,
			Tsx:      true,
			Options:  ignoreNonDOMOption,
			Settings: polymorphicAllowListSettings,
			Errors:   []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 15: Options coverage matrix — JSON path
		// ============================================================
		// Bare object (single-option CLI shape) — exercises GetOptionsMap's
		// `opts.(map[string]interface{})` arm.
		{
			Code:    `<Foo autoFocus />`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreNonDOM": false},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Array-wrapped (multi-element / rule_tester shape) — exercises the
		// `arr.([]interface{})` arm.
		{
			Code:    `<Foo autoFocus />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"ignoreNonDOM": false}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Malformed option (wrong type) is silently ignored → defaults to
		// ignoreNonDOM=false → custom component still reports.
		{
			Code:    `<Foo autoFocus />`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreNonDOM": "yes"},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 16: Real-world component patterns
		// ============================================================
		{
			Code:   `function LoginField() { return <input type="text" autoFocus />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const x = items.map(item => <li autoFocus key={item.id} />)`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Class component render with multiple offending children — each
		// reports independently.
		{
			Code: "class MyForm { render() { return <form><input autoFocus name=\"a\" /><input autoFocus name=\"b\" /></form>; } }",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError},
		},
		// Fragment + conditional rendering.
		{
			Code:   `const x = <>{cond && <input autoFocus />}</>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 17: Form patterns — search / login / signup
		// ============================================================
		{
			Code:   `function SearchBar() { return <input type="search" autoFocus placeholder="Search…" />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `function LoginForm() { return <form><input autoFocus name="user" /><input type="password" name="pass" /></form>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code: `function Signup() { return (<form><label>Email<input type="email" autoFocus /></label><label>Pass<input type="password" autoFocus /></label></form>); }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},

		// ============================================================
		// Group 18: Modal / Dialog with offending autoFocus
		// ============================================================
		{
			Code:   `function Modal() { return <dialog open><input autoFocus name="title" /><button>OK</button></dialog>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 19: HOC / forwardRef / memo wrappers carrying autoFocus
		// ============================================================
		{
			Code:   `const Enhanced = withTracking(({ value }) => <input value={value} autoFocus />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const FocusInput = React.forwardRef((props, ref) => <input ref={ref} autoFocus {...props} />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const Item = React.memo(({ id }) => <li id={id} autoFocus>{id}</li>);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 20: TypeScript generic JSX components
		// ============================================================
		// `<List<string> autoFocus />` — the listener must still see the
		// JsxAttribute despite the type args.
		{
			Code:   `<List<string> autoFocus />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<Cell<{a: number}> autoFocus={true} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 21: Long-chain member-expression tags (without ignoreNonDOM)
		// ============================================================
		{
			Code:   `<Foo.Bar.Baz autoFocus />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<this.Foo autoFocus />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 22: Hyphenated DOM tags without ignoreNonDOM
		// ============================================================
		// Listener still fires; only the truthy gate matters.
		{
			Code:   `<my-element autoFocus />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 23: Comments around / inside the prop don't suppress
		// ============================================================
		{
			Code: `<div /* a */ autoFocus /* b */ />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAutoFocus",
				Message:   errorMessage,
				Line:      1, Column: 14, EndLine: 1, EndColumn: 23,
			}},
		},
		{
			Code:   `<div autoFocus={/* truthy */ true} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 24: Complex value expressions: arrays / objects / new
		// ============================================================
		// All extract to truthy values per upstream's TYPES table; reports.
		{
			Code:   `<div autoFocus={[1, 2, 3]} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div autoFocus={{nested: true}} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div autoFocus={new Boolean(false)} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 25: Nullish coalescing / optional chain values
		// ============================================================
		{
			Code:   `<div autoFocus={cfg?.autoFocus} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div autoFocus={cfg ?? true} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 26: Conditional rendering forms
		// ============================================================
		// && short-circuit
		{
			Code:   `function Foo({cond}) { return cond && <input autoFocus />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Ternary with offending branch
		{
			Code:   `function Foo({cond}) { return cond ? <input autoFocus /> : <div />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Both branches offending
		{
			Code: `function Foo({a, b}) { return a ? <input autoFocus /> : <textarea autoFocus />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},

		// ============================================================
		// Group 27: Switch-case rendering
		// ============================================================
		{
			Code: `function Foo({type}) { switch(type) { case 'input': return <input autoFocus />; default: return null; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 28: Lists rendered via .map / .filter / .flatMap
		// ============================================================
		// Boolean attribute form inside arrow body — extracts to true → reports.
		{
			Code:   `const items = arr.map((x, i) => <li key={i} autoFocus>{x}</li>);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code: `const items = arr.filter(Boolean).flatMap(x => [<li autoFocus key={x.a}>{x.a}</li>, <li key={x.b}>{x.b}</li>]);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 29: Multiple components in one file
		// ============================================================
		{
			Code: "function A() { return <input autoFocus />; }\nfunction B() { return <input autoFocus />; }",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},

		// ============================================================
		// Group 30: Class component with state-driven JSX
		// ============================================================
		{
			Code: `class Form extends React.Component { state = {ready: true}; render() { return this.state.ready ? <input autoFocus /> : <div />; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 31: Generator / async / IIFE bodies
		// ============================================================
		{
			Code: `function* render() { yield <input autoFocus />; yield <textarea autoFocus />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError, expectedError,
			},
		},
		{
			Code:   `async function render() { return <input autoFocus />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const x = (() => <input autoFocus />)();`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Group 32: Object destructuring + JSX returning element
		// ============================================================
		{
			Code: `function Form({ initial: { autoFocus = true } }) { return <input autoFocus={autoFocus} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
	})
}
