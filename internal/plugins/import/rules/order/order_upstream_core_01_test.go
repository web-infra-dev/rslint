// TestOrderUpstreamCore01 migrates the upstream core cases from import-js/eslint-plugin-import at 1a39fb8672c83f8781c3a97f02eee00a7a2f0253.
// It preserves upstream messages, source positions, options, and iterative autofix outputs.
// Framework-only parser/resolver settings use rslint native equivalents; tsgo lock-ins live in order_extras_*_test.go.
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
			{Code: "\n        import { fromAbi } from '../../.checkpoint/models';\n        import { handleVotingPowerValidationMetadata } from '../common/ipfs';\n        import ExecutionStrategyAbi from './abis/executionStrategy.json';\n        import L1AvatarExecutionStrategyAbi from './abis/l1/L1AvatarExectionStrategy';\n        import { FullConfig } from './config';\n      ", Options: []any{map[string]any{"groups": []any{"builtin", "external", "internal", "type"}, "alphabetize": map[string]any{"order": "asc", "caseInsensitive": true}}}},
			// Upstream case 0:17.
			{Code: "\n        var index = require('./');\n        var path = require('path');\n      ", Options: []any{map[string]any{"groups": []any{"index", []any{"sibling", "parent", "external"}}}}},
			// Upstream case 0:18.
			{Code: "\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n        import sibling, {foo3} from './foo';\n        var fs = require('fs');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var index = require('./');\n      "},
			// Upstream case 0:19.
			{Code: "\n          import async, {foo1} from 'async';\n          import relParent2, {foo2} from '../foo/bar';\n          import sibling, {foo3} from './foo';\n          var fs = require('fs');\n          var util = require(\"util\");\n          var relParent1 = require('../foo');\n          var relParent3 = require('../');\n          var index = require('./');\n        "},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
