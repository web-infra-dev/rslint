import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('max-depth', {
  valid: [
    {
      code: 'function foo() { if (true) { if (false) { if (true) { } } } }',
      options: [3] as any,
    },
    {
      code: 'function foo() { if (true) { } else if (false) { } else if (true) { } else if (false) {} }',
      options: [3] as any,
    },
    {
      code: 'var foo = () => { if (true) { if (false) { if (true) { } } } }',
      options: [3] as any,
    },
    'function foo() { if (true) { if (false) { if (true) { } } } }',
    {
      code: 'function foo() { if (true) { if (false) { if (true) { } } } }',
      options: [{ max: 3 }] as any,
    },
    {
      code: 'class C { static { if (1) { if (2) {} } } }',
      options: [2] as any,
    },
    {
      code: 'class C { static { if (1) { if (2) {} } if (1) { if (2) {} } } }',
      options: [2] as any,
    },
    {
      code: 'class C { static { if (1) { if (2) {} } } static { if (1) { if (2) {} } } }',
      options: [2] as any,
    },
    {
      code: 'if (1) { class C { static { if (1) { if (2) {} } } } }',
      options: [2] as any,
    },
    {
      code: 'function foo() { if (1) { class C { static { if (1) { if (2) {} } } } } }',
      options: [2] as any,
    },
    {
      code: 'function foo() { if (1) { if (2) { class C { static { if (1) { if (2) {} } if (1) { if (2) {} } } } } } if (1) { if (2) {} } }',
      options: [2] as any,
    },
    // Class methods / object methods reset depth.
    {
      code: 'class C { method() { if (1) { if (2) { if (3) {} } } } }',
      options: [3] as any,
    },
    {
      code: 'var obj = { method() { if (1) { if (2) { if (3) {} } } } }',
      options: [3] as any,
    },
    // async / generator / async generator function-likes — same boundary.
    {
      code: 'async function f() { if (a) { if (b) { if (c) {} } } }',
      options: [3] as any,
    },
    {
      code: 'function* g() { if (a) { if (b) { if (c) {} } } }',
      options: [3] as any,
    },
    {
      code: 'async function* g() { if (a) { if (b) { if (c) {} } } }',
      options: [3] as any,
    },
    {
      code: 'var f = async () => { if (a) { if (b) { if (c) {} } } }',
      options: [3] as any,
    },
    // Class field arrow.
    {
      code: 'class C { handler = () => { if (a) { if (b) { if (c) {} } } } }',
      options: [3] as any,
    },
    // Type-only constructs add no depth.
    {
      code: 'interface I { x: { y: { z: { w: number } } } } function f() { if (a) {} }',
      options: [1] as any,
    },
    // Empty / degenerate sources.
    '',
    { code: 'function f() {}', options: [0] as any },
    { code: 'class C { static {} }', options: [0] as any },
    {
      code: 'function f() { try { } catch (e) { } finally { } }',
      options: [1] as any,
    },
    // Lock-in: ESLint's asymmetric `IfStatement:exit` pop leaves a residual
    // negative count after else-if chains, giving sibling code an artificial
    // discount. We mirror this exactly.
    {
      code: 'function f() { if (a) {} else if (b) {} else if (c) {} if (d) { if (e) { if (f) {} } } }',
      options: [2] as any,
    },
    // Lone blocks don't increase depth.
    { code: 'function f() { { { { if (a) {} } } } }', options: [1] as any },
    // for-await-of, labeled while, single-statement bodies, namespace.
    {
      code: 'async function f() { for await (const x of y) { if (a) {} } }',
      options: [2] as any,
    },
    {
      code: 'function f() { outer: while (true) { if (a) {} } }',
      options: [2] as any,
    },
    {
      code: 'function f() { while (true) if (a) for (;;) {} }',
      options: [3] as any,
    },
    { code: 'function f() { switch (x) {} }', options: [1] as any },
    {
      code: 'namespace N { if (a) { if (b) {} } }',
      options: [2] as any,
    },
  ],
  invalid: [
    {
      code: 'function foo() { if (true) { if (false) { if (true) { } } } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 43,
        },
      ],
    },
    {
      code: 'var foo = () => { if (true) { if (false) { if (true) { } } } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 44,
        },
      ],
    },
    {
      code: 'function foo() { if (true) {} else { for(;;) {} } }',
      options: [1] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 38,
        },
      ],
    },
    {
      code: 'function foo() { while (true) { if (true) {} } }',
      options: [1] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 33,
        },
      ],
    },
    {
      code: 'function foo() { for (let x of foo) { if (true) {} } }',
      options: [1] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 39,
        },
      ],
    },
    {
      code: 'function foo() { while (true) { if (true) { if (false) { } } } }',
      options: [1] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 33,
        },
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 45,
        },
      ],
    },
    {
      code: 'function foo() { if (true) { if (false) { if (true) { if (false) { if (true) { } } } } } }',
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 68,
        },
      ],
    },
    {
      code: 'function foo() { if (true) { if (false) { if (true) { } } } }',
      options: [{ max: 2 }] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 43,
        },
      ],
    },
    {
      code: 'function foo() { if (a) { if (b) { if (c) { if (d) { if (e) {} } } } } }',
      options: [{}] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 54,
        },
      ],
    },
    {
      code: 'function foo() { if (true) {} }',
      options: [{ max: 0 }] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 18,
        },
      ],
    },
    {
      code: 'class C { static { if (1) { if (2) { if (3) {} } } } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 38,
        },
      ],
    },
    {
      code: 'if (1) { class C { static { if (1) { if (2) { if (3) {} } } } } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 47,
        },
      ],
    },
    {
      code: 'function foo() { if (1) { class C { static { if (1) { if (2) { if (3) {} } } } } } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 64,
        },
      ],
    },
    {
      code: 'function foo() { if (1) { class C { static { if (1) { if (2) {} } } } if (2) { if (3) {} } } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 80,
        },
      ],
    },
    // Top-level (no enclosing function) deep nesting.
    {
      code: 'if (a) { if (b) { if (c) {} } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 19,
        },
      ],
    },
    // IIFE — body has its own scope; interior nesting still triggers.
    {
      code: '(function () { if (a) { if (b) { if (c) {} } } })()',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 34,
        },
      ],
    },
    // Async / generator function bodies.
    {
      code: 'async function f() { if (a) { if (b) { if (c) {} } } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 40,
        },
      ],
    },
    {
      code: 'function* g() { if (a) { if (b) { if (c) {} } } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 35,
        },
      ],
    },
    // Catch / finally body counted within the surrounding try.
    {
      code: 'function f() { try { } catch (e) { if (a) {} } }',
      options: [1] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 36,
        },
      ],
    },
    {
      code: 'function f() { try { } finally { if (a) {} } }',
      options: [1] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 34,
        },
      ],
    },
    // switch case clauses, deep block in a case.
    {
      code: 'function f() { switch (x) { case 1: { for (;;) { if (a) {} } break; } case 2: break; } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 50,
        },
      ],
    },
    // Multiple separate violations in one source.
    {
      code: 'function a() { if (1) { if (2) { if (3) {} } } }\nfunction b() { while (1) { while (2) { while (3) {} } } }',
      options: [2] as any,
      errors: [
        { messageId: 'tooDeeply', line: 1, column: 34 },
        { messageId: 'tooDeeply', line: 2, column: 40 },
      ],
    },
    // Sibling violations under shared parent.
    {
      code: 'function f() { if (1) { if (2) { if (3) {} } if (4) { if (5) {} } } }',
      options: [2] as any,
      errors: [
        { messageId: 'tooDeeply', line: 1, column: 34 },
        { messageId: 'tooDeeply', line: 1, column: 55 },
      ],
    },
    // Legacy `maximum` key.
    {
      code: 'function foo() { if (a) { if (b) { if (c) {} } } }',
      options: [{ maximum: 2 }] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 36,
        },
      ],
    },
    // `else { if (b) {} }` — block alternate, not chained — inner if pushes.
    {
      code: 'function f() { if (a) {} else { if (b) {} } }',
      options: [1] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 33,
        },
      ],
    },
    // Body without braces.
    {
      code: 'function f() { while (true) if (a) for (;;) {} }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 36,
        },
      ],
    },
    // Class expression with method.
    {
      code: 'var C = class { method() { if (a) { if (b) { if (c) {} } } } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 46,
        },
      ],
    },
    // Namespace body shares depth scope with surrounding scope.
    {
      code: 'namespace N { if (a) { if (b) { if (c) {} } } }',
      options: [2] as any,
      errors: [
        {
          messageId: 'tooDeeply',
          line: 1,
          column: 33,
        },
      ],
    },
  ],
});
