// TestNoDeprecatedApiExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
package no_deprecated_api_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/n/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/n/rules/no_deprecated_api"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDeprecatedApiExtras(t *testing.T) {
	v6 := map[string]interface{}{"version": "6.0.0"}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_deprecated_api.NoDeprecatedApiRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: graceful degradation (must not crash / mask) ----
			{Code: `var {} = require('buffer')`},              // empty object pattern
			{Code: `var {...rest} = require('buffer'); rest`}, // rest element only
			{Code: `var [a] = require('buffer'); a`},          // array pattern, no Buffer
			// Array pattern / array-literal binding isn't tracked — upstream's
			// _iterateLhsReferences has no ArrayPattern arm and _iteratePropertyReferences
			// stops at an ArrayExpression parent. We mirror that (no report).
			{Code: `var [b] = [require('buffer').Buffer]; new b()`, Options: v6},
			{Code: ``},                   // empty file
			{Code: `require('buffer');`}, // module READ that isn't deprecated

			// ---- Locks in iterateGlobalReferences isModifiedGlobal: declared global ----
			{Code: `function f(Buffer) { return new Buffer() }`},          // Buffer is a parameter → shadowed
			{Code: `function f(require) { return require('fs').exists }`}, // require shadowed
			{Code: `{ const require = (x) => x; require('fs').exists; }`}, // block-scoped require
			{Code: `var Buffer = null; new Buffer()`, Options: v6},        // top-level var shadows global

			// ---- Locks in iterateGlobalReferences isModifiedGlobal: written global ----
			{Code: `Buffer = function(){}; new Buffer()`, Options: v6}, // global Buffer written → whole name skipped
			// destructuring writes to a global also mark it modified (utils.IsWriteReference covers patterns)
			{Code: `[Buffer] = [function(){}]; new Buffer()`, Options: v6},
			{Code: `({Buffer} = {Buffer: function(){}}); new Buffer()`, Options: v6},

			// ---- Locks in compareNodes: non-constant require argument ----
			{Code: `require(mod).exists`},        // dynamic module name
			{Code: `require('fs')[name].exists`}, // dynamic element key on the chain

			// ---- Real-user: shadowed builtin name used as a plain local ----
			{Code: `function get(url) { return url.parse }`}, // url is a parameter, not the module

			// ---- Real-user: re-export that doesn't touch a deprecated member ----
			{Code: `export { request } from 'http'`},

			// ---- Real-user(n docs): parameter shadowing of a require-binding is out of scope ----
			{Code: `var fs = require('fs'); function g(fs) { return fs.lchmod }`, Options: map[string]interface{}{"version": "0.4.0"}},
			// ---- Real-user(n#350): non-deprecated API must not be flagged ----
			{Code: `process.nextTick(() => {})`, Options: v6},
			// ---- Real-user(n#65): trailing-slash specifier is userland, not core ----
			{Code: `require('punycode/');`, Options: map[string]interface{}{"version": "7.0.0"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver (tsgo keeps ParenthesizedExpression) ----
			{
				Code:   `(require('fs')).exists;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:   `((require('fs'))).exists;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: element access on the chain (a['x'] form) ----
			{
				Code:   `require('fs')['exists'];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: TS type-expression wrappers (as / non-null) ----
			{
				Code:   `(require('fs') as any).exists;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:   `require('fs')!.exists;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: optional chain (tsgo flag, no ChainExpression wrapper) ----
			{
				Code:   `require('fs')?.exists;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1}},
			},

			// ---- Locks in isPassThrough: comma sequence (last operand) ----
			{
				Code:    `(0, require('buffer')).Buffer();`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1}},
			},

			// ---- Locks in isPassThrough: conditional arm ----
			{
				Code:    `(0 ? 0 : require('buffer')).Buffer();`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1}},
			},

			// ---- Locks in iterateGlobalReferences globalObjectNames: global.X / globalThis.X ----
			{
				Code:    `global.Buffer();`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `globalThis.Buffer();`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'Buffer()' was deprecated since v6.0.0. Use 'Buffer.alloc()' or 'Buffer.from()' instead.", Line: 1, Column: 1}},
			},

			// ---- Locks in module path surviving a shadowed global (isModifiedGlobal) ----
			// Buffer is declared locally (no global match), but the require chain
			// still tracks `new buffer.Buffer()`.
			{
				Code:    `var Buffer = require('buffer').Buffer; new Buffer()`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 40}},
			},

			// ---- Locks in iteratePropertyReferences Parameter default value ----
			{
				Code:    `function f(b = require('buffer').Buffer) { return new b() }`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 51}},
			},

			// ---- Locks in version filtering: lower target drops not-yet-available replacement ----
			// fs.exists alternatives: fs.stat()@0.0.2, fs.access()@0.11.15. On Node 0.10,
			// fs.access isn't available yet, so only fs.stat() is suggested.
			{
				Code:    `require('fs').exists;`,
				Options: map[string]interface{}{"version": "0.10.0"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' instead.", Line: 1, Column: 1}},
			},

			// ---- Locks in node: prefix stripping in the reported name ----
			{
				Code:    `require('node:buffer').SlowBuffer;`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 1}},
			},

			// ---- Locks in graceful degradation: rest element doesn't mask sibling ----
			{
				Code:    `var {SlowBuffer, ...rest} = require('buffer'); rest;`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 6}},
			},

			// ---- Real-user: deep re-export of a deprecated member (export { x } from) ----
			{
				Code:   `export { parse } from 'url';`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'url.parse' was deprecated since v11.0.0. Use 'url.URL' constructor instead.", Line: 1, Column: 10}},
			},
			// ---- Export specifier with rename: covers the ExportSpecifier PropertyName arm ----
			{
				Code:   `export { parse as p } from 'url';`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'url.parse' was deprecated since v11.0.0. Use 'url.URL' constructor instead.", Line: 1, Column: 10}},
			},

			// ---- Real-user: namespace default access (import * as b; b.default.X) ----
			{
				Code:    `import * as b from 'buffer'; new b.default.Buffer()`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 30}},
			},

			// ---- Nesting: multi-hop intermediate variable (a -> b -> new b.Buffer()) ----
			{
				Code:    `var a = require('buffer'); var b = a; new b.Buffer()`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 39}},
			},
			// ---- Nesting: require inside a function body (global require not shadowed) ----
			{
				Code:   `function f() { return require('fs').exists }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 23}},
			},
			// ---- Nesting: assignment-destructuring LHS (ObjectLiteral) of a require result ----
			{
				Code:    `var b; ({Buffer: b} = require('buffer')); new b()`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'new buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 43}},
			},
			// ---- Constant folding: string `+` concat in module name / element key (aligned with ESLint) ----
			{
				Code:   `require('f' + 's').exists`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1}},
			},
			{
				Code:    `require('buffer')['Slow' + 'Buffer'];`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 1}},
			},
			// ---- tsgo wrapper: satisfies (alongside as / non-null already covered) ----
			{
				Code:   `(require('fs') satisfies any).exists`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1}},
			},
			// ---- Nesting: closure captures the require-binding ----
			{
				Code:   `var fs = require('fs'); function g() { return fs.exists }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 47}},
			},
			// ---- Nesting: element-access on an intermediate variable ----
			{
				Code:    `var b = require('buffer'); b['SlowBuffer']`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'buffer.SlowBuffer' was deprecated since v6.0.0. Use 'buffer.Buffer.allocUnsafeSlow()' instead.", Line: 1, Column: 28}},
			},
			// ---- Nesting: reassign binding to a different module, then access its deprecated member ----
			{
				Code:   `var b = require('buffer'); b = require('fs'); b.exists`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 47}},
			},
			// ---- tsgo optional-chain depth: require('fs')?.exists?.bind ----
			{
				Code:   `require('fs')?.exists?.bind`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'fs.exists' was deprecated since v4.0.0. Use 'fs.stat()' or 'fs.access()' instead.", Line: 1, Column: 1}},
			},
			// ---- Real-user: safe-buffer.Buffer is a re-export, deprecated like buffer.Buffer ----
			{
				Code:    `require('safe-buffer').Buffer()`,
				Options: v6,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "'safe-buffer.Buffer()' was deprecated since v6.0.0. Use 'buffer.Buffer.alloc()' or 'buffer.Buffer.from()' instead.", Line: 1, Column: 1}},
			},
		},
	)
}
