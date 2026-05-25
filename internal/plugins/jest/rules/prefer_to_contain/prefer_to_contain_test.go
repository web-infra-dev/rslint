package prefer_to_contain_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_to_contain"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToContainRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_to_contain.PreferToContainRule,
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
			{Code: `(expect('Model must be bound to an array if the multiple property is true') as any).toHaveBeenTipped()`},
			{Code: `expect(a.includes(b)).toEqual(0 as boolean);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(a.includes(b)).toEqual(true);`,
				Output: []string{`expect(a).toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 23},
				},
			},
			{
				Code:   `expect(a.includes(b,),).toEqual(true,);`,
				Output: []string{`expect(a,).toContain(b,);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 25},
				},
			},
			{
				Code:   `expect(a['includes'](b)).toEqual(true);`,
				Output: []string{`expect(a).toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 26},
				},
			},
			{
				Code:   `expect(a['includes'](b))['toEqual'](true);`,
				Output: []string{`expect(a).toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 26},
				},
			},
			{
				Code:   `expect(a['includes'](b)).toEqual(false);`,
				Output: []string{`expect(a).not.toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 26},
				},
			},
			{
				Code:   `expect(a['includes'](b)).not.toEqual(false);`,
				Output: []string{`expect(a).toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 30},
				},
			},
			{
				Code:   `expect(a['includes'](b))['not'].toEqual(false);`,
				Output: []string{`expect(a).toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 33},
				},
			},
			{
				Code:   `expect(a['includes'](b))['not']['toEqual'](false);`,
				Output: []string{`expect(a).toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 33},
				},
			},
			{
				Code:   `expect(a.includes(b)).toEqual(false);`,
				Output: []string{`expect(a).not.toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 23},
				},
			},
			{
				Code:   `expect(a.includes(b)).not.toEqual(false);`,
				Output: []string{`expect(a).toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 27},
				},
			},
			{
				Code:   `expect(a.includes(b)).not.toEqual(true);`,
				Output: []string{`expect(a).not.toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 27},
				},
			},
			{
				Code:   `expect(a.includes(b)).toBe(true);`,
				Output: []string{`expect(a).toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 23},
				},
			},
			{
				Code:   `expect(a.includes(b)).toBe(false);`,
				Output: []string{`expect(a).not.toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 23},
				},
			},
			{
				Code:   `expect(a.includes(b)).not.toBe(false);`,
				Output: []string{`expect(a).toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 27},
				},
			},
			{
				Code:   `expect(a.includes(b)).not.toBe(true);`,
				Output: []string{`expect(a).not.toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 27},
				},
			},
			{
				Code:   `expect(a.includes(b)).toStrictEqual(true);`,
				Output: []string{`expect(a).toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 23},
				},
			},
			{
				Code:   `expect(a.includes(b)).toStrictEqual(false);`,
				Output: []string{`expect(a).not.toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 23},
				},
			},
			{
				Code:   `expect(a.includes(b)).not.toStrictEqual(false);`,
				Output: []string{`expect(a).toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 27},
				},
			},
			{
				Code:   `expect(a.includes(b)).not.toStrictEqual(true);`,
				Output: []string{`expect(a).not.toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 27},
				},
			},
			{
				Code:   `expect(a.test(t).includes(b.test(p))).toEqual(true);`,
				Output: []string{`expect(a.test(t)).toContain(b.test(p));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 39},
				},
			},
			{
				Code:   `expect(a.test(t).includes(b.test(p))).toEqual(false);`,
				Output: []string{`expect(a.test(t)).not.toContain(b.test(p));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 39},
				},
			},
			{
				Code:   `expect(a.test(t).includes(b.test(p))).not.toEqual(true);`,
				Output: []string{`expect(a.test(t)).not.toContain(b.test(p));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 43},
				},
			},
			{
				Code:   `expect(a.test(t).includes(b.test(p))).not.toEqual(false);`,
				Output: []string{`expect(a.test(t)).toContain(b.test(p));`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 43},
				},
			},
			{
				Code:   `expect([{a:1}].includes({a:1})).toBe(true);`,
				Output: []string{`expect([{a:1}]).toContain({a:1});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 33},
				},
			},
			{
				Code:   `expect([{a:1}].includes({a:1})).toBe(false);`,
				Output: []string{`expect([{a:1}]).not.toContain({a:1});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 33},
				},
			},
			{
				Code:   `expect([{a:1}].includes({a:1})).not.toBe(true);`,
				Output: []string{`expect([{a:1}]).not.toContain({a:1});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 37},
				},
			},
			{
				Code:   `expect([{a:1}].includes({a:1})).not.toBe(false);`,
				Output: []string{`expect([{a:1}]).toContain({a:1});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 37},
				},
			},
			{
				Code:   `expect([{a:1}].includes({a:1})).toStrictEqual(true);`,
				Output: []string{`expect([{a:1}]).toContain({a:1});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 33},
				},
			},
			{
				Code:   `expect([{a:1}].includes({a:1})).toStrictEqual(false);`,
				Output: []string{`expect([{a:1}]).not.toContain({a:1});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 33},
				},
			},
			{
				Code:   `expect([{a:1}].includes({a:1})).not.toStrictEqual(true);`,
				Output: []string{`expect([{a:1}]).not.toContain({a:1});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 37},
				},
			},
			{
				Code:   `expect([{a:1}].includes({a:1})).not.toStrictEqual(false);`,
				Output: []string{`expect([{a:1}]).toContain({a:1});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 37},
				},
			},
			{
				Code: `import { expect as pleaseExpect } from '@jest/globals';

pleaseExpect([{a:1}].includes({a:1})).not.toStrictEqual(false);
`,
				Output: []string{`import { expect as pleaseExpect } from '@jest/globals';

pleaseExpect([{a:1}]).toContain({a:1});
`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 3, Column: 43},
				},
			},
			{
				Code:   `expect(a.includes(b)).toEqual(false as boolean);`,
				Output: []string{`expect(a).not.toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 23},
				},
			},
			{
				Code:   `expect(a.includes(b)).resolves.toBe(true);`,
				Output: []string{`expect(a).resolves.toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 32},
				},
			},
			{
				Code:   `expect(a.includes(b)).resolves.not.toBe(true);`,
				Output: []string{`expect(a).resolves.not.toContain(b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 36},
				},
			},
			// Regression: when an unknown matcher is chained after an equality
			// matcher, only the equality matcher (the inner CallExpression) must
			// be reported. The outer call must not produce a duplicate report or
			// a fix that drops the trailing chained call.
			{
				Code:   `expect(a.includes(b)).toBe(true).somethingElse(true);`,
				Output: []string{`expect(a).toContain(b).somethingElse(true);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToContain", Line: 1, Column: 23},
				},
			},
		},
	)
}
