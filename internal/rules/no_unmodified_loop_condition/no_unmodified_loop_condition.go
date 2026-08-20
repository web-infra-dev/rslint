package no_unmodified_loop_condition

import (
	_ "embed"
	"fmt"

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

// findConditionGroup returns the outermost ESLint-compatible binary or
// conditional expression between an identifier and the condition root.
func findConditionGroup(identifier *ast.Node, condition *ast.Node, checkConditionalExpressions bool) (*ast.Node, bool) {
	var group *ast.Node
	for current := identifier; current != nil; current = current.Parent {
		if isConditionGroup(current, checkConditionalExpressions) {
			group = current
		}
		if current == condition {
			return group, true
		}
	}
	return nil, false
}

// collectIdentifierSymbols mirrors ESLint's per-reference ancestor walk.
// Symbols are resolved via RefStore.Resolve, which falls back to the
// TypeChecker internally for identifiers its per-file binder walk can't place.
func collectIdentifierSymbols(condition *ast.Node, refs *rule.RefStore, checkConditionalExpressions bool) []identifierRef {
	if condition == nil {
		return nil
	}

	dynamicGroups := map[*ast.Node]bool{}
	checkedDynamicGroups := map[*ast.Node]bool{}
	result := []identifierRef{}
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		if isConditionBoundary(n) {
			return
		}
		if n.Kind == ast.KindIdentifier {
			sym := refs.Resolve(n)
			if sym == nil {
				return
			}
			group, ok := findConditionGroup(n, condition, checkConditionalExpressions)
			if !ok {
				return
			}
			if group != nil {
				if !checkedDynamicGroups[group] {
					checkedDynamicGroups[group] = true
					dynamicGroups[group] = hasDynamicExpression(group)
				}
				if dynamicGroups[group] {
					return
				}
			}
			result = append(result, identifierRef{symbol: sym, node: n, group: group})
			return
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(condition)
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

// isVarDeclarationWrittenInRange covers initialization writes that RefStore
// intentionally omits because declaration names are not references. ESLint
// counts these only for variables whose first definition is a var declaration.
func isVarDeclarationWrittenInRange(sym *ast.Symbol, rangeNode *ast.Node) bool {
	if sym == nil || rangeNode == nil || len(sym.Declarations) == 0 {
		return false
	}

	first := ast.GetRootDeclaration(sym.Declarations[0])
	if first == nil || first.Kind != ast.KindVariableDeclaration || first.Parent == nil ||
		!utils.IsVarKeyword(first.Parent) {
		return false
	}

	lo, hi := rangeNode.Pos(), rangeNode.End()
	for _, declaration := range sym.Declarations {
		root := ast.GetRootDeclaration(declaration)
		if root == nil || root.Kind != ast.KindVariableDeclaration || root.Parent == nil ||
			!utils.IsVarKeyword(root.Parent) ||
			root.Pos() < lo || root.End() > hi {
			continue
		}
		if utils.VariableDeclarationIntroducesWrite(root) {
			return true
		}
	}
	return false
}

func isModifiedInRange(refs *rule.RefStore, sym *ast.Symbol, rangeNode *ast.Node) bool {
	return isWrittenInRange(refs, sym, rangeNode) ||
		isVarDeclarationWrittenInRange(sym, rangeNode)
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

// isModifiedByCalledFunction checks if the symbol is modified inside a
// FunctionDeclaration that is referenced within the loop condition, body, or
// incrementor.
// This matches ESLint's secondary check: if a write reference to the variable
// is inside a FunctionDeclaration, and that function's name is referenced
// within the loop, the variable counts as modified.
func isModifiedByCalledFunction(refs *rule.RefStore, condition *ast.Node, loopBody *ast.Node, incrementor *ast.Node, sym *ast.Symbol) bool {
	for _, modifier := range refs.References(sym) {
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
		modifyingFuncSymbols := []*ast.Symbol{funcSym}
		if isReferencedInRange(refs, modifyingFuncSymbols, condition) ||
			isReferencedInRange(refs, modifyingFuncSymbols, loopBody) ||
			(incrementor != nil && isReferencedInRange(refs, modifyingFuncSymbols, incrementor)) {
			return true
		}
	}
	return false
}

type identifierGroup struct {
	refs []identifierRef
}

// groupIdentifierRefs keeps each ungrouped reference independent and combines
// only references that share the exact binary/conditional group node.
func groupIdentifierRefs(refs []identifierRef) []identifierGroup {
	groups := []identifierGroup{}
	indexes := map[*ast.Node]int{}
	for _, ref := range refs {
		if ref.group == nil {
			groups = append(groups, identifierGroup{refs: []identifierRef{ref}})
			continue
		}
		if index, ok := indexes[ref.group]; ok {
			groups[index].refs = append(groups[index].refs, ref)
			continue
		}
		indexes[ref.group] = len(groups)
		groups = append(groups, identifierGroup{refs: []identifierRef{ref}})
	}
	return groups
}

// checkLoopCondition checks identifiers in a loop condition and reports those
// that are not modified in the condition, body, or incrementor.
func checkLoopCondition(ctx rule.RuleContext, condition *ast.Node, body *ast.Node, incrementor *ast.Node, options ruleOptions) {
	if condition == nil || body == nil {
		return
	}

	refs := ctx.Refs
	idRefs := collectIdentifierSymbols(condition, refs, options.checkConditionalExpressions)
	groups := groupIdentifierRefs(idRefs)
	modifiedBySymbol := map[*ast.Symbol]bool{}
	checkedSymbol := map[*ast.Symbol]bool{}

	for _, group := range groups {
		// Check if any identifier in this group is modified
		anyModified := false
		for _, ref := range group.refs {
			modified := modifiedBySymbol[ref.symbol]
			if !checkedSymbol[ref.symbol] {
				modified = isModifiedInRange(refs, ref.symbol, condition) ||
					isModifiedInRange(refs, ref.symbol, body) ||
					(incrementor != nil && isModifiedInRange(refs, ref.symbol, incrementor)) ||
					isModifiedByCalledFunction(refs, condition, body, incrementor, ref.symbol)
				checkedSymbol[ref.symbol] = true
				modifiedBySymbol[ref.symbol] = modified
			}
			if modified {
				anyModified = true
				break
			}
		}

		if !anyModified {
			for _, ref := range group.refs {
				ctx.ReportNode(ref.node, buildLoopConditionNotModifiedMessage(ref.node.Text()))
			}
		}
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
