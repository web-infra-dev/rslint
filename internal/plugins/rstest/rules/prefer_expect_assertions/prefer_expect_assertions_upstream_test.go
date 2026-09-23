package prefer_expect_assertions_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_expect_assertions"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferExpectAssertionsUpstream migrates every valid/invalid case of
// vitest-dev/eslint-plugin-vitest v1.6.27
// tests/prefer-expect-assertions.test.ts. Diagnostic end positions and
// suggestion outputs were taken from running the upstream rule on each case.
// Cases whose expectation differs from upstream carry a comment explaining the
// Rstest behavior that decides them.
func TestPreferExpectAssertionsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_expect_assertions.PreferExpectAssertionsRule,
		[]rule_tester.ValidTestCase{
			// ---- prefer-expect-assertions ----
			{Code: `test("it1", () => {expect.assertions(0);})`},
			{Code: `test("it1", function() {expect.assertions(0);})`},
			{Code: `test("it1", function() {expect.hasAssertions();})`},
			{Code: `it("it1", function() {expect.assertions(0);})`},
			{Code: `test("it1")`},
			{Code: `itHappensToStartWithIt("foo", function() {})`},
			{Code: `testSomething("bar", function() {})`},
			{Code: `it(async () => {expect.assertions(0);})`},
			{Code: `test("example-fail", async ({ expect }) => {
    expect.assertions(1);
    await expect(Promise.resolve(null)).resolves.toBeNull();
  });
    `},
			{Code: `it("it1", () => {
    expect.assertions(0);
    const foo = { bar({ baz }) { baz(); } };
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
   `, Options: []interface{}{map[string]interface{}{"onlyFunctionsWithExpectInLoop": true}}},
			{Code: `
   it("returns numbers that are greater than five", function () {
    expect.assertions(2);
    for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
   }
   });
   `, Options: []interface{}{map[string]interface{}{"onlyFunctionsWithExpectInLoop": true}}},
			{Code: `it("returns things that are less than ten", function () {
    expect.hasAssertions();
    for (const thing in things) {
     expect(thing).toBeLessThan(10);
    }
   });`, Options: []interface{}{map[string]interface{}{"onlyFunctionsWithExpectInLoop": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- prefer-expect-assertions ----
			{
				Code: `it("it1", () => foo())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: `
it('my test description', ({ expect }) => {
  const a = 1;
  const b = 2;

  expect(sum(a, b)).toBe(a + b);
})
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 2, Column: 1, EndLine: 7, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestAddingHasAssertions", Output: `
it('my test description', ({ expect }) => {expect.hasAssertions();
  const a = 1;
  const b = 2;

  expect(sum(a, b)).toBe(a + b);
})
`},
						{MessageId: "suggestAddingAssertions", Output: `
it('my test description', ({ expect }) => {expect.assertions();
  const a = 1;
  const b = 2;

  expect(sum(a, b)).toBe(a + b);
})
`},
					}},
				},
			},
			{
				Code: `
it('my test description', (context) => {
  const a = 1;
  const b = 2;

  context.expect(sum(a, b)).toBe(a + b);
})
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 2, Column: 1, EndLine: 7, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestAddingHasAssertions", Output: `
it('my test description', (context) => {context.expect.hasAssertions();
  const a = 1;
  const b = 2;

  context.expect(sum(a, b)).toBe(a + b);
})
`},
						{MessageId: "suggestAddingAssertions", Output: `
it('my test description', (context) => {context.expect.assertions();
  const a = 1;
  const b = 2;

  context.expect(sum(a, b)).toBe(a + b);
})
`},
					}},
				},
			},
			{
				Code: `it('resolves', () => expect(staged()).toBe(true));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 50},
				},
			},
			{
				Code: `it('resolves', async () => expect(await staged()).toBe(true));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 62},
				},
			},
			{
				Code: `it('checks value', () => {
  expect(value).to.be.a("string")
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 3, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestAddingHasAssertions", Output: `it('checks value', () => {expect.hasAssertions();
  expect(value).to.be.a("string")
})`},
						{MessageId: "suggestAddingAssertions", Output: `it('checks value', () => {expect.assertions();
  expect(value).to.be.a("string")
})`},
					}},
				},
			},
			{
				Code: `it("it1", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestAddingHasAssertions", Output: `it("it1", () => {expect.hasAssertions();})`},
						{MessageId: "suggestAddingAssertions", Output: `it("it1", () => {expect.assertions();})`},
					}},
				},
			},
			{
				Code: `it("it1", () => { foo()})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestAddingHasAssertions", Output: `it("it1", () => {expect.hasAssertions(); foo()})`},
						{MessageId: "suggestAddingAssertions", Output: `it("it1", () => {expect.assertions(); foo()})`},
					}},
				},
			},
			{
				Code: `it("it1", function() {var a = 2;})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestAddingHasAssertions", Output: `it("it1", function() {expect.hasAssertions();var a = 2;})`},
						{MessageId: "suggestAddingAssertions", Output: `it("it1", function() {expect.assertions();var a = 2;})`},
					}},
				},
			},
			{
				Code: `it("it1", function() {expect.assertions();})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assertionsRequiresOneArgument", Line: 1, Column: 30, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code: `it("it1", function() {expect.assertions(1,2);})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assertionsRequiresOneArgument", Line: 1, Column: 43, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemovingExtraArguments", Output: `it("it1", function() {expect.assertions(1,);})`},
					}},
				},
			},
			{
				Code: `it("it1", function() {expect.assertions(1,2,);})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assertionsRequiresOneArgument", Line: 1, Column: 43, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemovingExtraArguments", Output: `it("it1", function() {expect.assertions(1,);})`},
					}},
				},
			},
			{
				Code: `it("it1", function() {expect.assertions("1");})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assertionsRequiresNumberArgument", Line: 1, Column: 41, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code: `it("it1", function() {expect.hasAssertions("1");})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "hasAssertionsTakesNoArguments", Line: 1, Column: 30, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemovingExtraArguments", Output: `it("it1", function() {expect.hasAssertions();})`},
					}},
				},
			},
			{
				Code: `it("it1", function() {expect.hasAssertions("1",);})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "hasAssertionsTakesNoArguments", Line: 1, Column: 30, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemovingExtraArguments", Output: `it("it1", function() {expect.hasAssertions();})`},
					}},
				},
			},
			{
				Code: `it("it1", function() {expect.hasAssertions("1", "2");})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "hasAssertionsTakesNoArguments", Line: 1, Column: 30, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestRemovingExtraArguments", Output: `it("it1", function() {expect.hasAssertions();})`},
					}},
				},
			},
			{
				Code: `it("it1", () => {
    expect.hasAssertions();

    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(0);
    }
     });

     it("it1", () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(0);
    }
     });`,
				Options: []interface{}{map[string]interface{}{"onlyFunctionsWithExpectInLoop": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 9, Column: 6, EndLine: 13, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestAddingHasAssertions", Output: `it("it1", () => {
    expect.hasAssertions();

    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(0);
    }
     });

     it("it1", () => {expect.hasAssertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(0);
    }
     });`},
						{MessageId: "suggestAddingAssertions", Output: `it("it1", () => {
    expect.hasAssertions();

    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(0);
    }
     });

     it("it1", () => {expect.assertions();
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(0);
    }
     });`},
					}},
				},
			},
			{
				Code: `it("returns numbers that are greater than four", async () => {
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
				Options: []interface{}{map[string]interface{}{"onlyFunctionsWithExpectInLoop": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 5, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestAddingHasAssertions", Output: `it("returns numbers that are greater than four", async () => {expect.hasAssertions();
     for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
     }
   });

   it("returns numbers that are greater than five", () => {
     for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
     }
   });
    `},
						{MessageId: "suggestAddingAssertions", Output: `it("returns numbers that are greater than four", async () => {expect.assertions();
     for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
     }
   });

   it("returns numbers that are greater than five", () => {
     for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
     }
   });
    `},
					}},
					{MessageId: "haveExpectAssertions", Line: 7, Column: 4, EndLine: 11, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestAddingHasAssertions", Output: `it("returns numbers that are greater than four", async () => {
     for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
     }
   });

   it("returns numbers that are greater than five", () => {expect.hasAssertions();
     for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
     }
   });
    `},
						{MessageId: "suggestAddingAssertions", Output: `it("returns numbers that are greater than four", async () => {
     for (const number of await getNumbers()) {
    expect(number).toBeGreaterThan(4);
     }
   });

   it("returns numbers that are greater than five", () => {expect.assertions();
     for (const number of getNumbers()) {
    expect(number).toBeGreaterThan(5);
     }
   });
    `},
					}},
				},
			},
			{
				Code: `it("it1", () => {
    const foo = { bar({ baz }) { baz(); } };
  });
    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 3, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestAddingHasAssertions", Output: `it("it1", () => {expect.hasAssertions();
    const foo = { bar({ baz }) { baz(); } };
  });
    `},
						{MessageId: "suggestAddingAssertions", Output: `it("it1", () => {expect.assertions();
    const foo = { bar({ baz }) { baz(); } };
  });
    `},
					}},
				},
			},
			{
				Code:    `it("it1", function() {expect.hasAssertions();})`,
				Options: []interface{}{map[string]interface{}{"disallowHasAssertions": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAssertionsOverHasAssertions", Line: 1, Column: 30, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestReplacingWithAssertions", Output: `it("it1", function() {expect.assertions();})`},
					}},
				},
			},
			{
				Code: `it('my test description', ({ expect }) => {
  expect.hasAssertions();
  const a = 1;
  expect(a).toBe(1);
})`,
				Options: []interface{}{map[string]interface{}{"disallowHasAssertions": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAssertionsOverHasAssertions", Line: 2, Column: 10, EndLine: 2, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestReplacingWithAssertions", Output: `it('my test description', ({ expect }) => {
  expect.assertions();
  const a = 1;
  expect(a).toBe(1);
})`},
					}},
				},
			},
			{
				Code:    `it("it1", () => {})`,
				Options: []interface{}{map[string]interface{}{"disallowHasAssertions": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestAddingAssertions", Output: `it("it1", () => {expect.assertions();})`},
					}},
				},
			},
		},
	)
}
