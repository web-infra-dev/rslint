package no_commented_out_tests

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const (
	testRootPattern            = `(test|it|describe)`
	testMemberPattern          = `(\s*(\.\s*\w+|\[\s*['"]\w+['"]\s*\]))`
	parameterizedMemberPattern = `(\.\s*(each|for)|\[\s*['"](each|for)['"]\s*\])`
	typeArgumentsPattern       = `(\s*<[^\x60\n]*>)?`
)

var commentedTestRegexps = []*regexp.Regexp{
	regexp.MustCompile(`(?m)^\s*` + testRootPattern + testMemberPattern + `*` + typeArgumentsPattern + `\s*\(`),
	regexp.MustCompile(`(?m)^\s*` + testRootPattern + testMemberPattern + `*` + `\s*` + parameterizedMemberPattern + typeArgumentsPattern + `\s*\x60`),
}

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

func hasCommentedOutTest(text string) bool {
	for _, re := range commentedTestRegexps {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

var NoCommentedOutTestsRule = rule.Rule{
	Name: "rstest/no-commented-out-tests",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		text := ctx.SourceFile.Text()
		for _, comment := range ctx.Comments.All() {
			if comment == nil {
				continue
			}
			inner := commentInnerText(text, comment)
			if inner == "" || !hasCommentedOutTest(inner) {
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
