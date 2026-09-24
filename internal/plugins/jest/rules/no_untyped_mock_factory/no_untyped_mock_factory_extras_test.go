// TestNoUntypedMockFactoryExtras covers added AST shapes, real usage, and
// upstream branch lock-ins. Migrated cases live in no_untyped_mock_factory_upstream_test.go.
package no_untyped_mock_factory

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
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
			// Locks in upstream parseJestFnCall top-of-chain gate.
			{Code: `consume(jest.mock('./service', () => ({})));`},
			{Code: `(jest.mock('./service', () => ({}))).value;`},
			// Locks in upstream CallExpression arm 1: non-member
			{Code: "mock('./service', () => ({}));"},
			// Locks in upstream CallExpression arm 2: arity
			{Code: "jest.mock();"},
			// Locks in upstream CallExpression arm 3: other API
			{Code: "jest.spyOn('./service', () => ({}));"},
			// Locks in upstream CallExpression arm 4: generic
			{Code: "jest.mock<any>('./service', () => ({}));"},
			// Locks in upstream CallExpression arm 5: parenthesized return annotation
			{Code: "jest.mock('./service', (((): object => ({}))));"},
			// Dimension 4: dynamic and numeric accessors
			{Code: "jest[method]('./service', () => ({})); jest[0]('./service', () => ({}));"},
			// Dimension 4: no call, tagged template and new expression
			{Code: "const factory = jest.mock; new jest.mock('./service', () => ({})); jest.mock`service`;"},
			// Real-user: #1313 virtual mock exemption
			{Code: "jest.mock('generated-runtime', () => ({ platform: 'test' }), { virtual: true });"},
			// Dimension 4: type wrapper is not an ESTree member
			{Code: "(jest.mock as any)('./service', () => ({}));"},
			// Local shadow
			{Code: "function setup(jest) { jest.mock('./service', () => ({})); }"},
			// Unrelated import
			{Code: "import { jest } from './helpers'; jest.mock('./service', () => ({}));"},
		}, []rule_tester.InvalidTestCase{
			// A CommonJS require binding cannot accept the type argument this
			// rule would otherwise insert.
			{Code: "const { jest } = require('@jest/globals'); jest.mock('./service', () => ({}));", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			// Locks in upstream static accessor handling and spread-factory branch.
			{Code: "jest[mock]('./service', () => ({}));", Output: []string{"jest[mock]<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "jest[`mock`]('./service', () => ({}));", Output: []string{"jest[`mock`]<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "jest.mock('./service', ...factories);", Output: []string{"jest.mock<typeof import('./service')>('./service', ...factories);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "jest.mock('./service', (() => ({})) as Factory);", Output: []string{"jest.mock<typeof import('./service')>('./service', (() => ({})) as Factory);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			// Only the outer mock call is part of the upstream rule's contract.
			{Code: "jest.mock('./service', () => ({})).mock('./next', () => ({}));", Output: []string{"jest.mock('./service', () => ({})).mock<typeof import('./next')>('./next', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},

			// Dimension 4: parenthesized callee
			{Code: "(jest.mock)('./service', () => ({}));", Output: []string{"(jest.mock)<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: nested receiver parentheses
			{Code: "((jest)).mock('./service', () => ({}));", Output: []string{"((jest)).mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: comments and escaped module text
			{Code: "jest.mock /* keep */ ('./ser\\x76ice', /* factory */ () => ({}));", Output: []string{"jest.mock<typeof import('./ser\\x76ice')> /* keep */ ('./ser\\x76ice', /* factory */ () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./ser\\x76ice')`"}}},
			// Dimension 4: function expression
			{Code: "jest.doMock('./service', function factory() { return {}; });", Output: []string{"jest.doMock<typeof import('./service')>('./service', function factory() { return {}; });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: async and generator
			{Code: "jest.mock('./service', async function* () { yield 1; });", Output: []string{"jest.mock<typeof import('./service')>('./service', async function* () { yield 1; });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`", Line: 1, Column: 1, EndLine: 1, EndColumn: 56}}},
			// Dimension 4: multiline diagnostic
			{Code: "jest.mock(\n  './service',\n  () => ({}),\n);", Output: []string{"jest.mock<typeof import('./service')>(\n  './service',\n  () => ({}),\n);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`", Line: 1, Column: 1, EndLine: 4, EndColumn: 2}}},
			// Dimension 4: quoted path and Unicode before call
			{Code: "/* 用户 */ jest.mock(\"./service\", () => ({}));", Output: []string{"/* 用户 */ jest.mock<typeof import(\"./service\")>(\"./service\", () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import(\"./service\")`"}}},
			// Dimension 4: parenthesized path
			{Code: "jest.mock((('./service')), () => ({}));", Output: []string{"jest.mock<typeof import('./service')>((('./service')), () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`", Line: 1, Column: 1, EndLine: 1, EndColumn: 39}}},
			// Locks in upstream findModuleName fallback
			{Code: "jest.mock(modulePath, () => ({}));", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import(./module-name)`", Line: 1, Column: 1, EndLine: 1, EndColumn: 34}}},
			// Real-user: #1313 partial module replacement
			{Code: "jest.mock('./service', () => ({ ...actualService, save: jest.fn() }));", Output: []string{"jest.mock<typeof import('./service')>('./service', () => ({ ...actualService, save: jest.fn() }));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`", Line: 1, Column: 1, EndLine: 1, EndColumn: 70}}},
			// Real-user: reusable factory
			{Code: "const makeService = () => ({ save: () => 1 }); jest.doMock('./service', makeService);", Output: []string{"const makeService = () => ({ save: () => 1 }); jest.doMock<typeof import('./service')>('./service', makeService);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: nested class method
			{Code: "class Setup { install() { jest.doMock('./service', () => ({})); } }", Output: []string{"class Setup { install() { jest.doMock<typeof import('./service')>('./service', () => ({})); } }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: accessor and optional call
			{Code: "jest[\"mock\"]('./service', () => ({}));", Output: []string{"jest[\"mock\"]<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: accessor and optional call
			{Code: "jest?.mock('./service', () => ({}));", Output: []string{"jest?.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: accessor and optional call
			{Code: "jest.mock?.('./service', () => ({}));", Output: []string{"jest.mock?.<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Dimension 4: accessor and optional call
			{Code: "jest[\"doMock\"]?.('./service', () => ({}));", Output: []string{"jest[\"doMock\"]?.<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Imported Jest alias
			{Code: "import { jest as j } from '@jest/globals'; j.mock('./service', () => ({}));", Output: []string{"import { jest as j } from '@jest/globals'; j.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Locks in upstream factory non-function
			{Code: "jest.mock('./service', null);", Output: []string{"jest.mock<typeof import('./service')>('./service', null);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`", Line: 1, Column: 1, EndLine: 1, EndColumn: 29}}},
			// Dimension 4: template module name
			{Code: "jest.mock(`./service`, () => ({}));", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import(./module-name)`", Line: 1, Column: 1, EndLine: 1, EndColumn: 35}}},
			// Dimension 4: type assertion on path
			{Code: "jest.mock('./service' as string, () => ({}));", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import(./module-name)`", Line: 1, Column: 1, EndLine: 1, EndColumn: 45}}},
		})
}

// TestNoUntypedMockFactoryEditDemand also exercises the binder-only backend.
func TestNoUntypedMockFactoryEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	compiler, file, err := helper.CreateTestProgram(`jest.mock('./service', () => ({}));
const factory = () => ({});
jest.doMock('./service', factory);
jest.mock(modulePath, () => ({}));`, "edit-demand.ts", "tsconfig.json")
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
