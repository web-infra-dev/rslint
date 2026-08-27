package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/web-infra-dev/rslint/internal/api"
	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestHandleLint_FixReturnsFinalGeneration(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"rules": { "@typescript-eslint/array-type": "error" },
		"plugins": ["@typescript-eslint"]
	}]`)
	target := filepath.Join(fixturesDir, "gap-fix.ts")

	response, err := (&Handler{}).HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
		Files:            []string{target},
		FileContents:     map[string]string{target: "let a: Array<string> = [];\n"},
		Fix:              true,
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.ErrorCount != 0 || response.FixableErrorCount != 0 || response.FixableWarningCount != 0 {
		t.Fatalf(
			"final counts = errors:%d fixable:%d/%d, want zero",
			response.ErrorCount,
			response.FixableErrorCount,
			response.FixableWarningCount,
		)
	}
	if len(response.Diagnostics) != 0 {
		t.Fatalf("final diagnostics = %+v, want none", response.Diagnostics)
	}
	if got := response.Output["gap-fix.ts"]; got != "let a: string[] = [];\n" {
		t.Fatalf("fixed output = %q", got)
	}
}

func TestHandleLint_FixOutputPreservesOverlayBOM(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"rules": { "@typescript-eslint/array-type": "error" },
		"plugins": ["@typescript-eslint"]
	}]`)
	target := filepath.Join(fixturesDir, "gap-fix-bom.ts")
	response, err := (&Handler{}).HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
		Files:            []string{target},
		FileContents:     map[string]string{target: utils.BOM + "let a: Array<string> = [];\n"},
		Fix:              true,
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	want := utils.BOM + "let a: string[] = [];\n"
	if got := response.Output["gap-fix-bom.ts"]; got != want {
		t.Fatalf("fixed BOM output = %q, want %q", got, want)
	}
}

func TestHandleLint_FixRunsNativeCascadesWithoutWritingDisk(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "input.ts")
	const source = "const value: Number = 1;\n"
	if err := os.WriteFile(target, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(`{"files":["input.ts"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"plugins": ["@typescript-eslint"],
		"rules": {
			"@typescript-eslint/no-wrapper-object-types": "error",
			"@typescript-eslint/no-inferrable-types": "error"
		}
	}]`)

	response, err := (&Handler{}).HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
		Fix:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := response.Output["input.ts"]; got != "const value = 1;\n" {
		t.Fatalf("cascade output = %q, want the second-round fix", got)
	}
	if len(response.Diagnostics) != 0 || response.ErrorCount != 0 || response.FixableErrorCount != 0 {
		t.Fatalf(
			"final diagnostics/counts are stale: %+v, errors=%d fixable=%d",
			response.Diagnostics,
			response.ErrorCount,
			response.FixableErrorCount,
		)
	}
	diskContent, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(diskContent) != source {
		t.Fatalf("API wrote fixes to disk: got %q, want %q", diskContent, source)
	}
}

func TestHandleLint_FixKeepsFilesIsolatedAcrossRounds(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first.js")
	second := filepath.Join(dir, "second.js")
	clean := filepath.Join(dir, "clean.js")
	response, err := (&Handler{}).HandleLint(api.LintRequest{
		Config:           json.RawMessage(`[{"rules":{"no-var":"error"}}]`),
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{first, second, clean},
		FileContents: map[string]string{
			first:  "var first = 1;\n",
			second: "var second = 2;\n",
			clean:  "const clean = 3;\n",
		},
		Fix: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Output) != 2 {
		t.Fatalf("output files = %+v, want only the two changed files", response.Output)
	}
	if got := response.Output["first.js"]; strings.Contains(got, "var") || !strings.Contains(got, "first") {
		t.Fatalf("first output was lost or contaminated: %q", got)
	}
	if got := response.Output["second.js"]; strings.Contains(got, "var") || !strings.Contains(got, "second") {
		t.Fatalf("second output was lost or contaminated: %q", got)
	}
	if _, exists := response.Output["clean.js"]; exists {
		t.Fatalf("unchanged file unexpectedly has output: %+v", response.Output)
	}
	if len(response.Diagnostics) != 0 {
		t.Fatalf("final diagnostics = %+v, want none", response.Diagnostics)
	}
}

func TestHandleLint_FixVerifiesAfterTenRounds(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "input.js")
	config := json.RawMessage(`[{"plugins":["community"],"rules":{"community/increment":"error"}}]`)

	var calls atomic.Int32
	requester := apiRequesterFunc(func(_ context.Context, kind ipc.MessageKind, payload any) (*ipc.Message, error) {
		if kind != api.KindPluginLint {
			return nil, fmt.Errorf("unexpected reverse request kind %q", kind)
		}
		request := payload.(linter.EslintPluginLintRequest)
		call := int(calls.Add(1))
		if len(request.Files) != 1 || request.Files[0].Text == nil {
			t.Fatalf("missing plugin source on call %d", call)
		}
		text := *request.Files[0].Text
		value, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			t.Fatalf("parse generation %q: %v", text, err)
		}
		if value != call-1 {
			t.Fatalf("plugin call %d saw generation %d", call, value)
		}
		return ipc.NewMessage(ipc.KindResponse, 1, linter.EslintPluginLintResult{
			Results: []linter.EslintPluginFileResult{{
				FilePath: request.Files[0].Path,
				Diagnostics: []linter.EslintPluginDiagnostic{{
					RuleName: "community/increment",
					Message:  "increment",
					StartPos: 0,
					EndPos:   len(strings.TrimSpace(text)),
					Fixes: []linter.EslintPluginFix{{
						Range: [2]int{0, len(strings.TrimSpace(text))},
						Text:  strconv.Itoa(value + 1),
					}},
				}},
			}},
		})
	})

	response, err := (&Handler{}).HandleLintWithContext(context.Background(), api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
		FileContents:     map[string]string{target: "0\n"},
		EslintPlugins: []api.EslintPluginEntry{{
			Prefix:    "community",
			RuleNames: []string{"increment"},
		}},
		Fix: true,
	}, requester)
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 11 {
		t.Fatalf("plugin calls = %d, want ten writable rounds plus final verification", calls.Load())
	}
	if got := response.Output["input.js"]; got != "10\n" {
		t.Fatalf("output = %q, want tenth write", got)
	}
	if len(response.Diagnostics) != 1 || response.ErrorCount != 1 || response.FixableErrorCount != 1 {
		t.Fatalf(
			"final verification was not projected: diagnostics=%+v errors=%d fixable=%d",
			response.Diagnostics,
			response.ErrorCount,
			response.FixableErrorCount,
		)
	}
	fixes := response.Diagnostics[0].Fixes
	if len(fixes) != 1 || fixes[0].StartPos != 0 || fixes[0].EndPos != 2 || fixes[0].Text != "11" {
		t.Fatalf("final diagnostic fix = %+v, want range [0,2] replacement 11", fixes)
	}
}

func TestHandleLint_FixReportsOutputAfterNetRestoration(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "input.js")
	var calls atomic.Int32
	requester := apiRequesterFunc(func(_ context.Context, kind ipc.MessageKind, payload any) (*ipc.Message, error) {
		if kind != api.KindPluginLint {
			return nil, fmt.Errorf("unexpected reverse request kind %q", kind)
		}
		request := payload.(linter.EslintPluginLintRequest)
		call := calls.Add(1)
		if len(request.Files) != 1 || request.Files[0].Text == nil {
			t.Fatalf("plugin call %d has no source", call)
		}
		text := *request.Files[0].Text
		replacement := "b"
		if text == "b" {
			replacement = "a"
		}
		return ipc.NewMessage(ipc.KindResponse, 1, linter.EslintPluginLintResult{
			Results: []linter.EslintPluginFileResult{{
				FilePath: request.Files[0].Path,
				Diagnostics: []linter.EslintPluginDiagnostic{{
					RuleName: "community/toggle",
					Message:  "toggle",
					StartPos: 0,
					EndPos:   1,
					Fixes:    []linter.EslintPluginFix{{Range: [2]int{0, 1}, Text: replacement}},
				}},
			}},
		})
	})

	response, err := (&Handler{}).HandleLintWithContext(context.Background(), api.LintRequest{
		Config:           json.RawMessage(`[{"plugins":["community"],"rules":{"community/toggle":"error"}}]`),
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
		FileContents:     map[string]string{target: "a"},
		EslintPlugins: []api.EslintPluginEntry{{
			Prefix:    "community",
			RuleNames: []string{"toggle"},
		}},
		Fix: true,
	}, requester)
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("plugin calls = %d, want two rounds ending at the initial source", calls.Load())
	}
	output, present := response.Output["input.js"]
	if !present || output != "a" {
		t.Fatalf("restored output = %q (present=%t), want original text reported", output, present)
	}
	if len(response.Diagnostics) != 1 || response.FixableErrorCount != 1 {
		t.Fatalf("restored final diagnostics = %+v", response.Diagnostics)
	}
}

func TestHandleLint_FixReportsEmptyOutput(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "input.js")
	var calls atomic.Int32
	requester := apiRequesterFunc(func(_ context.Context, kind ipc.MessageKind, payload any) (*ipc.Message, error) {
		if kind != api.KindPluginLint {
			return nil, fmt.Errorf("unexpected reverse request kind %q", kind)
		}
		request := payload.(linter.EslintPluginLintRequest)
		call := calls.Add(1)
		if len(request.Files) != 1 || request.Files[0].Text == nil {
			t.Fatalf("plugin call %d has no source", call)
		}
		text := *request.Files[0].Text
		result := linter.EslintPluginFileResult{FilePath: request.Files[0].Path}
		if text != "" {
			result.Diagnostics = []linter.EslintPluginDiagnostic{{
				RuleName: "community/empty",
				Message:  "empty",
				StartPos: 0,
				EndPos:   len(text),
				Fixes:    []linter.EslintPluginFix{{Range: [2]int{0, len(text)}, Text: ""}},
			}}
		}
		return ipc.NewMessage(ipc.KindResponse, 1, linter.EslintPluginLintResult{
			Results: []linter.EslintPluginFileResult{result},
		})
	})

	response, err := (&Handler{}).HandleLintWithContext(context.Background(), api.LintRequest{
		Config:           json.RawMessage(`[{"plugins":["community"],"rules":{"community/empty":"error"}}]`),
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
		FileContents:     map[string]string{target: "removeMe();"},
		EslintPlugins: []api.EslintPluginEntry{{
			Prefix:    "community",
			RuleNames: []string{"empty"},
		}},
		Fix: true,
	}, requester)
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("plugin calls = %d, want fix plus final verification", calls.Load())
	}
	output, present := response.Output["input.js"]
	if !present || output != "" {
		t.Fatalf("empty output = %q (present=%t), want an explicit empty string", output, present)
	}
	if len(response.Diagnostics) != 0 {
		t.Fatalf("final diagnostics = %+v, want none", response.Diagnostics)
	}
}

func TestHandleLint_FixProjectsOutputToRequestedAlias(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "real.ts")
	aliasPath := filepath.Join(dir, "alias.ts")
	const source = "const bad = 1;\n"
	if err := os.WriteFile(realPath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, "tsconfig.json"),
		[]byte(`{"compilerOptions":{"noLib":true},"files":["real.ts"]}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realPath, aliasPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	var calls atomic.Int32
	dispatch := func(_ context.Context, request linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		call := calls.Add(1)
		if len(request.Files) != 1 || request.Files[0].Text == nil {
			t.Fatalf("plugin call %d has no source", call)
		}
		text := *request.Files[0].Text
		result := linter.EslintPluginFileResult{FilePath: request.Files[0].Path}
		if strings.Contains(text, "bad") {
			result.Diagnostics = []linter.EslintPluginDiagnostic{{
				RuleName: "community/rename",
				Message:  "rename",
				StartPos: 6,
				EndPos:   9,
				Fixes:    []linter.EslintPluginFix{{Range: [2]int{6, 9}, Text: "good"}},
			}}
		}
		return &linter.EslintPluginLintResult{
			Results: []linter.EslintPluginFileResult{result},
		}, nil
	}

	response, err := (&Handler{}).handleLint(context.Background(), api.LintRequest{
		Files:          []string{aliasPath},
		CanonicalFiles: []string{realPath},
		Config: json.RawMessage(`[{
			"files":["**/*.ts"],
			"plugins":["community"],
			"languageOptions":{"parserOptions":{"project":["./tsconfig.json"]}},
			"rules":{"community/rename":"error"}
		}]`),
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		EslintPlugins: []api.EslintPluginEntry{{
			Prefix:    "community",
			RuleNames: []string{"rename"},
		}},
		Fix: true,
	}, dispatch, nil)
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("plugin calls = %d, want fix plus final verification", calls.Load())
	}
	if got := response.Output["alias.ts"]; got != "const good = 1;\n" {
		t.Fatalf("alias output = %q", got)
	}
	if _, present := response.Output["real.ts"]; present {
		t.Fatalf("output leaked the Program identity: %+v", response.Output)
	}
	if len(response.LintedFiles) != 1 || response.LintedFiles[0] != "alias.ts" {
		t.Fatalf("linted files = %+v, want requested alias", response.LintedFiles)
	}
	if len(response.Diagnostics) != 0 {
		t.Fatalf("final diagnostics = %+v, want none", response.Diagnostics)
	}
	disk, err := os.ReadFile(realPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(disk) != source {
		t.Fatalf("API wrote alias output to disk: %q", disk)
	}
}

func TestHandleLint_FixCancellationDuringFinalVerificationReturnsNoResponse(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "input.js")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls atomic.Int32
	requester := apiRequesterFunc(func(_ context.Context, kind ipc.MessageKind, payload any) (*ipc.Message, error) {
		if kind != api.KindPluginLint {
			return nil, fmt.Errorf("unexpected reverse request kind %q", kind)
		}
		request := payload.(linter.EslintPluginLintRequest)
		call := calls.Add(1)
		if len(request.Files) != 1 || request.Files[0].Text == nil {
			t.Fatalf("plugin call %d has no source", call)
		}
		if call == 2 {
			cancel()
			return nil, context.Canceled
		}
		return ipc.NewMessage(ipc.KindResponse, 1, linter.EslintPluginLintResult{
			Results: []linter.EslintPluginFileResult{{
				FilePath: request.Files[0].Path,
				Diagnostics: []linter.EslintPluginDiagnostic{{
					RuleName: "community/rename",
					Message:  "rename",
					StartPos: 6,
					EndPos:   9,
					Fixes:    []linter.EslintPluginFix{{Range: [2]int{6, 9}, Text: "good"}},
				}},
			}},
		})
	})

	response, err := (&Handler{}).HandleLintWithContext(ctx, api.LintRequest{
		Config:           json.RawMessage(`[{"plugins":["community"],"rules":{"community/rename":"error"}}]`),
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
		FileContents:     map[string]string{target: "const bad = 1;\n"},
		EslintPlugins: []api.EslintPluginEntry{{
			Prefix:    "community",
			RuleNames: []string{"rename"},
		}},
		Fix: true,
	}, requester)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if response != nil {
		t.Fatalf("canceled final verification exposed a partial response: %+v", response)
	}
	if calls.Load() != 2 {
		t.Fatalf("plugin calls = %d, want cancellation during final verification", calls.Load())
	}
}
