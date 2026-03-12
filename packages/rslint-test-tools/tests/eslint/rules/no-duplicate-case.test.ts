import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-duplicate-case', {
  valid: [
    'switch (a) { case 1: break; case 2: break; }',
    'switch (a) { case 1: break; case "1": break; }',
    'switch (a) { case 1: break; default: break; }',
    'switch (a) { case "a": break; case "b": break; }',
    'switch (a) { case a: break; case b: break; }',
    // String literals with comment-like content should not be corrupted
    'switch (a) { case "http://example.com": break; case "other": break; }',
    'switch (a) { case "a /* b */": break; case "a": break; }',
    // String literals with different whitespace should not be collapsed
    'switch (a) { case "hello  world": break; case "hello world": break; }',
    // Regex literals should not be confused with comments
    'switch (a) { case /foo/: break; case /bar/: break; }',
    'switch (a) { case /foo/i: break; case /bar/g: break; }',
  ],
  invalid: [
    {
      code: 'switch (a) { case 1: break; case 1: break; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'switch (a) { case "a": break; case "a": break; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'switch (a) { case 1: break; case 2: break; case 1: break; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'switch (a) { case a: break; case a: break; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Regex containing /* in character class should not be confused with comment
    {
      code: 'switch (a) { case /[/*]/: break; case /[/*]/: break; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Comments outside strings should still be stripped for comparison
    {
      code: 'switch (a) { case /*a*/ 1: break; case 1: break; }',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
