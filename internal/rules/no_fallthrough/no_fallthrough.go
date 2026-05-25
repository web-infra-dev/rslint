package no_fallthrough

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// defaultFallthroughPattern matches ESLint's default /falls?\s?through/iu
var defaultFallthroughPattern = regexp.MustCompile(`(?i)falls?\s?through`)

type noFallthroughOptions struct {
	commentPattern                 *regexp.Regexp
	allowEmptyCase                 bool
	reportUnusedFallthroughComment bool
}

func parseOptions(opts any) noFallthroughOptions {
	result := noFallthroughOptions{
		commentPattern:                 defaultFallthroughPattern,
		allowEmptyCase:                 false,
		reportUnusedFallthroughComment: false,
	}

	optsMap := utils.GetOptionsMap(opts)
	if optsMap == nil {
		return result
	}

	if pattern, ok := optsMap["commentPattern"].(string); ok && pattern != "" {
		if compiled, err := regexp.Compile("(?i)" + pattern); err == nil {
			result.commentPattern = compiled
		}
	}
	if allow, ok := optsMap["allowEmptyCase"].(bool); ok {
		result.allowEmptyCase = allow
	}
	if report, ok := optsMap["reportUnusedFallthroughComment"].(bool); ok {
		result.reportUnusedFallthroughComment = report
	}

	return result
}

// https://eslint.org/docs/latest/rules/no-fallthrough
var NoFallthroughRule = rule.Rule{
	Name: "no-fallthrough",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

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

					statements := currentClause.Statements
					if statements == nil || len(statements.Nodes) == 0 {
						// Empty cases (no statements) are always allowed.
						continue
					}

					if opts.allowEmptyCase && isCaseBodyEmpty(statements.Nodes) {
						continue
					}

					// Determine if control flow can fall through to the next case.
					//
					// Primary: the binder sets FallthroughFlowNode when flow can reach
					// the end of a case (handles break/return/throw/continue, if-else
					// with both branches, unreachable code after a terminal, etc.).
					// If nil, definitely no fallthrough.
					//
					// Refinement: the binder is conservative about certain patterns
					// that ESLint's CodePath analysis handles (infinite loops with
					// literal truthy conditions, try blocks whose catch is unreachable).
					// When FallthroughFlowNode is set, apply targeted AST checks for
					// these patterns.
					isFallthrough := currentClause.FallthroughFlowNode != nil
					if isFallthrough && hasASTTerminal(statements.Nodes) {
						isFallthrough = false
					}

					nextClause := clauses[i+1]
					lastStmt := statements.Nodes[len(statements.Nodes)-1]
					commentStart := lastStmt.End()
					nextKeywordPos := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, nextClause.Pos()).Pos()
					hasComment := hasFallthroughComment(sourceText, commentStart, nextKeywordPos, opts.commentPattern)

					if !isFallthrough {
						if opts.reportUnusedFallthroughComment && hasComment {
							ctx.ReportNode(nextClause, rule.RuleMessage{
								Id:          "unusedFallthroughComment",
								Description: "Found a comment that would permit fallthrough, but case cannot fall through.",
							})
						}
						continue
					}

					if hasComment {
						continue
					}

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

// isCaseBodyEmpty checks if all statements in a case are empty statements.
func isCaseBodyEmpty(nodes []*ast.Node) bool {
	for _, stmt := range nodes {
		if stmt.Kind != ast.KindEmptyStatement {
			return false
		}
	}
	return true
}

// ============================================================================
// AST-based refinements for patterns the binder's flow analysis misses.
// These only run when FallthroughFlowNode is non-nil (binder says "maybe
// fallthrough") and cover:
//   - Infinite loops with literal truthy conditions (while("x"), while(1_000))
//   - Try blocks whose catch is unreachable (try { break; } catch(e) { ... })
//   - Nested try/finally where finally overrides with a terminal
// ============================================================================

// hasASTTerminal checks if any statement in the list is a terminal that the
// binder's flow analysis missed.
func hasASTTerminal(nodes []*ast.Node) bool {
	for _, stmt := range nodes {
		if isASTTerminal(stmt) {
			return true
		}
	}
	return false
}

// isASTTerminal checks if a statement is terminal using AST heuristics,
// covering patterns the binder is conservative about.
func isASTTerminal(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindBlock:
		return hasASTTerminal(node.Statements())

	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt.ElseStatement == nil {
			return false
		}
		return isASTTerminal(ifStmt.ThenStatement) && isASTTerminal(ifStmt.ElseStatement)

	case ast.KindLabeledStatement:
		ls := node.AsLabeledStatement()
		if ls != nil && ls.Statement != nil {
			return isASTTerminal(ls.Statement)
		}
		return false

	case ast.KindWhileStatement:
		ws := node.AsWhileStatement()
		return ws != nil && isAlwaysTruthy(ws.Expression) && !hasBreakInLoopBody(node)

	case ast.KindForStatement:
		fs := node.AsForStatement()
		return fs != nil && isAlwaysTruthy(fs.Condition) && !hasBreakInLoopBody(node)

	case ast.KindDoStatement:
		ds := node.AsDoStatement()
		return ds != nil && isAlwaysTruthy(ds.Expression) && !hasBreakInLoopBody(node)

	case ast.KindTryStatement:
		return isTryASTTerminal(node)
	}
	return false
}

// isTryASTTerminal handles try/catch/finally patterns the binder misses:
//   - finally with a terminal overrides the try/catch completion
//   - try body that cannot throw makes the catch unreachable
func isTryASTTerminal(node *ast.Node) bool {
	ts := node.AsTryStatement()
	if ts == nil {
		return false
	}

	// If finally has a terminal, it overrides try/catch completion.
	if ts.FinallyBlock != nil && utils.BlockEndsWithTerminal(ts.FinallyBlock) {
		return true
	}

	// If the try body cannot throw before its terminal, catch is unreachable.
	if ts.TryBlock != nil && !utils.CanBlockThrow(ts.TryBlock) {
		return true
	}

	return false
}

// isAlwaysTruthy checks if an expression is a literal that is always truthy.
// The binder handles `true` and missing conditions (for(;;)), but not other
// literal types like non-zero numbers or non-empty strings.
func isAlwaysTruthy(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindTrueKeyword:
		return true
	case ast.KindNumericLiteral:
		text := strings.ReplaceAll(node.Text(), "_", "")
		if val, err := strconv.ParseInt(text, 0, 64); err == nil {
			return val != 0
		}
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			return val != 0
		}
		return false
	case ast.KindStringLiteral:
		return len(node.Text()) > 0
	}
	return false
}

// hasBreakInLoopBody checks if the loop body contains a break that targets
// this loop (unlabeled or matching the loop's label).
func hasBreakInLoopBody(loopNode *ast.Node) bool {
	var loopLabels []string
	for p := loopNode.Parent; p != nil && ast.IsLabeledStatement(p); p = p.Parent {
		ls := p.AsLabeledStatement()
		if ls != nil && ls.Label != nil {
			loopLabels = append(loopLabels, ls.Label.Text())
		}
	}

	found := false
	forEachDescendant(loopNode, func(n *ast.Node) bool {
		if found {
			return false
		}
		if ast.IsSwitchStatement(n) || ast.IsIterationStatement(n, false) {
			return false
		}
		if ast.IsFunctionLike(n) {
			return false
		}
		if n.Kind == ast.KindBreakStatement {
			bs := n.AsBreakStatement()
			if bs == nil || bs.Label == nil {
				found = true
			} else {
				label := bs.Label.Text()
				for _, l := range loopLabels {
					if label == l {
						found = true
						break
					}
				}
			}
			return false
		}
		return true
	})
	return found
}

// forEachDescendant walks all descendants of a node depth-first.
// If fn returns false for a node, its children are not visited.
func forEachDescendant(node *ast.Node, fn func(*ast.Node) bool) {
	if node == nil {
		return
	}
	node.ForEachChild(func(child *ast.Node) bool {
		if fn(child) {
			forEachDescendant(child, fn)
		}
		return false
	})
}

// hasFallthroughComment checks if there is a fallthrough comment matching the
// given pattern between start and end positions in the source text.
func hasFallthroughComment(sourceText string, start, end int, pattern *regexp.Regexp) bool {
	if start < 0 || end > len(sourceText) || start >= end {
		return false
	}
	return pattern.MatchString(sourceText[start:end])
}
