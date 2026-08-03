package cfg

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/utils"
)

func isBreakableStatement(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindSwitchStatement, ast.KindWhileStatement, ast.KindDoStatement,
		ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		return true
	}
	return false
}

// labelsOf collects every label wrapped directly around a loop or switch, so
// `outer: inner: while (…)` resolves `continue outer` and `break outer` as
// well as `inner`. (ESLint's code path analysis attaches only the innermost
// label and loses the back edge of a `continue` naming an outer one; this
// builder keeps the edge.)
func labelsOf(node *ast.Node) []string {
	var labels []string
	current := node
	for parent := current.Parent; parent != nil && parent.Kind == ast.KindLabeledStatement; parent = parent.Parent {
		labeled := parent.AsLabeledStatement()
		if labeled.Statement != current {
			break
		}
		labels = append(labels, labeled.Label.Text())
		current = parent
	}
	return labels
}

// isAlwaysTruthyTest mirrors ESLint's `getBooleanValueIfSimpleConstant`: only a
// literal test can mark the loop infinite, so the code after it is unreachable.
func isAlwaysTruthyTest(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindRegularExpressionLiteral:
		return true
	case ast.KindNumericLiteral:
		return utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text) != "0"
	case ast.KindBigIntLiteral:
		text := node.AsBigIntLiteral().Text
		return text != "0n" && text != "0"
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text != ""
	}
	return false
}

// isThrowableIdentifier reports whether completing this identifier is one of
// the points ESLint treats as able to throw inside a `try` block. Names that
// only declare or label something are excluded, as are JSX names, which ESLint
// sees as a separate node type.
func isThrowableIdentifier(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if ast.IsJsxTagName(node) {
		return false
	}
	switch parent.Kind {
	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement,
		ast.KindArrayBindingPattern, ast.KindImportSpecifier, ast.KindExportSpecifier,
		ast.KindImportClause, ast.KindNamespaceImport, ast.KindNamespaceExport,
		ast.KindJsxAttribute, ast.KindJsxNamespacedName:
		return false
	case ast.KindBindingElement:
		binding := parent.AsBindingElement()
		if binding.DotDotDotToken != nil || binding.PropertyName == node {
			return false
		}
		return parent.Parent != nil && parent.Parent.Kind == ast.KindObjectBindingPattern
	case ast.KindCatchClause:
		return false
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindClassDeclaration, ast.KindClassExpression, ast.KindVariableDeclaration,
		ast.KindPropertyAssignment, ast.KindPropertyDeclaration, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor, ast.KindEnumDeclaration,
		ast.KindModuleDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		return parent.Name() != node
	}
	return true
}
