package no_standalone_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_standalone_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoStandaloneExpectRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_standalone_expect.NoStandaloneExpectRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect.any(String)`},
			{Code: `expect.extend({})`},
			{Code: `expect.not.stringContaining("value")`},
			{Code: `describe("a test", () => { it("an it", () => {expect(1).toBe(1); }); });`},
			{Code: `describe("a test", () => { it("an it", () => { const func = () => { expect(1).toBe(1); }; }); });`},
			{Code: `describe("a test", () => { const func = () => { expect(1).toBe(1); }; });`},
			{Code: `describe("a test", () => { function func() { expect(1).toBe(1); }; });`},
			{Code: `describe("a test", () => { const func = function(){ expect(1).toBe(1); }; });`},
			{Code: `it("an it", () => expect(1).toBe(1))`},
			{Code: `const func = function(){ expect(1).toBe(1); };`},
			{Code: `const func = () => expect(1).toBe(1);`},
			{Code: `{}`},
			{Code: `it.each([1, true])("trues", value => { expect(value).toBe(true); });`},
			{Code: `it.each([1, true])("trues", value => { expect(value).toBe(true); }); it("an it", () => { expect(1).toBe(1) });`},
			{Code: `
      it.each` + "`" + `
        num   | value
        ${1} | ${true}
      ` + "`" + `('trues', ({ value }) => {
        expect(value).toBe(true);
      });
    `},
			{Code: `it.only("an only", value => { expect(value).toBe(true); });`},
			{Code: `it.concurrent("an concurrent", value => { expect(value).toBe(true); });`},
			{Code: `describe.each([1, true])("trues", value => { it("an it", () => expect(value).toBe(true) ); });`},
			{Code: `const helpers = { assert() { expect(1).toBe(1); } };`},
			{Code: `class Helper { assert() { expect(1).toBe(1); } get value() { expect(1).toBe(1); return 1; } }`},
			{
				Code: `
        describe('scenario', () => {
          const t = Math.random() ? it.only : it;
          t('testing', () => expect(true));
        });
      `,
				Options: []interface{}{
					map[string]interface{}{"additionalTestBlockFunctions": []interface{}{"t"}},
				},
			},
			{
				Code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
				Options: []interface{}{
					map[string]interface{}{"additionalTestBlockFunctions": []interface{}{"each.test"}},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "(() => {})('testing', () => expect(true).toBe(false))",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 29, EndColumn: 53},
				},
			},
			{
				Code: `expect.hasAssertions()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 1, EndColumn: 23},
				},
			},
			{
				Code: `expect().hasAssertions()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 1, EndColumn: 25},
				},
			},
			{
				Code: `
        describe('scenario', () => {
          const t = Math.random() ? it.only : it;
          t('testing', () => expect(true).toBe(false));
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 4, Column: 30, EndColumn: 54},
				},
			},
			{
				Code: `
        describe('scenario', () => {
          const t = Math.random() ? it.only : it;
          t('testing', () => expect(true).toBe(false));
        });
      `,
				Options: []interface{}{
					map[string]interface{}{"additionalTestBlockFunctions": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 4, Column: 30, EndColumn: 54},
				},
			},
			{
				Code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 7, Column: 11, EndColumn: 39},
				},
			},
			{
				Code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
				Options: []interface{}{
					map[string]interface{}{"additionalTestBlockFunctions": []interface{}{"each"}},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 7, Column: 11, EndColumn: 39},
				},
			},
			{
				Code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
				Options: []interface{}{
					map[string]interface{}{"additionalTestBlockFunctions": []interface{}{"test"}},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 7, Column: 11, EndColumn: 39},
				},
			},
			{
				Code: `describe("a test", () => { expect(1).toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 28, EndColumn: 45},
				},
			},
			{
				Code: `describe("a test", () => expect(1).toBe(1));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 26, EndColumn: 43},
				},
			},
			{
				Code: `describe("a test", () => { const func = () => { expect(1).toBe(1); }; expect(1).toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 71, EndColumn: 88},
				},
			},
			{
				Code: `describe("a test", () => {  it(() => { expect(1).toBe(1); }); expect(1).toBe(1); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 63, EndColumn: 80},
				},
			},
			{
				Code: `expect(1).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 1, EndColumn: 18},
				},
			},
			{
				Code: `{expect(1).toBe(1)}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 2, EndColumn: 19},
				},
			},
			{
				Code: `it.each([1, true])("trues", value => { expect(value).toBe(true); }); expect(1).toBe(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 70, EndColumn: 87},
				},
			},
			{
				Code: `it.only("an only", () => { expect(1).toBe(1); }); expect(2).toBe(2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 51, EndColumn: 68},
				},
			},
			{
				Code: `describe.each([1, true])("trues", value => { expect(value).toBe(true); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 1, Column: 46, EndColumn: 70},
				},
			},
			{
				Code: `
        describe.each` + "`" + `
          num   | value
          ${1} | ${true}
        ` + "`" + `('trues', ({ value }) => {
          expect(value).toBe(true);
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 6, Column: 11, EndColumn: 35},
				},
			},
			{
				Code: `
        it.each` + "`" + `
          num   | value
          ${1} | ${true}
        ` + "`" + `('trues', ({ value }) => {
          expect(value).toBe(true);
        });

        expect(1).toBe(1);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 9, Column: 9, EndColumn: 26},
				},
			},
			{
				Code: `
        import { expect as pleaseExpect } from '@jest/globals';

        describe("a test", () => { pleaseExpect(1).toBe(1); });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 4, Column: 36, EndColumn: 59},
				},
			},
			{
				Code: `
        import { expect as pleaseExpect } from '@jest/globals';

        beforeEach(() => pleaseExpect.hasAssertions());
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExpect", Line: 4, Column: 26, EndColumn: 54},
				},
			},
		},
	)
}
