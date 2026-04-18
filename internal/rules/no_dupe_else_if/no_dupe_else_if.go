package no_dupe_else_if

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-dupe-else-if
var NoDupeElseIfRule = rule.Rule{
	Name: "no-dupe-else-if",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		sf := ctx.SourceFile
		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				ifStmt := node.AsIfStatement()
				if ifStmt == nil || ifStmt.Expression == nil {
					return
				}
				test := ifStmt.Expression

				// ESLint: when the current test is a top-level `&&`, each AND
				// operand is also checked independently. A branch is dead if
				// any of {whole test, A, B, ...} becomes fully covered.
				conditionsToCheck := []*ast.Node{test}
				if isLogicalOp(test, ast.KindAmpersandAmpersandToken) {
					conditionsToCheck = append(conditionsToCheck, splitByAnd(test)...)
				}

				// For each condition to check: list of OR-operands, each an AND-set.
				listToCheck := make([][][]*ast.Node, len(conditionsToCheck))
				for i, c := range conditionsToCheck {
					listToCheck[i] = toOrAndMatrix(c)
				}

				current := node
				for current.Parent != nil && current.Parent.Kind == ast.KindIfStatement {
					parent := current.Parent
					parentIf := parent.AsIfStatement()
					if parentIf == nil || parentIf.ElseStatement != current {
						break
					}
					current = parent
					if parentIf.Expression == nil {
						break
					}

					priorOrAnd := toOrAndMatrix(parentIf.Expression)

					// Cumulative filter: drop OR-operands already covered by
					// any ancestor's OR-operand (i.e. an ancestor operand that
					// is an AND-subset of ours implies ours).
					anyEmpty := false
					for i, orOperands := range listToCheck {
						filtered := orOperands[:0]
						for _, orOp := range orOperands {
							covered := false
							for _, priorOp := range priorOrAnd {
								if andSetIsSubset(priorOp, orOp, sf) {
									covered = true
									break
								}
							}
							if !covered {
								filtered = append(filtered, orOp)
							}
						}
						listToCheck[i] = filtered
						if len(filtered) == 0 {
							anyEmpty = true
						}
					}

					if anyEmpty {
						ctx.ReportNode(ast.SkipParentheses(test), rule.RuleMessage{
							Id:          "unexpected",
							Description: "This branch can never execute. Its condition is a duplicate or covered by previous conditions in the if-else-if chain.",
						})
						return
					}
				}
			},
		}
	},
}

// toOrAndMatrix splits `node` by top-level `||` into OR-operands, then splits
// each OR-operand by top-level `&&` into AND-operands.
func toOrAndMatrix(node *ast.Node) [][]*ast.Node {
	orOps := splitByOr(node)
	matrix := make([][]*ast.Node, len(orOps))
	for i, o := range orOps {
		matrix[i] = splitByAnd(o)
	}
	return matrix
}

func splitByOr(node *ast.Node) []*ast.Node {
	return splitByLogicalOp(node, ast.KindBarBarToken)
}

func splitByAnd(node *ast.Node) []*ast.Node {
	return splitByLogicalOp(node, ast.KindAmpersandAmpersandToken)
}

// splitByLogicalOp flattens a binary expression tree along a single logical
// operator. Parentheses at the root are transparent to the split (matching
// ESLint's ESTree-based behavior, where parens don't appear as nodes).
func splitByLogicalOp(node *ast.Node, op ast.Kind) []*ast.Node {
	if node == nil {
		return nil
	}
	inner := ast.SkipParentheses(node)
	if inner.Kind == ast.KindBinaryExpression {
		bin := inner.AsBinaryExpression()
		if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == op {
			return append(splitByLogicalOp(bin.Left, op), splitByLogicalOp(bin.Right, op)...)
		}
	}
	return []*ast.Node{node}
}

// isLogicalOp reports whether `node` (ignoring outer parentheses) is a binary
// expression with the given logical operator.
func isLogicalOp(node *ast.Node, op ast.Kind) bool {
	inner := ast.SkipParentheses(node)
	if inner == nil || inner.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := inner.AsBinaryExpression()
	return bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == op
}

// nodesEqual mirrors ESLint's recursive `equal`: `||` and `&&` are treated as
// commutative at any depth; everything else falls back to token equality.
func nodesEqual(a, b *ast.Node, sf *ast.SourceFile) bool {
	ua := ast.SkipParentheses(a)
	ub := ast.SkipParentheses(b)
	if ua.Kind != ub.Kind {
		return false
	}
	if ua.Kind == ast.KindBinaryExpression {
		ba := ua.AsBinaryExpression()
		bb := ub.AsBinaryExpression()
		if ba != nil && bb != nil && ba.OperatorToken != nil && bb.OperatorToken != nil &&
			ba.OperatorToken.Kind == bb.OperatorToken.Kind &&
			(ba.OperatorToken.Kind == ast.KindBarBarToken ||
				ba.OperatorToken.Kind == ast.KindAmpersandAmpersandToken) {
			return (nodesEqual(ba.Left, bb.Left, sf) && nodesEqual(ba.Right, bb.Right, sf)) ||
				(nodesEqual(ba.Left, bb.Right, sf) && nodesEqual(ba.Right, bb.Left, sf))
		}
	}
	return tokenSignature(sf, ua) == tokenSignature(sf, ub)
}

// andSetIsSubset reports whether every element of `sub` has a structurally
// equal element in `super`, per nodesEqual semantics.
func andSetIsSubset(sub, super []*ast.Node, sf *ast.SourceFile) bool {
	for _, s := range sub {
		found := false
		for _, t := range super {
			if nodesEqual(s, t, sf) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// tokenSignature produces a canonical whitespace-delimited token string used
// for structural equality of leaf expressions. Matches ESLint's equalTokens
// approach of comparing token streams, ignoring trivia.
func tokenSignature(sf *ast.SourceFile, expr *ast.Node) string {
	var b strings.Builder
	src := sf.Text()
	first := true
	utils.ForEachToken(expr, func(token *ast.Node) {
		r := utils.TrimNodeTextRange(sf, token)
		if r.Pos() < r.End() {
			if !first {
				b.WriteByte(' ')
			}
			b.WriteString(src[r.Pos():r.End()])
			first = false
		}
	}, sf)
	return b.String()
}
