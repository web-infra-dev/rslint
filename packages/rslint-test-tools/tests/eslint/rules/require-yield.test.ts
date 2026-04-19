import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-yield', {
  valid: [
    // ---- Upstream ESLint suite ----
    `function foo() { return 0; }`,
    `function* foo() { yield 0; }`,
    `function* foo() { }`,
    `(function* foo() { yield 0; })();`,
    `(function* foo() { })();`,
    `var obj = { *foo() { yield 0; } };`,
    `var obj = { *foo() { } };`,
    `class A { *foo() { yield 0; } };`,
    `class A { *foo() { } };`,

    // ---- Yield variants ----
    `function* foo() { yield; }`,
    `function* foo() { yield* [1, 2, 3]; }`,

    // ---- Async generators ----
    `async function* foo() { yield 0; }`,
    `async function* foo() { }`,

    // ---- Yield inside control flow ----
    `function* foo() { if (true) { yield 1; } }`,
    `function* foo() { while (true) { yield 1; } }`,
    `function* foo() { try { yield 1; } catch (e) {} }`,
    `function* foo() { try { } catch (e) { yield 1; } }`,
    `function* foo() { try { } finally { yield 1; } }`,

    // ---- Export forms ----
    `export function* foo() { yield 1; }`,
    `export default function*() { yield 1; }`,

    // ---- Class method forms ----
    `class A { static *foo() { yield 0; } }`,
    `class A { *#foo() { yield 0; } }`,

    // ---- Computed / class expression / anonymous ----
    `var obj = { *['foo']() { yield 0; } };`,
    `const A = class { *foo() { yield 0; } };`,
    `(function*() { yield 0; })();`,

    // ---- Arrow inside generator ----
    `function* foo() { const x = () => 1; yield x; }`,

    // ---- Nested generators, both have yield ----
    `function* outer() { function* inner() { yield 1; } yield 2; }`,

    // ---- TS modifiers on class generator methods ----
    `class A { async *foo() { yield 0; } }`,
    `class A { public *foo() { yield 0; } }`,
    `class A { private *foo() { yield 0; } }`,
    `class A { protected *foo() { yield 0; } }`,
    `class A { public static async *foo() { yield 0; } }`,

    // ---- FE as object property value ----
    `var obj = { foo: function*() { yield 0; } };`,

    // ---- Control flow variants ----
    `function* foo() { for (const x of [1]) yield x; }`,
    `function* foo(x: number) { switch (x) { case 1: yield 1; break; default: yield 0; } }`,
    `function* foo() { do { yield 1; } while (false); }`,

    // ---- Generic generator ----
    `function* foo<T>(x: T): Generator<T> { yield x; }`,

    // ---- Nested FE inside arrow inside generator ----
    `function* foo() { const f = () => function*() { yield 1; }; yield 2; }`,

    // ---- Overload signatures ----
    `function* foo(x: string): Generator<string>; function* foo(x: number): Generator<number>; function* foo(x: any): Generator<any> { yield x; }`,

    // ---- Ambient declarations (no body) ----
    `declare function* foo(): Generator<number>;`,

    // ---- Class field with generator FE ----
    `class A { foo = function*() { yield 0; }; }`,

    // ---- Multi-line function head ----
    'function*\n  foo() {\n  yield 0;\n}',

    // ---- Decorator scenarios (with yield) ----
    `function dec(t: any, k: any, d: any) {} class A { @dec *foo() { yield 0; } }`,
    `function d1(t: any, k: any, d: any) {} function d2(t: any, k: any, d: any) {} class A { @d1 @d2 *foo() { yield 0; } }`,
    `function dec() { return (t: any, k: any, d: any) => {}; } class A { @dec() *foo() { yield 0; } }`,
    `function dec(t: any, k: any, d: any) {} class A { @dec public static *foo() { yield 0; } }`,
    `function dec(t: any, k: any, d: any) {} class A { @dec async *foo() { yield 0; } }`,
    `function classDec(t: any) {} @classDec class A { *foo() { yield 0; } }`,
    `function dec(t: any, k: any) {} class A { @dec foo = function*() { yield 0; }; }`,
    'function dec(t: any, k: any, d: any) {}\nclass A {\n  @dec\n  *foo() { yield 0; }\n}',

    // ---- JSDoc / comment scenarios (with yield) ----
    `/** doc */ function* foo() { yield 0; }`,
    '/**\n * doc\n */\nfunction* foo() { yield 0; }',
    `class A { /** doc */ *foo() { yield 0; } }`,
    `var o = { /** doc */ *foo() { yield 0; } };`,
    `var o = { /** doc */ foo: function*() { yield 0; } };`,
    `(/** doc */ function*() { yield 0; })();`,
    '// comment\nfunction* foo() { yield 0; }',
    `/* comment */ function* foo() { yield 0; }`,
    `class A { /** doc */ foo = function*() { yield 0; }; }`,

    // ---- Yield attribution boundary scenarios ----
    `function* foo() { const o = { [yield 1]() { return 0; } }; return 0; }`,
    `function* foo() { class C { [yield 1]() { return 0; } } return 0; }`,
    `function* foo() { class C extends (yield 1) {} return 0; }`,
    `function* foo() { yield yield 1; }`,
  ],
  invalid: [
    // ---- Upstream ESLint suite ----
    {
      code: `function* foo() { return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `(function* foo() { return 0; })();`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: `var obj = { *foo() { return 0; } }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: `class A { *foo() { return 0; } }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: `function* foo() { function* bar() { yield 0; } }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `function* foo() { function* bar() { return 0; } yield 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },

    // ---- Async generator ----
    {
      code: `async function* foo() { return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ---- Export forms ----
    {
      code: `export function* foo() { return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: `export default function*() { return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },

    // ---- Class method forms ----
    {
      code: `class A { static *foo() { return 0; } }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: `class A { *#foo() { return 0; } }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },

    // ---- Computed / class expression / anonymous FE ----
    {
      code: `var obj = { *['foo']() { return 0; } }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: `const A = class { *foo() { return 0; } };`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: `(function*() { return 0; })();`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },

    // ---- Control flow without yield in any branch ----
    {
      code: `function* foo() { if (true) { return 1; } else { return 2; } }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },

    // ---- Arrow inside generator ----
    {
      code: `function* foo() { const x = () => 1; return x; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },

    // ---- Multi-line ----
    {
      code: 'function* foo() {\n  return 0;\n}',
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },

    // ---- Async class method generator ----
    {
      code: `class A { async *foo() { return 0; } }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },

    // ---- Public/static combined ----
    {
      code: `class A { public static async *foo() { return 0; } }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },

    // ---- FE as object property value ----
    {
      code: `var obj = { foo: function*() { return 0; } };`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },

    // ---- Generic generator ----
    {
      code: `function* foo<T>(x: T): Generator<T> { return x; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },

    // ---- Overload signatures + impl with no yield ----
    {
      code: `function* foo(x: string): Generator<string>; function* foo(x: number): Generator<number>; function* foo(x: any): Generator<any> { return x; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 91,
          endLine: 1,
          endColumn: 104,
        },
      ],
    },

    // ---- Nested FE inside arrow: outer has no yield ----
    {
      code: `function* foo() { const f = () => function*() { yield 1; }; return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },

    // ---- Class field with generator FE ----
    {
      code: `class A { foo = function*() { return 0; }; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },

    // ---- Multi-line function head ----
    {
      code: 'function*\n  foo() {\n  return 0;\n}',
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 6,
        },
      ],
    },

    // ---- Decorator scenarios ----
    {
      code: 'declare function dec(t: any, k: any, d: any): void;\nclass A { @dec *foo() { return 0; } }',
      errors: [
        {
          messageId: 'missingYield',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'declare function dec(): (t: any, k: any, d: any) => void;\nclass A { @dec() *foo() { return 0; } }',
      errors: [
        {
          messageId: 'missingYield',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'declare function d1(t: any, k: any, d: any): void; declare function d2(t: any, k: any, d: any): void;\nclass A { @d1 @d2 *foo() { return 0; } }',
      errors: [
        {
          messageId: 'missingYield',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'declare function dec(t: any, k: any, d: any): void;\nclass A { @dec public static *foo() { return 0; } }',
      errors: [
        {
          messageId: 'missingYield',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'declare function dec(t: any, k: any, d: any): void;\nclass A { @dec async *foo() { return 0; } }',
      errors: [
        {
          messageId: 'missingYield',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'declare function classDec(t: any): void;\n@classDec\nclass A { *foo() { return 0; } }',
      errors: [
        {
          messageId: 'missingYield',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'declare function dec(t: any, k: any): void;\nclass A { @dec foo = function*() { return 0; }; }',
      errors: [
        {
          messageId: 'missingYield',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'declare function dec(t: any, k: any, d: any): void;\nclass A {\n  @dec\n  *foo() { return 0; }\n}',
      errors: [
        {
          messageId: 'missingYield',
          line: 3,
          column: 3,
          endLine: 4,
          endColumn: 7,
        },
      ],
    },

    // ---- JSDoc / comment scenarios ----
    {
      code: `/** doc */ function* foo() { return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '/**\n * doc\n */\nfunction* foo() { return 0; }',
      errors: [
        {
          messageId: 'missingYield',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `class A { /** doc */ *foo() { return 0; } }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: `var o = { /** doc */ *foo() { return 0; } };`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: `var o = { /** doc */ foo: function*() { return 0; } };`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: `(/** doc */ function*() { return 0; })();`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '// comment\nfunction* foo() { return 0; }',
      errors: [
        {
          messageId: 'missingYield',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 14,
        },
      ],
    },
    {
      code: `/* comment */ function* foo() { return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: `class A { /** doc */ foo = function*() { return 0; }; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },

    // ---- Illegal yield in nested non-generator scope must NOT rescue outer generator ----
    {
      code: `function* foo() { function inner() { yield 1; } return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `function* foo() { const f = function() { yield 1; }; return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `function* foo() { const f = () => { yield 1; }; return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `function* foo() { const f = () => yield 1; return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `function* foo() { class C { m() { yield 1; } } return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `function* foo() { const o = { get x() { yield 1; return 0; } }; return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `function* foo() { const o = { set x(v) { yield 1; } }; return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `function* foo() { class C { constructor() { yield 1; } } return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `function* foo() { class C { x = yield 1; } return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: `function* foo() { class C { m() { function* g() { yield 1; } } } return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },

    // ---- Illegal yield in parameter default values (nested non-gen) ----
    {
      code: `function* outer() { function bar(x = yield 1) { return x; } return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `function* outer() { const f = (x = yield 1) => x; return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `function* outer() { class C { m(x = yield 1) { return x; } } return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `function* outer() { const o = { set x(v = yield 1) {} }; return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `function* outer() { class C { constructor(x = yield 1) {} } return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },

    // ---- Class static block (illegal yield) ----
    {
      code: `function* outer() { class C { static { yield 1; } } return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `function* outer() { class C { static { function* g() { yield 1; } yield 2; } } return 0; }`,
      errors: [
        {
          messageId: 'missingYield',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
  ],
});
