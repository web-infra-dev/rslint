package no_redeclare

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestNoRedeclare(t *testing.T) {
	validTests := []rule_tester.ValidTestCase{
		{
			Code: `
var a = 3;
var b = function () {
  var a = 10;
};`,
		},
		{
			Code: `
var a = 3;
a = 10;`,
		},
		{
			Code: `
if (true) {
  let b = 2;
} else {
  let b = 3;
}`,
		},
		{
			Code: `var Object = 0;`,
			Options: map[string]interface{}{"builtinGlobals": false},
		},
		{
			Code: `
function foo({ bar }: { bar: string }) {
  console.log(bar);
}`,
		},
		{
			Code: `
function A<T>() {}
interface B<T> {}
type C<T> = Array<T>;
class D<T> {}`,
		},
		{
			Code: `
interface A {}
interface A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": true},
		},
		{
			Code: `
interface A {}
class A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": true},
		},
		{
			Code: `
class A {}
namespace A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": true},
		},
		{
			Code: `
interface A {}
class A {}
namespace A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": true},
		},
		{
			Code: `
enum A {}
namespace A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": true},
		},
		{
			Code: `
function A() {}
namespace A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": true},
		},
	}

	invalidTests := []rule_tester.InvalidTestCase{
		{
			Code: `
var a = 3;
var a = 10;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    5,
				},
			},
		},
		{
			Code: `
switch (foo) {
  case a:
    var b = 3;
  case b:
    var b = 4;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      6,
					Column:    9,
				},
			},
		},
		{
			Code: `
var a = {};
var a = [];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    5,
				},
			},
		},
		{
			Code: `
var a;
function a() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
function a() {}
function a() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
var a = function () {};
var a = function () {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    5,
				},
			},
		},
		{
			Code: `
var a = function () {};
var a = new Date();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    5,
				},
			},
		},
		{
			Code: `
var a = 3;
var a = 10;
var a = 15;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    5,
				},
				{
					MessageId: "redeclared",
					Line:      4,
					Column:    5,
				},
			},
		},
		{
			Code: `var Object = 0;`,
			Options: map[string]interface{}{"builtinGlobals": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclaredAsBuiltin",
					Line:      1,
					Column:    5,
				},
			},
		},
		{
			Code: `
var a;
var { a = 0, b: Object = 0 } = {};`,
			Options: map[string]interface{}{"builtinGlobals": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    7,
				},
				{
					MessageId: "redeclaredAsBuiltin",
					Line:      3,
					Column:    17,
				},
			},
		},
		{
			Code: `
type T = 1;
type T = 2;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    6,
				},
			},
		},
		{
			Code: `
interface A {}
interface A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code: `
interface A {}
class A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    7,
				},
			},
		},
		{
			Code: `
class A {}
namespace A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code: `
interface A {}
class A {}
namespace A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    7,
				},
				{
					MessageId: "redeclared",
					Line:      4,
					Column:    11,
				},
			},
		},
		{
			Code: `
class A {}
class A {}
namespace A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    7,
				},
			},
		},
		{
			Code: `
function A() {}
namespace A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code: `
function A() {}
function A() {}
namespace A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
function A() {}
class A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    7,
				},
			},
		},
		{
			Code: `
enum A {}
namespace A {}
enum A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      4,
					Column:    6,
				},
			},
		},
		{
			Code: `
function A() {}
class A {}
namespace A {}`,
			Options: map[string]interface{}{"ignoreDeclarationMerge": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    7,
				},
				{
					MessageId: "redeclared",
					Line:      4,
					Column:    11,
				},
			},
		},
		{
			Code: `
type something = string;
const something = 2;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "redeclared",
					Line:      3,
					Column:    7,
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRedeclareRule, validTests, invalidTests)
}