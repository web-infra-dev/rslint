package prefer_to_have_been_called_times

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferMatcherMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferMatcher",
		Description: "Use `toHaveBeenCalledTimes()`",
	}
}

func unwrapMockProperty(mockExpr *ast.Node) *ast.Node {
	if mockExpr == nil {
		return nil
	}
	mockExpr = ast.SkipParentheses(mockExpr)
	switch mockExpr.Kind {
	case ast.KindPropertyAccessExpression:
		if ast.IsOptionalChain(mockExpr) {
			return nil
		}
		pa := mockExpr.AsPropertyAccessExpression()
		if !jestUtils.IsNamedMember(pa.Name(), "mock") {
			return nil
		}
		return pa.Expression
	case ast.KindElementAccessExpression:
		if ast.IsOptionalChain(mockExpr) {
			return nil
		}
		el := mockExpr.AsElementAccessExpression()
		if !jestUtils.IsNamedMember(ast.SkipParentheses(el.ArgumentExpression), "mock") {
			return nil
		}
		return el.Expression
	default:
		return nil
	}
}

// unwrapMockCallsAccessProperty returns the mock function expression when arg is
// exactly `*.mock.calls` or `*["mock"].calls` (no indexing after `.calls`).
func unwrapMockCallsAccessProperty(arg *ast.Node) *ast.Node {
	if arg == nil {
		return nil
	}
	arg = ast.SkipParentheses(arg)
	switch arg.Kind {
	case ast.KindPropertyAccessExpression:
		if ast.IsOptionalChain(arg) {
			return nil
		}
		pa := arg.AsPropertyAccessExpression()
		if !jestUtils.IsNamedMember(pa.Name(), "calls") {
			return nil
		}
		return unwrapMockProperty(pa.Expression)
	case ast.KindElementAccessExpression:
		if ast.IsOptionalChain(arg) {
			return nil
		}
		el := arg.AsElementAccessExpression()
		if !jestUtils.IsNamedMember(ast.SkipParentheses(el.ArgumentExpression), "calls") {
			return nil
		}
		return unwrapMockProperty(el.Expression)
	default:
		return nil
	}
}

var PreferToHaveBeenCalledTimesRule = rule.Rule{
	Name: "jest/prefer-to-have-been-called-times",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil ||
					jestFnCall.Kind != jestUtils.JestFnTypeExpect ||
					jestFnCall.Matcher != "toHaveLength" {
					return
				}

				expectCall := jestFnCall.Head.Local.Node.Parent
				if expectCall == nil || expectCall.Kind != ast.KindCallExpression {
					return
				}

				args := expectCall.Arguments()
				if len(args) == 0 {
					return
				}

				inner := unwrapMockCallsAccessProperty(args[0])
				if inner == nil {
					return
				}

				matcherCall := node.AsCallExpression()
				if matcherCall == nil || matcherCall.Arguments == nil {
					return
				}

				matcherReceiver, matcherParent := jestUtils.GetAccessorReceiverAndParent(jestFnCall.MatcherEntry)
				if matcherReceiver == nil || matcherParent == nil {
					return
				}

				sourceFile := ast.GetSourceFileOfNode(node)
				if sourceFile == nil {
					return
				}

				reportNode := node
				if jestFnCall.MatcherEntry != nil && jestFnCall.MatcherEntry.Node != nil {
					reportNode = jestFnCall.MatcherEntry.Node
				}

				ctx.ReportNodeWithFixes(
					reportNode,
					buildPreferMatcherMessage(),
					rule.RuleFixReplaceRange(
						rslintUtils.TrimNodeTextRange(sourceFile, args[0]),
						rslintUtils.TrimmedNodeText(sourceFile, inner),
					),
					rule.RuleFixReplaceRange(
						core.NewTextRange(matcherReceiver.End(), matcherParent.End()),
						".toHaveBeenCalledTimes",
					),
				)
			},
		}
	},
}
