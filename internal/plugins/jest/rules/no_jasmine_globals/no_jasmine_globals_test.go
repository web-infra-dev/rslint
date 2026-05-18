package no_jasmine_globals_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_jasmine_globals"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoJasmineGlobalsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_jasmine_globals.NoJasmineGlobalsRule,
		[]rule_tester.ValidTestCase{
			{Code: `jest.spyOn()`},
			{Code: `jest.fn()`},
			{Code: `expect.extend()`},
			{Code: `expect.any()`},
			{Code: `it("foo", function () {})`},
			{Code: `test("foo", function () {})`},
			{Code: `foo()`},
			{Code: `require("foo")("bar")`},
			{Code: `(function(){})()`},
			{Code: `function callback(fail) { fail() }`},
			{Code: `var spyOn = require("actions"); spyOn("foo")`},
			{Code: `function callback(pending) { pending() }`},
			{Code: `jasmine.DEFAULT_TIMEOUT_INTERVAL += 1000;`},
			{Code: `jasmine["DEFAULT_TIMEOUT_INTERVAL"] += 1000;`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `spyOn(some, "object")`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalGlobal",
					Message:   "Illegal usage of global `spyOn`, prefer `jest.spyOn`",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `spyOnProperty(some, "object")`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalGlobal",
					Message:   "Illegal usage of global `spyOnProperty`, prefer `jest.spyOn`",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `fail()`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalFail",
					Message:   "Illegal usage of `fail`, prefer throwing an error, or the `done.fail` callback",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `pending()`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalPending",
					Message:   "Illegal usage of `pending`, prefer explicitly skipping a test using `test.skip`",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.DEFAULT_TIMEOUT_INTERVAL = 5000;`,
				Output: []string{`jest.setTimeout(5000);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.DEFAULT_TIMEOUT_INTERVAL = function() {}`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine["DEFAULT_TIMEOUT_INTERVAL"] = 5000;`,
				Output: []string{`jest.setTimeout(5000);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine["DEFAULT_TIMEOUT_INTERVAL"] = function() {}`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.addMatchers(matchers)`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalMethod",
					Message:   "Illegal usage of `jasmine.addMatchers`, prefer `expect.extend`",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.createSpy()`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalMethod",
					Message:   "Illegal usage of `jasmine.createSpy`, prefer `jest.fn`",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.any()`,
				Output: []string{`expect.any()`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalMethod",
					Message:   "Illegal usage of `jasmine.any`, prefer `expect.any`",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.anything()`,
				Output: []string{`expect.anything()`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalMethod",
					Message:   "Illegal usage of `jasmine.anything`, prefer `expect.anything`",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.arrayContaining()`,
				Output: []string{`expect.arrayContaining()`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalMethod",
					Message:   "Illegal usage of `jasmine.arrayContaining`, prefer `expect.arrayContaining`",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.objectContaining()`,
				Output: []string{`expect.objectContaining()`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalMethod",
					Message:   "Illegal usage of `jasmine.objectContaining`, prefer `expect.objectContaining`",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.stringMatching()`,
				Output: []string{`expect.stringMatching()`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalMethod",
					Message:   "Illegal usage of `jasmine.stringMatching`, prefer `expect.stringMatching`",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.getEnv()`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.empty()`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.falsy()`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.truthy()`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.arrayWithExactContents()`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.clock()`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine.MAX_PRETTY_PRINT_ARRAY_LENGTH = 42`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
			{
				Code:   `jasmine["MAX_PRETTY_PRINT_ARRAY_LENGTH"] = 42`,
				Output: []string{},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "illegalJasmine",
					Message:   "Illegal usage of jasmine global",
					Column:    1,
					Line:      1,
				}},
			},
		},
	)
}
