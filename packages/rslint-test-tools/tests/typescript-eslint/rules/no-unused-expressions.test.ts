import { RuleTester } from '@typescript-eslint/rule-tester';



const ruleTester = new RuleTester();

ruleTester.run('no-unused-expressions', {
  valid: [
    `
      test.age?.toLocaleString();
    `,
    `
      let a = (a?.b).c;
    `,
    `
      let b = a?.['b'];
    `,
    `
      let c = one[2]?.[3][4];
    `,
    `
      one[2]?.[3][4]?.();
    `,
    `
      a?.['b']?.c();
    `,
    `
      module Foo {
        'use strict';
      }
    `,
    `
      namespace Foo {
        'use strict';

        export class Foo {}
        export class Bar {}
      }
    `,
    `
      function foo() {
        'use strict';

        return null;
      }
    `,
    `
      import('./foo');
    `,
    `
      import('./foo').then(() => {});
    `,
    `
      class Foo<T> {}
      new Foo<string>();
    `,
    {
      code: 'foo && foo?.();',
      options: [{ allowShortCircuit: true }],
    },
    {
      code: "foo && import('./foo');",
      options: [{ allowShortCircuit: true }],
    },
    {
      code: "foo ? import('./foo') : import('./bar');",
      options: [{ allowTernary: true }],
    },

    // Side-effect expressions: assignments, update, delete, void, yield, await
    'a = b',
    'new a',
    'i++',
    'i--',
    '--i',
    '++i',
    'a += 1',
    'a &&= b',
    'a ||= b',
    'a ??= b',
    'delete foo.bar',
    'void new C',
    'function* foo(){ yield 0; }',
    'function* foo(){ yield; }',
    'async function foo() { await 5; }',

    // TS assertions wrapping calls (inner expression has side effects)
    'foo() as any;',
    'foo()!;',
    '<any>foo();',
    'foo() satisfies string;',

    // TS non-null assertion wrapping call in short-circuit
    { code: 'foo && foo()!;', options: [{ allowShortCircuit: true }] },

    // Instantiation expression wrapping call (has side effects)
    'declare function getSet(): Set<unknown>; getSet()<string>();',

    // Combined allowShortCircuit + allowTernary
    {
      code: 'a ? b && c() : d()',
      options: [{ allowShortCircuit: true, allowTernary: true }],
    },
    {
      code: 'a ?? b()',
      options: [{ allowShortCircuit: true }],
    },
    {
      code: 'a || b()',
      options: [{ allowShortCircuit: true }],
    },

    // Directives: multiple, different strings, arrow body
    '"use strict"; "use asm"; f();',
    'var foo = () => {"use strict"; return true; }',

    // String in variable declaration is not a directive (but also not a standalone expression)
    'function foo() { var foo = "use strict"; return true; }',

    // Deep nesting: TS wrappers + options combined
    {
      code: `
declare const foo: Function | undefined;
<any>(foo && foo()!)
      `,
      options: [{ allowShortCircuit: true }],
    },
    {
      code: '(Foo && Foo())<string, number>;',
      options: [{ allowShortCircuit: true }],
    },
    {
      code: 'a ? (b && c()) : (d || e())',
      options: [{ allowShortCircuit: true, allowTernary: true }],
    },

    // Chained optional calls
    'a?.()?.b();',

    // satisfies is NOT unwrapped (defaults to not disallowed)
    'declare const x: string; x satisfies string;',
    'foo() satisfies string;',

    // Class static block: leading string treated as directive
    "class C { static { 'use strict'; } }",
  ],
  invalid: [
    {
      code: 'if (0) 0;',
      errors: [
        {
          column: 8,
          endColumn: 10,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: 'f(0), {};',
      errors: [
        {
          column: 1,
          endColumn: 10,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: 'a, b();',
      errors: [
        {
          column: 1,
          endColumn: 8,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: `
a() &&
  function namedFunctionInExpressionContext() {
    f();
  };
      `,
      errors: [
        {
          column: 1,
          endColumn: 5,
          endLine: 5,
          line: 2,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: 'a?.b;',
      errors: [
        {
          column: 1,
          endColumn: 6,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: '(a?.b).c;',
      errors: [
        {
          column: 1,
          endColumn: 10,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: "a?.['b'];",
      errors: [
        {
          column: 1,
          endColumn: 10,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: "(a?.['b']).c;",
      errors: [
        {
          column: 1,
          endColumn: 14,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: 'a?.b()?.c;',
      errors: [
        {
          column: 1,
          endColumn: 11,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: '(a?.b()).c;',
      errors: [
        {
          column: 1,
          endColumn: 12,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: 'one[2]?.[3][4];',
      errors: [
        {
          column: 1,
          endColumn: 16,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: 'one.two?.three.four;',
      errors: [
        {
          column: 1,
          endColumn: 21,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: `
module Foo {
  const foo = true;
  'use strict';
}
      `,
      errors: [
        {
          column: 3,
          endColumn: 16,
          endLine: 4,
          line: 4,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: `
namespace Foo {
  export class Foo {}
  export class Bar {}

  'use strict';
}
      `,
      errors: [
        {
          column: 3,
          endColumn: 16,
          endLine: 6,
          line: 6,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: `
function foo() {
  const foo = true;

  ('use strict');
}
      `,
      errors: [
        {
          column: 3,
          endColumn: 18,
          endLine: 5,
          line: 5,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: 'foo && foo?.bar;',
      errors: [
        {
          column: 1,
          endColumn: 17,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
      options: [{ allowShortCircuit: true }],
    },
    {
      code: 'foo ? foo?.bar : bar.baz;',
      errors: [
        {
          column: 1,
          endColumn: 26,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
      options: [{ allowTernary: true }],
    },
    {
      code: `
class Foo<T> {}
Foo<string>;
      `,
      errors: [
        {
          column: 1,
          endColumn: 13,
          endLine: 3,
          line: 3,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: 'Map<string, string>;',
      errors: [
        {
          column: 1,
          endColumn: 21,
          endLine: 1,
          line: 1,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: `
declare const foo: number | undefined;
foo;
      `,
      errors: [
        {
          column: 1,
          endColumn: 5,
          endLine: 3,
          line: 3,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: `
declare const foo: number | undefined;
foo as any;
      `,
      errors: [
        {
          column: 1,
          endColumn: 12,
          endLine: 3,
          line: 3,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: `
declare const foo: number | undefined;
<any>foo;
      `,
      errors: [
        {
          column: 1,
          endColumn: 10,
          endLine: 3,
          line: 3,
          messageId: 'unusedExpression',
        },
      ],
    },
    {
      code: `
declare const foo: number | undefined;
foo!;
      `,
      errors: [
        {
          column: 1,
          endColumn: 6,
          endLine: 3,
          line: 3,
          messageId: 'unusedExpression',
        },
      ],
    },

    // Literals: boolean, null, bigint, regex
    {
      code: 'true;',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'false;',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'null;',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '1n;',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '/regex/;',
      errors: [{ messageId: 'unusedExpression' }],
    },

    // Unary without side effects: -, ~, typeof
    {
      code: '-a;',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '~a;',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'typeof foo;',
      errors: [{ messageId: 'unusedExpression' }],
    },

    // this / super as standalone
    {
      code: 'function foo() { this; }',
      errors: [{ messageId: 'unusedExpression' }],
    },

    // Nested parenthesized expressions
    {
      code: '((a));',
      errors: [{ messageId: 'unusedExpression' }],
    },

    // Arithmetic / comparison / bitwise binary expressions
    {
      code: 'a + b;',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'a === b;',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'a & b;',
      errors: [{ messageId: 'unusedExpression' }],
    },

    // NOTE: satisfies is NOT unwrapped (matches @typescript-eslint behavior).
    // `foo satisfies T;` is NOT flagged — see valid cases.

    // Nested TS assertions wrapping non-call
    {
      code: 'declare const foo: any; (foo as any)!;',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'declare const foo: any; (foo as any) as number;',
      errors: [{ messageId: 'unusedExpression' }],
    },

    // Comma expression (both sides are calls, but comma itself is disallowed)
    {
      code: 'f(), g();',
      errors: [{ messageId: 'unusedExpression' }],
    },

    // Ternary: one branch valid, one invalid
    {
      code: 'a ? b() || (c = d) : e',
      errors: [{ messageId: 'unusedExpression' }],
    },

    // Class static block: leading strings are directives, non-leading flagged
    {
      code: "class C { static { const x = 1; 'foo'; 'bar'; } }",
      errors: [
        { messageId: 'unusedExpression' },
        { messageId: 'unusedExpression' },
      ],
    },

    // allowShortCircuit: right side must be valid
    {
      code: 'a && (b ?? c);',
      options: [{ allowShortCircuit: true }],
      errors: [{ messageId: 'unusedExpression' }],
    },

    // allowTernary: both branches must be valid
    {
      code: 'a ? b() : c;',
      options: [{ allowTernary: true }],
      errors: [{ messageId: 'unusedExpression' }],
    },

    // Element access (no side effects)
    {
      code: "a['b'];",
      errors: [{ messageId: 'unusedExpression' }],
    },

    // Template literal with interpolation
    {
      code: '`hello ${world}`;',
      errors: [{ messageId: 'unusedExpression' }],
    },

    // Deep nesting invalid: instantiation > short-circuit > identifier (no call)
    {
      code: '(Foo && Foo)<string, number>;',
      options: [{ allowShortCircuit: true }],
      errors: [{ messageId: 'unusedExpression' }],
    },
    // Deeply nested ternary + short-circuit: one branch invalid
    {
      code: 'a ? (b && c()) : (d || e)',
      options: [{ allowShortCircuit: true, allowTernary: true }],
      errors: [{ messageId: 'unusedExpression' }],
    },
    // Chained optional member access after call (result is member access)
    {
      code: 'a?.()?.b;',
      errors: [{ messageId: 'unusedExpression' }],
    },
    // Arrow/function/class expression as statement
    {
      code: '(() => {});',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '(function() {});',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '(class {});',
      errors: [{ messageId: 'unusedExpression' }],
    },
    // Object literal as statement
    {
      code: '({a: 1});',
      errors: [{ messageId: 'unusedExpression' }],
    },
  ],
});
