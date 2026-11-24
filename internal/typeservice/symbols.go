package typeservice

import (
	"cmp"
	"math"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
		b.decl = b.owner.Declarations[0]
		switch b.decl.Kind {
		case ast.KindJSTypeAliasDeclaration, ast.KindJSImportDeclaration:
			b.decl = sourceJSDocDeclaration(b.decl)
		}
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
				if bySyntax := compareBindingNodes(a, b); bySyntax != 0 {
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

func compareBindingNodes(a, b symbolBinding) int {
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
		if byChild := cmp.Compare(bindingChildPhase(left), bindingChildPhase(right)); byChild != 0 && !unreachableBinding(left.Parent) {
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
		if child == parent.Expression() && !ast.IsOptionalChain(parent) && ast.IsFunctionExpressionOrArrowFunction(ast.SkipParentheses(child)) {
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

// Unreachable statements bind children in syntax order, including their expressions.
func unreachableBinding(node *ast.Node) bool {
	for ; node != nil; node = node.Parent {
		if node.Flags&ast.NodeFlagsUnreachable != 0 {
			return true
		}
		if ast.IsPotentiallyExecutableNode(node) || ast.IsFunctionLikeOrClassStaticBlockDeclaration(node) {
			break
		}
	}
	return false
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
