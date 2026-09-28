package no_unnecessary_assertion

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_unnecessary_assertion"
)

// headCall returns the call that invokes the expect head, looking through
// parentheses the way ESTree does, or nil when the head is not called.
func headCall(head *ast.Node) *ast.Node {
	if head == nil {
		return nil
	}
	callee := head
	for callee.Parent != nil && callee.Parent.Kind == ast.KindParenthesizedExpression {
		callee = callee.Parent
	}
	call := callee.Parent
	if call == nil || call.Kind != ast.KindCallExpression || call.AsCallExpression().Expression != callee {
		return nil
	}
	return call
}

// canBe reports whether the subject type, or any member of a union subject
// type, carries one of the desired flags. Like eslint-plugin-jest it looks
// only at the top-level flags: type parameters, intersections and void are
// not expanded.
func canBe(subjectType *checker.Type, desired checker.TypeFlags) bool {
	for _, part := range utils.UnionTypeParts(subjectType) {
		if checker.Type_flags(part)&desired != 0 {
			return true
		}
	}
	return false
}

var NoUnnecessaryAssertionRule = shared.NewRule(shared.Config{
	Name: "jest/no-unnecessary-assertion",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		return shared.Runtime{Check: func(node *ast.Node) *shared.Assertion {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil || parsed.MatcherEntry == nil {
				return nil
			}
			call := headCall(parsed.Head.Local.Node)
			if call == nil {
				return nil
			}

			matcher, ok := shared.CallMatcher(parsed.Matcher)
			if !ok || !shared.OnlyNotModifiers(parsed.Modifiers) {
				return nil
			}

			arguments := call.Arguments()
			// Without a subject the type checker has nothing to resolve and
			// answers with its error type, which counts as `any`.
			if len(arguments) == 0 || arguments[0] == nil {
				return nil
			}
			subjectType := ctx.TypeChecker.GetTypeAtLocation(ast.SkipParentheses(arguments[0]))
			if subjectType == nil ||
				canBe(subjectType, checker.TypeFlagsAny|checker.TypeFlagsUnknown|matcher.Flags) {
				return nil
			}

			return &shared.Assertion{Node: node, Matcher: matcher}
		}}
	},
})
