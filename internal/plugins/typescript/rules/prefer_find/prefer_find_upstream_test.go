// TestPreferFindUpstream migrates the full valid/invalid suite from
// typescript-eslint's tests/rules/prefer-find.test.ts 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases live
// in prefer_find_extras_test.go.
package prefer_find

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferFindUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferFindRule, []rule_tester.ValidTestCase{
		// ---- type without array-shaped filter signature ----
		{Code: `
interface JerkCode<T> {
  filter(predicate: (item: T) => boolean): JerkCode<T>;
}

declare const jerkCode: JerkCode<string>;

jerkCode.filter(item => item === 'aha')[0];
`},
		// ---- non-zero subscript ----
		{Code: `
declare const arr: readonly string[];
arr.filter(item => item === 'aha')[1];
`},
		// ---- .at(non-zero) ----
		{Code: `
declare const arr: string[];
arr.filter(item => item === 'aha').at(1);
`},
		// ---- nullable receiver type still passes through isArrayish ----
		{Code: `
declare const notNecessarilyAnArray: unknown[] | undefined | null | string;
notNecessarilyAnArray?.filter(item => true)[0];
`},
		// ---- optional [0] (don't change semantics of `?.[0]`) ----
		{Code: `[].filter(() => true)?.[0];`},
		// ---- optional .at?.() (don't change semantics) ----
		{Code: `[].filter(() => true)?.at?.(0);`},
		// ---- optional .filter?.() — call itself is optional ----
		{Code: `[].filter?.(() => true)[0];`},
		// ---- .at(-Infinity) — not zero ----
		{Code: `[1, 2, 3].filter(x => x > 0).at(-Infinity);`},
		// ---- .at(non-zero literal) with optional-chain receiver ----
		{Code: `
declare const arr: string[];
declare const cond: Parameters<Array<string>['filter']>[0];
const a = { arr };
a?.arr.filter(cond).at(1);
`},
		// ---- plain .filter() without subscript ----
		{Code: `['Just', 'a', 'filter'].filter(x => x.length > 4);`},
		// ---- plain .find() ----
		{Code: `['Just', 'a', 'find'].find(x => x.length > 4);`},
		// ---- receiver is undefined (not arrayish) ----
		{Code: `undefined.filter(x => x)[0];`},
		// ---- receiver is null (optional chain short-circuits, not arrayish) ----
		{Code: `null?.filter(x => x)[0];`},
		// ---- Issue #8386 regression: Symbol.for must not throw ----
		{Code: `
declare function foo(param: any): any;
foo(Symbol.for('foo'));
`},
		// ---- .at(symbol-typed const) ----
		{Code: `
declare const arr: string[];
const s = Symbol.for("Don't throw!");
arr.filter(item => item === 'aha').at(s);
`},
		// ---- [Symbol('0')] — String(Symbol) !== '0' ----
		{Code: `[1, 2, 3].filter(x => x)[Symbol('0')];`},
		// ---- [Symbol.for('0')] — same ----
		{Code: `[1, 2, 3].filter(x => x)[Symbol.for('0')];`},
		// ---- ternary where one branch is not a filter call ----
		{Code: `(Math.random() < 0.5 ? [1, 2, 3].filter(x => true) : [1, 2, 3])[0];`},
		// ---- ternary where one branch is .find(), other is .filter() ----
		{Code: `
(Math.random() < 0.5
  ? [1, 2, 3].find(x => true)
  : [1, 2, 3].filter(x => true))[0];
`},
	}, []rule_tester.InvalidTestCase{
		// ---- basic [0] ----
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha')[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// ---- [zero] where zero = 0 ----
		{
			Code: `
declare const arr: Array<string>;
const zero = 0;
arr.filter(item => item === 'aha')[zero];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      4,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: Array<string>;
const zero = 0;
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// ---- [zero] where zero = 0n ----
		{
			Code: `
declare const arr: Array<string>;
const zero = 0n;
arr.filter(item => item === 'aha')[zero];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      4,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: Array<string>;
const zero = 0n;
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// ---- [zero] where zero = -0n ----
		{
			Code: `
declare const arr: Array<string>;
const zero = -0n;
arr.filter(item => item === 'aha')[zero];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      4,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: Array<string>;
const zero = -0n;
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// ---- .at(0) on readonly array ----
		{
			Code: `
declare const arr: readonly string[];
arr.filter(item => item === 'aha').at(0);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: readonly string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// ---- sequence in receiver: only last operand matters ----
		{
			Code: `
declare const arr: ReadonlyArray<string>;
(undefined, arr.filter(item => item === 'aha')).at(0);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: ReadonlyArray<string>;
(undefined, arr.find(item => item === 'aha'));
`,
						},
					},
				},
			},
		},
		// ---- .at(zero) where zero = 0 ----
		{
			Code: `
declare const arr: string[];
const zero = 0;
arr.filter(item => item === 'aha').at(zero);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      4,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
const zero = 0;
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// ---- ['0'] string-zero subscript ----
		{
			Code: `
declare const arr: string[];
arr.filter(item => item === 'aha')['0'];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
arr.find(item => item === 'aha');
`,
						},
					},
				},
			},
		},
		// ---- single-line inline assignment ----
		{
			Code: `const two = [1, 2, 3].filter(item => item === 2)[0];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      1,
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output:    `const two = [1, 2, 3].find(item => item === 2);`,
						},
					},
				},
			},
		},
		// ---- [fltr] where fltr = "filter" (computed access via const) ----
		{
			Code: `const fltr = "filter"; (([] as unknown[]))[fltr] ((item) => { return item === 2 }  ) [ 0  ] ;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      1,
					Column:    24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output:    `const fltr = "filter"; (([] as unknown[]))["find"] ((item) => { return item === 2 }  )  ;`,
						},
					},
				},
			},
		},
		// ---- ?.["filter"] computed access ----
		{
			Code: `(([] as unknown[]))?.["filter"] ((item) => { return item === 2 }  ) [ 0  ] ;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output:    `(([] as unknown[]))?.["find"] ((item) => { return item === 2 }  )  ;`,
						},
					},
				},
			},
		},
		// ---- nullable receiver with optional chain ----
		{
			Code: `
declare const nullableArray: unknown[] | undefined | null;
nullableArray?.filter(item => true)[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const nullableArray: unknown[] | undefined | null;
nullableArray?.find(item => true);
`,
						},
					},
				},
			},
		},
		// ---- ([]?.filter(f))[0] ----
		{
			Code: `([]?.filter(f))[0];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output:    `([]?.find(f));`,
						},
					},
				},
			},
		},
		// ---- nested sequence in console.log with optional-chain filter call ----
		{
			Code: `
declare const objectWithArrayProperty: { arr: unknown[] };
declare function cond(x: unknown): boolean;
console.log((1, 2, objectWithArrayProperty?.arr['filter'](cond)).at(0));
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      4,
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const objectWithArrayProperty: { arr: unknown[] };
declare function cond(x: unknown): boolean;
console.log((1, 2, objectWithArrayProperty?.arr["find"](cond)));
`,
						},
					},
				},
			},
		},
		// ---- .at(NaN) — Number(NaN) is NaN → treated as 0 ----
		{
			Code: `
[1, 2, 3].filter(x => x > 0).at(NaN);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      2,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
[1, 2, 3].find(x => x > 0);
`,
						},
					},
				},
			},
		},
		// ---- .at(negative-fractional const) — Math.trunc rounds toward zero ----
		{
			Code: `
const idxToLookUp = -0.12635678;
[1, 2, 3].filter(x => x > 0).at(idxToLookUp);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
const idxToLookUp = -0.12635678;
[1, 2, 3].find(x => x > 0);
`,
						},
					},
				},
			},
		},
		// ---- [`at`](0) — template-literal property name ----
		{
			Code: "\n[1, 2, 3].filter(x => x > 0)[`at`](0);\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      2,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output:    "\n[1, 2, 3].find(x => x > 0);\n",
						},
					},
				},
			},
		},
		// ---- comments around .filter and .at, .at('0') string subscript ----
		{
			Code: `
declare const arr: string[];
declare const cond: Parameters<Array<string>['filter']>[0];
const a = { arr };
a?.arr
  .filter(cond) /* what a bad spot for a comment. Let's make sure
  there's some yucky symbols too. [ . ?. <>   ' ' \'] */
  .at('0');
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      5,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: string[];
declare const cond: Parameters<Array<string>['filter']>[0];
const a = { arr };
a?.arr
  .find(cond) /* what a bad spot for a comment. Let's make sure
  there's some yucky symbols too. [ . ?. <>   ' ' \'] */
  ;
`,
						},
					},
				},
			},
		},
		// ---- spread inside push, comments in ['filter'] and [`0`]! non-null ----
		{
			Code: "\nconst imNotActuallyAnArray = [\n  [1, 2, 3],\n  [2, 3, 4],\n] as const;\nconst butIAm = [4, 5, 6];\nbutIAm.push(\n  // line comment!\n  ...imNotActuallyAnArray[/* comment */ 'filter' /* another comment */](\n    x => x[1] > 0,\n  ) /**/[`0`]!,\n);\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      9,
					Column:    6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output:    "\nconst imNotActuallyAnArray = [\n  [1, 2, 3],\n  [2, 3, 4],\n] as const;\nconst butIAm = [4, 5, 6];\nbutIAm.push(\n  // line comment!\n  ...imNotActuallyAnArray[/* comment */ \"find\" /* another comment */](\n    x => x[1] > 0,\n  ) /**/!,\n);\n",
						},
					},
				},
			},
		},
		// ---- generic-constrained array, comments in subscript ----
		{
			Code: `
function actingOnArray<T extends string[]>(values: T) {
  return values.filter(filter => filter === 'filter')[
    /* filter */ -0.0 /* filter */
  ];
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
function actingOnArray<T extends string[]>(values: T) {
  return values.find(filter => filter === 'filter');
}
`,
						},
					},
				},
			},
		},
		// ---- deeply nested sequence ----
		// The `['0']` binds to the LAST operand of the outer comma sequence
		// (comma has lower precedence than subscript), so it's actually on
		// the MIDDLE paren — that's why both upstream and we report line 5
		// (`  (1,`), not line 3 (the outer paren).
		{
			Code: `
const nestedSequenceAbomination =
  (1,
  2,
  (1,
  2,
  3,
  (1, 2, 3, 4),
  (1, 2, 3, 4, 5, [1, 2, 3, 4, 5, 6].filter(x => x % 2 == 0)))['0']);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      5,
					Column:    3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
const nestedSequenceAbomination =
  (1,
  2,
  (1,
  2,
  3,
  (1, 2, 3, 4),
  (1, 2, 3, 4, 5, [1, 2, 3, 4, 5, 6].find(x => x % 2 == 0))));
`,
						},
					},
				},
			},
		},
		// ---- intersection of two arrays (with thisArg) ----
		{
			Code: `
declare const arr: { a: 1 }[] & { b: 2 }[];
arr.filter(f, thisArg)[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: { a: 1 }[] & { b: 2 }[];
arr.find(f, thisArg);
`,
						},
					},
				},
			},
		},
		// ---- intersection of array and (union of arrays) ----
		{
			Code: `
declare const arr: { a: 1 }[] & ({ b: 2 }[] | { c: 3 }[]);
arr.filter(f, thisArg)[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const arr: { a: 1 }[] & ({ b: 2 }[] | { c: 3 }[]);
arr.find(f, thisArg);
`,
						},
					},
				},
			},
		},
		// ---- ternary where both branches are filter() ----
		{
			Code: `
(Math.random() < 0.5
  ? [1, 2, 3].filter(x => false)
  : [1, 2, 3].filter(x => true))[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      2,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
(Math.random() < 0.5
  ? [1, 2, 3].find(x => false)
  : [1, 2, 3].find(x => true));
`,
						},
					},
				},
			},
		},
		// ---- statement-level ternary; outer ternary is .at(0) consumer ----
		{
			Code: `
Math.random() < 0.5
  ? [1, 2, 3].find(x => true)
  : [1, 2, 3].filter(x => true)[0];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      4,
					Column:    5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
Math.random() < 0.5
  ? [1, 2, 3].find(x => true)
  : [1, 2, 3].find(x => true);
`,
						},
					},
				},
			},
		},
		// ---- nested ternaries — every leaf is a filter call ----
		{
			Code: `
declare const f: (arg0: unknown, arg1: number, arg2: Array<unknown>) => boolean,
  g: (arg0: unknown) => boolean;
const nestedTernaries = (
  Math.random() < 0.5
    ? Math.random() < 0.5
      ? [1, 2, 3].filter(f)
      : []?.filter(x => 'shrug')
    : [2, 3, 4]['filter'](g)
).at(0.2);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      4,
					Column:    25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const f: (arg0: unknown, arg1: number, arg2: Array<unknown>) => boolean,
  g: (arg0: unknown) => boolean;
const nestedTernaries = (
  Math.random() < 0.5
    ? Math.random() < 0.5
      ? [1, 2, 3].find(f)
      : []?.find(x => 'shrug')
    : [2, 3, 4]["find"](g)
);
`,
						},
					},
				},
			},
		},
		// ---- nested ternaries with sequence expression inside a branch ----
		{
			Code: `
declare const f: (arg0: unknown) => boolean, g: (arg0: unknown) => boolean;
const nestedTernariesWithSequenceExpression = (
  Math.random() < 0.5
    ? ('sequence',
      'expression',
      Math.random() < 0.5 ? [1, 2, 3].filter(f) : []?.filter(x => 'shrug'))
    : [2, 3, 4]['filter'](g)
).at(0.2);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    47,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const f: (arg0: unknown) => boolean, g: (arg0: unknown) => boolean;
const nestedTernariesWithSequenceExpression = (
  Math.random() < 0.5
    ? ('sequence',
      'expression',
      Math.random() < 0.5 ? [1, 2, 3].find(f) : []?.find(x => 'shrug'))
    : [2, 3, 4]["find"](g)
);
`,
						},
					},
				},
			},
		},
		// ---- spread args inside filter ----
		{
			Code: `
declare const spreadArgs: [(x: unknown) => boolean];
[1, 2, 3].filter(...spreadArgs).at(0);
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferFind",
					Line:      3,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferFindSuggestion",
							Output: `
declare const spreadArgs: [(x: unknown) => boolean];
[1, 2, 3].find(...spreadArgs);
`,
						},
					},
				},
			},
		},
	})
}
