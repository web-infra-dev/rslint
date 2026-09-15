package prefer_as_const

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var preferConstAssertionMessage = rule.RuleMessage{
	Id:          "preferConstAssertion",
	Description: "Expected a `const` instead of a literal type assertion.",
}

var variableConstAssertionMessage = rule.RuleMessage{
	Id:          "variableConstAssertion",
	Description: "Expected a `const` assertion instead of a literal type annotation.",
}

var variableSuggestMessage = rule.RuleMessage{
	Id:          "variableSuggest",
	Description: "You should use `as const` instead of type annotation.",
}

const (
	constAssertionText = "const"
	asConstSuffixText  = " as const"
)

func trimmedLiteralRange(sourceText string, node *ast.Node) (core.TextRange, bool) {
	start := scanner.SkipTrivia(sourceText, node.Pos())
	end := node.End()
	if start < 0 || start > end || end > len(sourceText) {
		return core.TextRange{}, false
	}
	return core.NewTextRange(start, end), true
}

func isComparableLiteral(node *ast.Node) bool {
	kind := node.Kind
	// ESTree represents booleans as Literal nodes, while ts-go's literal-token
	// range excludes keyword literals. Null remains excluded because ESTree
	// represents a null type annotation as TSNullKeyword, not TSLiteralType.
	return kind >= ast.KindFirstLiteralToken && kind <= ast.KindLastLiteralToken ||
		kind == ast.KindTrueKeyword || kind == ast.KindFalseKeyword
}

func matchingLiteralRange(sourceText string, valueNode *ast.Node, typeNode *ast.Node) (core.TextRange, bool) {
	literalNode := typeNode.AsLiteralTypeNode().Literal
	if literalNode == nil || !isComparableLiteral(literalNode) ||
		literalNode.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return core.TextRange{}, false
	}

	valueRange, ok := trimmedLiteralRange(sourceText, valueNode)
	if !ok {
		return core.TextRange{}, false
	}
	literalRange, ok := trimmedLiteralRange(sourceText, literalNode)
	if !ok || valueRange.End()-valueRange.Pos() != literalRange.End()-literalRange.Pos() ||
		sourceText[valueRange.Pos():valueRange.End()] != sourceText[literalRange.Pos():literalRange.End()] {
		return core.TextRange{}, false
	}
	return literalRange, true
}

// TypeScript starts a parsed type node immediately after its annotation colon;
// trivia between the colon and the first type token belongs to the type node.
func typeAnnotationRange(sourceText string, typeNode *ast.Node) (core.TextRange, bool) {
	colonStart := typeNode.Pos() - 1
	if colonStart < 0 || typeNode.End() > len(sourceText) || sourceText[colonStart] != ':' {
		return core.TextRange{}, false
	}
	return core.NewTextRange(colonStart, typeNode.End()), true
}

func compareLiteralTypes(ctx *rule.RuleContext, sourceText string, valueNode *ast.Node, typeNode *ast.Node, canFix bool) {
	literalRange, matches := matchingLiteralRange(sourceText, valueNode, typeNode)
	if !matches {
		return
	}

	if canFix {
		ctx.ReportRangeWithDeferredFixes(literalRange, preferConstAssertionMessage, func() []rule.RuleFix {
			return []rule.RuleFix{rule.RuleFixReplaceRange(literalRange, constAssertionText)}
		})
		return
	}

	ctx.ReportRangeWithDeferredSuggestions(literalRange, variableConstAssertionMessage, func() []rule.RuleSuggestion {
		annotationRange, ok := typeAnnotationRange(sourceText, typeNode)
		if !ok {
			return nil
		}
		return []rule.RuleSuggestion{
			{
				Message: variableSuggestMessage,
				FixesArr: []rule.RuleFix{
					rule.RuleFixReplaceRange(annotationRange, ""),
					rule.RuleFixInsertAfter(valueNode, asConstSuffixText),
				},
			},
		}
	})
}

var PreferAsConstRule = rule.CreateRule(rule.Rule{
	Name:   "prefer-as-const",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		compareTypes := func(valueNode *ast.Node, typeNode *ast.Node, canFix bool) {
			if valueNode == nil || typeNode == nil ||
				!isComparableLiteral(valueNode) || typeNode.Kind != ast.KindLiteralType {
				return
			}
			compareLiteralTypes(&ctx, ctx.SourceFile.Text(), valueNode, typeNode, canFix)
		}
		return rule.RuleListeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				declaration := node.AsPropertyDeclaration()
				compareTypes(declaration.Initializer, declaration.Type, false)
			},
			ast.KindAsExpression: func(node *ast.Node) {
				expression := node.AsAsExpression()
				compareTypes(expression.Expression, expression.Type, true)
			},
			ast.KindTypeAssertionExpression: func(node *ast.Node) {
				expression := node.AsTypeAssertion()
				compareTypes(expression.Expression, expression.Type, true)
			},
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				compareTypes(declaration.Initializer, declaration.Type, false)
			},
		}
	},
})
