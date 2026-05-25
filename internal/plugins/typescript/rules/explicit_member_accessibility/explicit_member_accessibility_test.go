package explicit_member_accessibility

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExplicitMemberAccessibility(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitMemberAccessibilityRule, []rule_tester.ValidTestCase{
		// ---- accessibility: 'explicit' with overrides.parameterProperties: 'explicit' / 'off' ----
		{
			Code: `
class Test {
  public constructor(private foo: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "explicit",
				"overrides":     map[string]interface{}{"parameterProperties": "explicit"},
			},
		},
		{
			Code: `
class Test {
  public constructor(private readonly foo: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "explicit",
				"overrides":     map[string]interface{}{"parameterProperties": "explicit"},
			},
		},
		{
			Code: `
class Test {
  public constructor(private foo: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "explicit",
				"overrides":     map[string]interface{}{"parameterProperties": "off"},
			},
		},
		{
			Code: `
class Test {
  public constructor(protected foo: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "explicit",
				"overrides":     map[string]interface{}{"parameterProperties": "off"},
			},
		},
		{
			Code: `
class Test {
  public constructor(public foo: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "explicit",
				"overrides":     map[string]interface{}{"parameterProperties": "off"},
			},
		},
		{
			Code: `
class Test {
  public constructor(readonly foo: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "explicit",
				"overrides":     map[string]interface{}{"parameterProperties": "off"},
			},
		},
		{
			Code: `
class Test {
  public constructor(private readonly foo: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "explicit",
				"overrides":     map[string]interface{}{"parameterProperties": "off"},
			},
		},
		// ---- default options ----
		{
			Code: `
class Test {
  protected name: string;
  private x: number;
  public getX() {
    return this.x;
  }
}
      `,
		},
		{
			Code: `
class Test {
  protected name: string;
  protected foo?: string;
  public 'foo-bar'?: string;
}
      `,
		},
		{
			Code: `
class Test {
  public constructor({ x, y }: { x: number; y: number }) {}
}
      `,
		},
		{
			Code: `
class Test {
  protected name: string;
  protected foo?: string;
  public getX() {
    return this.x;
  }
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
		},
		// ---- accessibility: 'no-public' ----
		{
			Code: `
class Test {
  protected name: string;
  protected foo?: string;
  getX() {
    return this.x;
  }
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
		},
		{
			Code: `
class Test {
  name: string;
  foo?: string;
  getX() {
    return this.x;
  }
  get fooName(): string {
    return this.foo + ' ' + this.name;
  }
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
		},
		// ---- overrides: { accessors / constructors / methods } ----
		{
			Code: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  private set internalValue(value: number) {
    this.x = value;
  }
  public square(): number {
    return this.x * this.x;
  }
}
      `,
			Options: map[string]interface{}{"overrides": map[string]interface{}{"accessors": "off", "constructors": "off"}},
		},
		{
			Code: `
class Test {
  private x: number;
  public constructor(x: number) {
    this.x = x;
  }
  public get internalValue() {
    return this.x;
  }
  public set internalValue(value: number) {
    this.x = value;
  }
  public square(): number {
    return this.x * this.x;
  }
  half(): number {
    return this.x / 2;
  }
}
      `,
			Options: map[string]interface{}{"overrides": map[string]interface{}{"methods": "off"}},
		},
		// ---- parameter properties under accessibility: 'no-public' ----
		{
			Code: `
class Test {
  constructor(private x: number) {}
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
		},
		{
			Code: `
class Test {
  constructor(public x: number) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "no-public",
				"overrides":     map[string]interface{}{"parameterProperties": "off"},
			},
		},
		{
			Code: `
class Test {
  constructor(public foo: number) {}
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
		},
		// ---- ignoredMethodNames ----
		{
			Code: `
class Test {
  public getX() {
    return this.x;
  }
}
      `,
			Options: map[string]interface{}{"ignoredMethodNames": []interface{}{"getX"}},
		},
		{
			Code: `
class Test {
  public static getX() {
    return this.x;
  }
}
      `,
			Options: map[string]interface{}{"ignoredMethodNames": []interface{}{"getX"}},
		},
		{
			Code: `
class Test {
  get getX() {
    return this.x;
  }
}
      `,
			Options: map[string]interface{}{"ignoredMethodNames": []interface{}{"getX"}},
		},
		{
			Code: `
class Test {
  getX() {
    return this.x;
  }
}
      `,
			Options: map[string]interface{}{"ignoredMethodNames": []interface{}{"getX"}},
		},
		// ---- overrides.properties ----
		{
			Code: `
class Test {
  x = 2;
}
      `,
			Options: map[string]interface{}{"overrides": map[string]interface{}{"properties": "off"}},
		},
		{
			Code: `
class Test {
  private x = 2;
}
      `,
			Options: map[string]interface{}{"overrides": map[string]interface{}{"properties": "explicit"}},
		},
		{
			Code: `
class Test {
  x = 2;
  private x = 2;
}
      `,
			Options: map[string]interface{}{"overrides": map[string]interface{}{"properties": "no-public"}},
		},
		// ---- private fields (#name) are ignored ----
		{
			Code: `
class Test {
  #foo = 1;
  #bar() {}
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
		},
		// ---- accessor / abstract accessor with private modifier ----
		{
			Code: `
class Test {
  private accessor foo = 1;
}
      `,
		},
		{
			Code: `
abstract class Test {
  private abstract accessor foo: number;
}
      `,
		},
		// ---- Lock-in tests for branches typescript-eslint's own suite doesn't cover ----
		// JSON-options path: bare object matches single-element CLI shape.
		{
			Code: `
class Test {
  private foo() {}
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
		},
		// JSON-options path: nil options falls back to default ('explicit'),
		// which means the implicit-public method below is reported INVALID;
		// keep the negative case in the invalid section. Here we lock in that
		// default mode allows explicit modifiers.
		{Code: `class Test { public foo() {} }`},
		// Computed property name: locks the AST shape for ComputedPropertyName end position.
		{
			Code: `
class Test {
  public ['x'] = 1;
}
      `,
		},

		// ============================================================
		// tsgo-specific lock-in cases — shapes upstream's ESTree-based
		// suite doesn't exercise but rslint must handle correctly.
		// ============================================================

		// Class expression member: KindClassExpression is a sibling kind
		// to KindClassDeclaration and the listener must visit its members
		// the same way.
		{
			Code: `
const C = class {
  public foo() {}
  private bar = 1;
  public constructor() {}
};
      `,
		},
		// Nested class: outer rule pass should NOT bleed into the inner
		// class's members (each class is independent).
		{
			Code: `
class Outer {
  public makeInner() {
    return class Inner {
      public x = 1;
      private y = 2;
    };
  }
}
      `,
		},
		// Constructor overload signatures: the body-less signature is
		// merged with the implementation; an explicit modifier on the
		// implementation satisfies the rule.
		{
			Code: `
class Test {
  public constructor(x: string);
  public constructor(x: number, y: string);
  public constructor(x: string | number, y?: string) {}
}
      `,
		},
		// Method overload signatures with explicit accessibility on every
		// declaration.
		{
			Code: `
class Test {
  public foo(x: string): void;
  public foo(x: number, y: string): void;
  public foo(x: string | number, y?: string): void {}
}
      `,
		},
		// Numeric key with `no-public` allowed (no accessibility required).
		{
			Code: `
class Test {
  0 = 1;
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
		},
		// Computed key referencing a Symbol — locks ComputedPropertyName
		// handling beyond the simple ['x'] case.
		{
			Code: `
class Test {
  public [Symbol.iterator]() {}
}
      `,
		},
		// declare class with body-less members — `declare` is a modifier
		// flag in tsgo; rule should still see explicit accessibility on
		// each member.
		{
			Code: `
declare class Test {
  public foo: number;
  private bar(): void;
}
      `,
		},
		// override modifier on a subclass member with explicit access.
		{
			Code: `
class Base {
  public foo() {}
}
class Derived extends Base {
  public override foo() {}
}
      `,
		},
		// `static` + accessibility together — order in the modifier list
		// is parser-dependent; with the modifier-flag query the rule must
		// recognize `public` regardless of position.
		{
			Code: `
class Test {
  public static foo = 1;
  static public bar = 2;
}
      `,
		},
		// Interface members must NOT trigger the rule (only class-bound).
		{
			Code: `
interface ITest {
  foo(): void;
  bar: string;
  readonly baz: number;
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
		},
		// Type literal members must NOT trigger.
		{
			Code: `
type T = {
  foo(): void;
  bar: string;
};
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
		},
		// Object literal methods/properties must NOT trigger.
		{
			Code: `
const obj = {
  foo() {},
  bar: 1,
  get baz() { return 1; },
};
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
		},
		// Function-typed class field with explicit accessibility (locks
		// that arrow-function-initialized class fields are still treated
		// as properties, not methods).
		{
			Code: `
class Test {
  public foo = () => 1;
  private bar = function () {};
}
      `,
		},
		// Parameter property with default value — Initializer is set on
		// the Parameter node; Name() must still be the Identifier.
		{
			Code: `
class Test {
  public constructor(public foo: string = 'x') {}
}
      `,
		},
		// Constructor with destructured (non-parameter-property) parameter
		// alongside a real parameter property.
		{
			Code: `
class Test {
  public constructor({ x }: { x: number }, public y: number) {}
}
      `,
		},
		// `parameterProperties: 'no-public'` with bare `public foo` (no
		// readonly): upstream only flags `public readonly` parameter
		// properties because removing `public` from `public foo` would
		// turn it into a regular constructor parameter (silently
		// dropping the parameter-property semantics). Locks that we
		// match this behavior.
		{
			Code: `
class Test {
  public constructor(public foo: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "off",
				"overrides":     map[string]interface{}{"parameterProperties": "no-public"},
			},
		},
		// Bare class field declaration without initializer or type — must
		// still be classified as KindPropertyDeclaration and require a
		// modifier under `explicit`.
		{
			Code: `
class Test {
  public x;
}
      `,
		},
		// `accessor` field with explicit type but no initializer — locks
		// auto-accessor detection across initializer presence.
		{
			Code: `
class Test {
  private accessor foo: number;
}
      `,
		},
		// Non-abstract method inside an abstract class — should behave
		// the same as a method in a regular class.
		{
			Code: `
abstract class Test {
  public foo(): void {}
  protected bar = 1;
}
      `,
		},
		// ---- Real-world scenarios: namespace, JSX, export default ----
		// Class inside a TS namespace.
		{
			Code: `
namespace App {
  export class Foo {
    public bar = 1;
    private baz() {}
  }
}
      `,
		},
		// Class inside a module declaration.
		{
			Code: `
module App {
  export class Foo {
    public bar = 1;
  }
}
      `,
		},
		// `export default class { ... }` — anonymous default export.
		{
			Code: `
export default class {
  public foo() {}
  private bar = 1;
}
      `,
		},
		// JSX/TSX: class in a .tsx file. The TSX parser path must produce
		// the same AST shape so the rule still triggers consistently.
		{
			Code: `
class Component {
  public render() {
    return <div />;
  }
  private name = '';
}
      `,
			Tsx: true,
		},
		// Constructor with super call — modifier check must not be
		// influenced by the body shape.
		{
			Code: `
class Base {
  public constructor(public x: number) {}
}
class Derived extends Base {
  public constructor() {
    super(1);
  }
}
      `,
		},
		// ---- ignoredMethodNames: full name-form coverage ----
		// Multiple ignored names + matches both an Identifier method and
		// a quoted-string method.
		{
			Code: `
class Test {
  getX() {}
  'foo bar'() {}
}
      `,
			Options: map[string]interface{}{"ignoredMethodNames": []interface{}{"getX", `"foo bar"`}},
		},
		// `ignoredMethodNames` applied to a quoted numeric key — the
		// diagnostic name is `"0"` (quoted), so the option must be
		// `"0"` to match.
		{
			Code: `
class Test {
  0() {}
}
      `,
			Options: map[string]interface{}{"ignoredMethodNames": []interface{}{`"0"`}},
		},
		// `ignoredMethodNames` does NOT apply to properties (only methods);
		// upstream uses ignoredMethodNames only inside checkMethod.
		// This locks that we match upstream's scope.
		{
			Code: `
class Test {
  public foo = 1;
  bar() {}
}
      `,
			Options: map[string]interface{}{"ignoredMethodNames": []interface{}{"bar"}},
		},
		// ---- Modifier combinations: optional / definite / override ----
		// Optional class field `?`.
		{
			Code: `
class Test {
  public foo?: number;
  private bar?: string;
}
      `,
		},
		// Definite assignment assertion `!`.
		{
			Code: `
class Test {
  public foo!: number;
  private bar!: string;
}
      `,
		},
		// `override` + accessibility (subclass member explicitly typed).
		{
			Code: `
class Base {
  public foo() {}
  public bar = 1;
}
class Derived extends Base {
  public override foo() {}
  protected override bar = 2;
}
      `,
		},
		// `override` + readonly parameter property with explicit access.
		{
			Code: `
class Base {
  public constructor(public readonly value: string) {}
}
class Derived extends Base {
  public constructor(public override readonly value: string) {
    super(value);
  }
}
      `,
		},
		// Getter declared without modifier under `accessibility: 'no-public'`,
		// setter explicitly public — only setter is reported.
		{
			Code: `
class Test {
  get value() {
    return 1;
  }
  protected set value(v: number) {}
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
		},
		// ---- async / generator / async-generator methods ----
		// async method with explicit access.
		{
			Code: `
class Test {
  public async foo(): Promise<void> {}
}
      `,
		},
		// generator method with explicit access.
		{
			Code: `
class Test {
  private *foo(): Generator<number> {
    yield 1;
  }
}
      `,
		},
		// async generator method with explicit access.
		{
			Code: `
class Test {
  protected async *foo(): AsyncGenerator<number> {
    yield 1;
  }
}
      `,
		},
		// async + static combined with explicit access.
		{
			Code: `
class Test {
  public static async foo(): Promise<void> {}
}
      `,
		},
		// ---- generic / typed methods ----
		// Generic method with type parameter list.
		{
			Code: `
class Test {
  public foo<T>(x: T): T {
    return x;
  }
}
      `,
		},
		// ---- computed-key variants under `no-public` (no modifier needed) ----
		// Computed numeric key — must be classified as PropertyDeclaration's
		// ComputedPropertyName; rule should not crash.
		{
			Code: `
class Test {
  [0] = 1;
  [1n] = 2;
  ['x'] = 3;
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
		},
		// Computed key from a template literal.
		{
			Code: "\nclass Test {\n  [`foo`] = 1;\n}\n      ",
			Options: map[string]interface{}{"accessibility": "no-public"},
		},
		// ---- static + accessor / abstract + accessor + override ----
		// `static accessor` field — static auto-accessor with explicit access.
		{
			Code: `
class Test {
  public static accessor foo = 1;
}
      `,
		},
		// abstract class with computed-key abstract method.
		{
			Code: `
abstract class Test {
  public abstract ['x'](): void;
}
      `,
		},
		// ---- empty class body / structural edge cases ----
		// Empty class body — listener should never fire.
		{
			Code: `
class Test {}
class Other {
  public foo() {}
}
      `,
		},
		// Class field initialized to a class expression — the inner class's
		// members are checked independently.
		{
			Code: `
class Outer {
  public Inner = class {
    public bar() {}
  };
}
      `,
		},
		// Ambient module containing a class — listener should still fire
		// inside.
		{
			Code: `
declare module 'pkg' {
  class Foo {
    public bar(): void;
  }
}
      `,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- parameterProperties: 'explicit' ----
		{
			Code: `
export class XXXX {
  public constructor(readonly value: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "off",
				"overrides":     map[string]interface{}{"parameterProperties": "explicit"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 22, EndLine: 3, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
export class XXXX {
  public constructor(public readonly value: string) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
export class XXXX {
  public constructor(private readonly value: string) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
export class XXXX {
  public constructor(protected readonly value: string) {}
}
      `},
				}},
			},
		},
		{
			Code: `
export class WithParameterProperty {
  public constructor(readonly value: string) {}
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 22, EndLine: 3, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
export class WithParameterProperty {
  public constructor(public readonly value: string) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
export class WithParameterProperty {
  public constructor(private readonly value: string) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
export class WithParameterProperty {
  public constructor(protected readonly value: string) {}
}
      `},
				}},
			},
		},
		{
			Code: `
export class XXXX {
  public constructor(readonly samosa: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "off",
				"overrides":     map[string]interface{}{"constructors": "explicit", "parameterProperties": "explicit"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 22, EndLine: 3, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
export class XXXX {
  public constructor(public readonly samosa: string) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
export class XXXX {
  public constructor(private readonly samosa: string) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
export class XXXX {
  public constructor(protected readonly samosa: string) {}
}
      `},
				}},
			},
		},
		{
			Code: `
class Test {
  public constructor(readonly foo: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "explicit",
				"overrides":     map[string]interface{}{"parameterProperties": "explicit"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 22, EndLine: 3, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public constructor(public readonly foo: string) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public constructor(private readonly foo: string) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public constructor(protected readonly foo: string) {}
}
      `},
				}},
			},
		},
		// ---- explicit on class fields and methods ----
		{
			Code: `
class Test {
  x: number;
  public getX() {
    return this.x;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public x: number;
  public getX() {
    return this.x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  public getX() {
    return this.x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected x: number;
  public getX() {
    return this.x;
  }
}
      `},
				}},
			},
		},
		{
			Code: `
class Test {
  private x: number;
  getX() {
    return this.x;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  public getX() {
    return this.x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  private getX() {
    return this.x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  protected getX() {
    return this.x;
  }
}
      `},
				}},
			},
		},
		{
			Code: `
class Test {
  x?: number;
  getX?() {
    return this.x;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public x?: number;
  getX?() {
    return this.x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x?: number;
  getX?() {
    return this.x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected x?: number;
  getX?() {
    return this.x;
  }
}
      `},
				}},
				{MessageId: "missingAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  x?: number;
  public getX?() {
    return this.x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  x?: number;
  private getX?() {
    return this.x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  x?: number;
  protected getX?() {
    return this.x;
  }
}
      `},
				}},
			},
		},
		// ---- no-public reports unwantedPublicAccessibility with autofix ----
		{
			Code: `
class Test {
  protected name: string;
  protected foo?: string;
  public getX() {
    return this.x;
  }
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  protected name: string;
  protected foo?: string;
  getX() {
    return this.x;
  }
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 5, Column: 3, EndLine: 5, EndColumn: 9},
			},
		},
		{
			Code: `
class Test {
  protected name: string;
  public foo?: string;
  getX() {
    return this.x;
  }
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  protected name: string;
  foo?: string;
  getX() {
    return this.x;
  }
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 9},
			},
		},
		{
			Code: `
class Test {
  public x: number;
  public getX() {
    return this.x;
  }
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  x: number;
  getX() {
    return this.x;
  }
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
				{MessageId: "unwantedPublicAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 9},
			},
		},
		// ---- overrides.constructors: 'no-public' (constructor allowed without modifier, but accessors require it) ----
		{
			Code: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  set internalValue(value: number) {
    this.x = value;
  }
}
      `,
			Options: map[string]interface{}{"overrides": map[string]interface{}{"constructors": "no-public"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 7, Column: 3, EndLine: 7, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  public get internalValue() {
    return this.x;
  }
  set internalValue(value: number) {
    this.x = value;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  private get internalValue() {
    return this.x;
  }
  set internalValue(value: number) {
    this.x = value;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  protected get internalValue() {
    return this.x;
  }
  set internalValue(value: number) {
    this.x = value;
  }
}
      `},
				}},
				{MessageId: "missingAccessibility", Line: 10, Column: 3, EndLine: 10, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  public set internalValue(value: number) {
    this.x = value;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  private set internalValue(value: number) {
    this.x = value;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  protected set internalValue(value: number) {
    this.x = value;
  }
}
      `},
				}},
			},
		},
		// ---- default options: constructor + getter + setter ----
		{
			Code: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  set internalValue(value: number) {
    this.x = value;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  public constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  set internalValue(value: number) {
    this.x = value;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  private constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  set internalValue(value: number) {
    this.x = value;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  protected constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  set internalValue(value: number) {
    this.x = value;
  }
}
      `},
				}},
				{MessageId: "missingAccessibility", Line: 7, Column: 3, EndLine: 7, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  public get internalValue() {
    return this.x;
  }
  set internalValue(value: number) {
    this.x = value;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  private get internalValue() {
    return this.x;
  }
  set internalValue(value: number) {
    this.x = value;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  protected get internalValue() {
    return this.x;
  }
  set internalValue(value: number) {
    this.x = value;
  }
}
      `},
				}},
				{MessageId: "missingAccessibility", Line: 10, Column: 3, EndLine: 10, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  public set internalValue(value: number) {
    this.x = value;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  private set internalValue(value: number) {
    this.x = value;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x: number;
  constructor(x: number) {
    this.x = x;
  }
  get internalValue() {
    return this.x;
  }
  protected set internalValue(value: number) {
    this.x = value;
  }
}
      `},
				}},
			},
		},
		// ---- parameter property as missing in constructor with overrides.parameterProperties: 'no-public' (only constructor reported) ----
		{
			Code: `
class Test {
  constructor(public x: number) {}
  public foo(): string {
    return 'foo';
  }
}
      `,
			Options: map[string]interface{}{"overrides": map[string]interface{}{"parameterProperties": "no-public"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public constructor(public x: number) {}
  public foo(): string {
    return 'foo';
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private constructor(public x: number) {}
  public foo(): string {
    return 'foo';
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected constructor(public x: number) {}
  public foo(): string {
    return 'foo';
  }
}
      `},
				}},
			},
		},
		// ---- default: explicit on constructor with parameter property ----
		{
			Code: `
class Test {
  constructor(public x: number) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public constructor(public x: number) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private constructor(public x: number) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected constructor(public x: number) {}
}
      `},
				}},
			},
		},
		// ---- parameterProperties: 'no-public' on `public readonly` ----
		{
			Code: `
class Test {
  constructor(public readonly x: number) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "off",
				"overrides":     map[string]interface{}{"parameterProperties": "no-public"},
			},
			Output: []string{`
class Test {
  constructor(readonly x: number) {}
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 15, EndLine: 3, EndColumn: 21},
			},
		},
		// ---- overrides.properties: 'explicit' / 'no-public' ----
		{
			Code: `
class Test {
  x = 2;
}
      `,
			Options: map[string]interface{}{
				"accessibility": "off",
				"overrides":     map[string]interface{}{"properties": "explicit"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public x = 2;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x = 2;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected x = 2;
}
      `},
				}},
			},
		},
		{
			Code: `
class Test {
  public x = 2;
  private x = 2;
}
      `,
			Options: map[string]interface{}{
				"accessibility": "off",
				"overrides":     map[string]interface{}{"properties": "no-public"},
			},
			Output: []string{`
class Test {
  x = 2;
  private x = 2;
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
			},
		},
		// ---- explicit on constructor with `public x: any[]` parameter property ----
		{
			Code: `
class Test {
  constructor(public x: any[]) {}
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public constructor(public x: any[]) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private constructor(public x: any[]) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected constructor(public x: any[]) {}
}
      `},
				}},
			},
		},
		// ---- public-keyword removal preserves trailing comment / decorators ----
		{
			Code: `
class Test {
  public /*public*/constructor(private foo: string) {}
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  /*public*/constructor(private foo: string) {}
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
			},
		},
		{
			Code: `
class Test {
  @public
  public foo() {}
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  @public
  foo() {}
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 9},
			},
		},
		{
			Code: `
class Test {
  @public
  public foo;
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  @public
  foo;
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 9},
			},
		},
		{
			Code: `
class Test {
  public foo = '';
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  foo = '';
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
			},
		},
		// ---- removal stops at comment, no whitespace inserted ----
		{
			Code: `
class Test {
  constructor(public/* Hi there */ readonly foo) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "no-public",
				"overrides":     map[string]interface{}{"parameterProperties": "no-public"},
			},
			Output: []string{`
class Test {
  constructor(/* Hi there */ readonly foo) {}
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 15, EndLine: 3, EndColumn: 21},
			},
		},
		{
			Code: `
class Test {
  constructor(public readonly foo: string) {}
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  constructor(readonly foo: string) {}
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 15, EndLine: 3, EndColumn: 21},
			},
		},
		{
			Code: `
class EnsureWhiteSPaceSpan {
  public constructor() {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "no-public",
				"overrides":     map[string]interface{}{"parameterProperties": "no-public"},
			},
			Output: []string{`
class EnsureWhiteSPaceSpan {
  constructor() {}
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
			},
		},
		{
			Code: `
class EnsureWhiteSPaceSpan {
  public /* */ constructor() {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "no-public",
				"overrides":     map[string]interface{}{"parameterProperties": "no-public"},
			},
			Output: []string{`
class EnsureWhiteSPaceSpan {
  /* */ constructor() {}
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
			},
		},
		// ---- quoted member names ----
		{
			Code: `
class Test {
  public 'foo' = 1;
  public 'foo foo' = 2;
  public 'bar'() {}
  public 'bar bar'() {}
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  'foo' = 1;
  'foo foo' = 2;
  'bar'() {}
  'bar bar'() {}
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 9, Message: "Public accessibility modifier on class property foo."},
				{MessageId: "unwantedPublicAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 9, Message: `Public accessibility modifier on class property "foo foo".`},
				{MessageId: "unwantedPublicAccessibility", Line: 5, Column: 3, EndLine: 5, EndColumn: 9, Message: "Public accessibility modifier on method definition bar."},
				{MessageId: "unwantedPublicAccessibility", Line: 6, Column: 3, EndLine: 6, EndColumn: 9, Message: `Public accessibility modifier on method definition "bar bar".`},
			},
		},
		// ---- abstract methods/fields ----
		{
			Code: `
abstract class SomeClass {
  abstract method(): string;
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  public abstract method(): string;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  private abstract method(): string;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  protected abstract method(): string;
}
      `},
				}},
			},
		},
		{
			Code: `
abstract class SomeClass {
  public abstract method(): string;
}
      `,
			Options: map[string]interface{}{
				"accessibility": "no-public",
				"overrides":     map[string]interface{}{"parameterProperties": "no-public"},
			},
			Output: []string{`
abstract class SomeClass {
  abstract method(): string;
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
			},
		},
		{
			Code: `
abstract class SomeClass {
  abstract x: string;
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  public abstract x: string;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  private abstract x: string;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  protected abstract x: string;
}
      `},
				}},
			},
		},
		{
			Code: `
abstract class SomeClass {
  public abstract x: string;
}
      `,
			Options: map[string]interface{}{
				"accessibility": "no-public",
				"overrides":     map[string]interface{}{"parameterProperties": "no-public"},
			},
			Output: []string{`
abstract class SomeClass {
  abstract x: string;
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
			},
		},
		// ---- accessor properties ----
		{
			Code: `
class SomeClass {
  accessor foo = 1;
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class SomeClass {
  public accessor foo = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class SomeClass {
  private accessor foo = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class SomeClass {
  protected accessor foo = 1;
}
      `},
				}},
			},
		},
		{
			Code: `
abstract class SomeClass {
  abstract accessor foo: string;
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  public abstract accessor foo: string;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  private abstract accessor foo: string;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  protected abstract accessor foo: string;
}
      `},
				}},
			},
		},
		// ---- decorated class members ----
		{
			Code: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  public constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  private constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  protected constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
				}},
				{MessageId: "missingAccessibility", Line: 3, Column: 27, EndLine: 3, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() public readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() private readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() protected readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
				}},
				{MessageId: "missingAccessibility", Line: 4, Column: 15, EndLine: 4, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() public x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() private x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() protected x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
				}},
				{MessageId: "missingAccessibility", Line: 5, Column: 15, EndLine: 5, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() public getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() private getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() protected getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
				}},
				{MessageId: "missingAccessibility", Line: 10, Column: 3, EndLine: 10, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  public get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  private get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  protected get y() {
    return this.x;
  }
  @foo @bar() set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
				}},
				{MessageId: "missingAccessibility", Line: 13, Column: 15, EndLine: 13, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() public set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() private set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class DecoratedClass {
  constructor(@foo @bar() readonly arg: string) {}
  @foo @bar() x: string;
  @foo @bar() getX() {
    return this.x;
  }
  @foo
  @bar()
  get y() {
    return this.x;
  }
  @foo @bar() protected set z(@foo @bar() value: x) {
    this.x = x;
  }
}
      `},
				}},
			},
		},
		// ---- computed-name abstract methods (the head loc must include the `]`) ----
		{
			Code: `
abstract class SomeClass {
  abstract ['computed-method-name'](): string;
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  public abstract ['computed-method-name'](): string;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  private abstract ['computed-method-name'](): string;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
abstract class SomeClass {
  protected abstract ['computed-method-name'](): string;
}
      `},
				}},
			},
		},

		// ============================================================
		// tsgo-specific invalid lock-in cases.
		// `missingAccessibility` always carries 3 suggestions; we use
		// short snippets here to keep the assertions readable while still
		// covering the cases.
		// ============================================================

		// Class expression members are reported the same as class declarations.
		{
			Code: `
const C = class {
  foo() {}
  bar = 1;
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
const C = class {
  public foo() {}
  bar = 1;
};
      `},
					{MessageId: "addExplicitAccessibility", Output: `
const C = class {
  private foo() {}
  bar = 1;
};
      `},
					{MessageId: "addExplicitAccessibility", Output: `
const C = class {
  protected foo() {}
  bar = 1;
};
      `},
				}},
				{MessageId: "missingAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
const C = class {
  foo() {}
  public bar = 1;
};
      `},
					{MessageId: "addExplicitAccessibility", Output: `
const C = class {
  foo() {}
  private bar = 1;
};
      `},
					{MessageId: "addExplicitAccessibility", Output: `
const C = class {
  foo() {}
  protected bar = 1;
};
      `},
				}},
			},
		},
		// `public` removal must keep the rest of the modifier list intact
		// (here: `static`). Modifier-list ordering / preservation is a
		// tsgo-only concern (ESLint sees flat modifier flags).
		{
			Code: `
class Test {
  public static foo = 1;
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  static foo = 1;
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
			},
		},
		// `static` member missing accessibility — the head loc should
		// start at `static` (first non-decorator modifier), end at name end.
		{
			Code: `
class Test {
  static foo = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public static foo = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private static foo = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected static foo = 1;
}
      `},
				}},
			},
		},
		// declare class members are still checked.
		{
			Code: `
declare class Test {
  foo: number;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
declare class Test {
  public foo: number;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
declare class Test {
  private foo: number;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
declare class Test {
  protected foo: number;
}
      `},
				}},
			},
		},
		// Object literal must NOT report — `foo() {}` and `bar: 1` are
		// not class members. Only the class field reports.
		{
			Code: `
const obj = { foo() {}, bar: 1 };
class C {
  baz = 1;
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
const obj = { foo() {}, bar: 1 };
class C {
  public baz = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
const obj = { foo() {}, bar: 1 };
class C {
  private baz = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
const obj = { foo() {}, bar: 1 };
class C {
  protected baz = 1;
}
      `},
				}},
			},
		},
		// Interface members must NOT be reported.
		{
			Code: `
interface ITest {
  foo(): void;
  bar: string;
}
class C {
  baz = 1;
}
      `,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 7, Column: 3, EndLine: 7, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
interface ITest {
  foo(): void;
  bar: string;
}
class C {
  public baz = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
interface ITest {
  foo(): void;
  bar: string;
}
class C {
  private baz = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
interface ITest {
  foo(): void;
  bar: string;
}
class C {
  protected baz = 1;
}
      `},
				}},
			},
		},
		// Numeric-literal property name — `requiresQuoting` returns true for
		// digit-leading text, so the diagnostic name is quoted.
		{
			Code: `
class Test {
  0 = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 4,
					Message: `Missing accessibility modifier on class property "0".`,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public 0 = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private 0 = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected 0 = 1;
}
      `},
					},
				},
			},
		},
		// Decorator-only modifier list (no other modifier): the head loc
		// start must be the next token after the last decorator, and the
		// suggestion must insert before that next token (preserving the
		// decorator).
		{
			Code: `
class Test {
  @foo
  bar = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  @foo
  public bar = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  @foo
  private bar = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  @foo
  protected bar = 1;
}
      `},
				}},
			},
		},
		// `accessor` field with no initializer / no type — must still
		// require a modifier under `explicit`.
		{
			Code: `
class Test {
  accessor foo;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public accessor foo;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private accessor foo;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected accessor foo;
}
      `},
				}},
			},
		},
		// Bare field declaration `x;` (no initializer, no type) — locks
		// the head loc range (just the identifier) and the suggestion
		// insertion point.
		{
			Code: `
class Test {
  x;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public x;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private x;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected x;
}
      `},
				}},
			},
		},
		// ---- Real-world scenarios: namespace / JSX / export default ----
		// Class inside namespace — listener still fires on inner members.
		{
			Code: `
namespace App {
  export class Foo {
    bar = 1;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 4, Column: 5, EndLine: 4, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
namespace App {
  export class Foo {
    public bar = 1;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
namespace App {
  export class Foo {
    private bar = 1;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
namespace App {
  export class Foo {
    protected bar = 1;
  }
}
      `},
				}},
			},
		},
		// `export default class { ... }` with an un-modified member.
		{
			Code: `
export default class {
  foo() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
export default class {
  public foo() {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
export default class {
  private foo() {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
export default class {
  protected foo() {}
}
      `},
				}},
			},
		},
		// Class inside a .tsx file — render method missing accessibility.
		{
			Code: `
class Component {
  render() {
    return <div />;
  }
}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Component {
  public render() {
    return <div />;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Component {
  private render() {
    return <div />;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Component {
  protected render() {
    return <div />;
  }
}
      `},
				}},
			},
		},
		// ---- ignoredMethodNames: applies to METHODS only, not properties ----
		// `ignoredMethodNames: ['bar']` must NOT silence a property named `bar`.
		{
			Code: `
class Test {
  bar = 1;
}
      `,
			Options: map[string]interface{}{"ignoredMethodNames": []interface{}{"bar"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public bar = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private bar = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected bar = 1;
}
      `},
				}},
			},
		},
		// ---- Modifier combinations ----
		// Optional `?` field — head loc must include only the identifier
		// (not the `?`), matching upstream's `node.key.loc.end`.
		{
			Code: `
class Test {
  foo?: number;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public foo?: number;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private foo?: number;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected foo?: number;
}
      `},
				}},
			},
		},
		// Definite assignment assertion `!` — same as optional, the head
		// loc stops at the identifier.
		{
			Code: `
class Test {
  foo!: number;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public foo!: number;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private foo!: number;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected foo!: number;
}
      `},
				}},
			},
		},
		// `override` + missing accessibility — head loc anchors at
		// `override` (first non-decorator modifier).
		{
			Code: `
class Base {
  public foo() {}
}
class Derived extends Base {
  override foo() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 6, Column: 3, EndLine: 6, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Base {
  public foo() {}
}
class Derived extends Base {
  public override foo() {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Base {
  public foo() {}
}
class Derived extends Base {
  private override foo() {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Base {
  public foo() {}
}
class Derived extends Base {
  protected override foo() {}
}
      `},
				}},
			},
		},
		// `public override readonly` parameter property under
		// `parameterProperties: 'no-public'` — `public` is removed but
		// `override readonly` stay. Two parameter properties → two
		// reports + a single autofix pass that removes both.
		{
			Code: `
class Base {
  constructor(public readonly value: string) {}
}
class Derived extends Base {
  constructor(public override readonly value: string) {
    super(value);
  }
}
      `,
			Options: map[string]interface{}{
				"accessibility": "off",
				"overrides":     map[string]interface{}{"parameterProperties": "no-public"},
			},
			Output: []string{`
class Base {
  constructor(readonly value: string) {}
}
class Derived extends Base {
  constructor(override readonly value: string) {
    super(value);
  }
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unwantedPublicAccessibility", Line: 3, Column: 15, EndLine: 3, EndColumn: 21},
				{MessageId: "unwantedPublicAccessibility", Line: 6, Column: 15, EndLine: 6, EndColumn: 21},
			},
		},
		// ---- async / generator / async generator methods ----
		// async method missing accessibility — head loc starts at `async`
		// (first non-decorator modifier), ends at name end.
		{
			Code: `
class Test {
  async foo(): Promise<void> {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public async foo(): Promise<void> {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private async foo(): Promise<void> {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected async foo(): Promise<void> {}
}
      `},
				}},
			},
		},
		// Generator method missing accessibility — `*` token is part of
		// the method head; end is the name end (so `*` is not in the loc
		// because there is no modifier before `*`, only the name comes
		// after; the head loc covers the name only).
		{
			Code: `
class Test {
  *foo(): Generator<number> {
    yield 1;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public *foo(): Generator<number> {
    yield 1;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private *foo(): Generator<number> {
    yield 1;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected *foo(): Generator<number> {
    yield 1;
  }
}
      `},
				}},
			},
		},
		// Async generator method missing accessibility.
		{
			Code: `
class Test {
  async *foo(): AsyncGenerator<number> {
    yield 1;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public async *foo(): AsyncGenerator<number> {
    yield 1;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private async *foo(): AsyncGenerator<number> {
    yield 1;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected async *foo(): AsyncGenerator<number> {
    yield 1;
  }
}
      `},
				}},
			},
		},
		// async + static combined missing accessibility — head loc starts
		// at the first non-decorator modifier (`static` or `async`,
		// whichever comes first in source).
		{
			Code: `
class Test {
  static async foo(): Promise<void> {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public static async foo(): Promise<void> {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private static async foo(): Promise<void> {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected static async foo(): Promise<void> {}
}
      `},
				}},
			},
		},
		// ---- generic method missing accessibility ----
		{
			Code: `
class Test {
  foo<T>(x: T): T {
    return x;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public foo<T>(x: T): T {
    return x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private foo<T>(x: T): T {
    return x;
  }
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected foo<T>(x: T): T {
    return x;
  }
}
      `},
				}},
			},
		},
		// ---- computed-key edge cases ----
		// Computed BigInt key — head loc must include closing `]`.
		{
			Code: `
class Test {
  [1n] = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public [1n] = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private [1n] = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected [1n] = 1;
}
      `},
				}},
			},
		},
		// Computed template-literal key.
		{
			Code: "\nclass Test {\n  [`foo`] = 1;\n}\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: "\nclass Test {\n  public [`foo`] = 1;\n}\n      "},
					{MessageId: "addExplicitAccessibility", Output: "\nclass Test {\n  private [`foo`] = 1;\n}\n      "},
					{MessageId: "addExplicitAccessibility", Output: "\nclass Test {\n  protected [`foo`] = 1;\n}\n      "},
				}},
			},
		},
		// ---- static + accessor / inner class via class field ----
		// `static accessor foo = 1` missing accessibility.
		{
			Code: `
class Test {
  static accessor foo = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public static accessor foo = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private static accessor foo = 1;
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected static accessor foo = 1;
}
      `},
				}},
			},
		},
		// Inner class via class field initializer — both the outer field
		// and inner class member should be reported when both are
		// un-modified.
		{
			Code: `
class Outer {
  Inner = class {
    bar() {}
  };
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Outer {
  public Inner = class {
    bar() {}
  };
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Outer {
  private Inner = class {
    bar() {}
  };
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Outer {
  protected Inner = class {
    bar() {}
  };
}
      `},
				}},
				{MessageId: "missingAccessibility", Line: 4, Column: 5, EndLine: 4, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Outer {
  Inner = class {
    public bar() {}
  };
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Outer {
  Inner = class {
    private bar() {}
  };
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Outer {
  Inner = class {
    protected bar() {}
  };
}
      `},
				}},
			},
		},
		// Method overload signatures — each signature is its own
		// KindMethodDeclaration; each is independently checked.
		{
			Code: `
class Test {
  foo(x: string): void;
  foo(x: string | number) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAccessibility", Line: 3, Column: 3, EndLine: 3, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public foo(x: string): void;
  foo(x: string | number) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private foo(x: string): void;
  foo(x: string | number) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected foo(x: string): void;
  foo(x: string | number) {}
}
      `},
				}},
				{MessageId: "missingAccessibility", Line: 4, Column: 3, EndLine: 4, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  foo(x: string): void;
  public foo(x: string | number) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  foo(x: string): void;
  private foo(x: string | number) {}
}
      `},
					{MessageId: "addExplicitAccessibility", Output: `
class Test {
  foo(x: string): void;
  protected foo(x: string | number) {}
}
      `},
				}},
			},
		},

		// ============================================================
		// Message-text lock-ins. These dedicated cases assert
		// `InvalidTestCaseError.Message` (exact string match) for every
		// modifier-combination / name-form path, so any future
		// regression in getMemberName / nodeType / message templates is
		// caught immediately. Real-world differential testing on rspack
		// already surfaced two bugs whose go test suite was silent
		// because Message was never asserted.
		// ============================================================

		// Constructor — missingAccessibility name must be "constructor".
		{
			Code: `
class Test {
  constructor() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   "Missing accessibility modifier on method definition constructor.",
					Line:      3, Column: 3, EndLine: 3, EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public constructor() {}
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private constructor() {}
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected constructor() {}
}
      `},
					},
				},
			},
		},
		// Constructor — unwantedPublicAccessibility name must be "constructor".
		{
			Code: `
class Test {
  public constructor() {}
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  constructor() {}
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unwantedPublicAccessibility",
					Message:   "Public accessibility modifier on method definition constructor.",
					Line:      3, Column: 3, EndLine: 3, EndColumn: 9,
				},
			},
		},
		// Computed string-literal key — must be unwrapped and quoted.
		{
			Code: `
abstract class Test {
  abstract ['computed-method-name'](): void;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   `Missing accessibility modifier on method definition "computed-method-name".`,
					Line:      3, Column: 3, EndLine: 3, EndColumn: 36,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
abstract class Test {
  public abstract ['computed-method-name'](): void;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
abstract class Test {
  private abstract ['computed-method-name'](): void;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
abstract class Test {
  protected abstract ['computed-method-name'](): void;
}
      `},
					},
				},
			},
		},
		// Computed identifier key — unwrap to the inner Identifier
		// (no brackets, no quotes).
		{
			Code: `
declare const FOO: unique symbol;
class Test {
  [FOO] = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   "Missing accessibility modifier on class property FOO.",
					Line:      4, Column: 3, EndLine: 4, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
declare const FOO: unique symbol;
class Test {
  public [FOO] = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
declare const FOO: unique symbol;
class Test {
  private [FOO] = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
declare const FOO: unique symbol;
class Test {
  protected [FOO] = 1;
}
      `},
					},
				},
			},
		},
		// Computed MemberExpression key — name must be the source text
		// of the inner expression (no brackets).
		{
			Code: `
class Test {
  [Symbol.iterator]() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   "Missing accessibility modifier on method definition Symbol.iterator.",
					Line:      3, Column: 3, EndLine: 3, EndColumn: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public [Symbol.iterator]() {}
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private [Symbol.iterator]() {}
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected [Symbol.iterator]() {}
}
      `},
					},
				},
			},
		},
		// BigInt literal key — must strip the `n` suffix and quote.
		// Direct `1n = ...` syntax: tsgo parses this as a property
		// declaration with a BigIntLiteral key.
		{
			Code: `
class Test {
  1n = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   `Missing accessibility modifier on class property "1".`,
					Line:      3, Column: 3, EndLine: 3, EndColumn: 5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public 1n = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private 1n = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected 1n = 1;
}
      `},
					},
				},
			},
		},
		// Computed BigInt — unwrap to inner BigIntLiteral, name = `"1"`.
		{
			Code: `
class Test {
  [1n] = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   `Missing accessibility modifier on class property "1".`,
					Line:      3, Column: 3, EndLine: 3, EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public [1n] = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private [1n] = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected [1n] = 1;
}
      `},
					},
				},
			},
		},
		// Abstract method — message uses "method definition <name>".
		{
			Code: `
abstract class Test {
  abstract foo(): void;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   "Missing accessibility modifier on method definition foo.",
					Line:      3, Column: 3, EndLine: 3, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
abstract class Test {
  public abstract foo(): void;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
abstract class Test {
  private abstract foo(): void;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
abstract class Test {
  protected abstract foo(): void;
}
      `},
					},
				},
			},
		},
		// Abstract field — message uses "class property".
		{
			Code: `
abstract class Test {
  abstract foo: number;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   "Missing accessibility modifier on class property foo.",
					Line:      3, Column: 3, EndLine: 3, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
abstract class Test {
  public abstract foo: number;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
abstract class Test {
  private abstract foo: number;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
abstract class Test {
  protected abstract foo: number;
}
      `},
					},
				},
			},
		},
		// Auto-accessor field — also "class property".
		{
			Code: `
class Test {
  accessor foo = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   "Missing accessibility modifier on class property foo.",
					Line:      3, Column: 3, EndLine: 3, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public accessor foo = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private accessor foo = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected accessor foo = 1;
}
      `},
					},
				},
			},
		},
		// Static field — message stays "class property", head loc spans
		// `static <name>`.
		{
			Code: `
class Test {
  static foo = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   "Missing accessibility modifier on class property foo.",
					Line:      3, Column: 3, EndLine: 3, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public static foo = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private static foo = 1;
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected static foo = 1;
}
      `},
					},
				},
			},
		},
		// Getter — node type is "get property accessor".
		{
			Code: `
class Test {
  get value(): number {
    return 1;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   "Missing accessibility modifier on get property accessor value.",
					Line:      3, Column: 3, EndLine: 3, EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public get value(): number {
    return 1;
  }
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private get value(): number {
    return 1;
  }
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected get value(): number {
    return 1;
  }
}
      `},
					},
				},
			},
		},
		// Setter — node type is "set property accessor".
		{
			Code: `
class Test {
  set value(v: number) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   "Missing accessibility modifier on set property accessor value.",
					Line:      3, Column: 3, EndLine: 3, EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public set value(v: number) {}
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  private set value(v: number) {}
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  protected set value(v: number) {}
}
      `},
					},
				},
			},
		},
		// Public getter under no-public — unwantedPublic message uses
		// "get property accessor".
		{
			Code: `
class Test {
  public get value(): number {
    return 1;
  }
}
      `,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Output: []string{`
class Test {
  get value(): number {
    return 1;
  }
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unwantedPublicAccessibility",
					Message:   "Public accessibility modifier on get property accessor value.",
					Line:      3, Column: 3, EndLine: 3, EndColumn: 9,
				},
			},
		},
		// Parameter property — message uses "parameter property <name>".
		{
			Code: `
class Test {
  public constructor(readonly foo: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "explicit",
				"overrides":     map[string]interface{}{"parameterProperties": "explicit"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Message:   "Missing accessibility modifier on parameter property foo.",
					Line:      3, Column: 22, EndLine: 3, EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public constructor(public readonly foo: string) {}
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public constructor(private readonly foo: string) {}
}
      `},
						{MessageId: "addExplicitAccessibility", Output: `
class Test {
  public constructor(protected readonly foo: string) {}
}
      `},
					},
				},
			},
		},
		// Public-readonly parameter property under no-public —
		// unwantedPublic message uses "parameter property".
		{
			Code: `
class Test {
  constructor(public readonly foo: string) {}
}
      `,
			Options: map[string]interface{}{
				"accessibility": "off",
				"overrides":     map[string]interface{}{"parameterProperties": "no-public"},
			},
			Output: []string{`
class Test {
  constructor(readonly foo: string) {}
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unwantedPublicAccessibility",
					Message:   "Public accessibility modifier on parameter property foo.",
					Line:      3, Column: 15, EndLine: 3, EndColumn: 21,
				},
			},
		},
	})
}
