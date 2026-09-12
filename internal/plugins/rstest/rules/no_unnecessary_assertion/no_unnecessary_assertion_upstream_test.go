package no_unnecessary_assertion

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func diagnostic(thing string, line int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unnecessaryAssertion",
		Message:   "Unnecessary assertion, subject cannot be " + thing,
		Line:      line,
	}
}

func TestNoUnnecessaryAssertionGeneralUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		typeRoot(t), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: "expect"},
			{Code: "expect.hasAssertions"},
			{Code: "expect.hasAssertions()"},
			{Code: "expect(a).toBe(b)"},
		},
		[]rule_tester.InvalidTestCase{{
			Code:     "expect(x).toBe(y);",
			TSConfig: "tsconfig.unstrict.json",
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noStrictNullCheck"}},
		}},
	)
}

func nullishUpstreamCases(matcher, thing string) ([]rule_tester.ValidTestCase, []rule_tester.InvalidTestCase) {
	valid := []rule_tester.ValidTestCase{
		{Code: "expect." + matcher},
		{Code: fmt.Sprintf("expect.%s()", matcher)},
		{Code: "const add = (a, b) => a + b; expect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number) => a + b; expect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number): number => a + b; expect(add(1, 1)).toBe(2);"},
		{Code: fmt.Sprintf("const add = (a: number, b: number): number | %s => a + b; expect(add(1, 1)).not.%s();", thing, matcher)},
		{Code: fmt.Sprintf("declare function mx(): any; expect(mx()).%s(); expect(mx()).not.%s();", matcher, matcher)},
		{Code: fmt.Sprintf("declare function mx(): unknown; expect(mx()).%s(); expect(mx()).not.%s();", matcher, matcher)},
		{Code: fmt.Sprintf("declare function mx(): string | %s; expect(mx()).%s(); expect(mx()).not.%s();", thing, matcher, matcher)},
		{Code: fmt.Sprintf("declare function mx<T>(p: T): T | %s; expect(mx(%s)).%s(); expect(mx(%s)).not.%s(); expect(mx('hello')).%s(); expect(mx('world')).not.%s();", thing, thing, matcher, thing, matcher, matcher, matcher)},
		{Code: fmt.Sprintf("expect(%s).not.%s()", thing, matcher)},
		{Code: fmt.Sprintf("expect('hello' as %s).%s();", thing, matcher)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string | %s>; async function test() { await expect(mx()).resolves.%s(); }", thing, matcher)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string | %s>; async function test() { await expect(mx()).rejects.%s(); }", thing, matcher)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string | %s>; async function test() { await expect(mx()).rejects.not.%s(); }", thing, matcher)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string | %s>; async function test() { expect(await mx()).not.%s(); }", thing, matcher)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string>; async function test() { await expect(mx()).resolves.%s(); }", matcher)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string>; async function test() { await expect(mx()).rejects.not.%s(); }", matcher)},
	}

	invalid := []rule_tester.InvalidTestCase{
		{Code: fmt.Sprintf("expect(0).%s()", matcher), Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 1)}},
		{Code: fmt.Sprintf("expect('hello world').%s()", matcher), Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 1)}},
		{Code: fmt.Sprintf("expect({}).%s()", matcher), Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 1)}},
		{Code: fmt.Sprintf("expect([]).not.%s()", matcher), Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 1)}},
		{
			Code:   fmt.Sprintf("const x = 0;\nexpect(x).%s()\nexpect(x).not.%s()", matcher, matcher),
			Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 2), diagnostic(thing, 3)},
		},
		{Code: fmt.Sprintf("const add = (a: number, b: number) => a + b;\nexpect(add(1, 1)).%s();", matcher), Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 2)}},
		{Code: fmt.Sprintf("const add = (a: number, b: number): number => a + b;\nexpect(add(1, 1)).%s();", matcher), Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 2)}},
		{Code: fmt.Sprintf("const add = (a: number, b: number): number => a + b;\nexpect(add(1, 1)).not.%s();", matcher), Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 2)}},
		{Code: fmt.Sprintf("declare function mx(): never;\nexpect(mx()).not.%s();", matcher), Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 2)}},
		{
			Code:   fmt.Sprintf("const result = 'hello world'.match('sunshine') ?? [];\nexpect(result).not.%s();\nexpect(result).%s();", matcher, matcher),
			Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 2), diagnostic(thing, 3)},
		},
		{
			Code:   fmt.Sprintf("const result = 'hello world'.match('sunshine') || [];\nexpect(result).not.%s();\nexpect(result).%s();", matcher, matcher),
			Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 2), diagnostic(thing, 3)},
		},
		{Code: fmt.Sprintf("declare function mx(): Promise<string>;\nasync function test() { expect(await mx()).%s(); }", matcher), Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 2)}},
		{
			Code:   fmt.Sprintf("declare function mx(): string | number;\nexpect(mx()).%s();\nexpect(mx()).not.%s();", matcher, matcher),
			Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 2), diagnostic(thing, 3)},
		},
		{
			Code:   fmt.Sprintf("declare function mx<T>(p: T): T;\nexpect(mx(%s)).%s();\nexpect(mx(%s)).not.%s();\nexpect(mx('hello')).%s();\nexpect(mx('world')).not.%s();", thing, matcher, thing, matcher, matcher, matcher),
			Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 4), diagnostic(thing, 5)},
		},
		{
			Code:   fmt.Sprintf("declare function mx<T>(p: T): T extends string ? %s : T;\nexpect(mx('hello')).%s();\nexpect(mx('world')).not.%s();\nexpect(mx({})).%s();\nexpect(mx({})).not.%s();", thing, matcher, matcher, matcher, matcher),
			Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 4), diagnostic(thing, 5)},
		},
		{
			Code:   fmt.Sprintf("declare function mx(): string | %s;\nexpect(mx()!).%s();\nexpect(mx()!).not.%s();\nexpect(mx() as string).%s();\nexpect(mx() as string).not.%s();\nexpect(mx() as number).%s();\nexpect(mx() as number).not.%s();", thing, matcher, matcher, matcher, matcher, matcher, matcher),
			Errors: []rule_tester.InvalidTestCaseError{diagnostic(thing, 2), diagnostic(thing, 3), diagnostic(thing, 4), diagnostic(thing, 5), diagnostic(thing, 6), diagnostic(thing, 7)},
		},
	}
	return valid, invalid
}

func TestNoUnnecessaryAssertionNullishUpstream(t *testing.T) {
	for _, test := range []struct{ matcher, thing string }{
		{matcher: "toBeNull", thing: "null"},
		{matcher: "toBeDefined", thing: "undefined"},
		{matcher: "toBeUndefined", thing: "undefined"},
	} {
		t.Run(test.matcher, func(t *testing.T) {
			valid, invalid := nullishUpstreamCases(test.matcher, test.thing)
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryAssertionRule, valid, invalid)
		})
	}
}

func TestNoUnnecessaryAssertionNaNUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "expect.toBeNaN"},
		{Code: "expect.toBeNaN()"},
		{Code: "expect(0).toBeNaN()"},
		{Code: "expect(0).not.toBeNaN()"},
		{Code: "const x = 0; expect(x).toBeNaN(); expect(x).not.toBeNaN()"},
		{Code: "const add = (a, b) => a + b; expect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number) => a.toString() + b.toString(); expect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number): string => a.toString() + b.toString(); expect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number): string | number => a.toString() + b.toString(); expect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number): string | number => a.toString() + b.toString(); expect(add(1, 1)).toBeNaN();"},
		{Code: "const add = (a: number, b: number): string | number => a + b; expect(add(1, 1)).not.toBeNaN();"},
		{Code: "declare function mx(): string | number; expect(mx()).toBeNaN(); expect(mx()).not.toBeNaN();"},
		{Code: "declare function mx<T>(p: T): T | number; expect(mx(42)).toBeNaN(); expect(mx(4.2)).toBeNaN(); expect(mx(Infinity)).not.toBeNaN(); expect(mx('hello')).toBeNaN(); expect(mx('world')).not.toBeNaN();"},
		{Code: "expect('hello' as number).toBeNaN();"},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: "expect('hello world').toBeNaN()", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 1)}},
		{Code: "expect({}).toBeNaN()", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 1)}},
		{Code: "expect([]).not.toBeNaN()", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 1)}},
		{Code: "const x = 'hello world';\nexpect(x).toBeNaN();\nexpect(x).not.toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 2), diagnostic("a number", 3)}},
		{Code: "const join = (a: string, b: string) => a + b;\nexpect(join('hello', 'world')).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 2)}},
		{Code: "const join = (a: string, b: string): string => a + b;\nexpect(join('hello', 'world')).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 2)}},
		{Code: "const join = (a: string, b: string): string => a + b;\nexpect(join('hello', 'world')).not.toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 2)}},
		{Code: "const result = 'hello world'.match('sunshine') ?? [];\nexpect(result).not.toBeNaN();\nexpect(result).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 2), diagnostic("a number", 3)}},
		{Code: "const result = 'hello world'.match('sunshine') || [];\nexpect(result).not.toBeNaN();\nexpect(result).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 2), diagnostic("a number", 3)}},
		{Code: "declare function mx(): string | null;\nexpect(mx()).toBeNaN();\nexpect(mx()).not.toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 2), diagnostic("a number", 3)}},
		{Code: "declare function mx<T>(p: T): T;\nexpect(mx(0)).toBeNaN();\nexpect(mx(1)).not.toBeNaN();\nexpect(mx(NaN)).not.toBeNaN();\nexpect(mx('hello')).toBeNaN();\nexpect(mx('world')).not.toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 5), diagnostic("a number", 6)}},
		{Code: "declare function mx<T>(p: T): T extends string ? number : T;\nexpect(mx('hello')).toBeNaN();\nexpect(mx('world')).not.toBeNaN();\nexpect(mx({})).toBeNaN();\nexpect(mx({})).not.toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 4), diagnostic("a number", 5)}},
		{Code: "declare function mx(): string | number;\nexpect(mx() as string).toBeNaN();\nexpect(mx() as string).not.toBeNaN();\nexpect(mx() as number).toBeNaN();\nexpect(mx() as number).not.toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 2), diagnostic("a number", 3)}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryAssertionRule, valid, invalid)
}
