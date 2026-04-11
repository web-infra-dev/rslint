import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-script-url', {
  valid: [
    // Basic non-matching strings
    "var a = 'Hello World!';",
    'var a = 10;',
    "var a = '';",

    // Near misses
    "var url = 'xjavascript:'",
    'var url = `xjavascript:`',
    "var a = 'javascript';",
    "var a = ' javascript:';",
    "var a = 'about:blank';",
    "var a = 'mailto:user@example.com';",
    'var a = `https://example.com`;',

    // Template literals with substitutions — static value unknown
    'var url = `${foo}javascript:`',
    'var url = `javascript:${foo}`',
    'var url = `${a}javascript:${b}`',

    // Tagged templates — tag controls interpretation
    'var a = foo`javaScript:`;',
    'var a = obj.tag`javascript:`;',
    "var a = obj['tag']`javascript:`;",
    'var a = tag()`javascript:`;',
    // Nested tagged template — inner is also tagged
    'tag`${foo`javascript:`}`;',

    // String concatenation — individual parts don't start with "javascript:"
    "var a = 'java' + 'script:';",
  ],
  invalid: [
    // Basic string literals
    {
      code: "var a = 'javascript:void(0);';",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "var a = 'javascript:';",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: 'var a = "JavaScript:void(0)";',
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },

    // Case-insensitivity
    {
      code: "var a = 'JAVASCRIPT:';",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "var a = 'jAvAsCrIpT:void(0)';",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },

    // Template literals
    {
      code: 'var a = `javascript:`;',
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: 'var a = `JavaScript:`;',
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },

    // Nesting contexts
    {
      code: "location.href = 'javascript:void(0)';",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: 'location.href = `javascript:void(0)`;',
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "var obj = { href: 'javascript:void(0)' };",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "var arr = ['javascript:'];",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "fn('javascript:');",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "function f() { return 'javascript:'; }",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "var a = x ? 'javascript:' : 'safe';",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "function f(url = 'javascript:') {}",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "class A { url = 'javascript:' }",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "enum E { A = 'javascript:' }",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },

    // Multiple errors in one statement
    {
      code: "a('javascript:', 'javascript:void(0)');",
      errors: [
        { messageId: 'unexpectedScriptURL' },
        { messageId: 'unexpectedScriptURL' },
      ],
    },

    // String literal inside tagged template — only the outer template
    // is tagged; the inner string literal is still a script URL
    {
      code: "tag`${'javascript:'}`;",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },

    // Template literal inside tagged template — parent is TemplateSpan,
    // NOT TaggedTemplateExpression, so it should trigger
    {
      code: 'tag`${`javascript:`}`;',
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },

    // Arrow function implicit return
    {
      code: "const f = () => 'javascript:';",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },

    // Logical / nullish operators
    {
      code: "var a = x ?? 'javascript:';",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "var a = x || 'javascript:';",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },

    // Computed property / element access
    {
      code: "var a = obj['javascript:'];",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
    {
      code: "var obj = { ['javascript:']: 1 };",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },

    // TypeScript type position — string literal types are also checked
    {
      code: "type T = 'javascript:';",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },

    // Binary expression — matching side still triggers
    {
      code: "var a = 'javascript:' + path;",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },

    // Multi-line
    {
      code: "var a =\n  'javascript:';",
      errors: [{ messageId: 'unexpectedScriptURL' }],
    },
  ],
});
