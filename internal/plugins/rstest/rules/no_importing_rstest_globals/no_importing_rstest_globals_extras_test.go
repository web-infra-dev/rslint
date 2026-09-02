package no_importing_rstest_globals_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_importing_rstest_globals"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoImportingRstestGlobalsExtras locks in comments, wrappers, and require
// branches beyond the baseline suite in the sibling upstream file.
func TestNoImportingRstestGlobalsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_importing_rstest_globals.NoImportingRstestGlobalsRule,
		[]rule_tester.ValidTestCase{
			// N/A: optional chains and literal kinds do not affect import discovery.
			{Code: `const value = ({ test: local }).test;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: a parenthesized require initializer ----
			{Code: `const { test } = (require('@rstest/core'));`, Output: []string{""}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noRequiringRstestGlobals", Line: 1, Column: 9}}},
			// ---- Real-user: preserve a comment owned by the remaining import ----
			{Code: `import { defineConfig /* keep */, expect } from '@rstest/core';`, Output: []string{`import { defineConfig /* keep */ } from '@rstest/core';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 1, Column: 35}}},
			// Locks in import removal with a surviving type-only specifier.
			{Code: `import { type Mock, test } from '@rstest/core';`, Output: []string{`import { type Mock } from '@rstest/core';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 1, Column: 21}}},
			// ---- Real-user: namespace member invocation remains valid as a global ----
			{Code: "import { rs } from '@rstest/core';\nrs.fn();", Output: []string{"\nrs.fn();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 1, Column: 10}}},
		},
	)
}
