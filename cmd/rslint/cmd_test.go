package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
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

// createTestDiagnostic is a helper that creates a program from source, then builds a
// RuleDiagnostic spanning from startOffset to endOffset in the source text.
func createTestDiagnostic(t *testing.T, source string, startOffset, endOffset int) (rule.RuleDiagnostic, tspath.ComparePathsOptions) {
	t.Helper()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(`{"include":["./index.ts"]}`), 0644); err != nil {
		t.Fatalf("Failed to write tsconfig: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "index.ts"), []byte(source), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}

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

	diagnostic := rule.RuleDiagnostic{
		RuleName:   "test-rule",
		SourceFile: sourceFile,
		Range:      core.NewTextRange(startOffset, endOffset),
		Message:    rule.RuleMessage{Id: "test", Description: "Test diagnostic"},
		Severity:   rule.SeverityError,
	}
	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          tmpDir,
		UseCaseSensitiveFileNames: true,
	}
	return diagnostic, comparePathOptions
}

// renderDiagnostic renders a diagnostic and returns the output string.
func renderDiagnostic(t *testing.T, d rule.RuleDiagnostic, opts tspath.ComparePathsOptions) string {
	t.Helper()
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	printDiagnosticDefault(d, w, opts)
	w.Flush()
	return buf.String()
}

// TestPrintDiagnosticFold tests that long diagnostic ranges are folded (first 2 + ... + last 2)
// instead of showing "Error range is too big".
func TestPrintDiagnosticFold(t *testing.T) {
	// Generate a source file with many lines
	var sb strings.Builder
	sb.WriteString("const a = 1;\n")                // line 1
	for i := 2; i <= 20; i++ {
		sb.WriteString(fmt.Sprintf("const v%d = %d;\n", i, i)) // lines 2-20
	}
	source := sb.String()

	t.Run("short span - no fold", func(t *testing.T) {
		// Diagnostic spanning 2 lines (lines 2-3) → codebox = lines 1-4 (4 lines, diff=3 < 4)
		start := strings.Index(source, "const v2")
		end := strings.Index(source, "const v3")
		d, opts := createTestDiagnostic(t, source, start, end)
		output := renderDiagnostic(t, d, opts)

		// Should show all lines in range without folding
		if !strings.Contains(output, "v2") || !strings.Contains(output, "v3") {
			t.Errorf("Short span should show all lines.\nOutput:\n%s", output)
		}
		// Context line (v4) should also be visible (no fold)
		if !strings.Contains(output, "v4") {
			t.Errorf("Short span should show context line v4.\nOutput:\n%s", output)
		}
	})

	t.Run("5-line codebox - folds", func(t *testing.T) {
		// Diagnostic spanning 3 lines (lines 3-5) → codebox = lines 2-6 (5 lines, >= 5)
		start := strings.Index(source, "const v3")
		end := strings.Index(source, "const v6")
		d, opts := createTestDiagnostic(t, source, start, end)
		output := renderDiagnostic(t, d, opts)

		// Middle lines should be skipped (v4 is in the folded region)
		if strings.Contains(output, "v4") {
			t.Errorf("Folded region should not show middle lines.\nOutput:\n%s", output)
		}
		// Should show first 2 and last 2 codebox lines
		if !strings.Contains(output, "v2") {
			t.Errorf("Should show first context line.\nOutput:\n%s", output)
		}
		if !strings.Contains(output, "v6") {
			t.Errorf("Should show last context line.\nOutput:\n%s", output)
		}
	})

	t.Run("large span - previously too big", func(t *testing.T) {
		// Diagnostic spanning 15 lines → was previously "too big" (> 13 lines)
		start := strings.Index(source, "const v2")
		end := strings.Index(source, "const v17")
		d, opts := createTestDiagnostic(t, source, start, end)
		output := renderDiagnostic(t, d, opts)

		// Must NOT contain the old "too big" message
		if strings.Contains(output, "too big") || strings.Contains(output, "Skipping") {
			t.Errorf("Should fold instead of showing 'too big'.\nOutput:\n%s", output)
		}
		// Middle lines should be folded
		if strings.Contains(output, "v10") {
			t.Errorf("Middle line v10 should be folded.\nOutput:\n%s", output)
		}
		// Should show first and last context lines
		if !strings.Contains(output, "const a") {
			t.Errorf("Should show first context line.\nOutput:\n%s", output)
		}
		if !strings.Contains(output, "v17") {
			t.Errorf("Should show last context line.\nOutput:\n%s", output)
		}
	})

	t.Run("fold preserves highlight on last lines", func(t *testing.T) {
		start := strings.Index(source, "const v2")
		end := strings.Index(source, "const v15")
		d, opts := createTestDiagnostic(t, source, start, end)
		output := renderDiagnostic(t, d, opts)

		if !utf8.ValidString(output) {
			t.Errorf("Output is not valid UTF-8.\nOutput:\n%s", output)
		}
		if !strings.Contains(output, "╰") {
			t.Errorf("Should contain closing border.\nOutput:\n%s", output)
		}
		// The last displayed lines should contain the code content
		if !strings.Contains(output, "v15") {
			t.Errorf("Should show v15 in last displayed lines.\nOutput:\n%s", output)
		}
		if !strings.Contains(output, "v16") {
			t.Errorf("Should show v16 (context after diagnostic end).\nOutput:\n%s", output)
		}
		// Middle lines should be folded away
		if strings.Contains(output, "v8") {
			t.Errorf("Middle line v8 should be folded.\nOutput:\n%s", output)
		}
	})
}

// TestPrintDiagnosticSingleLine tests rendering a diagnostic on a single line (no fold).
func TestPrintDiagnosticSingleLine(t *testing.T) {
	source := "const x = 1;\nconst y = 2;\nconst z = 3;\n"
	start := strings.Index(source, "y")
	d, opts := createTestDiagnostic(t, source, start, start+1)
	output := renderDiagnostic(t, d, opts)

	// All 3 lines should be visible (no folding for a single-line diagnostic)
	if !strings.Contains(output, "const x") || !strings.Contains(output, "const y") || !strings.Contains(output, "const z") {
		t.Errorf("Single-line diagnostic should show all context lines.\nOutput:\n%s", output)
	}
}

// TestPrintDiagnosticEdgeCases tests boundary conditions.
func TestPrintDiagnosticEdgeCases(t *testing.T) {
	t.Run("diagnostic at file start", func(t *testing.T) {
		source := "const x = 1;\nconst y = 2;\n"
		d, opts := createTestDiagnostic(t, source, 0, 5) // "const"
		output := renderDiagnostic(t, d, opts)

		if !strings.Contains(output, "const") {
			t.Errorf("Should show diagnostic at file start.\nOutput:\n%s", output)
		}
		if !strings.Contains(output, "╰") {
			t.Errorf("Should have closing border.\nOutput:\n%s", output)
		}
	})

	t.Run("diagnostic at file end", func(t *testing.T) {
		source := "const a = 1;\nconst b = 2;\nconst c = 3;"
		end := len(source)
		start := strings.LastIndex(source, "const c")
		d, opts := createTestDiagnostic(t, source, start, end)
		output := renderDiagnostic(t, d, opts)

		if !strings.Contains(output, "const c") {
			t.Errorf("Should show diagnostic at file end.\nOutput:\n%s", output)
		}
	})

	t.Run("exact fold threshold - 5 line codebox", func(t *testing.T) {
		// 5 lines in source, diagnostic spanning lines 2-4 → codebox = lines 1-5 (diff=4, folds)
		source := "line1\nline2\nline3\nline4\nline5\n"
		start := strings.Index(source, "line2")
		end := strings.Index(source, "line5")
		d, opts := createTestDiagnostic(t, source, start, end)
		output := renderDiagnostic(t, d, opts)

		// Middle lines should be folded (line3 is in the folded region)
		if strings.Contains(output, "line3") {
			t.Errorf("Should fold middle lines at 5-line codebox.\nOutput:\n%s", output)
		}
		// First and last lines should be visible
		if !strings.Contains(output, "line1") || !strings.Contains(output, "line5") {
			t.Errorf("Should show first and last codebox lines.\nOutput:\n%s", output)
		}
	})

	t.Run("single line file", func(t *testing.T) {
		source := "const x = 1;"
		d, opts := createTestDiagnostic(t, source, 0, len(source))
		output := renderDiagnostic(t, d, opts)

		if !strings.Contains(output, "const x") {
			t.Errorf("Should render single-line file.\nOutput:\n%s", output)
		}
		if !strings.Contains(output, "╰") {
			t.Errorf("Should have closing border.\nOutput:\n%s", output)
		}
	})

	t.Run("diagnostic spans entire file with fold", func(t *testing.T) {
		// 6-line file, diagnostic covers everything → codebox = entire file, no extra context
		source := "aaa\nbbb\nccc\nddd\neee\nfff\n" // cspell:disable-line
		d, opts := createTestDiagnostic(t, source, 0, len(source)-1)
		output := renderDiagnostic(t, d, opts)

		// Should fold (6 lines >= 5)
		if strings.Contains(output, "ccc") {
			t.Errorf("Middle line should be folded.\nOutput:\n%s", output)
		}
		// First and last lines visible
		if !strings.Contains(output, "aaa") {
			t.Errorf("Should show first line.\nOutput:\n%s", output)
		}
		if !strings.Contains(output, "fff") {
			t.Errorf("Should show last line.\nOutput:\n%s", output)
		}
		if !strings.Contains(output, "╰") {
			t.Errorf("Should have closing border.\nOutput:\n%s", output)
		}
	})

	t.Run("just under fold threshold - 4 line codebox", func(t *testing.T) {
		// Diagnostic spanning 2 lines → codebox = 4 lines (diff=3, no fold)
		source := "line1\nline2\nline3\nline4\nline5\n"
		start := strings.Index(source, "line2")
		end := strings.Index(source, "line3")
		d, opts := createTestDiagnostic(t, source, start, end)
		output := renderDiagnostic(t, d, opts)

		// All codebox lines should be visible (no folding)
		if !strings.Contains(output, "line1") || !strings.Contains(output, "line4") {
			t.Errorf("4-line codebox should show all lines.\nOutput:\n%s", output)
		}
	})
}

// TestSyntaxErrorFormat tests that syntax errors produce clean, readable messages
// without dumping the entire file text.
func TestSyntaxErrorFormat(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(`{"include":["./index.ts"]}`), 0644); err != nil {
		t.Fatalf("Failed to write tsconfig: %v", err)
	}
	// Write a file with syntax error
	badSource := "const x = ;\nconst y = 2;\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.ts"), []byte(badSource), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	_, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err == nil {
		t.Fatal("Expected error for file with syntax errors")
	}

	errMsg := err.Error()

	// Should contain structured location info
	if !strings.Contains(errMsg, "(1,") {
		t.Errorf("Error should contain line number in tsc format.\nGot: %s", errMsg)
	}
	if !strings.Contains(errMsg, "error TS") {
		t.Errorf("Error should contain 'error TS' code.\nGot: %s", errMsg)
	}
	if !strings.Contains(errMsg, "syntactic error") {
		t.Errorf("Error should mention syntactic error count.\nGot: %s", errMsg)
	}

	// Must NOT contain the entire file source text
	if strings.Contains(errMsg, "const y = 2") {
		t.Errorf("Error message should not dump the entire file text.\nGot: %s", errMsg)
	}
}

// TestSyntaxErrorFormatMultiple tests that multiple syntax errors are each formatted on their own line.
func TestSyntaxErrorFormatMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(`{"include":["./index.ts"]}`), 0644); err != nil {
		t.Fatalf("Failed to write tsconfig: %v", err)
	}
	// Two syntax errors: missing expression on line 1, missing semicolon style on line 2
	badSource := "const x = ;\nconst y = ;\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.ts"), []byte(badSource), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	_, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err == nil {
		t.Fatal("Expected error for file with syntax errors")
	}

	errMsg := err.Error()

	// Should report the correct count
	if !strings.Contains(errMsg, "2 syntactic error") {
		t.Errorf("Should report 2 errors.\nGot: %s", errMsg)
	}

	// Both errors should have their own line with location info
	lines := strings.Split(errMsg, "\n")
	errorLines := 0
	for _, line := range lines {
		if strings.Contains(line, "error TS") {
			errorLines++
		}
	}
	if errorLines != 2 {
		t.Errorf("Expected 2 error lines with 'error TS', got %d.\nFull message:\n%s", errorLines, errMsg)
	}

	// Should reference both line 1 and line 2
	if !strings.Contains(errMsg, "(1,") {
		t.Errorf("Should contain line 1 reference.\nGot: %s", errMsg)
	}
	if !strings.Contains(errMsg, "(2,") {
		t.Errorf("Should contain line 2 reference.\nGot: %s", errMsg)
	}
}

// TestSyntacticErrorType tests that CreateProgram returns a *SyntacticError
// that can be type-asserted to access raw diagnostics.
func TestSyntacticErrorType(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(`{"include":["./index.ts"]}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "index.ts"), []byte("const x = ;\n"), 0644)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	_, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err == nil {
		t.Fatal("Expected error")
	}

	var syntacticErr *utils.SyntacticError
	if !errors.As(err, &syntacticErr) {
		t.Fatalf("Error should be *utils.SyntacticError, got %T", err)
	}
	if len(syntacticErr.Diagnostics) == 0 {
		t.Fatal("SyntacticError should contain diagnostics")
	}
	if syntacticErr.Diagnostics[0].File() == nil {
		t.Fatal("Diagnostic should have a file")
	}
}

// TestReportSyntacticErrorsPretty tests that reportSyntacticErrors renders
// syntax errors with code snippets (like tsc --pretty).
func TestReportSyntacticErrorsPretty(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(`{"include":["./index.ts"]}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "index.ts"), []byte("const x = ;\nconst y = 2;\n"), 0644)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	_, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err == nil {
		t.Fatal("Expected error")
	}

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          tmpDir,
		UseCaseSensitiveFileNames: true,
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	reported := reportSyntacticErrors(err, w, comparePathOptions)
	if !reported {
		t.Fatal("reportSyntacticErrors should return true for SyntacticError")
	}

	output := buf.String()

	// Should render with code snippet box (like rule diagnostics)
	if !strings.Contains(output, "╭") || !strings.Contains(output, "╰") {
		t.Errorf("Should render code snippet box.\nOutput:\n%s", output)
	}
	// Should show the source code context
	if !strings.Contains(output, "const x") {
		t.Errorf("Should show the error line.\nOutput:\n%s", output)
	}
	// Rule name should be the TS error code
	if !strings.Contains(output, "TS") {
		t.Errorf("Rule name should contain TS error code.\nOutput:\n%s", output)
	}
	// Should show file location
	if !strings.Contains(output, "index.ts:1:") {
		t.Errorf("Should show file location.\nOutput:\n%s", output)
	}
}

// TestReportSyntacticErrorsNonSyntactic tests that reportSyntacticErrors
// returns false for non-SyntacticError errors.
func TestReportSyntacticErrorsNonSyntactic(t *testing.T) {
	err := errors.New("some other error")
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	reported := reportSyntacticErrors(err, w, tspath.ComparePathsOptions{})
	if reported {
		t.Fatal("reportSyntacticErrors should return false for non-SyntacticError")
	}
	if buf.Len() != 0 {
		t.Fatal("Should not write anything for non-SyntacticError")
	}
}
