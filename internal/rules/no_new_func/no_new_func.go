package no_new_func

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

// https://eslint.org/docs/latest/rules/no-new-func

var msg = rule.RuleMessage{
	Id:          "noFunctionConstructor",
	Description: "The Function constructor is eval.",
}

var NoNewFuncRule = rule.Rule{
	Name:   "no-new-func",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysisReady := false
		globalFunctionHasDefinitions := false
		var localFunctionReferences map[*ast.Node]struct{}

		// isGlobalFunction checks whether an identifier resolves to ESLint's
		// declaration-less global Function variable rather than a file binding.
		isGlobalFunction := func(id *ast.Node) bool {
			// A config `/* global Function: off */` / `languageOptions.globals`
			// entry un-declares the builtin, so `Function` no longer resolves
			// to a known global — ESLint's `globalScope.set.get("Function")`
			// would be undefined and the rule stays silent.
			if !ctx.Globals.Access("Function").IsDeclared() {
				return false
			}
			if !analysisReady {
				analysisReady = true
				manager := scope.Build(ctx.SourceFile, scope.Options{
					CollectReferences: true,
					ReferenceNames:    map[string]struct{}{"Function": {}},
				})
				globalFunctionHasDefinitions = len(manager.Global.Declarations("Function")) != 0
				for _, reference := range manager.References {
					if reference.Resolved() != nil {
						if localFunctionReferences == nil {
							localFunctionReferences = make(map[*ast.Node]struct{})
						}
						localFunctionReferences[reference.Identifier] = struct{}{}
					}
				}
			}
			// In a module, authored declarations live in the module scope and do
			// not add definitions to ESLint's outer global Function variable.
			if ctx.LanguageOptions.EffectiveSourceType() != "module" && globalFunctionHasDefinitions {
				return false
			}
			_, local := localFunctionReferences[id]
			return !local
		}

		check := func(node *ast.Node) {
			var callee *ast.Node
			if node.Kind == ast.KindNewExpression {
				callee = node.AsNewExpression().Expression
			} else {
				callee = node.AsCallExpression().Expression
			}

			if callee == nil {
				return
			}

			// ESTree discards parentheses, but keeps TypeScript assertions as
			// separate expressions. Only the former are transparent here.
			unwrapped := ast.SkipParentheses(callee)

			// Case 1: new Function(...), Function(...), (Function)(...)
			if unwrapped.Kind == ast.KindIdentifier && unwrapped.AsIdentifier().Text == "Function" {
				if !isGlobalFunction(unwrapped) {
					return
				}
				ctx.ReportNode(node, msg)
				return
			}

			// Case 2: Function.call(...), Function.apply(...), Function.bind(...)
			// Only applies to CallExpression (not NewExpression)
			if node.Kind != ast.KindCallExpression {
				return
			}
			if !ast.IsAccessExpression(unwrapped) {
				return
			}

			if !utils.IsSpecificMemberAccess(unwrapped, "Function", "apply") &&
				!utils.IsSpecificMemberAccess(unwrapped, "Function", "bind") &&
				!utils.IsSpecificMemberAccess(unwrapped, "Function", "call") {
				return
			}

			obj := ast.SkipParentheses(utils.AccessExpressionObject(unwrapped))
			if obj == nil || obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != "Function" {
				return
			}

			if !isGlobalFunction(obj) {
				return
			}

			ctx.ReportNode(node, msg)
		}

		return rule.RuleListeners{
			ast.KindNewExpression:  check,
			ast.KindCallExpression: check,
		}
	},
}
