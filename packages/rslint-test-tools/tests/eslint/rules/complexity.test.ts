import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('complexity', {
  valid: [
    'function a(x) {}',
    { code: 'function b(x) {}', options: [1] as any },
    {
      code: 'function a(x) {if (true) {return x;}}',
      options: [2] as any,
    },
    {
      code: 'function a(x) {if (true) {return x;} else {return x+1;}}',
      options: [2] as any,
    },
    {
      code: 'function a(x) {if (true) {return x;} else if (false) {return x+1;} else {return 4;}}',
      options: [3] as any,
    },
    {
      code: 'function a(x) {for(var i = 0; i < 5; i ++) {x ++;} return x;}',
      options: [2] as any,
    },
    {
      code: 'function a(x) {try {x.getThis();} catch (e) {x.getThat();}}',
      options: [2] as any,
    },
    {
      code: 'function a(x) {return x === 4 ? 3 : 5;}',
      options: [2] as any,
    },
    { code: 'function a(x) {return x || 4;}', options: [2] as any },
    { code: 'function a(x) {x && 4;}', options: [2] as any },
    { code: 'function a(x) {x ?? 4;}', options: [2] as any },
    { code: 'function a(x) {x ||= 4;}', options: [2] as any },
    { code: 'function a(x) {x &&= 4;}', options: [2] as any },
    { code: 'function a(x) {x ??= 4;}', options: [2] as any },
    { code: 'function a(x) {x = 4;}', options: [1] as any },
    { code: 'function a(x) {x |= 4;}', options: [1] as any },
    {
      code: 'function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: 3;}}',
      options: [3] as any,
    },
    { code: 'function a(x) {while(true) {"foo";}}', options: [2] as any },
    { code: 'function a(x) {do {"foo";} while (true)}', options: [2] as any },
    { code: 'if (foo) { bar(); }', options: [3] as any },

    // Modified complexity.
    {
      code: 'function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: 3;}}',
      options: [{ max: 2, variant: 'modified' }] as any,
    },

    // Class fields.
    {
      code: 'function foo() { class C { x = a || b; y = c || d; } }',
      options: [2] as any,
    },
    {
      code: 'class C { x = a || b; y() { c || d; } z = e || f; }',
      options: [2] as any,
    },
    {
      code: 'class C { x = (() => { a || b }) || (() => { c || d }) }',
      options: [2] as any,
    },
    {
      code: 'class C { x; y = a; static z; static q = b; }',
      options: [1] as any,
    },

    // Class static blocks.
    {
      code: 'class C { static { a || b; } static { c || d; } }',
      options: [2] as any,
    },
    { code: 'class C { static { a } }', options: [1] as any },
    {
      code: 'class C { static { a || b; c || d; } }',
      options: [3] as any,
    },

    // Object property options.
    { code: 'function b(x) {}', options: [{ max: 1 }] as any },

    // Optional chaining.
    {
      code: 'function a(b) { b?.c; }',
      options: [{ max: 2 }] as any,
    },

    // Default parameter / destructuring values.
    { code: "function a(b = '') {}", options: [{ max: 2 }] as any },
    {
      code: "function a(b) { const { c = '' } = b; }",
      options: [{ max: 2 }] as any,
    },
    {
      code: "function a(b) { const [ c = '' ] = b; }",
      options: [{ max: 2 }] as any,
    },

    // Empty / degenerate sources.
    'function f() {}',
    'class C { static {} }',

    // async / generator function bodies.
    { code: 'async function f() { a || b; }', options: [2] as any },
    { code: 'function* g() { a || b; }', options: [2] as any },
    { code: 'async function* g() { a || b; }', options: [2] as any },
  ],
  invalid: [
    {
      code: 'function a(x) {}',
      options: [0] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: "function foo(x) {if (x > 10) {return 'x is greater than 10';} else if (x > 5) {return 'x is greater than 5';} else {return 'x is less than 5';}}",
      options: [2] as any,
      errors: [
        {
          messageId: 'complex',
          line: 1,
          column: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var func = function () {}',
      options: [0] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: 'var obj = { a(x) {} }',
      options: [0] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: 'class Test { a(x) {} }',
      options: [0] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: 'function a(x) {if (true) {return x;}}',
      options: [1] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: 'function a(x) {return x || 4;}',
      options: [1] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: 'function a(x) {x ||= 4;}',
      options: [1] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: 'var obj = { a(x) { return x ? 0 : 1; } };',
      options: [1] as any,
      errors: [{ messageId: 'complex' }],
    },

    // Modified complexity.
    {
      code: 'function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: 3;}}',
      options: [{ max: 1, variant: 'modified' }] as any,
      errors: [{ messageId: 'complex' }],
    },

    // Class field initializer.
    {
      code: 'class C { x = a || b; }',
      options: [1] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: 'class C { x = a || b; y = b || c || d; z = e || f; }',
      options: [2] as any,
      errors: [
        {
          messageId: 'complex',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    // Class field arrow value — classified as "method 'x'".
    {
      code: 'class C { x = () => a || b || c; }',
      options: [2] as any,
      errors: [{ messageId: 'complex' }],
    },

    // Class static block.
    {
      code: 'class C { static { a || b; }  }',
      options: [1] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: 'class C { static { a || b || c || d; } static { e || f || g; } }',
      options: [3] as any,
      errors: [
        {
          messageId: 'complex',
          column: 11,
          endColumn: 17,
        },
      ],
    },

    // Optional chaining.
    {
      code: 'function a(b) { b?.c; }',
      options: [{ max: 1 }] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: 'function a(b) { b?.c?.d; }',
      options: [{ max: 2 }] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: 'function a(b) { b?.c?.()?.(); }',
      options: [{ max: 3 }] as any,
      errors: [{ messageId: 'complex' }],
    },

    // Default parameter / destructuring values.
    {
      code: "function a(b = '') {}",
      options: [{ max: 1 }] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: "function a(b) { const { c = '' } = b; }",
      options: [{ max: 1 }] as any,
      errors: [{ messageId: 'complex' }],
    },
    {
      code: "function a(b) { const [ { c: d = '' } = {} ] = b; }",
      options: [{ max: 1 }] as any,
      errors: [{ messageId: 'complex' }],
    },

    // Multi-line.
    {
      code: 'function foo() {\n    if (a) {\n        if (b) {\n            if (c) {\n            }\n        }\n    }\n}',
      options: [2] as any,
      errors: [
        {
          messageId: 'complex',
          line: 1,
          column: 1,
        },
      ],
    },
  ],
});
