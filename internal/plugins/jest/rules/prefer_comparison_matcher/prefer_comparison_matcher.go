package prefer_comparison_matcher

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

type comparisonMatcher struct {
	matcher        string
	negatedMatcher string
}

var comparisonMatchers = map[ast.Kind]comparisonMatcher{
	ast.KindGreaterThanToken:       {matcher: "toBeGreaterThan", negatedMatcher: "toBeLessThanOrEqual"},
	ast.KindLessThanToken:          {matcher: "toBeLessThan", negatedMatcher: "toBeGreaterThanOrEqual"},
	ast.KindGreaterThanEqualsToken: {matcher: "toBeGreaterThanOrEqual", negatedMatcher: "toBeLessThan"},
	ast.KindLessThanEqualsToken:    {matcher: "toBeLessThanOrEqual", negatedMatcher: "toBeGreaterThan"},
}

func buildUseComparisonMatcherMessage(preferredMatcher string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToBeComparison",
		Description: "Prefer using `" + preferredMatcher + "` instead",
		Data: map[string]string{
			"preferredMatcher": preferredMatcher,
		},
	}
}

func isStringComparisonOperand(node *ast.Node) bool {
	return internalUtils.IsStringLiteralOrTemplate(utils.UnwrapBasicTypeAssertions(node))
}

func parseComparison(node *ast.Node) (left, right *ast.Node, matchers comparisonMatcher, ok bool) {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return nil, nil, comparisonMatcher{}, false
	}

	bin := node.AsBinaryExpression()
	matchers, ok = comparisonMatchers[bin.OperatorToken.Kind]
	if !ok || isStringComparisonOperand(bin.Left) || isStringComparisonOperand(bin.Right) {
		return nil, nil, comparisonMatcher{}, false
	}

	return bin.Left, bin.Right, matchers, true
}

func buildModifierText(jestFnCall *utils.ParsedJestFnCall) string {
	if len(jestFnCall.ModifierEntries) > 0 && jestFnCall.ModifierEntries[0].Name != "not" {
		return "." + jestFnCall.ModifierEntries[0].Name
	}
	return ""
}

var PreferComparisonMatcherRule = rule.Rule{
	Name: "jest/prefer-comparison-matcher",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil ||
					jestFnCall.Kind != utils.JestFnTypeExpect ||
					jestFnCall.MatcherEntry == nil ||
					!utils.EQUALITY_METHOD_NAMES[jestFnCall.Matcher] {
					return
				}

				expectCall := jestFnCall.Head.Local.Node.Parent
				if expectCall == nil || expectCall.Kind != ast.KindCallExpression {
					return
				}

				expectArgs := expectCall.Arguments()
				if len(expectArgs) == 0 {
					return
				}

				left, right, matchers, ok := parseComparison(expectArgs[0])
				if !ok {
					return
				}

				matcherEntry := jestFnCall.MatcherEntry
				if matcherEntry.Node == nil ||
					matcherEntry.Node.Parent == nil ||
					matcherEntry.Call == nil ||
					node != matcherEntry.Call {
					return
				}

				matcherArgs := matcherEntry.Call.AsCallExpression().Arguments.Nodes
				if len(matcherArgs) == 0 {
					return
				}

				matcherArg := matcherArgs[0]
				matcherValue, ok := utils.IsBooleanLiteral(matcherArg)
				if !ok {
					return
				}

				preferredMatcher := matchers.matcher
				if matcherValue == slices.Contains(jestFnCall.Modifiers, "not") {
					preferredMatcher = matchers.negatedMatcher
				}

				comparison := ast.SkipParentheses(expectArgs[0])
				leftText := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, left, false)
				rightText := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, right, false)
				modifierRange := core.NewTextRange(expectCall.End(), matcherEntry.Node.Parent.End())

				ctx.ReportNodeWithFixes(
					matcherEntry.Node,
					buildUseComparisonMatcherMessage(preferredMatcher),
					rule.RuleFixReplace(ctx.SourceFile, comparison, leftText),
					rule.RuleFixReplaceRange(modifierRange, buildModifierText(jestFnCall)+"."+preferredMatcher),
					rule.RuleFixReplace(ctx.SourceFile, matcherArg, rightText),
				)
			},
		}
	},
}
