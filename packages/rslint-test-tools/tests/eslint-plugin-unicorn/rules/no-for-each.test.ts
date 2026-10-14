import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string, filename?: string) => ({ code, filename });
const invalid = (code: string, errors = 1, filename?: string) => ({
  code,
  filename,
  errors: Array.from({ length: errors }, () => ({
    messageId: 'no-for-each/error',
  })),
});

// Mirrors the semantic groups in eslint-plugin-unicorn v77.0.0. Exact edit,
// suggestion, range, and TypeScript receiver behavior is asserted by the Go
// upstream/extras suites; this file verifies registration and the IPC shape.
ruleTester.run('no-for-each', null as never, {
  valid: [
    valid('new foo.forEach(element => bar())'),
    valid('forEach(element => bar())'),
    valid('foo.notForEach(element => bar())'),
    valid('React.Children.forEach(children, child => {})'),
    valid('Children.forEach(children, child => {})'),
    valid('await pIteration.forEach(plugins, async plugin => {})'),
    valid('Effect.forEach([1, 2, 3], n => Effect.succeed(n))'),
    valid('const map = new Map(); map.forEach(value => use(value))'),
    valid('const set = new Set(); set.forEach(value => use(value))'),
    valid(
      "import * as CB from 'strict-callbag-basics'; CB.forEach(value => use(value))",
    ),
    valid('foo["forEach"](value => use(value))'),
    valid(
      'function foo(map: Map<string, string>) { map.forEach(value => use(value)) }',
      'file.ts',
    ),
    valid(
      'function foo(set: ReadonlySet<string>) { set.forEach(value => use(value)) }',
      'file.ts',
    ),
    valid(
      'type Array<T> = Map<T, T>; function foo(value: Array<string>) { value.forEach(x => use(x)) }',
      'file.ts',
    ),
  ],
  invalid: [
    invalid('foo.forEach?.(element => bar(element))'),
    invalid('foo.forEach(element => bar(element), thisArgument)'),
    invalid('foo.forEach()'),
    invalid('const result = foo.forEach(element => bar(element))'),
    invalid('foo.forEach(bar)'),
    invalid('foo.forEach(async function (element) {})'),
    invalid('foo.forEach(function * (element) {})'),
    invalid('foo.forEach(() => bar())'),
    invalid('foo.forEach((element, index, array) => bar())'),
    invalid('property.forEach(({property}) => bar(property))'),
    invalid('foo.forEach((element = {}) => call(element))'),
    invalid('foo.forEach((...args) => bar(...args))'),
    invalid('[1, 2, 3].forEach(element => bar(element))'),
    invalid('const array = []; array.forEach(element => bar(element));'),
    invalid(
      'const array = []; array.forEach((element, index) => bar(element, index));',
    ),
    invalid(
      'const array = []; array.forEach(element => { element = next(element); });',
    ),
    invalid(
      'const array = []; array.forEach(element => { for (const item of element) { return; } });',
    ),
    invalid(
      'const typedArray = new Uint8Array(); typedArray.forEach(value => use(value));',
    ),
    invalid(
      'function foo(array: Array<string>) { array.forEach(value => use(value)); }',
      1,
      'file.ts',
    ),
    invalid(
      'function foo(array: ReadonlyArray<string>) { array.forEach(value => use(value)); }',
      1,
      'file.ts',
    ),
    invalid(
      'function foo(array: [string, string]) { array.forEach(value => use(value)); }',
      1,
      'file.ts',
    ),
    invalid('const array = []; array?.forEach(value => use(value));'),
    invalid('[foo()]?.forEach(value => use(value));'),
    invalid(
      'const arrays = [[1]]; arrays.forEach(array => array.forEach(value => use(value)));',
      2,
    ),
    invalid('React.NotChildren.forEach(callback)'),
    invalid('NotReact.Children.forEach(callback)'),
  ],
});
