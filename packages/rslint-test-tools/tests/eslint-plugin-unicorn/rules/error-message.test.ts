import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string) => ({ code, filename: 'file.js' });

const missingMessage = (constructorName: string) => ({
  messageId: 'missing-message',
  message: `Pass a message to the \`${constructorName}\` constructor.`,
});

const emptyMessage = {
  messageId: 'message-is-empty-string',
  message: 'Error message should not be an empty string.',
};

const notStringMessage = {
  messageId: 'message-is-not-a-string',
  message: 'Error message should be a string.',
};

ruleTester.run('error-message', null as never, {
  valid: [
    // ---- Core Builtin Errors ----
    valid("throw new Error('error')"),
    valid("throw new TypeError('error')"),
    valid("throw new MyCustomError('error')"),
    valid('throw new MyCustomError()'),
    valid('throw generateError()'),
    valid('throw foo()'),
    valid('throw err'),
    valid('throw 1'),
    valid("const err = TypeError('error');\nthrow err;"),
    valid('new Error("message", 0, 0)'),
    valid('new Error(foo)'),
    valid(
      "const errors = [];\nif (condition) {\n\terrors.push('hello');\n}\nif (errors.length) {\n\tthrow new Error(errors.join('\\n'));\n}",
    ),
    valid('new Error(...foo)'),
    valid('/* global x */\nconst a = x;\nthrow x;'),
    valid(
      "const Error = function () {};\nconst err = new Error({\n\tname: 'Unauthorized',\n});",
    ),

    // ---- AggregateError ----
    valid('new AggregateError(errors, "message")'),
    valid('new NotAggregateError(errors)'),
    valid('new AggregateError(...foo)'),
    valid('new AggregateError(...foo, "")'),
    valid('new AggregateError(errors, ...foo)'),
    valid('new AggregateError(errors, message, "")'),
    valid('new AggregateError("", message, "")'),

    // ---- SuppressedError ----
    valid('new SuppressedError(error, suppressed, "message")'),
    valid('new NotSuppressedError(error, suppressed)'),
    valid('new SuppressedError(...foo)'),
    valid('new SuppressedError(...foo, "")'),
    valid('new SuppressedError(error, suppressed, ...foo)'),
    valid('new SuppressedError(error, suppressed, message, "")'),
    valid('new SuppressedError("", "", message, "")'),
  ],
  invalid: [
    // ---- Core Builtin Errors ----
    {
      code: 'throw new Error()',
      filename: 'file.js',
      errors: [missingMessage('Error')],
    },
    {
      code: 'throw Error()',
      filename: 'file.js',
      errors: [missingMessage('Error')],
    },
    {
      code: "throw new Error('')",
      filename: 'file.js',
      errors: [emptyMessage],
    },
    {
      code: 'throw new Error(``)',
      filename: 'file.js',
      errors: [emptyMessage],
    },
    {
      code: 'const err = new Error();\nthrow err;',
      filename: 'file.js',
      errors: [missingMessage('Error')],
    },
    {
      code: 'let err = 1;\nerr = new Error();\nthrow err;',
      filename: 'file.js',
      errors: [missingMessage('Error')],
    },
    {
      code: 'let err = new Error();\nerr = 1;\nthrow err;',
      filename: 'file.js',
      errors: [missingMessage('Error')],
    },
    {
      code: 'const foo = new TypeError()',
      filename: 'file.js',
      errors: [missingMessage('TypeError')],
    },
    {
      code: 'const foo = new SyntaxError()',
      filename: 'file.js',
      errors: [missingMessage('SyntaxError')],
    },
    {
      code: 'const errorMessage = Object.freeze({errorMessage: 1}).errorMessage;\nthrow new Error(errorMessage)',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'throw new Error([])',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'throw new Error([foo])',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'throw new Error([0][0])',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'throw new Error({})',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'throw new Error({foo})',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'throw new Error({foo: 0}.foo)',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'throw new Error(lineNumber=2)',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'throw new Error(false)',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'const error = new RangeError;',
      filename: 'file.js',
      errors: [missingMessage('RangeError')],
    },
    {
      code: 'throw Object.assign(new Error(), {foo})',
      filename: 'file.js',
      errors: [missingMessage('Error')],
    },

    // ---- AggregateError ----
    {
      code: 'new AggregateError(errors)',
      filename: 'file.js',
      errors: [missingMessage('AggregateError')],
    },
    {
      code: 'AggregateError(errors)',
      filename: 'file.js',
      errors: [missingMessage('AggregateError')],
    },
    {
      code: 'new AggregateError(errors, "")',
      filename: 'file.js',
      errors: [emptyMessage],
    },
    {
      code: 'new AggregateError(errors, ``)',
      filename: 'file.js',
      errors: [emptyMessage],
    },
    {
      code: 'new AggregateError(errors, "", extraArgument)',
      filename: 'file.js',
      errors: [emptyMessage],
    },
    {
      code: 'const errorMessage = Object.freeze({errorMessage: 1}).errorMessage;\nthrow new AggregateError(errors, errorMessage)',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new AggregateError(errors, [])',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new AggregateError(errors, [foo])',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new AggregateError(errors, [0][0])',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new AggregateError(errors, {})',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new AggregateError(errors, {foo})',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new AggregateError(errors, {foo: 0}.foo)',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new AggregateError(errors, lineNumber=2)',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'const error = new AggregateError;',
      filename: 'file.js',
      errors: [missingMessage('AggregateError')],
    },

    // ---- SuppressedError ----
    {
      code: 'new SuppressedError(error, suppressed,)',
      filename: 'file.js',
      errors: [missingMessage('SuppressedError')],
    },
    {
      code: 'new SuppressedError(error,)',
      filename: 'file.js',
      errors: [missingMessage('SuppressedError')],
    },
    {
      code: 'new SuppressedError()',
      filename: 'file.js',
      errors: [missingMessage('SuppressedError')],
    },
    {
      code: 'SuppressedError(error, suppressed,)',
      filename: 'file.js',
      errors: [missingMessage('SuppressedError')],
    },
    {
      code: 'SuppressedError(error,)',
      filename: 'file.js',
      errors: [missingMessage('SuppressedError')],
    },
    {
      code: 'SuppressedError()',
      filename: 'file.js',
      errors: [missingMessage('SuppressedError')],
    },
    {
      code: 'new SuppressedError(error, suppressed, "")',
      filename: 'file.js',
      errors: [emptyMessage],
    },
    {
      code: 'new SuppressedError(error, suppressed, ``)',
      filename: 'file.js',
      errors: [emptyMessage],
    },
    {
      code: 'new SuppressedError(error, suppressed, "", options)',
      filename: 'file.js',
      errors: [emptyMessage],
    },
    {
      code: 'const errorMessage = Object.freeze({errorMessage: 1}).errorMessage;\nthrow new SuppressedError(error, suppressed, errorMessage)',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new SuppressedError(error, suppressed, [])',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new SuppressedError(error, suppressed, [foo])',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new SuppressedError(error, suppressed, [0][0])',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new SuppressedError(error, suppressed, {})',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new SuppressedError(error, suppressed, {foo})',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new SuppressedError(error, suppressed, {foo: 0}.foo)',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'new SuppressedError(error, suppressed, lineNumber=2)',
      filename: 'file.js',
      errors: [notStringMessage],
    },
    {
      code: 'const error = new SuppressedError;',
      filename: 'file.js',
      errors: [missingMessage('SuppressedError')],
    },
  ],
});
