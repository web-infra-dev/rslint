package no_named_as_default_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_named_as_default"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNamedAsDefaultExtras(t *testing.T) {
	rule_tester.RunRuleTester(namedDefaultRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &no_named_as_default.NoNamedAsDefaultRule,
		[]rule_tester.ValidTestCase{
			{Code: `import './base';`},
			{Code: `import { foo } from './base';`},
			{Code: `import * as foo from './base';`},
			{Code: `import { default as foo } from './base';`},
			{Code: `export { default as foo } from './base';`},
			{Code: `export * as foo from './base';`},
			{Code: `const foo = require('./base'); import('./base');`},
			{Code: `import foo = require('./base');`},
			{Code: `import unrelated from './base';`},
			{Code: `import foo from './missing';`},
			{Code: `import foo from 'node:fs';`},
			{Code: `import foo from './named-only';`},
			{Code: `import foo from './star-only';`},
			{Code: `import foo from './common.cjs';`},
			{Code: `import foo from './unresolved-star';`},
			{Code: `import foo from './base';`, Settings: map[string]interface{}{"import/ignore": []interface{}{`base\.ts$`}}},
			// v2.32.0 exempts direct re-exports from the same file, even if
			// they refer to different bindings in that file.
			{Code: `import foo from './same-source';`},
			{Code: `import foo from './same-binding';`},
			{Code: `import foo from './missing-default';`},
			{Code: `import foo from './unresolved-default';`},
			{Code: `import foo from './chained-missing-default';`},
			{Code: `import foo from './cycle-a';`},
			{Code: `import foo from './export-equals';`},
			{Code: `import foo from './interop-default';`},
			{Code: `import foo from './namespace-only';`},
			{Code: `import foo from './namespace-types';`},
			{Code: `import foo from './default-function';`},
			{Code: `import foo from './ambient-module';`},
			// A namespace re-export exposes its alias, not its members.
			// Upstream v2.32.0 also considers those members; see the docs.
			{Code: `import foo from './namespace-members';`},
			// tsgo accepts this return; this rule does not relay parser errors.
			{Code: `import foo from './invalid-return.js';`},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `import foo from './base';`, FileName: "consumer.js", Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './base';`, FileName: "consumer.mjs", Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './base';`, FileName: "consumer.mts", Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo, * as ns from './base';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './base'; const view = <foo />;`, Tsx: true, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import type Shape from './base';`, Errors: namedDefaultError("Shape", 1, 13, 18)},
			{Code: `import foo from './base' with { type: 'json' };`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: "import\n  /* local */ foo\n  from './base';", Errors: namedDefaultError("foo", 2, 15, 18)},
			{Code: `/* 😀 */ import café from './base';`, Errors: namedDefaultError("café", 1, 17, 21)},
			{Code: `import 𐐀 from './base';`, Errors: namedDefaultError("𐐀", 1, 8, 10)},
			{Code: `import \u0066oo from './base';`, Errors: namedDefaultError("foo", 1, 8, 16)},
			{Code: `import foo from './base';`, Options: []any{}, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './star';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './different-sources';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './star-name';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './local-default';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './local-named';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './explicit-missing-name';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './unresolved-name';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './through-local';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './cycle-star-a';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './default-namespace';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			// TypeScript interop makes these defaults available upstream.
			{Code: `import foo from './export-equals';`, TSConfig: "tsconfig.interop.json", Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './named-only';`, TSConfig: "tsconfig.interop.json", Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './interop-default';`, TSConfig: "tsconfig.interop.json", Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './namespace-only';`, TSConfig: "tsconfig.interop.json", Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './namespace-types';`, TSConfig: "tsconfig.interop.json", Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './quoted-namespace';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './local-string';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './named-types';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './star-types';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './export-equals-function';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './escaped-export';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './destructuring';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			// Documented differences: quoted re-export names are recognized,
			// and implicit esModuleInterop follows the TypeScript module option.
			{Code: `import foo from './quoted-named';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './quoted-default';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './quoted-local';`, Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './named-only';`, TSConfig: "tsconfig.nodenext.json", Errors: namedDefaultError("foo", 1, 8, 11)},
			{Code: `import foo from './namespace-only';`, TSConfig: "tsconfig.nodenext.json", Errors: namedDefaultError("foo", 1, 8, 11)},
		},
	)
}
