// Promise.any cases intentionally have no edits to preserve AggregateError.
// Ported from eslint-plugin-unicorn v77.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/no-single-promise-in-promise-methods.js
import path from 'node:path';
import { lint } from '@rslint/core/internal';
import { RuleTester } from '../rule-tester';
import { buildConfigForSettings } from '../../src/util/load-test-config';

const groups = [
  {
    valid: [],
    invalid: [
      {
        code: 'await Promise.race([(0, promise)])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 34,
            suggestions: [],
          },
        ],
        output: 'await (0, promise)',
      },
      {
        code: 'async function * foo() {await Promise.race([yield promise])}',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 44,
            endLine: 1,
            endColumn: 59,
            suggestions: [],
          },
        ],
        output: 'async function * foo() {await (yield promise)}',
      },
      {
        code: 'async function * foo() {await Promise.race([yield* promise])}',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 44,
            endLine: 1,
            endColumn: 60,
            suggestions: [],
          },
        ],
        output: 'async function * foo() {await (yield* promise)}',
      },
      {
        code: 'await Promise.race([() => promise,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 36,
            suggestions: [],
          },
        ],
        output: 'await (() => promise)',
      },
      {
        code: 'await Promise.race([a ? b : c,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 32,
            suggestions: [],
          },
        ],
        output: 'await (a ? b : c)',
      },
      {
        code: 'await Promise.race([x ??= y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 30,
            suggestions: [],
          },
        ],
        output: 'await (x ??= y)',
      },
      {
        code: 'await Promise.race([x ||= y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 30,
            suggestions: [],
          },
        ],
        output: 'await (x ||= y)',
      },
      {
        code: 'await Promise.race([x &&= y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 30,
            suggestions: [],
          },
        ],
        output: 'await (x &&= y)',
      },
      {
        code: 'await Promise.race([x |= y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 29,
            suggestions: [],
          },
        ],
        output: 'await (x |= y)',
      },
      {
        code: 'await Promise.race([x ^= y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 29,
            suggestions: [],
          },
        ],
        output: 'await (x ^= y)',
      },
      {
        code: 'await Promise.race([x | y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 28,
            suggestions: [],
          },
        ],
        output: 'await (x | y)',
      },
      {
        code: 'await Promise.race([x ^ y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 28,
            suggestions: [],
          },
        ],
        output: 'await (x ^ y)',
      },
      {
        code: 'await Promise.race([x & y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 28,
            suggestions: [],
          },
        ],
        output: 'await (x & y)',
      },
      {
        code: 'await Promise.race([x !== y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 30,
            suggestions: [],
          },
        ],
        output: 'await (x !== y)',
      },
      {
        code: 'await Promise.race([x == y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 29,
            suggestions: [],
          },
        ],
        output: 'await (x == y)',
      },
      {
        code: 'await Promise.race([x in y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 29,
            suggestions: [],
          },
        ],
        output: 'await (x in y)',
      },
      {
        code: 'await Promise.race([x >>> y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 30,
            suggestions: [],
          },
        ],
        output: 'await (x >>> y)',
      },
      {
        code: 'await Promise.race([x + y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 28,
            suggestions: [],
          },
        ],
        output: 'await (x + y)',
      },
      {
        code: 'await Promise.race([x / y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 28,
            suggestions: [],
          },
        ],
        output: 'await (x / y)',
      },
      {
        code: 'await Promise.race([x ** y,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 29,
            suggestions: [],
          },
        ],
        output: 'await (x ** y)',
      },
      {
        code: 'await Promise.race([promise,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 30,
            suggestions: [],
          },
        ],
        output: 'await promise',
      },
      {
        code: 'await Promise.race([getPromise(),],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 35,
            suggestions: [],
          },
        ],
        output: 'await getPromise()',
      },
      {
        code: 'await Promise.race([promises[0],],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 34,
            suggestions: [],
          },
        ],
        output: 'await promises[0]',
      },
      {
        code: 'await Promise.race([await promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 35,
            suggestions: [],
          },
        ],
        output: 'await await promise',
      },
      {
        code: 'await Promise.any([promise])',
        filename: 'src/virtual.js',
        name: 'Promise.any retains its AggregateError rejection semantics; no edits are offered',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.any()` is unnecessary.',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 28,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'await Promise.any([promise,],)',
        filename: 'src/virtual.js',
        name: 'Promise.any retains its AggregateError rejection semantics; no edits are offered',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.any()` is unnecessary.',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 29,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'await Promise.race([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 29,
            suggestions: [],
          },
        ],
        output: 'await promise',
      },
      {
        code: 'await Promise.race([new Promise(() => {})])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 43,
            suggestions: [],
          },
        ],
        output: 'await new Promise(() => {})',
      },
      {
        code: '+await Promise.race([+1])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 21,
            endLine: 1,
            endColumn: 25,
            suggestions: [],
          },
        ],
        output: '+await +1',
      },
      {
        code: 'await Promise.race([(x,y)])\n[0].toString()',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: 'await (x,y)\n[0].toString()',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'await Promise.resolve((x,y))\n[0].toString()',
              },
            ],
          },
        ],
        output: null,
      },
    ],
    name: 'Awaited',
  },
  {
    valid: [
      {
        code: 'Promise.race([promise, anotherPromise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise.race(notArrayLiteral)',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise.race([...promises])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise.any([promise, anotherPromise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise.notListedMethod([promise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise[race]([promise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise.race([,])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'NotPromise.race([promise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise?.race([promise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise.race?.([promise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise.race(...[promise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise.race([promise], extraArguments)',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise.race()',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'new Promise.race([promise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'globalThis.Promise.race([promise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise["race"]([promise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'Promise.allSettled([promise])',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
    ],
    invalid: [
      {
        code: 'Promise.race([promise,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 24,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: 'promise',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve(promise,)',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'foo\nPromise.race([(0, promise),],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 2,
            column: 14,
            endLine: 2,
            endColumn: 29,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: 'foo\n;(0, promise)',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'foo\nPromise.resolve((0, promise),)',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'foo\nPromise.race([[array][0],],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 2,
            column: 14,
            endLine: 2,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: 'foo\n;[array][0]',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'foo\nPromise.resolve([array][0],)',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([promise]).then()',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 23,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: 'promise.then()',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve(promise).then()',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([1]).then()',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 17,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: '(1).then()',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve(1).then()',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([1.]).then()',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 18,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: '(1.).then()',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve(1.).then()',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([.1]).then()',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 18,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: '(.1).then()',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve(.1).then()',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([(0, promise)]).then()',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 28,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: '(0, promise).then()',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve((0, promise)).then()',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'const _ = () => Promise.race([ a ?? b ,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 41,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: 'const _ = () => (a ?? b)',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'const _ = () => Promise.resolve( a ?? b ,)',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([ {a} = 1 ,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 26,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: '({a} = 1)',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve( {a} = 1 ,)',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([ function () {} ,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 33,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: '(function () {})',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve( function () {} ,)',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([ class {} ,],)',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: '(class {})',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve( class {} ,)',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([ new Foo ,],).then()',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 26,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: '(new Foo).then()',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve( new Foo ,).then()',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([ new Foo ,],).toString',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 26,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: '(new Foo).toString',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve( new Foo ,).toString',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'foo(Promise.race([promise]))',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: 'foo(promise)',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'foo(Promise.resolve(promise))',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.any([promise])',
        filename: 'src/virtual.js',
        name: 'Promise.any retains its AggregateError rejection semantics; no edits are offered',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.any()` is unnecessary.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 22,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'foo(Promise.any([promise]))',
        filename: 'src/virtual.js',
        name: 'Promise.any retains its AggregateError rejection semantics; no edits are offered',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.any()` is unnecessary.',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 26,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const obj = {p: Promise.all([promise])}',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 38,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([promise]).foo = 1',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 23,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: 'promise.foo = 1',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve(promise).foo = 1',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([promise])[0] ||= 1',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 23,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: 'promise[0] ||= 1',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve(promise)[0] ||= 1',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([undefined]).then()',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 25,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: 'undefined.then()',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve(undefined).then()',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([null]).then()',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 20,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: '(null).then()',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve(null).then()',
              },
            ],
          },
        ],
        output: null,
      },
    ],
    name: 'NotAwaited',
  },
  {
    valid: [],
    invalid: [
      {
        code: 'Promise.all([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 22,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'await Promise.all([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 28,
            suggestions: [],
          },
        ],
        output: 'await promise',
      },
      {
        code: 'const foo = () => Promise.all([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 31,
            endLine: 1,
            endColumn: 40,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = await Promise.all([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 31,
            endLine: 1,
            endColumn: 40,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'foo = await Promise.all([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 25,
            endLine: 1,
            endColumn: 34,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = await Promise.race([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 41,
            suggestions: [],
          },
        ],
        output: 'const foo = await promise',
      },
      {
        code: 'const foo = () => Promise.race([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 41,
            suggestions: [
              {
                messageId: 'no-single-promise-in-promise-methods/unwrap',
                desc: 'Use the value directly.',
                output: 'const foo = () => promise',
              },
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'const foo = () => Promise.resolve(promise)',
              },
            ],
          },
        ],
        output: null,
      },
      {
        code: 'foo = await Promise.race([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 35,
            suggestions: [],
          },
        ],
        output: 'foo = await promise',
      },
      {
        code: 'const results = await Promise.any([promise])',
        filename: 'src/virtual.js',
        name: 'Promise.any retains its AggregateError rejection semantics; no edits are offered',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.any()` is unnecessary.',
            line: 1,
            column: 35,
            endLine: 1,
            endColumn: 44,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const results = await Promise.race([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 36,
            endLine: 1,
            endColumn: 45,
            suggestions: [],
          },
        ],
        output: 'const results = await promise',
      },
      {
        code: 'const [foo] = await Promise.all([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 33,
            endLine: 1,
            endColumn: 42,
            suggestions: [],
          },
        ],
        output: 'const foo = await promise',
      },
      {
        code: '[foo] = await Promise.all([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 27,
            endLine: 1,
            endColumn: 36,
            suggestions: [],
          },
        ],
        output: 'foo = await promise',
      },
      {
        code: 'const foo = (await Promise.all([promise]))[0]',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 41,
            suggestions: [],
          },
        ],
        output: 'const foo = await promise',
      },
      {
        code: 'const foo = (await Promise.all([promise]))[0.0]',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 41,
            suggestions: [],
          },
        ],
        output: 'const foo = await promise',
      },
      {
        code: 'foo = (await Promise.all([promise]))[0]',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 35,
            suggestions: [],
          },
        ],
        output: 'foo = await promise',
      },
      {
        code: 'const [foo] = await Promise.all([a ? b : c])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 33,
            endLine: 1,
            endColumn: 44,
            suggestions: [],
          },
        ],
        output: 'const foo = await (a ? b : c)',
      },
      {
        code: 'const [foo = bar] = await Promise.all([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 39,
            endLine: 1,
            endColumn: 48,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const [...foo] = await Promise.all([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 36,
            endLine: 1,
            endColumn: 45,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const [, foo] = await Promise.all([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 35,
            endLine: 1,
            endColumn: 44,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const result = ([foo] = await Promise.all([promise]))',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 43,
            endLine: 1,
            endColumn: 52,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = (await Promise.all([promise]))[1]',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 41,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = (await Promise.all([promise])) /* comment */ [0]',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 41,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const [/* comment */ foo] = await Promise.all([promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 47,
            endLine: 1,
            endColumn: 56,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const [foo] = await Promise.all([/* comment */ promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 33,
            endLine: 1,
            endColumn: 56,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const [foo]: [Foo] = await Promise.all([promise])',
        filename: 'src/virtual.ts',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 40,
            endLine: 1,
            endColumn: 49,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'await Promise.race([/* comment */ promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 43,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'await Promise.race([promise /* comment */])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 43,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'await Promise.race([promise, /* comment */])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 44,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'await Promise.race([\n\t// comment\n\tpromise,\n])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 20,
            endLine: 4,
            endColumn: 2,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'await Promise.any([/* comment */ promise])',
        filename: 'src/virtual.js',
        name: 'Promise.any retains its AggregateError rejection semantics; no edits are offered',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.any()` is unnecessary.',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 42,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'Promise.race([/* comment */ promise])',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 37,
            suggestions: [
              {
                messageId:
                  'no-single-promise-in-promise-methods/use-promise-resolve',
                desc: 'Switch to `Promise.resolve(…)`.',
                output: 'Promise.resolve(/* comment */ promise)',
              },
            ],
          },
        ],
        output: null,
      },
    ],
    name: 'PromiseAll',
  },
  {
    name: 'Documentation',
    valid: [
      {
        code: 'const foo = await promise;',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'const promise = Promise.resolve(nonPromise);',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'const foo = await Promise.all(promises);',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'const foo = await Promise.any([promise, anotherPromise]);',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
      {
        code: 'const [{value: foo, reason: error}] = await Promise.allSettled([promise]);',
        filename: 'src/virtual.js',
        errors: [],
        output: null,
      },
    ],
    invalid: [
      {
        code: 'const foo = await Promise.all([promise]);',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 31,
            endLine: 1,
            endColumn: 40,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = await Promise.any([promise]);',
        filename: 'src/virtual.js',
        name: 'Promise.any retains its AggregateError rejection semantics; no edits are offered',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.any()` is unnecessary.',
            line: 1,
            column: 31,
            endLine: 1,
            endColumn: 40,
            suggestions: [],
          },
        ],
        output: null,
      },
      {
        code: 'const foo = await Promise.race([promise]);',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.race()` is unnecessary.',
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 41,
            suggestions: [],
          },
        ],
        output: 'const foo = await promise;',
      },
      {
        code: 'const promise = Promise.all([nonPromise]);',
        filename: 'src/virtual.js',
        errors: [
          {
            messageId: 'no-single-promise-in-promise-methods/error',
            message:
              'Wrapping single-element array with `Promise.all()` is unnecessary.',
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 41,
            suggestions: [],
          },
        ],
        output: null,
      },
    ],
  },
];

const ruleName = 'unicorn/no-single-promise-in-promise-methods';
for (const group of groups) {
  describe(group.name, () => {
    new RuleTester().run(
      'no-single-promise-in-promise-methods',
      null as never,
      group,
    );
  });
}

// The suite wrapper checks messages. Check the complete edit and range contract
// for the same upstream cases through the native IPC API.
test('upstream ranges, autofixes, and suggestions', async () => {
  const { config, configDirectory } = await buildConfigForSettings(
    path.resolve(import.meta.dirname, '../rslint.config.mjs'),
    undefined,
  );
  const apply = (
    code: string,
    fixes: { startPos: number; endPos: number; text: string }[] = [],
  ) =>
    [...fixes]
      .sort((a, b) => b.startPos - a.startPos)
      .reduce(
        (text, fix) =>
          text.slice(0, fix.startPos) + fix.text + text.slice(fix.endPos),
        code,
      );
  for (const group of groups) {
    for (const item of group.invalid) {
      const filename = path.resolve(import.meta.dirname, '..', item.filename);
      const result = await lint({
        workingDirectory: process.cwd(),
        configDirectory,
        config: [...config, { rules: { [ruleName]: 'error' } }],
        fileContents: { [filename]: item.code },
      });
      expect(result.fileCount).toBe(1);
      expect(result.ruleCount).toBe(1);
      expect(
        result.diagnostics.map((d) => ({
          ruleName: d.ruleName,
          messageId: d.messageId,
          message: d.message,
          range: d.range,
          suggestions: (d.suggestions ?? []).map((s) => ({
            messageId: s.messageId,
            desc: s.message,
            output: apply(item.code, s.fixes),
          })),
        })),
      ).toEqual(
        item.errors.map((e) => ({
          ruleName,
          messageId: e.messageId,
          message: e.message,
          range: {
            start: { line: e.line, column: e.column },
            end: { line: e.endLine, column: e.endColumn },
          },
          suggestions: e.suggestions,
        })),
      );
      const fixes = result.diagnostics.flatMap((d) => d.fixes ?? []);
      expect(fixes.length ? apply(item.code, fixes) : null).toEqual(
        item.output,
      );
    }
  }
});
