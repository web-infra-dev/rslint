package no_unsafe_string_replacement

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-unsafe-string-replacement"

func unsafeReplacementMessage(method string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: "Do not use a non-literal replacement value with `String#" + method + "()`.",
		Data:        map[string]string{"method": method},
	}
}

var NoUnsafeStringReplacementRule = rule.Rule{
	Name:   "unicorn/no-unsafe-string-replacement",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				checkCall(ctx, node)
			},
		}
	},
}

func checkCall(ctx rule.RuleContext, node *ast.Node) {
	argumentsLength := 2
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Methods:             []string{"replace", "replaceAll"},
		ArgumentsLength:     &argumentsLength,
		RejectSpreadElement: true,
		AllowOptionalCall:   true,
		AllowOptionalMember: true,
	})
	if !ok {
		return
	}

	replacement := node.Arguments()[1]
	if isAllowedReplacement(ctx, replacement) ||
		isPlainObjectReplacement(ctx, replacement) ||
		isKnownNonStringReceiver(ctx, call.Object) {
		return
	}

	// ESTree does not preserve argument parentheses in the reported node range.
	// Keep TS assertion wrappers, which upstream includes in the range.
	reportNode := ast.SkipParentheses(replacement)
	method := call.Property.AsIdentifier().Text
	ctx.ReportNode(reportNode, unsafeReplacementMessage(method))
}

func isAllowedReplacement(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	case ast.KindTaggedTemplateExpression:
		return isStaticStringRawTaggedTemplate(ctx, node)
	default:
		return false
	}
}

func isStaticStringRawTaggedTemplate(ctx rule.RuleContext, node *ast.Node) bool {
	tagged := node.AsTaggedTemplateExpression()
	if tagged == nil || tagged.Template == nil ||
		tagged.Template.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return false
	}

	tag := ast.SkipParentheses(tagged.Tag)
	if tag == nil || !ast.IsPropertyAccessExpression(tag) || ast.IsOptionalChainRoot(tag) {
		return false
	}
	access := tag.AsPropertyAccessExpression()
	name := access.Name()
	object := ast.SkipParentheses(access.Expression)
	return name != nil && ast.IsIdentifier(name) && name.AsIdentifier().Text == "raw" &&
		isGlobalStringReference(ctx, object)
}

func isGlobalStringReference(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil || !ast.IsIdentifier(node) || node.AsIdentifier().Text != "String" ||
		!ctx.Globals.Access("String").IsDeclared() {
		return false
	}
	if ctx.Refs != nil {
		if symbol := ctx.Refs.Resolve(node); symbol != nil {
			return !utils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile)
		}
	}
	return !utils.IsShadowed(node, "String")
}

var objectCoercionPropertyNames = map[string]bool{
	"__proto__": true,
	"toString":  true,
	"valueOf":   true,
}

func isPlainObjectReplacement(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if initializer := utils.GetConstVariableInitializer(node, ctx.TypeChecker); initializer != nil {
		node = utils.SkipAssertionsAndParens(initializer)
	}
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return false
	}

	object := node.AsObjectLiteralExpression()
	if object == nil || object.Properties == nil {
		return true
	}
	for _, property := range object.Properties.Nodes {
		if !isPlainObjectProperty(property) {
			return false
		}
	}
	return true
}

func isPlainObjectProperty(property *ast.Node) bool {
	if property == nil ||
		(property.Kind != ast.KindPropertyAssignment &&
			property.Kind != ast.KindShorthandPropertyAssignment) {
		return false
	}
	name := property.Name()
	if name == nil || name.Kind == ast.KindComputedPropertyName {
		return false
	}
	staticName, ok := utils.GetStaticPropertyName(name)
	if !ok && name.Kind == ast.KindBigIntLiteral {
		staticName, ok = utils.NormalizeBigIntLiteral(name.AsBigIntLiteral().Text), true
	}
	return ok && !objectCoercionPropertyNames[staticName]
}

func isKnownNonStringReceiver(ctx rule.RuleContext, node *ast.Node) bool {
	if classifyStringReceiverSyntax(ctx, node, map[*ast.Symbol]bool{}) == unicornutil.TypeNonTarget {
		return true
	}

	if ctx.TypeChecker == nil {
		return false
	}

	t := ctx.TypeChecker.GetTypeAtLocation(node)
	return unicornutil.ClassifyType(ctx, t, unicornutil.TypeClassifierOptions{
		HeritageSymbolFlags:        ast.SymbolFlagsClass | ast.SymbolFlagsInterface,
		UnknownSymbolLessTypeFlags: checker.TypeFlagsNumberLiteral | checker.TypeFlagsBigIntLiteral,
		IsTargetType: func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike)
		},
	}) == unicornutil.TypeNonTarget
}

// classifyStringReceiverSyntax mirrors the syntax-only part of Unicorn's
// createTypeCheckers. It matters for source-only TypeScript Programs: those
// files have binder / RefStore services but intentionally no TypeChecker.
func classifyStringReceiverSyntax(ctx rule.RuleContext, node *ast.Node, visiting map[*ast.Symbol]bool) unicornutil.TypeClass {
	if node == nil {
		return unicornutil.TypeUnknown
	}

	switch node.Kind {
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
		return classifyStringReceiverSyntax(ctx, node.Expression(), visiting)
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		if class := classifyStringTypeSyntax(ctx, node.Type(), visiting); class != unicornutil.TypeUnknown {
			return class
		}
		return classifyStringReceiverSyntax(ctx, node.Expression(), visiting)
	case ast.KindSatisfiesExpression:
		// `satisfies` validates without changing the expression's type.
		return classifyStringReceiverSyntax(ctx, node.Expression(), visiting)
	case ast.KindIdentifier:
		if ctx.Refs == nil {
			return unicornutil.TypeUnknown
		}
		symbol := ctx.Refs.Resolve(node)
		// getTypeFromVariable() in Unicorn only reasons about a variable with one
		// definition. A merged type/value symbol must stay unknown: selecting the
		// type declaration could otherwise hide the value receiver's diagnostic.
		if symbol == nil || visiting[symbol] || len(symbol.Declarations) != 1 {
			return unicornutil.TypeUnknown
		}
		visiting[symbol] = true
		defer delete(visiting, symbol)

		if typeAnnotation := bindingTypeAnnotation(symbol.Declarations[0]); typeAnnotation != nil {
			if class := classifyStringTypeSyntax(ctx, typeAnnotation, visiting); class != unicornutil.TypeUnknown {
				return class
			}
		}
		if initializer := utils.GetConstVariableInitializer(node, nil); initializer != nil {
			return classifyStringReceiverSyntax(ctx, initializer, visiting)
		}
	}

	return unicornutil.TypeUnknown
}

// bindingTypeAnnotation returns only a type annotation on the binding itself.
// In particular, FunctionDeclaration.Type() is its return type, not a type for
// the function binding; Unicorn's definition.name.typeAnnotation does not read
// that return type.
func bindingTypeAnnotation(declaration *ast.Node) *ast.Node {
	if declaration == nil {
		return nil
	}
	switch declaration.Kind {
	case ast.KindVariableDeclaration:
		return declaration.AsVariableDeclaration().Type
	case ast.KindParameter:
		return declaration.AsParameterDeclaration().Type
	default:
		return nil
	}
}

func classifyStringTypeSyntax(ctx rule.RuleContext, node *ast.Node, visiting map[*ast.Symbol]bool) unicornutil.TypeClass {
	if node == nil {
		return unicornutil.TypeUnknown
	}

	switch node.Kind {
	case ast.KindStringKeyword:
		return unicornutil.TypeTarget
	case ast.KindLiteralType:
		literal := node.AsLiteralTypeNode().Literal
		if literal != nil && ast.IsStringLiteralLike(literal) {
			return unicornutil.TypeTarget
		}
		return unicornutil.TypeNonTarget
	case ast.KindBigIntKeyword, ast.KindBooleanKeyword, ast.KindNeverKeyword,
		ast.KindNumberKeyword, ast.KindObjectKeyword, ast.KindSymbolKeyword,
		ast.KindUndefinedKeyword, ast.KindVoidKeyword, ast.KindArrayType, ast.KindTupleType,
		ast.KindTypeLiteral, ast.KindFunctionType, ast.KindConstructorType:
		return unicornutil.TypeNonTarget
	case ast.KindParenthesizedType:
		return classifyStringTypeSyntax(ctx, node.Type(), visiting)
	case ast.KindTypeOperator:
		operator := node.AsTypeOperatorNode()
		if operator.Operator == ast.KindReadonlyKeyword {
			return classifyStringTypeSyntax(ctx, operator.Type, visiting)
		}
		return unicornutil.TypeUnknown
	case ast.KindUnionType:
		return combineStringTypeSyntax(node.AsUnionTypeNode().Types.Nodes, ctx, visiting, false)
	case ast.KindIntersectionType:
		return combineStringTypeSyntax(node.AsIntersectionTypeNode().Types.Nodes, ctx, visiting, true)
	case ast.KindTypeReference:
		return classifyStringTypeReference(ctx, node.AsTypeReferenceNode().TypeName, visiting)
	}

	return unicornutil.TypeUnknown
}

func combineStringTypeSyntax(parts []*ast.Node, ctx rule.RuleContext, visiting map[*ast.Symbol]bool, intersection bool) unicornutil.TypeClass {
	if len(parts) == 0 {
		return unicornutil.TypeUnknown
	}

	allNonTarget := true
	allTarget := true
	for _, part := range parts {
		class := classifyStringTypeSyntax(ctx, part, visiting)
		if intersection && class == unicornutil.TypeTarget {
			return unicornutil.TypeTarget
		}
		allTarget = allTarget && class == unicornutil.TypeTarget
		allNonTarget = allNonTarget && class == unicornutil.TypeNonTarget
	}
	if allTarget {
		return unicornutil.TypeTarget
	}
	if allNonTarget {
		return unicornutil.TypeNonTarget
	}
	return unicornutil.TypeUnknown
}

func classifyStringTypeReference(ctx rule.RuleContext, typeName *ast.Node, visiting map[*ast.Symbol]bool) unicornutil.TypeClass {
	if typeName == nil || !ast.IsIdentifier(typeName) || ctx.Refs == nil {
		return unicornutil.TypeUnknown
	}
	symbol := ctx.Refs.Resolve(typeName)
	if symbol == nil || visiting[symbol] {
		return unicornutil.TypeUnknown
	}
	visiting[symbol] = true
	defer delete(visiting, symbol)

	for _, declaration := range symbol.Declarations {
		switch declaration.Kind {
		case ast.KindTypeAliasDeclaration:
			if class := classifyStringTypeSyntax(ctx, declaration.Type(), visiting); class != unicornutil.TypeUnknown {
				return class
			}
		case ast.KindTypeParameter:
			return classifyStringTypeSyntax(ctx, declaration.AsTypeParameterDeclaration().Constraint, visiting)
		case ast.KindInterfaceDeclaration:
			interfaces := ast.GetHeritageElements(declaration, ast.KindExtendsKeyword)
			if len(interfaces) == 0 {
				return unicornutil.TypeNonTarget
			}
			parts := make([]*ast.Node, 0, len(interfaces))
			for _, base := range interfaces {
				if base.Kind == ast.KindExpressionWithTypeArguments {
					parts = append(parts, base.AsExpressionWithTypeArguments().Expression)
				}
			}
			return combineStringTypeExpressions(parts, ctx, visiting)
		case ast.KindClassDeclaration:
			extends := ast.GetHeritageElements(declaration, ast.KindExtendsKeyword)
			if len(extends) == 0 {
				return unicornutil.TypeNonTarget
			}
			parts := make([]*ast.Node, 0, len(extends))
			for _, base := range extends {
				if base.Kind == ast.KindExpressionWithTypeArguments {
					parts = append(parts, base.AsExpressionWithTypeArguments().Expression)
				}
			}
			return combineStringTypeExpressions(parts, ctx, visiting)
		}
	}

	return unicornutil.TypeUnknown
}

func combineStringTypeExpressions(parts []*ast.Node, ctx rule.RuleContext, visiting map[*ast.Symbol]bool) unicornutil.TypeClass {
	if len(parts) == 0 {
		return unicornutil.TypeUnknown
	}
	allNonTarget := true
	for _, part := range parts {
		class := classifyStringTypeReference(ctx, part, visiting)
		if class == unicornutil.TypeTarget {
			return unicornutil.TypeTarget
		}
		allNonTarget = allNonTarget && class == unicornutil.TypeNonTarget
	}
	if allNonTarget {
		return unicornutil.TypeNonTarget
	}
	return unicornutil.TypeUnknown
}
