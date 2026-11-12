// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var domNodeTypeNames = utils.NewSetFromItems(
	"CharacterData", "Document", "DocumentFragment", "Element", "HTMLElement", "Node", "SVGElement", "Text",
)

// NewNonDOMNodeCheck creates a predicate for one rule run. It follows authored
// types before consulting the project checker. Each query starts with a fresh
// work limit and reuses the symbol map emptied by the previous walk.
func NewNonDOMNodeCheck(ctx rule.RuleContext) func(*ast.Node) bool {
	classifier := domNodeTypeClassifier{ctx: ctx}
	return func(node *ast.Node) bool {
		classifier.remaining = 4096
		return classifier.expression(node) == TypeNonTarget
	}
}

type domNodeTypeClassifier struct {
	ctx       rule.RuleContext
	remaining int // Bound work for deeply nested or repeatedly expanded types.
	visiting  map[*ast.Symbol]bool
}

func (c *domNodeTypeClassifier) expression(node *ast.Node) TypeClass {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || c.remaining == 0 {
		return TypeUnknown
	}
	c.remaining--
	class := TypeUnknown
	switch node.Kind {
	case ast.KindIdentifier:
		class = c.identifier(node)
	case ast.KindObjectBindingPattern:
		// ESTree attaches this annotation to the pattern itself.
		if parent := node.Parent; parent != nil {
			annotation := ast.GetTypeAnnotationNode(parent)
			if !utils.IsJSDocSyntaxNode(annotation) {
				class = c.annotation(annotation)
			}
		}
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		annotation := node.Type()
		class = c.annotation(annotation)
		if annotation != nil && (annotation.Kind == ast.KindAnyKeyword || annotation.Kind == ast.KindUnknownKeyword) {
			class = c.expression(node.Expression())
		}
	case ast.KindSatisfiesExpression:
		return c.expression(node.Expression())
	case ast.KindNonNullExpression:
		class = c.expression(node.Expression())
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		switch {
		case operator == ast.KindEqualsToken || operator == ast.KindCommaToken:
			class = c.expression(binary.Right)
		case ast.IsLogicalOrCoalescingBinaryOperator(operator):
			class = combineDOMTypes([]*ast.Node{binary.Left, binary.Right}, c.expression, false)
		}
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		class = combineDOMTypes([]*ast.Node{conditional.WhenTrue, conditional.WhenFalse}, c.expression, false)
	case ast.KindCallExpression:
		// An optional call inside a chain is still a CallExpression. A call
		// ending the chain is wrapped in an ESTree ChainExpression.
		if !ast.IsOptionalChain(node) || (node.Parent != nil && ast.IsOptionalChain(node.Parent) && node.Parent.Expression() == node) {
			class = c.callReturnType(utils.ESTreeCallCallee(node.Expression()))
		}
	}
	if class == TypeUnknown {
		return c.checkedType(node)
	}
	return class
}

func (c *domNodeTypeClassifier) checkedType(node *ast.Node) TypeClass {
	if c.ctx.TypeChecker == nil {
		return TypeUnknown
	}
	return ClassifyType(c.ctx, c.ctx.TypeChecker.GetTypeAtLocation(node), TypeClassifierOptions{
		IsTargetType: func(t *checker.Type) bool {
			symbol := checker.Type_symbol(t)
			return symbol != nil && domNodeTypeNames.Has(symbol.Name) && utils.IsSymbolFromDefaultLibrary(c.ctx.Program(), symbol)
		},
		HeritageSymbolFlags: ast.SymbolFlagsClass | ast.SymbolFlagsInterface,
	})
}

func (c *domNodeTypeClassifier) identifier(node *ast.Node) TypeClass {
	symbol := c.ctx.Refs.ResolveInFile(node)
	if symbol == nil || len(symbol.Declarations) != 1 || !c.enter(symbol) {
		return TypeUnknown
	}
	defer delete(c.visiting, symbol)

	declaration := symbol.Declarations[0]
	if !ast.IsVariableDeclaration(declaration) && !ast.IsParameterDeclaration(declaration) {
		return TypeUnknown
	}
	annotation, optional := c.bindingAnnotation(declaration)
	class := c.annotation(annotation)
	if class == TypeTarget && optional {
		class = TypeUnknown
	}
	if class != TypeUnknown || utils.IsJSDocSyntaxNode(annotation) {
		return class
	}
	if ast.IsVariableDeclaration(declaration) && ast.IsVarConst(declaration) && ast.IsIdentifier(declaration.Name()) {
		return c.expression(declaration.AsVariableDeclaration().Initializer)
	}
	return TypeUnknown
}

func hasNativeTypeAssertion(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKNonNullAssertions|ast.OEKSatisfies)
	return ast.IsAssertionExpression(node) && !utils.IsJSDocTypeCastWrapper(node)
}

func (c *domNodeTypeClassifier) annotation(node *ast.Node) TypeClass {
	if node == nil || c.remaining == 0 {
		return TypeUnknown
	}
	c.remaining--
	if utils.IsJSDocSyntaxNode(node) {
		switch node.Kind {
		case ast.KindArrayType, ast.KindTupleType, ast.KindNeverKeyword:
			// Unicorn's JSDoc parser leaves Array and unsupported types unknown.
			return TypeUnknown
		case ast.KindTypeLiteral:
			if len(node.Members()) == 0 {
				return TypeUnknown
			}
		case ast.KindJSDocTypeLiteral:
			if len(node.AsJSDocTypeLiteral().JSDocPropertyTags) == 0 {
				return TypeUnknown
			}
		case ast.KindTypeReference:
			name := node.AsTypeReferenceNode().TypeName
			if !ast.IsIdentifier(name) {
				return TypeUnknown
			}
			if domNodeTypeNames.Has(name.Text()) {
				return TypeTarget
			}
			switch name.Text() {
			case "Boolean", "String", "Number", "Void", "Null", "Undefined":
				return TypeNonTarget
			case "Object":
				if len(node.TypeArguments()) > 0 {
					return TypeNonTarget
				}
			}
			return TypeUnknown
		}
	}
	switch node.Kind {
	case ast.KindJSDocNonNullableType:
		return c.annotation(node.Type())
	case ast.KindParenthesizedType, ast.KindJSDocNullableType, ast.KindJSDocOptionalType:
		class := c.annotation(node.Type())
		if node.Kind != ast.KindParenthesizedType && class == TypeTarget {
			return TypeUnknown
		}
		return class
	case ast.KindTypeOperator:
		if node.AsTypeOperatorNode().Operator == ast.KindReadonlyKeyword {
			return c.annotation(node.Type())
		}
	case ast.KindTypeReference:
		return c.typeReference(node.AsTypeReferenceNode().TypeName)
	case ast.KindUnionType:
		return combineDOMTypes(node.AsUnionTypeNode().Types.Nodes, c.annotation, false)
	case ast.KindIntersectionType:
		return combineDOMTypes(node.AsIntersectionTypeNode().Types.Nodes, c.annotation, true)
	case ast.KindNullKeyword, ast.KindUndefinedKeyword, ast.KindBigIntKeyword, ast.KindBooleanKeyword,
		ast.KindNeverKeyword, ast.KindNumberKeyword, ast.KindStringKeyword, ast.KindSymbolKeyword, ast.KindVoidKeyword,
		ast.KindArrayType, ast.KindTupleType, ast.KindTypeLiteral, ast.KindLiteralType, ast.KindFunctionType,
		ast.KindConstructorType, ast.KindJSDocTypeLiteral:
		return TypeNonTarget
	}
	return TypeUnknown
}

func (c *domNodeTypeClassifier) typeReference(name *ast.Node) TypeClass {
	if name == nil || !ast.IsIdentifier(name) {
		return TypeUnknown
	}
	symbol := c.ctx.Refs.ResolveInFileWithMeaning(name, ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias)
	if symbol != nil && len(symbol.Declarations) > 0 {
		declaration := symbol.Declarations[0]
		if ast.IsImportSpecifier(declaration) &&
			(domNodeTypeNames.Has(declaration.PropertyNameOrName().Text()) || domNodeTypeNames.Has(name.Text())) {
			if class := c.checkedType(name.Parent); class != TypeUnknown {
				return class
			}
			return TypeNonTarget
		}
	}
	if domNodeTypeNames.Has(name.Text()) {
		return TypeTarget
	}
	if symbol == nil || len(symbol.Declarations) == 0 || !c.enter(symbol) {
		return TypeUnknown
	}
	defer delete(c.visiting, symbol)

	declaration := symbol.Declarations[0]
	switch declaration.Kind {
	case ast.KindTypeAliasDeclaration:
		return c.annotation(declaration.Type())
	case ast.KindTypeParameter:
		return c.annotation(declaration.AsTypeParameterDeclaration().Constraint)
	case ast.KindInterfaceDeclaration:
		return combineDOMTypes(ast.GetHeritageElements(declaration, ast.KindExtendsKeyword), func(base *ast.Node) TypeClass {
			if reference := base.AsTypeReferenceNode(); reference != nil {
				return c.typeReference(reference.TypeName)
			}
			return TypeUnknown
		}, true)
	case ast.KindClassDeclaration, ast.KindClassExpression:
		base := ast.GetClassExtendsHeritageElement(declaration)
		if base == nil {
			return TypeNonTarget
		}
		name := utils.ESTreeRuntimeExpression(base.AsExpressionWithTypeArguments().Expression)
		if name != nil && ast.IsIdentifier(name) {
			return c.classReference(name)
		}
	}
	return TypeUnknown
}

func (c *domNodeTypeClassifier) classReference(node *ast.Node) TypeClass {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || c.remaining == 0 {
		return TypeUnknown
	}
	c.remaining--
	if ast.IsClassLike(node) {
		base := ast.GetClassExtendsHeritageElement(node)
		if base == nil {
			return TypeNonTarget
		}
		return c.classReference(base.AsExpressionWithTypeArguments().Expression)
	}
	if !ast.IsIdentifier(node) {
		return TypeUnknown
	}
	symbol := c.ctx.Refs.ResolveInFileWithMeaning(node, ast.SymbolFlagsValue|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias)
	if symbol == nil || len(symbol.Declarations) == 0 {
		if domNodeTypeNames.Has(node.Text()) {
			return TypeTarget
		}
		return TypeUnknown
	}
	if !c.enter(symbol) {
		return TypeUnknown
	}
	defer delete(c.visiting, symbol)
	declaration := symbol.Declarations[0]
	if ast.IsClassLike(declaration) {
		return c.classReference(declaration)
	}
	if ast.IsVariableDeclaration(declaration) && ast.IsVarConst(declaration) && ast.IsIdentifier(declaration.Name()) {
		return c.classReference(declaration.AsVariableDeclaration().Initializer)
	}
	return TypeUnknown
}

func (c *domNodeTypeClassifier) callReturnType(callee *ast.Node) TypeClass {
	if callee == nil || !ast.IsIdentifier(callee) {
		return TypeUnknown
	}
	symbol := c.ctx.Refs.ResolveInFile(callee)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return TypeUnknown
	}
	declaration := symbol.Declarations[0]
	function, ok := UnmodifiedFunctionValue(c.ctx, symbol)
	if !ok {
		return TypeUnknown
	}
	if ast.IsParameterDeclaration(declaration) {
		initializer := declaration.Initializer()
		if hasNativeTypeAssertion(initializer) {
			return TypeUnknown
		}
		if value := utils.SkipAssertionsAndParens(initializer); value != nil && ast.IsFunctionLike(value) {
			if resultType := value.Type(); isAsyncOrGenerator(value) || (resultType != nil && !utils.IsJSDocSyntaxNode(resultType)) {
				return TypeUnknown
			}
		}
	}
	if function != nil {
		if annotation := function.Type(); annotation != nil && !utils.IsJSDocSyntaxNode(annotation) {
			return c.annotation(annotation)
		}
	}
	if ast.IsVariableDeclaration(declaration) || ast.IsParameterDeclaration(declaration) {
		annotation, optional := c.bindingAnnotation(declaration)
		if annotation != nil {
			if optional {
				return TypeUnknown
			}
			return classifyCallableReturnType(annotation, c.annotation)
		}
	}
	if function != nil {
		tag := c.uniqueJSDocTag(function, ast.KindJSDocReturnTag)
		return c.annotation(c.jsDocType(tag, function))
	}
	return TypeUnknown
}

func (c *domNodeTypeClassifier) enter(symbol *ast.Symbol) bool {
	if c.visiting[symbol] || c.remaining == 0 {
		return false
	}
	c.remaining--
	if c.visiting == nil {
		c.visiting = make(map[*ast.Symbol]bool)
	}
	c.visiting[symbol] = true
	return true
}

func combineDOMTypes(nodes []*ast.Node, classify func(*ast.Node) TypeClass, intersection bool) TypeClass {
	class := TypeNonTarget
	for i, node := range nodes {
		part := classify(node)
		if part == TypeUnknown || (!intersection && i > 0 && part != class) {
			return TypeUnknown
		}
		if !intersection || part == TypeTarget {
			class = part
		}
	}
	return class
}
