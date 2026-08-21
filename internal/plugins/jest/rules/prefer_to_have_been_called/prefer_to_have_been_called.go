package prefer_to_have_been_called

import (
	"github.com/microsoft/typescript-go/shim/ast"
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
	node = jestUtils.UnwrapBasicTypeAssertions(node)
	if node == nil {
		return false
	}

	return node.Kind == ast.KindNumericLiteral && node.AsNumericLiteral().Text == "0"
}

var PreferToHaveBeenCalledRule = rule.Rule{
	Name:   "jest/prefer-to-have-been-called",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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

				reportNode := node
				if jestFnCall.MatcherEntry != nil && jestFnCall.MatcherEntry.Node != nil {
					reportNode = jestFnCall.MatcherEntry.Node
				}
				message := buildPreferMatcherErrorMessage()
				reportWithoutFix := func() {
					ctx.ReportNode(reportNode, message)
				}

				_, matcherAccessor := jestUtils.GetAccessorReceiverAndParent(jestFnCall.MatcherEntry)
				if matcherAccessor == nil {
					reportWithoutFix()
					return
				}

				var notModifier *jestUtils.ParsedJestFnMemberEntry
				for i := range jestFnCall.ModifierEntries {
					if jestFnCall.ModifierEntries[i].Name == "not" {
						notModifier = &jestFnCall.ModifierEntries[i]
						break
					}
				}

				fixes := make([]rule.RuleFix, 0, 4)
				matcherFix, ok := jestUtils.ReplaceMemberNameFix(
					ctx,
					jestFnCall.MatcherEntry,
					"toHaveBeenCalled",
				)
				if !ok {
					reportWithoutFix()
					return
				}
				fixes = append(fixes, matcherFix)

				callFix, ok := jestUtils.ReplaceCallSuffixFix(ctx.SourceFile, node, "()")
				if !ok {
					reportWithoutFix()
					return
				}
				fixes = append(fixes, callFix)

				if notModifier != nil {
					removeNotFixes, ok := jestUtils.RemoveMemberAccessorFixes(ctx, notModifier)
					if !ok {
						reportWithoutFix()
						return
					}
					fixes = append(fixes, removeNotFixes...)
				} else {
					insertNotFix, ok := jestUtils.InsertMemberBeforeAccessorFix(
						ctx,
						jestFnCall.MatcherEntry,
						"not",
					)
					if !ok {
						reportWithoutFix()
						return
					}
					fixes = append(fixes, insertNotFix)
				}

				ctx.ReportNodeWithFixes(
					reportNode,
					message,
					fixes...,
				)
			},
		}
	},
}
