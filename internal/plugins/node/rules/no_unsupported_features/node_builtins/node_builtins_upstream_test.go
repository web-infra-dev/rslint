// cspell:ignore Fips Spkac fips ppid readv
// Upstream tests from eslint-plugin-n v18.3.0, including all expanded cases.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-unsupported-features/node-builtins.js
package node_builtins

import (
	"maps"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func builtinTestGlobals(overrides map[string]any) map[string]any {
	globals := map[string]any{
		"__dirname":                        "readonly",
		"__filename":                       "readonly",
		"AbortController":                  "readonly",
		"AbortSignal":                      "readonly",
		"atob":                             "readonly",
		"Blob":                             "readonly",
		"BroadcastChannel":                 "readonly",
		"btoa":                             "readonly",
		"Buffer":                           "readonly",
		"ByteLengthQueuingStrategy":        "readonly",
		"clearImmediate":                   "readonly",
		"clearInterval":                    "readonly",
		"clearTimeout":                     "readonly",
		"CloseEvent":                       "readonly",
		"CompressionStream":                "readonly",
		"console":                          "readonly",
		"CountQueuingStrategy":             "readonly",
		"crypto":                           "readonly",
		"Crypto":                           "readonly",
		"CryptoKey":                        "readonly",
		"CustomEvent":                      "readonly",
		"DecompressionStream":              "readonly",
		"DOMException":                     "readonly",
		"Event":                            "readonly",
		"EventTarget":                      "readonly",
		"exports":                          "writable",
		"fetch":                            "readonly",
		"File":                             "readonly",
		"FormData":                         "readonly",
		"global":                           "readonly",
		"Headers":                          "readonly",
		"MessageChannel":                   "readonly",
		"MessageEvent":                     "readonly",
		"MessagePort":                      "readonly",
		"module":                           "readonly",
		"navigator":                        "readonly",
		"Navigator":                        "readonly",
		"performance":                      "readonly",
		"Performance":                      "readonly",
		"PerformanceEntry":                 "readonly",
		"PerformanceMark":                  "readonly",
		"PerformanceMeasure":               "readonly",
		"PerformanceObserver":              "readonly",
		"PerformanceObserverEntryList":     "readonly",
		"PerformanceResourceTiming":        "readonly",
		"process":                          "readonly",
		"queueMicrotask":                   "readonly",
		"ReadableByteStreamController":     "readonly",
		"ReadableStream":                   "readonly",
		"ReadableStreamBYOBReader":         "readonly",
		"ReadableStreamBYOBRequest":        "readonly",
		"ReadableStreamDefaultController":  "readonly",
		"ReadableStreamDefaultReader":      "readonly",
		"Request":                          "readonly",
		"require":                          "readonly",
		"Response":                         "readonly",
		"setImmediate":                     "readonly",
		"setInterval":                      "readonly",
		"setTimeout":                       "readonly",
		"structuredClone":                  "readonly",
		"SubtleCrypto":                     "readonly",
		"TextDecoder":                      "readonly",
		"TextDecoderStream":                "readonly",
		"TextEncoder":                      "readonly",
		"TextEncoderStream":                "readonly",
		"TransformStream":                  "readonly",
		"TransformStreamDefaultController": "readonly",
		"URL":                              "readonly",
		"URLSearchParams":                  "readonly",
		"WebAssembly":                      "readonly",
		"WebSocket":                        "readonly",
		"WritableStream":                   "readonly",
		"WritableStreamDefaultController":  "readonly",
		"WritableStreamDefaultWriter":      "readonly",
	}
	maps.Copy(globals, overrides)
	return globals
}

func runBuiltinTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	for i := range valid {
		valid[i].Globals = builtinTestGlobals(valid[i].Globals)
		if valid[i].FileName == "" {
			valid[i].FileName = "input.js"
		}
		if valid[i].LanguageOptions.SourceType == "" {
			valid[i].LanguageOptions.SourceType = "module"
		}
	}
	for i := range invalid {
		invalid[i].Globals = builtinTestGlobals(invalid[i].Globals)
		if invalid[i].FileName == "" {
			invalid[i].FileName = "input.js"
		}
		if invalid[i].LanguageOptions.SourceType == "" {
			invalid[i].LanguageOptions.SourceType = "module"
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allowJs.json", t, &NodeBuiltinsRule, valid, invalid)
}

func TestNodeBuiltinsUpstream(t *testing.T) {
	runBuiltinTests(t,
		[]rule_tester.ValidTestCase{
			// assert: upstream valid 1.
			{Code: "require('assert').strictEqual()",
				Options: []any{map[string]any{"version": "0.12.0"}},
			},
			// assert: upstream valid 2.
			{Code: "var assert = require('assert'); assert(); assert.strictEqual()",
				Options: []any{map[string]any{"version": "0.12.0"}},
			},
			// assert: upstream valid 3.
			{Code: "require('assert').deepStrictEqual()",
				Options: []any{map[string]any{"version": "4.0.0"}},
			},
			// assert: upstream valid 4.
			{Code: "var assert = require('assert'); assert.deepStrictEqual()",
				Options: []any{map[string]any{"version": "4.0.0"}},
			},
			// assert: upstream valid 5.
			{Code: "var { deepStrictEqual } = require('assert'); deepStrictEqual()",
				Options: []any{map[string]any{"version": "4.0.0"}},
			},
			// assert: upstream valid 6.
			{Code: "import assert from 'assert'; assert.deepStrictEqual()",
				Options: []any{map[string]any{"version": "4.0.0"}},
			},
			// assert: upstream valid 7.
			{Code: "import { deepStrictEqual } from 'assert'; deepStrictEqual()",
				Options: []any{map[string]any{"version": "4.0.0"}},
			},
			// assert: upstream valid 8.
			{Code: "require('assert').notDeepStrictEqual()",
				Options: []any{map[string]any{"version": "4.0.0"}},
			},
			// assert: upstream valid 9.
			{Code: "require('assert').rejects()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// assert: upstream valid 10.
			{Code: "require('assert').doesNotReject()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// assert: upstream valid 11.
			{Code: "require('assert').strict.rejects()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// assert: upstream valid 12.
			{Code: "require('assert').strict.doesNotReject()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// assert: upstream valid 13.
			{Code: "var assert = require('assert').strict",
				Options: []any{map[string]any{"version": "9.9.0"}},
			},
			// assert: upstream valid 14.
			{Code: "var {strict: assert} = require('assert'); assert.rejects()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// assert: upstream valid 15.
			{Code: "require('assert').deepStrictEqual()",
				Options: []any{map[string]any{"version": "3.9.9", "ignores": []any{"assert.deepStrictEqual"}}},
			},
			// assert: upstream valid 16.
			{Code: "var assert = require('assert'); assert.deepStrictEqual()",
				Options: []any{map[string]any{"version": "3.9.9", "ignores": []any{"assert.deepStrictEqual"}}},
			},
			// assert: upstream valid 17.
			{Code: "var { deepStrictEqual } = require('assert'); deepStrictEqual()",
				Options: []any{map[string]any{"version": "3.9.9", "ignores": []any{"assert.deepStrictEqual"}}},
			},
			// assert: upstream valid 18.
			{Code: "import assert from 'assert'; assert.deepStrictEqual()",
				Options: []any{map[string]any{"version": "3.9.9", "ignores": []any{"assert.deepStrictEqual"}}},
			},
			// assert: upstream valid 19.
			{Code: "import { deepStrictEqual } from 'assert'; deepStrictEqual()",
				Options: []any{map[string]any{"version": "3.9.9", "ignores": []any{"assert.deepStrictEqual"}}},
			},
			// assert: upstream valid 20.
			{Code: "require('assert').notDeepStrictEqual()",
				Options: []any{map[string]any{"version": "3.9.9", "ignores": []any{"assert.notDeepStrictEqual"}}},
			},
			// assert: upstream valid 21.
			{Code: "require('assert').rejects()",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"assert.rejects"}}},
			},
			// assert: upstream valid 22.
			{Code: "require('assert').doesNotReject()",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"assert.doesNotReject"}}},
			},
			// assert: upstream valid 23.
			{Code: "require('assert').strict.rejects()",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"assert.strict.rejects"}}},
			},
			// assert: upstream valid 24.
			{Code: "require('assert').strict.doesNotReject()",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"assert.strict.doesNotReject"}}},
			},
			// assert: upstream valid 25.
			{Code: "var assert = require('assert').strict",
				Options: []any{map[string]any{"version": "9.8.9", "ignores": []any{"assert.strict"}}},
			},
			// assert: upstream valid 26.
			{Code: "var {strict: assert} = require('assert'); assert.rejects()",
				Options: []any{map[string]any{"version": "9.8.9", "ignores": []any{"assert.strict", "assert.strict.rejects"}}},
			},
			// assert: upstream valid 27.
			{Code: "const { CallTracker } = require('assert'); new CallTracker();",
				Options: []any{map[string]any{"version": "14.2.0", "ignores": []any{"assert.CallTracker"}}},
			},
			// assert: upstream valid 28.
			{Code: "import { CallTracker } from 'assert'; new CallTracker();",
				Options: []any{map[string]any{"version": "14.2.0", "ignores": []any{"assert.CallTracker"}}},
			},
			// assert: upstream valid 29.
			{Code: "const assert = require('node:assert'); assert.deepStrictEqual()",
				Options: []any{map[string]any{"version": "12.20.0"}},
			},
			// assert: upstream valid 30.
			{Code: "import assert from 'node:assert'; assert.deepStrictEqual()",
				Options: []any{map[string]any{"version": "14.13.1"}},
			},
			// assert: upstream valid 31.
			{Code: "require('node:assert').match()",
				Options: []any{map[string]any{"version": "15.0.0", "ignores": []any{"assert.match"}}},
			},
			// assert: upstream valid 32.
			{Code: "new Buffer(123)",
				Options: []any{map[string]any{"version": "6.0.0"}},
			},
			// assert: upstream valid 33.
			{Code: "require('tls').DEFAULT_CIPHERS",
				Options: []any{map[string]any{"version": "18.0.0"}},
			},
			// async_hooks: upstream valid 34.
			{Code: "require('async_hooks')",
				Options: []any{map[string]any{"version": "16.4.0"}},
			},
			// async_hooks: upstream valid 35.
			{Code: "import hooks from 'async_hooks'",
				Options: []any{map[string]any{"version": "16.4.0"}},
			},
			// async_hooks: upstream valid 36.
			{Code: "const { AsyncLocalStorage } = require('async_hooks'); new AsyncLocalStorage();",
				Options: []any{map[string]any{"version": "16.4.0"}},
			},
			// async_hooks: upstream valid 37.
			{Code: "import hooks from 'async_hooks'; new hooks.AsyncLocalStorage();",
				Options: []any{map[string]any{"version": "16.4.0"}},
			},
			// async_hooks: upstream valid 38.
			{Code: "require('async_hooks')",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 39.
			{Code: "import hooks from 'async_hooks'",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 40.
			{Code: "require('async_hooks').createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 41.
			{Code: "const hooks = require('async_hooks'); hooks.createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 42.
			{Code: "const { createHook } = require('async_hooks'); createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 43.
			{Code: "import * as hooks from 'async_hooks'; hooks.createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 44.
			{Code: "import hooks from 'async_hooks'; hooks.createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 45.
			{Code: "import { createHook } from 'async_hooks'; createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 46.
			{Code: "new require('async_hooks').AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 47.
			{Code: "const hooks = require('async_hooks'); new hooks.AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 48.
			{Code: "const { AsyncLocalStorage } = require('async_hooks'); new AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 49.
			{Code: "import * as hooks from 'async_hooks'; new hooks.AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 50.
			{Code: "import hooks from 'async_hooks'; new hooks.AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 51.
			{Code: "import { AsyncLocalStorage } from 'async_hooks'; new AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 52.
			{Code: "require('node:async_hooks').createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 53.
			{Code: "const hooks = require('node:async_hooks'); hooks.createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 54.
			{Code: "const { createHook } = require('node:async_hooks'); createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 55.
			{Code: "import * as hooks from 'node:async_hooks'; hooks.createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 56.
			{Code: "import hooks from 'node:async_hooks'; hooks.createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 57.
			{Code: "import { createHook } from 'node:async_hooks'; createHook()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 58.
			{Code: "new require('node:async_hooks').AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 59.
			{Code: "const hooks = require('node:async_hooks'); new hooks.AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 60.
			{Code: "const { AsyncLocalStorage } = require('node:async_hooks'); new AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 61.
			{Code: "import * as hooks from 'node:async_hooks'; new hooks.AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 62.
			{Code: "import hooks from 'node:async_hooks'; new hooks.AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 63.
			{Code: "import { AsyncLocalStorage } from 'node:async_hooks'; new AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "14.0.0", "ignores": []any{"async_hooks", "async_hooks.createHook", "async_hooks.AsyncLocalStorage"}}},
			},
			// async_hooks: upstream valid 64.
			{Code: "require('node:async_hooks')",
				Options: []any{map[string]any{"version": "16.4.0"}},
			},
			// async_hooks: upstream valid 65.
			{Code: "import hooks from 'node:async_hooks'",
				Options: []any{map[string]any{"version": "16.4.0"}},
			},
			// async_hooks: upstream valid 66.
			{Code: "require('node:async_hooks')",
				Options: []any{map[string]any{"version": "12.0.0", "ignores": []any{"async_hooks"}}},
			},
			// async_hooks: upstream valid 67.
			{Code: "import hooks from 'node:async_hooks'",
				Options: []any{map[string]any{"version": "12.0.0", "ignores": []any{"async_hooks"}}},
			},
			// buffer: upstream valid 68.
			{Code: "Buffer.alloc",
				Options: []any{map[string]any{"version": "4.5.0"}},
			},
			// buffer: upstream valid 69.
			{Code: "Buffer.allocUnsafe",
				Options: []any{map[string]any{"version": "4.5.0"}},
			},
			// buffer: upstream valid 70.
			{Code: "Buffer.allocUnsafeSlow",
				Options: []any{map[string]any{"version": "4.5.0"}},
			},
			// buffer: upstream valid 71.
			{Code: "Buffer.from",
				Options: []any{map[string]any{"version": "4.5.0"}},
			},
			// buffer: upstream valid 72.
			{Code: "require('buffer').constants",
				Options: []any{map[string]any{"version": "8.2.0"}},
			},
			// buffer: upstream valid 73.
			{Code: "var cp = require('buffer'); cp.constants",
				Options: []any{map[string]any{"version": "8.2.0"}},
			},
			// buffer: upstream valid 74.
			{Code: "var { constants } = require('buffer');",
				Options: []any{map[string]any{"version": "8.2.0"}},
			},
			// buffer: upstream valid 75.
			{Code: "import cp from 'buffer'; cp.constants",
				Options: []any{map[string]any{"version": "8.2.0"}},
			},
			// buffer: upstream valid 76.
			{Code: "import { constants } from 'buffer'",
				Options: []any{map[string]any{"version": "8.2.0"}},
			},
			// buffer: upstream valid 77.
			{Code: "var {Buffer: b} = require('buffer'); b.alloc",
				Options: []any{map[string]any{"version": "4.5.0"}},
			},
			// buffer: upstream valid 78.
			{Code: "var {Buffer: b} = require('buffer'); b.allocUnsafe",
				Options: []any{map[string]any{"version": "4.5.0"}},
			},
			// buffer: upstream valid 79.
			{Code: "var {Buffer: b} = require('buffer'); b.allocUnsafeSlow",
				Options: []any{map[string]any{"version": "4.5.0"}},
			},
			// buffer: upstream valid 80.
			{Code: "var {Buffer: b} = require('buffer'); b.from",
				Options: []any{map[string]any{"version": "4.5.0"}},
			},
			// buffer: upstream valid 81.
			{Code: "require('buffer').kMaxLength",
				Options: []any{map[string]any{"version": "3.0.0"}},
			},
			// buffer: upstream valid 82.
			{Code: "require('buffer').transcode",
				Options: []any{map[string]any{"version": "7.1.0"}},
			},
			// buffer: upstream valid 83.
			{Code: "Buffer.alloc",
				Options: []any{map[string]any{"version": "4.4.9", "ignores": []any{"Buffer.alloc"}}},
			},
			// buffer: upstream valid 84.
			{Code: "Buffer.allocUnsafe",
				Options: []any{map[string]any{"version": "4.4.9", "ignores": []any{"Buffer.allocUnsafe"}}},
			},
			// buffer: upstream valid 85.
			{Code: "Buffer.allocUnsafeSlow",
				Options: []any{map[string]any{"version": "4.4.9", "ignores": []any{"Buffer.allocUnsafeSlow"}}},
			},
			// buffer: upstream valid 86.
			{Code: "Buffer.from",
				Options: []any{map[string]any{"version": "4.4.9", "ignores": []any{"Buffer.from"}}},
			},
			// buffer: upstream valid 87.
			{Code: "require('buffer').constants",
				Options: []any{map[string]any{"version": "8.1.9", "ignores": []any{"buffer.constants"}}},
			},
			// buffer: upstream valid 88.
			{Code: "var cp = require('buffer'); cp.constants",
				Options: []any{map[string]any{"version": "8.1.9", "ignores": []any{"buffer.constants"}}},
			},
			// buffer: upstream valid 89.
			{Code: "var { constants } = require('buffer');",
				Options: []any{map[string]any{"version": "8.1.9", "ignores": []any{"buffer.constants"}}},
			},
			// buffer: upstream valid 90.
			{Code: "import cp from 'buffer'; cp.constants",
				Options: []any{map[string]any{"version": "8.1.9", "ignores": []any{"buffer.constants"}}},
			},
			// buffer: upstream valid 91.
			{Code: "import { constants } from 'buffer'",
				Options: []any{map[string]any{"version": "8.1.9", "ignores": []any{"buffer.constants"}}},
			},
			// buffer: upstream valid 92.
			{Code: "var {Buffer: b} = require('buffer'); b.alloc",
				Options: []any{map[string]any{"version": "4.4.9", "ignores": []any{"buffer.Buffer.alloc"}}},
			},
			// buffer: upstream valid 93.
			{Code: "var {Buffer: b} = require('buffer'); b.allocUnsafe",
				Options: []any{map[string]any{"version": "4.4.9", "ignores": []any{"buffer.Buffer.allocUnsafe"}}},
			},
			// buffer: upstream valid 94.
			{Code: "var {Buffer: b} = require('buffer'); b.allocUnsafeSlow",
				Options: []any{map[string]any{"version": "4.4.9", "ignores": []any{"buffer.Buffer.allocUnsafeSlow"}}},
			},
			// buffer: upstream valid 95.
			{Code: "var {Buffer: b} = require('buffer'); b.from",
				Options: []any{map[string]any{"version": "4.4.9", "ignores": []any{"buffer.Buffer.from"}}},
			},
			// buffer: upstream valid 96.
			{Code: "require('buffer').kMaxLength",
				Options: []any{map[string]any{"version": "2.9.9", "ignores": []any{"buffer.kMaxLength"}}},
			},
			// buffer: upstream valid 97.
			{Code: "require('buffer').transcode",
				Options: []any{map[string]any{"version": "7.0.9", "ignores": []any{"buffer.transcode"}}},
			},
			// buffer: upstream valid 98.
			{Code: "const { Blob } = require('buffer'); new Blob();",
				Options: []any{map[string]any{"version": "15.7.0", "ignores": []any{"buffer.Blob"}}},
			},
			// buffer: upstream valid 99.
			{Code: "import buffer from 'buffer'; new buffer.Blob();",
				Options: []any{map[string]any{"version": "15.7.0", "ignores": []any{"buffer.Blob"}}},
			},
			// child_process: upstream valid 100.
			{Code: "require('child_process').ChildProcess",
				Options: []any{map[string]any{"version": "2.2.0"}},
			},
			// child_process: upstream valid 101.
			{Code: "var cp = require('child_process'); cp.ChildProcess",
				Options: []any{map[string]any{"version": "2.2.0"}},
			},
			// child_process: upstream valid 102.
			{Code: "var { ChildProcess } = require('child_process'); ChildProcess",
				Options: []any{map[string]any{"version": "2.2.0"}},
			},
			// child_process: upstream valid 103.
			{Code: "import cp from 'child_process'; cp.ChildProcess",
				Options: []any{map[string]any{"version": "2.2.0"}},
			},
			// child_process: upstream valid 104.
			{Code: "import { ChildProcess } from 'child_process'",
				Options: []any{map[string]any{"version": "2.2.0"}},
			},
			// child_process: upstream valid 105.
			{Code: "require('child_process').ChildProcess",
				Options: []any{map[string]any{"version": "2.1.9", "ignores": []any{"child_process.ChildProcess"}}},
			},
			// child_process: upstream valid 106.
			{Code: "var cp = require('child_process'); cp.ChildProcess",
				Options: []any{map[string]any{"version": "2.1.9", "ignores": []any{"child_process.ChildProcess"}}},
			},
			// child_process: upstream valid 107.
			{Code: "var { ChildProcess } = require('child_process'); ChildProcess",
				Options: []any{map[string]any{"version": "2.1.9", "ignores": []any{"child_process.ChildProcess"}}},
			},
			// child_process: upstream valid 108.
			{Code: "import cp from 'child_process'; cp.ChildProcess",
				Options: []any{map[string]any{"version": "2.1.9", "ignores": []any{"child_process.ChildProcess"}}},
			},
			// child_process: upstream valid 109.
			{Code: "import { ChildProcess } from 'child_process'",
				Options: []any{map[string]any{"version": "2.1.9", "ignores": []any{"child_process.ChildProcess"}}},
			},
			// console: upstream valid 110.
			{Code: "console.clear()",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// console: upstream valid 111.
			{Code: "require('console').clear()",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// console: upstream valid 112.
			{Code: "var c = require('console'); c.clear()",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// console: upstream valid 113.
			{Code: "var { clear } = require('console'); clear()",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// console: upstream valid 114.
			{Code: "import c from 'console'; c.clear()",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// console: upstream valid 115.
			{Code: "console.count()",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// console: upstream valid 116.
			{Code: "console.countReset()",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// console: upstream valid 117.
			{Code: "console.debug()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// console: upstream valid 118.
			{Code: "console.dirxml()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// console: upstream valid 119.
			{Code: "console.group()",
				Options: []any{map[string]any{"version": "8.5.0"}},
			},
			// console: upstream valid 120.
			{Code: "console.groupCollapsed()",
				Options: []any{map[string]any{"version": "8.5.0"}},
			},
			// console: upstream valid 121.
			{Code: "console.groupEnd()",
				Options: []any{map[string]any{"version": "8.5.0"}},
			},
			// console: upstream valid 122.
			{Code: "console.table()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// console: upstream valid 123.
			{Code: "console.markTimeline()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// console: upstream valid 124.
			{Code: "console.profile()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// console: upstream valid 125.
			{Code: "console.profileEnd()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// console: upstream valid 126.
			{Code: "console.timeStamp()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// console: upstream valid 127.
			{Code: "console.timeline()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// console: upstream valid 128.
			{Code: "console.timelineEnd()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// console: upstream valid 129.
			{Code: "console.clear()",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"console.clear"}}},
			},
			// console: upstream valid 130.
			{Code: "require('console').clear()",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"console.clear"}}},
			},
			// console: upstream valid 131.
			{Code: "var c = require('console'); c.clear()",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"console.clear"}}},
			},
			// console: upstream valid 132.
			{Code: "var { clear } = require('console'); clear()",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"console.clear"}}},
			},
			// console: upstream valid 133.
			{Code: "import c from 'console'; c.clear()",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"console.clear"}}},
			},
			// console: upstream valid 134.
			{Code: "console.count()",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"console.count"}}},
			},
			// console: upstream valid 135.
			{Code: "console.countReset()",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"console.countReset"}}},
			},
			// console: upstream valid 136.
			{Code: "console.debug()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"console.debug"}}},
			},
			// console: upstream valid 137.
			{Code: "console.dirxml()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"console.dirxml"}}},
			},
			// console: upstream valid 138.
			{Code: "console.group()",
				Options: []any{map[string]any{"version": "8.4.9", "ignores": []any{"console.group"}}},
			},
			// console: upstream valid 139.
			{Code: "console.groupCollapsed()",
				Options: []any{map[string]any{"version": "8.4.9", "ignores": []any{"console.groupCollapsed"}}},
			},
			// console: upstream valid 140.
			{Code: "console.groupEnd()",
				Options: []any{map[string]any{"version": "8.4.9", "ignores": []any{"console.groupEnd"}}},
			},
			// console: upstream valid 141.
			{Code: "console.table()",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"console.table"}}},
			},
			// console: upstream valid 142.
			{Code: "console.profile()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"console.profile"}}},
			},
			// console: upstream valid 143.
			{Code: "console.profileEnd()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"console.profileEnd"}}},
			},
			// console: upstream valid 144.
			{Code: "console.timeStamp()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"console.timeStamp"}}},
			},
			// crypto: upstream valid 145.
			{Code: "require('crypto').constants",
				Options: []any{map[string]any{"version": "6.3.0"}},
			},
			// crypto: upstream valid 146.
			{Code: "var hooks = require('crypto'); hooks.constants",
				Options: []any{map[string]any{"version": "6.3.0"}},
			},
			// crypto: upstream valid 147.
			{Code: "var { constants } = require('crypto'); constants",
				Options: []any{map[string]any{"version": "6.3.0"}},
			},
			// crypto: upstream valid 148.
			{Code: "import crypto from 'crypto'; crypto.constants",
				Options: []any{map[string]any{"version": "6.3.0"}},
			},
			// crypto: upstream valid 149.
			{Code: "import { constants } from 'crypto'; constants",
				Options: []any{map[string]any{"version": "6.3.0"}},
			},
			// crypto: upstream valid 150.
			{Code: "require('crypto').Certificate.exportChallenge()",
				Options: []any{map[string]any{"version": "9.0.0"}},
			},
			// crypto: upstream valid 151.
			{Code: "var { Certificate: c } = require('crypto'); c.exportChallenge()",
				Options: []any{map[string]any{"version": "9.0.0"}},
			},
			// crypto: upstream valid 152.
			{Code: "var { Certificate: c } = require('crypto'); c.exportPublicKey()",
				Options: []any{map[string]any{"version": "9.0.0"}},
			},
			// crypto: upstream valid 153.
			{Code: "var { Certificate: c } = require('crypto'); c.verifySpkac()",
				Options: []any{map[string]any{"version": "9.0.0"}},
			},
			// crypto: upstream valid 154.
			{Code: "require('crypto').fips",
				Options: []any{map[string]any{"version": "6.0.0"}},
			},
			// crypto: upstream valid 155.
			{Code: "require('crypto').getCurves",
				Options: []any{map[string]any{"version": "2.3.0"}},
			},
			// crypto: upstream valid 156.
			{Code: "require('crypto').getFips",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// crypto: upstream valid 157.
			{Code: "require('crypto').privateEncrypt",
				Options: []any{map[string]any{"version": "1.1.0"}},
			},
			// crypto: upstream valid 158.
			{Code: "require('crypto').publicDecrypt",
				Options: []any{map[string]any{"version": "1.1.0"}},
			},
			// crypto: upstream valid 159.
			{Code: "require('crypto').randomFillSync",
				Options: []any{map[string]any{"version": "7.10.0"}},
			},
			// crypto: upstream valid 160.
			{Code: "require('crypto').randomFill",
				Options: []any{map[string]any{"version": "7.10.0"}},
			},
			// crypto: upstream valid 161.
			{Code: "require('crypto').scrypt",
				Options: []any{map[string]any{"version": "10.5.0"}},
			},
			// crypto: upstream valid 162.
			{Code: "require('crypto').scryptSync",
				Options: []any{map[string]any{"version": "10.5.0"}},
			},
			// crypto: upstream valid 163.
			{Code: "require('crypto').setFips",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// crypto: upstream valid 164.
			{Code: "require('crypto').timingSafeEqual",
				Options: []any{map[string]any{"version": "6.6.0"}},
			},
			// crypto: upstream valid 165.
			{Code: "require('crypto').constants",
				Options: []any{map[string]any{"version": "6.2.9", "ignores": []any{"crypto.constants"}}},
			},
			// crypto: upstream valid 166.
			{Code: "var hooks = require('crypto'); hooks.constants",
				Options: []any{map[string]any{"version": "6.2.9", "ignores": []any{"crypto.constants"}}},
			},
			// crypto: upstream valid 167.
			{Code: "var { constants } = require('crypto'); constants",
				Options: []any{map[string]any{"version": "6.2.9", "ignores": []any{"crypto.constants"}}},
			},
			// crypto: upstream valid 168.
			{Code: "import crypto from 'crypto'; crypto.constants",
				Options: []any{map[string]any{"version": "6.2.9", "ignores": []any{"crypto.constants"}}},
			},
			// crypto: upstream valid 169.
			{Code: "import { constants } from 'crypto'; constants",
				Options: []any{map[string]any{"version": "6.2.9", "ignores": []any{"crypto.constants"}}},
			},
			// crypto: upstream valid 170.
			{Code: "require('crypto').Certificate.exportChallenge()",
				Options: []any{map[string]any{"version": "8.9.9", "ignores": []any{"crypto.Certificate.exportChallenge"}}},
			},
			// crypto: upstream valid 171.
			{Code: "var { Certificate: c } = require('crypto'); c.exportChallenge()",
				Options: []any{map[string]any{"version": "8.9.9", "ignores": []any{"crypto.Certificate.exportChallenge"}}},
			},
			// crypto: upstream valid 172.
			{Code: "var { Certificate: c } = require('crypto'); c.exportPublicKey()",
				Options: []any{map[string]any{"version": "8.9.9", "ignores": []any{"crypto.Certificate.exportPublicKey"}}},
			},
			// crypto: upstream valid 173.
			{Code: "var { Certificate: c } = require('crypto'); c.verifySpkac()",
				Options: []any{map[string]any{"version": "8.9.9", "ignores": []any{"crypto.Certificate.verifySpkac"}}},
			},
			// crypto: upstream valid 174.
			{Code: "require('crypto').fips",
				Options: []any{map[string]any{"version": "5.9.9", "ignores": []any{"crypto.fips"}}},
			},
			// crypto: upstream valid 175.
			{Code: "require('crypto').getCurves",
				Options: []any{map[string]any{"version": "2.2.9", "ignores": []any{"crypto.getCurves"}}},
			},
			// crypto: upstream valid 176.
			{Code: "require('crypto').getFips",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"crypto.getFips"}}},
			},
			// crypto: upstream valid 177.
			{Code: "require('crypto').privateEncrypt",
				Options: []any{map[string]any{"version": "1.0.9", "ignores": []any{"crypto.privateEncrypt"}}},
			},
			// crypto: upstream valid 178.
			{Code: "require('crypto').publicDecrypt",
				Options: []any{map[string]any{"version": "1.0.9", "ignores": []any{"crypto.publicDecrypt"}}},
			},
			// crypto: upstream valid 179.
			{Code: "require('crypto').randomFillSync",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"crypto.randomFillSync"}}},
			},
			// crypto: upstream valid 180.
			{Code: "require('crypto').randomFill",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"crypto.randomFill"}}},
			},
			// crypto: upstream valid 181.
			{Code: "require('crypto').scrypt",
				Options: []any{map[string]any{"version": "10.4.9", "ignores": []any{"crypto.scrypt"}}},
			},
			// crypto: upstream valid 182.
			{Code: "require('crypto').scryptSync",
				Options: []any{map[string]any{"version": "10.4.9", "ignores": []any{"crypto.scryptSync"}}},
			},
			// crypto: upstream valid 183.
			{Code: "require('crypto').setFips",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"crypto.setFips"}}},
			},
			// crypto: upstream valid 184.
			{Code: "require('crypto').timingSafeEqual",
				Options: []any{map[string]any{"version": "6.5.9", "ignores": []any{"crypto.timingSafeEqual"}}},
			},
			// dns: upstream valid 185.
			{Code: "require('dns').Resolver",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// dns: upstream valid 186.
			{Code: "var hooks = require('dns'); hooks.Resolver",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// dns: upstream valid 187.
			{Code: "var { Resolver } = require('dns'); Resolver",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// dns: upstream valid 188.
			{Code: "import dns from 'dns'; dns.Resolver",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// dns: upstream valid 189.
			{Code: "import { Resolver } from 'dns'; Resolver",
				Options: []any{map[string]any{"version": "8.3.0"}},
			},
			// dns: upstream valid 190.
			{Code: "require('dns').resolvePtr",
				Options: []any{map[string]any{"version": "6.0.0"}},
			},
			// dns: upstream valid 191.
			{Code: "require('dns').promises",
				Options: []any{map[string]any{"version": "11.14.0"}},
			},
			// dns: upstream valid 192.
			{Code: "require('dns').Resolver",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"dns.Resolver"}}},
			},
			// dns: upstream valid 193.
			{Code: "var hooks = require('dns'); hooks.Resolver",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"dns.Resolver"}}},
			},
			// dns: upstream valid 194.
			{Code: "var { Resolver } = require('dns'); Resolver",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"dns.Resolver"}}},
			},
			// dns: upstream valid 195.
			{Code: "import dns from 'dns'; dns.Resolver",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"dns.Resolver"}}},
			},
			// dns: upstream valid 196.
			{Code: "import { Resolver } from 'dns'; Resolver",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"dns.Resolver"}}},
			},
			// dns: upstream valid 197.
			{Code: "require('dns').resolvePtr",
				Options: []any{map[string]any{"version": "5.9.9", "ignores": []any{"dns.resolvePtr"}}},
			},
			// dns: upstream valid 198.
			{Code: "require('dns').promises",
				Options: []any{map[string]any{"version": "11.13.9", "ignores": []any{"dns.promises"}}},
			},
			// fs: upstream valid 199.
			{Code: "require('fs').promises",
				Options: []any{map[string]any{"version": "11.14.0"}},
			},
			// fs: upstream valid 200.
			{Code: "var fs = require('fs'); fs.promises",
				Options: []any{map[string]any{"version": "11.14.0"}},
			},
			// fs: upstream valid 201.
			{Code: "var { promises } = require('fs'); promises",
				Options: []any{map[string]any{"version": "11.14.0"}},
			},
			// fs: upstream valid 202.
			{Code: "import fs from 'fs'; fs.promises",
				Options: []any{map[string]any{"version": "11.14.0"}},
			},
			// fs: upstream valid 203.
			{Code: "import { promises } from 'fs'",
				Options: []any{map[string]any{"version": "11.14.0"}},
			},
			// fs: upstream valid 204.
			{Code: "require('fs').copyFile",
				Options: []any{map[string]any{"version": "8.5.0"}},
			},
			// fs: upstream valid 205.
			{Code: "require('fs').copyFileSync",
				Options: []any{map[string]any{"version": "8.5.0"}},
			},
			// fs: upstream valid 206.
			{Code: "require('fs').mkdtemp",
				Options: []any{map[string]any{"version": "5.10.0"}},
			},
			// fs: upstream valid 207.
			{Code: "require('fs').mkdtempSync",
				Options: []any{map[string]any{"version": "5.10.0"}},
			},
			// fs: upstream valid 208.
			{Code: "require('fs').realpath.native",
				Options: []any{map[string]any{"version": "9.2.0"}},
			},
			// fs: upstream valid 209.
			{Code: "require('fs').realpathSync.native",
				Options: []any{map[string]any{"version": "9.2.0"}},
			},
			// fs: upstream valid 210.
			{Code: "require('fs').promises",
				Options: []any{map[string]any{"version": "11.13.9", "ignores": []any{"fs.promises"}}},
			},
			// fs: upstream valid 211.
			{Code: "var fs = require('fs'); fs.promises",
				Options: []any{map[string]any{"version": "11.13.9", "ignores": []any{"fs.promises"}}},
			},
			// fs: upstream valid 212.
			{Code: "var { promises } = require('fs'); promises",
				Options: []any{map[string]any{"version": "11.13.9", "ignores": []any{"fs.promises"}}},
			},
			// fs: upstream valid 213.
			{Code: "import fs from 'fs'; fs.promises",
				Options: []any{map[string]any{"version": "11.13.9", "ignores": []any{"fs.promises"}}},
			},
			// fs: upstream valid 214.
			{Code: "import { promises } from 'fs'",
				Options: []any{map[string]any{"version": "11.13.9", "ignores": []any{"fs.promises"}}},
			},
			// fs: upstream valid 215.
			{Code: "require('fs').copyFile",
				Options: []any{map[string]any{"version": "8.4.9", "ignores": []any{"fs.copyFile"}}},
			},
			// fs: upstream valid 216.
			{Code: "require('fs').copyFileSync",
				Options: []any{map[string]any{"version": "8.4.9", "ignores": []any{"fs.copyFileSync"}}},
			},
			// fs: upstream valid 217.
			{Code: "require('fs').mkdtemp",
				Options: []any{map[string]any{"version": "5.9.9", "ignores": []any{"fs.mkdtemp"}}},
			},
			// fs: upstream valid 218.
			{Code: "require('fs').mkdtempSync",
				Options: []any{map[string]any{"version": "5.9.9", "ignores": []any{"fs.mkdtempSync"}}},
			},
			// fs: upstream valid 219.
			{Code: "require('fs').realpath.native",
				Options: []any{map[string]any{"version": "9.1.9", "ignores": []any{"fs.realpath.native"}}},
			},
			// fs: upstream valid 220.
			{Code: "require('fs').realpathSync.native",
				Options: []any{map[string]any{"version": "9.1.9", "ignores": []any{"fs.realpathSync.native"}}},
			},
			// fs: upstream valid 221.
			{Code: "require('fs').readv",
				Options: []any{map[string]any{"version": "13.13.0"}},
			},
			// fs: upstream valid 222.
			{Code: "require('fs').readvSync",
				Options: []any{map[string]any{"version": "13.13.0"}},
			},
			// fs: upstream valid 223.
			{Code: "require('fs').readv",
				Options: []any{map[string]any{"version": "12.17.0"}},
			},
			// fs: upstream valid 224.
			{Code: "require('fs').readvSync",
				Options: []any{map[string]any{"version": "12.17.0"}},
			},
			// fs: upstream valid 225.
			{Code: "require('fs').readv",
				Options: []any{map[string]any{"version": "13.12.0", "ignores": []any{"fs.readv"}}},
			},
			// fs: upstream valid 226.
			{Code: "require('fs').readvSync",
				Options: []any{map[string]any{"version": "13.12.0", "ignores": []any{"fs.readvSync"}}},
			},
			// fs: upstream valid 227.
			{Code: "require('fs').lutimes",
				Options: []any{map[string]any{"version": "14.5.0"}},
			},
			// fs: upstream valid 228.
			{Code: "require('fs').lutimesSync",
				Options: []any{map[string]any{"version": "14.5.0"}},
			},
			// fs: upstream valid 229.
			{Code: "require('fs').lutimes",
				Options: []any{map[string]any{"version": "12.19.0"}},
			},
			// fs: upstream valid 230.
			{Code: "require('fs').lutimesSync",
				Options: []any{map[string]any{"version": "12.19.0"}},
			},
			// fs: upstream valid 231.
			{Code: "require('fs').lutimes",
				Options: []any{map[string]any{"version": "13.14.0", "ignores": []any{"fs.lutimes"}}},
			},
			// fs: upstream valid 232.
			{Code: "require('fs').lutimesSync",
				Options: []any{map[string]any{"version": "13.14.0", "ignores": []any{"fs.lutimesSync"}}},
			},
			// fs: upstream valid 233.
			{Code: "require('fs').opendir",
				Options: []any{map[string]any{"version": "12.12.0"}},
			},
			// fs: upstream valid 234.
			{Code: "require('fs').opendirSync",
				Options: []any{map[string]any{"version": "12.12.0"}},
			},
			// fs: upstream valid 235.
			{Code: "require('fs').opendir",
				Options: []any{map[string]any{"version": "12.11.0", "ignores": []any{"fs.opendir"}}},
			},
			// fs: upstream valid 236.
			{Code: "require('fs').opendirSync",
				Options: []any{map[string]any{"version": "12.11.0", "ignores": []any{"fs.opendirSync"}}},
			},
			// fs: upstream valid 237.
			{Code: "require('fs').rm",
				Options: []any{map[string]any{"version": "14.14.0"}},
			},
			// fs: upstream valid 238.
			{Code: "require('fs').rmSync",
				Options: []any{map[string]any{"version": "14.14.0"}},
			},
			// fs: upstream valid 239.
			{Code: "require('fs').rm",
				Options: []any{map[string]any{"version": "14.13.0", "ignores": []any{"fs.rm"}}},
			},
			// fs: upstream valid 240.
			{Code: "require('fs').rmSync",
				Options: []any{map[string]any{"version": "14.13.0", "ignores": []any{"fs.rmSync"}}},
			},
			// fs: upstream valid 241.
			{Code: "require('fs').read",
				Options: []any{map[string]any{"version": "13.11.0"}},
			},
			// fs: upstream valid 242.
			{Code: "require('fs').readSync",
				Options: []any{map[string]any{"version": "13.11.0"}},
			},
			// fs: upstream valid 243.
			{Code: "require('fs').read",
				Options: []any{map[string]any{"version": "12.17.0"}},
			},
			// fs: upstream valid 244.
			{Code: "require('fs').readSync",
				Options: []any{map[string]any{"version": "12.17.0"}},
			},
			// fs: upstream valid 245.
			{Code: "require('fs').read",
				Options: []any{map[string]any{"version": "13.10.0", "ignores": []any{"fs.read"}}},
			},
			// fs: upstream valid 246.
			{Code: "require('fs').readSync",
				Options: []any{map[string]any{"version": "13.10.0", "ignores": []any{"fs.readSync"}}},
			},
			// fs: upstream valid 247.
			{Code: "require('fs').Dir",
				Options: []any{map[string]any{"version": "12.12.0"}},
			},
			// fs: upstream valid 248.
			{Code: "require('fs').Dir",
				Options: []any{map[string]any{"version": "12.11.0", "ignores": []any{"fs.Dir"}}},
			},
			// fs: upstream valid 249.
			{Code: "require('fs').StatWatcher",
				Options: []any{map[string]any{"version": "14.3.0"}},
			},
			// fs: upstream valid 250.
			{Code: "require('fs').StatWatcher",
				Options: []any{map[string]any{"version": "14.2.0", "ignores": []any{"fs.StatWatcher"}}},
			},
			// fs: upstream valid 251.
			{Code: "require('fs').StatWatcher",
				Options: []any{map[string]any{"version": "12.20.0"}},
			},
			// fs: upstream valid 252.
			{Code: "require('fs').StatWatcher",
				Options: []any{map[string]any{"version": "12.19.0", "ignores": []any{"fs.StatWatcher"}}},
			},
			// fs/promises: upstream valid 253.
			{Code: "import * as fs from 'fs/promises';",
				Options: []any{map[string]any{"version": "14.0.0"}},
			},
			// fs/promises: upstream valid 254.
			{Code: "require('fs/promise')",
				Options: []any{map[string]any{"version": "14.0.0"}},
			},
			// fs/promises: upstream valid 255.
			{Code: "import * as fs from 'node:fs/promises';",
				Options: []any{map[string]any{"version": "14.13.1"}},
			},
			// fs/promises: upstream valid 256.
			{Code: "require('node:fs/promise')",
				Options: []any{map[string]any{"version": "14.13.1"}},
			},
			// fs/promises: upstream valid 257.
			{Code: "import * as fs from 'fs/promises';",
				Options: []any{map[string]any{"version": "13.14.0", "ignores": []any{"fs/promises"}}},
			},
			// fs/promises: upstream valid 258.
			{Code: "import * as fs from 'node:fs/promises';",
				Options: []any{map[string]any{"version": "13.14.0", "ignores": []any{"fs/promises"}}},
			},
			// fs/promises: upstream valid 259.
			{Code: "require('fs/promise')",
				Options: []any{map[string]any{"version": "13.14.0", "ignores": []any{"fs/promises"}}},
			},
			// fs/promises: upstream valid 260.
			{Code: "require('node:fs/promise')",
				Options: []any{map[string]any{"version": "13.14.0", "ignores": []any{"fs/promises"}}},
			},
			// http2: upstream valid 261.
			{Code: "require('http2')",
				Options: []any{map[string]any{"version": "10.10.0"}},
			},
			// http2: upstream valid 262.
			{Code: "import http2 from 'http2'",
				Options: []any{map[string]any{"version": "10.10.0"}},
			},
			// http2: upstream valid 263.
			{Code: "require('http2')",
				Options: []any{map[string]any{"version": "8.3.9", "ignores": []any{"http2"}}},
			},
			// http2: upstream valid 264.
			{Code: "import http2 from 'http2'",
				Options: []any{map[string]any{"version": "8.3.9", "ignores": []any{"http2"}}},
			},
			// inspector: upstream valid 265.
			{Code: "require('inspector')",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"inspector"}}},
			},
			// inspector: upstream valid 266.
			{Code: "import inspector from 'inspector'",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"inspector"}}},
			},
			// inspector: upstream valid 267.
			{Code: "import { open } from 'inspector'",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"inspector", "inspector.open"}}},
			},
			// module: upstream valid 268.
			{Code: "require.resolve.paths()",
				Options: []any{map[string]any{"version": "8.9.0"}},
			},
			// module: upstream valid 269.
			{Code: "require('module').builtinModules",
				Options: []any{map[string]any{"version": "9.3.0"}},
			},
			// module: upstream valid 270.
			{Code: "require.resolve.paths()",
				Options: []any{map[string]any{"version": "8.8.9", "ignores": []any{"require.resolve.paths"}}},
			},
			// module: upstream valid 271.
			{Code: "require('module').builtinModules",
				Options: []any{map[string]any{"version": "9.2.9", "ignores": []any{"module.builtinModules"}}},
			},
			// os: upstream valid 272.
			{Code: "require('os').constants",
				Options: []any{map[string]any{"version": "6.3.0"}},
			},
			// os: upstream valid 273.
			{Code: "var hooks = require('os'); hooks.constants",
				Options: []any{map[string]any{"version": "6.3.0"}},
			},
			// os: upstream valid 274.
			{Code: "var { constants } = require('os'); constants",
				Options: []any{map[string]any{"version": "6.3.0"}},
			},
			// os: upstream valid 275.
			{Code: "import os from 'os'; os.constants",
				Options: []any{map[string]any{"version": "6.3.0"}},
			},
			// os: upstream valid 276.
			{Code: "import { constants } from 'os'; constants",
				Options: []any{map[string]any{"version": "6.3.0"}},
			},
			// os: upstream valid 277.
			{Code: "require('os').homedir",
				Options: []any{map[string]any{"version": "2.3.0"}},
			},
			// os: upstream valid 278.
			{Code: "require('os').userInfo",
				Options: []any{map[string]any{"version": "6.0.0"}},
			},
			// os: upstream valid 279.
			{Code: "require('os').constants",
				Options: []any{map[string]any{"version": "6.2.9", "ignores": []any{"os.constants"}}},
			},
			// os: upstream valid 280.
			{Code: "var hooks = require('os'); hooks.constants",
				Options: []any{map[string]any{"version": "6.2.9", "ignores": []any{"os.constants"}}},
			},
			// os: upstream valid 281.
			{Code: "var { constants } = require('os'); constants",
				Options: []any{map[string]any{"version": "6.2.9", "ignores": []any{"os.constants"}}},
			},
			// os: upstream valid 282.
			{Code: "import os from 'os'; os.constants",
				Options: []any{map[string]any{"version": "6.2.9", "ignores": []any{"os.constants"}}},
			},
			// os: upstream valid 283.
			{Code: "import { constants } from 'os'; constants",
				Options: []any{map[string]any{"version": "6.2.9", "ignores": []any{"os.constants"}}},
			},
			// os: upstream valid 284.
			{Code: "require('os').homedir",
				Options: []any{map[string]any{"version": "2.2.9", "ignores": []any{"os.homedir"}}},
			},
			// os: upstream valid 285.
			{Code: "require('os').userInfo",
				Options: []any{map[string]any{"version": "5.9.9", "ignores": []any{"os.userInfo"}}},
			},
			// path: upstream valid 286.
			{Code: "require('path').toNamespacedPath()",
				Options: []any{map[string]any{"version": "9.0.0"}},
			},
			// path: upstream valid 287.
			{Code: "var path = require('path'); path.toNamespacedPath()",
				Options: []any{map[string]any{"version": "9.0.0"}},
			},
			// path: upstream valid 288.
			{Code: "var { toNamespacedPath } = require('path'); toNamespacedPath()",
				Options: []any{map[string]any{"version": "9.0.0"}},
			},
			// path: upstream valid 289.
			{Code: "import path from 'path'; path.toNamespacedPath()",
				Options: []any{map[string]any{"version": "9.0.0"}},
			},
			// path: upstream valid 290.
			{Code: "import { toNamespacedPath } from 'path'; toNamespacedPath()",
				Options: []any{map[string]any{"version": "9.0.0"}},
			},
			// path: upstream valid 291.
			{Code: "require('path').toNamespacedPath()",
				Options: []any{map[string]any{"version": "8.9.9", "ignores": []any{"path.toNamespacedPath"}}},
			},
			// path: upstream valid 292.
			{Code: "var path = require('path'); path.toNamespacedPath()",
				Options: []any{map[string]any{"version": "8.9.9", "ignores": []any{"path.toNamespacedPath"}}},
			},
			// path: upstream valid 293.
			{Code: "var { toNamespacedPath } = require('path'); toNamespacedPath()",
				Options: []any{map[string]any{"version": "8.9.9", "ignores": []any{"path.toNamespacedPath"}}},
			},
			// path: upstream valid 294.
			{Code: "import path from 'path'; path.toNamespacedPath()",
				Options: []any{map[string]any{"version": "8.9.9", "ignores": []any{"path.toNamespacedPath"}}},
			},
			// path: upstream valid 295.
			{Code: "import { toNamespacedPath } from 'path'; toNamespacedPath()",
				Options: []any{map[string]any{"version": "8.9.9", "ignores": []any{"path.toNamespacedPath"}}},
			},
			// perf_hooks: upstream valid 296.
			{Code: "require('perf_hooks')",
				Options: []any{map[string]any{"version": "8.5.0"}},
			},
			// perf_hooks: upstream valid 297.
			{Code: "import perf_hooks from 'perf_hooks'",
				Options: []any{map[string]any{"version": "8.5.0"}},
			},
			// perf_hooks: upstream valid 298.
			{Code: "require('perf_hooks')",
				Options: []any{map[string]any{"version": "8.4.9", "ignores": []any{"perf_hooks"}}},
			},
			// perf_hooks: upstream valid 299.
			{Code: "import perf_hooks from 'perf_hooks'",
				Options: []any{map[string]any{"version": "8.4.9", "ignores": []any{"perf_hooks"}}},
			},
			// process: upstream valid 300.
			{Code: "process.argv0",
				Options: []any{map[string]any{"version": "6.4.0"}},
			},
			// process: upstream valid 301.
			{Code: "require('process').argv0",
				Options: []any{map[string]any{"version": "6.4.0"}},
			},
			// process: upstream valid 302.
			{Code: "var c = require('process'); c.argv0",
				Options: []any{map[string]any{"version": "6.4.0"}},
			},
			// process: upstream valid 303.
			{Code: "var { argv0 } = require('process'); argv0",
				Options: []any{map[string]any{"version": "6.4.0"}},
			},
			// process: upstream valid 304.
			{Code: "import c from 'process'; c.argv0",
				Options: []any{map[string]any{"version": "6.4.0"}},
			},
			// process: upstream valid 305.
			{Code: "process.channel",
				Options: []any{map[string]any{"version": "7.1.0"}},
			},
			// process: upstream valid 306.
			{Code: "process.cpuUsage",
				Options: []any{map[string]any{"version": "6.1.0"}},
			},
			// process: upstream valid 307.
			{Code: "process.emitWarning",
				Options: []any{map[string]any{"version": "6.0.0"}},
			},
			// process: upstream valid 308.
			{Code: "process.getegid",
				Options: []any{map[string]any{"version": "2.0.0"}},
			},
			// process: upstream valid 309.
			{Code: "process.geteuid",
				Options: []any{map[string]any{"version": "2.0.0"}},
			},
			// process: upstream valid 310.
			{Code: "process.hasUncaughtExceptionCaptureCallback",
				Options: []any{map[string]any{"version": "9.3.0"}},
			},
			// process: upstream valid 311.
			{Code: "process.ppid",
				Options: []any{map[string]any{"version": "9.2.0"}},
			},
			// process: upstream valid 312.
			{Code: "process.release",
				Options: []any{map[string]any{"version": "3.0.0"}},
			},
			// process: upstream valid 313.
			{Code: "process.setegid",
				Options: []any{map[string]any{"version": "2.0.0"}},
			},
			// process: upstream valid 314.
			{Code: "process.seteuid",
				Options: []any{map[string]any{"version": "2.0.0"}},
			},
			// process: upstream valid 315.
			{Code: "process.setUncaughtExceptionCaptureCallback",
				Options: []any{map[string]any{"version": "9.3.0"}},
			},
			// process: upstream valid 316.
			{Code: "process.argv0",
				Options: []any{map[string]any{"version": "6.3.9", "ignores": []any{"process.argv0"}}},
			},
			// process: upstream valid 317.
			{Code: "require('process').argv0",
				Options: []any{map[string]any{"version": "6.3.9", "ignores": []any{"process.argv0"}}},
			},
			// process: upstream valid 318.
			{Code: "var c = require('process'); c.argv0",
				Options: []any{map[string]any{"version": "6.3.9", "ignores": []any{"process.argv0"}}},
			},
			// process: upstream valid 319.
			{Code: "var { argv0 } = require('process'); argv0",
				Options: []any{map[string]any{"version": "6.3.9", "ignores": []any{"process.argv0"}}},
			},
			// process: upstream valid 320.
			{Code: "import c from 'process'; c.argv0",
				Options: []any{map[string]any{"version": "6.3.9", "ignores": []any{"process.argv0"}}},
			},
			// process: upstream valid 321.
			{Code: "process.channel",
				Options: []any{map[string]any{"version": "7.0.9", "ignores": []any{"process.channel"}}},
			},
			// process: upstream valid 322.
			{Code: "process.cpuUsage",
				Options: []any{map[string]any{"version": "6.0.9", "ignores": []any{"process.cpuUsage"}}},
			},
			// process: upstream valid 323.
			{Code: "process.emitWarning",
				Options: []any{map[string]any{"version": "5.9.9", "ignores": []any{"process.emitWarning"}}},
			},
			// process: upstream valid 324.
			{Code: "process.getegid",
				Options: []any{map[string]any{"version": "1.9.9", "ignores": []any{"process.getegid"}}},
			},
			// process: upstream valid 325.
			{Code: "process.geteuid",
				Options: []any{map[string]any{"version": "1.9.9", "ignores": []any{"process.geteuid"}}},
			},
			// process: upstream valid 326.
			{Code: "process.hasUncaughtExceptionCaptureCallback",
				Options: []any{map[string]any{"version": "9.2.9", "ignores": []any{"process.hasUncaughtExceptionCaptureCallback"}}},
			},
			// process: upstream valid 327.
			{Code: "process.ppid",
				Options: []any{map[string]any{"version": "9.1.9", "ignores": []any{"process.ppid"}}},
			},
			// process: upstream valid 328.
			{Code: "process.release",
				Options: []any{map[string]any{"version": "2.9.9", "ignores": []any{"process.release"}}},
			},
			// process: upstream valid 329.
			{Code: "process.setegid",
				Options: []any{map[string]any{"version": "1.9.9", "ignores": []any{"process.setegid"}}},
			},
			// process: upstream valid 330.
			{Code: "process.seteuid",
				Options: []any{map[string]any{"version": "1.9.9", "ignores": []any{"process.seteuid"}}},
			},
			// process: upstream valid 331.
			{Code: "process.setUncaughtExceptionCaptureCallback",
				Options: []any{map[string]any{"version": "9.2.9", "ignores": []any{"process.setUncaughtExceptionCaptureCallback"}}},
			},
			// stream: upstream valid 332.
			{Code: "require('stream').finished()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// stream: upstream valid 333.
			{Code: "var hooks = require('stream'); hooks.finished()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// stream: upstream valid 334.
			{Code: "var { finished } = require('stream'); finished()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// stream: upstream valid 335.
			{Code: "import stream from 'stream'; stream.finished()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// stream: upstream valid 336.
			{Code: "import { finished } from 'stream'; finished()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// stream: upstream valid 337.
			{Code: "require('stream').pipeline()",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// stream: upstream valid 338.
			{Code: "require('stream').finished()",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"stream.finished"}}},
			},
			// stream: upstream valid 339.
			{Code: "var hooks = require('stream'); hooks.finished()",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"stream.finished"}}},
			},
			// stream: upstream valid 340.
			{Code: "var { finished } = require('stream'); finished()",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"stream.finished"}}},
			},
			// stream: upstream valid 341.
			{Code: "import stream from 'stream'; stream.finished()",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"stream.finished"}}},
			},
			// stream: upstream valid 342.
			{Code: "import { finished } from 'stream'; finished()",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"stream.finished"}}},
			},
			// stream: upstream valid 343.
			{Code: "require('stream').pipeline()",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"stream.pipeline"}}},
			},
			// trace_events: upstream valid 344.
			{Code: "require('trace_events')",
				Options: []any{map[string]any{"version": "10.0.0", "ignores": []any{"trace_events"}}},
			},
			// trace_events: upstream valid 345.
			{Code: "import trace_events from 'trace_events'",
				Options: []any{map[string]any{"version": "10.0.0", "ignores": []any{"trace_events"}}},
			},
			// url: upstream valid 346.
			{Code: "URL",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// url: upstream valid 347.
			{Code: "URLSearchParams",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// url: upstream valid 348.
			{Code: "require('url').URL",
				Options: []any{map[string]any{"version": "7.0.0"}},
			},
			// url: upstream valid 349.
			{Code: "require('url').URL",
				Options: []any{map[string]any{"version": "6.13.0"}},
			},
			// url: upstream valid 350.
			{Code: "var cp = require('url'); cp.URL",
				Options: []any{map[string]any{"version": "7.0.0"}},
			},
			// url: upstream valid 351.
			{Code: "var { URL } = require('url');",
				Options: []any{map[string]any{"version": "7.0.0"}},
			},
			// url: upstream valid 352.
			{Code: "import cp from 'url'; cp.URL",
				Options: []any{map[string]any{"version": "7.0.0"}},
			},
			// url: upstream valid 353.
			{Code: "import { URL } from 'url'",
				Options: []any{map[string]any{"version": "7.0.0"}},
			},
			// url: upstream valid 354.
			{Code: "require('url').URLSearchParams",
				Options: []any{map[string]any{"version": "7.5.0"}},
			},
			// url: upstream valid 355.
			{Code: "require('url').URLSearchParams",
				Options: []any{map[string]any{"version": "6.13.0"}},
			},
			// url: upstream valid 356.
			{Code: "require('url').domainToASCII",
				Options: []any{map[string]any{"version": "7.4.0"}},
			},
			// url: upstream valid 357.
			{Code: "require('url').domainToUnicode",
				Options: []any{map[string]any{"version": "7.4.0"}},
			},
			// url: upstream valid 358.
			{Code: "URL",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"URL"}}},
			},
			// url: upstream valid 359.
			{Code: "URLSearchParams",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"URLSearchParams"}}},
			},
			// url: upstream valid 360.
			{Code: "require('url').URL",
				Options: []any{map[string]any{"version": "6.9.9", "ignores": []any{"url.URL"}}},
			},
			// url: upstream valid 361.
			{Code: "var cp = require('url'); cp.URL",
				Options: []any{map[string]any{"version": "6.9.9", "ignores": []any{"url.URL"}}},
			},
			// url: upstream valid 362.
			{Code: "var { URL } = require('url');",
				Options: []any{map[string]any{"version": "6.9.9", "ignores": []any{"url.URL"}}},
			},
			// url: upstream valid 363.
			{Code: "import cp from 'url'; cp.URL",
				Options: []any{map[string]any{"version": "6.9.9", "ignores": []any{"url.URL"}}},
			},
			// url: upstream valid 364.
			{Code: "import { URL } from 'url'",
				Options: []any{map[string]any{"version": "6.9.9", "ignores": []any{"url.URL"}}},
			},
			// url: upstream valid 365.
			{Code: "require('url').URLSearchParams",
				Options: []any{map[string]any{"version": "7.4.9", "ignores": []any{"url.URLSearchParams"}}},
			},
			// url: upstream valid 366.
			{Code: "require('url').domainToASCII",
				Options: []any{map[string]any{"version": "7.3.9", "ignores": []any{"url.domainToASCII"}}},
			},
			// url: upstream valid 367.
			{Code: "require('url').domainToUnicode",
				Options: []any{map[string]any{"version": "7.3.9", "ignores": []any{"url.domainToUnicode"}}},
			},
			// util: upstream valid 368.
			{Code: "require('util').callbackify",
				Options: []any{map[string]any{"version": "8.2.0"}},
			},
			// util: upstream valid 369.
			{Code: "var hooks = require('util'); hooks.callbackify",
				Options: []any{map[string]any{"version": "8.2.0"}},
			},
			// util: upstream valid 370.
			{Code: "var { callbackify } = require('util'); callbackify",
				Options: []any{map[string]any{"version": "8.2.0"}},
			},
			// util: upstream valid 371.
			{Code: "import util from 'util'; util.callbackify",
				Options: []any{map[string]any{"version": "8.2.0"}},
			},
			// util: upstream valid 372.
			{Code: "import { callbackify } from 'util'; callbackify",
				Options: []any{map[string]any{"version": "8.2.0"}},
			},
			// util: upstream valid 373.
			{Code: "require('util').formatWithOptions",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// util: upstream valid 374.
			{Code: "require('util').getSystemErrorName",
				Options: []any{map[string]any{"version": "9.7.0"}},
			},
			// util: upstream valid 375.
			{Code: "require('util').inspect.custom",
				Options: []any{map[string]any{"version": "6.6.0"}},
			},
			// util: upstream valid 376.
			{Code: "require('util').inspect.defaultOptions",
				Options: []any{map[string]any{"version": "6.4.0"}},
			},
			// util: upstream valid 377.
			{Code: "require('util').isDeepStrictEqual",
				Options: []any{map[string]any{"version": "9.0.0"}},
			},
			// util: upstream valid 378.
			{Code: "require('util').promisify",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// util: upstream valid 379.
			{Code: "require('util').TextDecoder",
				Options: []any{map[string]any{"version": "8.9.0"}},
			},
			// util: upstream valid 380.
			{Code: "require('util').TextEncoder",
				Options: []any{map[string]any{"version": "8.9.0"}},
			},
			// util: upstream valid 381.
			{Code: "require('util').types",
				Options: []any{map[string]any{"version": "10.0.0"}},
			},
			// util: upstream valid 382.
			{Code: "require('util').styleText",
				Options: []any{map[string]any{"version": "21.7.0", "allowExperimental": true}},
			},
			// util: upstream valid 383.
			{Code: "import { styleText } from 'node:util'; styleText('green', 'ok')",
				Options: []any{map[string]any{"version": "21.7.0", "allowExperimental": true}},
			},
			// util: upstream valid 384.
			{Code: "require('util').callbackify",
				Options: []any{map[string]any{"version": "8.1.9", "ignores": []any{"util.callbackify"}}},
			},
			// util: upstream valid 385.
			{Code: "var hooks = require('util'); hooks.callbackify",
				Options: []any{map[string]any{"version": "8.1.9", "ignores": []any{"util.callbackify"}}},
			},
			// util: upstream valid 386.
			{Code: "var { callbackify } = require('util'); callbackify",
				Options: []any{map[string]any{"version": "8.1.9", "ignores": []any{"util.callbackify"}}},
			},
			// util: upstream valid 387.
			{Code: "import util from 'util'; util.callbackify",
				Options: []any{map[string]any{"version": "8.1.9", "ignores": []any{"util.callbackify"}}},
			},
			// util: upstream valid 388.
			{Code: "import { callbackify } from 'util'; callbackify",
				Options: []any{map[string]any{"version": "8.1.9", "ignores": []any{"util.callbackify"}}},
			},
			// util: upstream valid 389.
			{Code: "require('util').formatWithOptions",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"util.formatWithOptions"}}},
			},
			// util: upstream valid 390.
			{Code: "require('util').getSystemErrorName",
				Options: []any{map[string]any{"version": "9.6.9", "ignores": []any{"util.getSystemErrorName"}}},
			},
			// util: upstream valid 391.
			{Code: "require('util').inspect.custom",
				Options: []any{map[string]any{"version": "6.5.9", "ignores": []any{"util.inspect.custom"}}},
			},
			// util: upstream valid 392.
			{Code: "require('util').inspect.defaultOptions",
				Options: []any{map[string]any{"version": "6.3.9", "ignores": []any{"util.inspect.defaultOptions"}}},
			},
			// util: upstream valid 393.
			{Code: "require('util').isDeepStrictEqual",
				Options: []any{map[string]any{"version": "8.9.9", "ignores": []any{"util.isDeepStrictEqual"}}},
			},
			// util: upstream valid 394.
			{Code: "require('util').promisify",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"util.promisify"}}},
			},
			// util: upstream valid 395.
			{Code: "require('util').TextDecoder",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"util.TextDecoder"}}},
			},
			// util: upstream valid 396.
			{Code: "require('util').TextEncoder",
				Options: []any{map[string]any{"version": "8.2.9", "ignores": []any{"util.TextEncoder"}}},
			},
			// util: upstream valid 397.
			{Code: "require('util').types",
				Options: []any{map[string]any{"version": "9.9.9", "ignores": []any{"util.types"}}},
			},
			// v8: upstream valid 398.
			{Code: "require('v8')",
				Options: []any{map[string]any{"version": "1.0.0"}},
			},
			// v8: upstream valid 399.
			{Code: "import hooks from 'v8'",
				Options: []any{map[string]any{"version": "1.0.0"}},
			},
			// v8: upstream valid 400.
			{Code: "require('v8').cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// v8: upstream valid 401.
			{Code: "var hooks = require('v8'); hooks.cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// v8: upstream valid 402.
			{Code: "var { cachedDataVersionTag } = require('v8'); cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// v8: upstream valid 403.
			{Code: "import v8 from 'v8'; v8.cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// v8: upstream valid 404.
			{Code: "import { cachedDataVersionTag } from 'v8'; cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// v8: upstream valid 405.
			{Code: "require('v8').getHeapSpaceStatistics()",
				Options: []any{map[string]any{"version": "6.0.0"}},
			},
			// v8: upstream valid 406.
			{Code: "require('v8').serialize()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// v8: upstream valid 407.
			{Code: "require('v8').deserialize()",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// v8: upstream valid 408.
			{Code: "require('v8').Serializer",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// v8: upstream valid 409.
			{Code: "require('v8').Deserializer",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// v8: upstream valid 410.
			{Code: "require('v8').DefaultSerializer",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// v8: upstream valid 411.
			{Code: "require('v8').DefaultDeserializer",
				Options: []any{map[string]any{"version": "8.0.0"}},
			},
			// v8: upstream valid 412.
			{Code: "require('v8')",
				Options: []any{map[string]any{"version": "0.12.99", "ignores": []any{"v8"}}},
			},
			// v8: upstream valid 413.
			{Code: "import hooks from 'v8'",
				Options: []any{map[string]any{"version": "0.12.99", "ignores": []any{"v8"}}},
			},
			// v8: upstream valid 414.
			{Code: "import { cachedDataVersionTag } from 'v8'",
				Options: []any{map[string]any{"version": "0.12.99", "ignores": []any{"v8", "v8.cachedDataVersionTag"}}},
			},
			// v8: upstream valid 415.
			{Code: "require('v8').cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"v8.cachedDataVersionTag"}}},
			},
			// v8: upstream valid 416.
			{Code: "var hooks = require('v8'); hooks.cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"v8.cachedDataVersionTag"}}},
			},
			// v8: upstream valid 417.
			{Code: "var { cachedDataVersionTag } = require('v8'); cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"v8.cachedDataVersionTag"}}},
			},
			// v8: upstream valid 418.
			{Code: "import v8 from 'v8'; v8.cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"v8.cachedDataVersionTag"}}},
			},
			// v8: upstream valid 419.
			{Code: "import { cachedDataVersionTag } from 'v8'; cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"v8.cachedDataVersionTag"}}},
			},
			// v8: upstream valid 420.
			{Code: "require('v8').getHeapSpaceStatistics()",
				Options: []any{map[string]any{"version": "5.9.9", "ignores": []any{"v8.getHeapSpaceStatistics"}}},
			},
			// v8: upstream valid 421.
			{Code: "require('v8').serialize()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"v8.serialize"}}},
			},
			// v8: upstream valid 422.
			{Code: "require('v8').deserialize()",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"v8.deserialize"}}},
			},
			// v8: upstream valid 423.
			{Code: "require('v8').Serializer",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"v8.Serializer"}}},
			},
			// v8: upstream valid 424.
			{Code: "require('v8').Deserializer",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"v8.Deserializer"}}},
			},
			// v8: upstream valid 425.
			{Code: "require('v8').DefaultSerializer",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"v8.DefaultSerializer"}}},
			},
			// v8: upstream valid 426.
			{Code: "require('v8').DefaultDeserializer",
				Options: []any{map[string]any{"version": "7.9.9", "ignores": []any{"v8.DefaultDeserializer"}}},
			},
			// vm: upstream valid 427.
			{Code: "require('vm')",
				Options: []any{map[string]any{"version": "9.6.0"}},
			},
			// vm: upstream valid 428.
			{Code: "import vm from 'vm';",
				Options: []any{map[string]any{"version": "9.6.0"}},
			},
			// vm: upstream valid 429.
			{Code: "import * as vm from 'vm';",
				Options: []any{map[string]any{"version": "9.6.0"}},
			},
			// vm: upstream valid 430.
			{Code: "require('vm').Module",
				Options: []any{map[string]any{"version": "9.5.9", "ignores": []any{"vm.Module"}}},
			},
			// vm: upstream valid 431.
			{Code: "var vm = require('vm'); vm.Module",
				Options: []any{map[string]any{"version": "9.5.9", "ignores": []any{"vm.Module"}}},
			},
			// vm: upstream valid 432.
			{Code: "var { Module } = require('vm'); Module",
				Options: []any{map[string]any{"version": "9.5.9", "ignores": []any{"vm.Module"}}},
			},
			// vm: upstream valid 433.
			{Code: "import vm from 'vm'; vm.Module",
				Options: []any{map[string]any{"version": "9.5.9", "ignores": []any{"vm.Module"}}},
			},
			// vm: upstream valid 434.
			{Code: "import { Module } from 'vm'; Module",
				Options: []any{map[string]any{"version": "9.5.9", "ignores": []any{"vm.Module"}}},
			},
			// worker_threads: upstream valid 435.
			{Code: "require('worker_threads')",
				Options: []any{map[string]any{"version": "10.4.99", "ignores": []any{"worker_threads"}}},
			},
			// worker_threads: upstream valid 436.
			{Code: "import worker_threads from 'worker_threads'",
				Options: []any{map[string]any{"version": "10.4.99", "ignores": []any{"worker_threads"}}},
			},
			// worker_threads: upstream valid 437.
			{Code: "require('worker_threads')",
				Options: []any{map[string]any{"version": "12.11.0"}},
			},
			// worker_threads: upstream valid 438.
			{Code: "import worker_threads from 'worker_threads'",
				Options: []any{map[string]any{"version": "12.11.0"}},
			},
			// worker_threads: upstream valid 439.
			{Code: "import worker_threads from 'worker_threads'",
				Settings: map[string]any{"node": map[string]any{"version": "12.11.0"}},
			},
			// timers/promises: upstream valid 440.
			{Code: "\n                        import { scheduler } from 'node:timers/promises';\n                        await scheduler.wait( 1000 );\n                    ",
				Options: []any{map[string]any{"version": ">= 20.0.0", "ignores": []any{"timers/promises.scheduler.wait"}}},
			},
			// import.meta: upstream valid 441.
			{Code: "import.meta.resolve(specifier)",
				Options: []any{map[string]any{"version": "22.0.0"}},
			},
			// import.meta: upstream valid 442.
			{Code: "import.meta.resolve(specifier)",
				Options: []any{map[string]any{"version": "20.6.0"}},
			},
			// import.meta: upstream valid 443.
			{Code: "import.meta.resolve(specifier)",
				Options: []any{map[string]any{"version": "18.19.0"}},
			},
			// import.meta: upstream valid 444.
			{Code: "import.meta.resolve(specifier)",
				Options: []any{map[string]any{"version": "13.9.0", "allowExperimental": true}},
			},
			// import.meta: upstream valid 445.
			{Code: "import.meta.resolve(specifier)",
				Options: []any{map[string]any{"version": "12.16.2", "allowExperimental": true}},
			},
			// import.meta: upstream valid 446.
			{Code: "import.meta.resolve(specifier)",
				Options: []any{map[string]any{"version": "18.18.0", "ignores": []any{"import.meta.resolve"}}},
			},
			// import.meta: upstream valid 447.
			{Code: "import.meta.dirname;",
				Options: []any{map[string]any{"version": "22.16.0"}},
			},
			// import.meta: upstream valid 448.
			{Code: "import.meta.dirname;",
				Options: []any{map[string]any{"version": "22.0.0", "allowExperimental": true}},
			},
			// import.meta: upstream valid 449.
			{Code: "import.meta.dirname;",
				Options: []any{map[string]any{"version": "21.2.0", "allowExperimental": true}},
			},
			// import.meta: upstream valid 450.
			{Code: "import.meta.dirname;",
				Options: []any{map[string]any{"version": "20.11.0", "allowExperimental": true}},
			},
			// import.meta: upstream valid 451.
			{Code: "import.meta.dirname;",
				Options: []any{map[string]any{"version": "20.10.0", "ignores": []any{"import.meta.dirname"}}},
			},
			// import.meta: upstream valid 452.
			{Code: "import.meta.filename;",
				Options: []any{map[string]any{"version": "22.16.0"}},
			},
			// import.meta: upstream valid 453.
			{Code: "import.meta.filename;",
				Options: []any{map[string]any{"version": "22.0.0", "allowExperimental": true}},
			},
			// import.meta: upstream valid 454.
			{Code: "import.meta.filename;",
				Options: []any{map[string]any{"version": "21.2.0", "allowExperimental": true}},
			},
			// import.meta: upstream valid 455.
			{Code: "import.meta.filename;",
				Options: []any{map[string]any{"version": "20.11.0", "allowExperimental": true}},
			},
			// import.meta: upstream valid 456.
			{Code: "import.meta.filename;",
				Options: []any{map[string]any{"version": "20.10.0", "ignores": []any{"import.meta.filename"}}},
			},
			// fetch: upstream valid 457.
			{Code: "fetch('/asd')",
				Options: []any{map[string]any{"version": "16.16.0", "allowExperimental": true}},
			},
			// sqlite: upstream valid 458.
			{Code: "\n                        import { DatabaseSync } from 'node:sqlite';\n                        const database = new DatabaseSync(':memory:');\n                    ",
				Options: []any{map[string]any{"version": ">=22.5.0", "allowExperimental": true}},
			},
			// sqlite: upstream valid 459.
			{Code: "\n                        import { DatabaseSync } from 'node:sqlite';\n                        const database = new DatabaseSync(':memory:');\n                    ",
				Options: []any{map[string]any{"version": ">=22.3.0", "ignores": []any{"sqlite", "sqlite.DatabaseSync"}}},
			},
			// sqlite: upstream valid 460.
			{Code: "\n                        const { DatabaseSync } = require('node:sqlite');\n                        const database = new DatabaseSync(':memory:');\n                    ",
				Options: []any{map[string]any{"version": ">=22.5.0", "allowExperimental": true}},
			},
			// sqlite: upstream valid 461.
			{Code: "\n                        const { DatabaseSync } = process.getBuiltinModule('node:sqlite');\n                        const database = new DatabaseSync(':memory:');\n                    ",
				Options: []any{map[string]any{"version": ">=22.5.0", "allowExperimental": true}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// assert: upstream invalid 1.
			{Code: "require('assert').deepStrictEqual()",
				Options: []any{map[string]any{"version": "1.1.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.deepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			// assert: upstream invalid 2.
			{Code: "var assert = require('assert'); assert.deepStrictEqual()",
				Options: []any{map[string]any{"version": "1.1.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.deepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.", Line: 1, Column: 33, EndLine: 1, EndColumn: 55},
				},
			},
			// assert: upstream invalid 3.
			{Code: "var { deepStrictEqual } = require('assert'); deepStrictEqual()",
				Options: []any{map[string]any{"version": "1.1.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.deepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 22},
				},
			},
			// assert: upstream invalid 4.
			{Code: "import assert from 'assert'; assert.deepStrictEqual()",
				Options: []any{map[string]any{"version": "1.1.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.deepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.", Line: 1, Column: 30, EndLine: 1, EndColumn: 52},
				},
			},
			// assert: upstream invalid 5.
			{Code: "import { deepStrictEqual } from 'assert'; deepStrictEqual()",
				Options: []any{map[string]any{"version": "1.1.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.deepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 25},
				},
			},
			// assert: upstream invalid 6.
			{Code: "require('assert').notDeepStrictEqual()",
				Options: []any{map[string]any{"version": "1.1.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.notDeepStrictEqual' is still an experimental feature and is not supported until Node.js 1.2.0. The configured version range is '1.1.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},
			// assert: upstream invalid 7.
			{Code: "require('assert').rejects()",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.rejects' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// assert: upstream invalid 8.
			{Code: "require('assert').doesNotReject()",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.doesNotReject' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			// assert: upstream invalid 9.
			{Code: "require('assert').strict.rejects()",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.strict.rejects' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			// assert: upstream invalid 10.
			{Code: "require('assert').strict.doesNotReject()",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.strict.doesNotReject' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
				},
			},
			// assert: upstream invalid 11.
			{Code: "var assert = require('assert').strict",
				Options: []any{map[string]any{"version": "9.8.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.strict' is still an experimental feature and is not supported until Node.js 9.9.0 (backported: ^8.13.0). The configured version range is '9.8.9'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 38},
				},
			},
			// assert: upstream invalid 12.
			{Code: "var {strict: assert} = require('assert'); assert.rejects()",
				Options: []any{map[string]any{"version": "9.8.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert.strict' is still an experimental feature and is not supported until Node.js 9.9.0 (backported: ^8.13.0). The configured version range is '9.8.9'.", Line: 1, Column: 6, EndLine: 1, EndColumn: 20},
					{MessageId: "not-supported-till", Message: "The 'assert.strict.rejects' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.8.9'.", Line: 1, Column: 43, EndLine: 1, EndColumn: 57},
				},
			},
			// assert: upstream invalid 13.
			{Code: "const { CallTracker } = require('assert'); new CallTracker();",
				Options: []any{map[string]any{"version": "14.2.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'assert.CallTracker' is still an experimental feature The configured version range is '14.2.0'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20},
				},
			},
			// assert: upstream invalid 14.
			{Code: "import { CallTracker } from 'assert'; new CallTracker();",
				Options: []any{map[string]any{"version": "14.2.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'assert.CallTracker' is still an experimental feature The configured version range is '14.2.0'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 21},
				},
			},
			// assert: upstream invalid 15.
			{Code: "require('node:assert').deepStrictEqual()",
				Options: []any{map[string]any{"version": "3.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert' is still an experimental feature and is not supported until Node.js 14.13.1 (backported: ^12.20.0). The configured version range is '3.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			// assert: upstream invalid 16.
			{Code: "import assert from 'node:assert';",
				Options: []any{map[string]any{"version": "3.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'assert' is still an experimental feature and is not supported until Node.js 14.13.1 (backported: ^12.20.0). The configured version range is '3.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			// async_hooks: upstream invalid 17.
			{Code: "require('async_hooks')",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			// async_hooks: upstream invalid 18.
			{Code: "import hooks from 'async_hooks'",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			// async_hooks: upstream invalid 19.
			{Code: "require('async_hooks').createHook()",
				Options: []any{map[string]any{"version": "16.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			// async_hooks: upstream invalid 20.
			{Code: "const { createHook } = require('async_hooks')",
				Options: []any{map[string]any{"version": "16.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19},
				},
			},
			// async_hooks: upstream invalid 21.
			{Code: "const { createHook } = require('async_hooks'); createHook()",
				Options: []any{map[string]any{"version": "16.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19},
				},
			},
			// async_hooks: upstream invalid 22.
			{Code: "const hooks = require('async_hooks'); hooks.createHook()",
				Options: []any{map[string]any{"version": "16.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.", Line: 1, Column: 39, EndLine: 1, EndColumn: 55},
				},
			},
			// async_hooks: upstream invalid 23.
			{Code: "import { createHook } from 'async_hooks'",
				Options: []any{map[string]any{"version": "16.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 20},
				},
			},
			// async_hooks: upstream invalid 24.
			{Code: "import { createHook } from 'async_hooks'; createHook()",
				Options: []any{map[string]any{"version": "16.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 20},
				},
			},
			// async_hooks: upstream invalid 25.
			{Code: "import async_hooks from 'async_hooks'; async_hooks.createHook()",
				Options: []any{map[string]any{"version": "16.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'async_hooks.createHook' is still an experimental feature The configured version range is '16.5.0'.", Line: 1, Column: 40, EndLine: 1, EndColumn: 62},
				},
			},
			// async_hooks: upstream invalid 26.
			{Code: "const hooks = require('async_hooks'); new hooks.AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "13.9.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 37},
					{MessageId: "not-supported-till", Message: "The 'async_hooks.AsyncLocalStorage' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.", Line: 1, Column: 43, EndLine: 1, EndColumn: 66},
				},
			},
			// async_hooks: upstream invalid 27.
			{Code: "import * as hooks from 'async_hooks'; new hooks.AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "13.9.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
					{MessageId: "not-supported-till", Message: "The 'async_hooks.AsyncLocalStorage' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.", Line: 1, Column: 43, EndLine: 1, EndColumn: 66},
				},
			},
			// async_hooks: upstream invalid 28.
			{Code: "import hooks from 'async_hooks'; new hooks.AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "13.9.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
					{MessageId: "not-supported-till", Message: "The 'async_hooks.AsyncLocalStorage' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.", Line: 1, Column: 38, EndLine: 1, EndColumn: 61},
				},
			},
			// async_hooks: upstream invalid 29.
			{Code: "import { AsyncLocalStorage } from 'async_hooks'; new AsyncLocalStorage()",
				Options: []any{map[string]any{"version": "13.9.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 49},
					{MessageId: "not-supported-till", Message: "The 'async_hooks.AsyncLocalStorage' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '13.9.0'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
				},
			},
			// async_hooks: upstream invalid 30.
			{Code: "require('node:async_hooks')",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			// async_hooks: upstream invalid 31.
			{Code: "import hooks from 'node:async_hooks'",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'async_hooks' is still an experimental feature and is not supported until Node.js 16.4.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},
			// buffer: upstream invalid 32.
			{Code: "Buffer.alloc",
				Options: []any{map[string]any{"version": "4.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'Buffer.alloc' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			// buffer: upstream invalid 33.
			{Code: "Buffer.allocUnsafe",
				Options: []any{map[string]any{"version": "4.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'Buffer.allocUnsafe' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			// buffer: upstream invalid 34.
			{Code: "Buffer.allocUnsafeSlow",
				Options: []any{map[string]any{"version": "4.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'Buffer.allocUnsafeSlow' is still an experimental feature and is not supported until Node.js 5.12.0 (backported: ^4.5.0). The configured version range is '4.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			// buffer: upstream invalid 35.
			{Code: "Buffer.from",
				Options: []any{map[string]any{"version": "4.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'Buffer.from' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			// buffer: upstream invalid 36.
			{Code: "require('buffer').constants",
				Options: []any{map[string]any{"version": "8.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.constants' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			// buffer: upstream invalid 37.
			{Code: "var cp = require('buffer'); cp.constants",
				Options: []any{map[string]any{"version": "8.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.constants' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 41},
				},
			},
			// buffer: upstream invalid 38.
			{Code: "var { constants } = require('buffer');",
				Options: []any{map[string]any{"version": "8.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.constants' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 16},
				},
			},
			// buffer: upstream invalid 39.
			{Code: "import cp from 'buffer'; cp.constants",
				Options: []any{map[string]any{"version": "8.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.constants' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 38},
				},
			},
			// buffer: upstream invalid 40.
			{Code: "import { constants } from 'buffer'",
				Options: []any{map[string]any{"version": "8.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.constants' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				},
			},
			// buffer: upstream invalid 41.
			{Code: "var {Buffer: b} = require('buffer'); b.alloc",
				Options: []any{map[string]any{"version": "4.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.Buffer.alloc' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.", Line: 1, Column: 38, EndLine: 1, EndColumn: 45},
				},
			},
			// buffer: upstream invalid 42.
			{Code: "var {Buffer: b} = require('buffer'); b.allocUnsafe",
				Options: []any{map[string]any{"version": "4.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.Buffer.allocUnsafe' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.", Line: 1, Column: 38, EndLine: 1, EndColumn: 51},
				},
			},
			// buffer: upstream invalid 43.
			{Code: "var {Buffer: b} = require('buffer'); b.allocUnsafeSlow",
				Options: []any{map[string]any{"version": "4.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.Buffer.allocUnsafeSlow' is still an experimental feature and is not supported until Node.js 5.12.0 (backported: ^4.5.0). The configured version range is '4.4.9'.", Line: 1, Column: 38, EndLine: 1, EndColumn: 55},
				},
			},
			// buffer: upstream invalid 44.
			{Code: "var {Buffer: b} = require('buffer'); b.from",
				Options: []any{map[string]any{"version": "4.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.Buffer.from' is still an experimental feature and is not supported until Node.js 5.10.0 (backported: ^4.5.0). The configured version range is '4.4.9'.", Line: 1, Column: 38, EndLine: 1, EndColumn: 44},
				},
			},
			// buffer: upstream invalid 45.
			{Code: "require('buffer').kMaxLength",
				Options: []any{map[string]any{"version": "2.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.kMaxLength' is still an experimental feature and is not supported until Node.js 3.0.0. The configured version range is '2.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},
			// buffer: upstream invalid 46.
			{Code: "require('buffer').transcode",
				Options: []any{map[string]any{"version": "7.0.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.transcode' is still an experimental feature and is not supported until Node.js 7.1.0. The configured version range is '7.0.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			// buffer: upstream invalid 47.
			{Code: "const { Blob } = require('buffer'); new Blob();",
				Options: []any{map[string]any{"version": "15.7.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.Blob' is still an experimental feature and is not supported until Node.js 18.0.0 (backported: ^16.17.0). The configured version range is '15.7.0'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13},
				},
			},
			// buffer: upstream invalid 48.
			{Code: "import buffer from 'buffer'; new buffer.Blob();",
				Options: []any{map[string]any{"version": "15.7.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'buffer.Blob' is still an experimental feature and is not supported until Node.js 18.0.0 (backported: ^16.17.0). The configured version range is '15.7.0'.", Line: 1, Column: 34, EndLine: 1, EndColumn: 45},
				},
			},
			// child_process: upstream invalid 49.
			{Code: "require('child_process').ChildProcess",
				Options: []any{map[string]any{"version": "2.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'child_process.ChildProcess' is still an experimental feature and is not supported until Node.js 2.2.0. The configured version range is '2.1.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},
			// child_process: upstream invalid 50.
			{Code: "var cp = require('child_process'); cp.ChildProcess",
				Options: []any{map[string]any{"version": "2.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'child_process.ChildProcess' is still an experimental feature and is not supported until Node.js 2.2.0. The configured version range is '2.1.9'.", Line: 1, Column: 36, EndLine: 1, EndColumn: 51},
				},
			},
			// child_process: upstream invalid 51.
			{Code: "var { ChildProcess } = require('child_process'); ChildProcess",
				Options: []any{map[string]any{"version": "2.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'child_process.ChildProcess' is still an experimental feature and is not supported until Node.js 2.2.0. The configured version range is '2.1.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 19},
				},
			},
			// child_process: upstream invalid 52.
			{Code: "import cp from 'child_process'; cp.ChildProcess",
				Options: []any{map[string]any{"version": "2.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'child_process.ChildProcess' is still an experimental feature and is not supported until Node.js 2.2.0. The configured version range is '2.1.9'.", Line: 1, Column: 33, EndLine: 1, EndColumn: 48},
				},
			},
			// child_process: upstream invalid 53.
			{Code: "import { ChildProcess } from 'child_process'",
				Options: []any{map[string]any{"version": "2.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'child_process.ChildProcess' is still an experimental feature and is not supported until Node.js 2.2.0. The configured version range is '2.1.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 22},
				},
			},
			// console: upstream invalid 54.
			{Code: "console.clear()",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.clear' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			// console: upstream invalid 55.
			{Code: "require('console').clear()",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.clear' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			// console: upstream invalid 56.
			{Code: "var c = require('console'); c.clear()",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.clear' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 36},
				},
			},
			// console: upstream invalid 57.
			{Code: "var { clear } = require('console'); clear()",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.clear' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 12},
				},
			},
			// console: upstream invalid 58.
			{Code: "import c from 'console'; c.clear()",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.clear' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 33},
				},
			},
			// console: upstream invalid 59.
			{Code: "console.count()",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.count' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			// console: upstream invalid 60.
			{Code: "console.countReset()",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.countReset' is still an experimental feature and is not supported until Node.js 8.3.0 (backported: ^6.13.0). The configured version range is '8.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			// console: upstream invalid 61.
			{Code: "console.debug()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.debug' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			// console: upstream invalid 62.
			{Code: "console.dirxml()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.dirxml' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			// console: upstream invalid 63.
			{Code: "console.group()",
				Options: []any{map[string]any{"version": "8.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.group' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			// console: upstream invalid 64.
			{Code: "console.groupCollapsed()",
				Options: []any{map[string]any{"version": "8.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.groupCollapsed' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			// console: upstream invalid 65.
			{Code: "console.groupEnd()",
				Options: []any{map[string]any{"version": "8.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.groupEnd' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// console: upstream invalid 66.
			{Code: "console.table()",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.table' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			// console: upstream invalid 67.
			{Code: "console.profile()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.profile' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// console: upstream invalid 68.
			{Code: "console.profileEnd()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.profileEnd' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			// console: upstream invalid 69.
			{Code: "console.timeStamp()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'console.timeStamp' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			// crypto: upstream invalid 70.
			{Code: "require('crypto').constants",
				Options: []any{map[string]any{"version": "6.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.constants' is still an experimental feature and is not supported until Node.js 6.3.0. The configured version range is '6.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			// crypto: upstream invalid 71.
			{Code: "var hooks = require('crypto'); hooks.constants",
				Options: []any{map[string]any{"version": "6.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.constants' is still an experimental feature and is not supported until Node.js 6.3.0. The configured version range is '6.2.9'.", Line: 1, Column: 32, EndLine: 1, EndColumn: 47},
				},
			},
			// crypto: upstream invalid 72.
			{Code: "var { constants } = require('crypto'); constants",
				Options: []any{map[string]any{"version": "6.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.constants' is still an experimental feature and is not supported until Node.js 6.3.0. The configured version range is '6.2.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 16},
				},
			},
			// crypto: upstream invalid 73.
			{Code: "import crypto from 'crypto'; crypto.constants",
				Options: []any{map[string]any{"version": "6.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.constants' is still an experimental feature and is not supported until Node.js 6.3.0. The configured version range is '6.2.9'.", Line: 1, Column: 30, EndLine: 1, EndColumn: 46},
				},
			},
			// crypto: upstream invalid 74.
			{Code: "import { constants } from 'crypto'; constants",
				Options: []any{map[string]any{"version": "6.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.constants' is still an experimental feature and is not supported until Node.js 6.3.0. The configured version range is '6.2.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				},
			},
			// crypto: upstream invalid 75.
			{Code: "require('crypto').Certificate.exportChallenge()",
				Options: []any{map[string]any{"version": "8.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.Certificate.exportChallenge' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
				},
			},
			// crypto: upstream invalid 76.
			{Code: "var { Certificate: c } = require('crypto'); c.exportChallenge()",
				Options: []any{map[string]any{"version": "8.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.Certificate.exportChallenge' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.", Line: 1, Column: 45, EndLine: 1, EndColumn: 62},
				},
			},
			// crypto: upstream invalid 77.
			{Code: "var { Certificate: c } = require('crypto'); c.exportPublicKey()",
				Options: []any{map[string]any{"version": "8.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.Certificate.exportPublicKey' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.", Line: 1, Column: 45, EndLine: 1, EndColumn: 62},
				},
			},
			// crypto: upstream invalid 78.
			{Code: "var { Certificate: c } = require('crypto'); c.verifySpkac()",
				Options: []any{map[string]any{"version": "8.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.Certificate.verifySpkac' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.", Line: 1, Column: 45, EndLine: 1, EndColumn: 58},
				},
			},
			// crypto: upstream invalid 79.
			{Code: "require('crypto').fips",
				Options: []any{map[string]any{"version": "5.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.fips' is still an experimental feature and is not supported until Node.js 6.0.0. The configured version range is '5.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			// crypto: upstream invalid 80.
			{Code: "require('crypto').getCurves",
				Options: []any{map[string]any{"version": "2.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.getCurves' is still an experimental feature and is not supported until Node.js 2.3.0. The configured version range is '2.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			// crypto: upstream invalid 81.
			{Code: "require('crypto').getFips",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.getFips' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// crypto: upstream invalid 82.
			{Code: "require('crypto').privateEncrypt",
				Options: []any{map[string]any{"version": "1.0.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.privateEncrypt' is still an experimental feature and is not supported until Node.js 1.1.0. The configured version range is '1.0.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			// crypto: upstream invalid 83.
			{Code: "require('crypto').publicDecrypt",
				Options: []any{map[string]any{"version": "1.0.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.publicDecrypt' is still an experimental feature and is not supported until Node.js 1.1.0. The configured version range is '1.0.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			// crypto: upstream invalid 84.
			{Code: "require('crypto').randomFillSync",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.randomFillSync' is still an experimental feature and is not supported until Node.js 7.10.0 (backported: ^6.13.0). The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			// crypto: upstream invalid 85.
			{Code: "require('crypto').randomFill",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.randomFill' is still an experimental feature and is not supported until Node.js 7.10.0 (backported: ^6.13.0). The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},
			// crypto: upstream invalid 86.
			{Code: "require('crypto').scrypt",
				Options: []any{map[string]any{"version": "10.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.scrypt' is still an experimental feature and is not supported until Node.js 10.5.0. The configured version range is '10.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			// crypto: upstream invalid 87.
			{Code: "require('crypto').scryptSync",
				Options: []any{map[string]any{"version": "10.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.scryptSync' is still an experimental feature and is not supported until Node.js 10.5.0. The configured version range is '10.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},
			// crypto: upstream invalid 88.
			{Code: "require('crypto').setFips",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.setFips' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// crypto: upstream invalid 89.
			{Code: "require('crypto').timingSafeEqual",
				Options: []any{map[string]any{"version": "6.5.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'crypto.timingSafeEqual' is still an experimental feature and is not supported until Node.js 6.6.0. The configured version range is '6.5.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			// dns: upstream invalid 90.
			{Code: "require('dns').Resolver",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			// dns: upstream invalid 91.
			{Code: "var hooks = require('dns'); hooks.Resolver",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.2.9'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 43},
				},
			},
			// dns: upstream invalid 92.
			{Code: "var { Resolver } = require('dns'); Resolver",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.2.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 15},
				},
			},
			// dns: upstream invalid 93.
			{Code: "import dns from 'dns'; dns.Resolver",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.2.9'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 36},
				},
			},
			// dns: upstream invalid 94.
			{Code: "import { Resolver } from 'dns'; Resolver",
				Options: []any{map[string]any{"version": "8.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'dns.Resolver' is still an experimental feature and is not supported until Node.js 8.3.0. The configured version range is '8.2.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 18},
				},
			},
			// dns: upstream invalid 95.
			{Code: "require('dns').resolvePtr",
				Options: []any{map[string]any{"version": "5.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'dns.resolvePtr' is still an experimental feature and is not supported until Node.js 6.0.0. The configured version range is '5.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// dns: upstream invalid 96.
			{Code: "require('dns').promises",
				Options: []any{map[string]any{"version": "11.13.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'dns.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			// fs: upstream invalid 97.
			{Code: "require('fs').promises",
				Options: []any{map[string]any{"version": "11.13.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			// fs: upstream invalid 98.
			{Code: "var fs = require('fs'); fs.promises",
				Options: []any{map[string]any{"version": "11.13.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.", Line: 1, Column: 25, EndLine: 1, EndColumn: 36},
				},
			},
			// fs: upstream invalid 99.
			{Code: "var { promises } = require('fs'); promises",
				Options: []any{map[string]any{"version": "11.13.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 15},
				},
			},
			// fs: upstream invalid 100.
			{Code: "import fs from 'fs'; fs.promises",
				Options: []any{map[string]any{"version": "11.13.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 33},
				},
			},
			// fs: upstream invalid 101.
			{Code: "import { promises } from 'fs'",
				Options: []any{map[string]any{"version": "11.13.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.promises' is still an experimental feature and is not supported until Node.js 11.14.0 (backported: ^10.17.0). The configured version range is '11.13.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 18},
				},
			},
			// fs: upstream invalid 102.
			{Code: "require('fs').copyFile",
				Options: []any{map[string]any{"version": "8.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.copyFile' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			// fs: upstream invalid 103.
			{Code: "require('fs').copyFileSync",
				Options: []any{map[string]any{"version": "8.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.copyFileSync' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			// fs: upstream invalid 104.
			{Code: "require('fs').mkdtemp",
				Options: []any{map[string]any{"version": "5.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.mkdtemp' is still an experimental feature and is not supported until Node.js 5.10.0. The configured version range is '5.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			// fs: upstream invalid 105.
			{Code: "require('fs').mkdtempSync",
				Options: []any{map[string]any{"version": "5.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.mkdtempSync' is still an experimental feature and is not supported until Node.js 5.10.0. The configured version range is '5.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// fs: upstream invalid 106.
			{Code: "require('fs').realpath.native",
				Options: []any{map[string]any{"version": "9.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.realpath.native' is still an experimental feature and is not supported until Node.js 9.2.0. The configured version range is '9.1.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
			// fs: upstream invalid 107.
			{Code: "require('fs').realpathSync.native",
				Options: []any{map[string]any{"version": "9.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.realpathSync.native' is still an experimental feature and is not supported until Node.js 9.2.0. The configured version range is '9.1.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			// fs: upstream invalid 108.
			{Code: "require('fs').lutimes",
				Options: []any{map[string]any{"version": "14.4.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.lutimes' is still an experimental feature and is not supported until Node.js 14.5.0 (backported: ^12.19.0). The configured version range is '14.4.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			// fs: upstream invalid 109.
			{Code: "require('fs').lutimesSync",
				Options: []any{map[string]any{"version": "14.4.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.lutimesSync' is still an experimental feature and is not supported until Node.js 14.5.0 (backported: ^12.19.0). The configured version range is '14.4.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// fs: upstream invalid 110.
			{Code: "require('fs').readv",
				Options: []any{map[string]any{"version": "13.12.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.readv' is still an experimental feature and is not supported until Node.js 13.13.0 (backported: ^12.17.0). The configured version range is '13.12.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// fs: upstream invalid 111.
			{Code: "require('fs').readvSync",
				Options: []any{map[string]any{"version": "13.12.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.readvSync' is still an experimental feature and is not supported until Node.js 13.13.0 (backported: ^12.17.0). The configured version range is '13.12.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			// fs: upstream invalid 112.
			{Code: "require('fs').opendir",
				Options: []any{map[string]any{"version": "12.11.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.opendir' is still an experimental feature and is not supported until Node.js 12.12.0. The configured version range is '12.11.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			// fs: upstream invalid 113.
			{Code: "require('fs').opendirSync",
				Options: []any{map[string]any{"version": "12.11.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.opendirSync' is still an experimental feature and is not supported until Node.js 12.12.0. The configured version range is '12.11.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// fs: upstream invalid 114.
			{Code: "require('fs').rm",
				Options: []any{map[string]any{"version": "14.13.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rm' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '14.13.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// fs: upstream invalid 115.
			{Code: "require('fs').rmSync",
				Options: []any{map[string]any{"version": "14.13.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.rmSync' is still an experimental feature and is not supported until Node.js 14.14.0. The configured version range is '14.13.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// fs: upstream invalid 116.
			{Code: "require('fs').Dir",
				Options: []any{map[string]any{"version": "12.11.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.Dir' is still an experimental feature and is not supported until Node.js 12.12.0. The configured version range is '12.11.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			// fs: upstream invalid 117.
			{Code: "require('fs').StatWatcher",
				Options: []any{map[string]any{"version": "14.2.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.StatWatcher' is still an experimental feature and is not supported until Node.js 14.3.0 (backported: ^12.20.0). The configured version range is '14.2.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// fs: upstream invalid 118.
			{Code: "require('fs').StatWatcher",
				Options: []any{map[string]any{"version": "12.19.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs.StatWatcher' is still an experimental feature and is not supported until Node.js 14.3.0 (backported: ^12.20.0). The configured version range is '12.19.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// fs/promises: upstream invalid 119.
			{Code: "import * as fs from 'fs/promises';",
				Options: []any{map[string]any{"version": "13.14.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs/promises' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '13.14.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
				},
			},
			// fs/promises: upstream invalid 120.
			{Code: "require('fs/promises');",
				Options: []any{map[string]any{"version": "13.14.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs/promises' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '13.14.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			// fs/promises: upstream invalid 121.
			{Code: "const fs = require('fs/promises');",
				Options: []any{map[string]any{"version": "13.14.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs/promises' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '13.14.0'.", Line: 1, Column: 12, EndLine: 1, EndColumn: 34},
				},
			},
			// fs/promises: upstream invalid 122.
			{Code: "require('node:fs/promises');",
				Options: []any{map[string]any{"version": "13.14.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs/promises' is still an experimental feature and is not supported until Node.js 14.13.1. The configured version range is '13.14.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			// fs/promises: upstream invalid 123.
			{Code: "import * as fs from 'node:fs/promises';",
				Options: []any{map[string]any{"version": "13.14.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fs/promises' is still an experimental feature and is not supported until Node.js 14.13.1. The configured version range is '13.14.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
				},
			},
			// http2: upstream invalid 124.
			{Code: "require('http2')",
				Options: []any{map[string]any{"version": "8.3.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'http2' is still an experimental feature and is not supported until Node.js 10.10.0 (backported: ^8.13.0). The configured version range is '8.3.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// http2: upstream invalid 125.
			{Code: "import http2 from 'http2'",
				Options: []any{map[string]any{"version": "8.3.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'http2' is still an experimental feature and is not supported until Node.js 10.10.0 (backported: ^8.13.0). The configured version range is '8.3.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// http2: upstream invalid 126.
			{Code: "import { createServer } from 'http2'",
				Options: []any{map[string]any{"version": "8.3.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'http2' is still an experimental feature and is not supported until Node.js 10.10.0 (backported: ^8.13.0). The configured version range is '8.3.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
					{MessageId: "not-supported-till", Message: "The 'http2.createServer' is still an experimental feature and is not supported until Node.js 8.4.0. The configured version range is '8.3.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 22},
				},
			},
			// inspector: upstream invalid 127.
			{Code: "require('inspector')",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'inspector' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// inspector: upstream invalid 128.
			{Code: "import inspector from 'inspector'",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'inspector' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			// inspector: upstream invalid 129.
			{Code: "import { open } from 'inspector'",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'inspector' is still an experimental feature and is not supported until Node.js 14.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
					{MessageId: "not-supported-till", Message: "The 'inspector.open' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 14},
				},
			},
			// module: upstream invalid 130.
			{Code: "require.resolve.paths()",
				Options: []any{map[string]any{"version": "8.8.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'require.resolve.paths' is still an experimental feature and is not supported until Node.js 8.9.0. The configured version range is '8.8.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			// module: upstream invalid 131.
			{Code: "require('module').builtinModules",
				Options: []any{map[string]any{"version": "9.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'module.builtinModules' is still an experimental feature and is not supported until Node.js 9.3.0 (backported: ^8.10.0, ^6.13.0). The configured version range is '9.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			// os: upstream invalid 132.
			{Code: "require('os').constants",
				Options: []any{map[string]any{"version": "6.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'os.constants' is still an experimental feature and is not supported until Node.js 6.3.0 (backported: ^5.11.0). The configured version range is '6.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			// os: upstream invalid 133.
			{Code: "var hooks = require('os'); hooks.constants",
				Options: []any{map[string]any{"version": "6.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'os.constants' is still an experimental feature and is not supported until Node.js 6.3.0 (backported: ^5.11.0). The configured version range is '6.2.9'.", Line: 1, Column: 28, EndLine: 1, EndColumn: 43},
				},
			},
			// os: upstream invalid 134.
			{Code: "var { constants } = require('os'); constants",
				Options: []any{map[string]any{"version": "6.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'os.constants' is still an experimental feature and is not supported until Node.js 6.3.0 (backported: ^5.11.0). The configured version range is '6.2.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 16},
				},
			},
			// os: upstream invalid 135.
			{Code: "import os from 'os'; os.constants",
				Options: []any{map[string]any{"version": "6.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'os.constants' is still an experimental feature and is not supported until Node.js 6.3.0 (backported: ^5.11.0). The configured version range is '6.2.9'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 34},
				},
			},
			// os: upstream invalid 136.
			{Code: "import { constants } from 'os'; constants",
				Options: []any{map[string]any{"version": "6.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'os.constants' is still an experimental feature and is not supported until Node.js 6.3.0 (backported: ^5.11.0). The configured version range is '6.2.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				},
			},
			// os: upstream invalid 137.
			{Code: "require('os').homedir",
				Options: []any{map[string]any{"version": "2.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'os.homedir' is still an experimental feature and is not supported until Node.js 2.3.0. The configured version range is '2.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			// os: upstream invalid 138.
			{Code: "require('os').userInfo",
				Options: []any{map[string]any{"version": "5.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'os.userInfo' is still an experimental feature and is not supported until Node.js 6.0.0. The configured version range is '5.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			// path: upstream invalid 139.
			{Code: "require('path').toNamespacedPath()",
				Options: []any{map[string]any{"version": "8.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'path.toNamespacedPath' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			// path: upstream invalid 140.
			{Code: "var path = require('path'); path.toNamespacedPath()",
				Options: []any{map[string]any{"version": "8.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'path.toNamespacedPath' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 50},
				},
			},
			// path: upstream invalid 141.
			{Code: "var { toNamespacedPath } = require('path'); toNamespacedPath()",
				Options: []any{map[string]any{"version": "8.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'path.toNamespacedPath' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 23},
				},
			},
			// path: upstream invalid 142.
			{Code: "import path from 'path'; path.toNamespacedPath()",
				Options: []any{map[string]any{"version": "8.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'path.toNamespacedPath' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 47},
				},
			},
			// path: upstream invalid 143.
			{Code: "import { toNamespacedPath } from 'path'; toNamespacedPath()",
				Options: []any{map[string]any{"version": "8.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'path.toNamespacedPath' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 26},
				},
			},
			// perf_hooks: upstream invalid 144.
			{Code: "require('perf_hooks')",
				Options: []any{map[string]any{"version": "8.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'perf_hooks' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			// perf_hooks: upstream invalid 145.
			{Code: "import perf_hooks from 'perf_hooks'",
				Options: []any{map[string]any{"version": "8.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'perf_hooks' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 36},
				},
			},
			// perf_hooks: upstream invalid 146.
			{Code: "import { open } from 'perf_hooks'",
				Options: []any{map[string]any{"version": "8.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'perf_hooks' is still an experimental feature and is not supported until Node.js 8.5.0. The configured version range is '8.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			// process: upstream invalid 147.
			{Code: "process.argv0",
				Options: []any{map[string]any{"version": "6.3.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.argv0' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			// process: upstream invalid 148.
			{Code: "require('process').argv0",
				Options: []any{map[string]any{"version": "6.3.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.argv0' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			// process: upstream invalid 149.
			{Code: "var c = require('process'); c.argv0",
				Options: []any{map[string]any{"version": "6.3.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.argv0' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 36},
				},
			},
			// process: upstream invalid 150.
			{Code: "var { argv0 } = require('process'); argv0",
				Options: []any{map[string]any{"version": "6.3.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.argv0' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 12},
				},
			},
			// process: upstream invalid 151.
			{Code: "import c from 'process'; c.argv0",
				Options: []any{map[string]any{"version": "6.3.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.argv0' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 33},
				},
			},
			// process: upstream invalid 152.
			{Code: "process.channel",
				Options: []any{map[string]any{"version": "7.0.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.channel' is still an experimental feature and is not supported until Node.js 7.1.0. The configured version range is '7.0.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// process: upstream invalid 153.
			{Code: "process.cpuUsage",
				Options: []any{map[string]any{"version": "6.0.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.cpuUsage' is still an experimental feature and is not supported until Node.js 6.1.0. The configured version range is '6.0.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// process: upstream invalid 154.
			{Code: "process.emitWarning",
				Options: []any{map[string]any{"version": "5.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.emitWarning' is still an experimental feature and is not supported until Node.js 6.0.0. The configured version range is '5.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// process: upstream invalid 155.
			{Code: "process.getegid",
				Options: []any{map[string]any{"version": "1.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.getegid' is still an experimental feature and is not supported until Node.js 2.0.0. The configured version range is '1.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// process: upstream invalid 156.
			{Code: "process.geteuid",
				Options: []any{map[string]any{"version": "1.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.geteuid' is still an experimental feature and is not supported until Node.js 2.0.0. The configured version range is '1.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// process: upstream invalid 157.
			{Code: "process.hasUncaughtExceptionCaptureCallback",
				Options: []any{map[string]any{"version": "9.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.hasUncaughtExceptionCaptureCallback' is still an experimental feature and is not supported until Node.js 9.3.0. The configured version range is '9.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
				},
			},
			// process: upstream invalid 158.
			{Code: "process.ppid",
				Options: []any{map[string]any{"version": "9.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.ppid' is still an experimental feature and is not supported until Node.js 9.2.0 (backported: ^8.10.0, ^6.13.0). The configured version range is '9.1.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			// process: upstream invalid 159.
			{Code: "process.release",
				Options: []any{map[string]any{"version": "2.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.release' is still an experimental feature and is not supported until Node.js 3.0.0. The configured version range is '2.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// process: upstream invalid 160.
			{Code: "process.setegid",
				Options: []any{map[string]any{"version": "1.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.setegid' is still an experimental feature and is not supported until Node.js 2.0.0. The configured version range is '1.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// process: upstream invalid 161.
			{Code: "process.seteuid",
				Options: []any{map[string]any{"version": "1.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.seteuid' is still an experimental feature and is not supported until Node.js 2.0.0. The configured version range is '1.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// process: upstream invalid 162.
			{Code: "process.setUncaughtExceptionCaptureCallback",
				Options: []any{map[string]any{"version": "9.2.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'process.setUncaughtExceptionCaptureCallback' is still an experimental feature and is not supported until Node.js 9.3.0. The configured version range is '9.2.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
				},
			},
			// stream: upstream invalid 163.
			{Code: "require('stream').finished()",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'stream.finished' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			// stream: upstream invalid 164.
			{Code: "var hooks = require('stream'); hooks.finished()",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'stream.finished' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 32, EndLine: 1, EndColumn: 46},
				},
			},
			// stream: upstream invalid 165.
			{Code: "var { finished } = require('stream'); finished()",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'stream.finished' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 15},
				},
			},
			// stream: upstream invalid 166.
			{Code: "import stream from 'stream'; stream.finished()",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'stream.finished' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 30, EndLine: 1, EndColumn: 45},
				},
			},
			// stream: upstream invalid 167.
			{Code: "import { finished } from 'stream'; finished()",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'stream.finished' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 18},
				},
			},
			// stream: upstream invalid 168.
			{Code: "require('stream').pipeline()",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'stream.pipeline' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			// trace_events: upstream invalid 169.
			{Code: "require('trace_events')",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'trace_events' is still an experimental feature The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			// trace_events: upstream invalid 170.
			{Code: "import trace_events from 'trace_events'",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'trace_events' is still an experimental feature The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
				},
			},
			// trace_events: upstream invalid 171.
			{Code: "import { createTracing } from 'trace_events'",
				Options: []any{map[string]any{"version": "10.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'trace_events' is still an experimental feature The configured version range is '10.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
				},
			},
			// url: upstream invalid 172.
			{Code: "URL",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'URL' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 4},
				},
			},
			// url: upstream invalid 173.
			{Code: "URLSearchParams",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'URLSearchParams' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// url: upstream invalid 174.
			{Code: "require('url').URL",
				Options: []any{map[string]any{"version": "6.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'url.URL' is still an experimental feature and is not supported until Node.js 7.0.0 (backported: ^6.13.0). The configured version range is '6.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			// url: upstream invalid 175.
			{Code: "var cp = require('url'); cp.URL",
				Options: []any{map[string]any{"version": "6.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'url.URL' is still an experimental feature and is not supported until Node.js 7.0.0 (backported: ^6.13.0). The configured version range is '6.9.9'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 32},
				},
			},
			// url: upstream invalid 176.
			{Code: "var { URL } = require('url');",
				Options: []any{map[string]any{"version": "6.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'url.URL' is still an experimental feature and is not supported until Node.js 7.0.0 (backported: ^6.13.0). The configured version range is '6.9.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 10},
				},
			},
			// url: upstream invalid 177.
			{Code: "import cp from 'url'; cp.URL",
				Options: []any{map[string]any{"version": "6.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'url.URL' is still an experimental feature and is not supported until Node.js 7.0.0 (backported: ^6.13.0). The configured version range is '6.9.9'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 29},
				},
			},
			// url: upstream invalid 178.
			{Code: "import { URL } from 'url'",
				Options: []any{map[string]any{"version": "6.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'url.URL' is still an experimental feature and is not supported until Node.js 7.0.0 (backported: ^6.13.0). The configured version range is '6.9.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 13},
				},
			},
			// url: upstream invalid 179.
			{Code: "require('url').URLSearchParams",
				Options: []any{map[string]any{"version": "7.4.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'url.URLSearchParams' is still an experimental feature and is not supported until Node.js 7.5.0 (backported: ^6.13.0). The configured version range is '7.4.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			// url: upstream invalid 180.
			{Code: "require('url').domainToASCII",
				Options: []any{map[string]any{"version": "7.3.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'url.domainToASCII' is still an experimental feature and is not supported until Node.js 7.4.0 (backported: ^6.13.0). The configured version range is '7.3.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},
			// url: upstream invalid 181.
			{Code: "require('url').domainToUnicode",
				Options: []any{map[string]any{"version": "7.3.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'url.domainToUnicode' is still an experimental feature and is not supported until Node.js 7.4.0 (backported: ^6.13.0). The configured version range is '7.3.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			// util: upstream invalid 182.
			{Code: "require('util').callbackify",
				Options: []any{map[string]any{"version": "8.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.callbackify' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			// util: upstream invalid 183.
			{Code: "var hooks = require('util'); hooks.callbackify",
				Options: []any{map[string]any{"version": "8.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.callbackify' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.", Line: 1, Column: 30, EndLine: 1, EndColumn: 47},
				},
			},
			// util: upstream invalid 184.
			{Code: "var { callbackify } = require('util'); callbackify",
				Options: []any{map[string]any{"version": "8.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.callbackify' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 18},
				},
			},
			// util: upstream invalid 185.
			{Code: "import util from 'util'; util.callbackify",
				Options: []any{map[string]any{"version": "8.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.callbackify' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 42},
				},
			},
			// util: upstream invalid 186.
			{Code: "import { callbackify } from 'util'; callbackify",
				Options: []any{map[string]any{"version": "8.1.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.callbackify' is still an experimental feature and is not supported until Node.js 8.2.0. The configured version range is '8.1.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 21},
				},
			},
			// util: upstream invalid 187.
			{Code: "require('util').formatWithOptions",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.formatWithOptions' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			// util: upstream invalid 188.
			{Code: "require('util').getSystemErrorName",
				Options: []any{map[string]any{"version": "9.6.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.getSystemErrorName' is still an experimental feature and is not supported until Node.js 9.7.0 (backported: ^8.12.0). The configured version range is '9.6.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
				},
			},
			// util: upstream invalid 189.
			{Code: "require('util').inspect.custom",
				Options: []any{map[string]any{"version": "6.5.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.inspect.custom' is still an experimental feature and is not supported until Node.js 6.6.0. The configured version range is '6.5.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			// util: upstream invalid 190.
			{Code: "require('util').inspect.defaultOptions",
				Options: []any{map[string]any{"version": "6.3.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.inspect.defaultOptions' is still an experimental feature and is not supported until Node.js 6.4.0. The configured version range is '6.3.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
				},
			},
			// util: upstream invalid 191.
			{Code: "require('util').isDeepStrictEqual",
				Options: []any{map[string]any{"version": "8.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.isDeepStrictEqual' is still an experimental feature and is not supported until Node.js 9.0.0. The configured version range is '8.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			// util: upstream invalid 192.
			{Code: "require('util').promisify",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.promisify' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// util: upstream invalid 193.
			{Code: "require('util').TextDecoder",
				Options: []any{map[string]any{"version": "8.8.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.TextDecoder' is still an experimental feature and is not supported until Node.js 8.9.0. The configured version range is '8.8.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			// util: upstream invalid 194.
			{Code: "require('util').TextEncoder",
				Options: []any{map[string]any{"version": "8.8.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.TextEncoder' is still an experimental feature and is not supported until Node.js 8.9.0. The configured version range is '8.8.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			// util: upstream invalid 195.
			{Code: "require('util').types",
				Options: []any{map[string]any{"version": "9.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.types' is still an experimental feature and is not supported until Node.js 10.0.0. The configured version range is '9.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			// util: upstream invalid 196.
			{Code: "require('util').styleText",
				Options: []any{map[string]any{"version": "21.7.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'util.styleText' is still an experimental feature and is not supported until Node.js 23.5.0 (backported: ^22.13.0). The configured version range is '21.7.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// util: upstream invalid 197.
			{Code: "require('util').styleText",
				Options: []any{map[string]any{"version": "20.11.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-experimental-till", Message: "The 'util.styleText' is not an experimental feature until Node.js 21.7.0 (backported: ^20.12.0). The configured version range is '20.11.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// v8: upstream invalid 198.
			{Code: "require('v8')",
				Options: []any{map[string]any{"version": "0.12.99"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8' is still an experimental feature and is not supported until Node.js 1.0.0. The configured version range is '0.12.99'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			// v8: upstream invalid 199.
			{Code: "import hooks from 'v8'",
				Options: []any{map[string]any{"version": "0.12.99"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8' is still an experimental feature and is not supported until Node.js 1.0.0. The configured version range is '0.12.99'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			// v8: upstream invalid 200.
			{Code: "import { cachedDataVersionTag } from 'v8'",
				Options: []any{map[string]any{"version": "0.12.99"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8' is still an experimental feature and is not supported until Node.js 1.0.0. The configured version range is '0.12.99'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
					{MessageId: "not-supported-till", Message: "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '0.12.99'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
				},
			},
			// v8: upstream invalid 201.
			{Code: "require('v8').cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
				},
			},
			// v8: upstream invalid 202.
			{Code: "var hooks = require('v8'); hooks.cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 28, EndLine: 1, EndColumn: 54},
				},
			},
			// v8: upstream invalid 203.
			{Code: "var { cachedDataVersionTag } = require('v8'); cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 27},
				},
			},
			// v8: upstream invalid 204.
			{Code: "import v8 from 'v8'; v8.cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 45},
				},
			},
			// v8: upstream invalid 205.
			{Code: "import { cachedDataVersionTag } from 'v8'; cachedDataVersionTag()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.cachedDataVersionTag' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
				},
			},
			// v8: upstream invalid 206.
			{Code: "require('v8').getHeapSpaceStatistics()",
				Options: []any{map[string]any{"version": "5.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.getHeapSpaceStatistics' is still an experimental feature and is not supported until Node.js 6.0.0. The configured version range is '5.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},
			// v8: upstream invalid 207.
			{Code: "require('v8').serialize()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.serialize' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			// v8: upstream invalid 208.
			{Code: "require('v8').deserialize()",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.deserialize' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// v8: upstream invalid 209.
			{Code: "require('v8').Serializer",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.Serializer' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			// v8: upstream invalid 210.
			{Code: "require('v8').Deserializer",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.Deserializer' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			// v8: upstream invalid 211.
			{Code: "require('v8').DefaultSerializer",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.DefaultSerializer' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			// v8: upstream invalid 212.
			{Code: "require('v8').DefaultDeserializer",
				Options: []any{map[string]any{"version": "7.9.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'v8.DefaultDeserializer' is still an experimental feature and is not supported until Node.js 8.0.0. The configured version range is '7.9.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			// vm: upstream invalid 213.
			{Code: "require('vm').Module",
				Options: []any{map[string]any{"version": "9.5.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'vm.Module' is still an experimental feature The configured version range is '9.5.9'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// vm: upstream invalid 214.
			{Code: "var vm = require('vm'); vm.Module",
				Options: []any{map[string]any{"version": "9.5.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'vm.Module' is still an experimental feature The configured version range is '9.5.9'.", Line: 1, Column: 25, EndLine: 1, EndColumn: 34},
				},
			},
			// vm: upstream invalid 215.
			{Code: "var { Module } = require('vm'); Module",
				Options: []any{map[string]any{"version": "9.5.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'vm.Module' is still an experimental feature The configured version range is '9.5.9'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// vm: upstream invalid 216.
			{Code: "import vm from 'vm'; vm.Module",
				Options: []any{map[string]any{"version": "9.5.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'vm.Module' is still an experimental feature The configured version range is '9.5.9'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 31},
				},
			},
			// vm: upstream invalid 217.
			{Code: "import { Module } from 'vm'; Module",
				Options: []any{map[string]any{"version": "9.5.9"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'vm.Module' is still an experimental feature The configured version range is '9.5.9'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 16},
				},
			},
			// worker_threads: upstream invalid 218.
			{Code: "require('worker_threads')",
				Options: []any{map[string]any{"version": "10.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'worker_threads' is still an experimental feature and is not supported until Node.js 12.11.0. The configured version range is '10.5.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// worker_threads: upstream invalid 219.
			{Code: "import worker_threads from 'worker_threads'",
				Options: []any{map[string]any{"version": "10.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'worker_threads' is still an experimental feature and is not supported until Node.js 12.11.0. The configured version range is '10.5.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
				},
			},
			// worker_threads: upstream invalid 220.
			{Code: "import { Worker } from 'worker_threads'",
				Options: []any{map[string]any{"version": "10.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'worker_threads' is still an experimental feature and is not supported until Node.js 12.11.0. The configured version range is '10.5.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
				},
			},
			// worker_threads: upstream invalid 221.
			{Code: "import { Worker } from 'worker_threads'",
				Settings: map[string]any{"node": map[string]any{"version": "10.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'worker_threads' is still an experimental feature and is not supported until Node.js 12.11.0. The configured version range is '10.5.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
				},
			},
			// timers/promises: upstream invalid 222.
			{Code: "\n                        import { scheduler } from 'node:timers/promises';\n                        await scheduler.wait( 1000 );\n                    ",
				Options: []any{map[string]any{"version": ">= 20.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'timers/promises.scheduler.wait' is still an experimental feature The configured version range is '>= 20.0.0'.", Line: 3, Column: 31, EndLine: 3, EndColumn: 45},
				},
			},
			// import.meta: upstream invalid 223.
			{Code: "import.meta.resolve(specifier)",
				Options: []any{map[string]any{"version": "20.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'import.meta.resolve' is still an experimental feature and is not supported until Node.js 20.6.0 (backported: ^18.19.0). The configured version range is '20.5.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// import.meta: upstream invalid 224.
			{Code: "import.meta.resolve(specifier)",
				Options: []any{map[string]any{"version": "19.8.1"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'import.meta.resolve' is still an experimental feature and is not supported until Node.js 20.6.0 (backported: ^18.19.0). The configured version range is '19.8.1'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// import.meta: upstream invalid 225.
			{Code: "import.meta.resolve(specifier)",
				Options: []any{map[string]any{"version": "18.18.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'import.meta.resolve' is still an experimental feature and is not supported until Node.js 20.6.0 (backported: ^18.19.0). The configured version range is '18.18.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// import.meta: upstream invalid 226.
			{Code: "import.meta.resolve(specifier)",
				Options: []any{map[string]any{"version": "13.8.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-experimental-till", Message: "The 'import.meta.resolve' is not an experimental feature until Node.js 13.9.0 (backported: ^12.16.2). The configured version range is '13.8.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// import.meta: upstream invalid 227.
			{Code: "import.meta.resolve(specifier)",
				Options: []any{map[string]any{"version": "12.15.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-experimental-till", Message: "The 'import.meta.resolve' is not an experimental feature until Node.js 13.9.0 (backported: ^12.16.2). The configured version range is '12.15.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// import.meta: upstream invalid 228.
			{Code: "import.meta.dirname;",
				Options: []any{map[string]any{"version": "22.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'import.meta.dirname' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '22.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// import.meta: upstream invalid 229.
			{Code: "import.meta.dirname;",
				Options: []any{map[string]any{"version": "21.1.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'import.meta.dirname' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '21.1.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// import.meta: upstream invalid 230.
			{Code: "import.meta.dirname;",
				Options: []any{map[string]any{"version": "20.10.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'import.meta.dirname' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '20.10.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// import.meta: upstream invalid 231.
			{Code: "import.meta.dirname;",
				Options: []any{map[string]any{"version": "21.1.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-experimental-till", Message: "The 'import.meta.dirname' is not an experimental feature until Node.js 21.2.0 (backported: ^20.11.0). The configured version range is '21.1.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// import.meta: upstream invalid 232.
			{Code: "import.meta.dirname;",
				Options: []any{map[string]any{"version": "20.10.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-experimental-till", Message: "The 'import.meta.dirname' is not an experimental feature until Node.js 21.2.0 (backported: ^20.11.0). The configured version range is '20.10.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// import.meta: upstream invalid 233.
			{Code: "import.meta.filename;",
				Options: []any{map[string]any{"version": "22.0.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'import.meta.filename' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '22.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// import.meta: upstream invalid 234.
			{Code: "import.meta.filename;",
				Options: []any{map[string]any{"version": "21.1.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'import.meta.filename' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '21.1.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// import.meta: upstream invalid 235.
			{Code: "import.meta.filename;",
				Options: []any{map[string]any{"version": "20.10.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'import.meta.filename' is still an experimental feature and is not supported until Node.js 24.0.0 (backported: ^22.16.0). The configured version range is '20.10.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// import.meta: upstream invalid 236.
			{Code: "import.meta.filename;",
				Options: []any{map[string]any{"version": "21.1.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-experimental-till", Message: "The 'import.meta.filename' is not an experimental feature until Node.js 21.2.0 (backported: ^20.11.0). The configured version range is '21.1.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// import.meta: upstream invalid 237.
			{Code: "import.meta.filename;",
				Options: []any{map[string]any{"version": "20.10.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-experimental-till", Message: "The 'import.meta.filename' is not an experimental feature until Node.js 21.2.0 (backported: ^20.11.0). The configured version range is '20.10.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// fetch: upstream invalid 238.
			{Code: "fetch('/asd')",
				Options: []any{map[string]any{"version": "16.0.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-experimental-till", Message: "The 'fetch' is not an experimental feature until Node.js 17.5.0 (backported: ^16.15.0). The configured version range is '16.0.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// fetch: upstream invalid 239.
			{Code: "fetch('/asd')",
				Options: []any{map[string]any{"version": "16.16.0", "allowExperimental": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'fetch' is still an experimental feature and is not supported until Node.js 21.0.0. The configured version range is '16.16.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// sqlite: upstream invalid 240.
			{Code: "\n                        import { DatabaseSync } from 'node:sqlite';\n                        const database = new DatabaseSync(':memory:');\n                    ",
				Options: []any{map[string]any{"version": ">=22.5.0"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-yet", Message: "The 'sqlite' is still an experimental feature The configured version range is '>=22.5.0'.", Line: 2, Column: 25, EndLine: 2, EndColumn: 68},
				},
			},
			// sqlite: upstream invalid 241.
			{Code: "\n                        import { DatabaseSync } from 'node:sqlite';\n                        const database = new DatabaseSync(':memory:');\n                    ",
				Options: []any{map[string]any{"version": ">=22.3.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-experimental-till", Message: "The 'sqlite' is not an experimental feature until Node.js 22.5.0. The configured version range is '>=22.3.0'.", Line: 2, Column: 25, EndLine: 2, EndColumn: 68},
					{MessageId: "not-supported-till", Message: "The 'sqlite.DatabaseSync' is still an experimental feature and is not supported until Node.js 22.5.0. The configured version range is '>=22.3.0'.", Line: 2, Column: 34, EndLine: 2, EndColumn: 46},
				},
			},
			// sqlite: upstream invalid 242.
			{Code: "\n                        const { DatabaseSync } = require('node:sqlite');\n                        const database = new DatabaseSync(':memory:');\n                    ",
				Options: []any{map[string]any{"version": ">=22.3.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'sqlite.DatabaseSync' is still an experimental feature and is not supported until Node.js 22.5.0. The configured version range is '>=22.3.0'.", Line: 2, Column: 33, EndLine: 2, EndColumn: 45},
					{MessageId: "not-experimental-till", Message: "The 'sqlite' is not an experimental feature until Node.js 22.5.0. The configured version range is '>=22.3.0'.", Line: 2, Column: 50, EndLine: 2, EndColumn: 72},
				},
			},
			// sqlite: upstream invalid 243.
			{Code: "\n                        const { DatabaseSync } = process.getBuiltinModule('node:sqlite');\n                        const database = new DatabaseSync(':memory:');\n                    ",
				Options: []any{map[string]any{"version": ">=22.3.0", "allowExperimental": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "not-supported-till", Message: "The 'sqlite.DatabaseSync' is still an experimental feature and is not supported until Node.js 22.5.0. The configured version range is '>=22.3.0'.", Line: 2, Column: 33, EndLine: 2, EndColumn: 45},
					{MessageId: "not-experimental-till", Message: "The 'sqlite' is not an experimental feature until Node.js 22.5.0. The configured version range is '>=22.3.0'.", Line: 2, Column: 50, EndLine: 2, EndColumn: 89},
				},
			},
		},
	)
}
