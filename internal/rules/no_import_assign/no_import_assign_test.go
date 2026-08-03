package no_import_assign

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
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

func TestNoImportAssignRuleAdversarial(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoImportAssignRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `import { a } from "mod";
				{ let a = 0; a = 1; }
				const object = { a: 1 };
				const shorthand = { a };
				const computed = { [a]: 1 };
				object.a = 2;
				export { a };
				a: for (;;) { break a; }`,
			},
			{
				Code: `import * as ns from "mod";
					{ const Object = factory; ((Object)).assign(((ns)), source); }
					Object[method](((ns)), source);`,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				// References from bindings in one declaration must retain source order.
				Code: `import { a, b } from "mod";
					b = 1;
					a = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Message: "'b' is read-only.", Line: 2, Column: 6},
					{MessageId: "readonly", Message: "'a' is read-only.", Line: 3, Column: 6},
				},
			},
			{
				// Separate declarations retain declaration-grouped report order.
				Code: `import { a } from "a";
					import { b } from "b";
					b = 1;
					a = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Message: "'a' is read-only.", Line: 4, Column: 6},
					{MessageId: "readonly", Message: "'b' is read-only.", Line: 3, Column: 6},
				},
			},
			{
				// Import declarations are hoisted, so references before the declaration count.
				Code: `a = 1;
					import { a } from "mod";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Message: "'a' is read-only.", Line: 1, Column: 1},
				},
			},
			{
				Code: `ns.value = 1;
					import * as ns from "mod";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Message: "The members of 'ns' are read-only.", Line: 1, Column: 1},
				},
			},
			{
				// Parentheses around a namespace reference are semantically transparent.
				Code: `import * as ns from "mod";
					((ns)).value = 1;
					Object.assign(((ns)), source);
					delete (((ns).removed));
					((Object)).defineProperty(((ns)), key, descriptor);
					Object["assign"](((ns)), source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Message: "The members of 'ns' are read-only.", Line: 2},
					{MessageId: "readonlyMember", Message: "The members of 'ns' are read-only.", Line: 3},
					{MessageId: "readonlyMember", Message: "The members of 'ns' are read-only.", Line: 4},
					{MessageId: "readonlyMember", Message: "The members of 'ns' are read-only.", Line: 5},
					{MessageId: "readonlyMember", Message: "The members of 'ns' are read-only.", Line: 6},
				},
			},
			{
				// RefStore resolves shorthand assignment patterns that the
				// checker returns as property symbols rather than import aliases.
				Code: `import { value } from "mod";
					({ value } = source);
					({ value = 0 } = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Message: "'value' is read-only.", Line: 2},
					{MessageId: "readonly", Message: "'value' is read-only.", Line: 3},
				},
			},
		},
	)
}

func TestNoImportAssignRuleWithoutRefStore(t *testing.T) {
	fallbackRule := NoImportAssignRule
	fallbackRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.Refs = nil
		return NoImportAssignRule.Run(ctx, options)
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&fallbackRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { value } from "mod"; { let value = 0; value = 1; }`},
			{
				Code: `import { first, second } from "mod";
					{ let first = 0; first = 1; }
					{ let second = {}; second.member = 1; }`,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `import { value } from "mod"; value = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Message: "'value' is read-only.", Line: 1, Column: 30},
				},
			},
			{
				Code: `import * as ns from "mod"; Object.assign(ns, value);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Message: "The members of 'ns' are read-only.", Line: 1, Column: 42},
				},
			},
			{
				Code: `import { value } from "mod"; ({ value = 0 } = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Message: "'value' is read-only.", Line: 1, Column: 33},
				},
			},
			{
				Code: `import { first, second } from "mod";
					second = 1;
					first = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Message: "'second' is read-only.", Line: 2},
					{MessageId: "readonly", Message: "'first' is read-only.", Line: 3},
				},
			},
		},
	)
}

func TestNoImportAssignRuleWithRefStoreWithoutTypeChecker(t *testing.T) {
	noCheckerRule := NoImportAssignRule
	noCheckerRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return NoImportAssignRule.Run(ctx, options)
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&noCheckerRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { value } from "mod"; { let value = 0; value = 1; }`},
			{
				Code: `import { first, second } from "mod";
					{ let first = 0; first = 1; }
					{ let second = {}; second.member = 1; }`,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `import { value } from "mod";
					{ let value = 0; value = 1; }
					value = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Message: "'value' is read-only.", Line: 3},
				},
			},
			{
				Code: `import * as namespace from "namespace";
					import { value } from "value";
					namespace.member = 1;
					value = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "readonlyMember",
						Message:   "The members of 'namespace' are read-only.",
						Line:      3,
					},
					{
						MessageId: "readonly",
						Message:   "'value' is read-only.",
						Line:      4,
					},
				},
			},
		},
	)
}

type noImportAssignDiagnosticFingerprint struct {
	messageID   string
	description string
	pos         int
	end         int
}

func lintNoImportAssignForComparison(
	t *testing.T,
	code string,
	mode noImportAssignRefStoreMode,
) []noImportAssignDiagnosticFingerprint {
	t.Helper()

	rootDir := fixtures.GetRootDir()
	fileName := "file.ts"
	fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatal(err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatal("comparison source file was not loaded")
		return nil
	}

	var diagnostics []noImportAssignDiagnosticFingerprint
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:      program,
		File:         sourceFile.FileName(),
		HasTypeInfo:  true,
		ExcludePaths: []string{},
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name:     NoImportAssignRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if mode == noImportAssignRefStoreDisabled {
						ctx.Refs = nil
					}
					return noImportAssignListeners(ctx, mode)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandNone,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, noImportAssignDiagnosticFingerprint{
					messageID:   diagnostic.Message.Id,
					description: diagnostic.Message.Description,
					pos:         diagnostic.Range.Pos(),
					end:         diagnostic.Range.End(),
				})
			},
		},
	})
	return diagnostics
}

func TestNoImportAssignStrategiesMatch(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{
			name: "direct writes and ordering",
			code: `before = 0;
				import before, { first as alpha, second as beta } from "one";
				beta += 1;
				[alpha] = source;
				before++;
				({ alpha = 0, beta: before } = source);`,
		},
		{
			name: "multiple declaration groups",
			code: `import { a } from "a";
				import { b, c } from "bc";
				b = 1;
				a = 2;
				c = 3;`,
		},
		{
			name: "namespace mutations and escapes",
			code: `ns.value = 1;
					import * as ns from "mod";
					ns["element"]++;
					delete ns.deleted;
					Object.assign(ns, source);
					Reflect.set(ns, key, value);
					((ns)).wrapped = 1;
					Object.assign(((ns)), source);
					Object["assign"](((ns)), source);
					(ns.escaped as any) = 1;
					ns.deep.value = 1;`,
		},
		{
			name: "integrated namespace mutation roots",
			code: `import * as ns from "ns";
				import { value } from "value";
				ns.property = 1;
				ns["element"]++;
				++ns.prefix;
				delete ((ns)).deleted;
				for (ns.key in source);
				for (ns.item of source);
				({ property: ns.destructured, ...ns } = source);
				Object.assign(((ns)), source);
				Object["defineProperty"](ns, key, descriptor);
				Reflect.set(ns, key, value);
				[value] = source;
				(value as any) = 1;
				ns.deep.property = 2;
				Object.assign(target, ns);`,
		},
		{
			name: "integrated nested write ordering and deduplication",
			code: `import { first, second } from "values";
				[(first = 1), second] = source;
				foo(first = 2);
				delete object[(second += 1)];`,
		},
		{
			name: "local shadows and syntactic names",
			code: `import { value } from "value";
				import * as ns from "ns";
				{ let value = 0; value = 1; }
				{ let ns = {}; ns.value = 1; }
				const object = { value: 1 };
				object.value = 2;
				value: for (;;) { break value; }
				value = 3;
				ns.value = 4;`,
		},
		{
			name: "shadowed mutation globals",
			code: `import * as ns from "ns";
				import { marker } from "marker";
				{ const Object = factory; Object.assign(ns, source); }
				Object.assign(ns, source);
				function f(Reflect) { Reflect.set(ns, key, value); }
				Reflect.set(ns, key, value);
				marker = 1;`,
		},
		{
			name: "type-only imports",
			code: `import type { T } from "types";
				import { type U, V } from "mixed";
				V = 1;
				T = 2;
				U = 3;`,
		},
		{
			name: "import attributes",
			code: `import data from "pkg" with { type: "json" };
				let type = 0;
				type = 1;
				data = replacement;`,
		},
		{
			name: "duplicate aliases in recovered syntax",
			code: `import { x as duplicate } from "x";
				import { y as duplicate } from "y";
				duplicate = 1;`,
		},
		{
			name: "duplicate local binding in recovered syntax",
			code: `import { x as duplicate } from "x";
				let duplicate;
				duplicate = 1;`,
		},
		{
			name: "nested import in recovered syntax",
			code: `if (condition) {
					import { nested } from "nested";
					nested = 1;
				}`,
		},
		{
			name: "mixed top-level and nested recovered imports",
			code: `import { first } from "first";
				if (condition) {
					import { nested } from "nested";
					nested = 1;
				}
				import { last } from "last";
				last = 2;
				first = 3;`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			auto := lintNoImportAssignForComparison(
				t,
				testCase.code,
				noImportAssignRefStoreAuto,
			)
			for _, strategy := range []struct {
				name string
				mode noImportAssignRefStoreMode
			}{
				{name: "RefStore", mode: noImportAssignRefStoreEnabled},
				{name: "fallback", mode: noImportAssignRefStoreDisabled},
			} {
				got := lintNoImportAssignForComparison(t, testCase.code, strategy.mode)
				if len(got) != len(auto) {
					t.Fatalf(
						"diagnostic count differs: auto=%+v %s=%+v",
						auto,
						strategy.name,
						got,
					)
				}
				for i := range auto {
					if got[i] != auto[i] {
						t.Fatalf(
							"diagnostic %d differs: auto=%+v %s=%+v",
							i,
							auto,
							strategy.name,
							got,
						)
					}
				}
			}
		})
	}
}

func TestNoImportAssignAutoHonorsDisableDirectives(t *testing.T) {
	diagnostics := lintNoImportAssignForComparison(t, `import { a, b } from "mod";
		// eslint-disable-next-line no-import-assign
		a = 1;
		b = 2;`, noImportAssignRefStoreAuto)

	if len(diagnostics) != 1 {
		t.Fatalf("got diagnostics %+v, want one unsuppressed diagnostic", diagnostics)
	}
	if diagnostics[0].description != "'b' is read-only." {
		t.Fatalf("got diagnostic %+v, want the write to b", diagnostics[0])
	}
}
