package linter

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// typeCheckerRule returns a rule that records whether TypeChecker was nil or not.
func typeCheckerRule(checkerWasNil *bool) []ConfiguredRule {
	return []ConfiguredRule{
		{
			Name:     "checker-probe",
			Severity: rule.SeverityWarning,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				*checkerWasNil = (ctx.TypeChecker == nil)
				return rule.RuleListeners{}
			},
		},
	}
}

func TestTypeInfoFiles_FileInSet_GetsTypeChecker(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
	})

	typeInfoFiles := map[string]struct{}{
		paths["a.ts"]: {},
	}

	var checkerWasNil bool
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			return typeCheckerRule(&checkerWasNil)
		},
		false,
		func(d rule.RuleDiagnostic) {},
		typeInfoFiles,
		nil,
	)

	if checkerWasNil {
		t.Error("TypeChecker should NOT be nil for files in typeInfoFiles")
	}
}

func TestTypeInfoFiles_FileNotInSet_NilTypeChecker(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
	})

	// typeInfoFiles exists but does NOT contain a.ts → gap file
	typeInfoFiles := map[string]struct{}{
		"/some/other/file.ts": {},
	}

	var checkerWasNil bool
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			return typeCheckerRule(&checkerWasNil)
		},
		false,
		func(d rule.RuleDiagnostic) {},
		typeInfoFiles,
		nil,
	)

	if !checkerWasNil {
		t.Error("TypeChecker should be nil for gap files (not in typeInfoFiles)")
	}

	// Verify the file was still linted (rule ran, just with nil checker)
	_ = paths
}

func TestTypeInfoFiles_Nil_AllGetTypeChecker(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
	})

	var checkerWasNil bool
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			return typeCheckerRule(&checkerWasNil)
		},
		false,
		func(d rule.RuleDiagnostic) {},
		nil, // nil means no gap-file distinction
		nil,
	)

	if checkerWasNil {
		t.Error("TypeChecker should NOT be nil when typeInfoFiles is nil")
	}
}

func TestTypeCheck_TypeInfoFilesDoesNotRestrictProgramWideDiagnostics(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';", // type error
	})

	// TypeInfoFiles gates lint-rule TypeChecker access only. Program-wide
	// type-check diagnostics are controlled by the Program skip mask.
	typeInfoFiles := map[string]struct{}{
		"/some/other/file.ts": {},
	}

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true, // typeCheck enabled
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
		typeInfoFiles,
		nil,
	)

	found := false
	for _, d := range diagnostics {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			found = true
		}
	}
	if !found {
		t.Error("expected program-wide semantic diagnostics despite the lint TypeInfoFiles set")
	}
	_ = paths
}

func TestTypeCheck_ReportsSemanticDiagnosticsForTypeInfoFiles(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x: number = 'hello';", // type error
	})

	// a.ts IS in typeInfoFiles → should get semantic diagnostics
	typeInfoFiles := map[string]struct{}{
		paths["a.ts"]: {},
	}

	var diagnostics []rule.RuleDiagnostic
	RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		true,
		func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
		typeInfoFiles,
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
		t.Error("Files in typeInfoFiles should get semantic diagnostics")
	}
}
