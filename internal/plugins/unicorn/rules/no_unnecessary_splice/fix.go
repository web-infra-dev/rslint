package no_unnecessary_splice

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func spliceFixes(ctx rule.RuleContext, call unicornutil.DotMethodCall, replacement string, effects *utils.StaticStringEvaluator) []rule.RuleFix {
	if replacement == "no-op" {
		return removeNoOp(ctx, call, effects)
	}
	source := ctx.SourceFile
	span := utils.TrimNodeTextRange(source, call.Call)
	if utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
		return nil
	}
	if replacement == "push" && slices.ContainsFunc(call.Call.Arguments()[2:], func(arg *ast.Node) bool {
		// splice reads length before evaluating the inserted values.
		return effects.HasSideEffect(arg, false) || readsKnownGetter(ctx, effects, arg)
	}) {
		return nil
	}
	if (replacement == "pop" || replacement == "push" || replacement == "empty" && len(call.Call.Arguments()) == 2) &&
		(effects.HasSideEffect(call.Object, false) || readsKnownGetter(ctx, effects, call.Object)) {
		// The length argument evaluates the receiver a second time.
		return nil
	}

	receiver := utils.TrimmedNodeText(source, call.Object)
	if !ast.IsParenthesizedExpression(call.Object) && unicornutil.ShouldAddParenthesesToMemberExpressionObjectAt(source, call.Object, call.Call) {
		receiver = "(" + receiver + ")"
	}

	var text string
	switch replacement {
	case "empty":
		text = receiver + ".length = 0"
	case "shift", "pop":
		text = receiver + "." + replacement + "()"
	case "unshift", "push":
		args := call.Call.Arguments()[2:]
		argTexts := make([]string, len(args))
		for index, arg := range args {
			// Keep source parentheses, especially around comma expressions.
			argTexts[index] = utils.TrimmedNodeText(source, arg)
		}
		text = receiver + "." + replacement + "(" + strings.Join(argTexts, ", ") + ")"
	}
	return []rule.RuleFix{rule.RuleFixReplaceRange(span, utils.SafeReplacementText(source, call.Call, text))}
}

func removeNoOp(ctx rule.RuleContext, call unicornutil.DotMethodCall, effects *utils.StaticStringEvaluator) []rule.RuleFix {
	statement := utils.ESTreeParent(call.Call)
	if !ast.IsExpressionStatement(statement) {
		return nil
	}
	container := statement.Parent
	if container.Kind != ast.KindSourceFile &&
		(container.Kind != ast.KindBlock || container.Parent.Kind == ast.KindClassStaticBlockDeclaration) {
		return nil
	}
	source := ctx.SourceFile
	span := utils.TrimNodeTextRange(source, statement)
	comments := utils.CommentsInSpan(ctx.Comments.All(), span.Pos(), len(source.Text()))
	if len(comments) > 0 && (comments[0].Pos() < span.End() || utils.IsSameLine(source, span.End(), comments[0].Pos())) {
		return nil
	}
	if effects.HasSideEffect(call.Object, false) || readsKnownGetter(ctx, effects, call.Object) {
		return nil
	}
	if args := call.Call.Arguments(); len(args) > 0 &&
		(effects.HasSideEffect(args[0], false) || readsKnownGetter(ctx, effects, args[0]) || hasObservableIndexConversion(args[0])) {
		return nil
	}
	// Retain a boundary even if a separate fix also removes the next statement.
	if previous, ok := utils.TokenBeforePosition(source, span.Pos()); ok &&
		unicornutil.NeedsSemicolonAfter(source, previous, "(", false) {
		return []rule.RuleFix{rule.RuleFixReplaceRange(span, ";")}
	}
	return []rule.RuleFix{rule.RuleFixRemoveRange(unicornutil.StatementRemovalRange(source, span))}
}

// An object used as a splice index can run user code when converted to a
// number. BigInt conversion throws even when deleteCount is zero.
func hasObservableIndexConversion(node *ast.Node) bool {
	switch ast.SkipOuterExpressions(node, expressionWrappers).Kind {
	case ast.KindObjectLiteralExpression, ast.KindBigIntLiteral:
		return true
	}
	return false
}

func readsKnownGetter(ctx rule.RuleContext, effects *utils.StaticStringEvaluator, node *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, expressionWrappers)
	if !ast.IsAccessExpression(node) {
		return false
	}
	if readsKnownGetter(ctx, effects, node.Expression()) ||
		ast.IsElementAccessExpression(node) && readsKnownGetter(ctx, effects, node.AsElementAccessExpression().ArgumentExpression) {
		return true
	}
	if ctx.TypeChecker == nil {
		return readsObjectLiteralGetter(effects, node)
	}
	base := ast.GetLeftmostExpression(node, true)
	for ast.IsParenthesizedExpression(base) {
		base = ast.GetLeftmostExpression(ast.SkipOuterExpressions(base, expressionWrappers), true)
	}
	// Skip full property resolution for an unresolved JavaScript receiver.
	// The checker otherwise searches for spelling suggestions for the name.
	if ast.IsInJSFile(node) && ast.IsIdentifier(base) && ctx.Refs.ResolveInFile(base) == nil &&
		ctx.TypeChecker.ResolveName(base.Text(), base, ast.SymbolFlagsValue, false) == nil {
		return false
	}
	return unicornutil.AccessHasGetter(ctx.TypeChecker, node)
}

func readsObjectLiteralGetter(effects *utils.StaticStringEvaluator, access *ast.Node) bool {
	object := ast.SkipOuterExpressions(access.Expression(), expressionWrappers)
	if initializer, ok := effects.ResolveIdentifierInitializer(object); ok {
		object = ast.SkipOuterExpressions(initializer, expressionWrappers)
	}
	if !ast.IsObjectLiteralExpression(object) || object.Symbol() == nil {
		return false
	}
	key, ok := effects.EvalAccessExpressionName(access)
	if !ok {
		return false
	}
	member := object.Symbol().Members[key]
	return member != nil && member.Flags&ast.SymbolFlagsGetAccessor != 0
}
