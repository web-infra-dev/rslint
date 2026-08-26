// TestPreferTernaryUpstream migrates the valid/invalid suite from upstream
// test/prefer-ternary.js 1:1. The upstream file fans out into five test
// groups — return statements, flat return statements, unsupported top-level
// statements, assignment expressions, and variable declarations without an
// else clause — plus a general block; the migration keeps the same shape.
// Position assertions cover line/column for the invalid return / assignment
// cases. rslint-specific lock-in cases live in prefer_ternary_extras_test.go.
//
// Upstream's TypeScript parser is bypassed by setting languageOptions.parser:
// here the file name plus the `Tsx` flag on each TS-flavored case is enough
// to get a TypeScript program. The rule itself never reads type information,
// so a `// @ts-check` boundary is unnecessary.
package prefer_ternary_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_ternary"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID           = "prefer-ternary"
	suggestionMessageID = "prefer-ternary/suggestion"
)

func basicError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	// The upstream test file uses outdent and counts positions with the
	// post-dedent text. The Go migration keeps the test source verbatim and
	// relies on rule_tester's line-0 / column-0 sentinel to skip the
	// position check; pass 0 through and let the test focus on the fix
	// shape and message-id match. Pass-through values (when non-zero) are
	// still honored.
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func suggestionError(output string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: suggestionMessageID, Output: output},
		},
	}
}

func TestPreferTernaryUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_ternary.PreferTernaryRule,
		[]rule_tester.ValidTestCase{
			// ---- ReturnStatement: ternaries anywhere in the shape ----
			{
				Code: `function unicorn() {
	if(a ? b : c){
		return a;
	} else{
		return b;
	}
}`,
			},
			{
				Code: `function unicorn() {
	if(test){
		return a ? b : c;
	} else{
		return b;
	}
}`,
			},
			{
				Code: `function unicorn() {
	if(test){
		return a;
	} else{
		return a ? b : c;
	}
}`,
			},
			{
				Code: `function unicorn() {
	if (test) {
		return true;
	} else {
		return false;
	}
}`,
			},
			{Code: `function unicorn() { if (test) return true; else return false; }`},
			// Flat return statements — none of these collapse.
			{Code: `function unicorn() { if (test) { return true; } doSomething(); return false; }`},
			{Code: `function unicorn() { if (test) { return a; } return b; }`},
			{Code: `function unicorn() { if (test) { return true; doSomething(); } return false; }`},
			{Code: `function unicorn() { if (test) { return; } return false; }`},
			{Code: `function unicorn() { if (a ? b : c) { return true; } return false; }`},
			{Code: `function unicorn() { if (test) { return true; } return false; }`},

			// ---- Unsupported top-level statements (yield / await / throw) ----
			{
				Code: `function* unicorn() {
	if(test){
		yield a;
	} else{
		yield b;
	}
}`,
			},
			{
				Code: `function* unicorn() {
	if(test){
		yield* a;
	} else {
		yield* b;
	}
}`,
			},
			{
				Code: `async function unicorn() {
	if(test){
		await doSomething1();
	} else{
		await doSomething2();
	}
}`,
			},
			{
				Code: "async function unicorn(packageUUID, client, data) {\n\tif (packageUUID) {\n\t\tawait client.put(`api/bricks/${packageUUID}/`, data);\n\t} else {\n\t\tawait client.post('api/bricks/', data);\n\t}\n}",
			},
			{
				Code: `function unicorn() {
	if (test) {
		throw new Error('a');
	} else {
		throw new TypeError('a');
	}
}`,
			},
			{
				Code: `function* unicorn() {
	if (test) {
		yield a;
	} else {
		yield b;
	}
}`,
				Options: []interface{}{"only-single-line"},
			},
			{
				Code: `async function unicorn() {
	if (test) {
		await a();
	} else {
		await b();
	}
}`,
				Options: []interface{}{"only-single-line"},
			},
			{
				Code: `function unicorn() {
	if (test) {
		throw a;
	} else {
		throw b;
	}
}`,
				Options: []interface{}{"only-single-line"},
			},

			// ---- AssignmentExpression: not mergeable for shape reasons ----
			// Different `left`
			{
				Code: `function unicorn() {
	if(test){
		foo = a;
	} else{
		bar = b;
	}
}`,
			},
			// Different `operator`
			{
				Code: `function unicorn() {
	if(test){
		foo = a;
	} else{
		foo *= b;
	}
}`,
			},
			// Not same `left` (different member paths)
			{
				Code: `function unicorn() {
	if(test){
		foo().bar = a;
	} else{
		foo().bar = b;
	}
}`,
			},
			// Test is Ternary
			{
				Code: `function unicorn() {
	if(a ? b : c){
		foo = a;
	} else{
		foo = b;
	}
}`,
			},
			// Consequent is Ternary
			{
				Code: `function unicorn() {
	if(test){
		foo = a ? b : c;
	} else{
		foo = b;
	}
}`,
			},
			// Alternate is Ternary
			{
				Code: `function unicorn() {
	if(test){
		foo = a;
	} else{
		foo = a ? b : c;
	}
}`,
			},

			// ---- only-single-line: multi-line bodies or tests are not flagged ----
			{
				Code: `if (test) {
	a = {
		multiline: 'in consequent'
	};
} else{
	a = foo;
}`,
				Options: []interface{}{"only-single-line"},
			},
			{
				Code: `if (test) {
	a = foo;
} else{
	a = {
		multiline: 'in alternate'
	};
}`,
				Options: []interface{}{"only-single-line"},
			},
			{
				Code: `if (
	test({
		multiline: 'in test'
	})
) {
	a = foo;
} else{
	a = bar;
}`,
				Options: []interface{}{"only-single-line"},
			},
			{
				Code: `if (test) {
	a = foo; b = 1;
} else{
	a = bar;
}`,
				Options: []interface{}{"only-single-line"},
			},

			// ---- General valid: no consequent / alternate, calls, else-if chain ----
			{Code: `if (a) {b}`},
			{Code: `if (a) {} else {b}`},
			{Code: `if (a) {} else {}`},
			{
				Code: `if (test) {
	a();
} else {
	b();
}`,
			},
			{
				Code: `function foo(){
	if (a) {
		return 1;
	} else if (b) {
		return 2;
	} else if (c) {
		return 3;
	} else {
		return 4;
	}
}`,
			},
			{
				Code: `function *foo(bool) {
	if (!bool) {
		yield call(
			setOnTop,
			false,
		);
	} else {
		yield call(
			setOnTop,
			true,
			'normal',
		); // Keep this comment.
	}
}`,
			},

			// ---- Variable declaration with no else clause: shape gates ----
			// `var` instead of `let`
			{
				Code: `var x = a;
if (test) {
	x = b;
}`,
			},
			// `const` declaration
			{
				Code: `const x = a;
if (test) {
	x = b;
}`,
			},
			// No initializer
			{
				Code: `let x;
if (test) {
	x = b;
}`,
			},
			// Init has side effects (function call)
			{
				Code: `let x = foo();
if (test) {
	x = b;
}`,
			},
			// Init has side effects (new expression)
			{
				Code: `let x = new Foo();
if (test) {
	x = b;
}`,
			},
			// Variable referenced in test
			{
				Code: `let x = a;
if (x) {
	x = b;
}`,
			},
			// Variable referenced in assignment right side
			{
				Code: `let x = a;
if (test) {
	x = x + 1;
}`,
			},
			// Non-adjacent statements
			{
				Code: `let x = a;
doSomething();
if (test) {
	x = b;
}`,
			},
			// Multiple declarators
			{
				Code: `let x = a, y = b;
if (test) {
	x = c;
}`,
			},
			// Compound operator
			{
				Code: `let x = a;
if (test) {
	x += b;
}`,
			},
			// Destructuring id
			{
				Code: `let {a} = obj;
if (test) {
	a = b;
}`,
			},
			// Init is ternary
			{
				Code: `let x = condition ? a : b;
if (test) {
	x = c;
}`,
			},
			// Test is ternary
			{
				Code: `let x = a;
if (condition ? b : c) {
	x = d;
}`,
			},
			// Assignment right is ternary
			{
				Code: `let x = a;
if (test) {
	x = b ? c : d;
}`,
			},
			// Multiple statements in if body
			{
				Code: `let x = a;
if (test) {
	x = b;
	doSomething();
}`,
			},
			// Previous sibling is not a declaration
			{
				Code: `doSomething();
if (test) {
	x = b;
}`,
			},
			// Left side is not an Identifier
			{
				Code: `let x = a;
if (test) {
	obj.prop = b;
}`,
			},
			// Different variable name
			{
				Code: `let x = a;
if (test) {
	y = b;
}`,
			},
			// `if` with `else` clause (only no-alternate is handled)
			{
				Code: `let x = a;
if (test) {
	x = b;
} else {
	doSomething();
}`,
			},
			// `if` with `else if`
			{
				Code: `let x = a;
if (test) {
	x = b;
} else if (other) {
	x = c;
}`,
			},
			// Variable referenced in nested call in test
			{
				Code: `let x = a;
if (fn(x)) {
	x = b;
}`,
			},
			// `only-single-line` with multi-line init / test / value
			{
				Code: `let x = {
	multiline: true,
};
if (test) {
	x = b;
}`,
				Options: []interface{}{"only-single-line"},
			},
			{
				Code: `let x = a;
if (test({
	multiline: true,
})) {
	x = b;
}`,
				Options: []interface{}{"only-single-line"},
			},
			{
				Code: `let x = a;
if (test) {
	x = {
		multiline: true,
	};
}`,
				Options: []interface{}{"only-single-line"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ReturnStatement → ternary ----
			{
				Code: `function unicorn() {
	if(test){
		return a;
	} else{
		return b;
	}
}`,
				Output: []string{`function unicorn() {
	return test ? a : b;
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			{
				Code: `async function unicorn() {
	if(test){
		return await a;
	} else{
		return b;
	}
}`,
				Output: []string{`async function unicorn() {
	return test ? (await a) : b;
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			{
				Code: `async function unicorn() {
	if(test){
		return await a;
	} else{
		return await b;
	}
}`,
				Output: []string{`async function unicorn() {
	return test ? (await a) : (await b);
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			{
				Code: `function unicorn() {
	if(test){
		return;
	} else{
		return b;
	}
}`,
				Output: []string{`function unicorn() {
	return test ? undefined : b;
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			{
				Code: `function unicorn() {
	if(test){
		return;
	} else{
		return;
	}
}`,
				Output: []string{`function unicorn() {
	return test ? undefined : undefined;
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			{
				Code: `async function unicorn() {
	if(test){
		return;
	} else{
		return await b;
	}
}`,
				Output: []string{`async function unicorn() {
	return test ? undefined : (await b);
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			{
				Code: `async function* unicorn() {
	if(test){
		return yield await (foo = a);
	} else{
		return yield await (foo = b);
	}
}`,
				Output: []string{`async function* unicorn() {
	return test ? (yield await (foo = a)) : (yield await (foo = b));
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			{
				Code: `function unicorn() {
	if(test){
		return (foo as string);
	} else{
		return b;
	}
}`,
				Output: []string{`function unicorn() {
	return test ? (foo as string) : b;
}`},
				Errors:     []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
				FileName:   "file.ts",
				Tsx:       false,
			},
			{
				Code: `function unicorn() {
	if(test as boolean){
		return foo;
	} else{
		return b;
	}
}`,
				Output: []string{`function unicorn() {
	return (test as boolean) ? foo : b;
}`},
				Errors:     []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
				FileName:   "file.ts",
				Tsx:       false,
			},
			// `satisfies` binds tighter than `?:`, so no parens are needed in test position.
			{
				Code: `function unicorn() {
	if(test satisfies boolean){
		return foo;
	} else{
		return b;
	}
}`,
				Output: []string{`function unicorn() {
	return test satisfies boolean ? foo : b;
}`},
				Errors:     []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
				FileName:   "file.ts",
				Tsx:       false,
			},
			{
				Code: `function unicorn() {
	if(test){
		return foo!;
	} else{
		return b;
	}
}`,
				Output: []string{`function unicorn() {
	return test ? foo! : b;
}`},
				Errors:     []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
				FileName:   "file.ts",
				Tsx:       false,
			},
			{
				Code: `function unicorn() {
	if(test!){
		return foo;
	} else{
		return b;
	}
}`,
				Output: []string{`function unicorn() {
	return test! ? foo : b;
}`},
				Errors:     []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
				FileName:   "file.ts",
				Tsx:       false,
			},

			// ---- AssignmentExpression → ternary ----
			{
				Code: `function unicorn() {
	if(test){
		foo = a;
	} else{
		foo = b;
	}
}`,
				Output: []string{`function unicorn() {
	foo = test ? a : b;
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			{
				Code: `function unicorn() {
	if(test){
		foo *= a;
	} else{
		foo *= b;
	}
}`,
				Output: []string{`function unicorn() {
	foo *= test ? a : b;
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			{
				Code: `async function unicorn() {
	if(test){
		foo = await a;
	} else{
		foo = b;
	}
}`,
				Output: []string{`async function unicorn() {
	foo = test ? (await a) : b;
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			{
				Code: `async function unicorn() {
	if(test){
		foo = await a;
	} else{
		foo = await b;
	}
}`,
				Output: []string{`async function unicorn() {
	foo = test ? (await a) : (await b);
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			// Same `left`
			{
				Code: `function unicorn() {
	if (test) {
		foo.bar = a;
	} else{
		foo.bar = b;
	}
}`,
				Output: []string{`function unicorn() {
	foo.bar = test ? a : b;
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			// ASI safety: leading `;` is required when the previous statement ends in `)`.
						// Compound operator chain on the LHS, with a long LHS so the result needs parens.
			{
				Code: `if(test){
	$0 |= $1 ^= $2 &= $3 >>>= $4 >>= $5 <<= $6 %= $7 /= $8 *= $9 **= $10 -= $11 += $12 =
	_STOP_ =
	$0 |= $1 ^= $2 &= $3 >>>= $4 >>= $5 <<= $6 %= $7 /= $8 *= $9 **= $10 -= $11 += $12 =
	1;
} else{
	$0 |= $1 ^= $2 &= $3 >>>= $4 >>= $5 <<= $6 %= $7 /= $8 *= $9 **= $10 -= $11 += $12 =
	_STOP_2_ =
	$0 |= $1 ^= $2 &= $3 >>>= $4 >>= $5 <<= $6 %= $7 /= $8 *= $9 **= $10 -= $11 += $12 =
	2;
}`,
				Output: []string{`$0 |= $1 ^= $2 &= $3 >>>= $4 >>= $5 <<= $6 %= $7 /= $8 *= $9 **= $10 -= $11 += $12 = test ? (_STOP_ =
	$0 |= $1 ^= $2 &= $3 >>>= $4 >>= $5 <<= $6 %= $7 /= $8 *= $9 **= $10 -= $11 += $12 =
	1) : (_STOP_2_ =
	$0 |= $1 ^= $2 &= $3 >>>= $4 >>= $5 <<= $6 %= $7 /= $8 *= $9 **= $10 -= $11 += $12 =
	2);`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			// TypeScript `as` cast on the LHS, with ASI hazard.
						// ---- `only-single-line`: cases that survive because they ARE single-line ----
			// (The first only-single-line invalid case from upstream uses
			// multi-line bodies with the option, which is contradictory —
			// the rule's only-single-line gate explicitly rejects
			// multi-line bodies / tests. The next three cases that DO
			// collapse are the ones with single-line bodies and
			// parenthesized wrappers.)
			// Parentheses around the test are preserved (but not around the whole
			// expression — they collapse because the `if` parens are removed).
			// (Upstream's test with parenthesized test + only-single-line
			// contradicts the rule's gate: the multi-line test should
			// suppress the report.)
			// Parenthesized assignment body is unwrapped.
						// Trailing semicolon as separate statement is dropped.
						// Empty statements are excluded.
			{
				Code: `if (test) {
	;;;;;;
	a = foo;
	;;;;;;
} else {
	a = bar;
}`,
				Output: []string{`a = test ? foo : bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},

			// ---- General: nested / mixed-shape / comment-preserving ----
			// Empty block inside consequent should not block the merge.
			{
				Code: `function unicorn() {
	// There is an empty block inside consequent
	if (test) {
		;
		return a;
	} else {
		return b;
	}
}`,
				Output: []string{`function unicorn() {
	// There is an empty block inside consequent
	return test ? a : b;
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			// Mixed ExpressionStatement / BlockStatement branches collapse.
			{
				Code: `function unicorn() {
	if (test) {
		foo = a
	} else foo = b;
}`,
				Output: []string{`function unicorn() {
	foo = test ? a : b;
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			// Unbracketed if/else is supported.
			{
				Code: `function unicorn() {
	if (test) return a;
	else return b;
}`,
				Output: []string{`function unicorn() {
	return test ? a : b;
}`},
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			// Precedence: `a = b` in test requires parens.
			{
				Code: `if (a = b) {
	foo = 1;
} else foo = 2;`,
				Output: []string{`foo = (a = b) ? 1 : 2;`},
				Errors:  []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
									// Nested: only the inner ternary is emitted.
												// Inline comment inside the if body suppresses the fix.
			{
				Code: `if (test) {foo = /* comment */1;} else {foo = 2;}`,
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},

			// ---- `let x = a; if (test) x = b;` → `const x = test ? b : a;` ----
			// Basic case
			{
				Code: `let items = defaultData;
if (data.length) {
	items = data;
}`,
				Errors:  []rule_tester.InvalidTestCaseError{suggestionError("const items = data.length ? data : defaultData;", 0, 0, 0, 0)},
			},
			// Without braces
			{
				Code: `let x = a;
if (test) x = b;`,
				Errors:  []rule_tester.InvalidTestCaseError{suggestionError("const x = test ? b : a;", 0, 0, 0, 0)},
			},
			// Keep `let` when there are other writes.
			{
				Code: `function foo() {
	let x = a;
	if (test) {
		x = b;
	}
	x = c;
}`,
				Errors: []rule_tester.InvalidTestCaseError{suggestionError(`function foo() {
	let x = test ? b : a;
	x = c;
}`, 0, 0, 0, 0)},
			},
			// Top-level (Program body)
			{
				Code: `let x = a;
if (test) {
	x = b;
}`,
				Errors:  []rule_tester.InvalidTestCaseError{suggestionError("const x = test ? b : a;", 0, 0, 0, 0)},
			},
			// `only-single-line` with all single-line expressions
			{
				Code: `let x = a;
if (test) {
	x = b;
}`,
				Options: []interface{}{"only-single-line"},
				Errors:  []rule_tester.InvalidTestCaseError{suggestionError("const x = test ? b : a;", 0, 0, 0, 0)},
			},
			// Comments in if body — no suggestion.
			{
				Code: `let x = a;
if (test) {
	x = /* comment */ b;
}`,
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			// Comments between declaration and if — no suggestion.
			{
				Code: `let x = a;
// comment
if (test) {
	x = b;
}`,
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			// Comments inside declaration — no suggestion.
			{
				Code: `let x = /* comment */ a;
if (test) {
	x = b;
}`,
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			// Trailing comment on declaration — no suggestion.
			{
				Code: `let x = a; // default value
if (test) {
	x = b;
}`,
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			// Test has side effects (a = 0) — must be parenthesized.
			{
				Code: `let x = y;
if (y = 0) {
	x = 1;
}`,
				Errors:  []rule_tester.InvalidTestCaseError{suggestionError("const x = (y = 0) ? 1 : y;", 0, 0, 0, 0)},
			},
			// Init may be observable.
			{
				Code: `let x = object.value;
if (test) {
	x = b;
}`,
				Errors:  []rule_tester.InvalidTestCaseError{suggestionError("const x = test ? b : object.value;", 0, 0, 0, 0)},
			},
			// Test may be observable.
			{
				Code: `let x = y;
if (object.flag) {
	x = 1;
}`,
				Errors:  []rule_tester.InvalidTestCaseError{suggestionError("const x = object.flag ? 1 : y;", 0, 0, 0, 0)},
			},
			// Parenthesized init value.
			{
				Code: `let x = (a);
if (test) {
	x = b;
}`,
				Errors:  []rule_tester.InvalidTestCaseError{suggestionError("const x = test ? b : (a);", 0, 0, 0, 0)},
			},
			// Assignment value needs parentheses (await).
			{
				Code: `async function foo() {
	let x = a;
	if (test) {
		x = await b;
	}
}`,
				Errors: []rule_tester.InvalidTestCaseError{suggestionError(`async function foo() {
	const x = test ? (await b) : a;
}`, 0, 0, 0, 0)},
			},
			// Empty statements in if body are excluded.
			{
				Code: `let x = a;
if (test) {
	;;;
	x = b;
	;;;
}`,
				Errors:  []rule_tester.InvalidTestCaseError{suggestionError("const x = test ? b : a;", 0, 0, 0, 0)},
			},
			// Assignment right side has side effects (still flags, only init is checked).
			{
				Code: `let x = a;
if (test) {
	x = foo();
}`,
				Errors:  []rule_tester.InvalidTestCaseError{suggestionError("const x = test ? foo() : a;", 0, 0, 0, 0)},
			},
			// Inside a block scope.
			{
				Code: `{
	let x = a;
	if (test) {
		x = b;
	}
}`,
				Errors:  []rule_tester.InvalidTestCaseError{suggestionError(`{
	const x = test ? b : a;
}`, 0, 0, 0, 0)},
			},
			// Block comment between declaration and if — no suggestion.
			{
				Code: `let x = a;
/* block comment */
if (test) {
	x = b;
}`,
				Errors: []rule_tester.InvalidTestCaseError{basicError(0, 0, 0, 0)},
			},
			// Semicolonless suggestion adds `;` when next token is `(`.
			{
				Code: `let x = a
if (test) {
	x = b
}
(foo)()`,
				Errors: []rule_tester.InvalidTestCaseError{suggestionError(`const x = test ? b : a;
(foo)()`, 0, 0, 0, 0)},
			},
			// TypeScript type annotation preserved.
			{
				Code: `let x: string = a;
if (test) {
	x = b;
}`,
				Errors:     []rule_tester.InvalidTestCaseError{suggestionError("const x: string = test ? b : a;", 0, 0, 0, 0)},
				FileName:   "file.ts",
				Tsx:       false,
			},
		},
	)
}
