package no_non_null_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// isAssignmentTarget checks if a node is used as the left side of an assignment.
func isAssignmentTarget(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if ast.IsAssignmentExpression(parent, true) {
		return parent.AsBinaryExpression().Left == node
	}
	if parent.Kind == ast.KindPropertyAccessExpression || parent.Kind == ast.KindElementAccessExpression {
		return isAssignmentTarget(parent)
	}
	return false
}

var NoNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-non-null-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		msg := rule.RuleMessage{
			Id:          "noNonNull",
			Description: "Forbidden non-null assertion.",
		}
		suggestMsg := rule.RuleMessage{
			Id:          "suggestOptionalChain",
			Description: "Consider using the optional chain operator `?.` instead. This operator includes runtime checks, so it is safer than the compile-only non-null assertion operator.",
		}

		return rule.RuleListeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				expression := node.AsNonNullExpression().Expression
				parent := node.Parent

				// Get the ! token range using scanner
				exclamScanner := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
				exclamRange := exclamScanner.TokenRange()

				// Check if we can provide an optional chain suggestion
				var suggestion *rule.RuleSuggestion

				if parent.Kind == ast.KindPropertyAccessExpression && parent.Expression() == node {
					propAccess := parent.AsPropertyAccessExpression()

					// Don't suggest if this is an assignment target
					if !isAssignmentTarget(parent) {
						if propAccess.QuestionDotToken != nil {
							// Already optional (x!?.y) → just remove !
							suggestion = &rule.RuleSuggestion{
								Message:  suggestMsg,
								FixesArr: []rule.RuleFix{rule.RuleFixRemoveRange(exclamRange)},
							}
						} else {
							// Dot notation (x!.y) → remove ! and replace . with ?.
							exclamScanner.Scan() // advance to . token
							dotRange := exclamScanner.TokenRange()
							suggestion = &rule.RuleSuggestion{
								Message: suggestMsg,
								FixesArr: []rule.RuleFix{
									rule.RuleFixRemoveRange(exclamRange),
									rule.RuleFixReplaceRange(dotRange, "?."),
								},
							}
						}
					}
				} else if parent.Kind == ast.KindElementAccessExpression && parent.Expression() == node {
					elemAccess := parent.AsElementAccessExpression()

					if !isAssignmentTarget(parent) {
						if elemAccess.QuestionDotToken != nil {
							// Already optional (x!?.[y]) → just remove !
							suggestion = &rule.RuleSuggestion{
								Message:  suggestMsg,
								FixesArr: []rule.RuleFix{rule.RuleFixRemoveRange(exclamRange)},
							}
						} else {
							// Computed access (x![y]) → replace ! with ?.
							suggestion = &rule.RuleSuggestion{
								Message:  suggestMsg,
								FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(exclamRange, "?.")},
							}
						}
					}
				} else if parent.Kind == ast.KindCallExpression && parent.Expression() == node {
					callExpr := parent.AsCallExpression()

					if !isAssignmentTarget(parent) {
						if callExpr.QuestionDotToken != nil {
							// Already optional (x!?.()) → just remove !
							suggestion = &rule.RuleSuggestion{
								Message:  suggestMsg,
								FixesArr: []rule.RuleFix{rule.RuleFixRemoveRange(exclamRange)},
							}
						} else {
							// Call (x!()) → replace ! with ?.
							suggestion = &rule.RuleSuggestion{
								Message:  suggestMsg,
								FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(exclamRange, "?.")},
							}
						}
					}
				}

				if suggestion != nil {
					ctx.ReportNodeWithSuggestions(node, msg, *suggestion)
				} else {
					ctx.ReportNode(node, msg)
				}
			},
		}
	},
})
