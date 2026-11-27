package utils_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
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
			{Code: `import * as core from '@rstest/core'; const expect = 'expect'; core[expect].extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const key = 'expect'; core[key].extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const expect = 'expect'; const extend = 'extend'; core[expect][extend]({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const api = 'expect'; const register = 'ex' + 'tend'; core[api][register]({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { expect: check } = core; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { expect: { extend } } = core; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const api = 'expect'; const register = 'ex' + 'tend'; const { [api]: { [register]: reg } } = core; reg({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const ns = core; const { expect: check } = ns; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { [getKey()]: check } = core; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { [Symbol.iterator]: check } = core; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { other: check } = core; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; let { expect: check } = core; check = other; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const config = { name: 'other' }; function mutate(x: any) { x.name = 'expect'; } mutate(config); const key = config.name; core[key].extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const options = { method: 'extend', verbose: false }; function enableVerbose(x: any) { x.verbose = true; } enableVerbose(options); const method = options.method; core.expect[method]({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; core.expect[getKey()](1); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; var register = core.expect.extend; register({ toBe() {} }); var register = () => {}; probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; var register = () => {}; register({ toBe() {} }); var register = core.expect.extend; probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import { expect } from '@rstest/core'; const { ...mock } = expect; mock.extend = () => {}; mock.extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import { expect } from '@rstest/core'; const { ...mock } = expect; Object.assign(mock, { extend() {} }); mock.extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import { expect } from '@rstest/core'; const { ...copy } = expect; copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { ...api } = core; Object.freeze(api); api.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { ...api } = core; Object.keys(api); api.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import { expect } from '@rstest/core'; const { ...copy } = expect; (copy as any).tag = 'v1'; copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; let install: any; install?.({ toBeDefined() {} }); install = core.expect.extend; probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; let install: any; function setup() { install = core.expect.extend; } install?.({ toBeDefined() {} }); setup(); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; let install: any; function setup() { install = core.expect.extend; } setup(); install({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import { createRequire } from 'node:module'; const require = createRequire(import.meta.url); require('@rstest/core').expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `declare const require: any; require('@rstest/core').expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `function run(require: any) { require('@rstest/core').expect.extend({ toBeDefined() {} }); } probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const require = () => ({ expect: { extend() {} } }); require('@rstest/core').expect.extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; core[getKey()].extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; let key = 'other'; key = 'expect'; core[key].extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import.meta.rstest[getKey()].extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const core = import.meta.rstest!; core.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { expect: check } = require('@rstest/core'); check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `expect.extend({ ['to' + 'Be']() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const sym = Symbol.iterator; core[sym].extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const sym = Symbol.for('x'); core[sym].extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
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
			{Code: `import * as core from '@rstest/core'; let install: any; setup(); install({ toBe() {} }); function setup() { install = core.expect.extend; } probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; let install: any; install?.({ toBeDefined() {} }); setup(); function setup() { install = core.expect.extend; } probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; let install: any; function setup() { install = core.expect.extend; } function init() { setup(); } init(); install({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; let install: any; function setup() { install = core.expect.extend; } function init() { setup(); } install?.({ toBeDefined() {} }); init(); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `let register: any = expect.extend; if (false) register = () => {}; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let register: any = expect.extend; register ??= () => {}; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let register: any; register ??= expect.extend; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let extend: any; ({ extend } = expect); extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const copy = { ...expect }; copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const mock = { ...expect, extend: () => {} }; mock.extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const { ...copy } = expect; Object.assign(copy, { tag: 'v1' }); copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { ...api } = core; (api.expect as any).tag = 'v1'; api.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...mock } = expect; delete (mock.extend); mock.extend?.({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `test('setup', ({ expect: { extend } }) => { extend({ toBe() {} }); }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `(0, expect.extend)({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `expect.extend.call(expect, { toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `expect.extend.apply(expect, [{ toBe() {} }]); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `expect.extend.bind(expect)({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const { [{ __proto__: ['expect'] } as any]: check } = core; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...rstest } = require('@rstest/core'); rstest.extend?.({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const { ...rstest } = require('@rstest/core'); rstest.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `class C { run = () => fn({ toBe() {} }); } let fn: any = () => {}; fn = expect.extend; new C().run(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Cycle detection must not poison alias caches.
			{Code: `let a: any = expect.extend; let b: any; if (cond) a = b; b = a; a({}); b({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; let a: any; let b: any; function f() { a = b; b = a; } a = r.expect.extend; f(); a(); b({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Shallow-copy MUST-overwrite precision (conditional, post-use, sibling, per-property, rest-assign, alias).
			{Code: `let mock: any; ({ ...mock } = expect); mock.extend = () => {}; mock.extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const { mock = { ...expect } } = {}; mock.extend = () => {}; mock.extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const { ...mock } = expect; const alias = mock; mock.extend = () => {}; alias.extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const { ...copy } = expect; if (false) { copy.extend = () => {}; } copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...copy } = expect; function setup() { copy.extend({ toBe() {} }); } setup(); copy.extend = () => {}; probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...copy } = expect; for (let i = 0; i < 2; i++) { copy.extend({ toBe() {} }); copy.extend = () => {}; } probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...copy } = expect; afterAll(() => { copy.extend = () => {}; }); copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...copy } = expect; (copy as any)[Symbol.toStringTag] = 'custom'; copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { ...api } = r; (api as any).other.extend = () => {}; api.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const a = { ...expect, extend: () => {} }; a.addEqualityTesters([tester]); probe();`, Errors: want("toBe=false toBeDefined=false equality=true")},
			{Code: `const b = { ...expect, addEqualityTesters: () => {} }; b.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { ...c } = expect; c.addEqualityTesters = () => {}; c.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const d = { ...expect, extend: expect.extend }; d.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Logical `&&`, multi-kind unions, repeated function invocations, same-stmt / optional-chain / destructuring-default writes.
			{Code: `const check = import.meta.rstest && import.meta.rstest.expect; check?.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const reg = expect && expect.extend; reg({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const noop = expect.extend && (() => {}); noop({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const fn = cond ? expect.addEqualityTesters : expect.extend; fn({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=true")},
			{Code: `import * as core from '@rstest/core'; let x: any; if (cond) x = core; else x = core.expect; x.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let install: any; function run() { install?.({ toBe() {} }); install = expect.extend; } run(); run(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let install: any; function setup() { install = expect.extend; } function run() { install?.({ toBe() {} }); setup(); } run(); run(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let fn: any; fn = (fn = () => {}, expect.extend); fn({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let fn: any = expect.extend; const obj: any = null; obj?.foo(fn = () => {}); fn({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let fn: any = expect.extend; let x: any; [x = (fn = () => {})] = [1]; fn({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Bound args, spread args, Reflect.apply, Object.assign({}, expect), destructuring defaults, reassigned ctx, chained optional call.
			{Code: `expect.extend.bind(expect, { toBe() {} })(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const reg = expect.extend.bind(expect, { toBe() {} }); reg(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `expect.extend.bind(expect, { toBe() {} })({}); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const args = [expect, { toBe() {} }] as const; expect.extend.call(...args); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `Reflect.apply(expect.extend, expect, [{ toBe() {} }]); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let extend: any; ({ extend = expect.extend } = {}); extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let fn: any; ({ extend: fn = () => {} } = expect); fn({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; let extend2: any; ({ expect: { extend: extend2 } = {} } = core); extend2({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `test('setup', ctx => { ctx.expect.extend({ toBe() {} }); ctx = other; }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const copy = Object.assign({}, expect); copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `expect.extend({ toBe() {} })?.foo(1); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `expect.addEqualityTesters([])?.foo(1); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			// Restored overwrites, reassigned shallow copies, attribute-only defineProperty, shorthand/multi-source Object.assign, destructuring RHS evaluation order.
			{Code: `const mock = { ...expect }; mock.extend = () => {}; mock.extend = expect.extend; mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let mock = { ...expect }; mock.extend = () => {}; mock = { ...expect }; mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let mock = { ...expect }; mock.extend = () => {}; mock = expect; const alias = mock; alias.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const mock = { ...expect }; Object.defineProperty(mock, "extend", { enumerable: false }); mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const { extend } = expect; const mock = { ...expect }; Object.assign(mock, { extend }); mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const mock = { ...expect }; Object.assign(mock, { extend() {} }, expect); mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let fn: any; ({ extend: fn } = ((fn = () => {}), expect)); fn({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let extend: any; ({ extend } = ((extend = () => {}), expect)); extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `var { extend: fn } = ((fn = () => {}), expect); fn({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let fn: any = () => {}; const { a = fn({ toBe() {} }) } = ((fn = expect.extend), {}); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Reflect.apply with bound expect.extend uses the bound first argument, not Reflect.apply's extra args.
			{Code: `import * as r from '@rstest/core'; const register = r.expect.extend.bind(r.expect, { toBe() {} }); Reflect.apply(register, r.expect, [{ toBeDefined() {} }]); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const register = expect.extend.bind(expect, { toBe() {} }).bind(expect, { toBeDefined() {} }); register(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; let register: any = () => {}; register = r.expect.extend.bind(r.expect, { toBe() {} }); Reflect.apply(register, r.expect, [{ toBeDefined() {} }]); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { register = r.expect.extend.bind(r.expect, { toBe() {} }) } = {}; Reflect.apply(register, r.expect, [{ toBeDefined() {} }]); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const register = r.expect && r.expect.extend.bind(r.expect, { toBe() {} }); Reflect.apply(register, r.expect, [{ toBeDefined() {} }]); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const register = cond ? expect.extend.bind(expect, { toBe() {} }) : expect.extend; register({ toBeDefined() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `const f1 = expect.extend.bind(expect, { toBe() {} }); const f2 = f1.bind(expect, { toBeDefined() {} }); const f3 = f2.bind(expect, { toBeDefined() {} }); const f4 = f3.bind(expect, { toBeDefined() {} }); Reflect.apply(f4, expect, [{ toBeDefined() {} }]); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const register = expect.extend.bind(expect, { toBe() {} }); Reflect.apply(register, expect); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			// Destructuring default skipped when source object literal explicitly defines the property as non-undefined.
			{Code: `import * as r from '@rstest/core'; function ignore(_m: any) {} const { register = r.expect.extend } = { register: ignore }; register({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { register = r.expect.extend } = { register: () => {} }; register({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { register = r.expect.extend } = { register: undefined }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; let register: any; ({ register = r.expect.extend } = { register: () => {} }); register({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			// Nested destructuring assignment default must NOT mistake fallback object literal for outer RHS.
			{Code: `import * as r from '@rstest/core'; let register: any; ({ a: { register = r.expect.extend } = { register: () => {} } } = { a: {} }); register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Reassigned function/class and ambient declarations are NOT guaranteed defined.
			{Code: `import * as r from '@rstest/core'; function ignore(_m: any) {} ignore = undefined as any; const { register = r.expect.extend } = { register: ignore }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; declare function ambientIgnore(_m: any): void; const { register = r.expect.extend } = { register: ambientIgnore }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; class Helper {} Helper = undefined as any; const { register = r.expect.extend } = { register: Helper }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Skipping destructuring default must preserve expect-config values under arbitrary object-literal property keys.
			{Code: `import * as r from '@rstest/core'; const { myExpect = r.expect } = { myExpect: { ...r.expect } }; myExpect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { myExpect = r.expect } = { myExpect: { ...r.expect, extend: () => {} } }; myExpect.extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; let myExpect: any; ({ myExpect = r.expect } = { myExpect: { ...r.expect } }); myExpect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { a: { myExpect = r.expect } } = { a: { myExpect: { ...r.expect } } }; myExpect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Non-computed `__proto__` sets prototype rather than own property; computed `["__proto__"]` defines own property.
			{Code: `import * as r from '@rstest/core'; const { __proto__: register = r.expect.extend } = { __proto__: null }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { "__proto__": register = r.expect.extend } = { "__proto__": null }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { ["__proto__"]: register = r.expect.extend } = { ["__proto__"]: () => {} }; register({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			// Spread after bound property, member-access bound calls, alias destructuring, and exact `{ extend: bound }` precision.
			{Code: `import * as r from '@rstest/core'; const other = { register: r.expect.extend }; const { register } = { register: r.expect.extend.bind(r.expect, { toBe() {} }), ...other }; register({ toBeDefined() {} }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `import * as r from '@rstest/core'; const wrapper = { extend: r.expect.extend.bind(r.expect, { toBe() {} }) }; wrapper.extend(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; ({ extend: r.expect.extend.bind(r.expect, { toBe() {} }) }).extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const wrapper = { extend: r.expect.extend.bind(r.expect, { toBe() {} }) }; const { extend } = wrapper; extend(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { extend } = { extend: r.expect.extend.bind(r.expect, { toBe() {} }) }; extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Self-referential unconditional writes and conditional self-bind precision.
			{Code: `import * as r from '@rstest/core'; let fn: any = r.expect.extend; fn = fn.bind(r.expect, { toBe() {} }); fn(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; let a: any; ({ a } = { a: r.expect.extend }); ({ a } = { a }); a({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; let a: any = r.expect.extend; const b = a; a = b; a({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; let ex: any = r.expect; ex = { ...ex }; ex.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; let fn: any = r.expect.extend.bind(r.expect, { toBe() {} }); if (cond) { fn = fn.bind(r.expect, { toBeDefined() {} }); } fn(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Nested destructuring fallbacks, spreads, dynamic keys, and definite-parent assignment default skip.
			{Code: `import * as r from '@rstest/core'; const { a: { register } = { register: r.expect.extend } } = {}; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; let register: any; ({ a: { register } = { register: r.expect.extend } } = {}); register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { a: { extend } = { extend: r.expect.extend.bind(r.expect, { toBe() {} }) } } = {}; extend(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const [{ register } = { register: r.expect.extend }] = []; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { a: { register } } = { a: { register: r.expect.extend }, ...other }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { [getKey()]: register } = { myExtend: r.expect.extend }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; let extend: any; ({ a: { extend } = r.expect } = { a: {} }); extend({ toBeDefined() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			// Ambient `declare global`, hoisted non-const enum, and conditional FunctionDeclaration are NOT definitely defined.
			{Code: `import * as r from '@rstest/core'; declare global { class AmbientGlobalClass {} } const { register = r.expect.extend } = { register: AmbientGlobalClass }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; function run() { const { register = r.expect.extend } = { register: E }; register({ toBe() {} }); } run(); enum E { A } probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `if (false) function ignore() {} const { register = expect.extend } = { register: ignore }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// `__proto__` prototype inheritance and getters in object literals / defineProperty / Object.assign.
			{Code: `import * as r from '@rstest/core'; const { register = () => {} } = { __proto__: { register: r.expect.extend } }; register({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const copy = { __proto__: r.expect }; copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const copy = { ...r.expect, get extend() { return r.expect.extend; } }; copy.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const mock = { ...r.expect }; Object.defineProperty(mock, 'extend', { get: () => r.expect.extend }); mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const mock = { ...r.expect }; Object.assign(mock, { get extend() { return r.expect.extend; } }); mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Shallow-copy restores via nested blocks, helper functions, aliases, logical/destructuring writes, and Object.assign spreads.
			{Code: `import * as r from '@rstest/core'; const mock = { ...r.expect }; mock.extend = () => {}; if (cond) { mock.extend = r.expect.extend; } mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const mock = { ...r.expect }; function restore() { mock.extend = r.expect.extend; } mock.extend = () => {}; restore(); mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { ...mock } = r.expect; const alias = mock; mock.extend = () => {}; alias.extend = r.expect.extend; alias.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const { ...mock } = r.expect; const alias = mock; mock.extend = () => {}; alias.extend = r.expect.extend; mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const mock = { ...r.expect }; delete (mock as any).extend; mock.extend ??= r.expect.extend; mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const mock = { ...r.expect }; ({ extend: mock.extend } = r.expect); mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const mock = { ...r.expect }; const sources = [r.expect]; Object.assign(mock, { extend: () => {} }, ...sources); mock.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Mutual alias cycle in overwrite summary must not poison cache.
			{Code: `import * as r from '@rstest/core'; let a: any = { ...r.expect }; let b: any = { ...r.expect }; if (true) { b = a; } a.extend = () => {}; b.extend = r.expect.extend; a.extend({ toBe() {} }); a = b; probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Escaped holder object (alias or non-first call arg) must not drop call-site matchers.
			{Code: `import * as r from '@rstest/core'; const holder = { extend: r.expect.extend.bind(r.expect, { toHaveLength() {} }) }; const alias = holder; alias.extend = r.expect.extend; holder.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const holder = { extend: r.expect.extend.bind(r.expect, { toHaveLength() {} }) }; replaceExtend(0, holder); holder.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			// Bound extend through getter, object spread, and const identifier `__proto__`.
			{Code: `import * as r from '@rstest/core'; const holder = { get extend() { return r.expect.extend.bind(r.expect, { toBe() {} }); } }; holder.extend(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const spreadObj = { ...{ extend: r.expect.extend.bind(r.expect, { toBe() {} }) } }; spreadObj.extend(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as r from '@rstest/core'; const proto = { extend: r.expect.extend.bind(r.expect, { toBe() {} }) }; const protoObj = { __proto__: proto }; protoObj.extend(); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
		},
	)
}

func TestRstestExpectCustomizationLinearScaling(t *testing.T) {
	var b strings.Builder
	b.WriteString("import * as core from '@rstest/core';\n")
	// 1. 1,000 repeated helper calls.
	b.WriteString("let helperInstall: any;\n")
	b.WriteString("function helper() { helperInstall = () => {}; }\n")
	for range 1000 {
		b.WriteString("helper();\n")
	}
	// 2. 1,000 sequential reassignments on a variable WITH a config write.
	b.WriteString("let reg: any = core.expect.extend;\n")
	for range 1000 {
		b.WriteString("if (cond) { reg = core.expect.extend; } reg = () => {}; reg({});\n")
	}
	// 3. 1,000 shallow-copy reads + 1,000 unrelated property writes + 1 relevant overwrite.
	b.WriteString("test('shallow', () => {\nconst { ...copy } = core.expect;\n")
	for range 500 {
		b.WriteString("copy.extend({});\n(copy as any).tag = 1;\n")
	}
	b.WriteString("copy.extend = () => {};\n});\n")
	// 4. 500 pre-setup reads + 500 post-read helper calls in same function.
	b.WriteString("let deferred: any;\nfunction setupDeferred() { deferred = core.expect.extend; }\nfunction suite() {\n")
	for range 500 {
		b.WriteString("deferred?.({});\n")
	}
	for range 500 {
		b.WriteString("setupDeferred();\n")
	}
	b.WriteString("}\n")
	// 5. 30-level alias diamond DAG (would explode 2^30 without memoization).
	b.WriteString("const d0 = core.expect;\n")
	for i := 1; i <= 30; i++ {
		fmt.Fprintf(&b, "const d%d = cond ? d%d : d%d;\n", i, i-1, i-1)
	}
	b.WriteString("d30.extend({ toBe() {} });\n")
	// 6. 30-level const and let bound-call diamond DAGs (would explode 2^30 without memoization).
	b.WriteString("const bc0 = core.expect.extend.bind(core.expect, { toBe() {} });\n")
	for i := 1; i <= 30; i++ {
		fmt.Fprintf(&b, "const bc%d = cond ? bc%d : bc%d;\n", i, i-1, i-1)
	}
	b.WriteString("bc30();\n")
	b.WriteString("let bl0: any = core.expect.extend.bind(core.expect, { toBe() {} });\n")
	for i := 1; i <= 30; i++ {
		fmt.Fprintf(&b, "let bl%d: any; if (cond) bl%d = bl%d; if (cond) bl%d = bl%d;\n", i, i, i-1, i, i-1)
	}
	b.WriteString("bl30();\n")
	// 7. 26-level shallow-copy alias chain with repeated assignments (would blow up 2^26 events without dedup).
	b.WriteString("const oa0 = { ...core.expect };\n")
	for i := 1; i <= 26; i++ {
		fmt.Fprintf(&b, "let oa%d: any = oa%d; oa%d = oa%d;\n", i, i-1, i, i-1)
	}
	b.WriteString("oa26.extend = core.expect.extend;\noa0.extend({ toBe() {} });\n")
	// 8. Initialized mutable fn with 300 conditional binds + 300 trailing calls (would be O(K*M) without post-write cache).
	b.WriteString("let multiFn: any = core.expect.extend;\n")
	for i := range 300 {
		fmt.Fprintf(&b, "if (cond) multiFn = core.expect.extend.bind(core.expect, { m%d() {} });\n", i)
	}
	for range 300 {
		b.WriteString("multiFn();\n")
	}
	b.WriteString("probe();\n")

	start := time.Now()
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectCustomizationProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code:   b.String(),
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "customization", Message: "toBe=true toBeDefined=false equality=false"}},
			},
		},
	)
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("customization analysis took %v; expected linear-time execution", elapsed)
	}
}
