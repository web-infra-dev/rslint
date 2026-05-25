package linter

import (
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// --- Type-check basic functionality ---

func TestTypeCheck_ReportsSemanticErrors(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true, // typeCheck enabled
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	if len(diagnostics) == 0 {
		t.Fatal("Expected type-check diagnostics, got none")
	}

	found := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected at least one diagnostic with TS prefix")
	}
}

func TestTypeCheck_Disabled_NoSemanticErrors(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		false, // typeCheck disabled
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			t.Errorf("Expected no TS diagnostics when typeCheck=false, got %s", d.RuleName)
		}
	}
}

func TestTypeCheck_RuleNameFormat(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if !strings.HasPrefix(d.RuleName, "TypeScript(") {
			continue
		}
		// TS diagnostics should have format TypeScript(TS{code}) where code is numeric
		if !strings.HasSuffix(d.RuleName, ")") {
			t.Errorf("TS diagnostic rule name missing closing paren: %s", d.RuleName)
		}
		inner := d.RuleName[len("TypeScript(") : len(d.RuleName)-1] // extract "TS1234"
		if !strings.HasPrefix(inner, "TS") {
			t.Errorf("TS diagnostic inner code missing TS prefix: %s", d.RuleName)
		}
		for _, c := range inner[2:] {
			if c < '0' || c > '9' {
				t.Errorf("TS diagnostic rule name has non-numeric code: %s", d.RuleName)
				break
			}
		}
	}
}

func TestTypeCheck_SeverityIsError(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") && d.Severity != rule.SeverityError {
			t.Errorf("Expected TS diagnostic severity=error, got %s for %s", d.Severity, d.RuleName)
		}
	}
}

// --- Type-check with valid code ---

func TestTypeCheck_NoErrorsOnValidCode(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 42;",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			t.Errorf("Expected no TS diagnostics on valid code, got %s: %s", d.RuleName, d.Message.Description)
		}
	}
}

// --- Type-check coexists with lint rules ---

func TestTypeCheck_CoexistsWithLintRules(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		// Type error: string assigned to number
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	hasLint := false
	hasTypeCheck := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			hasTypeCheck = true
		} else {
			hasLint = true
		}
	}

	if !hasLint {
		t.Error("Expected lint diagnostics alongside type-check diagnostics")
	}
	if !hasTypeCheck {
		t.Error("Expected type-check diagnostics alongside lint diagnostics")
	}
}

// --- Type-check respects file filtering ---

// Type-check is program-scoped (tsc-aligned) and is NOT constrained by
// allowFiles. allowFiles only narrows lint-rule visit; type errors continue
// to be reported for every file in the program.
func TestTypeCheck_NotConstrainedByAllowFiles(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
		"b.ts": "const y: string = 42;",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, []string{paths["a.ts"]}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	gotA, gotB := false, false
	for _, d := range diagnostics {
		if d.SourceFile.FileName() == paths["a.ts"] {
			gotA = true
		}
		if d.SourceFile.FileName() == paths["b.ts"] {
			gotB = true
		}
	}
	if !gotA {
		t.Error("expected type error from a.ts (in allowFiles)")
	}
	if !gotB {
		t.Error("expected type error from b.ts even though it is outside allowFiles (program-scoped tsc-aligned semantics)")
	}
}

// Same idea for allowDirs: program-level type-check is unaffected.
func TestTypeCheck_NotConstrainedByAllowDirs(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"src/a.ts": "const x: number = 'hello';",
		"lib/b.ts": "const y: string = 42;",
	})

	srcDir := tmpDirPath(t, paths, "src/a.ts")
	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, []string{srcDir}, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	gotSrc, gotLib := false, false
	for _, d := range diagnostics {
		if strings.Contains(d.SourceFile.FileName(), "/src/") {
			gotSrc = true
		}
		if strings.Contains(d.SourceFile.FileName(), "/lib/") {
			gotLib = true
		}
	}
	if !gotSrc {
		t.Error("expected type error from src/a.ts (in allowDirs)")
	}
	if !gotLib {
		t.Error("expected type error from lib/b.ts even though outside allowDirs (program-scoped tsc-aligned semantics)")
	}
}

// --- Type-check via RunLinter (top-level) ---

func TestTypeCheck_RunLinter_Integration(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	_, err := runLinterPositional([]*compiler.Program{program}, true, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}

	found := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected type-check diagnostics via RunLinter")
	}
}

func TestTypeCheck_RunLinter_DisabledIntegration(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	_, err := runLinterPositional([]*compiler.Program{program}, true, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		false,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			t.Errorf("Expected no TS diagnostics when typeCheck=false, got %s", d.RuleName)
		}
	}
}

// --- Multiple error types ---

func TestTypeCheck_MultipleErrorTypes(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
const x: number = 'hello';
const y: string = 42;
const z: boolean = 'true';
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	tsCount := 0
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			tsCount++
		}
	}
	if tsCount < 3 {
		t.Errorf("Expected at least 3 type errors, got %d", tsCount)
	}
}

// --- Message content ---

func TestTypeCheck_MessageNotEmpty(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			if d.Message.Description == "" {
				t.Errorf("Expected non-empty message for %s", d.RuleName)
			}
		}
	}
}

func TestTypeCheck_MessageContainsAssignability(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") && strings.Contains(d.Message.Description, "not assignable") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected type error message to contain 'not assignable'")
	}
}

// --- Range correctness ---

func TestTypeCheck_RangeIsValid(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			if d.Range.Pos() < 0 {
				t.Errorf("Expected non-negative start position, got %d", d.Range.Pos())
			}
			if d.Range.End() <= d.Range.Pos() {
				t.Errorf("Expected end > start, got start=%d end=%d", d.Range.Pos(), d.Range.End())
			}
			if d.SourceFile == nil {
				t.Error("Expected non-nil SourceFile")
			}
		}
	}
}

// --- Edge cases ---

func TestTypeCheck_EmptyFile(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			t.Errorf("Expected no TS diagnostics on empty file, got %s", d.RuleName)
		}
	}
}

func TestTypeCheck_OnlyComments(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "// just a comment\n/* block comment */",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			t.Errorf("Expected no TS diagnostics on comments-only file, got %s", d.RuleName)
		}
	}
}

func TestTypeCheck_UndefinedVariable(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "console.log(nonExistentVar);",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		// TS2304: Cannot find name 'nonExistentVar'
		if d.RuleName == "TypeScript(TS2304)" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected TS2304 for undefined variable")
	}
}

func TestTypeCheck_PropertyAccessOnWrongType(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
const x: number = 42;
x.foo;
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		// TS2339: Property 'foo' does not exist on type 'number'
		if d.RuleName == "TypeScript(TS2339)" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected TS2339 for property access on wrong type")
	}
}

func TestTypeCheck_MultipleFiles(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
		"b.ts": "const y: string = 42;",
		"c.ts": "const z: number = 1;", // valid
	})

	filesWithErrors := make(map[string]bool)
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) {
			if strings.HasPrefix(d.RuleName, "TypeScript(") {
				filesWithErrors[d.SourceFile.FileName()] = true
			}
		}, nil,
		nil,
	)

	if len(filesWithErrors) != 2 {
		t.Errorf("Expected type errors in 2 files, got %d", len(filesWithErrors))
	}
}

// --- flattenDiagnosticMessage ---

func TestTypeCheck_MessageChainIncluded(t *testing.T) {
	// Object type mismatch produces message chains
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
interface Foo { a: number; }
const x: Foo = { a: 'hello' };
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			// Message chain produces multi-line output with indentation
			if strings.Contains(d.Message.Description, "\n") {
				found = true
				break
			}
		}
	}
	if !found {
		// Print actual messages for debugging
		for _, d := range diagnostics {
			if strings.HasPrefix(d.RuleName, "TypeScript(") {
				t.Logf("Diagnostic %s: %q", d.RuleName, d.Message.Description)
			}
		}
		t.Error("Expected multi-line message (message chain) for object type mismatch")
	}
}

// --- skipFiles filtering ---

// node_modules participation in type-check is governed by typescript-go's
// SkipTypeChecking (skipLibCheck etc.), not by rslint's lint-scope skipFiles.
// When a node_modules file is part of the program and skipLibCheck is off,
// errors there are reported.
func TestTypeCheck_NodeModulesIncludedWhenInProgram(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts":              "const x: number = 'hello';",
		"node_modules/b.ts": "const y: string = 42;",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, []string{"node_modules"},
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	var aDiags, bDiags int
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			if strings.Contains(d.SourceFile.FileName(), "node_modules") {
				bDiags++
			} else {
				aDiags++
			}
		}
	}
	if aDiags == 0 {
		t.Error("expected type errors from a.ts")
	}
	if bDiags == 0 {
		t.Error("expected node_modules/b.ts type errors to also be reported (program-scoped tsc-aligned semantics; ExcludePaths is a lint-only filter)")
	}
}

// --- @ts-expect-error / @ts-ignore ---

func TestTypeCheck_TsExpectErrorSuppresses(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
// @ts-expect-error
const x: number = 'hello';
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") && strings.Contains(d.Message.Description, "not assignable") {
			t.Errorf("Expected @ts-expect-error to suppress type error, got: %s", d.Message.Description)
		}
	}
}

func TestTypeCheck_TsIgnoreSuppresses(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
// @ts-ignore
const x: number = 'hello';
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") && strings.Contains(d.Message.Description, "not assignable") {
			t.Errorf("Expected @ts-ignore to suppress type error, got: %s", d.Message.Description)
		}
	}
}

func TestTypeCheck_TsExpectErrorUnused(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
// @ts-expect-error
const x: number = 42;
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	// TS2578: Unused '@ts-expect-error' directive
	found := false
	for _, d := range diagnostics {
		if d.RuleName == "TypeScript(TS2578)" {
			found = true
			break
		}
	}
	if !found {
		// Print for debugging
		for _, d := range diagnostics {
			t.Logf("Diagnostic: %s: %s", d.RuleName, d.Message.Description)
		}
		t.Error("Expected TS2578 for unused @ts-expect-error directive")
	}
}

// --- No fixes/suggestions on type-check diagnostics ---

func TestTypeCheck_NoFixesOrSuggestions(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			if d.FixesPtr != nil {
				t.Errorf("Expected no fixes on TS diagnostic %s, got %v", d.RuleName, d.FixesPtr)
			}
			if d.Suggestions != nil {
				t.Errorf("Expected no suggestions on TS diagnostic %s, got %v", d.RuleName, d.Suggestions)
			}
		}
	}
}

// --- lintedFileCount correctness ---

func TestTypeCheck_LintedFileCount(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
		"b.ts": "const y: string = 42;",
		"c.ts": "const z = 1;",
	})

	count := RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if count != 3 {
		t.Errorf("Expected lintedFileCount=3, got %d", count)
	}
}

func TestTypeCheck_LintedFileCountWithAllowFiles(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
		"b.ts": "const y: string = 42;",
	})

	count := RunLinterInProgram(program, []string{paths["a.ts"]}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) {}, nil,
		nil,
	)

	if count != 1 {
		t.Errorf("Expected lintedFileCount=1 with allowFiles, got %d", count)
	}
}

// --- SourceFile pointer matches correct file ---

func TestTypeCheck_SourceFileMatchesDiagnosticOrigin(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
		"b.ts": "const y: number = 1;", // valid
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			if d.SourceFile.FileName() != paths["a.ts"] {
				t.Errorf("Expected TS diagnostic from a.ts, got from %s", d.SourceFile.FileName())
			}
		}
	}
}

// --- Cross-file type errors ---

func TestTypeCheck_CrossFileTypeError(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"lib.ts": "export const value: number = 42;",
		"main.ts": `
import { value } from './lib';
const x: string = value;
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") && strings.Contains(d.SourceFile.FileName(), "main.ts") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected cross-file type error in main.ts")
	}
}

// --- Function argument type mismatch (TS2345) ---

func TestTypeCheck_ArgumentTypeMismatch(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
function greet(name: string): void {}
greet(123);
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		// TS2345: Argument of type 'number' is not assignable to parameter of type 'string'
		if d.RuleName == "TypeScript(TS2345)" {
			found = true
			break
		}
	}
	if !found {
		for _, d := range diagnostics {
			t.Logf("Diagnostic: %s: %s", d.RuleName, d.Message.Description)
		}
		t.Error("Expected TS2345 for argument type mismatch")
	}
}

// --- Function return type error ---

func TestTypeCheck_ReturnTypeMismatch(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
function getNum(): number {
  return 'hello';
}
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") && strings.Contains(d.Message.Description, "not assignable") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected type error for return type mismatch")
	}
}

// --- typeCheck=true with type-safe code: only lint diagnostics ---

func TestTypeCheck_TypeSafeCodeOnlyLintDiagnostics(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 42;",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return noopRule() },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	hasLint := false
	hasTypeCheck := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			hasTypeCheck = true
		} else {
			hasLint = true
		}
	}

	if !hasLint {
		t.Error("Expected lint diagnostics on type-safe code with noopRule")
	}
	if hasTypeCheck {
		t.Error("Expected no type-check diagnostics on type-safe code")
	}
}

// --- Multiple programs with type-check ---

func TestTypeCheck_MultiplePrograms(t *testing.T) {
	program1, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})
	program2, _ := createTestProgramWithFiles(t, map[string]string{
		"b.ts": "const y: string = 42;",
	})

	var mu sync.Mutex
	var diagnostics []rule.RuleDiagnostic
	result, err := runLinterPositional(
		[]*compiler.Program{program1, program2},
		true,
		nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) {
			mu.Lock()
			diagnostics = append(diagnostics, d)
			mu.Unlock()
		}, nil,
		nil,
	)

	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if result.LintedFileCount < 2 {
		t.Errorf("Expected lintedFileCount >= 2, got %d", result.LintedFileCount)
	}

	tsCount := 0
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			tsCount++
		}
	}
	if tsCount < 2 {
		t.Errorf("Expected at least 2 type errors across programs, got %d", tsCount)
	}
}

// --- Range points to correct token ---

func TestTypeCheck_RangePointsToCorrectToken(t *testing.T) {
	code := "const x: number = 'hello';"
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": code,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			start := d.Range.Pos()
			end := d.Range.End()
			fileText := d.SourceFile.Text()
			if end > len(fileText) {
				t.Errorf("Range end %d exceeds file length %d", end, len(fileText))
				continue
			}
			snippet := fileText[start:end]
			// TS2322 points to the assignment target (variable name), not the value
			if d.RuleName == "TypeScript(TS2322)" {
				if snippet != "x" {
					t.Errorf("Expected TS2322 range to point to 'x', got %q", snippet)
				}
			}
		}
	}
}

// --- No duplicate diagnostics ---

func TestTypeCheck_NoDuplicateDiagnostics(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	type diagKey struct {
		ruleName string
		pos      int
		end      int
		file     string
	}
	seen := make(map[diagKey]bool)
	for _, d := range diagnostics {
		if !strings.HasPrefix(d.RuleName, "TypeScript(") {
			continue
		}
		key := diagKey{
			ruleName: d.RuleName,
			pos:      d.Range.Pos(),
			end:      d.Range.End(),
			file:     d.SourceFile.FileName(),
		}
		if seen[key] {
			t.Errorf("Duplicate diagnostic: %s at %d-%d in %s", d.RuleName, d.Range.Pos(), d.Range.End(), d.SourceFile.FileName())
		}
		seen[key] = true
	}
}

// --- allowFiles empty slice (non-nil) blocks type-check ---

// Empty allowFiles blocks all lint-rule visits (LintedFileCount=0) but does
// NOT suppress the program-level type-check phase. Type errors are still
// reported, mirroring tsc.
func TestTypeCheck_AllowFilesEmptySliceBlocksLintOnly(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	count := RunLinterInProgram(program, []string{}, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	if count != 0 {
		t.Errorf("Expected lintedFileCount=0 (lint-rule visits), got %d", count)
	}
	tsCount := 0
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			tsCount++
		}
	}
	if tsCount == 0 {
		t.Error("expected at least one type-check diagnostic; type-check is program-level and must not be silenced by an empty allowFiles slice")
	}
}

// --- Specific TS2322 error code ---

func TestTypeCheck_TS2322_TypeAssignment(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		if d.RuleName == "TypeScript(TS2322)" {
			found = true
			break
		}
	}
	if !found {
		for _, d := range diagnostics {
			t.Logf("Diagnostic: %s: %s", d.RuleName, d.Message.Description)
		}
		t.Error("Expected TS2322 for type assignment error")
	}
}

// --- Interface missing required property ---

func TestTypeCheck_MissingProperty(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
interface Foo { a: number; b: string; }
const x: Foo = { a: 1 };
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		// TS2741: Property 'b' is missing in type '{ a: number; }' but required in type 'Foo'
		if strings.HasPrefix(d.RuleName, "TypeScript(") && strings.Contains(d.Message.Description, "missing") {
			found = true
			break
		}
	}
	if !found {
		for _, d := range diagnostics {
			t.Logf("Diagnostic: %s: %s", d.RuleName, d.Message.Description)
		}
		t.Error("Expected type error for missing property")
	}
}

// --- Generic type error ---

func TestTypeCheck_GenericTypeError(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
function identity<T>(x: T): T { return x; }
const result: string = identity<number>(42);
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") && strings.Contains(d.Message.Description, "not assignable") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected type error for generic type mismatch")
	}
}

// --- Enum type error ---

func TestTypeCheck_EnumTypeMismatch(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
enum Color { Red, Green, Blue }
const c: Color = 'red';
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected type error for enum type mismatch")
	}
}

// --- Tuple type error ---

func TestTypeCheck_TupleTypeMismatch(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
const x: [number, string] = [1, 2];
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected type error for tuple type mismatch")
	}
}

// --- Union type narrowing error ---

func TestTypeCheck_ExcessPropertyCheck(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": `
interface Foo { a: number; }
const x: Foo = { a: 1, b: 2 };
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		// TS2353 or TS2322: excess property
		if strings.HasPrefix(d.RuleName, "TypeScript(") && (strings.Contains(d.Message.Description, "not assignable") || strings.Contains(d.Message.Description, "does not exist")) {
			found = true
			break
		}
	}
	if !found {
		for _, d := range diagnostics {
			t.Logf("Diagnostic: %s: %s", d.RuleName, d.Message.Description)
		}
		t.Error("Expected type error for excess property")
	}
}

// --- Type-check with JSX ---

func TestTypeCheck_JSXTypeError(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.tsx": `
const x: number = 'hello';
`,
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected type-check to work on .tsx files")
	}
}

// --- Message.Id should be empty for type-check diagnostics ---

func TestTypeCheck_MessageIdEmpty(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
	})

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		nil,
	)

	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") && d.Message.Id != "" {
			t.Errorf("Expected empty Message.Id for TS diagnostic, got %q", d.Message.Id)
		}
	}
}

// fileFilter rejecting a file suppresses lint-rule diagnostics and excludes
// the file from LintedFileCount, but does NOT silence type-check
// diagnostics. Type-check is program-scoped and aligned with tsc; the
// per-program filter is a lint-rule concept (config `ignores`,
// multi-config ownership) and must not be re-applied to type errors.
func TestTypeCheck_FileFilterSuppressesLintRulesOnly(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"kept.ts":    "const x: number = 'hello';",
		"ignored.ts": "const y: number = 'world';",
	})

	ignoredPath := paths["ignored.ts"]
	keptPath := paths["kept.ts"]

	var diagnostics []rule.RuleDiagnostic
	count := RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }, nil,
		func(fileName string) bool { return fileName != ignoredPath },
	)

	// Lint-rule visit count excludes the filtered file.
	if count != 1 {
		t.Errorf("LintedFileCount = %d, want 1 (filter applies to lint phase)", count)
	}

	// Both files contribute type errors — the filter is lint-only.
	var keptDiags, ignoredDiags int
	for _, d := range diagnostics {
		if !strings.HasPrefix(d.RuleName, "TypeScript(") {
			continue
		}
		switch d.SourceFile.FileName() {
		case keptPath:
			keptDiags++
		case ignoredPath:
			ignoredDiags++
		}
	}
	if keptDiags == 0 {
		t.Error("expected kept.ts to produce TS diagnostics")
	}
	if ignoredDiags == 0 {
		t.Error("expected ignored.ts to STILL produce TS diagnostics — type-check is not gated by per-program filter")
	}
}

// Parallel scenario: even when every program has a reject-all filter, the
// type-check phase runs on each program. Lint-rule visits collapse to zero,
// but type errors are still surfaced.
func TestTypeCheck_FileFilterDoesNotMaskTypeChecksAcrossPrograms(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';",
		"b.ts": "const y: number = 'world';",
	})

	rejectAll := func(string) bool { return false }

	var diagCount int
	var mu sync.Mutex
	result, err := runLinterPositional(
		[]*compiler.Program{program},
		true,
		nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) {
			mu.Lock()
			defer mu.Unlock()
			if strings.HasPrefix(d.RuleName, "TypeScript(") {
				diagCount++
			}
		},
		nil,
		[]func(string) bool{rejectAll},
	)
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if result.LintedFileCount != 0 {
		t.Errorf("LintedFileCount = %d, want 0 (filter rejects all → no lint visits)", result.LintedFileCount)
	}
	if diagCount < 2 {
		t.Errorf("expected at least 2 type-check diagnostics (a.ts + b.ts), got %d", diagCount)
	}
	_ = paths["a.ts"]
	_ = paths["b.ts"]
}
