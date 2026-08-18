import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-div-regex', {
  valid: [
    "var f = function() { return /foo/ig.test('bar'); };",
    'var f = function() { return /\\=foo/; };',
  ],
  invalid: [
    {
      code: 'var f = function() { return /=foo/; };',
      errors: [{ messageId: 'unexpected', line: 1, column: 29 }],
    },
  ],
});
