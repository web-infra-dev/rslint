package utils_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

var expectCustomizationProbe = rule.Rule{
	Name: "rstest/expect-customization-probe",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := node.Expression()
				if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "probe" {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id: "customization",
					Description: fmt.Sprintf(
						"toBe=%t toBeDefined=%t equality=%t",
						analysis.IsExpectMatcherOverridden("toBe"),
						analysis.IsExpectMatcherOverridden("toBeDefined"),
						analysis.HasCustomEqualityTesters(),
					),
				})
			},
		}
	},
}

func TestRstestExpectCustomizationAnalysis(t *testing.T) {
	want := func(message string) []rule_tester.InvalidTestCaseError {
		return []rule_tester.InvalidTestCaseError{{MessageId: "customization", Message: message}}
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectCustomizationProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{Code: `expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; check['extend']({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: "const check = expect; check[`addEqualityTesters`]([tester]); probe();", Errors: want("toBe=false toBeDefined=false equality=true")},
			{Code: `let check = expect; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; const again = (check as any)!; again['extend']({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; const extend = check.extend; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; const { extend } = check; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import { expect as original } from '@rstest/core'; const check = original; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const check = core['expect']; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const core = require('rstack/test'); const check = core.expect; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = import.meta.rstest.expect; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `test('case', ctx => { const check = ctx.expect; check.extend({ toBe() {} }); }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; check.addEqualityTesters([tester]); probe();`, Errors: want("toBe=false toBeDefined=false equality=true")},
			{Code: `const check = expect; const { addEqualityTesters: add } = check; add([tester]); probe();`, Errors: want("toBe=false toBeDefined=false equality=true")},
			{Code: `const check = expect; check.addEqualityTesters([]); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `let check = expect; check = other; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const check = other; const other = check; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `let check; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const check = expect(); check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const check = expect.soft; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const expect = 'other'; const check = core[expect]; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const check = import.meta.rstest.expect(); check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import { expect as original } from 'vitest'; const check = original; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `function configure(expect: any) { const check = expect; check.extend({ toBe() {} }); } probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const check = expect; { const check = other; check.extend({ toBe() {} }); } probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `expect.extend({ ["toBe"]() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `expect.extend({ ...matchers }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `expect.extend(matchers); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `expect.addEqualityTesters([]); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `expect.addEqualityTesters([tester]); probe();`, Errors: want("toBe=false toBeDefined=false equality=true")},
			{Code: `import { expect as check } from "@rstest/core"; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from "rstack/test"; core.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import.meta.rstest.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const extend = expect.extend; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let extend = expect.extend; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from "@rstest/core"; const extend = core.expect.extend; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const add = import.meta.rstest.expect.addEqualityTesters; add([]); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const { extend } = expect; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let { extend } = expect; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let extend = expect.extend; extend = other; extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `let { extend } = expect; extend = other; extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from "@rstest/core"; const { addEqualityTesters: add } = core.expect; add([]); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `test("case", ({ expect }) => { expect.extend({ toBe() {} }); }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import { expect as check } from "@rstest/core"; check.addEqualityTesters([tester]); probe();`, Errors: want("toBe=false toBeDefined=false equality=true")},
			{Code: `import { expect as check } from "vitest"; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `function configure(expect: any) { expect.extend({ toBe() {} }); } probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},

			// Folded static keys, destructuring, shallow copies, and non-null assertions.
			{Code: `import.meta.rstest!.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const api = 'expect'; const register = 'ex' + 'tend'; core[api][register]({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { expect: check } = core; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `test('setup', ({ expect: { extend } }) => { extend({ toBe() {} }); }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { ...api } = core; Object.freeze(api); api.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...copy } = expect; copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...mock } = expect; mock.extend = () => {}; mock.extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `expect.addEqualityTesters([])?.foo(1); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `expect(expect.extend).toBeDefined(); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; var register = r.expect.extend; register({ toBe() {} }); var register = () => {}; probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const options = { method: 'extend', verbose: false }; function enableVerbose(x: any) { x.verbose = true; } enableVerbose(options); const method = options.method; r.expect[method]({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=true")},
			{Code: `import * as r from '@rstest/core'; r[getKey()].extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},

			// Escaped references to `expect` / `expect.extend` / `expect.addEqualityTesters`
			// conservatively fall back to marking customization active instead of
			// attempting to symbolically execute getters, setters, `.bind()`, or helpers.
			{Code: `expect.extend.bind(expect, { toBe() {} })(); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			{Code: `Reflect.apply(expect.extend, expect, [{ toBe() {} }]); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			{Code: `const copy = Object.assign({}, expect); copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			{Code: `import * as core from '@rstest/core'; let extend2: any; ({ expect: { extend: extend2 } = {} } = core); extend2({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			{Code: `import * as r from '@rstest/core'; const noop = (_m: any) => {}; const { prepare, register } = { register: noop, set enabled(v: boolean) { if (v) (this as any).register = r.expect.extend; }, configure(v: boolean) { this.enabled = v; }, get prepare() { this.configure(true); return 1; } }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `import * as r from '@rstest/core'; const noop = (_m: any) => {}; const { prepare, register } = { register: noop, configure(_x = ((this as any).register = r.expect.extend)) { return 1; }, get prepare() { this.configure(); return 1; } }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `import * as r from '@rstest/core'; const noop = (_m: any) => {}; const { prepare, register } = { register: noop, configure({ extend } = r.expect) { (this as any).register = extend; }, get prepare() { this.configure(); return 1; } }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			// Objects with getters/setters that do not reference `expect.extend` remain clean.
			{Code: `import * as r from '@rstest/core'; const noop = (_m: any) => {}; const { prepare, register } = { register: noop, set last(v: string) { (this as any)._last = v; }, log(msg: string) { this.last = msg; }, get prepare() { this.log('ready'); return 1; } }; register({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},

			// Red-team FN regressions: ambient/type-only expect, receiver escapes, hoisted kill bypasses, cyclic aliases, globalThis/import.meta/dynamic import/rest ctx, nested destructuring, callback matchers.
			{Code: `import type { expect } from '@rstest/core'; const { extend } = expect; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `declare const expect: any; const ext = expect.extend; ext({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `export {}; declare global { const expect: any; } const check = expect; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let extend2: any; ({ expect: { extend: extend2 } = {} } = require('@rstest/core')); extend2({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			{Code: `import * as r from '@rstest/core'; const obj = { r }; obj.r.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			{Code: `import * as r from '@rstest/core'; const c = { ...r }; c.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			{Code: `import * as r from '@rstest/core'; let c: any; c ??= r; c.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let check = expect; run(); check = other; function run() { check.extend({ toBe() {} }); } probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let extend = expect.extend; extend = run(); function run() { extend({ toBe() {} }); return other; } probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `done: { var check = expect; break done; check = other; } check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...mock } = expect; run(); mock.extend = () => {}; function run() { mock.extend({ toBe() {} }); } probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `let a = expect; a = a; a.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `var a: any = expect; var b = a; if (false) { var a = b; } b.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `globalThis.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `window.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { expect: e } = globalThis; e.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `(import.meta as any).rstest.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import.meta['rs' + 'test'].expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `async function f() { const { expect } = await import('@rstest/core'); expect.extend({ toBe() {} }); } probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `test('t', (...args) => { args[0].expect.extend({ toBe() {} }); }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `test('t', (...[ctx]) => { ctx.expect.extend({ toBe() {} }); }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `expect.extend({ toBe() {} }); if (false) { expect = other; } probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			{Code: `expect(expect.extend).toEqual({ asymmetricMatch(ext: any) { ext({ toBe() {} }); return true; } }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `(expect(expect) as any)._obj.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			{Code: `const { extend: { call } } = expect; call(null, { toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `test('t', ({ expect: { extend: { apply } } }) => { apply(null, [{ toBe() {} }]); }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},

			// Red-team FP regressions: realistic truthiness/type guards, JSX/module non-value identifiers, expect.soft, local reassignment targets.
			{Code: `const check = expect; if (check) { check(1).toBe(1); } probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `if (expect.extend) {} probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `'extend' in expect; expect instanceof Function; expect && expect(1).toBe(1); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `export * as expect from './helpers'; probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import data from './x.json' with { expect: '1' }; probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const el = <MyComponent expect={1} />; const el2 = <expect />; probe();`, Tsx: true, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `expect.soft(expect.extend).toBeDefined(); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `let check = expect; check(1).toBe(1); check = other; probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `let ext = expect.extend; ext({ toFoo() {} }); ext = other; probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `test('t', () => expect(expect.extend).toBeDefined()); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},

			// Round 2 red-team regressions: assignment-expr escapes, destructuring-assignment defaults, member-access cycles, sibling destructuring/comma kills, Symbol.hasInstance, overridden safe matchers, rest-array/import.meta/globalThis destructuring, non-instantiated namespace expect.
			{Code: `let tmp: any; const obj = { fn: (tmp = expect.extend) }; obj.fn({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `let tmp: any; (tmp = expect.extend)({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `let ext: any; ({ a: ext = expect.extend } = {}); ext({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `let e: any; ([e = expect] = []); e.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			{Code: `let a: any = expect; a = a.extend; a({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `import * as r from '@rstest/core'; let a: any = r; let b = a.expect; a = b; a.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let { extend, missing = run() } = expect as any; extend = other; function run() { extend({ toBe() {} }); return 1; } probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { expect: { ...mock }, missing = run() } = core as any; mock.extend = () => {}; function run() { mock.extend({ toBe() {} }); return 1; } probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `const { ...mock } = expect; mock[(run(), 'extend')] = () => {}; function run() { mock.extend({ toBe() {} }); } probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `const trap = { [Symbol.hasInstance](e: any) { e.extend({ toBe() {} }); return true; } }; expect instanceof (trap as any); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
			{Code: `const trap = { [Symbol.hasInstance](ext: any) { ext({ toBe() {} }); return true; } }; expect(expect.extend).toBeInstanceOf(trap as any); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `expect.extend({ toBeDefined(ext: any) { ext({ toBe() {} }); return { pass: true, message: () => '' }; } }); expect(expect.extend).toBeDefined(); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `function check() { expect(expect.extend).toBeDefined(); } expect.extend({ toBeDefined(ext: any) { ext({ toBe() {} }); return { pass: true, message: () => '' }; } }); check(); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `test('t', (...args) => { const { 0: ctx } = args; ctx.expect.extend({ toBe() {} }); }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `test('t', (...args) => { const { ...copy } = args; copy[0].expect.extend({ toBe() {} }); }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...m } = import.meta; m.rstest.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...g } = globalThis; g.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `namespace expect { export type T = number; } const check = expect; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let v0: any, v1: any, v2: any, v3: any, v4: any, v5: any, v6: any, v7: any, v8: any, v9: any, v10: any, v11: any, v12: any, v13: any, v14: any, v15: any, v16: any, v17: any, v18: any; if (false) { v0 = v1.foo; v1 = v2.foo; v2 = v3.foo; v3 = v4.foo; v4 = v5.foo; v5 = v6.foo; v6 = v7.foo; v7 = v8.foo; v8 = v9.foo; v9 = v10.foo; v10 = v11.foo; v11 = v12.foo; v12 = v13.foo; v13 = v14.foo; v14 = v15.foo; v15 = v16.foo; v16 = v17.foo; v17 = v18.foo; v18 = v0.foo; } v1 = expect; v2 = v1; v3 = v2; v4 = v3; v5 = v4; v6 = v5; v7 = v6; v8 = v7; v9 = v8; v10 = v9; v11 = v10; v12 = v11; v13 = v12; v14 = v13; v15 = v14; v16 = v15; v17 = v16; v18 = v17; v18.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `function check() { expect(expect).toBeTruthy(); } expect.extend({ toBeDefined(ext: any) { ext({ toBeTruthy(e: any) { e.addEqualityTesters([tester]); return { pass: true, message: () => '' }; } }); return { pass: true, message: () => '' }; } }); expect(expect.extend).toBeDefined(); check(); probe();`, Errors: want("toBe=true toBeDefined=true equality=true")},
		},
	)
}

func TestRstestExpectCustomizationLinearScaling(t *testing.T) {
	var b strings.Builder
	b.WriteString("import * as core from '@rstest/core';\n")
	for i := range 500 {
		fmt.Fprintf(&b, "const alias%d = core.expect;\n", i)
		fmt.Fprintf(&b, "alias%d(1).toBe(1);\n", i)
	}
	for i := range 500 {
		fmt.Fprintf(&b, "let v%d = %d;\n", i, i)
		fmt.Fprintf(&b, "v%d = %d;\n", i, i+1)
	}
	// 30-variable branching cycle (would be O(2^30) without Tarjan SCC memoization).
	for i := range 30 {
		fmt.Fprintf(&b, "let cyc%d: any;\n", i)
	}
	for i := range 29 {
		fmt.Fprintf(&b, "cyc%d = cyc%d;\n", i, i+1)
		if i+2 < 30 {
			fmt.Fprintf(&b, "cyc%d = cyc%d;\n", i, i+2)
		}
	}
	b.WriteString("cyc29 = cyc0;\n")
	b.WriteString("core.expect.extend({ toBe() {} });\nprobe();\n")

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(b.String(), "scaling.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []rule.RuleDiagnostic
	start := time.Now()
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     lintprogram.NewFromCompiler(program),
		File:        sourceFile.FileName(),
		HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     expectCustomizationProbe.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return expectCustomizationProbe.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandNone,
			Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			},
		},
	})
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("took %v; expected linear-time execution", elapsed)
	}
	if len(diagnostics) != 1 || diagnostics[0].Message.Description != "toBe=true toBeDefined=false equality=false" {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
}
