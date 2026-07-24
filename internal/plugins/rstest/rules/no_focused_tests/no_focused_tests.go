package no_focused_tests

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

func buildErrorFocusedTestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "focusedTest",
		Description: "Unexpected focused test",
	}
}

func buildErrorSuggestRemoveFocusMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestRemoveFocus",
		Description: "Suggest removing focus from test",
	}
}

var NoFocusedTestsRule = rule.Rule{
	Name:   "rstest/no-focused-tests",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := rstestUtils.ParseRstestFnCall(node, ctx)
				if parsed == nil ||
					(parsed.Kind != rstestUtils.RstestFnTypeDescribe &&
						parsed.Kind != rstestUtils.RstestFnTypeTest) {
					return
				}

				onlyEntries := make([]rstestUtils.ParsedRstestFnMemberEntry, 0, 1)
				for _, entry := range parsed.MemberEntries {
					if entry.Name == "only" {
						onlyEntries = append(onlyEntries, entry)
					}
				}
				if len(onlyEntries) == 0 {
					return
				}

				reportNode := onlyEntries[0].Node
				fixes, ok := removalFixes(onlyEntries, ctx)
				if !ok {
					ctx.ReportNode(reportNode, buildErrorFocusedTestMessage())
					return
				}

				ctx.ReportNodeWithSuggestions(
					reportNode,
					buildErrorFocusedTestMessage(),
					rule.RuleSuggestion{
						Message:  buildErrorSuggestRemoveFocusMessage(),
						FixesArr: fixes,
					},
				)
			},
		}
	},
}

func removalFixes(
	entries []rstestUtils.ParsedRstestFnMemberEntry,
	ctx rule.RuleContext,
) ([]rule.RuleFix, bool) {
	ranges := make([]core.TextRange, 0, len(entries))
	for _, entry := range entries {
		textRange, ok := memberRemovalRange(entry.Node)
		if !ok || rslintUtils.HasCommentInSpan(ctx.Comments.All(), textRange.Pos(), textRange.End()) {
			return nil, false
		}
		ranges = append(ranges, textRange)
	}

	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Pos() < ranges[j].Pos()
	})

	fixes := make([]rule.RuleFix, 0, len(ranges))
	for index, textRange := range ranges {
		if index > 0 && ranges[index-1].End() > textRange.Pos() {
			return nil, false
		}
		fixes = append(fixes, rule.RuleFixRemoveRange(textRange))
	}
	return fixes, true
}

func memberRemovalRange(nameNode *ast.Node) (core.TextRange, bool) {
	if nameNode == nil {
		return core.TextRange{}, false
	}

	memberNode := nameNode
	for memberNode.Parent != nil && memberNode.Parent.Kind == ast.KindParenthesizedExpression {
		memberNode = memberNode.Parent
	}
	if memberNode.Parent == nil {
		return core.TextRange{}, false
	}

	accessor := memberNode.Parent
	var receiver *ast.Node
	switch accessor.Kind {
	case ast.KindPropertyAccessExpression:
		property := accessor.AsPropertyAccessExpression()
		if property.Name() != memberNode {
			return core.TextRange{}, false
		}
		receiver = property.Expression
	case ast.KindElementAccessExpression:
		element := accessor.AsElementAccessExpression()
		if ast.SkipParentheses(element.ArgumentExpression) != nameNode {
			return core.TextRange{}, false
		}
		receiver = element.Expression
	default:
		return core.TextRange{}, false
	}

	if receiver == nil || receiver.End() >= accessor.End() {
		return core.TextRange{}, false
	}
	return core.NewTextRange(receiver.End(), accessor.End()), true
}
