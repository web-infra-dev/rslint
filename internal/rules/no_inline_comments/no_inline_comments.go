package no_inline_comments

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed no_inline_comments.schema.json
var schemaJSON []byte

var unexpectedInlineCommentMessage = rule.RuleMessage{
	Id:          "unexpectedInlineComment",
	Description: "Unexpected comment inline with code.",
}

// eslintDirectivePattern ports astUtils.ESLINT_DIRECTIVE_PATTERN, used to
// recognize block-comment directives (`/*eslint ...*/`, `/*global ...*/`,
// `/*exported ...*/`) by their leading text.
var eslintDirectivePattern = esregexp.MustCompile(`^(?:eslint[- ]|(?:globals?|exported) )`, "u")

type options struct {
	ignorePattern *esregexp.RegExp
}

func parseOptions(raw []any) options {
	opts := options{}
	if len(raw) == 0 {
		return opts
	}
	m, _ := raw[0].(map[string]any)
	if pattern, ok := m["ignorePattern"].(string); ok && pattern != "" {
		if re, err := esregexp.Compile(pattern, "u"); err == nil {
			opts.ignorePattern = re
		}
	}
	return opts
}

// commentValue extracts the text between a comment's delimiters — the same
// substring ESLint exposes as `comment.value` — without any trimming.
func commentValue(text string, comment *ast.CommentRange) string {
	switch comment.Kind {
	case ast.KindSingleLineCommentTrivia:
		return text[comment.Pos()+2 : comment.End()]
	case ast.KindMultiLineCommentTrivia:
		// A block comment left unterminated at end of file still parses, and
		// then has no closing delimiter to strip.
		end := comment.End()
		if end-comment.Pos() >= 4 && text[end-2:end] == "*/" {
			end -= 2
		}
		return text[comment.Pos()+2 : end]
	default:
		return ""
	}
}

// isDirectiveComment ports astUtils.isDirectiveComment: a Line comment is a
// directive if its trimmed text starts with "eslint-"; a Block comment is a
// directive if its trimmed text matches eslintDirectivePattern. Both forms
// also accept the "rslint-" prefix this linter recognizes alongside "eslint-".
func isDirectiveComment(kind ast.Kind, trimmedValue string) bool {
	isLintDirective := strings.HasPrefix(trimmedValue, "eslint-") ||
		strings.HasPrefix(trimmedValue, "rslint-")

	switch kind {
	case ast.KindSingleLineCommentTrivia:
		return isLintDirective
	case ast.KindMultiLineCommentTrivia:
		return isLintDirective || eslintDirectivePattern.Test(trimmedValue)
	default:
		return false
	}
}

// NoInlineCommentsRule disallows comments placed on the same line as code on
// either side of them.
//
// https://eslint.org/docs/latest/rules/no-inline-comments
var NoInlineCommentsRule = rule.Rule{
	Name:   "no-inline-comments",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		sourceFile := ctx.SourceFile
		text := sourceFile.Text()
		lineStarts := scanner.GetECMALineStarts(sourceFile)

		lineText := func(line int) string {
			start := int(lineStarts[line])
			end := len(text)
			if line+1 < len(lineStarts) {
				end = int(lineStarts[line+1])
			}
			return strings.TrimRight(text[start:end], "\r\n")
		}

		check := func(comment *ast.CommentRange) {
			startLine := scanner.ComputeLineOfPosition(lineStarts, comment.Pos())
			endLine := scanner.ComputeLineOfPosition(lineStarts, comment.End())
			startCol := comment.Pos() - int(lineStarts[startLine])
			endCol := comment.End() - int(lineStarts[endLine])

			startLineText := lineText(startLine)
			endLineText := lineText(endLine)
			if startCol > len(startLineText) {
				startCol = len(startLineText)
			}
			if endCol > len(endLineText) {
				endCol = len(endLineText)
			}

			preamble := ecmascript.StringTrim(startLineText[:startCol])
			postamble := ecmascript.StringTrim(endLineText[endCol:])
			isPreambleEmpty := preamble == ""
			isPostambleEmpty := postamble == ""

			// Nothing on both sides.
			if isPreambleEmpty && isPostambleEmpty {
				return
			}

			value := commentValue(text, comment)

			// Matches the ignore pattern.
			if opts.ignorePattern != nil && opts.ignorePattern.TestOrTimeout(value) {
				return
			}

			// JSX exception: a comment that is the sole content of a JSX
			// expression container (`{/* comment */}`, `{// comment\n}`) is
			// not "inline with code" — there is no code, only the braces.
			if (isPreambleEmpty || preamble == "{") && (isPostambleEmpty || postamble == "}") {
				enclosing := ast.GetNodeAtPosition(sourceFile, comment.Pos(), false)
				if enclosing != nil && enclosing.Kind == ast.KindJsxExpression &&
					enclosing.AsJsxExpression().Expression == nil {
					return
				}
			}

			// Don't report ESLint directive comments.
			if isDirectiveComment(comment.Kind, ecmascript.StringTrim(value)) {
				return
			}

			ctx.ReportRange(comment.TextRange, unexpectedInlineCommentMessage)
		}

		for _, comment := range ctx.Comments.All() {
			check(comment)
		}

		return rule.RuleListeners{}
	},
}
