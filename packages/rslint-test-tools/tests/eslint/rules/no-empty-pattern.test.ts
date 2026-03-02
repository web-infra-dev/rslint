import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-empty-pattern', {
  valid: [
    'var {a} = obj;',
    'var [a] = arr;',
    'var {a = 1} = obj;',
    'var [a = 1] = arr;',
    'function foo({a}) {}',
    'function foo([a]) {}',
    'var {a: {b}} = obj;',
    'var {a: [b]} = obj;',
    // allowObjectPatternsAsParameters: direct parameter
    {
      code: 'function foo({}) {}',
      options: { allowObjectPatternsAsParameters: true },
    },
    // allowObjectPatternsAsParameters: parameter with empty object default
    {
      code: 'function foo({} = {}) {}',
      options: { allowObjectPatternsAsParameters: true },
    },
  ],
  invalid: [
    {
      code: 'var {} = obj;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var [] = arr;',
      errors: [{ messageId: 'unexpected' }],
    },
    // Nested empty patterns
    {
      code: 'var {a: {}} = obj;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var {a: []} = obj;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function foo({}) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function foo([]) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    // allowObjectPatternsAsParameters: non-empty default should still report
    {
      code: 'function foo({} = {a: 1}) {}',
      options: { allowObjectPatternsAsParameters: true },
      errors: [{ messageId: 'unexpected' }],
    },
    // allowObjectPatternsAsParameters: non-object default should still report
    {
      code: 'function foo({} = bar) {}',
      options: { allowObjectPatternsAsParameters: true },
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
