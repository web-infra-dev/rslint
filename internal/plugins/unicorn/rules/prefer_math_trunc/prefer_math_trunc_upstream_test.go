package prefer_math_trunc_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_math_trunc"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-math-trunc.js
// Snapshot cases use the diagnostics and suggestions from Unicorn 79.0.0 and ESLint 10.4.0.
func TestPreferMathTruncUpstream(t *testing.T) {
	root := fixtures.GetRootDir()
	t.Run("Parse-int suggestions with comments", func(t *testing.T) {
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &prefer_math_trunc.PreferMathTruncRule,
			nil, []rule_tester.InvalidTestCase{
				truncationCase("input.js", "function foo() {return(parseInt)(String(read(/* keep */)), 10);}",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 23, 1, 63,
						"function foo() {return Math.trunc(read(/* keep */));}"),
				),
				truncationCase("input.js", "function foo() {throw(Number.parseInt)(String(read(/* keep */)), 10);}",
					truncationError("parse-int", "Number.parseInt(String(...), 10)", 1, 22, 1, 69,
						"function foo() {throw Math.trunc(read(/* keep */));}"),
				),
				truncationCase("input.js", "async function foo() {return await(parseInt)(String(read(/* keep */)), 10);}",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 35, 1, 75,
						"async function foo() {return await Math.trunc(read(/* keep */));}"),
				),
			})
	})
	t.Run("Bitwise suggestions with comments", func(t *testing.T) {
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &prefer_math_trunc.PreferMathTruncRule,
			nil, []rule_tester.InvalidTestCase{
				truncationCase("input.js", "read(/* keep */) | 0;",
					truncationError("bitwise", "| 0", 1, 1, 1, 21,
						"Math.trunc(read(/* keep */));"),
				),
				truncationCase("input.js", "(/* keep */ value) >> 0;",
					truncationError("bitwise", ">> 0", 1, 1, 1, 24,
						"Math.trunc((/* keep */ value));"),
				),
				truncationCase("input.js", "values[/* keep */ index] << 0;",
					truncationError("bitwise", "<< 0", 1, 1, 1, 30,
						"Math.trunc(values[/* keep */ index]);"),
				),
				truncationCase("input.js", "(left + /* keep */ right) ^ 0;",
					truncationError("bitwise", "^ 0", 1, 1, 1, 30,
						"Math.trunc(left + /* keep */ right);"),
				),
				truncationCase("input.js", "(read(/* keep */), value) | 0;",
					truncationError("bitwise", "| 0", 1, 1, 1, 30,
						"Math.trunc((read(/* keep */), value));"),
				),
				truncationCase("input.js", "function foo() {return[/* keep */ value][0] | 0;}",
					truncationError("bitwise", "| 0", 1, 23, 1, 48,
						"function foo() {return Math.trunc([/* keep */ value][0]);}"),
				),
				truncationCase("input.js", "function foo() {\r\n  return(value // keep\r\n  ) >> 0;\r\n}",
					truncationError("bitwise", ">> 0", 2, 9, 3, 9,
						"function foo() {\r\n  return Math.trunc((value // keep\r\n  ));\r\n}"),
				),
				truncationCase("input.ts", "(value as /* keep */ number) | 0;",
					truncationError("bitwise", "| 0", 1, 1, 1, 33,
						"Math.trunc(value as /* keep */ number);"),
				),
				truncationCase("input.ts", "values[/* keep */ index]! ^ 0;",
					truncationError("bitwise", "^ 0", 1, 1, 1, 30,
						"Math.trunc(values[/* keep */ index]!);"),
				),
				truncationCase("input.js", "value | /* keep */ 0;",
					truncationError("bitwise", "| 0", 1, 1, 1, 21,
						""),
				),
				truncationCase("input.js", "value | (/* keep */ 0);",
					truncationError("bitwise", "| 0", 1, 1, 1, 23,
						""),
				),
				truncationCase("input.js", "read(/* keep */) /* drop */ | 0;",
					truncationError("bitwise", "| 0", 1, 1, 1, 32,
						""),
				),
				truncationCase("input.js", "values[/* keep */ index] |= 0;",
					truncationError("bitwise", "|= 0", 1, 1, 1, 30,
						""),
				),
				truncationCase("input.js", "(/* keep */ value) >>= 0;",
					truncationError("bitwise", ">>= 0", 1, 1, 1, 25,
						""),
				),
			})
	})
	t.Run("Double-NOT suggestions with comments", func(t *testing.T) {
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &prefer_math_trunc.PreferMathTruncRule,
			nil, []rule_tester.InvalidTestCase{
				truncationCase("input.js", "~~getValue(/* keep */);",
					truncationError("bitwise-not", "~~", 1, 1, 1, 23,
						"Math.trunc(getValue(/* keep */));"),
				),
				truncationCase("input.js", "~~(/* keep */ value);",
					truncationError("bitwise-not", "~~", 1, 1, 1, 21,
						"Math.trunc((/* keep */ value));"),
				),
				truncationCase("input.js", "~~values[/* keep */ index];",
					truncationError("bitwise-not", "~~", 1, 1, 1, 27,
						"Math.trunc(values[/* keep */ index]);"),
				),
				truncationCase("input.js", "~~(left + /* keep */ right);",
					truncationError("bitwise-not", "~~", 1, 1, 1, 28,
						"Math.trunc(left + /* keep */ right);"),
				),
				truncationCase("input.js", "~~(first(/* keep */), value);",
					truncationError("bitwise-not", "~~", 1, 1, 1, 29,
						"Math.trunc((first(/* keep */), value));"),
				),
				truncationCase("input.js", "~~~getValue(/* keep */);",
					truncationError("bitwise-not", "~~", 1, 2, 1, 24,
						"~Math.trunc(getValue(/* keep */));"),
				),
				truncationCase("input.js", "function foo() {return~~getValue(/* keep */);}",
					truncationError("bitwise-not", "~~", 1, 23, 1, 45,
						"function foo() {return Math.trunc(getValue(/* keep */));}"),
				),
				truncationCase("input.js", "async function foo() { return ~~(await /* keep */ getValue()); }",
					truncationError("bitwise-not", "~~", 1, 31, 1, 62,
						"async function foo() { return Math.trunc(await /* keep */ getValue()); }"),
				),
				truncationCase("input.js", "function* foo() { return ~~(yield /* keep */ getValue()); }",
					truncationError("bitwise-not", "~~", 1, 26, 1, 57,
						"function* foo() { return Math.trunc(yield /* keep */ getValue()); }"),
				),
				truncationCase("input.js", "function foo() {\r\n  return~~(value // keep\r\n  );\r\n}",
					truncationError("bitwise-not", "~~", 2, 9, 3, 4,
						"function foo() {\r\n  return Math.trunc((value // keep\r\n  ));\r\n}"),
				),
				truncationCase("input.ts", "~~(value as /* keep */ number);",
					truncationError("bitwise-not", "~~", 1, 1, 1, 31,
						"Math.trunc(value as /* keep */ number);"),
				),
				truncationCase("input.ts", "~~values[/* keep */ index]!;",
					truncationError("bitwise-not", "~~", 1, 1, 1, 28,
						"Math.trunc(values[/* keep */ index]!);"),
				),
				truncationCase("input.js", "~~ /* keep */ value;",
					truncationError("bitwise-not", "~~", 1, 1, 1, 20,
						""),
				),
				truncationCase("input.js", "~(/* keep */ ~value);",
					truncationError("bitwise-not", "~~", 1, 1, 1, 21,
						""),
				),
			})
	})
	t.Run("Upstream snapshots", func(t *testing.T) {
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &prefer_math_trunc.PreferMathTruncRule,
			[]rule_tester.ValidTestCase{
				{Code: "const foo = 1 | 1;", FileName: "input.js"},
				{Code: "const foo = 0 | 1;", FileName: "input.js"},
				{Code: "const foo = 1.4 | +0;", FileName: "input.js"},
				{Code: "const foo = 1.4 | -0;", FileName: "input.js"},
				{Code: "const foo = 1.4 | (.5 - 0.5);", FileName: "input.js"},
				{Code: "const foo = 1.4 & 0xFFFFFFFF", FileName: "input.js"},
				{Code: "const foo = 1.4 & 0xFF", FileName: "input.js"},
				{Code: "const foo = 1.4 & 0x0", FileName: "input.js"},
				{Code: "const foo = 1.4 & 0", FileName: "input.js"},
				{Code: "const foo = ~3.9;", FileName: "input.js"},
				{Code: "const foo = 1.1 >> 1", FileName: "input.js"},
				{Code: "const foo = 0 << 1", FileName: "input.js"},
				{Code: `let foo = 0;
foo |= 1;`, FileName: "input.js"},
				{Code: `let foo = 1.2; // comment 1
foo |= 1; // comment 2 and 1.2 | 0`, FileName: "input.js"},
				{Code: "parseInt(value, 10);", FileName: "input.js"},
				{Code: "parseInt(String(value));", FileName: "input.js"},
				{Code: "parseInt(String(value), 16);", FileName: "input.js"},
				{Code: "Number.parseInt(String(value), 16);", FileName: "input.js"},
				{Code: "parseInt(String(value), \"10\");", FileName: "input.js"},
				{Code: "parseInt?.(String(value), 10);", FileName: "input.js"},
				{Code: "parseInt(String(...value), 10);", FileName: "input.js"},
				{Code: "Number.parseInt(String(...value), 10);", FileName: "input.js"},
				{Code: "parseInt(String(value), ...[10]);", FileName: "input.js"},
				{Code: "Number.parseInt(String(value), ...[10]);", FileName: "input.js"},
				{Code: "parseInt(String(value, extra), 10);", FileName: "input.js"},
				{Code: "parseInt(new String(value), 10);", FileName: "input.js"},
				{Code: "parseInt(String(value), 10, extra);", FileName: "input.js"},
				{Code: "Number.parseInt(String(value), 10, extra);", FileName: "input.js"},
				{Code: "Number.parseInt?.(String(value), 10);", FileName: "input.js"},
				{Code: "Number?.parseInt(String(value), 10);", FileName: "input.js"},
				{Code: "Number[\"parseInt\"](String(value), 10);", FileName: "input.js"},
				{Code: "const parseInt = value => value; parseInt(String(value), 10);", FileName: "input.js"},
				{Code: "const Number = {parseInt() {}}; Number.parseInt(String(value), 10);", FileName: "input.js"},
				{Code: "const String = value => value; parseInt(String(value), 10);", FileName: "input.js"},
				{Code: "const String = value => value; Number.parseInt(String(value), 10);", FileName: "input.js"},
			}, []rule_tester.InvalidTestCase{
				truncationCase("input.js", "const foo = parseInt(String(number), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 41,
						"const foo = Math.trunc(number);"),
				),
				truncationCase("input.js", "const foo = Number.parseInt(String(number), 10);",
					truncationError("parse-int", "Number.parseInt(String(...), 10)", 1, 13, 1, 48,
						"const foo = Math.trunc(number);"),
				),
				truncationCase("input.js", "const foo = parseInt(String((number)), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 43,
						"const foo = Math.trunc((number));"),
				),
				truncationCase("input.js", "const foo = parseInt(String(number + 1), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 45,
						"const foo = Math.trunc(number + 1);"),
				),
				truncationCase("input.js", "const foo = parseInt(String(number,), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 42,
						"const foo = Math.trunc(number,);"),
				),
				truncationCase("input.js", "const foo = parseInt(String(/* comment */ number), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 55,
						"const foo = Math.trunc(/* comment */ number);"),
				),
				truncationCase("input.js", "const foo = parseInt(String(number /* comment */), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 55,
						"const foo = Math.trunc(number /* comment */);"),
				),
				truncationCase("input.js", "const foo = parseInt(/* comment */ String(number), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 55,
						""),
				),
				truncationCase("input.js", "const foo = parseInt(String /* comment */ (number), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 56,
						""),
				),
				truncationCase("input.js", "const foo = parseInt(String(number), /* comment */ 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 55,
						""),
				),
				truncationCase("input.js", "const foo = parseInt(String((0, number)), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 46,
						"const foo = Math.trunc((0, number));"),
				),
				truncationCase("input.ts", "const foo = parseInt(String(number as number), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 51,
						"const foo = Math.trunc(number as number);"),
				),
				truncationCase("input.ts", "const foo = parseInt(String<() => number>(number), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 55,
						"const foo = Math.trunc(number);"),
				),
				truncationCase("input.ts", "const foo = parseInt(String(number!), 10);",
					truncationError("parse-int", "parseInt(String(...), 10)", 1, 13, 1, 42,
						"const foo = Math.trunc(number!);"),
				),
				truncationCase("input.js", "const foo = 1.1 | 0;",
					truncationError("bitwise", "| 0", 1, 13, 1, 20,
						"const foo = Math.trunc(1.1);"),
				),
				truncationCase("input.js", "const foo = 111 | 0;",
					truncationError("bitwise", "| 0", 1, 13, 1, 20,
						"const foo = Math.trunc(111);"),
				),
				truncationCase("input.js", "const foo = (1 + 2 / 3.4) | 0;",
					truncationError("bitwise", "| 0", 1, 13, 1, 30,
						"const foo = Math.trunc(1 + 2 / 3.4);"),
				),
				truncationCase("input.js", "const foo = bar((1.4 | 0) + 2);",
					truncationError("bitwise", "| 0", 1, 18, 1, 25,
						"const foo = bar((Math.trunc(1.4)) + 2);"),
				),
				truncationCase("input.js", "const foo = (0, 1.4) | 0;",
					truncationError("bitwise", "| 0", 1, 13, 1, 25,
						"const foo = Math.trunc((0, 1.4));"),
				),
				truncationCase("input.js", "function foo() {return.1 | 0;}",
					truncationError("bitwise", "| 0", 1, 23, 1, 29,
						"function foo() {return Math.trunc(.1);}"),
				),
				truncationCase("input.js", "const foo = 1.4 | 0.;",
					truncationError("bitwise", "| 0.", 1, 13, 1, 21,
						"const foo = Math.trunc(1.4);"),
				),
				truncationCase("input.js", "const foo = 1.4 | .0;",
					truncationError("bitwise", "| .0", 1, 13, 1, 21,
						"const foo = Math.trunc(1.4);"),
				),
				truncationCase("input.js", "const foo = 1.4 | 0.0000_0000_0000;",
					truncationError("bitwise", "| 0.0000_0000_0000", 1, 13, 1, 35,
						"const foo = Math.trunc(1.4);"),
				),
				truncationCase("input.js", "const foo = 1.4 | 0b0;",
					truncationError("bitwise", "| 0b0", 1, 13, 1, 22,
						"const foo = Math.trunc(1.4);"),
				),
				truncationCase("input.js", "const foo = 1.4 | 0x0000_0000_0000;",
					truncationError("bitwise", "| 0x0000_0000_0000", 1, 13, 1, 35,
						"const foo = Math.trunc(1.4);"),
				),
				truncationCase("input.js", "const foo = 1.4 | 0o0;",
					truncationError("bitwise", "| 0o0", 1, 13, 1, 22,
						"const foo = Math.trunc(1.4);"),
				),
				truncationCase("input.js", "const foo = 1.23 | 0 | 4;",
					truncationError("bitwise", "| 0", 1, 13, 1, 21,
						"const foo = Math.trunc(1.23) | 4;"),
				),
				truncationCase("input.js", "const foo = ~~3.9;",
					truncationError("bitwise-not", "~~", 1, 13, 1, 18,
						"const foo = Math.trunc(3.9);"),
				),
				truncationCase("input.js", "const foo = ~~111;",
					truncationError("bitwise-not", "~~", 1, 13, 1, 18,
						"const foo = Math.trunc(111);"),
				),
				truncationCase("input.js", "const foo = ~~(1 + 2 / 3.4);",
					truncationError("bitwise-not", "~~", 1, 13, 1, 28,
						"const foo = Math.trunc(1 + 2 / 3.4);"),
				),
				truncationCase("input.js", "const foo = ~~1 + 2 / 3.4;",
					truncationError("bitwise-not", "~~", 1, 13, 1, 16,
						"const foo = Math.trunc(1) + 2 / 3.4;"),
				),
				truncationCase("input.js", "const foo = ~~(0, 1.4);",
					truncationError("bitwise-not", "~~", 1, 13, 1, 23,
						"const foo = Math.trunc((0, 1.4));"),
				),
				truncationCase("input.js", "const foo = ~~~10.01;",
					truncationError("bitwise-not", "~~", 1, 14, 1, 21,
						"const foo = ~Math.trunc(10.01);"),
				),
				truncationCase("input.js", "const foo = ~~(~10.01);",
					truncationError("bitwise-not", "~~", 1, 14, 1, 23,
						"const foo = ~Math.trunc(10.01);"),
				),
				truncationCase("input.js", "const foo = ~(~~10.01);",
					truncationError("bitwise-not", "~~", 1, 15, 1, 22,
						"const foo = ~(Math.trunc(10.01));"),
				),
				truncationCase("input.js", "const foo = ~~-10.01;",
					truncationError("bitwise-not", "~~", 1, 13, 1, 21,
						"const foo = Math.trunc(-10.01);"),
				),
				truncationCase("input.js", "const foo = ~~~~10.01;",
					truncationError("bitwise-not", "~~", 1, 15, 1, 22,
						"const foo = ~~Math.trunc(10.01);"),
				),
				truncationCase("input.js", "function foo() {return~~3.9;}",
					truncationError("bitwise-not", "~~", 1, 23, 1, 28,
						"function foo() {return Math.trunc(3.9);}"),
				),
				truncationCase("input.js", "const foo = bar >> 0;",
					truncationError("bitwise", ">> 0", 1, 13, 1, 21,
						"const foo = Math.trunc(bar);"),
				),
				truncationCase("input.js", "const foo = (a + b) >> 0;",
					truncationError("bitwise", ">> 0", 1, 13, 1, 25,
						"const foo = Math.trunc(a + b);"),
				),
				truncationCase("input.ts", "const foo = ~~bar!;",
					truncationError("bitwise-not", "~~", 1, 13, 1, 19,
						"const foo = Math.trunc(bar!);"),
				),
				truncationCase("input.js", "const foo = ~~(a ? b : c);",
					truncationError("bitwise-not", "~~", 1, 13, 1, 26,
						"const foo = Math.trunc(a ? b : c);"),
				),
				truncationCase("input.js", "const foo = bar << 0;",
					truncationError("bitwise", "<< 0", 1, 13, 1, 21,
						"const foo = Math.trunc(bar);"),
				),
				truncationCase("input.js", "const foo = bar ^ 0;",
					truncationError("bitwise", "^ 0", 1, 13, 1, 20,
						"const foo = Math.trunc(bar);"),
				),
				truncationCase("input.js", "function foo() {return.1 ^0;}",
					truncationError("bitwise", "^ 0", 1, 23, 1, 28,
						"function foo() {return Math.trunc(.1);}"),
				),
				truncationCase("input.js", `const foo = {a: {b: {c: 3}}};
const bar = a.b.c | 0;`,
					truncationError("bitwise", "| 0", 2, 13, 2, 22,
						`const foo = {a: {b: {c: 3}}};
const bar = Math.trunc(a.b.c);`),
				),
				truncationCase("input.js", `const foo = {a: {b: {c: 3}}};
const bar = a.b?.c | 0;`,
					truncationError("bitwise", "| 0", 2, 13, 2, 23,
						`const foo = {a: {b: {c: 3}}};
const bar = Math.trunc(a.b?.c);`),
				),
				truncationCase("input.js", `const foo = {a: {b: {c: 3}}};
const bar = ~~a.b?.c;`,
					truncationError("bitwise-not", "~~", 2, 13, 2, 21,
						`const foo = {a: {b: {c: 3}}};
const bar = Math.trunc(a.b?.c);`),
				),
				truncationCase("input.js", `const foo = 3;
const bar = foo | 0;`,
					truncationError("bitwise", "| 0", 2, 13, 2, 20,
						`const foo = 3;
const bar = Math.trunc(foo);`),
				),
				truncationCase("input.js", `const foo = 3;
const bar = ~~foo;`,
					truncationError("bitwise-not", "~~", 2, 13, 2, 18,
						`const foo = 3;
const bar = Math.trunc(foo);`),
				),
				truncationCase("input.js", `let foo = 2;
foo |= 0;`,
					truncationError("bitwise", "|= 0", 2, 1, 2, 9,
						`let foo = 2;
foo = Math.trunc(foo);`),
				),
				truncationCase("input.js", `const foo = {a: {b: 3.4}};
foo.a.b |= 0;`,
					truncationError("bitwise", "|= 0", 2, 1, 2, 13,
						`const foo = {a: {b: 3.4}};
foo.a.b = Math.trunc(foo.a.b);`),
				),
				truncationCase("input.js", `const foo = 10.01;
const bar = ~~foo;`,
					truncationError("bitwise-not", "~~", 2, 13, 2, 18,
						`const foo = 10.01;
const bar = Math.trunc(foo);`),
				),
				truncationCase("input.js", `let foo = 10.01;
foo >>= 0;`,
					truncationError("bitwise", ">>= 0", 2, 1, 2, 10,
						`let foo = 10.01;
foo = Math.trunc(foo);`),
				),
				truncationCase("input.js", `let foo = 10.01;
foo <<= 0;`,
					truncationError("bitwise", "<<= 0", 2, 1, 2, 10,
						`let foo = 10.01;
foo = Math.trunc(foo);`),
				),
				truncationCase("input.js", `let foo = 10.01;
foo ^= 0;`,
					truncationError("bitwise", "^= 0", 2, 1, 2, 9,
						`let foo = 10.01;
foo = Math.trunc(foo);`),
				),
				truncationCase("input.js", "function foo() {return[foo][0] ^= 0;};",
					truncationError("bitwise", "^= 0", 1, 23, 1, 36,
						"function foo() {return[foo][0] = Math.trunc([foo][0]);};"),
				),
				truncationCase("input.js", "const foo = /* first comment */ 3.4 | 0; // A B C",
					truncationError("bitwise", "| 0", 1, 33, 1, 40,
						"const foo = /* first comment */ Math.trunc(3.4); // A B C"),
				),
				truncationCase("input.js", "const foo = /* first comment */ ~~3.4; // A B C",
					truncationError("bitwise-not", "~~", 1, 33, 1, 38,
						"const foo = /* first comment */ Math.trunc(3.4); // A B C"),
				),
				truncationCase("input.js", `const foo = {a: {b: 3.4}};
foo /* Comment 1 */ .a /* Comment 2 */ . /* Comment 3 */ b |= /* Comment 4 */ 0 /* Comment 5 */;`,
					truncationError("bitwise", "|= 0", 2, 1, 2, 80,
						""),
				),
				truncationCase("input.js", `const foo = {a: {b: 3.4}};
const bar = /* Comment 1 */ ~~ a /* Comment 3 */ . /* Comment 4 */ b /* Comment 5 */;`,
					truncationError("bitwise-not", "~~", 2, 29, 2, 69,
						`const foo = {a: {b: 3.4}};
const bar = /* Comment 1 */ Math.trunc(a /* Comment 3 */ . /* Comment 4 */ b) /* Comment 5 */;`),
				),
				truncationCase("input.js", "const foo = /* will keep */ 3.4 /* will remove 1 */ | /* will remove 2 */ 0;",
					truncationError("bitwise", "| 0", 1, 29, 1, 76,
						""),
				),
				truncationCase("input.js", "const foo = /* will keep */ ~ /* will remove 1 */ ~ /* will remove 2 */ 3.4;",
					truncationError("bitwise-not", "~~", 1, 29, 1, 76,
						""),
				),
				truncationCase("input.js", `const foo = 3.4; // comment 1
foo |= 0; // comment 2`,
					truncationError("bitwise", "|= 0", 2, 1, 2, 9,
						`const foo = 3.4; // comment 1
foo = Math.trunc(foo); // comment 2`),
				),
				truncationCase("input.js", `const foo = 3.4; // comment 1
const bar = ~~foo; // comment 2`,
					truncationError("bitwise-not", "~~", 2, 13, 2, 18,
						`const foo = 3.4; // comment 1
const bar = Math.trunc(foo); // comment 2`),
				),
				truncationCase("input.js", "const foo = ~~bar | 0;",
					truncationError("bitwise", "| 0", 1, 13, 1, 22,
						"const foo = Math.trunc(~~bar);"),
					truncationError("bitwise-not", "~~", 1, 13, 1, 18,
						"const foo = Math.trunc(bar) | 0;"),
				),
				truncationCase("input.js", "const foo = ~~(bar| 0);",
					truncationError("bitwise-not", "~~", 1, 13, 1, 23,
						"const foo = Math.trunc(bar| 0);"),
					truncationError("bitwise", "| 0", 1, 16, 1, 22,
						"const foo = ~~(Math.trunc(bar));"),
				),
				truncationCase("input.js", "const foo = bar | 0 | 0;",
					truncationError("bitwise", "| 0", 1, 13, 1, 24,
						"const foo = Math.trunc(bar | 0);"),
					truncationError("bitwise", "| 0", 1, 13, 1, 20,
						"const foo = Math.trunc(bar) | 0;"),
				),
				truncationCase("input.js", "const foo = ~~~~((bar | 0 | 0) >> 0 >> 0 << 0 << 0 ^ 0 ^0);",
					truncationError("bitwise-not", "~~", 1, 15, 1, 59,
						"const foo = ~~Math.trunc((bar | 0 | 0) >> 0 >> 0 << 0 << 0 ^ 0 ^0);"),
					truncationError("bitwise", "^ 0", 1, 18, 1, 58,
						"const foo = ~~~~(Math.trunc((bar | 0 | 0) >> 0 >> 0 << 0 << 0 ^ 0));"),
					truncationError("bitwise", "^ 0", 1, 18, 1, 55,
						"const foo = ~~~~(Math.trunc((bar | 0 | 0) >> 0 >> 0 << 0 << 0) ^0);"),
					truncationError("bitwise", "<< 0", 1, 18, 1, 51,
						"const foo = ~~~~(Math.trunc((bar | 0 | 0) >> 0 >> 0 << 0) ^ 0 ^0);"),
					truncationError("bitwise", "<< 0", 1, 18, 1, 46,
						"const foo = ~~~~(Math.trunc((bar | 0 | 0) >> 0 >> 0) << 0 ^ 0 ^0);"),
					truncationError("bitwise", ">> 0", 1, 18, 1, 41,
						"const foo = ~~~~(Math.trunc((bar | 0 | 0) >> 0) << 0 << 0 ^ 0 ^0);"),
					truncationError("bitwise", ">> 0", 1, 18, 1, 36,
						"const foo = ~~~~(Math.trunc(bar | 0 | 0) >> 0 << 0 << 0 ^ 0 ^0);"),
					truncationError("bitwise", "| 0", 1, 19, 1, 30,
						"const foo = ~~~~((Math.trunc(bar | 0)) >> 0 >> 0 << 0 << 0 ^ 0 ^0);"),
					truncationError("bitwise", "| 0", 1, 19, 1, 26,
						"const foo = ~~~~((Math.trunc(bar) | 0) >> 0 >> 0 << 0 << 0 ^ 0 ^0);"),
				),
				truncationCase("input.js", `const foo = Array.from({length: 10}, (_, index) => (index + 1) + (index + 1) /100);
let i = 0;
while (i < foo.length) {
	foo[i++] |= 0;
}
console.log(foo);`,
					truncationError("bitwise", "|= 0", 4, 2, 4, 15,
						""),
				),
			})
	})
	t.Run("Explicit suggestion regressions", func(t *testing.T) {
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &prefer_math_trunc.PreferMathTruncRule,
			nil, []rule_tester.InvalidTestCase{
				truncationCase("input.js", "const foo = ~~(/* keep */ 1 + 2 / 3.4);",
					truncationError("bitwise-not", "~~", 1, 13, 1, 39,
						"const foo = Math.trunc((/* keep */ 1 + 2 / 3.4));"),
				),
				truncationCase("input.js", "const foo = /* keep */ ~ /* drop */ ~ 3.4;",
					truncationError("bitwise-not", "~~", 1, 24, 1, 42,
						""),
				),
				truncationCase("input.js", "const foo = x /* keep */ << 0;",
					truncationError("bitwise", "<< 0", 1, 13, 1, 30,
						""),
				),
				truncationCase("input.js", "const foo = ~~(1 + 2 / 3.4);",
					truncationError("bitwise-not", "~~", 1, 13, 1, 28,
						"const foo = Math.trunc(1 + 2 / 3.4);"),
				),
				truncationCase("input.js", "const foo = ~~(2 ** 32);",
					truncationError("bitwise-not", "~~", 1, 13, 1, 24,
						"const foo = Math.trunc(2 ** 32);"),
				),
				truncationCase("input.js", "const foo = x << 0;",
					truncationError("bitwise", "<< 0", 1, 13, 1, 19,
						"const foo = Math.trunc(x);"),
				),
				truncationCase("input.js", "foo |= /* keep */ 0;",
					truncationError("bitwise", "|= 0", 1, 1, 1, 20,
						""),
				),
			})
	})
	t.Run("Runtime regressions", func(t *testing.T) {
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &prefer_math_trunc.PreferMathTruncRule,
			nil, []rule_tester.InvalidTestCase{
				truncationCase("input.js", "(function* () { return ~~(yield read(/* keep */)); })()",
					truncationError("bitwise-not", "~~", 1, 24, 1, 50,
						"(function* () { return Math.trunc(yield read(/* keep */)); })()"),
				),
				truncationCase("input.js", "~~getValue(/* keep */);",
					truncationError("bitwise-not", "~~", 1, 1, 1, 23,
						"Math.trunc(getValue(/* keep */));"),
				),
				truncationCase("input.js", "getValue(/* keep */) | 0;",
					truncationError("bitwise", "| 0", 1, 1, 1, 25,
						"Math.trunc(getValue(/* keep */));"),
				),
				truncationCase("input.js", "getValue(/* keep */) >> 0;",
					truncationError("bitwise", ">> 0", 1, 1, 1, 26,
						"Math.trunc(getValue(/* keep */));"),
				),
				truncationCase("input.js", "getValue(/* keep */) << 0;",
					truncationError("bitwise", "<< 0", 1, 1, 1, 26,
						"Math.trunc(getValue(/* keep */));"),
				),
				truncationCase("input.js", "getValue(/* keep */) ^ 0;",
					truncationError("bitwise", "^ 0", 1, 1, 1, 25,
						"Math.trunc(getValue(/* keep */));"),
				),
			})
	})
	// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-math-trunc.md
	t.Run("Documentation examples", func(t *testing.T) {
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &prefer_math_trunc.PreferMathTruncRule,
			[]rule_tester.ValidTestCase{
				{Code: `// ✅
const foo = 37.4;
console.log(foo | 3);`, FileName: "input.js"},
				{Code: `// ✅
const foo = 37.4;
console.log(~foo);`, FileName: "input.js"},
				{Code: `// ✅
const foo = 37.4;
console.log(foo >> 3);`, FileName: "input.js"},
				{Code: `// ✅
const foo = 37.4;
console.log(foo << 3);`, FileName: "input.js"},
				{Code: `// ✅
const foo = 37.4;
console.log(foo ^ 3);`, FileName: "input.js"},
			}, []rule_tester.InvalidTestCase{
				truncationCase("input.js", `const foo = 37.4;

// ❌
console.log(foo | 0);

// ❌
console.log(~~foo);

// ❌
console.log(foo << 0);

// ❌
console.log(foo >> 0);

// ❌
console.log(foo.bar ^ 0);

// ❌
console.log(parseInt(String(foo), 10));

// ❌
console.log(Number.parseInt(String(foo), 10));

// ✅
console.log(Math.trunc(foo));`,
					truncationError("bitwise", "| 0", 4, 13, 4, 20,
						`const foo = 37.4;

// ❌
console.log(Math.trunc(foo));

// ❌
console.log(~~foo);

// ❌
console.log(foo << 0);

// ❌
console.log(foo >> 0);

// ❌
console.log(foo.bar ^ 0);

// ❌
console.log(parseInt(String(foo), 10));

// ❌
console.log(Number.parseInt(String(foo), 10));

// ✅
console.log(Math.trunc(foo));`),
					truncationError("bitwise-not", "~~", 7, 13, 7, 18,
						`const foo = 37.4;

// ❌
console.log(foo | 0);

// ❌
console.log(Math.trunc(foo));

// ❌
console.log(foo << 0);

// ❌
console.log(foo >> 0);

// ❌
console.log(foo.bar ^ 0);

// ❌
console.log(parseInt(String(foo), 10));

// ❌
console.log(Number.parseInt(String(foo), 10));

// ✅
console.log(Math.trunc(foo));`),
					truncationError("bitwise", "<< 0", 10, 13, 10, 21,
						`const foo = 37.4;

// ❌
console.log(foo | 0);

// ❌
console.log(~~foo);

// ❌
console.log(Math.trunc(foo));

// ❌
console.log(foo >> 0);

// ❌
console.log(foo.bar ^ 0);

// ❌
console.log(parseInt(String(foo), 10));

// ❌
console.log(Number.parseInt(String(foo), 10));

// ✅
console.log(Math.trunc(foo));`),
					truncationError("bitwise", ">> 0", 13, 13, 13, 21,
						`const foo = 37.4;

// ❌
console.log(foo | 0);

// ❌
console.log(~~foo);

// ❌
console.log(foo << 0);

// ❌
console.log(Math.trunc(foo));

// ❌
console.log(foo.bar ^ 0);

// ❌
console.log(parseInt(String(foo), 10));

// ❌
console.log(Number.parseInt(String(foo), 10));

// ✅
console.log(Math.trunc(foo));`),
					truncationError("bitwise", "^ 0", 16, 13, 16, 24,
						`const foo = 37.4;

// ❌
console.log(foo | 0);

// ❌
console.log(~~foo);

// ❌
console.log(foo << 0);

// ❌
console.log(foo >> 0);

// ❌
console.log(Math.trunc(foo.bar));

// ❌
console.log(parseInt(String(foo), 10));

// ❌
console.log(Number.parseInt(String(foo), 10));

// ✅
console.log(Math.trunc(foo));`),
					truncationError("parse-int", "parseInt(String(...), 10)", 19, 13, 19, 38,
						`const foo = 37.4;

// ❌
console.log(foo | 0);

// ❌
console.log(~~foo);

// ❌
console.log(foo << 0);

// ❌
console.log(foo >> 0);

// ❌
console.log(foo.bar ^ 0);

// ❌
console.log(Math.trunc(foo));

// ❌
console.log(Number.parseInt(String(foo), 10));

// ✅
console.log(Math.trunc(foo));`),
					truncationError("parse-int", "Number.parseInt(String(...), 10)", 22, 13, 22, 45,
						`const foo = 37.4;

// ❌
console.log(foo | 0);

// ❌
console.log(~~foo);

// ❌
console.log(foo << 0);

// ❌
console.log(foo >> 0);

// ❌
console.log(foo.bar ^ 0);

// ❌
console.log(parseInt(String(foo), 10));

// ❌
console.log(Math.trunc(foo));

// ✅
console.log(Math.trunc(foo));`),
				),
				truncationCase("input.js", `let foo = 37.4;

// ❌
foo |= 0;

// ✅
foo = Math.trunc(foo);`,
					truncationError("bitwise", "|= 0", 4, 1, 4, 9,
						`let foo = 37.4;

// ❌
foo = Math.trunc(foo);

// ✅
foo = Math.trunc(foo);`),
				),
			})
	})
}

func truncationCase(fileName, code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{FileName: fileName, Code: code, Errors: errors}
}

// An empty output expects no suggestion. Every case also expects no autofix.
func truncationError(kind, pattern string, line, column, endLine, endColumn int, output string) rule_tester.InvalidTestCaseError {
	expected := rule_tester.InvalidTestCaseError{
		MessageId: "error-" + kind,
		Message:   "Use `Math.trunc` instead of `" + pattern + "`.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
	}
	if output != "" {
		expected.Suggestions = append(expected.Suggestions, rule_tester.InvalidTestCaseSuggestion{
			MessageId: "suggestion-" + kind, Output: output,
		})
	}
	return expected
}
