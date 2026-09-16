// Upstream tests and documentation examples from eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-deprecated-api.js
package no_deprecated_api

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDeprecatedAPIUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Upstream valid 1
		{Code: "require('buffer').Buffer"},
		// Upstream valid 2
		{Code: "require('node:buffer').Buffer"},
		// Upstream valid 3
		{Code: "foo(require('buffer').Buffer)"},
		// Upstream valid 4
		{Code: "new (require('another-buffer').Buffer)()"},
		// Upstream valid 5
		{Code: "var http = require('http'); http.request()"},
		// Upstream valid 6
		{Code: "var {request} = require('http'); request()"},
		// Upstream valid 7
		{Code: "(s ? require('https') : require('http')).request()"},
		// Upstream valid 8
		{Code: "require(HTTP).createClient"},
		// Upstream valid 9
		{Code: "import {Buffer} from 'another-buffer'; new Buffer()",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// Upstream valid 10
		{Code: "import {request} from 'http'; request()",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// Upstream valid 11
		{Code: "const {Buffer} = process.getBuiltinModule('another-buffer'); new Buffer()"},
		// Upstream valid 12
		{Code: "const {request} = process.getBuiltinModule('http'); request()"},
		// Upstream valid 13
		{Code: "require('fs').existsSync;"},
		// Upstream valid 14
		{Code: "require('domain/');"},
		// Upstream valid 15
		{Code: "import domain from 'domain/';",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		},
		// Upstream valid 16
		{Code: "undefinedVar = require('fs')"},
		// Upstream valid 17
		{Code: "new (require('buffer').Buffer)()",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"new buffer.Buffer()"}}},
		},
		// Upstream valid 18
		{Code: "require('buffer').Buffer()",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"buffer.Buffer()"}}},
		},
		// Upstream valid 19
		{Code: "require('node:buffer').Buffer()",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"buffer.Buffer()"}}},
		},
		// Upstream valid 20
		{Code: "require('domain');",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"domain"}}},
		},
		// Upstream valid 21
		{Code: "require('events').EventEmitter.listenerCount;",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"events.EventEmitter.listenerCount"}}},
		},
		// Upstream valid 22
		{Code: "require('events').listenerCount;",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"events.listenerCount"}}},
		},
		// Upstream valid 23
		{Code: "new Buffer;",
			Options: []any{map[string]any{"ignoreGlobalItems": []any{"new Buffer()"}}},
		},
		// Upstream valid 24
		{Code: "Buffer();",
			Options: []any{map[string]any{"ignoreGlobalItems": []any{"Buffer()"}}},
		},
		// Upstream valid 25
		{Code: "Intl.v8BreakIterator;",
			Options: []any{map[string]any{"ignoreGlobalItems": []any{"Intl.v8BreakIterator"}}},
		},
		// Upstream valid 26
		{Code: "let {env: {NODE_REPL_HISTORY_FILE}} = process;",
			Options: []any{map[string]any{"ignoreGlobalItems": []any{"process.env.NODE_REPL_HISTORY_FILE"}}},
		},
		// Upstream valid 27
		{Code: "require(\"domain/\")",
			Options: []any{map[string]any{"ignoreIndirectDependencies": true}},
		},
		// Upstream valid 28
		{Code: "let fs = fs || require(\"fs\")"},
		// Documentation example 2
		{Code: "const buffer = require(\"buffer\"); const data = new buffer.Buffer(10);",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"new buffer.Buffer()"}}},
		},
		// Documentation example 3
		{Code: "const data = new Buffer(10);",
			Options: []any{map[string]any{"ignoreGlobalItems": []any{"new Buffer()"}}},
		},
		// Documentation example 4
		{Code: "require(foo).aDeprecatedProperty;\nrequire(\"http\")[A_DEPRECATED_PROPERTY]();"},
		// Documentation example 5
		{Code: "var obj = {Buffer: require(\"buffer\").Buffer}; new obj.Buffer();"},
		// Documentation example 6
		{Code: "var obj = {}; obj.Buffer = require(\"buffer\").Buffer; new obj.Buffer();"},
		// Documentation example 7
		{Code: "(function(Buffer) { new Buffer(); })(require(\"buffer\").Buffer);"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Upstream invalid 1
		{Code: "new (require('buffer').Buffer)()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
			},
		},
		// Upstream invalid 2
		{Code: "new (require('node:buffer').Buffer)()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
			},
		},
		// Upstream invalid 3
		{Code: "require('buffer').Buffer()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			},
		},
		// Upstream invalid 4
		{Code: "require('node:buffer').Buffer()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			},
		},
		// Upstream invalid 5
		{Code: "var b = require('buffer'); new b.Buffer()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 28, EndLine: 1, EndColumn: 42},
			},
		},
		// Upstream invalid 6
		{Code: "var b = require('buffer'); new b['Buffer']()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 28, EndLine: 1, EndColumn: 45},
			},
		},
		// Upstream invalid 7
		{Code: "var b = require('buffer'); new b[`Buffer`]()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 28, EndLine: 1, EndColumn: 45},
			},
		},
		// Upstream invalid 8
		{Code: "var b = require('buffer').Buffer; new b()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 35, EndLine: 1, EndColumn: 42},
			},
		},
		// Upstream invalid 9
		{Code: "var b; new ((b = require('buffer')).Buffer)(); new b.Buffer()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 8, EndLine: 1, EndColumn: 46},
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 48, EndLine: 1, EndColumn: 62},
			},
		},
		// Upstream invalid 10
		{Code: "var {Buffer: b} = require('buffer'); new b()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 38, EndLine: 1, EndColumn: 45},
			},
		},
		// Upstream invalid 11
		{Code: "var {['Buffer']: b = null} = require('buffer'); new b()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 49, EndLine: 1, EndColumn: 56},
			},
		},
		// Upstream invalid 12
		{Code: "var {'Buffer': b = null} = require('buffer'); new b()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 47, EndLine: 1, EndColumn: 54},
			},
		},
		// Upstream invalid 13
		{Code: "var {Buffer: b = require('buffer').Buffer} = {}; new b()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 50, EndLine: 1, EndColumn: 57},
			},
		},
		// Upstream invalid 14
		{Code: "require('buffer').SlowBuffer",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
			},
		},
		// Upstream invalid 15
		{Code: "require('node:buffer').SlowBuffer",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			},
		},
		// Upstream invalid 16
		{Code: "var b = require('buffer'); b.SlowBuffer",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 28, EndLine: 1, EndColumn: 40},
			},
		},
		// Upstream invalid 17
		{Code: "var {SlowBuffer: b} = require('buffer');",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 6, EndLine: 1, EndColumn: 19},
			},
		},
		// Upstream invalid 18
		{Code: "require('_linklist');",
			Options: []any{map[string]any{"version": "5.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'_linklist' module was deprecated since v5.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			},
		},
		// Upstream invalid 19
		{Code: "require('async_hooks').currentId;",
			Options: []any{map[string]any{"version": "8.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'async_hooks.currentId' was deprecated since v8.2.0. Use 'async_hooks.executionAsyncId()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
			},
		},
		// Upstream invalid 20
		{Code: "require('async_hooks').triggerId;",
			Options: []any{map[string]any{"version": "8.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'async_hooks.triggerId' was deprecated since v8.2.0. Use 'async_hooks.triggerAsyncId()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
			},
		},
		// Upstream invalid 21
		{Code: "require('constants');",
			Options: []any{map[string]any{"version": "6.3.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'constants' module was deprecated since v6.3.0. Use 'constants' property of each module instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			},
		},
		// Upstream invalid 22
		{Code: "require('crypto').Credentials;",
			Options: []any{map[string]any{"version": "0.12.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'crypto.Credentials' was deprecated since v0.12.0. Use 'tls.SecureContext' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			},
		},
		// Upstream invalid 23
		{Code: "require('crypto').createCredentials;",
			Options: []any{map[string]any{"version": "0.12.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'crypto.createCredentials' was deprecated since v0.12.0. Use 'tls.createSecureContext()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 36},
			},
		},
		// Upstream invalid 24
		{Code: "require('domain');",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'domain' module was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
			},
		},
		// Upstream invalid 25
		{Code: "require('events').EventEmitter.listenerCount;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'events.EventEmitter.listenerCount' was deprecated since v4.0.0. Use 'events.EventEmitter#listenerCount()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
			},
		},
		// Upstream invalid 26
		{Code: "require('events').listenerCount;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'events.listenerCount' was deprecated since v4.0.0. Use 'events.EventEmitter#listenerCount()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			},
		},
		// Upstream invalid 27
		{Code: "require('freelist');",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'freelist' module was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
			},
		},
		// Upstream invalid 28
		{Code: "require('fs').SyncWriteStream;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.SyncWriteStream' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			},
		},
		// Upstream invalid 29
		{Code: "require('fs').exists;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			},
		},
		// Upstream invalid 30
		{Code: "require('fs').lchmod;",
			Options: []any{map[string]any{"version": "0.4.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.lchmod' was deprecated since v0.4.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			},
		},
		// Upstream invalid 31
		{Code: "require('fs').lchmodSync;",
			Options: []any{map[string]any{"version": "0.4.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.lchmodSync' was deprecated since v0.4.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			},
		},
		// Upstream invalid 32
		{Code: "require('http').createClient;",
			Options: []any{map[string]any{"version": "0.10.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'http.createClient' was deprecated since v0.10.0. Use 'http.request()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
			},
		},
		// Upstream invalid 33
		{Code: "require('module').requireRepl;",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'module.requireRepl' was deprecated since v6.0.0. Use 'require(\"repl\")' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			},
		},
		// Upstream invalid 34
		{Code: "require('module').Module.requireRepl;",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'module.Module.requireRepl' was deprecated since v6.0.0. Use 'require(\"repl\")' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
			},
		},
		// Upstream invalid 35
		{Code: "require('module')._debug;",
			Options: []any{map[string]any{"version": "9.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'module._debug' was deprecated since v9.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			},
		},
		// Upstream invalid 36
		{Code: "require('module').Module._debug;",
			Options: []any{map[string]any{"version": "9.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'module.Module._debug' was deprecated since v9.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			},
		},
		// Upstream invalid 37
		{Code: "require('os').getNetworkInterfaces;",
			Options: []any{map[string]any{"version": "0.6.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'os.getNetworkInterfaces' was deprecated since v0.6.0. Use 'os.networkInterfaces()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
			},
		},
		// Upstream invalid 38
		{Code: "require('os').tmpDir;",
			Options: []any{map[string]any{"version": "7.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'os.tmpDir' was deprecated since v7.0.0. Use 'os.tmpdir()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			},
		},
		// Upstream invalid 39
		{Code: "require('path')._makeLong;",
			Options: []any{map[string]any{"version": "9.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'path._makeLong' was deprecated since v9.0.0. Use 'path.toNamespacedPath()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
			},
		},
		// Upstream invalid 40
		{Code: "require('punycode');",
			Options: []any{map[string]any{"version": "7.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'punycode' module was deprecated since v7.0.0. Use 'https://www.npmjs.com/package/punycode' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
			},
		},
		// Upstream invalid 41
		{Code: "require('readline').codePointAt;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'readline.codePointAt' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			},
		},
		// Upstream invalid 42
		{Code: "require('readline').getStringWidth;",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'readline.getStringWidth' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
			},
		},
		// Upstream invalid 43
		{Code: "require('readline').isFullWidthCodePoint;",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'readline.isFullWidthCodePoint' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
			},
		},
		// Upstream invalid 44
		{Code: "require('readline').stripVTControlCharacters;",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'readline.stripVTControlCharacters' was deprecated since v6.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
			},
		},
		// Upstream invalid 45
		{Code: "require('sys');",
			Options: []any{map[string]any{"version": "0.3.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'sys' module was deprecated since v0.3.0. Use 'util' module instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
			},
		},
		// Upstream invalid 46
		{Code: "require('tls').CleartextStream;",
			Options: []any{map[string]any{"version": "0.10.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'tls.CleartextStream' was deprecated since v0.10.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			},
		},
		// Upstream invalid 47
		{Code: "require('tls').CryptoStream;",
			Options: []any{map[string]any{"version": "0.12.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'tls.CryptoStream' was deprecated since v0.12.0. Use 'tls.TLSSocket' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
			},
		},
		// Upstream invalid 48
		{Code: "require('tls').SecurePair;",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'tls.SecurePair' was deprecated since v6.0.0. Use 'tls.TLSSocket' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
			},
		},
		// Upstream invalid 49
		{Code: "require('tls').createSecurePair;",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'tls.createSecurePair' was deprecated since v6.0.0. Use 'tls.TLSSocket' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			},
		},
		// Upstream invalid 50
		{Code: "require('tls').parseCertString;",
			Options: []any{map[string]any{"version": "8.6.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'tls.parseCertString' was deprecated since v8.6.0. Use 'querystring.parse()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			},
		},
		// Upstream invalid 51
		{Code: "require('tty').setRawMode;",
			Options: []any{map[string]any{"version": "0.10.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'tty.setRawMode' was deprecated since v0.10.0. Use 'tty.ReadStream#setRawMode()' (e.g. 'process.stdin.setRawMode()') instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
			},
		},
		// Upstream invalid 52
		{Code: "require('util').debug;",
			Options: []any{map[string]any{"version": "0.12.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.debug' was deprecated since v0.12.0. Use 'console.error()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			},
		},
		// Upstream invalid 53
		{Code: "require('util').error;",
			Options: []any{map[string]any{"version": "0.12.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.error' was deprecated since v0.12.0. Use 'console.error()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			},
		},
		// Upstream invalid 54
		{Code: "require('util').isArray;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isArray' was deprecated since v4.0.0. Use 'Array.isArray()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			},
		},
		// Upstream invalid 55
		{Code: "require('util').isBoolean;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isBoolean' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
			},
		},
		// Upstream invalid 56
		{Code: "require('util').isBuffer;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isBuffer' was deprecated since v4.0.0. Use 'Buffer.isBuffer()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			},
		},
		// Upstream invalid 57
		{Code: "require('util').isDate;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isDate' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
			},
		},
		// Upstream invalid 58
		{Code: "require('util').isError;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isError' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			},
		},
		// Upstream invalid 59
		{Code: "require('util').isFunction;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isFunction' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			},
		},
		// Upstream invalid 60
		{Code: "require('util').isNull;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isNull' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
			},
		},
		// Upstream invalid 61
		{Code: "require('util').isNullOrUndefined;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isNullOrUndefined' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			},
		},
		// Upstream invalid 62
		{Code: "require('util').isNumber;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isNumber' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			},
		},
		// Upstream invalid 63
		{Code: "require('util').isObject;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isObject' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			},
		},
		// Upstream invalid 64
		{Code: "require('util').isPrimitive;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isPrimitive' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
			},
		},
		// Upstream invalid 65
		{Code: "require('util').isRegExp;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isRegExp' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			},
		},
		// Upstream invalid 66
		{Code: "require('util').isString;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isString' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			},
		},
		// Upstream invalid 67
		{Code: "require('util').isSymbol;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isSymbol' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			},
		},
		// Upstream invalid 68
		{Code: "require('util').isUndefined;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.isUndefined' was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
			},
		},
		// Upstream invalid 69
		{Code: "require('util').log;",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.log' was deprecated since v6.0.0. Use a third party module instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
			},
		},
		// Upstream invalid 70
		{Code: "require('util').print;",
			Options: []any{map[string]any{"version": "0.12.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.print' was deprecated since v0.12.0. Use 'console.log()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			},
		},
		// Upstream invalid 71
		{Code: "require('util').pump;",
			Options: []any{map[string]any{"version": "0.10.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.pump' was deprecated since v0.10.0. Use 'stream.Readable#pipe()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			},
		},
		// Upstream invalid 72
		{Code: "require('util').puts;",
			Options: []any{map[string]any{"version": "0.12.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util.puts' was deprecated since v0.12.0. Use 'console.log()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			},
		},
		// Upstream invalid 73
		{Code: "require('util')._extend;",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'util._extend' was deprecated since v6.0.0. Use 'Object.assign()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			},
		},
		// Upstream invalid 74
		{Code: "require('vm').runInDebugContext;",
			Options: []any{map[string]any{"version": "8.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'vm.runInDebugContext' was deprecated since v8.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			},
		},
		// Upstream invalid 75
		{Code: "import b from 'buffer'; new b.Buffer()",
			Options:         []any{map[string]any{"version": "6.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 25, EndLine: 1, EndColumn: 39},
			},
		},
		// Upstream invalid 76
		{Code: "import b from 'node:buffer'; new b.Buffer()",
			Options:         []any{map[string]any{"version": "6.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 30, EndLine: 1, EndColumn: 44},
			},
		},
		// Upstream invalid 77
		{Code: "import * as b from 'buffer'; new b.Buffer()",
			Options:         []any{map[string]any{"version": "6.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 30, EndLine: 1, EndColumn: 44},
			},
		},
		// Upstream invalid 78
		{Code: "import * as b from 'buffer'; new b.default.Buffer()",
			Options:         []any{map[string]any{"version": "6.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 30, EndLine: 1, EndColumn: 52},
			},
		},
		// Upstream invalid 79
		{Code: "import {Buffer as b} from 'buffer'; new b()",
			Options:         []any{map[string]any{"version": "6.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 37, EndLine: 1, EndColumn: 44},
			},
		},
		// Upstream invalid 80
		{Code: "import b from 'buffer'; b.SlowBuffer",
			Options:         []any{map[string]any{"version": "6.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 25, EndLine: 1, EndColumn: 37},
			},
		},
		// Upstream invalid 81
		{Code: "import * as b from 'buffer'; b.SlowBuffer",
			Options:         []any{map[string]any{"version": "6.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 30, EndLine: 1, EndColumn: 42},
			},
		},
		// Upstream invalid 82
		{Code: "import * as b from 'buffer'; b.default.SlowBuffer",
			Options:         []any{map[string]any{"version": "6.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 30, EndLine: 1, EndColumn: 50},
			},
		},
		// Upstream invalid 83
		{Code: "import {SlowBuffer as b} from 'buffer';",
			Options:         []any{map[string]any{"version": "6.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 9, EndLine: 1, EndColumn: 24},
			},
		},
		// Upstream invalid 84
		{Code: "import domain from 'domain';",
			Options:         []any{map[string]any{"version": "4.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'domain' module was deprecated since v4.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
			},
		},
		// Upstream invalid 85
		{Code: "new (require('buffer').Buffer)()",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"buffer.Buffer()"}, "ignoreGlobalItems": []any{"Buffer()", "new Buffer()"}, "version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
			},
		},
		// Upstream invalid 86
		{Code: "require('buffer').Buffer()",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"new buffer.Buffer()"}, "ignoreGlobalItems": []any{"Buffer()", "new Buffer()"}, "version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			},
		},
		// Upstream invalid 87
		{Code: "require('module').createRequireFromPath()",
			Options: []any{map[string]any{"version": "12.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'module.createRequireFromPath' was deprecated since v12.2.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
			},
		},
		// Upstream invalid 88
		{Code: "require('module').createRequireFromPath()",
			Options: []any{map[string]any{"version": "12.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'module.createRequireFromPath' was deprecated since v12.2.0. Use 'module.createRequire()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
			},
		},
		// Upstream invalid 89
		{Code: "const b = process.getBuiltinModule('buffer'); new b.Buffer()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 47, EndLine: 1, EndColumn: 61},
			},
		},
		// Upstream invalid 90
		{Code: "const b = process.getBuiltinModule('node:buffer'); new b.Buffer()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 52, EndLine: 1, EndColumn: 66},
			},
		},
		// Upstream invalid 91
		{Code: "const {Buffer} = process.getBuiltinModule('buffer'); new Buffer()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 54, EndLine: 1, EndColumn: 66},
			},
		},
		// Upstream invalid 92
		{Code: "const {Buffer:b} = process.getBuiltinModule('buffer'); new b()",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 56, EndLine: 1, EndColumn: 63},
			},
		},
		// Upstream invalid 93
		{Code: "const b = process.getBuiltinModule('buffer'); b.SlowBuffer",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 47, EndLine: 1, EndColumn: 59},
			},
		},
		// Upstream invalid 94
		{Code: "const domain = process.getBuiltinModule('domain');",
			Options:         []any{map[string]any{"version": "4.0.0"}},
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'domain' module was deprecated since v4.0.0.", Line: 1, Column: 16, EndLine: 1, EndColumn: 50},
			},
		},
		// Upstream invalid 95
		{Code: "new (process.getBuiltinModule('buffer').Buffer)()",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"buffer.Buffer()"}, "ignoreGlobalItems": []any{"Buffer()", "new Buffer()"}, "version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 50},
			},
		},
		// Upstream invalid 96
		{Code: "process.getBuiltinModule('buffer').Buffer()",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"new buffer.Buffer()"}, "ignoreGlobalItems": []any{"Buffer()", "new Buffer()"}, "version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
			},
		},
		// Upstream invalid 97
		{Code: "process.getBuiltinModule('module').createRequireFromPath()",
			Options: []any{map[string]any{"version": "12.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'module.createRequireFromPath' was deprecated since v12.2.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 57},
			},
		},
		// Upstream invalid 98
		{Code: "process.getBuiltinModule('module').createRequireFromPath()",
			Options: []any{map[string]any{"version": "12.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'module.createRequireFromPath' was deprecated since v12.2.0. Use 'module.createRequire()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 57},
			},
		},
		// Upstream invalid 99
		{Code: "new Buffer;",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
			},
		},
		// Upstream invalid 100
		{Code: "Buffer();",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
			},
		},
		// Upstream invalid 101
		{Code: "GLOBAL; /*globals GLOBAL*/",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'GLOBAL' was deprecated since v6.0.0. Use 'global' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			},
		},
		// Upstream invalid 102
		{Code: "Intl.v8BreakIterator;",
			Options: []any{map[string]any{"version": "7.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "removed", Message: "'Intl.v8BreakIterator' was deprecated since v7.0.0, and removed in v9.0.0.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			},
		},
		// Upstream invalid 103
		{Code: "require.extensions;",
			Options: []any{map[string]any{"version": "0.12.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'require.extensions' was deprecated since v0.12.0. Use compiling them ahead of time instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
			},
		},
		// Upstream invalid 104
		{Code: "root;",
			Options: []any{map[string]any{"version": "6.0.0"}},
			Globals: map[string]any{"root": "readonly"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'root' was deprecated since v6.0.0. Use 'global' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
			},
		},
		// Upstream invalid 105
		{Code: "process.EventEmitter;",
			Options: []any{map[string]any{"version": "0.6.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'process.EventEmitter' was deprecated since v0.6.0. Use 'require(\"events\")' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			},
		},
		// Upstream invalid 106
		{Code: "process.env.NODE_REPL_HISTORY_FILE;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'process.env.NODE_REPL_HISTORY_FILE' was deprecated since v4.0.0. Use 'NODE_REPL_HISTORY' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
			},
		},
		// Upstream invalid 107
		{Code: "let {env: {NODE_REPL_HISTORY_FILE}} = process;",
			Options: []any{map[string]any{"version": "4.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'process.env.NODE_REPL_HISTORY_FILE' was deprecated since v4.0.0. Use 'NODE_REPL_HISTORY' instead.", Line: 1, Column: 12, EndLine: 1, EndColumn: 34},
			},
		},
		// Upstream invalid 108
		{Code: "new Buffer()",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"buffer.Buffer()", "new buffer.Buffer()"}, "ignoreGlobalItems": []any{"Buffer()"}, "version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
			},
		},
		// Upstream invalid 109
		{Code: "Buffer()",
			Options: []any{map[string]any{"ignoreModuleItems": []any{"buffer.Buffer()", "new buffer.Buffer()"}, "ignoreGlobalItems": []any{"new Buffer()"}, "version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
			},
		},
		// Upstream invalid 110
		{Code: "Buffer()",
			Options:  []any{map[string]any{"ignoreModuleItems": []any{"buffer.Buffer()", "new buffer.Buffer()"}, "ignoreGlobalItems": []any{"new Buffer()"}}},
			Settings: map[string]any{"node": map[string]any{"version": "6.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
			},
		},
		// Documentation example 1
		{Code: "var fs = require(\"fs\");\nfs.exists(\"./foo.js\", function() {});\nvar exists = require(\"fs\").exists;\nconst {exists: existsAgain} = require(\"fs\");",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 2, Column: 1, EndLine: 2, EndColumn: 10},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 3, Column: 14, EndLine: 3, EndColumn: 34},
				{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 4, Column: 8, EndLine: 4, EndColumn: 27},
			},
		},
		// Documentation example 8
		{Code: "var Buffer = require(\"buffer\").Buffer; Buffer = require(\"another-buffer\"); new Buffer();",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 76, EndLine: 1, EndColumn: 88},
			},
		},
	}

	for i := range valid {
		valid[i].Globals = deprecatedTestGlobals(valid[i].Globals)
		if valid[i].LanguageOptions.SourceType == "" {
			valid[i].LanguageOptions.SourceType = "commonjs"
		}
	}
	for i := range invalid {
		invalid[i].Globals = deprecatedTestGlobals(invalid[i].Globals)
		if invalid[i].LanguageOptions.SourceType == "" {
			invalid[i].LanguageOptions.SourceType = "commonjs"
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDeprecatedAPIRule, valid, invalid)
}

func deprecatedTestGlobals(extra map[string]any) map[string]any {
	globals := map[string]any{"require": "readonly", "Buffer": "readonly", "process": "readonly", "global": "readonly", "Intl": "readonly"}
	for key, value := range extra {
		globals[key] = value
	}
	return globals
}
