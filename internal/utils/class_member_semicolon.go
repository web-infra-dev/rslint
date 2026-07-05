package utils

import "github.com/microsoft/typescript-go/shim/ast"

// ClassMemberLeadingSemicolonOptions controls how
// NeedsClassMemberLeadingSemicolon treats class fields without initializers.
type ClassMemberLeadingSemicolonOptions struct {
	// IncludePropertiesWithoutInitializers also treats plain fields like `foo`
	// as hazards. Type-only fields like `foo: string` are always considered
	// because the trailing type can merge with a following computed member.
	IncludePropertiesWithoutInitializers bool
}

// NeedsClassMemberLeadingSemicolon reports whether an edit that removes or
// rewrites member would leave nextToken at the start of a class member in a
// position where the previous property declaration could consume it as part of
// its initializer or type.
func NeedsClassMemberLeadingSemicolon(
	sourceFile *ast.SourceFile,
	classNode *ast.Node,
	member *ast.Node,
	nextToken SourceToken,
	options ClassMemberLeadingSemicolonOptions,
) bool {
	if sourceFile == nil || classNode == nil || member == nil {
		return false
	}
	if classNode.Kind != ast.KindClassDeclaration && classNode.Kind != ast.KindClassExpression {
		return false
	}
	if !canClassMemberTokenContinueExpression(nextToken) {
		return false
	}

	members := classNode.Members()
	idx := -1
	for i, m := range members {
		if m == member {
			idx = i
			break
		}
	}
	if idx <= 0 {
		return false
	}

	prev := members[idx-1]
	if !ast.IsPropertyDeclaration(prev) || classPropertyEndsWithSemicolon(sourceFile, prev) {
		return false
	}

	prop := prev.AsPropertyDeclaration()
	if prop == nil {
		return false
	}
	if prop.Initializer == nil {
		return options.IncludePropertiesWithoutInitializers || prop.Type != nil
	}

	init := prop.Initializer
	// Postfix ++/-- are restricted productions, so ASI fires before the next
	// class member token and no explicit semicolon is needed.
	if init.Kind == ast.KindPostfixUnaryExpression {
		return false
	}
	// Arrow functions with block bodies terminate at their own `}`; a following
	// `[`/`in`/`instanceof`/`*` cannot become a member access on the initializer.
	if init.Kind == ast.KindArrowFunction {
		if body := init.Body(); body != nil && body.Kind == ast.KindBlock {
			return false
		}
	}
	return true
}

func canClassMemberTokenContinueExpression(token SourceToken) bool {
	switch token.Kind {
	case ast.KindOpenBracketToken, ast.KindAsteriskToken, ast.KindInKeyword, ast.KindInstanceOfKeyword:
		return true
	case ast.KindIdentifier:
		return token.Text == "in" || token.Text == "instanceof"
	default:
		return false
	}
}

func classPropertyEndsWithSemicolon(sourceFile *ast.SourceFile, node *ast.Node) bool {
	text := sourceFile.Text()
	end := SkipTrailingWhitespace(text, node.Pos(), node.End())
	return end > node.Pos() && end <= len(text) && text[end-1] == ';'
}
