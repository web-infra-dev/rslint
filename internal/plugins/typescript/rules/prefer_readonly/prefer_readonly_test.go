package prefer_readonly

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferReadonlyRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferReadonlyRule, []rule_tester.ValidTestCase{
		// Non-class code
		{Code: `function ignore() {}`},
		{Code: `const ignore = function () {};`},
		{Code: `const ignore = () => {};`},
		{Code: `
const container = { member: true };
container.member;
		`},
		{Code: `
const container = { member: 1 };
+container.member;
		`},
		{Code: `
const container = { member: 1 };
++container.member;
		`},
		{Code: `
const container = { member: 1 };
container.member++;
		`},
		{Code: `
const container = { member: 1 };
-container.member;
		`},
		{Code: `
const container = { member: 1 };
--container.member;
		`},
		{Code: `
const container = { member: 1 };
container.member--;
		`},
		// Empty class
		{Code: `class TestEmpty {}`},
		// Already readonly - private keyword
		{Code: `
class TestReadonlyStatic {
  private static readonly correctlyReadonlyStatic = 7;
}
		`},
		// Already readonly - private field
		{Code: `
class TestReadonlyStatic {
  static readonly #correctlyReadonlyStatic = 7;
}
		`},
		// Modified static in constructor - private keyword
		{Code: `
class TestModifiableStatic {
  private static correctlyModifiableStatic = 7;

  public constructor() {
    TestModifiableStatic.correctlyModifiableStatic += 1;
  }
}
		`},
		// Modified static in constructor - private field
		{Code: `
class TestModifiableStatic {
  static #correctlyModifiableStatic = 7;

  public constructor() {
    TestModifiableStatic.#correctlyModifiableStatic += 1;
  }
}
		`},
		// Already readonly inline - private keyword
		{Code: `
class TestReadonlyInline {
  private readonly correctlyReadonlyInline = 7;
}
		`},
		// Already readonly inline - private field
		{Code: `
class TestReadonlyInline {
  readonly #correctlyReadonlyInline = 7;
}
		`},
		// Readonly delayed (constructor assignment with readonly) - private keyword
		{Code: `
class TestReadonlyDelayed {
  private readonly correctlyReadonlyDelayed = 7;

  public constructor() {
    this.correctlyReadonlyDelayed += 1;
  }
}
		`},
		// Readonly delayed - private field
		{Code: `
class TestReadonlyDelayed {
  readonly #correctlyReadonlyDelayed = 7;

  public constructor() {
    this.#correctlyReadonlyDelayed += 1;
  }
}
		`},
		// Nested class expression: both modified - private keyword
		{Code: `
class TestModifiableInline {
  private correctlyModifiableInline = 7;

  public mutate() {
    this.correctlyModifiableInline += 1;

    return class {
      private correctlyModifiableInline = 7;

      mutate() {
        this.correctlyModifiableInline += 1;
      }
    };
  }
}
		`},
		// Nested class expression: both modified - private field
		{Code: `
class TestModifiableInline {
  #correctlyModifiableInline = 7;

  public mutate() {
    this.#correctlyModifiableInline += 1;

    return class {
      #correctlyModifiableInline = 7;

      mutate() {
        this.#correctlyModifiableInline += 1;
      }
    };
  }
}
		`},
		// Modified in method - private keyword
		{Code: `
class TestModifiableDelayed {
  private correctlyModifiableDelayed = 7;

  public mutate() {
    this.correctlyModifiableDelayed += 1;
  }
}
		`},
		// Modified in method - private field
		{Code: `
class TestModifiableDelayed {
  #correctlyModifiableDelayed = 7;

  public mutate() {
    this.#correctlyModifiableDelayed += 1;
  }
}
		`},
		// Deleted property
		{Code: `
class TestModifiableDeleted {
  private correctlyModifiableDeleted = 7;

  public mutate() {
    delete this.correctlyModifiableDeleted;
  }
}
		`},
		// Modified in constructor arrow function - private keyword
		{Code: `
class TestModifiableWithinConstructor {
  private correctlyModifiableWithinConstructor = 7;

  public constructor() {
    (() => {
      this.correctlyModifiableWithinConstructor += 1;
    })();
  }
}
		`},
		// Modified in constructor arrow function - private field
		{Code: `
class TestModifiableWithinConstructor {
  #correctlyModifiableWithinConstructor = 7;

  public constructor() {
    (() => {
      this.#correctlyModifiableWithinConstructor += 1;
    })();
  }
}
		`},
		// Modified in constructor arrow function (variant naming) - private keyword
		{Code: `
class TestModifiableWithinConstructorArrowFunction {
  private correctlyModifiableWithinConstructorArrowFunction = 7;

  public constructor() {
    (() => {
      this.correctlyModifiableWithinConstructorArrowFunction += 1;
    })();
  }
}
		`},
		// Modified in constructor arrow function (variant naming) - private field
		{Code: `
class TestModifiableWithinConstructorArrowFunction {
  #correctlyModifiableWithinConstructorArrowFunction = 7;

  public constructor() {
    (() => {
      this.#correctlyModifiableWithinConstructorArrowFunction += 1;
    })();
  }
}
		`},
		// Modified via self in constructor function expression - private keyword
		{Code: `
class TestModifiableWithinConstructorInFunctionExpression {
  private correctlyModifiableWithinConstructorInFunctionExpression = 7;

  public constructor() {
    const self = this;

    (() => {
      self.correctlyModifiableWithinConstructorInFunctionExpression += 1;
    })();
  }
}
		`},
		// Modified via self in constructor function expression - private field
		{Code: `
class TestModifiableWithinConstructorInFunctionExpression {
  #correctlyModifiableWithinConstructorInFunctionExpression = 7;

  public constructor() {
    const self = this;

    (() => {
      self.#correctlyModifiableWithinConstructorInFunctionExpression += 1;
    })();
  }
}
		`},
		// Modified via self in constructor get accessor - private keyword
		{Code: `
class TestModifiableWithinConstructorInGetAccessor {
  private correctlyModifiableWithinConstructorInGetAccessor = 7;

  public constructor() {
    const self = this;

    const confusingObject = {
      get accessor() {
        return (self.correctlyModifiableWithinConstructorInGetAccessor += 1);
      },
    };
  }
}
		`},
		// Modified via self in constructor get accessor - private field
		{Code: `
class TestModifiableWithinConstructorInGetAccessor {
  #correctlyModifiableWithinConstructorInGetAccessor = 7;

  public constructor() {
    const self = this;

    const confusingObject = {
      get accessor() {
        return (self.#correctlyModifiableWithinConstructorInGetAccessor += 1);
      },
    };
  }
}
		`},
		// Modified via self in constructor method declaration - private keyword
		{Code: `
class TestModifiableWithinConstructorInMethodDeclaration {
  private correctlyModifiableWithinConstructorInMethodDeclaration = 7;

  public constructor() {
    const self = this;

    const confusingObject = {
      methodDeclaration() {
        self.correctlyModifiableWithinConstructorInMethodDeclaration = 7;
      },
    };
  }
}
		`},
		// Modified via self in constructor method declaration - private field
		{Code: `
class TestModifiableWithinConstructorInMethodDeclaration {
  #correctlyModifiableWithinConstructorInMethodDeclaration = 7;

  public constructor() {
    const self = this;

    const confusingObject = {
      methodDeclaration() {
        self.#correctlyModifiableWithinConstructorInMethodDeclaration = 7;
      },
    };
  }
}
		`},
		// Modified via self in constructor set accessor - private keyword
		{Code: `
class TestModifiableWithinConstructorInSetAccessor {
  private correctlyModifiableWithinConstructorInSetAccessor = 7;

  public constructor() {
    const self = this;

    const confusingObject = {
      set accessor(value: number) {
        self.correctlyModifiableWithinConstructorInSetAccessor += value;
      },
    };
  }
}
		`},
		// Modified via self in constructor set accessor - private field
		{Code: `
class TestModifiableWithinConstructorInSetAccessor {
  #correctlyModifiableWithinConstructorInSetAccessor = 7;

  public constructor() {
    const self = this;

    const confusingObject = {
      set accessor(value: number) {
        self.#correctlyModifiableWithinConstructorInSetAccessor += value;
      },
    };
  }
}
		`},
		// Post decremented via -= : private keyword
		{Code: `
class TestModifiablePostDecremented {
  private correctlyModifiablePostDecremented = 7;

  public mutate() {
    this.correctlyModifiablePostDecremented -= 1;
  }
}
		`},
		// Post decremented via -= : private field
		{Code: `
class TestModifiablePostDecremented {
  #correctlyModifiablePostDecremented = 7;

  public mutate() {
    this.#correctlyModifiablePostDecremented -= 1;
  }
}
		`},
		// Post incremented via += : private keyword
		{Code: `
class TestyModifiablePostIncremented {
  private correctlyModifiablePostIncremented = 7;

  public mutate() {
    this.correctlyModifiablePostIncremented += 1;
  }
}
		`},
		// Post incremented via += : private field
		{Code: `
class TestyModifiablePostIncremented {
  #correctlyModifiablePostIncremented = 7;

  public mutate() {
    this.#correctlyModifiablePostIncremented += 1;
  }
}
		`},
		// Pre decremented via -- : private keyword
		{Code: `
class TestModifiablePreDecremented {
  private correctlyModifiablePreDecremented = 7;

  public mutate() {
    --this.correctlyModifiablePreDecremented;
  }
}
		`},
		// Pre decremented via -- : private field
		{Code: `
class TestModifiablePreDecremented {
  #correctlyModifiablePreDecremented = 7;

  public mutate() {
    --this.#correctlyModifiablePreDecremented;
  }
}
		`},
		// Pre incremented via ++ : private keyword
		{Code: `
class TestModifiablePreIncremented {
  private correctlyModifiablePreIncremented = 7;

  public mutate() {
    ++this.correctlyModifiablePreIncremented;
  }
}
		`},
		// Pre incremented via ++ : private field
		{Code: `
class TestModifiablePreIncremented {
  #correctlyModifiablePreIncremented = 7;

  public mutate() {
    ++this.#correctlyModifiablePreIncremented;
  }
}
		`},
		// Protected and public are not reported
		{Code: `
class TestProtectedModifiable {
  protected protectedModifiable = 7;
}
		`},
		{Code: `
class TestPublicModifiable {
  public publicModifiable = 7;
}
		`},
		// Readonly parameter
		{Code: `
class TestReadonlyParameter {
  public constructor(private readonly correctlyReadonlyParameter = 7) {}
}
		`},
		// Modified parameter
		{Code: `
class TestCorrectlyModifiableParameter {
  public constructor(private correctlyModifiableParameter = 7) {}

  public mutate() {
    this.correctlyModifiableParameter += 1;
  }
}
		`},
		// onlyInlineLambdas: non-lambda is skipped - private keyword
		{
			Code: `
class TestCorrectlyNonInlineLambdas {
  private correctlyNonInlineLambda = 7;
}
			`,
			Options: []interface{}{map[string]interface{}{"onlyInlineLambdas": true}},
		},
		// onlyInlineLambdas: non-lambda is skipped - private field
		{
			Code: `
class TestCorrectlyNonInlineLambdas {
  #correctlyNonInlineLambda = 7;
}
			`,
			Options: []interface{}{map[string]interface{}{"onlyInlineLambdas": true}},
		},
		// Computed property
		{Code: `
class TestComputedParameter {
  public mutate() {
    this['computed'] = 1;
  }
}
		`},
		{Code: `
class TestComputedParameter {
  private ['computed-ignored-by-rule'] = 1;
}
		`},
		// Destructuring assignment - private keyword
		{Code: `
class Foo {
  private value: number = 0;

  bar(newValue: { value: number }) {
    ({ value: this.value } = newValue);
    return this.value;
  }
}
		`},
		// Destructuring assignment - private field
		{Code: `
class Foo {
  #value: number = 0;

  bar(newValue: { value: number }) {
    ({ value: this.#value } = newValue);
    return this.#value;
  }
}
		`},
		// Spread destructuring - private keyword
		{Code: `
class Foo {
  private value: Record<string, number> = {};

  bar(newValue: Record<string, number>) {
    ({ ...this.value } = newValue);
    return this.value;
  }
}
		`},
		// Spread destructuring - private field
		{Code: `
class Foo {
  #value: Record<string, number> = {};

  bar(newValue: Record<string, number>) {
    ({ ...this.#value } = newValue);
    return this.#value;
  }
}
		`},
		// Array spread destructuring - private keyword
		{Code: `
class Foo {
  private value: number[] = [];

  bar(newValue: number[]) {
    [...this.value] = newValue;
    return this.value;
  }
}
		`},
		// Array spread destructuring - private field
		{Code: `
class Foo {
  #value: number[] = [];

  bar(newValue: number[]) {
    [...this.#value] = newValue;
    return this.#value;
  }
}
		`},
		// Array element destructuring - private keyword
		{Code: `
class Foo {
  private value: number = 0;

  bar(newValue: number[]) {
    [this.value] = newValue;
    return this.value;
  }
}
		`},
		// Array element destructuring - private field
		{Code: `
class Foo {
  #value: number = 0;

  bar(newValue: number[]) {
    [this.#value] = newValue;
    return this.#value;
  }
}
		`},
		// Object that is reassigned - private keyword
		{Code: `
class Test {
  private testObj = {
    prop: '',
  };

  public test(): void {
    this.testObj = '';
  }
}
		`},
		// Object that is reassigned - private field
		{Code: `
class Test {
  #testObj = {
    prop: '',
  };

  public test(): void {
    this.#testObj = '';
  }
}
		`},
		// TestObject reassigned - private keyword
		{Code: `
class TestObject {
  public prop: number;
}

class Test {
  private testObj = new TestObject();

  public test(): void {
    this.testObj = new TestObject();
  }
}
		`},
		// TestObject reassigned - private field
		{Code: `
class TestObject {
  public prop: number;
}

class Test {
  #testObj = new TestObject();

  public test(): void {
    this.#testObj = new TestObject();
  }
}
		`},
		// Intersection type
		{Code: `
class TestIntersection {
  private prop: number = 3;

  test() {
    const that = {} as this & { _foo: 'bar' };
    that.prop = 1;
  }
}
		`},
		// Union type
		{Code: `
class TestUnion {
  private prop: number = 3;

  test() {
    const that = {} as this | (this & { _foo: 'bar' });
    that.prop = 1;
  }
}
		`},
		// Static intersection
		{Code: `
class TestStaticIntersection {
  private static prop: number;

  test() {
    const that = {} as typeof TestStaticIntersection & { _foo: 'bar' };
    that.prop = 1;
  }
}
		`},
		// Static union
		{Code: `
class TestStaticUnion {
  private static prop: number = 1;

  test() {
    const that = {} as
      | typeof TestStaticUnion
      | (typeof TestStaticUnion & { _foo: 'bar' });
    that.prop = 1;
  }
}
		`},
		// Both intersection - order 1
		{Code: `
class TestBothIntersection {
  private prop1: number = 1;
  private static prop2: number;

  test() {
    const that = {} as typeof TestBothIntersection & this;
    that.prop1 = 1;
    that.prop2 = 1;
  }
}
		`},
		// Both intersection - order 2
		{Code: `
class TestBothIntersection {
  private prop1: number = 1;
  private static prop2: number;

  test() {
    const that = {} as this & typeof TestBothIntersection;
    that.prop1 = 1;
    that.prop2 = 1;
  }
}
		`},
		// Accessor properties
		{Code: `
class TestStaticPrivateAccessor {
  private static accessor staticAcc = 1;
}
		`},
		{Code: `
class TestStaticPrivateFieldAccessor {
  static accessor #staticAcc = 1;
}
		`},
		{Code: `
class TestPrivateAccessor {
  private accessor acc = 3;
}
		`},
		{Code: `
class TestPrivateFieldAccessor {
  accessor #acc = 3;
}
		`},
		// Function returning class with method modification - private keyword
		{Code: `
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    private _name: string;

    public test(value: string) {
      this._name = value;
    }
  };
}
		`},
		// Function returning class with method modification - private field
		{Code: `
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    #name: string;

    public test(value: string) {
      this.#name = value;
    }
  };
}
		`},
	}, []rule_tester.InvalidTestCase{
		// Basic private static
		{
			Code: `
class TestIncorrectlyModifiableStatic {
  private static incorrectlyModifiableStatic = 7;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiableStatic {
  private static readonly incorrectlyModifiableStatic = 7;
}
			`},
		},
		// Private field static
		{
			Code: `
class TestIncorrectlyModifiableStatic {
  static #incorrectlyModifiableStatic = 7;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiableStatic {
  static readonly #incorrectlyModifiableStatic = 7;
}
			`},
		},
		// Static arrow - private keyword
		{
			Code: `
class TestIncorrectlyModifiableStaticArrow {
  private static incorrectlyModifiableStaticArrow = () => 7;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiableStaticArrow {
  private static readonly incorrectlyModifiableStaticArrow = () => 7;
}
			`},
		},
		// Static arrow - private field
		{
			Code: `
class TestIncorrectlyModifiableStaticArrow {
  static #incorrectlyModifiableStaticArrow = () => 7;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiableStaticArrow {
  static readonly #incorrectlyModifiableStaticArrow = () => 7;
}
			`},
		},
		// Nested class expressions - both reported (private keyword)
		{
			Code: `
class TestIncorrectlyModifiableInline {
  private incorrectlyModifiableInline = 7;

  public createConfusingChildClass() {
    return class {
      private incorrectlyModifiableInline = 7;
    };
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 7},
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiableInline {
  private readonly incorrectlyModifiableInline = 7;

  public createConfusingChildClass() {
    return class {
      private readonly incorrectlyModifiableInline = 7;
    };
  }
}
			`},
		},
		// Nested class expressions - both reported (private field)
		{
			Code: `
class TestIncorrectlyModifiableInline {
  #incorrectlyModifiableInline = 7;

  public createConfusingChildClass() {
    return class {
      #incorrectlyModifiableInline = 7;
    };
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 7},
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiableInline {
  readonly #incorrectlyModifiableInline = 7;

  public createConfusingChildClass() {
    return class {
      readonly #incorrectlyModifiableInline = 7;
    };
  }
}
			`},
		},
		// Constructor assignment with literal type widening - private keyword
		{
			Code: `
class TestIncorrectlyModifiableDelayed {
  private incorrectlyModifiableDelayed = 7;

  public constructor() {
    this.incorrectlyModifiableDelayed = 7;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiableDelayed {
  private readonly incorrectlyModifiableDelayed: number = 7;

  public constructor() {
    this.incorrectlyModifiableDelayed = 7;
  }
}
			`},
		},
		// Constructor assignment - private field
		{
			Code: `
class TestIncorrectlyModifiableDelayed {
  #incorrectlyModifiableDelayed = 7;

  public constructor() {
    this.#incorrectlyModifiableDelayed = 7;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiableDelayed {
  readonly #incorrectlyModifiableDelayed = 7;

  public constructor() {
    this.#incorrectlyModifiableDelayed = 7;
  }
}
			`},
		},
		// Child class expression modifiable - only outer reported (private keyword)
		{
			Code: `
class TestChildClassExpressionModifiable {
  private childClassExpressionModifiable = 7;

  public createConfusingChildClass() {
    return class {
      private childClassExpressionModifiable = 7;

      mutate() {
        this.childClassExpressionModifiable += 1;
      }
    };
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestChildClassExpressionModifiable {
  private readonly childClassExpressionModifiable = 7;

  public createConfusingChildClass() {
    return class {
      private childClassExpressionModifiable = 7;

      mutate() {
        this.childClassExpressionModifiable += 1;
      }
    };
  }
}
			`},
		},
		// Child class expression modifiable - only outer reported (private field)
		{
			Code: `
class TestChildClassExpressionModifiable {
  #childClassExpressionModifiable = 7;

  public createConfusingChildClass() {
    return class {
      #childClassExpressionModifiable = 7;

      mutate() {
        this.#childClassExpressionModifiable += 1;
      }
    };
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestChildClassExpressionModifiable {
  readonly #childClassExpressionModifiable = 7;

  public createConfusingChildClass() {
    return class {
      #childClassExpressionModifiable = 7;

      mutate() {
        this.#childClassExpressionModifiable += 1;
      }
    };
  }
}
			`},
		},
		// Binary expr minus (not assignment) - private keyword
		{
			Code: `
class TestIncorrectlyModifiablePostMinus {
  private incorrectlyModifiablePostMinus = 7;

  public mutate() {
    this.incorrectlyModifiablePostMinus - 1;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiablePostMinus {
  private readonly incorrectlyModifiablePostMinus = 7;

  public mutate() {
    this.incorrectlyModifiablePostMinus - 1;
  }
}
			`},
		},
		// Binary expr plus (not assignment) - private keyword
		{
			Code: `
class TestIncorrectlyModifiablePostPlus {
  private incorrectlyModifiablePostPlus = 7;

  public mutate() {
    this.incorrectlyModifiablePostPlus + 1;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiablePostPlus {
  private readonly incorrectlyModifiablePostPlus = 7;

  public mutate() {
    this.incorrectlyModifiablePostPlus + 1;
  }
}
			`},
		},
		// Unary minus (not --) - private keyword
		{
			Code: `
class TestIncorrectlyModifiablePreMinus {
  private incorrectlyModifiablePreMinus = 7;

  public mutate() {
    -this.incorrectlyModifiablePreMinus;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiablePreMinus {
  private readonly incorrectlyModifiablePreMinus = 7;

  public mutate() {
    -this.incorrectlyModifiablePreMinus;
  }
}
			`},
		},
		// Unary plus (not ++) - private keyword
		{
			Code: `
class TestIncorrectlyModifiablePrePlus {
  private incorrectlyModifiablePrePlus = 7;

  public mutate() {
    +this.incorrectlyModifiablePrePlus;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiablePrePlus {
  private readonly incorrectlyModifiablePrePlus = 7;

  public mutate() {
    +this.incorrectlyModifiablePrePlus;
  }
}
			`},
		},
		// Overlapping class variable
		{
			Code: `
class TestOverlappingClassVariable {
  private overlappingClassVariable = 7;

  public workWithSimilarClass(other: SimilarClass) {
    other.overlappingClassVariable = 7;
  }
}

class SimilarClass {
  public overlappingClassVariable = 7;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestOverlappingClassVariable {
  private readonly overlappingClassVariable = 7;

  public workWithSimilarClass(other: SimilarClass) {
    other.overlappingClassVariable = 7;
  }
}

class SimilarClass {
  public overlappingClassVariable = 7;
}
			`},
		},
		// Constructor parameter property
		{
			Code: `
class TestIncorrectlyModifiableParameter {
  public constructor(private incorrectlyModifiableParameter = 7) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiableParameter {
  public constructor(private readonly incorrectlyModifiableParameter = 7) {}
}
			`},
		},
		// onlyInlineLambdas: inline lambda should be reported
		{
			Code: `
class TestCorrectlyNonInlineLambdas {
  private incorrectlyInlineLambda = () => 7;
}
			`,
			Options: []interface{}{map[string]interface{}{"onlyInlineLambdas": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestCorrectlyNonInlineLambdas {
  private readonly incorrectlyInlineLambda = () => 7;
}
			`},
		},
		// Object sub-property write - private keyword
		{
			Code: `
class Test {
  private testObj = {
    prop: '',
  };

  public test(): void {
    this.testObj.prop = '';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly testObj = {
    prop: '',
  };

  public test(): void {
    this.testObj.prop = '';
  }
}
			`},
		},
		// TestObject sub-property write - private keyword
		{
			Code: `
class TestObject {
  public prop: number;
}

class Test {
  private testObj = new TestObject();

  public test(): void {
    this.testObj.prop = 10;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 7},
			},
			Output: []string{`
class TestObject {
  public prop: number;
}

class Test {
  private readonly testObj = new TestObject();

  public test(): void {
    this.testObj.prop = 10;
  }
}
			`},
		},
		// Optional chaining - still report
		{
			Code: `
class Test {
  private testObj = {};
  public test(): void {
    this.testObj?.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly testObj = {};
  public test(): void {
    this.testObj?.prop;
  }
}
			`},
		},
		// Non-null assertion - still report
		{
			Code: `
class Test {
  private testObj = {};
  public test(): void {
    this.testObj!.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly testObj = {};
  public test(): void {
    this.testObj!.prop;
  }
}
			`},
		},
		// Nested property access (not direct assignment)
		{
			Code: `
class Test {
  private testObj = {};
  public test(): void {
    this.testObj.prop.prop = '';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly testObj = {};
  public test(): void {
    this.testObj.prop.prop = '';
  }
}
			`},
		},
		// Intersection: unrelated prop write, class prop should be readonly
		{
			Code: `
class Test {
  private prop: number = 3;

  test() {
    const that = {} as this & { _foo: 'bar' };
    that._foo = 1;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop: number = 3;

  test() {
    const that = {} as this & { _foo: 'bar' };
    that._foo = 1;
  }
}
			`},
		},
		// Union: prop only read, should be readonly
		{
			Code: `
class Test {
  private prop: number = 3;

  test() {
    const that = {} as this | (this & { _foo: 'bar' });
    that.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop: number = 3;

  test() {
    const that = {} as this | (this & { _foo: 'bar' });
    that.prop;
  }
}
			`},
		},
		// String literal with constructor modification - type annotation added
		{
			Code: `
class Test {
  private prop = 'hello';

  constructor() {
    this.prop = 'world';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop: string = 'hello';

  constructor() {
    this.prop = 'world';
  }
}
			`},
		},
		// Number literal with constructor modification
		{
			Code: `
class Test {
  private prop = 10;

  constructor() {
    this.prop = 11;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop: number = 10;

  constructor() {
    this.prop = 11;
  }
}
			`},
		},
		// Boolean literal with constructor modification
		{
			Code: `
class Test {
  private prop = true;

  constructor() {
    this.prop = false;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop: boolean = true;

  constructor() {
    this.prop = false;
  }
}
			`},
		},
		// With existing type annotation - no additional annotation
		{
			Code: `
class Test {
  private prop: string = 'hello';
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop: string = 'hello';
}
			`},
		},
		// Constructor assignment without literal modification - no type widening
		{
			Code: `
class Test {
  private prop: string;

  constructor() {
    this.prop = 'hello';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop: string;

  constructor() {
    this.prop = 'hello';
  }
}
			`},
		},
		// Binary expr minus - private field
		{
			Code: `
class TestIncorrectlyModifiablePostMinus {
  #incorrectlyModifiablePostMinus = 7;

  public mutate() {
    this.#incorrectlyModifiablePostMinus - 1;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiablePostMinus {
  readonly #incorrectlyModifiablePostMinus = 7;

  public mutate() {
    this.#incorrectlyModifiablePostMinus - 1;
  }
}
			`},
		},
		// Binary expr plus - private field
		{
			Code: `
class TestIncorrectlyModifiablePostPlus {
  #incorrectlyModifiablePostPlus = 7;

  public mutate() {
    this.#incorrectlyModifiablePostPlus + 1;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiablePostPlus {
  readonly #incorrectlyModifiablePostPlus = 7;

  public mutate() {
    this.#incorrectlyModifiablePostPlus + 1;
  }
}
			`},
		},
		// Unary minus - private field
		{
			Code: `
class TestIncorrectlyModifiablePreMinus {
  #incorrectlyModifiablePreMinus = 7;

  public mutate() {
    -this.#incorrectlyModifiablePreMinus;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiablePreMinus {
  readonly #incorrectlyModifiablePreMinus = 7;

  public mutate() {
    -this.#incorrectlyModifiablePreMinus;
  }
}
			`},
		},
		// Unary plus - private field
		{
			Code: `
class TestIncorrectlyModifiablePrePlus {
  #incorrectlyModifiablePrePlus = 7;

  public mutate() {
    +this.#incorrectlyModifiablePrePlus;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class TestIncorrectlyModifiablePrePlus {
  readonly #incorrectlyModifiablePrePlus = 7;

  public mutate() {
    +this.#incorrectlyModifiablePrePlus;
  }
}
			`},
		},
		// Multi-parameter constructor property
		{
			Code: `
class TestIncorrectlyModifiableParameter {
  public constructor(
    public ignore: boolean,
    private incorrectlyModifiableParameter = 7,
  ) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 5},
			},
			Output: []string{`
class TestIncorrectlyModifiableParameter {
  public constructor(
    public ignore: boolean,
    private readonly incorrectlyModifiableParameter = 7,
  ) {}
}
			`},
		},
		// ClassWithName - unused private (private keyword)
		{
			Code: `
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    private _name: string;
  };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 4},
			},
			Output: []string{`
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    private readonly _name: string;
  };
}
			`},
		},
		// ClassWithName - unused private (private field)
		{
			Code: `
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    #name: string;
  };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 4},
			},
			Output: []string{`
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    readonly #name: string;
  };
}
			`},
		},
		// testObj sub-property write - private field
		{
			Code: `
class Test {
  #testObj = {
    prop: '',
  };

  public test(): void {
    this.#testObj.prop = '';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  readonly #testObj = {
    prop: '',
  };

  public test(): void {
    this.#testObj.prop = '';
  }
}
			`},
		},
		// TestObject sub-property write - private field
		{
			Code: `
class TestObject {
  public prop: number;
}

class Test {
  #testObj = new TestObject();

  public test(): void {
    this.#testObj.prop = 10;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 7},
			},
			Output: []string{`
class TestObject {
  public prop: number;
}

class Test {
  readonly #testObj = new TestObject();

  public test(): void {
    this.#testObj.prop = 10;
  }
}
			`},
		},
		// Object property read - private keyword
		{
			Code: `
class Test {
  private testObj = {
    prop: '',
  };
  public test(): void {
    this.testObj.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly testObj = {
    prop: '',
  };
  public test(): void {
    this.testObj.prop;
  }
}
			`},
		},
		// Object property read - private field
		{
			Code: `
class Test {
  #testObj = {
    prop: '',
  };
  public test(): void {
    this.#testObj.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  readonly #testObj = {
    prop: '',
  };
  public test(): void {
    this.#testObj.prop;
  }
}
			`},
		},
		// Optional chaining - private field
		{
			Code: `
class Test {
  #testObj = {};
  public test(): void {
    this.#testObj?.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  readonly #testObj = {};
  public test(): void {
    this.#testObj?.prop;
  }
}
			`},
		},
		// Non-null assertion - private field
		{
			Code: `
class Test {
  #testObj = {};
  public test(): void {
    this.#testObj!.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  readonly #testObj = {};
  public test(): void {
    this.#testObj!.prop;
  }
}
			`},
		},
		// Nested prop.prop write - private field
		{
			Code: `
class Test {
  #testObj = {};
  public test(): void {
    this.#testObj.prop.prop = '';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  readonly #testObj = {};
  public test(): void {
    this.#testObj.prop.prop = '';
  }
}
			`},
		},
		// Method call on nested prop - private keyword
		{
			Code: `
class Test {
  private testObj = {};
  public test(): void {
    this.testObj.prop.doesSomething();
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly testObj = {};
  public test(): void {
    this.testObj.prop.doesSomething();
  }
}
			`},
		},
		// Method call on nested prop - private field
		{
			Code: `
class Test {
  #testObj = {};
  public test(): void {
    this.#testObj.prop.doesSomething();
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  readonly #testObj = {};
  public test(): void {
    this.#testObj.prop.doesSomething();
  }
}
			`},
		},
		// Optional chaining ?.prop.prop - private keyword
		{
			Code: `
class Test {
  private testObj = {};
  public test(): void {
    this.testObj?.prop.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly testObj = {};
  public test(): void {
    this.testObj?.prop.prop;
  }
}
			`},
		},
		// Optional chaining ?.prop.prop - private field
		{
			Code: `
class Test {
  #testObj = {};
  public test(): void {
    this.#testObj?.prop.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  readonly #testObj = {};
  public test(): void {
    this.#testObj?.prop.prop;
  }
}
			`},
		},
		// Optional chaining ?.prop?.prop - private keyword
		{
			Code: `
class Test {
  private testObj = {};
  public test(): void {
    this.testObj?.prop?.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly testObj = {};
  public test(): void {
    this.testObj?.prop?.prop;
  }
}
			`},
		},
		// Optional chaining ?.prop?.prop - private field
		{
			Code: `
class Test {
  #testObj = {};
  public test(): void {
    this.#testObj?.prop?.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  readonly #testObj = {};
  public test(): void {
    this.#testObj?.prop?.prop;
  }
}
			`},
		},
		// Optional chaining .prop?.prop - private keyword
		{
			Code: `
class Test {
  private testObj = {};
  public test(): void {
    this.testObj.prop?.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly testObj = {};
  public test(): void {
    this.testObj.prop?.prop;
  }
}
			`},
		},
		// Optional chaining .prop?.prop - private field
		{
			Code: `
class Test {
  #testObj = {};
  public test(): void {
    this.#testObj.prop?.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  readonly #testObj = {};
  public test(): void {
    this.#testObj.prop?.prop;
  }
}
			`},
		},
		// Non-null + optional !.prop?.prop - private keyword
		{
			Code: `
class Test {
  private testObj = {};
  public test(): void {
    this.testObj!.prop?.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly testObj = {};
  public test(): void {
    this.testObj!.prop?.prop;
  }
}
			`},
		},
		// Non-null + optional !.prop?.prop - private field
		{
			Code: `
class Test {
  #testObj = {};
  public test(): void {
    this.#testObj!.prop?.prop;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  readonly #testObj = {};
  public test(): void {
    this.#testObj!.prop?.prop;
  }
}
			`},
		},
		// Intersection type in constructor
		{
			Code: `
class Test {
  private prop: number;

  constructor() {
    const that = {} as this & { _foo: 'bar' };
    that.prop = 1;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop: number;

  constructor() {
    const that = {} as this & { _foo: 'bar' };
    that.prop = 1;
  }
}
			`},
		},
		// String literal no constructor
		{
			Code: `
class Test {
  private prop = 'hello';
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop = 'hello';
}
			`},
		},
		// Declared const 'hello' with constructor
		{
			Code: `
declare const hello: 'hello';

class Test {
  private prop = hello;

  constructor() {
    this.prop = 'world';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 5},
			},
			Output: []string{`
declare const hello: 'hello';

class Test {
  private readonly prop = hello;

  constructor() {
    this.prop = 'world';
  }
}
			`},
		},
		// Declared const 'hello' without constructor
		{
			Code: `
declare const hello: 'hello';

class Test {
  private prop = hello;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 5},
			},
			Output: []string{`
declare const hello: 'hello';

class Test {
  private readonly prop = hello;
}
			`},
		},
		// Number literal without constructor
		{
			Code: `
class Test {
  private prop = 10;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop = 10;
}
			`},
		},
		// Declared const 10 with constructor
		{
			Code: `
declare const hello: 10;

class Test {
  private prop = hello;

  constructor() {
    this.prop = 11;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 5},
			},
			Output: []string{`
declare const hello: 10;

class Test {
  private readonly prop = hello;

  constructor() {
    this.prop = 11;
  }
}
			`},
		},
		// Boolean literal without constructor
		{
			Code: `
class Test {
  private prop = true;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop = true;
}
			`},
		},
		// Declared const true with constructor
		{
			Code: `
declare const hello: true;

class Test {
  private prop = hello;

  constructor() {
    this.prop = false;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 5},
			},
			Output: []string{`
declare const hello: true;

class Test {
  private readonly prop = hello;

  constructor() {
    this.prop = false;
  }
}
			`},
		},
		// Enum with constructor
		{
			Code: `
enum Foo {
  Bar,
  Bazz,
}

class Test {
  private prop = Foo.Bar;

  constructor() {
    this.prop = Foo.Bazz;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 8},
			},
			Output: []string{`
enum Foo {
  Bar,
  Bazz,
}

class Test {
  private readonly prop = Foo.Bar;

  constructor() {
    this.prop = Foo.Bazz;
  }
}
			`},
		},
		// Enum without constructor
		{
			Code: `
enum Foo {
  Bar,
  Bazz,
}

class Test {
  private prop = Foo.Bar;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 8},
			},
			Output: []string{`
enum Foo {
  Bar,
  Bazz,
}

class Test {
  private readonly prop = Foo.Bar;
}
			`},
		},
		// Enum const foo with constructor
		{
			Code: `
enum Foo {
  Bar,
  Bazz,
}

const foo = Foo.Bar;

class Test {
  private prop = foo;

  constructor() {
    this.prop = foo;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 10},
			},
			Output: []string{`
enum Foo {
  Bar,
  Bazz,
}

const foo = Foo.Bar;

class Test {
  private readonly prop = foo;

  constructor() {
    this.prop = foo;
  }
}
			`},
		},
		// Enum const foo without constructor
		{
			Code: `
enum Foo {
  Bar,
  Bazz,
}

const foo = Foo.Bar;

class Test {
  private prop = foo;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 10},
			},
			Output: []string{`
enum Foo {
  Bar,
  Bazz,
}

const foo = Foo.Bar;

class Test {
  private readonly prop = foo;
}
			`},
		},
		// Declare const foo: Foo
		{
			Code: `
enum Foo {
  Bar,
  Bazz,
}

declare const foo: Foo;

class Test {
  private prop = foo;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 10},
			},
			Output: []string{`
enum Foo {
  Bar,
  Bazz,
}

declare const foo: Foo;

class Test {
  private readonly prop = foo;
}
			`},
		},
		// Enum with const shadowing by value
		{
			Code: `
enum Foo {
  Bar,
  Bazz,
}

const bar = Foo.Bar;

function wrapper() {
  const Foo = 10;

  class Test {
    private prop = bar;

    constructor() {
      this.prop = bar;
    }
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 13},
			},
			Output: []string{`
enum Foo {
  Bar,
  Bazz,
}

const bar = Foo.Bar;

function wrapper() {
  const Foo = 10;

  class Test {
    private readonly prop = bar;

    constructor() {
      this.prop = bar;
    }
  }
}
			`},
		},
		// Enum with type shadowing
		{
			Code: `
enum Foo {
  Bar,
  Bazz,
}

const bar = Foo.Bar;

function wrapper() {
  type Foo = 10;

  class Test {
    private prop = bar;

    constructor() {
      this.prop = bar;
    }
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 13},
			},
			Output: []string{`
enum Foo {
  Bar,
  Bazz,
}

const bar = Foo.Bar;

function wrapper() {
  type Foo = 10;

  class Test {
    private readonly prop = bar;

    constructor() {
      this.prop = bar;
    }
  }
}
			`},
		},
		// IIFE enum const
		{
			Code: `
const Bar = (function () {
  enum Foo {
    Bar,
    Bazz,
  }

  return Foo;
})();

const bar = Bar.Bar;

class Test {
  private prop = bar;

  constructor() {
    this.prop = bar;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 14},
			},
			Output: []string{`
const Bar = (function () {
  enum Foo {
    Bar,
    Bazz,
  }

  return Foo;
})();

const bar = Bar.Bar;

class Test {
  private readonly prop = bar;

  constructor() {
    this.prop = bar;
  }
}
			`},
		},
		// Object literal without constructor
		{
			Code: `
class Test {
  private prop = { foo: 'bar' };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop = { foo: 'bar' };
}
			`},
		},
		// Object literal with constructor
		{
			Code: `
class Test {
  private prop = { foo: 'bar' };

  constructor() {
    this.prop = { foo: 'bazz' };
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop = { foo: 'bar' };

  constructor() {
    this.prop = { foo: 'bazz' };
  }
}
			`},
		},
		// Array literal without constructor
		{
			Code: `
class Test {
  private prop = [1, 2, 'three'];
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop = [1, 2, 'three'];
}
			`},
		},
		// Array literal with constructor
		{
			Code: `
class Test {
  private prop = [1, 2, 'three'];

  constructor() {
    this.prop = [1, 2, 'four'];
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop = [1, 2, 'three'];

  constructor() {
    this.prop = [1, 2, 'four'];
  }
}
			`},
		},
		// Boolean conditional constructor assignment with type widening
		{
			Code: `
class X {
  private _isValid = true;

  getIsValid = () => this._isValid;

  constructor(data?: {}) {
    if (!data) {
      this._isValid = false;
    }
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class X {
  private readonly _isValid: boolean = true;

  getIsValid = () => this._isValid;

  constructor(data?: {}) {
    if (!data) {
      this._isValid = false;
    }
  }
}
			`},
		},
		// String | number type annotation
		{
			Code: `
class Test {
  private prop: string | number = 'hello';
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop: string | number = 'hello';
}
			`},
		},
		// Prop no type annotation, no initializer, constructor only
		{
			Code: `
class Test {
  private prop;

  constructor() {
    this.prop = 'hello';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop;

  constructor() {
    this.prop = 'hello';
  }
}
			`},
		},
		// Conditional constructor if/else
		{
			Code: `
class Test {
  private prop;

  constructor(x: boolean) {
    if (x) {
      this.prop = 'hello';
    } else {
      this.prop = 10;
    }
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop;

  constructor(x: boolean) {
    if (x) {
      this.prop = 'hello';
    } else {
      this.prop = 10;
    }
  }
}
			`},
		},
		// Union type declared const
		{
			Code: `
declare const hello: 'hello' | 10;

class Test {
  private prop = hello;

  constructor() {
    this.prop = 10;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 5},
			},
			Output: []string{`
declare const hello: 'hello' | 10;

class Test {
  private readonly prop = hello;

  constructor() {
    this.prop = 10;
  }
}
			`},
		},
		// Null without constructor
		{
			Code: `
class Test {
  private prop = null;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop = null;
}
			`},
		},
		// Null with constructor
		{
			Code: `
class Test {
  private prop = null;

  constructor() {
    this.prop = null;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop = null;

  constructor() {
    this.prop = null;
  }
}
			`},
		},
		// As expression
		{
			Code: `
class Test {
  private prop = 'hello' as string;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop = 'hello' as string;
}
			`},
		},
		// Promise.resolve
		{
			Code: `
class Test {
  private prop = Promise.resolve('hello');
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferReadonly", Line: 3},
			},
			Output: []string{`
class Test {
  private readonly prop = Promise.resolve('hello');
}
			`},
		},
	})
}
