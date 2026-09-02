package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsExportedInternalBindings(t *testing.T) {
	script := rule.LanguageOptions{SourceType: "script"}
	unusedT := func(line, column int, suggestionOutput string) rule_tester.InvalidTestCaseError {
		return extraUnusedErrorWithSuggestion("T", false, line, column, column+1, suggestionOutput)
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			// The effective source goal, rather than import/export syntax, selects
			// the global program scope.
			{
				Code: `/* exported T */
type T = string;
export {};`,
				LanguageOptions: script,
			},
			{
				Code:            `/* exported T */ export type T = string;`,
				LanguageOptions: script,
			},
			{
				Code:            `/* exported T */ type T = string;`,
				FileName:        "exported.ts",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			// An ignored internal binding is still unused. It must not become a
			// usedIgnoredVar merely because its spelling appears in the directive.
			{
				Code:            `/* exported T */ type A<T> = string; consume(null as A<number>);`,
				LanguageOptions: script,
				Options: map[string]interface{}{
					"varsIgnorePattern":       "^T$",
					"reportUsedIgnorePattern": true,
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `/* exported T */
type A<T> = string;
consume(null as A<number>);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 8, `/* exported T */
type A<> = string;
consume(null as A<number>);`)},
			},
			{
				Code: `/* exported T */
interface A<T> { value: string }
consume(null as A<number>);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 13, `/* exported T */
interface A<> { value: string }
consume(null as A<number>);`)},
			},
			{
				Code: `/* exported T */
class A<T> {}
consume(A);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 9, `/* exported T */
class A<> {}
consume(A);`)},
			},
			{
				Code: `/* exported T */
function f<T>() {}
f();`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 12, `/* exported T */
function f<>() {}
f();`)},
			},
			{
				Code: `/* exported T */
type A = <T>() => void;
consume(null as A);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 11, `/* exported T */
type A = <>() => void;
consume(null as A);`)},
			},
			{
				Code: `/* exported T */
interface A { <T>(): void }
consume(null as A);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 16, `/* exported T */
interface A { <>(): void }
consume(null as A);`)},
			},
			{
				Code: `/* exported T */
interface A { new <T>(): object }
consume(null as A);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 20, `/* exported T */
interface A { new <>(): object }
consume(null as A);`)},
			},
			{
				Code: `/* exported T */
interface A { f<T>(): void }
consume(null as A);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 17, `/* exported T */
interface A { f<>(): void }
consume(null as A);`)},
			},
			{
				Code: `/* exported T */
type A<U> = { [T in keyof U]: string };
consume(null as A<{ x: 1 }>);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 16, `/* exported T */
type A<U> = { [ in keyof U]: string };
consume(null as A<{ x: 1 }>);`)},
			},
			{
				Code: `/* exported T */
type A<U> = U extends infer T ? U : never;
consume(null as A<number>);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 29, `/* exported T */
type A<U> = U extends infer  ? U : never;
consume(null as A<number>);`)},
			},
			{
				Code: `/* exported T */
enum A { T }
consume(A);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 10, `/* exported T */
enum A {  }
consume(A);`)},
			},
			{
				Code: `/* exported T */
type A = (T: number) => void;
consume(null as A);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("T", false, 2, 11, 20, `/* exported T */
type A = () => void;
consume(null as A);`),
				},
			},
			// The global T is exported, but the shadowing type parameter is not.
			{
				Code: `/* exported T */
type T = number;
type A<T> = string;
consume(null as A<number>);`,
				LanguageOptions: script,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(3, 8, `/* exported T */
type T = number;
type A<> = string;
consume(null as A<number>);`)},
			},
			// Flat-config defaults make a .ts file a module even with no module syntax.
			{
				Code: `/* exported T */
type T = string;`,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 6, `/* exported T */
type  = string;`)},
			},
			// A JavaScript CommonJS file has a wrapper scope, not a global program scope.
			{
				Code:     "/* exported T */\nvar T;",
				FileName: "exported.cjs",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("T", false, 2, 5, 6, "/* exported T */\n"),
				},
			},
			// vars:local skips the outer global binding, not its internal type scope.
			{
				Code: `type A<T> = string;
consume(null as A<number>);`,
				LanguageOptions: script,
				Options:         map[string]interface{}{"vars": "local"},
				Errors: []rule_tester.InvalidTestCaseError{unusedT(1, 8, `type A<> = string;
consume(null as A<number>);`)},
			},
		},
	)
}
