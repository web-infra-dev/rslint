package default_param_last

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// DefaultParamLastRule enforces default parameters to be last
var DefaultParamLastRule = rule.CreateRule(rule.Rule{
	Name: "default-param-last",
	Run:  run,
})

var shouldBeLastMessage = rule.RuleMessage{
	Id:          "shouldBeLast",
	Description: "Default parameters should be last.",
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	checkDefaultParamLast := func(node *ast.Node) {
		params := node.Parameters()
		if len(params) < 2 {
			return
		}

		hasSeenPlainParam := false
		for i := len(params) - 1; i >= 0; i-- {
			current := params[i]
			if current == nil {
				continue
			}

			param := current.AsParameterDeclaration()
			if param.DotDotDotToken != nil {
				continue
			}
			if param.Initializer == nil && param.QuestionToken == nil {
				hasSeenPlainParam = true
				continue
			}
			if hasSeenPlainParam {
				ctx.ReportNode(current, shouldBeLastMessage)
			}
		}
	}

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: checkDefaultParamLast,
		ast.KindFunctionExpression:  checkDefaultParamLast,
		ast.KindArrowFunction:       checkDefaultParamLast,
		ast.KindMethodDeclaration:   checkDefaultParamLast,
		ast.KindConstructor:         checkDefaultParamLast,
	}
}
