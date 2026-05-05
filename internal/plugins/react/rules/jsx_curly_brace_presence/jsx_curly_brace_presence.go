package jsx_curly_brace_presence

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	optAlways = "always"
	optNever  = "never"
	optIgnore = "ignore"
)

type curlyBraceOptions struct {
	props             string
	children          string
	propElementValues string
}

func defaultOptions() curlyBraceOptions {
	return curlyBraceOptions{
		props:             optNever,
		children:          optNever,
		propElementValues: optIgnore,
	}
}

func parseOptions(raw any) curlyBraceOptions {
	opts := defaultOptions()
	if raw == nil {
		return opts
	}
	// Upstream accepts EITHER an object OR a bare string at options[0].
	// `utils.GetOptionsMap` only handles the object case, so we route the
	// string shorthand here first.
	if s := utils.GetOptionsString(raw); s != "" {
		if s == optAlways || s == optNever || s == optIgnore {
			opts.props = s
			opts.children = s
		}
		return opts
	}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if s, ok := m["props"].(string); ok && (s == optAlways || s == optNever || s == optIgnore) {
		opts.props = s
	}
	if s, ok := m["children"].(string); ok && (s == optAlways || s == optNever || s == optIgnore) {
		opts.children = s
	}
	if s, ok := m["propElementValues"].(string); ok && (s == optAlways || s == optNever || s == optIgnore) {
		opts.propElementValues = s
	}
	return opts
}

var (
	htmlEntityRegex     = regexp.MustCompile(`&[A-Za-z\d#]+;`)
	leadingTrailingWS   = regexp.MustCompile(`^\s|\s$`)
	disallowedJSXChars  = regexp.MustCompile(`[{<>}]`)
	quoteCharsRegex     = regexp.MustCompile(`['"]`)
	multilineCommentSeq = regexp.MustCompile(`/\*`)
)

// containsLineTerminators mirrors upstream's `/[\n\r  ]/`.
func containsLineTerminators(s string) bool {
	for _, r := range s {
		switch r {
		case '\n', '\r', ' ', ' ':
			return true
		}
	}
	return false
}

func containsHTMLEntity(s string) bool { return htmlEntityRegex.MatchString(s) }

func containsOnlyHTMLEntities(s string) bool {
	return strings.TrimSpace(htmlEntityRegex.ReplaceAllString(s, "")) == ""
}

func containsDisallowedJSXChars(s string) bool { return disallowedJSXChars.MatchString(s) }
func containsQuoteChars(s string) bool         { return quoteCharsRegex.MatchString(s) }

func containsMultilineCommentMarker(s string) bool {
	return multilineCommentSeq.MatchString(s)
}

func isLineBreak(s string) bool {
	return containsLineTerminators(s) && strings.TrimSpace(s) == ""
}

func isAllWhitespace(s string) bool {
	return strings.TrimSpace(s) == ""
}

func isStringWithTrailingWhitespaces(s string) bool {
	return leadingTrailingWS.MatchString(s)
}

// jsStringify mirrors `JSON.stringify(s)` for the limited subset of strings
// the rule emits. Unlike Go's encoding/json, `<`, `>`, `&` are kept as-is
// so JSX-fixed output preserves HTML entities verbatim.
func jsStringify(s string) string {
	var sb strings.Builder
	sb.Grow(len(s) + 2)
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			sb.WriteString(`\\`)
		case '"':
			sb.WriteString(`\"`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		case '\b':
			sb.WriteString(`\b`)
		case '\f':
			sb.WriteString(`\f`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&sb, `\u%04x`, r)
			} else {
				sb.WriteRune(r)
			}
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

// wrapNonHTMLEntities mirrors upstream's helper: split text at HTML-entity
// boundaries, wrap non-entity slices in `{"…"}`, splice entities back in.
func wrapNonHTMLEntities(text string) string {
	parts := htmlEntityRegex.Split(text, -1)
	entities := htmlEntityRegex.FindAllString(text, -1)
	var sb strings.Builder
	for i, p := range parts {
		if p != "" {
			sb.WriteByte('{')
			sb.WriteString(jsStringify(p))
			sb.WriteByte('}')
		}
		if i < len(entities) {
			sb.WriteString(entities[i])
		}
	}
	return sb.String()
}

// wrapJsxTextWithCurlyBraces mirrors upstream's `wrapWithCurlyBraces`. For
// single-line text → `{"…"}`. For multi-line text, processes line by line
// preserving leading whitespace and pure-entity lines.
func wrapJsxTextWithCurlyBraces(rawText string) string {
	if !containsLineTerminators(rawText) {
		return "{" + jsStringify(rawText) + "}"
	}
	lines := strings.Split(rawText, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		firstCharIdx := indexFirstNonSpace(line)
		left := line[:firstCharIdx]
		text := line[firstCharIdx:]
		if containsHTMLEntity(line) {
			lines[i] = left + wrapNonHTMLEntities(text)
		} else {
			lines[i] = left + "{" + jsStringify(text) + "}"
		}
	}
	return strings.Join(lines, "\n")
}

// indexFirstNonSpace mirrors upstream's `line.search(/[^\s]/)` — JS regex
// `\s` covers Unicode whitespace (Zs category, NBSP, BOM, line separators).
func indexFirstNonSpace(s string) int {
	for i, r := range s {
		if !isJsRegexWhitespace(r) {
			return i
		}
	}
	return len(s)
}

func isJsRegexWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\v', '\f', '\u00A0', '\uFEFF':
		return true
	}
	return unicode.Is(unicode.Zs, r)
}

func escapeBackslashes(s string) string {
	return strings.ReplaceAll(s, `\`, `\\`)
}

func escapeDoubleQuotes(s string) string {
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// isJSXLike re-exports reactutil.IsJsxLike for readability inside this file.
// Upstream `jsxUtil.isJSX` covers JSXElement and JSXFragment; in tsgo,
// JsxElement and JsxSelfClosingElement both correspond to JSXElement.
var isJSXLike = reactutil.IsJsxLike

// jsxExpressionHasComments mirrors upstream's
// `sourceCode.getCommentsInside(JSXExpressionNode).length > 0`. The rule
// declines to remove curly braces around a comment-bearing expression
// since the comment cannot survive the rewrite.
//
// We must catch comments at ANY position inside the `{…}` (leading,
// between-tokens, trailing) — `getCommentsInside` is a full-range scan in
// upstream. The TS scanner's `GetLeadingCommentRanges` only returns
// comments adjacent to its `pos` argument, so we walk the full interior
// byte-by-byte. String/template literal regions are skipped to avoid
// matching `//` or `/*` sequences inside string contents.
func jsxExpressionHasComments(text string, je *ast.Node) bool {
	start := je.Pos()
	end := je.End()
	for start < end && text[start] != '{' {
		start++
	}
	if start >= end {
		return false
	}
	start++ // skip `{`
	// Skip the matching `}` at the end (if present).
	scanEnd := end
	if scanEnd > 0 && scanEnd <= len(text) && text[scanEnd-1] == '}' {
		scanEnd--
	}

	i := start
	for i < scanEnd {
		c := text[i]
		switch c {
		case '"', '\'':
			// Skip string literal — terminates at matching unescaped quote
			// or end of line.
			quote := c
			i++
			for i < scanEnd && text[i] != quote {
				if text[i] == '\\' && i+1 < scanEnd {
					i += 2
					continue
				}
				if text[i] == '\n' {
					break
				}
				i++
			}
			if i < scanEnd {
				i++ // consume closing quote
			}
		case '`':
			// Skip template literal — including interpolations. Nested
			// `${...}` may itself contain comments, but those are inside an
			// expression sub-tree, not inside the JsxExpression's textual
			// surface — upstream's `getCommentsInside` would still flag
			// them, so we DO descend through `${…}` braces by simply
			// continuing the outer loop without skipping over them.
			i++
			for i < scanEnd && text[i] != '`' {
				if text[i] == '\\' && i+1 < scanEnd {
					i += 2
					continue
				}
				if text[i] == '$' && i+1 < scanEnd && text[i+1] == '{' {
					// Drop into expression mode (nested braces).
					depth := 1
					i += 2
					for i < scanEnd && depth > 0 {
						switch text[i] {
						case '{':
							depth++
						case '}':
							depth--
						case '/':
							if i+1 < scanEnd && (text[i+1] == '/' || text[i+1] == '*') {
								return true
							}
						}
						i++
					}
					continue
				}
				i++
			}
			if i < scanEnd {
				i++ // consume closing backtick
			}
		case '/':
			if i+1 < scanEnd && (text[i+1] == '/' || text[i+1] == '*') {
				return true
			}
			i++
		default:
			i++
		}
	}
	return false
}

func stringLiteralRawText(text string, node *ast.Node) string {
	return text[node.Pos():node.End()]
}

func jsxTextRawText(text string, node *ast.Node) string {
	return text[node.Pos():node.End()]
}

func trimQuotes(s string) string {
	if len(s) < 2 {
		return s
	}
	return s[1 : len(s)-1]
}

func containsWhitespaceExpression(child *ast.Node) bool {
	if child == nil || child.Kind != ast.KindJsxExpression {
		return false
	}
	je := child.AsJsxExpression()
	if je == nil || je.Expression == nil {
		return false
	}
	expr := je.Expression
	if expr.Kind == ast.KindStringLiteral {
		return isAllWhitespace(expr.AsStringLiteral().Text)
	}
	if expr.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return isAllWhitespace(expr.Text())
	}
	return false
}

// adjacentSiblings mirrors upstream's `getAdjacentSiblings`. The inner loop
// deliberately runs `i = 1; i < len-1` and special-cases first/last index
// — keep verbatim or the boundary-position pairs swap.
func adjacentSiblings(node *ast.Node, children []*ast.Node) []*ast.Node {
	for i := 1; i < len(children)-1; i++ {
		if node == children[i] {
			return []*ast.Node{children[i-1], children[i+1]}
		}
	}
	if len(children) >= 2 && node == children[0] {
		return []*ast.Node{children[1]}
	}
	if len(children) >= 2 && node == children[len(children)-1] {
		return []*ast.Node{children[len(children)-2]}
	}
	return nil
}

func filterOutWhitespaceExpressions(children []*ast.Node) []*ast.Node {
	out := make([]*ast.Node, 0, len(children))
	for _, c := range children {
		if !containsWhitespaceExpression(c) {
			out = append(out, c)
		}
	}
	return out
}

func hasAdjacentJsxExpressionContainers(node *ast.Node, children []*ast.Node) bool {
	if children == nil {
		return false
	}
	filtered := filterOutWhitespaceExpressions(children)
	for _, sib := range adjacentSiblings(node, filtered) {
		if sib != nil && sib.Kind == ast.KindJsxExpression {
			return true
		}
	}
	return false
}

func hasAdjacentJsx(node *ast.Node, children []*ast.Node) bool {
	if children == nil {
		return false
	}
	filtered := filterOutWhitespaceExpressions(children)
	for _, sib := range adjacentSiblings(node, filtered) {
		if sib != nil && (sib.Kind == ast.KindJsxExpression || isJSXLike(sib)) {
			return true
		}
	}
	return false
}

func needToEscapeForJSX(raw string, parentIsAttribute bool) bool {
	if strings.Contains(raw, `\`) {
		return true
	}
	if containsHTMLEntity(raw) {
		return true
	}
	if !parentIsAttribute && containsDisallowedJSXChars(raw) {
		return true
	}
	return false
}

var JsxCurlyBracePresenceRule = rule.Rule{
	Name: "react/jsx-curly-brace-presence",
	Run: func(ctx rule.RuleContext, raw any) rule.RuleListeners {
		opts := parseOptions(raw)
		text := ctx.SourceFile.Text()

		unnecessaryMsg := rule.RuleMessage{
			Id:          "unnecessaryCurly",
			Description: "Curly braces are unnecessary here.",
		}
		missingMsg := rule.RuleMessage{
			Id:          "missingCurly",
			Description: "Need to wrap this literal in a JSX expression.",
		}

		// reportUnnecessaryCurlyOnExpr — replaces the whole `{…}` JsxExpression
		// with the unwrapped content per upstream's fix logic.
		reportUnnecessaryCurlyOnExpr := func(jsxExpr *ast.Node) {
			je := jsxExpr.AsJsxExpression()
			if je == nil {
				return
			}
			// Skip ParenthesizedExpression so `{('foo')}` and `{(<Foo />)}`
			// classify (and emit replacement text) based on their inner
			// content, matching upstream's ESTree-flattened view.
			expr := ast.SkipParentheses(je.Expression)
			parent := jsxExpr.Parent
			parentIsAttribute := parent != nil && parent.Kind == ast.KindJsxAttribute

			var replacement string
			switch {
			case isJSXLike(expr):
				replacement = utils.TrimmedNodeText(ctx.SourceFile, expr)
			case parentIsAttribute:
				switch expr.Kind {
				case ast.KindNoSubstitutionTemplateLiteral:
					ntl := expr.AsNoSubstitutionTemplateLiteral()
					rawText := ntl.RawText
					if rawText == "" {
						// tsgo may not populate RawText for substitution-free
						// templates; fall back to source-text slicing.
						src := utils.TrimmedNodeText(ctx.SourceFile, expr)
						if len(src) >= 2 && src[0] == '`' && src[len(src)-1] == '`' {
							rawText = src[1 : len(src)-1]
						} else {
							rawText = ntl.Text
						}
					}
					replacement = `"` + rawText + `"`
				case ast.KindStringLiteral:
					rawWithQuotes := stringLiteralRawText(text, expr)
					inner := trimQuotes(rawWithQuotes)
					if strings.Contains(inner, `"`) {
						replacement = rawWithQuotes
					} else {
						replacement = `"` + inner + `"`
					}
				default:
					replacement = utils.TrimmedNodeText(ctx.SourceFile, expr)
				}
			default:
				switch expr.Kind {
				case ast.KindNoSubstitutionTemplateLiteral:
					// Use cooked text (matches upstream's
					// `quasis[0].value.cooked`).
					replacement = expr.AsNoSubstitutionTemplateLiteral().Text
				case ast.KindStringLiteral:
					replacement = expr.AsStringLiteral().Text
				default:
					replacement = utils.TrimmedNodeText(ctx.SourceFile, expr)
				}
			}

			ctx.ReportNodeWithFixes(jsxExpr, unnecessaryMsg,
				rule.RuleFixReplace(ctx.SourceFile, jsxExpr, replacement))
		}

		// reportMissingCurlyOnLiteral — wraps the literal in `{"…"}` for
		// attribute initializers and uses line-aware wrapping for JsxText.
		reportMissingCurlyOnLiteral := func(literal *ast.Node) {
			if isJSXLike(literal) {
				inner := utils.TrimmedNodeText(ctx.SourceFile, literal)
				ctx.ReportNodeWithFixes(literal, missingMsg,
					rule.RuleFixReplace(ctx.SourceFile, literal, "{"+inner+"}"))
				return
			}

			parent := literal.Parent
			parentIsAttribute := parent != nil && parent.Kind == ast.KindJsxAttribute

			var rawWithDelimiters string
			if literal.Kind == ast.KindStringLiteral {
				rawWithDelimiters = stringLiteralRawText(text, literal)
			} else {
				rawWithDelimiters = jsxTextRawText(text, literal)
			}

			// Bail-outs that mirror upstream's `fix(fixer)` returning null.
			// Still emit the report (no fix attached).
			if parentIsAttribute && containsLineTerminators(rawWithDelimiters) {
				ctx.ReportNode(literal, missingMsg)
				return
			}
			if isLineBreak(rawWithDelimiters) {
				return
			}
			if containsOnlyHTMLEntities(rawWithDelimiters) {
				ctx.ReportNode(literal, missingMsg)
				return
			}

			var replacement string
			if parentIsAttribute {
				inner := trimQuotes(rawWithDelimiters)
				escaped := escapeDoubleQuotes(escapeBackslashes(inner))
				replacement = `{"` + escaped + `"}`
				ctx.ReportNodeWithFixes(literal, missingMsg,
					rule.RuleFixReplace(ctx.SourceFile, literal, replacement))
				return
			}
			replacement = wrapJsxTextWithCurlyBraces(rawWithDelimiters)
			// JsxText leading whitespace/newlines are trivia for the TS
			// scanner, so the default `TrimNodeTextRange` skips them — both
			// the report range and the fix range need to span the raw
			// `[Pos, End)` to match upstream (which uses the JSXText node's
			// own range, not a trivia-skipped one).
			rawRange := core.NewTextRange(literal.Pos(), literal.End())
			ctx.ReportRangeWithFixes(rawRange, missingMsg,
				rule.RuleFixReplaceRange(rawRange, replacement))
		}

		areRuleConditionsSatisfied := func(parent *ast.Node, condition string) bool {
			if parent == nil {
				return false
			}
			if parent.Kind == ast.KindJsxAttribute && opts.props == condition {
				return true
			}
			if (parent.Kind == ast.KindJsxElement || parent.Kind == ast.KindJsxFragment) && opts.children == condition {
				return true
			}
			return false
		}

		shouldCheckForUnnecessaryCurly := func(jsxExpr *ast.Node) bool {
			je := jsxExpr.AsJsxExpression()
			if je == nil || je.Expression == nil {
				return false
			}
			parent := jsxExpr.Parent

			if parent != nil && parent.Kind == ast.KindJsxAttribute {
				// Strip parens / TS type wrappers when classifying — tsgo
				// preserves these as nodes where ESTree flattens them.
				inner := ast.SkipParentheses(je.Expression)
				exprKind := inner.Kind
				if exprKind != ast.KindStringLiteral &&
					exprKind != ast.KindNoSubstitutionTemplateLiteral &&
					exprKind != ast.KindTemplateExpression {
					return false
				}
			}

			if parent != nil && (parent.Kind == ast.KindJsxElement || parent.Kind == ast.KindJsxFragment) {
				children := reactutil.GetJsxChildren(parent)
				if hasAdjacentJsxExpressionContainers(jsxExpr, children) {
					return false
				}
				if containsWhitespaceExpression(jsxExpr) && hasAdjacentJsx(jsxExpr, children) {
					return false
				}
				if len(children) == 1 && containsWhitespaceExpression(jsxExpr) {
					return false
				}
			}

			return areRuleConditionsSatisfied(parent, optNever)
		}

		shouldCheckForMissingCurly := func(literal *ast.Node) bool {
			if isJSXLike(literal) {
				return opts.propElementValues != optIgnore
			}
			var raw string
			if literal.Kind == ast.KindStringLiteral {
				raw = stringLiteralRawText(text, literal)
			} else {
				raw = jsxTextRawText(text, literal)
			}
			if isLineBreak(raw) || containsOnlyHTMLEntities(raw) {
				return false
			}
			parent := literal.Parent
			children := reactutil.GetJsxChildren(parent)
			if len(children) == 1 && containsWhitespaceExpression(children[0]) {
				return false
			}
			return areRuleConditionsSatisfied(parent, optAlways)
		}

		lintUnnecessaryCurly := func(jsxExpr *ast.Node) {
			je := jsxExpr.AsJsxExpression()
			if je == nil {
				return
			}
			rawExpr := je.Expression
			if rawExpr == nil {
				return
			}
			// Match upstream's logical view: skip ParenthesizedExpression
			// wrappers introduced by tsgo. Children may still be JSX so we
			// must NOT collapse `(<Foo />)` to its inner element via
			// `SkipParentheses` in the FIX path; we only use `expr` here for
			// CLASSIFICATION. The fix paths re-derive their replacement
			// text from the source.
			expr := ast.SkipParentheses(rawExpr)
			if jsxExpressionHasComments(text, jsxExpr) {
				return
			}
			parent := jsxExpr.Parent
			parentIsAttribute := parent != nil && parent.Kind == ast.KindJsxAttribute

			if expr.Kind == ast.KindStringLiteral {
				value := expr.AsStringLiteral().Text
				rawWithQuotes := stringLiteralRawText(text, expr)
				if parentIsAttribute && isAllWhitespace(value) {
					return
				}
				if !parentIsAttribute && isStringWithTrailingWhitespaces(value) {
					return
				}
				if containsMultilineCommentMarker(value) {
					return
				}
				if needToEscapeForJSX(rawWithQuotes, parentIsAttribute) {
					return
				}
				reportUnnecessaryCurlyOnExpr(jsxExpr)
				return
			}

			if expr.Kind == ast.KindNoSubstitutionTemplateLiteral {
				ntl := expr.AsNoSubstitutionTemplateLiteral()
				rawText := ntl.RawText
				if rawText == "" {
					src := utils.TrimmedNodeText(ctx.SourceFile, expr)
					if len(src) >= 2 && src[0] == '`' && src[len(src)-1] == '`' {
						rawText = src[1 : len(src)-1]
					} else {
						rawText = ntl.Text
					}
				}
				cooked := ntl.Text
				if strings.Contains(rawText, "\n") {
					return
				}
				if isStringWithTrailingWhitespaces(rawText) {
					return
				}
				if needToEscapeForJSX(rawText, parentIsAttribute) {
					return
				}
				if containsQuoteChars(cooked) {
					return
				}
				reportUnnecessaryCurlyOnExpr(jsxExpr)
				return
			}

			if expr.Kind == ast.KindTemplateExpression {
				return
			}

			if isJSXLike(expr) {
				reportUnnecessaryCurlyOnExpr(jsxExpr)
				return
			}
		}

		return rule.RuleListeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				attr := node.AsJsxAttribute()
				if attr == nil || attr.Initializer == nil {
					return
				}
				init := attr.Initializer
				if isJSXLike(init) {
					if shouldCheckForMissingCurly(init) {
						reportMissingCurlyOnLiteral(init)
					}
					return
				}
				if init.Kind == ast.KindStringLiteral {
					if shouldCheckForMissingCurly(init) {
						reportMissingCurlyOnLiteral(init)
					}
					return
				}
			},

			ast.KindJsxExpression: func(node *ast.Node) {
				je := node.AsJsxExpression()
				if je == nil {
					return
				}
				if je.DotDotDotToken != nil {
					return
				}

				parent := node.Parent
				parentIsAttribute := parent != nil && parent.Kind == ast.KindJsxAttribute

				// `JSXAttribute > JSXExpressionContainer > JSXElement`:
				// propElementValues='never' reports unnecessary curly on
				// the container. Skip parens around the element so a
				// `prop={(<div />)}` shape still matches.
				if parentIsAttribute && je.Expression != nil &&
					isJSXLike(ast.SkipParentheses(je.Expression)) {
					if opts.propElementValues == optNever {
						reportUnnecessaryCurlyOnExpr(node)
					}
					return
				}

				if shouldCheckForUnnecessaryCurly(node) {
					lintUnnecessaryCurly(node)
				}
			},

			ast.KindJsxText: func(node *ast.Node) {
				if shouldCheckForMissingCurly(node) {
					reportMissingCurlyOnLiteral(node)
				}
			},
		}
	},
}
