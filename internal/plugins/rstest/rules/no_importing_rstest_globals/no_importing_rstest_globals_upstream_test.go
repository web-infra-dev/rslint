package no_importing_rstest_globals_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_importing_rstest_globals"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoImportingRstestGlobalsUpstream covers the complete baseline contract
// for the rule. Rslint-specific edge shapes live in the sibling extras file.
func TestNoImportingRstestGlobalsUpstream(t *testing.T) {
	globals := []string{"test", "describe", "it", "expect", "afterAll", "afterEach", "beforeAll", "beforeEach", "rstest", "rs", "assert", "onTestFinished", "onTestFailed"}
	invalid := make([]rule_tester.InvalidTestCase, 0, len(globals)+5)
	for _, name := range globals {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:   fmt.Sprintf("import { %s } from '@rstest/core';", name),
			Output: []string{""},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noImportingRstestGlobals", Line: 1, Column: 10,
			}},
		})
	}
	invalid = append(invalid,
		rule_tester.InvalidTestCase{
			Code:   `import { defineConfig, expect } from 'rstack/test';`,
			Output: []string{`import { defineConfig } from 'rstack/test';`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Message: "Do not import `expect` from `rstack/test`; it is available as a global.", Line: 1, Column: 24}},
		},
		rule_tester.InvalidTestCase{
			Code:   `const { describe, defineConfig } = require('@rstest/core');`,
			Output: []string{`const { defineConfig } = require('@rstest/core');`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noRequiringRstestGlobals", Message: "Do not require `describe` from `@rstest/core`; it is available as a global.", Line: 1, Column: 9}},
		},
		rule_tester.InvalidTestCase{
			Code:   `import { it as test } from '@rstest/core'; test('works', () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 1, Column: 10}},
		},
		rule_tester.InvalidTestCase{
			Code:   `import { expect } from '@rstest/core'; const matcher = expect;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 1, Column: 10}},
		},
		rule_tester.InvalidTestCase{
			Code: `const { it = fallback, expect, ...others } = require('@rstest/core');`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRequiringRstestGlobals", Line: 1, Column: 9},
				{MessageId: "noRequiringRstestGlobals", Line: 1, Column: 24},
			},
		},
	)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_importing_rstest_globals.NoImportingRstestGlobalsRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { defineConfig } from '@rstest/core';`},
			{Code: `import type { Mock } from '@rstest/core';`},
			{Code: `import { type Mock } from '@rstest/core';`},
			{Code: `import * as core from '@rstest/core';`},
			{Code: `import core from '@rstest/core';`},
			{Code: `import { test } from 'vitest';`},
			{Code: `import { test } from 'rstack/lib';`},
			{Code: `import { test } from 'rstack';`},
			{Code: "const { test } = require(`@rstest/core${version}`);"},
		}, invalid)
}
