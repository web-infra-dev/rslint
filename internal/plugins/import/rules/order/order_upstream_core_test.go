// cspell:ignore asdas barfoo nonegate
//
// The TestOrderUpstreamCore* functions migrate the upstream core suite from
// eslint-plugin-import v2.32.0 (tests/src/rules/order.js).
// They preserve upstream messages, source positions, options, and iterative
// autofix outputs. rslint-specific lock-ins live in order_extras_*_test.go.
package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderUpstreamCore01(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// Upstream case 0:0.
			{Code: "\n        var fs = require('fs');\n        var async = require('async');\n        var relParent1 = require('../foo');\n        var relParent2 = require('../foo/bar');\n        var relParent3 = require('../');\n        var relParent4 = require('..');\n        var sibling = require('./foo');\n        var index = require('./');"},
			// Upstream case 0:1.
			{Code: "\n        import fs from 'fs';\n        import async, {foo1} from 'async';\n        import relParent1 from '../foo';\n        import relParent2, {foo2} from '../foo/bar';\n        import relParent3 from '../';\n        import sibling, {foo3} from './foo';\n        import index from './';"},
			// Upstream case 0:2.
			{Code: "\n        var fs = require('fs');\n        var fs = require('fs');\n        var path = require('path');\n        var _ = require('lodash');\n        var async = require('async');"},
			// Upstream case 0:3.
			{Code: "\n        var index = require('./');\n        var sibling = require('./foo');\n        var relParent3 = require('../');\n        var relParent2 = require('../foo/bar');\n        var relParent1 = require('../foo');\n        var async = require('async');\n        var fs = require('fs');\n      ", Options: []any{map[string]any{"groups": []any{"index", "sibling", "parent", "external", "builtin"}}}},
			// Upstream case 0:4.
			{Code: "\n        var path = require('path');\n        var _ = require('lodash');\n        var async = require('async');\n        var fs = require('f' + 's');"},
			// Upstream case 0:5.
			{Code: "\n        var path = require('path');\n        var result = add(1, 2);\n        var _ = require('lodash');"},
			// Upstream case 0:6.
			{Code: "\n        var index = require('./');\n        function foo() {\n          var fs = require('fs');\n        }\n        () => require('fs');\n        if (a) {\n          require('fs');\n        }"},
			// Upstream case 0:7.
			{Code: "\n        const foo = [\n          require('./foo'),\n          require('fs'),\n        ]"},
			// Upstream case 0:8.
			{Code: "const foo = `${require('./a')} ${require('fs')}`"},
			// Upstream case 0:9.
			{Code: "\n        var unknown1 = require('/unknown1');\n        var fs = require('fs');\n        var unknown2 = require('/unknown2');\n        var async = require('async');\n        var unknown3 = require('/unknown3');\n        var foo = require('../foo');\n        var unknown4 = require('/unknown4');\n        var bar = require('../foo/bar');\n        var unknown5 = require('/unknown5');\n        var parent = require('../');\n        var unknown6 = require('/unknown6');\n        var foo = require('./foo');\n        var unknown7 = require('/unknown7');\n        var index = require('./');\n        var unknown8 = require('/unknown8');\n    "},
			// Upstream case 0:10.
			{Code: "\n        require('./foo');\n        require('fs');\n        var path = require('path');\n    "},
			// Upstream case 0:11.
			{Code: "\n        import './foo';\n        import 'fs';\n        import path from 'path';\n    "},
			// Upstream case 0:12.
			{Code: "\n        function add(a, b) {\n          return a + b;\n        }\n        var foo;\n    "},
			// Upstream case 0:13.
			{Code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n\n        var sibling = require('./foo');\n        var relParent3 = require('../');\n        var async = require('async');\n        var relParent1 = require('../foo');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling", "parent", "external"}}}}},
			// Upstream case 0:14.
			{Code: "\n        import async from 'async';\n        import fs from 'fs';\n        import path from 'path';\n\n        import index from '.';\n        import relParent3 from '../';\n        import relParent1 from '../foo';\n        import sibling from './foo';\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "external"}}, "alphabetize": map[string]any{"order": "asc", "caseInsensitive": true}}}},
			// Upstream case 0:15.
			{Code: "\n      import { fooz } from '../baz.js'\n      import { foo } from './bar.js'\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc", "caseInsensitive": true}, "groups": []any{"builtin", "external", "internal", []any{"parent", "sibling", "index"}, "object"}, "newlines-between": "always", "warnOnUnassignedImports": true}}},
			// Upstream case 0:16.
			{Code: "\n        var index = require('./');\n        var path = require('path');\n      ", Options: []any{map[string]any{"groups": []any{"index", []any{"sibling", "parent", "external"}}}}},
			// Upstream case 0:17.
			{Code: "\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n        import sibling, {foo3} from './foo';\n        var fs = require('fs');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var index = require('./');\n      "},
			// Upstream case 0:18.
			{Code: "\n          import async, {foo1} from 'async';\n          import relParent2, {foo2} from '../foo/bar';\n          import sibling, {foo3} from './foo';\n          var fs = require('fs');\n          var util = require(\"util\");\n          var relParent1 = require('../foo');\n          var relParent3 = require('../');\n          var index = require('./');\n        "},
		},
		[]rule_tester.InvalidTestCase{},
	)
}

func TestOrderUpstreamCore02(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// Upstream case 0:19.
			{Code: "\n          export import CreateSomething = _CreateSomething;\n        "},
			// Upstream case 0:20.
			{Code: "\n        import fs from 'fs';\n        import { Input } from '-/components/Input';\n        import { Button } from '-/components/Button';\n        import { add } from './helper';", Options: []any{map[string]any{"groups": []any{"builtin", "external", "unknown", "parent", "sibling", "index"}}}},
			// Upstream case 0:21.
			{Code: "\n        import { Input } from '-/components/Input';\n        import { Button } from '-/components/Button';\n        import fs from 'fs';\n        import { add } from './helper';", Options: []any{map[string]any{"groups": []any{"unknown", "builtin", "external", "parent", "sibling", "index"}}}},
			// Upstream case 0:22.
			{Code: "\n        import fs from 'fs';\n\n        import { Input } from '-/components/Input';\n        import { Button } from '-/components/Button';\n\n        import p from '..';\n        import q from '../';\n\n        import { add } from './helper';\n\n        import i from '.';\n        import j from './';", Options: []any{map[string]any{"newlines-between": "always", "groups": []any{"builtin", "external", "unknown", "parent", "sibling", "index"}}}},
			// Upstream case 0:23.
			{Code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { Input } from '~/components/Input';\n        import { Button } from '#/components/Button';\n        import { add } from './helper';", Options: []any{map[string]any{"pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external", "position": "after"}, map[string]any{"pattern": "#/**", "group": "external", "position": "after"}}}}},
			// Upstream case 0:24.
			{Code: "\n        import fs from 'fs';\n        import { Input } from '~/components/Input';\n        import async from 'async';\n        import { Button } from '#/components/Button';\n        import _ from 'lodash';\n        import { add } from './helper';", Options: []any{map[string]any{"pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external"}, map[string]any{"pattern": "#/**", "group": "external"}}}}},
			// Upstream case 0:25.
			{Code: "\n        import fs from 'fs';\n\n        import { Input } from '~/components/Input';\n\n        import { Button } from '#/components/Button';\n\n        import _ from 'lodash';\n\n        import { add } from './helper';", Options: []any{map[string]any{"newlines-between": "always", "pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external", "position": "before"}, map[string]any{"pattern": "#/**", "group": "external", "position": "before"}}}}},
			// Upstream case 0:26.
			{Code: "\n        import fs from 'fs';\n\n        import _ from 'lodash';\n\n        import { Input } from '~/components/Input';\n\n        import { Button } from '!/components/Button';\n\n        import { add } from './helper';", Options: []any{map[string]any{"newlines-between": "always", "pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external", "position": "after"}, map[string]any{"pattern": "!/**", "patternOptions": map[string]any{"nonegate": true}, "group": "external", "position": "after"}}}}},
			// Upstream case 0:27.
			{Code: "\n        import fs from 'fs';\n\n        import { Input } from '@app/components/Input';\n\n        import { Button } from '@app2/components/Button';\n\n        import _ from 'lodash';\n\n        import { add } from './helper';", Options: []any{map[string]any{"newlines-between": "always", "pathGroupsExcludedImportTypes": []any{"builtin"}, "pathGroups": []any{map[string]any{"pattern": "@app/**", "group": "external", "position": "before"}, map[string]any{"pattern": "@app2/**", "group": "external", "position": "before"}}}}},
			// Upstream case 0:28.
			{Code: "\n        import fs from 'fs';\n        import external from 'external';\n        import externalTooPlease from './make-me-external';\n\n        import sibling from './sibling';", Options: []any{map[string]any{"newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}, "pathGroups": []any{map[string]any{"pattern": "./make-me-external", "group": "external"}}, "groups": []any{[]any{"builtin", "external"}, "internal", "parent", "sibling", "index"}}}},
			// Upstream case 0:29.
			{Code: "\n        import _ from 'lodash';\n        import m from '@test-scope/some-module';\n\n        import bar from './bar';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}, Settings: map[string]any{"import/external-module-folders": []any{"node_modules", "symlinked-module"}}},
			// Upstream case 0:30.
			{Code: "\n        import _ from 'lodash';\n        import m from '@test-scope/some-module';\n\n        import bar from './bar';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}, Settings: map[string]any{"import/external-module-folders": []any{"node_modules", "symlinked-module"}}},
			// Upstream case 0:31.
			{Code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n\n\n\n        var sibling = require('./foo');\n\n\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling"}, []any{"parent", "external"}}, "newlines-between": "always"}}},
			// Upstream case 0:32.
			{Code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling"}, []any{"parent", "external"}}, "newlines-between": "never"}}},
			// Upstream case 0:33.
			{Code: "\n      var fs = require('fs');\n\n      var index = require('./');\n      var path = require('path');\n      var sibling = require('./foo');\n\n\n      var relParent1 = require('../foo');\n\n      var relParent3 = require('../');\n      var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling"}, []any{"parent", "external"}}, "newlines-between": "ignore"}}},
			// Upstream case 0:34.
			{Code: "\n      var fs = require('fs');\n\n      var index = require('./');\n      var path = require('path');\n      var sibling = require('./foo');\n\n\n      var relParent1 = require('../foo');\n\n      var relParent3 = require('../');\n\n      var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling"}, []any{"parent", "external"}}}}},
			// Upstream case 0:35.
			{Code: "\n        import path from 'path';\n\n        import {\n            I,\n            Want,\n            Couple,\n            Imports,\n            Here\n        } from 'bar';\n        import external from 'external'\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
			// Upstream case 0:36.
			{Code: "\n        import path from 'path';\n        import net\n          from 'net';\n\n        import external from 'external'\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
			// Upstream case 0:37.
			{Code: "\n        import foo\n          from '../../../../this/will/be/very/long/path/and/therefore/this/import/has/to/be/in/two/lines';\n\n        import bar\n          from './sibling';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
			// Upstream case 0:38.
			{Code: "\n        import path from 'path';\n\n        import 'loud-rejection';\n        import 'something-else';\n\n        import _ from 'lodash';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
		},
		[]rule_tester.InvalidTestCase{},
	)
}

func TestOrderUpstreamCore03(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// Upstream case 0:39.
			{Code: "\n        import path from 'path';\n        import 'loud-rejection';\n        import 'something-else';\n        import _ from 'lodash';\n      ", Options: []any{map[string]any{"newlines-between": "never"}}},
			// Upstream case 0:40.
			{Code: "\n        var path = require('path');\n\n        require('loud-rejection');\n        require('something-else');\n\n        var _ = require('lodash');\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
			// Upstream case 0:41.
			{Code: "\n        var path = require('path');\n        require('loud-rejection');\n        require('something-else');\n        var _ = require('lodash');\n      ", Options: []any{map[string]any{"newlines-between": "never"}}},
			// Upstream case 0:42.
			{Code: "\n        var some = require('asdas');\n        var config = {\n          port: 4444,\n          runner: {\n            server_path: require('runner-binary').path,\n\n            cli_args: {\n                'webdriver.chrome.driver': require('browser-binary').path\n            }\n          }\n        }\n      ", Options: []any{map[string]any{"newlines-between": "never"}}},
			// Upstream case 0:43.
			{Code: "\n        var some = require('asdas');\n        var config = {\n          port: 4444,\n          runner: {\n            server_path: require('runner-binary').path,\n            cli_args: {\n                'webdriver.chrome.driver': require('browser-binary').path\n            }\n          }\n        }\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
			// Upstream case 0:44.
			{Code: "\n        var fs = require('fs');\n        var path = require('path');\n\n        var util = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n        var relParent2 = require('../');\n\n        var relParent3 = require('../bar');\n\n        var sibling = require('./foo');\n        var sibling2 = require('./bar');\n\n        var sibling3 = require('./foobar');\n      ", Options: []any{map[string]any{"newlines-between": "always-and-inside-groups"}}},
			// Upstream case 0:45.
			{Code: "\n        var fs = require('fs');\n        var path = require('path');\n        var util = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent2 } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var sibling = require('./foo');\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      ", Options: []any{map[string]any{"newlines-between": "always-and-inside-groups", "consolidateIslands": "inside-groups"}}},
			// Upstream case 0:46.
			{Code: "\n        import a from 'foo';\n        import b from 'bar';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "ignore"}}}},
			// Upstream case 0:47.
			{Code: "\n        import c from 'Bar';\n        import b from 'bar';\n        import a from 'foo';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:48.
			{Code: "\n        import a from 'foo';\n        import b from 'bar';\n        import c from 'Bar';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 0:49.
			{Code: "\n        import a from \"foo\";\n        import c from \"foo/bar\";\n        import d from \"foo/barfoo\";\n        import b from \"foo-bar\";\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:50.
			{Code: "\n        import a from \"foo\";\n        import c from \"foo/foobar/bar\";\n        import d from \"foo/foobar/barfoo\";\n        import b from \"foo-bar\";\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:51.
			{Code: "\n        import b from \"foo-bar\";\n        import d from \"foo/barfoo\";\n        import c from \"foo/bar\";\n        import a from \"foo\";\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 0:52.
			{Code: "\n        import b from \"foo-bar\";\n        import c from \"foo,bar\";\n        import d from \"foo/barfoo\";\n        import a from \"foo\";", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 0:53.
			{Code: "\n        import b from 'Bar';\n        import c from 'bar';\n        import a from 'foo';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "asc"}, "newlines-between": "always"}}},
			// Upstream case 0:54.
			{Code: "\n        import { hello } from './hello';\n        import { int } from './int';\n        const blah = require('./blah');\n        const { cello } = require('./cello');\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:55.
			{Code: "\n        import React from 'react';\n        import { BrowserRouter } from 'react-router-dom';\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:56.
			{Code: "\n        import { UserInputError } from 'apollo-server-express';\n\n        import { new as assertNewEmail } from '~/Assertions/Email';\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"caseInsensitive": true, "order": "asc"}, "pathGroups": []any{map[string]any{"pattern": "~/*", "group": "internal"}}, "groups": []any{"builtin", "external", "internal", "parent", "sibling", "index"}, "newlines-between": "always"}}},
			// Upstream case 0:57.
			{Code: "\n        import { ReactElement, ReactNode } from 'react';\n\n        import { util } from 'Internal/lib';\n\n        import { parent } from '../parent';\n\n        import { sibling } from './sibling';\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"caseInsensitive": true, "order": "asc"}, "pathGroups": []any{map[string]any{"pattern": "Internal/**/*", "group": "internal"}}, "groups": []any{"builtin", "external", "internal", "parent", "sibling", "index"}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}}}},
			// Upstream case 0:58.
			{Code: "\n        import fs from 'fs';\n\n        import express from 'express';\n\n        import service from '@/api/service';\n\n        import fooParent from '../foo';\n\n        import fooSibling from './foo';\n\n        import index from './';\n\n        import internalDoesNotExistSoIsUnknown from '@/does-not-exist';\n      ", Options: []any{map[string]any{"groups": []any{"builtin", "external", "internal", "parent", "sibling", "index", "unknown"}, "newlines-between": "always"}}, TSConfig: "tsconfig.order-paths.json"},
		},
		[]rule_tester.InvalidTestCase{},
	)
}

func TestOrderUpstreamCore04(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// Upstream case 0:59.
			{Code: "\n        import A from 'a';\n\n        import C from 'c';\n\n        import B from 'b';\n      ", Options: []any{map[string]any{"newlines-between": "always", "distinctGroup": true, "pathGroupsExcludedImportTypes": []any{}, "pathGroups": []any{map[string]any{"pattern": "a", "group": "external", "position": "before"}, map[string]any{"pattern": "b", "group": "external", "position": "after"}}}}},
			// Upstream case 0:60.
			{Code: "\n        import A from 'a';\n        import C from 'c';\n        import B from 'b';\n      ", Options: []any{map[string]any{"newlines-between": "always", "distinctGroup": false, "pathGroupsExcludedImportTypes": []any{}, "pathGroups": []any{map[string]any{"pattern": "a", "group": "external", "position": "before"}, map[string]any{"pattern": "b", "group": "external", "position": "after"}}}}},
			// Upstream case 0:61.
			{Code: "\n        import A from 'a';\n\n        import b from './b';\n        import B from './B';\n      ", Options: []any{map[string]any{"newlines-between": "always", "distinctGroup": false, "pathGroupsExcludedImportTypes": []any{}, "pathGroups": []any{map[string]any{"pattern": "a", "group": "external"}, map[string]any{"pattern": "b", "group": "internal", "position": "before"}}}}},
			// Upstream case 0:62.
			{Code: "\n        import A from \"baz\";\n        import B from \"Bar\";\n        import C from \"Foo\";\n\n        import D from \"..\";\n        import E from \"../\";\n        import F from \"../baz\";\n        import G from \"../Bar\";\n        import H from \"../Foo\";\n\n        import I from \".\";\n        import J from \"./baz\";\n        import K from \"./Bar\";\n        import L from \"./Foo\";\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"caseInsensitive": false, "order": "asc"}, "newlines-between": "always", "groups": []any{[]any{"builtin", "external", "internal", "unknown", "object", "type"}, "parent", []any{"sibling", "index"}}, "distinctGroup": false, "pathGroupsExcludedImportTypes": []any{}, "pathGroups": []any{map[string]any{"pattern": "./", "group": "sibling", "position": "before"}, map[string]any{"pattern": ".", "group": "sibling", "position": "before"}, map[string]any{"pattern": "..", "group": "parent", "position": "before"}, map[string]any{"pattern": "../", "group": "parent", "position": "before"}, map[string]any{"pattern": "[a-z]*", "group": "external", "position": "before"}, map[string]any{"pattern": "../[a-z]*", "group": "parent", "position": "before"}, map[string]any{"pattern": "./[a-z]*", "group": "sibling", "position": "before"}}}}},
			// Upstream case 0:63.
			{Code: "\n        import B from './B';\n        import b from './b';\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc", "orderImportKind": "asc", "caseInsensitive": true}}}},
			// Upstream case 0:64.
			{Code: "\n        import { a, B as C, Z } from './Z';\n        const { D, n: c, Y } = require('./Z');\n        export { C, D };\n        export { A, B, C as default } from \"./Z\";\n\n        const { [\"ignore require-statements with non-identifier imports\"]: z, d } = require(\"./Z\");\n        exports = { [\"ignore exports statements with non-identifiers\"]: Z, D };\n      ", Options: []any{map[string]any{"named": true, "alphabetize": map[string]any{"order": "asc", "caseInsensitive": true}}}},
			// Upstream case 0:65.
			{Code: "\n        const { b, A } = require('./Z');\n      ", Options: []any{map[string]any{"named": true, "alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 0:66.
			{Code: "\n        import { A, B } from \"./Z\";\n        export { Z, A } from \"./Z\";\n        export { N, P } from \"./Z\";\n        const { X, Y } = require(\"./Z\");\n      ", Options: []any{map[string]any{"named": map[string]any{"require": true, "import": true, "export": false}}}},
			// Upstream case 0:67.
			{Code: "\n        import { B, A } from \"./Z\";\n        const { D, C } = require(\"./Z\");\n        export { B, A } from \"./Z\";\n      ", Options: []any{map[string]any{"named": map[string]any{"require": false, "import": false, "export": false}}}},
			// Upstream case 0:68.
			{Code: "\n        import { B, A, R } from \"foo\";\n        const { D, O, G } = require(\"tunes\");\n        export { B, A, Z } from \"foo\";\n      ", Options: []any{map[string]any{"named": map[string]any{"enabled": false}}}},
			// Upstream case 0:69.
			{Code: "\n        import { A as A, A as B, A as C } from \"./Z\";\n        const { a, a: b, a: c } = require(\"./Z\");\n      ", Options: []any{map[string]any{"named": true}}},
			// Upstream case 0:70.
			{Code: "\n        import { A, B, C } from \"./Z\";\n        exports = { A, B, C };\n        module.exports = { a: A, b: B, c: C };\n      ", Options: []any{map[string]any{"named": map[string]any{"cjsExports": true}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:71.
			{Code: "\n        module.exports.A = { };\n        module.exports.A.B = { };\n        module.exports.B = { };\n        exports.C = { };\n      ", Options: []any{map[string]any{"named": map[string]any{"cjsExports": true}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:72.
			{Code: "\n        var exports = null;\n        var module = null;\n        exports = { };\n        module = { };\n        module.exports = { };\n        module.exports.U = { };\n        module.exports.N = { };\n        module.exports.C = { };\n        exports.L = { };\n        exports.E = { };\n      ", Options: []any{map[string]any{"named": map[string]any{"cjsExports": true}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:73.
			{Code: "\n        exports[\"B\"] = { };\n        exports[\"C\"] = { };\n        exports[\"A\"] = { };\n      ", Options: []any{map[string]any{"named": map[string]any{"cjsExports": true}, "alphabetize": map[string]any{"order": "asc"}}}},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream case 0:0.
			{Code: "\n        var async = require('async');\n        var fs = require('fs');\n      ", Output: []string{"\n        var fs = require('fs');\n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:1.
			{Code: "\n        var async = require('async');\n        var fs = require('fs'); \n      ", Output: []string{"\n        var fs = require('fs'); \n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:2.
			{Code: "\n        var async = require('async');\n        var fs = require('fs'); /* comment */\n      ", Output: []string{"\n        var fs = require('fs'); /* comment */\n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:3.
			{Code: "\n        /* comment1 */  var async = require('async'); /* comment2 */\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n      ", Output: []string{"\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n        /* comment1 */  var async = require('async'); /* comment2 */\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 34, EndLine: 3, EndColumn: 47}}},
		},
	)
}

func TestOrderUpstreamCore05(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Upstream case 0:4.
			{Code: "\n        /* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n      ", Output: []string{"\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n        /* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 34, EndLine: 3, EndColumn: 47}}},
			// Upstream case 0:5.
			{Code: "/* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\r\n/* comment3 */  var fs = require('fs'); /* comment4 */\r\n", Output: []string{"/* comment3 */  var fs = require('fs'); /* comment4 */\r\n/* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\r\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 2, Column: 26, EndLine: 2, EndColumn: 39}}},
			// Upstream case 0:6.
			{Code: "\n        /* multi-line1\n          comment1 */  var async = require('async'); /* multi-line2\n          comment2 */  var fs = require('fs'); /* multi-line3\n          comment3 */\n      ", Output: []string{"\n        /* multi-line1\n          comment1 */  var fs = require('fs'); \n  var async = require('async'); /* multi-line2\n          comment2 *//* multi-line3\n          comment3 */\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 4, Column: 33, EndLine: 4, EndColumn: 46}}},
			// Upstream case 0:7.
			{Code: "\n        var {b} = require('async');\n        var {a} = require('fs');\n      ", Output: []string{"\n        var {a} = require('fs');\n        var {b} = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 19, EndLine: 3, EndColumn: 32}}},
			// Upstream case 0:8.
			{Code: "\n        var async = require('async');\n        var fs =\n          require('fs');\n      ", Output: []string{"\n        var fs =\n          require('fs');\n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 4, Column: 11, EndLine: 4, EndColumn: 24}}},
			// Upstream case 0:9.
			{Code: "\n        var async = require('async');\n        var fs = require('fs');", Output: []string{"\n        var fs = require('fs');\n        var async = require('async');\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:10.
			{Code: "\n        import async from 'async';\n        import fs from 'fs';\n      ", Output: []string{"\n        import fs from 'fs';\n        import async from 'async';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 9, EndLine: 3, EndColumn: 29}}},
			// Upstream case 0:11.
			{Code: "\n        var async = require('async');\n        import fs from 'fs';\n      ", Output: []string{"\n        import fs from 'fs';\n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 9, EndLine: 3, EndColumn: 29}}},
			// Upstream case 0:12.
			{Code: "\n        var parent = require('../parent');\n        var async = require('async');\n      ", Output: []string{"\n        var async = require('async');\n        var parent = require('../parent');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`async` import should occur before import of `../parent`", Line: 3, Column: 21, EndLine: 3, EndColumn: 37}}},
			// Upstream case 0:13.
			{Code: "\n        var sibling = require('./sibling');\n        var parent = require('../parent');\n      ", Output: []string{"\n        var parent = require('../parent');\n        var sibling = require('./sibling');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`../parent` import should occur before import of `./sibling`", Line: 3, Column: 22, EndLine: 3, EndColumn: 42}}},
			// Upstream case 0:14.
			{Code: "\n        var index = require('./');\n        var sibling = require('./sibling');\n      ", Output: []string{"\n        var sibling = require('./sibling');\n        var index = require('./');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./sibling` import should occur before import of `./`", Line: 3, Column: 23, EndLine: 3, EndColumn: 43}}},
			// Upstream case 0:15.
			{Code: "\n          var sibling = require('./sibling');\n          var async = require('async');\n          var fs = require('fs');\n        ", Output: []string{"\n          var async = require('async');\n          var sibling = require('./sibling');\n          var fs = require('fs');\n        ", "\n          var fs = require('fs');\n          var async = require('async');\n          var sibling = require('./sibling');\n        "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`async` import should occur before import of `./sibling`", Line: 3, Column: 23, EndLine: 3, EndColumn: 39}, {MessageId: "order", Message: "`fs` import should occur before import of `./sibling`", Line: 4, Column: 20, EndLine: 4, EndColumn: 33}}},
			// Upstream case 0:16.
			{Code: "\n        var index = require('./');\n        var fs = require('fs');\n        var path = require('path');\n        var _ = require('lodash');\n        var foo = require('foo');\n        var bar = require('bar');\n      ", Output: []string{"\n        var fs = require('fs');\n        var path = require('path');\n        var _ = require('lodash');\n        var foo = require('foo');\n        var bar = require('bar');\n        var index = require('./');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./` import should occur after import of `bar`", Line: 2, Column: 21, EndLine: 2, EndColumn: 34}}},
			// Upstream case 0:17.
			{Code: "\n        var fs = require('fs');\n        var index = require('./');\n      ", Options: []any{map[string]any{"groups": []any{"index", "sibling", "parent", "external", "builtin"}}}, Output: []string{"\n        var index = require('./');\n        var fs = require('fs');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./` import should occur before import of `fs`", Line: 3, Column: 21, EndLine: 3, EndColumn: 34}}},
			// Upstream case 0:18.
			{Code: "\n        var foo = require('./foo').bar;\n        var fs = require('fs');\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `./foo`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:19.
			{Code: "\n        var foo = require('./foo').bar.bar.bar;\n        var fs = require('fs');\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `./foo`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:20.
			{Code: "\n        var foo = require('./foo').bar\n          .bar\n          .bar;\n        var fs = require('fs');\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `./foo`", Line: 5, Column: 18, EndLine: 5, EndColumn: 31}}},
			// Upstream case 0:21.
			{Code: "\n        var foo = require('./foo');\n        var fs = require('fs').bar\n          .bar\n          .bar;\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `./foo`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:22.
			{Code: "\n        var fs = require('fs');\n        var index = require('./');\n        var sibling = require('./foo');\n        var path = require('path');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling", "parent", "external"}}}}, Output: []string{"\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`path` import should occur before import of `./foo`", Line: 5, Column: 20, EndLine: 5, EndColumn: 35}}},
			// Upstream case 0:23.
			{Code: "\n        var path = require('path');\n        var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{"index", []any{"sibling", "parent", "external", "internal"}}}}, Output: []string{"\n        var async = require('async');\n        var path = require('path');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`async` import should occur before import of `path`", Line: 3, Column: 21, EndLine: 3, EndColumn: 37}}},
		},
	)
}

func TestOrderUpstreamCore06(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Upstream case 0:24.
			{Code: "\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n        var fs = require('fs');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        import sibling, {foo3} from './foo';\n        var index = require('./');\n      ", Output: []string{"\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n        import sibling, {foo3} from './foo';\n        var fs = require('fs');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var index = require('./');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./foo` import should occur before import of `fs`", Line: 7, Column: 9, EndLine: 7, EndColumn: 45}}},
			// Upstream case 0:25.
			{Code: "\n        var fs = require('fs');\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n      ", Output: []string{"\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n        var fs = require('fs');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur after import of `../foo/bar`", Line: 2, Column: 18, EndLine: 2, EndColumn: 31}}},
			// Upstream case 0:26.
			{Code: "\n          var fs = require('fs');\n          import async, {foo1} from 'async';\n          import bar = require(\"../foo/bar\");\n        ", Output: []string{"\n          import async, {foo1} from 'async';\n          import bar = require(\"../foo/bar\");\n          var fs = require('fs');\n        "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur after import of `../foo/bar`", Line: 2, Column: 20, EndLine: 2, EndColumn: 33}}},
			// Upstream case 0:27.
			{Code: "\n          var async = require('async');\n          var fs = require('fs');\n        ", Output: []string{"\n          var fs = require('fs');\n          var async = require('async');\n        "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 20, EndLine: 3, EndColumn: 33}}},
			// Upstream case 0:28.
			{Code: "\n          import sync = require('sync');\n          import async, {foo1} from 'async';\n\n          import index from './';\n        ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n          import async, {foo1} from 'async';\n          import sync = require('sync');\n\n          import index from './';\n        "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`async` import should occur before import of `sync`", Line: 3, Column: 11, EndLine: 3, EndColumn: 45}}},
			// Upstream case 0:29.
			{Code: "\n          import log = console.log;\n          import blah = require('./blah');", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./blah` import should occur before import of `console.log`", Line: 3, Column: 11, EndLine: 3, EndColumn: 43}}},
			// Upstream case 0:30.
			{Code: "\n        import { Button } from '-/components/Button';\n        import { add } from './helper';\n        import fs from 'fs';\n      ", Options: []any{map[string]any{"groups": []any{"builtin", "external", "unknown", "parent", "sibling", "index"}}}, Output: []string{"\n        import fs from 'fs';\n        import { Button } from '-/components/Button';\n        import { add } from './helper';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `-/components/Button`", Line: 4, Column: 9, EndLine: 4, EndColumn: 29}}},
			// Upstream case 0:31.
			{Code: "\n        import fs from 'fs';\n        import { Button } from '-/components/Button';\n        import { LinkButton } from '-/components/Link';\n        import { add } from './helper';\n      ", Options: []any{map[string]any{"groups": []any{"builtin", "external", "unknown", "parent", "sibling", "index"}, "newlines-between": "always"}}, Output: []string{"\n        import fs from 'fs';\n\n        import { Button } from '-/components/Button';\n        import { LinkButton } from '-/components/Link';\n\n        import { add } from './helper';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 2, Column: 9, EndLine: 2, EndColumn: 29}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 4, Column: 9, EndLine: 4, EndColumn: 56}}},
			// Upstream case 0:32.
			{Code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n\n        var sibling = require('./foo');\n\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling"}, []any{"parent", "external"}}, "newlines-between": "never"}}, Output: []string{"\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be no empty line between import groups", Line: 4, Column: 20, EndLine: 4, EndColumn: 35}, {MessageId: "groupNewline", Message: "There should be no empty line between import groups", Line: 6, Column: 23, EndLine: 6, EndColumn: 39}}},
			// Upstream case 0:33.
			{Code: "\n        var fs = require('fs'); /* comment */\n\n        var index = require('./');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin"}, []any{"index"}}, "newlines-between": "never"}}, Output: []string{"\n        var fs = require('fs'); /* comment */\n        var index = require('./');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be no empty line between import groups", Line: 2, Column: 18, EndLine: 2, EndColumn: 31}}},
			// Upstream case 0:34.
			{Code: "\n        var fs = require('fs'); /* multi-line\n        comment */\n\n        var index = require('./');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin"}, []any{"index"}}, "newlines-between": "never"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be no empty line between import groups", Line: 2, Column: 18, EndLine: 2, EndColumn: 31}}},
			// Upstream case 0:35.
			{Code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling"}, []any{"parent", "external"}}, "newlines-between": "always"}}, Output: []string{"\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n\n        var sibling = require('./foo');\n\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 4, Column: 20, EndLine: 4, EndColumn: 35}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 5, Column: 23, EndLine: 5, EndColumn: 39}}},
			// Upstream case 0:36.
			{Code: "\n        var fs = require('fs');\n\n        var path = require('path');\n        var index = require('./');\n\n        var sibling = require('./foo');\n\n        var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling", "parent", "external"}}, "newlines-between": "always"}}, Output: []string{"\n        var fs = require('fs');\n        var path = require('path');\n        var index = require('./');\n\n        var sibling = require('./foo');\n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "withinGroupNewline", Message: "There should be no empty line within import group", Line: 2, Column: 18, EndLine: 2, EndColumn: 31}, {MessageId: "withinGroupNewline", Message: "There should be no empty line within import group", Line: 7, Column: 23, EndLine: 7, EndColumn: 39}}},
			// Upstream case 0:37.
			{Code: "\n        import path from 'path';\n        import 'loud-rejection';\n\n        import 'something-else';\n        import _ from 'lodash';\n      ", Options: []any{map[string]any{"newlines-between": "never", "warnOnUnassignedImports": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be no empty line between import groups", Line: 2, Column: 9, EndLine: 2, EndColumn: 33}}},
			// Upstream case 0:38.
			{Code: "\n        import path from 'path';\n        import 'loud-rejection';\n\n        import 'something-else';\n        import _ from 'lodash';\n      ", Options: []any{map[string]any{"newlines-between": "never", "warnOnUnassignedImports": true}}, Output: []string{"\n        import path from 'path';\n        import 'loud-rejection';\n        import 'something-else';\n        import _ from 'lodash';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be no empty line between import groups", Line: 3, Column: 9, EndLine: 3, EndColumn: 33}}},
			// Upstream case 0:39.
			{Code: "\n        import path from 'path';\n        export const abc = 123;\n\n        import 'something-else';\n        import _ from 'lodash';\n      ", Options: []any{map[string]any{"newlines-between": "never"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be no empty line between import groups", Line: 2, Column: 9, EndLine: 2, EndColumn: 33}}},
			// Upstream case 0:40.
			{Code: "\n        import path from 'path';\n        import 'loud-rejection';\n        import 'something-else';\n        import _ from 'lodash';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}, Output: []string{"\n        import path from 'path';\n\n        import 'loud-rejection';\n        import 'something-else';\n        import _ from 'lodash';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 2, Column: 9, EndLine: 2, EndColumn: 33}}},
			// Upstream case 0:41.
			{Code: "\n        import path from 'path'; // comment\n        import _ from 'lodash';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}, Output: []string{"\n        import path from 'path'; // comment\n\n        import _ from 'lodash';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 2, Column: 9, EndLine: 2, EndColumn: 33}}},
			// Upstream case 0:42.
			{Code: "\n        import path from 'path'; /* comment */ /* comment */\n        import _ from 'lodash';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}, Output: []string{"\n        import path from 'path'; /* comment */ /* comment */\n\n        import _ from 'lodash';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 2, Column: 9, EndLine: 2, EndColumn: 33}}},
			// Upstream case 0:43.
			{Code: "\n        import path from 'path'; /* 1\n        2 */\n        import _ from 'lodash';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}, Output: []string{"\n        import path from 'path';\n /* 1\n        2 */\n        import _ from 'lodash';\n      ", "\n        import path from 'path';\n\n /* 1\n        2 */\n        import _ from 'lodash';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 2, Column: 9, EndLine: 2, EndColumn: 33}}},
		},
	)
}

func TestOrderUpstreamCore07(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Upstream case 0:44.
			{Code: "\n        const local = require('./local');\n\n        fn_call();\n\n        const global1 = require('global1');\n        const global2 = require('global2');\n\n        fn_call();\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./local` import should occur after import of `global2`", Line: 2, Column: 23, EndLine: 2, EndColumn: 41}}},
			// Upstream case 0:45.
			{Code: "\n        const local = require('./local');\n        fn_call();\n        const global1 = require('global1');\n        const global2 = require('global2');\n\n        fn_call();\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./local` import should occur after import of `global2`", Line: 2, Column: 23, EndLine: 2, EndColumn: 41}}},
			// Upstream case 0:46.
			{Code: "\n        const local1 = require('./local1');\n        const local2 = require('./local2');\n        const local3 = require('./local3');\n        const local4 = require('./local4');\n        fn_call();\n        const global1 = require('global1');\n        const global2 = require('global2');\n        const global3 = require('global3');\n        const global4 = require('global4');\n        const global5 = require('global5');\n        fn_call();\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./local1` import should occur after import of `global5`", Line: 2, Column: 24, EndLine: 2, EndColumn: 43}, {MessageId: "order", Message: "`./local2` import should occur after import of `global5`", Line: 3, Column: 24, EndLine: 3, EndColumn: 43}, {MessageId: "order", Message: "`./local3` import should occur after import of `global5`", Line: 4, Column: 24, EndLine: 4, EndColumn: 43}, {MessageId: "order", Message: "`./local4` import should occur after import of `global5`", Line: 5, Column: 24, EndLine: 5, EndColumn: 43}}},
			// Upstream case 0:47.
			{Code: "\n        const local = require('./local');\n        const global1 = require('global1');\n        const global2 = require('global2');\n        fn_call();\n        const global3 = require('global3');\n\n        fn_call();\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./local` import should occur after import of `global3`", Line: 2, Column: 23, EndLine: 2, EndColumn: 41}}},
			// Upstream case 0:48.
			{Code: "\n        const local1 = require('./local1');\n        const global1 = require('global1');\n        const global2 = require('global2');\n        fn_call();\n        const local2 = require('./local2');\n        const global3 = require('global3');\n        const global4 = require('global4');\n\n        fn_call();\n      ", Output: []string{"\n        const local1 = require('./local1');\n        const global1 = require('global1');\n        const global2 = require('global2');\n        fn_call();\n        const global3 = require('global3');\n        const global4 = require('global4');\n        const local2 = require('./local2');\n\n        fn_call();\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./local1` import should occur after import of `global4`", Line: 2, Column: 24, EndLine: 2, EndColumn: 43}, {MessageId: "order", Message: "`./local2` import should occur after import of `global4`", Line: 6, Column: 24, EndLine: 6, EndColumn: 43}}},
			// Upstream case 0:49.
			{Code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { add } from './helper';\n        import { Input } from '~/components/Input';\n        ", Options: []any{map[string]any{"pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external", "position": "after"}}}}, Output: []string{"\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { Input } from '~/components/Input';\n        import { add } from './helper';\n        "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`~/components/Input` import should occur before import of `./helper`", Line: 5, Column: 9, EndLine: 5, EndColumn: 52}}},
			// Upstream case 0:50.
			{Code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { add } from './helper';\n        import { Input } from '~/components/Input';\n        import async from 'async';\n        ", Options: []any{map[string]any{"pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external"}}}}, Output: []string{"\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { Input } from '~/components/Input';\n        import async from 'async';\n        import { add } from './helper';\n        "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./helper` import should occur after import of `async`", Line: 4, Column: 9, EndLine: 4, EndColumn: 40}}},
			// Upstream case 0:51.
			{Code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { add } from './helper';\n        import { Input } from '~/components/Input';\n        ", Options: []any{map[string]any{"pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external", "position": "before"}}}}, Output: []string{"\n        import fs from 'fs';\n        import { Input } from '~/components/Input';\n        import _ from 'lodash';\n        import { add } from './helper';\n        "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`~/components/Input` import should occur before import of `lodash`", Line: 5, Column: 9, EndLine: 5, EndColumn: 52}}},
			// Upstream case 0:52.
			{Code: "\n        import fs from 'fs';\n        import { Import } from '$/components/Import';\n        import _ from 'lodash';\n        import { Output } from '~/components/Output';\n        import { Input } from '#/components/Input';\n        import { add } from './helper';\n        import { Export } from '-/components/Export';\n        ", Options: []any{map[string]any{"pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external", "position": "after"}, map[string]any{"pattern": "#/**", "group": "external", "position": "after"}, map[string]any{"pattern": "-/**", "group": "external", "position": "before"}, map[string]any{"pattern": "$/**", "group": "external", "position": "before"}}}}, Output: []string{"\n        import fs from 'fs';\n        import { Export } from '-/components/Export';\n        import { Import } from '$/components/Import';\n        import _ from 'lodash';\n        import { Output } from '~/components/Output';\n        import { Input } from '#/components/Input';\n        import { add } from './helper';\n        "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`-/components/Export` import should occur before import of `$/components/Import`", Line: 8, Column: 9, EndLine: 8, EndColumn: 54}}},
			// Upstream case 0:53.
			{Code: "\n        import fs from 'fs';\n        import { Export } from '-/components/Export';\n        import { Import } from '$/components/Import';\n        import _ from 'lodash';\n        import { Input } from '#/components/Input';\n        import { add } from './helper';\n        import { Output } from '~/components/Output';\n        ", Options: []any{map[string]any{"pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external", "position": "after"}, map[string]any{"pattern": "#/**", "group": "external", "position": "after"}, map[string]any{"pattern": "-/**", "group": "external", "position": "before"}, map[string]any{"pattern": "$/**", "group": "external", "position": "before"}}}}, Output: []string{"\n        import fs from 'fs';\n        import { Export } from '-/components/Export';\n        import { Import } from '$/components/Import';\n        import _ from 'lodash';\n        import { Output } from '~/components/Output';\n        import { Input } from '#/components/Input';\n        import { add } from './helper';\n        "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`~/components/Output` import should occur before import of `#/components/Input`", Line: 8, Column: 9, EndLine: 8, EndColumn: 54}}},
			// Upstream case 0:54.
			{Code: "\n        import path from 'path';\n        import { namespace } from '@namespace';\n        import { a } from 'a';\n        import { b } from 'b';\n        import { c } from 'c';\n        import { d } from 'd';\n        import { e } from 'e';\n        import { f } from 'f';\n        import { g } from 'g';\n        import { h } from 'h';\n        import { i } from 'i';\n        import { j } from 'j';\n        import { k } from 'k';", Options: []any{map[string]any{"groups": []any{"builtin", "external", "internal"}, "pathGroups": []any{map[string]any{"pattern": "@namespace", "group": "external", "position": "after"}, map[string]any{"pattern": "a", "group": "internal", "position": "before"}, map[string]any{"pattern": "b", "group": "internal", "position": "before"}, map[string]any{"pattern": "c", "group": "internal", "position": "before"}, map[string]any{"pattern": "d", "group": "internal", "position": "before"}, map[string]any{"pattern": "e", "group": "internal", "position": "before"}, map[string]any{"pattern": "f", "group": "internal", "position": "before"}, map[string]any{"pattern": "g", "group": "internal", "position": "before"}, map[string]any{"pattern": "h", "group": "internal", "position": "before"}, map[string]any{"pattern": "i", "group": "internal", "position": "before"}}, "newlines-between": "always", "pathGroupsExcludedImportTypes": []any{"builtin"}}}, Settings: map[string]any{"import/internal-regex": "^(a|b|c|d|e|f|g|h|i|j|k)(\\/|$)"}, Output: []string{"\n        import path from 'path';\n\n        import { namespace } from '@namespace';\n\n        import { a } from 'a';\n\n        import { b } from 'b';\n\n        import { c } from 'c';\n\n        import { d } from 'd';\n\n        import { e } from 'e';\n\n        import { f } from 'f';\n\n        import { g } from 'g';\n\n        import { h } from 'h';\n\n        import { i } from 'i';\n\n        import { j } from 'j';\n        import { k } from 'k';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 2, Column: 9, EndLine: 2, EndColumn: 33}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 3, Column: 9, EndLine: 3, EndColumn: 48}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 4, Column: 9, EndLine: 4, EndColumn: 31}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 5, Column: 9, EndLine: 5, EndColumn: 31}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 6, Column: 9, EndLine: 6, EndColumn: 31}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 7, Column: 9, EndLine: 7, EndColumn: 31}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 8, Column: 9, EndLine: 8, EndColumn: 31}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 9, Column: 9, EndLine: 9, EndColumn: 31}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 10, Column: 9, EndLine: 10, EndColumn: 31}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 11, Column: 9, EndLine: 11, EndColumn: 31}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 12, Column: 9, EndLine: 12, EndColumn: 31}}},
			// Upstream case 0:55.
			{Code: "\n        import external from 'external';\n        import a from '@namespace/a';\n        import b from '@namespace/b';\n        import { parent } from '../../parent';\n        import local from './local';\n        import './side-effect';", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc", "caseInsensitive": true}, "groups": []any{"type", "builtin", "external", "internal", "parent", "sibling", "index", "object"}, "newlines-between": "always", "pathGroups": []any{map[string]any{"pattern": "@namespace", "group": "external", "position": "after"}, map[string]any{"pattern": "@namespace/**", "group": "external", "position": "after"}}, "pathGroupsExcludedImportTypes": []any{}}}, Output: []string{"\n        import external from 'external';\n\n        import a from '@namespace/a';\n        import b from '@namespace/b';\n\n        import { parent } from '../../parent';\n\n        import local from './local';\n        import './side-effect';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 2, Column: 9, EndLine: 2, EndColumn: 41}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 4, Column: 9, EndLine: 4, EndColumn: 38}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 5, Column: 9, EndLine: 5, EndColumn: 47}}},
			// Upstream case 0:56.
			{Code: "\n        var async = require('async');\n        fn_call();\n        var fs = require('fs');\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 4, Column: 18, EndLine: 4, EndColumn: 31}}},
			// Upstream case 0:57.
			{Code: "\n        const env = require('./config');\n\n        Object.keys(env);\n\n        const http = require('http');\n        const express = require('express');\n\n        http.createServer(express());\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./config` import should occur after import of `express`", Line: 2, Column: 21, EndLine: 2, EndColumn: 40}}},
			// Upstream case 0:58.
			{Code: "\n        var async = require('async');\n        var a = require('./value.js')(a);\n        var fs = require('fs');\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 4, Column: 18, EndLine: 4, EndColumn: 31}}},
			// Upstream case 0:59.
			{Code: "\n        var async = require('async');\n        var fs = require('fs')(a);\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:60.
			{Code: "\n        var async = require('async')(a);\n        var fs = require('fs');\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:61.
			{Code: "\n        var async = require('async');\n        require('./aa');\n        var fs = require('fs');\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 4, Column: 18, EndLine: 4, EndColumn: 31}}},
			// Upstream case 0:62.
			{Code: "\n        import async from 'async';\n        fn_call();\n        import fs from 'fs';\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 4, Column: 9, EndLine: 4, EndColumn: 29}}},
			// Upstream case 0:63.
			{Code: "\n        import async from 'async';\n        var a = 1;\n        import fs from 'fs';\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 4, Column: 9, EndLine: 4, EndColumn: 29}}},
		},
	)
}

func TestOrderUpstreamCore08(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Upstream case 0:64.
			{Code: "\n        import async from 'async';\n        var a = require('./value.js')(a);\n        import fs from 'fs';\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 4, Column: 9, EndLine: 4, EndColumn: 29}}},
			// Upstream case 0:65.
			{Code: "\n        import async from 'async';\n        import './aa';\n        import fs from 'fs';\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 4, Column: 9, EndLine: 4, EndColumn: 29}}},
			// Upstream case 0:66.
			{Code: "\n        import b from 'bar';\n        import c from 'Bar';\n        import a from 'foo';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n        import c from 'Bar';\n        import b from 'bar';\n        import a from 'foo';\n\n        import index from './';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`Bar` import should occur before import of `bar`", Line: 3, Column: 9, EndLine: 3, EndColumn: 29}}},
			// Upstream case 0:67.
			{Code: "\n        import a from 'foo';\n        import c from 'Bar';\n        import b from 'bar';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "desc"}}}, Output: []string{"\n        import a from 'foo';\n        import b from 'bar';\n        import c from 'Bar';\n\n        import index from './';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`bar` import should occur before import of `Bar`", Line: 4, Column: 9, EndLine: 4, EndColumn: 29}}},
			// Upstream case 0:68.
			{Code: "\n        import a from \"foo\";\n        import b from \"foo-bar\";\n        import c from \"foo/bar\";\n        import d from \"foo/barfoo\";\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n        import a from \"foo\";\n        import c from \"foo/bar\";\n        import d from \"foo/barfoo\";\n        import b from \"foo-bar\";\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`foo-bar` import should occur after import of `foo/barfoo`", Line: 3, Column: 9, EndLine: 3, EndColumn: 33}}},
			// Upstream case 0:69.
			{Code: "\n        import b from 'foo';\n        import a from 'Bar';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "asc", "caseInsensitive": true}}}, Output: []string{"\n        import a from 'Bar';\n        import b from 'foo';\n\n        import index from './';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`Bar` import should occur before import of `foo`", Line: 3, Column: 9, EndLine: 3, EndColumn: 29}}},
			// Upstream case 0:70.
			{Code: "\n        import a from 'Bar';\n        import b from 'foo';\n\n        import index from './';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "desc", "caseInsensitive": true}}}, Output: []string{"\n        import b from 'foo';\n        import a from 'Bar';\n\n        import index from './';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`foo` import should occur before import of `Bar`", Line: 3, Column: 9, EndLine: 3, EndColumn: 29}}},
			// Upstream case 0:71.
			{Code: "\n        const b = require('./b').get();\n        const a = require('./a');\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n        const a = require('./a');\n        const b = require('./b').get();\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./a` import should occur before import of `./b`", Line: 3, Column: 19, EndLine: 3, EndColumn: 33}}},
			// Upstream case 0:72.
			{Code: "\n        import a from '../a';\n        import p from '..';\n      ", Options: []any{map[string]any{"groups": []any{"external", "index"}, "alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n        import p from '..';\n        import a from '../a';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`..` import should occur before import of `../a`", Line: 3, Column: 9, EndLine: 3, EndColumn: 28}}},
			// Upstream case 0:73.
			{Code: "\n        import A from 'a';\n        import C from './c';\n      ", Options: []any{map[string]any{"newlines-between": "always", "distinctGroup": false, "pathGroupsExcludedImportTypes": []any{}}}, Output: []string{"\n        import A from 'a';\n\n        import C from './c';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 2, Column: 9, EndLine: 2, EndColumn: 27}}},
			// Upstream case 0:74.
			{Code: "\n        import A from 'a';\n\n        import C from 'c';\n      ", Options: []any{map[string]any{"newlines-between": "always", "distinctGroup": false, "pathGroupsExcludedImportTypes": []any{}, "pathGroups": []any{map[string]any{"pattern": "a", "group": "external", "position": "before"}, map[string]any{"pattern": "c", "group": "external", "position": "after"}}}}, Output: []string{"\n        import A from 'a';\n        import C from 'c';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "withinGroupNewline", Message: "There should be no empty line within import group", Line: 2, Column: 9, EndLine: 2, EndColumn: 27}}},
			// Upstream case 0:75.
			{Code: "\n        var { B, A: R } = require(\"./Z\");\n        import { O as G, D } from \"./Z\";\n        import { K, L, J } from \"./Z\";\n        export { Z, X, Y } from \"./Z\";\n      ", Options: []any{map[string]any{"named": true, "alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n        var { A: R, B } = require(\"./Z\");\n        import { D, O as G } from \"./Z\";\n        import { J, K, L } from \"./Z\";\n        export { X, Y, Z } from \"./Z\";\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`A` import should occur before import of `B`", Line: 2, Column: 18, EndLine: 2, EndColumn: 22}, {MessageId: "order", Message: "`D` import should occur before import of `O`", Line: 3, Column: 26, EndLine: 3, EndColumn: 27}, {MessageId: "order", Message: "`J` import should occur before import of `K`", Line: 4, Column: 24, EndLine: 4, EndColumn: 25}, {MessageId: "order", Message: "`Z` export should occur after export of `Y`", Line: 5, Column: 18, EndLine: 5, EndColumn: 19}}},
			// Upstream case 0:76.
			{Code: "\n        import { D, C } from \"./Z\";\n        var { B, A } = require(\"./Z\");\n        export { B, A };\n      ", Options: []any{map[string]any{"named": map[string]any{"require": false, "import": true, "export": true}, "alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n        import { C, D } from \"./Z\";\n        var { B, A } = require(\"./Z\");\n        export { A, B };\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`C` import should occur before import of `D`", Line: 2, Column: 21, EndLine: 2, EndColumn: 22}, {MessageId: "order", Message: "`A` export should occur before export of `B`", Line: 4, Column: 21, EndLine: 4, EndColumn: 22}}},
			// Upstream case 0:77.
			{Code: "\n        import { A as B, A as C, A } from \"./Z\";\n        export { A, A as D, A as B, A as C } from \"./Z\";\n        const { a: b, a: c, a } = require(\"./Z\");\n      ", Options: []any{map[string]any{"named": true, "alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n        import { A, A as B, A as C } from \"./Z\";\n        export { A, A as B, A as C, A as D } from \"./Z\";\n        const { a, a: b, a: c } = require(\"./Z\");\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`A` import should occur before import of `A as B`", Line: 2, Column: 34, EndLine: 2, EndColumn: 35}, {MessageId: "order", Message: "`A as D` export should occur after export of `A as C`", Line: 3, Column: 21, EndLine: 3, EndColumn: 27}, {MessageId: "order", Message: "`a` import should occur before import of `a as b`", Line: 4, Column: 29, EndLine: 4, EndColumn: 30}}},
			// Upstream case 0:78.
			{Code: "\n        import { A, B, C } from \"./Z\";\n        exports = { B, C, A };\n        module.exports = { c: C, a: A, b: B };\n      ", Options: []any{map[string]any{"named": map[string]any{"cjsExports": true}, "alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n        import { A, B, C } from \"./Z\";\n        exports = { A, B, C };\n        module.exports = { a: A, b: B, c: C };\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`A` export should occur before export of `B`", Line: 3, Column: 27, EndLine: 3, EndColumn: 28}, {MessageId: "order", Message: "`c` export should occur after export of `b`", Line: 4, Column: 28, EndLine: 4, EndColumn: 32}}},
			// Upstream case 0:79.
			{Code: "\n        exports.B = { };\n        module.exports.A = { };\n        module.exports.C = { };\n      ", Options: []any{map[string]any{"named": map[string]any{"cjsExports": true}, "alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n        module.exports.A = { };\n        exports.B = { };\n        module.exports.C = { };\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`A` export should occur before export of `B`", Line: 3, Column: 9, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:80.
			{Code: "\n        exports.A.C = { };\n        module.exports.A.A = { };\n        exports.A.B = { };\n      ", Options: []any{map[string]any{"named": map[string]any{"cjsExports": true}, "alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n        module.exports.A.A = { };\n        exports.A.B = { };\n        exports.A.C = { };\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`A.C` export should occur after export of `A.B`", Line: 2, Column: 9, EndLine: 2, EndColumn: 26}}},
			// Upstream case 0:81.
			{Code: "\n        const {\n          F: O,\n          O: B,\n          /* Hello World */\n          A: R\n        } = require(\"./Z\");\n        import {\n          Y,\n          X,\n        } from \"./Z\";\n        export {\n          Z, A,\n          B\n        } from \"./Z\";\n        module.exports = {\n          a: A, o: O,\n          b: B\n        };\n      ", Options: []any{map[string]any{"named": map[string]any{"enabled": true}, "alphabetize": map[string]any{"order": "asc"}}}, Output: []string{"\n        const {\n          /* Hello World */\n          A: R,\n          F: O,\n          O: B\n        } = require(\"./Z\");\n        import {\n          X,\n          Y,\n        } from \"./Z\";\n        export { A,\n          B,\n          Z\n        } from \"./Z\";\n        module.exports = {\n          a: A,\n          b: B, o: O\n        };\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`A` import should occur before import of `F`", Line: 6, Column: 11, EndLine: 6, EndColumn: 15}, {MessageId: "order", Message: "`X` import should occur before import of `Y`", Line: 10, Column: 11, EndLine: 10, EndColumn: 12}, {MessageId: "order", Message: "`Z` export should occur after export of `B`", Line: 13, Column: 11, EndLine: 13, EndColumn: 12}, {MessageId: "order", Message: "`b` export should occur before export of `o`", Line: 18, Column: 11, EndLine: 18, EndColumn: 15}}},
			// Upstream case 0:82.
			{Code: "\n          const { cello } = require('./cello');\n          import { int } from './int';\n          const blah = require('./blah');\n          import { hello } from './hello';\n        ", Output: []string{"\n          import { int } from './int';\n          const { cello } = require('./cello');\n          const blah = require('./blah');\n          import { hello } from './hello';\n        ", "\n          import { int } from './int';\n          import { hello } from './hello';\n          const { cello } = require('./cello');\n          const blah = require('./blah');\n        "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./int` import should occur before import of `./cello`", Line: 3, Column: 11, EndLine: 3, EndColumn: 39}, {MessageId: "order", Message: "`./hello` import should occur before import of `./cello`", Line: 5, Column: 11, EndLine: 5, EndColumn: 43}}},
			// Upstream case 0:83.
			{Code: "\n        var fs = require('fs');\n        var path = require('path');\n        var { util1, util2, util3 } = require('util');\n        var async = require('async');\n        var relParent1 = require('../foo');\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n        var relParent3 = require('../bar');\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      ", Options: []any{map[string]any{"newlines-between": "always-and-inside-groups", "consolidateIslands": "inside-groups"}}, Output: []string{"\n        var fs = require('fs');\n        var path = require('path');\n        var { util1, util2, util3 } = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 4, Column: 39, EndLine: 4, EndColumn: 54}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 5, Column: 21, EndLine: 5, EndColumn: 37}, {MessageId: "consolidate", Message: "There should be at least one empty line between this import and the multi-line import that follows it", Line: 6, Column: 26, EndLine: 6, EndColumn: 43}, {MessageId: "consolidate", Message: "There should be at least one empty line between this multi-line import and the import that follows it", Line: 12, Column: 13, EndLine: 12, EndColumn: 27}, {MessageId: "groupNewline", Message: "There should be at least one empty line between import groups", Line: 13, Column: 26, EndLine: 13, EndColumn: 43}, {MessageId: "consolidate", Message: "There should be at least one empty line between this multi-line import and the import that follows it", Line: 15, Column: 34, EndLine: 15, EndColumn: 50}}},
		},
	)
}

func TestOrderUpstreamCore09(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Upstream case 0:84.
			{Code: "\n        var fs = require('fs');\n\n        var path = require('path');\n\n        var { util1, util2, util3 } = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n\n        var sibling2 = require('./bar');\n\n        var sibling3 = require('./foobar');\n      ", Options: []any{map[string]any{"newlines-between": "always-and-inside-groups", "consolidateIslands": "inside-groups"}}, Output: []string{"\n        var fs = require('fs');\n        var path = require('path');\n        var { util1, util2, util3 } = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consolidate", Message: "There should be no empty lines between this single-line import and the single-line import that follows it", Line: 2, Column: 18, EndLine: 2, EndColumn: 31}, {MessageId: "consolidate", Message: "There should be no empty lines between this single-line import and the single-line import that follows it", Line: 4, Column: 20, EndLine: 4, EndColumn: 35}, {MessageId: "consolidate", Message: "There should be no empty lines between this single-line import and the single-line import that follows it", Line: 24, Column: 24, EndLine: 24, EndColumn: 40}}},
		},
	)
}
