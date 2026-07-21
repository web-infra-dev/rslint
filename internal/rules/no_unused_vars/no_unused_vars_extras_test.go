// TestNoUnusedVarsExtras locks in cases that are absent from ESLint's
// upstream suite. Scope-sensitive and discarded self-update cases live in
// no_unused_vars_extras_scopes_test.go.
package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			// Expression wrappers are usage sites, so nested parentheses and TS
			// assertions must preserve the underlying reference.
			{Code: `const wrapped = 1; ((wrapped as number)!);`},

			// Computed and shorthand keys are value references.
			{Code: `const computed = "key"; ({ [computed]: true });`},
			{Code: `const shorthand = 1; ({ shorthand });`},

			// Closure containers preserve captured references.
			{Code: `let captured; const read = () => captured; read();`},

			// String-form options skip only the global scope.
			{Code: `const globalOnly = 1;`, Options: []interface{}{"local"}},

			// JavaScript-compatible regexp syntax uses regexp2.
			{Code: `const _unused = 1;`, Options: map[string]interface{}{"varsIgnorePattern": `(?<=_)unused`}},

			// Locks in upstream collectUnusedVariables() global-scope branch:
			// `local` ignores a script global but not an ES module binding.
			{Code: `const scriptGlobal = 1;`, Options: map[string]interface{}{"vars": "local"}},
			{Code: `const { x, nested: [y] } = source;`, Options: map[string]interface{}{"vars": "local"}},
			{Code: `{ var blockVar = 1; } for (var loopVar of source) { consume(); }`, Options: map[string]interface{}{"vars": "local"}},

			// Exported directives mark declaration bindings only in the script-global
			// scope. Explicit `/* global */` names remain global even in a module.
			// Cover lexical declarations as well as var, which upstream exercises.
			{Code: `/*exported globalLet, GlobalClass, globalFn*/ let globalLet=1; class GlobalClass{} function globalFn(){}`},
			{Code: `/*exported blockVar*/ { var blockVar=1; }`},
			{
				Code: `/*global _declared*/ /*exported _declared*/ export {};`,
				Options: map[string]interface{}{
					"varsIgnorePattern":       "^_",
					"reportUsedIgnorePattern": true,
				},
			},

			// Locks in upstream array-pattern precedence: only an identifier that
			// is a direct ArrayPattern child matches this option.
			{Code: `const [_direct] = source;`, Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"}},
			{Code: `const [[_nestedDirect]] = source;`, Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"}},
			{Code: `let _assigned; [_assigned] = source;`, Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"}},
			{
				Code: `const [_default = 1] = source;`,
				Options: map[string]interface{}{
					"destructuredArrayIgnorePattern": "^unused$",
					"varsIgnorePattern":              "^_",
				},
			},

			// Locks in upstream isForInOfRef() direct-target arm. The equivalent
			// destructured targets are invalid cases below.
			{Code: `function f() { for (let item of source) { return true; } } f();`},
			{Code: `function f() { let item; for (item of source) { return true; } } f();`},

			// Locks in the final ignoreUsingDeclarations/hasRestSpreadSibling
			// gates for genuinely unused bindings.
			{Code: `using resource = getResource();`, Options: map[string]interface{}{"ignoreUsingDeclarations": true}},
			{Code: `const { ignored, ...rest } = source; consume(rest);`, Options: map[string]interface{}{"ignoreRestSiblings": true}},

			// A reference outside a class declaration consumes its outer binding.
			{Code: `class UsedClass { static make() { return new UsedClass(); } } consume(UsedClass);`},
		},
		[]rule_tester.InvalidTestCase{
			// A dot-property name is not a reference to the local binding.
			extraUnusedCase(
				`const local = 1; object.local;`,
				"local",
				true,
				1,
				7,
				12,
				` object.local;`,
			),

			// An inner parameter does not use a shadowed outer binding.
			extraUnusedCase(
				`const item = 1; function use(item) { return item; } use(1);`,
				"item",
				true,
				1,
				7,
				11,
				` function use(item) { return item; } use(1);`,
			),

			// String-form "local" still checks nested scopes.
			{
				Code:    "function outer() { const local = 1; }\nouter();",
				Options: []interface{}{"local"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("local", true, 1, 26, 31, "function outer() {  }\nouter();"),
				},
			},

			// Diagnostics remain in source order across declarations.
			{
				Code: "const later = 1;\nconst earlier = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("later", true, 1, 7, 12, "\nconst earlier = 2;"),
					extraUnusedError("earlier", true, 2, 7, 14, "const later = 1;\n"),
				},
			},

			// Real-user reports: report the last write, and keep comma-expression
			// updates unused when no later read consumes them.
			extraUnusedCase(
				"let x = [];\nx = x.concat(x);",
				"x",
				true,
				2,
				1,
				2,
				"",
			),
			extraUnusedCase(
				"let x = 0;\nx++, x = 0;\nx = 3;",
				"x",
				true,
				3,
				1,
				2,
				"",
			),

			// ---- Dimension 4: module containers remain local under vars:local ----
			{
				Code:    `export {}; const x = 1;`,
				Options: map[string]interface{}{"vars": "local"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("x", true, 1, 18, 19, `export {}; `),
				},
			},
			{
				Code:    `export {}; function x() {}`,
				Options: map[string]interface{}{"vars": "local"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("x", false, 1, 21, 22, `export {}; `),
				},
			},
			{
				Code:    `export {}; class X {}`,
				Options: map[string]interface{}{"vars": "local"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("X", false, 1, 18, 19, `export {}; `),
				},
			},
			{
				Code: `/*exported x*/ export {}; const x=1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("x", true, 1, 33, 34, `/*exported x*/ export {}; `),
				},
			},

			// ---- Dimension 4: vars:local distinguishes var and lexical loop/block scopes ----
			{
				Code:    `for(let x of source){}`,
				Options: map[string]interface{}{"vars": "local"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("x", true, 1, 9, 10, ""),
				},
			},
			{
				Code:    `{ let x=1; }`,
				Options: map[string]interface{}{"vars": "local"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("x", true, 1, 7, 8, `{  }`),
				},
			},

			// ---- Dimension 4: an export specifier never exports a nested shadow ----
			{
				Code: `const x=1; function f(){ const x=2; } f(); export {x};`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion(
						"x",
						true,
						1,
						32,
						33,
						`const x=1; function f(){  } f(); export {x};`,
					),
				},
			},
			{
				Code: `const x=1;for(let x of []){}export {x};`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("x", true, 1, 19, 20, ""),
				},
			},
			{
				Code: `const x=1;{const x=2;}export {x};`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("x", true, 1, 18, 19, `const x=1;{}export {x};`),
				},
			},
			{
				Code:    `const x=1;try{}catch(x){}export {x};`,
				Options: map[string]interface{}{"caughtErrors": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("x", false, 1, 22, 23, ""),
				},
			},

			// ---- Dimension 4: exported directives mark exact script-global names ----
			{
				Code: `/*exported x*/ var x=1; function f(){let x=2} f();`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("x", true, 1, 42, 43, `/*exported x*/ var x=1; function f(){} f();`),
				},
			},
			{
				Code: `/*exported x:false*/ var x=1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("x", true, 1, 26, 27, `/*exported x:false*/ `),
				},
			},
			{
				Code: `/* exported _x */ var _x=1;`,
				Options: map[string]interface{}{
					"varsIgnorePattern":       "^_",
					"reportUsedIgnorePattern": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUsedIgnoredError("_x", ". Used vars must not match /^_/u", 1, 23, 25),
				},
			},
			{
				Code: `/*exported f*/ { function f(){} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("f", false, 1, 27, 28, `/*exported f*/ {  }`),
				},
			},
			{
				Code: `/*exported x*/ for(let x of xs){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("x", true, 1, 24, 25, ""),
				},
			},

			// ---- Dimension 4: sibling lexical scopes never share references ----
			{
				Code: `{const x=1;} {const x=2;consume(x)}`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("x", true, 1, 8, 9, `{} {const x=2;consume(x)}`),
				},
			},

			// ---- Dimension 4: source ranges use ESLint's UTF-16 columns ----
			{
				Code: `"😀"; const 𝒳=1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("𝒳", true, 1, 13, 15, `"😀"; `),
				},
			},

			// ---- Dimension 4: class-local self references do not consume the declaration ----
			{
				Code: `class C{m(){return C}}`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("C", false, 1, 7, 8, ""),
				},
			},
			{
				Code: `class C extends C{}`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("C", false, 1, 7, 8, ""),
				},
			},

			// Locks in upstream direct ArrayPattern-parent checks: defaults and
			// rest elements are wrappers, so destructuredArrayIgnorePattern does
			// not apply to them.
			{
				Code:    `const [_x = 1] = arr;`,
				Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("_x", true, 1, 8, 10, ""),
				},
			},
			{
				Code:    `const [..._x] = arr;`,
				Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("_x", true, 1, 11, 13, ""),
				},
			},
			{
				Code:    `const [[_x = 1]] = arr;`,
				Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("_x", true, 1, 9, 11, ""),
				},
			},
			{
				Code:    `let _x; [..._x] = arr;`,
				Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("_x", true, 1, 13, 15, ""),
				},
			},
			{
				Code:    `const [x=1]=arr;`,
				Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("x", true, 1, 8, 9, ""),
				},
			},

			// Locks in upstream collectUnusedVariables() ordering: the array
			// pattern is considered before vars/args/caughtErrors categories.
			{
				Code: `const [_x] = arr; consume(_x);`,
				Options: map[string]interface{}{
					"destructuredArrayIgnorePattern": "^_",
					"varsIgnorePattern":              "^_",
					"reportUsedIgnorePattern":        true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUsedIgnoredError("_x", ". Used elements of array destructuring must not match /^_/u", 1, 8, 10),
				},
			},
			{
				Code: `function f([_x]){consume(_x)} f([]);`,
				Options: map[string]interface{}{
					"destructuredArrayIgnorePattern": "^_",
					"args":                           "none",
					"reportUsedIgnorePattern":        true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUsedIgnoredError("_x", ". Used elements of array destructuring must not match /^_/u", 1, 13, 15),
				},
			},
			{
				Code: `try{}catch([_x]){consume(_x)}`,
				Options: map[string]interface{}{
					"destructuredArrayIgnorePattern": "^_",
					"caughtErrors":                   "none",
					"reportUsedIgnorePattern":        true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUsedIgnoredError("_x", ". Used elements of array destructuring must not match /^_/u", 1, 13, 15),
				},
			},

			// Locks in upstream isForInOfRef(): the return-first exception does
			// not reach identifiers nested inside a destructuring target.
			{
				Code: `function f(){for(let [x] of obj){return true}} f()`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("x", true, 1, 23, 24, ""),
				},
			},
			{
				Code: `function f(){let x;for([x] of obj){return true}} f()`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("x", true, 1, 25, 26, ""),
				},
			},

			// Locks in upstream final suppression order: used bindings still
			// trigger reportUsedIgnorePattern before using/rest suppression.
			{
				Code: `const {_x,...rest}=obj; consume(_x,rest);`,
				Options: map[string]interface{}{
					"ignoreRestSiblings":      true,
					"varsIgnorePattern":       "^_",
					"reportUsedIgnorePattern": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUsedIgnoredError("_x", ". Used vars must not match /^_/u", 1, 8, 10),
				},
			},
			{
				Code: `using resource=getResource(); consume(resource);`,
				Options: map[string]interface{}{
					"ignoreUsingDeclarations": true,
					"varsIgnorePattern":       "^resource$",
					"reportUsedIgnorePattern": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUsedIgnoredError("resource", ". Used vars must not match /^resource$/u", 1, 7, 15),
				},
			},
			{
				Code: `await using resource=getResource(); consume(resource);`,
				Options: map[string]interface{}{
					"ignoreUsingDeclarations": true,
					"varsIgnorePattern":       "^resource$",
					"reportUsedIgnorePattern": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUsedIgnoredError("resource", ". Used vars must not match /^resource$/u", 1, 13, 21),
				},
			},
		},
	)
}

func TestNoUnusedVarsWithoutTypeChecker(t *testing.T) {
	t.Parallel()
	listeners := NoUnusedVarsRule.Run(rule.RuleContext{}, nil)
	if len(listeners) != 0 {
		t.Fatalf("expected graceful degradation without a type checker, got %d listeners", len(listeners))
	}
}

func extraUnusedCase(code string, name string, assigned bool, line int, column int, endColumn int, suggestionOutput string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			extraUnusedError(name, assigned, line, column, endColumn, suggestionOutput),
		},
	}
}

func extraUnusedError(name string, assigned bool, line int, column int, endColumn int, suggestionOutput string) rule_tester.InvalidTestCaseError {
	action := "defined"
	if assigned {
		action = "assigned a value"
	}
	result := rule_tester.InvalidTestCaseError{
		MessageId: "unusedVar",
		Message:   "'" + name + "' is " + action + " but never used.",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
	if suggestionOutput != "" {
		result.Suggestions = []rule_tester.InvalidTestCaseSuggestion{
			{
				MessageId: "removeVar",
				Output:    suggestionOutput,
			},
		}
	}
	return result
}

func extraUnusedErrorWithSuggestion(name string, assigned bool, line int, column int, endColumn int, suggestionOutput string) rule_tester.InvalidTestCaseError {
	result := extraUnusedError(name, assigned, line, column, endColumn, "")
	result.Suggestions = []rule_tester.InvalidTestCaseSuggestion{
		{
			MessageId: "removeVar",
			Output:    suggestionOutput,
		},
	}
	return result
}

func extraUsedIgnoredError(name string, additional string, line int, column int, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "usedIgnoredVar",
		Message:   "'" + name + "' is marked as ignored but is used" + additional + ".",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}
