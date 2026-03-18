package no_dupe_else_if

import (
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-dupe-else-if
var NoDupeElseIfRule = rule.Rule{
	Name: "no-dupe-else-if",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				// Only check if this IfStatement is in an "else if" position.
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindIfStatement {
					return
				}
				parentIf := parent.AsIfStatement()
				if parentIf == nil || parentIf.ElseStatement != node {
					return
				}

				ifStmt := node.AsIfStatement()
				if ifStmt == nil || ifStmt.Expression == nil {
					return
				}

				// Split current condition into its OR-operands, each split into AND-operands.
				currentOrOperands := splitByOr(ctx.SourceFile, ifStmt.Expression)

				// Walk up the if-else-if chain collecting all prior conditions.
				ancestor := parent
				for ancestor != nil && ancestor.Kind == ast.KindIfStatement {
					ancestorIf := ancestor.AsIfStatement()
					if ancestorIf == nil || ancestorIf.Expression == nil {
						break
					}

					priorOrOperands := splitByOr(ctx.SourceFile, ancestorIf.Expression)

					// Check if every or-operand in the current condition is a subset
					// of at least one or-operand in the prior condition.
					// This means the current condition can never be true if the prior was false.
					if isSubset(currentOrOperands, priorOrOperands) {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "unexpected",
							Description: "This branch can never execute. Its condition is a duplicate or covered by previous conditions in the if-else-if chain.",
						})
						return
					}

					// Move to the next ancestor in the chain.
					grandparent := ancestor.Parent
					if grandparent == nil || grandparent.Kind != ast.KindIfStatement {
						break
					}
					grandparentIf := grandparent.AsIfStatement()
					if grandparentIf == nil || grandparentIf.ElseStatement != ancestor {
						break
					}
					ancestor = grandparent
				}
			},
		}
	},
}

// operandSet represents a set of AND-operands (token signatures), sorted for comparison.
type operandSet []string

// splitByOr splits an expression by top-level || into OR-operands,
// then each OR-operand is split by top-level && into AND-operands.
// Returns a slice of operandSets, one per OR-operand.
func splitByOr(sf *ast.SourceFile, expr *ast.Node) []operandSet {
	orNodes := splitByOperator(expr, ast.KindBarBarToken)
	result := make([]operandSet, 0, len(orNodes))
	for _, orNode := range orNodes {
		andNodes := splitByOperator(orNode, ast.KindAmpersandAmpersandToken)
		sigs := make(operandSet, 0, len(andNodes))
		for _, andNode := range andNodes {
			sigs = append(sigs, getExpressionTokenSignature(sf, andNode))
		}
		// Sort for commutative comparison (a && b == b && a)
		sort.Strings(sigs)
		result = append(result, sigs)
	}
	return result
}

// splitByOperator recursively splits a binary expression tree by the given operator.
func splitByOperator(node *ast.Node, op ast.Kind) []*ast.Node {
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindParenthesizedExpression {
		pe := node.AsParenthesizedExpression()
		if pe != nil && pe.Expression != nil {
			return splitByOperator(pe.Expression, op)
		}
	}
	if node.Kind == ast.KindBinaryExpression {
		bin := node.AsBinaryExpression()
		if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == op {
			left := splitByOperator(bin.Left, op)
			right := splitByOperator(bin.Right, op)
			return append(left, right...)
		}
	}
	return []*ast.Node{node}
}

// isSubset checks if every OR-operand in `current` is covered by some OR-operand in `prior`.
// An OR-operand A is covered by OR-operand B if B's AND-set is a subset of A's AND-set.
// (If B ⊆ A as sets of AND-operands, then whenever B is true, A is also true.)
func isSubset(current, prior []operandSet) bool {
	for _, curOp := range current {
		covered := false
		for _, priorOp := range prior {
			if andSetIsSubset(priorOp, curOp) {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}
	return true
}

// andSetIsSubset checks if every element of `sub` is contained in `super`.
// Both slices are sorted.
func andSetIsSubset(sub, super operandSet) bool {
	if len(sub) > len(super) {
		return false
	}
	j := 0
	for _, s := range sub {
		found := false
		for j < len(super) {
			cmp := strings.Compare(s, super[j])
			j++
			if cmp == 0 {
				found = true
				break
			}
			if cmp < 0 {
				// sub element is smaller than current super element, not found
				return false
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// getExpressionTokenSignature produces a canonical string from an expression's tokens.
func getExpressionTokenSignature(sourceFile *ast.SourceFile, expr *ast.Node) string {
	var result strings.Builder
	sourceText := sourceFile.Text()
	first := true
	utils.ForEachToken(expr, func(token *ast.Node) {
		trimmedRange := utils.TrimNodeTextRange(sourceFile, token)
		start := trimmedRange.Pos()
		end := trimmedRange.End()
		if start < end {
			if !first {
				result.WriteByte(' ')
			}
			result.WriteString(sourceText[start:end])
			first = false
		}
	}, sourceFile)
	return result.String()
}
