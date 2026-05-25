package prefer_for_of

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferForOf(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferForOfRule, []rule_tester.ValidTestCase{
		// ---- Index used with two different arrays ----
		{Code: `
for (let i = 0; i < arr1.length; i++) {
  const x = arr1[i] === arr2[i];
}`},
		// ---- Index used as assignment target ----
		{Code: `
for (let i = 0; i < arr.length; i++) {
  arr[i] = 0;
}`},
		// ---- Index used directly (not just as array index) ----
		{Code: `
for (var c = 0; c < arr.length; c++) {
  doMath(c);
}`},
		{Code: `for (var d = 0; d < arr.length; d++) doMath(d);`},
		// ---- Index used both directly and as array index ----
		{Code: `
for (var e = 0; e < arr.length; e++) {
  if (e > 5) {
    doMath(e);
  }
  console.log(arr[e]);
}`},
		// ---- Not comparing with .length ----
		{Code: `
for (var f = 0; f <= 40; f++) {
  doMath(f);
}`},
		{Code: `for (var g = 0; g <= 40; g++) doMath(g);`},
		// ---- Multiple declarations in init ----
		{Code: `for (var h = 0, len = arr.length; h < len; h++) {}`},
		{Code: `for (var i = 0, len = arr.length; i < len; i++) arr[i];`},
		// ---- No init ----
		{Code: `
var m = 0;
for (;;) {
  if (m > 3) break;
  console.log(m);
  m++;
}`},
		{Code: `
var n = 0;
for (; n < 9; n++) {
  console.log(n);
}`},
		{Code: `
var o = 0;
for (; o < arr.length; o++) {
  console.log(arr[o]);
}`},
		{Code: `for (; x < arr.length; x++) {}`},
		// ---- Missing test or update ----
		{Code: `for (let x = 0; ; x++) {}`},
		{Code: `for (let x = 0; x < arr.length; ) {}`},
		// ---- Mismatched variable names ----
		{Code: `for (let x = 0; NOTX < arr.length; x++) {}`},
		{Code: `for (let x = 0; x < arr.length; NOTX++) {}`},
		{Code: `for (let NOTX = 0; x < arr.length; x++) {}`},
		// ---- Wrong update direction or operator ----
		{Code: `for (let x = 0; x < arr.length; x--) {}`},
		// ---- Wrong comparison operator ----
		{Code: `for (let x = 0; x <= arr.length; x++) {}`},
		// ---- Non-zero initializer ----
		{Code: `for (let x = 1; x < arr.length; x++) {}`},
		// ---- .length() is a function call, not property ----
		{Code: `for (let x = 0; x < arr.length(); x++) {}`},
		// ---- Wrong increment amount ----
		{Code: `for (let x = 0; x < arr.length; x += 11) {}`},
		// ---- Reverse iteration ----
		{Code: `for (let x = arr.length; x > 1; x -= 1) {}`},
		// ---- Multiply update ----
		{Code: `for (let x = 0; x < arr.length; x *= 2) {}`},
		// ---- Wrong addition amount in assignment ----
		{Code: `for (let x = 0; x < arr.length; x = x + 11) {}`},
		// ---- Index modified in body ----
		{Code: `
for (let x = 0; x < arr.length; x++) {
  x++;
}`},
		// ---- Test is not a comparison ----
		{Code: `for (let x = 0; true; x++) {}`},
		// ---- for-in and for-of (not for) ----
		{Code: `
for (var q in obj) {
  if (obj.hasOwnProperty(q)) {
    console.log(q);
  }
}`},
		{Code: `
for (var r of arr) {
  console.log(r);
}`},
		// ---- Index used in arithmetic with array index ----
		{Code: `
for (let x = 0; x < arr.length; x++) {
  let y = arr[x + 1];
}`},
		// ---- delete arr[i] ----
		{Code: `
for (let i = 0; i < arr.length; i++) {
  delete arr[i];
}`},
		// ---- arr[i]++ (update expression on element access) ----
		{Code: `
for (let i = 0; i < arr.length; i++) {
  arr[i]++;
}`},
		// ---- Non-null assertion + update ----
		{Code: `
for (let i = 0; i < arr.length; i++) {
  arr[i]!++;
}`},
		{Code: `
for (let i = 0; i < arr.length; i++) {
  arr[i]!!!++;
}`},
		// ---- Type assertion + update ----
		{Code: `
for (let i = 0; i < arr.length; i++) {
  (arr[i] as number)++;
}`},
		{Code: `
for (let i = 0; i < arr.length; i++) {
  (<number>arr[i])++;
}`},
		{Code: `
for (let i = 0; i < arr.length; i++) {
  (arr[i] as unknown as number)++;
}`},
		// ---- satisfies + update ----
		{Code: `
for (let i = 0; i < arr.length; i++) {
  (arr[i] satisfies number)++;
}`},
		{Code: `
for (let i = 0; i < arr.length; i++) {
  (arr[i]! satisfies number)++;
}`},
		// ---- Array destructuring assignment ----
		{Code: `
for (let i = 0; i < arr.length; i++) {
  [arr[i]] = [1];
}`},
		{Code: `
for (let i = 0; i < arr.length; i++) {
  [...arr[i]] = [1];
}`},
		{Code: `
for (let i = 0; i < arr.length; i++) {
  [...arr[i]!] = [1];
}`},
		// ---- Object destructuring assignment ----
		{Code: `
for (let i = 0; i < arr.length; i++) {
  ({ foo: arr[i] }) = { foo: 0 };
}`},
		// ---- Optional chaining on length (arr?.length) ----
		{Code: `
for (let i = 0; i < arr1?.length; i++) {
  const x = arr1[i] === arr2[i];
}`},
		{Code: `
for (let i = 0; i < arr?.length; i++) {
  arr[i] = 0;
}`},
		{Code: `
for (var c = 0; c < arr?.length; c++) {
  doMath(c);
}`},
		{Code: `for (var d = 0; d < arr?.length; d++) doMath(d);`},
		// ---- Optional chaining on function call (not length) ----
		{Code: `
for (var c = 0; c < arr.length; c++) {
  doMath?.(c);
}`},
		{Code: `for (var d = 0; d < arr.length; d++) doMath?.(d);`},
		// ---- this[i] (object is this) ----
		{Code: `
for (let i = 0; i < test.length; ++i) {
  this[i];
}`},
		// ---- this.length with this[i] ----
		{Code: `
for (let i = 0; i < this.length; ++i) {
  yield this[i];
}`},
		// ======== Extra edge cases: init variants ========
		// const init (can't be incremented, rejected by const flag check)
		{Code: `for (const i = 0; i < arr.length; i++) {}`},
		// TS type annotation on loop variable
		{Code: `
for (let i: number = 0; i < arr.length; i++) {
  arr[i] = 0;
}`},

		// ======== Extra edge cases: isIncrement uncovered branches ========
		// Locks in: AssignmentExpression `=` where RHS is not BinaryExpression
		{Code: `for (let x = 0; x < arr.length; x = 1) {}`},
		// Locks in: AssignmentExpression `=` where RHS BinaryExpression operator is not `+`
		{Code: `for (let x = 0; x < arr.length; x = x - 1) {}`},

		// ======== Extra edge cases: isAssignee uncovered branches ========
		// Compound assignment += on arr[i]
		{Code: `
for (let i = 0; i < arr.length; i++) {
  arr[i] += 1;
}`},
		// Logical assignment &&= on arr[i]
		{Code: `
for (let i = 0; i < arr.length; i++) {
  arr[i] &&= true;
}`},
		// Prefix decrement --arr[i]
		{Code: `
for (let i = 0; i < arr.length; i++) {
  --arr[i];
}`},
		// Nested array destructuring [[arr[i]]] = [[1]]
		{Code: `
for (let i = 0; i < arr.length; i++) {
  [[arr[i]]] = [[1]];
}`},

		// ======== Extra edge cases: optional chaining on .length ========
		// arr?.length implies arr may be null/undefined, for-of would throw
		{Code: `
for (let i = 0; i < arr?.length; i++) {
  console.log(arr[i]);
}`},

		// ======== Extra edge cases: parenthesized arr[((i))] as assignee ========
		{Code: `
for (let i = 0; i < arr.length; i++) {
  arr[((i))] = 0;
}`},

		// ======== Extra edge cases: non-null assertion on array object ========
		// arr!.length text is "arr!" which doesn't match "arr" in arr[i]
		{Code: `
for (let i = 0; i < arr!.length; i++) {
  arr[i] = 0;
}`},

		// ======== Extra edge cases: index used in nested closure ========
		// i is used inside arrow function — not a direct arr[i] at body level
		{Code: `
for (let i = 0; i < arr.length; i++) {
  setTimeout(() => console.log(i));
}`},

		// ======== Extra edge cases: deeply nested for loops (3 levels) ========
		{Code: `
for (let i = 0; i < arr.length; i++) {
  for (let j = 0; j < arr2.length; j++) {
    for (let k = 0; k < arr3.length; k++) {
      console.log(i, j, k);
    }
  }
}`},
	}, []rule_tester.InvalidTestCase{
		// ---- Nested member access (obj.arr) ----
		{
			Code: `
for (var a = 0; a < obj.arr.length; a++) {
  console.log(obj.arr[a]);
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Single-statement body (no braces) ----
		{
			Code: `for (var b = 0; b < arr.length; b++) console.log(arr[b]);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 1, Column: 1, EndLine: 1, EndColumn: 58},
			},
		},
		// ---- Basic let ----
		{
			Code: `
for (let a = 0; a < arr.length; a++) {
  console.log(arr[a]);
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Optional chaining in body (not on length) ----
		{
			Code: `for (var b = 0; b < arr.length; b++) console?.log(arr[b]);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 1, Column: 1, EndLine: 1, EndColumn: 59},
			},
		},
		{
			Code: `
for (let a = 0; a < arr.length; a++) {
  console?.log(arr[a]);
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Prefix increment (++a) ----
		{
			Code: `
for (let a = 0; a < arr.length; ++a) {
  arr[a].whatever();
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Empty body ----
		{
			Code: `for (let x = 0; x < arr.length; x++) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 1, Column: 1, EndLine: 1, EndColumn: 40,
					Message: "Expected a `for-of` loop instead of a `for` loop with this simple iteration."},
			},
		},
		// ---- x += 1 ----
		{
			Code: `for (let x = 0; x < arr.length; x += 1) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
			},
		},
		// ---- x = x + 1 ----
		{
			Code: `for (let x = 0; x < arr.length; x = x + 1) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
			},
		},
		// ---- x = 1 + x ----
		{
			Code: `for (let x = 0; x < arr.length; x = 1 + x) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
			},
		},
		// ---- Nested for loops with same variable name (shadow) ----
		{
			Code: `
for (let shadow = 0; shadow < arr.length; shadow++) {
  for (let shadow = 0; shadow < arr.length; shadow++) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				{MessageId: "preferForOf", Line: 3, Column: 3, EndLine: 3, EndColumn: 57},
			},
		},
		// ---- arr[i] used as computed key (not assignee) ----
		{
			Code: `
for (let i = 0; i < arr.length; i++) {
  obj[arr[i]] = 1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- delete on outer object, not arr[i] ----
		{
			Code: `
for (let i = 0; i < arr.length; i++) {
  delete obj[arr[i]];
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Destructuring on outer object ----
		{
			Code: `
for (let i = 0; i < arr.length; i++) {
  [obj[arr[i]]] = [1];
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		{
			Code: `
for (let i = 0; i < arr.length; i++) {
  [...obj[arr[i]]] = [1];
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		{
			Code: `
for (let i = 0; i < arr.length; i++) {
  ({ foo: obj[arr[i]] } = { foo: 1 });
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- this.item[i] (this is not the object directly) ----
		{
			Code: `
for (let i = 0; i < this.item.length; ++i) {
  this.item[i];
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- this.array[i] with yield ----
		{
			Code: `
for (let i = 0; i < this.array.length; ++i) {
  yield this.array[i];
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Extra: TS type annotation on loop variable — still reports ----
		{
			Code: `
for (let i: number = 0; i < arr.length; i++) {
  console.log(arr[i]);
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Extra: multiple valid refs in body (all read-only) ----
		{
			Code: `
for (let i = 0; i < arr.length; i++) {
  console.log(arr[i] + arr[i]);
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Extra: index used inside arrow function but only as arr[i] ----
		{
			Code: `
for (let i = 0; i < arr.length; i++) {
  setTimeout(() => console.log(arr[i]));
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Extra: deep chain a.b.c.length / a.b.c[i] ----
		{
			Code: `
for (let i = 0; i < a.b.c.length; i++) {
  console.log(a.b.c[i]);
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Extra: parenthesized index arr[(i)] — aligned with ESTree paren-stripping ----
		{
			Code: `
for (let i = 0; i < arr.length; i++) {
  console.log(arr[(i)]);
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Extra: multi-level parenthesized index arr[((i))] ----
		{
			Code: `
for (let i = 0; i < arr.length; i++) {
  console.log(arr[((i))]);
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// ---- Extra: deeply nested for loops, all read-only ----
		{
			Code: `
for (let i = 0; i < a.length; i++) {
  for (let j = 0; j < b.length; j++) {
    for (let k = 0; k < c.length; k++) {
      console.log(a[i], b[j], c[k]);
    }
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferForOf", Line: 2, Column: 1, EndLine: 8, EndColumn: 2},
				{MessageId: "preferForOf", Line: 3, Column: 3, EndLine: 7, EndColumn: 4},
				{MessageId: "preferForOf", Line: 4, Column: 5, EndLine: 6, EndColumn: 6},
			},
		},
	})
}
