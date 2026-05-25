package prefer_template

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferTemplate(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferTemplateRule,
		[]rule_tester.ValidTestCase{
			{Code: `'use strict';`},
			{Code: `var foo = 'foo' + '\0';`},
			{Code: `var foo = 'bar';`},
			{Code: `var foo = 'bar' + 'baz';`},
			{Code: `var foo = foo + +'100';`},
			{Code: "var foo = `bar`;"},
			{Code: "var foo = `hello, ${name}!`;"},
			// https://github.com/eslint/eslint/issues/3507
			{Code: "var foo = `foo` + `bar` + \"hoge\";"},
			{Code: "var foo = `foo` +\n    `bar` +\n    \"hoge\";"},
			// Compound assignment is not `+`, so it must not trigger.
			{Code: `x += 'y';`},
			{Code: `let x = 0; x -= 1; x += 'y';`},
			// `+` on non-string operands stays untouched.
			{Code: `var foo = a + b;`},
			{Code: `var foo = 1 + 2;`},
			// Unary `+'100'` is treated as numeric coercion; no string concat.
			{Code: `var foo = a + +'100' + b;`},
			// Nothing to report: no string literal anywhere in the chain.
			{Code: `var foo = (a + b) * c;`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `var foo = 'hello, ' + name + '!';`,
				Output: []string{"var foo = `hello, ${  name  }!`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = bar + 'baz';`,
				Output: []string{"var foo = `${bar  }baz`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   "var foo = bar + `baz`;",
				Output: []string{"var foo = `${bar  }baz`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = +100 + 'yen';`,
				Output: []string{"var foo = `${+100  }yen`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = 'bar' + baz;`,
				Output: []string{"var foo = `bar${  baz}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = '￥' + (n * 1000) + '-'`,
				Output: []string{"var foo = `￥${  n * 1000  }-`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = 'aaa' + aaa; var bar = 'bbb' + bbb;`,
				Output: []string{"var foo = `aaa${  aaa}`; var bar = `bbb${  bbb}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 34},
				},
			},
			{
				Code:   `var string = (number + 1) + 'px';`,
				Output: []string{"var string = `${number + 1  }px`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 14},
				},
			},
			{
				Code:   `var foo = 'bar' + baz + 'qux';`,
				Output: []string{"var foo = `bar${  baz  }qux`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   "var foo = '0 backslashes: ${bar}' + baz;",
				Output: []string{"var foo = `0 backslashes: \\${bar}${  baz}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   "var foo = '1 backslash: \\${bar}' + baz;",
				Output: []string{"var foo = `1 backslash: \\${bar}${  baz}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   "var foo = '2 backslashes: \\\\${bar}' + baz;",
				Output: []string{"var foo = `2 backslashes: \\\\\\${bar}${  baz}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   "var foo = '3 backslashes: \\\\\\${bar}' + baz;",
				Output: []string{"var foo = `3 backslashes: \\\\\\${bar}${  baz}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   "var foo = bar + 'this is a backtick: `' + baz;",
				Output: []string{"var foo = `${bar  }this is a backtick: \\`${  baz}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   "var foo = bar + 'this is a backtick preceded by a backslash: \\`' + baz;",
				Output: []string{"var foo = `${bar  }this is a backtick preceded by a backslash: \\`${  baz}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   "var foo = bar + 'this is a backtick preceded by two backslashes: \\\\`' + baz;",
				Output: []string{"var foo = `${bar  }this is a backtick preceded by two backslashes: \\\\\\`${  baz}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   "var foo = bar + `${baz}foo`;",
				Output: []string{"var foo = `${bar  }${baz}foo`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code: "var foo = 'favorites: ' + favorites.map(f => {\n" +
					"    return f.name;\n" +
					"}) + ';';",
				Output: []string{"var foo = `favorites: ${  favorites.map(f => {\n" +
					"    return f.name;\n" +
					"})  };`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = bar + baz + 'qux';`,
				Output: []string{"var foo = `${bar + baz  }qux`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code: "var foo = 'favorites: ' +\n" +
					"    favorites.map(f => {\n" +
					"        return f.name;\n" +
					"    }) +\n" +
					"';';",
				Output: []string{"var foo = `favorites: ${ \n" +
					"    favorites.map(f => {\n" +
					"        return f.name;\n" +
					"    }) \n" +
					"};`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var foo = /* a */ 'bar' /* b */ + /* c */ baz /* d */ + 'qux' /* e */ ;`,
				Output: []string{"var foo = /* a */ `bar${ /* b */  /* c */ baz /* d */  }qux` /* e */ ;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 19},
				},
			},
			{
				Code:   `var foo = bar + ('baz') + 'qux' + (boop);`,
				Output: []string{"var foo = `${bar  }baz` + `qux${  boop}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   `foo + 'unescapes an escaped single quote in a single-quoted string: \''`,
				Output: []string{"`${foo  }unescapes an escaped single quote in a single-quoted string: '`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `foo + "unescapes an escaped double quote in a double-quoted string: \""`,
				Output: []string{"`${foo  }unescapes an escaped double quote in a double-quoted string: \"`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   "foo + 'does not unescape an escaped double quote in a single-quoted string: \\\"'",
				Output: []string{"`${foo  }does not unescape an escaped double quote in a single-quoted string: \\\"`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   "foo + \"does not unescape an escaped single quote in a double-quoted string: \\'\"",
				Output: []string{"`${foo  }does not unescape an escaped single quote in a double-quoted string: \\'`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				// "\x27" === "'"
				Code:   `foo + 'handles unicode escapes correctly: \x27'`,
				Output: []string{"`${foo  }handles unicode escapes correctly: \\x27`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			// NOTE: ESLint tests for octal escape sequences (\033, \8, \0\1, \08)
			// cannot be ported to Go tests because the TypeScript parser rejects
			// them as syntax errors. The no-autofix logic for those sequences is
			// still implemented — verified via code inspection.
			{
				Code:   `foo + '\\033'`,
				Output: []string{"`${foo  }\\\\033`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `foo + '\0'`,
				Output: []string{"`${foo  }\\0`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},

			// https://github.com/eslint/eslint/issues/15083
			{
				Code: "\"default-src 'self' https://*.google.com;\"\n" +
					"            + \"frame-ancestors 'none';\"\n" +
					"            + \"report-to \" + foo + \";\"",
				Output: []string{"`default-src 'self' https://*.google.com;`\n" +
					"            + `frame-ancestors 'none';`\n" +
					"            + `report-to ${  foo  };`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + 'b' + foo`,
				Output: []string{"`a` + `b${  foo}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + 'b' + foo + 'c' + 'd'`,
				Output: []string{"`a` + `b${  foo  }c` + `d`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + 'b + c' + foo + 'd' + 'e'`,
				Output: []string{"`a` + `b + c${  foo  }d` + `e`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + 'b' + foo + ('c' + 'd')`,
				Output: []string{"`a` + `b${  foo  }c` + `d`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + 'b' + foo + ('a' + 'b')`,
				Output: []string{"`a` + `b${  foo  }a` + `b`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + 'b' + foo + ('c' + 'd') + ('e' + 'f')`,
				Output: []string{"`a` + `b${  foo  }c` + `d` + `e` + `f`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `foo + ('a' + 'b') + ('c' + 'd')`,
				Output: []string{"`${foo  }a` + `b` + `c` + `d`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + foo + ('b' + 'c') + ('d' + bar + 'e')`,
				Output: []string{"`a${  foo  }b` + `c` + `d${  bar  }e`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `foo + ('b' + 'c') + ('d' + bar + 'e')`,
				Output: []string{"`${foo  }b` + `c` + `d${  bar  }e`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + 'b' + foo + ('c' + 'd' + 'e')`,
				Output: []string{"`a` + `b${  foo  }c` + `d` + `e`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + 'b' + foo + ('c' + bar + 'd')`,
				Output: []string{"`a` + `b${  foo  }c${  bar  }d`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + 'b' + foo + ('c' + bar + ('d' + 'e') + 'f')`,
				Output: []string{"`a` + `b${  foo  }c${  bar  }d` + `e` + `f`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + 'b' + foo + ('c' + bar + 'e') + 'f' + test`,
				Output: []string{"`a` + `b${  foo  }c${  bar  }e` + `f${  test}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + foo + ('b' + bar + 'c') + ('d' + test)`,
				Output: []string{"`a${  foo  }b${  bar  }c` + `d${  test}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + foo + ('b' + 'c') + ('d' + bar)`,
				Output: []string{"`a${  foo  }b` + `c` + `d${  bar}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `foo + ('a' + bar + 'b') + 'c' + test`,
				Output: []string{"`${foo  }a${  bar  }b` + `c${  test}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   "'a' + '`b`' + c",
				Output: []string{"`a` + `\\`b\\`${  c}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   "'a' + '`b` + `c`' + d",
				Output: []string{"`a` + `\\`b\\` + \\`c\\`${  d}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   "'a' + b + ('`c`' + '`d`')",
				Output: []string{"`a${  b  }\\`c\\`` + `\\`d\\``"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   "'`a`' + b + ('`c`' + '`d`')",
				Output: []string{"`\\`a\\`${  b  }\\`c\\`` + `\\`d\\``"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   "foo + ('`a`' + bar + '`b`') + '`c`' + test",
				Output: []string{"`${foo  }\\`a\\`${  bar  }\\`b\\`` + `\\`c\\`${  test}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + ('b' + 'c') + d`,
				Output: []string{"`a` + `b` + `c${  d}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   "'a' + ('`b`' + '`c`') + d",
				Output: []string{"`a` + `\\`b\\`` + `\\`c\\`${  d}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `a + ('b' + 'c') + d`,
				Output: []string{"`${a  }b` + `c${  d}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `a + ('b' + 'c') + (d + 'e')`,
				Output: []string{"`${a  }b` + `c${  d  }e`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   "a + ('`b`' + '`c`') + d",
				Output: []string{"`${a  }\\`b\\`` + `\\`c\\`${  d}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   "a + ('`b` + `c`' + '`d`') + e",
				Output: []string{"`${a  }\\`b\\` + \\`c\\`` + `\\`d\\`${  e}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + ('b' + 'c' + 'd') + e`,
				Output: []string{"`a` + `b` + `c` + `d${  e}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'a' + ('b' + 'c' + 'd' + (e + 'f') + 'g' +'h' + 'i') + j`,
				Output: []string{"`a` + `b` + `c` + `d${  e  }fg` +`h` + `i${  j}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `a + (('b' + 'c') + 'd')`,
				Output: []string{"`${a  }b` + `c` + `d`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `(a + 'b') + ('c' + 'd') + e`,
				Output: []string{"`${a  }b` + `c` + `d${  e}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `var foo = "Hello " + "world " + "another " + test`,
				Output: []string{"var foo = `Hello ` + `world ` + `another ${  test}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 11},
				},
			},
			{
				Code:   `'Hello ' + '"world" ' + test`,
				Output: []string{"`Hello ` + `\"world\" ${  test}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `"Hello " + "'world' " + test`,
				Output: []string{"`Hello ` + `'world' ${  test}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},

			// ---- Additional coverage: operand kinds ----
			// Unary prefix on the left.
			{
				Code:   `+x + 'y'`,
				Output: []string{"`${+x  }y`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `typeof x + ' yen'`,
				Output: []string{"`${typeof x  } yen`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			// Call / property / element / optional-chain access.
			{
				Code:   `foo() + ' done'`,
				Output: []string{"`${foo()  } done`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'count: ' + foo(a, b)`,
				Output: []string{"`count: ${  foo(a, b)}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `obj.prop + '!'`,
				Output: []string{"`${obj.prop  }!`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `arr[0] + '!'`,
				Output: []string{"`${arr[0]  }!`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `a?.b + '!'`,
				Output: []string{"`${a?.b  }!`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},

			// ---- Mixed operators inside parentheses ----
			// Non-`+` binary expressions become a single `${...}` slot.
			{
				Code:   `(a * b) + 'px'`,
				Output: []string{"`${a * b  }px`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `(a || b) + '?'`,
				Output: []string{"`${a || b  }?`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'is: ' + (a ?? b)`,
				Output: []string{"`is: ${  a ?? b}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},

			// ---- TypeScript-specific operand kinds ----
			{
				Code:   `'v=' + (x as any)`,
				Output: []string{"`v=${  x as any}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'v=' + y!`,
				Output: []string{"`v=${  y!}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},

			// ---- No-space and whitespace edge cases ----
			{
				Code:   `'a'+b`,
				Output: []string{"`a${b}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `a+'b'`,
				Output: []string{"`${a}b`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},

			// ---- Nested in other contexts ----
			// Call argument: fix only rewrites the inner concat.
			{
				Code:   `foo('x=' + value, flag);`,
				Output: []string{"foo(`x=${  value}`, flag);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 5},
				},
			},
			// Return statement.
			{
				Code:   `function f(v) { return 'v=' + v; }`,
				Output: []string{"function f(v) { return `v=${  v}`; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 24},
				},
			},
			// Object property value.
			{
				Code:   `var o = { k: 'v=' + v };`,
				Output: []string{"var o = { k: `v=${  v}` };"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 14},
				},
			},
			// Computed property key.
			{
				Code:   `var o = { ['k-' + i]: 1 };`,
				Output: []string{"var o = { [`k-${  i}`]: 1 };"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 12},
				},
			},
			// Array element.
			{
				Code:   `var a = ['x-' + i, 'y'];`,
				Output: []string{"var a = [`x-${  i}`, 'y'];"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 10},
				},
			},
			// Inside a template span — the inner concat is fixed; the outer
			// template is untouched.
			{
				Code:   "var x = `[${'k=' + v}]`;",
				Output: []string{"var x = `[${`k=${  v}`}]`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 13},
				},
			},

			// ---- Tagged template as a non-string operand ----
			{
				Code:   "tag`t` + 'x'",
				Output: []string{"`${tag`t`  }x`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},

			// ---- Two independent concat chains on the same line report twice ----
			{
				Code:   `var a = 'x' + v; var b = 'y' + w;`,
				Output: []string{"var a = `x${  v}`; var b = `y${  w}`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 9},
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 26},
				},
			},

			// ---- Deeply nested parens around a string literal ----
			{
				Code:   `((('a'))) + b`,
				Output: []string{"`a${  b}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},

			// ---- String with regular escape sequences survives conversion ----
			{
				Code:   `'line\n' + x`,
				Output: []string{"`line\\n${  x}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `'\u{1F600}' + x`,
				Output: []string{"`\\u{1F600}${  x}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},

			// ---- Walk-up must terminate in non-concat statement contexts ----
			// Parent chain: BinaryExpression -> ThrowStatement (not concat).
			{
				Code:   `function f(e) { throw 'error: ' + e; }`,
				Output: []string{"function f(e) { throw `error: ${  e}`; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 23},
				},
			},
			// Parent chain: BinaryExpression -> IfStatement condition.
			{
				Code:   `if ('k=' + k) {}`,
				Output: []string{"if (`k=${  k}`) {}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 5},
				},
			},
			// Concat inside one branch of a ternary — the other branch is untouched.
			{
				Code:   `var r = cond ? 'a' + x : 'b';`,
				Output: []string{"var r = cond ? `a${  x}` : 'b';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 16},
				},
			},

			// ---- Expression-kind operands that land in the fallback `${expr}` slot ----
			{
				Code:   `new Foo() + 'x'`,
				Output: []string{"`${new Foo()  }x`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x++ + 'y'`,
				Output: []string{"`${x++  }y`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			// Comma expression flattens inside the `${}` slot — parens are shed,
			// but `${a, b}` is still a valid template-curly expression.
			{
				Code:   `(a, b) + 'x'`,
				Output: []string{"`${a, b  }x`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			// `await` operand in an async function body.
			{
				Code:   `async function g() { return 'v=' + await fetch(); }`,
				Output: []string{"async function g() { return `v=${  await fetch()}`; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 29},
				},
			},

			// ---- Paren-with-internal-whitespace (ESLint `getTokensBetween`
			// alignment). Whitespace inside the parens around an operand must
			// be merged into the resulting template curly, not dropped. ----
			{
				Code:   `( 'baz' ) + x`,
				Output: []string{"`baz${   x}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `x + (  'baz'  )`,
				Output: []string{"`${x    }baz`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
			{
				Code:   `( ( x ) ) + 'y'`,
				Output: []string{"`${x    }y`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation", Line: 1, Column: 1},
				},
			},
		},
	)
}
