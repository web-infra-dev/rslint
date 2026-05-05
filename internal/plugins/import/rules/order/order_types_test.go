package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestOrderRuleTypesMatrix exhaustively covers `sortTypesGroup` interactions
// with the major `groups` shapes and the `newlines-between` /
// `newlines-between-types` knobs.
//
// `sortTypesGroup` causes type-only imports to be re-ranked into a parallel
// hierarchy that mirrors the value-import group order, with sub-ranks of
// `type.rank + valueGroupRank/10`. The rule is sensitive to:
//   - whether `"type"` is in `groups` at all
//   - whether `sortTypesGroup` is on
//   - whether `newlines-between-types` is set (defaults to `newlines-between`)
//
// We cover the truth-table corners that materially change behaviour and skip
// the redundant cells.
func TestOrderRuleTypesMatrix(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// type NOT in groups → type imports interleave by their value rank
			// ============================================================

			// Default groups: "type" omitted → all omitted types share the
			// trailing rank, so a type import classifies by its value group.
			{
				Code: `
import fs from 'fs';
import type {T} from 'fs';
import async from 'async';
import type {U} from 'async';`,
			},

			// ============================================================
			// "type" in groups, sortTypesGroup=false (default)
			// → type imports cluster together at the "type" group's rank
			// ============================================================

			{
				Code: `
import fs from 'fs';
import async from 'async';
import type {T} from 'fs';
import type {U} from 'async';`,
				Options: map[string]interface{}{
					"groups": []interface{}{"builtin", "external", "type"},
				},
			},

			// "type" first
			{
				Code: `
import type {T} from 'fs';
import type {U} from 'async';
import fs from 'fs';
import async from 'async';`,
				Options: map[string]interface{}{
					"groups": []interface{}{"type", "builtin", "external"},
				},
			},

			// ============================================================
			// sortTypesGroup=true → type imports form a parallel hierarchy
			// ============================================================

			// Type imports re-grouped by value-rank: builtin types first,
			// then external types, followed by value imports in same order.
			{
				Code: `
import type {T} from 'fs';
import type {U} from 'async';
import type {V} from '../foo';
import fs from 'fs';
import async from 'async';
import parent from '../foo';`,
				Options: map[string]interface{}{
					"groups":         []interface{}{"type", "builtin", "external", "parent"},
					"sortTypesGroup": true,
				},
			},

			// sortTypesGroup with alphabetize: types alphabetized within
			// their sub-group, values alphabetized within theirs.
			{
				Code: `
import type {A} from 'a';
import type {B} from 'b';
import a from 'a';
import b from 'b';`,
				Options: map[string]interface{}{
					"groups":         []interface{}{"type", "external"},
					"sortTypesGroup": true,
					"alphabetize":    map[string]interface{}{"order": "asc"},
				},
			},

			// ============================================================
			// newlines-between-types overrides newlines-between for the
			// type sub-group transitions
			// ============================================================

			// newlines-between=always; newlines-between-types=ignore →
			// type sub-group transitions don't require a newline.
			{
				Code: `
import type {T} from 'fs';
import type {U} from 'async';

import fs from 'fs';

import async from 'async';`,
				Options: map[string]interface{}{
					"groups":                 []interface{}{"type", "builtin", "external"},
					"sortTypesGroup":         true,
					"newlines-between":       "always",
					"newlines-between-types": "ignore",
				},
			},

			// newlines-between=never + sortTypesGroup → still no blank lines
			// even between type → value transition.
			{
				Code: `
import type {T} from 'fs';
import type {U} from 'async';
import fs from 'fs';
import async from 'async';`,
				Options: map[string]interface{}{
					"groups":           []interface{}{"type", "builtin", "external"},
					"sortTypesGroup":   true,
					"newlines-between": "never",
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// "type" in groups, sortTypesGroup=false → type cluster violated
			// ============================================================

			// type appears between two value imports of the same value group.
			{
				Code: `
import fs from 'fs';
import type {T} from 'fs';
import async from 'async';`,
				Output: []string{`
import fs from 'fs';
import async from 'async';
import type {T} from 'fs';
`},
				Options: map[string]interface{}{
					"groups": []interface{}{"builtin", "external", "type"},
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},

			// ============================================================
			// sortTypesGroup=true → type sub-rank ordering violated
			// ============================================================

			// type-external before type-builtin when groups put builtin first.
			{
				Code: `
import type {U} from 'async';
import type {T} from 'fs';
import fs from 'fs';
import async from 'async';`,
				Output: []string{`
import type {T} from 'fs';
import type {U} from 'async';
import fs from 'fs';
import async from 'async';`},
				Options: map[string]interface{}{
					"groups":         []interface{}{"type", "builtin", "external"},
					"sortTypesGroup": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},

			// ============================================================
			// newlines-between-types: always — missing newline between
			// type sub-groups
			// ============================================================

			// builtin-types should be separated from external-types by a blank
			// line (since they're now distinct sub-groups).
			{
				Code: `
import type {T} from 'fs';
import type {U} from 'async';

import fs from 'fs';

import async from 'async';`,
				Output: []string{`
import type {T} from 'fs';

import type {U} from 'async';

import fs from 'fs';

import async from 'async';`},
				Options: map[string]interface{}{
					"groups":                 []interface{}{"type", "builtin", "external"},
					"sortTypesGroup":         true,
					"newlines-between":       "always",
					"newlines-between-types": "always",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline"}},
			},
		},
	)
}
