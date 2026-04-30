package prefer_to_be

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message Builders

func buildUseToBeErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToBe",
		Description: "Use `toBe` when expecting primitive literals",
	}
}

func buildUseToBeUndefinedErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToBeUndefined",
		Description: "Use `toBeUndefined` instead",
	}
}

func buildUseToBeDefinedErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToBeDefined",
		Description: "Use `toBeDefined` instead",
	}
}

func buildUseToBeNullErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToBeNull",
		Description: "Use `toBeNull` instead",
	}
}

func buildUseToBeNaNErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToBeNaN",
		Description: "Use `toBeNaN` instead",
	}
}

var PreferToBeRule = rule.Rule{
	Name: "jest/prefer-to-be",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil || jestFnCall.Kind != utils.JestFnTypeHook {
					return
				}
			},
		}
	},
}
