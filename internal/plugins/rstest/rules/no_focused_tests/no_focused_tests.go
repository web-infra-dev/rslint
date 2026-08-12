package no_focused_tests

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
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
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		reported := map[*ast.Node]bool{}
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil ||
					(parsed.Kind != rstestUtils.RstestFnTypeDescribe &&
						parsed.Kind != rstestUtils.RstestFnTypeTest) {
					return
				}

				focusEntries := parsed.FocusEntries()
				if len(focusEntries) == 0 {
					return
				}

				var reportNode *ast.Node
				for _, entry := range focusEntries {
					if entry.Node != nil && !reported[entry.Node] {
						reportNode = entry.Node
						break
					}
				}
				if reportNode == nil {
					return
				}
				for _, entry := range focusEntries {
					if entry.Node != nil {
						reported[entry.Node] = true
					}
				}

				fixes, ok := removalFixes(focusEntries, ctx)
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
	ranges := make([]core.TextRange, 0, len(entries)*2)
	for _, entry := range entries {
		entryRanges, ok := testFramework.RemoveAccessorEntryRanges(
			ctx.SourceFile,
			ctx.Comments.All(),
			&entry,
		)
		if !ok {
			return nil, false
		}
		ranges = append(ranges, entryRanges...)
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
