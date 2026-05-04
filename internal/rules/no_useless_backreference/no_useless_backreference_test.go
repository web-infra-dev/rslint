package no_useless_backreference

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessBackreferenceRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessBackreferenceRule,
		[]rule_tester.ValidTestCase{
			// ---- not a regular expression ----
			{Code: `'\\1(a)'`},
			{Code: `regExp('\\1(a)')`},
			{Code: `new Regexp('\\1(a)', 'u')`},
			{Code: `RegExp.foo('\\1(a)', 'u')`},
			{Code: `new foo.RegExp('\\1(a)')`},

			// ---- unknown pattern ----
			{Code: `RegExp(p)`},
			{Code: `new RegExp(p, 'u')`},
			{Code: `RegExp('\\1(a)' + suffix)`},
			{Code: "new RegExp(`${prefix}\\\\1(a)`)"},

			// ---- not the global RegExp ----
			{Code: `let RegExp; new RegExp('\\1(a)');`},
			{Code: `function foo() { var RegExp; RegExp('\\1(a)', 'u'); }`},
			{Code: `function foo(RegExp) { new RegExp('\\1(a)'); }`},
			{Code: `if (foo) { const RegExp = bar; RegExp('\\1(a)'); }`},
			// SKIP: rslint does not support ESLint's /*globals*/ directive comments
			// `/* globals RegExp:off */ new RegExp('\\1(a)');`
			// `RegExp('\\1(a)');` with languageOptions.globals { RegExp: "off" }

			// ---- no capturing groups ----
			{Code: `/(?:)/`},
			{Code: `/(?:a)/`},
			{Code: `new RegExp('')`},
			{Code: `RegExp('(?:a)|(?:b)*')`},
			{Code: `/^ab|[cd].\n$/`},

			// ---- no backreferences ----
			{Code: `/(a)/`},
			{Code: `RegExp('(a)|(b)')`},
			{Code: `new RegExp('\\n\\d(a)')`},
			{Code: `/\0(a)/`},
			{Code: `/\0(a)/u`},
			{Code: `/(?<=(a))(b)(?=(c))/`},
			{Code: `/(?<!(a))(b)(?!(c))/`},
			{Code: `/(?<foo>a)/`},

			// ---- not really a backreference ----
			// SKIP: rslint fixtures use strict TS, which rejects '\1' octal escapes in string literals
			// `RegExp('\1(a)')` // string octal escape
			{Code: `RegExp('\\\\1(a)')`}, // escaped backslash → JS string "\\1(a)" → pattern "\\1(a)" (literal \, then 1(a))
			{Code: `/\\1(a)/`},           // escaped backslash in literal
			{Code: `/\1/`},               // group 1 doesn't exist, regex octal escape
			{Code: `/^\1$/`},
			{Code: `/\2(a)/`},
			{Code: `/\1(?:a)/`},
			{Code: `/\1(?=a)/`},
			{Code: `/\1(?!a)/`},
			{Code: `/^[\1](a)$/`},                  // \N in a character class is octal
			{Code: `new RegExp('[\\1](a)')`},
			{Code: `/\11(a)/`},                     // octal escape \11
			{Code: `/\k<foo>(a)/`},                 // no named groups → literal "k<foo>a"
			{Code: `/^(a)\1\2$/`},                  // \1 backref, \2 octal

			// ---- valid backreferences: correct position, after the group ----
			{Code: `/(a)\1/`},
			{Code: `/(a).\1/`},
			{Code: `RegExp('(a)\\1(b)')`},
			{Code: `/(a)(b)\2(c)/`},
			{Code: `/(?<foo>a)\k<foo>/`},
			{Code: `new RegExp('(.)\\1')`},
			{Code: `RegExp('(a)\\1(?:b)')`},
			{Code: `/(a)b\1/`},
			{Code: `/((a)\2)/`},
			{Code: `/((a)b\2c)/`},
			{Code: `/^(?:(a)\1)$/`},
			{Code: `/^((a)\2)$/`},
			{Code: `/^(((a)\3))|b$/`},
			{Code: `/a(?<foo>(.)b\2)/`},
			{Code: `/(a)?(b)*(\1)(c)/`},
			{Code: `/(a)?(b)*(\2)(c)/`},
			{Code: `/(?<=(a))b\1/`},
			{Code: `/(?<=(?=(a)\1))b/`},

			// ---- valid backreferences: correct position before the group when both in same lookbehind ----
			{Code: `/(?<!\1(a))b/`},
			{Code: `/(?<=\1(a))b/`},
			{Code: `/(?<!\1.(a))b/`},
			{Code: `/(?<=\1.(a))b/`},
			{Code: `/(?<=(?:\1.(a)))b/`},
			{Code: `/(?<!(?:\1)((a)))b/`},
			{Code: `/(?<!(?:\2)((a)))b/`},
			{Code: `/(?=(?<=\1(a)))b/`},
			{Code: `/(?=(?<!\1(a)))b/`},
			{Code: `/(.)(?<=\2(a))b/`},

			// ---- valid: not a reference into another alternative ----
			{Code: `/^(a)\1|b/`},
			{Code: `/^a|(b)\1/`},
			{Code: `/^a|(b|c)\1/`},
			{Code: `/^(a)|(b)\2/`},
			{Code: `/^(?:(a)|(b)\2)$/`},
			{Code: `/^a|(?:.|(b)\1)/`},
			{Code: `/^a|(?:.|(b).(\1))/`},
			{Code: `/^a|(?:.|(?:(b)).(\1))/`},
			{Code: `/^a|(?:.|(?:(b)|c).(\1))/`},
			{Code: `/^a|(?:.|(?:(b)).(\1|c))/`},
			{Code: `/^a|(?:.|(?:(b)|c).(\1|d))/`},

			// ---- valid: not a reference into a negative lookaround (reference from within is allowed) ----
			{Code: `/.(?=(b))\1/`},
			{Code: `/.(?<=(b))\1/`},
			{Code: `/a(?!(b)\1)./`},
			{Code: `/a(?<!\1(b))./`},
			{Code: `/a(?!(b)(\1))./`},
			{Code: `/a(?!(?:(b)\1))./`},
			{Code: `/a(?!(?:(b))\1)./`},
			{Code: `/a(?<!(?:\1)(b))./`},
			{Code: `/a(?<!(?:(?:\1)(b)))./`},
			{Code: `/(?<!(a))(b)(?!(c))\2/`},
			{Code: `/a(?!(b|c)\1)./`},

			// ---- ignore regular expressions with syntax errors ----
			{Code: `RegExp('\\1(a)[')`},                              // unterminated [
			{Code: `new RegExp('\\1(a){', 'u')`},                     // unterminated { in u mode
			{Code: `new RegExp('\\1(a)\\2', 'ug')`},                  // \2 syntax error in u mode
			{Code: `const flags = 'gus'; RegExp('\\1(a){', flags);`}, // u flag known via const
			{Code: `RegExp('\\1(a)\\k<foo>', 'u')`},                  // \k<foo> with no named group is syntax error in u
			{Code: `new RegExp('\\k<foo>(?<foo>a)\\k<bar>')`},        // \k<bar> for unknown group is syntax error

			// ---- ES2024 ----
			{Code: `new RegExp('([[A--B]])\\1', 'v')`},
			{Code: `new RegExp('[[]\\1](a)', 'v')`}, // SyntaxError

			// ---- ES2025 ----
			{Code: `/((?<foo>bar)\k<foo>|(?<foo>baz))/`},

			// ---- Extra edge cases (rslint-side hardening) ----
			// Deeply nested capturing groups with valid backref
			{Code: `/(((((((((a)))))))))\1/`},
			// Multi-digit valid backref (group 10 referenced by \10)
			{Code: `/(a)(b)(c)(d)(e)(f)(g)(h)(i)(j)\10/`},
			// Quantified group followed by valid backref
			{Code: `/(a)+\1/`},
			{Code: `/(a){2,5}\1/`},
			// Backref inside character class is treated as octal escape
			{Code: `/(a)[\1]/`},
			// Unicode-named group with \k<...>
			{Code: `/(?<日本>a)\k<日本>/u`},
			// Empty alternative + backref (edge: `(|a)\1` valid because \1 is after group)
			{Code: `/(|a)\1/`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- full message tests ----
			{
				Code: `/(b)(\2a)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nested", Message: `Backreference '\2' will be ignored. It references group '(\2a)' from within that group.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/\k<foo>(?<foo>bar)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `RegExp('(a|bc)|\\1')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "disjunctive", Message: `Backreference '\1' will be ignored. It references group '(a|bc)' which is in another alternative.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('(?!(?<foo>\\n))\\1')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "intoNegativeLookaround", Message: `Backreference '\1' will be ignored. It references group '(?<foo>\n)' which is in a negative lookaround.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/(?<!(a)\1)b/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "backward", Message: `Backreference '\1' will be ignored. It references group '(a)' which appears before in the same lookbehind.`, Line: 1, Column: 1},
				},
			},

			// ---- nested ----
			{Code: `new RegExp('(\\1)')`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/^(a\1)$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/^((a)\1)$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `new RegExp('^(a\\1b)$')`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `RegExp('^((\\1))$')`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/((\2))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/a(?<foo>(.)b\1)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/a(?<foo>\k<foo>)b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/^(\1)*$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/^(?:a)(?:((?:\1)))*$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(?!(\1))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/a|(b\1c)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(a|(\1))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(a|(\2))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(?:a|(\1))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(a)?(b)*(\3)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(?<=(a\1))b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},

			// ---- forward ----
			{Code: `/\1(a)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/\1.(a)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:\1)(?:(a))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:\1)(?:((a)))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:\2)(?:((a)))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:\1)(?:((?:a)))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(\2)(a)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `RegExp('(a)\\2(b)')`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:a)(b)\2(c)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/\k<foo>(?<foo>a)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:a(b)\2)(c)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `new RegExp('(a)(b)\\3(c)')`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/\1(?<=(a))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/\1(?<!(a))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?<=\1)(?<=(a))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?<!\1)(?<!(a))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?=\1(a))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?!\1(a))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},

			// ---- backward in the same lookbehind ----
			{Code: `/(?<=(a)\1)b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<!.(a).\1.)b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(.)(?<!(b|c)\2)d/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<=(?:(a)\1))b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<=(?:(a))\1)b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<=(a)(?:\1))b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<!(?:(a))(?:\1))b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<!(?:(a))(?:\1)|.)b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/.(?!(?<!(a)\1))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/.(?=(?<!(a)\1))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/.(?!(?<=(a)\1))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/.(?=(?<=(a)\1))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},

			// ---- into another alternative ----
			{Code: `/(a)|\1b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(a)|\1b)$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(a)|b(?:c|\1))$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:a|b(?:(c)|\1))$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(a(?!b))|\1b)+$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(?:(a)(?!b))|\1b)+$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(a(?=a))|\1b)+$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(a)(?=a)|\1b)+$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/.(?:a|(b)).|(?:(\1)|c)./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/.(?!(a)|\1)./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/.(?<=\1|(a))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},

			// ---- into a negative lookaround ----
			{Code: `/a(?!(b)).\1/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/(?<!(a))b\1/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/(?<!(a))(?:\1)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/.(?<!a|(b)).\1/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/.(?!(a)).(?!\1)./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/.(?<!(a)).(?<!\1)./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/.(?=(?!(a))\1)./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/.(?<!\1(?!(a)))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},

			// ---- valid and invalid ----
			{Code: `/\1(a)(b)\2/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/\1(a)\1/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},

			// ---- multiple invalid ----
			{
				Code: `/\1(a)\2(b)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			{
				Code: `/\1.(?<=(a)\1)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
					{MessageId: "backward", Line: 1, Column: 1},
				},
			},
			{
				Code: `/(?!\1(a)).\1/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
					{MessageId: "intoNegativeLookaround", Line: 1, Column: 1},
				},
			},
			{
				Code: `/(a)\2(b)/; RegExp('(\\1)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
					{MessageId: "nested", Line: 1, Column: 13},
				},
			},

			// ---- non-evaluable flags assumed to lack 'u' ----
			{
				Code: `RegExp('\\1(a){', flags);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},

			// ---- statically known expressions ----
			{
				Code: `const r = RegExp, p = '\\1', s = '(a)'; new r(p + s);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 41},
				},
			},

			// ---- ES2024 ----
			{
				Code: `new RegExp('\\1([[A--B]])', 'v')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\1' will be ignored. It references group '([[A--B]])' which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},

			// ---- ES2025 ----
			{
				Code: `/\k<foo>((?<foo>bar)|(?<foo>baz))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and another group which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?<foo>bar)|\k<foo>(?<foo>baz))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>baz)' which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/\k<foo>((?<foo>bar)|(?<foo>baz)|(?<foo>qux))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and other 2 groups which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?<foo>bar)|\k<foo>(?<foo>baz)|(?<foo>qux))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>baz)' which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?<foo>bar)|\k<foo>|(?<foo>baz))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "disjunctive", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and another group which is in another alternative.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?<foo>bar)|\k<foo>|(?<foo>baz)|(?<foo>qux))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "disjunctive", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and other 2 groups which is in another alternative.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?<foo>bar)|(?<foo>baz\k<foo>)|(?<foo>qux\k<foo>))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nested", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>baz\k<foo>)' from within that group.`, Line: 1, Column: 1},
					{MessageId: "nested", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>qux\k<foo>)' from within that group.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/(?<=((?<foo>bar)|(?<foo>baz))\k<foo>)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "backward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and another group which appears before in the same lookbehind.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?!(?<foo>bar))|(?!(?<foo>baz)))\k<foo>/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "intoNegativeLookaround", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and another group which is in a negative lookaround.`, Line: 1, Column: 1},
				},
			},

			// ---- Extra edge cases (rslint-side hardening) ----
			// Multi-digit forward reference (group 10 referenced before declared)
			{
				Code: `/\10(a)(b)(c)(d)(e)(f)(g)(h)(i)(j)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			// Deeply nested forward reference
			{
				Code: `/((((((\1))))))(a)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nested", Line: 1, Column: 1},
				},
			},
			// Quantified backref before its group
			{
				Code: `/\1+(a)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			// Unicode-named group forward reference
			{
				Code: `/\k<日本>(?<日本>a)/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			// Nested lookbehind — inner lookbehind still triggers backward
			{
				Code: `/(?<=(?<=(a)\1))b/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "backward", Line: 1, Column: 1},
				},
			},
		},
	)
}
