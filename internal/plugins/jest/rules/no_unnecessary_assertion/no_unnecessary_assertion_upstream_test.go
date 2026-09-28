package no_unnecessary_assertion

import (
	"fmt"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// typeRoot adds the projects used by the strictNullChecks cases to the shared
// Jest fixtures.
func typeRoot() rule_tester.Root {
	root := fixtures.GetRootDir()
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "tsconfig.unstrict.json"):    `{"extends":"./tsconfig.json","compilerOptions":{"strict":false}}`,
		tspath.ResolvePath(root.Dir, "tsconfig.override.json"):    `{"extends":"./tsconfig.json","compilerOptions":{"strictNullChecks":false}}`,
		tspath.ResolvePath(root.Dir, "tsconfig.strict-null.json"): `{"extends":"./tsconfig.json","compilerOptions":{"strict":false,"strictNullChecks":true}}`,
	})
	return root
}

func diagnostic(thing string, line int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unnecessaryAssertion",
		Message:   "Unnecessary assertion, subject cannot be " + thing,
		Line:      line,
	}
}

func errors(thing string, lines ...int) []rule_tester.InvalidTestCaseError {
	result := make([]rule_tester.InvalidTestCaseError, 0, len(lines))
	for _, line := range lines {
		result = append(result, diagnostic(thing, line))
	}
	return result
}

func TestNoUnnecessaryAssertionGeneralUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		typeRoot(), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
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
		{Code: "const add = (a, b) => a + b;\n\nexpect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number) => a + b;\n\nexpect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number): number => a + b;\n\nexpect(add(1, 1)).toBe(2);"},
		{Code: fmt.Sprintf("const add = (a: number, b: number): number | %s => a + b;\n\nexpect(add(1, 1)).not.%s();", thing, matcher)},
		{Code: fmt.Sprintf("declare function mx(): any;\n\nexpect(mx()).%s();\nexpect(mx()).not.%s();", matcher, matcher)},
		{Code: fmt.Sprintf("declare function mx(): unknown;\n\nexpect(mx()).%s();\nexpect(mx()).not.%s();", matcher, matcher)},
		{Code: fmt.Sprintf("declare function mx(): string | %s;\n\nexpect(mx()).%s();\nexpect(mx()).not.%s();", thing, matcher, matcher)},
		{Code: fmt.Sprintf("declare function mx<T>(p: T): T | %s;\n\nexpect(mx(%s)).%s();\nexpect(mx(%s)).not.%s();\n\nexpect(mx('hello')).%s();\nexpect(mx('world')).not.%s();", thing, thing, matcher, thing, matcher, matcher, matcher)},
		{Code: fmt.Sprintf("expect(%s).not.%s()", thing, matcher)},
		{Code: fmt.Sprintf("expect(\"hello\" as %s).%s();", thing, matcher)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string | %s>;\n\nit('is async', async () => {\n  await expect(mx()).resolves.%s();\n});", thing, matcher)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string | %s>;\n\nit('is async', async () => {\n  await expect(mx()).rejects.%s();\n});", thing, matcher)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string | %s>;\n\nit('is async', async () => {\n  await expect(mx()).rejects.not.%s();\n});", thing, matcher)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string | %s>;\n\nit('is async', async () => {\n  expect(await mx()).not.%s();\n});", thing, matcher)},
		// Upstream todo: promise subjects are not resolved yet.
		{Code: fmt.Sprintf("declare function mx(): Promise<string>;\n\nit('is async', async () => {\n  await expect(mx()).resolves.%s();\n});", matcher)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string>;\n\nit('is async', async () => {\n  await expect(mx()).rejects.not.%s();\n});", matcher)},
	}

	invalid := []rule_tester.InvalidTestCase{
		{Code: fmt.Sprintf("expect(0).%s()", matcher), Errors: errors(thing, 1)},
		{Code: fmt.Sprintf("expect(\"hello world\").%s()", matcher), Errors: errors(thing, 1)},
		{Code: fmt.Sprintf("expect({}).%s()", matcher), Errors: errors(thing, 1)},
		{Code: fmt.Sprintf("expect([]).not.%s()", matcher), Errors: errors(thing, 1)},
		{Code: fmt.Sprintf("const x = 0;\n\nexpect(x).%s()\nexpect(x).not.%s()", matcher, matcher), Errors: errors(thing, 3, 4)},
		{Code: fmt.Sprintf("const add = (a: number, b: number) => a + b;\n\nexpect(add(1, 1)).%s();", matcher), Errors: errors(thing, 3)},
		{Code: fmt.Sprintf("const add = (a: number, b: number): number => a + b;\n\nexpect(add(1, 1)).%s();", matcher), Errors: errors(thing, 3)},
		{Code: fmt.Sprintf("const add = (a: number, b: number): number => a + b;\n\nexpect(add(1, 1)).not.%s();", matcher), Errors: errors(thing, 3)},
		{Code: fmt.Sprintf("declare function mx(): never;\n\nexpect(mx()).not.%s();", matcher), Errors: errors(thing, 3)},
		{Code: fmt.Sprintf("const result = \"hello world\".match(\"sunshine\") ?? [];\n\nexpect(result).not.%s();\nexpect(result).%s();", matcher, matcher), Errors: errors(thing, 3, 4)},
		{Code: fmt.Sprintf("const result = \"hello world\".match(\"sunshine\") || [];\n\nexpect(result).not.%s();\nexpect(result).%s();", matcher, matcher), Errors: errors(thing, 3, 4)},
		{Code: fmt.Sprintf("declare function mx(): Promise<string>;\n\nit('is async', async () => {\n  expect(await mx()).%s();\n});", matcher), Errors: errors(thing, 4)},
		{Code: fmt.Sprintf("declare function mx(): string | number;\n\nexpect(mx()).%s();\nexpect(mx()).not.%s();", matcher, matcher), Errors: errors(thing, 3, 4)},
		{
			Code:   fmt.Sprintf("declare function mx<T>(p: T): T;\n\nexpect(mx(%s)).%s();\nexpect(mx(%s)).not.%s();\n\nexpect(mx('hello')).%s();\nexpect(mx('world')).not.%s();", thing, matcher, thing, matcher, matcher, matcher),
			Errors: errors(thing, 6, 7),
		},
		{
			Code:   fmt.Sprintf("declare function mx<T>(p: T): T extends string ? %s : T;\n\nexpect(mx('hello')).%s();\nexpect(mx('world')).not.%s();\n\nexpect(mx({})).%s();\nexpect(mx({})).not.%s();", thing, matcher, matcher, matcher, matcher),
			Errors: errors(thing, 6, 7),
		},
		{
			Code:   fmt.Sprintf("declare function mx(): string | %s;\n\nexpect(mx()!).%s();\nexpect(mx()!).not.%s();\n\nexpect(mx() as string).%s();\nexpect(mx() as string).not.%s();\n\nexpect(mx() as number).%s();\nexpect(mx() as number).not.%s();", thing, matcher, matcher, matcher, matcher, matcher, matcher),
			Errors: errors(thing, 3, 4, 6, 7, 9, 10),
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
			rule_tester.RunRuleTester(typeRoot(), "tsconfig.json", t, &NoUnnecessaryAssertionRule, valid, invalid)
		})
	}
}

func TestNoUnnecessaryAssertionNaNUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "expect.toBeNaN"},
		{Code: "expect.toBeNaN()"},
		{Code: "expect(0).toBeNaN()"},
		{Code: "expect(0).not.toBeNaN()"},
		{Code: "const x = 0;\n\nexpect(x).toBeNaN()\nexpect(x).not.toBeNaN()"},
		{Code: "const add = (a, b) => a + b;\n\nexpect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number) => a.toString() + b.toString();\n\nexpect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number): string => a.toString() + b.toString();\n\nexpect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number): string | number => a.toString() + b.toString();\n\nexpect(add(1, 1)).toBe(2);"},
		{Code: "const add = (a: number, b: number): string | number => a.toString() + b.toString();\n\nexpect(add(1, 1)).toBeNaN();"},
		{Code: "const add = (a: number, b: number): string | number => a + b;\n\nexpect(add(1, 1)).not.toBeNaN();"},
		{Code: "declare function mx(): string | number;\n\nexpect(mx()).toBeNaN();\nexpect(mx()).not.toBeNaN();"},
		{Code: "declare function mx<T>(p: T): T | number;\n\nexpect(mx(42)).toBeNaN();\nexpect(mx(4.2)).toBeNaN();\nexpect(mx(Infinity)).not.toBeNaN();\n\nexpect(mx('hello')).toBeNaN();\nexpect(mx('world')).not.toBeNaN();"},
		{Code: "expect(\"hello\" as number).toBeNaN();"},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: "expect(\"hello world\").toBeNaN()", Errors: errors("a number", 1)},
		{Code: "expect({}).toBeNaN()", Errors: errors("a number", 1)},
		{Code: "expect([]).not.toBeNaN()", Errors: errors("a number", 1)},
		{Code: "const x = 'hello world';\n\nexpect(x).toBeNaN()\nexpect(x).not.toBeNaN()", Errors: errors("a number", 3, 4)},
		{Code: "const join = (a: string, b: string) => a + b;\n\nexpect(join('hello', 'world')).toBeNaN();", Errors: errors("a number", 3)},
		{Code: "const join = (a: string, b: string): string => a + b;\n\nexpect(join('hello', 'world')).toBeNaN();", Errors: errors("a number", 3)},
		{Code: "const join = (a: string, b: string): string => a + b;\n\nexpect(join('hello', 'world')).not.toBeNaN();", Errors: errors("a number", 3)},
		{Code: "const result = \"hello world\".match(\"sunshine\") ?? [];\n\nexpect(result).not.toBeNaN();\nexpect(result).toBeNaN();", Errors: errors("a number", 3, 4)},
		{Code: "const result = \"hello world\".match(\"sunshine\") || [];\n\nexpect(result).not.toBeNaN();\nexpect(result).toBeNaN();", Errors: errors("a number", 3, 4)},
		{Code: "declare function mx(): string | null;\n\nexpect(mx()).toBeNaN();\nexpect(mx()).not.toBeNaN();", Errors: errors("a number", 3, 4)},
		{
			Code:   "declare function mx<T>(p: T): T;\n\nexpect(mx(0)).toBeNaN();\nexpect(mx(1)).not.toBeNaN();\nexpect(mx(NaN)).not.toBeNaN();\n\nexpect(mx('hello')).toBeNaN();\nexpect(mx('world')).not.toBeNaN();",
			Errors: errors("a number", 7, 8),
		},
		{
			Code:   "declare function mx<T>(p: T): T extends string ? number : T;\n\nexpect(mx('hello')).toBeNaN();\nexpect(mx('world')).not.toBeNaN();\n\nexpect(mx({})).toBeNaN();\nexpect(mx({})).not.toBeNaN();",
			Errors: errors("a number", 6, 7),
		},
		{
			Code:   "declare function mx(): string | number;\n\nexpect(mx() as string).toBeNaN();\nexpect(mx() as string).not.toBeNaN();\n\nexpect(mx() as number).toBeNaN();\nexpect(mx() as number).not.toBeNaN();",
			Errors: errors("a number", 3, 4),
		},
	}
	rule_tester.RunRuleTester(typeRoot(), "tsconfig.json", t, &NoUnnecessaryAssertionRule, valid, invalid)
}
