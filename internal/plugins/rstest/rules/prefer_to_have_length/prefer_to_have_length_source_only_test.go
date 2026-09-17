package prefer_to_have_length

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferToHaveLengthSourceOnlyAndEditDemand(t *testing.T) {
	t.Parallel()
	code := `import { expect as check, test } from '@rstest/core';
check(values.length).toBe(2);
check([1, 2].length).toBe(2);
test('context', ({ expect }) => { expect(value.length).toStrictEqual(2); });
function shadow(check: any) { check(value.length).toBe(2); }
let { expect: changed } = require('@rstest/core');
changed = replacement;
changed(value.length).toBe(2);`

	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "prefer-to-have-length-source-only.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames: []string{fileName},
		Host:          utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: &core.CompilerOptions{
			Module: core.ModuleKindESNext,
		},
		SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("expected source-only program")
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: program,
			File:    fileName,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     PreferToHaveLengthRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return PreferToHaveLengthRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: got %d diagnostics, want 3", demand, len(diagnostics))
		}
		return diagnostics
	}

	baseline := run(rule.EditDemandNone)
	for _, demand := range []rule.EditDemand{rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		diagnostics := run(demand)
		for index := range diagnostics {
			wantFix := (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll) && index == 1
			if (diagnostics[index].FixesPtr != nil) != wantFix {
				t.Fatalf("demand %d diagnostic %d: unexpected fix materialization", demand, index)
			}
			diagnostics[index].FixesPtr = nil
		}
		if !reflect.DeepEqual(baseline, diagnostics) {
			t.Fatalf("demand %d changed diagnostics", demand)
		}
	}
}
