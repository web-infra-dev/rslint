package no_useless_switch_case

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const (
	messageIDError      = "no-useless-switch-case/error"
	messageIDSuggestion = "no-useless-switch-case/suggestion"
)

var NoUselessSwitchCaseRule = rule.Rule{
	Name:   "unicorn/no-useless-switch-case",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				switchStmt := node.AsSwitchStatement()
				if switchStmt == nil || switchStmt.CaseBlock == nil {
					return
				}

				caseBlock := switchStmt.CaseBlock.AsCaseBlock()
				if caseBlock == nil || caseBlock.Clauses == nil {
					return
				}

				clauses := caseBlock.Clauses.Nodes
				if len(clauses) < 2 || clauses[len(clauses)-1].Kind != ast.KindDefaultClause {
					return
				}

				// Upstream walks backward from the final `default` and stops at
				// the first non-empty case, but ESLint still emits diagnostics in
				// source order. Remember the boundary instead of allocating a
				// temporary slice for every switch that ends in `default`.
				firstCandidate := len(clauses) - 1
				for index := len(clauses) - 2; index >= 0; index-- {
					clause := clauses[index]
					if !isEmptySwitchCase(clause) {
						break
					}

					firstCandidate = index
				}

				for index := firstCandidate; index < len(clauses)-1; index++ {
					clause := clauses[index]
					headRange := switchCaseHeadRange(clause, ctx.SourceFile)
					removalRange := headRange.WithEnd(clause.End())
					ctx.ReportRangeWithDeferredSuggestions(
						headRange,
						rule.RuleMessage{
							Id:          messageIDError,
							Description: "Useless case in switch statement.",
						},
						func() []rule.RuleSuggestion {
							return []rule.RuleSuggestion{
								{
									Message: rule.RuleMessage{
										Id:          messageIDSuggestion,
										Description: "Remove this case.",
									},
									FixesArr: []rule.RuleFix{rule.RuleFixRemoveRange(removalRange)},
								},
							}
						},
					)
				}
			},
		}
	},
}

func isEmptySwitchCase(node *ast.Node) bool {
	clause := node.AsCaseOrDefaultClause()
	if clause == nil || clause.Statements == nil {
		return true
	}
	for _, statement := range clause.Statements.Nodes {
		if !isEmptyNode(statement) {
			return false
		}
	}
	return true
}

func isEmptyNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsEmptyStatement(node) {
		return true
	}
	if ast.IsBlock(node) {
		block := node.AsBlock()
		if block == nil || block.Statements == nil {
			return true
		}
		for _, statement := range block.Statements.Nodes {
			if !isEmptyNode(statement) {
				return false
			}
		}
		return true
	}
	return false
}

func switchCaseHeadRange(node *ast.Node, sourceFile *ast.SourceFile) core.TextRange {
	start := scanner.GetTokenPosOfNode(node, sourceFile, false)
	scanStart := start
	if clause := node.AsCaseOrDefaultClause(); clause != nil && clause.Expression != nil {
		scanStart = clause.Expression.End()
	}

	text := sourceFile.Text()
	colonStart := scanner.SkipTrivia(text, scanStart)
	if colonStart < node.End() && colonStart < len(text) && text[colonStart] == ':' {
		return core.NewTextRange(start, colonStart+1)
	}

	// Preserve the recovery behavior for malformed clauses where the next
	// non-trivia byte is not the separator the parser associated with the case.
	s := scanner.GetScannerForSourceFile(sourceFile, scanStart)
	for s.Token() != ast.KindEndOfFile && s.TokenStart() < node.End() {
		if s.Token() == ast.KindColonToken {
			return core.NewTextRange(start, s.TokenEnd())
		}
		s.Scan()
	}

	return core.NewTextRange(start, start)
}
