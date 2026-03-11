package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"gotest.tools/v3/assert"
)

// parseFunctionNode parses code and finds the first function-like node in the AST.
func parseFunctionNode(t *testing.T, code string) *ast.Node {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "file.ts")
	fs := NewOverlayVFSForFile(filePath, code)
	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	assert.NilError(t, err, "couldn't create program for code: "+code)
	sourceFile := program.GetSourceFile(filePath)

	var funcNode *ast.Node
	var findFunc func(node *ast.Node) bool
	findFunc = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		switch node.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			funcNode = node
			return true
		}
		return node.ForEachChild(findFunc)
	}
	for _, stmt := range sourceFile.Statements.Nodes {
		if findFunc(stmt) {
			break
		}
	}
	if funcNode == nil {
		t.Fatalf("no function-like node found in code: %s", code)
	}
	return funcNode
}

func TestAnalyzeFunctionReturns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		code               string
		endReachable       bool
		hasReturnWithValue bool
		hasEmptyReturn     bool
	}{
		// Basic cases
		{
			name:         "empty function",
			code:         `function foo() {}`,
			endReachable: true,
		},
		{
			name:               "single return with value",
			code:               `function foo() { return 1; }`,
			hasReturnWithValue: true,
		},
		{
			name:           "single empty return",
			code:           `function foo() { return; }`,
			hasEmptyReturn: true,
		},
		{
			name: "throw only",
			code: `function foo() { throw new Error(); }`,
		},

		// Control flow
		{
			name:               "if-else both return",
			code:               `function foo() { if (x) { return 1; } else { return 2; } }`,
			hasReturnWithValue: true,
		},
		{
			name:               "if without else (falls through)",
			code:               `function foo() { if (x) { return 1; } }`,
			endReachable:       true,
			hasReturnWithValue: true,
		},
		{
			name:               "if-else one returns, one throws",
			code:               `function foo() { if (x) { return 1; } else { throw new Error(); } }`,
			hasReturnWithValue: true,
		},
		{
			name: "if-else both throw",
			code: `function foo() { if (x) { throw new Error("a"); } else { throw new Error("b"); } }`,
		},

		// Switch
		{
			name:               "switch with default, all return",
			code:               `function foo() { switch(x) { case 1: return 1; default: return 2; } }`,
			hasReturnWithValue: true,
		},
		{
			name:               "switch without default (falls through)",
			code:               `function foo() { switch(x) { case 1: return 1; } }`,
			endReachable:       true,
			hasReturnWithValue: true,
		},

		// Try/catch
		{
			name:               "try-catch both return",
			code:               `function foo() { try { return 1; } catch(e) { return 2; } }`,
			hasReturnWithValue: true,
		},
		{
			name:               "try returns, catch falls through",
			code:               `function foo() { try { return 1; } catch(e) { } }`,
			endReachable:       true,
			hasReturnWithValue: true,
		},
		{
			name:               "try returns, catch throws",
			code:               `function foo() { try { return 1; } catch(e) { throw e; } }`,
			hasReturnWithValue: true,
		},
		{
			name:               "finally with empty return overrides try",
			code:               `function foo() { try { return 1; } finally { return; } }`,
			hasReturnWithValue: true,
			hasEmptyReturn:     true,
		},
		{
			name:               "finally with return value",
			code:               `function foo() { try { return 1; } finally { return 2; } }`,
			hasReturnWithValue: true,
		},

		// Mixed returns
		{
			name:               "mixed return and empty return",
			code:               `function foo() { if (x) { return 1; } else { return; } }`,
			hasReturnWithValue: true,
			hasEmptyReturn:     true,
		},

		// Getter
		{
			name:               "getter with return",
			code:               `var foo = { get bar() { return 1; } }`,
			hasReturnWithValue: true,
		},
		{
			name:         "empty getter",
			code:         `var foo = { get bar() {} }`,
			endReachable: true,
		},

		// Arrow function with block body
		{
			name:               "arrow with block return",
			code:               `var foo = () => { return 1; }`,
			hasReturnWithValue: true,
		},
		{
			name:         "arrow with empty block",
			code:         `var foo = () => {}`,
			endReachable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			funcNode := parseFunctionNode(t, tt.code)
			result := AnalyzeFunctionReturns(funcNode)

			if result.EndReachable != tt.endReachable {
				t.Errorf("EndReachable: got %v, want %v", result.EndReachable, tt.endReachable)
			}
			if result.HasReturnWithValue != tt.hasReturnWithValue {
				t.Errorf("HasReturnWithValue: got %v, want %v", result.HasReturnWithValue, tt.hasReturnWithValue)
			}
			if result.HasEmptyReturn != tt.hasEmptyReturn {
				t.Errorf("HasEmptyReturn: got %v, want %v", result.HasEmptyReturn, tt.hasEmptyReturn)
			}
		})
	}
}

func TestIsFunctionEndReachable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"empty function", `function foo() {}`, true},
		{"return only", `function foo() { return 1; }`, false},
		{"throw only", `function foo() { throw new Error(); }`, false},
		{"if without else", `function foo() { if (x) { return 1; } }`, true},
		{"if-else both return", `function foo() { if (x) { return 1; } else { return 2; } }`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			funcNode := parseFunctionNode(t, tt.code)
			result := IsFunctionEndReachable(funcNode)
			if result != tt.expected {
				t.Errorf("IsFunctionEndReachable: got %v, want %v", result, tt.expected)
			}
		})
	}

	// Test nil case separately
	t.Run("nil node", func(t *testing.T) {
		result := IsFunctionEndReachable(nil)
		if result != true {
			t.Errorf("IsFunctionEndReachable(nil): got %v, want true", result)
		}
	})
}
