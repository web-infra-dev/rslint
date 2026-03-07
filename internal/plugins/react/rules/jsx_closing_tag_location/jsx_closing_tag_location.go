package jsx_closing_tag_location

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// JsxClosingTagLocationRule enforces the closing tag location for multiline JSX.
var JsxClosingTagLocationRule = rule.Rule{
	Name: "react/jsx-closing-tag-location",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Default: "tag-aligned"
		location := "tag-aligned"

		// Parse options
		if options != nil {
			if optArray, ok := options.([]interface{}); ok && len(optArray) > 0 {
				if s, ok := optArray[0].(string); ok {
					location = s
				} else if m, ok := optArray[0].(map[string]interface{}); ok {
					if loc, ok := m["location"].(string); ok {
						location = loc
					}
				}
			} else if s, ok := options.(string); ok {
				location = s
			} else {
				optsMap := utils.GetOptionsMap(options)
				if optsMap != nil {
					if loc, ok := optsMap["location"].(string); ok {
						location = loc
					}
				}
			}
		}

		checkClosing := func(node *ast.Node) {
			var openingNode *ast.Node

			switch node.Kind {
			case ast.KindJsxClosingElement:
				// The parent of a JsxClosingElement is the JsxElement
				parent := node.Parent
				if parent == nil || !ast.IsJsxElement(parent) {
					return
				}
				openingNode = parent.AsJsxElement().OpeningElement
			case ast.KindJsxClosingFragment:
				parent := node.Parent
				if parent == nil || !ast.IsJsxFragment(parent) {
					return
				}
				openingNode = parent.AsJsxFragment().OpeningFragment
			default:
				return
			}

			lineStarts := ctx.SourceFile.ECMALineMap()
			text := ctx.SourceFile.Text()

			openingTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, openingNode)
			closingTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)

			openingLine := scanner.ComputeLineOfPosition(lineStarts, openingTrimmed.Pos())
			closingLine := scanner.ComputeLineOfPosition(lineStarts, closingTrimmed.Pos())

			// If opening and closing are on the same line, skip
			if openingLine == closingLine {
				return
			}

			closingLineStart := int(lineStarts[closingLine])

			// Check if the closing tag is the first non-whitespace on its line
			isFirstOnLine := true
			for i := closingLineStart; i < closingTrimmed.Pos(); i++ {
				ch := text[i]
				if ch != ' ' && ch != '\t' {
					isFirstOnLine = false
					break
				}
			}

			openingCol := openingTrimmed.Pos() - int(lineStarts[openingLine])
			closingCol := closingTrimmed.Pos() - closingLineStart

			switch location {
			case "tag-aligned":
				if openingCol != closingCol {
					if !isFirstOnLine {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "onOwnLine",
							Description: "Closing tag of a multiline JSX expression must be on its own line.",
						})
					} else {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "matchIndent",
							Description: "Expected closing tag to match indentation of opening.",
						})
					}
				}
			case "line-aligned":
				// The expected column is the indentation of the line containing the opening tag
				lineStart := int(lineStarts[openingLine])
				expectedCol := 0
				for i := lineStart; i < len(text); i++ {
					ch := text[i]
					if ch == ' ' || ch == '\t' {
						expectedCol++
					} else {
						break
					}
				}
				if closingCol != expectedCol {
					if !isFirstOnLine {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "onOwnLine",
							Description: "Closing tag of a multiline JSX expression must be on its own line.",
						})
					} else {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "alignWithOpening",
							Description: "Expected closing tag to be aligned with the line containing the opening tag",
						})
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindJsxClosingElement:  checkClosing,
			ast.KindJsxClosingFragment: checkClosing,
		}
	},
}
