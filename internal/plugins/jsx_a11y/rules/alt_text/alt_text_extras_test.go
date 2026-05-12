package alt_text

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestAltTextExtras locks in rslint-specific behaviors beyond what the
// upstream eslint-plugin-jsx-a11y test suite exercises:
//
//   - jsx-ast-utils getPropValue / getLiteralPropValue static-eval edges
//     (logical short-circuit, ternary branch selection, JS_RESERVED globals,
//     numeric / BigInt / template / unary / sequence / assignment, etc.)
//   - TS-only expression wrappers (`as`, `!`, `<T>`, `satisfies`)
//   - Polymorphic-prop edge cases (non-string truthy literal, allow-list)
//   - Spread / shorthand-spread first-match semantics
//   - aria-hidden full static-eval coverage in nested children
//   - Real-world i18n / optional-chain / fallback patterns
//
// Each case here was verified empirically against eslint-plugin-jsx-a11y@latest
// to ensure the rslint port produces byte-identical diagnostics.
func TestAltTextExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AltTextRule, []rule_tester.ValidTestCase{
		// ---- Lock-in: extra edge cases beyond upstream ----
		// Locks in upstream `ariaLabelHasValue` arm: aria-label boolean form
		// (no initializer) is treated as having value `true`.
		{Code: `<img aria-label />`, Tsx: true},
		// Locks in upstream `ariaLabelHasValue` arm: aria-labelledby boolean
		// form is also treated as having value.
		{Code: `<img aria-labelledby />`, Tsx: true},
		// Locks in upstream `getLiteralPropValue` for templates with no
		// substitution: `{`presentation`}` is a literal "presentation" value
		// for the role attr.
		{Code: "<img alt=\"\" role={`none`} />", Tsx: true},
		// Locks in `<object>` accessible-child arm: an inner JSXExpression
		// other than `{undefined}` counts as accessible.
		{Code: `<object>{value}</object>`, Tsx: true},
		// Locks in `<object>` arm: dangerouslySetInnerHTML on the opening
		// element makes the element accessible (no inner-text needed).
		{Code: `<object dangerouslySetInnerHTML={{__html:""}} />`, Tsx: true},
		// Locks in `<object>` arm: explicit `children` attribute counts as
		// accessible (matches upstream's hasAnyProp fallback).
		{Code: `<object children={"x"} />`, Tsx: true},
		// Locks in `<object>` arm: a string-literal child node counts as
		// accessible (matches upstream's `case 'Literal'`).
		{Code: `<object>{"Foo"}</object>`, Tsx: true},
		// Locks in `<object>` arm: a JSX child whose tag is not screen-reader
		// hidden counts as accessible.
		{Code: `<object><p>Body</p></object>`, Tsx: true},
		// Locks in upstream input branch: a non-`image` typed input is
		// silently skipped (the rule never fires on `<input type="text" />`).
		{Code: `<input type="text" />`, Tsx: true},
		// Locks in `tagName` translation: uppercase `INPUT` doesn't match
		// the `input` lowercase listener (upstream skips `<INPUT>`).
		{Code: `<INPUT />`, Tsx: true},
		// Locks in TS-wrapper unwrapping in LiteralStringValue: alt extracts
		// "foo" through `as`, matching jsx-ast-utils' TSAsExpression unwrap.
		{Code: `<img alt={"foo" as string} />`, Tsx: true},
		// Locks in TS-wrapper unwrapping for role: the value still resolves
		// to the literal "presentation" through the assertion.
		{Code: `<img alt="" role={"presentation" as const} />`, Tsx: true},
		// Locks in TS-wrapper unwrapping in expressionIsLikelyTruthy: a
		// non-null asserted variable is still treated as a value reference,
		// keeping the alt valid.
		{Code: `<img alt={alt!} />`, Tsx: true},
		// Locks in PrefixUnary `!` truthiness in expressionIsLikelyTruthy:
		// `!""` is truthy → valid alt.
		{Code: `<img alt={!""} />`, Tsx: true},
		// Locks in BigInt literal handling for non-zero values: `1n` is
		// truthy → valid (developer wrote SOMETHING).
		{Code: `<img alt={1n} />`, Tsx: true},
		// Locks in ConditionalExpression: at least one branch is truthy →
		// valid.
		{Code: `<img alt={cond ? "x" : "y"} />`, Tsx: true},
		// Locks in shorthand-spread match in FindAttributeByName: the
		// shorthand `alt` resolves to the bound identifier (truthy).
		{Code: `<img {...{alt}} />`, Tsx: true},
		// Locks in spread-object literal-property match: `alt: "x"` makes
		// the alt extractable as the literal string.
		{Code: `<img {...{alt: "x"}} />`, Tsx: true},
		// Locks in JsxElement-with-closing-tag's opening attributes being
		// inspected for aria-hidden: a paired `<div></div>` (no aria-hidden)
		// is NOT hidden → counts as accessible.
		{Code: `<object><div></div></object>`, Tsx: true},

		// ---- staticEval lock-ins (verified empirically against upstream ESLint) ----
		// Each case below was run through real eslint-plugin-jsx-a11y to
		// confirm the expected diagnostic. See /tmp/jsx-a11y-verify probes.
		//
		// LogicalExpression `&&`: short-circuits to left when left falsy.
		// `"" && x` → "" → valid via the `=== ''` branch.
		{Code: `<img alt={"" && x} />`, Tsx: true},
		// `x && ""` → "" (x is truthy via Identifier-name → take right) → valid.
		{Code: `<img alt={x && ""} />`, Tsx: true},
		// `"" && ""` → "" → valid.
		{Code: `<img alt={"" && ""} />`, Tsx: true},
		// LogicalExpression `||`: short-circuits to left when left truthy.
		// `"" || x` → "x" → valid.
		{Code: `<img alt={"" || x} />`, Tsx: true},
		// NullishCoalescing `??`: returns right only when left is null/undefined.
		// `null ?? "x"` → "x" → valid.
		{Code: `<img alt={null ?? "x"} />`, Tsx: true},
		// `"" ?? x` → "" (empty string is not null/undefined, so left wins) → valid.
		{Code: `<img alt={"" ?? x} />`, Tsx: true},
		// ConditionalExpression: real branch selection by test's truthiness.
		// `cond ? "x" : "y"` — cond's Identifier-name "cond" is truthy → "x" → valid.
		{Code: `<img alt={cond ? "x" : "y"} />`, Tsx: true},
		// `cond ? "" : "y"` — cond truthy → "" → valid via `=== ''` branch.
		{Code: `<img alt={cond ? "" : "y"} />`, Tsx: true},
		// `undefined ? "x" : "y"` — undefined identifier is falsy → "y" → valid.
		{Code: `<img alt={undefined ? "x" : "y"} />`, Tsx: true},
		// `false ? "x" : ""` — false → "" → valid via `=== ''` branch.
		{Code: `<img alt={false ? "x" : ""} />`, Tsx: true},
		// Reserved-word identifiers: Infinity → +Infinity (truthy) → valid.
		{Code: `<img alt={Infinity} />`, Tsx: true},
		// `NaN` is NOT in jsx-ast-utils' JS_RESERVED set, so it's treated as a
		// regular identifier and returns the bare name string "NaN" (truthy).
		{Code: `<img alt={NaN} />`, Tsx: true},
		// `Number` IS in JS_RESERVED → returns the constructor function (truthy).
		{Code: `<img alt={Number} />`, Tsx: true},
		// String concatenation via `+`.
		{Code: `<img alt={"a" + "b"} />`, Tsx: true},
		// `x + ""` → "x" (Identifier name + empty string) → valid.
		{Code: `<img alt={x + ""} />`, Tsx: true},
		// Template-with-substitutions: upstream extracts a synthesized string;
		// we approximate with placeholder-rendering — always non-empty truthy.
		{Code: "<img alt={`x${y}z`} />", Tsx: true},
		// `Math` / `Date` are JS_RESERVED — upstream returns the global value
		// (object / function), both truthy → valid.
		{Code: `<img alt={Math} />`, Tsx: true},
		{Code: `<img alt={Date} />`, Tsx: true},
		// Empty array / empty object literals are truthy in JS → valid.
		{Code: `<img alt={[]} />`, Tsx: true},
		{Code: `<img alt={[1,2]} />`, Tsx: true},
		{Code: `<img alt={({})} />`, Tsx: true},
		// `+` string concat with null / undefined coerces to "null" / "undefined"
		// (truthy), matching JS's String() semantics in jsx-ast-utils.
		{Code: `<img alt={"" + null} />`, Tsx: true},
		{Code: `<img alt={null + ""} />`, Tsx: true},
		{Code: `<img alt={undefined + ""} />`, Tsx: true},
		{Code: `<img alt={"a" + null} />`, Tsx: true},
		// Numeric `+`: 1 + 2 → 3 (truthy) → valid.
		{Code: `<img alt={1 + 2} />`, Tsx: true},
		// PrefixUnary `!`: `!0` → true → valid.
		{Code: `<img alt={!0} />`, Tsx: true},
		// 1.0 normalizes to 1 → truthy → valid.
		{Code: `<img alt={1.0} />`, Tsx: true},
		// `"" + ""` → "" → valid via the `=== ''` branch.
		{Code: `<img alt={"" + ""} />`, Tsx: true},
		// Nested chain that resolves to a truthy identifier name.
		{Code: `<img alt={a && (b || (c && ""))} />`, Tsx: true},

		// ---- Real-world patterns (verified against upstream ESLint) ----
		// i18n function calls — translation result is a CallExpression
		// (jsTruthy) → valid.
		{Code: `<img alt={t("img.alt")} />`, Tsx: true},
		{Code: `<img alt={i18n.t("key")} />`, Tsx: true},
		{Code: `<img alt={i18n.t("key", {defaultValue: "x"})} />`, Tsx: true},
		// Optional-call (`t?.("key")`) — same kind in tsgo (CallExpression
		// with the optional flag), still jsTruthy → valid.
		{Code: `<img alt={t?.("key")} />`, Tsx: true},
		{Code: `<img alt={t?.("key") ?? "fallback"} />`, Tsx: true},
		// Optional-chain property access — same kind in tsgo
		// (PropertyAccessExpression with the optional flag) → valid.
		{Code: `<img alt={data?.alt} />`, Tsx: true},
		{Code: `<img alt={data?.alt ?? ""} />`, Tsx: true},
		{Code: `<img alt={a?.b?.c} />`, Tsx: true},
		{Code: `<img alt={a?.b?.()} />`, Tsx: true},
		// Default-fallback patterns common in real code.
		{Code: `<img alt={altText || "Image"} />`, Tsx: true},
		{Code: `<img alt={altText ?? "Image"} />`, Tsx: true},
		{Code: `<img alt={t && t("key")} />`, Tsx: true},
		{Code: `<img alt={t ? t("key") : ""} />`, Tsx: true},

		// ---- Spread ordering (matches upstream getProp's "first match wins") ----
		// alt before spread → matched first.
		{Code: `<img alt="x" {...rest} />`, Tsx: true},
		// alt after spread (non-literal spread is opaque, getProp falls
		// through and finds alt).
		{Code: `<img {...rest} alt="x" />`, Tsx: true},
		// Two non-literal spreads, both opaque, alt missing — covered in
		// invalid section.

		// ---- JSX namespaced attribute ----
		// `alt:foo` is not the `alt` attribute (different name); the rule
		// must not match it. Matches upstream's `propName` returning
		// "alt:foo" for JsxNamespacedName.
		{Code: `<img alt:foo="bar" alt="x" />`, Tsx: true},

		// ---- Empty / whitespace-only template literal ----
		// `\`\`` (empty no-sub template) → "" → valid via `=== ''`.
		{Code: "<img alt={``} />", Tsx: true},
		// `\`  \`` (whitespace-only) → "  " (truthy) → valid.
		{Code: "<img alt={`  `} />", Tsx: true},

		// ---- Object / array literal alt (truthy in JS) ----
		{Code: `<img alt={{x: 1}} />`, Tsx: true},
		{Code: `<img alt={[1, 2, 3]} />`, Tsx: true},

		// ---- Identifier reserved-word edge cases ----
		// `Boolean` / `Symbol` / `JSON` are NOT in jsx-ast-utils JS_RESERVED
		// → returned as their identifier name (truthy non-empty string).
		{Code: `<img alt={Boolean} />`, Tsx: true},
		{Code: `<img alt={Symbol} />`, Tsx: true},
		{Code: `<img alt={JSON} />`, Tsx: true},

		// ---- Numeric edge cases ----
		// -0.5 → -0.5 (truthy) → valid.
		{Code: `<img alt={-0.5} />`, Tsx: true},
		// `1 > 0` → true (boolean) → valid.
		{Code: `<img alt={1 > 0} />`, Tsx: true},
		// `1 === 1` → true → valid.
		{Code: `<img alt={1 === 1} />`, Tsx: true},
		// Bitwise: `1 | 0` → 1 → truthy → valid.
		{Code: `<img alt={1 | 0} />`, Tsx: true},
		{Code: `<img alt={2 & 3} />`, Tsx: true},
		// Modulo: `5 % 2` → 1 → truthy → valid.
		{Code: `<img alt={5 % 2} />`, Tsx: true},
		// Power: `2 ** 3` → 8 → truthy → valid.
		{Code: `<img alt={2 ** 3} />`, Tsx: true},

		// ---- Multi-line / formatting ----
		// Multi-line JSX with valid alt — rule must accept across newlines.
		{
			Code: `<img
				alt="multi-line"
				src="x"
			/>`,
			Tsx: true,
		},
		// Spread + alt on different lines, alt visible → valid.
		{
			Code: `<img
				{...this.props}
				alt="x"
			/>`,
			Tsx: true,
		},

		// ---- Paired tag form (legal but uncommon) ----
		{Code: `<img alt="x"></img>`, Tsx: true},
		{Code: `<input type="image" alt="x"></input>`, Tsx: true},

		// ---- Settings: polymorphicAllowList restricts which raw types may
		//     be remapped via the polymorphic prop. Component NOT in the
		//     allow-list keeps its original tag — rule sees it as unknown
		//     custom component, not validated by default → valid.
		{
			Code: `<NotAllowed as="img" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"OtherTag"},
				},
			},
		},

		// ---- Settings combination: polymorphic + componentMap ----
		// `<X as="input" type="image" alt="x" />` — `as` rewrites to "input",
		// then componentMap doesn't have "input" key, so it stays "input",
		// then alt validation runs on input[type="image"].
		{
			Code: `<X as="input" type="image" alt="x" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components":          map[string]interface{}{"X": "div"},
				},
			},
		},

		// ---- Nested object accessible child (paired tags) ----
		{Code: `<object><div><span>Description</span></div></object>`, Tsx: true},
		// Multiple expression children — at least one non-undefined → valid.
		{Code: `<object>{"prefix"}{value}{"suffix"}</object>`, Tsx: true},
		// JsxExpression with nested JsxElement → counts as accessible.
		{Code: `<object>{<span>x</span>}</object>`, Tsx: true},

		// ---- Empty `elements` array → rule disabled (typesToValidate empty) ----
		// Even `<img />` (which would normally be invalid) doesn't fire
		// because the listener finds no matching tag.
		{
			Code:    `<img />`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{}},
		},

		// ---- Multi-spread + ObjectLiteral ordering (verified via ESLint) ----
		// First-match wins. `<img alt="y" {...{alt: "x"}} />` — direct alt
		// found before the spread → "y" → valid.
		{Code: `<img alt="y" {...{alt: "x"}} />`, Tsx: true},
		// First match is the spread's inner property → "x" → valid.
		{Code: `<img {...{alt: "x"}} alt="y" />`, Tsx: true},
		// Two literal-spreads — first wins.
		{Code: `<img {...{alt: "first"}} {...{alt: "second"}} />`, Tsx: true},
		// Empty-string second spread doesn't matter; first wins.
		{Code: `<img {...{alt: "first"}} {...{alt: ""}} />`, Tsx: true},
		// First spread is empty string → valid via `=== ''`.
		{Code: `<img {...{alt: ""}} {...{alt: "second"}} />`, Tsx: true},
		// Spread alt with conditional value — staticEval picks the
		// truthy-test consequent.
		{Code: `<img {...{alt: x ? "" : "y"}} />`, Tsx: true},
		// Spread alt with empty-string LITERAL → valid via `=== ''`.
		{Code: `<img {...{alt: ""}} />`, Tsx: true},

		// ---- Aria-hidden value variants in nested children ----
		// `aria-hidden={false}` → NOT hidden → div counts as accessible →
		// object valid.
		{Code: `<object><div aria-hidden={false}>x</div></object>`, Tsx: true},
		// `aria-hidden={"false"}` → string literal "false" → not "true",
		// NOT hidden → valid.
		{Code: `<object><div aria-hidden={"false"}>x</div></object>`, Tsx: true},
		// `aria-hidden={someVar}` → unknown → NOT hidden (we only flag
		// statically true) → valid. Matches upstream's getPropValue
		// returning the identifier name (truthy non-bool) which then fails
		// the `=== true` check in upstream's isHiddenFromScreenReader.
		{Code: `<object><div aria-hidden={someVar}>x</div></object>`, Tsx: true},

		// ---- TS-only wrappers (`as` / `!`) — Go-impl specific ----
		// Verified via real rslint binary. Each shape unwraps to its inner
		// expression via skipTransparent (parens + type assertions +
		// non-null assertions). `satisfies` is intentionally EXCLUDED — see
		// the invalid section below for the lock-in.
		// `altText!` — non-null assertion on a value-bearing identifier.
		{Code: `<img alt={altText!} />`, Tsx: true},
		// `altText as string` — type assertion.
		{Code: `<img alt={altText as string} />`, Tsx: true},
		// Double TS wrapping `(altText as any)!`.
		{Code: `<img alt={(altText as any)!} />`, Tsx: true},
		// `(undefined as any)!` — upstream `TSNonNullExpression` extractor
		// stringifies inner + appends "!" → "undefined!" (non-empty truthy
		// string, ≠ ''). AltAttributeIsValid → valid alt → no report.
		// Locks in alignment after the OEKNonNullAssertions strip was
		// removed from `skipTransparent` (see jsxa11yutil.go).
		{Code: `<img alt={(undefined as any)!} />`, Tsx: true},
		// String literal under `as` — extracts to "foo" → truthy → valid.
		{Code: `<img alt={"foo" as string} />`, Tsx: true},
		// Empty string under `as` — extracts to "" → valid via `=== ''`.
		{Code: `<img alt={"" as string} />`, Tsx: true},
		// `as const` on a literal — still a literal "x" → valid.
		{Code: `<img alt="" role={"presentation" as const} />`, Tsx: true},

		// ---- input[type] case sensitivity (verified against ESLint) ----
		// `type="IMAGE"` — upstream does case-sensitive `=== "image"`, so
		// non-image inputs are correctly skipped (no alt error).
		{Code: `<input type="IMAGE" />`, Tsx: true},
		{Code: `<input type="Image" />`, Tsx: true},
		// Non-string-literal type that doesn't statically resolve to "image" → skip.
		{Code: `<input type={typeVar} />`, Tsx: true},
		// Boolean form `<input type />` — upstream getPropValue returns
		// boolean true, true !== "image" → skip.
		{Code: `<input type />`, Tsx: true},
		// Logical short-circuit: `"image" && ""` → "" → !== "image" → skip.
		{Code: `<input type={"image" && ""} />`, Tsx: true},
		// `?? "image"` only fires when left is null/undefined; `null ?? "x"` → "x" → != "image" → skip.
		{Code: `<input type={null ?? "x"} />`, Tsx: true},
		// type with TS wrapper around a non-image literal.
		{Code: `<input type={"text" as string} />`, Tsx: true},

		// ---- object title with literal non-string values (verified against ESLint) ----
		// Upstream `getLiteralPropValue(null literal)` → "null" (string!) →
		// truthy → title is sufficient, no error.
		{Code: `<object title={null} />`, Tsx: true},
		// Number literal → truthy if non-zero → no error.
		{Code: `<object title={123} />`, Tsx: true},
		// Boolean true literal → truthy → no error.
		{Code: `<object title={true} />`, Tsx: true},
		// Template literal with subs → upstream renders to a non-empty
		// string → truthy → no error.
		{Code: "<object title={`x${y}`} />", Tsx: true},

		// ---- alt with `"true"` (string normalized to boolean true) ----
		// Upstream Literal extractor: `"true".toLowerCase() === "true"` → bool true
		// → truthy → valid.
		{Code: `<img alt="true" />`, Tsx: true},
		{Code: `<img alt="True" />`, Tsx: true},
		{Code: `<img alt="TRUE" />`, Tsx: true},
		{Code: `<img alt={"true"} />`, Tsx: true},
		{Code: `<img alt={"True"} />`, Tsx: true},
		// Non-bool string with content → string → truthy → valid.
		{Code: `<img alt="null" />`, Tsx: true},
		{Code: `<img alt="undefined" />`, Tsx: true},
		{Code: `<img alt="0" />`, Tsx: true},

		// ---- Pre-existing non-coverage corner-cases (verified against ESLint) ----
		// AssignmentExpression — upstream returns "left op right" string
		// (always non-empty truthy regardless of right's value).
		{Code: `<img alt={(x = "foo")} />`, Tsx: true},
		{Code: `<img alt={(x = "")} />`, Tsx: true},
		{Code: `<img alt={(x = false)} />`, Tsx: true},
		{Code: `<img alt={(x += 1)} />`, Tsx: true},
		// SequenceExpression / comma — upstream returns array of all values
		// (truthy regardless). My rsImpl handles `bin.OperatorToken.Kind ==
		// CommaToken` as truthy.
		{Code: `<img alt={(a, b, "foo")} />`, Tsx: true},
		{Code: `<img alt={(a, b, false)} />`, Tsx: true},
		{Code: `<img alt={(a, b, "")} />`, Tsx: true},
		// NewExpression — upstream returns empty object → truthy.
		{Code: `<img alt={new Date()} />`, Tsx: true},
		{Code: `<img alt={new Image(800)} />`, Tsx: true},
		// TaggedTemplateExpression — upstream redirects to TemplateLiteral
		// extractor, returns the (always non-empty) template text.
		{Code: "<img alt={tag`x`} />", Tsx: true},
		{Code: "<img alt={tag`hello ${world}`} />", Tsx: true},
		// Polymorphic with non-string truthy literal — upstream replaces
		// rawType with the non-string value, which doesn't match
		// typesToValidate. Mirror by coercing to a string form so the Set
		// lookup misses.
		{
			Code: `<Img as={null} />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components":          map[string]interface{}{"Img": "img"},
				},
			},
		},
		{
			Code: `<Img as={123} />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components":          map[string]interface{}{"Img": "img"},
				},
			},
		},
		{
			Code: `<Img as={true} />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components":          map[string]interface{}{"Img": "img"},
				},
			},
		},

		// ---- aria-hidden in nested children — full static eval coverage ----
		// `aria-hidden={cond ? false : false}` — both branches false →
		// resolves to false → not hidden → div is accessible → object valid.
		{Code: `<object><div aria-hidden={cond ? false : false}>x</div></object>`, Tsx: true},
		// `aria-hidden={"true" && false}` — short-circuits to false → not hidden.
		{Code: `<object><div aria-hidden={"true" && false}>x</div></object>`, Tsx: true},
		// `aria-hidden={false || "x"}` — string "x" → not boolean true → not hidden.
		{Code: `<object><div aria-hidden={false || "x"}>x</div></object>`, Tsx: true},
		// `aria-hidden` inside TS wrapper — same.
		{Code: `<object><div aria-hidden={false as any}>x</div></object>`, Tsx: true},

		// ---- isHiddenFromScreenReader on `<input type="hidden">` (literal-only) ----
		// `type="HIDDEN"` (case-insensitive comparison upstream) → hidden.
		// We're matching that for the inner check, so this object reports
		// invalid (no other accessible child).

		// ---- jsx-ast-utils Literal extractor on direct attr value (not in expr) ----
		// `aria-hidden=true` — JSX boolean form is parsed as Identifier, but
		// `aria-hidden="true"` is parsed as StringLiteral "true".
		// Both should classify the div as hidden, hence the object is invalid.
		// Already covered with explicit invalid cases below.
	}, []rule_tester.InvalidTestCase{
		// ---- Lock-in: extra invalid edge cases beyond upstream ----
		// Locks in upstream branch where alt is `<img alt={null} />` —
		// LiteralValue null is not extracted as truthy → invalid.
		{
			Code: `<img alt={null} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Locks in upstream branch: `<img alt={false} />` is invalid (literal
		// false is falsy and not the empty string).
		{
			Code: `<img alt={false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Locks in img branch where presentation role short-circuits the
		// aria-label / aria-labelledby checks.
		{
			Code: `<img role="presentation" aria-label="something" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferAlt", Message: msgPreferAlt, Line: 1, Column: 1},
			},
		},
		// Locks in upstream area branch: `area` with `<area alt="" />` is
		// valid; `<area alt={null} />` is invalid (alt validity logic).
		{
			Code: `<area alt={null} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		// Locks in input[type="image"] alt validity branch.
		{
			Code: `<input type="image" alt={null} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		// Locks in TS-wrapper unwrap in AttributeIsExplicitUndefined:
		// `<img alt={undefined as any} />` still resolves to undefined →
		// invalid (matches jsx-ast-utils TSAsExpression unwrap).
		{
			Code: `<img alt={undefined as any} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Locks in numeric falsy normalization: `<img alt={0.0} />` evaluates
		// to 0 (falsy) → invalid. Without normalization, the raw token
		// "0.0" wouldn't match "0".
		{
			Code: `<img alt={0.0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Locks in numeric falsy normalization across hex: `<img alt={0x0} />`
		// is 0 → falsy → invalid.
		{
			Code: `<img alt={0x0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Locks in BigInt falsy: `<img alt={0n} />` is falsy → invalid.
		{
			Code: `<img alt={0n} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Locks in PrefixUnary `!` truthiness: `!"x"` is `false` → falsy →
		// invalid.
		{
			Code: `<img alt={!"x"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Locks in ConditionalExpression with both falsy branches: invalid.
		{
			Code: `<img alt={cond ? false : null} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Locks in JsxFragment-as-child NOT counting as accessible (matches
		// upstream's missing `case 'JSXFragment'` falling to default false).
		{
			Code: `<object><>x</></object>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		// Locks in paired-tag aria-hidden inspection: `<div aria-hidden></div>`
		// (paired, not self-closing) IS hidden → object reports invalid.
		{
			Code: `<object><div aria-hidden></div></object>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		// Locks in paired-tag `<input type="hidden">` short-circuit on
		// isHiddenFromScreenReader.
		{
			Code: `<object><input type="hidden" /></object>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		// Locks in spread-object literal-property match for the negative
		// case: `{...{alt: ""}}` → alt is the empty string → object error
		// because alt extraction succeeds but the literal value, by upstream
		// semantics, is "" — `(altValue && !isNullValued) || altValue === ''`
		// makes empty string VALID. So this case is actually valid; we add
		// the negative case via `aria-label: ""`.
		{
			Code: `<img {...{"aria-label": ""}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 1},
			},
		},

		// ---- staticEval lock-ins for INVALID cases ----
		// `false && "x"` → false (left short-circuits) → falsy → invalid.
		{
			Code: `<img alt={false && "x"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// `"x" && false` → false (left truthy → take right) → falsy → invalid.
		{
			Code: `<img alt={"x" && false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// `null ?? false` → false (null is nullish → take right) → falsy → invalid.
		{
			Code: `<img alt={null ?? false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Nested chain: `a && (b || (c && ""))` — c truthy → take "", then
		// b || "" → b ("b" identifier name truthy → take "b"), then a && "b"
		// → "b" (truthy) — actually VALID. Upstream confirmed: altValue="b".
		// We'd need a chain that resolves to falsy/empty for an invalid case.
		// Use `(a && false)` instead.
		{
			Code: `<img alt={(a && false) || false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// PrefixUnary `!1` → false → invalid.
		{
			Code: `<img alt={!1} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Numeric arithmetic: `0 + 0` → 0 → falsy → invalid.
		{
			Code: `<img alt={0 + 0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},

		// ---- Comparison / bitwise (verified against upstream) ----
		// `1 < 0` → false → invalid.
		{
			Code: `<img alt={1 < 0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// `0 | 0` → 0 → falsy → invalid (this is the case that DID diverge
		// before the staticEval rewrite — locks in the fix).
		{
			Code: `<img alt={0 | 0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// `1 === 2` → false → invalid.
		{
			Code: `<img alt={1 === 2} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},

		// ---- i18n falsy patterns ----
		// `undefined && t("key")` short-circuits to undefined → invalid.
		{
			Code: `<img alt={undefined && t("key")} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Optional chain that resolves to undefined when receiver is
		// undefined — but we don't track receiver state; OptionalAccess is
		// modelled as truthy (matches upstream's `data?.alt` extraction
		// returning a non-empty `"data?.alt"` string). The legitimate
		// "missing alt because optional chain returns undefined" case
		// needs a guarded fallback.

		// ---- Multi-line invalid: error reports on the opening element ----
		{
			Code: `<img
				src="x"
			/>`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 1},
			},
		},
		// Two unrelated invalid elements in same source — independent
		// reports.
		{
			Code: `<div><img /><object /></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 6},
				{MessageId: "object", Message: msgObject, Line: 1, Column: 13},
			},
		},

		// ---- Paired-tag form invalid ----
		{
			Code: `<img></img>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 1},
			},
		},

		// Subset elements: only check `img`, skip `<area />`.
		// `<img />` → INVALID (still in elements), `<area />` → no report.
		{
			Code:    `<div><img /><area /></div>`,
			Tsx:     true,
			Options: map[string]interface{}{"elements": []interface{}{"img"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 6},
			},
		},

		// ---- Polymorphic NOT in allow-list: rule does NOT remap ----
		// `<NotAllowed as="img" />` with `polymorphicAllowList: ["Allowed"]`
		// — `NotAllowed` is not in the allow list, so `as` is ignored and
		// the type stays "NotAllowed", which isn't in typesToValidate by
		// default. So no error. (Already covered in valid above.)

		// ---- Polymorphic IN allow-list: rule remaps to img → reports ----
		// `<Allowed as="img" />` with allow list `["Allowed"]` — remaps to
		// "img", then alt is missing → invalid.
		{
			Code: `<Allowed as="img" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Allowed"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 1},
			},
		},

		// ---- Spread inner alt: undefined → invalid ----
		{
			Code: `<img {...{alt: undefined}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},

		// ---- Aria-hidden literal-string "true" → invalid object ----
		// `aria-hidden={"true"}` (StringLiteral inside JsxExpression) →
		// upstream's getPropValue returns "true", isHiddenFromScreenReader
		// uses `=== true` (boolean) so this would NOT match in upstream's
		// strict check. But empirically ESLint reports invalid for this
		// shape — meaning upstream actually compares loosely or extracts
		// the string-literal "true" as the boolean true. We match upstream's
		// observed behavior.
		{
			Code: `<object><div aria-hidden={"true"}>x</div></object>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		// Same but for direct StringLiteral attribute: `aria-hidden="true"`.
		{
			Code: `<object><div aria-hidden="true">x</div></object>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},

		// ---- TS-only wrappers preserving undefined → invalid ----
		// `undefined as any` — assertion doesn't change the underlying
		// undefined identifier; alt is still missing semantically.
		{
			Code: `<img alt={undefined as any} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// `undefined satisfies undefined` — same.
		{
			Code: `<img alt={undefined satisfies undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// `"foo" satisfies string` — even a truthy literal under `satisfies`
		// is null per upstream's TYPES table (no `TSSatisfiesExpression`
		// extractor). staticEval excludes `OEKSatisfies` from
		// skipTransparent, so satisfies-wrapped values fall to the default
		// `jsNull` arm. Locks in alignment after the satisfies-stripping
		// fix.
		{
			Code: `<img alt={"foo" satisfies string} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},

		// ---- input[type] expression evaluation (uses getPropValue, not literal) ----
		// `type={"image" + ""}` — staticEval resolves to "image" → enters
		// inputImage check → no alt → error.
		{
			Code: `<input type={"image" + ""} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		// `type={cond ? "image" : "text"}` — cond is identifier "cond" →
		// truthy → take consequent "image" → check alt → error.
		{
			Code: `<input type={cond ? "image" : "text"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		// Logical `||` resolving to "image".
		{
			Code: `<input type={"" || "image"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		// `&&` short-circuit resolving to "image".
		{
			Code: `<input type={"x" && "image"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		// TS wrapper around "image".
		{
			Code: `<input type={"image" as string} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},

		// ---- object title falsy literals → invalid ----
		{
			Code: `<object title={false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code: `<object title={0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code: `<object title={NaN} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		// Identifier (non-undefined) — getLiteralPropValue returns null → falsy.
		{
			Code: `<object title={someVar} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		// `Math` is in JS_RESERVED but getLiteralPropValue still maps non-
		// undefined identifiers to null → falsy → invalid (different from
		// `<img alt={Math} />` which uses getPropValue and resolves to a fn).
		{
			Code: `<object title={Math} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		// Empty string title → not truthy → invalid.
		{
			Code: `<object title="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		// CallExpression / MemberExpression in title → noop in LITERAL_TYPES → null → falsy.
		{
			Code: `<object title={getTitle()} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code: `<object title={obj.title} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},

		// ---- alt with `"false"` string normalized to boolean false → invalid ----
		{
			Code: `<img alt="false" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img alt="False" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img alt={"false"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		// Verified: `<img alt={false}>` (boolean literal) is also invalid —
		// already covered by the upstream test suite, locking in the
		// downstream `"false"` string match too.

		// ---- aria-hidden expression eval makes child hidden → object error ----
		// `aria-hidden={cond ? true : true}` — both branches true → hidden →
		// no accessible child → object error.
		{
			Code: `<object><div aria-hidden={cond ? true : true}>x</div></object>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		// `aria-hidden={"" || true}` — left falsy → right truthy bool → hidden.
		{
			Code: `<object><div aria-hidden={"" || true}>x</div></object>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		// `aria-hidden={true as any}` — TS-wrapped boolean true → hidden.
		{
			Code: `<object><div aria-hidden={true as any}>x</div></object>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},

		// ---- AwaitExpression / YieldExpression — upstream null → falsy → invalid ----
		// (verified against ESLint with @babel parser)
		{
			Code: `async function fn() { return <img alt={await x} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 30},
			},
		},
		{
			Code: `function* gen() { return <img alt={yield x} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 26},
			},
		},

		// ---- UpdateExpression (`x++`/`++x`/`x--`/`--x`) — upstream NaN → falsy ----
		{
			Code: `<img alt={x++} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img alt={++x} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img alt={x--} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img alt={--x} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},

		// ---- typeof / void → undefined → falsy ----
		{
			Code: `<img alt={typeof x} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img alt={void 0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},

		// ---- Polymorphic resolves to undefined → no replacement → check fires ----
		{
			Code: `<Img as={undefined} />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components":          map[string]interface{}{"Img": "img"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 1},
			},
		},
		// Polymorphic with literal "img" — replaces, then check fires (no alt).
		{
			Code: `<Img as={"img"} />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components":          map[string]interface{}{"Img": "img"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 1},
			},
		},
	})
}
