package no_import_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoImportAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoImportAssignRule,
		// Valid cases — aligned with ESLint's no-import-assign test suite
		[]rule_tester.ValidTestCase{
			// Default import — member writes are allowed (not the binding itself)
			{Code: `import mod from 'mod'; mod.prop = 0`},
			{Code: `import mod from 'mod'; mod.prop += 0`},
			{Code: `import mod from 'mod'; mod.prop++`},
			{Code: `import mod from 'mod'; delete mod.prop`},
			{Code: `import mod from 'mod'; for (mod.prop in foo);`},
			{Code: `import mod from 'mod'; for (mod.prop of foo);`},
			{Code: `import mod from 'mod'; [mod.prop] = foo;`},
			{Code: `import mod from 'mod'; [...mod.prop] = foo;`},
			{Code: `import mod from 'mod'; ({ bar: mod.prop } = foo);`},
			{Code: `import mod from 'mod'; ({ ...mod.prop } = foo);`},

			// Named import — member writes are allowed
			{Code: `import {named} from 'mod'; named.prop = 0`},
			{Code: `import {named} from 'mod'; named.prop += 0`},
			{Code: `import {named} from 'mod'; named.prop++`},
			{Code: `import {named} from 'mod'; delete named.prop`},
			{Code: `import {named} from 'mod'; for (named.prop in foo);`},
			{Code: `import {named} from 'mod'; for (named.prop of foo);`},
			{Code: `import {named} from 'mod'; [named.prop] = foo;`},
			{Code: `import {named} from 'mod'; [...named.prop] = foo;`},
			{Code: `import {named} from 'mod'; ({ bar: named.prop } = foo);`},
			{Code: `import {named} from 'mod'; ({ ...named.prop } = foo);`},

			// Namespace import — nested member writes (depth >= 2) are allowed
			{Code: `import * as mod from 'mod'; mod.named.prop = 0`},
			{Code: `import * as mod from 'mod'; mod.named.prop += 0`},
			{Code: `import * as mod from 'mod'; mod.named.prop++`},
			{Code: `import * as mod from 'mod'; delete mod.named.prop`},
			{Code: `import * as mod from 'mod'; for (mod.named.prop in foo);`},
			{Code: `import * as mod from 'mod'; for (mod.named.prop of foo);`},
			{Code: `import * as mod from 'mod'; [mod.named.prop] = foo;`},
			{Code: `import * as mod from 'mod'; [...mod.named.prop] = foo;`},
			{Code: `import * as mod from 'mod'; ({ bar: mod.named.prop } = foo);`},
			{Code: `import * as mod from 'mod'; ({ ...mod.named.prop } = foo);`},

			// Namespace import used as computed key or non-assignment
			{Code: `import * as mod from 'mod'; obj[mod] = 0`},
			{Code: `import * as mod from 'mod'; obj[mod.named] = 0`},
			{Code: `import * as mod from 'mod'; for (var foo in mod.named);`},
			{Code: `import * as mod from 'mod'; for (var foo of mod.named);`},
			{Code: `import * as mod from 'mod'; [bar = mod.named] = foo;`},
			{Code: `import * as mod from 'mod'; ({ bar = mod.named } = foo);`},
			{Code: `import * as mod from 'mod'; ({ bar: baz = mod.named } = foo);`},
			{Code: `import * as mod from 'mod'; ({ [mod.named]: bar } = foo);`},
			{Code: `import * as mod from 'mod'; var obj = { ...mod.named };`},
			{Code: `import * as mod from 'mod'; var obj = { foo: mod.named };`},

			// Block-scoped shadow — local redeclaration is allowed
			{Code: `import mod from 'mod'; { let mod = 0; mod = 1 }`},
			{Code: `import * as mod from 'mod'; { let mod = 0; mod = 1 }`},
			{Code: `import * as mod from 'mod'; { let mod = 0; mod.named = 1 }`},

			// Type assertion wrappers — intentional bypass, not flagged (aligns with ESLint)
			// PropertyAccess with various assertion styles
			{Code: `import * as mod from 'mod'; (mod.named as any) = 0`},
			{Code: `import * as mod from 'mod'; (mod.named as any) += 0`},
			{Code: `import * as mod from 'mod'; (mod.named as any)++`},
			{Code: `import * as mod from 'mod'; (<any>mod.named) = 0`},
			{Code: `import * as mod from 'mod'; (mod.named!) = 0`},
			// ElementAccess with type assertion
			{Code: `import * as mod from 'mod'; (mod["named"] as any) = 0`},
			// Nested parentheses around type assertion
			{Code: `import * as mod from 'mod'; ((mod.named as any)) = 0`},
			// Chained type assertions
			{Code: `import * as mod from 'mod'; (mod.named as any as unknown) = 0`},
			// delete with type assertion
			{Code: `import * as mod from 'mod'; delete (mod.named as any)`},
			// for-in/of with type assertion
			{Code: `import * as mod from 'mod'; for ((mod.named as any) in foo);`},
			{Code: `import * as mod from 'mod'; for ((mod.named as any) of foo);`},
			// Destructuring with type assertion
			{Code: `import * as mod from 'mod'; [(mod.named as any)] = foo`},
			{Code: `import * as mod from 'mod'; ({ bar: (mod.named as any) } = foo)`},
			// Mutation function with type assertion on the namespace
			{Code: `import * as mod from 'mod'; Object.assign(mod as any, obj)`},

			// Object/Reflect locally shadowed — mutation calls are safe
			{Code: `import * as mod from 'mod'; { var Object; Object.assign(mod, obj); }`},
			{Code: `import * as mod from 'mod'; var Object; Object.assign(mod, obj);`},

			// Empty / bare imports
			{Code: `import {} from 'mod'`},
			{Code: `import 'mod'`},

			// Object/Reflect methods — allowed on default/named imports
			{Code: `import mod from 'mod'; Object.assign(mod, obj);`},
			{Code: `import {named} from 'mod'; Object.assign(named, obj);`},

			// Namespace as non-first argument or safe method calls
			{Code: `import * as mod from 'mod'; Object.assign(mod.prop, obj);`},
			{Code: `import * as mod from 'mod'; Object.assign(obj, mod, other);`},
			{Code: `import * as mod from 'mod'; Object[assign](mod, obj);`},
			{Code: `import * as mod from 'mod'; Object.getPrototypeOf(mod);`},
			{Code: `import * as mod from 'mod'; Reflect.set(obj, key, mod);`},
			{Code: `import * as mod from 'mod'; Object.seal(mod, obj)`},
			{Code: `import * as mod from 'mod'; Object.preventExtensions(mod)`},
			{Code: `import * as mod from 'mod'; Reflect.preventExtensions(mod)`},

			// Re-export is not a write
			{Code: `import {a} from 'mod'; export {a}`},

			// Read-only usage
			{Code: `import mod from 'mod'; console.log(mod)`},
			{Code: `import {named} from 'mod'; console.log(named)`},
			{Code: `import * as mod from 'mod'; console.log(mod)`},

			// Calling imports is not a write
			{Code: `import mod from 'mod'; mod()`},
			{Code: `import {named} from 'mod'; named()`},
		},
		// Invalid cases — aligned with ESLint's no-import-assign test suite
		[]rule_tester.InvalidTestCase{
			// ========== Default import — direct reassignment ==========
			{
				Code:   `import mod1 from 'mod'; mod1 = 0`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 25}},
			},
			{
				Code:   `import mod2 from 'mod'; mod2 += 0`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 25}},
			},
			{
				Code:   `import mod3 from 'mod'; mod3++`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 25}},
			},
			{
				Code:   `import mod4 from 'mod'; for (mod4 in foo);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 30}},
			},
			{
				Code:   `import mod5 from 'mod'; for (mod5 of foo);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 30}},
			},
			{
				Code:   `import mod6 from 'mod'; [mod6] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 26}},
			},
			{
				Code:   `import mod7 from 'mod'; [mod7 = 0] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 26}},
			},
			{
				Code:   `import mod8 from 'mod'; [...mod8] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 29}},
			},
			{
				Code:   `import mod9 from 'mod'; ({ bar: mod9 } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 33}},
			},
			{
				Code:   `import mod10 from 'mod'; ({ bar: mod10 = 0 } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 34}},
			},
			{
				Code:   `import mod11 from 'mod'; ({ ...mod11 } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 32}},
			},

			// ========== Named import — direct reassignment ==========
			{
				Code:   `import {named1} from 'mod'; named1 = 0`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 29}},
			},
			{
				Code:   `import {named2} from 'mod'; named2 += 0`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 29}},
			},
			{
				Code:   `import {named3} from 'mod'; named3++`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 29}},
			},
			{
				Code:   `import {named4} from 'mod'; for (named4 in foo);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 34}},
			},
			{
				Code:   `import {named5} from 'mod'; for (named5 of foo);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 34}},
			},
			{
				Code:   `import {named6} from 'mod'; [named6] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 30}},
			},
			{
				Code:   `import {named7} from 'mod'; [named7 = 0] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 30}},
			},
			{
				Code:   `import {named8} from 'mod'; [...named8] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 33}},
			},
			{
				Code:   `import {named9} from 'mod'; ({ bar: named9 } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 37}},
			},
			{
				Code:   `import {named10} from 'mod'; ({ bar: named10 = 0 } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 38}},
			},
			{
				Code:   `import {named11} from 'mod'; ({ ...named11 } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 36}},
			},

			// ========== Namespace import — direct reassignment ==========
			{
				Code:   `import * as mod1 from 'mod'; mod1 = 0`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 30}},
			},
			{
				Code:   `import * as mod2 from 'mod'; mod2 += 0`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 30}},
			},
			{
				Code:   `import * as mod3 from 'mod'; mod3++`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 30}},
			},
			{
				Code:   `import * as mod4 from 'mod'; for (mod4 in foo);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 35}},
			},
			{
				Code:   `import * as mod5 from 'mod'; for (mod5 of foo);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 35}},
			},
			{
				Code:   `import * as mod6 from 'mod'; [mod6] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 31}},
			},
			{
				Code:   `import * as mod7 from 'mod'; [mod7 = 0] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 31}},
			},
			{
				Code:   `import * as mod8 from 'mod'; [...mod8] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 34}},
			},
			{
				Code:   `import * as mod9 from 'mod'; ({ bar: mod9 } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 38}},
			},
			{
				Code:   `import * as mod10 from 'mod'; ({ bar: mod10 = 0 } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 39}},
			},
			{
				Code:   `import * as mod11 from 'mod'; ({ ...mod11 } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonly", Line: 1, Column: 37}},
			},

			// ========== Namespace import — member modification ==========
			{
				Code:   `import * as mod1 from 'mod'; mod1.named = 0`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 30}},
			},
			{
				Code:   `import * as mod2 from 'mod'; mod2.named += 0`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 30}},
			},
			{
				Code:   `import * as mod3 from 'mod'; mod3.named++`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 30}},
			},
			{
				Code:   `import * as mod4 from 'mod'; for (mod4.named in foo);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 35}},
			},
			{
				Code:   `import * as mod5 from 'mod'; for (mod5.named of foo);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 35}},
			},
			{
				Code:   `import * as mod6 from 'mod'; [mod6.named] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 31}},
			},
			{
				Code:   `import * as mod7 from 'mod'; [mod7.named = 0] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 31}},
			},
			{
				Code:   `import * as mod8 from 'mod'; [...mod8.named] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 34}},
			},
			{
				Code:   `import * as mod9 from 'mod'; ({ bar: mod9.named } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 38}},
			},
			{
				Code:   `import * as mod10 from 'mod'; ({ bar: mod10.named = 0 } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 39}},
			},
			{
				Code:   `import * as mod11 from 'mod'; ({ ...mod11.named } = foo)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 37}},
			},

			// ========== Namespace import — delete and mutation functions ==========
			{
				Code:   `import * as mod12 from 'mod'; delete mod12.named`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 38}},
			},
			{
				Code:   `import * as mod from 'mod'; Object.assign(mod, obj)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 43}},
			},
			{
				Code:   `import * as mod from 'mod'; Object.defineProperty(mod, key, d)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 51}},
			},
			{
				Code:   `import * as mod from 'mod'; Object.defineProperties(mod, d)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 53}},
			},
			{
				Code:   `import * as mod from 'mod'; Object.setPrototypeOf(mod, proto)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 51}},
			},
			{
				Code:   `import * as mod from 'mod'; Object.freeze(mod)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 43}},
			},
			{
				Code:   `import * as mod from 'mod'; Reflect.defineProperty(mod, key, d)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 52}},
			},
			{
				Code:   `import * as mod from 'mod'; Reflect.deleteProperty(mod, key)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 52}},
			},
			{
				Code:   `import * as mod from 'mod'; Reflect.set(mod, key, value)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 41}},
			},
			{
				Code:   `import * as mod from 'mod'; Reflect.setPrototypeOf(mod, proto)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 52}},
			},

			// ========== Namespace import — element access ==========
			{
				Code:   `import * as mod from 'mod'; mod["named"] = 0`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 29}},
			},

			// ========== Optional chaining ==========
			{
				Code:   `import * as mod from 'mod'; Object?.defineProperty(mod, key, d)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 52}},
			},
			{
				Code:   `import * as mod from 'mod'; (Object?.defineProperty)(mod, key, d)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 54}},
			},
			{
				Code:   `import * as mod from 'mod'; delete mod?.named`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 36}},
			},

			// ========== Mixed imports ==========
			{
				Code:   `import mod, * as mod_ns from 'mod'; mod.prop = 0; mod_ns.prop = 0`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readonlyMember", Line: 1, Column: 51}},
			},
		},
	)
}
