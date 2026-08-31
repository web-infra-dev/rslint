import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const tooDeep = {
  message: 'Do not nest ternary expressions.',
  messageId: 'no-nested-ternary/too-deep',
};

const shouldParen = {
  message: 'Nested ternary expression should be parenthesized.',
  messageId: 'no-nested-ternary/should-parenthesized',
};

const valid = (code: string) => ({ code });

ruleTester.run('no-nested-ternary', null as never, {
  valid: [
    valid('const foo = i > 5 ? true : false;'),
    valid('const foo = i > 5 ? true : (i < 100 ? true : false);'),
    valid('const foo = i > 5 ? (i < 100 ? true : false) : true;'),
    valid(
      'const foo = i > 5 ? (i < 100 ? true : false) : (i < 100 ? true : false);',
    ),
    valid(
      'const foo = i > 5 ? true : (i < 100 ? FOO(i > 50 ? false : true) : false);',
    ),
    // Parenthesized ternary in the test position
    valid('const foo = (a ? b : c) ? d : e;'),
    valid('const foo = (a ? b : c) ? (d ? e : f) : g;'),
    valid('foo ? doBar() : doBaz();'),
    valid('var foo = bar === baz ? qux : quxx;'),
  ],
  invalid: [
    {
      code: 'const foo = i > 5 ? true : (i < 100 ? true : (i < 1000 ? true : false));',
      errors: [tooDeep],
    },
    {
      code: 'const foo = i > 5 ? true : (i < 100 ? (i > 50 ? false : true) : false);',
      errors: [tooDeep],
    },
    {
      code: 'const foo = i > 5 ? i < 100 ? true : false : true;',
      output: 'const foo = i > 5 ? (i < 100 ? true : false) : true;',
      errors: [shouldParen],
    },
    {
      code: 'const foo = i > 5 ? i < 100 ? true : false : i < 100 ? true : false;',
      output:
        'const foo = i > 5 ? (i < 100 ? true : false) : (i < 100 ? true : false);',
      errors: [shouldParen, shouldParen],
    },
    {
      code: 'const foo = i > 5 ? true : i < 100 ? true : false;',
      output: 'const foo = i > 5 ? true : (i < 100 ? true : false);',
      errors: [shouldParen],
    },
    {
      code: 'foo ? bar : baz === qux ? quxx : foobar;',
      output: 'foo ? bar : (baz === qux ? quxx : foobar);',
      errors: [shouldParen],
    },
    {
      code: 'foo ? baz === qux ? quxx : foobar : bar;',
      output: 'foo ? (baz === qux ? quxx : foobar) : bar;',
      errors: [shouldParen],
    },
  ],
});
