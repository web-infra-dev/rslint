package no_focused_tests

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
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
		suggested := map[*ast.Node]bool{}
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

				reportNode := focusedCallReportNode(node, parsed, focusEntries)
				hasUnsuggestedEntry := false
				for _, entry := range focusEntries {
					if entry.Node != nil && !suggested[entry.Node] {
						hasUnsuggestedEntry = true
						break
					}
				}
				if !hasUnsuggestedEntry {
					ctx.ReportNode(reportNode, buildErrorFocusedTestMessage())
					return
				}

				fixes, ok := removalFixes(focusEntries, ctx)
				if !ok {
					ctx.ReportNode(reportNode, buildErrorFocusedTestMessage())
					return
				}
				for _, entry := range focusEntries {
					if entry.Node != nil {
						suggested[entry.Node] = true
					}
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

func focusedCallReportNode(
	node *ast.Node,
	parsed *rstestUtils.ParsedRstestFnCall,
	focusEntries []rstestUtils.ParsedRstestFnMemberEntry,
) *ast.Node {
	for _, focusEntry := range focusEntries {
		for _, memberEntry := range parsed.MemberEntries {
			if focusEntry.Node != nil && focusEntry.Node == memberEntry.Node {
				return focusEntry.Node
			}
		}
	}

	// No focus entry belongs to this call site, so the focus came in through an
	// alias. The call site's own members are then whatever was chained onto the
	// alias — `.skip`, `.each` — and pointing "Unexpected focused test" at one
	// of those names the reader did not write as focus is actively misleading.
	// The identifier that resolves to the focused API is the honest anchor, and
	// it is what the member-less alias case already reports on.
	if parsed.Head.Local.Node != nil {
		return parsed.Head.Local.Node
	}
	if len(parsed.MemberEntries) > 0 {
		lastEntry := parsed.MemberEntries[len(parsed.MemberEntries)-1]
		if lastEntry.Node != nil {
			return lastEntry.Node
		}
	}
	return node
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
