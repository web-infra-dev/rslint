// Upstream cases and documentation examples, eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-deprecated-api.js
import { RuleTester } from '../rule-tester';

new RuleTester({
  languageOptions: {
    sourceType: 'commonjs',
    globals: {
      Buffer: 'readonly',
      process: 'readonly',
      require: 'readonly',
      global: 'readonly',
      Intl: 'readonly',
    },
  },
  fixtureFiles: { 'package.json': '{}' },
}).run(
  'no-deprecated-api',
  {},
  {
    valid: [
      {
        code: "require('buffer').Buffer",
        name: 'Upstream valid 1',
      },
      {
        code: "require('node:buffer').Buffer",
        name: 'Upstream valid 2',
      },
      {
        code: "foo(require('buffer').Buffer)",
        name: 'Upstream valid 3',
      },
      {
        code: "new (require('another-buffer').Buffer)()",
        name: 'Upstream valid 4',
      },
      {
        code: "var http = require('http'); http.request()",
        name: 'Upstream valid 5',
      },
      {
        code: "var {request} = require('http'); request()",
        name: 'Upstream valid 6',
      },
      {
        code: "(s ? require('https') : require('http')).request()",
        name: 'Upstream valid 7',
      },
      {
        code: 'require(HTTP).createClient',
        name: 'Upstream valid 8',
      },
      {
        code: "import {Buffer} from 'another-buffer'; new Buffer()",
        languageOptions: {
          sourceType: 'module',
        },
        name: 'Upstream valid 9',
      },
      {
        code: "import {request} from 'http'; request()",
        languageOptions: {
          sourceType: 'module',
        },
        name: 'Upstream valid 10',
      },
      {
        code: "const {Buffer} = process.getBuiltinModule('another-buffer'); new Buffer()",
        name: 'Upstream valid 11',
      },
      {
        code: "const {request} = process.getBuiltinModule('http'); request()",
        name: 'Upstream valid 12',
      },
      {
        code: "require('fs').existsSync;",
        name: 'Upstream valid 13',
      },
      {
        code: "require('domain/');",
        name: 'Upstream valid 14',
      },
      {
        code: "import domain from 'domain/';",
        languageOptions: {
          sourceType: 'module',
        },
        name: 'Upstream valid 15',
      },
      {
        code: "undefinedVar = require('fs')",
        name: 'Upstream valid 16',
      },
      {
        code: "new (require('buffer').Buffer)()",
        options: [
          {
            ignoreModuleItems: ['new buffer.Buffer()'],
          },
        ],
        name: 'Upstream valid 17',
      },
      {
        code: "require('buffer').Buffer()",
        options: [
          {
            ignoreModuleItems: ['buffer.Buffer()'],
          },
        ],
        name: 'Upstream valid 18',
      },
      {
        code: "require('node:buffer').Buffer()",
        options: [
          {
            ignoreModuleItems: ['buffer.Buffer()'],
          },
        ],
        name: 'Upstream valid 19',
      },
      {
        code: "require('domain');",
        options: [
          {
            ignoreModuleItems: ['domain'],
          },
        ],
        name: 'Upstream valid 20',
      },
      {
        code: "require('events').EventEmitter.listenerCount;",
        options: [
          {
            ignoreModuleItems: ['events.EventEmitter.listenerCount'],
          },
        ],
        name: 'Upstream valid 21',
      },
      {
        code: "require('events').listenerCount;",
        options: [
          {
            ignoreModuleItems: ['events.listenerCount'],
          },
        ],
        name: 'Upstream valid 22',
      },
      {
        code: 'new Buffer;',
        options: [
          {
            ignoreGlobalItems: ['new Buffer()'],
          },
        ],
        name: 'Upstream valid 23',
      },
      {
        code: 'Buffer();',
        options: [
          {
            ignoreGlobalItems: ['Buffer()'],
          },
        ],
        name: 'Upstream valid 24',
      },
      {
        code: 'Intl.v8BreakIterator;',
        options: [
          {
            ignoreGlobalItems: ['Intl.v8BreakIterator'],
          },
        ],
        name: 'Upstream valid 25',
      },
      {
        code: 'let {env: {NODE_REPL_HISTORY_FILE}} = process;',
        options: [
          {
            ignoreGlobalItems: ['process.env.NODE_REPL_HISTORY_FILE'],
          },
        ],
        name: 'Upstream valid 26',
      },
      {
        code: 'require("domain/")',
        options: [
          {
            ignoreIndirectDependencies: true,
          },
        ],
        name: 'Upstream valid 27',
      },
      {
        code: 'let fs = fs || require("fs")',
        name: 'Upstream valid 28',
      },
      {
        code: 'const buffer = require("buffer"); const data = new buffer.Buffer(10);',
        options: [
          {
            ignoreModuleItems: ['new buffer.Buffer()'],
          },
        ],
        name: 'Documentation example 2',
      },
      {
        code: 'const data = new Buffer(10);',
        options: [
          {
            ignoreGlobalItems: ['new Buffer()'],
          },
        ],
        name: 'Documentation example 3',
      },
      {
        code: 'require(foo).aDeprecatedProperty;\nrequire("http")[A_DEPRECATED_PROPERTY]();',
        name: 'Documentation example 4',
      },
      {
        code: 'var obj = {Buffer: require("buffer").Buffer}; new obj.Buffer();',
        name: 'Documentation example 5',
      },
      {
        code: 'var obj = {}; obj.Buffer = require("buffer").Buffer; new obj.Buffer();',
        name: 'Documentation example 6',
      },
      {
        code: '(function(Buffer) { new Buffer(); })(require("buffer").Buffer);',
        name: 'Documentation example 7',
      },
    ],
    invalid: [
      {
        code: "new (require('buffer').Buffer)()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
        ],
        name: 'Upstream invalid 1',
      },
      {
        code: "new (require('node:buffer').Buffer)()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 38,
          },
        ],
        name: 'Upstream invalid 2',
      },
      {
        code: "require('buffer').Buffer()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 27,
          },
        ],
        name: 'Upstream invalid 3',
      },
      {
        code: "require('node:buffer').Buffer()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
        name: 'Upstream invalid 4',
      },
      {
        code: "var b = require('buffer'); new b.Buffer()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 28,
            endLine: 1,
            endColumn: 42,
          },
        ],
        name: 'Upstream invalid 5',
      },
      {
        code: "var b = require('buffer'); new b['Buffer']()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 28,
            endLine: 1,
            endColumn: 45,
          },
        ],
        name: 'Upstream invalid 6',
      },
      {
        code: "var b = require('buffer'); new b[`Buffer`]()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 28,
            endLine: 1,
            endColumn: 45,
          },
        ],
        name: 'Upstream invalid 7',
      },
      {
        code: "var b = require('buffer').Buffer; new b()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 35,
            endLine: 1,
            endColumn: 42,
          },
        ],
        name: 'Upstream invalid 8',
      },
      {
        code: "var b; new ((b = require('buffer')).Buffer)(); new b.Buffer()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 46,
          },
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 48,
            endLine: 1,
            endColumn: 62,
          },
        ],
        name: 'Upstream invalid 9',
      },
      {
        code: "var {Buffer: b} = require('buffer'); new b()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 38,
            endLine: 1,
            endColumn: 45,
          },
        ],
        name: 'Upstream invalid 10',
      },
      {
        code: "var {['Buffer']: b = null} = require('buffer'); new b()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 49,
            endLine: 1,
            endColumn: 56,
          },
        ],
        name: 'Upstream invalid 11',
      },
      {
        code: "var {'Buffer': b = null} = require('buffer'); new b()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 47,
            endLine: 1,
            endColumn: 54,
          },
        ],
        name: 'Upstream invalid 12',
      },
      {
        code: "var {Buffer: b = require('buffer').Buffer} = {}; new b()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 50,
            endLine: 1,
            endColumn: 57,
          },
        ],
        name: 'Upstream invalid 13',
      },
      {
        code: "require('buffer').SlowBuffer",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 29,
          },
        ],
        name: 'Upstream invalid 14',
      },
      {
        code: "require('node:buffer').SlowBuffer",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
        name: 'Upstream invalid 15',
      },
      {
        code: "var b = require('buffer'); b.SlowBuffer",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.",
            line: 1,
            column: 28,
            endLine: 1,
            endColumn: 40,
          },
        ],
        name: 'Upstream invalid 16',
      },
      {
        code: "var {SlowBuffer: b} = require('buffer');",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.",
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 19,
          },
        ],
        name: 'Upstream invalid 17',
      },
      {
        code: "require('_linklist');",
        options: [
          {
            version: '5.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'_linklist' module was deprecated since v5.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
        name: 'Upstream invalid 18',
      },
      {
        code: "require('async_hooks').currentId;",
        options: [
          {
            version: '8.2.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'async_hooks.currentId' was deprecated since v8.2.0. Use 'async_hooks.executionAsyncId()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
        ],
        name: 'Upstream invalid 19',
      },
      {
        code: "require('async_hooks').triggerId;",
        options: [
          {
            version: '8.2.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'async_hooks.triggerId' was deprecated since v8.2.0. Use 'async_hooks.triggerAsyncId()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
        ],
        name: 'Upstream invalid 20',
      },
      {
        code: "require('constants');",
        options: [
          {
            version: '6.3.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'constants' module was deprecated since v6.3.0. Use 'constants' property of each module instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
        name: 'Upstream invalid 21',
      },
      {
        code: "require('crypto').Credentials;",
        options: [
          {
            version: '0.12.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'crypto.Credentials' was deprecated since v0.12.0. Use 'tls.SecureContext' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 30,
          },
        ],
        name: 'Upstream invalid 22',
      },
      {
        code: "require('crypto').createCredentials;",
        options: [
          {
            version: '0.12.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'crypto.createCredentials' was deprecated since v0.12.0. Use 'tls.createSecureContext()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 36,
          },
        ],
        name: 'Upstream invalid 23',
      },
      {
        code: "require('domain');",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'domain' module was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 18,
          },
        ],
        name: 'Upstream invalid 24',
      },
      {
        code: "require('events').EventEmitter.listenerCount;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'events.EventEmitter.listenerCount' was deprecated since v4.0.0. Use 'events.EventEmitter#listenerCount()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 45,
          },
        ],
        name: 'Upstream invalid 25',
      },
      {
        code: "require('events').listenerCount;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'events.listenerCount' was deprecated since v4.0.0. Use 'events.EventEmitter#listenerCount()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
        name: 'Upstream invalid 26',
      },
      {
        code: "require('freelist');",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'freelist' module was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
        name: 'Upstream invalid 27',
      },
      {
        code: "require('fs').SyncWriteStream;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'fs.SyncWriteStream' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 30,
          },
        ],
        name: 'Upstream invalid 28',
      },
      {
        code: "require('fs').exists;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
        name: 'Upstream invalid 29',
      },
      {
        code: "require('fs').lchmod;",
        options: [
          {
            version: '0.4.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'fs.lchmod' was deprecated since v0.4.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
        name: 'Upstream invalid 30',
      },
      {
        code: "require('fs').lchmodSync;",
        options: [
          {
            version: '0.4.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'fs.lchmodSync' was deprecated since v0.4.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
        name: 'Upstream invalid 31',
      },
      {
        code: "require('http').createClient;",
        options: [
          {
            version: '0.10.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'http.createClient' was deprecated since v0.10.0. Use 'http.request()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 29,
          },
        ],
        name: 'Upstream invalid 32',
      },
      {
        code: "require('module').requireRepl;",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'module.requireRepl' was deprecated since v6.0.0. Use 'require(\"repl\")' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 30,
          },
        ],
        name: 'Upstream invalid 33',
      },
      {
        code: "require('module').Module.requireRepl;",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'module.Module.requireRepl' was deprecated since v6.0.0. Use 'require(\"repl\")' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 37,
          },
        ],
        name: 'Upstream invalid 34',
      },
      {
        code: "require('module')._debug;",
        options: [
          {
            version: '9.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'module._debug' was deprecated since v9.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
        name: 'Upstream invalid 35',
      },
      {
        code: "require('module').Module._debug;",
        options: [
          {
            version: '9.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'module.Module._debug' was deprecated since v9.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
        name: 'Upstream invalid 36',
      },
      {
        code: "require('os').getNetworkInterfaces;",
        options: [
          {
            version: '0.6.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'os.getNetworkInterfaces' was deprecated since v0.6.0. Use 'os.networkInterfaces()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 35,
          },
        ],
        name: 'Upstream invalid 37',
      },
      {
        code: "require('os').tmpDir;",
        options: [
          {
            version: '7.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'os.tmpDir' was deprecated since v7.0.0. Use 'os.tmpdir()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
        name: 'Upstream invalid 38',
      },
      {
        code: "require('path')._makeLong;",
        options: [
          {
            version: '9.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'path._makeLong' was deprecated since v9.0.0. Use 'path.toNamespacedPath()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
        name: 'Upstream invalid 39',
      },
      {
        code: "require('punycode');",
        options: [
          {
            version: '7.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'punycode' module was deprecated since v7.0.0. Use 'https://www.npmjs.com/package/punycode' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
        name: 'Upstream invalid 40',
      },
      {
        code: "require('readline').codePointAt;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'readline.codePointAt' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
        name: 'Upstream invalid 41',
      },
      {
        code: "require('readline').getStringWidth;",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'readline.getStringWidth' was deprecated since v6.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 35,
          },
        ],
        name: 'Upstream invalid 42',
      },
      {
        code: "require('readline').isFullWidthCodePoint;",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'readline.isFullWidthCodePoint' was deprecated since v6.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 41,
          },
        ],
        name: 'Upstream invalid 43',
      },
      {
        code: "require('readline').stripVTControlCharacters;",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'readline.stripVTControlCharacters' was deprecated since v6.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 45,
          },
        ],
        name: 'Upstream invalid 44',
      },
      {
        code: "require('sys');",
        options: [
          {
            version: '0.3.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'sys' module was deprecated since v0.3.0. Use 'util' module instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 15,
          },
        ],
        name: 'Upstream invalid 45',
      },
      {
        code: "require('tls').CleartextStream;",
        options: [
          {
            version: '0.10.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'tls.CleartextStream' was deprecated since v0.10.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 31,
          },
        ],
        name: 'Upstream invalid 46',
      },
      {
        code: "require('tls').CryptoStream;",
        options: [
          {
            version: '0.12.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'tls.CryptoStream' was deprecated since v0.12.0. Use 'tls.TLSSocket' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
        name: 'Upstream invalid 47',
      },
      {
        code: "require('tls').SecurePair;",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'tls.SecurePair' was deprecated since v6.0.0. Use 'tls.TLSSocket' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
        name: 'Upstream invalid 48',
      },
      {
        code: "require('tls').createSecurePair;",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'tls.createSecurePair' was deprecated since v6.0.0. Use 'tls.TLSSocket' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
        name: 'Upstream invalid 49',
      },
      {
        code: "require('tls').parseCertString;",
        options: [
          {
            version: '8.6.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'tls.parseCertString' was deprecated since v8.6.0. Use 'querystring.parse()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 31,
          },
        ],
        name: 'Upstream invalid 50',
      },
      {
        code: "require('tty').setRawMode;",
        options: [
          {
            version: '0.10.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'tty.setRawMode' was deprecated since v0.10.0. Use 'tty.ReadStream#setRawMode()' (e.g. 'process.stdin.setRawMode()') instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
        name: 'Upstream invalid 51',
      },
      {
        code: "require('util').debug;",
        options: [
          {
            version: '0.12.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'util.debug' was deprecated since v0.12.0. Use 'console.error()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'Upstream invalid 52',
      },
      {
        code: "require('util').error;",
        options: [
          {
            version: '0.12.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'util.error' was deprecated since v0.12.0. Use 'console.error()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'Upstream invalid 53',
      },
      {
        code: "require('util').isArray;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'util.isArray' was deprecated since v4.0.0. Use 'Array.isArray()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 24,
          },
        ],
        name: 'Upstream invalid 54',
      },
      {
        code: "require('util').isBoolean;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isBoolean' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 26,
          },
        ],
        name: 'Upstream invalid 55',
      },
      {
        code: "require('util').isBuffer;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'util.isBuffer' was deprecated since v4.0.0. Use 'Buffer.isBuffer()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
        name: 'Upstream invalid 56',
      },
      {
        code: "require('util').isDate;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isDate' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
        name: 'Upstream invalid 57',
      },
      {
        code: "require('util').isError;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isError' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 24,
          },
        ],
        name: 'Upstream invalid 58',
      },
      {
        code: "require('util').isFunction;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isFunction' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 27,
          },
        ],
        name: 'Upstream invalid 59',
      },
      {
        code: "require('util').isNull;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isNull' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
        name: 'Upstream invalid 60',
      },
      {
        code: "require('util').isNullOrUndefined;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isNullOrUndefined' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 34,
          },
        ],
        name: 'Upstream invalid 61',
      },
      {
        code: "require('util').isNumber;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isNumber' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
        name: 'Upstream invalid 62',
      },
      {
        code: "require('util').isObject;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isObject' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
        name: 'Upstream invalid 63',
      },
      {
        code: "require('util').isPrimitive;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isPrimitive' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
        name: 'Upstream invalid 64',
      },
      {
        code: "require('util').isRegExp;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isRegExp' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
        name: 'Upstream invalid 65',
      },
      {
        code: "require('util').isString;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isString' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
        name: 'Upstream invalid 66',
      },
      {
        code: "require('util').isSymbol;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isSymbol' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
        name: 'Upstream invalid 67',
      },
      {
        code: "require('util').isUndefined;",
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'util.isUndefined' was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 28,
          },
        ],
        name: 'Upstream invalid 68',
      },
      {
        code: "require('util').log;",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'util.log' was deprecated since v6.0.0. Use a third party module instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 20,
          },
        ],
        name: 'Upstream invalid 69',
      },
      {
        code: "require('util').print;",
        options: [
          {
            version: '0.12.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'util.print' was deprecated since v0.12.0. Use 'console.log()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'Upstream invalid 70',
      },
      {
        code: "require('util').pump;",
        options: [
          {
            version: '0.10.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'util.pump' was deprecated since v0.10.0. Use 'stream.Readable#pipe()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
        name: 'Upstream invalid 71',
      },
      {
        code: "require('util').puts;",
        options: [
          {
            version: '0.12.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'util.puts' was deprecated since v0.12.0. Use 'console.log()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
        name: 'Upstream invalid 72',
      },
      {
        code: "require('util')._extend;",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'util._extend' was deprecated since v6.0.0. Use 'Object.assign()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 24,
          },
        ],
        name: 'Upstream invalid 73',
      },
      {
        code: "require('vm').runInDebugContext;",
        options: [
          {
            version: '8.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message: "'vm.runInDebugContext' was deprecated since v8.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
        name: 'Upstream invalid 74',
      },
      {
        code: "import b from 'buffer'; new b.Buffer()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        languageOptions: {
          sourceType: 'module',
        },
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 25,
            endLine: 1,
            endColumn: 39,
          },
        ],
        name: 'Upstream invalid 75',
      },
      {
        code: "import b from 'node:buffer'; new b.Buffer()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        languageOptions: {
          sourceType: 'module',
        },
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 44,
          },
        ],
        name: 'Upstream invalid 76',
      },
      {
        code: "import * as b from 'buffer'; new b.Buffer()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        languageOptions: {
          sourceType: 'module',
        },
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 44,
          },
        ],
        name: 'Upstream invalid 77',
      },
      {
        code: "import * as b from 'buffer'; new b.default.Buffer()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        languageOptions: {
          sourceType: 'module',
        },
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 52,
          },
        ],
        name: 'Upstream invalid 78',
      },
      {
        code: "import {Buffer as b} from 'buffer'; new b()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        languageOptions: {
          sourceType: 'module',
        },
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 37,
            endLine: 1,
            endColumn: 44,
          },
        ],
        name: 'Upstream invalid 79',
      },
      {
        code: "import b from 'buffer'; b.SlowBuffer",
        options: [
          {
            version: '6.0.0',
          },
        ],
        languageOptions: {
          sourceType: 'module',
        },
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.",
            line: 1,
            column: 25,
            endLine: 1,
            endColumn: 37,
          },
        ],
        name: 'Upstream invalid 80',
      },
      {
        code: "import * as b from 'buffer'; b.SlowBuffer",
        options: [
          {
            version: '6.0.0',
          },
        ],
        languageOptions: {
          sourceType: 'module',
        },
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.",
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 42,
          },
        ],
        name: 'Upstream invalid 81',
      },
      {
        code: "import * as b from 'buffer'; b.default.SlowBuffer",
        options: [
          {
            version: '6.0.0',
          },
        ],
        languageOptions: {
          sourceType: 'module',
        },
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.",
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 50,
          },
        ],
        name: 'Upstream invalid 82',
      },
      {
        code: "import {SlowBuffer as b} from 'buffer';",
        options: [
          {
            version: '6.0.0',
          },
        ],
        languageOptions: {
          sourceType: 'module',
        },
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 24,
          },
        ],
        name: 'Upstream invalid 83',
      },
      {
        code: "import domain from 'domain';",
        options: [
          {
            version: '4.0.0',
          },
        ],
        languageOptions: {
          sourceType: 'module',
        },
        errors: [
          {
            messageId: 'deprecated',
            message: "'domain' module was deprecated since v4.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 29,
          },
        ],
        name: 'Upstream invalid 84',
      },
      {
        code: "new (require('buffer').Buffer)()",
        options: [
          {
            ignoreModuleItems: ['buffer.Buffer()'],
            ignoreGlobalItems: ['Buffer()', 'new Buffer()'],
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 33,
          },
        ],
        name: 'Upstream invalid 85',
      },
      {
        code: "require('buffer').Buffer()",
        options: [
          {
            ignoreModuleItems: ['new buffer.Buffer()'],
            ignoreGlobalItems: ['Buffer()', 'new Buffer()'],
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 27,
          },
        ],
        name: 'Upstream invalid 86',
      },
      {
        code: "require('module').createRequireFromPath()",
        options: [
          {
            version: '12.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'module.createRequireFromPath' was deprecated since v12.2.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 40,
          },
        ],
        name: 'Upstream invalid 87',
      },
      {
        code: "require('module').createRequireFromPath()",
        options: [
          {
            version: '12.2.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'module.createRequireFromPath' was deprecated since v12.2.0. Use 'module.createRequire()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 40,
          },
        ],
        name: 'Upstream invalid 88',
      },
      {
        code: "const b = process.getBuiltinModule('buffer'); new b.Buffer()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 47,
            endLine: 1,
            endColumn: 61,
          },
        ],
        name: 'Upstream invalid 89',
      },
      {
        code: "const b = process.getBuiltinModule('node:buffer'); new b.Buffer()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 52,
            endLine: 1,
            endColumn: 66,
          },
        ],
        name: 'Upstream invalid 90',
      },
      {
        code: "const {Buffer} = process.getBuiltinModule('buffer'); new Buffer()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 54,
            endLine: 1,
            endColumn: 66,
          },
        ],
        name: 'Upstream invalid 91',
      },
      {
        code: "const {Buffer:b} = process.getBuiltinModule('buffer'); new b()",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 56,
            endLine: 1,
            endColumn: 63,
          },
        ],
        name: 'Upstream invalid 92',
      },
      {
        code: "const b = process.getBuiltinModule('buffer'); b.SlowBuffer",
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.",
            line: 1,
            column: 47,
            endLine: 1,
            endColumn: 59,
          },
        ],
        name: 'Upstream invalid 93',
      },
      {
        code: "const domain = process.getBuiltinModule('domain');",
        options: [
          {
            version: '4.0.0',
          },
        ],
        languageOptions: {
          sourceType: 'module',
        },
        errors: [
          {
            messageId: 'deprecated',
            message: "'domain' module was deprecated since v4.0.0.",
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 50,
          },
        ],
        name: 'Upstream invalid 94',
      },
      {
        code: "new (process.getBuiltinModule('buffer').Buffer)()",
        options: [
          {
            ignoreModuleItems: ['buffer.Buffer()'],
            ignoreGlobalItems: ['Buffer()', 'new Buffer()'],
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 50,
          },
        ],
        name: 'Upstream invalid 95',
      },
      {
        code: "process.getBuiltinModule('buffer').Buffer()",
        options: [
          {
            ignoreModuleItems: ['new buffer.Buffer()'],
            ignoreGlobalItems: ['Buffer()', 'new Buffer()'],
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 44,
          },
        ],
        name: 'Upstream invalid 96',
      },
      {
        code: "process.getBuiltinModule('module').createRequireFromPath()",
        options: [
          {
            version: '12.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'module.createRequireFromPath' was deprecated since v12.2.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 57,
          },
        ],
        name: 'Upstream invalid 97',
      },
      {
        code: "process.getBuiltinModule('module').createRequireFromPath()",
        options: [
          {
            version: '12.2.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'module.createRequireFromPath' was deprecated since v12.2.0. Use 'module.createRequire()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 57,
          },
        ],
        name: 'Upstream invalid 98',
      },
      {
        code: 'new Buffer;',
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 11,
          },
        ],
        name: 'Upstream invalid 99',
      },
      {
        code: 'Buffer();',
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 9,
          },
        ],
        name: 'Upstream invalid 100',
      },
      {
        code: 'GLOBAL; /*globals GLOBAL*/',
        options: [
          {
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'GLOBAL' was deprecated since v6.0.0. Use 'global' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
          },
        ],
        name: 'Upstream invalid 101',
      },
      {
        code: 'Intl.v8BreakIterator;',
        options: [
          {
            version: '7.0.0',
          },
        ],
        errors: [
          {
            messageId: 'removed',
            message:
              "'Intl.v8BreakIterator' was deprecated since v7.0.0, and removed in v9.0.0.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
        name: 'Upstream invalid 102',
      },
      {
        code: 'require.extensions;',
        options: [
          {
            version: '0.12.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'require.extensions' was deprecated since v0.12.0. Use compiling them ahead of time instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 19,
          },
        ],
        name: 'Upstream invalid 103',
      },
      {
        code: 'root;',
        options: [
          {
            version: '6.0.0',
          },
        ],
        languageOptions: {
          globals: {
            root: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'root' was deprecated since v6.0.0. Use 'global' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 5,
          },
        ],
        name: 'Upstream invalid 104',
      },
      {
        code: 'process.EventEmitter;',
        options: [
          {
            version: '0.6.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'process.EventEmitter' was deprecated since v0.6.0. Use 'require(\"events\")' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 21,
          },
        ],
        name: 'Upstream invalid 105',
      },
      {
        code: 'process.env.NODE_REPL_HISTORY_FILE;',
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'process.env.NODE_REPL_HISTORY_FILE' was deprecated since v4.0.0. Use 'NODE_REPL_HISTORY' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 35,
          },
        ],
        name: 'Upstream invalid 106',
      },
      {
        code: 'let {env: {NODE_REPL_HISTORY_FILE}} = process;',
        options: [
          {
            version: '4.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'process.env.NODE_REPL_HISTORY_FILE' was deprecated since v4.0.0. Use 'NODE_REPL_HISTORY' instead.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 34,
          },
        ],
        name: 'Upstream invalid 107',
      },
      {
        code: 'new Buffer()',
        options: [
          {
            ignoreModuleItems: ['buffer.Buffer()', 'new buffer.Buffer()'],
            ignoreGlobalItems: ['Buffer()'],
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 13,
          },
        ],
        name: 'Upstream invalid 108',
      },
      {
        code: 'Buffer()',
        options: [
          {
            ignoreModuleItems: ['buffer.Buffer()', 'new buffer.Buffer()'],
            ignoreGlobalItems: ['new Buffer()'],
            version: '6.0.0',
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 9,
          },
        ],
        name: 'Upstream invalid 109',
      },
      {
        code: 'Buffer()',
        settings: {
          node: {
            version: '6.0.0',
          },
        },
        options: [
          {
            ignoreModuleItems: ['buffer.Buffer()', 'new buffer.Buffer()'],
            ignoreGlobalItems: ['new Buffer()'],
          },
        ],
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 9,
          },
        ],
        name: 'Upstream invalid 110',
      },
      {
        code: 'var fs = require("fs");\nfs.exists("./foo.js", function() {});\nvar exists = require("fs").exists;\nconst {exists: existsAgain} = require("fs");',
        name: 'Documentation example 1',
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.",
            line: 2,
            column: 1,
            endLine: 2,
            endColumn: 10,
          },
          {
            messageId: 'deprecated',
            message:
              "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.",
            line: 3,
            column: 14,
            endLine: 3,
            endColumn: 34,
          },
          {
            messageId: 'deprecated',
            message:
              "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.",
            line: 4,
            column: 8,
            endLine: 4,
            endColumn: 27,
          },
        ],
      },
      {
        code: 'var Buffer = require("buffer").Buffer; Buffer = require("another-buffer"); new Buffer();',
        name: 'Documentation example 8',
        errors: [
          {
            messageId: 'deprecated',
            message:
              "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.",
            line: 1,
            column: 76,
            endLine: 1,
            endColumn: 88,
          },
        ],
      },
    ],
  },
);
