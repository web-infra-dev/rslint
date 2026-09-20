package prefer_hooks_on_top_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_hooks_on_top"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferHooksOnTopExtras locks in the Rstest sources, call shapes and scope
// branches the migrated suite does not reach. Each case carries an inline
// comment for the behavior it protects so a future refactor cannot silently
// regress it.
func TestPreferHooksOnTopExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_hooks_on_top.PreferHooksOnTopRule,
		[]rule_tester.ValidTestCase{
			// ---- Every Rstest source resolves to the same hook and test APIs ----
			{Code: `import { beforeEach, test } from "@rstest/core";
beforeEach(() => {});
test("charges the card", () => {});`},
			{Code: `import { beforeEach as setup, test as scenario } from "@rstest/core";
setup(() => {});
scenario("charges the card", () => {});`},
			{Code: `import * as rstest from "@rstest/core";
rstest.beforeEach(() => {});
rstest.it("charges the card", () => {});`},
			{Code: `const { beforeEach, test } = require("@rstest/core");
beforeEach(() => {});
test("charges the card", () => {});`},
			{Code: `import.meta.rstest.beforeEach(() => {});
import.meta.rstest.it("charges the card", () => {});`},
			{Code: `import { beforeEach, test } from "rstack/test";
beforeEach(() => {});
test("charges the card", () => {});`},
			// ---- An extended test API is built, not registered, so it does not
			// close the scope for the hooks that follow ----
			{Code: `import { test } from "@rstest/core";
test.extend({ db: openDatabase });
beforeEach(() => {});`},
			{Code: `import { test } from "@rstest/core";
const dbTest = test.extend({ db: openDatabase });
beforeEach(() => {});
dbTest("charges the card", ({ db }) => {});`},
			// ---- Hooks written inside a test body belong to that body, not to
			// the suite the test was registered in ----
			{Code: `describe("checkout", () => {
  test("charges the card", () => {
    beforeEach(() => {});
  });
});`},
			// ---- onTestFinished and onTestFailed run inside a test, so they are
			// not suite hooks and never report ----
			{Code: `describe("checkout", () => {
  test("charges the card", () => {});
  onTestFinished(() => {});
  onTestFailed(() => {});
});`},
			// ---- Names that do not resolve to Rstest are neither tests nor hooks ----
			{Code: `import { beforeEach } from "vitest";
test("charges the card", () => {});
beforeEach(() => {});`},
			{Code: `const beforeEach = createHook();
test("charges the card", () => {});
beforeEach(() => {});`},
			{Code: `import { test } from "@rstest/core";
test("charges the card", () => {});
test.beforeEach(() => {});`},
			// ---- Playwright exposes the hooks as members of the test object ----
			{Code: `import { test } from "@rstest/playwright";
test.beforeEach(async ({ page }) => {});
test("charges the card", async ({ page }) => {});`},
			{Code: `import { test } from "@rstest/playwright";
const appTest = test.extend({});
appTest.beforeEach(async ({ page }) => {});
appTest("charges the card", async ({ page }) => {});`},
			// ---- Each suite body is scored on its own, including parameterized ones ----
			{Code: `describe.each([1, 2])("checkout %i", (attempt) => {
  beforeEach(() => {});
  test("charges the card", () => {});
});`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- The file body is a scope of its own ----
			{
				Code: `test("charges the card", () => {});
beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 2, Column: 1},
				},
			},
			// ---- An invoked extend() chain IS a registration: the member name
			// alone must not exempt it ----
			{
				Code: `import { test } from "@rstest/core";
describe("checkout", () => {
  test.extend({ db: openDatabase })("charges the card", ({ db }) => {});
  beforeEach(() => {});
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 4, Column: 3},
				},
			},
			// ---- A registration through an extended alias closes the scope too ----
			{
				Code: `import { test } from "@rstest/core";
const dbTest = test.extend({ db: openDatabase });
describe("checkout", () => {
  dbTest("charges the card", ({ db }) => {});
  beforeEach(() => {});
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 5, Column: 3},
				},
			},
			// ---- Renamed, namespace, CommonJS and import.meta sources all report ----
			{
				Code: `import { beforeEach as setup, test as scenario } from "@rstest/core";
scenario("charges the card", () => {});
setup(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 3, Column: 1},
				},
			},
			{
				Code: `import * as rstest from "@rstest/core";
rstest.it("charges the card", () => {});
rstest.beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 3, Column: 1},
				},
			},
			{
				Code: `const { beforeEach, test } = require("@rstest/core");
test("charges the card", () => {});
beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 3, Column: 1},
				},
			},
			{
				Code: `import.meta.rstest.it("charges the card", () => {});
import.meta.rstest.beforeEach(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 2, Column: 1},
				},
			},
			// ---- Playwright member hooks report against Playwright registrations ----
			{
				Code: `import { test } from "@rstest/playwright";
test("charges the card", async ({ page }) => {});
test.beforeEach(async ({ page }) => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 3, Column: 1},
				},
			},
			// ---- .for and .each rows are registrations ----
			{
				Code: `describe("checkout", () => {
  test.for([1, 2])("charges the card %i", (attempt) => {});
  beforeEach(() => {});
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 3, Column: 3},
				},
			},
			// ---- A parameterized suite body is scored like any other ----
			{
				Code: `describe.each([1, 2])("checkout %i", (attempt) => {
  test("charges the card", () => {});
  beforeEach(() => {});
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 3, Column: 3},
				},
			},
		},
	)
}
