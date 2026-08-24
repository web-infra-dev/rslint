// TestFuncNamesUpstream migrates the full valid/invalid suite from ESLint
// v10.8.1 tests/lib/rules/func-names.js 1:1. languageOptions.ecmaVersion is
// dropped from ported cases: rslint's parser does not gate syntax on it.
// languageOptions.sourceType is dropped too — every upstream case that sets
// it already contains real import/export syntax, so rslint infers
// module-ness the same way regardless of the flag. rslint-specific lock-ins
// live in func_names_extras_test.go.
package func_names

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestFuncNamesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&FuncNamesRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid: default ("always" implied) ----
			{Code: "Foo.prototype.bar = function bar(){};"},
			{Code: "Foo.prototype.bar = () => {}"},
			{Code: "function foo(){}"},
			{Code: "function test(d, e, f) {}"},
			{Code: "new function bar(){}"},
			{Code: "exports = { get foo() { return 1; }, set bar(val) { return val; } };"},
			{Code: "({ foo() { return 1; } });"},
			{Code: "class A { constructor(){} foo(){} get bar(){} set baz(value){} static qux(){}}"},

			// ---- upstream valid: "always" ----
			{Code: "function foo() {}", Options: []any{"always"}},
			{Code: "var a = function foo() {};", Options: []any{"always"}},

			// ---- upstream valid: "as-needed" ----
			{Code: "class A { constructor(){} foo(){} get bar(){} set baz(value){} static qux(){}}", Options: []any{"as-needed"}},
			{Code: "({ foo() {} });", Options: []any{"as-needed"}},
			{Code: "var foo = function(){};", Options: []any{"as-needed"}},
			{Code: "({foo: function(){}});", Options: []any{"as-needed"}},
			{Code: "(foo = function(){});", Options: []any{"as-needed"}},
			{Code: "({foo = function(){}} = {});", Options: []any{"as-needed"}},
			{Code: "({key: foo = function(){}} = {});", Options: []any{"as-needed"}},
			{Code: "[foo = function(){}] = [];", Options: []any{"as-needed"}},
			{Code: "function fn(foo = function(){}) {}", Options: []any{"as-needed"}},

			// ---- upstream valid: "never" ----
			{Code: "function foo() {}", Options: []any{"never"}},
			{Code: "var a = function() {};", Options: []any{"never"}},
			{Code: "var a = function foo() { foo(); };", Options: []any{"never"}},
			{Code: "var foo = {bar: function() {}};", Options: []any{"never"}},
			{Code: "$('#foo').click(function() {});", Options: []any{"never"}},
			{Code: "Foo.prototype.bar = function() {};", Options: []any{"never"}},
			{Code: "class A { constructor(){} foo(){} get bar(){} set baz(value){} static qux(){}}", Options: []any{"never"}},
			{Code: "({ foo() {} });", Options: []any{"never"}},

			// ---- upstream valid: export default ----
			{Code: "export default function foo() {}", Options: []any{"always"}},
			{Code: "export default function foo() {}", Options: []any{"as-needed"}},
			{Code: "export default function foo() {}", Options: []any{"never"}},
			{Code: "export default function() {}", Options: []any{"never"}},

			// ---- upstream valid: generators ----
			{Code: "var foo = bar(function *baz() {});", Options: []any{"always"}},
			{Code: "var foo = bar(function *baz() {});", Options: []any{"always", map[string]any{"generators": "always"}}},
			{Code: "var foo = bar(function *baz() {});", Options: []any{"always", map[string]any{"generators": "as-needed"}}},
			{Code: "var foo = function*() {};", Options: []any{"always", map[string]any{"generators": "as-needed"}}},
			{Code: "var foo = bar(function *baz() {});", Options: []any{"as-needed"}},
			{Code: "var foo = function*() {};", Options: []any{"as-needed"}},
			{Code: "var foo = bar(function *baz() {});", Options: []any{"as-needed", map[string]any{"generators": "always"}}},
			{Code: "var foo = bar(function *baz() {});", Options: []any{"as-needed", map[string]any{"generators": "as-needed"}}},
			{Code: "var foo = function*() {};", Options: []any{"as-needed", map[string]any{"generators": "as-needed"}}},
			{Code: "var foo = bar(function *baz() {});", Options: []any{"never", map[string]any{"generators": "always"}}},
			{Code: "var foo = bar(function *baz() {});", Options: []any{"never", map[string]any{"generators": "as-needed"}}},
			{Code: "var foo = function*() {};", Options: []any{"never", map[string]any{"generators": "as-needed"}}},

			{Code: "var foo = bar(function *() {});", Options: []any{"never"}},
			{Code: "var foo = function*() {};", Options: []any{"never"}},
			{Code: "(function*() {}())", Options: []any{"never"}},
			{Code: "var foo = bar(function *() {});", Options: []any{"never", map[string]any{"generators": "never"}}},
			{Code: "var foo = function*() {};", Options: []any{"never", map[string]any{"generators": "never"}}},
			{Code: "(function*() {}())", Options: []any{"never", map[string]any{"generators": "never"}}},
			{Code: "var foo = bar(function *() {});", Options: []any{"always", map[string]any{"generators": "never"}}},
			{Code: "var foo = function*() {};", Options: []any{"always", map[string]any{"generators": "never"}}},
			{Code: "(function*() {}())", Options: []any{"always", map[string]any{"generators": "never"}}},
			{Code: "var foo = bar(function *() {});", Options: []any{"as-needed", map[string]any{"generators": "never"}}},
			{Code: "var foo = function*() {};", Options: []any{"as-needed", map[string]any{"generators": "never"}}},
			{Code: "(function*() {}())", Options: []any{"as-needed", map[string]any{"generators": "never"}}},

			// ---- upstream valid: class fields ----
			{Code: "class C { foo = function() {}; }", Options: []any{"as-needed"}},
			{Code: "class C { [foo] = function() {}; }", Options: []any{"as-needed"}},
			{Code: "class C { #foo = function() {}; }", Options: []any{"as-needed"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid: default ("always" implied) ----
			{
				Code: "Foo.prototype.bar = function() {};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 21, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: "(function(){}())",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 2, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: "f(function(){})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 3, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code: "var a = new Date(function() {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 18, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: "var test = function(d, e, f) {};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 12, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: "new function() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- upstream invalid: "as-needed" ----
			{
				Code:    "Foo.prototype.bar = function() {};",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 21, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:    "(function(){}())",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 2, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    "f(function(){})",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 3, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    "var a = new Date(function() {});",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 18, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "new function() {}",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    "var {foo} = function(){};",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "({ a: obj.prop = function(){} } = foo);",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 18, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "[obj.prop = function(){}] = foo;",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "var { a: [b] = function(){} } = foo;",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "function foo({ a } = function(){}) {};",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 22, EndLine: 1, EndColumn: 30},
				},
			},

			// ---- upstream invalid: "never" ----
			{
				Code:    "var x = function foo() {};",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named function 'foo'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "Foo.prototype.bar = function foo() {};",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named function 'foo'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:    "({foo: function foo() {}})",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named method 'foo'.", Line: 1, Column: 3, EndLine: 1, EndColumn: 20},
				},
			},

			// ---- upstream invalid: export default ----
			{
				Code:    "export default function() {}",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "export default function() {}",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "export default (function(){});",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 17, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- upstream invalid: generators ----
			{
				Code:    "var foo = bar(function *() {});",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "var foo = function*() {};",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 11, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "(function*() {}())",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 2, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    "var foo = bar(function *() {});",
				Options: []any{"always", map[string]any{"generators": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "var foo = function*() {};",
				Options: []any{"always", map[string]any{"generators": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 11, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "(function*() {}())",
				Options: []any{"always", map[string]any{"generators": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 2, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    "var foo = bar(function *() {});",
				Options: []any{"always", map[string]any{"generators": "as-needed"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "(function*() {}())",
				Options: []any{"always", map[string]any{"generators": "as-needed"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 2, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    "var foo = bar(function *() {});",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "(function*() {}())",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 2, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    "var foo = bar(function *() {});",
				Options: []any{"as-needed", map[string]any{"generators": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "var foo = function*() {};",
				Options: []any{"as-needed", map[string]any{"generators": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 11, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "(function*() {}())",
				Options: []any{"as-needed", map[string]any{"generators": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 2, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    "var foo = bar(function *() {});",
				Options: []any{"as-needed", map[string]any{"generators": "as-needed"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "(function*() {}())",
				Options: []any{"as-needed", map[string]any{"generators": "as-needed"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 2, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    "var foo = bar(function *() {});",
				Options: []any{"never", map[string]any{"generators": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "var foo = function*() {};",
				Options: []any{"never", map[string]any{"generators": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 11, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "(function*() {}())",
				Options: []any{"never", map[string]any{"generators": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 2, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    "var foo = bar(function *() {});",
				Options: []any{"never", map[string]any{"generators": "as-needed"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "(function*() {}())",
				Options: []any{"never", map[string]any{"generators": "as-needed"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed generator function.", Line: 1, Column: 2, EndLine: 1, EndColumn: 11},
				},
			},

			{
				Code:    "var foo = bar(function *baz() {});",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named generator function 'baz'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "var foo = bar(function *baz() {});",
				Options: []any{"never", map[string]any{"generators": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named generator function 'baz'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "var foo = bar(function *baz() {});",
				Options: []any{"always", map[string]any{"generators": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named generator function 'baz'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "var foo = bar(function *baz() {});",
				Options: []any{"as-needed", map[string]any{"generators": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named generator function 'baz'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28},
				},
			},

			// ---- upstream invalid: class fields ----
			{
				Code:    "class C { foo = function() {} }",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "class C { [foo] = function() {} }",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed method.", Line: 1, Column: 11, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    "class C { #foo = function() {} }",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed private method #foo.", Line: 1, Column: 11, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "class C { foo = bar(function() {}) }",
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnamed", Message: "Unexpected unnamed function.", Line: 1, Column: 21, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:    "class C { foo = function bar() {} }",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "named", Message: "Unexpected named method 'foo'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 29},
				},
			},
		},
	)
}
