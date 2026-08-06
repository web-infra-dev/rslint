package prefer_expect_assertions_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_expect_assertions"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferExpectAssertionsUpstream migrates the full valid/invalid suite from upstream
// packages/rslint-test-tools/tests/eslint-plugin-jest/rules/prefer-expect-assertions.test.ts 1:1. Position assertions cover line/column for every invalid
// case. Rule implementation is pending — the test is skipped until then.
func TestPreferExpectAssertionsUpstream(t *testing.T) {
	t.Skip("rule not implemented yet")
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_expect_assertions.PreferExpectAssertionsRule,
		[]rule_tester.ValidTestCase{
			{Code: `test("nonsense", [])`},
			{Code: `test("it1", () => {expect.assertions(0);})`},
			{Code: `test("it1", function() {expect.assertions(0);})`},
			{Code: `test("it1", function() {expect.hasAssertions();})`},
			{Code: `it("it1", function() {expect.assertions(0);})`},
			{Code: `
      it("it1", function() {
        expect.assertions(1);
        expect(someValue).toBe(true)
      });
    `},
			{Code: `test("it1")`},
			{Code: `itHappensToStartWithIt("foo", function() {})`},
			{Code: `testSomething("bar", function() {})`},
			{Code: `it(async () => {expect.assertions(0);})`},
			{Code: `
      it("returns numbers that are greater than four", function() {
        expect.assertions(2);

        for(let thing in things) {
          expect(number).toBeGreaterThan(4);
        }
      });
    `},
			{Code: `
      it("returns numbers that are greater than four", function() {
        expect.hasAssertions();

        for (let i = 0; i < things.length; i++) {
          expect(number).toBeGreaterThan(4);
        }
      });
    `},
			{Code: `
        it("it1", async () => {
          expect.assertions(1);
          expect(someValue).toBe(true)
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}}},
			{Code: `
        it("it1", function() {
          expect(someValue).toBe(true)
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}}},
			{Code: `it("it1", () => {})`, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}}},
			{Code: `
        it("returns numbers that are greater than four", async () => {
          expect.assertions(2);

          for(let thing in things) {
            expect(number).toBeGreaterThan(4);
          }
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}}},
			{Code: `
        it("returns numbers that are greater than four", () => {
          for(let thing in things) {
            expect(number).toBeGreaterThan(4);
          }
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}}},
			{Code: `
        import { expect as pleaseExpect } from '@jest/globals';

        it("returns numbers that are greater than four", function() {
          pleaseExpect.assertions(2);

          for(let thing in things) {
            pleaseExpect(number).toBeGreaterThan(4);
          }
        });
      `},
			{Code: `
      beforeEach(() => expect.hasAssertions());

      it('responds ok', function () {
        client.get('/user', response => {
          expect(response.status).toBe(200);
        });
      });

      it("is a number that is greater than four", () => {
        expect(number).toBeGreaterThan(4);
      });
    `},
			{Code: `
      afterEach(() => {
        expect.hasAssertions();
      });

      it('responds ok', function () {
        client.get('/user', response => {
          expect(response.status).toBe(200);
        });
      });

      it("is a number that is greater than four", () => {
        expect(number).toBeGreaterThan(4);
      });
    `},
			{Code: `
      afterEach(() => {
        expect.hasAssertions();
      });

      it('responds ok', function () {
        client.get('/user', response => {
          expect(response.status).toBe(200);
        });
      });

      it("is a number that is greater than four", () => {
        expect.hasAssertions();

        expect(number).toBeGreaterThan(4);
      });
    `},
			{Code: `
      beforeEach(() => { expect.hasAssertions(); });

      describe('my tests', () => {
        it('responds ok', function () {
          client.get('/user', response => {
            expect(response.status).toBe(200);
          });
        });

        it("is a number that is greater than four", () => {
          expect.hasAssertions();

          expect(number).toBeGreaterThan(4);
        });
      });
    `},
			{Code: `
      describe('my tests', () => {
        beforeEach(() => { expect.hasAssertions(); });

        describe('left', () => {
          describe('inner', () => {
            it('responds ok', function () {
              client.get('/user', response => {
                expect(response.status).toBe(200);
              });
            });
          });
        });

        describe('right', () => {
          it("is a number that is greater than four", () => {
            expect(number).toBeGreaterThan(4);
          });
        });
      });
    `},
			{Code: `
      describe('my tests', () => {
        beforeEach(() => { expect.hasAssertions(); });

        describe('left', () => {
          it('responds ok', function () {
            client.get('/user', response => {
              expect(response.status).toBe(200);
            });
          });
        });

        describe('right', () => {
          it("is a number that is greater than four", () => {
            expect(number).toBeGreaterThan(4);
          });
        });
      });
    `},
			{Code: `
      describe('my tests', () => {
        beforeEach(() => { expect.hasAssertions(); });

        describe('left', () => {
          beforeEach(() => { expect.hasAssertions(); });

          it('responds ok', function () {
            client.get('/user', response => {
              expect(response.status).toBe(200);
            });
          });
        });

        describe('right', () => {
          it("is a number that is greater than four", () => {
            expect(number).toBeGreaterThan(4);
          });
        });
      });
    `},
			{Code: `
      describe('my tests', () => {
        beforeEach(() => { expect.hasAssertions(); });

        describe('left', () => {
          afterEach(() => { expect.hasAssertions(); });

          it('responds ok', function () {
            client.get('/user', response => {
              expect(response.status).toBe(200);
            });
          });
        });

        describe('right', () => {
          it("is a number that is greater than four", () => {
            expect(number).toBeGreaterThan(4);
          });
        });
      });
    `},
			{Code: `
      describe('my tests', () => {
        beforeEach(() => { expect.hasAssertions(); });

        it('responds ok', function () {
          client.get('/user', response => {
            expect(response.status).toBe(200);
          });
        });

        it("is a number that is greater than four", () => {
          expect.hasAssertions();

          expect(number).toBeGreaterThan(4);
        });
      });
    `},
			{Code: `
      beforeEach(() => {
        setTimeout(() => expect.hasAssertions(), 5000);
      });

      it('only returns numbers that are greater than six', () => {
        for (const number of getNumbers()) {
          expect(number).toBeGreaterThan(6);
        }
      });
    `},
			{Code: `
        const expectNumbersToBeGreaterThan = (numbers, value) => {
          for (let number of numbers) {
            expect(number).toBeGreaterThan(value);
          }
        };

        it('returns numbers that are greater than two', function () {
          expectNumbersToBeGreaterThan(getNumbers(), 2);
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}}},
			{Code: `
        it("returns numbers that are greater than five", function () {
          expect.assertions(2);

          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(5);
          }
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}}},
			{Code: `
        it("returns things that are less than ten", function () {
          expect.hasAssertions();

          for (const thing in things) {
            expect(thing).toBeLessThan(10);
          }
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}}},
			{Code: `
        const expectNumbersToBeGreaterThan = (numbers, value) => {
          numbers.forEach(number => {
            expect(number).toBeGreaterThan(value);
          });
        };

        it('returns numbers that are greater than two', function () {
          expectNumbersToBeGreaterThan(getNumbers(), 2);
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}}},
			{Code: `
        it('returns numbers that are greater than two', function () {
          expect.assertions(2);

          const expectNumbersToBeGreaterThan = (numbers, value) => {
            for (let number of numbers) {
              expect(number).toBeGreaterThan(value);
            }
          };

          expectNumbersToBeGreaterThan(getNumbers(), 2);
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}}},
			{Code: `
        beforeEach(() => expect.hasAssertions());

        it('returns numbers that are greater than two', function () {
          const expectNumbersToBeGreaterThan = (numbers, value) => {
            for (let number of numbers) {
              expect(number).toBeGreaterThan(value);
            }
          };

          expectNumbersToBeGreaterThan(getNumbers(), 2);
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}}},
			{Code: `
        it("returns numbers that are greater than five", function () {
          expect.assertions(2);

          getNumbers().forEach(number => {
            expect(number).toBeGreaterThan(5);
          });
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}}},
			{Code: `
        it("returns things that are less than ten", function () {
          expect.hasAssertions();

          things.forEach(thing => {
            expect(thing).toBeLessThan(10);
          });
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}}},
			{Code: `
        it('sends the data as a string', () => {
          expect.hasAssertions();

          const stream = openStream();

          stream.on('data', data => {
            expect(data).toBe(expect.any(String));
          });
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}}},
			{Code: `
        it('responds ok', function () {
          expect.assertions(1);

          client.get('/user', response => {
            expect(response.status).toBe(200);
          });
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}}},
			{Code: "\n        it.each([1, 2, 3])(\"returns ok\", id => {\n          expect.assertions(3);\n\n          client.get(`/users/${id}`, response => {\n            expect(response.status).toBe(200);\n          });\n        });\n\n        it(\"is a number that is greater than four\", () => {\n          expect(number).toBeGreaterThan(4);\n        });\n      ", Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}}},
			{Code: `
        it('is a test', () => {
          expect(expected).toBe(actual);
        });

        describe('my test', () => {
          it('is another test', () => {
            expect(expected).toBe(actual);
          });
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}}},
			{Code: `
        it('responds ok', function () {
          expect.assertions(1);

          client.get('/user', response => {
            expect(response.status).toBe(200);
          });
        });

        describe('my test', () => {
          beforeEach(() => expect.hasAssertions());

          it('responds ok', function () {
            client.get('/user', response => {
              expect(response.status).toBe(200);
            });
          });
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}}},
			{Code: `
        it('responds ok', function () {
          expect.assertions(1);

          client.get('/user', response => {
            expect(response.status).toBe(200);
          });
        });

        describe('my test', () => {
          afterEach(() => expect.hasAssertions());

          it('responds ok', function () {
            client.get('/user', response => {
              expect(response.status).toBe(200);
            });
          });
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}}},
			{Code: `
        it('only returns numbers that are greater than zero', async () => {
          expect.hasAssertions();

          for (const number of await getNumbers()) {
            expect(number).toBeGreaterThan(0);
          }
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true, `onlyFunctionsWithExpectInLoop`: true}}},
			{Code: `
        it('only returns numbers that are greater than zero', async () => {
          expect.assertions(2);

          for (const number of await getNumbers()) {
            expect(number).toBeGreaterThan(0);
          }
        });
      `, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true, `onlyFunctionsWithExpectInLoop`: true}}},
			{Code: `test.each()("is fine", () => { expect.assertions(0); })`},
			{Code: "test.each``(\"is fine\", () => { expect.assertions(0); })"},
			{Code: `test.each()("is fine", () => { expect.hasAssertions(); })`},
			{Code: "test.each``(\"is fine\", () => { expect.hasAssertions(); })"},
			{Code: `it.each()("is fine", () => { expect.assertions(0); })`},
			{Code: "it.each``(\"is fine\", () => { expect.assertions(0); })"},
			{Code: `it.each()("is fine", () => { expect.hasAssertions(); })`},
			{Code: "it.each``(\"is fine\", () => { expect.hasAssertions(); })"},
			{Code: `test.each()("is fine", () => {})`, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}}},
			{Code: "test.each``(\"is fine\", () => {})", Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}}},
			{Code: `it.each()("is fine", () => {})`, Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}}},
			{Code: "it.each``(\"is fine\", () => {})", Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}}},
			{Code: `
      describe.each(['hello'])('%s', () => {
        it('is fine', () => {
          expect.assertions(0);
        });
      });
    `},
			{Code: "\n      describe.each``('%s', () => {\n        it('is fine', () => {\n          expect.assertions(0);\n        });\n      });\n    "},
			{Code: `
      describe.each(['hello'])('%s', () => {
        it('is fine', () => {
          expect.hasAssertions();
        });
      });
    `},
			{Code: "\n      describe.each``('%s', () => {\n        it('is fine', () => {\n          expect.hasAssertions();\n        });\n      });\n    "},
			{Code: `
      describe.each(['hello'])('%s', () => {
        it.each()('is fine', () => {
          expect.assertions(0);
        });
      });
    `},
			{Code: "\n      describe.each``('%s', () => {\n        it.each()('is fine', () => {\n          expect.assertions(0);\n        });\n      });\n    "},
			{Code: `
      describe.each(['hello'])('%s', () => {
        it.each()('is fine', () => {
          expect.hasAssertions();
        });
      });
    `},
			{Code: "\n      describe.each``('%s', () => {\n        it.each()('is fine', () => {\n          expect.hasAssertions();\n        });\n      });\n    "},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `it("it1", () => foo())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1},
				},
			},
			{
				Code: `it('resolves', () => expect(staged()).toBe(true));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1},
				},
			},
			{
				Code: `it('resolves', async () => expect(await staged()).toBe(true));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1},
				},
			},
			{
				Code: `it("it1", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `it("it1", () => {expect.hasAssertions();})`}, {MessageId: `suggestAddingAssertions`, Output: `it("it1", () => {expect.assertions();})`}}},
				},
			},
			{
				Code: `it("it1", () => { foo()})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `it("it1", () => {expect.hasAssertions(); foo()})`}, {MessageId: `suggestAddingAssertions`, Output: `it("it1", () => {expect.assertions(); foo()})`}}},
				},
			},
			{
				Code: `
        it("it1", function() {
          someFunctionToDo();
          someFunctionToDo2();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("it1", function() {expect.hasAssertions();
                  someFunctionToDo();
                  someFunctionToDo2();
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("it1", function() {expect.assertions();
                  someFunctionToDo();
                  someFunctionToDo2();
                });
              `}}},
				},
			},
			{
				Code: `
        it("it1", function() {
          someFunctionToDo();
          someFunctionToDo2();
        });

        describe('some tests', () => {
          beforeEach(() => { expect.hasAssertions(); });
          it("it1", function() {
            someFunctionToDo();
            someFunctionToDo2();
          });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("it1", function() {expect.hasAssertions();
                  someFunctionToDo();
                  someFunctionToDo2();
                });

                describe('some tests', () => {
                  beforeEach(() => { expect.hasAssertions(); });
                  it("it1", function() {
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("it1", function() {expect.assertions();
                  someFunctionToDo();
                  someFunctionToDo2();
                });

                describe('some tests', () => {
                  beforeEach(() => { expect.hasAssertions(); });
                  it("it1", function() {
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });
              `}}},
				},
			},
			{
				Code: `
        it("it1", function() {
          someFunctionToDo();
          someFunctionToDo2();
        });

        describe('some tests', () => {
          afterEach(() => { expect.hasAssertions(); });
          it("it1", function() {
            someFunctionToDo();
            someFunctionToDo2();
          });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("it1", function() {expect.hasAssertions();
                  someFunctionToDo();
                  someFunctionToDo2();
                });

                describe('some tests', () => {
                  afterEach(() => { expect.hasAssertions(); });
                  it("it1", function() {
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("it1", function() {expect.assertions();
                  someFunctionToDo();
                  someFunctionToDo2();
                });

                describe('some tests', () => {
                  afterEach(() => { expect.hasAssertions(); });
                  it("it1", function() {
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });
              `}}},
				},
			},
			{
				Code: `
        describe('some tests', () => {
          it("it1", function() {
            someFunctionToDo();
            someFunctionToDo2();
          });

          beforeEach(() => { expect.hasAssertions(); });

          it("it1", function() {
            someFunctionToDo();
            someFunctionToDo2();
          });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 2, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                describe('some tests', () => {
                  it("it1", function() {expect.hasAssertions();
                    someFunctionToDo();
                    someFunctionToDo2();
                  });

                  beforeEach(() => { expect.hasAssertions(); });

                  it("it1", function() {
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                describe('some tests', () => {
                  it("it1", function() {expect.assertions();
                    someFunctionToDo();
                    someFunctionToDo2();
                  });

                  beforeEach(() => { expect.hasAssertions(); });

                  it("it1", function() {
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });
              `}}},
				},
			},
			{
				Code: `
        describe('some tests', () => {
          beforeEach(() => { expect.hasAssertions(); });
          it("it1", function() {
            someFunctionToDo();
            someFunctionToDo2();
          });
        });

        it("it1", function() {
          someFunctionToDo();
          someFunctionToDo2();
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 9, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                describe('some tests', () => {
                  beforeEach(() => { expect.hasAssertions(); });
                  it("it1", function() {
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });

                it("it1", function() {expect.hasAssertions();
                  someFunctionToDo();
                  someFunctionToDo2();
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                describe('some tests', () => {
                  beforeEach(() => { expect.hasAssertions(); });
                  it("it1", function() {
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });

                it("it1", function() {expect.assertions();
                  someFunctionToDo();
                  someFunctionToDo2();
                });
              `}}},
				},
			},
			{
				Code: `
        describe('some tests', () => {
          beforeEach(() => { expect.hasAssertions(); });
          it("it1", function() {
            someFunctionToDo();
            someFunctionToDo2();
          });
        });

        describe('more tests', () => {
          it("it1", function() {
            someFunctionToDo();
            someFunctionToDo2();
          });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 10, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                describe('some tests', () => {
                  beforeEach(() => { expect.hasAssertions(); });
                  it("it1", function() {
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });

                describe('more tests', () => {
                  it("it1", function() {expect.hasAssertions();
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                describe('some tests', () => {
                  beforeEach(() => { expect.hasAssertions(); });
                  it("it1", function() {
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });

                describe('more tests', () => {
                  it("it1", function() {expect.assertions();
                    someFunctionToDo();
                    someFunctionToDo2();
                  });
                });
              `}}},
				},
			},
			{
				Code: `it("it1", function() {var a = 2;})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `it("it1", function() {expect.hasAssertions();var a = 2;})`}, {MessageId: `suggestAddingAssertions`, Output: `it("it1", function() {expect.assertions();var a = 2;})`}}},
				},
			},
			{
				Code: `it("it1", function() {expect.assertions();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assertionsRequiresOneArgument`, Line: 1, Column: 30},
				},
			},
			{
				Code: `it("it1", function() {expect.assertions(1,2);})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assertionsRequiresOneArgument`, Line: 1, Column: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestRemovingExtraArguments`, Output: `it("it1", function() {expect.assertions(1,);})`}}},
				},
			},
			{
				Code: `it("it1", function() {expect.assertions(1,2,);})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assertionsRequiresOneArgument`, Line: 1, Column: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestRemovingExtraArguments`, Output: `it("it1", function() {expect.assertions(1,);})`}}},
				},
			},
			{
				Code: `it("it1", function() {expect.assertions("1");})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assertionsRequiresNumberArgument`, Line: 1, Column: 41},
				},
			},
			{
				Code: `beforeEach(() => { expect.hasAssertions("1") })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `hasAssertionsTakesNoArguments`, Line: 1, Column: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestRemovingExtraArguments`, Output: `beforeEach(() => { expect.hasAssertions() })`}}},
				},
			},
			{
				Code: `beforeEach(() => expect.hasAssertions("1"))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `hasAssertionsTakesNoArguments`, Line: 1, Column: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestRemovingExtraArguments`, Output: `beforeEach(() => expect.hasAssertions())`}}},
				},
			},
			{
				Code: `afterEach(() => { expect.hasAssertions("1") })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `hasAssertionsTakesNoArguments`, Line: 1, Column: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestRemovingExtraArguments`, Output: `afterEach(() => { expect.hasAssertions() })`}}},
				},
			},
			{
				Code: `afterEach(() => expect.hasAssertions("1"))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `hasAssertionsTakesNoArguments`, Line: 1, Column: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestRemovingExtraArguments`, Output: `afterEach(() => expect.hasAssertions())`}}},
				},
			},
			{
				Code: `it("it1", function() {expect.hasAssertions("1");})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `hasAssertionsTakesNoArguments`, Line: 1, Column: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestRemovingExtraArguments`, Output: `it("it1", function() {expect.hasAssertions();})`}}},
				},
			},
			{
				Code: `it("it1", function() {expect.hasAssertions("1",);})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `hasAssertionsTakesNoArguments`, Line: 1, Column: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestRemovingExtraArguments`, Output: `it("it1", function() {expect.hasAssertions();})`}}},
				},
			},
			{
				Code: `it("it1", function() {expect.hasAssertions("1", "2");})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `hasAssertionsTakesNoArguments`, Line: 1, Column: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestRemovingExtraArguments`, Output: `it("it1", function() {expect.hasAssertions();})`}}},
				},
			},
			{
				Code: `
        it("it1", function() {
          expect.hasAssertions(() => {
            someFunctionToDo();
            someFunctionToDo2();
          });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `hasAssertionsTakesNoArguments`, Line: 2, Column: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestRemovingExtraArguments`, Output: `
                it("it1", function() {
                  expect.hasAssertions();
                });
              `}}},
				},
			},
			{
				Code: `
        it("it1", async function() {
          expect(someValue).toBe(true);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("it1", async function() {expect.hasAssertions();
                  expect(someValue).toBe(true);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("it1", async function() {expect.assertions();
                  expect(someValue).toBe(true);
                });
              `}}},
				},
			},
			{
				Code: `
        it("returns numbers that are greater than four", async () => {
          for(let thing in things) {
            expect(number).toBeGreaterThan(4);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("returns numbers that are greater than four", async () => {expect.hasAssertions();
                  for(let thing in things) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("returns numbers that are greater than four", async () => {expect.assertions();
                  for(let thing in things) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it("returns numbers that are greater than four", async () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("returns numbers that are greater than four", async () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("returns numbers that are greater than four", async () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it("returns numbers that are greater than four", async () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });

        it("returns numbers that are greater than five", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(5);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("returns numbers that are greater than four", async () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(5);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("returns numbers that are greater than four", async () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(5);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        beforeAll(() => { expect.hasAssertions(); });

        it("returns numbers that are greater than four", async () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });

        it("returns numbers that are greater than five", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(5);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 3, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                beforeAll(() => { expect.hasAssertions(); });

                it("returns numbers that are greater than four", async () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(5);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                beforeAll(() => { expect.hasAssertions(); });

                it("returns numbers that are greater than four", async () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(5);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        afterAll(() => { expect.hasAssertions(); });

        it("returns numbers that are greater than four", async () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });

        it("returns numbers that are greater than five", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(5);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 3, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                afterAll(() => { expect.hasAssertions(); });

                it("returns numbers that are greater than four", async () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(5);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                afterAll(() => { expect.hasAssertions(); });

                it("returns numbers that are greater than four", async () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(5);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it('only returns numbers that are greater than six', () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(6);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('only returns numbers that are greater than six', () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(6);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('only returns numbers that are greater than six', () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(6);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it('returns numbers that are greater than two', function () {
          const expectNumbersToBeGreaterThan = (numbers, value) => {
            for (let number of numbers) {
              expect(number).toBeGreaterThan(value);
            }
          };

          expectNumbersToBeGreaterThan(getNumbers(), 2);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('returns numbers that are greater than two', function () {expect.hasAssertions();
                  const expectNumbersToBeGreaterThan = (numbers, value) => {
                    for (let number of numbers) {
                      expect(number).toBeGreaterThan(value);
                    }
                  };

                  expectNumbersToBeGreaterThan(getNumbers(), 2);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('returns numbers that are greater than two', function () {expect.assertions();
                  const expectNumbersToBeGreaterThan = (numbers, value) => {
                    for (let number of numbers) {
                      expect(number).toBeGreaterThan(value);
                    }
                  };

                  expectNumbersToBeGreaterThan(getNumbers(), 2);
                });
              `}}},
				},
			},
			{
				Code: `
        it("only returns numbers that are greater than seven", function () {
          const numbers = getNumbers();

          for (let i = 0; i < numbers.length; i++) {
            expect(numbers[i]).toBeGreaterThan(7);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("only returns numbers that are greater than seven", function () {expect.hasAssertions();
                  const numbers = getNumbers();

                  for (let i = 0; i < numbers.length; i++) {
                    expect(numbers[i]).toBeGreaterThan(7);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("only returns numbers that are greater than seven", function () {expect.assertions();
                  const numbers = getNumbers();

                  for (let i = 0; i < numbers.length; i++) {
                    expect(numbers[i]).toBeGreaterThan(7);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it('has the number two', () => {
          expect(number).toBe(2);
        });

        it('only returns numbers that are less than twenty', () => {
          for (const number of getNumbers()) {
            expect(number).toBeLessThan(20);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 5, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('has the number two', () => {
                  expect(number).toBe(2);
                });

                it('only returns numbers that are less than twenty', () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeLessThan(20);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('has the number two', () => {
                  expect(number).toBe(2);
                });

                it('only returns numbers that are less than twenty', () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeLessThan(20);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it("is wrong");

        it("is a test", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 3, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("is wrong");

                it("is a test", () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("is wrong");

                it("is a test", () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it("is a number that is greater than four", () => {
          expect(number).toBeGreaterThan(4);
        });

        it("returns numbers that are greater than four", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });

        it("returns numbers that are greater than five", () => {
          expect(number).toBeGreaterThan(5);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 5, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });

                it("returns numbers that are greater than four", () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {
                  expect(number).toBeGreaterThan(5);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });

                it("returns numbers that are greater than four", () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {
                  expect(number).toBeGreaterThan(5);
                });
              `}}},
				},
			},
			{
				Code: `
        describe('my tests', () => {
          beforeEach(expect.hasAssertions);
          it("is a number that is greater than four", () => {
            expect(number).toBeGreaterThan(4);
          });
        });

        describe('more tests', () => {
          it("returns numbers that are greater than four", () => {
            for (const number of getNumbers()) {
              expect(number).toBeGreaterThan(4);
            }
          });
        });

        it("returns numbers that are greater than five", () => {
          expect(number).toBeGreaterThan(5);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 9, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                describe('my tests', () => {
                  beforeEach(expect.hasAssertions);
                  it("is a number that is greater than four", () => {
                    expect(number).toBeGreaterThan(4);
                  });
                });

                describe('more tests', () => {
                  it("returns numbers that are greater than four", () => {expect.hasAssertions();
                    for (const number of getNumbers()) {
                      expect(number).toBeGreaterThan(4);
                    }
                  });
                });

                it("returns numbers that are greater than five", () => {
                  expect(number).toBeGreaterThan(5);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                describe('my tests', () => {
                  beforeEach(expect.hasAssertions);
                  it("is a number that is greater than four", () => {
                    expect(number).toBeGreaterThan(4);
                  });
                });

                describe('more tests', () => {
                  it("returns numbers that are greater than four", () => {expect.assertions();
                    for (const number of getNumbers()) {
                      expect(number).toBeGreaterThan(4);
                    }
                  });
                });

                it("returns numbers that are greater than five", () => {
                  expect(number).toBeGreaterThan(5);
                });
              `}}},
				},
			},
			{
				Code: `
        it.each([1, 2, 3])("returns numbers that are greater than four", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });

        it("is a number that is greater than four", () => {
          expect(number).toBeGreaterThan(4);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it.each([1, 2, 3])("returns numbers that are greater than four", () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it.each([1, 2, 3])("returns numbers that are greater than four", () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });
              `}}},
				},
			},
			{
				Code: `
        it("returns numbers that are greater than four", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });

        it("is a number that is greater than four", () => {
          expect(number).toBeGreaterThan(4);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("returns numbers that are greater than four", () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("returns numbers that are greater than four", () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });
              `}}},
				},
			},
			{
				Code: `
        it("returns numbers that are greater than four", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });

        it("is a number that is greater than four", () => {
          expect.hasAssertions();

          expect(number).toBeGreaterThan(4);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("returns numbers that are greater than four", () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("is a number that is greater than four", () => {
                  expect.hasAssertions();

                  expect(number).toBeGreaterThan(4);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("returns numbers that are greater than four", () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("is a number that is greater than four", () => {
                  expect.hasAssertions();

                  expect(number).toBeGreaterThan(4);
                });
              `}}},
				},
			},
			{
				Code: `
        it("it1", () => {
          expect.hasAssertions();

          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(0);
          }
        });

        it("it1", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(0);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 9, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("it1", () => {
                  expect.hasAssertions();

                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(0);
                  }
                });

                it("it1", () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(0);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("it1", () => {
                  expect.hasAssertions();

                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(0);
                  }
                });

                it("it1", () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(0);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it("returns numbers that are greater than four", async () => {
          for (const number of await getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });

        it("returns numbers that are greater than five", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(5);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("returns numbers that are greater than four", async () => {expect.hasAssertions();
                  for (const number of await getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(5);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("returns numbers that are greater than four", async () => {expect.assertions();
                  for (const number of await getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(5);
                  }
                });
              `}}},
					{MessageId: `haveExpectAssertions`, Line: 7, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("returns numbers that are greater than four", async () => {
                  for (const number of await getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(5);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("returns numbers that are greater than four", async () => {
                  for (const number of await getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("returns numbers that are greater than five", () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(5);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it("it1", async () => {
          expect.hasAssertions();

          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });

        it("it1", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 9, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("it1", async () => {
                  expect.hasAssertions();

                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("it1", () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("it1", async () => {
                  expect.hasAssertions();

                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("it1", () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        describe('my tests', () => {
          beforeEach(() => { expect.hasAssertions(); });
 
          it("it1", async () => {
            for (const number of getNumbers()) {
              expect(number).toBeGreaterThan(4);
            }
          });
        });

        it("it1", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 11, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                describe('my tests', () => {
                  beforeEach(() => { expect.hasAssertions(); });

                  it("it1", async () => {
                    for (const number of getNumbers()) {
                      expect(number).toBeGreaterThan(4);
                    }
                  });
                });

                it("it1", () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                describe('my tests', () => {
                  beforeEach(() => { expect.hasAssertions(); });

                  it("it1", async () => {
                    for (const number of getNumbers()) {
                      expect(number).toBeGreaterThan(4);
                    }
                  });
                });

                it("it1", () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        describe('my tests', () => {
          afterEach(() => { expect.hasAssertions(); });
 
          it("it1", async () => {
            for (const number of getNumbers()) {
              expect(number).toBeGreaterThan(4);
            }
          });
        });

        it("it1", () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 11, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                describe('my tests', () => {
                  afterEach(() => { expect.hasAssertions(); });

                  it("it1", async () => {
                    for (const number of getNumbers()) {
                      expect(number).toBeGreaterThan(4);
                    }
                  });
                });

                it("it1", () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                describe('my tests', () => {
                  afterEach(() => { expect.hasAssertions(); });

                  it("it1", async () => {
                    for (const number of getNumbers()) {
                      expect(number).toBeGreaterThan(4);
                    }
                  });
                });

                it("it1", () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}}},
				},
			},
			{
				Code:    "\n        it.skip.each``(\"it1\", async () => {\n          expect.hasAssertions();\n\n          for (const number of getNumbers()) {\n            expect(number).toBeGreaterThan(4);\n          }\n        });\n\n        it(\"it1\", () => {\n          for (const number of getNumbers()) {\n            expect(number).toBeGreaterThan(4);\n          }\n        });\n      ",
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 9, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: "\n                it.skip.each``(\"it1\", async () => {\n                  expect.hasAssertions();\n\n                  for (const number of getNumbers()) {\n                    expect(number).toBeGreaterThan(4);\n                  }\n                });\n\n                it(\"it1\", () => {expect.hasAssertions();\n                  for (const number of getNumbers()) {\n                    expect(number).toBeGreaterThan(4);\n                  }\n                });\n              "}, {MessageId: `suggestAddingAssertions`, Output: "\n                it.skip.each``(\"it1\", async () => {\n                  expect.hasAssertions();\n\n                  for (const number of getNumbers()) {\n                    expect(number).toBeGreaterThan(4);\n                  }\n                });\n\n                it(\"it1\", () => {expect.assertions();\n                  for (const number of getNumbers()) {\n                    expect(number).toBeGreaterThan(4);\n                  }\n                });\n              "}}},
				},
			},
			{
				Code: `
        it("it1", async () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });

        it("it1", () => {
          expect.hasAssertions();

          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("it1", async () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("it1", () => {
                  expect.hasAssertions();

                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("it1", async () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });

                it("it1", () => {
                  expect.hasAssertions();

                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        describe('my tests', () => {
          it("it1", async () => {
            for (const number of getNumbers()) {
              expect(number).toBeGreaterThan(4);
            }
          });
        });

        it("it1", () => {
          expect.hasAssertions();

          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(4);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 2, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                describe('my tests', () => {
                  it("it1", async () => {expect.hasAssertions();
                    for (const number of getNumbers()) {
                      expect(number).toBeGreaterThan(4);
                    }
                  });
                });

                it("it1", () => {
                  expect.hasAssertions();

                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                describe('my tests', () => {
                  it("it1", async () => {expect.assertions();
                    for (const number of getNumbers()) {
                      expect(number).toBeGreaterThan(4);
                    }
                  });
                });

                it("it1", () => {
                  expect.hasAssertions();

                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it('sends the data as a string', () => {
          const stream = openStream();

          stream.on('data', data => {
            expect(data).toBe(expect.any(String));
          });
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('sends the data as a string', () => {expect.hasAssertions();
                  const stream = openStream();

                  stream.on('data', data => {
                    expect(data).toBe(expect.any(String));
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('sends the data as a string', () => {expect.assertions();
                  const stream = openStream();

                  stream.on('data', data => {
                    expect(data).toBe(expect.any(String));
                  });
                });
              `}}},
				},
			},
			{
				Code: `
        it('responds ok', function () {
          client.get('/user', response => {
            expect(response.status).toBe(200);
          });
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('responds ok', function () {expect.hasAssertions();
                  client.get('/user', response => {
                    expect(response.status).toBe(200);
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('responds ok', function () {expect.assertions();
                  client.get('/user', response => {
                    expect(response.status).toBe(200);
                  });
                });
              `}}},
				},
			},
			{
				Code: `
        it('responds ok', function () {
          client.get('/user', response => {
            expect.assertions(1);

            expect(response.status).toBe(200);
          });
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('responds ok', function () {expect.hasAssertions();
                  client.get('/user', response => {
                    expect.assertions(1);

                    expect(response.status).toBe(200);
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('responds ok', function () {expect.assertions();
                  client.get('/user', response => {
                    expect.assertions(1);

                    expect(response.status).toBe(200);
                  });
                });
              `}}},
				},
			},
			{
				Code: `
        it('responds ok', function () {
          const expectOkResponse = response => {
            expect.assertions(1);

            expect(response.status).toBe(200);
          };

          client.get('/user', expectOkResponse);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('responds ok', function () {expect.hasAssertions();
                  const expectOkResponse = response => {
                    expect.assertions(1);

                    expect(response.status).toBe(200);
                  };

                  client.get('/user', expectOkResponse);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('responds ok', function () {expect.assertions();
                  const expectOkResponse = response => {
                    expect.assertions(1);

                    expect(response.status).toBe(200);
                  };

                  client.get('/user', expectOkResponse);
                });
              `}}},
				},
			},
			{
				Code: `
        it('returns numbers that are greater than two', function () {
          const expectNumberToBeGreaterThan = (number, value) => {
            expect(number).toBeGreaterThan(value);
          };

          expectNumberToBeGreaterThan(1, 2);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('returns numbers that are greater than two', function () {expect.hasAssertions();
                  const expectNumberToBeGreaterThan = (number, value) => {
                    expect(number).toBeGreaterThan(value);
                  };

                  expectNumberToBeGreaterThan(1, 2);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('returns numbers that are greater than two', function () {expect.assertions();
                  const expectNumberToBeGreaterThan = (number, value) => {
                    expect(number).toBeGreaterThan(value);
                  };

                  expectNumberToBeGreaterThan(1, 2);
                });
              `}}},
				},
			},
			{
				Code: `
        it('returns numbers that are greater than two', function () {
          const expectNumbersToBeGreaterThan = (numbers, value) => {
            for (let number of numbers) {
              expect(number).toBeGreaterThan(value);
            }
          };

          expectNumbersToBeGreaterThan(getNumbers(), 2);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('returns numbers that are greater than two', function () {expect.hasAssertions();
                  const expectNumbersToBeGreaterThan = (numbers, value) => {
                    for (let number of numbers) {
                      expect(number).toBeGreaterThan(value);
                    }
                  };

                  expectNumbersToBeGreaterThan(getNumbers(), 2);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('returns numbers that are greater than two', function () {expect.assertions();
                  const expectNumbersToBeGreaterThan = (numbers, value) => {
                    for (let number of numbers) {
                      expect(number).toBeGreaterThan(value);
                    }
                  };

                  expectNumbersToBeGreaterThan(getNumbers(), 2);
                });
              `}}},
				},
			},
			{
				Code: `
        it('only returns numbers that are greater than six', () => {
          getNumbers().forEach(number => {
            expect(number).toBeGreaterThan(6);
          });
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('only returns numbers that are greater than six', () => {expect.hasAssertions();
                  getNumbers().forEach(number => {
                    expect(number).toBeGreaterThan(6);
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('only returns numbers that are greater than six', () => {expect.assertions();
                  getNumbers().forEach(number => {
                    expect(number).toBeGreaterThan(6);
                  });
                });
              `}}},
				},
			},
			{
				Code: `
        it("is wrong");

        it('responds ok', function () {
          const expectOkResponse = response => {
            expect.assertions(1);

            expect(response.status).toBe(200);
          };

          client.get('/user', expectOkResponse);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 3, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("is wrong");

                it('responds ok', function () {expect.hasAssertions();
                  const expectOkResponse = response => {
                    expect.assertions(1);

                    expect(response.status).toBe(200);
                  };

                  client.get('/user', expectOkResponse);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("is wrong");

                it('responds ok', function () {expect.assertions();
                  const expectOkResponse = response => {
                    expect.assertions(1);

                    expect(response.status).toBe(200);
                  };

                  client.get('/user', expectOkResponse);
                });
              `}}},
				},
			},
			{
				Code: `
        it("is a number that is greater than four", () => {
          expect(number).toBeGreaterThan(4);
        });

        it('responds ok', function () {
          const expectOkResponse = response => {
            expect(response.status).toBe(200);
          };

          client.get('/user', expectOkResponse);
        });

        it("returns numbers that are greater than five", () => {
          expect(number).toBeGreaterThan(5);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 5, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });

                it('responds ok', function () {expect.hasAssertions();
                  const expectOkResponse = response => {
                    expect(response.status).toBe(200);
                  };

                  client.get('/user', expectOkResponse);
                });

                it("returns numbers that are greater than five", () => {
                  expect(number).toBeGreaterThan(5);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });

                it('responds ok', function () {expect.assertions();
                  const expectOkResponse = response => {
                    expect(response.status).toBe(200);
                  };

                  client.get('/user', expectOkResponse);
                });

                it("returns numbers that are greater than five", () => {
                  expect(number).toBeGreaterThan(5);
                });
              `}}},
				},
			},
			{
				Code: `
        it("is a number that is greater than four", () => {
          expect(number).toBeGreaterThan(4);
        });

        it("returns numbers that are greater than four", () => {
          getNumbers().map(number => {
            expect(number).toBeGreaterThan(0);
          });
        });

        it("returns numbers that are greater than five", () => {
          expect(number).toBeGreaterThan(5);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 5, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });

                it("returns numbers that are greater than four", () => {expect.hasAssertions();
                  getNumbers().map(number => {
                    expect(number).toBeGreaterThan(0);
                  });
                });

                it("returns numbers that are greater than five", () => {
                  expect(number).toBeGreaterThan(5);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });

                it("returns numbers that are greater than four", () => {expect.assertions();
                  getNumbers().map(number => {
                    expect(number).toBeGreaterThan(0);
                  });
                });

                it("returns numbers that are greater than five", () => {
                  expect(number).toBeGreaterThan(5);
                });
              `}}},
				},
			},
			{
				Code:    "\n        it.each([1, 2, 3])(\"returns ok\", id => {\n          client.get(`/users/${id}`, response => {\n            expect(response.status).toBe(200);\n          });\n        });\n\n        it(\"is a number that is greater than four\", () => {\n          expect(number).toBeGreaterThan(4);\n        });\n      ",
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: "\n                it.each([1, 2, 3])(\"returns ok\", id => {expect.hasAssertions();\n                  client.get(`/users/${id}`, response => {\n                    expect(response.status).toBe(200);\n                  });\n                });\n\n                it(\"is a number that is greater than four\", () => {\n                  expect(number).toBeGreaterThan(4);\n                });\n              "}, {MessageId: `suggestAddingAssertions`, Output: "\n                it.each([1, 2, 3])(\"returns ok\", id => {expect.assertions();\n                  client.get(`/users/${id}`, response => {\n                    expect(response.status).toBe(200);\n                  });\n                });\n\n                it(\"is a number that is greater than four\", () => {\n                  expect(number).toBeGreaterThan(4);\n                });\n              "}}},
				},
			},
			{
				Code: `
        it('responds ok', function () {
          client.get('/user', response => {
            expect(response.status).toBe(200);
          });
        });

        it("is a number that is greater than four", () => {
          expect(number).toBeGreaterThan(4);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('responds ok', function () {expect.hasAssertions();
                  client.get('/user', response => {
                    expect(response.status).toBe(200);
                  });
                });

                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('responds ok', function () {expect.assertions();
                  client.get('/user', response => {
                    expect(response.status).toBe(200);
                  });
                });

                it("is a number that is greater than four", () => {
                  expect(number).toBeGreaterThan(4);
                });
              `}}},
				},
			},
			{
				Code: `
        it('responds ok', function () {
          client.get('/user', response => {
            expect(response.status).toBe(200);
          });
        });

        it("is a number that is greater than four", () => {
          expect.hasAssertions();

          expect(number).toBeGreaterThan(4);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('responds ok', function () {expect.hasAssertions();
                  client.get('/user', response => {
                    expect(response.status).toBe(200);
                  });
                });

                it("is a number that is greater than four", () => {
                  expect.hasAssertions();

                  expect(number).toBeGreaterThan(4);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('responds ok', function () {expect.assertions();
                  client.get('/user', response => {
                    expect(response.status).toBe(200);
                  });
                });

                it("is a number that is greater than four", () => {
                  expect.hasAssertions();

                  expect(number).toBeGreaterThan(4);
                });
              `}}},
				},
			},
			{
				Code: `
        it("it1", () => {
          expect.hasAssertions();

          getNumbers().forEach(number => {
            expect(number).toBeGreaterThan(0);
          });
        });

        it("it1", () => {
          getNumbers().forEach(number => {
            expect(number).toBeGreaterThan(0);
          });
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 9, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("it1", () => {
                  expect.hasAssertions();

                  getNumbers().forEach(number => {
                    expect(number).toBeGreaterThan(0);
                  });
                });

                it("it1", () => {expect.hasAssertions();
                  getNumbers().forEach(number => {
                    expect(number).toBeGreaterThan(0);
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("it1", () => {
                  expect.hasAssertions();

                  getNumbers().forEach(number => {
                    expect(number).toBeGreaterThan(0);
                  });
                });

                it("it1", () => {expect.assertions();
                  getNumbers().forEach(number => {
                    expect(number).toBeGreaterThan(0);
                  });
                });
              `}}},
				},
			},
			{
				Code: `
        it('responds ok', function () {
          expect.hasAssertions();

          client.get('/user', response => {
            expect(response.status).toBe(200);
          });
        });

        it('responds not found', function () {
          client.get('/user', response => {
            expect(response.status).toBe(404);
          });
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 9, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('responds ok', function () {
                  expect.hasAssertions();

                  client.get('/user', response => {
                    expect(response.status).toBe(200);
                  });
                });

                it('responds not found', function () {expect.hasAssertions();
                  client.get('/user', response => {
                    expect(response.status).toBe(404);
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('responds ok', function () {
                  expect.hasAssertions();

                  client.get('/user', response => {
                    expect(response.status).toBe(200);
                  });
                });

                it('responds not found', function () {expect.assertions();
                  client.get('/user', response => {
                    expect(response.status).toBe(404);
                  });
                });
              `}}},
				},
			},
			{
				Code:    "\n        it.skip.each``(\"it1\", async () => {\n          expect.hasAssertions();\n\n          client.get('/user', response => {\n            expect(response.status).toBe(200);\n          });\n        });\n\n        it(\"responds ok\", () => {\n          client.get('/user', response => {\n            expect(response.status).toBe(200);\n          });\n        });\n      ",
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInCallback`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 9, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: "\n                it.skip.each``(\"it1\", async () => {\n                  expect.hasAssertions();\n\n                  client.get('/user', response => {\n                    expect(response.status).toBe(200);\n                  });\n                });\n\n                it(\"responds ok\", () => {expect.hasAssertions();\n                  client.get('/user', response => {\n                    expect(response.status).toBe(200);\n                  });\n                });\n              "}, {MessageId: `suggestAddingAssertions`, Output: "\n                it.skip.each``(\"it1\", async () => {\n                  expect.hasAssertions();\n\n                  client.get('/user', response => {\n                    expect(response.status).toBe(200);\n                  });\n                });\n\n                it(\"responds ok\", () => {expect.assertions();\n                  client.get('/user', response => {\n                    expect(response.status).toBe(200);\n                  });\n                });\n              "}}},
				},
			},
			{
				Code: `
        it("returns numbers that are greater than four", function(expect) {
          expect.assertions(2);

          for(let thing in things) {
            expect(number).toBeGreaterThan(4);
          }
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Column: 1, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("returns numbers that are greater than four", function(expect) {expect.hasAssertions();
                  expect.assertions(2);

                  for(let thing in things) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("returns numbers that are greater than four", function(expect) {expect.assertions();
                  expect.assertions(2);

                  for(let thing in things) {
                    expect(number).toBeGreaterThan(4);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it('only returns numbers that are greater than zero', () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(0);
          }
        });

        it("is zero", () => {
          expect.hasAssertions();

          expect(0).toBe(0);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('only returns numbers that are greater than zero', () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(0);
                  }
                });

                it("is zero", () => {
                  expect.hasAssertions();

                  expect(0).toBe(0);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('only returns numbers that are greater than zero', () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(0);
                  }
                });

                it("is zero", () => {
                  expect.hasAssertions();

                  expect(0).toBe(0);
                });
              `}}},
				},
			},
			{
				Code: `
        it('only returns numbers that are greater than zero', () => {
          expect.hasAssertions();

          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(0);
          }
        });

        it('only returns numbers that are less than 100', () => {
          for (const number of getNumbers()) {
            expect(number).toBeLessThan(0);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 9, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('only returns numbers that are greater than zero', () => {
                  expect.hasAssertions();

                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(0);
                  }
                });

                it('only returns numbers that are less than 100', () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeLessThan(0);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('only returns numbers that are greater than zero', () => {
                  expect.hasAssertions();

                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(0);
                  }
                });

                it('only returns numbers that are less than 100', () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeLessThan(0);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        it("to be true", async function() {
          expect(someValue).toBe(true);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true, `onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it("to be true", async function() {expect.hasAssertions();
                  expect(someValue).toBe(true);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it("to be true", async function() {expect.assertions();
                  expect(someValue).toBe(true);
                });
              `}}},
				},
			},
			{
				Code: `
        it('only returns numbers that are greater than zero', async () => {
          for (const number of getNumbers()) {
            expect(number).toBeGreaterThan(0);
          }
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true, `onlyFunctionsWithExpectInLoop`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it('only returns numbers that are greater than zero', async () => {expect.hasAssertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(0);
                  }
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it('only returns numbers that are greater than zero', async () => {expect.assertions();
                  for (const number of getNumbers()) {
                    expect(number).toBeGreaterThan(0);
                  }
                });
              `}}},
				},
			},
			{
				Code: `
        test.each()("is not fine", () => {
          expect(someValue).toBe(true);
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                test.each()("is not fine", () => {expect.hasAssertions();
                  expect(someValue).toBe(true);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                test.each()("is not fine", () => {expect.assertions();
                  expect(someValue).toBe(true);
                });
              `}}},
				},
			},
			{
				Code: `
        describe.each()('something', () => {
          it("is not fine", () => {
            expect(someValue).toBe(true);
          });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 2, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                describe.each()('something', () => {
                  it("is not fine", () => {expect.hasAssertions();
                    expect(someValue).toBe(true);
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                describe.each()('something', () => {
                  it("is not fine", () => {expect.assertions();
                    expect(someValue).toBe(true);
                  });
                });
              `}}},
				},
			},
			{
				Code: `
        describe.each()('something', () => {
          test.each()("is not fine", () => {
            expect(someValue).toBe(true);
          });
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 2, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                describe.each()('something', () => {
                  test.each()("is not fine", () => {expect.hasAssertions();
                    expect(someValue).toBe(true);
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                describe.each()('something', () => {
                  test.each()("is not fine", () => {expect.assertions();
                    expect(someValue).toBe(true);
                  });
                });
              `}}},
				},
			},
			{
				Code: `
        test.each()("is not fine", async () => {
          expect(someValue).toBe(true);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                test.each()("is not fine", async () => {expect.hasAssertions();
                  expect(someValue).toBe(true);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                test.each()("is not fine", async () => {expect.assertions();
                  expect(someValue).toBe(true);
                });
              `}}},
				},
			},
			{
				Code: `
        it.each()("is not fine", async () => {
          expect(someValue).toBe(true);
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 1, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                it.each()("is not fine", async () => {expect.hasAssertions();
                  expect(someValue).toBe(true);
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                it.each()("is not fine", async () => {expect.assertions();
                  expect(someValue).toBe(true);
                });
              `}}},
				},
			},
			{
				Code: `
        describe.each()('something', () => {
          test.each()("is not fine", async () => {
            expect(someValue).toBe(true);
          });
        });
      `,
				Options: []interface{}{map[string]interface{}{`onlyFunctionsWithAsyncKeyword`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `haveExpectAssertions`, Line: 2, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: `suggestAddingHasAssertions`, Output: `
                describe.each()('something', () => {
                  test.each()("is not fine", async () => {expect.hasAssertions();
                    expect(someValue).toBe(true);
                  });
                });
              `}, {MessageId: `suggestAddingAssertions`, Output: `
                describe.each()('something', () => {
                  test.each()("is not fine", async () => {expect.assertions();
                    expect(someValue).toBe(true);
                  });
                });
              `}}},
				},
			},
		},
	)
}
