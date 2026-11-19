// Upstream reference: eslint-plugin-jest@29.16.6; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Shared source templates; reporting ranges remain explicit in each matrix.
const containsMethodsDeclaration = `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

`;
const consoleDeclaration = `class Console {
log(str) {
  process.stdout.write(str);
}
}

const console = new Console();

`;

new RuleTester(fixtureConfig()).run('unbound-method', {
  valid: [
    // Paired controls require real method types; missing console types must fail.
    {
      name: 'typed console.log is safe to pass unbound',
      code: 'const f = console.log;',
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        parserOptions: { project: ['fixtures/upstream/tsconfig.json'] },
      },
    },
    ...(
      [
        `export {};
class Service { method() {} }
const service = new Service();
expect(service.method).toHaveBeenCalledTimes(1);`,
        `export {};
class Service { method() {} }
const service = new Service();
expect(service.method).not.toHaveBeenCalled();`,
        `export {};
class Service { method() {} }
const service = new Service();
expect(service.method).toStrictEqual(somethingElse);`,
        `export {};
class Service { method() {} }
const service = new Service();
expect(value).toBe(service.method);`,
        `export {};
class Service { method() {} }
const service = new Service();
expect(service.method).rejects.toBe(1);`,
        `export {};
class Service { method() {} }
const service = new Service();
jest.mocked(service.method).mockImplementation(() => {});`,
        `export {};
class Service { method() {} }
const service = new Service();
const calls = jest.mocked(service.method).mock.calls;`,
        `export {};
class Service { method() {} }
const service = new Service();
jest['mocked'](service.method);`,
        `export {};
class Service { method() {} }
const service = new Service();
import { jest as j } from '@jest/globals';
j.mocked(service.method);`,
        `export {};
class Service { method() {} }
const service = new Service();
expect(() => { expect(service.method).toHaveBeenCalled(); }).not.toThrow();`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 0 valid 10',
      code: `export {};
class Static { static method() {} }
expect(Static.method).toThrow();`,
      options: [
        {
          ignoreStatic: true,
        },
      ],
    },
    ...(
      [
        `class Console {
  log(str) {
    process.stdout.write(str);
  }
}

const console = new Console();
expect(Console.prototype.log).toHaveBeenCalledTimes(1);`,
        `class Console {
  log(str) {
    process.stdout.write(str);
  }
}

const console = new Console();
expect(Console.prototype.log).not.toHaveBeenCalled();`,
        `class Console {
  log(str) {
    process.stdout.write(str);
  }
}

const console = new Console();
expect(Console.prototype.log).toStrictEqual(somethingElse);`,
        `class Console {
  log(str) {
    process.stdout.write(str);
  }
}

const console = new Console();
jest.mocked(Console.prototype.log).mockImplementation(() => {});`,
        `expect(() => {
   \u0020` +
          consoleDeclaration +
          `    expect(Console.prototype.log).toHaveBeenCalledTimes(1);
  }).not.toThrow();`,
        'expect(() => Promise.resolve().then(console.log)).not.toThrow();',
        'expect(console.log).not.toThrow();',
        'expect(console.log).not.toThrowError();',
        'expect(console.log).not.toThrowErrorMatchingSnapshot();',
        'expect(console.log).not.toThrowErrorMatchingInlineSnapshot();',
        'expect(console.log).toThrow();',
        'expect(console.log).toThrowError();',
        'expect(console.log).toThrowErrorMatchingSnapshot();',
        'expect(console.log).toThrowErrorMatchingInlineSnapshot();',
        `class Service {
  method() {}
}

const service = new Service();
const parameter = jest.mocked(service.method).mock.calls[0][0];`,
        `class Service {
  method() {}
}

const service = new Service();
const calls = jest.mocked(service.method).mock.calls;`,
        `class Service {
  method() {}
}

const service = new Service();
const lastCall = jest.mocked(service.method).mock.calls[0];`,
        `class Service {
  method() {}
}

const service = new Service();
const mockedMethod = jest.mocked(service.method); const parameter = mockedMethod.mock.calls[0][0];`,
        `class Service {
  method() {}
}

const service = new Service();
jest.mocked(service.method).mock;`,
        `class Service {
  method() {}
}

const service = new Service();
const mockProp = jest.mocked(service.method).mock;`,
        `class Service {
  method() {}
}

const service = new Service();
const result = jest.mocked(service.method, true);`,
        `class Service {
  method() {}
}

const service = new Service();
jest.mocked(service.method, { shallow: true });`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream unbound-method jest edition / valid ${index}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
    })),
    ...(
      [
        'Promise.resolve().then(console.log);',
        "['1', '2', '3'].map(Number.parseInt);",
        '[5.2, 7.1, 3.6].map(Math.floor);',
        `
      const foo = Number;
      ['1', '2', '3'].map(foo.parseInt);
   \u0020`,
        `
      const foo = Math;
      [5.2, 7.1, 3.6].map(foo.floor);
   \u0020`,
        "['1', '2', '3'].map(Number['floor']);",
        'const x = console.log;',
        'const x = Object.defineProperty;',
        `
      const foo = Object;
      const x = foo.defineProperty;
   \u0020`,
        'const x = String.fromCharCode;',
        `
      const foo = String;
      const x = foo.fromCharCode;
   \u0020`,
        'const x = RegExp.prototype;',
        'const x = Symbol.keyFor;',
        `
      const foo = Symbol;
      const x = foo.keyFor;
   \u0020`,
        'const x = Array.isArray;',
        `
      const foo = Array;
      const x = foo.isArray;
   \u0020`,
        `
      class Foo extends Array {}
      const x = Foo.isArray;
   \u0020`,
        'const x = Proxy.revocable;',
        `
      const foo = Proxy;
      const x = foo.revocable;
   \u0020`,
        'const x = Date.parse;',
        `
      const foo = Date;
      const x = foo.parse;
   \u0020`,
        'const x = Atomics.load;',
        `
      const foo = Atomics;
      const x = foo.load;
   \u0020`,
        'const x = Reflect.deleteProperty;',
        'const x = JSON.stringify;',
        `
      const foo = JSON;
      const x = foo.stringify;
   \u0020`,
        `
      const o = {
        f: function (this: void) {},
      };
      const f = o.f;
   \u0020`,
        `
      const { alert } = window;
   \u0020`,
        `
      let b = window.blur;
   \u0020`,
        `
      function foo() {}
      const fooObject = { foo };
      const { foo: bar } = fooObject;
   \u0020`,
        containsMethodsDeclaration +
          `instance.bound();
 \u0020`,
        containsMethodsDeclaration +
          `instance.unbound();
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.boundStatic();
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.unboundStatic();
 \u0020`,
        containsMethodsDeclaration +
          `const bound = instance.bound;
 \u0020`,
        containsMethodsDeclaration +
          `const boundStatic = ContainsMethods;
 \u0020`,
        containsMethodsDeclaration +
          `const { bound } = instance;
 \u0020`,
        containsMethodsDeclaration +
          `const { boundStatic } = ContainsMethods;
 \u0020`,
        containsMethodsDeclaration +
          `(instance.bound)();
 \u0020`,
        containsMethodsDeclaration +
          `(instance.unbound)();
 \u0020`,
        containsMethodsDeclaration +
          `(ContainsMethods.boundStatic)();
 \u0020`,
        containsMethodsDeclaration +
          `(ContainsMethods.unboundStatic)();
 \u0020`,
        containsMethodsDeclaration +
          `instance.bound\`\`;
 \u0020`,
        containsMethodsDeclaration +
          `instance.unbound\`\`;
 \u0020`,
        containsMethodsDeclaration +
          `if (instance.bound) { }
 \u0020`,
        containsMethodsDeclaration +
          `if (instance.unbound) { }
 \u0020`,
        containsMethodsDeclaration +
          `if (instance.bound !== undefined) { }
 \u0020`,
        containsMethodsDeclaration +
          `if (instance.unbound !== undefined) { }
 \u0020`,
        containsMethodsDeclaration +
          `if (ContainsMethods.boundStatic) { }
 \u0020`,
        containsMethodsDeclaration +
          `if (ContainsMethods.unboundStatic) { }
 \u0020`,
        containsMethodsDeclaration +
          `if (ContainsMethods.boundStatic !== undefined) { }
 \u0020`,
        containsMethodsDeclaration +
          `if (ContainsMethods.unboundStatic !== undefined) { }
 \u0020`,
        containsMethodsDeclaration +
          `if (ContainsMethods.boundStatic && instance) { }
 \u0020`,
        containsMethodsDeclaration +
          `if (ContainsMethods.unboundStatic && instance) { }
 \u0020`,
        containsMethodsDeclaration +
          `if (instance.bound || instance) { }
 \u0020`,
        containsMethodsDeclaration +
          `if (instance.unbound || instance) { }
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.unboundStatic && 0 || ContainsMethods;
 \u0020`,
        containsMethodsDeclaration +
          `(instance.bound || instance) ? 1 : 0
 \u0020`,
        containsMethodsDeclaration +
          `(instance.unbound || instance) ? 1 : 0
 \u0020`,
        containsMethodsDeclaration +
          `while (instance.bound) { }
 \u0020`,
        containsMethodsDeclaration +
          `while (instance.unbound) { }
 \u0020`,
        containsMethodsDeclaration +
          `while (instance.bound !== undefined) { }
 \u0020`,
        containsMethodsDeclaration +
          `while (instance.unbound !== undefined) { }
 \u0020`,
        containsMethodsDeclaration +
          `while (ContainsMethods.boundStatic) { }
 \u0020`,
        containsMethodsDeclaration +
          `while (ContainsMethods.unboundStatic) { }
 \u0020`,
        containsMethodsDeclaration +
          `while (ContainsMethods.boundStatic !== undefined) { }
 \u0020`,
        containsMethodsDeclaration +
          `while (ContainsMethods.unboundStatic !== undefined) { }
 \u0020`,
        containsMethodsDeclaration +
          `instance.bound as any;
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.boundStatic as any;
 \u0020`,
        containsMethodsDeclaration +
          `instance.bound++;
 \u0020`,
        containsMethodsDeclaration +
          `+instance.bound;
 \u0020`,
        containsMethodsDeclaration +
          `++instance.bound;
 \u0020`,
        containsMethodsDeclaration +
          `instance.bound--;
 \u0020`,
        containsMethodsDeclaration +
          `-instance.bound;
 \u0020`,
        containsMethodsDeclaration +
          `--instance.bound;
 \u0020`,
        containsMethodsDeclaration +
          `instance.bound += 1;
 \u0020`,
        containsMethodsDeclaration +
          `instance.bound -= 1;
 \u0020`,
        containsMethodsDeclaration +
          `instance.bound *= 1;
 \u0020`,
        containsMethodsDeclaration +
          `instance.bound /= 1;
 \u0020`,
        containsMethodsDeclaration +
          `instance.bound || 0;
 \u0020`,
        containsMethodsDeclaration +
          `instance.bound && 0;
 \u0020`,
        containsMethodsDeclaration +
          `instance.bound ? 1 : 0;
 \u0020`,
        containsMethodsDeclaration +
          `instance.unbound ? 1 : 0;
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.boundStatic++;
 \u0020`,
        containsMethodsDeclaration +
          `+ContainsMethods.boundStatic;
 \u0020`,
        containsMethodsDeclaration +
          `++ContainsMethods.boundStatic;
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.boundStatic--;
 \u0020`,
        containsMethodsDeclaration +
          `-ContainsMethods.boundStatic;
 \u0020`,
        containsMethodsDeclaration +
          `--ContainsMethods.boundStatic;
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.boundStatic += 1;
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.boundStatic -= 1;
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.boundStatic *= 1;
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.boundStatic /= 1;
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.boundStatic || 0;
 \u0020`,
        containsMethodsDeclaration +
          `instane.boundStatic && 0;
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.boundStatic ? 1 : 0;
 \u0020`,
        containsMethodsDeclaration +
          `ContainsMethods.unboundStatic ? 1 : 0;
 \u0020`,
        containsMethodsDeclaration +
          `typeof instance.bound === 'function';
 \u0020`,
        containsMethodsDeclaration +
          `typeof instance.unbound === 'function';
 \u0020`,
        containsMethodsDeclaration +
          `typeof ContainsMethods.boundStatic === 'function';
 \u0020`,
        containsMethodsDeclaration +
          `typeof ContainsMethods.unboundStatic === 'function';
 \u0020`,
        containsMethodsDeclaration +
          `instance.unbound = () => {};
 \u0020`,
        containsMethodsDeclaration +
          `instance.unbound = instance.unbound.bind(instance);
 \u0020`,
        containsMethodsDeclaration +
          `if (!!instance.unbound) {}
 \u0020`,
        containsMethodsDeclaration +
          `void instance.unbound
 \u0020`,
        containsMethodsDeclaration +
          `delete instance.unbound
 \u0020`,
        containsMethodsDeclaration +
          `const { double } = arith;
 \u0020`,
        `
interface RecordA {
  readonly type: 'A';
  readonly a: {};
}
interface RecordB {
  readonly type: 'B';
  readonly b: {};
}
type AnyRecord = RecordA | RecordB;

function test(obj: AnyRecord) {
  switch (obj.type) {
  }
}
   \u0020`,
        `
class CommunicationError {
  constructor() {
    const x = CommunicationError.prototype;
  }
}
   \u0020`,
        `
class CommunicationError {}
const x = CommunicationError.prototype;
   \u0020`,
        `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

function foo(instance: ContainsMethods | null) {
  instance?.bound();
  instance?.unbound();

  if (instance?.bound) {
  }
  if (instance?.unbound) {
  }

  typeof instance?.bound === 'function';
  typeof instance?.unbound === 'function';
}
   \u0020`,
        `
interface OptionalMethod {
  mightBeDefined?(): void;
}

const x: OptionalMethod = {};
declare const myCondition: boolean;
if (myCondition || x.mightBeDefined) {
  console.log('hello world');
}
   \u0020`,
        `
class A {
  unbound(): void {
    this.unbound = undefined;
    this.unbound = this.unbound.bind(this);
  }
}
   \u0020`,
        'const { parseInt } = Number;',
        'const { log } = console;',
        `
let parseInt;
({ parseInt } = Number);
   \u0020`,
        `
let log;
({ log } = console);
   \u0020`,
        `
const foo = {
  bar: 'bar',
};
const { bar } = foo;
   \u0020`,
        `
class Foo {
  unbnound() {}
  bar = 4;
}
const { bar } = new Foo();
   \u0020`,
        `
class Foo {
  bound = () => 'foo';
}
const { bound } = new Foo();
   \u0020`,
        `
class Foo {
  bound = () => 'foo';
}
function foo({ bound } = new Foo()) {}
   \u0020`,
        `
class Foo {
  bound = () => 'foo';
}
declare const bar: Foo;
function foo({ bound }: Foo) {}
   \u0020`,
        `
class Foo {
  bound = () => 'foo';
}
class Bar {
  bound = () => 'bar';
}
function foo({ bound }: Foo | Bar) {}
   \u0020`,
        `
class Foo {
  bound = () => 'foo';
}
type foo = ({ bound }: Foo) => void;
   \u0020`,
        `
class Foo {
  unbound = function () {};
}
type foo = ({ unbound }: Foo) => void;
   \u0020`,
        `
class Foo {
  bound = () => 'foo';
}
class Bar {
  bound = () => 'bar';
}
function foo({ bound }: Foo & Bar) {}
   \u0020`,
        `
class Foo {
  unbound = function () {};
}
declare const { unbound }: Foo;
   \u0020`,
        "declare const { unbound } = '***';",
        `
class Foo {
  unbound = function () {};
}
type foo = (a: (b: (c: ({ unbound }: Foo) => void) => void) => void) => void;
   \u0020`,
        `
class Foo {
  unbound = function () {};
}
class Bar {
  property: ({ unbound }: Foo) => void;
}
   \u0020`,
        `
class Foo {
  unbound = function () {};
}
function foo<T extends ({ unbound }: Foo) => void>() {}
   \u0020`,
        `
class Foo {
  unbound = function () {};
}
abstract class Bar {
  abstract foo({ unbound }: Foo);
}
   \u0020`,
        `
class Foo {
  unbound = function () {};
}
declare class Bar {
  foo({ unbound }: Foo);
}
   \u0020`,
        `
class Foo {
  unbound = function () {};
}
declare function foo({ unbound }: Foo);
   \u0020`,
        `
class Foo {
  unbound = function () {};
}
interface Bar {
  foo: ({ unbound }: Foo) => void;
}
   \u0020`,
        `
class Foo {
  unbound = function () {};
}
interface Bar {
  foo({ unbound }: Foo): void;
}
   \u0020`,
        `
class Foo {
  unbound = function () {};
}
interface Bar {
  new ({ unbound }: Foo): Foo;
}
   \u0020`,
        `
class Foo {
  unbound = function () {};
}
type foo = new ({ unbound }: Foo) => void;
   \u0020`,
        'const { unbound } = { unbound: () => {} };',
        'function foo({ unbound }: { unbound: () => void } = { unbound: () => {} }) {}',
        `
class BaseClass {
  x: number = 42;
  logThis() {}
}
class OtherClass extends BaseClass {
  superLogThis: any;
  constructor() {
    super();
    this.superLogThis = super.logThis;
  }
}
const oc = new OtherClass();
oc.superLogThis();
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream unbound-method / valid ${index}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
    })),
  ],
  invalid: [
    // Paired controls require real method types; missing console types must fail.
    {
      name: 'typed console custom method requires its receiver',
      code: 'const f = console.customMethod;',
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        parserOptions: { project: ['fixtures/upstream/tsconfig.json'] },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    ...(
      [
        ...[
          'toThrow',
          'toThrowError',
          'toThrowErrorMatchingSnapshot',
          'toThrowErrorMatchingInlineSnapshot',
        ].flatMap((matcher) =>
          ['', 'not.'].map(
            (modifier): readonly [string, number, number, number] => [
              `export {};
class Service { method() {} }
const service = new Service();
expect(service.method).${modifier}${matcher}();`,
              4,
              8,
              22,
            ],
          ),
        ),
        [
          `export {};
class Service { method() {} }
const service = new Service();
expect(service.method);`,
          4,
          8,
          22,
        ],
        [
          `export {};
class Service { method() {} }
const service = new Service();
expect(service?.method).toBe(1);`,
          4,
          8,
          23,
        ],
        [
          `export {};
class Service { method() {} }
const service = new Service();
const method = service.method;
jest.mocked(method);`,
          4,
          16,
          30,
        ],
        [
          `export {};
class Service { method() {} }
const service = new Service();
Promise.resolve().then(service.method);`,
          4,
          24,
          38,
        ],
        [
          `export {};
class Service { method() {} }
const service = new Service();
expect(() => { Promise.resolve().then(service.method); }).not.toThrow();`,
          4,
          39,
          53,
        ],
        [
          `export {};
class Service { method() {} }
const service = new Service();
const { method } = service;`,
          4,
          9,
          15,
        ],
        [
          `export {};
class Static { static method() {} }
expect(Static.method).toThrow();`,
          3,
          8,
          21,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          consoleDeclaration + '    expect(Console.prototype.log)',
          'upstream unbound-method jest edition / invalid 0',
          9,
          12,
          33,
        ],
        // This reads the matcher method without calling it. The fixture must
        // declare expect's return type as an interface with a method; otherwise
        // an any result makes both linters falsely accept this upstream case.
        [
          'expect(Console.prototype.log).toHaveBeenCalledTimes',
          'upstream unbound-method jest edition / invalid 1',
          1,
          1,
          52,
        ],
        [
          `expect(() => {
     \u0020` +
            consoleDeclaration +
            `      Promise.resolve().then(console.log);
    }).not.toThrow();`,
          'upstream unbound-method jest edition / invalid 2',
          10,
          30,
          41,
        ],
        [
          consoleDeclaration + '    expect(console.log).toThrow();',
          'upstream unbound-method jest edition / invalid 3',
          9,
          12,
          23,
        ],
        [
          consoleDeclaration + '    expect(console.log).toThrowError();',
          'upstream unbound-method jest edition / invalid 4',
          9,
          12,
          23,
        ],
        [
          consoleDeclaration +
            '    expect(console.log).toThrowErrorMatchingSnapshot();',
          'upstream unbound-method jest edition / invalid 5',
          9,
          12,
          23,
        ],
        [
          consoleDeclaration +
            '    expect(console.log).toThrowErrorMatchingInlineSnapshot();',
          'upstream unbound-method jest edition / invalid 6',
          9,
          12,
          23,
        ],
        [
          `class Service {
method() {}
}

const service = new Service();

    const method = service.method;
    jest.mocked(method);`,
          'upstream unbound-method jest edition / invalid 7',
          7,
          20,
          34,
        ],
        [
          `class Service {
method() {}
}

const service = new Service();

    const method = service.method;`,
          'upstream unbound-method jest edition / invalid 8',
          7,
          20,
          34,
        ],
        [
          `class Service {
method() {}
}

const service = new Service();

    Promise.resolve().then(service.method);`,
          'upstream unbound-method jest edition / invalid 9',
          7,
          28,
          42,
        ],
        [
          consoleDeclaration + '    expect(console.log).not.toThrow();',
          'upstream unbound-method jest edition / invalid 10',
          9,
          12,
          23,
        ],
        [
          consoleDeclaration + '    expect(console.log).not.toThrowError();',
          'upstream unbound-method jest edition / invalid 11',
          9,
          12,
          23,
        ],
        [
          consoleDeclaration +
            '    expect(console.log).not.toThrowErrorMatchingSnapshot();',
          'upstream unbound-method jest edition / invalid 12',
          9,
          12,
          23,
        ],
        [
          consoleDeclaration +
            '    expect(console.log).not.toThrowErrorMatchingInlineSnapshot();',
          'upstream unbound-method jest edition / invalid 13',
          9,
          12,
          23,
        ],
      ] as const
    ).map(([code, name, line, column, endColumn]): TestCase => ({
      name,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `
class Console {
  log(str) {
    process.stdout.write(str);
  }
}

const console = new Console();

Promise.resolve().then(console.log);
     \u0020`,
          10,
          24,
        ],
        [
          `
import { console } from './class';
const x = console.log;
     \u0020`,
          3,
          11,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `upstream unbound-method / invalid ${index}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line,
          column,
          endLine: line,
          endColumn: column + 11,
        },
      ],
    })),
    {
      name: 'upstream unbound-method / invalid 2',
      code:
        containsMethodsDeclaration +
        `
function foo(arg: ContainsMethods | null) {
  const unbound = arg?.unbound;
  arg.unbound += 1;
  arg?.unbound as any;
}
     \u0020
 \u0020`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 20,
          column: 19,
          endLine: 20,
          endColumn: 31,
        },
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 21,
          column: 3,
          endLine: 21,
          endColumn: 14,
        },
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 22,
          column: 3,
          endLine: 22,
          endColumn: 15,
        },
      ],
    },
    ...(
      [
        [
          containsMethodsDeclaration +
            `const unbound = instance.unbound;
 \u0020`,
          17,
          33,
        ],
        [
          containsMethodsDeclaration +
            `const unboundStatic = ContainsMethods.unboundStatic;
 \u0020`,
          23,
          52,
        ],
        [
          containsMethodsDeclaration +
            `const { unbound } = instance;
 \u0020`,
          9,
          16,
        ],
        [
          containsMethodsDeclaration +
            `const { unboundStatic } = ContainsMethods;
 \u0020`,
          9,
          22,
        ],
        [
          containsMethodsDeclaration +
            `<any>instance.unbound;
 \u0020`,
          6,
          22,
        ],
        [
          containsMethodsDeclaration +
            `instance.unbound as any;
 \u0020`,
          1,
          17,
        ],
        [
          containsMethodsDeclaration +
            `<any>ContainsMethods.unboundStatic;
 \u0020`,
          6,
          35,
        ],
        [
          containsMethodsDeclaration +
            `ContainsMethods.unboundStatic as any;
 \u0020`,
          1,
          30,
        ],
        [
          containsMethodsDeclaration +
            `instance.unbound || 0;
 \u0020`,
          1,
          17,
        ],
        [
          containsMethodsDeclaration +
            `ContainsMethods.unboundStatic || 0;
 \u0020`,
          1,
          30,
        ],
        [
          containsMethodsDeclaration +
            `instance.unbound ? instance.unbound : null
 \u0020`,
          20,
          36,
        ],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `upstream unbound-method / invalid ${index + 3}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 18,
          column,
          endLine: 18,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream unbound-method / invalid 14',
      code: `
class ContainsMethods {
  unbound?(): void;

  static unboundStatic?(): void;
}

new ContainsMethods().unbound;

ContainsMethods.unboundStatic;
     \u0020`,
      filename: 'fixtures/upstream/file.ts',
      options: [
        {
          ignoreStatic: true,
        },
      ],
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 8,
          column: 1,
          endLine: 8,
          endColumn: 30,
        },
      ],
    },
    ...(
      [
        [
          `
class CommunicationError {
  foo() {}
}
const x = CommunicationError.prototype.foo;
     \u0020`,
          5,
          11,
          43,
        ],
        ['const x = Promise.all;', 1, 11, 22],
        [
          `
class Foo {
  unbound() {}
}
const instance = new Foo();

let x;

x = instance.unbound; // THIS SHOULD ERROR
instance.unbound = x; // THIS SHOULD NOT
     \u0020`,
          9,
          5,
          21,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `upstream unbound-method / invalid ${index + 15}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream unbound-method / invalid 18',
      code: `
class Foo extends Number {
  static parseInt = function (string: string, radix?: number): number {};
}
const foo = Foo;
['1', '2', '3'].map(foo.parseInt);
     \u0020`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 6,
          column: 21,
          endLine: 6,
          endColumn: 33,
        },
      ],
    },
    ...(
      [
        [
          `
declare const foo: Number;
const x = foo.toFixed;
     \u0020`,
          22,
        ],
        [
          `
declare const foo: Object;
const x = foo.hasOwnProperty;
     \u0020`,
          29,
        ],
        [
          `
declare const foo: String;
const x = foo.slice;
     \u0020`,
          20,
        ],
        [
          `
declare const foo: Date;
const x = foo.getTime;
     \u0020`,
          22,
        ],
        [
          `
class Foo extends Number {}
const x = Foo.parseInt;
     \u0020`,
          23,
        ],
        [
          `
class Foo extends String {}
const x = Foo.fromCharCode;
     \u0020`,
          27,
        ],
        [
          `
class Foo extends Object {}
const x = Foo.defineProperty;
     \u0020`,
          29,
        ],
        [
          `
class Foo extends Date {}
const x = Foo.parse;
     \u0020`,
          20,
        ],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream unbound-method / invalid ${index + 19}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream unbound-method / invalid 27',
      code: `
class Foo {
  unbound = function () {};
}
const unbound = new Foo().unbound;
     \u0020`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 5,
          column: 17,
          endLine: 5,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'upstream unbound-method / invalid 28',
      code: `
class Foo {
  unbound() {}
}
const { unbound } = new Foo();
     \u0020`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'upstream unbound-method / invalid 29',
      code: `
class Foo {
  unbound = function () {};
}
const { unbound } = new Foo();
     \u0020`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'upstream unbound-method / invalid 30',
      code: `
class Foo {
  unbound() {}
}
let unbound;
({ unbound } = new Foo());
     \u0020`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 6,
          column: 4,
          endLine: 6,
          endColumn: 11,
        },
      ],
    },
    ...(
      [
        [
          `
class Foo {
  unbound = function () {};
}
let unbound;
({ unbound } = new Foo());
     \u0020`,
          6,
          4,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
function foo({ unbound }: Foo = new Foo()) {}
     \u0020`,
          5,
          16,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
declare const bar: Foo;
function foo({ unbound }: Foo = bar) {}
     \u0020`,
          6,
          16,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
declare const bar: Foo;
function foo({ unbound }: Foo = { unbound: () => {} }) {}
     \u0020`,
          6,
          16,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `upstream unbound-method / invalid ${index + 31}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line,
          column,
          endLine: line,
          endColumn: column + 7,
        },
      ],
    })),
    {
      name: 'upstream unbound-method / invalid 35',
      code: `
class Foo {
  unbound = function () {};
}
declare const bar: Foo;
function foo({ unbound }: Foo = { unbound: function () {} }) {}
     \u0020`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 23,
        },
      ],
    },
    ...(
      [
        [
          `
class Foo {
  unbound = function () {};
}
function foo({ unbound }: Foo) {}
     \u0020`,
          5,
          16,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
function bar(cb: (arg: Foo) => void) {}
bar(({ unbound }) => {});
     \u0020`,
          6,
          8,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
function bar(cb: (arg: { unbound: () => void }) => void) {}
bar(({ unbound } = new Foo()) => {});
     \u0020`,
          6,
          8,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
for (const { unbound } of [new Foo(), new Foo()]) {
}
     \u0020`,
          5,
          14,
        ],
        [
          `
class Foo {
  unbound = function () {};

  foo({ unbound }: Foo) {}
}
     \u0020`,
          5,
          9,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = function () {};
}
function foo({ unbound }: Foo | Bar) {}
     \u0020`,
          8,
          16,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
function foo({ unbound }: { unbound: () => string } | Foo) {}
     \u0020`,
          5,
          16,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = () => {};
}
function foo({ unbound }: Foo | Bar) {}
     \u0020`,
          8,
          16,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
const foo = ({ unbound }: Foo & { foo: () => 'bar' }) => {};
     \u0020`,
          5,
          16,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = () => {};
}
const foo = ({ unbound }: (Foo & { foo: () => 'bar' }) | Bar) => {};
     \u0020`,
          8,
          16,
        ],
        [
          `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = () => {};
}
const foo = ({ unbound }: Foo & Bar) => {};
     \u0020`,
          8,
          16,
        ],
        [
          `
class Foo {
  unbound = function () {};

  other = function () {};
}
class Bar {
  unbound = () => {};
}
const foo = ({ unbound, ...rest }: Foo & Bar) => {};
     \u0020`,
          10,
          16,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `upstream unbound-method / invalid ${index + 36}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line,
          column,
          endLine: line,
          endColumn: column + 7,
        },
      ],
    })),
    ...(
      [
        ['const { unbound } = { unbound: function () {} };', 1, 9],
        [
          `
function foo(
  { unbound }: { unbound: () => void } = { unbound: function () {} },
) {}
     \u0020`,
          3,
          5,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `upstream unbound-method / invalid ${index + 48}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line,
          column,
          endLine: line,
          endColumn: column + 7,
        },
      ],
    })),
    {
      name: 'upstream unbound-method / invalid 50',
      code: `
class Foo {
  floor = function () {};
}

const { floor } = Math.random() > 0.5 ? new Foo() : Math;
     \u0020`,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unbound',
          message:
            'Avoid referencing unbound methods which may cause unintentional scoping of `this`.',
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    ...(
      [
        [
          `
class CommunicationError {
  foo() {}
}
const { foo } = CommunicationError.prototype;
     \u0020`,
          5,
          9,
          12,
        ],
        [
          `
class CommunicationError {
  foo() {}
}
let foo;
({ foo } = CommunicationError.prototype);
     \u0020`,
          6,
          4,
          7,
        ],
        [
          `
import { console } from './class';
const { log } = console;
     \u0020`,
          3,
          9,
          12,
        ],
        ['const { all } = Promise;', 1, 9, 12],
        [
          `
class BaseClass {
  logThis() {}
}
class OtherClass extends BaseClass {
  constructor() {
    super();
    const x = super.logThis;
  }
}
     \u0020`,
          8,
          15,
          28,
        ],
        [
          `
class BaseClass {
  logThis() {}
}
class OtherClass extends BaseClass {
  constructor() {
    super();
    let x;
    x = super.logThis;
  }
}
     \u0020`,
          9,
          9,
          22,
        ],
        [
          `
const values = {
  a() {},
  b: () => {},
};

const { a, b } = values;
     \u0020`,
          7,
          9,
          10,
        ],
        [
          `
const values = {
  a() {},
  b: () => {},
};

const { a: c } = values;
     \u0020`,
          7,
          9,
          10,
        ],
        [
          `
const values = {
  a() {},
  b: () => {},
};

const { b, a } = values;
     \u0020`,
          7,
          12,
          13,
        ],
        [
          `
const objectLiteral = {
  f: function () {},
};
const f = objectLiteral.f;
     \u0020`,
          5,
          11,
          26,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `upstream unbound-method / invalid ${index + 51}`,
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        sourceType: 'module',
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
      errors: [
        {
          messageId: 'unboundWithoutThisAnnotation',
          message: `Avoid referencing unbound methods which may cause unintentional scoping of \`this\`.
If your function does not access \`this\`, you can annotate it with \`this: void\`, or consider using an arrow function instead.`,
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
  ],
});

// Upstream removes its optional Node base-rule dependency in this group.
// The native rule has no corresponding module-mocking execution mode.
new RuleTester(fixtureConfig()).run('unbound-method', {
  valid: [
    ...(
      [
        `class Console {
  log(str) {
    process.stdout.write(str);
  }
}

const console = new Console();
expect(Console.prototype.log).toHaveBeenCalledTimes(1);`,
        `class Console {
  log(str) {
    process.stdout.write(str);
  }
}

const console = new Console();
expect(Console.prototype.log).not.toHaveBeenCalled();`,
        `class Console {
  log(str) {
    process.stdout.write(str);
  }
}

const console = new Console();
expect(Console.prototype.log).toStrictEqual(somethingElse);`,
        `class Console {
  log(str) {
    process.stdout.write(str);
  }
}

const console = new Console();
jest.mocked(Console.prototype.log).mockImplementation(() => {});`,
        `expect(() => {
   \u0020` +
          consoleDeclaration +
          `    expect(Console.prototype.log).toHaveBeenCalledTimes(1);
  }).not.toThrow();`,
        'expect(() => Promise.resolve().then(console.log)).not.toThrow();',
        'expect(console.log).not.toThrow();',
        'expect(console.log).not.toThrowError();',
        'expect(console.log).not.toThrowErrorMatchingSnapshot();',
        'expect(console.log).not.toThrowErrorMatchingInlineSnapshot();',
        'expect(console.log).toThrow();',
        'expect(console.log).toThrowError();',
        'expect(console.log).toThrowErrorMatchingSnapshot();',
        'expect(console.log).toThrowErrorMatchingInlineSnapshot();',
        `class Service {
  method() {}
}

const service = new Service();
const parameter = jest.mocked(service.method).mock.calls[0][0];`,
        `class Service {
  method() {}
}

const service = new Service();
const calls = jest.mocked(service.method).mock.calls;`,
        `class Service {
  method() {}
}

const service = new Service();
const lastCall = jest.mocked(service.method).mock.calls[0];`,
        `class Service {
  method() {}
}

const service = new Service();
const mockedMethod = jest.mocked(service.method); const parameter = mockedMethod.mock.calls[0][0];`,
        `class Service {
  method() {}
}

const service = new Service();
jest.mocked(service.method).mock;`,
        `class Service {
  method() {}
}

const service = new Service();
const mockProp = jest.mocked(service.method).mock;`,
        `class Service {
  method() {}
}

const service = new Service();
const result = jest.mocked(service.method, true);`,
        `class Service {
  method() {}
}

const service = new Service();
jest.mocked(service.method, { shallow: true });`,
        consoleDeclaration + '    expect(Console.prototype.log)',
        'expect(Console.prototype.log).toHaveBeenCalledTimes',
        `expect(() => {
     \u0020` +
          consoleDeclaration +
          `      Promise.resolve().then(console.log);
    }).not.toThrow();`,
        consoleDeclaration + '    expect(console.log).toThrow();',
        consoleDeclaration + '    expect(console.log).toThrowError();',
        consoleDeclaration +
          '    expect(console.log).toThrowErrorMatchingSnapshot();',
        consoleDeclaration +
          '    expect(console.log).toThrowErrorMatchingInlineSnapshot();',
        `class Service {
method() {}
}

const service = new Service();

    const method = service.method;
    jest.mocked(method);`,
        `class Service {
method() {}
}

const service = new Service();

    const method = service.method;`,
        `class Service {
method() {}
}

const service = new Service();

    Promise.resolve().then(service.method);`,
        consoleDeclaration + '    expect(console.log).not.toThrow();',
        consoleDeclaration + '    expect(console.log).not.toThrowError();',
        consoleDeclaration +
          '    expect(console.log).not.toThrowErrorMatchingSnapshot();',
        consoleDeclaration +
          '    expect(console.log).not.toThrowErrorMatchingInlineSnapshot();',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream unbound-method jest edition without type service / valid ${index}`,
      skip: 'The upstream case removes @typescript-eslint/eslint-plugin through Node module mocking; the native rule has no optional Node base-rule dependency to remove.',
      code,
      filename: 'fixtures/upstream/file.ts',
      languageOptions: {
        parserOptions: {
          project: ['fixtures/upstream/tsconfig.json'],
        },
      },
    })),
  ],
  invalid: [],
});
