package jsx_newline

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// The upstream if/then constraint is expressed as anyOf for Draft 4.
//
//go:embed jsx_newline.schema.json
var schemaJSON []byte

var JsxNewlineRule = rule.Rule{
	Name:   "react/jsx-newline",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		prevent, allowMultilines := false, false
		if len(options) > 0 {
			if configuration, ok := options[0].(map[string]any); ok {
				prevent, _ = configuration["prevent"].(bool)
				allowMultilines, _ = configuration["allowMultilines"].(bool)
			}
		}
		text := ctx.SourceFile.Text()
		lineMap := ctx.SourceFile.ECMALineMap()
		isBlockComment := func(node *ast.Node) bool {
			return node.Kind == ast.KindJsxExpression && strings.HasPrefix(utils.TrimmedNodeText(ctx.SourceFile, node), "{/*")
		}
		isMultiline := func(node *ast.Node) bool {
			range_ := utils.TrimNodeTextRange(ctx.SourceFile, node)
			return scanner.ComputeLineOfPosition(lineMap, range_.Pos()) != scanner.ComputeLineOfPosition(lineMap, range_.End())
		}
		check := func(parent *ast.Node) {
			children := reactutil.GetJsxChildren(parent)
			for index, element := range children {
				if !isElementOrExpression(element) || index+2 >= len(children) || isBlockComment(element) {
					continue
				}
				spacing, sibling := children[index+1], children[index+2]
				if spacing.Kind != ast.KindJsxText {
					continue
				}
				raw := text[spacing.Pos():spacing.End()]
				value := raw
				if strings.ContainsRune(value, '&') {
					value = ecmascript.DecodeJSXEntities(value)
				}
				hasBlankLine := containsBlankLine(value)
				multiline := false
				if allowMultilines {
					multiline = isMultiline(element)
					if !multiline {
						// Comments stick to the following element. Fragments and spread
						// children are not JSXElement/JSXExpressionContainer in ESTree.
						for _, next := range children[index+2:] {
							if isElementOrExpression(next) && !isBlockComment(next) {
								multiline = isMultiline(next)
								break
							}
						}
					}
				}
				remove := prevent && !multiline
				if hasBlankLine != remove {
					continue
				}
				message := rule.RuleMessage{Id: "require", Description: "JSX element should start in a new line"}
				if multiline {
					message = rule.RuleMessage{Id: "allowMultilines", Description: "Multiline JSX elements should start in a new line"}
				} else if prevent {
					message = rule.RuleMessage{Id: "prevent", Description: "JSX element should not start in a new line"}
				}
				ctx.ReportNodeWithDeferredFixes(sibling, message, func() []rule.RuleFix {
					return []rule.RuleFix{{Range: spacing.Loc, Text: fixNewlines(raw, remove)}}
				})
			}
		}
		return rule.RuleListeners{
			ast.KindJsxElement:  check,
			ast.KindJsxFragment: check,
		}
	},
}

func isElementOrExpression(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		return true
	case ast.KindJsxExpression:
		return node.AsJsxExpression().DotDotDotToken == nil
	}
	return false
}

// containsBlankLine matches upstream's /\n\s*\n/ using JavaScript whitespace.
func containsBlankLine(text string) bool {
	for newline := strings.IndexByte(text, '\n'); newline >= 0; {
		text = text[newline+1:]
		newline = strings.IndexByte(text, '\n')
		if newline >= 0 && ecmascript.IsBlank(text[:newline]) {
			return true
		}
	}
	return false
}

// fixNewlines reproduces /(\n)(?!.*\1)/g or /(\n\n)(?!.*\1)/g.
// The lookahead stops at any JavaScript line terminator. In particular, do
// not normalize CRLF or remove spaces from indented blank lines.
func fixNewlines(text string, prevent bool) string {
	match, replacement := "\n", "\n\n"
	if prevent {
		match, replacement = replacement, match
	}
	var result strings.Builder
	start := 0
	for position := 0; position < len(text); {
		offset := strings.Index(text[position:], match)
		if offset < 0 {
			break
		}
		position += offset
		end := position + len(match)
		next := strings.IndexAny(text[end:], ecmascript.LineTerminators)
		if next >= 0 && strings.HasPrefix(text[end+next:], match) {
			position++
			continue
		}
		result.WriteString(text[start:position])
		result.WriteString(replacement)
		start, position = end, end
	}
	result.WriteString(text[start:])
	return result.String()
}
