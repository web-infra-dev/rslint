package no_duplicate_set_values_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_duplicate_set_values"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDuplicateSetValuesExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_duplicate_set_values.NoDuplicateSetValuesRule,
		[]rule_tester.ValidTestCase{
			// Only a single, non-spread array literal argument is checked.
			valid(`new Set`),
			valid(`new Set()`),
			valid(`new Set([1, 1], extra)`),
			valid(`new Set(...[[1, 1]])`),
			valid(`new foo.Set([1, 1])`),

			// Distinct values and expressions without a modeled identity.
			valid(`new Set([1n, 1])`),
			valid(`new Set([true, 1])`),
			valid(`new Set([null, undefined])`),
			valid(`new Set([0, "0"])`),
			valid("new Set([`${foo}`, `${foo}`])"),
			valid(`new Set([foo?.(), foo?.()])`),
			valid(`new Set([foo[bar()], foo[bar()]])`),
			valid(`new Set([foo.bar, foo["baz"]])`),
			valid(`new Set([new Set([1]), new Set([1])])`),
			valid(`new Set([() => 1, () => 1])`),
			valid("class A { #a; #b; m() { return new Set([this.#a, this.#b]); } }"),

			// Reassigned bindings have no static value.
			valid("let foo = 1;\nfoo = 2;\nnew Set([foo, 1]);"),
		},
		[]rule_tester.InvalidTestCase{
			// ESTree drops parentheses around elements, arguments and the callee.
			invalid(`new Set([(1), ((1))])`, duplicate(1, 17, 18, "1")),
			invalid(`new Set(([1, 1]))`, duplicate(1, 14, 15, "1")),
			invalid(`new (Set)([1, 1])`, duplicate(1, 15, 16, "1")),
			invalid(`new Set([1, /* c */ 1])`, duplicate(1, 21, 22, "1")),
			// JSDoc casts are not ESTree nodes either.
			invalid(`new Set([/** @type {number} */ (1), 1])`, duplicate(1, 37, 38, "1")),
			invalid(`new Set([1, /** @type {number} */ (1)])`, duplicate(1, 36, 37, "1")),

			// TypeScript wrappers are transparent for values and references.
			invalidTS(`new Set<number>([1, 1])`, duplicate(1, 21, 22, "1")),
			invalidTS(`new Set([1, 1] as number[])`, duplicate(1, 13, 14, "1")),
			invalidTS(`new Set(<number[]>[1, 1])`, duplicate(1, 23, 24, "1")),
			invalidTS(`new Set(([1, 1] as const))`, duplicate(1, 14, 15, "1")),
			// A duplicate hole is reported on the wrapped argument.
			invalidTS(`new Set([,,] as const)`, duplicate(1, 9, 22, "undefined")),
			invalidTS(`new Set([1 as number, 1])`, duplicate(1, 23, 24, "1")),
			invalidTS(`new Set([foo!, foo])`, duplicate(1, 16, 19, "foo")),
			invalidTS(`new Set([foo satisfies number, <number>foo])`, duplicate(1, 32, 43, "<number>foo")),
			// Upstream's reference comparison unwraps the assertions and then
			// compares regular-expression literals by pattern and flags.
			invalidTS(`new Set([/a/ as RegExp, /a/ as RegExp])`, duplicate(1, 25, 38, "/a/ as RegExp")),

			// References: optional chains and computed keys.
			invalid(`new Set([foo?.bar, foo.bar])`, duplicate(1, 20, 27, "foo.bar")),
			invalid(`new Set([foo[0], foo["0"]])`, duplicate(1, 18, 26, `foo["0"]`)),
			invalid(`new Set([foo[bar], foo[bar]])`, duplicate(1, 20, 28, "foo[bar]")),
			invalid("class A { #a; m() { return new Set([this.#a, this.#a]); } }", duplicate(1, 46, 53, "this.#a")),

			// Static values use SameValueZero.
			invalid("new Set([`a${\"b\"}`, \"ab\"])", duplicate(1, 21, 25, `"ab"`)),
			invalid(`new Set([1n, 1n])`, duplicate(1, 14, 16, "1n")),
			invalid(`new Set([0x10, 16])`, duplicate(1, 16, 18, "16")),
			invalid(`new Set([NaN, 0 / 0])`, duplicate(1, 15, 20, "0 / 0")),
			invalid(`new Set([void 0, undefined])`, duplicate(1, 18, 27, "undefined")),
			invalid(`new Set([Symbol.iterator, Symbol.iterator])`, duplicate(1, 27, 42, "Symbol.iterator")),
			invalid("const a = 1; const b = a; new Set([a, b]);", duplicate(1, 39, 40, "b")),

			// Every later repetition is reported.
			invalid(`new Set([1, 1, 1])`, duplicate(1, 13, 14, "1"), duplicate(1, 16, 17, "1")),
			invalid(`new Set([, , undefined])`, duplicate(1, 9, 24, "undefined"), duplicate(1, 14, 23, "undefined")),

			// Nested sets are checked independently.
			invalid(`new Set([new Set([1, 1])])`, duplicate(1, 22, 23, "1")),

			// Like upstream, a shadowed `Set` is still checked.
			invalid(`function f(Set) { return new Set([1, 1]); }`, duplicate(1, 38, 39, "1")),
		},
	)
}

// TestNoDuplicateSetValuesInstantiationExpressions covers TypeScript
// instantiation expressions, which upstream unwraps like other TypeScript
// wrappers around the argument, static values and references.
func TestNoDuplicateSetValuesInstantiationExpressions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_duplicate_set_values.NoDuplicateSetValuesRule,
		[]rule_tester.ValidTestCase{
			validTS(`new Set([foo<A>, bar<A>]);`),
		},
		[]rule_tester.InvalidTestCase{
			invalidTS(`new Set([foo<string>, foo]);`, duplicate(1, 23, 26, "foo")),
			invalidTS(`new Set([foo.bar<T>, foo.bar]);`, duplicate(1, 22, 29, "foo.bar")),
			invalidTS(`new Set([foo<A>, foo<B>]);`, duplicate(1, 18, 24, "foo<B>")),
			invalidTS(`new Set([foo?.bar<T>, foo.bar]);`, duplicate(1, 23, 30, "foo.bar")),
			invalidTS(`const a = 1; new Set([a<T>, 1]);`, duplicate(1, 29, 30, "1")),
			invalidTS(`new Set([1, 1]<string>);`, duplicate(1, 13, 14, "1")),
			invalidTS(`new Set([,,]<x>);`, duplicate(1, 9, 16, "undefined")),
		},
	)
}

// TestNoDuplicateSetValuesLargeLiteral covers the indexed comparison path: a
// generated allowlist of distinct static values reports only the real
// duplicates, including a static value that repeats an unresolved reference.
func TestNoDuplicateSetValuesLargeLiteral(t *testing.T) {
	var elements []string
	for i := range 10_000 {
		elements = append(elements, strconv.Itoa(i))
	}
	elements = append(elements, "foo", "1e3", "foo")
	code := "new Set([" + strings.Join(elements, ", ") + "]);"
	column := func(element string) int { return strings.LastIndex(code, element) + 1 }

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_duplicate_set_values.NoDuplicateSetValuesRule,
		nil,
		[]rule_tester.InvalidTestCase{
			invalid(code,
				duplicate(1, column("1e3"), column("1e3")+len("1e3"), "1e3"),
				duplicate(1, column("foo"), column("foo")+len("foo"), "foo"),
			),
			// A static value never matches an unresolved reference by identity.
			invalid("const foo = 1;\nnew Set([bar, foo, bar, 1]);",
				duplicate(2, 20, 23, "bar"),
				duplicate(2, 25, 26, "1"),
			),
		},
	)
}
