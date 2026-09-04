package prefer_rs_mocked

import (
	"reflect"
	"sort"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/consistent_rstest_namespace"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferRsMocked(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferRsMockedRule,
		[]rule_tester.ValidTestCase{
			{Code: `import type { Mock } from './test-helpers';
(getUser as Mock).mockReturnValue(user);`},
			{Code: `type Mock = { mockReturnValue(value: unknown): void };
(getUser as Mock).mockReturnValue(user);`},
			{Code: `(getUser as Mock).mockReturnValue(user);`},
			{Code: `import type { MockedFunction, MockedClass, MockedObject } from '@rstest/core';
(getUser as MockedFunction).mock;
(User as MockedClass).mock;
(service as MockedObject).mock;`},
			{Code: `import type { Mock } from '@rstest/core';
getUser satisfies Mock;`},
			{Code: `import type { Mock } from '@rstest/core';
const getUser: Mock = createMock();`},
			{Code: `import * as core from '@rstest/core';
(getUser as core.Mock).mockReturnValue(user);`},
			{Code: `import type { Mock } from '@rstest/core';
rs.mocked(getUser).mockReturnValue(user);
rstest.mocked(getUser).mockReturnValue(user);
import.meta.rstest!.rs.mocked(getUser).mockReturnValue(user);`},
			{Code: `import type { Mock } from '@rstest/core';
(getUser as Mock as unknown);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `import { rs, type Mock } from '@rstest/core';
(getUser as Mock).mockReturnValue(user);`,
				Output: []string{`import { rs, type Mock } from '@rstest/core';
(rs.mocked(getUser)).mockReturnValue(user);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferRsMocked",
					Message:   "Prefer `rs.mocked()` over type assertions",
					Line:      2,
					Column:    2,
					EndLine:   2,
					EndColumn: 17,
				}},
			},
			{
				Code: `import { rstest, type Mocked } from '@rstest/core';
(mod as Mocked<typeof mod>).init();`,
				Output: []string{`import { rstest, type Mocked } from '@rstest/core';
(rstest.mocked(mod)).init();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferRsMocked",
					Line:      2,
					Column:    2,
					EndLine:   2,
					EndColumn: 27,
				}},
			},
			{
				Code: `import type { MockInstance } from '@rstest/core';
(spy as MockInstance).mockRestore();`,
				Output: []string{`import type { MockInstance } from '@rstest/core';
(rs.mocked(spy)).mockRestore();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 2, Column: 2}},
			},
			{
				Code: `import { type Mock } from '@rstest/core';
(service.getUser as Mock).mockReturnValue(user);`,
				Output: []string{`import { type Mock } from '@rstest/core';
(rs.mocked(service.getUser)).mockReturnValue(user);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 2, Column: 2}},
			},
			{
				Code: `import type { Mock } from 'rstack/test';
(getUser as unknown as Mock).mockReturnValue(user);`,
				Output: []string{`import type { Mock } from 'rstack/test';
(rs.mocked(getUser)).mockReturnValue(user);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 2, Column: 2}},
			},
			{
				Code: `import type { Mock as RstestMock } from '@rstest/core';
const mocked = (getUser as RstestMock);`,
				Output: []string{`import type { Mock as RstestMock } from '@rstest/core';
const mocked = (rs.mocked(getUser));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 2, Column: 17}},
			},
			{
				Code: `import { rs } from '@rstest/core';
import type { Mock } from '@rstest/core';
(<Mock>getUser).mockReturnValue(user);`,
				Output: []string{`import { rs } from '@rstest/core';
import type { Mock } from '@rstest/core';
(rs.mocked(getUser)).mockReturnValue(user);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 3, Column: 2}},
			},
			{
				// An aliased import is written under the name the file binds.
				Code: `import { rs as r, type Mock } from '@rstest/core';
(getUser as Mock).mockReturnValue(user);`,
				Output: []string{`import { rs as r, type Mock } from '@rstest/core';
(r.mocked(getUser)).mockReturnValue(user);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 2, Column: 2}},
			},
			{
				// An export named with a string literal binds the same
				// namespace as the identifier spelling.
				Code: `import { "rs" as r, type Mock } from '@rstest/core';
(getUser as Mock).mockReturnValue(user);`,
				Output: []string{`import { "rs" as r, type Mock } from '@rstest/core';
(r.mocked(getUser)).mockReturnValue(user);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 2, Column: 2}},
			},
			{
				Code: `import { "rstest" as rstest, type Mock } from '@rstest/core';
(getUser as Mock).mockReturnValue(user);`,
				Output: []string{`import { "rstest" as rstest, type Mock } from '@rstest/core';
(rstest.mocked(getUser)).mockReturnValue(user);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 2, Column: 2}},
			},
			{
				Code: `import { rstest as helper, type Mocked } from '@rstest/core';
(mod as Mocked<typeof mod>).init();`,
				Output: []string{`import { rstest as helper, type Mocked } from '@rstest/core';
(helper.mocked(mod)).init();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 2, Column: 2}},
			},
			{
				// A namespace destructured out of a `require` binds the same
				// namespace an ESM named import does.
				Code: `import type { Mock } from '@rstest/core';
const { rs } = require('@rstest/core');
(getUser as Mock).mockReturnValue(user);`,
				Output: []string{`import type { Mock } from '@rstest/core';
const { rs } = require('@rstest/core');
(rs.mocked(getUser)).mockReturnValue(user);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 3, Column: 2}},
			},
			{
				// A renamed `require` binding is written under the name it
				// binds, not under the property it reads.
				Code: `import type { Mock } from '@rstest/core';
const { rstest: helper } = require('rstack/test');
(getUser as Mock).mockReturnValue(user);`,
				Output: []string{`import type { Mock } from '@rstest/core';
const { rstest: helper } = require('rstack/test');
(helper.mocked(getUser)).mockReturnValue(user);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 3, Column: 2}},
			},
			{
				// A `require` inside a block binds the namespace just as a
				// top-level one does.
				Code: `import type { Mock } from '@rstest/core';
function run() {
  const { rstest: helper } = require('@rstest/core');
  (getUser as Mock).mockReturnValue(user);
}`,
				Output: []string{`import type { Mock } from '@rstest/core';
function run() {
  const { rstest: helper } = require('@rstest/core');
  (helper.mocked(getUser)).mockReturnValue(user);
}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 4, Column: 4}},
			},
			{
				// A computed property name spelled by a static string names the
				// namespace as plainly as an identifier does.
				Code: `import type { Mock } from '@rstest/core';
const { ["rs"]: helper } = require('@rstest/core');
(getUser as Mock).mockReturnValue(user);`,
				Output: []string{`import type { Mock } from '@rstest/core';
const { ["rs"]: helper } = require('@rstest/core');
(helper.mocked(getUser)).mockReturnValue(user);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 3, Column: 2}},
			},
			{
				// A computed property that is not a static string names no
				// namespace the rule can write, so no edit is offered.
				Code: `import type { Mock } from '@rstest/core';
const { [key]: rs } = require('@rstest/core');
(getUser as Mock).mockReturnValue(user);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 3, Column: 2}},
			},
			{
				// A `require` of an unrelated module binds nothing the rule may
				// reach for, and the name it captures blocks the edit.
				Code: `import type { Mock } from '@rstest/core';
const { rs } = require('./helpers');
(getUser as Mock).mockReturnValue(user);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 3, Column: 2}},
			},
			{
				// A binding that captures the alias leaves the report unfixed.
				Code: `import { rs as helper, type Mock } from '@rstest/core';
function run(helper: unknown) {
  (getUser as Mock).mockReturnValue(user);
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 3, Column: 4}},
			},
			{
				// `import type { rs }` binds no runtime namespace to call.
				Code: `import type { rs, Mock } from '@rstest/core';
(getUser as Mock).mockReturnValue(user);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 2, Column: 2}},
			},
			{
				Code: `import type { Mock } from '@rstest/core';
const rs = createUtilities();
(getUser as Mock).mockReturnValue(user);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferRsMocked", Line: 3, Column: 2}},
			},
			{
				Code: `import type { Mock } from '@rstest/core';
const rs = createUtilities();
const rstest = createRunner();
import.meta.rstest!.expect(1).toBe(1);
(getUser as Mock).mockReturnValue(user);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferRsMocked",
					Line:      5,
					Column:    2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "preferRsMocked",
						Output: `import type { Mock } from '@rstest/core';
const rs = createUtilities();
const rstest = createRunner();
import.meta.rstest!.expect(1).toBe(1);
(import.meta.rstest.rs.mocked(getUser)).mockReturnValue(user);`,
					}},
				}},
			},
		},
	)
}

func TestPreferRsMockedEditDemand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		code           string
		wantAutofix    bool
		wantSuggestion bool
	}{
		{
			name: "autofix",
			code: `import { rs, type Mock } from '@rstest/core';
(getUser as Mock);`,
			wantAutofix: true,
		},
		{
			name: "suggestion",
			code: `import type { Mock } from '@rstest/core';
const rs = localRs;
const rstest = localRstest;
import.meta.rstest!.expect(1).toBe(1);
(getUser as Mock);`,
			wantSuggestion: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(test.code, test.name+"-edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}

			run := func(demand rule.EditDemand) rule.RuleDiagnostic {
				t.Helper()
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program),
					File:    sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name:     PreferRsMockedRule.Name,
							Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return PreferRsMockedRule.Run(ctx, nil)
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
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
				}
				return diagnostics[0]
			}

			diagnostics := map[rule.EditDemand]rule.RuleDiagnostic{
				rule.EditDemandNone:       run(rule.EditDemandNone),
				rule.EditDemandAutofix:    run(rule.EditDemandAutofix),
				rule.EditDemandSuggestion: run(rule.EditDemandSuggestion),
				rule.EditDemandAll:        run(rule.EditDemandAll),
			}
			withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
				diagnostic.FixesPtr = nil
				diagnostic.Suggestions = nil
				return diagnostic
			}
			wantDiagnostic := withoutEdits(diagnostics[rule.EditDemandAll])
			for demand, diagnostic := range diagnostics {
				if got := withoutEdits(diagnostic); !reflect.DeepEqual(got, wantDiagnostic) {
					t.Errorf("demand %d changed the diagnostic:\ngot  %#v\nwant %#v", demand, got, wantDiagnostic)
				}
			}

			if diagnostics[rule.EditDemandNone].FixesPtr != nil || diagnostics[rule.EditDemandNone].Suggestions != nil {
				t.Fatal("diagnostics-only demand materialized edits")
			}
			if diagnostics[rule.EditDemandAutofix].Suggestions != nil {
				t.Fatal("autofix demand materialized suggestions")
			}
			if diagnostics[rule.EditDemandSuggestion].FixesPtr != nil {
				t.Fatal("suggestion demand materialized fixes")
			}
			if got := diagnostics[rule.EditDemandAutofix].FixesPtr != nil; got != test.wantAutofix {
				t.Errorf("autofix present = %v, want %v", got, test.wantAutofix)
			}
			if got := diagnostics[rule.EditDemandSuggestion].Suggestions != nil; got != test.wantSuggestion {
				t.Errorf("suggestion present = %v, want %v", got, test.wantSuggestion)
			}
		})
	}
}

func TestPreferRsMockedConvergesWithConsistentNamespace(t *testing.T) {
	t.Parallel()

	code := `import { rs, type Mock } from '@rstest/core';
(getUser as Mock).mockReturnValue(user);`
	want := `import { rstest, type Mock } from '@rstest/core';
(rstest.mocked(getUser)).mockReturnValue(user);`
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())

	for range 3 {
		program, sourceFile, err := helper.CreateTestProgram(code, "consistent-namespace.ts", "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program),
			File:    sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{
					{
						Name:     PreferRsMockedRule.Name,
						Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return PreferRsMockedRule.Run(ctx, nil)
						},
					},
					{
						Name:     consistent_rstest_namespace.ConsistentRstestNamespaceRule.Name,
						Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return consistent_rstest_namespace.ConsistentRstestNamespaceRule.Run(
								ctx,
								[]any{map[string]any{"fn": "rstest"}},
							)
						},
					},
				}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: rule.EditDemandAutofix,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		fixedCode, _, fixed := linter.ApplyRuleFixes(code, diagnostics)
		if !fixed {
			break
		}
		code = fixedCode
	}

	if code != want {
		t.Fatalf("combined fixes converged to:\n%s\nwant:\n%s", code, want)
	}
}

func TestPreferRsMockedResolvesInSourceOnlyProgram(t *testing.T) {
	if PreferRsMockedRule.RequiresTypeInfo {
		t.Fatal("rstest/prefer-rs-mocked must not require type information")
	}

	code := `
import { Mock } from '@rstest/core';
import type { Mocked } from '@rstest/core';
import { type MockInstance } from 'rstack/test';
import type { Mock as ForeignMock } from './test-helpers';

(first as Mock);
(second as Mocked<typeof second>);
(third as MockInstance);
(foreign as ForeignMock);
`
	want := []string{"first as Mock", "second as Mocked<typeof second>", "third as MockInstance"}

	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "prefer-rs-mocked-source-only.ts")
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
				Name:     PreferRsMockedRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return PreferRsMockedRule.Run(ctx, nil)
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
