import { RuleTester } from '@rslint/rule-tester';
import { config } from '../../config';

// Migrated from the existing suite. Each active case pins the full diagnostic and edits.
// Preserve the existing Go binding ranges and enum-shadow diagnostic while strengthening the assertions.
new RuleTester(config).run('no-shadow', {
  valid: [
    {
      code: "function fn() { type Record = string; }",
      options: [
        {
          builtinGlobals: true,
        },
      ],
    },
    {
      code: "function fn(Record: number) {}",
      options: [
        {
          builtinGlobals: false,
        },
      ],
    },
    {
      code: "function fn(NodeListOf: number, HTMLElement: number) {}",
      options: [
        {
          builtinGlobals: true,
        },
      ],
    },
    {
      code: "function foo<T = (arg: any) => any>(arg: T) {}",
    },
    {
      code: "function foo<T = ([arg]: [any]) => any>(arg: T) {}",
    },
    {
      code: "function foo<T = ({ args }: { args: any }) => any>(arg: T) {}",
    },
    {
      code: "function foo<T = (...args: any[]) => any>(fn: T, args: any[]) {}",
    },
    {
      code: "function foo<T extends (...args: any[]) => any>(fn: T, args: any[]) {}",
    },
    {
      code: "function foo<T extends (...args: any[]) => any>(fn: T, ...args: any[]) {}",
    },
    {
      code: "function foo<T extends ([args]: any[]) => any>(fn: T, args: any[]) {}",
    },
    {
      code: "function foo<T extends ([...args]: any[]) => any>(fn: T, args: any[]) {}",
    },
    {
      code: "function foo<T extends ({ args }: { args: any }) => any>(fn: T, args: any) {}",
    },
    {
      code: "\nfunction foo<T extends (id: string, ...args: any[]) => any>(\n  fn: T,\n  ...args: any[]\n) {}\n    ",
    },
    {
      code: "\ntype Args = 1;\nfunction foo<T extends (Args: any) => void>(arg: T) {}\n    ",
    },
    {
      code: "\nexport type ArrayInput<Func> = Func extends (arg0: Array<infer T>) => any\n  ? T[]\n  : Func extends (...args: infer T) => any\n    ? T\n    : never;\n    ",
    },
    {
      code: "\nfunction foo() {\n  var Object = 0;\n}\n    ",
    },
    {
      code: "\nfunction test(this: Foo) {\n  function test2(this: Bar) {}\n}\n    ",
    },
    {
      code: "\nclass Foo {\n  prop = 1;\n}\nnamespace Foo {\n  export const v = 2;\n}\n    ",
    },
    {
      code: "\nfunction Foo() {}\nnamespace Foo {\n  export const v = 2;\n}\n    ",
    },
    {
      code: "\nclass Foo {\n  prop = 1;\n}\ninterface Foo {\n  prop2: string;\n}\n    ",
    },
    {
      code: "\nimport type { Foo } from 'bar';\n\ndeclare module 'bar' {\n  export interface Foo {\n    x: string;\n  }\n}\n    ",
    },
    {
      code: "\nconst x = 1;\ntype x = string;\n    ",
    },
    {
      code: "\nconst x = 1;\n{\n  type x = string;\n}\n    ",
    },
    {
      code: "\ntype Foo = 1;\n      ",
      options: [
        {
          ignoreTypeValueShadow: true,
        },
      ],
      languageOptions: {
        globals: {
          Foo: "writable",
        },
      },
    },
    {
      code: "\ntype Foo = 1;\n      ",
      options: [
        {
          builtinGlobals: false,
          ignoreTypeValueShadow: false,
        },
      ],
      languageOptions: {
        globals: {
          Foo: "writable",
        },
      },
    },
    {
      code: "\nenum Direction {\n  left = 'left',\n  right = 'right',\n}\n    ",
    },
    {
      code: "\nconst test = 1;\ntype Fn = (test: string) => typeof test;\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: "\ntype Fn = (Foo: string) => typeof Foo;\n      ",
      options: [
        {
          builtinGlobals: false,
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
      languageOptions: {
        globals: {
          Foo: "writable",
        },
      },
    },
    {
      code: "\nconst arg = 0;\n\ninterface Test {\n  (arg: string): typeof arg;\n}\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ninterface Test {\n  p1(arg: string): typeof arg;\n}\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ndeclare function test(arg: string): typeof arg;\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ndeclare const test: (arg: string) => typeof arg;\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ndeclare class Test {\n  p1(arg: string): typeof arg;\n}\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ndeclare const Test: {\n  new (arg: string): typeof arg;\n};\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ntype Bar = new (arg: number) => typeof arg;\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ndeclare namespace Lib {\n  function test(arg: string): typeof arg;\n}\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: "\n        declare global {\n          interface ArrayConstructor {}\n        }\n        export {};\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
    },
    {
      code: "\n      declare global {\n        const a: string;\n\n        namespace Foo {\n          const a: number;\n        }\n      }\n      export {};\n    ",
    },
    {
      code: "\n        declare global {\n          type A = 'foo';\n\n          namespace Foo {\n            type A = 'bar';\n          }\n        }\n        export {};\n      ",
      options: [
        {
          ignoreTypeValueShadow: false,
        },
      ],
    },
    {
      code: "\n        declare global {\n          const foo: string;\n          type Fn = (foo: number) => void;\n        }\n        export {};\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
    },
    {
      code: "\nexport class Wrapper<Wrapped> {\n  private constructor(private readonly wrapped: Wrapped) {}\n\n  unwrap(): Wrapped {\n    return this.wrapped;\n  }\n\n  static create<Wrapped>(wrapped: Wrapped) {\n    return new Wrapper<Wrapped>(wrapped);\n  }\n}\n    ",
    },
    {
      code: "\nfunction makeA() {\n  return class A<T> {\n    constructor(public value: T) {}\n\n    static make<T>(value: T) {\n      return new A<T>(value);\n    }\n  };\n}\n    ",
    },
    {
      code: "\nimport type { foo } from './foo';\ntype bar = number;\n\n// 'foo' is already declared in the upper scope\n// 'bar' is fine\nfunction doThing(foo: number, bar: number) {}\n      ",
      options: [
        {
          ignoreTypeValueShadow: true,
        },
      ],
    },
    {
      code: "\nimport { type foo } from './foo';\n\n// 'foo' is already declared in the upper scope\nfunction doThing(foo: number) {}\n      ",
      options: [
        {
          ignoreTypeValueShadow: true,
        },
      ],
    },
    {
      code: "const a = [].find(a => a);",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "\nconst a = [].find(function (a) {\n  return a;\n});\n      ",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const [a = [].find(a => true)] = dummy;",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const { a = [].find(a => true) } = dummy;",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "function func(a = [].find(a => true)) {}",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "\nfor (const a in [].find(a => true)) {\n}\n      ",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "\nfor (const a of [].find(a => true)) {\n}\n      ",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const a = [].map(a => true).filter(a => a === 'b');",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "\nconst a = []\n  .map(a => true)\n  .filter(a => a === 'b')\n  .find(a => a === 'c');\n      ",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const { a } = (({ a }) => ({ a }))();",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "\nconst person = people.find(item => {\n  const person = item.name;\n  return person === 'foo';\n});\n      ",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "var y = bar || foo(y => y);",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "var y = bar && foo(y => y);",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "var z = bar(foo(z => z));",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "var z = boo(bar(foo(z => z)));",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "\nvar match = function (person) {\n  return person.name === 'foo';\n};\nconst person = [].find(match);\n      ",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const a = foo(x || (a => {}));",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const { a = 1 } = foo(a => {});",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const person = { ...people.find(person => person.firstName.startsWith('s')) };",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2021,
      },
    },
    {
      code: "\nconst person = {\n  firstName: people\n    .filter(person => person.firstName.startsWith('s'))\n    .map(person => person.firstName)[0],\n};\n      ",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2021,
      },
    },
    {
      code: "\n() => {\n  const y = foo(y => y);\n};\n      ",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const x = (x => x)();",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "var y = bar || (y => y)();",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "var y = bar && (y => y)();",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "var x = (x => x)((y => y)());",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const { a = 1 } = (a => {})();",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "\n() => {\n  const y = (y => y)();\n};\n      ",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const [x = y => y] = [].map(y => y);",
    },
    {
      code: "\ntype Foo<A> = 1;\ntype A = 1;\n      ",
      options: [
        {
          hoist: "never",
        },
      ],
    },
    {
      code: "\ninterface Foo<A> {}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "never",
        },
      ],
    },
    {
      code: "\ninterface Foo<A> {}\ninterface A {}\n      ",
      options: [
        {
          hoist: "never",
        },
      ],
    },
    {
      code: "\ntype Foo<A> = 1;\ninterface A {}\n      ",
      options: [
        {
          hoist: "never",
        },
      ],
    },
    {
      code: "\n{\n  type A = 1;\n}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "never",
        },
      ],
    },
    {
      code: "\n{\n  interface Foo<A> {}\n}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "never",
        },
      ],
    },
    {
      code: "\ntype Foo<A> = 1;\ntype A = 1;\n      ",
      options: [
        {
          hoist: "functions",
        },
      ],
    },
    {
      code: "\ninterface Foo<A> {}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "functions",
        },
      ],
    },
    {
      code: "\ninterface Foo<A> {}\ninterface A {}\n      ",
      options: [
        {
          hoist: "functions",
        },
      ],
    },
    {
      code: "\ntype Foo<A> = 1;\ninterface A {}\n      ",
      options: [
        {
          hoist: "functions",
        },
      ],
    },
    {
      code: "\n{\n  type A = 1;\n}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "functions",
        },
      ],
    },
    {
      code: "\n{\n  interface Foo<A> {}\n}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "functions",
        },
      ],
    },
    {
      code: "\nimport type { Foo } from 'bar';\n\ndeclare module 'bar' {\n  export type Foo = string;\n}\n      ",
    },
    {
      code: "\nimport type { Foo } from 'bar';\n\ndeclare module 'bar' {\n  interface Foo {\n    x: string;\n  }\n}\n      ",
    },
    {
      code: "\nimport { type Foo } from 'bar';\n\ndeclare module 'bar' {\n  export type Foo = string;\n}\n      ",
    },
    {
      code: "\nimport { type Foo } from 'bar';\n\ndeclare module 'bar' {\n  export interface Foo {\n    x: string;\n  }\n}\n      ",
    },
    {
      code: "\nimport { type Foo } from 'bar';\n\ndeclare module 'bar' {\n  type Foo = string;\n}\n      ",
    },
    {
      code: "\nimport { type Foo } from 'bar';\n\ndeclare module 'bar' {\n  interface Foo {\n    x: string;\n  }\n}\n      ",
    },
    {
      code: "\ndeclare const foo1: boolean;\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          foo1: false,
        },
      },
      filename: "decl.d.ts",
    },
    {
      code: "\ndeclare let foo2: boolean;\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          foo2: false,
        },
      },
      filename: "decl.d.ts",
    },
    {
      code: "\ndeclare var foo3: boolean;\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          foo3: false,
        },
      },
      filename: "decl.d.ts",
    },
    {
      code: "\nfunction foo4(name: string): void;\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          foo4: false,
        },
      },
      filename: "decl.d.ts",
    },
    {
      code: "\ndeclare class Foopy1 {\n  name: string;\n}\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          Foopy1: false,
        },
      },
      filename: "decl.d.ts",
    },
    {
      code: "\ndeclare interface Foopy2 {\n  name: string;\n}\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          Foopy2: false,
        },
      },
      filename: "decl.d.ts",
    },
    {
      code: "\ndeclare type Foopy3 = {\n  x: number;\n};\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          Foopy3: false,
        },
      },
      filename: "decl.d.ts",
    },
    {
      code: "\ndeclare enum Foopy4 {\n  x,\n}\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          Foopy4: false,
        },
      },
      filename: "decl.d.ts",
    },
    {
      code: "\ndeclare namespace Foopy5 {}\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          Foopy5: false,
        },
      },
      filename: "decl.d.ts",
    },
    {
      code: "\ndeclare;\nfoo5: boolean;\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          foo5: false,
        },
      },
      filename: "decl.d.ts",
    },
  ],
  invalid: [
    {
      code: "const f = (function f() {} as unknown);",
      errors: [
        {
          messageId: "noShadow",
          message: "'f' is already declared in the upper scope on line 1 column 7.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "enum A { A }",
      errors: [
        {
          messageId: "noEnumShadow",
          message: "Enum members are added to the enum scope, so references to 'A' in enum member initializers resolve to this member instead of the declaration in the upper scope on line 1 column 6.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: "function fn(Record: number) {}",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: "noShadowGlobal",
          message: "'Record' is already a global variable.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: "function fn() { type Record = string; }",
      options: [
        {
          builtinGlobals: true,
          ignoreTypeValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadowGlobal",
          message: "'Record' is already a global variable.",
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: "import type { Record } from 'pkg';",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: "noShadowGlobal",
          message: "'Record' is already a global variable.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: "type Fn = (Record: number) => void;",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: "noShadowGlobal",
          message: "'Record' is already a global variable.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: "interface Record {} namespace Record { export const value = 1; }",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: "noShadowGlobal",
          message: "'Record' is already a global variable.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "class C { static method<Record>() {} }",
      options: [
        {
          builtinGlobals: true,
          ignoreTypeValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadowGlobal",
          message: "'Record' is already a global variable.",
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: "\nimport { request, type Server } from 'node:http';\nfunction use(request: unknown) {}\n      ",
      errors: [
        {
          messageId: "noShadow",
          message: "'request' is already declared in the upper scope on line 2 column 10.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: "\ntype T = 1;\n{\n  type T = 2;\n}\n      ",
      errors: [
        {
          messageId: "noShadow",
          message: "'T' is already declared in the upper scope on line 2 column 6.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\ntype T = 1;\nfunction foo<T>(arg: T) {}\n      ",
      errors: [
        {
          messageId: "noShadow",
          message: "'T' is already declared in the upper scope on line 2 column 6.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: "\nfunction foo<T>() {\n  return function <T>() {};\n}\n      ",
      errors: [
        {
          messageId: "noShadow",
          message: "'T' is already declared in the upper scope on line 2 column 14.",
          line: 3,
          column: 20,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: "\ntype T = string;\nfunction foo<T extends (arg: any) => void>(arg: T) {}\n      ",
      errors: [
        {
          messageId: "noShadow",
          message: "'T' is already declared in the upper scope on line 2 column 6.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: "\nconst x = 1;\n{\n  type x = string;\n}\n      ",
      options: [
        {
          ignoreTypeValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'x' is already declared in the upper scope on line 2 column 7.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\ntype Foo = 1;\n      ",
      options: [
        {
          builtinGlobals: true,
          ignoreTypeValueShadow: false,
        },
      ],
      languageOptions: {
        globals: {
          Foo: "writable",
        },
      },
      errors: [
        {
          messageId: "noShadowGlobal",
          message: "'Foo' is already a global variable.",
          line: 2,
          column: 6,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\nconst test = 1;\ntype Fn = (test: string) => typeof test;\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'test' is already declared in the upper scope on line 2 column 7.",
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\ntype Fn = (Foo: string) => typeof Foo;\n      ",
      options: [
        {
          builtinGlobals: true,
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      languageOptions: {
        globals: {
          Foo: "writable",
        },
      },
      errors: [
        {
          messageId: "noShadowGlobal",
          message: "'Foo' is already a global variable.",
          line: 2,
          column: 12,
          endLine: 2,
          endColumn: 15,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ninterface Test {\n  (arg: string): typeof arg;\n}\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 5,
          column: 4,
          endLine: 5,
          endColumn: 7,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ninterface Test {\n  p1(arg: string): typeof arg;\n}\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 5,
          column: 6,
          endLine: 5,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ndeclare function test(arg: string): typeof arg;\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 4,
          column: 23,
          endLine: 4,
          endColumn: 26,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ndeclare const test: (arg: string) => typeof arg;\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 4,
          column: 22,
          endLine: 4,
          endColumn: 25,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ndeclare class Test {\n  p1(arg: string): typeof arg;\n}\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 5,
          column: 6,
          endLine: 5,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ndeclare const Test: {\n  new (arg: string): typeof arg;\n};\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 5,
          column: 8,
          endLine: 5,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ntype Bar = new (arg: number) => typeof arg;\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 4,
          column: 17,
          endLine: 4,
          endColumn: 20,
        },
      ],
    },
    {
      code: "\nconst arg = 0;\n\ndeclare namespace Lib {\n  function test(arg: string): typeof arg;\n}\n      ",
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 5,
          column: 17,
          endLine: 5,
          endColumn: 20,
        },
      ],
    },
    {
      code: "\nimport type { foo } from './foo';\nfunction doThing(foo: number) {}\n      ",
      options: [
        {
          ignoreTypeValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'foo' is already declared in the upper scope on line 2 column 15.",
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: "\nimport { type foo } from './foo';\nfunction doThing(foo: number) {}\n      ",
      options: [
        {
          ignoreTypeValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'foo' is already declared in the upper scope on line 2 column 15.",
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: "\nimport { foo } from './foo';\nfunction doThing(foo: number, bar: number) {}\n      ",
      options: [
        {
          ignoreTypeValueShadow: true,
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'foo' is already declared in the upper scope on line 2 column 10.",
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: "\ninterface Foo {}\n\ndeclare module 'bar' {\n  export interface Foo {\n    x: string;\n  }\n}\n      ",
      errors: [
        {
          messageId: "noShadow",
          message: "'Foo' is already declared in the upper scope on line 2 column 11.",
          line: 5,
          column: 20,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: "\nimport type { Foo } from 'bar';\n\ndeclare module 'baz' {\n  export interface Foo {\n    x: string;\n  }\n}\n      ",
      errors: [
        {
          messageId: "noShadow",
          message: "'Foo' is already declared in the upper scope on line 2 column 15.",
          line: 5,
          column: 20,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: "\nimport { type Foo } from 'bar';\n\ndeclare module 'baz' {\n  export interface Foo {\n    x: string;\n  }\n}\n      ",
      errors: [
        {
          messageId: "noShadow",
          message: "'Foo' is already declared in the upper scope on line 2 column 15.",
          line: 5,
          column: 20,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: "\nlet x = foo((x, y) => {});\nlet y;\n      ",
      options: [
        {
          hoist: "all",
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: "noShadow",
          message: "'x' is already declared in the upper scope on line 2 column 5.",
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 15,
        },
        {
          messageId: "noShadow",
          message: "'y' is already declared in the upper scope on line 3 column 5.",
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: "\nlet x = foo((x, y) => {});\nlet y;\n      ",
      options: [
        {
          hoist: "functions",
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: "noShadow",
          message: "'x' is already declared in the upper scope on line 2 column 5.",
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 15,
        },
      ],
    },
    {
      code: "\ntype Foo<A> = 1;\ntype A = 1;\n      ",
      options: [
        {
          hoist: "types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\ninterface Foo<A> {}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\ninterface Foo<A> {}\ninterface A {}\n      ",
      options: [
        {
          hoist: "types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\ntype Foo<A> = 1;\ninterface A {}\n      ",
      options: [
        {
          hoist: "types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\n{\n  type A = 1;\n}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\n{\n  interface A {}\n}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: "\ntype Foo<A> = 1;\ntype A = 1;\n      ",
      options: [
        {
          hoist: "all",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\ninterface Foo<A> {}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "all",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\ninterface Foo<A> {}\ninterface A {}\n      ",
      options: [
        {
          hoist: "all",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\ntype Foo<A> = 1;\ninterface A {}\n      ",
      options: [
        {
          hoist: "all",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\n{\n  type A = 1;\n}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "all",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\n{\n  interface A {}\n}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "all",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: "\ntype Foo<A> = 1;\ntype A = 1;\n      ",
      options: [
        {
          hoist: "functions-and-types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\ninterface Foo<A> {}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "functions-and-types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\ninterface Foo<A> {}\ninterface A {}\n      ",
      options: [
        {
          hoist: "functions-and-types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\ntype Foo<A> = 1;\ninterface A {}\n      ",
      options: [
        {
          hoist: "functions-and-types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\n{\n  type A = 1;\n}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "functions-and-types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\n{\n  interface A {}\n}\ntype A = 1;\n      ",
      options: [
        {
          hoist: "functions-and-types",
        },
      ],
      errors: [
        {
          messageId: "noShadow",
          message: "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: "\nfunction foo<T extends (...args: any[]) => any>(fn: T, args: any[]) {}\n      ",
      options: [
        {
          builtinGlobals: true,
          ignoreTypeValueShadow: false,
        },
      ],
      languageOptions: {
        globals: {
          args: "writable",
        },
      },
      errors: [
        {
          messageId: "noShadowGlobal",
          message: "'args' is already a global variable.",
          line: 2,
          column: 56,
          endLine: 2,
          endColumn: 60,
        },
      ],
    },
    {
      code: "\ndeclare const has = (environment: 'dev' | 'prod' | 'test') => boolean;\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          has: false,
        },
      },
      errors: [
        {
          messageId: "noShadowGlobal",
          message: "'has' is already a global variable.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: "\ndeclare const has: (environment: 'dev' | 'prod' | 'test') => boolean;\nconst fn = (has: string) => {};\n      ",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          has: false,
        },
      },
      filename: "decl.d.ts",
      errors: [
        {
          messageId: "noShadow",
          message: "'has' is already declared in the upper scope on line 2 column 15.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
  ],
});
