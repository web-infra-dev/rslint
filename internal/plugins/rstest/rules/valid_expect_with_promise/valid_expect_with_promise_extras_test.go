package valid_expect_with_promise

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestValidExpectWithPromiseExtras(t *testing.T) {
	thenables := map[string]any{"checkThenables": true}
	valid := []rule_tester.ValidTestCase{
		{Code: `expect(async () => { throw new Error('failure'); }).rejects.toThrow();`},
		{Code: `class MyPromise extends Promise<number> {} expect(MyPromise).toBeDefined();`},
		{Code: `expect(Promise.resolve(1)).to.be.a('promise').and.resolves.toBe(1);`},
		{Code: `expect.soft(Promise.resolve(1)).to.be.a('promise').and.resolves.toBe(1);`},
		{Code: `expect(async () => { throw new Error(); }).to.be.a('function').and.rejects.toThrow();`},
		{Code: `expect(1).toBe(1).result.resolves.toBe(1);`},
		{Code: `expect(1).toBe(1)[key].rejects.toBe(1);`},
		{Code: `declare function f<T = Promise<never>>(): T; expect(f).rejects.toThrow();`},
		{Code: `declare function f<T extends Promise<never>>(): T; expect(f).rejects.toThrow();`},
		{Code: `declare function f(...args: []): Promise<never>; declare function f(x: string): number; expect(f).rejects.toThrow();`},
		{Code: `expect(() => Promise.reject(new Error())).rejects.toThrow();`},
		{Code: `expect.soft(async () => { throw new Error(); }).rejects.toThrow();`},
		{Code: `expect.soft(Promise.resolve(1)).resolves.toBe(1);`},
		{Code: `expect(() => 1).toBeDefined();`},
		{Code: `expect(async () => 1).toBeDefined();`},
		{Code: `expect.poll(async () => 1).toBe(1);`},
		{Code: `expect.poll(() => 1).resolves.toBe(1);`},
		{Code: `expect.poll(Promise.resolve(1)).toBe(1);`},
		{Code: `expect.element(locator).toBeVisible();`},
		{Code: `expect.element(locator).rejects.toBeVisible();`},
		{Code: `expect.element(Promise.resolve(1)).toBeVisible();`},
		{Code: `expect.any(Promise); expect.assertions(1); expect.hasAssertions();`},
		{Code: `expect(Promise.resolve(1)).resolves.to.be.ok;`},
		{Code: `expect['poll'](() => 1).rejects.toBe(1);`},
		{Code: `expect['element'](locator).resolves.toBeVisible();`},
		{Code: `declare const modifier: string; expect(Promise.resolve(1))[modifier].toBe(1);`},
		{Code: `expect(); expect(Promise.resolve()); expect(Promise.resolve()).toBe;`},
		{Code: `class Promise<T> {} declare const x: Promise<number>; expect(x).toBe(1); export {};`},
		{Code: `declare const x: (() => Promise<never>) & { then(a: () => void): unknown }; expect(x).rejects.toThrow();`, Options: thenables},
		{Code: `function f(expect: any) { expect(Promise.resolve(1)).toBe(1); }`},
		{Code: `import { expect } from 'other'; expect(Promise.resolve(1)).toBe(1);`},
		{Code: `import type { expect as check } from '@rstest/core'; check(Promise.resolve(1)).toBe(1);`},
		{Code: `let expect = other; expect = third; expect(Promise.resolve(1)).toBe(1);`},
		{Code: `declare const p: Promise<number> | number; expect(p).toBe(1);`},
		{Code: `declare const p: Promise<number> | Promise<string>; expect(p).resolves.toBe(1);`},
		{Code: `declare const f: (() => Promise<number>) | Promise<string>; expect(f).rejects.toThrow();`},
		{Code: `function f<T extends () => Promise<never>>(value: T) { expect(value).rejects.toThrow(); }`},
		{Code: `declare function f(): Promise<number>; declare function f(x: string): number; expect(f).rejects.toThrow();`},
		{Code: `declare const p: PromiseLike<number>; expect(p).toBe(1);`},
		{Code: customPromiseDeclaration + ` expect(promised).toBe(1);`, Options: map[string]any{"checkThenables": false}},
		{Code: customPromiseDeclaration + ` expect(() => promised).rejects.toBe(1);`, Options: thenables},
		{Code: `declare const x: { then(a: () => void): unknown }; expect(x).toBe(1);`, Options: thenables},
		{Code: `declare const x: { then(a: string, b: () => void): unknown }; expect(x).toBe(1);`, Options: thenables},
		{Code: `declare const x: { then(a: () => void, b: string): unknown }; expect(x).toBe(1);`, Options: thenables},
		{Code: `declare const x: { then(a: () => void, b: any): unknown }; expect(x).toBe(1);`, Options: thenables},
		{Code: `declare const x: { then(...callbacks: (() => void)[]): unknown }; expect(x).toBe(1);`, Options: thenables},
		{Code: `declare const x: { then(a: () => void, ...b: (() => void)[]): unknown }; expect(x).toBe(1);`, Options: thenables},
		{Code: `declare const x: { then(a: () => void): unknown; then(a: string, b: () => void): unknown }; expect(x).toBe(1);`, Options: thenables},
		{Code: customPromiseDeclaration + ` declare const x: CustomPromise<string> | number; expect(x).toBe(1);`, Options: thenables},
		{Code: `function f<T>(x: T) { expect(x).toBe(1); }`, Options: thenables},
		{Code: `declare const x: unknown; expect(x).toBe(1);`, Options: thenables},
		{Code: `declare const x: never; expect(x).toBe(1);`, Options: thenables},
	}
	var invalid []rule_tester.InvalidTestCase
	missing := []string{
		`expect.soft(Promise.resolve(1)).toBe(1);`,
		`expect(Promise.resolve(1), 'message').not.toEqual(1);`,
		`expect(Promise.resolve(1)).to.be.ok;`,
		`expect(Promise.resolve(1)).to.be.ok.and.equal(1);`,
		`expect(Promise.resolve(1)).toEqual(1).then(() => {});`,
		`expect((Promise.resolve(1))).toEqual(1);`,
		`expect(Promise.resolve(1) satisfies Promise<number>).toBe(1);`,
		`expect(Promise.resolve(1)!).toBe(1);`,
		`expect<number>(Promise.resolve(1))['toBe'](1);`,
		`expect?.(Promise.resolve(1))?.toBe?.(1);`,
		`(expect(Promise.resolve(1)).toBe)(1);`,
		`expect(Promise.resolve(1)) /* comment */ .toBe(1);`,
		`import { expect as check } from '@rstest/core'; check(Promise.resolve(1)).toBe(1);`,
		`import * as api from '@rstest/core'; api.expect(Promise.resolve(1)).toBe(1);`,
		`import { expect } from 'rstack/test'; expect(Promise.resolve(1)).toBe(1);`,
		`const { expect: check } = require('@rstest/core'); check(Promise.resolve(1)).toBe(1);`,
		`const api = require('@rstest/core'); api.expect(Promise.resolve(1)).toBe(1);`,
		`import.meta.rstest.expect(Promise.resolve(1)).toBe(1);`,
		`test('context', ({ expect: check }) => { check(Promise.resolve(1)).toBe(1); });`,
		`test('context', ctx => { ctx.expect.soft(Promise.resolve(1)).toBe(1); });`,
		`import { expect } from '@rstest/playwright'; expect(Promise.resolve(1)).toBe(1);`,
		`declare const x: Promise<number> | Promise<string>; expect(x).toBe(1);`,
		`declare const x: Promise<number> & { tag: string }; expect(x).toBe(1);`,
		`function f<T extends Promise<number>>(x: T) { expect(x).toBe(1); }`,
		`class Parent extends Promise<number> {} class Child extends Parent {} declare const x: Child; expect(x).toBe(1);`,
	}
	for _, code := range missing {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "poorlyExpectedPromise", Message: "Subject is a promise so resolve or reject should be used"}}})
	}
	for _, code := range []string{
		`expect(() => 1).rejects.toBe(1);`,
		`class MyPromise extends Promise<number> {} expect(MyPromise).resolves.toBeDefined();`,
		`class MyPromise extends Promise<number> {} expect(MyPromise).rejects.toBeDefined();`,
		`expect(1).to.be.a('number').and.resolves.toBe(1);`,
		`expect.soft(1).to.be.a('number').and.rejects.toBe(1);`,
		`expect(1).to.be.a('number').and['resolves'].toBe(1);`,
		`function f(): number; function f(x: string): Promise<never>; function f(x?: string): number | Promise<never> { return x ? Promise.reject(new Error()) : 1; } expect(f).rejects.toThrow();`,
		`declare function f(x?: number): number; declare function f(): Promise<never>; expect(f).rejects.toThrow();`,
		`declare function f(...args: [string]): Promise<never>; declare function f(): number; expect(f).rejects.toThrow();`,
		`declare function f<T = number>(): T; declare function f(x: string): Promise<never>; expect(f).rejects.toThrow();`,
		`expect(() => { throw new Error(); }).rejects.toThrow();`,
		`expect(async () => 1).resolves.toBe(1);`,
		`expect.soft(1).resolves.toBe(1);`,
		`expect.soft(() => 1).rejects.toBe(1);`,
		`expect(Promise).resolves.toBe(1);`,
		`expect(1).not.rejects.toBe(1);`,
		`declare const x: Promise<number> | number; expect(x).resolves.toBe(1);`,
		`declare const x: (() => number) | (() => Promise<number>); expect(x).rejects.toBe(1);`,
		`declare const x: (() => number) & Promise<number>; expect(x).rejects.toBe(1);`,
		`declare const x: any; expect(x).resolves.toBe(1);`,
		`declare const x: unknown; expect(x).rejects.toBe(1);`,
		customPromiseDeclaration + ` expect(() => promised).rejects.toBe(1);`,
	} {
		modifier := "resolves"
		if strings.Contains(code, ".rejects") {
			modifier = "rejects"
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unneededRejectResolve", Message: "Subject is not a promise so " + modifier + " is not needed"}}})
	}
	for _, code := range []string{
		`declare const x: PromiseLike<number>; expect(x).toBe(1);`,
		customPromiseDeclaration + ` function f<T extends CustomPromise<number>>(x: T) { expect(x).toBe(1); }`,
		customPromiseDeclaration + ` declare const x: CustomPromise<number> | CustomPromise<string>; expect(x).toBe(1);`,
		customPromiseDeclaration + ` declare const x: CustomPromise<number> & { tag: string }; expect(x).toBe(1);`,
		`declare const x: { then: ((a: () => void, b: () => void) => unknown) | number }; expect(x).toBe(1);`,
		`declare const x: { then(a?: (() => void) | null, b?: (() => void) | null): unknown }; expect(x).toBe(1);`,
		`declare const x: { then(a: string): unknown; then(a: () => void, b: () => void): unknown }; expect(x).toBe(1);`,
		`declare const x: { then<F extends () => void, R extends () => void>(a: F, b: R): unknown }; expect(x).toBe(1);`,
		`declare const x: { then(a: (() => void) & { tag: string }, b: () => void): unknown }; expect(x).toBe(1);`,
		`declare const x: { then(a: (() => void) & ((x: number) => void), b: () => void): unknown }; expect(x).toBe(1);`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: thenables, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "poorlyExpectedPromise"}}})
	}
	invalid = append(invalid,
		rule_tester.InvalidTestCase{Code: `expect(1)["resolves"].toBe(1);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unneededRejectResolve", Line: 1, Column: 11, EndLine: 1, EndColumn: 21}}},
		rule_tester.InvalidTestCase{Code: "expect(1)[`rejects`].toBe(1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unneededRejectResolve", Line: 1, Column: 11, EndLine: 1, EndColumn: 20}}},
		rule_tester.InvalidTestCase{Code: `declare const x: { then(a: () => void): unknown }; expect(x).resolves.toBe(1);`, Options: thenables, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unneededRejectResolve"}}},
	)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ValidExpectWithPromiseRule, valid, invalid)
}

func TestValidExpectWithPromiseSchema(t *testing.T) {
	for _, options := range [][]any{[]any{}, []any{map[string]any{}}, []any{map[string]any{"checkThenables": false}}, []any{map[string]any{"checkThenables": true}}} {
		if err := ValidExpectWithPromiseRule.Schema.Validate(options); err != nil {
			t.Fatal(err)
		}
	}
	for _, options := range [][]any{[]any{map[string]any{"checkThenables": "true"}}, []any{map[string]any{"unknown": true}}, []any{map[string]any{}, map[string]any{}}} {
		if err := ValidExpectWithPromiseRule.Schema.Validate(options); err == nil {
			t.Fatalf("accepted invalid options %v", options)
		}
	}
}
