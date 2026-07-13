import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-extra-semi', {
  valid: [
    'var x = 5;',
    'function foo(){}',
    'for(;;);',
    'while(0);',
    'do;while(0);',
    'for(a in b);',
    'if(true);',
    'if(true); else;',
    'foo: ;',
    'with(foo);',

    // Class body.
    'class A { }',
    'var A = class { };',
    'class A { a() { this; } }',
    'var A = class { a() { this; } };',
    'class A { } a;',
    'class A { field; }',
    'class A { field = 0; }',
    'class A { static { foo; } }',

    // modules
    {
      code: 'export const x = 42;',
      parserOptions: { sourceType: 'module' } as any,
    },
    {
      code: 'export default 42;',
      parserOptions: { sourceType: 'module' } as any,
    },
  ],
  invalid: [
    {
      code: 'var x = 5;;',
      output: 'var x = 5;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function foo(){};',
      output: 'function foo(){}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'for(;;);;',
      output: 'for(;;);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'while(0);;',
      output: 'while(0);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'do;while(0);;',
      output: 'do;while(0);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'for(a in b);;',
      output: 'for(a in b);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if(true);;',
      output: 'if(true);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if(true){} else;;',
      output: 'if(true){} else;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if(true){;} else {;}',
      output: 'if(true){} else {}',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: 'foo:;;',
      output: 'foo:;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'with(foo);;',
      output: 'with(foo);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'with(foo){;}',
      output: 'with(foo){}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { static { ; } }',
      output: 'class A { static {  } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { static { a;; } }',
      output: 'class A { static { a; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { ; }',
      output: 'class A {  }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { /*a*/; }',
      output: 'class A { /*a*/ }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { ; a() {} }',
      output: 'class A {  a() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { a() {}; }',
      output: 'class A { a() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { a() {}; b() {} }',
      output: 'class A { a() {} b() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A {; a() {}; b() {}; }',
      output: 'class A { a() {} b() {} }',
      errors: [
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
      ],
    },
    {
      code: 'class A { a() {}; get b() {} }',
      output: 'class A { a() {} get b() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { field;; }',
      output: 'class A { field; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { static {}; }',
      output: 'class A { static {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { static { a; }; foo(){} }',
      output: 'class A { static { a; } foo(){} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "; 'use strict'",
      output: "; 'use strict'", // wait, output is null/unchanged
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "; ; 'use strict'",
      output: " ; 'use strict'",
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: "debugger;\n;\n'use strict'",
      output: "debugger;\n;\n'use strict'", // output is null/unchanged
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "function foo() { ; 'bar'; }",
      output: "function foo() { ; 'bar'; }", // output is null/unchanged
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "{ ; 'foo'; }",
      output: "{  'foo'; }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "; ('use strict');",
      output: " ('use strict');",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '; 1;',
      output: ' 1;',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
