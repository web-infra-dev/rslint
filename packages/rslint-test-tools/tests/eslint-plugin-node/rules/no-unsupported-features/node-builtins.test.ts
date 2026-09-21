// Complete upstream suite: eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-unsupported-features/node-builtins.js
import { RuleTester } from '../../rule-tester';

new RuleTester({
  languageOptions: {
    sourceType: 'module',
    globals: {
      __dirname: 'readonly',
      __filename: 'readonly',
      AbortController: 'readonly',
      AbortSignal: 'readonly',
      atob: 'readonly',
      Blob: 'readonly',
      BroadcastChannel: 'readonly',
      btoa: 'readonly',
      Buffer: 'readonly',
      ByteLengthQueuingStrategy: 'readonly',
      clearImmediate: 'readonly',
      clearInterval: 'readonly',
      clearTimeout: 'readonly',
      CloseEvent: 'readonly',
      CompressionStream: 'readonly',
      console: 'readonly',
      CountQueuingStrategy: 'readonly',
      crypto: 'readonly',
      Crypto: 'readonly',
      CryptoKey: 'readonly',
      CustomEvent: 'readonly',
      DecompressionStream: 'readonly',
      DOMException: 'readonly',
      Event: 'readonly',
      EventTarget: 'readonly',
      exports: 'writable',
      fetch: 'readonly',
      File: 'readonly',
      FormData: 'readonly',
      global: 'readonly',
      Headers: 'readonly',
      MessageChannel: 'readonly',
      MessageEvent: 'readonly',
      MessagePort: 'readonly',
      module: 'readonly',
      navigator: 'readonly',
      Navigator: 'readonly',
      performance: 'readonly',
      Performance: 'readonly',
      PerformanceEntry: 'readonly',
      PerformanceMark: 'readonly',
      PerformanceMeasure: 'readonly',
      PerformanceObserver: 'readonly',
      PerformanceObserverEntryList: 'readonly',
      PerformanceResourceTiming: 'readonly',
      process: 'readonly',
      queueMicrotask: 'readonly',
      ReadableByteStreamController: 'readonly',
      ReadableStream: 'readonly',
      ReadableStreamBYOBReader: 'readonly',
      ReadableStreamBYOBRequest: 'readonly',
      ReadableStreamDefaultController: 'readonly',
      ReadableStreamDefaultReader: 'readonly',
      Request: 'readonly',
      require: 'readonly',
      Response: 'readonly',
      setImmediate: 'readonly',
      setInterval: 'readonly',
      setTimeout: 'readonly',
      structuredClone: 'readonly',
      SubtleCrypto: 'readonly',
      TextDecoder: 'readonly',
      TextDecoderStream: 'readonly',
      TextEncoder: 'readonly',
      TextEncoderStream: 'readonly',
      TransformStream: 'readonly',
      TransformStreamDefaultController: 'readonly',
      URL: 'readonly',
      URLSearchParams: 'readonly',
      WebAssembly: 'readonly',
      WebSocket: 'readonly',
      WritableStream: 'readonly',
      WritableStreamDefaultController: 'readonly',
      WritableStreamDefaultWriter: 'readonly',
    },
  },
}).run(
  'no-unsupported-features/node-builtins',
  {},
  {
    valid: [
      {
        name: 'assert: upstream valid 1',
        code: "require('assert').strictEqual()",
        options: [
          {
            version: '0.12.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 2',
        code: "var assert = require('assert'); assert(); assert.strictEqual()",
        options: [
          {
            version: '0.12.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 3',
        code: "require('assert').deepStrictEqual()",
        options: [
          {
            version: '4.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 4',
        code: "var assert = require('assert'); assert.deepStrictEqual()",
        options: [
          {
            version: '4.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 5',
        code: "var { deepStrictEqual } = require('assert'); deepStrictEqual()",
        options: [
          {
            version: '4.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 6',
        code: "import assert from 'assert'; assert.deepStrictEqual()",
        options: [
          {
            version: '4.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 7',
        code: "import { deepStrictEqual } from 'assert'; deepStrictEqual()",
        options: [
          {
            version: '4.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 8',
        code: "require('assert').notDeepStrictEqual()",
        options: [
          {
            version: '4.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 9',
        code: "require('assert').rejects()",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 10',
        code: "require('assert').doesNotReject()",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 11',
        code: "require('assert').strict.rejects()",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 12',
        code: "require('assert').strict.doesNotReject()",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 13',
        code: "var assert = require('assert').strict",
        options: [
          {
            version: '9.9.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 14',
        code: "var {strict: assert} = require('assert'); assert.rejects()",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 15',
        code: "require('assert').deepStrictEqual()",
        options: [
          {
            version: '3.9.9',
            ignores: ['assert.deepStrictEqual'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 16',
        code: "var assert = require('assert'); assert.deepStrictEqual()",
        options: [
          {
            version: '3.9.9',
            ignores: ['assert.deepStrictEqual'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 17',
        code: "var { deepStrictEqual } = require('assert'); deepStrictEqual()",
        options: [
          {
            version: '3.9.9',
            ignores: ['assert.deepStrictEqual'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 18',
        code: "import assert from 'assert'; assert.deepStrictEqual()",
        options: [
          {
            version: '3.9.9',
            ignores: ['assert.deepStrictEqual'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 19',
        code: "import { deepStrictEqual } from 'assert'; deepStrictEqual()",
        options: [
          {
            version: '3.9.9',
            ignores: ['assert.deepStrictEqual'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 20',
        code: "require('assert').notDeepStrictEqual()",
        options: [
          {
            version: '3.9.9',
            ignores: ['assert.notDeepStrictEqual'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 21',
        code: "require('assert').rejects()",
        options: [
          {
            version: '9.9.9',
            ignores: ['assert.rejects'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 22',
        code: "require('assert').doesNotReject()",
        options: [
          {
            version: '9.9.9',
            ignores: ['assert.doesNotReject'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 23',
        code: "require('assert').strict.rejects()",
        options: [
          {
            version: '9.9.9',
            ignores: ['assert.strict.rejects'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 24',
        code: "require('assert').strict.doesNotReject()",
        options: [
          {
            version: '9.9.9',
            ignores: ['assert.strict.doesNotReject'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 25',
        code: "var assert = require('assert').strict",
        options: [
          {
            version: '9.8.9',
            ignores: ['assert.strict'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 26',
        code: "var {strict: assert} = require('assert'); assert.rejects()",
        options: [
          {
            version: '9.8.9',
            ignores: ['assert.strict', 'assert.strict.rejects'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 27',
        code: "const { CallTracker } = require('assert'); new CallTracker();",
        options: [
          {
            version: '14.2.0',
            ignores: ['assert.CallTracker'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 28',
        code: "import { CallTracker } from 'assert'; new CallTracker();",
        options: [
          {
            version: '14.2.0',
            ignores: ['assert.CallTracker'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 29',
        code: "const assert = require('node:assert'); assert.deepStrictEqual()",
        options: [
          {
            version: '12.20.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 30',
        code: "import assert from 'node:assert'; assert.deepStrictEqual()",
        options: [
          {
            version: '14.13.1',
          },
        ],
      },
      {
        name: 'assert: upstream valid 31',
        code: "require('node:assert').match()",
        options: [
          {
            version: '15.0.0',
            ignores: ['assert.match'],
          },
        ],
      },
      {
        name: 'assert: upstream valid 32',
        code: 'new Buffer(123)',
        options: [
          {
            version: '6.0.0',
          },
        ],
      },
      {
        name: 'assert: upstream valid 33',
        code: "require('tls').DEFAULT_CIPHERS",
        options: [
          {
            version: '18.0.0',
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 34',
        code: "require('async_hooks')",
        options: [
          {
            version: '16.4.0',
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 35',
        code: "import hooks from 'async_hooks'",
        options: [
          {
            version: '16.4.0',
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 36',
        code: "const { AsyncLocalStorage } = require('async_hooks'); new AsyncLocalStorage();",
        options: [
          {
            version: '16.4.0',
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 37',
        code: "import hooks from 'async_hooks'; new hooks.AsyncLocalStorage();",
        options: [
          {
            version: '16.4.0',
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 38',
        code: "require('async_hooks')",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 39',
        code: "import hooks from 'async_hooks'",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 40',
        code: "require('async_hooks').createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 41',
        code: "const hooks = require('async_hooks'); hooks.createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 42',
        code: "const { createHook } = require('async_hooks'); createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 43',
        code: "import * as hooks from 'async_hooks'; hooks.createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 44',
        code: "import hooks from 'async_hooks'; hooks.createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 45',
        code: "import { createHook } from 'async_hooks'; createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 46',
        code: "new require('async_hooks').AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 47',
        code: "const hooks = require('async_hooks'); new hooks.AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 48',
        code: "const { AsyncLocalStorage } = require('async_hooks'); new AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 49',
        code: "import * as hooks from 'async_hooks'; new hooks.AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 50',
        code: "import hooks from 'async_hooks'; new hooks.AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 51',
        code: "import { AsyncLocalStorage } from 'async_hooks'; new AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 52',
        code: "require('node:async_hooks').createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 53',
        code: "const hooks = require('node:async_hooks'); hooks.createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 54',
        code: "const { createHook } = require('node:async_hooks'); createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 55',
        code: "import * as hooks from 'node:async_hooks'; hooks.createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 56',
        code: "import hooks from 'node:async_hooks'; hooks.createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 57',
        code: "import { createHook } from 'node:async_hooks'; createHook()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 58',
        code: "new require('node:async_hooks').AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 59',
        code: "const hooks = require('node:async_hooks'); new hooks.AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 60',
        code: "const { AsyncLocalStorage } = require('node:async_hooks'); new AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 61',
        code: "import * as hooks from 'node:async_hooks'; new hooks.AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 62',
        code: "import hooks from 'node:async_hooks'; new hooks.AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 63',
        code: "import { AsyncLocalStorage } from 'node:async_hooks'; new AsyncLocalStorage()",
        options: [
          {
            version: '14.0.0',
            ignores: [
              'async_hooks',
              'async_hooks.createHook',
              'async_hooks.AsyncLocalStorage',
            ],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 64',
        code: "require('node:async_hooks')",
        options: [
          {
            version: '16.4.0',
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 65',
        code: "import hooks from 'node:async_hooks'",
        options: [
          {
            version: '16.4.0',
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 66',
        code: "require('node:async_hooks')",
        options: [
          {
            version: '12.0.0',
            ignores: ['async_hooks'],
          },
        ],
      },
      {
        name: 'async_hooks: upstream valid 67',
        code: "import hooks from 'node:async_hooks'",
        options: [
          {
            version: '12.0.0',
            ignores: ['async_hooks'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 68',
        code: 'Buffer.alloc',
        options: [
          {
            version: '4.5.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 69',
        code: 'Buffer.allocUnsafe',
        options: [
          {
            version: '4.5.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 70',
        code: 'Buffer.allocUnsafeSlow',
        options: [
          {
            version: '4.5.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 71',
        code: 'Buffer.from',
        options: [
          {
            version: '4.5.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 72',
        code: "require('buffer').constants",
        options: [
          {
            version: '8.2.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 73',
        code: "var cp = require('buffer'); cp.constants",
        options: [
          {
            version: '8.2.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 74',
        code: "var { constants } = require('buffer');",
        options: [
          {
            version: '8.2.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 75',
        code: "import cp from 'buffer'; cp.constants",
        options: [
          {
            version: '8.2.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 76',
        code: "import { constants } from 'buffer'",
        options: [
          {
            version: '8.2.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 77',
        code: "var {Buffer: b} = require('buffer'); b.alloc",
        options: [
          {
            version: '4.5.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 78',
        code: "var {Buffer: b} = require('buffer'); b.allocUnsafe",
        options: [
          {
            version: '4.5.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 79',
        code: "var {Buffer: b} = require('buffer'); b.allocUnsafeSlow",
        options: [
          {
            version: '4.5.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 80',
        code: "var {Buffer: b} = require('buffer'); b.from",
        options: [
          {
            version: '4.5.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 81',
        code: "require('buffer').kMaxLength",
        options: [
          {
            version: '3.0.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 82',
        code: "require('buffer').transcode",
        options: [
          {
            version: '7.1.0',
          },
        ],
      },
      {
        name: 'buffer: upstream valid 83',
        code: 'Buffer.alloc',
        options: [
          {
            version: '4.4.9',
            ignores: ['Buffer.alloc'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 84',
        code: 'Buffer.allocUnsafe',
        options: [
          {
            version: '4.4.9',
            ignores: ['Buffer.allocUnsafe'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 85',
        code: 'Buffer.allocUnsafeSlow',
        options: [
          {
            version: '4.4.9',
            ignores: ['Buffer.allocUnsafeSlow'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 86',
        code: 'Buffer.from',
        options: [
          {
            version: '4.4.9',
            ignores: ['Buffer.from'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 87',
        code: "require('buffer').constants",
        options: [
          {
            version: '8.1.9',
            ignores: ['buffer.constants'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 88',
        code: "var cp = require('buffer'); cp.constants",
        options: [
          {
            version: '8.1.9',
            ignores: ['buffer.constants'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 89',
        code: "var { constants } = require('buffer');",
        options: [
          {
            version: '8.1.9',
            ignores: ['buffer.constants'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 90',
        code: "import cp from 'buffer'; cp.constants",
        options: [
          {
            version: '8.1.9',
            ignores: ['buffer.constants'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 91',
        code: "import { constants } from 'buffer'",
        options: [
          {
            version: '8.1.9',
            ignores: ['buffer.constants'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 92',
        code: "var {Buffer: b} = require('buffer'); b.alloc",
        options: [
          {
            version: '4.4.9',
            ignores: ['buffer.Buffer.alloc'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 93',
        code: "var {Buffer: b} = require('buffer'); b.allocUnsafe",
        options: [
          {
            version: '4.4.9',
            ignores: ['buffer.Buffer.allocUnsafe'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 94',
        code: "var {Buffer: b} = require('buffer'); b.allocUnsafeSlow",
        options: [
          {
            version: '4.4.9',
            ignores: ['buffer.Buffer.allocUnsafeSlow'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 95',
        code: "var {Buffer: b} = require('buffer'); b.from",
        options: [
          {
            version: '4.4.9',
            ignores: ['buffer.Buffer.from'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 96',
        code: "require('buffer').kMaxLength",
        options: [
          {
            version: '2.9.9',
            ignores: ['buffer.kMaxLength'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 97',
        code: "require('buffer').transcode",
        options: [
          {
            version: '7.0.9',
            ignores: ['buffer.transcode'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 98',
        code: "const { Blob } = require('buffer'); new Blob();",
        options: [
          {
            version: '15.7.0',
            ignores: ['buffer.Blob'],
          },
        ],
      },
      {
        name: 'buffer: upstream valid 99',
        code: "import buffer from 'buffer'; new buffer.Blob();",
        options: [
          {
            version: '15.7.0',
            ignores: ['buffer.Blob'],
          },
        ],
      },
      {
        name: 'child_process: upstream valid 100',
        code: "require('child_process').ChildProcess",
        options: [
          {
            version: '2.2.0',
          },
        ],
      },
      {
        name: 'child_process: upstream valid 101',
        code: "var cp = require('child_process'); cp.ChildProcess",
        options: [
          {
            version: '2.2.0',
          },
        ],
      },
      {
        name: 'child_process: upstream valid 102',
        code: "var { ChildProcess } = require('child_process'); ChildProcess",
        options: [
          {
            version: '2.2.0',
          },
        ],
      },
      {
        name: 'child_process: upstream valid 103',
        code: "import cp from 'child_process'; cp.ChildProcess",
        options: [
          {
            version: '2.2.0',
          },
        ],
      },
      {
        name: 'child_process: upstream valid 104',
        code: "import { ChildProcess } from 'child_process'",
        options: [
          {
            version: '2.2.0',
          },
        ],
      },
      {
        name: 'child_process: upstream valid 105',
        code: "require('child_process').ChildProcess",
        options: [
          {
            version: '2.1.9',
            ignores: ['child_process.ChildProcess'],
          },
        ],
      },
      {
        name: 'child_process: upstream valid 106',
        code: "var cp = require('child_process'); cp.ChildProcess",
        options: [
          {
            version: '2.1.9',
            ignores: ['child_process.ChildProcess'],
          },
        ],
      },
      {
        name: 'child_process: upstream valid 107',
        code: "var { ChildProcess } = require('child_process'); ChildProcess",
        options: [
          {
            version: '2.1.9',
            ignores: ['child_process.ChildProcess'],
          },
        ],
      },
      {
        name: 'child_process: upstream valid 108',
        code: "import cp from 'child_process'; cp.ChildProcess",
        options: [
          {
            version: '2.1.9',
            ignores: ['child_process.ChildProcess'],
          },
        ],
      },
      {
        name: 'child_process: upstream valid 109',
        code: "import { ChildProcess } from 'child_process'",
        options: [
          {
            version: '2.1.9',
            ignores: ['child_process.ChildProcess'],
          },
        ],
      },
      {
        name: 'console: upstream valid 110',
        code: 'console.clear()',
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 111',
        code: "require('console').clear()",
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 112',
        code: "var c = require('console'); c.clear()",
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 113',
        code: "var { clear } = require('console'); clear()",
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 114',
        code: "import c from 'console'; c.clear()",
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 115',
        code: 'console.count()',
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 116',
        code: 'console.countReset()',
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 117',
        code: 'console.debug()',
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 118',
        code: 'console.dirxml()',
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 119',
        code: 'console.group()',
        options: [
          {
            version: '8.5.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 120',
        code: 'console.groupCollapsed()',
        options: [
          {
            version: '8.5.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 121',
        code: 'console.groupEnd()',
        options: [
          {
            version: '8.5.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 122',
        code: 'console.table()',
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 123',
        code: 'console.markTimeline()',
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 124',
        code: 'console.profile()',
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 125',
        code: 'console.profileEnd()',
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 126',
        code: 'console.timeStamp()',
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 127',
        code: 'console.timeline()',
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 128',
        code: 'console.timelineEnd()',
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'console: upstream valid 129',
        code: 'console.clear()',
        options: [
          {
            version: '8.2.9',
            ignores: ['console.clear'],
          },
        ],
      },
      {
        name: 'console: upstream valid 130',
        code: "require('console').clear()",
        options: [
          {
            version: '8.2.9',
            ignores: ['console.clear'],
          },
        ],
      },
      {
        name: 'console: upstream valid 131',
        code: "var c = require('console'); c.clear()",
        options: [
          {
            version: '8.2.9',
            ignores: ['console.clear'],
          },
        ],
      },
      {
        name: 'console: upstream valid 132',
        code: "var { clear } = require('console'); clear()",
        options: [
          {
            version: '8.2.9',
            ignores: ['console.clear'],
          },
        ],
      },
      {
        name: 'console: upstream valid 133',
        code: "import c from 'console'; c.clear()",
        options: [
          {
            version: '8.2.9',
            ignores: ['console.clear'],
          },
        ],
      },
      {
        name: 'console: upstream valid 134',
        code: 'console.count()',
        options: [
          {
            version: '8.2.9',
            ignores: ['console.count'],
          },
        ],
      },
      {
        name: 'console: upstream valid 135',
        code: 'console.countReset()',
        options: [
          {
            version: '8.2.9',
            ignores: ['console.countReset'],
          },
        ],
      },
      {
        name: 'console: upstream valid 136',
        code: 'console.debug()',
        options: [
          {
            version: '7.9.9',
            ignores: ['console.debug'],
          },
        ],
      },
      {
        name: 'console: upstream valid 137',
        code: 'console.dirxml()',
        options: [
          {
            version: '7.9.9',
            ignores: ['console.dirxml'],
          },
        ],
      },
      {
        name: 'console: upstream valid 138',
        code: 'console.group()',
        options: [
          {
            version: '8.4.9',
            ignores: ['console.group'],
          },
        ],
      },
      {
        name: 'console: upstream valid 139',
        code: 'console.groupCollapsed()',
        options: [
          {
            version: '8.4.9',
            ignores: ['console.groupCollapsed'],
          },
        ],
      },
      {
        name: 'console: upstream valid 140',
        code: 'console.groupEnd()',
        options: [
          {
            version: '8.4.9',
            ignores: ['console.groupEnd'],
          },
        ],
      },
      {
        name: 'console: upstream valid 141',
        code: 'console.table()',
        options: [
          {
            version: '9.9.9',
            ignores: ['console.table'],
          },
        ],
      },
      {
        name: 'console: upstream valid 142',
        code: 'console.profile()',
        options: [
          {
            version: '7.9.9',
            ignores: ['console.profile'],
          },
        ],
      },
      {
        name: 'console: upstream valid 143',
        code: 'console.profileEnd()',
        options: [
          {
            version: '7.9.9',
            ignores: ['console.profileEnd'],
          },
        ],
      },
      {
        name: 'console: upstream valid 144',
        code: 'console.timeStamp()',
        options: [
          {
            version: '7.9.9',
            ignores: ['console.timeStamp'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 145',
        code: "require('crypto').constants",
        options: [
          {
            version: '6.3.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 146',
        code: "var hooks = require('crypto'); hooks.constants",
        options: [
          {
            version: '6.3.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 147',
        code: "var { constants } = require('crypto'); constants",
        options: [
          {
            version: '6.3.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 148',
        code: "import crypto from 'crypto'; crypto.constants",
        options: [
          {
            version: '6.3.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 149',
        code: "import { constants } from 'crypto'; constants",
        options: [
          {
            version: '6.3.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 150',
        code: "require('crypto').Certificate.exportChallenge()",
        options: [
          {
            version: '9.0.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 151',
        code: "var { Certificate: c } = require('crypto'); c.exportChallenge()",
        options: [
          {
            version: '9.0.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 152',
        code: "var { Certificate: c } = require('crypto'); c.exportPublicKey()",
        options: [
          {
            version: '9.0.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 153',
        code: "var { Certificate: c } = require('crypto'); c.verifySpkac()",
        options: [
          {
            version: '9.0.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 154',
        code: "require('crypto').fips",
        options: [
          {
            version: '6.0.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 155',
        code: "require('crypto').getCurves",
        options: [
          {
            version: '2.3.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 156',
        code: "require('crypto').getFips",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 157',
        code: "require('crypto').privateEncrypt",
        options: [
          {
            version: '1.1.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 158',
        code: "require('crypto').publicDecrypt",
        options: [
          {
            version: '1.1.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 159',
        code: "require('crypto').randomFillSync",
        options: [
          {
            version: '7.10.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 160',
        code: "require('crypto').randomFill",
        options: [
          {
            version: '7.10.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 161',
        code: "require('crypto').scrypt",
        options: [
          {
            version: '10.5.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 162',
        code: "require('crypto').scryptSync",
        options: [
          {
            version: '10.5.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 163',
        code: "require('crypto').setFips",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 164',
        code: "require('crypto').timingSafeEqual",
        options: [
          {
            version: '6.6.0',
          },
        ],
      },
      {
        name: 'crypto: upstream valid 165',
        code: "require('crypto').constants",
        options: [
          {
            version: '6.2.9',
            ignores: ['crypto.constants'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 166',
        code: "var hooks = require('crypto'); hooks.constants",
        options: [
          {
            version: '6.2.9',
            ignores: ['crypto.constants'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 167',
        code: "var { constants } = require('crypto'); constants",
        options: [
          {
            version: '6.2.9',
            ignores: ['crypto.constants'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 168',
        code: "import crypto from 'crypto'; crypto.constants",
        options: [
          {
            version: '6.2.9',
            ignores: ['crypto.constants'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 169',
        code: "import { constants } from 'crypto'; constants",
        options: [
          {
            version: '6.2.9',
            ignores: ['crypto.constants'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 170',
        code: "require('crypto').Certificate.exportChallenge()",
        options: [
          {
            version: '8.9.9',
            ignores: ['crypto.Certificate.exportChallenge'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 171',
        code: "var { Certificate: c } = require('crypto'); c.exportChallenge()",
        options: [
          {
            version: '8.9.9',
            ignores: ['crypto.Certificate.exportChallenge'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 172',
        code: "var { Certificate: c } = require('crypto'); c.exportPublicKey()",
        options: [
          {
            version: '8.9.9',
            ignores: ['crypto.Certificate.exportPublicKey'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 173',
        code: "var { Certificate: c } = require('crypto'); c.verifySpkac()",
        options: [
          {
            version: '8.9.9',
            ignores: ['crypto.Certificate.verifySpkac'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 174',
        code: "require('crypto').fips",
        options: [
          {
            version: '5.9.9',
            ignores: ['crypto.fips'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 175',
        code: "require('crypto').getCurves",
        options: [
          {
            version: '2.2.9',
            ignores: ['crypto.getCurves'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 176',
        code: "require('crypto').getFips",
        options: [
          {
            version: '9.9.9',
            ignores: ['crypto.getFips'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 177',
        code: "require('crypto').privateEncrypt",
        options: [
          {
            version: '1.0.9',
            ignores: ['crypto.privateEncrypt'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 178',
        code: "require('crypto').publicDecrypt",
        options: [
          {
            version: '1.0.9',
            ignores: ['crypto.publicDecrypt'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 179',
        code: "require('crypto').randomFillSync",
        options: [
          {
            version: '7.9.9',
            ignores: ['crypto.randomFillSync'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 180',
        code: "require('crypto').randomFill",
        options: [
          {
            version: '7.9.9',
            ignores: ['crypto.randomFill'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 181',
        code: "require('crypto').scrypt",
        options: [
          {
            version: '10.4.9',
            ignores: ['crypto.scrypt'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 182',
        code: "require('crypto').scryptSync",
        options: [
          {
            version: '10.4.9',
            ignores: ['crypto.scryptSync'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 183',
        code: "require('crypto').setFips",
        options: [
          {
            version: '9.9.9',
            ignores: ['crypto.setFips'],
          },
        ],
      },
      {
        name: 'crypto: upstream valid 184',
        code: "require('crypto').timingSafeEqual",
        options: [
          {
            version: '6.5.9',
            ignores: ['crypto.timingSafeEqual'],
          },
        ],
      },
      {
        name: 'dns: upstream valid 185',
        code: "require('dns').Resolver",
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'dns: upstream valid 186',
        code: "var hooks = require('dns'); hooks.Resolver",
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'dns: upstream valid 187',
        code: "var { Resolver } = require('dns'); Resolver",
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'dns: upstream valid 188',
        code: "import dns from 'dns'; dns.Resolver",
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'dns: upstream valid 189',
        code: "import { Resolver } from 'dns'; Resolver",
        options: [
          {
            version: '8.3.0',
          },
        ],
      },
      {
        name: 'dns: upstream valid 190',
        code: "require('dns').resolvePtr",
        options: [
          {
            version: '6.0.0',
          },
        ],
      },
      {
        name: 'dns: upstream valid 191',
        code: "require('dns').promises",
        options: [
          {
            version: '11.14.0',
          },
        ],
      },
      {
        name: 'dns: upstream valid 192',
        code: "require('dns').Resolver",
        options: [
          {
            version: '8.2.9',
            ignores: ['dns.Resolver'],
          },
        ],
      },
      {
        name: 'dns: upstream valid 193',
        code: "var hooks = require('dns'); hooks.Resolver",
        options: [
          {
            version: '8.2.9',
            ignores: ['dns.Resolver'],
          },
        ],
      },
      {
        name: 'dns: upstream valid 194',
        code: "var { Resolver } = require('dns'); Resolver",
        options: [
          {
            version: '8.2.9',
            ignores: ['dns.Resolver'],
          },
        ],
      },
      {
        name: 'dns: upstream valid 195',
        code: "import dns from 'dns'; dns.Resolver",
        options: [
          {
            version: '8.2.9',
            ignores: ['dns.Resolver'],
          },
        ],
      },
      {
        name: 'dns: upstream valid 196',
        code: "import { Resolver } from 'dns'; Resolver",
        options: [
          {
            version: '8.2.9',
            ignores: ['dns.Resolver'],
          },
        ],
      },
      {
        name: 'dns: upstream valid 197',
        code: "require('dns').resolvePtr",
        options: [
          {
            version: '5.9.9',
            ignores: ['dns.resolvePtr'],
          },
        ],
      },
      {
        name: 'dns: upstream valid 198',
        code: "require('dns').promises",
        options: [
          {
            version: '11.13.9',
            ignores: ['dns.promises'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 199',
        code: "require('fs').promises",
        options: [
          {
            version: '11.14.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 200',
        code: "var fs = require('fs'); fs.promises",
        options: [
          {
            version: '11.14.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 201',
        code: "var { promises } = require('fs'); promises",
        options: [
          {
            version: '11.14.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 202',
        code: "import fs from 'fs'; fs.promises",
        options: [
          {
            version: '11.14.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 203',
        code: "import { promises } from 'fs'",
        options: [
          {
            version: '11.14.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 204',
        code: "require('fs').copyFile",
        options: [
          {
            version: '8.5.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 205',
        code: "require('fs').copyFileSync",
        options: [
          {
            version: '8.5.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 206',
        code: "require('fs').mkdtemp",
        options: [
          {
            version: '5.10.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 207',
        code: "require('fs').mkdtempSync",
        options: [
          {
            version: '5.10.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 208',
        code: "require('fs').realpath.native",
        options: [
          {
            version: '9.2.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 209',
        code: "require('fs').realpathSync.native",
        options: [
          {
            version: '9.2.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 210',
        code: "require('fs').promises",
        options: [
          {
            version: '11.13.9',
            ignores: ['fs.promises'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 211',
        code: "var fs = require('fs'); fs.promises",
        options: [
          {
            version: '11.13.9',
            ignores: ['fs.promises'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 212',
        code: "var { promises } = require('fs'); promises",
        options: [
          {
            version: '11.13.9',
            ignores: ['fs.promises'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 213',
        code: "import fs from 'fs'; fs.promises",
        options: [
          {
            version: '11.13.9',
            ignores: ['fs.promises'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 214',
        code: "import { promises } from 'fs'",
        options: [
          {
            version: '11.13.9',
            ignores: ['fs.promises'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 215',
        code: "require('fs').copyFile",
        options: [
          {
            version: '8.4.9',
            ignores: ['fs.copyFile'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 216',
        code: "require('fs').copyFileSync",
        options: [
          {
            version: '8.4.9',
            ignores: ['fs.copyFileSync'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 217',
        code: "require('fs').mkdtemp",
        options: [
          {
            version: '5.9.9',
            ignores: ['fs.mkdtemp'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 218',
        code: "require('fs').mkdtempSync",
        options: [
          {
            version: '5.9.9',
            ignores: ['fs.mkdtempSync'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 219',
        code: "require('fs').realpath.native",
        options: [
          {
            version: '9.1.9',
            ignores: ['fs.realpath.native'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 220',
        code: "require('fs').realpathSync.native",
        options: [
          {
            version: '9.1.9',
            ignores: ['fs.realpathSync.native'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 221',
        code: "require('fs').readv",
        options: [
          {
            version: '13.13.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 222',
        code: "require('fs').readvSync",
        options: [
          {
            version: '13.13.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 223',
        code: "require('fs').readv",
        options: [
          {
            version: '12.17.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 224',
        code: "require('fs').readvSync",
        options: [
          {
            version: '12.17.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 225',
        code: "require('fs').readv",
        options: [
          {
            version: '13.12.0',
            ignores: ['fs.readv'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 226',
        code: "require('fs').readvSync",
        options: [
          {
            version: '13.12.0',
            ignores: ['fs.readvSync'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 227',
        code: "require('fs').lutimes",
        options: [
          {
            version: '14.5.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 228',
        code: "require('fs').lutimesSync",
        options: [
          {
            version: '14.5.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 229',
        code: "require('fs').lutimes",
        options: [
          {
            version: '12.19.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 230',
        code: "require('fs').lutimesSync",
        options: [
          {
            version: '12.19.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 231',
        code: "require('fs').lutimes",
        options: [
          {
            version: '13.14.0',
            ignores: ['fs.lutimes'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 232',
        code: "require('fs').lutimesSync",
        options: [
          {
            version: '13.14.0',
            ignores: ['fs.lutimesSync'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 233',
        code: "require('fs').opendir",
        options: [
          {
            version: '12.12.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 234',
        code: "require('fs').opendirSync",
        options: [
          {
            version: '12.12.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 235',
        code: "require('fs').opendir",
        options: [
          {
            version: '12.11.0',
            ignores: ['fs.opendir'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 236',
        code: "require('fs').opendirSync",
        options: [
          {
            version: '12.11.0',
            ignores: ['fs.opendirSync'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 237',
        code: "require('fs').rm",
        options: [
          {
            version: '14.14.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 238',
        code: "require('fs').rmSync",
        options: [
          {
            version: '14.14.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 239',
        code: "require('fs').rm",
        options: [
          {
            version: '14.13.0',
            ignores: ['fs.rm'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 240',
        code: "require('fs').rmSync",
        options: [
          {
            version: '14.13.0',
            ignores: ['fs.rmSync'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 241',
        code: "require('fs').read",
        options: [
          {
            version: '13.11.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 242',
        code: "require('fs').readSync",
        options: [
          {
            version: '13.11.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 243',
        code: "require('fs').read",
        options: [
          {
            version: '12.17.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 244',
        code: "require('fs').readSync",
        options: [
          {
            version: '12.17.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 245',
        code: "require('fs').read",
        options: [
          {
            version: '13.10.0',
            ignores: ['fs.read'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 246',
        code: "require('fs').readSync",
        options: [
          {
            version: '13.10.0',
            ignores: ['fs.readSync'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 247',
        code: "require('fs').Dir",
        options: [
          {
            version: '12.12.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 248',
        code: "require('fs').Dir",
        options: [
          {
            version: '12.11.0',
            ignores: ['fs.Dir'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 249',
        code: "require('fs').StatWatcher",
        options: [
          {
            version: '14.3.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 250',
        code: "require('fs').StatWatcher",
        options: [
          {
            version: '14.2.0',
            ignores: ['fs.StatWatcher'],
          },
        ],
      },
      {
        name: 'fs: upstream valid 251',
        code: "require('fs').StatWatcher",
        options: [
          {
            version: '12.20.0',
          },
        ],
      },
      {
        name: 'fs: upstream valid 252',
        code: "require('fs').StatWatcher",
        options: [
          {
            version: '12.19.0',
            ignores: ['fs.StatWatcher'],
          },
        ],
      },
      {
        name: 'fs/promises: upstream valid 253',
        code: "import * as fs from 'fs/promises';",
        options: [
          {
            version: '14.0.0',
          },
        ],
      },
      {
        name: 'fs/promises: upstream valid 254',
        code: "require('fs/promise')",
        options: [
          {
            version: '14.0.0',
          },
        ],
      },
      {
        name: 'fs/promises: upstream valid 255',
        code: "import * as fs from 'node:fs/promises';",
        options: [
          {
            version: '14.13.1',
          },
        ],
      },
      {
        name: 'fs/promises: upstream valid 256',
        code: "require('node:fs/promise')",
        options: [
          {
            version: '14.13.1',
          },
        ],
      },
      {
        name: 'fs/promises: upstream valid 257',
        code: "import * as fs from 'fs/promises';",
        options: [
          {
            version: '13.14.0',
            ignores: ['fs/promises'],
          },
        ],
      },
      {
        name: 'fs/promises: upstream valid 258',
        code: "import * as fs from 'node:fs/promises';",
        options: [
          {
            version: '13.14.0',
            ignores: ['fs/promises'],
          },
        ],
      },
      {
        name: 'fs/promises: upstream valid 259',
        code: "require('fs/promise')",
        options: [
          {
            version: '13.14.0',
            ignores: ['fs/promises'],
          },
        ],
      },
      {
        name: 'fs/promises: upstream valid 260',
        code: "require('node:fs/promise')",
        options: [
          {
            version: '13.14.0',
            ignores: ['fs/promises'],
          },
        ],
      },
      {
        name: 'http2: upstream valid 261',
        code: "require('http2')",
        options: [
          {
            version: '10.10.0',
          },
        ],
      },
      {
        name: 'http2: upstream valid 262',
        code: "import http2 from 'http2'",
        options: [
          {
            version: '10.10.0',
          },
        ],
      },
      {
        name: 'http2: upstream valid 263',
        code: "require('http2')",
        options: [
          {
            version: '8.3.9',
            ignores: ['http2'],
          },
        ],
      },
      {
        name: 'http2: upstream valid 264',
        code: "import http2 from 'http2'",
        options: [
          {
            version: '8.3.9',
            ignores: ['http2'],
          },
        ],
      },
      {
        name: 'inspector: upstream valid 265',
        code: "require('inspector')",
        options: [
          {
            version: '7.9.9',
            ignores: ['inspector'],
          },
        ],
      },
      {
        name: 'inspector: upstream valid 266',
        code: "import inspector from 'inspector'",
        options: [
          {
            version: '7.9.9',
            ignores: ['inspector'],
          },
        ],
      },
      {
        name: 'inspector: upstream valid 267',
        code: "import { open } from 'inspector'",
        options: [
          {
            version: '7.9.9',
            ignores: ['inspector', 'inspector.open'],
          },
        ],
      },
      {
        name: 'module: upstream valid 268',
        code: 'require.resolve.paths()',
        options: [
          {
            version: '8.9.0',
          },
        ],
      },
      {
        name: 'module: upstream valid 269',
        code: "require('module').builtinModules",
        options: [
          {
            version: '9.3.0',
          },
        ],
      },
      {
        name: 'module: upstream valid 270',
        code: 'require.resolve.paths()',
        options: [
          {
            version: '8.8.9',
            ignores: ['require.resolve.paths'],
          },
        ],
      },
      {
        name: 'module: upstream valid 271',
        code: "require('module').builtinModules",
        options: [
          {
            version: '9.2.9',
            ignores: ['module.builtinModules'],
          },
        ],
      },
      {
        name: 'os: upstream valid 272',
        code: "require('os').constants",
        options: [
          {
            version: '6.3.0',
          },
        ],
      },
      {
        name: 'os: upstream valid 273',
        code: "var hooks = require('os'); hooks.constants",
        options: [
          {
            version: '6.3.0',
          },
        ],
      },
      {
        name: 'os: upstream valid 274',
        code: "var { constants } = require('os'); constants",
        options: [
          {
            version: '6.3.0',
          },
        ],
      },
      {
        name: 'os: upstream valid 275',
        code: "import os from 'os'; os.constants",
        options: [
          {
            version: '6.3.0',
          },
        ],
      },
      {
        name: 'os: upstream valid 276',
        code: "import { constants } from 'os'; constants",
        options: [
          {
            version: '6.3.0',
          },
        ],
      },
      {
        name: 'os: upstream valid 277',
        code: "require('os').homedir",
        options: [
          {
            version: '2.3.0',
          },
        ],
      },
      {
        name: 'os: upstream valid 278',
        code: "require('os').userInfo",
        options: [
          {
            version: '6.0.0',
          },
        ],
      },
      {
        name: 'os: upstream valid 279',
        code: "require('os').constants",
        options: [
          {
            version: '6.2.9',
            ignores: ['os.constants'],
          },
        ],
      },
      {
        name: 'os: upstream valid 280',
        code: "var hooks = require('os'); hooks.constants",
        options: [
          {
            version: '6.2.9',
            ignores: ['os.constants'],
          },
        ],
      },
      {
        name: 'os: upstream valid 281',
        code: "var { constants } = require('os'); constants",
        options: [
          {
            version: '6.2.9',
            ignores: ['os.constants'],
          },
        ],
      },
      {
        name: 'os: upstream valid 282',
        code: "import os from 'os'; os.constants",
        options: [
          {
            version: '6.2.9',
            ignores: ['os.constants'],
          },
        ],
      },
      {
        name: 'os: upstream valid 283',
        code: "import { constants } from 'os'; constants",
        options: [
          {
            version: '6.2.9',
            ignores: ['os.constants'],
          },
        ],
      },
      {
        name: 'os: upstream valid 284',
        code: "require('os').homedir",
        options: [
          {
            version: '2.2.9',
            ignores: ['os.homedir'],
          },
        ],
      },
      {
        name: 'os: upstream valid 285',
        code: "require('os').userInfo",
        options: [
          {
            version: '5.9.9',
            ignores: ['os.userInfo'],
          },
        ],
      },
      {
        name: 'path: upstream valid 286',
        code: "require('path').toNamespacedPath()",
        options: [
          {
            version: '9.0.0',
          },
        ],
      },
      {
        name: 'path: upstream valid 287',
        code: "var path = require('path'); path.toNamespacedPath()",
        options: [
          {
            version: '9.0.0',
          },
        ],
      },
      {
        name: 'path: upstream valid 288',
        code: "var { toNamespacedPath } = require('path'); toNamespacedPath()",
        options: [
          {
            version: '9.0.0',
          },
        ],
      },
      {
        name: 'path: upstream valid 289',
        code: "import path from 'path'; path.toNamespacedPath()",
        options: [
          {
            version: '9.0.0',
          },
        ],
      },
      {
        name: 'path: upstream valid 290',
        code: "import { toNamespacedPath } from 'path'; toNamespacedPath()",
        options: [
          {
            version: '9.0.0',
          },
        ],
      },
      {
        name: 'path: upstream valid 291',
        code: "require('path').toNamespacedPath()",
        options: [
          {
            version: '8.9.9',
            ignores: ['path.toNamespacedPath'],
          },
        ],
      },
      {
        name: 'path: upstream valid 292',
        code: "var path = require('path'); path.toNamespacedPath()",
        options: [
          {
            version: '8.9.9',
            ignores: ['path.toNamespacedPath'],
          },
        ],
      },
      {
        name: 'path: upstream valid 293',
        code: "var { toNamespacedPath } = require('path'); toNamespacedPath()",
        options: [
          {
            version: '8.9.9',
            ignores: ['path.toNamespacedPath'],
          },
        ],
      },
      {
        name: 'path: upstream valid 294',
        code: "import path from 'path'; path.toNamespacedPath()",
        options: [
          {
            version: '8.9.9',
            ignores: ['path.toNamespacedPath'],
          },
        ],
      },
      {
        name: 'path: upstream valid 295',
        code: "import { toNamespacedPath } from 'path'; toNamespacedPath()",
        options: [
          {
            version: '8.9.9',
            ignores: ['path.toNamespacedPath'],
          },
        ],
      },
      {
        name: 'perf_hooks: upstream valid 296',
        code: "require('perf_hooks')",
        options: [
          {
            version: '8.5.0',
          },
        ],
      },
      {
        name: 'perf_hooks: upstream valid 297',
        code: "import perf_hooks from 'perf_hooks'",
        options: [
          {
            version: '8.5.0',
          },
        ],
      },
      {
        name: 'perf_hooks: upstream valid 298',
        code: "require('perf_hooks')",
        options: [
          {
            version: '8.4.9',
            ignores: ['perf_hooks'],
          },
        ],
      },
      {
        name: 'perf_hooks: upstream valid 299',
        code: "import perf_hooks from 'perf_hooks'",
        options: [
          {
            version: '8.4.9',
            ignores: ['perf_hooks'],
          },
        ],
      },
      {
        name: 'process: upstream valid 300',
        code: 'process.argv0',
        options: [
          {
            version: '6.4.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 301',
        code: "require('process').argv0",
        options: [
          {
            version: '6.4.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 302',
        code: "var c = require('process'); c.argv0",
        options: [
          {
            version: '6.4.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 303',
        code: "var { argv0 } = require('process'); argv0",
        options: [
          {
            version: '6.4.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 304',
        code: "import c from 'process'; c.argv0",
        options: [
          {
            version: '6.4.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 305',
        code: 'process.channel',
        options: [
          {
            version: '7.1.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 306',
        code: 'process.cpuUsage',
        options: [
          {
            version: '6.1.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 307',
        code: 'process.emitWarning',
        options: [
          {
            version: '6.0.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 308',
        code: 'process.getegid',
        options: [
          {
            version: '2.0.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 309',
        code: 'process.geteuid',
        options: [
          {
            version: '2.0.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 310',
        code: 'process.hasUncaughtExceptionCaptureCallback',
        options: [
          {
            version: '9.3.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 311',
        code: 'process.ppid',
        options: [
          {
            version: '9.2.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 312',
        code: 'process.release',
        options: [
          {
            version: '3.0.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 313',
        code: 'process.setegid',
        options: [
          {
            version: '2.0.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 314',
        code: 'process.seteuid',
        options: [
          {
            version: '2.0.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 315',
        code: 'process.setUncaughtExceptionCaptureCallback',
        options: [
          {
            version: '9.3.0',
          },
        ],
      },
      {
        name: 'process: upstream valid 316',
        code: 'process.argv0',
        options: [
          {
            version: '6.3.9',
            ignores: ['process.argv0'],
          },
        ],
      },
      {
        name: 'process: upstream valid 317',
        code: "require('process').argv0",
        options: [
          {
            version: '6.3.9',
            ignores: ['process.argv0'],
          },
        ],
      },
      {
        name: 'process: upstream valid 318',
        code: "var c = require('process'); c.argv0",
        options: [
          {
            version: '6.3.9',
            ignores: ['process.argv0'],
          },
        ],
      },
      {
        name: 'process: upstream valid 319',
        code: "var { argv0 } = require('process'); argv0",
        options: [
          {
            version: '6.3.9',
            ignores: ['process.argv0'],
          },
        ],
      },
      {
        name: 'process: upstream valid 320',
        code: "import c from 'process'; c.argv0",
        options: [
          {
            version: '6.3.9',
            ignores: ['process.argv0'],
          },
        ],
      },
      {
        name: 'process: upstream valid 321',
        code: 'process.channel',
        options: [
          {
            version: '7.0.9',
            ignores: ['process.channel'],
          },
        ],
      },
      {
        name: 'process: upstream valid 322',
        code: 'process.cpuUsage',
        options: [
          {
            version: '6.0.9',
            ignores: ['process.cpuUsage'],
          },
        ],
      },
      {
        name: 'process: upstream valid 323',
        code: 'process.emitWarning',
        options: [
          {
            version: '5.9.9',
            ignores: ['process.emitWarning'],
          },
        ],
      },
      {
        name: 'process: upstream valid 324',
        code: 'process.getegid',
        options: [
          {
            version: '1.9.9',
            ignores: ['process.getegid'],
          },
        ],
      },
      {
        name: 'process: upstream valid 325',
        code: 'process.geteuid',
        options: [
          {
            version: '1.9.9',
            ignores: ['process.geteuid'],
          },
        ],
      },
      {
        name: 'process: upstream valid 326',
        code: 'process.hasUncaughtExceptionCaptureCallback',
        options: [
          {
            version: '9.2.9',
            ignores: ['process.hasUncaughtExceptionCaptureCallback'],
          },
        ],
      },
      {
        name: 'process: upstream valid 327',
        code: 'process.ppid',
        options: [
          {
            version: '9.1.9',
            ignores: ['process.ppid'],
          },
        ],
      },
      {
        name: 'process: upstream valid 328',
        code: 'process.release',
        options: [
          {
            version: '2.9.9',
            ignores: ['process.release'],
          },
        ],
      },
      {
        name: 'process: upstream valid 329',
        code: 'process.setegid',
        options: [
          {
            version: '1.9.9',
            ignores: ['process.setegid'],
          },
        ],
      },
      {
        name: 'process: upstream valid 330',
        code: 'process.seteuid',
        options: [
          {
            version: '1.9.9',
            ignores: ['process.seteuid'],
          },
        ],
      },
      {
        name: 'process: upstream valid 331',
        code: 'process.setUncaughtExceptionCaptureCallback',
        options: [
          {
            version: '9.2.9',
            ignores: ['process.setUncaughtExceptionCaptureCallback'],
          },
        ],
      },
      {
        name: 'stream: upstream valid 332',
        code: "require('stream').finished()",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'stream: upstream valid 333',
        code: "var hooks = require('stream'); hooks.finished()",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'stream: upstream valid 334',
        code: "var { finished } = require('stream'); finished()",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'stream: upstream valid 335',
        code: "import stream from 'stream'; stream.finished()",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'stream: upstream valid 336',
        code: "import { finished } from 'stream'; finished()",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'stream: upstream valid 337',
        code: "require('stream').pipeline()",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'stream: upstream valid 338',
        code: "require('stream').finished()",
        options: [
          {
            version: '9.9.9',
            ignores: ['stream.finished'],
          },
        ],
      },
      {
        name: 'stream: upstream valid 339',
        code: "var hooks = require('stream'); hooks.finished()",
        options: [
          {
            version: '9.9.9',
            ignores: ['stream.finished'],
          },
        ],
      },
      {
        name: 'stream: upstream valid 340',
        code: "var { finished } = require('stream'); finished()",
        options: [
          {
            version: '9.9.9',
            ignores: ['stream.finished'],
          },
        ],
      },
      {
        name: 'stream: upstream valid 341',
        code: "import stream from 'stream'; stream.finished()",
        options: [
          {
            version: '9.9.9',
            ignores: ['stream.finished'],
          },
        ],
      },
      {
        name: 'stream: upstream valid 342',
        code: "import { finished } from 'stream'; finished()",
        options: [
          {
            version: '9.9.9',
            ignores: ['stream.finished'],
          },
        ],
      },
      {
        name: 'stream: upstream valid 343',
        code: "require('stream').pipeline()",
        options: [
          {
            version: '9.9.9',
            ignores: ['stream.pipeline'],
          },
        ],
      },
      {
        name: 'trace_events: upstream valid 344',
        code: "require('trace_events')",
        options: [
          {
            version: '10.0.0',
            ignores: ['trace_events'],
          },
        ],
      },
      {
        name: 'trace_events: upstream valid 345',
        code: "import trace_events from 'trace_events'",
        options: [
          {
            version: '10.0.0',
            ignores: ['trace_events'],
          },
        ],
      },
      {
        name: 'url: upstream valid 346',
        code: 'URL',
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 347',
        code: 'URLSearchParams',
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 348',
        code: "require('url').URL",
        options: [
          {
            version: '7.0.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 349',
        code: "require('url').URL",
        options: [
          {
            version: '6.13.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 350',
        code: "var cp = require('url'); cp.URL",
        options: [
          {
            version: '7.0.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 351',
        code: "var { URL } = require('url');",
        options: [
          {
            version: '7.0.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 352',
        code: "import cp from 'url'; cp.URL",
        options: [
          {
            version: '7.0.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 353',
        code: "import { URL } from 'url'",
        options: [
          {
            version: '7.0.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 354',
        code: "require('url').URLSearchParams",
        options: [
          {
            version: '7.5.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 355',
        code: "require('url').URLSearchParams",
        options: [
          {
            version: '6.13.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 356',
        code: "require('url').domainToASCII",
        options: [
          {
            version: '7.4.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 357',
        code: "require('url').domainToUnicode",
        options: [
          {
            version: '7.4.0',
          },
        ],
      },
      {
        name: 'url: upstream valid 358',
        code: 'URL',
        options: [
          {
            version: '9.9.9',
            ignores: ['URL'],
          },
        ],
      },
      {
        name: 'url: upstream valid 359',
        code: 'URLSearchParams',
        options: [
          {
            version: '9.9.9',
            ignores: ['URLSearchParams'],
          },
        ],
      },
      {
        name: 'url: upstream valid 360',
        code: "require('url').URL",
        options: [
          {
            version: '6.9.9',
            ignores: ['url.URL'],
          },
        ],
      },
      {
        name: 'url: upstream valid 361',
        code: "var cp = require('url'); cp.URL",
        options: [
          {
            version: '6.9.9',
            ignores: ['url.URL'],
          },
        ],
      },
      {
        name: 'url: upstream valid 362',
        code: "var { URL } = require('url');",
        options: [
          {
            version: '6.9.9',
            ignores: ['url.URL'],
          },
        ],
      },
      {
        name: 'url: upstream valid 363',
        code: "import cp from 'url'; cp.URL",
        options: [
          {
            version: '6.9.9',
            ignores: ['url.URL'],
          },
        ],
      },
      {
        name: 'url: upstream valid 364',
        code: "import { URL } from 'url'",
        options: [
          {
            version: '6.9.9',
            ignores: ['url.URL'],
          },
        ],
      },
      {
        name: 'url: upstream valid 365',
        code: "require('url').URLSearchParams",
        options: [
          {
            version: '7.4.9',
            ignores: ['url.URLSearchParams'],
          },
        ],
      },
      {
        name: 'url: upstream valid 366',
        code: "require('url').domainToASCII",
        options: [
          {
            version: '7.3.9',
            ignores: ['url.domainToASCII'],
          },
        ],
      },
      {
        name: 'url: upstream valid 367',
        code: "require('url').domainToUnicode",
        options: [
          {
            version: '7.3.9',
            ignores: ['url.domainToUnicode'],
          },
        ],
      },
      {
        name: 'util: upstream valid 368',
        code: "require('util').callbackify",
        options: [
          {
            version: '8.2.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 369',
        code: "var hooks = require('util'); hooks.callbackify",
        options: [
          {
            version: '8.2.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 370',
        code: "var { callbackify } = require('util'); callbackify",
        options: [
          {
            version: '8.2.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 371',
        code: "import util from 'util'; util.callbackify",
        options: [
          {
            version: '8.2.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 372',
        code: "import { callbackify } from 'util'; callbackify",
        options: [
          {
            version: '8.2.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 373',
        code: "require('util').formatWithOptions",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 374',
        code: "require('util').getSystemErrorName",
        options: [
          {
            version: '9.7.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 375',
        code: "require('util').inspect.custom",
        options: [
          {
            version: '6.6.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 376',
        code: "require('util').inspect.defaultOptions",
        options: [
          {
            version: '6.4.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 377',
        code: "require('util').isDeepStrictEqual",
        options: [
          {
            version: '9.0.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 378',
        code: "require('util').promisify",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 379',
        code: "require('util').TextDecoder",
        options: [
          {
            version: '8.9.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 380',
        code: "require('util').TextEncoder",
        options: [
          {
            version: '8.9.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 381',
        code: "require('util').types",
        options: [
          {
            version: '10.0.0',
          },
        ],
      },
      {
        name: 'util: upstream valid 382',
        code: "require('util').styleText",
        options: [
          {
            version: '21.7.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'util: upstream valid 383',
        code: "import { styleText } from 'node:util'; styleText('green', 'ok')",
        options: [
          {
            version: '21.7.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'util: upstream valid 384',
        code: "require('util').callbackify",
        options: [
          {
            version: '8.1.9',
            ignores: ['util.callbackify'],
          },
        ],
      },
      {
        name: 'util: upstream valid 385',
        code: "var hooks = require('util'); hooks.callbackify",
        options: [
          {
            version: '8.1.9',
            ignores: ['util.callbackify'],
          },
        ],
      },
      {
        name: 'util: upstream valid 386',
        code: "var { callbackify } = require('util'); callbackify",
        options: [
          {
            version: '8.1.9',
            ignores: ['util.callbackify'],
          },
        ],
      },
      {
        name: 'util: upstream valid 387',
        code: "import util from 'util'; util.callbackify",
        options: [
          {
            version: '8.1.9',
            ignores: ['util.callbackify'],
          },
        ],
      },
      {
        name: 'util: upstream valid 388',
        code: "import { callbackify } from 'util'; callbackify",
        options: [
          {
            version: '8.1.9',
            ignores: ['util.callbackify'],
          },
        ],
      },
      {
        name: 'util: upstream valid 389',
        code: "require('util').formatWithOptions",
        options: [
          {
            version: '9.9.9',
            ignores: ['util.formatWithOptions'],
          },
        ],
      },
      {
        name: 'util: upstream valid 390',
        code: "require('util').getSystemErrorName",
        options: [
          {
            version: '9.6.9',
            ignores: ['util.getSystemErrorName'],
          },
        ],
      },
      {
        name: 'util: upstream valid 391',
        code: "require('util').inspect.custom",
        options: [
          {
            version: '6.5.9',
            ignores: ['util.inspect.custom'],
          },
        ],
      },
      {
        name: 'util: upstream valid 392',
        code: "require('util').inspect.defaultOptions",
        options: [
          {
            version: '6.3.9',
            ignores: ['util.inspect.defaultOptions'],
          },
        ],
      },
      {
        name: 'util: upstream valid 393',
        code: "require('util').isDeepStrictEqual",
        options: [
          {
            version: '8.9.9',
            ignores: ['util.isDeepStrictEqual'],
          },
        ],
      },
      {
        name: 'util: upstream valid 394',
        code: "require('util').promisify",
        options: [
          {
            version: '7.9.9',
            ignores: ['util.promisify'],
          },
        ],
      },
      {
        name: 'util: upstream valid 395',
        code: "require('util').TextDecoder",
        options: [
          {
            version: '8.2.9',
            ignores: ['util.TextDecoder'],
          },
        ],
      },
      {
        name: 'util: upstream valid 396',
        code: "require('util').TextEncoder",
        options: [
          {
            version: '8.2.9',
            ignores: ['util.TextEncoder'],
          },
        ],
      },
      {
        name: 'util: upstream valid 397',
        code: "require('util').types",
        options: [
          {
            version: '9.9.9',
            ignores: ['util.types'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 398',
        code: "require('v8')",
        options: [
          {
            version: '1.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 399',
        code: "import hooks from 'v8'",
        options: [
          {
            version: '1.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 400',
        code: "require('v8').cachedDataVersionTag()",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 401',
        code: "var hooks = require('v8'); hooks.cachedDataVersionTag()",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 402',
        code: "var { cachedDataVersionTag } = require('v8'); cachedDataVersionTag()",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 403',
        code: "import v8 from 'v8'; v8.cachedDataVersionTag()",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 404',
        code: "import { cachedDataVersionTag } from 'v8'; cachedDataVersionTag()",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 405',
        code: "require('v8').getHeapSpaceStatistics()",
        options: [
          {
            version: '6.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 406',
        code: "require('v8').serialize()",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 407',
        code: "require('v8').deserialize()",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 408',
        code: "require('v8').Serializer",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 409',
        code: "require('v8').Deserializer",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 410',
        code: "require('v8').DefaultSerializer",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 411',
        code: "require('v8').DefaultDeserializer",
        options: [
          {
            version: '8.0.0',
          },
        ],
      },
      {
        name: 'v8: upstream valid 412',
        code: "require('v8')",
        options: [
          {
            version: '0.12.99',
            ignores: ['v8'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 413',
        code: "import hooks from 'v8'",
        options: [
          {
            version: '0.12.99',
            ignores: ['v8'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 414',
        code: "import { cachedDataVersionTag } from 'v8'",
        options: [
          {
            version: '0.12.99',
            ignores: ['v8', 'v8.cachedDataVersionTag'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 415',
        code: "require('v8').cachedDataVersionTag()",
        options: [
          {
            version: '7.9.9',
            ignores: ['v8.cachedDataVersionTag'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 416',
        code: "var hooks = require('v8'); hooks.cachedDataVersionTag()",
        options: [
          {
            version: '7.9.9',
            ignores: ['v8.cachedDataVersionTag'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 417',
        code: "var { cachedDataVersionTag } = require('v8'); cachedDataVersionTag()",
        options: [
          {
            version: '7.9.9',
            ignores: ['v8.cachedDataVersionTag'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 418',
        code: "import v8 from 'v8'; v8.cachedDataVersionTag()",
        options: [
          {
            version: '7.9.9',
            ignores: ['v8.cachedDataVersionTag'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 419',
        code: "import { cachedDataVersionTag } from 'v8'; cachedDataVersionTag()",
        options: [
          {
            version: '7.9.9',
            ignores: ['v8.cachedDataVersionTag'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 420',
        code: "require('v8').getHeapSpaceStatistics()",
        options: [
          {
            version: '5.9.9',
            ignores: ['v8.getHeapSpaceStatistics'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 421',
        code: "require('v8').serialize()",
        options: [
          {
            version: '7.9.9',
            ignores: ['v8.serialize'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 422',
        code: "require('v8').deserialize()",
        options: [
          {
            version: '7.9.9',
            ignores: ['v8.deserialize'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 423',
        code: "require('v8').Serializer",
        options: [
          {
            version: '7.9.9',
            ignores: ['v8.Serializer'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 424',
        code: "require('v8').Deserializer",
        options: [
          {
            version: '7.9.9',
            ignores: ['v8.Deserializer'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 425',
        code: "require('v8').DefaultSerializer",
        options: [
          {
            version: '7.9.9',
            ignores: ['v8.DefaultSerializer'],
          },
        ],
      },
      {
        name: 'v8: upstream valid 426',
        code: "require('v8').DefaultDeserializer",
        options: [
          {
            version: '7.9.9',
            ignores: ['v8.DefaultDeserializer'],
          },
        ],
      },
      {
        name: 'vm: upstream valid 427',
        code: "require('vm')",
        options: [
          {
            version: '9.6.0',
          },
        ],
      },
      {
        name: 'vm: upstream valid 428',
        code: "import vm from 'vm';",
        options: [
          {
            version: '9.6.0',
          },
        ],
      },
      {
        name: 'vm: upstream valid 429',
        code: "import * as vm from 'vm';",
        options: [
          {
            version: '9.6.0',
          },
        ],
      },
      {
        name: 'vm: upstream valid 430',
        code: "require('vm').Module",
        options: [
          {
            version: '9.5.9',
            ignores: ['vm.Module'],
          },
        ],
      },
      {
        name: 'vm: upstream valid 431',
        code: "var vm = require('vm'); vm.Module",
        options: [
          {
            version: '9.5.9',
            ignores: ['vm.Module'],
          },
        ],
      },
      {
        name: 'vm: upstream valid 432',
        code: "var { Module } = require('vm'); Module",
        options: [
          {
            version: '9.5.9',
            ignores: ['vm.Module'],
          },
        ],
      },
      {
        name: 'vm: upstream valid 433',
        code: "import vm from 'vm'; vm.Module",
        options: [
          {
            version: '9.5.9',
            ignores: ['vm.Module'],
          },
        ],
      },
      {
        name: 'vm: upstream valid 434',
        code: "import { Module } from 'vm'; Module",
        options: [
          {
            version: '9.5.9',
            ignores: ['vm.Module'],
          },
        ],
      },
      {
        name: 'worker_threads: upstream valid 435',
        code: "require('worker_threads')",
        options: [
          {
            version: '10.4.99',
            ignores: ['worker_threads'],
          },
        ],
      },
      {
        name: 'worker_threads: upstream valid 436',
        code: "import worker_threads from 'worker_threads'",
        options: [
          {
            version: '10.4.99',
            ignores: ['worker_threads'],
          },
        ],
      },
      {
        name: 'worker_threads: upstream valid 437',
        code: "require('worker_threads')",
        options: [
          {
            version: '12.11.0',
          },
        ],
      },
      {
        name: 'worker_threads: upstream valid 438',
        code: "import worker_threads from 'worker_threads'",
        options: [
          {
            version: '12.11.0',
          },
        ],
      },
      {
        name: 'worker_threads: upstream valid 439',
        code: "import worker_threads from 'worker_threads'",
        settings: {
          node: {
            version: '12.11.0',
          },
        },
      },
      {
        name: 'timers/promises: upstream valid 440',
        code: "\n                        import { scheduler } from 'node:timers/promises';\n                        await scheduler.wait( 1000 );\n                    ",
        options: [
          {
            version: '>= 20.0.0',
            ignores: ['timers/promises.scheduler.wait'],
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 441',
        code: 'import.meta.resolve(specifier)',
        options: [
          {
            version: '22.0.0',
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 442',
        code: 'import.meta.resolve(specifier)',
        options: [
          {
            version: '20.6.0',
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 443',
        code: 'import.meta.resolve(specifier)',
        options: [
          {
            version: '18.19.0',
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 444',
        code: 'import.meta.resolve(specifier)',
        options: [
          {
            version: '13.9.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 445',
        code: 'import.meta.resolve(specifier)',
        options: [
          {
            version: '12.16.2',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 446',
        code: 'import.meta.resolve(specifier)',
        options: [
          {
            version: '18.18.0',
            ignores: ['import.meta.resolve'],
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 447',
        code: 'import.meta.dirname;',
        options: [
          {
            version: '22.16.0',
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 448',
        code: 'import.meta.dirname;',
        options: [
          {
            version: '22.0.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 449',
        code: 'import.meta.dirname;',
        options: [
          {
            version: '21.2.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 450',
        code: 'import.meta.dirname;',
        options: [
          {
            version: '20.11.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 451',
        code: 'import.meta.dirname;',
        options: [
          {
            version: '20.10.0',
            ignores: ['import.meta.dirname'],
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 452',
        code: 'import.meta.filename;',
        options: [
          {
            version: '22.16.0',
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 453',
        code: 'import.meta.filename;',
        options: [
          {
            version: '22.0.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 454',
        code: 'import.meta.filename;',
        options: [
          {
            version: '21.2.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 455',
        code: 'import.meta.filename;',
        options: [
          {
            version: '20.11.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'import.meta: upstream valid 456',
        code: 'import.meta.filename;',
        options: [
          {
            version: '20.10.0',
            ignores: ['import.meta.filename'],
          },
        ],
      },
      {
        name: 'fetch: upstream valid 457',
        code: "fetch('/asd')",
        options: [
          {
            version: '16.16.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'sqlite: upstream valid 458',
        code: "\n                        import { DatabaseSync } from 'node:sqlite';\n                        const database = new DatabaseSync(':memory:');\n                    ",
        options: [
          {
            version: '>=22.5.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'sqlite: upstream valid 459',
        code: "\n                        import { DatabaseSync } from 'node:sqlite';\n                        const database = new DatabaseSync(':memory:');\n                    ",
        options: [
          {
            version: '>=22.3.0',
            ignores: ['sqlite', 'sqlite.DatabaseSync'],
          },
        ],
      },
      {
        name: 'sqlite: upstream valid 460',
        code: "\n                        const { DatabaseSync } = require('node:sqlite');\n                        const database = new DatabaseSync(':memory:');\n                    ",
        options: [
          {
            version: '>=22.5.0',
            allowExperimental: true,
          },
        ],
      },
      {
        name: 'sqlite: upstream valid 461',
        code: "\n                        const { DatabaseSync } = process.getBuiltinModule('node:sqlite');\n                        const database = new DatabaseSync(':memory:');\n                    ",
        options: [
          {
            version: '>=22.5.0',
            allowExperimental: true,
          },
        ],
      },
    ],
    invalid: [
      {
        name: 'assert: upstream invalid 1',
        code: "require('assert').deepStrictEqual()",
        options: [
          {
            version: '1.1.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.deepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 2',
        code: "var assert = require('assert'); assert.deepStrictEqual()",
        options: [
          {
            version: '1.1.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.deepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.",
            line: 1,
            column: 33,
            endLine: 1,
            endColumn: 55,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 3',
        code: "var { deepStrictEqual } = require('assert'); deepStrictEqual()",
        options: [
          {
            version: '1.1.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.deepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 4',
        code: "import assert from 'assert'; assert.deepStrictEqual()",
        options: [
          {
            version: '1.1.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.deepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.",
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 52,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 5',
        code: "import { deepStrictEqual } from 'assert'; deepStrictEqual()",
        options: [
          {
            version: '1.1.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.deepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 25,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 6',
        code: "require('assert').notDeepStrictEqual()",
        options: [
          {
            version: '1.1.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.notDeepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 7',
        code: "require('assert').rejects()",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.rejects' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 8',
        code: "require('assert').doesNotReject()",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.doesNotReject' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 9',
        code: "require('assert').strict.rejects()",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.strict.rejects' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 10',
        code: "require('assert').strict.doesNotReject()",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.strict.doesNotReject' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 11',
        code: "var assert = require('assert').strict",
        options: [
          {
            version: '9.8.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.strict' is still an experimental feature and is not supported until Node.js 9.9.0 (backported: ^8.13.0). The configured version range is '9.8.9'.",
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 12',
        code: "var {strict: assert} = require('assert'); assert.rejects()",
        options: [
          {
            version: '9.8.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.strict' is still an experimental feature and is not supported until Node.js 9.9.0 (backported: ^8.13.0). The configured version range is '9.8.9'.",
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 20,
          },
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert.strict.rejects' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.8.9'.",
            line: 1,
            column: 43,
            endLine: 1,
            endColumn: 57,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 13',
        code: "const { CallTracker } = require('assert'); new CallTracker();",
        options: [
          {
            version: '14.2.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'assert.CallTracker' is still an experimental feature The configured version range is '14.2.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 14',
        code: "import { CallTracker } from 'assert'; new CallTracker();",
        options: [
          {
            version: '14.2.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'assert.CallTracker' is still an experimental feature The configured version range is '14.2.0'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 15',
        code: "require('node:assert').deepStrictEqual()",
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert' is still an experimental feature and is not supported until Node.js 14.13.1 (backported: ^12.20.0). The configured version range is '3.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'assert: upstream invalid 16',
        code: "import assert from 'node:assert';",
        options: [
          {
            version: '3.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'assert' is still an experimental feature and is not supported until Node.js 14.13.1 (backported: ^12.20.0). The configured version range is '3.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 17',
        code: "require('async_hooks')",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 18',
        code: "import hooks from 'async_hooks'",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 19',
        code: "require('async_hooks').createHook()",
        options: [
          {
            version: '16.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 20',
        code: "const { createHook } = require('async_hooks')",
        options: [
          {
            version: '16.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 21',
        code: "const { createHook } = require('async_hooks'); createHook()",
        options: [
          {
            version: '16.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 22',
        code: "const hooks = require('async_hooks'); hooks.createHook()",
        options: [
          {
            version: '16.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.",
            line: 1,
            column: 39,
            endLine: 1,
            endColumn: 55,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 23',
        code: "import { createHook } from 'async_hooks'",
        options: [
          {
            version: '16.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 24',
        code: "import { createHook } from 'async_hooks'; createHook()",
        options: [
          {
            version: '16.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 25',
        code: "import async_hooks from 'async_hooks'; async_hooks.createHook()",
        options: [
          {
            version: '16.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.",
            line: 1,
            column: 40,
            endLine: 1,
            endColumn: 62,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 26',
        code: "const hooks = require('async_hooks'); new hooks.AsyncLocalStorage()",
        options: [
          {
            version: '13.9.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.",
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 37,
          },
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks.AsyncLocalStorage' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.",
            line: 1,
            column: 43,
            endLine: 1,
            endColumn: 66,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 27',
        code: "import * as hooks from 'async_hooks'; new hooks.AsyncLocalStorage()",
        options: [
          {
            version: '13.9.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 38,
          },
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks.AsyncLocalStorage' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.",
            line: 1,
            column: 43,
            endLine: 1,
            endColumn: 66,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 28',
        code: "import hooks from 'async_hooks'; new hooks.AsyncLocalStorage()",
        options: [
          {
            version: '13.9.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks.AsyncLocalStorage' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.",
            line: 1,
            column: 38,
            endLine: 1,
            endColumn: 61,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 29',
        code: "import { AsyncLocalStorage } from 'async_hooks'; new AsyncLocalStorage()",
        options: [
          {
            version: '13.9.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 49,
          },
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks.AsyncLocalStorage' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 30',
        code: "require('node:async_hooks')",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: 'async_hooks: upstream invalid 31',
        code: "import hooks from 'node:async_hooks'",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 32',
        code: 'Buffer.alloc',
        options: [
          {
            version: '4.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'Buffer.alloc' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 13,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 33',
        code: 'Buffer.allocUnsafe',
        options: [
          {
            version: '4.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'Buffer.allocUnsafe' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 34',
        code: 'Buffer.allocUnsafeSlow',
        options: [
          {
            version: '4.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'Buffer.allocUnsafeSlow' is still an experimental feature and is not supported until Node.js 5.12.0 (backported: ^4.5.0). The configured version range is '4.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 35',
        code: 'Buffer.from',
        options: [
          {
            version: '4.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'Buffer.from' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 12,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 36',
        code: "require('buffer').constants",
        options: [
          {
            version: '8.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.constants' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 37',
        code: "var cp = require('buffer'); cp.constants",
        options: [
          {
            version: '8.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.constants' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 41,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 38',
        code: "var { constants } = require('buffer');",
        options: [
          {
            version: '8.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.constants' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 39',
        code: "import cp from 'buffer'; cp.constants",
        options: [
          {
            version: '8.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.constants' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.",
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 40',
        code: "import { constants } from 'buffer'",
        options: [
          {
            version: '8.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.constants' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 41',
        code: "var {Buffer: b} = require('buffer'); b.alloc",
        options: [
          {
            version: '4.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.Buffer.alloc' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.",
            line: 1,
            column: 38,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 42',
        code: "var {Buffer: b} = require('buffer'); b.allocUnsafe",
        options: [
          {
            version: '4.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.Buffer.allocUnsafe' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.",
            line: 1,
            column: 38,
            endLine: 1,
            endColumn: 51,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 43',
        code: "var {Buffer: b} = require('buffer'); b.allocUnsafeSlow",
        options: [
          {
            version: '4.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.Buffer.allocUnsafeSlow' is still an experimental feature and is not supported until Node.js 5.12.0 (backported: ^4.5.0). The configured version range is '4.4.9'.",
            line: 1,
            column: 38,
            endLine: 1,
            endColumn: 55,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 44',
        code: "var {Buffer: b} = require('buffer'); b.from",
        options: [
          {
            version: '4.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.Buffer.from' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.",
            line: 1,
            column: 38,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 45',
        code: "require('buffer').kMaxLength",
        options: [
          {
            version: '2.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.kMaxLength' is still an experimental feature and is not supported until Node.js 3.0.0. The configured version range is '2.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 29,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 46',
        code: "require('buffer').transcode",
        options: [
          {
            version: '7.0.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.transcode' is still an experimental feature and is not supported until Node.js 7.1.0. The configured version range is '7.0.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 47',
        code: "const { Blob } = require('buffer'); new Blob();",
        options: [
          {
            version: '15.7.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.Blob' is still an experimental feature and is not supported until Node.js 18.0.0 (backported: ^16.17.0). The configured version range is '15.7.0'.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 13,
          },
        ],
      },
      {
        name: 'buffer: upstream invalid 48',
        code: "import buffer from 'buffer'; new buffer.Blob();",
        options: [
          {
            version: '15.7.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'buffer.Blob' is still an experimental feature and is not supported until Node.js 18.0.0 (backported: ^16.17.0). The configured version range is '15.7.0'.",
            line: 1,
            column: 34,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'child_process: upstream invalid 49',
        code: "require('child_process').ChildProcess",
        options: [
          {
            version: '2.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'child_process.ChildProcess' is still an experimental feature and is not supported until Node.js 2.2.0. The configured version range is '2.1.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 38,
          },
        ],
      },
      {
        name: 'child_process: upstream invalid 50',
        code: "var cp = require('child_process'); cp.ChildProcess",
        options: [
          {
            version: '2.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'child_process.ChildProcess' is still an experimental feature and is not supported until Node.js 2.2.0. The configured version range is '2.1.9'.",
            line: 1,
            column: 36,
            endLine: 1,
            endColumn: 51,
          },
        ],
      },
      {
        name: 'child_process: upstream invalid 51',
        code: "var { ChildProcess } = require('child_process'); ChildProcess",
        options: [
          {
            version: '2.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'child_process.ChildProcess' is still an experimental feature and is not supported until Node.js 2.2.0. The configured version range is '2.1.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'child_process: upstream invalid 52',
        code: "import cp from 'child_process'; cp.ChildProcess",
        options: [
          {
            version: '2.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'child_process.ChildProcess' is still an experimental feature and is not supported until Node.js 2.2.0. The configured version range is '2.1.9'.",
            line: 1,
            column: 33,
            endLine: 1,
            endColumn: 48,
          },
        ],
      },
      {
        name: 'child_process: upstream invalid 53',
        code: "import { ChildProcess } from 'child_process'",
        options: [
          {
            version: '2.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'child_process.ChildProcess' is still an experimental feature and is not supported until Node.js 2.2.0. The configured version range is '2.1.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'console: upstream invalid 54',
        code: 'console.clear()',
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.clear' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        name: 'console: upstream invalid 55',
        code: "require('console').clear()",
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.clear' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
      },
      {
        name: 'console: upstream invalid 56',
        code: "var c = require('console'); c.clear()",
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.clear' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'console: upstream invalid 57',
        code: "var { clear } = require('console'); clear()",
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.clear' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 12,
          },
        ],
      },
      {
        name: 'console: upstream invalid 58',
        code: "import c from 'console'; c.clear()",
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.clear' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.",
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'console: upstream invalid 59',
        code: 'console.count()',
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.count' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        name: 'console: upstream invalid 60',
        code: 'console.countReset()',
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.countReset' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'console: upstream invalid 61',
        code: 'console.debug()',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.debug' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        name: 'console: upstream invalid 62',
        code: 'console.dirxml()',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.dirxml' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 15,
          },
        ],
      },
      {
        name: 'console: upstream invalid 63',
        code: 'console.group()',
        options: [
          {
            version: '8.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.group' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        name: 'console: upstream invalid 64',
        code: 'console.groupCollapsed()',
        options: [
          {
            version: '8.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.groupCollapsed' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'console: upstream invalid 65',
        code: 'console.groupEnd()',
        options: [
          {
            version: '8.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.groupEnd' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'console: upstream invalid 66',
        code: 'console.table()',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.table' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        name: 'console: upstream invalid 67',
        code: 'console.profile()',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.profile' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'console: upstream invalid 68',
        code: 'console.profileEnd()',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.profileEnd' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'console: upstream invalid 69',
        code: 'console.timeStamp()',
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'console.timeStamp' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 70',
        code: "require('crypto').constants",
        options: [
          {
            version: '6.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.constants' is still an experimental feature and is not supported until Node.js 6.3.0. The configured version range is '6.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 71',
        code: "var hooks = require('crypto'); hooks.constants",
        options: [
          {
            version: '6.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.constants' is still an experimental feature and is not supported until Node.js 6.3.0. The configured version range is '6.2.9'.",
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 47,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 72',
        code: "var { constants } = require('crypto'); constants",
        options: [
          {
            version: '6.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.constants' is still an experimental feature and is not supported until Node.js 6.3.0. The configured version range is '6.2.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 73',
        code: "import crypto from 'crypto'; crypto.constants",
        options: [
          {
            version: '6.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.constants' is still an experimental feature and is not supported until Node.js 6.3.0. The configured version range is '6.2.9'.",
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 46,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 74',
        code: "import { constants } from 'crypto'; constants",
        options: [
          {
            version: '6.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.constants' is still an experimental feature and is not supported until Node.js 6.3.0. The configured version range is '6.2.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 75',
        code: "require('crypto').Certificate.exportChallenge()",
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.Certificate.exportChallenge' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 46,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 76',
        code: "var { Certificate: c } = require('crypto'); c.exportChallenge()",
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.Certificate.exportChallenge' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 45,
            endLine: 1,
            endColumn: 62,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 77',
        code: "var { Certificate: c } = require('crypto'); c.exportPublicKey()",
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.Certificate.exportPublicKey' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 45,
            endLine: 1,
            endColumn: 62,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 78',
        code: "var { Certificate: c } = require('crypto'); c.verifySpkac()",
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.Certificate.verifySpkac' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 45,
            endLine: 1,
            endColumn: 58,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 79',
        code: "require('crypto').fips",
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.fips' is still an experimental feature and is not supported until Node.js 6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 80',
        code: "require('crypto').getCurves",
        options: [
          {
            version: '2.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.getCurves' is still an experimental feature and is not supported until Node.js 2.3.0. The configured version range is '2.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 81',
        code: "require('crypto').getFips",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.getFips' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 82',
        code: "require('crypto').privateEncrypt",
        options: [
          {
            version: '1.0.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.privateEncrypt' is still an experimental feature and is not supported until Node.js 1.1.0. The configured version range is '1.0.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 83',
        code: "require('crypto').publicDecrypt",
        options: [
          {
            version: '1.0.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.publicDecrypt' is still an experimental feature and is not supported until Node.js 1.1.0. The configured version range is '1.0.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 84',
        code: "require('crypto').randomFillSync",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.randomFillSync' is still an experimental feature and is not supported until Node.js 7.10.0 (backported: ^6.13.0). The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 85',
        code: "require('crypto').randomFill",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.randomFill' is still an experimental feature and is not supported until Node.js 7.10.0 (backported: ^6.13.0). The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 29,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 86',
        code: "require('crypto').scrypt",
        options: [
          {
            version: '10.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.scrypt' is still an experimental feature and is not supported until Node.js 10.5.0. The configured version range is '10.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 87',
        code: "require('crypto').scryptSync",
        options: [
          {
            version: '10.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.scryptSync' is still an experimental feature and is not supported until Node.js 10.5.0. The configured version range is '10.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 29,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 88',
        code: "require('crypto').setFips",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.setFips' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'crypto: upstream invalid 89',
        code: "require('crypto').timingSafeEqual",
        options: [
          {
            version: '6.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'crypto.timingSafeEqual' is still an experimental feature and is not supported until Node.js 6.6.0. The configured version range is '6.5.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'dns: upstream invalid 90',
        code: "require('dns').Resolver",
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 24,
          },
        ],
      },
      {
        name: 'dns: upstream invalid 91',
        code: "var hooks = require('dns'); hooks.Resolver",
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.2.9'.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 43,
          },
        ],
      },
      {
        name: 'dns: upstream invalid 92',
        code: "var { Resolver } = require('dns'); Resolver",
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.2.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 15,
          },
        ],
      },
      {
        name: 'dns: upstream invalid 93',
        code: "import dns from 'dns'; dns.Resolver",
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.2.9'.",
            line: 1,
            column: 24,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'dns: upstream invalid 94',
        code: "import { Resolver } from 'dns'; Resolver",
        options: [
          {
            version: '8.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.2.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'dns: upstream invalid 95',
        code: "require('dns').resolvePtr",
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'dns.resolvePtr' is still an experimental feature and is not supported until Node.js 6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'dns: upstream invalid 96',
        code: "require('dns').promises",
        options: [
          {
            version: '11.13.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'dns.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 24,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 97',
        code: "require('fs').promises",
        options: [
          {
            version: '11.13.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 98',
        code: "var fs = require('fs'); fs.promises",
        options: [
          {
            version: '11.13.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.",
            line: 1,
            column: 25,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 99',
        code: "var { promises } = require('fs'); promises",
        options: [
          {
            version: '11.13.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 15,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 100',
        code: "import fs from 'fs'; fs.promises",
        options: [
          {
            version: '11.13.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.",
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 101',
        code: "import { promises } from 'fs'",
        options: [
          {
            version: '11.13.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 102',
        code: "require('fs').copyFile",
        options: [
          {
            version: '8.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.copyFile' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 103',
        code: "require('fs').copyFileSync",
        options: [
          {
            version: '8.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.copyFileSync' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 104',
        code: "require('fs').mkdtemp",
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.mkdtemp' is still an experimental feature and is not supported until Node.js 5.10.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 105',
        code: "require('fs').mkdtempSync",
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.mkdtempSync' is still an experimental feature and is not supported until Node.js 5.10.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 106',
        code: "require('fs').realpath.native",
        options: [
          {
            version: '9.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.realpath.native' is still an experimental feature and is not supported until Node.js 9.2.0. The configured version range is '9.1.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 30,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 107',
        code: "require('fs').realpathSync.native",
        options: [
          {
            version: '9.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.realpathSync.native' is still an experimental feature and is not supported until Node.js 9.2.0. The configured version range is '9.1.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 108',
        code: "require('fs').lutimes",
        options: [
          {
            version: '14.4.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.lutimes' is still an experimental feature and is not supported until Node.js 14.5.0 (backported: ^12.19.0). The configured version range is '14.4.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 109',
        code: "require('fs').lutimesSync",
        options: [
          {
            version: '14.4.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.lutimesSync' is still an experimental feature and is not supported until Node.js 14.5.0 (backported: ^12.19.0). The configured version range is '14.4.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 110',
        code: "require('fs').readv",
        options: [
          {
            version: '13.12.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.readv' is still an experimental feature and is not supported until Node.js 13.13.0 (backported: ^12.17.0). The configured version range is '13.12.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 111',
        code: "require('fs').readvSync",
        options: [
          {
            version: '13.12.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.readvSync' is still an experimental feature and is not supported until Node.js 13.13.0 (backported: ^12.17.0). The configured version range is '13.12.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 24,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 112',
        code: "require('fs').opendir",
        options: [
          {
            version: '12.11.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.opendir' is still an experimental feature and is not supported until Node.js 12.12.0. The configured version range is '12.11.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 113',
        code: "require('fs').opendirSync",
        options: [
          {
            version: '12.11.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.opendirSync' is still an experimental feature and is not supported until Node.js 12.12.0. The configured version range is '12.11.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 114',
        code: "require('fs').rm",
        options: [
          {
            version: '14.13.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '14.13.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 115',
        code: "require('fs').rmSync",
        options: [
          {
            version: '14.13.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.rmSync' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '14.13.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 116',
        code: "require('fs').Dir",
        options: [
          {
            version: '12.11.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.Dir' is still an experimental feature and is not supported until Node.js 12.12.0. The configured version range is '12.11.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 117',
        code: "require('fs').StatWatcher",
        options: [
          {
            version: '14.2.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.StatWatcher' is still an experimental feature and is not supported until Node.js 14.3.0 (backported: ^12.20.0). The configured version range is '14.2.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'fs: upstream invalid 118',
        code: "require('fs').StatWatcher",
        options: [
          {
            version: '12.19.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs.StatWatcher' is still an experimental feature and is not supported until Node.js 14.3.0 (backported: ^12.20.0). The configured version range is '12.19.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'fs/promises: upstream invalid 119',
        code: "import * as fs from 'fs/promises';",
        options: [
          {
            version: '13.14.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs/promises' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '13.14.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 35,
          },
        ],
      },
      {
        name: 'fs/promises: upstream invalid 120',
        code: "require('fs/promises');",
        options: [
          {
            version: '13.14.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs/promises' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '13.14.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'fs/promises: upstream invalid 121',
        code: "const fs = require('fs/promises');",
        options: [
          {
            version: '13.14.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs/promises' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '13.14.0'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'fs/promises: upstream invalid 122',
        code: "require('node:fs/promises');",
        options: [
          {
            version: '13.14.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs/promises' is still an experimental feature and is not supported until Node.js 14.13.1. The configured version range is '13.14.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: 'fs/promises: upstream invalid 123',
        code: "import * as fs from 'node:fs/promises';",
        options: [
          {
            version: '13.14.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fs/promises' is still an experimental feature and is not supported until Node.js 14.13.1. The configured version range is '13.14.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 40,
          },
        ],
      },
      {
        name: 'http2: upstream invalid 124',
        code: "require('http2')",
        options: [
          {
            version: '8.3.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'http2' is still an experimental feature and is not supported until Node.js 10.10.0 (backported: ^8.13.0). The configured version range is '8.3.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'http2: upstream invalid 125',
        code: "import http2 from 'http2'",
        options: [
          {
            version: '8.3.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'http2' is still an experimental feature and is not supported until Node.js 10.10.0 (backported: ^8.13.0). The configured version range is '8.3.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'http2: upstream invalid 126',
        code: "import { createServer } from 'http2'",
        options: [
          {
            version: '8.3.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'http2' is still an experimental feature and is not supported until Node.js 10.10.0 (backported: ^8.13.0). The configured version range is '8.3.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 37,
          },
          {
            messageId: 'not-supported-till',
            message:
              "The 'http2.createServer' is still an experimental feature and is not supported until Node.js 8.4.0. The configured version range is '8.3.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'inspector: upstream invalid 127',
        code: "require('inspector')",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'inspector' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'inspector: upstream invalid 128',
        code: "import inspector from 'inspector'",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'inspector' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'inspector: upstream invalid 129',
        code: "import { open } from 'inspector'",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'inspector' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
          {
            messageId: 'not-supported-till',
            message:
              "The 'inspector.open' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        name: 'module: upstream invalid 130',
        code: 'require.resolve.paths()',
        options: [
          {
            version: '8.8.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'require.resolve.paths' is still an experimental feature and is not supported until Node.js 8.9.0. The configured version range is '8.8.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'module: upstream invalid 131',
        code: "require('module').builtinModules",
        options: [
          {
            version: '9.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'module.builtinModules' is still an experimental feature and is not supported until Node.js 9.3.0 (backported: ^8.10.0, ^6.13.0). The configured version range is '9.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'os: upstream invalid 132',
        code: "require('os').constants",
        options: [
          {
            version: '6.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'os.constants' is still an experimental feature and is not supported until Node.js 6.3.0 (backported: ^5.11.0). The configured version range is '6.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 24,
          },
        ],
      },
      {
        name: 'os: upstream invalid 133',
        code: "var hooks = require('os'); hooks.constants",
        options: [
          {
            version: '6.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'os.constants' is still an experimental feature and is not supported until Node.js 6.3.0 (backported: ^5.11.0). The configured version range is '6.2.9'.",
            line: 1,
            column: 28,
            endLine: 1,
            endColumn: 43,
          },
        ],
      },
      {
        name: 'os: upstream invalid 134',
        code: "var { constants } = require('os'); constants",
        options: [
          {
            version: '6.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'os.constants' is still an experimental feature and is not supported until Node.js 6.3.0 (backported: ^5.11.0). The configured version range is '6.2.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'os: upstream invalid 135',
        code: "import os from 'os'; os.constants",
        options: [
          {
            version: '6.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'os.constants' is still an experimental feature and is not supported until Node.js 6.3.0 (backported: ^5.11.0). The configured version range is '6.2.9'.",
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'os: upstream invalid 136',
        code: "import { constants } from 'os'; constants",
        options: [
          {
            version: '6.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'os.constants' is still an experimental feature and is not supported until Node.js 6.3.0 (backported: ^5.11.0). The configured version range is '6.2.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'os: upstream invalid 137',
        code: "require('os').homedir",
        options: [
          {
            version: '2.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'os.homedir' is still an experimental feature and is not supported until Node.js 2.3.0. The configured version range is '2.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'os: upstream invalid 138',
        code: "require('os').userInfo",
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'os.userInfo' is still an experimental feature and is not supported until Node.js 6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'path: upstream invalid 139',
        code: "require('path').toNamespacedPath()",
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'path.toNamespacedPath' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'path: upstream invalid 140',
        code: "var path = require('path'); path.toNamespacedPath()",
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'path.toNamespacedPath' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 50,
          },
        ],
      },
      {
        name: 'path: upstream invalid 141',
        code: "var { toNamespacedPath } = require('path'); toNamespacedPath()",
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'path.toNamespacedPath' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'path: upstream invalid 142',
        code: "import path from 'path'; path.toNamespacedPath()",
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'path.toNamespacedPath' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 47,
          },
        ],
      },
      {
        name: 'path: upstream invalid 143',
        code: "import { toNamespacedPath } from 'path'; toNamespacedPath()",
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'path.toNamespacedPath' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'perf_hooks: upstream invalid 144',
        code: "require('perf_hooks')",
        options: [
          {
            version: '8.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'perf_hooks' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'perf_hooks: upstream invalid 145',
        code: "import perf_hooks from 'perf_hooks'",
        options: [
          {
            version: '8.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'perf_hooks' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'perf_hooks: upstream invalid 146',
        code: "import { open } from 'perf_hooks'",
        options: [
          {
            version: '8.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'perf_hooks' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'process: upstream invalid 147',
        code: 'process.argv0',
        options: [
          {
            version: '6.3.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.argv0' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        name: 'process: upstream invalid 148',
        code: "require('process').argv0",
        options: [
          {
            version: '6.3.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.argv0' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
      },
      {
        name: 'process: upstream invalid 149',
        code: "var c = require('process'); c.argv0",
        options: [
          {
            version: '6.3.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.argv0' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'process: upstream invalid 150',
        code: "var { argv0 } = require('process'); argv0",
        options: [
          {
            version: '6.3.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.argv0' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 12,
          },
        ],
      },
      {
        name: 'process: upstream invalid 151',
        code: "import c from 'process'; c.argv0",
        options: [
          {
            version: '6.3.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.argv0' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.",
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'process: upstream invalid 152',
        code: 'process.channel',
        options: [
          {
            version: '7.0.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.channel' is still an experimental feature and is not supported until Node.js 7.1.0. The configured version range is '7.0.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'process: upstream invalid 153',
        code: 'process.cpuUsage',
        options: [
          {
            version: '6.0.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.cpuUsage' is still an experimental feature and is not supported until Node.js 6.1.0. The configured version range is '6.0.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'process: upstream invalid 154',
        code: 'process.emitWarning',
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.emitWarning' is still an experimental feature and is not supported until Node.js 6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'process: upstream invalid 155',
        code: 'process.getegid',
        options: [
          {
            version: '1.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.getegid' is still an experimental feature and is not supported until Node.js 2.0.0. The configured version range is '1.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'process: upstream invalid 156',
        code: 'process.geteuid',
        options: [
          {
            version: '1.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.geteuid' is still an experimental feature and is not supported until Node.js 2.0.0. The configured version range is '1.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'process: upstream invalid 157',
        code: 'process.hasUncaughtExceptionCaptureCallback',
        options: [
          {
            version: '9.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.hasUncaughtExceptionCaptureCallback' is still an experimental feature and is not supported until Node.js 9.3.0. The configured version range is '9.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'process: upstream invalid 158',
        code: 'process.ppid',
        options: [
          {
            version: '9.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.ppid' is still an experimental feature and is not supported until Node.js 9.2.0 (backported: ^8.10.0, ^6.13.0). The configured version range is '9.1.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 13,
          },
        ],
      },
      {
        name: 'process: upstream invalid 159',
        code: 'process.release',
        options: [
          {
            version: '2.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.release' is still an experimental feature and is not supported until Node.js 3.0.0. The configured version range is '2.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'process: upstream invalid 160',
        code: 'process.setegid',
        options: [
          {
            version: '1.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.setegid' is still an experimental feature and is not supported until Node.js 2.0.0. The configured version range is '1.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'process: upstream invalid 161',
        code: 'process.seteuid',
        options: [
          {
            version: '1.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.seteuid' is still an experimental feature and is not supported until Node.js 2.0.0. The configured version range is '1.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'process: upstream invalid 162',
        code: 'process.setUncaughtExceptionCaptureCallback',
        options: [
          {
            version: '9.2.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'process.setUncaughtExceptionCaptureCallback' is still an experimental feature and is not supported until Node.js 9.3.0. The configured version range is '9.2.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'stream: upstream invalid 163',
        code: "require('stream').finished()",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'stream.finished' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'stream: upstream invalid 164',
        code: "var hooks = require('stream'); hooks.finished()",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'stream.finished' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 46,
          },
        ],
      },
      {
        name: 'stream: upstream invalid 165',
        code: "var { finished } = require('stream'); finished()",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'stream.finished' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 15,
          },
        ],
      },
      {
        name: 'stream: upstream invalid 166',
        code: "import stream from 'stream'; stream.finished()",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'stream.finished' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'stream: upstream invalid 167',
        code: "import { finished } from 'stream'; finished()",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'stream.finished' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'stream: upstream invalid 168',
        code: "require('stream').pipeline()",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'stream.pipeline' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'trace_events: upstream invalid 169',
        code: "require('trace_events')",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'trace_events' is still an experimental feature The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 24,
          },
        ],
      },
      {
        name: 'trace_events: upstream invalid 170',
        code: "import trace_events from 'trace_events'",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'trace_events' is still an experimental feature The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 40,
          },
        ],
      },
      {
        name: 'trace_events: upstream invalid 171',
        code: "import { createTracing } from 'trace_events'",
        options: [
          {
            version: '10.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'trace_events' is still an experimental feature The configured version range is '10.0.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'url: upstream invalid 172',
        code: 'URL',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'URL' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 4,
          },
        ],
      },
      {
        name: 'url: upstream invalid 173',
        code: 'URLSearchParams',
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'URLSearchParams' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'url: upstream invalid 174',
        code: "require('url').URL",
        options: [
          {
            version: '6.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'url.URL' is still an experimental feature and is not supported until Node.js 7.0.0 (backported: ^6.13.0). The configured version range is '6.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'url: upstream invalid 175',
        code: "var cp = require('url'); cp.URL",
        options: [
          {
            version: '6.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'url.URL' is still an experimental feature and is not supported until Node.js 7.0.0 (backported: ^6.13.0). The configured version range is '6.9.9'.",
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 32,
          },
        ],
      },
      {
        name: 'url: upstream invalid 176',
        code: "var { URL } = require('url');",
        options: [
          {
            version: '6.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'url.URL' is still an experimental feature and is not supported until Node.js 7.0.0 (backported: ^6.13.0). The configured version range is '6.9.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 10,
          },
        ],
      },
      {
        name: 'url: upstream invalid 177',
        code: "import cp from 'url'; cp.URL",
        options: [
          {
            version: '6.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'url.URL' is still an experimental feature and is not supported until Node.js 7.0.0 (backported: ^6.13.0). The configured version range is '6.9.9'.",
            line: 1,
            column: 23,
            endLine: 1,
            endColumn: 29,
          },
        ],
      },
      {
        name: 'url: upstream invalid 178',
        code: "import { URL } from 'url'",
        options: [
          {
            version: '6.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'url.URL' is still an experimental feature and is not supported until Node.js 7.0.0 (backported: ^6.13.0). The configured version range is '6.9.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 13,
          },
        ],
      },
      {
        name: 'url: upstream invalid 179',
        code: "require('url').URLSearchParams",
        options: [
          {
            version: '7.4.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'url.URLSearchParams' is still an experimental feature and is not supported until Node.js 7.5.0 (backported: ^6.13.0). The configured version range is '7.4.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 31,
          },
        ],
      },
      {
        name: 'url: upstream invalid 180',
        code: "require('url').domainToASCII",
        options: [
          {
            version: '7.3.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'url.domainToASCII' is still an experimental feature and is not supported until Node.js 7.4.0 (backported: ^6.13.0). The configured version range is '7.3.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 29,
          },
        ],
      },
      {
        name: 'url: upstream invalid 181',
        code: "require('url').domainToUnicode",
        options: [
          {
            version: '7.3.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'url.domainToUnicode' is still an experimental feature and is not supported until Node.js 7.4.0 (backported: ^6.13.0). The configured version range is '7.3.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 31,
          },
        ],
      },
      {
        name: 'util: upstream invalid 182',
        code: "require('util').callbackify",
        options: [
          {
            version: '8.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.callbackify' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: 'util: upstream invalid 183',
        code: "var hooks = require('util'); hooks.callbackify",
        options: [
          {
            version: '8.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.callbackify' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.",
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 47,
          },
        ],
      },
      {
        name: 'util: upstream invalid 184',
        code: "var { callbackify } = require('util'); callbackify",
        options: [
          {
            version: '8.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.callbackify' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'util: upstream invalid 185',
        code: "import util from 'util'; util.callbackify",
        options: [
          {
            version: '8.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.callbackify' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.",
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 42,
          },
        ],
      },
      {
        name: 'util: upstream invalid 186',
        code: "import { callbackify } from 'util'; callbackify",
        options: [
          {
            version: '8.1.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.callbackify' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'util: upstream invalid 187',
        code: "require('util').formatWithOptions",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.formatWithOptions' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'util: upstream invalid 188',
        code: "require('util').getSystemErrorName",
        options: [
          {
            version: '9.6.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.getSystemErrorName' is still an experimental feature and is not supported until Node.js 9.7.0 (backported: ^8.12.0). The configured version range is '9.6.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 35,
          },
        ],
      },
      {
        name: 'util: upstream invalid 189',
        code: "require('util').inspect.custom",
        options: [
          {
            version: '6.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.inspect.custom' is still an experimental feature and is not supported until Node.js 6.6.0. The configured version range is '6.5.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 31,
          },
        ],
      },
      {
        name: 'util: upstream invalid 190',
        code: "require('util').inspect.defaultOptions",
        options: [
          {
            version: '6.3.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.inspect.defaultOptions' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        name: 'util: upstream invalid 191',
        code: "require('util').isDeepStrictEqual",
        options: [
          {
            version: '8.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.isDeepStrictEqual' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'util: upstream invalid 192',
        code: "require('util').promisify",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.promisify' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'util: upstream invalid 193',
        code: "require('util').TextDecoder",
        options: [
          {
            version: '8.8.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.TextDecoder' is still an experimental feature and is not supported until Node.js 8.9.0. The configured version range is '8.8.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: 'util: upstream invalid 194',
        code: "require('util').TextEncoder",
        options: [
          {
            version: '8.8.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.TextEncoder' is still an experimental feature and is not supported until Node.js 8.9.0. The configured version range is '8.8.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: 'util: upstream invalid 195',
        code: "require('util').types",
        options: [
          {
            version: '9.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.types' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'util: upstream invalid 196',
        code: "require('util').styleText",
        options: [
          {
            version: '21.7.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'util.styleText' is still an experimental feature and is not supported until Node.js 23.5.0 (backported: ^22.13.0). The configured version range is '21.7.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'util: upstream invalid 197',
        code: "require('util').styleText",
        options: [
          {
            version: '20.11.0',
            allowExperimental: true,
          },
        ],
        errors: [
          {
            messageId: 'not-experimental-till',
            message:
              "The 'util.styleText' is not an experimental feature until Node.js 21.7.0 (backported: ^20.12.0). The configured version range is '20.11.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 198',
        code: "require('v8')",
        options: [
          {
            version: '0.12.99',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8' is still an experimental feature and is not supported until Node.js 1.0.0. The configured version range is '0.12.99'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 199',
        code: "import hooks from 'v8'",
        options: [
          {
            version: '0.12.99',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8' is still an experimental feature and is not supported until Node.js 1.0.0. The configured version range is '0.12.99'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 200',
        code: "import { cachedDataVersionTag } from 'v8'",
        options: [
          {
            version: '0.12.99',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8' is still an experimental feature and is not supported until Node.js 1.0.0. The configured version range is '0.12.99'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 42,
          },
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '0.12.99'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 30,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 201',
        code: "require('v8').cachedDataVersionTag()",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 35,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 202',
        code: "var hooks = require('v8'); hooks.cachedDataVersionTag()",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 28,
            endLine: 1,
            endColumn: 54,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 203',
        code: "var { cachedDataVersionTag } = require('v8'); cachedDataVersionTag()",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 204',
        code: "import v8 from 'v8'; v8.cachedDataVersionTag()",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 205',
        code: "import { cachedDataVersionTag } from 'v8'; cachedDataVersionTag()",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 30,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 206',
        code: "require('v8').getHeapSpaceStatistics()",
        options: [
          {
            version: '5.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.getHeapSpaceStatistics' is still an experimental feature and is not supported until Node.js 6.0.0. The configured version range is '5.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 207',
        code: "require('v8').serialize()",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.serialize' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 24,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 208',
        code: "require('v8').deserialize()",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.deserialize' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 209',
        code: "require('v8').Serializer",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.Serializer' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 210',
        code: "require('v8').Deserializer",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.Deserializer' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 211',
        code: "require('v8').DefaultSerializer",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.DefaultSerializer' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
      },
      {
        name: 'v8: upstream invalid 212',
        code: "require('v8').DefaultDeserializer",
        options: [
          {
            version: '7.9.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'v8.DefaultDeserializer' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'vm: upstream invalid 213',
        code: "require('vm').Module",
        options: [
          {
            version: '9.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'vm.Module' is still an experimental feature The configured version range is '9.5.9'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'vm: upstream invalid 214',
        code: "var vm = require('vm'); vm.Module",
        options: [
          {
            version: '9.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'vm.Module' is still an experimental feature The configured version range is '9.5.9'.",
            line: 1,
            column: 25,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'vm: upstream invalid 215',
        code: "var { Module } = require('vm'); Module",
        options: [
          {
            version: '9.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'vm.Module' is still an experimental feature The configured version range is '9.5.9'.",
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 13,
          },
        ],
      },
      {
        name: 'vm: upstream invalid 216',
        code: "import vm from 'vm'; vm.Module",
        options: [
          {
            version: '9.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'vm.Module' is still an experimental feature The configured version range is '9.5.9'.",
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 31,
          },
        ],
      },
      {
        name: 'vm: upstream invalid 217',
        code: "import { Module } from 'vm'; Module",
        options: [
          {
            version: '9.5.9',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'vm.Module' is still an experimental feature The configured version range is '9.5.9'.",
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'worker_threads: upstream invalid 218',
        code: "require('worker_threads')",
        options: [
          {
            version: '10.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'worker_threads' is still an experimental feature and is not supported until Node.js 12.11.0. The configured version range is '10.5.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'worker_threads: upstream invalid 219',
        code: "import worker_threads from 'worker_threads'",
        options: [
          {
            version: '10.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'worker_threads' is still an experimental feature and is not supported until Node.js 12.11.0. The configured version range is '10.5.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'worker_threads: upstream invalid 220',
        code: "import { Worker } from 'worker_threads'",
        options: [
          {
            version: '10.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'worker_threads' is still an experimental feature and is not supported until Node.js 12.11.0. The configured version range is '10.5.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 40,
          },
        ],
      },
      {
        name: 'worker_threads: upstream invalid 221',
        code: "import { Worker } from 'worker_threads'",
        settings: {
          node: {
            version: '10.5.0',
          },
        },
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'worker_threads' is still an experimental feature and is not supported until Node.js 12.11.0. The configured version range is '10.5.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 40,
          },
        ],
      },
      {
        name: 'timers/promises: upstream invalid 222',
        code: "\n                        import { scheduler } from 'node:timers/promises';\n                        await scheduler.wait( 1000 );\n                    ",
        options: [
          {
            version: '>= 20.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'timers/promises.scheduler.wait' is still an experimental feature The configured version range is '>= 20.0.0'.",
            line: 3,
            column: 31,
            endLine: 3,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 223',
        code: 'import.meta.resolve(specifier)',
        options: [
          {
            version: '20.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'import.meta.resolve' is still an experimental feature and is not supported until Node.js 20.6.0 (backported: ^18.19.0). The configured version range is '20.5.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 224',
        code: 'import.meta.resolve(specifier)',
        options: [
          {
            version: '19.8.1',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'import.meta.resolve' is still an experimental feature and is not supported until Node.js 20.6.0 (backported: ^18.19.0). The configured version range is '19.8.1'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 225',
        code: 'import.meta.resolve(specifier)',
        options: [
          {
            version: '18.18.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'import.meta.resolve' is still an experimental feature and is not supported until Node.js 20.6.0 (backported: ^18.19.0). The configured version range is '18.18.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 226',
        code: 'import.meta.resolve(specifier)',
        options: [
          {
            version: '13.8.0',
            allowExperimental: true,
          },
        ],
        errors: [
          {
            messageId: 'not-experimental-till',
            message:
              "The 'import.meta.resolve' is not an experimental feature until Node.js 13.9.0 (backported: ^12.16.2). The configured version range is '13.8.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 227',
        code: 'import.meta.resolve(specifier)',
        options: [
          {
            version: '12.15.0',
            allowExperimental: true,
          },
        ],
        errors: [
          {
            messageId: 'not-experimental-till',
            message:
              "The 'import.meta.resolve' is not an experimental feature until Node.js 13.9.0 (backported: ^12.16.2). The configured version range is '12.15.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 228',
        code: 'import.meta.dirname;',
        options: [
          {
            version: '22.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'import.meta.dirname' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '22.0.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 229',
        code: 'import.meta.dirname;',
        options: [
          {
            version: '21.1.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'import.meta.dirname' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '21.1.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 230',
        code: 'import.meta.dirname;',
        options: [
          {
            version: '20.10.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'import.meta.dirname' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '20.10.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 231',
        code: 'import.meta.dirname;',
        options: [
          {
            version: '21.1.0',
            allowExperimental: true,
          },
        ],
        errors: [
          {
            messageId: 'not-experimental-till',
            message:
              "The 'import.meta.dirname' is not an experimental feature until Node.js 21.2.0 (backported: ^20.11.0). The configured version range is '21.1.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 232',
        code: 'import.meta.dirname;',
        options: [
          {
            version: '20.10.0',
            allowExperimental: true,
          },
        ],
        errors: [
          {
            messageId: 'not-experimental-till',
            message:
              "The 'import.meta.dirname' is not an experimental feature until Node.js 21.2.0 (backported: ^20.11.0). The configured version range is '20.10.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 233',
        code: 'import.meta.filename;',
        options: [
          {
            version: '22.0.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'import.meta.filename' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '22.0.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 234',
        code: 'import.meta.filename;',
        options: [
          {
            version: '21.1.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'import.meta.filename' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '21.1.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 235',
        code: 'import.meta.filename;',
        options: [
          {
            version: '20.10.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'import.meta.filename' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '20.10.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 236',
        code: 'import.meta.filename;',
        options: [
          {
            version: '21.1.0',
            allowExperimental: true,
          },
        ],
        errors: [
          {
            messageId: 'not-experimental-till',
            message:
              "The 'import.meta.filename' is not an experimental feature until Node.js 21.2.0 (backported: ^20.11.0). The configured version range is '21.1.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'import.meta: upstream invalid 237',
        code: 'import.meta.filename;',
        options: [
          {
            version: '20.10.0',
            allowExperimental: true,
          },
        ],
        errors: [
          {
            messageId: 'not-experimental-till',
            message:
              "The 'import.meta.filename' is not an experimental feature until Node.js 21.2.0 (backported: ^20.11.0). The configured version range is '20.10.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'fetch: upstream invalid 238',
        code: "fetch('/asd')",
        options: [
          {
            version: '16.0.0',
            allowExperimental: true,
          },
        ],
        errors: [
          {
            messageId: 'not-experimental-till',
            message:
              "The 'fetch' is not an experimental feature until Node.js 17.5.0 (backported: ^16.15.0). The configured version range is '16.0.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 6,
          },
        ],
      },
      {
        name: 'fetch: upstream invalid 239',
        code: "fetch('/asd')",
        options: [
          {
            version: '16.16.0',
            allowExperimental: false,
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '16.16.0'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 6,
          },
        ],
      },
      {
        name: 'sqlite: upstream invalid 240',
        code: "\n                        import { DatabaseSync } from 'node:sqlite';\n                        const database = new DatabaseSync(':memory:');\n                    ",
        options: [
          {
            version: '>=22.5.0',
          },
        ],
        errors: [
          {
            messageId: 'not-supported-yet',
            message:
              "The 'sqlite' is still an experimental feature The configured version range is '>=22.5.0'.",
            line: 2,
            column: 25,
            endLine: 2,
            endColumn: 68,
          },
        ],
      },
      {
        name: 'sqlite: upstream invalid 241',
        code: "\n                        import { DatabaseSync } from 'node:sqlite';\n                        const database = new DatabaseSync(':memory:');\n                    ",
        options: [
          {
            version: '>=22.3.0',
            allowExperimental: true,
          },
        ],
        errors: [
          {
            messageId: 'not-experimental-till',
            message:
              "The 'sqlite' is not an experimental feature until Node.js 22.5.0. The configured version range is '>=22.3.0'.",
            line: 2,
            column: 25,
            endLine: 2,
            endColumn: 68,
          },
          {
            messageId: 'not-supported-till',
            message:
              "The 'sqlite.DatabaseSync' is still an experimental feature and is not supported until Node.js 22.5.0. The configured version range is '>=22.3.0'.",
            line: 2,
            column: 34,
            endLine: 2,
            endColumn: 46,
          },
        ],
      },
      {
        name: 'sqlite: upstream invalid 242',
        code: "\n                        const { DatabaseSync } = require('node:sqlite');\n                        const database = new DatabaseSync(':memory:');\n                    ",
        options: [
          {
            version: '>=22.3.0',
            allowExperimental: true,
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'sqlite.DatabaseSync' is still an experimental feature and is not supported until Node.js 22.5.0. The configured version range is '>=22.3.0'.",
            line: 2,
            column: 33,
            endLine: 2,
            endColumn: 45,
          },
          {
            messageId: 'not-experimental-till',
            message:
              "The 'sqlite' is not an experimental feature until Node.js 22.5.0. The configured version range is '>=22.3.0'.",
            line: 2,
            column: 50,
            endLine: 2,
            endColumn: 72,
          },
        ],
      },
      {
        name: 'sqlite: upstream invalid 243',
        code: "\n                        const { DatabaseSync } = process.getBuiltinModule('node:sqlite');\n                        const database = new DatabaseSync(':memory:');\n                    ",
        options: [
          {
            version: '>=22.3.0',
            allowExperimental: true,
          },
        ],
        errors: [
          {
            messageId: 'not-supported-till',
            message:
              "The 'sqlite.DatabaseSync' is still an experimental feature and is not supported until Node.js 22.5.0. The configured version range is '>=22.3.0'.",
            line: 2,
            column: 33,
            endLine: 2,
            endColumn: 45,
          },
          {
            messageId: 'not-experimental-till',
            message:
              "The 'sqlite' is not an experimental feature until Node.js 22.5.0. The configured version range is '>=22.3.0'.",
            line: 2,
            column: 50,
            endLine: 2,
            endColumn: 89,
          },
        ],
      },
    ],
  },
);
