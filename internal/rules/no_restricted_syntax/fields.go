package no_restricted_syntax

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// nodesAtField enumerates the children of `parent` that sit at the named
// ESTree-style field. The result may have zero, one, or multiple entries
// depending on whether the field is a list-shaped collection. The matcher
// uses this to evaluate `Foo.bar`-style class selectors.
func nodesAtField(parent *ast.Node, field string) []*ast.Node {
	switch field {
	case "key":
		switch parent.Kind {
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindPropertyDeclaration:
			return single(parent.Name())
		}
	case "value":
		switch parent.Kind {
		case ast.KindPropertyAssignment:
			return single(parent.AsPropertyAssignment().Initializer)
		case ast.KindPropertyDeclaration:
			return single(parent.AsPropertyDeclaration().Initializer)
		}
	case "id":
		switch parent.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindClassDeclaration, ast.KindClassExpression, ast.KindVariableDeclaration:
			return single(parent.Name())
		}
	case "init":
		switch parent.Kind {
		case ast.KindVariableDeclaration:
			return single(parent.AsVariableDeclaration().Initializer)
		case ast.KindBindingElement:
			return single(parent.AsBindingElement().Initializer)
		case ast.KindForStatement:
			return single(parent.AsForStatement().Initializer)
		}
	case "test":
		switch parent.Kind {
		case ast.KindIfStatement:
			return single(parent.AsIfStatement().Expression)
		case ast.KindConditionalExpression:
			return single(parent.AsConditionalExpression().Condition)
		case ast.KindWhileStatement:
			return single(parent.AsWhileStatement().Expression)
		case ast.KindDoStatement:
			return single(parent.AsDoStatement().Expression)
		case ast.KindForStatement:
			return single(parent.AsForStatement().Condition)
		}
	case "consequent":
		switch parent.Kind {
		case ast.KindIfStatement:
			return single(parent.AsIfStatement().ThenStatement)
		case ast.KindConditionalExpression:
			return single(parent.AsConditionalExpression().WhenTrue)
		}
	case "alternate":
		switch parent.Kind {
		case ast.KindIfStatement:
			return single(parent.AsIfStatement().ElseStatement)
		case ast.KindConditionalExpression:
			return single(parent.AsConditionalExpression().WhenFalse)
		}
	case "body":
		switch parent.Kind {
		case ast.KindFunctionDeclaration:
			return single(parent.AsFunctionDeclaration().Body)
		case ast.KindFunctionExpression:
			return single(parent.AsFunctionExpression().Body)
		case ast.KindArrowFunction:
			return single(parent.AsArrowFunction().Body)
		case ast.KindMethodDeclaration:
			return single(parent.AsMethodDeclaration().Body)
		case ast.KindIfStatement:
			return single(parent.AsIfStatement().ThenStatement)
		case ast.KindWhileStatement:
			return single(parent.AsWhileStatement().Statement)
		case ast.KindDoStatement:
			return single(parent.AsDoStatement().Statement)
		case ast.KindForStatement:
			return single(parent.AsForStatement().Statement)
		case ast.KindBlock:
			return statements(parent.AsBlock().Statements)
		case ast.KindSourceFile:
			return statements(parent.AsSourceFile().Statements)
		case ast.KindClassDeclaration:
			return statements(parent.AsClassDeclaration().Members)
		case ast.KindClassExpression:
			return statements(parent.AsClassExpression().Members)
		}
	case "callee":
		switch parent.Kind {
		case ast.KindCallExpression:
			return single(parent.AsCallExpression().Expression)
		case ast.KindNewExpression:
			return single(parent.AsNewExpression().Expression)
		case ast.KindTaggedTemplateExpression:
			return single(parent.AsTaggedTemplateExpression().Tag)
		}
	case "arguments":
		switch parent.Kind {
		case ast.KindCallExpression:
			return statements(parent.AsCallExpression().Arguments)
		case ast.KindNewExpression:
			return statements(parent.AsNewExpression().Arguments)
		}
	case "expression":
		switch parent.Kind {
		case ast.KindExpressionStatement:
			return single(parent.AsExpressionStatement().Expression)
		case ast.KindParenthesizedExpression:
			return single(parent.AsParenthesizedExpression().Expression)
		}
	case "object":
		switch parent.Kind {
		case ast.KindPropertyAccessExpression:
			return single(parent.AsPropertyAccessExpression().Expression)
		case ast.KindElementAccessExpression:
			return single(parent.AsElementAccessExpression().Expression)
		}
	case "property":
		switch parent.Kind {
		case ast.KindPropertyAccessExpression:
			return single(parent.AsPropertyAccessExpression().Name())
		case ast.KindElementAccessExpression:
			return single(parent.AsElementAccessExpression().ArgumentExpression)
		}
	case "source":
		switch parent.Kind {
		case ast.KindImportDeclaration:
			return single(parent.AsImportDeclaration().ModuleSpecifier)
		case ast.KindExportDeclaration:
			return single(parent.AsExportDeclaration().ModuleSpecifier)
		}
	case "argument":
		switch parent.Kind {
		case ast.KindAwaitExpression:
			return single(parent.AsAwaitExpression().Expression)
		case ast.KindYieldExpression:
			return single(parent.AsYieldExpression().Expression)
		case ast.KindSpreadElement:
			return single(parent.AsSpreadElement().Expression)
		case ast.KindSpreadAssignment:
			return single(parent.AsSpreadAssignment().Expression)
		case ast.KindReturnStatement:
			return single(parent.AsReturnStatement().Expression)
		case ast.KindThrowStatement:
			return single(parent.AsThrowStatement().Expression)
		case ast.KindPrefixUnaryExpression:
			return single(parent.AsPrefixUnaryExpression().Operand)
		case ast.KindPostfixUnaryExpression:
			return single(parent.AsPostfixUnaryExpression().Operand)
		case ast.KindTypeOfExpression:
			return single(parent.AsTypeOfExpression().Expression)
		case ast.KindVoidExpression:
			return single(parent.AsVoidExpression().Expression)
		case ast.KindDeleteExpression:
			return single(parent.AsDeleteExpression().Expression)
		}
	case "left":
		switch parent.Kind {
		case ast.KindBinaryExpression:
			return single(parent.AsBinaryExpression().Left)
		case ast.KindForInStatement, ast.KindForOfStatement:
			return single(parent.AsForInOrOfStatement().Initializer)
		}
	case "right":
		switch parent.Kind {
		case ast.KindBinaryExpression:
			return single(parent.AsBinaryExpression().Right)
		case ast.KindForInStatement, ast.KindForOfStatement:
			return single(parent.AsForInOrOfStatement().Expression)
		}
	case "label":
		switch parent.Kind {
		case ast.KindBreakStatement:
			return single(parent.AsBreakStatement().Label)
		case ast.KindContinueStatement:
			return single(parent.AsContinueStatement().Label)
		case ast.KindLabeledStatement:
			return single(parent.AsLabeledStatement().Label)
		}
	case "block":
		if parent.Kind == ast.KindTryStatement {
			return single(parent.AsTryStatement().TryBlock)
		}
	case "handler":
		if parent.Kind == ast.KindTryStatement {
			return single(parent.AsTryStatement().CatchClause)
		}
	case "finalizer":
		if parent.Kind == ast.KindTryStatement {
			return single(parent.AsTryStatement().FinallyBlock)
		}
	case "param":
		if parent.Kind == ast.KindCatchClause {
			cc := parent.AsCatchClause()
			if cc.VariableDeclaration != nil {
				return single(cc.VariableDeclaration.AsVariableDeclaration().Name())
			}
		}
	case "params":
		params := parent.Parameters()
		if params == nil {
			return nil
		}
		return params
	case "elements":
		switch parent.Kind {
		case ast.KindArrayLiteralExpression:
			return statements(parent.AsArrayLiteralExpression().Elements)
		case ast.KindArrayBindingPattern:
			return statements(parent.AsBindingPattern().Elements)
		case ast.KindNamedImports:
			return statements(parent.AsNamedImports().Elements)
		case ast.KindNamedExports:
			return statements(parent.AsNamedExports().Elements)
		}
	case "properties":
		switch parent.Kind {
		case ast.KindObjectLiteralExpression:
			return statements(parent.AsObjectLiteralExpression().Properties)
		case ast.KindObjectBindingPattern:
			return statements(parent.AsBindingPattern().Elements)
		}
	case "declarations":
		if parent.Kind == ast.KindVariableStatement {
			dl := parent.AsVariableStatement().DeclarationList
			if dl != nil {
				return statements(dl.AsVariableDeclarationList().Declarations)
			}
		}
	case "specifiers":
		switch parent.Kind {
		case ast.KindImportDeclaration:
			return collectImportSpecifiers(parent)
		case ast.KindExportDeclaration:
			return collectExportSpecifiers(parent)
		}
	}
	return nil
}

// listChildrenOf returns every list-shaped collection on the parent. The
// matcher's :nth-child evaluation iterates over these lists looking for
// the searched node.
func listChildrenOf(parent *ast.Node) [][]*ast.Node {
	var out [][]*ast.Node
	add := func(s []*ast.Node) {
		if len(s) > 0 {
			out = append(out, s)
		}
	}
	switch parent.Kind {
	case ast.KindSourceFile:
		add(statements(parent.AsSourceFile().Statements))
	case ast.KindBlock:
		add(statements(parent.AsBlock().Statements))
	case ast.KindArrayLiteralExpression:
		add(statements(parent.AsArrayLiteralExpression().Elements))
	case ast.KindObjectLiteralExpression:
		add(statements(parent.AsObjectLiteralExpression().Properties))
	case ast.KindCallExpression:
		add(statements(parent.AsCallExpression().Arguments))
	case ast.KindNewExpression:
		add(statements(parent.AsNewExpression().Arguments))
	case ast.KindClassDeclaration:
		add(statements(parent.AsClassDeclaration().Members))
	case ast.KindClassExpression:
		add(statements(parent.AsClassExpression().Members))
	case ast.KindNamedImports:
		add(statements(parent.AsNamedImports().Elements))
	case ast.KindNamedExports:
		add(statements(parent.AsNamedExports().Elements))
	case ast.KindArrayBindingPattern, ast.KindObjectBindingPattern:
		add(statements(parent.AsBindingPattern().Elements))
	case ast.KindCaseClause, ast.KindDefaultClause:
		add(statements(parent.AsCaseOrDefaultClause().Statements))
	case ast.KindCaseBlock:
		add(statements(parent.AsCaseBlock().Clauses))
	case ast.KindVariableDeclarationList:
		add(statements(parent.AsVariableDeclarationList().Declarations))
	}
	if isFunctionLikeForParams(parent) {
		if params := parent.Parameters(); len(params) > 0 {
			add(params)
		}
	}
	return out
}

// isFunctionLikeForParams gates the unconditional Parameters() call in
// listChildrenOf — Node.Parameters() panics for kinds that don't carry a
// parameter list (everything except function-like declarations).
func isFunctionLikeForParams(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindConstructor,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindFunctionType,
		ast.KindConstructorType,
		ast.KindCallSignature,
		ast.KindConstructSignature,
		ast.KindMethodSignature,
		ast.KindIndexSignature:
		return true
	}
	return false
}

func single(n *ast.Node) []*ast.Node {
	if n == nil {
		return nil
	}
	return []*ast.Node{n}
}

func statements(list *ast.NodeList) []*ast.Node {
	if list == nil {
		return nil
	}
	out := make([]*ast.Node, 0, len(list.Nodes))
	out = append(out, list.Nodes...)
	return out
}
