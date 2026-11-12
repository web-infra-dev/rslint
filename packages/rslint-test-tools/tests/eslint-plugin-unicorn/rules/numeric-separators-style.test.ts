// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/numeric-separators-style.js
import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { describe } from 'rstack/test';

function invalid(
  code: string,
  output: string,
  ranges: [line: number, column: number, endColumn: number][],
  options?: unknown[],
): TestCase {
  return {
    code,
    output,
    options,
    errors: ranges.map(([line, column, endColumn]) => ({
      messageId: 'numeric-separators-style',
      message: 'Invalid group length in numeric value.',
      line,
      column,
      endLine: line,
      endColumn,
    })),
  };
}

const tester = new RuleTester({ plugin: 'unicorn' });
const ruleName = 'numeric-separators-style';

tester.run(ruleName, {
  valid: [
    // Hexadecimal.
    'const foo = 0xAB_CD',
    'const foo = 0xAB',
    'const foo = 0xA',
    'const foo = 0xA_BC_DE_F0',
    'const foo = 0xab_e8_12',
    'const foo = 0xe',
    'const foo = 0Xab_e3_cd',
    // Octal.
    'const foo = 0o1234_5670',
    'const foo = 0o7777',
    'const foo = 0o01',
    'const foo = 0o12_7000_0000',
    'const foo = 0O1111_1111',
    // Binary.
    'const foo = 0b1010_0001_1000_0101',
    'const foo = 0b0000',
    'const foo = 0b10',
    'const foo = 0b1_0111_0101_0101',
    'const foo = 0B1010',
    // BigInt.
    'const foo = 0b1010n',
    'const foo = 0b1010_1010n',
    'const foo = 9_223_372_036_854_775_807n',
    'const foo = 807n',
    'const foo = 1n',
    'const foo = 9_372_854_807n',
    'const foo = 9807n',
    'const foo = 0n',
    // Numbers.
    'const foo = 12_345_678',
    'const foo = 123',
    'const foo = 1',
    'const foo = 1234',
    {
      code: 'const foo = 1_234',
      options: [{ number: { minimumDigits: 0, groupLength: 3 } }],
    },
    // Fractional parts.
    'const foo = 9807.123',
    'const foo = 3819.123432',
    'const foo = 138_789.12343242',
    'const foo = .0000001',
    'const foo = .00000',
    'const foo = 0.00000',
    'const foo = 0.55228474983',
    {
      code: 'const foo = 1.234_567',
      options: [{ number: { fractionGroupLength: 3 } }],
    },
    {
      code: 'const foo = 149_597_870_700.31415_92653_58979',
      options: [
        {
          number: { minimumDigits: 0, groupLength: 3, fractionGroupLength: 5 },
        },
      ],
    },
    {
      code: 'const foo = 1.2345',
      options: [{ number: { fractionGroupLength: 2 } }],
    },
    // Negative numbers.
    'const foo = -3000',
    'const foo = -10_000_000',
    // Exponential notation.
    'const foo = 1e10_000',
    'const foo = 39_804e1000',
    'const foo = -123_456e-100',
    'const foo = -100_000e-100_000',
    'const foo = -100_000e+100_000',
    'const foo = 3.6e12_000',
    'const foo = 3.6E12_000',
    'const foo = -1_200_000e5',
    // Other values.
    'const foo = -282_932 - (1938 / 10_000) * .1 + 18.1000002',
    'const foo = NaN',
    'const foo = Infinity',
    'const foo = "1234567n"',
    // Varying options.
    { code: 'const foo = 10000', options: [{ number: { minimumDigits: 6 } }] },
    {
      code: 'const foo = 100_0000_0000',
      options: [{ number: { groupLength: 4 } }],
    },
    {
      code: 'const foo = 0xA_B_C_D_E_1_2_3_4',
      options: [{ hexadecimal: { groupLength: 1 } }],
    },
    {
      code: 'const foo = 0b111',
      options: [{ number: { minimumDigits: 3, groupLength: 1 } }],
    },
    {
      code: `const binary = 0b10101010;
const octal = 0o76543210;
const hexadecimal = 0xfedcba97;
const number = 12345678.12345678e12345678;`,
      options: [{ onlyIfContainsSeparator: true }],
    },
    {
      code: `const binary = 0b1010_1010;
const octal = 0o76543210;
const hexadecimal = 0xfedcba97;
const number = 12345678.12345678e12345678;`,
      options: [
        {
          onlyIfContainsSeparator: true,
          binary: { onlyIfContainsSeparator: false },
        },
      ],
    },
    {
      code: `const binary = 0b10_10_10_10;
const octal = 0o76543210;
const hexadecimal = 0xfedcba97;
const number = 12345678.12345678e12345678;`,
      options: [
        {
          onlyIfContainsSeparator: true,
          binary: { onlyIfContainsSeparator: false, groupLength: 2 },
        },
      ],
    },
    {
      code: `const binary = 0b10101010;
const octal = 0o7654_3210;
const hexadecimal = 0xfe_dc_ba_97;
const number = 12_345_678.12345678e12_345_678;`,
      options: [{ binary: { onlyIfContainsSeparator: true } }],
    },
    {
      code: 'const foo = 12345',
      options: [{ number: { onlyIfContainsSeparator: true } }],
    },
    {
      code: 'const foo = 12345678',
      options: [{ number: { onlyIfContainsSeparator: true } }],
    },
    {
      code: 'const foo = 12_345',
      options: [{ number: { onlyIfContainsSeparator: true } }],
    },
    {
      code: 'const foo = -100_000e+100_000',
      options: [{ number: { onlyIfContainsSeparator: true } }],
    },
    {
      code: 'const foo = -100000e+100000',
      options: [{ number: { onlyIfContainsSeparator: true } }],
    },
    {
      code: 'const foo = 0xA_B_C_D_E',
      options: [
        { hexadecimal: { onlyIfContainsSeparator: true, groupLength: 1 } },
      ],
    },
    {
      code: 'const foo = 0o7777',
      options: [{ octal: { onlyIfContainsSeparator: true, minimumDigits: 4 } }],
    },
    {
      code: 'const foo = 0xABCDEF012',
      options: [{ hexadecimal: { onlyIfContainsSeparator: true } }],
    },
    {
      code: 'const foo = 0o777777',
      options: [{ octal: { onlyIfContainsSeparator: true, minimumDigits: 3 } }],
    },
    {
      code: 'const foo = 0o777777',
      options: [
        {
          octal: {
            onlyIfContainsSeparator: true,
            minimumDigits: 3,
            groupLength: 2,
          },
        },
      ],
    },
    {
      code: 'const foo = 0o777_777',
      options: [
        {
          octal: {
            onlyIfContainsSeparator: true,
            minimumDigits: 2,
            groupLength: 3,
          },
        },
      ],
    },
    {
      code: 'const foo = 0b01010101',
      options: [
        {
          onlyIfContainsSeparator: true,
          binary: { onlyIfContainsSeparator: true },
        },
      ],
    },
    {
      code: 'const foo = 0b0101_0101',
      options: [
        {
          onlyIfContainsSeparator: false,
          binary: { onlyIfContainsSeparator: true },
        },
      ],
    },
    {
      code: 'const foo = 0b0101_0101',
      options: [
        {
          onlyIfContainsSeparator: false,
          binary: { onlyIfContainsSeparator: false },
        },
      ],
    },
    // Documentation examples.
    'const foo = 1_234_444;',
    'const foo = 1234.56789;',
    'const foo = 0xAB_CD_EF;',
    'const foo = 0b1000_1111;',
    'const foo = 0o10_4421;',
    'const foo = 1_294_287_712n;',
    {
      code: 'const number = 100000; // Pass, this number does not contain separators',
      options: [
        {
          onlyIfContainsSeparator: true,
          binary: { onlyIfContainsSeparator: false },
        },
      ],
    },
    {
      code: 'const groupedInteger = 12_345;',
      options: [{ number: { minimumDigits: 0, groupLength: 3 } }],
    },
    {
      code: 'const fractional = 0.0000001;',
      options: [{ number: { minimumDigits: 0, groupLength: 3 } }],
    },
    {
      code: 'const groupedFractional = 0.000_000_1;',
      options: [
        {
          number: { minimumDigits: 0, groupLength: 3, fractionGroupLength: 3 },
        },
      ],
    },
    {
      code: 'const groupedLongFractional = 123.100_000_1;',
      options: [
        {
          number: { minimumDigits: 0, groupLength: 3, fractionGroupLength: 3 },
        },
      ],
    },
    {
      code: 'const groupedBinary = 0b10_1010;',
      options: [{ binary: { minimumDigits: 0, groupLength: 4 } }],
    },
    {
      code: 'const groupedLongBinary = 0b1_0101_0001;',
      options: [{ binary: { minimumDigits: 0, groupLength: 4 } }],
    },
    {
      code: 'const groupedHexadecimal = 0xAB_CD_EF;',
      options: [{ hexadecimal: { minimumDigits: 0, groupLength: 2 } }],
    },
    {
      code: 'const smallNumber = 100;',
      options: [{ number: { minimumDigits: 0, groupLength: 3 } }],
    },
    {
      code: 'const groupedNumber = 1_000;',
      options: [{ number: { minimumDigits: 0, groupLength: 3 } }],
    },
    {
      code: 'const largeGroupedNumber = 1_000_000;',
      options: [{ number: { minimumDigits: 0, groupLength: 3 } }],
    },
    {
      code: 'const foo = 1000;',
      options: [{ number: { minimumDigits: 5, groupLength: 3 } }],
    },
    {
      code: 'const octal = 0o7777;',
      options: [{ octal: { minimumDigits: 0, groupLength: 4 } }],
    },
    {
      code: 'const anotherOctal = 0o7777;',
      options: [{ octal: { minimumDigits: 0, groupLength: 4 } }],
    },
    {
      code: 'const groupedOctal = 0o12_7777;',
      options: [{ octal: { minimumDigits: 0, groupLength: 4 } }],
    },
    {
      code: 'const value = 1024;',
      options: [{ number: { minimumDigits: 5 } }],
    },
    {
      code: 'const pi = 3.14159_26535_89793;',
      options: [{ number: { fractionGroupLength: 5 } }],
    },
    {
      code: 'const values = [12_345, 0xAB_CD, 0b1010_1010, 0o1234_5670, 0.5522847498];',
      options: [
        {
          onlyIfContainsSeparator: false,
          hexadecimal: { minimumDigits: 0, groupLength: 2 },
          binary: { minimumDigits: 0, groupLength: 4 },
          octal: { minimumDigits: 0, groupLength: 4 },
          number: { minimumDigits: 5, groupLength: 3 },
        },
      ],
    },
  ],
  invalid: [
    // Hexadecimal.
    invalid('const foo = 0xA_B_CDE_F0', 'const foo = 0xA_BC_DE_F0', [
      [1, 13, 25],
    ]),
    invalid('const foo = 0xABCDEF', 'const foo = 0xAB_CD_EF', [[1, 13, 21]]),
    invalid('const foo = 0xA_B', 'const foo = 0xAB', [[1, 13, 18]]),
    invalid('const foo = 0XAB_C_D', 'const foo = 0XAB_CD', [[1, 13, 21]]),
    // Octal.
    invalid('const foo = 0o12_34_5670', 'const foo = 0o1234_5670', [
      [1, 13, 25],
    ]),
    invalid('const foo = 0o7_7_77', 'const foo = 0o7777', [[1, 13, 21]]),
    invalid('const foo = 0o010101010101', 'const foo = 0o0101_0101_0101', [
      [1, 13, 27],
    ]),
    invalid('const foo = 0O010101010101', 'const foo = 0O0101_0101_0101', [
      [1, 13, 27],
    ]),
    // Binary.
    invalid('const foo = 0b10_10_0001', 'const foo = 0b1010_0001', [
      [1, 13, 25],
    ]),
    invalid('const foo = 0b0_00_0', 'const foo = 0b0000', [[1, 13, 21]]),
    invalid('const foo = 0b10101010101010', 'const foo = 0b10_1010_1010_1010', [
      [1, 13, 29],
    ]),
    invalid('const foo = 0B10101010101010', 'const foo = 0B10_1010_1010_1010', [
      [1, 13, 29],
    ]),
    // BigInt.
    invalid('const foo = 0b10101010n', 'const foo = 0b1010_1010n', [
      [1, 13, 24],
    ]),
    invalid('const foo = 1_9_223n', 'const foo = 19_223n', [[1, 13, 21]]),
    invalid('const foo = 80_7n', 'const foo = 807n', [[1, 13, 18]]),
    invalid('const foo = 123456789_100n', 'const foo = 123_456_789_100n', [
      [1, 13, 27],
    ]),
    // Numbers.
    invalid('const foo = 1_2_345_678', 'const foo = 12_345_678', [[1, 13, 24]]),
    invalid('const foo = 12_3', 'const foo = 123', [[1, 13, 17]]),
    invalid('const foo = 1234567890', 'const foo = 1_234_567_890', [
      [1, 13, 23],
    ]),
    // Fractional parts and numeric member access.
    invalid('const foo = 0.552_284_749_83', 'const foo = 0.55228474983', [
      [1, 13, 29],
    ]),
    invalid('const foo = 3819.123_4325', 'const foo = 3819.1234325', [
      [1, 13, 26],
    ]),
    invalid('const foo = 138789.12343_2_42', 'const foo = 138_789.12343242', [
      [1, 13, 30],
    ]),
    invalid('const foo = .000000_1', 'const foo = .0000001', [[1, 13, 22]]),
    invalid(
      'const foo = 12345678..toString()',
      'const foo = 12_345_678..toString()',
      [[1, 13, 22]],
    ),
    invalid(
      'const foo = 12345678 .toString()',
      'const foo = 12_345_678 .toString()',
      [[1, 13, 21]],
    ),
    invalid(
      'const foo = 1.234567',
      'const foo = 1.234_567',
      [[1, 13, 21]],
      [{ number: { fractionGroupLength: 3 } }],
    ),
    invalid(
      'const foo = 1789.123_432_42',
      'const foo = 1789.12343242',
      [[1, 13, 28]],
      [{ number: { onlyIfContainsSeparator: true } }],
    ),
    invalid(
      'const foo = -282_932 - (1938 / 10_000) * .1 + 18.100_000_2',
      'const foo = -282_932 - (1938 / 10_000) * .1 + 18.1000002',
      [[1, 47, 59]],
      [{ number: { onlyIfContainsSeparator: true } }],
    ),
    // Negative numbers.
    invalid('const foo = -100000_1', 'const foo = -1_000_001', [[1, 14, 22]]),
    // Exponential notation.
    invalid('const foo = 1e10000', 'const foo = 1e10_000', [[1, 13, 20]]),
    invalid('const foo = 39804e10000', 'const foo = 39_804e10_000', [
      [1, 13, 24],
    ]),
    invalid('const foo = -123456e100', 'const foo = -123_456e100', [
      [1, 14, 24],
    ]),
    invalid('const foo = -100000e-10000', 'const foo = -100_000e-10_000', [
      [1, 14, 27],
    ]),
    invalid('const foo = -1000e+10000', 'const foo = -1000e+10_000', [
      [1, 14, 25],
    ]),
    invalid('const foo = -1000e+00010000', 'const foo = -1000e+00_010_000', [
      [1, 14, 28],
    ]),
    invalid('const foo = 3.6e12000', 'const foo = 3.6e12_000', [[1, 13, 22]]),
    invalid('const foo = -1200000e5', 'const foo = -1_200_000e5', [
      [1, 14, 23],
    ]),
    invalid('const foo = 3.65432E12000', 'const foo = 3.65432E12_000', [
      [1, 13, 26],
    ]),
    // Varying options.
    invalid(
      'const foo = 1000000',
      'const foo = 1_000_000',
      [[1, 13, 20]],
      [{ number: { minimumDigits: 6 } }],
    ),
    invalid(
      'const foo = 10_000_000_000',
      'const foo = 100_0000_0000',
      [[1, 13, 27]],
      [{ number: { groupLength: 4 } }],
    ),
    invalid(
      'const foo = 0xA_B_CD',
      'const foo = 0xA_B_C_D',
      [[1, 13, 21]],
      [{ hexadecimal: { groupLength: 1 } }],
    ),
    invalid(
      'const foo = 0b1_11',
      'const foo = 0b111',
      [[1, 13, 19]],
      [{ number: { minimumDigits: 3, groupLength: 2 } }],
    ),
    invalid(
      'const foo = -100000e+100000',
      'const foo = -100_000e+100_000',
      [[1, 14, 28]],
      [{ number: { onlyIfContainsSeparator: false } }],
    ),
    invalid(
      `const binary = 0b10_101010;
const octal = 0o76_543210;
const hexadecimal = 0xfe_dcba97;
const number = 12_345678.12345678e12345678;`,
      `const binary = 0b1010_1010;
const octal = 0o7654_3210;
const hexadecimal = 0xfe_dc_ba_97;
const number = 12_345_678.12345678e12_345_678;`,
      [
        [1, 16, 27],
        [2, 15, 26],
        [3, 21, 32],
        [4, 16, 43],
      ],
      [{ onlyIfContainsSeparator: true }],
    ),
    invalid(
      `const binary = 0b10101010;
const octal = 0o76_543210;
const hexadecimal = 0xfe_dcba97;
const number = 12_345678.12345678e12345678;`,
      `const binary = 0b1010_1010;
const octal = 0o7654_3210;
const hexadecimal = 0xfe_dc_ba_97;
const number = 12_345_678.12345678e12_345_678;`,
      [
        [1, 16, 26],
        [2, 15, 26],
        [3, 21, 32],
        [4, 16, 43],
      ],
      [
        {
          onlyIfContainsSeparator: true,
          binary: { onlyIfContainsSeparator: false },
        },
      ],
    ),
    invalid(
      `const binary = 0b10101010;
const octal = 0o76_543210;
const hexadecimal = 0xfe_dcba97;
const number = 12_345678.12345678e12345678;`,
      `const binary = 0b10_10_10_10;
const octal = 0o7654_3210;
const hexadecimal = 0xfe_dc_ba_97;
const number = 12_345_678.12345678e12_345_678;`,
      [
        [1, 16, 26],
        [2, 15, 26],
        [3, 21, 32],
        [4, 16, 43],
      ],
      [
        {
          onlyIfContainsSeparator: true,
          binary: { onlyIfContainsSeparator: false, groupLength: 2 },
        },
      ],
    ),
    invalid(
      `const binary = 0b10_101010;
const octal = 0o76543210;
const hexadecimal = 0xfedcba97;
const number = 12345678.12345678e12345678;`,
      `const binary = 0b1010_1010;
const octal = 0o7654_3210;
const hexadecimal = 0xfe_dc_ba_97;
const number = 12_345_678.12345678e12_345_678;`,
      [
        [1, 16, 27],
        [2, 15, 25],
        [3, 21, 31],
        [4, 16, 42],
      ],
      [{ binary: { onlyIfContainsSeparator: true } }],
    ),
    // Snapshot cases.
    invalid('console.log(0XdeEdBeeFn)', 'console.log(0Xde_Ed_Be_eFn)', [
      [1, 13, 24],
    ]),
    invalid(
      'const foo = 12345678..toString()',
      'const foo = 12_345_678..toString()',
      [[1, 13, 22]],
    ),
    // Documentation examples.
    invalid('const foo = 1_23_4444;', 'const foo = 1_234_444;', [[1, 13, 22]]),
    invalid('const foo = 1_234.567_89;', 'const foo = 1234.56789;', [
      [1, 13, 25],
    ]),
    invalid('const foo = 0xAB_C_D_EF;', 'const foo = 0xAB_CD_EF;', [
      [1, 13, 24],
    ]),
    invalid('const foo = 0b10_00_1111;', 'const foo = 0b1000_1111;', [
      [1, 13, 25],
    ]),
    invalid('const foo = 0o1_0_44_21;', 'const foo = 0o10_4421;', [
      [1, 13, 24],
    ]),
    invalid('const foo = 1_294_28771_2n;', 'const foo = 1_294_287_712n;', [
      [1, 13, 27],
    ]),
    invalid(
      "const binary = 0b101010001; // Fail, `binary` type don't require separators",
      "const binary = 0b1_0101_0001; // Fail, `binary` type don't require separators",
      [[1, 16, 27]],
      [
        {
          onlyIfContainsSeparator: true,
          binary: { onlyIfContainsSeparator: false },
        },
      ],
    ),
    invalid(
      "const hexadecimal = 0xD_EED_BEE_F; // Fail, it contains separators and it's incorrectly grouped",
      "const hexadecimal = 0xDE_ED_BE_EF; // Fail, it contains separators and it's incorrectly grouped",
      [[1, 21, 34]],
      [
        {
          onlyIfContainsSeparator: true,
          binary: { onlyIfContainsSeparator: false },
        },
      ],
    ),
    invalid(
      'const integer = 12345;',
      'const integer = 12_345;',
      [[1, 17, 22]],
      [{ number: { minimumDigits: 0, groupLength: 3 } }],
    ),
    invalid(
      'const fractional = 0.000_0001;',
      'const fractional = 0.000_000_1;',
      [[1, 20, 30]],
      [
        {
          number: { minimumDigits: 0, groupLength: 3, fractionGroupLength: 3 },
        },
      ],
    ),
    invalid(
      'const longFractional = 123.1_000_001;',
      'const longFractional = 123.100_000_1;',
      [[1, 24, 37]],
      [
        {
          number: { minimumDigits: 0, groupLength: 3, fractionGroupLength: 3 },
        },
      ],
    ),
    invalid(
      'const binary = 0b101010;',
      'const binary = 0b10_1010;',
      [[1, 16, 24]],
      [{ binary: { minimumDigits: 0, groupLength: 4 } }],
    ),
    invalid(
      'const longBinary = 0b1010_10001;',
      'const longBinary = 0b1_0101_0001;',
      [[1, 20, 32]],
      [{ binary: { minimumDigits: 0, groupLength: 4 } }],
    ),
    invalid(
      'const hexadecimal = 0xA_B_CD_EF;',
      'const hexadecimal = 0xAB_CD_EF;',
      [[1, 21, 32]],
      [{ hexadecimal: { minimumDigits: 0, groupLength: 2 } }],
    ),
    invalid(
      'const value = 1_024;',
      'const value = 1024;',
      [[1, 15, 20]],
      [{ number: { minimumDigits: 5 } }],
    ),
    invalid(
      'const value = 0.5522847498;',
      'const value = 0.55228_47498;',
      [[1, 15, 27]],
      [{ number: { fractionGroupLength: 5 } }],
    ),
  ],
});

// The Go upstream suite checks the rule directly on these recovered literals.
describe.skip('Legacy leading-zero literals are rejected by tsgo', () => {
  tester.run(ruleName, {
    valid: [
      {
        code: 'const foo = 0777777',
        languageOptions: { sourceType: 'script' },
      },
      {
        code: 'var foo = 0999999',
        languageOptions: { sourceType: 'script' },
      },
      {
        code: 'let foo = 0111222',
        languageOptions: { sourceType: 'script' },
      },
    ],
    invalid: [],
  });
});

describe.skip('TOML and SOML require ESLint language plugins', () => {
  tester.run(ruleName, {
    valid: [
      // TOML cases.
      {
        code: 'value = 1_000_000',
        filename: 'example.toml',
      },
      {
        code: 'value = [-1234, +1234, -12_345, +12_345]',
        filename: 'example.toml',
      },
      {
        code: 'value = [inf, +inf, -inf, nan, +nan, -nan]',
        filename: 'example.toml',
      },
      {
        code: 'value = ["1234567", 1979-05-27T07:32:00Z, 1979-05-27, 07:32:00]',
        filename: 'example.toml',
      },
      {
        code: 'value = [0xAB_CD, 0o1234_5670, 0b1010_1010]',
        filename: 'example.toml',
      },
      {
        code: 'value = [1234567, 0xABCDEF]',
        options: [{ onlyIfContainsSeparator: true }],
        filename: 'example.toml',
      },
      {
        code: 'value = [-12345, +12345]',
        options: [{ number: { minimumDigits: 6 } }],
        filename: 'example.toml',
      },
      // SOML cases.
      {
        code: 'value: [1_000_000, 0xAB_CD, 0b1010_1010, 0o1234_5670]',
        filename: 'input.soml',
      },
      {
        code: 'value: [infinity, -infinity]',
        filename: 'input.soml',
      },
      {
        code: "value: ['1234567', 2026-09-19T14:00:00Z, 12345s]",
        filename: 'input.soml',
      },
      {
        code: 'value: 1234567',
        options: [{ onlyIfContainsSeparator: true }],
        filename: 'input.soml',
      },
    ],
    invalid: [
      // TOML cases.
      {
        ...invalid(
          'value = 1000000 # Preserve this comment.',
          'value = 1_000_000 # Preserve this comment.',
          [[1, 9, 16]],
        ),
        filename: 'example.toml',
      },
      {
        ...invalid(
          'value = [-123456, +123456]',
          'value = [-123_456, +123_456]',
          [
            [1, 10, 17],
            [1, 19, 26],
          ],
        ),
        filename: 'example.toml',
      },
      {
        ...invalid(
          'value = 9223372036854775807',
          'value = 9_223_372_036_854_775_807',
          [[1, 9, 28]],
        ),
        filename: 'example.toml',
      },
      {
        ...invalid(
          'value = [0xABCDEF, 0o12345670, 0b10101010]',
          'value = [0xAB_CD_EF, 0o1234_5670, 0b1010_1010]',
          [
            [1, 10, 18],
            [1, 20, 30],
            [1, 32, 42],
          ],
        ),
        filename: 'example.toml',
      },
      {
        ...invalid(
          'value = {nested = [12345, 1_2_345, 0.123_456]}',
          'value = {nested = [12_345, 12_345, 0.123456]}',
          [
            [1, 20, 25],
            [1, 27, 34],
            [1, 36, 45],
          ],
        ),
        filename: 'example.toml',
      },
      {
        ...invalid(
          'value = [-12345.123456E+10000, +12345.123456e-10000]',
          'value = [-12_345.123456E+10_000, +12_345.123456e-10_000]',
          [
            [1, 10, 30],
            [1, 32, 52],
          ],
        ),
        filename: 'example.toml',
      },
      {
        ...invalid(
          'value = -1234567.12345678',
          'value = -123_4567.123_456_78',
          [[1, 9, 26]],
          [{ number: { groupLength: 4, fractionGroupLength: 3 } }],
        ),
        filename: 'example.toml',
      },
      {
        ...invalid(
          'value = [+123, -123]',
          'value = [+1_2_3, -1_2_3]',
          [
            [1, 10, 14],
            [1, 16, 20],
          ],
          [{ number: { minimumDigits: 0, groupLength: 1 } }],
        ),
        filename: 'example.toml',
      },
      {
        ...invalid(
          'value = [1234567, 12_34567, 0xABCDEF]',
          'value = [1234567, 1_234_567, 0xABC_DEF]',
          [
            [1, 19, 27],
            [1, 29, 37],
          ],
          [
            {
              onlyIfContainsSeparator: true,
              hexadecimal: { onlyIfContainsSeparator: false, groupLength: 3 },
            },
          ],
        ),
        filename: 'example.toml',
      },
      // SOML cases.
      {
        ...invalid('value: 1000000 # keep', 'value: 1_000_000 # keep', [
          [1, 8, 15],
        ]),
        filename: 'input.soml',
      },
      {
        ...invalid(
          'value: [-123456, 9223372036854775807]',
          'value: [-123_456, 9_223_372_036_854_775_807]',
          [
            [1, 9, 16],
            [1, 18, 37],
          ],
        ),
        filename: 'input.soml',
      },
      {
        ...invalid(
          'value: [0xABCDEF, 0b10101010, 0o12345670]',
          'value: [0xAB_CD_EF, 0b1010_1010, 0o1234_5670]',
          [
            [1, 9, 17],
            [1, 19, 29],
            [1, 31, 41],
          ],
        ),
        filename: 'input.soml',
      },
      {
        ...invalid('value: 1_2_345.123_456e-12', 'value: 12_345.123456e-12', [
          [1, 8, 27],
        ]),
        filename: 'input.soml',
      },
      {
        ...invalid(
          'value: -1234567.12345678',
          'value: -123_4567.123_456_78',
          [[1, 8, 25]],
          [{ number: { groupLength: 4, fractionGroupLength: 3 } }],
        ),
        filename: 'input.soml',
      },
    ],
  });
});
