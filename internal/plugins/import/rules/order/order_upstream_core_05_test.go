// TestOrderUpstreamCore05 migrates the upstream core cases from import-js/eslint-plugin-import at 1a39fb8672c83f8781c3a97f02eee00a7a2f0253.
// It preserves upstream messages, source positions, options, and iterative autofix outputs.
// Framework-only parser/resolver settings use rslint native equivalents; tsgo lock-ins live in order_extras_*_test.go.
package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderUpstreamCore05(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Upstream case 0:5.
			{Code: "\n        /* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n      ", Output: []string{"\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n        /* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 34, EndLine: 3, EndColumn: 47}}},
			// Upstream case 0:6.
			{Code: "/* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\r\n/* comment3 */  var fs = require('fs'); /* comment4 */\r\n", Output: []string{"/* comment3 */  var fs = require('fs'); /* comment4 */\r\n/* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\r\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 2, Column: 26, EndLine: 2, EndColumn: 39}}},
			// Upstream case 0:7.
			{Code: "\n        /* multi-line1\n          comment1 */  var async = require('async'); /* multi-line2\n          comment2 */  var fs = require('fs'); /* multi-line3\n          comment3 */\n      ", Output: []string{"\n        /* multi-line1\n          comment1 */  var fs = require('fs'); \n  var async = require('async'); /* multi-line2\n          comment2 *//* multi-line3\n          comment3 */\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 4, Column: 33, EndLine: 4, EndColumn: 46}}},
			// Upstream case 0:8.
			{Code: "\n        var {b} = require('async');\n        var {a} = require('fs');\n      ", Output: []string{"\n        var {a} = require('fs');\n        var {b} = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 19, EndLine: 3, EndColumn: 32}}},
			// Upstream case 0:9.
			{Code: "\n        var async = require('async');\n        var fs =\n          require('fs');\n      ", Output: []string{"\n        var fs =\n          require('fs');\n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 4, Column: 11, EndLine: 4, EndColumn: 24}}},
			// Upstream case 0:10.
			{Code: "\n        var async = require('async');\n        var fs = require('fs');", Output: []string{"\n        var fs = require('fs');\n        var async = require('async');\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:11.
			{Code: "\n        import async from 'async';\n        import fs from 'fs';\n      ", Output: []string{"\n        import fs from 'fs';\n        import async from 'async';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 9, EndLine: 3, EndColumn: 29}}},
			// Upstream case 0:12.
			{Code: "\n        var async = require('async');\n        import fs from 'fs';\n      ", Output: []string{"\n        import fs from 'fs';\n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 9, EndLine: 3, EndColumn: 29}}},
			// Upstream case 0:13.
			{Code: "\n        var parent = require('../parent');\n        var async = require('async');\n      ", Output: []string{"\n        var async = require('async');\n        var parent = require('../parent');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`async` import should occur before import of `../parent`", Line: 3, Column: 21, EndLine: 3, EndColumn: 37}}},
			// Upstream case 0:14.
			{Code: "\n        var sibling = require('./sibling');\n        var parent = require('../parent');\n      ", Output: []string{"\n        var parent = require('../parent');\n        var sibling = require('./sibling');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`../parent` import should occur before import of `./sibling`", Line: 3, Column: 22, EndLine: 3, EndColumn: 42}}},
			// Upstream case 0:15.
			{Code: "\n        var index = require('./');\n        var sibling = require('./sibling');\n      ", Output: []string{"\n        var sibling = require('./sibling');\n        var index = require('./');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./sibling` import should occur before import of `./`", Line: 3, Column: 23, EndLine: 3, EndColumn: 43}}},
			// Upstream case 0:16.
			{Code: "\n          var sibling = require('./sibling');\n          var async = require('async');\n          var fs = require('fs');\n        ", Output: []string{"\n          var async = require('async');\n          var sibling = require('./sibling');\n          var fs = require('fs');\n        ", "\n          var fs = require('fs');\n          var async = require('async');\n          var sibling = require('./sibling');\n        "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`async` import should occur before import of `./sibling`", Line: 3, Column: 23, EndLine: 3, EndColumn: 39}, {MessageId: "order", Message: "`fs` import should occur before import of `./sibling`", Line: 4, Column: 20, EndLine: 4, EndColumn: 33}}},
			// Upstream case 0:17.
			{Code: "\n        var index = require('./');\n        var fs = require('fs');\n        var path = require('path');\n        var _ = require('lodash');\n        var foo = require('foo');\n        var bar = require('bar');\n      ", Output: []string{"\n        var fs = require('fs');\n        var path = require('path');\n        var _ = require('lodash');\n        var foo = require('foo');\n        var bar = require('bar');\n        var index = require('./');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./` import should occur after import of `bar`", Line: 2, Column: 21, EndLine: 2, EndColumn: 34}}},
			// Upstream case 0:18.
			{Code: "\n        var fs = require('fs');\n        var index = require('./');\n      ", Options: []any{map[string]any{"groups": []any{"index", "sibling", "parent", "external", "builtin"}}}, Output: []string{"\n        var index = require('./');\n        var fs = require('fs');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./` import should occur before import of `fs`", Line: 3, Column: 21, EndLine: 3, EndColumn: 34}}},
			// Upstream case 0:19.
			{Code: "\n        var foo = require('./foo').bar;\n        var fs = require('fs');\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `./foo`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:20.
			{Code: "\n        var foo = require('./foo').bar.bar.bar;\n        var fs = require('fs');\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `./foo`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:21.
			{Code: "\n        var foo = require('./foo').bar\n          .bar\n          .bar;\n        var fs = require('fs');\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `./foo`", Line: 5, Column: 18, EndLine: 5, EndColumn: 31}}},
			// Upstream case 0:22.
			{Code: "\n        var foo = require('./foo');\n        var fs = require('fs').bar\n          .bar\n          .bar;\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `./foo`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:23.
			{Code: "\n        var fs = require('fs');\n        var index = require('./');\n        var sibling = require('./foo');\n        var path = require('path');\n      ", Options: []any{map[string]any{"groups": []any{[]any{"builtin", "index"}, []any{"sibling", "parent", "external"}}}}, Output: []string{"\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`path` import should occur before import of `./foo`", Line: 5, Column: 20, EndLine: 5, EndColumn: 35}}},
			// Upstream case 0:24.
			{Code: "\n        var path = require('path');\n        var async = require('async');\n      ", Options: []any{map[string]any{"groups": []any{"index", []any{"sibling", "parent", "external", "internal"}}}}, Output: []string{"\n        var async = require('async');\n        var path = require('path');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`async` import should occur before import of `path`", Line: 3, Column: 21, EndLine: 3, EndColumn: 37}}},
		},
	)
}
