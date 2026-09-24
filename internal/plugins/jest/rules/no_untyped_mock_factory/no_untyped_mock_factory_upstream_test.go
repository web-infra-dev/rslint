// TestNoUntypedMockFactoryUpstream migrates every eslint-plugin-jest v29.16.6
// case. Rstest reverses the bracket-call case because its transform ignores it.
// Additional shapes and branch coverage live in no_untyped_mock_factory_extras_test.go.
package no_untyped_mock_factory

import (
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoUntypedMockFactoryUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUntypedMockFactoryRule,
		[]rule_tester.ValidTestCase{
			{Code: "jest.mock('random-number');"},
			{Code: "jest.mock<typeof import('../moduleName')>('../moduleName', () => {\n  return jest.fn(() => 42);\n});"},
			{Code: "jest.mock<typeof import('./module')>('./module', () => ({\n  ...jest.requireActual('./module'),\n  foo: jest.fn()\n}));"},
			{Code: "jest.mock<typeof import('foo')>('bar', () => ({\n  ...jest.requireActual('bar'),\n  foo: jest.fn()\n}));"},
			{Code: "jest.doMock('./module', (): typeof import('./module') => ({\n  ...jest.requireActual('./module'),\n  foo: jest.fn()\n}));"},
			{Code: "jest.mock('../moduleName', function (): typeof import('../moduleName') {\n  return jest.fn(() => 42);\n});"},
			{Code: "jest.mock<() => number>('random-num', () => {\n  return jest.fn(() => 42);\n});"},
			{Code: "jest['doMock']<() => number>('random-num', () => {\n  return jest.fn(() => 42);\n});"},
			{Code: "jest.mock<any>('random-num', () => {\n  return jest.fn(() => 42);\n});"},
			{Code: "jest.mock(\n  '../moduleName',\n  () => {\n    return jest.fn(() => 42)\n  },\n  {virtual: true},\n);"},
			{Code: "jest.mock('../moduleName', function (): (() => number) {\n  return jest.fn(() => 42);\n});"},
			{Code: "mockito<() => number>('foo', () => {\n  return jest.fn(() => 42);\n});"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "jest.mock('../moduleName', () => {\n  return jest.fn(() => 42);\n});", Output: []string{"jest.mock<typeof import('../moduleName')>('../moduleName', () => {\n  return jest.fn(() => 42);\n});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "jest.mock(\"./module\", () => ({\n  ...jest.requireActual('./module'),\n  foo: jest.fn()\n}));", Output: []string{"jest.mock<typeof import(\"./module\")>(\"./module\", () => ({\n  ...jest.requireActual('./module'),\n  foo: jest.fn()\n}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "jest.mock('random-num', () => {\n  return jest.fn(() => 42);\n});", Output: []string{"jest.mock<typeof import('random-num')>('random-num', () => {\n  return jest.fn(() => 42);\n});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "jest.doMock('random-num', () => {\n  return jest.fn(() => 42);\n});", Output: []string{"jest.doMock<typeof import('random-num')>('random-num', () => {\n  return jest.fn(() => 42);\n});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "jest['mock']('random-num', () => {\n  return jest.fn(() => 42);\n});", Output: []string{"jest['mock']<typeof import('random-num')>('random-num', () => {\n  return jest.fn(() => 42);\n});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "const moduleToMock = 'random-num';\njest.mock(moduleToMock, () => {\n  return jest.fn(() => 42);\n});", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
		})
}
