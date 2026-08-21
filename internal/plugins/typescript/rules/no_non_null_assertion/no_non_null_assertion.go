package no_non_null_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// isAssignmentTarget checks if a node is used as the left side of an assignment.
func isAssignmentTarget(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if ast.IsAssignmentExpression(parent, true) {
		return parent.AsBinaryExpression().Left == node
	}
	if parent.Kind == ast.KindPropertyAccessExpression || parent.Kind == ast.KindElementAccessExpression {
		return isAssignmentTarget(parent)
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
		if parent.Expression() != node || isAssignmentTarget(parent) {
			return nil
		}
		if parent.AsPropertyAccessExpression().QuestionDotToken != nil {
			return singleOptionalChainSuggestion(
				suggestMsg,
				rule.RuleFixRemoveRange(nonNullAssertionTokenRange(sourceFile, node, expression)),
			)
		}

		// Dot notation (x!.y) removes ! and replaces . with ?.; scan only
		// when the consumer actually requests suggestions.
		dotScanner := scanner.GetScannerForSourceFile(sourceFile, expression.End())
		dotScanner.Scan()
		exclamRange := nonNullAssertionTokenRange(sourceFile, node, expression)
		return []rule.RuleSuggestion{{
			Message: suggestMsg,
			FixesArr: []rule.RuleFix{
				rule.RuleFixRemoveRange(exclamRange),
				rule.RuleFixReplaceRange(dotScanner.TokenRange(), "?."),
			},
		}}
	case ast.KindElementAccessExpression:
		if parent.Expression() != node || isAssignmentTarget(parent) {
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
		if parent.Expression() != node || isAssignmentTarget(parent) {
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
