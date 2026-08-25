import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = 'Do not use a reference value as the fill value.';
const js = (code: string) => ({ code, filename: 'file.js' });
const ts = (code: string) => ({ code, filename: 'file.ts' });
const invalidJs = (code: string) => ({
  code,
  filename: 'file.js',
  errors: [{ message }],
});
const invalidTs = (code: string) => ({
  code,
  filename: 'file.ts',
  errors: [{ message }],
});

ruleTester.run('no-array-fill-with-reference-type', null as never, {
  valid: [
    js('array.fill(0)'),
    js('array.fill("x")'),
    js('array.fill(false)'),
    js('array.fill(null)'),
    js('array.fill(undefined)'),
    js('array.fill(10n)'),
    js('array.fill(Symbol("x"))'),
    js('array.fill(Symbol.for("x"))'),
    js('array.fill(Symbol.iterator)'),
    js('array.fill(`x`)'),
    js('array.fill(() => {})'),
    js('array.fill(function () {})'),
    js('array.fill(/x/)'),
    js('array.fill(new RegExp("x"))'),
    js('const value = new RegExp("x"); array.fill(value)'),
    js('let value = {}; value = 1; array.fill(value)'),
    js('var value = {}; array.fill(value)'),
    js('const value = {}; const alias = value; array.fill(alias)'),
    js('array.fill(object.value)'),
    js('array.fill(this.value)'),
    js('"x".fill({})'),
    js('array?.fill({})'),
    js('array.fill?.({})'),
    js('array[fill]({})'),
    js('Array.from({length: 3}, () => value)'),
    js('Array.from({length: 3}).map(() => value)'),
    js('const {value = {}} = object; array.fill(value)'),
    js('function foo(value = {}) { array.fill(value); }'),
    ts('array.fill(/x/ as RegExp)'),
    ts('array.fill((() => {}) as Function)'),
    ts('array.fill(object.value as Foo)'),
    ts('const value = {}; const alias = value as Foo; array.fill(alias)'),
    ts('function f(foo: Uint8Array) { foo.fill({}); }'),
    ts('const foo = new Uint8Array(3); foo.fill({});'),
    ts('function f(foo: Set<object>) { foo.fill({}); }'),
    ts(
      'function f(foo: Uint8Array) { (foo as any).fill({}); (foo as unknown).fill({}); }',
    ),
    ts(
      'function f(foo: Set<object>) { (foo as any).fill({}); (foo as unknown).fill({}); }',
    ),
    ts(
      'interface Array { fill(value: object): void } function f(foo: Array) { foo.fill({}); }',
    ),
    ts(
      'interface ReadonlyArray { fill(value: object): void } function f(foo: ReadonlyArray) { foo.fill({}); }',
    ),
  ],
  invalid: [
    invalidJs('new Array(3).fill({})'),
    invalidJs('Array(3).fill([])'),
    invalidJs('Array.from({length: 3}).fill(new Map())'),
    invalidJs('[1, 2, 3].fill(new Set())'),
    invalidJs('const value = {}; array.fill(value)'),
    invalidJs('const value = []; array.fill(value)'),
    invalidJs('const value = new Map(); array.fill(value)'),
    invalidJs('const value = new class {}; array.fill(value)'),
    invalidJs('const RegExp = class {}; array.fill(new RegExp())'),
    invalidJs('array.fill(class {})'),
    invalidJs('const value = {};\narray.fill(value, 1);'),
    invalidTs('array.fill({} as Foo)'),
    invalidTs('array.fill(<Foo>{})'),
    invalidTs('array.fill({} satisfies Foo)'),
    invalidTs('array.fill({}!)'),
    invalidTs('const value = {} as Foo; array.fill(value)'),
    invalidTs('const value = {}; array.fill(value!)'),
    invalidTs('function f(foo: object[]) { foo.fill({}); }'),
    invalidTs('function f(foo: Uint8Array) { (foo as object[]).fill({}); }'),
    invalidTs('function f(foo) { foo.fill({}); }'),
  ],
});
