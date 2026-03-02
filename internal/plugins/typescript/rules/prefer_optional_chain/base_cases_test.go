package prefer_optional_chain

import (
	"fmt"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

type baseCase struct {
	id          int
	chain       string
	declaration string
	outputChain string
}

type baseCaseOptions struct {
	operator           string // "&&" or "||"
	mutateCode         func(string) string
	mutateDeclaration  func(string) string
	mutateOutput       func(string) string
	skipIds            []int
	useSuggestionFixer bool
}

func rawBaseCases(operator string) []baseCase {
	return []baseCase{
		// chained members
		{
			id:          1,
			chain:       fmt.Sprintf("foo %s foo.bar;", operator),
			declaration: "declare const foo: {bar: number} | null | undefined;",
			outputChain: "foo?.bar;",
		},
		{
			id:          2,
			chain:       fmt.Sprintf("foo.bar %s foo.bar.baz;", operator),
			declaration: "declare const foo: {bar: {baz: number} | null | undefined};",
			outputChain: "foo.bar?.baz;",
		},
		{
			id:          3,
			chain:       fmt.Sprintf("foo %s foo();", operator),
			declaration: "declare const foo: (() => number) | null | undefined;",
			outputChain: "foo?.();",
		},
		{
			id:          4,
			chain:       fmt.Sprintf("foo.bar %s foo.bar();", operator),
			declaration: "declare const foo: {bar: (() => number) | null | undefined};",
			outputChain: "foo.bar?.();",
		},
		{
			id:          5,
			chain:       fmt.Sprintf("foo %s foo.bar %s foo.bar.baz %s foo.bar.baz.buzz;", operator, operator, operator),
			declaration: "declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;",
			outputChain: "foo?.bar?.baz?.buzz;",
		},
		{
			id:          6,
			chain:       fmt.Sprintf("foo.bar %s foo.bar.baz %s foo.bar.baz.buzz;", operator, operator),
			declaration: "declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined};",
			outputChain: "foo.bar?.baz?.buzz;",
		},
		// case with a jump (i.e. a non-nullish prop)
		{
			id:          7,
			chain:       fmt.Sprintf("foo %s foo.bar %s foo.bar.baz.buzz;", operator, operator),
			declaration: "declare const foo: {bar: {baz: {buzz: number}} | null | undefined} | null | undefined;",
			outputChain: "foo?.bar?.baz.buzz;",
		},
		{
			id:          8,
			chain:       fmt.Sprintf("foo.bar %s foo.bar.baz.buzz;", operator),
			declaration: "declare const foo: {bar: {baz: {buzz: number}} | null | undefined};",
			outputChain: "foo.bar?.baz.buzz;",
		},
		// case where for some reason there is a doubled up expression
		{
			id:          9,
			chain:       fmt.Sprintf("foo %s foo.bar %s foo.bar.baz %s foo.bar.baz %s foo.bar.baz.buzz;", operator, operator, operator, operator),
			declaration: "declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;",
			outputChain: "foo?.bar?.baz?.buzz;",
		},
		{
			id:          10,
			chain:       fmt.Sprintf("foo.bar %s foo.bar.baz %s foo.bar.baz %s foo.bar.baz.buzz;", operator, operator, operator),
			declaration: "declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;",
			outputChain: "foo.bar?.baz?.buzz;",
		},
		// chained members with element access
		{
			id:          11,
			chain:       fmt.Sprintf("foo %s foo[bar] %s foo[bar].baz %s foo[bar].baz.buzz;", operator, operator, operator),
			declaration: "declare const bar: string;\ndeclare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;",
			outputChain: "foo?.[bar]?.baz?.buzz;",
		},
		{
			id:          12,
			chain:       fmt.Sprintf("foo %s foo[bar].baz %s foo[bar].baz.buzz;", operator, operator),
			declaration: "declare const bar: string;\ndeclare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;",
			outputChain: "foo?.[bar].baz?.buzz;",
		},
		// case with a property access in computed property
		{
			id:          13,
			chain:       fmt.Sprintf("foo %s foo[bar.baz] %s foo[bar.baz].buzz;", operator, operator),
			declaration: "declare const bar: {baz: string};\ndeclare const foo: {[k: string]: {buzz: number} | null | undefined} | null | undefined;",
			outputChain: "foo?.[bar.baz]?.buzz;",
		},
		// chained calls
		{
			id:          14,
			chain:       fmt.Sprintf("foo %s foo.bar %s foo.bar.baz %s foo.bar.baz.buzz();", operator, operator, operator),
			declaration: "declare const foo: {bar: {baz: {buzz: () => number} | null | undefined} | null | undefined} | null | undefined;",
			outputChain: "foo?.bar?.baz?.buzz();",
		},
		{
			id:          15,
			chain:       fmt.Sprintf("foo %s foo.bar %s foo.bar.baz %s foo.bar.baz.buzz %s foo.bar.baz.buzz();", operator, operator, operator, operator),
			declaration: "declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;",
			outputChain: "foo?.bar?.baz?.buzz?.();",
		},
		{
			id:          16,
			chain:       fmt.Sprintf("foo.bar %s foo.bar.baz %s foo.bar.baz.buzz %s foo.bar.baz.buzz();", operator, operator, operator),
			declaration: "declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined};",
			outputChain: "foo.bar?.baz?.buzz?.();",
		},
		// case with a jump (i.e. a non-nullish prop)
		{
			id:          17,
			chain:       fmt.Sprintf("foo %s foo.bar %s foo.bar.baz.buzz();", operator, operator),
			declaration: "declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined} | null | undefined;",
			outputChain: "foo?.bar?.baz.buzz();",
		},
		{
			id:          18,
			chain:       fmt.Sprintf("foo.bar %s foo.bar.baz.buzz();", operator),
			declaration: "declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined};",
			outputChain: "foo.bar?.baz.buzz();",
		},
		{
			id:          19,
			chain:       fmt.Sprintf("foo %s foo.bar %s foo.bar.baz.buzz %s foo.bar.baz.buzz();", operator, operator, operator),
			declaration: "declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined}} | null | undefined} | null | undefined;",
			outputChain: "foo?.bar?.baz.buzz?.();",
		},
		{
			id:          20,
			chain:       fmt.Sprintf("foo.bar %s foo.bar() %s foo.bar().baz %s foo.bar().baz.buzz %s foo.bar().baz.buzz();", operator, operator, operator, operator),
			declaration: "declare const foo: {bar: () => ({baz: {buzz: (() => number) | null | undefined} | null | undefined}) | null | undefined};",
			outputChain: "foo.bar?.()?.baz?.buzz?.();",
		},
		// chained calls with element access
		{
			id:          21,
			chain:       fmt.Sprintf("foo %s foo.bar %s foo.bar.baz %s foo.bar.baz[buzz]();", operator, operator, operator),
			declaration: "declare const buzz: string;\ndeclare const foo: {bar: {baz: {[k: string]: () => number} | null | undefined} | null | undefined} | null | undefined;",
			outputChain: "foo?.bar?.baz?.[buzz]();",
		},
		{
			id:          22,
			chain:       fmt.Sprintf("foo %s foo.bar %s foo.bar.baz %s foo.bar.baz[buzz] %s foo.bar.baz[buzz]();", operator, operator, operator, operator),
			declaration: "declare const buzz: string;\ndeclare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;",
			outputChain: "foo?.bar?.baz?.[buzz]?.();",
		},
		// (partially) pre-optional chained
		{
			id:          23,
			chain:       fmt.Sprintf("foo %s foo?.bar %s foo?.bar.baz %s foo?.bar.baz[buzz] %s foo?.bar.baz[buzz]();", operator, operator, operator, operator),
			declaration: "declare const buzz: string;\ndeclare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;",
			outputChain: "foo?.bar?.baz?.[buzz]?.();",
		},
		{
			id:          24,
			chain:       fmt.Sprintf("foo %s foo?.bar.baz %s foo?.bar.baz[buzz];", operator, operator),
			declaration: "declare const buzz: string;\ndeclare const foo: {bar: {baz: {[k: string]: number} | null | undefined}} | null | undefined;",
			outputChain: "foo?.bar.baz?.[buzz];",
		},
		{
			id:          25,
			chain:       fmt.Sprintf("foo %s foo?.() %s foo?.().bar;", operator, operator),
			declaration: "declare const foo: (() => ({bar: number} | null | undefined)) | null | undefined;",
			outputChain: "foo?.()?.bar;",
		},
		{
			id:          26,
			chain:       fmt.Sprintf("foo.bar %s foo.bar?.() %s foo.bar?.().baz;", operator, operator),
			declaration: "declare const foo: {bar: () => ({baz: number} | null | undefined)};",
			outputChain: "foo.bar?.()?.baz;",
		},
	}
}

func identity(s string) string {
	return s
}

func generateInvalidBaseCases(opts baseCaseOptions) []rule_tester.InvalidTestCase {
	mutateCode := opts.mutateCode
	if mutateCode == nil {
		mutateCode = identity
	}
	mutateDeclaration := opts.mutateDeclaration
	if mutateDeclaration == nil {
		mutateDeclaration = identity
	}
	mutateOutput := opts.mutateOutput
	if mutateOutput == nil {
		mutateOutput = mutateCode
	}

	skipIdsSet := make(map[int]bool)
	for _, id := range opts.skipIds {
		skipIdsSet[id] = true
	}

	cases := rawBaseCases(opts.operator)
	var result []rule_tester.InvalidTestCase

	for _, bc := range cases {
		if skipIdsSet[bc.id] {
			continue
		}

		declaration := mutateDeclaration(bc.declaration)
		code := fmt.Sprintf("// %d\n%s\n%s", bc.id, declaration, mutateCode(bc.chain))
		output := fmt.Sprintf("// %d\n%s\n%s", bc.id, declaration, mutateOutput(bc.outputChain))

		if opts.useSuggestionFixer {
			result = append(result, rule_tester.InvalidTestCase{
				Code: code,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferOptionalChain",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "optionalChainSuggest", Output: output},
						},
					},
				},
			})
		} else {
			result = append(result, rule_tester.InvalidTestCase{
				Code:   code,
				Output: []string{output},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
				},
			})
		}
	}

	return result
}

func generateValidBaseCases(opts baseCaseOptions) []rule_tester.ValidTestCase {
	mutateCode := opts.mutateCode
	if mutateCode == nil {
		mutateCode = identity
	}
	mutateDeclaration := opts.mutateDeclaration
	if mutateDeclaration == nil {
		mutateDeclaration = identity
	}

	skipIdsSet := make(map[int]bool)
	for _, id := range opts.skipIds {
		skipIdsSet[id] = true
	}

	cases := rawBaseCases(opts.operator)
	var result []rule_tester.ValidTestCase

	for _, bc := range cases {
		if skipIdsSet[bc.id] {
			continue
		}

		declaration := mutateDeclaration(bc.declaration)
		code := fmt.Sprintf("// %d\n%s\n%s", bc.id, declaration, mutateCode(bc.chain))

		result = append(result, rule_tester.ValidTestCase{
			Code: code,
		})
	}

	return result
}

