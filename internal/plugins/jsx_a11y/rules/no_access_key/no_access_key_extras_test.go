package no_access_key

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoAccessKeyExtras covers rule behavior the upstream test suite leaves
// unaudited — Dimension 4 universal edge shapes (TS wrappers, spread
// literals, paired vs self-closing element kinds), upstream getPropValue
// branches that have no dedicated upstream test (boolean attribute form,
// numeric / null / false literals, "true"/"false" string coercion), exact
// position assertions across the JSX surface, and the listener boundary
// between nested elements.
func TestNoAccessKeyExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoAccessKeyRule, []rule_tester.ValidTestCase{
		// ---- Empty-ish / falsy literal values ----
		// Locks in upstream getPropValue → empty string is falsy and the
		// `accessKey && accessKeyValue` gate filters this out.
		{Code: `<div accessKey="" />`, Tsx: true},
		{Code: `<div accessKey={""} />`, Tsx: true},
		// Locks in upstream Literal.js' "true"/"false" string coercion: the
		// string `"false"` extracts to JS boolean `false`, hence falsy. This
		// applies ONLY to ESTree `Literal` (string) — not to TemplateLiteral.
		{Code: `<div accessKey="false" />`, Tsx: true},
		{Code: `<div accessKey={"false"} />`, Tsx: true},
		// Numeric / null / boolean falsy literals.
		{Code: `<div accessKey={0} />`, Tsx: true},
		{Code: `<div accessKey={null} />`, Tsx: true},
		{Code: `<div accessKey={false} />`, Tsx: true},
		// `void 0` evaluates to undefined under upstream's UnaryExpression
		// extractor.
		{Code: `<div accessKey={void 0} />`, Tsx: true},

		// ---- Logical / conditional short-circuits resolving to falsy ----
		// Locks in upstream's Logical / Conditional extractors with
		// statically-known operands.
		{Code: `<div accessKey={false && "h"} />`, Tsx: true},
		{Code: `<div accessKey={"" || ""} />`, Tsx: true},
		{Code: `<div accessKey={true ? "" : "h"} />`, Tsx: true},
		{Code: `<div accessKey={false ? "h" : ""} />`, Tsx: true},

		// ---- TS wrapper around a falsy literal — staticEval unwraps ----
		{Code: `<div accessKey={undefined as any} />`, Tsx: true},
		{Code: `<div accessKey={(undefined)} />`, Tsx: true},
		{Code: `<div accessKey={"" as string} />`, Tsx: true},

		// ---- Spread of an object literal whose accessKey is falsy ----
		// Locks in FindAttributeByName's literal-spread walk + PropValueIsTruthy
		// stripping of TS wrappers around the synthesized property value.
		{Code: `<div {...{accessKey: undefined}} />`, Tsx: true},
		{Code: `<div {...{accessKey: ""}} />`, Tsx: true},
		{Code: `<div {...{accesskey: false}} />`, Tsx: true},

		// ---- Spread of a non-literal — opaque to FindAttributeByName ----
		// `{...this.props}` cannot be statically known to set accessKey, so
		// the rule remains silent. Locks in the "spread non-literal is opaque"
		// behavior so a future refactor can't start treating spreads as
		// matches.
		{Code: `<div {...this.props} />`, Tsx: true},

		// ---- Element kind survey: rule is type-agnostic ----
		// Make sure the falsy gate trips for every common element shape so a
		// future "only check intrinsics" or "only check anchors" refactor
		// would have to re-pass these.
		{Code: `<a />`, Tsx: true},
		{Code: `<input />`, Tsx: true},
		{Code: `<Component />`, Tsx: true},
		{Code: `<svg:path />`, Tsx: true},
		{Code: `<UX.Layout>x</UX.Layout>`, Tsx: true},

		// ---- Paired (non-self-closing) element with no accesskey ----
		{Code: `<div>content</div>`, Tsx: true},

		// ============================================================
		// Differential-locked valid cases — every code shape below was
		// run through eslint-plugin-jsx-a11y v6.10.2 and produced the
		// same verdict (no diagnostic). Adding to lock against future
		// drift in jsxa11yutil.staticEval / FindAttributeByName.
		// ============================================================

		// ---- Numeric edge cases that resolve to a falsy number ----
		// `-0` is falsy under jsTruthy_ (Num != 0 check); upstream
		// PrefixUnary returns Number(-0).
		{Code: `<div accessKey={-0} />`, Tsx: true},
		// BigInt 0n is the only BigInt value that is falsy.
		{Code: `<div accessKey={0n} />`, Tsx: true},

		// ---- String concatenation resolving to empty ----
		{Code: `<div accessKey={"" + ""} />`, Tsx: true},

		// ---- Logical / nullish-coalescing → falsy ----
		// `null ?? ""` → "" (empty string, falsy). `undefined ?? null`
		// → null (falsy). Locks in staticEvalBinary's `??` arm.
		{Code: `<div accessKey={null ?? ""} />`, Tsx: true},
		{Code: `<div accessKey={undefined ?? null} />`, Tsx: true},

		// ---- `satisfies` is OPAQUE for jsx-ast-utils ----
		// jsx-ast-utils' TYPES table has no entry for SatisfiesExpression
		// → console.error → null (falsy). Our staticEval's
		// `skipTransparent` deliberately excludes OEKSatisfies for the
		// same reason; the SatisfiesExpression node falls to the default
		// `jsNull` arm. Locks in that intentional asymmetry between TS
		// assertion wrappers (transparent) and `satisfies` (opaque).
		{Code: `<div accessKey={"h" satisfies string} />`, Tsx: true},

		// ---- Multiple spreads with same prop — first-match wins ----
		// jsx-ast-utils' getProp returns the FIRST matching property
		// across attributes (in source order, walking literal spreads
		// as it goes). When the first hit is falsy, the rule passes,
		// regardless of later truthy declarations.
		{Code: `<div {...{accessKey: ""}} {...{accessKey: "h"}} />`, Tsx: true},
		{Code: `<div {...{accessKey: ""}} accessKey="h" />`, Tsx: true},
		{Code: `<div accessKey="" {...{accessKey: "h"}} />`, Tsx: true},

		// ---- Same key declared twice in one literal spread ----
		// Object literal duplicate-key: at runtime the LATER value wins,
		// but jsx-ast-utils uses Array.prototype.find on the parsed
		// properties, so the FIRST occurrence wins. Locks in this
		// AST-walk semantic.
		{Code: `<div {...{accessKey: "", accessKey: "h"}} />`, Tsx: true},

		// ---- TS-wrapped spread argument is opaque ----
		// upstream's `attribute.argument.type === 'ObjectExpression'`
		// check is strict; an `as any` / `!` wrapped object literal
		// fails the type check and is treated as opaque. Mirrors
		// FindAttributeByName's "strip parens only" policy.
		{Code: `<div {...({accessKey: "h"} as any)} />`, Tsx: true},
		{Code: `<div {...({accessKey: "h"})!} />`, Tsx: true},

		// ---- Computed / non-Identifier property keys are skipped ----
		// upstream getNodeName returns `key.name`; computed keys
		// (`["accesskey"]`), numeric keys (`0:`), and string keys
		// (`"accesskey":`) all have no `.name` (or it's undefined),
		// so they fall out of the case-insensitive comparison.
		{Code: `<div {...{["accessKey"]: "h"}} />`, Tsx: true},
		{Code: `<div {...{["accesskey"]: "h"}} />`, Tsx: true},
		{Code: `<div {...{0: "h"}} />`, Tsx: true},
		{Code: `<div {...{"accesskey": "h"}} />`, Tsx: true},

		// ---- Namespaced attribute name does NOT match "accesskey" ----
		// `xml:accesskey` becomes the composite name "xml:accesskey"
		// (per reactutil.GetJsxPropName), which is not equal-fold to
		// "accesskey". The rule does not match.
		{Code: `<div xml:accesskey="h" />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Exact position on a self-closing element ----
		// Two-row position assertion (Line/Column/EndLine/EndColumn) — we
		// emit on the JsxSelfClosingElement node, which spans `<div … />`.
		{
			Code: `<div accessKey="h" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAccessKey",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 1, EndColumn: 22,
			}},
		},
		// Multi-line element — position must span the entire opening
		// self-closing tag.
		{
			Code: "<div\n  accessKey=\"h\"\n  className=\"x\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAccessKey",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 4, EndColumn: 3,
			}},
		},

		// ---- Paired element: report on the OPENING element only ----
		// tsgo emits KindJsxOpeningElement for `<div ...>`; the listener
		// fires once. Also locks in that the position covers only the
		// opening tag (not the entire JsxElement). `<div accessKey="h">`
		// is 19 chars wide → EndColumn 20 (exclusive).
		{
			Code: `<div accessKey="h">child</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noAccessKey",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 1, EndColumn: 20,
			}},
		},

		// ---- Listener boundary: nested elements both report ----
		// Outer `<a accessKey="h">` AND inner `<span accessKey="i">` each
		// emit a diagnostic. Locks in that the listener doesn't dedupe and
		// doesn't bleed across the nesting boundary.
		{
			Code: `<a accessKey="h"><span accessKey="i" /></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noAccessKey", Message: errorMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				{MessageId: "noAccessKey", Message: errorMessage, Line: 1, Column: 18, EndLine: 1, EndColumn: 40},
			},
		},

		// ---- Element kind survey: rule fires regardless of element type ----
		{Code: `<a accessKey="h" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<input accessKey="h" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Component accessKey="h" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<UX.Layout accessKey="h" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Boolean attribute form ----
		// Upstream's getPropValue maps a missing attribute initializer
		// (`<div accesskey />`) to JS boolean `true` via the
		// null-attribute-value branch. The truthy gate trips. Locks in
		// AttributeIsBooleanForm + PropValueIsTruthy's boolean-form path.
		{Code: `<div accessKey />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- "true" / non-zero / non-empty literal coercions ----
		// Locks in jsxAstUtilsLiteralCoerce for the boolean-truthy direction.
		{Code: `<div accessKey="true" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={"true"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// NoSubstitutionTemplateLiteral does NOT go through the
		// `"true"`/`"false"` → boolean coercion (it routes through ESTree's
		// TemplateLiteral extractor in jsx-ast-utils, which only joins the
		// quasi text). Both `` `true` `` and `` `false` `` therefore extract
		// to non-empty strings → truthy. Locks in the staticEval branch
		// after the 2026-05 alignment fix; differential-verified against
		// eslint-plugin-jsx-a11y v6.10.2.
		{Code: "<div accessKey={`true`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<div accessKey={`false`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Numeric / non-empty-string truthy literals.
		{Code: `<div accessKey={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={"0"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- TS wrappers around a truthy expression ----
		// Locks in skipTransparent unwrapping inside staticEval.
		{Code: `<div accessKey={"h" as string} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={"h"!} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={("h")} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Conditional / logical resolving to truthy ----
		// Locks in staticEval Conditional / Logical short-circuits.
		{Code: `<div accessKey={true && "h"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={"" || "h"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={true ? "h" : ""} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- CallExpression / MemberExpression — upstream truthy synthesis ----
		// Upstream synthesizes a non-empty string for these via the call /
		// member extractors. Locks in jsTruthy via staticEval's jsTruthy
		// fallback for these kinds.
		{Code: `<div accessKey={fn()} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={obj.x} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={obj?.x} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Spread of an object literal whose accessKey is truthy ----
		// Locks in the FindAttributeByName literal-spread walk for both
		// PropertyAssignment and ShorthandPropertyAssignment forms; the
		// shorthand resolves to an Identifier with the property's name (a
		// non-empty string → truthy under PropValueIsTruthy).
		{Code: `<div {...{accessKey: "h"}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div {...{accessKey: "h"}} className="x" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Mixed-case key inside a literal spread — case-insensitive lookup.
		{Code: `<div {...{ACCESSKEY: "h"}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Shorthand-property-assignment form: `<div {...{accessKey}} />`.
		// FindAttributeByName returns the ShorthandPropertyAssignment, whose
		// initializer (per AttributeInitializer) is the bound Identifier with
		// name "accessKey" — a non-undefined identifier → name string →
		// truthy.
		{Code: `<div {...{accessKey}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Differential-locked invalid cases — every code shape below was
		// run through eslint-plugin-jsx-a11y v6.10.2 with the SAME
		// verdict.
		// ============================================================

		// ---- Numeric edge cases that resolve to truthy ----
		// `NaN` is an Identifier (not in JS_RESERVED), so jsx-ast-utils'
		// Identifier extractor returns the raw name string "NaN" →
		// truthy. Note: this is counter-intuitive (a NaN value would be
		// falsy, but jsx-ast-utils sees the *identifier name*, not the
		// runtime value). Locks in this surprising shape.
		{Code: `<div accessKey={NaN} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `Infinity` IS in JS_RESERVED → upstream returns the actual
		// +Infinity number; truthy.
		{Code: `<div accessKey={Infinity} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// PrefixUnary `-1` evaluates to -1, truthy.
		{Code: `<div accessKey={-1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// BigInt non-zero values are truthy.
		{Code: `<div accessKey={1n} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `NaN + 0` — Identifier "NaN" treated as the string "NaN", and
		// `+` with one string operand stringifies the other → "NaN0"
		// (truthy). Mirrors staticEvalBinary's `+` string-concat path.
		{Code: `<div accessKey={NaN + 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- String concatenation resolving to truthy ----
		{Code: `<div accessKey={"h" + "i"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Identifier on the right — staticEval treats it as the name
		// string "x" (matching upstream's Identifier extractor for the
		// non-reserved name path).
		{Code: `<div accessKey={"" + x} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Composite container literals — always truthy ----
		// Array / Object / RegExp / JSX-element / arrow / function-expression
		// all extract to truthy values per jsx-ast-utils' TYPES table.
		{Code: `<div accessKey={[]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={[1, 2]} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={{}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={{x: 1}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={/foo/} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={<span />} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={() => {}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div accessKey={function() {}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Logical / nullish-coalescing → truthy ----
		// `??` arm coverage: left non-null/undef → returns left.
		{Code: `<div accessKey={x ?? "h"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Left null → returns right.
		{Code: `<div accessKey={null ?? "h"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Composite `||` chain.
		{Code: `<div accessKey={x || y || ""} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Tagged template literal — upstream truthy synthesis ----
		// staticEval's KindTaggedTemplateExpression returns jsTruthy.
		{Code: "<div accessKey={tag`x`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- Real-world component patterns ----
		// Locks in that the listener fires reliably inside common React
		// shapes: function components, class render methods, JSX inside
		// .map() callbacks, fragment + conditional, ternary-rendered
		// element, hooks consumers.
		{
			Code: `function SubmitButton() { return <button accessKey="s" type="submit">Submit</button>; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code: `function HomeLink() { return <a href="/home" accessKey="h" target="_blank">Home</a>; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code: `function NameField() { return <input type="text" accessKey="n" placeholder="name" />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code: `function ToolbarBtn() { return <div role="button" tabIndex={0} accessKey="t" onClick={fn} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Fragment + conditional rendering.
		{
			Code:   `const x = <>{cond && <div accessKey="h" />}</>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// .map() callback — JSX inside arrow body.
		{
			Code:   `const x = items.map(item => <li accessKey={item.key} key={item.id} />)`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Ternary-rendered element.
		{
			Code:   `const x = cond ? <div accessKey="h" /> : <div />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Class component render with multiple offending children — each
		// reports independently.
		{
			Code: "class MyForm { render() { return <form><input accessKey=\"u\" name=\"username\" /><input accessKey=\"p\" name=\"password\" type=\"password\" /></form>; } }",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError},
		},
		// Hooks pattern.
		{
			Code:   `function Stateful() { return <input value={v} onChange={onChange} accessKey="i" />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- Spread + direct attribute mix — multiple orderings ----
		// In all these orderings, `accessKey="h"` is reachable as the
		// first matching property, so the rule reports.
		{
			Code:   `<div {...spread1} {...spread2} accessKey="h" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div accessKey="h" {...spread1} {...spread2} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div {...spread1} accessKey="h" {...spread2} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- Multiple spread literals — first-match wins (truthy first) ----
		// First match `{accessKey: "h"}` is truthy, second `""` is
		// ignored. Mirror image of the corresponding valid case.
		{
			Code:   `<div {...{accessKey: "h"}} {...{accessKey: ""}} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Same-spread duplicate keys, first found is truthy.
		{
			Code:   `<div {...{accessKey: "h", accessKey: ""}} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- Multi-property literal spread ----
		// Sibling property in the same literal does not interfere with
		// the accesskey match.
		{
			Code:   `<div {...{className: "x", accessKey: "h"}} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- Nested SpreadAssignment inside literal spread ----
		// upstream's getProp filters `type === 'Property'`, so the
		// inner SpreadAssignment is opaque, but the sibling
		// `accessKey: "h"` property is still found.
		{
			Code:   `<div {...{...other, accessKey: "h"}} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div {...{accessKey: "h", ...other}} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
	})
}
