package no_commented_out_tests

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

type commentBlockSegment struct {
	start   int
	comment *ast.CommentRange
}

type commentBlock struct {
	text     string
	segments []commentBlockSegment
}

func startsRstestTypeOperator(text, operator string) bool {
	if !strings.HasPrefix(text, operator) {
		return false
	}
	return len(text) == len(operator) ||
		text[len(operator)] == ' ' ||
		text[len(operator)] == '\t' ||
		text[len(operator)] == '\v' ||
		text[len(operator)] == '\f' ||
		text[len(operator)] >= utf8.RuneSelf ||
		text[len(operator)] == '/'
}

func skipRstestHorizontalWhitespace(text string, offset int) int {
	for offset < len(text) {
		switch text[offset] {
		case ' ', '\t', '\v', '\f':
			offset++
			continue
		}
		if text[offset] < utf8.RuneSelf {
			break
		}
		char, size := utf8.DecodeRuneInString(text[offset:])
		if !ecmascript.IsWhiteSpace(char) {
			break
		}
		offset += size
	}
	return offset
}

func startsRstestRootCandidate(text string) bool {
	offset := 0
	for offset < len(text) && (text[offset] == ' ' || text[offset] == '\t') {
		offset++
	}
	for offset < len(text) && (text[offset] == ';' || text[offset] == '(') {
		offset++
		for offset < len(text) && (text[offset] == ' ' || text[offset] == '\t') {
			offset++
		}
	}
	for _, root := range [...]string{"test", "it", "describe"} {
		if !strings.HasPrefix(text[offset:], root) {
			continue
		}
		end := offset + len(root)
		if end < len(text) {
			next := text[end]
			if next == '_' || next == '$' ||
				(next >= '0' && next <= '9') ||
				(next >= 'A' && next <= 'Z') ||
				(next >= 'a' && next <= 'z') {
				continue
			}
		}
		end = skipRstestHorizontalWhitespace(text, end)
		for end < len(text) && text[end] == ')' {
			end++
			end = skipRstestHorizontalWhitespace(text, end)
		}
		if end == len(text) {
			// A root at physical line end may continue in the next adjacent
			// comment, e.g. `// test` followed by `//   .only(...)`.
			return true
		}
		switch text[end] {
		case '(', '.', '[', '?', '!', '/':
			return true
		}
		if startsRstestTypeOperator(text[end:], "as") ||
			startsRstestTypeOperator(text[end:], "satisfies") {
			return true
		}
	}
	return false
}

func blockBodyMayContainRstestRoot(body string) bool {
	for {
		lineEnd := len(body)
		if index := strings.IndexAny(body, "\r\n\u2028\u2029"); index >= 0 {
			lineEnd = index
		}
		line := body[:lineEnd]
		offset := 0
		for offset < len(line) && (line[offset] == ' ' || line[offset] == '\t') {
			offset++
		}
		if offset < len(line) && line[offset] == '*' {
			offset++
			if offset < len(line) && (line[offset] == ' ' || line[offset] == '\t') {
				offset++
			}
		}
		if startsRstestRootCandidate(line[offset:]) {
			return true
		}
		if lineEnd == len(body) {
			return false
		}
		body = body[lineEnd+lineTerminatorLength(body, lineEnd):]
	}
}

// mayContainCommentedRstestRoot is a conservative byte-level gate before the
// expensive canonical comment collection and candidate parsing. It may return
// true for comment-like text inside strings or templates, but every real line
// or block comment whose normalized logical line starts with a supported root
// returns true and proceeds through the existing exact analyzer.
func mayContainCommentedRstestRoot(sourceText string) bool {
	for searchStart := 0; searchStart < len(sourceText); {
		relativeSlash := strings.IndexByte(sourceText[searchStart:], '/')
		if relativeSlash < 0 {
			return false
		}
		slash := searchStart + relativeSlash
		if slash+1 >= len(sourceText) {
			return false
		}
		switch sourceText[slash+1] {
		case '/':
			end := len(sourceText)
			if index := strings.IndexAny(sourceText[slash+2:], "\r\n\u2028\u2029"); index >= 0 {
				end = slash + 2 + index
			}
			if startsRstestRootCandidate(sourceText[slash+2 : end]) {
				return true
			}
			// This byte pair may live inside a string or template. Keep
			// searching from the second slash so a later real comment on the
			// same physical line cannot be skipped.
			searchStart = slash + 1
		case '*':
			end := len(sourceText)
			if index := strings.Index(sourceText[slash+2:], "*/"); index >= 0 {
				end = slash + 2 + index
			}
			if blockBodyMayContainRstestRoot(sourceText[slash+2 : end]) {
				return true
			}
			// As above, do not trust that this is lexically a comment. A later
			// real comment may begin before the apparent closing delimiter.
			searchStart = slash + 1
		default:
			searchStart = slash + 1
		}
	}
	return false
}

func lineTerminatorLength(text string, offset int) int {
	if offset < 0 || offset >= len(text) {
		return 0
	}
	switch text[offset] {
	case '\n':
		return 1
	case '\r':
		if offset+1 < len(text) && text[offset+1] == '\n' {
			return 2
		}
		return 1
	}
	if strings.HasPrefix(text[offset:], "\u2028") || strings.HasPrefix(text[offset:], "\u2029") {
		return len("\u2028")
	}
	return 0
}

func normalizeBlockCommentBody(body string) string {
	var normalized strings.Builder
	normalized.Grow(len(body))

	for offset := 0; offset < len(body); {
		lineEnd := offset
		for lineEnd < len(body) && lineTerminatorLength(body, lineEnd) == 0 {
			_, size := utf8.DecodeRuneInString(body[lineEnd:])
			lineEnd += size
		}

		content := body[offset:lineEnd]
		leaderEnd := 0
		for leaderEnd < len(content) && (content[leaderEnd] == ' ' || content[leaderEnd] == '\t') {
			leaderEnd++
		}
		if leaderEnd < len(content) && content[leaderEnd] == '*' {
			leaderEnd++
			if leaderEnd < len(content) && (content[leaderEnd] == ' ' || content[leaderEnd] == '\t') {
				leaderEnd++
			}
			content = content[leaderEnd:]
		}
		normalized.WriteString(content)

		terminatorLength := lineTerminatorLength(body, lineEnd)
		if terminatorLength == 0 {
			break
		}
		// The normalized text is synthetic and block comments always map back
		// to one physical comment. Canonicalizing every ECMAScript line
		// terminator to LF lets the multiline candidate matcher behave
		// consistently for CR, CRLF, U+2028, and U+2029.
		normalized.WriteByte('\n')
		offset = lineEnd + terminatorLength
	}

	return normalized.String()
}

func buildCommentedTestsMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "commentedTests",
		Description: "Do not comment out tests",
	}
}

func commentBody(sourceText string, comment *ast.CommentRange) (string, bool) {
	if comment == nil || comment.Pos() < 0 || comment.End() > len(sourceText) {
		return "", false
	}

	switch comment.Kind {
	case ast.KindSingleLineCommentTrivia:
		if comment.Pos()+2 > comment.End() || sourceText[comment.Pos():comment.Pos()+2] != "//" {
			return "", false
		}
		return sourceText[comment.Pos()+2 : comment.End()], true
	case ast.KindMultiLineCommentTrivia:
		if comment.Pos()+4 > comment.End() ||
			sourceText[comment.Pos():comment.Pos()+2] != "/*" ||
			sourceText[comment.End()-2:comment.End()] != "*/" {
			return "", false
		}
		// Documentation examples are not disabled tests. Regular block comments
		// remain eligible, including multiline comments containing full test
		// bodies.
		if sourceText[comment.Pos():comment.Pos()+3] == "/**" {
			return "", false
		}
		return normalizeBlockCommentBody(sourceText[comment.Pos()+2 : comment.End()-2]), true
	default:
		return "", false
	}
}

func containsExactlyOneLineTerminator(text string) bool {
	count := 0
	for offset := 0; offset < len(text); {
		if size := lineTerminatorLength(text, offset); size != 0 {
			count++
			if count > 1 {
				return false
			}
			offset += size
			continue
		}

		r, size := utf8.DecodeRuneInString(text[offset:])
		if r == utf8.RuneError && size == 1 || !ecmascript.IsWhiteSpace(r) {
			return false
		}
		offset += size
	}
	return count == 1
}

func lineCommentsAreAdjacent(sourceText string, current, next *ast.CommentRange) bool {
	if current == nil || next == nil ||
		current.Kind != ast.KindSingleLineCommentTrivia ||
		next.Kind != ast.KindSingleLineCommentTrivia ||
		current.End() > next.Pos() {
		return false
	}

	between := sourceText[current.End():next.Pos()]
	return containsExactlyOneLineTerminator(between)
}

func buildCommentBlocks(sourceText string, comments []*ast.CommentRange) []commentBlock {
	blocks := make([]commentBlock, 0, len(comments))

	for i := 0; i < len(comments); {
		comment := comments[i]
		body, ok := commentBody(sourceText, comment)
		if !ok {
			i++
			continue
		}

		var text strings.Builder
		text.Grow(len(body))
		text.WriteString(body)
		block := commentBlock{
			segments: []commentBlockSegment{{
				start:   0,
				comment: comment,
			}},
		}
		i++

		if comment.Kind == ast.KindSingleLineCommentTrivia {
			previous := comment
			for i < len(comments) && lineCommentsAreAdjacent(sourceText, previous, comments[i]) {
				next := comments[i]
				nextBody, nextOK := commentBody(sourceText, next)
				if !nextOK {
					break
				}
				text.WriteByte('\n')
				block.segments = append(block.segments, commentBlockSegment{
					start:   text.Len(),
					comment: next,
				})
				text.WriteString(nextBody)
				previous = next
				i++
			}
		}

		block.text = text.String()
		blocks = append(blocks, block)
	}

	return blocks
}

func (block commentBlock) commentAt(offset int) *ast.CommentRange {
	if len(block.segments) == 0 {
		return nil
	}
	for i := len(block.segments) - 1; i >= 0; i-- {
		if offset >= block.segments[i].start {
			return block.segments[i].comment
		}
	}
	return block.segments[0].comment
}

var NoCommentedOutTestsRule = rule.Rule{
	Name:   "rstest/no-commented-out-tests",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		sourceText := ctx.SourceFile.Text()
		if !mayContainCommentedRstestRoot(sourceText) {
			return nil
		}
		for _, block := range buildCommentBlocks(sourceText, ctx.Comments.All()) {
			reportedComments := make(map[*ast.CommentRange]struct{})
			for _, offset := range findCommentedRstestRegistrations(block.text, ctx.SourceFile.ScriptKind) {
				comment := block.commentAt(offset)
				if comment == nil {
					continue
				}
				if _, reported := reportedComments[comment]; reported {
					continue
				}
				reportedComments[comment] = struct{}{}
				ctx.ReportRange(
					core.NewTextRange(comment.Pos(), comment.End()),
					buildCommentedTestsMessage(),
				)
			}
		}
		return nil
	},
}
