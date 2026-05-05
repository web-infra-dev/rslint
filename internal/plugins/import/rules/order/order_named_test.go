package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestOrderRuleNamed exercises the `named` option, including the `import`,
// `export`, `require`, and `cjsExports` toggles plus the `types` modifier
// (`mixed` / `types-first` / `types-last`).
func TestOrderRuleNamed(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// ---- named imports already in alphabetical order ----
			{
				Code: `
import { a, b, c } from 'foo';`,
				Options: map[string]interface{}{
					"named":       true,
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
			},
			// ---- named imports without alphabetize → no enforcement ----
			{
				Code: `
import { c, a, b } from 'foo';`,
				Options: map[string]interface{}{
					"named": true,
				},
			},
			// ---- named.import false: don't check imports ----
			{
				Code: `
import { c, a, b } from 'foo';`,
				Options: map[string]interface{}{
					"named":       map[string]interface{}{"enabled": false, "import": false},
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
			},
			// ---- single named import — no comparison possible ----
			{
				Code: `
import { onlyOne } from 'foo';`,
				Options: map[string]interface{}{
					"named":       true,
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
			},
			// ---- destructured require named in alphabetical order ----
			{
				Code: `
var { a, b } = require('foo');`,
				Options: map[string]interface{}{
					"named":       map[string]interface{}{"enabled": true, "require": true},
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
			},
			// ---- export specifiers in alphabetical order ----
			{
				Code: `
export { a, b } from 'foo';`,
				Options: map[string]interface{}{
					"named":       map[string]interface{}{"enabled": true, "export": true},
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
			},
			// ---- types-last: type imports come after value imports in same list ----
			{
				Code: `
import { a, b, type T } from 'foo';`,
				Options: map[string]interface{}{
					"named":       map[string]interface{}{"enabled": true, "types": "types-last"},
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
			},
			// ---- types-first: type imports come first ----
			{
				Code: `
import { type T, a, b } from 'foo';`,
				Options: map[string]interface{}{
					"named":       map[string]interface{}{"enabled": true, "types": "types-first"},
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
			},
			// ---- mixed: types intermingled alphabetically ----
			{
				Code: `
import { a, type b, c } from 'foo';`,
				Options: map[string]interface{}{
					"named":       map[string]interface{}{"enabled": true, "types": "mixed"},
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
			},
			// ---- cjsExports shadowed by user binding → no false positive ----
			// `let module = ...` shadows the ambient `module`, so the rule
			// must not treat the assignment as a CJS export.
			{
				Code: `
let module: any;
const a = 1, b = 2, c = 3;
module.exports = { c, a, b };`,
				FileName: "cjs-shadow.ts",
				Options: map[string]interface{}{
					"named":       map[string]interface{}{"enabled": true, "cjsExports": true},
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- named imports out of order ----
			// reverse-direction yields fewer reports (c should move to the end,
			// not a/b move to the start) — picks the minimal report set.
			{
				Code: `
import { c, a, b } from 'foo';`,
				Options: map[string]interface{}{
					"named":       true,
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "namedOrder", Message: "`c` import should occur after import of `b`"},
				},
			},
			// ---- destructured require with out-of-order names ----
			{
				Code: `
var { c, a, b } = require('foo');`,
				Options: map[string]interface{}{
					"named":       map[string]interface{}{"enabled": true, "require": true},
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "namedOrder", Message: "`c` import should occur after import of `b`"},
				},
			},
			// ---- export specifiers out of order ----
			{
				Code: `
export { b, a } from 'foo';`,
				Options: map[string]interface{}{
					"named":       map[string]interface{}{"enabled": true, "export": true},
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "namedOrder"},
				},
			},
			// ---- types-last violated: type appears before value ----
			{
				Code: `
import { type T, a } from 'foo';`,
				Options: map[string]interface{}{
					"named":       map[string]interface{}{"enabled": true, "types": "types-last"},
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "namedOrder"},
				},
			},

			// ---- cjsExports: shorthand `module.exports = { c, a, b }` ----
			// Upstream's check requires both keys AND values to be identifiers,
			// so use the shorthand form here.
			{
				Code: `
const a = 1, b = 2, c = 3;
module.exports = { c, a, b };`,
				FileName: "cjs.ts",
				Options: map[string]interface{}{
					"named":       map[string]interface{}{"enabled": true, "cjsExports": true},
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "namedOrder"},
				},
			},
		},
	)
}
