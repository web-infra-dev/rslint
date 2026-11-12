package prefer_math_min_max

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func mathCallFixes(ctx *rule.RuleContext, evaluator *utils.StaticStringEvaluator, node *ast.Node, span core.TextRange, left, right *ast.Node, method string) []rule.RuleFix {
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

	leftSpan := left.Loc.WithPos(scanner.GetTokenPosOfNode(left, ctx.SourceFile, false))
	rightSpan := right.Loc.WithPos(scanner.GetTokenPosOfNode(right, ctx.SourceFile, false))
	for _, comment := range utils.CommentsInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
		inLeft := leftSpan.Pos() <= comment.Pos() && comment.End() <= leftSpan.End()
		inRight := rightSpan.Pos() <= comment.Pos() && comment.End() <= rightSpan.End()
		if !inLeft && !inRight {
			return nil
		}
	}

	fixes := unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)
	call := "Math." + method + "(" + argumentText(ctx, left, leftSpan) + ", " + argumentText(ctx, right, rightSpan) + ")"
	return append(fixes, rule.RuleFixReplaceRange(span, call))
}

func argumentText(ctx *rule.RuleContext, operand *ast.Node, span core.TextRange) string {
	source := ctx.SourceFile.Text()
	expression := utils.ESTreeRuntimeExpression(operand)
	if expression != operand {
		inner := expression.Loc.WithPos(scanner.GetTokenPosOfNode(expression, ctx.SourceFile, false))
		comments := ctx.Comments.All()
		if utils.HasCommentInSpan(comments, span.Pos(), inner.Pos()) ||
			utils.HasCommentInSpan(comments, inner.End(), span.End()) {
			return source[span.Pos():span.End()]
		}
		span = inner
	}
	text := source[span.Pos():span.End()]
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
