package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"gotest.tools/v3/assert"
)

// parseAndFindNode parses code and finds the first node matching the given kind.
func parseAndFindNode(t *testing.T, code string, kind ast.Kind) (*ast.Node, *ast.SourceFile) {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir.Dir, "file.ts")
	fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: code})
	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
	assert.NilError(t, err, "couldn't create program for code: "+code)
	sourceFile := program.GetSourceFile(filePath)

	var found *ast.Node
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node.Kind == kind {
			found = node
			return true
		}
		return node.ForEachChild(walk)
	}
	for _, stmt := range sourceFile.Statements.Nodes {
		if walk(stmt) {
			break
		}
	}
	if found == nil {
		t.Fatalf("no node of kind %v found in code: %s", kind, code)
	}
	return found, sourceFile
}

func TestIsInStrictMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     string
		kind     ast.Kind
		expected bool
	}{
		// === ES Module (import/export) → strict ===
		{
			name:     "ES module with export",
			code:     `export {}; if (true) function f() {}`,
			kind:     ast.KindFunctionDeclaration,
			expected: true,
		},
		{
			name:     "ES module with import",
			code:     `import "foo"; if (true) function f() {}`,
			kind:     ast.KindFunctionDeclaration,
			expected: true,
		},

		// === "use strict" at file level → strict ===
		{
			name:     "use strict directive at file level",
			code:     `"use strict"; if (true) function f() {}`,
			kind:     ast.KindFunctionDeclaration,
			expected: true,
		},
		{
			name:     "use strict with single quotes",
			code:     `'use strict'; if (true) function f() {}`,
			kind:     ast.KindFunctionDeclaration,
			expected: true,
		},
		{
			name:     "escaped use strict at file level",
			code:     `"use\x20strict"; if (true) function f() {}`,
			kind:     ast.KindFunctionDeclaration,
			expected: false,
		},

		// === "use strict" in function body → strict inside that function ===
		{
			name:     "use strict in enclosing function",
			code:     `function outer() { "use strict"; if (true) { var x = 1; } }`,
			kind:     ast.KindVariableStatement,
			expected: true,
		},
		{
			name:     "escaped use strict in enclosing function",
			code:     `function outer() { "use\x20strict"; if (true) { var x = 1; } }`,
			kind:     ast.KindVariableStatement,
			expected: false,
		},

		// === Class body (implicit strict) ===
		{
			name:     "function inside class method",
			code:     `class C { method() { if (true) function f() {} } }`,
			kind:     ast.KindFunctionDeclaration,
			expected: true,
		},
		{
			name:     "function inside class static block",
			code:     `class C { static { if (true) function f() {} } }`,
			kind:     ast.KindFunctionDeclaration,
			expected: true,
		},

		// === TypeScript namespace/module scopes are strict in typescript-eslint ===
		{
			name:     "function inside namespace",
			code:     `namespace N { if (true) function f() {} }`,
			kind:     ast.KindFunctionDeclaration,
			expected: true,
		},

		// === Non-strict (script without module/strict/class) ===
		{
			name:     "plain script file",
			code:     `if (true) function f() {}`,
			kind:     ast.KindFunctionDeclaration,
			expected: false,
		},
		{
			name:     "var in script function without use strict",
			code:     `function outer() { if (true) { var x = 1; } }`,
			kind:     ast.KindVariableStatement,
			expected: false,
		},

		// === Nested: strict function inside non-strict file ===
		{
			name:     "use strict in nested function only",
			code:     `function outer() { "use strict"; var x = 1; }`,
			kind:     ast.KindVariableStatement,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			node, sourceFile := parseAndFindNode(t, tt.code, tt.kind)
			result := IsInStrictMode(node, sourceFile)
			assert.Equal(t, result, tt.expected, "IsInStrictMode mismatch for: %s", tt.code)
		})
	}
}

func TestHasUseStrictDirective(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name:     "double-quoted directive",
			code:     `"use strict"; var x = 1;`,
			expected: true,
		},
		{
			name:     "single-quoted directive",
			code:     `'use strict'; var x = 1;`,
			expected: true,
		},
		{
			name:     "directive after another prologue directive",
			code:     `"use asm"; "use strict"; var x = 1;`,
			expected: true,
		},
		{
			name: "directive after comments",
			code: `/* before */ // still before
"use strict"; var x = 1;`,
			expected: true,
		},
		{
			name:     "directive after byte order mark",
			code:     "\ufeff\"use strict\"; var x = 1;",
			expected: true,
		},
		{
			name:     "directive after hashbang",
			code:     "#!/usr/bin/env node\n\"use strict\"; var x = 1;",
			expected: true,
		},
		{
			name:     "no directive",
			code:     `var x = 1;`,
			expected: false,
		},
		{
			name:     "other directive only",
			code:     `"use asm"; var x = 1;`,
			expected: false,
		},
		{
			name:     "directive after prologue ends",
			code:     `0; "use strict"; var x = 1;`,
			expected: false,
		},
		{
			name:     "parenthesized string is not a directive",
			code:     `("use strict"); var x = 1;`,
			expected: false,
		},
		{
			name:     "hex escape",
			code:     `"use\x20strict"; var x = 1;`,
			expected: false,
		},
		{
			name:     "unicode escape",
			code:     `"use\u0020strict"; var x = 1;`,
			expected: false,
		},
		{
			name:     "escaped first character",
			code:     `"\x75se strict"; var x = 1;`,
			expected: false,
		},
		{
			name:     "identity escape",
			code:     `"use\ strict"; var x = 1;`,
			expected: false,
		},
		{
			name: "line continuation",
			code: `"use\
 strict"; var x = 1;`,
			expected: false,
		},
		{
			name:     "exact directive after escaped lookalike",
			code:     `"use\x20strict"; "use strict"; var x = 1;`,
			expected: true,
		},
		{
			name:     "empty file",
			code:     ``,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rootDir := fixtures.GetRootDir()
			filePath := tspath.ResolvePath(rootDir.Dir, "file.ts")
			fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: tt.code})
			program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
			assert.NilError(t, err)
			sourceFile := program.GetSourceFile(filePath)
			result := HasUseStrictDirective(sourceFile.AsNode())
			assert.Equal(t, result, tt.expected, "HasUseStrictDirective mismatch for: %s", tt.code)
		})
	}
}

func TestHasUseStrictDirectiveInFunctionBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name:     "exact directive",
			code:     `function f() { "use strict"; work(); }`,
			expected: true,
		},
		{
			name:     "escaped lookalike",
			code:     `function f() { "use\x20strict"; work(); }`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body, _ := parseAndFindNode(t, tt.code, ast.KindBlock)
			assert.Equal(t, HasUseStrictDirective(body), tt.expected)
		})
	}
}

func TestHasUseStrictDirectiveWithoutSourceFile(t *testing.T) {
	t.Parallel()

	if HasUseStrictDirective(nil) {
		t.Fatal("HasUseStrictDirective(nil) = true, want false")
	}

	body, _ := parseAndFindNode(t, `function f() { "use strict"; }`, ast.KindBlock)
	body.Parent = nil
	if HasUseStrictDirective(body) {
		t.Fatal("HasUseStrictDirective(detached block) = true, want false")
	}
}
