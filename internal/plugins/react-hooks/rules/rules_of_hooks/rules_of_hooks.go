package rules_of_hooks

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message functions for different error types
func buildConditionalHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "conditionalHook",
		Description: `React Hook "` + hookName + `" is called conditionally. React Hooks must be ` +
			"called in the exact same order in every component render.",
	}
}

func buildLoopHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "loopHook",
		Description: `React Hook "` + hookName + `" may be executed more than once. Possibly ` +
			"because it is called in a loop. React Hooks must be called in the " +
			"exact same order in every component render.",
	}
}

func buildFunctionHookMessage(hookName, functionName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "functionHook",
		Description: `React Hook "` + hookName + `" is called in function "` + functionName + `" that is neither ` +
			"a React function component nor a custom React Hook function." +
			" React component names must start with an uppercase letter." +
			" React Hook names must start with the word \"use\".",
	}
}

func buildGenericHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "genericHook",
		Description: `React Hook "` + hookName + `" cannot be called inside a callback. React Hooks ` +
			"must be called in a React function component or a custom React " +
			"Hook function.",
	}
}

func buildTopLevelHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "topLevelHook",
		Description: `React Hook "` + hookName + `" cannot be called at the top level. React Hooks ` +
			"must be called in a React function component or a custom React " +
			"Hook function.",
	}
}

func buildClassHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "classHook",
		Description: `React Hook "` + hookName + `" cannot be called in a class component. React Hooks ` +
			"must be called in a React function component or a custom React " +
			"Hook function.",
	}
}

func buildAsyncComponentHookMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "asyncComponentHook",
		Description: `React Hook "` + hookName + `" cannot be called in an async function.`,
	}
}

func buildTryCatchUseMessage(hookName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tryCatchUse",
		Description: `React Hook "` + hookName + `" cannot be called inside a try/catch block.`,
	}
}

var RulesOfHooksRule = rule.Rule{
	Name: "react-hooks/rules-of-hooks",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return utils.VisitModules(func(source, node *ast.Node) {

		}, utils.VisitModulesOptions{
			Commonjs: true,
			ESModule: true,
		})
	},
}
