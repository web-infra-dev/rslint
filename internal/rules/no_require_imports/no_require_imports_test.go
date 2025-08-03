package no_require_imports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoRequireImportsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRequireImportsRule, []rule_tester.ValidTestCase{
		// ES6 imports are valid
		{Code: "import { l } from 'lib';"},
		{Code: "var lib3 = load('not_an_import');"},
		{Code: "var lib4 = lib2.subImport;"},
		{Code: "var lib7 = 700;"},
		{Code: "import lib9 = lib2.anotherSubImport;"},
		{Code: "import lib10 from 'lib10';"},
		{Code: "var lib3 = load?.('not_an_import');"},

		// Local require should be allowed
		{Code: `
import { createRequire } from 'module';
const require = createRequire();
require('remark-preset-prettier');
		`},

		// Allow patterns
		{
			Code:    "const pkg = require('./package.json');",
			Options: map[string]interface{}{"allow": []interface{}{"/package\\.json$"}},
		},
		{
			Code:    "const pkg = require('../package.json');",
			Options: map[string]interface{}{"allow": []interface{}{"/package\\.json$"}},
		},
		{
			Code:    "const pkg = require(`./package.json`);",
			Options: map[string]interface{}{"allow": []interface{}{"/package\\.json$"}},
		},
		{
			Code:    "const pkg = require('../packages/package.json');",
			Options: map[string]interface{}{"allow": []interface{}{"/package\\.json$"}},
		},
		{
			Code:    "import pkg = require('../packages/package.json');",
			Options: map[string]interface{}{"allow": []interface{}{"/package\\.json$"}},
		},
		{
			Code:    "import pkg = require('data.json');",
			Options: map[string]interface{}{"allow": []interface{}{"\\.json$"}},
		},
		{
			Code:    "import pkg = require('some-package');",
			Options: map[string]interface{}{"allow": []interface{}{"^some-package$"}},
		},

		// AllowAsImport option
		{
			Code:    "import foo = require('foo');",
			Options: map[string]interface{}{"allowAsImport": true},
		},
		{
			Code: `
let require = bazz;
trick(require('foo'));
			`,
			Options: map[string]interface{}{"allowAsImport": true},
		},
		{
			Code: `
let require = bazz;
const foo = require('./foo.json') as Foo;
			`,
			Options: map[string]interface{}{"allowAsImport": true},
		},
		{
			Code: `
let require = bazz;
const foo: Foo = require('./foo.json').default;
			`,
			Options: map[string]interface{}{"allowAsImport": true},
		},
		{
			Code: `
let require = bazz;
const foo = <Foo>require('./foo.json');
			`,
			Options: map[string]interface{}{"allowAsImport": true},
		},
		{
			Code: `
let require = bazz;
const configValidator = new Validator(require('./a.json'));
configValidator.addSchema(require('./a.json'));
			`,
			Options: map[string]interface{}{"allowAsImport": true},
		},
		{
			Code: `
let require = bazz;
require('foo');
			`,
			Options: map[string]interface{}{"allowAsImport": true},
		},
		{
			Code: `
let require = bazz;
require?.('foo');
			`,
			Options: map[string]interface{}{"allowAsImport": true},
		},
		{
			Code: `
import { createRequire } from 'module';
const require = createRequire();
require('remark-preset-prettier');
			`,
			Options: map[string]interface{}{"allowAsImport": true},
		},
	}, []rule_tester.InvalidTestCase{
		// Basic require() calls
		{
			Code: "var lib = require('lib');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "let lib2 = require('lib2');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `
var lib5 = require('lib5'),
  lib6 = require('lib6');
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      2,
					Column:    12,
				},
				{
					MessageId: "noRequireImports",
					Line:      3,
					Column:    10,
				},
			},
		},

		// import = require() style
		{
			Code: "import lib8 = require('lib8');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    15,
				},
			},
		},

		// Optional chaining
		{
			Code: "var lib = require?.('lib');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "let lib2 = require?.('lib2');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `
var lib5 = require?.('lib5'),
  lib6 = require?.('lib6');
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      2,
					Column:    12,
				},
				{
					MessageId: "noRequireImports",
					Line:      3,
					Column:    10,
				},
			},
		},

		// Disallowed even with allow patterns that don't match
		{
			Code: "const pkg = require('./package.json');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code:    "const pkg = require('./package.jsonc');",
			Options: map[string]interface{}{"allow": []interface{}{"/package\\.json$"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code:    "const pkg = require(`./package.jsonc`);",
			Options: map[string]interface{}{"allow": []interface{}{"/package\\.json$"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: "import pkg = require('./package.json');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    14,
				},
			},
		},
		{
			Code:    "import pkg = require('./package.jsonc');",
			Options: map[string]interface{}{"allow": []interface{}{"/package\\.json$"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    14,
				},
			},
		},
		{
			Code:    "import pkg = require('./package.json');",
			Options: map[string]interface{}{"allow": []interface{}{"^some-package$"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    14,
				},
			},
		},

		// With allowAsImport but not import = require
		{
			Code:    "var foo = require?.('foo');",
			Options: map[string]interface{}{"allowAsImport": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code:    "let foo = trick(require?.('foo'));",
			Options: map[string]interface{}{"allowAsImport": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    17,
				},
			},
		},
		{
			Code:    "trick(require('foo'));",
			Options: map[string]interface{}{"allowAsImport": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    7,
				},
			},
		},
		{
			Code:    "const foo = require('./foo.json') as Foo;",
			Options: map[string]interface{}{"allowAsImport": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code:    "const foo: Foo = require('./foo.json').default;",
			Options: map[string]interface{}{"allowAsImport": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    18,
				},
			},
		},
		{
			Code:    "const foo = <Foo>require('./foo.json');",
			Options: map[string]interface{}{"allowAsImport": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    18,
				},
			},
		},
		{
			Code: `
const configValidator = new Validator(require('./a.json'));
configValidator.addSchema(require('./a.json'));
			`,
			Options: map[string]interface{}{"allowAsImport": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      2,
					Column:    39,
				},
				{
					MessageId: "noRequireImports",
					Line:      3,
					Column:    27,
				},
			},
		},
		{
			Code:    "require(foo);",
			Options: map[string]interface{}{"allowAsImport": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    "require?.(foo);",
			Options: map[string]interface{}{"allowAsImport": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noRequireImports",
					Line:      1,
					Column:    1,
				},
			},
		},
	})
}
