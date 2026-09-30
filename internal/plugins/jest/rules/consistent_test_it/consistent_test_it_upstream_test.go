package consistent_test_it_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/consistent_test_it"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Cases from eslint-plugin-jest v29.16.6
// src/rules/__tests__/consistent-test-it.test.ts, one subtest per upstream
// ruleTester.run group, plus the examples in docs/rules/consistent-test-it.md.
// Positions and outputs were checked against the upstream rule.
func TestConsistentTestItUpstream(t *testing.T) {
	t.Run("consistent-test-it with fn=test", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "test(\"foo\")", Options: map[string]any{"fn": "test"}},
				{Code: "test.only(\"foo\")", Options: map[string]any{"fn": "test"}},
				{Code: "test.skip(\"foo\")", Options: map[string]any{"fn": "test"}},
				{Code: "test.concurrent(\"foo\")", Options: map[string]any{"fn": "test"}},
				{Code: "xtest(\"foo\")", Options: map[string]any{"fn": "test"}},
				{Code: "test.each([])(\"foo\")", Options: map[string]any{"fn": "test"}},
				{Code: "test.each``(\"foo\")", Options: map[string]any{"fn": "test"}},
				{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "test"}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code: "it(\"foo\")", Options: map[string]any{"fn": "test"},
					Output: []string{"test(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 3}},
				},
				{
					Code: "import { it } from '@jest/globals';\n\nit(\"foo\")", Options: map[string]any{"fn": "test"},
					// Differs from upstream, which calls the new name without importing it.
					Output: []string{"import { test, it } from '@jest/globals';\n\ntest(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 3, Column: 1, EndLine: 3, EndColumn: 3}},
				},
				{
					Code: "import { it as testThisThing } from '@jest/globals';\n\ntestThisThing(\"foo\")", Options: map[string]any{"fn": "test"},
					// Differs from upstream, which calls the new name without importing it.
					Output: []string{"import { test, it as testThisThing } from '@jest/globals';\n\ntest(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 3, Column: 1, EndLine: 3, EndColumn: 14}},
				},
				{
					Code: "xit(\"foo\")", Options: map[string]any{"fn": "test"},
					Output: []string{"xtest(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
				},
				{
					Code: "fit(\"foo\")", Options: map[string]any{"fn": "test"},
					Output: []string{"test.only(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
				},
				{
					Code: "it.skip(\"foo\")", Options: map[string]any{"fn": "test"},
					Output: []string{"test.skip(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}},
				},
				{
					Code: "it.concurrent(\"foo\")", Options: map[string]any{"fn": "test"},
					Output: []string{"test.concurrent(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 14}},
				},
				{
					Code: "it.only(\"foo\")", Options: map[string]any{"fn": "test"},
					Output: []string{"test.only(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}},
				},
				{
					Code: "it.each([])(\"foo\")", Options: map[string]any{"fn": "test"},
					Output: []string{"test.each([])(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 12}},
				},
				{
					Code: "it.each``(\"foo\")", Options: map[string]any{"fn": "test"},
					Output: []string{"test.each``(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 10}},
				},
				{
					Code: "describe.each``(\"foo\", () => { it.each``(\"bar\") })", Options: map[string]any{"fn": "test"},
					Output: []string{"describe.each``(\"foo\", () => { test.each``(\"bar\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 32, EndLine: 1, EndColumn: 41}},
				},
				{
					Code: "describe.each``(\"foo\", () => { test.each``(\"bar\") })", Options: map[string]any{"fn": "it"},
					Output: []string{"describe.each``(\"foo\", () => { it.each``(\"bar\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 32, EndLine: 1, EndColumn: 43}},
				},
				{
					Code: "describe.each()(\"%s\", () => {\n  test(\"is valid, but should not be\", () => {});\n\n  it(\"is not valid, but should be\", () => {});\n});", Options: map[string]any{"fn": "test", "withinDescribe": "it"},
					Output: []string{"describe.each()(\"%s\", () => {\n  it(\"is valid, but should not be\", () => {});\n\n  it(\"is not valid, but should be\", () => {});\n});"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 2, Column: 3, EndLine: 2, EndColumn: 7}},
				},
				{
					Code: "describe.only.each()(\"%s\", () => {\n  test(\"is valid, but should not be\", () => {});\n\n  it(\"is not valid, but should be\", () => {});\n});", Options: map[string]any{"fn": "test", "withinDescribe": "it"},
					Output: []string{"describe.only.each()(\"%s\", () => {\n  it(\"is valid, but should not be\", () => {});\n\n  it(\"is not valid, but should be\", () => {});\n});"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 2, Column: 3, EndLine: 2, EndColumn: 7}},
				},
				{
					Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "test"},
					Output: []string{"describe(\"suite\", () => { test(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 29}},
				},
			},
		)
	})
	t.Run("consistent-test-it with fn=it", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "it(\"foo\")", Options: map[string]any{"fn": "it"}},
				{Code: "fit(\"foo\")", Options: map[string]any{"fn": "it"}},
				{Code: "xit(\"foo\")", Options: map[string]any{"fn": "it"}},
				{Code: "it.only(\"foo\")", Options: map[string]any{"fn": "it"}},
				{Code: "it.skip(\"foo\")", Options: map[string]any{"fn": "it"}},
				{Code: "it.concurrent(\"foo\")", Options: map[string]any{"fn": "it"}},
				{Code: "it.each([])(\"foo\")", Options: map[string]any{"fn": "it"}},
				{Code: "it.each``(\"foo\")", Options: map[string]any{"fn": "it"}},
				{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "it"}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code: "test(\"foo\")", Options: map[string]any{"fn": "it"},
					Output: []string{"it(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 5}},
				},
				{
					Code: "xtest(\"foo\")", Options: map[string]any{"fn": "it"},
					Output: []string{"xit(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 6}},
				},
				{
					Code: "test.skip(\"foo\")", Options: map[string]any{"fn": "it"},
					Output: []string{"it.skip(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 10}},
				},
				{
					Code: "test.concurrent(\"foo\")", Options: map[string]any{"fn": "it"},
					Output: []string{"it.concurrent(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 16}},
				},
				{
					Code: "test.only(\"foo\")", Options: map[string]any{"fn": "it"},
					Output: []string{"it.only(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 10}},
				},
				{
					Code: "test.each([])(\"foo\")", Options: map[string]any{"fn": "it"},
					Output: []string{"it.each([])(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 14}},
				},
				{
					Code: "describe.each``(\"foo\", () => { test.each``(\"bar\") })", Options: map[string]any{"fn": "it"},
					Output: []string{"describe.each``(\"foo\", () => { it.each``(\"bar\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 32, EndLine: 1, EndColumn: 43}},
				},
				{
					Code: "test.each``(\"foo\")", Options: map[string]any{"fn": "it"},
					Output: []string{"it.each``(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 12}},
				},
				{
					Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "it"},
					Output: []string{"describe(\"suite\", () => { it(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 31}},
				},
			},
		)
	})
	t.Run("consistent-test-it with fn=test and withinDescribe=it", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "test(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
				{Code: "test.only(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
				{Code: "test.skip(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
				{Code: "test.concurrent(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
				{Code: "xtest(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
				{Code: "[1,2,3].forEach(() => { test(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"},
					Output: []string{"describe(\"suite\", () => { it(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 31}},
				},
				{
					Code: "describe(\"suite\", () => { test.only(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"},
					Output: []string{"describe(\"suite\", () => { it.only(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 36}},
				},
				{
					Code: "describe(\"suite\", () => { xtest(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"},
					Output: []string{"describe(\"suite\", () => { xit(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 32}},
				},
				{
					Code: "import { xtest as dontTestThis } from '@jest/globals';\n\ndescribe(\"suite\", () => { dontTestThis(\"foo\") });", Options: map[string]any{"fn": "test", "withinDescribe": "it"},
					// Differs from upstream, which calls the new name without importing it.
					Output: []string{"import { xit, xtest as dontTestThis } from '@jest/globals';\n\ndescribe(\"suite\", () => { xit(\"foo\") });"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 3, Column: 27, EndLine: 3, EndColumn: 39}},
				},
				{
					Code: "import { describe as context, xtest as dontTestThis } from '@jest/globals';\n\ncontext(\"suite\", () => { dontTestThis(\"foo\") });", Options: map[string]any{"fn": "test", "withinDescribe": "it"},
					// Differs from upstream, which calls the new name without importing it.
					Output: []string{"import { describe as context, xit, xtest as dontTestThis } from '@jest/globals';\n\ncontext(\"suite\", () => { xit(\"foo\") });"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 3, Column: 26, EndLine: 3, EndColumn: 38}},
				},
				{
					Code: "describe(\"suite\", () => { test.skip(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"},
					Output: []string{"describe(\"suite\", () => { it.skip(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 36}},
				},
				{
					Code: "describe(\"suite\", () => { test.concurrent(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"},
					Output: []string{"describe(\"suite\", () => { it.concurrent(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 42}},
				},
			},
		)
	})
	t.Run("consistent-test-it with fn=it and withinDescribe=test", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "it(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
				{Code: "it.only(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
				{Code: "it.skip(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
				{Code: "it.concurrent(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
				{Code: "xit(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
				{Code: "[1,2,3].forEach(() => { it(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"},
					Output: []string{"describe(\"suite\", () => { test(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 29}},
				},
				{
					Code: "describe(\"suite\", () => { it.only(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"},
					Output: []string{"describe(\"suite\", () => { test.only(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 34}},
				},
				{
					Code: "describe(\"suite\", () => { xit(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"},
					Output: []string{"describe(\"suite\", () => { xtest(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 30}},
				},
				{
					Code: "describe(\"suite\", () => { it.skip(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"},
					Output: []string{"describe(\"suite\", () => { test.skip(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 34}},
				},
				{
					Code: "describe(\"suite\", () => { it.concurrent(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"},
					Output: []string{"describe(\"suite\", () => { test.concurrent(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 40}},
				},
			},
		)
	})
	t.Run("consistent-test-it with fn=test and withinDescribe=test", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "test"}},
				{Code: "test(\"foo\");", Options: map[string]any{"fn": "test", "withinDescribe": "test"}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "test"},
					Output: []string{"describe(\"suite\", () => { test(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 29}},
				},
				{
					Code: "it(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "test"},
					Output: []string{"test(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 3}},
				},
			},
		)
	})
	t.Run("consistent-test-it with fn=it and withinDescribe=it", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "it"}},
				{Code: "it(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "it"}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "it"},
					Output: []string{"describe(\"suite\", () => { it(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 31}},
				},
				{
					Code: "test(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "it"},
					Output: []string{"it(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1, EndLine: 1, EndColumn: 5}},
				},
			},
		)
	})
	t.Run("consistent-test-it defaults without config object", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "test(\"foo\")"},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code:   "describe(\"suite\", () => { test(\"foo\") })",
					Output: []string{"describe(\"suite\", () => { it(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 31}},
				},
			},
		)
	})
	t.Run("consistent-test-it with withinDescribe=it", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "test(\"foo\")", Options: map[string]any{"withinDescribe": "it"}},
				{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"withinDescribe": "it"}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code: "it(\"foo\")", Options: map[string]any{"withinDescribe": "it"},
					Output: []string{"test(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 3}},
				},
				{
					Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"withinDescribe": "it"},
					Output: []string{"describe(\"suite\", () => { it(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 31}},
				},
			},
		)
	})
	t.Run("consistent-test-it with withinDescribe=test", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{
				{Code: "test(\"foo\")", Options: map[string]any{"withinDescribe": "test"}},
				{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"withinDescribe": "test"}},
			},
			[]rule_tester.InvalidTestCase{
				{
					Code: "it(\"foo\")", Options: map[string]any{"withinDescribe": "test"},
					Output: []string{"test(\"foo\")"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1, EndLine: 1, EndColumn: 3}},
				},
				{
					Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"withinDescribe": "test"},
					Output: []string{"describe(\"suite\", () => { test(\"foo\") })"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27, EndLine: 1, EndColumn: 29}},
				},
			},
		)
	})
	t.Run("documentation", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_test_it.ConsistentTestItRule,
			[]rule_tester.ValidTestCase{},
			[]rule_tester.InvalidTestCase{
				{
					Code: "test('foo'); // valid\ntest.only('foo'); // valid\n\nit('foo'); // invalid\nit.only('foo'); // invalid", Options: map[string]any{"fn": "test"},
					Output: []string{"test('foo'); // valid\ntest.only('foo'); // valid\n\ntest('foo'); // invalid\ntest.only('foo'); // invalid"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 4, Column: 1, EndLine: 4, EndColumn: 3}, {MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 5, Column: 1, EndLine: 5, EndColumn: 8}},
				},
				{
					Code: "it('foo'); // valid\nit.only('foo'); // valid\n\ntest('foo'); // invalid\ntest.only('foo'); // invalid", Options: map[string]any{"fn": "it"},
					Output: []string{"it('foo'); // valid\nit.only('foo'); // valid\n\nit('foo'); // invalid\nit.only('foo'); // invalid"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 4, Column: 1, EndLine: 4, EndColumn: 5}, {MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 5, Column: 1, EndLine: 5, EndColumn: 10}},
				},
				{
					Code: "it('foo'); // valid\ndescribe('foo', function () {\n  test('bar'); // valid\n});\n\ntest('foo'); // invalid\ndescribe('foo', function () {\n  it('bar'); // invalid\n});", Options: map[string]any{"fn": "it", "withinDescribe": "test"},
					Output: []string{"it('foo'); // valid\ndescribe('foo', function () {\n  test('bar'); // valid\n});\n\nit('foo'); // invalid\ndescribe('foo', function () {\n  test('bar'); // invalid\n});"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 6, Column: 1, EndLine: 6, EndColumn: 5}, {MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 8, Column: 3, EndLine: 8, EndColumn: 5}},
				},
				{
					Code:   "test('foo'); // valid\ndescribe('foo', function () {\n  it('bar'); // valid\n});\n\nit('foo'); // invalid\ndescribe('foo', function () {\n  test('bar'); // invalid\n});",
					Output: []string{"test('foo'); // valid\ndescribe('foo', function () {\n  it('bar'); // valid\n});\n\ntest('foo'); // invalid\ndescribe('foo', function () {\n  it('bar'); // invalid\n});"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 6, Column: 1, EndLine: 6, EndColumn: 3}, {MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 8, Column: 3, EndLine: 8, EndColumn: 7}},
				},
			},
		)
	})
}
