package utils

import (
	"fmt"
	"iter"
	"slices"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

func TrimNodeTextRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	return scanner.GetRangeOfTokenAtPosition(sourceFile, node.Pos()).WithEnd(node.End())
}

func GetCommentsInRange(sourceFile *ast.SourceFile, inRange core.TextRange) iter.Seq[ast.CommentRange] {
	nodeFactory := ast.NewNodeFactory(ast.NodeFactoryHooks{})

	return func(yield func(ast.CommentRange) bool) {
		// Simple approach: get all comments from position 0 and filter
		// This is less efficient but more reliable than trying to optimize the start position
		seenComments := make(map[string]bool)

		// Get all leading comments from the beginning of the file
		for commentRange := range scanner.GetLeadingCommentRanges(nodeFactory, sourceFile.Text(), 0) {
			// Check if comment overlaps with our range (more flexible)
			if commentRange.Pos() < inRange.End() && commentRange.End() > inRange.Pos() {
				key := fmt.Sprintf("%d-%d", commentRange.Pos(), commentRange.End())
				if !seenComments[key] {
					seenComments[key] = true
					if !yield(commentRange) {
						return
					}
				}
			}
		}

		// Get all trailing comments from the beginning of the file
		for commentRange := range scanner.GetTrailingCommentRanges(nodeFactory, sourceFile.Text(), 0) {
			// Check if comment overlaps with our range (more flexible)
			if commentRange.Pos() < inRange.End() && commentRange.End() > inRange.Pos() {
				key := fmt.Sprintf("%d-%d", commentRange.Pos(), commentRange.End())
				if !seenComments[key] {
					seenComments[key] = true
					if !yield(commentRange) {
						return
					}
				}
			}
		}
	}
}

func HasCommentsInRange(sourceFile *ast.SourceFile, inRange core.TextRange) bool {
	// First try the scanner-based approach
	for range GetCommentsInRange(sourceFile, inRange) {
		return true
	}

	// Fallback: directly check the source text for comment patterns
	sourceText := sourceFile.Text()
	if inRange.Pos() >= 0 && inRange.End() <= len(sourceText) {
		rangeText := sourceText[inRange.Pos():inRange.End()]
		// Check for /* */ comments and // comments
		if containsBlockComment(rangeText) || containsLineComment(rangeText) {
			return true
		}
	}

	return false
}

func containsBlockComment(text string) bool {
	i := 0
	for i < len(text)-1 {
		if text[i] == '/' && text[i+1] == '*' {
			return true
		}
		i++
	}
	return false
}

func containsLineComment(text string) bool {
	i := 0
	for i < len(text)-1 {
		if text[i] == '/' && text[i+1] == '/' {
			return true
		}
		i++
	}
	return false
}

func TypeRecurser(t *checker.Type, predicate func(t *checker.Type) /* should stop */ bool) bool {
	if IsTypeFlagSet(t, checker.TypeFlagsUnionOrIntersection) {
		for _, subtype := range t.Types() {
			if TypeRecurser(subtype, predicate) {
				return true
			}
		}
		return false
	} else {
		return predicate(t)
	}
}

// SUPER DIRTY HACK FOR OPTIONAL FIELDS :(
func Ref[T any](a T) *T {
	return &a
}

func GetNumberIndexType(typeChecker *checker.Checker, t *checker.Type) *checker.Type {
	return checker.Checker_getIndexTypeOfType(typeChecker, t, checker.Checker_numberType(typeChecker))
}
func GetHeritageClauses(node *ast.Node) *ast.NodeList {
	switch node.Kind {
	case ast.KindClassDeclaration:
		return node.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		return node.AsClassExpression().HeritageClauses
	case ast.KindInterfaceDeclaration:
		return node.AsInterfaceDeclaration().HeritageClauses
	}
	return nil
}

// Source: typescript-go/internal/core/core.go
func Filter[T any](slice []T, f func(T) bool) []T {
	for i, value := range slice {
		if !f(value) {
			result := slices.Clone(slice[:i])
			for i++; i < len(slice); i++ {
				value = slice[i]
				if f(value) {
					result = append(result, value)
				}
			}
			return result
		}
	}
	return slice
}

// Source: typescript-go/internal/core/core.go
func FilterIndex[T any](slice []T, f func(T, int, []T) bool) []T {
	for i, value := range slice {
		if !f(value, i, slice) {
			result := slices.Clone(slice[:i])
			for i++; i < len(slice); i++ {
				value = slice[i]
				if f(value, i, slice) {
					result = append(result, value)
				}
			}
			return result
		}
	}
	return slice
}

// Source: typescript-go/internal/core/core.go
func Map[T, U any](slice []T, f func(T) U) []U {
	if len(slice) == 0 {
		return nil
	}
	result := make([]U, len(slice))
	for i, value := range slice {
		result[i] = f(value)
	}
	return result
}

// Source: typescript-go/internal/core/core.go
func Some[T any](slice []T, f func(T) bool) bool {
	for _, value := range slice {
		if f(value) {
			return true
		}
	}
	return false
}

// Source: typescript-go/internal/core/core.go
func Every[T any](slice []T, f func(T) bool) bool {
	for _, value := range slice {
		if !f(value) {
			return false
		}
	}
	return true
}

// Source: typescript-go/internal/core/core.go
func Flatten[T any](array [][]T) []T {
	var result []T
	for _, subArray := range array {
		result = append(result, subArray...)
	}
	return result
}

func IncludesModifier(node interface{ Modifiers() *ast.ModifierList }, modifier ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	return Some(modifiers.NodeList.Nodes, func(m *ast.Node) bool {
		return m.Kind == modifier
	})
}

// Source: https://github.com/microsoft/typescript-go/blob/5652e65d5ae944375676d3955f9755e554576d41/internal/jsnum/string.go#L99
func IsStrWhiteSpace(r rune) bool {
	// This is different than stringutil.IsWhiteSpaceLike.

	// https://tc39.es/ecma262/2024/multipage/ecmascript-language-lexical-grammar.html#prod-LineTerminator
	// https://tc39.es/ecma262/2024/multipage/ecmascript-language-lexical-grammar.html#prod-WhiteSpace

	switch r {
	// LineTerminator
	case '\n', '\r', 0x2028, 0x2029:
		return true
	// WhiteSpace
	case '\t', '\v', '\f', 0xFEFF:
		return true
	}

	// WhiteSpace
	return unicode.Is(unicode.Zs, r)
}

// GetFunctionHeadLoc returns the location of a function's "head" - the part before the body.
// This includes the function keyword, name, and parameters, but excludes the function body.
// Used for precise error reporting in ESLint rules.
func GetFunctionHeadLoc(node *ast.Node, sourceFile *ast.SourceFile) core.TextRange {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		start := node.Pos()

		// Find the end position after the parameter list
		if len(funcDecl.Parameters.Nodes) > 0 {
			lastParam := funcDecl.Parameters.Nodes[len(funcDecl.Parameters.Nodes)-1]
			// Look for closing parenthesis after the last parameter
			s := scanner.GetScannerForSourceFile(sourceFile, lastParam.End())
			for s.Token() != ast.KindCloseParenToken && s.Token() != ast.KindEndOfFile {
				s.Scan()
			}
			if s.Token() == ast.KindCloseParenToken {
				return core.NewTextRange(start, s.TokenEnd())
			}
		}

		// Fallback: find the opening brace and go back
		return findFunctionHeadEnd(sourceFile, start, funcDecl.Body)

	case ast.KindFunctionExpression:
		funcExpr := node.AsFunctionExpression()
		start := node.Pos()

		if len(funcExpr.Parameters.Nodes) > 0 {
			lastParam := funcExpr.Parameters.Nodes[len(funcExpr.Parameters.Nodes)-1]
			s := scanner.GetScannerForSourceFile(sourceFile, lastParam.End())
			for s.Token() != ast.KindCloseParenToken && s.Token() != ast.KindEndOfFile {
				s.Scan()
			}
			if s.Token() == ast.KindCloseParenToken {
				return core.NewTextRange(start, s.TokenEnd())
			}
		}

		return findFunctionHeadEnd(sourceFile, start, funcExpr.Body)

	case ast.KindArrowFunction:
		arrowFunc := node.AsArrowFunction()
		start := node.Pos()

		// For arrow functions, we need to find the '=>' token
		searchStart := start
		if len(arrowFunc.Parameters.Nodes) > 0 {
			lastParam := arrowFunc.Parameters.Nodes[len(arrowFunc.Parameters.Nodes)-1]
			searchStart = lastParam.End()
		}

		// Find the '=>' token
		s := scanner.GetScannerForSourceFile(sourceFile, searchStart)
		for s.Token() != ast.KindEqualsGreaterThanToken && s.Token() != ast.KindEndOfFile {
			s.Scan()
		}
		if s.Token() == ast.KindEqualsGreaterThanToken {
			return core.NewTextRange(start, s.TokenEnd())
		}

		// Fallback
		return findFunctionHeadEnd(sourceFile, start, arrowFunc.Body)

	case ast.KindMethodDeclaration:
		methodDecl := node.AsMethodDeclaration()
		start := node.Pos()

		if len(methodDecl.Parameters.Nodes) > 0 {
			lastParam := methodDecl.Parameters.Nodes[len(methodDecl.Parameters.Nodes)-1]
			s := scanner.GetScannerForSourceFile(sourceFile, lastParam.End())
			for s.Token() != ast.KindCloseParenToken && s.Token() != ast.KindEndOfFile {
				s.Scan()
			}
			if s.Token() == ast.KindCloseParenToken {
				return core.NewTextRange(start, s.TokenEnd())
			}
		}

		return findFunctionHeadEnd(sourceFile, start, methodDecl.Body)

	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		start := node.Pos()

		if len(accessor.Parameters.Nodes) > 0 {
			lastParam := accessor.Parameters.Nodes[len(accessor.Parameters.Nodes)-1]
			s := scanner.GetScannerForSourceFile(sourceFile, lastParam.End())
			for s.Token() != ast.KindCloseParenToken && s.Token() != ast.KindEndOfFile {
				s.Scan()
			}
			if s.Token() == ast.KindCloseParenToken {
				return core.NewTextRange(start, s.TokenEnd())
			}
		}

		return findFunctionHeadEnd(sourceFile, start, accessor.Body)

	case ast.KindSetAccessor:
		accessor := node.AsSetAccessorDeclaration()
		start := node.Pos()

		if len(accessor.Parameters.Nodes) > 0 {
			lastParam := accessor.Parameters.Nodes[len(accessor.Parameters.Nodes)-1]
			s := scanner.GetScannerForSourceFile(sourceFile, lastParam.End())
			for s.Token() != ast.KindCloseParenToken && s.Token() != ast.KindEndOfFile {
				s.Scan()
			}
			if s.Token() == ast.KindCloseParenToken {
				return core.NewTextRange(start, s.TokenEnd())
			}
		}

		return findFunctionHeadEnd(sourceFile, start, accessor.Body)

	default:
		// Fallback for unknown function types
		return TrimNodeTextRange(sourceFile, node)
	}
}

// Helper function to find the end of a function head by looking for the opening brace
func findFunctionHeadEnd(sourceFile *ast.SourceFile, start int, body *ast.Node) core.TextRange {
	if body == nil {
		// No body, use the entire node
		return core.NewTextRange(start, start)
	}

	// Find the opening brace of the function body
	s := scanner.GetScannerForSourceFile(sourceFile, body.Pos())
	for s.Token() != ast.KindOpenBraceToken && s.Token() != ast.KindEndOfFile && s.TokenStart() >= body.Pos() {
		s.Scan()
	}

	if s.Token() == ast.KindOpenBraceToken {
		// Go back to find the last non-whitespace token before the brace
		end := s.TokenStart()
		return core.NewTextRange(start, end)
	}

	// Fallback: use the start of the body
	return core.NewTextRange(start, body.Pos())
}
