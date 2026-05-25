package parameter_properties

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestParameterProperties(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ParameterPropertiesRule, []rule_tester.ValidTestCase{
		// ============================================================
		// prefer: "class-property" (default) — valid cases
		// ============================================================

		// --- Basic: no parameter properties ---
		{Code: `class Foo { constructor(name: string) {} }`},
		{Code: `class Foo { constructor(name: string) {} }`, Options: map[string]interface{}{"prefer": "class-property"}},
		{Code: `class Foo { constructor(...name: string[]) {} }`},
		{Code: `class Foo { constructor(name: string, age: number) {} }`},

		// --- Constructor overloads without parameter properties ---
		{Code: `
class Foo {
  constructor(name: string) {}
  constructor(name: string, age?: number) {}
}`},

		// --- Allow specific modifiers (all 7 combinations) ---
		{Code: `class Foo { constructor(readonly name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"readonly"}}},
		{Code: `class Foo { constructor(private name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"private"}}},
		{Code: `class Foo { constructor(protected name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"protected"}}},
		{Code: `class Foo { constructor(public name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"public"}}},
		{Code: `class Foo { constructor(private readonly name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"private readonly"}}},
		{Code: `class Foo { constructor(protected readonly name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"protected readonly"}}},
		{Code: `class Foo { constructor(public readonly name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"public readonly"}}},

		// --- Multiple allowed modifiers ---
		{
			Code: `
class Foo {
  constructor(
    readonly name: string,
    private age: number,
  ) {}
}`,
			Options: map[string]interface{}{"allow": []interface{}{"readonly", "private"}},
		},
		{
			Code: `
class Foo {
  constructor(
    public readonly name: string,
    private age: number,
  ) {}
}`,
			Options: map[string]interface{}{"allow": []interface{}{"public readonly", "private"}},
		},

		// --- Semantically invalid: rest / destructured parameter properties ---
		{Code: `class Foo { constructor(private ...name: string[]) {} }`},
		{Code: `class Foo { constructor(private [test]: [string]) {} }`},

		// ============================================================
		// prefer: "parameter-property" — valid cases
		// ============================================================

		// --- Parameter property already used (no class property to convert) ---
		{Code: `class Foo { constructor(private name: string[]) {} }`, Options: map[string]interface{}{"prefer": "parameter-property"}},
		{Code: `class Foo { constructor(...name: string[]) {} }`, Options: map[string]interface{}{"prefer": "parameter-property"}},
		{Code: `class Foo { constructor(age: string, ...name: string[]) {} }`, Options: map[string]interface{}{"prefer": "parameter-property"}},
		{
			Code: `
class Foo {
  constructor(
    private age: string,
    ...name: string[]
  ) {}
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Type mismatch (different types) ---
		{
			Code: `
class Foo {
  public age: number;
  constructor(age: string) {
    this.age = age;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Property has initializer ---
		{
			Code: `
class Foo {
  public age = '';
  constructor(age: string) {
    this.age = age;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- One has type, the other doesn't ---
		{
			Code: `
class Foo {
  public age;
  constructor(age: string) {
    this.age = age;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},
		{
			Code: `
class Foo {
  public age: string;
  constructor(age) {
    this.age = age;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Non-leading assignment (chain broken before this.X = X) ---
		{
			Code: `
class Foo {
  public age: string;
  constructor(age: string) {
    console.log('unrelated');
    this.age = age;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Different property name vs parameter name ---
		{
			Code: `
class Foo {
  other: string;
  constructor(age: string) {
    this.other = age;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Assignment RHS is not the parameter identifier ---
		{
			Code: `
class Foo {
  age: string;
  constructor(age: string) {
    this.age = '';
    console.log(age);
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Property is a method, not a field ---
		{
			Code: `
class Foo {
  age() {
    return '';
  }
  constructor(age: string) {
    this.age = age;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Allow specific modifiers with prefer: parameter-property (all 6 combinations) ---
		{
			Code: `
class Foo {
  public age: string;
  constructor(age: string) { this.age = age; }
}`,
			Options: map[string]interface{}{"allow": []interface{}{"public"}, "prefer": "parameter-property"},
		},
		{
			Code: `
class Foo {
  public readonly age: string;
  constructor(age: string) { this.age = age; }
}`,
			Options: map[string]interface{}{"allow": []interface{}{"public readonly"}, "prefer": "parameter-property"},
		},
		{
			Code: `
class Foo {
  protected age: string;
  constructor(age: string) { this.age = age; }
}`,
			Options: map[string]interface{}{"allow": []interface{}{"protected"}, "prefer": "parameter-property"},
		},
		{
			Code: `
class Foo {
  protected readonly age: string;
  constructor(age: string) { this.age = age; }
}`,
			Options: map[string]interface{}{"allow": []interface{}{"protected readonly"}, "prefer": "parameter-property"},
		},
		{
			Code: `
class Foo {
  private age: string;
  constructor(age: string) { this.age = age; }
}`,
			Options: map[string]interface{}{"allow": []interface{}{"private"}, "prefer": "parameter-property"},
		},
		{
			Code: `
class Foo {
  private readonly age: string;
  constructor(age: string) { this.age = age; }
}`,
			Options: map[string]interface{}{"allow": []interface{}{"private readonly"}, "prefer": "parameter-property"},
		},

		// ============================================================
		// Edge cases — prefer: "parameter-property" — valid
		// ============================================================

		// --- Element access with string literal: this['member'] = member ---
		// ESLint MemberExpression check passes, but property.type is Literal (not
		// Identifier), so it breaks. Valid (no match).
		{
			Code: `
class Foo {
  member: string;
  constructor(member: string) {
    this['member'] = member;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Computed property name on class property ---
		{
			Code: `
class Foo {
  ['member']: string;
  constructor(member: string) {
    this.member = member;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Empty constructor body ---
		{
			Code: `
class Foo {
  member: string;
  constructor(member: string) {}
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Constructor without body (overload declaration) ---
		{
			Code: `
class Foo {
  member: string;
  constructor(member: string);
  constructor(member: string, extra?: number) {
    console.log(extra);
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Non-identifier RHS breaks chain: this.x = x + 1 ---
		{
			Code: `
class Foo {
  member: string;
  constructor(member: string) {
    this.member = member + '!';
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Assignment chain broken — second property not tracked ---
		{
			Code: `
class Foo {
  b: string;
  constructor(a: string, b: string) {
    this.a = a;
    console.log('break');
    this.b = b;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Inner class property should NOT match outer constructor ---
		{
			Code: `
class Outer {
  constructor(member: string) {
    this.member = member;
  }
}
class Inner {
  member: string;
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Nested: outer property + inner constructor should NOT cross-match ---
		{
			Code: `
class Outer {
  member: string;
  method() {
    class Inner {
      constructor(member: string) {
        this.member = member;
      }
    }
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Parameter already IS a parameter property → should NOT report ---
		// ESLint skips TSParameterProperty in the constructor handler;
		// in Go we must explicitly skip parameters with modifiers.
		{
			Code: `
class Foo {
  member: string;
  constructor(private member: string) {
    this.member = member;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- super() before assignments breaks the chain ---
		{
			Code: `
class Foo extends Bar {
  member: string;
  constructor(member: string) {
    super();
    this.member = member;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Compound assignment (+=) does NOT break the chain ---
		// ESLint checks for AssignmentExpression which includes all assignment
		// operators, so this.member += member is still treated as a matching
		// assignment (same as this.member = member). This is a quirk of the
		// original rule — it matches the syntactic pattern, not the semantics.
		// Moved to invalid cases below.

		// --- Destructured parameter → constructorParameter not set ---
		{
			Code: `
class Foo {
  member: string;
  constructor({ member }: { member: string }) {
    this.member = member;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Rest parameter → constructorParameter not set ---
		// ESLint: RestElement !== Identifier, so rest params are skipped.
		{
			Code: `
class Foo {
  args: string[];
  constructor(...args: string[]) {
    this.args = args;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Default value parameter → constructorParameter not set ---
		// ESLint: AssignmentPattern !== Identifier, so params with defaults are skipped.
		{
			Code: `
class Foo {
  x: string;
  constructor(x: string = 'default') {
    this.x = x;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},

		// --- Type annotation whitespace mismatch → no match ---
		// ESLint compares getText(TSTypeAnnotation) which includes trivia after
		// the colon. Different whitespace means different text → types don't match.
		{
			Code: `
class Foo {
  member:  string;
  constructor(member: string) {
    this.member = member;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
		},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// prefer: "class-property" (default) — invalid cases
		// ============================================================

		// --- All 7 modifier combinations ---
		{
			Code:   "\nclass Foo {\n  constructor(readonly name: string) {}\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 3, Column: 15}},
		},
		{
			Code:   "\nclass Foo {\n  constructor(private name: string) {}\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 3, Column: 15}},
		},
		{
			Code:   "\nclass Foo {\n  constructor(protected name: string) {}\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 3, Column: 15}},
		},
		{
			Code:   "\nclass Foo {\n  constructor(public name: string) {}\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 3, Column: 15}},
		},
		{
			Code:   "\nclass Foo {\n  constructor(private readonly name: string) {}\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 3, Column: 15}},
		},
		{
			Code:   "\nclass Foo {\n  constructor(protected readonly name: string) {}\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 3, Column: 15}},
		},
		{
			Code:   "\nclass Foo {\n  constructor(public readonly name: string) {}\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 3, Column: 15}},
		},

		// --- Mixed: one param-property + one plain ---
		{
			Code: `
class Foo {
  constructor(
    public name: string,
    age: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 4, Column: 5}},
		},

		// --- Multiple parameter properties ---
		{
			Code: `
class Foo {
  constructor(
    private name: string,
    private age: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferClassProperty", Line: 4, Column: 5},
				{MessageId: "preferClassProperty", Line: 5, Column: 5},
			},
		},
		{
			Code: `
class Foo {
  constructor(
    protected name: string,
    protected age: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferClassProperty", Line: 4, Column: 5},
				{MessageId: "preferClassProperty", Line: 5, Column: 5},
			},
		},
		{
			Code: `
class Foo {
  constructor(
    public name: string,
    public age: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferClassProperty", Line: 4, Column: 5},
				{MessageId: "preferClassProperty", Line: 5, Column: 5},
			},
		},

		// --- Constructor overloads ---
		{
			Code: `
class Foo {
  constructor(name: string) {}
  constructor(
    private name: string,
    age?: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 5, Column: 5}},
		},
		{
			Code: `
class Foo {
  constructor(private name: string) {}
  constructor(
    private name: string,
    age?: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferClassProperty", Line: 3, Column: 15},
				{MessageId: "preferClassProperty", Line: 5, Column: 5},
			},
		},
		{
			Code: `
class Foo {
  constructor(private name: string) {}
  constructor(
    private name: string,
    private age?: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferClassProperty", Line: 3, Column: 15},
				{MessageId: "preferClassProperty", Line: 5, Column: 5},
				{MessageId: "preferClassProperty", Line: 6, Column: 5},
			},
		},
		{
			Code: `
class Foo {
  constructor(name: string) {}
  constructor(
    protected name: string,
    age?: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 5, Column: 5}},
		},
		{
			Code: `
class Foo {
  constructor(protected name: string) {}
  constructor(
    protected name: string,
    age?: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferClassProperty", Line: 3, Column: 15},
				{MessageId: "preferClassProperty", Line: 5, Column: 5},
			},
		},
		{
			Code: `
class Foo {
  constructor(protected name: string) {}
  constructor(
    protected name: string,
    protected age?: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferClassProperty", Line: 3, Column: 15},
				{MessageId: "preferClassProperty", Line: 5, Column: 5},
				{MessageId: "preferClassProperty", Line: 6, Column: 5},
			},
		},
		{
			Code: `
class Foo {
  constructor(name: string) {}
  constructor(
    public name: string,
    age?: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 5, Column: 5}},
		},
		{
			Code: `
class Foo {
  constructor(public name: string) {}
  constructor(
    public name: string,
    age?: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferClassProperty", Line: 3, Column: 15},
				{MessageId: "preferClassProperty", Line: 5, Column: 5},
			},
		},
		{
			Code: `
class Foo {
  constructor(public name: string) {}
  constructor(
    public name: string,
    public age?: number,
  ) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferClassProperty", Line: 3, Column: 15},
				{MessageId: "preferClassProperty", Line: 5, Column: 5},
				{MessageId: "preferClassProperty", Line: 6, Column: 5},
			},
		},

		// --- Allow options ---
		{
			Code:    `class Foo { constructor(readonly name: string) {} }`,
			Options: map[string]interface{}{"allow": []interface{}{"private"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty"}},
		},
		{
			Code:    `class Foo { constructor(private name: string) {} }`,
			Options: map[string]interface{}{"allow": []interface{}{"readonly"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty"}},
		},
		{
			Code:    `class Foo { constructor(protected name: string) {} }`,
			Options: map[string]interface{}{"allow": []interface{}{"readonly", "private", "public", "protected readonly"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty"}},
		},
		{
			Code:    `class Foo { constructor(public name: string) {} }`,
			Options: map[string]interface{}{"allow": []interface{}{"readonly", "private", "protected", "protected readonly", "public readonly"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty"}},
		},
		{
			Code:    `class Foo { constructor(private readonly name: string) {} }`,
			Options: map[string]interface{}{"allow": []interface{}{"readonly", "private"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty"}},
		},
		{
			Code:    `class Foo { constructor(protected readonly name: string) {} }`,
			Options: map[string]interface{}{"allow": []interface{}{"readonly", "protected", "private readonly", "public readonly"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty"}},
		},
		{
			Code: `
class Foo {
  constructor(private name: string) {}
  constructor(
    private name: string,
    protected age?: number,
  ) {}
}`,
			Options: map[string]interface{}{"allow": []interface{}{"private"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 6, Column: 5}},
		},

		// ============================================================
		// prefer: "class-property" — edge case invalid
		// ============================================================

		// --- Nested class: inner class reports its own parameter property ---
		{
			Code: `
class Outer {
  method() {
    class Inner {
      constructor(private x: string) {}
    }
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 5, Column: 19}},
		},

		// --- Class expression ---
		{
			Code:   `const A = class { constructor(private x: string) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty"}},
		},

		// --- Class expression nested in class ---
		{
			Code: `
class Outer {
  foo = class {
    constructor(private x: string) {}
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 4, Column: 17}},
		},

		// --- Decorator on parameter property (still reported) ---
		{
			Code: `
class Foo {
  constructor(@Inject private name: string) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty"}},
		},

		// --- Default value on parameter property ---
		{
			Code:   `class Foo { constructor(private name: string = 'default') {} }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty"}},
		},

		// --- Abstract class ---
		{
			Code: `
abstract class Foo {
  constructor(private name: string) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 3, Column: 15}},
		},

		// --- Generic class ---
		{
			Code: `
class Foo<T> {
  constructor(private name: T) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 3, Column: 15}},
		},

		// --- Extends class ---
		{
			Code: `
class Foo extends Bar {
  constructor(private name: string) {
    super();
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty", Line: 3, Column: 15}},
		},

		// --- Optional parameter property ---
		{
			Code:   `class Foo { constructor(private name?: string) {} }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferClassProperty"}},
		},

		// --- Both outer and inner classes have parameter properties ---
		// Note: inner class is visited and exited before outer in AST traversal,
		// so inner's error is emitted first.
		{
			Code: `
class Outer {
  constructor(private x: string) {}
  method() {
    class Inner {
      constructor(private y: string) {}
    }
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferClassProperty", Line: 3, Column: 15},
				{MessageId: "preferClassProperty", Line: 6, Column: 19},
			},
		},

		// ============================================================
		// prefer: "parameter-property" — invalid cases
		// ============================================================

		// --- Basic: property before constructor ---
		{
			Code: `
class Foo {
  member: string;

  constructor(member: string) {
    this.member = member;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferParameterProperty", Line: 3, Column: 3}},
		},

		// --- Property after constructor ---
		{
			Code: `
class Foo {
  constructor(member: string) {
    this.member = member;
  }

  member: string;
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferParameterProperty", Line: 7, Column: 3}},
		},

		// --- Both have no type annotation ---
		{
			Code: `
class Foo {
  member;
  constructor(member) {
    this.member = member;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferParameterProperty", Line: 3, Column: 3}},
		},

		// --- With allow list ---
		{
			Code: `
class Foo {
  public member: string;
  constructor(member: string) {
    this.member = member;
  }
}`,
			Options: map[string]interface{}{"allow": []interface{}{"protected", "private", "readonly"}, "prefer": "parameter-property"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferParameterProperty", Line: 3, Column: 3}},
		},

		// ============================================================
		// prefer: "parameter-property" — edge case invalid
		// ============================================================

		// --- Multiple leading assignments → all tracked ---
		{
			Code: `
class Foo {
  a: string;
  b: string;
  constructor(a: string, b: string) {
    this.a = a;
    this.b = b;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferParameterProperty", Line: 3, Column: 3},
				{MessageId: "preferParameterProperty", Line: 4, Column: 3},
			},
		},

		// --- Assignment chain broken → only first property tracked ---
		{
			Code: `
class Foo {
  a: string;
  b: string;
  constructor(a: string, b: string) {
    this.a = a;
    console.log('break');
    this.b = b;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferParameterProperty", Line: 3, Column: 3}},
		},

		// --- Class expression ---
		{
			Code: `
const A = class {
  member: string;
  constructor(member: string) {
    this.member = member;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferParameterProperty", Line: 3, Column: 3}},
		},

		// --- Nested classes: both outer and inner have violations ---
		// Inner class exits first in AST traversal, so inner's error is emitted first.
		{
			Code: `
class Outer {
  name: string;
  constructor(name: string) {
    this.name = name;
  }
  method() {
    class Inner {
      age: string;
      constructor(age: string) {
        this.age = age;
      }
    }
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferParameterProperty", Line: 9, Column: 7},
				{MessageId: "preferParameterProperty", Line: 3, Column: 3},
			},
		},

		// --- Nested: same property name in outer and inner (independent scopes) ---
		{
			Code: `
class Outer {
  member: string;
  constructor(member: string) {
    this.member = member;
  }
  method() {
    class Inner {
      member: string;
      constructor(member: string) {
        this.member = member;
      }
    }
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferParameterProperty", Line: 9, Column: 7},
				{MessageId: "preferParameterProperty", Line: 3, Column: 3},
			},
		},

		// --- Computed element access this[member] = member (Identifier arg) ---
		// ESLint MemberExpression covers both this.x and this[x]. When the
		// argument is an Identifier, it passes the property.type === Identifier check.
		{
			Code: `
class Foo {
  member: string;
  constructor(member: string) {
    this[member] = member;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferParameterProperty", Line: 3, Column: 3}},
		},

		// --- Compound assignment (+=) still matches (ESLint quirk) ---
		{
			Code: `
class Foo {
  member: string;
  constructor(member: string) {
    this.member += member;
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferParameterProperty", Line: 3, Column: 3}},
		},

		// --- Nested class expression ---
		{
			Code: `
class Outer {
  bar = class {
    member: string;
    constructor(member: string) {
      this.member = member;
    }
  }
}`,
			Options: map[string]interface{}{"prefer": "parameter-property"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferParameterProperty", Line: 4, Column: 5}},
		},
	})
}
