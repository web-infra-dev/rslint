package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// IsLexicalDeclaration mirrors ESLint astUtils.isLexicalDeclaration: true for
// let/const/using/await-using variable declarations and function/class
// declarations.
//
// NOTE: Unlike ESLint (which only sees JavaScript), tsgo also parses
// TypeScript declarations that are equally illegal as an unbraced
// control-statement body (`if (a) enum E {}` is a syntax error). They are
// treated the same as lexical declarations so an autofix never strips a
// required block.
func IsLexicalDeclaration(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindVariableStatement:
		declList := node.AsVariableStatement().DeclarationList
		return declList.Flags&ast.NodeFlagsBlockScoped != 0
	case ast.KindFunctionDeclaration,
		ast.KindClassDeclaration,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindImportEqualsDeclaration:
		return true
	}
	return false
}

// HasUnsafeIf mirrors ESLint astUtils.areBracesNecessary's internal
// hasUnsafeIf: reports whether node contains an `if` statement that would
// become associated with an `else` keyword appended directly after node.
func HasUnsafeIf(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt.ElseStatement == nil {
			return true
		}
		return HasUnsafeIf(ifStmt.ElseStatement)
	case ast.KindForStatement:
		return HasUnsafeIf(node.AsForStatement().Statement)
	case ast.KindForInStatement, ast.KindForOfStatement:
		return HasUnsafeIf(node.AsForInOrOfStatement().Statement)
	case ast.KindLabeledStatement:
		return HasUnsafeIf(node.AsLabeledStatement().Statement)
	case ast.KindWithStatement:
		return HasUnsafeIf(node.AsWithStatement().Statement)
	case ast.KindWhileStatement:
		return HasUnsafeIf(node.AsWhileStatement().Statement)
	default:
		return false
	}
}

// IsFollowedByElseKeyword reports whether the next non-trivia token after
// node is the `else` keyword.
func IsFollowedByElseKeyword(sourceFile *ast.SourceFile, node *ast.Node) bool {
	token, ok := TokenAtOrAfter(sourceFile, node.End())
	return ok && token.Kind == ast.KindElseKeyword
}

// AreBracesNecessary mirrors ESLint astUtils.areBracesNecessary: reports
// whether the existing curly braces around a single-statement BlockStatement
// body are necessary to preserve the semantics of the code — either because
// the statement is a lexical declaration, or because it ends in an `if` that
// would capture a trailing `else` once the braces are removed.
func AreBracesNecessary(sourceFile *ast.SourceFile, block *ast.Node) bool {
	statement := block.AsBlock().Statements.Nodes[0]
	return IsLexicalDeclaration(statement) ||
		(HasUnsafeIf(statement) && IsFollowedByElseKeyword(sourceFile, block))
}
