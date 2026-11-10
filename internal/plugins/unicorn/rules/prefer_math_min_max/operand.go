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
		annotation := cast.Type()
		boxedNumber := annotation.Kind == ast.KindTypeReference &&
			isNamedIdentifier(annotation.AsTypeReferenceNode().TypeName, "Number")
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
		if annotation != nil && !utils.IsJSDocSyntaxNode(annotation) && annotation.Kind != ast.KindNumberKeyword {
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
		isNamedIdentifier(utils.ESTreeRuntimeExpression(node.Expression()), "Date")
}

func nonNumberLiteral(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindNoSubstitutionTemplateLiteral ||
		(utils.IsESTreeLiteralKind(node.Kind) && node.Kind != ast.KindNumericLiteral))
}

func isNamedIdentifier(node *ast.Node, name string) bool {
	return node != nil && ast.IsIdentifier(node) && node.Text() == name
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
	if bigIntFromSyntax(ctx, node, make(map[*ast.Node]bool)) {
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

// Type information is optional. Recognize BigInt from the operand's syntax
// and authored annotations even in a project-free file.
func bigIntFromSyntax(ctx *rule.RuleContext, node *ast.Node, bindings map[*ast.Node]bool) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindBigIntLiteral:
		return true
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		annotation := node.Type()
		return bigIntAnnotation(ctx, annotation, bindings) ||
			((annotation.Kind == ast.KindAnyKeyword || annotation.Kind == ast.KindUnknownKeyword) &&
				bigIntFromSyntax(ctx, node.Expression(), bindings))
	case ast.KindSatisfiesExpression:
		return bigIntAnnotation(ctx, node.Type(), bindings) || bigIntFromSyntax(ctx, node.Expression(), bindings)
	case ast.KindNonNullExpression:
		return bigIntFromSyntax(ctx, node.Expression(), bindings)
	case ast.KindIdentifier:
		symbol := ctx.Refs.ResolveInFile(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if result, ok := bindings[declaration]; ok {
			return result
		}
		annotation, _ := unicornutil.ValueTypeAnnotation(*ctx, declaration)
		if annotation != nil {
			return bigIntAnnotation(ctx, annotation, bindings)
		}
		if declaration.Kind != ast.KindVariableDeclaration || !ast.IsVarConst(declaration.Parent) {
			return false
		}
		// Break cycles and reuse the result for repeated aliases in this operand.
		bindings[declaration] = false
		result := bigIntFromSyntax(ctx, declaration.Initializer(), bindings)
		bindings[declaration] = result
		return result
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

func bigIntAnnotation(ctx *rule.RuleContext, annotation *ast.Node, bindings map[*ast.Node]bool) bool {
	if annotation == nil {
		return false
	}
	switch annotation.Kind {
	case ast.KindBigIntKeyword:
		return true
	case ast.KindLiteralType:
		literal := annotation.AsLiteralTypeNode().Literal
		if ast.IsPrefixUnaryExpression(literal) {
			literal = literal.AsPrefixUnaryExpression().Operand
		}
		return ast.IsBigIntLiteral(literal)
	case ast.KindParenthesizedType, ast.KindJSDocNullableType, ast.KindJSDocNonNullableType, ast.KindJSDocOptionalType:
		return bigIntAnnotation(ctx, annotation.Type(), bindings)
	case ast.KindTypeReference:
		symbol := ctx.Refs.ResolveInFileWithMeaning(annotation.AsTypeReferenceNode().TypeName, ast.SymbolFlagsType)
		if symbol == nil {
			return false
		}
		declaration := ast.GetDeclarationOfKind(symbol, ast.KindTypeAliasDeclaration)
		if declaration == nil {
			return false
		}
		if result, ok := bindings[declaration]; ok {
			return result
		}
		bindings[declaration] = false
		result := bigIntAnnotation(ctx, declaration.Type(), bindings)
		bindings[declaration] = result
		return result
	case ast.KindUnionType:
		return slices.ContainsFunc(annotation.AsUnionTypeNode().Types.Nodes, func(part *ast.Node) bool {
			return bigIntAnnotation(ctx, part, bindings)
		})
	case ast.KindIntersectionType:
		parts := annotation.AsIntersectionTypeNode().Types.Nodes
		for _, part := range parts {
			switch part.Kind {
			case ast.KindAnyKeyword, ast.KindUnknownKeyword, ast.KindJSDocAllType:
				return false
			}
		}
		return slices.ContainsFunc(parts, func(part *ast.Node) bool {
			return bigIntAnnotation(ctx, part, bindings)
		})
	}
	return false
}

func bigIntCall(ctx *rule.RuleContext, node *ast.Node, bindings map[*ast.Node]bool) bool {
	if ast.IsOptionalChain(node) {
		return false
	}
	callee := utils.ESTreeCallCallee(node.Expression())
	if isNamedIdentifier(callee, "BigInt") {
		return true
	}
	if call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Methods: []string{"asIntN", "asUintN"},
	}); ok && isNamedIdentifier(utils.ESTreeRuntimeExpression(call.Object), "BigInt") {
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
		initializer := utils.ESTreeRuntimeExpression(declaration.Initializer())
		if initializer != nil {
			initializer = ast.SkipOuterExpressions(initializer, ast.OEKParentheses|ast.OEKSatisfies|ast.OEKNonNullAssertions)
			if ast.IsFunctionExpressionOrArrowFunction(initializer) {
				function = initializer
			}
		}
	}
	if function != nil && (ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) || function.BodyData().AsteriskToken != nil) {
		return false
	}
	annotation, optional := unicornutil.ValueTypeAnnotation(*ctx, declaration)
	if function != declaration {
		if annotation != nil {
			if optional || annotation.Kind != ast.KindFunctionType {
				return false
			}
			annotation = annotation.Type()
		} else if function != nil && ast.IsVariableDeclaration(declaration) {
			annotation, _ = unicornutil.ValueTypeAnnotation(*ctx, function)
		}
	}
	return bigIntAnnotation(ctx, annotation, bindings) &&
		!slices.ContainsFunc(ctx.Refs.References(symbol), utils.IsWriteReference)
}
