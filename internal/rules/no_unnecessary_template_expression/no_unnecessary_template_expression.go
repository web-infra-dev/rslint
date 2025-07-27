package no_unnecessary_template_expression

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

func buildNoUnnecessaryTemplateExpressionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noUnnecessaryTemplateExpression",
		Description: "Template literal expression is unnecessary and can be simplified.",
	}
}

func isUnderlyingTypeString(t *checker.Type) bool {
	return utils.Every(utils.UnionTypeParts(t), func(t *checker.Type) bool {
		return utils.Some(utils.IntersectionTypeParts(t), func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike)
		})
	})
}

func isAnyLiteral(node *ast.Node) bool {
	return ast.IsLiteralExpression(node) || ast.IsBooleanLiteral(node) || node.Kind == ast.KindNullKeyword
}

func isFixableIdentifier(node *ast.Node) bool {
	if ast.IsIdentifier(node) {
		name := node.AsIdentifier().Text
		return name == "undefined" || name == "Infinity" || name == "NaN"
	}
	return node.Kind == ast.KindUndefinedKeyword
}

func startsWithNewline(str string) bool {
	return strings.HasPrefix(str, "\n") || strings.HasPrefix(str, "\r\n")
}

func isWhitespace(str string) bool {
	// allow empty string too since we went to allow
	// `      ${''}
	// `;
	//
	// in addition to
	// `${'        '}
	// `;

	for _, r := range str {
		if !utils.IsStrWhiteSpace(r) {
			return false
		}
	}
	return true
}

var NoUnnecessaryTemplateExpressionRule = rule.Rule{
	Name: "no-unnecessary-template-expression",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		reportSingleInterpolation := func(spanExpr *ast.Node, spanLiteral *ast.Node) {
			ctx.ReportRange(core.NewTextRange(spanExpr.Pos()-2, spanLiteral.Pos()+1), buildNoUnnecessaryTemplateExpressionMessage())
		}

		isUnnecessaryValueInterpolation := func(expression *ast.Node, prevQuasiEnd int, nextQuasiLiteral *ast.TemplateMiddleOrTail) bool {
			// Check for comments in the entire template expression span
			// From the end of the previous quasi (which includes ${) to the start of the next quasi
			templateExprStart := prevQuasiEnd - 2 // Position of `${`
			templateExprEnd := nextQuasiLiteral.Pos() // Position of the template literal after `}`
			
			// Check broadly for comments in the entire template expression
			if utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(templateExprStart, templateExprEnd)) {
				return false
			}
			
			// Also check around the expression itself for any comments that might be adjacent
			exprStart := expression.Pos()
			exprEnd := expression.End()
			if utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(exprStart, exprEnd + 15)) { // Check after expression
				return false
			}
			if utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(exprStart - 15, exprStart)) { // Check before expression  
				return false
			}

			if ast.IsLiteralTypeNode(expression) {
				expression = expression.AsLiteralTypeNode().Literal
			}

			if isFixableIdentifier(expression) {
				return true
			}

			if ast.IsStringLiteralLike(expression) {
				var raw string
				if nextQuasiLiteral.Kind == ast.KindTemplateMiddle {
					raw = nextQuasiLiteral.AsTemplateMiddle().RawText
				} else {
					raw = nextQuasiLiteral.AsTemplateTail().RawText
				}

				// allow trailing whitespace literal
				return !startsWithNewline(raw) || !isWhitespace(expression.Text())
			}

			return isAnyLiteral(expression) || ast.IsTemplateExpression(expression)
		}

		isTrivialInterpolation := func(templateSpans *ast.NodeList, head *ast.TemplateHeadNode, firstSpanLiteral *ast.Node) bool {
			if len(templateSpans.Nodes) != 1 || head.AsTemplateHead().Text != "" || firstSpanLiteral.Text() != "" {
				return false
			}
			// Check for comments in the template expression ${...}
			templateExprStart := head.End() - 2 // Position of `${`
			templateExprEnd := firstSpanLiteral.Pos() // Position of the template literal after `}`
			
			// Check the main range
			if utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(templateExprStart, templateExprEnd)) {
				return false
			}
			
			// Also check a broader range to catch edge cases
			if utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(templateExprStart, templateExprEnd + 15)) {
				return false
			}
			
			return true
		}

		isEnumMemberType := func(t *checker.Type) bool {
			return utils.TypeRecurser(t, func(t *checker.Type) bool {
				symbol := checker.Type_symbol(t)
				return symbol != nil && symbol.ValueDeclaration != nil && ast.IsEnumMember(symbol.ValueDeclaration)
			})
		}

		checkTemplateSpans := func(templateSpans *ast.NodeList, head *ast.TemplateHeadNode) {
			for i := len(templateSpans.Nodes) - 1; i >= 0; i-- {
				span := templateSpans.Nodes[i]
				var prevQuasiEnd int
				if i == 0 {
					prevQuasiEnd = head.End()
				} else {
					prevQuasiEnd = templateSpans.Nodes[i-1].End()
				}

				var expr *ast.Node
				var literal *ast.TemplateMiddleOrTail
				if span.Kind == ast.KindTemplateSpan {
					s := span.AsTemplateSpan()
					expr = s.Expression
					literal = s.Literal
				} else {
					s := span.AsTemplateLiteralTypeSpan()
					expr = s.Type
					literal = s.Literal
				}

				if !isUnnecessaryValueInterpolation(expr, prevQuasiEnd, literal) {
					continue
				}

				// TODO(port): implement fixes
				ctx.ReportRange(core.NewTextRange(prevQuasiEnd-2, utils.TrimNodeTextRange(ctx.SourceFile, literal).Pos()+1), buildNoUnnecessaryTemplateExpressionMessage())
			}
		}

		return rule.RuleListeners{
			ast.KindTemplateExpression: func(node *ast.Node) {
				if ast.IsTaggedTemplateExpression(node.Parent) {
					return
				}

				expr := node.AsTemplateExpression()
				firstSpan := expr.TemplateSpans.Nodes[0].AsTemplateSpan()

				if isTrivialInterpolation(expr.TemplateSpans, expr.Head, firstSpan.Literal) {
					constraintType, _ := utils.GetConstraintInfo(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(firstSpan.Expression))

					if constraintType != nil && isUnderlyingTypeString(constraintType) {
						reportSingleInterpolation(firstSpan.Expression, firstSpan.Literal)
						return
					}
				}

				checkTemplateSpans(expr.TemplateSpans, expr.Head)
			},
			ast.KindTemplateLiteralType: func(node *ast.Node) {
				expr := node.AsTemplateLiteralTypeNode()
				firstSpan := expr.TemplateSpans.Nodes[0].AsTemplateLiteralTypeSpan()

				if isTrivialInterpolation(expr.TemplateSpans, expr.Head, firstSpan.Literal) {
					constraintType, isTypeParameter := utils.GetConstraintInfo(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(firstSpan.Type))

					if constraintType != nil && !isTypeParameter && isUnderlyingTypeString(constraintType) && !isEnumMemberType(constraintType) {
						reportSingleInterpolation(firstSpan.Type, firstSpan.Literal)
						return
					}
				}

				checkTemplateSpans(expr.TemplateSpans, expr.Head)
			},
		}
	},
}
