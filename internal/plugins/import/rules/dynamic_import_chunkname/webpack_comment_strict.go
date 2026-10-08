package dynamic_import_chunkname

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// strictReservedWords cannot be used as identifiers in strict mode code.
var strictReservedWords = map[string]struct{}{
	"implements": {}, "interface": {}, "let": {}, "package": {}, "private": {},
	"protected": {}, "public": {}, "static": {}, "yield": {},
}

// containsStrictModeError reports strict-mode early errors inside class
// expressions. A class body is always strict code, and the TypeScript parser
// parses the comment as sloppy code, so these errors are not reported by it.
// Every node of a class is checked, including method bodies that never run,
// because an early error is a syntax error wherever it appears.
func containsStrictModeError(sourceFile *ast.SourceFile) bool {
	checker := &strictModeChecker{sourceFile: sourceFile}
	sourceFile.AsNode().ForEachChild(checker.visit)
	return checker.invalid
}

type strictModeChecker struct {
	sourceFile *ast.SourceFile
	classDepth int
	invalid    bool
}

func (c *strictModeChecker) visit(node *ast.Node) bool {
	if node.Kind == ast.KindClassExpression {
		c.classDepth++
		node.ForEachChild(c.visit)
		c.classDepth--
		return c.invalid
	}
	if c.classDepth > 0 && c.violates(node) {
		c.invalid = true
		return true
	}
	return node.ForEachChild(c.visit)
}

// violates reports whether node is a strict-mode early error on its own.
func (c *strictModeChecker) violates(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindDeleteExpression:
		// `delete name` is a SyntaxError in strict mode, wherever the name is.
		return ast.SkipParentheses(node.AsDeleteExpression().Expression).Kind == ast.KindIdentifier
	case ast.KindNumericLiteral:
		return node.AsNumericLiteral().TokenFlags&(ast.TokenFlagsOctal|ast.TokenFlagsContainsLeadingZero) != 0
	case ast.KindStringLiteral:
		raw := utils.TrimNodeTextRange(c.sourceFile, node)
		return hasLegacyOctalEscape(c.sourceFile.Text()[raw.Pos():raw.End()])
	case ast.KindIdentifier:
		name := node.Text()
		if _, reserved := strictReservedWords[name]; reserved && !isNameOfParent(node) {
			return true
		}
		return (name == "eval" || name == "arguments") && isAssignedOrBound(node)
	case ast.KindWithStatement:
		return true
	case ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return hasDuplicateParameter(node)
	}
	return false
}

// isNameOfParent reports whether node is the name of a property, method or
// member access, where it is a name rather than a reference.
func isNameOfParent(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression, ast.KindPropertyAssignment, ast.KindMethodDeclaration,
		ast.KindPropertyDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return parent.Name() == node
	}
	return false
}

// isAssignedOrBound reports whether node is the target of an assignment or
// update, or the name of a variable or parameter.
func isAssignedOrBound(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.Left == node && ast.IsAssignmentOperator(binary.OperatorToken.Kind)
	case ast.KindPrefixUnaryExpression:
		unary := parent.AsPrefixUnaryExpression()
		return unary.Operand == node && (unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken)
	case ast.KindPostfixUnaryExpression:
		unary := parent.AsPostfixUnaryExpression()
		return unary.Operand == node
	case ast.KindVariableDeclaration, ast.KindParameter, ast.KindBindingElement:
		return parent.Name() == node
	}
	return false
}

// hasDuplicateParameter reports two parameters that bind the same simple name.
// Binding patterns are not compared.
func hasDuplicateParameter(function *ast.Node) bool {
	if function.ParameterList() == nil {
		return false
	}
	seen := map[string]bool{}
	for _, parameter := range function.Parameters() {
		name := parameter.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			continue
		}
		if seen[name.Text()] {
			return true
		}
		seen[name.Text()] = true
	}
	return false
}

// hasLegacyOctalEscape reports a string literal's raw text containing an
// octal escape (`\1`, `\01`) or `\8`, `\9`. `\0` alone is allowed.
func hasLegacyOctalEscape(raw string) bool {
	for i := 0; i+1 < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		next := raw[i+1]
		switch {
		case next >= '1' && next <= '9':
			return true
		case next == '0' && i+2 < len(raw) && raw[i+2] >= '0' && raw[i+2] <= '9':
			return true
		}
		i++
	}
	return false
}
