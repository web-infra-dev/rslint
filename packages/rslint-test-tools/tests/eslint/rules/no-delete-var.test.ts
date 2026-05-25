import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-delete-var', {
  valid: [
    // Property access
    'delete x.prop;',
    'delete foo.bar.baz;',
    // Element access
    'delete obj["key"];',
    'delete obj[0];',
    // Optional chaining
    'delete a?.b;',
    // Computed property
    'delete obj[`key`];',
    // Parenthesized member expression (inner is not Identifier)
    'delete (x.prop);',
    'delete ((obj["key"]));',
    // TypeScript type assertion wrapping identifier (ESLint sees TSAsExpression, not Identifier)
    'delete (x as any);',
    // TypeScript non-null assertion wrapping identifier
    'delete x!;',
    // TypeScript angle-bracket assertion
    'delete (<any>x);',
    // TypeScript satisfies expression
    'delete (x satisfies any);',
    // Comma expression — inner result is not a plain Identifier in ESTree
    'delete (0, x);',
  ],
  invalid: [
    // Basic
    {
      code: 'delete x',
      errors: [{ messageId: 'unexpected' }],
    },
    // With var declaration
    {
      code: 'var x; delete x;',
      errors: [{ messageId: 'unexpected' }],
    },
    // With let declaration
    {
      code: 'let y = 1; delete y;',
      errors: [{ messageId: 'unexpected' }],
    },
    // Parenthesized identifier
    {
      code: 'delete (x);',
      errors: [{ messageId: 'unexpected' }],
    },
    // Double parenthesized
    {
      code: 'delete ((x));',
      errors: [{ messageId: 'unexpected' }],
    },
    // Nested in if
    {
      code: 'var x; if (true) { delete x; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Nested in function
    {
      code: 'function f() { delete x; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Nested in arrow function
    {
      code: 'const f = () => { delete x; };',
      errors: [{ messageId: 'unexpected' }],
    },
    // In for loop
    {
      code: 'for (;;) { delete x; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In try-catch
    {
      code: 'try { delete x; } catch(e) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    // In class method
    {
      code: 'class C { m() { delete x; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Multiple violations
    {
      code: 'delete x; delete y;',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    // Special identifier names
    {
      code: 'delete arguments;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'delete NaN;',
      errors: [{ messageId: 'unexpected' }],
    },
    // Parenthesized + nested context
    {
      code: 'if (true) { delete (x); }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In switch case
    {
      code: 'switch(0) { case 0: delete x; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In while
    {
      code: 'while (false) { delete x; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In async function
    {
      code: 'async function f() { delete x; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In generator function
    {
      code: 'function* g() { delete x; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Multi-line
    {
      code: 'delete\nx',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
