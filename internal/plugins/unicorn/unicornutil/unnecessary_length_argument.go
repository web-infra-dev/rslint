// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type unnecessaryLengthReferenceIndexKey struct{}
type unnecessaryLengthStaticEvaluatorKey struct{}

// ReportUnnecessaryLengthArgument checks the second argument of an already
// matched two-argument slice/splice call. Callers own receiver restrictions.
func ReportUnnecessaryLengthArgument(ctx rule.RuleContext, call DotMethodCall, messageID, argumentName string) {
	raw := call.Call.Arguments()[1]
	argument := utils.ESTreeRuntimeExpression(raw)
	object := utils.ESTreeRuntimeExpression(call.Object)
	description := lengthOrInfinityDescription(ctx, argument, object)
	if description == "" {
		return
	}
	if lengthMember(argument) != nil &&
		(!isRepeatableReference(ctx, object) || hasSideEffect(call.Call.Arguments()[0], true)) {
		return
	}
	message := rule.RuleMessage{Id: messageID, Description: "Passing `" + description + "` as the `" + argumentName + "` argument is unnecessary.", Data: map[string]string{"description": description, "argumentName": argumentName}}
	ctx.ReportNodeWithDeferredFixes(argument, message, func() []rule.RuleFix {
		// Preserve comments before the separator and any existing trailing comma.
		comma, ok := utils.TokenBeforePosition(ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, raw).Pos())
		if !ok || comma.Kind != ast.KindCommaToken {
			return nil
		}
		return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(comma.Start, raw.End()))}
	})
}
func lengthOrInfinityDescription(ctx rule.RuleContext, argument, object *ast.Node) string {
	if ast.IsIdentifier(argument) && argument.Text() == "Infinity" &&
		isPristineGlobalReference(ctx, argument, "Infinity") {
		return "Infinity"
	}
	member := lengthMember(argument)
	if member == nil {
		return ""
	}
	receiver := utils.ESTreeRuntimeExpression(member.Expression)
	if member.Name().Text() == "POSITIVE_INFINITY" && !ast.IsOptionalChain(argument) &&
		ast.IsIdentifier(receiver) && receiver.Text() == "Number" &&
		isPristineGlobalReference(ctx, receiver, "Number") {
		return "Number.POSITIVE_INFINITY"
	}
	if member.Name().Text() != "length" || !sameStaticReference(ctx, object, receiver) {
		return ""
	}
	name := "…"
	if ast.IsIdentifier(object) {
		name = object.Text()
	}
	if ast.IsOptionalChain(argument) {
		return name + "?.length"
	}
	return name + ".length"
}

func lengthMember(node *ast.Node) *ast.PropertyAccessExpression {
	if node == nil || !ast.IsPropertyAccessExpression(node) {
		return nil
	}
	member := node.AsPropertyAccessExpression()
	if member == nil || member.Name() == nil || !ast.IsIdentifier(member.Name()) {
		return nil
	}
	return member
}

func isPristineGlobalReference(ctx rule.RuleContext, node *ast.Node, name string) bool {
	if node == nil || ctx.Refs == nil || !ctx.Globals.Access(name).IsDeclared() ||
		!ctx.Refs.IsGlobalReference(node) {
		return false
	}
	index := rule.CachedByFile(ctx, unnecessaryLengthReferenceIndexKey{}, func() *utils.ReferenceIndex {
		return utils.NewReferenceIndex(ctx.SourceFile, ctx.TypeChecker)
	})
	pristine := true
	index.ForEachReferenceByName(name, nil, func(reference *ast.Node) bool {
		if reference.Pos() >= node.Pos() {
			return true
		}
		if ctx.Refs.IsGlobalReference(reference) && utils.IsWriteReference(reference) {
			pristine = false
			return true
		}
		return false
	})
	return pristine
}

func sameStaticReference(ctx rule.RuleContext, left, right *ast.Node) bool {
	left = utils.SkipAssertionsAndParens(left)
	right = utils.SkipAssertionsAndParens(right)
	if left == nil || right == nil {
		return left == right
	}
	if ast.IsAccessExpression(left) && ast.IsAccessExpression(right) {
		evaluator := rule.CachedByFile(ctx, unnecessaryLengthStaticEvaluatorKey{}, func() *utils.StaticStringEvaluator {
			return utils.NewStaticStringEvaluatorWithReferenceResolver(
				ctx.TypeChecker, ctx.SourceFile, ctx.Refs,
			)
		})
		leftName, leftOK := evaluator.EvalAccessExpressionName(left)
		rightName, rightOK := evaluator.EvalAccessExpressionName(right)
		if !leftOK || !rightOK || leftName != rightName {
			return false
		}
		return sameStaticReference(
			ctx,
			utils.AccessExpressionObject(left),
			utils.AccessExpressionObject(right),
		)
	}
	return utils.IsSameReference(left, right, false)
}

func isRepeatableReference(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword:
		return true
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if accessHasGetter(ctx, node) {
			return false
		}
		return isRepeatableReference(ctx, utils.AccessExpressionObject(node))
	default:
		return false
	}
}

func accessHasGetter(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.TypeChecker == nil || node == nil {
		return false
	}
	var location *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		location = node.AsPropertyAccessExpression().Name()
	case ast.KindElementAccessExpression:
		location = node
	default:
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(location)
	if symbol == nil {
		return false
	}

	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindGetAccessor ||
			ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAccessor) {
			return true
		}
	}
	return false
}

// TODO: Extract this together with prefer_ternary.hasSideEffect into
// internal/utils once the shared contract includes configurable getter handling.
func hasSideEffect(node *ast.Node, considerGetters bool) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindFunctionDeclaration:
		// Function bodies are deferred until invocation.
		return false
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindAwaitExpression,
		ast.KindYieldExpression, ast.KindDeleteExpression, ast.KindPostfixUnaryExpression,
		ast.KindTaggedTemplateExpression:
		return true
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix != nil && (prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken) {
			return true
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil &&
			ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			return true
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if considerGetters {
			return true
		}
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return deferredMemberHasSideEffect(node, considerGetters)
	}
	return node.ForEachChild(func(child *ast.Node) bool {
		return hasSideEffect(child, considerGetters)
	})
}

func deferredMemberHasSideEffect(node *ast.Node, considerGetters bool) bool {
	for _, decorator := range node.Decorators() {
		if hasSideEffect(decorator, considerGetters) {
			return true
		}
	}
	if node.Kind == ast.KindConstructor {
		return false
	}
	return hasSideEffect(node.Name(), considerGetters)
}
