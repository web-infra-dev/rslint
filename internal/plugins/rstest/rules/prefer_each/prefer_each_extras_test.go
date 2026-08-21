package prefer_each_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_each"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferEachExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_each.PreferEachRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: loop kinds without registration ----
			{Code: `for (const row of rows) { consume(row); }`},
			{Code: `for (const key in rows) { consume(rows[key]); }`},
			{Code: `for (let i = 0; i < rows.length; i++) { consume(rows[i]); }`},

			// ---- Real-user: #1048 business loop inside a test callback ----
			{Code: `test('business loop', () => { for (const row of rows) { consume(row); } });`},
			// ---- Real-user: nested-test business loop inside a callback ----
			{Code: `test('case', () => { for (const row of rows) { test(row.name, () => {}); } });`},

			// ---- test callback exclusion ----
			{Code: `test('outer', () => { test('inner', () => {}); for (const row of rows) { consume(row); } });`},

			// ---- overload boundaries ----
			{Code: `test('case', () => { for (const row of rows) { consume(row); } }, 1000);`},
			{Code: `test('case', { timeout: 1000 }, () => { for (const row of rows) { consume(row); } });`},
			{Code: `function body() { for (const row of rows) { consume(row); } } test('case', body);`},
			{Code: `const body = () => { for (const row of rows) { consume(row); } }; test('case', body);`},

			// ---- parser lock: factories are not final registrations ----
			{Code: `for (const row of rows) { test.each(rows); test.for(rows); test.runIf(flag); test.skipIf(flag); test.extend({}); }`},
			{Code: `for (const row of rows) { const makeEach = test.each(rows); const makeFor = test.for(rows); consume(makeEach, makeFor); }`},
			{Code: "for (const row of rows) { test.each(rows)`case ${row.name}`(() => {}); }"},

			// ---- non-rstest negatives ----
			{Code: `import { test } from 'vitest'; for (const row of rows) { test(row.name, () => {}); }`},
			{Code: `import { test } from '@jest/globals'; for (const row of rows) { test(row.name, () => {}); }`},
			{Code: `import { test } from '@playwright/test'; for (const row of rows) { test(row.name, () => {}); }`},
			{Code: `declare function createRunner(): any; const test = createRunner(); for (const row of rows) { test(row.name, () => {}); }`},

			// ---- Dimension 4: wrappers / empty loops / same-kind nesting ----
			{Code: `for (const row of rows);`},
			{Code: `for (const row of rows) ((consume))(row);`},
			// Locks in existing Rslint/Jest behavior: the nested loop enter clears
			// the outer pending sequence, so neither loop reports.
			{Code: `for (const suite of suites) {
  test(suite.name, () => {});
  for (const item of suite.items) {
    consume(item);
  }
}`},
			// N/A: optional call on the registration itself is not a supported rstest parser shape.
			// N/A: computed dynamic registration member names are intentionally unresolved by ParseFnCall.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Real-user: #1048 top-level batch registrations ----
			{
				Code: `for (const scenario of scenarios) {
  test(scenario.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      1,
					Column:    1,
					EndLine:   3,
					EndColumn: 2,
				}},
			},
			// ---- Real-user: #1048 describe-scoped batch registrations ----
			{
				Code: `describe('suite', () => {
  for (const scenario of scenarios) {
    test(scenario.name, () => {});
  }
});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      2,
					Column:    3,
					EndLine:   4,
					EndColumn: 4,
				}},
			},

			// ---- all rstest source variants ----
			{
				Code: `import { test } from '@rstest/core';
for (const row of rows) {
  test(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import { test as case_ } from '@rstest/core';
for (const row of rows) {
  case_(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `const { test } = require('@rstest/core');
for (const row of rows) {
  test(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import * as rstest from '@rstest/core';
for (const row of rows) {
  rstest.test(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `const rstest = require('@rstest/core');
for (const row of rows) {
  rstest.test(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `for (const row of rows) {
  import.meta.rstest.test(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      1,
					Column:    1,
				}},
			},
			{
				Code: `const { test: case_ } = import.meta.rstest;
for (const row of rows) {
  case_(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `const api = import.meta.rstest;
for (const row of rows) {
  api.test(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `const case_ = test;
for (const row of rows) {
  case_(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `const case_ = it;
for (const row of rows) {
  case_(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `it.each` rather than a manual loop",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import { test } from '@rstest/playwright';
for (const row of rows) {
  test(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import { test } from '@rstest/playwright';
for (const row of rows) {
  test.describe(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `describe.each` rather than a manual loop",
					Line:      2,
					Column:    1,
				}},
			},

			// ---- test / describe / hook combinations ----
			{
				Code: `for (const row of rows) {
  test(row.name, () => {});
  test(row.id, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `describe.each` rather than a manual loop", Line: 1, Column: 1}},
			},
			{
				Code: `for (const row of rows) {
  describe(row.name, () => {
    test('inner', () => {});
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `describe.each` rather than a manual loop", Line: 1, Column: 1}},
			},
			{
				Code: `for (const row of rows) {
  beforeEach(() => setup(row));
  test(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `describe.each` rather than a manual loop", Line: 1, Column: 1}},
			},
			{
				Code: `for (const row of rows) {
  beforeAll(() => setup(row));
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `describe.each` rather than a manual loop", Line: 1, Column: 1}},
			},

			// ---- Dimension 1: alternate loop kinds with registrations ----
			{
				Code: `for (const key in rows) {
  test(key, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      1,
					Column:    1,
					EndLine:   3,
					EndColumn: 2,
				}},
			},
			{
				Code: `for (let i = 0; i < rows.length; i++) {
  test(String(i), () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferEach",
					Message:   "prefer using `test.each` rather than a manual loop",
					Line:      1,
					Column:    1,
					EndLine:   3,
					EndColumn: 2,
				}},
			},

			// ---- overloads and named callbacks ----
			{
				Code: `for (const row of rows) {
  test(row.name, () => {}, 1000);
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `test.each` rather than a manual loop", Line: 1, Column: 1}},
			},
			{
				Code: `for (const row of rows) {
  test(row.name, { timeout: 1000 }, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `test.each` rather than a manual loop", Line: 1, Column: 1}},
			},
			{
				Code: `function named() {}
for (const row of rows) {
  test(row.name, named);
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `test.each` rather than a manual loop", Line: 2, Column: 1}},
			},

			// ---- parser lock: direct final registration / tagged each / factory alias ----
			{
				Code: `for (const row of rows) {
  test.each(rows)(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `test.each` rather than a manual loop", Line: 1, Column: 1}},
			},
			{
				Code:   "for (const row of rows) {\n  test.each`\nname\n${row.name}\n`(\"$name\", () => {});\n}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `test.each` rather than a manual loop", Line: 1, Column: 1}},
			},
			{
				Code: `const cases = test.each(rows);
for (const row of rows) {
  cases(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `test.each` rather than a manual loop", Line: 2, Column: 1}},
			},
			{
				Code: `const cases = test.for(rows);
for (const row of rows) {
  cases(row.name, () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `test.each` rather than a manual loop", Line: 2, Column: 1}},
			},

			// ---- nested loop contracts ----
			{
				// Locks in existing Rslint/Jest behavior: only the inner loop reports.
				Code: `for (const suite of suites) {
  for (const item of suite.items) {
    test(item.name, () => {});
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferEach", Message: "prefer using `test.each` rather than a manual loop", Line: 2, Column: 3}},
			},
		},
	)
}
