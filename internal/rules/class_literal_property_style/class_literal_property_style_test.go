package class_literal_property_style

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestClassLiteralPropertyStyleRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ClassLiteralPropertyStyleRule, []rule_tester.ValidTestCase{
		{Code: `
class Mx {
  declare readonly p1 = 1;
}
		`},
		{Code: `
class Mx {
  readonly p1 = 'hello world';
}
		`},
		{Code: `
class Mx {
  p1 = 'hello world';
}
		`},
		{Code: `
class Mx {
  static p1 = 'hello world';
}
		`},
		{Code: `
class Mx {
  p1: string;
}
		`},
		{Code: `
class Mx {
  get p1();
}
		`},
		{Code: `
class Mx {
  get p1() {}
}
		`},
		{Code: `
abstract class Mx {
  abstract get p1(): string;
}
		`},
		{Code: `
class Mx {
  get mySetting() {
    if (this._aValue) {
      return 'on';
    }

    return 'off';
  }
}
		`},
		{Code: `
class Mx {
  get mySetting() {
    return ` + "`build-${process.env.build}`" + `;
  }
}
		`},
		{Code: `
class Mx {
  getMySetting() {
    if (this._aValue) {
      return 'on';
    }

    return 'off';
  }
}
		`},
		{Code: `
class Mx {
  public readonly myButton = styled.button` + "`\n    color: ${props => (props.primary ? 'hotpink' : 'turquoise')};\n  `" + `;
}
		`},
		{Code: `
class Mx {
  set p1(val) {}
  get p1() {
    return '';
  }
}
		`},
		{Code: `
let p1 = 'p1';
class Mx {
  set [p1](val) {}
  get [p1]() {
    return '';
  }
}
		`},
		{Code: `
let p1 = 'p1';
class Mx {
  set [/* before set */ p1 /* after set */](val) {}
  get [/* before get */ p1 /* after get */]() {
    return '';
  }
}
		`},
		{Code: `
class Mx {
  set ['foo'](val) {}
  get foo() {
    return '';
  }
  set bar(val) {}
  get ['bar']() {
    return '';
  }
  set ['baz'](val) {}
  get baz() {
    return '';
  }
}
		`},
		{
			Code: `
class Mx {
  public get myButton() {
    return styled.button` + "`\n      color: ${props => (props.primary ? 'hotpink' : 'turquoise')};\n    `" + `;
  }
}
			`,
			Options: []interface{}{"fields"},
		},
		{
			Code: `
class Mx {
  declare public readonly foo = 1;
}
			`,
			Options: []interface{}{"getters"},
		},
		{
			Code: `
class Mx {
  get p1() {
    return 'hello world';
  }
}
			`,
			Options: []interface{}{"getters"},
		},
		{Code: `
class Mx {
  p1 = 'hello world';
}
		`, Options: []interface{}{"getters"}},
		{Code: `
class Mx {
  p1: string;
}
		`, Options: []interface{}{"getters"}},
		{Code: `
class Mx {
  readonly p1 = [1, 2, 3];
}
		`, Options: []interface{}{"getters"}},
		{Code: `
class Mx {
  static p1: string;
}
		`, Options: []interface{}{"getters"}},
		{Code: `
class Mx {
  static get p1() {
    return 'hello world';
  }
}
		`, Options: []interface{}{"getters"}},
		{Code: `
class Mx {
  public readonly myButton = styled.button` + "`\n    color: ${props => (props.primary ? 'hotpink' : 'turquoise')};\n  `" + `;
}
		`, Options: []interface{}{"getters"}},
		{Code: `
class Mx {
  public get myButton() {
    return styled.button` + "`\n      color: ${props => (props.primary ? 'hotpink' : 'turquoise')};\n    `" + `;
  }
}
		`, Options: []interface{}{"getters"}},
		{Code: `
class A {
  private readonly foo: string = 'bar';
  constructor(foo: string) {
    this.foo = foo;
  }
}
		`, Options: []interface{}{"getters"}},
		{Code: `
class A {
  private readonly foo: string = 'bar';
  constructor(foo: string) {
    this['foo'] = foo;
  }
}
		`, Options: []interface{}{"getters"}},
		{Code: `
class A {
  private readonly foo: string = 'bar';
  constructor(foo: string) {
    const bar = new (class {
      private readonly foo: string = 'baz';
      constructor() {
        this.foo = 'qux';
      }
    })();
    this['foo'] = foo;
  }
}
		`, Options: []interface{}{"getters"}},
		{
			Code: `
declare abstract class BaseClass {
  get cursor(): string;
}

class ChildClass extends BaseClass {
  override get cursor() {
    return 'overridden value';
  }
}
			`,
		},
		{
			Code: `
declare abstract class BaseClass {
  protected readonly foo: string;
}

class ChildClass extends BaseClass {
  protected override readonly foo = 'bar';
}
			`,
			Options: []interface{}{"getters"},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
class Mx {
  get p1() {
    return 'hello world';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFieldStyle",
					Line:      3,
					Column:    7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFieldStyleSuggestion",
							Output:    `
class Mx {
  readonly p1 = 'hello world';
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  get p1() {
    return ` + "`hello world`" + `;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFieldStyle",
					Line:      3,
					Column:    7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFieldStyleSuggestion",
							Output:    `
class Mx {
  readonly p1 = ` + "`hello world`" + `;
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  static get p1() {
    return 'hello world';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFieldStyle",
					Line:      3,
					Column:    14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFieldStyleSuggestion",
							Output:    `
class Mx {
  static readonly p1 = 'hello world';
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  public static get foo() {
    return 1;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFieldStyle",
					Line:      3,
					Column:    21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFieldStyleSuggestion",
							Output:    `
class Mx {
  public static readonly foo = 1;
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  public get [myValue]() {
    return 'a literal value';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFieldStyle",
					Line:      3,
					Column:    15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFieldStyleSuggestion",
							Output:    `
class Mx {
  public readonly [myValue] = 'a literal value';
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  public get [myValue]() {
    return 12345n;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFieldStyle",
					Line:      3,
					Column:    15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFieldStyleSuggestion",
							Output:    `
class Mx {
  public readonly [myValue] = 12345n;
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  public readonly [myValue] = 'a literal value';
}
			`,
			Options: []interface{}{"getters"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferGetterStyle",
					Line:      3,
					Column:    20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferGetterStyleSuggestion",
							Output:    `
class Mx {
  public get [myValue]() { return 'a literal value'; }
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  readonly p1 = 'hello world';
}
			`,
			Options: []interface{}{"getters"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferGetterStyle",
					Line:      3,
					Column:    12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferGetterStyleSuggestion",
							Output:    `
class Mx {
  get p1() { return 'hello world'; }
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  readonly p1 = ` + "`hello world`" + `;
}
			`,
			Options: []interface{}{"getters"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferGetterStyle",
					Line:      3,
					Column:    12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferGetterStyleSuggestion",
							Output:    `
class Mx {
  get p1() { return ` + "`hello world`" + `; }
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  static readonly p1 = 'hello world';
}
			`,
			Options: []interface{}{"getters"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferGetterStyle",
					Line:      3,
					Column:    19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferGetterStyleSuggestion",
							Output:    `
class Mx {
  static get p1() { return 'hello world'; }
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  protected get p1() {
    return 'hello world';
  }
}
			`,
			Options: []interface{}{"fields"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFieldStyle",
					Line:      3,
					Column:    17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFieldStyleSuggestion",
							Output:    `
class Mx {
  protected readonly p1 = 'hello world';
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  protected readonly p1 = 'hello world';
}
			`,
			Options: []interface{}{"getters"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferGetterStyle",
					Line:      3,
					Column:    22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferGetterStyleSuggestion",
							Output:    `
class Mx {
  protected get p1() { return 'hello world'; }
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  public static get p1() {
    return 'hello world';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFieldStyle",
					Line:      3,
					Column:    21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFieldStyleSuggestion",
							Output:    `
class Mx {
  public static readonly p1 = 'hello world';
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  public static readonly p1 = 'hello world';
}
			`,
			Options: []interface{}{"getters"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferGetterStyle",
					Line:      3,
					Column:    26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferGetterStyleSuggestion",
							Output:    `
class Mx {
  public static get p1() { return 'hello world'; }
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  public get myValue() {
    return gql` + "`\n      {\n        user(id: 5) {\n          firstName\n          lastName\n        }\n      }\n    `" + `;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFieldStyle",
					Line:      3,
					Column:    14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFieldStyleSuggestion",
							Output:    `
class Mx {
  public readonly myValue = gql` + "`\n      {\n        user(id: 5) {\n          firstName\n          lastName\n        }\n      }\n    `" + `;
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class Mx {
  public readonly myValue = gql` + "`\n    {\n      user(id: 5) {\n        firstName\n        lastName\n      }\n    }\n  `" + `;
}
			`,
			Options: []interface{}{"getters"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferGetterStyle",
					Line:      3,
					Column:    19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferGetterStyleSuggestion",
							Output:    `
class Mx {
  public get myValue() { return gql` + "`\n    {\n      user(id: 5) {\n        firstName\n        lastName\n      }\n    }\n  `" + `; }
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class A {
  private readonly foo: string = 'bar';
  constructor(foo: string) {
    const bar = new (class {
      private readonly foo: string = 'baz';
      constructor() {
        this.foo = 'qux';
      }
    })();
  }
}
			`,
			Options: []interface{}{"getters"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferGetterStyle",
					Line:      3,
					Column:    20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferGetterStyleSuggestion",
							Output:    `
class A {
  private get foo() { return 'bar'; }
  constructor(foo: string) {
    const bar = new (class {
      private readonly foo: string = 'baz';
      constructor() {
        this.foo = 'qux';
      }
    })();
  }
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class A {
  private readonly ['foo']: string = 'bar';
  constructor(foo: string) {
    const bar = new (class {
      private readonly foo: string = 'baz';
      constructor() {}
    })();

    if (bar) {
      this.foo = 'baz';
    }
  }
}
			`,
			Options: []interface{}{"getters"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferGetterStyle",
					Line:      6,
					Column:    24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferGetterStyleSuggestion",
							Output:    `
class A {
  private readonly ['foo']: string = 'bar';
  constructor(foo: string) {
    const bar = new (class {
      private get foo() { return 'baz'; }
      constructor() {}
    })();

    if (bar) {
      this.foo = 'baz';
    }
  }
}
			`,
						},
					},
				},
			},
		},
		{
			Code: `
class A {
  private readonly foo: string = 'bar';
  constructor(foo: string) {
    function func() {
      this.foo = 'aa';
    }
  }
}
			`,
			Options: []interface{}{"getters"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferGetterStyle",
					Line:      3,
					Column:    20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferGetterStyleSuggestion",
							Output:    `
class A {
  private get foo() { return 'bar'; }
  constructor(foo: string) {
    function func() {
      this.foo = 'aa';
    }
  }
}
			`,
						},
					},
				},
			},
		},
	})
}