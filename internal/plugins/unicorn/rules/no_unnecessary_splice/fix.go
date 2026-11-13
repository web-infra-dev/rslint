package no_unnecessary_splice

import (
	"math"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
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
		return hasSpliceSideEffect(ctx, effects, arg)
	}) {
		return nil
	}
	if (replacement == "pop" || replacement == "push" || replacement == "empty" && len(call.Call.Arguments()) == 2) &&
		hasSpliceSideEffect(ctx, effects, call.Object) {
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
	if hasSpliceSideEffect(ctx, effects, call.Object) {
		return nil
	}
	if args := call.Call.Arguments(); len(args) > 0 &&
		(hasSpliceSideEffect(ctx, effects, args[0]) || !hasSafeIndexConversion(effects, args[0])) {
		return nil
	}
	if precedesDirective(statement) {
		return []rule.RuleFix{rule.RuleFixReplaceRange(span, ";")}
	}
	// Retain a boundary even if a separate fix also removes the next statement.
	if previous, ok := utils.TokenBeforePosition(source, span.Pos()); ok &&
		unicornutil.NeedsSemicolonAfter(source, previous, "(", false) {
		return []rule.RuleFix{rule.RuleFixReplaceRange(span, ";")}
	}
	return []rule.RuleFix{rule.RuleFixRemoveRange(unicornutil.StatementRemovalRange(source, span))}
}

// The last removed call before a string must keep the directive prologue
// closed, even when earlier calls in the same statement list are removed.
func precedesDirective(statement *ast.Node) bool {
	container := statement.Parent
	if container.Kind != ast.KindSourceFile && !ast.IsFunctionLikeDeclaration(container.Parent) {
		return false
	}
	statements := container.Statements()
	index := ast.IndexOfNode(statements, statement)
	return index >= 0 && index+1 < len(statements) && ast.IsPrologueDirective(statements[index+1])
}

// Objects can run conversion methods and BigInt or Symbol indices throw.
// Only remove a call when the index can be converted without those effects.
func hasSafeIndexConversion(effects *utils.StaticStringEvaluator, node *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, expressionWrappers)
	switch node.Kind {
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	case ast.KindPrefixUnaryExpression:
		return !math.IsNaN(staticNumber(node))
	case ast.KindIdentifier:
		// A stable binding's initializer has already been evaluated. Folding a
		// compound index directly could miss evaluation, as in `(+1n, 0)`.
	default:
		return false
	}
	value, known := effects.EvalControlFlowValue(node)
	if !known {
		return false
	}
	switch value.(type) {
	case jsnum.Number, string, bool:
		return true
	}
	return false
}

func hasSpliceSideEffect(ctx rule.RuleContext, effects *utils.StaticStringEvaluator, node *ast.Node) bool {
	return effects.HasSideEffect(node, false) || hasImplicitSpliceEffect(ctx, effects, node)
}

// HasSideEffect handles explicit evaluation. Check getter reads and iterator
// execution as well, since either can mutate the array during argument evaluation.
func hasImplicitSpliceEffect(ctx rule.RuleContext, effects *utils.StaticStringEvaluator, node *ast.Node) bool {
	if node == nil || ast.IsPartOfTypeNode(node) {
		return false
	}
	node = ast.SkipOuterExpressions(node, expressionWrappers)
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return false
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		// Only a computed name is evaluated when creating a method.
		return hasImplicitSpliceEffect(ctx, effects, node.Name())
	case ast.KindPropertyDeclaration:
		if !ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
			return hasImplicitSpliceEffect(ctx, effects, node.Name())
		}
	case ast.KindSpreadAssignment, ast.KindJsxSpreadAttribute:
		return true
	case ast.KindSpreadElement:
		iterable := ast.SkipOuterExpressions(node.Expression(), expressionWrappers)
		// A fresh, dense array or a string literal has no custom iterator of
		// its own. HasSideEffect already checks for iterator prototype writes.
		if !ast.IsStringLiteralLike(iterable) &&
			(!ast.IsArrayLiteralExpression(iterable) || slices.ContainsFunc(iterable.Elements(), ast.IsOmittedExpression)) {
			return true
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if readsKnownGetter(ctx, effects, node) {
			return true
		}
	}
	return node.ForEachChild(func(child *ast.Node) bool { return hasImplicitSpliceEffect(ctx, effects, child) })
}

func readsKnownGetter(ctx rule.RuleContext, effects *utils.StaticStringEvaluator, node *ast.Node) bool {
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
