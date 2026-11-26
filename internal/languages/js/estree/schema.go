package estree

import (
	"fmt"
	"math"
	"strings"
)

// NodeID identifies a projected node inside a Tree. NoNode (0) represents an
// absent node (null child or hole in a sparse array).
type NodeID uint32

const NoNode NodeID = 0

// TypeID densely identifies an ESTree/TSESTree node schema (0..TypeCount()-1).
// UnknownType is the sentinel for unrecognized or wildcard type names because
// Program occupies TypeID 0.
type TypeID uint16

const UnknownType TypeID = math.MaxUint16

// FieldID densely identifies a public ESTree field name (1..fieldCount-1).
// NoField (0) represents an unknown or absent field.
type FieldID uint8

const NoField FieldID = 0

// Schema is ordered identically to FRAME_SCHEMAS in the Node decoder.
// Fields are final ESTree fields; $ fields carry JS-only literal values.
// visitorKeys preserves the canonical @typescript-eslint/visitor-keys child order.
type nodeSchema struct {
	name        string
	fields      []string
	visitorKeys []string
}

const (
	typeProgram TypeID = iota
	typeBlockStatement
	typeTSModuleBlock
	typeEmptyStatement
	typeDebuggerStatement
	typeExpressionStatement
	typeReturnStatement
	typeThrowStatement
	typeIfStatement
	typeWhileStatement
	typeDoWhileStatement
	typeForStatement
	typeForInStatement
	typeForOfStatement
	typeBreakStatement
	typeContinueStatement
	typeWithStatement
	typeSwitchStatement
	typeSwitchCase
	typeTryStatement
	typeCatchClause
	typeLabeledStatement
	typeVariableDeclaration
	typeVariableDeclarator
	typeFunctionDeclaration
	typeTSDeclareFunction
	typeFunctionExpression
	typeTSEmptyBodyFunctionExpression
	typeTSClassImplements
	typeArrowFunctionExpression
	typeIdentifier
	typePrivateIdentifier
	typeThisExpression
	typeSuper
	typeLiteral
	typeBinaryExpression
	typeLogicalExpression
	typeAssignmentExpression
	typeSequenceExpression
	typeConditionalExpression
	typeUnaryExpression
	typeUpdateExpression
	typeAwaitExpression
	typeYieldExpression
	typeTSAsExpression
	typeTSSatisfiesExpression
	typeTSNonNullExpression
	typeTSTypeAssertion
	typeChainExpression
	typeMemberExpression
	typeCallExpression
	typeNewExpression
	typeImportExpression
	typeArrayExpression
	typeObjectExpression
	typeProperty
	typeSpreadElement
	typeRestElement
	typeObjectPattern
	typeArrayPattern
	typeAssignmentPattern
	typeTemplateLiteral
	typeTemplateElement
	typeTaggedTemplateExpression
	typeClassDeclaration
	typeClassExpression
	typeClassBody
	typeMethodDefinition
	typePropertyDefinition
	typeStaticBlock
	typeTSParameterProperty
	typeDecorator
	typeImportDeclaration
	typeImportDefaultSpecifier
	typeImportNamespaceSpecifier
	typeImportSpecifier
	typeExportNamedDeclaration
	typeExportAllDeclaration
	typeExportSpecifier
	typeExportDefaultDeclaration
	typeTSExportAssignment
	typeImportAttribute
	typeTSInterfaceDeclaration
	typeTSInterfaceBody
	typeTSTypeAliasDeclaration
	typeTSPropertySignature
	typeTSMethodSignature
	typeTSIndexSignature
	typeTSTypeReference
	typeTSQualifiedName
	typeTSTypeLiteral
	typeTSUnionType
	typeTSIntersectionType
	typeTSArrayType
	typeTSLiteralType
	typeTSTypeAnnotation
	typeTSTypeParameterDeclaration
	typeTSTypeParameterInstantiation
	typeTSTypeParameter
	typeTSStringKeyword
	typeTSNumberKeyword
	typeTSBooleanKeyword
	typeTSAnyKeyword
	typeTSUnknownKeyword
	typeTSVoidKeyword
	typeTSUndefinedKeyword
	typeTSNeverKeyword
	typeTSObjectKeyword
	typeTSSymbolKeyword
	typeTSBigIntKeyword
	typeTSIntrinsicKeyword
	typeTSFunctionType
	typeTSConstructorType
	typeTSTupleType
	typeTSNamedTupleMember
	typeTSOptionalType
	typeTSRestType
	typeTSIndexedAccessType
	typeTSTypeQuery
	typeTSTypeOperator
	typeTSInterfaceHeritage
	typeJSXElement
	typeJSXOpeningElement
	typeJSXClosingElement
	typeJSXFragment
	typeJSXOpeningFragment
	typeJSXClosingFragment
	typeJSXIdentifier
	typeJSXMemberExpression
	typeJSXNamespacedName
	typeJSXAttribute
	typeJSXSpreadAttribute
	typeJSXExpressionContainer
	typeJSXEmptyExpression
	typeJSXSpreadChild
	typeJSXText
	typeMetaProperty
	typeTSCallSignatureDeclaration
	typeTSConstructSignatureDeclaration
	typeTSTypePredicate
	typeTSImportType
	typeTSEnumBody
	typeTSEnumDeclaration
	typeTSEnumMember
	typeTSModuleDeclaration
	typeTSImportEqualsDeclaration
	typeTSExternalModuleReference
	typeTSConditionalType
	typeTSMappedType
	typeTSInferType
	typeTSTemplateLiteralType
	typeTSThisType
	typeTSAbstractMethodDefinition
	typeTSAbstractPropertyDefinition
	typeAccessorProperty
	typeTSAbstractAccessorProperty
	typeTSInstantiationExpression
	typeTSNamespaceExportDeclaration
	typeTSNullKeyword
	schemaCount
)

var schemas = [...]nodeSchema{
	typeProgram:                         {"Program", []string{"body", "sourceType"}, []string{"body"}},
	typeBlockStatement:                  {"BlockStatement", []string{"body"}, []string{"body"}},
	typeTSModuleBlock:                   {"TSModuleBlock", []string{"body"}, []string{"body"}},
	typeEmptyStatement:                  {"EmptyStatement", []string{}, nil},
	typeDebuggerStatement:               {"DebuggerStatement", []string{}, nil},
	typeExpressionStatement:             {"ExpressionStatement", []string{"expression", "directive"}, []string{"expression"}},
	typeReturnStatement:                 {"ReturnStatement", []string{"argument"}, []string{"argument"}},
	typeThrowStatement:                  {"ThrowStatement", []string{"argument"}, []string{"argument"}},
	typeIfStatement:                     {"IfStatement", []string{"test", "consequent", "alternate"}, []string{"test", "consequent", "alternate"}},
	typeWhileStatement:                  {"WhileStatement", []string{"test", "body"}, []string{"test", "body"}},
	typeDoWhileStatement:                {"DoWhileStatement", []string{"body", "test"}, []string{"body", "test"}},
	typeForStatement:                    {"ForStatement", []string{"init", "test", "update", "body"}, []string{"init", "test", "update", "body"}},
	typeForInStatement:                  {"ForInStatement", []string{"left", "right", "body"}, []string{"left", "right", "body"}},
	typeForOfStatement:                  {"ForOfStatement", []string{"await", "left", "right", "body"}, []string{"left", "right", "body"}},
	typeBreakStatement:                  {"BreakStatement", []string{"label"}, []string{"label"}},
	typeContinueStatement:               {"ContinueStatement", []string{"label"}, []string{"label"}},
	typeWithStatement:                   {"WithStatement", []string{"object", "body"}, []string{"object", "body"}},
	typeSwitchStatement:                 {"SwitchStatement", []string{"discriminant", "cases"}, []string{"discriminant", "cases"}},
	typeSwitchCase:                      {"SwitchCase", []string{"test", "consequent"}, []string{"test", "consequent"}},
	typeTryStatement:                    {"TryStatement", []string{"block", "handler", "finalizer"}, []string{"block", "handler", "finalizer"}},
	typeCatchClause:                     {"CatchClause", []string{"param", "body"}, []string{"param", "body"}},
	typeLabeledStatement:                {"LabeledStatement", []string{"label", "body"}, []string{"label", "body"}},
	typeVariableDeclaration:             {"VariableDeclaration", []string{"declarations", "declare", "kind"}, []string{"declarations"}},
	typeVariableDeclarator:              {"VariableDeclarator", []string{"id", "init", "definite"}, []string{"id", "init"}},
	typeFunctionDeclaration:             {"FunctionDeclaration", []string{"id", "params", "body", "generator", "async", "expression", "declare", "typeParameters", "returnType"}, []string{"id", "typeParameters", "params", "returnType", "body"}},
	typeTSDeclareFunction:               {"TSDeclareFunction", []string{"id", "params", "body", "generator", "async", "expression", "declare", "typeParameters", "returnType"}, []string{"id", "typeParameters", "params", "returnType", "body"}},
	typeFunctionExpression:              {"FunctionExpression", []string{"id", "params", "body", "generator", "async", "expression", "declare", "typeParameters", "returnType"}, []string{"id", "typeParameters", "params", "returnType", "body"}},
	typeTSEmptyBodyFunctionExpression:   {"TSEmptyBodyFunctionExpression", []string{"id", "params", "body", "generator", "async", "expression", "declare", "typeParameters", "returnType"}, []string{"id", "typeParameters", "params", "returnType"}},
	typeTSClassImplements:               {"TSClassImplements", []string{"expression", "typeArguments"}, []string{"expression", "typeArguments"}},
	typeArrowFunctionExpression:         {"ArrowFunctionExpression", []string{"id", "params", "body", "generator", "async", "expression", "typeParameters", "returnType"}, []string{"typeParameters", "params", "returnType", "body"}},
	typeIdentifier:                      {"Identifier", []string{"decorators", "name", "optional", "typeAnnotation"}, []string{"decorators", "typeAnnotation"}},
	typePrivateIdentifier:               {"PrivateIdentifier", []string{"name"}, nil},
	typeThisExpression:                  {"ThisExpression", []string{}, nil},
	typeSuper:                           {"Super", []string{}, nil},
	typeLiteral:                         {"Literal", []string{"value", "raw", "bigint", "$regexPattern", "$regexFlags"}, nil},
	typeBinaryExpression:                {"BinaryExpression", []string{"left", "operator", "right"}, []string{"left", "right"}},
	typeLogicalExpression:               {"LogicalExpression", []string{"left", "operator", "right"}, []string{"left", "right"}},
	typeAssignmentExpression:            {"AssignmentExpression", []string{"left", "operator", "right"}, []string{"left", "right"}},
	typeSequenceExpression:              {"SequenceExpression", []string{"expressions"}, []string{"expressions"}},
	typeConditionalExpression:           {"ConditionalExpression", []string{"test", "consequent", "alternate"}, []string{"test", "consequent", "alternate"}},
	typeUnaryExpression:                 {"UnaryExpression", []string{"operator", "prefix", "argument"}, []string{"argument"}},
	typeUpdateExpression:                {"UpdateExpression", []string{"operator", "prefix", "argument"}, []string{"argument"}},
	typeAwaitExpression:                 {"AwaitExpression", []string{"argument"}, []string{"argument"}},
	typeYieldExpression:                 {"YieldExpression", []string{"argument", "delegate"}, []string{"argument"}},
	typeTSAsExpression:                  {"TSAsExpression", []string{"expression", "typeAnnotation"}, []string{"expression", "typeAnnotation"}},
	typeTSSatisfiesExpression:           {"TSSatisfiesExpression", []string{"expression", "typeAnnotation"}, []string{"expression", "typeAnnotation"}},
	typeTSNonNullExpression:             {"TSNonNullExpression", []string{"expression"}, []string{"expression"}},
	typeTSTypeAssertion:                 {"TSTypeAssertion", []string{"expression", "typeAnnotation"}, []string{"typeAnnotation", "expression"}},
	typeChainExpression:                 {"ChainExpression", []string{"expression"}, []string{"expression"}},
	typeMemberExpression:                {"MemberExpression", []string{"object", "property", "computed", "optional"}, []string{"object", "property"}},
	typeCallExpression:                  {"CallExpression", []string{"callee", "arguments", "optional", "typeArguments"}, []string{"callee", "typeArguments", "arguments"}},
	typeNewExpression:                   {"NewExpression", []string{"callee", "arguments", "typeArguments"}, []string{"callee", "typeArguments", "arguments"}},
	typeImportExpression:                {"ImportExpression", []string{"source", "options"}, []string{"source", "options"}},
	typeArrayExpression:                 {"ArrayExpression", []string{"elements"}, []string{"elements"}},
	typeObjectExpression:                {"ObjectExpression", []string{"properties"}, []string{"properties"}},
	typeProperty:                        {"Property", []string{"key", "value", "kind", "method", "shorthand", "computed", "optional"}, []string{"key", "value"}},
	typeSpreadElement:                   {"SpreadElement", []string{"argument"}, []string{"argument"}},
	typeRestElement:                     {"RestElement", []string{"decorators", "argument", "optional", "typeAnnotation"}, []string{"decorators", "argument", "typeAnnotation"}},
	typeObjectPattern:                   {"ObjectPattern", []string{"properties", "decorators", "optional", "typeAnnotation"}, []string{"decorators", "properties", "typeAnnotation"}},
	typeArrayPattern:                    {"ArrayPattern", []string{"elements", "decorators", "optional", "typeAnnotation"}, []string{"decorators", "elements", "typeAnnotation"}},
	typeAssignmentPattern:               {"AssignmentPattern", []string{"decorators", "left", "optional", "right", "typeAnnotation"}, []string{"decorators", "left", "right", "typeAnnotation"}},
	typeTemplateLiteral:                 {"TemplateLiteral", []string{"expressions", "quasis"}, []string{"quasis", "expressions"}},
	typeTemplateElement:                 {"TemplateElement", []string{"tail", "$cooked", "$raw"}, nil},
	typeTaggedTemplateExpression:        {"TaggedTemplateExpression", []string{"tag", "quasi", "typeArguments"}, []string{"tag", "typeArguments", "quasi"}},
	typeClassDeclaration:                {"ClassDeclaration", []string{"abstract", "body", "declare", "decorators", "id", "implements", "superClass", "superTypeArguments", "typeParameters"}, []string{"decorators", "id", "typeParameters", "superClass", "superTypeArguments", "implements", "body"}},
	typeClassExpression:                 {"ClassExpression", []string{"abstract", "body", "declare", "decorators", "id", "implements", "superClass", "superTypeArguments", "typeParameters"}, []string{"decorators", "id", "typeParameters", "superClass", "superTypeArguments", "implements", "body"}},
	typeClassBody:                       {"ClassBody", []string{"body"}, []string{"body"}},
	typeMethodDefinition:                {"MethodDefinition", []string{"accessibility", "computed", "decorators", "key", "kind", "optional", "override", "static", "value"}, []string{"decorators", "key", "value"}},
	typePropertyDefinition:              {"PropertyDefinition", []string{"accessibility", "computed", "declare", "decorators", "definite", "key", "optional", "override", "readonly", "static", "typeAnnotation", "value"}, []string{"decorators", "key", "typeAnnotation", "value"}},
	typeStaticBlock:                     {"StaticBlock", []string{"body"}, []string{"body"}},
	typeTSParameterProperty:             {"TSParameterProperty", []string{"accessibility", "decorators", "override", "parameter", "readonly", "static"}, []string{"decorators", "parameter"}},
	typeDecorator:                       {"Decorator", []string{"expression"}, []string{"expression"}},
	typeImportDeclaration:               {"ImportDeclaration", []string{"attributes", "importKind", "source", "specifiers", "phase"}, []string{"specifiers", "source", "attributes"}},
	typeImportDefaultSpecifier:          {"ImportDefaultSpecifier", []string{"local"}, []string{"local"}},
	typeImportNamespaceSpecifier:        {"ImportNamespaceSpecifier", []string{"local"}, []string{"local"}},
	typeImportSpecifier:                 {"ImportSpecifier", []string{"imported", "importKind", "local"}, []string{"imported", "local"}},
	typeExportNamedDeclaration:          {"ExportNamedDeclaration", []string{"attributes", "declaration", "exportKind", "source", "specifiers"}, []string{"declaration", "specifiers", "source", "attributes"}},
	typeExportAllDeclaration:            {"ExportAllDeclaration", []string{"attributes", "exported", "exportKind", "source"}, []string{"exported", "source", "attributes"}},
	typeExportSpecifier:                 {"ExportSpecifier", []string{"exported", "exportKind", "local"}, []string{"local", "exported"}},
	typeExportDefaultDeclaration:        {"ExportDefaultDeclaration", []string{"declaration", "exportKind"}, []string{"declaration"}},
	typeTSExportAssignment:              {"TSExportAssignment", []string{"expression"}, []string{"expression"}},
	typeImportAttribute:                 {"ImportAttribute", []string{"key", "value"}, []string{"key", "value"}},
	typeTSInterfaceDeclaration:          {"TSInterfaceDeclaration", []string{"id", "typeParameters", "extends", "body", "declare"}, []string{"id", "typeParameters", "extends", "body"}},
	typeTSInterfaceBody:                 {"TSInterfaceBody", []string{"body"}, []string{"body"}},
	typeTSTypeAliasDeclaration:          {"TSTypeAliasDeclaration", []string{"declare", "id", "typeAnnotation", "typeParameters"}, []string{"id", "typeParameters", "typeAnnotation"}},
	typeTSPropertySignature:             {"TSPropertySignature", []string{"accessibility", "computed", "key", "optional", "readonly", "static", "typeAnnotation"}, []string{"key", "typeAnnotation"}},
	typeTSMethodSignature:               {"TSMethodSignature", []string{"accessibility", "computed", "key", "kind", "optional", "params", "readonly", "returnType", "static", "typeParameters"}, []string{"key", "typeParameters", "params", "returnType"}},
	typeTSIndexSignature:                {"TSIndexSignature", []string{"accessibility", "parameters", "readonly", "static", "typeAnnotation"}, []string{"parameters", "typeAnnotation"}},
	typeTSTypeReference:                 {"TSTypeReference", []string{"typeName", "typeArguments"}, []string{"typeName", "typeArguments"}},
	typeTSQualifiedName:                 {"TSQualifiedName", []string{"left", "right"}, []string{"left", "right"}},
	typeTSTypeLiteral:                   {"TSTypeLiteral", []string{"members"}, []string{"members"}},
	typeTSUnionType:                     {"TSUnionType", []string{"types"}, []string{"types"}},
	typeTSIntersectionType:              {"TSIntersectionType", []string{"types"}, []string{"types"}},
	typeTSArrayType:                     {"TSArrayType", []string{"elementType"}, []string{"elementType"}},
	typeTSLiteralType:                   {"TSLiteralType", []string{"literal"}, []string{"literal"}},
	typeTSTypeAnnotation:                {"TSTypeAnnotation", []string{"typeAnnotation"}, []string{"typeAnnotation"}},
	typeTSTypeParameterDeclaration:      {"TSTypeParameterDeclaration", []string{"params"}, []string{"params"}},
	typeTSTypeParameterInstantiation:    {"TSTypeParameterInstantiation", []string{"params"}, []string{"params"}},
	typeTSTypeParameter:                 {"TSTypeParameter", []string{"const", "constraint", "default", "in", "name", "out"}, []string{"name", "constraint", "default"}},
	typeTSStringKeyword:                 {"TSStringKeyword", []string{}, nil},
	typeTSNumberKeyword:                 {"TSNumberKeyword", []string{}, nil},
	typeTSBooleanKeyword:                {"TSBooleanKeyword", []string{}, nil},
	typeTSAnyKeyword:                    {"TSAnyKeyword", []string{}, nil},
	typeTSUnknownKeyword:                {"TSUnknownKeyword", []string{}, nil},
	typeTSVoidKeyword:                   {"TSVoidKeyword", []string{}, nil},
	typeTSUndefinedKeyword:              {"TSUndefinedKeyword", []string{}, nil},
	typeTSNeverKeyword:                  {"TSNeverKeyword", []string{}, nil},
	typeTSObjectKeyword:                 {"TSObjectKeyword", []string{}, nil},
	typeTSSymbolKeyword:                 {"TSSymbolKeyword", []string{}, nil},
	typeTSBigIntKeyword:                 {"TSBigIntKeyword", []string{}, nil},
	typeTSIntrinsicKeyword:              {"TSIntrinsicKeyword", []string{}, nil},
	typeTSFunctionType:                  {"TSFunctionType", []string{"params", "returnType", "typeParameters"}, []string{"typeParameters", "params", "returnType"}},
	typeTSConstructorType:               {"TSConstructorType", []string{"abstract", "params", "returnType", "typeParameters"}, []string{"typeParameters", "params", "returnType"}},
	typeTSTupleType:                     {"TSTupleType", []string{"elementTypes"}, []string{"elementTypes"}},
	typeTSNamedTupleMember:              {"TSNamedTupleMember", []string{"elementType", "label", "optional"}, []string{"label", "elementType"}},
	typeTSOptionalType:                  {"TSOptionalType", []string{"typeAnnotation"}, []string{"typeAnnotation"}},
	typeTSRestType:                      {"TSRestType", []string{"typeAnnotation"}, []string{"typeAnnotation"}},
	typeTSIndexedAccessType:             {"TSIndexedAccessType", []string{"objectType", "indexType"}, []string{"objectType", "indexType"}},
	typeTSTypeQuery:                     {"TSTypeQuery", []string{"exprName", "typeArguments"}, []string{"exprName", "typeArguments"}},
	typeTSTypeOperator:                  {"TSTypeOperator", []string{"operator", "typeAnnotation"}, []string{"typeAnnotation"}},
	typeTSInterfaceHeritage:             {"TSInterfaceHeritage", []string{"expression", "typeArguments"}, []string{"expression", "typeArguments"}},
	typeJSXElement:                      {"JSXElement", []string{"openingElement", "children", "closingElement"}, []string{"openingElement", "children", "closingElement"}},
	typeJSXOpeningElement:               {"JSXOpeningElement", []string{"name", "attributes", "selfClosing", "typeArguments"}, []string{"name", "typeArguments", "attributes"}},
	typeJSXClosingElement:               {"JSXClosingElement", []string{"name"}, []string{"name"}},
	typeJSXFragment:                     {"JSXFragment", []string{"openingFragment", "children", "closingFragment"}, []string{"openingFragment", "children", "closingFragment"}},
	typeJSXOpeningFragment:              {"JSXOpeningFragment", []string{"attributes", "selfClosing"}, nil},
	typeJSXClosingFragment:              {"JSXClosingFragment", []string{}, nil},
	typeJSXIdentifier:                   {"JSXIdentifier", []string{"name"}, nil},
	typeJSXMemberExpression:             {"JSXMemberExpression", []string{"object", "property"}, []string{"object", "property"}},
	typeJSXNamespacedName:               {"JSXNamespacedName", []string{"namespace", "name"}, []string{"namespace", "name"}},
	typeJSXAttribute:                    {"JSXAttribute", []string{"name", "value"}, []string{"name", "value"}},
	typeJSXSpreadAttribute:              {"JSXSpreadAttribute", []string{"argument"}, []string{"argument"}},
	typeJSXExpressionContainer:          {"JSXExpressionContainer", []string{"expression"}, []string{"expression"}},
	typeJSXEmptyExpression:              {"JSXEmptyExpression", []string{}, nil},
	typeJSXSpreadChild:                  {"JSXSpreadChild", []string{"expression"}, []string{"expression"}},
	typeJSXText:                         {"JSXText", []string{"value", "raw"}, nil},
	typeMetaProperty:                    {"MetaProperty", []string{"meta", "property"}, []string{"meta", "property"}},
	typeTSCallSignatureDeclaration:      {"TSCallSignatureDeclaration", []string{"params", "returnType", "typeParameters"}, []string{"typeParameters", "params", "returnType"}},
	typeTSConstructSignatureDeclaration: {"TSConstructSignatureDeclaration", []string{"params", "returnType", "typeParameters"}, []string{"typeParameters", "params", "returnType"}},
	typeTSTypePredicate:                 {"TSTypePredicate", []string{"asserts", "parameterName", "typeAnnotation"}, []string{"parameterName", "typeAnnotation"}},
	typeTSImportType:                    {"TSImportType", []string{"options", "qualifier", "source", "typeArguments"}, []string{"source", "options", "qualifier", "typeArguments"}},
	typeTSEnumBody:                      {"TSEnumBody", []string{"members"}, []string{"members"}},
	typeTSEnumDeclaration:               {"TSEnumDeclaration", []string{"body", "const", "declare", "id"}, []string{"id", "body"}},
	typeTSEnumMember:                    {"TSEnumMember", []string{"id", "initializer"}, []string{"id", "initializer"}},
	typeTSModuleDeclaration:             {"TSModuleDeclaration", []string{"body", "declare", "global", "id", "kind"}, []string{"id", "body"}},
	typeTSImportEqualsDeclaration:       {"TSImportEqualsDeclaration", []string{"id", "importKind", "moduleReference"}, []string{"id", "moduleReference"}},
	typeTSExternalModuleReference:       {"TSExternalModuleReference", []string{"expression"}, []string{"expression"}},
	typeTSConditionalType:               {"TSConditionalType", []string{"checkType", "extendsType", "falseType", "trueType"}, []string{"checkType", "extendsType", "trueType", "falseType"}},
	typeTSMappedType:                    {"TSMappedType", []string{"constraint", "key", "nameType", "optional", "readonly", "typeAnnotation"}, []string{"key", "constraint", "nameType", "typeAnnotation"}},
	typeTSInferType:                     {"TSInferType", []string{"typeParameter"}, []string{"typeParameter"}},
	typeTSTemplateLiteralType:           {"TSTemplateLiteralType", []string{"quasis", "types"}, []string{"quasis", "types"}},
	typeTSThisType:                      {"TSThisType", []string{}, nil},
	typeTSAbstractMethodDefinition:      {"TSAbstractMethodDefinition", []string{"accessibility", "computed", "decorators", "key", "kind", "optional", "override", "static", "value"}, []string{"key", "value"}},
	typeTSAbstractPropertyDefinition:    {"TSAbstractPropertyDefinition", []string{"accessibility", "computed", "declare", "decorators", "definite", "key", "optional", "override", "readonly", "static", "typeAnnotation", "value"}, []string{"decorators", "key", "typeAnnotation"}},
	typeAccessorProperty:                {"AccessorProperty", []string{"accessibility", "computed", "declare", "decorators", "definite", "key", "optional", "override", "readonly", "static", "typeAnnotation", "value"}, []string{"decorators", "key", "typeAnnotation", "value"}},
	typeTSAbstractAccessorProperty:      {"TSAbstractAccessorProperty", []string{"accessibility", "computed", "declare", "decorators", "definite", "key", "optional", "override", "readonly", "static", "typeAnnotation", "value"}, []string{"decorators", "key", "typeAnnotation"}},
	typeTSInstantiationExpression:       {"TSInstantiationExpression", []string{"expression", "typeArguments"}, []string{"expression", "typeArguments"}},
	typeTSNamespaceExportDeclaration:    {"TSNamespaceExportDeclaration", []string{"id"}, []string{"id"}},
	typeTSNullKeyword:                   {"TSNullKeyword", []string{}, nil},
}

const maxFieldID = 128

type childSlot struct {
	slot  uint8
	field FieldID
}

var (
	typeIDs         map[string]uint16
	fieldIDs        map[string]FieldID
	fieldNames      []string
	fieldSlot       [schemaCount][maxFieldID]int8
	childSlots      [schemaCount][]childSlot
	jsUndefinedMask [schemaCount]uint16

	fieldValue          FieldID
	fieldRegex          FieldID
	fieldTypeAnnotation FieldID
	fieldOptional       FieldID
	fieldDecorators     FieldID
)

const (
	slotLiteralValue        = 0
	slotLiteralRaw          = 1
	slotLiteralBigInt       = 2
	slotLiteralRegexPattern = 3
	slotLiteralRegexFlags   = 4

	slotTemplateTail   = 0
	slotTemplateCooked = 1
	slotTemplateRaw    = 2
)

func init() {
	typeIDs = make(map[string]uint16, len(schemas))
	fieldIDs = make(map[string]FieldID, maxFieldID)
	fieldNames = []string{""} // NoField == 0

	internField := func(name string) FieldID {
		if id, ok := fieldIDs[name]; ok {
			return id
		}
		if len(fieldNames) >= maxFieldID {
			panic(fmt.Sprintf("estree field count exceeds %d", maxFieldID-1))
		}
		id := FieldID(len(fieldNames))
		fieldNames = append(fieldNames, name)
		fieldIDs[name] = id
		return id
	}

	for i := range fieldSlot {
		for j := range fieldSlot[i] {
			fieldSlot[i][j] = -1
		}
	}

	for id := range schemas {
		schema := &schemas[id]
		if len(schema.fields) > 16 {
			panic(fmt.Sprintf("estree schema %s has %d fields (> 16)", schema.name, len(schema.fields)))
		}
		typeIDs[schema.name] = uint16(id)

		var jsMask uint16
		for slot, field := range schema.fields {
			if strings.HasPrefix(field, "$") {
				continue
			}
			fid := internField(field)
			fieldSlot[id][fid] = int8(slot)

			switch field {
			case "abstract", "accessibility", "declare", "definite", "exportKind", "importKind",
				"implements", "override", "readonly", "returnType", "superTypeArguments",
				"typeAnnotation", "typeArguments", "typeParameters":
				jsMask |= 1 << slot
			case "optional":
				if TypeID(id) != typeMemberExpression && TypeID(id) != typeCallExpression {
					jsMask |= 1 << slot
				}
			case "decorators":
				switch TypeID(id) {
				case typeIdentifier, typeArrayPattern, typeObjectPattern, typeRestElement, typeAssignmentPattern:
					jsMask |= 1 << slot
				}
			}
		}
		jsUndefinedMask[id] = jsMask

		if len(schema.visitorKeys) > 0 {
			slots := make([]childSlot, len(schema.visitorKeys))
			for idx, key := range schema.visitorKeys {
				fid, ok := fieldIDs[key]
				if !ok || fieldSlot[id][fid] < 0 {
					panic(fmt.Sprintf("estree visitor key %s.%s is not a public schema field", schema.name, key))
				}
				slots[idx] = childSlot{slot: uint8(fieldSlot[id][fid]), field: fid}
			}
			childSlots[id] = slots
		}
	}

	// Register synthesized ESTree fields backed by $-prefixed wire columns.
	fieldValue = internField("value")
	fieldRegex = internField("regex")
	fieldTypeAnnotation = internField("typeAnnotation")
	fieldOptional = internField("optional")
	fieldDecorators = internField("decorators")
}

// TypeCount returns the number of ESTree/TSESTree node schemas.
func TypeCount() int {
	return len(schemas)
}

// TypeIDByName resolves an exact-case ESTree node type name to its TypeID.
func TypeIDByName(name string) (TypeID, bool) {
	id, ok := typeIDs[name]
	if !ok {
		return UnknownType, false
	}
	return TypeID(id), true
}

// String returns the ESTree node type name for id, or "" for UnknownType.
func (id TypeID) String() string {
	if int(id) >= len(schemas) {
		return ""
	}
	return schemas[id].name
}

// FieldIDByName resolves a public ESTree field name to its FieldID.
// Internal $-prefixed wire fields are never exposed.
func FieldIDByName(name string) (FieldID, bool) {
	id, ok := fieldIDs[name]
	return id, ok
}

// String returns the field name for id, or "" for NoField.
func (id FieldID) String() string {
	if int(id) >= len(fieldNames) {
		return ""
	}
	return fieldNames[id]
}
