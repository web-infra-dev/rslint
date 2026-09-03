package prefer_importing_rstest_globals_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_importing_rstest_globals"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferImportingRstestGlobalsExtras locks in identifier shapes and merge
// branches beyond the baseline suite in the sibling upstream file.
func TestPreferImportingRstestGlobalsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_importing_rstest_globals.PreferImportingRstestGlobalsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: member names are not global references ----
			{Code: `service.expect(); const value = { describe: callback };`},
			// N/A: optional-chain flags do not change identifier binding resolution.
			{Code: `const rs = service?.rs; rs.fn();`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: globals used through member and optional-call chains ----
			{Code: `expect.any(String); rs?.fn();`, Output: []string{`import { expect, rs } from '@rstest/core';
expect.any(String); rs?.fn();`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 1, Column: 1}}},
			// ---- Real-user: merge into an existing destructured require ----
			{Code: `const { defineConfig } = require('rstack/test');
beforeEach(setup);`, Output: []string{`import { beforeEach, defineConfig } from 'rstack/test';
beforeEach(setup);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
			// ---- Real-user: preserve an existing import alias while merging ----
			{Code: `import { it as testCase } from '@rstest/core';
expect(value);`, Output: []string{`import { it as testCase, expect } from '@rstest/core';
expect(value);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
			// ---- Real-user: merging keeps a default binding on the same import ----
			{Code: `import core, { defineConfig } from '@rstest/core';
expect(value);`, Output: []string{`import core, { defineConfig, expect } from '@rstest/core';
expect(value);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
			// ---- Real-user: merging keeps a comment inside the existing braces ----
			{Code: `import { defineConfig /* keep */ } from '@rstest/core';
expect(value);`, Output: []string{`import { defineConfig, expect /* keep */ } from '@rstest/core';
expect(value);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
			// Locks in that only a string literal opens the directive prologue.
			{Code: "`use strict`;\nexpect(value);", Output: []string{"import { expect } from '@rstest/core';\n`use strict`;\nexpect(value);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
			{Code: "('use strict');\nexpect(value);", Output: []string{"import { expect } from '@rstest/core';\n('use strict');\nexpect(value);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
			// Locks in de-duplication and source-order diagnostics for all 13 names.
			{Code: `test; describe; it; expect; afterAll; afterEach; beforeAll; beforeEach; rstest; rs; assert; onTestFinished; onTestFailed; test;`, Output: []string{`import { afterAll, afterEach, assert, beforeAll, beforeEach, describe, expect, it, onTestFailed, onTestFinished, rs, rstest, test } from '@rstest/core';
test; describe; it; expect; afterAll; afterEach; beforeAll; beforeEach; rstest; rs; assert; onTestFinished; onTestFailed; test;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 1, Column: 1}}},
		},
	)
}
