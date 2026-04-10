package no_unmodified_loop_condition

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
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
func collectIdentifierSymbols(node *ast.Node, tc *checker.Checker) []identifierRef {
	if node == nil {
		return nil
	}
	if hasDynamicExpression(node) {
		return nil
	}
	var refs []identifierRef
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		if n.Kind == ast.KindIdentifier {
			sym := tc.GetSymbolAtLocation(n)
			if sym != nil {
				// Deduplicate by symbol
				dup := false
				for _, r := range refs {
					if r.symbol == sym {
						dup = true
						break
					}
				}
				if !dup {
					refs = append(refs, identifierRef{symbol: sym, node: n})
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
	return refs
}

// isSymbolWrittenInBody walks the body (and optionally the incrementor) looking for
// any write reference to the given symbol. Does NOT skip function boundaries —
// ESLint uses range-based checking where any write within the loop's text range
// counts as a modification, even inside nested functions.
func isSymbolWrittenInBody(body *ast.Node, sym *ast.Symbol, tc *checker.Checker) bool {
	if body == nil {
		return false
	}

	found := false
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}
		if n.Kind == ast.KindIdentifier {
			refSym := tc.GetSymbolAtLocation(n)
			if refSym == sym && utils.IsWriteReference(n) {
				found = true
				return
			}
		}
		// ShorthandPropertyAssignment in destructuring: ({x} = {x: 1})
		// TypeChecker resolves shorthand name to property symbol, not variable symbol.
		// Use GetShorthandAssignmentValueSymbol to get the variable symbol.
		if n.Kind == ast.KindShorthandPropertyAssignment && utils.IsInDestructuringAssignment(n) {
			valSym := tc.GetShorthandAssignmentValueSymbol(n)
			if valSym == sym {
				found = true
				return
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(body)
	return found
}

// isModifiedByCalledFunction checks if the symbol is modified inside a
// FunctionDeclaration that is called within the loop (body or incrementor).
// This matches ESLint's secondary check: if a write reference to the variable
// is inside a FunctionDeclaration, and that function's name is referenced
// within the loop, the variable counts as modified.
func isModifiedByCalledFunction(loopBody *ast.Node, incrementor *ast.Node, sym *ast.Symbol, tc *checker.Checker) bool {
	scope := utils.FindEnclosingScope(loopBody)
	if scope == nil {
		return false
	}

	// Step 1: find FunctionDeclarations that write to sym anywhere in scope.
	var modifyingFuncSymbols []*ast.Symbol
	var findFuncs func(n *ast.Node)
	findFuncs = func(n *ast.Node) {
		if n == nil {
			return
		}
		if ast.IsFunctionDeclaration(n) && n.Name() != nil {
			if functionBodyWritesSymbol(n, sym, tc) {
				funcSym := tc.GetSymbolAtLocation(n.Name())
				if funcSym != nil {
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
	if nodeReferencesAnySymbol(loopBody, modifyingFuncSymbols, tc) {
		return true
	}
	if incrementor != nil && nodeReferencesAnySymbol(incrementor, modifyingFuncSymbols, tc) {
		return true
	}
	return false
}

// functionBodyWritesSymbol checks if a function body contains a write to sym.
// Does NOT skip nested functions — ESLint uses range-based scope analysis.
func functionBodyWritesSymbol(funcNode *ast.Node, sym *ast.Symbol, tc *checker.Checker) bool {
	body := funcNode.Body()
	if body == nil {
		return false
	}
	return isSymbolWrittenInBody(body, sym, tc)
}

// nodeReferencesAnySymbol checks if a node tree contains any identifier
// referencing one of the given symbols. Does not skip function boundaries
// (ESLint uses range-based checking).
func nodeReferencesAnySymbol(node *ast.Node, symbols []*ast.Symbol, tc *checker.Checker) bool {
	if node == nil {
		return false
	}
	found := false
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}
		if n.Kind == ast.KindIdentifier {
			refSym := tc.GetSymbolAtLocation(n)
			if refSym != nil {
				for _, s := range symbols {
					if refSym == s {
						found = true
						return
					}
				}
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(node)
	return found
}

// checkLoopCondition checks identifiers in a loop condition and reports those
// that are not modified in the loop body (or incrementor for for-statements).
func checkLoopCondition(ctx rule.RuleContext, condition *ast.Node, body *ast.Node, incrementor *ast.Node) {
	if condition == nil || body == nil {
		return
	}

	tc := ctx.TypeChecker
	groups := extractGroups(condition)

	for _, group := range groups {
		refs := collectIdentifierSymbols(group, tc)
		if refs == nil {
			continue // dynamic expression found, skip this group
		}

		// Check if any identifier in this group is modified
		anyModified := false
		for _, ref := range refs {
			if isSymbolWrittenInBody(body, ref.symbol, tc) ||
				(incrementor != nil && isSymbolWrittenInBody(incrementor, ref.symbol, tc)) ||
				isModifiedByCalledFunction(body, incrementor, ref.symbol, tc) {
				anyModified = true
				break
			}
		}

		if !anyModified {
			for _, ref := range refs {
				ctx.ReportNode(ref.node, buildLoopConditionNotModifiedMessage(ref.node.Text()))
			}
		}
	}
}

// NoUnmodifiedLoopConditionRule disallows variables in loop conditions that are not modified in the loop
var NoUnmodifiedLoopConditionRule = rule.Rule{
	Name: "no-unmodified-loop-condition",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
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
