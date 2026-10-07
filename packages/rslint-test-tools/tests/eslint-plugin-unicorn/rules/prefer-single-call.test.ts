// Ported from eslint-plugin-unicorn v77.0.0 tests and documentation; see LICENSE.
import { RuleTester } from '../rule-tester';

// Array push
new RuleTester().run('prefer-single-call', {} as never, {
  valid: [
    {
      code: 'function f(foo: {push(value: number): void}) { foo.push(1); foo.push(2); }',
      filename: 'case.ts',
    },
    {
      code: 'function makeSink() { return {push(value: number) {}}; } const sink = makeSink(); sink.push(1); sink.push(2);',
      filename: 'case.ts',
    },
    { code: 'foo.forEach(fn);\nfoo.forEach(fn);', filename: 'case.js' },
    { code: 'foo.push(1);', filename: 'case.js' },
    { code: 'foo.push(1);\nfoo.unshift(2);', filename: 'case.js' },
    {
      code: 'foo.push(1);; // <- there is an "EmptyStatement" between\nfoo.push(2);',
      filename: 'case.js',
    },
    { code: 'foo.push(1);\nbar.push(2);', filename: 'case.js' },
    { code: 'foo.push(1);push(2)', filename: 'case.js' },
    { code: 'push(1);foo.push(2)', filename: 'case.js' },
    { code: 'new foo.push(1);foo.push(2)', filename: 'case.js' },
    { code: 'foo.push(1);new foo.push(2)', filename: 'case.js' },
    { code: 'foo[push](1);foo.push(2)', filename: 'case.js' },
    { code: 'foo.push(1);foo[push](2)', filename: 'case.js' },
    { code: 'foo.push(foo.push(1));', filename: 'case.js' },
    { code: 'const length = foo.push(1);\nfoo.push(2);', filename: 'case.js' },
    { code: 'foo.push(1);\nconst length = foo.push(2);', filename: 'case.js' },
    { code: 'foo().push(1);\nfoo().push(2);', filename: 'case.js' },
    { code: 'foo().bar.push(1);\nfoo().bar.push(2);', filename: 'case.js' },
    {
      code: "const stream = new Readable();\nstream.push('one string');\nstream.push('another string');",
      filename: 'case.js',
    },
    {
      code: 'class FooReadable extends Readable {\n\tpushAndEnd(chunk) {\n\t\tthis.push(chunk);\n\t\tthis.push(null);\n\t}\n}',
      filename: 'case.js',
    },
    {
      code: 'class Foo {\n\tpushAndEnd(chunk) {\n\t\tthis.stream.push(chunk);\n\t\tthis.stream.push(null);\n\t}\n}',
      filename: 'case.js',
    },
    {
      code: 'process.stdin.push(chunk);\nprocess.stdin.push(null);',
      filename: 'case.js',
    },
    {
      code: 'process.stdout.push(chunk);\nprocess.stdout.push(null);',
      filename: 'case.js',
    },
    {
      code: 'process.stderr.push(chunk);\nprocess.stderr.push(null);',
      filename: 'case.js',
    },
    {
      code: 'foo.push(1);\nfoo.push(2);\nfoo.bar.push(1);\nfoo.bar.push(2);',
      filename: 'case.js',
      options: [{ ignore: ['foo.push', 'foo.bar.push'] }],
    },
    { code: 'for (const _ of []) foo.push(bar);', filename: 'case.js' },
    { code: 'function bar() {}\nfoo.push(bindEvents);', filename: 'case.js' },
    { code: 'foo.push?.(1);\nfoo.push?.(2);', filename: 'case.js' },
    { code: 'foo.push(1);\nfoo.push?.(2);', filename: 'case.js' },
    { code: 'foo.push?.(1);\nfoo.push(2);', filename: 'case.js' },
    {
      code: 'const foo = new Foo(); foo.push(1); foo.push(2);',
      filename: 'case.js',
    },
  ],
  invalid: [
    {
      code: "const object = {value: 0};\nObject.defineProperty(object, 'value', {get() { return foo.length; }});\nfoo.push(1);\nfoo.push(object.value);",
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1);\nfoo.push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: '(foo.push)(1);\n(foo.push)(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.bar.push(1);\nfoo.bar.push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1);\n(foo).push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push();\nfoo.push();',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1);\nfoo.push();',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push();\nfoo.push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1, 2);\nfoo.push((3), (4));',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1, 2,);\nfoo.push(3, 4);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1, 2);\nfoo.push(3, 4,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1, 2,);\nfoo.push(3, 4,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1, 2, ...a,);\nfoo.push(...b,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(bar());\nfoo.push(1);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1);\nfoo.push(bar());',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1,);\nfoo.push(2,);\nfoo.push(3,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'if (a) {\n\tfoo.push(1);\n\tfoo.push(2);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'switch (a) {\n\tdefault:\n\t\tfoo.push(1);\n\t\tfoo.push(2);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'function a() {\n\tfoo.push(1);\n\tfoo.push(2);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1)\nfoo.push(2)\n;[foo].forEach(bar)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: "foo.bar.push(1);\n(foo)['bar'].push(2);",
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1);\nfoo.push(2);\nstream.push(1);\nstream.push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.bar.push(1);\nfoo.bar.push(2);\nfoo.push(1);\nfoo.push(2);\nbar.foo.push(1);\nbar.foo.push(2);',
      filename: 'case.js',
      options: [{ ignore: ['foo', 'foo.bar'] }],
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.push(1);\nfoo?.push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo?.push(1);\nfoo.push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo?.push(1);\nfoo?.push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo?.bar.push(1);\nfoo?.bar.push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: '(foo as any[]).push(1);\n(foo as any[]).push(2);',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo!.push(1);\nfoo!.push(2);',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'function f(foo: number[]) { foo.push(1); foo.push(2); }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    // rslint always supplies TypeScript type information; the identical type-aware upstream case above/below covers this input.
    // {"code": "const array = [0].map(value => value); array.push(1); array.push(2);", "filename": "case.ts", "errors": [{"messageId": "error/array-push", "message": "Do not call `Array#push()` multiple times."}]},
    {
      code: 'const array = [0].map(value => value); array.push(1); array.push(2);',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'declare const receiver: number[] | {push(value: number): void}; receiver.push(1); receiver.push(2);',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    // rslint always supplies TypeScript type information; the identical type-aware upstream case above/below covers this input.
    // {"code": "function makeSink() { return {push(value: number) {}}; } const sink = makeSink(); sink.push(1); sink.push(2);", "filename": "case.ts", "errors": [{"messageId": "error/array-push", "message": "Do not call `Array#push()` multiple times."}]},
    {
      code: 'const container = {data: {entries: {push(value) { console.log(value); }}}}; container.data.entries.push(1); container.data.entries.push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'const values = []; values.push(1); values.push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
  ],
});

// Argument evaluation
new RuleTester().run('prefer-single-call', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'const array = []; array.push(1); array.push(array.length);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'function f(array: unknown[], value: unknown) { array.push(1); array.push(value); }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'function f(array: unknown[], value: unknown) { array.push(value); array.push(2); }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'const array = [1]; array.push(2); array.push(...array);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'function f(array: unknown[]) { array.push(array = []); array.push(2); }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'function f(array: unknown[]) { const values = { *[Symbol.iterator]() { array = []; yield 1; } }; array.push(...values); array.push(2); }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'function f(array: unknown[]) { const values = { *[Symbol.iterator]() { array = []; yield 1; } }; array.push([...values]); array.push(2); }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
  ],
});

// Array unshift
new RuleTester().run('prefer-single-call', {} as never, {
  valid: [
    {
      code: 'function f(foo: {unshift(value: number): void}) { foo.unshift(1); foo.unshift(2); }',
      filename: 'case.ts',
    },
    { code: 'foo.unshift(1);', filename: 'case.js' },
    { code: 'foo.push(1);\nfoo.unshift(2);', filename: 'case.js' },
    { code: 'foo.unshift(1);\nfoo.push(2);', filename: 'case.js' },
    {
      code: 'foo.unshift(1);; // <- there is an "EmptyStatement" between\nfoo.unshift(2);',
      filename: 'case.js',
    },
    { code: 'foo.unshift(1);\nbar.unshift(2);', filename: 'case.js' },
    { code: 'foo.unshift(1);unshift(2)', filename: 'case.js' },
    { code: 'unshift(1);foo.unshift(2)', filename: 'case.js' },
    { code: 'new foo.unshift(1);foo.unshift(2)', filename: 'case.js' },
    { code: 'foo.unshift(1);new foo.unshift(2)', filename: 'case.js' },
    { code: 'foo[unshift](1);foo.unshift(2)', filename: 'case.js' },
    { code: 'foo.unshift(1);foo[unshift](2)', filename: 'case.js' },
    { code: 'foo.unshift(foo.unshift(1));', filename: 'case.js' },
    {
      code: 'const length = foo.unshift(1);\nfoo.unshift(2);',
      filename: 'case.js',
    },
    {
      code: 'foo.unshift(1);\nconst length = foo.unshift(2);',
      filename: 'case.js',
    },
    { code: 'foo().unshift(1);\nfoo().unshift(2);', filename: 'case.js' },
    {
      code: 'foo().bar.unshift(1);\nfoo().bar.unshift(2);',
      filename: 'case.js',
    },
    {
      code: "const stream = new Readable();\nstream.unshift('one string');\nstream.unshift('another string');",
      filename: 'case.js',
    },
    {
      code: 'class FooReadable extends Readable {\n\tunshiftAndEnd(chunk) {\n\t\tthis.unshift(chunk);\n\t\tthis.unshift(null);\n\t}\n}',
      filename: 'case.js',
    },
    {
      code: 'class Foo {\n\tunshiftAndEnd(chunk) {\n\t\tthis.stream.unshift(chunk);\n\t\tthis.stream.unshift(null);\n\t}\n}',
      filename: 'case.js',
    },
    {
      code: 'process.stdin.unshift(chunk);\nprocess.stdin.unshift(null);',
      filename: 'case.js',
    },
    {
      code: 'process.stdout.unshift(chunk);\nprocess.stdout.unshift(null);',
      filename: 'case.js',
    },
    {
      code: 'process.stderr.unshift(chunk);\nprocess.stderr.unshift(null);',
      filename: 'case.js',
    },
    {
      code: 'foo.unshift(1);\nfoo.unshift(2);\nfoo.bar.unshift(1);\nfoo.bar.unshift(2);',
      filename: 'case.js',
      options: [{ ignore: ['foo.unshift', 'foo.bar.unshift'] }],
    },
    { code: 'for (const _ of []) foo.unshift(bar);', filename: 'case.js' },
    {
      code: 'function bar() {}\nfoo.unshift(bindEvents);',
      filename: 'case.js',
    },
    { code: 'foo.unshift?.(1);\nfoo.unshift?.(2);', filename: 'case.js' },
    { code: 'foo.unshift(1);\nfoo.unshift?.(2);', filename: 'case.js' },
    { code: 'foo.unshift?.(1);\nfoo.unshift(2);', filename: 'case.js' },
    {
      code: 'const foo = new Foo(); foo.unshift(1); foo.unshift(2);',
      filename: 'case.js',
    },
  ],
  invalid: [
    {
      code: 'foo.unshift(1);\nfoo.unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: '(foo.unshift)(1);\n(foo.unshift)(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.bar.unshift(1);\nfoo.bar.unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1);\n(foo).unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'bar()\nfoo.unshift(1);\n(foo).unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift();\nfoo.unshift();',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1);\nfoo.unshift();',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift();\nfoo.unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1, 2);\nfoo.unshift((3), (4));',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1, 2,);\nfoo.unshift(3, 4);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1, 2);\nfoo.unshift(3, 4,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1, 2,);\nfoo.unshift(3, 4,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1, 2, ...a,);\nfoo.unshift(...b,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(bar());\nfoo.unshift(1);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1);\nfoo.unshift(bar());',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(x);\nfoo.unshift(foo.length);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1);\n// Keep this comment\nfoo.unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1,);\nfoo.unshift(2,);\nfoo.unshift(3,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'if (a) {\n\tfoo.unshift(1);\n\tfoo.unshift(2);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'switch (a) {\n\tdefault:\n\t\tfoo.unshift(1);\n\t\tfoo.unshift(2);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'function a() {\n\tfoo.unshift(1);\n\tfoo.unshift(2);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1)\nfoo.unshift(2)\n;[foo].forEach(bar)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: "foo.bar.unshift(1);\n(foo)['bar'].unshift(2);",
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1);\nfoo?.unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.unshift(1);\nfoo?.unshift(2,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo?.unshift(1);\nfoo.unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo?.unshift(1);\nfoo?.unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo?.bar.unshift(1);\nfoo?.bar.unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: '(foo as any[]).unshift(1);\n(foo as any[]).unshift(2);',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'foo!.unshift(1);\nfoo!.unshift(2);',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'function f(foo: number[]) { foo.unshift(1); foo.unshift(2); }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'const container = {data: {entries: {unshift(value) { console.log(value); }}}}; container.data.entries.unshift(1); container.data.entries.unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'const values = []; values.unshift(1); values.unshift(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'const array = []; array.unshift(1); array.unshift(array.length);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
  ],
});

// classList
new RuleTester().run('prefer-single-call', {} as never, {
  valid: [
    {
      code: "foo.classList.toggle('foo');\nfoo.classList.toggle('bar');",
      filename: 'case.js',
    },
    { code: 'foo.classList.add("foo");', filename: 'case.js' },
    {
      code: 'foo.classList.add("foo");\nfoo.classList.remove("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add("foo");; // <- there is an "EmptyStatement" between\nfoo.classList.add("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add("foo");\nbar.classList.add("bar");',
      filename: 'case.js',
    },
    { code: 'foo.classList.add("foo");add("bar")', filename: 'case.js' },
    { code: 'add("foo");foo.classList("bar")', filename: 'case.js' },
    {
      code: 'new foo.classList.add("foo");foo.classList.add("bar")',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add("foo");new foo.classList.add("bar")',
      filename: 'case.js',
    },
    {
      code: 'foo.classList[add]("foo");foo.classList.add("bar")',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add("foo");foo.classList[add]("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add(foo.classList.add("foo"));',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add("foo");\nfoo[classList].add("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add("foo");\nclassList.add("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add("foo");\n(new foo.classList).add("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add("foo");\nfoo.classList.add?.("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.notClassList.add("foo");\nfoo.notClassList.add("bar");',
      filename: 'case.js',
    },
    {
      code: 'classList.add("foo");\nclassList.add("bar");',
      filename: 'case.js',
    },
    {
      code: 'const _ = foo.classList.add("foo");\nfoo.classList.add("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add("foo");\nconst _ = foo.classList.add("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo().classList.add("foo");\nfoo().classList.add("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo().bar.classList.add("foo");\nfoo().bar.classList.add("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList?.add("foo");\nfoo.classList.add("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add("foo");\nfoo.classList?.add("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.add?.("foo");\nfoo.classList.add("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList?.remove("foo");\nfoo.classList.remove("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.remove("foo");\nfoo.classList?.remove("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.remove?.("foo");\nfoo.classList.remove("bar");',
      filename: 'case.js',
    },
    {
      code: 'foo.classList.remove("foo");\nfoo.classList.remove?.("bar");',
      filename: 'case.js',
    },
  ],
  invalid: [
    {
      code: 'foo.classList.add("foo");\nfoo.classList.add("bar");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.remove("foo");\nfoo.classList.remove("bar");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.remove()` multiple times.',
        },
      ],
    },
    {
      code: '(foo.classList.add)("foo");\n(foo.classList.add)("bar");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.bar.classList.add("foo");\nfoo.bar.classList.add("bar");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add("foo");\n(foo).classList.add("bar");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add();\nfoo.classList.add();',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add("foo");\nfoo.classList.add();',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add();\nfoo.classList.add(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add(a, b);\nfoo.classList.add((c), (d));',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add.push(a, b,);\nfoo.classList.add.push(c, d);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add(a, b);\nfoo.classList.add(c, d,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add(a, b,);\nfoo.classList.add(c, d,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add(a, b, ...c,);\nfoo.classList.add(...d,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add(bar());\nfoo.classList.add("foo");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add(a);\nfoo.classList.add(bar());',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add(a,);\nfoo.classList.add(b,);\nfoo.classList.add(c,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'if (a) {\n\tfoo.classList.add(a);\n\tfoo.classList.add(b);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'switch (a) {\n\tdefault:\n\t\tfoo.classList.add(a);\n\t\tfoo.classList.add(b);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'function _() {\n\tfoo.classList.add(a);\n\tfoo.classList.add(b);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add(a)\nfoo.classList.add(b)\n;[foo].forEach(bar)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: "foo.bar.classList.add(a);\n(foo)['bar'].classList.add(b);",
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo?.classList.add("foo");\nfoo.classList.add("bar");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add("foo");\nfoo?.classList.add("bar");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'foo?.classList.add("foo");\nfoo?.classList.add("bar");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
  ],
});

// importScripts
new RuleTester().run('prefer-single-call', {} as never, {
  valid: [
    {
      code: "importScripts('foo.js');\nnotImportScripts('bar.js');",
      filename: 'case.js',
    },
    { code: 'importScripts("foo.js");', filename: 'case.js' },
    {
      code: 'importScripts("foo.js");; // <- there is an "EmptyStatement" between\nimportScripts("bar.js");',
      filename: 'case.js',
    },
    {
      code: 'new importScripts("foo.js");importScripts("bar.js")',
      filename: 'case.js',
    },
    {
      code: 'importScripts("foo.js");new importScripts("bar.js")',
      filename: 'case.js',
    },
    {
      code: 'const _ = importScripts("foo.js");\nimportScripts("bar.js");',
      filename: 'case.js',
    },
    {
      code: 'importScripts("foo.js");\nconst _ = importScripts("bar.js");',
      filename: 'case.js',
    },
    {
      code: 'importScripts("foo.js");\nimportScripts("bar.js");',
      filename: 'case.js',
      options: [{ ignore: ['importScripts'] }],
    },
  ],
  invalid: [
    {
      code: 'importScripts("foo.js");\nimportScripts("bar.js");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: '(importScripts)("foo.js");\n(importScripts)("bar.js");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts();\nimportScripts();',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts("foo.js");\nimportScripts();',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts();\nimportScripts(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts(a, b);\nimportScripts((c), (d));',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts(a, b,);\nimportScripts(c, d);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts(a, b);\nimportScripts(c, d,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts(a, b,);\nimportScripts(c, d,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'foo.classList.add(a, b, ...c,);\nfoo.classList.add(...d,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts(bar());\nimportScripts("foo.js");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts(a);\nimportScripts(bar());',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts(a,);\nimportScripts(b,);\nimportScripts(c,);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'if (a) {\n\timportScripts(a);\n\timportScripts(b);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'switch (a) {\n\tdefault:\n\t\timportScripts(a);\n\t\timportScripts(b);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'function _() {\n\timportScripts(a);\n\timportScripts(b);\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts(a)\nimportScripts(b)\n;[foo].forEach(bar)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts?.("foo.js");\nimportScripts("bar.js");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts("foo.js");\nimportScripts?.("bar.js");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
    {
      code: 'importScripts?.("foo.js");\nimportScripts?.("bar.js");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
  ],
});

// Reference identity
new RuleTester().run('prefer-single-call', {} as never, {
  valid: [
    {
      code: "'1'.someMagicPropertyReturnsAnArray.push(1);\n(1).someMagicPropertyReturnsAnArray.push(2);\n\n/a/i.someMagicPropertyReturnsAnArray.push(1);\n/b/g.someMagicPropertyReturnsAnArray.push(2);\n\n1n.someMagicPropertyReturnsAnArray.push(1);\n2n.someMagicPropertyReturnsAnArray.push(2);\n\n(true).someMagicPropertyReturnsAnArray.push(1);\n(false).someMagicPropertyReturnsAnArray.push(2);",
      filename: 'case.js',
    },
  ],
  invalid: [
    {
      code: "class A extends B {\n\tfoo() {\n\t\tthis.x.push(1);\n\t\tthis.x.push(2);\n\n\t\tsuper.x.push(1);\n\t\tsuper.x.push(2);\n\n\t\t((a?.x).y).push(1);\n\t\t(a.x?.y).push(1);\n\n\t\t((a?.x.y).z).push(1);\n\t\t((a.x?.y).z).push(1);\n\n\t\ta[null].push(1);\n\t\ta['null'].push(1);\n\n\t\t'1'.someMagicPropertyReturnsAnArray.push(1);\n\t\t'1'.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(1);\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t1n.someMagicPropertyReturnsAnArray.push(1);\n\t\t1n.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t(true).someMagicPropertyReturnsAnArray.push(1);\n\t\t(true).someMagicPropertyReturnsAnArray.push(2);\n\t}\n}",
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'a[x].push(1);\na[x].push(2);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
  ],
});

// Receiver reads
new RuleTester().run('prefer-single-call', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'const result = [];\nresult.push("a");\nresult.push(result.length);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'const result = [];\nresult.push("a");\nresult.push(String(result));',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: 'const result = [];\nresult.unshift("a");\nresult.unshift(result.length);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: 'const result = [];\nresult.push("a");\nresult.push(1);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
  ],
});

// Documentation
new RuleTester().run('prefer-single-call', {} as never, {
  valid: [
    { code: '// ✅\nfoo.push(1, 2, 3);', filename: 'case.js' },
    { code: '// ✅\nfoo.unshift(2, 3, 1);', filename: 'case.js' },
    {
      code: "// ✅\nelement.classList.add('foo', 'bar', 'baz');",
      filename: 'case.js',
    },
    {
      code: '// ✅\nimportScripts(\n\t"https://example.com/foo.js",\n\t"https://example.com/bar.js",\n);',
      filename: 'case.js',
    },
    {
      code: '/* eslint unicorn/prefer-single-call: ["error", {"ignore": ["readable.push"]}] */\nimport {Readable} from \'node:stream\';\n\nconst readable = new Readable();\nreadable.push(\'one\');\nreadable.push(\'another\');\nreadable.push(null);\n',
      filename: 'case.js',
      options: [{ ignore: ['readable.push'] }],
    },
  ],
  invalid: [
    {
      code: '// ❌\nfoo.push(1);\nfoo.push(2, 3);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#push()` multiple times.',
        },
      ],
    },
    {
      code: '// ❌\nfoo.unshift(1);\nfoo.unshift(2, 3);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Array#unshift()` multiple times.',
        },
      ],
    },
    {
      code: "// ❌\nelement.classList.add('foo');\nelement.classList.add('bar', 'baz');",
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `Element#classList.add()` multiple times.',
        },
      ],
    },
    {
      code: '// ❌\nimportScripts("https://example.com/foo.js");\nimportScripts("https://example.com/bar.js");',
      filename: 'case.js',
      errors: [
        {
          messageId: 'error/array-push',
          message: 'Do not call `importScripts()` multiple times.',
        },
      ],
    },
  ],
});
