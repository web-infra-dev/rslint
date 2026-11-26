package estree

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"unicode"

	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	jsxtx "github.com/microsoft/TypeScript/tsc/shim/transformers/jsxtransforms"
)

func (p *estreeProjector) convert(node *ast.Node, pattern bool) uint32 {
	if p.mapping == nil {
		return p.convertNode(node, pattern)
	}
	previous := p.origin
	p.origin = node
	id := p.convertNode(node, pattern)
	p.origin = previous
	p.registerReverse(node, id)
	return id
}

func (p *estreeProjector) registerReverse(node *ast.Node, id uint32) {
	m := p.mapping
	if m == nil || node == nil || id == 0 {
		return
	}
	index := m.table.GetIndex(node)
	if index == 0 || m.Reverse[index] != 0 {
		return
	}
	m.Reverse[index] = id
	if id == m.chainRecord {
		m.chainReverse = append(m.chainReverse, index)
	}
}

func (p *estreeProjector) convertNode(node *ast.Node, pattern bool) uint32 {
	if node == nil {
		return 0
	}
	p.check()
	if expression := utils.JSDocTypeCastExpression(node); expression != nil {
		return p.convert(expression, pattern)
	}
	if isJSDocCommentSyntax(node) {
		return 0
	}
	p.recordTokenCover(node, false)
	switch node.Kind {
	case ast.KindSourceFile:
		file := node.AsSourceFile()
		body := p.convertStatements(file.Statements, true)
		// Preserve the rule-facing span: TS begins at its first syntax token,
		// JS at zero, and both end at EOF. Disable directives use this position.
		start := uint32(0)
		end := p.offsets.At(node.End())
		if !p.isJS {
			start = p.offsets.At(scanner.SkipTrivia(p.text, 0))
		}
		return p.add(typeProgram, start, end, body, p.stringValue(p.sourceType))

	case ast.KindBlock:
		return p.addNode(typeBlockStatement, node, p.convertStatements(node.AsBlock().Statements, node.Parent != nil && ast.IsFunctionLike(node.Parent)))
	case ast.KindModuleBlock:
		return p.addNode(typeTSModuleBlock, node, p.convertStatements(node.AsModuleBlock().Statements, true))
	case ast.KindEmptyStatement:
		return p.addNode(typeEmptyStatement, node)
	case ast.KindDebuggerStatement:
		return p.addNode(typeDebuggerStatement, node)
	case ast.KindExpressionStatement:
		return p.convertExpressionStatement(node, false)
	case ast.KindReturnStatement:
		return p.addNode(typeReturnStatement, node, nodeValue(p.convert(node.AsReturnStatement().Expression, false)))
	case ast.KindThrowStatement:
		return p.addNode(typeThrowStatement, node, nodeValue(p.convert(node.AsThrowStatement().Expression, false)))
	case ast.KindIfStatement:
		n := node.AsIfStatement()
		return p.addNode(typeIfStatement, node,
			nodeValue(p.convert(n.Expression, false)),
			nodeValue(p.convertStatement(n.ThenStatement)),
			nodeValue(p.convertStatement(n.ElseStatement)))
	case ast.KindWhileStatement:
		n := node.AsWhileStatement()
		return p.addNode(typeWhileStatement, node,
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convertStatement(n.Statement)))
	case ast.KindDoStatement:
		n := node.AsDoStatement()
		return p.addNode(typeDoWhileStatement, node,
			nodeValue(p.convertStatement(n.Statement)), nodeValue(p.convert(n.Expression, false)))
	case ast.KindForStatement:
		n := node.AsForStatement()
		return p.addNode(typeForStatement, node,
			nodeValue(p.convert(n.Initializer, false)),
			nodeValue(p.convert(n.Condition, false)),
			nodeValue(p.convert(n.Incrementor, false)),
			nodeValue(p.convertStatement(n.Statement)))
	case ast.KindForInStatement, ast.KindForOfStatement:
		n := node.AsForInOrOfStatement()
		left := p.convert(n.Initializer, true)
		if node.Kind == ast.KindForInStatement {
			return p.addNode(typeForInStatement, node,
				nodeValue(left), nodeValue(p.convert(n.Expression, false)), nodeValue(p.convertStatement(n.Statement)))
		}
		return p.addNode(typeForOfStatement, node,
			boolValue(n.AwaitModifier != nil), nodeValue(left),
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convertStatement(n.Statement)))
	case ast.KindBreakStatement:
		return p.addNode(typeBreakStatement, node, nodeValue(p.convert(node.AsBreakStatement().Label, false)))
	case ast.KindContinueStatement:
		return p.addNode(typeContinueStatement, node, nodeValue(p.convert(node.AsContinueStatement().Label, false)))
	case ast.KindWithStatement:
		n := node.AsWithStatement()
		return p.addNode(typeWithStatement, node,
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convertStatement(n.Statement)))
	case ast.KindSwitchStatement:
		n := node.AsSwitchStatement()
		var clauses []*ast.Node
		if n.CaseBlock != nil && n.CaseBlock.AsCaseBlock().Clauses != nil {
			clauses = n.CaseBlock.AsCaseBlock().Clauses.Nodes
		}
		return p.addNode(typeSwitchStatement, node,
			nodeValue(p.convert(n.Expression, false)), p.convertList(clauses, false))
	case ast.KindCaseClause, ast.KindDefaultClause:
		n := node.AsCaseOrDefaultClause()
		test := uint32(0)
		if node.Kind == ast.KindCaseClause {
			test = p.convert(n.Expression, false)
		}
		return p.addNode(typeSwitchCase, node, nodeValue(test), p.convertStatements(n.Statements, false))
	case ast.KindTryStatement:
		n := node.AsTryStatement()
		return p.addNode(typeTryStatement, node,
			nodeValue(p.convert(n.TryBlock, false)),
			nodeValue(p.convert(n.CatchClause, false)),
			nodeValue(p.convert(n.FinallyBlock, false)))
	case ast.KindCatchClause:
		n := node.AsCatchClause()
		var parameter uint32
		if n.VariableDeclaration != nil {
			declaration := n.VariableDeclaration.AsVariableDeclaration()
			parameter = p.identifierWithAnnotation(declaration.Name(), declaration.Type)
		}
		return p.addNode(typeCatchClause, node, nodeValue(parameter), nodeValue(p.convert(n.Block, false)))
	case ast.KindLabeledStatement:
		n := node.AsLabeledStatement()
		return p.addNode(typeLabeledStatement, node,
			nodeValue(p.convert(n.Label, false)), nodeValue(p.convertStatement(n.Statement)))

	case ast.KindVariableStatement:
		n := node.AsVariableStatement()
		return p.convertVariableDeclarationList(n.DeclarationList, node)
	case ast.KindVariableDeclarationList:
		return p.convertVariableDeclarationList(node, node)
	case ast.KindVariableDeclaration:
		n := node.AsVariableDeclaration()
		return p.addNode(typeVariableDeclarator, node,
			nodeValue(p.identifierWithAnnotation(n.Name(), n.Type)),
			nodeValue(p.convert(n.Initializer, false)),
			boolValue(n.ExclamationToken != nil))
	case ast.KindFunctionDeclaration:
		return p.convertFunction(node, typeFunctionDeclaration)
	case ast.KindFunctionExpression:
		return p.convertFunction(node, typeFunctionExpression)
	case ast.KindArrowFunction:
		return p.convertArrowFunction(node)
	case ast.KindParameter:
		return p.convertParameter(node)

	case ast.KindIdentifier:
		return p.addNode(typeIdentifier, node,
			p.listValue(nil), p.stringValue(node.Text()), boolValue(false), undefinedValue())
	case ast.KindComputedPropertyName:
		return p.convert(node.AsComputedPropertyName().Expression, false)
	case ast.KindPrivateIdentifier:
		return p.addNode(typePrivateIdentifier, node, p.stringValue(strings.TrimPrefix(node.Text(), "#")))
	case ast.KindThisKeyword:
		return p.addNode(typeThisExpression, node)
	case ast.KindSuperKeyword:
		return p.addNode(typeSuper, node)

	case ast.KindStringLiteral:
		if node.Parent != nil && node.Parent.Kind == ast.KindJsxAttribute {
			raw := p.rawText(node)
			inner := ""
			if len(raw) >= 2 && raw[len(raw)-1] == raw[0] {
				inner = raw[1 : len(raw)-1]
			} else if len(raw) >= 1 {
				inner = raw[1:]
			}
			return p.addLiteral(node, p.stringValue(p.jsxValue(inner, false)), undefinedValue(), undefinedValue(), undefinedValue())
		}
		return p.addLiteral(node, p.stringValue(node.Text()), undefinedValue(), undefinedValue(), undefinedValue())
	case ast.KindNumericLiteral:
		return p.addLiteral(node, p.numberValue(parseNumberLiteral(node.Text())), undefinedValue(), undefinedValue(), undefinedValue())
	case ast.KindBigIntLiteral:
		raw := p.rawText(node)
		bigint := strings.TrimSuffix(strings.ReplaceAll(raw, "_", ""), "n")
		integer, ok := new(big.Int).SetString(bigint, 0)
		if !ok {
			failProjection(fmt.Errorf("invalid bigint literal %q", bigint))
		}
		bigint = integer.String()
		return p.addLiteral(node, nullValue(), p.stringValue(bigint), undefinedValue(), undefinedValue())
	case ast.KindRegularExpressionLiteral:
		raw := node.Text()
		lastSlash := strings.LastIndex(raw, "/")
		patternText, flags := "", ""
		if lastSlash > 0 {
			patternText, flags = raw[1:lastSlash], raw[lastSlash+1:]
		}
		return p.addLiteral(node, nullValue(), undefinedValue(), p.stringValue(patternText), p.stringValue(flags))
	case ast.KindTrueKeyword:
		return p.addLiteral(node, boolValue(true), undefinedValue(), undefinedValue(), undefinedValue())
	case ast.KindFalseKeyword:
		return p.addLiteral(node, boolValue(false), undefinedValue(), undefinedValue(), undefinedValue())
	case ast.KindNullKeyword:
		return p.addLiteral(node, nullValue(), undefinedValue(), undefinedValue(), undefinedValue())

	case ast.KindParenthesizedExpression:
		expression := node.AsParenthesizedExpression().Expression
		if ast.IsJSDocTypeAssertion(node) && expression != nil && expression.Kind == ast.KindAsExpression {
			// `/** @type {T} */ (value)` is ordinary JavaScript syntax. tsgo
			// internally reparses it as an AsExpression for checking, whereas the
			// ESTree exposed to ESLint rules contains only `value`.
			return p.convert(expression.AsAsExpression().Expression, pattern)
		}
		return p.convert(expression, pattern)
	case ast.KindBinaryExpression:
		return p.convertBinary(node, pattern)
	case ast.KindConditionalExpression:
		n := node.AsConditionalExpression()
		return p.addNode(typeConditionalExpression, node,
			nodeValue(p.convert(n.Condition, false)),
			nodeValue(p.convert(n.WhenTrue, false)),
			nodeValue(p.convert(n.WhenFalse, false)))
	case ast.KindPrefixUnaryExpression:
		n := node.AsPrefixUnaryExpression()
		operator := scanner.TokenToString(n.Operator)
		typeID := typeUnaryExpression
		if n.Operator == ast.KindPlusPlusToken || n.Operator == ast.KindMinusMinusToken {
			typeID = typeUpdateExpression
		}
		return p.addNode(typeID, node,
			p.stringValue(operator), boolValue(true), nodeValue(p.convert(n.Operand, false)))
	case ast.KindPostfixUnaryExpression:
		n := node.AsPostfixUnaryExpression()
		return p.addNode(typeUpdateExpression, node,
			p.stringValue(scanner.TokenToString(n.Operator)), boolValue(false), nodeValue(p.convert(n.Operand, false)))
	case ast.KindDeleteExpression, ast.KindVoidExpression, ast.KindTypeOfExpression:
		var operator string
		switch node.Kind {
		case ast.KindDeleteExpression:
			operator = "delete"
		case ast.KindVoidExpression:
			operator = "void"
		case ast.KindTypeOfExpression:
			operator = "typeof"
		}
		return p.addNode(typeUnaryExpression, node,
			p.stringValue(operator), boolValue(true), nodeValue(p.convert(node.Expression(), false)))
	case ast.KindAwaitExpression:
		return p.addNode(typeAwaitExpression, node, nodeValue(p.convert(node.AsAwaitExpression().Expression, false)))
	case ast.KindYieldExpression:
		n := node.AsYieldExpression()
		return p.addNode(typeYieldExpression, node,
			nodeValue(p.convert(n.Expression, false)), boolValue(n.AsteriskToken != nil))
	case ast.KindAsExpression:
		n := node.AsAsExpression()
		return p.addNode(typeTSAsExpression, node,
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convert(n.Type, false)))
	case ast.KindSatisfiesExpression:
		n := node.AsSatisfiesExpression()
		return p.addNode(typeTSSatisfiesExpression, node,
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convert(n.Type, false)))
	case ast.KindNonNullExpression:
		expression, chain := p.convertChainOperand(node.AsNonNullExpression().Expression)
		id := p.addNode(typeTSNonNullExpression, node, nodeValue(expression))
		return p.wrapOptionalChain(node, id, chain)
	case ast.KindTypeAssertionExpression:
		n := node.AsTypeAssertion()
		return p.addNode(typeTSTypeAssertion, node,
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convert(n.Type, false)))
	case ast.KindMetaProperty:
		n := node.AsMetaProperty()
		start, _ := p.rangeOf(node)
		keyword := scanner.TokenToString(n.KeywordToken)
		meta := p.addWithOrigin(typeIdentifier, nil, start, start+uint32(len(keyword)),
			p.listValue(nil), p.stringValue(keyword), boolValue(false), undefinedValue())
		if p.mapping != nil {
			// The parser maps this identifier to the MetaProperty's first token.
			if p.mapping.FirstTokens == nil {
				p.mapping.FirstTokens = make(map[uint32]uint32)
			}
			p.mapping.FirstTokens[meta] = p.mapping.table.GetIndex(node)
		}
		return p.addNode(typeMetaProperty, node,
			nodeValue(meta), nodeValue(p.convert(n.Name(), false)))

	case ast.KindPropertyAccessExpression:
		n := node.AsPropertyAccessExpression()
		object, chain := p.convertChainOperand(n.Expression)
		id := p.addNode(typeMemberExpression, node,
			nodeValue(object), nodeValue(p.convert(n.Name(), false)),
			boolValue(false), boolValue(n.QuestionDotToken != nil))
		return p.wrapOptionalChain(node, id, chain)
	case ast.KindElementAccessExpression:
		n := node.AsElementAccessExpression()
		object, chain := p.convertChainOperand(n.Expression)
		id := p.addNode(typeMemberExpression, node,
			nodeValue(object), nodeValue(p.convert(n.ArgumentExpression, false)),
			boolValue(true), boolValue(n.QuestionDotToken != nil))
		return p.wrapOptionalChain(node, id, chain)
	case ast.KindCallExpression:
		return p.convertCall(node)
	case ast.KindNewExpression:
		n := node.AsNewExpression()
		arguments := []*ast.Node(nil)
		if n.Arguments != nil {
			arguments = n.Arguments.Nodes
		}
		return p.addNode(typeNewExpression, node,
			nodeValue(p.convert(n.Expression, false)), p.convertList(arguments, false),
			p.typeParameters(n.TypeArguments, false))
	case ast.KindArrayLiteralExpression:
		elements := p.convertListWithNulls(node.AsArrayLiteralExpression().Elements.Nodes, pattern)
		if pattern {
			return p.addNode(typeArrayPattern, node, elements, p.listValue(nil), boolValue(false), undefinedValue())
		}
		return p.addNode(typeArrayExpression, node, elements)
	case ast.KindObjectLiteralExpression:
		properties := p.convertList(node.AsObjectLiteralExpression().Properties.Nodes, pattern)
		if pattern {
			return p.addNode(typeObjectPattern, node, properties, p.listValue(nil), boolValue(false), undefinedValue())
		}
		return p.addNode(typeObjectExpression, node, properties)
	case ast.KindPropertyAssignment:
		n := node.AsPropertyAssignment()
		name := n.Name()
		return p.addNode(typeProperty, node,
			nodeValue(p.convert(name, false)), nodeValue(p.convert(n.Initializer, pattern)),
			p.stringValue("init"), boolValue(false), boolValue(false),
			boolValue(name != nil && name.Kind == ast.KindComputedPropertyName), boolValue(false))
	case ast.KindShorthandPropertyAssignment:
		n := node.AsShorthandPropertyAssignment()
		name := n.Name()
		value := p.convert(name, false)
		if n.ObjectAssignmentInitializer != nil {
			value = p.addNode(typeAssignmentPattern, node,
				p.listValue(nil), nodeValue(value), boolValue(false),
				nodeValue(p.convert(n.ObjectAssignmentInitializer, false)), undefinedValue())
		}
		return p.addNode(typeProperty, node,
			nodeValue(p.convert(name, false)), nodeValue(value), p.stringValue("init"),
			boolValue(false), boolValue(true), boolValue(false), boolValue(false))
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		if pattern {
			return p.addNode(typeRestElement, node,
				p.listValue(nil), nodeValue(p.convert(node.Expression(), true)),
				boolValue(false), undefinedValue())
		}
		return p.addNode(typeSpreadElement, node, nodeValue(p.convert(node.Expression(), false)))
	case ast.KindObjectBindingPattern:
		return p.addNode(typeObjectPattern, node,
			p.convertList(node.AsBindingPattern().Elements.Nodes, true),
			p.listValue(nil), boolValue(false), undefinedValue())
	case ast.KindArrayBindingPattern:
		return p.addNode(typeArrayPattern, node,
			p.convertListWithNulls(node.AsBindingPattern().Elements.Nodes, true),
			p.listValue(nil), boolValue(false), undefinedValue())
	case ast.KindBindingElement:
		return p.convertBindingElement(node)

	case ast.KindNoSubstitutionTemplateLiteral:
		return p.convertNoSubstitutionTemplate(node)
	case ast.KindTemplateExpression:
		return p.convertTemplateExpression(node)
	case ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail:
		return p.convertTemplateElement(node)
	case ast.KindTaggedTemplateExpression:
		n := node.AsTaggedTemplateExpression()
		return p.addNode(typeTaggedTemplateExpression, node,
			nodeValue(p.convert(n.Tag, false)), nodeValue(p.convert(n.Template, false)),
			p.typeParameters(n.TypeArguments, false))

	case ast.KindClassDeclaration:
		return p.convertClass(node, true)
	case ast.KindClassExpression:
		return p.convertClass(node, false)
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		if node.Parent != nil && (node.Parent.Kind == ast.KindTypeLiteral || node.Parent.Kind == ast.KindInterfaceDeclaration) {
			return p.convertMethodSignature(node)
		}
		return p.convertMethod(node)
	case ast.KindPropertyDeclaration:
		return p.convertPropertyDefinition(node)
	case ast.KindClassStaticBlockDeclaration:
		body := node.AsClassStaticBlockDeclaration().Body
		return p.addNode(typeStaticBlock, node, p.convertStatements(body.AsBlock().Statements, !p.isJS))

	case ast.KindDecorator:
		return p.addNode(typeDecorator, node, nodeValue(p.convert(node.AsDecorator().Expression, false)))
	case ast.KindNamespaceExportDeclaration:
		return p.addNode(typeTSNamespaceExportDeclaration, node, nodeValue(p.convert(node.Name(), false)))
	case ast.KindImportDeclaration:
		return p.convertImportDeclaration(node)
	case ast.KindNamespaceImport:
		return p.addNode(typeImportNamespaceSpecifier, node,
			nodeValue(p.convert(node.AsNamespaceImport().Name(), false)))
	case ast.KindImportSpecifier:
		n := node.AsImportSpecifier()
		imported := n.PropertyName
		if imported == nil {
			imported = n.Name()
		}
		kind := "value"
		if n.IsTypeOnly {
			kind = "type"
		}
		return p.addNode(typeImportSpecifier, node,
			nodeValue(p.convert(imported, false)), p.stringValue(kind), nodeValue(p.convert(n.Name(), false)))
	case ast.KindImportClause:
		name := node.AsImportClause().Name()
		start, end := p.rangeOf(name)
		return p.addWithOrigin(typeImportDefaultSpecifier, node, start, end, nodeValue(p.convert(name, false)))
	case ast.KindImportAttribute:
		n := node.AsImportAttribute()
		return p.addNode(typeImportAttribute, node,
			nodeValue(p.convert(n.Name(), false)), nodeValue(p.convert(n.Value, false)))
	case ast.KindImportEqualsDeclaration:
		n := node.AsImportEqualsDeclaration()
		importKind := "value"
		if n.IsTypeOnly {
			importKind = "type"
		}
		return p.addNode(typeTSImportEqualsDeclaration, node,
			nodeValue(p.convert(n.Name(), false)), p.stringValue(importKind),
			nodeValue(p.convert(n.ModuleReference, false)))
	case ast.KindExternalModuleReference:
		return p.addNode(typeTSExternalModuleReference, node,
			nodeValue(p.convert(node.AsExternalModuleReference().Expression, false)))
	case ast.KindExportDeclaration:
		return p.convertExportDeclaration(node)
	case ast.KindExportSpecifier:
		n := node.AsExportSpecifier()
		local := n.PropertyName
		if local == nil {
			local = n.Name()
		}
		kind := "value"
		if n.IsTypeOnly {
			kind = "type"
		}
		return p.addNode(typeExportSpecifier, node,
			nodeValue(p.convert(n.Name(), false)), p.stringValue(kind), nodeValue(p.convert(local, false)))
	case ast.KindExportAssignment:
		n := node.AsExportAssignment()
		if n.IsExportEquals {
			return p.addNode(typeTSExportAssignment, node, nodeValue(p.convert(n.Expression, false)))
		}
		return p.addNode(typeExportDefaultDeclaration, node,
			nodeValue(p.convert(n.Expression, false)), p.stringValue("value"))

	case ast.KindInterfaceDeclaration:
		return p.convertInterface(node)
	case ast.KindTypeAliasDeclaration, ast.KindJSTypeAliasDeclaration:
		n := node.AsTypeAliasDeclaration()
		return p.addNode(typeTSTypeAliasDeclaration, node,
			boolValue(p.hasModifier(node, ast.ModifierFlagsAmbient)),
			nodeValue(p.convert(n.Name(), false)), nodeValue(p.convert(n.Type, false)),
			p.typeParameters(n.TypeParameters, true))
	case ast.KindPropertySignature:
		return p.convertPropertySignature(node)
	case ast.KindMethodSignature:
		return p.convertMethodSignature(node)
	case ast.KindIndexSignature:
		n := node.AsIndexSignatureDeclaration()
		return p.addNode(typeTSIndexSignature, node,
			p.accessibility(node), p.convertList(n.Parameters.Nodes, false),
			boolValue(p.hasModifier(node, ast.ModifierFlagsReadonly)),
			boolValue(p.hasModifier(node, ast.ModifierFlagsStatic)), p.typeAnnotation(n.Type))
	case ast.KindCallSignature:
		n := node.AsCallSignatureDeclaration()
		parameters := []*ast.Node(nil)
		if n.Parameters != nil {
			parameters = n.Parameters.Nodes
		}
		return p.addNode(typeTSCallSignatureDeclaration, node,
			p.convertParameters(parameters), p.typeAnnotation(n.Type),
			p.typeParameters(n.TypeParameters, true))
	case ast.KindConstructSignature:
		n := node.AsConstructSignatureDeclaration()
		parameters := []*ast.Node(nil)
		if n.Parameters != nil {
			parameters = n.Parameters.Nodes
		}
		return p.addNode(typeTSConstructSignatureDeclaration, node,
			p.convertParameters(parameters), p.typeAnnotation(n.Type),
			p.typeParameters(n.TypeParameters, true))
	case ast.KindTypeReference:
		n := node.AsTypeReferenceNode()
		return p.addNode(typeTSTypeReference, node,
			nodeValue(p.convert(n.TypeName, false)), p.typeParameters(n.TypeArguments, false))
	case ast.KindQualifiedName:
		n := node.AsQualifiedName()
		return p.addNode(typeTSQualifiedName, node,
			nodeValue(p.convert(n.Left, false)), nodeValue(p.convert(n.Right, false)))
	case ast.KindTypeLiteral:
		return p.addNode(typeTSTypeLiteral, node, p.convertList(node.AsTypeLiteralNode().Members.Nodes, false))
	case ast.KindUnionType:
		return p.addNode(typeTSUnionType, node, p.convertList(node.AsUnionTypeNode().Types.Nodes, false))
	case ast.KindIntersectionType:
		return p.addNode(typeTSIntersectionType, node, p.convertList(node.AsIntersectionTypeNode().Types.Nodes, false))
	case ast.KindConditionalType:
		n := node.AsConditionalTypeNode()
		return p.addNode(typeTSConditionalType, node,
			nodeValue(p.convert(n.CheckType, false)), nodeValue(p.convert(n.ExtendsType, false)),
			nodeValue(p.convert(n.FalseType, false)), nodeValue(p.convert(n.TrueType, false)))
	case ast.KindMappedType:
		return p.convertMappedType(node)
	case ast.KindInferType:
		return p.addNode(typeTSInferType, node,
			nodeValue(p.convert(node.AsInferTypeNode().TypeParameter, false)))
	case ast.KindThisType:
		return p.addNode(typeTSThisType, node)
	case ast.KindArrayType:
		return p.addNode(typeTSArrayType, node, nodeValue(p.convert(node.AsArrayTypeNode().ElementType, false)))
	case ast.KindLiteralType:
		if node.AsLiteralTypeNode().Literal.Kind == ast.KindNullKeyword {
			return p.addNode(typeTSNullKeyword, node.AsLiteralTypeNode().Literal)
		}
		return p.addNode(typeTSLiteralType, node, nodeValue(p.convert(node.AsLiteralTypeNode().Literal, false)))
	case ast.KindTypeParameter:
		n := node.AsTypeParameterDeclaration()
		return p.addNode(typeTSTypeParameter, node,
			boolValue(p.hasModifier(node, ast.ModifierFlagsConst)),
			optionalNodeValue(p.convert(n.Constraint, false)), optionalNodeValue(p.convert(n.DefaultType, false)),
			boolValue(p.hasModifier(node, ast.ModifierFlagsIn)), nodeValue(p.convert(n.Name(), false)),
			boolValue(p.hasModifier(node, ast.ModifierFlagsOut)))
	case ast.KindStringKeyword:
		return p.addNode(typeTSStringKeyword, node)
	case ast.KindNumberKeyword:
		return p.addNode(typeTSNumberKeyword, node)
	case ast.KindBooleanKeyword:
		return p.addNode(typeTSBooleanKeyword, node)
	case ast.KindAnyKeyword:
		return p.addNode(typeTSAnyKeyword, node)
	case ast.KindUnknownKeyword:
		return p.addNode(typeTSUnknownKeyword, node)
	case ast.KindVoidKeyword:
		return p.addNode(typeTSVoidKeyword, node)
	case ast.KindUndefinedKeyword:
		return p.addNode(typeTSUndefinedKeyword, node)
	case ast.KindNeverKeyword:
		return p.addNode(typeTSNeverKeyword, node)
	case ast.KindObjectKeyword:
		return p.addNode(typeTSObjectKeyword, node)
	case ast.KindSymbolKeyword:
		return p.addNode(typeTSSymbolKeyword, node)
	case ast.KindBigIntKeyword:
		return p.addNode(typeTSBigIntKeyword, node)
	case ast.KindIntrinsicKeyword:
		return p.addNode(typeTSIntrinsicKeyword, node)
	case ast.KindFunctionType:
		return p.convertFunctionType(node, false)
	case ast.KindConstructorType:
		return p.convertFunctionType(node, true)
	case ast.KindTupleType:
		return p.addNode(typeTSTupleType, node, p.convertList(node.AsTupleTypeNode().Elements.Nodes, false))
	case ast.KindNamedTupleMember:
		n := node.AsNamedTupleMember()
		id := p.addNode(typeTSNamedTupleMember, node,
			nodeValue(p.convert(n.Type, false)), nodeValue(p.convert(n.Name(), false)), boolValue(n.QuestionToken != nil))
		if n.DotDotDotToken != nil {
			p.nodes[id-1].start = p.offsets.At(p.tokenStart(n.Name()))
			return p.addNode(typeTSRestType, node, nodeValue(id))
		}
		return id
	case ast.KindOptionalType:
		return p.addNode(typeTSOptionalType, node, nodeValue(p.convert(node.AsOptionalTypeNode().Type, false)))
	case ast.KindRestType:
		return p.addNode(typeTSRestType, node, nodeValue(p.convert(node.AsRestTypeNode().Type, false)))
	case ast.KindParenthesizedType:
		return p.convert(node.AsParenthesizedTypeNode().Type, false)
	case ast.KindIndexedAccessType:
		n := node.AsIndexedAccessTypeNode()
		return p.addNode(typeTSIndexedAccessType, node,
			nodeValue(p.convert(n.ObjectType, false)), nodeValue(p.convert(n.IndexType, false)))
	case ast.KindTypeQuery:
		n := node.AsTypeQueryNode()
		return p.addNode(typeTSTypeQuery, node,
			nodeValue(p.convertTypeQueryName(n.ExprName)), p.typeParameters(n.TypeArguments, false))
	case ast.KindTypeOperator:
		n := node.AsTypeOperatorNode()
		return p.addNode(typeTSTypeOperator, node,
			p.stringValue(scanner.TokenToString(n.Operator)), nodeValue(p.convert(n.Type, false)))
	case ast.KindTypePredicate:
		n := node.AsTypePredicateNode()
		annotation := nullValue()
		if n.Type != nil {
			annotation = p.typeAnnotation(n.Type)
		}
		return p.addNode(typeTSTypePredicate, node,
			boolValue(n.AssertsModifier != nil), nodeValue(p.convert(n.ParameterName, false)), annotation)
	case ast.KindImportType:
		return p.convertImportType(node)
	case ast.KindTemplateLiteralType:
		return p.convertTemplateLiteralType(node)
	case ast.KindTemplateLiteralTypeSpan, ast.KindOmittedExpression:
		return 0
	case ast.KindExpressionWithTypeArguments:
		n := node.AsExpressionWithTypeArguments()
		return p.addNode(typeTSInstantiationExpression, node,
			nodeValue(p.convert(n.Expression, false)), p.typeParameters(n.TypeArguments, false))

	case ast.KindEnumDeclaration:
		return p.convertEnum(node)
	case ast.KindEnumMember:
		n := node.AsEnumMember()
		return p.addNode(typeTSEnumMember, node,
			nodeValue(p.convert(n.Name(), false)), optionalNodeValue(p.convert(n.Initializer, false)))
	case ast.KindModuleDeclaration:
		return p.convertModuleDeclaration(node)

	case ast.KindJsxElement:
		n := node.AsJsxElement()
		return p.addNode(typeJSXElement, node,
			nodeValue(p.convert(n.OpeningElement, false)), p.convertList(n.Children.Nodes, false),
			nodeValue(p.convert(n.ClosingElement, false)))
	case ast.KindJsxSelfClosingElement:
		return p.convertJsxSelfClosing(node)
	case ast.KindJsxOpeningElement:
		n := node.AsJsxOpeningElement()
		return p.addNode(typeJSXOpeningElement, node,
			nodeValue(p.convertJsxName(n.TagName)), p.convertList(n.Attributes.AsJsxAttributes().Properties.Nodes, false),
			boolValue(false), p.typeParameters(n.TypeArguments, false))
	case ast.KindJsxClosingElement:
		return p.addNode(typeJSXClosingElement, node,
			nodeValue(p.convertJsxName(node.AsJsxClosingElement().TagName)))
	case ast.KindJsxFragment:
		n := node.AsJsxFragment()
		return p.addNode(typeJSXFragment, node,
			nodeValue(p.convert(n.OpeningFragment, false)), p.convertList(n.Children.Nodes, false),
			nodeValue(p.convert(n.ClosingFragment, false)))
	case ast.KindJsxOpeningFragment:
		attributes, selfClosing := undefinedValue(), undefinedValue()
		if p.isJS {
			attributes, selfClosing = p.listValue(nil), boolValue(false)
		}
		return p.addNode(typeJSXOpeningFragment, node, attributes, selfClosing)
	case ast.KindJsxClosingFragment:
		return p.addNode(typeJSXClosingFragment, node)
	case ast.KindJsxAttribute:
		n := node.AsJsxAttribute()
		return p.addNode(typeJSXAttribute, node,
			nodeValue(p.convertJsxName(n.Name())), nodeValue(p.convert(n.Initializer, false)))
	case ast.KindJsxSpreadAttribute:
		return p.addNode(typeJSXSpreadAttribute, node,
			nodeValue(p.convert(node.AsJsxSpreadAttribute().Expression, false)))
	case ast.KindJsxExpression:
		return p.convertJsxExpression(node)
	case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces:
		raw := p.rawTextFull(node)
		return p.addNode(typeJSXText, node, p.stringValue(p.jsxValue(raw, true)), p.stringValue(raw))
	case ast.KindJsxNamespacedName:
		n := node.AsJsxNamespacedName()
		return p.addNode(typeJSXNamespacedName, node,
			nodeValue(p.convertJsxName(n.Namespace)), nodeValue(p.convertJsxName(n.Name())))
	}

	return p.unsupported(node)
}

func (p *estreeProjector) convertStatements(list *ast.NodeList, prologue bool) uint32 {
	if list == nil || len(list.Nodes) == 0 {
		return p.listValue(nil)
	}
	base := len(p.listScratch)
	for _, statement := range list.Nodes {
		// Advance the prologue once instead of rescanning every preceding
		// string statement.
		prologue = prologue && statement.Kind == ast.KindExpressionStatement &&
			statement.AsExpressionStatement().Expression != nil &&
			statement.AsExpressionStatement().Expression.Kind == ast.KindStringLiteral
		var id uint32
		if prologue {
			id = p.convertExpressionStatement(statement, true)
			p.registerReverse(statement, id)
		} else {
			id = p.convertStatement(statement)
		}
		if id != 0 {
			p.listScratch = append(p.listScratch, id)
		}
	}
	res := p.listValue(p.listScratch[base:])
	p.listScratch = p.listScratch[:base]
	return res
}

func (p *estreeProjector) convertExpressionStatement(node *ast.Node, prologue bool) uint32 {
	p.check()
	expression := node.AsExpressionStatement().Expression
	directive := undefinedValue()
	if prologue && expression != nil {
		raw := p.rawText(expression)
		if len(raw) >= 2 && raw[len(raw)-1] == raw[0] {
			directive = p.stringValue(raw[1 : len(raw)-1])
		} else if len(raw) >= 1 {
			directive = p.stringValue(raw[1:])
		}
	}
	return p.addNode(typeExpressionStatement, node, nodeValue(p.convert(expression, false)), directive)
}

func (p *estreeProjector) convertStatement(node *ast.Node) uint32 {
	if node == nil {
		return 0
	}
	id := p.convert(node, false)
	if id == 0 || node.Kind == ast.KindModuleDeclaration || !p.hasModifier(node, ast.ModifierFlagsExport) {
		return id
	}
	return p.wrapExport(node, node, id)
}

func (p *estreeProjector) wrapExport(node, origin *ast.Node, declaration uint32) uint32 {
	// The wrapper owns export/default; the declaration begins at its own syntax.
	start := p.tokenStart(node)
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier.Kind != ast.KindExportKeyword && modifier.Kind != ast.KindDefaultKeyword {
				continue
			}
			start = max(start, scanner.SkipTrivia(p.text, modifier.End()))
		}
	}
	p.nodes[declaration-1].start = p.offsets.At(start)

	kind := "value"
	if node.Kind == ast.KindInterfaceDeclaration || node.Kind == ast.KindTypeAliasDeclaration || p.hasModifier(node, ast.ModifierFlagsAmbient) {
		kind = "type"
	}
	var result uint32
	if p.hasModifier(node, ast.ModifierFlagsDefault) {
		result = p.addNode(typeExportDefaultDeclaration, origin, nodeValue(declaration), p.stringValue("value"))
	} else {
		result = p.addNode(typeExportNamedDeclaration, origin, p.listValue(nil), nodeValue(declaration), p.stringValue(kind), nullValue(), p.listValue(nil))
	}
	for _, modifier := range node.Modifiers().Nodes {
		if modifier.Kind == ast.KindExportKeyword {
			p.nodes[result-1].start = p.offsets.At(p.tokenStart(modifier))
			break
		}
	}
	return result
}

func (p *estreeProjector) convertVariableDeclarationList(node, rangeNode *ast.Node) uint32 {
	if node == nil {
		return 0
	}
	n := node.AsVariableDeclarationList()
	kind := "var"
	flags := node.Flags & ast.NodeFlagsBlockScoped
	switch flags {
	case ast.NodeFlagsAwaitUsing:
		kind = "await using"
	case ast.NodeFlagsUsing:
		kind = "using"
	case ast.NodeFlagsConst:
		kind = "const"
	case ast.NodeFlagsLet:
		kind = "let"
	}
	// A standalone ESTree VariableDeclaration spans the VariableStatement,
	// including its trailing semicolon. tsgo keeps that semicolon on the outer
	// VariableStatement while DeclarationList.End stops before it. For-loop
	// declaration lists have no outer statement and therefore use their own
	// range. Token-sensitive rules such as @stylistic/semi depend on this exact
	// distinction.
	return p.addNode(typeVariableDeclaration, rangeNode,
		p.convertList(n.Declarations.Nodes, false),
		boolValue(node.Parent != nil && p.hasModifier(node.Parent, ast.ModifierFlagsAmbient)),
		p.stringValue(kind))
}

func (p *estreeProjector) addLiteral(node *ast.Node, value, bigint, regexPattern, regexFlags uint32) uint32 {
	start := p.tokenStart(node)
	raw := ""
	if start >= 0 && node.End() >= start && node.End() <= len(p.text) {
		raw = p.text[start:node.End()]
	}
	return p.addWithOrigin(typeLiteral, node, p.offsets.At(start), p.offsets.At(node.End()),
		value, p.stringValue(raw), bigint, regexPattern, regexFlags)
}

func (p *estreeProjector) rawText(node *ast.Node) string {
	start := p.tokenStart(node)
	if start < 0 || node.End() < start || node.End() > len(p.text) {
		return ""
	}
	return p.text[start:node.End()]
}

func (p *estreeProjector) rawTextFull(node *ast.Node) string {
	if node.Pos() < 0 || node.End() < node.Pos() || node.End() > len(p.text) {
		return ""
	}
	return p.text[node.Pos():node.End()]
}

func (p *estreeProjector) unsupported(node *ast.Node) uint32 {
	failProjection(fmt.Errorf("ESTree projection: unsupported syntax %s at byte %d", node.Kind, node.Pos()))
	return 0
}

func (p *estreeProjector) wrapOptionalChain(node *ast.Node, id uint32, chain bool) uint32 {
	if !chain && node.Flags&ast.NodeFlagsOptionalChain == 0 {
		return id
	}
	wrapper := p.addNode(typeChainExpression, node, nodeValue(id))
	if p.mapping != nil {
		// A chain operand can reclaim only the last record it emitted.
		// Track every reverse key for that wrapper, including transparent nodes.
		p.mapping.chainRecord = wrapper
		p.mapping.chainReverse = p.mapping.chainReverse[:0]
	}
	return wrapper
}

func (p *estreeProjector) convertFunction(node *ast.Node, requestedType TypeID) uint32 {
	function := node.FunctionLikeData()
	body := node.BodyData()
	typeID := requestedType
	if requestedType == typeFunctionDeclaration && (body == nil || body.Body == nil) {
		typeID = typeTSDeclareFunction
	}
	parameters := []*ast.Node(nil)
	if function.Parameters != nil {
		parameters = function.Parameters.Nodes
	}
	asterisk := body != nil && body.AsteriskToken != nil
	bodyNode := (*ast.Node)(nil)
	if body != nil {
		bodyNode = body.Body
	}
	bodyVal := nodeValue(p.convert(bodyNode, false))
	if typeID == typeTSDeclareFunction && bodyNode == nil {
		bodyVal = undefinedValue()
	}
	return p.addNode(typeID, node,
		nodeValue(p.convert(node.Name(), false)),
		p.convertParameters(parameters),
		bodyVal,
		boolValue(asterisk),
		boolValue(p.hasModifier(node, ast.ModifierFlagsAsync)),
		boolValue(false),
		boolValue(p.hasModifier(node, ast.ModifierFlagsAmbient)),
		p.typeParameters(function.TypeParameters, true),
		p.typeAnnotation(function.Type),
	)
}

func (p *estreeProjector) convertArrowFunction(node *ast.Node) uint32 {
	n := node.AsArrowFunction()
	parameters := []*ast.Node(nil)
	if n.Parameters != nil {
		parameters = n.Parameters.Nodes
	}
	return p.addNode(typeArrowFunctionExpression, node,
		nullValue(), p.convertParameters(parameters), nodeValue(p.convert(n.Body, false)),
		boolValue(false), boolValue(p.hasModifier(node, ast.ModifierFlagsAsync)),
		boolValue(n.Body != nil && n.Body.Kind != ast.KindBlock),
		p.typeParameters(n.TypeParameters, true), p.typeAnnotation(n.Type))
}

func (p *estreeProjector) convertParameters(parameters []*ast.Node) uint32 {
	if len(parameters) == 0 {
		return p.listValue(nil)
	}
	base := len(p.listScratch)
	for _, parameter := range parameters {
		if id := p.convertParameter(parameter); id != 0 {
			p.registerReverse(parameter, id)
			p.listScratch = append(p.listScratch, id)
		}
	}
	res := p.listValue(p.listScratch[base:])
	p.listScratch = p.listScratch[:base]
	return res
}

func (p *estreeProjector) convertParameter(node *ast.Node) uint32 {
	previous := p.origin
	if p.mapping != nil {
		p.origin = node
	}
	id := p.convertParameterNode(node)
	p.origin = previous
	p.registerReverse(node, id)
	return id
}

func (p *estreeProjector) convertParameterNode(node *ast.Node) uint32 {
	if node == nil || isJSDocCommentSyntax(node) {
		return 0
	}
	n := node.AsParameterDeclaration()
	var parameter uint32
	if n.DotDotDotToken == nil {
		parameter = p.identifierWithAnnotation(n.Name(), n.Type)
	} else {
		parameter = p.convert(n.Name(), true)
	}
	if parameter == 0 {
		return 0
	}
	optional := n.QuestionToken != nil && !isJSDocCommentSyntax(n.QuestionToken)
	if optional && n.DotDotDotToken == nil {
		record := p.nodes[parameter-1]
		if slot := fieldSlot[record.typeID][fieldOptional]; slot >= 0 {
			p.values[int(record.valueStart)+int(slot)] = boolValue(true)
		}
		p.nodes[parameter-1].end = max(p.nodes[parameter-1].end, p.offsets.At(n.QuestionToken.End()))
	}
	result := parameter
	if n.DotDotDotToken != nil {
		result = p.addNode(typeRestElement, node,
			p.listValue(nil), nodeValue(parameter), boolValue(optional),
			p.typeAnnotation(n.Type))
	} else if n.Initializer != nil {
		result = p.addNode(typeAssignmentPattern, node,
			p.listValue(nil), nodeValue(parameter), boolValue(optional),
			nodeValue(p.convert(n.Initializer, false)), undefinedValue())
	}
	// Accessibility and readonly belong to TSParameterProperty, not its binding.
	p.nodes[result-1].start = p.nodes[parameter-1].start
	if n.DotDotDotToken != nil {
		p.nodes[result-1].start = p.offsets.At(p.tokenStart(n.DotDotDotToken))
	}
	if utils.ESTreeModifierFlags(node)&ast.ModifierFlagsParameterPropertyModifier != 0 {
		return p.addNode(typeTSParameterProperty, node,
			p.accessibility(node), p.decorators(node),
			boolValue(p.hasModifier(node, ast.ModifierFlagsOverride)), nodeValue(result),
			boolValue(p.hasModifier(node, ast.ModifierFlagsReadonly)),
			boolValue(p.hasModifier(node, ast.ModifierFlagsStatic)))
	}
	if slot := fieldSlot[p.nodes[result-1].typeID][fieldDecorators]; slot >= 0 {
		decorators := p.decorators(node)
		p.values[int(p.nodes[result-1].valueStart)+int(slot)] = decorators
	}
	return result
}

func (p *estreeProjector) convertBindingElement(node *ast.Node) uint32 {
	n := node.AsBindingElement()
	// tsgo represents an elision in an array binding pattern (for example the
	// first slot in `const [, value] = tuple`) as an empty BindingElement on
	// some parser paths. ESTree represents that slot as a literal null in the
	// ArrayPattern.elements array, not as a Property with null key/value.
	if n.Name() == nil {
		return 0
	}
	value := p.convert(n.Name(), true)
	if n.DotDotDotToken != nil {
		return p.addNode(typeRestElement, node,
			p.listValue(nil), nodeValue(value), boolValue(false), undefinedValue())
	}
	if n.Initializer != nil && value != 0 {
		start := p.nodes[value-1].start
		value = p.addNode(typeAssignmentPattern, node,
			p.listValue(nil), nodeValue(value), boolValue(false),
			nodeValue(p.convert(n.Initializer, false)), undefinedValue())
		// A renamed object's BindingElement includes its source property key;
		// only the binding value belongs to the AssignmentPattern.
		p.nodes[value-1].start = start
	}
	// Object binding elements become ESTree Property nodes, but array binding
	// elements are patterns directly. tsgo uses BindingElement for both shapes;
	// wrapping `[value]` as a Property would expose an AST shape no ESLint rule
	// expects and would also turn array defaults into Property(AssignmentPattern).
	if node.Parent != nil && node.Parent.Kind == ast.KindArrayBindingPattern {
		return value
	}
	key := n.PropertyName
	if key == nil {
		key = n.Name()
	}
	return p.addNode(typeProperty, node,
		nodeValue(p.convert(key, false)), nodeValue(value), p.stringValue("init"),
		boolValue(false), boolValue(n.PropertyName == nil),
		boolValue(n.PropertyName != nil && n.PropertyName.Kind == ast.KindComputedPropertyName), boolValue(false))
}

func (p *estreeProjector) convertBinary(node *ast.Node, pattern bool) uint32 {
	n := node.AsBinaryExpression()
	p.recordTokenCover(n.OperatorToken, false)
	operator := scanner.TokenToString(n.OperatorToken.Kind)
	if n.OperatorToken.Kind == ast.KindCommaToken {
		// Collect the native left spine before projecting it. Syntax-only
		// callers avoid detached prefixes; semantic maps preserve their aliases.
		rights := []*ast.Node{n.Right}
		var prefixes []*ast.Node
		left := n.Left
		for left.Kind == ast.KindBinaryExpression && left.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
			p.check()
			nested := left.AsBinaryExpression()
			if p.mapping != nil {
				prefixes = append(prefixes, left)
			}
			rights = append(rights, nested.Right)
			left = nested.Left
		}
		// Parenthesized operands stay separate sequences. Convert in source
		// order, just as for other expression lists.
		base := len(p.listScratch)
		leftID := p.convert(left, false)
		p.listScratch = append(p.listScratch, leftID)
		for index := len(rights) - 1; index >= 0; index-- {
			rightID := p.convert(rights[index], false)
			p.listScratch = append(p.listScratch, rightID)
		}
		listVal := p.listValue(p.listScratch[base:])
		id := p.addNode(typeSequenceExpression, node, listVal)
		if len(prefixes) != 0 {
			items := make([]uint32, len(p.listScratch)-base)
			copy(items, p.listScratch[base:])
			m := p.mapping
			if m.SequenceMembers == nil {
				m.SequenceMembers = make(map[uint32][]uint32)
			}
			m.SequenceMembers[id] = items
			for i, prefix := range prefixes {
				start, end := p.rangeOf(prefix)
				m.Aliases = append(m.Aliases, SequenceAlias{m.table.GetIndex(prefix), id, uint32(len(items) - i - 1), start, end})
			}
		}
		p.listScratch = p.listScratch[:base]
		return id
	}
	typeID := typeBinaryExpression
	if n.OperatorToken.Kind >= ast.KindFirstAssignment && n.OperatorToken.Kind <= ast.KindLastAssignment {
		typeID = typeAssignmentExpression
	} else if n.OperatorToken.Kind == ast.KindAmpersandAmpersandToken ||
		n.OperatorToken.Kind == ast.KindBarBarToken ||
		n.OperatorToken.Kind == ast.KindQuestionQuestionToken {
		typeID = typeLogicalExpression
	}
	if pattern && typeID == typeAssignmentExpression {
		return p.addNode(typeAssignmentPattern, node,
			p.listValue(nil), nodeValue(p.convert(n.Left, true)), boolValue(false),
			nodeValue(p.convert(n.Right, false)), undefinedValue())
	}
	return p.addNode(typeID, node,
		nodeValue(p.convert(n.Left, typeID == typeAssignmentExpression)),
		p.stringValue(operator), nodeValue(p.convert(n.Right, false)))
}

func (p *estreeProjector) convertCall(node *ast.Node) uint32 {
	n := node.AsCallExpression()
	arguments := []*ast.Node(nil)
	if n.Arguments != nil {
		arguments = n.Arguments.Nodes
	}
	if n.Expression != nil && n.Expression.Kind == ast.KindImportKeyword {
		var source, options uint32
		if len(arguments) > 0 {
			source = p.convert(arguments[0], false)
		}
		if len(arguments) > 1 {
			options = p.convert(arguments[1], false)
		}
		return p.addNode(typeImportExpression, node, nodeValue(source), nodeValue(options))
	}
	callee, chain := p.convertChainOperand(n.Expression)
	id := p.addNode(typeCallExpression, node,
		nodeValue(callee), p.convertList(arguments, false), boolValue(n.QuestionDotToken != nil),
		p.typeParameters(n.TypeArguments, false))
	return p.wrapOptionalChain(node, id, chain)
}

func (p *estreeProjector) convertNoSubstitutionTemplate(node *ast.Node) uint32 {
	start, end := p.rangeOf(node)
	raw := p.rawText(node)
	if len(raw) >= 2 && raw[len(raw)-1] == '`' {
		raw = raw[1 : len(raw)-1]
	} else if len(raw) >= 1 && raw[0] == '`' {
		raw = raw[1:]
	}
	element := p.add(typeTemplateElement, start, end,
		boolValue(true), p.templateCooked(node), p.stringValue(templateRaw(node, raw)))
	return p.addNode(typeTemplateLiteral, node, p.listValue(nil), p.listValue([]uint32{element}))
}

func (p *estreeProjector) convertTemplateExpression(node *ast.Node) uint32 {
	n := node.AsTemplateExpression()
	var exprBuf [8]uint32
	var quasiBuf [8]uint32
	expressions := exprBuf[:0]
	quasis := quasiBuf[:0]
	quasis = append(quasis, p.convert(n.Head, false))
	for _, spanNode := range n.TemplateSpans.Nodes {
		span := spanNode.AsTemplateSpan()
		expressions = append(expressions, p.convert(span.Expression, false))
		quasis = append(quasis, p.convert(span.Literal, false))
	}
	return p.addNode(typeTemplateLiteral, node, p.listValue(expressions), p.listValue(quasis))
}

func (p *estreeProjector) convertTemplateElement(node *ast.Node) uint32 {
	tail := node.Kind == ast.KindTemplateTail
	raw := node.RawText()
	return p.addNode(typeTemplateElement, node,
		boolValue(tail), p.templateCooked(node), p.stringValue(templateRaw(node, raw)))
}

func templateRaw(node *ast.Node, raw string) string {
	// TypeScript template-literal types preserve source newlines in raw;
	// expression templates use ECMAScript's normalized template raw value.
	if node.Parent != nil {
		switch node.Parent.Kind {
		case ast.KindLiteralType, ast.KindTemplateLiteralType, ast.KindTemplateLiteralTypeSpan:
			return raw
		}
	}
	if strings.IndexByte(raw, '\r') < 0 {
		return raw
	}
	return strings.ReplaceAll(strings.ReplaceAll(raw, "\r\n", "\n"), "\r", "\n")
}

func (p *estreeProjector) jsxValue(raw string, text bool) string {
	if !p.isJS {
		return decodeTypeScriptJSX(raw)
	}
	if text {
		return ecmascript.DecodeJSXText(raw)
	}
	return ecmascript.DecodeJSXAttribute(raw)
}

func decodeTypeScriptJSX(raw string) string {
	if !strings.ContainsRune(raw, '&') {
		return raw
	}
	var result strings.Builder
	result.Grow(len(raw))
	for {
		amp := strings.IndexByte(raw, '&')
		if amp < 0 {
			break
		}
		result.WriteString(raw[:amp])
		raw = raw[amp:]
		semi := strings.IndexByte(raw, ';')
		if semi < 0 {
			break
		}
		// Only the last ampersand before this semicolon can start a reference.
		// Consume its literal prefix once, including runs of invalid ampersands.
		if lastAmp := strings.LastIndexByte(raw[1:semi], '&'); lastAmp >= 0 {
			prefix := lastAmp + 1
			result.WriteString(raw[:prefix])
			raw = raw[prefix:]
			semi -= prefix
		}
		// TypeScript ESTree leaves invalid code points verbatim; tsgo's JSX
		// transformer replaces them. Decode only validated references.
		decoded, ok := jsxtx.DecodeEntity(raw[1:semi])
		if ok && decoded >= 0 && decoded <= unicode.MaxRune {
			result.WriteString(jsxtx.DecodeEntities(raw[:semi+1]))
			raw = raw[semi+1:]
		} else {
			result.WriteString(raw[:semi+1])
			raw = raw[semi+1:]
		}
	}
	result.WriteString(raw)
	return result.String()
}

func (p *estreeProjector) convertClass(node *ast.Node, declaration bool) uint32 {
	n := node.ClassLikeData()
	var superClass uint32
	superTypeArguments := undefinedValue()
	implementsBase := len(p.listScratch)
	if n.HeritageClauses != nil {
		for _, clauseNode := range n.HeritageClauses.Nodes {
			clause := clauseNode.AsHeritageClause()
			if clause.Types == nil {
				continue
			}
			for index, heritageNode := range clause.Types.Nodes {
				previous := p.origin
				if p.mapping != nil {
					p.origin = heritageNode
				}
				heritageExpression, heritageTypeArguments := heritageParts(heritageNode)
				if clause.Token == ast.KindExtendsKeyword && index == 0 {
					superClass = p.convertHeritageName(heritageExpression)
					superTypeArguments = p.typeParameters(heritageTypeArguments, false)
				} else if clause.Token == ast.KindImplementsKeyword {
					id := p.addNode(typeTSClassImplements, heritageNode,
						nodeValue(p.convertHeritageName(heritageExpression)),
						p.typeParameters(heritageTypeArguments, false))
					p.registerReverse(heritageNode, id)
					p.listScratch = append(p.listScratch, id)
				}
				p.origin = previous
			}
		}
	}
	implementsList := p.listValue(p.listScratch[implementsBase:])
	p.listScratch = p.listScratch[:implementsBase]

	membersBase := len(p.listScratch)
	if n.Members != nil {
		for _, member := range n.Members.Nodes {
			if member.Kind == ast.KindSemicolonClassElement {
				continue
			}
			if id := p.convert(member, false); id != 0 {
				p.listScratch = append(p.listScratch, id)
			}
		}
	}
	membersList := p.listValue(p.listScratch[membersBase:])
	p.listScratch = p.listScratch[:membersBase]

	end := p.offsets.At(node.End())
	start := p.offsets.At(p.bodyBrace(node, n.Members))
	body := p.add(typeClassBody, start, end, membersList)
	typeID := typeClassExpression
	if declaration {
		typeID = typeClassDeclaration
	}
	return p.addNode(typeID, node,
		boolValue(p.hasModifier(node, ast.ModifierFlagsAbstract)), nodeValue(body),
		boolValue(p.hasModifier(node, ast.ModifierFlagsAmbient)), p.decorators(node),
		nodeValue(p.convert(node.Name(), false)), implementsList, nodeValue(superClass),
		superTypeArguments, p.typeParameters(n.TypeParameters, true))
}

func (p *estreeProjector) convertMethod(node *ast.Node) uint32 {
	function := node.FunctionLikeData()
	bodyData := node.BodyData()
	parameters := []*ast.Node(nil)
	if function != nil && function.Parameters != nil {
		parameters = function.Parameters.Nodes
	}
	bodyNode := (*ast.Node)(nil)
	asterisk := false
	if bodyData != nil {
		bodyNode = bodyData.Body
		asterisk = bodyData.AsteriskToken != nil
	}
	functionType := typeFunctionExpression
	if bodyNode == nil {
		functionType = typeTSEmptyBodyFunctionExpression
	}
	value := p.addNode(functionType, node,
		nullValue(), p.convertParameters(parameters), nodeValue(p.convert(bodyNode, false)),
		boolValue(asterisk), boolValue(p.hasModifier(node, ast.ModifierFlagsAsync)),
		boolValue(false), boolValue(false), p.typeParameters(function.TypeParameters, true),
		p.typeAnnotation(function.Type))
	name := node.Name()
	if function.Parameters == nil {
		failProjection(errors.New("missing method parameter list"))
	}
	functionStart := max(0, function.Parameters.Pos()-1)
	if function.TypeParameters != nil && !p.isJS && (len(function.TypeParameters.Nodes) == 0 || !isJSDocCommentSyntax(function.TypeParameters.Nodes[0])) {
		functionStart = max(0, function.TypeParameters.Pos()-1)
	}
	p.nodes[value-1].start = p.offsets.At(functionStart)
	kind := "method"
	var key uint32
	switch node.Kind {
	case ast.KindGetAccessor:
		kind = "get"
	case ast.KindSetAccessor:
		kind = "set"
	case ast.KindConstructor:
		// tsgo also uses Constructor for a static method named constructor.
		if !p.hasModifier(node, ast.ModifierFlagsStatic) {
			kind = "constructor"
		}
		key = p.constructorKey(node)
	}
	if key == 0 {
		key = p.convert(name, false)
	}
	computed := name != nil && name.Kind == ast.KindComputedPropertyName
	optional := false
	if postfix := node.PostfixToken(); postfix != nil {
		optional = postfix.Kind == ast.KindQuestionToken
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
		propertyKind := kind
		if kind == "method" {
			propertyKind = "init"
		}
		return p.addNode(typeProperty, node,
			nodeValue(key), nodeValue(value), p.stringValue(propertyKind), boolValue(kind == "method"),
			boolValue(false), boolValue(computed), boolValue(optional))
	}
	methodType := typeMethodDefinition
	if p.hasModifier(node, ast.ModifierFlagsAbstract) {
		methodType = typeTSAbstractMethodDefinition
	}
	return p.addNode(methodType, node,
		p.accessibility(node), boolValue(computed), p.decorators(node), nodeValue(key),
		p.stringValue(kind), boolValue(optional), boolValue(p.hasModifier(node, ast.ModifierFlagsOverride)),
		boolValue(p.hasModifier(node, ast.ModifierFlagsStatic)), nodeValue(value))
}

func (p *estreeProjector) constructorKey(node *ast.Node) uint32 {
	// tsgo omits the constructor name node. The keyword or quoted name is
	// the first token after its parsed modifiers, regardless of trivia.
	start := node.Pos()
	if modifiers := node.Modifiers(); modifiers != nil && len(modifiers.Nodes) > 0 {
		start = modifiers.Nodes[len(modifiers.Nodes)-1].End()
	}
	scan := &p.scan
	scan.SetText(p.text)
	scan.SetSkipTrivia(true)
	scan.ResetTokenState(start)
	kind := scan.Scan()
	start, end := scan.TokenStart(), scan.TokenEnd()
	switch kind {
	case ast.KindConstructorKeyword:
		return p.add(typeIdentifier, p.offsets.At(start), p.offsets.At(end),
			p.listValue(nil), p.stringValue("constructor"), boolValue(false), undefinedValue())
	case ast.KindStringLiteral:
		if scan.TokenValue() == "constructor" {
			return p.add(typeLiteral, p.offsets.At(start), p.offsets.At(end),
				p.stringValue("constructor"), p.stringValue(p.text[start:end]),
				undefinedValue(), undefinedValue(), undefinedValue())
		}
	}
	failProjection(errors.New("missing constructor name token"))
	return 0
}

func (p *estreeProjector) convertPropertyDefinition(node *ast.Node) uint32 {
	n := node.AsPropertyDeclaration()
	postfix := n.PostfixToken
	optional := postfix != nil && postfix.Kind == ast.KindQuestionToken
	definite := postfix != nil && postfix.Kind == ast.KindExclamationToken
	propertyType := typePropertyDefinition
	if p.hasModifier(node, ast.ModifierFlagsAccessor) {
		propertyType = typeAccessorProperty
	}
	if p.hasModifier(node, ast.ModifierFlagsAbstract) {
		if propertyType == typeAccessorProperty {
			propertyType = typeTSAbstractAccessorProperty
		} else {
			propertyType = typeTSAbstractPropertyDefinition
		}
	}
	return p.addNode(propertyType, node,
		p.accessibility(node), boolValue(n.Name() != nil && n.Name().Kind == ast.KindComputedPropertyName),
		boolValue(p.hasModifier(node, ast.ModifierFlagsAmbient)), p.decorators(node), boolValue(definite),
		nodeValue(p.convert(n.Name(), false)), boolValue(optional),
		boolValue(p.hasModifier(node, ast.ModifierFlagsOverride)),
		boolValue(p.hasModifier(node, ast.ModifierFlagsReadonly)),
		boolValue(p.hasModifier(node, ast.ModifierFlagsStatic)), p.typeAnnotation(n.Type),
		nodeValue(p.convert(n.Initializer, false)))
}

func (p *estreeProjector) convertImportDeclaration(node *ast.Node) uint32 {
	n := node.AsImportDeclaration()
	base := len(p.listScratch)
	importKind := "value"
	phase := nullValue()
	if p.isJS {
		phase = undefinedValue()
	}
	if n.ImportClause != nil {
		clause := n.ImportClause.AsImportClause()
		if clause.PhaseModifier == ast.KindTypeKeyword {
			importKind = "type"
		} else if clause.PhaseModifier != ast.KindUnknown {
			phase = p.stringValue(scanner.TokenToString(clause.PhaseModifier))
		}
		if clause.Name() != nil {
			spec := p.convert(n.ImportClause, false)
			p.listScratch = append(p.listScratch, spec)
		}
		if clause.NamedBindings != nil {
			switch clause.NamedBindings.Kind {
			case ast.KindNamespaceImport:
				spec := p.convert(clause.NamedBindings, false)
				p.listScratch = append(p.listScratch, spec)
			case ast.KindNamedImports:
				for _, specifier := range clause.NamedBindings.AsNamedImports().Elements.Nodes {
					spec := p.convert(specifier, false)
					p.listScratch = append(p.listScratch, spec)
				}
			}
		}
	}
	specifiers := p.listValue(p.listScratch[base:])
	p.listScratch = p.listScratch[:base]
	return p.addNode(typeImportDeclaration, node,
		p.convertImportAttributes(n.Attributes), p.stringValue(importKind),
		nodeValue(p.convert(n.ModuleSpecifier, false)), specifiers, phase)
}

func (p *estreeProjector) convertImportAttributes(node *ast.Node) uint32 {
	if node == nil {
		return p.listValue(nil)
	}
	attributes := node.AsImportAttributes().Attributes
	if attributes == nil {
		return p.listValue(nil)
	}
	return p.convertList(attributes.Nodes, false)
}

func (p *estreeProjector) convertExportDeclaration(node *ast.Node) uint32 {
	n := node.AsExportDeclaration()
	exportKind := "value"
	if n.IsTypeOnly {
		exportKind = "type"
	}
	if n.ExportClause != nil && n.ExportClause.Kind == ast.KindNamedExports {
		return p.addNode(typeExportNamedDeclaration, node,
			p.convertImportAttributes(n.Attributes), nullValue(), p.stringValue(exportKind),
			nodeValue(p.convert(n.ModuleSpecifier, false)),
			p.convertList(n.ExportClause.AsNamedExports().Elements.Nodes, false))
	}
	var exported uint32
	if n.ExportClause != nil && n.ExportClause.Kind == ast.KindNamespaceExport {
		exported = p.convert(n.ExportClause.Name(), false)
	}
	return p.addNode(typeExportAllDeclaration, node,
		p.convertImportAttributes(n.Attributes), nodeValue(exported), p.stringValue(exportKind),
		nodeValue(p.convert(n.ModuleSpecifier, false)))
}

func (p *estreeProjector) convertInterface(node *ast.Node) uint32 {
	n := node.AsInterfaceDeclaration()
	base := len(p.listScratch)
	if n.HeritageClauses != nil {
		for _, clauseNode := range n.HeritageClauses.Nodes {
			clause := clauseNode.AsHeritageClause()
			if clause.Token != ast.KindExtendsKeyword || clause.Types == nil {
				continue
			}
			for _, heritageNode := range clause.Types.Nodes {
				previous := p.origin
				if p.mapping != nil {
					p.origin = heritageNode
				}
				heritageExpression, heritageTypeArguments := heritageParts(heritageNode)
				id := p.addNode(typeTSInterfaceHeritage, heritageNode,
					nodeValue(p.convertHeritageName(heritageExpression)),
					p.typeParameters(heritageTypeArguments, false))
				p.origin = previous
				p.registerReverse(heritageNode, id)
				p.listScratch = append(p.listScratch, id)
			}
		}
	}
	extendsList := p.listValue(p.listScratch[base:])
	p.listScratch = p.listScratch[:base]
	end := p.offsets.At(node.End())
	start := p.offsets.At(p.bodyBrace(node, n.Members))
	body := p.add(typeTSInterfaceBody, start, end, p.convertList(n.Members.Nodes, false))
	return p.addNode(typeTSInterfaceDeclaration, node,
		nodeValue(p.convert(n.Name(), false)), p.typeParameters(n.TypeParameters, true),
		extendsList, nodeValue(body), boolValue(p.hasModifier(node, ast.ModifierFlagsAmbient)))
}

func (p *estreeProjector) convertPropertySignature(node *ast.Node) uint32 {
	n := node.AsPropertySignatureDeclaration()
	optional := n.PostfixToken != nil && n.PostfixToken.Kind == ast.KindQuestionToken
	return p.addNode(typeTSPropertySignature, node,
		p.accessibility(node), boolValue(n.Name() != nil && n.Name().Kind == ast.KindComputedPropertyName),
		nodeValue(p.convert(n.Name(), false)), boolValue(optional),
		boolValue(p.hasModifier(node, ast.ModifierFlagsReadonly)),
		boolValue(p.hasModifier(node, ast.ModifierFlagsStatic)), p.typeAnnotation(n.Type))
}

func (p *estreeProjector) convertMethodSignature(node *ast.Node) uint32 {
	n := node.FunctionLikeData()
	kind := "method"
	switch node.Kind {
	case ast.KindGetAccessor:
		kind = "get"
	case ast.KindSetAccessor:
		kind = "set"
	}
	optional := node.PostfixToken() != nil && node.PostfixToken().Kind == ast.KindQuestionToken
	return p.addNode(typeTSMethodSignature, node,
		p.accessibility(node), boolValue(node.Name() != nil && node.Name().Kind == ast.KindComputedPropertyName),
		nodeValue(p.convert(node.Name(), false)), p.stringValue(kind), boolValue(optional),
		p.convertParameters(n.Parameters.Nodes), boolValue(p.hasModifier(node, ast.ModifierFlagsReadonly)),
		p.typeAnnotation(n.Type), boolValue(p.hasModifier(node, ast.ModifierFlagsStatic)),
		p.typeParameters(n.TypeParameters, true))
}

func (p *estreeProjector) convertFunctionType(node *ast.Node, constructor bool) uint32 {
	n := node.FunctionLikeData()
	parameters := []*ast.Node(nil)
	if n.Parameters != nil {
		parameters = n.Parameters.Nodes
	}
	if constructor {
		return p.addNode(typeTSConstructorType, node,
			boolValue(p.hasModifier(node, ast.ModifierFlagsAbstract)), p.convertParameters(parameters),
			p.typeAnnotation(n.Type), p.typeParameters(n.TypeParameters, true))
	}
	return p.addNode(typeTSFunctionType, node,
		p.convertParameters(parameters), p.typeAnnotation(n.Type), p.typeParameters(n.TypeParameters, true))
}

func (p *estreeProjector) convertMappedType(node *ast.Node) uint32 {
	n := node.AsMappedTypeNode()
	typeParameter := n.TypeParameter.AsTypeParameterDeclaration()
	optional := boolValue(false)
	if n.QuestionToken != nil {
		if n.QuestionToken.Kind == ast.KindQuestionToken {
			optional = boolValue(true)
		} else {
			optional = p.stringValue(scanner.TokenToString(n.QuestionToken.Kind))
		}
	}
	readonly := undefinedValue()
	if n.ReadonlyToken != nil {
		if n.ReadonlyToken.Kind == ast.KindReadonlyKeyword {
			readonly = boolValue(true)
		} else {
			readonly = p.stringValue(scanner.TokenToString(n.ReadonlyToken.Kind))
		}
	}
	return p.addNode(typeTSMappedType, node,
		nodeValue(p.convert(typeParameter.Constraint, false)),
		nodeValue(p.convert(typeParameter.Name(), false)),
		nodeValue(p.convert(n.NameType, false)), optional, readonly,
		optionalNodeValue(p.convert(n.Type, false)))
}

func (p *estreeProjector) convertTemplateLiteralType(node *ast.Node) uint32 {
	n := node.AsTemplateLiteralTypeNode()
	var quasiBuf [8]uint32
	var typeBuf [8]uint32
	quasis := quasiBuf[:0]
	types := typeBuf[:0]
	quasis = append(quasis, p.convert(n.Head, false))
	if n.TemplateSpans != nil {
		for _, spanNode := range n.TemplateSpans.Nodes {
			span := spanNode.AsTemplateLiteralTypeSpan()
			types = append(types, p.convert(span.Type, false))
			quasis = append(quasis, p.convert(span.Literal, false))
		}
	}
	return p.addNode(typeTSTemplateLiteralType, node, p.listValue(quasis), p.listValue(types))
}

func (p *estreeProjector) convertImportType(node *ast.Node) uint32 {
	n := node.AsImportTypeNode()
	source := uint32(0)
	if n.Argument != nil && n.Argument.Kind == ast.KindLiteralType {
		source = p.convert(n.Argument.AsLiteralTypeNode().Literal, false)
	} else {
		failProjection(errors.New("invalid import type source"))
	}
	options := uint32(0)
	if n.Attributes != nil {
		attributes := n.Attributes.AsImportAttributes()
		propsBase := len(p.listScratch)
		if attributes.Attributes != nil {
			for _, attributeNode := range attributes.Attributes.Nodes {
				attribute := attributeNode.AsImportAttribute()
				name := attribute.Name()
				prop := p.addNode(typeProperty, attributeNode,
					nodeValue(p.convert(name, false)), nodeValue(p.convert(attribute.Value, false)),
					p.stringValue("init"), boolValue(false), boolValue(false),
					boolValue(false), boolValue(false))
				p.listScratch = append(p.listScratch, prop)
			}
		}
		propsList := p.listValue(p.listScratch[propsBase:])
		p.listScratch = p.listScratch[:propsBase]
		inner := p.addNode(typeObjectExpression, n.Attributes, propsList)
		keyword := "with"
		if attributes.Token == ast.KindAssertKeyword {
			keyword = "assert"
		}
		// tsgo does not give the outer options object or with/assert key their
		// own nodes. Read their authored ranges with the existing scanner.
		scan := &p.scan
		scan.SetText(p.text)
		scan.SetSkipTrivia(true)
		scan.ResetTokenState(n.Argument.End())
		if scan.Scan() != ast.KindCommaToken || scan.Scan() != ast.KindOpenBraceToken {
			failProjection(errors.New("missing import type options"))
		}
		outerStart := p.offsets.At(scan.TokenStart())
		scan.Scan()
		keyStart, keyEnd := p.offsets.At(scan.TokenStart()), p.offsets.At(scan.TokenEnd())
		end := p.offsets.At(n.Attributes.End())
		key := p.add(typeIdentifier, keyStart, keyEnd,
			p.listValue(nil), p.stringValue(keyword), boolValue(false), undefinedValue())
		outerProperty := p.add(typeProperty, keyStart, end,
			nodeValue(key), nodeValue(inner), p.stringValue("init"),
			boolValue(false), boolValue(false), boolValue(false), boolValue(false))
		scan.ResetTokenState(n.Attributes.End())
		if scan.Scan() == ast.KindCommaToken {
			scan.Scan()
		}
		if scan.Token() != ast.KindCloseBraceToken {
			failProjection(errors.New("missing import type options closing brace"))
		}
		options = p.add(typeObjectExpression, outerStart, p.offsets.At(scan.TokenEnd()), p.listValue([]uint32{outerProperty}))
	}
	typeArguments := p.typeParameters(n.TypeArguments, false)
	if typeArguments == undefinedValue() {
		typeArguments = nullValue()
	}
	result := p.addNode(typeTSImportType, node,
		nodeValue(options), nodeValue(p.convert(n.Qualifier, false)),
		nodeValue(source), typeArguments)
	if n.IsTypeOf {
		start := p.tokenStart(node)
		p.nodes[result-1].start = p.offsets.At(scanner.SkipTrivia(p.text, start+len("typeof")))
		return p.addNode(typeTSTypeQuery, node, nodeValue(result), undefinedValue())
	}
	return result
}

func (p *estreeProjector) convertEnum(node *ast.Node) uint32 {
	n := node.AsEnumDeclaration()
	members := []*ast.Node(nil)
	if n.Members != nil {
		members = n.Members.Nodes
	}
	memberList := p.convertList(members, false)
	start, end := p.rangeOf(node)
	if n.Members != nil {
		bodyStart := p.bodyBrace(node, n.Members)
		start = p.offsets.At(bodyStart)
	}
	body := p.add(typeTSEnumBody, start, end, memberList)
	return p.addNode(typeTSEnumDeclaration, node,
		nodeValue(body), boolValue(p.hasModifier(node, ast.ModifierFlagsConst)),
		boolValue(p.hasModifier(node, ast.ModifierFlagsAmbient)),
		nodeValue(p.convert(n.Name(), false)))
}

func (p *estreeProjector) convertModuleDeclaration(node *ast.Node) uint32 {
	n := node.AsModuleDeclaration()
	kind := "module"
	global := n.Keyword == ast.KindGlobalKeyword
	if n.Keyword == ast.KindNamespaceKeyword {
		kind = "namespace"
	} else if global {
		kind = "global"
	}
	// Upstream constructs namespace identifiers directly, without registering
	// reverse aliases through child conversion. String module names use both maps.
	var name uint32
	if !global && n.Name() != nil && n.Name().Kind == ast.KindIdentifier {
		name = p.convertNode(n.Name(), false)
	} else {
		name = p.convert(n.Name(), false)
	}
	inner := n
	declare := p.hasModifier(node, ast.ModifierFlagsAmbient)
	// Dotted namespaces have one ESTree declaration and a qualified id.
	for inner.Body != nil && inner.Body.Kind == ast.KindModuleDeclaration {
		inner = inner.Body.AsModuleDeclaration()
		right := p.convertNode(inner.Name(), false)
		if name != 0 && right != 0 {
			name = p.addWithOrigin(typeTSQualifiedName, inner.Name(), p.nodes[name-1].start, p.nodes[right-1].end, nodeValue(name), nodeValue(right))
		}
		declare = declare || p.hasModifier(inner.AsNode(), ast.ModifierFlagsAmbient)
	}
	declaration := p.addNode(typeTSModuleDeclaration, node, optionalNodeValue(p.convert(inner.Body, false)), boolValue(declare), boolValue(global), nodeValue(name), p.stringValue(kind))
	if p.hasModifier(node, ast.ModifierFlagsExport) {
		// typescript-estree fixes exports on the last segment after building the declaration from the first.
		p.registerReverse(inner.AsNode(), declaration)
		return p.wrapExport(node, inner.AsNode(), declaration)
	}
	return declaration
}

func (p *estreeProjector) convertJsxName(node *ast.Node) uint32 {
	previous := p.origin
	if p.mapping != nil {
		p.origin = node
	}
	id := p.convertJsxNameNode(node)
	p.origin = previous
	p.registerReverse(node, id)
	return id
}

func (p *estreeProjector) convertJsxNameNode(node *ast.Node) uint32 {
	if node == nil {
		return 0
	}
	p.recordTokenCover(node, true)
	switch node.Kind {
	case ast.KindIdentifier:
		return p.addNode(typeJSXIdentifier, node, p.stringValue(node.Text()))
	case ast.KindThisKeyword:
		return p.addNode(typeJSXIdentifier, node, p.stringValue("this"))
	case ast.KindPropertyAccessExpression:
		n := node.AsPropertyAccessExpression()
		return p.addNode(typeJSXMemberExpression, node,
			nodeValue(p.convertJsxName(n.Expression)), nodeValue(p.convertJsxName(n.Name())))
	case ast.KindJsxNamespacedName:
		n := node.AsJsxNamespacedName()
		return p.addNode(typeJSXNamespacedName, node,
			nodeValue(p.convertJsxName(n.Namespace)), nodeValue(p.convertJsxName(n.Name())))
	default:
		return p.convert(node, false)
	}
}

func (p *estreeProjector) convertJsxSelfClosing(node *ast.Node) uint32 {
	n := node.AsJsxSelfClosingElement()
	opening := p.addNode(typeJSXOpeningElement, node,
		nodeValue(p.convertJsxName(n.TagName)), p.convertList(n.Attributes.AsJsxAttributes().Properties.Nodes, false),
		boolValue(true), p.typeParameters(n.TypeArguments, false))
	return p.addNode(typeJSXElement, node, nodeValue(opening), p.listValue(nil), nullValue())
}

func (p *estreeProjector) convertJsxExpression(node *ast.Node) uint32 {
	n := node.AsJsxExpression()
	expression := p.convert(n.Expression, false)
	if expression == 0 {
		start, end := p.rangeOf(node)
		emptyStart := min(start+1, end)
		emptyEnd := emptyStart
		if end > emptyStart {
			emptyEnd = end - 1
		}
		expression = p.add(typeJSXEmptyExpression, emptyStart, emptyEnd)
	}
	if n.DotDotDotToken != nil {
		return p.addNode(typeJSXSpreadChild, node, nodeValue(expression))
	}
	return p.addNode(typeJSXExpressionContainer, node, nodeValue(expression))
}

func (p *estreeProjector) convertChainOperand(node *ast.Node) (uint32, bool) {
	id := p.convert(node, false)
	if id == 0 || node.Kind == ast.KindParenthesizedExpression {
		return id, false
	}
	// A terminal non-null assertion may lack tsgo's optional-chain flag.
	// Preserve the converted operand's chain until the outer expression wraps
	// it; parentheses remain an explicit boundary for short-circuiting.
	record := p.nodes[id-1]
	if TypeID(record.typeID) != typeChainExpression {
		return id, false
	}
	expression := p.values[record.valueStart] & payloadMask
	if m := p.mapping; m != nil {
		origin := m.Origins[id]
		for _, index := range m.chainReverse {
			m.Reverse[index] = detachedChainReverse
			m.ChainAliases = append(m.ChainAliases, ChainAlias{index, origin, expression, record.start, record.end})
		}
		m.chainRecord = 0
		m.chainReverse = m.chainReverse[:0]
		m.Origins = m.Origins[:id]
	}
	// The wrapper is the last record emitted by convert and has no users yet.
	// Its parent consumes the inner expression, so reclaim it before proceeding.
	p.nodes = p.nodes[:id-1]
	p.values = p.values[:record.valueStart]
	if !p.treeMode {
		p.frameBytes -= 14 + 4 // Node columns and the single expression field.
	}
	return expression, true
}

func (p *estreeProjector) templateCooked(node *ast.Node) uint32 {
	if node.TemplateLiteralLikeData().TemplateFlags&ast.TokenFlagsContainsInvalidEscape != 0 {
		return nullValue()
	}
	return p.stringValue(node.Text())
}

// NodeList.Pos starts immediately after the opening brace, before its trivia.
func (p *estreeProjector) bodyBrace(node *ast.Node, list *ast.NodeList) int {
	if list != nil {
		pos := list.Pos() - 1
		if pos >= 0 && pos < len(p.text) && p.text[pos] == '{' {
			return pos
		}
	}
	failProjection(fmt.Errorf("missing body brace for %s", node.Kind))
	return 0
}

func heritageParts(node *ast.Node) (*ast.Node, *ast.NodeList) {
	switch node.Kind {
	case ast.KindTypeReference:
		data := node.AsTypeReferenceNode()
		return data.TypeName, data.TypeArguments
	case ast.KindExpressionWithTypeArguments:
		data := node.AsExpressionWithTypeArguments()
		return data.Expression, data.TypeArguments
	}
	failProjection(fmt.Errorf("unsupported heritage syntax %s", node.Kind))
	return nil, nil
}

func (p *estreeProjector) convertTypeQueryName(node *ast.Node) uint32 {
	id := p.convertTypeQueryNameNode(node)
	p.registerReverse(node, id)
	return id
}

func (p *estreeProjector) convertTypeQueryNameNode(node *ast.Node) uint32 {
	if node == nil {
		return 0
	}
	if node.Kind == ast.KindIdentifier && node.Text() == "this" {
		p.check()
		// Type queries encode this as an identifier in tsgo, including its
		// token classification, but ESTree uses an expression in this role.
		p.recordTokenCover(node, false)
		return p.addNode(typeThisExpression, node)
	}
	if node.Kind == ast.KindQualifiedName {
		p.check()
		data := node.AsQualifiedName()
		return p.addNode(typeTSQualifiedName, node,
			nodeValue(p.convertTypeQueryName(data.Left)), nodeValue(p.convert(data.Right, false)))
	}
	return p.convert(node, false)
}

func (p *estreeProjector) convertHeritageName(node *ast.Node) uint32 {
	id := p.convertHeritageNameNode(node)
	p.registerReverse(node, id)
	return id
}

func (p *estreeProjector) convertHeritageNameNode(node *ast.Node) uint32 {
	if node == nil {
		return 0
	}
	if node.Kind == ast.KindQualifiedName {
		data := node.AsQualifiedName()
		return p.addNode(typeMemberExpression, node, nodeValue(p.convertHeritageName(data.Left)), nodeValue(p.convert(data.Right, false)), boolValue(false), boolValue(false))
	}
	return p.convert(node, false)
}
