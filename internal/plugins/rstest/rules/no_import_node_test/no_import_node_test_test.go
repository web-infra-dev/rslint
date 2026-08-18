package no_import_node_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	no_import_node "github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_import_node_test"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoImportNodeTestRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_import_node.NoImportNodeTestRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { test } from '@rstest/core'`},
			{Code: `import { test } from 'vitest'`},
			{Code: `import { test } from 'node:testing'`},
			{Code: `const assert = require('node:assert')`},
			{Code: `const t = import('node:test')`},
			{Code: `import assert = require('node:assert')`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `import { test } from 'node:test'`,
				Output: []string{`import { test } from '@rstest/core'`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noImportNodeTest",
					Message:   "Do not import the Node test runner in Rstest test files",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 33,
				}},
			},
			{
				Code:   `import { test } from "node:test"`,
				Output: []string{`import { test } from "@rstest/core"`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import { test, describe } from 'node:test'`,
				Output: []string{`import { test, describe } from '@rstest/core'`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import { test as nodeTest, describe } from 'node:test'`,
				Output: []string{`import { test as nodeTest, describe } from '@rstest/core'`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import nodeTest from 'node:test'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import * as nodeTest from 'node:test'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import 'node:test'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import { mock } from 'node:test'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import { test, mock } from 'node:test'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import nodeTest, { test } from 'node:test'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import { tap } from 'node:test/reporters'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import { test } from 'node:test/reporters'`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `const nodeTest = require('node:test')`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `const { tap } = require('node:test/reporters')`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import nodeTest = require('node:test')`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import reporters = require('node:test/reporters')`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `const nodeTest = (require)('node:test')`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `const nodeTest = require(('node:test'))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code:   `import { test as mock } from 'node:test'`,
				Output: []string{`import { test as mock } from '@rstest/core'`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
		},
	)
}
