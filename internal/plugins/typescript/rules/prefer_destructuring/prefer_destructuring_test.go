package prefer_destructuring

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferDestructuringRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferDestructuringRule, []rule_tester.ValidTestCase{
		// ---- type annotated (default: skip) ----
		{Code: "declare const object: { foo: string };\nvar foo: string = object.foo;"},
		{Code: "declare const array: number[];\nconst bar: number = array[0];"},

		// ---- enforceForDeclarationWithTypeAnnotation: true (valid cases) ----
		{
			Code: "declare const object: { foo: string };\nvar { foo } = object;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
		},
		{
			Code: "declare const object: { foo: string };\nvar { foo }: { foo: number } = object;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
		},
		{
			Code: "declare const array: number[];\nvar [foo] = array;",
			Options: []interface{}{
				map[string]interface{}{"array": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
		},
		{
			Code: "declare const array: number[];\nvar [foo]: [foo: number] = array;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
		},
		{
			// Renamed prop: var name != property name, valid even with enforcement
			Code: "declare const object: { bar: string };\nvar foo: unknown = object.bar;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
		},
		{
			Code: "declare const object: { foo: string };\nvar { foo: bar } = object;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
		},
		{
			Code: "declare const object: { foo: boolean };\nvar { foo: bar }: { foo: boolean } = object;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
		},
		{
			// super.foo — always skipped
			Code: "declare class Foo { foo: string; }\nclass Bar extends Foo {\n  static foo() {\n    var foo: any = super.foo;\n  }\n}",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
		},

		// ---- numeric property for iterable / non-iterable ----
		{Code: "let x: { 0: unknown };\nlet y = x[0];"},
		{Code: "let x: { 0: unknown };\ny = x[0];"},
		{Code: "let x: unknown;\nlet y = x[0];"},
		{Code: "let x: unknown;\ny = x[0];"},
		{Code: "let x: { 0: unknown } | unknown[];\nlet y = x[0];"},
		{Code: "let x: { 0: unknown } | unknown[];\ny = x[0];"},
		{Code: "let x: { 0: unknown } & (() => void);\nlet y = x[0];"},
		{Code: "let x: { 0: unknown } & (() => void);\ny = x[0];"},
		{Code: "let x: Record<number, unknown>;\nlet y = x[0];"},
		{Code: "let x: Record<number, unknown>;\ny = x[0];"},
		{
			Code: "let x: { 0: unknown };\nlet { 0: y } = x;",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code: "let x: { 0: unknown };\n({ 0: y } = x);",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			// Non-iterable + only array enabled → valid
			Code: "let x: { 0: unknown };\nlet y = x[0];",
			Options: []interface{}{
				map[string]interface{}{"array": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code: "let x: { 0: unknown };\ny = x[0];",
			Options: []interface{}{
				map[string]interface{}{"array": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			// Per-context: VarDeclarator object=false → valid for let
			Code: "let x: { 0: unknown };\nlet y = x[0];",
			Options: []interface{}{
				map[string]interface{}{
					"AssignmentExpression": map[string]interface{}{"array": true, "object": true},
					"VariableDeclarator":   map[string]interface{}{"array": true, "object": false},
				},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			// Per-context: AssignExpr object=false → valid for assignment
			Code: "let x: { 0: unknown };\ny = x[0];",
			Options: []interface{}{
				map[string]interface{}{
					"AssignmentExpression": map[string]interface{}{"array": true, "object": false},
					"VariableDeclarator":   map[string]interface{}{"array": true, "object": true},
				},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		// Non-integer-literal index: x[i] where i is a variable → not array literal access
		{
			Code: "let x: Record<number, unknown>;\nlet i: number = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": false},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code: "let x: Record<number, unknown>;\nlet i: 0 = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": false},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code: "let x: Record<number, unknown>;\nlet i: 0 | 1 | 2 = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": false},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code: "let x: unknown[];\nlet i: number = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": false},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code: "let x: unknown[];\nlet i: 0 = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": false},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code: "let x: unknown[];\nlet i: 0 | 1 | 2 = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": false},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code: "let x: unknown[];\nlet i: number = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": false},
			},
		},
		{
			// Compound assignment operators should not be reported
			Code: "let x: { 0: unknown };\ny += x[0];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			// super[0] — always skipped
			Code: "class Bar { public [0]: unknown; }\nclass Foo extends Bar {\n  static foo() { let y = super[0]; }\n}",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code: "class Bar { public [0]: unknown; }\nclass Foo extends Bar {\n  static foo() { y = super[0]; }\n}",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},

		// ---- already destructured ----
		{Code: "let xs: unknown[] = [1];\nlet [x] = xs;"},
		{Code: "const obj: { x: unknown } = { x: 1 };\nconst { x } = obj;"},
		{Code: "var obj: { x: unknown } = { x: 1 };\nvar { x: y } = obj;"},
		{Code: "let obj: { x: unknown } = { x: 1 };\nlet key: 'x' = 'x';\nlet { [key]: foo } = obj;"},
		{Code: "const obj: { x: unknown } = { x: 1 };\nlet x: unknown;\n({ x } = obj);"},

		// ---- valid unless enforceForRenamedProperties is true ----
		{Code: "let obj: { x: unknown } = { x: 1 };\nlet y = obj.x;"},
		{Code: "var obj: { x: unknown } = { x: 1 };\nvar y: unknown;\ny = obj.x;"},
		{Code: "const obj: { x: unknown } = { x: 1 };\nconst y = obj['x'];"},
		{Code: "let obj: Record<string, unknown> = {};\nlet key = 'abc';\nvar y = obj[key];"},

		// ---- shorthand operators (should NOT be reported) ----
		{Code: "let obj: { x: number } = { x: 1 };\nlet x = 10;\nx += obj.x;"},
		{Code: "let obj: { x: boolean } = { x: false };\nlet x = true;\nx ||= obj.x;"},
		{Code: "const xs: number[] = [1];\nlet x = 3;\nx *= xs[0];"},
		// &&= and ??= operators
		{Code: "let obj: { x: number } = { x: 1 };\nlet x = 10;\nx &&= obj.x;"},
		{Code: "let obj: Record<string, number> = { foo: 1 };\nlet x = 10;\nx ??= obj['foo'];"},

		// ---- optional chaining (should NOT be reported) ----
		{Code: "let xs: unknown[] | undefined;\nlet x = xs?.[0];"},
		{Code: "let obj: Record<string, unknown> | undefined;\nlet x = obj?.x;"},

		// ---- private identifiers ----
		{Code: "class C { #foo: string = ''; method() { const foo: unknown = this.#foo; } }"},
		{Code: "class C { #foo: string = ''; method() { let foo: unknown; foo = this.#foo; } }"},
		{
			Code: "class C { #foo: string = ''; method() { const bar: unknown = this.#foo; } }",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
		},
		{
			Code: "class C { #foo: string = ''; method(another: C) { let bar: unknown; bar = another.#foo; } }",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
		},
		{
			Code: "class C { #foo: string = ''; method() { const foo: unknown = this.#foo; } }",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
		},

		// ---- ESLint core valid cases ----
		{Code: "var [foo] = [1];"},
		{Code: "var { foo } = { foo: 1 };"},
		{Code: "var foo: any;"},
		{
			Code: "var foo = ({ bar: 1 } as any).bar;",
			Options: []interface{}{
				map[string]interface{}{"VariableDeclarator": map[string]interface{}{"object": true}},
			},
		},
		{
			Code:    "var foo = ({ bar: 1 } as any).bar;",
			Options: []interface{}{map[string]interface{}{"object": true}},
		},
		{
			Code: "var foo = ({ bar: 1 } as any).bar;",
			Options: []interface{}{
				map[string]interface{}{"VariableDeclarator": map[string]interface{}{"object": true}},
				map[string]interface{}{"enforceForRenamedProperties": false},
			},
		},
		{
			Code: "var foo = ({ bar: 1 } as any).bar;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": false},
			},
		},
		{
			Code:    "var { bar: foo } = { bar: 1 } as any;",
			Options: []interface{}{map[string]interface{}{"object": true}, map[string]interface{}{"enforceForRenamedProperties": true}},
		},
		{
			Code:    "var { [bar]: foo } = { bar: 1 } as any;",
			Options: []interface{}{map[string]interface{}{"object": true}, map[string]interface{}{"enforceForRenamedProperties": true}},
		},
		{
			Code:    "var foo = [1][0];",
			Options: []interface{}{map[string]interface{}{"VariableDeclarator": map[string]interface{}{"array": false}}},
		},
		{
			Code:    "var foo = [1][0];",
			Options: []interface{}{map[string]interface{}{"array": false}},
		},
		{
			Code: "var foo = ({ foo: 1 } as any).foo;",
			Options: []interface{}{
				map[string]interface{}{"VariableDeclarator": map[string]interface{}{"object": false}},
			},
		},
		{Code: "({ foo } = { foo: 1 } as any);"},
		{
			Code: "var foo = [1][0];",
			Options: []interface{}{
				map[string]interface{}{"VariableDeclarator": map[string]interface{}{"array": false}},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code:    "var foo = [1][0];",
			Options: []interface{}{map[string]interface{}{"array": false}, map[string]interface{}{"enforceForRenamedProperties": true}},
		},
		{Code: "[foo] = [1];"},
		{Code: "foo += [1][0]"},
		{Code: "foo += ({ foo: 1 } as any).foo"},
		{
			Code: "foo = ({ foo: 1 } as any).foo;",
			Options: []interface{}{
				map[string]interface{}{"AssignmentExpression": map[string]interface{}{"object": false}},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code: "foo = ({ foo: 1 } as any).foo;",
			Options: []interface{}{
				map[string]interface{}{"AssignmentExpression": map[string]interface{}{"object": false}},
				map[string]interface{}{"enforceForRenamedProperties": false},
			},
		},
		{
			Code: "foo = [1][0];",
			Options: []interface{}{
				map[string]interface{}{"AssignmentExpression": map[string]interface{}{"array": false}},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
		},
		{
			Code: "foo = [1][0];",
			Options: []interface{}{
				map[string]interface{}{"AssignmentExpression": map[string]interface{}{"array": false}},
				map[string]interface{}{"enforceForRenamedProperties": false},
			},
		},
		{
			Code: "foo = [1][0];",
			Options: []interface{}{
				map[string]interface{}{
					"VariableDeclarator":   map[string]interface{}{"array": true},
					"AssignmentExpression": map[string]interface{}{"array": false},
				},
				map[string]interface{}{"enforceForRenamedProperties": false},
			},
		},
		{
			Code: "var foo = [1][0];",
			Options: []interface{}{
				map[string]interface{}{
					"VariableDeclarator":   map[string]interface{}{"array": false},
					"AssignmentExpression": map[string]interface{}{"array": true},
				},
				map[string]interface{}{"enforceForRenamedProperties": false},
			},
		},
		{
			Code: "foo = ({ foo: 1 } as any).foo;",
			Options: []interface{}{
				map[string]interface{}{
					"VariableDeclarator":   map[string]interface{}{"object": true},
					"AssignmentExpression": map[string]interface{}{"object": false},
				},
			},
		},
		{
			Code: "var foo = ({ foo: 1 } as any).foo;",
			Options: []interface{}{
				map[string]interface{}{
					"VariableDeclarator":   map[string]interface{}{"object": false},
					"AssignmentExpression": map[string]interface{}{"object": true},
				},
			},
		},
		{Code: "class Foo extends (Object as any) { static foo() {var foo = super.foo} }"},
		{Code: "foo = ({ bar: 1 } as any)[foo];"},
		{Code: "var foo = ({ bar: 1 } as any)[foo];"},
		{
			Code:    "var {foo: {bar}} = { foo: { bar: 1 } } as any;",
			Options: []interface{}{map[string]interface{}{"object": true}},
		},
		{
			Code:    "var {bar} = ({ foo: { bar: 1 } } as any).foo;",
			Options: []interface{}{map[string]interface{}{"object": true}},
		},

		// ---- optional chaining (ESLint core) ----
		{Code: "var foo = [1]?.[0];"},
		{Code: "var foo = ({ foo: 1 } as any)?.foo;"},

		// ---- private identifiers (ESLint core) ----
		{Code: "class C { #x: number = 0; foo() { const x = this.#x; } }"},
		{Code: "class C { #x: number = 0; foo() { x = this.#x; } }"},
		{Code: "class C { #x: number = 0; foo(a: C) { x = a.#x; } }"},
		{
			Code:    "class C { #x: number = 0; foo() { const x = this.#x; } }",
			Options: []interface{}{map[string]interface{}{"array": true, "object": true}, map[string]interface{}{"enforceForRenamedProperties": true}},
		},
		{
			Code:    "class C { #x: number = 0; foo() { const y = this.#x; } }",
			Options: []interface{}{map[string]interface{}{"array": true, "object": true}, map[string]interface{}{"enforceForRenamedProperties": true}},
		},
		{
			Code:    "class C { #x: number = 0; foo() { x = this.#x; } }",
			Options: []interface{}{map[string]interface{}{"array": true, "object": true}, map[string]interface{}{"enforceForRenamedProperties": true}},
		},
		{
			Code:    "class C { #x: number = 0; foo() { y = this.#x; } }",
			Options: []interface{}{map[string]interface{}{"array": true, "object": true}, map[string]interface{}{"enforceForRenamedProperties": true}},
		},
		{
			Code:    "class C { #x: number = 0; foo(a: C) { x = a.#x; } }",
			Options: []interface{}{map[string]interface{}{"array": true, "object": true}, map[string]interface{}{"enforceForRenamedProperties": true}},
		},
		{
			Code:    "class C { #x: number = 0; foo(a: C) { y = a.#x; } }",
			Options: []interface{}{map[string]interface{}{"array": true, "object": true}, map[string]interface{}{"enforceForRenamedProperties": true}},
		},
		{
			Code:    "class C { #x: number = 0; foo() { x = this.a.#x; } }",
			Options: []interface{}{map[string]interface{}{"array": true, "object": true}, map[string]interface{}{"enforceForRenamedProperties": true}},
		},

		// ---- ESLint core: computed bracket access, renamed not enforced ----
		{
			Code: "let object: Record<string, unknown> = {};\nvar foo = object['bar'];",
			Options: []interface{}{
				map[string]interface{}{"VariableDeclarator": map[string]interface{}{"object": true}},
				map[string]interface{}{"enforceForRenamedProperties": false},
			},
		},
		{
			Code: "let object: Record<string, unknown> = {};\nvar foo = object[bar];",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": false},
			},
		},
		{
			// object disabled via per-context, bracket string literal
			Code: "var foo = ({ foo: 1 } as any)['foo'];",
			Options: []interface{}{
				map[string]interface{}{"VariableDeclarator": map[string]interface{}{"object": false}},
			},
		},
		// ---- &&= with array index ----
		{Code: "const xs: number[] = [1];\nlet x = 3;\nx &&= xs[0];"},

		// ---- using / await using (explicit resource management) ----
		{Code: "using foo = [1][0];"},
		{Code: "using foo = ({ foo: 1 } as any).foo;"},
		{Code: "await using foo = [1][0];"},
		{Code: "await using foo = ({ foo: 1 } as any).foo;"},
	}, []rule_tester.InvalidTestCase{
		// ---- enforceForDeclarationWithTypeAnnotation: true ----
		{
			Code: "var foo: string = object.foo;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "var foo: string = array[0];",
			Options: []interface{}{
				map[string]interface{}{"array": true},
				map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "var foo: unknown = object.bar;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{
					"enforceForDeclarationWithTypeAnnotation": true,
					"enforceForRenamedProperties":             true,
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- numeric property for iterable / non-iterable ----
		{
			Code: "let x: [1, 2, 3];\nlet y = x[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "let x: [1, 2, 3];\ny = x[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "function* it() { yield 1; }\nlet y = it()[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "function* it() { yield 1; }\ny = it()[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "let x: any;\nlet y = x[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "let x: any;\ny = x[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "let x: string[] | { [Symbol.iterator]: unknown };\nlet y = x[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "let x: string[] | { [Symbol.iterator]: unknown };\ny = x[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "let x: object & unknown[];\nlet y = x[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "let x: object & unknown[];\ny = x[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		// Non-iterable with enforceForRenamedProperties → object
		{
			Code: "let x: { 0: string };\nlet y = x[0];",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let x: { 0: string };\ny = x[0];",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let x: { 0: string };\nlet y = x[0];",
			Options: []interface{}{
				map[string]interface{}{
					"AssignmentExpression": map[string]interface{}{"array": false, "object": false},
					"VariableDeclarator":   map[string]interface{}{"array": false, "object": true},
				},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let x: { 0: string };\ny = x[0];",
			Options: []interface{}{
				map[string]interface{}{
					"AssignmentExpression": map[string]interface{}{"array": false, "object": true},
					"VariableDeclarator":   map[string]interface{}{"array": false, "object": false},
				},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		// Non-integer-literal index with enforceForRenamedProperties → object
		{
			Code: "let x: Record<number, unknown>;\nlet i: number = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let x: Record<number, unknown>;\nlet i: 0 = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let x: Record<number, unknown>;\nlet i: 0 | 1 | 2 = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let x: unknown[];\nlet i: number = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let x: unknown[];\nlet i: 0 = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let x: unknown[];\nlet i: 0 | 1 | 2 = 0;\ny = x[i];",
			Options: []interface{}{
				map[string]interface{}{"array": true, "object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			// Union of iterable + non-iterable: non-iterable member makes it object
			Code: "let x: { 0: unknown } | unknown[];\nlet y = x[0];",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let x: { 0: unknown } | unknown[];\ny = x[0];",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- auto fixes (same-name property access) ----
		{
			Code:   "let obj = { foo: 'bar' };\nconst foo = obj.foo;",
			Output: []string{"let obj = { foo: 'bar' };\nconst {foo} = obj;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code:   "const obj = { asdf: { qwer: null } };\nconst qwer = obj.asdf.qwer;",
			Output: []string{"const obj = { asdf: { qwer: null } };\nconst {qwer} = obj.asdf;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			// Comment before identifier: fix preserves it
			Code:   "const obj = { foo: 100 };\nconst /* comment */ foo = obj.foo;",
			Output: []string{"const obj = { foo: 100 };\nconst /* comment */ {foo} = obj;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		// ---- autofix: parenthesized expression cases ----
		{
			Code:   "let x: null = null; let obj = { foo: 1 };\nconst foo = (x, obj).foo;",
			Output: []string{"let x: null = null; let obj = { foo: 1 };\nconst {foo} = (x, obj);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code:   "const call = (() => null).call;",
			Output: []string{"const {call} = () => null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code:   "const obj = { foo: 'bar' };\nlet a: any;\nvar foo = (a = obj).foo;",
			Output: []string{"const obj = { foo: 'bar' };\nlet a: any;\nvar {foo} = a = obj;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			// (a || b).foo → {foo} = a || b
			Code:   "let a: any; let b: any;\nvar foo = (a || b).foo;",
			Output: []string{"let a: any; let b: any;\nvar {foo} = a || b;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			// (f()).foo → {foo} = f()
			Code:   "function f(): any { return {}; }\nvar foo = (f()).foo;",
			Output: []string{"function f(): any { return {}; }\nvar {foo} = f();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- comment-related fix/no-fix ----
		{
			// Comment after identifier before = → suppress fix
			Code: "var foo /* comment */ = ({ foo: 1 } as any).foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			// Comment between = and expression → suppress fix
			Code: "var foo = /* comment */ ({ foo: 1 } as any).foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- enforceForRenamedProperties: true ----
		{
			Code: "let obj = { foo: 'bar' };\nconst x = obj.foo;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let obj = { foo: 'bar' };\nlet x: unknown;\nx = obj.foo;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let obj: Record<string, unknown> = {};\nlet key = 'abc';\nconst x = obj[key];",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "let obj: Record<string, unknown> = {};\nlet key = 'abc';\nlet x: unknown;\nx = obj[key];",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- ESLint core invalid cases ----
		{
			Code:   "var foo = ({ foo: 1 } as any).foo;",
			Output: []string{"var {foo} = { foo: 1 } as any;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code:   "var foo = ({ bar: { foo: 1 } } as any).bar.foo;",
			Output: []string{"var {foo} = ({ bar: { foo: 1 } } as any).bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "var foobar = ({ bar: 1 } as any).bar;",
			Options: []interface{}{
				map[string]interface{}{"VariableDeclarator": map[string]interface{}{"object": true}},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "var foobar = ({ bar: 1 } as any).bar;",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "var foo = ({ bar: 1 } as any)[bar];",
			Options: []interface{}{
				map[string]interface{}{"VariableDeclarator": map[string]interface{}{"object": true}},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "var foo = ({ bar: 1 } as any)[bar];",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "var foo = ({ foo: 1 } as any)['foo'];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "foo = ({ foo: 1 } as any).foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "foo = ({ foo: 1 } as any)['foo'];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code: "foo = [1][0];",
			Options: []interface{}{
				map[string]interface{}{"AssignmentExpression": map[string]interface{}{"array": true}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "var foo = [1][0];",
			Options: []interface{}{
				map[string]interface{}{
					"VariableDeclarator":   map[string]interface{}{"array": true},
					"AssignmentExpression": map[string]interface{}{"array": false},
				},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "var foo = [1][0];",
			Options: []interface{}{
				map[string]interface{}{
					"VariableDeclarator":   map[string]interface{}{"array": true},
					"AssignmentExpression": map[string]interface{}{"array": false},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "foo = [1][0];",
			Options: []interface{}{
				map[string]interface{}{
					"VariableDeclarator":   map[string]interface{}{"array": false},
					"AssignmentExpression": map[string]interface{}{"array": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "foo = ({ foo: 1 } as any).foo;",
			Options: []interface{}{
				map[string]interface{}{
					"VariableDeclarator":   map[string]interface{}{"array": true, "object": false},
					"AssignmentExpression": map[string]interface{}{"object": true},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			Code:   "class Foo extends (Object as any) { static foo() {var bar = super.foo.bar} }",
			Output: []string{"class Foo extends (Object as any) { static foo() {var {bar} = super.foo} }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- Options JSON path test (map[string]interface{}) ----
		{
			Code:    "var foo = ({ foo: 1 } as any).foo;",
			Options: map[string]interface{}{"object": true},
			Output:  []string{"var {foo} = { foo: 1 } as any;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring"},
			},
		},

		// ---- Symbol.iterator invalid cases ----
		{
			Code: "let x: { [Symbol.iterator]: unknown };\nlet y = x[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "let x: { [Symbol.iterator]: unknown };\ny = x[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},

		// ---- Line/Column position assertions ----
		{
			Code: "var foo = ({ foo: 1 } as any).foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Line: 1, Column: 5},
			},
			Output: []string{"var {foo} = { foo: 1 } as any;"},
		},
		{
			Code: "let obj = { foo: 'bar' };\nconst foo = obj.foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Line: 2, Column: 7},
			},
			Output: []string{"let obj = { foo: 'bar' };\nconst {foo} = obj;"},
		},
		{
			Code: "foo = ({ foo: 1 } as any).foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Line: 1, Column: 1},
			},
		},

		// ---- Deeply nested member expression ----
		{
			Code:   "const a = { b: { c: { d: 1 } } };\nconst d = a.b.c.d;",
			Output: []string{"const a = { b: { c: { d: 1 } } };\nconst {d} = a.b.c;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- Parenthesized object on RHS ----
		{
			// Multiple parens: ((obj)).foo → {foo} = obj
			Code:   "const obj: any = {};\nconst foo = ((obj)).foo;",
			Output: []string{"const obj: any = {};\nconst {foo} = obj;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring"},
			},
		},

		// ---- TS type-assertion wrappers ----
		{
			// (obj as any).foo → {foo} = obj as any
			Code:   "const obj = { foo: 1 };\nconst foo = (obj as any).foo;",
			Output: []string{"const obj = { foo: 1 };\nconst {foo} = obj as any;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring"},
			},
		},

		// ---- Element access with string literal key (same name) ----
		{
			Code: "const obj: { foo: number } = { foo: 1 };\nconst foo = obj['foo'];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- Default options: var foo = array[0] with iterable type ----
		{
			Code: "const array = [1, 2, 3];\nvar foo = array[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "const array = [1, 2, 3];\nfoo = array[0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},

		// ---- Non-null assertion wrapping (TS-specific) ----
		{
			Code:   "const obj: { foo: number } | undefined = { foo: 1 };\nconst foo = obj!.foo;",
			Output: []string{"const obj: { foo: number } | undefined = { foo: 1 };\nconst {foo} = obj!;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- ESLint core: self-referential computed access ----
		{
			Code: "var foo = ({ foo: 1 } as any)[foo];",
			Options: []interface{}{
				map[string]interface{}{"object": true},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- ESLint core: per-context VD-only key with enforceForRenamedProperties ----
		{
			Code: "var foo = [1][0];",
			Options: []interface{}{
				map[string]interface{}{"VariableDeclarator": map[string]interface{}{"array": true}},
				map[string]interface{}{"enforceForRenamedProperties": true},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},

		// ---- ESLint core: comment edge cases (fix produced) ----
		{
			// Comment before identifier: fix works
			Code:   "var /* comment */ foo = ({ foo: 1 } as any).foo;",
			Output: []string{"var /* comment */ {foo} = { foo: 1 } as any;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			// Comment inside object expression (function call args): fix works
			Code:   "function bar(...args: any[]): any { return args; }\nvar foo = bar(/* comment */).foo;",
			Output: []string{"function bar(...args: any[]): any { return args; }\nvar {foo} = bar(/* comment */);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- ESLint core: comment edge cases (fix suppressed) ----
		{
			// Line comment between = and expression
			Code: "var foo = // comment\n({ foo: 1 } as any).foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			// Comment inside parenthesized object
			Code: "var foo = (/* comment */ { foo: 1 } as any).foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			// Comment between object and .property
			Code: "var foo = ({ foo: 1 } as any)// comment\n.foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},
		{
			// Comment between dot and property name
			Code: "var foo = ({ foo: 1 } as any)./* comment */foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use object destructuring."},
			},
		},

		// ---- Non-decimal numeric literal indexes (tsgo normalizes to decimal) ----
		{
			Code: "let x: number[] = [1, 2, 3];\nlet y = x[0x0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "let x: number[] = [1, 2];\nlet y = x[0b1];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "let x: number[] = [1, 2, 3];\nlet y = x[0o2];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
		{
			Code: "let x = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11];\nlet y = x[1_0];",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferDestructuring", Message: "Use array destructuring."},
			},
		},
	})
}
