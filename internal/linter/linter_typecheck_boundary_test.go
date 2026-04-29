package linter

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Boundary tests for --type-check semantics. These pin invariants that
// would otherwise drift silently between typescript-go versions or under
// internal API refactors.

// === Boundary 1: default-lib invariant === //
//
// typescript-go ships its lib.*.d.ts files (lib.es5, lib.dom, etc.)
// under bundled:///libs/... and treats them as default-library files.
// They MUST type-check cleanly under every supported lib selection — if
// any of them ever had an error, this would explode user output.
//
// We assert: across every reasonable lib/target combination, the
// type-check phase produces ZERO diagnostics whose file is a
// bundled:///libs/* path.
func TestBoundary_DefaultLibsTypeCheckClean(t *testing.T) {
	matrix := []struct {
		name    string
		options string
	}{
		{"default", `{}`},
		{"target_es5_lib_es5", `{"target":"es2015","lib":["es5"]}`},
		{"target_es2020", `{"target":"es2020"}`},
		{"target_esnext_dom", `{"target":"esnext","lib":["esnext","dom"]}`},
		{"strict_target_es2020", `{"target":"es2020","strict":true}`},
		{"skipDefaultLibCheck_off", `{"target":"es2020","skipDefaultLibCheck":false}`},
	}
	for _, tc := range matrix {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, map[string]string{
				"a.ts":          "export const x = 1;\n",
				"tsconfig.json": `{"compilerOptions":` + tc.options + `,"include":["a.ts"]}`,
			})
			program := createProgramFromTsconfigDir(t, dir)
			diags := runProgramTypeCheck(t, program)

			for _, d := range diags {
				if d.SourceFile == nil {
					continue
				}
				name := d.SourceFile.FileName()
				if strings.HasPrefix(name, "bundled:") || strings.Contains(name, "/libs/lib.") {
					t.Errorf("default-lib invariant violated under %s: diagnostic %s in %s",
						tc.name, d.RuleName, name)
				}
			}
		})
	}
}

// Same invariant under multi-program load. If pool-level checker
// invocations were ever to surface lib diagnostics in just one program's
// view but not another, this would catch it.
func TestBoundary_DefaultLibsCleanAcrossPrograms(t *testing.T) {
	mk := func(name string, opts string) *compiler.Program {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"a.ts":          "export const " + name + " = 1;\n",
			"tsconfig.json": `{"compilerOptions":` + opts + `,"include":["a.ts"]}`,
		})
		return createProgramFromTsconfigDir(t, dir)
	}
	progs := []*compiler.Program{
		mk("a", `{"target":"es2020","strict":true}`),
		mk("b", `{"target":"es2015","lib":["es2015"]}`),
		mk("c", `{"target":"esnext","lib":["esnext","dom"]}`),
	}

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        progs,
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:       true,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			if strings.HasPrefix(d.RuleName, "TypeScript(") {
				diags = append(diags, d)
			}
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}

	for _, d := range diags {
		if d.SourceFile == nil {
			continue
		}
		name := d.SourceFile.FileName()
		if strings.HasPrefix(name, "bundled:") || strings.Contains(name, "/libs/lib.") {
			t.Errorf("default-lib invariant violated across programs: %s in %s",
				d.RuleName, name)
		}
	}
}

// === Boundary 2: program-level vs per-file equivalence === //
//
// rslint's runTypeCheckForProgram calls GetDiagnosticsOfAnyProgram with
// file=nil, which goes through compilerCheckerPool.forEachCheckerGroupDo
// internally. The legacy file-scoped path called GetSemanticDiagnostics
// per file in a Goroutine pool that calls SkipTypeChecking BEFORE the
// per-file collect function.
//
// These two paths can in principle diverge if SkipTypeChecking is gated
// differently. This test pins the invariant: the per-file *semantic*
// diagnostic set (gathered by walking each source file individually) is
// a SUBSET of what we extract from GetDiagnosticsOfAnyProgram. Failure
// means the program-level path silently drops a diagnostic that per-file
// would have surfaced.
func TestBoundary_ProgramLevelMatchesPerFileSemanticDiagnostics(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		// Mix: source-level errors + d.ts errors + unused @ts-expect-error
		// to exercise multiple diagnostic paths.
		"entry.ts": `// @ts-expect-error
const x: number = 1;
const bad: string = 42;
`,
		"local.d.ts": `import * as M from 'this-package-does-not-exist';
export type T = typeof M;
`,
		"tsconfig.json": `{"compilerOptions":{"skipLibCheck":false},"include":["entry.ts","local.d.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)
	ctx := context.Background()

	progDiags := compiler.GetDiagnosticsOfAnyProgram(
		ctx, program, nil, false,
		program.GetBindDiagnostics,
		program.GetSemanticDiagnostics,
	)

	var perFileDiags []*ast.Diagnostic
	for _, sf := range program.GetSourceFiles() {
		perFileDiags = append(perFileDiags, program.GetSemanticDiagnostics(ctx, sf)...)
	}

	type key struct {
		file string
		code int32
		pos  int
		end  int
	}
	semanticOnly := func(diags []*ast.Diagnostic) map[key]struct{} {
		out := make(map[key]struct{})
		for _, d := range diags {
			if d.File() == nil {
				continue
			}
			out[key{d.File().FileName(), d.Code(), d.Loc().Pos(), d.Loc().End()}] = struct{}{}
		}
		return out
	}

	progSet := semanticOnly(progDiags)
	perFileSet := semanticOnly(perFileDiags)

	// Sanity: there must be at least one entry in each set, otherwise
	// the test passes trivially.
	if len(perFileSet) == 0 {
		t.Fatal("baseline: per-file semantic walk produced 0 diagnostics; fixture broken")
	}
	if len(progSet) == 0 {
		t.Fatal("baseline: program-level walk produced 0 diagnostics; fixture broken")
	}

	for k := range perFileSet {
		if _, ok := progSet[k]; !ok {
			t.Errorf("program-level path missing (file=%s code=TS%d pos=%d end=%d) that per-file produced",
				k.file, k.code, k.pos, k.end)
		}
	}
}

// === Boundary 3: stability — same program, same diagnostics === //
//
// typescript-go's checker pool is concurrent. If anything in our type-
// check phase introduced nondeterminism (race-induced reordering with a
// downstream effect, missing locks, etc.) repeated runs over the same
// program could yield slightly different diagnostic sets. This test runs
// the same fixture 5 times and asserts the resulting fingerprint
// (code, suffix, line, col) sets are byte-for-byte identical.
func TestBoundary_DiagnosticsStableAcrossRepeatedRuns(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"entry.ts":      "const x: number = 'oops';\nconst y: string = 42;\n",
		"local.d.ts":    "import * as M from 'this-package-does-not-exist';\nexport type T = typeof M;\n",
		"tsconfig.json": `{"compilerOptions":{"skipLibCheck":false},"include":["entry.ts","local.d.ts"]}`,
	})

	type fp struct {
		code   string
		suffix string
		line   int
		col    int
	}
	captureRun := func(t *testing.T) []fp {
		program := createProgramFromTsconfigDir(t, dir)
		diags := runProgramTypeCheck(t, program)
		out := make([]fp, 0, len(diags))
		for _, d := range diags {
			if d.SourceFile == nil {
				continue
			}
			line, char := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, d.Range.Pos())
			suffix := d.SourceFile.FileName()
			if i := strings.LastIndex(suffix, "/"); i >= 0 {
				suffix = suffix[i+1:]
			}
			out = append(out, fp{d.RuleName, suffix, line + 1, int(char) + 1})
		}
		sort.Slice(out, func(i, j int) bool {
			if out[i].code != out[j].code {
				return out[i].code < out[j].code
			}
			if out[i].suffix != out[j].suffix {
				return out[i].suffix < out[j].suffix
			}
			if out[i].line != out[j].line {
				return out[i].line < out[j].line
			}
			return out[i].col < out[j].col
		})
		return out
	}

	first := captureRun(t)
	if len(first) == 0 {
		t.Fatal("baseline: expected at least one diagnostic; fixture broken")
	}

	for i := range 4 {
		next := captureRun(t)
		if !reflect.DeepEqual(first, next) {
			t.Errorf("run %d produced different diagnostics:\n  first: %v\n  this:  %v",
				i+1, first, next)
		}
	}
}

// Same stability check exercising parallel program processing
// (SingleThreaded=false) to catch any concurrency issue in the type-check
// phase.
func TestBoundary_DiagnosticsStableUnderParallelism(t *testing.T) {
	mk := func(name string) *compiler.Program {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"entry.ts": "const x: number = 'oops';\nconst y: string = 42;\n",
			// Distinct package name per program so the d.ts errors don't
			// dedup across them.
			"local.d.ts":    "import * as M from 'missing-pkg-" + name + "';\nexport type T = typeof M;\n",
			"tsconfig.json": `{"compilerOptions":{"skipLibCheck":false},"include":["entry.ts","local.d.ts"]}`,
		})
		return createProgramFromTsconfigDir(t, dir)
	}

	type fp struct {
		code   string
		suffix string
		line   int
		col    int
	}
	captureRun := func(t *testing.T) []fp {
		programs := []*compiler.Program{mk("a"), mk("b"), mk("c")}
		var mu sync.Mutex
		var diags []rule.RuleDiagnostic
		_, err := RunLinter(RunLinterOptions{
			Programs:        programs,
			SingleThreaded:  false,
			GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
			TypeCheck:       true,
			OnDiagnostic: func(d rule.RuleDiagnostic) {
				if !strings.HasPrefix(d.RuleName, "TypeScript(") {
					return
				}
				mu.Lock()
				diags = append(diags, d)
				mu.Unlock()
			},
		})
		if err != nil {
			t.Fatalf("RunLinter: %v", err)
		}
		out := make([]fp, 0, len(diags))
		for _, d := range diags {
			if d.SourceFile == nil {
				continue
			}
			line, char := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, d.Range.Pos())
			parts := strings.Split(d.SourceFile.FileName(), "/")
			tail := parts[len(parts)-1]
			out = append(out, fp{d.RuleName, tail, line + 1, int(char) + 1})
		}
		sort.Slice(out, func(i, j int) bool {
			if out[i].code != out[j].code {
				return out[i].code < out[j].code
			}
			if out[i].suffix != out[j].suffix {
				return out[i].suffix < out[j].suffix
			}
			if out[i].line != out[j].line {
				return out[i].line < out[j].line
			}
			return out[i].col < out[j].col
		})
		return out
	}

	first := captureRun(t)
	if len(first) == 0 {
		t.Fatal("baseline: expected diagnostics; fixture broken")
	}

	for i := range 4 {
		next := captureRun(t)
		if !reflect.DeepEqual(first, next) {
			t.Errorf("parallel run %d produced different diagnostics:\n  first: %v\n  this:  %v",
				i+1, first, next)
		}
	}
}

// === Boundary 4: file=nil diagnostics are intentionally dropped === //
//
// rslint's design decision: skip diagnostics with no file location
// (config-parsing, option errors). Pin that decision: produce a fixture
// that yields a file=nil diagnostic via tsconfig and assert it is NOT
// surfaced through the rslint pipeline.
func TestBoundary_NoSourceLocationDiagnosticsDropped(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		// `include:["src"]` matches no files → typescript-go emits TS18003
		// ("No inputs were found...") with file=nil.
		"tsconfig.json": `{"include":["src"]}`,
	})

	program := createProgramFromTsconfigDir(t, dir)
	raw := compiler.GetDiagnosticsOfAnyProgram(
		context.Background(), program, nil, false,
		program.GetBindDiagnostics,
		program.GetSemanticDiagnostics,
	)
	hasFilelessTS18003 := false
	for _, d := range raw {
		if d.Code() == 18003 && d.File() == nil {
			hasFilelessTS18003 = true
			break
		}
	}
	if !hasFilelessTS18003 {
		t.Skip("fixture did not yield the expected file=nil TS18003 (typescript-go behavior changed); skipping invariant")
	}

	var got []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:       true,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			got = append(got, d)
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	for _, d := range got {
		if d.RuleName == "TypeScript(TS18003)" {
			t.Errorf("expected fileless TS18003 to be dropped, got: %+v", d)
		}
	}
}
