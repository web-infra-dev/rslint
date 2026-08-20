package id_match

// TestIdMatchExtrasBranches locks in every reachable branch of the upstream
// rule source, including the ones upstream itself never tests. Each case names
// the branch it pins. Its siblings are id_match_extras_dim4_test.go and
// id_match_extras_realuser_test.go.

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestIdMatchExtrasBranches(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&IdMatchRule,
		[]rule_tester.ValidTestCase{
			// ---- Locks in upstream Identifier() arm 1: a declared global is left alone ----
			{
				Code:    `const a = Object.keys(b);`,
				Options: []any{`^[a-z]+$`, map[string]any{"properties": true}},
			},
			// ---- Locks in upstream isImportAttributeKey() arm 1: static import attribute ----
			{
				Code:    `import foo from 'm' with { type_x: 'json' };`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Locks in upstream isImportAttributeKey() arm 2: dynamic import options ----
			{
				Code:    `import('m', { with: { type_x: 'json' } });`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Locks in upstream isImportAttributeKey() arm 2: shorthand attribute key ----
			{
				Code:    `import('m', { with: { type_x } });`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Locks in upstream isImportAttributeKey() arm 2: string-literal outer key ----
			{
				Code:    `import('m', { 'with': { type_x: 'json' } });`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Locks in upstream Identifier() arm 3: MetaProperty ----
			{
				Code:    `function f() { return new.target; }`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Locks in upstream Identifier() MemberExpression arm 0: properties off short-circuits ----
			{
				Code:    `a_1.b_1 = c_1.d_1;`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Locks in upstream Identifier() MemberExpression arm 2: an update expression does not ----
			{
				Code:    `obj.a_1++;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Locks in upstream Identifier() MemberExpression arm 2: an equality comparison does not ----
			{
				Code:    `obj.a_1 === 1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Locks in upstream Identifier() MemberExpression arm 3: right side is a member access ----
			{
				Code:    `obj.x = obj.y_1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Locks in upstream Identifier() MemberExpression arm 3: nested receiver is never assigned ----
			{
				Code:    `a.b_1.c = 1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Locks in upstream Identifier() Property arm: object-literal key needs properties ----
			{
				Code:    `const o = { a_1: 1 };`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Locks in upstream Identifier() Property arm: object-literal value needs properties ----
			{
				Code:    `const o = { k: a_1 };`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Locks in upstream Identifier() Property arm: object-literal value is not a declaration ----
			{
				Code:    `const o = { k: a_1 };`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true, "onlyDeclarations": true}},
			},
			// ---- Locks in upstream Identifier() ObjectPattern arm 1: ignoreDestructuring silences it ----
			{
				Code:    `const { a_1 = 1 } = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"ignoreDestructuring": true}},
			},
			// ---- Locks in upstream Identifier() ObjectPattern arm 2: the renamed-from key is skipped ----
			{
				Code:    `const { a_1: b } = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Locks in upstream Identifier() ObjectPattern arm 3: ignoreDestructuring keeps both quiet ----
			{
				Code:    `const { a_1: a_1 } = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true, "ignoreDestructuring": true}},
			},
			// ---- Locks in upstream Identifier() ObjectPattern arm 3: a nested pattern value has no name ----
			{
				Code:    `const { a_1: { b } } = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Locks in upstream Identifier() Property arm 5: ignoreDestructuring inside a pattern ----
			{
				Code:    `const [{ a_1 }] = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true, "ignoreDestructuring": true}},
			},
			// ---- Locks in upstream Identifier() AssignmentPattern arm: parameter default needs properties ----
			{
				Code:    `function f(a_1 = 1) {}`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Locks in upstream Identifier() AssignmentPattern arm: onlyDeclarations still applies ----
			{
				Code:    `function f(a_1 = 1) {}`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true, "onlyDeclarations": true}},
			},
			// ---- Locks in upstream Identifier() import arm: the exported name is not the local one ----
			{
				Code:    `import { a_1 as b } from 'm';`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Locks in upstream Identifier() PropertyDefinition arm: classFields off ----
			{
				Code:    `class C { a_1 = 1; }`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Locks in upstream Identifier() PropertyDefinition arm: the initializer is a field too ----
			{
				Code:    `class C { a = b_1; }`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Locks in upstream shouldReport() arm 1: an arrow parameter is not ----
			{
				Code:    `const g = (a_1) => a_1;`,
				Options: []any{`^[^_]+$`, map[string]any{"onlyDeclarations": true}},
			},
			// ---- Locks in upstream shouldReport() arm 1: a TS parameter property is not ----
			{
				Code:    `class C { constructor(private a_1) {} }`,
				Options: []any{`^[^_]+$`, map[string]any{"onlyDeclarations": true}},
			},
			// ---- Locks in upstream shouldReport() arm 2: a call argument is allowed ----
			{
				Code:    `f(a_1);`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Locks in upstream shouldReport() arm 2: a call callee is allowed ----
			{
				Code:    `a_1();`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Locks in upstream shouldReport() arm 2: a new callee is allowed ----
			{
				Code:    `new a_1();`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Locks in upstream PrivateIdentifier() arm 1: a private field obeys classFields ----
			{
				Code:    `class C { #a_1 = 1; }`,
				Options: []any{`^[^_]+$`},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Locks in upstream Identifier() arm 1: a shadowed global is not left alone ----
			{
				Code:    `function f(Object) { return Object.keys(b); }`,
				Options: []any{`^[a-z]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Object' does not match the pattern '^[a-z]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 18,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Object' does not match the pattern '^[a-z]+$'.`,
						Line:      1,
						Column:    29,
						EndLine:   1,
						EndColumn: 35,
					},
				},
			},
			// ---- Locks in upstream Identifier() arm 1: a property named like a global is not a reference ----
			{
				Code:    `foo.Object = 1;`,
				Options: []any{`^[a-z]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Object' does not match the pattern '^[a-z]+$'.`,
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 11,
					},
				},
			},
			// ---- Locks in upstream isImportAttributeKey() arm 2: a computed key is not an attribute key ----
			{
				Code:    `import('m', { with: { [type_x]: 'json' } });`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'type_x' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    24,
						EndLine:   1,
						EndColumn: 30,
					},
				},
			},
			// ---- Locks in upstream isImportAttributeKey() arm 2: only the second argument is options ----
			{
				Code:    `import({ with: { type_x: 'json' } });`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'type_x' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    18,
						EndLine:   1,
						EndColumn: 24,
					},
				},
			},
			// ---- Locks in upstream isImportAttributeKey() arm 2: a plain call is not an import ----
			{
				Code:    `call('m', { with: { type_x: 'json' } });`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'type_x' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 27,
					},
				},
			},
			// ---- Locks in upstream Identifier() MemberExpression arm 1: object named like the property ----
			{
				Code:    `a_1.a_1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 4,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 8,
					},
				},
			},
			// ---- Locks in upstream Identifier() MemberExpression arm 2: assigned member property ----
			{
				Code:    `obj.a_1 = 1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 8,
					},
				},
			},
			// ---- Locks in upstream Identifier() MemberExpression arm 2: compound assignment counts ----
			{
				Code:    `obj.a_1 += 1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 8,
					},
				},
			},
			// ---- Locks in upstream Identifier() MemberExpression arm 2: logical assignment counts ----
			{
				Code:    `obj.a_1 ||= 1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 8,
					},
				},
			},
			// ---- Locks in upstream Identifier() MemberExpression arm 3: right side is not a member access ----
			{
				Code:    `obj.x = y_1.z;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'y_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    9,
						EndLine:   1,
						EndColumn: 12,
					},
				},
			},
			// ---- Locks in upstream Identifier() Property arm: object-literal value with properties ----
			{
				Code:    `const o = { k: a_1 };`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    16,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			// ---- Locks in upstream Identifier() ObjectPattern arm 1: shorthand default reports the key ----
			{
				Code:    `const { a_1 = 1 } = o;`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    9,
						EndLine:   1,
						EndColumn: 12,
					},
				},
			},
			// ---- Locks in upstream Identifier() ObjectPattern arm 3: key equal to value reports both nodes ----
			{
				Code:    `const { a_1: a_1 } = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    9,
						EndLine:   1,
						EndColumn: 12,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 17,
					},
				},
			},
			// ---- Locks in upstream Identifier() ObjectPattern arm 3: a renamed binding is always reported ----
			{
				Code:    `const { a: b_1 } = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true, "ignoreDestructuring": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'b_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			// ---- Locks in upstream Identifier() Property arm 4: computed keys bypass the properties gate ----
			{
				Code:    `const o = { [a_1]: 1 };`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 17,
					},
				},
			},
			// ---- Locks in upstream Identifier() Property arm 5: an array pattern is not an object pattern ----
			{
				Code:    `const [a_1] = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true, "ignoreDestructuring": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    8,
						EndLine:   1,
						EndColumn: 11,
					},
				},
			},
			// ---- Locks in upstream Identifier() Property arm 6: the default value itself is not re-reported ----
			{
				Code:    `const b_1 = 1; const { a = b_1 } = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'b_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    7,
						EndLine:   1,
						EndColumn: 10,
					},
				},
			},
			// ---- Locks in upstream Identifier() AssignmentPattern arm: parameter default with properties ----
			{
				Code:    `function f(a_1 = 1) {}`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			// ---- Locks in upstream Identifier() AssignmentPattern arm: array pattern default ----
			{
				Code:    `const [a_1 = 1] = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    8,
						EndLine:   1,
						EndColumn: 11,
					},
				},
			},
			// ---- Locks in upstream Identifier() AssignmentPattern arm: destructuring assignment default ----
			{
				Code:    `[a_1 = 1] = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    2,
						EndLine:   1,
						EndColumn: 5,
					},
				},
			},
			// ---- Locks in upstream Identifier() import arm: default import binding ----
			{
				Code:    `import a_1 from 'm';`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    8,
						EndLine:   1,
						EndColumn: 11,
					},
				},
			},
			// ---- Locks in upstream Identifier() import arm: namespace import binding ----
			{
				Code:    `import * as a_1 from 'm';`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 16,
					},
				},
			},
			// ---- Locks in upstream Identifier() import arm: a type-only import still binds a name ----
			{
				Code:    `import { type a_1 } from 'm';`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    15,
						EndLine:   1,
						EndColumn: 18,
					},
				},
			},
			// ---- Locks in upstream Identifier() import arm: export specifiers are not import specifiers ----
			{
				Code: `const a_1 = 1;
export { a_1 as b_1 };`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    7,
						EndLine:   1,
						EndColumn: 10,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      2,
						Column:    10,
						EndLine:   2,
						EndColumn: 13,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'b_1' does not match the pattern '^[^_]+$'.`,
						Line:      2,
						Column:    17,
						EndLine:   2,
						EndColumn: 20,
					},
				},
			},
			// ---- Locks in upstream Identifier() PropertyDefinition arm: classFields on ----
			{
				Code:    `class C { a_1 = 1; }`,
				Options: []any{`^[^_]+$`, map[string]any{"classFields": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 14,
					},
				},
			},
			// ---- Locks in upstream Identifier() PropertyDefinition arm: the initializer is a field too ----
			{
				Code:    `class C { a = b_1; }`,
				Options: []any{`^[^_]+$`, map[string]any{"classFields": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'b_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    15,
						EndLine:   1,
						EndColumn: 18,
					},
				},
			},
			// ---- Locks in upstream Identifier() PropertyDefinition arm: an auto-accessor is not a field ----
			{
				Code:    `class C { accessor a_1 = 1; }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    20,
						EndLine:   1,
						EndColumn: 23,
					},
				},
			},
			// ---- Locks in upstream shouldReport() arm 1: onlyDeclarations keeps declarations ----
			{
				Code: `var a_1 = 1;
function b_1() {}`,
				Options: []any{`^[^_]+$`, map[string]any{"onlyDeclarations": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 8,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'b_1' does not match the pattern '^[^_]+$'.`,
						Line:      2,
						Column:    10,
						EndLine:   2,
						EndColumn: 13,
					},
				},
			},
			// ---- Locks in upstream shouldReport() arm 1: a function-declaration parameter is a declaration ----
			{
				Code:    `function f(a_1) {}`,
				Options: []any{`^[^_]+$`, map[string]any{"onlyDeclarations": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			// ---- Locks in upstream shouldReport() arm 1: a TS parameter property without onlyDeclarations ----
			{
				Code:    `class C { constructor(private a_1) {} }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    31,
						EndLine:   1,
						EndColumn: 34,
					},
				},
			},
			// ---- Locks in upstream shouldReport() arm 2: a tagged template tag is not a call ----
			{
				Code:    "a_1`x`;",
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 4,
					},
				},
			},
			// ---- Locks in upstream PrivateIdentifier() arm 2: a private method never does ----
			{
				Code:    `class C { #a_1() {} }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatchPrivate",
						Message:   `Identifier '#a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			// ---- Locks in upstream PrivateIdentifier() arm 2: a private member access never does ----
			{
				Code:    `class C { #a_1 = 1; m() { return this.#a_1; } }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatchPrivate",
						Message:   `Identifier '#a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    39,
						EndLine:   1,
						EndColumn: 43,
					},
				},
			},
			// ---- Locks in upstream PrivateIdentifier() arm 2: an auto-accessor is not a field ----
			{
				Code:    `class C { accessor #a_1 = 1; }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatchPrivate",
						Message:   `Identifier '#a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    20,
						EndLine:   1,
						EndColumn: 24,
					},
				},
			},
		},
	)
}
