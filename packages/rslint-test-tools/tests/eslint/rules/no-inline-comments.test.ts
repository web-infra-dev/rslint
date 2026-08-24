import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-inline-comments', {
  valid: [
    '// A valid comment before code\nvar a = 1;',
    'var a = 2;\n// A valid comment after code',
    '// A solitary comment',
    'var a = 1; // eslint-disable-line no-debugger',
    'var a = 1; /* eslint-disable-line no-debugger */',
    'foo(); /* global foo */',
    'foo(); /* globals foo */',
    'var foo; /* exported foo */',

    // JSX exception
    {
      code: 'var a = (\n            <div>\n            {/*comment*/}\n            </div>\n        )',
      filename: 'src/virtual.tsx',
    },
    {
      code: 'var a = (\n            <div>\n            { /* comment */ }\n            <h1>Some heading</h1>\n            </div>\n        )',
      filename: 'src/virtual.tsx',
    },
    {
      code: 'var a = (\n            <div>\n            {// comment\n            }\n            </div>\n        )',
      filename: 'src/virtual.tsx',
    },
    {
      code: 'var a = (\n            <div>\n            { // comment\n            }\n            </div>\n        )',
      filename: 'src/virtual.tsx',
    },
    {
      code: 'var a = (\n            <div>\n            {/* comment 1 */\n            /* comment 2 */}\n            </div>\n        )',
      filename: 'src/virtual.tsx',
    },
    {
      code: 'var a = (\n            <div>\n            {/*\n              * comment 1\n              */\n             /*\n              * comment 2\n              */}\n            </div>\n        )',
      filename: 'src/virtual.tsx',
    },
    {
      code: 'var a = (\n            <div>\n            {/*\n               multi\n               line\n               comment\n            */}\n            </div>\n        )',
      filename: 'src/virtual.tsx',
    },
    {
      code: `import(/* webpackChunkName: "my-chunk-name" */ './locale/en');`,
      options: { ignorePattern: '(?:webpackChunkName):\\s.+' },
    },
    {
      code: 'var foo = 2; // Note: This comment is legal.',
      options: { ignorePattern: 'Note: ' },
    },
  ],

  invalid: [
    {
      code: 'var a = 1; /*A block comment inline after code*/',
      errors: [{ messageId: 'unexpectedInlineComment', line: 1, column: 12 }],
    },
    {
      code: '/*A block comment inline before code*/ var a = 2;',
      errors: [{ messageId: 'unexpectedInlineComment', line: 1, column: 1 }],
    },
    {
      code: '/* something */ var a = 2;',
      options: { ignorePattern: 'otherthing' },
      errors: [{ messageId: 'unexpectedInlineComment', line: 1, column: 1 }],
    },
    {
      code: 'var a = 3; //A comment inline with code',
      errors: [{ messageId: 'unexpectedInlineComment', line: 1, column: 12 }],
    },
    {
      code: 'var a = 3; // someday use eslint-disable-line here',
      errors: [{ messageId: 'unexpectedInlineComment', line: 1, column: 12 }],
    },
    {
      code: 'var a = 3; // other line comment',
      options: { ignorePattern: 'something' },
      errors: [{ messageId: 'unexpectedInlineComment', line: 1, column: 12 }],
    },
    {
      code: 'var a = 4;\n/**A\n * block\n * comment\n * inline\n * between\n * code*/ var foo = a;',
      errors: [{ messageId: 'unexpectedInlineComment', line: 2, column: 1 }],
    },
    {
      code: 'var a = \n{/**/}',
      errors: [{ messageId: 'unexpectedInlineComment', line: 2, column: 2 }],
    },

    // JSX
    {
      code: 'var a = (\n                <div>{/* comment */}</div>\n            )',
      filename: 'src/virtual.tsx',
      errors: [{ messageId: 'unexpectedInlineComment', line: 2, column: 23 }],
    },
    {
      code: 'var a = (\n                <div>{// comment\n                }\n                </div>\n            )',
      filename: 'src/virtual.tsx',
      errors: [{ messageId: 'unexpectedInlineComment', line: 2, column: 23 }],
    },
    {
      code: 'var a = (\n                <div>{/* comment */\n                }\n                </div>\n            )',
      filename: 'src/virtual.tsx',
      errors: [{ messageId: 'unexpectedInlineComment', line: 2, column: 23 }],
    },
    {
      code: 'var a = (\n                <div>\n                { /* two comments on the same line... */ /* ...are not allowed, same as with a non-JSX code */}\n                </div>\n            )',
      filename: 'src/virtual.tsx',
      errors: [
        { messageId: 'unexpectedInlineComment', line: 3, column: 19 },
        { messageId: 'unexpectedInlineComment', line: 3, column: 58 },
      ],
    },
    {
      code: 'var a = (\n                <div>\n                {\n                    /* overlapping\n                    */ /*\n                       lines */\n                }\n                </div>\n            )',
      filename: 'src/virtual.tsx',
      errors: [
        { messageId: 'unexpectedInlineComment', line: 4, column: 21 },
        { messageId: 'unexpectedInlineComment', line: 5, column: 24 },
      ],
    },
  ],
});
