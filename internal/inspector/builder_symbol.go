package inspector

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// BuildShallowSymbolInfo builds minimal SymbolInfo for nested symbols
// Returns nil for internal/synthetic symbols (compiler-generated virtual symbols)
func (b *Builder) BuildShallowSymbolInfo(symbol *ast.Symbol) *SymbolInfo {
	if symbol == nil {
		return nil
	}

	// Skip internal symbols - they are compiler-generated virtual symbols
	if IsInternalSymbol(symbol) {
		return nil
	}

	info := &SymbolInfo{
		Name:      symbol.Name,
		Flags:     uint32(symbol.Flags),
		FlagNames: GetSymbolFlagNames(symbol.Flags),
	}

	// Set Pos from declaration (for fetching full info later)
	var decl *ast.Node
	if symbol.ValueDeclaration != nil {
		decl = symbol.ValueDeclaration
	} else if len(symbol.Declarations) > 0 {
		decl = symbol.Declarations[0]
	}
	if decl != nil {
		// Use name position if available, otherwise declaration position
		if name := decl.Name(); name != nil {
			info.Pos = b.GetTokenPos(name.AsNode())
		} else {
			info.Pos = b.GetTokenPos(decl)
		}
		// Check if declaration is from external file
		declFile := ast.GetSourceFileOfNode(decl)
		if declFile != nil && declFile != b.sourceFile {
			info.FileName = declFile.FileName()
		}
	}

	return info
}

// BuildSymbolInfo builds SymbolInfo from a Symbol
// Returns nil for internal/synthetic symbols (compiler-generated virtual symbols)
func (b *Builder) BuildSymbolInfo(symbol *ast.Symbol) *SymbolInfo {
	if symbol == nil {
		return nil
	}

	// Skip internal symbols - they are compiler-generated virtual symbols
	if IsInternalSymbol(symbol) {
		return nil
	}

	info := &SymbolInfo{
		Id:             uint64(ast.GetSymbolId(symbol)),
		Name:           symbol.Name,
		Flags:          uint32(symbol.Flags),
		FlagNames:      GetSymbolFlagNames(symbol.Flags),
		CheckFlags:     uint32(symbol.CheckFlags),
		CheckFlagNames: GetCheckFlagNames(symbol.CheckFlags),
	}

	// Position for on-demand fetch
	var decl *ast.Node
	if symbol.ValueDeclaration != nil {
		decl = symbol.ValueDeclaration
	} else if len(symbol.Declarations) > 0 {
		decl = symbol.Declarations[0]
	}
	if decl != nil {
		if name := decl.Name(); name != nil {
			info.Pos = b.GetTokenPos(name.AsNode())
		} else {
			info.Pos = b.GetTokenPos(decl)
		}
		// Check if declaration is from external file
		declFile := ast.GetSourceFileOfNode(decl)
		if declFile != nil && declFile != b.sourceFile {
			info.FileName = declFile.FileName()
		}
	}

	// Declarations (as NodeInfo for lazy loading)
	if len(symbol.Declarations) > 0 {
		info.Declarations = make([]*NodeInfo, 0, len(symbol.Declarations))
		for _, decl := range symbol.Declarations {
			info.Declarations = append(info.Declarations, b.BuildShallowNodeInfo(decl))
		}
	}

	// Value declaration (as NodeInfo for lazy loading)
	if symbol.ValueDeclaration != nil {
		info.ValueDeclaration = b.BuildShallowNodeInfo(symbol.ValueDeclaration)
	}

	// Members (full for detailed info)
	if len(symbol.Members) > 0 {
		info.Members = make([]*SymbolInfo, 0, len(symbol.Members))
		for _, member := range symbol.Members {
			info.Members = append(info.Members, b.BuildSymbolInfo(member))
		}
	}

	// Exports (full for detailed info)
	if len(symbol.Exports) > 0 {
		info.Exports = make([]*SymbolInfo, 0, len(symbol.Exports))
		for _, exp := range symbol.Exports {
			info.Exports = append(info.Exports, b.BuildSymbolInfo(exp))
		}
	}

	return info
}
