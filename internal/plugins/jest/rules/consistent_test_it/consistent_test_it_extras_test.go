package consistent_test_it_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/consistent_test_it"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Diagnostics below match eslint-plugin-jest v29.16.6 run with
// @typescript-eslint/parser. Outputs match too, except in the "chain" subtest.
func TestConsistentTestItExtras(t *testing.T) {
	// ESTree/tsgo shape differences: parentheses, computed and optional access, type arguments, assertions and non-registrations.
	t.Run("shape", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "(it as any)('x')", Options: map[string]any{"fn": "test"}},
				{Code: "(<any>it)('x')", Options: map[string]any{"fn": "test"}},
				{Code: "it!('x')", Options: map[string]any{"fn": "test"}},
				{Code: "it.each([])", Options: map[string]any{"fn": "test"}},
				{Code: "it.only", Options: map[string]any{"fn": "test"}},
				{Code: "it.foo('x')", Options: map[string]any{"fn": "test"}},
				{Code: "new it('x')", Options: map[string]any{"fn": "test"}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code: "(it)('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"(test)('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 2, EndLine: 1, EndColumn: 4}},
				},
				{
					Code: "(it).only('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"(test).only('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 10}},
				},
				{
					Code: "(it.only)('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"(test.only)('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 2, EndLine: 1, EndColumn: 9}},
				},
				{
					Code: "(it.each([]))('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"(test.each([]))('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 2, EndLine: 1, EndColumn: 13}},
				},
				{
					Code: "it['only']('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test['only']('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 11}},
				},
				{
					Code: "it[`only`]('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test[`only`]('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 11}},
				},
				{
					Code: "it?.('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test?.('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 3}},
				},
				{
					Code: "it.only?.('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.only?.('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}},
				},
				{
					Code: "it<string>('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test<string>('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 3}},
				},
				{
					Code: "it.each<string>([])('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.each<string>([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}},
				},
				{
					Code: "it.each`a`('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.each`a`('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 11}},
				},
				{
					Code: "it\n  .only('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test\n  .only('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 2, EndColumn: 8}},
				},
				{
					Code: "it/*c*/.only('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test/*c*/.only('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
				},
				{
					Code:   "describe('𝒜', () => {\n  test('é')\n})",
					Output: []string{"describe('𝒜', () => {\n  it('é')\n})"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 2, Column: 3, EndLine: 2, EndColumn: 7}},
				},
			},
		)
	})
	// Prefixed names, fit rewriting and modifiers that upstream tests do not cover.
	t.Run("names", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "fit('x')", Options: map[string]any{"fn": "it"}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code: "fit('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.only('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
				},
				{
					Code: "fit.each([])('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.only.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
				},
				{
					Code: "xit.each([])('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"xtest.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 13}},
				},
				{
					Code: "xtest.each([])('x')", Options: map[string]any{"fn": "it"},
					Output: []string{"xit.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 15}},
				},
				{
					Code: "xit.failing('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"xtest.failing('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 12}},
				},
				{
					Code: "fit.failing('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.only.failing('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 12}},
				},
				{
					Code: "describe('s', () => { fit('x') })", Options: map[string]any{"fn": "it", "withinDescribe": "test"},
					Output: []string{"describe('s', () => { test.only('x') })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 23, EndLine: 1, EndColumn: 26}},
				},
				{
					Code: "it.failing('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.failing('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 11}},
				},
				{
					Code: "it.todo('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.todo('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}},
				},
				{
					Code: "test.todo('x')", Options: map[string]any{"fn": "it"},
					Output: []string{"it.todo('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 10}},
				},
			},
		)
	})
	// Imports, CommonJS, shadowing, other modules and settings.jest.globalAliases.
	t.Run("binding", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "import * as j from '@jest/globals';\nj.it('x')", Options: map[string]any{"fn": "test"}},
				{Code: "import { it } from 'vitest';\nit('x')", Options: map[string]any{"fn": "test"}},
				{Code: "function f(it) { it('x') }", Options: map[string]any{"fn": "test"}},
				{Code: "const it = () => {};\nit('x')", Options: map[string]any{"fn": "test"}},
				{Code: "function f(describe) { describe('s', () => {}) }\ntest('x')"},
				{Code: "specify('x')", Options: map[string]any{"fn": "it"}, Settings: map[string]any{"jest": map[string]any{"globalAliases": map[string]any{"it": []any{"specify"}}}}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code: "import { it } from '@jest/globals';\nit('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"import { it } from '@jest/globals';\ntest('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 2, Column: 1, EndLine: 2, EndColumn: 3}},
				},
				{
					Code: "import { it as t } from '@jest/globals';\nt.only('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"import { it as t } from '@jest/globals';\ntest.only('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 2, Column: 1, EndLine: 2, EndColumn: 7}},
				},
				{
					Code: "import { fit as t } from '@jest/globals';\nt('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"import { fit as t } from '@jest/globals';\ntest.only('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 2, Column: 1, EndLine: 2, EndColumn: 2}},
				},
				{
					Code: "import { xit as t } from '@jest/globals';\nt.each([])('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"import { xit as t } from '@jest/globals';\nxtest.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 2, Column: 1, EndLine: 2, EndColumn: 11}},
				},
				{
					Code: "const { it } = require('@jest/globals');\nit('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"const { it } = require('@jest/globals');\ntest('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 2, Column: 1, EndLine: 2, EndColumn: 3}},
				},
				{
					Code: "specify('x')", Options: map[string]any{"fn": "test"}, Settings: map[string]any{"jest": map[string]any{"globalAliases": map[string]any{"it": []any{"specify"}}}},
					Output: []string{"test('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}},
				},
				{
					Code: "context('s', () => { test('x') })", Settings: map[string]any{"jest": map[string]any{"globalAliases": map[string]any{"describe": []any{"context"}}}},
					Output: []string{"context('s', () => { it('x') })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 22, EndLine: 1, EndColumn: 26}},
				},
			},
		)
	})
	// Suite depth: describe variants, factories without a suite call, callbacks passed by reference and describe inside a test.
	t.Run("depth", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "describe.each([])"},
				{Code: "const helper = () => { test('x') };\ndescribe('a', helper)"},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code:   "describe('a', () => {}); it('x')",
					Output: []string{"describe('a', () => {}); test('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 26, EndLine: 1, EndColumn: 28}},
				},
				{
					Code:   "describe('a', () => { describe('b', () => { test('y') }); test('x') }); test('z')",
					Output: []string{"describe('a', () => { describe('b', () => { it('y') }); it('x') }); test('z')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 45, EndLine: 1, EndColumn: 49}, {MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 59, EndLine: 1, EndColumn: 63}},
				},
				{
					Code:   "describe.each([])('a', () => { test('x') }); it('z')",
					Output: []string{"describe.each([])('a', () => { it('x') }); test('z')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 32, EndLine: 1, EndColumn: 36}, {MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 46, EndLine: 1, EndColumn: 48}},
				},
				{
					Code:   "describe.skip('a', () => { test('x') })",
					Output: []string{"describe.skip('a', () => { it('x') })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 28, EndLine: 1, EndColumn: 32}},
				},
				{
					Code:   "xdescribe('a', () => { test('x') })",
					Output: []string{"xdescribe('a', () => { it('x') })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 24, EndLine: 1, EndColumn: 28}},
				},
				{
					Code:   "fdescribe('a', () => { test('x') })",
					Output: []string{"fdescribe('a', () => { it('x') })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 24, EndLine: 1, EndColumn: 28}},
				},
				{
					Code:   "describe.only.each``('a', () => { test('x') })",
					Output: []string{"describe.only.each``('a', () => { it('x') })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 35, EndLine: 1, EndColumn: 39}},
				},
				{
					Code:   "describe('a', () => { test('x', () => { it('y') }) })",
					Output: []string{"describe('a', () => { it('x', () => { it('y') }) })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 23, EndLine: 1, EndColumn: 27}},
				},
				{
					Code:   "test('a', () => { describe('b', () => { test('x') }) })",
					Output: []string{"test('a', () => { describe('b', () => { it('x') }) })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 41, EndLine: 1, EndColumn: 45}},
				},
			},
		)
	})
	// Calls that are one link of a longer chain (their result is called, accessed or passed on) are not registrations; .fails is not a Jest API, .failing.each is.
	t.Run("parser", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "(describe('a', () => { test('x') })).foo"},
				{Code: "describe('a', () => { test('x') })?.foo"},
				{Code: "describe('a', () => { test('x') })['foo']"},
				{Code: "wrap(describe('a', () => { test('x') }))"},
				{Code: "describe('a', () => { test('x') }).only('b', () => { test('y') })"},
				{Code: "it('x').only('y')", Options: map[string]any{"fn": "test"}},
				{Code: "it('x').foo", Options: map[string]any{"fn": "test"}},
				{Code: "wrap(it('x'))", Options: map[string]any{"fn": "test"}},
				{Code: "it.fails('x')", Options: map[string]any{"fn": "test"}},
				{Code: "test.skip.fails('x')", Options: map[string]any{"fn": "it"}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code:   "describe('a', () => { test('x') }).foo; it('y')",
					Output: []string{"describe('a', () => { test('x') }).foo; test('y')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 41, EndLine: 1, EndColumn: 43}},
				},
				{
					Code:   "describe('a', () => { test('x') })()",
					Output: []string{"describe('a', () => { it('x') })()"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 23, EndLine: 1, EndColumn: 27}},
				},
				{
					Code:   "describe('a', () => { test('x') })!.foo",
					Output: []string{"describe('a', () => { it('x') })!.foo"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 23, EndLine: 1, EndColumn: 27}},
				},
				{
					Code:   "(describe('a', () => { test('x') }) as any).foo",
					Output: []string{"(describe('a', () => { it('x') }) as any).foo"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 24, EndLine: 1, EndColumn: 28}},
				},
				{
					Code: "const t = it('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"const t = test('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 11, EndLine: 1, EndColumn: 13}},
				},
			},
		)
	})
	// Chains with more than one member keep every modifier. Upstream rewrites the whole object and drops them (it.only.each becomes test.each); rslint replaces only the root.
	t.Run("chain", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{},
			[]rule_tester.InvalidTestCase{
				{
					Code: "it.only.each([])('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.only.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}},
				},
				{
					Code: "it.skip.each([])('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.skip.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}},
				},
				{
					Code: "it.only.failing('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.only.failing('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 16}},
				},
				{
					Code: "test.concurrent.only.each([])('x')", Options: map[string]any{"fn": "it"},
					Output: []string{"it.concurrent.only.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}},
				},
				{
					Code:   "describe('s', () => { test.concurrent.skip.each``('x') })",
					Output: []string{"describe('s', () => { it.concurrent.skip.each``('x') })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 23, EndLine: 1, EndColumn: 50}},
				},
				{
					Code: "import { it as t } from '@jest/globals';\nt.only.each([])('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"import { it as t } from '@jest/globals';\ntest.only.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 2, Column: 1, EndLine: 2, EndColumn: 16}},
				},
				{
					Code: "it.skip.failing.each([])('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.skip.failing.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 25}},
				},
				{
					Code: "test.failing.each``('x')", Options: map[string]any{"fn": "it"},
					Output: []string{"it.failing.each``('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}},
				},
				{
					Code: "xit.failing.each([])('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"xtest.failing.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 21}},
				},
				{
					Code: "fit.failing.each([])('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.only.failing.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 21}},
				},
				{
					Code: "it.concurrent.failing('x')", Options: map[string]any{"fn": "test"},
					Output: []string{"test.concurrent.failing('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 22}},
				},
				{
					Code: "test.concurrent.failing.only.each([])('x')", Options: map[string]any{"fn": "it"},
					Output: []string{"it.concurrent.failing.only.each([])('x')"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 38}},
				},
			},
		)
	})
}

// TestConsistentTestItEditDemand ensures requesting edits never changes the
// diagnostics, and that fixes are built only when autofixes are requested.
func TestConsistentTestItEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"describe('s', () => { test('x') });\nfit('y');",
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}
	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: consistent_test_it.ConsistentTestItRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return consistent_test_it.ConsistentTestItRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}
	withoutEdits := func(diagnostics []rule.RuleDiagnostic) []rule.RuleDiagnostic {
		out := make([]rule.RuleDiagnostic, len(diagnostics))
		for i, diagnostic := range diagnostics {
			diagnostic.FixesPtr = nil
			diagnostic.Suggestions = nil
			out[i] = diagnostic
		}
		return out
	}
	allEdits := run(rule.EditDemandAll)
	want := withoutEdits(allEdits)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		diagnostics := run(demand)
		if got := withoutEdits(diagnostics); !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostics:\ngot  %#v\nwant %#v", demand, got, want)
		}
		for i, diagnostic := range diagnostics {
			if diagnostic.Suggestions != nil {
				t.Fatalf("demand %d, diagnostic %d: unexpected suggestions", demand, i)
			}
			if demand != rule.EditDemandAutofix {
				if diagnostic.FixesPtr != nil {
					t.Fatalf("demand %d, diagnostic %d: autofix built without being requested", demand, i)
				}
				continue
			}
			if diagnostic.FixesPtr == nil || !reflect.DeepEqual(*diagnostic.FixesPtr, *allEdits[i].FixesPtr) {
				t.Fatalf("diagnostic %d: requested autofix does not match all-edits output", i)
			}
		}
	}
}
