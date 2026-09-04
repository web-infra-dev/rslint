package prefer_importing_rstest_globals_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_importing_rstest_globals"
	"github.com/web-infra-dev/rslint/internal/rule"
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
			// A shorthand that is assigned to writes the binding instead of reading it.
			{Code: `({ expect } = holder);`},
			// A write to the global cannot be satisfied by an import, whose
			// binding is read-only.
			{Code: `expect = value;`},
			{Code: `expect += 1;`},
			{Code: `expect++;`},
			{Code: `[expect] = holder;`},
			// A type position does not use the API at runtime.
			{Code: `const value: expect = input;`},
			{Code: `type Alias = typeof expect;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: globals used through member and optional-call chains ----
			{Code: `expect.any(String); rs?.fn();`, Output: []string{`import { expect, rs } from '@rstest/core';
expect.any(String); rs?.fn();`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 1, Column: 1}}},
			// ---- Real-user: merge into an existing destructured require ----
			{Code: `const { defineConfig } = require('rstack/test');
beforeEach(setup);`, Output: []string{`const { defineConfig, beforeEach } = require('rstack/test');
beforeEach(setup);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
			// ---- Real-user: an aliased require binding keeps its own syntax ----
			{Code: `const { it: testCase } = require('@rstest/core');
expect(value);`, Output: []string{`const { it: testCase, expect } = require('@rstest/core');
expect(value);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
			// A rest element has to stay last in the pattern.
			{Code: `const { defineConfig, ...others } = require('@rstest/core');
expect(value);`, Output: []string{`const { defineConfig, expect, ...others } = require('@rstest/core');
expect(value);`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
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
			// Locks in that the insertion clears the whole directive prologue, so
			// `use strict` keeps applying to the file.
			{Code: "\"use asm\";\n'use strict';\nexpect(value);", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Output: []string{"\"use asm\";\n'use strict';\nconst { expect } = require('@rstest/core');\nexpect(value);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 3, Column: 1}}},
			// ---- Real-user: a type-only import binds no runtime value, so the
			// file still needs the import this rule asks for ----
			// No fix: every edit available here would bind the name twice.
			{Code: "import { type expect } from '@rstest/core';\nexpect(value);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
			{Code: "import type { expect } from '@rstest/core';\nexpect(value);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
			// A type-only import of a different name still merges normally.
			{Code: "import type { Mock } from '@rstest/core';\nexpect(value);", Output: []string{"import { expect } from '@rstest/core';\nimport type { Mock } from '@rstest/core';\nexpect(value);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 2, Column: 1}}},
			// ---- Dimension 4: a shorthand property value reads the global ----
			{Code: `const value = { expect };`, Output: []string{`import { expect } from '@rstest/core';
const value = { expect };`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 1, Column: 17}}},
			// Locks in de-duplication and source-order diagnostics for all 13 names.
			{Code: `test; describe; it; expect; afterAll; afterEach; beforeAll; beforeEach; rstest; rs; assert; onTestFinished; onTestFailed; test;`, Output: []string{`import { afterAll, afterEach, assert, beforeAll, beforeEach, describe, expect, it, onTestFailed, onTestFinished, rs, rstest, test } from '@rstest/core';
test; describe; it; expect; afterAll; afterEach; beforeAll; beforeEach; rstest; rs; assert; onTestFinished; onTestFailed; test;`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferImportingRstestGlobals", Line: 1, Column: 1}}},
		},
	)
}
