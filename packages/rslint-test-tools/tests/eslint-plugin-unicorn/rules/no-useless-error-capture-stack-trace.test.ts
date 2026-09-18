// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import { RuleTester, type ValidTestCase } from '../rule-tester';

const defaults = {
  filename: 'src/virtual.js',
  languageOptions: { sourceType: 'module' },
} satisfies Pick<ValidTestCase, 'filename' | 'languageOptions'>;

const messages = {
  error: {
    messageId: 'no-useless-error-capture-stack-trace/error',
    message: 'Unnecessary `Error.captureStackTrace(…)` call.',
  },
};

new RuleTester().run('no-useless-error-capture-stack-trace', {} as never, {
  valid: [
    {
      ...defaults,
      code: 'class MyError {constructor() {Error.captureStackTrace(this, MyError)}}',
    },
    {
      ...defaults,
      code: 'class MyError extends NotABuiltinError {constructor() {Error.captureStackTrace(this, MyError)}}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(not_this, MyError)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, NotClassName)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError, ...extraArguments)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(..._, MyError)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, ..._)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(...[this, MyError])\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tNotError.captureStackTrace(this, MyError)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.not_captureStackTrace(this, MyError)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tnew Error.captureStackTrace(this, MyError)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError?.captureStackTrace(this, MyError)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, this?.constructor)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, this.notConstructor)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, import.meta)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tfunction foo() {\n\tError.captureStackTrace(this, MyError)\n}\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tnotConstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tfunction foo() {\n\t\t\tError.captureStackTrace(this, MyError)\n\t\t}\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor(MyError) {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tstatic {\n\t\tError.captureStackTrace(this, MyError)\n\n\t\tfunction foo() {\n\t\t\tError.captureStackTrace(this, MyError)\n\t\t}\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tclass NotAErrorSubclass {\n\t\t\tconstructor() {\n\t\t\t\tError.captureStackTrace(this, new.target)\n\t\t\t}\n\t\t}\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class Error {}\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class Error {}\nclass MyError extends RangeError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor(): void;\n\tstatic {\n\t\tError.captureStackTrace(this, MyError)\n\n\t\tfunction foo() {\n\t\t\tError.captureStackTrace(this, MyError)\n\t\t}\n\t}\n}',
      filename: 'src/virtual.ts',
    },
  ],
  invalid: [
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError);\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, MyError);\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, this.constructor);\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, this.constructor);\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, new.target);\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, new.target);\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends EvalError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends RangeError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends ReferenceError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends SyntaxError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends TypeError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends URIError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends AggregateError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends SuppressedError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tconst foo = () => {\n\t\t\tError.captureStackTrace(this, MyError)\n\t\t}\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tif (a) Error.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tconst x = () => Error.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'class MyError extends Error {\n\tconstructor() {\n\t\tvoid Error.captureStackTrace(this, MyError)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'export default class extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, new.target)\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'export default (\n\tclass extends Error {\n\t\tconstructor() {\n\t\t\tError.captureStackTrace(this, new.target)\n\t\t}\n\t}\n)',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: '// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError);\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: '// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, MyError);\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: '// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, this.constructor);\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: '// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, this.constructor);\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: '// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, new.target);\n\t}\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: '// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, new.target);\n\t}\n}',
      errors: [messages.error],
    },
  ],
});
