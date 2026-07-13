package no_restricted_jest_methods_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_restricted_jest_methods"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRestrictedJestMethodsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_restricted_jest_methods.NoRestrictedJestMethodsRule,
		[]rule_tester.ValidTestCase{
			{Code: `jest`},
			{Code: `jest()`},
			{Code: `jest.mock()`},
			{Code: `expect(a).rejects;`},
			{Code: `expect(a);`},
			{
				Code: `import { jest } from '@jest/globals';

jest;
`,
			},
			{
				Code: `const jest = { fn: () => {} };
jest.fn();`,
				Options: []interface{}{
					map[string]interface{}{"fn": nil},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `jest.fn()`,
				Options: []interface{}{
					map[string]interface{}{"fn": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedJestMethod",
						Message:   "Use of `fn` is disallowed",
						Line:      1,
						Column:    6,
					},
				},
			},
			{
				Code: `jest["fn"]()`,
				Options: []interface{}{
					map[string]interface{}{"fn": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedJestMethod",
						Message:   "Use of `fn` is disallowed",
						Line:      1,
						Column:    6,
					},
				},
			},
			{
				Code: `jest.fn()`,
				Options: []interface{}{
					map[string]interface{}{"fn": ""},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedJestMethod",
						Message:   "Use of `fn` is disallowed",
						Line:      1,
						Column:    6,
					},
				},
			},
			{
				Code: `jest.fn().mockImplementation(() => {})`,
				Options: []interface{}{
					map[string]interface{}{"fn": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedJestMethod",
						Message:   "Use of `fn` is disallowed",
						Line:      1,
						Column:    6,
					},
				},
			},
			{
				Code: `(jest.fn)()`,
				Options: []interface{}{
					map[string]interface{}{"fn": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedJestMethod",
						Message:   "Use of `fn` is disallowed",
						Line:      1,
						Column:    7,
					},
				},
			},
			{
				Code: `jest.mock()`,
				Options: []interface{}{
					map[string]interface{}{"mock": "Do not use mocks"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedJestMethodWithMessage",
						Message:   "Do not use mocks",
						Line:      1,
						Column:    6,
					},
				},
			},
			{
				Code: `jest["mock"]()`,
				Options: []interface{}{
					map[string]interface{}{"mock": "Do not use mocks"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedJestMethodWithMessage",
						Message:   "Do not use mocks",
						Line:      1,
						Column:    6,
					},
				},
			},
			{
				Code: `import { jest } from '@jest/globals';

jest.advanceTimersByTime();
`,
				Options: []interface{}{
					map[string]interface{}{"advanceTimersByTime": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedJestMethod",
						Message:   "Use of `advanceTimersByTime` is disallowed",
						Line:      3,
						Column:    6,
					},
				},
			},
			{
				Code: `const { jest } = require('@jest/globals');

jest.advanceTimersByTime();
`,
				Options: []interface{}{
					map[string]interface{}{"advanceTimersByTime": nil},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "restrictedJestMethod",
						Message:   "Use of `advanceTimersByTime` is disallowed",
						Line:      3,
						Column:    6,
					},
				},
			},
		},
	)
}
