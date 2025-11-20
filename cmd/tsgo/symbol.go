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
	Name       string `json:"name"`
	Flags      int    `json:"flags"`
	CheckFlags int    `json:"check_flags"`
}

type TypeInfo struct {
	Id    checker.TypeId `json:"id"`
	Flags int            `json:"flags"`
}
type SymbolTable = map[NodeReference]SymbolInfo
type TypeTable = map[NodeReference]TypeInfo

// collect_symbol_table walks every AST node in the program once and records the
// symbol (if any) associated with that node keyed by its file/span tuple.
func CollectSemantic(program *compiler.Program) (SymbolTable, TypeTable) {
	symbol_table := SymbolTable{}
	type_table := TypeTable{}

	if program == nil {
		return symbol_table, type_table
	}

	tc, done := program.GetTypeChecker(context.Background())
	defer done()

	for id, sourceFile := range program.GetSourceFiles() {
		CollectSemanticInFile(tc, sourceFile, &symbol_table, &type_table, id)
	}
	return symbol_table, type_table
}

func CollectSemanticInFile(tc *checker.Checker, file *ast.SourceFile,
	symbol_table *SymbolTable,
	type_table *TypeTable,
	sourceFileId int) {
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

				if ty := tc.GetTypeAtLocation(node); ty != nil {
					key := NodeReference{
						SourceFileId: sourceFileId,
						Start:        node.Pos(),
						End:          node.End(),
					}
					if _, exists := (*symbol_table)[key]; !exists {
						(*symbol_table)[key] = SymbolInfo{
							Name:       symbol.Name,
							Flags:      int(symbol.Flags),
							CheckFlags: int(symbol.CheckFlags),
						}
					}
					if _, exists := (*type_table)[key]; !exists {
						(*type_table)[key] = TypeInfo{
							Id:    ty.Id(),
							Flags: int(ty.Flags()),
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
