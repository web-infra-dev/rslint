// TestCamelcaseExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch, Dimension 4 row, or tsgo AST quirk it covers; the full
// upstream migration lives in camelcase_upstream_test.go.
package camelcase

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCamelcaseExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&CamelcaseRule,
		[]rule_tester.ValidTestCase{
			// ---- Regression: upstream's MemberExpression target helper excludes updates and loop headers ----
			{Code: `obj.snake_case++; ++obj.other_case; for (obj.third_case in source); for (obj.fourth_case of source);`},
			// ---- WRAP-03/WRAP-05: TypeScript wrappers stop member assignment-target classification ----
			{Code: `[obj.snake_case!] = source; [obj.other_case as any] = source;`},
			// ---- AST-16/VIS-01: lowercase intrinsic JSX tags are not runtime references ----
			{Code: `const element = <snake_case></snake_case>`, Tsx: true},
			// ---- Dimension 4: parentheses are transparent to direct-call exclusions ----
			{Code: `(snake_case)()`},
			// ---- Dimension 4: optional-chain read-only property ----
			{Code: `source?.snake_case`},
			// ---- Dimension 4: string, numeric, template, and computed-static keys are not Identifier keys ----
			{Code: "const value = { \"snake_case\": 1, 0: 2, [\"other_case\"]: 3, [`template_case`]: 4 }"},
			// ---- Dimension 4: element access is outside the dotted-property listener ----
			{Code: `target["snake_case"] = 1`},
			// ---- Dimension 4: TS signature-only declarations have no upstream runtime listener ----
			{Code: `declare function snake_case(value_name: number): void; interface Shape { property_name: string }`},
			// ---- A configured global value in a type query honors ignoreGlobals ----
			{Code: `let value: typeof configured_global`, Globals: map[string]any{"configured_global": "readonly"}, Options: map[string]any{"ignoreGlobals": true}},
			// ---- Dimension 4: body-less abstract members are TS-only declaration forms ----
			{Code: `abstract class Shape { abstract propertyName: string; abstract methodName(value_name: string): void }`},
			// ---- tsgo collapses TSAbstract* and AccessorProperty into ordinary class-member kinds ----
			{Code: `abstract class Shape {
				abstract property_name: string
				abstract method_name(value_name: string): void
				abstract get getter_name(): string
				abstract set setter_name(value_name: string)
			}
			class Box { accessor property_name = 1; accessor #private_name = 2 }`},
			// ---- Dimension 4: empty/spread containers degrade without masking siblings ----
			{Code: `const empty = {}; const copy = { ...source }; const [] = list; const {} = source`},
			// ---- Dimension 4: private read uses do not create a second property report ----
			{Code: `class Box { #snakeCase; read() { return this.#snakeCase } }`},
			// ---- Locks in isUnderscored() leading/trailing trim and uppercase arm ----
			{Code: `const __snakeCase__ = 1; const API_RESPONSE_CODE = 2`},
			// ---- Locks in isAllowed() exact-match arm ----
			{Code: `const snake_case = 1`, Options: map[string]any{"allow": []any{"snake_case"}}},
			// ---- Locks in isAllowed() RegExp arm, including JavaScript lookahead ----
			{Code: `const snake_case = 1`, Options: map[string]any{"allow": []any{`^(?=snake_)snake_case$`}}},
			// ---- Locks in equalsToOriginalName() ImportSpecifier arm ----
			{Code: `import { snake_case } from "pkg"`, Options: map[string]any{"ignoreImports": true}},
			// ---- Real-user: eslint/eslint#15572 shorthand reuse with both exemptions ----
			{Code: `var { some_property } = obj; doSomething({ some_property });`, Options: map[string]any{"properties": "never", "ignoreDestructuring": true}},
			// ---- Real-user: eslint/eslint#16028 PascalCase has no internal underscore ----
			{Code: `const FooBar = "hello"`},
			// ---- Real-user: eslint/eslint#16604 external dotted names can be opted out ----
			{Code: `this.data.nested.variable_from_backend = "value"`, Options: map[string]any{"properties": "never"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Type queries retain their leftmost value reference only ----
			{
				Code: `let value: typeof snake_case; let other: typeof namespace_case.member_name;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notCamelCase", Line: 1, Column: 19, EndLine: 1, EndColumn: 29},
					{MessageId: "notCamelCase", Line: 1, Column: 49, EndLine: 1, EndColumn: 63},
				},
			},
			// ---- AST-03/RANGE-01: an implementation reports the first overload identifier ----
			{
				Code: `function snake_case(value_name: string): void; function snake_case(value_name: string) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notCamelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 20},
					{MessageId: "notCamelCase", Line: 1, Column: 68, EndLine: 1, EndColumn: 86},
				},
			},
			// ---- Dimension 4: parenthesized declaration initializer remains a reference ----
			{
				Code:   `const value = (snake_case)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Message: "Identifier 'snake_case' is not in camel case.", Line: 1, Column: 16, EndLine: 1, EndColumn: 26}},
			},
			// ---- Dimension 4: computed dynamic keys are runtime references ----
			{
				Code:   `const value = { [snake_case]: 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 1, Column: 18, EndLine: 1, EndColumn: 28}},
			},
			// ---- Dimension 4: function declaration/expression/arrow/method containers ----
			{
				Code: `function snake_case() {} const f = function other_case() {}; const g = (value_name) => value_name; const o = { m(method_arg) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notCamelCase", Line: 1, Column: 10, EndLine: 1, EndColumn: 20},
					{MessageId: "notCamelCase", Line: 1, Column: 45, EndLine: 1, EndColumn: 55},
					{MessageId: "notCamelCase", Line: 1, Column: 73, EndLine: 1, EndColumn: 83},
					{MessageId: "notCamelCase", Line: 1, Column: 88, EndLine: 1, EndColumn: 98},
					{MessageId: "notCamelCase", Line: 1, Column: 114, EndLine: 1, EndColumn: 124},
				},
			},
			// ---- Dimension 4: class declaration/expression and nested scopes ----
			{
				Code: `class outer_case { method() { class inner_case {} } } const C = class expression_case {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notCamelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 17},
					{MessageId: "notCamelCase", Line: 1, Column: 37, EndLine: 1, EndColumn: 47},
					{MessageId: "notCamelCase", Line: 1, Column: 71, EndLine: 1, EndColumn: 86},
				},
			},
			// ---- Dimension 4: TS assertion wrappers are not transparent call children ----
			{
				Code:   `consume(snake_case as string)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 1, Column: 9, EndLine: 1, EndColumn: 19}},
			},
			// ---- Dimension 4: object property multi-line range ----
			{
				Code:   "const value = {\n  property_name: 1,\n}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 2, Column: 3, EndLine: 2, EndColumn: 16}},
			},
			// ---- Dimension 4: private property is a separate equivalence class ----
			{
				Code:   "class Box {\n  #private_name;\n}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCasePrivate", Message: "#private_name is not in camel case.", Line: 2, Column: 3, EndLine: 2, EndColumn: 16}},
			},
			// ---- Ambient class members remain PropertyDefinition/MethodDefinition upstream ----
			{
				Code: `declare class Box { property_name: string; method_name(): void }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notCamelCase", Line: 1, Column: 21, EndLine: 1, EndColumn: 34},
					{MessageId: "notCamelCase", Line: 1, Column: 44, EndLine: 1, EndColumn: 55},
				},
			},
			// ---- Dimension 4: static identifier and dynamic string keys do not pair ----
			{
				Code: `const property_name = "snake_case"; const value = { [property_name]: 1, ["property_name"]: 2 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notCamelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 20},
					{MessageId: "notCamelCase", Line: 1, Column: 54, EndLine: 1, EndColumn: 67},
				},
			},
			// ---- Locks in isAssignmentTarget() AssignmentExpression arm ----
			{
				Code:   `target.property_name = 1`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}},
			},
			// ---- Locks in isAssignmentTarget() Property/ObjectPattern arm ----
			{
				Code:   `({ value: target.property_name } = source)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 1, Column: 18, EndLine: 1, EndColumn: 31}},
			},
			// ---- Locks in isAssignmentTarget() ArrayPattern and RestElement arms ----
			{
				Code: `[target.first_name, ...target.rest_name] = source`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notCamelCase", Line: 1, Column: 9, EndLine: 1, EndColumn: 19},
					{MessageId: "notCamelCase", Line: 1, Column: 31, EndLine: 1, EndColumn: 40},
				},
			},
			// ---- Locks in equalsToOriginalName() object alias mismatch ----
			{
				Code:    `const { source_name: local_name } = source`,
				Options: map[string]any{"ignoreDestructuring": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 1, Column: 22, EndLine: 1, EndColumn: 32}},
			},
			// ---- Locks in reportReferenceId() default-value direct-child exclusion boundary ----
			{
				Code: `const source_name = 1; const { value = source_name + 1 } = input`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notCamelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 18},
					{MessageId: "notCamelCase", Line: 1, Column: 40, EndLine: 1, EndColumn: 51},
				},
			},
			// ---- Locks in Program() configured-global branch when ignoreGlobals is false ----
			{
				Code:    `configured_global`,
				Globals: map[string]any{"configured_global": "readonly"},
				Options: map[string]any{"ignoreGlobals": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 1, Column: 1, EndLine: 1, EndColumn: 18}},
			},
			// ---- Locks in Program() through-reference branch despite ignoreGlobals ----
			{
				Code:    `undefined_global`,
				Options: map[string]any{"ignoreGlobals": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}},
			},
			// ---- Locks in property listener default and explicit-empty option equivalence ----
			{
				Code:   `const value = { property_name: 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 1, Column: 17, EndLine: 1, EndColumn: 30}},
			},
			{
				Code:    `const value = { property_name: 1 }`,
				Options: map[string]any{},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 1, Column: 17, EndLine: 1, EndColumn: 30}},
			},
			// ---- Locks in properties:"never" exempting only the key half of shorthand ----
			{
				Code:    `const some_property = 1; const value = { some_property }`,
				Options: map[string]any{"properties": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notCamelCase", Line: 1, Column: 7, EndLine: 1, EndColumn: 20},
					{MessageId: "notCamelCase", Line: 1, Column: 42, EndLine: 1, EndColumn: 55},
				},
			},
			// ---- Locks in re-export listener ----
			{
				Code:   `export { value as exported_name }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 1, Column: 19, EndLine: 1, EndColumn: 32}},
			},
			// ---- Locks in all three label listener arms ----
			{
				Code: `outer_label: for (;;) { if (stop) break outer_label; continue outer_label }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notCamelCase", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
					{MessageId: "notCamelCase", Line: 1, Column: 41, EndLine: 1, EndColumn: 52},
					{MessageId: "notCamelCase", Line: 1, Column: 63, EndLine: 1, EndColumn: 74},
				},
			},
			// ---- Real-user: eslint/eslint#16604 default property policy reports the final write ----
			{
				Code:   `this.data.nested.variable_from_backend = "value"`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notCamelCase", Line: 1, Column: 18, EndLine: 1, EndColumn: 39}},
			},
		},
	)
}

func TestCamelcaseAllowSchemaRejectsInvalidRegexp(t *testing.T) {
	invalid := []any{map[string]any{"allow": []any{"["}}}
	if err := CamelcaseRule.Schema.Validate(invalid); err == nil {
		t.Fatal("expected an invalid allow regexp to fail schema validation")
	}

	valid := []any{map[string]any{"allow": []any{`^(?=snake_)snake_case$`}}}
	if err := CamelcaseRule.Schema.Validate(valid); err != nil {
		t.Fatalf("expected a valid JavaScript regexp to pass schema validation: %v", err)
	}
}
