package unicornutil

import "github.com/microsoft/typescript-go/shim/ast"

// PlainParameterIdentifier returns the identifier declared by a parameter
// whose ESTree shape is a plain Identifier. Type annotations are allowed;
// rest and default parameters are not.
func PlainParameterIdentifier(parameter *ast.Node) *ast.Node {
	if !ast.IsParameterDeclaration(parameter) {
		return nil
	}
	declaration := parameter.AsParameterDeclaration()
	if declaration == nil || declaration.DotDotDotToken != nil ||
		declaration.Initializer != nil {
		return nil
	}
	name := declaration.Name()
	if name == nil || !ast.IsIdentifier(name) {
		return nil
	}
	return name
}

// IsSameIdentifier reports whether two expressions are identifiers with the
// same name. Parentheses are transparent because ESTree does not preserve
// them; TypeScript assertion wrappers deliberately remain significant.
func IsSameIdentifier(left *ast.Node, right *ast.Node) bool {
	left = ast.SkipParentheses(left)
	right = ast.SkipParentheses(right)
	return left != nil && right != nil &&
		ast.IsIdentifier(left) && ast.IsIdentifier(right) &&
		left.AsIdentifier().Text == right.AsIdentifier().Text
}
