package prefer_lowercase_title_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_lowercase_title"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferLowercaseTitleRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_lowercase_title.PreferLowercaseTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: `it.each()`},
			{Code: `it.each()(1)`},
			{Code: `randomFunction()`},
			{Code: `foo.bar()`},
			{Code: `it()`},
			{Code: `it(' ', function () {})`},
			{Code: `it(true, function () {})`},
			{Code: `it(MY_CONSTANT, function () {})`},
			{Code: `it(" ", function () {})`},
			{Code: "it(` `, function () {})"},
			{Code: `it('foo', function () {})`},
			{Code: `it("foo", function () {})`},
			{Code: "it(`foo`, function () {})"},
			{Code: `it("<Foo/>", function () {})`},
			{Code: `it("123 foo", function () {})`},
			{Code: `it(42, function () {})`},
			{Code: "it(``)"},
			{Code: `it("")`},
			{Code: `it(42)`},
			{Code: `test()`},
			{Code: `test('foo', function () {})`},
			{Code: `test("foo", function () {})`},
			{Code: "test(`foo`, function () {})"},
			{Code: `test("<Foo/>", function () {})`},
			{Code: `test("123 foo", function () {})`},
			{Code: `test("42", function () {})`},
			{Code: "test(``)"},
			{Code: `test("")`},
			{Code: `test(42)`},
			{Code: `describe()`},
			{Code: `describe('foo', function () {})`},
			{Code: `describe("foo", function () {})`},
			{Code: "describe(`foo`, function () {})"},
			{Code: `describe("<Foo/>", function () {})`},
			{Code: `describe("123 foo", function () {})`},
			{Code: `describe("42", function () {})`},
			{Code: `describe(function () {})`},
			{Code: "describe(``)"},
			{Code: `describe("")`},
			{Code: "describe.each()(1);\ndescribe.each()(2);"},
			{Code: `jest.doMock("my-module")`},
			{Code: "import { jest } from '@jest/globals';\n\njest.doMock('my-module');"},
			{Code: `describe(42)`},
			{
				Code:    `describe(42)`,
				Options: []any{map[string]interface{}{}},
			},
			// extras: already-lowercase Unicode title
			{Code: "it('über', function () {})"},
			// extras: parenthesized string argument
			{Code: "it(('foo'), function () {})"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `it('Foo', function () {})`,
				Output: []string{`it('foo', function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 4},
				},
			},
			{
				Code:   `xit('Foo', function () {})`,
				Output: []string{`xit('foo', function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 5},
				},
			},
			{
				Code:   `it("Foo", function () {})`,
				Output: []string{`it("foo", function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 4},
				},
			},
			{
				Code:   "it(`Foo`, function () {})",
				Output: []string{"it(`foo`, function () {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 4},
				},
			},
			{
				Code:   `test('Foo', function () {})`,
				Output: []string{`test('foo', function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 6},
				},
			},
			{
				Code:   `xtest('Foo', function () {})`,
				Output: []string{`xtest('foo', function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 7},
				},
			},
			{
				Code:   `test("Foo", function () {})`,
				Output: []string{`test("foo", function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 6},
				},
			},
			{
				Code:   "test(`Foo`, function () {})",
				Output: []string{"test(`foo`, function () {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 6},
				},
			},
			{
				Code:   `describe('Foo', function () {})`,
				Output: []string{`describe('foo', function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 10},
				},
			},
			{
				Code:   `describe("Foo", function () {})`,
				Output: []string{`describe("foo", function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 10},
				},
			},
			{
				Code:   "describe(`Foo`, function () {})",
				Output: []string{"describe(`foo`, function () {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 10},
				},
			},
			{
				Code:   "import { describe as context } from '@jest/globals';\n\ncontext(`Foo`, () => {});",
				Output: []string{"import { describe as context } from '@jest/globals';\n\ncontext(`foo`, () => {});"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 3, Column: 9},
				},
			},
			{
				Code:   "describe(`Some longer description`, function () {})",
				Output: []string{"describe(`some longer description`, function () {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 10},
				},
			},
			{
				Code:   "fdescribe(`Some longer description`, function () {})",
				Output: []string{"fdescribe(`some longer description`, function () {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 11},
				},
			},
			{
				Code:   "it.each(['green', 'black'])('Should return %', () => {})",
				Output: []string{"it.each(['green', 'black'])('should return %', () => {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 29},
				},
			},
			{
				Code:   "describe.each(['green', 'black'])('Should return %', () => {})",
				Output: []string{"describe.each(['green', 'black'])('should return %', () => {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 35},
				},
			},
			// extras: uppercase Unicode title
			{
				Code:   "it('Über', function () {})",
				Output: []string{"it('über', function () {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 4},
				},
			},
			// extras: parenthesized string argument
			{
				Code:   "it(('Foo'), function () {})",
				Output: []string{"it(('foo'), function () {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 5},
				},
			},
		},
	)
}

func TestPreferLowercaseTitleIgnoreDescribe(t *testing.T) {
	opts := []any{map[string]interface{}{"ignore": []interface{}{"describe"}}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_lowercase_title.PreferLowercaseTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe('Foo', function () {})`, Options: opts},
			{Code: `describe("Foo", function () {})`, Options: opts},
			{Code: "describe(`Foo`, function () {})", Options: opts},
			{Code: "fdescribe(`Foo`, function () {})", Options: opts},
			{Code: "describe.skip(`Foo`, function () {})", Options: opts},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `test('Foo', function () {})`,
				Output:  []string{`test('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 6},
				},
			},
			{
				Code:    `xit('Foo', function () {})`,
				Output:  []string{`xit('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 5},
				},
			},
		},
	)
}

func TestPreferLowercaseTitleIgnoreTest(t *testing.T) {
	opts := []any{map[string]interface{}{"ignore": []interface{}{"test"}}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_lowercase_title.PreferLowercaseTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: `test('Foo', function () {})`, Options: opts},
			{Code: `test("Foo", function () {})`, Options: opts},
			{Code: "test(`Foo`, function () {})", Options: opts},
			{Code: "xtest(`Foo`, function () {})", Options: opts},
			{Code: "test.only(`Foo`, function () {})", Options: opts},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `describe('Foo', function () {})`,
				Output:  []string{`describe('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 10},
				},
			},
			{
				Code:    `it('Foo', function () {})`,
				Output:  []string{`it('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 4},
				},
			},
			{
				Code:    `xit('Foo', function () {})`,
				Output:  []string{`xit('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 5},
				},
			},
		},
	)
}

func TestPreferLowercaseTitleIgnoreIt(t *testing.T) {
	opts := []any{map[string]interface{}{"ignore": []interface{}{"it"}}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_lowercase_title.PreferLowercaseTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: `it('Foo', function () {})`, Options: opts},
			{Code: `it("Foo", function () {})`, Options: opts},
			{Code: "it(`Foo`, function () {})", Options: opts},
			{Code: "fit(`Foo`, function () {})", Options: opts},
			{Code: "it.skip(`Foo`, function () {})", Options: opts},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `describe('Foo', function () {})`,
				Output:  []string{`describe('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 10},
				},
			},
			{
				Code:    `test('Foo', function () {})`,
				Output:  []string{`test('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 6},
				},
			},
			{
				Code:    `xtest('Foo', function () {})`,
				Output:  []string{`xtest('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 7},
				},
			},
		},
	)
}

func TestPreferLowercaseTitleAllowedPrefixes(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_lowercase_title.PreferLowercaseTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: `it('GET /live', function () {})`, Options: []any{map[string]interface{}{"allowedPrefixes": []interface{}{"GET"}}}},
			{Code: `it("POST /live", function () {})`, Options: []any{map[string]interface{}{"allowedPrefixes": []interface{}{"GET", "POST"}}}},
			{Code: "it(`PATCH /live`, function () {})", Options: []any{map[string]interface{}{"allowedPrefixes": []interface{}{"GET", "PATCH"}}}},
		},
		nil,
	)
}

func TestPreferLowercaseTitleIgnoreTopLevelDescribe(t *testing.T) {
	opts := []any{map[string]interface{}{"ignoreTopLevelDescribe": true}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_lowercase_title.PreferLowercaseTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe("MyClass", () => {});`, Options: opts},
			{
				Code: "describe('MyClass', () => {\n  describe('#myMethod', () => {\n    it('does things', () => {});\n  });\n});",
				Options: opts,
			},
			{
				Code: "describe('Strings', () => {\n  it('are strings', () => { expect('abc').toBe('abc'); });\n});\n\ndescribe('Booleans', () => {\n  it('are booleans', () => { expect(true).toBe(true); });\n});",
				Options: opts,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `it("Works!", () => {});`,
				Output:  []string{`it("works!", () => {});`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 4},
				},
			},
			{
				Code: "describe('MyClass', () => {\n  describe('MyMethod', () => {\n    it('Does things', () => {});\n  });\n});",
				Output: []string{"describe('MyClass', () => {\n  describe('myMethod', () => {\n    it('does things', () => {});\n  });\n});"},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 2, Column: 12},
					{MessageId: "unexpectedCase", Line: 3, Column: 8},
				},
			},
			{
				Code: "import { describe, describe as context } from '@jest/globals';\n\ndescribe('MyClass', () => {\n  context('MyMethod', () => {\n    it('Does things', () => {});\n  });\n});",
				Output: []string{"import { describe, describe as context } from '@jest/globals';\n\ndescribe('MyClass', () => {\n  context('myMethod', () => {\n    it('does things', () => {});\n  });\n});"},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 4, Column: 11},
					{MessageId: "unexpectedCase", Line: 5, Column: 8},
				},
			},
		},
	)
}

func TestPreferLowercaseTitleIgnoreTopLevelDescribeFalse(t *testing.T) {
	opts := []any{map[string]interface{}{"ignoreTopLevelDescribe": false}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_lowercase_title.PreferLowercaseTitleRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: "describe('MyClass', () => {\n  describe('MyMethod', () => {\n    it('Does things', () => {});\n  });\n});",
				Output: []string{"describe('myClass', () => {\n  describe('myMethod', () => {\n    it('does things', () => {});\n  });\n});"},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 10},
					{MessageId: "unexpectedCase", Line: 2, Column: 12},
					{MessageId: "unexpectedCase", Line: 3, Column: 8},
				},
			},
		},
	)
}

func TestPreferLowercaseTitleIgnoreTodos(t *testing.T) {
	opts := []any{map[string]interface{}{"ignoreTodos": true}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_lowercase_title.PreferLowercaseTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: "test.todo(`Foo`, function () {})", Options: opts},
			{Code: "it.todo(`Foo`, function () {})", Options: opts},
			{Code: `it.todo("Foo", () => {})`, Options: opts},
			{Code: `it.only.todo("Foo", () => {})`, Options: opts},
			{Code: `it.todo.only("Foo", () => {})`, Options: opts},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `describe('Foo', function () {})`,
				Output:  []string{`describe('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 10},
				},
			},
			{
				Code:    `it('Foo', function () {})`,
				Output:  []string{`it('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 4},
				},
			},
			{
				Code:    `test('Foo', function () {})`,
				Output:  []string{`test('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 6},
				},
			},
			{
				Code:    `test.only('Foo', function () {})`,
				Output:  []string{`test.only('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 11},
				},
			},
			{
				Code:    `test.skip('Foo', function () {})`,
				Output:  []string{`test.skip('foo', function () {})`},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 11},
				},
			},
		},
	)
}
