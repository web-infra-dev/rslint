// TestPreferToContainUpstream migrates every valid and invalid case from
// @vitest/eslint-plugin@v1.6.27 tests/prefer-to-contain.test.ts. Rstest-only
// semantic exclusions and tsgo edge shapes live in the sibling extras suite.
package prefer_to_contain

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToContainUpstream(t *testing.T) {
	invalid := func(code, output string, column int) rule_tester.InvalidTestCase {
		return rule_tester.InvalidTestCase{
			Code: code, Output: []string{output},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: column}},
		}
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferToContainRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect.hasAssertions`},
			{Code: `expect.hasAssertions()`},
			{Code: `expect.assertions(1)`},
			{Code: `expect().toBe(false);`},
			{Code: `expect(a).toContain(b);`},
			{Code: `expect(a.name).toBe('b');`},
			{Code: `expect(a).toBe(true);`},
			{Code: `expect(a).toEqual(b)`},
			{Code: `expect(a.test(c)).toEqual(b)`},
			{Code: `expect(a.includes(b)).toEqual()`},
			{Code: `expect(a.includes(b)).toEqual("test")`},
			{Code: `expect(a.includes(b)).toBe("test")`},
			{Code: `expect(a.includes()).toEqual()`},
			{Code: `expect(a.includes()).toEqual(true)`},
			{Code: `expect(a.includes(b,c)).toBe(true)`},
			{Code: `expect([{a:1}]).toContain({a:1})`},
			{Code: `expect([1].includes(1)).toEqual`},
			{Code: `expect([1].includes).toEqual`},
			{Code: `expect([1].includes).not`},
			{Code: `expect(a.test(b)).resolves.toEqual(true)`},
			{Code: `expect(a.test(b)).resolves.not.toEqual(true)`},
			{Code: `expect(a).not.toContain(b)`},
			{Code: `expect(a.includes(...[])).toBe(true)`},
			{Code: `expect(a.includes(b)).toBe(...true)`},
			{Code: `expect(a);`},
			{Code: `expect(a).to.be.a("string");`},
		},
		[]rule_tester.InvalidTestCase{
			invalid(`expect(a.includes(b)).toEqual(true);`, `expect(a).toContain(b);`, 23),
			invalid(`expect(a.includes(b,),).toEqual(true,);`, `expect(a,).toContain(b,);`, 25),
			invalid(`expect(a['includes'](b)).toEqual(true);`, `expect(a).toContain(b);`, 26),
			invalid(`expect(a['includes'](b))['toEqual'](true);`, `expect(a)['toContain'](b);`, 26),
			invalid(`expect(a['includes'](b)).toEqual(false);`, `expect(a).not.toContain(b);`, 26),
			invalid(`expect(a['includes'](b)).not.toEqual(false);`, `expect(a).toContain(b);`, 30),
			invalid(`expect(a['includes'](b))['not'].toEqual(false);`, `expect(a).toContain(b);`, 33),
			invalid(`expect(a['includes'](b))['not']['toEqual'](false);`, `expect(a)['toContain'](b);`, 33),
			invalid(`expect(a.includes(b)).toEqual(false);`, `expect(a).not.toContain(b);`, 23),
			invalid(`expect(a.includes(b)).not.toEqual(false);`, `expect(a).toContain(b);`, 27),
			invalid(`expect(a.includes(b)).not.toEqual(true);`, `expect(a).not.toContain(b);`, 27),
			invalid(`expect(a.includes(b)).toBe(true);`, `expect(a).toContain(b);`, 23),
			invalid(`expect(a.includes(b)).toBe(false);`, `expect(a).not.toContain(b);`, 23),
			invalid(`expect(a.includes(b)).not.toBe(false);`, `expect(a).toContain(b);`, 27),
			invalid(`expect(a.includes(b)).not.toBe(true);`, `expect(a).not.toContain(b);`, 27),
			invalid(`expect(a.includes(b)).toStrictEqual(true);`, `expect(a).toContain(b);`, 23),
			invalid(`expect(a.includes(b)).toStrictEqual(false);`, `expect(a).not.toContain(b);`, 23),
			invalid(`expect(a.includes(b)).not.toStrictEqual(false);`, `expect(a).toContain(b);`, 27),
			invalid(`expect(a.includes(b)).not.toStrictEqual(true);`, `expect(a).not.toContain(b);`, 27),
			// ADAPT: Rstest reports these upstream cases but withholds the fix;
			// moving b.test(p) after expect(...) can reorder observable effects.
			{Code: `expect(a.test(t).includes(b.test(p))).toEqual(true);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 39}}},
			{Code: `expect(a.test(t).includes(b.test(p))).toEqual(false);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 39}}},
			{Code: `expect(a.test(t).includes(b.test(p))).not.toEqual(true);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 43}}},
			{Code: `expect(a.test(t).includes(b.test(p))).not.toEqual(false);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToContain", Line: 1, Column: 43}}},
			invalid(`expect([{a:1}].includes({a:1})).toBe(true);`, `expect([{a:1}]).toContain({a:1});`, 33),
			invalid(`expect([{a:1}].includes({a:1})).toBe(false);`, `expect([{a:1}]).not.toContain({a:1});`, 33),
			invalid(`expect([{a:1}].includes({a:1})).not.toBe(true);`, `expect([{a:1}]).not.toContain({a:1});`, 37),
			invalid(`expect([{a:1}].includes({a:1})).not.toBe(false);`, `expect([{a:1}]).toContain({a:1});`, 37),
			invalid(`expect([{a:1}].includes({a:1})).toStrictEqual(true);`, `expect([{a:1}]).toContain({a:1});`, 33),
			invalid(`expect([{a:1}].includes({a:1})).toStrictEqual(false);`, `expect([{a:1}]).not.toContain({a:1});`, 33),
			invalid(`expect([{a:1}].includes({a:1})).not.toStrictEqual(true);`, `expect([{a:1}]).not.toContain({a:1});`, 37),
			invalid(`expect([{a:1}].includes({a:1})).not.toStrictEqual(false);`, `expect([{a:1}]).toContain({a:1});`, 37),
		},
	)
}
