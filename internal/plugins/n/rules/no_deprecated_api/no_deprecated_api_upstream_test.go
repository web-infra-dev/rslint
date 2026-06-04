// TestNoDeprecatedApiUpstream migrates the full valid/invalid suite from upstream
// tests/lib/rules/no-deprecated-api.js 1:1. Position assertions cover line/column
// for every invalid case. rslint-specific lock-in cases live in the
// no_deprecated_api_extras_test.go file.
package no_deprecated_api_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/n/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/n/rules/no_deprecated_api"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDeprecatedApiUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_deprecated_api.NoDeprecatedApiRule,
		[]rule_tester.ValidTestCase{
			{Code: `require('buffer').Buffer`},
			{Code: `require('node:buffer').Buffer`},
			{Code: `foo(require('buffer').Buffer)`},
			{Code: `new (require('another-buffer').Buffer)()`},
			{Code: `var http = require('http'); http.request()`},
			{Code: `var {request} = require('http'); request()`},
			{Code: `(s ? require('https') : require('http')).request()`},
			{Code: `require(HTTP).createClient`},
			{Code: `import {Buffer} from 'another-buffer'; new Buffer()`},
			{Code: `import {request} from 'http'; request()`},
			{Code: `const {Buffer} = process.getBuiltinModule('another-buffer'); new Buffer()`},
			{Code: `const {request} = process.getBuiltinModule('http'); request()`},

			// On Node v6.8.0, fs.existsSync revived.
			{Code: `require('fs').existsSync;`},

			// use third parties.
			{Code: `require('domain/');`},
			{Code: `import domain from 'domain/';`},

			// https://github.com/mysticatea/eslint-plugin-node/issues/55
			{Code: `undefinedVar = require('fs')`},

			// ignore options
			{Code: `new (require('buffer').Buffer)()`, Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"new buffer.Buffer()"}}},
			{Code: `require('buffer').Buffer()`, Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"buffer.Buffer()"}}},
			{Code: `require('node:buffer').Buffer()`, Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"buffer.Buffer()"}}},
			{Code: `require('domain');`, Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"domain"}}},
			{Code: `require('events').EventEmitter.listenerCount;`, Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"events.EventEmitter.listenerCount"}}},
			{Code: `require('events').listenerCount;`, Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"events.listenerCount"}}},
			{Code: `new Buffer;`, Options: map[string]interface{}{"ignoreGlobalItems": []interface{}{"new Buffer()"}}},
			{Code: `Buffer();`, Options: map[string]interface{}{"ignoreGlobalItems": []interface{}{"Buffer()"}}},
			{Code: `Intl.v8BreakIterator;`, Options: map[string]interface{}{"ignoreGlobalItems": []interface{}{"Intl.v8BreakIterator"}}},
			{Code: `let {env: {NODE_REPL_HISTORY_FILE}} = process;`, Options: map[string]interface{}{"ignoreGlobalItems": []interface{}{"process.env.NODE_REPL_HISTORY_FILE"}}},

			// https://github.com/mysticatea/eslint-plugin-node/issues/65
			{Code: `require("domain/")`, Options: map[string]interface{}{"ignoreIndirectDependencies": true}},

			// https://github.com/mysticatea/eslint-plugin-node/issues/87
			{Code: `let fs = fs || require("fs")`},
		},
		[]rule_tester.InvalidTestCase{
			//----------------------------------------------------------------------
			// Modules
			//----------------------------------------------------------------------
			{
				Code:    `new (require('buffer').Buffer)()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `new (require('node:buffer').Buffer)()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('buffer').Buffer()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('node:buffer').Buffer()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `var b = require('buffer'); new b.Buffer()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 28}},
			},
			{
				Code:    `var b = require('buffer'); new b['Buffer']()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 28}},
			},
			{
				Code:    "var b = require('buffer'); new b[`Buffer`]()",
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 28}},
			},
			{
				Code:    `var b = require('buffer').Buffer; new b()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 35}},
			},
			{
				Code:    `var b; new ((b = require('buffer')).Buffer)(); new b.Buffer()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 8},
					{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 48},
				},
			},
			{
				Code:    `var {Buffer: b} = require('buffer'); new b()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 38}},
			},
			{
				Code:    `var {['Buffer']: b = null} = require('buffer'); new b()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 49}},
			},
			{
				Code:    `var {'Buffer': b = null} = require('buffer'); new b()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 47}},
			},
			{
				Code:    `var {Buffer: b = require('buffer').Buffer} = {}; new b()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 50}},
			},
			{
				Code:    `require('buffer').SlowBuffer`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('node:buffer').SlowBuffer`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `var b = require('buffer'); b.SlowBuffer`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 28}},
			},
			{
				Code:    `var {SlowBuffer: b} = require('buffer');`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 6}},
			},

			//----------------------------------------------------------------------
			{
				Code:    `require('_linklist');`,
				Options: map[string]interface{}{"version": "5.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'_linklist' module was deprecated since v5.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('async_hooks').currentId;`,
				Options: map[string]interface{}{"version": "8.2.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'async_hooks.currentId' was deprecated since v8.2.0. Use 'async_hooks.executionAsyncId()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('async_hooks').triggerId;`,
				Options: map[string]interface{}{"version": "8.2.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'async_hooks.triggerId' was deprecated since v8.2.0. Use 'async_hooks.triggerAsyncId()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('constants');`,
				Options: map[string]interface{}{"version": "6.3.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'constants' module was deprecated since v6.3.0. Use 'constants' property of each module instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('crypto').Credentials;`,
				Options: map[string]interface{}{"version": "0.12.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'crypto.Credentials' was deprecated since v0.12.0. Use 'tls.SecureContext' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('crypto').createCredentials;`,
				Options: map[string]interface{}{"version": "0.12.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'crypto.createCredentials' was deprecated since v0.12.0. Use 'tls.createSecureContext()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('domain');`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'domain' module was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('events').EventEmitter.listenerCount;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'events.EventEmitter.listenerCount' was deprecated since v4.0.0. Use 'events.EventEmitter#listenerCount()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('events').listenerCount;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'events.listenerCount' was deprecated since v4.0.0. Use 'events.EventEmitter#listenerCount()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('freelist');`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'freelist' module was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('fs').SyncWriteStream;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.SyncWriteStream' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('fs').exists;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('fs').lchmod;`,
				Options: map[string]interface{}{"version": "0.4.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.lchmod' was deprecated since v0.4.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('fs').lchmodSync;`,
				Options: map[string]interface{}{"version": "0.4.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.lchmodSync' was deprecated since v0.4.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('http').createClient;`,
				Options: map[string]interface{}{"version": "0.10.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'http.createClient' was deprecated since v0.10.0. Use 'http.request()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('module').requireRepl;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'module.requireRepl' was deprecated since v6.0.0. Use 'require(\"repl\")' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('module').Module.requireRepl;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'module.Module.requireRepl' was deprecated since v6.0.0. Use 'require(\"repl\")' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('module')._debug;`,
				Options: map[string]interface{}{"version": "9.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'module._debug' was deprecated since v9.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('module').Module._debug;`,
				Options: map[string]interface{}{"version": "9.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'module.Module._debug' was deprecated since v9.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('os').getNetworkInterfaces;`,
				Options: map[string]interface{}{"version": "0.6.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'os.getNetworkInterfaces' was deprecated since v0.6.0. Use 'os.networkInterfaces()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('os').tmpDir;`,
				Options: map[string]interface{}{"version": "7.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'os.tmpDir' was deprecated since v7.0.0. Use 'os.tmpdir()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('path')._makeLong;`,
				Options: map[string]interface{}{"version": "9.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'path._makeLong' was deprecated since v9.0.0. Use 'path.toNamespacedPath()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('punycode');`,
				Options: map[string]interface{}{"version": "7.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'punycode' module was deprecated since v7.0.0. Use 'https://www.npmjs.com/package/punycode' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('readline').codePointAt;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'readline.codePointAt' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('readline').getStringWidth;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'readline.getStringWidth' was deprecated since v6.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('readline').isFullWidthCodePoint;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'readline.isFullWidthCodePoint' was deprecated since v6.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('readline').stripVTControlCharacters;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'readline.stripVTControlCharacters' was deprecated since v6.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('sys');`,
				Options: map[string]interface{}{"version": "0.3.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'sys' module was deprecated since v0.3.0. Use 'util' module instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('tls').CleartextStream;`,
				Options: map[string]interface{}{"version": "0.10.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'tls.CleartextStream' was deprecated since v0.10.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('tls').CryptoStream;`,
				Options: map[string]interface{}{"version": "0.12.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'tls.CryptoStream' was deprecated since v0.12.0. Use 'tls.TLSSocket' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('tls').SecurePair;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'tls.SecurePair' was deprecated since v6.0.0. Use 'tls.TLSSocket' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('tls').createSecurePair;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'tls.createSecurePair' was deprecated since v6.0.0. Use 'tls.TLSSocket' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('tls').parseCertString;`,
				Options: map[string]interface{}{"version": "8.6.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'tls.parseCertString' was deprecated since v8.6.0. Use 'querystring.parse()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('tty').setRawMode;`,
				Options: map[string]interface{}{"version": "0.10.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'tty.setRawMode' was deprecated since v0.10.0. Use 'tty.ReadStream#setRawMode()' (e.g. 'process.stdin.setRawMode()') instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').debug;`,
				Options: map[string]interface{}{"version": "0.12.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.debug' was deprecated since v0.12.0. Use 'console.error()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').error;`,
				Options: map[string]interface{}{"version": "0.12.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.error' was deprecated since v0.12.0. Use 'console.error()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isArray;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isArray' was deprecated since v4.0.0. Use 'Array.isArray()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isBoolean;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isBoolean' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isBuffer;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isBuffer' was deprecated since v4.0.0. Use 'Buffer.isBuffer()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isDate;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isDate' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isError;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isError' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isFunction;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isFunction' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isNull;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isNull' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isNullOrUndefined;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isNullOrUndefined' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isNumber;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isNumber' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isObject;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isObject' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isPrimitive;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isPrimitive' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isRegExp;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isRegExp' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isString;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isString' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isSymbol;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isSymbol' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').isUndefined;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.isUndefined' was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').log;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.log' was deprecated since v6.0.0. Use a third party module instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').print;`,
				Options: map[string]interface{}{"version": "0.12.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.print' was deprecated since v0.12.0. Use 'console.log()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').pump;`,
				Options: map[string]interface{}{"version": "0.10.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.pump' was deprecated since v0.10.0. Use 'stream.Readable#pipe()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util').puts;`,
				Options: map[string]interface{}{"version": "0.12.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util.puts' was deprecated since v0.12.0. Use 'console.log()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('util')._extend;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'util._extend' was deprecated since v6.0.0. Use 'Object.assign()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('vm').runInDebugContext;`,
				Options: map[string]interface{}{"version": "8.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'vm.runInDebugContext' was deprecated since v8.0.0.", Line: 1, Column: 1}},
			},

			// ES2015 Modules
			{
				Code:    `import b from 'buffer'; new b.Buffer()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 25}},
			},
			{
				Code:    `import b from 'node:buffer'; new b.Buffer()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 30}},
			},
			{
				Code:    `import * as b from 'buffer'; new b.Buffer()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 30}},
			},
			{
				Code:    `import * as b from 'buffer'; new b.default.Buffer()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 30}},
			},
			{
				Code:    `import {Buffer as b} from 'buffer'; new b()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 37}},
			},
			{
				Code:    `import b from 'buffer'; b.SlowBuffer`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 25}},
			},
			{
				Code:    `import * as b from 'buffer'; b.SlowBuffer`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 30}},
			},
			{
				Code:    `import * as b from 'buffer'; b.default.SlowBuffer`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 30}},
			},
			{
				Code:    `import {SlowBuffer as b} from 'buffer';`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 9}},
			},
			{
				Code:    `import domain from 'domain';`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'domain' module was deprecated since v4.0.0.", Line: 1, Column: 1}},
			},

			// ignore combinations
			{
				Code:    `new (require('buffer').Buffer)()`,
				Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"buffer.Buffer()"}, "ignoreGlobalItems": []interface{}{"Buffer()", "new Buffer()"}, "version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('buffer').Buffer()`,
				Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"new buffer.Buffer()"}, "ignoreGlobalItems": []interface{}{"Buffer()", "new Buffer()"}, "version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('module').createRequireFromPath()`,
				Options: map[string]interface{}{"version": "12.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'module.createRequireFromPath' was deprecated since v12.2.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('module').createRequireFromPath()`,
				Options: map[string]interface{}{"version": "12.2.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'module.createRequireFromPath' was deprecated since v12.2.0. Use 'module.createRequire()' instead.", Line: 1, Column: 1}},
			},

			// process.getBuiltinModule()
			{
				Code:    `const b = process.getBuiltinModule('buffer'); new b.Buffer()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 47}},
			},
			{
				Code:    `const b = process.getBuiltinModule('node:buffer'); new b.Buffer()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 52}},
			},
			{
				Code:    `const {Buffer} = process.getBuiltinModule('buffer'); new Buffer()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 54}},
			},
			{
				Code:    `const {Buffer:b} = process.getBuiltinModule('buffer'); new b()`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 56}},
			},
			{
				Code:    `const b = process.getBuiltinModule('buffer'); b.SlowBuffer`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 47}},
			},
			{
				Code:    `const domain = process.getBuiltinModule('domain');`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'domain' module was deprecated since v4.0.0.", Line: 1, Column: 16}},
			},
			{
				Code:    `new (process.getBuiltinModule('buffer').Buffer)()`,
				Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"buffer.Buffer()"}, "ignoreGlobalItems": []interface{}{"Buffer()", "new Buffer()"}, "version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `process.getBuiltinModule('buffer').Buffer()`,
				Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"new buffer.Buffer()"}, "ignoreGlobalItems": []interface{}{"Buffer()", "new Buffer()"}, "version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `process.getBuiltinModule('module').createRequireFromPath()`,
				Options: map[string]interface{}{"version": "12.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'module.createRequireFromPath' was deprecated since v12.2.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `process.getBuiltinModule('module').createRequireFromPath()`,
				Options: map[string]interface{}{"version": "12.2.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'module.createRequireFromPath' was deprecated since v12.2.0. Use 'module.createRequire()' instead.", Line: 1, Column: 1}},
			},

			//----------------------------------------------------------------------
			// Global Variables
			//----------------------------------------------------------------------
			{
				Code:    `new Buffer;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `Buffer();`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			// SKIP upstream's `/*globals GLOBAL*/` directive — rslint does not
			// support ESLint global-directive comments; the bare reference is a
			// live global under rslint's "undeclared name is global" semantics.
			{
				Code:    `GLOBAL;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'GLOBAL' was deprecated since v6.0.0. Use 'global' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `Intl.v8BreakIterator;`,
				Options: map[string]interface{}{"version": "7.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "removed", Message: "'Intl.v8BreakIterator' was deprecated since v7.0.0, and removed in v9.0.0.", Line: 1, Column: 1}},
			},
			{
				Code:    `require.extensions;`,
				Options: map[string]interface{}{"version": "0.12.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'require.extensions' was deprecated since v0.12.0. Use compiling them ahead of time instead.", Line: 1, Column: 1}},
			},
			// SKIP upstream's `languageOptions.globals: { root: false }` — rslint
			// has no global-declaration framework concept; the bare reference is a
			// live global here.
			{
				Code:    `root;`,
				Options: map[string]interface{}{"version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'root' was deprecated since v6.0.0. Use 'global' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `process.EventEmitter;`,
				Options: map[string]interface{}{"version": "0.6.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'process.EventEmitter' was deprecated since v0.6.0. Use 'require(\"events\")' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `process.env.NODE_REPL_HISTORY_FILE;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'process.env.NODE_REPL_HISTORY_FILE' was deprecated since v4.0.0. Use 'NODE_REPL_HISTORY' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `let {env: {NODE_REPL_HISTORY_FILE}} = process;`,
				Options: map[string]interface{}{"version": "4.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'process.env.NODE_REPL_HISTORY_FILE' was deprecated since v4.0.0. Use 'NODE_REPL_HISTORY' instead.", Line: 1, Column: 12}},
			},
			{
				Code:    `new Buffer()`,
				Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"buffer.Buffer()", "new buffer.Buffer()"}, "ignoreGlobalItems": []interface{}{"Buffer()"}, "version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `Buffer()`,
				Options: map[string]interface{}{"ignoreModuleItems": []interface{}{"buffer.Buffer()", "new buffer.Buffer()"}, "ignoreGlobalItems": []interface{}{"new Buffer()"}, "version": "6.0.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			// settings.node.version is honoured by rslint via ctx.Settings.
			{
				Code:     `Buffer()`,
				Settings: map[string]interface{}{"node": map[string]interface{}{"version": "6.0.0"}},
				Options:  map[string]interface{}{"ignoreModuleItems": []interface{}{"buffer.Buffer()", "new buffer.Buffer()"}, "ignoreGlobalItems": []interface{}{"new Buffer()"}},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1}},
			},
		},
	)
}
