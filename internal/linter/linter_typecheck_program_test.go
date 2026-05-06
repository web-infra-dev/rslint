package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// createProgramFromTsconfigDir builds a Program from an existing tsconfig.json
// in dir. Use createTsconfigProject to set the directory up.
func createProgramFromTsconfigDir(t *testing.T, dir string) *compiler.Program {
	t.Helper()
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(dir, fs)
	program, err := utils.CreateProgram(true, fs, dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("create program in %s: %v", dir, err)
	}
	return program
}

// writeFiles creates files under dir relative to the directory.
func writeFiles(t *testing.T, dir string, files map[string]string) map[string]string {
	t.Helper()
	out := make(map[string]string, len(files))
	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
		out[rel] = tspath.NormalizePath(full)
	}
	return out
}

// New regression tests covering the program-scoped, tsc-aligned --type-check
// semantics introduced after the linter refactor.

// (1) Owned-set boundary: a `.ts` source file pulled in by import resolution
// but not listed in tsconfig include must still get its type errors reported.
func TestTypeCheck_ReportsErrorsInTransitivelyImportedSource(t *testing.T) {
	dir := t.TempDir()
	paths := writeFiles(t, dir, map[string]string{
		"entry.ts": `import {bad} from './lib';
const v: number = bad;
`,
		"lib.ts": `export const bad: string = 'oops';
`,
		"tsconfig.json": `{"include":["entry.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:       true,
		OnDiagnostic:    func(d rule.RuleDiagnostic) { diags = append(diags, d) },
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}

	gotEntryErr := false
	for _, d := range diags {
		if !strings.HasPrefix(d.RuleName, "TypeScript(") {
			continue
		}
		if d.SourceFile != nil && d.SourceFile.FileName() == paths["entry.ts"] {
			gotEntryErr = true
		}
	}
	if !gotEntryErr {
		t.Fatal("expected entry.ts to report a type error caused by lib.ts (transitive import)")
	}
}

// (2) skipLibCheck:false + a hand-written .d.ts that imports a missing
// package: the d.ts itself produces TS2307. This is the resolution-bundler
// scenario in miniature.
func TestTypeCheck_ReportsDeclarationFileErrors_WhenSkipLibCheckFalse(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"types.d.ts": `import * as Missing from 'this-package-does-not-exist';
export type T = typeof Missing;
`,
		"entry.ts": `import type {T} from './types';
export const x: T = {} as T;
`,
		"tsconfig.json": `{"compilerOptions":{"skipLibCheck":false},"include":["entry.ts","types.d.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:       true,
		OnDiagnostic:    func(d rule.RuleDiagnostic) { diags = append(diags, d) },
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}

	hasDtsErr := false
	for _, d := range diags {
		if d.RuleName == "TypeScript(TS2307)" && strings.Contains(d.SourceFile.FileName(), "types.d.ts") {
			hasDtsErr = true
			break
		}
	}
	if !hasDtsErr {
		t.Fatal("expected TS2307 from types.d.ts when skipLibCheck=false")
	}
}

// (3) skipLibCheck:true masks d.ts errors (typescript-go's SkipTypeChecking
// natural behaviour; we should not over-report).
func TestTypeCheck_SuppressesDeclarationFileErrors_WhenSkipLibCheckTrue(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"types.d.ts": `import * as Missing from 'this-package-does-not-exist';
export type T = typeof Missing;
`,
		"entry.ts": `import type {T} from './types';
export const x: T = {} as T;
`,
		"tsconfig.json": `{"compilerOptions":{"skipLibCheck":true},"include":["entry.ts","types.d.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:       true,
		OnDiagnostic:    func(d rule.RuleDiagnostic) { diags = append(diags, d) },
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}

	for _, d := range diags {
		if d.RuleName == "TypeScript(TS2307)" && d.SourceFile != nil && strings.HasSuffix(d.SourceFile.FileName(), "types.d.ts") {
			t.Errorf("did not expect TS2307 from types.d.ts under skipLibCheck:true, got: %s", d.Message.Description)
		}
	}
}

// (4) Cross-program dedup: same shared file, same error, two programs.
// Should be reported exactly once.
func TestTypeCheck_DedupsAcrossPrograms(t *testing.T) {
	root := t.TempDir()

	// Shared .ts file under root.
	sharedRel := "shared/util.ts"
	writeFiles(t, root, map[string]string{
		sharedRel: `export const broken: number = 'string';
`,
	})

	// Two sibling tsconfig dirs each include the same shared file via a
	// relative path.
	a := filepath.Join(root, "a")
	if err := os.MkdirAll(a, 0755); err != nil {
		t.Fatalf("mkdir a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(a, "tsconfig.json"),
		[]byte(`{"include":["../shared/util.ts"]}`), 0644); err != nil {
		t.Fatalf("write a/tsconfig: %v", err)
	}
	b := filepath.Join(root, "b")
	if err := os.MkdirAll(b, 0755); err != nil {
		t.Fatalf("mkdir b: %v", err)
	}
	if err := os.WriteFile(filepath.Join(b, "tsconfig.json"),
		[]byte(`{"include":["../shared/util.ts"]}`), 0644); err != nil {
		t.Fatalf("write b/tsconfig: %v", err)
	}

	progA := createProgramFromTsconfigDir(t, a)
	progB := createProgramFromTsconfigDir(t, b)

	var ts2322 int
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{progA, progB},
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:       true,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			if d.RuleName == "TypeScript(TS2322)" {
				ts2322++
			}
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}

	if ts2322 != 1 {
		t.Errorf("expected exactly 1 TS2322 (deduped across programs), got %d", ts2322)
	}
}

// (5) SkipTypeCheckPrograms entry true skips that program's type-check phase.
func TestTypeCheck_SkipTypeCheckProgramsHonored(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"a.ts": `const x: number = 'oops';
`,
		"tsconfig.json": `{"include":["a.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:              []*compiler.Program{program},
		SingleThreaded:        true,
		GetRulesForFile:       func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:             true,
		SkipTypeCheckPrograms: []bool{true},
		OnDiagnostic:          func(d rule.RuleDiagnostic) { diags = append(diags, d) },
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}

	for _, d := range diags {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			t.Fatalf("did not expect any TS diagnostic when program is in skip mask, got: %s", d.RuleName)
		}
	}
}

// (6) TS2578 ("unused @ts-expect-error") surfaces under the new path.
// Verifies that GetDiagnosticsOfAnyProgram emits this when the directive
// turns out unused.
func TestTypeCheck_UnusedTsExpectErrorStillReportsTS2578(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"a.ts": `// @ts-expect-error
const x: number = 1;
`,
		"tsconfig.json": `{"include":["a.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)

	var found bool
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:       true,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			if d.RuleName == "TypeScript(TS2578)" {
				found = true
			}
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if !found {
		t.Error("expected TS2578 (unused @ts-expect-error) to be reported")
	}
}
