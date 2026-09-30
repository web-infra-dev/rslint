package no_async_promise_finally_test

import (
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
			invalid("promise.finally((async () => {}));", "async () => {}"),
			invalid("const cleanup = async () => {}; promise.finally((cleanup));", "cleanup"),
			invalidTS("type Callback = () => void; promise.finally(((async () => {}) as Callback));", "(async () => {}) as Callback"),
			invalidTS("type Callback = () => void; promise.finally(((async () => {}) satisfies Callback));", "(async () => {}) satisfies Callback"),
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
	diagnostics := lintNoAsyncPromiseFinally(t, code, false)
	if len(diagnostics) != 3 {
		t.Fatalf("project:false diagnostics = %d, want 3: %+v", len(diagnostics), diagnostics)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Message.Id != "no-async-promise-finally" {
			t.Fatalf("project:false message id = %q", diagnostic.Message.Id)
		}
	}
}

func TestNoAsyncPromiseFinallyJSDocRanges(t *testing.T) {
	for _, mode := range []struct {
		name      string
		typeAware bool
	}{
		{name: "source-only"},
		{name: "type-aware", typeAware: true},
	} {
		t.Run(mode.name, func(t *testing.T) {
			for _, testCase := range []struct {
				name       string
				code       string
				reportText string
			}{
				{
					name:       "type-cast-arrow",
					code:       "promise.finally(/** @type {() => Promise<void>} */ (((async () => {}))));",
					reportText: "async () => {}",
				},
				{
					name:       "type-cast-reference",
					code:       "const cleanup = async () => {}; promise.finally(/** @type {() => Promise<void>} */ ((cleanup)));",
					reportText: "cleanup",
				},
				{
					name:       "satisfies-cast-arrow",
					code:       "promise.finally(/** @satisfies {() => Promise<void>} */ (((async () => {}))));",
					reportText: "async () => {}",
				},
				{
					name: "disabled-type-cast-arrow",
					code: "promise.finally(\n  /** @type {() => Promise<void>} */ (\n    (\n      // eslint-disable-next-line unicorn/no-async-promise-finally\n      async () => {}\n    )\n  )\n);",
				},
				{
					name: "disabled-type-cast-reference",
					code: "const cleanup = async () => {};\npromise.finally(\n  /** @type {() => Promise<void>} */ (\n    (\n      // eslint-disable-next-line unicorn/no-async-promise-finally\n      cleanup\n    )\n  )\n);",
				},
				{
					name: "disabled-satisfies-cast-arrow",
					code: "promise.finally(\n  /** @satisfies {() => Promise<void>} */ (\n    (\n      // eslint-disable-next-line unicorn/no-async-promise-finally\n      async () => {}\n    )\n  )\n);",
				},
			} {
				t.Run(testCase.name, func(t *testing.T) {
					diagnostics := lintNoAsyncPromiseFinally(t, testCase.code, mode.typeAware)
					if testCase.reportText == "" {
						if len(diagnostics) != 0 {
							t.Fatalf("disabled callback produced %d diagnostics, want 0", len(diagnostics))
						}
						return
					}
					if len(diagnostics) != 1 {
						t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
					}
					diagnostic := diagnostics[0]
					if diagnostic.Message.Id != "no-async-promise-finally" {
						t.Fatalf("message id = %q", diagnostic.Message.Id)
					}
					if got := testCase.code[diagnostic.Range.Pos():diagnostic.Range.End()]; got != testCase.reportText {
						t.Errorf("reported text = %q, want %q", got, testCase.reportText)
					}
				})
			}
		})
	}
}

func lintNoAsyncPromiseFinally(t *testing.T, code string, typeAware bool) []rule.RuleDiagnostic {
	t.Helper()
	var program *lintprogram.Program
	var fileName string
	if typeAware {
		root := fixtures.GetRootDir()
		fileName = tspath.ResolvePath(root.Dir, "case.js")
		fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
		compiled, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", utils.CreateCompilerHost(root.Dir, fs))
		if err != nil {
			t.Fatalf("create type-aware program: %v", err)
		}
		program = lintprogram.NewFromCompiler(compiled)
	} else {
		dir := tspath.NormalizePath(t.TempDir())
		fileName = tspath.ResolvePath(dir, "case.js")
		fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{fileName: code})
		var err error
		program, err = lintprogram.NewFromRoots(lintprogram.RootOptions{
			RootFileNames:   []string{fileName},
			Host:            utils.CreateCompilerHost(dir, fs),
			CompilerOptions: &core.CompilerOptions{Target: core.ScriptTargetESNext},
			SingleThreaded:  true,
		})
		if err != nil {
			t.Fatalf("create project:false program: %v", err)
		}
	}
	if got := program.CanProvideTypeChecker(program.GetSourceFile(fileName)); got != typeAware {
		t.Fatalf("CanProvideTypeChecker = %v, want %v", got, typeAware)
	}

	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     program,
		File:        fileName,
		HasTypeInfo: typeAware,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     no_async_promise_finally.NoAsyncPromiseFinallyRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if got := ctx.TypeChecker != nil; got != typeAware {
						t.Fatalf("TypeChecker available = %v, want %v", got, typeAware)
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
