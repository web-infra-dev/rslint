// TestConsistentTestItExtras covers Rstest APIs, scope-safe fixes, AST shapes,
// and historical regressions. The upstream suite is in the sibling file.
package consistent_test_it

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestConsistentTestItExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTestItRule,
		[]rule_tester.ValidTestCase{
			// Dimension 4: unrelated bindings, unsupported APIs, factories, dynamic keys, and TS wrappers.
			{Code: "fit('case'); xit('case'); xtest('case');"},
			// Dimension 4: unrelated bindings, unsupported APIs, factories, dynamic keys, and TS wrappers.
			{Code: "const it = custom; it('case');"},
			// Dimension 4: unrelated bindings, unsupported APIs, factories, dynamic keys, and TS wrappers.
			{Code: "function register(it) { it('case'); }"},
			// A TypeScript namespace is a local value, not the Rstest global.
			{Code: "namespace it {}; it('case');"},
			// Dimension 4: unrelated bindings, unsupported APIs, factories, dynamic keys, and TS wrappers.
			{Code: "import { it } from 'vitest'; it('case');"},
			// Dimension 4: unrelated bindings, unsupported APIs, factories, dynamic keys, and TS wrappers.
			{Code: "import { it } from '@jest/globals'; it('case');"},
			// Dimension 4: unrelated bindings, unsupported APIs, factories, dynamic keys, and TS wrappers.
			{Code: "import { it } from './fixtures'; it('case');"},
			// Dimension 4: unrelated bindings, unsupported APIs, factories, dynamic keys, and TS wrappers.
			{Code: "it.extend({}); it.each([1]); it.for([1]); it.skipIf(true);"},
			// Dimension 4: unrelated bindings, unsupported APIs, factories, dynamic keys, and TS wrappers.
			{Code: "it[method]('case'); it[1]('case'); it[/skip/]('case');"},
			// Dimension 4: unrelated bindings, unsupported APIs, factories, dynamic keys, and TS wrappers.
			{Code: "(it as any)('case'); it!('case'); (it satisfies Function)('case');"},
			// Dimension 4: unrelated bindings, unsupported APIs, factories, dynamic keys, and TS wrappers.
			{Code: "suite('group', () => { test('case'); });"},
			// Dimension 4: unrelated bindings, unsupported APIs, factories, dynamic keys, and TS wrappers.
			{Code: "rs.test('case'); rstest.it('case');"},
			// Real-user: jest #711 explicitly setting fn also sets suite preference.
			{Code: "describe('suite', () => { test('case'); });", Options: map[string]any{"fn": "test"}},
			// Describe exit resets nesting depth.
			{Code: "describe('suite', () => { it('nested'); }); test('outside');"},
			// Playwright has no it export.
			{Code: "import { test } from '@rstest/playwright'; describe('suite', () => { test('browser', () => {}); });"},
			// Playwright exemption survives uniform it option.
			{Code: "import { test } from '@rstest/playwright'; test('browser', () => {});", Options: map[string]any{"fn": "it"}},
			// Playwright has no it export.
			{Code: "import { test as base } from '@rstest/playwright'; const test = base.extend({}); describe('suite', () => { test('browser', () => {}); });"},
			// Playwright exemption survives uniform it option.
			{Code: "import { test as base } from '@rstest/playwright'; const test = base.extend({}); test('browser', () => {});", Options: map[string]any{"fn": "it"}},
			// Playwright has no it export.
			{Code: "import * as pw from '@rstest/playwright'; const test = pw.test.extend({}); describe('suite', () => { test('browser', () => {}); });"},
			// Playwright exemption survives uniform it option.
			{Code: "import * as pw from '@rstest/playwright'; const test = pw.test.extend({}); test('browser', () => {});", Options: map[string]any{"fn": "it"}},
			// Playwright has no it export.
			{Code: "const { test } = require('@rstest/playwright'); describe('suite', () => { test('browser', () => {}); });"},
			// Playwright exemption survives uniform it option.
			{Code: "const { test } = require('@rstest/playwright'); test('browser', () => {});", Options: map[string]any{"fn": "it"}},
			// Playwright namespace and test.describe.
			{Code: "import * as pw from '@rstest/playwright'; pw.test.describe('suite', () => { pw.test.only('browser'); });"},
			// Real-user: Vitest #956 fixture binding already uses the preferred name.
			{Code: "import { test, describe } from '@rstest/core'; const it = test.extend({ account: {} }); describe('suite', () => { it('case'); });"},
			// Explicit canonical import alias chooses the written convention.
			{Code: "import { it as test } from '@rstest/core'; test('case');"},
			// Real-user: Vitest #884 fixture factory is not a registration.
			{Code: "import { test as base } from '@rstest/core'; const test = base.extend({}); test('case');"},
		}, []rule_tester.InvalidTestCase{
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it('registers a case', () => {});", Output: []string{"test('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 3}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test('registers a case', () => {}); });", Output: []string{"describe('group', () => { it('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 31}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.only('registers a case', () => {});", Output: []string{"test.only('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.only('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.only('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 36}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.skip('registers a case', () => {});", Output: []string{"test.skip('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.skip('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.skip('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 36}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.todo('registers a case', () => {});", Output: []string{"test.todo('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.todo('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.todo('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 36}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.fails('registers a case', () => {});", Output: []string{"test.fails('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.fails('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.fails('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 37}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.concurrent('registers a case', () => {});", Output: []string{"test.concurrent('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 14}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.concurrent('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.concurrent('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 42}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.sequential('registers a case', () => {});", Output: []string{"test.sequential('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 14}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.sequential('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.sequential('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 42}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.skip.only('registers a case', () => {});", Output: []string{"test.skip.only('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 13}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.skip.only('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.skip.only('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 41}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.runIf(enabled)('registers a case', () => {});", Output: []string{"test.runIf(enabled)('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 18}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.runIf(enabled)('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.runIf(enabled)('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 46}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.skipIf(disabled)('registers a case', () => {});", Output: []string{"test.skipIf(disabled)('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.skipIf(disabled)('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.skipIf(disabled)('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 48}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.runIf(true).skipIf(false).concurrent('registers a case', () => {});", Output: []string{"test.runIf(true).skipIf(false).concurrent('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 40}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.runIf(true).skipIf(false).concurrent('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.runIf(true).skipIf(false).concurrent('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 68}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.each([[1]])('registers a case', () => {});", Output: []string{"test.each([[1]])('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 15}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.each([[1]])('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.each([[1]])('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 43}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.for([[1]])('registers a case', () => {});", Output: []string{"test.for([[1]])('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 14}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.for([[1]])('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.for([[1]])('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 42}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.only.concurrent.each([[1]])('registers a case', () => {});", Output: []string{"test.only.concurrent.each([[1]])('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 31}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.only.concurrent.each([[1]])('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.only.concurrent.each([[1]])('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 59}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.extend({ account: {} })('registers a case', () => {});", Output: []string{"test.extend({ account: {} })('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 27}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.extend({ account: {} })('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.extend({ account: {} })('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 55}}},
			// Rstest registration chains preserve all modifiers and fixtures.
			{Code: "it.extend({}).extend({}).skip.for([1])('registers a case', () => {});", Output: []string{"test.extend({}).extend({}).skip.for([1])('registers a case', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 39}}},
			// Inside suite registration chain.
			{Code: "describe('group', () => { test.extend({}).extend({}).skip.for([1])('registers a case', () => {}); });", Output: []string{"describe('group', () => { it.extend({}).extend({}).skip.for([1])('registers a case', () => {}); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 67}}},
			// Add import atomically.
			{Code: "import { it } from '@rstest/core'; it('case');", Output: []string{"import { test, it } from '@rstest/core'; test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 36, EndLine: 1, EndColumn: 38}}},
			// Reuse existing named import.
			{Code: "import { it, test } from '@rstest/core'; it('case');", Output: []string{"import { it, test } from '@rstest/core'; test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 42, EndLine: 1, EndColumn: 44}}},
			// Renamed import cannot discard its binding.
			{Code: "import { it as scenario } from '@rstest/core'; scenario('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 48, EndLine: 1, EndColumn: 56}}},
			// ESM namespace keeps modifiers.
			{Code: "import * as core from '@rstest/core'; core.it.only('case');", Output: []string{"import * as core from '@rstest/core'; core.test.only('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 39, EndLine: 1, EndColumn: 51}}},
			// Dimension 4: quoted namespace accessor.
			{Code: "import * as core from '@rstest/core'; core['it'].skip('case');", Output: []string{"import * as core from '@rstest/core'; core['test'].skip('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 39, EndLine: 1, EndColumn: 54}}},
			// CommonJS lacks a preferred binding.
			{Code: "const { it } = require('@rstest/core'); it('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 41, EndLine: 1, EndColumn: 43}}},
			// CommonJS can reuse the existing named base API.
			{Code: "import { test } from '@rstest/core'; const { it } = require('@rstest/core'); it('case');", Output: []string{"import { test } from '@rstest/core'; const { it } = require('@rstest/core'); test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 78, EndLine: 1, EndColumn: 80}}},
			// A reassignable CommonJS binding no longer holds what its pattern declares.
			{Code: "import { test } from '@rstest/core'; let { it } = require('@rstest/core'); it = it.extend({ account: {} }); it('case', ({ account }) => {});", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 109, EndLine: 1, EndColumn: 111}}},
			// A reassignable CommonJS binding no longer holds what its pattern declares.
			{Code: "import { test } from '@rstest/core'; let { it } = require('@rstest/core'); it('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 76, EndLine: 1, EndColumn: 78}}},
			// A class static block is a variable environment of its own.
			{Code: "import { it, test } from '@rstest/core'; class C { static { if (true) { var test = () => {}; } it('case', () => {}); } }", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 96, EndLine: 1, EndColumn: 98}}},
			// Mutable CommonJS namespace is not rewritten.
			{Code: "const core = require('@rstest/core'); core.it('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 39, EndLine: 1, EndColumn: 46}}},
			// Add import atomically.
			{Code: "import { it } from 'rstack/test'; it('case');", Output: []string{"import { test, it } from 'rstack/test'; test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 35, EndLine: 1, EndColumn: 37}}},
			// Reuse existing named import.
			{Code: "import { it, test } from 'rstack/test'; it('case');", Output: []string{"import { it, test } from 'rstack/test'; test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 41, EndLine: 1, EndColumn: 43}}},
			// Renamed import cannot discard its binding.
			{Code: "import { it as scenario } from 'rstack/test'; scenario('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 47, EndLine: 1, EndColumn: 55}}},
			// ESM namespace keeps modifiers.
			{Code: "import * as core from 'rstack/test'; core.it.only('case');", Output: []string{"import * as core from 'rstack/test'; core.test.only('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 38, EndLine: 1, EndColumn: 50}}},
			// Dimension 4: quoted namespace accessor.
			{Code: "import * as core from 'rstack/test'; core['it'].skip('case');", Output: []string{"import * as core from 'rstack/test'; core['test'].skip('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 38, EndLine: 1, EndColumn: 53}}},
			// CommonJS lacks a preferred binding.
			{Code: "const { it } = require('rstack/test'); it('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 40, EndLine: 1, EndColumn: 42}}},
			// CommonJS can reuse the existing named base API.
			{Code: "import { test } from 'rstack/test'; const { it } = require('rstack/test'); it('case');", Output: []string{"import { test } from 'rstack/test'; const { it } = require('rstack/test'); test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 76, EndLine: 1, EndColumn: 78}}},
			// Mutable CommonJS namespace is not rewritten.
			{Code: "const core = require('rstack/test'); core.it('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 38, EndLine: 1, EndColumn: 45}}},
			// In-source API preserves member chains.
			{Code: "import.meta.rstest.it('case');", Output: []string{"import.meta.rstest.test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 22}}},
			// In-source API preserves member chains.
			{Code: "import.meta.rstest.it['only']('case');", Output: []string{"import.meta.rstest.test['only']('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}},
			// In-source API preserves member chains.
			{Code: "import.meta.rstest.it?.skip('case');", Output: []string{"import.meta.rstest.test?.skip('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 28}}},
			// In-source API preserves member chains.
			{Code: "import.meta.rstest.it.each`value\\n${1}`('case');", Output: []string{"import.meta.rstest.test.each`value\\n${1}`('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 40}}},
			// Dimension 4: template accessor replacement.
			{Code: "import.meta['rstest'][`it`]('case');", Output: []string{"import.meta['rstest'][`test`]('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 28}}},
			// Dimension 4: explicit parentheses.
			{Code: "(it)('case');", Output: []string{"(test)('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 5}}},
			// Dimension 4: comma expression preserves left operand.
			{Code: "(0, it)('case');", Output: []string{"(0, test)('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},
			// Dimension 4: optional invocation.
			{Code: "it?.('case');", Output: []string{"test?.('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 3}}},
			// Dimension 4: trivia retained.
			{Code: "it /* keep */ .skip /* keep */ .each([1])('case');", Output: []string{"test /* keep */ .skip /* keep */ .each([1])('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 42}}},
			// Same-file aliases retain captured modifiers and fixtures.
			{Code: "const scenario = it; scenario('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 22, EndLine: 1, EndColumn: 30}}},
			// Same-file aliases retain captured modifiers and fixtures.
			{Code: "const scenario = it.skip; scenario('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 27, EndLine: 1, EndColumn: 35}}},
			// Same-file aliases retain captured modifiers and fixtures.
			{Code: "const scenario = it.extend({ account: {} }); scenario('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 46, EndLine: 1, EndColumn: 54}}},
			// Same-file aliases retain captured modifiers and fixtures.
			{Code: "const cases = it.each([1]); cases('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 29, EndLine: 1, EndColumn: 34}}},
			// Same-file aliases retain captured modifiers and fixtures.
			{Code: "import { it as base } from '@rstest/core'; const it = base.extend({}); it('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 72, EndLine: 1, EndColumn: 74}}},
			// Same-file aliases retain captured modifiers and fixtures.
			{Code: "const { it } = import.meta.rstest; it('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 36, EndLine: 1, EndColumn: 38}}},
			// Same-file aliases retain captured modifiers and fixtures.
			{Code: "const core = import.meta.rstest; core.it('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 34, EndLine: 1, EndColumn: 41}}},
			// Scope collision prevents introducing preferred binding.
			{Code: "function register(test) { it('case'); }", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 27, EndLine: 1, EndColumn: 29}}},
			// Scope collision prevents introducing preferred binding.
			{Code: "const test = custom; it('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 22, EndLine: 1, EndColumn: 24}}},
			// Scope collision prevents introducing preferred binding.
			{Code: "import { test } from 'vitest'; it('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 32, EndLine: 1, EndColumn: 34}}},
			// Scope collision prevents introducing preferred binding.
			{Code: "try {} catch (test) { it('case'); }", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 23, EndLine: 1, EndColumn: 25}}},
			// Scope collision prevents introducing preferred binding.
			{Code: "for (const test of values) { it('case'); }", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 30, EndLine: 1, EndColumn: 32}}},
			// Scope collision prevents introducing preferred binding.
			{Code: "function register({ test }) { it('case'); }", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 31, EndLine: 1, EndColumn: 33}}},
			// Scope collision prevents introducing preferred binding.
			{Code: "namespace group { const test = custom; it('case'); }", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 40, EndLine: 1, EndColumn: 42}}},
			// Scope collision prevents introducing preferred binding.
			{Code: "const register = function test() { it('case'); };", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 36, EndLine: 1, EndColumn: 38}}},
			// Scope collision prevents introducing preferred binding.
			{Code: "import type { test } from '@rstest/core'; it('case');", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 43, EndLine: 1, EndColumn: 45}}},
			// An aliased preferred export does not bind test.
			{Code: "import { test as otherTest, it } from '@rstest/core'; it('case');", Output: []string{"import { test as otherTest, test, it } from '@rstest/core'; test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 55, EndLine: 1, EndColumn: 57}}},
			// Import insertion preserves comments.
			{Code: "import { /* important */ it /* retained */ } from '@rstest/core'; it('case');", Output: []string{"import { /* important */ test, it /* retained */ } from '@rstest/core'; test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 67, EndLine: 1, EndColumn: 69}}},
			// Exports retain original binding.
			{Code: "import { it } from '@rstest/core'; export { it }; it('case');", Output: []string{"import { test, it } from '@rstest/core'; export { it }; test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 51, EndLine: 1, EndColumn: 53}}},
			// Overlapping import insertions converge without duplicating the preferred import.
			{Code: "import { it } from '@rstest/core'; it('one'); it('two');", Output: []string{"import { test, it } from '@rstest/core'; test('one'); it('two');", "import { test, it } from '@rstest/core'; test('one'); test('two');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Line: 1, Column: 36}, {MessageId: "consistentMethod", Line: 1, Column: 47}}},
			// Real-user: jest #795 parameterized suite nesting.
			{Code: "describe('outer', () => { describe.for([1])('inner', () => { test('case'); }); });", Output: []string{"describe('outer', () => { describe.for([1])('inner', () => { it('case'); }); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 62, EndLine: 1, EndColumn: 66}}},
			// Real-user: jest #795 chained suite factory counts only final registration.
			{Code: "describe.only.each([1])('suite', () => { test('case'); });", Output: []string{"describe.only.each([1])('suite', () => { it('case'); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 42, EndLine: 1, EndColumn: 46}}},
			// Real-user: jest #711 fn fallback.
			{Code: "describe('suite', () => { it('case'); });", Options: map[string]any{"fn": "test"}, Output: []string{"describe('suite', () => { test('case'); });"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 29}}},
			// Named callback scope is lexical.
			{Code: "function cases() { it('case'); } describe('suite', cases);", Output: []string{"function cases() { test('case'); } describe('suite', cases);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 20, EndLine: 1, EndColumn: 22}}},
			// Type-only imports do not shadow the global value API.
			{Code: "import type { it } from '@rstest/core'; it('case');", Output: []string{"import type { it } from '@rstest/core'; test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 41, EndLine: 1, EndColumn: 43}}},
			// Type-only imports do not shadow the global value API.
			{Code: "import { type it } from '@rstest/core'; it('case');", Output: []string{"import { type it } from '@rstest/core'; test('case');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 41, EndLine: 1, EndColumn: 43}}},
		})
}

// TestConsistentTestItEditDemand checks source-only operation and ensures
// optional edits never change diagnostics.
func TestConsistentTestItEditDemand(t *testing.T) {
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "consistent-test-it-demand.ts")
	code := `import { it } from '@rstest/core';
it('uses import');
import * as core from 'rstack/test';
core.it('uses namespace');
const custom = it.extend({}); custom('retains fixtures');
import { it as foreign } from 'vitest'; foreign('ignored');
import { test as browser } from '@rstest/playwright';
describe('browser', () => { browser('ignored'); });
namespace test {}; it('namespace collision');
namespace local { namespace it {}; it('not a registration'); }`
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("expected source-only Program")
	}
	var baseline []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: program, File: fileName,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: ConsistentTestItRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners { return ConsistentTestItRule.Run(ctx, nil) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
		})
		if len(diagnostics) != 4 {
			t.Fatalf("demand %d: got %d diagnostics, want 4: %+v", demand, len(diagnostics), diagnostics)
		}
		for i := range diagnostics {
			wantFix := (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll) && i == 1
			if (diagnostics[i].FixesPtr != nil && len(*diagnostics[i].FixesPtr) > 0) != wantFix {
				t.Fatalf("demand %d, diagnostic %d: unexpected fixes", demand, i)
			}
			diagnostics[i].FixesPtr = nil
		}
		if baseline == nil {
			baseline = diagnostics
		} else if !reflect.DeepEqual(baseline, diagnostics) {
			t.Fatalf("demand %d changed diagnostics", demand)
		}
	}
}
