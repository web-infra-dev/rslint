// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-bigint-literals.js
import { RuleTester, type ValidTestCase } from '../rule-tester';

const ruleName = 'prefer-bigint-literals';
const error = {
  messageId: 'prefer-bigint-literals/error',
  message: 'Prefer using bigint literal over `BigInt(…)`.',
};

function js(code: string): ValidTestCase {
  return { code, filename: 'src/virtual.js' };
}

function invalid(code: string, count = 1) {
  return {
    ...js(code),
    errors: Array.from({ length: count }, () => error),
  };
}

// The Go upstream suite asserts the complete ranges, autofixes and suggestions.
new RuleTester().run(ruleName, {} as never, {
  valid: [
    // Numbers and strings.
    js('1n'),
    js('BigInt()'),
    js('BigInt(1, 1)'),
    js('BigInt(...[1])'),
    js('BigInt(true)'),
    js('BigInt(null)'),
    js('new BigInt(1)'),
    js('Not_BigInt(1)'),
    js('BigInt("1.0")'),
    js('BigInt("1.1")'),
    js('BigInt("1e3")'),
    js('BigInt("1" + "2")'),
    js('BigInt?.(1)'),
    js('BigInt(1.1)'),
    js('typeof BigInt'),
    js('BigInt(1n)'),
    js('BigInt("not-number")'),
    js('BigInt("1_2")'),
    js('BigInt(!1)'),
    js('BigInt(~1)'),
    js('BigInt("++1")'),
    js('BigInt("+ 1")'),
    js('BigInt(void 0)'),
    js('BigInt(NaN)'),
    js('BigInt(Infinity)'),
    js('BigInt(-Infinity)'),
    js('BigInt(BigInt)'),
    js('BigInt(globalThis.x)'),
    js('BigInt(x)'),
    js('BigInt?.("1")'),
    // Templates and escapes.
    js('BigInt(`1.1`)'),
    js('BigInt(`not-number`)'),
    js('BigInt(tag`1`)'),
    js('BigInt(`1${value}`)'),
    // Documentation examples.
    js('const bigint = 1n;'),
    js('const large = 9007199254740991n;'),
    js('const hex = 0xFFn;'),
    js('const computed = BigInt(getSomeNumber());'),
    js('const astronomicalNumber = 999999999999999999999999n;'),
  ],
  invalid: [
    // Numbers and strings.
    invalid('BigInt("0")'),
    invalid('BigInt("  0  ")'),
    invalid('BigInt("9007199254740993")'),
    invalid('BigInt("0B11")'),
    invalid('BigInt("0O777")'),
    invalid('BigInt("0XFe")'),
    invalid(`BigInt("${'9'.repeat(100)}")`),
    invalid('BigInt(0)'),
    invalid('BigInt(0B11_11)'),
    invalid('BigInt(0O777_777)'),
    invalid('BigInt(0XFe_fE)'),
    invalid('BigInt(9007199254740993)'),
    invalid('BigInt(0x20000000000001)'),
    invalid('BigInt(9_007_199_254_740_993)'),
    invalid('BigInt(0x20_00_00_00_00_00_01)'),
    invalid('BigInt(1.0)'),
    invalid('BigInt(1e2)'),
    invalid('BigInt(/* comment */1)'),
    invalid(`BigInt(${'9'.repeat(100)})`),
    invalid('BigInt("-1")'),
    invalid('BigInt("+1")'),
    invalid('BigInt(-1)'),
    invalid('BigInt(+1)'),
    invalid('-BigInt(-1)'),
    invalid('-BigInt("-1")'),
    invalid('-BigInt(1)'),
    invalid('-BigInt("1")'),
    invalid('BigInt(" +1 ")'),
    invalid('BigInt(" -1 ")'),
    invalid('\n\t\tfoo\n\t\tBigInt("-1")\n\t\t'),
    invalid('-BigInt("+1")'),
    invalid('-BigInt(/* comment */-1)'),
    invalid('-(BigInt(-1))'),
    invalid('2n - BigInt("-1")'),
    invalid('2n -BigInt("-1")'),
    invalid('BigInt(-1).toString()'),
    invalid('functionCall(-BigInt(1))'),
    invalid('obj[BigInt(1)]'),
    invalid('arr[BigInt(0)]'),
    invalid('void BigInt(1)'),
    invalid('typeof BigInt(1)'),
    invalid('BigInt(1) + BigInt(2)', 2),
    invalid('BigInt(1) ** 2n'),
    invalid('condition ? BigInt(1) : BigInt(2)', 2),
    invalid('+BigInt(1)'),
    invalid('+BigInt(-1)'),
    invalid('+BigInt("1")'),
    invalid('BigInt(+0)'),
    invalid('BigInt(-0)'),
    invalid('BigInt("+0")'),
    invalid('-BigInt("-0")'),
    invalid('1n - (( BigInt("-1") ))'),
    invalid('BigInt(-(+(-10)))'),
    // Templates and escapes.
    invalid('BigInt(`1`)'),
    invalid('BigInt(``)'),
    invalid('BigInt(`  00001  `)'),
    invalid('BigInt(`0XFe`)'),
    invalid('BigInt(`-1`).toString()'),
    invalid('BigInt(`+1`)'),
    invalid('BigInt(/* comment */ `1`)'),
    invalid('BigInt("1\\\n2")'),
    invalid('BigInt(`\\u{31}`)'),
    invalid('BigInt("\\u{31}")'),
    // Leading zeros and exponent notation.
    invalid('BigInt("00001");'),
    invalid('BigInt("007");'),
    invalid('BigInt("-00001");'),
    invalid('BigInt("  00001  ");'),
    invalid('BigInt("-0");'),
    invalid('BigInt("08");'),
    invalid('BigInt("09");'),
    invalid('BigInt("080");'),
    invalid('BigInt("0000008");'),
    invalid('BigInt(1e+21);'),
    // Documentation examples.
    invalid('const bigint = BigInt(1);'),
    invalid("const large = BigInt('9007199254740991');"),
    invalid("const hex = BigInt('0xFF');"),
  ],
});

// The native API surfaces TS1121/TS1489 before running rules, even in script
// mode. The Go upstream suite checks these cases on the recovered AST.
test.skip.each(['BigInt(0777)', 'BigInt(0888)'])(
  'Legacy script literal rejected by the TypeScript parser (%s)',
  () => {},
);
