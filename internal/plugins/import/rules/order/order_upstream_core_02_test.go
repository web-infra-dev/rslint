// TestOrderUpstreamCore02 migrates the upstream core cases from import-js/eslint-plugin-import at 1a39fb8672c83f8781c3a97f02eee00a7a2f0253.
// It preserves upstream messages, source positions, options, and iterative autofix outputs.
// Framework-only parser/resolver settings use rslint native equivalents; tsgo lock-ins live in order_extras_*_test.go.
package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderUpstreamCore02(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// Upstream case 0:20.
			{Code: "\n          export import CreateSomething = _CreateSomething;\n        "},
			// Upstream case 0:21.
			{Code: "\n        import fs from 'fs';\n        import { Input } from '-/components/Input';\n        import { Button } from '-/components/Button';\n        import { add } from './helper';", Options: []any{map[string]any{"groups": []any{"builtin", "external", "unknown", "parent", "sibling", "index"}}}},
			// Upstream case 0:22.
			{Code: "\n        import { Input } from '-/components/Input';\n        import { Button } from '-/components/Button';\n        import fs from 'fs';\n        import { add } from './helper';", Options: []any{map[string]any{"groups": []any{"unknown", "builtin", "external", "parent", "sibling", "index"}}}},
			// Upstream case 0:23.
			{Code: "\n        import fs from 'fs';\n\n        import { Input } from '-/components/Input';\n        import { Button } from '-/components/Button';\n\n        import p from '..';\n        import q from '../';\n\n        import { add } from './helper';\n\n        import i from '.';\n        import j from './';", Options: []any{map[string]any{"newlines-between": "always", "groups": []any{"builtin", "external", "unknown", "parent", "sibling", "index"}}}},
			// Upstream case 0:24.
			{Code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { Input } from '~/components/Input';\n        import { Button } from '#/components/Button';\n        import { add } from './helper';", Options: []any{map[string]any{"pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external", "position": "after"}, map[string]any{"pattern": "#/**", "group": "external", "position": "after"}}}}},
			// Upstream case 0:25.
			{Code: "\n        import fs from 'fs';\n        import { Input } from '~/components/Input';\n        import async from 'async';\n        import { Button } from '#/components/Button';\n        import _ from 'lodash';\n        import { add } from './helper';", Options: []any{map[string]any{"pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external"}, map[string]any{"pattern": "#/**", "group": "external"}}}}},
			// Upstream case 0:26.
			{Code: "\n        import fs from 'fs';\n\n        import { Input } from '~/components/Input';\n\n        import { Button } from '#/components/Button';\n\n        import _ from 'lodash';\n\n        import { add } from './helper';", Options: []any{map[string]any{"newlines-between": "always", "pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external", "position": "before"}, map[string]any{"pattern": "#/**", "group": "external", "position": "before"}}}}},
			// Upstream case 0:27.
			{Code: "\n        import fs from 'fs';\n\n        import _ from 'lodash';\n\n        import { Input } from '~/components/Input';\n\n        import { Button } from '!/components/Button';\n\n        import { add } from './helper';", Options: []any{map[string]any{"newlines-between": "always", "pathGroups": []any{map[string]any{"pattern": "~/**", "group": "external", "position": "after"}, map[string]any{"pattern": "!/**", "patternOptions": map[string]any{"nonegate": true}, "group": "external", "position": "after"}}}}},
			// Upstream case 0:28.
			{Code: "\n        import fs from 'fs';\n\n        import { Input } from '@app/components/Input';\n\n        import { Button } from '@app2/components/Button';\n\n        import _ from 'lodash';\n\n        import { add } from './helper';", Options: []any{map[string]any{"newlines-between": "always", "pathGroupsExcludedImportTypes": []any{"builtin"}, "pathGroups": []any{map[string]any{"pattern": "@app/**", "group": "external", "position": "before"}, map[string]any{"pattern": "@app2/**", "group": "external", "position": "before"}}}}},
			// Upstream case 0:29.
			{Code: "\n        import fs from 'fs';\n        import external from 'external';\n        import externalTooPlease from './make-me-external';\n\n        import sibling from './sibling';", Options: []any{map[string]any{"newlines-between": "always", "pathGroupsExcludedImportTypes": []any{}, "pathGroups": []any{map[string]any{"pattern": "./make-me-external", "group": "external"}}, "groups": []any{[]any{"builtin", "external"}, "internal", "parent", "sibling", "index"}}}},
			// Upstream case 0:30.
			{Code: "\n        import _ from 'lodash';\n        import m from '@test-scope/some-module';\n\n        import bar from './bar';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}, Settings: map[string]any{"import/external-module-folders": []any{"node_modules", "symlinked-module"}}},
			// Upstream case 0:31.
			{Code: "\n        import _ from 'lodash';\n        import m from '@test-scope/some-module';\n\n        import bar from './bar';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}, Settings: map[string]any{"import/external-module-folders": []any{"node_modules", "symlinked-module"}}},
			// Upstream case 0:32.
			{Code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n\n\n\n        var sibling = require('./foo');\n\n\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling"}, []any{"parent", "external"}}, "newlines-between": "always"}}},
			// Upstream case 0:33.
			{Code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling"}, []any{"parent", "external"}}, "newlines-between": "never"}}},
			// Upstream case 0:34.
			{Code: "\n      var fs = require('fs');\n\n      var index = require('./');\n      var path = require('path');\n      var sibling = require('./foo');\n\n\n      var relParent1 = require('../foo');\n\n      var relParent3 = require('../');\n      var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling"}, []any{"parent", "external"}}, "newlines-between": "ignore"}}},
			// Upstream case 0:35.
			{Code: "\n      var fs = require('fs');\n\n      var index = require('./');\n      var path = require('path');\n      var sibling = require('./foo');\n\n\n      var relParent1 = require('../foo');\n\n      var relParent3 = require('../');\n\n      var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling"}, []any{"parent", "external"}}}}},
			// Upstream case 0:36.
			{Code: "\n        import path from 'path';\n\n        import {\n            I,\n            Want,\n            Couple,\n            Imports,\n            Here\n        } from 'bar';\n        import external from 'external'\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
			// Upstream case 0:37.
			{Code: "\n        import path from 'path';\n        import net\n          from 'net';\n\n        import external from 'external'\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
			// Upstream case 0:38.
			{Code: "\n        import foo\n          from '../../../../this/will/be/very/long/path/and/therefore/this/import/has/to/be/in/two/lines';\n\n        import bar\n          from './sibling';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
			// Upstream case 0:39.
			{Code: "\n        import path from 'path';\n\n        import 'loud-rejection';\n        import 'something-else';\n\n        import _ from 'lodash';\n      ", Options: []any{map[string]any{"newlines-between": "always"}}},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
