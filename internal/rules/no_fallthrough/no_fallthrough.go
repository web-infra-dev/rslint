package no_fallthrough

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// defaultFallthroughPattern matches ESLint's default /falls?\s?through/iu
var defaultFallthroughPattern = regexp.MustCompile(`(?i)falls?\s?through`)

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

					// Check if the last statement is a terminal statement.
					// Pass the switch node so break/continue targeting is correct.
					lastStmt := currentClause.Statements.Nodes[len(currentClause.Statements.Nodes)-1]
					if isTerminalStatement(lastStmt, node) {
						continue
					}

					// Check for "falls through" comment between this case and next.
					nextClause := clauses[i+1]
					commentStart := getStatementsEnd(currentClause)
					nextKeywordPos := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, nextClause.Pos()).Pos()
					if hasFallthroughComment(sourceText, commentStart, nextKeywordPos) {
						continue
					}

					// Report on the next case/default clause
					var msgID string
					var description string
					if nextClause.Kind == ast.KindDefaultClause {
						msgID = "default"
						description = "Expected a 'break' statement before 'default'."
					} else {
						msgID = "case"
						description = "Expected a 'break' statement before 'case'."
					}

					ctx.ReportNode(nextClause, rule.RuleMessage{
						Id:          msgID,
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

// isTerminalStatement checks if a statement terminates control flow
// for the given switch statement (switchNode).
// A break without label terminates if it targets this switch (not a nested one).
// A labeled break terminates only if the label targets something outside this switch.
func isTerminalStatement(node *ast.Node, switchNode *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement:
		return true

	case ast.KindContinueStatement:
		// continue always exits the current switch case (goes to loop)
		return true

	case ast.KindBreakStatement:
		breakStmt := node.AsBreakStatement()
		if breakStmt == nil {
			return true
		}
		// Labeled break: only terminal if label is NOT inside this switch
		if breakStmt.Label != nil {
			return !isLabelInsideNode(breakStmt.Label.Text(), switchNode)
		}
		// Unlabeled break: only terminal if it targets THIS switch,
		// not a nested switch/loop. Check if there's a nested switch/loop between
		// this break and our switchNode.
		return !hasNestedBreakTarget(node, switchNode)

	case ast.KindBlock:
		block := node.AsBlock()
		if block.Statements != nil && len(block.Statements.Nodes) > 0 {
			return isTerminalStatement(block.Statements.Nodes[len(block.Statements.Nodes)-1], switchNode)
		}
		return false

	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt.ElseStatement == nil {
			return false
		}
		return isTerminalStatement(ifStmt.ThenStatement, switchNode) && isTerminalStatement(ifStmt.ElseStatement, switchNode)

	case ast.KindTryStatement:
		tryStmt := node.AsTryStatement()
		if tryStmt == nil {
			return false
		}
		// Try is terminal if both try-block and catch-clause terminate
		// (finally doesn't affect fallthrough analysis here)
		tryTerminal := false
		if tryStmt.TryBlock != nil {
			tryTerminal = isTerminalStatement(tryStmt.TryBlock, switchNode)
		}
		if !tryTerminal {
			return false
		}
		if tryStmt.CatchClause != nil {
			cc := tryStmt.CatchClause.AsCatchClause()
			if cc == nil || cc.Block == nil {
				return false
			}
			return isTerminalStatement(cc.Block, switchNode)
		}
		// No catch clause but try terminates — terminal
		return true
	}

	return false
}

// hasNestedBreakTarget checks if there is a switch/loop between node and switchNode
// that would capture an unlabeled break.
func hasNestedBreakTarget(node *ast.Node, switchNode *ast.Node) bool {
	current := node.Parent
	for current != nil && current != switchNode {
		switch current.Kind {
		case ast.KindSwitchStatement,
			ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindWhileStatement, ast.KindDoStatement:
			return true
		}
		current = current.Parent
	}
	return false
}

// isLabelInsideNode checks if a label with the given name is defined inside the node.
func isLabelInsideNode(labelName string, node *ast.Node) bool {
	found := false
	walkNode(node, func(n *ast.Node) bool {
		if found {
			return false
		}
		if n.Kind == ast.KindLabeledStatement {
			ls := n.AsLabeledStatement()
			if ls != nil && ls.Label != nil && ls.Label.Text() == labelName {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

// walkNode walks all children of a node, calling fn for each.
// If fn returns false, children of that node are not visited.
func walkNode(node *ast.Node, fn func(*ast.Node) bool) {
	if node == nil {
		return
	}
	node.ForEachChild(func(child *ast.Node) bool {
		if fn(child) {
			walkNode(child, fn)
		}
		return false
	})
}

// hasFallthroughComment checks if there is a fallthrough comment matching
// ESLint's default pattern /falls?\s?through/iu.
func hasFallthroughComment(sourceText string, start, end int) bool {
	if start < 0 || end > len(sourceText) || start >= end {
		return false
	}
	text := strings.ToLower(sourceText[start:end])
	return defaultFallthroughPattern.MatchString(text)
}
