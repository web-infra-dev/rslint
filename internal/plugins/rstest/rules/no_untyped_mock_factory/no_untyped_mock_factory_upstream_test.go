// TestNoUntypedMockFactoryUpstream migrates every eslint-plugin-jest v29.16.6
// case. Rstest reverses the bracket-call case because its transform ignores it.
// Additional shapes and branch coverage live in no_untyped_mock_factory_extras_test.go.
package no_untyped_mock_factory

import (
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoUntypedMockFactoryUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUntypedMockFactoryRule,
		[]rule_tester.ValidTestCase{
			{Code: "rs.mock('random-number');"},
			{Code: "rs.mock<typeof import('../moduleName')>('../moduleName', () => {\n  return rs.fn(() => 42);\n});"},
			{Code: "rs.mock<typeof import('./module')>('./module', () => ({\n  ...rs.requireActual('./module'),\n  foo: rs.fn()\n}));"},
			{Code: "rs.mock<typeof import('foo')>('bar', () => ({\n  ...rs.requireActual('bar'),\n  foo: rs.fn()\n}));"},
			{Code: "rs.doMock('./module', (): typeof import('./module') => ({\n  ...rs.requireActual('./module'),\n  foo: rs.fn()\n}));"},
			{Code: "rs.mock('../moduleName', function (): typeof import('../moduleName') {\n  return rs.fn(() => 42);\n});"},
			{Code: "rs.mock<() => number>('random-num', () => {\n  return rs.fn(() => 42);\n});"},
			{Code: "rs['doMock']<() => number>('random-num', () => {\n  return rs.fn(() => 42);\n});"},
			{Code: "rs.mock<any>('random-num', () => {\n  return rs.fn(() => 42);\n});"},
			{Code: "rs.mock(\n  '../moduleName',\n  () => {\n    return rs.fn(() => 42)\n  },\n  {virtual: true},\n);"},
			{Code: "rs.mock('../moduleName', function (): (() => number) {\n  return rs.fn(() => 42);\n});"},
			{Code: "mockito<() => number>('foo', () => {\n  return rs.fn(() => 42);\n});"},
			{Code: "rs['mock']('random-num', () => {\n  return rs.fn(() => 42);\n});"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "rs.mock('../moduleName', () => {\n  return rs.fn(() => 42);\n});", Output: []string{"rs.mock<typeof import('../moduleName')>('../moduleName', () => {\n  return rs.fn(() => 42);\n});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "rs.mock(\"./module\", () => ({\n  ...rs.requireActual('./module'),\n  foo: rs.fn()\n}));", Output: []string{"rs.mock<typeof import(\"./module\")>(\"./module\", () => ({\n  ...rs.requireActual('./module'),\n  foo: rs.fn()\n}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "rs.mock('random-num', () => {\n  return rs.fn(() => 42);\n});", Output: []string{"rs.mock<typeof import('random-num')>('random-num', () => {\n  return rs.fn(() => 42);\n});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "rs.doMock('random-num', () => {\n  return rs.fn(() => 42);\n});", Output: []string{"rs.doMock<typeof import('random-num')>('random-num', () => {\n  return rs.fn(() => 42);\n});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
			{Code: "const moduleToMock = 'random-num';\nrs.mock(moduleToMock, () => {\n  return rs.fn(() => 42);\n});", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "addTypeParameterToModuleMock"}}},
		})
}
