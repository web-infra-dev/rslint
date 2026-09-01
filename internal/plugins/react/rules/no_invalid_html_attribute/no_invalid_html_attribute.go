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

func message(id, description string, data map[string]string) rule.RuleMessage {
	return rule.RuleMessage{Id: id, Description: description, Data: data}
}

func suggestion(id, description string, fix rule.RuleFix) []rule.RuleSuggestion {
	return []rule.RuleSuggestion{{Message: message(id, description, nil), FixesArr: []rule.RuleFix{fix}}}
}

// stringParts is the JS /\\S+/g equivalent with byte offsets. JavaScript's
// whitespace set is deliberately used instead of Go's Unicode definition.
func stringParts(value string) [][2]int {
	var out [][2]int
	for start := 0; start < len(value); {
		for start < len(value) {
			r, n := utf8.DecodeRuneInString(value[start:])
			if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				break
			}
			start += n
		}
		end := start
		for end < len(value) {
			r, n := utf8.DecodeRuneInString(value[end:])
			if ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				break
			}
			end += n
		}
		if start < end {
			out = append(out, [2]int{start, end})
		}
		start = end
	}
	return out
}

func whitespaceParts(value string) [][2]int {
	var out [][2]int
	for start := 0; start < len(value); {
		r, n := utf8.DecodeRuneInString(value[start:])
		if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			start += n
			continue
		}
		end := start + n
		for end < len(value) {
			r, n = utf8.DecodeRuneInString(value[end:])
			if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				break
			}
			end += n
		}
		out = append(out, [2]int{start, end})
		start = end
	}
	return out
}

func literalString(node *ast.Node) (string, bool) {
	if node != nil && node.Kind == ast.KindStringLiteral {
		return node.AsStringLiteral().Text, true
	}
	return "", false
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
	return core.NewTextRange(r.Pos()+1+offset[0], r.Pos()+1+offset[1])
}

// https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-invalid-html-attribute.js
var NoInvalidHtmlAttributeRule = rule.Rule{
	Name: "react/no-invalid-html-attribute", Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !configured(options) {
			return rule.RuleListeners{}
		}
		reportLiteral := func(valueNode, removeNode *ast.Node, element string) {
			value, isString := literalString(valueNode)
			if !isString {
				ctx.ReportNodeWithDeferredSuggestions(valueNode, message("onlyStrings", "“rel” attribute only supports strings.", map[string]string{"attributeName": "rel"}), func() []rule.RuleSuggestion {
					return suggestion("suggestRemoveNonString", "remove non-string value in “rel”", rule.RuleFixRemove(ctx.SourceFile, removeNode))
				})
				return
			}
			if ecmascript.IsBlank(value) {
				ctx.ReportNodeWithDeferredSuggestions(valueNode, message("noEmpty", "An empty “rel” attribute is meaningless.", map[string]string{"attributeName": "rel", "reportingValue": value}), func() []rule.RuleSuggestion {
					return suggestion("suggestRemoveEmpty", "remove empty attribute rel", rule.RuleFixRemove(ctx.SourceFile, removeNode))
				})
				return
			}
			for _, part := range stringParts(value) {
				word := value[part[0]:part[1]]
				tags, exists := relValues[word]
				data := map[string]string{"attributeName": "rel", "reportingValue": word}
				id, desc := "", ""
				if !exists {
					id, desc = "neverValid", "“"+word+"” is never a valid “rel” attribute value."
				} else if _, allowed := tags[element]; !allowed {
					id, desc = "notValidFor", "“"+word+"” is not a valid “rel” attribute value for <"+element+">."
					data["elementName"] = element
				}
				if id != "" {
					partRange := literalRange(ctx, valueNode, part)
					ctx.ReportNodeWithDeferredSuggestions(valueNode, message(id, desc, data), func() []rule.RuleSuggestion {
						return suggestion("suggestRemoveInvalid", "“remove invalid attribute "+word+"”", rule.RuleFixRemoveRange(partRange))
					})
				}
			}
			parts := stringParts(value)
			for i, part := range parts {
				if value[part[0]:part[1]] == "shortcut" {
					second := ""
					last := value[part[0]:part[1]]
					first := last
					if i+1 < len(parts) {
						pairText := value[part[0]:parts[i+1][1]]
						pairParts := strings.Split(pairText, " ")
						first = pairParts[0]
						last = pairParts[len(pairParts)-1]
						if len(pairParts) > 1 {
							second = pairParts[1]
						}
					}
					if first == "shortcut" && last != "icon" {
						id, desc := "notAlone", "“shortcut” must be directly followed by “icon”."
						if second != "" {
							id, desc = "notPaired", "“shortcut” can not be directly followed by “"+second+"” without “icon”."
						}
						ctx.ReportNode(valueNode, message(id, desc, map[string]string{"reportingValue": "shortcut", "secondValue": second, "missingValue": "icon"}))
					}
				}
			}
			for _, part := range whitespaceParts(value) {
				r := literalRange(ctx, valueNode, part)
				whitespace := value[part[0]:part[1]]
				if part[0] == 0 || part[1] == len(value) || whitespace != " " {
					replacement := " "
					if part[0] == 0 || part[1] == len(value) {
						replacement = ""
					}
					ctx.ReportNodeWithDeferredSuggestions(valueNode, message("spaceDelimited", "”rel“ attribute values should be space delimited.", map[string]string{"attributeName": "rel"}), func() []rule.RuleSuggestion {
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
				ctx.ReportNodeWithDeferredSuggestions(name, message("onlyMeaningfulFor", "The ”rel“ attribute only has meaning on the tags: \"<link>\", \"<a>\", \"<area>\", \"<form>\"", map[string]string{"attributeName": "rel", "tagNames": "\"<link>\", \"<a>\", \"<area>\", \"<form>\""}), func() []rule.RuleSuggestion {
					return suggestion("suggestRemoveDefault", "remove rel", rule.RuleFixRemove(ctx.SourceFile, node))
				})
				return
			}
			init := attr.Initializer
			if init == nil {
				ctx.ReportNodeWithDeferredSuggestions(name, message("emptyIsMeaningless", "An empty “rel” attribute is meaningless.", map[string]string{"attributeName": "rel"}), func() []rule.RuleSuggestion {
					return suggestion("suggestRemoveEmpty", "remove empty attribute rel", rule.RuleFixRemove(ctx.SourceFile, node))
				})
				return
			}
			if init.Kind == ast.KindStringLiteral {
				reportLiteral(init, node, element)
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
				reportLiteral(expr, node, element)
				return
			}
			if expr.Kind == ast.KindObjectLiteralExpression || (expr.Kind == ast.KindIdentifier && expr.AsIdentifier().Text == "undefined") {
				ctx.ReportNodeWithDeferredSuggestions(init, message("onlyStrings", "“rel” attribute only supports strings.", map[string]string{"attributeName": "rel"}), func() []rule.RuleSuggestion {
					return suggestion("suggestRemoveDefault", "remove rel", rule.RuleFixRemove(ctx.SourceFile, node))
				})
			}
		}
		checkCall := func(node *ast.Node) {
			call := node.AsCallExpression()
			if !reactutil.IsCreateElementCall(call.Expression, reactutil.DefaultReactPragma) || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
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
					ctx.ReportNode(key, message("onlyMeaningfulFor", "The ”rel“ attribute only has meaning on the tags: \"<link>\", \"<a>\", \"<area>\", \"<form>\"", map[string]string{"attributeName": "rel", "tagNames": "\"<link>\", \"<a>\", \"<area>\", \"<form>\""}))
					continue
				}
				if method {
					ctx.ReportNode(property, message("noMethod", "The ”rel“ attribute cannot be a method.", map[string]string{"attributeName": "rel"}))
					continue
				}
				if shorthand || computed {
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
		ctx.ReportNodeWithDeferredSuggestions(value, message("neverValid", "“"+text+"” is never a valid “rel” attribute value.", map[string]string{"attributeName": "rel", "reportingValue": text}), func() []rule.RuleSuggestion {
			return suggestion("suggestRemoveInvalid", "“remove invalid attribute "+text+"”", removeStaticValue(ctx, value, text))
		})
		return
	}
	if _, allowed := tags[element]; !allowed {
		raw := sourceText(ctx, value)
		ctx.ReportNode(value, message("notValidFor", "“"+raw+"” is not a valid “rel” attribute value for <"+element+">.", map[string]string{"attributeName": "rel", "reportingValue": raw, "elementName": element}))
	}
}
