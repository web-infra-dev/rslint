package no_confusing_non_null_assertion

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

var confusingOperators = map[ast.Kind]bool {
	ast.KindEqualsToken:               true, // =
	ast.KindEqualsEqualsToken:         true, // ==
	ast.KindEqualsEqualsEqualsToken:   true, // ===
	ast.KindInKeyword:                 true, // in
	ast.KindInstanceOfKeyword:         true, // instanceof
}

var NoConfusingNonNullAssertionRule = rule.Rule{
	Name: "no-confusing-non-null-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		sourceFile := ctx.SourceFile
		
		checkNode := func(node *ast.Node) {
			var operator ast.Kind
			var left *ast.Node
			var operatorToken ast.Kind
			
			// For Go's TypeScript AST, both binary expressions and assignments are KindBinaryExpression
			if node.Kind != ast.KindBinaryExpression {
				return
			}
			
			binaryExpr := node.AsBinaryExpression()
			operator = binaryExpr.OperatorToken.Kind
			left = binaryExpr.Left
			operatorToken = binaryExpr.OperatorToken.Kind
			
			// Check if it's a confusing operator
			if !confusingOperators[operator] {
				return
			}
			
			// Get the last token of the left side
			leftRange := utils.TrimNodeTextRange(sourceFile, left)
			
			// Check if the left side ends with an exclamation mark
			// Get the text of the left side to check for exclamation mark
			leftText := string(sourceFile.Text()[leftRange.Pos():leftRange.End()])
			if !strings.HasSuffix(leftText, "!") {
				return
			}
			
			// Find the position of the exclamation mark
			exclamationPos := leftRange.End() - 1
			
			// Check various cases where we should NOT report:
			if exclamationPos > 0 {
				charBeforeExclamation := sourceFile.Text()[exclamationPos-1]
				// 1. If there's a closing parenthesis before !, like (a)!
				if charBeforeExclamation == ')' {
					return
				}
				// 2. If there's another ! before !, like a!!
				if charBeforeExclamation == '!' {
					return
				}
			}
			
			// Determine message and suggestions based on operator
			var message rule.RuleMessage
			var suggestions []rule.RuleSuggestion
			
			operatorStr := getOperatorString(operatorToken)
			
			switch operator {
			case ast.KindEqualsToken:
				message = rule.RuleMessage{
					Id:          "confusingAssign",
					Description: "Confusing combination of non-null assertion and assignment like `a! = b`, which looks very similar to `a != b`.",
				}
				suggestions = []rule.RuleSuggestion{
					{
						Message: rule.RuleMessage{
							Id:          "wrapUpLeft",
							Description: "Wrap the left-hand side in parentheses to avoid confusion with \"" + operatorStr + "\" operator.",
						},
						FixesArr: wrapUpLeftFixes(sourceFile, left, exclamationPos),
					},
				}
				
			case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken:
				message = rule.RuleMessage{
					Id:          "confusingEqual",
					Description: "Confusing combination of non-null assertion and equality test like `a! == b`, which looks very similar to `a !== b`.",
				}
				suggestions = []rule.RuleSuggestion{
					{
						Message: rule.RuleMessage{
							Id:          "wrapUpLeft",
							Description: "Wrap the left-hand side in parentheses to avoid confusion with \"" + operatorStr + "\" operator.",
						},
						FixesArr: wrapUpLeftFixes(sourceFile, left, exclamationPos),
					},
				}
				
			case ast.KindInKeyword, ast.KindInstanceOfKeyword:
				message = rule.RuleMessage{
					Id:          "confusingOperator",
					Description: "Confusing combination of non-null assertion and `" + operatorStr + "` operator like `a! " + operatorStr + " b`, which might be misinterpreted as `!(a " + operatorStr + " b)`.",
				}
				suggestions = []rule.RuleSuggestion{
					{
						Message: rule.RuleMessage{
							Id:          "wrapUpLeft",
							Description: "Wrap the left-hand side in parentheses to avoid confusion with \"" + operatorStr + "\" operator.",
						},
						FixesArr: wrapUpLeftFixes(sourceFile, left, exclamationPos),
					},
				}
			}
			
			ctx.ReportNodeWithSuggestions(node, message, suggestions...)
		}
		
		return rule.RuleListeners{
			ast.KindBinaryExpression: checkNode,
		}
	},
}

func getOperatorString(operatorToken ast.Kind) string {
	switch operatorToken {
	case ast.KindEqualsToken:
		return "="
	case ast.KindEqualsEqualsToken:
		return "=="
	case ast.KindEqualsEqualsEqualsToken:
		return "==="
	case ast.KindInKeyword:
		return "in"
	case ast.KindInstanceOfKeyword:
		return "instanceof"
	default:
		return ""
	}
}

func wrapUpLeftFixes(sourceFile *ast.SourceFile, left *ast.Node, exclamationPos int) []rule.RuleFix {
	return []rule.RuleFix{
		rule.RuleFixInsertBefore(sourceFile, left, "("),
		rule.RuleFixInsertAfter(left, ")"),
	}
}