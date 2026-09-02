import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-sort-props', {} as never, {
  valid: [
    { code: '<App />' },
    { code: '<App {...props} a b />' },
    { code: '<App a z onBar onFoo />', options: [{ callbacksLast: true }] },
    { code: '<App a b="b" />', options: [{ shorthandFirst: true }] },
    { code: '<App a="a" b="b" x y />', options: [{ shorthandLast: true }] },
    {
      code: '<App children={<App />} key={0} ref="r" a />',
      options: [{ reservedFirst: true }],
    },
  ],
  invalid: [
    {
      code: '<App b a />',
      errors: [{ message: 'Props should be sorted alphabetically' }],
    },
    {
      code: '<App a onBar onFoo z />',
      options: [{ callbacksLast: true }],
      errors: [{ message: 'Callbacks must be listed after all other props' }],
    },
    {
      code: '<App a="a" b />',
      options: [{ shorthandFirst: true }],
      errors: [
        { message: 'Shorthand props must be listed before all other props' },
      ],
    },
    {
      code: '<App a key={1} />',
      options: [{ reservedFirst: true }],
      errors: [
        { message: 'Reserved props must be listed before all other props' },
      ],
    },
    {
      code: '<App key={4} />',
      options: [{ reservedFirst: [] }],
      errors: [
        { message: 'A customized reserved first list must not be empty' },
      ],
    },
  ],
});
