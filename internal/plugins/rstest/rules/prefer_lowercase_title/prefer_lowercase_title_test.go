package prefer_lowercase_title_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_lowercase_title"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferLowercaseTitleRstest(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_lowercase_title.PreferLowercaseTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: `test()`},
			{Code: `test('foo', function () {})`},
			{Code: `test("foo", function () {})`},
			{Code: "test(`foo`, function () {})"},
			{Code: `test(42)`},
			{Code: `test("")`},
			{Code: `describe()`},
			{Code: `describe('foo', function () {})`},
			{Code: `describe("foo", function () {})`},
			{Code: "describe(`foo`, function () {})"},
			{Code: `describe(42)`},
			{Code: `describe("")`},
			{Code: `it('foo', function () {})`},
			{Code: `it()`},
			{Code: `randomFunction()`},
			{Code: `foo.bar()`},
			// imported from rstest
			{Code: "import { test } from 'rstack/test';\ntest('foo', () => {});"},
			{Code: "import { describe } from 'rstack/test';\ndescribe('foo', () => {});"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `test('Foo', function () {})`,
				Output: []string{`test('foo', function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 6},
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
				Code:   `it('Foo', function () {})`,
				Output: []string{`it('foo', function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 4},
				},
			},
			{
				Code:   `test('Doesn\'t mutate', () => {})`,
				Output: []string{`test('doesn\'t mutate', () => {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 6},
				},
			},
			{
				Code:   `const todoTest = test.todo; todoTest('Should work');`,
				Output: []string{`const todoTest = test.todo; todoTest('should work');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 1, Column: 38},
				},
			},
			{
				Code:   "import { test } from 'rstack/test';\ntest('Foo', () => {});",
				Output: []string{"import { test } from 'rstack/test';\ntest('foo', () => {});"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 2, Column: 6},
				},
			},
			{
				Code:   "import { describe } from 'rstack/test';\ndescribe('Foo', () => {});",
				Output: []string{"import { describe } from 'rstack/test';\ndescribe('foo', () => {});"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 2, Column: 10},
				},
			},
		},
	)
}

func TestPreferLowercaseTitleRstestIgnoreTodos(t *testing.T) {
	opts := []any{map[string]interface{}{"ignoreTodos": true}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_lowercase_title.PreferLowercaseTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: "test.todo(`Foo`, function () {})", Options: opts},
			{Code: `test.todo("Foo", () => {})`, Options: opts},
			{Code: `const todoTest = test.todo; todoTest('Should work');`, Options: opts},
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
		},
	)
}

func TestPreferLowercaseTitleRstestIgnoreTopLevelDescribe(t *testing.T) {
	opts := []any{map[string]interface{}{"ignoreTopLevelDescribe": true}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_lowercase_title.PreferLowercaseTitleRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe("MyClass", () => {});`, Options: opts},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    "describe('MyClass', () => {\n  describe('MyMethod', () => {\n    test('Does things', () => {});\n  });\n});",
				Output:  []string{"describe('MyClass', () => {\n  describe('myMethod', () => {\n    test('does things', () => {});\n  });\n});"},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 2, Column: 12},
					{MessageId: "unexpectedCase", Line: 3, Column: 10},
				},
			},
			{
				Code:    "describe('Outer', suiteBody);\n\nfunction suiteBody() {\n  describe('Inner', () => {});\n}",
				Output:  []string{"describe('Outer', suiteBody);\n\nfunction suiteBody() {\n  describe('inner', () => {});\n}"},
				Options: opts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCase", Line: 4, Column: 12},
				},
			},
		},
	)
}
