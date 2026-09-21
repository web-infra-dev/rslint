// Every test group and ignore variant from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-unsupported-features/es-builtins.js
import { RuleTester } from '../../rule-tester';

function unsupported(
  name: string,
  supported: string,
  version: string,
  line: number,
  column: number,
  endLine: number,
  endColumn: number,
) {
  return {
    messageId: 'not-supported-till',
    message: `The '${name}' is still an experimental feature and is not supported until Node.js ${supported}. The configured version range is '${version}'.`,
    line,
    column,
    endLine,
    endColumn,
  };
}

type Cases = Parameters<RuleTester['run']>[2];
const patterns: ({ keyword: string } & Cases)[] = [
  {
    keyword: 'AggregateError',
    valid: [
      {
        code: 'if (error instanceof AggregateError) {}',
        options: [{ version: '15.0.0' }],
      },
    ],
    invalid: [
      {
        code: 'if (error instanceof AggregateError) {}',
        options: [{ version: '14.0.0' }],
        errors: [
          unsupported('AggregateError', '15.0.0', '14.0.0', 1, 22, 1, 36),
        ],
      },
    ],
  },
  {
    keyword: 'Array.from',
    valid: [
      { code: 'Array.foo(a)', options: [{ version: '3.9.9' }] },
      {
        code: '(function(Array) { Array.from(a) }(b))',
        options: [{ version: '3.9.9' }],
      },
      { code: 'Array.from(a)', options: [{ version: '4.0.0' }] },
    ],
    invalid: [
      {
        code: 'Array.from(a)',
        options: [{ version: '3.9.9' }],
        errors: [unsupported('Array.from', '4.0.0', '3.9.9', 1, 1, 1, 11)],
      },
    ],
  },
  {
    keyword: 'Array.of',
    valid: [
      { code: 'Array.foo(a)', options: [{ version: '3.9.9' }] },
      {
        code: '(function(Array) { Array.of(a) }(b))',
        options: [{ version: '3.9.9' }],
      },
      { code: 'Array.of(a)', options: [{ version: '4.0.0' }] },
    ],
    invalid: [
      {
        code: 'Array.of(a)',
        options: [{ version: '3.9.9' }],
        errors: [unsupported('Array.of', '4.0.0', '3.9.9', 1, 1, 1, 9)],
      },
    ],
  },
  {
    keyword: 'BigInt',
    valid: [
      { code: 'bigint', options: [{ version: '10.3.0' }] },
      {
        code: '(function(BigInt) { BigInt }(b))',
        options: [{ version: '10.3.0' }],
      },
      { code: 'BigInt', options: [{ version: '10.4.0' }] },
    ],
    invalid: [
      {
        code: 'BigInt',
        options: [{ version: '10.3.0' }],
        errors: [unsupported('BigInt', '10.4.0', '10.3.0', 1, 1, 1, 7)],
      },
      {
        code: '(function() { BigInt })()',
        options: [{ version: '10.3.0' }],
        errors: [unsupported('BigInt', '10.4.0', '10.3.0', 1, 15, 1, 21)],
      },
    ],
  },
  {
    keyword: 'FinalizationRegistry',
    valid: [
      {
        code: 'new FinalizationRegistry(() => {})',
        options: [{ version: '14.6.0' }],
      },
    ],
    invalid: [
      {
        code: 'new FinalizationRegistry(() => {})',
        options: [{ version: '14.5.0' }],
        errors: [
          unsupported('FinalizationRegistry', '14.6.0', '14.5.0', 1, 5, 1, 25),
        ],
      },
    ],
  },
  {
    keyword: 'Map',
    valid: [
      { code: 'map', options: [{ version: '0.11.9' }] },
      { code: '(function(Map) { Map }(b))', options: [{ version: '0.11.9' }] },
      { code: 'Map', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Map',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Map', '0.12.0', '0.11.9', 1, 1, 1, 4)],
      },
      {
        code: '(function() { Map })()',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Map', '0.12.0', '0.11.9', 1, 15, 1, 18)],
      },
    ],
  },
  {
    keyword: 'Math.acosh',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.acosh(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.acosh(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.acosh(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.acosh', '0.12.0', '0.11.9', 1, 1, 1, 11)],
      },
    ],
  },
  {
    keyword: 'Math.asinh',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.asinh(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.asinh(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.asinh(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.asinh', '0.12.0', '0.11.9', 1, 1, 1, 11)],
      },
    ],
  },
  {
    keyword: 'Math.atanh',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.atanh(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.atanh(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.atanh(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.atanh', '0.12.0', '0.11.9', 1, 1, 1, 11)],
      },
    ],
  },
  {
    keyword: 'Math.cbrt',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.cbrt(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.cbrt(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.cbrt(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.cbrt', '0.12.0', '0.11.9', 1, 1, 1, 10)],
      },
    ],
  },
  {
    keyword: 'Math.clz32',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.clz32(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.clz32(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.clz32(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.clz32', '0.12.0', '0.11.9', 1, 1, 1, 11)],
      },
    ],
  },
  {
    keyword: 'Math.cosh',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.cosh(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.cosh(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.cosh(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.cosh', '0.12.0', '0.11.9', 1, 1, 1, 10)],
      },
    ],
  },
  {
    keyword: 'Math.expm1',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.expm1(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.expm1(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.expm1(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.expm1', '0.12.0', '0.11.9', 1, 1, 1, 11)],
      },
    ],
  },
  {
    keyword: 'Math.fround',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.fround(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.fround(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.fround(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.fround', '0.12.0', '0.11.9', 1, 1, 1, 12)],
      },
    ],
  },
  {
    keyword: 'Math.hypot',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.hypot(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.hypot(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.hypot(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.hypot', '0.12.0', '0.11.9', 1, 1, 1, 11)],
      },
    ],
  },
  {
    keyword: 'Math.imul',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.imul(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.imul(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.imul(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.imul', '0.12.0', '0.11.9', 1, 1, 1, 10)],
      },
    ],
  },
  {
    keyword: 'Math.log10',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.log10(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.log10(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.log10(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.log10', '0.12.0', '0.11.9', 1, 1, 1, 11)],
      },
    ],
  },
  {
    keyword: 'Math.log1p',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.log1p(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.log1p(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.log1p(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.log1p', '0.12.0', '0.11.9', 1, 1, 1, 11)],
      },
    ],
  },
  {
    keyword: 'Math.log2',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.log2(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.log2(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.log2(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.log2', '0.12.0', '0.11.9', 1, 1, 1, 10)],
      },
    ],
  },
  {
    keyword: 'Math.sign',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.sign(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.sign(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.sign(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.sign', '0.12.0', '0.11.9', 1, 1, 1, 10)],
      },
    ],
  },
  {
    keyword: 'Math.sinh',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.sinh(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.sinh(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.sinh(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.sinh', '0.12.0', '0.11.9', 1, 1, 1, 10)],
      },
    ],
  },
  {
    keyword: 'Math.tanh',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.tanh(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.tanh(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.tanh(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.tanh', '0.12.0', '0.11.9', 1, 1, 1, 10)],
      },
    ],
  },
  {
    keyword: 'Math.trunc',
    valid: [
      { code: 'Math.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Math) { Math.trunc(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Math.trunc(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Math.trunc(a)',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Math.trunc', '0.12.0', '0.11.9', 1, 1, 1, 11)],
      },
    ],
  },
  {
    keyword: 'Number.isFinite',
    valid: [
      { code: 'Number.foo(a)', options: [{ version: '0.9.9' }] },
      {
        code: '(function(Number) { Number.isFinite(a) }(b))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Number.isFinite(a)', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Number.isFinite(a)',
        options: [{ version: '0.9.9' }],
        errors: [
          unsupported('Number.isFinite', '0.10.0', '0.9.9', 1, 1, 1, 16),
        ],
      },
    ],
  },
  {
    keyword: 'Number.isInteger',
    valid: [
      { code: 'Number.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Number) { Number.isInteger(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Number.isInteger(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Number.isInteger(a)',
        options: [{ version: '0.11.9' }],
        errors: [
          unsupported('Number.isInteger', '0.12.0', '0.11.9', 1, 1, 1, 17),
        ],
      },
    ],
  },
  {
    keyword: 'Number.isNaN',
    valid: [
      { code: 'Number.foo(a)', options: [{ version: '0.9.9' }] },
      {
        code: '(function(Number) { Number.isNaN(a) }(b))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Number.isNaN(a)', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Number.isNaN(a)',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Number.isNaN', '0.10.0', '0.9.9', 1, 1, 1, 13)],
      },
    ],
  },
  {
    keyword: 'Number.isSafeInteger',
    valid: [
      { code: 'Number.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Number) { Number.isSafeInteger(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Number.isSafeInteger(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Number.isSafeInteger(a)',
        options: [{ version: '0.11.9' }],
        errors: [
          unsupported('Number.isSafeInteger', '0.12.0', '0.11.9', 1, 1, 1, 21),
        ],
      },
    ],
  },
  {
    keyword: 'Number.parseFloat',
    valid: [
      { code: 'Number.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Number) { Number.parseFloat(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Number.parseFloat(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Number.parseFloat(a)',
        options: [{ version: '0.11.9' }],
        errors: [
          unsupported('Number.parseFloat', '0.12.0', '0.11.9', 1, 1, 1, 18),
        ],
      },
    ],
  },
  {
    keyword: 'Number.parseInt',
    valid: [
      { code: 'Number.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Number) { Number.parseInt(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Number.parseInt(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Number.parseInt(a)',
        options: [{ version: '0.11.9' }],
        errors: [
          unsupported('Number.parseInt', '0.12.0', '0.11.9', 1, 1, 1, 16),
        ],
      },
    ],
  },
  {
    keyword: 'Object.assign',
    valid: [
      { code: 'Object.foo(a)', options: [{ version: '3.9.9' }] },
      {
        code: '(function(Object) { Object.assign(a) }(b))',
        options: [{ version: '3.9.9' }],
      },
      { code: 'Object.assign(a)', options: [{ version: '4.0.0' }] },
    ],
    invalid: [
      {
        code: 'Object.assign(a)',
        options: [{ version: '3.9.9' }],
        errors: [unsupported('Object.assign', '4.0.0', '3.9.9', 1, 1, 1, 14)],
      },
    ],
  },
  {
    keyword: 'Object.getOwnPropertySymbols',
    valid: [
      { code: 'Object.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Object) { Object.getOwnPropertySymbols(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      {
        code: 'Object.getOwnPropertySymbols(a)',
        options: [{ version: '0.12.0' }],
      },
    ],
    invalid: [
      {
        code: 'Object.getOwnPropertySymbols(a)',
        options: [{ version: '0.11.9' }],
        errors: [
          unsupported(
            'Object.getOwnPropertySymbols',
            '0.12.0',
            '0.11.9',
            1,
            1,
            1,
            29,
          ),
        ],
      },
    ],
  },
  {
    keyword: 'Object.is',
    valid: [
      { code: 'Object.foo(a)', options: [{ version: '0.9.9' }] },
      {
        code: '(function(Object) { Object.is(a) }(b))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Object.is(a)', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Object.is(a)',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Object.is', '0.10.0', '0.9.9', 1, 1, 1, 10)],
      },
    ],
  },
  {
    keyword: 'Object.setPrototypeOf',
    valid: [
      { code: 'Object.foo(a)', options: [{ version: '0.11.9' }] },
      {
        code: '(function(Object) { Object.setPrototypeOf(a) }(b))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Object.setPrototypeOf(a)', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Object.setPrototypeOf(a)',
        options: [{ version: '0.11.9' }],
        errors: [
          unsupported('Object.setPrototypeOf', '0.12.0', '0.11.9', 1, 1, 1, 22),
        ],
      },
    ],
  },
  {
    keyword: 'Promise',
    valid: [
      {
        code: '(function(Promise) { Promise }(a))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Promise', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Promise',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Promise', '0.12.0', '0.11.9', 1, 1, 1, 8)],
      },
      {
        code: 'function wrap() { Promise }',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Promise', '0.12.0', '0.11.9', 1, 19, 1, 26)],
      },
    ],
  },
  {
    keyword: 'Promise.allSettled',
    valid: [
      {
        code: '(function(Promise) { Promise.allSettled }(a))',
        options: [{ version: '12.8.1' }],
      },
      { code: 'Promise.allSettled', options: [{ version: '12.9.0' }] },
    ],
    invalid: [
      {
        code: 'Promise.allSettled',
        options: [{ version: '12.8.1' }],
        errors: [
          unsupported('Promise.allSettled', '12.9.0', '12.8.1', 1, 1, 1, 19),
        ],
      },
      {
        code: 'function wrap() { Promise.allSettled }',
        options: [{ version: '12.8.1' }],
        errors: [
          unsupported('Promise.allSettled', '12.9.0', '12.8.1', 1, 19, 1, 37),
        ],
      },
    ],
  },
  {
    keyword: 'Promise.any',
    valid: [
      {
        code: '(function(Promise) { Promise.any }(a))',
        options: [{ version: '14.0.0' }],
      },
      { code: 'Promise.any', options: [{ version: '15.0.0' }] },
    ],
    invalid: [
      {
        code: 'Promise.any',
        options: [{ version: '14.0.0' }],
        errors: [unsupported('Promise.any', '15.0.0', '14.0.0', 1, 1, 1, 12)],
      },
      {
        code: 'function wrap() { Promise.any }',
        options: [{ version: '14.0.0' }],
        errors: [unsupported('Promise.any', '15.0.0', '14.0.0', 1, 19, 1, 30)],
      },
    ],
  },
  {
    keyword: 'Proxy',
    valid: [
      {
        code: '(function(Proxy) { Proxy }(a))',
        options: [{ version: '5.9.9' }],
      },
      { code: 'Proxy', options: [{ version: '6.0.0' }] },
    ],
    invalid: [
      {
        code: 'Proxy',
        options: [{ version: '5.9.9' }],
        errors: [unsupported('Proxy', '6.0.0', '5.9.9', 1, 1, 1, 6)],
      },
      {
        code: 'function wrap() { Proxy }',
        options: [{ version: '5.9.9' }],
        errors: [unsupported('Proxy', '6.0.0', '5.9.9', 1, 19, 1, 24)],
      },
    ],
  },
  {
    keyword: 'Reflect',
    valid: [
      {
        code: '(function(Reflect) { Reflect }(a))',
        options: [{ version: '5.9.9' }],
      },
      { code: 'Reflect', options: [{ version: '6.0.0' }] },
    ],
    invalid: [
      {
        code: 'Reflect',
        options: [{ version: '5.9.9' }],
        errors: [unsupported('Reflect', '6.0.0', '5.9.9', 1, 1, 1, 8)],
      },
      {
        code: 'function wrap() { Reflect }',
        options: [{ version: '5.9.9' }],
        errors: [unsupported('Reflect', '6.0.0', '5.9.9', 1, 19, 1, 26)],
      },
    ],
  },
  {
    keyword: 'Set',
    valid: [
      { code: '(function(Set) { Set }(a))', options: [{ version: '0.11.9' }] },
      { code: 'Set', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Set',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Set', '0.12.0', '0.11.9', 1, 1, 1, 4)],
      },
      {
        code: 'function wrap() { Set }',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Set', '0.12.0', '0.11.9', 1, 19, 1, 22)],
      },
    ],
  },
  {
    keyword: 'String.fromCodePoint',
    valid: [
      { code: 'String.foo(a)', options: [{ version: '3.9.9' }] },
      {
        code: '(function(String) { String.fromCodePoint(a) }(b))',
        options: [{ version: '3.9.9' }],
      },
      { code: 'String.fromCodePoint(a)', options: [{ version: '4.0.0' }] },
    ],
    invalid: [
      {
        code: 'String.fromCodePoint(a)',
        options: [{ version: '3.9.9' }],
        errors: [
          unsupported('String.fromCodePoint', '4.0.0', '3.9.9', 1, 1, 1, 21),
        ],
      },
    ],
  },
  {
    keyword: 'String.raw',
    valid: [
      { code: 'String.foo(a)', options: [{ version: '3.9.9' }] },
      {
        code: '(function(String) { String.raw(a) }(b))',
        options: [{ version: '3.9.9' }],
      },
      { code: 'String.raw(a)', options: [{ version: '4.0.0' }] },
    ],
    invalid: [
      {
        code: 'String.raw(a)',
        options: [{ version: '3.9.9' }],
        errors: [unsupported('String.raw', '4.0.0', '3.9.9', 1, 1, 1, 11)],
      },
    ],
  },
  {
    keyword: 'Symbol',
    valid: [
      {
        code: '(function(Symbol) { Symbol }(a))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'Symbol', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'Symbol',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Symbol', '0.12.0', '0.11.9', 1, 1, 1, 7)],
      },
      {
        code: 'function wrap() { Symbol }',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('Symbol', '0.12.0', '0.11.9', 1, 19, 1, 25)],
      },
    ],
  },
  {
    keyword: 'Int8Array',
    valid: [
      {
        code: '(function(Int8Array) { Int8Array }(a))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Int8Array', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Int8Array',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Int8Array', '0.10.0', '0.9.9', 1, 1, 1, 10)],
      },
      {
        code: 'function wrap() { Int8Array }',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Int8Array', '0.10.0', '0.9.9', 1, 19, 1, 28)],
      },
    ],
  },
  {
    keyword: 'Uint8Array',
    valid: [
      {
        code: '(function(Uint8Array) { Uint8Array }(a))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Uint8Array', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Uint8Array',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Uint8Array', '0.10.0', '0.9.9', 1, 1, 1, 11)],
      },
      {
        code: 'function wrap() { Uint8Array }',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Uint8Array', '0.10.0', '0.9.9', 1, 19, 1, 29)],
      },
    ],
  },
  {
    keyword: 'Uint8ClampedArray',
    valid: [
      {
        code: '(function(Uint8ClampedArray) { Uint8ClampedArray }(a))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Uint8ClampedArray', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Uint8ClampedArray',
        options: [{ version: '0.9.9' }],
        errors: [
          unsupported('Uint8ClampedArray', '0.10.0', '0.9.9', 1, 1, 1, 18),
        ],
      },
      {
        code: 'function wrap() { Uint8ClampedArray }',
        options: [{ version: '0.9.9' }],
        errors: [
          unsupported('Uint8ClampedArray', '0.10.0', '0.9.9', 1, 19, 1, 36),
        ],
      },
    ],
  },
  {
    keyword: 'Int16Array',
    valid: [
      {
        code: '(function(Int16Array) { Int16Array }(a))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Int16Array', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Int16Array',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Int16Array', '0.10.0', '0.9.9', 1, 1, 1, 11)],
      },
      {
        code: 'function wrap() { Int16Array }',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Int16Array', '0.10.0', '0.9.9', 1, 19, 1, 29)],
      },
    ],
  },
  {
    keyword: 'Uint16Array',
    valid: [
      {
        code: '(function(Uint16Array) { Uint16Array }(a))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Uint16Array', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Uint16Array',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Uint16Array', '0.10.0', '0.9.9', 1, 1, 1, 12)],
      },
      {
        code: 'function wrap() { Uint16Array }',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Uint16Array', '0.10.0', '0.9.9', 1, 19, 1, 30)],
      },
    ],
  },
  {
    keyword: 'Int32Array',
    valid: [
      {
        code: '(function(Int32Array) { Int32Array }(a))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Int32Array', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Int32Array',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Int32Array', '0.10.0', '0.9.9', 1, 1, 1, 11)],
      },
      {
        code: 'function wrap() { Int32Array }',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Int32Array', '0.10.0', '0.9.9', 1, 19, 1, 29)],
      },
    ],
  },
  {
    keyword: 'Uint32Array',
    valid: [
      {
        code: '(function(Uint32Array) { Uint32Array }(a))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Uint32Array', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Uint32Array',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Uint32Array', '0.10.0', '0.9.9', 1, 1, 1, 12)],
      },
      {
        code: 'function wrap() { Uint32Array }',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Uint32Array', '0.10.0', '0.9.9', 1, 19, 1, 30)],
      },
    ],
  },
  {
    keyword: 'BigInt64Array',
    valid: [
      {
        code: '(function(BigInt64Array) { BigInt64Array }(b))',
        options: [{ version: '10.3.0' }],
      },
      { code: 'BigInt64Array', options: [{ version: '10.4.0' }] },
    ],
    invalid: [
      {
        code: 'BigInt64Array',
        options: [{ version: '10.3.0' }],
        errors: [unsupported('BigInt64Array', '10.4.0', '10.3.0', 1, 1, 1, 14)],
      },
      {
        code: '(function() { BigInt64Array })()',
        options: [{ version: '10.3.0' }],
        errors: [
          unsupported('BigInt64Array', '10.4.0', '10.3.0', 1, 15, 1, 28),
        ],
      },
    ],
  },
  {
    keyword: 'BigUint64Array',
    valid: [
      {
        code: '(function(BigUint64Array) { BigUint64Array }(b))',
        options: [{ version: '10.3.0' }],
      },
      { code: 'BigUint64Array', options: [{ version: '10.4.0' }] },
    ],
    invalid: [
      {
        code: 'BigUint64Array',
        options: [{ version: '10.3.0' }],
        errors: [
          unsupported('BigUint64Array', '10.4.0', '10.3.0', 1, 1, 1, 15),
        ],
      },
      {
        code: '(function() { BigUint64Array })()',
        options: [{ version: '10.3.0' }],
        errors: [
          unsupported('BigUint64Array', '10.4.0', '10.3.0', 1, 15, 1, 29),
        ],
      },
    ],
  },
  {
    keyword: 'Float32Array',
    valid: [
      {
        code: '(function(Float32Array) { Float32Array }(a))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Float32Array', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Float32Array',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Float32Array', '0.10.0', '0.9.9', 1, 1, 1, 13)],
      },
      {
        code: 'function wrap() { Float32Array }',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Float32Array', '0.10.0', '0.9.9', 1, 19, 1, 31)],
      },
    ],
  },
  {
    keyword: 'Float64Array',
    valid: [
      {
        code: '(function(Float64Array) { Float64Array }(a))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'Float64Array', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'Float64Array',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Float64Array', '0.10.0', '0.9.9', 1, 1, 1, 13)],
      },
      {
        code: 'function wrap() { Float64Array }',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('Float64Array', '0.10.0', '0.9.9', 1, 19, 1, 31)],
      },
    ],
  },
  {
    keyword: 'DataView',
    valid: [
      {
        code: '(function(DataView) { DataView }(a))',
        options: [{ version: '0.9.9' }],
      },
      { code: 'DataView', options: [{ version: '0.10.0' }] },
    ],
    invalid: [
      {
        code: 'DataView',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('DataView', '0.10.0', '0.9.9', 1, 1, 1, 9)],
      },
      {
        code: 'function wrap() { DataView }',
        options: [{ version: '0.9.9' }],
        errors: [unsupported('DataView', '0.10.0', '0.9.9', 1, 19, 1, 27)],
      },
    ],
  },
  {
    keyword: 'WeakMap',
    valid: [
      {
        code: '(function(WeakMap) { WeakMap }(a))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'WeakMap', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'WeakMap',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('WeakMap', '0.12.0', '0.11.9', 1, 1, 1, 8)],
      },
      {
        code: 'function wrap() { WeakMap }',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('WeakMap', '0.12.0', '0.11.9', 1, 19, 1, 26)],
      },
    ],
  },
  {
    keyword: 'WeakRef',
    valid: [
      {
        code: '(function(WeakRef) { WeakRef }(a))',
        options: [{ version: '14.5.0' }],
      },
      { code: 'WeakRef', options: [{ version: '14.6.0' }] },
    ],
    invalid: [
      {
        code: 'WeakRef',
        options: [{ version: '14.5.0' }],
        errors: [unsupported('WeakRef', '14.6.0', '14.5.0', 1, 1, 1, 8)],
      },
      {
        code: 'function wrap() { WeakRef }',
        options: [{ version: '14.5.0' }],
        errors: [unsupported('WeakRef', '14.6.0', '14.5.0', 1, 19, 1, 26)],
      },
    ],
  },
  {
    keyword: 'WeakSet',
    valid: [
      {
        code: '(function(WeakSet) { WeakSet }(a))',
        options: [{ version: '0.11.9' }],
      },
      { code: 'WeakSet', options: [{ version: '0.12.0' }] },
    ],
    invalid: [
      {
        code: 'WeakSet',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('WeakSet', '0.12.0', '0.11.9', 1, 1, 1, 8)],
      },
      {
        code: 'function wrap() { WeakSet }',
        options: [{ version: '0.11.9' }],
        errors: [unsupported('WeakSet', '0.12.0', '0.11.9', 1, 19, 1, 26)],
      },
    ],
  },
  {
    keyword: 'Atomics',
    valid: [
      {
        code: '(function(Atomics) { Atomics }(a))',
        options: [{ version: '8.9.9' }],
      },
      { code: 'Atomics', options: [{ version: '8.10.0' }] },
    ],
    invalid: [
      {
        code: 'Atomics',
        options: [{ version: '8.9.9' }],
        errors: [unsupported('Atomics', '8.10.0', '8.9.9', 1, 1, 1, 8)],
      },
      {
        code: 'function wrap() { Atomics }',
        options: [{ version: '8.9.9' }],
        errors: [unsupported('Atomics', '8.10.0', '8.9.9', 1, 19, 1, 26)],
      },
    ],
  },
  {
    keyword: 'Object.values',
    valid: [
      { code: 'Object.foo(a)', options: [{ version: '6.9.9' }] },
      {
        code: '(function(Object) { Object.values(a) }(b))',
        options: [{ version: '6.9.9' }],
      },
      { code: 'Object.values(a)', options: [{ version: '7.0.0' }] },
    ],
    invalid: [
      {
        code: 'Object.values(a)',
        options: [{ version: '6.9.9' }],
        errors: [unsupported('Object.values', '7.0.0', '6.9.9', 1, 1, 1, 14)],
      },
    ],
  },
  {
    keyword: 'Object.entries',
    valid: [
      { code: 'Object.foo(a)', options: [{ version: '6.9.9' }] },
      {
        code: '(function(Object) { Object.entries(a) }(b))',
        options: [{ version: '6.9.9' }],
      },
      { code: 'Object.entries(a)', options: [{ version: '7.0.0' }] },
    ],
    invalid: [
      {
        code: 'Object.entries(a)',
        options: [{ version: '6.9.9' }],
        errors: [unsupported('Object.entries', '7.0.0', '6.9.9', 1, 1, 1, 15)],
      },
    ],
  },
  {
    keyword: 'Object.getOwnPropertyDescriptors',
    valid: [
      { code: 'Object.foo(a)', options: [{ version: '6.9.9' }] },
      {
        code: '(function(Object) { Object.getOwnPropertyDescriptors(a) }(b))',
        options: [{ version: '6.9.9' }],
      },
      {
        code: 'Object.getOwnPropertyDescriptors(a)',
        options: [{ version: '7.0.0' }],
      },
    ],
    invalid: [
      {
        code: 'Object.getOwnPropertyDescriptors(a)',
        options: [{ version: '6.9.9' }],
        errors: [
          unsupported(
            'Object.getOwnPropertyDescriptors',
            '7.0.0',
            '6.9.9',
            1,
            1,
            1,
            33,
          ),
        ],
      },
    ],
  },
  {
    keyword: 'SharedArrayBuffer',
    valid: [
      {
        code: '(function(SharedArrayBuffer) { SharedArrayBuffer }(a))',
        options: [{ version: '8.9.9' }],
      },
      { code: 'SharedArrayBuffer', options: [{ version: '8.10.0' }] },
    ],
    invalid: [
      {
        code: 'SharedArrayBuffer',
        options: [{ version: '8.9.9' }],
        errors: [
          unsupported('SharedArrayBuffer', '8.10.0', '8.9.9', 1, 1, 1, 18),
        ],
      },
      {
        code: 'function wrap() { SharedArrayBuffer }',
        options: [{ version: '8.9.9' }],
        errors: [
          unsupported('SharedArrayBuffer', '8.10.0', '8.9.9', 1, 19, 1, 36),
        ],
      },
    ],
  },
  {
    keyword: 'globalThis',
    valid: [
      {
        code: '(function(globalThis) { globalThis }(a))',
        options: [{ version: '12.0.0' }],
      },
      { code: 'globalThis', options: [{ version: '12.0.0' }] },
      { code: 'globalThis', settings: { node: { version: '12.0.0' } } },
    ],
    invalid: [
      {
        code: 'globalThis',
        options: [{ version: '11.9.9' }],
        errors: [unsupported('globalThis', '12.0.0', '11.9.9', 1, 1, 1, 11)],
      },
      {
        code: 'function wrap() { globalThis }',
        options: [{ version: '11.9.9' }],
        errors: [unsupported('globalThis', '12.0.0', '11.9.9', 1, 19, 1, 29)],
      },
      {
        code: 'function wrap() { globalThis }',
        settings: { node: { version: '11.9.9' } },
        errors: [unsupported('globalThis', '12.0.0', '11.9.9', 1, 19, 1, 29)],
      },
      {
        code: 'function wrap() { globalThis }',
        settings: { node: { version: '>=11.9.9' } },
        errors: [unsupported('globalThis', '12.0.0', '>=11.9.9', 1, 19, 1, 29)],
      },
    ],
  },
];

const ruleTester = new RuleTester();
for (const pattern of patterns) {
  const ignored = pattern.invalid.map(({ errors: _errors, ...item }) => ({
    ...item,
    options: [
      {
        ...(item.options as Record<string, unknown>[] | undefined)?.[0],
        ignores: [pattern.keyword],
      },
    ],
  }));
  ruleTester.run(
    'no-unsupported-features/es-builtins',
    {},
    {
      valid: [...pattern.valid, ...ignored].map((item) => ({
        ...item,
        name: pattern.keyword,
      })),
      invalid: pattern.invalid.map((item) => ({
        ...item,
        name: pattern.keyword,
      })),
    },
  );
}

// The two configuration examples in the pinned documentation.
for (const settings of [undefined, { node: { version: '>=16.0.0' } }]) {
  const options = [
    { ...(settings ? {} : { version: '>=16.0.0' }), ignores: [] },
  ];
  ruleTester.run(
    'no-unsupported-features/es-builtins',
    {},
    {
      valid: [
        {
          name: 'documented version configuration',
          code: 'Object.fromEntries(entries)',
          options,
          settings,
        },
      ],
      invalid: [
        {
          name: 'documented version configuration',
          code: 'Map.groupBy(items, key)',
          options,
          settings,
          errors: [
            unsupported('Map.groupBy', '21.0.0', '>=16.0.0', 1, 1, 1, 12),
          ],
        },
      ],
    },
  );
}
