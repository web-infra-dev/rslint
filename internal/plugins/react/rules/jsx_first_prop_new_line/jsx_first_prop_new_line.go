package jsx_first_prop_new_line

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// JsxFirstPropNewLineRule enforces the position of the first property in a JSX element.
var JsxFirstPropNewLineRule = rule.Rule{
	Name: "react/jsx-first-prop-new-line",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Default option
		option := "multiline-multiprop"

		// Parse options - can be a string or in an options array
		if options != nil {
			if optArray, ok := options.([]interface{}); ok && len(optArray) > 0 {
				if s, ok := optArray[0].(string); ok {
					option = s
				}
			} else if s, ok := options.(string); ok {
				option = s
			} else {
				optsMap := utils.GetOptionsMap(options)
				if optsMap != nil {
					if s, ok := optsMap["option"].(string); ok {
						option = s
					}
				}
			}
		}

		check := func(node *ast.Node) {
			var props []*ast.Node
			var openingNode *ast.Node

			switch node.Kind {
			case ast.KindJsxOpeningElement:
				opening := node.AsJsxOpeningElement()
				openingNode = node
				attrs := opening.Attributes.AsJsxAttributes()
				if attrs.Properties != nil {
					props = attrs.Properties.Nodes
				}
			case ast.KindJsxSelfClosingElement:
				self := node.AsJsxSelfClosingElement()
				openingNode = node
				attrs := self.Attributes.AsJsxAttributes()
				if attrs.Properties != nil {
					props = attrs.Properties.Nodes
				}
			}

			if len(props) == 0 {
				return
			}

			lineMap := ctx.SourceFile.ECMALineMap()
			text := ctx.SourceFile.Text()
			firstProp := props[0]

			// Use the trimmed position for the opening element
			openingTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, openingNode)
			openingLine := scanner.ComputeLineOfPosition(lineMap, openingTrimmed.Pos())
			firstPropTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, firstProp)
			firstPropLine := scanner.ComputeLineOfPosition(lineMap, firstPropTrimmed.Pos())

			// Determine if element is multiline (opening tag start to opening tag end)
			openingEndLine := scanner.ComputeLineOfPosition(lineMap, openingTrimmed.End())
			isMultiline := openingLine != openingEndLine

			propOnNewLineMsg := rule.RuleMessage{
				Id:          "propOnNewLine",
				Description: "Property should be placed on a new line",
			}
			propOnSameLineMsg := rule.RuleMessage{
				Id:          "propOnSameLine",
				Description: "Property should be placed on the same line as the component declaration",
			}

			// Fix: replace whitespace before prop with newline
			newLineFix := func() rule.RuleFix {
				propStart := firstPropTrimmed.Pos()
				wsStart := propStart
				for wsStart > 0 && (text[wsStart-1] == ' ' || text[wsStart-1] == '\t') {
					wsStart--
				}
				return rule.RuleFix{Text: "\n", Range: core.NewTextRange(wsStart, propStart)}
			}
			// Fix: replace whitespace/newlines before prop with a space
			sameLineFix := func() rule.RuleFix {
				propStart := firstPropTrimmed.Pos()
				wsStart := propStart
				for wsStart > 0 && (text[wsStart-1] == ' ' || text[wsStart-1] == '\t' || text[wsStart-1] == '\n' || text[wsStart-1] == '\r') {
					wsStart--
				}
				return rule.RuleFix{Text: " ", Range: core.NewTextRange(wsStart, propStart)}
			}

			switch option {
			case "always":
				if openingLine == firstPropLine {
					ctx.ReportNodeWithFixes(firstProp, propOnNewLineMsg, newLineFix())
				}
			case "never":
				if openingLine != firstPropLine {
					ctx.ReportNodeWithFixes(firstProp, propOnSameLineMsg, sameLineFix())
				}
			case "multiline":
				if isMultiline && openingLine == firstPropLine {
					ctx.ReportNodeWithFixes(firstProp, propOnNewLineMsg, newLineFix())
				}
			case "multiline-multiprop":
				if isMultiline && len(props) > 1 && openingLine == firstPropLine {
					ctx.ReportNodeWithFixes(firstProp, propOnNewLineMsg, newLineFix())
				}
			case "multiprop":
				if len(props) > 1 && openingLine == firstPropLine {
					ctx.ReportNodeWithFixes(firstProp, propOnNewLineMsg, newLineFix())
				} else if len(props) <= 1 && isMultiline && openingLine != firstPropLine {
					ctx.ReportNodeWithFixes(firstProp, propOnSameLineMsg, sameLineFix())
				}
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
