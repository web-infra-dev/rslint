package prefer_math_min_max

import (
	"math/big"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func skipOperand(ctx *rule.RuleContext, evaluator *utils.StaticStringEvaluator, operand *ast.Node) bool {
	node := utils.ESTreeRuntimeExpression(operand)
	if node == nil {
		return false
	}
	if isDateConstruction(node) {
		return true
	}
	// Check the outer authored cast, if any. Non-null assertions do not
	// contribute a type annotation.
	cast := ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKNonNullAssertions)
	if ast.IsAsExpression(cast) || ast.IsTypeAssertion(cast) {
		annotation := ast.SkipTypeParentheses(cast.Type())
		boxedNumber := annotation.Kind == ast.KindTypeReference &&
			unicornutil.NodeMatchesPath(annotation.AsTypeReferenceNode().TypeName, "Number")
		if annotation.Kind != ast.KindNumberKeyword && !boxedNumber {
			return true
		}
	}

	binding := ctx.Refs.ResolveInFile(unwrapOperand(node))
	if binding == nil {
		return isBigInt(ctx, evaluator, operand)
	}
	for _, definition := range binding.Declarations {
		// ESTree's variable definition points to the entire declarator even
		// when the identifier is part of a destructuring pattern.
		declaration := ast.GetRootDeclaration(definition)
		switch declaration.Kind {
		case ast.KindVariableDeclaration:
			if ast.IsCatchClauseVariableDeclarationOrBindingElement(declaration) {
				continue
			}
		case ast.KindParameter:
			// The type of a destructured or rest parameter belongs to the
			// pattern, while a destructured name can have its own default.
			declaration = definition
			if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
				continue
			}
		default:
			continue
		}
		annotation := declaration.Type()
		if annotation != nil && !utils.IsJSDocSyntaxNode(annotation) && ast.SkipTypeParentheses(annotation).Kind != ast.KindNumberKeyword {
			return true
		}
		initializer := utils.ESTreeRuntimeExpression(declaration.Initializer())
		if nonNumberLiteral(initializer) || isBigInt(ctx, evaluator, initializer) ||
			(declaration.Kind == ast.KindVariableDeclaration && isDateConstruction(initializer)) {
			return true
		}
	}
	return isBigInt(ctx, evaluator, operand)
}

func isDateConstruction(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindNewExpression &&
		unicornutil.NodeMatchesPath(node.Expression(), "Date")
}

func nonNumberLiteral(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindNoSubstitutionTemplateLiteral ||
		(utils.IsESTreeLiteralKind(node.Kind) && node.Kind != ast.KindNumericLiteral))
}

func isBigInt(ctx *rule.RuleContext, evaluator *utils.StaticStringEvaluator, node *ast.Node) bool {
	if node == nil {
		return false
	}
	if utils.IsESTreeLiteralKind(node.Kind) || node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return ast.IsBigIntLiteral(node)
	}
	if ctx.TypeChecker != nil && typeIncludesBigInt(utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)) {
		return true
	}
	if bigIntFromSyntax(ctx, node, make(map[*ast.Node]bigIntAnnotationClass)) {
		return true
	}
	value, ok := evaluator.EvalValueIfNoSideEffects(node)
	_, isBigInt := value.(*big.Int)
	return ok && isBigInt && !unicornutil.HasJSDocTypeReference(*ctx, node, make(map[*ast.Symbol]bool))
}

func typeIncludesBigInt(value *checker.Type) bool {
	if utils.IsTypeFlagSet(value, checker.TypeFlagsBigIntLike) {
		return true
	}
	return (utils.IsUnionType(value) || utils.IsIntersectionType(value)) &&
		slices.ContainsFunc(value.Types(), typeIncludesBigInt)
}

type bigIntAnnotationClass uint8

const (
	annotationOther bigIntAnnotationClass = iota
	annotationBigInt
	annotationAny
)

// Type information is optional. Recognize BigInt from the operand's syntax
// and authored annotations even in a project-free file.
func bigIntFromSyntax(ctx *rule.RuleContext, node *ast.Node, bindings map[*ast.Node]bigIntAnnotationClass) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindBigIntLiteral:
		return true
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		annotation := node.Type()
		return classifyBigIntAnnotation(ctx, annotation, bindings) == annotationBigInt ||
			((annotation.Kind == ast.KindAnyKeyword || annotation.Kind == ast.KindUnknownKeyword) &&
				bigIntFromSyntax(ctx, node.Expression(), bindings))
	case ast.KindSatisfiesExpression:
		return classifyBigIntAnnotation(ctx, node.Type(), bindings) == annotationBigInt ||
			bigIntFromSyntax(ctx, node.Expression(), bindings)
	case ast.KindNonNullExpression:
		return bigIntFromSyntax(ctx, node.Expression(), bindings)
	case ast.KindIdentifier:
		symbol := ctx.Refs.ResolveInFile(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if result, ok := bindings[declaration]; ok {
			return result == annotationBigInt
		}
		annotation, _ := unicornutil.ValueTypeAnnotation(*ctx, declaration)
		if annotation != nil {
			return classifyBigIntAnnotation(ctx, annotation, bindings) == annotationBigInt
		}
		if declaration.Kind != ast.KindVariableDeclaration || !ast.IsVarConst(declaration) {
			return false
		}
		// Break cycles and reuse the result for repeated aliases in this operand.
		bindings[declaration] = annotationOther
		if bigIntFromSyntax(ctx, declaration.Initializer(), bindings) {
			bindings[declaration] = annotationBigInt
			return true
		}
		return false
	case ast.KindCallExpression:
		return bigIntCall(ctx, node, bindings)
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return bigIntFromSyntax(ctx, conditional.WhenTrue, bindings) || bigIntFromSyntax(ctx, conditional.WhenFalse, bindings)
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		switch unary.Operator {
		case ast.KindMinusToken, ast.KindTildeToken, ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return bigIntFromSyntax(ctx, unary.Operand, bindings)
		}
	case ast.KindPostfixUnaryExpression:
		return bigIntFromSyntax(ctx, node.AsPostfixUnaryExpression().Operand, bindings)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindCommaToken, ast.KindEqualsToken:
			return bigIntFromSyntax(ctx, binary.Right, bindings)
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return bigIntFromSyntax(ctx, binary.Left, bindings) || bigIntFromSyntax(ctx, binary.Right, bindings)
		case ast.KindPlusToken, ast.KindMinusToken, ast.KindAsteriskToken, ast.KindSlashToken, ast.KindPercentToken,
			ast.KindAsteriskAsteriskToken, ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken,
			ast.KindBarToken, ast.KindCaretToken, ast.KindAmpersandToken:
			return bigIntFromSyntax(ctx, binary.Left, bindings) && bigIntFromSyntax(ctx, binary.Right, bindings)
		}
	}
	return false
}

func classifyBigIntAnnotation(ctx *rule.RuleContext, annotation *ast.Node, bindings map[*ast.Node]bigIntAnnotationClass) bigIntAnnotationClass {
	if annotation == nil {
		return annotationOther
	}
	annotation = ast.SkipTypeParentheses(annotation)
	var parts []*ast.Node
	switch annotation.Kind {
	case ast.KindBigIntKeyword:
		return annotationBigInt
	case ast.KindAnyKeyword, ast.KindJSDocAllType:
		return annotationAny
	case ast.KindLiteralType:
		literal := annotation.AsLiteralTypeNode().Literal
		if ast.IsPrefixUnaryExpression(literal) {
			literal = literal.AsPrefixUnaryExpression().Operand
		}
		if ast.IsBigIntLiteral(literal) {
			return annotationBigInt
		}
		return annotationOther
	case ast.KindJSDocNullableType, ast.KindJSDocNonNullableType, ast.KindJSDocOptionalType:
		return classifyBigIntAnnotation(ctx, annotation.Type(), bindings)
	case ast.KindTypeReference:
		symbol := ctx.Refs.ResolveInFileWithMeaning(annotation.AsTypeReferenceNode().TypeName, ast.SymbolFlagsType)
		if symbol == nil {
			return annotationOther
		}
		declaration := ast.GetDeclarationOfKind(symbol, ast.KindTypeAliasDeclaration)
		var referencedType *ast.Node
		if declaration != nil {
			referencedType = declaration.Type()
		} else {
			declaration = ast.GetDeclarationOfKind(symbol, ast.KindTypeParameter)
			// A generic alias's constraint alone cannot determine its type
			// argument. Follow constraints from the enclosing scope.
			if declaration == nil || ast.IsTypeAliasDeclaration(declaration.Parent) {
				return annotationOther
			}
			referencedType = declaration.AsTypeParameterDeclaration().Constraint
		}
		if result, ok := bindings[declaration]; ok {
			return result
		}
		bindings[declaration] = annotationOther
		result := classifyBigIntAnnotation(ctx, referencedType, bindings)
		bindings[declaration] = result
		return result
	case ast.KindUnionType:
		parts = annotation.AsUnionTypeNode().Types.Nodes
	case ast.KindIntersectionType:
		parts = annotation.AsIntersectionTypeNode().Types.Nodes
	default:
		return annotationOther
	}

	// An explicit any overrides the other members, even through an alias.
	// An unknown member contributes no evidence and is neutral in an intersection.
	result := annotationOther
	for _, part := range parts {
		switch classifyBigIntAnnotation(ctx, part, bindings) {
		case annotationAny:
			return annotationAny
		case annotationBigInt:
			result = annotationBigInt
		}
	}
	return result
}

func bigIntCall(ctx *rule.RuleContext, node *ast.Node, bindings map[*ast.Node]bigIntAnnotationClass) bool {
	if ast.IsOptionalChain(node) {
		return false
	}
	callee := utils.ESTreeCallCallee(node.Expression())
	if unicornutil.NodeMatchesPath(callee, "BigInt") {
		return true
	}
	if call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Methods: []string{"asIntN", "asUintN"},
	}); ok && unicornutil.NodeMatchesPath(call.Object, "BigInt") {
		return true
	}
	symbol := ctx.Refs.ResolveInFile(callee)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	var function *ast.Node
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		function = declaration
	case ast.KindVariableDeclaration, ast.KindParameter:
		// An explicit TypeScript binding type takes precedence over JSDoc,
		// even when only the checker can resolve its return type.
		if bindingType := declaration.Type(); bindingType != nil && !utils.IsJSDocSyntaxNode(bindingType) &&
			!ast.IsFunctionTypeNode(ast.SkipTypeParentheses(bindingType)) {
			return false
		}
		initializer := utils.ESTreeRuntimeExpression(declaration.Initializer())
		if initializer != nil {
			initializer = ast.SkipOuterExpressions(initializer, ast.OEKParentheses|ast.OEKSatisfies|ast.OEKNonNullAssertions)
			if ast.IsFunctionExpressionOrArrowFunction(initializer) {
				function = initializer
			}
		}
	default:
		return false
	}
	if function != nil && ast.GetFunctionFlags(function)&ast.FunctionFlagsAsyncGenerator != 0 {
		return false
	}
	annotation := unicornutil.CallableReturnTypeAnnotation(*ctx, declaration)
	if annotation == nil && function != nil && ast.IsVariableDeclaration(declaration) {
		// Only infer a variable's return type from its initializer when the
		// binding has no explicit type of its own.
		if bindingType, _ := unicornutil.ValueTypeAnnotation(*ctx, declaration); bindingType == nil {
			annotation = unicornutil.CallableReturnTypeAnnotation(*ctx, function)
		}
	}
	return classifyBigIntAnnotation(ctx, annotation, bindings) == annotationBigInt &&
		!slices.ContainsFunc(ctx.Refs.References(symbol), utils.IsWriteReference)
}
