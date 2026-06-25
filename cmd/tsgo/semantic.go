package main

import (
	"context"
	"strings"
	_ "unsafe"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
)

// sanitizeSymbolName replaces the internal symbol name prefix (\xFE) with "__"
// so that consumers receive valid UTF-8 symbol names.
func sanitizeSymbolName(name string) []byte {
	if strings.HasPrefix(name, ast.InternalSymbolNamePrefix) {
		return []byte("__" + name[len(ast.InternalSymbolNamePrefix):])
	}
	return []byte(name)
}

//go:linkname getAliasedSymbol github.com/microsoft/typescript-go/internal/checker.(*Checker).GetAliasedSymbol
func getAliasedSymbol(recv *checker.Checker, symbol *ast.Symbol) *ast.Symbol

type CString = []byte
type SourceFileId = int

// NodeReference uniquely identifies a node by file name and span.
type NodeReference struct {
	SourceFileId SourceFileId `json:"sourcefile_id"`
	Start        int          `json:"start"`
	End          int          `json:"end"`
}
type SymbolInfo struct {
	Id         ast.SymbolId `json:"id"`
	Name       CString      `json:"name"`
	Flags      int          `json:"flags"`
	CheckFlags int          `json:"check_flags"`
	// Declaration node reference (if available)
	Decl *NodeReference `json:"decl,omitempty"`
}
type TypeExtra struct {
	Name map[int]CString      `json:"name"`
	Func map[int]FunctionData `json:"func"`
}
type FunctionData struct {
	Signatures []FuncSignature `json:"signatures"`
}
type FuncSignature struct {
	Result checker.TypeId `json:"result"`
}
type TypeInfo struct {
	Id    checker.TypeId `json:"id"`
	Flags int            `json:"flags"`
}
type SymbolTable = map[NodeReference]SymbolInfo
type TypeTable = map[NodeReference]TypeInfo

// collect_symbol_table walks every AST node in the program once and records the
// symbol (if any) associated with that node keyed by its file/span tuple.
func CollectSemantic(program *compiler.Program) Semantic {
	semantic := NewSemantic()
	if program == nil {
		return semantic
	}

	tc, done := program.GetTypeChecker(context.Background())
	defer done()

	sourceFiles := program.GetSourceFiles()
	sourceFileIds := make(map[*ast.SourceFile]SourceFileId, len(sourceFiles))
	for id, sourceFile := range sourceFiles {
		sourceFileIds[sourceFile] = id
	}

	for id, sourceFile := range sourceFiles {
		CollectSemanticInFile(tc, sourceFile, &semantic, id, sourceFileIds)
	}
	return semantic
}

// primitive types from https://github.com/quininer/typescript-go/blob/da56f163200ee7880c2134cf821ef08372383f7b/internal/checker/checker.go#L892
type PrimTypes struct {
	String    checker.TypeId `json:"string"`
	Any       checker.TypeId `json:"any"`
	Error     checker.TypeId `json:"error"`
	Unknown   checker.TypeId `json:"unknown"`
	Undefined checker.TypeId `json:"undefined"`
	Null      checker.TypeId `json:"null"`
	Number    checker.TypeId `json:"number"`
	Bigint    checker.TypeId `json:"bigint"`
	False     checker.TypeId `json:"false"`
	True      checker.TypeId `json:"true"`
	Void      checker.TypeId `json:"void"`
	Bool      checker.TypeId `json:"bool"`
	Never     checker.TypeId `json:"never"`
}
type Semantic struct {
	Symtab       map[ast.SymbolId]SymbolInfo      `json:"symtab"`
	Typetab      map[checker.TypeId]TypeInfo      `json:"typetab"`
	Sym2type     map[ast.SymbolId]checker.TypeId  `json:"sym2type"`
	AliasSymbols map[ast.SymbolId]ast.SymbolId    `json:"alias_symbols"`
	Node2sym     map[NodeReference]ast.SymbolId   `json:"node2sym"`
	Node2type    map[NodeReference]checker.TypeId `json:"node2type"`
	NodeFlags    map[NodeReference]uint32         `json:"node_flags"`
	Primtypes    PrimTypes                        `json:"primtypes"`
	TypeExtra    TypeExtra                        `json:"type_extra"`
	FuncData     FunctionData                     `json:"func_data"`
	// ShorthandSymbols maps node reference to the value symbol for shorthand property assignments
	// (node -> value_symbol_id)
	ShorthandSymbols map[NodeReference]ast.SymbolId `json:"shorthand_symbols"`
	// ParameterPropertySymbols maps a parameter property name node to the other symbol declared at that location.
	// The primary symbol remains recorded in Node2sym.
	ParameterPropertySymbols map[NodeReference]ast.SymbolId `json:"parameter_property_symbols"`
}

func NewSemantic() Semantic {
	return Semantic{
		Symtab:                   make(map[ast.SymbolId]SymbolInfo),
		Typetab:                  make(map[checker.TypeId]TypeInfo),
		Sym2type:                 make(map[ast.SymbolId]checker.TypeId),
		AliasSymbols:             make(map[ast.SymbolId]ast.SymbolId),
		Node2sym:                 make(map[NodeReference]ast.SymbolId),
		Node2type:                make(map[NodeReference]checker.TypeId),
		NodeFlags:                make(map[NodeReference]uint32),
		ShorthandSymbols:         make(map[NodeReference]ast.SymbolId),
		ParameterPropertySymbols: make(map[NodeReference]ast.SymbolId),
		Primtypes:                PrimTypes{},
		TypeExtra: TypeExtra{
			Name: make(map[int]CString),
			Func: make(map[int]FunctionData),
		},
		FuncData: FunctionData{
			Signatures: []FuncSignature{},
		},
	}
}
func initPrimitiveTypes(tc *checker.Checker, semantic *Semantic) {
	semantic.Primtypes = PrimTypes{
		String:    tc.GetStringType().Id(),
		Number:    tc.GetNumberType().Id(),
		Any:       tc.GetAnyType().Id(),
		Error:     tc.GetErrorType().Id(),
		Unknown:   tc.GetUnknownType().Id(),
		Undefined: tc.GetUndefinedType().Id(),
		Null:      tc.GetNullType().Id(),
		Void:      tc.GetVoidType().Id(),
		Bool:      tc.GetBooleanType().Id(),
		Never:     tc.GetNeverType().Id(),
	}
}
func CollectSemanticInFile(tc *checker.Checker, file *ast.SourceFile, semantic *Semantic, sourceFileId int, sourceFileIds map[*ast.SourceFile]SourceFileId) {
	if tc == nil || file == nil {
		return
	}

	// The AST encoder converts positions to UTF-16 code units. We must use the
	// same encoding here so that the consumer can match semantic data to AST nodes.
	positionMap := file.GetPositionMap()
	utf16 := func(pos int) int {
		return positionMap.UTF8ToUTF16(pos)
	}
	nodeReference := func(node *ast.Node) *NodeReference {
		if node == nil || node.Pos() < 0 || node.End() < 0 {
			return nil
		}

		nodeSourceFile := ast.GetSourceFileOfNode(node)
		if nodeSourceFile == nil {
			return nil
		}
		nodeSourceFileId, ok := sourceFileIds[nodeSourceFile]
		if !ok {
			return nil
		}

		nodePositionMap := nodeSourceFile.GetPositionMap()
		return &NodeReference{
			SourceFileId: nodeSourceFileId,
			Start:        nodePositionMap.UTF8ToUTF16(node.Pos()),
			End:          nodePositionMap.UTF8ToUTF16(node.End()),
		}
	}

	recordType := func(ty *checker.Type) checker.TypeId {
		if ty == nil {
			return 0
		}

		typeID := ty.Id()
		if _, exists := semantic.Typetab[typeID]; !exists {
			semantic.Typetab[typeID] = TypeInfo{
				Id:    typeID,
				Flags: int(ty.Flags()),
			}
			semantic.TypeExtra.Name[int(typeID)] = []byte(tc.TypeToString(ty))
			callSignatures := tc.GetCallSignatures(ty)
			if len(callSignatures) > 0 {
				signatures := []FuncSignature{}
				for _, sig := range callSignatures {
					returnType := checker.Checker_getReturnTypeOfSignature(tc, sig)
					signatures = append(signatures, FuncSignature{
						Result: returnType.Id(),
					})

				}
				semantic.TypeExtra.Func[int(typeID)] = FunctionData{
					Signatures: signatures,
				}
			}
		}

		return typeID
	}

	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}

		// Skip synthetic nodes without stable positions.
		if node.Pos() >= 0 && node.End() >= 0 {
			key := NodeReference{
				SourceFileId: sourceFileId,
				Start:        utf16(node.Pos()),
				End:          utf16(node.End()),
			}
			if node.Flags != 0 {
				semantic.NodeFlags[key] = uint32(node.Flags)
			}
			// typescript will panic if we pass typeDeclaration to GetTypeAtLocation
			if !ast.IsTypeDeclaration(node) {
				if tyAtNode := tc.GetTypeAtLocation(node); tyAtNode != nil {
					if typeID := recordType(tyAtNode); typeID != 0 {
						semantic.Node2type[key] = typeID
					}
				}
			}

			if symbol := tc.GetSymbolAtLocation(node); symbol != nil {

				if ty := tc.GetTypeOfSymbol(symbol); ty != nil {
					typeID := recordType(ty)
					sym_id := ast.GetSymbolId(symbol)

					declRef := nodeReference(symbol.ValueDeclaration)

					semantic.Symtab[sym_id] = SymbolInfo{
						Id:         sym_id,
						Name:       sanitizeSymbolName(symbol.Name),
						Flags:      int(symbol.Flags),
						CheckFlags: int(symbol.CheckFlags),
						Decl:       declRef,
					}
					semantic.Sym2type[sym_id] = typeID
					(semantic.Node2sym)[key] = sym_id
					semantic.Node2type[key] = typeID

					// Resolve alias symbol if this is an alias
					if symbol.Flags&ast.SymbolFlagsAlias != 0 {
						if aliasedSymbol := getAliasedSymbol(tc, symbol); aliasedSymbol != nil {
							aliased_id := ast.GetSymbolId(aliasedSymbol)
							if aliased_id != sym_id {
								semantic.AliasSymbols[sym_id] = aliased_id
							}
						}
					}

				}

			}

			// Collect shorthand assignment value symbol
			if valueSymbol := tc.GetShorthandAssignmentValueSymbol(node); valueSymbol != nil {
				value_sym_id := ast.GetSymbolId(valueSymbol)
				// Also record this symbol if not already recorded
				if _, exists := semantic.Symtab[value_sym_id]; !exists {
					if ty := tc.GetTypeOfSymbol(valueSymbol); ty != nil {
						typeID := recordType(ty)
						declRef := nodeReference(valueSymbol.ValueDeclaration)
						semantic.ShorthandSymbols[key] = value_sym_id
						semantic.Symtab[value_sym_id] = SymbolInfo{
							Id:         value_sym_id,
							Name:       sanitizeSymbolName(valueSymbol.Name),
							Flags:      int(valueSymbol.Flags),
							CheckFlags: int(valueSymbol.CheckFlags),
							Decl:       declRef,
						}
						semantic.Sym2type[value_sym_id] = typeID
					}
				}
			}

			if ast.IsParameterPropertyDeclaration(node, node.Parent) {
				name := node.Name()
				if name != nil && ast.IsIdentifier(name) {
					if nameKey := nodeReference(name); nameKey != nil {
						parameterSymbol, _ := tc.GetSymbolsOfParameterPropertyDeclaration(node, name.Text())
						if parameterSymbol != nil {
							parameterSymbolID := ast.GetSymbolId(parameterSymbol)
							if _, exists := semantic.Symtab[parameterSymbolID]; !exists {
								if ty := tc.GetTypeOfSymbol(parameterSymbol); ty != nil {
									typeID := recordType(ty)
									declRef := nodeReference(parameterSymbol.ValueDeclaration)
									semantic.ParameterPropertySymbols[*nameKey] = parameterSymbolID
									semantic.Symtab[parameterSymbolID] = SymbolInfo{
										Id:         parameterSymbolID,
										Name:       sanitizeSymbolName(parameterSymbol.Name),
										Flags:      int(parameterSymbol.Flags),
										CheckFlags: int(parameterSymbol.CheckFlags),
										Decl:       declRef,
									}
									semantic.Sym2type[parameterSymbolID] = typeID
								}
							}
						}
					}
				}
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}

	visit(file.AsNode())
}

func IsTypeFlagSet(t *checker.Type, flags checker.TypeFlags) bool {
	return t != nil && checker.Type_flags(t)&flags != 0
}

func IsIntrinsicType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsIntrinsic)
}
