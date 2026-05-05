package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestOrderRuleEdges covers SKILL.md Dimension 4 universal edge shapes plus
// tsgo-specific quirks that upstream's test file cannot exercise (paren AST,
// numeric-literal normalization, optional chain flag, multi-byte / CRLF).
//
// These tests lock in current behaviour even where upstream wouldn't produce
// the same diagnostics, so future refactors can't silently flip semantics.
func TestOrderRuleEdges(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		buildEdgeValidCases(),
		buildEdgeInvalidCases(),
	)
}

func buildEdgeValidCases() []rule_tester.ValidTestCase {
	return []rule_tester.ValidTestCase{
		// ============================================================
		// Dimension 4 — receiver / expression wrappers on require()
		// ============================================================

		// (require('foo')) — parenthesised require still recognised.
		{
			Code: `
var fs = (require('fs'));
var sibling = require('./foo');`,
		},
		// require('foo')!.bar — TS non-null assertion on require result.
		{
			Code: `
var fs = require('fs')!;
var sibling = require('./foo');`,
		},
		// (require('foo') as any) — TS as-expression wrapper.
		{
			Code: `
var fs = (require('fs') as any);
var sibling = require('./foo');`,
		},
		// require('foo') satisfies T — TS satisfies expression.
		{
			Code: `
var fs = require('fs') satisfies any;
var sibling = require('./foo');`,
		},
		// (require('foo') as any).bar — wrapped chain, still recognised.
		{
			Code: `
var fs = (require('fs') as any).x;
var sibling = require('./foo');`,
		},
		// require?.('foo') — optional-chain CallExpression. tsgo's IsRequireCall
		// (and upstream's isStaticRequire) match it as a normal require —
		// rank ordering still applies. We lock in this parity here.
		{
			Code: `
var fs = require('fs');
var x = require?.('async');
var sibling = require('./foo');`,
		},

		// ============================================================
		// Dimension 4 — destructured require key forms
		// ============================================================

		// Computed-key destructure → entire list bailed out (no named report).
		{
			Code: `
var { [key]: x } = require('foo');`,
			Options: map[string]interface{}{"named": true},
		},
		// String-literal key in destructure → bail out.
		{
			Code: `
var { 'foo-bar': fooBar } = require('foo');`,
			Options: map[string]interface{}{"named": true},
		},

		// ============================================================
		// Dimension 4 — declaration / container forms
		// ============================================================

		// Single-segment declare module name vs nested form.
		{
			Code: `
declare module 'A' {
  import fs from 'fs';
  import async from 'async';
}
declare module 'B' {
  import path from 'path';
  import lodash from 'lodash';
}`,
		},

		// ============================================================
		// Dimension 4 — nesting / traversal boundaries
		// ============================================================

		// Out-of-order require inside a function body must not bleed into the
		// top-level check.
		{
			Code: `
import fs from 'fs';
import async from 'async';
function f() {
  const sibling = require('./foo');
  const path = require('path');
}`,
		},
		// Top-level imports OK; an inner declare-module's check is independent
		// (and itself OK).
		{
			Code: `
import async from 'async';
import sibling from './foo';

declare module 'x' {
  import path from 'path';
  import { foo } from 'foo';
}`,
		},

		// ============================================================
		// Dimension 4 — graceful degradation
		// ============================================================

		// Empty file.
		{Code: ``},
		// Only comments.
		{Code: `// nothing here\n/* nothing here */`},
		// Single import — no comparisons possible.
		{Code: `import fs from 'fs';`},

		// ============================================================
		// tsgo-specific — node:* prefix and joined builtin names
		// ============================================================

		// `node:fs` and `fs/promises` both classify as builtin.
		{
			Code: `
import fs from 'node:fs';
import promises from 'fs/promises';
import async from 'async';`,
		},
		// `node:test` (Node test runner) is in our builtin table.
		{
			Code: `
import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import async from 'async';`,
		},

		// ============================================================
		// CRLF endings — fix output must preserve line discipline
		// ============================================================

		{
			Code: "import fs from 'fs';\r\nimport async from 'async';\r\n",
		},

		// ============================================================
		// Multi-byte identifiers in module specifiers (BMP + surrogate)
		// ============================================================

		{
			Code: `
import α from 'α';
import β from 'β';`,
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
		},

		// ============================================================
		// Settings: import/internal-regex
		// ============================================================

		// "@my/" prefix marks internal — placed before externals via groups.
		{
			Code: `
import a from '@my/a';
import async from 'async';`,
			Options: map[string]interface{}{
				"groups": []interface{}{"internal", "external"},
			},
			Settings: map[string]interface{}{
				"import/internal-regex": "^@my/",
			},
		},

		// ============================================================
		// Settings: import/core-modules
		// ============================================================

		// "my-builtin" is treated as builtin via settings.
		{
			Code: `
import x from 'my-builtin';
import async from 'async';
import sibling from './foo';`,
			Settings: map[string]interface{}{
				"import/core-modules": []interface{}{"my-builtin"},
			},
		},
	}
}

func buildEdgeInvalidCases() []rule_tester.InvalidTestCase {
	return []rule_tester.InvalidTestCase{
		// ============================================================
		// upstream semantic walk lock-ins
		// ============================================================

		// Lock in: `require('foo').bar.baz` is still recognised as a static require.
		// (upstream `getRequireBlock` walks the MemberExpression chain.)
		{
			Code: `
var x = require('./a').b.c;
var fs = require('fs');`,
			Output: []string{`
var fs = require('fs');
var x = require('./a').b.c;
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},
		// Lock in: `require('foo')()` (immediately invoked require).
		{
			Code: `
var x = require('./a')();
var fs = require('fs');`,
			Output: []string{`
var fs = require('fs');
var x = require('./a')();
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// Comments — leading and trailing must follow the import
		// ============================================================

		// Same-line trailing comment moves with import.
		{
			Code: `
import b from 'b'; // inline
import a from 'a';`,
			Output: []string{`
import a from 'a';
import b from 'b'; // inline
`},
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// declare module + out-of-order body
		// ============================================================

		// Body inside `declare module` independently re-ordered; top stays put.
		{
			Code: `
import fs from 'fs';
declare module 'x' {
  import sibling from './foo';
  import path from 'path';
}`,
			Output: []string{`
import fs from 'fs';
declare module 'x' {
  import path from 'path';
  import sibling from './foo';
}`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// pathGroups maxPosition rounding
		// ============================================================

		// Many `before` pathGroups on the same target group should still
		// resolve into a coherent rank ordering. Lock current behaviour.
		{
			Code: `
import a from '~/a';
import b from '~/b';
import path from 'path';`,
			Output: []string{`
import path from 'path';
import a from '~/a';
import b from '~/b';
`},
			Options: map[string]interface{}{
				"groups": []interface{}{"builtin", "external"},
				"pathGroups": []interface{}{
					map[string]interface{}{"pattern": "~/**", "group": "external", "position": "before"},
				},
				"pathGroupsExcludedImportTypes": []interface{}{"builtin"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// reverse-direction wins — fewer "after" reports than "before"
		// ============================================================

		// Two correctly-placed imports plus one stray "first" import:
		// scanning forward finds 2 stragglers, scanning reverse finds just 1
		// (the stray). Rule should pick the reverse and emit 1 report.
		{
			Code: `
import sibling from './foo';
import fs from 'fs';
import async from 'async';`,
			Output: []string{`
import fs from 'fs';
import async from 'async';
import sibling from './foo';
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// Internal-regex setting
		// ============================================================

		{
			Code: `
import a from '@my/a';
import async from 'async';`,
			Output: []string{`
import async from 'async';
import a from '@my/a';
`},
			Options: map[string]interface{}{
				// internal sits AFTER external, so '@my/a' should come last.
				"groups": []interface{}{"external", "internal"},
			},
			Settings: map[string]interface{}{
				"import/internal-regex": "^@my/",
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// distinctGroup interplay with newlines-between
		// ============================================================

		// distinctGroup=true (default) + pathGroup that DOES apply (because we
		// remove "external" from pathGroupsExcludedImportTypes) → @app/a gets
		// a sub-rank inside external, distinct from plain "external" → demands
		// a newline between them.
		{
			Code: `
import path from 'path';
import async from 'async';
import a from '@app/a';`,
			Output: []string{`
import path from 'path';

import async from 'async';

import a from '@app/a';`},
			Options: map[string]interface{}{
				"pathGroups":                    []interface{}{map[string]interface{}{"pattern": "@app/**", "group": "external", "position": "after"}},
				"pathGroupsExcludedImportTypes": []interface{}{"builtin", "object"},
				"newlines-between":              "always",
				// distinctGroup defaults to true.
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "groupNewline"},
				{MessageId: "groupNewline"},
			},
		},

		// ============================================================
		// Type imports without sortTypesGroup (default behaviour)
		// ============================================================

		// `import type` placed before non-type external when "type" group is
		// listed AFTER external in groups.
		{
			Code: `
import type {T} from 'foo';
import async from 'async';`,
			Output: []string{`
import async from 'async';
import type {T} from 'foo';
`},
			Options: map[string]interface{}{
				"groups": []interface{}{"external", "type"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// Multi-line import block — fix must preserve braces/newlines
		// ============================================================

		{
			Code: `
import {
  foo,
  bar,
} from 'b';
import a from 'a';`,
			Output: []string{`
import a from 'a';
import {
  foo,
  bar,
} from 'b';
`},
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 6, Column: 1}},
		},

		// ============================================================
		// canReorder: a non-import statement between blocks suppresses fix
		// ============================================================

		// Out-of-order is still reported, but the autofix is dropped because
		// the unrelated statement between the two requires can't be stepped
		// over safely.
		{
			Code: `
var sibling = require('./foo');
var unrelated = doStuff();
var fs = require('fs');`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// Malformed `groups` → single Program-level configError diagnostic
		// ============================================================

		{
			Code:    `import fs from 'fs';`,
			Options: map[string]interface{}{"groups": []interface{}{"not-a-known-bucket"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "configError"}},
		},
		{
			Code:    `import fs from 'fs';`,
			Options: map[string]interface{}{"groups": []interface{}{"builtin", "builtin"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "configError"}},
		},

		// ============================================================
		// Chain form — call after member access on require
		// ============================================================

		{
			Code: `
var x = require('./a').foo();
var fs = require('fs');`,
			Output: []string{`
var fs = require('fs');
var x = require('./a').foo();
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// Leading same-line comments travel with the import
		// ============================================================

		{
			Code: `
/* leader-b */ import b from 'b';
/* leader-a */ import a from 'a';`,
			Output: []string{`
/* leader-a */ import a from 'a';
/* leader-b */ import b from 'b';
`},
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// File extensions: .mts / .cts / .mjs / .cjs
		// ============================================================

		// .mts — ESM TypeScript module file. Should behave identical to .ts.
		{
			Code: `
import async from 'async';
import fs from 'fs';`,
			FileName: "edge-ext.mts",
			Output: []string{`
import fs from 'fs';
import async from 'async';
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// .cts — CommonJS TypeScript module file. Same.
		{
			Code: `
import async from 'async';
import fs from 'fs';`,
			FileName: "edge-ext.cts",
			Output: []string{`
import fs from 'fs';
import async from 'async';
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},
	}
}
