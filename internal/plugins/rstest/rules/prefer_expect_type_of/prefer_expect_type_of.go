package prefer_expect_type_of

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func buildPreferExpectTypeOfMessage(value string, typeText string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "preferExpectTypeOf",
		Description: fmt.Sprintf(
			"Use `expect(%s).toBeTypeOf(%s)` instead of `expect(typeof %s).toBe(%s)`",
			value,
			typeText,
			value,
			typeText,
		),
	}
}

// typeofAssertion describes one matched `expect(typeof value).toBe(type)`
// chain. The rule keeps the nodes rather than their text so that the eager
// pass stays free of source slicing beyond the two message placeholders.
type typeofAssertion struct {
	// typeofNode is the `typeof value` expression inside the expect head.
	typeofNode *ast.Node
	// operand is what `typeof` is applied to, as written.
	operand *ast.Node
	// matcherNode is the matcher accessor to rewrite to `toBeTypeOf`.
	matcherNode *ast.Node
	// matcherArgument is the single argument the matcher was called with.
	matcherArgument *ast.Node
}

var PreferExpectTypeOfRule = rule.Rule{
	Name:   "rstest/prefer-expect-type-of",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				assertion, ok := matchTypeofAssertion(analysis.ParseExpectCall(node))
				if !ok {
					return
				}

				message := buildPreferExpectTypeOfMessage(
					utils.TrimmedNodeText(ctx.SourceFile, assertion.operand),
					utils.TrimmedNodeText(ctx.SourceFile, assertion.matcherArgument),
				)
				ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
					return buildFixes(ctx, assertion)
				})
			},
		}
	},
}

func matchTypeofAssertion(parsed *rstestUtils.ParsedRstestExpectCall) (typeofAssertion, bool) {
	if parsed == nil ||
		parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
		parsed.MatcherEntry == nil ||
		parsed.Head == nil {
		return typeofAssertion{}, false
	}

	// `expect.poll(fn)` takes a function and `expect.element(locator)` takes a
	// locator, so neither can carry the `typeof value` argument this rule
	// rewrites; static members such as `expect.assertions(1)` have no head.
	if parsed.Entry != rstestUtils.RstestExpectEntryCall &&
		parsed.Entry != rstestUtils.RstestExpectEntrySoft {
		return typeofAssertion{}, false
	}

	if parsed.Matcher != "toBe" && parsed.Matcher != "toEqual" {
		return typeofAssertion{}, false
	}
	if len(parsed.Matchers) == 0 || parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall {
		return typeofAssertion{}, false
	}

	// A computed identifier key names its matcher at runtime, so `toBe` here is
	// the name of a variable rather than of a matcher. Reporting it would flag
	// whatever assertion that variable holds.
	matcherNode := parsed.MatcherEntry.Node
	if matcherNode == nil || testFramework.IsComputedIdentifierAccessor(matcherNode) {
		return typeofAssertion{}, false
	}

	headArguments := parsed.Head.Arguments()
	if len(headArguments) == 0 {
		return typeofAssertion{}, false
	}
	typeofNode := ast.SkipParentheses(headArguments[0])
	if typeofNode == nil || typeofNode.Kind != ast.KindTypeOfExpression {
		return typeofAssertion{}, false
	}
	operand := typeofNode.AsTypeOfExpression().Expression
	if operand == nil {
		return typeofAssertion{}, false
	}

	matcherArg := matcherArgument(parsed.MatcherEntry.Call)
	if matcherArg == nil {
		return typeofAssertion{}, false
	}

	return typeofAssertion{
		typeofNode:      typeofNode,
		operand:         operand,
		matcherNode:     matcherNode,
		matcherArgument: matcherArg,
	}, true
}

// matcherArgument returns the matcher's single argument, or nil when the
// matcher takes a different number of arguments or spreads them — neither maps
// onto `toBeTypeOf(type)`.
func matcherArgument(call *ast.Node) *ast.Node {
	if call == nil {
		return nil
	}
	arguments := call.Arguments()
	if len(arguments) != 1 || arguments[0].Kind == ast.KindSpreadElement {
		return nil
	}
	return arguments[0]
}

// buildFixes rewrites the two places that differ from the desired form and
// leaves the rest of the call untouched, so `expect(actual, message)`, the
// `expect.soft` factory, a renamed or `import.meta.rstest` expect root, the
// modifier chain and the matcher argument all survive the fix.
func buildFixes(ctx rule.RuleContext, assertion typeofAssertion) []rule.RuleFix {
	typeofStart := utils.TrimNodeTextRange(ctx.SourceFile, assertion.typeofNode).Pos()
	operandStart := utils.TrimNodeTextRange(ctx.SourceFile, assertion.operand).Pos()
	if operandStart <= typeofStart {
		return nil
	}

	matcherRange, matcherText, ok := testFramework.AccessorReplacement(
		ctx.SourceFile,
		assertion.matcherNode,
		"toBeTypeOf",
	)
	if !ok {
		return nil
	}

	return []rule.RuleFix{
		rule.RuleFixRemoveRange(core.NewTextRange(typeofStart, operandStart)),
		rule.RuleFixReplaceRange(matcherRange, matcherText),
	}
}
