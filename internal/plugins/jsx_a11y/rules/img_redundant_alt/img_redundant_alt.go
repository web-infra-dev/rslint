package img_redundant_alt

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// errorMessage mirrors upstream's `errorMessage` constant verbatim — including
// the curly-quote in `don’t` and the trailing comma after `photo,` inside
// the quoted list, both reproduced from the upstream string.
const errorMessage = "Redundant alt attribute. Screen-readers already announce `img` tags as an image. You don’t need to use the words `image`, `photo,` or `picture` (or any specified custom words) in the alt prop."

// defaultRedundantWords mirrors upstream's `REDUNDANT_WORDS` constant.
var defaultRedundantWords = []string{"image", "photo", "picture"}

type options struct {
	components []string
	words      []string
}

func parseOptions(raw any) options {
	opts := options{}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if rawComponents, ok := m["components"]; ok {
		if arr, ok := rawComponents.([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					opts.components = append(opts.components, s)
				}
			}
		}
	}
	if rawWords, ok := m["words"]; ok {
		if arr, ok := rawWords.([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					opts.words = append(opts.words, s)
				}
			}
		}
	}
	return opts
}

var ImgRedundantAltRule = rule.Rule{
	Name: "jsx-a11y/img-redundant-alt",
	Run: func(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
		rawOptions := rule.LegacyUnwrapOptions(_rawOptions)
		opts := parseOptions(rawOptions)

		// Upstream: `typesToValidate = ['img'].concat(componentOptions)`.
		typesToValidate := make(map[string]struct{}, 1+len(opts.components))
		typesToValidate["img"] = struct{}{}
		for _, c := range opts.components {
			typesToValidate[c] = struct{}{}
		}

		// Precompute the lowercased redundant-word list once. Upstream lowercases
		// the words on every call; we hoist that here since the list never
		// changes after options parsing.
		redundantWords := make([]string, 0, len(defaultRedundantWords)+len(opts.words))
		redundantWords = append(redundantWords, defaultRedundantWords...)
		redundantWords = append(redundantWords, opts.words...)
		for i := range redundantWords {
			redundantWords[i] = strings.ToLower(redundantWords[i])
		}

		elementType := func(node *ast.Node) string {
			return jsxa11yutil.GetElementType(node, ctx.Settings)
		}

		check := func(node *ast.Node) {
			nodeType := elementType(node)
			if _, ok := typesToValidate[nodeType]; !ok {
				return
			}

			attrs := reactutil.GetJsxElementAttributes(node)
			altProp := jsxa11yutil.FindAttributeByName(attrs, "alt")
			if altProp == nil {
				return
			}

			// Upstream: `value = getLiteralPropValue(altProp)`, then guarded by
			// `typeof value === 'string'`. LiteralPropStringValue collapses the
			// extract-then-typeof check into a single (value, ok) result, and
			// crucially does NOT unwrap TS wrappers (`as` / `!` / `satisfies`)
			// — jsx-ast-utils' LITERAL_TYPES maps those to noop -> null, so
			// e.g. `alt={"photo" as string}` is upstream-valid (null is not a
			// string).
			value, ok := jsxa11yutil.LiteralPropStringValue(altProp)
			if !ok {
				return
			}

			// Upstream: `isVisible = isHiddenFromScreenReader(...) === false`.
			// Invert and early-return on the hidden case.
			if jsxa11yutil.IsHiddenFromScreenReader(node, elementType) {
				return
			}

			if containsRedundantWord(value, redundantWords) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "redundantAlt",
					Description: errorMessage,
				})
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

// containsAnyASCIIPrintable reports whether the string contains at least one
// character in U+0020 through U+007F. Mirrors upstream's
// `safeRegexTest(/[\x20-\x7F]+/)` predicate. A value mixing CJK with even one
// ASCII space (e.g. "画像 photo") qualifies as ASCII because the space
// is in range — that pushes the value onto the word-tokenizer path.
func containsAnyASCIIPrintable(s string) bool {
	for i := range len(s) {
		c := s[i]
		if c >= 0x20 && c <= 0x7F {
			return true
		}
	}
	return false
}

// isJSWhitespace reports whether r is matched by ECMAScript's `\s` — the
// union of `WhiteSpace` and `LineTerminator` per the spec. We do NOT use
// `unicode.IsSpace` because Go's whitespace set differs from JS in two
// places that show up in real-world alt text:
//
//   - U+0085 (NEL) — Go matches, JS does NOT.
//   - U+FEFF (ZWNBSP) — JS matches, Go does NOT.
//
// Without this hand-rolled predicate, `<img alt="photo<U+0085>bar" />` would
// silently report (Go-only false positive) and `<img alt="photo<U+FEFF>bar"
// />` would silently pass (Go-only false negative). Both are surfaced by the
// differential check against eslint-plugin-jsx-a11y v6.10.2.
//
// Reference: ECMAScript 2024, §11.2 White Space + §11.3 Line Terminators.
//   - WhiteSpace: U+0009, U+000B, U+000C, U+0020, U+00A0, U+FEFF, plus all
//     code points in Unicode category `Space_Separator` (Zs).
//   - LineTerminator: U+000A, U+000D, U+2028, U+2029.
func isJSWhitespace(r rune) bool {
	switch r {
	case '\t', '\v', '\f', ' ', '\n', '\r':
		// U+0009, U+000B, U+000C, U+0020, U+000A, U+000D
		return true
	case '\u00A0', '\u1680', '\u2028', '\u2029', '\u202F', '\u205F', '\u3000', '\uFEFF':
		// NBSP, OGHAM SPACE MARK, LINE SEPARATOR, PARAGRAPH SEPARATOR,
		// NARROW NO-BREAK SPACE, MEDIUM MATHEMATICAL SPACE, IDEOGRAPHIC SPACE,
		// ZERO WIDTH NO-BREAK SPACE (BOM).
		return true
	}
	// U+2000 through U+200A — assorted en/em/thin/hair/punctuation spaces,
	// all in category Space_Separator and matched by JS `\s`.
	if r >= '\u2000' && r <= '\u200A' {
		return true
	}
	return false
}

// splitOnJSWhitespace mirrors JS `value.split(/\s+/)` for the purpose of
// token equality matching. Unlike upstream's regex split, this returns ONLY
// the non-empty tokens — JS produces empty strings at leading / trailing
// whitespace edges, but those never equal any redundant word, so dropping
// them is observationally identical and avoids one nested loop per call.
func splitOnJSWhitespace(value string) []string {
	tokens := make([]string, 0, 4)
	start := -1
	for i, r := range value {
		if isJSWhitespace(r) {
			if start >= 0 {
				tokens = append(tokens, value[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		tokens = append(tokens, value[start:])
	}
	return tokens
}

// containsRedundantWord mirrors upstream's `containsRedundantWord`.
//
// `redundantWords` MUST already be lowercased by the caller (we hoist the
// `.toLowerCase()` map out of the hot path).
//
// Branch 1 — ASCII path: split by ECMAScript `\s+` and exact-match each
// token (case-insensitively) against the redundant list. Substrings inside
// a token DON'T match: "Photography" stays valid because it's a single
// token.
//
// Branch 2 — non-ASCII path: substring-match each redundant word against
// the lowercased value. This is what makes `<img alt="イメージです" />`
// with `words: ['イメージ']` invalid even though there's no
// whitespace boundary — upstream specifically uses substring-includes for
// non-ASCII to handle scripts that don't use space-delimited words.
func containsRedundantWord(value string, redundantWords []string) bool {
	if containsAnyASCIIPrintable(value) {
		for _, token := range splitOnJSWhitespace(value) {
			lower := strings.ToLower(token)
			for _, w := range redundantWords {
				if lower == w {
					return true
				}
			}
		}
		return false
	}
	lowerValue := strings.ToLower(value)
	for _, w := range redundantWords {
		if strings.Contains(lowerValue, w) {
			return true
		}
	}
	return false
}
