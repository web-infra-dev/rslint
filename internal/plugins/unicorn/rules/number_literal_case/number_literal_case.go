package number_literal_case

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed number_literal_case.schema.json
var schemaJSON []byte

var message = rule.RuleMessage{
	Id:          "number-literal-case",
	Description: "Invalid number literal casing.",
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/number-literal-case.js
var NumberLiteralCaseRule = rule.Rule{
	Name:   "unicorn/number-literal-case",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		lowercaseHexadecimal := false
		if len(options) > 0 {
			if object, ok := options[0].(map[string]any); ok {
				lowercaseHexadecimal = object["hexadecimalValue"] == "lowercase"
			}
		}

		checkLiteral := func(node *ast.Node) {
			// NumericLiteral.Text is normalized by tsgo; casing and separators
			// must come from the original token instead.
			raw := utils.TrimmedNodeText(ctx.SourceFile, node)
			value := raw
			if node.Kind == ast.KindBigIntLiteral {
				value = strings.TrimSuffix(value, "n")
			}
			fixed := ecmascript.StringToLowerCase(value)
			if strings.HasPrefix(fixed, "0x") && !lowercaseHexadecimal {
				fixed = "0x" + ecmascript.StringToUpperCase(fixed[2:])
			}
			if node.Kind == ast.KindBigIntLiteral {
				fixed += "n"
			}
			if raw == fixed {
				return
			}

			ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
				return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, fixed)}
			})
		}
		return rule.RuleListeners{
			ast.KindNumericLiteral: checkLiteral,
			ast.KindBigIntLiteral:  checkLiteral,
		}
	},
}
