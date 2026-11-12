package unicornutil

import (
	"maps"
	"math/big"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsBigInt recognizes BigInt expressions using available types, authored syntax,
// and side-effect-free static values.
func IsBigInt(ctx *rule.RuleContext, evaluator *utils.StaticStringEvaluator, node *ast.Node) bool {
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
	return ok && isBigInt && !HasJSDocTypeReference(*ctx, node, make(map[*ast.Symbol]bool))
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
		annotation, _ := ValueTypeAnnotation(*ctx, declaration)
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

type bigIntAliasInstance struct {
	declaration *ast.Node
	arguments   string
}

type bigIntAnnotationWalk struct {
	ctx       *rule.RuleContext
	instances map[bigIntAliasInstance]bigIntAnnotationClass
}

func classifyBigIntAnnotation(ctx *rule.RuleContext, annotation *ast.Node, bindings map[*ast.Node]bigIntAnnotationClass) bigIntAnnotationClass {
	walk := bigIntAnnotationWalk{ctx: ctx}
	return walk.classify(annotation, bindings)
}

func (walk *bigIntAnnotationWalk) classify(annotation *ast.Node, bindings map[*ast.Node]bigIntAnnotationClass) bigIntAnnotationClass {
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
		return walk.classify(annotation.Type(), bindings)
	case ast.KindTypeReference:
		return walk.classifyReference(annotation, bindings)
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
		switch walk.classify(part, bindings) {
		case annotationAny:
			return annotationAny
		case annotationBigInt:
			result = annotationBigInt
		}
	}
	return result
}

func (walk *bigIntAnnotationWalk) classifyReference(reference *ast.Node, bindings map[*ast.Node]bigIntAnnotationClass) bigIntAnnotationClass {
	symbol := walk.ctx.Refs.ResolveInFileWithMeaning(reference.AsTypeReferenceNode().TypeName, ast.SymbolFlagsType)
	if symbol == nil {
		return annotationOther
	}
	declaration := ast.GetDeclarationOfKind(symbol, ast.KindTypeAliasDeclaration)
	var referencedType *ast.Node
	var parameters []*ast.Node
	if declaration != nil {
		referencedType = declaration.Type()
		parameters = declaration.TypeParameters()
	} else {
		declaration = ast.GetDeclarationOfKind(symbol, ast.KindTypeParameter)
		if declaration == nil {
			return annotationOther
		}
		referencedType = declaration.AsTypeParameterDeclaration().Constraint
	}
	if result, ok := bindings[declaration]; ok {
		return result
	}
	// A generic alias's unbound parameter has no concrete type argument.
	if declaration.Kind == ast.KindTypeParameter && ast.IsTypeAliasDeclaration(declaration.Parent) {
		return annotationOther
	}

	if len(parameters) != 0 {
		// Keep each instantiation separate. Supplied arguments use the caller's
		// bindings, while defaults can reference earlier parameters of this alias.
		arguments := reference.TypeArguments()
		classes := make([]byte, min(len(arguments), len(parameters)))
		for i := range classes {
			classes[i] = byte(walk.classify(arguments[i], bindings))
		}
		key := bigIntAliasInstance{declaration: declaration, arguments: string(classes)}
		if result, ok := walk.instances[key]; ok {
			return result
		}
		if walk.instances == nil {
			walk.instances = make(map[bigIntAliasInstance]bigIntAnnotationClass)
		}
		// A default can resolve another instance of this alias with explicit arguments.
		walk.instances[key] = annotationOther
		instance := maps.Clone(bindings)
		for i, parameter := range parameters {
			if i < len(classes) {
				instance[parameter] = bigIntAnnotationClass(classes[i])
			} else {
				instance[parameter] = walk.classify(parameter.AsTypeParameterDeclaration().DefaultType, instance)
			}
		}
		result := walk.classify(referencedType, instance)
		walk.instances[key] = result
		return result
	}
	bindings[declaration] = annotationOther
	result := walk.classify(referencedType, bindings)
	bindings[declaration] = result
	return result
}

func isGlobalBigInt(ctx *rule.RuleContext, node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && ast.IsIdentifier(node) && node.Text() == "BigInt" &&
		ctx.Globals.Access("BigInt").IsDeclared() && IsGlobalReference(*ctx, node)
}

func bigIntCall(ctx *rule.RuleContext, node *ast.Node, bindings map[*ast.Node]bigIntAnnotationClass) bool {
	if ast.IsOptionalChain(node) {
		return false
	}
	callee := utils.ESTreeCallCallee(node.Expression())
	if isGlobalBigInt(ctx, callee) {
		return true
	}
	if call, ok := MatchDotMethodCall(node, DotMethodCallOptions{
		Methods: []string{"asIntN", "asUintN"},
	}); ok && isGlobalBigInt(ctx, call.Object) {
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
	annotation := CallableReturnTypeAnnotation(*ctx, declaration)
	if annotation == nil && function != nil && ast.IsVariableDeclaration(declaration) {
		// Only infer a variable's return type from its initializer when the
		// binding has no explicit type of its own.
		if bindingType, _ := ValueTypeAnnotation(*ctx, declaration); bindingType == nil {
			annotation = CallableReturnTypeAnnotation(*ctx, function)
		}
	}
	return classifyBigIntAnnotation(ctx, annotation, bindings) == annotationBigInt &&
		!slices.ContainsFunc(ctx.Refs.References(symbol), utils.IsWriteReference)
}
