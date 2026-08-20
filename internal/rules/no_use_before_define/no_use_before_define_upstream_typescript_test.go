// TestNoUseBeforeDefineUpstreamTypescript migrates the TypeScript-parser half
// of the upstream valid/invalid suite from
// tests/lib/rules/no-use-before-define.js 1:1 — the `ruleTesterTypeScript`
// block covering the `enums`, `typedefs`, and `ignoreTypeReferences` options
// plus TS-only syntax. Position assertions cover line/column for every invalid
// case. The JavaScript half lives in no_use_before_define_upstream_test.go, and
// rslint-specific lock-in cases live in no_use_before_define_extras_test.go.
package no_use_before_define

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUseBeforeDefineUpstreamTypescript(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUseBeforeDefineRule,
		[]rule_tester.ValidTestCase{
			{Code: `
type foo = 1;
const x: foo = 1;
`},
			{Code: `
type foo = 1;
type bar = foo;
`},
			{Code: `
interface Foo {}
const x: Foo = {};
`},
			{Code: `
var a = 10;
alert(a);
`},
			{Code: `
function b(a) {
  alert(a);
}
`},
			{Code: `Object.hasOwnProperty.call(a);`},
			{Code: `
function a() {
  alert(arguments);
}
`},
			{Code: `declare function a();`},
			{Code: `
declare class a {
  foo();
}
`},
			{Code: `const updatedAt = data?.updatedAt;`},
			{Code: `
function f() {
  return function t() {};
}
f()?.();
`},
			{Code: `
var a = { b: 5 };
alert(a?.b);
`},
			{Code: `
a();
function a() {
  alert(arguments);
}
`, Options: "nofunc"},
			{Code: `
(() => {
  var a = 42;
  alert(a);
})();
`},
			{Code: `
a();
try {
  throw new Error();
} catch (a) {}
`},
			{Code: `
class A {}
new A();
`},
			{Code: `
var a = 0,
  b = a;
`},
			{Code: `var { a = 0, b = a } = {};`},
			{Code: `var [a = 0, b = a] = {};`},
			{Code: `
function foo() {
  foo();
}
`},
			{Code: `
var foo = function () {
  foo();
};
`},
			{Code: `
var a;
for (a in a) {
}
`},
			{Code: `
var a;
for (a of a) {
}
`},

			// ---- Block-level bindings ----
			{Code: `
'use strict';
a();
{
  function a() {}
}
`},
			{Code: `
'use strict';
{
  a();
  function a() {}
}
`, Options: "nofunc"},
			{Code: `
switch (foo) {
  case 1: {
    a();
  }
  default: {
    let a;
  }
}
`},
			{Code: `
a();
{
  let a = function () {};
}
`},

			// ---- object style options ----
			{Code: `
a();
function a() {
  alert(arguments);
}
`, Options: map[string]any{"functions": false}},
			{Code: `
'use strict';
{
  a();
  function a() {}
}
`, Options: map[string]any{"functions": false}},
			{Code: `
function foo() {
  new A();
}
class A {}
`, Options: map[string]any{"classes": false}},

			// ---- "variables" option ----
			{Code: `
function foo() {
  bar;
}
var bar;
`, Options: map[string]any{"variables": false}},
			{Code: `
var foo = () => bar;
var bar;
`, Options: map[string]any{"variables": false}},

			// ---- "typedefs" option ----
			{Code: `
var x: Foo = 2;
type Foo = string | number;
`, Options: map[string]any{"typedefs": false}},
			{Code: `
var x: Foo = {};
interface Foo {}
`, Options: map[string]any{"typedefs": false, "ignoreTypeReferences": false}},
			{Code: `
let myVar: String;
type String = string;
`, Options: map[string]any{"typedefs": false, "ignoreTypeReferences": false}},

			// ---- https://github.com/typescript-eslint/typescript-eslint/issues/2572 ----
			{Code: `
interface Bar {
  type: typeof Foo;
}

const Foo = 2;
`, Options: map[string]any{"ignoreTypeReferences": true}},
			{Code: `
interface Bar {
  type: typeof Foo.FOO;
}

class Foo {
  public static readonly FOO = '';
}
`, Options: map[string]any{"ignoreTypeReferences": true}},
			{Code: `
interface Bar {
  type: typeof Foo.Bar.Baz;
}

const Foo = {
  Bar: {
    Baz: 1,
  },
};
`, Options: map[string]any{"ignoreTypeReferences": true}},

			// ---- https://github.com/bradzacher/eslint-plugin-typescript/issues/141 ----
			{Code: `
interface ITest {
  first: boolean;
  second: string;
  third: boolean;
}

let first = () => console.log('first');

export let second = () => console.log('second');

export namespace Third {
  export let third = () => console.log('third');
}
`},

			// ---- https://github.com/eslint/typescript-eslint-parser/issues/550 ----
			{Code: `
function test(file: Blob) {
  const slice: typeof file.slice =
    file.slice || (file as any).webkitSlice || (file as any).mozSlice;
  return slice;
}
`},

			// ---- https://github.com/eslint/typescript-eslint-parser/issues/435 ----
			{Code: `
interface Foo {
  bar: string;
}
const bar = 'blah';
`},

			// ---- "enums" option ----
			{Code: `
function foo(): Foo {
  return Foo.FOO;
}

enum Foo {
  FOO,
}
`, Options: map[string]any{"enums": false}},
			{Code: `
let foo: Foo;

enum Foo {
  FOO,
}
`, Options: map[string]any{"enums": false}},
			{Code: `
class Test {
  foo(args: Foo): Foo {
    return Foo.FOO;
  }
}

enum Foo {
  FOO,
}
`, Options: map[string]any{"enums": false}},

			// ---- "allowNamedExports" option ----
			{Code: `
export { a };
const a = 1;
`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `
export { a as b };
const a = 1;
`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `
export { a, b };
let a, b;
`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `
export { a };
var a;
`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `
export { f };
function f() {}
`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `
export { C };
class C {}
`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `
export { Foo };

enum Foo {
  BAR,
}
`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `
export { Foo };

namespace Foo {
  export let bar = () => console.log('bar');
}
`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `
export { Foo, baz };

enum Foo {
  BAR,
}

let baz: Enum;
enum Enum {}
`, Options: map[string]any{"allowNamedExports": true}},

			// ---- https://github.com/typescript-eslint/typescript-eslint/issues/2502 ----
			{Code: `
import * as React from 'react';

<div />;
`, Tsx: true},
			{Code: `
import React from 'react';

<div />;
`, Tsx: true},
			// Upstream sets `parserOptions.jsxPragma: 'h'` here. rslint does not
			// model a JSX pragma, so `<div />` simply references nothing — which
			// is the same verdict.
			{Code: `
import { h } from 'preact';

<div />;
`, Tsx: true},
			{Code: `
const React = require('react');

<div />;
`, Tsx: true},

			// ---- https://github.com/typescript-eslint/typescript-eslint/issues/2527 ----
			{Code: `
type T = (value: unknown) => value is Id;
`},
			{Code: `
global.foo = true;

declare global {
  namespace NodeJS {
    interface Global {
      foo?: boolean;
    }
  }
}
`},

			// ---- https://github.com/typescript-eslint/typescript-eslint/issues/2824 ----
			{Code: `
@Directive({
  selector: '[rcCidrIpPattern]',
  providers: [
    {
      provide: NG_VALIDATORS,
      useExisting: CidrIpPatternDirective,
      multi: true,
    },
  ],
})
export class CidrIpPatternDirective implements Validator {}
`},
			{Code: `
@Directive({
  selector: '[rcCidrIpPattern]',
  providers: [
    {
      provide: NG_VALIDATORS,
      useExisting: CidrIpPatternDirective,
      multi: true,
    },
  ],
})
export class CidrIpPatternDirective implements Validator {}
`, Options: map[string]any{"classes": false}},

			// ---- https://github.com/typescript-eslint/typescript-eslint/issues/2941 ----
			{Code: `
class A {
  constructor(printName) {
    this.printName = printName;
  }

  openPort(printerName = this.printerName) {
    this.tscOcx.ActiveXopenport(printerName);

    return this;
  }
}
`},
			{Code: "\nconst obj = {\n  foo: 'foo-value',\n  bar: 'bar-value',\n} satisfies {\n  [key in 'foo' | 'bar']: `${key}-value`;\n};\n",
				Options: map[string]any{"ignoreTypeReferences": false}},
			{Code: "\nconst obj = {\n  foo: 'foo-value',\n  bar: 'bar-value',\n} as {\n  [key in 'foo' | 'bar']: `${key}-value`;\n};\n",
				Options: map[string]any{"ignoreTypeReferences": false}},
			{Code: `
const obj = {
  foo: {
    foo: 'foo',
  } as {
    [key in 'foo' | 'bar']: key;
  },
};
`, Options: map[string]any{"ignoreTypeReferences": false}},
			{Code: `
const foo = {
  bar: 'bar',
} satisfies {
  bar: typeof baz;
};

const baz = '';
`, Options: map[string]any{"ignoreTypeReferences": true}},
			{Code: `
namespace A.X.Y {}

import Z = A.X.Y;

const X = 23;
`},
			{Code: `
namespace A {
  export namespace X {
    export namespace Y {
      export const foo = 40;
    }
  }
}

import Z = A.X.Y;

const X = 23;
`},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream repeats this case for `sourceType: module`,
			// `ecmaVersion: 6`, and the tester default; rslint models none of
			// those as scope-affecting, so one case covers all three.
			{
				Code: `
a++;
var a = 19;
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Message: "'a' was used before it was defined.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2}},
			},
			{
				Code: `
a();
var a = function () {};
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 1}},
			},
			{
				Code: `
alert(a[1]);
var a = [1, 3];
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 7}},
			},
			{
				Code: `
a();
function a() {
  alert(b);
  var b = 10;
  a();
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 2, Column: 1},
					{MessageId: "usedBeforeDefined", Line: 4, Column: 9},
				},
			},
			{
				Code: `
a();
var a = function () {};
`,
				Options: "nofunc",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 1}},
			},
			{
				Code: `
(() => {
  alert(a);
  var a = 42;
})();
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 9}},
			},
			{
				Code: `
(() => a())();
function a() {}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 8}},
			},
			{
				Code: `
a();
try {
  throw new Error();
} catch (foo) {
  var a;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 1}},
			},
			{
				Code: `
var f = () => a;
var a;
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 15}},
			},
			{
				Code: `
new A();
class A {}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 5}},
			},
			{
				Code: `
function foo() {
  new A();
}
class A {}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 7}},
			},
			{
				Code: `
new A();
var A = class {};
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 5}},
			},
			{
				Code: `
function foo() {
  new A();
}
var A = class {};
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 7}},
			},

			// ---- Block-level bindings ----
			{
				Code: `
a++;
{
  var a;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 1}},
			},
			{
				Code: `
'use strict';
{
  a();
  function a() {}
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 4, Column: 3}},
			},
			{
				Code: `
{
  a;
  let a = 1;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 3}},
			},
			{
				Code: `
switch (foo) {
  case 1:
    a();
  default:
    let a;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 4, Column: 5}},
			},
			{
				Code: `
if (true) {
  function foo() {
    a;
  }
  let a;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 4, Column: 5}},
			},

			// ---- object style options ----
			{
				Code: `
a();
var a = function () {};
`,
				Options: map[string]any{"classes": false, "functions": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 1}},
			},
			{
				Code: `
new A();
var A = class {};
`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 5}},
			},
			{
				Code: `
function foo() {
  new A();
}
var A = class {};
`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 7}},
			},

			// ---- invalid initializers ----
			{
				Code:   `var a = a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 9}},
			},
			{
				Code:   `let a = a + b;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 9}},
			},
			{
				Code:   `const a = foo(a);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code:   `function foo(a = a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 18}},
			},
			{
				Code:   `var { a = a } = [];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 11}},
			},
			{
				Code:   `var [a = a] = [];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:   `var { b = a, a } = {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 11}},
			},
			{
				Code:   `var [b = a, a] = {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:   `var { a = 0 } = a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 17}},
			},
			{
				Code:   `var [a = 0] = a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code: `
for (var a in a) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 15}},
			},
			{
				Code: `
for (var a of a) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 15}},
			},

			// ---- "ignoreTypeReferences" option ----
			{
				Code: `
interface Bar {
  type: typeof Foo;
}

const Foo = 2;
`,
				Options: map[string]any{"ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 16}},
			},
			{
				Code: `
let var1: StringOrNumber;

type StringOrNumber = string | number;
`,
				Options: map[string]any{"ignoreTypeReferences": false, "typedefs": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 11}},
			},
			{
				Code: `
interface Bar {
  type: typeof Foo.FOO;
}

class Foo {
  public static readonly FOO = '';
}
`,
				Options: map[string]any{"ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 16}},
			},
			{
				Code: `
interface Bar {
  type: typeof Foo.Bar.Baz;
}

const Foo = {
  Bar: {
    Baz: 1,
  },
};
`,
				Options: map[string]any{"ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 16}},
			},
			{
				Code: `
const foo = {
  bar: 'bar',
} satisfies {
  bar: typeof baz;
};

const baz = '';
`,
				Options: map[string]any{"ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 5, Column: 15}},
			},

			// ---- "variables" option ----
			{
				Code: `
function foo() {
  bar;
  var bar = 1;
}
var bar;
`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 3}},
			},

			// ---- "enums" option ----
			{
				Code: `
class Test {
  foo(args: Foo): Foo {
    return Foo.FOO;
  }
}

enum Foo {
  FOO,
}
`,
				Options: map[string]any{"enums": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 4, Column: 12}},
			},
			{
				Code: `
function foo(): Foo {
  return Foo.FOO;
}

enum Foo {
  FOO,
}
`,
				Options: map[string]any{"enums": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 10}},
			},
			{
				Code: `
const foo = Foo.Foo;

enum Foo {
  FOO,
}
`,
				Options: map[string]any{"enums": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 13}},
			},

			// ---- "allowNamedExports" option ----
			{
				Code: `
export { a };
const a = 1;
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 10}},
			},
			{
				Code: `
export { a };
const a = 1;
`,
				Options: map[string]any{},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 10}},
			},
			{
				Code: `
export { a };
const a = 1;
`,
				Options: map[string]any{"allowNamedExports": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 10}},
			},
			{
				Code: `
export { a };
const a = 1;
`,
				Options: "nofunc",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 10}},
			},
			{
				Code: `
export { a as b };
const a = 1;
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 10}},
			},
			{
				Code: `
export { a, b };
let a, b;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 2, Column: 10},
					{MessageId: "usedBeforeDefined", Line: 2, Column: 13},
				},
			},
			{
				Code: `
export { a };
var a;
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 10}},
			},
			{
				Code: `
export { f };
function f() {}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 10}},
			},
			{
				Code: `
export { C };
class C {}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 10}},
			},
			{
				Code: `
export const foo = a;
const a = 1;
`,
				Options: map[string]any{"allowNamedExports": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 20}},
			},
			{
				Code: `
export function foo() {
  return a;
}
const a = 1;
`,
				Options: map[string]any{"allowNamedExports": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 10}},
			},
			{
				Code: `
export class C {
  foo() {
    return a;
  }
}
const a = 1;
`,
				Options: map[string]any{"allowNamedExports": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 4, Column: 12}},
			},
			{
				Code: `
export { Foo };

enum Foo {
  BAR,
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 10}},
			},
			{
				Code: `
export { Foo };

namespace Foo {
  export let bar = () => console.log('bar');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 10}},
			},
			{
				Code: `
export { Foo, baz };

enum Foo {
  BAR,
}

let baz: Enum;
enum Enum {}
`,
				Options: map[string]any{"allowNamedExports": false, "ignoreTypeReferences": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 2, Column: 10},
					{MessageId: "usedBeforeDefined", Line: 2, Column: 15},
				},
			},
			{
				Code: `
f();
function f() {}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 1}},
			},
			{
				Code: `
alert(a);
var a = 10;
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 7}},
			},
			{
				Code: `
f()?.();
function f() {
  return function t() {};
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 1}},
			},
			{
				Code: `
alert(a?.b);
var a = { b: 5 };
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 7}},
			},
			{
				Code: `
@decorator
class C {
  static x = 'foo';
  [C.x]() {}
}
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 5, Column: 4}},
			},
		},
	)
}
