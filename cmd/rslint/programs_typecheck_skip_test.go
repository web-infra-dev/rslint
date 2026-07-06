package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func writeProgramTestFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}
}

func collectProgramTypeDiagnostics(
	t *testing.T,
	programs []*compiler.Program,
	skip []bool,
	typeInfoFiles map[string]struct{},
) []rule.RuleDiagnostic {
	t.Helper()

	var diags []rule.RuleDiagnostic
	_, err := linter.RunLinter(linter.RunLinterOptions{
		Programs:              programs,
		SingleThreaded:        true,
		TypeCheck:             true,
		SkipTypeCheckPrograms: skip,
		TypeInfoFiles:         typeInfoFiles,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			diags = append(diags, d)
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	return diags
}

func containsTSDiagnostic(diags []rule.RuleDiagnostic, code string) bool {
	needle := "TypeScript(" + code + ")"
	for _, d := range diags {
		if d.RuleName == needle {
			return true
		}
	}
	return false
}

func TestTypeCheck_SkipsNoTsconfigFallbackPrograms(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"rslint.config.ts": `import type { Bad } from './bad';
export default [{ files: ['**/*.ts'], rules: {} }] satisfies Bad;
`,
		"bad.d.ts": `export type Bad = MissingGlobalType;
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	cfg := rslintconfig.RslintConfig{{Files: []string{"**/*.ts"}}}
	programs, hasTsConfig, exitCode := createProgramsForConfig(
		dir,
		cfg,
		true,
		fs,
		nil,
		parseCache,
	)
	if exitCode != 0 {
		t.Fatalf("createProgramsForConfig exit code = %d", exitCode)
	}
	if hasTsConfig {
		t.Fatalf("expected no tsconfig to be found")
	}
	if len(programs) != 0 {
		t.Fatalf("expected no programs before gap fallback, got %d", len(programs))
	}

	programs, _, _, _, _ = buildProgramsWithGapFallback(
		programs, nil, cfg, dir, fs, nil, nil, parseCache, true, hasTsConfig, nil,
	)
	if len(programs) != 1 {
		t.Fatalf("expected one synthesized fallback program, got %d", len(programs))
	}
	if got := programs[0].Options().ConfigFilePath; got != "" {
		t.Fatalf("expected synthesized fallback program to have no ConfigFilePath, got %q", got)
	}
	skip := buildTypeCheckSkipMask(programs)
	if len(skip) != 1 || !skip[0] {
		t.Fatalf("expected no-tsconfig fallback program to be skipped for type-check, got %v", skip)
	}

	if diags := collectProgramTypeDiagnostics(t, programs, skip, nil); containsTSDiagnostic(diags, "TS2304") {
		t.Fatalf("did not expect declaration diagnostics from a no-tsconfig fallback program: %+v", diags)
	}

	// Prove the fixture is capable of surfacing the bad declaration if the
	// synthesized Program is not skipped. This pins the regression directly.
	if diags := collectProgramTypeDiagnostics(t, programs, nil, nil); !containsTSDiagnostic(diags, "TS2304") {
		t.Fatalf("expected TS2304 without the skip mask, got: %+v", diags)
	}
}

func TestTypeCheck_TsconfigBackedProgramReportsCoveredDeclarationErrors(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"tsconfig.json": `{
  "compilerOptions": { "skipLibCheck": false },
  "include": ["rslint.config.ts"]
}
`,
		"rslint.config.ts": `import type { Bad } from './bad';
export const value: Bad | null = null;
`,
		"bad.d.ts": `export type Bad = MissingGlobalType;
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	programs, hasTsConfig, exitCode := createProgramsForConfig(
		dir,
		rslintconfig.RslintConfig{{
			Files: []string{"**/*.ts"},
			LanguageOptions: &rslintconfig.LanguageOptions{
				ParserOptions: &rslintconfig.ParserOptions{
					Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
				},
			},
		}},
		true,
		fs,
		nil,
		utils.NewParseCache(),
	)
	if exitCode != 0 {
		t.Fatalf("createProgramsForConfig exit code = %d", exitCode)
	}
	if !hasTsConfig {
		t.Fatalf("expected tsconfig to be found")
	}
	if len(programs) != 1 {
		t.Fatalf("expected one tsconfig-backed program, got %d", len(programs))
	}
	if got := programs[0].Options().ConfigFilePath; got == "" {
		t.Fatal("expected tsconfig-backed program to carry ConfigFilePath")
	}
	skip := buildTypeCheckSkipMask(programs)
	if skip != nil {
		t.Fatalf("expected tsconfig-backed program to participate in type-check, got %v", skip)
	}

	diags := collectProgramTypeDiagnostics(t, programs, skip, nil)
	if !containsTSDiagnostic(diags, "TS2304") {
		var rendered []string
		for _, d := range diags {
			rendered = append(rendered, d.RuleName+": "+d.Message.Description)
		}
		t.Fatalf("expected TS2304 from the tsconfig-covered declaration graph, got:\n%s", strings.Join(rendered, "\n"))
	}
}

func TestTypeCheck_SkipsGapFallbackPrograms(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"tsconfig.json": `{
  "compilerOptions": { "skipLibCheck": false },
  "include": ["src/in-project.ts"]
}
`,
		"src/in-project.ts": `export const ok = 1;
`,
		"gap.ts": `import type { Bad } from './bad';
export const value: Bad | null = null;
`,
		"bad.d.ts": `export type Bad = MissingGlobalType;
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	parseCache := utils.NewParseCache()
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
			},
		},
	}}
	programs, hasTsConfig, exitCode := createProgramsForConfig(
		dir,
		cfg,
		true,
		fs,
		nil,
		parseCache,
	)
	if exitCode != 0 {
		t.Fatalf("createProgramsForConfig exit code = %d", exitCode)
	}
	if !hasTsConfig {
		t.Fatalf("expected tsconfig to be found")
	}
	if len(programs) != 1 {
		t.Fatalf("expected one tsconfig-backed program before gap fallback, got %d", len(programs))
	}
	if skip := buildTypeCheckSkipMask(programs); skip != nil {
		t.Fatalf("expected tsconfig-backed program to participate in type-check, got %v", skip)
	}

	programs, typeInfoFiles, gapFiles, _, _ := buildProgramsWithGapFallback(
		programs,
		nil,
		cfg,
		dir,
		fs,
		nil,
		nil,
		parseCache,
		true,
		hasTsConfig,
		nil,
	)
	if len(programs) != 2 {
		t.Fatalf("expected the gap fallback program to be appended, got %d programs", len(programs))
	}
	gapPath := filepath.ToSlash(filepath.Join(dir, "gap.ts"))
	if !slices.Contains(gapFiles, gapPath) {
		t.Fatalf("expected gap.ts to be discovered as a gap file, got %v", gapFiles)
	}
	if got := programs[0].Options().ConfigFilePath; got == "" {
		t.Fatal("expected original tsconfig-backed program to carry ConfigFilePath")
	}
	if got := programs[1].Options().ConfigFilePath; got != "" {
		t.Fatalf("expected gap fallback program to have no ConfigFilePath, got %q", got)
	}
	skip := buildTypeCheckSkipMask(programs)
	if len(skip) != 2 || skip[0] || !skip[1] {
		t.Fatalf("expected only the gap fallback program to be skipped, got %v", skip)
	}

	if diags := collectProgramTypeDiagnostics(t, programs, skip, typeInfoFiles); containsTSDiagnostic(diags, "TS2304") {
		t.Fatalf("did not expect declaration diagnostics from a gap fallback program: %+v", diags)
	}
}
