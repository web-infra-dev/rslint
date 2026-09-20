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
type unnecessaryLengthGlobalObjectWritesKey struct{}

var unnecessaryLengthGlobalObjectNames = map[string]bool{
	"global":     true,
	"globalThis": true,
	"self":       true,
	"window":     true,
}

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
		(!isRepeatableReference(ctx, object) || !isSideEffectFreeArgument(call.Call.Arguments()[0])) {
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
	return pristine && !hasEarlierGlobalObjectPropertyWrite(ctx, name, node.Pos())
}

func hasEarlierGlobalObjectPropertyWrite(ctx rule.RuleContext, name string, before int) bool {
	firstWrites := rule.CachedByFile(ctx, unnecessaryLengthGlobalObjectWritesKey{}, func() map[string]int {
		writes := map[string]int{}
		evaluator := rule.CachedByFile(ctx, unnecessaryLengthStaticEvaluatorKey{}, func() *utils.StaticStringEvaluator {
			return utils.NewStaticStringEvaluatorWithReferenceResolver(
				ctx.TypeChecker, ctx.SourceFile, ctx.Refs,
			)
		})
		var visit func(*ast.Node)
		visit = func(node *ast.Node) {
			if ast.IsAccessExpression(node) && isGlobalObjectPropertyWrite(ctx, node) {
				property, ok := evaluator.EvalAccessExpressionName(node)
				if ok {
					if previous, exists := writes[property]; !exists || node.Pos() < previous {
						writes[property] = node.Pos()
					}
				}
			}
			node.ForEachChild(func(child *ast.Node) bool {
				visit(child)
				return false
			})
		}
		visit(ctx.SourceFile.AsNode())
		return writes
	})
	position, ok := firstWrites[name]
	return ok && position < before
}

func isGlobalObjectPropertyWrite(ctx rule.RuleContext, node *ast.Node) bool {
	if !utils.IsWriteReference(node) {
		parent := node.Parent
		if parent == nil || parent.Kind != ast.KindDeleteExpression ||
			parent.AsDeleteExpression().Expression != node {
			return false
		}
	}
	root := utils.SkipAssertionsAndParens(utils.AccessExpressionObject(node))
	if !ast.IsIdentifier(root) || !unnecessaryLengthGlobalObjectNames[root.Text()] ||
		ctx.Refs == nil || !ctx.Globals.Access(root.Text()).IsDeclared() {
		return false
	}
	return ctx.Refs.IsGlobalReference(root)
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
		// Without type information we cannot distinguish a data property from a
		// getter. Re-evaluating a member chain can therefore change which object
		// receives the splice, so only checker-backed member paths are repeatable.
		if ctx.TypeChecker == nil || accessHasGetter(ctx, node) {
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

func isSideEffectFreeArgument(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword,
		ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindRegularExpressionLiteral,
		ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		switch prefix.Operator {
		case ast.KindPlusToken, ast.KindMinusToken, ast.KindExclamationToken,
			ast.KindTildeToken, ast.KindTypeOfKeyword, ast.KindVoidKeyword:
			return isSideEffectFreeArgument(prefix.Operand)
		}
	}
	return false
}
