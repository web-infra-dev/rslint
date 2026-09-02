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
// The walk stops at the first deferred-execution boundary, which is what keeps
// the rule honest in two directions. Downwards, it means only the registration
// that the `if` itself decides to run is reported: in
// `if (x) { describe('a', () => { test('b', fn) }) }` the inner `test` walk
// stops at the describe callback's body, so only the outer `describe` is
// reported. Upwards, it means a condition inside an unrelated enclosing
// function never leaks onto a registration that function merely contains.
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
		if isDeferredExecutionBoundary(child, parent) {
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

// isDeferredExecutionBoundary reports whether stepping from child up to parent
// crosses into code that runs later than, and independently of, whatever
// controls parent's own evaluation — so a registration on the far side of it
// is no longer conditional on anything wrapping parent.
//
// A function, method, constructor or accessor body, and the runtime-evaluated
// parts of a parameter, qualify: none runs when the declaration itself is
// evaluated, only later when the function is called. An instance field
// initializer (`p = expr`, no `static`) qualifies too: it runs once per `new`,
// not when the class declaration is evaluated.
//
// A class static block, a static field initializer (`static p = expr`), a
// computed member name (`[expr]`), and a decorator (`@dec(expr)`) do not
// qualify even though each sits inside a class body: all four run
// synchronously as part of evaluating the class declaration itself, so a
// registration inside one is exactly as conditional as the class declaration
// that contains it. Climbing out of a computed name or a decorator lands back
// on the method or accessor it belongs to — which is itself function-like —
// so the boundary check below only treats that step as deferred when child is
// specifically the declaration's body or one of its parameters, not merely
// because parent is function-like.
func isDeferredExecutionBoundary(child *ast.Node, parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindComputedPropertyName, ast.KindDecorator, ast.KindClassStaticBlockDeclaration:
		return false
	case ast.KindParameter:
		// Parameter decorators run while the enclosing class is evaluated.
		// Defaults and binding patterns run later, when the function is called.
		return child.Kind != ast.KindDecorator
	case ast.KindPropertyDeclaration:
		// A non-static field's initializer is deferred until construction, but
		// its computed name and decorators run while the class is evaluated.
		property := parent.AsPropertyDeclaration()
		return !ast.HasStaticModifier(parent) && child == property.Initializer
	}
	if !ast.IsFunctionLikeDeclaration(parent) {
		return false
	}
	return child == parent.Body()
}
