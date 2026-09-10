package jsx_props_no_multi_spaces

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// JsxPropsNoMultiSpacesRule disallows multiple spaces between inline JSX props.
var JsxPropsNoMultiSpacesRule = rule.Rule{
	Name:   "react/jsx-props-no-multi-spaces",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		text := ctx.SourceFile.Text()
		lineStarts := ctx.SourceFile.ECMALineMap()

		check := func(node *ast.Node) {
			comments := ctx.Comments.All()
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
			checkGap(ctx, text, lineStarts, comments, tagNameEnd, firstTrimmed.Pos(), firstTrimmed.End(), props[0], tagDisplayName, getDisplayName(ctx, props[0]))

			// Check between consecutive attributes
			for i := 1; i < len(props); i++ {
				prevTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, props[i-1])
				currTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, props[i])
				checkGap(ctx, text, lineStarts, comments, prevTrimmed.End(), currTrimmed.Pos(), currTrimmed.End(), props[i], getDisplayName(ctx, props[i-1]), getDisplayName(ctx, props[i]))
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
		// NodeList.End points at the parsed closing angle token. Looking the
		// token up from that parser-owned boundary avoids mistaking a `>` in
		// trailing trivia (notably a block comment) for the delimiter.
		closeAngle := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, typeArgs.End())
		text := ctx.SourceFile.Text()
		if closeAngle.Pos() < len(text) && text[closeAngle.Pos()] == '>' {
			return closeAngle.End()
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

func getDisplayName(ctx rule.RuleContext, node *ast.Node) string {
	if ast.IsJsxSpreadAttribute(node) {
		return utils.TrimmedNodeText(ctx.SourceFile, node.AsJsxSpreadAttribute().Expression)
	}
	if name := reactutil.GetJsxPropName(node); name != "" {
		return name
	}
	return "element"
}

// hasEmptyLines mirrors upstream's comparison of the previous JSX node, every
// comment before the current attribute, and the current attribute. A comment is
// one item regardless of newlines inside it, while any ECMAScript whitespace on
// an otherwise empty intervening line still contributes to the location gap.
func hasEmptyLines(lineStarts []core.TextPos, comments []*ast.CommentRange, startPos, endPos int) bool {
	previousEnd := startPos
	for _, comment := range utils.CommentsInSpan(comments, startPos, endPos) {
		if scanner.ComputeLineOfPosition(lineStarts, comment.Pos())-scanner.ComputeLineOfPosition(lineStarts, previousEnd) >= 2 {
			return true
		}
		if comment.End() > previousEnd {
			previousEnd = comment.End()
		}
	}
	return scanner.ComputeLineOfPosition(lineStarts, endPos)-scanner.ComputeLineOfPosition(lineStarts, previousEnd) >= 2
}

func checkGap(ctx rule.RuleContext, text string, lineStarts []core.TextPos, comments []*ast.CommentRange, prevEnd, currStart, currEnd int, reportNode *ast.Node, prevName, currName string) {
	prevEndLine := scanner.ComputeLineOfPosition(lineStarts, prevEnd)
	currEndLine := scanner.ComputeLineOfPosition(lineStarts, currEnd)

	if prevEndLine == currEndLine {
		// Same line (both prev end and curr end on same line) - check for multiple spaces
		between := text[prevEnd:currStart]
		if between != " " {
			ctx.ReportNodeWithFixes(reportNode, rule.RuleMessage{
				Id:          "onlyOneSpace",
				Description: fmt.Sprintf("Expected only one space between \"%s\" and \"%s\"", prevName, currName),
			}, rule.RuleFix{
				Text:  " ",
				Range: core.NewTextRange(prevEnd, currStart),
			})
		}
	} else {
		// Different lines - comments bridge gaps, but empty lines around them do not.
		if hasEmptyLines(lineStarts, comments, prevEnd, currStart) {
			ctx.ReportNode(reportNode, rule.RuleMessage{
				Id:          "noLineGap",
				Description: fmt.Sprintf("Expected no line gap between \"%s\" and \"%s\"", prevName, currName),
			})
		}
	}
}
