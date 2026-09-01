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
			{Code: `import { test } from 'rstack/test'`},
			{Code: `import { test } from 'vitest'`},
			{Code: `import { test } from 'node:testing'`},
			{Code: `const assert = require('node:assert')`},
			{Code: `const t = import('node:test')`},
			{Code: `import assert = require('node:assert')`},
			{Code: `import x from a.b;`},
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

// The same test API is reachable as `@rstest/core` and as `rstack/test`, and a
// project set up through the Rstack CLI depends on `rstack` alone. The fix
// therefore writes whichever specifier the file itself already uses, so that a
// fixed import resolves in the project it was written for.
func TestNoImportNodeTestReplacementFollowsFile(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_import_node.NoImportNodeTestRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// A sibling import names the specifier this project uses.
			{
				Code: `import { expect } from 'rstack/test';
import { test } from 'node:test';`,
				Output: []string{`import { expect } from 'rstack/test';
import { test } from 'rstack/test';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			// A namespace import, a require and an import-equals name it too.
			{
				Code: `import * as rstest from 'rstack/test';
import { test } from 'node:test';`,
				Output: []string{`import * as rstest from 'rstack/test';
import { test } from 'rstack/test';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code: `const { expect } = require('rstack/test');
import { test } from 'node:test';`,
				Output: []string{`const { expect } = require('rstack/test');
import { test } from 'rstack/test';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			{
				Code: `import rstest = require('rstack/test');
import { test } from 'node:test';`,
				Output: []string{`import rstest = require('rstack/test');
import { test } from 'rstack/test';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			// A globals type reference names it without any import.
			{
				Code: `/// <reference types="rstack/test/globals" />
import { test } from 'node:test';`,
				Output: []string{`/// <reference types="rstack/test/globals" />
import { test } from 'rstack/test';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			// A file that says `@rstest/core` keeps `@rstest/core`, even
			// alongside an Rstack import: the direct dependency is known to
			// resolve, so it is the safer of the two to write.
			{
				Code: `import { expect } from '@rstest/core';
import { describe } from 'rstack/test';
import { test } from 'node:test';`,
				Output: []string{`import { expect } from '@rstest/core';
import { describe } from 'rstack/test';
import { test } from '@rstest/core';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			// A sibling subpath of the same package is a different API and
			// says nothing about which test specifier the project uses.
			{
				Code: `import { defineConfig } from 'rstack/lib';
import { test } from 'node:test';`,
				Output: []string{`import { defineConfig } from 'rstack/lib';
import { test } from '@rstest/core';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
			// With nothing to go on, the fix keeps `@rstest/core`.
			{
				Code:   `import { test } from 'node:test';`,
				Output: []string{`import { test } from '@rstest/core';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noImportNodeTest"}},
			},
		},
	)
}
