package no_useless_return

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"

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
		// Only the code paths that hold a bare return need a graph.
		var rootOrder []*ast.Node
		seen := make(map[*ast.Node]bool)

		listeners := rule.RuleListeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				if node.AsReturnStatement().Expression != nil {
					return
				}
				root := cfg.RootOf(node)
				if root == nil || seen[root] {
					return
				}
				seen[root] = true
				rootOrder = append(rootOrder, root)
			},
		}

		statements := ctx.SourceFile.Statements
		if statements == nil || len(statements.Nodes) == 0 {
			return listeners
		}
		lastTopLevelNode := statements.Nodes[len(statements.Nodes)-1]
		listeners[rule.ListenerOnExit(lastTopLevelNode.Kind)] = func(node *ast.Node) {
			if node == lastTopLevelNode {
				reportUselessReturns(&ctx, rootOrder)
			}
		}
		return listeners
	},
}

// reportUselessReturns lays out each collected code path and reports the returns
// nothing runs after. ESLint reports one code path at a time and its reports
// come out ordered by position, so the collected returns are sorted first.
func reportUselessReturns(ctx *rule.RuleContext, rootOrder []*ast.Node) {
	var reports []*ast.Node
	for _, root := range rootOrder {
		reports = append(reports, uselessReturns(root)...)
	}
	slices.SortFunc(reports, func(a, b *ast.Node) int { return a.Pos() - b.Pos() })
	for _, returnStatement := range reports {
		node := returnStatement
		ctx.ReportNodeWithDeferredFixes(node, unnecessaryReturn, func() []rule.RuleFix {
			return buildFix(ctx.SourceFile, node)
		})
	}
}

// ---------------------------------------------------------------------------
// what runs after a return
// ---------------------------------------------------------------------------

// event is what one statement means for the bare returns standing where it is
// reached.
type eventKind uint8

const (
	// eventReturn is a bare return, held until something clears it.
	eventReturn eventKind = iota
	// eventClear is a statement that stands where those returns do, which
	// gives every one of them something it skips.
	eventClear
)

type event struct {
	kind eventKind
	node *ast.Node
}

type block = cfg.Block[event]

// uselessReturns returns the bare returns in root that nothing runs after.
//
// ESLint holds each bare return against the code path segment it stands on and
// lets any statement on a segment that segment leads to clear it again —
// deliberately reaching into the unreachable segments the return itself
// created, which is the path the code would take without it. The same question
// over a control-flow graph is whether any clearing statement stands in a block
// the return's block leads to.
func uselessReturns(root *ast.Node) []*ast.Node {
	graph := cfg.Build(root, cfg.Hooks[event]{
		Statement: func(b *cfg.Builder[event], node *ast.Node) {
			if node.Kind != ast.KindReturnStatement {
				if clears(node) {
					b.Emit(event{kind: eventClear, node: node})
				}
				return
			}
			if node.AsReturnStatement().Expression != nil {
				// The value it leaves with is what the earlier returns skip.
				b.Emit(event{kind: eventClear, node: node})
				return
			}
			// A `return` inside a loop leaves the loop, and one inside a
			// `finally` block overrides the value the statement was leaving
			// with, so neither is ever redundant. One control never reaches
			// says nothing about the code around it.
			if b.Current().Reachable && !isInLoop(node) && !isInFinally(node) {
				b.Emit(event{kind: eventReturn, node: node})
			}
		},
	})

	w := walk{
		root:        root,
		leftThrough: make([]bool, len(graph.Blocks)),
		seen:        make([]int, len(graph.Blocks)),
	}
	// A block that holds a bare return is one control left through, which is
	// what lets the walk step from it into the code that no longer runs.
	for _, blk := range graph.Blocks {
		for _, e := range blk.Events {
			if e.kind == eventReturn {
				w.leftThrough[blk.Index()] = true
				break
			}
		}
	}

	var reports []*ast.Node
	for _, blk := range graph.Blocks {
		for i, e := range blk.Events {
			if e.kind == eventReturn && !w.runsAfter(blk, i) {
				reports = append(reports, e.node)
			}
		}
	}
	return reports
}

// walk answers one graph's returns. Its scratch is kept across them: seen holds
// the round each block was last reached in, so a round costs no clearing pass
// of its own, and frontier is refilled rather than reallocated.
type walk struct {
	root        *ast.Node
	leftThrough []bool
	seen        []int
	round       int
	frontier    []*block
}

// runsAfter reports whether a clearing statement stands anywhere the return
// recorded at start.Events[index] leads to.
func (w *walk) runsAfter(start *block, index int) bool {
	shields := shieldsOf(start.Events[index].node, w.root)
	if clearedBy(start.Events[index+1:], shields) {
		return true
	}

	w.round++
	w.seen[start.Index()] = w.round
	w.frontier = w.frontier[:0]
	w.step(start)
	for len(w.frontier) > 0 {
		blk := w.frontier[len(w.frontier)-1]
		w.frontier = w.frontier[:len(w.frontier)-1]
		if w.seen[blk.Index()] == w.round {
			continue
		}
		w.seen[blk.Index()] = w.round
		if clearedBy(blk.Events, shields) {
			return true
		}
		w.step(blk)
	}
	return false
}

// step queues the blocks the returns standing at blk still stand at. Code
// control reaches carries them along as usual; the code after an abrupt exit
// carries them only where that exit was a return, so a `break` does not hand
// them on to the statements it skips.
func (w *walk) step(blk *block) {
	carries := !blk.Reachable || w.leftThrough[blk.Index()]
	for _, next := range blk.Successors {
		if next.Reachable || carries {
			w.frontier = append(w.frontier, next)
		}
	}
}

func clearedBy(events []event, shields []shield) bool {
	for _, e := range events {
		if e.kind != eventClear {
			continue
		}
		if !slices.ContainsFunc(shields, func(s shield) bool { return s.covers(e.node) }) {
			return true
		}
	}
	return false
}

// clears reports whether a statement standing where a bare return is pending
// means something runs after that return.
//
// ESLint clears them on every statement node except `FunctionDeclaration`,
// `BlockStatement`, and `BreakStatement` — a function declaration is hoisted
// and runs either way, a block only holds statements of its own, and a break
// merely goes to the next segment. Nothing clears them for the TypeScript-only
// declarations either, because ESLint's listener list names ESTree types and
// those parse to types of their own; an `export` modifier puts one back inside
// an `ExportNamedDeclaration`, which does clear.
func clears(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindBlock, ast.KindBreakStatement:
		return false

	case ast.KindExportAssignment:
		// `export = x` is TypeScript's own; `export default x` is not.
		return !node.AsExportAssignment().IsExportEquals

	case ast.KindNamespaceExportDeclaration:
		// `export as namespace X` spells its own `export`, so it stays a
		// TypeScript-only node rather than moving inside an export wrapper.
		return false

	case ast.KindFunctionDeclaration, ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration,
		ast.KindModuleDeclaration, ast.KindImportEqualsDeclaration,
		ast.KindMissingDeclaration:
		return hasExportKeyword(node)

	default:
		return true
	}
}

// hasExportKeyword reports whether the declaration spells `export` in the
// source. A dotted namespace such as `namespace N.M {}` nests a declaration of
// its own for each name after the first, and tsgo gives those an `export`
// modifier it synthesizes rather than one the source writes; only a written
// `export` moves a declaration inside an `ExportNamedDeclaration`.
func hasExportKeyword(node *ast.Node) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	return slices.ContainsFunc(modifiers.Nodes, func(modifier *ast.Node) bool {
		return modifier.Kind == ast.KindExportKeyword && modifier.Flags&ast.NodeFlagsReparsed == 0
	})
}

// shield is one `try` block a return stands in. ESLint keeps such a return in
// the pending list for as long as it is traversing the rest of that `try`
// statement, so nothing in the `catch` clause or the `finally` block clears it.
type shield struct {
	statement *ast.Node
	block     *ast.Node
}

// covers reports whether node stands in the part of the `try` statement that
// follows the shielded block.
func (s shield) covers(node *ast.Node) bool {
	return node.Pos() >= s.block.End() && node.End() <= s.statement.End()
}

func shieldsOf(node *ast.Node, root *ast.Node) []shield {
	var shields []shield
	for current := node; current != nil && current != root && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if parent.Kind == ast.KindTryStatement && parent.AsTryStatement().TryBlock == current {
			shields = append(shields, shield{statement: parent, block: current})
		}
	}
	return shields
}

// ---------------------------------------------------------------------------
// candidates
// ---------------------------------------------------------------------------

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
// fix
// ---------------------------------------------------------------------------

// buildFix removes the return statement. The edit spans the enclosing function
// even though only the return is dropped from it, which is how ESLint keeps
// this fix from being applied in the same pass as an overlapping one.
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

// retainedRange spans the enclosing function, matching the range ESLint keeps
// for the edit. tsgo folds a method's name and its function value into one
// node, where ESTree splits them across a definition and an inner
// FunctionExpression, so for those the range runs from the parameter list to
// the end of the body: a computed key or a decorator holds its own functions,
// and their returns get to be fixed in the same pass.
func retainedRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	for current := node; current != nil; current = current.Parent {
		if !isFunction(current) {
			continue
		}
		switch current.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			if pos, end, ok := utils.BodyLikeRange(current); ok {
				return core.NewTextRange(pos, end)
			}
		}
		return utils.TrimNodeTextRange(sourceFile, current)
	}
	return core.NewTextRange(0, len(sourceFile.Text()))
}
