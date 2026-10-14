package estree

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

func (p *estreeProjector) convert(node *ast.Node, pattern bool) uint32 {
	if node == nil {
		return 0
	}
	p.check()
	if expression := utils.JSDocTypeCastExpression(node); expression != nil {
		return p.convert(expression, pattern)
	}
	if utils.IsJSDocSyntaxNode(node) {
		return 0
	}
	p.recordTokenCover(node)
	switch node.Kind {
	case ast.KindSourceFile:
		file := node.AsSourceFile()
		body := make([]uint32, 0, len(file.Statements.Nodes))
		for _, statement := range file.Statements.Nodes {
			if id := p.convertStatement(statement); id != 0 {
				body = append(body, id)
			}
		}
		// Preserve the rule-facing span: TS begins at its first syntax token,
		// JS at zero, and both end at EOF. Disable directives use this position.
		start := uint32(0)
		end := p.offsets.At(node.End())
		if !p.isJavaScriptFile() {
			start = p.offsets.At(scanner.SkipTrivia(p.text, 0))
		}
		return p.add("Program", start, end, p.listValue(body), p.stringValue(p.sourceType))

	case ast.KindBlock:
		return p.addNode("BlockStatement", node, p.convertStatements(node.AsBlock().Statements))
	case ast.KindModuleBlock:
		return p.addNode("TSModuleBlock", node, p.convertStatements(node.AsModuleBlock().Statements))
	case ast.KindEmptyStatement:
		return p.addNode("EmptyStatement", node)
	case ast.KindDebuggerStatement:
		return p.addNode("DebuggerStatement", node)
	case ast.KindExpressionStatement:
		n := node.AsExpressionStatement()
		directive := undefinedValue()
		if n.Expression != nil && n.Expression.Kind == ast.KindStringLiteral && p.isDirective(node) {
			raw := p.rawText(n.Expression)
			directive = p.stringValue(raw[1 : len(raw)-1])
		}
		return p.addNode("ExpressionStatement", node, nodeValue(p.convert(n.Expression, false)), directive)
	case ast.KindReturnStatement:
		return p.addNode("ReturnStatement", node, nodeValue(p.convert(node.AsReturnStatement().Expression, false)))
	case ast.KindThrowStatement:
		return p.addNode("ThrowStatement", node, nodeValue(p.convert(node.AsThrowStatement().Expression, false)))
	case ast.KindIfStatement:
		n := node.AsIfStatement()
		return p.addNode("IfStatement", node,
			nodeValue(p.convert(n.Expression, false)),
			nodeValue(p.convertStatement(n.ThenStatement)),
			nodeValue(p.convertStatement(n.ElseStatement)))
	case ast.KindWhileStatement:
		n := node.AsWhileStatement()
		return p.addNode("WhileStatement", node,
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convertStatement(n.Statement)))
	case ast.KindDoStatement:
		n := node.AsDoStatement()
		return p.addNode("DoWhileStatement", node,
			nodeValue(p.convertStatement(n.Statement)), nodeValue(p.convert(n.Expression, false)))
	case ast.KindForStatement:
		n := node.AsForStatement()
		return p.addNode("ForStatement", node,
			nodeValue(p.convert(n.Initializer, false)),
			nodeValue(p.convert(n.Condition, false)),
			nodeValue(p.convert(n.Incrementor, false)),
			nodeValue(p.convertStatement(n.Statement)))
	case ast.KindForInStatement, ast.KindForOfStatement:
		n := node.AsForInOrOfStatement()
		left := p.convert(n.Initializer, true)
		if node.Kind == ast.KindForInStatement {
			return p.addNode("ForInStatement", node,
				nodeValue(left), nodeValue(p.convert(n.Expression, false)), nodeValue(p.convertStatement(n.Statement)))
		}
		return p.addNode("ForOfStatement", node,
			boolValue(n.AwaitModifier != nil), nodeValue(left),
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convertStatement(n.Statement)))
	case ast.KindBreakStatement:
		return p.addNode("BreakStatement", node, nodeValue(p.convert(node.AsBreakStatement().Label, false)))
	case ast.KindContinueStatement:
		return p.addNode("ContinueStatement", node, nodeValue(p.convert(node.AsContinueStatement().Label, false)))
	case ast.KindWithStatement:
		n := node.AsWithStatement()
		return p.addNode("WithStatement", node,
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convertStatement(n.Statement)))
	case ast.KindSwitchStatement:
		n := node.AsSwitchStatement()
		var clauses []*ast.Node
		if n.CaseBlock != nil && n.CaseBlock.AsCaseBlock().Clauses != nil {
			clauses = n.CaseBlock.AsCaseBlock().Clauses.Nodes
		}
		return p.addNode("SwitchStatement", node,
			nodeValue(p.convert(n.Expression, false)), p.convertList(clauses, false))
	case ast.KindCaseClause, ast.KindDefaultClause:
		n := node.AsCaseOrDefaultClause()
		test := uint32(0)
		if node.Kind == ast.KindCaseClause {
			test = p.convert(n.Expression, false)
		}
		return p.addNode("SwitchCase", node, nodeValue(test), p.convertStatements(n.Statements))
	case ast.KindTryStatement:
		n := node.AsTryStatement()
		return p.addNode("TryStatement", node,
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
		return p.addNode("CatchClause", node, nodeValue(parameter), nodeValue(p.convert(n.Block, false)))
	case ast.KindLabeledStatement:
		n := node.AsLabeledStatement()
		return p.addNode("LabeledStatement", node,
			nodeValue(p.convert(n.Label, false)), nodeValue(p.convertStatement(n.Statement)))

	case ast.KindVariableStatement:
		n := node.AsVariableStatement()
		return p.convertVariableDeclarationList(n.DeclarationList, node)
	case ast.KindVariableDeclarationList:
		return p.convertVariableDeclarationList(node, node)
	case ast.KindVariableDeclaration:
		n := node.AsVariableDeclaration()
		return p.addNode("VariableDeclarator", node,
			nodeValue(p.identifierWithAnnotation(n.Name(), n.Type)),
			nodeValue(p.convert(n.Initializer, false)),
			boolValue(n.ExclamationToken != nil))
	case ast.KindFunctionDeclaration:
		return p.convertFunction(node, "FunctionDeclaration")
	case ast.KindFunctionExpression:
		return p.convertFunction(node, "FunctionExpression")
	case ast.KindArrowFunction:
		return p.convertArrowFunction(node)
	case ast.KindParameter:
		return p.convertParameter(node)

	case ast.KindIdentifier:
		return p.addNode("Identifier", node,
			p.listValue(nil), p.stringValue(node.Text()), boolValue(false), undefinedValue())
	case ast.KindComputedPropertyName:
		return p.convert(node.AsComputedPropertyName().Expression, false)
	case ast.KindPrivateIdentifier:
		return p.addNode("PrivateIdentifier", node, p.stringValue(strings.TrimPrefix(node.Text(), "#")))
	case ast.KindThisKeyword:
		return p.addNode("ThisExpression", node)
	case ast.KindSuperKeyword:
		return p.addNode("Super", node)

	case ast.KindStringLiteral:
		if node.Parent != nil && node.Parent.Kind == ast.KindJsxAttribute {
			raw := p.rawText(node)
			return p.addLiteral(node, p.stringValue(ecmascript.DecodeJSXAttribute(raw[1:len(raw)-1])), undefinedValue(), undefinedValue(), undefinedValue())
		}
		return p.addLiteral(node, p.stringValue(node.Text()), undefinedValue(), undefinedValue(), undefinedValue())
	case ast.KindNumericLiteral:
		return p.addLiteral(node, p.numberValue(parseNumberLiteral(node.Text())), undefinedValue(), undefinedValue(), undefinedValue())
	case ast.KindBigIntLiteral:
		raw := p.rawText(node)
		bigint := strings.TrimSuffix(strings.ReplaceAll(raw, "_", ""), "n")
		integer, ok := new(big.Int).SetString(bigint, 0)
		if !ok {
			panic(fmt.Errorf("invalid bigint literal %q", bigint))
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
		return p.addNode("ConditionalExpression", node,
			nodeValue(p.convert(n.Condition, false)),
			nodeValue(p.convert(n.WhenTrue, false)),
			nodeValue(p.convert(n.WhenFalse, false)))
	case ast.KindPrefixUnaryExpression:
		n := node.AsPrefixUnaryExpression()
		operator := scanner.TokenToString(n.Operator)
		typeName := "UnaryExpression"
		if n.Operator == ast.KindPlusPlusToken || n.Operator == ast.KindMinusMinusToken {
			typeName = "UpdateExpression"
		}
		return p.addNode(typeName, node,
			p.stringValue(operator), boolValue(true), nodeValue(p.convert(n.Operand, false)))
	case ast.KindPostfixUnaryExpression:
		n := node.AsPostfixUnaryExpression()
		return p.addNode("UpdateExpression", node,
			p.stringValue(scanner.TokenToString(n.Operator)), boolValue(false), nodeValue(p.convert(n.Operand, false)))
	case ast.KindDeleteExpression, ast.KindVoidExpression, ast.KindTypeOfExpression:
		operator := map[ast.Kind]string{
			ast.KindDeleteExpression: "delete",
			ast.KindVoidExpression:   "void",
			ast.KindTypeOfExpression: "typeof",
		}[node.Kind]
		return p.addNode("UnaryExpression", node,
			p.stringValue(operator), boolValue(true), nodeValue(p.convert(node.Expression(), false)))
	case ast.KindAwaitExpression:
		return p.addNode("AwaitExpression", node, nodeValue(p.convert(node.AsAwaitExpression().Expression, false)))
	case ast.KindYieldExpression:
		n := node.AsYieldExpression()
		return p.addNode("YieldExpression", node,
			nodeValue(p.convert(n.Expression, false)), boolValue(n.AsteriskToken != nil))
	case ast.KindAsExpression:
		n := node.AsAsExpression()
		return p.addNode("TSAsExpression", node,
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convert(n.Type, false)))
	case ast.KindSatisfiesExpression:
		n := node.AsSatisfiesExpression()
		return p.addNode("TSSatisfiesExpression", node,
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convert(n.Type, false)))
	case ast.KindNonNullExpression:
		id := p.addNode("TSNonNullExpression", node, nodeValue(p.convertChainOperand(node.AsNonNullExpression().Expression)))
		return p.wrapOptionalChain(node, id)
	case ast.KindTypeAssertionExpression:
		n := node.AsTypeAssertion()
		return p.addNode("TSTypeAssertion", node,
			nodeValue(p.convert(n.Expression, false)), nodeValue(p.convert(n.Type, false)))
	case ast.KindMetaProperty:
		n := node.AsMetaProperty()
		start, _ := p.rangeOf(node)
		keyword := scanner.TokenToString(n.KeywordToken)
		meta := p.add("Identifier", start, start+uint32(len(keyword)),
			p.listValue(nil), p.stringValue(keyword), boolValue(false), undefinedValue())
		return p.addNode("MetaProperty", node,
			nodeValue(meta), nodeValue(p.convert(n.Name(), false)))

	case ast.KindPropertyAccessExpression:
		n := node.AsPropertyAccessExpression()
		object := p.convertChainOperand(n.Expression)
		id := p.addNode("MemberExpression", node,
			nodeValue(object), nodeValue(p.convert(n.Name(), false)),
			boolValue(false), boolValue(n.QuestionDotToken != nil))
		return p.wrapOptionalChain(node, id)
	case ast.KindElementAccessExpression:
		n := node.AsElementAccessExpression()
		object := p.convertChainOperand(n.Expression)
		id := p.addNode("MemberExpression", node,
			nodeValue(object), nodeValue(p.convert(n.ArgumentExpression, false)),
			boolValue(true), boolValue(n.QuestionDotToken != nil))
		return p.wrapOptionalChain(node, id)
	case ast.KindCallExpression:
		return p.convertCall(node)
	case ast.KindNewExpression:
		n := node.AsNewExpression()
		arguments := []*ast.Node(nil)
		if n.Arguments != nil {
			arguments = n.Arguments.Nodes
		}
		callee, typeArguments := n.Expression, n.TypeArguments
		inner := callee
		for inner != nil && inner.Kind == ast.KindParenthesizedExpression {
			inner = inner.AsParenthesizedExpression().Expression
		}
		if inner != nil && inner.Kind == ast.KindExpressionWithTypeArguments {
			instantiation := inner.AsExpressionWithTypeArguments()
			callee, typeArguments = instantiation.Expression, instantiation.TypeArguments
		}
		return p.addNode("NewExpression", node,
			nodeValue(p.convert(callee, false)), p.convertList(arguments, false),
			p.typeParameters(typeArguments, false))
	case ast.KindArrayLiteralExpression:
		elements := p.convertListWithNulls(node.AsArrayLiteralExpression().Elements.Nodes, pattern)
		if pattern {
			return p.addNode("ArrayPattern", node, elements, p.listValue(nil), boolValue(false), undefinedValue())
		}
		return p.addNode("ArrayExpression", node, elements)
	case ast.KindObjectLiteralExpression:
		properties := p.convertList(node.AsObjectLiteralExpression().Properties.Nodes, pattern)
		if pattern {
			return p.addNode("ObjectPattern", node, properties, p.listValue(nil), boolValue(false), undefinedValue())
		}
		return p.addNode("ObjectExpression", node, properties)
	case ast.KindPropertyAssignment:
		n := node.AsPropertyAssignment()
		name := n.Name()
		return p.addNode("Property", node,
			nodeValue(p.convert(name, false)), nodeValue(p.convert(n.Initializer, pattern)),
			p.stringValue("init"), boolValue(false), boolValue(false),
			boolValue(name != nil && name.Kind == ast.KindComputedPropertyName), boolValue(false))
	case ast.KindShorthandPropertyAssignment:
		n := node.AsShorthandPropertyAssignment()
		name := n.Name()
		value := p.convert(name, false)
		if n.ObjectAssignmentInitializer != nil {
			value = p.addNode("AssignmentPattern", node,
				p.listValue(nil), nodeValue(value), boolValue(false),
				nodeValue(p.convert(n.ObjectAssignmentInitializer, false)), undefinedValue())
		}
		return p.addNode("Property", node,
			nodeValue(p.convert(name, false)), nodeValue(value), p.stringValue("init"),
			boolValue(false), boolValue(true), boolValue(false), boolValue(false))
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		if pattern {
			return p.addNode("RestElement", node,
				p.listValue(nil), nodeValue(p.convert(node.Expression(), true)),
				boolValue(false), undefinedValue())
		}
		return p.addNode("SpreadElement", node, nodeValue(p.convert(node.Expression(), false)))
	case ast.KindObjectBindingPattern:
		return p.addNode("ObjectPattern", node,
			p.convertList(node.AsBindingPattern().Elements.Nodes, true),
			p.listValue(nil), boolValue(false), undefinedValue())
	case ast.KindArrayBindingPattern:
		return p.addNode("ArrayPattern", node,
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
		return p.addNode("TaggedTemplateExpression", node,
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
		return p.addNode("StaticBlock", node, p.convertStatements(body.AsBlock().Statements))

	case ast.KindDecorator:
		return p.addNode("Decorator", node, nodeValue(p.convert(node.AsDecorator().Expression, false)))
	case ast.KindNamespaceExportDeclaration:
		return p.addNode("TSNamespaceExportDeclaration", node, nodeValue(p.convert(node.Name(), false)))
	case ast.KindImportDeclaration:
		return p.convertImportDeclaration(node)
	case ast.KindNamespaceImport:
		return p.addNode("ImportNamespaceSpecifier", node,
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
		return p.addNode("ImportSpecifier", node,
			nodeValue(p.convert(imported, false)), p.stringValue(kind), nodeValue(p.convert(n.Name(), false)))
	case ast.KindImportClause:
		name := node.AsImportClause().Name()
		return p.addNode("ImportDefaultSpecifier", name, nodeValue(p.convert(name, false)))
	case ast.KindImportAttribute:
		n := node.AsImportAttribute()
		return p.addNode("ImportAttribute", node,
			nodeValue(p.convert(n.Name(), false)), nodeValue(p.convert(n.Value, false)))
	case ast.KindImportEqualsDeclaration:
		n := node.AsImportEqualsDeclaration()
		importKind := "value"
		if n.IsTypeOnly {
			importKind = "type"
		}
		return p.addNode("TSImportEqualsDeclaration", node,
			nodeValue(p.convert(n.Name(), false)), p.stringValue(importKind),
			nodeValue(p.convert(n.ModuleReference, false)))
	case ast.KindExternalModuleReference:
		return p.addNode("TSExternalModuleReference", node,
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
		return p.addNode("ExportSpecifier", node,
			nodeValue(p.convert(n.Name(), false)), p.stringValue(kind), nodeValue(p.convert(local, false)))
	case ast.KindExportAssignment:
		n := node.AsExportAssignment()
		if n.IsExportEquals {
			return p.addNode("TSExportAssignment", node, nodeValue(p.convert(n.Expression, false)))
		}
		return p.addNode("ExportDefaultDeclaration", node,
			nodeValue(p.convert(n.Expression, false)), p.stringValue("value"))

	case ast.KindInterfaceDeclaration:
		return p.convertInterface(node)
	case ast.KindTypeAliasDeclaration, ast.KindJSTypeAliasDeclaration:
		n := node.AsTypeAliasDeclaration()
		return p.addNode("TSTypeAliasDeclaration", node,
			boolValue(p.hasModifier(node, ast.ModifierFlagsAmbient)),
			nodeValue(p.convert(n.Name(), false)), nodeValue(p.convert(n.Type, false)),
			p.typeParameters(n.TypeParameters, true))
	case ast.KindPropertySignature:
		return p.convertPropertySignature(node)
	case ast.KindMethodSignature:
		return p.convertMethodSignature(node)
	case ast.KindIndexSignature:
		n := node.AsIndexSignatureDeclaration()
		return p.addNode("TSIndexSignature", node,
			p.accessibility(node), p.convertList(n.Parameters.Nodes, false),
			boolValue(p.hasModifier(node, ast.ModifierFlagsReadonly)),
			boolValue(p.hasModifier(node, ast.ModifierFlagsStatic)), p.typeAnnotation(n.Type))
	case ast.KindCallSignature:
		n := node.AsCallSignatureDeclaration()
		parameters := []*ast.Node(nil)
		if n.Parameters != nil {
			parameters = n.Parameters.Nodes
		}
		return p.addNode("TSCallSignatureDeclaration", node,
			p.convertParameters(parameters), p.typeAnnotation(n.Type),
			p.typeParameters(n.TypeParameters, true))
	case ast.KindConstructSignature:
		n := node.AsConstructSignatureDeclaration()
		parameters := []*ast.Node(nil)
		if n.Parameters != nil {
			parameters = n.Parameters.Nodes
		}
		return p.addNode("TSConstructSignatureDeclaration", node,
			p.convertParameters(parameters), p.typeAnnotation(n.Type),
			p.typeParameters(n.TypeParameters, true))
	case ast.KindTypeReference:
		n := node.AsTypeReferenceNode()
		return p.addNode("TSTypeReference", node,
			nodeValue(p.convert(n.TypeName, false)), p.typeParameters(n.TypeArguments, false))
	case ast.KindQualifiedName:
		n := node.AsQualifiedName()
		return p.addNode("TSQualifiedName", node,
			nodeValue(p.convert(n.Left, false)), nodeValue(p.convert(n.Right, false)))
	case ast.KindTypeLiteral:
		return p.addNode("TSTypeLiteral", node, p.convertList(node.AsTypeLiteralNode().Members.Nodes, false))
	case ast.KindUnionType:
		return p.addNode("TSUnionType", node, p.convertList(node.AsUnionTypeNode().Types.Nodes, false))
	case ast.KindIntersectionType:
		return p.addNode("TSIntersectionType", node, p.convertList(node.AsIntersectionTypeNode().Types.Nodes, false))
	case ast.KindConditionalType:
		n := node.AsConditionalTypeNode()
		return p.addNode("TSConditionalType", node,
			nodeValue(p.convert(n.CheckType, false)), nodeValue(p.convert(n.ExtendsType, false)),
			nodeValue(p.convert(n.FalseType, false)), nodeValue(p.convert(n.TrueType, false)))
	case ast.KindMappedType:
		return p.convertMappedType(node)
	case ast.KindInferType:
		return p.addNode("TSInferType", node,
			nodeValue(p.convert(node.AsInferTypeNode().TypeParameter, false)))
	case ast.KindThisType:
		return p.addNode("TSThisType", node)
	case ast.KindArrayType:
		return p.addNode("TSArrayType", node, nodeValue(p.convert(node.AsArrayTypeNode().ElementType, false)))
	case ast.KindLiteralType:
		if node.AsLiteralTypeNode().Literal.Kind == ast.KindNullKeyword {
			return p.addNode("TSNullKeyword", node)
		}
		return p.addNode("TSLiteralType", node, nodeValue(p.convert(node.AsLiteralTypeNode().Literal, false)))
	case ast.KindTypeParameter:
		n := node.AsTypeParameterDeclaration()
		return p.addNode("TSTypeParameter", node,
			boolValue(p.hasModifier(node, ast.ModifierFlagsConst)),
			nodeValue(p.convert(n.Constraint, false)), nodeValue(p.convert(n.DefaultType, false)),
			boolValue(p.hasModifier(node, ast.ModifierFlagsIn)), nodeValue(p.convert(n.Name(), false)),
			boolValue(p.hasModifier(node, ast.ModifierFlagsOut)))
	case ast.KindStringKeyword:
		return p.addNode("TSStringKeyword", node)
	case ast.KindNumberKeyword:
		return p.addNode("TSNumberKeyword", node)
	case ast.KindBooleanKeyword:
		return p.addNode("TSBooleanKeyword", node)
	case ast.KindAnyKeyword:
		return p.addNode("TSAnyKeyword", node)
	case ast.KindUnknownKeyword:
		return p.addNode("TSUnknownKeyword", node)
	case ast.KindVoidKeyword:
		return p.addNode("TSVoidKeyword", node)
	case ast.KindUndefinedKeyword:
		return p.addNode("TSUndefinedKeyword", node)
	case ast.KindNeverKeyword:
		return p.addNode("TSNeverKeyword", node)
	case ast.KindObjectKeyword:
		return p.addNode("TSObjectKeyword", node)
	case ast.KindSymbolKeyword:
		return p.addNode("TSSymbolKeyword", node)
	case ast.KindBigIntKeyword:
		return p.addNode("TSBigIntKeyword", node)
	case ast.KindIntrinsicKeyword:
		return p.addNode("TSIntrinsicKeyword", node)
	case ast.KindFunctionType:
		return p.convertFunctionType(node, false)
	case ast.KindConstructorType:
		return p.convertFunctionType(node, true)
	case ast.KindTupleType:
		return p.addNode("TSTupleType", node, p.convertList(node.AsTupleTypeNode().Elements.Nodes, false))
	case ast.KindNamedTupleMember:
		n := node.AsNamedTupleMember()
		id := p.addNode("TSNamedTupleMember", node,
			nodeValue(p.convert(n.Type, false)), nodeValue(p.convert(n.Name(), false)), boolValue(n.QuestionToken != nil))
		if n.DotDotDotToken != nil {
			p.nodes[id-1].start = p.offsets.At(scanner.GetTokenPosOfNode(n.Name(), p.file, false))
			return p.addNode("TSRestType", node, nodeValue(id))
		}
		return id
	case ast.KindOptionalType:
		return p.addNode("TSOptionalType", node, nodeValue(p.convert(node.AsOptionalTypeNode().Type, false)))
	case ast.KindRestType:
		return p.addNode("TSRestType", node, nodeValue(p.convert(node.AsRestTypeNode().Type, false)))
	case ast.KindParenthesizedType:
		return p.convert(node.AsParenthesizedTypeNode().Type, false)
	case ast.KindIndexedAccessType:
		n := node.AsIndexedAccessTypeNode()
		return p.addNode("TSIndexedAccessType", node,
			nodeValue(p.convert(n.ObjectType, false)), nodeValue(p.convert(n.IndexType, false)))
	case ast.KindTypeQuery:
		n := node.AsTypeQueryNode()
		return p.addNode("TSTypeQuery", node,
			nodeValue(p.convert(n.ExprName, false)), p.typeParameters(n.TypeArguments, false))
	case ast.KindTypeOperator:
		n := node.AsTypeOperatorNode()
		return p.addNode("TSTypeOperator", node,
			p.stringValue(scanner.TokenToString(n.Operator)), nodeValue(p.convert(n.Type, false)))
	case ast.KindTypePredicate:
		n := node.AsTypePredicateNode()
		annotation := nullValue()
		if n.Type != nil {
			annotation = p.typeAnnotation(n.Type)
		}
		return p.addNode("TSTypePredicate", node,
			boolValue(n.AssertsModifier != nil), nodeValue(p.convert(n.ParameterName, false)), annotation)
	case ast.KindImportType:
		return p.convertImportType(node)
	case ast.KindTemplateLiteralType:
		return p.convertTemplateLiteralType(node)
	case ast.KindTemplateLiteralTypeSpan, ast.KindOmittedExpression:
		return 0
	case ast.KindExpressionWithTypeArguments:
		n := node.AsExpressionWithTypeArguments()
		return p.addNode("TSInstantiationExpression", node,
			nodeValue(p.convert(n.Expression, false)), p.typeParameters(n.TypeArguments, false))

	case ast.KindEnumDeclaration:
		return p.convertEnum(node)
	case ast.KindEnumMember:
		n := node.AsEnumMember()
		return p.addNode("TSEnumMember", node,
			nodeValue(p.convert(n.Name(), false)), nodeValue(p.convert(n.Initializer, false)))
	case ast.KindModuleDeclaration:
		return p.convertModuleDeclaration(node)

	case ast.KindJsxElement:
		n := node.AsJsxElement()
		return p.addNode("JSXElement", node,
			nodeValue(p.convert(n.OpeningElement, false)), p.convertList(n.Children.Nodes, false),
			nodeValue(p.convert(n.ClosingElement, false)))
	case ast.KindJsxSelfClosingElement:
		return p.convertJsxSelfClosing(node)
	case ast.KindJsxOpeningElement:
		n := node.AsJsxOpeningElement()
		return p.addNode("JSXOpeningElement", node,
			nodeValue(p.convertJsxName(n.TagName)), p.convertList(n.Attributes.AsJsxAttributes().Properties.Nodes, false),
			boolValue(false), p.typeParameters(n.TypeArguments, false))
	case ast.KindJsxClosingElement:
		return p.addNode("JSXClosingElement", node,
			nodeValue(p.convertJsxName(node.AsJsxClosingElement().TagName)))
	case ast.KindJsxFragment:
		n := node.AsJsxFragment()
		return p.addNode("JSXFragment", node,
			nodeValue(p.convert(n.OpeningFragment, false)), p.convertList(n.Children.Nodes, false),
			nodeValue(p.convert(n.ClosingFragment, false)))
	case ast.KindJsxOpeningFragment:
		return p.addNode("JSXOpeningFragment", node)
	case ast.KindJsxClosingFragment:
		return p.addNode("JSXClosingFragment", node)
	case ast.KindJsxAttribute:
		n := node.AsJsxAttribute()
		return p.addNode("JSXAttribute", node,
			nodeValue(p.convertJsxName(n.Name())), nodeValue(p.convert(n.Initializer, false)))
	case ast.KindJsxSpreadAttribute:
		return p.addNode("JSXSpreadAttribute", node,
			nodeValue(p.convert(node.AsJsxSpreadAttribute().Expression, false)))
	case ast.KindJsxExpression:
		return p.convertJsxExpression(node)
	case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces:
		raw := p.rawTextFull(node)
		return p.addNode("JSXText", node, p.stringValue(ecmascript.DecodeJSXText(raw)), p.stringValue(raw))
	case ast.KindJsxNamespacedName:
		n := node.AsJsxNamespacedName()
		return p.addNode("JSXNamespacedName", node,
			nodeValue(p.convertJsxName(n.Namespace)), nodeValue(p.convertJsxName(n.Name())))
	}

	if ast.IsJSDocNodeKind(node.Kind) {
		return 0
	}
	return p.unsupported(node)
}

func (p *estreeProjector) convertStatements(list *ast.NodeList) uint32 {
	if list == nil {
		return p.listValue(nil)
	}
	ids := make([]uint32, 0, len(list.Nodes))
	for _, statement := range list.Nodes {
		if id := p.convertStatement(statement); id != 0 {
			ids = append(ids, id)
		}
	}
	return p.listValue(ids)
}

func (p *estreeProjector) convertStatement(node *ast.Node) uint32 {
	if node == nil {
		return 0
	}
	id := p.convert(node, false)
	if id == 0 || !p.hasModifier(node, ast.ModifierFlagsExport) {
		return id
	}
	return p.wrapExport(node, id)
}

func (p *estreeProjector) wrapExport(node *ast.Node, declaration uint32) uint32 {
	// The wrapper owns export/default; the declaration begins at its own syntax.
	start := scanner.GetTokenPosOfNode(node, p.file, false)
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
	if node.Kind == ast.KindInterfaceDeclaration || node.Kind == ast.KindTypeAliasDeclaration {
		kind = "type"
	}
	var result uint32
	if p.hasModifier(node, ast.ModifierFlagsDefault) {
		result = p.addNode("ExportDefaultDeclaration", node, nodeValue(declaration), p.stringValue("value"))
	} else {
		result = p.addNode("ExportNamedDeclaration", node, p.listValue(nil), nodeValue(declaration), p.stringValue(kind), nullValue(), p.listValue(nil))
	}
	for _, modifier := range node.Modifiers().Nodes {
		if modifier.Kind == ast.KindExportKeyword {
			p.nodes[result-1].start = p.offsets.At(scanner.GetTokenPosOfNode(modifier, p.file, false))
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
	return p.addNode("VariableDeclaration", rangeNode,
		p.convertList(n.Declarations.Nodes, false),
		boolValue(node.Parent != nil && p.hasModifier(node.Parent, ast.ModifierFlagsAmbient)),
		p.stringValue(kind))
}

func (p *estreeProjector) addLiteral(node *ast.Node, value, bigint, regexPattern, regexFlags uint32) uint32 {
	return p.addNode("Literal", node,
		value, p.stringValue(p.rawText(node)), bigint, regexPattern, regexFlags)
}

func (p *estreeProjector) rawText(node *ast.Node) string {
	start := scanner.GetTokenPosOfNode(node, p.file, false)
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
	panic(fmt.Errorf("ESTree projection: unsupported syntax %s at byte %d", node.Kind, node.Pos()))
}

func (p *estreeProjector) unwrapChain(id uint32) uint32 {
	if id == 0 || int(id) > len(p.nodes) {
		return id
	}
	record := p.nodes[id-1]
	if schemas[record.typeID].name != "ChainExpression" {
		return id
	}
	value := p.values[record.valueStart]
	if value>>tagShift == tagNode {
		return value & payloadMask
	}
	return id
}

func (p *estreeProjector) wrapOptionalChain(node *ast.Node, id uint32) uint32 {
	if node.Flags&ast.NodeFlagsOptionalChain == 0 {
		return id
	}
	return p.addNode("ChainExpression", node, nodeValue(id))
}

func (p *estreeProjector) convertFunction(node *ast.Node, requestedType string) uint32 {
	function := node.FunctionLikeData()
	body := node.BodyData()
	typeName := requestedType
	if requestedType == "FunctionDeclaration" && (body == nil || body.Body == nil) {
		typeName = "TSDeclareFunction"
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
	values := []uint32{
		nodeValue(p.convert(node.Name(), false)),
		p.convertParameters(parameters),
		nodeValue(p.convert(bodyNode, false)),
		boolValue(asterisk),
		boolValue(p.hasModifier(node, ast.ModifierFlagsAsync)),
		boolValue(false),
		boolValue(p.hasModifier(node, ast.ModifierFlagsAmbient)),
		p.typeParameters(function.TypeParameters, true),
		p.typeAnnotation(function.Type),
	}
	return p.addNode(typeName, node, values...)
}

func (p *estreeProjector) convertArrowFunction(node *ast.Node) uint32 {
	n := node.AsArrowFunction()
	parameters := []*ast.Node(nil)
	if n.Parameters != nil {
		parameters = n.Parameters.Nodes
	}
	return p.addNode("ArrowFunctionExpression", node,
		nullValue(), p.convertParameters(parameters), nodeValue(p.convert(n.Body, false)),
		boolValue(false), boolValue(p.hasModifier(node, ast.ModifierFlagsAsync)),
		boolValue(n.Body != nil && n.Body.Kind != ast.KindBlock),
		p.typeParameters(n.TypeParameters, true), p.typeAnnotation(n.Type))
}

func (p *estreeProjector) convertParameters(parameters []*ast.Node) uint32 {
	ids := make([]uint32, 0, len(parameters))
	for _, parameter := range parameters {
		if id := p.convertParameter(parameter); id != 0 {
			ids = append(ids, id)
		}
	}
	return p.listValue(ids)
}

func (p *estreeProjector) convertParameter(node *ast.Node) uint32 {
	if node == nil || utils.IsJSDocSyntaxNode(node) {
		return 0
	}
	n := node.AsParameterDeclaration()
	var parameter uint32
	if n.DotDotDotToken == nil {
		parameter = p.identifierWithAnnotation(n.Name(), n.Type)
	} else {
		parameter = p.convert(n.Name(), true)
	}
	if n.QuestionToken != nil {
		parameter = p.cloneBindingOptional(parameter, true)
		p.nodes[parameter-1].end = max(p.nodes[parameter-1].end, p.offsets.At(n.QuestionToken.End()))
	}
	result := parameter
	if n.DotDotDotToken != nil {
		result = p.addNode("RestElement", node,
			p.listValue(nil), nodeValue(parameter), boolValue(n.QuestionToken != nil),
			p.typeAnnotation(n.Type))
	} else if n.Initializer != nil {
		result = p.addNode("AssignmentPattern", node,
			p.listValue(nil), nodeValue(parameter), boolValue(n.QuestionToken != nil),
			nodeValue(p.convert(n.Initializer, false)), undefinedValue())
	}
	// Accessibility and readonly belong to TSParameterProperty, not its binding.
	p.nodes[result-1].start = p.nodes[parameter-1].start
	if n.DotDotDotToken != nil {
		p.nodes[result-1].start = p.offsets.At(scanner.GetTokenPosOfNode(n.DotDotDotToken, p.file, false))
	}
	if utils.ESTreeModifierFlags(node)&ast.ModifierFlagsParameterPropertyModifier != 0 {
		return p.addNode("TSParameterProperty", node,
			p.accessibility(node), p.decorators(node),
			boolValue(p.hasModifier(node, ast.ModifierFlagsOverride)), nodeValue(result),
			boolValue(p.hasModifier(node, ast.ModifierFlagsReadonly)),
			boolValue(p.hasModifier(node, ast.ModifierFlagsStatic)))
	}
	for index, field := range schemas[p.nodes[result-1].typeID].fields {
		if field == "decorators" {
			decorators := p.decorators(node)
			p.values[int(p.nodes[result-1].valueStart)+index] = decorators
			break
		}
	}
	return result
}

func (p *estreeProjector) cloneBindingOptional(id uint32, optional bool) uint32 {
	if id == 0 || int(id) > len(p.nodes) {
		return id
	}
	record := p.nodes[id-1]
	if schemas[record.typeID].name != "Identifier" {
		return id
	}
	start := int(record.valueStart)
	values := append([]uint32(nil), p.values[start:start+4]...)
	values[2] = boolValue(optional)
	return p.add("Identifier", record.start, record.end, values...)
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
		return p.addNode("RestElement", node,
			p.listValue(nil), nodeValue(value), boolValue(false), undefinedValue())
	}
	if n.Initializer != nil {
		start := p.nodes[value-1].start
		value = p.addNode("AssignmentPattern", node,
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
	return p.addNode("Property", node,
		nodeValue(p.convert(key, false)), nodeValue(value), p.stringValue("init"),
		boolValue(false), boolValue(n.PropertyName == nil),
		boolValue(n.PropertyName != nil && n.PropertyName.Kind == ast.KindComputedPropertyName), boolValue(false))
}

func (p *estreeProjector) convertBinary(node *ast.Node, pattern bool) uint32 {
	n := node.AsBinaryExpression()
	p.recordTokenCover(n.OperatorToken)
	operator := scanner.TokenToString(n.OperatorToken.Kind)
	if n.OperatorToken.Kind == ast.KindCommaToken {
		left := p.convert(n.Left, false)
		items := p.sequenceItems(left)
		items = append(items, p.convert(n.Right, false))
		return p.addNode("SequenceExpression", node, p.listValue(items))
	}
	typeName := "BinaryExpression"
	if n.OperatorToken.Kind >= ast.KindFirstAssignment && n.OperatorToken.Kind <= ast.KindLastAssignment {
		typeName = "AssignmentExpression"
	} else if n.OperatorToken.Kind == ast.KindAmpersandAmpersandToken ||
		n.OperatorToken.Kind == ast.KindBarBarToken ||
		n.OperatorToken.Kind == ast.KindQuestionQuestionToken {
		typeName = "LogicalExpression"
	}
	if pattern && typeName == "AssignmentExpression" {
		return p.addNode("AssignmentPattern", node,
			p.listValue(nil), nodeValue(p.convert(n.Left, true)), boolValue(false),
			nodeValue(p.convert(n.Right, false)), undefinedValue())
	}
	return p.addNode(typeName, node,
		nodeValue(p.convert(n.Left, typeName == "AssignmentExpression")),
		p.stringValue(operator), nodeValue(p.convert(n.Right, false)))
}

func (p *estreeProjector) sequenceItems(id uint32) []uint32 {
	if id == 0 || int(id) > len(p.nodes) {
		return []uint32{id}
	}
	record := p.nodes[id-1]
	if schemas[record.typeID].name != "SequenceExpression" {
		return []uint32{id}
	}
	encoded := p.values[record.valueStart]
	if encoded>>tagShift != tagList {
		return []uint32{id}
	}
	offset := encoded & payloadMask
	length := p.lists[offset]
	return append([]uint32(nil), p.lists[offset+1:offset+1+length]...)
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
		return p.addNode("ImportExpression", node, nodeValue(source), nodeValue(options))
	}
	callee := p.convertChainOperand(n.Expression)
	id := p.addNode("CallExpression", node,
		nodeValue(callee), p.convertList(arguments, false), boolValue(n.QuestionDotToken != nil),
		p.typeParameters(n.TypeArguments, false))
	return p.wrapOptionalChain(node, id)
}

func (p *estreeProjector) convertNoSubstitutionTemplate(node *ast.Node) uint32 {
	start, end := p.rangeOf(node)
	raw := p.rawText(node)
	if len(raw) >= 2 {
		raw = raw[1 : len(raw)-1]
	}
	element := p.add("TemplateElement", start, end,
		boolValue(true), p.templateCooked(node), p.stringValue(raw))
	return p.addNode("TemplateLiteral", node, p.listValue(nil), p.listValue([]uint32{element}))
}

func (p *estreeProjector) convertTemplateExpression(node *ast.Node) uint32 {
	n := node.AsTemplateExpression()
	expressions := make([]uint32, 0, len(n.TemplateSpans.Nodes))
	quasis := make([]uint32, 0, len(n.TemplateSpans.Nodes)+1)
	quasis = append(quasis, p.convert(n.Head, false))
	for _, spanNode := range n.TemplateSpans.Nodes {
		span := spanNode.AsTemplateSpan()
		expressions = append(expressions, p.convert(span.Expression, false))
		quasis = append(quasis, p.convert(span.Literal, false))
	}
	return p.addNode("TemplateLiteral", node, p.listValue(expressions), p.listValue(quasis))
}

func (p *estreeProjector) convertTemplateElement(node *ast.Node) uint32 {
	tail := node.Kind == ast.KindTemplateTail
	raw := node.RawText()
	return p.addNode("TemplateElement", node,
		boolValue(tail), p.templateCooked(node), p.stringValue(raw))
}

func (p *estreeProjector) convertClass(node *ast.Node, declaration bool) uint32 {
	n := node.ClassLikeData()
	var superClass uint32
	superTypeArguments := undefinedValue()
	implements := make([]uint32, 0)
	if n.HeritageClauses != nil {
		for _, clauseNode := range n.HeritageClauses.Nodes {
			clause := clauseNode.AsHeritageClause()
			if clause.Types == nil {
				continue
			}
			for index, heritageNode := range clause.Types.Nodes {
				heritageExpression, heritageTypeArguments := heritageParts(heritageNode)
				if clause.Token == ast.KindExtendsKeyword && index == 0 {
					superClass = p.convertHeritageName(heritageExpression)
					superTypeArguments = p.typeParameters(heritageTypeArguments, false)
				} else if clause.Token == ast.KindImplementsKeyword {
					implements = append(implements, p.addNode("TSClassImplements", heritageNode,
						nodeValue(p.convertHeritageName(heritageExpression)),
						p.typeParameters(heritageTypeArguments, false)))
				}
			}
		}
	}
	members := []*ast.Node(nil)
	if n.Members != nil {
		members = n.Members.Nodes
	}
	memberIDs := make([]uint32, 0, len(members))
	for _, member := range members {
		if member.Kind == ast.KindSemicolonClassElement {
			continue
		}
		if id := p.convert(member, false); id != 0 {
			memberIDs = append(memberIDs, id)
		}
	}
	_, end := p.rangeOf(node)
	start := p.offsets.At(p.bodyBrace(node, n.Members))
	body := p.add("ClassBody", start, end, p.listValue(memberIDs))
	typeName := "ClassExpression"
	if declaration {
		typeName = "ClassDeclaration"
	}
	return p.addNode(typeName, node,
		boolValue(p.hasModifier(node, ast.ModifierFlagsAbstract)), nodeValue(body),
		boolValue(p.hasModifier(node, ast.ModifierFlagsAmbient)), p.decorators(node),
		nodeValue(p.convert(node.Name(), false)), p.listValue(implements), nodeValue(superClass),
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
	functionType := "FunctionExpression"
	if bodyNode == nil {
		functionType = "TSEmptyBodyFunctionExpression"
	}
	value := p.addNode(functionType, node,
		nullValue(), p.convertParameters(parameters), nodeValue(p.convert(bodyNode, false)),
		boolValue(asterisk), boolValue(p.hasModifier(node, ast.ModifierFlagsAsync)),
		boolValue(false), boolValue(false), p.typeParameters(function.TypeParameters, true),
		p.typeAnnotation(function.Type))
	name := node.Name()
	if function.Parameters == nil {
		panic(errors.New("missing method parameter list"))
	}
	functionStart := function.Parameters.Pos() - 1
	if function.TypeParameters != nil {
		functionStart = function.TypeParameters.Pos() - 1
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
		return p.addNode("Property", node,
			nodeValue(key), nodeValue(value), p.stringValue(propertyKind), boolValue(kind == "method"),
			boolValue(false), boolValue(computed), boolValue(optional))
	}
	methodType := "MethodDefinition"
	if p.hasModifier(node, ast.ModifierFlagsAbstract) {
		methodType = "TSAbstractMethodDefinition"
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
	scan := scanner.NewScanner()
	scan.SetText(p.text)
	scan.SetSkipTrivia(true)
	scan.ResetTokenState(start)
	kind := scan.Scan()
	start, end := scan.TokenStart(), scan.TokenEnd()
	switch kind {
	case ast.KindConstructorKeyword:
		return p.add("Identifier", p.offsets.At(start), p.offsets.At(end),
			p.listValue(nil), p.stringValue("constructor"), boolValue(false), undefinedValue())
	case ast.KindStringLiteral:
		if scan.TokenValue() == "constructor" {
			return p.add("Literal", p.offsets.At(start), p.offsets.At(end),
				p.stringValue("constructor"), p.stringValue(p.text[start:end]),
				undefinedValue(), undefinedValue(), undefinedValue())
		}
	}
	panic(errors.New("missing constructor name token"))
}

func (p *estreeProjector) convertPropertyDefinition(node *ast.Node) uint32 {
	n := node.AsPropertyDeclaration()
	postfix := n.PostfixToken
	optional := postfix != nil && postfix.Kind == ast.KindQuestionToken
	definite := postfix != nil && postfix.Kind == ast.KindExclamationToken
	propertyType := "PropertyDefinition"
	if p.hasModifier(node, ast.ModifierFlagsAccessor) {
		propertyType = "AccessorProperty"
	}
	if p.hasModifier(node, ast.ModifierFlagsAbstract) {
		if propertyType == "AccessorProperty" {
			propertyType = "TSAbstractAccessorProperty"
		} else {
			propertyType = "TSAbstractPropertyDefinition"
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
	specifiers := make([]uint32, 0, 4)
	importKind := "value"
	phase := nullValue()
	if n.ImportClause != nil {
		clause := n.ImportClause.AsImportClause()
		if clause.PhaseModifier == ast.KindTypeKeyword {
			importKind = "type"
		} else if clause.PhaseModifier != ast.KindUnknown {
			phase = p.stringValue(scanner.TokenToString(clause.PhaseModifier))
		}
		if clause.Name() != nil {
			specifiers = append(specifiers, p.convert(n.ImportClause, false))
		}
		if clause.NamedBindings != nil {
			switch clause.NamedBindings.Kind {
			case ast.KindNamespaceImport:
				specifiers = append(specifiers, p.convert(clause.NamedBindings, false))
			case ast.KindNamedImports:
				for _, specifier := range clause.NamedBindings.AsNamedImports().Elements.Nodes {
					specifiers = append(specifiers, p.convert(specifier, false))
				}
			}
		}
	}
	return p.addNode("ImportDeclaration", node,
		p.convertImportAttributes(n.Attributes), p.stringValue(importKind),
		nodeValue(p.convert(n.ModuleSpecifier, false)), p.listValue(specifiers), phase)
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
		return p.addNode("ExportNamedDeclaration", node,
			p.convertImportAttributes(n.Attributes), nullValue(), p.stringValue(exportKind),
			nodeValue(p.convert(n.ModuleSpecifier, false)),
			p.convertList(n.ExportClause.AsNamedExports().Elements.Nodes, false))
	}
	var exported uint32
	if n.ExportClause != nil && n.ExportClause.Kind == ast.KindNamespaceExport {
		exported = p.convert(n.ExportClause.Name(), false)
	}
	return p.addNode("ExportAllDeclaration", node,
		p.convertImportAttributes(n.Attributes), nodeValue(exported), p.stringValue(exportKind),
		nodeValue(p.convert(n.ModuleSpecifier, false)))
}

func (p *estreeProjector) convertInterface(node *ast.Node) uint32 {
	n := node.AsInterfaceDeclaration()
	extends := make([]uint32, 0)
	if n.HeritageClauses != nil {
		for _, clauseNode := range n.HeritageClauses.Nodes {
			clause := clauseNode.AsHeritageClause()
			if clause.Token != ast.KindExtendsKeyword || clause.Types == nil {
				continue
			}
			for _, heritageNode := range clause.Types.Nodes {
				heritageExpression, heritageTypeArguments := heritageParts(heritageNode)
				extends = append(extends, p.addNode("TSInterfaceHeritage", heritageNode,
					nodeValue(p.convertHeritageName(heritageExpression)),
					p.typeParameters(heritageTypeArguments, false)))
			}
		}
	}
	_, end := p.rangeOf(node)
	start := p.offsets.At(p.bodyBrace(node, n.Members))
	body := p.add("TSInterfaceBody", start, end, p.convertList(n.Members.Nodes, false))
	return p.addNode("TSInterfaceDeclaration", node,
		nodeValue(p.convert(n.Name(), false)), p.typeParameters(n.TypeParameters, true),
		p.listValue(extends), nodeValue(body), boolValue(p.hasModifier(node, ast.ModifierFlagsAmbient)))
}

func (p *estreeProjector) convertPropertySignature(node *ast.Node) uint32 {
	n := node.AsPropertySignatureDeclaration()
	optional := n.PostfixToken != nil && n.PostfixToken.Kind == ast.KindQuestionToken
	return p.addNode("TSPropertySignature", node,
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
	return p.addNode("TSMethodSignature", node,
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
		return p.addNode("TSConstructorType", node,
			boolValue(p.hasModifier(node, ast.ModifierFlagsAbstract)), p.convertParameters(parameters),
			p.typeAnnotation(n.Type), p.typeParameters(n.TypeParameters, true))
	}
	return p.addNode("TSFunctionType", node,
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
	return p.addNode("TSMappedType", node,
		nodeValue(p.convert(typeParameter.Constraint, false)),
		nodeValue(p.convert(typeParameter.Name(), false)),
		nodeValue(p.convert(n.NameType, false)), optional, readonly,
		nodeValue(p.convert(n.Type, false)))
}

func (p *estreeProjector) convertTemplateLiteralType(node *ast.Node) uint32 {
	n := node.AsTemplateLiteralTypeNode()
	spanCount := 0
	if n.TemplateSpans != nil {
		spanCount = len(n.TemplateSpans.Nodes)
	}
	quasis := make([]uint32, 0, spanCount+1)
	types := make([]uint32, 0, spanCount)
	quasis = append(quasis, p.convert(n.Head, false))
	if n.TemplateSpans != nil {
		for _, spanNode := range n.TemplateSpans.Nodes {
			span := spanNode.AsTemplateLiteralTypeSpan()
			types = append(types, p.convert(span.Type, false))
			quasis = append(quasis, p.convert(span.Literal, false))
		}
	}
	return p.addNode("TSTemplateLiteralType", node, p.listValue(quasis), p.listValue(types))
}

func (p *estreeProjector) convertImportType(node *ast.Node) uint32 {
	n := node.AsImportTypeNode()
	source := uint32(0)
	if n.Argument != nil && n.Argument.Kind == ast.KindLiteralType {
		source = p.convert(n.Argument.AsLiteralTypeNode().Literal, false)
	} else {
		panic(errors.New("invalid import type source"))
	}
	options := uint32(0)
	if n.Attributes != nil {
		attributes := n.Attributes.AsImportAttributes()
		properties := make([]uint32, 0)
		if attributes.Attributes != nil {
			properties = make([]uint32, 0, len(attributes.Attributes.Nodes))
			for _, attributeNode := range attributes.Attributes.Nodes {
				attribute := attributeNode.AsImportAttribute()
				name := attribute.Name()
				properties = append(properties, p.addNode("Property", attributeNode,
					nodeValue(p.convert(name, false)), nodeValue(p.convert(attribute.Value, false)),
					p.stringValue("init"), boolValue(false), boolValue(false),
					boolValue(false), boolValue(false)))
			}
		}
		inner := p.addNode("ObjectExpression", n.Attributes, p.listValue(properties))
		keyword := "with"
		if attributes.Token == ast.KindAssertKeyword {
			keyword = "assert"
		}
		// tsgo does not give the outer options object or with/assert key their
		// own nodes. Read their authored ranges with the existing scanner.
		scan := scanner.NewScanner()
		scan.SetText(p.text)
		scan.SetSkipTrivia(true)
		scan.ResetTokenState(n.Argument.End())
		if scan.Scan() != ast.KindCommaToken || scan.Scan() != ast.KindOpenBraceToken {
			panic(errors.New("missing import type options"))
		}
		outerStart := p.offsets.At(scan.TokenStart())
		scan.Scan()
		keyStart, keyEnd := p.offsets.At(scan.TokenStart()), p.offsets.At(scan.TokenEnd())
		_, end := p.rangeOf(n.Attributes)
		key := p.add("Identifier", keyStart, keyEnd,
			p.listValue(nil), p.stringValue(keyword), boolValue(false), undefinedValue())
		outerProperty := p.add("Property", keyStart, end,
			nodeValue(key), nodeValue(inner), p.stringValue("init"),
			boolValue(false), boolValue(false), boolValue(false), boolValue(false))
		scan.ResetTokenState(n.Attributes.End())
		if scan.Scan() == ast.KindCommaToken {
			scan.Scan()
		}
		if scan.Token() != ast.KindCloseBraceToken {
			panic(errors.New("missing import type options closing brace"))
		}
		options = p.add("ObjectExpression", outerStart, p.offsets.At(scan.TokenEnd()), p.listValue([]uint32{outerProperty}))
	}
	result := p.addNode("TSImportType", node,
		nodeValue(options), nodeValue(p.convert(n.Qualifier, false)),
		nodeValue(source), p.typeParameters(n.TypeArguments, false))
	if n.IsTypeOf {
		start := scanner.GetTokenPosOfNode(node, p.file, false)
		p.nodes[result-1].start = p.offsets.At(scanner.SkipTrivia(p.text, start+len("typeof")))
		return p.addNode("TSTypeQuery", node, nodeValue(result), undefinedValue())
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
	body := p.add("TSEnumBody", start, end, memberList)
	return p.addNode("TSEnumDeclaration", node,
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
	name := p.convert(n.Name(), false)
	body := n.Body
	declare := p.hasModifier(node, ast.ModifierFlagsAmbient)
	// Dotted namespaces have one ESTree declaration and a qualified id.
	for body != nil && body.Kind == ast.KindModuleDeclaration {
		nested := body.AsModuleDeclaration()
		right := p.convert(nested.Name(), false)
		name = p.add("TSQualifiedName", p.nodes[name-1].start, p.nodes[right-1].end, nodeValue(name), nodeValue(right))
		declare = declare || p.hasModifier(body, ast.ModifierFlagsAmbient)
		body = nested.Body
	}
	return p.addNode("TSModuleDeclaration", node, nodeValue(p.convert(body, false)), boolValue(declare), boolValue(global), nodeValue(name), p.stringValue(kind))
}

func (p *estreeProjector) convertJsxName(node *ast.Node) uint32 {
	if node == nil {
		return 0
	}
	p.recordTokenCover(node)
	switch node.Kind {
	case ast.KindIdentifier:
		return p.addNode("JSXIdentifier", node, p.stringValue(node.Text()))
	case ast.KindThisKeyword:
		return p.addNode("JSXIdentifier", node, p.stringValue("this"))
	case ast.KindPropertyAccessExpression:
		n := node.AsPropertyAccessExpression()
		return p.addNode("JSXMemberExpression", node,
			nodeValue(p.convertJsxName(n.Expression)), nodeValue(p.convertJsxName(n.Name())))
	case ast.KindJsxNamespacedName:
		n := node.AsJsxNamespacedName()
		return p.addNode("JSXNamespacedName", node,
			nodeValue(p.convertJsxName(n.Namespace)), nodeValue(p.convertJsxName(n.Name())))
	default:
		return p.convert(node, false)
	}
}

func (p *estreeProjector) convertJsxSelfClosing(node *ast.Node) uint32 {
	n := node.AsJsxSelfClosingElement()
	opening := p.addNode("JSXOpeningElement", node,
		nodeValue(p.convertJsxName(n.TagName)), p.convertList(n.Attributes.AsJsxAttributes().Properties.Nodes, false),
		boolValue(true), p.typeParameters(n.TypeArguments, false))
	return p.addNode("JSXElement", node, nodeValue(opening), p.listValue(nil), nullValue())
}

func (p *estreeProjector) convertJsxExpression(node *ast.Node) uint32 {
	n := node.AsJsxExpression()
	expression := p.convert(n.Expression, false)
	if expression == 0 {
		start, end := p.rangeOf(node)
		expression = p.add("JSXEmptyExpression", min(start+1, end), max(start+1, end-1))
	}
	if n.DotDotDotToken != nil {
		return p.addNode("JSXSpreadChild", node, nodeValue(expression))
	}
	return p.addNode("JSXExpressionContainer", node, nodeValue(expression))
}

func (p *estreeProjector) convertChainOperand(node *ast.Node) uint32 {
	id := p.convert(node, false)
	if node != nil && node.Kind == ast.KindParenthesizedExpression {
		return id
	}
	return p.unwrapChain(id)
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
	panic(fmt.Errorf("missing body brace for %s", node.Kind))
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
	panic(fmt.Errorf("unsupported heritage syntax %s", node.Kind))
}
func (p *estreeProjector) convertHeritageName(node *ast.Node) uint32 {
	if node.Kind == ast.KindQualifiedName {
		data := node.AsQualifiedName()
		return p.addNode("MemberExpression", node, nodeValue(p.convertHeritageName(data.Left)), nodeValue(p.convert(data.Right, false)), boolValue(false), boolValue(false))
	}
	return p.convert(node, false)
}

func (p *estreeProjector) isDirective(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	var statements *ast.NodeList
	if parent.Kind == ast.KindSourceFile {
		statements = parent.AsSourceFile().Statements
	} else if parent.Kind == ast.KindBlock && parent.Parent != nil && ast.IsFunctionLike(parent.Parent) {
		statements = parent.AsBlock().Statements
	} else {
		return false
	}
	for _, statement := range statements.Nodes {
		if statement == node {
			return true
		}
		if statement.Kind != ast.KindExpressionStatement || statement.AsExpressionStatement().Expression.Kind != ast.KindStringLiteral {
			return false
		}
	}
	return false
}
