package valid_title_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/valid_title"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	patMustNotHashBad = `(?:#(?!unit|e2e))\w+`
	patMustMatchHash  = `^[^#]+$|(?:#(?:unit|e2e))`
	patHashTag        = `#(?:unit|integration|e2e)`
)

func TestValidTitleRule(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `describe("the correct way to properly handle all the things", () => {});`},
		{Code: `test("that all is as it should be", () => {});`},
		{
			Code: `it("correctly sets the value", () => {});`,
			Options: []interface{}{
				map[string]interface{}{
					"ignoreTypeOfDescribeName": false,
					"disallowedWords":          []interface{}{"correct"},
				},
			},
		},
		{
			Code: `it("correctly sets the value", () => {});`,
			Options: []interface{}{
				map[string]interface{}{"disallowedWords": nil},
			},
		},
		{
			Code: `it("correctly sets the value", () => {});`,
			Options: []interface{}{
				map[string]interface{}{
					"mustMatch": map[string]interface{}{},
				},
			},
		},
		{
			Code: `it("correctly sets the value", () => {});`,
			Options: []interface{}{
				map[string]interface{}{
					"mustMatch": ` `,
				},
			},
		},
		{
			Code: `it("correctly sets the value", () => {});`,
			Options: []interface{}{
				map[string]interface{}{
					"mustMatch": []interface{}{` `},
				},
			},
		},
		{
			Code: `it("correctly sets the value #unit", () => {});`,
			Options: []interface{}{
				map[string]interface{}{
					"mustMatch": patHashTag,
				},
			},
		},
		{
			Code: `it("correctly sets the value", () => {});`,
			Options: []interface{}{
				map[string]interface{}{
					"mustMatch": patMustMatchHash,
				},
			},
		},
		{
			Code: `it("correctly sets the value", () => {});`,
			Options: []interface{}{
				map[string]interface{}{"mustMatch": map[string]interface{}{"test": patHashTag}},
			},
		},
		{
			Code: `
	        describe('things to test', () => {
	          describe('unit tests #unit', () => {
	            it('is true', () => {
	              expect(true).toBe(true);
	            });
	          });
	
	          describe('e2e tests #e2e', () => {
	            it('is another test #jest4life', () => {});
	          });
	        });
	      `,
			Options: []interface{}{
				map[string]interface{}{"mustMatch": map[string]interface{}{"test": patMustMatchHash}},
			},
		},

		{Code: `it("is a string", () => {});`},
		{Code: `it("is" + " a " + " string", () => {});`},
		{Code: `it(1 + " + " + 1, () => {});`},
		{Code: `it(foo - "bar", () => {});`},
		{Code: `it("is a string" + suffix, () => {});`},
		{Code: `test("is a string", () => {});`},
		{Code: `xtest("is a string", () => {});`},
		{Code: "xtest(`${myFunc} is a string`, () => {});"},
		{Code: `describe("is a string", () => {});`},
		{Code: `describe.skip("is a string", () => {});`},
		{Code: "describe.skip(`${myFunc} is a string`, () => {});"},
		{Code: `fdescribe("is a string", () => {});`},
		{
			Code: `describe(String(/.+/), () => {});`,
			Options: []interface{}{
				map[string]interface{}{"ignoreTypeOfDescribeName": true},
			},
		},
		{
			Code: `it(String(/.+/), () => {});`,
			Options: []interface{}{
				map[string]interface{}{"ignoreTypeOfTestName": true},
			},
		},
		{
			Code: `const foo = "my-title"; it(foo, () => {});`,
			Options: []interface{}{
				map[string]interface{}{"ignoreTypeOfTestName": true},
			},
		},
		{
			Code: `describe(myFunction, () => {});`,
			Options: []interface{}{
				map[string]interface{}{"ignoreTypeOfDescribeName": true},
			},
		},
		{
			Code: `xdescribe(skipFunction, () => {});`,
			Options: []interface{}{
				map[string]interface{}{
					"ignoreTypeOfDescribeName": true,
					"disallowedWords":          []interface{}{},
				},
			},
		},

		{Code: `describe()`},
		{Code: `someFn("", function () {})`},
		{Code: `describe("foo", function () {})`},
		{Code: `describe("foo", function () { it("bar", function () {}) })`},
		{Code: "test(`foo`, function () {})"},
		{Code: "test.concurrent(`foo`, function () {})"},
		{Code: "test(`\\nfoo\\n`, function () {})"},
		{Code: "test(`${foo}`, function () {})"},
		{Code: "test.concurrent(`${foo}`, function () {})"},
		{Code: `it('foo', function () {})`},
		{Code: `it.each([])()`},
		{Code: `it.concurrent('foo', function () {})`},
		{Code: `xdescribe('foo', function () {})`},
		{Code: `xit('foo', function () {})`},
		{Code: `xtest('foo', function () {})`},

		{Code: `it()`},
		{Code: `it.concurrent()`},
		{Code: `it.each()()`},
		{Code: `describe("foo", function () {})`},
		{Code: `fdescribe("foo", function () {})`},
		{Code: `xdescribe("foo", function () {})`},
		{Code: `it("foo", function () {})`},
		{Code: `it.concurrent("foo", function () {})`},
		{Code: `fit("foo", function () {})`},
		{Code: `fit.concurrent("foo", function () {})`},
		{Code: `xit("foo", function () {})`},
		{Code: `test("foo", function () {})`},
		{Code: `test.concurrent("foo", function () {})`},
		{Code: `xtest("foo", function () {})`},
		{Code: "xtest(`foo`, function () {})"},
		{Code: `someFn("foo", function () {})`},
		{Code: `
      describe('foo', () => {
        it('bar', () => {})
      })
    `},
		{
			Code: "it(`GIVEN... \\n  `, () => {});",
			Options: []interface{}{
				map[string]interface{}{"ignoreSpaces": true},
			},
		},

		{Code: "xdescribe(`foo`, function () {})"},
		{Code: `test('foo', function () {})`},
		{Code: `test("foo test", function () {})`},
		{Code: `xtest("foo test", function () {})`},
		{Code: `it("foos it correctly", function () {})`},
		{Code: `
      describe('foo', () => {
        it('describes things correctly', () => {})
      })
    `},
		{Code: "\n      it.each`\n        num   | value\n        ${1} | ${true}\n      `('trues', ({ value }) => {\n        expect(value).toBe(true);\n      });\n    "},

		{Code: `
      test.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i) = %i', (a, b, expected) => {
        expect(a + b).toBe(expected);
      });
    `},
		{Code: `
      test.only.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i)', (a, b, expected) => {
        expect(a + b).toBe(expected);
      });
    `},
		{Code: `
      describe.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i)', (a, b, expected) => {
        it('is correct', () => {
          expect(a + b).toBe(expected);
        });
      });
    `},
		{Code: `
      describe.skip.each([
        [1, 1, 2],
        [1, 2, 3],
        [2, 1, 3],
      ])('.add(%i, %i)', (a, b, expected) => {
        it('is correct', () => {
          expect(a + b).toBe(expected);
        });
      });
    `},
		{Code: `
      test.each([
        {a: 1, b: 1, expected: 2},
        {a: 1, b: 2, expected: 3},
        {a: 2, b: 1, expected: 3},
      ])('.add($a, $b)', ({a, b, expected}) => {
        expect(a + b).toBe(expected);
      });
    `},
		{Code: `it("returns 100%", () => {})`},
		{Code: `it("returns %100%", () => {})`},
		{Code: `it("returns %100", () => {})`},
		{Code: `it("returns a percent%", () => {})`},
		{Code: `it("returns a %percent%", () => {})`},
		{Code: `it("returns a %percent", () => {})`},
		{Code: `it.each([])("%i %p", () => {})`},
		{Code: `it.skip.each([])("%%%", () => {})`},
		{Code: `it.only.each([])("%%%%", () => {})`},
		{Code: `it.failing.each([])("%%%%%", () => {})`},
		{Code: `it.each([])("x%%y", () => {})`},
		{Code: `it.each([])("x %% y", () => {})`},
		{Code: `it.each([])("%int", () => {})`},
		{Code: `
      const x = '%x';

      it.each([])(` + "`has a value of ${x}`" + `, () => {});
    `},
		{Code: `
      const x = '%x';

      it.each([])(` + "`has a value of %${x}`" + `, () => {});
    `},
		{Code: "\n      test.each" + "`\n        a    | b    | expected\n        ${1} | ${1} | ${2}\n        ${1} | ${2} | ${3}\n        ${2} | ${1} | ${3}\n      `('returns $expected when $a is added to $b', ({a, b, expected}) => {\n        expect(a + b).toBe(expected);\n      });\n    "},
		{Code: "\n      test.each" + "`\n        a    | b    | expected\n        ${1} | ${1} | ${2}\n        ${1} | ${2} | ${3}\n        ${2} | ${1} | ${3}\n      `('returns %expected when %a is added to %b', ({a, b, expected}) => {\n        expect(a + b).toBe(expected);\n      });\n    "},
	}
	for _, param := range []string{"p", "s", "d", "i", "f", "j", "o", "#", "$"} {
		valid = append(valid,
			rule_tester.ValidTestCase{Code: fmt.Sprintf(`it.each([])("%%%s", () => {})`, param)},
			rule_tester.ValidTestCase{Code: fmt.Sprintf(`it.each([])("%%%%%s", () => {})`, param)},
			rule_tester.ValidTestCase{Code: fmt.Sprintf(`it.each([])("%%%s%%", () => {})`, param)},
			rule_tester.ValidTestCase{Code: fmt.Sprintf(`it.each([])("%%%%%s%%", () => {})`, param)},
			rule_tester.ValidTestCase{Code: fmt.Sprintf(`it.each([])("%%%s!", () => {})`, param)},
			rule_tester.ValidTestCase{Code: fmt.Sprintf(`
	        test.each([
	          [1, 1, 2],
	          [1, 2, 3],
	          [2, 1, 3],
	        ])('.add(%%%s, %%%s)', (a, b, expected) => {
	          expect(a + b).toBe(expected);
	        });
	      `, param, param)},
			rule_tester.ValidTestCase{Code: fmt.Sprintf(`
	        test.each([
	          [1, 1, 2],
	          [1, 2, 3],
	          [2, 1, 3],
	        ])('.add(%%%s, 1)', (a, b, expected) => {
	          expect(a + b).toBe(expected);
	        });
	      `, param)},
		)
	}
	invalid := []rule_tester.InvalidTestCase{
		{
			Code:    `test("the correct way to properly handle all things", () => {});`,
			Options: []interface{}{map[string]interface{}{"disallowedWords": []interface{}{"correct", "properly", "all"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "disallowedWord"}},
		},
		{
			Code:    `it("is correct", () => {});`,
			Options: []interface{}{map[string]interface{}{"disallowedWords": []interface{}{"correct|properly"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "disallowedWord"}},
		},
		{
			Code: `it("it foo", () => {});`,
			Options: []interface{}{map[string]interface{}{
				"mustMatch": "(unclosed",
			}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidPattern",
				Message:   "Invalid regular expression in `mustMatch` option: `(unclosed`: error parsing regexp: missing closing ) in `(unclosed`",
				Line:      1,
				Column:    1,
			}},
		},
		{
			Code: `it("it foo", () => {});`,
			Options: []interface{}{map[string]interface{}{
				"mustNotMatch": map[string]interface{}{"describe": "(unterminated"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidPattern",
				Message:   "Invalid regular expression in `mustNotMatch.describe` option: `(unterminated`: error parsing regexp: missing closing ) in `(unterminated`",
				Line:      1,
				Column:    1,
			}},
		},
		{
			Code: `it("it foo", () => {});`,
			Options: []interface{}{map[string]interface{}{
				"disallowedWords": []interface{}{"ok", "(unterminated"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidPattern",
				Message:   "Invalid regular expression in `disallowedWords` option: `(?i)\\b(ok|(unterminated)\\b`: error parsing regexp: missing closing ) in `(?i)\\b(ok|(unterminated)\\b`",
				Line:      1,
				Column:    1,
			}},
		},
		{
			Code: `it("it foo", () => {});`,
			Options: []interface{}{map[string]interface{}{
				"disallowedWords": []interface{}{"ok", "(unterminated"},
				"mustMatch":       "^should ",
			}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidPattern",
				Message:   "Invalid regular expression in `disallowedWords` option: `(?i)\\b(ok|(unterminated)\\b`: error parsing regexp: missing closing ) in `(?i)\\b(ok|(unterminated)\\b`",
				Line:      1,
				Column:    1,
			}},
		},
		{
			Code:    `describe("the correct way to do things", function () {})`,
			Options: []interface{}{map[string]interface{}{"disallowedWords": []interface{}{"correct"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "disallowedWord"}},
		},
		{
			Code:    `it("has ALL the things", () => {})`,
			Options: []interface{}{map[string]interface{}{"disallowedWords": []interface{}{"all"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "disallowedWord"}},
		},
		{
			Code:    `xdescribe("every single one of them", function () {})`,
			Options: []interface{}{map[string]interface{}{"disallowedWords": []interface{}{"every"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "disallowedWord"}},
		},
		{
			Code:    `describe('Very Descriptive Title Goes Here', function () {})`,
			Options: []interface{}{map[string]interface{}{"disallowedWords": []interface{}{"descriptive"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "disallowedWord"}},
		},
		{
			Code:    "test(`that the value is set properly`, function () {})",
			Options: []interface{}{map[string]interface{}{"disallowedWords": []interface{}{"properly"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "disallowedWord"}},
		},
		{
			Code: `
	        describe('things to test', () => {
	          describe('unit tests #unit', () => {
	            it('is true', () => {
	              expect(true).toBe(true);
	            });
	          });
	
	          describe('e2e tests #e4e', () => {
	            it('is another test #e2e #jest4life', () => {});
	          });
	        });
	      `,
			Options: []interface{}{map[string]interface{}{"mustNotMatch": patMustNotHashBad, "mustMatch": patMustMatchHash}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "mustNotMatch"},
				{MessageId: "mustNotMatch"},
			},
		},
		{
			Code: `
	        import { describe, describe as context, it as thisTest } from '@jest/globals';
	
	        describe('things to test', () => {
	          context('unit tests #unit', () => {
	            thisTest('is true', () => {
	              expect(true).toBe(true);
	            });
	          });
	
	          context('e2e tests #e4e', () => {
	            thisTest('is another test #e2e #jest4life', () => {});
	          });
	        });
	      `,
			Options: []interface{}{map[string]interface{}{"mustNotMatch": patMustNotHashBad, "mustMatch": patMustMatchHash}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "mustNotMatch"},
				{MessageId: "mustNotMatch"},
			},
		},
		{
			Code: `
	        describe('things to test', () => {
	          describe('unit tests #unit', () => {
	            it('is true', () => {
	              expect(true).toBe(true);
	            });
	          });
	
	          describe('e2e tests #e4e', () => {
	            it('is another test #e2e #jest4life', () => {});
	          });
	        });
	      `,
			Options: []interface{}{map[string]interface{}{
				"mustNotMatch": []interface{}{patMustNotHashBad, `Please include "#unit" or "#e2e" in titles`},
				"mustMatch":    []interface{}{patMustMatchHash, `Please include "#unit" or "#e2e" in titles`},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "mustNotMatchCustom"},
				{MessageId: "mustNotMatchCustom"},
			},
		},
		{
			Code: `
	        describe('things to test', () => {
	          describe('unit tests #unit', () => {
	            it('is true', () => {
	              expect(true).toBe(true);
	            });
	          });
	
	          describe('e2e tests #e4e', () => {
	            it('is another test #e2e #jest4life', () => {});
	          });
	        });
	      `,
			Options: []interface{}{map[string]interface{}{
				"mustNotMatch": map[string]interface{}{"describe": []interface{}{patMustNotHashBad}},
				"mustMatch":    map[string]interface{}{"describe": patMustMatchHash},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "mustNotMatch"},
			},
		},
		{
			Code: `
	        describe('things to test', () => {
	          describe('unit tests #unit', () => {
	            it('is true', () => {
	              expect(true).toBe(true);
	            });
	          });
	
	          describe('e2e tests #e4e', () => {
	            it('is another test #e2e #jest4life', () => {});
	          });
	        });
	      `,
			Options: []interface{}{map[string]interface{}{
				"mustNotMatch": map[string]interface{}{
					"describe": []interface{}{patMustNotHashBad, `Please include "#unit" or "#e2e" in describe titles`},
				},
				"mustMatch": map[string]interface{}{"describe": patMustMatchHash},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "mustNotMatchCustom"},
			},
		},
		{
			Code: `
	        describe('things to test', () => {
	          describe('unit tests #unit', () => {
	            it('is true', () => {
	              expect(true).toBe(true);
	            });
	          });
	
	          describe('e2e tests #e4e', () => {
	            it('is another test #e2e #jest4life', () => {});
	          });
	        });
	      `,
			Options: []interface{}{map[string]interface{}{
				"mustNotMatch": map[string]interface{}{"describe": patMustNotHashBad},
				"mustMatch":    map[string]interface{}{"it": patMustMatchHash},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "mustNotMatch"},
			},
		},
		{
			Code: `
	        describe('things to test', () => {
	          describe('unit tests #unit', () => {
	            it('is true #jest4life', () => {
	              expect(true).toBe(true);
	            });
	          });
	
	          describe('e2e tests #e4e', () => {
	            it('is another test #e2e #jest4life', () => {});
	          });
	        });
	      `,
			Options: []interface{}{map[string]interface{}{
				"mustNotMatch": map[string]interface{}{
					"describe": []interface{}{patMustNotHashBad, `Please include "#unit" or "#e2e" in describe titles`},
				},
				"mustMatch": map[string]interface{}{
					"it": []interface{}{patMustMatchHash, `Please include "#unit" or "#e2e" in it titles`},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "mustMatchCustom"},
				{MessageId: "mustNotMatchCustom"},
			},
		},
		{
			Code: `test("the correct way to properly handle all things", () => {});`,
			Options: []interface{}{
				map[string]interface{}{"mustMatch": patHashTag},
			},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mustMatch",
				Message:   `test should match /#(?:unit|integration|e2e)/u`,
			}},
		},
		{
			Code: `describe("the test", () => {});`,
			Options: []interface{}{
				map[string]interface{}{
					"mustMatch": map[string]interface{}{"describe": patHashTag},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "mustMatch"}},
		},
		{
			Code: `xdescribe("the test", () => {});`,
			Options: []interface{}{
				map[string]interface{}{
					"mustMatch": map[string]interface{}{"describe": patHashTag},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "mustMatch"}},
		},
		{
			Code: `describe.skip("the test", () => {});`,
			Options: []interface{}{
				map[string]interface{}{
					"mustMatch": map[string]interface{}{"describe": patHashTag},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "mustMatch"}},
		},
		{
			Code:   `it.each([])(1, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:    `it(String(/.+/), () => {});`,
			Options: []interface{}{map[string]interface{}{"ignoreTypeOfTestName": false}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:    `const foo = "my-title"; it(foo, () => {});`,
			Options: []interface{}{map[string]interface{}{"ignoreTypeOfTestName": false}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `it.skip.each([])(1, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   "it.skip.each``(1, () => {});",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `it(123, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `it.concurrent(123, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `it(1 + 2 + 3, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `it('a' || 'b', () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `it(cond && 'b', () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `it(x = 'foo', () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   "it(foo + `suffix${x}`, () => {});",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   "it(`prefix${x}` + foo, () => {});",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `it.concurrent(1 + 2 + 3, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:    `test.skip(123, () => {});`,
			Options: []interface{}{map[string]interface{}{"ignoreTypeOfDescribeName": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `describe(String(/.+/), () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:    `describe(myFunction, () => 1);`,
			Options: []interface{}{map[string]interface{}{"ignoreTypeOfDescribeName": false}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `describe(myFunction, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `xdescribe(myFunction, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `describe(6, function () {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `describe.skip(123, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code: `describe("", function () {})`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "emptyTitle",
				Line:      1,
				Column:    1,
			}},
		},
		{
			Code: `
	        describe('foo', () => {
	          it('', () => {});
	        });
	      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},
		{
			Code:   `it("", function () {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},
		{
			Code:   `it.concurrent("", function () {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},
		{
			Code:   `test("", function () {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},
		{
			Code:   `test.concurrent("", function () {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},
		{
			Code:   "test(``, function () {})",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},
		{
			Code:   "test.concurrent(``, function () {})",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},
		{
			Code:   `xdescribe('', () => {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},
		{
			Code:   `xit('', () => {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},
		{
			Code:   `xtest('', () => {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},
		{
			Code:   `describe(" foo", function () {})`,
			Output: []string{`describe("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `describe.each()(" foo", function () {})`,
			Output: []string{`describe.each()("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `describe.only.each()(" foo", function () {})`,
			Output: []string{`describe.only.each()("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `describe(" foo foe fum", function () {})`,
			Output: []string{`describe("foo foe fum", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `describe("foo foe fum ", function () {})`,
			Output: []string{`describe("foo foe fum", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `fdescribe(" foo", function () {})`,
			Output: []string{`fdescribe("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `fdescribe(' foo', function () {})`,
			Output: []string{`fdescribe('foo', function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `xdescribe(" foo", function () {})`,
			Output: []string{`xdescribe("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `it(" foo", function () {})`,
			Output: []string{`it("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `it.concurrent(" foo", function () {})`,
			Output: []string{`it.concurrent("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `fit(" foo", function () {})`,
			Output: []string{`fit("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `it.skip(" foo", function () {})`,
			Output: []string{`it.skip("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `fit("foo ", function () {})`,
			Output: []string{`fit("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `it.skip("foo ", function () {})`,
			Output: []string{`it.skip("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code: `
	        import { test as testThat } from '@jest/globals';
	
	        testThat('foo works ', () => {});
	      `,
			Output: []string{`
	        import { test as testThat } from '@jest/globals';
	
	        testThat('foo works', () => {});
	      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `xit(" foo", function () {})`,
			Output: []string{`xit("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `test(" foo", function () {})`,
			Output: []string{`test("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `test.concurrent(" foo", function () {})`,
			Output: []string{`test.concurrent("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   "test(` foo`, function () {})",
			Output: []string{"test(`foo`, function () {})"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   "test.concurrent(` foo`, function () {})",
			Output: []string{"test.concurrent(`foo`, function () {})"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   "test(` foo bar bang`, function () {})",
			Output: []string{"test(`foo bar bang`, function () {})"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   "test.concurrent(` foo bar bang`, function () {})",
			Output: []string{"test.concurrent(`foo bar bang`, function () {})"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   "test(` foo bar bang  `, function () {})",
			Output: []string{"test(`foo bar bang`, function () {})"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   "test.concurrent(` foo bar bang  `, function () {})",
			Output: []string{"test.concurrent(`foo bar bang`, function () {})"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `xtest(" foo", function () {})`,
			Output: []string{`xtest("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `xtest(" foo  ", function () {})`,
			Output: []string{`xtest("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code: `
	        describe(' foo', () => {
	          it('bar', () => {})
	        })
	      `,
			Output: []string{`
	        describe('foo', () => {
	          it('bar', () => {})
	        })
	      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code: `
	        describe('foo', () => {
	          it(' bar', () => {})
	        })
	      `,
			Output: []string{`
	        describe('foo', () => {
	          it('bar', () => {})
	        })
	      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `describe("describe foo", function () {})`,
			Output: []string{`describe("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `fdescribe("describe foo", function () {})`,
			Output: []string{`fdescribe("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `xdescribe("describe foo", function () {})`,
			Output: []string{`xdescribe("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `describe('describe foo', function () {})`,
			Output: []string{`describe('foo', function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `test("test foo", function () {})`,
			Output: []string{`test("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `xtest("test foo", function () {})`,
			Output: []string{`xtest("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `it("it foo", function () {})`,
			Output: []string{`it("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `it("it   foo", function () {})`,
			Output: []string{`it("  foo", function () {})`, `it("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   "describe('describe \tfoo', () => {})",
			Output: []string{"describe('\tfoo', () => {})"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `fit("it foo", function () {})`,
			Output: []string{`fit("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `xit("it foo", function () {})`,
			Output: []string{`xit("foo", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `it("it foos it correctly", function () {})`,
			Output: []string{`it("foos it correctly", function () {})`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code: `
	        describe('describe foo', () => {
	          it('bar', () => {})
	        })
	      `,
			Output: []string{`
	        describe('foo', () => {
	          it('bar', () => {})
	        })
	      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code: `
	        describe('describe foo', () => {
	          it('describes things correctly', () => {})
	        })
	      `,
			Output: []string{`
	        describe('foo', () => {
	          it('describes things correctly', () => {})
	        })
	      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code: `
	        describe('foo', () => {
	          it('it bar', () => {})
	        })
	      `,
			Output: []string{`
	        describe('foo', () => {
	          it('bar', () => {})
	        })
	      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `it.each([])("%y", () => {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		{
			Code:   `it.each([])("%s+%y", () => {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		{
			Code:   `it.each([])("%y+%%y", () => {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		{
			Code:   `it.each([])("%%y+%y", () => {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		{
			Code:   `it.each([])("%%%x", () => {})`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		{Code: `
        test.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])('.add(%x, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}}},
		{Code: `
        test.only.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])('.add(%x, %y)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}}},
		{Code: `
        describe.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])('.add(%x, %y)', (a, b, expected) => {
          it('is correct', () => {
            expect(a + b).toBe(expected);
          });
        });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}}},
		{Code: `
        test.each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ])('.add(%i, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}}},
		{Code: `
        it.skip.each(entries)('.add(%i, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}}},
		{Code: `
        const entries = [
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ];

        test.each(entries)('.add(%i, %x)', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}}},
	}
	for _, letter := range []string{"P", "S", "D", "I", "F", "J", "O"} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:   fmt.Sprintf(`it.each([])("%%%s", () => {})`, letter),
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		})
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&valid_title.ValidTitleRule,
		valid,
		invalid,
	)
}
