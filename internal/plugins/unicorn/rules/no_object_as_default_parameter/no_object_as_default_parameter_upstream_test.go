// TestNoObjectAsDefaultParameterUpstream migrates the full valid/invalid suite
// from upstream test/no-object-as-default-parameter.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases live
// in no_object_as_default_parameter_extras_test.go.
package no_object_as_default_parameter_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_object_as_default_parameter "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_object_as_default_parameter"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	identifierMessageID    = "identifier"
	nonIdentifierMessageID = "non-identifier"
	nonIdentifierMessage   = "Do not use an object literal as default."
)

func identifierError(parameter string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: identifierMessageID,
		Message:   "Do not use an object literal as default for parameter `" + parameter + "`.",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func nonIdentifierError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: nonIdentifierMessageID,
		Message:   nonIdentifierMessage,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestNoObjectAsDefaultParameterUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_object_as_default_parameter.NoObjectAsDefaultParameterRule,
		[]rule_tester.ValidTestCase{
			{Code: `const abc = {};`, FileName: "file.js"},
			{Code: `const abc = {foo: 123};`, FileName: "file.js"},
			{Code: `function abc(foo) {}`, FileName: "file.js"},
			{Code: `function abc(foo = null) {}`, FileName: "file.js"},
			{Code: `function abc(foo = undefined) {}`, FileName: "file.js"},
			{Code: `function abc(foo = 123) {}`, FileName: "file.js"},
			{Code: `function abc(foo = true) {}`, FileName: "file.js"},
			{Code: `function abc(foo = "bar") {}`, FileName: "file.js"},
			{Code: `function abc(foo = 123, bar = "foo") {}`, FileName: "file.js"},
			{Code: `function abc(foo = {}) {}`, FileName: "file.js"},
			{Code: `function abc({foo = 123} = {}) {}`, FileName: "file.js"},
			{Code: `(function abc() {})(foo = {a: 123})`, FileName: "file.js"},
			{Code: `const abc = foo => {};`, FileName: "file.js"},
			{Code: `const abc = (foo = null) => {};`, FileName: "file.js"},
			{Code: `const abc = (foo = undefined) => {};`, FileName: "file.js"},
			{Code: `const abc = (foo = 123) => {};`, FileName: "file.js"},
			{Code: `const abc = (foo = true) => {};`, FileName: "file.js"},
			{Code: `const abc = (foo = "bar") => {};`, FileName: "file.js"},
			{Code: `const abc = (foo = 123, bar = "foo") => {};`, FileName: "file.js"},
			{Code: `const abc = (foo = {}) => {};`, FileName: "file.js"},
			{Code: `const abc = ({a = true, b = "foo"}) => {};`, FileName: "file.js"},
			{Code: `const abc = function(foo = 123) {}`, FileName: "file.js"},
			{Code: `const {abc = {foo: 123}} = bar;`, FileName: "file.js"},
			{Code: `const {abc = {null: "baz"}} = bar;`, FileName: "file.js"},
			{Code: `const {abc = {foo: undefined}} = undefined;`, FileName: "file.js"},
			{Code: `const abc = ([{foo = false, bar = 123}]) => {};`, FileName: "file.js"},
			{Code: `const abc = ({foo = {a: 123}}) => {};`, FileName: "file.js"},
			{Code: `const abc = ([foo = {a: 123}]) => {};`, FileName: "file.js"},
			{Code: `const abc = ({foo: bar = {a: 123}}) => {};`, FileName: "file.js"},
			{Code: `const abc = () => (foo = {a: 123});`, FileName: "file.js"},
			{Code: "class A {\n\t[foo = {a: 123}]() {}\n}", FileName: "file.js"},
			{Code: "class A extends (foo = {a: 123}) {\n\ta() {}\n}", FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `function abc(foo = {a: 123}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 1, 14, 1, 17)}},
			{Code: `async function * abc(foo = {a: 123}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 1, 22, 1, 25)}},
			{Code: `function abc(foo = {a: false}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 1, 14, 1, 17)}},
			{Code: `function abc(foo = {a: "bar"}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 1, 14, 1, 17)}},
			{Code: `function abc(foo = {a: "bar", b: {c: true}}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 1, 14, 1, 17)}},
			{Code: `const abc = (foo = {a: false}) => {};`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 1, 14, 1, 17)}},
			{Code: `const abc = (foo = {a: 123, b: false}) => {};`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 1, 14, 1, 17)}},
			{Code: `const abc = (foo = {a: false, b: 1, c: "test", d: null}) => {};`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 1, 14, 1, 17)}},
			{Code: `const abc = function(foo = {a: 123}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 1, 22, 1, 25)}},
			{Code: "class A {\n\tabc(foo = {a: 123}) {}\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 2, 6, 2, 9)}},
			{Code: "class A {\n\tconstructor(foo = {a: 123}) {}\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 2, 14, 2, 17)}},
			{Code: "class A {\n\tset abc(foo = {a: 123}) {}\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 2, 10, 2, 13)}},
			{Code: "class A {\n\tstatic abc(foo = {a: 123}) {}\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 2, 13, 2, 16)}},
			{Code: "class A {\n\t* abc(foo = {a: 123}) {}\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 2, 8, 2, 11)}},
			{Code: "class A {\n\tstatic async * abc(foo = {a: 123}) {}\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 2, 21, 2, 24)}},
			{Code: "class A {\n\t[foo = {a: 123}](foo = {a: 123}) {}\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 2, 19, 2, 22)}},
			{Code: "const A = class {\n\tabc(foo = {a: 123}) {}\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 2, 6, 2, 9)}},
			{Code: "object = {\n\tabc(foo = {a: 123}) {}\n};", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 2, 6, 2, 9)}},
			{Code: "const A = class {\n\tabc({a} = {a: 123}) {}\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{nonIdentifierError(2, 12, 2, 20)}},

			// ---- Upstream snapshot cases ----
			{Code: `/**/function abc(foo = {a: 123}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 1, 18, 1, 21)}},
			{Code: `const abc = (foo = {a: false}) => {};`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("foo", 1, 14, 1, 17)}},
			{Code: `function abc({a} = {a: 123}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{nonIdentifierError(1, 20, 1, 28)}},
			{Code: `function abc([a] = {a: 123}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{nonIdentifierError(1, 20, 1, 28)}},
		},
	)
}
