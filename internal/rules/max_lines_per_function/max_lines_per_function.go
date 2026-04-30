package max_lines_per_function

import (
	"fmt"
	"sort"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// MaxLinesPerFunctionRule enforces a maximum number of lines per function.
// https://eslint.org/docs/latest/rules/max-lines-per-function
var MaxLinesPerFunctionRule = rule.Rule{
	Name: "max-lines-per-function",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		state := newLineState(ctx.SourceFile)

		process := func(node *ast.Node) {
			// Overload signatures, abstract / declare members, and TS interface
			// methods reach this listener with no body. ESLint never visits
			// them (they don't exist as FunctionExpression in ESTree), so
			// produce no diagnostic for parity.
			if node.Body() == nil {
				return
			}
			if !opts.iifes && isIIFE(node) {
				return
			}

			startLine, endLine := state.nodeLineRange(node)
			lineCount := 0
			for i := startLine; i <= endLine; i++ {
				if opts.skipComments && state.isFullLineCommentLine(i) {
					continue
				}
				if opts.skipBlankLines && utils.IsECMABlankLine(state.lineContent(i)) {
					continue
				}
				lineCount++
			}

			if lineCount > opts.max {
				name := upperCaseFirst(utils.GetFunctionNameWithKind(node))
				ctx.ReportNode(node, rule.RuleMessage{
					Id: "exceed",
					Description: fmt.Sprintf(
						"%s has too many lines (%d). Maximum allowed is %d.",
						name, lineCount, opts.max,
					),
				})
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: process,
			ast.KindFunctionExpression:  process,
			ast.KindArrowFunction:       process,
			ast.KindMethodDeclaration:   process,
			ast.KindGetAccessor:         process,
			ast.KindSetAccessor:         process,
			ast.KindConstructor:         process,
		}
	},
}

type maxLinesPerFunctionOptions struct {
	max            int
	skipComments   bool
	skipBlankLines bool
	iifes          bool
}

func parseOptions(opts any) maxLinesPerFunctionOptions {
	result := maxLinesPerFunctionOptions{max: 50}
	if opts == nil {
		return result
	}
	// JS tests pass options as [50] or [{ max: 50, ... }]; the CLI may pass a
	// bare number, bare object, or array-wrapped value.
	if arr, ok := opts.([]interface{}); ok {
		if len(arr) == 0 {
			return result
		}
		opts = arr[0]
	}
	if n, ok := toInt(opts); ok {
		result.max = n
		return result
	}
	if m, ok := opts.(map[string]interface{}); ok {
		if v, ok := m["max"]; ok {
			if n, ok := toInt(v); ok {
				result.max = n
			}
		}
		if v, ok := m["skipComments"].(bool); ok {
			result.skipComments = v
		}
		if v, ok := m["skipBlankLines"].(bool); ok {
			result.skipBlankLines = v
		}
		if v, ok := m["IIFEs"].(bool); ok {
			result.iifes = v
		}
	}
	return result
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case float32:
		return int(n), true
	}
	return 0, false
}

// upperCaseFirst mirrors ESLint's shared/string-utils upperCaseFirst — used to
// capitalize the leading word of `getFunctionNameWithKind`'s output before
// embedding it into the diagnostic message ("function 'foo'" → "Function
// 'foo'", "arrow function" → "Arrow function").
func upperCaseFirst(s string) string {
	for i, r := range s {
		return string(unicode.ToUpper(r)) + s[i+len(string(r)):]
	}
	return s
}

// isIIFE reports whether the given function-like node is the callee of a call
// expression — matching ESLint's astUtils-style IIFE check, but accounting for
// tsgo's explicit `ParenthesizedExpression` wrappers (ESTree strips parens, so
// `(function(){})()` is a call whose callee is the function directly; tsgo
// preserves the parens, requiring an unwrap step).
//
// Method / getter / setter / constructor nodes can never be a CallExpression
// callee, so they're filtered out up front.
func isIIFE(node *ast.Node) bool {
	if node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction {
		return false
	}
	cur := node
	for cur.Parent != nil && ast.IsParenthesizedExpression(cur.Parent) {
		cur = cur.Parent
	}
	parent := cur.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	return parent.AsCallExpression().Expression == cur
}

// lineState caches per-source-file derived data shared across every function
// visited by the rule (line starts, comment-line table). Computing these
// eagerly per-file is cheaper than re-deriving them per function in files with
// many small functions.
type lineState struct {
	sourceFile *ast.SourceFile
	text       string
	lineStarts []core.TextPos
	nLines     int
	// lineComment[i] is the last comment whose range touches line i (1-indexed)
	// — mirroring ESLint's `getCommentLineNumbers` map. nil entries mean no
	// comment on that line.
	lineComment []*ast.CommentRange
}

func newLineState(sourceFile *ast.SourceFile) *lineState {
	text := sourceFile.Text()
	lineStarts := scanner.GetECMALineStarts(sourceFile)
	if len(lineStarts) == 0 {
		lineStarts = []core.TextPos{0}
	}
	nLines := len(lineStarts)

	var comments []*ast.CommentRange
	// ESLint's sourceCode.getAllComments() includes the hashbang (`#!`) line,
	// but tsgo's ForEachComment skips past it. Synthesize a comment range so
	// skipComments filters it the same way ESLint does.
	if shebang := scanner.GetShebang(text); shebang != "" {
		comments = append(comments, &ast.CommentRange{
			TextRange: core.NewTextRange(0, len(shebang)),
			Kind:      ast.KindSingleLineCommentTrivia,
		})
	}
	utils.ForEachComment(sourceFile.AsNode(), func(c *ast.CommentRange) {
		copied := *c
		comments = append(comments, &copied)
	}, sourceFile)
	// ForEachComment may surface a comment twice (once as a token's trailing
	// range, once as the next token's leading range) and not strictly in source
	// order. Sort so a deterministic "last write wins" assignment to
	// lineComment matches ESLint's iteration order.
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].Pos() < comments[j].Pos()
	})

	// nLines lines (0-indexed) → store at index = line. We use 0-indexed lines
	// throughout this rule for consistency with scanner.ComputeLineOfPosition.
	lineComment := make([]*ast.CommentRange, nLines)
	for _, cmt := range comments {
		if cmt.End() <= cmt.Pos() {
			continue
		}
		startLine := scanner.ComputeLineOfPosition(lineStarts, cmt.Pos())
		endLine := scanner.ComputeLineOfPosition(lineStarts, cmt.End()-1)
		for l := startLine; l <= endLine && l < nLines; l++ {
			if l >= 0 {
				lineComment[l] = cmt
			}
		}
	}

	return &lineState{
		sourceFile:  sourceFile,
		text:        text,
		lineStarts:  lineStarts,
		nLines:      nLines,
		lineComment: lineComment,
	}
}

// nodeLineRange returns the inclusive 0-indexed line range of the node,
// excluding any leading trivia. Mirrors ESLint's `node.loc.start.line` /
// `node.loc.end.line` (which are based on the first / last source token of
// the node).
func (s *lineState) nodeLineRange(node *ast.Node) (int, int) {
	startPos := utils.TrimNodeTextRange(s.sourceFile, node).Pos()
	endPos := node.End()
	startLine := scanner.ComputeLineOfPosition(s.lineStarts, startPos)
	endLineIdx := endPos - 1
	if endLineIdx < startPos {
		endLineIdx = startPos
	}
	endLine := scanner.ComputeLineOfPosition(s.lineStarts, endLineIdx)
	return startLine, endLine
}

// lineContent returns the content of the i-th 0-indexed line, without its
// terminator — matching ESLint's `sourceCode.lines[i]`.
func (s *lineState) lineContent(i int) string {
	if i < 0 || i >= s.nLines {
		return ""
	}
	start := int(s.lineStarts[i])
	if i+1 >= s.nLines {
		return s.text[start:]
	}
	return s.text[start:utils.LineContentEnd(s.text, int(s.lineStarts[i+1]))]
}

// isFullLineCommentLine reports whether the given 0-indexed line is entirely
// covered by a single comment (i.e. there is no source code on it). Mirrors
// ESLint's `isFullLineComment` via the `getCommentLineNumbers` map: only the
// LAST comment touching this line is consulted, so a line like
// `/* a */ /* b */` is treated as having code (because /* a */ precedes /* b */).
func (s *lineState) isFullLineCommentLine(i int) bool {
	if i < 0 || i >= s.nLines {
		return false
	}
	cmt := s.lineComment[i]
	if cmt == nil {
		return false
	}
	startLine := scanner.ComputeLineOfPosition(s.lineStarts, cmt.Pos())
	endLine := scanner.ComputeLineOfPosition(s.lineStarts, cmt.End()-1)
	line := s.lineContent(i)

	// Mimics ESLint's two arms:
	//   isFirstTokenOnLine: comment starts on this line && nothing before it
	//   isLastTokenOnLine : comment ends on this line && nothing after it
	// Either arm passes when the comment crosses past this line on that side.
	startOK := startLine < i
	if !startOK && startLine == i {
		col := cmt.Pos() - int(s.lineStarts[i])
		if col >= 0 && col <= len(line) && utils.IsECMABlankLine(line[:col]) {
			startOK = true
		}
	}
	if !startOK {
		return false
	}

	if endLine > i {
		return true
	}
	if endLine == i {
		col := cmt.End() - int(s.lineStarts[i])
		if col >= 0 && col <= len(line) && utils.IsECMABlankLine(line[col:]) {
			return true
		}
	}
	return false
}
