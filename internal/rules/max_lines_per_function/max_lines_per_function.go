package max_lines_per_function

import (
	_ "embed"
	"fmt"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed max_lines_per_function.schema.json
var schemaJSON []byte

// MaxLinesPerFunctionRule enforces a maximum number of lines per function.
// https://eslint.org/docs/latest/rules/max-lines-per-function
var MaxLinesPerFunctionRule = rule.Rule{
	Name:   "max-lines-per-function",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		sourceFile := ctx.SourceFile
		text := sourceFile.Text()
		state := &lineState{}

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

			startPos := scanner.GetTokenPosOfNode(node, sourceFile, false)
			endPos := node.End()
			textRange := core.NewTextRange(startPos, endPos)
			lineCount := 0
			if len(state.lineStarts) == 0 && opts.max >= 1 && isSingleLineRange(text, startPos, endPos) {
				lineCount = 1
			} else {
				if len(state.lineStarts) == 0 {
					var comments []*ast.CommentRange
					if opts.skipComments {
						comments = ctx.Comments.All()
					}
					state.initialize(sourceFile, comments, opts.skipComments, opts.skipBlankLines)
				}
				startLine, endLine := state.lineRange(startPos, endPos)
				lineCount = state.countLines(startLine, endLine)
			}

			if lineCount > opts.max {
				name := upperCaseFirst(utils.GetFunctionNameWithKind(node))
				ctx.ReportRange(textRange, rule.RuleMessage{
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

// isSingleLineRange reports whether a validated source range contains no
// ECMAScript line terminator. A single-line function always counts as exactly
// one line, even when comment or blank-line skipping is enabled, so callers can
// avoid materializing the source file's line map and comment table entirely.
// Synthetic or otherwise unusual ranges retain the established line-map path.
func isSingleLineRange(text string, startPos, endPos int) bool {
	return startPos >= 0 && endPos >= startPos && endPos <= len(text) &&
		!ecmascript.ContainsLineTerminator(text, startPos, endPos)
}

type maxLinesPerFunctionOptions struct {
	max            int
	skipComments   bool
	skipBlankLines bool
	iifes          bool
}

// parseOptions reads the single option, which is either a bare maximum
// (`["error", 50]`) or an object (`["error", { max: 50, ... }]`).
func parseOptions(options []any) maxLinesPerFunctionOptions {
	result := maxLinesPerFunctionOptions{max: 50}
	if len(options) == 0 {
		return result
	}
	result.max = utils.ResolveLegacyMaxOption(options[0], 50)
	m, _ := options[0].(map[string]interface{})
	if v, ok := m["skipComments"].(bool); ok {
		result.skipComments = v
	}
	if v, ok := m["skipBlankLines"].(bool); ok {
		result.skipBlankLines = v
	}
	if v, ok := m["IIFEs"].(bool); ok {
		result.iifes = v
	}
	return result
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
// visited by the rule. When a skip option is enabled, countedLinePrefix turns
// each function's line count into a pair of indexed lookups. This is especially
// important for nested functions, whose source ranges overlap.
type lineState struct {
	text              string
	lineStarts        []core.TextPos
	nLines            int
	fullLineComment   []bool
	skipBlankLines    bool
	prefixStartLine   int
	prefixEndLine     int
	countedLinePrefix []int
}

func (s *lineState) initialize(
	sourceFile *ast.SourceFile,
	sourceComments []*ast.CommentRange,
	skipComments bool,
	skipBlankLines bool,
) {
	text := sourceFile.Text()
	lineStarts := scanner.GetECMALineStarts(sourceFile)
	if len(lineStarts) == 0 {
		lineStarts = []core.TextPos{0}
	}
	nLines := len(lineStarts)
	s.text = text
	s.lineStarts = lineStarts
	s.nLines = nLines
	s.skipBlankLines = skipBlankLines
	if !skipComments && !skipBlankLines {
		return
	}

	if skipComments {
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
		comments = append(comments, sourceComments...)
		if len(comments) == 0 {
			return
		}

		fullLineComment := make([]bool, nLines)
		fullLineCommentCount := 0
		// CommentStore guarantees that sourceComments is ordered and
		// deduplicated. Assigning rather than accumulating is intentional:
		// ESLint only consults the last comment touching a line.
		for _, cmt := range comments {
			if cmt.End() <= cmt.Pos() {
				continue
			}
			startLine := scanner.ComputeLineOfPosition(lineStarts, cmt.Pos())
			endLine := scanner.ComputeLineOfPosition(lineStarts, cmt.End()-1)
			for line := max(startLine, 0); line <= endLine && line < nLines; line++ {
				coversWholeLine := s.commentCoversWholeLine(cmt, startLine, endLine, line)
				if coversWholeLine != fullLineComment[line] {
					if coversWholeLine {
						fullLineCommentCount++
					} else {
						fullLineCommentCount--
					}
					fullLineComment[line] = coversWholeLine
				}
			}
		}
		if fullLineCommentCount > 0 {
			s.fullLineComment = fullLineComment
		}
	}
}

// lineRange returns the inclusive 0-indexed line range for an already-trimmed
// source range. Mirrors ESLint's `node.loc.start.line` / `node.loc.end.line`.
func (s *lineState) lineRange(startPos, endPos int) (int, int) {
	startLine := scanner.ComputeLineOfPosition(s.lineStarts, startPos)
	endLineIdx := endPos - 1
	if endLineIdx < startPos {
		endLineIdx = startPos
	}
	endLine := scanner.ComputeLineOfPosition(s.lineStarts, endLineIdx)
	return startLine, endLine
}

func (s *lineState) countLines(startLine, endLine int) int {
	if !s.skipBlankLines && len(s.fullLineComment) == 0 {
		return endLine - startLine + 1
	}

	if len(s.countedLinePrefix) == 0 || startLine > s.prefixEndLine+1 {
		s.resetCountedLinePrefix(startLine, endLine)
	} else if startLine < s.prefixStartLine {
		// Listener traversal is source-ordered, so this is only reachable for a
		// synthetic or manually invoked node. Keep that uncommon path correct
		// without making the normal cache more complex.
		return s.countLinesDirect(startLine, endLine)
	} else if endLine > s.prefixEndLine {
		s.extendCountedLinePrefix(endLine)
	}

	startIndex := startLine - s.prefixStartLine
	endIndex := endLine - s.prefixStartLine + 1
	return s.countedLinePrefix[endIndex] - s.countedLinePrefix[startIndex]
}

func (s *lineState) resetCountedLinePrefix(startLine, endLine int) {
	needed := endLine - startLine + 2
	if cap(s.countedLinePrefix) < needed {
		s.countedLinePrefix = make([]int, needed)
	} else {
		s.countedLinePrefix = s.countedLinePrefix[:needed]
	}
	s.countedLinePrefix[0] = 0
	for line := startLine; line <= endLine; line++ {
		index := line - startLine + 1
		s.countedLinePrefix[index] = s.countedLinePrefix[index-1]
		if s.lineIsCounted(line) {
			s.countedLinePrefix[index]++
		}
	}
	s.prefixStartLine = startLine
	s.prefixEndLine = endLine
}

func (s *lineState) extendCountedLinePrefix(endLine int) {
	additional := endLine - s.prefixEndLine
	needed := len(s.countedLinePrefix) + additional
	if needed > cap(s.countedLinePrefix) {
		// Once adjacent function ranges reveal a dense cluster, grow in chunks
		// instead of reallocating for every short function. Disjoint ranges reset
		// and reuse this buffer, so sparse files do not reserve for their gaps.
		newCapacity := max(cap(s.countedLinePrefix)*2, 1024, needed)
		maxCapacity := s.nLines - s.prefixStartLine + 1
		newCapacity = min(newCapacity, maxCapacity)
		prefix := make([]int, len(s.countedLinePrefix), newCapacity)
		copy(prefix, s.countedLinePrefix)
		s.countedLinePrefix = prefix
	}
	for line := s.prefixEndLine + 1; line <= endLine; line++ {
		count := s.countedLinePrefix[len(s.countedLinePrefix)-1]
		if s.lineIsCounted(line) {
			count++
		}
		s.countedLinePrefix = append(s.countedLinePrefix, count)
	}
	s.prefixEndLine = endLine
}

func (s *lineState) countLinesDirect(startLine, endLine int) int {
	count := 0
	for line := startLine; line <= endLine; line++ {
		if s.lineIsCounted(line) {
			count++
		}
	}
	return count
}

func (s *lineState) lineIsCounted(line int) bool {
	if len(s.fullLineComment) > 0 && s.fullLineComment[line] {
		return false
	}
	return !s.skipBlankLines || !ecmascript.IsBlank(s.lineContent(line))
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

// commentCoversWholeLine reports whether cmt is the only non-whitespace source
// on line. The caller processes comments in source order and retains the last
// result for each line, matching ESLint's getCommentLineNumbers map.
func (s *lineState) commentCoversWholeLine(cmt *ast.CommentRange, startLine, endLine, lineIndex int) bool {
	line := s.lineContent(lineIndex)

	// Mimics ESLint's two arms:
	//   isFirstTokenOnLine: comment starts on this line && nothing before it
	//   isLastTokenOnLine : comment ends on this line && nothing after it
	// Either arm passes when the comment crosses past this line on that side.
	startOK := startLine < lineIndex
	if !startOK && startLine == lineIndex {
		col := cmt.Pos() - int(s.lineStarts[lineIndex])
		if col >= 0 && col <= len(line) && ecmascript.IsBlank(line[:col]) {
			startOK = true
		}
	}
	if !startOK {
		return false
	}

	if endLine > lineIndex {
		return true
	}
	if endLine == lineIndex {
		col := cmt.End() - int(s.lineStarts[lineIndex])
		if col >= 0 && col <= len(line) && ecmascript.IsBlank(line[col:]) {
			return true
		}
	}
	return false
}
