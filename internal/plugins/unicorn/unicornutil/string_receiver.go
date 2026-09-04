package unicornutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsKnownNonString reports whether node's type is definitely not string-like.
// It follows local TypeScript annotations and const initializers before using
// the checker, matching Unicorn's source-only behavior for files that are not
// part of a configured project. Unknown receivers return false so syntactic
// string rules can still report them.
func IsKnownNonString(ctx rule.RuleContext, node *ast.Node) bool {
	return classifyStringReceiver(ctx, node, &stringReceiverWalkState{
		visiting: map[*ast.Symbol]bool{},
		memo:     map[*ast.Symbol]TypeClass{},
	}) == TypeNonTarget
}

// IsKnownNonStringType reports whether the checker alone proves that node is
// not string-like. Rules that require type information and intentionally do
// not follow Unicorn's syntax classifier use this narrower helper.
func IsKnownNonStringType(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.TypeChecker == nil || node == nil {
		return false
	}

	return classifyRequiredStringType(ctx, node) == TypeNonTarget
}

type stringReceiverWalkState struct {
	visiting map[*ast.Symbol]bool
	memo     map[*ast.Symbol]TypeClass
	depth    int
	steps    int
}

const (
	maxStringReceiverDepth = 1024
	maxStringReceiverSteps = 4096
)

func classifyStringReceiver(ctx rule.RuleContext, node *ast.Node, state *stringReceiverWalkState) TypeClass {
	if node == nil || state.depth >= maxStringReceiverDepth || state.steps >= maxStringReceiverSteps {
		return TypeUnknown
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return TypeUnknown
	}

	state.depth++
	state.steps++
	defer func() { state.depth-- }()

	// Unicorn ignores a satisfies annotation and classifies the expression.
	if node.Kind == ast.KindSatisfiesExpression {
		return classifyStringReceiver(ctx, node.Expression(), state)
	}
	// An informative assertion wins. `any` and `unknown` fall through to the
	// wrapped expression, matching upstream's getTypeFromExpression.
	if node.Kind == ast.KindAsExpression || node.Kind == ast.KindTypeAssertionExpression {
		if class := classifyStringTypeNode(ctx, node.Type(), map[*ast.Symbol]bool{}); class != TypeUnknown {
			return class
		}
		return classifyStringReceiver(ctx, node.Expression(), state)
	}
	if node.Kind == ast.KindNonNullExpression {
		return classifyStringReceiver(ctx, node.Expression(), state)
	}

	if ast.IsIdentifier(node) {
		if class := classifyStringIdentifier(ctx, node, state); class != TypeUnknown {
			return class
		}
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		class := combineStringUnion([]TypeClass{
			classifyStringReceiver(ctx, conditional.WhenTrue, state),
			classifyStringReceiver(ctx, conditional.WhenFalse, state),
		})
		if class != TypeUnknown {
			return class
		}
	}
	if ast.IsBinaryExpression(node) &&
		node.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
		if class := classifyStringReceiver(ctx, node.AsBinaryExpression().Right, state); class != TypeUnknown {
			return class
		}
	}

	if ctx.TypeChecker != nil {
		return classifyStringType(ctx, node)
	}
	return TypeUnknown
}

func classifyStringIdentifier(ctx rule.RuleContext, node *ast.Node, state *stringReceiverWalkState) TypeClass {
	if ctx.Refs == nil {
		return TypeUnknown
	}
	symbol := ctx.Refs.ResolveInFile(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return TypeUnknown
	}
	if class, ok := state.memo[symbol]; ok {
		return class
	}
	if state.visiting[symbol] {
		return TypeUnknown
	}
	state.visiting[symbol] = true
	defer delete(state.visiting, symbol)

	declaration := symbol.Declarations[0]
	if declaration == nil {
		return TypeUnknown
	}
	if class := classifyStringTypeNode(ctx, stringBindingTypeAnnotation(declaration), map[*ast.Symbol]bool{}); class != TypeUnknown {
		state.memo[symbol] = class
		return class
	}
	if !ast.IsVariableDeclaration(declaration) {
		return TypeUnknown
	}

	declarationList := declaration.Parent
	variable := declaration.AsVariableDeclaration()
	if declarationList == nil || !ast.IsVariableDeclarationList(declarationList) ||
		ast.IsVarUsing(declarationList) || ast.IsVarAwaitUsing(declarationList) ||
		declarationList.Flags&ast.NodeFlagsConst == 0 || variable.Name() == nil ||
		!ast.IsIdentifier(variable.Name()) || variable.Initializer == nil {
		return TypeUnknown
	}
	class := classifyStringReceiver(ctx, variable.Initializer, state)
	state.memo[symbol] = class
	return class
}

func stringBindingTypeAnnotation(declaration *ast.Node) *ast.Node {
	if declaration == nil {
		return nil
	}
	switch declaration.Kind {
	case ast.KindVariableDeclaration, ast.KindParameter:
		return declaration.Type()
	default:
		return nil
	}
}

func classifyStringTypeNode(
	ctx rule.RuleContext,
	node *ast.Node,
	visitedSymbols map[*ast.Symbol]bool,
) TypeClass {
	if node == nil {
		return TypeUnknown
	}
	for node.Kind == ast.KindParenthesizedType {
		node = node.AsParenthesizedTypeNode().Type
		if node == nil {
			return TypeUnknown
		}
	}
	if node.Kind == ast.KindTypeOperator {
		operator := node.AsTypeOperatorNode()
		if operator.Operator != ast.KindReadonlyKeyword {
			return TypeUnknown
		}
		return classifyStringTypeNode(ctx, operator.Type, visitedSymbols)
	}

	switch node.Kind {
	case ast.KindStringKeyword:
		return TypeTarget
	case ast.KindLiteralType:
		literal := node.AsLiteralTypeNode().Literal
		if literal != nil && literal.Kind == ast.KindStringLiteral {
			return TypeTarget
		}
		return TypeNonTarget
	case ast.KindNullKeyword, ast.KindUndefinedKeyword,
		ast.KindBigIntKeyword, ast.KindBooleanKeyword, ast.KindNeverKeyword,
		ast.KindNumberKeyword, ast.KindSymbolKeyword, ast.KindVoidKeyword,
		ast.KindArrayType, ast.KindTupleType, ast.KindTypeLiteral,
		ast.KindFunctionType, ast.KindConstructorType:
		return TypeNonTarget
	case ast.KindUnionType:
		types := node.AsUnionTypeNode().Types.Nodes
		classes := make([]TypeClass, 0, len(types))
		for _, part := range types {
			classes = append(classes, classifyStringTypeNode(ctx, part, visitedSymbols))
		}
		return combineStringUnion(classes)
	case ast.KindIntersectionType:
		types := node.AsIntersectionTypeNode().Types.Nodes
		classes := make([]TypeClass, 0, len(types))
		for _, part := range types {
			classes = append(classes, classifyStringTypeNode(ctx, part, visitedSymbols))
		}
		return combineStringIntersection(classes)
	case ast.KindTypeReference:
		return classifyStringTypeReference(ctx, node.AsTypeReferenceNode().TypeName, visitedSymbols)
	default:
		return TypeUnknown
	}
}

func classifyStringTypeReference(
	ctx rule.RuleContext,
	typeName *ast.Node,
	visitedSymbols map[*ast.Symbol]bool,
) TypeClass {
	if typeName == nil || !ast.IsIdentifier(typeName) || ctx.Refs == nil {
		return TypeUnknown
	}
	symbol := ctx.Refs.ResolveInFileWithMeaning(
		typeName, ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias,
	)
	if symbol == nil || visitedSymbols[symbol] {
		return TypeUnknown
	}
	visitedSymbols[symbol] = true
	defer delete(visitedSymbols, symbol)

	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		switch declaration.Kind {
		case ast.KindTypeAliasDeclaration:
			return classifyStringTypeNode(ctx, declaration.AsTypeAliasDeclaration().Type, visitedSymbols)
		case ast.KindTypeParameter:
			return classifyStringTypeNode(ctx, declaration.AsTypeParameterDeclaration().Constraint, visitedSymbols)
		case ast.KindInterfaceDeclaration:
			return classifyStringInterface(ctx, declaration, visitedSymbols)
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return classifyStringClass(ctx, declaration, visitedSymbols)
		}
	}
	return TypeUnknown
}

func classifyStringInterface(
	ctx rule.RuleContext,
	node *ast.Node,
	visitedSymbols map[*ast.Symbol]bool,
) TypeClass {
	declaration := node.AsInterfaceDeclaration()
	if declaration.HeritageClauses == nil || len(declaration.HeritageClauses.Nodes) == 0 {
		return TypeNonTarget
	}
	classes := make([]TypeClass, 0, len(declaration.HeritageClauses.Nodes))
	for _, clauseNode := range declaration.HeritageClauses.Nodes {
		clause := clauseNode.AsHeritageClause()
		if clause == nil || clause.Token != ast.KindExtendsKeyword || clause.Types == nil {
			continue
		}
		for _, heritageNode := range clause.Types.Nodes {
			heritage := heritageNode.AsExpressionWithTypeArguments()
			if heritage == nil {
				classes = append(classes, TypeUnknown)
				continue
			}
			classes = append(classes, classifyStringTypeReference(ctx, heritage.Expression, visitedSymbols))
		}
	}
	if len(classes) == 0 {
		return TypeNonTarget
	}
	return combineStringIntersection(classes)
}

func classifyStringClass(
	ctx rule.RuleContext,
	node *ast.Node,
	visitedSymbols map[*ast.Symbol]bool,
) TypeClass {
	heritageClauses := utils.GetHeritageClauses(node)
	if heritageClauses == nil {
		return TypeNonTarget
	}
	for _, clauseNode := range heritageClauses.Nodes {
		clause := clauseNode.AsHeritageClause()
		if clause == nil || clause.Token != ast.KindExtendsKeyword || clause.Types == nil || len(clause.Types.Nodes) == 0 {
			continue
		}
		heritage := clause.Types.Nodes[0].AsExpressionWithTypeArguments()
		if heritage == nil {
			return TypeUnknown
		}
		return classifyStringClassReference(ctx, heritage.Expression, visitedSymbols)
	}
	return TypeNonTarget
}

func classifyStringClassReference(
	ctx rule.RuleContext,
	node *ast.Node,
	visitedSymbols map[*ast.Symbol]bool,
) TypeClass {
	node = ast.SkipParentheses(node)
	if node == nil {
		return TypeUnknown
	}
	if node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression {
		return classifyStringClass(ctx, node, visitedSymbols)
	}
	if !ast.IsIdentifier(node) || ctx.Refs == nil {
		return TypeUnknown
	}

	symbol := ctx.Refs.ResolveInFileWithMeaning(
		node, ast.SymbolFlagsValue|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias,
	)
	if symbol == nil || visitedSymbols[symbol] {
		return TypeUnknown
	}
	visitedSymbols[symbol] = true
	defer delete(visitedSymbols, symbol)

	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		switch declaration.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return classifyStringClass(ctx, declaration, visitedSymbols)
		case ast.KindVariableDeclaration:
			declarationList := declaration.Parent
			variable := declaration.AsVariableDeclaration()
			if declarationList == nil || !ast.IsVariableDeclarationList(declarationList) ||
				declarationList.Flags&ast.NodeFlagsConst == 0 || variable.Name() == nil ||
				!ast.IsIdentifier(variable.Name()) || variable.Initializer == nil {
				return TypeUnknown
			}
			return classifyStringClassReference(ctx, variable.Initializer, visitedSymbols)
		}
	}
	return TypeUnknown
}

func combineStringUnion(classes []TypeClass) TypeClass {
	if len(classes) == 0 {
		return TypeNonTarget
	}
	allTarget := true
	allNonTarget := true
	for _, class := range classes {
		allTarget = allTarget && class == TypeTarget
		allNonTarget = allNonTarget && class == TypeNonTarget
	}
	if allTarget {
		return TypeTarget
	}
	if allNonTarget {
		return TypeNonTarget
	}
	return TypeUnknown
}

func combineStringIntersection(classes []TypeClass) TypeClass {
	allNonTarget := true
	for _, class := range classes {
		if class == TypeTarget {
			return TypeTarget
		}
		allNonTarget = allNonTarget && class == TypeNonTarget
	}
	if allNonTarget {
		return TypeNonTarget
	}
	return TypeUnknown
}

func classifyStringType(ctx rule.RuleContext, node *ast.Node) TypeClass {
	t := ctx.TypeChecker.GetTypeAtLocation(node)
	return classifyStringCheckerType(ctx, t)
}

func classifyStringCheckerType(ctx rule.RuleContext, t *checker.Type) TypeClass {
	if t == nil || utils.IsTypeAnyType(t) || utils.IsTypeUnknownType(t) || utils.IsIntrinsicErrorType(t) {
		return TypeUnknown
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
		return TypeNonTarget
	}
	if utils.IsTypeParameter(t) {
		constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, t)
		if constraint == nil {
			return TypeUnknown
		}
		return classifyStringCheckerType(ctx, constraint)
	}
	if utils.IsUnionType(t) {
		parts := utils.UnionTypeParts(t)
		classes := make([]TypeClass, 0, len(parts))
		for _, part := range parts {
			classes = append(classes, classifyStringCheckerType(ctx, part))
		}
		return combineStringUnion(classes)
	}
	if utils.IsIntersectionType(t) {
		parts := utils.IntersectionTypeParts(t)
		classes := make([]TypeClass, 0, len(parts))
		for _, part := range parts {
			classes = append(classes, classifyStringCheckerType(ctx, part))
		}
		return combineStringIntersection(classes)
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsStringLiteral) ||
		(utils.IsIntrinsicType(t) && t.AsIntrinsicType().IntrinsicName() == "string") {
		return TypeTarget
	}

	constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, t)
	if constraint != nil && constraint != t {
		return classifyStringCheckerType(ctx, constraint)
	}
	if stringTypeHasTargetBase(ctx, t) {
		return TypeTarget
	}
	if utils.IsIntrinsicType(t) {
		return TypeNonTarget
	}
	if _, ok := TypeSymbolName(t); !ok {
		return TypeUnknown
	}
	// is-string.js configures createTypeCheckers with no target type names, so
	// every named non-intrinsic type is a known non-string at this point.
	return TypeNonTarget
}

func stringTypeHasTargetBase(ctx rule.RuleContext, t *checker.Type) bool {
	symbol := checker.Type_symbol(t)
	if symbol == nil || symbol.Flags&(ast.SymbolFlagsClass|ast.SymbolFlagsInterface) == 0 {
		return false
	}
	declared := checker.Checker_getDeclaredTypeOfSymbol(ctx.TypeChecker, symbol)
	if declared == nil {
		return false
	}
	for _, base := range checker.Checker_getBaseTypes(ctx.TypeChecker, declared) {
		if classifyStringCheckerType(ctx, base) == TypeTarget {
			return true
		}
	}
	return false
}

func classifyRequiredStringType(ctx rule.RuleContext, node *ast.Node) TypeClass {
	t := ctx.TypeChecker.GetTypeAtLocation(node)
	return ClassifyType(ctx, t, TypeClassifierOptions{
		HeritageSymbolFlags:          ast.SymbolFlagsClass | ast.SymbolFlagsInterface,
		NonTargetSymbolLessTypeFlags: checker.TypeFlagsObject,
		IsTargetType: func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike)
		},
	})
}
