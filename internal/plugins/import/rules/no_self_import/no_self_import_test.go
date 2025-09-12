package no_self_import_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_self_import"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSelfImportRule(t *testing.T) {
	errors := make([]rule_tester.InvalidTestCaseError, 6)
	for i, err := range errors {
		err.MessageId = "import/no-self-import"
		err.Line = i + 2
		err.Column = 1
		errors[i] = err
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_self_import.NoSelfImportRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { bar } from "./bar.ts"`, FileName: "foo.ts"},
			{Code: `import { bar } from "./bar.json" with { type: "json" }`, FileName: "foo.ts"},
			{Code: `require("./bar.ts")`, FileName: "foo.ts"},
			{Code: `require()`, FileName: "foo.ts"},
			{Code: `require(123)`, FileName: "foo.ts"},
			{Code: `require("./foo.ts", 123)`, FileName: "foo.ts"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
import './foo.ts';
import * as fooStar from './foo.ts';
import fooDefault from './foo.ts';
import { foo } from './foo.ts';
import('./foo.ts');
import('./foo.ts', { with: { type: 'module' } });
`,
				FileName: "foo.ts",
				Errors:   errors,
			},
			{
				Code: `
import './foo.js';
import * as fooStar from './foo.js';
import fooDefault from './foo.js';
import { foo } from './foo.js';
import('./foo.js');
import('./foo.js', { with: { type: 'module' } });
`,
				FileName: "foo.ts",
				Errors:   errors,
			},
			{
				Code: `
import './foo';
import * as fooStar from './foo';
import fooDefault from './foo';
import { foo } from './foo';
import('./foo');
import('./foo', { with: { type: 'module' } });
`,
				FileName: "foo.ts",
				Errors:   errors,
			},
		},
	)
}
