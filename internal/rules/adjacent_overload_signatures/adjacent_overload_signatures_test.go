package adjacent_overload_signatures

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestAdjacentOverloadSignaturesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AdjacentOverloadSignaturesRule, []rule_tester.ValidTestCase{
		{Code: `
function error(a: string);
function error(b: number);
function error(ab: string | number) {}
export { error };
		`},
		{Code: `
import { connect } from 'react-redux';
export interface ErrorMessageModel {
  message: string;
}
function mapStateToProps() {}
function mapDispatchToProps() {}
export default connect(mapStateToProps, mapDispatchToProps)(ErrorMessage);
		`},
		{Code: `
export const foo = 'a',
  bar = 'b';
export interface Foo {}
export class Foo {}
		`},
		{Code: `
export interface Foo {}
export const foo = 'a',
  bar = 'b';
export class Foo {}
		`},
		{Code: `
const foo = 'a',
  bar = 'b';
interface Foo {}
class Foo {}
		`},
		{Code: `
interface Foo {}
const foo = 'a',
  bar = 'b';
class Foo {}
		`},
		{Code: `
export class Foo {}
export class Bar {}
export type FooBar = Foo | Bar;
		`},
		{Code: `
export interface Foo {}
export class Foo {}
export class Bar {}
export type FooBar = Foo | Bar;
		`},
		{Code: `
export function foo(s: string);
export function foo(n: number);
export function foo(sn: string | number) {}
export function bar(): void {}
export function baz(): void {}
		`},
		{Code: `
function foo(s: string);
function foo(n: number);
function foo(sn: string | number) {}
function bar(): void {}
function baz(): void {}
		`},
		{Code: `
declare function foo(s: string);
declare function foo(n: number);
declare function foo(sn: string | number);
declare function bar(): void;
declare function baz(): void;
		`},
		{Code: `
declare module 'Foo' {
  export function foo(s: string): void;
  export function foo(n: number): void;
  export function foo(sn: string | number): void;
  export function bar(): void;
  export function baz(): void;
}
		`},
		{Code: `
declare namespace Foo {
  export function foo(s: string): void;
  export function foo(n: number): void;
  export function foo(sn: string | number): void;
  export function bar(): void;
  export function baz(): void;
}
		`},
		{Code: `
type Foo = {
  foo(s: string): void;
  foo(n: number): void;
  foo(sn: string | number): void;
  bar(): void;
  baz(): void;
};
		`},
		{Code: `
type Foo = {
  foo(s: string): void;
  ['foo'](n: number): void;
  foo(sn: string | number): void;
  bar(): void;
  baz(): void;
};
		`},
		{Code: `
interface Foo {
  (s: string): void;
  (n: number): void;
  (sn: string | number): void;
  foo(n: number): void;
  bar(): void;
  baz(): void;
}
		`},
		{Code: `
interface Foo {
  (s: string): void;
  (n: number): void;
  (sn: string | number): void;
  foo(n: number): void;
  bar(): void;
  baz(): void;
  call(): void;
}
		`},
		{Code: `
interface Foo {
  foo(s: string): void;
  foo(n: number): void;
  foo(sn: string | number): void;
  bar(): void;
  baz(): void;
}
		`},
		{Code: `
interface Foo {
  foo(s: string): void;
  ['foo'](n: number): void;
  foo(sn: string | number): void;
  bar(): void;
  baz(): void;
}
		`},
		{Code: `
interface Foo {
  foo(): void;
  bar: {
    baz(s: string): void;
    baz(n: number): void;
    baz(sn: string | number): void;
  };
}
		`},
		{Code: `
interface Foo {
  new (s: string);
  new (n: number);
  new (sn: string | number);
  foo(): void;
}
		`},
		{Code: `
class Foo {
  constructor(s: string);
  constructor(n: number);
  constructor(sn: string | number) {}
  bar(): void {}
  baz(): void {}
}
		`},
		{Code: `
class Foo {
  foo(s: string): void;
  foo(n: number): void;
  foo(sn: string | number): void {}
  bar(): void {}
  baz(): void {}
}
		`},
		{Code: `
class Foo {
  foo(s: string): void;
  ['foo'](n: number): void;
  foo(sn: string | number): void {}
  bar(): void {}
  baz(): void {}
}
		`},
		{Code: `
class Foo {
  name: string;
  foo(s: string): void;
  foo(n: number): void;
  foo(sn: string | number): void {}
  bar(): void {}
  baz(): void {}
}
		`},
		{Code: `
class Foo {
  name: string;
  static foo(s: string): void;
  static foo(n: number): void;
  static foo(sn: string | number): void {}
  bar(): void {}
  baz(): void {}
}
		`},
		{Code: `
class Test {
  static test() {}
  untest() {}
  test() {}
}
		`},
		{Code: `export default function <T>(foo: T) {}`},
		{Code: `export default function named<T>(foo: T) {}`},
		{Code: `
interface Foo {
  [Symbol.toStringTag](): void;
  [Symbol.iterator](): void;
}
		`},
		{Code: `
class Test {
  #private(): void;
  #private(arg: number): void {}

  bar() {}

  '#private'(): void;
  '#private'(arg: number): void {}
}
		`},
		{Code: `
function wrap() {
  function foo(s: string);
  function foo(n: number);
  function foo(sn: string | number) {}
}
		`},
		{Code: `
if (true) {
  function foo(s: string);
  function foo(n: number);
  function foo(sn: string | number) {}
}
		`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
function wrap() {
  function foo(s: string);
  function foo(n: number);
  type bar = number;
  function foo(sn: string | number) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      6,
					Column:    3,
				},
			},
		},
		{
			Code: `
if (true) {
  function foo(s: string);
  function foo(n: number);
  let a = 1;
  function foo(sn: string | number) {}
  foo(a);
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      6,
					Column:    3,
				},
			},
		},
		{
			Code: `
export function foo(s: string);
export function foo(n: number);
export function bar(): void {}
export function baz(): void {}
export function foo(sn: string | number) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      6,
					Column:    1,
				},
			},
		},
		{
			Code: `
export function foo(s: string);
export function foo(n: number);
export type bar = number;
export type baz = number | string;
export function foo(sn: string | number) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      6,
					Column:    1,
				},
			},
		},
		{
			Code: `
function foo(s: string);
function foo(n: number);
function bar(): void {}
function baz(): void {}
function foo(sn: string | number) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      6,
					Column:    1,
				},
			},
		},
		{
			Code: `
function foo(s: string);
function foo(n: number);
type bar = number;
type baz = number | string;
function foo(sn: string | number) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      6,
					Column:    1,
				},
			},
		},
		{
			Code: `
function foo(s: string) {}
function foo(n: number) {}
const a = '';
const b = '';
function foo(sn: string | number) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      6,
					Column:    1,
				},
			},
		},
		{
			Code: `
function foo(s: string) {}
function foo(n: number) {}
class Bar {}
function foo(sn: string | number) {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      5,
					Column:    1,
				},
			},
		},
		{
			Code: `
function foo(s: string) {}
function foo(n: number) {}
function foo(sn: string | number) {}
class Bar {
  foo(s: string);
  foo(n: number);
  name: string;
  foo(sn: string | number) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      9,
					Column:    3,
				},
			},
		},
		{
			Code: `
declare function foo(s: string);
declare function foo(n: number);
declare function bar(): void;
declare function baz(): void;
declare function foo(sn: string | number);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      6,
					Column:    1,
				},
			},
		},
		{
			Code: `
declare function foo(s: string);
declare function foo(n: number);
const a = '';
const b = '';
declare function foo(sn: string | number);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      6,
					Column:    1,
				},
			},
		},
		{
			Code: `
declare module 'Foo' {
  export function foo(s: string): void;
  export function foo(n: number): void;
  export function bar(): void;
  export function baz(): void;
  export function foo(sn: string | number): void;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
declare module 'Foo' {
  export function foo(s: string): void;
  export function foo(n: number): void;
  export function foo(sn: string | number): void;
  function baz(s: string): void;
  export function bar(): void;
  function baz(n: number): void;
  function baz(sn: string | number): void;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      8,
					Column:    3,
				},
			},
		},
		{
			Code: `
declare namespace Foo {
  export function foo(s: string): void;
  export function foo(n: number): void;
  export function bar(): void;
  export function baz(): void;
  export function foo(sn: string | number): void;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
declare namespace Foo {
  export function foo(s: string): void;
  export function foo(n: number): void;
  export function foo(sn: string | number): void;
  function baz(s: string): void;
  export function bar(): void;
  function baz(n: number): void;
  function baz(sn: string | number): void;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      8,
					Column:    3,
				},
			},
		},
		{
			Code: `
type Foo = {
  foo(s: string): void;
  foo(n: number): void;
  bar(): void;
  baz(): void;
  foo(sn: string | number): void;
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
type Foo = {
  foo(s: string): void;
  ['foo'](n: number): void;
  bar(): void;
  baz(): void;
  foo(sn: string | number): void;
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
type Foo = {
  foo(s: string): void;
  name: string;
  foo(n: number): void;
  foo(sn: string | number): void;
  bar(): void;
  baz(): void;
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      5,
					Column:    3,
				},
			},
		},
		{
			Code: `
interface Foo {
  (s: string): void;
  foo(n: number): void;
  (n: number): void;
  (sn: string | number): void;
  bar(): void;
  baz(): void;
  call(): void;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      5,
					Column:    3,
				},
			},
		},
		{
			Code: `
interface Foo {
  foo(s: string): void;
  foo(n: number): void;
  bar(): void;
  baz(): void;
  foo(sn: string | number): void;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
interface Foo {
  foo(s: string): void;
  ['foo'](n: number): void;
  bar(): void;
  baz(): void;
  foo(sn: string | number): void;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
interface Foo {
  foo(s: string): void;
  'foo'(n: number): void;
  bar(): void;
  baz(): void;
  foo(sn: string | number): void;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
interface Foo {
  foo(s: string): void;
  name: string;
  foo(n: number): void;
  foo(sn: string | number): void;
  bar(): void;
  baz(): void;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      5,
					Column:    3,
				},
			},
		},
		{
			Code: `
interface Foo {
  foo(): void;
  bar: {
    baz(s: string): void;
    baz(n: number): void;
    foo(): void;
    baz(sn: string | number): void;
  };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      8,
					Column:    5,
				},
			},
		},
		{
			Code: `
interface Foo {
  new (s: string);
  new (n: number);
  foo(): void;
  bar(): void;
  new (sn: string | number);
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
interface Foo {
  new (s: string);
  foo(): void;
  new (n: number);
  bar(): void;
  new (sn: string | number);
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      5,
					Column:    3,
				},
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Foo {
  constructor(s: string);
  constructor(n: number);
  bar(): void {}
  baz(): void {}
  constructor(sn: string | number) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Foo {
  foo(s: string): void;
  foo(n: number): void;
  bar(): void {}
  baz(): void {}
  foo(sn: string | number): void {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Foo {
  foo(s: string): void;
  ['foo'](n: number): void;
  bar(): void {}
  baz(): void {}
  foo(sn: string | number): void {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Foo {
  // prettier-ignore
  "foo"(s: string): void;
  foo(n: number): void;
  bar(): void {}
  baz(): void {}
  foo(sn: string | number): void {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      8,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Foo {
  constructor(s: string);
  name: string;
  constructor(n: number);
  constructor(sn: string | number) {}
  bar(): void {}
  baz(): void {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      5,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Foo {
  foo(s: string): void;
  name: string;
  foo(n: number): void;
  foo(sn: string | number): void {}
  bar(): void {}
  baz(): void {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      5,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Foo {
  static foo(s: string): void;
  name: string;
  static foo(n: number): void;
  static foo(sn: string | number): void {}
  bar(): void {}
  baz(): void {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      5,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Test {
  #private(): void;
  '#private'(): void;
  #private(arg: number): void {}
  '#private'(arg: number): void {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "adjacentSignature",
					Line:      5,
					Column:    3,
				},
				{
					MessageId: "adjacentSignature",
					Line:      6,
					Column:    3,
				},
			},
		},
	})
}