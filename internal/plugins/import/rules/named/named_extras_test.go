package named_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/named"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNamedExtras(t *testing.T) {
	rule_tester.RunRuleTester(namedRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &named.NamedRule,
		[]rule_tester.ValidTestCase{
			// The shared export collector recognizes quoted re-export names.
			{Code: `import { quoted } from './quoted.js'`},
			{Code: `import { value } from "typed-package"`},
			{Code: `import { missing } from "typed-package"`, Settings: map[string]any{"import/ignore": []any{"node_modules"}}},
			{Code: `import './leaf.js'; import def, * as ns from './leaf.js'; export * from './leaf.js'; export * as all from './leaf.js'`},
			{Code: `const missing = 1; export { missing };`},
			{Code: `import type { missing } from './empty.js'; export type { missing } from './empty.js'; import { type missing, present } from './leaf.js'`},
			{Code: `import { present } from './stars.js'; import { present as cycled } from './cycle-b.js'`},
			{Code: `import { renamed } from './self-alias.js'`},
			{Code: `import { present } from './shared.js'; import { found } from './shared-alias.js'`},
			{Code: `import { missing } from './unknown.js'; import { remote } from './unknown-named.js'`},
			{Code: `import { ns } from './namespace.js'; import { missing } from './common.cjs'; import { '' as empty } from './leaf.js'`},
			{Code: `import { missing } from './leaf.js'`, Settings: map[string]any{"import/ignore": []any{"leaf\\.js$"}}},
			{Code: `import { missing } from './leaf.js'`, Settings: map[string]any{"import/extensions": []any{".ts"}}},
			{Code: `import { outer } from './second.js'`, Settings: map[string]any{"import/ignore": []any{"leaf\\.js$"}}},
			{Code: `const { missing } = require('./leaf.js')`, Options: []any{map[string]any{}}},
			{Code: `const { present, ...rest } = require('./leaf.js')`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `const { required } = (require)(('conditional-package'))`, FileName: "consumer.mts", TSConfig: "tsconfig.nodenext.json", Options: []any{map[string]any{"commonjs": true}}},
			{Code: `const { 'missing': value, ['missing']: computed, [getName()]: dynamic, [ns.missing]: member } = require('./leaf.js')`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `const {} = require('./leaf.js'); const [missing] = require('./leaf.js'); const value = require('./leaf.js')`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `const { missing } = load('./leaf.js'); const { other } = obj.require('./leaf.js'); const { value } = require?.('./leaf.js')`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: "const { missing } = require(); const { other } = require('./leaf.js', 1); const { value } = require(source); const { x } = require(`./leaf.js`)", Options: []any{map[string]any{"commonjs": true}}},
			{Code: `const { missing } = (obj?.load)('./leaf.js'); const { other } = (require as any)('./leaf.js'); const { x } = require('./leaf.js') as any`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `const { [missing as string]: alias } = require('./leaf.js'); const { missing } = require('./leaf.js')!`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `const { missing } = require('./common.cjs'); const { other } = require('./unresolved.js')`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `function consume({ missing }) {} try {} catch ({ missing }) {}`, Options: []any{map[string]any{"commonjs": true}}},
		}, []rule_tester.InvalidTestCase{
			{Code: `import { default as value } from './types'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './types'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `import { default as value } from './types'`, TSConfig: "tsconfig.interop.json", Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './types'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `import { default as value } from './types'`, TSConfig: "tsconfig.nodenext.json", Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './types'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			// A namespace export exposes only its alias, not the members as peers.
			{Code: `import { present } from './namespace.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "present not found in './namespace.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			// Program resolves declaration entries; a CommonJS runtime entry does not hide them.
			{Code: `import { missing } from "typed-package"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in 'typed-package'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `import { missing as local } from './leaf.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `export { missing as publicName } from './leaf.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `export { type missing } from './leaf.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 15, EndLine: 1, EndColumn: 22}}},
			{Code: `import { missing } from './empty.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './empty.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `/* 😀 */ import { '\u006dissing' as alias } from './leaf.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 19, EndLine: 1, EndColumn: 33}}},
			{Code: `import {
  present,
  "不存在" as alias
} from "./leaf.js"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "不存在 not found in './leaf.js'", Line: 3, Column: 3, EndLine: 3, EndColumn: 8}}},
			{Code: `import { '' as alias } from './empty.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: " not found in './empty.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 12}}},
			{Code: `import { missing } from './cycle-a.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './cycle-a.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `import { missing } from './self.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './self.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `import { missing } from './shared.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './shared.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `import { renamed } from './shared-broken.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "renamed not found via shared-broken.js -> shared.js", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `import { renamed } from './star.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "renamed not found in './star.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `import { default as value } from './stars.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './stars.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `import { missing } from './types'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './types'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `import { outer as local } from './second.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "outer not found via second.js -> first.js -> leaf.js", Line: 1, Column: 10, EndLine: 1, EndColumn: 15}}},
			{Code: `export { outer as publicName } from './second.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "outer not found via second.js -> first.js -> leaf.js", Line: 1, Column: 10, EndLine: 1, EndColumn: 15}}},
			{Code: `import { missing } from './cycle-named-a.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found via cycle-named-a.js -> cycle-named-b.js -> cycle-named-a.js", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `const { missing: alias = 1 } = require('./leaf.js')`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
			{Code: `const { missing = 1 } = require('./leaf.js')`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
			{Code: `const { missing: { nested } } = require('./leaf.js')`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
			{Code: `const { [missing]: alias } = require('./leaf.js')`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `const { [(missing)]: alias } = (require)(('./leaf.js'))`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 11, EndLine: 1, EndColumn: 18}}},
			{Code: `const { missing } = /** @type {any} */ (require('./leaf.js'))`, FileName: "consumer.js", Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
			{Code: `const { missing } = (/** @type {any} */ (require))('./leaf.js')`, FileName: "consumer.js", Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
			{Code: `const { missing } = require(/** @type {string} */ ('./leaf.js'))`, FileName: "consumer.js", Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
			{Code: `import { missing } from './leaf.js' with { type: 'json' }`, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `const { missing } = ((require)(('./leaf.js')))`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
			{Code: `const { imported } = (require)(('conditional-package'))`, FileName: "consumer.mts", TSConfig: "tsconfig.nodenext.json", Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "imported not found in 'conditional-package'", Line: 1, Column: 9, EndLine: 1, EndColumn: 17}}},
			{Code: `function f(require) { const { missing } = require('./leaf.js') }`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 31, EndLine: 1, EndColumn: 38}}},
			{Code: `const { outer: alias } = require('./second.js')`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "outer not found via second.js -> first.js -> leaf.js", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
			{Code: `import { missing } from './leaf.js'`, Settings: map[string]any{"import/extensions": []any{".ts"}, "import/parsers": map[string]any{"custom-parser": []any{".js"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './leaf.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
		})
}

func TestNamedESModuleDefault(t *testing.T) {
	root := namedRoot(t, "testdata/extras.txtar")
	for _, config := range []string{"tsconfig.json", "tsconfig.interop.json", "tsconfig.nodenext.json"} {
		t.Run(config, func(t *testing.T) {
			rule_tester.RunRuleTester(root, config, t, &named.NamedRule,
				[]rule_tester.ValidTestCase{
					{Code: `import { default as value } from './leaf.js'`},
					{Code: `import { default as value } from './real-defaults.mjs'`},
					{Code: `import { default as value } from './namespace-default.mjs'`},
					{Code: `export { default as value } from './real-defaults.mjs'`},
					{Code: `import { default as value } from './common.cjs'`},
				}, []rule_tester.InvalidTestCase{
					{Code: `import { default as value } from './named-only.mjs'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './named-only.mjs'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
					{Code: `import { default as value } from './typed-only.mts'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './typed-only.mts'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
					{Code: `import { default as value } from './namespace.js'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './namespace.js'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
					{Code: `import { default as value } from './umd-global.mjs'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './umd-global.mjs'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
					{Code: `export { default as value } from './named-only.mjs'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './named-only.mjs'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
					{Code: `import { default as value } from './missing-default.mjs'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found via missing-default.mjs -> named-only.mjs", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
				})
		})
	}
	for _, config := range []string{"tsconfig.interop.json", "tsconfig.nodenext.json"} {
		t.Run("export-equals-"+config, func(t *testing.T) {
			rule_tester.RunRuleTester(root, config, t, &named.NamedRule,
				[]rule_tester.ValidTestCase{{Code: `import { default as value } from './export-equals.cjs'`}}, nil)
		})
	}
}

func TestNamedNodeDefault(t *testing.T) {
	rule_tester.RunRuleTester(namedRoot(t, "testdata/extras.txtar"), "tsconfig.nodenext.json", t, &named.NamedRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { default as value } from './compiled.cjs'`, FileName: "consumer.mts"},
			{Code: `export { default as value } from './compiled.cjs'`, FileName: "consumer.mts"},
			{Code: `import { default as value } from './compiled-barrel.mjs'`, FileName: "consumer.cts"},
		}, []rule_tester.InvalidTestCase{
			{Code: `import { default as value } from './compiled.cjs'`, FileName: "consumer.cts", Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './compiled.cjs'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `const { default: value } = (require)(('./compiled.cjs'))`, FileName: "consumer.mts", Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './compiled.cjs'", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
			{Code: `import { missing } from './compiled.cjs'`, FileName: "consumer.mts", Errors: []rule_tester.InvalidTestCaseError{{Message: "missing not found in './compiled.cjs'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
		})
}
