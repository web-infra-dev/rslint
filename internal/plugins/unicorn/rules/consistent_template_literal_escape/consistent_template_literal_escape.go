package consistent_template_literal_escape

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "consistent-template-literal-escape"

var message = rule.RuleMessage{
	Id:          messageID,
	Description: "Use `\\${` instead of `$\\{` to escape in template literals.",
}

var ConsistentTemplateLiteralEscapeRule = rule.Rule{
	Name:   "unicorn/consistent-template-literal-escape",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			if isInsideTaggedTemplate(node) {
				return
			}

			raw, contentRange, ok := templateElementRaw(ctx.SourceFile, node)
			if !ok {
				return
			}
			fixed, changed := normalizeTemplateEscape(raw)
			if !changed {
				return
			}

			ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
				return []rule.RuleFix{rule.RuleFixReplaceRange(contentRange, fixed)}
			})
		}

		return rule.RuleListeners{
			ast.KindNoSubstitutionTemplateLiteral: check,
			ast.KindTemplateHead:                  check,
			ast.KindTemplateMiddle:                check,
			ast.KindTemplateTail:                  check,
		}
	},
}

func templateElementRaw(sourceFile *ast.SourceFile, node *ast.Node) (string, core.TextRange, bool) {
	if sourceFile == nil || node == nil {
		return "", core.TextRange{}, false
	}

	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	startOffset := 1
	endOffset := 1
	switch node.Kind {
	case ast.KindTemplateHead, ast.KindTemplateMiddle:
		endOffset = 2
	case ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateTail:
	default:
		return "", core.TextRange{}, false
	}

	start := nodeRange.Pos() + startOffset
	end := nodeRange.End() - endOffset
	if start > end {
		return "", core.TextRange{}, false
	}

	contentRange := core.NewTextRange(start, end)
	return sourceFile.Text()[start:end], contentRange, true
}

func normalizeTemplateEscape(raw string) (string, bool) {
	if !strings.Contains(raw, "$\\{") {
		return raw, false
	}

	output := make([]byte, 0, len(raw))
	changed := false

	for index := 0; index < len(raw); {
		if index+2 < len(raw) && raw[index] == '$' && raw[index+1] == '\\' && raw[index+2] == '{' {
			backslashes := 0
			for previous := index - 1; previous >= 0 && raw[previous] == '\\'; previous-- {
				backslashes++
			}
			if backslashes%2 == 1 && len(output) > 0 {
				output = output[:len(output)-1]
			}
			output = append(output, '\\', '$', '{')
			index += 3
			changed = true
			continue
		}

		output = append(output, raw[index])
		index++
	}

	if !changed {
		return raw, false
	}
	return string(output), true
}

func isInsideTaggedTemplate(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression
	case ast.KindTemplateHead:
		return node.Parent != nil &&
			node.Parent.Parent != nil &&
			node.Parent.Parent.Kind == ast.KindTaggedTemplateExpression
	case ast.KindTemplateMiddle, ast.KindTemplateTail:
		return node.Parent != nil &&
			node.Parent.Parent != nil &&
			node.Parent.Parent.Parent != nil &&
			node.Parent.Parent.Parent.Kind == ast.KindTaggedTemplateExpression
	default:
		return false
	}
}
