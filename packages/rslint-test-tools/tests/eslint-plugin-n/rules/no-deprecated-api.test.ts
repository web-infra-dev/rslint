import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Mirrors the upstream Layer-1 valid/invalid semantics. Edge-shape, branch
// lock-in, and real-user cases live in the Go suite
// (internal/plugins/n/rules/no_deprecated_api/no_deprecated_api_extras_test.go).
ruleTester.run('no-deprecated-api', {} as never, {
  valid: [
    { code: `require('buffer').Buffer` },
    { code: `require('fs').existsSync;` },
    { code: `var http = require('http'); http.request()` },
    { code: `var {request} = require('http'); request()` },
    { code: `import {Buffer} from 'another-buffer'; new Buffer()` },
    { code: `const {request} = process.getBuiltinModule('http'); request()` },
    {
      code: `require('domain');`,
      options: [{ ignoreModuleItems: ['domain'] }],
    },
    { code: `new Buffer;`, options: [{ ignoreGlobalItems: ['new Buffer()'] }] },
    { code: `let fs = fs || require("fs")` },
  ],
  invalid: [
    {
      code: `require('fs').exists;`,
      errors: [
        `'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.`,
      ],
    },
    {
      code: `require('buffer').Buffer()`,
      options: [{ version: '6.0.0' }],
      errors: [
        `'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.`,
      ],
    },
    {
      code: `var b = require('buffer'); new b.Buffer()`,
      options: [{ version: '6.0.0' }],
      errors: [
        `'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.`,
      ],
    },
    {
      code: `require('buffer').SlowBuffer`,
      options: [{ version: '6.0.0' }],
      errors: [
        `'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.`,
      ],
    },
    {
      code: `require('domain');`,
      options: [{ version: '4.0.0' }],
      errors: [`'domain' module was deprecated since v4.0.0.`],
    },
    {
      code: `import b from 'buffer'; new b.Buffer()`,
      options: [{ version: '6.0.0' }],
      errors: [
        `'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.`,
      ],
    },
    {
      code: `const b = process.getBuiltinModule('buffer'); new b.Buffer()`,
      options: [{ version: '6.0.0' }],
      errors: [
        `'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.`,
      ],
    },
    {
      code: `new Buffer;`,
      options: [{ version: '6.0.0' }],
      errors: [
        `'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.`,
      ],
    },
    {
      code: `process.EventEmitter;`,
      options: [{ version: '0.6.0' }],
      errors: [
        `'process.EventEmitter' was deprecated since v0.6.0. Use 'require("events")' instead.`,
      ],
    },
    {
      code: `Intl.v8BreakIterator;`,
      options: [{ version: '7.0.0' }],
      errors: [
        `'Intl.v8BreakIterator' was deprecated since v7.0.0, and removed in v9.0.0.`,
      ],
    },
    {
      code: `require('module').createRequireFromPath()`,
      options: [{ version: '12.2.0' }],
      errors: [
        `'module.createRequireFromPath' was deprecated since v12.2.0. Use 'module.createRequire()' instead.`,
      ],
    },
  ],
});
