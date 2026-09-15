package default_param_last

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var shouldBeLastMessage = rule.RuleMessage{
	Id:          "shouldBeLast",
	Description: "Default parameters should be last.",
}

// DefaultParamLastRule enforces default parameters to be last.
var DefaultParamLastRule = rule.Rule{
	Name:   "default-param-last",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			if node.Body() == nil {
				return
			}
			params := utils.ESTreeParameters(node)
			isRequiredParameter := func(current *ast.Node) bool {
				if current == nil {
					return false
				}
				param := current.AsParameterDeclaration()
				questionToken := param.QuestionToken
				if utils.IsJSDocSyntaxNode(questionToken) {
					questionToken = nil
				}
				return param.Initializer == nil &&
					param.DotDotDotToken == nil &&
					questionToken == nil
			}

			lastRequired := -1
			for i := len(params) - 1; i >= 0; i-- {
				if isRequiredParameter(params[i]) {
					lastRequired = i
					break
				}
			}
			for i := range lastRequired {
				if !isRequiredParameter(params[i]) {
					reportRange := utils.NodeTextRangeSkippingDecorators(ctx.SourceFile, params[i])
					if ast.IsParameterPropertyDeclaration(params[i], node) {
						reportRange = utils.TrimNodeTextRange(ctx.SourceFile, params[i])
					}
					ctx.ReportRange(reportRange, shouldBeLastMessage)
				}
			}
		}

		// ESLint's FunctionExpression listener also reaches methods and
		// constructors through ESTree wrapper nodes. tsgo models them directly.
		return rule.RuleListeners{
			ast.KindFunctionDeclaration: check,
			ast.KindFunctionExpression:  check,
			ast.KindArrowFunction:       check,
			ast.KindMethodDeclaration:   check,
			ast.KindConstructor:         check,
			ast.KindGetAccessor:         check,
			ast.KindSetAccessor:         check,
		}
	},
}
