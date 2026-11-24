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
type symbolOrder map[*ast.Symbol]int

func (order symbolOrder) table(table ast.SymbolTable) {
	rank := len(order) + 1
	for _, symbol := range table {
		if order[symbol] == 0 {
			order[symbol] = rank
		}
	}
}

func (order symbolOrder) symbol(symbol *ast.Symbol) {
	if symbol != nil && order[symbol] == 0 {
		order[symbol] = len(order) + 1
	}
}

func bindingPosition(decl *ast.Node) int {
	// TypeScript binds function declarations before other statements in the block.
	if decl.Kind == ast.KindFunctionDeclaration {
		return decl.Parent.Pos() - 1
	}
	return decl.Pos()
}

func bindingSymbol(symbol *ast.Symbol) *ast.Symbol {
	// The binder adds the implicit prototype when it enters the class.
	if symbol.Flags&ast.SymbolFlagsPrototype != 0 {
		return symbol.Parent
	}
	return symbol
}

func (order symbolOrder) sort(c *checker.Checker, symbols []*ast.Symbol) {
	slices.SortFunc(symbols, func(a, b *ast.Symbol) int {
		if byTable := cmp.Compare(cmp.Or(order[a], math.MaxInt), cmp.Or(order[b], math.MaxInt)); byTable != 0 {
			return byTable
		}
		a, b = bindingSymbol(a), bindingSymbol(b)
		if len(a.Declarations) != 0 && len(b.Declarations) != 0 {
			left, right := a.Declarations[0], b.Declarations[0]
			if (left.Kind == ast.KindFunctionDeclaration || right.Kind == ast.KindFunctionDeclaration) &&
				ast.GetSourceFileOfNode(left) == ast.GetSourceFileOfNode(right) {
				if byPosition := cmp.Compare(bindingPosition(left), bindingPosition(right)); byPosition != 0 {
					return byPosition
				}
			}
		}
		return c.CompareSymbols(a, b)
	})
}

func moduleExports(c *checker.Checker, module *ast.Symbol) []*ast.Symbol {
	exports := c.GetExportsOfModule(module)
	if len(exports) < 2 {
		return exports
	}
	var order symbolOrder
	root := c.ResolveExternalModuleSymbol(module)
	if root != module || root.Exports[ast.InternalSymbolNameExportStar] != nil {
		order = make(symbolOrder)
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

func (order symbolOrder) scopes(c *checker.Checker, location *ast.Node) {
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
			order.table(c.GetMergedSymbol(symbol).Members)
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
