package test_framework

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// This file is the framework-neutral counterpart of eslint-plugin-jest's
// src/rules/utils/accessors.ts and its replaceAccessorFixer helper.
//
// It exists because ast.Node.Pos() is not ESTree's node.range[0]: Pos() points
// at the end of the previous token, so it includes the node's leading trivia
// (whitespace, newlines, comments). A fix range built straight from Pos() will
// silently overwrite that trivia. Report ranges are unaffected — ctx.ReportNode
// already routes through utils.TrimNodeTextRange — so only hand-built ranges
// need these helpers.

// IsAccessorNode reports whether node is one of the node kinds GetMemberEntries
// stores in MemberEntry.Node.
func IsAccessorNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindPrivateIdentifier,
		ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return true
	default:
		return false
	}
}

// AccessorRange returns the accessor's own source range, excluding leading
// trivia. For string and template literals the range includes the delimiters.
func AccessorRange(sourceFile *ast.SourceFile, node *ast.Node) (core.TextRange, bool) {
	if sourceFile == nil || !IsAccessorNode(node) {
		return core.TextRange{}, false
	}
	return utils.TrimNodeTextRange(sourceFile, node), true
}

// AccessorValueRange returns the range of the accessor's textual value: the
// token itself for an identifier, and the span between the delimiters for a
// string or template literal.
func AccessorValueRange(sourceFile *ast.SourceFile, node *ast.Node) (core.TextRange, bool) {
	tokenRange, ok := AccessorRange(sourceFile, node)
	if !ok {
		return core.TextRange{}, false
	}

	switch node.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return tokenRange, true
	default:
		// Both `'x'` and `` `x` `` use a single-character delimiter per side.
		// Guard against unterminated literals in recovered parses.
		if tokenRange.End()-tokenRange.Pos() < 2 {
			return core.TextRange{}, false
		}
		return core.NewTextRange(tokenRange.Pos()+1, tokenRange.End()-1), true
	}
}

// AccessorReplacement returns the range and text that rewrite an accessor to
// name. The accessor's written form is preserved: a bare identifier stays bare,
// and a string or template literal keeps its original delimiters because only
// the text between them is replaced.
//
// Preserving the delimiters is a deliberate divergence from upstream's
// replaceAccessorFixer, which always emits `'name'`. Normalizing quotes makes an
// autofix rewrite `["toEqual"]` into `['toStrictEqual']`, which introduces a
// fresh `quotes` violation in a double-quoted codebase. jest/no-alias-methods
// already preserved delimiters before this helper existed; this makes the whole
// plugin consistent with that.
//
// Returns false for accessors that cannot meaningfully be renamed — a private
// identifier is never a matcher or a framework API member.
func AccessorReplacement(sourceFile *ast.SourceFile, node *ast.Node, name string) (core.TextRange, string, bool) {
	if node == nil {
		return core.TextRange{}, "", false
	}

	switch node.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		valueRange, ok := AccessorValueRange(sourceFile, node)
		if !ok {
			return core.TextRange{}, "", false
		}
		return valueRange, name, true
	default:
		return core.TextRange{}, "", false
	}
}
