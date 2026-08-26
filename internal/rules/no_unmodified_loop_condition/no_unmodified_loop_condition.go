package no_unmodified_loop_condition

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_unmodified_loop_condition.schema.json
var schemaJSON []byte

type ruleOptions struct {
	checkConditionalExpressions bool
}

func parseOptions(options []any) ruleOptions {
	if len(options) == 0 {
		return ruleOptions{}
	}
	value, _ := options[0].(map[string]any)
	check, _ := value["checkConditionalExpressions"].(bool)
	return ruleOptions{checkConditionalExpressions: check}
}

func buildLoopConditionNotModifiedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "loopConditionNotModified",
		Description: "'" + name + "' is not modified in this loop.",
	}
}

// hasDynamicExpression checks if an expression contains any dynamic sub-expression
// (call, member access, new, tagged template, yield) that could have side effects.
// Skips traversal into function/class expressions (matching ESLint's SKIP_PATTERN).
func hasDynamicExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsFunctionLike(node) || node.Kind == ast.KindClassExpression {
		return false
	}

	switch node.Kind {
	case ast.KindCallExpression:
		// ESTree represents import() as ImportExpression, which is not one of
		// upstream's dynamic sentinels. tsgo represents it as CallExpression.
		if !ast.IsImportCall(node) {
			return true
		}
	case ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindNewExpression,
		ast.KindTaggedTemplateExpression,
		ast.KindYieldExpression:
		return true
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

// identifierRef holds an identifier's symbol, AST node, and the binary or
// conditional expression group it belongs to. A nil group means ESLint checks
// this reference independently.
type identifierRef struct {
	symbol *ast.Symbol
	node   *ast.Node
	group  *ast.Node
}

// isConditionGroup reports whether node corresponds to one of the ESTree node
// kinds grouped by ESLint: BinaryExpression, plus ConditionalExpression unless
// checkConditionalExpressions is enabled. tsgo also represents logical,
// assignment, and sequence expressions as BinaryExpression, so those operator
// families must be excluded here.
func isConditionGroup(node *ast.Node, checkConditionalExpressions bool) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindConditionalExpression {
		return !checkConditionalExpressions
	}
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := node.AsBinaryExpression()
	if bin == nil || bin.OperatorToken == nil {
		return false
	}
	operator := bin.OperatorToken.Kind
	return operator != ast.KindCommaToken &&
		!ast.IsLogicalOrCoalescingBinaryOperator(operator) &&
		!ast.IsAssignmentOperator(operator)
}

// isConditionBoundary matches the expression/statement/declaration sentinels
// that stop ESLint's ancestor walk. Tagged templates are intentionally not a
// boundary: upstream only treats them as dynamic when they occur inside a
// binary or conditional group.
func isConditionBoundary(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsFunctionLike(node) || node.Kind == ast.KindClassExpression ||
		node.Kind == ast.KindClassDeclaration || ast.IsStatement(node) {
		return true
	}
	switch node.Kind {
	case ast.KindCallExpression:
		return !ast.IsImportCall(node)
	case ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindNewExpression,
		ast.KindYieldExpression:
		return true
	default:
		return false
	}
}

// collectIdentifierSymbols mirrors ESLint's per-reference ancestor walk.
// Symbols are resolved via RefStore.Resolve, which falls back to the
// TypeChecker internally for identifiers its per-file binder walk can't place.
func collectIdentifierSymbols(condition *ast.Node, refs *rule.RefStore, checkConditionalExpressions bool) []identifierRef {
	if condition == nil {
		return nil
	}

	var result []identifierRef
	var walk func(n *ast.Node, group *ast.Node)
	walk = func(n *ast.Node, group *ast.Node) {
		if n == nil {
			return
		}
		if isConditionBoundary(n) {
			return
		}
		// Walking from the condition root means the first eligible ancestor is
		// the outermost group that findConditionGroup used to discover by
		// walking back up from every identifier. Outermost groups have disjoint
		// subtrees, so their dynamic status can be checked once at entry without
		// per-loop maps.
		if group == nil && isConditionGroup(n, checkConditionalExpressions) {
			group = n
			if hasDynamicExpression(group) {
				return
			}
		}
		if n.Kind == ast.KindIdentifier {
			sym := refs.Resolve(n)
			if sym == nil {
				return
			}
			result = append(result, identifierRef{symbol: sym, node: n, group: group})
			return
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child, group)
			return false
		})
	}
	walk(condition, nil)
	return result
}

func nodeIsInRange(node *ast.Node, rangeNode *ast.Node) bool {
	if node == nil || rangeNode == nil {
		return false
	}
	return node.Pos() >= rangeNode.Pos() && node.End() <= rangeNode.End()
}

func nodeIsInLoopRange(node *ast.Node, condition *ast.Node, body *ast.Node, incrementor *ast.Node) bool {
	return nodeIsInRange(node, condition) ||
		nodeIsInRange(node, body) ||
		nodeIsInRange(node, incrementor)
}

// isVarDeclarationWrittenInLoop covers initialization writes that RefStore
// intentionally omits because declaration names are not references. ESLint
// counts these only for variables whose first definition is a var declaration.
func isVarDeclarationWrittenInLoop(sym *ast.Symbol, condition *ast.Node, body *ast.Node, incrementor *ast.Node) bool {
	if sym == nil || len(sym.Declarations) == 0 {
		return false
	}

	first := ast.GetRootDeclaration(sym.Declarations[0])
	if first == nil || first.Kind != ast.KindVariableDeclaration || first.Parent == nil ||
		!utils.IsVarKeyword(first.Parent) {
		return false
	}

	for _, declaration := range sym.Declarations {
		root := ast.GetRootDeclaration(declaration)
		if root == nil || root.Kind != ast.KindVariableDeclaration || root.Parent == nil ||
			!utils.IsVarKeyword(root.Parent) ||
			!nodeIsInLoopRange(root, condition, body, incrementor) {
			continue
		}
		if utils.VariableDeclarationIntroducesWrite(root) {
			return true
		}
	}
	return false
}

// isReferencedInLoop reports whether sym has a reference in any part of the
// loop. The caller always has exactly one function symbol, so accepting it
// directly avoids constructing a single-element slice for every candidate
// modifier.
func isReferencedInLoop(refs *rule.RefStore, sym *ast.Symbol, condition *ast.Node, body *ast.Node, incrementor *ast.Node) bool {
	for _, ref := range refs.References(sym) {
		if nodeIsInLoopRange(ref, condition, body, incrementor) {
			return true
		}
	}
	return false
}

// enclosingFunctionDeclaration returns the nearest FunctionDeclaration that
// contains a write reference. Function expressions do not stop the walk,
// matching ESLint's getEncloseFunctionDeclaration helper.
func enclosingFunctionDeclaration(ref *ast.Node) *ast.Node {
	for node := ref; node != nil; node = node.Parent {
		if ast.IsFunctionDeclaration(node) {
			if node.Name() == nil {
				return nil
			}
			return node
		}
	}
	return nil
}

// isModifiedInLoop scans the symbol's shared whole-file reference list once
// for direct writes before applying ESLint's secondary called-function check.
// The previous implementation scanned the same list independently for the
// condition, body, incrementor, and called-function path.
func isModifiedInLoop(refs *rule.RefStore, sym *ast.Symbol, condition *ast.Node, body *ast.Node, incrementor *ast.Node) bool {
	symbolRefs := refs.References(sym)
	for _, ref := range symbolRefs {
		if utils.IsWriteReference(ref) && nodeIsInLoopRange(ref, condition, body, incrementor) {
			return true
		}
	}
	if isVarDeclarationWrittenInLoop(sym, condition, body, incrementor) {
		return true
	}

	// A write outside the loop still counts when it belongs to a named
	// FunctionDeclaration whose binding is referenced inside the loop.
	for _, modifier := range symbolRefs {
		if !utils.IsWriteReference(modifier) {
			continue
		}
		funcNode := enclosingFunctionDeclaration(modifier)
		if funcNode == nil {
			continue
		}
		funcSym := funcNode.Symbol()
		if funcSym == nil {
			continue
		}
		if isReferencedInLoop(refs, funcSym, condition, body, incrementor) {
			return true
		}
	}
	return false
}

// checkLoopCondition checks identifiers in a loop condition and reports those
// that are not modified in the condition, body, or incrementor.
func checkLoopCondition(ctx rule.RuleContext, condition *ast.Node, body *ast.Node, incrementor *ast.Node, options ruleOptions) {
	if condition == nil || body == nil {
		return
	}

	refs := ctx.Refs
	idRefs := collectIdentifierSymbols(condition, refs, options.checkConditionalExpressions)
	if len(idRefs) == 0 {
		return
	}
	if len(idRefs) == 1 {
		ref := idRefs[0]
		if !isModifiedInLoop(refs, ref.symbol, condition, body, incrementor) {
			ctx.ReportNode(ref.node, buildLoopConditionNotModifiedMessage(ref.node.Text()))
		}
		return
	}

	modifiedBySymbol := map[*ast.Symbol]bool{}

	for groupStart := 0; groupStart < len(idRefs); {
		groupEnd := groupStart + 1
		groupNode := idRefs[groupStart].group
		if groupNode != nil {
			// References assigned to an outermost group form one contiguous DFS
			// run. Nil-group references remain intentionally independent.
			for groupEnd < len(idRefs) && idRefs[groupEnd].group == groupNode {
				groupEnd++
			}
		}
		group := idRefs[groupStart:groupEnd]

		// Check if any identifier in this group is modified
		anyModified := false
		for _, ref := range group {
			modified, checked := modifiedBySymbol[ref.symbol]
			if !checked {
				modified = isModifiedInLoop(refs, ref.symbol, condition, body, incrementor)
				modifiedBySymbol[ref.symbol] = modified
			}
			if modified {
				anyModified = true
				break
			}
		}

		if !anyModified {
			for _, ref := range group {
				ctx.ReportNode(ref.node, buildLoopConditionNotModifiedMessage(ref.node.Text()))
			}
		}
		groupStart = groupEnd
	}
}

// NoUnmodifiedLoopConditionRule disallows variables in loop conditions that are not modified in the loop
var NoUnmodifiedLoopConditionRule = rule.Rule{
	Name:   "no-unmodified-loop-condition",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		return rule.RuleListeners{
			ast.KindWhileStatement: func(node *ast.Node) {
				whileStmt := node.AsWhileStatement()
				if whileStmt == nil {
					return
				}
				checkLoopCondition(ctx, whileStmt.Expression, whileStmt.Statement, nil, opts)
			},
			ast.KindDoStatement: func(node *ast.Node) {
				doStmt := node.AsDoStatement()
				if doStmt == nil {
					return
				}
				checkLoopCondition(ctx, doStmt.Expression, doStmt.Statement, nil, opts)
			},
			ast.KindForStatement: func(node *ast.Node) {
				forStmt := node.AsForStatement()
				if forStmt == nil {
					return
				}
				checkLoopCondition(ctx, forStmt.Condition, forStmt.Statement, forStmt.Incrementor, opts)
			},
		}
	},
}
