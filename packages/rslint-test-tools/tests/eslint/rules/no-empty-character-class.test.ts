import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-empty-character-class', {
  valid: [
    'var foo = /^abc[a-zA-Z]/;',
    'var regExp = new RegExp("^abc[]");',
    'var foo = /^abc/;',
    String.raw`var foo = /[\[]/;`,
    String.raw`var foo = /[\]]/;`,
    'var foo = /[^]/;',
    String.raw`var foo = /\[]/`,
    'var foo = /[[]/;',
    // v-flag valid cases (ES2024 unicodeSets)
    'var foo = /[[^]]/v;',
    String.raw`var foo = /[[\]]]/v;`,
    String.raw`var foo = /[[\[]]/v;`,
    'var foo = /[a--b]/v;',
    'var foo = /[a&&b]/v;',
    'var foo = /[[a][b]]/v;',
  ],
  invalid: [
    {
      code: 'var foo = /^abc[]/;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var foo = /foo[]bar/;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var foo = /[]]/;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`var foo = /\[[]/;`,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: String.raw`var foo = /\[\[\]a-z[]/;`,
      errors: [{ messageId: 'unexpected' }],
    },
    // v-flag invalid cases (ES2024 unicodeSets)
    {
      code: 'var foo = /[]/v;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var foo = /[[]]/v;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var foo = /[[a][]]/v;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var foo = /[a[[b[]c]]d]/v;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var foo = /[a--[]]/v;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var foo = /[[]--b]/v;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var foo = /[a&&[]]/v;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var foo = /[[]&&b]/v;',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
