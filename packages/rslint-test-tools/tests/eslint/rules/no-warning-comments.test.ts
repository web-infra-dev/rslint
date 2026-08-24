import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-warning-comments', {
  valid: [
    { code: '// any comment', options: [{ terms: ['fixme'] }] as any },
    { code: '// any comment', options: [{ terms: ['fixme', 'todo'] }] as any },
    '// any comment',
    { code: '// any comment', options: [{ location: 'anywhere' }] as any },
    {
      code: '// any comment with TODO, FIXME or XXX',
      options: [{ location: 'start' }] as any,
    },
    '// any comment with TODO, FIXME or XXX',
    { code: '/* any block comment */', options: [{ terms: ['fixme'] }] as any },
    {
      code: '/* any block comment */',
      options: [{ terms: ['fixme', 'todo'] }] as any,
    },
    '/* any block comment */',
    {
      code: '/* any block comment */',
      options: [{ location: 'anywhere' }] as any,
    },
    {
      code: '/* any block comment with TODO, FIXME or XXX */',
      options: [{ location: 'start' }] as any,
    },
    '/* any block comment with TODO, FIXME or XXX */',
    "/* any block comment with (TODO, FIXME's or XXX!) */",
    {
      code: '// comments containing terms as substrings like TodoMVC',
      options: [{ terms: ['todo'], location: 'anywhere' }] as any,
    },
    {
      code: "// special regex characters don't cause a problem",
      options: [{ terms: ['[aeiou]'], location: 'anywhere' }] as any,
    },
    '/*eslint no-warning-comments: [2, { "terms": ["todo", "fixme", "any other term"], "location": "anywhere" }]*/\n\nvar x = 10;\n',
    {
      code: '/*eslint no-warning-comments: [2, { "terms": ["todo", "fixme", "any other term"], "location": "anywhere" }]*/\n\nvar x = 10;\n',
      options: [{ location: 'anywhere' }] as any,
    },
    { code: '// foo', options: [{ terms: ['foo-bar'] }] as any },
    '/** multi-line block comment with lines starting with\nTODO\nFIXME or\nXXX\n*/',
    { code: '//!TODO ', options: [{ decoration: ['*'] }] as any },
  ],
  invalid: [
    {
      code: '// fixme',
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// any fixme',
      options: [{ location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// any fixme',
      options: [{ terms: ['fixme'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// any FIXME',
      options: [{ terms: ['fixme'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '/* any fixme */',
      options: [{ terms: ['FIXME'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// any fixme or todo',
      options: [{ terms: ['fixme', 'todo'], location: 'anywhere' }] as any,
      errors: [
        { messageId: 'unexpectedComment', line: 1, column: 1 },
        { messageId: 'unexpectedComment', line: 1, column: 1 },
      ],
    },
    {
      code: '/* fixme and todo */',
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '/* fixme and todo */',
      options: [{ location: 'anywhere' }] as any,
      errors: [
        { messageId: 'unexpectedComment', line: 1, column: 1 },
        { messageId: 'unexpectedComment', line: 1, column: 1 },
      ],
    },
    {
      code: '/* fixme! */',
      options: [{ terms: ['fixme'] }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// regex [litera|$]',
      options: [{ terms: ['[litera|$]'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '/* eslint one-var: 2 */',
      options: [{ terms: ['eslint'] }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '/* eslint one-var: 2 */',
      options: [{ terms: ['one'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '/* any block comment with TODO, FIXME or XXX */',
      options: [{ location: 'anywhere' }] as any,
      errors: [
        { messageId: 'unexpectedComment', line: 1, column: 1 },
        { messageId: 'unexpectedComment', line: 1, column: 1 },
        { messageId: 'unexpectedComment', line: 1, column: 1 },
      ],
    },
    {
      code: "/** \n *any block comment \n*with (TODO, FIXME's or XXX!) **/",
      options: [{ location: 'anywhere' }] as any,
      errors: [
        { messageId: 'unexpectedComment', line: 1, column: 1 },
        { messageId: 'unexpectedComment', line: 1, column: 1 },
        { messageId: 'unexpectedComment', line: 1, column: 1 },
      ],
    },
    {
      code: '// TODO: something small',
      options: [{ location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// TODO: something really longer than 40 characters',
      options: [{ location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '/* TODO: something \n really longer than 40 characters \n and also a new line */',
      options: [{ location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// https://github.com/eslint/eslint/pull/13522#discussion_r470293411 TODO',
      options: [{ location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// Comment ending with term followed by punctuation TODO!',
      options: [{ terms: ['todo'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// Comment ending with term including punctuation TODO!',
      options: [{ terms: ['todo!'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// !TODO comment starting with term preceded by punctuation',
      options: [{ terms: ['todo'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// !TODO comment starting with term including punctuation',
      options: [{ terms: ['!todo'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// FIX!term ending with punctuation followed word character',
      options: [{ terms: ['FIX!'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '// Term starting with punctuation preceded word character!FIX',
      options: [{ terms: ['!FIX'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '//!XXX comment starting with no spaces (anywhere)',
      options: [{ terms: ['!xxx'], location: 'anywhere' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '//!XXX comment starting with no spaces (start)',
      options: [{ terms: ['!xxx'], location: 'start' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '/*\nTODO undecorated multi-line block comment (start)\n*/',
      options: [{ terms: ['todo'], location: 'start' }] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '///// TODO decorated single-line comment with decoration array \n /////',
      options: [
        { terms: ['todo'], location: 'start', decoration: ['*', '/'] },
      ] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
    {
      code: '//**TODO term starts with a decoration character',
      options: [
        { terms: ['*todo'], location: 'start', decoration: ['*'] },
      ] as any,
      errors: [{ messageId: 'unexpectedComment', line: 1, column: 1 }],
    },
  ],
});
