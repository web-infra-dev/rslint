// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-type-error.js
import {
  RuleTester,
  type ValidTestCase,
  type InvalidTestCase,
} from '../rule-tester';

const languageOptions = { sourceType: 'module' as const };

const valid: ValidTestCase[] = [
  {
    code: "if (MrFuManchu.name !== 'Fu Manchu' || MrFuManchu.isMale === false) {\n\tthrow new Error('How cant Fu Manchu be Fu Manchu?');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (wrapper.g.ary.isArray(foo) || wrapper.f.g.ary.isView(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (wrapper.g.ary(foo) || wrapper.f.g.ary.isPiew(foo)) {\n\tthrow new Error();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (Array.isArray()) {\n\tthrow new Error('Woohoo - isArray is broken!');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\tthrow new CustomError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (Array.isArray(foo)) {\n\tthrow new Error.foo();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (Array.isArray(foo)) {\n\tthrow new Error.foo;\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (Array.isArray(foo)) {\n\tthrow new foo.Error;\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (Array.isArray(foo)) {\n\tthrow new foo.Error('My name is Foo Manchu');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\tthrow Error('This is fo FooBar', foo);\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\tnew Error('This is fo FooBar', foo);\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "function test(foo) {\n\tif (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\t\treturn new Error('This is fo FooBar', foo);\n\t}\n\treturn foo;\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (Array.isArray(foo) || ArrayBuffer.isView(foo)) {\n\tlastError = new Error('This is fo FooBar', foo);\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (!isFinite(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (isNaN(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (isArray(foo)) {\n\tthrow new Error();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (foo instanceof boo) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (typeof boo === 'Boo') {\n\tthrow new TypeError();\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (typeof boo === 'Boo') {\n\tsome.thing.else.happens.before();\n\tthrow new Error();\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (Number.isNaN(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (Number.isFinite(foo) && Number.isSafeInteger(foo) && Number.isInteger(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (Array.isArray(foo) || (Blob.isBlob(foo) || Blip.isBlip(foo))) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (typeof foo === 'object' || (Object.isFrozen(foo) || 'String' === typeof foo)) {\n\tthrow new TypeError();\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (isNaN) {\n\tthrow new Error();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (isObjectLike) {\n\tthrow new Error();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (isNaN.foo()) {\n\tthrow new Error();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (typeof foo !== 'object' || foo.bar() === false) {\n\tthrow new TypeError('Expected Foo being bar!');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (foo instanceof Foo) {\n\tthrow new TypeError('Expected Foo being bar!');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (!foo instanceof Foo) {\n\tthrow new TypeError('Expected Foo being bar!');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (foo instanceof Foo === false) {\n\tthrow new TypeError('Expected Foo being bar!');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "throw new Error('💣')",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (!Number.isNaN(foo) && foo === 10) {\n\tthrow new Error('foo is not 10!');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "function foo(foo) {\n\tif (!Number.isNaN(foo) && foo === 10) {\n\t\ttimesFooWas10 += 1;\n\t\tif (calculateAnswerToLife() !== 42) {\n\t\t\topenIssue('Your program is buggy!');\n\t\t} else {\n\t\t\treturn printAwesomeAnswer(42);\n\t\t}\n\t\tthrow new Error('foo is 10');\n\t}\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "function foo(foo) {\n\tif (!Number.isNaN(foo)) {\n\t\ttimesFooWas10 += 1;\n\t\tif (calculateAnswerToLife({with: foo}) !== 42) {\n\t\t\topenIssue('Your program is buggy!');\n\t\t} else {\n\t\t\treturn printAwesomeAnswer(42);\n\t\t}\n\t\tthrow new Error('foo is 10');\n\t}\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (!x.isFudge()) {\n\tthrow new Error('x is no fudge!');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (!_.isFudge(x)) {\n\tthrow new Error('x is no fudge!');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "switch (something) {\n\tcase 1:\n\t\tbreak;\n\tdefault:\n\t\tthrow new Error('Unknown');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (foo instanceof Error) throw new Error("message")',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (foo instanceof CustomError) throw new Error("message")',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (foo instanceof lib.Error) throw new Error("message")',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: 'if (foo instanceof lib.CustomError) throw new Error("message")',
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (typeof window !== 'undefined') {\n\tthrow new Error('This package requires a browser environment.');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (typeof process === 'undefined') {\n\tthrow new Error('This package requires Node.js.');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if ('undefined' === typeof self) {\n\tthrow new Error('No global available.');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (typeof window !== 'undefined' && typeof foo !== 'string') {\n\tthrow new Error('Mixed environment and type check.');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (typeof window != 'undefined') {\n\tthrow new Error('This package requires a browser environment.');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
  {
    code: "if (typeof x !== 'undefined' || Array.isArray(x)) {\n\tthrow new Error('message');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
  },
];

const invalid: InvalidTestCase[] = [
  {
    code: 'if (!isFinite(foo)) {\n\tthrow new Error();\n}\n',
    output: 'if (!isFinite(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: 'if (Array.isArray(foo) || _.isString(foo)) {\n\tthrow new Error();\n}\n',
    output:
      'if (Array.isArray(foo) || _.isString(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: 'if (isNaN(foo) === false) {\n\tthrow new Error();\n}\n',
    output: 'if (isNaN(foo) === false) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (Array.isArray(foo)) {\n\tthrow new Error('foo is an Array');\n}\n",
    output:
      "if (Array.isArray(foo)) {\n\tthrow new TypeError('foo is an Array');\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: 'if (foo instanceof bar) {\n\tthrow new Error(foobar);\n}\n',
    output: 'if (foo instanceof bar) {\n\tthrow new TypeError(foobar);\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: 'if (_.isElement(foo)) {\n\tthrow new Error();\n}\n',
    output: 'if (_.isElement(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: 'if (_.isElement(foo)) {\n\tthrow new Error;\n}\n',
    output: 'if (_.isElement(foo)) {\n\tthrow new TypeError;\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: 'if (wrapper._.isElement(foo)) {\n\tthrow new Error;\n}\n',
    output: 'if (wrapper._.isElement(foo)) {\n\tthrow new TypeError;\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (typeof foo == 'Foo' || 'Foo' === typeof foo) {\n\tthrow new Error();\n}\n",
    output:
      "if (typeof foo == 'Foo' || 'Foo' === typeof foo) {\n\tthrow new TypeError();\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (typeof foo === 'string') {\n\tthrow new Error();\n}\n",
    output: "if (typeof foo === 'string') {\n\tthrow new TypeError();\n}\n",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: 'if (Number.isFinite(foo) && Number.isSafeInteger(foo) && Number.isInteger(foo)) {\n\tthrow new Error();\n}\n',
    output:
      'if (Number.isFinite(foo) && Number.isSafeInteger(foo) && Number.isInteger(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: 'if (wrapper.n.isFinite(foo) && wrapper.n.isSafeInteger(foo) && wrapper.n.isInteger(foo)) {\n\tthrow new Error();\n}\n',
    output:
      'if (wrapper.n.isFinite(foo) && wrapper.n.isSafeInteger(foo) && wrapper.n.isInteger(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: 'if (wrapper.f.g.n.isFinite(foo) && wrapper.g.n.isSafeInteger(foo) && wrapper.n.isInteger(foo)) {\n\tthrow new Error();\n}\n',
    output:
      'if (wrapper.f.g.n.isFinite(foo) && wrapper.g.n.isSafeInteger(foo) && wrapper.n.isInteger(foo)) {\n\tthrow new TypeError();\n}\n',
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isArguments(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isArguments(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isArray(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isArray(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isArrayBuffer(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isArrayBuffer(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isArrayLike(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isArrayLike(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isArrayLikeObject(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isArrayLikeObject(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isBigInt(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isBigInt(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isBoolean(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isBoolean(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isBuffer(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isBuffer(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isDate(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isDate(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isElement(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isElement(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isError(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isError(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isFinite(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isFinite(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isFunction(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isFunction(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isInteger(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isInteger(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isLength(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isLength(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isMap(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isMap(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isNaN(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isNaN(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isNative(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isNative(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isNil(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isNil(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isNull(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isNull(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isNumber(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isNumber(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isObject(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isObject(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isObjectLike(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isObjectLike(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isPlainObject(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isPlainObject(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isPrototypeOf(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isPrototypeOf(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isRegExp(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isRegExp(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isSafeInteger(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isSafeInteger(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isSet(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isSet(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isString(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isString(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isSymbol(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isSymbol(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isTypedArray(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isTypedArray(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isUndefined(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isUndefined(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isView(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isView(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isWeakMap(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isWeakMap(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isWeakSet(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isWeakSet(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isWindow(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isWindow(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
  {
    code: "if (SomeThing.isXMLDoc(foo) === bar) {\n\tthrow new Error('foo is bar');\n}",
    output:
      "if (SomeThing.isXMLDoc(foo) === bar) {\n\tthrow new TypeError('foo is bar');\n}",
    filename: 'src/virtual.js',
    languageOptions,
    errors: [
      {
        messageId: 'prefer-type-error',
        message:
          '`new Error()` is too unspecific for a type check. Use `new TypeError()` instead.',
      },
    ],
  },
];

new RuleTester().run('prefer-type-error', {} as never, { valid, invalid });
