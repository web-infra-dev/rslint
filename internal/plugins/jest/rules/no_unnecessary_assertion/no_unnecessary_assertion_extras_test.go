package no_unnecessary_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Every expectation below was recorded from eslint-plugin-jest 29.16.6 under
// @typescript-eslint/parser with a strict project, including report ranges.

// Call shapes.
func TestNoUnnecessaryAssertionCallShapes(t *testing.T) {
	rule_tester.RunRuleTester(
		typeRoot(), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: "expect().toBeNull();"},
			{Code: "expect.soft('a').toBeNull();"},
			{Code: "expect('a').resolves.toBeNull();"},
			{Code: "expect('a').toBeNull;"},
			{Code: "expect.not.toBeNull();"},
			{Code: "jest.expect('a').toBeNull();"},
			{Code: "expect('a').not.not.toBeNull();"},
			{Code: "expect('a').resolves.not.toBeNull();"},
			{Code: "expect('a').toHaveBeenCalled();"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "expect('a')['toBeNull']();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 26}}},
			{Code: "expect('a')[`toBeNull`]();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 26}}},
			{Code: "expect(('a')).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 25}}},
			{Code: "expect('a', null).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 29}}},
			{Code: "expect('a')?.toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 24}}},
			{Code: "expect('a').not.toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 27}}},
			{Code: "expect(\n  'a'\n).not\n.toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 4, EndColumn: 12}}},
			{Code: "expect('a').toBeDefined(1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 27}}},
			{Code: "describe('x', () => { it('y', () => { expect('a').toBeNull(); }); });", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 39, EndLine: 1, EndColumn: 61}}},
			{Code: "expect('a').toBeUndefined().toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 39}, {MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 28}}},
			{Code: "(expect)('a').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 25}}},
			{Code: "expect('a').not['toBeNull']();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}},
			{Code: "expect('a')['not'].toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}},
			{Code: "expect('a').toBeNull?.();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 25}}},
			{Code: "expect?.('a').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 25}}},
			{Code: "expect(/* c */ 'a' /* d */).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 39}}},
			{Code: "expect('a').toBeNull(); expect('b').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}, {MessageId: "unnecessaryAssertion", Line: 1, Column: 25, EndLine: 1, EndColumn: 47}}},
		},
	)
}

// Expect provenance.
func TestNoUnnecessaryAssertionProvenance(t *testing.T) {
	rule_tester.RunRuleTester(
		typeRoot(), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: "function g(expect: any){ expect('a').toBeNull(); }"},
			{Code: "import * as g from '@jest/globals'; g.expect('a').toBeNull();"},
			{Code: "import { expect } from 'vitest'; expect('a').toBeNull();"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "import {expect as e} from '@jest/globals'; e('a').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 44, EndLine: 1, EndColumn: 61}}},
			{Code: "const {expect: ex} = require('@jest/globals'); ex('a').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 48, EndLine: 1, EndColumn: 66}}},
			{Code: "import { expect } from '@jest/globals'; expect('a').toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 41, EndLine: 1, EndColumn: 62}}},
		},
	)
}

// Spread subjects.
func TestNoUnnecessaryAssertionSpreadSubjects(t *testing.T) {
	rule_tester.RunRuleTester(
		typeRoot(), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: "declare const a: [null]; expect(...a).toBeNull();"},
			{Code: "expect(...([null] as const)).toBeNull();"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "declare const a: string[]; expect(...a).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 28, EndLine: 1, EndColumn: 51}}},
			{Code: "expect(...[]).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 25}}},
		},
	)
}

// Subject types. Only the top-level type flags, or those of each union
// member, are inspected: type parameters, intersections, void and object types
// such as {} are not expanded.
func TestNoUnnecessaryAssertionSubjectTypes(t *testing.T) {
	rule_tester.RunRuleTester(
		typeRoot(), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: "function f<T extends string|null>(v: T){ expect(v).toBeNull(); }"},
			{Code: "enum E {A,B} declare const e: E; expect(e).toBeNaN();"},
			{Code: "expect(undefined).toBeDefined();"},
			{Code: "expect(void 0).toBeDefined();"},
			{Code: "function h<T>(k: keyof T){ expect(k).toBeNaN(); }"},
			{Code: "expect(NaN).toBeNaN();"},
			{Code: "expect(<any>'a').toBeNull();"},
			{Code: "declare const e: 1 | 2; expect(e).toBeNaN();"},
			{Code: "declare const v: any[]; expect(v[0]).toBeNull();"},
			{Code: "declare const o: { a?: string }; expect(o.a).toBeUndefined();"},
			{Code: "declare const o: { a?: string }; expect(o.a).toBeDefined();"},
			{Code: "function f(a?: string) { expect(a).toBeUndefined(); }"},
			{Code: "expect(1 as const).toBeNaN();"},
			{Code: "declare const x: null; expect(x).not.toBeNull();"},
			{Code: "test.each([1])('x', (v) => { expect(v).toBeNull(); });"},
			{Code: "declare function mx(): Promise<null>; async function t(){ expect(await mx()).toBeNull(); }"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "function f<T>(v: T){ expect(v).toBeNull(); }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 22, EndLine: 1, EndColumn: 42}}},
			{Code: "declare function f(): void; expect(f()).toBeUndefined();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 29, EndLine: 1, EndColumn: 56}}},
			{Code: "declare const b: boolean; expect(b).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 27, EndLine: 1, EndColumn: 47}}},
			{Code: "declare const n: number & {x:1}; expect(n).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 34, EndLine: 1, EndColumn: 53}}},
			{Code: "declare const o: {}; expect(o).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 22, EndLine: 1, EndColumn: 41}}},
			{Code: "declare const o: Number; expect(o).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 26, EndLine: 1, EndColumn: 45}}},
			{Code: "declare const x: unknown; expect(x as string).toBeUndefined();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 27, EndLine: 1, EndColumn: 62}}},
			{Code: "declare const x: 1n; expect(x).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 22, EndLine: 1, EndColumn: 41}}},
			{Code: "declare const x: symbol; expect(x).toBeUndefined();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 26, EndLine: 1, EndColumn: 51}}},
			{Code: "expect(null!).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 25}}},
			{Code: "declare const k: keyof {a:1}; expect(k).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 31, EndLine: 1, EndColumn: 50}}},
			{Code: "declare const s: string | number & {}; expect(s).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 40, EndLine: 1, EndColumn: 59}}},
			{Code: "expect(x => x).toBeUndefined();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 31}}},
			{Code: "declare const u: string | undefined; expect(u!).toBeDefined();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 38, EndLine: 1, EndColumn: 62}}},
			{Code: "let v: string | null = null as any; v ??= 'a'; expect(v).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 48, EndLine: 1, EndColumn: 68}}},
			{Code: "expect('a' satisfies string).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 40}}},
			{Code: "declare const e: E; const enum E { A = 'a' } expect(e).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 46, EndLine: 1, EndColumn: 65}}},
			{Code: "declare const r: Record<string, string>; expect(r.x).toBeUndefined();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 42, EndLine: 1, EndColumn: 69}}},
			{Code: "expect(Symbol()).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndLine: 1, EndColumn: 28}}},
			{Code: "declare const x: undefined; expect(x).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 29, EndLine: 1, EndColumn: 49}}},
			{Code: "declare const s: `a${string}`; expect(s).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 32, EndLine: 1, EndColumn: 52}}},
			{Code: "declare const s: unique symbol; expect(s).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 33, EndLine: 1, EndColumn: 53}}},
			{Code: "declare const t: [number]; expect(t).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 28, EndLine: 1, EndColumn: 47}}},
			{Code: "declare let x: string | null; x = 'a'; expect(x).toBeNull();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Line: 1, Column: 40, EndLine: 1, EndColumn: 60}}},
		},
	)
}

// strictNullChecks may come from `strict` or be set on its own, and an explicit
// `strictNullChecks: false` wins over `strict`. Assertions are still checked
// after the compiler-option diagnostic.
func TestNoUnnecessaryAssertionCompilerOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		typeRoot(), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: "expect(value).toBe(other)", TSConfig: "tsconfig.strict-null.json"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "expect(value).toBe(other)", TSConfig: "tsconfig.override.json", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStrictNullCheck", Line: 1, Column: 1}}},
			{
				Code:     "expect('a').toBeNull();",
				TSConfig: "tsconfig.unstrict.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noStrictNullCheck", Line: 1, Column: 1},
					{MessageId: "unnecessaryAssertion", Line: 1, Column: 1, EndColumn: 23},
				},
			},
		},
	)
}
