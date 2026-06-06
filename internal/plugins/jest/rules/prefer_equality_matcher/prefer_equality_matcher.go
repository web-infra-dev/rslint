package prefer_equality_matcher

import (
	"maps"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var suggestedEqualityMatchers = slices.Sorted(maps.Keys(utils.EQUALITY_METHOD_NAMES))

func buildUseEqualityMatcherErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useEqualityMatcher",
		Description: "Prefer using one of the equality matchers instead",
	}
}

func buildSuggestEqualityMatcherMessage(equalityMatcher string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestEqualityMatcher",
		Description: "Use `" + equalityMatcher + "`",
	}
}

func isBooleanLiteral(node *ast.Node) (value bool, ok bool) {
	if node == nil {
		return false, false
	}
	switch node.Kind {
	case ast.KindTrueKeyword:
		return true, true
	case ast.KindFalseKeyword:
		return false, true
	default:
		return false, false
	}
}

func parseStrictEqualityComparison(node *ast.Node) (left, right *ast.Node, negated bool, ok bool) {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return nil, nil, false, false
	}

	bin := node.AsBinaryExpression()
	switch bin.OperatorToken.Kind {
	case ast.KindEqualsEqualsEqualsToken:
		return bin.Left, bin.Right, false, true
	case ast.KindExclamationEqualsEqualsToken:
		return bin.Left, bin.Right, true, true
	default:
		return nil, nil, false, false
	}
}

func buildModifierText(jestFnCall *utils.ParsedJestFnCall, addNotModifier bool) string {
	modifierText := ""
	if len(jestFnCall.ModifierEntries) > 0 && jestFnCall.ModifierEntries[0].Name != "not" {
		modifierText = "." + jestFnCall.ModifierEntries[0].Name
	}
	if addNotModifier {
		modifierText += ".not"
	}
	return modifierText
}

var PreferEqualityMatcherRule = rule.Rule{
	Name: "jest/prefer-equality-matcher",
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

				left, right, negated, ok := parseStrictEqualityComparison(expectArgs[0])
				if !ok {
					return
				}

				matcherArgs := node.AsCallExpression().Arguments.Nodes
				if len(matcherArgs) == 0 {
					return
				}

				matcherValue, ok := isBooleanLiteral(matcherArgs[0])
				if !ok {
					return
				}

				matcherEntry := jestFnCall.MatcherEntry
				if matcherEntry.Node == nil || matcherEntry.Node.Parent == nil {
					return
				}

				hasNot := slices.Contains(jestFnCall.Modifiers, "not")
				modifierText := buildModifierText(jestFnCall, (negated != matcherValue) == hasNot)

				leftText := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, left, false)
				rightText := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, right, false)
				replaceComparison := rule.RuleFixReplace(ctx.SourceFile, expectArgs[0], leftText)
				replaceMatcherArg := rule.RuleFixReplace(ctx.SourceFile, matcherArgs[0], rightText)
				modifierRange := core.NewTextRange(expectCall.End(), matcherEntry.Node.Parent.End())

				suggestions := make([]rule.RuleSuggestion, len(suggestedEqualityMatchers))
				for i, equalityMatcher := range suggestedEqualityMatchers {
					suggestions[i] = rule.RuleSuggestion{
						Message: buildSuggestEqualityMatcherMessage(equalityMatcher),
						FixesArr: []rule.RuleFix{
							replaceComparison,
							rule.RuleFixReplaceRange(modifierRange, modifierText+"."+equalityMatcher),
							replaceMatcherArg,
						},
					}
				}

				ctx.ReportNodeWithSuggestions(
					matcherEntry.Node,
					buildUseEqualityMatcherErrorMessage(),
					suggestions...,
				)
			},
		}
	},
}
