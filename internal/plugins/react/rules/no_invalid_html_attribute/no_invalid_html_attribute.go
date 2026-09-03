package no_invalid_html_attribute

import (
	_ "embed"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed no_invalid_html_attribute.schema.json
var schemaJSON []byte

var relValues = map[string]map[string]struct{}{
	"alternate": {"link": {}, "area": {}, "a": {}}, "apple-touch-icon": {"link": {}},
	"apple-touch-startup-image": {"link": {}}, "author": {"link": {}, "area": {}, "a": {}},
	"bookmark": {"area": {}, "a": {}}, "canonical": {"link": {}}, "dns-prefetch": {"link": {}},
	"external": {"area": {}, "a": {}, "form": {}}, "help": {"link": {}, "area": {}, "a": {}, "form": {}},
	"icon": {"link": {}}, "license": {"link": {}, "area": {}, "a": {}, "form": {}}, "manifest": {"link": {}},
	"mask-icon": {"link": {}}, "modulepreload": {"link": {}}, "next": {"link": {}, "area": {}, "a": {}, "form": {}},
	"nofollow": {"area": {}, "a": {}, "form": {}}, "noopener": {"area": {}, "a": {}, "form": {}},
	"noreferrer": {"area": {}, "a": {}, "form": {}}, "opener": {"area": {}, "a": {}, "form": {}},
	"pingback": {"link": {}}, "preconnect": {"link": {}}, "prefetch": {"link": {}}, "preload": {"link": {}},
	"prerender": {"link": {}}, "prev": {"link": {}, "area": {}, "a": {}, "form": {}},
	"search": {"link": {}, "area": {}, "a": {}, "form": {}}, "shortcut": {"link": {}},
	"shortcut icon": {"link": {}}, "stylesheet": {"link": {}}, "tag": {"area": {}, "a": {}},
}

var htmlElements = map[string]struct{}{
	"a": {}, "abbr": {}, "acronym": {}, "address": {}, "applet": {}, "area": {}, "article": {}, "aside": {}, "audio": {}, "b": {}, "base": {}, "basefont": {}, "bdi": {}, "bdo": {}, "bgsound": {}, "big": {}, "blink": {}, "blockquote": {}, "body": {}, "br": {}, "button": {}, "canvas": {}, "caption": {}, "center": {}, "cite": {}, "code": {}, "col": {}, "colgroup": {}, "content": {}, "data": {}, "datalist": {}, "dd": {}, "del": {}, "details": {}, "dfn": {}, "dialog": {}, "dir": {}, "div": {}, "dl": {}, "dt": {}, "em": {}, "embed": {}, "fieldset": {}, "figcaption": {}, "figure": {}, "font": {}, "footer": {}, "form": {}, "frame": {}, "frameset": {}, "h1": {}, "h2": {}, "h3": {}, "h4": {}, "h5": {}, "h6": {}, "head": {}, "header": {}, "hgroup": {}, "hr": {}, "html": {}, "i": {}, "iframe": {}, "image": {}, "img": {}, "input": {}, "ins": {}, "kbd": {}, "keygen": {}, "label": {}, "legend": {}, "li": {}, "link": {}, "main": {}, "map": {}, "mark": {}, "marquee": {}, "math": {}, "menu": {}, "menuitem": {}, "meta": {}, "meter": {}, "nav": {}, "nobr": {}, "noembed": {}, "noframes": {}, "noscript": {}, "object": {}, "ol": {}, "optgroup": {}, "option": {}, "output": {}, "p": {}, "param": {}, "picture": {}, "plaintext": {}, "portal": {}, "pre": {}, "progress": {}, "q": {}, "rb": {}, "rp": {}, "rt": {}, "rtc": {}, "ruby": {}, "s": {}, "samp": {}, "script": {}, "section": {}, "select": {}, "shadow": {}, "slot": {}, "small": {}, "source": {}, "spacer": {}, "span": {}, "strike": {}, "strong": {}, "style": {}, "sub": {}, "summary": {}, "sup": {}, "svg": {}, "table": {}, "tbody": {}, "td": {}, "template": {}, "textarea": {}, "tfoot": {}, "th": {}, "thead": {}, "time": {}, "title": {}, "tr": {}, "track": {}, "tt": {}, "u": {}, "ul": {}, "var": {}, "video": {}, "wbr": {}, "xmp": {},
}

var relElements = map[string]struct{}{"link": {}, "a": {}, "area": {}, "form": {}}

func configured(options []any) bool {
	if len(options) == 0 {
		return true
	}
	attrs, ok := options[0].([]interface{})
	if !ok {
		return true
	}
	for _, attr := range attrs {
		if attr == "rel" {
			return true
		}
	}
	return false
}

func message(id, description string) rule.RuleMessage {
	return rule.RuleMessage{Id: id, Description: description}
}

func suggestion(id, description string, fix rule.RuleFix) []rule.RuleSuggestion {
	return []rule.RuleSuggestion{{Message: message(id, description), FixesArr: []rule.RuleFix{fix}}}
}

// stringParts is the JS /\\S+/g equivalent. It records UTF-8 byte spans for
// slicing the decoded Go string and UTF-16 spans for ESLint-style ranges.
type valuePart struct {
	bytes [2]int
	units [2]int
}

func stringParts(value string) []valuePart {
	var out []valuePart
	units := 0
	for start := 0; start < len(value); {
		for start < len(value) {
			r, n := utf8.DecodeRuneInString(value[start:])
			if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				break
			}
			start += n
			units += utf16Width(r)
		}
		end := start
		startUnits := units
		for end < len(value) {
			r, n := utf8.DecodeRuneInString(value[end:])
			if ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				break
			}
			end += n
			units += utf16Width(r)
		}
		if start < end {
			out = append(out, valuePart{bytes: [2]int{start, end}, units: [2]int{startUnits, units}})
		}
		start = end
	}
	return out
}

func whitespaceParts(value string) []valuePart {
	var out []valuePart
	units := 0
	for start := 0; start < len(value); {
		r, n := utf8.DecodeRuneInString(value[start:])
		if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			start += n
			units += utf16Width(r)
			continue
		}
		startUnits := units
		end := start + n
		units += utf16Width(r)
		for end < len(value) {
			r, n = utf8.DecodeRuneInString(value[end:])
			if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				break
			}
			end += n
			units += utf16Width(r)
		}
		out = append(out, valuePart{bytes: [2]int{start, end}, units: [2]int{startUnits, units}})
		start = end
	}
	return out
}

func utf16Width(r rune) int {
	if r > 0xFFFF {
		return 2
	}
	return 1
}

func literalString(node *ast.Node) (string, bool) {
	if node != nil && node.Kind == ast.KindStringLiteral {
		return node.AsStringLiteral().Text, true
	}
	return "", false
}

func jsxLiteralString(ctx rule.RuleContext, node *ast.Node) (string, bool) {
	value, ok := literalString(node)
	if !ok {
		return "", false
	}
	raw := sourceText(ctx, node)
	if len(raw) >= 2 && ((raw[0] == '"' && raw[len(raw)-1] == '"') || (raw[0] == '\'' && raw[len(raw)-1] == '\'')) {
		return ecmascript.DecodeJSXEntities(raw[1 : len(raw)-1]), true
	}
	return value, true
}

func isCreateElementCall(callee *ast.Node) bool {
	if reactutil.IsCreateElementCall(callee, reactutil.DefaultReactPragma) {
		return true
	}
	if callee == nil {
		return false
	}
	wasParenthesized := callee.Kind == ast.KindParenthesizedExpression
	callee = ast.SkipParentheses(callee)
	if callee == nil || (wasParenthesized && ast.IsOptionalChain(callee)) || callee.Kind != ast.KindElementAccessExpression {
		return false
	}
	access := callee.AsElementAccessExpression()
	object := ast.SkipParentheses(access.Expression)
	property := ast.SkipParentheses(access.ArgumentExpression)
	return object != nil && object.Kind == ast.KindIdentifier && object.AsIdentifier().Text == reactutil.DefaultReactPragma &&
		property != nil && property.Kind == ast.KindIdentifier && property.AsIdentifier().Text == "createElement"
}

func reportShortcutPairs(ctx rule.RuleContext, node *ast.Node, value string) {
	for start := range len(value) {
		if !isASCIIWord(value[start]) || (start > 0 && isASCIIWord(value[start-1])) {
			continue
		}
		firstEnd := start
		for firstEnd < len(value) {
			r, size := utf8.DecodeRuneInString(value[firstEnd:])
			if ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				break
			}
			firstEnd += size
		}
		pairEnd := firstEnd
		secondStart := firstEnd
		for secondStart < len(value) {
			r, size := utf8.DecodeRuneInString(value[secondStart:])
			if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				break
			}
			secondStart += size
		}
		if secondStart < len(value) {
			pairEnd = secondStart
			for pairEnd < len(value) {
				r, size := utf8.DecodeRuneInString(value[pairEnd:])
				if ecmascript.IsWhiteSpaceOrLineTerminator(r) {
					break
				}
				pairEnd += size
			}
		}
		pairText := value[start:pairEnd]
		pairParts := strings.Split(pairText, " ")
		first := pairParts[0]
		if first != "shortcut" {
			continue
		}
		last := pairParts[len(pairParts)-1]
		second := ""
		if len(pairParts) > 1 {
			second = pairParts[1]
		}
		if last == "icon" {
			continue
		}
		id, desc := "notAlone", "“shortcut” must be directly followed by “icon”."
		if second != "" {
			id, desc = "notPaired", "“shortcut” can not be directly followed by “"+second+"” without “icon”."
		}
		ctx.ReportNode(node, message(id, desc))
	}
}

func isASCIIWord(value byte) bool {
	return value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z' || value >= '0' && value <= '9' || value == '_'
}

func isESTreeLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral:
		return true
	default:
		return false
	}
}

func sourceText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil || ctx.SourceFile == nil {
		return ""
	}
	r := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return ctx.SourceFile.Text()[r.Pos():r.End()]
}

func removeStaticValue(ctx rule.RuleContext, node *ast.Node, value string) rule.RuleFix {
	raw := sourceText(ctx, node)
	raw = strings.Replace(raw, value, "", 1)
	return rule.RuleFixReplace(ctx.SourceFile, node, raw)
}

func literalRange(ctx rule.RuleContext, node *ast.Node, offset [2]int) core.TextRange {
	r := utils.TrimNodeTextRange(ctx.SourceFile, node)
	raw := ctx.SourceFile.Text()[r.Pos():r.End()]
	if len(raw) >= 2 {
		inner := raw[1 : len(raw)-1]
		return core.NewTextRange(
			r.Pos()+1+rawByteOffsetForUTF16(inner, offset[0]),
			r.Pos()+1+rawByteOffsetForUTF16(inner, offset[1]),
		)
	}
	return core.NewTextRange(r.Pos()+offset[0], r.Pos()+offset[1])
}

func rawByteOffsetForUTF16(value string, target int) int {
	if target <= 0 {
		return 0
	}
	units := 0
	for pos := 0; pos < len(value); {
		r, size := utf8.DecodeRuneInString(value[pos:])
		if units >= target {
			return pos
		}
		units += utf16Width(r)
		pos += size
	}
	return len(value)
}

// https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-invalid-html-attribute.js
var NoInvalidHtmlAttributeRule = rule.Rule{
	Name: "react/no-invalid-html-attribute", Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !configured(options) {
			return rule.RuleListeners{}
		}
		reportLiteral := func(valueNode, nonStringRemoveNode, emptyRemoveNode *ast.Node, element string, decodeJSX bool) {
			value, isString := literalString(valueNode)
			if decodeJSX {
				value, isString = jsxLiteralString(ctx, valueNode)
			}
			if !isString {
				ctx.ReportNodeWithDeferredSuggestions(valueNode, message("onlyStrings", "“rel” attribute only supports strings."), func() []rule.RuleSuggestion {
					return suggestion("suggestRemoveNonString", "remove non-string value in “rel”", rule.RuleFixRemove(ctx.SourceFile, nonStringRemoveNode))
				})
				return
			}
			if ecmascript.IsBlank(value) {
				ctx.ReportNodeWithDeferredSuggestions(valueNode, message("noEmpty", "An empty “rel” attribute is meaningless."), func() []rule.RuleSuggestion {
					return suggestion("suggestRemoveEmpty", "\"remove empty attribute rel\"", rule.RuleFixRemove(ctx.SourceFile, emptyRemoveNode))
				})
				return
			}
			for _, part := range stringParts(value) {
				word := value[part.bytes[0]:part.bytes[1]]
				tags, exists := relValues[word]
				id, desc := "", ""
				if !exists {
					id, desc = "neverValid", "“"+word+"” is never a valid “rel” attribute value."
				} else if _, allowed := tags[element]; !allowed {
					id, desc = "notValidFor", "“"+word+"” is not a valid “rel” attribute value for <"+element+">."
				}
				if id != "" {
					partRange := literalRange(ctx, valueNode, part.units)
					ctx.ReportNodeWithDeferredSuggestions(valueNode, message(id, desc), func() []rule.RuleSuggestion {
						return suggestion("suggestRemoveInvalid", "“remove invalid attribute "+word+"”", rule.RuleFixRemoveRange(partRange))
					})
				}
			}
			reportShortcutPairs(ctx, valueNode, value)
			for _, part := range whitespaceParts(value) {
				r := literalRange(ctx, valueNode, part.units)
				whitespace := value[part.bytes[0]:part.bytes[1]]
				nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, valueNode)
				if r.Pos() == nodeRange.Pos()+1 || r.End() == nodeRange.End()-1 || whitespace != " " {
					replacement := " "
					if part.units[0] == 0 || part.units[1] == int(core.UTF16Len(value)) {
						replacement = ""
					}
					ctx.ReportNodeWithDeferredSuggestions(valueNode, message("spaceDelimited", "”rel“ attribute values should be space delimited."), func() []rule.RuleSuggestion {
						return suggestion("suggestRemoveWhitespaces", "remove whitespaces in “rel”", rule.RuleFixReplaceRange(r, replacement))
					})
				}
			}
		}
		checkJsx := func(node *ast.Node) {
			attr := node.AsJsxAttribute()
			name := attr.Name()
			if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "rel" {
				return
			}
			parent := reactutil.GetJsxParentElement(node)
			tag := reactutil.GetJsxTagName(parent)
			if tag == nil || tag.Kind != ast.KindIdentifier {
				return
			}
			element := tag.AsIdentifier().Text
			if _, ok := htmlElements[element]; !ok {
				return
			}
			if _, ok := relElements[element]; !ok {
				ctx.ReportNodeWithDeferredSuggestions(name, message("onlyMeaningfulFor", "The ”rel“ attribute only has meaning on the tags: \"<link>\", \"<a>\", \"<area>\", \"<form>\""), func() []rule.RuleSuggestion {
					return suggestion("suggestRemoveDefault", "\"remove rel\"", rule.RuleFixRemove(ctx.SourceFile, node))
				})
				return
			}
			init := attr.Initializer
			if init == nil {
				ctx.ReportNodeWithDeferredSuggestions(name, message("emptyIsMeaningless", "An empty “rel” attribute is meaningless."), func() []rule.RuleSuggestion {
					return suggestion("suggestRemoveEmpty", "\"remove empty attribute rel\"", rule.RuleFixRemove(ctx.SourceFile, node))
				})
				return
			}
			if init.Kind == ast.KindStringLiteral {
				reportLiteral(init, node, node, element, true)
				return
			}
			if init.Kind != ast.KindJsxExpression {
				return
			}
			expr := ast.SkipParentheses(init.AsJsxExpression().Expression)
			if expr == nil {
				return
			}
			if expr.Kind == ast.KindStringLiteral || expr.Kind == ast.KindNumericLiteral || expr.Kind == ast.KindBigIntLiteral || expr.Kind == ast.KindTrueKeyword || expr.Kind == ast.KindFalseKeyword || expr.Kind == ast.KindNullKeyword || expr.Kind == ast.KindRegularExpressionLiteral {
				reportLiteral(expr, node, init, element, false)
				return
			}
			if expr.Kind == ast.KindObjectLiteralExpression || (expr.Kind == ast.KindIdentifier && expr.AsIdentifier().Text == "undefined") {
				ctx.ReportNodeWithDeferredSuggestions(init, message("onlyStrings", "“rel” attribute only supports strings."), func() []rule.RuleSuggestion {
					return suggestion("suggestRemoveDefault", "\"remove rel\"", rule.RuleFixRemove(ctx.SourceFile, node))
				})
			}
		}
		checkCall := func(node *ast.Node) {
			call := node.AsCallExpression()
			if !isCreateElementCall(call.Expression) || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
				return
			}
			tagNode := ast.SkipParentheses(call.Arguments.Nodes[0])
			if !isESTreeLiteral(tagNode) {
				return
			}
			element, isString := literalString(tagNode)
			if isString {
				if _, ok := htmlElements[element]; !ok {
					return
				}
			}
			if len(call.Arguments.Nodes) < 2 {
				return
			}
			props := ast.SkipParentheses(call.Arguments.Nodes[1])
			if props.Kind != ast.KindObjectLiteralExpression {
				return
			}
			for _, property := range props.AsObjectLiteralExpression().Properties.Nodes {
				var key, value *ast.Node
				method := false
				shorthand := false
				computed := false
				accessor := false
				switch property.Kind {
				case ast.KindMethodDeclaration:
					method = true
					key = property.AsMethodDeclaration().Name()
				case ast.KindPropertyAssignment:
					p := property.AsPropertyAssignment()
					key, value = p.Name(), p.Initializer
				case ast.KindShorthandPropertyAssignment:
					shorthand = true
					key = property.AsShorthandPropertyAssignment().Name()
				case ast.KindGetAccessor, ast.KindSetAccessor:
					accessor = true
					key = property.Name()
				default:
					continue
				}
				if key != nil && key.Kind == ast.KindComputedPropertyName {
					computed = true
					key = ast.SkipParentheses(key.AsComputedPropertyName().Expression)
				}
				if key == nil || key.Kind != ast.KindIdentifier || key.AsIdentifier().Text != "rel" {
					continue
				}
				if !isString || (element != "link" && element != "a" && element != "area" && element != "form") {
					ctx.ReportNode(key, message("onlyMeaningfulFor", "The ”rel“ attribute only has meaning on the tags: \"<link>\", \"<a>\", \"<area>\", \"<form>\""))
					continue
				}
				if method {
					ctx.ReportNode(property, message("noMethod", "The ”rel“ attribute cannot be a method."))
					continue
				}
				if shorthand || computed || accessor {
					continue
				}
				value = ast.SkipParentheses(value)
				if value != nil && value.Kind == ast.KindArrayLiteralExpression {
					for _, item := range value.AsArrayLiteralExpression().Elements.Nodes {
						if item != nil {
							reportCreateElementValue(ctx, item, element)
						}
					}
					continue
				}
				reportCreateElementValue(ctx, value, element)
			}
		}
		return rule.RuleListeners{ast.KindJsxAttribute: checkJsx, ast.KindCallExpression: checkCall}
	},
}

func reportCreateElementValue(ctx rule.RuleContext, value *ast.Node, element string) {
	if value == nil {
		return
	}
	value = ast.SkipParentheses(value)
	// The upstream createElement branch assumes every ESTree Literal is a
	// string. Do not reproduce that bug: non-string values must not enter the
	// string lookup or suggestion path.
	if value.Kind != ast.KindStringLiteral {
		return
	}
	text, ok := utils.GetStaticExpressionValue(value)
	if !ok {
		return
	}
	tags, exists := relValues[text]
	if !exists {
		ctx.ReportNodeWithDeferredSuggestions(value, message("neverValid", "“"+text+"” is never a valid “rel” attribute value."), func() []rule.RuleSuggestion {
			return suggestion("suggestRemoveInvalid", "“remove invalid attribute "+text+"”", removeStaticValue(ctx, value, text))
		})
		return
	}
	if _, allowed := tags[element]; !allowed {
		raw := sourceText(ctx, value)
		ctx.ReportNode(value, message("notValidFor", "“"+raw+"” is not a valid “rel” attribute value for <"+element+">."))
	}
}
