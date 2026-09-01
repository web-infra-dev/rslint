// TestNoInnerDeclarationsUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/no-inner-declarations.js at ESLint v10.8.1 1:1.
// Position assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in no_inner_declarations_extras_test.go.
package no_inner_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func noInnerLanguage(ecmaVersion int) rule.LanguageOptions {
	return rule.LanguageOptions{ECMAVersion: ecmaVersion, SourceType: "script"}
}

func noInnerValid(code string, options any, ecmaVersion int) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{
		Code:            code,
		Options:         options,
		LanguageOptions: noInnerLanguage(ecmaVersion),
	}
}

func noInnerError(message string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "moveDeclToRoot",
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func noInnerInvalid(
	code string,
	options any,
	ecmaVersion int,
	errors ...rule_tester.InvalidTestCaseError,
) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:            code,
		Options:         options,
		LanguageOptions: noInnerLanguage(ecmaVersion),
		Errors:          errors,
	}
}

func TestNoInnerDeclarationsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInnerDeclarationsRule,
		[]rule_tester.ValidTestCase{
			noInnerValid(`function doSomething() { }`, nil, 5),
			noInnerValid(`function doSomething() { function somethingElse() { } }`, nil, 5),
			noInnerValid(`(function() { function doSomething() { } }());`, nil, 5),
			noInnerValid(`if (test) { var fn = function() { }; }`, nil, 5),
			noInnerValid(`if (test) { var fn = function expr() { }; }`, nil, 5),
			noInnerValid(`function decl() { var fn = function expr() { }; }`, nil, 5),
			noInnerValid(`function decl(arg) { var fn; if (arg) { fn = function() { }; } }`, nil, 5),
			noInnerValid(`var x = {doSomething() {function doSomethingElse() {}}}`, nil, 2015),
			noInnerValid(`function decl(arg) { var fn; if (arg) { fn = function expr() { }; } }`, nil, 2015),
			noInnerValid(`function decl(arg) { var fn; if (arg) { fn = function expr() { }; } }`, nil, 5),
			noInnerValid(`if (test) { var foo; }`, nil, 5),
			noInnerValid(`if (test) { let x = 1; }`, []any{"both"}, 2015),
			noInnerValid(`if (test) { const x = 1; }`, []any{"both"}, 2015),
			noInnerValid(`if (test) { using x = 1; }`, []any{"both"}, 2026),
			noInnerValid(`if (test) { await using x = 1; }`, []any{"both"}, 2026),
			noInnerValid(`function doSomething() { while (test) { var foo; } }`, nil, 5),
			noInnerValid(`var foo;`, []any{"both"}, 5),
			noInnerValid(`var foo = 42;`, []any{"both"}, 5),
			noInnerValid(`function doSomething() { var foo; }`, []any{"both"}, 5),
			noInnerValid(`(function() { var foo; }());`, []any{"both"}, 5),
			noInnerValid(`foo(() => { function bar() { } });`, nil, 2015),
			noInnerValid(`var fn = () => {var foo;}`, []any{"both"}, 2015),
			noInnerValid(`var x = {doSomething() {var foo;}}`, []any{"both"}, 2015),
			noInnerValid(`export var foo;`, []any{"both"}, 2015),
			noInnerValid(`export function bar() {}`, []any{"both"}, 2015),
			noInnerValid(`export default function baz() {}`, []any{"both"}, 2015),
			noInnerValid(`exports.foo = () => {}`, []any{"both"}, 2015),
			noInnerValid(`exports.foo = function(){}`, []any{"both"}, 5),
			noInnerValid(`module.exports = function foo(){}`, []any{"both"}, 5),
			noInnerValid(`class C { method() { function foo() {} } }`, []any{"both"}, 2022),
			noInnerValid(`class C { method() { var x; } }`, []any{"both"}, 2022),
			noInnerValid(`class C { static { function foo() {} } }`, []any{"both"}, 2022),
			noInnerValid(`class C { static { var x; } }`, []any{"both"}, 2022),
			noInnerValid(`'use strict'
 if (test) { function doSomething() { } }`, []any{"functions", map[string]any{"blockScopedFunctions": "allow"}}, 2022),
			noInnerValid(`'use strict'
 if (test) { function doSomething() { } }`, []any{"functions"}, 2022),
			noInnerValid(`function foo() {'use strict'
 if (test) { function doSomething() { } } }`, []any{"functions", map[string]any{"blockScopedFunctions": "allow"}}, 2015),
			{
				Code: `function foo() { { function bar() { } } }`,
				Options: []any{
					"functions",
					map[string]any{"blockScopedFunctions": "allow"},
				},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022, SourceType: "module"},
			},
			noInnerValid(`class C { method() { if(test) { function somethingElse() { } } } }`, []any{"functions", map[string]any{"blockScopedFunctions": "allow"}}, 2022),
			noInnerValid(`const C = class { method() { if(test) { function somethingElse() { } } } }`, []any{"functions", map[string]any{"blockScopedFunctions": "allow"}}, 2022),
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:            `export {}; function foo() { { function bar() { } } }`,
				Options:         []any{"functions", map[string]any{"blockScopedFunctions": "allow"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022, SourceType: "script"},
				Errors:          []rule_tester.InvalidTestCaseError{noInnerError("Move function declaration to function body root.", 1, 31, 1, 49)},
			},
			noInnerInvalid(
				`if (test) { function doSomething() { } }`,
				[]any{"both"},
				5,
				noInnerError("Move function declaration to program root.", 1, 13, 1, 39),
			),
			noInnerInvalid(
				`if (foo) var a; `,
				[]any{"both"},
				5,
				noInnerError("Move variable declaration to program root.", 1, 10, 1, 16),
			),
			noInnerInvalid(
				`if (foo) /* some comments */ var a; `,
				[]any{"both"},
				5,
				noInnerError("Move variable declaration to program root.", 1, 30, 1, 36),
			),
			noInnerInvalid(
				`if (foo){ function f(){ if(bar){ var a; } } }`,
				[]any{"both"},
				5,
				noInnerError("Move function declaration to program root.", 1, 11, 1, 44),
				noInnerError("Move variable declaration to function body root.", 1, 34, 1, 40),
			),
			noInnerInvalid(
				`if (foo) function f(){ if(bar) var a; }`,
				[]any{"both"},
				5,
				noInnerError("Move function declaration to program root.", 1, 10, 1, 40),
				noInnerError("Move variable declaration to function body root.", 1, 32, 1, 38),
			),
			noInnerInvalid(
				`if (foo) { var fn = function(){} } `,
				[]any{"both"},
				5,
				noInnerError("Move variable declaration to program root.", 1, 12, 1, 33),
			),
			noInnerInvalid(
				`if (foo)  function f(){} `,
				nil,
				5,
				noInnerError("Move function declaration to program root.", 1, 11, 1, 25),
			),
			noInnerInvalid(
				`function bar() { if (foo) function f(){}; }`,
				[]any{"both"},
				5,
				noInnerError("Move function declaration to function body root.", 1, 27, 1, 41),
			),
			noInnerInvalid(
				`function bar() { if (foo) var a; }`,
				[]any{"both"},
				5,
				noInnerError("Move variable declaration to function body root.", 1, 27, 1, 33),
			),
			noInnerInvalid(
				`if (foo) { var a; }`,
				[]any{"both"},
				5,
				noInnerError("Move variable declaration to program root.", 1, 12, 1, 18),
			),
			noInnerInvalid(
				`function doSomething() { do { function somethingElse() { } } while (test); }`,
				nil,
				5,
				noInnerError("Move function declaration to function body root.", 1, 31, 1, 59),
			),
			noInnerInvalid(
				`(function() { if (test) { function doSomething() { } } }());`,
				nil,
				5,
				noInnerError("Move function declaration to function body root.", 1, 27, 1, 53),
			),
			noInnerInvalid(
				`while (test) { var foo; }`,
				[]any{"both"},
				5,
				noInnerError("Move variable declaration to program root.", 1, 16, 1, 24),
			),
			noInnerInvalid(
				`function doSomething() { if (test) { var foo = 42; } }`,
				[]any{"both"},
				5,
				noInnerError("Move variable declaration to function body root.", 1, 38, 1, 51),
			),
			noInnerInvalid(
				`(function() { if (test) { var foo; } }());`,
				[]any{"both"},
				5,
				noInnerError("Move variable declaration to function body root.", 1, 27, 1, 35),
			),
			noInnerInvalid(
				`const doSomething = () => { if (test) { var foo = 42; } }`,
				[]any{"both"},
				2015,
				noInnerError("Move variable declaration to function body root.", 1, 41, 1, 54),
			),
			noInnerInvalid(
				`class C { method() { if(test) { var foo; } } }`,
				[]any{"both"},
				2015,
				noInnerError("Move variable declaration to function body root.", 1, 33, 1, 41),
			),
			noInnerInvalid(
				`class C { static { if (test) { var foo; } } }`,
				[]any{"both"},
				2022,
				noInnerError("Move variable declaration to class static block body root.", 1, 32, 1, 40),
			),
			noInnerInvalid(
				`class C { static { if (test) { function foo() {} } } }`,
				[]any{"both", map[string]any{"blockScopedFunctions": "disallow"}},
				2022,
				noInnerError("Move function declaration to class static block body root.", 1, 32, 1, 49),
			),
			noInnerInvalid(
				`class C { static { if (test) { if (anotherTest) { var foo; } } } }`,
				[]any{"both"},
				2022,
				noInnerError("Move variable declaration to class static block body root.", 1, 51, 1, 59),
			),
			noInnerInvalid(
				`if (test) { function doSomething() { } }`,
				[]any{"both", map[string]any{"blockScopedFunctions": "allow"}},
				5,
				noInnerError("Move function declaration to program root.", 1, 13, 1, 39),
			),
			noInnerInvalid(
				`if (test) { function doSomething() { } }`,
				[]any{"both", map[string]any{"blockScopedFunctions": "disallow"}},
				2022,
				noInnerError("Move function declaration to program root.", 1, 13, 1, 39),
			),
			noInnerInvalid(
				`'use strict'
 if (test) { function doSomething() { } }`,
				[]any{"both", map[string]any{"blockScopedFunctions": "disallow"}},
				2022,
				noInnerError("Move function declaration to program root.", 2, 14, 2, 40),
			),
			noInnerInvalid(
				`'use strict'
 if (test) { function doSomething() { } }`,
				[]any{"both", map[string]any{"blockScopedFunctions": "disallow"}},
				5,
				noInnerError("Move function declaration to program root.", 2, 14, 2, 40),
			),
			noInnerInvalid(
				`'use strict'
 if (test) { function doSomething() { } }`,
				[]any{"both", map[string]any{"blockScopedFunctions": "allow"}},
				5,
				noInnerError("Move function declaration to program root.", 2, 14, 2, 40),
			),
			noInnerInvalid(
				`function foo() {'use strict'
 { function bar() { } } }`,
				[]any{"both", map[string]any{"blockScopedFunctions": "disallow"}},
				2022,
				noInnerError("Move function declaration to function body root.", 2, 4, 2, 22),
			),
			noInnerInvalid(
				`function foo() {'use strict'
 { function bar() { } } }`,
				[]any{"both", map[string]any{"blockScopedFunctions": "disallow"}},
				5,
				noInnerError("Move function declaration to function body root.", 2, 4, 2, 22),
			),
			noInnerInvalid(
				`function doSomething() { 'use strict'
 do { function somethingElse() { } } while (test); }`,
				[]any{"both", map[string]any{"blockScopedFunctions": "disallow"}},
				5,
				noInnerError("Move function declaration to function body root.", 2, 7, 2, 35),
			),
			noInnerInvalid(
				`{ function foo () {'use strict'
 console.log('foo called'); } }`,
				[]any{"both"},
				2022,
				noInnerError("Move function declaration to program root.", 1, 3, 2, 30),
			),
		},
	)
}
