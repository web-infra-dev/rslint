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

		// Regression: parameter names inside type-only signatures create no
		// binding, so the "parameter" selector must not validate them — matches
		// the real-world false positive where `System` in
		// `as (System: any, ...globalArgs: any[]) => void` (a cast, not a real
		// function declaration) was incorrectly flagged.
		{
			Code: `const evaluateThunk = (0, eval)("") as (System: any, ...globalArgs: any[]) => void;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "parameter",
					"format":   []interface{}{"camelCase"},
				},
			},
		},
		{
			Code: `type Cast = (System: any) => void;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "parameter",
					"format":   []interface{}{"camelCase"},
				},
			},
		},
		{
			Code: `interface Foo { method(System: any): void; }`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "parameter",
					"format":   []interface{}{"camelCase"},
				},
			},
		},
		{
			Code: `interface Foo { (System: any): void; }`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "parameter",
					"format":   []interface{}{"camelCase"},
				},
			},
		},
		{
			Code: `interface Foo { new (System: any): void; }`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "parameter",
					"format":   []interface{}{"camelCase"},
				},
			},
		},
		{
			Code: `interface Foo { [System: string]: any; }`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "parameter",
					"format":   []interface{}{"camelCase"},
				},
			},
		},

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

		// Anonymous class expression: the `class` selector must not infer a
		// name from the enclosing variable declarator (real
		// @typescript-eslint/naming-convention only checks the class's own
		// identifier, which is absent here; `clz` is a variable name, not a
		// class name).
		{Code: `const clz = class extends Object {};`},

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

		// Unused modifier never applies to members: a type property read via
		// property access must keep the base format, and even an unreferenced
		// member stays on the base format (typescript-eslint only computes
		// `unused` for scope-participating declarations).
		{
			Code: "type foo_type = {\n  used_prop: string;\n  extra_prop: number;\n};\nexport declare const foo_var: foo_type;\nexport const prop_value = foo_var.used_prop;",
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

		// A `filter` on an earlier, less-specific selector (memberLike) must
		// not defeat a later, more specific selector (property) that has no
		// filter at all — the more specific selector should still win.
		{
			Code: `interface Foo { Uppercase: string; } const obj = { Lowercase: 1 };`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "memberLike",
					"format":   []interface{}{"camelCase"},
					"filter": map[string]interface{}{
						"regex": "^([0-9]+|[A-Za-z]+_[A-Za-z]+)$",
						"match": false,
					},
				},
				map[string]interface{}{
					"selector": "property",
					"format":   nil,
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Modifier-specific selectors must still distinguish explicit private
		// and protected members from the implicit public default when modifier
		// computation is limited to the bits needed by this config.
		{
			Code: "class MyClass {\n  private privateValue = 1;\n  protected ProtectedValue = 2;\n  PublicMethod() {}\n}",
			Options: []interface{}{
				map[string]interface{}{
					"format":            []interface{}{"camelCase"},
					"leadingUnderscore": "require",
					"modifiers":         []interface{}{"private"},
					"selector":          "memberLike",
				},
				map[string]interface{}{
					"format":            []interface{}{"camelCase"},
					"leadingUnderscore": "require",
					"modifiers":         []interface{}{"protected"},
					"selector":          "memberLike",
				},
				map[string]interface{}{
					"format":    []interface{}{"camelCase", "UPPER_CASE"},
					"modifiers": []interface{}{"public"},
					"selector":  "method",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingUnderscore", Line: 2, Column: 11},
				{MessageId: "missingUnderscore", Line: 3, Column: 13},
				{MessageId: "doesNotMatchFormat", Line: 4, Column: 3},
			},
		},

		// Reference tracking must distinguish same-named bindings in nested
		// scopes. The shorthand reads only the inner binding: it uses the base
		// snake_case config and fails, while the outer binding is unused and
		// correctly matches the PascalCase override.
		{
			Code: "function use_value() {\n  const SameName = 1;\n  return { SameName };\n}\nuse_value();\nconst SameName = 2;",
			Options: []interface{}{
				map[string]interface{}{
					"format":   []interface{}{"snake_case"},
					"selector": "variable",
				},
				map[string]interface{}{
					"format":    []interface{}{"PascalCase"},
					"modifiers": []interface{}{"unused"},
					"selector":  "variable",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 2, Column: 9},
			},
		},

		// Members never get the `unused` modifier: an unreferenced type
		// property is still checked against the base format, not the
		// unused override.
		{
			Code: "export type foo_type = {\n  ExtraProp: string;\n};",
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 2, Column: 3},
			},
		},

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

		// Named class expression violating PascalCase: the class's own
		// identifier is still checked by the `class` selector.
		{
			Code: `const x = class myClass {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat", Line: 1, Column: 17},
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

		// requireDouble on a lone `_`: stripping `__` fails because only one
		// underscore exists, so missingUnderscore fires.
		{
			Code: `const _ = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector":          "variable",
					"format":            []interface{}{"camelCase"},
					"leadingUnderscore": "requireDouble",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingUnderscore", Line: 1, Column: 7},
			},
		},
	})
}

// TestNamingConventionRegexSchema locks in that both regex-valued options are
// validated as regexes: `filter`'s bare-string form and the `regex` key of
// its object form (shared by `custom`). The rule compiles each one, so an
// unparsable pattern would otherwise be dropped and the selector would apply
// to everything. Upstream's schema declares bare strings.
func TestNamingConventionRegexSchema(t *testing.T) {
	selector := func(filter any) []any {
		return []any{map[string]any{
			"selector": "variable",
			"format":   []any{"camelCase"},
			"filter":   filter,
		}}
	}
	for name, filter := range map[string]any{
		"bare string":  "(",
		"regex object": map[string]any{"match": true, "regex": "("},
	} {
		if err := NamingConventionRule.Schema.Validate(selector(filter)); err == nil {
			t.Errorf("expected an invalid filter %s to fail schema validation", name)
		}
	}
	for name, filter := range map[string]any{
		"bare string":  "^ignored",
		"regex object": map[string]any{"match": true, "regex": "^ignored"},
	} {
		if err := NamingConventionRule.Schema.Validate(selector(filter)); err != nil {
			t.Errorf("expected a valid filter %s to pass schema validation, got: %v", name, err)
		}
	}
}

// TestNamingConventionCustomLookahead locks in that `custom.regex` (and by
// extension `filter`, which shares the same compile path) is matched with
// the same ECMAScript regex engine the schema validates it against
// (regexp2), not Go's RE2. RE2 cannot compile lookahead, so a JS-only
// pattern that passes schema validation would otherwise silently fail to
// compile and never flag a mismatched name.
func TestNamingConventionCustomLookahead(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{
			Code: `const myVar = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"camelCase"},
					"custom": map[string]interface{}{
						"regex": "^(?!deprecated).*$",
						"match": true,
					},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `const deprecatedVar = 1;`,
			Options: []interface{}{
				map[string]interface{}{
					"selector": "variable",
					"format":   []interface{}{"camelCase"},
					"custom": map[string]interface{}{
						"regex": "^(?!deprecated).*$",
						"match": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "satisfyCustom", Line: 1, Column: 7},
			},
		},
	})
}
