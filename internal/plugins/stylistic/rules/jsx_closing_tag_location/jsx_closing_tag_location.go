// Package jsx_closing_tag_location ports `@stylistic/jsx-closing-tag-location`
// to rslint. It enforces the closing tag location for multiline JSX elements
// and fragments: under `tag-aligned` (default) the closing tag must sit in the
// same column as the opening tag, and under `line-aligned` it must sit in the
// column where the line containing the opening tag begins.
//
// The rule originated as `react/jsx-closing-tag-location` and was adopted into
// eslint-stylistic. Two upstream-stylistic specifics are mirrored here: the
// autofix indent is always emitted as spaces (never the opening line's literal
// indentation, so tab indents collapse to spaces), and the "move the tag to
// its own line" fix is a pure insertion that does NOT strip whitespace already
// preceding the tag (upstream's `insertTextBefore`).
package jsx_closing_tag_location

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const defaultLocation = "tag-aligned"

var messages = map[string]string{
	"onOwnLine":        "Closing tag of a multiline JSX expression must be on its own line.",
	"matchIndent":      "Expected closing tag to match indentation of opening.",
	"alignWithOpening": "Expected closing tag to be aligned with the line containing the opening tag",
}

// messageLocation maps the option to the messageId emitted when the closing
// tag is alone on its line (upstream's MESSAGE_LOCATION table). When the
// closing tag shares its line with preceding content the rule always emits
// `onOwnLine` instead.
var messageLocation = map[string]string{
	"tag-aligned":  "matchIndent",
	"line-aligned": "alignWithOpening",
}

// parseOption accepts every shape rslint's loader can deliver for upstream's
// single-string schema (`enum: ['tag-aligned', 'line-aligned']`, default
// `tag-aligned`): nil, a bare string (single-option CLI form), or a
// single-element array (rule-tester form). Any value outside the enum falls
// back to the default — upstream's JSON schema rejects those before the rule
// runs, so this branch is only reachable through hand-written rslint configs.
func parseOption(options any) string {
	raw := options
	if arr, ok := options.([]interface{}); ok {
		if len(arr) == 0 {
			return defaultLocation
		}
		raw = arr[0]
	}
	if s, ok := raw.(string); ok && (s == "tag-aligned" || s == "line-aligned") {
		return s
	}
	return defaultLocation
}

// JsxClosingTagLocationRule enforces closing tag location for multiline JSX.
var JsxClosingTagLocationRule = rule.Rule{
	Name: "@stylistic/jsx-closing-tag-location",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		location := parseOption(options)

		checkClosing := func(node *ast.Node) {
			// Resolve the matching opening tag from the parent element /
			// fragment (mirrors upstream's `node.parent.openingElement` /
			// `openingFragment`). A closing tag always has a parent, but guard
			// defensively against parser recovery shapes.
			var openingNode *ast.Node
			switch node.Kind {
			case ast.KindJsxClosingElement:
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
			if openingNode == nil {
				return
			}

			text := ctx.SourceFile.Text()
			lineStarts := ctx.SourceFile.ECMALineMap()

			openingTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, openingNode)
			openingPos := openingTrimmed.Pos()
			openingLine := scanner.ComputeLineOfPosition(lineStarts, openingPos)
			openingLineStart := int(lineStarts[openingLine])

			closingTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
			closingPos := closingTrimmed.Pos()
			closingLine := scanner.ComputeLineOfPosition(lineStarts, closingPos)
			closingLineStart := int(lineStarts[closingLine])

			// Single-line element (opening and closing tags share a line) —
			// nothing to enforce.
			if openingLine == closingLine {
				return
			}

			// Columns are UTF-16 code-unit offsets to match ESLint's
			// `loc.column`; a multi-byte character before either tag would
			// otherwise shift the comparison off ESLint's.
			closingColumn := reactutil.UTF16Length(text[closingLineStart:closingPos])

			// targetColumn is upstream's getIndentation(): the opening tag's
			// own column under `tag-aligned`, or the indentation of the line
			// containing the opening tag under `line-aligned`. It doubles as
			// both the column the closing tag is checked against and the width
			// of the autofix indent.
			var targetColumn int
			if location == "line-aligned" {
				targetColumn = reactutil.UTF16Length(reactutil.HorizontalWhitespacePrefix(text[openingLineStart:]))
			} else {
				targetColumn = reactutil.UTF16Length(text[openingLineStart:openingPos])
			}

			if closingColumn == targetColumn {
				return
			}

			indent := strings.Repeat(" ", targetColumn)

			// First-in-line iff only whitespace precedes the closing tag back
			// to its line start. Equivalent to upstream's
			// `isNodeFirstInLine`, which walks back over whitespace-only JSXText
			// and reports whether the preceding token lands on an earlier line.
			leadingWS := reactutil.HorizontalWhitespacePrefix(text[closingLineStart:closingPos])
			if len(leadingWS) == closingPos-closingLineStart {
				// Closing tag is alone on its line — only its indentation is
				// wrong. Replace the leading whitespace of the closing line
				// with `indent` (upstream's replaceTextRange over
				// [lineStart, node.start]).
				msgID := messageLocation[location]
				ctx.ReportNodeWithFixes(node, rule.RuleMessage{
					Id:          msgID,
					Description: messages[msgID],
				}, rule.RuleFix{
					Text:  indent,
					Range: core.NewTextRange(closingLineStart, closingPos),
				})
				return
			}

			// Closing tag shares its line with preceding content — move it to
			// its own line. Upstream uses insertTextBefore(node, "\n"+indent),
			// a pure insertion that leaves any whitespace already before the
			// tag in place.
			ctx.ReportNodeWithFixes(node, rule.RuleMessage{
				Id:          "onOwnLine",
				Description: messages["onOwnLine"],
			}, rule.RuleFix{
				Text:  "\n" + indent,
				Range: core.NewTextRange(closingPos, closingPos),
			})
		}

		return rule.RuleListeners{
			ast.KindJsxClosingElement:  checkClosing,
			ast.KindJsxClosingFragment: checkClosing,
		}
	},
}
