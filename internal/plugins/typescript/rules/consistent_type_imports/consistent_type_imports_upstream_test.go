package consistent_type_imports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestConsistentTypeImportsUpstream migrates the valid/invalid semantic groups
// from packages/eslint-plugin/tests/rules/consistent-type-imports.test.ts.
// Position assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in consistent_type_imports_extras_test.go.
func TestConsistentTypeImportsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTypeImportsRule, []rule_tester.ValidTestCase{
		// ---- value references and unused bindings ----
		{Code: `import Foo from 'foo'; const foo: Foo = new Foo();`},
		{Code: `import foo from 'foo'; const foo: foo.Foo = foo.fn();`},
		{Code: `import { A, B } from 'foo'; const foo: A = B(); const bar = new A();`},
		{Code: `import Foo from 'foo';`},
		{Code: `import { A, B } from 'foo'; const b = B;`},
		{Code: `import {} from 'foo';`},
		{Code: `import 'foo';`},

		// ---- shadowing ----
		{Code: `import Foo from 'foo'; type T<Foo> = Foo;`},
		{Code: `import Foo from 'foo'; function fn() { type Foo = {}; let foo: Foo; }`},

		// ---- already type-only / inline type imports ----
		{Code: `import type Foo from 'foo'; type T = Foo;`},
		{Code: `import type { A, B } from 'foo'; type T = A | B;`},
		{Code: `import type * as Foo from 'foo'; type T = Foo.Bar;`},
		{Code: `import { type A, B } from 'foo'; type T = A; const b = B();`},
		{Code: `import { type A } from 'foo';`},

		// ---- options ----
		{Code: `import Foo from 'foo'; type T = Foo;`, Options: map[string]interface{}{"prefer": "no-type-imports"}},
		{Code: `let foo: import('foo').Foo;`, Options: map[string]interface{}{"disallowTypeAnnotations": false}},
		{Code: `import { A, type B } from 'foo'; const a = A; type T = B;`, Options: map[string]interface{}{"fixStyle": "inline-type-imports"}},

		// ---- exports and import attributes ----
		{Code: `import Foo from 'foo'; export { Foo };`},
		{Code: `import Foo from 'foo'; export default Foo;`},
		{Code: `import * as Type from 'foo' with { type: 'json' }; type T = typeof Type;`},
	}, []rule_tester.InvalidTestCase{
		// ---- all specifiers are types ----
		{
			Code:   `import Foo from 'foo'; type T = Foo;`,
			Output: []string{`import type Foo from 'foo'; type T = Foo;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typeOverValue", Message: "All imports in the declaration are only used as types. Use `import type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}},
		},
		{
			Code:    `import Foo from 'foo'; type T = Foo;`,
			Options: map[string]interface{}{},
			Output:  []string{`import type Foo from 'foo'; type T = Foo;`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "typeOverValue", Line: 1, Column: 1}},
		},
		{
			Code:   `import { Type } from 'foo'; type T = Type;`,
			Output: []string{`import type { Type } from 'foo'; type T = Type;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typeOverValue", Line: 1, Column: 1}},
		},
		{
			Code:   `import * as Type from 'foo'; type T = typeof Type.foo;`,
			Output: []string{`import type * as Type from 'foo'; type T = typeof Type.foo;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typeOverValue", Line: 1, Column: 1}},
		},
		{
			Code:   `import { Type as Alias } from 'foo'; interface T { value: Alias }`,
			Output: []string{`import type { Type as Alias } from 'foo'; interface T { value: Alias }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typeOverValue", Line: 1, Column: 1}},
		},

		// ---- mixed type/value imports ----
		{
			Code: `import { A, B } from 'foo'; const foo: A = B();`,
			Output: []string{`import type { A} from 'foo';
import { B } from 'foo'; const foo: A = B();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "someImportsAreOnlyTypes", Message: `Imports "A" are only used as type.`, Line: 1, Column: 1}},
		},
		{
			Code: `import { A, B, C } from 'foo'; const foo: A = B(); let bar: C;`,
			Output: []string{`import type { A, C } from 'foo';
import { B } from 'foo'; const foo: A = B(); let bar: C;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "someImportsAreOnlyTypes", Line: 1, Column: 1}},
		},
		{
			Code: `import A, { B } from 'foo'; B(); type T = A;`,
			Output: []string{`import type A from 'foo';
import { B } from 'foo'; B(); type T = A;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "someImportsAreOnlyTypes", Line: 1, Column: 1}},
		},
		{
			Code:    `import { A, B } from 'foo'; const foo: A = B();`,
			Options: map[string]interface{}{"fixStyle": "inline-type-imports"},
			Output:  []string{`import { type A, B } from 'foo'; const foo: A = B();`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "someImportsAreOnlyTypes", Line: 1, Column: 1}},
		},
		{
			Code:    `import Value, { Type } from 'foo'; const value = Value; type T = Type;`,
			Options: map[string]interface{}{"fixStyle": "inline-type-imports"},
			Output:  []string{`import Value, { type Type } from 'foo'; const value = Value; type T = Type;`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "someImportsAreOnlyTypes", Line: 1, Column: 1}},
		},

		// ---- merge into an existing type-only named import ----
		{
			Code: `import type { Already } from 'foo';
import { A, B } from 'foo'; const b = B; type T = A;`,
			Output: []string{`import type { Already , A} from 'foo';
import { B } from 'foo'; const b = B; type T = A;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "someImportsAreOnlyTypes", Line: 2, Column: 1}},
		},

		// ---- prefer no-type-imports ----
		{
			Code:    `import type Foo from 'foo'; type T = Foo;`,
			Options: map[string]interface{}{"prefer": "no-type-imports"},
			Output:  []string{`import Foo from 'foo'; type T = Foo;`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "avoidImportType", Message: "Use an `import` instead of an `import type`.", Line: 1, Column: 1}},
		},
		{
			Code:    `import { type Foo, type Bar } from 'foo'; type T = Foo | Bar;`,
			Options: map[string]interface{}{"prefer": "no-type-imports"},
			Output:  []string{`import { Foo, Bar } from 'foo'; type T = Foo | Bar;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "avoidImportType", Line: 1, Column: 10},
				{MessageId: "avoidImportType", Line: 1, Column: 20},
			},
		},

		// ---- import() annotations ----
		{
			Code:    `let foo: import('foo').Foo;`,
			Options: map[string]interface{}{"disallowTypeAnnotations": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noImportTypeAnnotations", Message: "`import()` type annotations are forbidden.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27}},
		},

		// ---- type queries, computed type keys and type-only exports ----
		{
			Code:   `import Type from 'foo'; type T = typeof Type;`,
			Output: []string{`import type Type from 'foo'; type T = typeof Type;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typeOverValue", Line: 1, Column: 1}},
		},
		{
			Code:   `import key from 'foo'; type T = { [key]: string };`,
			Output: []string{`import type key from 'foo'; type T = { [key]: string };`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typeOverValue", Line: 1, Column: 1}},
		},
		{
			Code:   `import Type from 'foo'; export type { Type };`,
			Output: []string{`import type Type from 'foo'; export type { Type };`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typeOverValue", Line: 1, Column: 1}},
		},
	})
}
