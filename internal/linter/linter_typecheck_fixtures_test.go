package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Real-fixture tests covering scenarios that synthetic in-tempdir setups
// cannot easily exercise:
//   - node_modules-resolved d.ts under skipLibCheck on/off
//   - project references (composite + references)
//   - skipDefaultLibCheck / noLib interaction with user code that uses global types
//
// All assertions follow the same precision discipline as the matrix file:
// exact code, exact file suffix, exact 1-based (line, col), exact counts,
// and a baseline run wherever a test could otherwise pass trivially.

// programHasFileWithSuffix returns true if any source file in the program
// has a path ending with the given suffix. Use this for baseline checks
// instead of comparing against `filepath.Join(tempDir, ...)`: macOS
// `t.TempDir()` returns `/var/folders/...` while typescript-go realpaths
// to `/private/var/folders/...`, so direct equality fails.
func programHasFileWithSuffix(program *compiler.Program, suffix string) bool {
	for _, f := range program.GetSourceFiles() {
		if strings.HasSuffix(f.FileName(), suffix) {
			return true
		}
	}
	return false
}

// === A real-world node_modules d.ts === //
//
// Layout under tmpDir:
//   ./node_modules/foo/package.json   {"types": "index.d.ts"}
//   ./node_modules/foo/index.d.ts     declares an error (TS2307 missing module)
//   ./entry.ts                        imports 'foo'
//   ./tsconfig.json                   moduleResolution: "node"
//
// Under skipLibCheck:false the d.ts inside node_modules MUST produce its
// TS2307 (typescript-go's SkipTypeChecking does NOT special-case
// node_modules path; only IsDeclarationFile + skipLibCheck matters).
// Under skipLibCheck:true the same diagnostic must be suppressed.
func writeNodeModulesFooFixture(t *testing.T, dir string, skipLibCheck *bool) {
	t.Helper()
	files := map[string]string{
		"node_modules/foo/package.json": `{"name":"foo","types":"index.d.ts"}`,
		"node_modules/foo/index.d.ts": `import * as M from 'this-package-does-not-exist';
export type T = typeof M;
export declare const foo: T;
`,
		"entry.ts": `import {foo} from 'foo';
export const x = foo;
`,
	}
	var opts string
	switch {
	case skipLibCheck == nil:
		opts = ""
	case *skipLibCheck:
		opts = `,"skipLibCheck":true`
	default:
		opts = `,"skipLibCheck":false`
	}
	files["tsconfig.json"] = `{"compilerOptions":{"moduleResolution":"node16","module":"node16","target":"es2020"` + opts + `},"include":["entry.ts"]}`
	writeFiles(t, dir, files)
}

func TestFixture_NodeModulesDts_SkipLibCheckFalse_ReportsError(t *testing.T) {
	dir := t.TempDir()
	skip := false
	writeNodeModulesFooFixture(t, dir, &skip)
	program := createProgramFromTsconfigDir(t, dir)

	// Baseline: program loaded the d.ts under node_modules.
	if !programHasFileWithSuffix(program, "node_modules/foo/index.d.ts") {
		t.Fatal("baseline: node_modules/foo/index.d.ts not loaded; fixture broken")
	}

	diags := runProgramTypeCheck(t, program)
	// `node_modules/foo/index.d.ts:1:20` — opening quote of the missing
	// 'this-package-does-not-exist' module specifier.
	assertOneDiag(t, diags, "TS2307", "node_modules/foo/index.d.ts", 1, 20)
	assertExactDiagCount(t, diags, 1)
}

func TestFixture_NodeModulesDts_SkipLibCheckTrue_Suppresses(t *testing.T) {
	dir := t.TempDir()
	skip := true
	writeNodeModulesFooFixture(t, dir, &skip)
	program := createProgramFromTsconfigDir(t, dir)

	if !programHasFileWithSuffix(program, "node_modules/foo/index.d.ts") {
		t.Fatal("baseline: node_modules/foo/index.d.ts not loaded; fixture broken")
	}

	diags := runProgramTypeCheck(t, program)
	assertExactDiagCount(t, diags, 0)
}

// Negative pair-control: same fixture, default skipLibCheck (unspecified).
// Default-off means errors must be reported — proves the suppression in the
// previous test is caused by skipLibCheck:true and not by some other
// implicit filter (e.g. ExcludePaths suppressing node_modules).
func TestFixture_NodeModulesDts_SkipLibCheckUnspecified_ReportsError(t *testing.T) {
	dir := t.TempDir()
	writeNodeModulesFooFixture(t, dir, nil)
	program := createProgramFromTsconfigDir(t, dir)

	if !programHasFileWithSuffix(program, "node_modules/foo/index.d.ts") {
		t.Fatal("baseline: node_modules/foo/index.d.ts not loaded; fixture broken")
	}

	diags := runProgramTypeCheck(t, program)
	assertOneDiag(t, diags, "TS2307", "node_modules/foo/index.d.ts", 1, 20)
	assertExactDiagCount(t, diags, 1)
}

// === Project references === //
//
// Layout:
//   /a/tsconfig.json   composite, includes a.ts, references ../b
//   /a/a.ts            imports from '../b/b'
//   /b/tsconfig.json   composite, includes b.ts, declaration outputs in dist/
//   /b/b.ts            real type error (TS2322 on `broken`)
//   /b/dist/b.d.ts     pre-built declaration so A can resolve to B's outputs
//   /b/dist/b.d.ts.map sourcemap so source positions line up
//
// When both A and B programs are passed to RunLinter:
//   - typescript-go marks B's b.ts as IsSourceFromProjectReference inside
//     program A's view, so A skips reporting it.
//   - Program B owns b.ts and reports TS2322.
//   - Cross-program dedup is the secondary safety net.
// Assert the TS2322 appears exactly once. There may be additional
// scaffolding diagnostics (e.g. TS6305 if outputs go stale); we pin the
// b.ts:TS2322 anchor and the count of TS2322 specifically.
func TestFixture_ProjectReferences_FileReportedExactlyOnce(t *testing.T) {
	root := t.TempDir()

	// Project B (referenced) — pre-built so A can resolve `../b/b` cleanly.
	bDir := filepath.Join(root, "b")
	if err := os.MkdirAll(filepath.Join(bDir, "dist"), 0755); err != nil {
		t.Fatalf("mkdir b/dist: %v", err)
	}
	writeFiles(t, bDir, map[string]string{
		"tsconfig.json": `{
  "compilerOptions": {
    "composite": true,
    "outDir": "dist",
    "rootDir": ".",
    "declaration": true
  },
  "include": ["b.ts"]
}`,
		"b.ts": "export const broken: number = 'oops';\n",
		// Pre-built declaration so A resolves '../b/b' against B's outputs.
		"dist/b.d.ts":          "export declare const broken: number;\n",
		"dist/b.d.ts.map":      `{"version":3,"file":"b.d.ts","sourceRoot":"","sources":["../b.ts"],"names":[],"mappings":""}`,
		"dist/tsconfig.tsbuildinfo": `{"version":"5.6.0"}`,
	})

	// Project A (referencing B)
	aDir := filepath.Join(root, "a")
	if err := os.MkdirAll(aDir, 0755); err != nil {
		t.Fatalf("mkdir a: %v", err)
	}
	writeFiles(t, aDir, map[string]string{
		"tsconfig.json": `{
  "compilerOptions": {
    "composite": true,
    "outDir": "dist",
    "rootDir": "."
  },
  "include": ["a.ts"],
  "references": [{"path": "../b"}]
}`,
		"a.ts": "import {broken} from '../b/b';\nexport const v: number = broken;\n",
	})

	progA := createProgramFromTsconfigDir(t, aDir)
	progB := createProgramFromTsconfigDir(t, bDir)

	if !programHasFileWithSuffix(progA, "/a/a.ts") {
		t.Fatal("baseline: a.ts not in program A")
	}
	if !programHasFileWithSuffix(progB, "/b/b.ts") {
		t.Fatal("baseline: b.ts not in program B")
	}

	var diags []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{progA, progB},
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

	// `b.ts:1:14` — variable identifier `broken` (col 14, 1-based) gets
	// the TS2322. There must be exactly one such diagnostic — proving:
	//   (1) program B reports it,
	//   (2) program A does NOT double-report it (IsSourceFromProjectReference),
	//   (3) cross-program dedup keeps a single entry.
	assertOneDiag(t, diags, "TS2322", "b/b.ts", 1, 14)
	assertExactCodeCount(t, diags, "TS2322", 1)
}

// === Default lib check === //
//
// Verifies that under default skipDefaultLibCheck:false the program loads
// its default library AND user source code that uses global types compiles
// cleanly (i.e. the default lib is actually being type-checked but produces
// zero errors itself). This is the de-facto baseline that any real project
// relies on.
func TestFixture_DefaultLib_LoadedAndCleanByDefault(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"a.ts":          `export const arr = Array.from('abc');` + "\n",
		"tsconfig.json": `{"compilerOptions":{"target":"es2020"},"include":["a.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)
	diags := runProgramTypeCheck(t, program)
	assertExactDiagCount(t, diags, 0)
}

// Toggling skipDefaultLibCheck:true must not cause user code that depends
// on global types (Array.from) to start failing — the lib is still loaded,
// just not type-checked. So user diagnostics remain at zero for valid code.
func TestFixture_SkipDefaultLibCheckTrue_UserCodeStillTypeChecked(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"a.ts":          `export const arr = Array.from('abc');` + "\n",
		"tsconfig.json": `{"compilerOptions":{"target":"es2020","skipDefaultLibCheck":true},"include":["a.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)
	diags := runProgramTypeCheck(t, program)
	assertExactDiagCount(t, diags, 0)
}

// And introducing a real user error under skipDefaultLibCheck:true MUST
// surface — confirming the previous test isn't passing because the type-
// check phase silently bailed out.
func TestFixture_SkipDefaultLibCheckTrue_UserErrorStillReported(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"a.ts":          "const x: number = 'oops';\n",
		"tsconfig.json": `{"compilerOptions":{"target":"es2020","skipDefaultLibCheck":true},"include":["a.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)
	diags := runProgramTypeCheck(t, program)
	assertOneDiag(t, diags, "TS2322", "a.ts", 1, 7)
	assertExactDiagCount(t, diags, 1)
}

// `lib`-narrowed configurations: with only `lib:["es2015"]`, `BigInt` is
// unknown (introduced in es2020). The user code should produce a
// diagnostic at the `BigInt` identifier — proof that the lib selection
// is actually being honoured by the type-check phase.
func TestFixture_LibNarrowed_MissingGlobalReported(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"a.ts": "export const b = BigInt(1);\n",
		// es2015 lib excludes BigInt (introduced in es2020).
		"tsconfig.json": `{"compilerOptions":{"target":"es2015","lib":["es2015"]},"include":["a.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)
	diags := runProgramTypeCheck(t, program)

	// `export const b = BigInt(1);`
	//                   ^ col 18 (1-based)
	// TS2583 = "Cannot find name '{0}'. Do you need to change your target
	// library? Try changing the `lib` compiler option to include 'esXX'."
	assertOneDiag(t, diags, "TS2583", "a.ts", 1, 18)
	assertExactDiagCount(t, diags, 1)
}

// Pair-control to the previous test: same source under `target:es2020`
// (which includes BigInt) compiles cleanly. This confirms the `lib`
// setting is the cause of the previous test's failure rather than some
// unrelated rejection.
func TestFixture_LibIncludesBigInt_NoErrors(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"a.ts":          "export const b = BigInt(1);\n",
		"tsconfig.json": `{"compilerOptions":{"target":"es2020","lib":["es2020"]},"include":["a.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)
	diags := runProgramTypeCheck(t, program)
	assertExactDiagCount(t, diags, 0)
}
