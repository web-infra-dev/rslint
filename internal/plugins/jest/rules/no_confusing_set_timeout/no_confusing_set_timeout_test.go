package no_confusing_set_timeout_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_confusing_set_timeout"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConfusingSetTimeoutRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_confusing_set_timeout.NoConfusingSetTimeoutRule,
		[]rule_tester.ValidTestCase{
			{Code: `
      jest.setTimeout(1000);
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
    `},
			{Code: `
      jest.setTimeout(1000);
      window.setTimeout(6000)
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('test foo', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
    `},
			{Code: `
        import { handler } from 'dep/mod';
        jest.setTimeout(800);
        describe('A', () => {
          beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        });
      `},
			{Code: `
      function handler() {}
      jest.setTimeout(800);
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
    `},
			{Code: `
      const { handler } = require('dep/mod');
      jest.setTimeout(800);
      describe('A', () => {
        beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
      });
    `},
			{Code: `
      jest.setTimeout(1000);
      window.setTimeout(60000);
    `},
			{Code: `window.setTimeout(60000);`},
			{Code: `setTimeout(1000);`},
			{Code: `
      jest.setTimeout(1000);
      test('test case', () => {
        setTimeout(() => {
          Promise.resolv();
        }, 5000);
      });
    `},
			{Code: `
      test('test case', () => {
        setTimeout(() => {
          Promise.resolv();
        }, 5000);
      });
    `},
			{Code: `
      jest['setTimeout'](1000);
      describe('A', () => {
        it('A.1', () => {});
      });
    `},
			{Code: `for (var i = 0; i < 1; i++) jest.setTimeout(1000);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
        jest.setTimeout(1000);
        setTimeout(1000);
        window.setTimeout(1000);
        describe('A', () => {
          beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        });
        jest.setTimeout(800);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "orderSetTimeout", Line: 10, Column: 9},
					{MessageId: "multipleSetTimeouts", Line: 10, Column: 9},
				},
			},
			{
				Code: `
        describe('A', () => {
          jest.setTimeout(800);
          beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalSetTimeout", Line: 3, Column: 11},
					{MessageId: "orderSetTimeout", Line: 3, Column: 11},
				},
			},
			{
				Code: `
        describe('B', () => {
          it('B.1', async () => {
            await new Promise((resolve) => {
              jest.setTimeout(1000);
              setTimeout(resolve, 10000).unref();
            });
          });
          it('B.2', async () => {
            await new Promise((resolve) => { setTimeout(resolve, 10000).unref(); });
          });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalSetTimeout", Line: 5, Column: 15},
					{MessageId: "orderSetTimeout", Line: 5, Column: 15},
				},
			},
			{
				Code: `
        test('test-suite', () => {
          jest.setTimeout(1000);
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalSetTimeout", Line: 3, Column: 11},
					{MessageId: "orderSetTimeout", Line: 3, Column: 11},
				},
			},
			{
				Code: `
        describe('A', () => {
          beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        });
        jest.setTimeout(1000);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "orderSetTimeout", Line: 7, Column: 9},
				},
			},
			{
				Code: `
        import { jest } from '@jest/globals';
        {
          jest.setTimeout(800);
        }
        describe('A', () => {
          beforeEach(async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.1', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
          it('A.2', async () => { await new Promise(resolve => { setTimeout(resolve, 10000).unref(); });});
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalSetTimeout", Line: 4, Column: 11},
				},
			},
			{
				Code: `
        jest.setTimeout(800);
        jest.setTimeout(900);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleSetTimeouts", Line: 3, Column: 9},
				},
			},
			{
				Code: `
        expect(1 + 2).toEqual(3);
        jest.setTimeout(800);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "orderSetTimeout", Line: 3, Column: 9},
				},
			},
			{
				Code: `
        import { jest as Jest } from '@jest/globals';
        {
          Jest.setTimeout(800);
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalSetTimeout", Line: 4, Column: 11},
				},
			},
			{
				Code: `
        namespace JestScope {
          jest.setTimeout(800);
        }
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalSetTimeout", Line: 3, Column: 11},
				},
			},
			{
				Code: "for (let i = 0; i < 1; i++) {\n  jest.setTimeout(1000);\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalSetTimeout", Line: 2, Column: 3},
				},
			},
			{
				Code: "for (const value of values) jest.setTimeout(1000);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalSetTimeout", Line: 1, Column: 29},
				},
			},
			{
				Code: "switch (value) {\n  case 1:\n    jest.setTimeout(1000);\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalSetTimeout", Line: 3, Column: 5},
				},
			},
			{
				Code: "try {} catch ({ value = jest.setTimeout(1000) }) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalSetTimeout", Line: 1, Column: 25},
				},
			},
			{
				Code: "enum Timeout {\n  Value = jest.setTimeout(1000),\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalSetTimeout", Line: 2, Column: 11},
				},
			},
		},
	)
}
