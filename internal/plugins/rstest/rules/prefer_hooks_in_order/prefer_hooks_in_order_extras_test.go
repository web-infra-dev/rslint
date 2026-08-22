package prefer_hooks_in_order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_hooks_in_order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferHooksInOrderExtras locks in branches and edge shapes that the
// upstream Jest suite does not exercise for Rstest. Each case carries an inline
// comment for the continuity rule, source shape, or hook-frame branch it
// protects so future refactors cannot silently regress it.
func TestPreferHooksInOrderExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_hooks_in_order.PreferHooksInOrderRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: Rstest sources resolve to the same hook names ----
			{Code: `import { beforeAll, beforeEach, afterEach, afterAll } from "@rstest/core";
beforeAll(() => {});
beforeEach(() => {});
afterEach(() => {});
afterAll(() => {});`},
			{Code: `import { beforeAll as setup, afterAll as teardown } from "@rstest/core";
setup(() => {});
teardown(() => {});`},
			{Code: `const { beforeAll, afterAll } = require("@rstest/core");
beforeAll(() => {});
afterAll(() => {});`},
			{Code: `import * as rstest from "@rstest/core";
rstest.beforeAll(() => {});
rstest.afterAll(() => {});`},
			{Code: `import.meta.rstest.beforeAll(() => {});
import.meta.rstest.afterAll(() => {});`},
			{Code: `const api = import.meta.rstest;
api.beforeAll(() => {});
api.afterAll(() => {});`},
			{Code: `const setup = beforeAll;
const teardown = afterAll;
setup(() => {});
teardown(() => {});`},
			// ---- Real-user: Playwright member hooks and extend() share the same ordering ----
			{Code: `import { test } from "@rstest/playwright";
test.beforeAll(() => {});
test.beforeEach(() => {});
test.afterEach(() => {});
test.afterAll(() => {});`},
			{Code: `import { test } from "@rstest/playwright";
const appTest = test.extend({});
appTest.beforeAll(() => {});
appTest.beforeEach(() => {});
appTest.afterEach(() => {});
appTest.afterAll(() => {});`},
			// ---- Dimension 4: mixed sources participate in one continuous run ----
			{Code: `import { beforeAll as setup } from "@rstest/core";
setup(() => {});
import.meta.rstest.beforeEach(() => {});
const { afterEach } = require("@rstest/core");
afterEach(() => {});
afterAll(() => {});`},
			// ---- Locks in call-event continuity: non-call statements do not reset ----
			{Code: `beforeAll(() => {});
const marker = 1;
afterAll(() => {});`},
			{Code: `beforeAll(() => {});
if (condition) {}
afterAll(() => {});`},
			// ---- Locks in barrier semantics: non-hook calls reset the run ----
			{Code: `afterAll(() => {});
doSomething();
beforeAll(() => {});`},
			{Code: `afterAll(() => {});
test("case", () => {});
beforeAll(() => {});`},
			{Code: `afterAll(() => {});
describe("suite", () => {});
beforeAll(() => {});`},
			// ---- Locks in hook callback isolation: calls inside a hook callback are ignored ----
			{Code: `beforeAll(() => {
  doSomething();
  helper();
});
afterAll(() => {});`},
			// ---- Dimension 4: core invalid member hooks become barriers, not hooks ----
			{Code: `import { test } from "@rstest/core";
afterAll(() => {});
test.beforeEach(() => {});
beforeAll(() => {});`},
			// ---- Dimension 4: invalid Playwright member chains also become barriers ----
			{Code: `import { test } from "@rstest/playwright";
afterAll(() => {});
test.skip.beforeEach(() => {});
beforeAll(() => {});`},
			{Code: `import { test } from "@rstest/playwright";
afterAll(() => {});
test.beforeEach.only(() => {});
beforeAll(() => {});`},
			// ---- Dimension 4: foreign imports and local shadows are barriers ----
			{Code: `import { beforeAll } from "vitest";
afterAll(() => {});
beforeAll(() => {});
beforeAll(() => {});
beforeAll(() => {});`},
			{Code: `const beforeAll = createHook();
afterAll(() => {});
beforeAll(() => {});
beforeAll(() => {});`},
			// ---- Dimension 4: execution-time and around APIs are barriers ----
			{Code: `afterAll(() => {});
onTestFinished(() => {});
beforeAll(() => {});`},
			{Code: `afterAll(() => {});
aroundEach(() => {});
beforeAll(() => {});`},
			// ---- Dimension 4: optional import.meta hook source parses and participates ----
			{Code: `import.meta.rstest?.beforeAll(() => {});
import.meta.rstest?.afterAll(() => {});`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: alias import participates in the same ordering semantics ----
			{
				Code: `import { afterAll as teardown, beforeAll as setup } from "@rstest/core";
teardown(() => {});
setup(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 3, 1),
				},
			},
			// ---- Dimension 4: require destructuring participates in the same ordering semantics ----
			{
				Code: `const { afterEach, beforeEach } = require("@rstest/core");
afterEach(() => {});
beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterEach", 3, 1),
				},
			},
			// ---- Dimension 4: namespace imports participate in the same ordering semantics ----
			{
				Code: `import * as rstest from "@rstest/core";
rstest.afterAll(() => {});
rstest.beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 3, 1),
				},
			},
			// ---- Dimension 4: import.meta direct and object alias sources participate ----
			{
				Code: `import.meta.rstest.afterEach(() => {});
import.meta.rstest.beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterEach", 2, 1),
				},
			},
			{
				Code: `const api = import.meta.rstest;
api.afterAll(() => {});
api.beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 3, 1),
				},
			},
			// ---- Real-user: same-file aliases preserve resolved hook names ----
			{
				Code: `const teardown = afterAll;
const setup = beforeAll;
teardown(() => {});
setup(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 4, 1),
				},
			},
			// ---- Real-user: Playwright member hooks participate in ordering ----
			{
				Code: `import { test } from "@rstest/playwright";
test.afterAll(() => {});
test.beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 3, 1),
				},
			},
			{
				Code: `import { test } from "@rstest/playwright";
const appTest = test.extend({});
appTest.afterEach(() => {});
appTest.beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterEach", 4, 1),
				},
			},
			// ---- Dimension 4: mixed sources remain one continuous run ----
			{
				Code: `import { afterAll as teardown } from "@rstest/core";
teardown(() => {});
import.meta.rstest.beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 3, 1),
				},
			},
			// ---- Locks in call-event continuity: non-call statements do not reset ----
			{
				Code: `afterAll(() => {});
const marker = 1;
beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 3, 1),
				},
			},
			// ---- Locks in call-event continuity: comments and blank lines do not reset ----
			{
				Code: `afterAll(() => {});

// comment-only gap

beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 5, 1),
				},
			},
			// ---- Locks in barrier semantics: external non-hook call resets ----
			{
				Code: `afterAll(() => {});
doSomething();
afterEach(() => {});
beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterEach", 4, 1),
				},
			},
			// ---- Locks in barrier semantics: invalid core member chain resets ----
			{
				Code: `import { test } from "@rstest/core";
afterAll(() => {});
test.beforeEach(() => {});
afterEach(() => {});
beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterEach", 5, 1),
				},
			},
			// ---- Locks in barrier semantics: invalid Playwright member chains reset ----
			{
				Code: `import { test } from "@rstest/playwright";
afterAll(() => {});
test.skip.beforeEach(() => {});
afterEach(() => {});
beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterEach", 5, 1),
				},
			},
			{
				Code: `import { test } from "@rstest/playwright";
afterAll(() => {});
test.beforeEach.only(() => {});
afterEach(() => {});
beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterEach", 5, 1),
				},
			},
			// ---- Locks in barrier semantics: execution-time APIs are barriers ----
			{
				Code: `afterAll(() => {});
onTestFinished(() => {});
afterEach(() => {});
beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterEach", 4, 1),
				},
			},
			{
				Code: `afterAll(() => {});
aroundEach(() => {});
afterEach(() => {});
beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterEach", 4, 1),
				},
			},
			// ---- Locks in shared frame fix: nested hook callbacks do not end the outer run ----
			{
				Code: `afterAll(() => {
  beforeEach(() => {});
  doSomething();
});
beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 5, 1),
				},
			},
			// ---- Locks in shared frame fix: a nested run is checked on its own terms ----
			{
				Code: `afterAll(() => {
  beforeEach(() => {});
  afterEach(() => {});
});
beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 5, 1),
				},
			},
			// ---- Dimension 4: optional import.meta source still reports with exact location ----
			{
				Code: `import.meta.rstest?.afterAll(() => {});
import.meta.rstest?.beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 2, 1),
				},
			},
		},
	)
}
