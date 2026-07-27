package no_unmodified_loop_condition

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildLoopConditionNotModifiedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "loopConditionNotModified",
		Description: fmt.Sprintf("'%s' is not modified in this loop.", name),
	}
}

// hasDynamicExpression checks if an expression contains any dynamic sub-expression
// (call, member access, new, tagged template, yield) that could have side effects.
// Skips traversal into function/class expressions (matching ESLint's SKIP_PATTERN).
func hasDynamicExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindCallExpression,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindNewExpression,
		ast.KindTaggedTemplateExpression,
		ast.KindYieldExpression:
		return true
	// Skip function/class expressions — side effects inside them
	// don't execute during condition evaluation.
	case ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindClassExpression:
		return false
	}

	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if hasDynamicExpression(child) {
			found = true
			return true
		}
		return false
	})
	return found
}

// identifierRef holds an identifier's symbol and AST node for reporting.
type identifierRef struct {
	symbol *ast.Symbol
	node   *ast.Node
}

// extractGroups walks a condition expression and returns groups of sub-expressions.
// Logical operators (||, &&, ??) split operands into independent groups.
// Comparison/arithmetic BinaryExpressions and ConditionalExpressions form a single group.
func extractGroups(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindBinaryExpression {
		bin := node.AsBinaryExpression()
		if bin != nil && bin.OperatorToken != nil && ast.IsLogicalOrCoalescingBinaryOperator(bin.OperatorToken.Kind) {
			// Logical operators split into independent groups
			left := extractGroups(bin.Left)
			right := extractGroups(bin.Right)
			return append(left, right...)
		}
	}
	// Everything else (comparison binary, conditional, single identifier, etc.)
	// is a single group.
	return []*ast.Node{node}
}

// collectIdentifierSymbols collects unique identifier references (by symbol) from a node.
// Returns nil if any dynamic expression is found.
func collectIdentifierSymbols(node *ast.Node, refs *rule.RefStore) []identifierRef {
	if node == nil {
		return nil
	}
	if hasDynamicExpression(node) {
		return nil
	}
	var result []identifierRef
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		if n.Kind == ast.KindIdentifier {
			sym := refs.Resolve(n)
			if sym != nil {
				// Deduplicate by symbol
				dup := false
				for _, r := range result {
					if r.symbol == sym {
						dup = true
						break
					}
				}
				if !dup {
					result = append(result, identifierRef{symbol: sym, node: n})
				}
			}
			return
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(node)
	return result
}

// isWrittenInRange reports whether sym has a write reference positioned inside
// rangeNode. It scans sym's whole-file reference list (built once per symbol
// and shared across every loop/group that queries it) rather than re-walking
// rangeNode's subtree per query — the old TypeChecker-based implementation
// walked the (potentially large) loop body/incrementor once per condition
// identifier, which dominates cost when loops are large or numerous.
func isWrittenInRange(refs *rule.RefStore, sym *ast.Symbol, rangeNode *ast.Node) bool {
	if rangeNode == nil {
		return false
	}
	lo, hi := rangeNode.Pos(), rangeNode.End()
	for _, ref := range refs.References(sym) {
		if ref.Pos() >= lo && ref.End() <= hi && utils.IsWriteReference(ref) {
			return true
		}
	}
	return false
}

// isReferencedInRange reports whether any symbol in funcSymbols has a
// reference positioned inside rangeNode, using the same whole-file reference
// lists as isWrittenInRange instead of walking rangeNode per query.
func isReferencedInRange(refs *rule.RefStore, funcSymbols []*ast.Symbol, rangeNode *ast.Node) bool {
	if rangeNode == nil {
		return false
	}
	lo, hi := rangeNode.Pos(), rangeNode.End()
	for _, funcSym := range funcSymbols {
		for _, ref := range refs.References(funcSym) {
			if ref.Pos() >= lo && ref.End() <= hi {
				return true
			}
		}
	}
	return false
}

// isModifiedByCalledFunction checks if the symbol is modified inside a
// FunctionDeclaration that is called within the loop (body or incrementor).
// This matches ESLint's secondary check: if a write reference to the variable
// is inside a FunctionDeclaration, and that function's name is referenced
// within the loop, the variable counts as modified.
func isModifiedByCalledFunction(refs *rule.RefStore, loopBody *ast.Node, incrementor *ast.Node, sym *ast.Symbol) bool {
	scope := utils.FindEnclosingScope(loopBody)
	if scope == nil {
		return false
	}

	// Step 1: find (top-level, non-nested) FunctionDeclarations in scope that
	// write to sym anywhere in their body.
	var modifyingFuncSymbols []*ast.Symbol
	var findFuncs func(n *ast.Node)
	findFuncs = func(n *ast.Node) {
		if n == nil {
			return
		}
		if ast.IsFunctionDeclaration(n) && n.Name() != nil {
			if isWrittenInRange(refs, sym, n) {
				if funcSym := n.Symbol(); funcSym != nil {
					modifyingFuncSymbols = append(modifyingFuncSymbols, funcSym)
				}
			}
			return
		}
		n.ForEachChild(func(child *ast.Node) bool {
			findFuncs(child)
			return false
		})
	}
	findFuncs(scope)

	if len(modifyingFuncSymbols) == 0 {
		return false
	}

	// Step 2: check if any of those functions are referenced in the loop
	// (body or incrementor). ESLint uses range-based checking — any reference
	// (not just calls) to the function within the loop counts.
	if isReferencedInRange(refs, modifyingFuncSymbols, loopBody) {
		return true
	}
	if incrementor != nil && isReferencedInRange(refs, modifyingFuncSymbols, incrementor) {
		return true
	}
	return false
}

// checkLoopCondition checks identifiers in a loop condition and reports those
// that are not modified in the loop body (or incrementor for for-statements).
func checkLoopCondition(ctx rule.RuleContext, condition *ast.Node, body *ast.Node, incrementor *ast.Node) {
	if condition == nil || body == nil {
		return
	}

	refs := ctx.Refs
	groups := extractGroups(condition)

	for _, group := range groups {
		idRefs := collectIdentifierSymbols(group, refs)
		if idRefs == nil {
			continue // dynamic expression found, skip this group
		}

		// Check if any identifier in this group is modified
		anyModified := false
		for _, ref := range idRefs {
			if isWrittenInRange(refs, ref.symbol, body) ||
				(incrementor != nil && isWrittenInRange(refs, ref.symbol, incrementor)) ||
				isModifiedByCalledFunction(refs, body, incrementor, ref.symbol) {
				anyModified = true
				break
			}
		}

		if !anyModified {
			for _, ref := range idRefs {
				ctx.ReportNode(ref.node, buildLoopConditionNotModifiedMessage(ref.node.Text()))
			}
		}
	}
}

// NoUnmodifiedLoopConditionRule disallows variables in loop conditions that are not modified in the loop
var NoUnmodifiedLoopConditionRule = rule.Rule{
	Name:             "no-unmodified-loop-condition",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Defense-in-depth: RequiresTypeInfo: true filters this rule out for
		// gap files / inferred-project files, but if a future caller bypasses
		// the filter we still want to no-op rather than nil-deref.
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			ast.KindWhileStatement: func(node *ast.Node) {
				whileStmt := node.AsWhileStatement()
				if whileStmt == nil {
					return
				}
				checkLoopCondition(ctx, whileStmt.Expression, whileStmt.Statement, nil)
			},
			ast.KindDoStatement: func(node *ast.Node) {
				doStmt := node.AsDoStatement()
				if doStmt == nil {
					return
				}
				checkLoopCondition(ctx, doStmt.Expression, doStmt.Statement, nil)
			},
			ast.KindForStatement: func(node *ast.Node) {
				forStmt := node.AsForStatement()
				if forStmt == nil {
					return
				}
				checkLoopCondition(ctx, forStmt.Condition, forStmt.Statement, forStmt.Incrementor)
			},
		}
	},
}
