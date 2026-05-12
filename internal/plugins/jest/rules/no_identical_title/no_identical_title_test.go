package no_identical_title_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_identical_title"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoIdenticalTitleRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_identical_title.NoIdenticalTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: "it(); it();"},
			{Code: "describe(); describe();"},
			{Code: "describe(\"foo\", () => {}); it(\"foo\", () => {});"},
			{Code: `describe("foo", () => {
  it("works", () => {});
});
`},
			{Code: `it('one', () => {});
it('two', () => {});
`},
			{Code: `describe('foo', () => {});
describe('foe', () => {});
`},
			{Code: "it(`one`, () => {});\nit(`two`, () => {});\n"},
			{Code: "describe(`foo`, () => {});\ndescribe(`foe`, () => {});\n"},
			{Code: `describe('foo', () => {
  test('this', () => {});
  test('that', () => {});
});
`},
			{Code: `test.concurrent('this', () => {});
test.concurrent('that', () => {});
`},
			{Code: `test.concurrent('this', () => {});
test.only.concurrent('that', () => {});
`},
			{Code: `test.only.concurrent('this', () => {});
test.concurrent('that', () => {});
`},
			{Code: `test.only.concurrent('this', () => {});
test.only.concurrent('that', () => {});
`},
			{Code: `test.only('this', () => {});
test.only('that', () => {});
`},
			{Code: `describe('foo', () => {
  it('works', () => {});

  describe('foe', () => {
    it('works', () => {});
  });
});
`},
			{Code: `describe('foo', () => {
  describe('foe', () => {
    it('works', () => {});
  });

  it('works', () => {});
});
`},
			{Code: "describe('foo', () => describe('foe', () => {}));"},
			{Code: `describe('foo', () => {
  describe('foe', () => {});
});

describe('foe', () => {});
`},
			{Code: "test(\"number\" + n, function() {});"},
			{Code: "test(\"number\" + n, function() {}); test(\"number\" + n, function() {});"},
			{Code: "it(`${n}`, function() {});"},
			{Code: "it(`${n}`, function() {}); it(`${n}`, function() {});"},
			{Code: `describe('a class named ' + myClass.name, () => {
  describe('#myMethod', () => {});
});

describe('something else', () => {});
`},
			{Code: `describe('my class', () => {
  describe('#myMethod', () => {});
  describe('a class named ' + myClass.name, () => {});
});
`},
			{Code: "describe(\"foo\", () => {\n  it(`ignores ${someVar} with the same title`, () => {});\n  it(`ignores ${someVar} with the same title`, () => {});\n});\n"},
			{Code: "const test = { content: () => \"foo\" };\ntest.content(`something that is not from jest`, () => {});\ntest.content(`something that is not from jest`, () => {});\n"},
			{Code: "const describe = { content: () => \"foo\" };\ndescribe.content(`something that is not from jest`, () => {});\ndescribe.content(`something that is not from jest`, () => {});\n"},
			{Code: "describe.each`\n  description\n  ${'b'}\n`('$description', () => {});\n\ndescribe.each`\n  description\n  ${'a'}\n`('$description', () => {});\n"},
			{Code: "describe('top level', () => {\n  describe.each``('nested each', () => {\n    describe.each``('nested nested each', () => {});\n  });\n\n  describe('nested', () => {});\n});\n"},
			{Code: "describe.each``('my title', value => {});\ndescribe.each``('my title', value => {});\ndescribe.each([])('my title', value => {});\ndescribe.each([])('my title', value => {});\n"},
			{Code: `describe.each([])('when the value is %s', value => {});
describe.each([])('when the value is %s', value => {});
`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `describe('foo', () => {
  it('works', () => {});
  it('works', () => {});
});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 3, Column: 6},
				},
			},
			{
				Code: `it('works', () => {});
it('works', () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 4},
				},
			},
			{
				Code: `test.only('this', () => {});
test('this', () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 6},
				},
			},
			{
				Code: `xtest('this', () => {});
test('this', () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 6},
				},
			},
			{
				Code: `test.only('this', () => {});
test.only('this', () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 11},
				},
			},
			{
				Code: `test.concurrent('this', () => {});
test.concurrent('this', () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 17},
				},
			},
			{
				Code: `test.only('this', () => {});
test.concurrent('this', () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 17},
				},
			},
			{
				Code: `describe('foo', () => {});
describe('foo', () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDescribeTitle", Line: 2, Column: 10},
				},
			},
			{
				Code: `describe('foo', () => {});
xdescribe('foo', () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDescribeTitle", Line: 2, Column: 11},
				},
			},
			{
				Code: `fdescribe('foo', () => {});
describe('foo', () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDescribeTitle", Line: 2, Column: 10},
				},
			},
			{
				Code: `describe('foo', () => {
  describe('foe', () => {});
});
describe('foo', () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDescribeTitle", Line: 4, Column: 10},
				},
			},
			{
				Code: "describe(\"foo\", () => {\n  it(`catches backticks with the same title`, () => {});\n  it(`catches backticks with the same title`, () => {});\n});\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 3, Column: 6},
				},
			},
			{
				Code: `context('foo', () => {
  describe('foe', () => {});
});
describe('foo', () => {});
`,
				Settings: map[string]interface{}{
					"jest": map[string]interface{}{
						"globalAliases": map[string]interface{}{
							"describe": []interface{}{"context"},
						},
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDescribeTitle", Line: 4, Column: 10},
				},
			},
			{
				Code: `spec('works', () => {});
it('works', () => {});
`,
				Settings: map[string]interface{}{
					"jest": map[string]interface{}{
						"globalAliases": map[string]interface{}{
							"it": []interface{}{"spec"},
						},
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleTestTitle", Line: 2, Column: 4},
				},
			},
		},
	)
}
