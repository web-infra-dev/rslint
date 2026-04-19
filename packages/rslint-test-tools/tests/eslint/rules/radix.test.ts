import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('radix', {
  valid: [
    // ---- Valid: integer radix literals ----
    'parseInt("10", 10);',
    'parseInt("10", 2);',
    'parseInt("10", 36);',
    'parseInt("10", 0x10);',
    'parseInt("10", 1.6e1);',
    'parseInt("10", 10.0);',

    // ---- Valid: variable radix ----
    'parseInt("10", foo);',
    'Number.parseInt("10", foo);',

    // ---- Valid: non-literal radix / template literal ----
    'parseInt("10", -1);',
    'parseInt("10", x + 1);',
    'parseInt("10", x ? 10 : 16);',
    'parseInt("10", Math.floor(x));',
    'parseInt("10", `10`);',
    'parseInt("10", NaN);',
    'parseInt("10", Infinity);',
    'parseInt("10", (10));',
    'parseInt("10", ((10)));',

    // ---- Valid: numeric edge cases inside [2, 36] ----
    'parseInt("10", 2.0);',
    'parseInt("10", 0b10);',
    'parseInt("10", 0o10);',
    'parseInt("10", 0x24);',

    // ---- Valid: more than 2 arguments ----
    'parseInt("10", 10, extra);',
    'Number.parseInt("10", 10, extra);',

    // ---- Valid: not a call of the tracked functions ----
    'parseInt',
    'Number.foo();',
    'Number[parseInt]();',

    // ---- Valid: private identifier Number.#parseInt ----
    'class C { #parseInt; foo() { Number.#parseInt(); } }',
    'class C { #parseInt; foo() { Number.#parseInt(foo); } }',
    'class C { #parseInt; foo() { Number.#parseInt(foo, "bar"); } }',

    // ---- Valid: shadowing (various declaration kinds) ----
    'var parseInt; parseInt();',
    'var Number; Number.parseInt();',
    'let parseInt = foo; parseInt();',
    'const parseInt = foo; parseInt();',
    'function parseInt() {} parseInt();',
    'function f(parseInt) { parseInt(); }',
    'function f(Number) { Number.parseInt("x", 1); }',
    'import parseInt from "x"; parseInt();',
    'import { parseInt } from "x"; parseInt();',
    'function f() { var parseInt; parseInt(); }',
    'function f() { var Number = {}; Number.parseInt("x", 1); }',
    'function f() { var parseInt = g; function h() { parseInt(); } }',

    // ---- Valid: parseInt used in non-call positions ----
    'const obj = { parseInt: 1 };',
    'const { parseInt } = obj;',
    'obj.parseInt();',
    'foo.parseInt("10");',

    // ---- Valid: Number in non-method-access positions ----
    'const x = Number;',
    'const x = Number(foo);',

    // ---- Valid: NewExpression / TaggedTemplate are not CallExpressions ----
    'new parseInt("10");',
    'parseInt`10`;',

    // ---- Valid: additional shadowing scopes ----
    'try {} catch (parseInt) { parseInt(); }',
    'for (let parseInt = 0; parseInt < 1; parseInt++) {}',
    'for (const parseInt of []) { parseInt(); }',
    'for (const parseInt in {}) { parseInt(); }',
    '{ var parseInt = foo; parseInt(); }',

    // ---- Valid: named function expression / class expression binding ----
    'var fn = function parseInt() { parseInt(); };',
    'var fn = function Number() { Number.parseInt("x", 1); };',
    'const C = class parseInt { static foo() { parseInt(); } };',
    'const C = class Number { static foo() { Number.parseInt("x", 1); } };',

    // ---- Valid: destructuring binds parseInt ----
    'const [parseInt] = arr; parseInt();',
    'const [...parseInt] = arr; parseInt();',
    'const { parseInt = foo } = obj; parseInt();',
    'const { x: parseInt } = obj; parseInt();',

    // ---- Valid: TS declare / export ----
    'declare const parseInt: any; parseInt();',
    'export const parseInt = foo; parseInt();',

    // ---- Valid: Unicode-escaped identifier resolves to parseInt ----
    'var \\u0070arseInt = foo; parseInt();',

    // ---- Valid: TS-only wrappers on callee (ESLint 1:1) ----
    'parseInt!("10");',
    '(parseInt as Function)("10");',
    'Number!.parseInt("10");',
    '(Number as any).parseInt("10");',

    // ---- Valid: labeled statement ----
    'parseInt: for (;;) { parseInt("10", 10); break parseInt; }',

    // ---- Valid: deprecated options "always" / "as-needed" (no behavior change) ----
    { code: 'parseInt("10", 10);', options: ['always'] as any },
    { code: 'parseInt("10", 10);', options: ['as-needed'] as any },
    { code: 'parseInt("10", 8);', options: ['always'] as any },
    { code: 'parseInt("10", 8);', options: ['as-needed'] as any },
    { code: 'parseInt("10", foo);', options: ['always'] as any },
    { code: 'parseInt("10", foo);', options: ['as-needed'] as any },
  ],
  invalid: [
    // ---- Missing parameters ----
    {
      code: 'parseInt();',
      errors: [{ messageId: 'missingParameters', line: 1, column: 1 }],
    },

    // ---- Missing radix ----
    {
      code: 'parseInt("10");',
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10",);',
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt((0, "10"));',
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt((0, "10"),);',
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },

    // ---- Invalid radix literals ----
    {
      code: 'parseInt("10", null);',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10", undefined);',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10", true);',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10", "foo");',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10", "123");',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10", 1);',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10", 37);',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10", 10.5);',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },

    // ---- Number.parseInt ----
    {
      code: 'Number.parseInt();',
      errors: [{ messageId: 'missingParameters', line: 1, column: 1 }],
    },
    {
      code: 'Number.parseInt("10");',
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },
    {
      code: 'Number.parseInt("10", 1);',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'Number.parseInt("10", 37);',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'Number.parseInt("10", 10.5);',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },

    // ---- Optional chaining ----
    {
      code: 'parseInt?.("10");',
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },
    {
      code: 'Number.parseInt?.("10");',
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },
    {
      code: 'Number?.parseInt("10");',
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },
    {
      code: '(Number?.parseInt)("10");',
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },

    // ---- Deprecated options still trigger ----
    {
      code: 'parseInt();',
      options: ['always'] as any,
      errors: [{ messageId: 'missingParameters', line: 1, column: 1 }],
    },
    {
      code: 'parseInt();',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'missingParameters', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10");',
      options: ['always'] as any,
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10");',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10", 1);',
      options: ['always'] as any,
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10", 1);',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'Number.parseInt();',
      options: ['always'] as any,
      errors: [{ messageId: 'missingParameters', line: 1, column: 1 }],
    },
    {
      code: 'Number.parseInt();',
      options: ['as-needed'] as any,
      errors: [{ messageId: 'missingParameters', line: 1, column: 1 }],
    },

    // ---- Numeric literal edge cases ----
    {
      code: 'parseInt("10", 0);',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },
    {
      code: 'parseInt("10", 0x25);',
      errors: [{ messageId: 'invalidRadix', line: 1, column: 1 }],
    },

    // ---- Multiple errors per file ----
    {
      code: "parseInt();\nparseInt('10');\nparseInt('10', 1);",
      errors: [
        { messageId: 'missingParameters', line: 1, column: 1 },
        { messageId: 'missingRadix', line: 2, column: 1 },
        { messageId: 'invalidRadix', line: 3, column: 1 },
      ],
    },

    // ---- Nested parseInt calls ----
    {
      code: 'parseInt(parseInt("10"));',
      errors: [
        { messageId: 'missingRadix', line: 1, column: 1 },
        { messageId: 'missingRadix', line: 1, column: 10 },
      ],
    },

    // ---- Class / arrow / IIFE contexts ----
    {
      code: "class C { foo() { parseInt('x'); } }",
      errors: [{ messageId: 'missingRadix', line: 1, column: 19 }],
    },
    {
      code: "class C { x = parseInt('x'); }",
      errors: [{ messageId: 'missingRadix', line: 1, column: 15 }],
    },
    {
      code: "class C { static { parseInt('x'); } }",
      errors: [{ messageId: 'missingRadix', line: 1, column: 20 }],
    },
    {
      code: "const f = () => parseInt('x');",
      errors: [{ messageId: 'missingRadix', line: 1, column: 17 }],
    },
    {
      code: "(function () { parseInt('x'); })();",
      errors: [{ messageId: 'missingRadix', line: 1, column: 16 }],
    },

    // ---- Shadow scope does not reach ----
    {
      code: 'function f() { parseInt(); }',
      errors: [{ messageId: 'missingParameters', line: 1, column: 16 }],
    },
    {
      code: "parseInt('x');\nfunction f() { var parseInt; parseInt(); }",
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },

    // ---- Suggestion preserves internal comments ----
    {
      code: 'parseInt("10" /* hi */);',
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },

    // ---- Spread argument ----
    {
      code: 'parseInt(...args);',
      errors: [{ messageId: 'missingRadix', line: 1, column: 1 }],
    },

    // ---- Real-world containers ----
    {
      code: 'async function f() { return parseInt("x"); }',
      errors: [{ messageId: 'missingRadix', line: 1, column: 29 }],
    },
    {
      code: 'function* g() { yield parseInt("x"); }',
      errors: [{ messageId: 'missingRadix', line: 1, column: 23 }],
    },
    {
      code: 'const obj = { m() { parseInt("x"); } };',
      errors: [{ messageId: 'missingRadix', line: 1, column: 21 }],
    },
    {
      code: 'throw parseInt("x");',
      errors: [{ messageId: 'missingRadix', line: 1, column: 7 }],
    },
    {
      code: 'const x = parseInt(`${foo}`);',
      errors: [{ messageId: 'missingRadix', line: 1, column: 11 }],
    },
    {
      code: 'class C { [parseInt("x")]() {} }',
      errors: [{ messageId: 'missingRadix', line: 1, column: 12 }],
    },
    {
      code: 'function f(x = parseInt("x")) { return x; }',
      errors: [{ messageId: 'missingRadix', line: 1, column: 16 }],
    },
    {
      code: 'export default parseInt("x");',
      errors: [{ messageId: 'missingRadix', line: 1, column: 16 }],
    },
  ],
});
