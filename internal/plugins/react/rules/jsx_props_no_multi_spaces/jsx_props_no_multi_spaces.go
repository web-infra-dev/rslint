package jsx_props_no_multi_spaces

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// JsxPropsNoMultiSpacesRule disallows multiple spaces between inline JSX props.
var JsxPropsNoMultiSpacesRule = rule.Rule{
	Name: "react/jsx-props-no-multi-spaces",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		text := ctx.SourceFile.Text()
		lineStarts := ctx.SourceFile.ECMALineMap()

		check := func(node *ast.Node) {
			var props []*ast.Node
			var tagNameEnd int
			var tagName *ast.Node

			switch node.Kind {
			case ast.KindJsxOpeningElement:
				opening := node.AsJsxOpeningElement()
				tagName = opening.TagName
				tagNameEnd = getTagNameEnd(ctx, opening.TagName, opening.TypeArguments)
				attrs := opening.Attributes.AsJsxAttributes()
				if attrs.Properties != nil {
					props = attrs.Properties.Nodes
				}
			case ast.KindJsxSelfClosingElement:
				self := node.AsJsxSelfClosingElement()
				tagName = self.TagName
				tagNameEnd = getTagNameEnd(ctx, self.TagName, self.TypeArguments)
				attrs := self.Attributes.AsJsxAttributes()
				if attrs.Properties != nil {
					props = attrs.Properties.Nodes
				}
			}

			if len(props) == 0 {
				return
			}

			// Check between tag name and first attribute
			firstTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, props[0])
			tagDisplayName := getTagDisplayName(ctx, tagName)
			checkGap(ctx, text, lineStarts, tagNameEnd, firstTrimmed.Pos(), firstTrimmed.End(), props[0], tagDisplayName, getDisplayName(props[0]))

			// Check between consecutive attributes
			for i := 1; i < len(props); i++ {
				prevTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, props[i-1])
				currTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, props[i])
				checkGap(ctx, text, lineStarts, prevTrimmed.End(), currTrimmed.Pos(), currTrimmed.End(), props[i], getDisplayName(props[i-1]), getDisplayName(props[i]))
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

// getTagNameEnd returns the effective end position of the tag name area,
// including TypeArguments if present (e.g., `<App<T>` -> end of `>`).
func getTagNameEnd(ctx rule.RuleContext, tagName *ast.Node, typeArgs *ast.NodeList) int {
	if typeArgs != nil && len(typeArgs.Nodes) > 0 {
		// Use the end of the type arguments list (includes the closing `>`)
		trimmed := utils.TrimNodeTextRange(ctx.SourceFile, typeArgs.Nodes[len(typeArgs.Nodes)-1])
		// The closing `>` of type arguments is after the last type arg node.
		text := ctx.SourceFile.Text()
		pos := trimmed.End()
		for pos < len(text) && text[pos] != '>' {
			pos++
		}
		if pos < len(text) {
			return pos + 1 // past the `>`
		}
	}
	trimmed := utils.TrimNodeTextRange(ctx.SourceFile, tagName)
	return trimmed.End()
}

func getTagDisplayName(ctx rule.RuleContext, tagName *ast.Node) string {
	if tagName == nil {
		return "element"
	}
	trimmed := utils.TrimNodeTextRange(ctx.SourceFile, tagName)
	return ctx.SourceFile.Text()[trimmed.Pos():trimmed.End()]
}

func getDisplayName(node *ast.Node) string {
	if name := reactutil.GetJsxPropName(node); name != "" {
		return name
	}
	return "element"
}

// hasBlankLineBetween checks if there is a truly blank line (only whitespace, no comments)
// between two positions in the source text.
func hasBlankLineBetween(text string, lineStarts []core.TextPos, startPos, endPos int) bool {
	startLine := scanner.ComputeLineOfPosition(lineStarts, startPos)
	endLine := scanner.ComputeLineOfPosition(lineStarts, endPos)

	for line := startLine + 1; line < endLine; line++ {
		lineStart := int(lineStarts[line])
		lineEnd := len(text)
		if line+1 < len(lineStarts) {
			lineEnd = int(lineStarts[line+1])
		}
		isBlank := true
		for i := lineStart; i < lineEnd; i++ {
			ch := text[i]
			if ch != ' ' && ch != '\t' && ch != '\r' && ch != '\n' {
				isBlank = false
				break
			}
		}
		if isBlank {
			return true
		}
	}
	return false
}

func checkGap(ctx rule.RuleContext, text string, lineStarts []core.TextPos, prevEnd, currStart, currEnd int, reportNode *ast.Node, prevName, currName string) {
	prevEndLine := scanner.ComputeLineOfPosition(lineStarts, prevEnd)
	currEndLine := scanner.ComputeLineOfPosition(lineStarts, currEnd)

	if prevEndLine == currEndLine {
		// Same line (both prev end and curr end on same line) - check for multiple spaces
		between := text[prevEnd:currStart]
		if between != " " {
			ctx.ReportNode(reportNode, rule.RuleMessage{
				Id:          "onlyOneSpace",
				Description: fmt.Sprintf("Expected only one space between \"%s\" and \"%s\"", prevName, currName),
			})
		}
	} else {
		// Different lines - check for truly blank lines between them (comments bridge gaps)
		if hasBlankLineBetween(text, lineStarts, prevEnd, currStart) {
			ctx.ReportNode(reportNode, rule.RuleMessage{
				Id:          "noLineGap",
				Description: fmt.Sprintf("Expected no line gap between \"%s\" and \"%s\"", prevName, currName),
			})
		}
	}
}
