package no_confusing_non_null_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildConfusingAssignMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "confusingAssign",
		Description: "Confusing combination of non-null assertion and assignment like `a! = b`, which looks very similar to `a != b`.",
	}
}

func buildConfusingEqualMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "confusingEqual",
		Description: "Confusing combination of non-null assertion and equality test like `a! == b`, which looks very similar to `a !== b`.",
	}
}

func buildConfusingOperatorMessage(op string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "confusingOperator",
		Description: "Confusing combination of non-null assertion and `" + op + "` operator like `a! " + op + " b`, which might be misinterpreted as `!(a " + op + " b)`.",
		Data:        map[string]string{"operator": op},
	}
}

func buildNotNeedInAssignMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notNeedInAssign",
		Description: "Remove unnecessary non-null assertion (!) in assignment left-hand side.",
	}
}

func buildNotNeedInEqualTestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notNeedInEqualTest",
		Description: "Remove unnecessary non-null assertion (!) in equality test.",
	}
}

func buildNotNeedInOperatorMessage(op string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notNeedInOperator",
		Description: "Remove possibly unnecessary non-null assertion (!) in the left operand of the `" + op + "` operator.",
		Data:        map[string]string{"operator": op},
	}
}

func buildWrapUpLeftMessage(op string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "wrapUpLeft",
		Description: `Wrap the left-hand side in parentheses to avoid confusion with "` + op + `" operator.`,
		Data:        map[string]string{"operator": op},
	}
}

var NoConfusingNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-confusing-non-null-assertion",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin.OperatorToken == nil {
					return
				}

				var operator string
				var primaryMsg rule.RuleMessage
				switch bin.OperatorToken.Kind {
				case ast.KindEqualsToken:
					operator = "="
					primaryMsg = buildConfusingAssignMessage()
				case ast.KindEqualsEqualsToken:
					operator = "=="
					primaryMsg = buildConfusingEqualMessage()
				case ast.KindEqualsEqualsEqualsToken:
					operator = "==="
					primaryMsg = buildConfusingEqualMessage()
				case ast.KindInKeyword:
					operator = "in"
					primaryMsg = buildConfusingOperatorMessage(operator)
				case ast.KindInstanceOfKeyword:
					operator = "instanceof"
					primaryMsg = buildConfusingOperatorMessage(operator)
				default:
					return
				}

				left := bin.Left
				if left == nil {
					return
				}

				// Mirror upstream's `getLastToken(node.left).value === '!'` +
				// `getTokenAfter(node.left).value !== ')'` check. In tsgo,
				// ParenthesizedExpression is an explicit AST node, so when the
				// left side is wrapped in parens its End() lands right after
				// the closing `)`. Probing the byte just before End() collapses
				// both conditions: the byte is `!` iff the unparenthesized
				// left ends in a non-null assertion.
				text := ctx.SourceFile.Text()
				leftEnd := left.End()
				if leftEnd < 1 || leftEnd > len(text) || text[leftEnd-1] != '!' {
					return
				}

				wrapUpLeftFix := []rule.RuleFix{
					rule.RuleFixInsertBefore(ctx.SourceFile, left, "("),
					rule.RuleFixInsertAfter(left, ")"),
				}

				if left.Kind == ast.KindNonNullExpression {
					expression := left.AsNonNullExpression().Expression
					s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
					removeBangFix := []rule.RuleFix{rule.RuleFixRemoveRange(s.TokenRange())}

					var suggestions []rule.RuleSuggestion
					switch bin.OperatorToken.Kind {
					case ast.KindEqualsToken:
						suggestions = []rule.RuleSuggestion{
							{Message: buildNotNeedInAssignMessage(), FixesArr: removeBangFix},
						}
					case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken:
						suggestions = []rule.RuleSuggestion{
							{Message: buildNotNeedInEqualTestMessage(), FixesArr: removeBangFix},
						}
					case ast.KindInKeyword, ast.KindInstanceOfKeyword:
						suggestions = []rule.RuleSuggestion{
							{Message: buildNotNeedInOperatorMessage(operator), FixesArr: removeBangFix},
							{Message: buildWrapUpLeftMessage(operator), FixesArr: wrapUpLeftFix},
						}
					}
					ctx.ReportNodeWithSuggestions(node, primaryMsg, suggestions...)
					return
				}

				ctx.ReportNodeWithSuggestions(node, primaryMsg, rule.RuleSuggestion{
					Message:  buildWrapUpLeftMessage(operator),
					FixesArr: wrapUpLeftFix,
				})
			},
		}
	},
})
