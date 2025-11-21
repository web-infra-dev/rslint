package main

import (
	"context"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
)

// NodeReference uniquely identifies a node by file name and span.
type NodeReference struct {
	SourceFileId int `json:"sourcefile_id"`
	Start        int `json:"start"`
	End          int `json:"end"`
}
type SymbolInfo struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Flags      int    `json:"flags"`
	CheckFlags int    `json:"check_flags"`
}

type TypeInfo struct {
	Id    int `json:"id"`
	Flags int `json:"flags"`
}
type SymbolTable = map[NodeReference]SymbolInfo
type TypeTable = map[NodeReference]TypeInfo

// collect_symbol_table walks every AST node in the program once and records the
// symbol (if any) associated with that node keyed by its file/span tuple.
func CollectSemantic(program *compiler.Program) Semantic {
	semantic := Semantic{
		Symtab:   make(map[ast.SymbolId]SymbolInfo),
		Typetab:  make(map[checker.TypeId]TypeInfo),
		Sym2type: make(map[ast.SymbolId]checker.TypeId),
		Node2sym: make(map[NodeReference]ast.SymbolId),
	}
	if program == nil {
		return semantic
	}

	tc, done := program.GetTypeChecker(context.Background())
	defer done()

	for id, sourceFile := range program.GetSourceFiles() {
		CollectSemanticInFile(tc, sourceFile, &semantic, id)
	}
	return semantic
}

// primitive types from https://github.com/quininer/typescript-go/blob/da56f163200ee7880c2134cf821ef08372383f7b/internal/checker/checker.go#L892
type PrimTypes struct {
	String checker.TypeId `json:"string"`
	// any       checker.TypeId
	// error     checker.TypeId
	// unknown   checker.TypeId
	// undefined checker.TypeId
	// null      checker.TypeId
	// number    checker.TypeId
	// bigint    checker.TypeId
	// false     checker.TypeId
	// true      checker.TypeId
	// void      checker.TypeId
	// bool      checker.TypeId
}
type Semantic struct {
	Symtab    map[ast.SymbolId]SymbolInfo     `json:"symtab"`
	Typetab   map[checker.TypeId]TypeInfo     `json:"typetab"`
	Sym2type  map[ast.SymbolId]checker.TypeId `json:"sym2type"`
	Node2sym  map[NodeReference]ast.SymbolId  `json:"node2sym"`
	Primtypes PrimTypes                       `json:"primtypes"`
}

func NewSemantic() Semantic {
	return Semantic{
		Symtab:    make(map[ast.SymbolId]SymbolInfo),
		Typetab:   make(map[checker.TypeId]TypeInfo),
		Sym2type:  make(map[ast.SymbolId]checker.TypeId),
		Node2sym:  make(map[NodeReference]ast.SymbolId),
		Primtypes: PrimTypes{},
	}
}
func initPrimitiveTypes(tc *checker.Checker, semantic *Semantic) {
	semantic.Primtypes = PrimTypes{
		String: tc.GetStringType().Id(),
	}
}
func CollectSemanticInFile(tc *checker.Checker, file *ast.SourceFile, semantic *Semantic, sourceFileId int) {
	if tc == nil || file == nil {
		return
	}

	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}

		// Skip synthetic nodes without stable positions.
		if node.Pos() >= 0 && node.End() >= 0 {

			if symbol := tc.GetSymbolAtLocation(node); symbol != nil {

				if ty := tc.GetTypeOfSymbol(symbol); ty != nil {
					key := NodeReference{
						SourceFileId: sourceFileId,
						Start:        node.Pos(),
						End:          node.End(),
					}
					sym_id := ast.GetSymbolId(symbol)
					type_id := ty.Id()
					semantic.Symtab[sym_id] = SymbolInfo{
						Id:         int(sym_id),
						Name:       symbol.Name,
						Flags:      int(symbol.Flags),
						CheckFlags: int(symbol.CheckFlags),
					}
					semantic.Typetab[type_id] = TypeInfo{
						Id:    int(type_id),
						Flags: int(ty.Flags()),
					}
					semantic.Sym2type[sym_id] = type_id
					(semantic.Node2sym)[key] = sym_id

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
