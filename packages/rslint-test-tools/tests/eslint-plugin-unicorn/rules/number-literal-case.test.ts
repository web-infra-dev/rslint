// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/test/number-literal-case.js
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const valid = (code: string) => ({ code, filename: 'file.js' });
const invalid = (code: string, output: string) => ({
  code,
  output,
  filename: 'file.js',
  errors: [
    {
      messageId: 'number-literal-case',
      message: 'Invalid number literal casing.',
    },
  ],
});

ruleTester.run('number-literal-case', null as never, {
  valid: [
    ...[
      // Number.
      '1234',
      '0b10',
      '0o1234567',
      '0xABCDEF',
      // BigInt.
      '1234n',
      '0b10n',
      '0o1234567n',
      '0xABCDEFn',
      // Symbolic values.
      'NaN',
      '+Infinity',
      '-Infinity',
      // Exponential notation.
      '1.2e3',
      '1.2e-3',
      '1.2e+3',
      // Not numbers.
      "'0Xff'",
      "'0Xffn'",
      // Numeric separators.
      '123_456',
      '0b10_10',
      '0o1_234_567',
      '0xDEED_BEEF',
      '123_456n',
      '0b10_10n',
      '0o1_234_567n',
      '0xDEED_BEEFn',
      // Negative numbers.
      '-1234',
      '-0b10',
      '-0o1234567',
      '-0xABCDEF',
    ].map((value) => valid(`const foo = ${value}`)),
  ],
  invalid: [
    ...[
      // Number.
      ['0B10', '0b10'],
      ['0O1234567', '0o1234567'],
      ['0XaBcDeF', '0xABCDEF'],
      // BigInt.
      ['0B10n', '0b10n'],
      ['0O1234567n', '0o1234567n'],
      ['0XaBcDeFn', '0xABCDEFn'],
      // BigInt zero.
      ['0B0n', '0b0n'],
      ['0O0n', '0o0n'],
      ['0X0n', '0x0n'],
      // Exponential notation, including integer mantissas.
      ['1.2E3', '1.2e3'],
      ['5E3', '5e3'],
      ['5E+3', '5e+3'],
      ['1.2E-3', '1.2e-3'],
      ['1.2E+3', '1.2e+3'],
      // Numeric separators.
      ['0XdeEd_Beefn', '0xDEED_BEEFn'],
      // Negative numbers.
      ['-0B10', '-0b10'],
      ['-0O1234567', '-0o1234567'],
      ['-0XaBcDeF', '-0xABCDEF'],
      ['-0XaBcn', '-0xABCn'],
    ].map(([value, fixed]) =>
      invalid(`const foo = ${value}`, `const foo = ${fixed}`),
    ),
    invalid(
      "const foo = 255;\n\nif (foo === 0xff) {\n\tconsole.log('invalid');\n}",
      "const foo = 255;\n\nif (foo === 0xFF) {\n\tconsole.log('invalid');\n}",
    ),
    // Lowercase hexadecimal values.
    ...[
      ['0XaBcDeF', '0xabcdef'],
      ['0xaBcDeF', '0xabcdef'],
      ['0XaBcDeFn', '0xabcdefn'],
      ['0XdeEd_Beefn', '0xdeed_beefn'],
    ].map(([value, fixed]) => ({
      ...invalid(`const foo = ${value}`, `const foo = ${fixed}`),
      options: [{ hexadecimalValue: 'lowercase' }],
    })),
  ],
});

// tsgo rejects these sloppy-mode literals before running rules (TS1121/TS1489).
// The Go upstream suite checks their casing directly on the parsed literals.
test.skip.each(['var foo = 0777', 'var foo = 0888'])(
  'Legacy literal parser limitation: %s',
  () => {},
);

// rslint does not provide the Vue parser service used by these upstream cases.
test.skip.each([
  '<template><input value="0XdeEd_Beef"></div></template>',
  '<template><div v-if="0xDEED_BEEF > 0"></div></template>',
  '<template><div v-if="0XdeEd_Beef > 0"></div></template>',
  '<template><div v-if="0XdeEd_Beefn > 0n"></div></template>',
  '<template><div>{{1.2E3}}</div></template>',
  '<template><div>{{0B1n}}</div></template>',
  '<script>export default {data() {return {n: 0XdeEd_Beefn}}}</script>',
])('Vue parser service: %s', () => {});
