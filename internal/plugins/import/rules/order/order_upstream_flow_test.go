// TestOrderUpstreamFlow migrates the upstream Flow-parser cases from
// import-js/eslint-plugin-import at 1a39fb8672c83f8781c3a97f02eee00a7a2f0253.
// The generic case remains active; each parser-incompatible case has an
// explicit SKIP marker. tsgo lock-ins live in order_extras_*_test.go.
package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderUpstreamFlow(t *testing.T) {
	// TypeScript's parser intentionally rejects Flow's `import typeof` syntax.
	// The four affected upstream cases stay visible here instead of silently
	// disappearing from the migration inventory.
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// SKIP: TypeScript's parser rejects Flow `import typeof` syntax.
			{
				Skip: true,
				Code: `
        import type {Bar} from 'common';
        import typeof {foo} from 'common';
        import {bar} from 'common';
      `,
				Options: []any{map[string]any{
					"alphabetize": map[string]any{"order": "asc", "orderImportKind": "asc"},
				}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// SKIP: TypeScript's parser rejects Flow `import typeof` syntax.
			{
				Skip: true,
				Code: `
        import type {Bar} from 'common';
        import {bar} from 'common';
        import typeof {foo} from 'common';
      `,
				Options: []any{map[string]any{
					"alphabetize": map[string]any{"order": "asc", "orderImportKind": "asc"},
				}},
				Output: []string{`
        import type {Bar} from 'common';
        import typeof {foo} from 'common';
        import {bar} from 'common';
      `},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "order", Message: "`common` typeof import should occur before import of `common`",
					Line: 4, Column: 9, EndLine: 4, EndColumn: 43,
				}},
			},
			// SKIP: TypeScript's parser rejects Flow `import typeof` syntax.
			{
				Skip: true,
				Code: `
        import type {Bar} from 'common';
        import {bar} from 'common';
        import typeof {foo} from 'common';
      `,
				Options: []any{map[string]any{
					"alphabetize": map[string]any{"order": "asc", "orderImportKind": "desc"},
				}},
				Output: []string{`
        import {bar} from 'common';
        import typeof {foo} from 'common';
        import type {Bar} from 'common';
      `},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "order", Message: "`common` type import should occur after typeof import of `common`",
					Line: 2, Column: 9, EndLine: 2, EndColumn: 41,
				}},
			},
			// SKIP: TypeScript's parser rejects Flow `import typeof` syntax.
			{
				Skip: true,
				Code: `
        import type {Bar} from './local/sub';
        import {bar} from './local/sub';
        import {baz} from './local-sub';
        import typeof {foo} from './local/sub';
      `,
				Options: []any{map[string]any{
					"alphabetize": map[string]any{"order": "asc", "orderImportKind": "asc"},
				}},
				Output: []string{`
        import type {Bar} from './local/sub';
        import typeof {foo} from './local/sub';
        import {bar} from './local/sub';
        import {baz} from './local-sub';
      `},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "order", Message: "`./local/sub` typeof import should occur before import of `./local/sub`",
					Line: 5, Column: 9, EndLine: 5, EndColumn: 48,
				}},
			},
			{
				Code: `
        import { cfg } from 'path/path/path/src/Cfg';
        import { l10n } from 'path/src/l10n';
        import { helpers } from 'path/path/path/helpers';
        import { tip } from 'path/path/tip';

        import { controller } from '../../../../path/path/path/controller';
        import { component } from '../../../../path/path/path/component';
      `,
				Options: []any{map[string]any{
					"groups": []any{[]any{"builtin", "external"}, "internal", []any{"sibling", "parent"}, "object", "type"},
					"pathGroups": []any{
						map[string]any{"pattern": "react", "group": "builtin", "position": "before", "patternOptions": map[string]any{"matchBase": true}},
						map[string]any{"pattern": "*.+(css|svg)", "group": "type", "position": "after", "patternOptions": map[string]any{"matchBase": true}},
					},
					"pathGroupsExcludedImportTypes": []any{},
					"alphabetize":                   map[string]any{"order": "asc"},
					"newlines-between":              "always",
				}},
				Output: []string{
					`
        import { helpers } from 'path/path/path/helpers';
        import { cfg } from 'path/path/path/src/Cfg';
        import { l10n } from 'path/src/l10n';
        import { tip } from 'path/path/tip';

        import { component } from '../../../../path/path/path/component';
        import { controller } from '../../../../path/path/path/controller';
      `,
					`
        import { helpers } from 'path/path/path/helpers';
        import { cfg } from 'path/path/path/src/Cfg';
        import { tip } from 'path/path/tip';
        import { l10n } from 'path/src/l10n';

        import { component } from '../../../../path/path/path/component';
        import { controller } from '../../../../path/path/path/controller';
      `,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "order", Message: "`path/path/path/helpers` import should occur before import of `path/path/path/src/Cfg`", Line: 4, Column: 9, EndLine: 4, EndColumn: 58},
					{MessageId: "order", Message: "`path/path/tip` import should occur before import of `path/src/l10n`", Line: 5, Column: 9, EndLine: 5, EndColumn: 45},
					{MessageId: "order", Message: "`../../../../path/path/path/component` import should occur before import of `../../../../path/path/path/controller`", Line: 8, Column: 9, EndLine: 8, EndColumn: 74},
				},
			},
		},
	)
}
