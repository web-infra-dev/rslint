package jsx_curly_newline

import (
	_ "embed"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed jsx_curly_newline.schema.json
var schemaJSON []byte

const (
	consistent = "consistent"
	require    = "require"
	forbid     = "forbid"
)

type options struct {
	singleline string
	multiline  string
}

func parseOptions(raw []any) options {
	result := options{singleline: consistent, multiline: consistent}
	if len(raw) == 0 {
		return result
	}
	if value, ok := raw[0].(string); ok && value == "never" {
		return options{singleline: forbid, multiline: forbid}
	}
	if value, ok := raw[0].(map[string]any); ok {
		if singleline, ok := value["singleline"].(string); ok {
			result.singleline = singleline
		}
		if multiline, ok := value["multiline"].(string); ok {
			result.multiline = multiline
		}
	}
	return result
}

func requiresNewlines(option options, start, end int, lineMap []core.TextPos, hasLeftNewline bool) bool {
	multiline := start != end && scanner.ComputeLineOfPosition(lineMap, start) != scanner.ComputeLineOfPosition(lineMap, end)
	mode := option.singleline
	if multiline {
		mode = option.multiline
	}
	switch mode {
	case forbid:
		return false
	case require:
		return true
	default:
		return hasLeftNewline
	}
}

// JsxCurlyNewlineRule enforces matching line breaks immediately within JSX
// expression-container braces. JSX spread children are excluded because the
// upstream listener is JSXExpressionContainer-only.
var JsxCurlyNewlineRule = rule.Rule{
	Name:   "react/jsx-curly-newline",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		option := parseOptions(rawOptions)
		text := ctx.SourceFile.Text()
		lineMap := ctx.SourceFile.ECMALineMap()
		comments := ctx.Comments.All()

		return rule.RuleListeners{
			ast.KindJsxExpression: func(node *ast.Node) {
				expression := node.AsJsxExpression()
				if expression == nil || expression.DotDotDotToken != nil {
					return
				}
				trimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
				openPos, closePos := trimmed.Pos(), trimmed.End()-1
				if openPos < 0 || closePos <= openPos || closePos >= len(text) || text[openPos] != '{' || text[closePos] != '}' {
					return
				}

				innerLow, innerHigh := openPos+1, closePos
				// ESLint's token APIs exclude comments. Use the parser-backed comment
				// ranges rather than scanning text, which would mistake comment-like
				// content in strings, regexps, and templates for trivia.
				firstCode := skipLeadingComments(text, innerLow, innerHigh, comments)
				lastCodeEnd := skipTrailingComments(text, innerLow, innerHigh, comments)
				hasLeftNewline := ecmascript.ContainsLineTerminator(text, innerLow, firstCode)
				hasRightNewline := ecmascript.ContainsLineTerminator(text, lastCodeEnd, innerHigh)
				// ESTree expression ranges exclude brace-adjacent trivia and tsgo-only
				// ParenthesizedExpression wrappers. Empty JSX expressions use their
				// authored interior.
				spanStart, spanEnd := firstCode, lastCodeEnd
				if expression.Expression != nil {
					estreeExpression := ast.SkipParentheses(expression.Expression)
					estreeRange := utils.TrimNodeTextRange(ctx.SourceFile, estreeExpression)
					spanStart, spanEnd = estreeRange.Pos(), estreeRange.End()
				}
				needsNewlines := requiresNewlines(option, spanStart, spanEnd, lineMap, hasLeftNewline)

				report := func(position int, id, description string, fix *rule.RuleFix) {
					range_ := core.NewTextRange(position, position+1)
					message := rule.RuleMessage{Id: id, Description: description}
					if fix == nil {
						ctx.ReportRange(range_, message)
					} else {
						ctx.ReportRangeWithFixes(range_, message, *fix)
					}
				}

				if hasLeftNewline && !needsNewlines {
					var fix *rule.RuleFix
					if ecmascript.StringTrim(text[innerLow:firstCode]) == "" {
						candidate := rule.RuleFixRemoveRange(core.NewTextRange(innerLow, firstCode))
						fix = &candidate
					}
					report(openPos, "unexpectedAfter", "Unexpected newline after '{'.", fix)
				} else if !hasLeftNewline && needsNewlines {
					fix := rule.RuleFix{Text: "\n", Range: core.NewTextRange(innerLow, innerLow)}
					report(openPos, "expectedAfter", "Expected newline after '{'.", &fix)
				}

				if hasRightNewline && !needsNewlines {
					var fix *rule.RuleFix
					if ecmascript.StringTrim(text[lastCodeEnd:innerHigh]) == "" {
						candidate := rule.RuleFixRemoveRange(core.NewTextRange(lastCodeEnd, innerHigh))
						fix = &candidate
					}
					report(closePos, "unexpectedBefore", "Unexpected newline before '}'.", fix)
				} else if !hasRightNewline && needsNewlines {
					fix := rule.RuleFix{Text: "\n", Range: core.NewTextRange(innerHigh, innerHigh)}
					report(closePos, "expectedBefore", "Expected newline before '}'.", &fix)
				}
			},
		}
	},
}

func skipLeadingComments(text string, start, end int, comments []*ast.CommentRange) int {
	position := start
	index := sort.Search(len(comments), func(index int) bool {
		return comments[index].Pos() >= position
	})
	for ; index < len(comments); index++ {
		comment := comments[index]
		if comment.End() > end || !ecmascript.IsBlank(text[position:comment.Pos()]) {
			break
		}
		position = comment.End()
	}
	return ecmascript.SkipLeadingWhitespace(text, position, end)
}

func skipTrailingComments(text string, start, end int, comments []*ast.CommentRange) int {
	position := end
	index := sort.Search(len(comments), func(index int) bool {
		return comments[index].End() > position
	}) - 1
	for ; index >= 0; index-- {
		comment := comments[index]
		if comment.Pos() < start || !ecmascript.IsBlank(text[comment.End():position]) {
			break
		}
		position = comment.Pos()
	}
	return ecmascript.SkipTrailingWhitespace(text, start, position)
}
