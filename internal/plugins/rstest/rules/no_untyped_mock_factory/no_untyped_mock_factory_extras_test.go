// TestNoUntypedMockFactoryExtras covers added AST shapes, real usage, and
// upstream branch lock-ins. Migrated cases live in no_untyped_mock_factory_upstream_test.go.
package no_untyped_mock_factory

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"reflect"
	"testing"
)

func TestNoUntypedMockFactoryFixedOutputTypeChecks(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	testCases := []struct {
		name     string
		code     string
		want2347 bool
	}{
		{
			name: "generic callee accepts fix",
			code: `declare const rs: { mock<T>(path: string, factory: () => Partial<T>): void };
rs.mock<typeof import('./async-mock-factories')>('./async-mock-factories', () => ({}));`,
		},
		{
			name: "any callee rejects fix",
			code: `declare const rs: any;
rs.mock<typeof import('./async-mock-factories')>('./async-mock-factories', () => ({}));`,
			want2347: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			program, file, err := helper.CreateTestProgram(testCase.code, "fixed-output.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			found2347 := false
			for _, diagnostic := range program.GetSemanticDiagnostics(context.Background(), file) {
				if diagnostic.Code() == 2347 {
					found2347 = true
				}
			}
			if found2347 != testCase.want2347 {
				t.Fatalf("TS2347 present = %v, want %v", found2347, testCase.want2347)
			}
		})
	}
}

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
			// A non-hoisted API reads a var before its initializer has run, so
			// the value passed here is undefined rather than a factory.
			{Code: "rs.doMockRequire('./service', factory); var factory = () => ({});"},
			// Hoisted APIs run before ordinary variable initializers, whatever
			// declaration keyword was used and wherever the declaration appears.
			{Code: "var factory = () => ({}); rs.mock('./service', factory);"},
			{Code: "const factory = () => ({}); rs.mockRequire('./service', factory);"},
			// A conditional var initializer may not run before the later call.
			{Code: "if (enabled) { var factory = () => ({}); } rs.doMock('./service', factory);"},
		}, []rule_tester.InvalidTestCase{
			// Function declarations are available to both ordinary and hoisted calls.
			{Code: "function factory() { return {}; } rs.mock('./service', factory);", Output: []string{"function factory() { return {}; } rs.mock<typeof import('./service')>('./service', factory);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			// rs.hoisted is the variable-initializer form that deliberately runs
			// before a hoisted module mock.
			{Code: "const factory = rs.hoisted(() => () => ({})); rs.mock('./service', factory);", Output: []string{"const factory = rs.hoisted(() => () => ({})); rs.mock<typeof import('./service')>('./service', factory);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			// A dominating declaration also covers calls in a later nested statement.
			{Code: "let factory = () => ({}); if (enabled) { rs.doMock('./service', factory); }", Output: []string{"let factory = () => ({}); if (enabled) { rs.doMock<typeof import('./service')>('./service', factory); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
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
			{Code: "(rs as any).mock('./service', () => ({}));", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest CommonJS, namespace and type wrappers
			{Code: "rs!.mock('./service', () => ({}));", Output: []string{"rs!.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest CommonJS, namespace and type wrappers
			{Code: "(rs.mock as any)('./service', () => ({}));", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest literal receiver transformation
			{Code: "import { rs } from 'rstack/test'; rs.mock('./service', () => ({}));", Output: []string{"import { rs } from 'rstack/test'; rs.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest literal receiver transformation
			{Code: "const { rs } = require('@rstest/core'); rs.mock('./service', () => ({}));", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest literal receiver transformation
			{Code: "import { rs } from '@rstest/core'; rs.mock('./service', () => ({}));", Output: []string{"import { rs } from '@rstest/core'; rs.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// Rstest literal receiver transformation
			{Code: "const rs = { mock() {} }; rs.mock('./service', () => ({}));", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock", Message: "Add a type parameter to the mock factory such as `typeof import('./service')`"}}},
			// A locally typed generic callee accepts the inserted type argument.
			{Code: "declare const rs: { mock<T>(path: string, factory: () => T): void }; rs.mock('./service', () => ({}));", Output: []string{"declare const rs: { mock<T>(path: string, factory: () => T): void }; rs.mock<typeof import('./service')>('./service', () => ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
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
	// Exercise stable const, let, and var factories in both backends. The
	// reassigned let binding must remain exempt even though its initializer is
	// a function, because the value passed to doMock is no longer that factory.
	compiler, file, err := helper.CreateTestProgram(`rs.mock('./service', () => ({}));
const factory = () => ({});
rs.doMock('./service', factory);
let mutableFactory = () => ({});
rs.doMock('./mutable-service', mutableFactory);
var legacyFactory = function () { return {}; };
rs.doMockRequire('./legacy-service', legacyFactory);
function declaredFactory() { return {}; }
rs.mock('./declared-service', declaredFactory);
const liftedFactory = rs.hoisted(() => () => ({}));
rs.mock('./lifted-service', liftedFactory);
let reassignedFactory = () => ({});
reassignedFactory = { spy: true };
rs.doMock('./reassigned-service', reassignedFactory);
rs.doMockRequire('./early-service', earlyFactory);
var earlyFactory = () => ({});
var ordinaryFactory = () => ({});
rs.mock('./hoisted-service', ordinaryFactory);
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
			if len(diagnostics) != 7 {
				t.Fatalf("typed=%v demand=%d: got %d diagnostics", typed, demand, len(diagnostics))
			}
			for i := range diagnostics {
				d := &diagnostics[i]
				wantFix := i < 6 && demand&rule.EditDemandAutofix != 0
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

func TestNoUntypedMockFactoryDeclarationTiming(t *testing.T) {
	testCases := []struct {
		name           string
		code           string
		sourceOnlyWant int
		typedWant      int
	}{
		{name: "non-hoisted stable let", code: `let factory = () => ({}); rs.doMock('./service', factory);`, sourceOnlyWant: 1, typedWant: 1},
		{name: "non-hoisted stable var", code: `var factory = () => ({}); rs.doMockRequire('./service', factory);`, sourceOnlyWant: 1, typedWant: 1},
		{name: "non-hoisted use before var", code: `rs.doMockRequire('./service', factory); var factory = () => ({});`},
		{name: "non-hoisted use before let", code: `rs.doMock('./service', factory); let factory = () => ({});`},
		{name: "non-hoisted use before const", code: `rs.doMock('./service', factory); const factory = () => ({});`},
		{name: "outer declaration dominates nested block", code: `let factory = () => ({}); if (enabled) { rs.doMock('./service', factory); }`, sourceOnlyWant: 1, typedWant: 1},
		{name: "conditional initializer does not dominate", code: `if (enabled) { var factory = () => ({}); } rs.doMock('./service', factory);`},
		{name: "switch case initializer does not dominate", code: `switch (kind) { case 0: var factory = () => ({}); break; default: rs.doMock('./service', factory); }`},
		{name: "for initializer is conservatively skipped", code: `for (let factory = () => ({}); enabled; ) { rs.doMock('./service', factory); break; }`},
		{name: "cross-function timing is conservatively skipped", code: `const factory = () => ({}); function install() { rs.doMock('./service', factory); }`},
		{name: "direct reassignment", code: `let factory = () => ({}); factory = other; rs.doMock('./service', factory);`},
		{name: "logical reassignment", code: `let factory = () => ({}); factory ||= other; rs.doMock('./service', factory);`},
		{name: "destructuring reassignment", code: `let factory = () => ({}); ({ factory } = other); rs.doMock('./service', factory);`},
		{name: "loop reassignment", code: `let factory = () => ({}); for (factory of factories) {} rs.doMock('./service', factory);`},
		{name: "closure reassignment", code: `let factory = () => ({}); function replace() { factory = other; } rs.doMock('./service', factory);`},
		{name: "multiple var initializers are conservatively skipped", code: `var factory = () => ({}); var factory = { spy: true }; rs.doMock('./service', factory);`},
		{name: "callable result after initialization", code: `declare function makeFactory(): () => object; const factory = makeFactory(); rs.doMock('./service', factory);`, typedWant: 1},
		{name: "non-hoisted imported factory uses type information", code: `import { syncModuleFactory } from './async-mock-factories'; rs.doMock('./service', syncModuleFactory);`, typedWant: 1},
		{name: "non-hoisted member factory uses type information", code: `const holder = { factory: () => ({}) }; rs.doMock('./service', holder.factory);`, typedWant: 1},
		{name: "hoisted mock rejects ordinary var", code: `var factory = () => ({}); rs.mock('./service', factory);`},
		{name: "hoisted mock rejects ordinary const", code: `const factory = () => ({}); rs.mockRequire('./service', factory);`},
		{name: "hoisted mock rejects nested function declaration", code: `function install() { function factory() { return {}; } rs.mock('./service', factory); }`},
		{name: "hoisted mock rejects ambient function", code: `declare function factory(): object; rs.mock('./service', factory);`},
		{name: "hoisted mock rejects callable parameter", code: `function install(factory: () => object) { rs.mock('./service', factory); }`},
		{name: "hoisted mock rejects imported factory", code: `import { syncModuleFactory } from './async-mock-factories'; rs.mock('./service', syncModuleFactory);`},
		{name: "hoisted mock rejects local member", code: `const holder = { factory: () => ({}) }; rs.mock('./service', holder.factory);`},
		{name: "hoisted mock accepts top-level function declaration", code: `function factory() { return {}; } rs.mock('./service', factory);`, sourceOnlyWant: 1, typedWant: 1},
		{name: "hoisted mock accepts top-level overload implementation", code: `function factory(): object; function factory() { return {}; } rs.mock('./service', factory);`, sourceOnlyWant: 1, typedWant: 1},
		{name: "hoisted mock accepts direct hoisted binding", code: `const factory = rs.hoisted(() => () => ({})); rs.mock('./service', factory);`, sourceOnlyWant: 1, typedWant: 1},
		{name: "hoisted mock accepts wrapped rstest hoisted binding", code: `const factory = (rstest as any).hoisted(() => () => ({})); rs.mockRequire('./service', factory);`, sourceOnlyWant: 1, typedWant: 1},
		{name: "hoisted mock accepts object destructuring", code: `declare const rs: { hoisted<T>(callback: () => T): T }; const { factory } = rs.hoisted(() => ({ factory: () => ({}) })); rs.mock('./service', factory);`, sourceOnlyWant: 1, typedWant: 1},
		{name: "hoisted mock accepts renamed object destructuring", code: `declare const rs: { hoisted<T>(callback: () => T): T }; const { factory: makeFactory } = rs.hoisted(() => ({ factory: () => ({}) })); rs.mock('./service', makeFactory);`, sourceOnlyWant: 1, typedWant: 1},
		{name: "hoisted mock accepts array destructuring", code: `declare const rs: { hoisted<T>(callback: () => T): T }; const [factory] = rs.hoisted(() => [() => ({})]); rs.mock('./service', factory);`, sourceOnlyWant: 1, typedWant: 1},
		{name: "hoisted object spread is conservatively skipped", code: `declare const rs: { hoisted<T>(callback: () => T): T }; declare const other: { factory: { spy: true } }; const { factory } = rs.hoisted(() => ({ factory: () => ({}), ...other })); rs.mock('./service', factory);`},
		{name: "nested hoisted destructuring does not match a top-level decoy", code: `declare const rs: { hoisted<T>(callback: () => T): T }; const { nested: { factory } } = rs.hoisted(() => ({ factory: () => ({}), nested: { factory: { spy: true as const } } })); rs.mock('./service', factory);`},
		{name: "computed hoisted property may override the static key", code: `declare const rs: { hoisted<T>(callback: () => T): T }; declare const mockKey: 'factory'; const { factory } = rs.hoisted(() => ({ factory: () => ({}), [mockKey]: { spy: true as const } })); rs.mock('./service', factory);`},
		{name: "duplicate hoisted property may override the factory", code: `declare const rs: { hoisted<T>(callback: () => T): T }; const { factory } = rs.hoisted(() => ({ factory: () => ({}), factory: { spy: true as const } })); rs.mock('./service', factory);`},
		{name: "hoisted options remain exempt", code: `const options = rs.hoisted(() => ({ spy: true as const })); rs.mock('./service', options);`},
		{name: "hoisted callable options union remains exempt", code: `const value = rs.hoisted((): (() => object) | { spy: true } => condition ? () => ({}) : { spy: true }); rs.mock('./service', value);`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			for _, typed := range []bool{false, true} {
				want := testCase.sourceOnlyWant
				if typed {
					want = testCase.typedWant
				}
				if got := lintNoUntypedMockFactory(t, testCase.code, typed); got != want {
					t.Fatalf("typed=%v: diagnostics = %d, want %d", typed, got, want)
				}
			}
		})
	}
}

func lintNoUntypedMockFactory(t *testing.T, code string, typed bool) int {
	t.Helper()
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	compiler, file, err := helper.CreateTestProgram(code, "declaration-timing.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lintprogram.NewFromBoundSources(compiler, compiler.SourceFiles())
	if err != nil {
		t.Fatal(err)
	}
	if typed {
		program = lintprogram.NewFromCompiler(compiler)
	}

	count := 0
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     program,
		File:        file.FileName(),
		HasTypeInfo: typed,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     NoUntypedMockFactoryRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoUntypedMockFactoryRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandNone, Report: func(rule.RuleDiagnostic) { count++ }},
	})
	return count
}
