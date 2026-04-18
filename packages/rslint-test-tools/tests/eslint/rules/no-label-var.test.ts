import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-label-var', {
  valid: [
    // ---- Upstream ESLint suite ----
    'function bar() { q: for(;;) { break q; } } function foo () { var q = t; }',
    'function bar() { var x = foo; q: for(;;) { break q; } }',

    // ---- Top-level / sibling-scope ----
    'q: for(;;) { break q; }',
    'function bar() { a: for(;;) { b: for(;;) { break b; } } }',
    'function bar(y) { q: for(;;) { break q; } }',
    'for (let i = 0; i < 1; i++) { q: for(;;) { break q; } }',
    'for (const k in obj) { q: for(;;) { break q; } }',
    'for (const v of arr) { q: for(;;) { break q; } }',

    // ---- TS type-only declarations should NOT clash (only values count) ----
    'interface X {} X: for(;;) { break X; }',
    'type X = number; X: for(;;) { break X; }',

    // ---- Nested label inside iteration; outer var has different name ----
    'var x = 1; function f() { y: for(;;) { break y; } }',

    // ---- Catch parameter has different name ----
    'try {} catch (e) { q: for(;;) { break q; } }',

    // ---- Method / arrow / generator scopes don't leak names ----
    'class C { m() { q: for(;;) { break q; } } }',
    'const fn = (a) => { q: for(;;) { break q; } };',
    'function* gen() { q: for(;;) { break q; } }',
  ],
  invalid: [
    // ---- Upstream ESLint suite ----
    {
      code: 'var x = foo; function bar() { x: for(;;) { break x; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 31 }],
    },
    {
      code: 'function bar() { var x = foo; x: for(;;) { break x; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 31 }],
    },
    {
      code: 'function bar(x) { x: for(;;) { break x; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 19 }],
    },

    // ---- Local-binding clashes (strategy A path) ----
    {
      code: 'function bar() { let x = 1; x: for(;;) { break x; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 29 }],
    },
    {
      code: 'function bar() { const x = 1; x: for(;;) { break x; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 31 }],
    },
    {
      code: 'function bar() { function x() {} x: for(;;) { break x; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 34 }],
    },
    {
      code: 'class X {}\nX: for(;;) { break X; }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 2, column: 1 }],
    },
    {
      code: 'function bar({ x }) { x: for(;;) { break x; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 23 }],
    },
    {
      code: 'try {} catch (e) { e: for(;;) { break e; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 20 }],
    },
    {
      code: 'try {} catch ({ a }) { a: for(;;) { break a; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 24 }],
    },
    {
      code: 'for (let i = 0; i < 1; i++) { i: for(;;) { break i; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 31 }],
    },
    {
      code: 'for (let v of arr) { v: for(;;) { break v; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 22 }],
    },
    {
      code: 'for (const k in obj) { k: for(;;) { break k; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 24 }],
    },
    {
      code: 'function f() { x: for(;;) { break x; } { var x = 1; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 16 }],
    },
    {
      code: 'function foo() { foo: for(;;) { break foo; } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 18 }],
    },
    {
      code: '(function fee() { fee: for(;;) { break fee; } })()',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 19 }],
    },
    {
      code: "import x from 'mod'; x: for(;;) { break x; }",
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 22 }],
    },
    {
      code: "import { x } from 'mod'; x: for(;;) { break x; }",
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 26 }],
    },
    {
      code: "import * as x from 'mod'; x: for(;;) { break x; }",
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 27 }],
    },
    {
      code: "import { y as x } from 'mod'; x: for(;;) { break x; }",
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 31 }],
    },
    {
      code: "import type { X } from 'mod'; X: for(;;) { break X; }",
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 31 }],
    },
    {
      code: 'enum X { A } X: for(;;) { break X; }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 14 }],
    },
    {
      code: 'namespace N { export const x = 1; } N: for(;;) { break N; }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 37 }],
    },
    {
      code: 'var a = 1; function f() { a: for(;;) { b: for(;;) { break b; } } }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 27 }],
    },

    // ---- Globals from tsgo lib (strategy B path; requires TypeChecker) ----
    // Use ES standard built-ins so the assertion holds without `lib: ["dom"]`.
    {
      code: 'Promise: for (;;) { break Promise; }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 1 }],
    },
    {
      code: 'Array: for (;;) { break Array; }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 1 }],
    },
    {
      code: 'Math: for (;;) { break Math; }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 1 }],
    },
    {
      code: 'Symbol: for (;;) { break Symbol; }',
      errors: [{ messageId: 'identifierClashWithLabel', line: 1, column: 1 }],
    },
  ],
});
