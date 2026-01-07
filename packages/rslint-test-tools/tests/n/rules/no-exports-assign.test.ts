import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-exports-assign', null as any, {
  valid: [
    { code: 'module.exports.foo = 1' },
    { code: 'exports.bar = 2' },
    { code: 'module.exports = {}' },
    { code: 'module.exports = exports = {}' },
    { code: 'exports = module.exports = {}' },
    { code: 'exports = module.exports' },
    { code: 'var exports = {}' },
    { code: 'let exports = {}' },
    { code: 'const exports = {}' },
    // { code: 'function foo(exports) { exports = {}; }' }, // TODO: shadowing support
  ],
  invalid: [
    {
      code: 'exports = {}',
      errors: [
        {
          messageId: 'noExportsAssign',
          line: 1,
          column: 1,
        },
      ],
    },
    {
      code: 'exports = 1',
      errors: [
        {
          messageId: 'noExportsAssign',
          line: 1,
          column: 1,
        },
      ],
    },
  ],
});
