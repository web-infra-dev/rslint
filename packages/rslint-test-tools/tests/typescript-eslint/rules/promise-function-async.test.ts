// Every case from typescript-eslint v8.29.1, with messages and full ranges
// additionally pinned from the same upstream version. No upstream expectations are relaxed.
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.29.1/packages/eslint-plugin/tests/rules/promise-function-async.test.ts
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  fixtureFiles: {
    'tsconfig.json':
      '{"compilerOptions":{"jsx":"preserve","target":"es2015","module":"commonjs","strict":true,"types":[],"lib":["es2015","es2017","esnext"],"experimentalDecorators":true},"include":["*.ts","*.tsx"]}',
  },
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
    },
  },
});

ruleTester.run('promise-function-async', {
  valid: [
    {
      code: `
const nonAsyncNonPromiseArrowFunction = (n: number) => n;
    `,
      filename: 'file.ts',
    },
    {
      code: `
function nonAsyncNonPromiseFunctionDeclaration(n: number) {
  return n;
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
const asyncPromiseFunctionExpressionA = async function (p: Promise<void>) {
  return p;
};
    `,
      filename: 'file.ts',
    },
    {
      code: `
const asyncPromiseFunctionExpressionB = async function () {
  return new Promise<void>();
};
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Test {
  public nonAsyncNonPromiseArrowFunction = (n: number) => n;
  public nonAsyncNonPromiseMethod() {
    return 0;
  }

  public async asyncPromiseMethodA(p: Promise<void>) {
    return p;
  }

  public async asyncPromiseMethodB() {
    return new Promise<void>();
  }
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
class InvalidAsyncModifiers {
  public constructor() {
    return new Promise<void>();
  }
  public get asyncGetter() {
    return new Promise<void>();
  }
  public set asyncGetter(p: Promise<void>) {
    return p;
  }
  public get asyncGetterFunc() {
    return async () => new Promise<void>();
  }
  public set asyncGetterFunc(p: () => Promise<void>) {
    return p;
  }
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
const invalidAsyncModifiers = {
  get asyncGetter() {
    return new Promise<void>();
  },
  set asyncGetter(p: Promise<void>) {
    return p;
  },
  get asyncGetterFunc() {
    return async () => new Promise<void>();
  },
  set asyncGetterFunc(p: () => Promise<void>) {
    return p;
  },
};
    `,
      filename: 'file.ts',
    },
    {
      code: `
      export function valid(n: number) {
        return n;
      }
    `,
      filename: 'file.ts',
    },
    {
      code: `
      export default function invalid(n: number) {
        return n;
      }
    `,
      filename: 'file.ts',
    },
    {
      code: `
      class Foo {
        constructor() {}
      }
    `,
      filename: 'file.ts',
    },
    {
      code: `
class Foo {
  async catch<T>(arg: Promise<T>) {
    return arg;
  }
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
function returnsAny(): any {
  return 0;
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowAny: true,
        },
      ],
    },
    {
      code: `
function returnsUnknown(): unknown {
  return 0;
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowAny: true,
        },
      ],
    },
    {
      code: `
interface ReadableStream {}
interface Options {
  stream: ReadableStream;
}

type Return = ReadableStream | Promise<void>;
const foo = (options: Options): Return => {
  return options.stream ? asStream(options) : asPromise(options);
};
      `,
      filename: 'file.ts',
    },
    {
      code: `
function foo(): Promise<string> | boolean {
  return Math.random() > 0.5 ? Promise.resolve('value') : false;
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
abstract class Test {
  abstract test1(): Promise<number>;

  // abstract method with body is always an error but it still parses into valid AST
  abstract test2(): Promise<number> {
    return Promise.resolve(1);
  }
}
      `,
      filename: 'file.ts',
    },
    {
      code: `
function promiseInUnionWithExplicitReturnType(
  p: boolean,
): Promise<number> | number {
  return p ? Promise.resolve(5) : 5;
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
function explicitReturnWithPromiseInUnion(): Promise<number> | number {
  return 5;
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
async function asyncFunctionReturningUnion(p: boolean) {
  return p ? Promise.resolve(5) : 5;
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
function overloadingThatCanReturnPromise(): Promise<number>;
function overloadingThatCanReturnPromise(a: boolean): number;
function overloadingThatCanReturnPromise(
  a?: boolean,
): Promise<number> | number {
  return Promise.resolve(5);
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
function overloadingThatCanReturnPromise(a: boolean): number;
function overloadingThatCanReturnPromise(): Promise<number>;
function overloadingThatCanReturnPromise(
  a?: boolean,
): Promise<number> | number {
  return Promise.resolve(5);
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
function a(): Promise<void>;
function a(x: boolean): void;
function a(x?: boolean) {
  if (x == null) return Promise.reject(new Error());
  throw new Error();
}
    `,
      filename: 'file.ts',
    },
    {
      code: `
function overloadingThatIncludeUnknown(): number;
function overloadingThatIncludeUnknown(a: boolean): unknown;
function overloadingThatIncludeUnknown(a?: boolean): unknown | number {
  return Promise.resolve(5);
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowAny: true,
        },
      ],
    },
    {
      code: `
function overloadingThatIncludeAny(): number;
function overloadingThatIncludeAny(a: boolean): any;
function overloadingThatIncludeAny(a?: boolean): any | number {
  return Promise.resolve(5);
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowAny: true,
        },
      ],
    },
  ],
  invalid: [
    {
      code: `
function returnsAny(): any {
  return 0;
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowAny: false,
        },
      ],
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 20,
        },
      ],
    },
    {
      code: `
function returnsUnknown(): unknown {
  return 0;
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowAny: false,
        },
      ],
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
const nonAsyncPromiseFunctionExpressionA = function (p: Promise<void>) {
  return p;
};
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 44,
          endLine: 2,
          endColumn: 53,
        },
      ],
      output: `
const nonAsyncPromiseFunctionExpressionA = async function (p: Promise<void>) {
  return p;
};
      `,
    },
    {
      code: `
const nonAsyncPromiseFunctionExpressionB = function () {
  return new Promise<void>();
};
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 44,
          endLine: 2,
          endColumn: 53,
        },
      ],
      output: `
const nonAsyncPromiseFunctionExpressionB = async function () {
  return new Promise<void>();
};
      `,
    },
    {
      code: `
function nonAsyncPromiseFunctionDeclarationA(p: Promise<void>) {
  return p;
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 45,
        },
      ],
      output: `
async function nonAsyncPromiseFunctionDeclarationA(p: Promise<void>) {
  return p;
}
      `,
    },
    {
      code: `
function nonAsyncPromiseFunctionDeclarationB() {
  return new Promise<void>();
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 45,
        },
      ],
      output: `
async function nonAsyncPromiseFunctionDeclarationB() {
  return new Promise<void>();
}
      `,
    },
    {
      code: `
const nonAsyncPromiseArrowFunctionA = (p: Promise<void>) => p;
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 58,
          endLine: 2,
          endColumn: 60,
        },
      ],
      output: `
const nonAsyncPromiseArrowFunctionA = async (p: Promise<void>) => p;
      `,
    },
    {
      code: `
const nonAsyncPromiseArrowFunctionB = () => new Promise<void>();
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 42,
          endLine: 2,
          endColumn: 44,
        },
      ],
      output: `
const nonAsyncPromiseArrowFunctionB = async () => new Promise<void>();
      `,
    },
    {
      code: `
const functions = {
  nonAsyncPromiseMethod() {
    return Promise.resolve(1);
  },
};
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 24,
        },
      ],
      output: `
const functions = {
  async nonAsyncPromiseMethod() {
    return Promise.resolve(1);
  },
};
      `,
    },
    {
      code: `
class Test {
  public nonAsyncPromiseMethodA(p: Promise<void>) {
    return p;
  }

  public static nonAsyncPromiseMethodB() {
    return new Promise<void>();
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 32,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 7,
          column: 3,
          endLine: 7,
          endColumn: 39,
        },
      ],
      output: `
class Test {
  public async nonAsyncPromiseMethodA(p: Promise<void>) {
    return p;
  }

  public static async nonAsyncPromiseMethodB() {
    return new Promise<void>();
  }
}
      `,
    },
    {
      code: `
const nonAsyncPromiseFunctionExpression = function (p: Promise<void>) {
  return p;
};

function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = (p: Promise<void>) => p;

class Test {
  public nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          checkArrowFunctions: false,
        },
      ],
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 43,
          endLine: 2,
          endColumn: 52,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 44,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 13,
          column: 3,
          endLine: 13,
          endColumn: 31,
        },
      ],
      output: `
const nonAsyncPromiseFunctionExpression = async function (p: Promise<void>) {
  return p;
};

async function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = (p: Promise<void>) => p;

class Test {
  public async nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
    },
    {
      code: `
const nonAsyncPromiseFunctionExpression = function (p: Promise<void>) {
  return p;
};

function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = (p: Promise<void>) => p;

class Test {
  public nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          checkFunctionDeclarations: false,
        },
      ],
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 43,
          endLine: 2,
          endColumn: 52,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 10,
          column: 57,
          endLine: 10,
          endColumn: 59,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 13,
          column: 3,
          endLine: 13,
          endColumn: 31,
        },
      ],
      output: `
const nonAsyncPromiseFunctionExpression = async function (p: Promise<void>) {
  return p;
};

function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = async (p: Promise<void>) => p;

class Test {
  public async nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
    },
    {
      code: `
const nonAsyncPromiseFunctionExpression = function (p: Promise<void>) {
  return p;
};

function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = (p: Promise<void>) => p;

class Test {
  public nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          checkFunctionExpressions: false,
        },
      ],
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 44,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 10,
          column: 57,
          endLine: 10,
          endColumn: 59,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 13,
          column: 3,
          endLine: 13,
          endColumn: 31,
        },
      ],
      output: `
const nonAsyncPromiseFunctionExpression = function (p: Promise<void>) {
  return p;
};

async function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = async (p: Promise<void>) => p;

class Test {
  public async nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
    },
    {
      code: `
const nonAsyncPromiseFunctionExpression = function (p: Promise<void>) {
  return p;
};

function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = (p: Promise<void>) => p;

class Test {
  public nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          checkMethodDeclarations: false,
        },
      ],
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 43,
          endLine: 2,
          endColumn: 52,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 44,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 10,
          column: 57,
          endLine: 10,
          endColumn: 59,
        },
      ],
      output: `
const nonAsyncPromiseFunctionExpression = async function (p: Promise<void>) {
  return p;
};

async function nonAsyncPromiseFunctionDeclaration(p: Promise<void>) {
  return p;
}

const nonAsyncPromiseArrowFunction = async (p: Promise<void>) => p;

class Test {
  public nonAsyncPromiseMethod(p: Promise<void>) {
    return p;
  }
}
      `,
    },
    {
      code: `
class PromiseType {}

const returnAllowedType = () => new PromiseType();
      `,
      filename: 'file.ts',
      options: [
        {
          allowedPromiseNames: ['PromiseType'],
        },
      ],
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 4,
          column: 30,
          endLine: 4,
          endColumn: 32,
        },
      ],
      output: `
class PromiseType {}

const returnAllowedType = async () => new PromiseType();
      `,
    },
    {
      code: `
interface SPromise<T> extends Promise<T> {}
function foo(): Promise<string> | SPromise<boolean> {
  return Math.random() > 0.5
    ? Promise.resolve('value')
    : Promise.resolve(false);
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowedPromiseNames: ['SPromise'],
        },
      ],
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 13,
        },
      ],
      output: `
interface SPromise<T> extends Promise<T> {}
async function foo(): Promise<string> | SPromise<boolean> {
  return Math.random() > 0.5
    ? Promise.resolve('value')
    : Promise.resolve(false);
}
      `,
    },
    {
      code: `
class Test {
  @decorator
  public test() {
    return Promise.resolve(123);
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 14,
        },
      ],
      output: `
class Test {
  @decorator
  public async test() {
    return Promise.resolve(123);
  }
}
      `,
    },
    {
      code: `
class Test {
  @decorator(async () => {})
  static protected[(1)]() {
    return Promise.resolve(1);
  }
  public'bar'() {
    return Promise.resolve(2);
  }
  private['baz']() {
    return Promise.resolve(3);
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 24,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 7,
          column: 3,
          endLine: 7,
          endColumn: 14,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 10,
          column: 3,
          endLine: 10,
          endColumn: 17,
        },
      ],
      output: `
class Test {
  @decorator(async () => {})
  static protected async [(1)]() {
    return Promise.resolve(1);
  }
  public async 'bar'() {
    return Promise.resolve(2);
  }
  private async ['baz']() {
    return Promise.resolve(3);
  }
}
      `,
    },
    {
      code: `
class Foo {
  catch() {
    return Promise.resolve(1);
  }

  public default() {
    return Promise.resolve(2);
  }

  @decorator
  private case<T>() {
    return Promise.resolve(3);
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 8,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 7,
          column: 3,
          endLine: 7,
          endColumn: 17,
        },
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 12,
          column: 3,
          endLine: 12,
          endColumn: 18,
        },
      ],
      output: `
class Foo {
  async catch() {
    return Promise.resolve(1);
  }

  public async default() {
    return Promise.resolve(2);
  }

  @decorator
  private async case<T>() {
    return Promise.resolve(3);
  }
}
      `,
    },
    {
      code: `
const foo = {
  catch() {
    return Promise.resolve(1);
  },
};
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 8,
        },
      ],
      output: `
const foo = {
  async catch() {
    return Promise.resolve(1);
  },
};
      `,
    },
    {
      code: `
function promiseInUnionWithoutExplicitReturnType(p: boolean) {
  return p ? Promise.resolve(5) : 5;
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 49,
        },
      ],
      output: `
async function promiseInUnionWithoutExplicitReturnType(p: boolean) {
  return p ? Promise.resolve(5) : 5;
}
      `,
    },
    {
      code: `
function overloadingThatCanReturnPromise(): Promise<number>;
function overloadingThatCanReturnPromise(a: boolean): Promise<string>;
function overloadingThatCanReturnPromise(
  a?: boolean,
): Promise<number | string> {
  return Promise.resolve(5);
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 41,
        },
      ],
      output: `
function overloadingThatCanReturnPromise(): Promise<number>;
function overloadingThatCanReturnPromise(a: boolean): Promise<string>;
async function overloadingThatCanReturnPromise(
  a?: boolean,
): Promise<number | string> {
  return Promise.resolve(5);
}
      `,
    },
    {
      code: `
function overloadingThatIncludeAny(): number;
function overloadingThatIncludeAny(a: boolean): any;
function overloadingThatIncludeAny(a?: boolean): any | number {
  return Promise.resolve(5);
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowAny: false,
        },
      ],
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 35,
        },
      ],
    },
    {
      code: `
function overloadingThatIncludeUnknown(): number;
function overloadingThatIncludeUnknown(a: boolean): unknown;
function overloadingThatIncludeUnknown(a?: boolean): unknown | number {
  return Promise.resolve(5);
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowAny: false,
        },
      ],
      errors: [
        {
          messageId: 'missingAsync',
          message: 'Functions that return promises must be async.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 39,
        },
      ],
    },
  ],
});
