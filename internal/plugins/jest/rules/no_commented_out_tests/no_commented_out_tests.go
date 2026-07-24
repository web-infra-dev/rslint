package no_commented_out_tests

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Port of eslint-plugin-jest no-commented-out-tests:
// /^\s*[xf]?(test|it|describe)(\.\w+|\[['"]\w+['"]\])?\s*\(/mu
var commentedTestRegexp = regexp.MustCompile(`(?m)^\s*[xf]?(test|it|describe)(\.\w+|\[['"]\w+['"]\])?\s*\(`)

func buildCommentedTestsMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "commentedTests",
		Description: "Do not comment out tests",
	}
}

func commentInnerText(sourceText string, comment *ast.CommentRange) string {
	if comment == nil || comment.Pos() < 0 || comment.End() > len(sourceText) {
		return ""
	}
	switch comment.Kind {
	case ast.KindSingleLineCommentTrivia:
		start := comment.Pos() + 2 // past //
		if start >= comment.End() {
			return ""
		}
		return sourceText[start:comment.End()]
	case ast.KindMultiLineCommentTrivia:
		start := comment.Pos() + 2 // past /*
		end := comment.End() - 2   // before */
		if start >= end {
			return ""
		}
		return sourceText[start:end]
	default:
		return ""
	}
}

var NoCommentedOutTestsRule = rule.Rule{
	Name: "jest/no-commented-out-tests",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		text := ctx.SourceFile.Text()
		for _, comment := range ctx.Comments.All() {
			if comment == nil {
				continue
			}
			inner := commentInnerText(text, comment)
			if inner == "" || !commentedTestRegexp.MatchString(inner) {
				continue
			}
			ctx.ReportRange(
				core.NewTextRange(comment.Pos(), comment.End()),
				buildCommentedTestsMessage(),
			)
		}
		return rule.RuleListeners{}
	},
}
