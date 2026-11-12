package no_mismatched_map_key

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (a *analyzer) sameReference(left, right *ast.Node) bool {
	if !unicornutil.IsSameComparisonOperandWithKeyEvaluator(left, right, a.computedKeys.Eval) {
		return false
	}
	leftRoot, rightRoot := rootIdentifier(left), rootIdentifier(right)
	return leftRoot == nil || rightRoot == nil || a.sameBinding(leftRoot, rightRoot)
}

func (a *analyzer) sameReceiver(left, right *ast.Node) bool {
	return unicornutil.IsSameComparisonOperandWithKeyEvaluator(left, right, a.computedKeys.Eval) &&
		a.sameReceiverBindings(left, right)
}

// The structural comparison already matched property names. Check the
// bindings of identifiers that are read to obtain the receiver or an index.
func (a *analyzer) sameReceiverBindings(left, right *ast.Node) bool {
	left, right = unwrap(left), unwrap(right)
	if left == nil || right == nil {
		return true
	}
	if ast.IsIdentifier(left) && ast.IsIdentifier(right) {
		return a.sameBinding(left, right)
	}
	if !ast.IsAccessExpression(left) || !ast.IsAccessExpression(right) {
		return true
	}
	if ast.IsElementAccessExpression(left) && ast.IsElementAccessExpression(right) &&
		!a.sameReceiverBindings(left.AsElementAccessExpression().ArgumentExpression, right.AsElementAccessExpression().ArgumentExpression) {
		return false
	}
	return a.sameReceiverBindings(left.Expression(), right.Expression())
}

func rootIdentifier(node *ast.Node) *ast.Node {
	for node = unwrap(node); node != nil; node = unwrap(node.Expression()) {
		if ast.IsIdentifier(node) {
			return node
		}
		if !ast.IsAccessExpression(node) {
			return nil
		}
		// A chain used as a separate expression, such as (cache?.map),
		// is an ESTree ChainExpression. Its identifier is not a root.
		if ast.IsOptionalChain(node) &&
			(node.Parent == nil || !ast.IsOptionalChain(node.Parent) || node.Parent.Expression() != node) {
			return nil
		}
	}
	return nil
}

func (a *analyzer) sameBinding(left, right *ast.Node) bool {
	if left.Text() != right.Text() {
		return false
	}
	rightSymbol := a.ctx.Refs.ResolveInFile(right)
	if rightSymbol == nil && right.Parent != nil && ast.IsDeclarationBindingElement(right.Parent) {
		rightSymbol = utils.BindingNameSymbol(right)
	}
	return a.ctx.Refs.ResolveInFile(left) == rightSymbol
}

func (a *analyzer) writesReceiver(node, receiver *ast.Node) bool {
	var target *ast.Node
	switch node.Kind {
	case ast.KindVariableDeclaration:
		if node.Initializer() != nil {
			target = node.Name()
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) && !utils.IsDefaultValueInDestructuringAssignment(node) {
			target = binary.Left
		}
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken {
			target = prefix.Operand
		}
	case ast.KindPostfixUnaryExpression:
		target = node.AsPostfixUnaryExpression().Operand
	case ast.KindForInStatement, ast.KindForOfStatement:
		target = node.Initializer()
		if target != nil && ast.IsVariableDeclarationList(target) {
			return false
		}
	}
	return a.patternWritesReceiver(target, receiver)
}

func (a *analyzer) patternWritesReceiver(pattern, receiver *ast.Node) bool {
	pattern = unwrap(pattern)
	if pattern == nil {
		return false
	}
	if ast.IsIdentifier(pattern) || ast.IsAccessExpression(pattern) {
		return a.changesReceiver(receiver, pattern)
	}
	for _, element := range ast.GetElementsOfBindingOrAssignmentPattern(pattern) {
		if a.patternWritesReceiver(ast.GetTargetOfBindingOrAssignmentElement(element), receiver) {
			return true
		}
	}
	return false
}

func (a *analyzer) changesReceiver(receiver, target *ast.Node) bool {
	for node := unwrap(receiver); node != nil; node = unwrap(node.Expression()) {
		if a.sameReceiver(node, target) {
			return true
		}
		if !ast.IsAccessExpression(node) {
			break
		}
		if ast.IsElementAccessExpression(node) &&
			a.changesReceiver(node.AsElementAccessExpression().ArgumentExpression, target) {
			return true
		}
	}
	return false
}

var mapTypeOptions = unicornutil.TypeClassifierOptions{
	TargetTypeNames:            utils.NewSetFromItems("Map", "ReadonlyMap"),
	HeritageSymbolFlags:        ast.SymbolFlagsClass | ast.SymbolFlagsInterface,
	PreferKnownAnnotationNames: true,
}

func (a *analyzer) isKnownNonMap(node *ast.Node) bool {
	class := a.sourceReceiverType(node, map[*ast.Symbol]unicornutil.TypeClass{})
	if class == unicornutil.TypeUnknown && a.ctx.TypeChecker != nil && !ast.IsSourceFileJS(a.ctx.SourceFile) {
		class = unicornutil.ClassifyType(a.ctx, a.ctx.TypeChecker.GetTypeAtLocation(node), mapTypeOptions)
	}
	return class == unicornutil.TypeNonTarget
}

func (a *analyzer) sourceReceiverType(node *ast.Node, inferred map[*ast.Symbol]unicornutil.TypeClass) unicornutil.TypeClass {
	if node == nil {
		return unicornutil.TypeUnknown
	}
	node = ast.SkipOuterExpressions(node, ast.OEKExpressionTypePassthrough|ast.OEKNonNullAssertions|ast.OEKSatisfies)
	if ast.IsOptionalChain(node) {
		return unicornutil.TypeUnknown
	}
	classify := func(annotation *ast.Node) unicornutil.TypeClass {
		return unicornutil.ClassifyTypeAnnotation(a.ctx, annotation, mapTypeOptions)
	}
	switch node.Kind {
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		annotation := node.Type()
		if annotation != nil && (annotation.Kind == ast.KindAnyKeyword || annotation.Kind == ast.KindUnknownKeyword) {
			return a.sourceReceiverType(node.Expression(), inferred)
		}
		return classify(annotation)
	case ast.KindIdentifier:
		symbol := a.ctx.Refs.ResolveInFile(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			break
		}
		if class, known := inferred[symbol]; known {
			return class
		}
		declaration := symbol.Declarations[0]
		if annotation, optional := unicornutil.ValueTypeAnnotation(a.ctx, declaration); annotation != nil {
			class := classify(annotation)
			if class == unicornutil.TypeNonTarget || class == unicornutil.TypeTarget && !optional {
				return class
			}
		}
		if ast.IsVariableDeclaration(declaration) && ast.IsVarConst(declaration) {
			// An alias that recurs during inference is unknown.
			inferred[symbol] = unicornutil.TypeUnknown
			class := a.sourceReceiverType(declaration.Initializer(), inferred)
			inferred[symbol] = class
			return class
		}
	case ast.KindCallExpression:
		callee := utils.ESTreeCallCallee(node.Expression())
		symbol := a.ctx.Refs.ResolveInFile(callee)
		if symbol == nil || len(symbol.Declarations) != 1 {
			break
		}
		function, ok := unicornutil.UnmodifiedFunctionValue(a.ctx, symbol)
		if ok {
			if function != nil {
				if annotation := unicornutil.CallableReturnTypeAnnotation(a.ctx, function); annotation != nil {
					return classify(annotation)
				}
			}
			return classify(unicornutil.CallableReturnTypeAnnotation(a.ctx, symbol.Declarations[0]))
		}
	case ast.KindNewExpression:
		callee := utils.ESTreeCallCallee(node.Expression())
		if callee != nil && ast.IsIdentifier(callee) && callee.Text() == "Map" {
			return unicornutil.TypeTarget
		}
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return a.combineReceiverTypes(conditional.WhenTrue, conditional.WhenFalse, inferred)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if ast.IsLogicalOrCoalescingBinaryOperator(binary.OperatorToken.Kind) {
			return a.combineReceiverTypes(binary.Left, binary.Right, inferred)
		}
	}
	return unicornutil.TypeUnknown
}

func (a *analyzer) combineReceiverTypes(left, right *ast.Node, inferred map[*ast.Symbol]unicornutil.TypeClass) unicornutil.TypeClass {
	leftType := a.sourceReceiverType(left, inferred)
	if leftType != unicornutil.TypeUnknown && leftType == a.sourceReceiverType(right, inferred) {
		return leftType
	}
	return unicornutil.TypeUnknown
}
