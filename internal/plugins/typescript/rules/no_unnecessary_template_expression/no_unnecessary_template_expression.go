package no_unnecessary_template_expression

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
	for _, r := range str {
		return ecmascript.IsLineTerminator(r)
	}
	return false
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
		if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			return false
		}
	}
	return true
}

func endsWithUnescapedDollarSign(text string) bool {
	if !strings.HasSuffix(text, "$") {
		return false
	}

	backslashes := 0
	for i := len(text) - 2; i >= 0 && text[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 0
}

func escapeTemplateMetacharacters(text string) string {
	var output strings.Builder
	output.Grow(len(text))
	backslashes := 0
	for i := range len(text) {
		ch := text[i]
		if ch == '\\' {
			output.WriteByte(ch)
			backslashes++
			continue
		}

		if backslashes%2 == 0 && (ch == '`' || ch == '$' && i+1 < len(text) && text[i+1] == '{') {
			output.WriteByte('\\')
		}
		output.WriteByte(ch)
		backslashes = 0
	}
	return output.String()
}

func escapeTrailingDollarSign(text string) string {
	return text[:len(text)-1] + "\\$"
}

func interpolationValueNode(node *ast.Node) *ast.Node {
	if ast.IsLiteralTypeNode(node) {
		return node.AsLiteralTypeNode().Literal
	}
	return node
}

func templateHeadRawText(sourceFile *ast.SourceFile, node *ast.TemplateHeadNode) string {
	textRange := utils.TrimNodeTextRange(sourceFile, node)
	if textRange.Pos()+1 > textRange.End()-2 {
		return ""
	}
	return sourceFile.Text()[textRange.Pos()+1 : textRange.End()-2]
}

func templateLiteralRawText(sourceFile *ast.SourceFile, node *ast.TemplateMiddleOrTail) string {
	textRange := utils.TrimNodeTextRange(sourceFile, node)
	endOffset := 1
	if node.Kind == ast.KindTemplateMiddle {
		endOffset = 2
	}
	if textRange.Pos()+1 > textRange.End()-endOffset {
		return ""
	}
	return sourceFile.Text()[textRange.Pos()+1 : textRange.End()-endOffset]
}

func interpolationFixes(sourceFile *ast.SourceFile, interpolation *ast.Node, nextCharacterIsOpeningCurlyBrace bool) ([]rule.RuleFix, bool, bool) {
	node := interpolationValueNode(interpolation)
	textRange := utils.TrimNodeTextRange(sourceFile, node)
	sourceText := sourceFile.Text()
	if textRange.Pos() < 0 || textRange.End() > len(sourceText) || textRange.Pos() >= textRange.End() {
		return nil, nextCharacterIsOpeningCurlyBrace, false
	}

	switch node.Kind {
	case ast.KindStringLiteral:
		raw := sourceText[textRange.Pos():textRange.End()]
		if len(raw) < 2 {
			return nil, nextCharacterIsOpeningCurlyBrace, false
		}
		replacement := escapeTemplateMetacharacters(raw[1 : len(raw)-1])
		if nextCharacterIsOpeningCurlyBrace && endsWithUnescapedDollarSign(replacement) {
			replacement = escapeTrailingDollarSign(replacement)
		}
		if replacement != "" {
			nextCharacterIsOpeningCurlyBrace = strings.HasPrefix(replacement, "{")
		}
		return []rule.RuleFix{rule.RuleFixReplaceRange(textRange, replacement)}, nextCharacterIsOpeningCurlyBrace, true

	case ast.KindNoSubstitutionTemplateLiteral:
		rawText := sourceText[textRange.Pos()+1 : textRange.End()-1]
		fixes := []rule.RuleFix{
			rule.RuleFixRemoveRange(core.NewTextRange(textRange.Pos(), textRange.Pos()+1)),
			rule.RuleFixRemoveRange(core.NewTextRange(textRange.End()-1, textRange.End())),
		}
		if nextCharacterIsOpeningCurlyBrace && endsWithUnescapedDollarSign(rawText) {
			fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(textRange.End()-2, textRange.End()-2), "\\"))
		}
		if rawText != "" {
			nextCharacterIsOpeningCurlyBrace = strings.HasPrefix(rawText, "{")
		}
		return fixes, nextCharacterIsOpeningCurlyBrace, true

	case ast.KindTemplateExpression:
		template := node.AsTemplateExpression()
		fixes := []rule.RuleFix{
			rule.RuleFixRemoveRange(core.NewTextRange(textRange.Pos(), textRange.Pos()+1)),
			rule.RuleFixRemoveRange(core.NewTextRange(textRange.End()-1, textRange.End())),
		}
		lastSpan := template.TemplateSpans.Nodes[len(template.TemplateSpans.Nodes)-1].AsTemplateSpan()
		if nextCharacterIsOpeningCurlyBrace && endsWithUnescapedDollarSign(templateLiteralRawText(sourceFile, lastSpan.Literal)) {
			fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(textRange.End()-2, textRange.End()-2), "\\"))
		}
		return fixes, nextCharacterIsOpeningCurlyBrace, true
	}

	if isFixableIdentifier(node) {
		return nil, false, true
	}

	replacement, ok := utils.GetStaticExpressionValue(node)
	if !ok {
		return nil, false, false
	}
	if node.Kind == ast.KindRegularExpressionLiteral {
		replacement = strings.ReplaceAll(replacement, "\\", "\\\\")
	}
	replacement = escapeTemplateMetacharacters(replacement)
	if nextCharacterIsOpeningCurlyBrace && endsWithUnescapedDollarSign(replacement) {
		replacement = escapeTrailingDollarSign(replacement)
	}
	if replacement != "" {
		nextCharacterIsOpeningCurlyBrace = strings.HasPrefix(replacement, "{")
	}
	return []rule.RuleFix{rule.RuleFixReplaceRange(textRange, replacement)}, nextCharacterIsOpeningCurlyBrace, true
}

var NoUnnecessaryTemplateExpressionRule = rule.CreateRule(rule.Rule{
	Name:             "no-unnecessary-template-expression",
	Schema:           rule.EmptyArraySchema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		reportSingleInterpolation := func(templateNode *ast.Node, spanExpr *ast.Node, spanLiteral *ast.Node) {
			expressionRange := utils.TrimNodeTextRange(ctx.SourceFile, spanExpr)
			replacement := ctx.SourceFile.Text()[expressionRange.Pos():expressionRange.End()]
			if !utils.IsStrongPrecedenceNode(spanExpr) && typescriptutil.IsWeakPrecedenceParent(templateNode) {
				replacement = "(" + replacement + ")"
			}
			ctx.ReportRangeWithFixes(
				core.NewTextRange(spanExpr.Pos()-2, spanLiteral.Pos()+1),
				buildNoUnnecessaryTemplateExpressionMessage(),
				rule.RuleFixReplace(ctx.SourceFile, templateNode, replacement),
			)
		}

		isUnnecessaryValueInterpolation := func(expression *ast.Node, prevQuasiEnd int, nextQuasiLiteral *ast.TemplateMiddleOrTail) bool {
			if utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(prevQuasiEnd, nextQuasiLiteral.Pos())) || utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(nextQuasiLiteral.Pos(), utils.TrimNodeTextRange(ctx.SourceFile, nextQuasiLiteral).Pos())) {
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
			return len(templateSpans.Nodes) == 1 && head.AsTemplateHead().Text == "" && firstSpanLiteral.Text() == "" && !utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(head.End(), firstSpanLiteral.Pos())) && !utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(firstSpanLiteral.Pos(), utils.TrimNodeTextRange(ctx.SourceFile, firstSpanLiteral).Pos()))
		}

		isEnumMemberType := func(t *checker.Type) bool {
			return utils.TypeRecurser(t, func(t *checker.Type) bool {
				symbol := checker.Type_symbol(t)
				return symbol != nil && symbol.ValueDeclaration != nil && ast.IsEnumMember(symbol.ValueDeclaration)
			})
		}

		type interpolationInfo struct {
			expression           *ast.Node
			literal              *ast.TemplateMiddleOrTail
			previousQuasiEnd     int
			previousQuasiRawText string
			nextQuasiRawText     string
		}

		checkTemplateSpans := func(templateSpans *ast.NodeList, head *ast.TemplateHeadNode) {
			infos := make([]interpolationInfo, 0, len(templateSpans.Nodes))
			for i := range len(templateSpans.Nodes) {
				span := templateSpans.Nodes[i]
				var prevQuasiEnd int
				var prevQuasiRawText string
				if i == 0 {
					prevQuasiEnd = head.End()
					prevQuasiRawText = templateHeadRawText(ctx.SourceFile, head)
				} else {
					var previousLiteral *ast.TemplateMiddleOrTail
					if templateSpans.Nodes[i-1].Kind == ast.KindTemplateSpan {
						previousLiteral = templateSpans.Nodes[i-1].AsTemplateSpan().Literal
					} else {
						previousLiteral = templateSpans.Nodes[i-1].AsTemplateLiteralTypeSpan().Literal
					}
					prevQuasiEnd = previousLiteral.End()
					prevQuasiRawText = templateLiteralRawText(ctx.SourceFile, previousLiteral)
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
				infos = append(infos, interpolationInfo{
					expression:           expr,
					literal:              literal,
					previousQuasiEnd:     prevQuasiEnd,
					previousQuasiRawText: prevQuasiRawText,
					nextQuasiRawText:     templateLiteralRawText(ctx.SourceFile, literal),
				})
			}

			nextCharacterIsOpeningCurlyBrace := false
			for i := len(infos) - 1; i >= 0; i-- {
				info := infos[i]
				if info.nextQuasiRawText != "" {
					nextCharacterIsOpeningCurlyBrace = strings.HasPrefix(info.nextQuasiRawText, "{")
				}

				expressionFixes, nextIsCurly, ok := interpolationFixes(ctx.SourceFile, info.expression, nextCharacterIsOpeningCurlyBrace)
				nextCharacterIsOpeningCurlyBrace = nextIsCurly
				reportRange := core.NewTextRange(info.previousQuasiEnd-2, utils.TrimNodeTextRange(ctx.SourceFile, info.literal).Pos()+1)
				if !ok {
					ctx.ReportRange(reportRange, buildNoUnnecessaryTemplateExpressionMessage())
					continue
				}

				expressionRange := utils.TrimNodeTextRange(ctx.SourceFile, info.expression)
				fixes := []rule.RuleFix{
					rule.RuleFixRemoveRange(core.NewTextRange(reportRange.Pos(), expressionRange.Pos())),
					rule.RuleFixRemoveRange(core.NewTextRange(expressionRange.End(), reportRange.End())),
				}
				fixes = append(fixes, expressionFixes...)
				if nextCharacterIsOpeningCurlyBrace && endsWithUnescapedDollarSign(info.previousQuasiRawText) {
					fixes = append(fixes, rule.RuleFixReplaceRange(
						core.NewTextRange(info.previousQuasiEnd-3, info.previousQuasiEnd-2),
						"\\$",
					))
				}
				ctx.ReportRangeWithFixes(reportRange, buildNoUnnecessaryTemplateExpressionMessage(), fixes...)
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
						reportSingleInterpolation(node, firstSpan.Expression, firstSpan.Literal)
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
						reportSingleInterpolation(node, firstSpan.Type, firstSpan.Literal)
						return
					}
				}

				checkTemplateSpans(expr.TemplateSpans, expr.Head)
			},
		}
	},
})
