package typeservice

import (
	"cmp"
	"math"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A JS SymbolTable retains insertion order. Keep the checker's choice of
// symbols and order them by the tables visited, then by binder order.
type symbolOrder struct {
	ranks map[*ast.Symbol]int
	next  int
}

func newSymbolOrder(symbols []*ast.Symbol) symbolOrder {
	ranks := make(map[*ast.Symbol]int, len(symbols))
	for _, symbol := range symbols {
		ranks[symbol] = math.MaxInt
	}
	return symbolOrder{ranks: ranks}
}

func (order *symbolOrder) table(table ast.SymbolTable) {
	order.next++
	// A filtered scope can contain two locals inside a class with thousands of members.
	if len(order.ranks) < len(table) {
		for symbol, rank := range order.ranks {
			if rank == math.MaxInt && table[symbol.Name] == symbol {
				order.ranks[symbol] = order.next
			}
		}
		return
	}
	for _, symbol := range table {
		if order.ranks[symbol] == math.MaxInt {
			order.ranks[symbol] = order.next
		}
	}
}

func (order *symbolOrder) symbol(symbol *ast.Symbol) {
	if order.ranks[symbol] == math.MaxInt {
		order.next++
		order.ranks[symbol] = order.next
	}
}

type symbolBinding struct {
	symbol, owner *ast.Symbol
	decl, file    *ast.Node
	rank, depth   int
	mergePhase    uint8
	filePhase     uint8
}

func bindingOf(symbol *ast.Symbol, rank int) symbolBinding {
	b := symbolBinding{symbol: symbol, owner: symbol, rank: rank}
	if symbol.Flags&ast.SymbolFlagsPrototype != 0 {
		b.owner = symbol.Parent // The binder inserts prototype when it enters the class.
	}
	if len(b.owner.Declarations) != 0 {
		b.decl = sourceJSDocDeclaration(b.owner.Declarations[0])
		for n := b.decl; n != nil; n = n.Parent {
			b.depth++
			b.file = n
			switch n.Kind {
			// The JS binder finishes the file, then binds JSDoc aliases, then JSDoc imports.
			case ast.KindJSDocTypedefTag, ast.KindJSDocCallbackTag, ast.KindJSTypeAliasDeclaration:
				b.filePhase = 1
			case ast.KindJSDocImportTag, ast.KindJSImportDeclaration:
				b.filePhase = 2
			case ast.KindModuleDeclaration:
				// The checker merges global augmentations before external module augmentations.
				if ast.IsGlobalScopeAugmentation(n) {
					b.mergePhase = 1
				} else if ast.IsExternalModuleAugmentation(n) {
					b.mergePhase = 2
				}
			}
		}
	}
	return b
}

func (order symbolOrder) sort(c *checker.Checker, symbols []*ast.Symbol) {
	if len(symbols) < 2 {
		return
	}
	bindings := make([]symbolBinding, len(symbols))
	for i, symbol := range symbols {
		bindings[i] = bindingOf(symbol, order.ranks[symbol])
	}
	flow := make(bindingFlow)
	slices.SortFunc(bindings, func(a, b symbolBinding) int {
		if byTable := cmp.Compare(a.rank, b.rank); byTable != 0 {
			return byTable
		}
		if a.decl != nil && b.decl != nil {
			if byMerge := cmp.Compare(a.mergePhase, b.mergePhase); byMerge != 0 {
				return byMerge
			}
			if a.file == b.file {
				if byPhase := cmp.Compare(a.filePhase, b.filePhase); byPhase != 0 {
					return byPhase
				}
				if bySyntax := compareBindingNodes(a, b, flow); bySyntax != 0 {
					return bySyntax
				}
			}
		}
		return c.CompareSymbols(a.owner, b.owner)
	})
	for i, b := range bindings {
		symbols[i] = b.symbol
	}
}

func compareBindingNodes(a, b symbolBinding, flow bindingFlow) int {
	left, right := a.decl, b.decl
	for range a.depth - b.depth {
		left = left.Parent
	}
	for range b.depth - a.depth {
		right = right.Parent
	}
	if left != right {
		for left.Parent != right.Parent {
			left, right = left.Parent, right.Parent
		}
		if byChild := cmp.Compare(bindingChildPhase(left), bindingChildPhase(right)); byChild != 0 && !flow.unreachable(left.Parent) {
			return byChild
		}
	}
	return cmp.Compare(a.decl.Pos(), b.decl.Pos())
}

// The binder visits most siblings in source order. These children run earlier
// or later than their position in the syntax tree.
func bindingChildPhase(child *ast.Node) int {
	parent := child.Parent
	switch parent.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock:
		if child.Kind == ast.KindFunctionDeclaration {
			return -1
		}
	case ast.KindForStatement:
		if child == parent.AsForStatement().Incrementor {
			return 1
		}
	case ast.KindForInStatement, ast.KindForOfStatement:
		if child == parent.AsForInOrOfStatement().Expression {
			return -1
		}
	case ast.KindCallExpression:
		if child == parent.Expression() && immediateCallee(parent) != nil {
			return 1
		}
	case ast.KindParameter, ast.KindBindingElement:
		if child == parent.Name() {
			return 1
		}
	case ast.KindBinaryExpression:
		if child == parent.AsBinaryExpression().Left && ast.IsDestructuringAssignment(parent) {
			if target := ast.GetAssignmentTarget(parent); target != nil && ast.IsDestructuringAssignment(target) {
				return 1 // A nested pattern binds its default value before its target.
			}
		}
	}
	return 0
}

func immediateCallee(node *ast.Node) *ast.Node {
	if node.Kind == ast.KindCallExpression && !ast.IsOptionalChain(node) {
		if callee := ast.SkipParentheses(node.Expression()); ast.IsFunctionExpressionOrArrowFunction(callee) {
			return callee
		}
	}
	return nil
}

type entryFlow uint8

const (
	noFlow entryFlow = iota
	deadFlow
	liveFlow
)

// Cache the binder's flow at each reordering site for this sort.
type bindingFlow map[*ast.Node]entryFlow

func (flow bindingFlow) unreachable(node *ast.Node) bool {
	if state := flow.entry(node); state != noFlow {
		return state == deadFlow
	}
	// An empty pattern containing only class field initializers may have no
	// observation in its own flow. Its surrounding expression supplies the entry.
	return flow.enclosingUnreachable(node, true)
}

func (flow bindingFlow) entry(node *ast.Node) entryFlow {
	if state, ok := flow[node]; ok {
		return state
	}
	state := liveFlow
	if callee := immediateCallee(node); callee != nil {
		// An argument can terminate flow before the binder visits the callee.
		if flow.observed(callee) != liveFlow && !slices.ContainsFunc(node.AsCallExpression().Arguments.Nodes, func(arg *ast.Node) bool {
			return flow.observed(arg) == liveFlow
		}) {
			state = deadFlow
		}
	} else if ast.IsDestructuringAssignment(node) {
		state = flow.children(node)
	} else if flow.enclosingUnreachable(node, false) {
		state = deadFlow
	}
	flow[node] = state
	return state
}

func (flow bindingFlow) children(node *ast.Node) entryFlow {
	state := noFlow
	node.ForEachChild(func(child *ast.Node) bool {
		state = max(state, flow.observed(child))
		return state == liveFlow
	})
	return state
}

func (flow bindingFlow) observed(node *ast.Node) entryFlow {
	if data := node.FlowNodeData(); data != nil && data.FlowNode != nil {
		if data.FlowNode.Flags&ast.FlowFlagsUnreachable != 0 {
			return deadFlow
		}
		return liveFlow
	}
	// A class static block runs in the surrounding flow. A field initializer
	// or nested function gets its own flow even in an unreachable argument.
	if binder.GetContainerFlags(node)&binder.ContainerFlagsIsControlFlowContainer != 0 && !ast.IsClassStaticBlockDeclaration(node) {
		return noFlow
	}
	if immediateCallee(node) != nil || ast.IsDestructuringAssignment(node) {
		return flow.entry(node)
	}
	return flow.children(node)
}

// Unreachable statements bind children in syntax order, including their expressions.
func (flow bindingFlow) enclosingUnreachable(node *ast.Node, missingEntry bool) bool {
	var child *ast.Node
	for ; node != nil; child, node = node, node.Parent {
		if node.Flags&ast.NodeFlagsUnreachable != 0 {
			return true
		}
		if missingEntry && child != nil {
			if skippedBindingChild(node, child) {
				return true
			}
		}
		if immediateCallee(node) != nil {
			return flow.entry(node) == deadFlow
		}
		if node.Kind == ast.KindBlock && !ast.IsFunctionLike(node.Parent) {
			// Blocks have no entry flow of their own. The binder records it on
			// their first statement, skipping hoisted declarations.
			if first := firstFlowStatement(node); first != nil {
				if ast.KindFirstStatement <= first.Kind && first.Kind <= ast.KindLastStatement {
					return first.FlowNodeData().FlowNode == nil
				}
				return first.Flags&ast.NodeFlagsUnreachable != 0
			}
		}
		if ast.IsFunctionExpressionOrArrowFunction(node) {
			return ast.GetFunctionFlags(node) == ast.FunctionFlagsNormal && ast.GetImmediatelyInvokedFunctionExpression(node) != nil && flow.observed(node) != liveFlow
		}
		if ast.IsPotentiallyExecutableNode(node) || binder.GetContainerFlags(node)&binder.ContainerFlagsIsControlFlowContainer != 0 {
			break
		}
	}
	return false
}

// With no flow-bearing reference in a pattern, only its evaluation context
// can tell whether the binder reached it.
func skippedBindingChild(parent, child *ast.Node) bool {
	switch parent.Kind {
	case ast.KindBinaryExpression:
		b := parent.AsBinaryExpression()
		if child == b.Right {
			switch b.OperatorToken.Kind {
			case ast.KindAmpersandAmpersandToken:
				return impossibleBindingCondition(b.Left, true)
			case ast.KindBarBarToken:
				return impossibleBindingCondition(b.Left, false)
			case ast.KindCommaToken:
				return endsBindingFlow(b.Left)
			}
		}
	case ast.KindConditionalExpression:
		c := parent.AsConditionalExpression()
		return child != c.Condition && impossibleBindingCondition(c.Condition, child == c.WhenTrue)
	case ast.KindCallExpression:
		args := parent.AsCallExpression().Arguments.Nodes
		if i := slices.Index(args, child); i >= 0 {
			return slices.ContainsFunc(args[:i], endsBindingFlow)
		}
	}
	return false
}

func impossibleBindingCondition(node *ast.Node, truth bool) bool {
	if ast.IsBooleanLiteral(node) {
		return (node.Kind == ast.KindTrueKeyword) != truth
	}
	if !ast.IsLogicalExpression(node) {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPrefixUnaryExpression {
		return impossibleBindingCondition(node.AsPrefixUnaryExpression().Operand, !truth)
	}
	b := node.AsBinaryExpression()
	if b.OperatorToken.Kind != ast.KindAmpersandAmpersandToken && b.OperatorToken.Kind != ast.KindBarBarToken {
		return false
	}
	if (b.OperatorToken.Kind == ast.KindAmpersandAmpersandToken) == truth {
		return impossibleBindingCondition(b.Left, truth) || impossibleBindingCondition(b.Right, truth)
	}
	return impossibleBindingCondition(b.Left, truth) && impossibleBindingCondition(b.Right, truth)
}

func endsBindingFlow(node *ast.Node) bool {
	callee := immediateCallee(ast.SkipParentheses(node))
	if callee == nil || ast.GetFunctionFlags(callee) != ast.FunctionFlagsNormal || callee.Flags&ast.NodeFlagsHasImplicitReturn != 0 || callee.Body().Kind != ast.KindBlock {
		return false
	}
	return !ast.ForEachReturnStatement(callee.Body(), func(stmt *ast.Node) bool {
		data := stmt.FlowNodeData().FlowNode
		return data != nil && data.Flags&ast.FlowFlagsUnreachable == 0
	})
}

func firstFlowStatement(block *ast.Node) *ast.Node {
	for _, stmt := range block.StatementList().Nodes {
		if ast.KindFirstStatement <= stmt.Kind && stmt.Kind <= ast.KindLastStatement || ast.IsPotentiallyExecutableNode(stmt) {
			return stmt
		}
		if stmt.Kind == ast.KindBlock {
			if first := firstFlowStatement(stmt); first != nil {
				return first
			}
		}
	}
	return nil
}

func moduleExports(c *checker.Checker, module *ast.Symbol) []*ast.Symbol {
	exports := c.GetExportsOfModule(module)
	if len(exports) < 2 {
		return exports
	}
	var order symbolOrder
	root := c.ResolveExternalModuleSymbol(module)
	if root != module || root.Exports[ast.InternalSymbolNameExportStar] != nil {
		order = newSymbolOrder(exports)
		seen := make(map[*ast.Symbol]bool)
		var visit func(*ast.Symbol)
		visit = func(current *ast.Symbol) {
			if current == nil || seen[current] {
				return
			}
			seen[current] = true
			order.table(current.Exports)
			if stars := current.Exports[ast.InternalSymbolNameExportStar]; stars != nil {
				for _, decl := range stars.Declarations {
					visit(c.GetSymbolAtLocation(decl.ModuleSpecifier()))
				}
			}
		}
		// With export =, the checker appends type-only exports of the original module.
		// Those symbols sort after the tables visited from the target.
		visit(root)
	}
	order.sort(c, exports)
	return exports
}

func (order *symbolOrder) scopes(c *checker.Checker, location *ast.Node, meaning ast.SymbolFlags) {
	for node := location; node != nil; node = node.Parent {
		if locals := node.Locals(); locals != nil && !ast.IsGlobalSourceFile(node) {
			order.table(locals)
		}
		symbol := node.Symbol()
		switch node.Kind {
		case ast.KindSourceFile, ast.KindModuleDeclaration, ast.KindEnumDeclaration:
			if symbol != nil {
				order.table(c.GetMergedSymbol(symbol).Exports)
			}
		case ast.KindClassExpression:
			order.symbol(symbol)
			fallthrough
		case ast.KindClassDeclaration, ast.KindInterfaceDeclaration:
			if meaning&ast.SymbolFlagsType != 0 {
				order.table(c.GetMergedSymbol(symbol).Members)
			}
		case ast.KindFunctionExpression:
			order.symbol(symbol)
		}
		switch node.Kind {
		case ast.KindMethodDeclaration, ast.KindMethodSignature, ast.KindConstructor, ast.KindGetAccessor,
			ast.KindSetAccessor, ast.KindFunctionDeclaration, ast.KindFunctionExpression:
			order.symbol(c.GetArgumentsSymbol())
		}
	}
	// The checker installs globalThis before binding any source files.
	order.symbol(c.GetGlobalSymbol("globalThis", ast.SymbolFlagsValue, nil))
}
