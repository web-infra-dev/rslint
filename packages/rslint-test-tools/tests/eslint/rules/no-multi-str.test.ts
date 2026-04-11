import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-multi-str', {
  valid: [
    // ---- Single-line strings ----
    "var a = 'Line 1 Line 2';",
    'var a = "double quoted";',
    "var a = '';",
    'var a = "";',

    // ---- Escape sequences (not real newlines in raw source) ----
    "var a = 'hello\\nworld';",
    'var a = "hello\\nworld";',
    "var a = 'hello\\rworld';",
    "var a = 'hello\\u2028world';",
    "var a = 'hello\\u2029world';",

    // ---- Template literals can span multiple lines ----
    'var a = `Line 1\nLine 2`;',
    'var a = `\n`;',

    // ---- String concatenation across lines ----
    "var a = 'Line 1' +\n'Line 2';",

    // ---- Non-string expressions ----
    'var a = 42;',
    'var a = true;',
  ],
  invalid: [
    // ================================================================
    // Line break types
    // ================================================================
    {
      code: "var x = 'Line 1 \\\n Line 2'",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "'foo\\\rbar';",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "'foo\\\r\nbar';",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "'foo\\\u2028bar';",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "'foo\\\u2029ar';",
      errors: [{ messageId: 'multilineString' }],
    },

    // ================================================================
    // Quote types
    // ================================================================
    {
      code: "'foo\\\nbar';",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: '"foo\\\nbar";',
      errors: [{ messageId: 'multilineString' }],
    },

    // ================================================================
    // Nesting contexts
    // ================================================================
    {
      code: "test('Line 1 \\\n Line 2');",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "var obj = { key: 'foo\\\nbar' };",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "var obj = { ['foo\\\nbar']: 1 };",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "var arr = ['foo\\\nbar'];",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "function f() { return 'foo\\\nbar'; }",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "var a = x ? 'foo\\\nbar' : 'safe';",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "function f(x = 'foo\\\nbar') {}",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "class A { prop = 'foo\\\nbar' }",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "const f = () => 'foo\\\nbar';",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "var a = 'foo\\\nbar' + x;",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "var a = x || 'foo\\\nbar';",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "var a = x ?? 'foo\\\nbar';",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "x = 'foo\\\nbar';",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "throw 'foo\\\nbar';",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "switch (x) { case 'foo\\\nbar': break; }",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "var a = obj['foo\\\nbar'];",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "var a = ('foo\\\nbar');",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "fn1(fn2('foo\\\nbar'));",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "var obj = { a: { b: 'foo\\\nbar' } };",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "var a = `${'foo\\\nbar'}`;",
      errors: [{ messageId: 'multilineString' }],
    },

    // ================================================================
    // TypeScript-specific contexts
    // ================================================================
    {
      code: "enum E { A = 'foo\\\nbar' }",
      errors: [{ messageId: 'multilineString' }],
    },
    {
      code: "type T = 'foo\\\nbar';",
      errors: [{ messageId: 'multilineString' }],
    },

    // ================================================================
    // Edge cases
    // ================================================================
    // 3+ lines
    {
      code: "'line1\\\nline2\\\nline3';",
      errors: [{ messageId: 'multilineString' }],
    },
    // Empty continuation
    {
      code: "'\\\n';",
      errors: [{ messageId: 'multilineString' }],
    },
    // String on a later line
    {
      code: "var x;\nvar y = 'foo\\\nbar';",
      errors: [{ messageId: 'multilineString' }],
    },

    // ================================================================
    // Multiple errors
    // ================================================================
    {
      code: "'foo\\\nbar';\n'baz\\\nqux';",
      errors: [
        { messageId: 'multilineString' },
        { messageId: 'multilineString' },
      ],
    },
    {
      code: "fn('one\\\ntwo', 'three\\\nfour');",
      errors: [
        { messageId: 'multilineString' },
        { messageId: 'multilineString' },
      ],
    },
  ],
});
