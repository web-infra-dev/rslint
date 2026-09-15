// TestNoAsyncMockFactorySourceOnly locks the rule's behaviour on a program that
// carries no TypeChecker. The two syntax layers have to keep working there, and
// the type layer has to fall silent rather than panic.
package no_async_mock_factory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/testutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoAsyncMockFactorySourceOnly(t *testing.T) {
	testCases := []struct {
		name   string
		source string
		want   int
	}{
		{
			name:   "async arrow",
			source: `rs.mock('./sum', async () => ({ sum: () => 0 }));`,
			want:   1,
		},
		{
			name:   "promise-producing body",
			source: `rs.doMock('./sum', () => Promise.resolve({ sum: 0 }));`,
			want:   1,
		},
		{
			name: "declaration in this file",
			source: `const factory = async () => ({ sum: () => 0 });
rs.mock('./sum', factory);`,
			want: 1,
		},
		{
			// The syntax layer matches `Promise` by the name as written, and
			// this file declares its own. With no types left to ask, the rule
			// stays silent rather than guess.
			name: "shadowed Promise",
			source: `const Promise = { resolve: (value: unknown) => value };
rs.mock('./sum', () => Promise.resolve({ sum: 0 }));`,
			want: 0,
		},
		{
			// `importActual` is one of the two members whose rewrite honours an
			// ordinary binding of the receiver.
			name: "locally declared receiver",
			source: `const rs = { mock: (p: string, f: () => unknown) => f(), importActual: () => ({ sum: 0 }) };
rs.mock('./sum', () => rs.importActual('./sum'));`,
			want: 0,
		},
		{
			// The binder's reachability is available without a TypeChecker, so
			// the fall-through path is seen here too.
			name:   "path that returns nothing",
			source: `rs.doMockRequire('./sum', () => { if (flag) return Promise.resolve({ sum: 0 }); });`,
			want:   0,
		},
		{
			name:   "synchronous factory",
			source: `rs.mock('./sum', () => ({ sum: () => 0 }));`,
			want:   0,
		},
		{
			// Without a TypeChecker there is nothing left to ask, so the rule
			// stays silent. This is the one shape it knowingly misses.
			name:   "undecidable without types",
			source: `rs.mock('./sum', buildMock);`,
			want:   0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := lintSourceOnly(t, testCase.source); got != testCase.want {
				t.Fatalf("source-only diagnostics = %d, want %d", got, testCase.want)
			}
		})
	}
}

func lintSourceOnly(t *testing.T, source string) int {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := tspath.NormalizePath(filepath.Join(tmpDir, "case.ts"))
	if err := os.WriteFile(filePath, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Target:          core.ScriptTargetESNext,
		Module:          core.ModuleKindCommonJS,
		ESModuleInterop: core.TSTrue,
		SkipLibCheck:    core.TSTrue,
	}, []string{filePath}, utils.CreateCompilerHost(tmpDir, bundled.WrapFS(cachedvfs.From(osvfs.FS()))))
	if err != nil {
		t.Fatal(err)
	}
	sourceProgram, err := lintprogram.NewFromBoundSources(program, program.SourceFiles())
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	testutil.LintProgram(t, testutil.LintProgramOptions{
		Program:                sourceProgram,
		ExcludedPathSubstrings: testutil.DefaultExcludedPathSubstrings,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []rule.ConfiguredRule {
			if sourceFile.FileName() != filePath {
				return nil
			}
			return []rule.ConfiguredRule{{
				Name:     NoAsyncMockFactoryRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Fatal("expected a program without a TypeChecker")
					}
					return NoAsyncMockFactoryRule.Run(ctx, nil)
				},
			}}
		},
		OnDiagnostic: func(rule.RuleDiagnostic) { count++ },
	})
	return count
}
