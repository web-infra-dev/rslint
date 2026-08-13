package no_unsafe_assignment

import (
	"slices"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

type assignmentTestCase struct {
	code               string
	col                int
	endCol             int
	skipAssignmentExpr bool
	messageID          string
}

func assignmentTest(tests []assignmentTestCase) []rule_tester.InvalidTestCase {
	res := make([]rule_tester.InvalidTestCase, 0, 3*len(tests))
	for _, test := range tests {
		res = append(res,
			// VariableDeclaration
			rule_tester.InvalidTestCase{
				Code: "const " + test.code,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Column:    test.col + 6,
						EndColumn: test.endCol + 6,
						Line:      1,
						MessageId: test.messageID,
					},
				},
			},
			// AssignmentPattern
			rule_tester.InvalidTestCase{
				Code: "function foo(" + test.code + ") {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Column:    test.col + 13,
						EndColumn: test.endCol + 13,
						Line:      1,
						MessageId: test.messageID,
					},
				},
			},
		)
		if !test.skipAssignmentExpr {
			// AssignmentExpression
			res = append(res, rule_tester.InvalidTestCase{
				Code: "(" + test.code + ")",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Column:    test.col + 1,
						EndColumn: test.endCol + 1,
						Line:      1,
						MessageId: test.messageID,
					},
				},
			})
		}
	}
	return res
}

func TestNoUnsafeAssignmentRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noImplicitThis.json", t, &NoUnsafeAssignmentRule, []rule_tester.ValidTestCase{
		{Code: "const x = 1;"},
		{Code: "const x: number = 1;"},
		{Code: `
const x = 1,
  y = 1;
    `},
		{Code: "let x;"},
		{Code: `
let x = 1,
  y;
    `},
		{Code: "function foo(a = 1) {}"},
		{Code: `
class Foo {
  constructor(private a = 1) {}
}
    `},
		{Code: `
class Foo {
  private a = 1;
}
    `},
		{Code: `
class Foo {
  accessor a = 1;
}
    `},
		{Code: "const x: Set<string> = new Set();"},
		{Code: "const x: Set<string> = new Set<string>();"},
		{Code: "const [x] = [1];"},
		{Code: "const [x, y] = [1, 2] as number[];"},
		{Code: "const [x, ...y] = [1, 2, 3, 4, 5];"},
		{Code: "const [x, ...y] = [1];"},
		{Code: "const [{ ...x }] = [{ x: 1 }] as [{ x: any }];"},
		{Code: "function foo(x = 1) {}"},
		{Code: "function foo([x] = [1]) {}"},
		{Code: "function foo([x, ...y] = [1, 2, 3, 4, 5]) {}"},
		{Code: "function foo([x, ...y] = [1]) {}"},
		{Code: "const x = new Set<any>();"},
		{Code: "const x = { y: 1 };"},
		// TODO(port): this is invalid TypeScript code
		{Skip: true, Code: "const x = { y = 1 };"},
		{Code: "const x = { y(){} };"},
		{Code: "const x: { y: number } = { y: 1 };"},
		{Code: "const x = [...[1, 2, 3]];"},
		{Code: "const [{ [`x${1}`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }];"},
		{Code: `
type T = [string, T[]];
const test: T = ['string', []] as T;
    `},
		{
			Code: `
type Props = { a: string };
declare function Foo(props: Props): never;
<Foo a={'foo'} />;
      `,
			Tsx: true,
		},
		{
			Code: `
declare function Foo(props: { a: string }): never;
<Foo a="foo" />;
      `,
			Tsx: true,
		},
		{
			Code: `
declare function Foo(props: { a: string }): never;
<Foo a={} />;
      `,
			Tsx: true,
		},
		{Code: "const x: unknown = y as any;"},
		{Code: "const x: unknown[] = y as any[];"},
		{Code: "const x: Set<unknown> = y as Set<any>;"},
		{Code: "const x: Map<string, string> = new Map();"},
		{Code: `
type Foo = { bar: unknown };
const bar: any = 1;
const foo: Foo = { bar };
    `},

		{Code: `
declare const foo: any;
let a = 1;

a+= foo;
		`},
		{
			Code: `
declare const dynamic: any;
const booleanResult = dynamic && true;
const options: { refresh: boolean } = { refresh: booleanResult };
      `,
			TSConfig: "tsconfig.unstrict.json",
		},
		{
			Code: `
const isEnabled = (value: MissingType | undefined, flag: boolean) => {
  if (typeof value === 'boolean' || value === undefined) {
    return (value ?? true) && flag;
  }
  return value.enabled !== false && flag;
};
let options: boolean | { enabled?: boolean } | undefined;
options = { enabled: isEnabled(spooky, true) };
      `,
			TSConfig: "tsconfig.unstrict.json",
		},
		{
			Code: `
declare function parseResource(value: string): { path: string; query: string; fragment: string };
declare const input: any;
const parsed = input && parseResource(input);
let path: string;
let query: string;
let fragment: string;
path = parsed ? parsed.path : undefined;
query = parsed ? parsed.query : undefined;
fragment = parsed ? parsed.fragment : undefined;
      `,
			TSConfig: "tsconfig.unstrict.json",
		},
		{
			Code: `
declare const dynamic: any;
const values = [...(dynamic && [1])];
      `,
			TSConfig: "tsconfig.unstrict.json",
		},
		{
			Code: `
declare const dynamic: any;
const genericLogical = <T>(value: T) => dynamic && value;
const genericString: string = genericLogical('safe');
const explicitString: string = genericLogical<string>(dynamic);
const genericUnrelated = <T, U>(value: T, unused: U) => dynamic && value;
const unrelatedString: string = genericUnrelated('safe', dynamic);
const genericArray = <T>(value: T[]) => dynamic && value;
const stringArray: string[] = genericArray(['safe']);
const annotatedParameter = (value: boolean) => dynamic && value;
const annotatedBoolean: boolean = annotatedParameter(dynamic);
      `,
			TSConfig: "tsconfig.unstrict.json",
		},
		{
			Code: `
declare const dynamic: any;
function recursive(flag: boolean) {
  if (flag) {
    return dynamic && true;
  }
  return recursive(true);
}
const recursiveBoolean: boolean = recursive(false);
      `,
			TSConfig: "tsconfig.unstrict.json",
		},
	}, slices.Concat([]rule_tester.InvalidTestCase{
		{
			Code: "const x = 1 as any;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
					Message:   "Unsafe assignment of an `any` value.",
				},
			},
		},
		{
			Code: `
const x = 1 as any,
  y = 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: "function foo(a = 1 as any) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
class Foo {
  constructor(private a = 1 as any) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
class Foo {
  private a = 1 as any;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
class Foo {
  accessor a = 1 as any;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
const [x] = spooky;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
const [[[x]]] = [spooky];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArrayPatternFromTuple",
				},
			},
		},
		{
			Code: `
const {
  x: { y: z },
} = { x: spooky };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeObjectPattern",
					Message:   "Unsafe object destructuring of a property with an error typed value.",
				},
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
let value: number;

value = spooky;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
const [x] = 1 as any;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
const [x] = [] as any[];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArrayPattern",
				},
			},
		},
		{
			Code: "const x: Set<string> = new Set<any>();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeAssignment",
					Message:   "Unsafe assignment of type `Set<any>` to a variable of type `Set<string>`.",
				},
			},
		},
		{
			Code: "const x: Map<string, string> = new Map<string, any>();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeAssignment",
				},
			},
		},
		{
			Code: "const x: Set<string[]> = new Set<any[]>();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeAssignment",
				},
			},
		},
		{
			Code: `
declare const dynamic: any;
const booleanResult = dynamic && true;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment"},
			},
		},
		{
			Code: `
declare const dynamic: any;
const result = dynamic;
      `,
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment"},
			},
		},
		{
			Code: `
declare function getAny(): any;
let options: { enabled?: boolean };
options = { enabled: getAny() };
      `,
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment"},
			},
		},
		{
			Code: `
declare const dynamic: any;
const genericLogical = <T>(value: T) => dynamic && value;
const genericAny: boolean = genericLogical(dynamic);
      `,
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment"},
			},
		},
		{
			Code: `
declare const dynamic: any;
const genericArray = <T>(value: T[]) => dynamic && value;
const unsafeArray: string[] = genericArray(dynamic as any[]);
      `,
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeAssignment",
					Message:   "Unsafe assignment of type `any[]` to a variable of type `string[]`.",
				},
			},
		},
		{
			Code: `
declare const dynamic: any;
var duplicate = dynamic;
var duplicate = dynamic && true;
const duplicateTarget: boolean = duplicate;
      `,
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment"},
				{MessageId: "anyAssignment"},
			},
		},
		{
			Code: `
declare const dynamic: any;
let sink: unknown;
const assignmentProperty = {
  first: (sink = dynamic),
  second: dynamic,
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyAssignment"},
				{MessageId: "anyAssignment"},
			},
		},
		{
			Code: "const x: Set<Set<Set<string>>> = new Set<Set<Set<any>>>();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeAssignment",
				},
			},
		},
	},
		assignmentTest([]assignmentTestCase{
			{"[x] = [1] as [any]", 2, 3, false, "unsafeArrayPatternFromTuple"},
			{"[[[[x]]]] = [[[[1 as any]]]]", 5, 6, false, "unsafeArrayPatternFromTuple"},
			{"[[[[x]]]] = [1 as any]", 2, 9, true, "unsafeArrayPatternFromTuple"},
			{"[{x}] = [{x: 1}] as [{x: any}]", 3, 4, false, "unsafeObjectPattern"},
			{"[{['x']: x}] = [{['x']: 1}] as [{['x']: any}]", 10, 11, false, "unsafeObjectPattern"},
			{"[{[`x`]: x}] = [{[`x`]: 1}] as [{[`x`]: any}]", 10, 11, false, "unsafeObjectPattern"},
		}),

		[]rule_tester.InvalidTestCase{
			{
				Code: "[[[[x]]]] = [1 as any];",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeAssignment",
						Line:      1,
						Column:    1,
						EndColumn: 23,
					},
				},
			},
			{
				Code: `
const x = [...(1 as any)];
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeArraySpread",
					},
				},
			},
			{
				Code: `
const x = [...([] as any[])];
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeArraySpread",
					},
				},
			},
		},
		assignmentTest([]assignmentTestCase{
			{"{x} = {x: 1} as {x: any}", 2, 3, false, "unsafeObjectPattern"},
			{"{x: y} = {x: 1} as {x: any}", 5, 6, false, "unsafeObjectPattern"},
			{"{x: {y}} = {x: {y: 1}} as {x: {y: any}}", 6, 7, false, "unsafeObjectPattern"},
			{"{x: [y]} = {x: {y: 1}} as {x: [any]}", 6, 7, false, "unsafeArrayPatternFromTuple"},
		}),

		[]rule_tester.InvalidTestCase{
			{
				Code: "const x = { y: 1 as any };",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Column:    13,
						EndColumn: 24,
					},
				},
			},
			{
				Code: "const x = { y: { z: 1 as any } };",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Column:    18,
						EndColumn: 29,
					},
				},
			},
			{
				Code: "const x: { y: Set<Set<Set<string>>> } = { y: new Set<Set<Set<any>>>() };",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeAssignment",
						Column:    43,
						EndColumn: 70,
					},
				},
			},
			{
				Code: "const x = { ...(1 as any) };",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Column:    7,
						EndColumn: 28,
					},
				},
			},
			{
				Code: `
type Props = { a: string };
declare function Foo(props: Props): never;
<Foo a={1 as any} />;
      `,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Line:      4,
						Column:    9,
						EndColumn: 17,
					},
				},
			},
			{
				Code: `
function foo() {
  const bar = this;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignmentThis",
						Line:      3,
						Column:    9,
						EndColumn: 19,
					},
				},
			},
			{
				Code: `
type T = [string, T[]];
const test: T = ['string', []] as any;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Line:      3,
						Column:    7,
						EndColumn: 38,
					},
				},
			},
			{
				Code: `
type Foo = { bar: number };
const bar: any = 1;
const foo: Foo = { bar };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Line:      4,
						Column:    20,
						EndColumn: 23,
					},
				},
			},

			{
				Code: `
declare const foo: any;
interface Bar {
  bar: number
}

class Foo {
  constructor(
    private readonly param: Bar = Object.create(null)
  ) {}
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Line:      9,
						Column:    5,
						EndColumn: 54,
					},
				},
			},
			{
				Code: `
let foo: { foo: 1 };

foo = { bar: 2 } as any;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Line:      4,
						Column:    1,
						EndColumn: 24,
					},
				},
			},
		}))
}
