package id_match

// TestIdMatchExtrasRealuser locks in code shapes taken from the upstream rule's
// issue tracker. Its siblings are id_match_extras_dim4_test.go and
// id_match_extras_branches_test.go.

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestIdMatchExtrasRealuser(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&IdMatchRule,
		[]rule_tester.ValidTestCase{
			// ---- Real-user: eslint#4042 - a constructor name is never a declaration ----
			{
				Code: `var startDate = new Date();
var myArray = new Array();`,
				Options: []any{`^[a-z$]+([A-Z][a-z]+)*$`},
			},
			// ---- Real-user: eslint#5885 - onlyDeclarations must not reach a property value ----
			{
				Code:    `export default { context: __dirname };`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true, "onlyDeclarations": true}},
			},
			// ---- Real-user: eslint#15395 - a native class reference is not renamed ----
			{
				Code: `const a = Object.keys(b);
const c = Array.from(b);`,
				Options: []any{`^\$?[a-z]+([A-Z0-9][a-z0-9]+)*$`, map[string]any{"properties": true}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Real-user: eslint#14005 - a named import binds a local name and is checked ----
			{
				Code:    `import { stub_is_valid } from '../stubs';`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'stub_is_valid' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 23,
					},
				},
			},
			// ---- Real-user: eslint#15123 - properties and onlyDeclarations together ----
			{
				Code: `const foo = {
    foo_one: 1,
    fooBar: 2,
};`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true, "onlyDeclarations": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'foo_one' does not match the pattern '^[^_]+$'.`,
						Line:      2,
						Column:    5,
						EndLine:   2,
						EndColumn: 12,
					},
				},
			},
		},
	)
}
