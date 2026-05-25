package first

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// parseStatements parses code and returns the top-level statement nodes.
func parseStatements(t *testing.T, code string) []*ast.Node {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	fileName := "file.ts"
	fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sf := program.GetSourceFile(fileName)
	if sf == nil || sf.Statements == nil {
		t.Fatal("source file has no statements")
	}
	return sf.Statements.Nodes
}

// TestHasReferenceBeforeImportName tests the name-based fallback path
// (used when TypeChecker is unavailable).
func TestHasReferenceBeforeImportName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     string
		impIndex int // index of the import in body to check
		want     bool
	}{
		{
			name:     "value reference before import — detected",
			code:     "console.log(x);\nimport { x } from './foo';",
			impIndex: 1,
			want:     true,
		},
		{
			name:     "no reference before import — not detected",
			code:     "var a = 1;\nimport { x } from './foo';",
			impIndex: 1,
			want:     false,
		},
		{
			name:     "declaration name same as import — NOT a reference (IsDeclarationIdentifier)",
			code:     "function foo() {}\nimport { foo } from './mod';",
			impIndex: 1,
			want:     false,
		},
		{
			name:     "variable declaration name same as import — NOT a reference",
			code:     "var x = 1;\nimport { x } from './foo';",
			impIndex: 1,
			want:     false,
		},
		{
			name:     "class declaration name same as import — NOT a reference",
			code:     "class Foo {}\nimport { Foo } from './mod';",
			impIndex: 1,
			want:     false,
		},
		{
			name:     "reference in nested expression — detected",
			code:     "var a = foo();\nimport { foo } from './mod';",
			impIndex: 1,
			want:     true,
		},
		{
			name:     "reference in nested function body — detected",
			code:     "function setup() { x(); }\nimport { x } from './foo';",
			impIndex: 1,
			want:     true,
		},
		{
			name:     "side-effect import (no bindings) — not detected",
			code:     "var a = 1;\nimport './side-effect';",
			impIndex: 1,
			want:     false,
		},
		{
			name:     "parameter with same name as import — IS a reference (name-based is conservative)",
			code:     "function bar(x: number) { return x; }\nimport { x } from './foo';",
			impIndex: 1,
			// Name-based: `x` as parameter declaration is skipped by IsDeclarationIdentifier,
			// but `return x` is a value reference with matching name → true (conservative).
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body := parseStatements(t, tt.code)
			if tt.impIndex >= len(body) {
				t.Fatalf("impIndex %d out of range (body has %d statements)", tt.impIndex, len(body))
			}
			importNode := body[tt.impIndex]
			bindingNodes := utils.GetImportBindingNodes(importNode)
			got := hasReferenceBeforeImportName(body, tt.impIndex, bindingNodes)
			if got != tt.want {
				t.Errorf("hasReferenceBeforeImportName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestBuildFixWithNonASCII verifies that buildFix handles source text
// containing non-ASCII characters (e.g. in comments) without panicking
// or producing incorrect output.
func TestBuildFixWithNonASCII(t *testing.T) {
	t.Parallel()

	code := "import a from 'a';\nvar x = 1; // 中文注释\nimport b from 'b';"
	body := parseStatements(t, code)

	// body[0] = import a (legal)
	// body[1] = var x = 1 (non-import)
	// body[2] = import b (misplaced)
	if len(body) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(body))
	}

	lastLegalImp := body[0]
	sortNodes := []errorInfo{
		{
			node:      body[2],
			rangeFrom: body[1].End(),
			rangeTo:   body[2].End(),
		},
	}

	fixes := buildFix(code, body, lastLegalImp, sortNodes)
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}

	// Apply the fix manually: replace [0, overallEnd) with fix text
	result := fixes[0].Text + code[fixes[0].Range.End():]

	// The comment with Chinese characters should be preserved
	if !contains(result, "中文注释") {
		t.Errorf("non-ASCII comment lost in fix output: %s", result)
	}
	// import b should come after import a
	if !contains(result, "import a from 'a';") || !contains(result, "import b from 'b';") {
		t.Errorf("imports missing in fix output: %s", result)
	}
}

// TestBuildFixWithEmoji verifies that emoji in comments are preserved
// through the fix and don't interfere with the whitespace detection logic.
func TestBuildFixWithEmoji(t *testing.T) {
	t.Parallel()

	code := "import a from 'a';\nvar x = 1; // 🚀 rocket\nimport b from 'b';"
	body := parseStatements(t, code)
	if len(body) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(body))
	}

	lastLegalImp := body[0]
	sortNodes := []errorInfo{
		{
			node:      body[2],
			rangeFrom: body[1].End(),
			rangeTo:   body[2].End(),
		},
	}

	fixes := buildFix(code, body, lastLegalImp, sortNodes)
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}

	result := fixes[0].Text + code[fixes[0].Range.End():]

	if !contains(result, "🚀 rocket") {
		t.Errorf("emoji comment lost in fix output: %s", result)
	}
	if !contains(result, "import b from 'b';") {
		t.Errorf("import missing in fix output: %s", result)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexSubstring(s, substr) >= 0
}

func indexSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
