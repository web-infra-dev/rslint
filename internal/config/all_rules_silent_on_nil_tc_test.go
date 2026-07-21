package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestAllRules_SilentOnNilTypeCheckerImpliesRequiresTypeInfo is the runtime
// counterpart to TestAllRules_NilTypeCheckerEarlyReturnImpliesRequiresTypeInfo.
//
// The static check matches `if ctx.TypeChecker == nil { return RuleListeners{} }`
// at the top of Run. Some rules instead push the nil-TC gate into a helper
// (e.g. no-obj-calls's checkCallee, no-const-assign's checkIdentifierWrite,
// no-ex-assign's checkReassignments, no-use-before-define's checkIdentifier)
// — every listener funnels through that helper, so the rule emits zero
// diagnostics without TC even though Run itself is unguarded.
//
// This test runs each rule on a fixture *known to trigger it* under two
// configurations:
//
//  1. The rule's normal Run with a real, non-nil TypeChecker — must produce
//     ≥1 diagnostic (proves the fixture is correct and the rule actually
//     fires).
//  2. The same fixture but with `ctx.TypeChecker` forcibly set to nil before
//     calling Run — if the rule emits 0 diagnostics here while emitting ≥1
//     above, the rule is silently broken in CLI gap-file mode and would also
//     misbehave in LSP inferred-project mode, so it MUST declare
//     RequiresTypeInfo: true.
//
// Adding a new rule to this table is the regression guard: if you write a
// rule whose logic is meaningless without a TypeChecker, this test forces you
// to declare the flag.
func TestAllRules_SilentOnNilTypeCheckerImpliesRequiresTypeInfo(t *testing.T) {
	RegisterAllRules()
	registry := GlobalRuleRegistry.GetAllRules()

	cases := []struct {
		ruleKey  string
		fileName string
		source   string
	}{
		{
			ruleKey:  "no-undef",
			fileName: "no-undef.ts",
			source:   `someUndefinedThing();`,
		},
		{
			ruleKey:  "no-obj-calls",
			fileName: "no-obj-calls.ts",
			source:   `Math();`,
		},
		{
			ruleKey:  "no-const-assign",
			fileName: "no-const-assign.ts",
			source:   `const x = 1; x = 2;`,
		},
		{
			ruleKey:  "no-ex-assign",
			fileName: "no-ex-assign.ts",
			source:   `try { throw 1; } catch (e) { e = 2; }`,
		},
		{
			ruleKey:  "prefer-const",
			fileName: "prefer-const.ts",
			source:   `let neverReassigned = 1; console.log(neverReassigned);`,
		},
		{
			ruleKey:  "no-unmodified-loop-condition",
			fileName: "no-unmodified-loop-condition.ts",
			source:   `let n = 1; while (n < 10) { console.log(n); }`,
		},
		{
			ruleKey:  "no-loop-func",
			fileName: "no-loop-func.ts",
			source:   `for (var i = 0; i < 3; i++) { setTimeout(function() { console.log(i); }); }`,
		},
		{
			ruleKey:  "@typescript-eslint/no-use-before-define",
			fileName: "no-use-before-define.ts",
			source:   `useBefore(); function useBefore() {}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.ruleKey, func(t *testing.T) {
			impl, ok := registry[tc.ruleKey]
			if !ok {
				t.Fatalf("rule %q is not registered", tc.ruleKey)
			}

			withTC := countDiagnosticsForRule(t, tc.fileName, tc.source, impl, true)
			if withTC == 0 {
				t.Fatalf("fixture for %q produced 0 diagnostics with a real TypeChecker; the test fixture is wrong (it must trigger the rule so the without-TC comparison is meaningful)", tc.ruleKey)
			}

			withoutTC := countDiagnosticsForRule(t, tc.fileName, tc.source, impl, false)
			if withoutTC == 0 && !impl.RequiresTypeInfo {
				t.Fatalf("rule %q emits %d diagnostics with TypeChecker but 0 without; it is silently useless on gap files / LSP inferred-project files and MUST declare RequiresTypeInfo: true", tc.ruleKey, withTC)
			}
		})
	}
}

// countDiagnosticsForRule runs a single rule on a single-file program and
// returns how many diagnostics it produced. When withTypeChecker is false, the
// rule receives a nil TypeChecker — simulating the gap-file / inferred-project
// path that the existing FilterNonTypeAwareRules infrastructure is meant to
// guard against.
func countDiagnosticsForRule(t *testing.T, fileName, source string, impl rule.Rule, withTypeChecker bool) int {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := tspath.NormalizePath(filepath.Join(tmpDir, fileName))
	if err := os.WriteFile(filePath, []byte(source), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	compilerOptions := &core.CompilerOptions{
		Target:          core.ScriptTargetESNext,
		Module:          core.ModuleKindCommonJS,
		ESModuleInterop: core.TSTrue,
		SkipLibCheck:    core.TSTrue,
	}
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgramFromOptionsLenient(true, compilerOptions, []string{filePath}, host)
	if err != nil {
		t.Fatalf("create program: %v", err)
	}

	configured := linter.ConfiguredRule{
		Name:     impl.Name,
		Severity: rule.SeverityWarning,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return impl.Run(ctx, nil)
		},
	}

	var typeInfoFiles map[string]struct{}
	if withTypeChecker {
		// nil typeInfoFiles → linter passes the real TypeChecker to every rule.
		typeInfoFiles = nil
	} else {
		// Non-nil but empty → file is treated as a gap file, so the linter
		// substitutes a nil TypeChecker (matching CLI gap-file behavior).
		typeInfoFiles = map[string]struct{}{}
	}

	count := 0
	linter.RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []linter.ConfiguredRule {
			if sf.FileName() != filePath {
				return nil
			}
			return []linter.ConfiguredRule{configured}
		},
		false,
		func(d rule.RuleDiagnostic) { count++ },
		typeInfoFiles,
		nil,
	)
	return count
}

// Compile-time anchor: keep the linter import live in case future trims of
// the imports remove the only call site by accident.
var _ = compiler.Program{}
