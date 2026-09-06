// TestRequireToThrowMessageResolvesLocalShadowInSourceOnlyProgram verifies the
// rule's provenance decision when no tsconfig supplies a TypeChecker. The main
// upstream and edge-shape matrices live in the sibling upstream/extras files.
package require_to_throw_message

import (
	"sort"
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

func TestRequireToThrowMessageResolvesLocalShadowInSourceOnlyProgram(t *testing.T) {
	if RequireToThrowMessageRule.RequiresTypeInfo {
		t.Fatal("rstest/require-to-throw-message must run without type information")
	}

	tests := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "global and local shadow",
			code: `expect(globalRun).toThrow();
{
  const expect = createAssertionLibrary();
  expect(localRun).toThrow();
}`,
			want: []string{"toThrow"},
		},
		{
			name: "named import",
			code: `import { expect } from '@rstest/core';
expect(run).toThrow();`,
			want: []string{"toThrow"},
		},
		{
			name: "require destructuring",
			code: `const { expect } = require('@rstest/core');
expect(run).toThrowError();`,
			want: []string{"toThrowError"},
		},
		{
			name: "import meta destructuring",
			code: `const { expect } = import.meta.rstest;
expect(run).toThrow();`,
			want: []string{"toThrow"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runSourceOnlyCase(t, test.code, test.want)
		})
	}
}

func runSourceOnlyCase(t *testing.T, code string, want []string) {
	t.Helper()
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "require-to-throw-message-source-only.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	host := utils.CreateCompilerHost(root.Dir, fs)
	sourceProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            host,
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatalf("NewFromRoots: %v", err)
	}
	if sourceProgram.CanProvideTypeChecker(sourceProgram.SourceFiles()[0]) {
		t.Fatal("expected a source-only Program with no TypeChecker")
	}

	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{sourceProgram},
		TargetsByProgram: [][]string{{fileName}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     RequireToThrowMessageRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return RequireToThrowMessageRule.Run(ctx, nil)
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}

	type reported struct {
		pos  int
		text string
	}
	var got []reported
	if _, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{
			Report: func(diagnostic rule.RuleDiagnostic) {
				got = append(got, reported{
					pos:  diagnostic.Range.Pos(),
					text: code[diagnostic.Range.Pos():diagnostic.Range.End()],
				})
			},
		},
	}); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	sort.Slice(got, func(left, right int) bool { return got[left].pos < got[right].pos })

	if len(got) != len(want) {
		t.Fatalf("reported %d assertions, want %d: %v", len(got), len(want), got)
	}
	for index, expected := range want {
		if got[index].text != expected {
			t.Errorf("diagnostic %d anchored to %q, want %q", index, got[index].text, expected)
		}
	}
}
