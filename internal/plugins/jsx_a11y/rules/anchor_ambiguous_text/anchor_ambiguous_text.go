// Package anchor_ambiguous_text ports eslint-plugin-jsx-a11y's
// `anchor-ambiguous-text` rule. The rule reports `<a>` elements whose
// accessible text — after normalization (case-fold, whitespace collapse,
// stripped sentence-ending punctuation) — exactly matches one of a
// configurable list of ambiguous link phrases ("click here", "here", "link",
// "a link", "learn more" by default).
//
// The accessible-text computation mirrors upstream's `getAccessibleChildText`
// recursion: aria-label wins (early return), `<img alt="…">` substitutes the
// alt text for the element's children, aria-hidden elements contribute the
// empty string, and otherwise child JsxText / StringLiteral / nested JsxElement
// pieces are joined with spaces and normalized once at the top.
package anchor_ambiguous_text

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	jsxtx "github.com/microsoft/typescript-go/shim/transformers/jsxtransforms"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed anchor_ambiguous_text.schema.json
var schemaJSON []byte

// defaultAmbiguousWords mirrors upstream's `DEFAULT_AMBIGUOUS_WORDS`. Order
// matters because it surfaces in the user-facing diagnostic via
// `words.join('", "')` — keep it identical to upstream so the message text
// matches byte-for-byte.
var defaultAmbiguousWords = []string{
	"click here",
	"here",
	"link",
	"a link",
	"learn more",
}

// punctuationRE / multiSpaceRE mirror upstream's two `replace` calls in
// `standardizeSpaceAndCase`. The punctuation class is intentionally narrow —
// it covers sentence-ending marks (including Spanish-inverted `¿`/`¡` and the
// interrobang `‽`) but NOT quotation marks, parentheses, or hyphens. Anything
// outside the class is preserved verbatim, then case-folded.
//
// multiSpaceRE matches runs of two-or-more Unicode whitespace characters.
// Upstream's `/\s\s+/g` uses JS regex `\s`, which (per ECMA-262 21.2.2.12)
// covers ASCII whitespace + NBSP (U+00A0) + the U+1680..U+3000 space block
// + U+2028/U+2029 line separators. Go regex's `\s` is ASCII-only and RE2 does
// NOT expose the Unicode `White_Space` property as a named class, so we
// assemble the equivalent: Go `\s` (`\t\n\f\r `) + `\v` (U+000B, omitted
// from Go's `\s` but in JS's) + `\p{Z}` (Unicode Separator category — Zs
// covers NBSP, OGHAM SPACE MARK, the U+2000..U+200A block, NARROW NO-BREAK
// SPACE, MEDIUM MATHEMATICAL SPACE, IDEOGRAPHIC SPACE; Zl is U+2028; Zp is
// U+2029). The only JS `\s` member
// not in the result is U+FEFF ZWNBSP — a BOM that can't realistically
// appear inside JSX text. Without this fix, NBSP runs slip past plain
// `\s\s+` and the rule silently misses inputs like
// `<a>learn[NBSP][NBSP]more</a>`.
var (
	punctuationRE = regexp.MustCompile(`[,.?¿!‽¡;:]`)
	multiSpaceRE  = regexp.MustCompile(`[\s\v\p{Z}]{2,}`)
)

// standardizeSpaceAndCase mirrors upstream's same-named helper:
//
//	input.trim()
//	     .replace(/[,.?¿!‽¡;:]/g, '')
//	     .replace(/\s\s+/g, ' ')
//	     .toLowerCase()
//
// Applied at every recursion level (and again on the joined result of the
// top-level traversal) so that punctuation removal and whitespace collapse
// compose across nested elements, not just at the leaves.
func standardizeSpaceAndCase(input string) string {
	s := strings.TrimSpace(input)
	s = punctuationRE.ReplaceAllString(s, "")
	s = multiSpaceRE.ReplaceAllString(s, " ")
	return strings.ToLower(s)
}

// getAccessibleChildText mirrors upstream's recursive
// `getAccessibleChildText(node, elementType)`. `node` is the JSX root —
// either a JsxElement (paired form) or a JsxSelfClosingElement (self-closing
// form). Returns the normalized accessible text the caller compares against
// `ambiguousWords`.
//
// Resolution order, matching upstream:
//  1. If the element carries an `aria-label` with a literal-typed string
//     value, return its standardized form. This is the "escape hatch" path —
//     children are ignored entirely.
//  2. If the element's resolved type is `img` AND it carries an `alt` with a
//     literal-typed string value, return the standardized alt. Note the
//     asymmetry: `alt` is only consulted when the tag is an img-like element;
//     a non-img anchor with `alt="…"` (technically invalid HTML) falls
//     through to the children path.
//  3. If `isHiddenFromScreenReader` reports the element is hidden
//     (`aria-hidden="true"` / boolean form, or `<input type="hidden">`),
//     return the empty string — the element contributes nothing to the
//     enclosing anchor's accessible text.
//  4. Otherwise, walk children and concatenate per-node contributions joined
//     by a single space, then normalize.
//
// Per-child handling (upstream's children map):
//   - JsxText / JsxTextAllWhiteSpaces → the raw `.Text` (tsgo splits ESTree's
//     `JSXText` into two kinds based on whitespace-only-ness; both expose the
//     same `.Text` field and behave identically here).
//   - StringLiteral → the literal string value (matches upstream's
//     `case 'Literal'` — uncommon as a JSX child but legal e.g. after Babel).
//   - JsxElement / JsxSelfClosingElement → recurse. Self-closing elements
//     enter the same resolution order, so a self-closing `<img alt="x"/>`
//     contributes its alt text in step 2.
//   - Everything else (JsxExpression, JsxFragment, …) → empty string. This
//     mirrors upstream's switch which falls to `default: return ”`.
//
// The leaf-empty-string contributions are intentionally preserved (rather
// than skipped): they introduce the spaces that — combined with the
// top-level `\s\s+` collapse — give `<a>a<i></i> link</a>` the value
// `"a link"` rather than `"a  link"` or `"alink"`.
func getAccessibleChildText(node *ast.Node, elementType func(*ast.Node) string) string {
	if node == nil {
		return ""
	}

	// OpeningElementOf normalizes the paired-vs-self-closing AST split so
	// downstream prop / type lookups see the same "opening element +
	// attributes" surface regardless of form. Returns (nil, nil) for any
	// non-JsxElement / non-JsxSelfClosingElement kind, including nested
	// JsxFragment children that recursion hands us — those collapse to ""
	// here, matching upstream's switch which has no JSXFragment arm.
	opening, openingAttrs := jsxa11yutil.OpeningElementOf(node)
	if opening == nil {
		return ""
	}
	// reactutil.GetJsxChildren handles JsxElement and JsxFragment (returns
	// nil for JsxSelfClosingElement, which has none) — single source of
	// truth for "give me the JSX children list", same helper the other
	// jsx-a11y rules use.
	children := reactutil.GetJsxChildren(node)
	resolvedType := elementType(opening)

	// Step 1: aria-label early return.
	//
	// Upstream uses `getLiteralPropValue(getProp(attrs, 'aria-label'))` and
	// gates on JS truthiness. LiteralPropStringValue mirrors the string-typed
	// extraction path of `getLiteralPropValue` and additionally filters out
	// the empty string here (which JS truthiness also rejects). The boolean
	// attribute form `<a aria-label />` returns ("", false) from
	// LiteralPropStringValue, so it falls through — this is intentional and
	// better than upstream, which would crash trying to `.trim()` a JS
	// boolean. See the package-level comment for the divergence rationale.
	if ariaLabel := jsxa11yutil.FindAttributeByName(openingAttrs, "aria-label"); ariaLabel != nil {
		if v, ok := jsxa11yutil.LiteralPropStringValue(ariaLabel); ok && v != "" {
			return standardizeSpaceAndCase(v)
		}
	}

	// Step 2: alt substitution for img-like elements.
	//
	// `resolvedType` is the settings-resolved type (polymorphic prop +
	// components map applied), so `<Image alt="…">` with
	// `components: { Image: 'img' }` enters this branch — matching upstream's
	// `elementType(node.openingElement) === 'img'`.
	if resolvedType == "img" {
		if altAttr := jsxa11yutil.FindAttributeByName(openingAttrs, "alt"); altAttr != nil {
			if v, ok := jsxa11yutil.LiteralPropStringValue(altAttr); ok && v != "" {
				return standardizeSpaceAndCase(v)
			}
		}
	}

	// Step 3: aria-hidden suppression. Upstream's `isHiddenFromScreenReader`
	// takes the (already-resolved) elementType + attrs, so we mirror with
	// the *FromTagAttrs variant — same boolean truth-table, no double
	// resolution of the tag name.
	if jsxa11yutil.IsHiddenFromScreenReaderFromTagAttrs(resolvedType, openingAttrs) {
		return ""
	}

	// Step 4: walk children.
	parts := make([]string, 0, len(children))
	for _, child := range children {
		switch child.Kind {
		case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces:
			// tsgo's JsxText.Text preserves the RAW source bytes from the
			// scanner (`s.tokenValue = s.text[fullStartPos:pos]`), so HTML
			// entities like `&nbsp;` / `&amp;` / `&#104;` remain in
			// `&…;` form. Babel / jsx-eslint's JSXText.value, by contrast,
			// is entity-decoded at parse time, and jsx-ast-utils'
			// `getAccessibleChildText` consumes that decoded form via
			// `String(currentNode.value)`. Apply `jsxtx.DecodeEntities`
			// here to realign — same helper jsxa11yutil uses for the JSX
			// attribute path (see directAttributeStringValue's rationale
			// comment for the general principle).
			parts = append(parts, jsxtx.DecodeEntities(child.AsJsxText().Text))
		case ast.KindStringLiteral:
			// StringLiteral as a JSX child comes from `<a>{"foo"}</a>`-style
			// inputs (uncommon outside compiler outputs). The value is a JS
			// string literal — already escape-decoded by the parser — so no
			// entity decode is needed; only JsxText carries raw `&…;` form.
			parts = append(parts, child.AsStringLiteral().Text)
		case ast.KindNoSubstitutionTemplateLiteral:
			// Upstream's `case 'Literal'` covers any literal in JSX child
			// position. tsgo splits ESTree's `Literal` across several Kind*
			// values; NoSubstitutionTemplateLiteral is the back-tick variant
			// that has no `${…}` substitutions, so it carries a fully-known
			// string and should contribute it. TemplateExpression (with
			// substitutions) is NOT a Literal upstream → falls to default.
			parts = append(parts, child.AsNoSubstitutionTemplateLiteral().Text)
		case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
			parts = append(parts, getAccessibleChildText(child, elementType))
		default:
			// JsxExpression (`{expr}`), JsxFragment, comments, and any other
			// kind contribute the empty string — matches upstream's switch
			// default. The empty string still participates in the join so
			// surrounding spaces collapse correctly at the top.
			parts = append(parts, "")
		}
	}
	return standardizeSpaceAndCase(strings.Join(parts, " "))
}

// errorMessage mirrors upstream's diagnostic template:
//
//	Ambiguous text within anchor. Screen reader users rely on link text for
//	context; the words "<words.join('", "')>" are ambiguous and do not provide
//	enough context.
//
// The wordlist is interpolated verbatim from the active option (or the
// defaults), so changing `words` propagates into the user-visible message —
// upstream behaves the same way.
func errorMessage(words []string) string {
	return fmt.Sprintf(
		`Ambiguous text within anchor. Screen reader users rely on link text for context; the words "%s" are ambiguous and do not provide enough context.`,
		strings.Join(words, `", "`),
	)
}

type options struct {
	// words is the active ambiguous-phrase list. Defaults to
	// `defaultAmbiguousWords` when the option is absent or unparseable.
	words []string
}

func parseOptions(raw []any) options {
	opts := options{words: defaultAmbiguousWords}
	if len(raw) == 0 {
		return opts
	}
	m, _ := raw[0].(map[string]interface{})
	// Upstream: `const { words = DEFAULT_AMBIGUOUS_WORDS } = options;`. The
	// destructuring default only kicks in when `words` is `undefined`. An
	// explicit `[]` REPLACES the defaults and disables the rule for that run
	// (no phrase can ever match the empty Set). Mirror with a presence check
	// — StringSliceOption returns nil for non-array values (keep defaults)
	// and a (possibly empty) []string for any present array (use as-is).
	if rawWords, ok := m["words"]; ok {
		if w := jsxa11yutil.StringSliceOption(rawWords); w != nil {
			opts.words = w
		}
	}
	return opts
}

var AnchorAmbiguousTextRule = rule.Rule{
	Name:   "jsx-a11y/anchor-ambiguous-text",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		// Pre-build the lookup set so each check is O(1) on the wordlist.
		// Upstream uses `new Set(words)` for the same reason.
		ambiguousWords := make(map[string]struct{}, len(opts.words))
		for _, w := range opts.words {
			ambiguousWords[w] = struct{}{}
		}

		// Empty wordlist — the lookup can never hit. Skip the per-element
		// elementType call entirely. Matches upstream's behavior for the
		// `{ words: [] }` "disable" idiom but avoids the per-listener
		// overhead on large files.
		if len(ambiguousWords) == 0 {
			return rule.RuleListeners{}
		}

		message := errorMessage(opts.words)

		elementType := func(node *ast.Node) string {
			return jsxa11yutil.GetElementType(node, ctx.Settings)
		}

		check := func(node *ast.Node) {
			// Upstream's `typesToValidate = ['a']` is a single literal — the
			// rule does NOT take a `components: string[]` option (unlike
			// anchor-has-content / anchor-is-valid). Custom-component
			// matching happens via settings.components / polymorphicProp,
			// already resolved by GetElementType.
			if elementType(node) != "a" {
				return
			}
			nodeText := getAccessibleChildText(jsxa11yutil.JsxAccessibleChildRoot(node), elementType)
			if _, ok := ambiguousWords[nodeText]; !ok {
				return
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "anchorAmbiguousText",
				Description: message,
			})
		}

		// tsgo splits ESTree's single `JSXOpeningElement` into two kinds —
		// see no-distracting-elements / anchor-has-content for the same
		// pattern. Register both so paired (`<a>x</a>`) and self-closing
		// (`<a />`) forms reach the same check.
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
