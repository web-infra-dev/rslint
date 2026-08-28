package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// FunctionHeadRangeLocator returns function-head report ranges matching
// ESLint's built-in rules. The shared GetFunctionHeadLoc helper follows
// typescript-eslint's decorator behavior, so Range restores the owning
// member's decorator-inclusive start where the two implementations differ.
type FunctionHeadRangeLocator struct {
	sourceFile *ast.SourceFile
	tokens     *scanner.Scanner
}

// NewFunctionHeadRangeLocator returns a reusable locator for one source file.
func NewFunctionHeadRangeLocator(sourceFile *ast.SourceFile) *FunctionHeadRangeLocator {
	return &FunctionHeadRangeLocator{sourceFile: sourceFile}
}

// Range returns the report range for node's function head.
func (l *FunctionHeadRangeLocator) Range(node *ast.Node) core.TextRange {
	// GetFunctionHeadLoc has to reproduce SourceCode.getTokenBefore() for a
	// single-parameter field arrow. Its general implementation lazily builds a
	// token index for the whole file; the owning field gives rules that only
	// need these report ranges a narrower, local token boundary.
	if loc, ok := l.singleParameterFieldArrowHeadRange(node); ok {
		return loc
	}

	loc := GetFunctionHeadLoc(l.sourceFile, node)
	member := node
	switch {
	case isFunctionHeadMember(node):
	case node.Kind == ast.KindArrowFunction || node.Kind == ast.KindFunctionExpression:
		owner := ESTreeParent(node)
		if owner == nil || owner.Kind != ast.KindPropertyDeclaration ||
			ast.HasSyntacticModifier(owner, ast.ModifierFlagsAccessor) {
			return loc
		}
		member = owner
	default:
		return loc
	}

	if !hasDecorator(member) {
		return loc
	}
	return core.NewTextRange(TrimNodeTextRange(l.sourceFile, member).Pos(), loc.End())
}

func (l *FunctionHeadRangeLocator) singleParameterFieldArrowHeadRange(node *ast.Node) (core.TextRange, bool) {
	if node.Kind != ast.KindArrowFunction {
		return core.TextRange{}, false
	}
	arrow := node.AsArrowFunction()
	if arrow.Parameters == nil || len(arrow.Parameters.Nodes) != 1 {
		return core.TextRange{}, false
	}

	owner := ESTreeParent(node)
	if owner == nil {
		return core.TextRange{}, false
	}
	switch owner.Kind {
	case ast.KindPropertyAssignment:
	case ast.KindPropertyDeclaration:
		if ast.HasSyntacticModifier(owner, ast.ModifierFlagsAccessor) {
			return core.TextRange{}, false
		}
	default:
		return core.TextRange{}, false
	}
	initializer := owner.Initializer()
	if initializer == nil ||
		(initializer != node && ESTreeRuntimeExpression(initializer) != node) {
		return core.TextRange{}, false
	}

	parameterStart := TrimNodeTextRange(l.sourceFile, arrow.Parameters.Nodes[0]).Pos()
	// Start at the owning property's initializer. This is the narrowest source
	// boundary that includes every ESTree-elided parenthesis and JSDoc cast
	// around the arrow, so the scan still observes a `(` immediately before a
	// single parameter without falling back to a file-wide token lookup.
	scanStart := TrimNodeTextRange(l.sourceFile, initializer).Pos()

	end := parameterStart
	var previousKind ast.Kind
	previousStart := 0
	foundPrevious := false
	if l.tokens == nil {
		l.tokens = scanner.NewScanner()
	}
	// Reuse one scanner per file to avoid per-diagnostic allocations.
	l.tokens.SetText(l.sourceFile.Text())
	l.tokens.SetLanguageVariant(l.sourceFile.LanguageVariant)
	l.tokens.ResetTokenState(scanStart)
	l.tokens.Scan()
	for l.tokens.Token() != ast.KindEndOfFile && l.tokens.TokenStart() < parameterStart {
		if l.tokens.TokenEnd() <= parameterStart {
			previousKind = l.tokens.Token()
			previousStart = l.tokens.TokenStart()
			foundPrevious = true
		}
		l.tokens.Scan()
	}
	if foundPrevious && previousKind == ast.KindOpenParenToken {
		end = previousStart
	}

	return core.NewTextRange(TrimNodeTextRange(l.sourceFile, owner).Pos(), end), true
}

func isFunctionHeadMember(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return true
	default:
		return false
	}
}

func hasDecorator(node *ast.Node) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindDecorator {
			return true
		}
	}
	return false
}
