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
			// A computed key that is not a static string names no known export;
			// `expect` here is read as a value, not used as the property name.
			{Code: `const { [expect]: local } = require('@rstest/core');`},
			{Code: `const { [key]: local } = require('@rstest/core');`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: a parenthesized require initializer ----
			{Code: `const { test } = (require('@rstest/core'));`, Output: []string{""}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noRequiringRstestGlobals", Line: 1, Column: 9}}},
			// ---- Real-user: preserve a comment owned by the remaining import ----
			{Code: `import { defineConfig /* keep */, expect } from '@rstest/core';`, Output: []string{`import { defineConfig /* keep */ } from '@rstest/core';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 1, Column: 35}}},
			// ---- Real-user: a comma inside a comment is not the separator ----
			{Code: `import { defineConfig, /* comma, in comment */ expect } from '@rstest/core';`, Output: []string{`import { defineConfig } from '@rstest/core';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 1, Column: 48}}},
			{Code: `import { expect /* comma, in comment */, defineConfig } from '@rstest/core';`, Output: []string{`import { defineConfig } from '@rstest/core';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 1, Column: 10}}},
			// ---- Real-user: a default import survives the named binding removal ----
			// The removal range ends at the end of the named-bindings node, which
			// covers the closing brace, so only the default binding is left.
			{Code: `import core, { expect } from '@rstest/core'; expect(value);`, Output: []string{`import core from '@rstest/core'; expect(value);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 1, Column: 16}}},
			{Code: "import core, {\n  expect,\n} from '@rstest/core';\nexpect(value);", Output: []string{"import core from '@rstest/core';\nexpect(value);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 2, Column: 3}}},
			// Locks in import removal with a surviving type-only specifier.
			{Code: `import { type Mock, test } from '@rstest/core';`, Output: []string{`import { type Mock } from '@rstest/core';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 1, Column: 21}}},
			// ---- Real-user: an aliased sibling keeps the declaration alive ----
			{Code: `import { expect, it as test } from '@rstest/core'; expect(1); test('x');`, Output: []string{`import { it as test } from '@rstest/core'; expect(1); test('x');`}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noImportingRstestGlobals", Line: 1, Column: 10},
				{MessageId: "noImportingRstestGlobals", Line: 1, Column: 18},
			}},
			{Code: `const { expect, it: test } = require('@rstest/core'); expect(1); test('x');`, Output: []string{`const { it: test } = require('@rstest/core'); expect(1); test('x');`}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noRequiringRstestGlobals", Line: 1, Column: 9},
				{MessageId: "noRequiringRstestGlobals", Line: 1, Column: 17},
			}},
			// A sibling used as a value, not invoked, keeps the declaration alive too.
			{Code: `import { expect, test } from '@rstest/core'; expect(1); const runner = test;`, Output: []string{`import { test } from '@rstest/core'; expect(1); const runner = test;`}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noImportingRstestGlobals", Line: 1, Column: 10},
				{MessageId: "noImportingRstestGlobals", Line: 1, Column: 18},
			}},
			// ---- Real-user: a string-literal key names the same export as an
			// identifier key, so it is reported and removed the same way ----
			{Code: `const { 'expect': expect, defineConfig } = require('@rstest/core');`, Output: []string{`const { defineConfig } = require('@rstest/core');`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noRequiringRstestGlobals", Line: 1, Column: 9}}},
			{Code: `const { ['expect']: expect, defineConfig } = require('@rstest/core');`, Output: []string{`const { defineConfig } = require('@rstest/core');`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noRequiringRstestGlobals", Line: 1, Column: 9}}},
			// ---- Real-user: namespace member invocation remains valid as a global ----
			{Code: "import { rs } from '@rstest/core';\nrs.fn();", Output: []string{"\nrs.fn();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportingRstestGlobals", Line: 1, Column: 10}}},
		},
	)
}
