package no_done_callback_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_done_callback"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDoneCallbackRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_done_callback.NoDoneCallbackRule,
		[]rule_tester.ValidTestCase{
			{Code: `test("something", () => {})`},
			{Code: `test("something", async () => {})`},
			{Code: `test("something", function() {})`},
			{Code: "test.each``(\"something\", ({ a, b }) => {})"},
			{Code: `test.each()("something", ({ a, b }) => {})`},
			{Code: `it.each()("something", ({ a, b }) => {})`},
			{Code: `it.each([])("something", (a, b) => {})`},
			{Code: "it.each``(\"something\", ({ a, b }) => {})"},
			{Code: `it.each([])("something", (a, b) => { a(); b(); })`},
			{Code: "it.each``(\"something\", ({ a, b }) => { a(); b(); })"},
			{Code: `test("something", async function () {})`},
			{Code: `test("something", someArg)`},
			{Code: `beforeEach(() => {})`},
			{Code: `beforeAll(async () => {})`},
			{Code: `afterAll(() => {})`},
			{Code: `afterAll(async function () {})`},
			{Code: `afterAll(async function () {}, 5)`},
			{Code: "test.each``(\"only one arg\")"},
			{Code: "it.each``(\"only one arg\")"},
			{Code: `test.each([])("only one arg")`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `test("something", (...args) => {args[0]();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    20,
					},
				},
			},
			{
				Code: `test("something", done => {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `test("something", () => {return new Promise(done => {done();})})`,
							},
						},
					},
				},
			},
			{
				Code: `test("something", (done,) => {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `test("something", () => {return new Promise(done => {done();})})`,
							},
						},
					},
				},
			},
			{
				Code: `test("something", finished => {finished();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `test("something", () => {return new Promise(finished => {finished();})})`,
							},
						},
					},
				},
			},
			{
				Code: `test("something", (done) => {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `test("something", () => {return new Promise(done => {done();})})`,
							},
						},
					},
				},
			},
			{
				Code: `test("something", done => done())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `test("something", () => new Promise(done => done()))`,
							},
						},
					},
				},
			},
			{
				Code: `test("something", (done) => done())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `test("something", () => new Promise(done => done()))`,
							},
						},
					},
				},
			},
			{
				Code: `test("something", function(done) {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `test("something", function() {return new Promise(done => {done();})})`,
							},
						},
					},
				},
			},
			{
				Code: `test("something", function (done) {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    29,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `test("something", function () {return new Promise(done => {done();})})`,
							},
						},
					},
				},
			},
			{
				Code: `test("something", async done => {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useAwaitInsteadOfCallback", Line: 1, Column: 25},
				},
			},
			{
				Code: `test("something", async done => done())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useAwaitInsteadOfCallback", Line: 1, Column: 25},
				},
			},
			{
				Code: `test("something", async function (done) {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useAwaitInsteadOfCallback", Line: 1, Column: 35},
				},
			},
			{
				Code: `
        test('my test', async (done) => {
          await myAsyncTask();
          expect(true).toBe(false);
          done();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useAwaitInsteadOfCallback", Line: 2, Column: 32},
				},
			},
			{
				Code: `
        test('something', (done) => {
          done();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      2,
						Column:    28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output: `
        test('something', () => {return new Promise(done => {
          done();
        })});
      `,
							},
						},
					},
				},
			},
			{
				Code: `afterEach((...args) => {args[0]();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    12,
					},
				},
			},
			{
				Code: `beforeAll(done => {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `beforeAll(() => {return new Promise(done => {done();})})`,
							},
						},
					},
				},
			},
			{
				Code: `beforeAll(finished => {finished();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `beforeAll(() => {return new Promise(finished => {finished();})})`,
							},
						},
					},
				},
			},
			{
				Code: `beforeEach((done) => {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `beforeEach(() => {return new Promise(done => {done();})})`,
							},
						},
					},
				},
			},
			{
				Code: `afterAll(done => done())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `afterAll(() => new Promise(done => done()))`,
							},
						},
					},
				},
			},
			{
				Code: `afterEach((done) => done())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    12,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `afterEach(() => new Promise(done => done()))`,
							},
						},
					},
				},
			},
			{
				Code: `beforeAll(function(done) {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `beforeAll(function() {return new Promise(done => {done();})})`,
							},
						},
					},
				},
			},
			{
				Code: `afterEach(function (done) {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    `afterEach(function () {return new Promise(done => {done();})})`,
							},
						},
					},
				},
			},
			{
				Code: `beforeAll(async done => {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useAwaitInsteadOfCallback", Line: 1, Column: 17},
				},
			},
			{
				Code: `beforeAll(async done => done())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useAwaitInsteadOfCallback", Line: 1, Column: 17},
				},
			},
			{
				Code: `beforeAll(async function (done) {done();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useAwaitInsteadOfCallback", Line: 1, Column: 27},
				},
			},
			{
				Code: `
        afterAll(async (done) => {
          await myAsyncTask();
          done();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useAwaitInsteadOfCallback", Line: 2, Column: 25},
				},
			},
			{
				Code: `
        beforeEach((done) => {
          done();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      2,
						Column:    21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output: `
        beforeEach(() => {return new Promise(done => {
          done();
        })});
      `,
							},
						},
					},
				},
			},
			{
				Code: `
        import { beforeEach } from '@jest/globals';

        beforeEach((done) => {
          done();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      4,
						Column:    21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output: `
        import { beforeEach } from '@jest/globals';

        beforeEach(() => {return new Promise(done => {
          done();
        })});
      `,
							},
						},
					},
				},
			},
			{
				Code: `
        import { beforeEach as atTheStartOfEachTest } from '@jest/globals';

        atTheStartOfEachTest((done) => {
          done();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      4,
						Column:    31,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output: `
        import { beforeEach as atTheStartOfEachTest } from '@jest/globals';

        atTheStartOfEachTest(() => {return new Promise(done => {
          done();
        })});
      `,
							},
						},
					},
				},
			},
			{
				Code: "test.each``(\"something\", ({ a, b }, done) => { done(); })",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    37,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    "test.each``(\"something\", () => {return new Promise(done => { done(); })})",
							},
						},
					},
				},
			},
			{
				Code: "it.each``(\"something\", ({ a, b }, done) => { done(); })",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDoneCallback",
						Line:      1,
						Column:    35,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestWrappingInPromise",
								Output:    "it.each``(\"something\", () => {return new Promise(done => { done(); })})",
							},
						},
					},
				},
			},
		},
	)
}
