package brace_style

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/stylisticutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const (
	msgNextLineOpen    = "nextLineOpen"
	msgSameLineOpen    = "sameLineOpen"
	msgBlockSameLine   = "blockSameLine"
	msgNextLineClose   = "nextLineClose"
	msgSingleLineClose = "singleLineClose"
	msgSameLineClose   = "sameLineClose"

	descNextLineOpen    = "Opening curly brace does not appear on the same line as controlling statement."
	descSameLineOpen    = "Opening curly brace appears on the same line as controlling statement."
	descBlockSameLine   = "Statement inside of curly braces should be on next line."
	descNextLineClose   = "Closing curly brace does not appear on the same line as the subsequent block."
	descSingleLineClose = "Closing curly brace should be on the same line as opening curly brace or on the line after the previous block."
	descSameLineClose   = "Closing curly brace appears on the same line as the subsequent block."
)

const (
	style1tbs       = "1tbs"
	styleStroustrup = "stroustrup"
	styleAllman     = "allman"
)

type options struct {
	style           string
	allowSingleLine bool
}

// parseOptions mirrors upstream's option layout.
//
//	rule: ['brace-style']                                  → '1tbs', allowSingleLine=false
//	rule: ['brace-style', '1tbs']                          → '1tbs', allowSingleLine=false
//	rule: ['brace-style', 'stroustrup']                    → 'stroustrup', allowSingleLine=false
//	rule: ['brace-style', 'allman']                        → 'allman', allowSingleLine=false
//	rule: ['brace-style', <style>, { allowSingleLine: b }] → <style>, allowSingleLine=b
//
// rslint's config loader collapses a single-element options array, so the
// first slot might travel as a bare string. We accept both shapes here.
// `GetOptionsMap` is not used because the primary option is a string, not an
// object — the standard helper would discard it.
func parseOptions(raw any) options {
	opts := options{style: style1tbs}

	var arr []any
	switch v := raw.(type) {
	case []any:
		arr = v
	case string:
		arr = []any{v}
	}

	if len(arr) > 0 {
		if s, ok := arr[0].(string); ok {
			switch s {
			case style1tbs, styleStroustrup, styleAllman:
				opts.style = s
			}
		}
	}
	if len(arr) > 1 {
		if m, ok := arr[1].(map[string]any); ok {
			if b, ok := m["allowSingleLine"].(bool); ok {
				opts.allowSingleLine = b
			}
		}
	}
	return opts
}

// curlyPair is the resolved (`{`, `}`) pair for one container. `searchStart`
// is a position guaranteed to precede the opening curly — used to locate
// `tokenBeforeOpeningCurly` via a forward scan.
type curlyPair struct {
	openPos     int
	closePos    int
	searchStart int
}

// shouldSkipBlock returns true when a Block node lives in a parent position
// where upstream's `BlockStatement` listener would NOT fire — i.e. its parent
// type is in STATEMENT_LIST_PARENTS (`Program`, `BlockStatement`,
// `StaticBlock`, `SwitchCase`). In ESTree, `StaticBlock` is its own node and
// has no nested BlockStatement, so we also skip Blocks whose parent is a
// `ClassStaticBlockDeclaration` — that block IS the static block in ESTree,
// and gets handled by the dedicated `KindClassStaticBlockDeclaration` listener.
func shouldSkipBlock(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return true
	}
	switch parent.Kind {
	case ast.KindSourceFile,
		ast.KindBlock,
		ast.KindCaseClause,
		ast.KindDefaultClause,
		ast.KindClassStaticBlockDeclaration:
		return true
	}
	return false
}

// validateOpen runs the three "opening side" checks on a curly pair:
//
//	nextLineOpen   — non-allman: `{` must be on same line as preceding token
//	sameLineOpen   — allman: `{` must NOT be on same line as preceding token
//	blockSameLine  — `{` must NOT be on same line as the first inner token
//
// `allowSingleLine` toggles a per-pair exception on the last two checks; the
// first runs unconditionally because it's about the OPENING brace's position
// relative to what comes BEFORE it (single-line-block exception doesn't apply).
func validateOpen(
	ctx rule.RuleContext,
	text string,
	opts options,
	pair curlyPair,
) {
	sf := ctx.SourceFile
	openPos := pair.openPos
	closePos := pair.closePos
	openEnd := openPos + 1

	isAllman := opts.style == styleAllman

	// tokenAfterOpeningCurly: first token start at or after `{` + 1.
	afterStart := scanner.SkipTrivia(text, openEnd)
	if afterStart > closePos {
		afterStart = closePos
	}

	// tokenBeforeOpeningCurly via forward scan from the caller-supplied
	// position. If we can't locate it (parser recovery / synthetic ranges),
	// fall back to assuming same-line — that's the safer default since the
	// nextLineOpen / sameLineOpen reports would otherwise fire spuriously.
	prevOpenEnd, hasPrevOpen := stylisticutil.FindPrevTokenEnd(sf, pair.searchStart, openPos)
	if !hasPrevOpen {
		prevOpenEnd = openPos
	}

	singleLineException := opts.allowSingleLine && stylisticutil.SameLineByPos(sf, openPos, closePos)

	if !isAllman && !stylisticutil.SameLineByPos(sf, prevOpenEnd, openPos) {
		var fixes []rule.RuleFix
		if !stylisticutil.CommentsExistBetween(text, prevOpenEnd, openPos) {
			fixes = []rule.RuleFix{{
				Text:  " ",
				Range: core.NewTextRange(prevOpenEnd, openPos),
			}}
		}
		ctx.ReportRangeWithFixes(
			core.NewTextRange(openPos, openEnd),
			rule.RuleMessage{Id: msgNextLineOpen, Description: descNextLineOpen},
			fixes...,
		)
	}

	if isAllman && stylisticutil.SameLineByPos(sf, prevOpenEnd, openPos) && !singleLineException {
		ctx.ReportRangeWithFixes(
			core.NewTextRange(openPos, openEnd),
			rule.RuleMessage{Id: msgSameLineOpen, Description: descSameLineOpen},
			rule.RuleFix{
				Text:  "\n",
				Range: core.NewTextRange(openPos, openPos),
			},
		)
	}

	nonEmpty := afterStart < closePos
	if nonEmpty && stylisticutil.SameLineByPos(sf, openPos, afterStart) && !singleLineException {
		ctx.ReportRangeWithFixes(
			core.NewTextRange(openPos, openEnd),
			rule.RuleMessage{Id: msgBlockSameLine, Description: descBlockSameLine},
			rule.RuleFix{
				Text:  "\n",
				Range: core.NewTextRange(openEnd, openEnd),
			},
		)
	}
}

// validateClose runs the closing-side `singleLineClose` check. Allman /
// stroustrup / 1tbs all share this check; only `allowSingleLine` and the
// emptiness of the block can suppress it.
func validateClose(
	ctx rule.RuleContext,
	text string,
	opts options,
	pair curlyPair,
) {
	sf := ctx.SourceFile
	openPos := pair.openPos
	closePos := pair.closePos
	openEnd := openPos + 1
	closeEnd := closePos + 1

	// tokenAfterOpeningCurly: first token start at or after `{` + 1.
	afterStart := scanner.SkipTrivia(text, openEnd)
	if afterStart > closePos {
		afterStart = closePos
	}

	prevCloseEnd, ok := stylisticutil.FindPrevTokenEnd(sf, openEnd, closePos)
	if !ok {
		prevCloseEnd = openEnd
	}

	singleLineException := opts.allowSingleLine && stylisticutil.SameLineByPos(sf, openPos, closePos)
	nonEmpty := afterStart < closePos

	if nonEmpty && stylisticutil.SameLineByPos(sf, prevCloseEnd, closePos) && !singleLineException {
		ctx.ReportRangeWithFixes(
			core.NewTextRange(closePos, closeEnd),
			rule.RuleMessage{Id: msgSingleLineClose, Description: descSingleLineClose},
			rule.RuleFix{
				Text:  "\n",
				Range: core.NewTextRange(closePos, closePos),
			},
		)
	}
}

// validateCurlyBeforeKeyword checks the relationship between a closing `}`
// and the subsequent keyword (`else`, `catch`, `finally`):
//
//	nextLineClose   — 1tbs: `}` must be on same line as the keyword
//	sameLineClose   — stroustrup/allman: `}` must NOT be on same line as the keyword
func validateCurlyBeforeKeyword(
	ctx rule.RuleContext,
	text string,
	opts options,
	closePos int,
) {
	sf := ctx.SourceFile
	closeEnd := closePos + 1
	keywordStart := scanner.SkipTrivia(text, closeEnd)
	if keywordStart >= len(text) {
		return
	}

	is1tbs := opts.style == style1tbs

	if is1tbs && !stylisticutil.SameLineByPos(sf, closePos, keywordStart) {
		var fixes []rule.RuleFix
		if !stylisticutil.CommentsExistBetween(text, closeEnd, keywordStart) {
			fixes = []rule.RuleFix{{
				Text:  " ",
				Range: core.NewTextRange(closeEnd, keywordStart),
			}}
		}
		ctx.ReportRangeWithFixes(
			core.NewTextRange(closePos, closeEnd),
			rule.RuleMessage{Id: msgNextLineClose, Description: descNextLineClose},
			fixes...,
		)
		return
	}

	if !is1tbs && stylisticutil.SameLineByPos(sf, closePos, keywordStart) {
		ctx.ReportRangeWithFixes(
			core.NewTextRange(closePos, closeEnd),
			rule.RuleMessage{Id: msgSameLineClose, Description: descSameLineClose},
			rule.RuleFix{
				Text:  "\n",
				Range: core.NewTextRange(closeEnd, closeEnd),
			},
		)
	}
}

// blockCurlyPair locates `{` and `}` of a Block-like node whose text range
// already starts at `{` and ends at `}` + 1. Used by KindBlock, KindModuleBlock,
// and the inner Body of KindClassStaticBlockDeclaration.
func blockCurlyPair(sf *ast.SourceFile, node *ast.Node, searchStart int) (curlyPair, bool) {
	text := sf.Text()
	openPos := scanner.SkipTrivia(text, node.Pos())
	closePos := node.End() - 1
	if openPos >= closePos || closePos >= len(text) {
		return curlyPair{}, false
	}
	if text[openPos] != '{' || text[closePos] != '}' {
		return curlyPair{}, false
	}
	return curlyPair{openPos: openPos, closePos: closePos, searchStart: searchStart}, true
}

// classBodyCurlyPair locates `{` and `}` of a class body. tsgo doesn't model
// `ClassBody` as a separate node — the class members live directly on
// `ClassDeclaration` / `ClassExpression`. The opening curly is at
// `Members.Pos() - 1`; the closing curly is at `node.End() - 1`.
func classBodyCurlyPair(sf *ast.SourceFile, node *ast.Node, members *ast.NodeList) (curlyPair, bool) {
	text := sf.Text()
	if members == nil {
		return curlyPair{}, false
	}
	openPos := members.Pos() - 1
	closePos := node.End() - 1
	if openPos < node.Pos() || openPos >= closePos || closePos >= len(text) {
		return curlyPair{}, false
	}
	if text[openPos] != '{' || text[closePos] != '}' {
		return curlyPair{}, false
	}
	return curlyPair{openPos: openPos, closePos: closePos, searchStart: node.Pos()}, true
}

// switchCurlyPair locates `{` and `}` around a switch's CaseBlock. The
// CaseBlock is itself the `{ cases... }` container, so its `Pos()` (after
// trivia) lands on `{` and `End() - 1` lands on `}`.
func switchCurlyPair(sf *ast.SourceFile, switchNode *ast.Node) (curlyPair, bool) {
	ss := switchNode.AsSwitchStatement()
	if ss == nil || ss.CaseBlock == nil {
		return curlyPair{}, false
	}
	return blockCurlyPair(sf, ss.CaseBlock, switchNode.Pos())
}

// closeBeforeKeywordTrigger inspects a Block being exited and reports
// `nextLineClose` / `sameLineClose` when the block is the consequent of an
// IfStatement with an alternate, the try-block / catch-body of a TryStatement
// whose next sibling is a keyword.
//
// Centralizing the trigger here (rather than in IfStatement / TryStatement
// enter listeners) preserves source-order output: the close-before-keyword
// report is emitted right after the block's own `singleLineClose` (same
// position) and before the next block's open-side reports.
func closeBeforeKeywordTrigger(
	ctx rule.RuleContext,
	text string,
	opts options,
	block *ast.Node,
) {
	parent := block.Parent
	if parent == nil {
		return
	}
	closePos := block.End() - 1

	switch parent.Kind {
	case ast.KindIfStatement:
		is := parent.AsIfStatement()
		if is == nil {
			return
		}
		if is.ThenStatement != block || is.ElseStatement == nil {
			return
		}
		validateCurlyBeforeKeyword(ctx, text, opts, closePos)
	case ast.KindTryStatement:
		ts := parent.AsTryStatement()
		if ts == nil {
			return
		}
		// `try { ... } catch (...) { ... } [finally { ... }]` — the close
		// keyword check fires for the try block (next token: `catch` or
		// `finally`) unconditionally.
		if ts.TryBlock == block {
			validateCurlyBeforeKeyword(ctx, text, opts, closePos)
		}
	case ast.KindCatchClause:
		cc := parent.AsCatchClause()
		if cc == nil || cc.Block != block {
			return
		}
		grand := parent.Parent
		if grand == nil || grand.Kind != ast.KindTryStatement {
			return
		}
		ts := grand.AsTryStatement()
		if ts == nil || ts.CatchClause != parent || ts.FinallyBlock == nil {
			return
		}
		validateCurlyBeforeKeyword(ctx, text, opts, closePos)
	}
}

// BraceStyleRule enforces consistent brace style across block-bearing
// constructs (functions, classes, control-flow blocks, switch, try/catch,
// TS namespaces). Ported from @stylistic/eslint-plugin's brace-style.
//
// The rule is layout-only — it inserts and removes line breaks between
// braces and adjacent tokens, never reflows whole statements. As upstream
// notes for static blocks: when the autofix introduces a newline the
// resulting indentation may be wrong; that's expected, the `indent` rule
// handles indentation in a separate pass.
//
// Enter / exit listener split: opening-side reports (nextLineOpen,
// sameLineOpen, blockSameLine) fire on enter; closing-side reports
// (singleLineClose) plus close-before-keyword (nextLineClose, sameLineClose)
// fire on exit. This yields source-order diagnostics for nested constructs
// and for the if/else and try/catch/finally cases where reports on adjacent
// tokens come from different AST nodes.
var BraceStyleRule = rule.Rule{
	Name: "@stylistic/brace-style",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		text := ctx.SourceFile.Text()

		open := func(pair curlyPair) {
			validateOpen(ctx, text, opts, pair)
		}
		closeFn := func(pair curlyPair) {
			validateClose(ctx, text, opts, pair)
		}

		// Block enter/exit.
		blockEnter := func(node *ast.Node) {
			if shouldSkipBlock(node) {
				return
			}
			searchStart := node.Parent.Pos()
			if pair, ok := blockCurlyPair(ctx.SourceFile, node, searchStart); ok {
				open(pair)
			}
		}
		blockExit := func(node *ast.Node) {
			if shouldSkipBlock(node) {
				return
			}
			// Emit close-before-keyword BEFORE the block's own singleLineClose
			// when both fire at the same `}` position. This matches ESLint's
			// "parent-listener fires before child-listener" emission order
			// (IfStatement / TryStatement fire before the consequent's
			// BlockStatement). The two reports are at the same column, so any
			// downstream consumer that sorts by location sees identical order
			// in both implementations.
			closeBeforeKeywordTrigger(ctx, text, opts, node)
			searchStart := node.Parent.Pos()
			if pair, ok := blockCurlyPair(ctx.SourceFile, node, searchStart); ok {
				closeFn(pair)
			}
		}

		// ClassStaticBlockDeclaration enter/exit.
		staticEnter := func(node *ast.Node) {
			decl := node.AsClassStaticBlockDeclaration()
			if decl == nil || decl.Body == nil {
				return
			}
			if pair, ok := blockCurlyPair(ctx.SourceFile, decl.Body, node.Pos()); ok {
				open(pair)
			}
		}
		staticExit := func(node *ast.Node) {
			decl := node.AsClassStaticBlockDeclaration()
			if decl == nil || decl.Body == nil {
				return
			}
			if pair, ok := blockCurlyPair(ctx.SourceFile, decl.Body, node.Pos()); ok {
				closeFn(pair)
			}
		}

		// ClassDeclaration / ClassExpression enter/exit.
		classDeclEnter := func(node *ast.Node) {
			cd := node.AsClassDeclaration()
			if cd == nil {
				return
			}
			if pair, ok := classBodyCurlyPair(ctx.SourceFile, node, cd.Members); ok {
				open(pair)
			}
		}
		classDeclExit := func(node *ast.Node) {
			cd := node.AsClassDeclaration()
			if cd == nil {
				return
			}
			if pair, ok := classBodyCurlyPair(ctx.SourceFile, node, cd.Members); ok {
				closeFn(pair)
			}
		}
		classExprEnter := func(node *ast.Node) {
			ce := node.AsClassExpression()
			if ce == nil {
				return
			}
			if pair, ok := classBodyCurlyPair(ctx.SourceFile, node, ce.Members); ok {
				open(pair)
			}
		}
		classExprExit := func(node *ast.Node) {
			ce := node.AsClassExpression()
			if ce == nil {
				return
			}
			if pair, ok := classBodyCurlyPair(ctx.SourceFile, node, ce.Members); ok {
				closeFn(pair)
			}
		}

		// SwitchStatement enter/exit (validate on the CaseBlock's braces).
		switchEnter := func(node *ast.Node) {
			if pair, ok := switchCurlyPair(ctx.SourceFile, node); ok {
				open(pair)
			}
		}
		switchExit := func(node *ast.Node) {
			if pair, ok := switchCurlyPair(ctx.SourceFile, node); ok {
				closeFn(pair)
			}
		}

		// ModuleBlock enter/exit — `namespace Foo { ... }` / `module "Foo" { ... }`.
		moduleEnter := func(node *ast.Node) {
			searchStart := node.Pos()
			if node.Parent != nil {
				searchStart = node.Parent.Pos()
			}
			if pair, ok := blockCurlyPair(ctx.SourceFile, node, searchStart); ok {
				open(pair)
			}
		}
		moduleExit := func(node *ast.Node) {
			searchStart := node.Pos()
			if node.Parent != nil {
				searchStart = node.Parent.Pos()
			}
			if pair, ok := blockCurlyPair(ctx.SourceFile, node, searchStart); ok {
				closeFn(pair)
			}
		}

		return rule.RuleListeners{
			ast.KindBlock:                                              blockEnter,
			rule.ListenerOnExit(ast.KindBlock):                         blockExit,
			ast.KindClassStaticBlockDeclaration:                        staticEnter,
			rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration):   staticExit,
			ast.KindClassDeclaration:                                   classDeclEnter,
			rule.ListenerOnExit(ast.KindClassDeclaration):              classDeclExit,
			ast.KindClassExpression:                                    classExprEnter,
			rule.ListenerOnExit(ast.KindClassExpression):               classExprExit,
			ast.KindSwitchStatement:                                    switchEnter,
			rule.ListenerOnExit(ast.KindSwitchStatement):               switchExit,
			ast.KindModuleBlock:                                        moduleEnter,
			rule.ListenerOnExit(ast.KindModuleBlock):                   moduleExit,
		}
	},
}
