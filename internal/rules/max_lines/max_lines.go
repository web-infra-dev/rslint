package max_lines

import (
	_ "embed"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed max_lines.schema.json
var schemaJSON []byte

// MaxLinesRule enforces a maximum number of lines per file.
// https://eslint.org/docs/latest/rules/max-lines
var MaxLinesRule = rule.Rule{
	Name:   "max-lines",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// The linter never fires a KindSourceFile listener, so run eagerly.
		checkMaxLines(ctx, parseOptions(options))
		return rule.RuleListeners{}
	},
}

type maxLinesOptions struct {
	max            int
	skipComments   bool
	skipBlankLines bool
}

// parseOptions reads the single option, which is either a bare maximum
// (`["error", 2]`) or an object (`["error", { max: 2, ... }]`).
func parseOptions(options []any) maxLinesOptions {
	result := maxLinesOptions{max: 300}
	if len(options) == 0 {
		return result
	}
	result.max = utils.ResolveLegacyMaxOption(options[0], 300)
	m, _ := options[0].(map[string]interface{})
	if v, ok := m["skipComments"].(bool); ok {
		result.skipComments = v
	}
	if v, ok := m["skipBlankLines"].(bool); ok {
		result.skipBlankLines = v
	}
	return result
}

func checkMaxLines(ctx rule.RuleContext, opts maxLinesOptions) {
	sourceFile := ctx.SourceFile
	text := sourceFile.Text()
	lineStarts := scanner.GetECMALineStarts(sourceFile)
	if len(lineStarts) == 0 {
		lineStarts = []core.TextPos{0}
	}
	realLineCount := len(lineStarts)
	// SourceCode.lines includes an empty final entry when the file ends in a
	// line terminator. ESLint drops that entry before enforcing max-lines.
	if realLineCount > 1 && int(lineStarts[realLineCount-1]) == len(text) {
		realLineCount--
	}

	var comments []*ast.CommentRange
	shebangEnd := 0
	if opts.skipComments {
		comments = ctx.Comments.All()
		shebangEnd = len(scanner.GetShebang(text))
		// With no comments there is nothing for skipComments to filter. This
		// also lets comment-free files use the direct-count path below.
		if len(comments) == 0 && shebangEnd == 0 {
			opts.skipComments = false
		}
	}

	excessIdx := opts.max
	if excessIdx < 0 {
		// ESLint's schema requires max >= 0; if a consumer bypasses the
		// schema, report from the first counted line rather than panic.
		excessIdx = 0
	}

	keptCount := realLineCount
	startLineIdx := excessIdx
	if opts.skipBlankLines || opts.skipComments {
		keptCount = 0
		startLineIdx = -1
		commentIdx := 0
		for i := range realLineCount {
			start := int(lineStarts[i])
			nextStart := len(text)
			contentEnd := nextStart
			if i+1 < len(lineStarts) {
				nextStart = int(lineStarts[i+1])
				contentEnd = utils.LineContentEnd(text, nextStart)
			}

			if opts.skipComments && lineIsCommentOnly(
				text,
				start,
				contentEnd,
				nextStart,
				comments,
				&commentIdx,
				shebangEnd,
			) {
				continue
			}
			if opts.skipBlankLines && ecmascript.IsBlank(text[start:contentEnd]) {
				continue
			}

			keptCount++
			if startLineIdx < 0 && keptCount > excessIdx {
				startLineIdx = i
			}
		}
	}

	if keptCount <= opts.max {
		return
	}
	if startLineIdx < 0 {
		// Only possible for an invalid negative max when every line is
		// skipped. There is no counted line to point at, so use file start.
		startLineIdx = 0
	}
	ctx.ReportRange(
		core.NewTextRange(int(lineStarts[startLineIdx]), len(text)),
		rule.RuleMessage{
			Id:          "exceed",
			Description: maxLinesDescription(keptCount, opts.max),
		},
	)
}

func maxLinesDescription(lines int, maximum int) string {
	var storage [128]byte
	description := append(storage[:0], "File has too many lines ("...)
	description = strconv.AppendInt(description, int64(lines), 10)
	description = append(description, "). Maximum allowed is "...)
	description = strconv.AppendInt(description, int64(maximum), 10)
	description = append(description, '.')
	return string(description)
}

// lineIsCommentOnly reports whether a line contains at least one comment and
// nothing but comments and ECMAScript whitespace. comments must be ordered by
// source position. commentIdx carries the first range that may overlap this or
// a later line, so the whole file and comment list are traversed only once.
func lineIsCommentOnly(
	text string,
	lineStart int,
	contentEnd int,
	nextLineStart int,
	comments []*ast.CommentRange,
	commentIdx *int,
	shebangEnd int,
) bool {
	firstComment := *commentIdx
	for firstComment < len(comments) && comments[firstComment].End() <= lineStart {
		firstComment++
	}
	*commentIdx = firstComment

	hasComment := false
	outsideStart := lineStart
	// CommentStore intentionally excludes the hashbang, while ESLint's
	// SourceCode.getAllComments includes it. Treat it as the first comment
	// without allocating a synthetic CommentRange.
	if lineStart == 0 && shebangEnd > 0 {
		hasComment = true
		outsideStart = min(shebangEnd, contentEnd)
	}

	for i := firstComment; i < len(comments); i++ {
		comment := comments[i]
		if comment.Pos() >= nextLineStart {
			break
		}
		if comment.End() <= lineStart {
			continue
		}
		hasComment = true

		commentStart := min(max(comment.Pos(), lineStart), contentEnd)
		if commentStart > outsideStart && !ecmascript.IsBlank(text[outsideStart:commentStart]) {
			return false
		}
		commentEnd := min(comment.End(), contentEnd)
		if commentEnd > outsideStart {
			outsideStart = commentEnd
		}
	}

	return hasComment && ecmascript.IsBlank(text[outsideStart:contentEnd])
}
