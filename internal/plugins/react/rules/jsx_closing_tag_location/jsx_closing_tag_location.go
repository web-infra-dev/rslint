package jsx_closing_tag_location

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
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

			// Compute expected indentation
			var expectedIndent string
			switch location {
			case "tag-aligned":
				expectedIndent = strings.Repeat(" ", openingCol)
			case "line-aligned":
				lineStart := int(lineStarts[openingLine])
				indentEnd := lineStart
				for indentEnd < len(text) && (text[indentEnd] == ' ' || text[indentEnd] == '\t') {
					indentEnd++
				}
				expectedIndent = text[lineStart:indentEnd]
			}

			if closingCol == len(expectedIndent) {
				return
			}

			if !isFirstOnLine {
				wsStart := closingTrimmed.Pos()
				for wsStart > closingLineStart && (text[wsStart-1] == ' ' || text[wsStart-1] == '\t') {
					wsStart--
				}
				ctx.ReportNodeWithFixes(node, rule.RuleMessage{
					Id:          "onOwnLine",
					Description: "Closing tag of a multiline JSX expression must be on its own line.",
				}, rule.RuleFix{
					Text:  "\n" + expectedIndent,
					Range: core.NewTextRange(wsStart, closingTrimmed.Pos()),
				})
			} else {
				msgId := "matchIndent"
				msgDesc := "Expected closing tag to match indentation of opening."
				if location == "line-aligned" {
					msgId = "alignWithOpening"
					msgDesc = "Expected closing tag to be aligned with the line containing the opening tag"
				}
				ctx.ReportNodeWithFixes(node, rule.RuleMessage{
					Id:          msgId,
					Description: msgDesc,
				}, rule.RuleFix{
					Text:  expectedIndent,
					Range: core.NewTextRange(closingLineStart, closingTrimmed.Pos()),
				})
			}
		}

		return rule.RuleListeners{
			ast.KindJsxClosingElement:  checkClosing,
			ast.KindJsxClosingFragment: checkClosing,
		}
	},
}
