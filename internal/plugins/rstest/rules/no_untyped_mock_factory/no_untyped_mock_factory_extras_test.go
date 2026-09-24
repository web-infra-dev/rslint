// TestNoUntypedMockFactoryExtras covers added AST shapes, real usage, and
// upstream branch lock-ins. Migrated cases live in no_untyped_mock_factory_upstream_test.go.
package no_untyped_mock_factory

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"reflect"
	"testing"
)

func TestNoUntypedMockFactoryExtras(t *testing.T) {
	// N/A: object/class key equivalence, Chai, test modifiers and TestContext
	// do not participate in module factory annotations.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUntypedMockFactoryRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream CallExpression arm 1: non-member
			{Code: "mock('./service', () => ({}));"},
			// Locks in upstream CallExpression arm 2: arity
			{Code: "rs.mock();"},
			// Locks in upstream CallExpression arm 3: other API
			{Code: "rs.spyOn('./service', () => ({}));"},
			// Locks in upstream CallExpression arm 4: generic
			{Code: "rs.mock<any>('./service', () => ({}));"},
			// Locks in upstream CallExpression arm 5: parenthesized return annotation
			{Code: "rs.mock('./service', (((): object => ({}))));"},
			// Dimension 4: dynamic and numeric accessors
			{Code: "rs[method]('./service', () => ({})); rs[0]('./service', () => ({}));"},
			// Dimension 4: no call, tagged template and new expression
			{Code: "const factory = rs.mock; new rs.mock('./service', () => ({})); rs.mock`service`;"},
			// Real-user: #1313 virtual mock exemption
			{Code: "rs.mock('generated-runtime', () => ({ platform: 'test' }), { virtual: true });"},
			// Dynamic import infers shape
			{Code: "rs.mock(import('./service'), () => ({}));"},
			// Wrapped dynamic import infers shape
			{Code: "rs.doMock((import('./service') as Promise<object>), () => ({}));"},
			// Inline spy options
			{Code: "rs.mock('./service', { spy: true });"},
			// Inline mock options
			{Code: "rs.mockRequire('./service', { mock: true });"},
			// Options reference
			{Code: "const options = { spy: true } as const; rs.doMock('./service', options);"},
			// Unknown argument
			{Code: "rs.mock('./service', factoryOrOptions);"},
			// Bodyless declaration and options parameter
			{Code: "declare const options: { spy: true }; rs.mock('./service', options);"},
			// Uninitialized declaration
			{Code: "let factory; rs.mock('./service', factory);"},
			// Reassignment to options
			{Code: "let factory = () => ({}); factory = { spy: true }; rs.doMock('./service', factory);"},
			// Dimension 4: spread argument
			{Code: "rs.mock('./service', ...factories);"},
			// Unsupported alias
			{Code: "import { rs as mocker } from '@rstest/core'; mocker.mock('./service', () => ({}));"},
			// Unsupported namespace
			{Code: "import * as core from '@rstest/core'; core.rs.mock('./service', () => ({}));"},
			// Unsupported import.meta namespace
			{Code: "import.meta.rstest.rs.mock('./service', () => ({}));"},
			// Unsupported computed member
			{Code: "rs['mock']('./service', () => ({}));"},
			// Unsupported optional receiver
			{Code: "rs?.mock('./service', () => ({}));"},
			// Unsupported optional call
			{Code: "rs.mock?.('./service', () => ({}));"},
			// Unsupported expression position
			{Code: "consume(rs.mock('./service', () => ({})));"},
			// Unsupported concise arrow position
			{Code: "const install = () => rs.mock('./service', () => ({}));"},
			// Type wrapper preserves annotated return
			{Code: "rs.mock('./service', (((): object => ({})) as Function));"},
			// No factory on CommonJS API
			{Code: "rs.mockRequire('./service'); rs.doMockRequire('./service');"},
		}, []rule_tester.InvalidTestCase{
			// Dimension 4: parenthesized callee
			{Code: "(rs.mock)('./service', () => ({}));", Output: []string{"(rs.mock)<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: nested receiver parentheses
			{Code: "((rs)).mock('./service', () => ({}));", Output: []string{"((rs)).mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: comments and escaped module text
			{Code: "rs.mock /* keep */ ('./ser\\x76ice', /* factory */ () => ({}));", Output: []string{"rs.mock<typeof import('./ser\\x76ice')> /* keep */ ('./ser\\x76ice', /* factory */ () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./ser\\x76ice')`"}}},
			// Dimension 4: function expression
			{Code: "rs.doMock('./service', function factory() { return {}; });", Output: []string{"rs.doMock<typeof import('./service')>('./service', function factory() { return {}; });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: async and generator
			{Code: "rs.mock('./service', async function* () { yield 1; });", Output: []string{"rs.mock<typeof import('./service')>('./service', async function* () { yield 1; });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`", Line: 1, Column: 1, EndLine: 1, EndColumn: 54}}},
			// Dimension 4: multiline diagnostic
			{Code: "rs.mock(\n  './service',\n  () => ({}),\n);", Output: []string{"rs.mock<typeof import('./service')>(\n  './service',\n  () => ({}),\n);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`", Line: 1, Column: 1, EndLine: 4, EndColumn: 2}}},
			// Dimension 4: quoted path and Unicode before call
			{Code: "/* 用户 */ rs.mock(\"./service\", () => ({}));", Output: []string{"/* 用户 */ rs.mock<typeof import(\"./service\")>(\"./service\", () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import(\"./service\")`"}}},
			// Dimension 4: parenthesized path
			{Code: "rs.mock((('./service')), () => ({}));", Output: []string{"rs.mock<typeof import('./service')>((('./service')), () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`", Line: 1, Column: 1, EndLine: 1, EndColumn: 37}}},
			// Locks in upstream findModuleName fallback
			{Code: "rs.mock(modulePath, () => ({}));", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import(./module-name)`", Line: 1, Column: 1, EndLine: 1, EndColumn: 32}}},
			// Real-user: #1313 partial module replacement
			{Code: "rs.mock('./service', () => ({ ...actualService, save: rs.fn() }));", Output: []string{"rs.mock<typeof import('./service')>('./service', () => ({ ...actualService, save: rs.fn() }));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`", Line: 1, Column: 1, EndLine: 1, EndColumn: 66}}},
			// Real-user: reusable factory
			{Code: "const makeService = () => ({ save: () => 1 }); rs.doMock('./service', makeService);", Output: []string{"const makeService = () => ({ save: () => 1 }); rs.doMock<typeof import('./service')>('./service', makeService);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: nested class method
			{Code: "class Setup { install() { rs.doMock('./service', () => ({})); } }", Output: []string{"class Setup { install() { rs.doMock<typeof import('./service')>('./service', () => ({})); } }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest CommonJS, namespace and type wrappers
			{Code: "rs.mockRequire('./service', () => ({}));", Output: []string{"rs.mockRequire<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest CommonJS, namespace and type wrappers
			{Code: "rstest.doMockRequire('./service', () => ({}));", Output: []string{"rstest.doMockRequire<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest CommonJS, namespace and type wrappers
			{Code: "rstest.mock('./service', () => ({}));", Output: []string{"rstest.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest CommonJS, namespace and type wrappers
			{Code: "(rs as any).mock('./service', () => ({}));", Output: []string{"(rs as any).mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest CommonJS, namespace and type wrappers
			{Code: "rs!.mock('./service', () => ({}));", Output: []string{"rs!.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest CommonJS, namespace and type wrappers
			{Code: "(rs.mock as any)('./service', () => ({}));", Output: []string{"(rs.mock as any)<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest literal receiver transformation
			{Code: "import { rs } from 'rstack/test'; rs.mock('./service', () => ({}));", Output: []string{"import { rs } from 'rstack/test'; rs.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest literal receiver transformation
			{Code: "const { rs } = require('@rstest/core'); rs.mock('./service', () => ({}));", Output: []string{"const { rs } = require('@rstest/core'); rs.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest literal receiver transformation
			{Code: "import { rs } from '@rstest/core'; rs.mock('./service', () => ({}));", Output: []string{"import { rs } from '@rstest/core'; rs.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest literal receiver transformation
			{Code: "const rs = { mock() {} }; rs.mock('./service', () => ({}));", Output: []string{"const rs = { mock() {} }; rs.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Declared factory
			{Code: "function factory() { return {}; } rs.doMock('./service', factory);", Output: []string{"function factory() { return {}; } rs.doMock<typeof import('./service')>('./service', factory);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Callable parameter with type information
			{Code: "function install(factory: () => object) { rs.doMock('./service', factory); }", Output: []string{"function install(factory: () => object) { rs.doMock<typeof import('./service')>('./service', factory); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Type assertion on path
			{Code: "rs.mock('./service' as string, () => ({}));", Output: []string{"rs.mock<typeof import('./service')>('./service' as string, () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`", Line: 1, Column: 1, EndLine: 1, EndColumn: 43}}},
		})
}

// TestNoUntypedMockFactoryEditDemand also exercises the binder-only backend.
func TestNoUntypedMockFactoryEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	compiler, file, err := helper.CreateTestProgram(`rs.mock('./service', () => ({}));
const factory = () => ({});
rs.doMock('./service', factory);
rs.mock(modulePath, () => ({}));`, "edit-demand.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	bound, err := lintprogram.NewFromBoundSources(compiler, compiler.SourceFiles())
	if err != nil {
		t.Fatal(err)
	}
	for _, typed := range []bool{false, true} {
		program := bound
		if typed {
			program = lintprogram.NewFromCompiler(compiler)
		}
		var baseline []rule.RuleDiagnostic
		for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
			var diagnostics []rule.RuleDiagnostic
			linter.LintSingleFile(linter.LintSingleFileOptions{
				Program: program, File: file.FileName(), HasTypeInfo: typed,
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{Name: NoUntypedMockFactoryRule.Name, Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							if !typed && ctx.TypeChecker != nil {
								t.Fatal("source-only path received a TypeChecker")
							}
							return NoUntypedMockFactoryRule.Run(ctx, nil)
						},
					}}
				},
				Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
			})
			if len(diagnostics) != 3 {
				t.Fatalf("typed=%v demand=%d: got %d diagnostics", typed, demand, len(diagnostics))
			}
			for i := range diagnostics {
				d := &diagnostics[i]
				wantFix := i < 2 && demand&rule.EditDemandAutofix != 0
				if (d.FixesPtr != nil && len(*d.FixesPtr) > 0) != wantFix {
					t.Fatalf("typed=%v demand=%d diagnostic=%d: wrong edits", typed, demand, i)
				}
				if d.Suggestions != nil {
					t.Fatal("unexpected suggestions")
				}
				d.FixesPtr = nil
			}
			if baseline == nil {
				baseline = diagnostics
			} else if !reflect.DeepEqual(baseline, diagnostics) {
				t.Fatal("diagnostics changed with edit demand")
			}
		}
	}
}
