package prefer_at

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func removeLength(ctx rule.RuleContext, length *ast.Node) (rule.RuleFix, bool) {
	span := utils.TrimNodeTextRange(ctx.SourceFile, length)
	if utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
		return rule.RuleFix{}, false
	}
	trailing := ctx.SourceFile.Text()[span.End():]
	end := span.End() + len(trailing) - len(strings.TrimLeftFunc(trailing, ecmascript.IsWhiteSpaceOrLineTerminator))
	return rule.RuleFixRemoveRange(span.WithEnd(end)), true
}

func indexFixes(ctx rule.RuleContext, node, length *ast.Node) []rule.RuleFix {
	source := ctx.SourceFile
	access := node.AsElementAccessExpression()
	var fixes []rule.RuleFix
	if length != nil {
		if isObviouslyUnsupportedReceiver(ctx, access.Expression) {
			return nil
		}
		fix, ok := removeLength(ctx, length)
		if !ok {
			return nil
		}
		fixes = append(fixes, fix)

		// Preserve formatting of parenthesized, commented, and non-decimal
		// operands. Only tighten a simple subtraction, e.g. "- 1" to "-1".
		index := utils.ESTreeRuntimeExpression(access.ArgumentExpression)
		if ast.IsBinaryExpression(index) && index.AsBinaryExpression().Left == length {
			binary := index.AsBinaryExpression()
			right := utils.ESTreeRuntimeExpression(binary.Right)
			if right == binary.Right && ast.IsNumericLiteral(right) {
				digits := utils.TrimmedNodeText(source, right)
				rightStart := utils.TrimNodeTextRange(source, right).Pos()
				minusEnd := binary.OperatorToken.End()
				space := source.Text()[minusEnd:rightStart]
				if digits != "" && strings.Trim(digits, "0123456789") == "" && ecmascript.StringTrim(space) == "" && minusEnd < rightStart {
					fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(minusEnd, rightStart)))
				}
			}
		}
	}

	open, _ := utils.TokenAtOrAfter(source, access.Expression.End())
	replacement := ".at("
	if ast.IsOptionalChainRoot(node) {
		open, _ = utils.TokenAtOrAfter(source, open.End)
		replacement = "at("
	}
	closeBracket, _ := utils.TokenAtOrAfter(source, access.ArgumentExpression.End())
	if open.Kind != ast.KindOpenBracketToken || closeBracket.Kind != ast.KindCloseBracketToken {
		return nil
	}
	fixes = append(fixes, rule.RuleFixReplaceRange(open.Range(), replacement), rule.RuleFixReplaceRange(closeBracket.Range(), ")"))
	if inUnparenthesizedConstructorCallee(node) {
		fixes = append(fixes, rule.RuleFixInsertBefore(source, node, "("), rule.RuleFixInsertAfter(node, ")"))
	}
	return fixes
}

func sliceFixes(ctx rule.RuleContext, call unicornutil.DotMethodCall, result *ast.Node, method unicornutil.DotMethodCall) []rule.RuleFix {
	source := ctx.SourceFile
	start := utils.TrimNodeTextRange(source, call.Call).Pos()
	if isObviouslyUnsupportedReceiver(ctx, call.Object) || utils.HasCommentInSpan(ctx.Comments.All(), start, result.End()) {
		return nil
	}
	fixes := []rule.RuleFix{rule.RuleFixReplace(source, call.Property, "at")}
	if args := call.Call.Arguments(); len(args) == 2 {
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(args[0].End(), call.Call.End()-1)))
	}
	if method.Call != nil {
		return append(fixes, unicornutil.RemoveMethodCallFixes(method)...)
	}
	return append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(result.Expression().End(), result.End())))
}

func lastFunctionFixes(ctx rule.RuleContext, node, receiver *ast.Node) []rule.RuleFix {
	if isObviouslyUnsupportedReceiver(ctx, receiver) {
		return nil
	}
	source := ctx.SourceFile
	callRange := utils.TrimNodeTextRange(source, node)
	receiverRange := utils.TrimNodeTextRange(source, receiver)
	comments := ctx.Comments.All()
	if utils.HasCommentInSpan(comments, callRange.Pos(), receiverRange.Pos()) ||
		utils.HasCommentInSpan(comments, receiverRange.End(), callRange.End()) {
		return nil
	}
	text := source.Text()[receiverRange.Pos():receiverRange.End()]
	if !ast.IsParenthesizedExpression(receiver) && unicornutil.ShouldAddParenthesesToMemberExpressionObjectAt(source, receiver, node) {
		text = "(" + text + ")"
	}
	text += ".at(-1)"
	if unicornutil.NeedsSemicolonBefore(source, node, text) {
		text = ";" + text
	}
	return append([]rule.RuleFix{rule.RuleFixReplaceRange(callRange, text)}, unicornutil.SpaceAroundKeywordFixes(source, node)...)
}
