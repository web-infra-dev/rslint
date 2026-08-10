// TestOrderUpstreamCore03 migrates the upstream core cases from import-js/eslint-plugin-import at 1a39fb8672c83f8781c3a97f02eee00a7a2f0253.
// It preserves upstream messages, source positions, options, and iterative autofix outputs.
// Framework-only parser/resolver settings use rslint native equivalents; tsgo lock-ins live in order_extras_*_test.go.
package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderUpstreamCore03(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// Upstream case 0:40.
			{Code: "\n        import path from 'path';\n        import 'loud-rejection';\n        import 'something-else';\n        import _ from 'lodash';\n      ", Options: []any{map[string]any{"newlines-between": "never"}}},
			// Upstream case 0:41.
			{Code: "\n        var path = require('path');\n\n        require('loud-rejection');\n        require('something-else');\n\n        var _ = require('lodash');\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
			// Upstream case 0:42.
			{Code: "\n        var path = require('path');\n        require('loud-rejection');\n        require('something-else');\n        var _ = require('lodash');\n      ", Options: []any{map[string]any{"newlines-between": "never"}}},
			// Upstream case 0:43.
			{Code: "\n        var some = require('asdas');\n        var config = {\n          port: 4444,\n          runner: {\n            server_path: require('runner-binary').path,\n\n            cli_args: {\n                'webdriver.chrome.driver': require('browser-binary').path\n            }\n          }\n        }\n      ", Options: []any{map[string]any{"newlines-between": "never"}}},
			// Upstream case 0:44.
			{Code: "\n        var some = require('asdas');\n        var config = {\n          port: 4444,\n          runner: {\n            server_path: require('runner-binary').path,\n            cli_args: {\n                'webdriver.chrome.driver': require('browser-binary').path\n            }\n          }\n        }\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
			// Upstream case 0:45.
			{Code: "\n        var fs = require('fs');\n        var path = require('path');\n\n        var util = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n        var relParent2 = require('../');\n\n        var relParent3 = require('../bar');\n\n        var sibling = require('./foo');\n        var sibling2 = require('./bar');\n\n        var sibling3 = require('./foobar');\n      ", Options: []any{map[string]any{"newlines-between": "always-and-inside-groups"}}},
			// Upstream case 0:46.
			{Code: "\n        var fs = require('fs');\n        var path = require('path');\n        var util = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent2 } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var sibling = require('./foo');\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      ", Options: []any{map[string]any{"newlines-between": "always-and-inside-groups", "consolidateIslands": "inside-groups"}}},
			// Upstream case 0:47.
			{Code: "\n        import a from 'foo';\n        import b from 'bar';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "ignore"}}}},
			// Upstream case 0:48.
			{Code: "\n        import c from 'Bar';\n        import b from 'bar';\n        import a from 'foo';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:49.
			{Code: "\n        import a from 'foo';\n        import b from 'bar';\n        import c from 'Bar';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 0:50.
			{Code: "\n        import a from \"foo\";\n        import c from \"foo/bar\";\n        import d from \"foo/barfoo\";\n        import b from \"foo-bar\";\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:51.
			{Code: "\n        import a from \"foo\";\n        import c from \"foo/foobar/bar\";\n        import d from \"foo/foobar/barfoo\";\n        import b from \"foo-bar\";\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:52.
			{Code: "\n        import b from \"foo-bar\";\n        import d from \"foo/barfoo\";\n        import c from \"foo/bar\";\n        import a from \"foo\";\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 0:53.
			{Code: "\n        import b from \"foo-bar\";\n        import c from \"foo,bar\";\n        import d from \"foo/barfoo\";\n        import a from \"foo\";", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 0:54.
			{Code: "\n        import b from 'Bar';\n        import c from 'bar';\n        import a from 'foo';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "asc"}, "newlines-between": "always"}}},
			// Upstream case 0:55.
			{Code: "\n        import { hello } from './hello';\n        import { int } from './int';\n        const blah = require('./blah');\n        const { cello } = require('./cello');\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:56.
			{Code: "\n        import React from 'react';\n        import { BrowserRouter } from 'react-router-dom';\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:57.
			{Code: "\n        import { UserInputError } from 'apollo-server-express';\n\n        import { new as assertNewEmail } from '~/Assertions/Email';\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"caseInsensitive": true, "order": "asc"}, "pathGroups": []any{map[string]any{"pattern": "~/*", "group": "internal"}}, "groups": []any{"builtin", "external", "internal", "parent", "sibling", "index"}, "newlines-between": "always"}}},
			// Upstream case 0:58.
			{Code: "\n        import { ReactElement, ReactNode } from 'react';\n\n        import { util } from 'Internal/lib';\n\n        import { parent } from '../parent';\n\n        import { sibling } from './sibling';\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"caseInsensitive": true, "order": "asc"}, "pathGroups": []any{map[string]any{"pattern": "Internal/**/*", "group": "internal"}}, "groups": []any{"builtin", "external", "internal", "parent", "sibling", "index"}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}}}},
			// Upstream case 0:59.
			{Code: "\n        import fs from 'fs';\n\n        import express from 'express';\n\n        import service from '@/api/service';\n\n        import fooParent from '../foo';\n\n        import fooSibling from './foo';\n\n        import index from './';\n\n        import internalDoesNotExistSoIsUnknown from '@/does-not-exist';\n      ", Options: []any{map[string]any{"groups": []any{"builtin", "external", "internal", "parent", "sibling", "index", "unknown"}, "newlines-between": "always"}}, TSConfig: "tsconfig.order-paths.json"},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
