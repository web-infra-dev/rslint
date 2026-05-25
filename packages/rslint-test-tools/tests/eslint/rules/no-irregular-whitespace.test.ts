import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-irregular-whitespace', {
  valid: [
    // Escaped Unicode in strings (no actual irregular chars)
    `'\\u000B';`,
    `'\\u00A0';`,
    `'\\u3000';`,

    // Actual irregular whitespace inside strings (skipStrings default true)
    "'\u000B';",
    "'\u000C';",
    "'\u0085';",
    "'\u00A0';",
    "'\u180E';",
    "'\uFEFF';",
    "'\u2000';",
    "'\u200B';",
    "'\u202F';",
    "'\u205F';",
    "'\u3000';",

    // skipComments: true
    { code: '// \u000B', options: { skipComments: true } },
    { code: '// \u00A0', options: { skipComments: true } },
    { code: '// \u3000', options: { skipComments: true } },
    { code: '/* \u000B */', options: { skipComments: true } },
    { code: '/* \u00A0 */', options: { skipComments: true } },
    { code: '/* \u3000 */', options: { skipComments: true } },

    // skipRegExps: true
    { code: '/\u000B/', options: { skipRegExps: true } },
    { code: '/\u00A0/', options: { skipRegExps: true } },
    { code: '/\u3000/', options: { skipRegExps: true } },

    // skipTemplates: true
    { code: '`\u000B`', options: { skipTemplates: true } },
    { code: '`\u00A0`', options: { skipTemplates: true } },
    { code: '`\u3000`', options: { skipTemplates: true } },
    { code: '`\u3000${foo}\u3000`', options: { skipTemplates: true } },
    { code: 'const error = ` \u3000 `;', options: { skipTemplates: true } },
    { code: 'const error = `\n\u3000`;', options: { skipTemplates: true } },

    // Unicode BOM at start of file
    '\uFEFFconsole.log("hello BOM");',

    // No irregular whitespace
    'var a = 1;',
  ],
  invalid: [
    // Irregular whitespace in code
    {
      code: "var any \u000B = 'thing';",
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },
    {
      code: "var any \u00A0 = 'thing';",
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },
    {
      code: "var any \uFEFF = 'thing';",
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },
    {
      code: "var any \u3000 = 'thing';",
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },

    // Line separators
    {
      code: "var any \u2028 = 'thing';",
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },
    {
      code: "var any \u2029 = 'thing';",
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },

    // Multiple errors
    {
      code: "var any \u3000 = 'thing', other \u3000 = 'thing';\nvar third \u3000 = 'thing';",
      errors: [
        { messageId: 'noIrregularWhitespace' },
        { messageId: 'noIrregularWhitespace' },
        { messageId: 'noIrregularWhitespace' },
      ],
    },

    // Comments (default skipComments: false)
    {
      code: '// \u000B',
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },
    {
      code: '/* \u00A0 */',
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },

    // Regex (default skipRegExps: false)
    {
      code: 'var any = /\u3000/, other = /\u000B/;',
      errors: [
        { messageId: 'noIrregularWhitespace' },
        { messageId: 'noIrregularWhitespace' },
      ],
    },

    // skipStrings: false
    {
      code: "var any = '\u3000', other = '\u000B';",
      options: { skipStrings: false },
      errors: [
        { messageId: 'noIrregularWhitespace' },
        { messageId: 'noIrregularWhitespace' },
      ],
    },

    // Template literals (default skipTemplates: false)
    {
      code: 'var any = `\u3000`, other = `\u000B`;',
      errors: [
        { messageId: 'noIrregularWhitespace' },
        { messageId: 'noIrregularWhitespace' },
      ],
    },

    // skipTemplates: true but irregular in expression part
    {
      code: '`something ${\u3000 10} another thing`',
      options: { skipTemplates: true },
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },

    // skipTemplates: true but irregular outside template
    {
      code: '\u3000\n`\u3000template`',
      options: { skipTemplates: true },
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },

    // Consecutive irregular chars
    {
      code: 'var foo = \u000B\u000B bar;',
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },

    // Just an irregular char
    {
      code: '\u000B',
      errors: [{ messageId: 'noIrregularWhitespace' }],
    },
  ],
});
