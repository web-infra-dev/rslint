package jsx_no_comment_textnodes

import (
	"strings"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// hasCommentLikeLine mirrors ESLint's `/^\s*\/(\/|\*)/m` test. It cannot be
// expressed as a Go `regexp`: ECMAScript's `\s` covers Unicode WhiteSpace
// (NBSP, BOM, Zs category, …) and `/m`'s `^` fires after LS/PS line
// separators — Go's `regexp` does neither. We iterate lines manually instead.
func hasCommentLikeLine(text string) bool {
	rest := text
	for {
		lineEnd, termLen := indexJsLineTerminator(rest)
		var line string
		if lineEnd < 0 {
			line = rest
		} else {
			line = rest[:lineEnd]
		}
		trimmed := strings.TrimLeftFunc(line, isJsWhitespace)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			return true
		}
		if lineEnd < 0 {
			return false
		}
		rest = rest[lineEnd+termLen:]
	}
}

// indexJsLineTerminator returns the byte offset and width of the first
// ECMAScript LineTerminator in s (LF, CR, CRLF, LS U+2028, PS U+2029), or
// -1, 0 if none.
func indexJsLineTerminator(s string) (int, int) {
	for i := 0; i < len(s); {
		c := s[i]
		if c == '\n' {
			return i, 1
		}
		if c == '\r' {
			if i+1 < len(s) && s[i+1] == '\n' {
				return i, 2
			}
			return i, 1
		}
		if c < 0x80 {
			i++
			continue
		}
		if c == 0xE2 && i+2 < len(s) && s[i+1] == 0x80 && (s[i+2] == 0xA8 || s[i+2] == 0xA9) {
			return i, 3
		}
		i++
	}
	return -1, 0
}

// isJsWhitespace matches ECMAScript `\s` minus LineTerminator (handled
// separately): tab, vertical tab, form feed, space, NBSP (U+00A0), BOM
// (U+FEFF), and any rune in Unicode category Zs.
func isJsWhitespace(r rune) bool {
	switch r {
	case '\t', '\v', '\f', ' ', '\u00A0', '\uFEFF':
		return true
	}
	return unicode.Is(unicode.Zs, r)
}

var JsxNoCommentTextnodesRule = rule.Rule{
	Name: "react/jsx-no-comment-textnodes",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			// In tsgo, comment-like source (e.g. `<div>// foo</div>`) is parsed
			// as a JsxText child of the surrounding JsxElement / JsxFragment.
			// ESLint additionally registers a `Literal` listener because Babel
			// can emit such text as a `Literal`, but tsgo never does — the two
			// listeners collapse into this one.
			ast.KindJsxText: func(node *ast.Node) {
				parent := node.Parent
				// ESLint's `parent.type !== 'JSXAttribute' && parent.type !==
				// 'JSXExpressionContainer' && parent.type.indexOf('JSX') !== -1`
				// collapses to this positive check — in tsgo, JsxText only ever
				// appears as a direct child of JsxElement / JsxFragment.
				if parent == nil || (!ast.IsJsxElement(parent) && !ast.IsJsxFragment(parent)) {
					return
				}
				// Fast path: the parser marks JsxText nodes that are entirely
				// whitespace/newlines — those can never match the comment regex.
				jsxText := node.AsJsxText()
				if jsxText == nil || jsxText.ContainsOnlyTriviaWhiteSpaces {
					return
				}
				// Use raw source text, not any cooked representation: the
				// upstream rule deliberately calls `getText(context, node)` so
				// that HTML entities such as `&#x2F;` are NOT decoded before
				// the comment test (see the `<pre>&#x2F;&#x2F; ...</pre>`
				// valid case in the ESLint test suite).
				//
				// WARNING: `scanner.GetSourceTextOfNodeFromSourceFile` and
				// `utils.TrimmedNodeText` can NOT be used here — they both
				// advance past leading trivia via the TypeScript scanner, which
				// interprets a JsxText starting with `//` or `/*` as a JS
				// comment and returns a start position past the end of the
				// node (producing an inverted slice). JsxText in JSX grammar
				// is not subject to JS trivia rules, so slicing the source
				// buffer directly is the only correct read.
				source := ctx.SourceFile.Text()
				startPos, endPos := node.Pos(), node.End()
				if startPos < 0 || endPos > len(source) || startPos >= endPos {
					return
				}
				raw := source[startPos:endPos]
				if !hasCommentLikeLine(raw) {
					return
				}
				// Report on the raw JsxText range, NOT via ReportNode: rslint's
				// ReportNode trims leading trivia via the TypeScript scanner,
				// which — faced with a JsxText that literally starts with `//`
				// or `/*` — would treat those characters as comment trivia and
				// skip past them, sending the diagnostic to the wrong position
				// (or into the next token entirely). `//` inside a JsxText is
				// not a comment in the JSX grammar, so we must keep the range
				// verbatim.
				ctx.ReportRange(core.NewTextRange(startPos, endPos), rule.RuleMessage{
					Id:          "putCommentInBraces",
					Description: "Comments inside children section of tag should be placed inside braces",
				})
			},
		}
	},
}
