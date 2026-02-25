package naming_convention

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNamingConventionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		// Default config: camelCase for most things, PascalCase for types
		{Code: `const myVariable = 1;`},
		{Code: `let anotherVar = "hello";`},
		{Code: `function myFunction() {}`},
		{Code: `class MyClass {}`},
		{Code: `interface MyInterface {}`},
		{Code: `type MyType = string;`},
		{Code: `enum MyEnum { a, b }`},

		// Leading/trailing underscore allowed by default
		{Code: `const _privateVar = 1;`},
		{Code: `const trailingVar_ = 1;`},
		{Code: `const UPPER_CASE = 1;`},

		// Variable can be UPPER_CASE (default config)
		{Code: `const MY_CONSTANT = 1;`},
		{Code: `const myVar = 1;`},

		// Import selectors (default: camelCase or PascalCase)
		{Code: `import myModule from 'module';`},
		{Code: `import MyModule from 'module';`},
		{Code: `import { myExport } from 'module';`},
		{Code: `import { MyExport } from 'module';`},

		// Custom config: snake_case for variables
		{
			Code: `const my_variable = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"snake_case"},
				},
			},
		},

		// Custom config: PascalCase for variables
		{
			Code: `const MyVariable = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"PascalCase"},
				},
			},
		},

		// Leading underscore required
		{
			Code: `const _myVariable = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector":          "variable",
					"format":            []interface{}{"camelCase"},
					"leadingUnderscore": "require",
				},
			},
		},

		// Trailing underscore required
		{
			Code: `const myVariable_ = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector":           "variable",
					"format":             []interface{}{"camelCase"},
					"trailingUnderscore": "require",
				},
			},
		},

		// Prefix required - after stripping prefix, remaining must match format
		{
			Code: `const isActive = true;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"PascalCase"},
					"prefix":   []interface{}{"is", "has", "should"},
				},
			},
		},

		// Suffix required
		{
			Code: `const nameStr = "hello";`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"camelCase"},
					"suffix":   []interface{}{"Str", "Num"},
				},
			},
		},

		// Format null: skip format check
		{
			Code: `const ANY_NAME = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   nil,
				},
			},
		},

		// Custom regex
		{
			Code: `const myVar123 = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"camelCase"},
					"custom": map[string]interface{}{
						"regex": "\\d+$",
						"match": true,
					},
				},
			},
		},

		// Filter
		{
			Code: `const __special__ = 1; const myNormal = 2;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"camelCase"},
					"filter": map[string]interface{}{
						"regex": "^__.*__$",
						"match": false,
					},
				},
			},
		},

		// Class members
		{
			Code: `class MyClass { myProperty = 1; myMethod() {} }`,
		},

		// Enum members with default config - camelCase required
		{Code: `enum MyEnum { camelCase = 1 }`},
		{Code: `enum MyEnum { myValue = 1 }`},

		// Type parameter
		{Code: `function foo<T>() {}`},
		{Code: `function foo<TData>() {}`},

		// Interface with PascalCase
		{Code: `interface MyInterface { myProp: string; }`},

		// Multiple formats allowed
		{
			Code: `const myVar = 1; const MY_VAR = 2;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"camelCase", "UPPER_CASE"},
				},
			},
		},

		// Parameter
		{Code: `function fn(myParam: string) {}`},

		// Destructured variable
		{
			Code: `const { myProp } = ({} as any);`,
		},

		// Function expression assigned to variable
		{Code: `const myFunc = () => {};`},
		{Code: `const myFunc = function() {};`},

		// Object literal properties
		{Code: `const obj = { myProp: 1 };`},
		{Code: `const obj = { myMethod() {} };`},

		// Type properties and methods
		{Code: `type MyType = { myProp: string; myMethod(): void; };`},

		// Accessor
		{Code: `class Foo { get myProp() { return 1; } set myProp(v: number) {} }`},

		// Strict camelCase
		{
			Code: `const myId = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"strictCamelCase"},
				},
			},
		},

		// Strict PascalCase
		{
			Code: `class MyClass {}`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "class",
					"format":   []interface{}{"StrictPascalCase"},
				},
			},
		},
		// Unused modifier: unused identifiers should match PascalCase
		{
			Code: "const UnusedVar = 1;\nfunction UnusedFunc(\n  UnusedParam: string,\n) {}\nclass UnusedClass {}\ninterface UnusedInterface {}\ntype UnusedType<\n  UnusedTypeParam,\n> = {};\n\nexport const used_var = 1;\nexport function used_func(\n  used_param: string,\n) {\n  return used_param;\n}\nexport class used_class {}\nexport interface used_interface {}\nexport type used_type<\n  used_typeparam,\n> = used_typeparam;",
			Options: []interface{}{
				map[string]interface{}{
					"format":   []interface{}{"snake_case"},
					"selector": "default",
				},
				map[string]interface{}{
					"format":    []interface{}{"PascalCase"},
					"modifiers": []interface{}{"unused"},
					"selector":  "default",
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Variable violating camelCase
		{
			Code: `const my_variable = 1;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 7},
			},
		},

		// Function violating camelCase
		{
			Code: `function MyFunction() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 10},
			},
		},

		// Class violating PascalCase
		{
			Code: `class myClass {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 7},
			},
		},

		// Interface violating PascalCase
		{
			Code: `interface myInterface {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 11},
			},
		},

		// Type alias violating PascalCase
		{
			Code: `type myType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 6},
			},
		},

		// Enum violating PascalCase (member A also violates camelCase from default)
		{
			Code: `enum myEnum { a }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 6},
			},
		},

		// Leading underscore forbidden
		{
			Code: `const _myVariable = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector":          "variable",
					"format":            []interface{}{"camelCase"},
					"leadingUnderscore": "forbid",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedUnderscore", Line: 1, Column: 7},
			},
		},

		// Trailing underscore forbidden
		{
			Code: `const myVariable_ = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector":           "variable",
					"format":             []interface{}{"camelCase"},
					"trailingUnderscore": "forbid",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedUnderscore", Line: 1, Column: 7},
			},
		},

		// Leading underscore required but missing
		{
			Code: `const myVariable = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector":          "variable",
					"format":            []interface{}{"camelCase"},
					"leadingUnderscore": "require",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingUnderscore", Line: 1, Column: 7},
			},
		},

		// Missing prefix
		{
			Code: `const active = true;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"camelCase"},
					"prefix":   []interface{}{"is", "has"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAffix", Line: 1, Column: 7},
			},
		},

		// Missing suffix
		{
			Code: `const name = "hello";`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"camelCase"},
					"suffix":   []interface{}{"Str", "Num"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAffix", Line: 1, Column: 7},
			},
		},

		// Custom regex not matching
		{
			Code: `const myVar = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"camelCase"},
					"custom": map[string]interface{}{
						"regex": "\\d+$",
						"match": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "satisfyCustom", Line: 1, Column: 7},
			},
		},

		// UPPER_CASE violating camelCase-only rule
		{
			Code: `const MY_VAR = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"camelCase"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 7},
			},
		},

		// snake_case variable violating PascalCase
		{
			Code: `const my_var = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"PascalCase"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 7},
			},
		},

		// Parameter violating format
		{
			Code: `function fn(MY_PARAM: string) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 13},
			},
		},

		// Strict camelCase violation (consecutive uppercase)
		{
			Code: `const myID = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"strictCamelCase"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 7},
			},
		},

		// Multiple violations in one file
		{
			Code: "const my_var = 1;\nfunction MyFunc() {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 7},
				{MessageId: "doesNotMatchFormat", Line: 2, Column: 10},
			},
		},
	})
}
