package no_fallthrough

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-fallthrough
var NoFallthroughRule = rule.Rule{
	Name: "no-fallthrough",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
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
				sourceText := ctx.SourceFile.Text()

				for i := range len(clauses) - 1 {
					currentClause := clauses[i].AsCaseOrDefaultClause()
					if currentClause == nil {
						continue
					}

					// Skip empty cases (no statements)
					if currentClause.Statements == nil || len(currentClause.Statements.Nodes) == 0 {
						continue
					}

					// Check if the last statement is a terminal statement
					lastStmt := currentClause.Statements.Nodes[len(currentClause.Statements.Nodes)-1]
					if isTerminalStatement(lastStmt) {
						continue
					}

					// Check for "falls through" comment between this case and next.
					// node.Pos() includes leading trivia, so a comment like
					// /* falls through */ before the next case keyword is part
					// of the next clause's leading trivia. We need to search
					// up to the actual keyword token position of the next clause.
					nextClause := clauses[i+1]
					commentStart := getStatementsEnd(currentClause)
					nextKeywordPos := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, nextClause.Pos()).Pos()
					if hasFallthroughComment(sourceText, commentStart, nextKeywordPos) {
						continue
					}

					// Report on the next case/default clause
					var msgId string
					var description string
					if nextClause.Kind == ast.KindDefaultClause {
						msgId = "default"
						description = "Expected a 'break' statement before 'default'."
					} else {
						msgId = "case"
						description = "Expected a 'break' statement before 'case'."
					}

					ctx.ReportNode(nextClause, rule.RuleMessage{
						Id:          msgId,
						Description: description,
					})
				}
			},
		}
	},
}

// getStatementsEnd returns the end position of the last statement in a clause.
func getStatementsEnd(clause *ast.CaseOrDefaultClause) int {
	if clause.Statements == nil || len(clause.Statements.Nodes) == 0 {
		return 0
	}
	lastStmt := clause.Statements.Nodes[len(clause.Statements.Nodes)-1]
	return lastStmt.End()
}

// isTerminalStatement checks if a statement is a terminal statement
// (break, return, throw, continue), or an if/else where all branches terminate.
func isTerminalStatement(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement,
		ast.KindBreakStatement, ast.KindContinueStatement:
		return true
	case ast.KindBlock:
		// Check the last statement in the block
		block := node.AsBlock()
		if block.Statements != nil && len(block.Statements.Nodes) > 0 {
			return isTerminalStatement(block.Statements.Nodes[len(block.Statements.Nodes)-1])
		}
		return false
	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		// Both branches must exist and terminate
		if ifStmt.ElseStatement == nil {
			return false
		}
		return isTerminalStatement(ifStmt.ThenStatement) && isTerminalStatement(ifStmt.ElseStatement)
	}

	return false
}

// hasFallthroughComment checks if there is a "falls through" or "fall through"
// comment in the source text between start and end positions.
func hasFallthroughComment(sourceText string, start, end int) bool {
	if start < 0 || end > len(sourceText) || start >= end {
		return false
	}
	text := strings.ToLower(sourceText[start:end])
	return strings.Contains(text, "falls through") || strings.Contains(text, "fall through")
}
