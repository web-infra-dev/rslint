package prefer_expect_assertions

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	jest "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_expect_assertions"
)

// PreferExpectAssertionsRule counts assertions set up in both beforeEach and
// afterEach: jest-circus checks expect.assertions and expect.hasAssertions on
// `test_done`, after every afterEach hook of the test has run.
var PreferExpectAssertionsRule = shared.NewRule(shared.Config{
	Name:          "jest/prefer-expect-assertions",
	Schema:        shared.Schema,
	CoveringHooks: []string{"beforeEach", "afterEach"},
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := jest.GetJestCallAnalysis(ctx)
		return shared.Runtime{
			Classify: func(node *ast.Node) shared.Registration {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return shared.Registration{}
				}
				arguments := node.AsCallExpression().Arguments.Nodes
				switch parsed.Kind {
				case jest.JestFnTypeDescribe:
					return shared.Registration{Kind: shared.RegistrationDescribe}
				case jest.JestFnTypeHook:
					registration := shared.Registration{Kind: shared.RegistrationHook, HookName: parsed.Name}
					if len(arguments) > 0 {
						registration.Callback = arguments[0]
					}
					return registration
				case jest.JestFnTypeTest:
					registration := shared.Registration{Kind: shared.RegistrationTest}
					if len(arguments) >= 2 {
						if callback := ast.SkipParentheses(arguments[1]); ast.IsFunctionExpressionOrArrowFunction(callback) {
							registration.Callback = callback
						}
					}
					return registration
				}
				return shared.Registration{}
			},
			ParseStatic: func(node *ast.Node) *shared.StaticCall {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil || parsed.Kind != jest.JestFnTypeExpect || len(parsed.MemberEntries) != 1 {
					return nil
				}
				entry := parsed.MemberEntries[0]
				if !shared.IsStaticMemberName(entry.Name) || !isStaticAccessOf(node, parsed.Head.Local.Node) {
					return nil
				}
				return &shared.StaticCall{Call: node, Member: entry.Name, MemberNode: entry.Node}
			},
			IsExpect: func(node *ast.Node) bool {
				parsed := analysis.ParseFnCall(node)
				return parsed != nil && parsed.Kind == jest.JestFnTypeExpect
			},
			IsHasAssertionsReference: func(node *ast.Node) bool {
				return isHasAssertionsReference(ctx, node)
			},
			ExpectSpelling: func(*ast.Node, *ast.Node) (string, bool) {
				return "expect", true
			},
		}
	},
})

// isStaticAccessOf reports whether call invokes a member read directly off the
// expect binding, as in `expect.hasAssertions()`. A matcher named like a static
// member, as in `expect(value).assertions(1)`, reads it off expect's result
// instead and does not declare an assertion count.
func isStaticAccessOf(call *ast.Node, head *ast.Node) bool {
	if head == nil || head.Parent == nil {
		return false
	}
	access := head.Parent
	if !ast.IsPropertyAccessExpression(access) && !ast.IsElementAccessExpression(access) {
		return false
	}
	return access.Expression() == head && ast.SkipParentheses(call.AsCallExpression().Expression) == access
}

// isHasAssertionsReference recognizes `beforeEach(expect.hasAssertions)`. Jest's
// hasAssertions reads the global matcher state rather than its receiver, so it
// works when a hook calls it detached.
func isHasAssertionsReference(ctx rule.RuleContext, node *ast.Node) bool {
	if !ast.IsPropertyAccessExpression(node) {
		return false
	}
	name := node.Name()
	root := node.Expression()
	if name == nil || name.Text() != "hasAssertions" || root == nil || root.Kind != ast.KindIdentifier || ctx.Refs == nil {
		return false
	}
	resolved, _, _ := testFramework.ResolveFunctionIdentifierReferenceFromSymbol(
		root.AsIdentifier().Text,
		root,
		ctx.Refs.Resolve(root),
		ctx.SourceFile,
		"@jest/globals",
	)
	return jest.ApplyGlobalJestAlias(resolved, ctx.Settings) == "expect"
}
