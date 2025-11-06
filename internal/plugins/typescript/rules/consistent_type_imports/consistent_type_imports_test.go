package consistent_type_imports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTypeImportsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTypeImportsRule, []rule_tester.ValidTestCase{
		// Default imports used as values
		{Code: `import Foo from 'foo'; const foo: Foo = new Foo();`},
		{Code: `import Foo from 'foo'; const foo = Foo;`},
		{Code: `import Foo from 'foo'; class Bar extends Foo {}`},

		// Named imports with mixed usage
		{Code: `import { A, B } from 'foo'; const a: A = B();`},
		{Code: `import { A, B } from 'foo'; const a = A; type T = B;`},

		// Empty imports
		{Code: `import {} from 'foo';`},
		{Code: `import 'foo';`},

		// Already type-only imports (valid with default prefer: 'type-imports')
		{Code: `import type Foo from 'foo'; type T = Foo;`},
		{Code: `import type { A, B } from 'foo'; type T = A | B;`},
		{Code: `import type * as Foo from 'foo'; type T = Foo.Bar;`},

		// Inline type imports (TypeScript 4.5+)
		{Code: `import { type A, B } from 'foo'; type T = A; const b = B();`},
		{Code: `import { A, type B } from 'foo'; const a = A(); type T = B;`},
		{Code: `import { type A, type B, C } from 'foo'; type T = A | B; const c = C();`},

		// Namespace imports used as values
		{Code: `import * as Foo from 'foo'; const foo = Foo.bar;`},
		{Code: `import * as Foo from 'foo'; Foo.doSomething();`},

		// Options: prefer: 'no-type-imports'
		{
			Code:    `import Foo from 'foo'; type T = Foo;`,
			Options: []interface{}{map[string]interface{}{"prefer": "no-type-imports"}},
		},
		{
			Code:    `import { A } from 'foo'; type T = A;`,
			Options: []interface{}{map[string]interface{}{"prefer": "no-type-imports"}},
		},

		// Options: disallowTypeAnnotations: false (allows import() in types)
		{
			Code:    `let foo: import('foo');`,
			Options: []interface{}{map[string]interface{}{"disallowTypeAnnotations": false}},
		},
		{
			Code:    `let foo: import('foo').Foo;`,
			Options: []interface{}{map[string]interface{}{"disallowTypeAnnotations": false}},
		},

		// Export re-exports
		{Code: `export { Foo } from 'foo';`},
		{Code: `export type { Foo } from 'foo';`},

		// Multiple imports from different modules
		{Code: `import type { A } from 'a'; import { B } from 'b'; type T = A; const b = B();`},

		// Side-effect imports
		{Code: `import './styles.css';`},
		{Code: `import 'reflect-metadata';`},
	}, []rule_tester.InvalidTestCase{
		// import() type annotations (disallowed by default)
		{
			Code: `let foo: import('foo');`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noImportTypeAnnotations"},
			},
		},
		{
			Code: `let foo: import('foo').Foo;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noImportTypeAnnotations"},
			},
		},
		{
			Code: `type T = import('foo').Foo;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noImportTypeAnnotations"},
			},
		},

		// Options: prefer: 'no-type-imports'
		{
			Code:    `import type Foo from 'foo'; type T = Foo;`,
			Options: []interface{}{map[string]interface{}{"prefer": "no-type-imports"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "avoidImportType"},
			},
		},
		{
			Code:    `import type { A } from 'foo'; type T = A;`,
			Options: []interface{}{map[string]interface{}{"prefer": "no-type-imports"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "avoidImportType"},
			},
		},
		{
			Code:    `import type * as Foo from 'foo'; type T = Foo.Bar;`,
			Options: []interface{}{map[string]interface{}{"prefer": "no-type-imports"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "avoidImportType"},
			},
		},
	})
}
