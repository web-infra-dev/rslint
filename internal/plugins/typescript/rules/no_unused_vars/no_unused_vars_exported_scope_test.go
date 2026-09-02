package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsExportedDirectiveInternalBindings(t *testing.T) {
	script := rule.LanguageOptions{SourceType: "script"}
	unusedT := func(line, column int) rule_tester.InvalidTestCaseError {
		return rule_tester.InvalidTestCaseError{MessageId: "unusedVar", Line: line, Column: column}
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
			// A named default export has distinct local and export symbols. The
			// local binding still belongs to the script global scope and vars:local
			// must skip it before reportUsedIgnorePattern is considered.
			{
				Code:            `export default function _f() {} consume(_f);`,
				LanguageOptions: script,
				Options: map[string]interface{}{
					"vars":                    "local",
					"varsIgnorePattern":       "^_",
					"reportUsedIgnorePattern": true,
				},
			},
			{
				Code:            `/* exported T */ type T = string;`,
				FileName:        "exported.ts",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			// An ignored internal binding is unused, not a directive-created use.
			{
				Code:            `/* exported T */ type A<T> = string; consume(null as A<number>);`,
				LanguageOptions: script,
				Options: map[string]interface{}{
					"varsIgnorePattern":       "^T$",
					"reportUsedIgnorePattern": true,
				},
			},
			// These scope-manager bindings are intentionally ignored by the
			// TypeScript rule, independently of the directive.
			{Code: `/* exported T */ type A<U> = { [T in keyof U]: string }; consume(null as A<{ x: 1 }>);`, LanguageOptions: script},
			{Code: `/* exported T */ enum A { T } consume(A);`, LanguageOptions: script},
			{Code: `/* exported T */ type A = (T: number) => void; consume(null as A);`, LanguageOptions: script},
			// vars:local skips every binding hoisted into the script global scope,
			// including destructuring and var declarations nested in statements.
			{
				Code: `{ var blockVar = 1; }
var { top } = source;
for (var loopVar of source) { consume(); }`,
				LanguageOptions: script,
				Options:         map[string]interface{}{"vars": "local"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `/* exported T */
type A<T> = string;
consume(null as A<number>);`,
				LanguageOptions: script,
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(2, 8)},
			},
			{
				Code: `/* exported T */
interface A<T> { value: string }
consume(null as A<number>);`,
				LanguageOptions: script,
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(2, 13)},
			},
			{
				Code: `/* exported T */
class A<T> {}
consume(A);`,
				LanguageOptions: script,
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(2, 9)},
			},
			{
				Code: `/* exported T */
function f<T>() {}
f();`,
				LanguageOptions: script,
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(2, 12)},
			},
			{
				Code: `/* exported T */
type A = <T>() => void;
consume(null as A);`,
				LanguageOptions: script,
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(2, 11)},
			},
			{
				Code: `/* exported T */
interface A { <T>(): void }
consume(null as A);`,
				LanguageOptions: script,
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(2, 16)},
			},
			{
				Code: `/* exported T */
interface A { new <T>(): object }
consume(null as A);`,
				LanguageOptions: script,
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(2, 20)},
			},
			{
				Code: `/* exported T */
interface A { f<T>(): void }
consume(null as A);`,
				LanguageOptions: script,
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(2, 17)},
			},
			{
				Code: `/* exported T */
type A<U> = U extends infer T ? U : never;
consume(null as A<number>);`,
				LanguageOptions: script,
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(2, 29)},
			},
			// The global T is exported, but the shadowing type parameter is not.
			{
				Code: `/* exported T */
type T = number;
type A<T> = string;
consume(null as A<number>);`,
				LanguageOptions: script,
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(3, 8)},
			},
			// The export facet paired with the outer default-export local must not
			// match an internal type parameter with the same spelling.
			{
				Code:            `export default function T<T>() {} consume(T);`,
				LanguageOptions: script,
				Options:         map[string]interface{}{"vars": "local"},
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(1, 27)},
			},
			// Flat-config defaults make a .ts file a module without module syntax.
			{
				Code: `/* exported T */
type T = string;`,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 6)},
			},
			// A JavaScript CommonJS file has a wrapper scope.
			{
				Code:     "/* exported T */\nvar T;",
				FileName: "exported.js",
				TSConfig: "tsconfig.allow-js.json",
				LanguageOptions: rule.LanguageOptions{
					SourceType: "commonjs",
				},
				Errors: []rule_tester.InvalidTestCaseError{unusedT(2, 5)},
			},
			// vars:local skips the outer global binding, not its type parameter.
			{
				Code: `type A<T> = string;
consume(null as A<number>);`,
				LanguageOptions: script,
				Options:         map[string]interface{}{"vars": "local"},
				Errors:          []rule_tester.InvalidTestCaseError{unusedT(1, 8)},
			},
			// Latest typescript-eslint treats an infer binding as an ordinary
			// type-scope variable and reports it when the true branch never uses it.
			{
				Code: `type A<U> = U extends infer T ? U : never;
consume(null as A<number>);`,
				Errors: []rule_tester.InvalidTestCaseError{unusedT(1, 29)},
			},
		},
	)
}
