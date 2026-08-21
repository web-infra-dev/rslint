package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// CoreFunctionHeadLocator mirrors core ESLint's getFunctionHeadLoc. The
// shared GetFunctionHeadLoc helper follows typescript-eslint's decorator
// behavior, so Loc restores the owning member's decorator-inclusive start
// required by core rules.
type CoreFunctionHeadLocator struct {
	sourceFile *ast.SourceFile
	tokens     *scanner.Scanner
}

// NewCoreFunctionHeadLocator returns a reusable locator for one source file.
func NewCoreFunctionHeadLocator(sourceFile *ast.SourceFile) *CoreFunctionHeadLocator {
	return &CoreFunctionHeadLocator{sourceFile: sourceFile}
}

// Loc returns the core ESLint report range for node's function head.
func (l *CoreFunctionHeadLocator) Loc(node *ast.Node) core.TextRange {
	// GetFunctionHeadLoc has to reproduce SourceCode.getTokenBefore() for a
	// single-parameter field arrow. Its general implementation scans from the
	// start of the source file, which becomes quadratic when a rule reports many
	// such fields. The owning field gives us a narrow, local token boundary.
	if loc, ok := l.singleParameterFieldArrowHeadLoc(node); ok {
		return loc
	}

	loc := GetFunctionHeadLoc(l.sourceFile, node)
	member := node
	switch {
	case isCoreFunctionMember(node):
	case node.Kind == ast.KindArrowFunction || node.Kind == ast.KindFunctionExpression:
		owner := ast.WalkUpParenthesizedExpressions(node.Parent)
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

func (l *CoreFunctionHeadLocator) singleParameterFieldArrowHeadLoc(node *ast.Node) (core.TextRange, bool) {
	if node.Kind != ast.KindArrowFunction {
		return core.TextRange{}, false
	}
	arrow := node.AsArrowFunction()
	if arrow.Parameters == nil || len(arrow.Parameters.Nodes) != 1 {
		return core.TextRange{}, false
	}

	owner := ast.WalkUpParenthesizedExpressions(node.Parent)
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

	parameterStart := TrimNodeTextRange(l.sourceFile, arrow.Parameters.Nodes[0]).Pos()
	scanStart := TrimNodeTextRange(l.sourceFile, node).Pos()
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		// A parenless arrow can itself be wrapped: `field = (value => {})`.
		// ESLint sees no wrapper node, so that `(` is the token before `value`.
		scanStart = TrimNodeTextRange(l.sourceFile, parent).Pos()
	}

	end := parameterStart
	var previousKind ast.Kind
	previousStart := 0
	foundPrevious := false
	if l.tokens == nil {
		l.tokens = scanner.NewScanner()
	}
	// Reuse one scanner per file: constructing one for every report would trade
	// the quadratic scan for avoidable per-diagnostic allocations.
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

func isCoreFunctionMember(node *ast.Node) bool {
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
