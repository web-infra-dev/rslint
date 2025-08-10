package no_non_null_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// buildNoNonNullAssertionMessage creates a standardized rule message for the no-non-null-assertion rule.
// This function returns a RuleMessage with a unique identifier and descriptive text about the rule violation.
func buildNoNonNullAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noNonNull",
		Description: "Non-null assertion operator (!) is not allowed.",
	}
}

// NoNonNullAssertionRule is a linting rule that detects and reports the usage of non-null assertion operators (!).
// The rule allows non-null assertions only in specific contexts where they are necessary, such as
// the left side of assignment expressions where TypeScript requires non-null types.
//
// Rule Configuration:
// - Name: "no-non-null-assertion"
// - Purpose: Prevents unsafe usage of non-null assertions that can lead to runtime errors
// - Exceptions: Assignment expressions where non-null assertion is required by TypeScript
//
// Example violations:
//
//	const value = obj!.property;        // ❌ Not allowed
//	const value = obj?.property;        // ✅ Use optional chaining instead
//
// Example allowed usage:
//
//	obj!.property = value;              // ✅ Allowed in assignment left side
var NoNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-non-null-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			// Listen for non-null assertion expressions (!)
			ast.KindNonNullExpression: func(node *ast.Node) {
				// Check if the non-null assertion is used in an assignment expression
				parent := node.Parent
				if parent != nil && ast.IsAssignmentExpression(parent, true) {
					// Allow non-null assertions on the left side of assignments
					// This is necessary when TypeScript requires non-null types for assignment targets
					binaryExpr := parent.AsBinaryExpression()
					if binaryExpr != nil && binaryExpr.Left == node {
						return
					}
				}

				// Report the non-null assertion usage as a violation
				// This helps developers identify potentially unsafe code patterns
				ctx.ReportNode(node, buildNoNonNullAssertionMessage())
			},
		}
	},
})
