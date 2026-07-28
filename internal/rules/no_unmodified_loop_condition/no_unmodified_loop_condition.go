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
	symbol     *ast.Symbol
	node       *ast.Node
	viaChecker bool
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
//
// Symbols are resolved via RefStore first; RefStore only knows about symbols
// bound within this file, so identifiers naming a cross-file, .d.ts, or
// standard-library global fall back to the TypeChecker. Those are flagged
// viaChecker so later modification checks use the matching (slower but
// program-wide) TypeChecker-based path instead of RefStore's per-file
// reference lists, which never index such symbols.
func collectIdentifierSymbols(node *ast.Node, refs *rule.RefStore, tc *checker.Checker) []identifierRef {
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
			viaChecker := false
			if sym == nil {
				sym = tc.GetSymbolAtLocation(n)
				viaChecker = true
			}
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
					result = append(result, identifierRef{symbol: sym, node: n, viaChecker: viaChecker})
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
// reference positioned inside rangeNode.
//
// This resolves each identifier directly via refs.Resolve rather than going
// through RefStore.References(funcSym)'s name-keyed cache: RefStore.Resolve
// maps a same-file self-reference to a default-exported function's name
// (`inc` in `export default function inc() {} ...  inc();`) to the export
// symbol, but that export symbol's own .Name is "default" (its slot in the
// module's export table), not "inc". Resolve() special-cases that name
// mismatch internally, but querying RefStore.References(exportSymbol)
// buckets by exportSymbol.Name = "default" and so never finds the "inc"
// identifiers. A direct per-identifier Resolve() walk sidesteps that
// name/symbol split instead of relying on the cache bucketing correctly.
func isReferencedInRange(refs *rule.RefStore, funcSymbols []*ast.Symbol, rangeNode *ast.Node) bool {
	if rangeNode == nil {
		return false
	}
	found := false
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}
		if n.Kind == ast.KindIdentifier {
			if sym := refs.Resolve(n); sym != nil {
				for _, s := range funcSymbols {
					if sym == s {
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
	walk(rangeNode)
	return found
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
			// Only the body executes when the function is called; a write inside
			// a parameter's default-value initializer only runs when that
			// argument is omitted, so it must not count as "written" here.
			if isWrittenInRange(refs, sym, n.Body()) {
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

// isSymbolWrittenInBodyTC walks the body (and optionally the incrementor) looking for
// any write reference to the given symbol. Does NOT skip function boundaries —
// ESLint uses range-based checking where any write within the loop's text range
// counts as a modification, even inside nested functions.
//
// Used as the fallback for symbols RefStore couldn't resolve (cross-file,
// .d.ts, or standard-library globals): RefStore's reference lists only index
// symbols bound within this file, so those symbols must be found by walking
// the range directly and asking the TypeChecker about each identifier.
func isSymbolWrittenInBodyTC(body *ast.Node, sym *ast.Symbol, tc *checker.Checker) bool {
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

// isModifiedByCalledFunctionTC is the TypeChecker-based counterpart of
// isModifiedByCalledFunction, used when sym came from the TypeChecker
// fallback in collectIdentifierSymbols.
func isModifiedByCalledFunctionTC(loopBody *ast.Node, incrementor *ast.Node, sym *ast.Symbol, tc *checker.Checker) bool {
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
			if functionBodyWritesSymbolTC(n, sym, tc) {
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
	if nodeReferencesAnySymbolTC(loopBody, modifyingFuncSymbols, tc) {
		return true
	}
	if incrementor != nil && nodeReferencesAnySymbolTC(incrementor, modifyingFuncSymbols, tc) {
		return true
	}
	return false
}

// functionBodyWritesSymbolTC checks if a function body contains a write to sym.
// Does NOT skip nested functions — ESLint uses range-based scope analysis.
func functionBodyWritesSymbolTC(funcNode *ast.Node, sym *ast.Symbol, tc *checker.Checker) bool {
	body := funcNode.Body()
	if body == nil {
		return false
	}
	return isSymbolWrittenInBodyTC(body, sym, tc)
}

// nodeReferencesAnySymbolTC checks if a node tree contains any identifier
// referencing one of the given symbols. Does not skip function boundaries
// (ESLint uses range-based checking).
func nodeReferencesAnySymbolTC(node *ast.Node, symbols []*ast.Symbol, tc *checker.Checker) bool {
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

	refs := ctx.Refs
	tc := ctx.TypeChecker
	groups := extractGroups(condition)

	for _, group := range groups {
		idRefs := collectIdentifierSymbols(group, refs, tc)
		if idRefs == nil {
			continue // dynamic expression found, skip this group
		}

		// Check if any identifier in this group is modified
		anyModified := false
		for _, ref := range idRefs {
			var modified bool
			if ref.viaChecker {
				modified = isSymbolWrittenInBodyTC(body, ref.symbol, tc) ||
					(incrementor != nil && isSymbolWrittenInBodyTC(incrementor, ref.symbol, tc)) ||
					isModifiedByCalledFunctionTC(body, incrementor, ref.symbol, tc)
			} else {
				modified = isWrittenInRange(refs, ref.symbol, body) ||
					(incrementor != nil && isWrittenInRange(refs, ref.symbol, incrementor)) ||
					isModifiedByCalledFunction(refs, body, incrementor, ref.symbol)
			}
			if modified {
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
