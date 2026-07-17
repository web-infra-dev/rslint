package prefer_catch

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const skipTransparent = ast.OEKParentheses

func buildPreferCatchToThenMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferCatchToThen",
		Description: "Prefer `catch` to `then(a, b)`/`then(null, b)`.",
	}
}

var PreferCatchRule = rule.Rule{
	Name:   "promise/prefer-catch",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				callee := ast.SkipOuterExpressions(callExpr.Expression, skipTransparent)
				if callee == nil || !ast.IsPropertyAccessExpression(callee) {
					return
				}
				prop := callee.AsPropertyAccessExpression()
				thenName := prop.Name()
				if thenName == nil || !ast.IsIdentifier(thenName) || thenName.AsIdentifier().Text != "then" {
					return
				}

				args := node.Arguments()
				if len(args) < 2 {
					return
				}

				sf := ctx.SourceFile
				msg := buildPreferCatchToThenMessage()

				firstArg := ast.SkipOuterExpressions(args[0], skipTransparent)
				isNullOrUndef := firstArg.Kind == ast.KindNullKeyword ||
					(ast.IsIdentifier(firstArg) && firstArg.AsIdentifier().Text == "undefined")

				if isNullOrUndef {
					// hey.then(null, fn2) → hey.catch(fn2)
					// Remove from start of arg0 to start of arg1 (removes "null, " with trailing whitespace)
					removeStart := utils.TrimNodeTextRange(sf, args[0]).Pos()
					removeEnd := utils.TrimNodeTextRange(sf, args[1]).Pos()
					ctx.ReportNodeWithFixes(thenName, msg,
						rule.RuleFixRemoveRange(core.NewTextRange(removeStart, removeEnd)),
						rule.RuleFixReplace(sf, thenName, "catch"),
					)
				} else {
					// hey.then(fn1, fn2) → hey.catch(fn2).then(fn1)
					// Get text of arg1 stripping outer parens to match ESLint's getText behavior
					innerArg1 := ast.SkipOuterExpressions(args[1], skipTransparent)
					catcherText := utils.TrimmedNodeText(sf, innerArg1)
					// Remove from the comma after arg0 to end of arg1, so trivia
					// between arg0 and the comma (e.g. a trailing comment) is kept.
					commaStart := scanner.GetRangeOfTokenAtPosition(sf, args[0].End()).Pos()
					removeRange := core.NewTextRange(commaStart, args[1].End())
					ctx.ReportNodeWithFixes(thenName, msg,
						rule.RuleFixRemoveRange(removeRange),
						rule.RuleFixInsertBefore(sf, thenName, "catch("+catcherText+")."),
					)
				}
			},
		}
	},
}
