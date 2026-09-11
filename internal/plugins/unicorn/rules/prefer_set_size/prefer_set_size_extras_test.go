package prefer_set_size_test

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
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_set_size"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferSetSizeExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_set_size.PreferSetSizeRule,
		[]rule_tester.ValidTestCase{
			// Static bracket access and optional access are distinct from upstream's
			// non-optional dot-member shape.
			{Code: "Array['from'](new Set(values)).length", FileName: "file.js"},
			{Code: "Array.from(new Set(values))?.length", FileName: "file.js"},
			{Code: "[...new Set(values)]?.length", FileName: "file.js"},
			// A TS assertion on the conversion itself is visible to upstream and does
			// not count as one of its two conversion shapes.
			{Code: "([...(set as Set<string>)] as string[]).length", FileName: "file.ts"},
			// Upstream's isMethodCall rejects spread arguments by default: the
			// source cannot replace the conversion as one member-expression object.
			{Code: "declare const args: [Set<number>]; Array.from(...args).length", FileName: "file.ts"},
		},
		[]rule_tester.InvalidTestCase{
			// Const aliases recurse, while `let` deliberately did not match in the
			// JavaScript upstream parser.
			invalid("const first = new Set(values); const second = first; [...second].length", "const first = new Set(values); const second = first; second.size", "file.js"),
			// A Set construction without constructor parentheses needs a wrapper once
			// it becomes the object of a member expression.
			invalid("[...new Set].length", "(new Set).size", "file.js"),
			invalid("[...new (Set)].length", "(new (Set)).size", "file.js"),
			// Rslint-specific runtime helper regressions; these are not part of
			// eslint-plugin-unicorn v74.0.0's prefer-set-size test corpus.
			invalid("[...(flag ? new Set() : new Set())].length", "(flag ? new Set() : new Set()).size", "file.js"),
			invalid("[...(sideEffect(), new Set())].length", "(sideEffect(), new Set()).size", "file.js"),
			invalid("function size(value: unknown) { return [...(new Set() satisfies Set)].length; }", "function size(value: unknown) { return (new Set() satisfies Set).size; }", "file.ts"),
			// Comments nested in Set survive; an outer conversion comment suppresses
			// the whole fix rather than dropping author text.
			invalid("[...new /* retained */ Set(values)].length", "new /* retained */ Set(values).size", "file.js"),
			{Code: "Array.from(/* outer */ new /* retained */ Set(values)).length", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{errorAtLength("Array.from(/* outer */ new /* retained */ Set(values)).length", 0)}},
			// Authored TypeScript wrappers remain visible in the replacement and are
			// parenthesized before they become a member-expression object.
			{Code: "function size(value: unknown) { return Array.from(value satisfies Set<string>).length; }", FileName: "file.ts"},
			// A project-backed non-null assertion narrows the nullable Set before
			// `isSet` unwraps it.
			invalid("declare const set: Set<string> | null; [...set!].length", "declare const set: Set<string> | null; (set!).size", "file.ts"),
		},
	)
}

func TestPreferSetSizeProjectFalseTypeSyntax(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		output string
	}{
		{
			name:   "parameter Set annotation",
			code:   "function getSize(set: Set<string>) { return Array.from(set).length; }",
			output: "function getSize(set: Set<string>) { return set.size; }",
		},
		{
			name:   "parameter ReadonlySet annotation",
			code:   "function getSize(set: ReadonlySet<string>) { return Array.from(set).length; }",
			output: "function getSize(set: ReadonlySet<string>) { return set.size; }",
		},
		{
			name:   "spread assertion",
			code:   "function getSize(set: unknown) { return [...(set as Set<string>)].length; }",
			output: "function getSize(set: unknown) { return (set as Set<string>).size; }",
		},
		{
			name:   "Array.from assertion",
			code:   "function getSize(set: unknown) { return Array.from(set as Set<string>).length; }",
			output: "function getSize(set: unknown) { return (set as Set<string>).size; }",
		},
		{
			name:   "type alias annotation",
			code:   "type SetAlias = Set<string>; function getSize(set: SetAlias) { return Array.from(set).length; }",
			output: "type SetAlias = Set<string>; function getSize(set: SetAlias) { return set.size; }",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diagnostics := lintPreferSetSizeProjectFalse(t, test.code)
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostic count = %d, want 1: %+v", len(diagnostics), diagnostics)
			}
			output, unapplied, fixed := linter.ApplyRuleFixes(test.code, diagnostics)
			if !fixed || len(unapplied) != 0 || output != test.output {
				t.Fatalf("source-only output = %q, fixed=%t, unapplied=%+v; want %q", output, fixed, unapplied, test.output)
			}
		})
	}
}

func lintPreferSetSizeProjectFalse(t *testing.T, code string) []rule.RuleDiagnostic {
	t.Helper()
	dir := tspath.NormalizePath(t.TempDir())
	fileName := tspath.NormalizePath(filepath.Join(dir, "file.ts"))
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

	diagnostics := []rule.RuleDiagnostic{}
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program,
		File:    fileName,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     prefer_set_size.PreferSetSizeRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Fatal("project:false fixture unexpectedly received a TypeChecker")
					}
					return prefer_set_size.PreferSetSizeRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAutofix,
			Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		},
	})
	return diagnostics
}
