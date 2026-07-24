package triple_slash_reference

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestTripleSlashReferenceRule(t *testing.T) {
	always := map[string]interface{}{
		"lib":   "always",
		"path":  "always",
		"types": "always",
	}
	never := map[string]interface{}{
		"lib":   "never",
		"path":  "never",
		"types": "never",
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&TripleSlashReferenceRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `
// <reference path="foo" />
// <reference types="bar" />
// <reference lib="baz" />
import * as foo from 'foo';
import * as bar from 'bar';
import * as baz from 'baz';
`,
				Options: never,
			},
			{
				Code: `
// <reference path="foo" />
// <reference types="bar" />
// <reference lib="baz" />
import foo = require('foo');
import bar = require('bar');
import baz = require('baz');
`,
				Options: never,
			},
			{
				Code: `
/// <reference path="foo" />
/// <reference types="bar" />
/// <reference lib="baz" />
import * as foo from 'foo';
import * as bar from 'bar';
import * as baz from 'baz';
`,
				Options: always,
			},
			{
				Code: `
/// <reference path="foo" />
/// <reference types="bar" />
/// <reference lib="baz" />
import foo = require('foo');
import bar = require('bar');
import baz = require('baz');
`,
				Options: always,
			},
			{
				Code: `
/// <reference path="foo" />
/// <reference types="bar" />
/// <reference lib="baz" />
import foo = foo;
import bar = bar;
import baz = baz;
`,
				Options: always,
			},
			{
				Code: `
/// <reference path="foo" />
/// <reference types="bar" />
/// <reference lib="baz" />
import foo = foo.foo;
import bar = bar.bar.bar.bar;
import baz = baz.baz;
`,
				Options: always,
			},
			{
				Code:    "import * as foo from 'foo';",
				Options: map[string]interface{}{"path": "never"},
			},
			{
				Code:    "import foo = require('foo');",
				Options: map[string]interface{}{"path": "never"},
			},
			{
				Code:    "import * as foo from 'foo';",
				Options: map[string]interface{}{"types": "never"},
			},
			{
				Code:    "import foo = require('foo');",
				Options: map[string]interface{}{"types": "never"},
			},
			{
				Code:    "import * as foo from 'foo';",
				Options: map[string]interface{}{"lib": "never"},
			},
			{
				Code:    "import foo = require('foo');",
				Options: map[string]interface{}{"lib": "never"},
			},
			{
				Code:    "import * as foo from 'foo';",
				Options: map[string]interface{}{"types": "prefer-import"},
			},
			{
				Code:    "import foo = require('foo');",
				Options: map[string]interface{}{"types": "prefer-import"},
			},
			{
				Code: `
/// <reference types="foo" />
import * as bar from 'bar';
`,
				Options: map[string]interface{}{"types": "prefer-import"},
			},
			{
				Code: `
/*
/// <reference types="foo" />
*/
import * as foo from 'foo';
`,
				Options: never,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
/// <reference types="foo" />
import * as foo from 'foo';
`,
				Options: map[string]interface{}{"types": "prefer-import"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(2, 1, "foo")},
			},
			{
				Code: `
/// <reference types="foo" />
import foo = require('foo');
`,
				Options: map[string]interface{}{"types": "prefer-import"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(2, 1, "foo")},
			},
			{
				Code:    `/// <reference path="foo" />`,
				Options: map[string]interface{}{"path": "never"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(1, 1, "foo")},
			},
			{
				Code:    `/// <reference types="foo" />`,
				Options: map[string]interface{}{"types": "never"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(1, 1, "foo")},
			},
			{
				Code:    `/// <reference lib="foo" />`,
				Options: map[string]interface{}{"lib": "never"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(1, 1, "foo")},
			},
		},
	)
}

func tripleSlashReferenceError(line int, column int, module string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "tripleSlashReference",
		Message:   "Do not use a triple slash reference for " + module + ", use `import` style instead.",
		Line:      line,
		Column:    column,
	}
}
