// TestPreferToHaveLengthExtras covers Rstest expect sources, framework
// boundaries, tsgo edge shapes, diagnostic ranges and fix safety. The complete
// upstream corpus is in prefer_to_have_length_upstream_test.go.
package prefer_to_have_length

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func lengthError(line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "useToHaveLength", Message: "Use `toHaveLength()` instead",
		Line: line, Column: column,
	}
}

func TestPreferToHaveLengthExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Rstest factory/runtime boundaries.
		{Code: `await expect.poll(() => values.length).toBe(2);`},
		{Code: `await expect(Promise.resolve(values.length)).resolves.toBe(2);`},
		{Code: `await expect(Promise.reject(values.length)).rejects.toEqual(2);`},
		{Code: `expect.element(locator).toEqual(2);`},
		// Chai matchers are not Jest-style equality matchers for this rule.
		{Code: `expect(values.length).to.equal(2);`},
		{Code: `expect(values).to.have.lengthOf(2);`},
		{Code: `expect(values).to.have.length(2);`},
		// Foreign, shadowed, type-only and reassigned expect roots.
		{Code: `import { expect } from 'vitest'; expect(values.length).toBe(2);`},
		{Code: `import { expect } from '@jest/globals'; expect(values.length).toBe(2);`},
		{Code: `const expect = createExpect(); expect(values.length).toBe(2);`},
		{Code: `import type { expect as check } from '@rstest/core'; check(values.length).toBe(2);`},
		{Code: `import { expect } from '@rstest/core'; function run(expect: any) { expect(values.length).toBe(2); }`},
		{Code: `let { expect: check } = require('@rstest/core'); check = replacement; check(values.length).toBe(2);`},
		// Dynamic and optional accessors cannot prove the target property.
		{Code: `expect(values[length]).toBe(2);`},
		{Code: `expect(values["len" + suffix]).toBe(2);`},
		{Code: `expect(values?.length).toBe(2);`},
		{Code: `expect(values.length)[matcher](2);`},
		{Code: `expect(values.length)[0](2);`},
		{Code: `expect(values.length).toBe;`},
		// TypeScript wrappers outside the expect head are parser boundaries.
		{Code: `(expect(values.length) as any).toBe(2);`},
		{Code: `expect(values.length)!.toBe(2);`},
	}

	invalid := []rule_tester.InvalidTestCase{
		// Visibly inert receivers can be rewritten without moving a getter.
		{
			Code:   `expect([1, 2, 3].length).toBe(3);`,
			Output: []string{`expect([1, 2, 3]).toHaveLength(3);`},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useToHaveLength", Message: "Use `toHaveLength()` instead",
				Line: 1, Column: 26, EndLine: 1, EndColumn: 30,
			}},
		},
		{
			Code:   `expect("hello"["length"]).not["toEqual"](+5);`,
			Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 31)},
		},
		{
			Code:   "expect((`hello`)[`length`]).toStrictEqual(5 as number);",
			Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 29)},
		},
		{
			Code:   `expect((function named(a, b) {}).length).toBe(2);`,
			Output: []string{`expect((function named(a, b) {})).toHaveLength(2);`},
			Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 42)},
		},
		{
			Code:   `expect(((a, b) => a + b).length)["toBe"](2);`,
			Output: []string{`expect(((a, b) => a + b))["toHaveLength"](2);`},
			Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 34)},
		},
		// Parentheses, optional matcher links and authored comments survive.
		{
			Code:   `expect((([1, 2]).length)). /* matcher */ toEqual?.((2));`,
			Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 42)},
		},
		{
			Code:   `expect([1, 2]. /* keep */ length).toBe(2);`,
			Output: []string{`expect([1, 2] /* keep */ ).toHaveLength(2);`},
			Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 35)},
		},
		// Type arguments belong to the old call signatures and are removed.
		{
			Code:   `expect<string[]>(["a"].length).toBe<number>(1);`,
			Output: []string{`expect(["a"]).toHaveLength(1);`},
			Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 32)},
		},
		{
			Code:   `expect<string[]>(["a"].length).toBe</* keep */ number>(1);`,
			Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 32)},
		},
		// Arbitrary receivers still report but are not fixed: `.length` may be a getter.
		{Code: `expect(values.length).toBe(2);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 23)}},
		{Code: `expect(loadValues().length).toEqual(2);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 29)}},
		{Code: `expect(holder.values.length).toStrictEqual(2);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 30)}},
		// Non-literal expected values, extra arguments and custom messages may run code.
		{Code: `expect([1, 2].length).toBe(expected);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 23)}},
		{Code: `expect([1, 2].length).toBe();`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 23)}},
		{Code: `expect([1, 2].length, message()).toBe(2);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 34)}},
		{Code: `expect([1, 2].length).toBe(2, notify());`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 23)}},
		{Code: `expect([1, 2].length).toBe(...lengths);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 23)}},
		{Code: `consume(expect([1, 2].length).toBe(2));`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 31)}},
		{Code: `returnValue = expect([1, 2].length).toBe(2);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 37)}},
		// Rstest source/provenance matrix. Identifier receivers report without fixes.
		{Code: `import { expect as check } from '@rstest/core'; check(values.length).toBe(2);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 70)}},
		{Code: `import * as core from '@rstest/core'; core.expect(values.length).toEqual(2);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 66)}},
		{Code: `const { expect } = require('rstack/test'); expect(values.length).toStrictEqual(2);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 66)}},
		{Code: `import.meta.rstest.expect(values.length).toBe(2);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 42)}},
		{Code: `import { expect } from '@rstest/playwright'; expect(values.length).toBe(2);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 68)}},
		{Code: `test('length', ({ expect: check }) => check(values.length).toBe(2));`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 60)}},
		// Rstest soft assertions use the same value matcher.
		{
			Code:   `expect.soft([1, 2].length).not.toEqual(0);`,
			Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 32)},
		},
		// toEqual and toStrictEqual can use custom equality testers that
		// toHaveLength does not consult, so they report without autofixes.
		{
			Code: `expect.addEqualityTesters([(actual, expected) => actual + 1 === expected]);
expect([].length).toEqual(1);`,
			Errors: []rule_tester.InvalidTestCaseError{lengthError(2, 19)},
		},
		{Code: `expect([].length).toStrictEqual(0);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 19)}},
		// toBe distinguishes negative zero from zero, while Rstest's
		// toHaveLength delegates to Chai's loose comparison.
		{Code: `expect([].length).toBe(-0);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 19)}},
		{Code: `expect([].length).toBe(-(0 as number));`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 19)}},
		{Code: `expect([].length).toBe(-(+0));`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 19)}},
		{
			Code:   `expect([].length).toBe(-(-0));`,
			Output: []string{`expect([]).toHaveLength(-(-0));`},
			Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 19)},
		},
		// Continued assertion chains still diagnose the first equality matcher,
		// but changing their shared subject would affect later operations.
		{
			Code: `expect(values.length).toBe(2).and.toEqual(2);`,
			Errors: []rule_tester.InvalidTestCaseError{
				lengthError(1, 23),
				lengthError(1, 35),
			},
		},
		{Code: `expect(values.length).toBe(2).to.be.ok;`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 23)}},
		{Code: `expect(values.length).toBe(2).message;`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 23)}},
		{Code: `expect(values.length).to.be.a('number').and.toBe(2);`, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 45)}},
		// property() changes Chai's assertion object to the selected property,
		// so a later equality matcher no longer checks the original length.
		{Code: `expect(values.length).to.have.property('x').and.toBe(2);`},
		{Code: `expect(values.length).to.have.ownPropertyDescriptor('x').and.toEqual(2);`},
		{Code: `expect([].length).not.toContain(1).and.toEqual(0);`},
		{Code: `expect({ length: () => { throw 2; } }.length).toThrow().and.toBe(2);`},
		{Code: `expect({ length: () => { throw 2; } }.length).toThrowError().and.toStrictEqual(2);`},
		{
			Code: `expect(values.length).toBe(2).and.toStrictEqual(2);`,
			Errors: []rule_tester.InvalidTestCaseError{
				lengthError(1, 23),
				lengthError(1, 35),
			},
		},
		// Diagnostic range uses UTF-16 columns and repeated assertions stay independent.
		{Code: `expect("😀".length).toBe(2);`, Output: []string{`expect("😀").toHaveLength(2);`}, Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 21)}},
		{
			Code: `expect([1].length).toBe(1);
expect([1, 2].length).toEqual(2);`,
			Output: []string{`expect([1]).toHaveLength(1);
expect([1, 2].length).toEqual(2);`},
			Errors: []rule_tester.InvalidTestCaseError{lengthError(1, 20), lengthError(2, 23)},
		},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferToHaveLengthRule,
		valid,
		invalid,
	)
}
