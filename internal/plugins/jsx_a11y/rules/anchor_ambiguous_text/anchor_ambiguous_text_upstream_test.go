package anchor_ambiguous_text

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// defaultMessage is the diagnostic emitted under the rule's default wordlist.
// Computed once so every upstream-mirrored invalid case stays in lock-step
// with the message helper.
var defaultMessage = errorMessage(defaultAmbiguousWords)

// linkComponentSettings mirrors upstream's
// `settings: { 'jsx-a11y': { components: { Link: 'a' } } }`. The componentMap
// remap promotes `<Link>` to be checked as an anchor; without it, custom
// component tags fall to the `nodeType !== 'a'` short-circuit and the rule
// never inspects their text.
var linkComponentSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Link": "a",
		},
	},
}

// imageComponentSettings remaps `<Image>` to `img`. Exercises the
// nested-element alt-substitution branch where the anchor itself stays `<a>`
// but its child component should be treated as an img for accessible-text
// purposes.
var imageComponentSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Image": "img",
		},
	},
}

// disablingDefaultsOpts overrides the default wordlist. Used by the upstream
// "options replace defaults, not extend" lock-in cases.
var disablingDefaultsOpts = map[string]interface{}{
	"words": []interface{}{"disabling the defaults"},
}

var disablingDefaultsWithComponentsOpts = map[string]interface{}{
	"words": []interface{}{"disabling the defaults with components"},
}

var customWordsOpts = map[string]interface{}{
	"words": []interface{}{"a disallowed word"},
}

// TestAnchorAmbiguousTextUpstream migrates the full valid/invalid suite from
// upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/anchor-ambiguous-text-test.js` 1:1. Position
// assertions cover line/column for every invalid case (the upstream JS suite
// does not assert positions itself, but they're stable on tsgo and locking
// them in protects against future regressions). rslint-specific lock-in
// cases live in anchor_ambiguous_text_extras_test.go.
func TestAnchorAmbiguousTextUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AnchorAmbiguousTextRule, []rule_tester.ValidTestCase{
		// ---- Plain, non-ambiguous text ----
		{Code: `<a>documentation</a>;`, Tsx: true},

		// ---- `${here}` in JSX child position is a literal "${here}" string,
		//      NOT a template substitution. Standardized → "${here}", not in
		//      the ambiguous Set → valid. Locks in the JsxText raw-value path.
		{Code: `<a>${here}</a>;`, Tsx: true},

		// ---- aria-label early-return: a non-ambiguous label hides the
		//      ambiguous "click here" children.
		{Code: `<a aria-label="tutorial on using eslint-plugin-jsx-a11y">click here</a>;`, Tsx: true},

		// ---- aria-label on a child element also wins for that child's
		//      contribution — outer anchor sees the non-ambiguous label text.
		{Code: `<a><span aria-label="tutorial on using eslint-plugin-jsx-a11y">click here</span></a>;`, Tsx: true},

		// ---- alt substitution on img-typed child. `<img>` is its own element
		//      type, so the anchor's accessible text comes from the alt only.
		{Code: `<a><img alt="documentation" /></a>;`, Tsx: true},

		// ---- Option override: explicit words list replaces defaults entirely.
		//      The literal "click here" is no longer ambiguous.
		{Code: `<a>click here</a>`, Tsx: true, Options: disablingDefaultsOpts},

		// ---- componentMap remap: Link → a, with non-ambiguous content.
		{Code: `<Link>documentation</Link>;`, Tsx: true, Settings: linkComponentSettings},

		// ---- componentMap remap of the CHILD (Image → img) — the anchor's
		//      accessible text is the remapped img's alt, not the tag name.
		{Code: `<a><Image alt="documentation" /></a>;`, Tsx: true, Settings: imageComponentSettings},

		// ---- componentMap remap with `${here}` literal child.
		{Code: `<Link>${here}</Link>;`, Tsx: true, Settings: linkComponentSettings},

		// ---- componentMap remap with aria-label early return.
		{Code: `<Link aria-label="tutorial on using eslint-plugin-jsx-a11y">click here</Link>;`, Tsx: true, Settings: linkComponentSettings},

		// ---- componentMap remap + options override stacked.
		{Code: `<Link>click here</Link>`, Tsx: true, Options: disablingDefaultsWithComponentsOpts, Settings: linkComponentSettings},
	}, []rule_tester.InvalidTestCase{
		// ---- Default wordlist: every bare phrase is ambiguous ----
		{Code: `<a>here</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- Case-insensitivity via toLowerCase: HERE → "here" ----
		{Code: `<a>HERE</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		{Code: `<a>click here</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},
		{Code: `<a>learn more</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- Whitespace collapse: multiple spaces → single space ----
		{Code: `<a>learn      more</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- Stripped sentence-ending punctuation: . , ? ! ; : ----
		{Code: `<a>learn more.</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},
		{Code: `<a>learn more?</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},
		{Code: `<a>learn more,</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},
		{Code: `<a>learn more!</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},
		{Code: `<a>learn more;</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},
		{Code: `<a>learn more:</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		{Code: `<a>link</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},
		{Code: `<a>a link</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- aria-label takes precedence over children: an ambiguous label
		//      still triggers even when the children are non-ambiguous.
		{Code: `<a aria-label="click here">something</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- Leading/trailing whitespace trimmed before lookup.
		{Code: `<a> a link </a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- Empty inner element contributes "" which collapses out:
		//      "a" + " " + "" + " " + " link" → "a   link" → "a link".
		{Code: `<a>a<i></i> link</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},
		{Code: `<a><i></i>a link</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- Recursion: nested span contributes its standardized children.
		{Code: `<a><span>click</span> here</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- Recursion + whitespace collapse across recursion boundaries:
		//      <span> normalizes to "click", outer joins with " here" →
		//      "click  here" → "click here".
		{Code: `<a><span> click </span> here</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- aria-hidden suppression on a child (boolean form).
		{Code: `<a><span aria-hidden>more text</span>learn more</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- aria-hidden suppression on a child (string "true" form).
		{Code: `<a><span aria-hidden="true">more text</span>learn more</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- alt on a self-closing img child contributes the alt as text.
		{Code: `<a><img alt="click here"/></a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- alt on a non-img tag is IGNORED (`alt` only counts when the
		//      element resolves to `img`); the rule falls through to
		//      children → "click here" → ambiguous.
		{Code: `<a alt="tutorial on using eslint-plugin-jsx-a11y">click here</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},
		{Code: `<a><span alt="tutorial on using eslint-plugin-jsx-a11y">click here</span></a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- Custom React-component child contributes its text recursively.
		{Code: `<a><CustomElement>click</CustomElement> here</a>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- componentMap remap of the anchor itself.
		{Code: `<Link>here</Link>`, Tsx: true, Settings: linkComponentSettings, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- componentMap remap of the CHILD (Image → img) with ambiguous alt.
		{Code: `<a><Image alt="click here" /></a>`, Tsx: true, Settings: imageComponentSettings, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "anchorAmbiguousText", Message: defaultMessage, Line: 1, Column: 1}}},

		// ---- Options override that ADDS a phrase: defaults are replaced, so
		//      only the listed phrase is ambiguous, AND the wordlist in the
		//      diagnostic reflects the override (Message tracks Options).
		{
			Code:    `<a>a disallowed word</a>`,
			Tsx:     true,
			Options: customWordsOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorAmbiguousText", Message: errorMessage([]string{"a disallowed word"}), Line: 1, Column: 1},
			},
		},
	})
}
