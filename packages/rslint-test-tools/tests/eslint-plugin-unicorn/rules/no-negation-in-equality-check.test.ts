// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import { RuleTester, type ValidTestCase } from '../rule-tester';

const defaults = {
  filename: 'src/virtual.js',
  languageOptions: { sourceType: 'module' },
} satisfies Pick<ValidTestCase, 'filename' | 'languageOptions'>;

const messages = {
  error: {
    messageId: 'no-negation-in-equality-check/error',
    message: 'Negated expression is not allowed in equality check.',
  },
};

new RuleTester().run('no-negation-in-equality-check', {} as never, {
  valid: [
    { ...defaults, code: '!foo instanceof bar' },
    { ...defaults, code: '+foo === bar' },
    { ...defaults, code: '!(foo === bar)' },
    { ...defaults, code: '!!foo === bar' },
    { ...defaults, code: '!!!foo === bar' },
    { ...defaults, code: 'foo === !bar' },
    { ...defaults, code: '// ✅\nif (foo !== bar) {}' },
    { ...defaults, code: '// ✅\nif (!(foo === bar)) {}' },
  ],
  invalid: [
    { ...defaults, code: '!foo === bar', errors: [messages.error] },
    { ...defaults, code: '!foo !== bar', errors: [messages.error] },
    { ...defaults, code: '!foo == bar', errors: [messages.error] },
    { ...defaults, code: '!foo != bar', errors: [messages.error] },
    {
      ...defaults,
      code: 'function x() {\n\treturn!foo === bar;\n}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: 'function x() {\n\treturn!\n\t\tfoo === bar;\n\tthrow!\n\t\tfoo === bar;\n}',
      errors: [messages.error, messages.error],
    },
    { ...defaults, code: 'foo\n!(a) === b', errors: [messages.error] },
    {
      ...defaults,
      code: "foo\n![a, b].join('') === c",
      errors: [messages.error],
    },
    {
      ...defaults,
      code: "foo\n! [a, b].join('') === c",
      errors: [messages.error],
    },
    {
      ...defaults,
      code: "foo\n!/* comment */[a, b].join('') === c",
      errors: [messages.error],
    },
    {
      ...defaults,
      code: '// ❌\nif (!foo === bar) {}',
      errors: [messages.error],
    },
    {
      ...defaults,
      code: '// ❌\nif (!foo !== bar) {}',
      errors: [messages.error],
    },
  ],
});
