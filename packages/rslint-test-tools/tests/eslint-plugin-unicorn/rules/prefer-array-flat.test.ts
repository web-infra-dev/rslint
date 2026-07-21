import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string, options?: unknown) => ({ code, options });
const invalid = (code: string, options?: unknown, errors = 1) => ({
  code,
  options,
  errors,
});

const validCases = [
  // `array.flatMap(x => x)`
  'array.flatMap',
  'new array.flatMap(x => x)',
  'flatMap(x => x)',
  'array.notFlatMap(x => x)',
  'array[flatMap](x => x)',
  'array.flatMap(x => x, thisArgument)',
  'array.flatMap(...[x => x])',
  'array.flatMap(function (x) { return x; })',
  'array.flatMap(async x => x)',
  'array.flatMap(function * (x) { return x;})',
  'array.flatMap(() => x)',
  'array.flatMap((x, y) => x)',
  'array.flatMap((x) => { return x; })',
  'array.flatMap(x => y)',

  // Obvious non-array flatMap receivers
  `const randomObject = {
    flatMap(function_) {
      function_();
    },
  };
  randomObject.flatMap(x => x);`,
  'Effects.flatMap(x => x)',
  `const effects = {
    flatMap(function_) {
      function_();
    },
  };
  effects.flatMap(x => x);`,
  'const effects = new Set(); effects.flatMap(x => x);',
  'const mapping = new Map(); mapping.flatMap(x => x);',
  'const text = ""; text.flatMap(x => x);',
  'const handler = () => {}; handler.flatMap(x => x);',
  'const collection = new Foo(); collection.flatMap(x => x);',

  // `array.reduce((a, b) => a.concat(b), [])`
  'new array.reduce((a, b) => a.concat(b), [])',
  'array.reduce',
  'reduce((a, b) => a.concat(b), [])',
  'array[reduce]((a, b) => a.concat(b), [])',
  'array.notReduce((a, b) => a.concat(b), [])',
  'array.reduce((a, b) => a.concat(b), [], EXTRA_ARGUMENT)',
  'array.reduce((a, b) => a.concat(b), NOT_EMPTY_ARRAY)',
  'array.reduce((a, b, extraParameter) => a.concat(b), [])',
  'array.reduce((a,) => a.concat(b), [])',
  'array.reduce(() => a.concat(b), [])',
  'array.reduce((a, b) => {return a.concat(b); }, [])',
  'array.reduce(function (a, b) { return a.concat(b); }, [])',
  'array.reduce((a, b) => b.concat(b), [])',
  'array.reduce((a, b) => a.concat(a), [])',
  'array.reduce((a, b) => b.concat(a), [])',
  'array.reduce((a, b) => a.notConcat(b), [])',
  'array.reduce((a, b) => a.concat, [])',

  // `array.reduce((a, b) => [...a, ...b], [])`
  'new array.reduce((a, b) => [...a, ...b], [])',
  'array[reduce]((a, b) => [...a, ...b], [])',
  'reduce((a, b) => [...a, ...b], [])',
  'array.notReduce((a, b) => [...a, ...b], [])',
  'array.reduce((a, b) => [...a, ...b], [], EXTRA_ARGUMENT)',
  'array.reduce((a, b) => [...a, ...b], NOT_EMPTY_ARRAY)',
  'array.reduce((a, b, extraParameter) => [...a, ...b], [])',
  'array.reduce((a,) => [...a, ...b], [])',
  'array.reduce(() => [...a, ...b], [])',
  'array.reduce((a, b) => {return [...a, ...b]; }, [])',
  'array.reduce(function (a, b) { return [...a, ...b]; }, [])',
  'array.reduce((a, b) => [...b, ...b], [])',
  'array.reduce((a, b) => [...a, ...a], [])',
  'array.reduce((a, b) => [...b, ...a], [])',
  'array.reduce((a, b) => [a, ...b], [])',
  'array.reduce((a, b) => [...a, b], [])',
  'array.reduce((a, b) => [a, b], [])',
  'array.reduce((a, b) => [...a, ...b, c], [])',
  'array.reduce((a, b) => [...a, ...b,,], [])',
  'array.reduce((a, b) => [,...a, ...b], [])',
  'array.reduce((a, b) => [, ], [])',
  'array.reduce((a, b) => [, ,], [])',

  // `[].concat(array)`
  '[].concat',
  'new [].concat(array)',
  '[][concat](array)',
  '[].notConcat(array)',
  '[,].concat(array)',
  '({}).concat(array)',
  '[].concat()',
  '[].concat(array, EXTRA_ARGUMENT)',
  '[]?.concat(array)',
  '[].concat?.(array)',

  // `[].concat(...array)`
  'new [].concat(...array)',
  '[][concat](...array)',
  '[].notConcat(...array)',
  '[,].concat(...array)',
  '({}).concat(...array)',
  '[].concat()',
  '[].concat(...array, EXTRA_ARGUMENT)',
  '[]?.concat(...array)',
  '[].concat?.(...array)',

  // `[].concat.{apply,call}`
  'new [].concat.apply([], array)',
  '[].concat.apply',
  '[].concat.apply([], ...array)',
  '[].concat.apply([], array, EXTRA_ARGUMENT)',
  '[].concat.apply([])',
  '[].concat.apply(NOT_EMPTY_ARRAY, array)',
  '[].concat.apply([,], array)',
  '[,].concat.apply([], array)',
  '[].concat[apply]([], array)',
  '[][concat].apply([], array)',
  '[].concat.notApply([], array)',
  '[].notConcat.apply([], array)',
  '[].concat.apply?.([], array)',
  '[].concat?.apply([], array)',
  '[]?.concat.apply([], array)',

  // `Array.prototype.concat.{apply,call}`
  'new Array.prototype.concat.apply([], array)',
  'Array.prototype.concat.apply',
  'Array.prototype.concat.apply([], ...array)',
  'Array.prototype.concat.apply([], array, EXTRA_ARGUMENT)',
  'Array.prototype.concat.apply([])',
  'Array.prototype.concat.apply(NOT_EMPTY_ARRAY, array)',
  'Array.prototype.concat.apply([,], array)',
  'Array.prototype.concat[apply]([], array)',
  'Array.prototype[concat].apply([], array)',
  'Array[prototype].concat.apply([], array)',
  'Array.prototype.concat.notApply([], array)',
  'Array.prototype.notConcat.apply([], array)',
  'Array.notPrototype.concat.apply([], array)',
  'NotArray.prototype.concat.apply([], array)',
  'Array.prototype.concat.apply?.([], array)',
  'Array.prototype.concat?.apply([], array)',
  'Array.prototype?.concat.apply([], array)',
  'Array?.prototype.concat.apply([], array)',
  'object.Array.prototype.concat.apply([], array)',

  // `_.flatten(array)`
  'new _.flatten(array)',
  '_.flatten',
  '_.flatten(array, EXTRA_ARGUMENT)',
  '_.flatten(...array)',
  '_[flatten](array)',
  '_.notFlatten(array)',
  'NOT_LODASH.flatten(array)',
  '_.flatten?.(array)',
  '_?.flatten(array)',
  'object._.flatten(array)',

  'array.flat()',
  'array.flat(1)',
].map((code) => valid(code));

const invalidCases = [
  // `array.flatMap(x => x)`
  'array.flatMap(x => x)',
  'array?.flatMap(x => x)',
  'function foo(){return[].flatMap(x => x)}',
  'foo.flatMap(x => x)instanceof Array',
  'array.flatMap((x) => x)',
  'Foo.bar.flatMap(x => x)',
  'const values = getValues(); values.flatMap(x => x);',
  'const values = []; values.flatMap(x => x);',
  'const Items = []; Items.flatMap(x => x);',
  `for (const value of values) {
    value.flatMap(x => x);
  }`,

  // `array.reduce((a, b) => a.concat(b), [])`
  'array.reduce((a, b) => a.concat(b), [])',
  'array?.reduce((a, b) => a.concat(b), [])',
  'function foo(){return[].reduce((a, b) => a.concat(b), [])}',
  'function foo(){return[]?.reduce((a, b) => a.concat(b), [])}',

  // `array.reduce((a, b) => [...a, ...b], [])`
  'array.reduce((a, b) => [...a, ...b], [])',
  'array.reduce((a, b) => [...a, ...b,], [])',
  'function foo(){return[].reduce((a, b) => [...a, ...b,], [])}',

  // `[].concat(array)`
  '[].concat(maybeArray)',
  '[].concat( ((0, maybeArray)) )',
  '[].concat( ((maybeArray)) )',
  '[].concat( [foo] )',
  '[].concat( [[foo]] )',
  'function foo(){return[].concat(maybeArray)}',

  // `[].concat(...array)`
  '[].concat(...array)',
  '[].concat(...(( array )))',
  '[].concat(...(( [foo] )))',
  '[].concat(...(( [[foo]] )))',
  'function foo(){return[].concat(...array)}',
  'class A extends[].concat(...array){}',
  'const A = class extends[].concat(...array){}',

  // `[].concat.{apply,call}`
  '[].concat.apply([], array)',
  '[].concat.apply([], ((0, array)))',
  '[].concat.apply([], ((array)))',
  '[].concat.apply([], [foo])',
  '[].concat.apply([], [[foo]])',
  '[].concat.call([], maybeArray)',
  '[].concat.call([], ((0, maybeArray)))',
  '[].concat.call([], ((maybeArray)))',
  '[].concat.call([], [foo])',
  '[].concat.call([], [[foo]])',
  '[].concat.call([], ...array)',
  '[].concat.call([], ...((0, array)))',
  '[].concat.call([], ...((array)))',
  '[].concat.call([], ...[foo])',
  '[].concat.call([], ...[[foo]])',
  'function foo(){return[].concat.call([], ...array)}',

  // `Array.prototype.concat.{apply,call}`
  'Array.prototype.concat.apply([], array)',
  'Array.prototype.concat.apply([], ((0, array)))',
  'Array.prototype.concat.apply([], ((array)))',
  'Array.prototype.concat.apply([], [foo])',
  'Array.prototype.concat.apply([], [[foo]])',
  'Array.prototype.concat.call([], maybeArray)',
  'Array.prototype.concat.call([], ((0, maybeArray)))',
  'Array.prototype.concat.call([], ((maybeArray)))',
  'Array.prototype.concat.call([], [foo])',
  'Array.prototype.concat.call([], [[foo]])',
  'Array.prototype.concat.call([], ...array)',
  'Array.prototype.concat.call([], ...((0, array)))',
  'Array.prototype.concat.call([], ...((array)))',
  'Array.prototype.concat.call([], ...[foo])',
  'Array.prototype.concat.call([], ...[[foo]])',

  // #1146
  '/**/[].concat.apply([], array)',
  'Array.prototype.concat.apply([], array)',

  // Lodash/Underscore
  '_.flatten(array)',
  'lodash.flatten(array)',
  'underscore.flatten(array)',

  // Fix boundaries
  `before()
  Array.prototype.concat.apply([], [array].concat(array))`,
  `before()
  Array.prototype.concat.apply([], +1)`,
  `before()
  Array.prototype.concat.call([], +1)`,
  'Array.prototype.concat.apply([], (0, array))',
  'Array.prototype.concat.call([], (0, array))',
  'async function a() { return [].concat(await getArray()); }',
  '_.flatten((0, array))',
  'async function a() { return _.flatten(await getArray()); }',
  'async function a() { return _.flatten((await getArray())); }',
  `before()
  Array.prototype.concat.apply([], 1)`,
  `before()
  Array.prototype.concat.call([], 1)`,
  `before()
  Array.prototype.concat.apply([], 1.)`,
  `before()
  Array.prototype.concat.call([], 1.)`,
  `before()
  Array.prototype.concat.apply([], .1)`,
  `before()
  Array.prototype.concat.call([], .1)`,
  `before()
  Array.prototype.concat.apply([], 1.0)`,
  `before()
  Array.prototype.concat.call([], 1.0)`,
  '[].concat(some./**/array)',
  '[/**/].concat(some./**/array)',
  '[/**/].concat(some.array)',
].map((code) => invalid(code));

const options = [
  { functions: ['flat', 'utils.flat', 'globalThis.lodash.flatten'] },
];

validCases.push(
  ...[
    'flat',
    'new flat(array)',
    'flat?.(array)',
    'object.flat?.(array)',
    'utils.flat',
    'new utils.flat(array)',
    'utils.flat?.(array)',
    'utils?.flat(array)',
    'utils.flat2(array)',
    'utils2.flat(array)',
    'object.utils.flat(array)',
    'globalThis.lodash.flatten',
    'new globalThis.lodash.flatten(array)',
    'globalThis.lodash.flatten?.(array)',
    'globalThis.lodash?.flatten(array)',
    'globalThis?.lodash.flatten(array)',
    'object.globalThis.lodash.flatten(array)',
    'GLOBALTHIS.lodash.flatten(array)',
    'globalthis.lodash.flatten(array)',
    'GLOBALTHIS.LODASH.FLATTEN(array)',
    'flat(array, EXTRA_ARGUMENT)',
    'flat(...array)',
  ].map((code) => valid(code, options)),
);

invalidCases.push(
  ...[
    'flat(array)',
    'flat(array,)',
    'utils.flat(array)',
    'globalThis.lodash.flatten(array)',
    `import {flatten as flat} from 'lodash-es';
    const foo = flat(bar);`,
    '_.flatten(array).length',
    'Array.prototype.concat.apply([], array)',
  ].map((code) => invalid(code, options)),
  invalid('flat(array).map(array => utils.flat(array))', options, 2),
);

const spacesInFunctions = [
  {
    functions: [
      '',
      ' ',
      ' flat1 ',
      'utils..flat2',
      'utils . flat3',
      'utils.fl at4',
      'utils.flat5  ',
      '  utils.flat6',
    ],
  },
];

validCases.push(
  ...['utils.flat2(x)', 'utils.flat3(x)', 'utils.flat4(x)'].map((code) =>
    valid(code, spacesInFunctions),
  ),
);

invalidCases.push(
  ...['flat1(x)', 'utils.flat5(x)', 'utils.flat6(x)'].map((code) =>
    invalid(code, spacesInFunctions),
  ),
);

ruleTester.run('prefer-array-flat', null as never, {
  valid: validCases,
  invalid: invalidCases,
});
