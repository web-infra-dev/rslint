package consistent_type_assertions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Helpers to keep the option maps terse.
func opt(m map[string]interface{}) []interface{} { return []interface{}{m} }

func TestConsistentTypeAssertionsRule(t *testing.T) {
	angleStyle := opt(map[string]interface{}{"assertionStyle": "angle-bracket"})
	neverStyle := opt(map[string]interface{}{"assertionStyle": "never"})
	objNever := opt(map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "never"})
	objParam := opt(map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter"})
	arrNever := opt(map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "never"})
	arrParam := opt(map[string]interface{}{"assertionStyle": "as", "arrayLiteralTypeAssertions": "allow-as-parameter"})
	bothParam := opt(map[string]interface{}{"assertionStyle": "as", "objectLiteralTypeAssertions": "allow-as-parameter", "arrayLiteralTypeAssertions": "allow-as-parameter"})
	angleObjNever := opt(map[string]interface{}{"assertionStyle": "angle-bracket", "objectLiteralTypeAssertions": "never"})
	angleArrNever := opt(map[string]interface{}{"assertionStyle": "angle-bracket", "arrayLiteralTypeAssertions": "never"})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTypeAssertionsRule, []rule_tester.ValidTestCase{
		// Default options (assertionStyle: 'as', literals 'allow')
		{Code: `const x = b as A;`},
		{Code: `const x = [1] as readonly number[];`},
		{Code: `const x = 'string' as a | b;`},
		{Code: `const x = () => ({ bar: 5 }) as Foo;`},
		{Code: `const x = { key: 'value' } as const;`},
		{Code: `const x = [1] as const;`},
		{Code: `const x = { foo: 1 } as Foo;`},
		{Code: `const x = [] as Foo[];`},
		{Code: `const x = new Generic<int>() as Foo;`},
		{Code: `const x = (foo as Bar).baz;`},
		{Code: `const x = [1, 2, 3] as number[];`},

		// assertionStyle: 'angle-bracket'
		{Code: `const x = <A>b;`, Options: angleStyle},
		{Code: `const x = <readonly number[]>[1];`, Options: angleStyle},
		{Code: `const x = <a | b>'string';`, Options: angleStyle},
		{Code: `const x = <Foo>(() => ({ bar: 5 }));`, Options: angleStyle},
		{Code: `const x = <const>{ key: 'value' };`, Options: angleStyle},
		{Code: `const x = <const>[1];`, Options: angleStyle},
		{Code: `const x = <Foo>{ foo: 1 };`, Options: angleStyle},
		{Code: `const x = <Foo[]>[];`, Options: angleStyle},

		// assertionStyle: 'never' — `const` is exempt
		{Code: `const x = { key: 'value' } as const;`, Options: neverStyle},
		{Code: `const x = <const>{ key: 'value' };`, Options: neverStyle},
		{Code: `const x = [1] as const;`, Options: neverStyle},
		{Code: `const x = <const>[1];`, Options: neverStyle},

		// objectLiteralTypeAssertions: 'never' — bare any/unknown/const & non-objects exempt
		{Code: `const x: Foo = { bar: 5 };`, Options: objNever},
		{Code: `const x = { bar: 5 } as any;`, Options: objNever},
		{Code: `const x = { bar: 5 } as unknown;`, Options: objNever},
		{Code: `const x = { bar: 5 } as const;`, Options: objNever},
		{Code: `const x = 'string' as Foo;`, Options: objNever},
		{Code: `const x = 123 as Foo;`, Options: objNever},
		{Code: `const x = true as Foo;`, Options: objNever},

		// objectLiteralTypeAssertions: 'allow-as-parameter'
		{Code: `const x: Foo = { bar: 5 };`, Options: objParam},
		{Code: `foo({ bar: 5 } as Foo);`, Options: objParam},
		{Code: `new Foo({ bar: 5 } as Foo);`, Options: objParam},
		{Code: `throw { bar: 5 } as Foo;`, Options: objParam},
		{Code: `const x = { bar: 5 } as any;`, Options: objParam},
		{Code: `const x = { bar: 5 } as unknown;`, Options: objParam},
		{Code: `const x = { bar: 5 } as const;`, Options: objParam},
		{Code: `function foo() { throw { bar: 5 } as Foo; }`, Options: objParam},
		{Code: `const foo = (x = { bar: 5 } as Foo) => {};`, Options: objParam},
		{Code: `const foo = ({ x = { bar: 5 } as Foo } = {}) => {};`, Options: objParam},
		{Code: `const foo = ({ x = {} as Record<string, string[]> }) => {};`, Options: objParam},
		{Code: `const foo = ([x = {} as Foo]) => {};`, Options: objParam},
		{Code: `function b({ x = {} as Foo.Bar }) {}`, Options: objParam},
		{Code: `const foo = ({ a: { b = {} as Foo } = {} }) => {};`, Options: objParam},
		{Code: `print?.({ bar: 5 } as Foo);`, Options: objParam},
		{Code: `print?.call({ bar: 5 } as Foo);`, Options: objParam},
		{Code: "print`${{ bar: 5 } as Foo}`;", Options: objParam},
		{Code: `const bar = <Foo style={{ bar: 5 } as Bar} />;`, Tsx: true, Options: objParam},
		{Code: `foo(({ bar: 5 } as Foo));`, Options: objParam},

		// arrayLiteralTypeAssertions: 'never'
		{Code: `const x: string[] = [];`, Options: arrNever},
		{Code: `const x = [] as any;`, Options: arrNever},
		{Code: `const x = [] as unknown;`, Options: arrNever},
		{Code: `const x = [] as const;`, Options: arrNever},
		{Code: `const x = 'string' as Foo;`, Options: arrNever},

		// arrayLiteralTypeAssertions: 'allow-as-parameter'
		{Code: `const x: string[] = [];`, Options: arrParam},
		{Code: `const foo = ({ x = [] as string[] }) => {};`, Options: arrParam},
		{Code: `const foo = ([x = [] as string[]]) => {};`, Options: arrParam},
		{Code: `function b(x = [5] as Foo.Bar) {}`, Options: arrParam},
		{Code: `print?.([5] as Foo);`, Options: arrParam},
		{Code: `print?.call([5] as Foo);`, Options: arrParam},
		{Code: "print`${[5] as Foo}`;", Options: arrParam},
		{Code: `const bar = <Foo style={[5] as Bar} />;`, Tsx: true, Options: arrParam},
		{Code: `foo(([5] as Foo));`, Options: arrParam},
		{Code: `foo([] as string[]);`, Options: arrParam},
		{Code: `new Foo([] as string[]);`, Options: arrParam},
		{Code: `throw [] as string[];`, Options: arrParam},
		{Code: `const x = [] as any;`, Options: arrParam},
		{Code: `const x = [] as unknown;`, Options: arrParam},
		{Code: `const x = [] as const;`, Options: arrParam},
		{Code: `function foo() { throw [] as string[]; }`, Options: arrParam},

		// Both objectLiteralTypeAssertions and arrayLiteralTypeAssertions: 'allow-as-parameter'
		{Code: `const x: Foo = { bar: 5 };`, Options: bothParam},
		{Code: `const x: string[] = [];`, Options: bothParam},
		{Code: `foo({ bar: 5 } as Foo);`, Options: bothParam},
		{Code: `foo([] as string[]);`, Options: bothParam},
		{Code: `foo({ bar: 5 } as Foo, [] as string[]);`, Options: bothParam},

		// angle-bracket with literal 'never' — literals via angle bracket allowed
		{Code: `const x: Foo = { bar: 5 };`, Options: angleObjNever},
		{Code: `const x = <any>{ bar: 5 };`, Options: angleObjNever},
		{Code: `const x = <unknown>{ bar: 5 };`, Options: angleObjNever},
		{Code: `const x = <const>{ bar: 5 };`, Options: angleObjNever},
		{Code: `const x: string[] = [];`, Options: angleArrNever},
		{Code: `const x = <any>[];`, Options: angleArrNever},
		{Code: `const x = <unknown>[];`, Options: angleArrNever},
		{Code: `const x = <const>[];`, Options: angleArrNever},

		// Additional edge cases — assertions on non object/array literals
		{Code: `const x = value as string | number;`},
		{Code: `const x = value as (string | number)[];`},
		{Code: `const x = (value as Foo) as Bar;`},
		{Code: `const x = (value as Foo).bar;`},
		{Code: `const x = (value as Foo)();`},
		{Code: "const x = `template ${value as string}`;"},
	}, []rule_tester.InvalidTestCase{
		// ---- assertionStyle: 'as' (default) — angle bracket used, autofix to `as` ----
		{
			Code:   `const x = <A>b;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`const x = b as A;`},
		},
		{
			Code:   `const x = <readonly number[]>[1];`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`const x = [1] as readonly number[];`},
		},
		{
			Code:   `const x = <Foo>{ bar: 5 };`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`const x = { bar: 5 } as Foo;`},
		},
		{
			Code:   `const x = <Foo>[];`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`const x = [] as Foo;`},
		},
		{
			Code:   `const x = <const>{ key: 'value' };`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`const x = { key: 'value' } as const;`},
		},
		{
			Code:   `const x = <const>[1];`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`const x = [1] as const;`},
		},
		// angle-bracket assertion on an object literal under 'as' style reports the
		// style mismatch (NOT the object-literal rule), and autofixes to `as`.
		{
			Code:    `const x = <Foo>{ bar: 5 };`,
			Options: objNever,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output:  []string{`const x = { bar: 5 } as Foo;`},
		},
		// Autofix precedence wrapping: the result is parenthesized only where the
		// surrounding context requires it, and the expression only where its own
		// precedence is lower than `as`.
		{
			Code:   `const x = <T>a + b;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`const x = (a as T) + b;`},
		},
		{
			Code:   `foo(<T>x);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`foo((x as T));`},
		},
		{
			Code:   `const x = (<T>a).b;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`const x = (a as T).b;`},
		},
		{
			Code:   `const x = <T>(a ? b : c);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`const x = (a ? b : c) as T;`},
		},
		{
			Code:   `const x = -<T>a;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`const x = -(a as T);`},
		},
		{
			Code:   `const x = <T>a.b;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "as"}},
			Output: []string{`const x = a.b as T;`},
		},

		// ---- assertionStyle: 'angle-bracket' — `as` used ----
		{Code: `const x = b as A;`, Options: angleStyle, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "angle-bracket"}}},
		{Code: `const x = [1] as readonly number[];`, Options: angleStyle, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "angle-bracket"}}},
		{Code: `const x = { bar: 5 } as Foo;`, Options: angleStyle, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "angle-bracket"}}},
		{Code: `const x = { key: 'value' } as const;`, Options: angleStyle, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "angle-bracket"}}},
		{Code: `const x = [1] as const;`, Options: angleStyle, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "angle-bracket"}}},

		// ---- assertionStyle: 'never' ----
		{Code: `const x = b as A;`, Options: neverStyle, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never"}}},
		{Code: `const x = <A>b;`, Options: neverStyle, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never"}}},
		{Code: `const x = { bar: 5 } as Foo;`, Options: neverStyle, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never"}}},
		{Code: `const x = <Foo>{ bar: 5 };`, Options: neverStyle, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "never"}}},

		// ---- objectLiteralTypeAssertions: 'never' ----
		{
			Code:    `const x = {} as Foo;`,
			Options: objNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithAnnotation", Output: `const x: Foo = {};`},
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const x = {} satisfies Foo;`},
				},
			}},
		},
		{
			Code:    `const x = { bar: 5 } as Foo;`,
			Options: objNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithAnnotation", Output: `const x: Foo = { bar: 5 };`},
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const x = { bar: 5 } satisfies Foo;`},
				},
			}},
		},
		{
			Code:    `const x = { bar: 5 } as Foo<int>;`,
			Options: objNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithAnnotation", Output: `const x: Foo<int> = { bar: 5 };`},
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const x = { bar: 5 } satisfies Foo<int>;`},
				},
			}},
		},
		{
			Code:    `const x = { bar: 5 } as a | b;`,
			Options: objNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithAnnotation", Output: `const x: a | b = { bar: 5 };`},
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const x = { bar: 5 } satisfies a | b;`},
				},
			}},
		},
		// `any`/`unknown` inside a union ARE reported (only the bare keyword is exempt).
		{
			Code:    `const x = {} as any | string;`,
			Options: objNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithAnnotation", Output: `const x: any | string = {};`},
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const x = {} satisfies any | string;`},
				},
			}},
		},
		{
			Code:    `const x = {} as unknown | string;`,
			Options: objNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithAnnotation", Output: `const x: unknown | string = {};`},
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const x = {} satisfies unknown | string;`},
				},
			}},
		},
		// Qualified-name target (`Foo.Bar`) is reported.
		{
			Code:    `const x = {} as Foo.Bar;`,
			Options: objNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithAnnotation", Output: `const x: Foo.Bar = {};`},
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const x = {} satisfies Foo.Bar;`},
				},
			}},
		},
		// Parenthesized object literal — `({ ... }) as T` — must be detected. The
		// report spans from the opening paren (column 11).
		{
			Code:    `const x = ({}) as Foo;`,
			Options: objNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Line:      1, Column: 11, EndColumn: 22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithAnnotation", Output: `const x: Foo = ({});`},
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const x = ({}) satisfies Foo;`},
				},
			}},
		},
		// Nested parens.
		{
			Code:    `const x = (({})) as Foo;`,
			Options: objNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithAnnotation", Output: `const x: Foo = ({});`},
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const x = ({}) satisfies Foo;`},
				},
			}},
		},
		// Arrow-function return of a parenthesized object literal (the real-world
		// shape from the migration report). Only the `satisfies` suggestion applies
		// (the assertion does not initialize a plain variable).
		{
			Code:    `const f = (): Foo => ({ bar: 5 }) as Foo;`,
			Options: objNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const f = (): Foo => ({ bar: 5 }) satisfies Foo;`},
				},
			}},
		},
		// Assertion nested in a larger expression — only `satisfies` suggestion.
		{
			Code:    `const x = ({} as A) + b;`,
			Options: objNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const x = ({} satisfies A) + b;`},
				},
			}},
		},

		// ---- objectLiteralTypeAssertions: 'allow-as-parameter' (non-parameter positions) ----
		{
			Code:    `const x = { bar: 5 } as Foo;`,
			Options: objParam,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithAnnotation", Output: `const x: Foo = { bar: 5 };`},
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `const x = { bar: 5 } satisfies Foo;`},
				},
			}},
		},
		{
			Code:    `class C { x = {} as Foo; }`,
			Options: objParam,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `class C { x = {} satisfies Foo; }`},
				},
			}},
		},
		// A property value is NOT an "as parameter" — it is reported.
		{
			Code:    `foo({ x: {} as Foo });`,
			Options: objParam,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedObjectTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceObjectTypeAssertionWithSatisfies", Output: `foo({ x: {} satisfies Foo });`},
				},
			}},
		},

		// ---- arrayLiteralTypeAssertions: 'never' ----
		{
			Code:    `const x = [] as Foo;`,
			Options: arrNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedArrayTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceArrayTypeAssertionWithAnnotation", Output: `const x: Foo = [];`},
					{MessageId: "replaceArrayTypeAssertionWithSatisfies", Output: `const x = [] satisfies Foo;`},
				},
			}},
		},
		// Parenthesized array literal — `([ ... ]) as T`.
		{
			Code:    `const x = ([]) as Foo;`,
			Options: arrNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedArrayTypeAssertion",
				Line:      1, Column: 11, EndColumn: 22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceArrayTypeAssertionWithAnnotation", Output: `const x: Foo = ([]);`},
					{MessageId: "replaceArrayTypeAssertionWithSatisfies", Output: `const x = ([]) satisfies Foo;`},
				},
			}},
		},
		{
			Code:    `const x = [5] as any | string[];`,
			Options: arrNever,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unexpectedArrayTypeAssertion",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replaceArrayTypeAssertionWithAnnotation", Output: `const x: any | string[] = [5];`},
					{MessageId: "replaceArrayTypeAssertionWithSatisfies", Output: `const x = [5] satisfies any | string[];`},
				},
			}},
		},
	})
}
