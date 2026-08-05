package no_useless_return

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/cfg"
)

var unnecessaryReturn = rule.RuleMessage{
	Id:          "unnecessaryReturn",
	Description: "Unnecessary return statement.",
}

// https://eslint.org/docs/latest/rules/no-useless-return
var NoUselessReturnRule = rule.Rule{
	Name:   "no-useless-return",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var byRoot map[*ast.Node][]*ast.Node
		var rootOrder []*ast.Node

		listeners := rule.RuleListeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				if !isBareReturn(node) {
					return
				}
				root := cfg.RootOf(node)
				if root == nil {
					return
				}
				if byRoot == nil {
					byRoot = make(map[*ast.Node][]*ast.Node)
				}
				if _, seen := byRoot[root]; !seen {
					rootOrder = append(rootOrder, root)
				}
				byRoot[root] = append(byRoot[root], node)
			},
		}

		statements := ctx.SourceFile.Statements
		if statements == nil || len(statements.Nodes) == 0 {
			return listeners
		}
		lastTopLevelNode := statements.Nodes[len(statements.Nodes)-1]
		listeners[rule.ListenerOnExit(lastTopLevelNode.Kind)] = func(node *ast.Node) {
			if node == lastTopLevelNode {
				reportUselessReturns(&ctx, byRoot, rootOrder)
			}
		}
		return listeners
	},
}

// reportUselessReturns keeps the returns nothing runs after. ESLint reports one
// code path at a time and its reports come out ordered by position, so the
// collected returns are sorted before they are reported.
func reportUselessReturns(ctx *rule.RuleContext, byRoot map[*ast.Node][]*ast.Node, rootOrder []*ast.Node) {
	var reports []*ast.Node
	for _, root := range rootOrder {
		reachable := reachableStatements(root)
		for _, returnStatement := range byRoot[root] {
			// A `return` nothing reaches says nothing about the code after it.
			if !reachable[returnStatement] {
				continue
			}
			walker := &walker{root: root, reachable: reachable, visited: map[*ast.Node]bool{}}
			if !walker.runsAfter(returnStatement) {
				reports = append(reports, returnStatement)
			}
		}
	}
	slices.SortFunc(reports, func(a, b *ast.Node) int { return a.Pos() - b.Pos() })
	for _, returnStatement := range reports {
		node := returnStatement
		ctx.ReportNodeWithDeferredFixes(node, unnecessaryReturn, func() []rule.RuleFix {
			return buildFix(ctx.SourceFile, node)
		})
	}
}

// reachableStatements lays out one code path and collects the statements
// control can reach. internal/utils/cfg stops laying a path out where it ends, so a
// statement the hook never sees is one ESLint's code path analysis would place
// on an unreachable segment.
func reachableStatements(root *ast.Node) map[*ast.Node]bool {
	reachable := make(map[*ast.Node]bool)
	cfg.Build(root, cfg.Hooks[struct{}]{
		Statement: func(_ *cfg.Builder[struct{}], node *ast.Node) {
			reachable[node] = true
		},
	})
	return reachable
}

// ---------------------------------------------------------------------------
// candidates
// ---------------------------------------------------------------------------

// isBareReturn reports whether node is a `return` with no value that ESLint
// would consider redundant if nothing ran after it. A `return` inside a loop
// leaves the loop, and one inside a `finally` block overrides the value the
// statement was leaving with, so neither is ever redundant.
func isBareReturn(node *ast.Node) bool {
	return node.AsReturnStatement().Expression == nil && !isInLoop(node) && !isInFinally(node)
}

func isInLoop(node *ast.Node) bool {
	for current := node; current != nil && !isFunction(current); current = current.Parent {
		switch current.Kind {
		case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement,
			ast.KindForInStatement, ast.KindForOfStatement:
			return true
		}
	}
	return false
}

func isInFinally(node *ast.Node) bool {
	for current := node; current != nil && current.Parent != nil && !isFunction(current); current = current.Parent {
		if current.Parent.Kind == ast.KindTryStatement &&
			current.Parent.AsTryStatement().FinallyBlock == current {
			return true
		}
	}
	return false
}

// isFunction covers the nodes ESLint's ESTree calls a function: a declaration,
// an arrow, and the function expression a method, constructor, or accessor is
// built from.
func isFunction(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// what runs after a return
// ---------------------------------------------------------------------------

// step is what scanning one statement says about the statements after it.
type step int

const (
	// stepContinue: nothing ESLint counts as executed happened here.
	stepContinue step = iota
	// stepRuns: a statement runs after the return, so the return is used.
	stepRuns
	// stepStop: control does not carry the return past this statement.
	stepStop
)

// walker answers "does anything run after this return" by following the code
// path the return would have if it were deleted.
//
// ESLint answers the same question by attaching the return to its code path
// segment and letting any statement on a segment the return can reach clear it
// again — deliberately walking through the unreachable segments the return
// itself created, which simulates the path the code would take without it.
// internal/utils/cfg stops laying a path out at a `return`, so that continuation has
// no edges in the graph and is followed over the syntax tree instead: from the
// return outwards through its enclosing statements, into the statements that
// follow each of them.
type walker struct {
	root      *ast.Node
	reachable map[*ast.Node]bool
	visited   map[*ast.Node]bool
}

// runsAfter reports whether a statement runs after node completes.
func (w *walker) runsAfter(node *ast.Node) bool {
	if w.visited[node] {
		return false
	}
	w.visited[node] = true

	for node != nil && node != w.root {
		parent := node.Parent
		if parent == nil {
			return false
		}

		switch parent.Kind {
		case ast.KindSourceFile:
			runs, _ := w.scanFrom(parent.AsSourceFile().Statements, node)
			return runs

		case ast.KindBlock:
			runs, stopped := w.scanFrom(parent.AsBlock().Statements, node)
			if runs || stopped {
				return runs
			}

		case ast.KindModuleBlock:
			runs, stopped := w.scanFrom(parent.AsModuleBlock().Statements, node)
			if runs || stopped {
				return runs
			}

		case ast.KindCaseClause, ast.KindDefaultClause:
			runs, stopped := w.scanFrom(parent.AsCaseOrDefaultClause().Statements, node)
			if runs || stopped {
				return runs
			}
			// Control falls out of a clause into the one after it.
			caseBlock := parent.Parent
			if caseBlock == nil || caseBlock.Kind != ast.KindCaseBlock {
				return false
			}
			clauses := caseBlock.AsCaseBlock().Clauses.Nodes
			index := slices.Index(clauses, parent)
			if index < 0 {
				return false
			}
			for _, later := range clauses[index+1:] {
				runs, stopped := w.scanList(later.AsCaseOrDefaultClause().Statements, 0)
				if runs || stopped {
					return runs
				}
			}
			node = caseBlock.Parent
			continue

		case ast.KindTryStatement:
			try := parent.AsTryStatement()
			// A return inside the `try` block keeps its place in the list
			// while the `catch` clause and the `finally` block are traversed,
			// so nothing in either of them clears it. A return in the `catch`
			// clause is past that point, and the `finally` block does clear it.
			if node == try.CatchClause && try.FinallyBlock != nil {
				runs, stopped := w.scanList(try.FinallyBlock.AsBlock().Statements, 0)
				if runs || stopped {
					return runs
				}
			}

		case ast.KindIfStatement, ast.KindLabeledStatement, ast.KindWithStatement,
			ast.KindCatchClause:
			// Nothing of these runs after the branch the return is in.

		default:
			// A function body, a class static block, or a loop the return
			// leaves: the code path ends here.
			return false
		}

		node = parent
	}
	return false
}

func (w *walker) scanFrom(list *ast.NodeList, after *ast.Node) (runs bool, stopped bool) {
	if list == nil {
		return false, false
	}
	index := slices.Index(list.Nodes, after)
	if index < 0 {
		return false, false
	}
	return w.scanList(list, index+1)
}

func (w *walker) scanList(list *ast.NodeList, start int) (runs bool, stopped bool) {
	if list == nil || start < 0 || start > len(list.Nodes) {
		return false, false
	}
	for _, statement := range list.Nodes[start:] {
		switch w.scan(statement) {
		case stepRuns:
			return true, false
		case stepStop:
			return false, true
		}
	}
	return false, false
}

// scan reports what one statement standing after the return means for it.
//
// ESLint clears a return on every statement node except `FunctionDeclaration`,
// `BlockStatement`, and `BreakStatement` — a function declaration is hoisted
// and runs either way, a block only contains statements, and a break merely
// goes to the next segment. Nothing clears a return for the TypeScript-only
// declarations either, because ESLint's listener list names ESTree types and
// those parse to types of their own; an `export` modifier puts one back inside
// an `ExportNamedDeclaration`, which does clear it.
func (w *walker) scan(statement *ast.Node) step {
	switch statement.Kind {
	case ast.KindExportAssignment:
		// `export = x` is TypeScript's own; `export default x` is not.
		if statement.AsExportAssignment().IsExportEquals {
			return stepContinue
		}
		return stepRuns

	case ast.KindFunctionDeclaration, ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration,
		ast.KindImportEqualsDeclaration, ast.KindMissingDeclaration:
		if hasExportModifier(statement) {
			return stepRuns
		}
		return stepContinue

	case ast.KindModuleDeclaration:
		if hasExportModifier(statement) {
			return stepRuns
		}
		body := statement.AsModuleDeclaration().Body
		for body != nil && body.Kind == ast.KindModuleDeclaration {
			body = body.AsModuleDeclaration().Body
		}
		if body == nil || body.Kind != ast.KindModuleBlock {
			return stepContinue
		}
		runs, stopped := w.scanList(body.AsModuleBlock().Statements, 0)
		if runs {
			return stepRuns
		}
		if stopped {
			return stepStop
		}
		return stepContinue

	case ast.KindBlock:
		runs, stopped := w.scanList(statement.AsBlock().Statements, 0)
		if runs {
			return stepRuns
		}
		if stopped {
			return stepStop
		}
		return stepContinue

	case ast.KindReturnStatement:
		if statement.AsReturnStatement().Expression != nil {
			return stepRuns
		}
		// A `return` of its own carries the earlier one along, on the same
		// "as if it were deleted" reading — unless this one is a return the
		// rule never collects, which leaves nothing to carry it.
		if w.reachable[statement] && !isBareReturn(statement) {
			return stepStop
		}
		return stepContinue

	case ast.KindBreakStatement:
		if w.runsAfterBreak(statement, breakTarget(statement)) {
			return stepRuns
		}
		if w.reachable[statement] {
			return stepStop
		}
		// A break control never reaches goes to the next segment merely, so
		// the return carries on past it.
		return stepContinue

	default:
		return stepRuns
	}
}

// runsAfterBreak follows a break out to the statement it leaves. On the way out
// every `finally` block still runs, except where the return is inside the `try`
// block that owns it and keeps its place in the list. A nil target is a break
// with nothing to leave, which ends the code path where it stands.
func (w *walker) runsAfterBreak(node *ast.Node, target *ast.Node) bool {
	for current := node; current != nil && current != target && current != w.root; current = current.Parent {
		parent := current.Parent
		if parent == nil {
			break
		}
		if parent.Kind != ast.KindTryStatement {
			continue
		}
		try := parent.AsTryStatement()
		if try.FinallyBlock == nil || current != try.CatchClause {
			continue
		}
		runs, stopped := w.scanList(try.FinallyBlock.AsBlock().Statements, 0)
		if runs {
			return true
		}
		if stopped {
			return false
		}
	}
	if target == nil {
		return false
	}
	return w.runsAfter(target)
}

func hasExportModifier(node *ast.Node) bool {
	return ast.HasSyntacticModifier(node, ast.ModifierFlagsExport)
}

// breakTarget returns the statement a `break` leaves, so the walk can carry on
// from after it.
func breakTarget(node *ast.Node) *ast.Node {
	label := node.AsBreakStatement().Label
	name := ""
	if label != nil {
		name = label.Text()
	}
	for current := node.Parent; current != nil; current = current.Parent {
		if isFunction(current) || current.Kind == ast.KindClassStaticBlockDeclaration ||
			current.Kind == ast.KindSourceFile {
			return nil
		}
		if name == "" {
			switch current.Kind {
			case ast.KindSwitchStatement, ast.KindWhileStatement, ast.KindDoStatement,
				ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
				return current
			}
			continue
		}
		if current.Kind == ast.KindLabeledStatement &&
			current.AsLabeledStatement().Label.Text() == name {
			return current
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// fix
// ---------------------------------------------------------------------------

// buildFix removes the return statement. The edit spans the whole enclosing
// function even though only the return is dropped from it, which is how ESLint
// keeps this fix from being applied in the same pass as an overlapping one.
func buildFix(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	if !isRemovable(node) || utils.HasCommentInsideNode(sourceFile, node) {
		return nil
	}
	removed := utils.TrimNodeTextRange(sourceFile, node)
	retained := retainedRange(sourceFile, node)
	text := sourceFile.Text()
	return []rule.RuleFix{rule.RuleFixReplaceRange(
		retained,
		text[retained.Pos():removed.Pos()]+text[removed.End():retained.End()],
	)}
}

// isRemovable reports whether the return stands in a list of statements, so
// dropping it leaves the code around it well formed.
func isRemovable(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindCaseClause, ast.KindDefaultClause:
		return true
	}
	return false
}

func retainedRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	for current := node; current != nil; current = current.Parent {
		if isFunction(current) {
			return utils.TrimNodeTextRange(sourceFile, current)
		}
	}
	return core.NewTextRange(0, len(sourceFile.Text()))
}
