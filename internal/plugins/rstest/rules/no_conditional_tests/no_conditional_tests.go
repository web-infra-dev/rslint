package no_conditional_tests

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

func buildNoConditionalTestsMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noConditionalTests",
		Description: "Avoid using if conditions in a test",
	}
}

var NoConditionalTestsRule = rule.Rule{
	Name:   "rstest/no-conditional-tests",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseFnCall(node)
				// Hooks are excluded: only test and suite registrations count,
				// matching upstream's `test` / `it` / `describe` name list.
				if parsed == nil ||
					(parsed.Kind != rstestUtils.RstestFnTypeTest &&
						parsed.Kind != rstestUtils.RstestFnTypeDescribe) {
					return
				}

				if !registeredUnderIfBranch(node) {
					return
				}

				reportNode := node
				if parsed.Head.Local.Node != nil {
					reportNode = parsed.Head.Local.Node
				}
				ctx.ReportRange(
					internalUtils.TrimNodeTextRange(ctx.SourceFile, reportNode),
					buildNoConditionalTestsMessage(),
				)
			},
		}
	},
}

// registeredUnderIfBranch walks up from a registration call and reports whether
// it is reached through the then or else branch of an `if`.
//
// The walk stops at the first function boundary, which is what keeps the rule
// honest in two directions. Downwards, it means only the registration that the
// `if` itself decides to run is reported: in
// `if (x) { describe('a', () => { test('b', fn) }) }` the inner `test` walk
// stops at the describe callback, so only the outer `describe` is reported.
// Upwards, it means a condition inside an unrelated enclosing function never
// leaks onto a registration that function merely contains.
//
// Arriving through an `if`'s condition is not a conditional registration —
// `if (test('a')) {}` runs the call unconditionally — so the walk continues
// past it rather than reporting, and an outer `if` can still catch it.
func registeredUnderIfBranch(node *ast.Node) bool {
	child := node
	for {
		parent := child.Parent
		if parent == nil || parent.Kind == ast.KindSourceFile {
			return false
		}
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(parent) {
			return false
		}
		if parent.Kind == ast.KindIfStatement {
			ifStatement := parent.AsIfStatement()
			if child == ifStatement.ThenStatement || child == ifStatement.ElseStatement {
				return true
			}
		}
		child = parent
	}
}
