package no_restricted_syntax

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// estreeKindMap maps an ESTree node-type name to the tsgo ast.Kind values it
// can match. ESLint selectors are written against ESTree, so this is the
// translation table the matcher consults when comparing a selector
// identifier against a tsgo node.
//
// A few ESTree types map to multiple tsgo kinds because tsgo splits cases
// that ESTree fuses (for example, `Property` covers PropertyAssignment +
// ShorthandPropertyAssignment + MethodDeclaration + Get/Set accessors, and
// `Literal` covers every literal-kind variant). The matcher treats this map
// as exhaustive — kinds not listed here cannot match the corresponding
// ESTree name.
var estreeKindMap = map[string][]ast.Kind{
	// Top-level
	"Program":    {ast.KindSourceFile},
	"SourceFile": {ast.KindSourceFile},

	// Statements
	"BlockStatement":      {ast.KindBlock},
	"BreakStatement":      {ast.KindBreakStatement},
	"ContinueStatement":   {ast.KindContinueStatement},
	"DebuggerStatement":   {ast.KindDebuggerStatement},
	"DoWhileStatement":    {ast.KindDoStatement},
	"EmptyStatement":      {ast.KindEmptyStatement},
	"ExpressionStatement": {ast.KindExpressionStatement},
	"ForInStatement":      {ast.KindForInStatement},
	"ForOfStatement":      {ast.KindForOfStatement},
	"ForStatement":        {ast.KindForStatement},
	"IfStatement":         {ast.KindIfStatement},
	"LabeledStatement":    {ast.KindLabeledStatement},
	"ReturnStatement":     {ast.KindReturnStatement},
	"SwitchStatement":     {ast.KindSwitchStatement},
	"ThrowStatement":      {ast.KindThrowStatement},
	"TryStatement":        {ast.KindTryStatement},
	"WhileStatement":      {ast.KindWhileStatement},
	"WithStatement":       {ast.KindWithStatement},

	// Switch parts
	"SwitchCase": {ast.KindCaseClause, ast.KindDefaultClause},

	// Variable & function declarations
	// ESTree's VariableDeclaration is the statement-level node (`var a = 1;`).
	// tsgo represents that with VariableStatement; the inner declarators
	// become VariableDeclaration nodes (mapped as VariableDeclarator below).
	"VariableDeclaration":     {ast.KindVariableStatement},
	"VariableDeclarator":      {ast.KindVariableDeclaration},
	"FunctionDeclaration":     {ast.KindFunctionDeclaration},
	"FunctionExpression":      {ast.KindFunctionExpression},
	"ArrowFunctionExpression": {ast.KindArrowFunction},

	// Classes
	"ClassDeclaration":   {ast.KindClassDeclaration},
	"ClassExpression":    {ast.KindClassExpression},
	"MethodDefinition":   {ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor},
	"PropertyDefinition": {ast.KindPropertyDeclaration},
	"StaticBlock":        {ast.KindClassStaticBlockDeclaration},
	"Decorator":          {ast.KindDecorator},

	// Expressions
	"ArrayExpression":          {ast.KindArrayLiteralExpression},
	"AssignmentExpression":     {ast.KindBinaryExpression}, // matcher checks operator
	"AwaitExpression":          {ast.KindAwaitExpression},
	"BinaryExpression":         {ast.KindBinaryExpression}, // matcher checks operator
	"CallExpression":           {ast.KindCallExpression},
	"ChainExpression":          {ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression, ast.KindNonNullExpression},
	"ConditionalExpression":    {ast.KindConditionalExpression},
	"LogicalExpression":        {ast.KindBinaryExpression}, // matcher checks operator
	"MemberExpression":         {ast.KindPropertyAccessExpression, ast.KindElementAccessExpression},
	"MetaProperty":             {ast.KindMetaProperty},
	"NewExpression":            {ast.KindNewExpression},
	"ObjectExpression":         {ast.KindObjectLiteralExpression},
	"SequenceExpression":       {ast.KindBinaryExpression}, // matcher checks operator (comma)
	"SpreadElement":            {ast.KindSpreadElement, ast.KindSpreadAssignment},
	"Super":                    {ast.KindSuperKeyword},
	"TaggedTemplateExpression": {ast.KindTaggedTemplateExpression},
	"TemplateElement":          {ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail},
	"TemplateLiteral":          {ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression},
	"ThisExpression":           {ast.KindThisKeyword},
	"UnaryExpression":          {ast.KindPrefixUnaryExpression, ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindDeleteExpression},
	"UpdateExpression":         {ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression},
	"YieldExpression":          {ast.KindYieldExpression},

	// Literals
	"Literal": {
		ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
	},
	"RegExpLiteral": {ast.KindRegularExpressionLiteral},

	// Identifiers
	"Identifier":        {ast.KindIdentifier},
	"PrivateIdentifier": {ast.KindPrivateIdentifier},

	// Object literal members
	"Property": {
		ast.KindPropertyAssignment,
		ast.KindShorthandPropertyAssignment,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
	},

	// Patterns
	"ArrayPattern":      {ast.KindArrayBindingPattern},
	"ObjectPattern":     {ast.KindObjectBindingPattern},
	"RestElement":       {ast.KindBindingElement, ast.KindParameter},
	"AssignmentPattern": {ast.KindBindingElement, ast.KindParameter},

	// Catch
	"CatchClause": {ast.KindCatchClause},

	// Modules
	"ImportDeclaration":        {ast.KindImportDeclaration},
	"ImportSpecifier":          {ast.KindImportSpecifier},
	"ImportDefaultSpecifier":   {ast.KindImportClause},
	"ImportNamespaceSpecifier": {ast.KindNamespaceImport},
	"ExportNamedDeclaration":   {ast.KindExportDeclaration},
	"ExportDefaultDeclaration": {ast.KindExportAssignment},
	"ExportAllDeclaration":     {ast.KindExportDeclaration},
	"ExportSpecifier":          {ast.KindExportSpecifier},

	// JSX (only meaningful when the file is parsed with JSX enabled)
	"JSXElement":             {ast.KindJsxElement, ast.KindJsxSelfClosingElement},
	"JSXFragment":            {ast.KindJsxFragment},
	"JSXOpeningElement":      {ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement},
	"JSXClosingElement":      {ast.KindJsxClosingElement},
	"JSXOpeningFragment":     {ast.KindJsxOpeningFragment},
	"JSXClosingFragment":     {ast.KindJsxClosingFragment},
	"JSXAttribute":           {ast.KindJsxAttribute},
	"JSXSpreadAttribute":     {ast.KindJsxSpreadAttribute},
	"JSXExpressionContainer": {ast.KindJsxExpression},
	"JSXText":                {ast.KindJsxText},
	"JSXIdentifier":          {ast.KindIdentifier}, // identifiers inside JSX share the regular kind
	"JSXNamespacedName":      {ast.KindJsxNamespacedName},
}

// kindsForEstreeName returns the tsgo kinds that match the given ESTree
// type name. Returns nil if the name is unknown — the matcher treats
// unknown names as never matching.
func kindsForEstreeName(name string) []ast.Kind {
	if kinds, ok := estreeKindMap[name]; ok {
		return kinds
	}
	return nil
}

// allInterestingKinds is the universe of tsgo kinds the wildcard `*`
// should listen on. We exclude pure trivia / token-only kinds the
// framework's visitor never hands to listeners (e.g. punctuation) and
// keep every syntactic kind that can show up as a node.
//
// The list is built lazily by enumerating estreeKindMap plus the small
// set of TypeScript-specific kinds the map doesn't cover but that ought
// to participate in `*` semantics.
var allInterestingKinds = func() []ast.Kind {
	seen := make(map[ast.Kind]struct{})
	for _, kinds := range estreeKindMap {
		for _, k := range kinds {
			seen[k] = struct{}{}
		}
	}
	// TypeScript-specific kinds that are common in real programs and
	// that users might select with `*` ~ `*`-style wildcards.
	extra := []ast.Kind{
		ast.KindParenthesizedExpression,
		ast.KindAsExpression,
		ast.KindSatisfiesExpression,
		ast.KindNonNullExpression,
		ast.KindTypeAssertionExpression,
		ast.KindTaggedTemplateExpression,
		ast.KindParameter,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindBindingElement,
		ast.KindSourceFile,
	}
	for _, k := range extra {
		seen[k] = struct{}{}
	}
	out := make([]ast.Kind, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}()
