package no_array_constructor

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func preferLiteralMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferLiteral",
		Description: "The array literal notation [] is preferable.",
	}
}

func useLiteralMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteral",
		Description: "Replace with an array literal.",
	}
}

func useLiteralAfterSemicolonMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteralAfterSemicolon",
		Description: "Replace with an array literal, add preceding semicolon.",
	}
}

// calleeArgsAndTypeArgs extracts the callee, argument list, and type-argument
// list shared by CallExpression and NewExpression, the two forms this rule
// inspects.
func calleeArgsAndTypeArgs(node *ast.Node) (callee *ast.Node, args *ast.NodeList, typeArgs *ast.NodeList) {
	switch node.Kind {
	case ast.KindCallExpression:
		callExpr := node.AsCallExpression()
		return callExpr.Expression, callExpr.Arguments, callExpr.TypeArguments
	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		return newExpr.Expression, newExpr.Arguments, newExpr.TypeArguments
	}
	return nil, nil, nil
}

// findOpenParen scans forward from calleeEnd (the position right after the
// possibly-parenthesized callee) for the call's own opening paren, stopping
// at nodeEnd. Only `?.` punctuation and trivia can appear before it, since
// type arguments were already excluded by the caller. found is false when the
// node has no parentheses at all, e.g. a bare `new Array`.
func findOpenParen(sourceFile *ast.SourceFile, calleeEnd int, nodeEnd int) (pos int, found bool) {
	text := sourceFile.Text()
	if calleeEnd < 0 || nodeEnd < calleeEnd || nodeEnd > len(text) {
		return 0, false
	}
	p := calleeEnd
	for p < nodeEnd {
		p = scanner.SkipTrivia(text, p)
		if p >= nodeEnd {
			break
		}
		if text[p] == '(' {
			return p, true
		}
		p++
	}
	return 0, false
}

func countNonSpread(args *ast.NodeList) int {
	if args == nil {
		return 0
	}
	count := 0
	for _, arg := range args.Nodes {
		if arg.Kind != ast.KindSpreadElement {
			count++
		}
	}
	return count
}

func diagnosticTouchesRange(diagnostics []*ast.Diagnostic, textRange core.TextRange) bool {
	for _, diagnostic := range diagnostics {
		// Missing-token diagnostics are zero-width and can sit exactly at the
		// recovered call's end, so intersection is intentionally inclusive.
		if diagnostic != nil && diagnostic.Pos() <= textRange.End() && diagnostic.End() >= textRange.Pos() {
			return true
		}
	}
	return false
}

type arrayConstructorEditPlan struct {
	sourceFile  *ast.SourceFile
	comments    *rule.CommentStore
	node        *ast.Node
	callee      *ast.Node
	args        *ast.NodeList
	reportRange core.TextRange

	classified    bool
	editable      bool
	shouldSuggest bool
	argsStart     int
	argsEnd       int
}

func (p *arrayConstructorEditPlan) classify() {
	if p.classified {
		return
	}
	p.classified = true

	sourceText := p.sourceFile.Text()
	start, end := p.reportRange.Pos(), p.reportRange.End()
	calleeEnd := p.callee.End()
	if start < 0 || start > end || end > len(sourceText) || calleeEnd < start || calleeEnd > end {
		return
	}
	if p.node.Flags&(ast.NodeFlagsThisNodeHasError|ast.NodeFlagsThisNodeOrAnySubNodesHasError) != 0 ||
		diagnosticTouchesRange(p.sourceFile.Diagnostics(), p.reportRange) ||
		// Despite its name, JSDiagnostics contains JavaScript-file syntax
		// diagnostics (for example, TypeScript-only syntax in a .js file).
		// JSDocDiagnostics is intentionally excluded: malformed comment types
		// do not recover the call AST, and the edit preserves argument text.
		diagnosticTouchesRange(p.sourceFile.JSDiagnostics(), p.reportRange) {
		return
	}
	if p.args != nil {
		for _, arg := range p.args.Nodes {
			if ast.NodeIsMissing(arg) || arg.Kind == ast.KindOmittedExpression {
				return
			}
		}
	}

	openParen, hasParens := findOpenParen(p.sourceFile, calleeEnd, end)
	commentBound := end
	if hasParens {
		closeParen := end - 1
		if openParen >= closeParen || sourceText[closeParen] != ')' {
			return
		}
		p.argsStart = openParen + 1
		p.argsEnd = closeParen
		commentBound = openParen
	} else {
		// Only NewExpression permits a valid constructor call without its own
		// parentheses (`new Array` or `new (Array)`). A CallExpression with a
		// nil argument list is necessarily recovered or synthetic.
		if p.node.Kind != ast.KindNewExpression || p.args != nil {
			return
		}
		p.argsStart = end
		p.argsEnd = end
	}

	p.editable = true
	p.shouldSuggest = ast.IsOptionalChainRoot(p.node) ||
		(p.args != nil && len(p.args.Nodes) > 0 && countNonSpread(p.args) < 2) ||
		utils.HasCommentInSpan(p.comments.All(), start, commentBound)
}

func (p *arrayConstructorEditPlan) replacement() (text string, addSemicolon bool) {
	addSemicolon = utils.IsStartOfExpressionStatement(p.sourceFile, p.node) &&
		utils.NeedsPrecedingSemicolon(p.sourceFile, p.node)
	if p.argsStart == p.argsEnd {
		if addSemicolon {
			return ";[]", true
		}
		return "[]", false
	}

	argsText := p.sourceFile.Text()[p.argsStart:p.argsEnd]
	if addSemicolon {
		return ";[" + argsText + "]", true
	}
	return "[" + argsText + "]", false
}

func (p *arrayConstructorEditPlan) buildFixes() []rule.RuleFix {
	p.classify()
	if !p.editable || p.shouldSuggest {
		return nil
	}
	fixText, _ := p.replacement()
	return []rule.RuleFix{rule.RuleFixReplaceRange(p.reportRange, fixText)}
}

func (p *arrayConstructorEditPlan) buildSuggestions() []rule.RuleSuggestion {
	p.classify()
	if !p.editable || !p.shouldSuggest {
		return nil
	}
	fixText, addSemicolon := p.replacement()
	msg := useLiteralMessage()
	if addSemicolon {
		msg = useLiteralAfterSemicolonMessage()
	}
	return []rule.RuleSuggestion{{
		Message:  msg,
		FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(p.reportRange, fixText)},
	}}
}

// https://eslint.org/docs/latest/rules/no-array-constructor
var NoArrayConstructorRule = rule.Rule{
	Name:   "no-array-constructor",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var shadowCache utils.ScopeManagerShadowCache
		check := func(node *ast.Node) {
			callee, args, typeArgs := calleeArgsAndTypeArgs(node)
			if callee == nil {
				return
			}
			// ESTree flattens grouping parentheses around the callee, so
			// `(Array)()` still has an Identifier callee; TS-only wrappers
			// (non-null assertion, `as`/`satisfies`) are deliberately left
			// alone since ESTree never collapses those back to a bare
			// Identifier either.
			callee = ast.SkipParentheses(callee)
			if callee == nil || callee.Kind != ast.KindIdentifier {
				return
			}
			if callee.AsIdentifier().Text != "Array" {
				return
			}
			// `Array<Foo>()` — explicit type arguments mean this isn't the
			// bare-constructor call the rule targets.
			if typeArgs != nil {
				return
			}
			// A single non-spread argument (`Array(9)`, `Array(x)`) may be
			// intentionally creating a sparse array of that length — leave
			// it alone, since the argument's runtime type can't be known
			// statically.
			if args != nil && len(args.Nodes) == 1 && args.Nodes[0].Kind != ast.KindSpreadElement {
				return
			}

			// A local declaration shadows the global constructor. The binder
			// lookup resolves the declaration once without rescanning every
			// enclosing body for each matching call. Keep the AST fallback for
			// manually assembled rule contexts that do not provide a RefStore.
			if ctx.Refs == nil {
				if utils.IsShadowed(callee, "Array") {
					return
				}
			} else if ctx.Refs.IsNameDefinedInFileWithMeaning(
				callee,
				"Array",
				ast.SymbolFlagsValue|ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias,
			) {
				return
			}
			// scope-manager keeps function parameters visible directly inside
			// parameter decorators and class type parameters visible inside
			// static members, where TypeScript's resolver hides them. It also
			// keeps a function's parameter initializers in the same scope as its
			// body declarations, where both TypeScript's resolver and IsShadowed
			// follow runtime lexical semantics instead.
			if shadowCache.HasEnclosingParameter(callee, "Array") ||
				shadowCache.HasEnclosingTypeParameter(callee, "Array") ||
				utils.HasEnclosingClassExpressionName(callee, "Array") ||
				shadowCache.IsShadowedFromParameterInitializer(callee, "Array") {
				return
			}
			// A config `/* global Array: off */` / `languageOptions.globals`
			// entry un-declares Espree's builtin. typescript-eslint/parser also
			// installs its library `Array` variable, which remains visible after
			// the override, so TypeScript-flavoured files still report. rslint does
			// not currently project parserOptions.lib into native rule contexts;
			// the documented divergence for an explicitly empty lib follows from
			// preserving that existing API boundary.
			if ast.IsInJSFile(callee) && !ctx.Globals.Access("Array").IsDeclared() {
				return
			}

			reportRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			plan := arrayConstructorEditPlan{
				sourceFile:  ctx.SourceFile,
				comments:    ctx.Comments,
				node:        node,
				callee:      callee,
				args:        args,
				reportRange: reportRange,
			}

			// NOTE: Unlike ESLint, utils.NeedsPrecedingSemicolon doesn't model
			// TypeScript-only node kinds (type positions, ambient/overload
			// function declarations, import-equals declarations), so it falls
			// back to the conservative "needs a semicolon" answer there. The
			// edit plan keeps that existing behavior, but computes it only when
			// the consumer requests the matching edit category.
			//
			// The same deferred plan classifies optional calls, risky spreads,
			// comments, and parser-recovery boundaries. A malformed call keeps
			// its diagnostic but never receives an edit that could drop source.
			ctx.ReportRangeWithDeferredFixesAndSuggestions(
				reportRange,
				preferLiteralMessage(),
				plan.buildFixes,
				plan.buildSuggestions,
			)
		}

		return rule.RuleListeners{
			ast.KindCallExpression: check,
			ast.KindNewExpression:  check,
		}
	},
}
