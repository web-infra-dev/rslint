package no_non_null_assertion

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// isAssignee matches typescript-eslint's isAssignee utility. In particular,
// member access is not itself transparent: in x!.y.z = value, x!.y is not the
// direct assignee and upstream still offers the optional-chain suggestion.
func isAssignee(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent

	if ast.IsAssignmentExpression(parent, false) {
		// tsgo uses BinaryExpression(=) for ESTree AssignmentPattern too.
		// Upstream isAssignee does not treat that wrapper as an assignee.
		if utils.IsDefaultValueInDestructuringAssignment(parent) {
			return false
		}
		return parent.AsBinaryExpression().Left == node
	}

	switch parent.Kind {
	case ast.KindDeleteExpression:
		return parent.AsDeleteExpression().Expression == node
	case ast.KindPostfixUnaryExpression:
		postfix := parent.AsPostfixUnaryExpression()
		return postfix.Operand == node &&
			(postfix.Operator == ast.KindPlusPlusToken || postfix.Operator == ast.KindMinusMinusToken)
	case ast.KindPrefixUnaryExpression:
		prefix := parent.AsPrefixUnaryExpression()
		return prefix.Operand == node &&
			(prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken)
	case ast.KindArrayLiteralExpression, ast.KindSpreadElement:
		return isAssignee(parent)
	case ast.KindSpreadAssignment:
		return parent.Parent != nil && isAssignee(parent.Parent)
	case ast.KindPropertyAssignment:
		property := parent.AsPropertyAssignment()
		return property.Initializer == node &&
			parent.Parent != nil &&
			parent.Parent.Kind == ast.KindObjectLiteralExpression &&
			isAssignee(parent.Parent)
	case ast.KindNonNullExpression,
		ast.KindAsExpression,
		ast.KindTypeAssertionExpression,
		ast.KindSatisfiesExpression,
		ast.KindParenthesizedExpression:
		return isAssignee(parent)
	}

	return false
}

func nonNullAssertionTokenRange(sourceFile *ast.SourceFile, node *ast.Node, expression *ast.Node) core.TextRange {
	end := node.End()
	start := end - 1
	text := sourceFile.Text()
	// Parsed non-null expressions end at their invariant ASCII `!` token. Keep
	// the scanner fallback for synthetic, reparsed, or otherwise unusual nodes.
	if start >= 0 && start >= expression.End() && end <= len(text) && text[start] == '!' {
		return core.NewTextRange(start, end)
	}
	s := scanner.GetScannerForSourceFile(sourceFile, expression.End())
	return s.TokenRange()
}

func singleOptionalChainSuggestion(suggestMsg rule.RuleMessage, fix rule.RuleFix) []rule.RuleSuggestion {
	return []rule.RuleSuggestion{{
		Message:  suggestMsg,
		FixesArr: []rule.RuleFix{fix},
	}}
}

func buildOptionalChainSuggestions(
	sourceFile *ast.SourceFile,
	node *ast.Node,
	suggestMsg rule.RuleMessage,
) []rule.RuleSuggestion {
	expression := node.AsNonNullExpression().Expression
	parent := node.Parent
	if parent == nil {
		return nil
	}

	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		if parent.Expression() != node || isAssignee(parent) {
			return nil
		}
		if parent.AsPropertyAccessExpression().QuestionDotToken != nil {
			return singleOptionalChainSuggestion(
				suggestMsg,
				rule.RuleFixRemoveRange(nonNullAssertionTokenRange(sourceFile, node, expression)),
			)
		}

		exclamRange := nonNullAssertionTokenRange(sourceFile, node, expression)
		// With no trivia, replacing ! with ? produces x?.y directly. This also
		// matches ESLint's merged suggestion edit and avoids scanning the dot.
		text := sourceFile.Text()
		if end := exclamRange.End(); end >= 0 && end < len(text) && text[end] == '.' {
			return singleOptionalChainSuggestion(
				suggestMsg,
				rule.RuleFixReplaceRange(exclamRange, "?"),
			)
		}

		// Preserve trivia between ! and . by editing both punctuators.
		dotScanner := scanner.GetScannerForSourceFile(sourceFile, expression.End())
		dotScanner.Scan()
		return []rule.RuleSuggestion{{
			Message: suggestMsg,
			FixesArr: []rule.RuleFix{
				rule.RuleFixRemoveRange(exclamRange),
				rule.RuleFixReplaceRange(dotScanner.TokenRange(), "?."),
			},
		}}
	case ast.KindElementAccessExpression:
		if parent.Expression() != node || isAssignee(parent) {
			return nil
		}
		if parent.AsElementAccessExpression().QuestionDotToken != nil {
			return singleOptionalChainSuggestion(
				suggestMsg,
				rule.RuleFixRemoveRange(nonNullAssertionTokenRange(sourceFile, node, expression)),
			)
		}
		return singleOptionalChainSuggestion(
			suggestMsg,
			rule.RuleFixReplaceRange(nonNullAssertionTokenRange(sourceFile, node, expression), "?."),
		)
	case ast.KindCallExpression:
		if parent.Expression() != node {
			return nil
		}
		if parent.AsCallExpression().QuestionDotToken != nil {
			return singleOptionalChainSuggestion(
				suggestMsg,
				rule.RuleFixRemoveRange(nonNullAssertionTokenRange(sourceFile, node, expression)),
			)
		}
		return singleOptionalChainSuggestion(
			suggestMsg,
			rule.RuleFixReplaceRange(nonNullAssertionTokenRange(sourceFile, node, expression), "?."),
		)
	default:
		return nil
	}
}

var NoNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name:   "no-non-null-assertion",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		msg := rule.RuleMessage{
			Id:          "noNonNull",
			Description: "Forbidden non-null assertion.",
		}
		suggestMsg := rule.RuleMessage{
			Id:          "suggestOptionalChain",
			Description: "Consider using the optional chain operator `?.` instead. This operator includes runtime checks, so it is safer than the compile-only non-null assertion operator.",
		}

		return rule.RuleListeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				ctx.ReportNodeWithDeferredSuggestions(node, msg, func() []rule.RuleSuggestion {
					return buildOptionalChainSuggestions(ctx.SourceFile, node, suggestMsg)
				})
			},
		}
	},
})
