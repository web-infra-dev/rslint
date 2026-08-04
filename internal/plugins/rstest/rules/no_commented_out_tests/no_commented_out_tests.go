package no_commented_out_tests

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type commentBlockSegment struct {
	start   int
	comment *ast.CommentRange
}

type commentBlock struct {
	text     string
	segments []commentBlockSegment
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
		if r == utf8.RuneError && size == 1 || !unicode.IsSpace(r) {
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
		return rule.RuleListeners{}
	},
}
