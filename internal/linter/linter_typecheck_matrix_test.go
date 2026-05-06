package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Cross-matrix tests for --type-check semantics.
//
// Axes covered:
//   A. skipLibCheck setting:        unspecified | true | false
//   B. file location:               own .ts source | project .d.ts | node_modules .d.ts
//   C. program count:               single | multi-program with mixed settings
//   D. compiler options:            noCheck | noEmit | strict differences
//   E. file-level directives:       // @ts-nocheck
//   F. skipDefaultLibCheck
//
// Every test asserts:
//   - exact diagnostic counts (not "at least one"),
//   - exact TypeScript error codes (not just "some TS error"),
//   - exact file paths,
//   - exact 1-based (line, column) for at least one anchor diagnostic.
// Most tests also include a baseline run (negative or positive control)
// to prove the program is actually exercising the type-check phase.

// runProgramTypeCheck runs RunLinter on a single program and returns the
// TypeScript(TSxxxx) diagnostics emitted.
func runProgramTypeCheck(t *testing.T, program *compiler.Program) []rule.RuleDiagnostic {
	t.Helper()
	var out []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:       true,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			if strings.HasPrefix(d.RuleName, "TypeScript(") {
				out = append(out, d)
			}
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	return out
}

// diagFingerprint summarises a diagnostic for stable error messages.
type diagFingerprint struct {
	code   string
	suffix string // file path suffix
	line   int    // 1-based
	col    int    // 1-based
}

func fingerprint(d rule.RuleDiagnostic) diagFingerprint {
	fp := diagFingerprint{code: d.RuleName}
	if d.SourceFile != nil {
		// 0-based from typescript-go → +1 to match user-visible line/col
		line, char := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, d.Range.Pos())
		fp.suffix = d.SourceFile.FileName()
		fp.line = line + 1
		fp.col = int(char) + 1
	}
	return fp
}

// assertOneDiag fails the test unless diags has exactly one diagnostic
// matching code/suffix/line/col.
func assertOneDiag(t *testing.T, diags []rule.RuleDiagnostic, code, fileSuffix string, line, col int) {
	t.Helper()
	matches := 0
	for _, d := range diags {
		fp := fingerprint(d)
		if fp.code == "TypeScript("+code+")" &&
			strings.HasSuffix(fp.suffix, fileSuffix) &&
			fp.line == line && fp.col == col {
			matches++
		}
	}
	if matches != 1 {
		t.Errorf("expected exactly 1 %s at %s:%d:%d, found %d. all diags: %s",
			code, fileSuffix, line, col, matches, dumpDiags(diags))
	}
}

// assertNoDiagInFile fails if any diagnostic's file ends with the given suffix.
func assertNoDiagInFile(t *testing.T, diags []rule.RuleDiagnostic, fileSuffix string) {
	t.Helper()
	for _, d := range diags {
		if d.SourceFile != nil && strings.HasSuffix(d.SourceFile.FileName(), fileSuffix) {
			t.Errorf("expected no diagnostic in %s, got %s at line %d. all diags: %s",
				fileSuffix, d.RuleName, fingerprint(d).line, dumpDiags(diags))
		}
	}
}

// assertExactDiagCount fails unless the slice has exactly n diagnostics.
func assertExactDiagCount(t *testing.T, diags []rule.RuleDiagnostic, n int) {
	t.Helper()
	if len(diags) != n {
		t.Errorf("expected exactly %d type diagnostics, got %d. all diags: %s",
			n, len(diags), dumpDiags(diags))
	}
}

// assertExactCodeCount fails unless exactly n diagnostics with the given code exist.
func assertExactCodeCount(t *testing.T, diags []rule.RuleDiagnostic, code string, n int) {
	t.Helper()
	got := 0
	for _, d := range diags {
		if d.RuleName == "TypeScript("+code+")" {
			got++
		}
	}
	if got != n {
		t.Errorf("expected exactly %d %s, got %d. all diags: %s",
			n, code, got, dumpDiags(diags))
	}
}

func dumpDiags(diags []rule.RuleDiagnostic) string {
	if len(diags) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, d := range diags {
		if i > 0 {
			b.WriteString(", ")
		}
		fp := fingerprint(d)
		b.WriteString(fp.code)
		b.WriteString("@")
		short := fp.suffix
		if i := strings.LastIndex(short, "/"); i >= 0 {
			short = short[i+1:]
		}
		b.WriteString(short)
		b.WriteString(":")
		b.WriteString(itoa(fp.line))
		b.WriteString(":")
		b.WriteString(itoa(fp.col))
	}
	b.WriteString("]")
	return b.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var s [12]byte
	i := len(s)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		s[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		s[i] = '-'
	}
	return string(s[i:])
}

// === Axis A × Axis B: skipLibCheck × file location ===

// (A=unspecified, B=project .d.ts) — typescript-go default treats absent
// skipLibCheck as off. The d.ts must produce TS2307 for the missing module
// import. The .ts user-side import should NOT produce its own redundant
// error.
func TestMatrix_SkipLibCheck_Unspecified_ProjectDtsErrorReported(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"local.d.ts": `import * as M from 'this-package-does-not-exist';
export type T = typeof M;
`,
		"entry.ts":      "import type {T} from './local';\nexport const x: T = {} as T;\n",
		"tsconfig.json": `{"include":["entry.ts","local.d.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)
	diags := runProgramTypeCheck(t, program)

	// Exactly one TS2307 in local.d.ts at column 20 — the opening quote of
	// the 'this-package-does-not-exist' module specifier in
	//   `import * as M from 'this-package-does-not-exist';`
	//                       ^ col 20
	assertOneDiag(t, diags, "TS2307", "local.d.ts", 1, 20)
	// And nothing else.
	assertExactDiagCount(t, diags, 1)
}

// (A=true, B=project-internal .d.ts NOT in node_modules) — skipLibCheck
// applies to ALL declaration files irrespective of location.
func TestMatrix_SkipLibCheck_True_SuppressesProjectDtsErrors(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"local.d.ts": `import * as M from 'this-package-does-not-exist';
export type T = typeof M;
`,
		"entry.ts":      "import type {T} from './local';\nexport const x: T = {} as T;\n",
		"tsconfig.json": `{"compilerOptions":{"skipLibCheck":true},"include":["entry.ts","local.d.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)
	diags := runProgramTypeCheck(t, program)

	// Baseline confirmation: program loaded both files (otherwise the test
	// would pass trivially because the program is empty).
	if program.GetSourceFile(filepath.Join(dir, "local.d.ts")) == nil {
		t.Fatal("baseline: local.d.ts not loaded into program — test setup broken")
	}
	if program.GetSourceFile(filepath.Join(dir, "entry.ts")) == nil {
		t.Fatal("baseline: entry.ts not loaded into program — test setup broken")
	}

	// No diagnostics anywhere — d.ts errors suppressed, entry.ts has no
	// real type errors of its own.
	assertExactDiagCount(t, diags, 0)
}

// (A=true, B=own .ts source) — skipLibCheck only affects declaration files.
// A real TS2322 in a .ts file MUST still be reported.
func TestMatrix_SkipLibCheck_True_DoesNotSuppressSourceErrors(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"a.ts":          "const x: number = 'oops';\n",
		"tsconfig.json": `{"compilerOptions":{"skipLibCheck":true},"include":["a.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)
	diags := runProgramTypeCheck(t, program)

	// `const x: number = 'oops';` → TS2322 at line 1, col 7 (variable name).
	assertOneDiag(t, diags, "TS2322", "a.ts", 1, 7)
	assertExactDiagCount(t, diags, 1)
}

// === Axis C: multi-program with mixed skipLibCheck on a shared d.ts ===

// Program A skipLibCheck:false → reports d.ts error.
// Program B skipLibCheck:true → suppresses d.ts error.
// Across both, dedup keeps a SINGLE TS2307. Asserts:
//   - exact code TS2307
//   - exact file (shared.d.ts)
//   - exact (line, col) (1, 21)
//   - exactly 1 occurrence — proving dedup works AND that B's silence
//     didn't suppress A's report.
func TestMatrix_MultiProgram_MixedSkipLibCheck_SharedDtsReportedExactlyOnce(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"shared.d.ts": `import * as M from 'this-package-does-not-exist';
export type T = typeof M;
`,
	})
	mk := func(subdir string, skipLib bool) *compiler.Program {
		p := filepath.Join(root, subdir)
		_ = os.MkdirAll(p, 0755)
		opts := `"skipLibCheck":false`
		if skipLib {
			opts = `"skipLibCheck":true`
		}
		_ = os.WriteFile(filepath.Join(p, "tsconfig.json"),
			[]byte(`{"compilerOptions":{`+opts+`},"include":["../shared.d.ts"]}`), 0644)
		return createProgramFromTsconfigDir(t, p)
	}
	progA := mk("a", false)
	progB := mk("b", true)

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

	assertOneDiag(t, diags, "TS2307", "shared.d.ts", 1, 20)
	assertExactDiagCount(t, diags, 1)
}

// Inverse control: BOTH programs skipLibCheck:true → ZERO TS2307 reported.
// Together with the previous test this proves:
//   1. the d.ts genuinely has the error (program A in the previous test reported it),
//   2. when ALL programs containing it have skipLibCheck:true, none reports.
func TestMatrix_MultiProgram_BothSkipLibCheck_NoDtsErrors(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"shared.d.ts": `import * as M from 'this-package-does-not-exist';
export type T = typeof M;
`,
	})
	mk := func(subdir string) *compiler.Program {
		p := filepath.Join(root, subdir)
		_ = os.MkdirAll(p, 0755)
		_ = os.WriteFile(filepath.Join(p, "tsconfig.json"),
			[]byte(`{"compilerOptions":{"skipLibCheck":true},"include":["../shared.d.ts"]}`), 0644)
		return createProgramFromTsconfigDir(t, p)
	}
	progA := mk("a")
	progB := mk("b")

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

	assertExactDiagCount(t, diags, 0)
}

// === Axis D: noCheck === //

// Baseline + noCheck: assert the SAME source produces TS2322 normally and
// zero diagnostics under noCheck:true.
func TestMatrix_NoCheck_SuppressesEverything(t *testing.T) {
	src := "const x: number = 'oops';\n"
	build := func(opts string) []rule.RuleDiagnostic {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"a.ts":          src,
			"tsconfig.json": `{"compilerOptions":{` + opts + `},"include":["a.ts"]}`,
		})
		return runProgramTypeCheck(t, createProgramFromTsconfigDir(t, dir))
	}

	// Baseline: no noCheck → TS2322 at line 1, col 7, exactly one diag.
	baseline := build(`"strict":true`)
	assertOneDiag(t, baseline, "TS2322", "a.ts", 1, 7)
	assertExactDiagCount(t, baseline, 1)

	// noCheck:true → zero.
	suppressed := build(`"strict":true,"noCheck":true`)
	assertExactDiagCount(t, suppressed, 0)
}

// === Axis D: noEmit ===

// noEmit:true does NOT suppress real type errors. Asserts the SAME error
// surfaces both with and without noEmit.
func TestMatrix_NoEmit_DoesNotSuppressTypeErrors(t *testing.T) {
	src := "const x: number = 'oops';\n"
	build := func(opts string) []rule.RuleDiagnostic {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"a.ts":          src,
			"tsconfig.json": `{"compilerOptions":{` + opts + `},"include":["a.ts"]}`,
		})
		return runProgramTypeCheck(t, createProgramFromTsconfigDir(t, dir))
	}
	baseline := build(``)
	withNoEmit := build(`"noEmit":true`)
	assertOneDiag(t, baseline, "TS2322", "a.ts", 1, 7)
	assertExactDiagCount(t, baseline, 1)
	assertOneDiag(t, withNoEmit, "TS2322", "a.ts", 1, 7)
	assertExactDiagCount(t, withNoEmit, 1)
}

// === Axis E: file-level @ts-nocheck === //

// Baseline + @ts-nocheck: same source produces 2 TS errors normally; with
// @ts-nocheck the file is fully suppressed.
func TestMatrix_TsNocheck_SuppressesEntireFile(t *testing.T) {
	build := func(src string) []rule.RuleDiagnostic {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"a.ts":          src,
			"tsconfig.json": `{"include":["a.ts"]}`,
		})
		return runProgramTypeCheck(t, createProgramFromTsconfigDir(t, dir))
	}

	// Baseline: no directive → 2 TS2322 (both lines have type errors).
	baseline := build("const x: number = 'oops';\nconst y: string = 42;\n")
	assertExactCodeCount(t, baseline, "TS2322", 2)

	// With @ts-nocheck → file-wide suppression.
	suppressed := build("// @ts-nocheck\nconst x: number = 'oops';\nconst y: string = 42;\n")
	assertNoDiagInFile(t, suppressed, "a.ts")
	assertExactDiagCount(t, suppressed, 0)
}

// === Axis F: skipDefaultLibCheck === //

// skipDefaultLibCheck:true alone (skipLibCheck not set) only skips
// lib.*.d.ts. A user-authored .d.ts with a real error must still produce
// TS2307. (We use skipLibCheck:false explicitly for clarity, but absence of
// skipLibCheck would be equivalent.)
func TestMatrix_SkipDefaultLibCheck_OnlySkipsDefaultLib(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"local.d.ts": `import * as M from 'this-package-does-not-exist';
export type T = typeof M;
`,
		"entry.ts": "import type {T} from './local';\nexport const x: T = {} as T;\n",
		"tsconfig.json": `{
"compilerOptions":{"skipDefaultLibCheck":true,"skipLibCheck":false},
"include":["entry.ts","local.d.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)
	diags := runProgramTypeCheck(t, program)

	assertOneDiag(t, diags, "TS2307", "local.d.ts", 1, 20)
	assertExactDiagCount(t, diags, 1)
}

// === TypeInfoFiles gate as backstop ===
//
// Baseline + with-gate: same source, same program. Without the gate, TS2322
// appears. With TypeInfoFiles excluding the file, the diagnostic is dropped.
func TestMatrix_TypeInfoFiles_GateActsAsBackstopForGapFiles(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"a.ts":          "const x: number = 'oops';\n",
		"tsconfig.json": `{"include":["a.ts"]}`,
	})
	program := createProgramFromTsconfigDir(t, dir)

	collect := func(infoFiles map[string]struct{}) []rule.RuleDiagnostic {
		var out []rule.RuleDiagnostic
		_, err := RunLinter(RunLinterOptions{
			Programs:        []*compiler.Program{program},
			SingleThreaded:  true,
			GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
			TypeCheck:       true,
			TypeInfoFiles:   infoFiles,
			OnDiagnostic: func(d rule.RuleDiagnostic) {
				if strings.HasPrefix(d.RuleName, "TypeScript(") {
					out = append(out, d)
				}
			},
		})
		if err != nil {
			t.Fatalf("RunLinter: %v", err)
		}
		return out
	}

	// Baseline: no gate → 1 TS2322.
	baseline := collect(nil)
	assertOneDiag(t, baseline, "TS2322", "a.ts", 1, 7)
	assertExactDiagCount(t, baseline, 1)

	// Gate that excludes a.ts → 0.
	gated := collect(map[string]struct{}{"/some/other/file.ts": {}})
	assertExactDiagCount(t, gated, 0)

	// Gate that includes a.ts → still 1 (the gate is a positive set).
	included := collect(map[string]struct{}{program.GetSourceFile(filepath.Join(dir, "a.ts")).FileName(): {}})
	assertOneDiag(t, included, "TS2322", "a.ts", 1, 7)
	assertExactDiagCount(t, included, 1)
}

// === Dedup with strict-only diagnostics across programs ===
//
// Same source loaded by two programs with different strictness:
//   - strict program reports TS7006 (param `x` implicitly any) at col 21
//   - loose program reports nothing
// Across both, exactly 1 TS7006 should reach the consumer. Asserts the
// dedup keeps strict's report and that the loose program's silence does
// not override it.
func TestMatrix_StrictOnlyDiagAcrossPrograms_KeptExactlyOnce(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"shared.ts": "export function fn(x) { return x; }\n",
	})
	mk := func(subdir, opts string) *compiler.Program {
		p := filepath.Join(root, subdir)
		_ = os.MkdirAll(p, 0755)
		_ = os.WriteFile(filepath.Join(p, "tsconfig.json"),
			[]byte(`{"compilerOptions":{`+opts+`},"include":["../shared.ts"]}`), 0644)
		return createProgramFromTsconfigDir(t, p)
	}
	strict := mk("strict", `"noImplicitAny":true`)
	loose := mk("loose", `"noImplicitAny":false`)

	var got []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{strict, loose},
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:       true,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			if strings.HasPrefix(d.RuleName, "TypeScript(") {
				got = append(got, d)
			}
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}

	// `export function fn(x) {...}` — parameter `x` at column 20 (1-based).
	assertOneDiag(t, got, "TS7006", "shared.ts", 1, 20)
	assertExactDiagCount(t, got, 1)
}

// === TypeCheck=false fully disables Phase 2 === //
//
// Same source as TestMatrix_SkipLibCheck_True_DoesNotSuppressSourceErrors:
// has 1 TS2322 when TypeCheck=true; 0 when TypeCheck=false.
func TestMatrix_TypeCheckFalse_NoPhaseTwo(t *testing.T) {
	src := "const x: number = 'oops';\n"
	build := func() *compiler.Program {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"a.ts":          src,
			"tsconfig.json": `{"include":["a.ts"]}`,
		})
		return createProgramFromTsconfigDir(t, dir)
	}

	// Baseline: TypeCheck=true → 1 TS2322.
	on := runProgramTypeCheck(t, build())
	assertOneDiag(t, on, "TS2322", "a.ts", 1, 7)
	assertExactDiagCount(t, on, 1)

	// TypeCheck=false → 0.
	var got []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{build()},
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:       false,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			if strings.HasPrefix(d.RuleName, "TypeScript(") {
				got = append(got, d)
			}
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	assertExactDiagCount(t, got, 0)
}

// === noImplicitAny strict toggle === //
//
// Verifies that strict checks are honored at the program level (the source
// file alone is identical; only compilerOptions differ).
func TestMatrix_NoImplicitAny_HonoredByProgram(t *testing.T) {
	src := "export function fn(x) { return x; }\n"
	build := func(opts string) []rule.RuleDiagnostic {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"a.ts":          src,
			"tsconfig.json": `{"compilerOptions":{` + opts + `},"include":["a.ts"]}`,
		})
		return runProgramTypeCheck(t, createProgramFromTsconfigDir(t, dir))
	}

	loose := build(`"noImplicitAny":false`)
	assertExactDiagCount(t, loose, 0)

	strict := build(`"noImplicitAny":true`)
	// `export function fn(x) {...}` — parameter `x` at col 20.
	assertOneDiag(t, strict, "TS7006", "a.ts", 1, 20)
	assertExactDiagCount(t, strict, 1)
}
