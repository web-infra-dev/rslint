package no_done_callback

import (
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message Builders

func buildErrorNoDoneCallbackMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noDoneCallback",
		Description: "Return a Promise instead of relying on callback parameter",
	}
}

func buildErrorSuggestWrappingInPromiseMessage(callbackName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestWrappingInPromise",
		Description: fmt.Sprintf("Wrap in `new Promise(%s => ...`", callbackName),
	}
}

func buildErrorUseAwaitInsteadOfCallbackMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useAwaitInsteadOfCallback",
		Description: "Use await instead of callback in async functions",
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func findCallbackArgument(callExpr *ast.CallExpression, jestFnCall *utils.ParsedJestFnCall, isJestEach bool) *ast.Node {
	nodes := callExpr.Arguments.Nodes
	argLength := len(nodes)
	if argLength == 0 {
		return nil
	}

	if (isJestEach || jestFnCall.Kind == utils.JestFnTypeTest) && argLength >= 2 {
		return nodes[1]
	} else if jestFnCall.Kind == utils.JestFnTypeHook && argLength >= 1 {
		return nodes[0]
	}

	return nil
}

func isFunction(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionDeclaration ||
		node.Kind == ast.KindFunctionExpression ||
		node.Kind == ast.KindArrowFunction
}

var NoDoneCallbackRule = rule.Rule{
	Name: "jest/no-done-callback",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil {
					return
				}

				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil {
					return
				}

				callee := callExpr.Expression
				if callee == nil {
					return
				}

				isJestEach := slices.Contains(jestFnCall.Members, "each")
				if isJestEach && callee.Kind != ast.KindTaggedTemplateExpression {
					return
				}

				callback := findCallbackArgument(callExpr, jestFnCall, isJestEach)
				callbackArgIndex := boolToInt(isJestEach)
				if callback == nil ||
					!isFunction(callback) ||
					len(callback.Parameters()) == 0 ||
					len(callback.Parameters()) != 1+callbackArgIndex {
					return
				}

				params := callback.Parameters()
				argument := params[callbackArgIndex]
				paramDecl := argument.AsParameterDeclaration()
				nameNode := paramDecl.Name()
				if paramDecl.DotDotDotToken != nil ||
					nameNode == nil ||
					nameNode.Kind != ast.KindIdentifier {
					ctx.ReportNode(argument, buildErrorNoDoneCallbackMessage())
					return
				}

				if ast.IsAsyncFunction(callback) {
					ctx.ReportNode(argument, buildErrorUseAwaitInsteadOfCallbackMessage())
					return
				}

				body := callback.Body()
				if body == nil {
					ctx.ReportNode(argument, buildErrorNoDoneCallbackMessage())
					return
				}

				text := ctx.SourceFile.Text()
				// Parser-recorded ParameterList.Loc spans the inside of the
				// `(...)` when the parameters are parenthesized; for
				// paren-less simple arrows (`x => ...`) it equals the lone
				// parameter's own Loc. So `text[Loc.End()] == ')'` is enough
				// to detect parens — no need to scan trivia ourselves.
				paramListLoc := callback.FunctionLikeData().Parameters.Loc
				hasParens := paramListLoc.End() < len(text) && text[paramListLoc.End()] == ')'

				bodyStart := scanner.GetTokenPosOfNode(body, ctx.SourceFile, false)
				bodyEnd := body.End()
				bodyIsBlock := body.Kind == ast.KindBlock
				callbackName := nameNode.Text()

				var fixes []rule.RuleFix
				if hasParens {
					fixes = append(fixes, rule.RuleFixRemoveRange(paramListLoc))
				} else {
					firstParam := params[0]
					firstParamStart := scanner.GetTokenPosOfNode(firstParam, ctx.SourceFile, false)
					fixes = append(fixes, rule.RuleFixReplaceRange(
						core.NewTextRange(firstParamStart, firstParam.End()),
						"()",
					))
				}

				beforeReplacement := "new Promise(" + callbackName + " => "
				afterReplacement := ")"
				if bodyIsBlock {
					beforeReplacement = "return " + beforeReplacement + "{"
					afterReplacement += "}"
					fixes = append(fixes, rule.RuleFixReplaceRange(
						core.NewTextRange(bodyStart+1, bodyStart+1),
						beforeReplacement,
					))
				} else {
					fixes = append(fixes, rule.RuleFixReplaceRange(
						core.NewTextRange(bodyStart, bodyStart),
						beforeReplacement,
					))
				}
				fixes = append(fixes, rule.RuleFixReplaceRange(
					core.NewTextRange(bodyEnd, bodyEnd),
					afterReplacement,
				))

				ctx.ReportNodeWithSuggestions(
					argument,
					buildErrorNoDoneCallbackMessage(),
					rule.RuleSuggestion{
						Message:  buildErrorSuggestWrappingInPromiseMessage(callbackName),
						FixesArr: fixes,
					},
				)
			},
		}
	},
}
