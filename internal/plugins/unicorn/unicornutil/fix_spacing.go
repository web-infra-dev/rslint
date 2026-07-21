package unicornutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func isProblematicKeywordToken(token utils.SourceToken) bool {
	if token.Text == "of" || token.Text == "await" {
		return true
	}
	// @typescript-eslint/parser exposes only ECMAScript keywords through
	// ESLint's `Keyword` token type. TypeScript-only contextual keywords such
	// as `as` and `satisfies` remain identifiers and need no inserted space.
	if token.Kind < ast.KindFirstKeyword ||
		token.Kind > ast.KindLastFutureReservedWord ||
		token.Text == "" {
		return false
	}
	for _, character := range token.Text {
		if character < 'a' || character > 'z' {
			return false
		}
	}
	return true
}

// SpaceAroundKeywordFixes mirrors Unicorn's fixSpaceAroundKeyword helper. It
// adds spaces only when the complete parenthesized range directly touches a
// lowercase keyword (plus the contextual `of` and `await` identifier tokens).
func SpaceAroundKeywordFixes(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	if sourceFile == nil || node == nil {
		return nil
	}

	outer := utils.OutermostParenthesizedExpression(node)
	textRange := utils.TrimNodeTextRange(sourceFile, outer)
	var fixes []rule.RuleFix

	if before, ok := utils.TokenBeforePosition(sourceFile, textRange.Pos()); ok &&
		before.End == textRange.Pos() && isProblematicKeywordToken(before) {
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(textRange.Pos(), textRange.Pos()),
			" ",
		))
	}

	if after, ok := utils.TokenAtOrAfter(sourceFile, textRange.End()); ok &&
		after.Start == textRange.End() && isProblematicKeywordToken(after) {
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(textRange.End(), textRange.End()),
			" ",
		))
	}
	return fixes
}
