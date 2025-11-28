package main

import (
	"testing"
)

func TestCollectSymbolTable(t *testing.T) {
	//t.Log("start")
	test_config := "fixture/tsconfig.json"
	program, err := CreateProgram(test_config)
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}

	CollectSemantic(program)
	//t.Logf("symbol_table:%+v\n", semantic)

	// expectedSymbols := []string{
	// 	"myFunction",
	// 	"MyClass",
	// 	"myVariable",
	// }

	// for _, symbol := range expectedSymbols {
	// 	if _, exists := symbolTable[symbol]; !exists {
	// 		t.Errorf("Expected symbol %s not found in symbol table", symbol)
	// 	}
	// }
}
