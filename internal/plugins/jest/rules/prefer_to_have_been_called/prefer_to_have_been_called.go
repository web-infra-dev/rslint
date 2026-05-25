package prefer_to_have_been_called

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message Builders

func buildPreferMatcherErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferMatcher",
		Description: "Use `toHaveBeenCalled`",
	}
}

func isZeroLiteral(node *ast.Node) bool {
	node = jestUtils.UnwrapTypeAssertions(node)
	if node == nil {
		return false
	}

	return node.Kind == ast.KindNumericLiteral && node.AsNumericLiteral().Text == "0"
}

var PreferToHaveBeenCalledRule = rule.Rule{
	Name: "jest/prefer-to-have-been-called",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil ||
					jestFnCall.Kind != jestUtils.JestFnTypeExpect ||
					(jestFnCall.Matcher != "toBeCalledTimes" && jestFnCall.Matcher != "toHaveBeenCalledTimes") {
					return
				}

				matcherCall := node.AsCallExpression()
				if matcherCall == nil || matcherCall.Arguments == nil || len(matcherCall.Arguments.Nodes) == 0 {
					return
				}

				if !isZeroLiteral(matcherCall.Arguments.Nodes[0]) {
					return
				}

				matcherReceiver, matcherParent := jestUtils.GetAccessorReceiverAndParent(jestFnCall.MatcherEntry)
				if matcherParent == nil || matcherReceiver == nil {
					return
				}

				var notModifier *jestUtils.ParsedJestFnMemberEntry
				for i := range jestFnCall.ModifierEntries {
					if jestFnCall.ModifierEntries[i].Name == "not" {
						notModifier = &jestFnCall.ModifierEntries[i]
						break
					}
				}

				replaceStart := matcherReceiver.End()
				if notModifier != nil {
					notReceiver, _ := jestUtils.GetAccessorReceiverAndParent(notModifier)
					if notReceiver == nil {
						return
					}
					replaceStart = notReceiver.End()
				}

				replacementMatcher := ".not.toHaveBeenCalled"
				if notModifier != nil {
					replacementMatcher = ".toHaveBeenCalled"
				}

				ctx.ReportNodeWithFixes(
					jestFnCall.MatcherEntry.Node,
					buildPreferMatcherErrorMessage(),
					rule.RuleFixReplaceRange(
						core.NewTextRange(matcherCall.Arguments.Pos(), matcherCall.Arguments.End()),
						"",
					),
					rule.RuleFixReplaceRange(
						core.NewTextRange(replaceStart, matcherParent.End()),
						replacementMatcher,
					),
				)
			},
		}
	},
}
