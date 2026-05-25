// Differential-validated lock-ins for jsx-a11y/img-redundant-alt.
// See img_redundant_alt_upstream_test.go for upstream-mirrored cases.
package img_redundant_alt

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestImgRedundantAltExtras locks in rslint-specific behaviors beyond what
// the upstream eslint-plugin-jsx-a11y test suite exercises:
//
//   - jsx-ast-utils LITERAL_TYPES dispatch coverage (Array / Object / JSX /
//     Regex / BigInt / NewExpression / SequenceExpression / AssignmentExpression
//     / Await / Yield / TaggedTemplate)
//   - TS-only expression wrappers (`as`, `as const`, `!`, `satisfies`,
//     `(x as any)!`)
//   - Spread-argument shape strictness (TS-wrapped object literals do NOT
//     match upstream's `argument.type === 'ObjectExpression'`)
//   - Computed-property-name / numeric-key / shorthand / nested spread
//   - Unicode whitespace classification via ECMAScript `\s` (not Go's
//     `unicode.IsSpace`) — covers U+0085, U+00A0, U+1680, U+2000-U+200A,
//     U+2028, U+2029, U+202F, U+205F, U+3000, U+FEFF
//   - aria-hidden full static-eval coverage (case-insensitive coerce, "yes"
//     not strictly true, conditional that resolves to true, etc.)
//   - Dimension-4 universal edge shapes (parens, position, paired/closing,
//     self-closing without space, member / namespaced / custom-element tags)
//   - Real-world React patterns (FC / forwardRef / memo / class / Array.map
//     / hook / generic / IIFE / object method / nested children / fragment
//     / ternary / spread + multi-aria mix)
//   - Settings shape edge cases (empty / missing-key / etc.)
//
// Each case here was verified empirically against
// `eslint-plugin-jsx-a11y@6.10.2` on a 164-case differential fixture; the
// rslint port produces byte-identical diagnostics.
func TestImgRedundantAltExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ImgRedundantAltRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4 universal-edge lockdowns ----
		// TS-wrapped values — jsx-ast-utils' LITERAL_TYPES maps
		// TSAsExpression / TSNonNullExpression / TypeCastExpression to
		// noop → null, so `getLiteralPropValue` returns null regardless of
		// what the TS wrapper hides. `typeof null !== 'string'`, so the
		// rule skips entirely; the value never reaches the redundancy check.
		{Code: `<img alt={"foo" as string} />`, Tsx: true},
		{Code: `<img alt={"bar" as const} />`, Tsx: true},
		{Code: `<img alt={"baz"!} />`, Tsx: true},
		// CRUCIAL: even when the wrapped value WOULD be redundant, the TS
		// wrapper short-circuits the rule. These three cases differ from
		// the bare-literal forms only in the TS wrapper, and locking them
		// as valid is the regression test that the wrapper transparency
		// stays disabled.
		{Code: `<img alt={"photo" as string} />`, Tsx: true},
		{Code: `<img alt={"picture" as const} />`, Tsx: true},
		{Code: `<img alt={"image"!} />`, Tsx: true},
		// `satisfies` operator — same noop classification.
		{Code: `<img alt={"photo" satisfies string} />`, Tsx: true},
		// Double-wrapped `(x as any)!` — both layers are noop. Locks that
		// LITERAL_TYPES short-circuits at the OUTERMOST TS wrapper without
		// recursing.
		{Code: `<img alt={("photo" as any)!} />`, Tsx: true},
		// Parenthesized literal — extra layer of unwrap.
		{Code: `<img alt={("foo")} />`, Tsx: true},
		{Code: `<img alt={(("foo"))} />`, Tsx: true},
		// Empty alt — empty string isASCII match is false (regex requires at
		// least one ASCII char); falls into the substring path with an empty
		// haystack; no redundant word can match.
		{Code: `<img alt="" />`, Tsx: true},
		// Single space — IS ASCII (0x20), splits to a single empty token,
		// which never equals any redundant word.
		{Code: `<img alt=" " />`, Tsx: true},
		// Boolean true literal — getLiteralPropValue evaluates to boolean
		// true; not a string → skipped.
		{Code: `<img alt={true} />`, Tsx: true},
		// Boolean false literal — likewise.
		{Code: `<img alt={false} />`, Tsx: true},
		// Numeric literal — not a string → skipped.
		{Code: `<img alt={42} />`, Tsx: true},
		// `null` literal — upstream LITERAL_TYPES.Literal maps null → "null"
		// string, but "null" doesn't match any redundant word.
		{Code: `<img alt={null} />`, Tsx: true},
		// String literal "true" / "false" coerce to booleans (not strings) —
		// they get skipped, even though they look like alt text.
		{Code: `<img alt="true" />`, Tsx: true},
		{Code: `<img alt="false" />`, Tsx: true},
		// `<img>` self-closing without space before `/>`.
		{Code: `<img alt="foo"/>`, Tsx: true},
		// Paired open/close form (no children).
		{Code: `<img alt="foo"></img>`, Tsx: true},
		// Hyphenated ASCII alt — "image-thing" is a single token, doesn't
		// equal "image" (split is by whitespace, not by hyphens).
		{Code: `<img alt="image-thing" />`, Tsx: true},
		// Comma-separated ASCII alt — "photo,picture" is a single token.
		{Code: `<img alt="photo,picture" />`, Tsx: true},
		// Period-separated — "photo.jpg" is a single token.
		{Code: `<img alt="photo.jpg" />`, Tsx: true},
		// Member tag (`<Foo.Bar>`) — elementType is "Foo.Bar", not "img".
		{Code: `<Foo.Image alt="photo" />`, Tsx: true},
		// Lowercase member tag — same: not "img".
		{Code: `<foo.img alt="photo" />`, Tsx: true},
		// Namespaced tag (`<svg:image>`) — elementType is "svg:image", not "img".
		{Code: `<svg:image alt="photo" />`, Tsx: true},
		// Custom element with hyphen — not "img".
		{Code: `<my-img alt="photo" />`, Tsx: true},
		// Pure spread — no `alt` attribute literal, FindAttributeByName looks
		// at the spread arg (Identifier `props`, not ObjectLiteral), returns
		// nil. No alt → skip.
		{Code: `<img {...props} />`, Tsx: true},
		// Spread of object literal containing `alt`. FindAttributeByName CAN
		// see this — the inner `alt: "foo"` PropertyAssignment matches.
		// "foo" isn't redundant → valid.
		{Code: `<img {...{alt: "foo"}} />`, Tsx: true},
		// Same shape, but JSX expression / template — exercises the
		// PropertyAssignment-with-template path.
		{Code: "<img {...{alt: `foo bar`}} />", Tsx: true},
		// aria-hidden with non-true static value — upstream's `getPropValue
		// === true` check fails, so the rule fires (visible).
		{Code: `<img aria-hidden="false" alt="foo" />`, Tsx: true},
		// aria-hidden with truthy non-boolean — `=== true` is strict, so
		// truthy string values like "yes" don't equal true → visible →
		// no redundancy here either.
		{Code: `<img aria-hidden="yes" alt="foo" />`, Tsx: true},
		// Comments interleaved with attributes — must not crash extraction.
		{Code: `<img /* leading */ alt="foo" /* trailing */ />`, Tsx: true},
		// JSX inside conditional render — listener fires through children.
		{Code: `function F() { return <div>{cond && <img alt="foo" />}</div>; }`, Tsx: true},
		// JSX inside Array.map.
		{Code: `function L({xs}) { return xs.map(x => <img alt="foo" key={x} />); }`, Tsx: true},
		// JSX inside fragment.
		{Code: `<>{cond && <img alt="foo" />}</>`, Tsx: true},
		// Hidden by aria-hidden + redundant alt — hidden short-circuits, no
		// report. Locks the visibility branch.
		{Code: `<img aria-hidden alt="photo of friend" />`, Tsx: true},
		{Code: `<img aria-hidden={true} alt="photo of friend" />`, Tsx: true},
		// aria-hidden case-insensitive "TRUE" string — staticEval coerces
		// case-insensitively, so `aria-hidden="TRUE"` is hidden too.
		{Code: `<img aria-hidden="TRUE" alt="photo of friend" />`, Tsx: true},
		// Empty options object — fall through to defaults.
		{
			Code:    `<img alt="foo" />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{}},
		},
		// Bare-map options shape (single-option CLI form).
		{
			Code:    `<img alt="foo" />`,
			Tsx:     true,
			Options: map[string]interface{}{"components": []interface{}{"Image"}},
		},
		// Non-ASCII custom word, value contains only CJK — substring match
		// against "イメージ" misses "画像写真".
		{
			Code:    `<img alt="画像写真" />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"words": []interface{}{"イメージ"}}},
		},
		// Pure CJK alt with no custom words — defaults are ASCII-only, so
		// substring path can't hit anything.
		{Code: `<img alt="画像" />`, Tsx: true},
		// `key` prop on img — not a real React semantic for alt, ignored.
		{Code: `<img key="x" alt="foo" />`, Tsx: true},

		// ---- LITERAL_TYPES inheritance edge cases (self-review gap fill) ----
		// Array literal as alt — LITERAL_TYPES.ArrayExpression returns an
		// array (after filtering nulls), typeof !== 'string' → skip.
		{Code: `<img alt={[]} />`, Tsx: true},
		{Code: `<img alt={["photo"]} />`, Tsx: true},
		// Object literal as alt — LITERAL_TYPES.ObjectExpression = noop → null.
		{Code: `<img alt={{x: 1}} />`, Tsx: true},
		// JSX element / fragment as alt — both noop.
		{Code: `<img alt={<span>photo</span>} />`, Tsx: true},
		{Code: `<img alt={<>photo</>} />`, Tsx: true},
		// Regex literal — Literal extractor returns the RegExp object → not a string.
		{Code: `<img alt={/photo/} />`, Tsx: true},
		// BigInt — Literal returns BigInt, not string.
		{Code: `<img alt={42n} />`, Tsx: true},
		// NewExpression — LITERAL_TYPES.NewExpression = noop → null.
		{Code: `<img alt={new String("photo")} />`, Tsx: true},
		// SequenceExpression — LITERAL_TYPES.SequenceExpression = noop → null.
		{Code: `<img alt={(1, "photo")} />`, Tsx: true},
		// await / yield — not in LITERAL_TYPES, default fallback → null.
		{Code: `async function ax() { return <img alt={await x} />; }`, Tsx: true},
		{Code: `function* gx() { return <img alt={yield "photo"} />; }`, Tsx: true},

		// ---- Spread argument shape strictness ----
		// jsx-ast-utils' getProp checks `argument.type === 'ObjectExpression'`
		// strictly. ESTree folds parens at parse time so `({...})` shows up
		// as a bare ObjectExpression, but TS wrappers (`as`, `!`) DON'T
		// fold and DON'T match. We mirror by stripping ONLY parens (not TS)
		// when sniffing the spread arg. Locked against
		// eslint-plugin-jsx-a11y@6.10.2.
		{Code: `<img {...({alt: "photo"} as any)} />`, Tsx: true},
		{Code: `<img {...({alt: "photo"})!} />`, Tsx: true},
		// Spread of array literal — not an ObjectExpression, skip.
		{Code: `<img {...["alt", "photo"]} />`, Tsx: true},
		// Computed-property-name `['alt']: 'photo'` — jsx-ast-utils requires
		// `key.type === 'Identifier'`, ComputedPropertyName fails → skip.
		{Code: `<img {...{['alt']: 'photo'}} />`, Tsx: true},
		// Numeric-key in spread object — not 'alt', skip.
		{Code: `<img {...{0: 'photo'}} />`, Tsx: true},
		// Nested spread — getProp scans only top-level props. Inner alt is
		// hidden under `nested.alt` and ignored.
		{Code: `<img {...{nested: {alt: "photo"}}} />`, Tsx: true},

		// ---- TaggedTemplateExpression (no redundancy) ----
		// jsx-ast-utils LITERAL_TYPES.TaggedTemplateExpression inherits from
		// TYPES (no override) and forwards to the inner template literal.
		{Code: "<img alt={tag`foo`} />", Tsx: true},
		{Code: "<img alt={tag`hello world`} />", Tsx: true},
		{Code: "<img alt={tag`${x}`} />", Tsx: true},

		// ---- Unicode whitespace classification (JS \s, not Go IsSpace) ----
		// NEL (U+0085) — JS does NOT match. The whole alt is one
		// token "photo<NEL>" which doesn't equal any redundant word.
		{Code: "<img alt=\"photo\u0085\" />", Tsx: true},
		// Same value via JsxExpression literal — same outcome.
		{Code: "<img alt={\"photo\\u0085\"} />", Tsx: true},
		// NBSP IS in JS \s but the token before it ("hello") doesn't
		// equal any redundant word either.
		{Code: "<img alt={\"hello\\u00A0\"} />", Tsx: true},

		// ---- Real-world React patterns (valid — no redundancy) ----
		{Code: `const Logo = () => <img alt="Brand badge" />;`, Tsx: true},
		{Code: `function F() { return <img alt="A friendly mascot" />; }`, Tsx: true},
		{Code: `class C extends Component { render() { return <img alt="cat sleeping" />; } }`, Tsx: true},
		{Code: `function L({xs}) { return xs.map(x => <img key={x} alt="A drawing" />); }`, Tsx: true},
		{Code: `<><img alt="ok" /><img alt="fine" /></>`, Tsx: true},
		// Multi-attribute valid: aria-label + alt non-redundant.
		{Code: `<img src="x" srcSet="x@2x" sizes="100vw" alt="A drawing" />`, Tsx: true},
		{Code: `<img onClick={f} onLoad={g} alt="ok" />`, Tsx: true},
		{Code: `<img data-test="x" data-foo="y" alt="ok" />`, Tsx: true},
		{Code: `<img role="presentation" alt="A drawing" />`, Tsx: true},
		{Code: `<img aria-labelledby="x" alt="A drawing" />`, Tsx: true},

		// ---- Settings shape edge cases ----
		// Empty jsx-a11y settings — should not panic, behaves like no settings.
		{
			Code:     `<img alt="foo" />`,
			Tsx:      true,
			Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{}},
		},
		// components map missing target key — falls through, rawType stays
		// the JSX tag string ("Image" → not "img" → skipped).
		{
			Code: `<Image alt="photo" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"Other": "img"},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Position assertions ----
		// Diagnostic position: the rule reports on the JsxOpeningElement /
		// JsxSelfClosingElement node. tsgo's self-closing span includes the
		// trailing `/>`. Locks the column boundary.
		{
			Code: `<img alt="Photo of friend." />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "redundantAlt",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 1, EndColumn: 31,
			}},
		},
		// Multi-line attributes: position spans across lines.
		{
			Code: "<img\n  alt=\"Photo of friend.\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "redundantAlt",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 3, EndColumn: 3,
			}},
		},
		// Paired opening/closing form — listener fires on the opening element.
		{
			Code: `<img alt="Photo of friend."></img>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "redundantAlt",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 1, EndColumn: 29,
			}},
		},

		// ---- Whitespace handling ----
		// Multiple ASCII tokens with leading/trailing whitespace — split by
		// `\s+` produces empty strings at edges; lockdown that those don't
		// confuse the matcher.
		{
			Code:   `<img alt="    photo    " />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Tab + newline whitespace inside alt via JsxExpression so the JS
		// string literal interprets escapes — JSX attribute string syntax
		// itself doesn't process `\t` / `\n` sequences.
		{
			Code:   "<img alt={\"\\tphoto\\n\"} />",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Mixed CJK + ASCII space — falls onto the ASCII (split) path because
		// the space is ASCII; "photo" matches.
		{
			Code:   `<img alt="画像 photo" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Three redundant words in a row — single report (each JSX element
		// fires the listener once). Locks "one report per element."
		{
			Code:   `<img alt="photo image picture" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Leading whitespace + redundant token.
		{
			Code:   `<img alt="    photo" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Trailing whitespace + redundant token.
		{
			Code:   `<img alt="photo    " />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Multi-token sentence with redundant in middle.
		{
			Code:   `<img alt="lovely photo today" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- Unicode whitespace classifications (JS \s) ----
		// Each entry uses JsxExpression + Go `\u` escape so tsgo's JS
		// string-literal parser interprets the escapes (JSX attribute strings
		// don't, and Go source can't carry a raw U+FEFF byte mid-file).
		{
			Code:   "<img alt={\"photo\\u00A0\"} />", // NBSP
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={\"photo\uFEFFbar\"} />", // ZWNBSP / BOM
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={\"photo\\u3000\"} />", // IDEOGRAPHIC SPACE
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={\"photo\\u2028\"} />", // LINE SEPARATOR
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={\"photo\\u2029\"} />", // PARAGRAPH SEPARATOR
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={\"photo\\u202F\"} />", // NARROW NO-BREAK SPACE
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={\"photo\\u205F\"} />", // MEDIUM MATHEMATICAL SPACE
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={\"photo\\u1680\"} />", // OGHAM SPACE MARK
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={\"photo\\u2002\"} />", // EN SPACE
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={\"photo\\u2003\"} />", // EM SPACE
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={\"photo\\u2009\"} />", // THIN SPACE
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={\"photo\\u200A\"} />", // HAIR SPACE
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- TaggedTemplateExpression with redundancy ----
		// LITERAL_TYPES forwards to TemplateLiteral on the inner quasi.
		{
			Code:   "<img alt={tag`photo`} />",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={tag`picture of cat`} />",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<img alt={tag`image doing ${x}`} />",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- AssignmentExpression as alt value ----
		// LITERAL_TYPES.AssignmentExpression is NOT noop — it inherits TYPES
		// and synthesizes `${left} ${operator} ${right}` using the TYPES
		// extract path on each side. Identifier sides stringify to their
		// name, so `x = "photo"` becomes the string "x = photo" and the
		// rule fires. Surfaced by the differential check; locked here.
		{
			Code:   `<img alt={x = "photo"} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<img alt={x += "image"} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<img alt={x ||= "picture"} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- Parenthesized ObjectLiteral spread ----
		// `{...({alt: "photo"})}` — ESTree folds parens, ObjectExpression
		// matches → reported. We strip parens (only) on the spread arg.
		{
			Code:   `<img {...({alt: "photo"})} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Spread of object literal containing redundant `alt`.
		{
			Code:   `<img {...{alt: "photo"}} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Spread of object literal with a template-literal alt.
		{
			Code:   "<img {...{alt: `photo of bar`}} />",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Spread of object literal with multiple props — alt buried inside.
		{
			Code:   `<img {...{src: "x", alt: "photo"}} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Multiple spreads with redundant alt at the end.
		{
			Code:   `<img {...a} {...b} alt="photo" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- Bare-literal forms with redundancy ----
		// Parenthesized redundant literal.
		{
			Code:   `<img alt={("photo")} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Multi-level parenthesized.
		{
			Code:   `<img alt={(("photo"))} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// NoSubstitutionTemplateLiteral.
		{
			Code:   "<img alt={`photo`} />",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- Listener boundary / nesting ----
		// img inside another container — listener fires on inner img.
		{
			Code: `<div><img alt="photo of bar" /></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "redundantAlt",
				Message:   errorMessage,
				Line:      1, Column: 6, EndLine: 1, EndColumn: 32,
			}},
		},
		// JSX inside Array.map callback.
		{
			Code:   `function L({xs}) { return xs.map(x => <img alt="photo" key={x} />); }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// JSX inside class component render method.
		{
			Code:   `class C { render() { return <img alt="picture" />; } }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// JSX inside generic function.
		{
			Code:   `function f<T>(x: T) { return <img alt="image" />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Conditional `&&` render via a JsxExpression child.
		{
			Code:   `function F() { return <div>{cond && <img alt="photo" />}</div>; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Inside a JSX fragment, conditionally rendered.
		{
			Code:   `<>{cond && <img alt="photo" />}</>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- Real-world React patterns (invalid — redundancy in real code) ----
		{
			Code:   `const Logo = () => <img alt="Photo of brand" />;`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const X = forwardRef((p, r) => <img ref={r} alt="photo of x" />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const Y = memo(() => <img alt="photo of memo" />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `function H() { useEffect(() => {}); return <img alt="picture hook" />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const j = (() => <img alt="photo iife" />)();`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const o = { render() { return <img alt="photo method" />; } };`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Ternary render with redundancy on both branches — listener fires
		// on each img.
		{
			Code:   `function G({c}) { return c ? <img alt="photo true" /> : <img alt="picture false" />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError},
		},
		// Nested children — redundant img buried 3 layers deep.
		{
			Code:   `const T = <section><header><img alt="photo header" /></header></section>;`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Two siblings both redundant — listener fires once per element.
		{
			Code:   `<><img alt="photo" /><img alt="picture" /></>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError, expectedError},
		},

		// ---- Multi-attribute mix where alt is redundant ----
		{
			Code:   `<img alt="photo cat" aria-label="Cat" aria-describedby="d1" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<img aria-labelledby="x" alt="photo of y" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<img src="x" srcSet="x@2x" sizes="100vw" alt="photo of bar" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<img onClick={f} onLoad={g} alt="photo handler" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<img data-test="x" data-foo="y" alt="photo data" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<img role="presentation" alt="photo role" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- Unicode + ASCII mix ----
		// Emoji + space + redundant — emoji is non-ASCII but the space is
		// ASCII so we go down the split path; "photo" matches.
		{
			Code:   `<img alt="🐱 photo of cat" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// All-caps redundant token.
		{
			Code:   `<img alt="A NICE PHOTO TODAY" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- polymorphicPropName + components map ----
		// Polymorphic prop redirects `<Box>` → `img`. Locks the
		// polymorphicPropName chain through to the listener.
		{
			Code: `<Box as="img" alt="photo of cat" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Components-map alias points at `img` for a custom JSX tag.
		{
			Code: `<MyImg alt="photo of cat" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"MyImg": "img"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ---- Options shape coverage ----
		// Bare-map options shape — single-option CLI form (no array wrap)
		// must still parse correctly. Locks the GetOptionsMap branch.
		{
			Code:    `<Image alt="photo of cat" />`,
			Tsx:     true,
			Options: map[string]interface{}{"components": []interface{}{"Image"}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Custom words via single-option map shape.
		{
			Code:    `<img alt="Bild of cat" />`,
			Tsx:     true,
			Options: map[string]interface{}{"words": []interface{}{"Bild"}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},
	})
}
