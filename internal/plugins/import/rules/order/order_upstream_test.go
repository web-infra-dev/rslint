package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestOrderRuleUpstream mirrors representative cases from
// eslint-plugin-import's `tests/src/rules/order.js`. Each block keeps the
// upstream comment header so a reader can cross-reference against the
// original. Message strings are asserted byte-for-byte to guard the public
// contract.
//
// Cases that depend on framework features rslint deliberately doesn't expose
// (TypeScript parser variants, babel-flow, `eslint-module-utils/resolve` with
// custom resolver paths) are noted inline as `// SKIP: <reason>` and not
// migrated.
func TestOrderRuleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		buildUpstreamValidCases(),
		buildUpstreamInvalidCases(),
	)
}

func buildUpstreamValidCases() []rule_tester.ValidTestCase {
	return []rule_tester.ValidTestCase{
		// ============================================================
		// Default order using require / import / mixed
		// ============================================================

		// upstream: line ~26 (default order using require)
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

		// upstream: line ~39 (default order using import)
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

		// upstream: line ~50 (multiple modules of same rank)
		{
			Code: `
var fs = require('fs');
var fs2 = require('fs');
var path = require('path');
var _ = require('lodash');
var async = require('async');`,
		},

		// upstream: line ~59 (reverse default order via groups)
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

		// upstream: line ~71 (groups with same-rank items in array)
		{
			Code: `
var fs = require('fs');
var index = require('./');
var path = require('path');
var sibling = require('./foo');`,
			Options: map[string]interface{}{
				"groups": []interface{}{
					[]interface{}{"builtin", "index"},
					[]interface{}{"sibling", "parent"},
				},
			},
		},

		// upstream: line ~80 (ignore dynamic require)
		{
			Code: `
var path = require('path');
var _ = require('lodash');
var async = require('async');
var fs = require('f' + 's');`,
		},

		// upstream: line ~87 (ignore non-require call)
		{
			Code: `
var path = require('path');
var result = add(1, 2);
var _ = require('lodash');`,
		},

		// upstream: line ~99 (ignore requires not at top level)
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

		// upstream: line ~107 (ignore template-literal require)
		{
			Code: "const foo = `${require('./a')} ${require('fs')}`",
		},

		// upstream: line ~111 (ignore unknown / weird-path requires)
		{
			Code: `
var unknown1 = require('/unknown1');
var fs = require('fs');
var unknown2 = require('/unknown2');
var async = require('async');
var unknown3 = require('/unknown3');
var foo = require('../foo');
var unknown4 = require('/unknown4');
var bar = require('../foo/bar');
var unknown5 = require('/unknown5');
var parent = require('../');
var unknown6 = require('/unknown6');
var sibling = require('./foo');
var unknown7 = require('/unknown7');
var index = require('./');`,
		},

		// ============================================================
		// Newlines-between
		// ============================================================

		// upstream: line ~226 (newlines-between: always — valid)
		// `path` (builtin) sits in the same array-group as `./` (index), so it
		// can come before `./foo` (sibling) without a newline-violation.
		{
			Code: `
var fs = require('fs');
var path = require('path');
var index = require('./');

var sibling = require('./foo');`,
			Options: map[string]interface{}{
				"newlines-between": "always",
				"groups":           []interface{}{[]interface{}{"builtin", "index"}, "sibling", []interface{}{"parent", "external"}},
			},
		},

		// upstream: line ~261 (always-and-inside-groups — valid)
		{
			Code: `
var fs = require('fs');

var path = require('path');

var sibling = require('./foo');`,
			Options: map[string]interface{}{
				"newlines-between": "always-and-inside-groups",
			},
		},

		// upstream: line ~284 (newlines-between: never — valid)
		{
			Code: `
var fs = require('fs');
var path = require('path');
var index = require('./');
var sibling = require('./foo');`,
			Options: map[string]interface{}{
				"newlines-between": "never",
				"groups":           []interface{}{[]interface{}{"builtin", "index"}, "sibling", []interface{}{"parent", "external"}},
			},
		},

		// ============================================================
		// alphabetize
		// ============================================================

		// upstream: alphabetize asc valid
		{
			Code: `
import async from 'async';
import lodash from 'lodash';`,
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
		},

		// upstream: alphabetize asc with type kinds
		{
			Code: `
import a from 'a';
import type {x} from 'a';
import b from 'b';`,
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc"},
			},
		},

		// upstream: alphabetize desc
		{
			Code: `
import lodash from 'lodash';
import async from 'async';`,
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "desc"},
			},
		},

		// upstream: caseInsensitive desc — high → low alphabetically
		{
			Code: `
import { compose } from 'xcompose';
import React from 'react';
import aTypes from 'prop-types';`,
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "desc", "caseInsensitive": true},
			},
		},

		// ============================================================
		// pathGroups
		// ============================================================

		// upstream: simple pathGroup before
		{
			Code: `
import a from '@app/a';
import path from 'path';`,
			Options: map[string]interface{}{
				"groups": []interface{}{"external", "builtin"},
				"pathGroups": []interface{}{
					map[string]interface{}{"pattern": "@app/**", "group": "external", "position": "before"},
				},
				"pathGroupsExcludedImportTypes": []interface{}{"builtin"},
			},
		},

		// upstream: pathGroup after with custom newlines
		{
			Code: `
import path from 'path';

import async from 'async';

import a from '@app/a';

import sibling from './foo';`,
			Options: map[string]interface{}{
				"newlines-between":              "always",
				"pathGroups":                    []interface{}{map[string]interface{}{"pattern": "@app/**", "group": "external", "position": "after"}},
				"pathGroupsExcludedImportTypes": []interface{}{},
			},
		},
	}
}

func buildUpstreamInvalidCases() []rule_tester.InvalidTestCase {
	return []rule_tester.InvalidTestCase{
		// ============================================================
		// Basic upstream invalid cases with exact message text
		// ============================================================

		// upstream: line 1295 (builtin before external — require)
		{
			Code: `
var async = require('async');
var fs = require('fs');`,
			Output: []string{`
var fs = require('fs');
var async = require('async');
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "order",
					Message:   "`fs` import should occur before import of `async`",
				},
			},
		},

		// upstream: line 1310 (trailing whitespace preserved by fix)
		{
			Code: "\nvar async = require('async');\nvar fs = require('fs'); \n",
			Output: []string{
				"\nvar fs = require('fs'); \nvar async = require('async');\n",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "order", Message: "`fs` import should occur before import of `async`"},
			},
		},

		// upstream: line 1324 (trailing comment kept with import)
		{
			Code: `
var async = require('async');
var fs = require('fs'); /* comment */`,
			Output: []string{`
var fs = require('fs'); /* comment */
var async = require('async');
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "order", Message: "`fs` import should occur before import of `async`"},
			},
		},

		// upstream: line ~1393 (destructured require — names alphabetized independently)
		// We don't enable `named: true`, so the order rule treats these as plain
		// requires and just sorts by module spec.
		{
			Code: `
var {b} = require('async');
var {a} = require('fs');`,
			Output: []string{`
var {a} = require('fs');
var {b} = require('async');
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "order", Message: "`fs` import should occur before import of `async`"},
			},
		},

		// ============================================================
		// Multi-line require — fix preserves the whole declaration
		// ============================================================

		// upstream: line 1407 (multi-line require)
		{
			Code: `
var async = require('async');
var fs =
	require('fs');`,
			Output: []string{`
var fs =
	require('fs');
var async = require('async');
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "order", Message: "`fs` import should occur before import of `async`"},
			},
		},

		// ============================================================
		// Imports (not require) with same wording
		// ============================================================

		{
			Code: `
import async from 'async';
import fs from 'fs';`,
			Output: []string{`
import fs from 'fs';
import async from 'async';
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "order", Message: "`fs` import should occur before import of `async`"},
			},
		},

		// ---- Type import message ----
		{
			Code: `
import async from 'async';
import type {T} from 'fs';`,
			Output: []string{`
import type {T} from 'fs';
import async from 'async';
`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "order",
					Message:   "`fs` type import should occur before import of `async`",
				},
			},
		},

		// ============================================================
		// newlines-between violations with exact message
		// ============================================================

		// upstream-style: newlines-between always missing
		{
			Code: `
var fs = require('fs');
var path = require('./foo');`,
			Output: []string{`
var fs = require('fs');

var path = require('./foo');`},
			Options: map[string]interface{}{"newlines-between": "always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "groupNewline",
					Message:   "There should be at least one empty line between import groups",
				},
			},
		},

		// newlines-between never with extra blank lines
		{
			Code: `
var fs = require('fs');

var path = require('./foo');`,
			Output: []string{`
var fs = require('fs');
var path = require('./foo');`},
			Options: map[string]interface{}{"newlines-between": "never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "groupNewline",
					Message:   "There should be no empty line between import groups",
				},
			},
		},

		// ============================================================
		// alphabetize violations
		// ============================================================

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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "order", Message: "`a` import should occur before import of `b`"},
			},
		},

		// alphabetize desc
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "order", Message: "`b` import should occur before import of `a`"},
			},
		},

		// caseInsensitive: 'AaA' before 'b' insensitive
		{
			Code: `
import b from 'b';
import AaA from 'AaA';`,
			Output: []string{`
import AaA from 'AaA';
import b from 'b';
`},
			Options: map[string]interface{}{
				"alphabetize": map[string]interface{}{"order": "asc", "caseInsensitive": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "order", Message: "`AaA` import should occur before import of `b`"},
			},
		},
	}
}
