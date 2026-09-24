package no_empty_named_blocks

import (
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-empty-named-blocks.js
var NoEmptyNamedBlocksRule = rule.Rule{
	Name:   "import/no-empty-named-blocks",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration.ImportClause == nil {
					return
				}
				clause := declaration.ImportClause.AsImportClause()
				if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports ||
					len(clause.NamedBindings.AsNamedImports().Elements.Nodes) != 0 {
					return
				}

				// Upstream uses literal messages without message IDs.
				message := rule.RuleMessage{Description: "Unexpected empty named import block"}
				nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				if clause.Name() != nil {
					ctx.ReportRangeWithDeferredFixes(nodeRange, message, func() []rule.RuleFix {
						comma, ok := utils.TokenAtOrAfter(ctx.SourceFile, clause.Name().End())
						if !ok || comma.Kind != ast.KindCommaToken {
							return nil
						}
						replacement := ""
						if after, ok := utils.TokenAtOrAfter(ctx.SourceFile, clause.NamedBindings.End()); ok &&
							comma.Start == clause.Name().End() && after.Start == clause.NamedBindings.End() &&
							!utils.CanTokenTextsBeAdjacent(utils.TrimmedNodeText(ctx.SourceFile, clause.Name()), after.Text) {
							replacement = " "
						}
						return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(comma.Start, clause.NamedBindings.End()), replacement)}
					})
					return
				}

				ctx.ReportRangeWithDeferredSuggestions(nodeRange, message, func() []rule.RuleSuggestion {
					suggestions := make([]rule.RuleSuggestion, 1, 2)
					suggestions[0] = rule.RuleSuggestion{
						Message:  rule.RuleMessage{Description: "Remove unused import"},
						FixesArr: []rule.RuleFix{rule.RuleFixRemoveRange(nodeRange)},
					}
					// Type import attributes such as resolution-mode require import type.
					if clause.IsTypeOnly() && declaration.Attributes != nil {
						return suggestions
					}
					from, ok := utils.TokenAtOrAfter(ctx.SourceFile, clause.NamedBindings.End())
					if !ok || from.Kind != ast.KindFromKeyword {
						return suggestions
					}
					return append(suggestions, rule.RuleSuggestion{
						Message:  rule.RuleMessage{Description: "Remove empty import block"},
						FixesArr: []rule.RuleFix{sideEffectFix(ctx.SourceFile, node, from)},
					})
				})
			},
		}
	},
}

func sideEffectFix(file *ast.SourceFile, node *ast.Node, from utils.SourceToken) rule.RuleFix {
	declaration := node.AsImportDeclaration()
	start := utils.TrimNodeTextRange(file, declaration.ImportClause).Pos()
	end := from.End

	// Match upstream's spacing, excluding whitespace inside comments. Limit
	// the scan and edit to this declaration, even when earlier imports exist.
	replacement := " "
	s := scanner.GetScannerForSourceFile(file, node.Pos())
	s.SetSkipTrivia(false)
	sourceStart := scanner.GetTokenPosOfNode(declaration.ModuleSpecifier, file, false)
	for kind := s.Scan(); kind != ast.KindEndOfFile && s.TokenStart() < sourceStart; kind = s.Scan() {
		if kind == ast.KindWhitespaceTrivia || kind == ast.KindNewLineTrivia {
			replacement = ""
			break
		}
	}

	// Consume at most one whitespace character after `from`, never part of
	// a comment or a UTF-8 character.
	ch, size := utf8.DecodeRuneInString(file.Text()[end:])
	if ch < utf8.RuneSelf && ecmascript.IsTriviaWhitespaceByte(byte(ch)) || ecmascript.IsTriviaWhitespaceRune(ch) {
		end += size
	}
	return rule.RuleFixReplaceRange(core.NewTextRange(start, end), replacement)
}
