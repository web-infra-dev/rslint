package consistent_json_file_read

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const utf8Literal = "'utf8'"

func insertEncoding(source *ast.SourceFile, read *ast.Node) []rule.RuleFix {
	call := read.AsCallExpression()
	path := call.Arguments.Nodes[0]
	if call.Arguments.HasTrailingComma() {
		comma, ok := utils.TokenAtOrAfter(source, path.End())
		if !ok || comma.Kind != ast.KindCommaToken {
			return nil
		}
		return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(comma.End, comma.End), " "+utf8Literal)}
	}
	return []rule.RuleFix{rule.RuleFixInsertAfter(path, ", "+utf8Literal)}
}

func replaceEncoding(ctx *rule.RuleContext, encoding *ast.Node) []rule.RuleFix {
	span := utils.TrimNodeTextRange(ctx.SourceFile, encoding)
	if utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
		return nil
	}
	return []rule.RuleFix{rule.RuleFixReplaceRange(span, utf8Literal)}
}

func removeEncoding(ctx *rule.RuleContext, args []*ast.Node) []rule.RuleFix {
	comma, ok := utils.TokenAtOrAfter(ctx.SourceFile, args[0].End())
	if !ok || comma.Kind != ast.KindCommaToken {
		return nil
	}
	span := core.NewTextRange(comma.Start, args[1].End())
	comments := utils.CommentsInSpan(ctx.Comments.All(), span.Pos(), span.End())
	if len(comments) == 0 {
		return []rule.RuleFix{rule.RuleFixRemoveRange(span)}
	}

	// Keep comments inside the removed argument and leave any trailing comma.
	var replacement strings.Builder
	text := ctx.SourceFile.Text()
	for _, comment := range comments {
		replacement.WriteByte(' ')
		replacement.WriteString(text[comment.Pos():comment.End()])
		if comment.Kind == ast.KindSingleLineCommentTrivia {
			replacement.WriteString(firstLineBreak(ctx.SourceFile))
		}
	}
	return []rule.RuleFix{rule.RuleFixReplaceRange(span, replacement.String())}
}

func firstLineBreak(source *ast.SourceFile) string {
	lines := source.ECMALineMap()
	if len(lines) < 2 {
		return "\n"
	}
	firstLine := source.Text()[:lines[1]]
	if strings.HasSuffix(firstLine, "\r\n") {
		return "\r\n"
	}
	_, width := utf8.DecodeLastRuneInString(firstLine)
	return firstLine[len(firstLine)-width:]
}
