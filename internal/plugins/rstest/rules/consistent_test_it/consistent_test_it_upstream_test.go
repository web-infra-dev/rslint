// TestConsistentTestItUpstream migrates all cases from eslint-plugin-jest v29.16.1
// and @vitest/eslint-plugin v1.6.27. Unsupported prefix APIs are explicitly
// skipped; import fixes retain bindings. Rstest augmentation is in the extras file.
package consistent_test_it

import (
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestConsistentTestItUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTestItRule,
		[]rule_tester.ValidTestCase{
			// jest group 0.
			{Code: "test(\"foo\")", Options: map[string]any{"fn": "test"}},
			// jest group 0.
			{Code: "test.only(\"foo\")", Options: map[string]any{"fn": "test"}},
			// jest group 0.
			{Code: "test.skip(\"foo\")", Options: map[string]any{"fn": "test"}},
			// jest group 0.
			{Code: "test.concurrent(\"foo\")", Options: map[string]any{"fn": "test"}},
			// jest group 0.
			{Code: "xtest(\"foo\")", Options: map[string]any{"fn": "test"}},
			// jest group 0.
			{Code: "test.each([])(\"foo\")", Options: map[string]any{"fn": "test"}},
			// jest group 0.
			{Code: "test.each``(\"foo\")", Options: map[string]any{"fn": "test"}},
			// jest group 0.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "test"}},
			// jest group 1.
			{Code: "it(\"foo\")", Options: map[string]any{"fn": "it"}},
			// jest group 1.
			{Code: "fit(\"foo\")", Options: map[string]any{"fn": "it"}},
			// jest group 1.
			{Code: "xit(\"foo\")", Options: map[string]any{"fn": "it"}},
			// jest group 1.
			{Code: "it.only(\"foo\")", Options: map[string]any{"fn": "it"}},
			// jest group 1.
			{Code: "it.skip(\"foo\")", Options: map[string]any{"fn": "it"}},
			// jest group 1.
			{Code: "it.concurrent(\"foo\")", Options: map[string]any{"fn": "it"}},
			// jest group 1.
			{Code: "it.each([])(\"foo\")", Options: map[string]any{"fn": "it"}},
			// jest group 1.
			{Code: "it.each``(\"foo\")", Options: map[string]any{"fn": "it"}},
			// jest group 1.
			{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "it"}},
			// jest group 2.
			{Code: "test(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
			// jest group 2.
			{Code: "test.only(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
			// jest group 2.
			{Code: "test.skip(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
			// jest group 2.
			{Code: "test.concurrent(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
			// jest group 2.
			{Code: "xtest(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
			// jest group 2.
			{Code: "[1,2,3].forEach(() => { test(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"}},
			// jest group 3.
			{Code: "it(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
			// jest group 3.
			{Code: "it.only(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
			// jest group 3.
			{Code: "it.skip(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
			// jest group 3.
			{Code: "it.concurrent(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
			// jest group 3.
			{Code: "xit(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
			// jest group 3.
			{Code: "[1,2,3].forEach(() => { it(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"}},
			// jest group 4.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "test"}},
			// jest group 4.
			{Code: "test(\"foo\");", Options: map[string]any{"fn": "test", "withinDescribe": "test"}},
			// jest group 5.
			{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "it"}},
			// jest group 5.
			{Code: "it(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "it"}},
			// jest group 6.
			{Code: "test(\"foo\")"},
			// jest group 7.
			{Code: "test(\"foo\")", Options: map[string]any{"withinDescribe": "it"}},
			// jest group 7.
			{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"withinDescribe": "it"}},
			// jest group 8.
			{Code: "test(\"foo\")", Options: map[string]any{"withinDescribe": "test"}},
			// jest group 8.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"withinDescribe": "test"}},
			// vitest group 9.
			{Code: "it(\"shows error\", () => {\n  expect(true).toBe(false);\n        });", Options: map[string]any{"fn": "it"}},
			// vitest group 9.
			{Code: "it(\"foo\", function () {\n         expect(true).toBe(false);\n     })", Options: map[string]any{"fn": "it"}},
			// vitest group 9.
			{Code: " it('foo', () => {\n      expect(true).toBe(false);\n  });\n  function myTest() { if ('bar') {} }", Options: map[string]any{"fn": "it"}},
			// vitest group 9.
			{Code: "bench(\"foo\", function () {\n        fibonacci(10);\n     })", Options: map[string]any{"fn": "it"}},
			// vitest group 10.
			{Code: "test(\"shows error\", () => {\n      expect(true).toBe(false);\n     });", Options: map[string]any{"fn": "test"}},
			// vitest group 10.
			{Code: "test.skip(\"foo\")", Options: map[string]any{"fn": "test"}},
			// vitest group 10.
			{Code: "test.concurrent(\"foo\")", Options: map[string]any{"fn": "test"}},
			// vitest group 10.
			{Code: "xtest(\"foo\")", Options: map[string]any{"fn": "test"}},
			// vitest group 10.
			{Code: "test.each([])(\"foo\")", Options: map[string]any{"fn": "test"}},
			// vitest group 10.
			{Code: "test.each``(\"foo\")", Options: map[string]any{"fn": "test"}},
			// vitest group 10.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "test"}},
			// vitest group 11.
			{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "it"}},
			// vitest group 11.
			{Code: "it(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "it"}},
			// vitest group 12.
			{Code: "test(\"shows error\", () => {});"},
			// vitest group 13.
			{Code: "test(\"foo\")", Options: map[string]any{"withinDescribe": "it"}},
			// vitest group 13.
			{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"withinDescribe": "it"}},
			// vitest group 14.
			{Code: "test(\"foo\")", Options: map[string]any{"withinDescribe": "test"}},
			// vitest group 14.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"withinDescribe": "test"}},
			// vitest group 14. Import-only violation is not a registration naming violation.
			{Code: "import { it as baseIt, test } from \"@rstest/core\"\nbaseIt(\"foo\")", Options: map[string]any{"fn": "it"}},
		}, []rule_tester.InvalidTestCase{
			// jest group 0, case 0.
			{Code: "it(\"foo\")", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test(\"foo\")"}},
			// jest group 0, case 1. Rstest adaptation: preserve existing bindings; imports are not separate diagnostics.
			{Code: "import { it } from '@rstest/core';\n\nit(\"foo\")", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 3, Column: 1}}, Output: []string{"import { test, it } from '@rstest/core';\n\ntest(\"foo\")"}},
			// jest group 0, case 2. Rstest adaptation: aliased calls report without replacing the binding.
			{Code: "import { it as testThisThing } from '@rstest/core';\n\ntestThisThing(\"foo\")", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 3, Column: 1}}, Output: []string{}},
			// jest group 0, case 3. SKIP: Rstest has no fit, xit, or xtest exports.
			{Code: "xit(\"foo\")", Options: map[string]any{"fn": "test"}, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod"}}},
			// jest group 0, case 4. SKIP: Rstest has no fit, xit, or xtest exports.
			{Code: "fit(\"foo\")", Options: map[string]any{"fn": "test"}, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod"}}},
			// jest group 0, case 5.
			{Code: "it.skip(\"foo\")", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test.skip(\"foo\")"}},
			// jest group 0, case 6.
			{Code: "it.concurrent(\"foo\")", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test.concurrent(\"foo\")"}},
			// jest group 0, case 7.
			{Code: "it.only(\"foo\")", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test.only(\"foo\")"}},
			// jest group 0, case 8.
			{Code: "it.each([])(\"foo\")", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test.each([])(\"foo\")"}},
			// jest group 0, case 9.
			{Code: "it.each``(\"foo\")", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test.each``(\"foo\")"}},
			// jest group 0, case 10.
			{Code: "describe.each``(\"foo\", () => { it.each``(\"bar\") })", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 32}}, Output: []string{"describe.each``(\"foo\", () => { test.each``(\"bar\") })"}},
			// jest group 0, case 11.
			{Code: "describe.each``(\"foo\", () => { test.each``(\"bar\") })", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 32}}, Output: []string{"describe.each``(\"foo\", () => { it.each``(\"bar\") })"}},
			// jest group 0, case 12.
			{Code: "describe.each()(\"%s\", () => {\n  test(\"is valid, but should not be\", () => {});\n\n  it(\"is not valid, but should be\", () => {});\n});", Options: map[string]any{"fn": "test", "withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 2, Column: 3}}, Output: []string{"describe.each()(\"%s\", () => {\n  it(\"is valid, but should not be\", () => {});\n\n  it(\"is not valid, but should be\", () => {});\n});"}},
			// jest group 0, case 13.
			{Code: "describe.only.each()(\"%s\", () => {\n  test(\"is valid, but should not be\", () => {});\n\n  it(\"is not valid, but should be\", () => {});\n});", Options: map[string]any{"fn": "test", "withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 2, Column: 3}}, Output: []string{"describe.only.each()(\"%s\", () => {\n  it(\"is valid, but should not be\", () => {});\n\n  it(\"is not valid, but should be\", () => {});\n});"}},
			// jest group 0, case 14.
			{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { test(\"foo\") })"}},
			// jest group 1, case 0.
			{Code: "test(\"foo\")", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1}}, Output: []string{"it(\"foo\")"}},
			// jest group 1, case 1. SKIP: Rstest has no fit, xit, or xtest exports.
			{Code: "xtest(\"foo\")", Options: map[string]any{"fn": "it"}, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod"}}},
			// jest group 1, case 2.
			{Code: "test.skip(\"foo\")", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1}}, Output: []string{"it.skip(\"foo\")"}},
			// jest group 1, case 3.
			{Code: "test.concurrent(\"foo\")", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1}}, Output: []string{"it.concurrent(\"foo\")"}},
			// jest group 1, case 4.
			{Code: "test.only(\"foo\")", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1}}, Output: []string{"it.only(\"foo\")"}},
			// jest group 1, case 5.
			{Code: "test.each([])(\"foo\")", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1}}, Output: []string{"it.each([])(\"foo\")"}},
			// jest group 1, case 6.
			{Code: "describe.each``(\"foo\", () => { test.each``(\"bar\") })", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 32}}, Output: []string{"describe.each``(\"foo\", () => { it.each``(\"bar\") })"}},
			// jest group 1, case 7.
			{Code: "test.each``(\"foo\")", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1}}, Output: []string{"it.each``(\"foo\")"}},
			// jest group 1, case 8.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { it(\"foo\") })"}},
			// jest group 2, case 0.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { it(\"foo\") })"}},
			// jest group 2, case 1.
			{Code: "describe(\"suite\", () => { test.only(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { it.only(\"foo\") })"}},
			// jest group 2, case 2. SKIP: Rstest has no fit, xit, or xtest exports.
			{Code: "describe(\"suite\", () => { xtest(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"}, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod"}}},
			// jest group 2, case 3. SKIP: Rstest has no fit, xit, or xtest exports.
			{Code: "import { xtest as dontTestThis } from '@rstest/core';\n\ndescribe(\"suite\", () => { dontTestThis(\"foo\") });", Options: map[string]any{"fn": "test", "withinDescribe": "it"}, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod"}}},
			// jest group 2, case 4. SKIP: Rstest has no fit, xit, or xtest exports.
			{Code: "import { describe as context, xtest as dontTestThis } from '@rstest/core';\n\ncontext(\"suite\", () => { dontTestThis(\"foo\") });", Options: map[string]any{"fn": "test", "withinDescribe": "it"}, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod"}}},
			// jest group 2, case 5.
			{Code: "describe(\"suite\", () => { test.skip(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { it.skip(\"foo\") })"}},
			// jest group 2, case 6.
			{Code: "describe(\"suite\", () => { test.concurrent(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { it.concurrent(\"foo\") })"}},
			// jest group 3, case 0.
			{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { test(\"foo\") })"}},
			// jest group 3, case 1.
			{Code: "describe(\"suite\", () => { it.only(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { test.only(\"foo\") })"}},
			// jest group 3, case 2. SKIP: Rstest has no fit, xit, or xtest exports.
			{Code: "describe(\"suite\", () => { xit(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"}, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod"}}},
			// jest group 3, case 3.
			{Code: "describe(\"suite\", () => { it.skip(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { test.skip(\"foo\") })"}},
			// jest group 3, case 4.
			{Code: "describe(\"suite\", () => { it.concurrent(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { test.concurrent(\"foo\") })"}},
			// jest group 4, case 0.
			{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "test", "withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { test(\"foo\") })"}},
			// jest group 4, case 1.
			{Code: "it(\"foo\")", Options: map[string]any{"fn": "test", "withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test(\"foo\")"}},
			// jest group 5, case 0.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { it(\"foo\") })"}},
			// jest group 5, case 1.
			{Code: "test(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1}}, Output: []string{"it(\"foo\")"}},
			// jest group 6, case 0.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { it(\"foo\") })"}},
			// jest group 7, case 0.
			{Code: "it(\"foo\")", Options: map[string]any{"withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test(\"foo\")"}},
			// jest group 7, case 1.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { it(\"foo\") })"}},
			// jest group 8, case 0.
			{Code: "it(\"foo\")", Options: map[string]any{"withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test(\"foo\")"}},
			// jest group 8, case 1.
			{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { test(\"foo\") })"}},
			// vitest group 9, case 0.
			{Code: "test(\"shows error\", () => {});", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1}}, Output: []string{"it(\"shows error\", () => {});"}},
			// vitest group 9, case 1.
			{Code: "test.skip(\"shows error\");", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1}}, Output: []string{"it.skip(\"shows error\");"}},
			// vitest group 9, case 2.
			{Code: "test.only('shows error');", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1}}, Output: []string{"it.only('shows error');"}},
			// vitest group 9, case 3.
			{Code: "describe('foo', () => { it('bar', () => {}); });", Options: map[string]any{"fn": "it", "withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 25}}, Output: []string{"describe('foo', () => { test('bar', () => {}); });"}},
			// vitest group 9, case 4. Rstest adaptation: preserve existing bindings; imports are not separate diagnostics.
			{Code: "import { test } from \"@rstest/core\"\ntest(\"shows error\", () => {});", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 2, Column: 1}}, Output: []string{"import { it, test } from \"@rstest/core\"\nit(\"shows error\", () => {});"}},
			// vitest group 9, case 5. Rstest adaptation: preserve existing bindings; imports are not separate diagnostics.
			{Code: "import { expect, test, it } from \"@rstest/core\"\ntest(\"shows error\", () => {});", Options: map[string]any{"fn": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 2, Column: 1}}, Output: []string{"import { expect, test, it } from \"@rstest/core\"\nit(\"shows error\", () => {});"}},
			// vitest group 10, case 0.
			{Code: "it(\"shows error\", () => {});", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test(\"shows error\", () => {});"}},
			// vitest group 10, case 1.
			{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"fn": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { test(\"foo\") })"}},
			// vitest group 11, case 0.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"fn": "it", "withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { it(\"foo\") })"}},
			// vitest group 11, case 1.
			{Code: "test(\"foo\")", Options: map[string]any{"fn": "it", "withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'it' instead of 'test'", Line: 1, Column: 1}}, Output: []string{"it(\"foo\")"}},
			// vitest group 12, case 0.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { it(\"foo\") })"}},
			// vitest group 13, case 0.
			{Code: "it(\"foo\")", Options: map[string]any{"withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test(\"foo\")"}},
			// vitest group 13, case 1.
			{Code: "describe(\"suite\", () => { test(\"foo\") })", Options: map[string]any{"withinDescribe": "it"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'it' instead of 'test' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { it(\"foo\") })"}},
			// vitest group 14, case 0.
			{Code: "it(\"foo\")", Options: map[string]any{"withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 1, Column: 1}}, Output: []string{"test(\"foo\")"}},
			// vitest group 14, case 1. Rstest adaptation: preserve existing bindings; imports are not separate diagnostics.
			{Code: "import { it } from \"@rstest/core\"\nit(\"foo\")", Options: map[string]any{"withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 2, Column: 1}}, Output: []string{"import { test, it } from \"@rstest/core\"\ntest(\"foo\")"}},
			// vitest group 14, case 3. Rstest adaptation: preserve existing bindings; imports are not separate diagnostics.
			{Code: "import { expect, it, test } from \"@rstest/core\"\nit(\"foo\")", Options: map[string]any{"withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethod", Message: "Prefer using 'test' instead of 'it'", Line: 2, Column: 1}}, Output: []string{"import { expect, it, test } from \"@rstest/core\"\ntest(\"foo\")"}},
			// vitest group 14, case 4.
			{Code: "describe(\"suite\", () => { it(\"foo\") })", Options: map[string]any{"withinDescribe": "test"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistentMethodWithinDescribe", Message: "Prefer using 'test' instead of 'it' within describe", Line: 1, Column: 27}}, Output: []string{"describe(\"suite\", () => { test(\"foo\") })"}},
		})
}
