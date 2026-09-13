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
		{
			name: "for of binding without initializer shadows outer fragment alias",
			code: "const F=React.Fragment; for (const F of widgets) { <F/>; }",
		},
		{
			name: "destructured for of binding without initializer shadows outer fragment alias",
			code: "const F=React.Fragment; for (const {F} of widgets) { <F/>; }",
		},
		{
			name: "catch parameter shadows outer fragment alias",
			code: "const F=React.Fragment; try {} catch (F) { <F/>; }",
		},
		{
			name: "uninitialized variable is not a fragment alias",
			code: "let F; <F/>;",
		},
		{
			name: "first variable definition wins over later function declaration",
			code: "function render(){ var F=React.Fragment; function F(){} return <F/>; }",
			want: 1,
		},
		{
			name: "first function definition wins over later variable declaration",
			code: "function render(){ function F(){} var F=React.Fragment; return <F/>; }",
		},
		{
			name: "first non-fragment variable definition wins",
			code: "function render(){ var F=Widget; var F=React.Fragment; return <F/>; }",
		},
		{
			name: "first fragment variable definition wins",
			code: "function render(){ var F=React.Fragment; var F=Widget; return <F/>; }",
			want: 1,
		},
		{
			name: "first child scope is searched",
			code: "function inner(){ const F=React.Fragment; } <F/>;",
			want: 1,
		},
		{
			name: "first grandchild scope is searched",
			code: "function first(){ function second(){ const F=React.Fragment; } } <F/>;",
			want: 1,
		},
		{
			name: "great grandchild scope is not searched",
			code: "function first(){ function second(){ function third(){ const F=React.Fragment; } } } <F/>;",
		},
		{
			name: "non-first child scope is not searched",
			code: "function first(){} function second(){ const F=React.Fragment; } <F/>;",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root := fixtures.GetRootDir()
			fileName := tspath.ResolvePath(root.Dir, "jsx-fragments-source-only.jsx")
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
