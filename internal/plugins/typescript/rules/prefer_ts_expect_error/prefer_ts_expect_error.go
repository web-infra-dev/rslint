package prefer_ts_expect_error

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const tsIgnoreDirective = "@ts-ignore"
const tsExpectErrorDirective = "@ts-expect-error"

func buildPreferExpectErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferExpectErrorComment",
		Description: "Use @ts-expect-error instead of @ts-ignore.",
	}
}

func findDirectiveInLineComment(commentText string) (int, int, bool) {
	idx := 0
	for idx < len(commentText) && commentText[idx] == '/' {
		idx++
	}
	for idx < len(commentText) && (commentText[idx] == ' ' || commentText[idx] == '\t') {
		idx++
	}
	if strings.HasPrefix(commentText[idx:], tsIgnoreDirective) {
		return idx, idx + len(tsIgnoreDirective), true
	}
	return 0, 0, false
}

func findDirectiveInBlockComment(commentText string) (int, int, bool) {
	if len(commentText) < 4 {
		return 0, 0, false
	}

	contentStart := 2
	contentEnd := len(commentText) - 2
	if contentEnd <= contentStart {
		return 0, 0, false
	}
	content := commentText[contentStart:contentEnd]

	lastLineStart := strings.LastIndexByte(content, '\n')
	if lastLineStart == -1 {
		lastLineStart = 0
	} else {
		lastLineStart++
	}

	line := content[lastLineStart:]
	idx := 0
	for idx < len(line) && (line[idx] == ' ' || line[idx] == '\t' || line[idx] == '\r') {
		idx++
	}
	for idx < len(line) && (line[idx] == '/' || line[idx] == '*') {
		idx++
	}
	for idx < len(line) && (line[idx] == ' ' || line[idx] == '\t') {
		idx++
	}
	if idx < len(line) && strings.HasPrefix(line[idx:], tsIgnoreDirective) {
		start := contentStart + lastLineStart + idx
		return start, start + len(tsIgnoreDirective), true
	}

	return 0, 0, false
}

func findTsIgnoreDirective(commentText string, kind ast.Kind) (int, int, bool) {
	if kind == ast.KindSingleLineCommentTrivia {
		return findDirectiveInLineComment(commentText)
	}
	if kind == ast.KindMultiLineCommentTrivia {
		return findDirectiveInBlockComment(commentText)
	}
	return 0, 0, false
}

var PreferTsExpectErrorRule = rule.CreateRule(rule.Rule{
	Name: "prefer-ts-expect-error",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		fullText := ctx.SourceFile.Text()
		utils.ForEachComment(ctx.SourceFile.AsNode(), func(comment *ast.CommentRange) {
			if comment == nil {
				return
			}
			commentText := fullText[comment.Pos():comment.End()]
			start, end, ok := findTsIgnoreDirective(commentText, comment.Kind)
			if !ok {
				return
			}

			fixRange := core.NewTextRange(comment.Pos()+start, comment.Pos()+end)
			fix := rule.RuleFixReplaceRange(fixRange, tsExpectErrorDirective)
			commentRange := core.NewTextRange(comment.Pos(), comment.End())
			ctx.ReportRangeWithFixes(commentRange, buildPreferExpectErrorMessage(), fix)
		}, ctx.SourceFile)

		return rule.RuleListeners{}
	},
})
