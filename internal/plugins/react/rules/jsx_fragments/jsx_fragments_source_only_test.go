package jsx_fragments

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestJsxFragmentsSourceOnlyBindings(t *testing.T) {
	for _, testCase := range []struct {
		name string
		code string
		want int
	}{
		{
			name: "unrelated inner fragment alias does not match outer component",
			code: "function f(){const F=React.Fragment;} const F=Widget; <F/>;",
		},
		{
			name: "parameter shadows outer fragment alias",
			code: "const F=React.Fragment; function f(F){return <F/>;}",
		},
		{
			name: "outer fragment alias resolves after inner declaration",
			code: "function f(){const F=Widget;} const F=React.Fragment; <F/>;",
			want: 1,
		},
		{
			name: "later fragment declaration resolves",
			code: "<F/>; const F=React.Fragment;",
			want: 1,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root := fixtures.GetRootDir()
			fileName := tspath.ResolvePath(root.Dir, "jsx-fragments-source-only.tsx")
			fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: testCase.code})
			program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
				RootFileNames: []string{fileName},
				Host:          utils.CreateCompilerHost(root.Dir, fs),
				CompilerOptions: &core.CompilerOptions{
					Jsx:    core.JsxEmitPreserve,
					Module: core.ModuleKindESNext,
				},
				SingleThreaded: true,
			})
			if err != nil {
				t.Fatal(err)
			}

			count := 0
			linter.LintSingleFile(linter.LintSingleFileOptions{
				Program: program,
				File:    fileName,
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{
						Name:     JsxFragmentsRule.Name,
						Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							if ctx.TypeChecker != nil {
								t.Fatal("expected source-only linting without a type checker")
							}
							return JsxFragmentsRule.Run(ctx, nil)
						},
					}}
				},
				Consumer: rule.DiagnosticConsumer{
					Report: func(rule.RuleDiagnostic) { count++ },
				},
			})
			if count != testCase.want {
				t.Fatalf("diagnostics = %d, want %d", count, testCase.want)
			}
		})
	}
}
