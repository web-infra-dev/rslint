package prefer_math_min_max

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func mathCallFixes(ctx *rule.RuleContext, evaluator *utils.StaticStringEvaluator, node, left, right *ast.Node, method string) []rule.RuleFix {
	// A local Math binding or a disabled global makes the suggested call unsafe.
	if !ctx.Globals.Access("Math").IsDeclared() || !ctx.Refs.IsGlobalNameReference(node, "Math", ast.SymbolFlagsValue) {
		return nil
	}

	// Template interpolation and tags can also execute user code. Treat tags
	// as calls even when the static evaluator recognizes the tag.
	if evaluator.HasSideEffect(left, false) || evaluator.HasSideEffect(right, false) ||
		hasTemplateCoercion(left) || hasTemplateCoercion(right) {
		return nil
	}

	span := utils.TrimNodeTextRange(ctx.SourceFile, node)
	leftSpan := utils.TrimNodeTextRange(ctx.SourceFile, left)
	rightSpan := utils.TrimNodeTextRange(ctx.SourceFile, right)
	for _, comment := range utils.CommentsInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
		inLeft := leftSpan.Pos() <= comment.Pos() && comment.End() <= leftSpan.End()
		inRight := rightSpan.Pos() <= comment.Pos() && comment.End() <= rightSpan.End()
		if !inLeft && !inRight {
			return nil
		}
	}

	fixes := unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)
	call := "Math." + method + "(" + argumentText(ctx, left) + ", " + argumentText(ctx, right) + ")"
	return append(fixes, rule.RuleFixReplaceRange(span, call))
}

func argumentText(ctx *rule.RuleContext, operand *ast.Node) string {
	expression := utils.ESTreeRuntimeExpression(operand)
	if expression != operand {
		outer := utils.TrimNodeTextRange(ctx.SourceFile, operand)
		inner := utils.TrimNodeTextRange(ctx.SourceFile, expression)
		comments := ctx.Comments.All()
		if utils.HasCommentInSpan(comments, outer.Pos(), inner.Pos()) ||
			utils.HasCommentInSpan(comments, inner.End(), outer.End()) {
			return utils.TrimmedNodeText(ctx.SourceFile, operand)
		}
	}
	text := utils.TrimmedNodeText(ctx.SourceFile, expression)
	if ast.IsCommaExpression(expression) {
		return "(" + text + ")"
	}
	return text
}

func hasTemplateCoercion(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || ast.IsTypeNode(node) {
		return false
	}
	if ast.IsFunctionLike(node) {
		// A computed name is evaluated now. The function body runs on invocation.
		return hasTemplateCoercion(node.Name())
	}
	return node.Kind == ast.KindTemplateExpression || node.Kind == ast.KindTaggedTemplateExpression ||
		node.ForEachChild(hasTemplateCoercion)
}
