import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-with', {
  valid: [
    // Normal code
    'foo.bar()',
    // "with" as property name
    'obj.with(1)',
    // "with" as method name in object literal
    'var obj = { with: function() {} }; obj.with();',
    // "with" in string literal
    'var s = "with";',
  ],
  invalid: [
    // Basic with statement
    {
      code: 'with(foo) { bar() }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // Single-statement body (no block)
    {
      code: 'with(foo) bar();',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // Nested with inside with — two errors
    {
      code: 'with(a) { with(b) { c() } }',
      errors: [
        { messageId: 'unexpectedWith' },
        { messageId: 'unexpectedWith' },
      ],
    },
    // with inside function
    {
      code: 'function f() { with(obj) { x; } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside arrow function
    {
      code: 'var f = () => { with(obj) { x; } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside if block
    {
      code: 'if (true) { with(obj) { x; } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside for loop
    {
      code: 'for (;;) { with(obj) { x; } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside while loop
    {
      code: 'while (true) { with(obj) { x; } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside try/catch
    {
      code: 'try { with(obj) { x; } } catch(e) {}',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside switch case
    {
      code: 'switch(a) { case 1: with(obj) { x; } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with with member expression
    {
      code: 'with(a.b) { c(); }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with with call expression
    {
      code: 'with(a()) { b(); }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // Multiple sequential with statements — two errors
    {
      code: 'with(a) { x; }\nwith(b) { y; }',
      errors: [
        { messageId: 'unexpectedWith' },
        { messageId: 'unexpectedWith' },
      ],
    },
    // Multi-line with statement
    {
      code: 'with (obj) {\n  foo();\n  bar();\n}',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with with empty body
    {
      code: 'with(obj) {}',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside class method
    {
      code: 'class C { method() { with(obj) { x; } } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside class constructor
    {
      code: 'class C { constructor() { with(obj) { x; } } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside static block
    {
      code: 'class C { static { with(obj) { x; } } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside do...while
    {
      code: 'do { with(obj) { x; } } while(true)',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside for...in
    {
      code: 'for (var k in obj) { with(obj) { x; } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside for...of
    {
      code: 'for (var v of arr) { with(v) { x; } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside else branch
    {
      code: 'if (false) {} else { with(obj) { x; } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside finally block
    {
      code: 'try {} finally { with(obj) { x; } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside catch block
    {
      code: 'try {} catch(e) { with(obj) { x; } }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // with inside labeled statement
    {
      code: 'label: with(obj) { x; }',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // Multi-byte characters (emoji surrogate pair) to verify UTF-16 code unit counting
    {
      code: '/* 🚀 */ with(obj) {}',
      errors: [{ messageId: 'unexpectedWith' }],
    },
    // deeply nested: with inside if inside function inside with
    {
      code: 'with(a) { function f() { if (true) { with(b) { x; } } } }',
      errors: [
        { messageId: 'unexpectedWith' },
        { messageId: 'unexpectedWith' },
      ],
    },
  ],
});
