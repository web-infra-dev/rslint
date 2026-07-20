// TestPreferDestructuringExtrasRealUser covers representative application
// code and reported ESLint issues in addition to synthetic AST branch cases.
package prefer_destructuring

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferDestructuringExtrasRealUser(t *testing.T) {
	assignmentsDisabled := []any{map[string]any{
		"AssignmentExpression": map[string]any{"array": false, "object": false},
	}}
	enforceObject := []any{
		map[string]any{"object": true},
		map[string]any{"enforceForRenamedProperties": true},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		[]rule_tester.ValidTestCase{
			// A dynamic collection index is not an integer literal and cannot be
			// represented by positional array destructuring.
			{Code: "const item = rows[index];"},
			// Renamed environment variables remain allowed by default.
			{Code: "const environment = process.env.ENVIRONMENT;"},
			// ESLint #16514: projects can explicitly keep control-flow
			// assignments when only declarations should be checked.
			{
				Code:    "let value;\nif (condition) {\n  value = source.value;\n}",
				Options: assignmentsDisabled,
			},
			// Member assignments are not identifier assignments and therefore
			// do not enter the default same-name object check.
			{Code: "cache.value = response.value;"},
			// Already-destructured application code is left untouched.
			{Code: "const { data } = await client.get(url);"},
		},
		[]rule_tester.InvalidTestCase{
			// ESLint #14918: build-time process.env access intentionally follows
			// the same same-name object rule as every other member expression.
			oneLineMatrixError(
				"const ENVIRONMENT = process.env.ENVIRONMENT;",
				7,
				"object",
				"const {ENVIRONMENT} = process.env;",
				nil,
				false,
			),
			{
				Code:   "const API_KEY = process.env.API_KEY; const DOMAIN = process.env.DOMAIN;",
				Output: []string{"const {API_KEY} = process.env; const {DOMAIN} = process.env;"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 7, 1, 36),
					preferError("object", 1, 44, 1, 71),
				},
			},
			// ESLint #16043: every integer Number literal is an array access,
			// including indexes beyond zero.
			oneLineMatrixError(
				"const second = \"some string\".split(\" \")[1];",
				7,
				"array",
				"",
				nil,
				false,
			),
			// A hook result is another common tuple-like access; array reports
			// are deliberately not autofixed because preceding slots matter.
			oneLineMatrixError(
				"const state = useState(initial)[0];",
				7,
				"array",
				"",
				nil,
				false,
			),
			{
				Code: "async function loadAll(tasks) { const results = await Promise.all(tasks); " +
					"const first = results[0]; return first; }",
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("array", 1, 81, 1, 99),
				},
			},
			// ESLint #16514: assignment expressions are checked by default even
			// when the variable was declared earlier or assignment is nested in
			// control flow. Assignment diagnostics never carry an autofix.
			{
				Code: "let value;\nif (condition) {\n  value = source.value;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 3, 3, 3, 23),
				},
			},
			// Await, calls, and stripped receiver parentheses retain an exact
			// expression when converted to a declaration destructuring fix.
			{
				Code:   "async function load() { const data = (await client.get(url)).data; }",
				Output: []string{"async function load() { const {data} = await client.get(url); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 31, 1, 66),
				},
			},
			oneLineMatrixError(
				"const value = event.target.value;",
				7,
				"object",
				"const {value} = event.target;",
				nil,
				false,
			),
			oneLineMatrixError(
				"const current = ref.current;",
				7,
				"object",
				"const {current} = ref;",
				nil,
				false,
			),
			oneLineMatrixError(
				"const id = response.data.user.id;",
				7,
				"object",
				"const {id} = response.data.user;",
				nil,
				false,
			),
			oneLineMatrixError(
				"const join = require(\"path\").join;",
				7,
				"object",
				"const {join} = require(\"path\");",
				nil,
				false,
			),
			// Type arguments belong to the retained call receiver.
			oneLineMatrixError(
				"const data = client.get<Result>().data;",
				7,
				"object",
				"const {data} = client.get<Result>();",
				nil,
				false,
			),
			{
				Code:   "function Component(props) { const title = props.title; return title; }",
				Output: []string{"function Component(props) { const {title} = props; return title; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 35, 1, 54),
				},
			},
			{
				Code:   "function reducer(state) { const user = state.auth.user; return user; }",
				Output: []string{"function reducer(state) { const {user} = state.auth; return user; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 33, 1, 55),
				},
			},
			{
				Code: "async function route(request) { const id = request.params.id; " +
					"const token = request.headers.token; return id + token; }",
				Output: []string{
					"async function route(request) { const {id} = request.params; " +
						"const {token} = request.headers; return id + token; }",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 1, 39, 1, 61),
					preferError("object", 1, 69, 1, 98),
				},
			},
			{
				Code: "const zero = rows[0], one = rows[1], dynamic = rows[index];",
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("array", 1, 7, 1, 21),
					preferError("array", 1, 23, 1, 36),
				},
			},
			// Renamed-property enforcement applies to real environment aliases
			// but still cannot offer the declaration-only same-name autofix.
			oneLineMatrixError(
				"const environment = process.env.ENVIRONMENT;",
				7,
				"object",
				"",
				enforceObject,
				false,
			),
			// UTF-16 columns and CRLF line boundaries must match ESLint while
			// the fix preserves all text outside the selected declarator.
			{
				Code:   "const before = \"😀\";\r\nconst value = event.target.value;",
				Output: []string{"const before = \"😀\";\r\nconst {value} = event.target;"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferError("object", 2, 7, 2, 33),
				},
			},
		},
	)
}
