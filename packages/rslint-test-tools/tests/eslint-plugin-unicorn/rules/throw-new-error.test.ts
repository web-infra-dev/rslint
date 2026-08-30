import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('throw-new-error', null as never, {
  valid: [
    { code: 'throw new Error()' },
    { code: 'new Error()' },
    { code: 'throw new TypeError()' },
    { code: 'throw new EvalError()' },
    { code: 'throw new RangeError()' },
    { code: 'throw new ReferenceError()' },
    { code: 'throw new SyntaxError()' },
    { code: 'throw new URIError()' },
    { code: 'throw new CustomError()' },
    { code: 'throw new ABCError()' },

    // Not `FooError` like.
    { code: 'throw getError()' },
    // Not a call expression.
    { code: 'throw CustomError' },
    // Callee is neither an identifier nor a static member.
    { code: 'throw getErrorConstructor()()' },
    // `MemberExpression.computed`.
    { code: 'throw lib[Error]()' },
    { code: 'throw lib["Error"]()' },
    // `new` cannot be applied to an optional-chained call.
    { code: 'throw Error?.()' },
    { code: 'throw lib?.Error()' },
    { code: 'throw lib?.foo.Error()' },
    { code: 'throw lib?.foo.Error().message' },
    // Not `FooError` like.
    { code: 'throw lib.getError()' },
    // https://github.com/sindresorhus/eslint-plugin-unicorn/issues/2654
    // `Data.TaggedError` is a factory, not a constructor.
    { code: "class QueryError extends Data.TaggedError('QueryError') {}" },
  ],
  invalid: [
    {
      code: 'throw Error()',
      errors: [
        {
          messageId: 'throw-new-error',
          message: 'Use `new` when creating an error.',
        },
      ],
    },
    {
      code: 'throw (Error)()',
      errors: [
        {
          messageId: 'throw-new-error',
          message: 'Use `new` when creating an error.',
        },
      ],
    },
    {
      code: 'throw lib.Error()',
      errors: [
        {
          messageId: 'throw-new-error',
          message: 'Use `new` when creating an error.',
        },
      ],
    },
    {
      code: 'throw lib.mod.Error()',
      errors: [
        {
          messageId: 'throw-new-error',
          message: 'Use `new` when creating an error.',
        },
      ],
    },
    {
      code: "throw CustomError('foo')",
      errors: [
        {
          messageId: 'throw-new-error',
          message: 'Use `new` when creating an error.',
        },
      ],
    },
    {
      code: 'throw TypeError()',
      errors: [
        {
          messageId: 'throw-new-error',
          message: 'Use `new` when creating an error.',
        },
      ],
    },
    {
      code: 'throw (( URIError ))()',
      errors: [
        {
          messageId: 'throw-new-error',
          message: 'Use `new` when creating an error.',
        },
      ],
    },
    {
      code: 'throw getGlobalThis().Error()',
      errors: [
        {
          messageId: 'throw-new-error',
          message: 'Use `new` when creating an error.',
        },
      ],
    },
    {
      // The rule watches every call expression, not only `throw`.
      code: 'const error = Error()',
      errors: [
        {
          messageId: 'throw-new-error',
          message: 'Use `new` when creating an error.',
        },
      ],
    },
  ],
});
