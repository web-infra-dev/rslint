// Package jsx_closing_bracket_location ports `@stylistic/jsx-closing-bracket-location`
// to rslint. Behavior mirrors the upstream rule (which originated as
// `react/jsx-closing-bracket-location` and added trailing-comment handling
// when it moved into eslint-stylistic). The trailing-comment branch — when
// expectedLocation is `after-props` or `after-tag` but a comment sits between
// the last attribute (or the tag name, in the zero-prop case) and the closing
// bracket, the location is upgraded to `line-aligned` — is the principal
// semantic delta vs. `react/jsx-closing-bracket-location`.
package jsx_closing_bracket_location

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const defaultLocation = "tag-aligned"

var locationMessages = map[string]string{
	"after-props":   "placed after the last prop",
	"after-tag":     "placed after the opening tag",
	"props-aligned": "aligned with the last prop",
	"tag-aligned":   "aligned with the opening tag",
	"line-aligned":  "aligned with the line containing the opening tag",
}

type bracketOptions struct {
	nonEmpty       string
	selfClosing    string
	nonEmptyOff    bool
	selfClosingOff bool
}

// parseOptions accepts every shape upstream's JSON schema permits:
//
//	string                              → both nonEmpty & selfClosing
//	{ location: string }                → both nonEmpty & selfClosing
//	{ nonEmpty?: string|false,
//	  selfClosing?: string|false }      → per-form configuration
//
// rslint's CLI / config loader collapses a single-element option array into
// the option value directly, so the rule may receive any of `string`,
// `map[string]interface{}`, or `[]interface{}` at the top level. We accept all
// three to stay byte-compatible with both the JS rule-tester and the
// production CLI shape.
func parseOptions(options any) bracketOptions {
	opts := bracketOptions{nonEmpty: defaultLocation, selfClosing: defaultLocation}
	if options == nil {
		return opts
	}

	raw := options
	if arr, ok := options.([]interface{}); ok {
		if len(arr) == 0 {
			return opts
		}
		raw = arr[0]
	}

	switch v := raw.(type) {
	case string:
		opts.nonEmpty = v
		opts.selfClosing = v
	case map[string]interface{}:
		applyMap(&opts, v)
	}
	return opts
}

func applyMap(opts *bracketOptions, m map[string]interface{}) {
	// Upstream's `'location' in config!` makes the two object shapes
	// mutually exclusive: if `location` is present, nonEmpty / selfClosing
	// are IGNORED (upstream destructures `{ nonEmpty: config.location,
	// selfClosing: config.location }` and never reads the other keys). The
	// JSON schema rejects mixing them upstream-side; rslint has no schema
	// validation, so we enforce the same precedence here in code to keep
	// runtime behavior identical on any input the user might pass.
	if v, exists := m["location"]; exists {
		if s, ok := v.(string); ok {
			opts.nonEmpty = s
			opts.selfClosing = s
		}
		return
	}
	if v, exists := m["nonEmpty"]; exists {
		if s, ok := v.(string); ok {
			opts.nonEmpty = s
			opts.nonEmptyOff = false
		} else if b, ok := v.(bool); ok && !b {
			opts.nonEmptyOff = true
		}
	}
	if v, exists := m["selfClosing"]; exists {
		if s, ok := v.(string); ok {
			opts.selfClosing = s
			opts.selfClosingOff = false
		} else if b, ok := v.(bool); ok && !b {
			opts.selfClosingOff = true
		}
	}
}

type tokenInfo struct {
	openingPos    int
	openingLine   int
	openingColumn int

	closingPos    int
	closingLine   int
	closingColumn int

	tagLine int

	hasLastProp       bool
	lastPropEnd       int
	lastPropFirstLine int
	lastPropLastLine  int
	lastPropColumn    int

	openingStartCol int

	openTab  bool
	closeTab bool

	selfClosing bool

	// Trailing-comment span. `hasTrailingComment` is upstream's
	//   lastComment.range[0] > (lastAttributeNode ?? node.name).range[1]
	// — i.e. the last comment **inside** the opening element starts AFTER the
	// last attribute (or the tag name when there are no attributes). When set,
	// the fix's range start is moved to lastCommentEnd so the comment isn't
	// swept into the rewrite.
	hasTrailingComment bool
	lastCommentEnd     int
}

func leadingWhitespace(text string, lineStart int) string {
	if lineStart >= len(text) {
		return ""
	}
	return reactutil.HorizontalWhitespacePrefix(text[lineStart:])
}

// JsxClosingBracketLocationRule enforces closing bracket location for JSX
// multiline elements.
var JsxClosingBracketLocationRule = rule.Rule{
	Name: "@stylistic/jsx-closing-bracket-location",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		check := func(node *ast.Node) {
			var tagName *ast.Node
			var attrs []*ast.Node
			var selfClosing bool

			switch node.Kind {
			case ast.KindJsxOpeningElement:
				opening := node.AsJsxOpeningElement()
				tagName = opening.TagName
				if opening.Attributes != nil {
					jsxAttrs := opening.Attributes.AsJsxAttributes()
					if jsxAttrs.Properties != nil {
						attrs = jsxAttrs.Properties.Nodes
					}
				}
				selfClosing = false
			case ast.KindJsxSelfClosingElement:
				self := node.AsJsxSelfClosingElement()
				tagName = self.TagName
				if self.Attributes != nil {
					jsxAttrs := self.Attributes.AsJsxAttributes()
					if jsxAttrs.Properties != nil {
						attrs = jsxAttrs.Properties.Nodes
					}
				}
				selfClosing = true
			default:
				return
			}

			// Parser error recovery: an incomplete `<` may produce a JSX
			// opening element with a missing tag name. Skip gracefully —
			// upstream's `getFirstToken(node.name)` would also throw.
			if tagName == nil {
				return
			}

			text := ctx.SourceFile.Text()
			lineStarts := ctx.SourceFile.ECMALineMap()

			elemTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
			openingPos := elemTrimmed.Pos()
			elemEnd := elemTrimmed.End()

			// elemEnd should land right after the closing `>`. The reverse
			// scan is defensive against parser recovery shapes that emit
			// trailing trivia. If `>` is genuinely absent (truncated source),
			// bail.
			gtPos := elemEnd - 1
			for gtPos > openingPos && gtPos < len(text) && text[gtPos] != '>' {
				gtPos--
			}
			if gtPos < 0 || gtPos >= len(text) || text[gtPos] != '>' {
				return
			}

			closingPos := gtPos
			if selfClosing {
				slash := gtPos - 1
				for slash > openingPos && (text[slash] == ' ' || text[slash] == '\t' || text[slash] == '\n' || text[slash] == '\r') {
					slash--
				}
				if slash <= openingPos || text[slash] != '/' {
					return
				}
				closingPos = slash
			}

			openingLine := scanner.ComputeLineOfPosition(lineStarts, openingPos)
			openingLineStart := int(lineStarts[openingLine])
			openingColumn := reactutil.UTF16Length(text[openingLineStart:openingPos])

			closingLine := scanner.ComputeLineOfPosition(lineStarts, closingPos)
			closingLineStart := int(lineStarts[closingLine])
			closingColumn := reactutil.UTF16Length(text[closingLineStart:closingPos])

			tagTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, tagName)
			tagLine := scanner.ComputeLineOfPosition(lineStarts, tagTrimmed.Pos())

			openingStartIndent := leadingWhitespace(text, openingLineStart)

			info := tokenInfo{
				openingPos:      openingPos,
				openingLine:     openingLine,
				openingColumn:   openingColumn,
				closingPos:      closingPos,
				closingLine:     closingLine,
				closingColumn:   closingColumn,
				tagLine:         tagLine,
				openingStartCol: reactutil.UTF16Length(openingStartIndent),
				openTab:         openingLineStart < len(text) && text[openingLineStart] == '\t',
				closeTab:        closingLineStart < len(text) && text[closingLineStart] == '\t',
				selfClosing:     selfClosing,
			}

			if len(attrs) > 0 {
				lastProp := attrs[len(attrs)-1]
				lpTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, lastProp)
				lpStart := lpTrimmed.Pos()
				lpEnd := lpTrimmed.End()
				lpFirstLine := scanner.ComputeLineOfPosition(lineStarts, lpStart)
				lpLastLine := scanner.ComputeLineOfPosition(lineStarts, lpEnd-1)
				lpLineStart := int(lineStarts[lpFirstLine])
				lpColumn := reactutil.UTF16Length(text[lpLineStart:lpStart])
				info.hasLastProp = true
				info.lastPropEnd = lpEnd
				info.lastPropFirstLine = lpFirstLine
				info.lastPropLastLine = lpLastLine
				info.lastPropColumn = lpColumn
			}

			// Locate the last comment **inside** the opening element that sits
			// AFTER the last attribute (or the tag name when zero attributes).
			// Mirrors upstream's `sourceCode.getCommentsInside(node).at(-1)` +
			// `lastComment.range[0] > (lastAttributeNode ?? node.name).range[1]`.
			gateEnd := tagTrimmed.End()
			if info.hasLastProp {
				gateEnd = info.lastPropEnd
			}
			info.hasTrailingComment, info.lastCommentEnd = findTrailingComment(text, gateEnd, closingPos)

			expectedLocation, disabled := getExpectedLocation(info, opts)
			if disabled {
				return
			}

			usingSameIndentation := true
			if expectedLocation == "tag-aligned" {
				usingSameIndentation = info.openTab == info.closeTab
			}

			// Upstream upgrade: when expectedLocation is `after-props` or
			// `after-tag` and the bracket is misplaced AND a trailing comment
			// exists, switch to `line-aligned`. Without this branch the fix
			// would move the closing bracket onto the comment's line, breaking
			// the comment.
			if (expectedLocation == "after-props" || expectedLocation == "after-tag") &&
				(!hasCorrectLocation(info, expectedLocation) || !usingSameIndentation) &&
				info.hasTrailingComment {
				expectedLocation = "line-aligned"
			}

			if hasCorrectLocation(info, expectedLocation) && usingSameIndentation {
				return
			}

			locationDesc, ok := locationMessages[expectedLocation]
			if !ok {
				return
			}

			details := ""
			expectedNextLine := info.hasLastProp && info.lastPropLastLine == info.closingLine
			correctColumn, hasCorrectColumn := getCorrectColumn(info, expectedLocation)
			if hasCorrectColumn {
				if expectedNextLine {
					details = fmt.Sprintf(" (expected column %d on the next line)", correctColumn+1)
				} else {
					details = fmt.Sprintf(" (expected column %d)", correctColumn+1)
				}
			}

			fix := buildFix(text, info, expectedLocation, expectedNextLine, correctColumn, lineStarts, elemEnd, tagTrimmed.End(), selfClosing)

			msg := rule.RuleMessage{
				Id:          "bracketLocation",
				Description: fmt.Sprintf("The closing bracket must be %s%s", locationDesc, details),
			}

			closingEnd := closingPos + 1
			if closingEnd > len(text) {
				closingEnd = len(text)
			}
			reportRange := core.NewTextRange(closingPos, closingEnd)

			if fix != nil {
				ctx.ReportRangeWithFixes(reportRange, msg, *fix)
			} else {
				ctx.ReportRange(reportRange, msg)
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

func getExpectedLocation(info tokenInfo, opts bracketOptions) (string, bool) {
	if !info.hasLastProp {
		return "after-tag", false
	}
	if info.openingLine == info.lastPropLastLine {
		return "after-props", false
	}
	if info.selfClosing {
		if opts.selfClosingOff {
			return "", true
		}
		return opts.selfClosing, false
	}
	if opts.nonEmptyOff {
		return "", true
	}
	return opts.nonEmpty, false
}

func hasCorrectLocation(info tokenInfo, expectedLocation string) bool {
	switch expectedLocation {
	case "after-tag":
		return info.tagLine == info.closingLine
	case "after-props":
		return info.lastPropLastLine == info.closingLine
	case "props-aligned", "tag-aligned", "line-aligned":
		col, ok := getCorrectColumn(info, expectedLocation)
		if !ok {
			return true
		}
		return col == info.closingColumn
	default:
		return true
	}
}

func getCorrectColumn(info tokenInfo, expectedLocation string) (int, bool) {
	switch expectedLocation {
	case "props-aligned":
		if !info.hasLastProp {
			return 0, false
		}
		return info.lastPropColumn, true
	case "tag-aligned":
		return info.openingColumn, true
	case "line-aligned":
		return info.openingStartCol, true
	default:
		return 0, false
	}
}

func getIndentation(text string, lineStarts []core.TextPos, info tokenInfo, expectedLocation string, correctColumn int) string {
	var indent string
	switch expectedLocation {
	case "props-aligned":
		if info.hasLastProp {
			indent = leadingWhitespace(text, int(lineStarts[info.lastPropFirstLine]))
		}
	case "tag-aligned", "line-aligned":
		indent = leadingWhitespace(text, int(lineStarts[info.openingLine]))
	}
	indentUTF16 := reactutil.UTF16Length(indent)
	if indentUTF16+1 < correctColumn+1 {
		// Non-whitespace characters precede the bracket on the reference line —
		// pad with spaces so the column matches.
		if correctColumn > indentUTF16 {
			indent = indent + strings.Repeat(" ", correctColumn-indentUTF16)
		}
	}
	return indent
}

func buildFix(text string, info tokenInfo, expectedLocation string, expectedNextLine bool, correctColumn int, lineStarts []core.TextPos, elemEnd int, tagNameEnd int, selfClosing bool) *rule.RuleFix {
	closingTag := ">"
	if selfClosing {
		closingTag = "/>"
	}

	var start, end int
	var newText string

	switch expectedLocation {
	case "after-tag":
		if info.hasLastProp {
			start = info.lastPropEnd
			end = elemEnd
			if expectedNextLine {
				newText = "\n" + closingTag
			} else {
				newText = closingTag
			}
		} else {
			start = tagNameEnd
			end = elemEnd
			if expectedNextLine {
				newText = "\n" + closingTag
			} else {
				newText = " " + closingTag
			}
		}
	case "after-props":
		if !info.hasLastProp {
			return nil
		}
		start = info.lastPropEnd
		end = elemEnd
		if expectedNextLine {
			newText = "\n" + closingTag
		} else {
			newText = closingTag
		}
	case "props-aligned", "tag-aligned", "line-aligned":
		// Upstream's `rangeStart = hasTrailingComment ? lastComment.range[1] :
		// cachedLastAttributeEndPos`. When a trailing comment exists, rewrite
		// from the byte AFTER the comment so the comment text stays intact.
		if info.hasTrailingComment {
			start = info.lastCommentEnd
		} else if info.hasLastProp {
			start = info.lastPropEnd
		} else {
			// Zero attributes AND no trailing comment — upstream's fixer
			// returns null for this branch (the `cachedLastAttributeEndPos`
			// it relies on is null). Mirror that: emit no fix.
			return nil
		}
		end = elemEnd
		indent := getIndentation(text, lineStarts, info, expectedLocation, correctColumn)
		newText = "\n" + indent + closingTag
	default:
		return nil
	}

	return &rule.RuleFix{
		Text:  newText,
		Range: core.NewTextRange(start, end),
	}
}

// findTrailingComment scans the byte range [gateEnd, closingPos) for `//` or
// `/* ... */` comments and returns whether one was found AND the end position
// of the LAST such comment (matching upstream's `at(-1)` selection). The scan
// is byte-oriented because (a) the scanner skips comment tokens — they are
// gone from the token stream by the time the rule visits the parent node, and
// (b) the bytes between the last attribute end and the closing bracket are by
// construction pure trivia (whitespace + comments only — no string / regex /
// template literal content can appear between attributes in an opening tag).
func findTrailingComment(text string, gateEnd, closingPos int) (bool, int) {
	if gateEnd < 0 {
		gateEnd = 0
	}
	if closingPos > len(text) {
		closingPos = len(text)
	}
	if gateEnd >= closingPos {
		return false, 0
	}
	found := false
	lastEnd := 0
	i := gateEnd
	for i < closingPos {
		c := text[i]
		switch {
		case c == '/' && i+1 < closingPos && text[i+1] == '/':
			// Line comment: consume up to the next LineTerminator (LF / CR /
			// LS / PS). The comment's end is the byte BEFORE the terminator,
			// matching ESLint's `Comment.range[1]` (exclusive of trailing \n).
			j := i + 2
			for j < closingPos {
				b := text[j]
				if b == '\n' || b == '\r' {
					break
				}
				if b == 0xE2 && j+2 < closingPos && text[j+1] == 0x80 &&
					(text[j+2] == 0xA8 || text[j+2] == 0xA9) {
					break
				}
				j++
			}
			found = true
			lastEnd = j
			i = j
		case c == '/' && i+1 < closingPos && text[i+1] == '*':
			// Block comment: consume up to and including `*/`.
			j := i + 2
			for j+1 < closingPos {
				if text[j] == '*' && text[j+1] == '/' {
					j += 2
					break
				}
				j++
			}
			found = true
			lastEnd = j
			i = j
		default:
			i++
		}
	}
	return found, lastEnd
}
