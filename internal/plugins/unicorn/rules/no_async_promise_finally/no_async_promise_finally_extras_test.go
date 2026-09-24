package no_async_promise_finally_test

import (
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_async_promise_finally"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoAsyncPromiseFinallyExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_async_promise_finally.NoAsyncPromiseFinallyRule,
		[]rule_tester.ValidTestCase{
			validTS("type Callback = () => void; promise.finally((async function * () {}) as Callback);"),
			validTS("function foo(object: {finally(handler: () => void): void} | {finally(handler: () => Promise<void>): void}) { object.finally(async () => {}); }"),
			valid("const method = getMethod(); promise[method](async () => {});"),
			valid("const cleanup = object.cleanup; promise.finally(cleanup);"),
		},
		[]rule_tester.InvalidTestCase{
			invalid("const run = async function cleanup() { promise.finally(cleanup); };", "cleanup"),
			invalidTS("type Callback = () => void; promise.finally((async () => {})!);", "(async () => {})!"),
			invalidTS("type Callback = () => void; promise.finally(<Callback>(async () => {}));", "<Callback>(async () => {})"),
			invalidTS("type Callback = () => void; const cleanup = (async () => {}) satisfies Callback; promise.finally(cleanup);", "cleanup"),
			invalidTS("function foo(promise: any) { promise.finally(async () => {}); }", "async () => {}"),
			invalidTS("function foo(promise: PromiseLike<string>) { promise.finally(async () => {}); }", "async () => {}"),
			invalidTS("function foo(promise: MissingPromiseType) { promise.finally(async () => {}); }", "async () => {}"),
			invalidTS("function foo(promise: Promise<string> | {finally(handler: () => void): void}) { promise.finally(async () => {}); }", "async () => {}"),
		},
	)
}

func TestNoAsyncPromiseFinallySourceOnly(t *testing.T) {
	code := "const method = \"finally\"; const cleanup = async () => {}; promise[method](cleanup); async function declared() {} promise.finally(declared); const run = async function named() { promise.finally(named); };"
	diagnostics := lintNoAsyncPromiseFinallySourceOnly(t, code)
	if len(diagnostics) != 3 {
		t.Fatalf("project:false diagnostics = %d, want 3: %+v", len(diagnostics), diagnostics)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Message.Id != "no-async-promise-finally" {
			t.Fatalf("project:false message id = %q", diagnostic.Message.Id)
		}
	}
}

func lintNoAsyncPromiseFinallySourceOnly(t *testing.T, code string) []rule.RuleDiagnostic {
	t.Helper()
	dir := tspath.NormalizePath(t.TempDir())
	fileName := tspath.NormalizePath(filepath.Join(dir, "file.js"))
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            utils.CreateCompilerHost(dir, fs),
		CompilerOptions: &core.CompilerOptions{Target: core.ScriptTargetESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatalf("create project:false program: %v", err)
	}
	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("project:false fixture unexpectedly received a TypeChecker")
	}

	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program,
		File:    fileName,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     no_async_promise_finally.NoAsyncPromiseFinallyRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Fatal("project:false fixture unexpectedly received a TypeChecker")
					}
					return no_async_promise_finally.NoAsyncPromiseFinallyRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandNone,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}
