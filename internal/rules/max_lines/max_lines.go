package max_lines

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// MaxLinesRule enforces a maximum number of lines per file.
// https://eslint.org/docs/latest/rules/max-lines
var MaxLinesRule = rule.Rule{
	Name: "max-lines",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// The linter never fires a KindSourceFile listener, so run eagerly.
		checkMaxLines(ctx, options)
		return rule.RuleListeners{}
	},
}

type maxLinesOptions struct {
	max            int
	skipComments   bool
	skipBlankLines bool
}

func parseOptions(opts any) maxLinesOptions {
	result := maxLinesOptions{max: 300}
	if opts == nil {
		return result
	}
	// JS tests pass options as [2] or [{ max: 2, ... }].
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

func checkMaxLines(ctx rule.RuleContext, options any) {
	opts := parseOptions(options)
	sourceFile := ctx.SourceFile
	text := sourceFile.Text()
	lineStarts := scanner.GetECMALineStarts(sourceFile)
	if len(lineStarts) == 0 {
		lineStarts = []core.TextPos{0}
	}
	nLines := len(lineStarts)

	// Slice out a single line's content (excluding its terminator). Matches the
	// semantics of ESLint's SourceCode.lines entries.
	lineContent := func(i int) string {
		start := int(lineStarts[i])
		if i+1 >= nLines {
			return text[start:]
		}
		return text[start:utils.LineContentEnd(text, int(lineStarts[i+1]))]
	}

	// If the file ends with a line terminator, the final entry is an extra
	// empty string that doesn't represent a real line. Drop it from counting
	// but keep it visible to the end-position calculation below.
	lastLineIsTrailingEmpty := nLines > 1 && lineContent(nLines-1) == ""

	keep := make([]bool, nLines)
	for i := range keep {
		keep[i] = true
	}
	if lastLineIsTrailingEmpty {
		keep[nLines-1] = false
	}

	if opts.skipBlankLines {
		for i := range nLines {
			if keep[i] && utils.IsECMABlankLine(lineContent(i)) {
				keep[i] = false
			}
		}
	}

	if opts.skipComments {
		for line := range commentOnlyLines(sourceFile, text, lineStarts) {
			idx := line - 1
			if idx >= 0 && idx < nLines {
				keep[idx] = false
			}
		}
	}

	kept := make([]int, 0, nLines)
	for i := range nLines {
		if keep[i] {
			kept = append(kept, i)
		}
	}
	if len(kept) <= opts.max {
		return
	}
	excessIdx := opts.max
	if excessIdx < 0 {
		// ESLint's schema requires max >= 0; if a consumer bypasses the
		// schema and supplies a negative value, report from the first line
		// rather than panic.
		excessIdx = 0
	}
	startLineIdx := kept[excessIdx]

	ctx.ReportRange(
		core.NewTextRange(int(lineStarts[startLineIdx]), len(text)),
		rule.RuleMessage{
			Id:          "exceed",
			Description: fmt.Sprintf("File has too many lines (%d). Maximum allowed is %d.", len(kept), opts.max),
		},
	)
}

// commentOnlyLines returns the set of 1-indexed line numbers that contain only
// comments and whitespace, matching ESLint's max-lines `getLinesWithoutCode`.
// A multi-line comment's first (last) line is excluded when non-comment code
// exists earlier (later) on that same line.
func commentOnlyLines(sourceFile *ast.SourceFile, text string, lineStarts []core.TextPos) map[int]bool {
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
	if len(comments) == 0 {
		return nil
	}
	// ForEachComment may surface a comment twice (once as a token's trailing
	// range, once as the next token's leading range) and not strictly in source
	// order. Duplicates are harmless for this algorithm, but advancing
	// commentIdx linearly requires sorted input.
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].Pos() < comments[j].Pos()
	})

	nLines := len(lineStarts)
	// minCodePos[line] / maxCodeEnd[line] bound the non-comment,
	// non-whitespace characters on each 0-indexed line. -1 means "none".
	minCodePos := make([]int, nLines)
	maxCodeEnd := make([]int, nLines)
	for i := range minCodePos {
		minCodePos[i] = -1
		maxCodeEnd[i] = -1
	}

	commentIdx := 0
	line := 0
	i := 0
	for i < len(text) {
		r, size := utf8.DecodeRuneInString(text[i:])
		// CRLF collapses to a single line break.
		if r == '\r' {
			if i+1 < len(text) && text[i+1] == '\n' {
				i += 2
			} else {
				i++
			}
			line++
			continue
		}
		if r == '\n' || r == 0x2028 || r == 0x2029 {
			i += size
			line++
			continue
		}

		// Advance past comments that ended at or before i.
		for commentIdx < len(comments) && comments[commentIdx].End() <= i {
			commentIdx++
		}
		// Inside a comment — ignore.
		if commentIdx < len(comments) && comments[commentIdx].Pos() <= i && i < comments[commentIdx].End() {
			i += size
			continue
		}

		// Line-terminator runes are consumed above, so at this point
		// IsStrWhiteSpace only matches ECMAScript WhiteSpace.
		if utils.IsStrWhiteSpace(r) {
			i += size
			continue
		}

		if minCodePos[line] == -1 {
			minCodePos[line] = i
		}
		maxCodeEnd[line] = i + size
		i += size
	}

	commentOnly := make(map[int]bool)
	for _, cmt := range comments {
		startPos := cmt.Pos()
		endPos := cmt.End()
		if endPos <= startPos {
			continue
		}
		startLine := scanner.ComputeLineOfPosition(lineStarts, startPos)
		endLine := scanner.ComputeLineOfPosition(lineStarts, endPos-1)

		if mc := minCodePos[startLine]; mc != -1 && mc < startPos {
			startLine++
		}
		if me := maxCodeEnd[endLine]; me != -1 && me > endPos {
			endLine--
		}
		for l := startLine; l <= endLine; l++ {
			if l >= 0 && l < nLines {
				commentOnly[l+1] = true
			}
		}
	}
	return commentOnly
}
