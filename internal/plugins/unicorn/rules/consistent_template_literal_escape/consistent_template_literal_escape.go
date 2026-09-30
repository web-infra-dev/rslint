package consistent_template_literal_escape

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
			if utils.IsTaggedTemplateElement(node) {
				return
			}

			raw, contentRange := utils.TemplateElementRaw(ctx.SourceFile, node)
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

func normalizeTemplateEscape(raw string) (string, bool) {
	if !strings.Contains(raw, "$\\{") {
		return raw, false
	}

	output := make([]byte, 0, len(raw))
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
			continue
		}

		output = append(output, raw[index])
		index++
	}

	return string(output), true
}
