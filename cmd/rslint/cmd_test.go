package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/cmd/rslint/internal/output"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func runLintPipelineForTest(t *testing.T, cwd string, args lintArgs) (int, string, string) {
	t.Helper()

	t.Chdir(cwd)

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	defer stdoutR.Close()
	defer stderrR.Close()

	originalStdout, originalStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdoutW, stderrW

	code := executeLintPipeline(args, context.Background(), nil)

	os.Stdout, os.Stderr = originalStdout, originalStderr
	if err := stdoutW.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	if err := stderrW.Close(); err != nil {
		t.Fatalf("close stderr writer: %v", err)
	}

	stdoutBytes, err := io.ReadAll(stdoutR)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	stderrBytes, err := io.ReadAll(stderrR)
	if err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	return code, string(stdoutBytes), string(stderrBytes)
}

// TestPrintDiagnosticUTF8 tests that the default output renderer correctly
// handles UTF-8 characters (Chinese, Japanese, Korean, Emoji).
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

			comparePathOptions := tspath.ComparePathsOptions{
				CurrentDirectory:          tmpDir,
				UseCaseSensitiveFileNames: true,
			}
			output := renderDiagnostic(t, diagnostic, comparePathOptions)

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
		FilePath:   sourceFile.FileName(),
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
	report := output.NewReport([]rule.RuleDiagnostic{d}, output.Metadata{
		Mode:      output.ModeLint,
		Threads:   1,
		StartedAt: time.Now(),
	})
	if err := output.Render(&buf, report, output.Options{
		Format:       output.FormatDefault,
		ComparePaths: opts,
	}); err != nil {
		t.Fatalf("render diagnostic: %v", err)
	}
	return buf.String()
}

// TestPrintDiagnosticFold tests that long diagnostic ranges are folded (first 2 + ... + last 2)
// instead of showing "Error range is too big".
func TestPrintDiagnosticFold(t *testing.T) {
	// Generate a source file with many lines
	var sb strings.Builder
	sb.WriteString("const a = 1;\n") // line 1
	for i := 2; i <= 20; i++ {
		fmt.Fprintf(&sb, "const v%d = %d;\n", i, i) // lines 2-20
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

	t.Run("source contains invalid UTF-8 bytes", func(t *testing.T) {
		// Regression: the codebox renderer iterated `for _, char := range
		// codeboxText` and advanced a manual byte counter by
		// `utf8.RuneLen(char)`. Go's range yields utf8.RuneError (U+FFFD)
		// for each invalid UTF-8 byte but only advances 1 byte — yet
		// utf8.RuneLen(RuneError) is 3 (the encoded length of U+FFFD).
		// The manual counter fell out of sync and downstream sliced the
		// source text past its length, panicking with
		// `slice bounds out of range [:17] with length 7`.
		//
		// Source: `//` + 5 invalid UTF-8 first bytes (0xFF 0xFE 0xFD
		// 0xFC 0xFB) — mirrors a real swc-loader fixture in rspack that
		// triggered this in production.
		source := "//\xff\xfe\xfd\xfc\xfb"
		d, opts := createTestDiagnostic(t, source, len(source), len(source))
		output := renderDiagnostic(t, d, opts)

		// The render must complete without panic and produce a non-empty
		// codebox containing the leading `//`.
		if !strings.Contains(output, "//") {
			t.Errorf("Should render the `//` prefix without panicking.\nOutput:\n%s", output)
		}
		if !strings.Contains(output, "╰") {
			t.Errorf("Should have closing border.\nOutput:\n%s", output)
		}
	})

	t.Run("codebox contains only whitespace", func(t *testing.T) {
		// Regression: indentSize was initialized to math.MaxInt and only
		// updated inside `if !lineIndentCalculated && !unicode.IsSpace`.
		// When every codebox line was whitespace-only (e.g. a 1-byte
		// `"\n"` source), indentSize stayed MaxInt, and
		// `lineMap[line] + indentSize` overflowed int — wrapping to a
		// large negative number that then sliced out of bounds.
		//
		// Source: single LF — the simplest shape that produces a
		// whitespace-only codebox. The diagnostic covers the LF itself
		// (mirrors what eol-last 'never' emits on a 1-byte file).
		source := "\n"
		d, opts := createTestDiagnostic(t, source, 0, 1)
		output := renderDiagnostic(t, d, opts)

		if !strings.Contains(output, "╰") {
			t.Errorf("Should have closing border.\nOutput:\n%s", output)
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

// ======== resolveStartTime tests ========

func TestResolveStartTime_Zero(t *testing.T) {
	before := time.Now()
	result := resolveStartTime(0)
	after := time.Now()

	if result.Before(before) || result.After(after) {
		t.Errorf("expected time.Now() when startTimeMs is 0, got %v", result)
	}
}

func TestResolveStartTime_Positive(t *testing.T) {
	ms := int64(1711800000000) // a fixed epoch millis
	result := resolveStartTime(ms)
	expected := time.UnixMilli(ms)

	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestResolveStartTime_Negative(t *testing.T) {
	before := time.Now()
	result := resolveStartTime(-1)
	after := time.Now()

	if result.Before(before) || result.After(after) {
		t.Errorf("expected time.Now() when startTimeMs is negative, got %v", result)
	}
}

// --- validateTypeCheckOnlyFlags ---
//
// These tests pin down the CLI compatibility policy for --type-check-only:
// it implies --type-check, but is rejected with exit code 2 when combined
// with --fix or --rule (both rely on the lint phase that this mode disables).

func TestValidateTypeCheckOnlyFlags_OffIsNoop(t *testing.T) {
	// typeCheckOnly=false → policy doesn't apply, every other combination ok.
	cases := []struct {
		fix       bool
		ruleFlags []string
	}{
		{false, nil},
		{true, nil},
		{false, []string{"no-console: error"}},
		{true, []string{"no-console: error"}},
	}
	for _, c := range cases {
		code, msg := validateTypeCheckOnlyFlags(false, c.fix, c.ruleFlags)
		if code != 0 || msg != "" {
			t.Errorf("typeCheckOnly=false should never reject (fix=%v rules=%v), got (%d, %q)", c.fix, c.ruleFlags, code, msg)
		}
	}
}

func TestValidateTypeCheckOnlyFlags_AloneIsOK(t *testing.T) {
	code, msg := validateTypeCheckOnlyFlags(true, false, nil)
	if code != 0 || msg != "" {
		t.Errorf("typeCheckOnly alone should be accepted, got (%d, %q)", code, msg)
	}
}

func TestValidateTypeCheckOnlyFlags_RejectsFix(t *testing.T) {
	code, msg := validateTypeCheckOnlyFlags(true, true, nil)
	if code != 2 {
		t.Errorf("expected exit code 2 for --type-check-only + --fix, got %d", code)
	}
	if !strings.Contains(msg, "--fix") || !strings.Contains(msg, "--type-check-only") {
		t.Errorf("expected message to mention both flags, got %q", msg)
	}
}

func TestValidateTypeCheckOnlyFlags_RejectsRule(t *testing.T) {
	code, msg := validateTypeCheckOnlyFlags(true, false, []string{"no-console: error"})
	if code != 2 {
		t.Errorf("expected exit code 2 for --type-check-only + --rule, got %d", code)
	}
	if !strings.Contains(msg, "--rule") || !strings.Contains(msg, "--type-check-only") {
		t.Errorf("expected message to mention both flags, got %q", msg)
	}
}

// TestValidateTypeCheckOnlyFlags_FixTakesPriority documents the diagnostic
// preference when both incompatible flags are present: the error message
// names --fix (not --rule) because --fix is checked first. This isn't a
// correctness property, just a stability guarantee for users with scripts
// scraping stderr.
func TestValidateTypeCheckOnlyFlags_FixTakesPriority(t *testing.T) {
	code, msg := validateTypeCheckOnlyFlags(true, true, []string{"no-console: error"})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(msg, "--fix") {
		t.Errorf("expected --fix to take priority in the error message, got %q", msg)
	}
}

func TestExecuteLintPipelineRejectsInvalidFormatBeforeWork(t *testing.T) {
	code, stdout, stderr := runLintPipelineForTest(t, t.TempDir(), lintArgs{
		Format:         "stylish",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stdout=%q stderr=%q", code, stdout, stderr)
	}
	if stdout != "" {
		t.Fatalf("invalid format wrote stdout: %q", stdout)
	}
	if !strings.Contains(stderr, `invalid output format "stylish"`) {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestExecuteLintPipelineInitIgnoresLintFormat(t *testing.T) {
	dir := t.TempDir()
	code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
		Init:           true,
		Format:         "stylish",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 0 || strings.Contains(stderr, "invalid output format") {
		t.Fatalf("init did not retain priority: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExecuteLintPipelineConfigCatalogSelection(t *testing.T) {
	dir := t.TempDir()
	target := tspath.NormalizePath(filepath.Join(dir, "index.js"))
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, "rslint.json"),
		[]byte(`[{"rules":{"no-debugger":"error"}}]`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	configDir := tspath.NormalizePath(dir)

	t.Run("explicit empty export remains a JS config", func(t *testing.T) {
		emptyConfig := rslintconfig.RslintConfig{}
		code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
			ConfigCatalog: &discovery.ConfigCatalog{
				Configs: map[string]rslintconfig.RslintConfig{configDir: emptyConfig},
				Targets: []discovery.DiscoveredTarget{{
					Path:            target,
					ConfigDirectory: configDir,
					Explicit:        true,
					MergedConfig:    emptyConfig.GetConfigForFile(target, configDir),
				}},
				Explicit: true,
			},
			AllowFiles:     []string{target},
			Format:         "default",
			NoColor:        true,
			SingleThreaded: true,
		})
		if code != 0 || strings.Contains(stdout, "no-debugger") {
			t.Fatalf("explicit empty JS config fell back to rslint.json: code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})

	t.Run("empty automatic catalog uses JSON fallback", func(t *testing.T) {
		code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
			ConfigCatalog:  &discovery.ConfigCatalog{Configs: map[string]rslintconfig.RslintConfig{}},
			AllowFiles:     []string{target},
			Format:         "jsonline",
			NoColor:        true,
			SingleThreaded: true,
		})
		if code != 1 || !strings.Contains(stdout, "no-debugger") {
			t.Fatalf("empty automatic catalog did not use rslint.json: code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})

	t.Run("malformed explicit catalog cannot fall back to JSON", func(t *testing.T) {
		code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
			ConfigCatalog:  &discovery.ConfigCatalog{Explicit: true},
			AllowFiles:     []string{target},
			Format:         "jsonline",
			NoColor:        true,
			SingleThreaded: true,
		})
		if code != 1 || stdout != "" || !strings.Contains(stderr, "explicit config catalog contains 0 configs") {
			t.Fatalf("malformed explicit catalog escaped its invariant: code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
}

func TestExecuteLintPipelineTypedCatalogEnforcesPluginDeclarations(t *testing.T) {
	dir := t.TempDir()
	target := tspath.NormalizePath(filepath.Join(dir, "index.ts"))
	if err := os.WriteFile(target, []byte("let value: any;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	configDir := tspath.NormalizePath(dir)
	for _, test := range []struct {
		name        string
		plugins     []string
		wantFailure bool
	}{
		{name: "undeclared plugin rule is gated"},
		{name: "declared plugin rule runs", plugins: []string{"@typescript-eslint"}, wantFailure: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			entries := rslintconfig.RslintConfig{{
				Files:   []string{"**/*.ts"},
				Plugins: test.plugins,
				Rules:   rslintconfig.Rules{"@typescript-eslint/no-explicit-any": "error"},
			}}
			code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
				ConfigCatalog: &discovery.ConfigCatalog{
					Configs: map[string]rslintconfig.RslintConfig{
						configDir: entries,
					},
					Targets: []discovery.DiscoveredTarget{{
						Path:            target,
						ConfigDirectory: configDir,
						Explicit:        true,
						MergedConfig:    entries.GetConfigForFile(target, configDir),
					}},
					Explicit: true,
				},
				AllowFiles:     []string{target},
				Format:         "jsonline",
				NoColor:        true,
				SingleThreaded: true,
			})
			gotFailure := code != 0 && strings.Contains(stdout, "@typescript-eslint/no-explicit-any")
			if gotFailure != test.wantFailure {
				t.Fatalf("plugin enforcement = %v, want %v: code=%d stdout=%q stderr=%q", gotFailure, test.wantFailure, code, stdout, stderr)
			}
		})
	}
}

// --- shouldShortCircuitOutput ---
//
// Locks two regressions:
//   1. `--type-check-only <dir>` returning exit 0 with no diagnostics
//      because Phase 1's LintedFileCount was always 0, tripping a
//      short-circuit meant for "user pointed at a nonexistent file"
//      (lint mode).
//   2. `--type-check <non-program-file.ts>` silently dropping Phase 2
//      diagnostics: lintedFileCount==0 (scope filtered out everything in
//      Phase 1) but Phase 2 ran program-wide and produced TS errors that
//      the short-circuit would have swallowed.
//
// Either type-check mode must never take the short-circuit.

func TestShouldShortCircuitOutput_NotInTypeCheckOnly(t *testing.T) {
	// All combinations of scope/lintedFileCount must NOT short-circuit
	// when type-check-only is on, because Phase 2 may have output.
	cases := []struct {
		scopeRestricted bool
		lintedFileCount int32
	}{
		{false, 0},
		{false, 5},
		{true, 0},
		{true, 5},
	}
	for _, c := range cases {
		// typeCheckOnly=true with typeCheck=false is non-canonical (main()
		// sets typeCheck=true when typeCheckOnly is on), but the guard must
		// still hold on its own to avoid coupling.
		if shouldShortCircuitOutput(true, false, c.scopeRestricted, c.lintedFileCount) {
			t.Errorf("type-check-only mode must never short-circuit (scope=%v lintedFiles=%d)", c.scopeRestricted, c.lintedFileCount)
		}
	}
}

func TestShouldShortCircuitOutput_NotInTypeCheckMode(t *testing.T) {
	// --type-check (without --type-check-only): Phase 2 runs program-wide
	// regardless of Scope, so even lintedFileCount==0 + scopeRestricted is
	// not enough to drop diagnostics. Locks review-111 Issue 1.
	cases := []struct {
		scopeRestricted bool
		lintedFileCount int32
	}{
		{false, 0},
		{false, 5},
		{true, 0}, // <-- the previously-buggy case
		{true, 5},
	}
	for _, c := range cases {
		if shouldShortCircuitOutput(false, true, c.scopeRestricted, c.lintedFileCount) {
			t.Errorf("--type-check mode must never short-circuit (scope=%v lintedFiles=%d)", c.scopeRestricted, c.lintedFileCount)
		}
	}
}

func TestShouldShortCircuitOutput_LintModeShortCircuitsWhenEmpty(t *testing.T) {
	if !shouldShortCircuitOutput(false, false, true, 0) {
		t.Error("lint mode with scope restriction and zero linted files should short-circuit")
	}
}

func TestShouldShortCircuitOutput_LintModeKeepsRunningOtherwise(t *testing.T) {
	cases := []struct {
		name            string
		scopeRestricted bool
		lintedFileCount int32
	}{
		{"no scope, no files", false, 0},
		{"no scope, files present", false, 5},
		{"scope, files present", true, 5},
	}
	for _, c := range cases {
		if shouldShortCircuitOutput(false, false, c.scopeRestricted, c.lintedFileCount) {
			t.Errorf("%s: lint mode must not short-circuit", c.name)
		}
	}
}

// --- allowFileWarning ---
//
// collectAllowFileWarnings is the structured form of "this CLI file won't
// be linted" diagnostics. formatAllowFileWarning is the message renderer.
// We test them separately so the policy (when to emit) and the wording
// (what the message looks like) are pinned independently.

func TestFormatAllowFileWarning_NotFound(t *testing.T) {
	opts := tspath.ComparePathsOptions{CurrentDirectory: "/work", UseCaseSensitiveFileNames: true}
	msg := formatAllowFileWarning(allowFileWarning{Path: "/work/missing.ts", Kind: allowFileNotFound}, opts)
	if !strings.Contains(msg, "missing.ts") {
		t.Errorf("message should contain file name, got %q", msg)
	}
	if !strings.Contains(msg, "was not found") {
		t.Errorf("message should explain missing file, got %q", msg)
	}
	if !strings.Contains(msg, "skipping") {
		t.Errorf("message should say 'skipping' (lint-side semantics), got %q", msg)
	}
	if strings.HasSuffix(msg, "\n") {
		t.Errorf("message should NOT include trailing newline (caller adds it via Fprintln), got %q", msg)
	}
}

func TestFormatAllowFileWarning_Ignored(t *testing.T) {
	opts := tspath.ComparePathsOptions{CurrentDirectory: "/work", UseCaseSensitiveFileNames: true}
	msg := formatAllowFileWarning(allowFileWarning{Path: "/work/src/x.ts", Kind: allowFileIgnored}, opts)
	if !strings.Contains(msg, "src/x.ts") {
		t.Errorf("message should contain relative path, got %q", msg)
	}
	if !strings.Contains(msg, "ignored because of a matching ignore pattern") {
		t.Errorf("message should reference the ignore pattern, got %q", msg)
	}
}

func TestFormatAllowFileWarning_UnknownKindIsEmpty(t *testing.T) {
	// Defensive: future Kind enum additions shouldn't crash callers if
	// formatter isn't updated. Empty string is the agreed sentinel.
	msg := formatAllowFileWarning(allowFileWarning{Path: "/x.ts", Kind: allowFileWarningKind(99)}, tspath.ComparePathsOptions{})
	if msg != "" {
		t.Errorf("unknown kind should produce empty message, got %q", msg)
	}
}

func TestCollectDiscoveryExplicitWarningsUsesAuthoritativeStatuses(t *testing.T) {
	catalog := &discovery.ConfigCatalog{
		ExplicitInputs: []discovery.ExplicitInputResult{
			{Path: "/work/kept.js", Status: discovery.ExplicitInputConfigured},
			{Path: "/work/config-ignored.js", Status: discovery.ExplicitInputIgnored},
			{Path: "/work/config-ignored.js", Status: discovery.ExplicitInputIgnored, Order: 99},
			{Path: "/work/unconfigured.txt", Status: discovery.ExplicitInputUnconfigured},
			{Path: "/outside/external.js", Status: discovery.ExplicitInputExternal},
			{Path: "/work/gitignored.js", Status: discovery.ExplicitInputIgnored},
		},
	}
	warnings := collectDiscoveryExplicitWarnings(catalog)
	want := []allowFileWarning{
		{Path: "/work/config-ignored.js", Kind: allowFileIgnored},
		{Path: "/work/unconfigured.txt", Kind: allowFileUnconfigured},
		{Path: "/outside/external.js", Kind: allowFileExternal},
		{Path: "/work/gitignored.js", Kind: allowFileIgnored},
	}
	if !reflect.DeepEqual(warnings, want) {
		t.Fatalf("warnings = %#v, want %#v", warnings, want)
	}
}

func TestExecuteLintPipelineDiscoveryFileWarningQuietAndMaxWarnings(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "ignored.ts")
	if err := os.WriteFile(target, []byte("const = ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	entries := rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"no-debugger": "error"}}}
	newCatalog := func() *discovery.ConfigCatalog {
		return &discovery.ConfigCatalog{
			Configs: map[string]rslintconfig.RslintConfig{root: entries},
			ExplicitInputs: []discovery.ExplicitInputResult{{
				Path: target, Status: discovery.ExplicitInputIgnored,
			}},
		}
	}
	tests := []struct {
		name           string
		quiet          bool
		maxWarnings    int
		wantCode       int
		wantStdout     bool
		wantMaxMessage bool
	}{
		{name: "normal", maxWarnings: -1, wantCode: 0, wantStdout: true},
		{name: "quiet", quiet: true, maxWarnings: -1, wantCode: 0},
		{name: "quiet max warnings", quiet: true, maxWarnings: 0, wantCode: 1, wantMaxMessage: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := runLintPipelineForTest(t, root, lintArgs{
				ConfigCatalog:  newCatalog(),
				AllowFiles:     []string{target},
				Format:         "jsonline",
				NoColor:        true,
				SingleThreaded: true,
				Quiet:          test.quiet,
				MaxWarnings:    test.maxWarnings,
			})
			if code != test.wantCode {
				t.Fatalf("code = %d, want %d; stdout=%q stderr=%q", code, test.wantCode, stdout, stderr)
			}
			if got := strings.Contains(stdout, "ignored because of a matching ignore pattern"); got != test.wantStdout {
				t.Fatalf("warning visibility = %v, want %v; stdout=%q", got, test.wantStdout, stdout)
			}
			if got := strings.Contains(stderr, "too many warnings"); got != test.wantMaxMessage {
				t.Fatalf("max-warning message = %v, want %v; stderr=%q", got, test.wantMaxMessage, stderr)
			}
		})
	}
}

func TestCollectAllowFileWarnings_EmptyReturnsNil(t *testing.T) {
	// No allowFiles → no work, no warnings. Important so callers can rely
	// on a non-nil result implying actual user-specified files.
	got := collectAllowFileWarnings(nil, nil, "/work", nil)
	if got != nil {
		t.Errorf("empty allowFiles should produce nil, got %+v", got)
	}
	got = collectAllowFileWarnings([]string{}, nil, "/work", nil)
	if got != nil {
		t.Errorf("empty allowFiles (non-nil slice) should still produce nil, got %+v", got)
	}
}

func TestCollectAllowFileWarnings_NoWarningForFilesScopeMiss(t *testing.T) {
	program := createTestProgram(t, map[string]string{
		"src/app.ts": "const value = 1;",
	})
	target := findProgramFileForTest(t, program, "src/app.ts")
	configDir := tspath.GetDirectoryPath(tspath.GetDirectoryPath(target))

	warnings := collectAllowFileWarnings(
		[]string{target},
		rslintconfig.RslintConfig{
			{Files: []string{"**/*.js"}, Rules: rslintconfig.Rules{"no-console": "error"}},
		},
		configDir,
		nil,
	)
	if len(warnings) != 0 {
		t.Fatalf("files scope miss should not emit warning, got %+v", warnings)
	}
}

func TestCollectAllowFileWarnings_NoWarningForExistingFile(t *testing.T) {
	target := filepath.Join(t.TempDir(), "outside.ts")
	if err := os.WriteFile(target, []byte("const outside = 1;\n"), 0o644); err != nil {
		t.Fatalf("write outside target: %v", err)
	}
	target = tspath.NormalizePath(target)

	warnings := collectAllowFileWarnings(
		[]string{target},
		rslintconfig.RslintConfig{
			{Rules: rslintconfig.Rules{"no-console": "error"}},
		},
		tspath.GetDirectoryPath(target),
		nil,
	)
	if len(warnings) != 0 {
		t.Fatalf("existing files should not produce warnings, got %+v", warnings)
	}
}

func TestCollectAllowFileWarnings_MissingFileWarns(t *testing.T) {
	target := tspath.NormalizePath(filepath.Join(t.TempDir(), "missing.ts"))
	warnings := collectAllowFileWarnings(
		[]string{target},
		rslintconfig.RslintConfig{
			{Rules: rslintconfig.Rules{"no-console": "error"}},
		},
		tspath.GetDirectoryPath(target),
		nil,
	)
	if len(warnings) != 1 {
		t.Fatalf("expected one missing-file warning, got %+v", warnings)
	}
	if warnings[0].Kind != allowFileNotFound {
		t.Fatalf("expected allowFileNotFound, got %+v", warnings[0])
	}
}

func TestCollectAllowFileWarnings_GlobalIgnoreStillWarns(t *testing.T) {
	program := createTestProgram(t, map[string]string{
		"src/app.ts": "const value = 1;",
	})
	target := findProgramFileForTest(t, program, "src/app.ts")
	configDir := tspath.GetDirectoryPath(tspath.GetDirectoryPath(target))

	warnings := collectAllowFileWarnings(
		[]string{target},
		rslintconfig.RslintConfig{
			{Ignores: []string{"src/**"}},
			{Rules: rslintconfig.Rules{"no-console": "error"}},
		},
		configDir,
		nil,
	)
	if len(warnings) != 1 {
		t.Fatalf("expected one warning, got %+v", warnings)
	}
	if warnings[0].Kind != allowFileIgnored {
		t.Fatalf("expected allowFileIgnored, got %+v", warnings[0])
	}
}

func TestCollectAllowFileWarnings_DefaultExcludedFileWarns(t *testing.T) {
	dir := t.TempDir()
	target := tspath.NormalizePath(filepath.Join(dir, "node_modules/pkg/a.ts"))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(target, []byte("const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	warnings := collectAllowFileWarnings(
		[]string{target},
		rslintconfig.RslintConfig{
			{Rules: rslintconfig.Rules{"no-console": "error"}},
		},
		tspath.NormalizePath(dir),
		nil,
	)
	if len(warnings) != 1 {
		t.Fatalf("expected one warning, got %+v", warnings)
	}
	if warnings[0].Kind != allowFileIgnored {
		t.Fatalf("expected allowFileIgnored, got %+v", warnings[0])
	}
}

func TestCLIRuleOverlayDoesNotAlterTargetDiscovery(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	write("a.ts", "debugger;\n")
	write("b.js", "debugger;\n")

	baseConfig := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		Rules: rslintconfig.Rules{"no-debugger": "off"},
	}}
	targetConfig := append(rslintconfig.RslintConfig(nil), baseConfig...)
	cliEntry, err := rslintconfig.BuildCLIRuleEntry([]string{"no-debugger: error"})
	if err != nil {
		t.Fatalf("BuildCLIRuleEntry: %v", err)
	}
	activeConfig := append(append(rslintconfig.RslintConfig(nil), baseConfig...), *cliEntry)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	programSet, err := createProgramSetForConfig(dir, activeConfig, true, fs, parseCache)
	if err != nil {
		t.Fatalf("createProgramSetForConfig: %v", err)
	}
	targetPlan, err := resolveLintTargetPlan(targetConfig, dir, fs, nil, []string{tspath.NormalizePath(dir)}, true)
	if err != nil {
		t.Fatalf("resolveLintTargetPlan: %v", err)
	}
	binding, err := bindLintTargetPlan(programSet, targetPlan, dir, fs, parseCache, true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}
	programs := binding.Programs
	typeInfoFiles := bindingTypeInfoFiles(binding)
	targetsByProgram := bindingTargetsByProgram(binding)
	targetFiles := make([]string, 0, len(targetPlan.Targets))
	for _, target := range targetPlan.Targets {
		targetFiles = append(targetFiles, target.Path)
	}
	if len(targetFiles) != 2 || !strings.HasSuffix(targetFiles[0], "/a.ts") || !strings.HasSuffix(targetFiles[1], "/b.js") {
		t.Fatalf("target discovery should retain the default baseline despite --rule overlay, got %v", targetFiles)
	}

	rslintconfig.RegisterAllRules()
	var diagnostics []rule.RuleDiagnostic
	_, err = linter.RunLinter(linter.RunLinterOptions{
		Programs:       programs,
		SingleThreaded: true,
		TargetFiles:    targetsByProgram,
		TypeInfoFiles:  typeInfoFiles,
		GetRulesForFile: func(sf *ast.SourceFile) []linter.ConfiguredRule {
			return rslintconfig.GlobalRuleRegistry.GetActiveRulesForFile(activeConfig, sf.FileName(), dir, false, typeInfoFiles)
		},
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, d)
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if len(diagnostics) != 2 {
		t.Fatalf("expected no-debugger diagnostics on both baseline targets, got %+v", diagnostics)
	}
	diagnosticFiles := []string{diagnostics[0].FilePath, diagnostics[1].FilePath}
	sort.Strings(diagnosticFiles)
	if !strings.HasSuffix(diagnosticFiles[0], "/a.ts") || !strings.HasSuffix(diagnosticFiles[1], "/b.js") {
		t.Fatalf("expected diagnostics on a.ts and b.js, got %+v", diagnostics)
	}
}

func TestPlainLintSkipsProjectResolutionWhenAllTargetsAreIgnored(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "rslint.json")
	target := filepath.Join(dir, "ignored.ts")
	if err := os.WriteFile(configPath, []byte(`[
		{"ignores":["ignored.ts"]},
		{
			"languageOptions":{"parserOptions":{"project":["./missing.json"]}},
			"rules":{"no-debugger":"error"}
		}
	]`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	// Deliberately malformed fixture: a global ignore removes it before Program
	// creation and syntax diagnostics, rather than treating parse failure as an
	// implicit ignore.
	if err := os.WriteFile(target, []byte("const = ;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	code, _, stderr := runLintPipelineForTest(t, dir, lintArgs{
		Config:         configPath,
		Format:         "default",
		AllowFiles:     []string{tspath.NormalizePath(target)},
		SingleThreaded: true,
	})
	if code != 0 {
		t.Fatalf("plain lint resolved an inactive project: code=%d stderr=%s", code, stderr)
	}

	code, _, stderr = runLintPipelineForTest(t, dir, lintArgs{
		Config:         configPath,
		Format:         "default",
		AllowFiles:     []string{tspath.NormalizePath(target)},
		SingleThreaded: true,
		TypeCheck:      true,
	})
	if code == 0 || !strings.Contains(stderr, "missing.json") {
		t.Fatalf("type-check must resolve every configured project: code=%d stderr=%s", code, stderr)
	}
}

func TestCLIExplicitJSONConfigNoArgsScopesToInvocationCWD(t *testing.T) {
	dir := t.TempDir()
	childDir := filepath.Join(dir, "child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatalf("mkdir child: %v", err)
	}
	write := func(base, name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(base, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	write(dir, "rslint.jsonc", `[{ "files": ["*.js"], "rules": { "no-debugger": "error" } }]`)
	write(dir, "parent.js", "debugger;\n")
	write(childDir, "child.js", "debugger;\n")

	code, stdout, stderr := runLintPipelineForTest(t, childDir, lintArgs{
		Config:         "../rslint.jsonc",
		Format:         "jsonline",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 1 {
		t.Fatalf("expected no-debugger to fail on child.js, got code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, `"filePath":"child.js"`) {
		t.Fatalf("expected child.js diagnostic relative to invocation cwd, stdout=%q stderr=%q", stdout, stderr)
	}
	if strings.Contains(stdout, "parent.js") {
		t.Fatalf("explicit --config must not widen no-args scope to config dir, stdout=%q", stdout)
	}
}

func TestCLIExplicitFileOutsideFilesCountsWithNoRules(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rslint.jsonc"), []byte(`[
		{ "files": ["**/*.ts"], "rules": { "no-debugger": "error" } }
	]`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	explicit := tspath.NormalizePath(filepath.Join(dir, "explicit.js"))
	if err := os.WriteFile(explicit, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write explicit file: %v", err)
	}

	code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
		Config:         "rslint.jsonc",
		Format:         "default",
		NoColor:        true,
		SingleThreaded: true,
		AllowFiles:     []string{explicit},
	})
	if code != 0 {
		t.Fatalf("expected explicit files-scope miss to exit cleanly, got code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "linted 1 file with 0 rules") {
		t.Fatalf("expected the explicit file to be counted with zero matching rules, stdout=%q stderr=%q", stdout, stderr)
	}
	if strings.Contains(stdout, "no-debugger") {
		t.Fatalf("files-scope miss must not run no-debugger, stdout=%q", stdout)
	}
}

func TestCLIExplicitMalformedFileOutsideFilesReportsSyntaxDiagnostic(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rslint.jsonc"), []byte(`[
		{ "files": ["**/*.ts"], "rules": { "no-debugger": "error" } }
	]`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	explicit := tspath.NormalizePath(filepath.Join(dir, "explicit.js"))
	if err := os.WriteFile(explicit, []byte("debugger;\nconst = ;\n"), 0o644); err != nil {
		t.Fatalf("write explicit file: %v", err)
	}

	code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
		Config:         "rslint.jsonc",
		Format:         "default",
		NoColor:        true,
		SingleThreaded: true,
		AllowFiles:     []string{explicit},
	})
	if code != 1 {
		t.Fatalf("expected malformed explicit target to exit 1, got code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "linted 1 file with 0 rules") {
		t.Fatalf("expected the explicit file to be counted with zero matching rules, stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stdout, "TypeScript(TS1134)") {
		t.Fatalf("selected zero-rule target must surface syntax diagnostics, stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestCLIExplicitMalformedFileWithRuleOverlayReportsSyntaxDiagnostic(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rslint.jsonc"), []byte(`[
		{ "files": ["**/*.ts"], "rules": {} }
	]`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	explicit := tspath.NormalizePath(filepath.Join(dir, "explicit.js"))
	if err := os.WriteFile(explicit, []byte("const = ;\n"), 0o644); err != nil {
		t.Fatalf("write explicit file: %v", err)
	}

	code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
		Config:         "rslint.jsonc",
		Format:         "default",
		NoColor:        true,
		SingleThreaded: true,
		AllowFiles:     []string{explicit},
		RuleFlags:      []string{"no-debugger:error"},
	})
	if code != 1 {
		t.Fatalf("expected syntax diagnostic to fail, got code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "TypeScript(TS") || !strings.Contains(stdout, "explicit.js") {
		t.Fatalf("expected syntax diagnostic for explicit.js, stdout=%q stderr=%q", stdout, stderr)
	}
	if strings.Contains(stdout, "no-debugger") {
		t.Fatalf("rules must not run when parsing fails, stdout=%q stderr=%q", stdout, stderr)
	}
}

func findProgramFileForTest(t *testing.T, program *compiler.Program, suffix string) string {
	t.Helper()
	normalizedSuffix := strings.ReplaceAll(suffix, "\\", "/")
	for _, sf := range program.GetSourceFiles() {
		name := sf.FileName()
		if strings.HasSuffix(name, normalizedSuffix) {
			return name
		}
	}
	t.Fatalf("program file with suffix %q not found", suffix)
	return ""
}

// TestGitlabReportState_EmptyProducesEmptyArray verifies a run with no
// diagnostics still produces a valid (empty) JSON array, not an empty file.
func TestGitlabReportState_EmptyProducesEmptyArray(t *testing.T) {
	var buf bytes.Buffer
	report := output.NewReport(nil, output.Metadata{})
	if err := output.Render(&buf, report, output.Options{Format: output.FormatGitLab}); err != nil {
		t.Fatalf("render empty GitLab report: %v", err)
	}

	if got := buf.String(); got != "[]\n" {
		t.Errorf("expected %q, got %q", "[]\n", got)
	}
}

// TestPrintDiagnosticGitLab verifies the gitlab format emits a single valid
// JSON array with the fields GitLab's Code Quality report requires:
// https://docs.gitlab.com/ci/testing/code_quality/
func TestPrintDiagnosticGitLab(t *testing.T) {
	source := "const unused = 42;\n"
	startOffset := strings.Index(source, "unused")
	diagWarning, opts := createTestDiagnostic(t, source, startOffset, startOffset+len("unused"))
	diagWarning.Severity = rule.SeverityWarning
	diagWarning.RuleName = "no-unused-vars"
	diagWarning.Message = rule.RuleMessage{Id: "test", Description: "'unused' is never read."}

	diagError, _ := createTestDiagnostic(t, source, 0, len("const"))
	diagError.Severity = rule.SeverityError
	diagError.RuleName = "prefer-const"
	diagError.Message = rule.RuleMessage{Id: "test", Description: "Use const."}

	var buf bytes.Buffer
	report := output.NewReport([]rule.RuleDiagnostic{diagWarning, diagError}, output.Metadata{})
	if err := output.Render(&buf, report, output.Options{
		Format:       output.FormatGitLab,
		ComparePaths: opts,
	}); err != nil {
		t.Fatalf("render GitLab report: %v", err)
	}

	var issues []struct {
		Description string `json:"description"`
		CheckName   string `json:"check_name"`
		Fingerprint string `json:"fingerprint"`
		Severity    string `json:"severity"`
		Location    struct {
			Path  string `json:"path"`
			Lines struct {
				Begin int `json:"begin"`
				End   int `json:"end"`
			} `json:"lines"`
			Positions struct {
				Begin struct {
					Line   int `json:"line"`
					Column int `json:"column"`
				} `json:"begin"`
				End struct {
					Line   int `json:"line"`
					Column int `json:"column"`
				} `json:"end"`
			} `json:"positions"`
		} `json:"location"`
	}
	if err := json.Unmarshal(buf.Bytes(), &issues); err != nil {
		t.Fatalf("output is not a valid JSON array: %v\noutput: %s", err, buf.String())
	}
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(issues))
	}

	if issues[0].Severity != "minor" {
		t.Errorf("warning should map to severity 'minor', got %q", issues[0].Severity)
	}
	if issues[1].Severity != "major" {
		t.Errorf("error should map to severity 'major', got %q", issues[1].Severity)
	}
	if issues[0].CheckName != "no-unused-vars" {
		t.Errorf("unexpected check_name: %q", issues[0].CheckName)
	}
	if issues[0].Location.Path != "index.ts" {
		t.Errorf("expected relative path 'index.ts', got %q", issues[0].Location.Path)
	}
	if issues[0].Fingerprint == "" || issues[0].Fingerprint == issues[1].Fingerprint {
		t.Errorf("each issue should have a distinct, non-empty fingerprint, got %q and %q", issues[0].Fingerprint, issues[1].Fingerprint)
	}
	if issues[0].Location.Positions.Begin.Line != 1 {
		t.Errorf("expected 1-based line number, got %d", issues[0].Location.Positions.Begin.Line)
	}
}

func TestDeduplicateTypeScriptDiagnosticsAcrossPathAliases(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "real.ts")
	aliasPath := filepath.Join(dir, "alias.ts")
	if err := os.WriteFile(realPath, []byte("let value: = 1;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realPath, aliasPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	base := rule.RuleDiagnostic{
		RuleName: "TypeScript(TS1110)",
		Origin:   rule.DiagnosticOriginTypeScript,
		Range:    core.NewTextRange(11, 11),
		Message:  rule.RuleMessage{Description: "Type expected."},
	}
	realDiagnostic := base
	realDiagnostic.FilePath = realPath
	aliasDiagnostic := base
	aliasDiagnostic.FilePath = aliasPath
	nonTypeScriptA := base
	nonTypeScriptA.RuleName = "some-rule"
	nonTypeScriptA.Origin = rule.DiagnosticOriginLint
	nonTypeScriptA.FilePath = realPath
	nonTypeScriptB := nonTypeScriptA
	nonTypeScriptB.FilePath = aliasPath

	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	for name, diagnostics := range map[string][]rule.RuleDiagnostic{
		"real-first":  {realDiagnostic, aliasDiagnostic, nonTypeScriptA, nonTypeScriptB},
		"alias-first": {aliasDiagnostic, realDiagnostic, nonTypeScriptA, nonTypeScriptB},
	} {
		t.Run(name, func(t *testing.T) {
			got := deduplicateTypeScriptDiagnostics(diagnostics, fsys)
			if len(got) != 3 {
				t.Fatalf("expected one TypeScript diagnostic and both rule diagnostics, got %d: %+v", len(got), got)
			}
			if got[0].FilePath != aliasPath {
				t.Fatalf("expected deterministic lexical alias survivor %q, got %q", aliasPath, got[0].FilePath)
			}
		})
	}
}

func TestDeduplicateTypeScriptDiagnosticsPrefersCallerTarget(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "z-real.ts")
	aliasPath := filepath.Join(dir, "a-alias.ts")
	if err := os.WriteFile(realPath, []byte("let value: = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realPath, aliasPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	base := rule.RuleDiagnostic{
		RuleName: "TypeScript(TS1110)",
		Origin:   rule.DiagnosticOriginTypeScript,
		Range:    core.NewTextRange(11, 11),
		Message:  rule.RuleMessage{Description: "Type expected."},
	}
	realDiagnostic := base
	realDiagnostic.FilePath = realPath
	aliasDiagnostic := base
	aliasDiagnostic.FilePath = aliasPath
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	preferred := map[string]string{canonicalHostPathID(realPath, fsys): realPath}

	got := deduplicateTypeScriptDiagnostics([]rule.RuleDiagnostic{aliasDiagnostic, realDiagnostic}, fsys, preferred)
	if len(got) != 1 || got[0].FilePath != realPath {
		t.Fatalf("expected caller-selected target %q to win over lexical alias, got %+v", realPath, got)
	}
	single := deduplicateTypeScriptDiagnostics([]rule.RuleDiagnostic{aliasDiagnostic}, fsys, preferred)
	if single[0].FilePath != realPath {
		t.Fatalf("expected a single aliased diagnostic to use caller target %q, got %+v", realPath, single)
	}
}

func TestWriteIndependentFixOutputsReturnsWriteError(t *testing.T) {
	directoryPath := t.TempDir()
	err := writeIndependentFixOutputs([]independentFixResult{{
		path: directoryPath, content: "b", didFix: true,
	}}, false)
	if err == nil {
		t.Fatal("expected a write error")
	}
	if !strings.Contains(err.Error(), directoryPath) {
		t.Fatalf("write error must identify the target path, got %v", err)
	}
}

func TestWriteIndependentFixOutputsSerializesLexicalAliasesWhenSingleThreaded(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory symlink setup requires privileges on Windows")
	}
	const source = `var value = "x";`
	root := t.TempDir()
	physicalDir := filepath.Join(root, "physical")
	if err := os.MkdirAll(physicalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	physical := filepath.Join(physicalDir, "file.js")
	if err := os.WriteFile(physical, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	aliasA := filepath.Join(root, "alias-a")
	aliasB := filepath.Join(root, "alias-b")
	if err := os.Symlink(physicalDir, aliasA); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(physicalDir, aliasB); err != nil {
		t.Fatal(err)
	}
	pathA := filepath.Join(aliasA, "file.js")
	pathB := filepath.Join(aliasB, "file.js")

	wantA := `var value = 'x';`
	wantB := `let value = "x";`
	if err := writeIndependentFixOutputs([]independentFixResult{
		{path: pathA, content: wantA, didFix: true},
		{path: pathB, content: wantB, didFix: true},
	}, true); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(physical)
	if err != nil {
		t.Fatal(err)
	}
	got := string(content)
	if got != wantB {
		t.Fatalf("single-threaded physical output = %q, want final complete lexical output %q (first was %q)", got, wantB, wantA)
	}
}

func TestExecuteLintPipelineFixesLexicalAliasesInIsolatedMemory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory symlink setup requires privileges on Windows")
	}
	root := t.TempDir()
	physicalDir := filepath.Join(root, "physical")
	if err := os.MkdirAll(physicalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	physical := filepath.Join(physicalDir, "file.ts")
	const source = "const pattern = /bar  baz/; const obj = { ['x']: 1 };\n"
	if err := os.WriteFile(physical, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	aliasA := filepath.Join(root, "alias-a")
	aliasB := filepath.Join(root, "alias-b")
	if err := os.Symlink(physicalDir, aliasA); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := os.Symlink(physicalDir, aliasB); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	pathA := filepath.Join(aliasA, "file.ts")
	pathB := filepath.Join(aliasB, "file.ts")
	configA := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		Rules: rslintconfig.Rules{"no-regex-spaces": "error"},
	}}
	configB := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		Rules: rslintconfig.Rules{"no-useless-computed-key": "error"},
	}}
	catalog := &discovery.ConfigCatalog{
		Configs: map[string]rslintconfig.RslintConfig{
			aliasA: configA,
			aliasB: configB,
		},
		Targets: []discovery.DiscoveredTarget{
			{
				Path: pathA, ConfigDirectory: aliasA,
				MergedConfig: configA.GetConfigForFile(pathA, aliasA),
			},
			{
				Path: pathB, ConfigDirectory: aliasB,
				MergedConfig: configB.GetConfigForFile(pathB, aliasB),
			},
		},
	}
	code, stdout, stderr := runLintPipelineForTest(t, root, lintArgs{
		ConfigCatalog:  catalog,
		Fix:            true,
		Format:         "jsonline",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 0 || stdout != "" || stderr != "" {
		t.Fatalf("fix pipeline failed: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	content, err := os.ReadFile(physical)
	if err != nil {
		t.Fatal(err)
	}
	wantA := "const pattern = /bar {2}baz/; const obj = { ['x']: 1 };\n"
	wantB := "const pattern = /bar  baz/; const obj = { 'x': 1 };\n"
	if got := string(content); got != wantB {
		t.Fatalf("single-threaded physical output = %q, want final isolated lexical output %q (first was %q)", got, wantB, wantA)
	}
}

func TestExecuteLintPipelinePreservesNativePOSIXBackslashTargetThroughCompilerAlias(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows uses backslash as a path separator")
	}
	root := t.TempDir()
	target := filepath.Join(root, `a\b.ts`)
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", "b.ts"), []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	entries := rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"no-debugger": "error"}}}
	catalog := &discovery.ConfigCatalog{
		Configs: map[string]rslintconfig.RslintConfig{root: entries},
		Targets: []discovery.DiscoveredTarget{{
			Path: target, ConfigDirectory: root,
			MergedConfig: entries.GetConfigForFile(target, root),
		}},
	}

	code, stdout, stderr := runLintPipelineForTest(t, root, lintArgs{
		ConfigCatalog:  catalog,
		Format:         "jsonline",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 1 || stderr != "" {
		t.Fatalf("lint pipeline failed: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	var diagnostic struct {
		RuleName string `json:"ruleName"`
		FilePath string `json:"filePath"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &diagnostic); err != nil {
		t.Fatalf("decode jsonline diagnostic %q: %v", stdout, err)
	}
	if diagnostic.RuleName != "no-debugger" || diagnostic.FilePath != `a\b.ts` {
		t.Fatalf("diagnostic lost native lexical target: %+v", diagnostic)
	}
}
