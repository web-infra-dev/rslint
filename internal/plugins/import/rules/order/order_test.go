package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Most invalid cases here intentionally check only `MessageId` (and sometimes
// the exact `Message` string) because rule_tester also validates the
// post-autofix `Output` byte-for-byte; covering position with Line/Column on
// every case would balloon the file without catching new classes of bugs.
// Where positions matter (multi-line imports, comment-bearing imports, named
// reorders), explicit Line/Column assertions are added.

func TestOrderRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		buildValidCases(),
		buildInvalidCases(),
	)
}

func buildValidCases() []rule_tester.ValidTestCase {
	return []rule_tester.ValidTestCase{
		// ============================================================
		// Default order
		// ============================================================

		// ---- Default order using require ----
		{
			Code: `
var fs = require('fs');
var async = require('async');
var relParent1 = require('../foo');
var relParent2 = require('../foo/bar');
var relParent3 = require('../');
var relParent4 = require('..');
var sibling = require('./foo');
var index = require('./');`,
		},
		// ---- Default order using import ----
		{
			Code: `
import fs from 'fs';
import async, {foo1} from 'async';
import relParent1 from '../foo';
import relParent2, {foo2} from '../foo/bar';
import relParent3 from '../';
import sibling, {foo3} from './foo';
import index from './';`,
		},
		// ---- Default order using ImportEqualsDeclaration ----
		{
			Code: `
import fs = require('fs');
import async = require('async');
import sibling = require('./foo');`,
		},
		// ---- Multiple modules of the same rank next to each other ----
		{
			Code: `
var fs = require('fs');
var fs2 = require('fs');
var path = require('path');
var _ = require('lodash');
var async = require('async');`,
		},
		// ---- import comes before require within the same group ----
		// upstream adds +100 to a require's rank, so all `require()`s sort
		// after all imports of the same classifying type.
		{
			Code: `
import fs from 'fs';
import async from 'async';
var path = require('path');
var _ = require('lodash');`,
		},
		// ---- Reverse default order via groups ----
		{
			Code: `
var index = require('./');
var sibling = require('./foo');
var relParent3 = require('../');
var relParent2 = require('../foo/bar');
var relParent1 = require('../foo');
var async = require('async');
var fs = require('fs');`,
			Options: map[string]interface{}{
				"groups": []interface{}{"index", "sibling", "parent", "external", "builtin"},
			},
		},
		// ---- Group items grouped together (array form) ----
		{
			Code: `
var path = require('path');
var sibling = require('./foo');
var fs = require('fs');
var index = require('./');`,
			Options: map[string]interface{}{
				// `[builtin, sibling]` → same rank, can interleave.
				"groups": []interface{}{
					[]interface{}{"builtin", "sibling"},
					"index",
				},
			},
		},

		// ============================================================
		// Edge: ignored / non-static / nested
		// ============================================================

		// ---- Ignore dynamic requires ----
		{
			Code: `
var path = require('path');
var _ = require('lodash');
var async = require('async');
var fs = require('f' + 's');`,
		},
		// ---- Ignore non-require call expressions ----
		{
			Code: `
var path = require('path');
var result = add(1, 2);
var _ = require('lodash');`,
		},
		// ---- Ignore requires not at top level ----
		{
			Code: `
var index = require('./');
function foo() {
	var fs = require('fs');
}
() => require('fs');
if (a) {
	require('fs');
}`,
		},
		// ---- Ignore requires in template literal ----
		{
			Code: "const foo = `${require('./a')} ${require('fs')}`",
		},
		// ---- Ignore unknown / weird-path requires (default behaviour: still checked) ----
		{
			Code: `
var unknown1 = require('/unknown1');
var fs = require('fs');
var unknown2 = require('/unknown2');`,
		},
		// ---- require in an array literal (not at statement level) ----
		{
			Code: `
const foo = [
  require('./foo'),
  require('fs'),
];`,
		},

		// ============================================================
		// Side-effect imports
		// ============================================================

		// ---- Ignore unassigned imports unless warnOnUnassignedImports ----
		{
			Code: `
import './styles.css';
import 'something-else';
import path from 'path';`,
		},

		// ============================================================
		// newlines-between
		// ============================================================

		// ---- always: one empty line between groups ----
		{
			Code: `
import fs from 'fs';

import async from 'async';

import sibling from './foo';`,
			Options: map[string]interface{}{"newlines-between": "always"},
		},
		// ---- always: same-rank consecutive imports without empty lines ----
		{
			Code: `
import fs from 'fs';
import path from 'path';

import async from 'async';`,
			Options: map[string]interface{}{"newlines-between": "always"},
		},
		// ---- never: no empty lines between any import ----
		{
			Code: `
import fs from 'fs';
import async from 'async';
import sibling from './foo';`,
			Options: map[string]interface{}{"newlines-between": "never"},
		},
		// ---- always-and-inside-groups: empty line allowed within group too ----
		{
			Code: `
import fs from 'fs';

import path from 'path';

import async from 'async';`,
			Options: map[string]interface{}{"newlines-between": "always-and-inside-groups"},
		},
		// ---- ignore: explicit no-op ----
		{
			Code: `
import fs from 'fs';


import async from 'async';
import sibling from './foo';`,
			Options: map[string]interface{}{"newlines-between": "ignore"},
		},

		// ============================================================
		// alphabetize
		// ============================================================

		// ---- asc: same group, sorted ----
		{
			Code: `
import a from 'a';
import b from 'b';
import c from 'c';`,
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
		},
		// ---- desc: same group, sorted descending ----
		{
			Code: `
import c from 'c';
import b from 'b';
import a from 'a';`,
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "desc"},
			},
		},
		// ---- caseInsensitive: PropTypes < React when ignoring case ----
		{
			Code: `
import aTypes from 'prop-types';
import React from 'react';
import { compose } from 'xcompose';`,
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc", "caseInsensitive": true},
			},
		},
		// ---- Path segment comparison: shorter path before its sub-paths ----
		// upstream treats path segments lexicographically; "foo" sorts before
		// "foo/a" because the longer side loses the tie-break.
		{
			Code: `
import c from 'foo';
import a from 'foo/a';
import b from 'foo/b';`,
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
		},
		// ---- orderImportKind asc: "type" < "value" lex, so type imports come first ----
		{
			Code: `
import type a2 from 'a';
import a from 'a';
import b from 'b';`,
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc", "orderImportKind": "asc"},
			},
		},

		// ============================================================
		// pathGroups
		// ============================================================

		// ---- @app/** placed after externals ----
		{
			Code: `
import path from 'path';
import async from 'async';
import a from '@app/foo';
import sibling from './foo';`,
			Options: map[string]interface{}{
				"pathGroups": []interface{}{
					map[string]interface{}{"pattern": "@app/**", "group": "external", "position": "after"},
				},
			},
		},
		// ---- @app/** placed before externals ----
		{
			Code: `
import path from 'path';
import a from '@app/foo';
import async from 'async';
import sibling from './foo';`,
			Options: map[string]interface{}{
				"pathGroups": []interface{}{
					map[string]interface{}{"pattern": "@app/**", "group": "external", "position": "before"},
				},
			},
		},
		// ---- pathGroupsExcludedImportTypes excludes external from path matching ----
		// "@scoped/foo" matches @app/* but is in external (excluded), so pattern doesn't apply.
		{
			Code: `
import a from '@app/a';
import b from '@app/b';
import path from 'path';`,
			Options: map[string]interface{}{
				"groups":                        []interface{}{"external", "builtin"},
				"pathGroups":                    []interface{}{map[string]interface{}{"pattern": "@app/**", "group": "external", "position": "before"}},
				"pathGroupsExcludedImportTypes": []interface{}{"builtin"},
			},
		},

		// ============================================================
		// Type-only imports (default — no sortTypesGroup)
		// ============================================================

		// ---- type group last when explicitly listed ----
		{
			Code: `
import fs from 'fs';
import async from 'async';
import type {T} from 'foo';`,
			Options: map[string]interface{}{
				"groups": []interface{}{"builtin", "external", "type"},
			},
		},
		// ---- type imports inline with values when type not in groups ----
		{
			Code: `
import fs from 'fs';
import type {T} from 'foo';
import async from 'async';`,
		},

		// ============================================================
		// sortTypesGroup
		// ============================================================

		// ---- sortTypesGroup: types sub-group mirrors values' group order ----
		{
			Code: `
import type A from 'fs';
import type C from '../foo';
import a from 'fs';
import c from '../foo';`,
			Options: map[string]interface{}{
				"groups":         []interface{}{"type", "builtin", "parent"},
				"alphabetize":    map[string]interface{}{"order": "asc"},
				"sortTypesGroup": true,
			},
		},

		// ============================================================
		// distinctGroup
		// ============================================================

		// ---- distinctGroup=false: a small rank delta (within 1) doesn't trigger
		// the "groups need a newline" check, so async + @app/a (delta 0.1) sit together. ----
		{
			Code: `
import path from 'path';

import async from 'async';
import a from '@app/a';`,
			Options: map[string]interface{}{
				"distinctGroup":    false,
				"pathGroups":       []interface{}{map[string]interface{}{"pattern": "@app/**", "group": "external", "position": "after"}},
				"newlines-between": "always",
			},
		},

		// ============================================================
		// declare module / namespace — independent scopes
		// ============================================================

		// ---- declare module: each block scoped independently ----
		{
			Code: `
import fs from 'fs';
import async from 'async';

declare module 'x' {
  import path from 'path';
  import { foo } from 'foo';
}`,
		},
		// ---- nested declare module 'A.B' ----
		{
			Code: `
declare module 'A' {
  import fs from 'fs';
  import async from 'async';
}`,
		},

		// ============================================================
		// Reserved-word + string-keyed specifiers (TS 5.0 string-named imports)
		// ============================================================

		{
			Code: `
import { default as foo } from 'a';
import { type Bar } from 'a';
import x from 'b';`,
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
		},

		// ============================================================
		// CRLF endings
		// ============================================================

		{
			Code: "import fs from 'fs';\r\nimport async from 'async';\r\n",
		},

		// ============================================================
		// node: prefix
		// ============================================================

		{
			Code: `
import fs from 'node:fs';
import path from 'node:path/posix';
import async from 'async';`,
		},

		// ============================================================
		// Empty file / only comments
		// ============================================================

		{
			Code: ``,
		},
		{
			Code: `// just a comment`,
		},
	}
}

func buildInvalidCases() []rule_tester.InvalidTestCase {
	return []rule_tester.InvalidTestCase{
		// ============================================================
		// Default order — basic violations
		// ============================================================

		// ---- 0: external before builtin ----
		// Diagnostic reports on the `require()` CallExpression (col 10 of
		// `var fs = require('fs')`), matching upstream's behaviour.
		{
			Code: `
var async = require('async');
var fs = require('fs');`,
			Output: []string{`
var fs = require('fs');
var async = require('async');
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 3, Column: 10}},
		},
		// ---- 1: sibling before parent ----
		{
			Code: `
var sibling = require('./foo');
var parent = require('../foo');`,
			Output: []string{`
var parent = require('../foo');
var sibling = require('./foo');
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},
		// ---- 2: index before sibling ----
		{
			Code: `
var index = require('./');
var sibling = require('./foo');`,
			Output: []string{`
var sibling = require('./foo');
var index = require('./');
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},
		// ---- 3: import after require with default order ----
		{
			Code: `
import sibling from './foo';
import fs = require('fs');`,
			Output: []string{`
import fs = require('fs');
import sibling from './foo';
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},
		// ---- 4: multiple out-of-order ----
		// rule_tester applies fixes iteratively until convergence; we end up
		// with two intermediate snapshots before the imports are stable.
		{
			Code: `
import sibling from './foo';
import async from 'async';
import fs from 'fs';`,
			Output: []string{
				`
import async from 'async';
import sibling from './foo';
import fs from 'fs';`,
				`
import fs from 'fs';
import async from 'async';
import sibling from './foo';
`,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "order"},
				{MessageId: "order"},
			},
		},

		// ============================================================
		// newlines-between
		// ============================================================

		// ---- 5: always — missing newline between two groups ----
		{
			Code: `
import fs from 'fs';
import async from 'async';`,
			Output: []string{`
import fs from 'fs';

import async from 'async';`},
			Options: map[string]interface{}{"newlines-between": "always"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline"}},
		},
		// ---- 6: never — extra empty line between imports ----
		{
			Code: `
import fs from 'fs';

import async from 'async';`,
			Output: []string{`
import fs from 'fs';
import async from 'async';`},
			Options: map[string]interface{}{"newlines-between": "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline"}},
		},
		// ---- 7: always — extra newline within same group ----
		{
			Code: `
import fs from 'fs';

import path from 'path';

import async from 'async';`,
			Output: []string{`
import fs from 'fs';
import path from 'path';

import async from 'async';`},
			Options: map[string]interface{}{"newlines-between": "always"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "withinGroupNewline"}},
		},

		// ============================================================
		// alphabetize
		// ============================================================

		// ---- 8: asc, same group, swap b and a ----
		{
			Code: `
import b from 'b';
import a from 'a';`,
			Output: []string{`
import a from 'a';
import b from 'b';
`},
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},
		// ---- 9: desc, same group ----
		{
			Code: `
import a from 'a';
import b from 'b';`,
			Output: []string{`
import b from 'b';
import a from 'a';
`},
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "desc"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},
		// ---- 10: caseInsensitive: 'a' before 'B' ----
		{
			Code: `
import B from 'B';
import a from 'a';`,
			Output: []string{`
import a from 'a';
import B from 'B';
`},
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc", "caseInsensitive": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// pathGroups
		// ============================================================

		// ---- 11: @app/** position after — placed before externals (wrong) ----
		{
			Code: `
import a from '@app/foo';
import path from 'path';`,
			Output: []string{`
import path from 'path';
import a from '@app/foo';
`},
			Options: map[string]interface{}{
				"pathGroups": []interface{}{
					map[string]interface{}{"pattern": "@app/**", "group": "external", "position": "after"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// Side-effect imports
		// ============================================================

		// ---- 12: warnOnUnassignedImports — side-effect out of order ----
		{
			Code: `
import fs from 'fs';
import './styles.css';
import path from 'path';`,
			Options: map[string]interface{}{"warnOnUnassignedImports": true},
			Output: []string{`
import fs from 'fs';
import path from 'path';
import './styles.css';
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// Type-only imports (no sortTypesGroup)
		// ============================================================

		// ---- 13: type group last but type appears before external ----
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
		// Multi-line imports
		// ============================================================

		// ---- 14: multi-line then misplaced ----
		{
			Code: `
import {
  a,
  b,
} from 'b';
import x from 'a';`,
			Output: []string{`
import x from 'a';
import {
  a,
  b,
} from 'b';
`},
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// Comments preservation
		// ============================================================

		// ---- 15: trailing same-line comment moves with its import ----
		{
			Code: `
import b from 'b'; // for foo
import a from 'a';`,
			Output: []string{`
import a from 'a';
import b from 'b'; // for foo
`},
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},

		// ============================================================
		// Cross-block: declare module body checked independently
		// ============================================================

		// ---- 16: top-level OK, declare module body has out-of-order ----
		{
			Code: `
import async from 'async';
import fs from 'fs';
declare module 'x' {
  import sibling from './foo';
  import fs2 from 'fs';
}`,
			Output: []string{`
import fs from 'fs';
import async from 'async';
declare module 'x' {
  import fs2 from 'fs';
  import sibling from './foo';
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "order"},
				{MessageId: "order"},
			},
		},

		// ============================================================
		// require() chain forms
		// ============================================================

		// ---- 17: var x = require('a').foo  (member access on require) ----
		{
			Code: `
var x = require('./a').foo;
var y = require('fs');`,
			Output: []string{`
var y = require('fs');
var x = require('./a').foo;
`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
		},
	}
}
