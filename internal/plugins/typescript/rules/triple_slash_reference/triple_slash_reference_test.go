package triple_slash_reference

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestTripleSlashReferenceRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&TripleSlashReferenceRule,
		[]rule_tester.ValidTestCase{
			// Double-slash comments with 'never' options
			{
				Code: `
// <reference path="foo" />
// <reference types="bar" />
// <reference lib="baz" />
import * as foo from 'foo';
`,
				Options: map[string]interface{}{
					"lib":   "never",
					"path":  "never",
					"types": "never",
				},
			},
			// Triple-slash comments with 'always' options
			{
				Code: `
/// <reference path="foo" />
/// <reference types="bar" />
/// <reference lib="baz" />
import * as foo from 'foo';
`,
				Options: map[string]interface{}{
					"lib":   "always",
					"path":  "always",
					"types": "always",
				},
			},
			// Triple-slash with prefer-import when no imports
			{
				Code: `
/// <reference types="foo" />
`,
				Options: map[string]interface{}{
					"types": "prefer-import",
				},
			},
			// Commented-out references
			{
				Code: `
/* /// <reference types="foo" /> */
import * as foo from 'foo';
`,
				Options: map[string]interface{}{
					"lib":   "never",
					"path":  "never",
					"types": "never",
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Triple-slash types with prefer-import (ES6 import)
			{
				Code: `
/// <reference types="foo" />
import * as foo from 'foo';
`,
				Options: map[string]interface{}{
					"types": "prefer-import",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tripleSlashReference",
						Line:      2,
						Column:    1,
					},
				},
			},
			// Triple-slash path when never allowed
			{
				Code: `/// <reference path="foo" />`,
				Options: map[string]interface{}{
					"path": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tripleSlashReference",
						Line:      1,
						Column:    1,
					},
				},
			},
			// Triple-slash types when never allowed
			{
				Code: `/// <reference types="foo" />`,
				Options: map[string]interface{}{
					"types": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tripleSlashReference",
						Line:      1,
						Column:    1,
					},
				},
			},
			// Triple-slash lib when never allowed
			{
				Code: `/// <reference lib="foo" />`,
				Options: map[string]interface{}{
					"lib": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tripleSlashReference",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}
