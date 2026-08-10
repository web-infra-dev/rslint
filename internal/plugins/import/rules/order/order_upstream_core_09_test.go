// TestOrderUpstreamCore09 migrates the upstream core cases from import-js/eslint-plugin-import at 1a39fb8672c83f8781c3a97f02eee00a7a2f0253.
// It preserves upstream messages, source positions, options, and iterative autofix outputs.
// Framework-only parser/resolver settings use rslint native equivalents; tsgo lock-ins live in order_extras_*_test.go.
package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderUpstreamCore09(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Upstream case 0:85.
			{Code: "\n        var fs = require('fs');\n\n        var path = require('path');\n\n        var { util1, util2, util3 } = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n\n        var sibling2 = require('./bar');\n\n        var sibling3 = require('./foobar');\n      ", Options: []any{map[string]any{"newlines-between": "always-and-inside-groups", "consolidateIslands": "inside-groups"}}, Output: []string{"\n        var fs = require('fs');\n        var path = require('path');\n        var { util1, util2, util3 } = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consolidate", Message: "There should be no empty lines between this single-line import and the single-line import that follows it", Line: 2, Column: 18, EndLine: 2, EndColumn: 31}, {MessageId: "consolidate", Message: "There should be no empty lines between this single-line import and the single-line import that follows it", Line: 4, Column: 20, EndLine: 4, EndColumn: 35}, {MessageId: "consolidate", Message: "There should be no empty lines between this single-line import and the single-line import that follows it", Line: 24, Column: 24, EndLine: 24, EndColumn: 40}}},
		},
	)
}
