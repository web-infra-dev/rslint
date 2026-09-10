package prefer_to_contain

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_to_contain"
)

func isExplicitNaN(node *ast.Node) bool {
	node = testFramework.FollowTypeAssertionChain(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text == "NaN"
	}
	var receiver *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		name := property.Name()
		if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "NaN" {
			return false
		}
		receiver = property.Expression
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		key := ast.SkipParentheses(element.ArgumentExpression)
		if key == nil || key.Kind != ast.KindStringLiteral || key.AsStringLiteral().Text != "NaN" {
			return false
		}
		receiver = element.Expression
	default:
		return false
	}
	receiver = ast.SkipParentheses(receiver)
	if receiver == nil || receiver.Kind != ast.KindIdentifier {
		return false
	}
	name := receiver.AsIdentifier().Text
	return name == "Number" || name == "globalThis"
}

func isKnownStringRegexpDifference(receiver, item *ast.Node) bool {
	receiver = testFramework.FollowTypeAssertionChain(receiver)
	item = testFramework.FollowTypeAssertionChain(item)
	return receiver != nil && item != nil &&
		(receiver.Kind == ast.KindStringLiteral || receiver.Kind == ast.KindNoSubstitutionTemplateLiteral) &&
		item.Kind == ast.KindRegularExpressionLiteral
}

func isSafeToMove(node *ast.Node) bool {
	node = testFramework.FollowTypeAssertionChain(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindSuperKeyword,
		ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		return unary != nil && unary.Operator != ast.KindPlusPlusToken && unary.Operator != ast.KindMinusMinusToken &&
			isSafeToMove(unary.Operand)
	case ast.KindArrayLiteralExpression:
		array := node.AsArrayLiteralExpression()
		if array == nil {
			return false
		}
		for _, element := range array.Elements.Nodes {
			if element.Kind == ast.KindSpreadElement || !isSafeToMove(element) {
				return false
			}
		}
		return true
	case ast.KindObjectLiteralExpression:
		object := node.AsObjectLiteralExpression()
		if object == nil {
			return false
		}
		for _, property := range object.Properties.Nodes {
			switch property.Kind {
			case ast.KindPropertyAssignment:
				assignment := property.AsPropertyAssignment()
				if assignment == nil || !isSafePropertyName(assignment.Name()) || !isSafeToMove(assignment.Initializer) {
					return false
				}
			case ast.KindShorthandPropertyAssignment:
				shorthand := property.AsShorthandPropertyAssignment()
				if shorthand == nil || shorthand.ObjectAssignmentInitializer != nil || !isSafeToMove(shorthand.Name()) {
					return false
				}
			default:
				return false
			}
		}
		return true
	default:
		return false
	}
}

func isSafePropertyName(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind != ast.KindComputedPropertyName {
		return true
	}
	return isSafeToMove(node.AsComputedPropertyName().Expression)
}

var PreferToContainRule = shared.NewRule(shared.Config{
	Name: "rstest/prefer-to-contain",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{Parse: func(node *ast.Node) *shared.ExpectCall {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil ||
				parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
				parsed.Head == nil ||
				(parsed.Entry != rstestUtils.RstestExpectEntryCall && parsed.Entry != rstestUtils.RstestExpectEntrySoft) ||
				parsed.MatcherEntry == nil ||
				len(parsed.Matchers) == 0 ||
				parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall ||
				testFramework.IsComputedIdentifierAccessor(parsed.MatcherEntry.Node) {
				return nil
			}

			for _, modifier := range parsed.Modifiers {
				if modifier == "resolves" || modifier == "rejects" {
					return nil
				}
			}

			matcherCall := testFramework.InvokedAccessorCall(parsed.MatcherEntry)
			if matcherCall == nil {
				return nil
			}
			return &shared.ExpectCall{
				HeadCall: parsed.Head, MatcherCall: matcherCall, MatcherEntry: *parsed.MatcherEntry,
				Matcher: parsed.Matcher, Modifiers: parsed.ModifierEntries,
			}
		}}
	},
	IgnoreMatch: func(_ rule.RuleContext, matched shared.Match) bool {
		// Array.prototype.includes uses SameValueZero, while Rstest's current
		// Vitest matcher layer delegates array containment to indexOf. Rewriting
		// an explicit NaN would therefore reverse a passing assertion.
		return isExplicitNaN(matched.Item) || isKnownStringRegexpDifference(matched.Receiver, matched.Item)
	},
	BuildFixes: func(ctx rule.RuleContext, matched shared.Match) []rule.RuleFix {
		if !isSafeToMove(matched.Item) {
			return nil
		}
		fixes := shared.BuildFixes(ctx, matched)
		if len(fixes) == 0 {
			return nil
		}
		if typeArguments, ok := testFramework.CallTypeArgumentListRange(ctx.SourceFile, matched.Expect.MatcherCall); ok {
			if utils.HasCommentInSpan(ctx.Comments.All(), typeArguments.Pos(), typeArguments.End()) {
				return nil
			}
			fixes = append(fixes, rule.RuleFixRemoveRange(typeArguments))
		}
		return fixes
	},
})
