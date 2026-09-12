package jsx_curly_newline

import (
	_ "embed"
	"strings"

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

func requiresNewlines(option options, expression *ast.Node, lineMap []core.TextPos, hasLeftNewline bool) bool {
	multiline := false
	if expression != nil {
		multiline = expression.Pos() != expression.End() && scanner.ComputeLineOfPosition(lineMap, expression.Pos()) != scanner.ComputeLineOfPosition(lineMap, expression.End())
	}
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
				firstToken := ecmascript.SkipLeadingWhitespace(text, innerLow, innerHigh)
				lastTokenEnd := ecmascript.SkipTrailingWhitespace(text, innerLow, innerHigh)
				// The upstream token APIs skip comments. scan to the first/last
				// non-comment token so comments participate in newline detection.
				firstCode := skipLeadingComments(text, firstToken, innerHigh)
				lastCodeEnd := skipTrailingComments(text, lastTokenEnd, innerLow)
				hasLeftNewline := ecmascript.ContainsLineTerminator(text, innerLow, firstCode)
				hasRightNewline := ecmascript.ContainsLineTerminator(text, lastCodeEnd, innerHigh)
				needsNewlines := requiresNewlines(option, expression.Expression, lineMap, hasLeftNewline)

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

func skipLeadingComments(text string, position, end int) int {
	for position+1 < end && text[position] == '/' {
		switch text[position+1] {
		case '*':
			position += 2
			for position+1 < end && (text[position] != '*' || text[position+1] != '/') {
				position++
			}
			if position+1 >= end {
				return end
			}
			position += 2
		case '/':
			position += 2
			for position < end && text[position] != '\n' && text[position] != '\r' {
				position++
			}
		default:
			return position
		}
		position = ecmascript.SkipLeadingWhitespace(text, position, end)
	}
	return position
}

func skipTrailingComments(text string, position, start int) int {
	for position > start {
		lineStart := strings.LastIndexByte(text[start:position], '\n') + start + 1
		if comment := strings.LastIndex(text[lineStart:position], "//"); comment >= 0 {
			position = lineStart + comment
			position = ecmascript.SkipTrailingWhitespace(text, start, position)
			continue
		}
		if position-2 < start || text[position-2] != '*' || text[position-1] != '/' {
			break
		}
		position -= 2
		for position-2 >= start && (text[position-2] != '/' || text[position-1] != '*') {
			position--
		}
		if position-2 < start {
			return start
		}
		position -= 2
		position = ecmascript.SkipTrailingWhitespace(text, start, position)
	}
	return position
}
