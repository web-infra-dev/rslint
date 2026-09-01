package utils_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// `rstack/test` is the Rstack CLI's re-export of the Rstest core API, so a
// registration imported from it parses exactly as one imported from
// `@rstest/core`. These cases mirror the `@rstest/core` source forms in
// TestParseRstestFnCallHooks and TestParseRstestExpectCallSources one for one.
func TestParseRstestFnCallRstackTestModule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &hookParseProbe,
		[]rule_tester.ValidTestCase{
			// Only the `test` subpath re-exports the API. Its sibling sub-paths
			// carry the build, app and lint APIs, and the bare package carries
			// the config helper, so a same-named binding out of any of them is
			// not a registration.
			{Code: `import { beforeAll } from 'rstack'; beforeAll(() => {});`},
			{Code: `import { beforeAll } from 'rstack/lib'; beforeAll(() => {});`},
			{Code: `import { beforeAll } from 'rstack/app'; beforeAll(() => {});`},
			{Code: `import { beforeAll } from 'rstack/lint'; beforeAll(() => {});`},
			{Code: `import { beforeAll } from 'rstack/test/globals'; beforeAll(() => {});`},
			{Code: `import { beforeAll } from '@rstack/test'; beforeAll(() => {});`},
			// Hooks accept no chained members, whichever specifier bound them.
			{Code: `import { beforeAll } from 'rstack/test'; beforeAll.skip(() => {});`},
			// `test.beforeEach` stays Playwright-only: `rstack/test` re-exports
			// the core API, where the test object carries no hooks.
			{Code: `import { test } from 'rstack/test'; test.beforeEach(() => {});`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `import { beforeAll } from 'rstack/test'; beforeAll(() => {});`,
				Errors: parsedHookError("hook", "beforeAll"),
			},
			{
				Code:   `import { beforeEach as setup } from 'rstack/test'; setup(() => {});`,
				Errors: parsedHookError("hook", "beforeEach"),
			},
			{
				Code:   `import * as rstest from 'rstack/test'; rstest.afterEach(() => {});`,
				Errors: parsedHookError("hook", "afterEach"),
			},
			{
				Code:   `const { afterAll } = require('rstack/test'); afterAll(() => {});`,
				Errors: parsedHookError("hook", "afterAll"),
			},
			{
				Code:   `const rstest = require('rstack/test'); rstest.beforeAll(() => {});`,
				Errors: parsedHookError("hook", "beforeAll"),
			},
			{
				Code:   `import { beforeAll } from 'rstack/test'; const setup = beforeAll; setup(() => {});`,
				Errors: parsedHookError("hook", "beforeAll"),
			},
			{
				Code:   `import { describe } from 'rstack/test'; describe('suite', () => {});`,
				Errors: parsedHookError("describe", "describe"),
			},
			{
				Code:   `import { test } from 'rstack/test'; test('case', () => {});`,
				Errors: parsedHookError("test", "test"),
			},
			{
				Code:   `import { it } from 'rstack/test'; it.each([1])('case %s', () => {});`,
				Errors: parsedHookError("test", "it"),
			},
			// A type-only import binds no value, so the call is the global
			// registration rather than the imported one — the same reading
			// `@rstest/core` gets for this source.
			{
				Code:   `import type { beforeAll } from 'rstack/test'; beforeAll(() => {});`,
				Errors: parsedHookError("hook", "beforeAll"),
			},
		},
	)
}

// A file may reach the same API through both specifiers at once — a partially
// migrated test, or one importing a helper alongside the registrations. Both
// bindings have to keep resolving: a per-specifier resolution would see the
// other specifier's import as a local declaration shadowing the global and
// abandon it.
func TestParseRstestFnCallMixedRstestModules(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &hookParseProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code: `import { beforeAll } from 'rstack/test';
import { afterAll } from '@rstest/core';
beforeAll(() => {});
afterAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "parsedFn", Message: "kind=hook name=beforeAll"},
					{MessageId: "parsedFn", Message: "kind=hook name=afterAll"},
				},
			},
			{
				Code: `import { describe } from '@rstest/core';
import { test } from 'rstack/test';
describe('suite', () => { test('case', () => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "parsedFn", Message: "kind=describe name=describe"},
					{MessageId: "parsedFn", Message: "kind=test name=test"},
				},
			},
		},
	)
}

func TestParseRstestExpectCallRstackTestModule(t *testing.T) {
	chain := "entry=expect head=true modifiers=[] matcher=toBe reason=none static=false"
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectParseProbe,
		[]rule_tester.ValidTestCase{
			{Code: `import { expect } from 'rstack/lib'; expect(1).toBe(1);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `import { expect } from 'rstack/test'; expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `import { expect as check } from 'rstack/test'; check(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `const { expect } = require('rstack/test'); expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `import * as rstest from 'rstack/test'; rstest.expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `import * as rstest from 'rstack/test'; rstest['expect'](1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `const rstest = require('rstack/test'); rstest.expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
		},
	)
}
