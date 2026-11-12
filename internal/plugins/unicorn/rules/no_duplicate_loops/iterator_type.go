package no_duplicate_loops

import (
	"maps"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var iteratorTypes = unicornutil.TypeClassifierOptions{
	TargetTypeNames: utils.NewSetFromItems(
		"ArrayIterator", "Generator", "IterableIterator", "Iterator", "IteratorObject",
		"MapIterator", "RegExpStringIterator", "SetIterator", "StringIterator",
	),
	HeritageSymbolFlags:      ast.SymbolFlagsClass | ast.SymbolFlagsInterface,
	AllowNullishInMixedUnion: true,
}

const annotationWrappers = ast.OEKParentheses | ast.OEKNonNullAssertions | ast.OEKSatisfies

func isKnownIteratorType(ctx *rule.RuleContext, expression *ast.Node) bool {
	switch classifyIteratorSourceType(ctx, expression, iteratorTypes) {
	case unicornutil.TypeTarget:
		options := iteratorTypesAt(ctx, expression)
		return options.TargetTypeNames == iteratorTypes.TargetTypeNames ||
			classifyIteratorSourceType(ctx, expression, options) == unicornutil.TypeTarget
	case unicornutil.TypeNonTarget:
		return false
	}

	// Use the TypeScript project for inferred and imported types. In JavaScript
	// rely on source evidence, as Unicorn does with its default parser.
	if ctx.TypeChecker == nil || ast.IsInJSFile(expression) {
		return false
	}
	t := ctx.TypeChecker.GetTypeAtLocation(expression)
	if unicornutil.ClassifyType(*ctx, t, iteratorTypes) != unicornutil.TypeTarget {
		return false
	}
	options := iteratorTypesAt(ctx, expression)
	return (options.TargetTypeNames == iteratorTypes.TargetTypeNames ||
		unicornutil.ClassifyType(*ctx, t, options) == unicornutil.TypeTarget) && !isArray(ctx, expression)
}

func isArray(ctx *rule.RuleContext, expression *ast.Node) bool {
	return unicornutil.IsArrayWithOptions(*ctx, expression, unicornutil.ArrayReceiverOptions{
		CheckArrayProducers: true,
		ReadJSDoc:           true,
	})
}

func iteratorTypesAt(ctx *rule.RuleContext, expression *ast.Node) unicornutil.TypeClassifierOptions {
	options := iteratorTypes
	for name := range iteratorTypes.TargetTypeNames.Keys() {
		if !ctx.Refs.IsNameDefinedInFileWithMeaning(expression, name, ast.SymbolFlagsValue|ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) {
			continue
		}
		if options.TargetTypeNames == iteratorTypes.TargetTypeNames {
			options.TargetTypeNames = &utils.Set[string]{M: maps.Clone(iteratorTypes.TargetTypeNames.Keys())}
		}
		options.TargetTypeNames.Delete(name)
	}
	return options
}

func classifyIteratorSourceType(ctx *rule.RuleContext, expression *ast.Node, options unicornutil.TypeClassifierOptions) unicornutil.TypeClass {
	node := ast.SkipOuterExpressions(expression, annotationWrappers|ast.OEKAssignments|ast.OEKComma)
	var annotation *ast.Node
	switch node.Kind {
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return classifyIteratorAlternatives(ctx, conditional.WhenTrue, conditional.WhenFalse, options)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if ast.IsLogicalOrCoalescingBinaryOperator(binary.OperatorToken.Kind) {
			return classifyIteratorAlternatives(ctx, binary.Left, binary.Right, options)
		}
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		if utils.IsJSDocTypeCastWrapper(node) {
			return classifyIteratorSourceType(ctx, node.Expression(), options)
		}
		annotation = node.Type()
		if annotation != nil && (annotation.Kind == ast.KindAnyKeyword || annotation.Kind == ast.KindUnknownKeyword) {
			return classifyIteratorSourceType(ctx, node.Expression(), options)
		}
	case ast.KindIdentifier:
		symbol := ctx.Refs.ResolveInFile(node)
		if symbol != nil && len(symbol.Declarations) == 1 {
			declaration := symbol.Declarations[0]
			if declaration.Kind == ast.KindParameter || declaration.Kind == ast.KindVariableDeclaration {
				annotation, _ = unicornutil.ValueTypeAnnotation(*ctx, declaration)
			}
		}
	case ast.KindCallExpression:
		annotation = iteratorReturnAnnotation(ctx, utils.ESTreeCallCallee(node.Expression()))
	}
	if annotation == nil {
		return unicornutil.TypeUnknown
	}
	class := unicornutil.ClassifyTypeAnnotation(*ctx, annotation, options)
	if class == unicornutil.TypeTarget && isArray(ctx, node) {
		return unicornutil.TypeNonTarget
	}
	return class
}

func classifyIteratorAlternatives(ctx *rule.RuleContext, left, right *ast.Node, options unicornutil.TypeClassifierOptions) unicornutil.TypeClass {
	leftType := classifyIteratorSourceType(ctx, left, options)
	rightType := classifyIteratorSourceType(ctx, right, options)
	switch {
	case leftType == unicornutil.TypeNullish:
		return rightType
	case rightType == unicornutil.TypeNullish || leftType == rightType:
		return leftType
	default:
		return unicornutil.TypeUnknown
	}
}

func iteratorReturnAnnotation(ctx *rule.RuleContext, callee *ast.Node) *ast.Node {
	symbol := ctx.Refs.ResolveInFile(callee)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	var function *ast.Node
	switch declaration.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		function = declaration
	case ast.KindVariableDeclaration, ast.KindParameter:
		if declaration.Type() != nil && !utils.IsJSDocSyntaxNode(declaration.Type()) {
			return nil
		}
		function = utils.ESTreeRuntimeExpression(declaration.Initializer())
		if function != nil {
			function = ast.SkipOuterExpressions(function, annotationWrappers)
		}
	default:
		return nil
	}
	if !bindingIsUnchanged(ctx, symbol) {
		return nil
	}
	if function != nil && ast.IsFunctionLike(function) {
		if ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) ||
			function.BodyData().AsteriskToken != nil || len(function.TypeParameters()) != 0 {
			return nil
		}
		if annotation := unicornutil.CallableReturnTypeAnnotation(*ctx, function); annotation != nil {
			return annotation
		}
	}
	if declaration != function {
		return unicornutil.CallableReturnTypeAnnotation(*ctx, declaration)
	}
	return nil
}
