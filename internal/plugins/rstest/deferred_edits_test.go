package rstest_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_focused_tests"
	noImportNode "github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_import_node_test"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_called_exactly_once_with"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_equality_matcher"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_importing_rstest_globals"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_expect"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_title"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type deferredEditCase struct {
	name        string
	rule        rule.Rule
	code        string
	options     []any
	messages    []string
	output      string
	suggestions map[int][]string
}

func deferredEditCases() []deferredEditCase {
	return []deferredEditCase{
		{
			name:        "focused alias deduplication after suppression",
			rule:        no_focused_tests.NoFocusedTestsRule,
			code:        "const focused = test.only;\n// rslint-disable-next-line rstest/no-focused-tests\nfocused('hidden', fn);\nfocused('first', fn);\nfocused('second', fn);",
			messages:    []string{"focusedTest", "focusedTest"},
			suggestions: map[int][]string{0: {"const focused = test;\n// rslint-disable-next-line rstest/no-focused-tests\nfocused('hidden', fn);\nfocused('first', fn);\nfocused('second', fn);"}},
		},
		{
			name:        "focused optional computed accessor with comments",
			rule:        no_focused_tests.NoFocusedTestsRule,
			code:        `test?.[/* keep */ "only"]('case', fn);`,
			messages:    []string{"focusedTest"},
			suggestions: map[int][]string{0: {`test?./* keep */ ('case', fn);`}},
		},
		{
			name:     "async insertion after suppressed descriptor",
			rule:     valid_expect.ValidExpectRule,
			code:     "test('case', () => {\n// rslint-disable-next-line rstest/valid-expect\nexpect(p).resolves.toBe(0);\nexpect(p).resolves.toBe(1);\nexpect(q).rejects.toThrow();\n});",
			messages: []string{"asyncMustBeAwaited", "asyncMustBeAwaited"},
			output:   "test('case', async () => {\n// rslint-disable-next-line rstest/valid-expect\nexpect(p).resolves.toBe(0);\nawait expect(p).resolves.toBe(1);\nawait expect(q).rejects.toThrow();\n});",
		},
		{
			name:     "async return and nested functions",
			rule:     valid_expect.ValidExpectRule,
			code:     `test('case', () => { function inner() { return (expect(p).resolves.toBe(1)); } return expect(q).rejects.toThrow(); });`,
			options:  []any{map[string]any{"alwaysAwait": true}},
			messages: []string{"asyncMustBeAwaited", "asyncMustBeAwaited"},
			output:   `test('case', async () => { async function inner() { await (expect(p).resolves.toBe(1)); } await expect(q).rejects.toThrow(); });`,
		},
		{
			name:     "top level async assertion remains report only",
			rule:     valid_expect.ValidExpectRule,
			code:     `expect(p).resolves.toBe(1);`,
			messages: []string{"asyncMustBeAwaited"},
		},
		{
			name:     "late rstack import selects replacement for every import",
			rule:     noImportNode.NoImportNodeTestRule,
			code:     "import { test as a } from 'node:test';\nimport { it as b } from \"node:test\";\nimport { expect } from 'rstack/test';",
			messages: []string{"noImportNodeTest", "noImportNodeTest"},
			output:   "import { test as a } from 'rstack/test';\nimport { it as b } from \"rstack/test\";\nimport { expect } from 'rstack/test';",
		},
		{
			name:     "core reference wins and unsafe imports remain report only",
			rule:     noImportNode.NoImportNodeTestRule,
			code:     "import { test } from 'node:test';\nimport { mock } from 'node:test';\nimport reporters from 'node:test/reporters';\nimport { expect } from 'rstack/test';\nconst api = require('@rstest/core');",
			messages: []string{"noImportNodeTest", "noImportNodeTest", "noImportNodeTest"},
			output:   "import { test } from '@rstest/core';\nimport { mock } from 'node:test';\nimport reporters from 'node:test/reporters';\nimport { expect } from 'rstack/test';\nconst api = require('@rstest/core');",
		},
		{
			name:     "title overlapping fixes converge",
			rule:     valid_title.ValidTitleRule,
			code:     `test('test title ', fn);`,
			messages: []string{"accidentalSpace", "duplicatePrefix"},
			output:   `test('title', fn);`,
		},
		{
			name:     "escaped title space remains report only",
			rule:     valid_title.ValidTitleRule,
			code:     `test('\u0020title', fn);`,
			messages: []string{"accidentalSpace"},
		},
		{
			name:     "matcher merge preserves computed access type arguments and comments",
			rule:     prefer_called_exactly_once_with.PreferCalledExactlyOnceWithRule,
			code:     "expect(spy).toHaveBeenCalledOnce();\nexpect(spy)[/* keep */ 'toHaveBeenCalledWith']<string>('value');",
			messages: []string{"preferCalledExactlyOnceWith"},
			output:   `expect(spy)[/* keep */ 'toHaveBeenCalledExactlyOnceWith']<string>('value');`,
		},
		{
			name:     "unstable second expect argument forbids merge fix",
			rule:     prefer_called_exactly_once_with.PreferCalledExactlyOnceWithRule,
			code:     "expect(spy, message()).toHaveBeenCalledOnce();\nexpect(spy, message()).toHaveBeenCalledWith('value');",
			messages: []string{"preferCalledExactlyOnceWith"},
		},
		{
			name:     "equality suggestion truth table and comma operands",
			rule:     prefer_equality_matcher.PreferEqualityMatcherRule,
			code:     `expect((a(), b) !== c).not.toBe(false);`,
			messages: []string{"useEqualityMatcher"},
			suggestions: map[int][]string{0: {
				`expect((a(), b)).not.toBe(c);`,
				`expect((a(), b)).not.toEqual(c);`,
				`expect((a(), b)).not.toStrictEqual(c);`,
			}},
		},
		{
			name:     "global import diagnostic order differs from sorted fix",
			rule:     prefer_importing_rstest_globals.PreferImportingRstestGlobalsRule,
			code:     `test('case', () => { expect(rs.fn()).toBeDefined(); });`,
			messages: []string{"preferImportingRstestGlobals"},
			output:   "import { expect, rs, test } from '@rstest/core';\ntest('case', () => { expect(rs.fn()).toBeDefined(); });",
		},
		{
			name:     "global write forbids import fix",
			rule:     prefer_importing_rstest_globals.PreferImportingRstestGlobalsRule,
			code:     `expect(1).toBe(1); expect = other;`,
			messages: []string{"preferImportingRstestGlobals"},
		},
	}
}

func deferredEditProgram(t *testing.T, code string, typed bool) (*lintprogram.Program, *ast.SourceFile) {
	t.Helper()
	root := fixtures.GetRootDir()
	if typed {
		program, sourceFile, err := rule_tester.NewProgramHelper(root).CreateTestProgram(code, "deferred-edits.ts", "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		return lintprogram.NewFromCompiler(program), sourceFile
	}
	fileName := tspath.ResolvePath(root.Dir, "deferred-edits.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames: []string{fileName}, Host: utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil || program.CanProvideTypeChecker(sourceFile) {
		t.Fatal("expected source-only program")
	}
	return program, sourceFile
}

func lintDeferredEdits(t *testing.T, test deferredEditCase, program *lintprogram.Program, sourceFile *ast.SourceFile, typed bool, demand rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()
	var diagnostics []rule.RuleDiagnostic
	options := rule_tester.ResolveTestCaseOptions(t, &test.rule, test.options)
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program, File: sourceFile.FileName(), HasTypeInfo: typed,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{Name: test.rule.Name, Severity: rule.SeverityWarning,
				Run: func(ctx rule.RuleContext) rule.RuleListeners { return test.rule.Run(ctx, options) },
			}}
		},
		Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
	})
	return diagnostics
}

func TestRstestDeferredEditDemand(t *testing.T) {
	for _, test := range deferredEditCases() {
		for _, typed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/typed=%t", test.name, typed), func(t *testing.T) {
				program, sourceFile := deferredEditProgram(t, test.code, typed)
				all := lintDeferredEdits(t, test, program, sourceFile, typed, rule.EditDemandAll)
				if len(all) != len(test.messages) {
					t.Fatalf("diagnostics = %d, want %d: %#v", len(all), len(test.messages), all)
				}
				for index, diagnostic := range all {
					if diagnostic.Message.Id != test.messages[index] {
						t.Fatalf("diagnostic %d = %s, want %s", index, diagnostic.Message.Id, test.messages[index])
					}
					want := test.suggestions[index]
					var suggestions []rule.RuleSuggestion
					if diagnostic.Suggestions != nil {
						suggestions = *diagnostic.Suggestions
					}
					if len(suggestions) != len(want) {
						t.Fatalf("diagnostic %d suggestions = %d, want %d", index, len(suggestions), len(want))
					}
					for i, suggestion := range suggestions {
						output, _, applied := linter.ApplyRuleFixes(test.code, []rule.RuleSuggestion{suggestion})
						if !applied || output != want[i] {
							t.Fatalf("suggestion %d output:\n%s\nwant:\n%s", i, output, want[i])
						}
						assertDeferredOutputParses(t, output)
					}
				}
				for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
					got := lintDeferredEdits(t, test, program, sourceFile, typed, demand)
					if len(got) != len(all) {
						t.Fatalf("demand %d changed diagnostic count", demand)
					}
					for index, expected := range all {
						if demand&rule.EditDemandAutofix == 0 {
							expected.FixesPtr = nil
						}
						if demand&rule.EditDemandSuggestion == 0 {
							expected.Suggestions = nil
						}
						if !reflect.DeepEqual(got[index], expected) {
							t.Fatalf("demand %d changed diagnostic %d:\ngot %#v\nwant %#v", demand, index, got[index], expected)
						}
					}
				}
				code := test.code
				for pass := range 5 {
					diagnostics := all
					if pass > 0 {
						program, sourceFile = deferredEditProgram(t, code, typed)
						diagnostics = lintDeferredEdits(t, test, program, sourceFile, typed, rule.EditDemandAutofix)
					}
					output, _, applied := linter.ApplyRuleFixes(code, diagnostics)
					if !applied {
						break
					}
					if pass == 4 {
						t.Fatal("autofix did not converge")
					}
					assertDeferredOutputParses(t, output)
					code = output
				}
				want := test.output
				if want == "" {
					want = test.code
				}
				if code != want {
					t.Fatalf("autofix output:\n%s\nwant:\n%s", code, want)
				}
				program, sourceFile = deferredEditProgram(t, "/* rslint-disable "+test.rule.Name+" */\n"+test.code, typed)
				if got := lintDeferredEdits(t, test, program, sourceFile, typed, rule.EditDemandAll); len(got) != 0 {
					t.Fatalf("suppressed diagnostics = %d", len(got))
				}
			})
		}
	}
}

func assertDeferredOutputParses(t *testing.T, code string) {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/fixed.ts", Path: "/fixed.ts"}, code, core.ScriptKindTS)
	if len(sourceFile.Diagnostics()) != 0 {
		t.Fatalf("fix produced parse diagnostics: %s", code)
	}
}

func BenchmarkRstestDeferredNodeImports(b *testing.B) {
	for _, count := range []int{500, 1000, 2000} {
		var source strings.Builder
		for i := range count {
			fmt.Fprintf(&source, "import { test as test%d } from 'node:test';\n", i)
		}
		source.WriteString("import { expect } from 'rstack/test';\n")
		sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/bench.ts", Path: "/bench.ts"}, source.String(), core.ScriptKindTS)
		for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix} {
			b.Run(fmt.Sprintf("imports=%d/demand=%d", count, demand), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					ctx := rule.RuleContext{SourceFile: sourceFile}.WithDiagnosticConsumer(noImportNode.NoImportNodeTestRule.Name, rule.SeverityError, rule.DiagnosticConsumer{Demand: demand, Report: func(rule.RuleDiagnostic) {}})
					listener := noImportNode.NoImportNodeTestRule.Run(ctx, nil)[ast.KindImportDeclaration]
					for _, statement := range sourceFile.Statements.Nodes {
						listener(statement)
					}
				}
			})
		}
	}
}
