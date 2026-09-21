// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-async-promise-finally.js
import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

const defaults = {
  filename: 'src/virtual.js',
  languageOptions: { sourceType: 'module' as const },
};

const typeAware = (code: string): ValidTestCase => ({
  ...defaults,
  code,
  filename: 'src/virtual.ts',
});

const message = {
  messageId: 'no-async-promise-finally',
  message: 'Do not pass an async function to `Promise#finally()`.',
};

const invalid = (
  code: string,
  filename = 'src/virtual.js',
): InvalidTestCase => ({
  ...defaults,
  code,
  filename,
  errors: [message],
});

const valid: ValidTestCase[] = [
  { ...defaults, code: 'promise.finally(() => {})' },
  { ...defaults, code: 'promise.finally(() => cleanup())' },
  { ...defaults, code: 'promise.finally(function () {})' },
  { ...defaults, code: 'promise.finally(async function * () {})' },
  { ...defaults, code: 'promise.finally()' },
  { ...defaults, code: 'promise.finally(undefined)' },
  { ...defaults, code: 'promise.finally(cleanup)' },
  { ...defaults, code: 'promise.finally(object.cleanup)' },
  { ...defaults, code: 'promise.then(async () => {})' },
  { ...defaults, code: 'promise.catch(async () => {})' },
  { ...defaults, code: 'finalizer(async () => {})' },
  { ...defaults, code: 'promise.notFinally(async () => {})' },
  { ...defaults, code: 'promise[method](async () => {})' },
  { ...defaults, code: 'const cleanup = () => {}; promise.finally(cleanup);' },
  {
    ...defaults,
    code: 'let cleanup = async () => {}; promise.finally(cleanup);',
  },
  {
    ...defaults,
    code: 'async function * cleanup() {} promise.finally(cleanup);',
  },
  {
    ...defaults,
    code: 'const cleanup = async function * () {}; promise.finally(cleanup);',
  },
  {
    ...defaults,
    code: 'const cleanup = async () => {}; promise.finally(...[cleanup]);',
  },
  {
    ...defaults,
    code: 'import {cleanup} from "./cleanup.js"; promise.finally(cleanup);',
  },
  typeAware(
    'function foo(object: {finally(handler: () => Promise<void>): void}) { object.finally(async () => {}); }',
  ),
  typeAware(
    'function foo(object: {finally(handler: () => Promise<void>): void}) { const cleanup = async () => {}; object.finally(cleanup); }',
  ),
];

const invalidCases: InvalidTestCase[] = [
  invalid('promise.finally(async () => {})'),
  invalid('promise.finally(async () => cleanup())'),
  invalid('promise.finally(async () => { await cleanup(); })'),
  invalid('promise.finally(async function () { await cleanup(); })'),
  invalid('Promise.resolve(value).finally(async () => cleanup())'),
  invalid('new Promise(resolve => resolve()).finally(async () => cleanup())'),
  invalid('promise?.finally(async () => {})'),
  invalid('promise.finally?.(async () => {})'),
  invalid('promise["finally"](async () => {})'),
  invalid('promise[`finally`](async () => {})'),
  invalid('const method = "finally"; promise[method](async () => {});'),
  invalid('async function cleanup() {} promise.finally(cleanup);'),
  invalid('const cleanup = async () => {}; promise.finally(cleanup);'),
  invalid('const cleanup = async function () {}; promise.finally(cleanup);'),
  invalid(
    'type Callback = () => void; promise.finally((async () => {}) as Callback);',
    'src/virtual.ts',
  ),
  invalid(
    'type Callback = () => void; const cleanup = (async () => {}) as Callback; promise.finally(cleanup);',
    'src/virtual.ts',
  ),
  invalid(
    'function foo(promise: Promise<string>) { promise.finally(async () => {}); }',
    'src/virtual.ts',
  ),
];

new RuleTester().run('no-async-promise-finally', {} as never, {
  valid,
  invalid: invalidCases,
});
