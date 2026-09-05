package prefer_importing_rstest_globals_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_importing_rstest_globals"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferImportingRstestGlobalsUpstream covers the complete baseline
// contract. Rslint-specific edge shapes live in the sibling extras file.
func TestPreferImportingRstestGlobalsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_importing_rstest_globals.PreferImportingRstestGlobalsRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { describe, expect, it } from '@rstest/core'; describe('suite', () => it('works', () => expect(1).toBe(1)));`},
			{Code: `import { describe } from 'rstack/test'; describe('suite', () => {});`},
			{Code: `const { it } = require('@rstest/core'); it('works', () => {});`},
			{Code: `function describe(title: string) {} describe('local');`},
			{Code: `function run(it: () => void) { it(); }`},
			{Code: `const expect = makeExpectation; expect(value);`},
			{Code: `import * as core from '@rstest/core'; core.describe('suite', () => {});`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "describe('user service', () => {\n  it('reads a user', () => {\n    expect(getUser('1')).toEqual(user);\n  });\n});",
				Output: []string{"import { describe, expect, it } from '@rstest/core';\ndescribe('user service', () => {\n  it('reads a user', () => {\n    expect(getUser('1')).toEqual(user);\n  });\n});"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Message: "Import `describe, it, expect` from `@rstest/core`.", Line: 1, Column: 1}},
			},
			{
				Code: `rs.fn();`, Output: []string{`import { rs } from '@rstest/core';
rs.fn();`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 1, Column: 1}},
			},
			{
				Code: `import { defineConfig } from 'rstack/test';
test('works', () => {});`, Output: []string{`import { defineConfig, test } from 'rstack/test';
test('works', () => {});`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}},
			},
			{
				Code: `'use strict';
expect(value);`, Output: []string{`'use strict';
import { expect } from '@rstest/core';
expect(value);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}},
			},
			{
				Code: `expect(value);`, LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Output: []string{`const { expect } = require('@rstest/core');
expect(value);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 1, Column: 1}},
			},
		},
	)
}
