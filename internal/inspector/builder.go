package inspector

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// Builder builds AST info responses from TypeScript AST nodes
type Builder struct {
	checker    *checker.Checker
	sourceFile *ast.SourceFile
}

// NewBuilder creates a new AST info builder
func NewBuilder(c *checker.Checker, sf *ast.SourceFile) *Builder {
	return &Builder{
		checker:    c,
		sourceFile: sf,
	}
}

// Checker returns the type checker
func (b *Builder) Checker() *checker.Checker {
	return b.checker
}

// SourceFile returns the source file
func (b *Builder) SourceFile() *ast.SourceFile {
	return b.sourceFile
}

// GetTokenPos returns the actual start position of a node (skipping trivia like whitespace/comments)
func (b *Builder) GetTokenPos(node *ast.Node) int {
	// Use the node's actual source file, not b.sourceFile, since the node might be from an external file
	nodeFile := ast.GetSourceFileOfNode(node)
	if nodeFile == nil {
		return node.Pos()
	}
	return scanner.GetTokenPosOfNode(node, nodeFile, false)
}

// SafeCall executes a function and recovers from any panic
func SafeCall(fn func()) {
	defer func() {
		recover()
	}()
	fn()
}

// AddListMeta adds NodeList metadata for an array property
func AddListMeta(info *NodeInfo, name string, list *ast.NodeList) {
	if list == nil {
		return
	}
	if info.ListMetas == nil {
		info.ListMetas = make(map[string]*NodeListMeta)
	}
	info.ListMetas[name] = &NodeListMeta{
		Pos:              list.Pos(),
		End:              list.End(),
		HasTrailingComma: list.HasTrailingComma(),
	}
}

// IsInternalSymbol checks if a symbol is an internal/synthetic symbol (like __call__, __new__, etc.)
func IsInternalSymbol(symbol *ast.Symbol) bool {
	// Internal symbols start with \xFE (InternalSymbolName prefix in TypeScript)
	return len(symbol.Name) > 0 && symbol.Name[0] == '\xFE'
}

// KindToString converts an AST Kind to a string representation with package prefix
func KindToString(kind ast.Kind) string {
	return "ast." + kind.String()
}

// SafeGetTarget safely gets the Target of a type, returning nil if it panics or doesn't apply
func (b *Builder) SafeGetTarget(t *checker.Type) (result *checker.Type) {
	defer func() {
		if r := recover(); r != nil {
			result = nil
		}
	}()
	return t.Target()
}

// BuildSourceFileNodeInfo builds NodeInfo directly from a SourceFile without Node conversion
func (b *Builder) BuildSourceFileNodeInfo(sf *ast.SourceFile) *NodeInfo {
	if sf == nil {
		return nil
	}

	node := sf.AsNode()
	info := &NodeInfo{
		Id:        uint64(ast.GetNodeId(node)),
		Kind:      int(node.Kind),
		KindName:  KindToString(node.Kind),
		Pos:       node.Pos(),
		End:       node.End(),
		Flags:     int(node.Flags),
		FlagNames: GetNodeFlagNames(node.Flags),
	}

	// Set fileName if this is an external file (different from builder's sourceFile)
	if sf != b.sourceFile {
		info.FileName = sf.FileName()
	}

	// Build statements directly from SourceFile (avoiding AsSourceFile conversion)
	info.Statements = b.buildNodeList(sf.Statements, "Statements", info)

	// EndOfFileToken (shallow)
	if sf.EndOfFileToken != nil {
		info.EndOfFileToken = b.BuildShallowNodeInfo(sf.EndOfFileToken.AsNode())
	}

	// Imports (shallow array)
	imports := sf.Imports()
	if len(imports) > 0 {
		info.Imports = make([]*NodeInfo, 0, len(imports))
		for _, imp := range imports {
			if imp != nil {
				info.Imports = append(info.Imports, b.BuildShallowNodeInfo(imp.AsNode()))
			}
		}
	}

	// SourceFile metadata
	info.IsDeclarationFile = sf.IsDeclarationFile
	info.ScriptKind = int(sf.ScriptKind)
	info.IdentifierCount = sf.IdentifierCount
	info.SymbolCount = sf.SymbolCount
	info.NodeCount = sf.NodeCount

	// Build locals
	b.buildLocals(node, info)

	return info
}
