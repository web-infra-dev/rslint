import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('consistent-this', {
  valid: [
    'var foo = 42, that = this',
    { code: 'var foo = 42, self = this', options: ['self'] as any },
    { code: 'var self = 42', options: ['that'] as any },
    { code: 'var self', options: ['that'] as any },
    { code: 'var self; self = this', options: ['self'] as any },
    { code: 'var foo, self; self = this', options: ['self'] as any },
    { code: 'var foo, self; foo = 42; self = this', options: ['self'] as any },
    { code: 'self = 42', options: ['that'] as any },
    { code: 'var foo = {}; foo.bar = this', options: ['self'] as any },
    { code: 'var self = this; var vm = this;', options: ['self', 'vm'] as any },
    { code: 'var {foo, bar} = this', options: ['self'] as any },
    { code: '({foo, bar} = this)', options: ['self'] as any },
    { code: 'var [foo, bar] = this', options: ['self'] as any },
    { code: '[foo, bar] = this', options: ['self'] as any },
  ],
  invalid: [
    {
      code: 'var context = this',
      errors: [{ messageId: 'unexpectedAlias', line: 1, column: 5 }],
    },
    {
      code: 'var that = this',
      options: ['self'] as any,
      errors: [{ messageId: 'unexpectedAlias', line: 1, column: 5 }],
    },
    {
      code: 'var foo = 42, self = this',
      options: ['that'] as any,
      errors: [{ messageId: 'unexpectedAlias', line: 1, column: 15 }],
    },
    {
      code: 'var self = 42',
      options: ['self'] as any,
      errors: [{ messageId: 'aliasNotAssignedToThis', line: 1, column: 5 }],
    },
    {
      code: 'var self',
      options: ['self'] as any,
      errors: [{ messageId: 'aliasNotAssignedToThis', line: 1, column: 5 }],
    },
    {
      code: 'var self; self = 42',
      options: ['self'] as any,
      errors: [
        { messageId: 'aliasNotAssignedToThis', line: 1, column: 5 },
        { messageId: 'aliasNotAssignedToThis', line: 1, column: 11 },
      ],
    },
    {
      code: 'context = this',
      options: ['that'] as any,
      errors: [{ messageId: 'unexpectedAlias', line: 1, column: 1 }],
    },
    {
      code: 'that = this',
      options: ['self'] as any,
      errors: [{ messageId: 'unexpectedAlias', line: 1, column: 1 }],
    },
    {
      code: 'self = this',
      options: ['that'] as any,
      errors: [{ messageId: 'unexpectedAlias', line: 1, column: 1 }],
    },
    {
      code: 'self += this',
      options: ['self'] as any,
      errors: [{ messageId: 'aliasNotAssignedToThis', line: 1, column: 1 }],
    },
    {
      code: 'var self; (function() { self = this; }())',
      options: ['self'] as any,
      errors: [{ messageId: 'aliasNotAssignedToThis', line: 1, column: 5 }],
    },
  ],
});
