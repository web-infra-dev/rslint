// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const filename = 'src/virtual.js';
const valid = [
  "const foo = `\\${a}`",
  'const foo = `hello`',
  'const foo = `$`',
  'const foo = `{`',
  'const foo = ``',
  'const foo = `${a}`',
  'const foo = `${a}${b}`',
  "const foo = String.raw`$\\{a}`",
  "const foo = html`$\\{a}`",
  "const foo = `\\\\\\${a}`",
  "const foo = '$\\{a}'",
].map((code) => ({ code, filename }));

const invalid = [
  "const foo = `$\\{a}`",
  "const foo = `\\$\\{a}`",
  "const foo = `$\\{a} and $\\{b}`",
  "const foo = `\\\\$\\{a}`",
  "const foo = `\\\\\\$\\{a}`",
  "const foo = `$\\{a}${expr}`",
  "const foo = `${expr}$\\{a}`",
  "const foo = `$\\{a}${expr}$\\{b}`",
].map((code) => ({
  code,
  filename,
  errors: [
    {
      messageId: 'consistent-template-literal-escape',
      message:
        'Use `\\${` instead of `$\\{` to escape in template literals.',
    },
  ],
}));

invalid[7]!.errors.push({
  messageId: 'consistent-template-literal-escape',
  message: 'Use `\\${` instead of `$\\{` to escape in template literals.',
});

ruleTester.run('consistent-template-literal-escape', null as never, {
  valid,
  invalid,
});
