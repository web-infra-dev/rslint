package no_test_prefixes_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_test_prefixes"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoTestPrefixesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_test_prefixes.NoTestPrefixesRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe("foo", function () {})`},
			{Code: `it("foo", function () {})`},
			{Code: `it.concurrent("foo", function () {})`},
			{Code: `test("foo", function () {})`},
			{Code: `test.concurrent("foo", function () {})`},
			{Code: `describe.only("foo", function () {})`},
			{Code: `it.only("foo", function () {})`},
			{Code: `it.each()("foo", function () {})`},
			{Code: "it.each``(\"foo\", function () {})"},
			{Code: `test.only("foo", function () {})`},
			{Code: `test.each()("foo", function () {})`},
			{Code: "test.each``(\"foo\", function () {})"},
			{Code: `describe.skip("foo", function () {})`},
			{Code: `it.skip("foo", function () {})`},
			{Code: `test.skip("foo", function () {})`},
			{Code: `foo()`},
			{Code: `[1,2,3].forEach()`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `fdescribe("foo", function () {})`,
				Output: []string{`describe.only("foo", function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 1, Column: 1},
				},
			},
			{
				Code:   `xdescribe.each([])("foo", function () {})`,
				Output: []string{`describe.skip.each([])("foo", function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 1, Column: 1},
				},
			},
			{
				Code:   `fit("foo", function () {})`,
				Output: []string{`it.only("foo", function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 1, Column: 1},
				},
			},
			{
				Code:   `xdescribe("foo", function () {})`,
				Output: []string{`describe.skip("foo", function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 1, Column: 1},
				},
			},
			{
				Code:   `xit("foo", function () {})`,
				Output: []string{`it.skip("foo", function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 1, Column: 1},
				},
			},
			{
				Code:   `xtest("foo", function () {})`,
				Output: []string{`test.skip("foo", function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 1, Column: 1},
				},
			},
			{
				Code:   "xit.each``(\"foo\", function () {})",
				Output: []string{"it.skip.each``(\"foo\", function () {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 1, Column: 1},
				},
			},
			{
				Code:   "xtest.each``(\"foo\", function () {})",
				Output: []string{"test.skip.each``(\"foo\", function () {})"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 1, Column: 1},
				},
			},
			{
				Code:   `xit.each([])("foo", function () {})`,
				Output: []string{`it.skip.each([])("foo", function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 1, Column: 1},
				},
			},
			{
				Code:   `xtest.each([])("foo", function () {})`,
				Output: []string{`test.skip.each([])("foo", function () {})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 1, Column: 1},
				},
			},
			{
				Code: `
import { xit } from '@jest/globals';

xit("foo", function () {})
`,
				Output: []string{`
import { xit } from '@jest/globals';

it.skip("foo", function () {})
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 4, Column: 1},
				},
			},
			{
				Code: `
import { xit as skipThis } from '@jest/globals';

skipThis("foo", function () {})
`,
				Output: []string{`
import { xit as skipThis } from '@jest/globals';

it.skip("foo", function () {})
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 4, Column: 1},
				},
			},
			{
				Code: `
import { fit as onlyThis } from '@jest/globals';

onlyThis("foo", function () {})
`,
				Output: []string{`
import { fit as onlyThis } from '@jest/globals';

it.only("foo", function () {})
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usePreferredName", Line: 4, Column: 1},
				},
			},
		},
	)
}
