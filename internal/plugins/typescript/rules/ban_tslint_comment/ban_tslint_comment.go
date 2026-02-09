package ban_tslint_comment

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// tslint regex
// https://github.com/palantir/tslint/blob/95d9d958833fd9dc0002d18cbe34db20d0fbf437/src/enableDisableRules.ts#L32
var enableDisableRegex = regexp.MustCompile(`^\s*tslint:(enable|disable)(?:-(line|next-line))?(:|\s|$)`)

// BanTslintCommentRule implements the ban-tslint-comment rule.
// Disallows tslint directive comments like // tslint:disable
var BanTslintCommentRule = rule.CreateRule(rule.Rule{
	Name: "ban-tslint-comment",
	Run:  run,
})

func run(ctx rule.RuleContext, _ any) rule.RuleListeners {
	text := ctx.SourceFile.Text()

	utils.ForEachComment(&ctx.SourceFile.Node, func(comment *ast.CommentRange) {
		commentValue := extractCommentValue(text, comment)
		if !enableDisableRegex.MatchString(commentValue) {
			return
		}

		formatted := formatCommentText(commentValue, comment.Kind == ast.KindSingleLineCommentTrivia)
		message := rule.RuleMessage{
			Id:          "commentDetected",
			Description: fmt.Sprintf("tslint comment detected: \"%s\"", formatted),
		}

		fixRange := buildFixRange(ctx.SourceFile, comment, len(text))
		ctx.ReportRangeWithFixes(
			core.NewTextRange(comment.Pos(), comment.End()),
			message,
			rule.RuleFixRemoveRange(fixRange),
		)
	}, ctx.SourceFile)

	return rule.RuleListeners{}
}

func extractCommentValue(text string, comment *ast.CommentRange) string {
	switch comment.Kind {
	case ast.KindSingleLineCommentTrivia:
		if comment.End() <= comment.Pos()+2 {
			return ""
		}
		return text[comment.Pos()+2 : comment.End()]
	case ast.KindMultiLineCommentTrivia:
		if comment.End() <= comment.Pos()+4 {
			return ""
		}
		return text[comment.Pos()+2 : comment.End()-2]
	default:
		return ""
	}
}

func formatCommentText(value string, isLine bool) string {
	trimmed := strings.TrimSpace(value)
	if isLine {
		return "// " + trimmed
	}
	return "/* " + trimmed + " */"
}

func buildFixRange(sourceFile *ast.SourceFile, comment *ast.CommentRange, textLen int) core.TextRange {
	text := sourceFile.Text()
	start := comment.Pos()
	end := comment.End()

	startLine, _ := scanner.GetECMALineAndCharacterOfPosition(sourceFile, start)
	lineStart := scanner.GetECMAPositionOfLineAndCharacter(sourceFile, startLine, 0)

	isStandalone := true
	for i := lineStart; i < start; i++ {
		if text[i] != ' ' && text[i] != '\t' {
			isStandalone = false
			break
		}
	}

	if isStandalone {
		start = lineStart
		if end < textLen {
			if text[end] == '\r' {
				end++
				if end < textLen && text[end] == '\n' {
					end++
				}
			} else if text[end] == '\n' {
				end++
			}
		}
	} else {
		for start > lineStart && (text[start-1] == ' ' || text[start-1] == '\t') {
			start--
		}
	}

	return core.NewTextRange(start, end)
}
