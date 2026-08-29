package unicornutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// CallExpressionParenthesesRange returns the actual opening and closing
// parenthesis tokens of a call. It starts after type arguments and after the
// final argument, so parentheses inside either cannot be mistaken for the
// call's delimiters.
func CallExpressionParenthesesRange(sourceFile *ast.SourceFile, node *ast.Node) (core.TextRange, core.TextRange, bool) {
	if sourceFile == nil || node == nil {
		return core.TextRange{}, core.TextRange{}, false
	}
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil {
		return core.TextRange{}, core.TextRange{}, false
	}
	// The parser records the argument-list span immediately after consuming
	// `(`, while the call node ends immediately after `)`. This is the common
	// non-recovery path and avoids allocating two scanners per call.
	if call.Arguments != nil {
		text := sourceFile.Text()
		openingPos, closingPos := call.Arguments.Pos()-1, node.End()-1
		if openingPos >= 0 && closingPos > openingPos && closingPos < len(text) &&
			text[openingPos] == '(' && text[closingPos] == ')' {
			return core.NewTextRange(openingPos, openingPos+1),
				core.NewTextRange(closingPos, closingPos+1), true
		}
	}

	scanStart := call.Expression.End()
	if call.TypeArguments != nil && len(call.TypeArguments.Nodes) > 0 {
		scanStart = call.TypeArguments.End()
	}
	s := scanner.GetScannerForSourceFile(sourceFile, scanStart)
	for s.Token() != ast.KindOpenParenToken && s.TokenStart() < node.End() {
		s.Scan()
	}
	if s.Token() != ast.KindOpenParenToken {
		return core.TextRange{}, core.TextRange{}, false
	}
	opening := s.TokenRange()

	suffixStart := opening.End()
	arguments := node.Arguments()
	if len(arguments) > 0 {
		suffixStart = arguments[len(arguments)-1].End()
	}
	s = scanner.GetScannerForSourceFile(sourceFile, suffixStart)
	if s.Token() == ast.KindCommaToken {
		s.Scan()
	}
	if s.Token() != ast.KindCloseParenToken || s.TokenEnd() != node.End() {
		return core.TextRange{}, core.TextRange{}, false
	}

	return opening, s.TokenRange(), true
}
