// TestInitDeclarationsUpstream migrates the full valid/invalid suite from
// upstream typescript-eslint's
//
//	packages/eslint-plugin/tests/rules/init-declarations.test.ts
//
// 1:1, including the cases inherited from the core ESLint test suite that
// upstream embeds as "checking compatibility with base rule" entries.
//
// rslint-specific lock-in cases (Dimension 4 edge shapes, branch lock-ins,
// real-user issue shapes) live in init_declarations_extras_test.go.
package init_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// arrayOption builds the array-wrapped option shape that exercises
// utils.GetOptionsMap's JSON path — passing a Go struct directly would
// short-circuit it and leave the CLI-facing wiring untested.
func arrayOption(values ...interface{}) []interface{} {
	return values
}

func TestInitDeclarationsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&InitDeclarationsRule,
		[]rule_tester.ValidTestCase{
			// ---- checking compatibility with base rule ----
			{Code: `var foo = null;`},
			{Code: `foo = true;`},
			{Code: `
var foo = 1,
  bar = false,
  baz = {};
      `},
			{Code: `
function foo() {
  var foo = 0;
  var bar = [];
}
      `},
			{Code: `var fn = function () {};`},
			{Code: `var foo = (bar = 2);`},
			{Code: `for (var i = 0; i < 1; i++) {}`},
			{Code: `
for (var foo in []) {
}
      `},
			{Code: `
for (var foo of []) {
}
      `},
			{Code: `let a = true;`, Options: arrayOption("always")},
			{Code: `const a = {};`, Options: arrayOption("always")},
			{Code: `
function foo() {
  let a = 1,
    b = false;
  if (a) {
    let c = 3,
      d = null;
  }
}
      `, Options: arrayOption("always")},
			{Code: `
function foo() {
  const a = 1,
    b = true;
  if (a) {
    const c = 3,
      d = null;
  }
}
      `, Options: arrayOption("always")},
			{Code: `
function foo() {
  let a = 1;
  const b = false;
  var c = true;
}
      `, Options: arrayOption("always")},
			{Code: `var foo;`, Options: arrayOption("never")},
			{Code: `var foo, bar, baz;`, Options: arrayOption("never")},
			{Code: `
function foo() {
  var foo;
  var bar;
}
      `, Options: arrayOption("never")},
			{Code: `let a;`, Options: arrayOption("never")},
			{Code: `const a = 1;`, Options: arrayOption("never")},
			{Code: `
function foo() {
  let a, b;
  if (a) {
    let c, d;
  }
}
      `, Options: arrayOption("never")},
			{Code: `
function foo() {
  const a = 1,
    b = true;
  if (a) {
    const c = 3,
      d = null;
  }
}
      `, Options: arrayOption("never")},
			{Code: `
function foo() {
  let a;
  const b = false;
  var c;
}
      `, Options: arrayOption("never")},
			{Code: `for (var i = 0; i < 1; i++) {}`, Options: arrayOption("never", map[string]interface{}{"ignoreForLoopInit": true})},
			{Code: `
for (var foo in []) {
}
      `, Options: arrayOption("never", map[string]interface{}{"ignoreForLoopInit": true})},
			{Code: `
for (var foo of []) {
}
      `, Options: arrayOption("never", map[string]interface{}{"ignoreForLoopInit": true})},
			{Code: `
function foo() {
  var bar = 1;
  let baz = 2;
  const qux = 3;
}
      `, Options: arrayOption("always")},

			// ---- typescript-eslint ----
			{Code: `declare const foo: number;`, Options: arrayOption("always")},
			{Code: `declare const foo: number;`, Options: arrayOption("never")},
			{Code: `
declare namespace myLib {
  let numberOfGreetings: number;
}
      `, Options: arrayOption("always")},
			{Code: `
declare namespace myLib {
  let numberOfGreetings: number;
}
      `, Options: arrayOption("never")},
			{Code: `
interface GreetingSettings {
  greeting: string;
  duration?: number;
  color?: string;
}
      `},
			{Code: `
interface GreetingSettings {
  greeting: string;
  duration?: number;
  color?: string;
}
      `, Options: arrayOption("never")},
			{Code: `type GreetingLike = string | (() => string) | Greeter;`},
			{Code: `type GreetingLike = string | (() => string) | Greeter;`, Options: arrayOption("never")},
			{Code: `
function foo() {
  var bar: string;
}
      `, Options: arrayOption("never")},
			{Code: `var bar: string;`, Options: arrayOption("never")},
			{Code: `
var bar: string = function (): string {
  return 'string';
};
      `, Options: arrayOption("always")},
			{Code: `
var bar: string = function (arg1: string): string {
  return 'string';
};
      `, Options: arrayOption("always")},
			{Code: `function foo(arg1: string = 'string'): void {}`, Options: arrayOption("never")},
			{Code: `const foo: string = 'hello';`, Options: arrayOption("never")},
			{Code: `
const class1 = class NAME {
  constructor() {
    var name1: string = 'hello';
  }
};
      `},
			{Code: `
const class1 = class NAME {
  static pi: number = 3.14;
};
      `},
			{Code: `
const class1 = class NAME {
  static pi: number = 3.14;
};
      `, Options: arrayOption("never")},
			{Code: `
interface IEmployee {
  empCode: number;
  empName: string;
  getSalary: (number) => number; // arrow function
  getManagerName(number): string;
}
      `},
			{Code: `
interface IEmployee {
  empCode: number;
  empName: string;
  getSalary: (number) => number; // arrow function
  getManagerName(number): string;
}
      `, Options: arrayOption("never")},
			{Code: `const foo: number = 'asd';`, Options: arrayOption("always")},
			{Code: `const foo: number;`, Options: arrayOption("never")},
			{Code: `
namespace myLib {
  let numberOfGreetings: number;
}
      `, Options: arrayOption("never")},
			{Code: `
namespace myLib {
  let numberOfGreetings: number = 2;
}
      `, Options: arrayOption("always")},
			{Code: `
declare namespace myLib1 {
  const foo: number;
  namespace myLib2 {
    let bar: string;
    namespace myLib3 {
      let baz: object;
    }
  }
}
      `, Options: arrayOption("always")},
			{Code: `
declare namespace myLib1 {
  const foo: number;
  namespace myLib2 {
    let bar: string;
    namespace myLib3 {
      let baz: object;
    }
  }
}
      `, Options: arrayOption("never")},
		},
		[]rule_tester.InvalidTestCase{
			// ---- checking compatibility with base rule ----
			{
				Code:    `var foo;`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 8, Message: "Variable 'foo' should be initialized on declaration."},
				},
			},
			{
				Code:    `for (var a in []) var foo;`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 23, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: `
var foo,
  bar = false,
  baz;
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 2, Column: 5, EndLine: 2, EndColumn: 8},
					{MessageId: "initialized", Line: 4, Column: 3, EndLine: 4, EndColumn: 6},
				},
			},
			{
				Code: `
function foo() {
  var foo = 0;
  var bar;
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 4, Column: 7, EndLine: 4, EndColumn: 10},
				},
			},
			{
				Code: `
function foo() {
  var foo;
  var bar = foo;
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 3, Column: 7, EndLine: 3, EndColumn: 10},
				},
			},
			{
				Code:    `let a;`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
				},
			},
			{
				Code: `
function foo() {
  let a = 1,
    b;
  if (a) {
    let c = 3,
      d = null;
  }
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 4, Column: 5, EndLine: 4, EndColumn: 6},
				},
			},
			{
				Code: `
function foo() {
  let a;
  const b = false;
  var c;
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 3, Column: 7, EndLine: 3, EndColumn: 8},
					{MessageId: "initialized", Line: 5, Column: 7, EndLine: 5, EndColumn: 8},
				},
			},
			{
				Code:    `var foo = (bar = 2);`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 20, Message: "Variable 'foo' should not be initialized on declaration."},
				},
			},
			{
				Code:    `var foo = true;`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `
var foo,
  bar = 5,
  baz = 3;
      `,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 3, Column: 3, EndLine: 3, EndColumn: 10},
					{MessageId: "notInitialized", Line: 4, Column: 3, EndLine: 4, EndColumn: 10},
				},
			},
			{
				Code: `
function foo() {
  var foo;
  var bar = foo;
}
      `,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 4, Column: 7, EndLine: 4, EndColumn: 16},
				},
			},
			{
				Code:    `let a = 1;`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: `
function foo() {
  let a = 'foo',
    b;
  if (a) {
    let c, d;
  }
}
      `,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 3, Column: 7, EndLine: 3, EndColumn: 16},
				},
			},
			{
				Code: `
function foo() {
  let a;
  const b = false;
  var c = 1;
}
      `,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 5, Column: 7, EndLine: 5, EndColumn: 12},
				},
			},
			{
				Code:    `for (var i = 0; i < 1; i++) {}`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `
for (var foo in []) {
}
      `,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 2, Column: 10, EndLine: 2, EndColumn: 13},
				},
			},
			{
				Code: `
for (var foo of []) {
}
      `,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 2, Column: 10, EndLine: 2, EndColumn: 13},
				},
			},
			{
				Code: `
function foo() {
  var bar;
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 3, Column: 7, EndLine: 3, EndColumn: 10},
				},
			},

			// ---- typescript-eslint ----
			{
				Code:    `let arr: string[] = ['arr', 'ar'];`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    `let arr: string = function () {};`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code: `
const class1 = class NAME {
  constructor() {
    var name1: string = 'hello';
  }
};
      `,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 4, Column: 9, EndLine: 4, EndColumn: 32},
				},
			},
			{
				Code:    `let arr: string;`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code: `
namespace myLib {
  let numberOfGreetings: number;
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 3, Column: 7, EndLine: 3, EndColumn: 24},
				},
			},
			{
				Code: `
namespace myLib {
  let numberOfGreetings: number = 2;
}
      `,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 3, Column: 7, EndLine: 3, EndColumn: 36},
				},
			},
			{
				Code: `
namespace myLib1 {
  const foo: number;
  namespace myLib2 {
    let bar: string;
    namespace myLib3 {
      let baz: object;
    }
  }
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 3, Column: 9, EndLine: 3, EndColumn: 12},
					{MessageId: "initialized", Line: 5, Column: 9, EndLine: 5, EndColumn: 12},
					{MessageId: "initialized", Line: 7, Column: 11, EndLine: 7, EndColumn: 14},
				},
			},
		},
	)
}
