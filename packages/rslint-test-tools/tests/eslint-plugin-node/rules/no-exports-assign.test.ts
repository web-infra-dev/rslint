import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester({
  languageOptions: {
    globals: { exports: 'writable', module: 'readonly' },
  },
});

const forbidden = {
  messageId: 'forbidden',
  message:
    "Unexpected assignment to 'exports' variable. Use 'module.exports' instead.",
  line: 1,
  column: 1,
  endLine: 1,
  endColumn: 13,
};

// Every case from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-exports-assign.js
ruleTester.run(
  'no-exports-assign',
  {},
  {
    valid: [
      { name: 'module property', code: 'module.exports.foo = 1' },
      { name: 'exports property', code: 'exports.bar = 1' },
      {
        name: 'outer module assignment',
        code: 'module.exports = exports = {}',
      },
      {
        name: 'inner module assignment',
        code: 'exports = module.exports = {}',
      },
      {
        name: 'shadowed parameter',
        code: 'function f(exports) { exports = {} }',
      },
      // Correct example from the documentation at the same tag.
      {
        name: 'documentation: correct',
        code: `module.exports.foo = 1
exports.bar = 2

module.exports = {}

// allows \`exports = {}\` if along with \`module.exports =\`
module.exports = exports = {}
exports = module.exports = {}`,
      },
    ],
    invalid: [
      { name: 'global assignment', code: 'exports = {}', errors: [forbidden] },
      // Both incorrect examples from the documentation at the same tag.
      {
        name: 'documentation: object is not exported',
        code: `// This assigned object is not exported.
// You need to use \`module.exports = { ... }\`.
exports = {
    foo: 1
}`,
        errors: [{ ...forbidden, line: 3, endLine: 5, endColumn: 2 }],
      },
      {
        name: 'documentation: incorrect',
        code: 'exports = {}',
        errors: [forbidden],
      },
    ],
  },
);
