package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestPrintDiagnosticUTF8 tests that printDiagnosticDefault correctly renders
// UTF-8 characters (Chinese, Japanese, Korean, Emoji) in diagnostic output.
func TestPrintDiagnosticUTF8(t *testing.T) {
	testCases := []struct {
		name         string
		source       string
		expectedText string
	}{
		{
			name:         "Chinese comment",
			source:       "// 未使用的变量\nconst unused = 42;",
			expectedText: "未使用的变量",
		},
		{
			name:         "Japanese comment",
			source:       "// 使用されていない変数\nconst unused = 42;",
			expectedText: "使用されていない変数",
		},
		{
			name:         "Korean comment",
			source:       "// 사용되지 않는 변수\nconst unused = 42;",
			expectedText: "사용되지 않는 변수",
		},
		{
			name:         "Emoji in comment",
			source:       "// 🎉 Celebration\nconst x = 1;",
			expectedText: "🎉",
		},
		{
			name:         "Mixed UTF-8 content",
			source:       "// Hello 世界 🌍\nconst world = '世界';",
			expectedText: "Hello 世界 🌍",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tsconfigPath := filepath.Join(tmpDir, "tsconfig.json")
			sourcePath := filepath.Join(tmpDir, "index.ts")

			// Write tsconfig.json
			if err := os.WriteFile(tsconfigPath, []byte(`{"include":["./index.ts"]}`), 0644); err != nil {
				t.Fatalf("Failed to write tsconfig: %v", err)
			}

			// Write source file with UTF-8 content
			if err := os.WriteFile(sourcePath, []byte(tc.source), 0644); err != nil {
				t.Fatalf("Failed to write source file: %v", err)
			}

			// Create program
			fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			host := utils.CreateCompilerHost(tmpDir, fs)
			program, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
			if err != nil {
				t.Fatalf("Failed to create program: %v", err)
			}

			// Find source file
			var sourceFile *ast.SourceFile
			for _, file := range program.GetSourceFiles() {
				if strings.HasSuffix(file.FileName(), "index.ts") {
					sourceFile = file
					break
				}
			}
			if sourceFile == nil {
				t.Fatal("Source file not found")
			}

			// Create diagnostic at position of variable name
			text := sourceFile.Text()
			var diagnosticStart int
			if idx := strings.Index(text, "unused"); idx != -1 {
				diagnosticStart = idx
			} else if idx := strings.Index(text, "x"); idx != -1 {
				diagnosticStart = idx
			} else if idx := strings.Index(text, "world"); idx != -1 {
				diagnosticStart = idx
			} else {
				diagnosticStart = 0
			}

			diagnostic := rule.RuleDiagnostic{
				RuleName:   "test-rule",
				SourceFile: sourceFile,
				Range:      core.NewTextRange(diagnosticStart, diagnosticStart+1),
				Message: rule.RuleMessage{
					Id:          "test",
					Description: "Test diagnostic for UTF-8 rendering",
				},
				Severity: rule.SeverityWarning,
			}

			// Capture diagnostic output
			var buf bytes.Buffer
			writer := bufio.NewWriter(&buf)

			comparePathOptions := tspath.ComparePathsOptions{
				CurrentDirectory:          tmpDir,
				UseCaseSensitiveFileNames: true,
			}

			// Call the actual function
			printDiagnosticDefault(diagnostic, writer, comparePathOptions)
			writer.Flush()

			output := buf.String()

			// Verify expected UTF-8 text is present in output
			if !strings.Contains(output, tc.expectedText) {
				t.Errorf("Output does not contain expected text %q.\nOutput:\n%s", tc.expectedText, output)
			}

			// Check for replacement character (indicates UTF-8 decoding error)
			// cspell:ignore FFFD
			if strings.Contains(output, "\uFFFD") {
				t.Errorf("Output contains replacement character (U+FFFD), indicating UTF-8 decoding error.\nOutput:\n%s", output)
			}

			// Verify all characters are valid UTF-8
			if !utf8.ValidString(output) {
				t.Errorf("Output is not valid UTF-8.\nOutput:\n%s", output)
			}
		})
	}
}
