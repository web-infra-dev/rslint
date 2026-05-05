package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestOrderRuleConsolidateIslands isolates the `consolidateIslands` knob.
//
// `consolidateIslands="inside-groups"` is meaningful only when paired with
// `newlines-between="always-and-inside-groups"` (or the type-only twin).
// Within that mode, multi-line imports must be separated from neighboring
// imports by a blank line, while consecutive single-line imports stay
// grouped together.
//
// Three observable rules:
//
//  1. Single-line + multi-line neighbors → blank line required between them.
//  2. Multi-line + single-line neighbors → blank line required between them.
//  3. Two single-line same-group imports separated by a blank line → that
//     blank line should be removed.
func TestOrderRuleConsolidateIslands(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// ---- Single-line islands stay together ----
			{
				Code: `
import a from 'a';
import b from 'b';

import {
	x,
	y,
} from 'c';

import d from 'd';`,
				Options: map[string]interface{}{
					"newlines-between":   "always-and-inside-groups",
					"consolidateIslands": "inside-groups",
					"alphabetize":        map[string]interface{}{"order": "asc"},
				},
			},
			// ---- consolidateIslands=never (default) → no extra enforcement ----
			{
				Code: `
import a from 'a';
import {
	x,
	y,
} from 'c';
import d from 'd';`,
				Options: map[string]interface{}{
					"newlines-between": "always-and-inside-groups",
				},
			},
			// ---- consolidateIslands without always-and-inside-groups: no-op ----
			{
				Code: `
import a from 'a';
import {
	x,
	y,
} from 'c';
import d from 'd';`,
				Options: map[string]interface{}{
					"newlines-between":   "always",
					"consolidateIslands": "inside-groups",
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- single-line directly before a multi-line: insert blank ----
			{
				Code: `
import a from 'a';
import {
	x,
	y,
} from 'b';`,
				Output: []string{`
import a from 'a';

import {
	x,
	y,
} from 'b';`},
				Options: map[string]interface{}{
					"newlines-between":   "always-and-inside-groups",
					"consolidateIslands": "inside-groups",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consolidate"}},
			},
			// ---- multi-line directly before a single-line: insert blank ----
			{
				Code: `
import {
	x,
	y,
} from 'a';
import b from 'b';`,
				Output: []string{`
import {
	x,
	y,
} from 'a';

import b from 'b';`},
				Options: map[string]interface{}{
					"newlines-between":   "always-and-inside-groups",
					"consolidateIslands": "inside-groups",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consolidate"}},
			},
			// ---- two single-line same-group imports with a stray blank between ----
			{
				Code: `
import a from 'a';

import b from 'b';`,
				Output: []string{`
import a from 'a';
import b from 'b';`},
				Options: map[string]interface{}{
					"newlines-between":   "always-and-inside-groups",
					"consolidateIslands": "inside-groups",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consolidate"}},
			},
		},
	)
}
