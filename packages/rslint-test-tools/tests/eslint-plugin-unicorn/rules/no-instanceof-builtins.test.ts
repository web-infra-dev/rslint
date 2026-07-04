import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const lines = (...parts: string[]) => parts.join('\n');
const js = (code: string) => ({ code, filename: 'file.js' });
const invalid = (code: string, options?: Record<string, unknown>) => ({
  ...js(code),
  ...(options ? { options: [options] } : {}),
  errors: [
    {
      message:
        'Avoid using `instanceof` for type checking as it can lead to unreliable results.',
    },
  ],
});

const looseStrategyInvalid = [
  'foo instanceof String',
  'foo instanceof Number',
  'foo instanceof Boolean',
  'foo instanceof BigInt',
  'foo instanceof Symbol',
  'foo instanceof Function',
  'foo instanceof Array',
];

const strictStrategyInvalid = [
  'foo instanceof Error',
  'foo instanceof EvalError',
  'foo instanceof RangeError',
  'foo instanceof ReferenceError',
  'foo instanceof SyntaxError',
  'foo instanceof TypeError',
  'foo instanceof URIError',
  'foo instanceof AggregateError',
  'foo instanceof SuppressedError',
  'foo instanceof Map',
  'foo instanceof Set',
  'foo instanceof WeakMap',
  'foo instanceof WeakRef',
  'foo instanceof WeakSet',
  'foo instanceof ArrayBuffer',
  'foo instanceof Int8Array',
  'foo instanceof Uint8Array',
  'foo instanceof Uint8ClampedArray',
  'foo instanceof Int16Array',
  'foo instanceof Uint16Array',
  'foo instanceof Int32Array',
  'foo instanceof Uint32Array',
  'foo instanceof Float16Array',
  'foo instanceof Float32Array',
  'foo instanceof Float64Array',
  'foo instanceof BigInt64Array',
  'foo instanceof BigUint64Array',
  'foo instanceof Object',
  'foo instanceof RegExp',
  'foo instanceof Promise',
  'foo instanceof Proxy',
  'foo instanceof DataView',
  'foo instanceof Date',
  'foo instanceof SharedArrayBuffer',
  'foo instanceof FinalizationRegistry',
];

ruleTester.run('no-instanceof-builtins', undefined as never, {
  valid: [
    js('fooLoose instanceof WebWorker'),
    ...strictStrategyInvalid.map((code) => js(code.replace('foo', 'fooLoose'))),
    {
      ...js('fooExclude instanceof Function'),
      options: [{ exclude: ['Function'] }],
    },
    {
      ...js('fooExclude instanceof Array'),
      options: [{ exclude: ['Array'] }],
    },
    {
      ...js('fooExclude instanceof String'),
      options: [{ exclude: ['String'] }],
    },
    js('Array.isArray(arr)'),
    js('arr instanceof array'),
    js("a instanceof 'array'"),
    js('a instanceof ArrayA'),
    js('a.x[2] instanceof foo()'),
    js('Array.isArray([1,2,3]) === true'),
    js('"arr instanceof Array"'),
  ],
  invalid: [
    ...looseStrategyInvalid.map((code) => invalid(code)),
    ...[...looseStrategyInvalid, ...strictStrategyInvalid].map((code) =>
      invalid(code.replace('foo', 'fooStrict'), { strategy: 'strict' }),
    ),
    invalid('fooErr instanceof Error', {
      useErrorIsError: true,
      strategy: 'loose',
    }),
    invalid('(fooErr) instanceof (Error)', {
      useErrorIsError: true,
      strategy: 'loose',
    }),
    ...[
      'err instanceof Error',
      'err instanceof EvalError',
      'err instanceof RangeError',
      'err instanceof ReferenceError',
      'err instanceof SyntaxError',
      'err instanceof TypeError',
      'err instanceof URIError',
      'err instanceof AggregateError',
      'err instanceof SuppressedError',
    ].map((code) =>
      invalid(code, { useErrorIsError: true, strategy: 'strict' }),
    ),
    invalid('fooInclude instanceof WebWorker', { include: ['WebWorker'] }),
    invalid('fooInclude instanceof HTMLElement', { include: ['HTMLElement'] }),
    invalid('arr instanceof Array'),
    invalid('[] instanceof Array'),
    invalid('[1,2,3] instanceof Array === true'),
    invalid('fun.call(1, 2, 3) instanceof Array'),
    invalid('obj.arr instanceof Array'),
    invalid('foo.bar[2] instanceof Array'),
    invalid('(0, array) instanceof Array'),
    invalid('function foo(){return[]instanceof Array}'),
    invalid(
      lines(
        '(',
        '\t// comment',
        '\t((',
        '\t\t// comment',
        '\t\t(',
        '\t\t\t// comment',
        '\t\t\tfoo',
        '\t\t\t// comment',
        '\t\t)',
        '\t\t// comment',
        '\t))',
        '\t// comment',
        ')',
        '// comment before instanceof\r      instanceof',
        '',
        '// comment after instanceof',
        '',
        '(',
        '\t// comment',
        '',
        '\t(',
        '',
        '\t\t// comment',
        '',
        '\t\tArray',
        '',
        '\t\t// comment',
        '\t)',
        '',
        '\t\t// comment',
        ')',
        '',
        '\t// comment',
      ),
    ),
  ],
});
