package media_has_caption

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMediaHasCaptionExtras covers behavior the upstream test suite leaves
// unaudited — Dimension 1-4 universal edge shapes, options coverage (bare-map
// vs array-wrapped JSON paths), settings coverage (components map /
// polymorphic prop / polymorphicAllowList), exact position assertions across
// the JSX surface, the listener boundary for nested media elements, and the
// muted / kind literal-extraction edge cases that upstream's "valid"-only
// tests don't pin down.
func TestMediaHasCaptionExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MediaHasCaptionRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Dimension 1: AST tag-shape variants — non-matching forms
		// ============================================================

		// PropertyAccess tag: full type is "UI.audio", not "audio". Locks in
		// that the dotted form is compared, not any segment in isolation.
		{Code: `<UI.audio><track /></UI.audio>`, Tsx: true},
		// Namespaced tag: "svg:audio" ≠ "audio".
		{Code: `<svg:audio><track /></svg:audio>`, Tsx: true},
		// Multi-segment property access.
		{Code: `<A.B.C />`, Tsx: true},
		{Code: `<this.Audio />`, Tsx: true},
		// Case-sensitive comparison: uppercase / mixed-case never match
		// without a components-map alias.
		{Code: `<AUDIO />`, Tsx: true},
		{Code: `<Audio />`, Tsx: true},
		{Code: `<AuDiO />`, Tsx: true},
		{Code: `<VIDEO />`, Tsx: true},
		{Code: `<Video />`, Tsx: true},

		// ---- Element-kind survey: rule is a no-op for every non-media tag ----
		{Code: `<a />`, Tsx: true},
		{Code: `<input />`, Tsx: true},
		{Code: `<Component><track kind="captions" /></Component>`, Tsx: true},

		// `<track>` outside a media element doesn't trigger anything — the
		// rule only runs the caption check when the OUTER element is media.
		{Code: `<div><track kind="subtitles" /></div>`, Tsx: true},
		{Code: `<track />`, Tsx: true},

		// ============================================================
		// Muted exemption — literal-value-extraction edge shapes
		// ============================================================

		// Boolean form (no initializer at all).
		{Code: `<audio muted />`, Tsx: true},
		{Code: `<video muted />`, Tsx: true},

		// Parenthesized boolean — parens unwrap on every layer.
		{Code: `<audio muted={(true)}></audio>`, Tsx: true},
		{Code: `<video muted={((true))}></video>`, Tsx: true},

		// String-literal "true" — jsxAstUtilsLiteralCoerce turns it into bool
		// true, so the strict `=== true` matches. (Mirrors upstream's
		// `getLiteralPropValue("true")` returning boolean true.)
		{Code: `<audio muted="true"></audio>`, Tsx: true},
		{Code: `<video muted="true"></video>`, Tsx: true},
		{Code: `<audio muted={"true"}></audio>`, Tsx: true},

		// Spread attribute carrying muted as an object literal — FindAttribute
		// walks literal spreads, then literalPropValue extracts the value.
		{Code: `<audio {...{muted: true}}></audio>`, Tsx: true},

		// PrefixUnary IS evaluated in LITERAL_TYPES (UnaryExpression isn't
		// noop'd). `!false` → bool true → matches → skip. Mirrors
		// jsx-ast-utils' `extractValueFromLiteralUnaryExpression`.
		{Code: `<audio muted={!false} />`, Tsx: true},

		// ============================================================
		// Captions extraction — kind-prop value edge shapes (valid side)
		// ============================================================

		// JsxExpression-wrapped string literal.
		{Code: `<audio><track kind={"captions"} /></audio>`, Tsx: true},
		{Code: `<video><track kind={"captions"} /></video>`, Tsx: true},

		// Parenthesized string literal.
		{Code: `<audio><track kind={("captions")} /></audio>`, Tsx: true},

		// Mixed-case kind values. EqualFold is case-insensitive — "CAPTIONS",
		// "Captions", "cAPtIoNS" all match.
		{Code: `<audio><track kind="CAPTIONS" /></audio>`, Tsx: true},
		{Code: `<audio><track kind="cAPtIoNS" /></audio>`, Tsx: true},

		// Multiple tracks, only one with captions — `some` short-circuits.
		{Code: `<audio><track kind="subtitles" /><track kind="captions" /></audio>`, Tsx: true},
		{Code: `<audio><track kind="subtitles" /><track kind="metadata" /><track kind="captions" /></audio>`, Tsx: true},

		// Track interspersed with non-track content (text, expressions, other elements)
		// — text and expressions are filtered out by `child.type === 'JSXElement'`.
		{Code: `<audio>Hello <track kind="captions" /></audio>`, Tsx: true},
		{Code: `<audio>{header}<track kind="captions" /></audio>`, Tsx: true},
		{Code: `<audio><source src="a.mp3" /><track kind="captions" /></audio>`, Tsx: true},

		// Self-closing track (the same shape as paired empty track).
		{Code: `<video><track kind="captions" /></video>`, Tsx: true},
		{Code: `<video><track kind="captions"></track></video>`, Tsx: true},

		// ============================================================
		// Options: bare-map AND array-wrapped JSON shapes both work
		// ============================================================

		// Bare-map shape (single-option CLI shape after config.go unwrap).
		{
			Code:    `<MyAudio><track kind="captions" /></MyAudio>`,
			Tsx:     true,
			Options: map[string]interface{}{"audio": []interface{}{"MyAudio"}},
		},
		// Array-wrapped shape (multi-element / rule_tester JSON shape).
		{
			Code:    `<MyAudio><track kind="captions" /></MyAudio>`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"audio": []interface{}{"MyAudio"}}},
		},

		// Empty options object — equivalent to default; non-media tag still passes.
		{Code: `<div />`, Tsx: true, Options: map[string]interface{}{}},

		// nil-valued audio/video/track — falls back to defaults (no extra
		// component names). Default audio/video/track still match.
		{Code: `<audio muted />`, Tsx: true, Options: map[string]interface{}{"audio": nil, "video": nil, "track": nil}},
		{Code: `<audio><track kind="captions" /></audio>`, Tsx: true, Options: map[string]interface{}{"track": nil}},

		// rslint-extension: non-string entries silently dropped by
		// StringSliceOption — `<MyAudio>` doesn't match the (empty post-filter)
		// audio list, so it's NOT considered a media element and the rule
		// short-circuits. (ESLint's JSON schema would reject this at config
		// load, so upstream never sees it.)
		{Code: `<MyAudio />`, Tsx: true, Options: map[string]interface{}{"audio": []interface{}{123, true}}},

		// ============================================================
		// Settings: components map without a matching entry — `<Audio>`
		// stays as "Audio", which is not in the default media list.
		// ============================================================
		{
			Code: `<Audio />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"OtherTag": "audio"},
				},
			},
		},

		// Polymorphic prop with allow-list NOT containing the tag.
		{
			Code: `<Foo as="audio" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Bar"},
				},
			},
		},

		// Reverse aliasing via components map: `<audio>` → "div". Rule no-op.
		{
			Code: `<audio />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"audio": "div"},
				},
			},
		},

		// Polymorphic reverse exemption: `<audio as="div">` → "div" → no
		// match → rule no-op.
		{
			Code: `<audio as="div" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},

		// ============================================================
		// Defensive type handling: settings shapes that ARE truthy but
		// don't match expected types. All silenced (no report).
		// ============================================================
		{Code: `<MyAudio />`, Tsx: true, Settings: map[string]interface{}{"jsx-a11y": "invalid"}},
		{Code: `<MyAudio />`, Tsx: true, Settings: map[string]interface{}{"jsx-a11y": nil}},

		// ============================================================
		// Spread literal edge cases — FindAttributeByName walks ObjectLiteral
		// spreads to find named props. Locks in the parens / TS-wrapper
		// asymmetry so a refactor of attributeInnerExpression can't silently
		// flip these.
		// ============================================================

		// Spread with multi-prop literal: muted is found, order doesn't matter.
		{Code: `<audio {...{muted: true, src: "a.mp3"}}></audio>`, Tsx: true},
		{Code: `<audio {...{src: "a.mp3", muted: true}}></audio>`, Tsx: true},

		// Parens-wrapped ObjectLiteral in spread — parens stripped per
		// FindAttributeByName's `OEKParentheses` unwrap.
		{Code: `<audio {...({muted: true})}></audio>`, Tsx: true},

		// Shorthand spread property: `{...{muted}}` — muted ShorthandProperty
		// resolves to the bound identifier "muted", which under literalPropValue
		// is a non-undefined identifier → null. NOT exactly true → rule should
		// NOT silence. (locked in invalid section below.)

		// ============================================================
		// Children whitespace / comment / format variants
		// ============================================================

		// Whitespace-only text between elements — JsxText is filtered out by
		// the JSXElement-only filter; track child still found.
		{Code: `<audio>  <track kind="captions" />  </audio>`, Tsx: true},
		{Code: "<audio>\n  <track kind=\"captions\" />\n</audio>", Tsx: true},

		// JsxText comment alongside the track.
		{Code: `<audio>{/* localized */}<track kind="captions" /></audio>`, Tsx: true},

		// ============================================================
		// Multi-track arrangements — captions can be at any position
		// ============================================================
		{Code: `<audio><track kind="captions" /><track kind="subtitles" /><track kind="metadata" /><track kind="chapters" /></audio>`, Tsx: true},
		{Code: `<audio><track kind="subtitles" /><track kind="metadata" /><track kind="chapters" /><track kind="captions" /></audio>`, Tsx: true},
		{Code: `<audio><track kind="subtitles" /><track kind="captions" /><track kind="metadata" /></audio>`, Tsx: true},
		// Multiple captions tracks — `some` short-circuits on first hit.
		{Code: `<video><track kind="captions" /><track kind="captions" /></video>`, Tsx: true},

		// ============================================================
		// Polymorphic prop value forms — JsxExpression-wrapped string,
		// template literal — both extract via LITERAL_TYPES and replace
		// rawType.
		// ============================================================
		{
			Code: `<Box as={"audio"} muted={true}></Box>`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},
		{
			Code: "<Box as={`audio`} muted={true}></Box>",
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},

		// ============================================================
		// Multi-line valid form — captions on its own line.
		// ============================================================
		{
			Code: "<video\n  controls\n  src=\"video.mp4\"\n>\n  <track kind=\"captions\" src=\"caps.vtt\" />\n</video>",
			Tsx:  true,
		},

		// ============================================================
		// Empty option arrays for audio/video/track — explicit empty list,
		// behaves the same as absent (no extension; default media still match).
		// ============================================================
		{Code: `<audio><track kind="captions" /></audio>`, Tsx: true,
			Options: map[string]interface{}{"audio": []interface{}{}, "video": []interface{}{}, "track": []interface{}{}}},

		// ============================================================
		// Duplicate `muted` attributes — JSX strict mode disallows but tsgo
		// accepts. FindAttributeByName returns the FIRST matching attribute,
		// so `<audio muted={true} muted={false}>` → first is true → silence.
		// (Differential-verified against eslint-plugin-jsx-a11y v6.10.2.)
		// ============================================================
		{Code: `<audio muted={true} muted={false}></audio>`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Position assertions: paired form — listener fires on the OPENING
		// element only. `<audio>` is 7 characters → EndColumn 8.
		// ============================================================
		{
			Code: `<audio>scrolling</audio>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mediaHasCaption",
				Message:   "Media elements such as <audio> and <video> must have a <track> for captions.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 8,
			}},
		},

		// Empty body still reports.
		{
			Code:   `<audio></audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Position assertions: self-closing form — `<audio />` is 10
		// characters → EndColumn 11.
		// ============================================================
		{
			Code: `<audio />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mediaHasCaption",
				Message:   "Media elements such as <audio> and <video> must have a <track> for captions.",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 10,
			}},
		},

		// Multi-line element — position must span the entire opening
		// self-closing tag.
		{
			Code: "<video\n  width=\"320\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mediaHasCaption",
				Message:   "Media elements such as <audio> and <video> must have a <track> for captions.",
				Line:      1, Column: 1, EndLine: 3, EndColumn: 3,
			}},
		},

		// ============================================================
		// Dimension 2: same-kind / cross-kind media nesting — both the
		// outer and inner element fire (each fails the caption check
		// independently).
		// ============================================================
		{
			Code: `<video><audio /></video>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "mediaHasCaption", Message: "Media elements such as <audio> and <video> must have a <track> for captions.",
					Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				{MessageId: "mediaHasCaption", Message: "Media elements such as <audio> and <video> must have a <track> for captions.",
					Line: 1, Column: 8, EndLine: 1, EndColumn: 17},
			},
		},
		{
			Code: `<audio><video><track kind="captions" /></video></audio>`,
			Tsx:  true,
			// Outer audio fails (no track child of audio); inner video
			// satisfies via captions track and is silent.
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Muted exemption — literal-value-extraction edge shapes (invalid)
		// ============================================================

		// String "true" coerces to bool true and SKIPS — but the
		// non-string-literal forms below do NOT.
		{Code: `<audio muted={undefined} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<audio muted={null} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<audio muted={someVar} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<audio muted={1} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<audio muted="anything" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Conditional / Logical / Unary all map to noop → null in
		// LITERAL_TYPES; not exactly true → reports.
		{Code: `<audio muted={cond ? true : false} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// PrefixUnary that evaluates to false (NOT true) — extracted via
		// staticEvalUnary, so `!true` → bool false → not exactly true → reports.
		{Code: `<audio muted={!true} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// TS as-assertion: literalPropValue rejects TSAsExpression → null →
		// not exactly true → reports.
		{Code: `<audio muted={true as any} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// No-substitution template `\`true\`` is treated as TemplateLiteral
		// (not Literal), so jsxAstUtilsLiteralCoerce is NOT applied — value
		// stays jvString "true" → not exactly bool true → reports.
		{Code: "<audio muted={`true`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// Captions extraction — kind-prop value edge shapes (invalid)
		// ============================================================

		// `kind` is dynamic identifier — literalPropValue → null → "" →
		// fails the "captions" match.
		{Code: `<audio><track kind={someVar} /></audio>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `kind` is conditional — noop → null.
		{Code: `<audio><track kind={cond ? "captions" : "subtitles"} /></audio>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Empty string kind — `"" === "captions"` is false.
		{Code: `<audio><track kind="" /></audio>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<audio><track kind={""} /></audio>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `kind="true"` — Literal coerces "true" → boolean true → not a
		// string → LiteralPropStringValue returns ("", false) → fails the
		// captions match. (Mirrors upstream's `(true || '').toLowerCase()`
		// throwing TypeError; we silently treat as no caption.)
		{Code: `<audio><track kind="true" /></audio>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// `kind` value carrying TS wrapper — literalPropValue rejects
		// TSAsExpression → null → empty string → fails.
		{Code: `<audio><track kind={"captions" as string} /></audio>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// Boolean form `<track kind />` — extractValue maps to bool true →
		// LiteralPropStringValue returns ("", false) → fails the captions
		// match.
		{Code: `<audio><track kind /></audio>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// Multiple tracks, none with captions.
		{
			Code:   `<audio><track kind="subtitles" /><track kind="metadata" /></audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// JsxFragment child — JSXFragment is NOT 'JSXElement' upstream, so
		// it's filtered out. Audio still has no track → reports.
		{
			Code:   `<audio><><track kind="captions" /></></audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// Track wrapped in a JsxExpression (e.g. `{cond && <track />}`) —
		// the child is a JsxExpression, NOT a JsxElement / JsxSelfClosingElement,
		// so it doesn't satisfy the filter. Audio has no track child → reports.
		{
			Code:   `<audio>{cond && <track kind="captions" />}</audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// Spread attributes don't carry muted → reports.
		{
			Code:   `<audio {...props}></audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Spread literal carrying muted=false → not exactly true → reports.
		{
			Code:   `<audio {...{muted: false}}></audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Options: array-wrapped shape with audio extension
		// ============================================================
		{
			Code:    `<MyAudio />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"audio": []interface{}{"MyAudio"}}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Track extension via options.track.
		{
			Code:    `<audio><MyTrack kind="subtitles" /></audio>`,
			Tsx:     true,
			Options: map[string]interface{}{"track": []interface{}{"MyTrack"}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Mixed: extension to BOTH audio AND track, with subtitles only.
		{
			Code:    `<MyAudio><MyTrack kind="subtitles" /></MyAudio>`,
			Tsx:     true,
			Options: map[string]interface{}{"audio": []interface{}{"MyAudio"}, "track": []interface{}{"MyTrack"}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Extension via options.video — captures `<MyVideo />`.
		{
			Code:    `<MyVideo />`,
			Tsx:     true,
			Options: map[string]interface{}{"video": []interface{}{"MyVideo"}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},
		// rslint extension: mixed-type array — non-string entries silently
		// dropped, "MyAudio" still honored. Reports.
		{
			Code:    `<MyAudio />`,
			Tsx:     true,
			Options: map[string]interface{}{"audio": []interface{}{"MyAudio", 123}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Settings: components map maps `<MyAudio>` to "audio"
		// ============================================================
		{
			Code: `<MyAudio />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"MyAudio": "audio"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code: `<MyVideo />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"MyVideo": "video"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// `<audio>` aliased to "audio" (self-map) — still reports.
		{
			Code: `<audio />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"audio": "audio"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// Polymorphic without allow-list — every truthy `as` value replaces
		// rawType, so `<Foo as="audio" />` reports.
		{
			Code: `<Foo as="audio" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// Polymorphic + components combo: polymorphic replaces FIRST, then
		// components looks up the (post-polymorphic) name. So `<Foo as="audio">`
		// with `components: {Foo: "div"}` still reports — polymorphic turns
		// rawType into "audio", and there's no "audio" key in components.
		{
			Code: `<Foo as="audio" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components":          map[string]interface{}{"Foo": "div"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// Polymorphic + components CHAIN: Foo → Bar (polymorphic) → audio
		// (components). Locks in that components looks up the post-polymorphic
		// name.
		{
			Code: `<Foo as="Bar" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName": "as",
					"components":          map[string]interface{}{"Bar": "audio"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Real-world component patterns — listener fires inside common
		// React shapes (function component, map callback, conditional, etc.).
		// ============================================================
		{
			Code:   `function Banner() { return <video />; }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const x = items.map(item => <video key={item.id} />)`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `const x = cond ? <audio /> : <div />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code: `class C { render() { return <div><audio /><video /></div>; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError,
				expectedError,
			},
		},

		// ============================================================
		// Listener boundary: media inside non-media wrapper (each fires
		// independently from any wrapping element).
		// ============================================================
		{
			Code:   `<div><audio /></div>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<><audio /></>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Spread literal edge cases (invalid side)
		// ============================================================
		// Non-literal spread is opaque — muted is unreachable through
		// `{...someObj}`. Mirrors jsx-ast-utils' getProp behavior.
		{
			Code:   `<audio {...someObj}></audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Shorthand spread property `{...{muted}}` — value is the bound
		// identifier "muted", which under literalPropValue is a non-undefined
		// identifier → null. Not exactly true → reports.
		{
			Code:   `<audio {...{muted}}></audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Children wrapping boundary — JsxExpression / array literal /
		// comment-only / whitespace-only wrapping all FAIL the
		// JSXElement-only filter, so the audio still has no track child.
		// ============================================================

		// Track in unconditional JsxExpression — `{<track ...>}`.
		{
			Code:   `<audio>{<track kind="captions" />}</audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Track in array literal in JsxExpression — `{[<track .../>]}`.
		{
			Code:   `<audio>{[<track kind="captions" key="x" />]}</audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Comment-only children — JsxExpression payload only contains a
		// comment. No track, no caption.
		{
			Code:   `<audio>{/* TODO add captions */}</audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Whitespace-only children — JsxText whitespace is filtered out
		// by the JSXElement-only filter; no track child found.
		{
			Code:   `<audio>     </audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<audio>\n\n\n</audio>",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Cross-form nesting: outer media without its own track child fires;
		// inner media with valid captions stays silent. Lock that the outer
		// listener does NOT walk into nested media's children for tracks.
		// ============================================================
		{
			Code:   `<video><audio><track kind="captions" /></audio></video>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Multi-line invalid form — position spans the entire opening tag.
		// `<video controls src="…">` opening tag spans lines 1-4; tsgo
		// EndColumn is exclusive, so `>` on line 4 col 1 → EndColumn 2.
		// ============================================================
		{
			Code: "<video\n  controls\n  src=\"video.mp4\"\n>\n</video>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mediaHasCaption",
				Message:   "Media elements such as <audio> and <video> must have a <track> for captions.",
				Line:      1, Column: 1, EndLine: 4, EndColumn: 2,
			}},
		},

		// ============================================================
		// kind extraction edge: dynamic references all map to noop → null in
		// LITERAL_TYPES → empty-string fallback → fail captions match.
		// ============================================================
		{
			Code:   `<audio><track kind={Foo.Bar} /></audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<audio><track kind={getKind()} /></audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<audio><track kind={someKind || "subtitles"} /></audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Real-world component patterns (extended) — listener fires inside
		// the React idioms users actually write.
		// ============================================================
		// React.memo wrap.
		{
			Code:   `const Audio = React.memo(() => <audio />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// React.forwardRef wrap with spread props.
		{
			Code:   `const Audio = React.forwardRef((props, ref) => <audio ref={ref} {...props} />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// React.Suspense fallback containing media.
		{
			Code:   `<React.Suspense fallback={<video />}>{children}</React.Suspense>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Switch-style rendering — both arms fire.
		{
			Code: `function R({type}) { switch(type) { case 'a': return <audio />; case 'v': return <video />; default: return null; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError,
				expectedError,
			},
		},
		// Iterator chain: `.filter().map(...)` returning JSX.
		{
			Code:   `const x = items.filter(i => i.kind === "media").map(i => <video key={i.id} src={i.src} />)`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// Array of JSX literals.
		{
			Code: `const items = [<audio key="1" />, <video key="2" />];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				expectedError,
				expectedError,
			},
		},
		// JSX as function argument.
		{
			Code:   `wrap(<audio />);`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		// JSX as object property value.
		{
			Code:   `const x = { content: <audio /> };`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},

		// ============================================================
		// Deep nesting (3 levels): each opening tag fires independently.
		// Position assertions lock in tsgo column counting on adjacent JSX.
		// `<video>` is 7 chars, so outer cols 1-8, middle 8-15, inner 15-24
		// (`<video />` is 9 chars).
		// ============================================================
		{
			Code: `<video><video><video /></video></video>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "mediaHasCaption", Message: "Media elements such as <audio> and <video> must have a <track> for captions.",
					Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				{MessageId: "mediaHasCaption", Message: "Media elements such as <audio> and <video> must have a <track> for captions.",
					Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				{MessageId: "mediaHasCaption", Message: "Media elements such as <audio> and <video> must have a <track> for captions.",
					Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
			},
		},

		// ============================================================
		// Mixed JsxExpression children — even if a track is inside an
		// expression, it still doesn't satisfy the filter, so the audio
		// is reported.
		// ============================================================
		{
			Code:   `<audio>{props.captions}{props.children}</audio>`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
	})
}
