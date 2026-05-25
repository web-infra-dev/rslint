package anchor_ambiguous_text

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// polymorphicSettings exercises the `polymorphicPropName` settings entry —
// `<Foo as="a">` should be treated as `<a>` after polymorphic-prop resolution
// (no componentMap fallback). Upstream's own test file doesn't exercise this
// branch for anchor-ambiguous-text, so we lock it in here against
// GetElementType regressions.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// emptyWordsOpts is the canonical "disable the rule" idiom: an explicit empty
// list REPLACES the defaults (the `||` fallback only fires when `words` is
// undefined), so no phrase can ever be ambiguous. The rule short-circuits at
// listener registration time when the active wordlist is empty.
var emptyWordsOpts = map[string]interface{}{
	"words": []interface{}{},
}

// customWordsOnlyOpts exercises the "options replace defaults entirely"
// semantic — the user adds a phrase but the original defaults stop being
// ambiguous.
var customWordsOnlyOpts = map[string]interface{}{
	"words": []interface{}{"custom phrase"},
}

// TestAnchorAmbiguousTextExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 shape / tsgo AST quirk it
// covers, so future refactors of either the rule or its shared helpers can't
// silently regress these without breaking a named lock-in.
func TestAnchorAmbiguousTextExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AnchorAmbiguousTextRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: container forms ----
		// Self-closing anchor — KindJsxSelfClosingElement listener fires, no
		// children, accessible text is "" which never appears in the default
		// wordlist.
		{Code: `<a />`, Tsx: true},
		// Paired empty anchor — KindJsxOpeningElement listener with empty
		// Children.Nodes. Same outcome via the other listener kind.
		{Code: `<a></a>`, Tsx: true},
		// Whitespace-only text — trimmed to "" before the lookup, so even an
		// "ambiguous" word made of whitespace wouldn't match by default.
		{Code: `<a>   </a>`, Tsx: true},

		// ---- Dimension 4: tag-name forms ----
		// Uppercase HTML tag is treated as a React component — nodeType "A"
		// is NOT in `['a']`, rule short-circuits. Upstream `typesToValidate`
		// is case-sensitive.
		{Code: `<A>here</A>`, Tsx: true},
		// Namespaced JSX (e.g. `<svg:a>`) resolves to nodeType "svg:a"; not
		// matched.
		{Code: `<svg:a>here</svg:a>`, Tsx: true},
		// Member-access tag resolves to a dotted nodeType; not matched.
		{Code: `<Foo.a>here</Foo.a>`, Tsx: true},

		// ---- Locks in upstream's switch: NON-Literal/JSXText children
		//      contribute "". JsxExpressionContainer with a StringLiteral
		//      payload is NOT a JSXText / Literal in upstream's `case`
		//      branches, so even the literal "click here" inside `{...}`
		//      falls to `return ''` and the anchor's accessible text is "".
		{Code: `<a>{"click here"}</a>`, Tsx: true},
		// Same shape with a bare Identifier — upstream's switch handles
		// JsxExpression default → "".
		{Code: `<a>{here}</a>`, Tsx: true},

		// ---- JsxFragment children contribute "" (upstream's switch has no
		//      JSXFragment case). Even an ambiguous phrase inside a fragment
		//      doesn't make the anchor's text match.
		{Code: `<a><>here</></a>`, Tsx: true},

		// ---- aria-label empty string → falls back to children. Children
		//      "read this" is not ambiguous → valid. Locks in upstream's
		//      `if (ariaLabel)` falsy branch for the empty-string case.
		{Code: `<a aria-label="">read this</a>`, Tsx: true},

		// ---- nodeType-resolved child via componentMap — img-tagged custom
		//      component with empty alt + no children → still valid.
		{Code: `<a><img /></a>`, Tsx: true},

		// ---- Polymorphic prop resolves to a NON-anchor → rule skips.
		//      Without this the polymorphicProp branch in GetElementType
		//      could be silently swapped for the componentMap branch.
		{Code: `<Foo as="div">here</Foo>`, Tsx: true, Settings: polymorphicSettings},

		// ---- Options override: explicit empty list disables the rule.
		//      `words: []` REPLACES defaults; the listener registration
		//      short-circuits and no element is ever inspected.
		{Code: `<a>click here</a>`, Tsx: true, Options: emptyWordsOpts},

		// ---- Options override: defaults gone, the new wordlist is the only
		//      thing checked. The previously-ambiguous "here" is now fine.
		{Code: `<a>here</a>`, Tsx: true, Options: customWordsOnlyOpts},

		// ---- Upstream's `typesToValidate` is hard-coded to `['a']` — the
		//      rule does NOT accept a `components` option (unlike
		//      anchor-has-content / anchor-is-valid). A `components` key in
		//      Options is silently ignored. `<CustomA>` resolves to nodeType
		//      "CustomA" and is never inspected — locks in the
		//      "components option does NOT extend typesToValidate" invariant.
		{
			Code:    `<CustomA>here</CustomA>`,
			Tsx:     true,
			Options: map[string]interface{}{"components": []interface{}{"CustomA"}},
		},

		// ---- aria-label takes precedence at every level. An anchor-level
		//      aria-label hides not just children text but also a nested
		//      img's alt. Locks in the "aria-label early return wins over
		//      everything below" semantic at the anchor level.
		{Code: `<a aria-label="documentation"><img alt="click here" /></a>`, Tsx: true},

		// ---- whitespace-only aria-label: LiteralPropStringValue returns
		//      ("   ", true) (non-empty after the `v != ""` guard);
		//      standardizeSpaceAndCase trims it to "". The early-return
		//      fires with "" — which doesn't match any default-word and
		//      thus the rule passes EVEN THOUGH the children would be
		//      ambiguous. Mirrors upstream's `if (ariaLabel)` on the truthy
		//      whitespace string, followed by the same normalize. This is a
		//      common foot-gun in real code and the test locks the
		//      upstream-aligned behavior in place.
		{Code: `<a aria-label="   ">click here</a>`, Tsx: true},

		// ---- alt on img with an empty-string value falls through (upstream
		//      treats `altTag` as falsy for ""). With NO outer text, the
		//      anchor's accessible text is "" → valid.
		{Code: `<a><img alt="" /></a>`, Tsx: true},

		// ---- alt on img with a non-literal Identifier value (`alt={altText}`)
		//      can't be statically resolved; falls through. No outer text →
		//      accessible text "" → valid.
		{Code: `<a><img alt={altText} /></a>`, Tsx: true},

		// ---- alt on img with a non-string literal (`alt={123}`) — upstream
		//      LITERAL_TYPES.Literal stringifies to "123" but the truthy
		//      branch needs altTag to be coerced as STRING; jsx-ast-utils'
		//      getLiteralPropValue returns 123 (number), and the condition
		//      `elementType === 'img' && altTag` (truthy) IS taken upstream
		//      → returns standardizeSpaceAndCase(123) which throws TypeError
		//      on `.trim()`. rslint's LiteralPropStringValue returns
		//      ("", false) for non-string literals, so we fall through
		//      gracefully — locks in our better-than-upstream resilience for
		//      this malformed shape.
		{Code: `<a><img alt={123} /></a>`, Tsx: true},

		// ---- Real-world i18n: i18n libraries render translated text via
		//      a function call inside a JsxExpression. The container's
		//      payload is opaque to static analysis, so it contributes ""
		//      — even when the i18n key suggests an ambiguous result. This
		//      is the upstream-aligned "we don't follow the expression"
		//      behavior.
		{Code: `<a>{t('learn-more')}</a>`, Tsx: true},
		{Code: `<a>{i18n.t('here')}</a>`, Tsx: true},

		// ---- Real-world variable passthrough: `<a>{children}</a>` is the
		//      idiomatic "render whatever the parent passed". JsxExpression
		//      → "" → valid.
		{Code: `<a>{children}</a>`, Tsx: true},

		// ---- Real-world custom i18n component child: anything that resolves
		//      to a non-img, non-anchor element with no static children
		//      contributes "" — the child component is opaque.
		{Code: `<a><Trans i18nKey="ambiguous-but-opaque" /></a>`, Tsx: true},

		// ---- Sibling aria-hidden coverage: when ALL non-text siblings are
		//      hidden, the anchor's accessible text is "" — valid even when
		//      the hidden text would have been ambiguous.
		{Code: `<a><span aria-hidden>click here</span><span aria-hidden>learn more</span></a>`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- jsx-ast-utils `getProp` default `ignoreCase: true`: attribute
		//      names match case-insensitively. Locks in FindAttributeByName
		//      semantics for aria-label specifically.
		{
			Code: `<a ARIA-LABEL="click here">x</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<a aria-LABEL="click here">x</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- HTML entity decoding on attribute string values. tsgo
		//      preserves `&#…;` source on StringLiteral.Text; jsx-ast-utils
		//      reads the DECODED text. directAttributeStringValue routes
		//      through jsxtransforms.DecodeEntities to realign. `&#104;` →
		//      "h", so `&#104;ere` → "here" → ambiguous.
		{
			Code: `<a aria-label="&#104;ere">x</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- JsxExpression-wrapped string literal as the aria-label value
		//      — LiteralPropStringValue must unwrap the JsxExpression and
		//      read the inner StringLiteral.
		{
			Code: `<a aria-label={"click here"}>x</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},
		// NoSubstitutionTemplateLiteral (back-tick form without `${…}`) —
		// same shape as a regular string but a different tsgo Kind. Locks
		// in LiteralPropStringValue's second arm.
		{
			Code: "<a aria-label={`click here`}>x</a>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- aria-label = `undefined` Identifier. Upstream's
		//      `getLiteralPropValue` returns null for this → falsy → falls
		//      through to children. We mirror via LiteralPropStringValue
		//      returning ("", false). Children "here" → ambiguous.
		{
			Code: `<a aria-label={undefined}>here</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},
		// aria-label = an arbitrary runtime variable — upstream's
		// LITERAL_TYPES.Identifier maps to null → fall through. We must NOT
		// take the variable's NAME ("someVar") as the label value.
		{
			Code: `<a aria-label={someVar}>here</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},
		// aria-label = empty string — `if (ariaLabel)` falsy → fall through.
		// LiteralPropStringValue returns ("", true) so the `v != ""` guard
		// in step 1 of getAccessibleChildText is what filters this out.
		{
			Code: `<a aria-label="">here</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},
		// Boolean-attribute form `<a aria-label>here</a>`. Upstream's
		// `getLiteralPropValue` returns boolean `true` and `if (true)`
		// proceeds to `.trim()` it — which throws a TypeError at runtime.
		// rslint instead treats boolean form as "no literal string value",
		// falls through to children, and reports correctly. This locks in
		// our graceful-degradation behavior; without the test, a future
		// refactor that "fixes" us to mirror upstream's crash would not be
		// caught.
		{
			Code: `<a aria-label>here</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Recursion: case-fold is applied at every level. Inner span
		//      standardizes "HERE" → "here"; outer joins and standardizes
		//      again to "here".
		{
			Code: `<a><span>HERE</span></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},
		// Recursion: punctuation stripping is applied at every level. Inner
		// span "HERE." → "here"; outer "here". Without the recursive
		// normalize, "here." would survive to the outer level and miss the
		// Set.
		{
			Code: `<a><span>HERE.</span></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},
		// Deeper nesting: <a><div><span>click</span> here</div></a>. Both
		// levels of getAccessibleChildText recurse normally; the outer
		// anchor's accessible text is "click here". Locks in multi-level
		// recursion.
		{
			Code: `<a><div><span>click</span> here</div></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Polymorphic prop resolves to an anchor → rule fires.
		{
			Code:     `<Foo as="a">here</Foo>`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- img with paired form `<a><img>click here</img></a>`. HTML5
		//      treats img as a void element, but JSX accepts the paired
		//      form. elementType resolves to "img"; without `alt`, the alt
		//      shortcut is skipped, and the rule walks children — giving
		//      "click here". Locks in the upstream `(elementType === 'img'
		//      && altTag)` condition with an absent alt.
		{
			Code: `<a><img>click here</img></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Options override: a NEW phrase entirely. Default phrases no
		//      longer count, but "custom phrase" does. The diagnostic's
		//      wordsList reflects the override (Message changes accordingly).
		{
			Code:    `<a>custom phrase</a>`,
			Tsx:     true,
			Options: customWordsOnlyOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: errorMessage([]string{"custom phrase"}), Line: 1, Column: 1},
			},
		},

		// ---- Capitalized aria-label gets case-folded by the
		//      standardizeSpaceAndCase pass on the aria-label value itself
		//      (not just on children-joined text). Locks in the
		//      normalization step on the aria-label branch.
		{
			Code: `<a aria-label="CLICK HERE">x</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Multi-child sequence: JsxText + JsxElement + JsxText collapse
		//      via the outer join + standardize. Different layout from the
		//      upstream `<a><span>click</span> here</a>` (which is JsxElement
		//      first, JsxText second). This is JsxText first, JsxElement
		//      second.
		{
			Code: `<a>click<span> here</span></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Real-world i18n + dynamic insert: text + JsxExpression + text.
		//      JsxText "click " + JsxExpression{Identifier} ("" contribution)
		//      + JsxText " here" — joined with " " separators, collapsed to
		//      "click here". This is the single most common real-world
		//      "ambiguous link" pattern: a translation that interpolates a
		//      variable around the ambiguous phrase.
		{
			Code: `<a>click {name} here</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- `<input type="hidden" />` child contributes "" via the second
		//      arm of upstream's `isHiddenFromScreenReader`: when the
		//      element type is uppercase-"INPUT" AND its `type` literal is
		//      uppercase-"HIDDEN", treat it as hidden. Locks in the
		//      INPUT-arm of IsHiddenFromScreenReaderFromTagAttrs.
		{
			Code: `<a><input type="hidden" />learn more</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- alt on img is non-literal (`alt={altText}`): falls through to
		//      the children path on the inner element (no children → ""),
		//      and the OUTER text "learn more" contributes via the anchor's
		//      child walk. Locks in the "alt fallthrough doesn't mask outer
		//      ambiguous text" semantic.
		{
			Code: `<a><img alt={altText} />learn more</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- alt="" on img + ambiguous outer text. img alt="" is falsy →
		//      falls through to children (empty) → ""; outer "click here"
		//      → ambiguous. Same shape as the above with an empty literal
		//      instead of a runtime Identifier.
		{
			Code: `<a><img alt="" />click here</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- aria-label on a nested img child wins over its alt. Outer
		//      anchor sees the nested element's aria-label-resolved text.
		//      Locks in aria-label-precedence-over-alt at the recursion
		//      step (different from the anchor-level early return tested
		//      via the upstream suite).
		{
			Code: `<a><img aria-label="here" alt="not here" /></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- aria-label content gets standardize-normalized too: leading
		//      and trailing spaces are trimmed, internal runs are collapsed
		//      to a single space. Without normalization on the aria-label
		//      branch, this would survive as "click    here" and miss the
		//      Set lookup.
		{
			Code: `<a aria-label="click    here">x</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Punctuation stripping on the aria-label branch (not just
		//      children). Locks in standardizeSpaceAndCase running on the
		//      aria-label value before the lookup.
		{
			Code: `<a aria-label="learn more!">x</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Multiple sentence-ending punctuation chars stripped in one
		//      pass (regex `g` flag). `here?!` → "here", `here..!?;,:` → "here".
		{
			Code: `<a>here?!</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<a>here..!?;,:</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Combined: leading/trailing whitespace + uppercase + trailing
		//      punctuation. Exercises ALL four standardize steps at once on
		//      a single string.
		{
			Code: `<a>  LEARN MORE.  </a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Deep nesting (3+ levels) — accessible text bubbles up through
		//      every intermediate JsxElement.
		{
			Code: `<a><div><span><b>here</b></span></div></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Self-closing empty inner element doesn't suppress sibling
		//      text. `<a><i />click here</a>` → contributions ["", "click
		//      here"] → "click here". Variant of the upstream-suite
		//      `<a>a<i></i> link</a>` with a self-closing inner.
		{
			Code: `<a><i />click here</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Multiple aria-hidden siblings with ONE visible sibling — the
		//      visible sibling's text still bubbles up unimpeded. Locks in
		//      "aria-hidden suppresses only the hidden element, not its
		//      siblings".
		{
			Code: `<a><span aria-hidden>X</span><span aria-hidden>Y</span>here</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- `<a alt="">click here</a>` — alt on a non-img is completely
		//      ignored (not even consulted for falsy-bypass purposes); the
		//      rule walks children directly. Variant of the upstream
		//      `<a alt="tutorial">click here</a>` with an empty alt to
		//      double-check the elementType gate is the deciding factor.
		{
			Code: `<a alt="">click here</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- aria-label early-return on the anchor wins over a dynamic
		//      child expression. The child contributes "" anyway, but the
		//      anchor's aria-label is the ambiguous value being reported.
		{
			Code: `<a aria-label="click here">{children}</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Unicode whitespace collapse: two NBSP characters (U+00A0)
		//      between "learn" and "more" must be treated as a single
		//      whitespace run and collapsed to " ". Locks in Go-vs-JS
		//      regex parity: Go's plain `\s\s+` is ASCII-only and would
		//      miss this; our `[\s\v\p{Z}]{2,}` covers NBSP via `\p{Z}`
		//      (Zs category). Source uses Go's `" "` to inject
		//      actual NBSP runes — tsgo preserves them on JsxText.Text,
		//      mirroring Babel's parse.
		{
			Code: "<a>learn  more</a>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- Unicode whitespace trim: a single leading/trailing NBSP is
		//      already removed by `strings.TrimSpace` (which uses
		//      `unicode.IsSpace`, the same Unicode `White_Space` set
		//      JS `String.prototype.trim()` consumes). This case verifies
		//      trim alignment independently of the collapse regex.
		{
			Code: "<a> here </a>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- HTML entity decode in JsxText children: named entity
		//      `&nbsp;` → U+00A0. Babel exposes JSXText.value with
		//      entities decoded; tsgo preserves raw `&…;` source on
		//      JsxText.Text and we route through
		//      `jsxtransforms.DecodeEntities` to realign. After decode +
		//      trim, the text is "here" (the NBSP is trimmed away),
		//      which is ambiguous.
		{
			Code: `<a>here&nbsp;</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- HTML entity decode: decimal numeric entity. `&#104;` is
		//      ASCII "h". After decode → "here" → ambiguous. Locks the
		//      decimal arm of `DecodeEntities`.
		{
			Code: `<a>&#104;ere</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- HTML entity decode: hex numeric entity. `&#x68;` is also
		//      ASCII "h". Locks the hex arm of `DecodeEntities`.
		{
			Code: `<a>&#x68;ere</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},

		// ---- HTML entity decode inside a nested element. Entity decode
		//      must happen at every recursion level, not only at the
		//      anchor's direct children — locks in the recursive call
		//      path through getAccessibleChildText.
		{
			Code: `<a><span>here&nbsp;</span></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1},
			},
		},
	})
}
