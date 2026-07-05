// TestNoUnusedPrivateClassMembersExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
//
// Dimension 4 walk (rows that don't apply to no-unused-private-class-members,
// with reasons):
//   - N/A element access forms (`this['#x']`, `this[expr]`): JavaScript
//     hash-private members can only be accessed through `.#x` or `#x in obj`.
//   - N/A optional chaining (`obj?.#x`): ECMAScript private identifiers do not
//     have an optional chaining form.
//   - N/A numeric/string/computed declaration keys: the core rule only tracks
//     `PrivateIdentifier` keys, not static public keys that happen to spell
//     `#x`.
//   - N/A autofix boundaries: upstream exposes a suggestion, not an autofix.
//   - N/A overload signatures / abstract / declare members: those are
//     TypeScript-only forms and are outside the core ESLint rule's declaration
//     surface.
package no_unused_private_class_members

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedPrivateClassMembersExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedPrivateClassMembersRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized receiver ----
			{Code: `
class C {
    #x;
    method() {
        return (this).#x;
    }
}`},
			{Code: `
class C {
    #x;
    method() {
        return ((this)).#x;
    }
}`},

			// ---- Dimension 4: receiver wrapped in TypeScript expression forms ----
			{Code: `
class C {
    #x;
    method() {
        return (this as C).#x;
    }
}`},
			{Code: `
class C {
    #x;
    method() {
        return (this satisfies C).#x;
    }
}`},
			{Code: `
class C {
    #x;
    method() {
        return (<C>this).#x;
    }
}`},
			{Code: `
class C {
    #x;
    method() {
        return this!.#x;
    }
}`},

			// ---- Dimension 4: declaration/container forms ----
			{Code: `
const C = class {
    #x;
    method() {
        return this.#x;
    }
};`},
			{Code: `
class C {
    static #x;
    static {
        this.#x;
    }
}`},
			{Code: `
class C {
    method() {
        return this.#x;
    }
    #x;
}`},
			{Code: `
class C {
    #x;
    reader = () => this.#x;
}`},
			{Code: `
class C {
    static #x;
    method() {
        return C.#x;
    }
}`},

			// ---- Dimension 4: async / generator method bodies ----
			{Code: `
class C {
    #x;
    async method() {
        return this.#x;
    }
}`},
			{Code: `
class C {
    #x;
    *method() {
        yield this.#x;
    }
}`},

			// ---- Dimension 4: nesting boundaries ----
			{Code: `
class Outer {
    #outer;
    method() {
        class Inner {
            #inner;
            method() {
                return this.#inner;
            }
        }
        return this.#outer;
    }
}`},

			// ---- Dimension 4: private names are lexically visible in nested bodies ----
			{Code: `
class Outer {
    #x;
    method(obj) {
        return class Inner {
            check() {
                return #x in obj;
            }
        };
    }
}`},
			{Code: `
class C {
    #x;
    method() {
        return function() {
            return this.#x;
        };
    }
}`},

			// ---- Dimension 4: private-in expression counts as a read ----
			// Locks in upstream PrivateIdentifier() arm: non-member-expression
			// private identifiers are references, not declarations.
			{Code: `
class C {
    #x;
    method(obj) {
        return #x in obj;
    }
}`},

			// ---- Dimension 4: expression-valued writes count as reads ----
			{Code: `
class C {
    #x;
    method() {
        return this.#x += 1;
    }
}`},
			{Code: `
class C {
    #x;
    method() {
        return this.#x++;
    }
}`},
			{Code: `
class C {
    #x;
    method() {
        return this.#x ||= fallback;
    }
}`},
			{Code: `
class C {
    #x;
    method() {
        this.#x = this.#x + 1;
    }
}`},

			// ---- tsgo AST quirk: nested destructuring computed keys are reads ----
			{Code: `
class C {
    #key;
    method() {
        ({ nested: { [this.#key]: value } } = source);
    }
}`},

			// ---- Real-user: TS private modifier request stays outside core ----
			{Code: `
class C {
    private value = 1;
}`},
			{Code: `
class C {
    private value = 1;
    method() {
        this.value = 2;
    }
}`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: class expressions are inspected ----
			invalidUnusedPrivateCase(
				`const C = class { #x; };`,
				unusedPrivateError("#x", 1, 19, 1, 21, `const C = class { };`),
			),

			// ---- Dimension 4: public keys that spell "#x" don't satisfy private #x ----
			invalidUnusedPrivateCase(
				`class C {
    #x;
    "#x"() {
        return 1;
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7, `class C {
    "#x"() {
        return 1;
    }
}`),
			),

			// ---- Real-user: write-only update from the original rule proposal ----
			// Locks in upstream isWriteOnlyAssignment()/UpdateExpression arm.
			invalidUnusedPrivateCase(
				`class C {
    #x;
    method() {
        this.#x++;
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7),
			),
			invalidUnusedPrivateCase(
				`class C {
    #x;
    method() {
        (++this.#x);
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7),
			),

			// ---- Real-user: suggestion preserves ASI before a keyword-named member ----
			// Locks in upstream getSemicolonInsertionToken() branch for `in`.
			invalidUnusedPrivateCase(
				`class C {
    foo = 1
    #unused
    in = 2
}`,
				unusedPrivateError("#unused", 3, 5, 3, 12, `class C {
    foo = 1;
    in = 2
}`),
			),

			// Locks in upstream PrivateIdentifier() branch: nested class shadowing
			// means the inner #x read does not mark the outer #x as used.
			invalidUnusedPrivateCase(
				`class Outer {
    #x;
    method() {
        return class Inner {
            #x;
            method() {
                return this.#x;
            }
        };
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7, `class Outer {
    method() {
        return class Inner {
            #x;
            method() {
                return this.#x;
            }
        };
    }
}`),
			),

			// Locks in lexical private-name lookup for `#x in obj`: an inner
			// class declaration shadows the outer private name.
			invalidUnusedPrivateCase(
				`class Outer {
    #x;
    method(obj) {
        return class Inner {
            #x;
            check() {
                return #x in obj;
            }
        };
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7, `class Outer {
    method(obj) {
        return class Inner {
            #x;
            check() {
                return #x in obj;
            }
        };
    }
}`),
			),

			// Locks in upstream isWriteOnlyAssignment() branch: compound
			// assignment used as a bare statement is still write-only.
			invalidUnusedPrivateCase(
				`class C {
    #x;
    method() {
        this.#x += 1;
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7),
			),
			invalidUnusedPrivateCase(
				`class C {
    #x;
    method() {
        this.#x ||= fallback;
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7),
			),
			invalidUnusedPrivateCase(
				`class C {
    #x;
    method() {
        this.#x = (this.#x = 1);
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7),
			),
			invalidUnusedPrivateCase(
				`class C {
    #x;
    method() {
        ((this as C)).#x = 1;
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7),
			),
			invalidUnusedPrivateCase(
				`class C {
    #x;
    method() {
        this!.#x = 1;
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7),
			),

			// Locks in upstream destructuring branch: property values in an
			// assignment pattern are writes, while computed keys are reads.
			invalidUnusedPrivateCase(
				`class C {
    #x;
    method() {
        ({ value: this.#x } = source);
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7),
			),
			invalidUnusedPrivateCase(
				`class C {
    #x;
    method() {
        ({ nested: { value: this.#x } } = source);
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7),
			),
			invalidUnusedPrivateCase(
				`class C {
    #x;
    method() {
        ({ nested: [...this.#x] } = source);
    }
}`,
				unusedPrivateError("#x", 2, 5, 2, 7),
			),
		},
	)
}
