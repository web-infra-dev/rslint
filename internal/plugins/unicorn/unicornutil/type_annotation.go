package unicornutil

import (
	"maps"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ClassifyTypeAnnotation classifies authored type syntax without requiring a
// TypeChecker. It resolves local aliases and constraints through the binder.
func ClassifyTypeAnnotation(ctx rule.RuleContext, node *ast.Node, options TypeClassifierOptions) TypeClass {
	walk := typeAnnotationWalk{
		visiting:  make(map[annotationReference]bool),
		completed: make(map[annotationReference]TypeClass),
	}
	class := classifyTypeAnnotation(ctx, node, &walk, options)
	if class == typeAny || class == typeCyclic {
		return TypeUnknown
	}
	return class
}

type typeAnnotationWalk struct {
	visiting      map[annotationReference]bool
	completed     map[annotationReference]TypeClass
	bindings      map[*ast.Symbol]TypeClass
	instantiation string
}

// Type aliases and typeof queries can resolve to the same merged symbol.
type annotationReference struct {
	symbol    *ast.Symbol
	value     bool
	arguments string
}

func (walk *typeAnnotationWalk) classifyReference(ref annotationReference, classify func() TypeClass) TypeClass {
	// Recursive aliases remain cyclic when their arguments change.
	active := ref
	active.arguments = ""
	if walk.visiting[active] {
		// Propagate cycles so an enclosing intersection cannot cache a
		// classification based only on its non-cyclic members.
		return typeCyclic
	}
	if completed, ok := walk.completed[ref]; ok {
		return completed
	}
	walk.visiting[active] = true
	class := classify()
	delete(walk.visiting, active)
	walk.completed[ref] = class
	return class
}

func classifyTypeAnnotation(
	ctx rule.RuleContext,
	node *ast.Node,
	walk *typeAnnotationWalk,
	options TypeClassifierOptions,
) TypeClass {
	if node == nil {
		return TypeUnknown
	}
	node = ast.SkipTypeParentheses(node)
	if node.Kind == ast.KindTypeOperator {
		typeOperator := node.AsTypeOperatorNode()
		if typeOperator.Operator != ast.KindReadonlyKeyword {
			return TypeUnknown
		}
		return classifyTypeAnnotation(ctx, typeOperator.Type, walk, options)
	}

	switch node.Kind {
	case ast.KindAnyKeyword, ast.KindJSDocAllType:
		return typeAny
	case ast.KindArrayType, ast.KindTupleType:
		if classifyKnownTypeName("Array", options) == TypeTarget {
			return TypeTarget
		}
		return TypeNonTarget
	case ast.KindJSDocNonNullableType:
		return classifyTypeAnnotation(ctx, node.Type(), walk, options)
	case ast.KindJSDocNullableType, ast.KindJSDocOptionalType:
		return includeNullishSourceType(classifyTypeAnnotation(ctx, node.Type(), walk, options), options)
	case ast.KindNullKeyword, ast.KindUndefinedKeyword:
		if options.AllowNullishInMixedUnion {
			return TypeNullish
		}
		return TypeNonTarget
	case ast.KindLiteralType:
		if literal := node.AsLiteralTypeNode().Literal; literal.Kind == ast.KindNullKeyword {
			return classifyTypeAnnotation(ctx, literal, walk, options)
		}
		return TypeNonTarget
	case ast.KindUnionType:
		types := node.AsUnionTypeNode().Types.Nodes
		classes := make([]TypeClass, 0, len(types))
		for _, part := range types {
			classes = append(classes, classifyTypeAnnotation(ctx, part, walk, options))
		}
		return combineSourceTypeUnion(classes, options)
	case ast.KindIntersectionType:
		types := node.AsIntersectionTypeNode().Types.Nodes
		classes := make([]TypeClass, 0, len(types))
		for _, part := range types {
			classes = append(classes, classifyTypeAnnotation(ctx, part, walk, options))
		}
		return combineSourceTypeIntersection(classes)
	case ast.KindTypeReference:
		typeName := node.AsTypeReferenceNode().TypeName
		if options.AllowNullishInMixedUnion && utils.IsJSDocSyntaxNode(node) &&
			(typeName.Text() == "Null" || typeName.Text() == "Undefined") {
			return TypeNullish
		}
		return classifyTypeReference(ctx, typeName, node.TypeArguments(), walk, options)
	case ast.KindTypeQuery:
		return classifyTypeQuery(ctx, node, walk, options)
	case ast.KindBigIntKeyword, ast.KindBooleanKeyword, ast.KindNeverKeyword,
		ast.KindNumberKeyword, ast.KindStringKeyword, ast.KindSymbolKeyword,
		ast.KindVoidKeyword, ast.KindTypeLiteral,
		ast.KindFunctionType, ast.KindConstructorType:
		return TypeNonTarget
	default:
		return TypeUnknown
	}
}

func classifyTypeReference(
	ctx rule.RuleContext,
	typeName *ast.Node,
	typeArguments []*ast.Node,
	walk *typeAnnotationWalk,
	options TypeClassifierOptions,
) TypeClass {
	if typeName == nil || !ast.IsIdentifier(typeName) {
		return TypeUnknown
	}
	name := typeName.AsIdentifier().Text
	known := classifyKnownTypeName(name, options)
	if options.PreferKnownAnnotationNames && known != TypeUnknown {
		return known
	}
	var symbol *ast.Symbol
	if ctx.Refs != nil {
		symbol = ctx.Refs.ResolveInFileWithMeaning(
			typeName, ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias,
		)
	}
	if symbol == nil {
		if name == "ReturnType" && len(typeArguments) == 1 {
			if function := ast.SkipTypeParentheses(typeArguments[0]); ast.IsFunctionTypeNode(function) {
				return classifyTypeAnnotation(ctx, function.Type(), walk, options)
			}
		}
		return known
	}
	if bound, ok := walk.bindings[symbol]; ok {
		return bound
	}
	for _, declaration := range symbol.Declarations {
		if declaration != nil && declaration.Kind == ast.KindTypeAliasDeclaration && len(declaration.TypeParameters()) != 0 {
			return classifyGenericTypeAlias(ctx, declaration, typeArguments, walk, options)
		}
	}
	ref := annotationReference{symbol: symbol}
	if symbol.Flags&ast.SymbolFlagsTypeParameter != 0 {
		// A nested function type's constraint can depend on alias arguments.
		ref.arguments = walk.instantiation
	}
	return walk.classifyReference(ref, func() TypeClass {
		return classifyTypeDeclaration(ctx, symbol, name, walk, options)
	})
}

func classifyGenericTypeAlias(ctx rule.RuleContext, declaration *ast.Node, typeArguments []*ast.Node, walk *typeAnnotationWalk, options TypeClassifierOptions) TypeClass {
	symbol := declaration.Symbol()
	if walk.visiting[annotationReference{symbol: symbol}] {
		return typeCyclic
	}
	// Evaluate supplied arguments in the caller's bindings. The classification
	// of each argument identifies this instantiation in the completed cache.
	arguments := make([]byte, len(typeArguments))
	for i, argument := range typeArguments {
		arguments[i] = byte(classifyTypeAnnotation(ctx, argument, walk, options))
	}
	ref := annotationReference{symbol: symbol, arguments: string(arguments)}
	return walk.classifyReference(ref, func() TypeClass {
		previous, previousInstantiation := walk.bindings, walk.instantiation
		walk.bindings = maps.Clone(previous)
		walk.instantiation = ref.arguments
		if walk.bindings == nil {
			walk.bindings = make(map[*ast.Symbol]TypeClass)
		}
		defer func() {
			walk.bindings = previous
			walk.instantiation = previousInstantiation
		}()
		for i, parameter := range declaration.TypeParameters() {
			class := TypeUnknown
			if i < len(arguments) {
				class = TypeClass(arguments[i])
			} else if fallback := parameter.AsTypeParameterDeclaration().DefaultType; fallback != nil {
				class = classifyTypeAnnotation(ctx, fallback, walk, options)
			}
			walk.bindings[parameter.Symbol()] = class
		}
		return classifyTypeAnnotation(ctx, declaration.AsTypeAliasDeclaration().Type, walk, options)
	})
}

func classifyTypeDeclaration(ctx rule.RuleContext, symbol *ast.Symbol, name string, walk *typeAnnotationWalk, options TypeClassifierOptions) TypeClass {
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		switch declaration.Kind {
		case ast.KindTypeAliasDeclaration:
			return classifyTypeAnnotation(
				ctx, declaration.AsTypeAliasDeclaration().Type, walk, options,
			)
		case ast.KindTypeParameter:
			constraint := classifyTypeAnnotation(ctx, declaration.AsTypeParameterDeclaration().Constraint, walk, options)
			if constraint == typeAny {
				return TypeUnknown
			}
			return constraint
		case ast.KindInterfaceDeclaration:
			return classifyInterfaceAnnotation(
				ctx, declaration, walk, options,
			)
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return classifyClassReference(ctx, declaration, make(map[*ast.Symbol]bool), options)
		case ast.KindImportClause, ast.KindImportSpecifier, ast.KindNamespaceImport:
			importedName := name
			if declaration.Kind == ast.KindImportSpecifier {
				if propertyName := declaration.PropertyName(); propertyName != nil && ast.IsIdentifier(propertyName) {
					importedName = propertyName.AsIdentifier().Text
				}
			}
			if classifyKnownTypeName(name, options) != TypeUnknown ||
				classifyKnownTypeName(importedName, options) != TypeUnknown {
				return TypeNonTarget
			}
			return TypeUnknown
		}
	}
	return TypeUnknown
}

func classifyTypeQuery(ctx rule.RuleContext, node *ast.Node, walk *typeAnnotationWalk, options TypeClassifierOptions) TypeClass {
	name := node.AsTypeQueryNode().ExprName
	if name == nil || !ast.IsIdentifier(name) || len(node.TypeArguments()) != 0 {
		return TypeUnknown
	}
	symbol := ctx.Refs.ResolveInFileWithMeaning(name, ast.SymbolFlagsValue|ast.SymbolFlagsAlias)
	if symbol == nil || symbol.ValueDeclaration == nil {
		return TypeUnknown
	}
	declaration := symbol.ValueDeclaration
	if declaration.Kind != ast.KindVariableDeclaration && declaration.Kind != ast.KindParameter {
		return TypeUnknown
	}
	ref := annotationReference{symbol: symbol, value: true, arguments: walk.instantiation}
	return walk.classifyReference(ref, func() TypeClass {
		annotation, _ := ValueTypeAnnotation(ctx, declaration)
		return classifyTypeAnnotation(ctx, annotation, walk, options)
	})
}

func classifyInterfaceAnnotation(
	ctx rule.RuleContext,
	node *ast.Node,
	walk *typeAnnotationWalk,
	options TypeClassifierOptions,
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
			heritage := heritageNode.AsTypeReferenceNode()
			if heritage == nil {
				classes = append(classes, TypeUnknown)
				continue
			}
			classes = append(classes, classifyTypeReference(
				ctx, heritage.TypeName, nil, walk, options,
			))
		}
	}
	if len(classes) == 0 {
		return TypeNonTarget
	}
	return combineSourceTypeIntersection(classes)
}

func classifyKnownTypeName(name string, options TypeClassifierOptions) TypeClass {
	if options.TargetTypeNames != nil && options.TargetTypeNames.Has(name) {
		return TypeTarget
	}
	if options.NonTargetTypeNames != nil && options.NonTargetTypeNames.Has(name) {
		return TypeNonTarget
	}
	return TypeUnknown
}

func classifyClassReference(ctx rule.RuleContext, node *ast.Node, visited map[*ast.Symbol]bool, options TypeClassifierOptions) TypeClass {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return TypeUnknown
	}
	if ast.IsIdentifier(node) {
		if ctx.Refs == nil {
			return classifyKnownTypeName(node.Text(), options)
		}
		symbol := ctx.Refs.ResolveInFile(node)
		if symbol == nil {
			return classifyKnownTypeName(node.Text(), options)
		}
		if visited[symbol] {
			return typeCyclic
		}
		if len(symbol.Declarations) == 0 {
			return TypeUnknown
		}
		visited[symbol] = true
		defer delete(visited, symbol)
		declaration := symbol.Declarations[0]
		if ast.IsVariableDeclaration(declaration) && ast.IsIdentifier(declaration.Name()) && ast.IsVarConst(declaration) {
			return classifyClassReference(ctx, declaration.AsVariableDeclaration().Initializer, visited, options)
		}
		if ast.IsClassDeclaration(declaration) || ast.IsClassExpression(declaration) {
			return classifyClassReference(ctx, declaration, visited, options)
		}
		return TypeUnknown
	}
	if !ast.IsClassDeclaration(node) && !ast.IsClassExpression(node) {
		return TypeUnknown
	}
	if options.HeritageSymbolFlags&ast.SymbolFlagsClass == 0 {
		return TypeNonTarget
	}
	if base := ast.GetClassExtendsHeritageElement(node); base != nil {
		return classifyClassReference(ctx, base.AsExpressionWithTypeArguments().Expression, visited, options)
	}
	return TypeNonTarget
}

func includeNullishSourceType(class TypeClass, options TypeClassifierOptions) TypeClass {
	if options.AllowNullishInMixedUnion {
		return class
	}
	return combineSourceTypeUnion([]TypeClass{class, TypeNonTarget}, options)
}

func combineSourceTypeUnion(classes []TypeClass, options TypeClassifierOptions) TypeClass {
	if options.TreatMixedUnionAsTarget && slices.Contains(classes, TypeTarget) {
		return TypeTarget
	}
	if slices.Contains(classes, typeAny) {
		return typeAny
	}
	if slices.Contains(classes, typeCyclic) {
		return typeCyclic
	}
	if options.AllowNullishInMixedUnion {
		classes = slices.DeleteFunc(classes, func(class TypeClass) bool { return class == TypeNullish })
		if len(classes) == 0 {
			return TypeNullish
		}
	}
	if options.TreatMixedUnionAsNonTarget && slices.Contains(classes, TypeNonTarget) {
		return TypeNonTarget
	}
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

func combineSourceTypeIntersection(classes []TypeClass) TypeClass {
	// An explicit any absorbs the intersection, even if another member is a
	// known target. A cyclic reference has no usable type for this analysis.
	if slices.Contains(classes, typeAny) {
		return typeAny
	}
	if slices.Contains(classes, typeCyclic) {
		return typeCyclic
	}
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
